/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package huffman

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

var _ HuffmanTabler = &EncodedTable{}

// EncodedTable is a model for the encoded huffman table
type EncodedTable struct {
	Tabler
	rootNode *InternalNode
}

// NewEncodedTable creates new encoded table for the provided Tabler
func NewEncodedTable(table Tabler) (*EncodedTable, error) {
	e := &EncodedTable{
		rootNode: &InternalNode{},
		Tabler:   table,
	}

	if err := e.ParseTable(); err != nil {
		return nil, err
	}

	return e, nil
}

// Decode decodes the provided root node
func (e *EncodedTable) Decode(r reader.StreamReader) (int64, error) {
	return e.rootNode.Decode(r)
}

// ParseTable parses the Tabler
func (e *EncodedTable) ParseTable() (err error) {
	r := e.StreamReader()

	var codeTable []*Code

	var prefLen, rangeLen, rangeLow int
	curRangeLow := e.HtLow()

	var temp uint64

	// Annex B.2 5) - decode table lines
	for curRangeLow < e.HtHigh() {
		temp, err = r.ReadBits(byte(e.HtPS()))
		if err != nil {
			return
		}
		prefLen = int(temp)
		temp, err = r.ReadBits(byte(e.HtRS()))
		if err != nil {
			return
		}
		rangeLen = int(temp)

		codeTable = append(codeTable, NewCode(prefLen, rangeLen, rangeLow, false))

		curRangeLow += (1 << uint(rangeLen))
	}

	// Annex B.2 6)
	temp, err = r.ReadBits(byte(e.HtPS()))
	if err != nil {
		return
	}
	prefLen = int(temp)

	// Annex B.2 7)
	rangeLen = 32
	rangeLow = e.HtLow() - 1
	codeTable = append(codeTable, NewCode(prefLen, rangeLen, rangeLow, true))

	// Annex B.2 8)
	temp, err = r.ReadBits(byte(e.HtPS()))
	if err != nil {
		return
	}
	prefLen = int(temp)

	// Annex B.2 9)
	rangeLen = 32
	rangeLow = e.HtHigh()
	codeTable = append(codeTable, NewCode(prefLen, rangeLen, rangeLow, false))

	//Annex B.2 10) oob table line
	if e.HtOOB() == 1 {
		temp, err = r.ReadBits(byte(e.HtPS()))
		if err != nil {
			return
		}
		prefLen = int(temp)
		codeTable = append(codeTable, NewCode(prefLen, -1, -1, false))
	}

	if err = e.InitTree(codeTable); err != nil {
		return
	}

	return nil
}

// RootNode returns the EncodedTable root node
func (e *EncodedTable) RootNode() *InternalNode {
	return e.rootNode
}

// InitTree implements HuffmanTabler interface
func (e *EncodedTable) InitTree(codeTable []*Code) error {
	preprocessCodes(codeTable)

	for _, c := range codeTable {
		if err := e.rootNode.append(c); err != nil {
			return err
		}
	}
	return nil
}

// String implements Stringer interface
func (e *EncodedTable) String() string {
	return e.rootNode.String() + "\n"
}

// func NewEncodedTable(table Tabler)

// Tabler is the interface common for the tables
type Tabler interface {
	HtHigh() int
	HtLow() int
	StreamReader() reader.StreamReader
	HtPS() int
	HtRS() int
	HtOOB() int
}
