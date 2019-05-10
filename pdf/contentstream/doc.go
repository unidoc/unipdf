/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package contentstream provides functionality for parsing and creating content streams for PDF files.
//
// For processing and manipulating content streams, it allows parse the content stream into a list of
// operands that can then be processed further for rendering or extraction of information.
// The ContentStreamProcessor offers a basic engine for processing the content stream and can be used
// to render or modify the contents.
//
// For creating content streams, see NewContentCreator.  It allows adding multiple operands and then can
// be converted to a string for embedding in a PDF file.
//
// The contentstream package uses the core and model packages.
package contentstream
