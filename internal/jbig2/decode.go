/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package jbig2

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// DecodeBytes decodes jbig2 'encode' byte slice data, with provided 'parameters' and optional 'globals'.
// The function decodes only a single page from the given input.
func DecodeBytes(encoded []byte, parameters decoder.Parameters, globals ...Globals) ([]byte, error) {
	var g Globals
	if len(globals) > 0 {
		g = globals[0]
	}
	d, err := decoder.Decode(encoded, parameters, g.ToDocumentGlobals())
	if err != nil {
		return nil, err
	}
	return d.DecodeNextPage()
}

// DecodeGlobals decodes globally defined data segments from the provided 'encoded' byte slice.
func DecodeGlobals(encoded []byte) (Globals, error) {
	const processName = "DecodeGlobals"
	r := reader.New(encoded)

	doc, err := document.DecodeDocument(r, nil)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	if doc.GlobalSegments == nil || (doc.GlobalSegments.Segments == nil) {
		return nil, errors.Error(processName, "no global segments found")
	}
	g := Globals{}
	for _, segment := range doc.GlobalSegments.Segments {
		g[int(segment.SegmentNumber)] = segment
	}
	return g, nil
}
