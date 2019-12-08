/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bytes"
	"image"
	"image/color"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// JBIG2CompressionType defines the enum compression type used by the JBIG2Encoder
type JBIG2CompressionType int

const (
	// JB2Generic is the JBIG2 compression type that uses generic region see 6.2.
	JB2Generic JBIG2CompressionType = iota
	// JB2SymbolCorrelation is the JBIG2 compression type that uses symbol dictionary and text region encoding procedure
	// with the correlation classification.
	JB2SymbolCorrelation
	// JB2SymbolRankHaus is the JBIG2 compression type that uses symbol dictionary and text region encoding procedure
	// with the rank hausdorff classification. RankHausMode uses the rank Hausdorff method that classifies the input images.
	// It is more robust, more susceptible to confusing components that should be in different classes.
	JB2SymbolRankHaus
)

/**

JBIG2Encoder/Decoder

*/

// JBIG2Encoder implements both jbig2 encoder and the decoder. The encoder allows to encode
// provided images (best used document scans) in multiple way. By default it uses single page generic
// encoder. It allows to store lossless data as a single segment. In order to obtain better compression results
// the encoder allows to encode the input in a lossy or lossless way with a component (symbol) mode. It divides the image into components.
// Then checks if any component is 'simillar' to the others and maps them together. The symbol classes are stored
// in the dictionary. Then the encoder creates text regions which uses the related symbol classes to fill it's space.
// The similarity is defined by the 'Threshold' variable (default: 0.95). The less the value is, the more components
// matches to single class, thus the compression is better, but the result might become lossy.
// In order to store multiple image documents use the 'FileMode' which allows to store more pages within single jbig2 document.
type JBIG2Encoder struct {
	JBIG2Document
	// Globals are the JBIG2 global segments.
	Globals document.Globals
	// IsChocolateData defines if the data is encoded such that
	// binary data '1' means black and '0' white.
	// otherwise the data is called vanilla.
	// Naming convention taken from: 'https://en.wikipedia.org/wiki/Binary_image#Interpretation'
	IsChocolateData bool
	// Document defines the JBIG2 Encoded document

	// DefaultPageSettings are the settings parameters used by the jbig2 encoder.
	DefaultPageSettings JBIG2PageSettings
}

// DecodeBytes decodes a slice of JBIG2 encoded bytes and returns the results.
func (enc *JBIG2Encoder) DecodeBytes(encoded []byte) ([]byte, error) {
	parameters := decoder.Parameters{UnpaddedData: true}
	if enc.IsChocolateData {
		parameters.Color = bitmap.Chocolate
	}
	return jbig2.DecodeBytes(encoded, parameters, enc.Globals)
}

// DecodeImages decodes the page images from the jbig2 'encoded' data input.
// The jbig2 document may contain multiple pages, thus the function can return multiple
// images. The images order corresponds to the page number.
func (enc *JBIG2Encoder) DecodeImages(encoded []byte) ([]image.Image, error) {
	const processName = "JBIG2Encoder.DecodeImages"
	parameters := decoder.Parameters{UnpaddedData: true}
	if enc.IsChocolateData {
		parameters.Color = bitmap.Chocolate
	}
	// create decoded document.
	d, err := decoder.Decode(encoded, parameters, enc.Globals)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	// get page number in the document.
	pageNumber, err := d.PageNumber()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	// decode all images
	images := []image.Image{}
	var img image.Image
	for i := 1; i <= pageNumber; i++ {
		img, err = d.DecodePageImage(i)
		if err != nil {
			return nil, errors.Wrapf(err, processName, "page: '%d'", i)
		}
		images = append(images, img)
	}
	return images, nil
}

// DecodeStream decodes a JBIG2 encoded stream and returns the result as a slice of bytes.
func (enc *JBIG2Encoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return enc.DecodeBytes(streamObj.Stream)
}

// EncodeBytes encodes the passed slice in slice of bytes into JBIG2.
// The input 'data' must be an image. In order to Decode it a user is responsible to
// load the codec ('png', 'jpg').
// Returns jbig2 single page encoded document byte slice. The encoder uses DefaultPageSettings
// to encode given image.
func (enc *JBIG2Encoder) EncodeBytes(data []byte) ([]byte, error) {
	const processName = "JBIG2Encoder.EncodeBytes"
	if len(data) == 0 {
		return nil, errors.Errorf(processName, "input 'data' not defined")
	}
	i, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, processName, "decode input image")
	}
	encoded, err := enc.encodeImage(i)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return encoded, nil
}

