/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitwise

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

// StreamReader is the interface that allows to read bit, bits, byte,
// bytes, integers change and get the stream position, align the bits.
type StreamReader interface {
	io.Reader
	io.ByteReader
	io.Seeker

	// Align resets the bits position of the given reader.
	// It returns how many bits left were skipped.
	Align() byte
	// BitPosition gets the current bit position.
	BitPosition() int
	// Mark marks a position in the stream to be returned to by a subsequent call to 'Reset'.
	Mark()
	// Length returns the length of the total data used by the reader.
	Length() uint64

	// ReadBit reads the next binary value from the current cache.
	// Equivalent of ReadBool method but returns an integer.
	ReadBit() (int, error)
	// ReadBits reads the bits of size 'n' from the reader.
	ReadBits(n byte) (uint64, error)
	// ReadBool reads the next binary value from the current cache.
	ReadBool() (bool, error)
	// ReadUint32 reads the unsigned uint32 from the reader.
	ReadUint32() (uint32, error)

	// Reset returns the stream pointer to its previous position, including the bit offset,
	// at the time of the most recent unmatched call to mark.
	Reset()
	// StreamPosition gets the stream position of the stream reader.
	StreamPosition() int64
}
