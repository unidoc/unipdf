/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
)

// TestExtract tests the extraction of the image.Rectangle from the source bitmap
func TestExtract(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	t.Run("SmallerSize", func(t *testing.T) {
		// Tests the extraction when the dimensions are smaller than in the source bitmap
		t.Run("SmallerWidth", func(t *testing.T) {
			// create the 10x10 bitmap
			src := New(10, 10)

			// set some pixels for testing purpose
			require.NoError(t, src.SetPixel(4, 9, 1))
			require.NoError(t, src.SetPixel(0, 0, 1))

			// create the rectangle 5x10
			roi := image.Rect(0, 0, 5, 10)

			res, err := Extract(roi, src)
			require.NoError(t, err)

			assert.Equal(t, 5, res.Width)
			assert.Equal(t, 10, res.Height)
			assert.True(t, res.GetPixel(0, 0))
			assert.True(t, res.GetPixel(4, 9))
		})

		t.Run("SmallerHeight", func(t *testing.T) {
			// create the 10x10 bitmap
			src := New(10, 10)

			require.NoError(t, src.SetPixel(0, 0, 1))
			require.NoError(t, src.SetPixel(3, 3, 1))

			// create the rectangle 10x5
			roi := image.Rect(0, 0, 10, 5)

			// extract the roi from source
			res, err := Extract(roi, src)
			require.NoError(t, err)

			assert.Equal(t, 10, res.Width)
			assert.Equal(t, 5, res.Height)
			assert.True(t, res.GetPixel(0, 0))
			assert.True(t, res.GetPixel(3, 3))
		})

		t.Run("BothSmaller", func(t *testing.T) {
			// create the 10x10 bitmap
			src := New(10, 10)

			require.NoError(t, src.SetPixel(0, 0, 1))
			require.NoError(t, src.SetPixel(5, 5, 1))

			// create the extraction rectangle 6x6
			roi := image.Rect(0, 0, 6, 6)

			// extract the resultant bitmap
			res, err := Extract(roi, src)
			require.NoError(t, err)

			assert.Equal(t, 6, res.Width)
			assert.Equal(t, 6, res.Height)
			assert.True(t, res.GetPixel(0, 0))
			assert.True(t, res.GetPixel(5, 5))
		})

		t.Run("Shifted", func(t *testing.T) {
			// create the 10x10 bitmap
			src := New(10, 10)

			// (3,3) and (5,5) is set to true
			require.NoError(t, src.SetPixel(3, 3, 1))
			require.NoError(t, src.SetPixel(5, 5, 1))

			// create the image 3x3 shifted by [3,3]
			roi := image.Rect(3, 3, 6, 6)

			// Extract the rectangle bitmap
			res, err := Extract(roi, src)
			require.NoError(t, err)

			// check the dimensions
			assert.Equal(t, 3, res.Width)
			assert.Equal(t, 3, res.Height)

			// check the shifted pixels Before (3,3) shifted (-3,-3) -> (0,0)
			assert.True(t, res.GetPixel(0, 0))
			assert.True(t, res.GetPixel(2, 2))
		})

		t.Run("BigShifted", func(t *testing.T) {
			// create the 50x50 bitmap
			src := New(50, 50)

			require.NoError(t, src.SetPixel(25, 25, 1))
			require.NoError(t, src.SetPixel(15, 15, 1))

			// create the shifted rectangle
			roi := image.Rect(12, 12, 30, 30)

			// Extract the image with copyLine
			res, err := Extract(roi, src)
			require.NoError(t, err)

			assert.Equal(t, 18, res.Height)
			assert.Equal(t, 18, res.Width)

			assert.True(t, res.GetPixel(3, 3))
			assert.True(t, res.GetPixel(13, 13))
		})

		t.Run("GreaterShifted", func(t *testing.T) {
			// create the 50x50 bitmap
			src := New(64, 64)

			require.NoError(t, src.SetPixel(25, 25, 1))
			require.NoError(t, src.SetPixel(15, 15, 1))

			// create the shifted rectangle
			roi := image.Rect(12, 12, 56, 56)

			// Extract the image with copyLine
			res, err := Extract(roi, src)
			require.NoError(t, err)

			assert.Equal(t, 44, res.Height)
			assert.Equal(t, 44, res.Width)

			assert.True(t, res.GetPixel(3, 3))
			assert.True(t, res.GetPixel(13, 13))
		})
	})

	t.Run("EqualSize", func(t *testing.T) {
		// Tests the extraction when the extractino is of equal size as the source bitmap
		src := New(5, 5)

		require.NoError(t, src.SetPixel(3, 3, 1))
		require.NoError(t, src.SetPixel(1, 2, 1))

		// create the 5x5 rectangle image
		roi := image.Rect(0, 0, 5, 5)

		// extract the bitmap
		res, err := Extract(roi, src)
		require.NoError(t, err)

		// check the sizes
		assert.Equal(t, 5, res.Width)
		assert.Equal(t, 5, res.Height)

		// check the pixels
		assert.True(t, res.GetPixel(3, 3))
		assert.True(t, res.GetPixel(1, 2))
		assert.False(t, res.GetPixel(0, 0))

		// these should have equal values but be different bitmaps
		assert.True(t, res.Equals(src))

		// the result should be a different struct
		assert.False(t, res == src)
	})

	t.Run("GreaterSize", func(t *testing.T) {
		// Tests the extraction when the rectangle is greater than the source bitmap
		src := New(5, 5)

		require.NoError(t, src.SetPixel(2, 2, 1))
		require.NoError(t, src.SetPixel(1, 4, 1))

		// create the rectanle 10x10 with base at (0,0)
		roi := image.Rect(0, 0, 10, 10)

		// extract the rectangle image
		_, err := Extract(roi, src)
		require.Error(t, err)
	})
}