// GetFilterName returns the name of the encoding filter.
func (enc *JBIG2Encoder) GetFilterName() string {
	return StreamEncodingFilterNameJBIG2
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on the current encoder settings.
func (enc *JBIG2Encoder) MakeDecodeParams() PdfObject {
	return MakeDict()
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
func (enc *JBIG2Encoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()
	if enc.IsChocolateData {
		// /Decode[1.0 0.0] - see note in the 'setChocolateData' method.
		dict.Set("Decode", MakeArray(MakeFloat(1.0), MakeFloat(0.0)))
	}
	dict.Set("Filter", MakeName(enc.GetFilterName()))
	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *JBIG2Encoder) UpdateParams(params *PdfObjectDictionary) {
	if decode := params.Get("Decode"); decode != nil {
		enc.setChocolateData(decode)
	}
}

func (enc *JBIG2Encoder) encodeImage(i image.Image) ([]byte, error) {
	const processName = "encodeImage"
	// conver the input into jbig2 image
	jimg, err := GoImageToJBIG2(i, 0.5)
	if err != nil {
		return nil, errors.Wrap(err, processName, "convert input image to jbig2 img")
	}
	if err = enc.AddPageImage(jimg, enc.DefaultPageSettings); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	common.Log.Debug("Error: Attempting to use unsupported encoding %s", enc.GetFilterName())
	return nil, ErrNoJBIG2Decode
}

// setChocolateData sets the chocolate data flag when the pdf stream object contains the 'Decode' object.
// Decode object ( PDF32000:2008 7.10.2 Type 0 (Sampled) Functions).
// NOTE: this function is a temporary helper until the samples handle Decode function.
func (enc *JBIG2Encoder) setChocolateData(decode PdfObject) {
	arr, ok := decode.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("JBIG2Encoder - Decode is not an array. %T", decode)
		return
	}

	// (PDF32000:2008 Table 39) The array should be of 2 x n size.
	// For binary images n stands for 1bit, thus the array should contain 2 numbers.
	vals, err := arr.GetAsFloat64Slice()
	if err != nil {
		common.Log.Debug("JBIG2Encoder unsupported Decode value. %s", arr.String())
		return
	}

	if len(vals) != 2 {
		return
	}

	first, second := int(vals[0]), int(vals[1])
	if first == 1 && second == 0 {
		enc.IsChocolateData = true
	} else if first == 0 && second == 1 {
		enc.IsChocolateData = false
	} else {
		common.Log.Debug("JBIG2Encoder unsupported DecodeParams->Decode value: %s", arr.String())
	}
}

func newJBIG2DecoderFromStream(streamObj *PdfObjectStream, decodeParams *PdfObjectDictionary) (*JBIG2Encoder, error) {
	const processName = "newJBIG2DecoderFromStream"
	encoder := &JBIG2Encoder{}
	encDict := streamObj.PdfObjectDictionary
	if encDict == nil {
		// No encoding dictionary.
		return encoder, nil
	}

	// If decodeParams not provided, see if we can get from the stream.
	if decodeParams == nil {
		obj := encDict.Get("DecodeParms")
		if obj != nil {
			switch t := obj.(type) {
			case *PdfObjectDictionary:
				decodeParams = t
			case *PdfObjectArray:
				if t.Len() == 1 {
					if dp, ok := GetDict(t.Get(0)); ok {
						decodeParams = dp
					}
				}
			default:
				common.Log.Error("DecodeParams not a dictionary %#v", obj)
				return nil, errors.Errorf(processName, "invalid DecodeParms type: %T", t)
			}
		}
	}

	if decodeParams != nil {
		if globals := decodeParams.Get("JBIG2Globals"); globals != nil {
			var err error

			globalsStream, ok := globals.(*PdfObjectStream)
			if !ok {
				err = errors.Error(processName, "jbig2.Globals stream should be an Object Stream")
				common.Log.Debug("ERROR: %s", err.Error())
				return nil, err
			}
			encoder.Globals, err = jbig2.DecodeGlobals(globalsStream.Stream)
			if err != nil {
				err = errors.Wrap(err, processName, "corrupted jbig2 encoded data")
				common.Log.Debug("ERROR: %s", err)
				return nil, err
			}
		}
	}

	// Inverse the bits on the 'Decode [1.0 0.0]' function (PDF32000:2008 7.10.2)
	if decode := streamObj.Get("Decode"); decode != nil {
		encoder.setChocolateData(decode)
	}
	return encoder, nil
}

/**

JBIG2Document

*/

// JBIG2Document defines the JBIG2 document which contains pages to encode.
type JBIG2Document struct {
	// FileMode defines if the jbig2 encoder should return full jbig2 file instead of
	// shortened pdf mode. This adds the file header to the jbig2 definition.
	FileMode bool
	d        *document.Document
}

// AddPageImage adds the page with the image 'img' to the document 'd'.
// The 'settings' defines what encoding type should be used by the encoder.
func (d *JBIG2Document) AddPageImage(img *JBIG2Image, settings JBIG2PageSettings) (err error) {
	const processName = "JBIG2Document.AddPageImage"
	if d == nil {
		return errors.Error(processName, "JBIG2Document is nil")
	}
	d.initializeIfNeeded()

	if err = settings.Validate(); err != nil {
		return errors.Wrap(err, processName, "")
	}

	// convert input 'img' to the bitmap.Bitmap
	b, err := img.toBitmap()
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	switch settings.Compression {
	case JB2Generic:
		if err = d.d.AddGenericPage(b, settings.DuplicatedLinesRemoval); err != nil {
			return errors.Wrap(err, processName, "")
		}
	case JB2SymbolCorrelation:
		return errors.Error(processName, "symbol correlation encoding not implemented yet")
	case JB2SymbolRankHaus:
		return errors.Error(processName, "symbol rank haus encoding not implemented yet")
	}
	return nil
}

// Encode encodes whole document and stores as the byte slice data.
func (d *JBIG2Document) Encode() (data []byte, err error) {
	const processName = "JBIG2Document.Encode"
	if d.d == nil {
		return nil, errors.Errorf(processName, "document input data not defined")
	}
	d.updateParameters()
	// encode the document
	data, err = d.d.Encode()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return data, nil
}

func (d *JBIG2Document) initializeIfNeeded() {
	// check if the document is initialized
	if d.d == nil {
		d.d = document.InitEncodeDocument(d.FileMode)
	}
}

func (d *JBIG2Document) updateParameters() {
	d.d.FullHeaders = d.FileMode
}

/**

JBIG2Image

*/

// JBIG2Image is the image structure used by the jbig2 encoder.
// This image should be used as one bit per component as the jbig2 accepts only binary images.
// In order to create binary image use the model.Image.ToJBIG2() method or use GoImageToJBIG2 function.
// If the image data contains the row bytes padding set the HasPadding to true.
type JBIG2Image struct {
	// Width and Height defines the image boundaries.
	Width, Height int
	// Data is the byte slice data for the input image
	Data []byte
	// HasPadding is the attribute that defines if the last byte of the data in the row contains
	// 0 bits padding.
	HasPadding bool
}

// ToGoImage converts the JBIG2Image to the golang image.Image.
func (j *JBIG2Image) ToGoImage() (image.Image, error) {
	const processName = "JBIG2Image.ToGoImage"
	bm, err := j.toBitmap()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return bm.ToImage(), nil
}

func (j *JBIG2Image) toBitmap() (b *bitmap.Bitmap, err error) {
	const processName = "JBIG2Image.toBitmap"
	if j.Data == nil {
		return nil, errors.Error(processName, "image data not defined")
	}
	if j.Width == 0 || j.Height == 0 {
		return nil, errors.Error(processName, "image height or width not defined")
	}
	// check if the data already has padding
	if j.HasPadding {
		b, err = bitmap.NewWithData(j.Width, j.Height, j.Data)
	} else {
		b, err = bitmap.NewWithUnpaddedData(j.Width, j.Height, j.Data)
	}
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return b, nil
}

// GoImageToJBIG2 creates a binary image on the base of 'i' golang image.Image.
// The 'bwThreshold' value should be in range (0.0, 1.0). The threshold checks if the grayscaled
// pixel (uint) value is greater or smaller than 'bwThreshold' * 255. Pixels inside the range will be white, and the others will be black
// If the 'bwThreshold' is equal to 0.0 then it's value would be set on the base of it's histogram using Triangle method.
// For more information go to:
// 	https://www.mathworks.com/matlabcentral/fileexchange/28047-gray-image-thresholding-using-the-triangle-method
func GoImageToJBIG2(i image.Image, bwThreshold float64) (*JBIG2Image, error) {
	const processName = "GoImageToJBIG2"
	if i == nil {
		return nil, errors.Error(processName, "image 'i' not defined")
	}
	var th uint8
	if bwThreshold == 0.0 {
		// autoThreshold using triangle method
		gray := imgToGray(i)
		histogram := grayImageHistogram(gray)
		th = autoThresholdTriangle(histogram)
		i = gray
	} else if bwThreshold > 1.0 || bwThreshold < 0.0 {
		// check if bwThreshold is unknown - set to 0.0 is not in the allowed range.
		return nil, errors.Error(processName, "provided threshold is not in a range {0.0, 1.0}")
	} else {
		th = uint8(255 * bwThreshold)
	}
	gray := imgToBinary(i, th)
	return bwToJBIG2Image(gray), nil
}

func bwToJBIG2Image(i *image.Gray) *JBIG2Image {
	bounds := i.Bounds()
	// compute the rowstride - number of bytes in the row.
	bm := bitmap.New(bounds.Dx(), bounds.Dy())
	ji := &JBIG2Image{Height: bounds.Dy(), Width: bounds.Dx(), HasPadding: true}
	// allocate the byte slice data
	var pix color.Gray
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			pix = i.GrayAt(x, y)
			// check if the pixel is black or white
			// where black pixel would be stored as '1' bit
			// and the white as '0' bit.

			// the pix is color.Black if it's Y value is '0'.
			if pix.Y == 0 {
				if err := bm.SetPixel(x, y, 1); err != nil {
					common.Log.Debug("can't set pixel at bitmap: %v", bm)
				}
			}
		}
	}
	ji.Data = bm.Data
	return ji
}

