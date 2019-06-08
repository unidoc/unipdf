/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// TestDecodeHeader test the segment header model decode process.
func TestDecodeHeader(t *testing.T) {
	data := []byte{
		// header
		0x00, 0x00, 0x00, 0x0B, 0x27, 0x00, 0x02, 0x00, 0x00, 0x00, 0x23,

		// data
		0x00, 0x00, 0x00, 0x36, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x04,
		0x00, 0x00, 0x00, 0x0B, 0x00, 0x08, 0x03, 0xFF, 0xFD, 0xFF, 0x02, 0xFE,
		0xFE, 0xFE, 0x04, 0xEE, 0xED, 0x87, 0xFB, 0xCB, 0x2B, 0xFF, 0xAC,
	}

	r := reader.New(data)
	d := &document{}
	h, err := NewHeader(d, r, 0, OSequential)
	require.NoError(t, err)

	assert.Equal(t, int64(11), h.HeaderLength)
	assert.Equal(t, uint64(11), h.SegmentDataStartOffset)

	s, err := h.subInputReader()
	require.NoError(t, err)

	b, err := s.ReadByte()
	require.NoError(t, err)
	assert.Equal(t, byte(0x00), b)

	three := make([]byte, 3)
	read, err := s.Read(three)
	require.NoError(t, err)

	assert.Equal(t, 3, read)
	assert.Equal(t, byte(0x36), three[2])
}
