/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package jbig2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/segments"
)

// TestDocument tests the jbig2.Document decoding.
func TestDocument(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	t.Run("AnnexH", func(t *testing.T) {
		data := []byte{
			0x97, 0x4A, 0x42, 0x32, 0x0D, 0x0A, 0x1A, 0x0A, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x18, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			0x00, 0x01, 0xE9, 0xCB, 0xF4, 0x00, 0x26, 0xAF, 0x04, 0xBF, 0xF0, 0x78, 0x2F, 0xE0, 0x00, 0x40,
			0x00, 0x00, 0x00, 0x01, 0x30, 0x00, 0x01, 0x00, 0x00, 0x00, 0x13, 0x00, 0x00, 0x00, 0x40, 0x00,
			0x00, 0x00, 0x38, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x02, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x1C, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00,
			0x00, 0x00, 0x02, 0xE5, 0xCD, 0xF8, 0x00, 0x79, 0xE0, 0x84, 0x10, 0x81, 0xF0, 0x82, 0x10, 0x86,
			0x10, 0x79, 0xF0, 0x00, 0x80, 0x00, 0x00, 0x00, 0x03, 0x07, 0x42, 0x00, 0x02, 0x01, 0x00, 0x00,
			0x00, 0x31, 0x00, 0x00, 0x00, 0x25, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00,
			0x00, 0x01, 0x00, 0x0C, 0x09, 0x00, 0x10, 0x00, 0x00, 0x00, 0x05, 0x01, 0x10, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x40, 0x07, 0x08,
			0x70, 0x41, 0xD0, 0x00, 0x00, 0x00, 0x04, 0x27, 0x00, 0x01, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00,
			0x00, 0x36, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x0B, 0x00, 0x01,
			0x26, 0xA0, 0x71, 0xCE, 0xA7, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xF8, 0xF0, 0x00, 0x00, 0x00, 0x05, 0x10, 0x01,
			0x01, 0x00, 0x00, 0x00, 0x2D, 0x01, 0x04, 0x04, 0x00, 0x00, 0x00, 0x0F, 0x20, 0xD1, 0x84, 0x61,
			0x18, 0x45, 0xF2, 0xF9, 0x7C, 0x8F, 0x11, 0xC3, 0x9E, 0x45, 0xF2, 0xF9, 0x7D, 0x42, 0x85, 0x0A,
			0xAA, 0x84, 0x62, 0x2F, 0xEE, 0xEC, 0x44, 0x62, 0x22, 0x35, 0x2A, 0x0A, 0x83, 0xB9, 0xDC, 0xEE,
			0x77, 0x80, 0x00, 0x00, 0x00, 0x06, 0x17, 0x20, 0x05, 0x01, 0x00, 0x00, 0x00, 0x57, 0x00, 0x00,
			0x00, 0x20, 0x00, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x0F, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x04, 0x00, 0x00, 0x00, 0xAA, 0xAA, 0xAA, 0xAA, 0x80, 0x08, 0x00, 0x80, 0x36, 0xD5, 0x55, 0x6B,
			0x5A, 0xD4, 0x00, 0x40, 0x04, 0x2E, 0xE9, 0x52, 0xD2, 0xD2, 0xD2, 0x8A, 0xA5, 0x4A, 0x00, 0x20,
			0x02, 0x23, 0xE0, 0x95, 0x24, 0xB4, 0x92, 0x8A, 0x4A, 0x92, 0x54, 0x92, 0xD2, 0x4A, 0x29, 0x2A,
			0x49, 0x40, 0x04, 0x00, 0x40, 0x00, 0x00, 0x00, 0x07, 0x31, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x08, 0x30, 0x00, 0x02, 0x00, 0x00, 0x00, 0x13, 0x00, 0x00, 0x00, 0x40, 0x00,
			0x00, 0x00, 0x38, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x09, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x1B, 0x08, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00,
			0x02, 0x00, 0x00, 0x00, 0x02, 0x4F, 0xE7, 0x8C, 0x20, 0x0E, 0x1D, 0xC7, 0xCF, 0x01, 0x11, 0xC4,
			0xB2, 0x6F, 0xFF, 0xAC, 0x00, 0x00, 0x00, 0x0A, 0x07, 0x40, 0x00, 0x09, 0x02, 0x00, 0x00, 0x00,
			0x1F, 0x00, 0x00, 0x00, 0x25, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x0C, 0x08, 0x00, 0x00, 0x00, 0x05, 0x8D, 0x6E, 0x5A, 0x12, 0x40, 0x85, 0xFF, 0xAC,
			0x00, 0x00, 0x00, 0x0B, 0x27, 0x00, 0x02, 0x00, 0x00, 0x00, 0x23, 0x00, 0x00, 0x00, 0x36, 0x00,
			0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x0B, 0x00, 0x08, 0x03, 0xFF, 0xFD,
			0xFF, 0x02, 0xFE, 0xFE, 0xFE, 0x04, 0xEE, 0xED, 0x87, 0xFB, 0xCB, 0x2B, 0xFF, 0xAC, 0x00, 0x00,
			0x00, 0x0C, 0x10, 0x01, 0x02, 0x00, 0x00, 0x00, 0x1C, 0x06, 0x04, 0x04, 0x00, 0x00, 0x00, 0x0F,
			0x90, 0x71, 0x6B, 0x6D, 0x99, 0xA7, 0xAA, 0x49, 0x7D, 0xF2, 0xE5, 0x48, 0x1F, 0xDC, 0x68, 0xBC,
			0x6E, 0x40, 0xBB, 0xFF, 0xAC, 0x00, 0x00, 0x00, 0x0D, 0x17, 0x20, 0x0C, 0x02, 0x00, 0x00, 0x00,
			0x3E, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00,
			0x0F, 0x00, 0x02, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x87, 0xCB, 0x82, 0x1E, 0x66, 0xA4, 0x14, 0xEB, 0x3C,
			0x4A, 0x15, 0xFA, 0xCC, 0xD6, 0xF3, 0xB1, 0x6F, 0x4C, 0xED, 0xBF, 0xA7, 0xBF, 0xFF, 0xAC, 0x00,
			0x00, 0x00, 0x0E, 0x31, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F, 0x30, 0x00,
			0x03, 0x00, 0x00, 0x00, 0x13, 0x00, 0x00, 0x00, 0x25, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x01, 0x00, 0x00,
			0x00, 0x00, 0x16, 0x08, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x4F,
			0xE7, 0x8D, 0x68, 0x1B, 0x14, 0x2F, 0x3F, 0xFF, 0xAC, 0x00, 0x00, 0x00, 0x11, 0x00, 0x21, 0x10,
			0x03, 0x00, 0x00, 0x00, 0x20, 0x08, 0x02, 0x02, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00,
			0x03, 0x00, 0x00, 0x00, 0x02, 0x4F, 0xE9, 0xD7, 0xD5, 0x90, 0xC3, 0xB5, 0x26, 0xA7, 0xFB, 0x6D,
			0x14, 0x98, 0x3F, 0xFF, 0xAC, 0x00, 0x00, 0x00, 0x12, 0x07, 0x20, 0x11, 0x03, 0x00, 0x00, 0x00,
			0x25, 0x00, 0x00, 0x00, 0x25, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x8C, 0x12, 0x00, 0x00, 0x00, 0x04, 0xA9, 0x5C, 0x8B, 0xF4, 0xC3, 0x7D, 0x96, 0x6A,
			0x28, 0xE5, 0x76, 0x8F, 0xFF, 0xAC, 0x00, 0x00, 0x00, 0x13, 0x31, 0x00, 0x03, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x14, 0x33, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}

		// create the document
		d, err := NewDocument(data)
		require.NoError(t, err)

		assert.Equal(t, uint32(3), d.NumberOfPages)
		assert.Equal(t, segments.OSequential, d.OrganizationType)
		assert.Equal(t, false, d.GBUseExtTemplate)

		p1 := d.Pages[1]
		require.NoError(t, err)

		t.Run("Segment#0", func(t *testing.T) {
			s1 := p1.GetSegment(0)
			require.NotNil(t, s1)

			seg, err := s1.GetSegmentData()
			require.NoError(t, err)

			sd, ok := seg.(*segments.SymbolDictionary)
			require.True(t, ok)

			assert.True(t, sd.IsHuffmanEncoded())
			assert.False(t, sd.UseRefinementAggregation())
			assert.Equal(t, uint32(1), sd.NumberOfExportedSymbols())
			assert.Equal(t, uint32(1), sd.NumberOfNewSymbols())

			bm, err := sd.GetDictionary()
			require.NoError(t, err)
			require.Len(t, bm, 1)

			pLetter := bm[0]
			symbol := pSymbol(t)
			assert.True(t, pLetter.Equals(symbol), fmt.Sprintf("P decoded: %s - Should be: %s", pLetter, symbol))
		})

		t.Run("Segment#1", func(t *testing.T) {
			h := p1.GetSegment(1)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			pi, ok := seg.(*segments.PageInformationSegment)
			require.True(t, ok)

			assert.Equal(t, 64, pi.PageBMWidth)
			assert.Equal(t, 56, pi.PageBMHeight)

			assert.Equal(t, 0, pi.ResolutionX)
			assert.Equal(t, 0, pi.ResolutionY)

			assert.Equal(t, bitmap.CmbOpOr, pi.CombinationOperator())
			assert.False(t, pi.IsStripe)
		})

		t.Run("Segment#3", func(t *testing.T) {
			h := p1.GetSegment(2)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			sd, ok := seg.(*segments.SymbolDictionary)
			require.True(t, ok)

			dict, err := sd.GetDictionary()
			require.NoError(t, err)

			require.Len(t, dict, 2)

			c := cSymbol(t)
			a := aSymbol(t)
			assert.True(t, dict[0].Equals(c))
			assert.True(t, dict[1].Equals(a))
		})

		t.Run("Segment#4", func(t *testing.T) {
			h := p1.GetSegment(3)
			require.NotNil(t, 3)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			tr, ok := seg.(*segments.TextRegion)
			require.True(t, ok)

			bm, err := tr.GetRegionBitmap()
			require.NoError(t, err)

			assert.Equal(t, 8, bm.Height)
			assert.Equal(t, 37, bm.Width)

			expected := bitmap.New(37, 8)
			err = bitmap.Blit(cSymbol(t), expected, 0, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(aSymbol(t), expected, 8, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(pSymbol(t), expected, 16, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(aSymbol(t), expected, 23, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(cSymbol(t), expected, 31, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			assert.True(t, expected.Equals(bm))
		})

		t.Run("Segment#5", func(t *testing.T) {
			h := p1.GetSegment(4)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			g, ok := seg.(*segments.GenericRegion)
			require.True(t, ok)

			bm, err := g.GetRegionBitmap()
			require.NoError(t, err)
			expected := getFrame(t)
			assert.True(t, expected.Equals(bm))
		})

		t.Run("Segment#6", func(t *testing.T) {
			h := p1.GetSegment(5)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			p, ok := seg.(*segments.PatternDictionary)
			require.True(t, ok)

			dict, err := p.GetDictionary()
			require.NoError(t, err)
			checkPatternDictionary(t, dict)
		})

		t.Run("Segment#7", func(t *testing.T) {
			header := p1.GetSegment(6)
			require.NotNil(t, header)

			seg, err := header.GetSegmentData()
			require.NoError(t, err)

			h, ok := seg.(*segments.HalftoneRegion)
			require.True(t, ok)

			patterns, err := h.GetPatterns()
			require.NoError(t, err)

			expected := getPatternsFirst(t)
			if assert.Equal(t, len(expected), len(patterns)) {
				for i, p := range patterns {
					assert.True(t, expected[i].Equals(p))
				}
			}

			_, err = h.GetRegionBitmap()
			require.NoError(t, err)
		})

		t.Run("Segment#8", func(t *testing.T) {
			h := p1.GetSegment(7)
			require.NotNil(t, h)

			assert.Equal(t, segments.TEndOfPage, h.Type)
		})

		t.Run("Page#1", func(t *testing.T) {
			_, err := p1.GetBitmap()
			require.NoError(t, err)
		})

		p2, err := d.GetPage(2)
		require.NoError(t, err)

		t.Run("Segment#9", func(t *testing.T) {
			h := p2.GetSegment(8)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			_, ok := seg.(*segments.PageInformationSegment)
			require.True(t, ok)
		})

		t.Run("Segment#10", func(t *testing.T) {
			h := p2.GetSegment(9)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			sd, ok := seg.(*segments.SymbolDictionary)
			require.True(t, ok)

			dict, err := sd.GetDictionary()
			require.NoError(t, err)

			require.Len(t, dict, 2)
			assert.True(t, cSymbol(t).Equals(dict[0]))
			assert.True(t, aSymbol(t).Equals(dict[1]))
		})

		t.Run("Segment#11", func(t *testing.T) {
			h := p2.GetSegment(10)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			tx, ok := seg.(*segments.TextRegion)
			require.True(t, ok)

			bm, err := tx.GetRegionBitmap()
			require.NoError(t, err)

			expected := bitmap.New(37, 8)
			err = bitmap.Blit(cSymbol(t), expected, 0, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(aSymbol(t), expected, 8, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(pSymbol(t), expected, 16, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(aSymbol(t), expected, 23, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(cSymbol(t), expected, 31, 0, bitmap.CmbOpOr)
			require.NoError(t, err)

			assert.True(t, expected.Equals(bm))
		})

		t.Run("Segment#12", func(t *testing.T) {
			h := p2.GetSegment(11)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			g, ok := seg.(*segments.GenericRegion)
			require.True(t, ok)

			bm, err := g.GetRegionBitmap()
			require.NoError(t, err)

			frame := getFrame(t)
			require.True(t, frame.Equals(bm))
		})

		t.Run("Segment#13", func(t *testing.T) {
			h := p2.GetSegment(12)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			pd, ok := seg.(*segments.PatternDictionary)
			require.True(t, ok)

			dict, err := pd.GetDictionary()
			require.NoError(t, err)

			checkPatternDictionary(t, dict)
		})

		t.Run("Segment#14", func(t *testing.T) {
			h := p2.GetSegment(13)
			require.NotNil(t, h)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			ht, ok := seg.(*segments.HalftoneRegion)
			require.True(t, ok)

			bm, err := ht.GetRegionBitmap()
			require.NoError(t, err)

			assert.Equal(t, 32, bm.Width)
			assert.Equal(t, 36, bm.Height)
		})

		// EOP segment
		t.Run("Page#2", func(t *testing.T) {
			bm, err := p2.GetBitmap()
			require.NoError(t, err)

			assert.NotEmpty(t, bm.Data)
		})

		p3, err := d.GetPage(3)
		require.NoError(t, err)
		require.NotNil(t, p3)

		t.Run("Segment#16", func(t *testing.T) {
			h := p3.GetSegment(15)

			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			pi, ok := seg.(*segments.PageInformationSegment)
			require.True(t, ok)

			assert.Equal(t, 37, pi.PageBMWidth)
			assert.Equal(t, 8, pi.PageBMHeight)
		})

		t.Run("Segment#17", func(t *testing.T) {
			h := p3.GetSegment(16)
			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			sd, ok := seg.(*segments.SymbolDictionary)
			require.True(t, ok)

			dict, err := sd.GetDictionary()
			require.NoError(t, err)
			require.Len(t, dict, 1)

			assert.True(t, dict[0].Equals(aSymbol(t)))
		})

		t.Run("Segment#18", func(t *testing.T) {
			h := p3.GetSegment(17)
			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			sd, ok := seg.(*segments.SymbolDictionary)
			require.True(t, ok)

			dict, err := sd.GetDictionary()
			require.NoError(t, err)

			require.Len(t, dict, 3)

			assert.True(t, dict[0].Equals(aSymbol(t)))
			assert.True(t, dict[1].Equals(cSymbol(t)))
			expected := bitmap.New(14, 6)

			require.NoError(t, bitmap.Blit(aSymbol(t), expected, 0, 0, bitmap.CmbOpOr))
			require.NoError(t, bitmap.Blit(cSymbol(t), expected, 8, 0, bitmap.CmbOpOr))
			assert.True(t, dict[2].Equals(expected), expected.String())
		})

		t.Run("Segment#19", func(t *testing.T) {
			h := p3.GetSegment(18)
			seg, err := h.GetSegmentData()
			require.NoError(t, err)

			tr, ok := seg.(*segments.TextRegion)
			require.True(t, ok)

			bm, err := tr.GetRegionBitmap()
			require.NoError(t, err)

			expected := bitmap.New(37, 8)
			err = bitmap.Blit(cSymbol(t), expected, 0, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(aSymbol(t), expected, 8, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(pSymbol(t), expected, 16, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(aSymbol(t), expected, 23, 0, bitmap.CmbOpOr)
			require.NoError(t, err)
			err = bitmap.Blit(cSymbol(t), expected, 31, 0, bitmap.CmbOpOr)
			require.NoError(t, err)

			assert.True(t, expected.Equals(bm))
		})
	})

	t.Run("AdobeExample", func(t *testing.T) {
		globalsData := []byte{
			// Symbol Dictionary Segment
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x32, 0x00, 0x00, 0x03, 0xFF, 0xFD, 0xFF,
			0x02, 0xFE, 0xFE, 0xFE, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x2A, 0xE2, 0x25,
			0xAE, 0xA9, 0xA5, 0xA5, 0x38, 0xB4, 0xD9, 0x99, 0x9C, 0x5C, 0x8E, 0x56, 0xEF, 0x0F, 0x87,
			0x27, 0xF2, 0xB5, 0x3D, 0x4E, 0x37, 0xEF, 0x79, 0x5C, 0xC5, 0x50, 0x6D, 0xFF, 0xAC,
		}

		gdoc, err := NewDocument(globalsData)
		require.NoError(t, err)

		data := []byte{
			// File Header
			0x97, 0x4A, 0x42, 0x32, 0x0D, 0x0A, 0x1A, 0x0A, 0x01, 0x00, 0x00, 0x00, 0x01,

			// Page Information Segment
			0x00, 0x00, 0x00, 0x01, 0x30, 0x00, 0x01, 0x00, 0x00, 0x00, 0x13, 0x00, 0x00, 0x00, 0x34,
			0x00, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00,

			// Text Region Segment
			0x00, 0x00, 0x00, 0x02, 0x06, 0x20, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1E, 0x00, 0x00, 0x00,
			0x34, 0x00, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00,
			0x10, 0x00, 0x00, 0x00, 0x02, 0x31, 0xDB, 0x51, 0xCE, 0x51, 0xFF, 0xAC,

			// EOP segment
			0x00, 0x00, 0x00, 0x03, 0x31, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,

			// EOF Segment
			0x00, 0x00, 0x00, 0x04, 0x33, 0x01, 0x00, 0x00, 0x00, 0x00,
		}

		// get the document
		d, err := NewDocumentWithGlobals(data, gdoc.GlobalSegments)
		require.NoError(t, err)

		assert.Len(t, d.GlobalSegments, 1)
	})
}

func pSymbol(t *testing.T) *bitmap.Bitmap {
	t.Helper()
	symbol := bitmap.New(5, 8)
	require.NoError(t, symbol.SetPixel(0, 0, 1))
	require.NoError(t, symbol.SetPixel(1, 0, 1))
	require.NoError(t, symbol.SetPixel(2, 0, 1))
	require.NoError(t, symbol.SetPixel(3, 0, 1))
	require.NoError(t, symbol.SetPixel(4, 1, 1))
	require.NoError(t, symbol.SetPixel(0, 1, 1))
	require.NoError(t, symbol.SetPixel(4, 2, 1))
	require.NoError(t, symbol.SetPixel(0, 2, 1))
	require.NoError(t, symbol.SetPixel(4, 3, 1))
	require.NoError(t, symbol.SetPixel(0, 3, 1))
	require.NoError(t, symbol.SetPixel(0, 4, 1))
	require.NoError(t, symbol.SetPixel(1, 4, 1))
	require.NoError(t, symbol.SetPixel(2, 4, 1))
	require.NoError(t, symbol.SetPixel(3, 4, 1))
	require.NoError(t, symbol.SetPixel(0, 5, 1))
	require.NoError(t, symbol.SetPixel(0, 6, 1))
	require.NoError(t, symbol.SetPixel(0, 7, 1))
	return symbol
}

func aSymbol(t *testing.T) *bitmap.Bitmap {
	t.Helper()
	a := bitmap.New(6, 6)
	require.NoError(t, a.SetPixel(1, 0, 1))
	require.NoError(t, a.SetPixel(2, 0, 1))
	require.NoError(t, a.SetPixel(3, 0, 1))
	require.NoError(t, a.SetPixel(4, 0, 1))
	require.NoError(t, a.SetPixel(5, 1, 1))
	require.NoError(t, a.SetPixel(1, 2, 1))
	require.NoError(t, a.SetPixel(2, 2, 1))
	require.NoError(t, a.SetPixel(3, 2, 1))
	require.NoError(t, a.SetPixel(4, 2, 1))
	require.NoError(t, a.SetPixel(5, 2, 1))
	require.NoError(t, a.SetPixel(0, 3, 1))
	require.NoError(t, a.SetPixel(5, 3, 1))
	require.NoError(t, a.SetPixel(0, 4, 1))
	require.NoError(t, a.SetPixel(5, 4, 1))
	require.NoError(t, a.SetPixel(1, 5, 1))
	require.NoError(t, a.SetPixel(2, 5, 1))
	require.NoError(t, a.SetPixel(3, 5, 1))
	require.NoError(t, a.SetPixel(4, 5, 1))
	require.NoError(t, a.SetPixel(5, 5, 1))
	return a
}

func cSymbol(t *testing.T) *bitmap.Bitmap {
	t.Helper()
	c := bitmap.New(6, 6)
	require.NoError(t, c.SetPixel(1, 0, 1))
	require.NoError(t, c.SetPixel(2, 0, 1))
	require.NoError(t, c.SetPixel(3, 0, 1))
	require.NoError(t, c.SetPixel(4, 0, 1))
	require.NoError(t, c.SetPixel(0, 1, 1))
	require.NoError(t, c.SetPixel(5, 1, 1))
	require.NoError(t, c.SetPixel(0, 2, 1))
	require.NoError(t, c.SetPixel(0, 3, 1))
	require.NoError(t, c.SetPixel(0, 4, 1))
	require.NoError(t, c.SetPixel(5, 4, 1))
	require.NoError(t, c.SetPixel(1, 5, 1))
	require.NoError(t, c.SetPixel(2, 5, 1))
	require.NoError(t, c.SetPixel(3, 5, 1))
	require.NoError(t, c.SetPixel(4, 5, 1))
	return c
}

func getFrame(t *testing.T) *bitmap.Bitmap {
	t.Helper()
	expected := bitmap.New(54, 44)
	for y := 0; y < expected.Height; y++ {
		switch y {
		case 0, 1, expected.Height - 1, expected.Height - 2:
			// full
			for x := 0; x < expected.Width; x++ {
				require.NoError(t, expected.SetPixel(x, y, 1))
			}
		default:
			require.NoError(t, expected.SetPixel(0, y, 1))
			require.NoError(t, expected.SetPixel(1, y, 1))
			require.NoError(t, expected.SetPixel(expected.Width-2, y, 1))
			require.NoError(t, expected.SetPixel(expected.Width-1, y, 1))
		}
	}
	return expected
}

func getPatternsFirst(t *testing.T) (patterns []*bitmap.Bitmap) {
	t.Helper()

	for i := 0; i < 16; i++ {
		toCompare := bitmap.New(4, 4)
		switch i {
		case 15:
			require.NoError(t, toCompare.SetPixel(0, 3, 1))
			fallthrough
		case 14:
			require.NoError(t, toCompare.SetPixel(3, 3, 1))
			fallthrough
		case 13:
			require.NoError(t, toCompare.SetPixel(3, 0, 1))
			fallthrough
		case 12:
			require.NoError(t, toCompare.SetPixel(0, 0, 1))
			fallthrough
		case 11:
			require.NoError(t, toCompare.SetPixel(3, 1, 1))
			fallthrough
		case 10:
			require.NoError(t, toCompare.SetPixel(2, 3, 1))
			fallthrough
		case 9:
			require.NoError(t, toCompare.SetPixel(1, 0, 1))
			require.NoError(t, toCompare.SetPixel(0, 2, 1))
			fallthrough
		case 8:
			require.NoError(t, toCompare.SetPixel(3, 2, 1))
			fallthrough
		case 7:
			require.NoError(t, toCompare.SetPixel(1, 3, 1))
			fallthrough
		case 6:
			require.NoError(t, toCompare.SetPixel(0, 1, 1))
			fallthrough
		case 5:
			require.NoError(t, toCompare.SetPixel(2, 0, 1))
			fallthrough
		case 4:
			require.NoError(t, toCompare.SetPixel(1, 2, 1))
			fallthrough
		case 3:
			require.NoError(t, toCompare.SetPixel(2, 2, 1))
			fallthrough
		case 2:
			require.NoError(t, toCompare.SetPixel(1, 1, 1))
			fallthrough
		case 1:
			require.NoError(t, toCompare.SetPixel(2, 1, 1))
		}
		patterns = append(patterns, toCompare)
	}
	return
}

func checkPatternDictionary(t *testing.T, dict []*bitmap.Bitmap) {
	t.Helper()
	for i, s := range dict {
		toCompare := bitmap.New(4, 4)
		switch i {
		case 15:
			require.NoError(t, toCompare.SetPixel(0, 3, 1))
			fallthrough
		case 14:
			require.NoError(t, toCompare.SetPixel(3, 3, 1))
			fallthrough
		case 13:
			require.NoError(t, toCompare.SetPixel(3, 0, 1))
			fallthrough
		case 12:
			require.NoError(t, toCompare.SetPixel(0, 0, 1))
			fallthrough
		case 11:
			require.NoError(t, toCompare.SetPixel(3, 1, 1))
			fallthrough
		case 10:
			require.NoError(t, toCompare.SetPixel(2, 3, 1))
			fallthrough
		case 9:
			require.NoError(t, toCompare.SetPixel(1, 0, 1))
			require.NoError(t, toCompare.SetPixel(0, 2, 1))
			fallthrough
		case 8:
			require.NoError(t, toCompare.SetPixel(3, 2, 1))
			fallthrough
		case 7:
			require.NoError(t, toCompare.SetPixel(1, 3, 1))
			fallthrough
		case 6:
			require.NoError(t, toCompare.SetPixel(0, 1, 1))
			fallthrough
		case 5:
			require.NoError(t, toCompare.SetPixel(2, 0, 1))
			fallthrough
		case 4:
			require.NoError(t, toCompare.SetPixel(1, 2, 1))
			fallthrough
		case 3:
			require.NoError(t, toCompare.SetPixel(2, 2, 1))
			fallthrough
		case 2:
			require.NoError(t, toCompare.SetPixel(1, 1, 1))
			fallthrough
		case 1:
			require.NoError(t, toCompare.SetPixel(2, 1, 1))
		}
		assert.True(t, toCompare.Equals(s), fmt.Sprintf("i: %d, %v, %v", i, s.String(), toCompare.String()))
	}
}