/**

JBIG2Page

*/

// JBIG2Page defines the JBIG2 page which contains the segment definitions.
type JBIG2Page struct {
	Settings JBIG2PageSettings
}

// JBIG2PageSettings contains the parameters and settings used by the JBIG2Encoder
type JBIG2PageSettings struct {
	// ResolutionX defines the 'x' axis input image resolution - used for single page encoding.
	ResolutionX int
	// ResolutionY defines the 'y' axis input image resolution - used for single page encoding.
	ResolutionY int
	// Threshold defines the threshold of the image corelation for
	// non generic compression.
	// Best results in range [0.7 - 0.98] - the less the better the compression would be
	// but the more lossy.
	// Default value: 0.95
	Threshold float64
	// Compression defines the compression type used for encoding the page.
	Compression JBIG2CompressionType
	// DuplicatedLinesRemoval code generic region in a way such that if the lines are duplicated the encoder
	// doesn't store it twice.
	DuplicatedLinesRemoval bool
	// DefaultPixelValue is the bit value initial for every pixel in the page.
	DefaultPixelValue uint8
}

// Validate validates the page settings for the JBIG2 encoder.
func (s JBIG2PageSettings) Validate() error {
	const processName = "validateEncoder"
	if s.Threshold < 0 || s.Threshold > 1.0 {
		return errors.Errorf(processName, "provided threshold value: '%v' must be in range [0.0, 1.0]", s.Threshold)
	}
	if s.ResolutionX < 0 {
		return errors.Errorf(processName, "provided x resoulution: '%d' must be positive value", s.ResolutionX)
	}
	if s.ResolutionY < 0 {
		return errors.Errorf(processName, "provided y resoulution: '%d' must be positive value", s.ResolutionY)
	}
	if s.DefaultPixelValue != 0 && s.DefaultPixelValue != 1 {
		return errors.Errorf(processName, "default pixel value: '%d' must be a value for the bit: {0,1}", s.DefaultPixelValue)
	}
	return nil
}

