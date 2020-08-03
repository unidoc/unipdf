/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"image"
	"image/color"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/imageutil"

	"github.com/unidoc/unipdf/v3/internal/jbig2"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// JBIG2CompressionType defines the enum compression type used by the JBIG2Encoder.
type JBIG2CompressionType int

const (
	// JB2Generic is the JBIG2 compression type that uses generic region see 6.2.
	JB2Generic JBIG2CompressionType = iota
	// JB2SymbolCorrelation is the JBIG2 compression type that uses symbol dictionary and text region encoding procedure
	// with the correlation classification.
	// NOT IMPLEMENTED YET.
	JB2SymbolCorrelation
	// JB2SymbolRankHaus is the JBIG2 compression type that uses symbol dictionary and text region encoding procedure
	// with the rank hausdorff classification. RankHausMode uses the rank Hausdorff method that classifies the input images.
	// It is more robust, more susceptible to confusing components that should be in different classes.
	// NOT IMPLEMENTED YET.
	JB2SymbolRankHaus
)

// JB2ImageAutoThreshold is the const value used by the 'GoImageToJBIG2Image function' used to set auto threshold
// for the histogram.
const JB2ImageAutoThreshold = -1.0

//
// JBIG2Encoder/Decoder
//

// JBIG2Encoder implements both jbig2 encoder and the decoder. The encoder allows to encode
// provided images (best used document scans) in multiple way. By default it uses single page generic
// encoder. It allows to store lossless data as a single segment.
// In order to store multiple image pages use the 'FileMode' which allows to store more pages within single jbig2 document.
// WIP: In order to obtain better compression results the encoder would allow to encode the input in a
// lossy or lossless way with a component (symbol) mode. It divides the image into components.
// Then checks if any component is 'similar' to the others and maps them together. The symbol classes are stored
// in the dictionary. Then the encoder creates text regions which uses the related symbol classes to fill it's space.
// The similarity is defined by the 'Threshold' variable (default: 0.95). The less the value is, the more components
// matches to single class, thus the compression is better, but the result might become lossy.
type JBIG2Encoder struct {
	// These values are required to be set for the 'EncodeBytes' method.
	// ColorComponents defines the number of color components for provided image.
	ColorComponents int
	// BitsPerComponent is the number of bits that stores per color component
	BitsPerComponent int
	// Width is the width of the image to encode
	Width int
	// Height is the height of the image to encode.
	Height int

	// Encode Page and Decode parameters
	d *document.Document
	// Globals are the JBIG2 global segments.
	Globals jbig2.Globals
	// IsChocolateData defines if the data is encoded such that
	// binary data '1' means black and '0' white.
	// otherwise the data is called vanilla.
	// Naming convention taken from: 'https://en.wikipedia.org/wiki/Binary_image#Interpretation'
	IsChocolateData bool
	// DefaultPageSettings are the settings parameters used by the jbig2 encoder.
	DefaultPageSettings JBIG2EncoderSettings
}

// NewJBIG2Encoder creates a new JBIG2Encoder.
func NewJBIG2Encoder() *JBIG2Encoder {
	return &JBIG2Encoder{
		d: document.InitEncodeDocument(false),
	}
}

// AddPageImage adds the page with the image 'img' to the encoder context in order to encode it jbig2 document.
// The 'settings' defines what encoding type should be used by the encoder.
func (enc *JBIG2Encoder) AddPageImage(img *JBIG2Image, settings *JBIG2EncoderSettings) (err error) {
	const processName = "JBIG2Document.AddPageImage"
	if enc == nil {
		return errors.Error(processName, "JBIG2Document is nil")
	}
	if settings == nil {
		settings = &enc.DefaultPageSettings
	}
	if enc.d == nil {
		enc.d = document.InitEncodeDocument(settings.FileMode)
	}

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
		if err = enc.d.AddGenericPage(b, settings.DuplicatedLinesRemoval); err != nil {
			return errors.Wrap(err, processName, "")
		}
	case JB2SymbolCorrelation:
		return errors.Error(processName, "symbol correlation encoding not implemented yet")
	case JB2SymbolRankHaus:
		return errors.Error(processName, "symbol rank haus encoding not implemented yet")
	default:
		return errors.Error(processName, "provided invalid compression")
	}
	return nil
}

// DecodeBytes decodes a slice of JBIG2 encoded bytes and returns the results.
func (enc *JBIG2Encoder) DecodeBytes(encoded []byte) ([]byte, error) {
	parameters := decoder.Parameters{UnpaddedData: true}
	return jbig2.DecodeBytes(encoded, parameters, enc.Globals)
}

// DecodeGlobals decodes 'encoded' byte stream and returns their Globally defined segments ('Globals').
func (enc *JBIG2Encoder) DecodeGlobals(encoded []byte) (jbig2.Globals, error) {
	return jbig2.DecodeGlobals(encoded)
}

