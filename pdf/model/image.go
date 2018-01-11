/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	goimage "image"
	gocolor "image/color"
	"image/draw"
	_ "image/gif"
	_ "image/png"
	"io"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/sampling"
)

// Basic representation of an image.
// The colorspace is not specified, but must be known when handling the image.
type Image struct {
	Width            int64  // The width of the image in samples
	Height           int64  // The height of the image in samples
	BitsPerComponent int64  // The number of bits per color component
	ColorComponents  int    // Color components per pixel
	Data             []byte // Image data stored as bytes.

	// Transparency data: alpha channel.
	// Stored in same bits per component as original data with 1 color component.
	alphaData []byte // Alpha channel data.
	hasAlpha  bool   // Indicates whether the alpha channel data is available.

	decode []float64 // [Dmin Dmax ... values for each color component]
}

// Threshold alpha channel.  Set all alpha values below threshold to transparent.
type AlphaMapFunc func(alpha byte) byte

// Allow mapping of alpha data for transformations.  Can allow custom filtering of alpha data etc.
func (this Image) AlphaMap(mapFunc AlphaMapFunc) {
	for idx, alpha := range this.alphaData {
		this.alphaData[idx] = mapFunc(alpha)
	}
}

// Convert the raw byte slice into samples which are stored in a uint32 bit array.
// Each sample is represented by BitsPerComponent consecutive bits in the raw data.
func (this *Image) GetSamples() []uint32 {
	samples := sampling.ResampleBytes(this.Data, int(this.BitsPerComponent))

	expectedLen := int(this.Width) * int(this.Height) * this.ColorComponents
	if len(samples) < expectedLen {
		// Return error, or fill with 0s?
		common.Log.Debug("Error: Too few samples (got %d, expecting %d)", len(samples), expectedLen)
		return samples
	} else if len(samples) > expectedLen {
		samples = samples[:expectedLen]
	}
	return samples
}

// Convert samples to byte-data.
func (this *Image) SetSamples(samples []uint32) {
	resampled := sampling.ResampleUint32(samples, int(this.BitsPerComponent), 8)
	data := []byte{}
	for _, val := range resampled {
		data = append(data, byte(val))
	}

	this.Data = data
}

// Converts the unidoc Image to a golang Image structure.
func (this *Image) ToGoImage() (goimage.Image, error) {
	common.Log.Trace("Converting to go image")
	bounds := goimage.Rect(0, 0, int(this.Width), int(this.Height))
	var img DrawableImage

	if this.ColorComponents == 1 {
		if this.BitsPerComponent == 16 {
			img = goimage.NewGray16(bounds)
		} else {
			img = goimage.NewGray(bounds)
		}
	} else if this.ColorComponents == 3 {
		if this.BitsPerComponent == 16 {
			img = goimage.NewRGBA64(bounds)
		} else {
			img = goimage.NewRGBA(bounds)
		}
	} else if this.ColorComponents == 4 {
		img = goimage.NewCMYK(bounds)
	} else {
		// XXX? Force RGB convert?
		common.Log.Debug("Unsupported number of colors components per sample: %d", this.ColorComponents)
		return nil, errors.New("Unsupported colors")
	}

	// Draw the data on the image..
	x := 0
	y := 0
	aidx := 0

	samples := this.GetSamples()
	//bytesPerColor := colorComponents * int(this.BitsPerComponent) / 8
	bytesPerColor := this.ColorComponents
	for i := 0; i+bytesPerColor-1 < len(samples); i += bytesPerColor {
		var c gocolor.Color
		if this.ColorComponents == 1 {
			if this.BitsPerComponent == 16 {
				val := uint16(samples[i])<<8 | uint16(samples[i+1])
				c = gocolor.Gray16{val}
			} else {
				val := uint8(samples[i] & 0xff)
				c = gocolor.Gray{val}
			}
		} else if this.ColorComponents == 3 {
			if this.BitsPerComponent == 16 {
				r := uint16(samples[i])<<8 | uint16(samples[i+1])
				g := uint16(samples[i+2])<<8 | uint16(samples[i+3])
				b := uint16(samples[i+4])<<8 | uint16(samples[i+5])
				a := uint16(0)
				if this.alphaData != nil && len(this.alphaData) > aidx+1 {
					a = (uint16(this.alphaData[aidx]) << 8) | uint16(this.alphaData[aidx+1])
					aidx += 2
				}
				c = gocolor.RGBA64{R: r, G: g, B: b, A: a}
			} else {
				r := uint8(samples[i] & 0xff)
				g := uint8(samples[i+1] & 0xff)
				b := uint8(samples[i+2] & 0xff)
				a := uint8(0)
				if this.alphaData != nil && len(this.alphaData) > aidx {
					a = uint8(this.alphaData[aidx])
					aidx++
				}
				c = gocolor.RGBA{R: r, G: g, B: b, A: a}
			}
		} else if this.ColorComponents == 4 {
			c1 := uint8(samples[i] & 0xff)
			m1 := uint8(samples[i+1] & 0xff)
			y1 := uint8(samples[i+2] & 0xff)
			k1 := uint8(samples[i+3] & 0xff)
			c = gocolor.CMYK{C: c1, M: m1, Y: y1, K: k1}
		}

		img.Set(x, y, c)
		x++
		if x == int(this.Width) {
			x = 0
			y++
		}
	}

	return img, nil
}

