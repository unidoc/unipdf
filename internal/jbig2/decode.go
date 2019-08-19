/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package jbig2

import (
	"errors"

	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// DecodeBytes decodes jbig2 'encode' byte slice data, with provided 'parameters' and optional 'globals'.
// The function decodes only a single page from the given input.
func DecodeBytes(encoded []byte, parameters decoder.Parameters, globals ...document.Globals) ([]byte, error) {
	d, err := decoder.Decode(encoded, parameters, globals...)
	if err != nil {
		return nil, err
	}
	return d.DecodeNextPage()
}

// DecodeGlobals decodes globally defined data segments from the provided 'encoded' byte slice.
func DecodeGlobals(encoded []byte) (document.Globals, error) {
	r := reader.New(encoded)

	doc, err := document.DecodeDocument(r)
	if err != nil {
		return nil, err
	}

	if doc.GlobalSegments == nil {
		return nil, errors.New("no global segments found")
	}
	return doc.GlobalSegments, nil
}
