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
)

func TestClipRectangle(t *testing.T) {
	box := image.Rect(0, 0, 100, 100)

	clipped, err := ClipBoxToRectangle(&box, 50, 50)
	require.NoError(t, err)

	assert.Equal(t, 0, clipped.Min.X)
	assert.Equal(t, 0, clipped.Min.Y)
	assert.Equal(t, 50, clipped.Max.X)
	assert.Equal(t, 50, clipped.Max.Y)
	assert.Equal(t, 50, clipped.Dx())

	box = image.Rect(-10, -10, 20, 20)

	clipped, err = ClipBoxToRectangle(&box, 25, 25)
	require.NoError(t, err)

	assert.Equal(t, 0, clipped.Min.X)
	assert.Equal(t, 0, clipped.Min.Y)
	assert.Equal(t, 10, clipped.Max.X)
	assert.Equal(t, 10, clipped.Max.Y)
	assert.Equal(t, 10, clipped.Dx())

	_, err = ClipBoxToRectangle(nil, 20, 20)
	require.Error(t, err)

	_, err = ClipBoxToRectangle(&box, -15, -15)
	require.Error(t, err)
}
