/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package model provides an interface for working with high-level objects (models) in PDF files, including
// reading and writing documents.
//
// The document structure of a PDF is constructed of a hierarchy of data models, representing a tree
// of information starting from the Document catalog (Figure 5 p. 80).
// It is based on the core package which handles core functionality such as file i/o, parsing and
// handling of primitive PDF objects (core.PdfObject).
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
