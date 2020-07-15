/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package imageutil

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRGBA16(t *testing.T) {
	i, err := NewImage(20, 20, 4, 3, nil, nil, nil)
	require.NoError(t, err)

	rgba, ok := i.(*NRGBA16)
	require.True(t, ok)

	x, y := 5, 5
	c, err := rgba.ColorAt(x, y)
	require.NoError(t, err)

	rgbaC, ok := c.(color.NRGBA)
	require.True(t, ok)

	assert.Equal(t, rgbaC, color.NRGBA{A: 0xff})

	// Set sample color.
	srcColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	rgba.Set(x, y, srcColor)

	c, err = rgba.ColorAt(x, y)
	require.NoError(t, err)

	rgbaC, ok = c.(color.NRGBA)
	require.True(t, ok)

	assert.Equal(t, rgbaC, srcColor)

	decode := []float64{1.0, 0, 1.0, 0, 1.0, 0.0}
	rgba.Decode = decode

	rgbaC = rgba.NRGBAAt(x, y)
	// Having 'decode' the values should be opposite to 0xff.
	assert.Equal(t, color.NRGBA{A: 0xff}, rgbaC)

	// Reverse only Red.
	decode = []float64{1.0, 0.0, 0.0, 15.0, 0.0, 15.0}
	rgba.Decode = decode

	rgbaC = rgba.NRGBAAt(x, y)
	// Having decode the values should be opposite to 0xff.
	assert.Equal(t, color.NRGBA{R: 0x00, G: 0xff, B: 0xff, A: 0xff}, rgbaC)

	rgba.MakeAlpha()
	srcColor = color.NRGBA{R: 0x3f, G: 0x3f, B: 0x3f, A: 0x3f}
	rgba.SetNRGBA(x, y, srcColor)

	rgba.Decode = nil
	rgbaC = rgba.NRGBAAt(x, y)

	assert.Equal(t, color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0x33}, rgbaC)
}

func TestRGBA32(t *testing.T) {
	i, err := NewImage(20, 20, 8, 3, nil, nil, nil)
	require.NoError(t, err)

	rgba, ok := i.(*NRGBA32)
	require.True(t, ok)

	x, y := 5, 5
	c, err := rgba.ColorAt(x, y)
	require.NoError(t, err)

	rgbaC, ok := c.(color.NRGBA)
	require.True(t, ok)

	assert.Equal(t, rgbaC, color.NRGBA{A: 0xff})

	// Set sample color.
	srcColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	rgba.Set(x, y, srcColor)

	c, err = rgba.ColorAt(x, y)
	require.NoError(t, err)

	rgbaC, ok = c.(color.NRGBA)
	require.True(t, ok)

	assert.Equal(t, rgbaC, srcColor)

	decode := []float64{1.0, 0, 1.0, 0, 1.0, 0.0}
	rgba.Decode = decode

	rgbaC = rgba.NRGBAAt(x, y)
	// Having 'decode' the values should be opposite to 0xff.
	assert.Equal(t, color.NRGBA{A: 0xff}, rgbaC)

	// Reverse only Red.
	decode = []float64{1.0, 0.0, 0.0, 255.0, 0.0, 255.0}
	rgba.Decode = decode

	rgbaC = rgba.NRGBAAt(x, y)
	// Having decode the values should be opposite to 0xff.
	assert.Equal(t, color.NRGBA{R: 0x00, G: 0xff, B: 0xff, A: 0xff}, rgbaC)

	rgba.MakeAlpha()
	srcColor = color.NRGBA{R: 0x3f, G: 0x3f, B: 0x3f, A: 0x3f}
	rgba.SetNRGBA(x, y, srcColor)

	rgba.Decode = nil
	rgbaC = rgba.NRGBAAt(x, y)

	assert.Equal(t, color.NRGBA{R: 0x3f, G: 0x3f, B: 0x3f, A: 0x3f}, rgbaC)
}

func TestRGBA64(t *testing.T) {
	i, err := NewImage(20, 20, 16, 3, nil, nil, nil)
	require.NoError(t, err)

	rgba, ok := i.(*NRGBA64)
	require.True(t, ok)

	x, y := 5, 5
	c, err := rgba.ColorAt(x, y)
	require.NoError(t, err)

	rgbaC, ok := c.(color.NRGBA64)
	require.True(t, ok)

	assert.Equal(t, rgbaC, color.NRGBA64{A: 0xffff})

	// Set sample color.
	srcColor := color.NRGBA64{R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}
	rgba.Set(x, y, srcColor)

	c, err = rgba.ColorAt(x, y)
	require.NoError(t, err)

	rgbaC, ok = c.(color.NRGBA64)
	require.True(t, ok)

	assert.Equal(t, rgbaC, srcColor)

	decode := []float64{1.0, 0, 1.0, 0, 1.0, 0.0}
	rgba.Decode = decode

	rgbaC = rgba.NRGBA64At(x, y)
	// Having 'decode' the values should be opposite to 0xff.
	assert.Equal(t, color.NRGBA64{A: 0xffff}, rgbaC)

	// Reverse only Red.
	decode = []float64{1.0, 0.0, 0.0, 65535.0, 0.0, 65535.0}
	rgba.Decode = decode

	rgbaC = rgba.NRGBA64At(x, y)
	// Having decode the values should be opposite to 0xff.
	assert.Equal(t, color.NRGBA64{R: 0x00, G: 0xffff, B: 0xffff, A: 0xffff}, rgbaC)

	rgba.MakeAlpha()
	srcColor = color.NRGBA64{R: 0x3f3f, G: 0x3f3f, B: 0x3f3f, A: 0x3f3f}
	rgba.SetNRGBA64(x, y, srcColor)

	rgba.Decode = nil
	rgbaC = rgba.NRGBA64At(x, y)

	assert.Equal(t, color.NRGBA64{R: 0x3f3f, G: 0x3f3f, B: 0x3f3f, A: 0x3f3f}, rgbaC)
}
