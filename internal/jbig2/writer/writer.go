/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package writer

import (
	"errors"
	"io"
)

// Writer is the structure used to write bits, bytes into predefined data.
// It allows to write the bits in two modes. The first and default
// writes bytes with the initial bitIndex 0 as the LSB (Least Significant Bit)
// The second mode writes bits in an opposite manner starting from the MSB (Most Significant Bit).
// The writer is being created by the methods: 'New' and 'NewMSB', where the first
// creates default writer and the second the 'msb' flagged writer.
// Implements io.Writer, io.ByteWriter interfaces.
type Writer struct {
	data      []byte
	bitIndex  uint8
	byteIndex int

	msb bool
}

var (
	// ErrInvalidBitValue defines the error when invalid 'bit' value is provided.
	ErrInvalidBitValue = errors.New("invalid bit value")
)

var (
	_ io.Writer     = &Writer{}
	_ io.ByteWriter = &Writer{}
)

// New creates new writer for the provided data.
func New(data []byte) *Writer {
	return &Writer{data: data}
}

// NewMSB creates new writer with the msb flag.
// While default writer writes single bits into LSB, the msbWriter writes single bits
// starting from the MSB.
// Example:
// 		InverseWriter contains following data:
//		data - 10010100 01001110 00000000
//							 	 ^
// 		The default current bit index is pointed by '^'.
// 		Writing new '1' bit to the following data would result as:
//		data - 10010100 01001110 10000000
func NewMSB(data []byte) *Writer {
	return &Writer{data: data, msb: true}
}

// Data gets the writer data.
func (w *Writer) Data() []byte {
	return w.data
}

// UseMSB gets the writer flag if it works on the MSB mode.
func (w *Writer) UseMSB() bool {
	return w.msb
}

// Write implements io.Writer interface.
func (w *Writer) Write(p []byte) (int, error) {
	if len(p) > w.byteCapacity() {
		return 0, io.EOF
	}

	for _, b := range p {
		if err := w.writeByte(b); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// WriteByte implements io.ByteWriter interface.
func (w *Writer) WriteByte(c byte) error {
	return w.writeByte(c)
}

// WriteBit writes single bit into provided bit writer data.
func (w *Writer) WriteBit(bit int) error {
	switch bit {
	case 0, 1:
		return w.writeBit(uint8(bit))
	}
	return ErrInvalidBitValue
}

func (w *Writer) byteCapacity() int {
	currentCapacity := len(w.data) - w.byteIndex
	if w.bitIndex != 0 {
		currentCapacity--
	}
	return currentCapacity
}

func (w *Writer) writeBit(b uint8) error {
	if len(w.data)-1 < w.byteIndex {
		return io.EOF
	}

	bitIndex := w.bitIndex
	if w.msb {
		bitIndex = 7 - w.bitIndex
	}

	w.data[w.byteIndex] |= byte(uint16(b<<bitIndex) & 0xff)
	w.bitIndex++

	if w.bitIndex == 8 {
		w.byteIndex++
		w.bitIndex = 0
	}
	return nil
}

func (w *Writer) writeByte(b byte) error {
	if w.byteIndex > len(w.data)-1 {
		return io.EOF
	}
	if w.byteIndex == len(w.data)-1 && w.bitIndex != 0 {
		return io.EOF
	}

	if w.bitIndex == 0 {
		w.data[w.byteIndex] = b
		w.byteIndex++
		return nil
	}

	if w.msb {
		w.data[w.byteIndex] |= b >> w.bitIndex
		w.byteIndex++
		w.data[w.byteIndex] = byte(uint16(b) << (8 - w.bitIndex) & 0xff)
	} else {
		w.data[w.byteIndex] |= byte(uint16(b) << w.bitIndex & 0xff)
		w.byteIndex++
		w.data[w.byteIndex] = b >> (8 - w.bitIndex)
	}
	return nil
}
