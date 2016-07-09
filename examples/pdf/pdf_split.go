/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

/*
 * Splits PDF files, tries to decrypt encrypted documents with an empty password
 * as best effort.
 *
 * Run as: go run pdf_split.go page_from page_to output.pdf input.pdf
 * To get only page 1 and 2 from input.pdf and save as output.pdf run: go run pdf_split.go 1 2 output.pdf input.pdf
 */

package main

import (
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
	if len(os.Args) < 5 {
		fmt.Printf("Requires at least 4 arguments: page_from page_to output.pdf input.pdf\n")
		fmt.Printf("Usage: To get only page 1 and 2 from input.pdf and save as output.pdf run: go run pdf_split.go 1 2 output.pdf input.pdf\n")
		os.Exit(1)
	}

	strSplitFrom := os.Args[1]
	splitFrom, err := strconv.Atoi(strSplitFrom)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	strSplitTo := os.Args[2]
	splitTo, err := strconv.Atoi(strSplitTo)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	outputPath := os.Args[3]
	fmt.Println("O: " + outputPath)

	inputPath := os.Args[4]
	fmt.Println("I: " + inputPath)

	err = initUniDoc("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = splitPdf(inputPath, outputPath, splitFrom, splitTo)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Complete, see output file: %s\n", outputPath)
}

func splitPdf(inputPath string, outputPath string, pageFrom int, pageTo int) error {
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

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return err
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	if numPages < pageTo {
		return err
	}

	for i := pageFrom; i <= pageTo; i++ {
		pageNum := i

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}

		err = pdfWriter.AddPage(page)
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
