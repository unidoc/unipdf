/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package huffman

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Node is the interface defined for all huffman tree nodes.
type Node interface {
	Decode(r reader.StreamReader) (int64, error)
	String() string
}

// OutOfBandNode represents an out of band node in a huffman tree.
type OutOfBandNode struct{}

// Compile time check the OutOfBandNode.
var _ Node = &OutOfBandNode{}

// Decode implements Node interface.
// Decodes the out of band node by returning max int64 value.
func (o *OutOfBandNode) Decode(r reader.StreamReader) (int64, error) {
	return int64(math.MaxInt64), nil
}

// String implements the Stringer interface returns the max int binary value.
func (o *OutOfBandNode) String() string {
	return fmt.Sprintf("%064b", int64(math.MaxInt64))
}

func newOufOfBandNode(c *Code) *OutOfBandNode {
	return &OutOfBandNode{}
}

// ValueNode represents a value node in a huffman tree. It is a leaf of a tree.
type ValueNode struct {
	rangeLen     int32
	rangeLow     int32
	isLowerRange bool
}

// Compile time check the ValueNode.
var _ Node = &ValueNode{}

// Decode implements Node interface.
func (v *ValueNode) Decode(r reader.StreamReader) (int64, error) {
	bits, err := r.ReadBits(byte(v.rangeLen))
	if err != nil {
		return 0, err
	}

	if v.isLowerRange {
		// B.4 4)
		bits = -bits
	}
	return int64(v.rangeLow) + int64(bits), nil
}

// String implements Stringer interface.
func (v *ValueNode) String() string {
	return fmt.Sprintf("%d/%d", v.rangeLen, v.rangeLow)
}

func newValueNode(c *Code) *ValueNode {
	return &ValueNode{
		rangeLen:     c.rangeLength,
		rangeLow:     c.rangeLow,
		isLowerRange: c.isLowerRange,
	}
}

// InternalNode represents an internal node of a huffman tree.
// It is defined as a pair of  two child nodes 'zero' and 'one' and a 'depth' level.
// Implements Node interface.
type InternalNode struct {
	depth int32
	zero  Node
	one   Node
}

// Compile time check for the InternalNode.
var _ Node = &InternalNode{}

// Decode implements Node interface.
func (i *InternalNode) Decode(r reader.StreamReader) (int64, error) {
	b, err := r.ReadBit()
	if err != nil {
		return 0, err
	}

	if b == 0 {
		return i.zero.Decode(r)
	}
	return i.one.Decode(r)
}

// String implements the Stringer interface.
func (i *InternalNode) String() string {
	b := &strings.Builder{}

	b.WriteString("\n")
	i.pad(b)
	b.WriteString("0: ")
	b.WriteString(i.zero.String() + "\n")
	i.pad(b)
	b.WriteString("1: ")
	b.WriteString(i.one.String() + "\n")
	return b.String()
}

func (i *InternalNode) append(c *Code) (err error) {
	// ignore unused codes
	if c.prefixLength == 0 {
		return nil
	}
	shift := c.prefixLength - 1 - i.depth

	if shift < 0 {
		return errors.New("Negative shifting is not allowed")
	}

	bit := (c.code >> uint(shift)) & 0x1
	if shift == 0 {
		if c.rangeLength == -1 {
			// the child will be OutOfBand
			if bit == 1 {
				if i.one != nil {
					return fmt.Errorf("OOB already set for code %s", c)
				}
				i.one = newOufOfBandNode(c)
			} else {
				if i.zero != nil {
					return fmt.Errorf("OOB already set for code %s", c)
				}
				i.zero = newOufOfBandNode(c)
			}
		} else {
			// the child will be a ValueNode
			if bit == 1 {
				if i.one != nil {
					return fmt.Errorf("Value Node already set for code %s", c)
				}
				i.one = newValueNode(c)
			} else {
				if i.zero != nil {
					return fmt.Errorf("Value Node already set for code %s", c)
				}
				i.zero = newValueNode(c)
			}
		}
	} else {
		// the child will be an Internal Node
		if bit == 1 {
			if i.one == nil {
				i.one = newInternalNode(i.depth + 1)
			}
			if err = i.one.(*InternalNode).append(c); err != nil {
				return err
			}
		} else {
			if i.zero == nil {
				i.zero = newInternalNode(i.depth + 1)
			}
			if err = i.zero.(*InternalNode).append(c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *InternalNode) pad(sb *strings.Builder) {
	for j := int32(0); j < i.depth; j++ {
		sb.WriteString("   ")
	}
}

// newInternalNode creates new internal node.
func newInternalNode(depth int32) *InternalNode {
	return &InternalNode{depth: depth}
}
