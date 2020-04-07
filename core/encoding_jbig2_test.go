/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
)

// TestImageToJBIG2Image tests conversion of image.Image to JBIG2Image
func TestImageToJBIG2Image(t *testing.T) {
	t.Run("BlackWhite", func(t *testing.T) {
		// having a test black white image of a frame.
		// The frame has a width of 2 bits.
		g := image.NewGray(image.Rect(0, 0, 50, 50))
		bounds := g.Bounds()
		bm := bitmap.New(50, 50)
		setPix := func(x, y int) {
			g.SetGray(x, y, color.Gray{})
			assert.NoError(t, bm.SetPixel(x, y, 1))
		}
		for x := 0; x < bounds.Dx(); x++ {
			for y := 0; y < bounds.Dy(); y++ {
				switch x {
				case 0, 1, 48, 49:
					setPix(x, y)
				default:
					if !(y > 1 && y < 48) {
						setPix(x, y)
					} else {
						g.SetGray(x, y, color.Gray{Y: 255})
					}
				}
			}
		}

		// execute GoImageToJBIG2 and check jbig2 images.
		jb2, err := GoImageToJBIG2(g, JB2ImageAutoThreshold)
		require.NoError(t, err)

		assert.Equal(t, jb2.Data, bm.Data)
	})
}
