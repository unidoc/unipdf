/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bufio"
	"errors"
	"io"

	"github.com/unidoc/unidoc/common"
)

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
