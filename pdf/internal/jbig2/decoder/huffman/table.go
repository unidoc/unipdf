/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package huffman

import (
	"fmt"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"strings"
)

// HuffmanTabler is the interface for all types of the huffman tables
type HuffmanTabler interface {
	Decode(r reader.StreamReader) (int64, error)
	InitTree(codeTable []*Code) error
	String() string
	RootNode() *InternalNode
}

// Code is the model for the huffman table code
type Code struct {
	prefixLength int
	rangeLength  int
	rangeLow     int
	isLowerRange bool
	code         int
}

// String implements Stringer interface
// returns stringified Code
func (c *Code) String() string {
	var temp string
	if c.code != -1 {
		temp = bitPattern(c.code, c.prefixLength)
	} else {
		temp = "?"
	}
	return fmt.Sprintf("%s/%d/%d/%d", temp, c.prefixLength, c.rangeLength, c.rangeLow)
}

// NewCode creates new huffman code
func NewCode(prefixLength, rangeLength, rangeLow int, isLowerRange bool) *Code {
	return &Code{
		prefixLength: prefixLength,
		rangeLength:  rangeLength,
		rangeLow:     rangeLow,
		isLowerRange: isLowerRange,
		code:         -1,
	}
}

func bitPattern(v, l int) string {
	var result = make([]rune, l)
	var temp int
	for i := 1; i <= l; i++ {
		temp = (v >> uint(l-i) & 1)
		if temp != 0 {
			result[i-1] = '1'
		} else {
			result[i-1] = '0'
		}
	}
	return string(result)
}

func preprocessCodes(codeTable []*Code) {
	// Annex B.3 1) build the histogram
	var maxPrefixLength int

	for _, c := range codeTable {
		maxPrefixLength = maxInt(maxPrefixLength, c.prefixLength)
	}

	var lenCount = make([]int, maxPrefixLength+1)

	for _, c := range codeTable {
		lenCount[c.prefixLength]++
	}

	var (
		curCode   int
		firstCode = make([]int, len(lenCount)+1)
	)
	lenCount[0] = 0

	// Annex B.3 3)
	for curLen := 1; curLen <= len(lenCount); curLen++ {
		// common.Log.Debug("First: %d, Lencount-1: %d, LenCount-1 shifted: %d", firstCode[curLen-1], lenCount[curLen-1], lenCount[curLen-1]<<1)
		firstCode[curLen] = (firstCode[curLen-1] + (lenCount[curLen-1])) << 1
		curCode = firstCode[curLen]
		// common.Log.Debug("CurCode %d at i: %d", curCode, curLen)
		for _, c := range codeTable {
			if c.prefixLength == curLen {
				c.code = curCode
				curCode++
			}
		}
	}
	// common.Log.Debug("Table: %v", codeTable)
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func codeTableToString(codeTable []*Code) string {
	sb := strings.Builder{}
	for _, c := range codeTable {
		sb.WriteString(c.String() + "\n")
	}
	return sb.String()
}
