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
	"math"

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
	BytesPerLine     int    // The number of bytes per line.

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
	img.BitsPerComponent = 1
	img.ColorComponents = 1
	img.setBytesPerLine()
	img.Data = tmpBM.Data
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

	if img.BitsPerComponent < 8 {
		// If the bits per component number is smaller than 8 - trim the padding at the end of each row.
		// We want only color values - not the padding.
		samples = img.samplesTrimPadding(samples)
	}
	expectedLen := int(img.Width) * int(img.Height) * img.ColorComponents
	if len(samples) < expectedLen {
		// Return error, or fill with 0s?
		common.Log.Debug("Error: Too few samples (got %d, expecting %d)", len(samples), expectedLen)
		return samples
	} else if len(samples) > expectedLen {
		common.Log.Debug("Error: Too many samples (got %d, expecting %d", len(samples), expectedLen)
		samples = samples[:expectedLen]
	}
	return samples
}

// SetSamples convert samples to byte-data and sets for the image.
// NOTE: The method resamples the data and this could lead to high memory usage,
// especially on large images. It should be used only when it is not possible
// to work with the image byte data directly.
func (img *Image) SetSamples(samples []uint32) {
	if img.BitsPerComponent < 8 {
		samples = img.samplesAddPadding(samples)
	}

	resampled := sampling.ResampleUint32(samples, int(img.BitsPerComponent), 8)
	data := make([]byte, len(resampled))
	for i, val := range resampled {
		data[i] = byte(val)
	}

	img.Data = data
}

// ColorAt returns the color of the image pixel specified by the x and y coordinates.
func (img *Image) ColorAt(x, y int) (gocolor.Color, error) {
	switch img.ColorComponents {
	case 1:
		return img.grayscaleColorAt(x, y)
	case 3:
		return img.rgbColorAt(x, y)
	case 4:
		return img.cmykColorAt(x, y)
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
	if img.BitsPerComponent == targetBitsPerComponent {
		// Nothing changes here.
		return
	}
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
		downSampling := img.BitsPerComponent - targetBitsPerComponent
		for i := range samples {
			samples[i] >>= uint(downSampling)
		}
	} else if targetBitsPerComponent > img.BitsPerComponent {
		upSampling := targetBitsPerComponent - img.BitsPerComponent
		for i := range samples {
			samples[i] <<= uint(upSampling)
		}
	} else {

		return
	}

	img.BitsPerComponent = targetBitsPerComponent

	// Samples data may require extra padding on each row if the bits per component are lower than 8.
	if img.BitsPerComponent < 8 {
		img.resampleLowBits(samples)
		return
	}

	// Write out row by row...
	data := make([]byte, img.BytesPerLine*int(img.Height))
	var (
		ind1, ind2, row, i int
		val                uint32
	)
	for row = 0; row < int(img.Height); row++ {
		ind1 = row * img.BytesPerLine
		ind2 = (row+1)*img.BytesPerLine - 1

		resampledRow := sampling.ResampleUint32(samples[ind1:ind2], int(targetBitsPerComponent), 8)
		for i, val = range resampledRow {
			data[i+ind1] = byte(val)
		}
	}
	img.Data = data
	// While changing the bits per component value for given samples, we also need to change
	// the bytes per line for image.
	img.setBytesPerLine()
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

	img := Image{}
	img.Width = int64(b.Dx())
	img.Height = int64(b.Dy())
	img.BitsPerComponent = 8 // RGBA colormap
	img.ColorComponents = 3
	img.Data = data
	img.setBytesPerLine()

	img.hasAlpha = hasAlpha
	if hasAlpha {
		img.alphaData = alphaData
	}
	return &img, nil
}

