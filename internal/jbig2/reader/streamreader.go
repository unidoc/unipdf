/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package reader

// StreamReader is the interface that allows to read bit, bits, byte, bytes change and get the
// stream position, align the bits.
// Implements io.Reader, io.Seeker interfaces
type StreamReader interface {
	// Align resets the bits position of the given reader.
	// It returns how many bits left were skipped.
	Align() byte

	// BitPosition gets the current bit position.
	BitPosition() int

	// Mark marks a position in the stream to be returned to by a subsequent call to 'Reset'.
	Mark()

	// Length returns the length of the total data used by the reader.
	Length() uint64

	// Read reads the bytes of the provided data length and stores them inside the data slice.
	Read(b []byte) (int, error)

	// ReadBit reads the next binary value from the current cache.
	// Equivalent of ReadBool method but returns an integer.
	ReadBit() (int, error)

	// ReadBits reads the bits of size 'n' from the reader.
	ReadBits(n byte) (uint64, error)

	// ReadBool reads the next binary value from the current cache.
	ReadBool() (bool, error)

	// ReadByte implements io.ByteReader.
	ReadByte() (byte, error)

	// Reset returns the stream pointer to its previous position, including the bit offset,
	// at the time of the most recent unmatched call to mark.
	Reset()

	// Seek implements io.Seek
	Seek(offset int64, whence int) (int64, error)

	// StreamPosition gets the stream position of the stream reader.
	StreamPosition() int64
}
