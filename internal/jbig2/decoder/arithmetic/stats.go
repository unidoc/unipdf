/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"fmt"
	"strings"
)

// DecoderStats is the structure that contains arithmetic decode context.
type DecoderStats struct {
	index              int32
	contextSize        int32
	codingContextTable []byte
	mps                []byte
}

// NewStats creates new DecoderStats of size 'contextSize'.
func NewStats(contextSize int32, index int32) *DecoderStats {
	return &DecoderStats{
		index:              index,
		contextSize:        contextSize,
		codingContextTable: make([]byte, contextSize),
		mps:                make([]byte, contextSize),
	}
}

// Copy copies the DecoderStats.
func (d *DecoderStats) Copy() *DecoderStats {
	cp := &DecoderStats{
		contextSize:        d.contextSize,
		codingContextTable: make([]byte, d.contextSize),
	}

	for i := 0; i < len(d.codingContextTable); i++ {
		cp.codingContextTable[i] = d.codingContextTable[i]
	}

	return cp
}

// Overwrite overwrites the codingContextTable from new DecoderStats 'dNew'.
func (d *DecoderStats) Overwrite(dNew *DecoderStats) {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = dNew.codingContextTable[i]
		d.mps[i] = dNew.mps[i]
	}
}

// Reset resets current decoder stats.
func (d *DecoderStats) Reset() {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = 0
		d.mps[i] = 0
	}
}

// SetIndex sets current decoder stats 'index'.
func (d *DecoderStats) SetIndex(index int32) {
	d.index = index
}

// String implements Stringer interface.
func (d *DecoderStats) String() string {
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("Stats:  %d\n", len(d.codingContextTable)))
	for i, v := range d.codingContextTable {
		if v != 0 {
			b.WriteString(fmt.Sprintf("Not zero at: %d - %d\n", i, v))
		}
	}
	return b.String()
}

func (d *DecoderStats) cx() byte {
	return d.codingContextTable[d.index]
}

func (d *DecoderStats) getMps() byte {
	return d.mps[d.index]
}

// setEntry sets the decoder stats coding context table with moreprobableSymbol.
func (d *DecoderStats) setEntry(value int) {
	v := byte(value & 0x7f)
	d.codingContextTable[d.index] = v
}

// toggleMps flips the bit in the actual more predictable index.
func (d *DecoderStats) toggleMps() {
	d.mps[d.index] ^= 1
}
