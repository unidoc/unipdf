/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package reader

import (
	"errors"
	"github.com/unidoc/unidoc/common"
	"io"
)

// SubstreamReader is the wrapper over the Reader
type SubstreamReader struct {
	streamPos uint64

	wrapped StreamReader

	// The position in the wrapped stream at which the window starts. Offset is an absolute value.
	offset uint64

	// The length of the window. Length is an relative value.
	length uint64

	// A buffer which is used to improve read performance.
	buffer []byte

	// Location of the first byte in the buffer with respect to the start of the stream.
	bufferBase uint64

	// Location of the last byte in the buffer with respect to the start of the stream.
	bufferTop uint64

	// unread bits are stored here
	cache byte

	// number of unread bits in cache
	bits byte

	mark     uint64
	markBits byte
}

// NewSubstreamReader
func NewSubstreamReader(r StreamReader, offset, length uint64) (*SubstreamReader, error) {
	if r == nil {
		return nil, errors.New("Root reader is nil")
	}
	common.Log.Debug("New substream at offset: %d with length: %d", offset, length)
	s := &SubstreamReader{

		wrapped: r,
		offset:  offset,
		length:  length,
		buffer:  make([]byte, length),
	}

	// if r.StreamPosition() != int64(offset) {

	// 	// seek if the offset is not at that initial position
	// 	n, err := r.Seek(int64(offset), io.SeekStart)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	common.Log.Debug("Seeked the root reader to offset: %d", n)
	// }

	// common.Log.Debug("SubstreamReader offset: %d. length: %d. Reader.offset: %d", offset, length, r.r)
	return s, nil
}

func (s *SubstreamReader) BitPosition() int {
	return int(s.bits)
}

func (s *SubstreamReader) ReadByte() (byte, error) {
	if s.bits == 0 {
		return s.readBufferByte()
	}
	return s.readUnalignedByte()
}

// Read reads the substreamReader reads
func (s *SubstreamReader) Read(b []byte) (n int, err error) {

	// if the stream position is
	if s.streamPos >= s.length {
		common.Log.Debug("StreamPos: '%d' >= length: '%d'", s.streamPos, s.length)
		return 0, io.EOF
	}

	for ; n < len(b); n++ {
		if b[n], err = s.readUnalignedByte(); err != nil {
			if err == io.EOF {
				return n, nil
			}
			return
		}
	}

	return
}

func (s *SubstreamReader) Length() uint64 {
	return s.length
}

// ReadBit reads the next binary value from the current cache
func (s *SubstreamReader) ReadBit() (b int, err error) {
	var bit bool
	bit, err = s.readBool()
	if err != nil {
		return
	}

	if bit {
		b = 1
	}

	// common.Log.Debug("Bit position after readingBit: %d", s.bits)
	return
}

func (s *SubstreamReader) ReadBool() (bool, error) {
	return s.readBool()
}

func (s *SubstreamReader) ReadBits(n byte) (u uint64, err error) {
	// common.Log.Debug("Bit position before reading bits n: %d,  %d", n, s.bits)
	// defer func() {
	// common.Log.Debug("Bit position after readingBits n:%d  b: %d", size, s.bits)
	// }()
	if n < s.bits {
		shift := s.bits - n
		u = uint64(s.cache >> shift)
		s.cache &= 1<<shift - 1
		s.bits = shift
		return
	}

	if n > s.bits {
		if s.bits > 0 {
			u = uint64(s.cache)
			n -= s.bits
		}
		var b byte
		for n >= 8 {
			b, err = s.readBufferByte()
			if err != nil {
				return
			}
			u = u<<8 + uint64(b)
			n -= 8
		}

		if n > 0 {
			if s.cache, err = s.readBufferByte(); err != nil {
				return 0, err
			}
			shift := 8 - n
			u = u<<n + uint64(s.cache>>shift)
			s.cache &= 1<<shift - 1
			s.bits = shift
		} else {
			s.bits = 0
		}
		return u, nil
	}

	s.bits = 0
	return uint64(s.cache), nil

}

func (s *SubstreamReader) Align() (skipped byte) {
	skipped = s.bits
	s.bits = 0
	return
}

// Mark marks a position in the stream to be returned to by a subsequent call to 'Reset'.
func (s *SubstreamReader) Mark() {
	s.mark = s.streamPos
	s.markBits = s.bits
}

// Offset returns current SubstreamReader offset
func (s *SubstreamReader) Offset() uint64 {
	return s.offset
}

// Reset returns the stream pointer to its previous position, including the bit offset,
// at the time of the most recent unmatched call to mark.
func (s *SubstreamReader) Reset() {
	s.streamPos = s.mark
	s.bits = s.markBits
}

// StreamPosition gets the stream position of the substream reader
func (s *SubstreamReader) StreamPosition() int64 {
	return int64(s.streamPos)
}

// Seek implements the io.Seeker interface
func (s *SubstreamReader) Seek(offset int64, whence int) (int64, error) {

	switch whence {
	case io.SeekStart:
		s.streamPos = uint64(offset)
	case io.SeekCurrent:
		s.streamPos += uint64(offset)
	case io.SeekEnd:
		s.streamPos = s.length + uint64(offset)
	default:
		return 0, errors.New("reader.SubstreamReader.Seek invalid whence")
	}

	if s.streamPos < 0 {
		return 0, errors.New("reader.Substream.Seek negative position")
	}
	s.bits = 0

	return int64(s.streamPos), nil
}

func (s *SubstreamReader) fillBuffer() error {
	if uint64(s.wrapped.StreamPosition()) != s.streamPos+s.offset {
		_, err := s.wrapped.Seek(int64(s.streamPos+s.offset), io.SeekStart)
		if err != nil {
			return err
		}
	}

	s.bufferBase = uint64(s.streamPos)

	toRead := min(uint64(len(s.buffer)), s.length-s.streamPos)

	bytes := make([]byte, toRead)
	// common.Log.Debug("toRead :%d", s.length)

	read, err := s.wrapped.Read(bytes)
	if err != nil {
		return err
	}
	var i uint64
	for i = 0; i < toRead; i++ {
		s.buffer[i] = bytes[i]
	}
	s.bufferTop = s.bufferBase + uint64(read)

	return nil
}

func (s *SubstreamReader) readBool() (b bool, err error) {
	if s.bits == 0 {
		s.cache, err = s.readBufferByte()
		if err != nil {
			return
		}
		b = (s.cache & 0x80) != 0
		s.cache, s.bits = s.cache&0x7f, 7
		return
	}

	s.bits--
	b = (s.cache & (1 << s.bits)) != 0
	s.cache &= 1<<s.bits - 1
	return
}

func (s *SubstreamReader) readUnalignedByte() (b byte, err error) {
	bits := s.bits
	b = s.cache << (8 - bits)
	s.cache, err = s.readBufferByte()
	if err != nil {
		return 0, err
	}
	b |= s.cache >> bits
	s.cache &= 1<<bits - 1
	return
}

func (s *SubstreamReader) readBufferByte() (b byte, err error) {
	if s.streamPos >= s.length {
		// common.Log.Debug("StreamPos: '%d' >= length: '%d'", s.streamPos, s.length)
		return 0, io.EOF
	}

	if s.streamPos >= s.bufferTop || s.streamPos < s.bufferBase {
		// common.Log.Debug("Fill the buffer.")
		if err := s.fillBuffer(); err != nil {
			return 0, err
		}
	}

	read := s.buffer[s.streamPos-s.bufferBase]
	s.streamPos += 1

	return read, nil

}

func min(f, s uint64) uint64 {
	if f < s {
		return f
	}
	return s
}
