package jbig2

import (
	"github.com/pkg/errors"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

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

var _ huffman.Tabler = &TableSegment{}

func (t *TableSegment) parseHeader() (err error) {
	var bit int
	bit, err = t.r.ReadBit()
	if err != nil {
		return
	}
	if bit == 1 {
		return errors.Errorf("B.2.1 Code Table flags: Bit 7 must be zero. Was: %d", bit)
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

func (t *TableSegment) Init(h *SegmentHeader, r reader.StreamReader) error {
	t.r = r
	return t.parseHeader()
}

// Getters
func (t *TableSegment) HtPS() int {
	return t.htPS
}

func (t *TableSegment) HtRS() int {
	return t.htRS
}

func (t *TableSegment) HtLow() int {
	return t.htLow
}

func (t *TableSegment) HtHigh() int {
	return t.htHight
}

func (t *TableSegment) HtOOB() int {
	return t.htOutOfBand
}

func (t *TableSegment) StreamReader() reader.StreamReader {
	return t.r
}
