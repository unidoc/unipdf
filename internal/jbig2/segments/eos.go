/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"math"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// EndOfStripe flags an end of stripe - see 7.4.9.
type EndOfStripe struct {
	r          reader.StreamReader
	lineNumber int
}

// Init implements Segmenter interface.
func (e *EndOfStripe) Init(h *Header, r reader.StreamReader) error {
	e.r = r
	return e.parseHeader(h, r)
}

// LineNumber gets the EndOfStripe line number.
func (e *EndOfStripe) LineNumber() int {
	return e.lineNumber
}

func (e *EndOfStripe) parseHeader(h *Header, r reader.StreamReader) error {
	temp, err := e.r.ReadBits(32)
	if err != nil {
		return err
	}
	e.lineNumber = int(temp & math.MaxInt32)
	return nil
}
