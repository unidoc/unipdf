/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

/*
 * Rotate pages in a PDF file.
 * Example of how to manipulate pages.
 *
 * Run as: go run pdf_rotate.go output.pdf input.pdf
 */

package main

import (
	"errors"
	"fmt"
	"os"

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
		fmt.Printf("Requires at least 2 arguments: output.pdf input.pdf\n")
		fmt.Printf("Usage: go run pdf_rotate.go output.pdf input.pdf\n")
		os.Exit(1)
	}

	outputPath := os.Args[1]
	inputPath := os.Args[2]

	err := initUniDoc("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = rotatePdf(inputPath, outputPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Complete, see output file: %s\n", outputPath)
}

// Rotate all pages by 90 degrees.
func rotatePdf(inputPath string, outputPath string) error {
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

		// Do the rotation.
		var rotation int64 = 0
		if page.Rotate != nil {
			rotation = *(page.Rotate)
		}
		rotation += 90 // Rotate by 90 deg.
		page.Rotate = &rotation

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
