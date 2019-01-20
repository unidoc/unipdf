package symboldict

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"

	// "github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"

	"io"
	"testing"
)

func TestSymbolDictionaryHeader(t *testing.T) {
	if testing.Verbose() {
		common.Log = common.NewConsoleLogger(common.LogLevelDebug)
	}

	t.Run("10thHeaderAnnexH", func(t *testing.T) {
		data := []byte{
			0x00, 0x00, 0x00, 0x09, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x1B,
		}

		buf := reader.New(data)

		h := &header.Header{}

		n, err := h.Decode(buf)
		if assert.True(t, err == nil || err == io.EOF) {
			assert.Equal(t, len(data), n)
			assert.Equal(t, 9, h.SegmentNumber)
			assert.Equal(t, kind.SymbolDictionary, kind.SegmentKind(h.SegmentType))
			assert.False(t, h.PageAssociationSizeSet)
			assert.Equal(t, 0, h.ReferredToSegmentCount)
			assert.Equal(t, []int{0x01}, h.RententionFlags)
			assert.Equal(t, 2, h.PageAssociation)
			assert.Equal(t, 27, h.DataLength)
		}
	})

	t.Run("10thAnnexH", func(t *testing.T) {

		data := []byte{
			// Header
			0x00, 0x00, 0x00, 0x09, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x1B,

			// Segment data
			0x08, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02,
			0x4F, 0xE7, 0x8C, 0x20, 0x0E, 0x1D, 0xC7, 0xCF, 0x01, 0x11, 0xC4, 0xB2,
			0x6F, 0xFF, 0xAC,
		}

		h := &header.Header{}

		r := reader.New(data)

		_, err := h.Decode(r)
		if assert.NoError(t, err) {
			assert.Equal(t, 9, h.SegmentNumber)
			assert.Equal(t, kind.SymbolDictionary, kind.SegmentKind(h.SegmentType))
			assert.False(t, h.PageAssociationSizeSet)
			assert.Equal(t, 0, h.ReferredToSegmentCount)
			assert.Equal(t, []int{0x01}, h.RententionFlags)
			assert.Equal(t, 2, h.PageAssociation)
			assert.Equal(t, 27, h.DataLength)

			s := New(container.New(), h)
			err = s.Decode(r)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, s.SDFlags.GetValue(SD_HUFF))
				assert.Equal(t, 0, s.SDFlags.GetValue(SD_REF_AGG))
				assert.Equal(t, 2, s.SDFlags.GetValue(SD_TEMPLATE))
				assert.Equal(t, 0, s.SDFlags.GetValue(SD_R_TEMPLATE))
				assert.Equal(t, 0, s.SDFlags.GetValue(BITMAP_CC_USED))
				assert.Equal(t, 0, s.SDFlags.GetValue(BITMAP_CC_RETAINED))
				if assert.Len(t, s.AdaptiveTemplateX, 4) {
					assert.Equal(t, int8(2), s.AdaptiveTemplateX[0])
				}

				if assert.Len(t, s.AdaptiveTemplateY, 4) {
					assert.Equal(t, int8(-1), s.AdaptiveTemplateY[0])
				}

			}
		}

	})
}