/**

private functions

*/

func autoThresholdTriangle(histogram [256]int) uint8 {
	var min, dmax, max, min2 int
	// find the min and the max of the histogram
	for i := 0; i < len(histogram); i++ {
		if histogram[i] > 0 {
			min = i
			break
		}
	}
	if min > 0 {
		min--
	}
	for i := 255; i > 0; i-- {
		if histogram[i] > 0 {
			min2 = i
			break
		}
	}
	if min2 < 255 {
		min2++
	}

	for i := 0; i < 256; i++ {
		if histogram[i] > dmax {
			max = i
			dmax = histogram[i]
		}
	}

	// find which is the furthest side
	var inverted bool
	if (max - min) < (min2 - max) {
		// reverse the histogram
		inverted = true
		var left int
		right := 255
		for left < right {
			temp := histogram[left]
			histogram[left] = histogram[right]
			histogram[right] = temp
			left++
			right--
		}
		min = 255 - min2
		max = 255 - max
	}

	if min == max {
		return uint8(min)
	}
	// nx is the max frequency
	nx := float64(histogram[max])
	ny := float64(min - max)
	d := math.Sqrt(nx*nx + ny*ny)
	nx /= d
	ny /= d
	d = nx*float64(min) + ny*float64(histogram[min])

	// find the split point
	split := min
	var splitDistance float64
	for i := min + 1; i <= max; i++ {
		newDistance := nx*float64(i) + ny*float64(histogram[i]) - d
		if newDistance > splitDistance {
			split = i
			splitDistance = newDistance
		}
	}
	split--
	if inverted {
		var left int
		right := 255
		for left < right {
			temp := histogram[left]
			histogram[left] = histogram[right]
			histogram[right] = temp
			left++
			right--
		}
		return uint8(255 - split)
	}
	return uint8(split)
}

