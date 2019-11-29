/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// TestPageInformationSegment tests the jbig2 page information segment.
func TestPageInformationSegment(t *testing.T) {
	t.Run("2nd", func(t *testing.T) {
		data := []byte{
			// Header
			0x00, 0x00, 0x00, 0x01, 0x30, 0x00, 0x01, 0x00, 0x00, 0x00, 0x13,

			// Data part
			0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x38, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}
		r := reader.New(data)
		d := &document{}
		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		p := &PageInformationSegment{}
		require.NoError(t, p.Init(h, r))

		assert.Equal(t, 64, p.PageBMWidth)
		assert.Equal(t, 56, p.PageBMHeight)
		assert.Equal(t, uint8(0), p.defaultPixelValue)
		assert.Equal(t, bitmap.CombinationOperator(0), p.combinationOperator)
		assert.False(t, p.IsStripe)
	})
}

func TestEncodePageInformationSegment(t *testing.T) {
	p := &PageInformationSegment{
		PageBMWidth:       64,
		PageBMHeight:      56,
		defaultPixelValue: 0,
		IsStripe:          false,
		IsLossless:        true,
	}
	h := &Header{SegmentNumber: 1, PageAssociation: 1, Type: TPageInformation, SegmentData: p}

	// initialize buffered writer
	w := writer.BufferedMSB()
	// encode page information segment
	n, err := h.Encode(w)
	require.NoError(t, err)

	assert.Equal(t, n, len(w.Data()), "n: %d", n)
	assert.Equal(t, int64(11), h.HeaderLength)
	assert.Equal(t, uint64(19), h.SegmentDataLength)

	// encoded data should be equal to:
	expected := []byte{
		// Header part
		// 00000000 00000000 00000000 00000001 - segment number
		0x00, 0x00, 0x00, 0x01,
		// 11000000 - segment flags
		0x30,
		// segment refered to count and flags
		0x00,
		// page association
		0x01,
		// segment data length - 19 bytes
		// 00000000 00000000 00000000 00010011
		0x00, 0x00, 0x00, 0x13,
		// 00000000 00000000 00000000 01000000 - bm width
		0x00, 0x00, 0x00, 0x40,
		// 00000000 00000000 00000000 00111000 - bm height
		0x00, 0x00, 0x00, 0x38,
		// 00000000 00000000 00000000 00000000 - x resolution
		0x00, 0x00, 0x00, 0x00,
		// 00000000 00000000 00000000 00000000 - y resolution
		0x00, 0x00, 0x00, 0x00,
		// 00000001 - flags
		0x01,
		// 00000000 00000000
		0x00, 0x00,
	}
	assert.Equal(t, expected, w.Data())
}
