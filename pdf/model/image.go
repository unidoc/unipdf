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
	"math"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/sampling"
)

// Image interface is a basic representation of an image used in PDF.
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

// AlphaMapFunc represents a alpha mapping function: byte -> byte. Can be used for
// thresholding the alpha channel, i.e. setting all alpha values below threshold to transparent.
type AlphaMapFunc func(alpha byte) byte

// AlphaMap performs mapping of alpha data for transformations. Allows custom filtering of alpha data etc.
func (img *Image) AlphaMap(mapFunc AlphaMapFunc) {
	for idx, alpha := range img.alphaData {
		img.alphaData[idx] = mapFunc(alpha)
	}
}

// GetSamples converts the raw byte slice into samples which are stored in a uint32 bit array.
// Each sample is represented by BitsPerComponent consecutive bits in the raw data.
func (img *Image) GetSamples() []uint32 {
	samples := sampling.ResampleBytes(img.Data, int(img.BitsPerComponent))

	expectedLen := int(img.Width) * int(img.Height) * img.ColorComponents
	if len(samples) < expectedLen {
		// Return error, or fill with 0s?
		common.Log.Debug("Error: Too few samples (got %d, expecting %d)", len(samples), expectedLen)
		return samples
	} else if len(samples) > expectedLen {
		samples = samples[:expectedLen]
	}
	return samples
}

// SetSamples convert samples to byte-data and sets for the image.
func (img *Image) SetSamples(samples []uint32) {
	resampled := sampling.ResampleUint32(samples, int(img.BitsPerComponent), 8)
	var data []byte
	for _, val := range resampled {
		data = append(data, byte(val))
	}

	img.Data = data
}

// Resample resamples the image data converting from current BitsPerComponent to a target BitsPerComponent
// value.  Sets the image's BitsPerComponent to the target value following resampling.
//
// For example, converting an 8-bit RGB image to 1-bit grayscale (common for scanned images):
//   // Convert RGB image to grayscale.
//   rgbColorSpace := pdf.NewPdfColorspaceDeviceRGB()
//   grayImage, err := rgbColorSpace.ImageToGray(rgbImage)
//   if err != nil {
//     return err
//   }
//   // Resample as 1 bit.
//   grayImage.Resample(1)
func (img *Image) Resample(targetBitsPerComponent int64) {
	samples := img.GetSamples()

	// Image data are stored row by row.  If the number of bits per row is not a multiple of 8, the end of the
	// row needs to be padded with extra bits to fill out the last byte.
	// Thus the processing is done on a row by row basis below.

	// This one simply resamples the data so that each component has target bits per component...
	// So if the original data was 10011010, then will have 1 0 0 1 1 0 1 0... much longer
	// The key to resampling is that we need to upsample/downsample,
	// i.e. 10011010 >> targetBitsPerComponent
	// Current bits: 8, target bits: 1... need to downsample by 8-1 = 7

	if targetBitsPerComponent < img.BitsPerComponent {
		downsampling := img.BitsPerComponent - targetBitsPerComponent
		for i := range samples {
			samples[i] >>= uint(downsampling)
		}
	} else if targetBitsPerComponent > img.BitsPerComponent {
		upsampling := targetBitsPerComponent - img.BitsPerComponent
		for i := range samples {
			samples[i] <<= uint(upsampling)
		}
	} else {
		return
	}

	// Write out row by row...
	var data []byte
	for i := int64(0); i < img.Height; i++ {
		ind1 := i * img.Width * int64(img.ColorComponents)
		ind2 := (i+1)*img.Width*int64(img.ColorComponents) - 1

		resampled := sampling.ResampleUint32(samples[ind1:ind2], int(targetBitsPerComponent), 8)
		for _, val := range resampled {
			data = append(data, byte(val))
		}
	}

	img.Data = data
	img.BitsPerComponent = int64(targetBitsPerComponent)
}

