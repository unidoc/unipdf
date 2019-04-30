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
