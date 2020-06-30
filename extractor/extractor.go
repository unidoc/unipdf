/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"

	"github.com/unidoc/unipdf/v3/model"
)

// Extractor stores and offers functionality for extracting content from PDF pages.
type Extractor struct {
	// stream contents and resources for page
	contents  string
	resources *model.PdfPageResources
	mediaBox  model.PdfRectangle

	// fontCache is a simple LRU cache that is used to prevent redundant constructions of PdfFonts
	// from PDF objects. NOTE: This is not a conventional glyph cache. It only caches PdfFonts.
	fontCache map[string]fontEntry

	// text results from running extractXYText on forms within the page.
	// TODO(peterwilliams97): Cache this map accross all pages in a PDF to speed up processing.
	formResults map[string]textResult

	// accessCount is used to set fontEntry.access to an incrementing number.
	accessCount int64

	// textCount is an incrementing number used to identify XYTest objects.
	textCount int
}

// New returns an Extractor instance for extracting content from the input PDF page.
func New(page *model.PdfPage) (*Extractor, error) {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return nil, err
	}

	// Uncomment these lines to see the contents of the page. For debugging.
	// fmt.Println("========================= +++ =========================")
	// fmt.Printf("%s\n", contents)
	// fmt.Println("========================= ::: =========================")

	mediaBox, err := page.GetMediaBox()
	if err != nil {
		return nil, fmt.Errorf("extractor requires mediaBox. %v", err)
	}
	e := &Extractor{
		contents:    contents,
		resources:   page.Resources,
		mediaBox:    *mediaBox,
		fontCache:   map[string]fontEntry{},
		formResults: map[string]textResult{},
	}
	return e, nil
}

// NewFromContents creates a new extractor from contents and page resources.
func NewFromContents(contents string, resources *model.PdfPageResources) (*Extractor, error) {
	e := &Extractor{
		contents:    contents,
		resources:   resources,
		fontCache:   map[string]fontEntry{},
		formResults: map[string]textResult{},
	}
	return e, nil
}
