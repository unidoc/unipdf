package tests

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2"
	"image/jpeg"
	"io/ioutil"
	"strings"

	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	pdf "github.com/unidoc/unidoc/pdf/model"
	"os"
	"path/filepath"
	"testing"
)

const (
	jbig2filesDir  = "jbig2files"
	jbig2ImagesDir = "jbig2_images"
)

func getFile(dirName, filename string) (*os.File, error) {
	return os.Open(filepath.Join(dirName, filename))
}

func readFileNames(dirname string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".pdf") {
			files = append(files, info.Name())
		}
		return nil
	})
	return files, err
}

func readJBIGZippedFiles(dirname string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasPrefix(info.Name(), "jbig2") && strings.HasSuffix(info.Name(), ".zip") {
				files = append(files, info.Name())
			}
		}
		return nil
	})
	return files, err
}

func readPDF(f *os.File, password ...string) (*pdf.PdfReader, error) {

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return nil, err
	}

	// check if is encrypted
	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}

	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, err
		}

		if !auth {
			if len(password) > 0 {
				auth, err = pdfReader.Decrypt([]byte(password[0]))
				if err != nil {
					return nil, err
				}
			}
			if !auth {
				return nil, fmt.Errorf("Reading the file: '%s' failed. Invalid password provided.", f.Name())
			}
		}
	}

	return pdfReader, nil
}

func rawFileName(filename string) string {
	if i := strings.IndexRune(filename, '.'); i != -1 {
		return filename[:i]
	}
	return filename
}

func extractImagesOnPage(filename string, page *pdf.PdfPage) ([]*pdf.Image, [][]byte, error) {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return nil, nil, err
	}
	return extractImagesInContentStream(filename, contents, page.Resources)
}

func extractImagesInContentStream(filename, contents string, resources *pdf.PdfPageResources) ([]*pdf.Image, [][]byte, error) {
	rgbImages := []*pdf.Image{}
	jbig2RawData := [][]byte{}
	cstreamParser := pdfcontent.NewContentStreamParser(contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		return nil, nil, err
	}

	processedXObjects := map[string]bool{}

	// Range through all the content stream operations.
	for _, op := range *operations {
		if op.Operand == "BI" && len(op.Params) == 1 {
			// BI: Inline image.

			// iimg, ok := op.Params[0].(*pdfcontent.ContentStreamInlineImage)
			// if !ok {
			// 	continue
			// }

			// iimg.

			// img, err := iimg.ToImage(resources)
			// if err != nil {
			// 	return nil, nil, err
			// }

			// cs, err := iimg.GetColorSpace(resources)
			// if err != nil {
			// 	return nil, nil, err
			// }
			// if cs == nil {
			// 	// Default if not specified?
			// 	cs = pdf.NewPdfColorspaceDeviceGray()
			// }
			// fmt.Printf("Cs: %T\n", cs)

			// rgbImg, err := cs.ImageToRGB(*img)
			// if err != nil {
			// 	return nil, nil, err
			// }

			// rgbImages = append(rgbImages, &rgbImg)
			// inlineImages++
		} else if op.Operand == "Do" && len(op.Params) == 1 {
			// Do: XObject.
			name := op.Params[0].(*pdfcore.PdfObjectName)

			// Only process each one once.
			_, has := processedXObjects[string(*name)]
			if has {
				continue
			}
			processedXObjects[string(*name)] = true

			xobj, xtype := resources.GetXObjectByName(*name)
			if xtype == pdf.XObjectTypeImage {
				filterObj := pdfcore.TraceToDirectObject(xobj.Get("Filter"))
				if filterObj == nil {
					continue
				}

				method, ok := filterObj.(*pdfcore.PdfObjectName)
				if !ok {
					continue
				}

				if method.String() != pdfcore.StreamEncodingFilterNameJBIG2 {
					continue
				}

				fmt.Printf(" XObject Image: %s\n", *name)

				ximg, err := resources.GetXObjectImageByName(*name)
				if err != nil {
					return nil, nil, err
				}

				common.Log.Debug("Components: %d", ximg.ColorSpace.GetNumComponents())
				common.Log.Debug("BitsPerComponent: %d", *ximg.BitsPerComponent)

				img, err := ximg.ToImage()
				if err != nil {
					return nil, nil, err
				}

				rgbImages = append(rgbImages, img)
				jbig2RawData = append(jbig2RawData, xobj.Stream)

			} else if xtype == pdf.XObjectTypeForm {
				// Go through the XObject Form content stream.
				xform, err := resources.GetXObjectFormByName(*name)
				if err != nil {
					return nil, nil, err
				}

				formContent, err := xform.GetContentStream()
				if err != nil {
					return nil, nil, err
				}

				// Process the content stream in the Form object too:
				formResources := xform.Resources
				if formResources == nil {
					formResources = resources
				}

				// Process the content stream in the Form object too:
				formRgbImages, jbig2data, err := extractImagesInContentStream(filename, string(formContent), formResources)
				if err != nil {
					return nil, nil, err
				}
				rgbImages = append(rgbImages, formRgbImages...)
				jbig2RawData = append(jbig2RawData, jbig2data...)
			}
		}
	}
	// common.Log.Debug("Extracted: '%d' XObject images.", xObjectImages)
	// common.Log.Debug("Extracted: '%d' Inline images.", inlineImages)

	return rgbImages, jbig2RawData, nil
}

func writeImages(t testing.TB, zw *zip.Writer, dirname, filename string, pageNo int, images ...*pdf.Image) {
	t.Helper()

	for idx, img := range images {
		fname := fmt.Sprintf("%s_%d_%d.jpg", rawFileName(filename), pageNo, idx)
		common.Log.Debug("Writing: %s", fname)
		f, err := zw.Create(fname)
		require.NoError(t, err)

		gimg, err := img.ToGoImage()
		require.NoError(t, err)

		// write to file
		// require.NoError(t, bmp.Encode(f, gimg))
		q := &jpeg.Options{Quality: 100}
		require.NoError(t, jpeg.Encode(f, gimg, q))
	}
}

func writeJBIG2Files(t testing.TB, zw *zip.Writer, dirname, filename string, pageNo int, files ...[]byte) {
	t.Helper()

	for idx, file := range files {
		fname := fmt.Sprintf("%s_%d_%d.jbig2", rawFileName(filename), pageNo, idx)
		common.Log.Debug("Writing: %s", fname)
		f, err := zw.Create(fname)
		require.NoError(t, err)

		_, err = f.Write(file)

		require.NoError(t, err)
	}
}

// TestDecodeSingle tests single file single file
func TestDecodeSingle(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}
	f, err := os.Open("000187_55_0.jbig2")
	require.NoError(t, err)

	data, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	d, err := jbig2.NewDocument(data)
	require.NoError(t, err)

	p, err := d.GetPage(1)
	require.NoError(t, err)

	bm, err := p.GetBitmap()
	require.NoError(t, err)

	h := md5.New()

	_, err = h.Write(bm.Data)
	require.NoError(t, err)

	t.Logf("Hash: %s", hex.EncodeToString(h.Sum(nil)))

}
