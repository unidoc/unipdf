/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// TestSymbolDictionaryDecode tests the symbol dictionary decode process.
func TestSymbolDictionaryDecode(t *testing.T) {
	setLogger()

	t.Run("1st", func(t *testing.T) {
		data := []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x18,
			// Data part
			0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0xE9, 0xCB,
			0xF4, 0x00, 0x26, 0xAF, 0x04, 0xBF, 0xF0, 0x78, 0x2F, 0xE0, 0x00, 0x40,
		}
		r := reader.New(data)
		d := &document{}
		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		assert.Equal(t, TSymbolDictionary, h.Type)
		assert.Equal(t, false, h.PageAssociationFieldSize)
		assert.Equal(t, false, h.RetainFlag)
		assert.Equal(t, 0, len(h.RTSegments))
		assert.Equal(t, uint64(24), h.SegmentDataLength)

		sg, err := h.GetSegmentData()
		require.NoError(t, err)

		s, ok := sg.(*SymbolDictionary)
		require.True(t, ok)

		assert.True(t, s.isHuffmanEncoded)
		assert.False(t, s.useRefinementAggregation)
		assert.Equal(t, 1, s.amountOfExportedSymbols)
		assert.Equal(t, 1, s.amountOfNewSymbols)

		bm, err := s.GetDictionary()
		require.NoError(t, err)

		if assert.NotEmpty(t, bm) {
			for _, b := range bm {
				t.Logf("Bitmap: %s", b.String())
			}
		}
	})

	t.Run("3rd", func(t *testing.T) {
		data := []byte{
			// Header
			0x00, 0x00, 0x00, 0x02, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x1C,
			// Data part
			0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0xE5, 0xCD,
			0xF8, 0x00, 0x79, 0xE0, 0x84, 0x10, 0x81, 0xF0, 0x82, 0x10, 0x86, 0x10,
			0x79, 0xF0, 0x00, 0x80,
		}
		r := reader.New(data)
		d := &document{}
		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		assert.Equal(t, TSymbolDictionary, h.Type)
		assert.Equal(t, false, h.PageAssociationFieldSize)
		assert.Equal(t, false, h.RetainFlag)
		assert.Equal(t, 0, len(h.RTSegments))
		assert.Equal(t, uint64(28), h.SegmentDataLength)

		sg, err := h.GetSegmentData()
		require.NoError(t, err)

		s, ok := sg.(*SymbolDictionary)
		require.True(t, ok)

		assert.True(t, s.isHuffmanEncoded)
		assert.False(t, s.useRefinementAggregation)
		assert.Equal(t, 2, s.amountOfExportedSymbols)
		assert.Equal(t, 2, s.amountOfNewSymbols)

		bm, err := s.GetDictionary()
		require.NoError(t, err)

		if assert.NotEmpty(t, bm) {
			for _, b := range bm {
				t.Logf("Bitmap: %s", b.String())
			}
		}
	})

	t.Run("10th", func(t *testing.T) {
		data := []byte{
			// Header
			0x00, 0x00, 0x00, 0x09, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x1B,

			// Segment data
			0x08, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02,
			0x4F, 0xE7, 0x8C, 0x20, 0x0E, 0x1D, 0xC7, 0xCF, 0x01, 0x11, 0xC4, 0xB2,
			0x6F, 0xFF, 0xAC,
		}
		r := reader.New(data)
		d := &document{}
		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		assert.Equal(t, TSymbolDictionary, h.Type)
		assert.Equal(t, false, h.PageAssociationFieldSize)
		assert.Equal(t, false, h.RetainFlag)
		assert.Equal(t, 0, len(h.RTSegments))
		assert.Equal(t, uint64(27), h.SegmentDataLength)

		sg, err := h.GetSegmentData()
		require.NoError(t, err)

		s, ok := sg.(*SymbolDictionary)
		require.True(t, ok)

		assert.False(t, s.isHuffmanEncoded)
		assert.Equal(t, int8(2), s.sdTemplate)
		assert.Equal(t, false, s.isCodingContextUsed)
		assert.Equal(t, false, s.isCodingContextRetained)

		assert.Equal(t, int8(2), s.sdATX[0])
		assert.Equal(t, int8(-1), s.sdATY[0])

		assert.Equal(t, 2, s.amountOfExportedSymbols)
		assert.Equal(t, 2, s.amountOfNewSymbols)

		bm, err := s.GetDictionary()
		require.NoError(t, err)

		if assert.NotEmpty(t, bm) {
			for _, b := range bm {
				t.Logf("Bitmap: %s", b.String())
			}
		}
	})

	t.Run("17th", func(t *testing.T) {
		data := []byte{
			// 17th segment
			// Header
			0x00, 0x00, 0x00, 0x10, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x16,

			// Data Part
			0x08, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
			0x01, 0x4F, 0xE7, 0x8D, 0x68, 0x1B, 0x14, 0x2F, 0x3F, 0xFF, 0xAC,

			// 18th segment
			// Header
			0x00, 0x00, 0x00, 0x11, 0x00, 0x21, 0x10, 0x03, 0x00, 0x00, 0x00, 0x20,

			// Data Part
			0x08, 0x02, 0x02, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x03,
			0x00, 0x00, 0x00, 0x02, 0x4F, 0xE9, 0xD7, 0xD5, 0x90, 0xC3, 0xB5, 0x26,
			0xA7, 0xFB, 0x6D, 0x14, 0x98, 0x3F, 0xFF, 0xAC,
		}

		r := reader.New(data)
		p3 := &page{segments: []*Header{}}
		d := &document{pages: []Pager{nil, nil, p3}}
		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		p3.setSegment(h)

		assert.Equal(t, TSymbolDictionary, h.Type)
		assert.Equal(t, false, h.PageAssociationFieldSize)

		assert.Equal(t, false, h.RetainFlag)
		assert.Equal(t, 0, len(h.RTSegments))
		assert.Equal(t, uint64(22), h.SegmentDataLength)

		sg, err := h.GetSegmentData()
		require.NoError(t, err)

		s, ok := sg.(*SymbolDictionary)
		require.True(t, ok)

		assert.False(t, s.isHuffmanEncoded)
		assert.False(t, s.useRefinementAggregation)
		assert.Equal(t, int8(2), s.sdTemplate)
		assert.False(t, s.isCodingContextUsed)
		assert.False(t, s.isCodingContextRetained)

		if assert.Len(t, s.sdATX, 1) {
			assert.Equal(t, s.sdATX[0], int8(2))
		}

		if assert.Len(t, s.sdATY, 1) {
			assert.Equal(t, s.sdATY[0], int8(-1))
		}

		assert.Equal(t, 1, s.amountOfExportedSymbols)
		assert.Equal(t, 1, s.amountOfNewSymbols)

		bm, err := s.GetDictionary()
		require.NoError(t, err)

		if assert.NotEmpty(t, bm) {
			for _, b := range bm {
				t.Logf("Bitmap: %s", b.String())
			}
		}
	})
}

var alreadySet bool

func setLogger() {
	if testing.Verbose() && !alreadySet {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}
}
