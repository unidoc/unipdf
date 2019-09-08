/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageResampling(t *testing.T) {
	img := Image{}

	// Case 1:
	// Data:
	// 4x8bit: 00000001 11101000 01101110 00001010
	// Resample as 1bit:
	//
	// 4x8bit: 00000001 11101000 01101110 00001010
	// Downsample to 1bit
	// 4x8bit: 00000000 00000001 00000000 00000000
	// 4x1bit: 0100
	// Padding with 4x00
	// -> 01000000 = 64 decimal
	//
	img.BitsPerComponent = 8
	img.Data = []byte{1, 232, 110, 10}
	//int(this.Width) * int(this.Height) * this.ColorComponents
	img.Width = 4
	img.ColorComponents = 1
	img.Height = 1
	img.Resample(1)
	if len(img.Data) != 1 {
		t.Errorf("Incorrect length != 1 (%d)", len(img.Data))
		return
	}
	if img.Data[0] != 64 {
		t.Errorf("Value != 4 (%d)", img.Data[0])
	}

	// Case 2:
	// Data:
	// 4x8bit: 00000001 11101000 01101110 00001010 00000001 11101000 01101110 00001010 00000001 11101000 01101110 00001010
	//         0        1        0        0        0        1        0        0        0        1        0        0
	// 010001000100
	// -> 01000100 0100(0000)
	// -> 68 64
	img.BitsPerComponent = 8
	img.Data = []byte{1, 232, 110, 10, 1, 232, 110, 10, 1, 232, 110, 10}
	img.Width = 12
	img.ColorComponents = 1
	img.Height = 1
	img.Resample(1)

	if len(img.Data) != 2 {
		t.Errorf("Incorrect length != 2 (%d)", len(img.Data))
		return
	}
	if img.Data[0] != 68 {
		t.Errorf("Value != 68 (%d)", img.Data[0])
	}
	if img.Data[1] != 64 {
		t.Errorf("Value != 64 (%d)", img.Data[1])
	}
}

