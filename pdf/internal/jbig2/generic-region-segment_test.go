package jbig2

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"testing"
)

func TestDecodeGenericRegionSegment(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}
	t.Run("AnnexH", func(t *testing.T) {
		t.Run("S-12", func(t *testing.T) {
			// headerData := []byte{
			//
			// }

			data := []byte{
				// header
				0x00, 0x00, 0x00, 0x0B, 0x27, 0x00, 0x02, 0x00, 0x00, 0x00, 0x23,

				// data
				0x00, 0x00, 0x00, 0x36, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x04,
				0x00, 0x00, 0x00, 0x0B, 0x00, 0x08, 0x03, 0xFF, 0xFD, 0xFF, 0x02, 0xFE,
				0xFE, 0xFE, 0x04, 0xEE, 0xED, 0x87, 0xFB, 0xCB, 0x2B, 0xFF, 0xAC,
			}

			r := reader.New(data)

			d := &Document{InputStream: r}

			h, err := NewHeader(d, r, 0, OSequential)
			require.NoError(t, err)

			assert.Equal(t, uint32(11), h.SegmentNumber)
			assert.Equal(t, uint64(35), h.SegmentDataLength)
			assert.Equal(t, 2, h.PageAssociation)

			s, err := NewGenericRegionSegment(h, r)
			require.NoError(t, err)

			assert.Equal(t, 44, s.RegionSegment.BitmapHeight)
			assert.Equal(t, 54, s.RegionSegment.BitmapWidth)
			assert.Equal(t, bitmap.CmbOpOr, s.RegionSegment.CombinaionOperator)
			assert.Equal(t, true, s.IsTPGDon)
			assert.Equal(t, byte(0), s.GBTemplate)

			bm, err := s.GetRegionBitmap()
			require.NoError(t, err)

			t.Log(bm.String())

		})
	})
}
