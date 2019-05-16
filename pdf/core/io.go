/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bufio"
	"errors"
	"io"

	"github.com/unidoc/unipdf/v3/common"
)

// Offset reader encapsulates io.ReadSeeker and offsets it by the specified
// offset, thus skipping the first offset bytes. Reading always occurs after the
// offset.
type offsetReader struct {
	reader io.ReadSeeker
	offset int64
}

func newOffsetReader(reader io.ReadSeeker, offset int64) (*offsetReader, error) {
	r := &offsetReader{
		reader: reader,
		offset: offset,
	}

	_, err := r.Seek(0, io.SeekStart)
	return r, err
}

func (r *offsetReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *offsetReader) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart {
		offset += r.offset
	}

	n, err := r.reader.Seek(offset, whence)
	if err != nil {
		return n, err
	}

	if whence == io.SeekCurrent {
		n -= r.offset
	}
	if n < 0 {
		return 0, errors.New("core.offsetReader.Seek: negative position")
	}

	return n, nil
}

// ReadAtLeast reads at least n bytes into slice p.
// Returns the number of bytes read (should always be == n), and an error on failure.
func (parser *PdfParser) ReadAtLeast(p []byte, n int) (int, error) {
	remaining := n
	start := 0
	numRounds := 0
	for remaining > 0 {
		nRead, err := parser.reader.Read(p[start:])
		if err != nil {
			common.Log.Debug("ERROR Failed reading (%d;%d) %s", nRead, numRounds, err.Error())
			return start, errors.New("failed reading")
		}
		numRounds++
		start += nRead
		remaining -= nRead
	}
	return start, nil
}

// GetFileOffset returns the current file offset, accounting for buffered position.
func (parser *PdfParser) GetFileOffset() int64 {
	offset, _ := parser.rs.Seek(0, io.SeekCurrent)
	offset -= int64(parser.reader.Buffered())
	return offset
}

// SetFileOffset sets the file to an offset position and resets buffer.
func (parser *PdfParser) SetFileOffset(offset int64) {
	if offset < 0 {
		offset = 0
	}

	parser.rs.Seek(offset, io.SeekStart)
	parser.reader = bufio.NewReader(parser.rs)
}

// ReadBytesAt reads byte content at specific offset and length within the PDF.
func (parser *PdfParser) ReadBytesAt(offset, len int64) ([]byte, error) {
	curPos := parser.GetFileOffset()

	_, err := parser.rs.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	bb := make([]byte, len)
	_, err = io.ReadAtLeast(parser.rs, bb, int(len))
	if err != nil {
		return nil, err
	}

	// Restore.
	parser.SetFileOffset(curPos)

	return bb, nil
}