func TestImageColorAt(t *testing.T) {
	img := &Image{}
	img.Data = []byte{
		// 01111111 10010000
		127, 144,
		// 01000000 01011101
		64, 93,
		// 10011110 00100101
		158, 37,
	}

	// 1 bit grayscale.
	img.Width = 16
	img.Height = 3
	img.BitsPerComponent = 1
	img.ColorComponents = 1

	c, err := img.ColorAt(3, 0)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 255 {
		t.Errorf("Expected 255. Got %d.", y) // b'1' translated in 0-255 range.
	}

	c, err = img.ColorAt(2, 1)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 0 {
		t.Errorf("Expected 0. Got %d.", y) // b'0' translated in 0-255 range.
	}

	// 2 bit grayscale.
	img.Width = 4
	img.Height = 6
	img.BitsPerComponent = 2

	c, err = img.ColorAt(0, 2)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 85 {
		t.Errorf("Expected 85. Got %d.", y) // b'01' translated in 0-255 range.
	}

	c, err = img.ColorAt(0, 1)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 170 {
		t.Errorf("Expected 170. Got %d.", y) // b'10' translated in 0-255 range.
	}

	// 4 bit grayscale.
	img.Width = 4
	img.Height = 3
	img.BitsPerComponent = 4

	c, err = img.ColorAt(0, 0)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 119 {
		t.Errorf("Expected 119. Got %d.", y) // b'0111' translated in 0-255 range.
	}

	c, err = img.ColorAt(3, 1)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 221 {
		t.Errorf("Expected 221. Got %d.", y) // b'1101' translated in 0-255 range.
	}

	// 8 bit grayscale.
	img.Width = 2
	img.Height = 3
	img.BitsPerComponent = 8

	c, err = img.ColorAt(1, 0)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 144 {
		t.Errorf("Expected 144. Got %d.", y)
	}

	c, err = img.ColorAt(1, 1)
	require.NoError(t, err)
	if y := c.(color.Gray).Y; y != 93 {
		t.Errorf("Expected 93. Got %d.", y)
	}

	// 16 bit grayscale.
	img.Width = 1
	img.Height = 3
	img.BitsPerComponent = 16

	c, err = img.ColorAt(0, 0)
	require.NoError(t, err)
	if y := c.(color.Gray16).Y; y != 32656 {
		t.Errorf("Expected 32656. Got %d.", y) // Bytes 127 and 144.
	}

	c, err = img.ColorAt(0, 1)
	require.NoError(t, err)
	if y := c.(color.Gray16).Y; y != 16477 {
		t.Errorf("Expected 16477. Got %d.", y) // Bytes 64 and 93.
	}

	// 4 bit/component RGB.
	img.Width = 2
	img.Height = 2
	img.BitsPerComponent = 4
	img.ColorComponents = 3

	c, err = img.ColorAt(1, 0)
	require.NoError(t, err)
	r, g, b, _ := c.RGBA()
	if v := r >> 8; v != 0 {
		t.Errorf("Expected 0 for R component. Got %d.", v) // b'0000' translated in 0-255 range.
	}
	if v := g >> 8; v != 68 {
		t.Errorf("Expected 68 for G component. Got %d.", v) // b'0100' translated in 0-255 range.
	}
	if v := b >> 8; v != 0 {
		t.Errorf("Expected 0 for B component. Got %d.", v) // b'0000' translated in 0-255 range.
	}

	c, err = img.ColorAt(1, 1)
	require.NoError(t, err)
	r, g, b, _ = c.RGBA()
	if v := r >> 8; v != 238 {
		t.Errorf("Expected 238 for R component. Got %d.", v) // b'1110' translated in 0-255 range.
	}
	if v := g >> 8; v != 34 {
		t.Errorf("Expected 34 for G component. Got %d.", v) // b'0010' translated in 0-255 range.
	}
	if v := b >> 8; v != 85 {
		t.Errorf("Expected 85 for B component. Got %d.", v) // b'0101' translated in 0-255 range.
	}

	// 8 bit/component RGB.
	img.Width = 2
	img.Height = 1
	img.BitsPerComponent = 8
	img.ColorComponents = 3

	c, err = img.ColorAt(0, 0)
	require.NoError(t, err)
	r, g, b, _ = c.RGBA()
	if v := r >> 8; v != 127 {
		t.Errorf("Expected 127 for R component. Got %d.", v)
	}
	if v := g >> 8; v != 144 {
		t.Errorf("Expected 144 for G component. Got %d.", v)
	}
	if v := b >> 8; v != 64 {
		t.Errorf("Expected 64 for B component. Got %d.", v)
	}

	c, err = img.ColorAt(1, 0)
	require.NoError(t, err)
	r, g, b, _ = c.RGBA()
	if v := r >> 8; v != 93 {
		t.Errorf("Expected 238 for R component. Got %d.", v)
	}
	if v := g >> 8; v != 158 {
		t.Errorf("Expected 34 for G component. Got %d.", v)
	}
	if v := b >> 8; v != 37 {
		t.Errorf("Expected 85 for B component. Got %d.", v)
	}

	// 16 bit/component RGB.
	img.Width = 1
	img.Height = 2
	img.BitsPerComponent = 16
	img.ColorComponents = 3
	img.Data = []byte{
		// 01111111 10010000
		127, 144,
		// 01000000 01011101
		64, 93,
		// 10011110 00100101
		158, 37,
		// 00101100 01001110
		44, 78,
		// 00001101 00110011
		13, 51,
		// 10100111 10111101
		167, 189,
	}

	c, err = img.ColorAt(0, 0)
	require.NoError(t, err)
	r, g, b, _ = c.RGBA()
	if v := r; v != 32656 {
		t.Errorf("Expected 32656 for R component. Got %d.", v) // Bytes 127 and 144.
	}
	if v := g; v != 16477 {
		t.Errorf("Expected 16477 for G component. Got %d.", v) // Bytes 63 and 93.
	}
	if v := b; v != 40485 {
		t.Errorf("Expected 40485 for B component. Got %d.", v) // Bytes 158 and 37.
	}

	c, err = img.ColorAt(0, 1)
	require.NoError(t, err)
	r, g, b, _ = c.RGBA()
	if v := r; v != 11342 {
		t.Errorf("Expected 11342 for R component. Got %d.", v) // Bytes 44 and 78.
	}
	if v := g; v != 3379 {
		t.Errorf("Expected 3379 for B component. Got %d.", v) // Bytes 13 and 51.
	}
	if v := b; v != 42941 {
		t.Errorf("Expected 42941 for G component. Got %d.", v) // Bytes 167 and 189.
	}

	// 8 bit/component CMYK.
	img.Width = 2
	img.Height = 2
	img.BitsPerComponent = 8
	img.ColorComponents = 4

	c, err = img.ColorAt(0, 0)
	require.NoError(t, err)
	cc := c.(color.CMYK)
	if cc.C != 127 || cc.M != 144 || cc.Y != 64 || cc.K != 93 {
		t.Errorf("Expected CMYK values (127,144,64,93). Got (%d,%d,%d,%d).", cc.C, cc.M, cc.Y, cc.K)
	}

	c, err = img.ColorAt(1, 0)
	require.NoError(t, err)
	cc = c.(color.CMYK)
	if cc.C != 158 || cc.M != 37 || cc.Y != 44 || cc.K != 78 {
		t.Errorf("Expected CMYK values (158,37,44,78). Got (%d,%d,%d,%d).", cc.C, cc.M, cc.Y, cc.K)
	}

	c, err = img.ColorAt(0, 1)
	require.NoError(t, err)
	cc = c.(color.CMYK)
	if cc.C != 13 || cc.M != 51 || cc.Y != 167 || cc.K != 189 {
		t.Errorf("Expected CMYK values (13,51,167,189). Got (%d,%d,%d,%d).", cc.C, cc.M, cc.Y, cc.K)
	}
}

