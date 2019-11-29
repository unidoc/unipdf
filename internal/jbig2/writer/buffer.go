/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package writer

import (
	"io"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

const (
	smallSize = 64
	maxInt    = int(^uint(0) >> 1)
)

// compile time check for Buffer to implement BinaryWriter interface.
var _ BinaryWriter = &Buffer{}

// Buffer is the Writer implementation that works on expandable data slices.
type Buffer struct {
	data      []byte
	bitIndex  uint8
	byteIndex int

	msb bool
}

// BufferedMSB creates a buffered writer with MSB bit writing method.
func BufferedMSB() *Buffer {
	return &Buffer{msb: true}
}

// Data gets the buffer byte slice data.
// The buffer is the owner of the byte slice data, thus the data is available until
// next buffer Reset.
func (b *Buffer) Data() []byte {
	return b.data
}

// FinishByte finishes current bit written byte and sets the current byte index pointe to the next byte.
func (b *Buffer) FinishByte() {
	if b.bitIndex == 0 {
		return
	}
	b.bitIndex = 0
	b.byteIndex++
}

// Len returns the number of bytes of the unwritten portion of the buffer.
func (b *Buffer) Len() int {
	return b.byteCapacity()
}

// Reset resets the data buffer as well as the bit and byte indexes.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
	b.byteIndex = 0
	b.bitIndex = 0
}

// ResetBitIndex resets the current bit index.
func (b *Buffer) ResetBitIndex() {
	b.bitIndex = 0
}

// SkipBits implements BitWriter interface.
func (b *Buffer) SkipBits(skip int) error {
	if skip == 0 {
		return nil
	}

	d := int(b.bitIndex) + skip
	if d >= 0 && d < 8 {
		// skip only bit index. The byte index is the same.
		b.bitIndex = uint8(d)
		return nil
	}

	// skip bit index as well as byte index.
	// The 'skip' value may be negative. Check if the summary bit index
	// is not lower than zero.
	d = int(b.bitIndex) + b.byteIndex*8 + skip
	if d < 0 {
		return errors.Errorf("Writer.SkipBits", "index out of range")
	}

	byteIndex := d / 8
	bitIndex := d % 8
	// common.Log.Trace("SkipBits")
	// common.Log.Trace("BitIndex: '%d' ByteIndex: '%d', FullBits: '%d', Len: '%d', Cap: '%d'", b.bitIndex, b.byteIndex, int(b.bitIndex)+(b.byteIndex)*8, len(b.data), cap(b.data))
	// common.Log.Trace("Skip: '%d', d: '%d', bitIndex: '%d'", skip, d, bitIndex)

	b.bitIndex = uint8(bitIndex)

	// expand if the the data doesnt have place for the given byte index
	if byteDiff := byteIndex - b.byteIndex; byteDiff > 0 && len(b.data)-1 < byteIndex {
		// 	common.Log.Trace("ByteDiff: %d", byteDiff)
		if b.bitIndex != 0 {
			byteDiff++
		}
		b.expandIfNeeded(byteDiff)
	}
	b.byteIndex = byteIndex

	// common.Log.Trace("BitIndex: '%d', ByteIndex: '%d'", b.bitIndex, b.byteIndex)
	return nil
}

// compile time check for the io.Writer interface.
var _ io.Writer = &Buffer{}

// Write implements io.Writer interface.
func (b *Buffer) Write(d []byte) (int, error) {
	b.expandIfNeeded(len(d))
	if b.bitIndex == 0 {
		return b.writeFullBytes(d), nil
	}
	return b.writeShiftedBytes(d), nil
}

// compile time check for the io.ByteWriter interface.
var _ io.ByteWriter = &Buffer{}

// WriteByte implements io.ByteWriter.
func (b *Buffer) WriteByte(bt byte) error {
	if b.byteIndex > len(b.data)-1 || (b.byteIndex == len(b.data)-1 && b.bitIndex != 0) {
		b.expandIfNeeded(1)
	}
	b.writeByte(bt)
	return nil
}

