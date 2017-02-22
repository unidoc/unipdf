/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/sampling"
)

// Basic representation of an image.
// The colorspace is not specified, but must be known when handling the image.
type Image struct {
	Width            int64  // The width of the image in samples
	Height           int64  // The height of the image in samples
	BitsPerComponent int64  // The number of bits per color component
	Data             []byte // Image data stored as bytes.
}

// Convert the raw byte slice into samples which are stored in a uint32 bit array.
// Each sample is represented by BitsPerComponent consecutive bits in the raw data.
func (this *Image) GetSamples() []uint32 {
	samples := sampling.ResampleBytes(this.Data, int(this.BitsPerComponent))
	return samples
}

// Convert samples to byte-data.
func (this *Image) SetSamples(samples []uint32) {
	resampled := sampling.ResampleUint32(samples, 8)
	data := []byte{}
	for _, val := range resampled {
		data = append(data, byte(val))
	}
	this.Data = data
}

type ImageHandler interface {
	// Read any image type and load into a new Image object.
	Read(r io.Reader) (*Image, error)
	// Compress an image.
	Compress(input *Image, quality int64) (*Image, error)
}

// Default implementation.

type DefaultImageHandler struct{}

// Reads an image and loads into a new Image object with an RGB
// colormap and 8 bits per component.
func (this DefaultImageHandler) Read(reader io.Reader) (*Image, error) {
	// Load the image with the native implementation.
	img, _, err := image.Decode(reader)
	if err != nil {
		common.Log.Debug("Error decoding file: %s", err)
		return nil, err
	}

	// Write image stream.
	var buf bytes.Buffer
	opt := jpeg.Options{}
	// Use full quality.
	opt.Quality = 100

	// Speed up jpeg encoding by converting to RGBA first.
	// Will not be required once the golang image/jpeg package is optimized.
	b := img.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), img, b.Min, draw.Src)
	err = jpeg.Encode(&buf, m, &opt)
	if err != nil {
		return nil, err
	}

	imag := Image{}
	imag.Width = int64(b.Dx())
	imag.Height = int64(b.Dy())
	imag.BitsPerComponent = 8 // RGBA colormap
	imag.Data = buf.Bytes()

	return &imag, nil
}

// To be implemented.
func (this DefaultImageHandler) Compress(input *Image, quality int64) (*Image, error) {
	return input, nil
}

var ImageHandling ImageHandler = DefaultImageHandler{}

func SetImageHandler(imgHandling ImageHandler) {
	ImageHandling = imgHandling
}
