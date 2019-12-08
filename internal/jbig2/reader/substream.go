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

// SubstreamReader is the wrapper over the Reader's parts that is allowed only to
// operate on the selected data space.
type SubstreamReader struct {
	// stream position
	streamPos uint64
	// wrapped stream reader
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

// NewSubstreamReader creates new SubStreamReader for the provided wrapped StreamReader 'r' with defined 'offset' and 'length'.
func NewSubstreamReader(r StreamReader, offset, length uint64) (*SubstreamReader, error) {
	if r == nil {
		return nil, errors.New("root reader is nil")
	}

	common.Log.Trace("New substream at offset: %d with length: %d", offset, length)
	return &SubstreamReader{
		wrapped: r,
		offset:  offset,
		length:  length,
		buffer:  make([]byte, length),
	}, nil
}

// Align implements StreamReader interface.
func (s *SubstreamReader) Align() (skipped byte) {
	skipped = s.bits
	s.bits = 0
	return skipped
}

// BitPosition implements StreamReader interface.
func (s *SubstreamReader) BitPosition() int {
	return int(s.bits)
}

// Length implements StreamReader interface.
func (s *SubstreamReader) Length() uint64 {
	return s.length
}

// Mark implements StreamReader interface.
func (s *SubstreamReader) Mark() {
	s.mark = s.streamPos
	s.markBits = s.bits
}

// Offset returns current SubstreamReader offset.
func (s *SubstreamReader) Offset() uint64 {
	return s.offset
}

// Read implements io.Reader interface.
func (s *SubstreamReader) Read(b []byte) (n int, err error) {
	if s.streamPos >= s.length {
		common.Log.Trace("StreamPos: '%d' >= length: '%d'", s.streamPos, s.length)
		return 0, io.EOF
	}

	for ; n < len(b); n++ {
		if b[n], err = s.readUnalignedByte(); err != nil {
			if err == io.EOF {
				return n, nil
			}
			return 0, err
		}
	}
	return n, nil
}

// ReadBit implements StreamReader interface.
func (s *SubstreamReader) ReadBit() (bit int, err error) {
	booleanBit, err := s.readBool()
	if err != nil {
		return 0, err
	}

	if booleanBit {
		bit = 1
	}
	return bit, nil
}

// ReadBits implements StreamReader interface.
func (s *SubstreamReader) ReadBits(n byte) (u uint64, err error) {
	if n < s.bits {
		shift := s.bits - n
		u = uint64(s.cache >> shift)
		s.cache &= 1<<shift - 1
		s.bits = shift
		return u, nil
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
				return 0, err
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

// ReadBool implements StreamReader interface.
func (s *SubstreamReader) ReadBool() (bool, error) {
	return s.readBool()
}

// ReadByte implements io.ByteReader.
func (s *SubstreamReader) ReadByte() (byte, error) {
	if s.bits == 0 {
		return s.readBufferByte()
	}
	return s.readUnalignedByte()
}

// ReadUint32 implements Streamreader interface.
func (s *SubstreamReader) ReadUint32() (uint32, error) {
	ub := make([]byte, 4)

	_, err := s.Read(ub)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(ub), nil
}

// Reset implements StreamReader interface.
func (s *SubstreamReader) Reset() {
	s.streamPos = s.mark
	s.bits = s.markBits
}

// Seek implements the io.Seeker interface.
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
	s.bits = 0
	return int64(s.streamPos), nil
}

// StreamPosition implements StreamReader interface.
func (s *SubstreamReader) StreamPosition() int64 {
	return int64(s.streamPos)
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

	read, err := s.wrapped.Read(bytes)
	if err != nil {
		return err
	}

	for i := uint64(0); i < toRead; i++ {
		s.buffer[i] = bytes[i]
	}
	s.bufferTop = s.bufferBase + uint64(read)

	return nil
}

func (s *SubstreamReader) readBool() (bit bool, err error) {
	if s.bits == 0 {
		s.cache, err = s.readBufferByte()
		if err != nil {
			return false, err
		}

		bit = (s.cache & 0x80) != 0
		s.cache, s.bits = s.cache&0x7f, 7
		return bit, nil
	}

	s.bits--
	bit = (s.cache & (1 << s.bits)) != 0
	s.cache &= 1<<s.bits - 1
	return bit, nil
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
	return b, nil
}

func (s *SubstreamReader) readBufferByte() (byte, error) {
	if s.streamPos >= s.length {
		return 0, io.EOF
	}

	if s.streamPos >= s.bufferTop || s.streamPos < s.bufferBase {
		if err := s.fillBuffer(); err != nil {
			return 0, err
		}
	}

	read := s.buffer[s.streamPos-s.bufferBase]
	s.streamPos++

	return read, nil
}

func min(f, s uint64) uint64 {
	if f < s {
		return f
	}
	return s
}