func TestSymbolDictionaryDecode(t *testing.T) {
	if testing.Verbose() {
		common.Log = common.NewConsoleLogger(common.LogLevelDebug)
	}

	t.Run("3rdAnnexHHuffman", func(t *testing.T) {
		// t.Skip("Skipping Huffman")
		if testing.Verbose() {
			common.Log = common.NewConsoleLogger(common.LogLevelDebug)
		}

		data := []byte{
			// Header
			0x00, 0x00, 0x00, 0x02, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x1C,

			// Segment data
			0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0xE5, 0xCD,
			0xF8, 0x00, 0x79, 0xE0, 0x84, 0x10, 0x81, 0xF0, 0x82, 0x10, 0x86, 0x10,
			0x79, 0xF0, 0x00, 0x80, 0x00, 0x00, 0x00, 0x03, 0x07, 0x42, 0x00, 0x02,
			0x01, 0x00, 0x00, 0x00, 0x31, 0x00, 0x00, 0x00, 0x00, 0x25, 0x00,
		}

		h := &header.Header{}

		r := reader.New(data)

		_, err := h.Decode(r)
		if assert.NoError(t, err) {

			s := New(container.New(), h)
			err = s.Decode(r)
			if err != io.EOF {
				require.NoError(t, err)
			}

			assert.Equal(t, 1, s.SDFlags.GetValue(SD_HUFF))
			assert.Equal(t, 0, s.SDFlags.GetValue(SD_REF_AGG))
			assert.Equal(t, uint32(2), s.ExportedSymbolsNumber)
			assert.Equal(t, uint32(2), s.NewSymbolsNumber)

			symbolC := [][]int{
				{0, 1, 1, 1, 1, 0},
				{1, 0, 0, 0, 0, 1},
				{1, 0, 0, 0, 0, 0},
				{1, 0, 0, 0, 0, 0},
				{1, 0, 0, 0, 0, 1},
				{0, 1, 1, 1, 1, 0},
			}

			if assert.Len(t, s.Bitmaps, 2) {
				bs := symbolToBitset(t, symbolC, 36)
				if assert.NotNil(t, bs) {
					assert.True(t, s.Bitmaps[0].Data.Equals(bs))
				}

			}
		}
	})

}

func symbolToBitset(t *testing.T, symbol [][]int, length int) *bitmap.BitSet {
	t.Helper()
	bs := bitmap.NewBitSet(length)
	for i := 0; i < len(symbol); i++ {
		for j := 0; j < len(symbol[i]); j++ {
			val := symbol[i][j]
			err := bs.Set(uint(i*len(symbol[i])+j), val == 1)
			if !assert.NoError(t, err) {
				return nil
			}
		}
	}
	return bs
}

func BenchmarkSymbolDictionaryHeaderDecode(b *testing.B) {

	decodeSymbolDictHeader := func(b *testing.B, data []byte) *SymbolDictionarySegment {
		b.Helper()

		// Get reader
		r := reader.New(data)

		// Create and decode header
		h := &header.Header{}
		var s *SymbolDictionarySegment
		_, err := h.Decode(r)
		if assert.NoError(b, err) {
			s = New(container.New(), h)

			err := s.readFlags(r)
			assert.NoError(b, err)
		}
		return s
	}

	b.Run("AnnexHFirst", func(b *testing.B) {

		var data []byte = []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x18,

			0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0xE9, 0xCB,
			0xF4, 0x00, 0x26, 0xAF, 0x04, 0xBF, 0xF0, 0x78, 0x2F, 0xE0, 0x00, 0x40,
		}

		for i := 0; i < b.N; i++ {
			s := decodeSymbolDictHeader(b, data)
			if assert.NotNil(b, s) {

			}
		}
	})

	b.Run("AnnexHThird", func(b *testing.B) {
		data := []byte{
			// Header
			0x00, 0x00, 0x00, 0x09, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x1B,

			// Segment data
			0x08, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02,
			0x4F, 0xE7, 0x8C, 0x20, 0x0E, 0x1D, 0xC7, 0xCF, 0x01, 0x11, 0xC4, 0xB2,
			0x6F, 0xFF, 0xAC,
		}

		for i := 0; i < b.N; i++ {
			s := decodeSymbolDictHeader(b, data)
			if assert.NotNil(b, s) {

			}
		}

	})

}

func BenchmarkSymbolDictDecode(b *testing.B) {

	decodeSymbolDict := func(b *testing.B, data []byte) {
		b.Helper()
		// Get reader
		r := reader.New(data)

		// Create and decode header
		h := &header.Header{}

		_, err := h.Decode(r)
		if assert.NoError(b, err) {
			s := New(container.New(), h)
			err := s.Decode(r)
			assert.NoError(b, err)
		}
	}

	b.Run("Small", func(b *testing.B) {
		data := []byte{
			// Header
			0x00, 0x00, 0x00, 0x09, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x1B,

			// Segment data
			0x08, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02,
			0x4F, 0xE7, 0x8C, 0x20, 0x0E, 0x1D, 0xC7, 0xCF, 0x01, 0x11, 0xC4, 0xB2,
			0x6F, 0xFF, 0xAC,
		}

		for i := 0; i < b.N; i++ {
			decodeSymbolDict(b, data)
		}
	})

}
