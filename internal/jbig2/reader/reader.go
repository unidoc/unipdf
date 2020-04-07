/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package reader

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/unidoc/unipdf/v3/common"
)

// Reader is the bit reader implementation.
// Implements io.Reader, io.ByteReader, io.Seeker interfaces.
type Reader struct {
	in           []byte
	cache        byte  // unread bits are stored here
	bits         byte  // number of unread bits in cache
	r            int64 // buf read positions
	lastByte     int
	lastRuneSize int

	mark     int64
	markBits byte
}

// compile time checks for the interface implementation of the Reader.
var (
	_ io.Reader     = &Reader{}
	_ io.ByteReader = &Reader{}
	_ io.Seeker     = &Reader{}
	_ StreamReader  = &Reader{}
)

// New creates a new reader.Reader using the byte slice data as input.
func New(data []byte) *Reader {
	return &Reader{in: data}
}

// Align implements StreamReader interface.
func (r *Reader) Align() (skipped byte) {
	skipped = r.bits
	r.bits = 0 // no need to clear cache - it would be overwritten on next read.
	return skipped
}

// ConsumeRemainingBits consumes the remaining bits from the given reader.
func (r *Reader) ConsumeRemainingBits() {
	if r.bits != 0 {
		_, err := r.ReadBits(r.bits)
		if err != nil {
			common.Log.Debug("ConsumeRemainigBits failed: %v", err)
		}
	}
}

// BitPosition implements StreamReader inteface.
func (r *Reader) BitPosition() int {
	return int(r.bits)
}

// Length implements StreamReader interface.
func (r *Reader) Length() uint64 {
	return uint64(len(r.in))
}

// Mark implements StreamReader interface.
func (r *Reader) Mark() {
	r.mark = r.r
	r.markBits = r.bits
}

// Read implements io.Reader interface.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.bits == 0 {
		return r.read(p)
	}

	for ; n < len(p); n++ {
		if p[n], err = r.readUnalignedByte(); err != nil {
			return 0, err
		}
	}
	return n, nil
}

// ReadBit implements StreamReader interface.
func (r *Reader) ReadBit() (bit int, err error) {
	boolean, err := r.readBool()
	if err != nil {
		return 0, err
	}

	if boolean {
		bit = 1
	}
	return bit, nil
}

// ReadBits implements StreamReader interface.
func (r *Reader) ReadBits(n byte) (u uint64, err error) {
	// Frequent optimization.
	if n < r.bits {
		// cache has all needed bits, there are also some extra which will be left in cache.
		shift := r.bits - n
		u = uint64(r.cache >> shift)
		r.cache &= 1<<shift - 1
		r.bits = shift
		return u, nil
	}

	if n > r.bits {
		if r.bits > 0 {
			u = uint64(r.cache)
			n -= r.bits
		}

		// Read whole bytes.
		for n >= 8 {
			b, err := r.readBufferByte()
			if err != nil {
				return 0, err
			}
			u = u<<8 + uint64(b)
			n -= 8
		}

		// Read last fraction if exists.
		if n > 0 {
			if r.cache, err = r.readBufferByte(); err != nil {
				return 0, err
			}
			shift := 8 - n
			u = u<<n + uint64(r.cache>>shift)
			r.cache &= 1<<shift - 1
			r.bits = shift
		} else {
			r.bits = 0
		}
		return u, nil
	}

	r.bits = 0 // no need to clear cache, will be overridden on next read
	return uint64(r.cache), nil
}

// ReadBool implements StreamReader interface.
func (r *Reader) ReadBool() (bool, error) {
	return r.readBool()
}

// ReadByte implements io.ByteReader.
func (r *Reader) ReadByte() (byte, error) {
	// r.bits will be the same after reading 8 bits, so we don't need to update that.
	if r.bits == 0 {
		return r.readBufferByte()
	}
	return r.readUnalignedByte()
}

// ReadUint32 implements StreamReader interface.
func (r *Reader) ReadUint32() (uint32, error) {
	ub := make([]byte, 4)

	_, err := r.Read(ub)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(ub), nil
}

// Reset implements StreamReader interface.
func (r *Reader) Reset() {
	r.r = r.mark
	r.bits = r.markBits
}

// Seek implements the io.Seeker interface.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.lastRuneSize = -1
	var abs int64

	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.r + offset
	case io.SeekEnd:
		abs = int64(len(r.in)) + offset
	default:
		return 0, errors.New("reader.Reader.Seek: invalid whence")
	}

	if abs < 0 {
		return 0, errors.New("reader.Reader.Seek: negative position")
	}
	r.r = abs
	r.bits = 0
	return abs, nil
}

// StreamPosition implements StreamReader interface.
func (r *Reader) StreamPosition() int64 {
	return r.r
}

func (r *Reader) read(p []byte) (int, error) {
	if r.r >= int64(len(r.in)) {
		return 0, io.EOF
	}

	r.lastRuneSize = -1
	n := copy(p, r.in[r.r:])
	r.r += int64(n)
	return n, nil
}

func (r *Reader) readBufferByte() (byte, error) {
	if r.r >= int64(len(r.in)) {
		return 0, io.EOF
	}
	r.lastRuneSize = -1
	c := r.in[r.r]
	r.r++
	r.lastByte = int(c)
	return c, nil
}

// readUnalignedByte reads the next 8 bits which are (may be) unaligned and returns them as a byte.
func (r *Reader) readUnalignedByte() (b byte, err error) {
	// r.bits will be the same after reading 8 bits, so we don't need to update that.
	bits := r.bits
	b = r.cache << (8 - bits)
	r.cache, err = r.readBufferByte()
	if err != nil {
		return 0, err
	}
	b |= r.cache >> bits
	r.cache &= 1<<bits - 1
	return b, nil
}

func (r *Reader) readBool() (bit bool, err error) {
	if r.bits == 0 {
		r.cache, err = r.readBufferByte()
		if err != nil {
			return false, err
		}
		bit = (r.cache & 0x80) != 0
		r.cache, r.bits = r.cache&0x7f, 7
		return bit, nil
	}

	r.bits--
	bit = (r.cache & (1 << r.bits)) != 0
	r.cache &= 1<<r.bits - 1
	return bit, nil
}
