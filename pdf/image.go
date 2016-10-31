/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"

	"github.com/unidoc/unidoc/common"
)

// Basic representation of an image.  The images are stored in the JPEG
// format.
type Image struct {
	Width  int64
	Height int64
	Data   *bytes.Buffer
}

type ImageHandler interface {
	// Read any image type and load into a new Image object.
	Read(r io.Reader) (*Image, error)
	// Compress an image.
	Compress(input *Image, quality int64) (*Image, error)
}

// Default implementation.

type DefaultImageHandler struct{}

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
	imag.Data = &buf

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
