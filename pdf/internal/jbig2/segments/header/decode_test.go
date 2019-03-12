package header

import (
	"github.com/stretchr/testify/assert"
	// "github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"io"
	"testing"
)

func TestSegmentHeader(t *testing.T) {
	// common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	t.Run("7.2.8/1", func(t *testing.T) {

		data := []byte{
			0x00, 0x00, 0x00, 0x20, // segment number
			0x86,             // header flags
			0x6B,             // referred-to segment count and retention flags
			0x02, 0x1E, 0x05, // referred-to segment numbers
			0x04, // page association number
		}
		buf := reader.New(data)

		h := &Header{}
		n, err := h.Decode(buf)
		if err == io.EOF || err == nil {
			assert.Equal(t, len(data), n)
			assert.Equal(t, 32, h.SegmentNumber)
			assert.False(t, h.PageAssociationSizeSet)
			assert.Equal(t, kind.ImmediateTextRegion, kind.SegmentKind(h.SegmentType))
			assert.True(t, h.DeferredNonRetainSet)
			assert.Equal(t, 3, h.ReferredToSegmentCount)
			assert.Equal(t, []int{0x0B}, h.RententionFlags)
			assert.Equal(t, []int{2, 30, 5}, h.ReferredToSegments)
			assert.Equal(t, 4, h.PageAssociation)
		}
	})
	t.Run("7.2.8/2", func(t *testing.T) {
		data := []byte{
			0x00, 0x00, 0x02, 0x34, // segment header
			0x40,                               // header flags
			0xE0, 0x00, 0x00, 0x09, 0x02, 0xFD, // referred-to segment count and retention flags
			// referred-to segment numbers, 2 bytes for each numbers because the
			// current segment number is strictly greater than 256
			0x01, 0x00,
			0x00, 0x02,
			0x00, 0x1E,
			0x00, 0x05,
			0x02, 0x00,
			0x02, 0x01,
			0x02, 0x02,
			0x02, 0x03,
			0x02, 0x04,
			// page association number
			0x00, 0x00, 0x04, 0x01,
		}
		buf := reader.New(data)

		h := &Header{}
		n, err := h.Decode(buf)
		if err == io.EOF || err == nil {
			assert.Equal(t, n, len(data))
			assert.Equal(t, 564, h.SegmentNumber)
			assert.Equal(t, kind.SymbolDictionary, kind.SegmentKind(h.SegmentType))
			assert.True(t, h.PageAssociationSizeSet)
			assert.False(t, h.DeferredNonRetainSet)
			assert.Equal(t, 9, h.ReferredToSegmentCount)
			if assert.NotEqual(t, 0, len(h.RententionFlags)) {
				assert.Equal(t, h.RententionFlags[0], 0x02)
				assert.Equal(t, h.RententionFlags[1], 0xFD)
			}
			assert.Equal(t, 1025, h.PageAssociation)
			t.Logf("Referred Segments: %v", h.ReferredToSegments)

		}

	})

	t.Run("Annex H", func(t *testing.T) {
		data := []byte{
			0x00, 0x00, 0x00, 0x00, // segment number
			0x00,                   // header flags
			0x01,                   // referred-to segment count and retention flags
			0x00,                   // page association size
			0x00, 0x00, 0x00, 0x18, // segment data length
		}
		buf := reader.New(data)

		h := &Header{}
		n, err := h.Decode(buf)
		if err == io.EOF || err == nil {
			assert.Equal(t, n, len(data))
			assert.Equal(t, 0, h.SegmentNumber)
			assert.Equal(t, kind.SymbolDictionary, kind.SegmentKind(h.SegmentType))
			assert.False(t, h.PageAssociationSizeSet)
			assert.Equal(t, 0, h.ReferredToSegmentCount)
			assert.Equal(t, h.RententionFlags, []int{0x01})
			assert.Equal(t, 0, h.PageAssociation)
			assert.Equal(t, 24, h.DataLength)
		}
	})
}

func BenchmarkHeaderDecode(b *testing.B) {

	headerDecoder := func(b *testing.B, data []byte) {
		b.Helper()
		r := reader.New(data)
		h := &Header{}
		_, err := h.Decode(r)
		if err != nil {
			if err != io.EOF {
				assert.NoError(b, err)
			}
		}

	}

	testCases := map[string][]byte{
		"AnnexH": {
			0x00, 0x00, 0x00, 0x00, // segment number
			0x00, // header flags
			0x01, // referred-to segment count and retention flags
			0x00, // page association size
			0x00, 0x00, 0x00, 0x18,
		},
		"Short": {
			0x00, 0x00, 0x00, 0x20, // segment number
			0x86,             // header flags
			0x6B,             // referred-to segment count and retention flags
			0x02, 0x1E, 0x05, // referred-to segment numbers
			0x04, // page association number
		},
		"Long": {
			0x00, 0x00, 0x02, 0x34, // segment header
			0x40,                               // header flags
			0xE0, 0x00, 0x00, 0x09, 0x02, 0xFD, // referred-to segment count and retention flags
			// referred-to segment numbers, 2 bytes for each numbers because the
			// current segment number is strictly greater than 256
			0x01, 0x00,
			0x00, 0x02,
			0x00, 0x1E,
			0x00, 0x05,
			0x02, 0x00,
			0x02, 0x01,
			0x02, 0x02,
			0x02, 0x03,
			0x02, 0x04,
			// page association number
			0x00, 0x00, 0x04, 0x01,
		},
	}

	for name, data := range testCases {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				headerDecoder(b, data)
			}
		})
	}
}
