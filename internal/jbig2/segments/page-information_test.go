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
