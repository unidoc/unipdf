/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package imageutil

import (
	"errors"
	"fmt"
	"image"
	"image/color"
)

// CMYK is the interface used for getting and operation on the CMYK interface.
type CMYK interface {
	CMYKAt(x, y int) color.CMYK
	SetCMYK(x, y int, c color.CMYK)
}

// CMYK32 is an Image implementation for the CMYK32 color colorConverter.
type CMYK32 struct {
	ImageBase
}

//
// CMYK32 Image interface methods.
//

// Compile time check if CMYK32 implements Image interface.
var _ Image = &CMYK32{}

// Base implements Image interface.
func (i *CMYK32) Base() *ImageBase {
	return &i.ImageBase
}

// ColorAt implements Image interface.
func (i *CMYK32) ColorAt(x, y int) (color.Color, error) {
	return ColorAtCMYK(x, y, i.Width, i.Data, i.Decode)
}

// Copy implements Image interface.
func (i *CMYK32) Copy() Image {
	return &CMYK32{ImageBase: i.copy()}
}

// Set implements draw.Image interface.
func (i *CMYK32) Set(x, y int, c color.Color) {
	idx := 4 * (y*i.Width + x)
	if idx+3 >= len(i.Data) {
		return
	}
	cmyk := color.CMYKModel.Convert(c).(color.CMYK)
	i.Data[idx] = cmyk.C
	i.Data[idx+1] = cmyk.M
	i.Data[idx+2] = cmyk.Y
	i.Data[idx+3] = cmyk.K
}

// Validate implements Image interface.
func (i *CMYK32) Validate() error {
	if len(i.Data) != 4*i.Width*i.Height {
		return errors.New("invalid image data size for provided dimensions")
	}
	return nil
}

//
// CMYK32 image.Image interface methods.
//

// ColorModel implements image.Image interface.
func (i *CMYK32) ColorModel() color.Model {
	return color.CMYKModel
}

// Bounds implements image.Image interface.
func (i *CMYK32) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

// At implements image.Image interface.
func (i *CMYK32) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

//
// CMYK32 CMYK interface methods.
//

// CMYKAt implements CMYK interface.
func (i *CMYK32) CMYKAt(x, y int) color.CMYK {
	c, _ := ColorAtCMYK(x, y, i.Width, i.Data, i.Decode)
	return c
}

// SetCMYK implements CMYK interface.
func (i *CMYK32) SetCMYK(x, y int, c color.CMYK) {
	idx := 4 * (y*i.Width + x)
	if idx+3 >= len(i.Data) {
		return
	}
	i.Data[idx] = c.C
	i.Data[idx+1] = c.M
	i.Data[idx+2] = c.Y
	i.Data[idx+3] = c.K
}

// ColorAtCMYK gets the color at CMYK32 colorspace at 'x' and 'y' position for an image with specific 'width'.
func ColorAtCMYK(x, y, width int, data []byte, decode []float64) (color.CMYK, error) {
	idx := 4 * (y*width + x)
	if idx+3 >= len(data) {
		return color.CMYK{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	C := data[idx] & 0xff
	M := data[idx+1] & 0xff
	Y := data[idx+2] & 0xff
	K := data[idx+3] & 0xff

	if len(decode) == 8 {
		C = uint8(uint32(LinearInterpolate(float64(C), 0, 255, decode[0], decode[1])) & 0xff)
		M = uint8(uint32(LinearInterpolate(float64(M), 0, 255, decode[2], decode[3])) & 0xff)
		Y = uint8(uint32(LinearInterpolate(float64(Y), 0, 255, decode[4], decode[5])) & 0xff)
		K = uint8(uint32(LinearInterpolate(float64(K), 0, 255, decode[6], decode[7])) & 0xff)
	}
	return color.CMYK{C: C, M: M, Y: Y, K: K}, nil

}

func cmykConverter(src image.Image) (Image, error) {
	if i, ok := src.(*CMYK32); ok {
		return i.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 8, 4, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	switch srcImg := src.(type) {
	case CMYK:
		// goimage.CMYK implementation.
		cmykConvertCMYKImage(srcImg, g.(CMYK), bounds)
	case Gray:
		cmykConvertGrayImage(srcImg, g.(CMYK), bounds)
	case NRGBA:
		cmykConvertNRGBAImage(srcImg, g.(CMYK), bounds)
	case RGBA:
		cmykConvertRGBAImage(srcImg, g.(CMYK), bounds)
	default:
		convertImages(src, g, bounds)
	}
	return g, nil
}

func cmykConvertCMYKImage(src, dst CMYK, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			dst.SetCMYK(x, y, src.CMYKAt(x, y))
		}
	}
}

func cmykConvertNRGBAImage(src NRGBA, dst CMYK, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := src.NRGBAAt(x, y)
			dst.SetCMYK(x, y, colorsNRGBAToCMYK(c))
		}
	}
}

func cmykConvertGrayImage(src Gray, dst CMYK, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			g := src.GrayAt(x, y)
			dst.SetCMYK(x, y, colorsGrayToCMYK(g))
		}
	}
}

func cmykConvertRGBAImage(src RGBA, dst CMYK, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := src.RGBAAt(x, y)
			dst.SetCMYK(x, y, colorsRGBAToCMYK(c))
		}
	}
}
