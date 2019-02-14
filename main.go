package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"

	unicommon "github.com/unidoc/unidoc/common"
	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/core/ccittfaxdecode"
	"github.com/unidoc/unidoc/pdf/creator"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

func getPixels(file io.Reader) ([][]byte, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	var pixels [][]byte
	for y := 0; y < h; y++ {
		var row []byte
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 65535 && g == 65535 && b == 65535 {
				// append white
				row = append(row, 1)
			} else {
				row = append(row, 0)
			}
		}
		pixels = append(pixels, row)
	}
	return pixels, nil
}

// sliceDiff compares two slices returning the first index of the different
// elements pair. Returns -1 if the slices contain the same elements
func slicesDiff(s1, s2 []byte) int {
	minLen := 0
	if len(s1) < len(s2) {
		minLen = len(s1)
	} else {
		minLen = len(s2)
	}
	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	return -1
}

/*func main() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	file, err := os.Open("/home/darkrengarius/Downloads/scan2.png")
	if err != nil {
		log.Fatalf("Error opening file: %v\n", err)
	}
	defer file.Close()
	pixels, err := getPixels(file)
	if err != nil {
		log.Fatalf("Error decoding the image: %v\n", err)
	}
	encoder := &ccitt.Encoder{BlackIs1: true}
	encoded := encoder.Encode(pixels)
	preparedBytes, err := ioutil.ReadFile("/home/darkrengarius/Downloads/scan2.gr3")
	if err != nil {
		log.Fatalf("Error opening gr3 file: %v\n", err)
	}
	log.Println(encoded)
	log.Println(preparedBytes)
	diffInd := slicesDiff(encoded, preparedBytes)
	if diffInd != -1 {
		log.Fatalf("Slices differ in %v. Encoded: %v, prepared: %v\n", diffInd,
			encoded[diffInd], preparedBytes[diffInd])
	}
	if len(encoded) != len(preparedBytes) {
		log.Fatalf("Slices differ in length")
	}
	log.Println("Slices are totally equal")
}
*/

var xObjectImages = 0
var inlineImages = 0

func main() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	file, err := os.Open("/home/darkrengarius/Downloads/p3_0.png")
	if err != nil {
		log.Fatalf("Error opening file: %v\n", err)
	}
	defer file.Close()
	pixels, err := getPixels(file)
	if err != nil {
		log.Fatalf("Error decoding the image: %v\n", err)
	}
	encoder := &ccittfaxdecode.Encoder{
		Columns:          2560,
		K:                -1,
		EndOfLine:        true,
		EncodedByteAlign: true,
		EndOfBlock:       true,
		//Rows:             3295,
	}
	encoded := encoder.Encode(pixels)

	f, err := os.Create("/home/darkrengarius/Downloads/K-1-Columns2560-EOL-Aligned-EOFB.gr4")
	if err != nil {
		log.Fatalf("Error creating file: %v\n", err)
	}

	_, err = f.Write(encoded)
	if err != nil {
		log.Fatalf("Error saving to file: %v\n", err)
	}

	log.Println()

	/*var files []string

	err := filepath.Walk("/home/darkrengarius/Downloads/4111112", func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".png" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error scanning directory: %v\n", err)
	}

	if err := imagesToPdf(files, "/home/darkrengarius/Downloads/testCombined4111112.pdf"); err != nil {
		log.Fatalf("Error writing images to pdf: %v\n", err)
	}*/

	// extract images from pdf to zip
	inputPath := "/home/darkrengarius/Downloads/testCombined000141.pdf"
	outputPath := "/home/darkrengarius/Downloads/testCombined000141.zip"

	fmt.Printf("Input file: %s\n", inputPath)
	err = extractImagesToArchive(inputPath, outputPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("-- Summary\n")
	fmt.Printf("%d XObject images extracted\n", xObjectImages)
	fmt.Printf("%d inline images extracted\n", inlineImages)
	fmt.Printf("Total %d images\n", xObjectImages+inlineImages)

}

