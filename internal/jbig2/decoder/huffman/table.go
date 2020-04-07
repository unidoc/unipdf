/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package huffman

import (
	"fmt"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Tabler is the interface for all types of the huffman tables.
type Tabler interface {
	Decode(r reader.StreamReader) (int64, error)
	InitTree(codeTable []*Code) error
	String() string
	RootNode() *InternalNode
}

// BasicTabler is the interface common for the huffman tables.
type BasicTabler interface {
	HtHigh() int32
	HtLow() int32
	StreamReader() reader.StreamReader
	HtPS() int32
	HtRS() int32
	HtOOB() int32
}

// Code is the model for the huffman table code.
type Code struct {
	prefixLength int32
	rangeLength  int32
	rangeLow     int32
	isLowerRange bool
	code         int32
}

// String implements Stringer interface.
func (c *Code) String() string {
	var temp string
	if c.code != -1 {
		temp = bitPattern(c.code, c.prefixLength)
	} else {
		temp = "?"
	}
	return fmt.Sprintf("%s/%d/%d/%d", temp, c.prefixLength, c.rangeLength, c.rangeLow)
}

// NewCode creates new huffman code.
func NewCode(prefixLength, rangeLength, rangeLow int32, isLowerRange bool) *Code {
	return &Code{
		prefixLength: prefixLength,
		rangeLength:  rangeLength,
		rangeLow:     rangeLow,
		isLowerRange: isLowerRange,
		code:         -1,
	}
}

func bitPattern(v, l int32) string {
	var temp int32
	result := make([]rune, l)

	for i := int32(1); i <= l; i++ {
		temp = v >> uint(l-i) & 1
		if temp != 0 {
			result[i-1] = '1'
		} else {
			result[i-1] = '0'
		}
	}
	return string(result)
}

func preprocessCodes(codeTable []*Code) {
	// Annex B.3 1) build the histogram.
	var maxPrefixLength int32

	for _, c := range codeTable {
		maxPrefixLength = maxInt(maxPrefixLength, c.prefixLength)
	}

	lenCount := make([]int32, maxPrefixLength+1)

	for _, c := range codeTable {
		lenCount[c.prefixLength]++
	}

	var curCode int32
	firstCode := make([]int32, len(lenCount)+1)
	lenCount[0] = 0

	// Annex B.3 3)
	for curLen := int32(1); curLen <= int32(len(lenCount)); curLen++ {
		firstCode[curLen] = (firstCode[curLen-1] + (lenCount[curLen-1])) << 1
		curCode = firstCode[curLen]

		for _, c := range codeTable {
			if c.prefixLength == curLen {
				c.code = curCode
				curCode++
			}
		}
	}
}

func maxInt(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}
