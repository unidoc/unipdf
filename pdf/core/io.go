/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bufio"
	"errors"
	"os"

	"github.com/unidoc/unidoc/common"
)

// ReadAtLeast reads at least n bytes into slice p.
// Returns the number of bytes read (should always be == n), and an error on failure.
// TODO (v3): Unexport.
func (parser *PdfParser) ReadAtLeast(p []byte, n int) (int, error) {
	remaining := n
	start := 0
	numRounds := 0
	for remaining > 0 {
		nRead, err := parser.reader.Read(p[start:])
		if err != nil {
			common.Log.Debug("ERROR Failed reading (%d;%d) %s", nRead, numRounds, err.Error())
			return start, errors.New("Failed reading")
		}
		numRounds++
		start += nRead
		remaining -= nRead
	}
	return start, nil
}

// Get the current file offset, accounting for buffered position.
// TODO (v3): Unexport.
func (parser *PdfParser) GetFileOffset() int64 {
	offset, _ := parser.rs.Seek(0, os.SEEK_CUR)
	offset -= int64(parser.reader.Buffered())
	return offset
}

// Seek the file to an offset position.
// TODO (v3): Unexport.
func (parser *PdfParser) SetFileOffset(offset int64) {
	parser.rs.Seek(offset, os.SEEK_SET)
	parser.reader = bufio.NewReader(parser.rs)
}
