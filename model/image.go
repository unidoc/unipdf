/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"
	goimage "image"
	gocolor "image/color"
	"image/draw"
	// Imported for initialization side effects.
	_ "image/gif"
	_ "image/png"
	"io"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/imageutil"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/sampling"
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

// ConvertToBinary converts current image into binary (bi-level) format.
// Binary images are composed of single bits per pixel (only black or white).
// If provided image has more color components, then it would be converted into binary image using
// histogram auto threshold function.
func (img *Image) ConvertToBinary() error {
	// check if  given image is already a binary image (1 bit per component - 1 color component - the size of the data
	// is equal to the multiplication of width and height.
	if img.ColorComponents == 1 && img.BitsPerComponent == 1 {
		return nil
	}
	i, err := img.ToGoImage()
	if err != nil {
		return err
	}
	gray := imageutil.ImgToGray(i)
	// check if 'img' is already a binary image.
	if !imageutil.IsGrayImgBlackAndWhite(gray) {
		threshold := imageutil.AutoThresholdTriangle(imageutil.GrayImageHistogram(gray))
		gray = imageutil.ImgToBinary(i, threshold)
	}
	// use JBIG2 bitmap as the temporary binary data converter - by default it uses
	tmpBM := bitmap.New(int(img.Width), int(img.Height))
	for y := 0; y < tmpBM.Height; y++ {
		for x := 0; x < tmpBM.Width; x++ {
			c := gray.GrayAt(x, y)
			// set only the white pixel - c.Y != 0
			if c.Y != 0 {
				if err = tmpBM.SetPixel(x, y, 1); err != nil {
					return err
				}
			}
		}
	}
	unpaddedData, err := tmpBM.GetUnpaddedData()
	if err != nil {
		return err
	}
	img.BitsPerComponent = 1
	img.ColorComponents = 1
	img.Data = unpaddedData
	return nil
}

// GetParamsDict returns *core.PdfObjectDictionary with a set of basic image parameters.
func (img *Image) GetParamsDict() *core.PdfObjectDictionary {
	params := core.MakeDict()
	params.Set("Width", core.MakeInteger(img.Width))
	params.Set("Height", core.MakeInteger(img.Height))
	params.Set("ColorComponents", core.MakeInteger(int64(img.ColorComponents)))
	params.Set("BitsPerComponent", core.MakeInteger(img.BitsPerComponent))
	return params
}

// GetSamples converts the raw byte slice into samples which are stored in a uint32 bit array.
// Each sample is represented by BitsPerComponent consecutive bits in the raw data.
// NOTE: The method resamples the image byte data before returning the result and
// this could lead to high memory usage, especially on large images. It should
// be avoided, when possible. It is recommended to access the Data field of the
// image directly or use the ColorAt method to extract individual pixels.
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
// NOTE: The method resamples the data and this could lead to high memory usage,
// especially on large images. It should be used only when it is not possible
// to work with the image byte data directly.
func (img *Image) SetSamples(samples []uint32) {
	resampled := sampling.ResampleUint32(samples, int(img.BitsPerComponent), 8)
	data := make([]byte, len(resampled))
	for i, val := range resampled {
		data[i] = byte(val)
	}

	img.Data = data
}

