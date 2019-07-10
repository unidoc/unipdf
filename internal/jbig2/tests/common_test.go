/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
)

const jbig2DecodedDirectory string = "jbig2_decoded_images"

func extractImagesOnPage(filename string, page *model.PdfPage) ([]*extractedImage, error) {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return nil, err
	}
	return extractImagesInContentStream(filename, contents, page.Resources)
}

type extractedImage struct {
	jbig2Data []byte
	pdfImage  *model.XObjectImage
}

func extractImagesInContentStream(filename, contents string, resources *model.PdfPageResources) ([]*extractedImage, error) {
	extractedImages := []*extractedImage{}
	processedXObjects := make(map[string]bool)
	cstreamParser := contentstream.NewContentStreamParser(contents)

	operations, err := cstreamParser.Parse()
	if err != nil {
		return nil, err
	}

	// Range through all the content stream operations.
	for _, op := range *operations {
		if !(op.Operand == "Do" && len(op.Params) == 1) {
			continue
		}
		// Do: XObject.
		name := op.Params[0].(*core.PdfObjectName)

		// Only process each one once.
		_, has := processedXObjects[string(*name)]
		if has {
			continue
		}
		processedXObjects[string(*name)] = true

		xobj, xtype := resources.GetXObjectByName(*name)
		if xtype == model.XObjectTypeImage {
			filterObj := core.TraceToDirectObject(xobj.Get("Filter"))
			if filterObj == nil {
				continue
			}

			method, ok := filterObj.(*core.PdfObjectName)
			if !ok {
				continue
			}

			if method.String() != core.StreamEncodingFilterNameJBIG2 {
				continue
			}

			ximg, err := resources.GetXObjectImageByName(*name)
			if err != nil {
				return nil, err
			}

			extracted := &extractedImage{
				pdfImage:  ximg,
				jbig2Data: xobj.Stream,
			}

			extractedImages = append(extractedImages, extracted)
		} else if xtype == model.XObjectTypeForm {
			// Go through the XObject Form content stream.
			xform, err := resources.GetXObjectFormByName(*name)
			if err != nil {
				return nil, err
			}

			formContent, err := xform.GetContentStream()
			if err != nil {
				return nil, err
			}

			// Process the content stream in the Form object too:
			formResources := xform.Resources
			if formResources == nil {
				formResources = resources
			}

			// Process the content stream in the Form object too:
			images, err := extractImagesInContentStream(filename, string(formContent), formResources)
			if err != nil {
				return nil, err
			}
			extractedImages = append(extractedImages, images...)
		}
	}
	return extractedImages, nil
}

type fileHash struct {
	fileName string
	hash     string
}

func getFile(dirName, filename string) (*os.File, error) {
	return os.Open(filepath.Join(dirName, filename))
}

func rawFileName(filename string) string {
	if i := strings.LastIndex(filename, "."); i != -1 {
		return filename[:i]
	}
	return filename
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
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".zip") {
			files = append(files, info.Name())
		}
		return nil
	})
	return files, err
}

func readPDF(f *os.File, password ...string) (*model.PdfReader, error) {
	pdfReader, err := model.NewPdfReader(f)
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
				return nil, fmt.Errorf("reading the file: '%s' failed. Invalid password provided", f.Name())
			}
		}
	}
	return pdfReader, nil
}

func writeExtractedImages(zw *zip.Writer, filename string, pageNo int, images ...*extractedImage) (hashes []fileHash, err error) {
	h := md5.New()

	// write images
	for idx, img := range images {
		fname := fmt.Sprintf("%s_%d_%d", rawFileName(filename), pageNo, idx)

		common.Log.Debug("Writing file: '%s'", fname)
		f, err := zw.Create(fname + ".jpg")
		if err != nil {
			return nil, err
		}

		cimg, err := img.pdfImage.ToImage()
		if err != nil {
			return nil, err
		}

		bm := bitmap.NewWithData(int(cimg.Width), int(cimg.Height), cimg.Data)

		gimg, err := cimg.ToGoImage()
		if err != nil {
			return nil, err
		}

		multiWriter := io.MultiWriter(f, h)

		// write to file
		q := &jpeg.Options{Quality: 100}
		if err = jpeg.Encode(multiWriter, gimg, q); err != nil {
			return nil, err
		}

		fh := fileHash{fileName: fname, hash: hex.EncodeToString(h.Sum(nil))}
		hashes = append(hashes, fh)
		h.Reset()

		f, err = zw.Create(fname + "_bitmap" + ".jpg")
		if err != nil {
			return nil, err
		}
		if err = jpeg.Encode(f, bm.ToImage(), q); err != nil {
			return nil, err
		}

		if err = writeJBIG2Stream(zw, fname+".jbig2", img.jbig2Data); err != nil {
			return nil, err
		}
	}
	return hashes, nil
}

func writeJBIG2Stream(zw *zip.Writer, filename string, data []byte) error {
	f, err := zw.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