// The ImageHandler interface implements common image loading and processing tasks.
// Implementing as an interface allows for the possibility to use non-standard libraries for faster
// loading and processing of images.
type ImageHandler interface {
	// Read any image type and load into a new Image object.
	Read(r io.Reader) (*Image, error)

	// Load a unidoc Image from a standard Go image structure.
	NewImageFromGoImage(goimg goimage.Image) (*Image, error)

	// Compress an image.
	Compress(input *Image, quality int64) (*Image, error)
}

// Default implementation.

type DefaultImageHandler struct{}

// Create a unidoc Image from a golang Image.
func (this DefaultImageHandler) NewImageFromGoImage(goimg goimage.Image) (*Image, error) {
	// Speed up jpeg encoding by converting to RGBA first.
	// Will not be required once the golang image/jpeg package is optimized.
	b := goimg.Bounds()
	m := goimage.NewRGBA(goimage.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), goimg, b.Min, draw.Src)

	alphaData := []byte{}
	hasAlpha := false

	data := []byte{}
	for i := 0; i < len(m.Pix); i += 4 {
		data = append(data, m.Pix[i], m.Pix[i+1], m.Pix[i+2])

		alpha := m.Pix[i+3]
		if alpha != 255 {
			// If all alpha values are 255 (opaque), means that the alpha transparency channel is unnecessary.
			hasAlpha = true
		}
		alphaData = append(alphaData, alpha)
	}

	imag := Image{}
	imag.Width = int64(b.Dx())
	imag.Height = int64(b.Dy())
	imag.BitsPerComponent = 8 // RGBA colormap
	imag.ColorComponents = 3
	imag.Data = data // buf.Bytes()

	imag.hasAlpha = hasAlpha
	if hasAlpha {
		imag.alphaData = alphaData
	}

	return &imag, nil
}

// Reads an image and loads into a new Image object with an RGB
// colormap and 8 bits per component.
func (this DefaultImageHandler) Read(reader io.Reader) (*Image, error) {
	// Load the image with the native implementation.
	goimg, _, err := goimage.Decode(reader)
	if err != nil {
		common.Log.Debug("Error decoding file: %s", err)
		return nil, err
	}

	return this.NewImageFromGoImage(goimg)
}

// To be implemented.
// Should be able to compress in terms of JPEG quality parameter,
// and DPI threshold (need to know bounding area dimensions).
func (this DefaultImageHandler) Compress(input *Image, quality int64) (*Image, error) {
	return input, nil
}

var ImageHandling ImageHandler = DefaultImageHandler{}

func SetImageHandler(imgHandling ImageHandler) {
	ImageHandling = imgHandling
}
