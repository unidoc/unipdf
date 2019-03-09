package reader

import (
	"errors"
	"github.com/unidoc/unidoc/common"
	"io"
)

// SubstreamReader is the wrapper over the Reader
type SubstreamReader struct {
	streamPos uint64

	wrapped *Reader

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
}

// NewSubstreamReader
func NewSubstreamReader(r *Reader, offset, length uint64) (*SubstreamReader, error) {
	if r == nil {
		return nil, errors.New("Root reader is nil")
	}

	s := &SubstreamReader{

		wrapped: r,
		offset:  offset,
		length:  length,
		buffer:  make([]byte, 4096),
	}

	common.Log.Debug("SubstreamReader offset: %d. length: %d. Reader.offset: %d", offset, length, r.r)
	return s, nil
}

func (s *SubstreamReader) ReadByte() (byte, error) {
	if s.streamPos >= s.length {
		common.Log.Debug("StreamPos: '%d' >= length: '%d'", s.streamPos, s.length)
		return 0, io.EOF
	}

	if s.streamPos >= s.bufferTop || s.streamPos < s.bufferBase {
		common.Log.Debug("Fill the buffer.")
		if err := s.fillBuffer(); err != nil {
			return 0, err
		}
	}

	common.Log.Debug("StreamPos: %d, BufferBase: %d, BufferTop: %d", s.streamPos, s.bufferBase, s.bufferTop)
	read := s.buffer[s.streamPos-s.bufferBase]
	s.streamPos += 1
	return read, nil
}

// Read reads the substreamReader reads
func (s *SubstreamReader) Read(b []byte) (int, error) {

	// if the stream position is
	if s.streamPos >= s.length {
		common.Log.Debug("StreamPos: '%d' >= length: '%d'", s.streamPos, s.length)
		return 0, io.EOF
	}

	if uint64(s.wrapped.r) != s.streamPos+s.offset {
		_, err := s.wrapped.Seek(int64(s.streamPos+s.offset), io.SeekCurrent)
		if err != nil {
			common.Log.Debug("Seak StreamPos: '%d' + s.offset: '%d'", s.streamPos, s.offset)
			return 0, io.EOF
		}
	}

	toRead := min(uint64(len(b)), s.length-s.streamPos)

	bytes := make([]byte, toRead)
	read, err := s.wrapped.Read(bytes)
	if err != nil {
		return 0, err
	}

	s.streamPos += uint64(read)
	for i := 0; i < len(bytes); i++ {
		b[i] = bytes[i]
	}

	return read, nil
}

func (s *SubstreamReader) Length() uint64 {
	return s.length
}

func (s *SubstreamReader) ReadBit() (b int, err error) {
	return s.wrapped.ReadBit()
}

func (s *SubstreamReader) Align() (skipped byte) {
	return s.wrapped.Align()
}

func (s *SubstreamReader) Seek(offset int64, whence int) (int64, error) {

	switch whence {
	case io.SeekStart:
		offset += int64(s.offset)
	case io.SeekCurrent:
	case io.SeekEnd:
		offset = int64(s.offset+s.length) + offset
	}

	s.streamPos = uint64(offset) - s.offset

	return s.wrapped.Seek(offset, whence)
}

// Offset returns current SubstreamReader offset
func (s *SubstreamReader) Offset() uint64 {
	return s.offset
}

// StreamPosition gets the stream position of the substream reader
func (s *SubstreamReader) StreamPosition() uint64 {
	return s.streamPos
}

func (s *SubstreamReader) fillBuffer() error {
	if uint64(s.wrapped.r) != s.streamPos+s.offset {
		_, err := s.wrapped.Seek(int64(s.streamPos+s.offset), io.SeekCurrent)
		if err != nil {
			return err
		}
	}

	s.bufferBase = uint64(s.streamPos)

	toRead := min(uint64(len(s.buffer)), s.length-s.streamPos)

	bytes := make([]byte, toRead)
	common.Log.Debug("toRead :%d", s.length)

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

func min(f, s uint64) uint64 {
	if f < s {
		return f
	}
	return s
}
