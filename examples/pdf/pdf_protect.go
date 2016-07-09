/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

/*
 * Protects PDF files by setting a password on it. This example both sets user
 * and opening password and hard-codes the protection bits here, but easily adjusted
 * in the code here although not on the command line.
 *
 *
 * When reading the input it tries to decrypt with empty password if the input file
 * is encrypted, if that fails we fail also.
 *
 * Run as: go run pdf_protect.go password output.pdf input.pdf
 * To protect input.pdf with password 'test' and save as output.pdf run: go run pdf_protect.go test output.pdf input.pdf
 */

package main

import (
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
	if len(os.Args) < 4 {
		fmt.Printf("Requires at least 3 arguments: password output.pdf input.pdf\n")
		fmt.Printf("Usage: To protect input.pdf with password 'test' and save as output.pdf run: go run pdf_protect.go test output.pdf input.pdf\n")
		os.Exit(1)
	}

	password := os.Args[1]

	outputPath := os.Args[2]
	inputPath := os.Args[3]

	err := initUniDoc("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = protectPdf(inputPath, outputPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Complete, see output file: %s\n", outputPath)
}

func protectPdf(inputPath string, outputPath string, password string) error {
	pdfWriter := unipdf.NewPdfWriter()

	// Feel free to change these values when testing.
	allowPrinting := false
	allowModifications := true
	allowCopying := true
	allowForm := false

	permissions := unipdf.AccessPermissions{}
	permissions.Printing = allowPrinting
	permissions.Modify = allowModifications
	permissions.Annotate = allowModifications
	permissions.RotateInsert = allowModifications
	permissions.ExtractGraphics = allowCopying
	permissions.DisabilityExtract = allowCopying
	permissions.FillForms = allowForm
	permissions.LimitPrintQuality = false

	encryptOptions := &unipdf.EncryptOptions{}
	encryptOptions.Permissions = permissions

	err := pdfWriter.Encrypt([]byte(password), []byte(password), encryptOptions)
	if err != nil {
		return err
	}

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
		_, err = pdfReader.Decrypt([]byte(password))
		if err != nil {
			// Fails, try fallback with empty password.
			_, err = pdfReader.Decrypt([]byte(""))
			if err != nil {
				return err
			}
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

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
