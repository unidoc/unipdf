/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

/*
 * Crop pages in a PDF file. Crops the view to a certain percentage
 * of the original.
 *
 * Run as: go run pdf_crop.go <percentage> output.pdf input.pdf
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	unicommon "github.com/unidoc/unidoc/common"
	unilicense "github.com/unidoc/unidoc/license"
	unipdf "github.com/unidoc/unidoc/pdf"
)

func initUniDoc(licenseKey string) error {
	if len(licenseKey) > 0 {
		err := unilicense.SetLicenseKey(licenseKey)
		if err != nil {
			return err
		}
	}

	// To make the library log we just have to initialise the logger which satisfies
	// the unicommon.Logger interface, unicommon.DummyLogger is the default and
	// does not do anything. Very easy to implement your own.
	unicommon.SetLogger(unicommon.DummyLogger{})

	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Requires at least 3 arguments: <percentage> output.pdf input.pdf\n")
		fmt.Printf("Usage: go run pdf_crop.go <percentage> output.pdf input.pdf\n")
		os.Exit(1)
	}

	percentageStr := os.Args[1]
	outputPath := os.Args[2]
	inputPath := os.Args[3]

	percentage, err := strconv.ParseInt(percentageStr, 10, 32)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if percentage < 0 || percentage > 100 {
		fmt.Printf("Percentage should be in the range 0 - 100 (%)\n")
		os.Exit(1)
	}

	err = initUniDoc("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = cropPdf(inputPath, outputPath, percentage)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Complete, see output file: %s\n", outputPath)
}

// Crop all pages by a given percentage.
func cropPdf(inputPath string, outputPath string, percentage int64) error {
	pdfWriter := unipdf.NewPdfWriter()

	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	pdfReader, err := unipdf.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	// Try decrypting both with given password and an empty one if that fails.
	if isEncrypted {
		success, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return err
		}
		if !success {
			return errors.New("Unable to decrypt pdf with empty pass")
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		obj, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}

		pageObj, ok := obj.(*unipdf.PdfIndirectObject)
		if !ok {
			return errors.New("Invalid page object")
		}

		pageDict, ok := pageObj.PdfObject.(*unipdf.PdfObjectDictionary)
		if !ok {
			return errors.New("Invalid page dictionary")
		}

		page, err := unipdf.NewPdfPage(*pageDict)
		if err != nil {
			return err
		}

		bbox, err := page.GetMediaBox()
		if err != nil {
			return err
		}

		// Zoom in on the page middle, with a scaled width and height.
		width := (*bbox).Urx - (*bbox).Llx
		height := (*bbox).Ury - (*bbox).Lly
		newWidth := width * float64(percentage) / 100.0
		newHeight := height * float64(percentage) / 100.0
		(*bbox).Llx += newWidth / 2
		(*bbox).Lly += newHeight / 2
		(*bbox).Urx -= newWidth / 2
		(*bbox).Ury -= newHeight / 2
		page.MediaBox = bbox

		// Swap out the page dictionary.
		pageObj.PdfObject = page.GetPageDict()

		err = pdfWriter.AddPage(pageObj)
		if err != nil {
			return err
		}
	}

	fWrite, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer fWrite.Close()

	err = pdfWriter.Write(fWrite)
	if err != nil {
		return err
	}

	return nil
}
