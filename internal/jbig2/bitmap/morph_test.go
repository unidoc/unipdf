/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCentroid tests the centroid function.
func TestCentroid(t *testing.T) {
	// The centroid is based on the pixel number of the bitmap.
	// Having the bitmap 10x10 with 2 pix border.
	bm, err := New(8, 8).AddBorder(2, 1)
	require.NoError(t, err)

	pt, err := Centroid(bm, nil, nil)
	require.NoError(t, err)

	// the following for-loops computes in less efficient way the 'xsum', 'ysum' and 'pixsum'
	// Where:
	// 	point.X = float32(xsum)/float32(pixsum)
	//	point.Y = float32(ysum)/float32(pixsum)
	var xsum, ysum, pixsum int
	for i := 0; i < bm.Height; i++ {
		for j := 0; j < bm.Width; j++ {
			if bm.GetPixel(j, i) {
				xsum += j
				ysum += i
				pixsum++
			}
		}
	}
	ptCompare := Point{float32(xsum) / float32(pixsum), float32(ysum) / float32(pixsum)}

	// thus the 'pt' and 'ptCompare' should have equal 'X' and 'Y' values.
	assert.Equal(t, ptCompare, pt)
}
