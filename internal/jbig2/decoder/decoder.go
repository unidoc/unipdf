/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package decoder

import (
	"errors"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Decoder is the structure used to decode JBIG2 encoded byte streams.
type Decoder struct {
	inputReader reader.StreamReader
	document    *document.Document

	currentDecodedPage int
	parameters         Parameters
}

// DecodePage decodes jbig2 encoded page for provided 'pageNumber' in the document.
func (d *Decoder) DecodePage(pageNumber int) ([]byte, error) {
	return d.decodePage(pageNumber)
}

// DecodeNextPage decodes next jbig2 encoded page and returns decoded byte stream
func (d *Decoder) DecodeNextPage() ([]byte, error) {
	d.currentDecodedPage++
	pageNumber := d.currentDecodedPage
	return d.decodePage(pageNumber)
}

func (d *Decoder) decodePage(pageNumber int) ([]byte, error) {
	if pageNumber < 0 {
		return nil, errors.New("invalid page number")
	}

	if pageNumber > int(d.document.NumberOfPages) {
		return nil, errors.New("no more images to decode")
	}

	page, err := d.document.GetPage(pageNumber)
	if err != nil {
		return nil, err
	}

	bm, err := page.GetBitmap()
	if err != nil {
		return nil, err
	}

	if d.parameters.Color == bitmap.Chocolate {
		bm.InverseData()
	}

	if !d.parameters.UnpaddedData {
		return bm.Data, nil
	}
	return bm.GetUnpaddedData()
}

// Decode prepares decoder for the jbig2 encoded 'input' data,
// with optional 'parameters' and optional Globally encoded
// data segments - 'globals'.
func Decode(input []byte, parameters Parameters, globals ...document.Globals) (*Decoder, error) {
	r := reader.New(input)

	doc, err := document.DecodeDocument(r, globals...)
	if err != nil {
		return nil, err
	}

	return &Decoder{inputReader: r, document: doc, parameters: parameters}, nil
}