// WriteBit implements BitWriter interface.
func (b *Buffer) WriteBit(bit int) error {
	if bit != 1 && bit != 0 {
		return errors.Errorf("Buffer.WriteBit", "bit value must be in range {0,1} but is: %d", bit)
	}

	if len(b.data)-1 < b.byteIndex {
		b.expandIfNeeded(1)
	}

	bitIndex := b.bitIndex
	if b.msb {
		bitIndex = 7 - b.bitIndex
	}

	b.data[b.byteIndex] |= byte(uint16(bit<<bitIndex) & 0xff)
	b.bitIndex++

	if b.bitIndex == 8 {
		b.byteIndex++
		b.bitIndex = 0
	}
	return nil
}

func (b *Buffer) byteCapacity() int {
	currentCapacity := len(b.data) - b.byteIndex
	if b.bitIndex != 0 {
		currentCapacity--
	}
	return currentCapacity
}

func (b *Buffer) expandIfNeeded(n int) {
	if !b.tryGrowByReslice(n) {
		b.grow(n)
	}
}

func (b *Buffer) fullOffset() int {
	off := b.byteIndex
	if b.bitIndex != 0 {
		off++
	}
	return off
}

func (b *Buffer) grow(n int) {
	if b.data == nil && n < smallSize {
		b.data = make([]byte, n, smallSize)
		return
	}

	m := len(b.data)
	if b.bitIndex != 0 {
		m++
	}
	c := cap(b.data)
	// common.Log.Trace("data before grow: %v", b.data)
	if n <= c/2-m {
		// reslice only
		common.Log.Trace("[Buffer] grow - reslice only. Len: '%d', Cap: '%d', n: '%d'", len(b.data), cap(b.data), n)
		common.Log.Trace(" n <= c / 2 -m. C: '%d', m: '%d'", c, m)
		copy(b.data, b.data[b.fullOffset():])
	} else if c > maxInt-c-n {
		panic("buffer too large")
	} else {
		// not enough space, need to allocate new slice.
		// common.Log.Trace("[Buffer] grow - allocate new slice. Len: '%d', Cap: '%d', N: '%d'", len(b.data), cap(b.data), n)
		buf := make([]byte, 2*c+n)
		copy(buf, b.data)
		b.data = buf
	}
	//	common.Log.Trace("data after grow: %v", b.data)
	//	common.Log.Trace("[Buffer] b.data[:m+n]. m: '%d', n:'%d'", m, n)
	b.data = b.data[:m+n]
	//	common.Log.Trace("[Buffer] grown data. Len: '%d', Cap: '%d'", len(b.data), cap(b.data))
	//	common.Log.Trace("[Buffer] b.data after grow: '%v'", b.data)
}

func (b *Buffer) writeByte(bt byte) {
	if b.bitIndex == 0 {
		b.data[b.byteIndex] = bt
		b.byteIndex++
	} else if b.msb {
		b.data[b.byteIndex] |= bt >> b.bitIndex
		b.byteIndex++
		b.data[b.byteIndex] = byte(uint16(bt) << (8 - b.bitIndex) & 0xff)
	} else {
		b.data[b.byteIndex] |= byte(uint16(bt) << b.bitIndex & 0xff)
		b.byteIndex++
		b.data[b.byteIndex] = bt >> (8 - b.bitIndex)
	}
}

func (b *Buffer) writeFullBytes(d []byte) int {
	copied := copy(b.data[b.fullOffset():], d)
	b.byteIndex += copied
	return copied
}

func (b *Buffer) writeShiftedBytes(d []byte) int {
	for _, bt := range d {
		b.writeByte(bt)
	}
	return len(d)
}
func (b *Buffer) tryGrowByReslice(n int) bool {
	if l := len(b.data); n <= cap(b.data)-l {
		// 	common.Log.Trace("[Buffer] Grow by reslice. Len: '%d', Cap: '%d', Additional space: '%d'", len(b.data), cap(b.data), n)
		b.data = b.data[:l+n]
		// 	common.Log.Trace("[Buffer] Current data. Len: '%d', Cap: '%d'", len(b.data), cap(b.data))
		return true
	}
	return false
}
