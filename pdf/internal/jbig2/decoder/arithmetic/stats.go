/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"fmt"
	"strings"
)

// DecoderStats is the structure that contains arithmetic
// decoder stats
type DecoderStats struct {
	index              int
	contextSize        int
	codingContextTable []byte
	mps                []byte
}

// NewStats creates new DecoderStats of size 'contextSize'
func NewStats(contextSize int, index int) *DecoderStats {
	d := &DecoderStats{
		index:              index,
		contextSize:        contextSize,
		codingContextTable: make([]byte, contextSize),
		mps:                make([]byte, contextSize),
	}

	return d
}

// Reset resets current decoder stats
func (d *DecoderStats) Reset() {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = 0
		d.mps[i] = 0
	}
}

// SetEntry sets the decoder stats coding context table with moreprobableSymbol
func (d *DecoderStats) SetEntry(value int) {
	v := byte(value & 0x7f)
	// common.Log.Debug("setCX for index: %d and value: %d and v: %d", d.index, value, v)
	d.codingContextTable[d.index] = v
}

// SetIndex sets the decoderStats index
func (d *DecoderStats) SetIndex(index int) {
	// common.Log.Debug("Setting index: %32b", index)
	d.index = int(uint(index))
}

// Overwrite overwrites the codingContextTable from new DecoderStats
func (d *DecoderStats) Overwrite(dNew *DecoderStats) {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = dNew.codingContextTable[i]
		d.mps[i] = dNew.mps[i]
	}
}

// flips the bit in the actual more predictable index
func (d *DecoderStats) toggleMps() {
	// common.Log.Debug("Before: %d", d.mps[d.index])
	d.mps[d.index] ^= 1
	// common.Log.Debug("After: %d", d.mps[d.index])
}

func (d *DecoderStats) getMps() byte {
	return d.mps[d.index]
}

// Copy copies the DecoderStats
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

func (d *DecoderStats) cx() byte {
	return d.codingContextTable[d.index]
}

// String implements Stringer interface
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
