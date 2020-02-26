/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package huffman

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Compile time check for the encoded table.
var _ Tabler = &EncodedTable{}

// EncodedTable is a model for the encoded huffman table.
type EncodedTable struct {
	BasicTabler
	rootNode *InternalNode
}

// NewEncodedTable creates new encoded table from the provided 'table' argument.
func NewEncodedTable(table BasicTabler) (*EncodedTable, error) {
	e := &EncodedTable{
		rootNode:    &InternalNode{},
		BasicTabler: table,
	}

	if err := e.parseTable(); err != nil {
		return nil, err
	}

	return e, nil
}

// Decode implenets Node interface.
func (e *EncodedTable) Decode(r reader.StreamReader) (int64, error) {
	return e.rootNode.Decode(r)
}

// InitTree implements Tabler interface.
func (e *EncodedTable) InitTree(codeTable []*Code) error {
	preprocessCodes(codeTable)

	for _, c := range codeTable {
		if err := e.rootNode.append(c); err != nil {
			return err
		}
	}
	return nil
}

// RootNode Implements Tabler interface.
func (e *EncodedTable) RootNode() *InternalNode {
	return e.rootNode
}

// String implements Stringer interface.
func (e *EncodedTable) String() string {
	return e.rootNode.String() + "\n"
}

// parseTable parses the encoded table 'e' into BasicTabler.
func (e *EncodedTable) parseTable() error {
	var (
		codeTable                   []*Code
		prefLen, rangeLen, rangeLow int32
		temp                        uint64
		err                         error
	)

	r := e.StreamReader()
	curRangeLow := e.HtLow()

	// Annex B.2 5) - decode table lines.
	for curRangeLow < e.HtHigh() {
		temp, err = r.ReadBits(byte(e.HtPS()))
		if err != nil {
			return err
		}
		prefLen = int32(temp)

		temp, err = r.ReadBits(byte(e.HtRS()))
		if err != nil {
			return err
		}
		rangeLen = int32(temp)

		codeTable = append(codeTable, NewCode(prefLen, rangeLen, rangeLow, false))
		curRangeLow += 1 << uint(rangeLen)
	}

	// Annex B.2 6)
	temp, err = r.ReadBits(byte(e.HtPS()))
	if err != nil {
		return err
	}
	prefLen = int32(temp)

	// Annex B.2 7)
	rangeLen = 32
	rangeLow = e.HtLow() - 1
	codeTable = append(codeTable, NewCode(prefLen, rangeLen, rangeLow, true))

	// Annex B.2 8)
	temp, err = r.ReadBits(byte(e.HtPS()))
	if err != nil {
		return err
	}
	prefLen = int32(temp)

	// Annex B.2 9)
	rangeLen = 32
	rangeLow = e.HtHigh()
	codeTable = append(codeTable, NewCode(prefLen, rangeLen, rangeLow, false))

	//Annex B.2 10) OOB table line.
	if e.HtOOB() == 1 {
		temp, err = r.ReadBits(byte(e.HtPS()))
		if err != nil {
			return err
		}
		prefLen = int32(temp)
		codeTable = append(codeTable, NewCode(prefLen, -1, -1, false))
	}

	if err = e.InitTree(codeTable); err != nil {
		return err
	}
	return nil
}