/*func imagesToPdf(inputPaths []string, outputPath string) error {
	c := creator.New()

	for _, imgPath := range inputPaths {
		unicommon.Log.Debug("Image: %s", imgPath)

		file, err := os.Open(imgPath)
		if err != nil {
			log.Fatalf("Error opening file: %v\n", err)
		}

		imgF, _, err := image.Decode(file)
		if err != nil {
			file.Close()

			return err
		}

		file.Close()

		modelImg, err := pdf.ImageHandling.NewImageFromGoImage(imgF)
		if err != nil {
			unicommon.Log.Debug("Error loading image: %v", err)
			return err
		}
		modelImg.BitsPerComponent = 1
		modelImg.ColorComponents = 1

		img, err := creator.NewImage(modelImg)
		if err != nil {
			unicommon.Log.Debug("Error loading image: %v", err)
			return err
		}
		img.ScaleToWidth(612.0)

		// Use page width of 612 points, and calculate the height proportionally based on the image.
		// Standard PPI is 72 points per inch, thus a width of 8.5"
		height := 612.0 * img.Height() / img.Width()
		c.SetPageSize(creator.PageSize{612, height})
		c.NewPage()
		img.SetPos(0, 0)

		enc := pdfcore.NewCCITTFaxEncoder()
		enc.K = -4
		enc.Columns = int(modelImg.Width)
		enc.EndOfBlock = true
		enc.EndOfLine = true
		img.SetEncoder(enc)

		_ = c.Draw(img)
	}

	err := c.WriteToFile(outputPath)
	return err
}*/

// Images to PDF.
func imagesToPdf(inputPaths []string, outputPath string) error {
	c := creator.New()

	for _, imgPath := range inputPaths {
		unicommon.Log.Debug("Image: %s", imgPath)

		file, err := os.Open(imgPath)
		if err != nil {
			log.Fatalf("Error opening file: %v\n", err)
		}

		imgF, _, err := image.Decode(file)
		if err != nil {
			file.Close()

			return err
		}

		file.Close()

		modelImg, err := pdf.ImageHandling.NewImageFromGoImage(imgF)
		if err != nil {
			unicommon.Log.Debug("Error loading image: %v", err)
			return err
		}
		modelImg.BitsPerComponent = 1
		modelImg.ColorComponents = 1

		img, err := c.NewImage(modelImg)
		if err != nil {
			unicommon.Log.Debug("Error loading image: %v", err)
			return err
		}

		img.ScaleToWidth(612.0)

		// Use page width of 612 points, and calculate the height proportionally based on the image.
		// Standard PPI is 72 points per inch, thus a width of 8.5"
		height := 612.0 * img.Height() / img.Width()
		c.SetPageSize(creator.PageSize{612, height})
		c.NewPage()

		encoder := pdfcore.NewCCITTFaxEncoder()
		encoder.Columns = int(modelImg.Width)
		encoder.EndOfBlock = true
		encoder.EndOfLine = true
		encoder.EncodedByteAlign = true
		encoder.K = -1
		img.SetEncoder(encoder)

		img.SetPos(0, 0)
		_ = c.Draw(img)
	}

	err := c.WriteToFile(outputPath)
	return err
}

