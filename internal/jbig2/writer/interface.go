/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package writer

import "io"

// BitWriter is the interface that allows to write single bits.
type BitWriter interface {
	// WriteBit writes the 'bit' - {0,1} value to the writer.
	WriteBit(bit int) error
	// WriteBits writes 'number' of 'bits'.
	WriteBits(bits uint64, number int) (n int, err error)
	// FinishByte sets the bitIndex to the end of given byte. This resets the bitIndex to 0
	// and the byte index to the next byte.
	FinishByte()
	// SkipBits skips the 'skip' number of bits in the writer - changes the index position of the bit and byte.
	// The value -1 sets the bitIndex to the beginning of the byte.
	SkipBits(skip int) error
}

// BinaryWriter is the interface that implements writer.BitWriter, io.Writer and io.ByteWriter.
type BinaryWriter interface {
	BitWriter
	io.Writer
	io.ByteWriter
	Data() []byte
}