func makeTestImage(x, y int, val byte) *image.RGBA {
	rect := image.Rect(0, 0, x, y)
	m := image.NewRGBA(rect)

	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			rgba := color.RGBA{R: val, G: val, B: val, A: 0}
			m.Set(i, j, rgba)
		}
	}
	return m
}

func BenchmarkImageCopyingDraw(b *testing.B) {
	for n := 0; n < b.N; n++ {
		img := makeTestImage(100, 100, byte(n))
		b := img.Bounds()
		m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

		// Image copying using draw.Draw.
		draw.Draw(m, img.Bounds(), img, b.Min, draw.Src)
	}
}

func BenchmarkImageCopyingAtSet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		img := makeTestImage(100, 100, byte(n))
		b := img.Bounds()
		m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

		// Image copying with At and direct set.
		for x := m.Rect.Min.X; x < m.Rect.Max.X; x++ {
			for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
				m.Set(x, y, img.At(x, y))
			}
		}
	}
}

func BenchmarkImageCopyingAtDirectSet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		img := makeTestImage(100, 100, byte(n))
		b := img.Bounds()
		m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

		// Image copying with At to get colors and set Pix directly.
		i := 0
		for x := m.Rect.Min.X; x < m.Rect.Max.X; x++ {
			for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
				r, g, b, a := img.At(x, y).RGBA()

				m.Pix[4*i], m.Pix[4*i+1], m.Pix[4*i+2], m.Pix[4*i+3] = byte(r), byte(g), byte(b), byte(a)
				i++
			}
		}
	}
}

func BenchmarkImageCopyingDirect(b *testing.B) {
	for n := 0; n < b.N; n++ {
		img := makeTestImage(100, 100, byte(n))
		b := img.Bounds()
		m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

		// Image copying with direct Pix access for getting and setting.
		i := 0
		for x := m.Rect.Min.X; x < m.Rect.Max.X; x++ {
			for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
				m.Pix[4*i], m.Pix[4*i+1], m.Pix[4*i+2], m.Pix[4*i+3] = img.Pix[4*i], img.Pix[4*i+1], img.Pix[4*i+2], img.Pix[4*i+3]
				i++
			}
		}
	}
}
