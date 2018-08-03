/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import "github.com/unidoc/unidoc/pdf/model"

// Extractor stores and offers functionality for extracting content from PDF pages.
type Extractor struct {
	contents  string
	resources *model.PdfPageResources
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

	e := &Extractor{contents: contents, resources: page.Resources}
	return e, nil
}
