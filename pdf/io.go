/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"bufio"
	"errors"
	"os"

	"github.com/unidoc/unidoc/common"
)

func (this *PdfParser) ReadAtLeast(p []byte, n int) (int, error) {
	remaining := n
	start := 0
	numRounds := 0
	for remaining > 0 {
		nRead, err := this.reader.Read(p[start:])
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
func (this *PdfParser) GetFileOffset() int64 {
	offset, _ := this.rs.Seek(0, os.SEEK_CUR)
	offset -= int64(this.reader.Buffered())
	return offset
}

// Seek the file to an offset position.
func (this *PdfParser) SetFileOffset(offset int64) {
	this.rs.Seek(offset, os.SEEK_SET)
	this.reader = bufio.NewReader(this.rs)
}
