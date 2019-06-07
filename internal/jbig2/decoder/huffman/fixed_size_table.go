/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package huffman

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// FixedSizeTable defines the table with the fixed size.
type FixedSizeTable struct {
	rootNode *InternalNode
}

// NewFixedSizeTable creates new fixedSizeTable.
func NewFixedSizeTable(codeTable []*Code) (*FixedSizeTable, error) {
	f := &FixedSizeTable{
		rootNode: &InternalNode{},
	}

	if err := f.InitTree(codeTable); err != nil {
		return nil, err
	}

	return f, nil
}

// Decode implements Tabler interface.
func (f *FixedSizeTable) Decode(r reader.StreamReader) (int64, error) {
	return f.rootNode.Decode(r)
}

// InitTree implements Tabler interface.
func (f *FixedSizeTable) InitTree(codeTable []*Code) error {
	preprocessCodes(codeTable)
	for _, c := range codeTable {
		err := f.rootNode.append(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// String implements Stringer interface.
func (f *FixedSizeTable) String() string {
	return f.rootNode.String() + "\n"
}

// RootNode implements Tabler interface.
func (f *FixedSizeTable) RootNode() *InternalNode {
	return f.rootNode
}
