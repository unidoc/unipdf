package pageinformation

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"testing"
)

func TestPageInformationDecode(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}
	t.Run("2ndAnnexH", func(t *testing.T) {
		var data []byte = []byte{

			// header
			0x00, 0x00, 0x00, 0x01, 0x30, 0x00, 0x01, 0x00, 0x00, 0x00, 0x13,

			// data part
			0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x38, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}

		r := reader.New(data)

		h := &header.Header{}

		_, err := h.Decode(r)
		require.NoError(t, err)

		assert.Equal(t, kind.PageInformation, kind.SegmentKind(h.SegmentType))
		assert.False(t, h.PageAssociationSizeSet)
		assert.False(t, h.DeferredNonRetainSet)
		assert.Equal(t, 0, h.ReferredToSegmentCount)
		assert.Equal(t, 1, h.PageAssociation)
		assert.Equal(t, 19, h.DataLength)

		s := New(container.New(), h)
		err = s.Decode(r)
		if assert.NoError(t, err) {
			assert.Equal(t, 64, s.PageBMWidth)
			assert.Equal(t, 56, s.PageBMHeight)
			assert.Equal(t, 0, s.PageInfoFlags.GetValue(DefaultPixelValue))
			assert.Equal(t, 0, s.PageInfoFlags.GetValue(DefaultCombinationOperator))
			assert.Equal(t, 0, s.pageStripping)

			common.Log.Debug("Page bitmap: \n%s", s.PageBitmap)
		}

	})
}
