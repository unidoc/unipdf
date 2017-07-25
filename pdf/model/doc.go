/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// The model package provides a convenient interface for reading, writing and working with PDF files.
// The package includes many high level PDF data models which can be used to access information or modifications.
// It is based on the core package which handles core functionality such as file i/o, parsing and handling of primitive
// PDF objects.
//
// As an example of the interface, the following snippet can read the PDF and output the number of pages:
//
//	f, err := os.Open(inputPath)
//		if err != nil {
//		return nil, err
//	}
//	defer f.Close()
// 	pdfReader, err := unipdf.NewPdfReader(f)
//	if err != nil {
//		fmt.Printf("Failed to read PDF file: %v\n", err)
//		os.Exit(1)
//	}
//	numPages, err := pdfReader.GetNumPages()
//	if err != nil {
//		fmt.Printf("Failed to get number of pages: %v\n", err)
//		os.Exit(1)
//	}
//	fmt.Printf("The PDF file has %d pages\n", numPages)
//
// For more examples, see the unidoc-examples repository on GitHub: https://github.com/unidoc/unidoc-examples
package model