// ToGoImage converts the unidoc Image to a golang Image structure.
func (img *Image) ToGoImage() (goimage.Image, error) {
	common.Log.Trace("Converting to go image")
	bounds := goimage.Rect(0, 0, int(img.Width), int(img.Height))
	var imgout DrawableImage

	if img.ColorComponents == 1 {
		if img.BitsPerComponent == 16 {
			imgout = goimage.NewGray16(bounds)
		} else {
			imgout = goimage.NewGray(bounds)
		}
	} else if img.ColorComponents == 3 {
		if img.BitsPerComponent == 16 {
			imgout = goimage.NewRGBA64(bounds)
		} else {
			imgout = goimage.NewRGBA(bounds)
		}
	} else if img.ColorComponents == 4 {
		imgout = goimage.NewCMYK(bounds)
	} else {
		// TODO: Force RGB convert?
		common.Log.Debug("Unsupported number of colors components per sample: %d", img.ColorComponents)
		return nil, errors.New("unsupported colors")
	}

	// Draw the data on the image..
	x := 0
	y := 0
	aidx := 0

	samples := img.GetSamples()
	//bytesPerColor := colorComponents * int(img.BitsPerComponent) / 8
	bytesPerColor := img.ColorComponents
	for i := 0; i+bytesPerColor-1 < len(samples); i += bytesPerColor {
		var c gocolor.Color
		if img.ColorComponents == 1 {
			if img.BitsPerComponent == 16 {
				val := uint16(samples[i])<<8 | uint16(samples[i+1])
				c = gocolor.Gray16{val}
			} else {
				// Account for 1-bit/2-bit color images.
				val := samples[i] * 255 / uint32(math.Pow(2, float64(img.BitsPerComponent))-1)
				c = gocolor.Gray{uint8(val & 0xff)}
			}
		} else if img.ColorComponents == 3 {
			if img.BitsPerComponent == 16 {
				r := uint16(samples[i])<<8 | uint16(samples[i+1])
				g := uint16(samples[i+2])<<8 | uint16(samples[i+3])
				b := uint16(samples[i+4])<<8 | uint16(samples[i+5])
				a := uint16(0xffff) // Default: solid (0xffff) whereas transparent=0.
				if img.alphaData != nil && len(img.alphaData) > aidx+1 {
					a = (uint16(img.alphaData[aidx]) << 8) | uint16(img.alphaData[aidx+1])
					aidx += 2
				}
				c = gocolor.RGBA64{R: r, G: g, B: b, A: a}
			} else {
				r := uint8(samples[i] & 0xff)
				g := uint8(samples[i+1] & 0xff)
				b := uint8(samples[i+2] & 0xff)
				a := uint8(0xff) // Default: solid (0xff) whereas transparent=0.
				if img.alphaData != nil && len(img.alphaData) > aidx {
					a = uint8(img.alphaData[aidx])
					aidx++
				}
				c = gocolor.RGBA{R: r, G: g, B: b, A: a}
			}
		} else if img.ColorComponents == 4 {
			c1 := uint8(samples[i] & 0xff)
			m1 := uint8(samples[i+1] & 0xff)
			y1 := uint8(samples[i+2] & 0xff)
			k1 := uint8(samples[i+3] & 0xff)
			c = gocolor.CMYK{C: c1, M: m1, Y: y1, K: k1}
		}

		imgout.Set(x, y, c)
		x++
		if x == int(img.Width) {
			x = 0
			y++
		}
	}

	return imgout, nil
}

// ImageHandler interface implements common image loading and processing tasks.
// Implementing as an interface allows for the possibility to use non-standard libraries for faster
// loading and processing of images.
type ImageHandler interface {
	// Read any image type and load into a new Image object.
	Read(r io.Reader) (*Image, error)

	// NewImageFromGoImage load a unidoc Image from a standard Go image structure.
	NewImageFromGoImage(goimg goimage.Image) (*Image, error)

	// Compress an image.
	Compress(input *Image, quality int64) (*Image, error)
}

// DefaultImageHandler is the default implementation of the ImageHandler using the standard go library.
type DefaultImageHandler struct{}

// NewImageFromGoImage creates a unidoc Image from a golang Image.
func (ih DefaultImageHandler) NewImageFromGoImage(goimg goimage.Image) (*Image, error) {
	// Speed up jpeg encoding by converting to RGBA first.
	// Will not be required once the golang image/jpeg package is optimized.
	b := goimg.Bounds()
	m := goimage.NewRGBA(goimage.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), goimg, b.Min, draw.Src)

	var alphaData []byte
	hasAlpha := false

	var data []byte
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

// Read reads an image and loads into a new Image object with an RGB
// colormap and 8 bits per component.
func (ih DefaultImageHandler) Read(reader io.Reader) (*Image, error) {
	// Load the image with the native implementation.
	goimg, _, err := goimage.Decode(reader)
	if err != nil {
		common.Log.Debug("Error decoding file: %s", err)
		return nil, err
	}

	return ih.NewImageFromGoImage(goimg)
}

// Compress is yet to be implemented.
// Should be able to compress in terms of JPEG quality parameter,
// and DPI threshold (need to know bounding area dimensions).
func (ih DefaultImageHandler) Compress(input *Image, quality int64) (*Image, error) {
	return input, nil
}

// ImageHandling is used for handling images.
var ImageHandling ImageHandler = DefaultImageHandler{}

// SetImageHandler sets the image handler used by the package.
func SetImageHandler(imgHandling ImageHandler) {
	ImageHandling = imgHandling
}