// DecodeImages decodes the page images from the jbig2 'encoded' data input.
// The jbig2 document may contain multiple pages, thus the function can return multiple
// images. The images order corresponds to the page number.
func (enc *JBIG2Encoder) DecodeImages(encoded []byte) ([]image.Image, error) {
	const processName = "JBIG2Encoder.DecodeImages"
	parameters := decoder.Parameters{UnpaddedData: true}
	// create decoded document.
	d, err := decoder.Decode(encoded, parameters, enc.Globals.ToDocumentGlobals())
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

// EncodeBytes encodes slice of bytes into JBIG2 encoding format.
// The input 'data' must be an image. In order to Decode it a user is responsible to
// load the codec ('png', 'jpg').
// Returns jbig2 single page encoded document byte slice. The encoder uses DefaultPageSettings
// to encode given image.
func (enc *JBIG2Encoder) EncodeBytes(data []byte) ([]byte, error) {
	const processName = "JBIG2Encoder.EncodeBytes"
	if enc.ColorComponents != 1 || enc.BitsPerComponent != 1 {
		return nil, errors.Errorf(processName, "provided invalid input image. JBIG2 Encoder requires binary images data")
	}
	b, err := bitmap.NewWithUnpaddedData(enc.Width, enc.Height, data)
	if err != nil {
		return nil, err
	}
	settings := enc.DefaultPageSettings
	if err = settings.Validate(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	switch settings.Compression {
	case JB2Generic:
		if err = enc.d.AddGenericPage(b, settings.DuplicatedLinesRemoval); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
	case JB2SymbolCorrelation:
		return nil, errors.Error(processName, "symbol correlation encoding not implemented yet")
	case JB2SymbolRankHaus:
		return nil, errors.Error(processName, "symbol rank haus encoding not implemented yet")
	default:
		return nil, errors.Error(processName, "provided invalid compression")
	}
	return enc.Encode()
}

// EncodeImage encodes 'img' golang image.Image into jbig2 encoded bytes document using default encoder settings.
func (enc *JBIG2Encoder) EncodeImage(img image.Image) ([]byte, error) {
	return enc.encodeImage(img)
}

// EncodeJBIG2Image encodes 'img' into jbig2 encoded bytes stream, using default encoder settings.
func (enc *JBIG2Encoder) EncodeJBIG2Image(img *JBIG2Image) ([]byte, error) {
	const processName = "core.EncodeJBIG2Image"
	if err := enc.AddPageImage(img, &enc.DefaultPageSettings); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return enc.Encode()
}

// Encode encodes previously prepare jbig2 document and stores it as the byte slice.
func (enc *JBIG2Encoder) Encode() (data []byte, err error) {
	const processName = "JBIG2Document.Encode"
	if enc.d == nil {
		return nil, errors.Errorf(processName, "document input data not defined")
	}
	enc.d.FullHeaders = enc.DefaultPageSettings.FileMode
	// encode the document
	data, err = enc.d.Encode()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return data, nil
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
	dict.Set("Filter", MakeName(enc.GetFilterName()))
	return dict
}

// UpdateParams updates the parameter values of the encoder.
// Implements StreamEncoder interface.
func (enc *JBIG2Encoder) UpdateParams(params *PdfObjectDictionary) {
	bpc, err := GetNumberAsInt64(params.Get("BitsPerComponent"))
	if err == nil {
		enc.BitsPerComponent = int(bpc)
	}
	width, err := GetNumberAsInt64(params.Get("Width"))
	if err == nil {
		enc.Width = int(width)
	}
	height, err := GetNumberAsInt64(params.Get("Height"))
	if err == nil {
		enc.Height = int(height)
	}
	colorComponents, err := GetNumberAsInt64(params.Get("ColorComponents"))
	if err == nil {
		enc.ColorComponents = int(colorComponents)
	}
}

func (enc *JBIG2Encoder) encodeImage(i image.Image) ([]byte, error) {
	const processName = "encodeImage"
	// convert the input into jbig2 image
	jbig2Image, err := GoImageToJBIG2(i, JB2ImageAutoThreshold)
	if err != nil {
		return nil, errors.Wrap(err, processName, "convert input image to jbig2 img")
	}
	if err = enc.AddPageImage(jbig2Image, &enc.DefaultPageSettings); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return enc.Encode()
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
	// if no decode params provided - end fast.
	if decodeParams == nil {
		return encoder, nil
	}
	// set image parameters.
	encoder.UpdateParams(decodeParams)
	globals := decodeParams.Get("JBIG2Globals")
	if globals == nil {
		return encoder, nil
	}
	// decode and set JBIG2 Globals.
	var err error
	globalsStream, ok := globals.(*PdfObjectStream)
	if !ok {
		err = errors.Error(processName, "jbig2.Globals stream should be an Object Stream")
		common.Log.Debug("ERROR: %v", err)
		return nil, err
	}
	encoder.Globals, err = jbig2.DecodeGlobals(globalsStream.Stream)
	if err != nil {
		err = errors.Wrap(err, processName, "corrupted jbig2 encoded data")
		common.Log.Debug("ERROR: %v", err)
		return nil, err
	}
	return encoder, nil
}

//
// JBIG2Image
//

// JBIG2Image is the image structure used by the jbig2 encoder. Its Data must be in a
// 1 bit per component and 1 component per pixel (1bpp). In order to create binary image
// use GoImageToJBIG2 function. If the image data contains the row bytes padding set the HasPadding to true.
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
// If the image is not a black/white image then the function converts provided input into
// JBIG2Image with 1bpp. For non grayscale images the function performs the conversion to the grayscale temp image.
// Then it checks the value of the gray image value if it's within bounds of the black white threshold.
// This 'bwThreshold' value should be in range (0.0, 1.0). The threshold checks if the grayscale pixel (uint) value
// is greater or smaller than 'bwThreshold' * 255. Pixels inside the range will be white, and the others will be black.
// If the 'bwThreshold' is equal to -1.0 - JB2ImageAutoThreshold then it's value would be set on the base of
// it's histogram using Triangle method. For more information go to:
// 	https://www.mathworks.com/matlabcentral/fileexchange/28047-gray-image-thresholding-using-the-triangle-method
func GoImageToJBIG2(i image.Image, bwThreshold float64) (*JBIG2Image, error) {
	const processName = "GoImageToJBIG2"
	if i == nil {
		return nil, errors.Error(processName, "image 'i' not defined")
	}
	var th uint8
	if bwThreshold == JB2ImageAutoThreshold {
		// autoThreshold using triangle method
		gray := imageutil.ImgToGray(i)
		histogram := imageutil.GrayImageHistogram(gray)
		th = imageutil.AutoThresholdTriangle(histogram)
		i = gray
	} else if bwThreshold > 1.0 || bwThreshold < 0.0 {
		// check if bwThreshold is unknown - set to 0.0 is not in the allowed range.
		return nil, errors.Error(processName, "provided threshold is not in a range {0.0, 1.0}")
	} else {
		th = uint8(255 * bwThreshold)
	}
	gray := imageutil.ImgToBinary(i, th)
	return bwToJBIG2Image(gray), nil
}

func bwToJBIG2Image(i *image.Gray) *JBIG2Image {
	bounds := i.Bounds()
	// compute the rowStride - number of bytes in the row.
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

// JBIG2EncoderSettings contains the parameters and settings used by the JBIG2Encoder.
// Current version works only on JB2Generic compression.
type JBIG2EncoderSettings struct {
	// FileMode defines if the jbig2 encoder should return full jbig2 file instead of
	// shortened pdf mode. This adds the file header to the jbig2 definition.
	FileMode bool
	// Compression is the setting that defines the compression type used for encoding the page.
	Compression JBIG2CompressionType
	// DuplicatedLinesRemoval code generic region in a way such that if the lines are duplicated the encoder
	// doesn't store it twice.
	DuplicatedLinesRemoval bool
	// DefaultPixelValue is the bit value initial for every pixel in the page.
	DefaultPixelValue uint8
	// ResolutionX optional setting that defines the 'x' axis input image resolution - used for single page encoding.
	ResolutionX int
	// ResolutionY optional setting that defines the 'y' axis input image resolution - used for single page encoding.
	ResolutionY int
	// Threshold defines the threshold of the image correlation for
	// non Generic compression.
	// User only for JB2SymbolCorrelation and JB2SymbolRankHaus methods.
	// Best results in range [0.7 - 0.98] - the less the better the compression would be
	// but the more lossy.
	// Default value: 0.95
	Threshold float64
}

// Validate validates the page settings for the JBIG2 encoder.
func (s JBIG2EncoderSettings) Validate() error {
	const processName = "validateEncoder"
	if s.Threshold < 0 || s.Threshold > 1.0 {
		return errors.Errorf(processName, "provided threshold value: '%v' must be in range [0.0, 1.0]", s.Threshold)
	}
	if s.ResolutionX < 0 {
		return errors.Errorf(processName, "provided x resolution: '%d' must be positive or zero value", s.ResolutionX)
	}
	if s.ResolutionY < 0 {
		return errors.Errorf(processName, "provided y resolution: '%d' must be positive or zero value", s.ResolutionY)
	}
	if s.DefaultPixelValue != 0 && s.DefaultPixelValue != 1 {
		return errors.Errorf(processName, "default pixel value: '%d' must be a value for the bit: {0,1}", s.DefaultPixelValue)
	}
	if s.Compression != JB2Generic {
		return errors.Errorf(processName, "provided compression is not implemented yet")
	}
	return nil
}
