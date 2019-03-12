package patterndict

import (
	"github.com/stretchr/testify/assert"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"testing"
)

func TestDecodePatternDictionary(t *testing.T) {

	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	// Annex H - segment 6th
	t.Run("6thAnnexH", func(t *testing.T) {
		var data []byte = []byte{
			// Header
			0x00, 0x00, 0x00, 0x05, 0x10, 0x01, 0x01, 0x00, 0x00, 0x00, 0x2D,

			// Data part
			0x01, 0x04, 0x04, 0x00, 0x00, 0x00, 0x0F, 0x20, 0xD1, 0x84,
			0x61, 0x18, 0x45, 0xF2, 0xF9, 0x7C, 0x8F, 0x11, 0xC3, 0x9E,
			0x45, 0xF2, 0xF9, 0x7D, 0x42, 0x85, 0x0A, 0xAA, 0x84, 0x62,
			0x2F, 0xEE, 0xEC, 0x44, 0x62, 0x22, 0x35, 0x2A, 0x0A, 0x83,
			0xB9, 0xDC, 0xEE, 0x77, 0x80,
		}

		// set the reader
		r := reader.New(data)

		// create header
		h := &header.Header{}

		_, err := h.Decode(r)
		if assert.NoError(t, err) {
			assert.Equal(t, 5, h.SegmentNumber)
			assert.Equal(t, kind.PatternDictionary, kind.SegmentKind(h.SegmentType))
			assert.False(t, h.DeferredNonRetainSet)
			assert.False(t, h.PageAssociationSizeSet)

			assert.Equal(t, 0, h.ReferredToSegmentCount)
			assert.Equal(t, 45, h.DataLength)

			p := New(container.New(), h)

			// Decode the pattern dictionary segment
			err = p.Decode(r)
			if assert.NoError(t, err) {

				patterns := [][][]int{
					{
						{0, 0, 0, 0},
						{0, 0, 0, 0},
						{0, 0, 0, 0},
						{0, 0, 0, 0},
					},
					{
						{0, 0, 0, 0},
						{0, 0, 1, 0},
						{0, 0, 0, 0},
						{0, 0, 0, 0},
					},
					{
						{0, 0, 0, 0},
						{0, 1, 1, 0},
						{0, 0, 0, 0},
						{0, 0, 0, 0},
					},
					{
						{0, 0, 0, 0},
						{0, 1, 1, 0},
						{0, 0, 1, 0},
						{0, 0, 0, 0},
					},
					{
						{0, 0, 0, 0},
						{0, 1, 1, 0},
						{0, 1, 1, 0},
						{0, 0, 0, 0},
					},
					{
						{0, 0, 1, 0},
						{0, 1, 1, 0},
						{0, 1, 1, 0},
						{0, 0, 0, 0},
					},
					{
						{0, 0, 1, 0},
						{1, 1, 1, 0},
						{0, 1, 1, 0},
						{0, 0, 0, 0},
					},

					{
						{0, 0, 1, 0},
						{1, 1, 1, 0},
						{0, 1, 1, 0},
						{0, 1, 0, 0},
					},

					{
						{0, 0, 1, 0},
						{1, 1, 1, 0},
						{0, 1, 1, 1},
						{0, 1, 0, 0},
					},
					{
						{0, 1, 1, 0},
						{1, 1, 1, 0},
						{1, 1, 1, 1},
						{0, 1, 0, 0},
					},
					{
						{0, 1, 1, 0},
						{1, 1, 1, 0},
						{1, 1, 1, 1},
						{0, 1, 1, 0},
					},
					{
						{0, 1, 1, 0},
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{0, 1, 1, 0},
					},
					{
						{1, 1, 1, 0},
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{0, 1, 1, 0},
					},
					{
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{0, 1, 1, 0},
					},
					{
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{0, 1, 1, 1},
					},
					{
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{1, 1, 1, 1},
						{1, 1, 1, 1},
					},
				}

				if assert.Len(t, p.Bitmaps, len(patterns)) {
					for i := range patterns {
						bs := patternToBitset(t, patterns[i], 16)
						if assert.NotNil(t, p.Bitmaps[i]) {
							t.Logf("Bitmap at '%d': %s", i, p.Bitmaps[0].Data)
							assert.True(t, p.Bitmaps[i].Data.Equals(bs))
						}
					}
				}

			}
		}
	})

}

func patternToBitset(t *testing.T, pattern [][]int, length int) *bitmap.BitSet {
	t.Helper()
	bs := bitmap.NewBitSet(length)
	for i := 0; i < len(pattern); i++ {
		for j := 0; j < len(pattern[i]); j++ {
			val := pattern[i][j]
			err := bs.Set(uint(i*len(pattern[i])+j), val == 1)
			if !assert.NoError(t, err) {
				return nil
			}
		}
	}
	return bs
}
