/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	goimage "image"
	gocolor "image/color"
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
	img    imageutil.Image
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
		threshold := imageutil.AutoThresholdTriangle(imageutil.GrayHistogram(gray))
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
// image directly or use the At method to extract individual pixels.
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

// SetDecode sets the decode image float slice.
func (img *Image) SetDecode(decode []float64) {
	img.decode = decode
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
	bytesPerLine := imageutil.BytesPerLine(int(img.Width), int(img.BitsPerComponent), img.ColorComponents)
	switch img.ColorComponents {
	case 1:
		return imageutil.ColorAtGrayscale(x, y, int(img.BitsPerComponent), bytesPerLine, img.Data, img.decode)
	case 3:
		return imageutil.ColorAtNRGBA(x, y, int(img.Width), bytesPerLine, int(img.BitsPerComponent), img.Data, img.alphaData, img.decode)
	case 4:
		return imageutil.ColorAtCMYK(x, y, int(img.Width), img.Data, img.decode)
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
		// Nothing to resample.
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
	}

	img.BitsPerComponent = targetBitsPerComponent

	// Samples data may require extra padding on each row if the bits per component are lower than 8.
	if img.BitsPerComponent < 8 {
		img.resampleLowBits(samples)
		return
	}
	bytesPerLine := imageutil.BytesPerLine(int(img.Width), int(img.BitsPerComponent), img.ColorComponents)
	// Write out row by row...
	data := make([]byte, bytesPerLine*int(img.Height))
	var (
		ind1, ind2, row, i int
		val                uint32
	)
	for row = 0; row < int(img.Height); row++ {
		ind1 = row * bytesPerLine
		ind2 = (row+1)*bytesPerLine - 1

		resampledRow := sampling.ResampleUint32(samples[ind1:ind2], int(targetBitsPerComponent), 8)
		for i, val = range resampledRow {
			data[i+ind1] = byte(val)
		}
	}
	img.Data = data
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
	iimg, err := imageutil.NewImage(int(img.Width), int(img.Height), int(img.BitsPerComponent), img.ColorComponents, img.Data, img.alphaData, img.decode)
	if err != nil {
		return nil, err
	}
	return iimg, nil
}

// ImageHandler interface implements common image loading and processing tasks.
// Implementing as an interface allows for the possibility to use non-standard libraries for faster
// loading and processing of images.
type ImageHandler interface {
	// Read any image type and load into a new Image object.
	Read(r io.Reader) (*Image, error)

	// NewImageFromGoImage loads a NRGBA32 unidoc Image from a standard Go image structure.
	NewImageFromGoImage(goimg goimage.Image) (*Image, error)

	// NewGrayImageFromGoImage loads a grayscale unidoc Image from a standard Go image structure.
	NewGrayImageFromGoImage(goimg goimage.Image) (*Image, error)

	// Compress an image.
	Compress(input *Image, quality int64) (*Image, error)
}

// DefaultImageHandler is the default implementation of the ImageHandler using the standard go library.
type DefaultImageHandler struct{}

// NewImageFromGoImage creates a new NRGBA32 unidoc Image from a golang Image.
// If `goimg` is grayscale (*goimage.Gray8) then calls NewGrayImageFromGoImage instead.
func (ih DefaultImageHandler) NewImageFromGoImage(goimg goimage.Image) (*Image, error) {
	img, err := imageutil.FromGoImage(goimg)
	if err != nil {
		return nil, err
	}
	res := imageFromBase(img.Base())
	return &res, nil
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
			iImg, err := imageutil.GrayConverter.Convert(goimg)
			if err != nil {
				return nil, err
			}
			img.Data = iImg.Pix()
		} else {
			img.Data = t.Pix
		}
	case *goimage.Gray16:
		img.BitsPerComponent = 16
		if len(t.Pix) != b.Dx()*b.Dy()*2 {
			// Detects when the image Pix data is not of correct format, typically happens
			// when m.Stride does not match the image width (extra bytes at end of each line for example).
			// Rearrange the data back such that the Pix data is arranged consistently.
			// Disadvantage of this is that it doubles the memory use as the data is
			// copied when creating the new structure.
			iImg, err := imageutil.Gray16Converter.Convert(goimg)
			if err != nil {
				return nil, err
			}
			img.Data = iImg.Pix()
		} else {
			img.Data = t.Pix
		}
	default:
		iImg, err := imageutil.GrayConverter.Convert(goimg)
		if err != nil {
			return nil, err
		}
		img.Data = iImg.Pix()
	}
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

func (img *Image) resampleLowBits(samples []uint32) {
	bytesPerLine := imageutil.BytesPerLine(int(img.Width), int(img.BitsPerComponent), img.ColorComponents)
	// Try to create full bytes from provided samples and set to the given data output.
	data := make([]byte, img.ColorComponents*bytesPerLine*int(img.Height))

	// Define how many samples are stored within a single line.
	samplesPerLine := int(img.BitsPerComponent) * img.ColorComponents * int(img.Width)
	bitPosition := uint8(8)
	var (
		sampleIndex, byteIndex int
		sample                 uint32
	)
	for y := 0; y < int(img.Height); y++ {
		byteIndex = y * bytesPerLine
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

	samplesPerLine := imageutil.BytesPerLine(int(img.Width), int(img.BitsPerComponent), img.ColorComponents) * (8 / int(img.BitsPerComponent))
	paddedLen := samplesPerLine * int(img.Height)
	if len(samples) == paddedLen {
		return samples
	}
	paddedSamples := make([]uint32, paddedLen)
	samplesPerTrimmedRow := int(img.Width) * img.ColorComponents
	for row := 0; row < int(img.Height); row++ {
		rowIndexTrimmed := row * int(img.Width)
		rowIndexPadded := row * samplesPerLine
		for i := 0; i < samplesPerTrimmedRow; i++ {
			paddedSamples[rowIndexPadded+i] = samples[rowIndexTrimmed+i]
		}
	}
	return paddedSamples
}

func (img *Image) samplesTrimPadding(samples []uint32) []uint32 {
	// initialize the samples that wouldn't contain extra padding at the end of each row.
	trimmedLength := img.ColorComponents * int(img.Width) * int(img.Height)
	if len(samples) == trimmedLength {
		return samples
	}
	trimmed := make([]uint32, trimmedLength)
	// get number of samples per trimmed row
	samplesPerTrimmedRow := int(img.Width) * img.ColorComponents
	var row, rowIndexTrimmed, rowIndexUntrimmed, i int
	bytesPerLine := imageutil.BytesPerLine(int(img.Width), int(img.BitsPerComponent), img.ColorComponents)
	// trim extra padding from each row and set it on the trimmed slice.
	for row = 0; row < int(img.Height); row++ {
		rowIndexTrimmed = row * int(img.Width)
		rowIndexUntrimmed = row * bytesPerLine
		for i = 0; i < samplesPerTrimmedRow; i++ {
			trimmed[rowIndexTrimmed+i] = samples[rowIndexUntrimmed+i]
		}
	}
	return trimmed
}
func imageFromBase(src *imageutil.ImageBase) (dst Image) {
	dst.Width = int64(src.Width)
	dst.Height = int64(src.Height)
	dst.BitsPerComponent = int64(src.BitsPerComponent)
	dst.ColorComponents = src.ColorComponents
	dst.Data = src.Data
	dst.decode = src.Decode
	dst.alphaData = src.Alpha
	return dst
}