// Extracts images and properties of a PDF specified by inputPath.
// The output images are stored into a zip archive whose path is given by outputPath.
func extractImagesToArchive(inputPath, outputPath string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	// Try decrypting with an empty one.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			// Encrypted and we cannot do anything about it.
			return err
		}
		if !auth {
			fmt.Println("Need to decrypt with password")
			return nil
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}
	fmt.Printf("PDF Num Pages: %d\n", numPages)

	// Prepare output archive.
	zipf, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer zipf.Close()
	zipw := zip.NewWriter(zipf)

	for i := 0; i < numPages; i++ {
		fmt.Printf("-----\nPage %d:\n", i+1)

		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return err
		}

		// List images on the page.
		rgbImages, err := extractImagesOnPage(page)
		if err != nil {
			return err
		}
		_ = rgbImages

		for idx, img := range rgbImages {
			fname := fmt.Sprintf("p%d_%d.png", i+1, idx)

			gimg, err := img.ToGoImage()
			if err != nil {
				return err
			}

			imgf, err := zipw.Create(fname)
			if err != nil {
				return err
			}
			err = png.Encode(imgf, gimg)
			if err != nil {
				return err
			}
		}
	}

	// Make sure to check the error on Close.
	err = zipw.Close()
	if err != nil {
		return err
	}

	return nil
}

func extractImagesOnPage(page *pdf.PdfPage) ([]*pdf.Image, error) {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return nil, err
	}

	return extractImagesInContentStream(contents, page.Resources)
}

func extractImagesInContentStream(contents string, resources *pdf.PdfPageResources) ([]*pdf.Image, error) {
	rgbImages := []*pdf.Image{}
	cstreamParser := pdfcontent.NewContentStreamParser(contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		return nil, err
	}

	processedXObjects := map[string]bool{}

	// Range through all the content stream operations.
	for _, op := range *operations {
		if op.Operand == "BI" && len(op.Params) == 1 {
			// BI: Inline image.

			iimg, ok := op.Params[0].(*pdfcontent.ContentStreamInlineImage)
			if !ok {
				continue
			}

			img, err := iimg.ToImage(resources)
			if err != nil {
				return nil, err
			}

			cs, err := iimg.GetColorSpace(resources)
			if err != nil {
				return nil, err
			}
			if cs == nil {
				// Default if not specified?
				cs = pdf.NewPdfColorspaceDeviceGray()
			}
			fmt.Printf("Cs: %T\n", cs)

			rgbImg, err := cs.ImageToRGB(*img)
			if err != nil {
				return nil, err
			}

			rgbImages = append(rgbImages, &rgbImg)
			inlineImages++
		} else if op.Operand == "Do" && len(op.Params) == 1 {
			// Do: XObject.
			name := op.Params[0].(*pdfcore.PdfObjectName)

			// Only process each one once.
			_, has := processedXObjects[string(*name)]
			if has {
				continue
			}
			processedXObjects[string(*name)] = true

			_, xtype := resources.GetXObjectByName(*name)
			if xtype == pdf.XObjectTypeImage {
				fmt.Printf(" XObject Image: %s\n", *name)

				ximg, err := resources.GetXObjectImageByName(*name)
				if err != nil {
					return nil, err
				}

				img, err := ximg.ToImage()
				if err != nil {
					return nil, err
				}

				//img.ColorComponents = 3
				//img.BitsPerComponent = 8

				goimg, err := img.ToGoImage()
				if err != nil {
					return nil, err
				}

				f, err := os.Create("/home/darkrengarius/Downloads/testDECODED.png")
				if err != nil {
					return nil, err
				}
				defer f.Close()

				err = png.Encode(f, goimg)
				if err != nil {
					return nil, err
				}

				cs := ximg.ColorSpace
				if cs == nil {
					// Default if not specified?
					cs = pdf.NewPdfColorspaceDeviceGray()
				}

				//cs = pdf.NewPdfColorspaceDeviceRGB()

				fmt.Printf("Cs: %T\n", cs)

				rgbImg, err := cs.ImageToRGB(*img)
				if err != nil {
					return nil, err
				}

				rgbImages = append(rgbImages, &rgbImg)
				xObjectImages++
			} else if xtype == pdf.XObjectTypeForm {
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
				formRgbImages, err := extractImagesInContentStream(string(formContent), formResources)
				if err != nil {
					return nil, err
				}
				rgbImages = append(rgbImages, formRgbImages...)
			}
		}
	}

	return rgbImages, nil
}