// ColorAt returns the color of the image pixel specified by the x and y coordinates.
func (img *Image) ColorAt(x, y int) (gocolor.Color, error) {
	data := img.Data
	lenData := len(img.Data)
	maxVal := uint32(1<<uint32(img.BitsPerComponent)) - 1

	switch img.ColorComponents {
	case 1:
		// Grayscale image.
		switch img.BitsPerComponent {
		case 1, 2, 4:
			// 1, 2 or 4 bit grayscale image.
			bpc := int(img.BitsPerComponent)
			divider := 8 / bpc

			// Calculate index of byte containing the gray value
			// in the image data, based on the specified x,y coordinates.
			idx := (y*int(img.Width) + x) / divider
			if idx >= lenData {
				return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
			}

			// Calculate bit position at which the color data starts.
			pos := 8 - uint(((y*int(img.Width)+x)%divider)*bpc+bpc)

			// Extract gray color value starting at the calculated position.
			val := float64(((1 << uint(img.BitsPerComponent)) - 1) & (data[idx] >> pos))
			if len(img.decode) == 2 {
				dMin, dMax := img.decode[0], img.decode[1]
				val = interpolate(float64(val), 0, float64(maxVal), dMin, dMax)
			}

			return gocolor.Gray{
				Y: uint8(uint32(val) * 255 / maxVal & 0xff),
			}, nil
		case 16:
			// 16 bit grayscale image.
			idx := (y*int(img.Width) + x) * 2
			if idx+1 >= lenData {
				return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
			}

			return gocolor.Gray16{
				Y: uint16(data[idx])<<8 | uint16(data[idx+1]),
			}, nil
		default:
			// Assuming 8 bit grayscale image.
			idx := y*int(img.Width) + x
			if idx >= lenData {
				return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
			}
			val := float64(data[idx])
			if len(img.decode) == 2 {
				dMin, dMax := img.decode[0], img.decode[1]
				val = interpolate(float64(val), 0, float64(maxVal), dMin, dMax)
			}

			return gocolor.Gray{
				Y: uint8(uint32(val) * 255 / maxVal & 0xff),
			}, nil
		}
	case 3:
		// RGB image.
		switch img.BitsPerComponent {
		case 4:
			// 4 bit per component RGB image.
			idx := (y*int(img.Width) + x) * 3 / 2
			if idx+1 >= lenData {
				return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
			}

			// Calculate bit position at which the color data starts.
			pos := (y*int(img.Width) + x) * 3 % 2

			var r, g, b uint8
			if pos == 0 {
				// The R and G components are contained by the current byte
				// and the B component is contained by the next byte.
				r = ((1 << uint(img.BitsPerComponent)) - 1) & (data[idx] >> uint(4))
				g = ((1 << uint(img.BitsPerComponent)) - 1) & (data[idx] >> uint(0))
				b = ((1 << uint(img.BitsPerComponent)) - 1) & (data[idx+1] >> uint(4))
			} else {
				// The R component is contained by the current byte and the
				// G and B components are contained by the next byte.
				r = ((1 << uint(img.BitsPerComponent)) - 1) & (data[idx] >> uint(0))
				g = ((1 << uint(img.BitsPerComponent)) - 1) & (data[idx+1] >> uint(4))
				b = ((1 << uint(img.BitsPerComponent)) - 1) & (data[idx+1] >> uint(0))
			}

			return gocolor.RGBA{
				R: uint8(uint32(r) * 255 / maxVal & 0xff),
				G: uint8(uint32(g) * 255 / maxVal & 0xff),
				B: uint8(uint32(b) * 255 / maxVal & 0xff),
				A: uint8(0xff),
			}, nil
		case 16:
			// 16 bit per component RGB image.
			idx := (y*int(img.Width) + x) * 2

			i := idx * 3
			if i+5 >= lenData {
				return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
			}

			a := uint16(0xffff)
			if img.alphaData != nil && len(img.alphaData) > idx+1 {
				a = uint16(img.alphaData[idx])<<8 | uint16(img.alphaData[idx+1])
			}

			return gocolor.RGBA64{
				R: uint16(data[i])<<8 | uint16(data[i+1]),
				G: uint16(data[i+2])<<8 | uint16(data[i+3]),
				B: uint16(data[i+4])<<8 | uint16(data[i+5]),
				A: a,
			}, nil
		default:
			// Assuming 8 bit per component RGB image.
			idx := y*int(img.Width) + x

			i := 3 * idx
			if i+2 >= lenData {
				return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
			}

			a := uint8(0xff)
			if img.alphaData != nil && len(img.alphaData) > idx {
				a = uint8(img.alphaData[idx])
			}

			return gocolor.RGBA{
				R: uint8(data[i] & 0xff),
				G: uint8(data[i+1] & 0xff),
				B: uint8(data[i+2] & 0xff),
				A: a,
			}, nil
		}
	case 4:
		// CMYK image.
		idx := 4 * (y*int(img.Width) + x)
		if idx+3 >= lenData {
			return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
		}

		return gocolor.CMYK{
			C: uint8(data[idx] & 0xff),
			M: uint8(data[idx+1] & 0xff),
			Y: uint8(data[idx+2] & 0xff),
			K: uint8(data[idx+3] & 0xff),
		}, nil
	}

	common.Log.Debug("ERROR: unsupported image. %d components, %d bits per component", img.ColorComponents, img.BitsPerComponent)
	return nil, errors.New("unsupported image colorspace")
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

// ToJBIG2Image converts current image to the core.JBIG2Image.
func (img *Image) ToJBIG2Image() (*core.JBIG2Image, error) {
	goImg, err := img.ToGoImage()
	if err != nil {
		return nil, err
	}
	return core.GoImageToJBIG2(goImg, core.JB2ImageAutoThreshold)
}

// ToGoImage converts the unidoc Image to a golang Image structure.
func (img *Image) ToGoImage() (goimage.Image, error) {
	common.Log.Trace("Converting to go image")
	bounds := goimage.Rect(0, 0, int(img.Width), int(img.Height))

	var imgout core.DrawableImage
	switch img.ColorComponents {
	case 1:
		if img.BitsPerComponent == 16 {
			imgout = goimage.NewGray16(bounds)
		} else {
			imgout = goimage.NewGray(bounds)
		}
	case 3:
		if img.BitsPerComponent == 16 {
			imgout = goimage.NewRGBA64(bounds)
		} else {
			imgout = goimage.NewRGBA(bounds)
		}
	case 4:
		imgout = goimage.NewCMYK(bounds)
	default:
		// TODO: Force RGB convert?
		common.Log.Debug("Unsupported number of colors components per sample: %d", img.ColorComponents)
		return nil, errors.New("unsupported colors")
	}

	for y := 0; y < int(img.Height); y++ {
		for x := 0; x < int(img.Width); x++ {
			color, err := img.ColorAt(x, y)
			if err != nil {
				common.Log.Debug("ERROR: %v. Image details: %d components, %d bits per component, %dx%d dimensions, %d data length",
					err, img.ColorComponents, img.BitsPerComponent, img.Width, img.Height, len(img.Data))
				continue
			}

			imgout.Set(x, y, color)
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

	// NewImageFromGoImage loads a RGBA unidoc Image from a standard Go image structure.
	NewImageFromGoImage(goimg goimage.Image) (*Image, error)

	// NewGrayImageFromGoImage loads a grayscale unidoc Image from a standard Go image structure.
	NewGrayImageFromGoImage(goimg goimage.Image) (*Image, error)

	// Compress an image.
	Compress(input *Image, quality int64) (*Image, error)
}

// DefaultImageHandler is the default implementation of the ImageHandler using the standard go library.
type DefaultImageHandler struct{}

// NewImageFromGoImage creates a new RGBA unidoc Image from a golang Image.
// If `goimg` is grayscale (*goimage.Gray) then calls NewGrayImageFromGoImage instead.
func (ih DefaultImageHandler) NewImageFromGoImage(goimg goimage.Image) (*Image, error) {
	b := goimg.Bounds()

	var m *goimage.NRGBA
	switch t := goimg.(type) {
	case *goimage.Gray, *goimage.Gray16:
		return ih.NewGrayImageFromGoImage(goimg)
	case *goimage.NRGBA:
		m = t
	default:
		// Speed up jpeg encoding by converting to NRGBA first.
		// Will not be required once the golang image/jpeg package is optimized.
		m = goimage.NewNRGBA(goimage.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(m, m.Bounds(), goimg, b.Min, draw.Src)
		b = m.Bounds()
	}

	numPixels := b.Dx() * b.Dy()
	data := make([]byte, 3*numPixels)
	alphaData := make([]byte, numPixels)
	hasAlpha := false

	i0 := m.PixOffset(b.Min.X, b.Min.Y)
	i1 := i0 + b.Dx()*4

	j := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for i := i0; i < i1; i += 4 {
			data[3*j], data[3*j+1], data[3*j+2] = m.Pix[i], m.Pix[i+1], m.Pix[i+2]
			alpha := m.Pix[i+3]
			if alpha != 255 {
				// If all alpha values are 255 (opaque), means that the alpha transparency channel is unnecessary.
				hasAlpha = true
			}
			alphaData[j] = alpha
			j++
		}

		i0 += m.Stride
		i1 += m.Stride
	}

	imag := Image{}
	imag.Width = int64(b.Dx())
	imag.Height = int64(b.Dy())
	imag.BitsPerComponent = 8 // RGBA colormap
	imag.ColorComponents = 3
	imag.Data = data

	imag.hasAlpha = hasAlpha
	if hasAlpha {
		imag.alphaData = alphaData
	}

	return &imag, nil
}

// NewGrayImageFromGoImage creates a new grayscale unidoc Image from a golang Image.
func (ih DefaultImageHandler) NewGrayImageFromGoImage(goimg goimage.Image) (*Image, error) {
	b := goimg.Bounds()

	var m *goimage.Gray
	switch t := goimg.(type) {
	case *goimage.Gray:
		m = t
		if len(m.Pix) != b.Dx()*b.Dy() {
			// Detects when the image Pix data is not of correct format, typically happens
			// when m.Stride does not match the image width (extra bytes at end of each line for example).
			// Rearrange the data back such that the Pix data is arranged consistently.
			// Disadvantage of this is that it doubles the memory use as the data is
			// copied when creating the new structure.
			m = goimage.NewGray(b)
			draw.Draw(m, b, goimg, b.Min, draw.Src)
		}
	default:
		m = goimage.NewGray(b)
		draw.Draw(m, b, goimg, b.Min, draw.Src)
	}

	return &Image{
		Width:            int64(b.Dx()),
		Height:           int64(b.Dy()),
		BitsPerComponent: 8,
		ColorComponents:  1,
		Data:             m.Pix,
	}, nil
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