func blackOrWhite(c, threshold uint8) uint8 {
	if c < threshold {
		return 255
	}
	return 0
}

func gray16ImageToBlackWhite(img *image.Gray16, th uint8) *image.Gray {
	bounds := img.Bounds()
	d := image.NewGray(bounds)
	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			pix := img.Gray16At(x, y)
			d.SetGray(x, y, color.Gray{blackOrWhite(uint8(pix.Y/256), th)})
		}
	}
	return d
}

func grayImageHistogram(img *image.Gray) (histogram [256]int) {
	for _, pix := range img.Pix {
		histogram[pix]++
	}
	return histogram
}

func grayImageToBlackWhite(img *image.Gray, th uint8) *image.Gray {
	bounds := img.Bounds()
	d := image.NewGray(bounds)
	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			c := img.GrayAt(x, y)
			d.SetGray(x, y, color.Gray{blackOrWhite(c.Y, th)})
		}
	}
	return d
}

func imgToBinary(i image.Image, threshold uint8) *image.Gray {
	switch img := i.(type) {
	case *image.Gray:
		if isGrayBlackWhite(img) {
			return img
		}
		return grayImageToBlackWhite(img, threshold)
	case *image.Gray16:
		return gray16ImageToBlackWhite(img, threshold)
	default:
		return rgbImageToBlackWhite(img, threshold)
	}
}

func imgToGray(i image.Image) *image.Gray {
	if g, ok := i.(*image.Gray); ok {
		return g
	}
	bounds := i.Bounds()
	g := image.NewGray(bounds)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := i.At(x, y)
			g.Set(x, y, c)
		}
	}
	return g
}

func isGrayBlackWhite(img *image.Gray) bool {
	for i := 0; i < len(img.Pix); i++ {
		if !isPix8BlackWhite(img.Pix[i]) {
			return false
		}
	}
	return true
}

func isPix8BlackWhite(pix uint8) bool {
	if pix == 0 || pix == 255 {
		return true
	}
	return false
}

func rgbImageToBlackWhite(i image.Image, th uint8) *image.Gray {
	bounds := i.Bounds()
	gray := image.NewGray(bounds)
	var (
		c  color.Color
		cg color.Gray
	)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			// get the color at x,y
			c = i.At(x, y)
			// set it to the grayscale
			gray.Set(x, y, c)
			// get the grayscale color value
			cg = gray.GrayAt(x, y)
			// set the black/white pixel at 'x', 'y'
			gray.SetGray(x, y, color.Gray{blackOrWhite(cg.Y, th)})
		}
	}
	return gray
}
