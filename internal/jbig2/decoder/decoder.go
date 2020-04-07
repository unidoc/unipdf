/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package decoder

import (
	"image"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
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

// DecodePageImage decodes page with 'pageNumber' from the document and stores it's result
// within go image.Image.
func (d *Decoder) DecodePageImage(pageNumber int) (image.Image, error) {
	const processName = "decoder.DecodePageImage"
	i, err := d.decodePageImage(pageNumber)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return i, nil
}

// DecodeNextPage decodes next jbig2 encoded page and returns decoded byte stream
func (d *Decoder) DecodeNextPage() ([]byte, error) {
	d.currentDecodedPage++
	pageNumber := d.currentDecodedPage
	return d.decodePage(pageNumber)
}

// PageNumber returns
func (d *Decoder) PageNumber() (int, error) {
	const processName = "Decoder.PageNumber"
	if d.document == nil {
		return 0, errors.Error(processName, "decoder not initialized yet")
	}
	return int(d.document.NumberOfPages), nil
}

func (d *Decoder) decodePage(pageNumber int) ([]byte, error) {
	const processName = "decodePage"
	if pageNumber < 0 {
		return nil, errors.Errorf(processName, "invalid page number: '%d'", pageNumber)
	}

	if pageNumber > int(d.document.NumberOfPages) {
		return nil, errors.Errorf(processName, "page: '%d' not found in the decoder", pageNumber)
	}

	// get the page from the document.
	page, err := d.document.GetPage(pageNumber)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	// get bitmap from this page.
	bm, err := page.GetBitmap()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	// inverse the color by default.
	bm.InverseData()

	if !d.parameters.UnpaddedData {
		return bm.Data, nil
	}
	return bm.GetUnpaddedData()
}

func (d *Decoder) decodePageImage(pageNumber int) (image.Image, error) {
	const processName = "decodePageImage"
	if pageNumber < 0 {
		return nil, errors.Errorf(processName, "invalid page number: '%d'", pageNumber)
	}

	if pageNumber > int(d.document.NumberOfPages) {
		return nil, errors.Errorf(processName, "page: '%d' not found in the decoder", pageNumber)
	}

	// get the page from the document.
	page, err := d.document.GetPage(pageNumber)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	// get bitmap from this page.
	bm, err := page.GetBitmap()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return bm.ToImage(), nil
}

// Decode prepares decoder for the jbig2 encoded 'input' data,
// with optional 'parameters' and optional Globally encoded
// data segments - 'globals'.
func Decode(input []byte, parameters Parameters, globals *document.Globals) (*Decoder, error) {
	r := reader.New(input)

	doc, err := document.DecodeDocument(r, globals)
	if err != nil {
		return nil, err
	}

	return &Decoder{inputReader: r, document: doc, parameters: parameters}, nil
}

// Parameters are the parameters used by the jbig2 decoder.
type Parameters struct {
	UnpaddedData bool
	Color        bitmap.Color
}
