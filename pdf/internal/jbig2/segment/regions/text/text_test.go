package text

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"testing"
)

func TestTextRegionDecode(t *testing.T) {
	var data []byte = []byte{
		// header
		0x00, 0x00, 0x00, 0x03, 0x07, 0x42, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x31,

		// data part
		0x00, 0x00, 0x25, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x04, 0x00, 0x00, 0x00, 0x01, 0x00, 0x0C, 0x09, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x05, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x0C, 0x40, 0x07, 0x08, 0x70, 0x41, 0xD0,
	}

	r := reader.New(data)

	h := &header.Header{}
	_, err := h.Decode(r)
	require.NoError(t, err)

	assert.Equal(t, kind.ImmediateLosslessTextRegion, kind.SegmentKind(h.SegmentType))
	assert.True(t, h.PageAssociationSizeSet)
	if assert.Equal(t, 2, h.ReferredToSegmentCount) {
		assert.Contains(t, 0, h.ReferredToSegments)
		assert.Contains(t, 2, h.ReferredToSegments)
	}

	assert.Equal(t, 49, h.DataLength)

	d := container.New()

	ts := New(d, h, true)
	err = ts.Decode(r)
}
