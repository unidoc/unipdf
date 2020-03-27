/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// TestDecodePatternDictionary tests the decode process of the pattern dictionary segment.
func TestDecodePatternDictionary(t *testing.T) {
	t.Run("AnnexH", func(t *testing.T) {
		if testing.Verbose() {
			common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
		}

		t.Run("S-6th", func(t *testing.T) {
			data := []byte{
				// Header
				0x00, 0x00, 0x00, 0x05, 0x10, 0x01, 0x01, 0x00, 0x00, 0x00, 0x2D,

				// Data part
				0x01, 0x04, 0x04, 0x00, 0x00, 0x00, 0x0F, 0x20, 0xD1, 0x84,
				0x61, 0x18, 0x45, 0xF2, 0xF9, 0x7C, 0x8F, 0x11, 0xC3, 0x9E,
				0x45, 0xF2, 0xF9, 0x7D, 0x42, 0x85, 0x0A, 0xAA, 0x84, 0x62,
				0x2F, 0xEE, 0xEC, 0x44, 0x62, 0x22, 0x35, 0x2A, 0x0A, 0x83,
				0xB9, 0xDC, 0xEE, 0x77, 0x80,
			}
			r := reader.New(data)
			d := &document{}
			h, err := NewHeader(d, r, 0, OSequential)
			require.NoError(t, err)

			ps, err := h.GetSegmentData()
			require.NoError(t, err)

			p, ok := ps.(*PatternDictionary)
			require.True(t, ok)

			assert.Equal(t, uint32(5), h.SegmentNumber)
			assert.Equal(t, uint64(45), h.SegmentDataLength)
			assert.Equal(t, 1, h.PageAssociation)

			require.Equal(t, true, p.IsMMREncoded)
			require.Equal(t, byte(4), p.HdpWidth)
			require.Equal(t, byte(4), p.HdpHeight)
			require.Equal(t, uint32(15), p.GrayMax)

			dict, err := p.GetDictionary()
			require.NoError(t, err)

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
		})
	})
}
