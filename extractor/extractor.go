/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"github.com/unidoc/unipdf/v3/model"
)

// Extractor stores and offers functionality for extracting content from PDF pages.
type Extractor struct {
	// stream contents and resources for page
	contents  string
	resources *model.PdfPageResources

	// fontCache is a simple LRU cache that is used to prevent redundant constructions of PdfFont's from
	// PDF objects. NOTE: This is not a conventional glyph cache. It only caches PdfFont's.
	fontCache map[string]fontEntry

	// text results from running extractXYText on forms within the page.
	// TODO(peterwilliams): Cache this map accross all pages in a PDF to speed up processig.
	formResults map[string]textResult

	// accessCount is used to set fontEntry.access to an incrementing number.
	accessCount int64

	// textCount is an incrementing number used to identify XYTest objects.
	textCount int64
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

	e := &Extractor{
		contents:    contents,
		resources:   page.Resources,
		fontCache:   map[string]fontEntry{},
		formResults: map[string]textResult{},
	}
	return e, nil
}
