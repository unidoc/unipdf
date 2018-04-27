/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// UniDoc is a comprehensive PDF library for Go (golang). The library has advanced capabilities for generating,
// processing and modifying PDFs. UniDoc is written and supported by the owners of the
// FoxyUtils.com website, where the library is used to power many of the PDF services offered.
//
// Getting More Information
//
// Check out the Getting Started and Example sections, which showcase how to install UniDoc and provide numerous
// examples of using UniDoc to generate, process or modify PDF files.
// https://unidoc.io/examples/getting_started/
//
// The godoc for unidoc provides a detailed breakdown of the API and documentation for packages, types and methods.
// https://godoc.org/github.com/unidoc/unidoc
//
// Overview of Major Packages
//
// The API is composed of a few major packages:
//
//   - common: Provides common shared types such as Logger and utilities to check
//     license validity.
//
//   - pdf/core: The core package defines the primitive PDF object types and handles
//     the file reading I/O and parsing the primitive objects.
//
//   - pdf/model: The model package builds on the core package, to represent the PDF as
//     a structured model of the PDF primitive types. It has a reader and a writer to
//     read and process a PDF file based on the structured model. This serves as a basis
//     to perform a number of numerous tasks and can be used to work with a PDF in a
//     medium to high level interface, although it does require an understanding of the
//     PDF format and structure.
//
//   - pdf/creator: The PDF creator makes it easy to create new PDFs or modify existing
//     PDFs. It can also enable loading a template PDF, adding text/images and
//     generating an output PDF. It can be used to add text, images, and generate text
//     and graphical reports. It is designed with simplicity in mind, with the goal of
//     making it easy to create reports without needing any knowledge about the PDF
//     format or specifications.
//
//   - pdf/extractor: Package extractor is used for quickly extracting PDF content
//     through a simple interface. Currently offers functionality for extracting textual
//     content.
package unidoc
