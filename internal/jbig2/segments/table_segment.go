/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"fmt"

	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// TableSegment is the model used for user defined Huffman Table Segment - see 7.4.13 and appendix B..
type TableSegment struct {
	r reader.StreamReader

	// Code Table Flags B.2.1
	htOutOfBand int
	htPS        int
	htRS        int

	// Code Table lowest value B.2.2
	htLow int

	// Code Table highest value B.2.3
	htHight int
}

// compile time check if the TableSegment implements huffman.BasicTabler
var _ huffman.BasicTabler = &TableSegment{}

func (t *TableSegment) parseHeader() (err error) {
	var bit int
	bit, err = t.r.ReadBit()
	if err != nil {
		return
	}
	if bit == 1 {
		return fmt.Errorf("B.2.1 Code Table flags: Bit 7 must be zero. Was: %d", bit)
	}

	var bits uint64

	// Bit 4-6
	if bits, err = t.r.ReadBits(3); err != nil {
		return
	}
	t.htRS = (int(bits) + 1) & 0xf

	// Bit 1-3
	if bits, err = t.r.ReadBits(3); err != nil {
		return
	}
	t.htPS = (int(bits) + 1) & 0xf

	// 4 bytes
	if bits, err = t.r.ReadBits(32); err != nil {
		return
	}
	t.htLow = int(bits & 0xffffffff)

	// 4 bytes
	if bits, err = t.r.ReadBits(32); err != nil {
		return
	}
	t.htHight = int(bits & 0xffffffff)

	return
}

// Init initializes the table segment.
// Implements Segmenter interface.
func (t *TableSegment) Init(h *Header, r reader.StreamReader) error {
	t.r = r
	return t.parseHeader()
}

// HtPS returns the HtPs value.
// Implements huffman.BasicTabler interface.
func (t *TableSegment) HtPS() int {
	return t.htPS
}

// HtRS returns the table HtRs value.
// Implements huffman.BasicTabler interface.
func (t *TableSegment) HtRS() int {
	return t.htRS
}

// HtLow returns the table HtLow value.
// Implements huffman.BasicTabler interface.
func (t *TableSegment) HtLow() int {
	return t.htLow
}

// HtHigh return the table HtHigh value.
// Implements huffman.BasicTabler interface.
func (t *TableSegment) HtHigh() int {
	return t.htHight
}

// HtOOB returns the table HtoutOfBand value.
// Implements huffman.BasicTabler interface.
func (t *TableSegment) HtOOB() int {
	return t.htOutOfBand
}

// StreamReader returns the table segment stream reader.
// Implements huffman.BasicTabler interface.
func (t *TableSegment) StreamReader() reader.StreamReader {
	return t.r
}