// NewGrayImageFromGoImage creates a new grayscale unidoc Image from a golang Image.
func (ih DefaultImageHandler) NewGrayImageFromGoImage(goimg goimage.Image) (*Image, error) {
	b := goimg.Bounds()

	img := &Image{
		Width:            int64(b.Dx()),
		Height:           int64(b.Dy()),
		ColorComponents:  1,
		BitsPerComponent: 8,
	}

	switch t := goimg.(type) {
	case *goimage.Gray:
		if len(t.Pix) != b.Dx()*b.Dy() {
			// Detects when the image Pix data is not of correct format, typically happens
			// when m.Stride does not match the image width (extra bytes at end of each line for example).
			// Rearrange the data back such that the Pix data is arranged consistently.
			// Disadvantage of this is that it doubles the memory use as the data is
			// copied when creating the new structure.
			t = goimage.NewGray(b)
			draw.Draw(t, b, goimg, b.Min, draw.Src)
		}
		img.Data = t.Pix
	case *goimage.Gray16:
		if len(t.Pix) != b.Dx()*b.Dy()*2 {
			// Detects when the image Pix data is not of correct format, typically happens
			// when m.Stride does not match the image width (extra bytes at end of each line for example).
			// Rearrange the data back such that the Pix data is arranged consistently.
			// Disadvantage of this is that it doubles the memory use as the data is
			// copied when creating the new structure.
			t = goimage.NewGray16(b)
			draw.Draw(t, b, goimg, b.Min, draw.Src)
		}
		img.BitsPerComponent = 16
		img.Data = t.Pix
	default:
		g := goimage.NewGray(b)
		draw.Draw(g, b, goimg, b.Min, draw.Src)
		img.Data = g.Pix
	}
	img.setBytesPerLine()
	return img, nil
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

//
// Private color getters.
//

// cmykColorAt gets the color of the CMYK image at 'x' and 'y' coordinates.
func (img *Image) cmykColorAt(x int, y int) (gocolor.Color, error) {
	idx := 4 * (y*int(img.Width) + x)
	if idx+3 >= len(img.Data) {
		return nil, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	return gocolor.CMYK{
		C: img.Data[idx] & 0xff,
		M: img.Data[idx+1] & 0xff,
		Y: img.Data[idx+2] & 0xff,
		K: img.Data[idx+3] & 0xff,
	}, nil
}

func (img *Image) grayscaleColorAt(x, y int) (gocolor.Color, error) {
	switch img.BitsPerComponent {
	case 1:
		return img.grayscaleBitColorAt(x, y)
	case 2:
		return img.grayscaleDiBitColorAt(x, y)
	case 4:
		return img.grayscaleQBitColorAt(x, y)
	case 8:
		return img.grayscale8bitColorAt(x, y)
	case 16:
		return img.grayscale16bitColorAt(x, y)
	default:
		return nil, fmt.Errorf("unsupported gray scale bits per component amount: '%d'", img.BitsPerComponent)
	}
}

// grayscaleBitColorAt gets the color from the 1 bit per component Grayscale image at 'x' and 'y' coordinates.
func (img *Image) grayscaleBitColorAt(x, y int) (gocolor.Gray, error) {
	idx := y*img.BytesPerLine + x>>3
	if idx >= len(img.Data) {
		return gocolor.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	byteValue := img.Data[idx] >> uint(7-(x&7)) & 1
	if len(img.decode) != 2 {
		return gocolor.Gray{Y: byteValue * 255}, nil
	}
	val := float64(byteValue)
	val = interpolate(val, 0, float64(1), img.decode[0], img.decode[1])
	return gocolor.Gray{Y: uint8(val*255) & 0xff}, nil
}

// grayscaleDiBitColorAt gets the color from the 2 bits per component Grayscale image at 'x' and 'y' coordinates.
func (img *Image) grayscaleDiBitColorAt(x, y int) (gocolor.Gray, error) {
	idx := y*img.BytesPerLine + x>>2
	if idx >= len(img.Data) {
		return gocolor.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	byteValue := img.Data[idx] >> uint(6-(x&3)*2) & 3
	if len(img.decode) == 0 {
		return gocolor.Gray{Y: byteValue * 85}, nil
	}
	val := float64(byteValue)
	val = interpolate(val, 0, float64(3), img.decode[0], img.decode[1])
	return gocolor.Gray{Y: uint8(val*255/3.0) & 0xff}, nil
}

// grayscaleQBitColorAt gets the color from the 4 bits per component Grayscale image at 'x' and 'y' coordinates.
func (img *Image) grayscaleQBitColorAt(x, y int) (gocolor.Gray, error) {
	idx := y*img.BytesPerLine + x>>1
	if idx >= len(img.Data) {
		return gocolor.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	byteValue := img.Data[idx] >> uint(4-(x&1)*4) & 15

	if len(img.decode) != 2 {
		return gocolor.Gray{Y: (byteValue * 17) & 0xff}, nil
	}
	val := float64(byteValue)
	val = interpolate(val, 0, float64(15), img.decode[0], img.decode[1])
	return gocolor.Gray{Y: uint8(val * 255 / 15.0)}, nil
}

// grayscale16bitColorAt gets the color from the 8 bits per component Grayscale image at 'x' and 'y' coordinates.
func (img *Image) grayscale8bitColorAt(x, y int) (gocolor.Gray, error) {
	idx := y*int(img.Width) + x
	if idx >= len(img.Data) {
		return gocolor.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	if len(img.decode) != 2 {
		return gocolor.Gray{Y: img.Data[idx]}, nil
	}
	val := interpolate(float64(img.Data[idx]), 0, float64(255), img.decode[0], img.decode[1])
	return gocolor.Gray{Y: uint8(uint32(val) & math.MaxUint8)}, nil
}

// grayscale16bitColorAt gets the color from the 16 bits per component Grayscale image at 'x' and 'y' coordinates.
func (img *Image) grayscale16bitColorAt(x, y int) (gocolor.Gray16, error) {
	idx := (y*int(img.Width) + x) * 2
	if idx+1 >= len(img.Data) {
		return gocolor.Gray16{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	colorValue := uint16(img.Data[idx])<<8 | uint16(img.Data[idx+1])
	if len(img.decode) != 2 {
		return gocolor.Gray16{Y: colorValue}, nil
	}
	val := interpolate(float64(colorValue), 0, float64(math.MaxUint16), img.decode[0], img.decode[1])
	return gocolor.Gray16{Y: uint16(uint32(val) & math.MaxUint16)}, nil
}

// rgbColorAt gets the color from the RGB image at 'x' and 'y' coordinates.
func (img *Image) rgbColorAt(x int, y int) (gocolor.Color, error) {
	switch img.BitsPerComponent {
	case 4:
		return img.rgb4BPCColorAt(x, y)
	case 8:
		return img.rgb8BPCColorAt(x, y)
	case 16:
		return img.rgb16BPCColorAt(x, y)
	default:
		return nil, fmt.Errorf("unsupported rgb image bits per component: '%d'", img.BitsPerComponent)
	}
}

// rgb4BPCColorAt gets the color from the 4 bit per component RGB image at 'x' and 'y' coordinates.
func (img *Image) rgb4BPCColorAt(x int, y int) (gocolor.RGBA, error) {
	// Index in the data is equal to rowIndex (y*img.BytesPerLine) + the numbers of bytes to the right.
	idx := y*img.BytesPerLine + x*3/2
	if idx+1 >= len(img.Data) {
		return gocolor.RGBA{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	// Calculate bit position at which the color data starts.
	var r, g, b uint8
	const max4bitValue = 0xf
	if x*3%2 == 0 {
		// The R and G components are contained by the current byte
		// and the B component is contained by the next byte.
		r = (img.Data[idx] >> uint(4)) & max4bitValue
		g = (img.Data[idx] >> uint(0)) & max4bitValue
		b = (img.Data[idx+1] >> uint(4)) & max4bitValue
	} else {
		// The R component is contained by the current byte and the
		// G and B components are contained by the next byte.
		r = (img.Data[idx] >> uint(0)) & max4bitValue
		g = (img.Data[idx+1] >> uint(4)) & max4bitValue
		b = (img.Data[idx+1] >> uint(0)) & max4bitValue
	}

	const max8BitValue = uint8(0xff)
	return gocolor.RGBA{
		R: uint8(uint32(r)*255/max4bitValue) & max8BitValue,
		G: uint8(uint32(g)*255/max4bitValue) & max8BitValue,
		B: uint8(uint32(b)*255/max4bitValue) & max8BitValue,
		A: max8BitValue,
	}, nil
}

// rgb8BPCColorAt gets the color from the 8 bit per component RGB image at 'x' and 'y' coordinates.
func (img *Image) rgb8BPCColorAt(x int, y int) (gocolor.RGBA, error) {
	idx := y*int(img.Width) + x

	i := 3 * idx
	if i+2 >= len(img.Data) {
		return gocolor.RGBA{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	a := uint8(0xff)
	if img.alphaData != nil && len(img.alphaData) > idx {
		a = img.alphaData[idx]
	}

	return gocolor.RGBA{
		R: img.Data[i],
		G: img.Data[i+1],
		B: img.Data[i+2],
		A: a,
	}, nil
}

// rgb16BPCColorAt gets the color from the 16 bit per component RGB image at 'x' and 'y' coordinates.
func (img *Image) rgb16BPCColorAt(x int, y int) (gocolor.RGBA64, error) {
	idx := (y*int(img.Width) + x) * 2

	i := idx * 3
	if i+5 >= len(img.Data) {
		return gocolor.RGBA64{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	a := uint16(0xffff)
	if img.alphaData != nil && len(img.alphaData) > idx+1 {
		a = uint16(img.alphaData[idx])<<8 | uint16(img.alphaData[idx+1])
	}

	return gocolor.RGBA64{
		R: uint16(img.Data[i])<<8 | uint16(img.Data[i+1]),
		G: uint16(img.Data[i+2])<<8 | uint16(img.Data[i+3]),
		B: uint16(img.Data[i+4])<<8 | uint16(img.Data[i+5]),
		A: a,
	}, nil
}

func (img *Image) resampleLowBits(samples []uint32) {
	// Set new BytesPerLine value.
	img.setBytesPerLine()

	// Try to create full bytes from provided samples and set to the given data output.
	data := make([]byte, img.ColorComponents*img.BytesPerLine*int(img.Height))

	// Define how many samples are stored within a single line.
	samplesPerLine := int(img.BitsPerComponent) * img.ColorComponents * int(img.Width)
	bitPosition := uint8(8)
	var (
		sampleIndex, byteIndex int
		sample                 uint32
	)
	for y := 0; y < int(img.Height); y++ {
		byteIndex = y * img.BytesPerLine
		for i := 0; i < samplesPerLine; i++ {
			sample = samples[sampleIndex]

			// The bitPosition defines current bit position where the value of the sample should be added.
			bitPosition -= uint8(img.BitsPerComponent)
			data[byteIndex] |= byte(sample) << bitPosition
			if bitPosition == 0 {
				bitPosition = 8
				byteIndex++
			}
			sampleIndex++
		}
	}
	img.Data = data
}

func (img *Image) samplesAddPadding(samples []uint32) []uint32 {
	paddedLen := img.ColorComponents * img.BytesPerLine * int(img.Height)
	if len(samples) == paddedLen {
		return samples
	}
	paddedSamples := make([]uint32, paddedLen)
	samplesPerTrimmedRow := int(img.Width) * img.ColorComponents
	for row := 0; row < int(img.Height); row++ {
		rowIndexTrimmed := row * int(img.Width)
		rowIndexPadded := row * img.BytesPerLine
		for i := 0; i < samplesPerTrimmedRow; i++ {
			paddedSamples[rowIndexPadded+i] = samples[rowIndexTrimmed+i]
		}
	}
	return paddedSamples
}

func (img *Image) samplesTrimPadding(samples []uint32) []uint32 {
	// initialize the samples that wouldn't contain extra padding at the end of each row.
	trimmed := make([]uint32, int64(img.ColorComponents)*img.Width*img.Height)
	// get number of samples per trimmed row
	samplesPerTrimmedRow := int(img.Width) * img.ColorComponents
	var row, rowIndexTrimmed, rowIndexUntrimmed, i int
	// trim extra padding from each row and set it on the trimmed slice.
	for row = 0; row < int(img.Height); row++ {
		rowIndexTrimmed = row * int(img.Width)
		rowIndexUntrimmed = row * img.BytesPerLine
		for i = 0; i < samplesPerTrimmedRow; i++ {
			trimmed[rowIndexTrimmed+i] = samples[rowIndexUntrimmed+i]
		}
	}
	return trimmed
}

func (img *Image) setBytesPerLine() {
	img.BytesPerLine = (int(img.Width)*int(img.BitsPerComponent)*img.ColorComponents + 7) >> 3
}
