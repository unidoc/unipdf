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

// TableSegment is the model used for user defined Huffman Table Segment - see 7.4.13 and appendix B.
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

// Compile time check if the TableSegment implements huffman.BasicTabler.
var _ huffman.BasicTabler = &TableSegment{}

// Init initializes the table segment.
// Implements Segmenter interface.
func (t *TableSegment) Init(h *Header, r reader.StreamReader) error {
	t.r = r
	return t.parseHeader()
}

// HtPS implements huffman.BasicTabler interface.
func (t *TableSegment) HtPS() int {
	return t.htPS
}

// HtRS implements huffman.BasicTabler interface.
func (t *TableSegment) HtRS() int {
	return t.htRS
}

// HtLow implements huffman.BasicTabler interface.
func (t *TableSegment) HtLow() int {
	return t.htLow
}

// HtHigh implements huffman.BasicTabler interface.
func (t *TableSegment) HtHigh() int {
	return t.htHight
}

// HtOOB implements huffman.BasicTabler interface.
func (t *TableSegment) HtOOB() int {
	return t.htOutOfBand
}

// StreamReader implements huffman.BasicTabler interface.
func (t *TableSegment) StreamReader() reader.StreamReader {
	return t.r
}

func (t *TableSegment) parseHeader() (err error) {
	var (
		bit  int
		bits uint64
	)

	bit, err = t.r.ReadBit()
	if err != nil {
		return
	}
	if bit == 1 {
		return fmt.Errorf("invalid table segment definition. B.2.1 Code Table flags: Bit 7 must be zero. Was: %d", bit)
	}

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
