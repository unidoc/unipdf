/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"bytes"
	"errors"
	"fmt"
	gocolor "image/color"
	"image/jpeg"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// Creates the encoder for the inline image's Filter and DecodeParms.
func newEncoderFromInlineImage(inlineImage *ContentStreamInlineImage) (core.StreamEncoder, error) {
	if inlineImage.Filter == nil {
		// No filter, return raw data back.
		return core.NewRawEncoder(), nil
	}

	// The filter should be a name or an array with a list of filter names.
	filterName, ok := inlineImage.Filter.(*core.PdfObjectName)
	if !ok {
		array, ok := inlineImage.Filter.(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("filter not a Name or Array object")
		}
		if array.Len() == 0 {
			// Empty array -> indicates raw filter (no filter).
			return core.NewRawEncoder(), nil
		}

		if array.Len() != 1 {
			menc, err := newMultiEncoderFromInlineImage(inlineImage)
			if err != nil {
				common.Log.Error("Failed creating multi encoder: %v", err)
				return nil, err
			}

			common.Log.Trace("Multi enc: %s\n", menc)
			return menc, nil
		}

		// Single element.
		filterObj := array.Get(0)
		filterName, ok = filterObj.(*core.PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("filter array member not a Name object")
		}
	}

	// From Table 94 p. 224 (PDF32000_2008):
	// Additional Abbreviations in an Inline Image Object:

	switch *filterName {
	case "AHx", "ASCIIHexDecode":
		return core.NewASCIIHexEncoder(), nil
	case "A85", "ASCII85Decode":
		return core.NewASCII85Encoder(), nil
	case "DCT", "DCTDecode":
		return newDCTEncoderFromInlineImage(inlineImage)
	case "Fl", "FlateDecode":
		return newFlateEncoderFromInlineImage(inlineImage, nil)
	case "LZW", "LZWDecode":
		return newLZWEncoderFromInlineImage(inlineImage, nil)
	case "CCF", "CCITTFaxDecode":
		return core.NewCCITTFaxEncoder(), nil
	case "RL", "RunLengthDecode":
		return core.NewRunLengthEncoder(), nil
	default:
		common.Log.Debug("Unsupported inline image encoding filter name : %s", *filterName)
		return nil, errors.New("unsupported inline encoding method")
	}
}

// Create a new flate decoder from an inline image object, getting all the encoding parameters
// from the DecodeParms stream object dictionary entry that can be provided optionally, usually
// only when a multi filter is used.
func newFlateEncoderFromInlineImage(inlineImage *ContentStreamInlineImage, decodeParams *core.PdfObjectDictionary) (*core.FlateEncoder, error) {
	encoder := core.NewFlateEncoder()

	// If decodeParams not provided, see if we can get from the stream.
	if decodeParams == nil {
		obj := inlineImage.DecodeParms
		if obj != nil {
			dp, isDict := core.GetDict(obj)
			if !isDict {
				common.Log.Debug("Error: DecodeParms not a dictionary (%T)", obj)
				return nil, fmt.Errorf("invalid DecodeParms")
			}
			decodeParams = dp
		}
	}
	if decodeParams == nil {
		// Can safely return here if no decode params, as the following depend on the decode params.
		return encoder, nil
	}

	common.Log.Trace("decode params: %s", decodeParams.String())
	obj := decodeParams.Get("Predictor")
	if obj == nil {
		common.Log.Debug("Error: Predictor missing from DecodeParms - Continue with default (1)")
	} else {
		predictor, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: Predictor specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("invalid Predictor")
		}
		encoder.Predictor = int(*predictor)
	}

	// Bits per component.  Use default if not specified (8).
	obj = decodeParams.Get("BitsPerComponent")
	if obj != nil {
		bpc, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("ERROR: Invalid BitsPerComponent")
			return nil, fmt.Errorf("invalid BitsPerComponent")
		}
		encoder.BitsPerComponent = int(*bpc)
	}

	if encoder.Predictor > 1 {
		// Columns.
		encoder.Columns = 1
		obj = decodeParams.Get("Columns")
		if obj != nil {
			columns, ok := obj.(*core.PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("predictor column invalid")
			}

			encoder.Columns = int(*columns)
		}

		// Colors.
		// Number of interleaved color components per sample (Default 1 if not specified)
		encoder.Colors = 1
		obj := decodeParams.Get("Colors")
		if obj != nil {
			colors, ok := obj.(*core.PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("predictor colors not an integer")
			}
			encoder.Colors = int(*colors)
		}
	}

	return encoder, nil
}

// Create a new LZW encoder/decoder based on an inline image object, getting all the encoding parameters
// from the DecodeParms stream object dictionary entry.
func newLZWEncoderFromInlineImage(inlineImage *ContentStreamInlineImage, decodeParams *core.PdfObjectDictionary) (*core.LZWEncoder, error) {
	// Start with default settings.
	encoder := core.NewLZWEncoder()

	// If decodeParams not provided, see if we can get from the inline image directly.
	if decodeParams == nil {
		if inlineImage.DecodeParms != nil {
			dp, isDict := core.GetDict(inlineImage.DecodeParms)
			if !isDict {
				common.Log.Debug("Error: DecodeParms not a dictionary (%T)", inlineImage.DecodeParms)
				return nil, fmt.Errorf("invalid DecodeParms")
			}
			decodeParams = dp
		}
	}

	if decodeParams == nil {
		// No decode parameters. Can safely return here if not set as the following options
		// are related to the decode Params.
		return encoder, nil
	}

	// The EarlyChange indicates when to increase code length, as different
	// implementations use a different mechanisms. Essentially this chooses
	// which LZW implementation to use.
	// The default is 1 (one code early)
	//
	// The EarlyChange parameter is specified in the object stream dictionary for regular streams,
	// but it is not specified explicitly where to check for it in the case of inline images.
	// We will check in the decodeParms for now, we can adjust later if we come across cases of this.
	obj := decodeParams.Get("EarlyChange")
	if obj != nil {
		earlyChange, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: EarlyChange specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("invalid EarlyChange")
		}
		if *earlyChange != 0 && *earlyChange != 1 {
			return nil, fmt.Errorf("invalid EarlyChange value (not 0 or 1)")
		}

		encoder.EarlyChange = int(*earlyChange)
	} else {
		encoder.EarlyChange = 1 // default
	}

	obj = decodeParams.Get("Predictor")
	if obj != nil {
		predictor, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: Predictor specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("invalid Predictor")
		}
		encoder.Predictor = int(*predictor)
	}

	// Bits per component.  Use default if not specified (8).
	obj = decodeParams.Get("BitsPerComponent")
	if obj != nil {
		bpc, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("ERROR: Invalid BitsPerComponent")
			return nil, fmt.Errorf("invalid BitsPerComponent")
		}
		encoder.BitsPerComponent = int(*bpc)
	}

	if encoder.Predictor > 1 {
		// Columns.
		encoder.Columns = 1
		obj = decodeParams.Get("Columns")
		if obj != nil {
			columns, ok := obj.(*core.PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("predictor column invalid")
			}

			encoder.Columns = int(*columns)
		}

		// Colors.
		// Number of interleaved color components per sample (Default 1 if not specified)
		encoder.Colors = 1
		obj = decodeParams.Get("Colors")
		if obj != nil {
			colors, ok := obj.(*core.PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("predictor colors not an integer")
			}
			encoder.Colors = int(*colors)
		}
	}

	common.Log.Trace("decode params: %s", decodeParams.String())
	return encoder, nil
}

// Create a new DCT encoder/decoder based on an inline image, getting all the encoding parameters
// from the stream object dictionary entry and the image data itself.
func newDCTEncoderFromInlineImage(inlineImage *ContentStreamInlineImage) (*core.DCTEncoder, error) {
	// Start with default settings.
	encoder := core.NewDCTEncoder()

	bufReader := bytes.NewReader(inlineImage.stream)

	cfg, err := jpeg.DecodeConfig(bufReader)
	//img, _, err := goimage.Decode(bufReader)
	if err != nil {
		common.Log.Debug("Error decoding file: %s", err)
		return nil, err
	}

	switch cfg.ColorModel {
	case gocolor.RGBAModel:
		encoder.BitsPerComponent = 8
		encoder.ColorComponents = 3 // alpha is not included in pdf.
	case gocolor.RGBA64Model:
		encoder.BitsPerComponent = 16
		encoder.ColorComponents = 3
	case gocolor.GrayModel:
		encoder.BitsPerComponent = 8
		encoder.ColorComponents = 1
	case gocolor.Gray16Model:
		encoder.BitsPerComponent = 16
		encoder.ColorComponents = 1
	case gocolor.CMYKModel:
		encoder.BitsPerComponent = 8
		encoder.ColorComponents = 4
	case gocolor.YCbCrModel:
		// YCbCr is not supported by PDF, but it could be a different colorspace
		// with 3 components.  Would be specified by the ColorSpace entry.
		encoder.BitsPerComponent = 8
		encoder.ColorComponents = 3
	default:
		return nil, errors.New("unsupported color model")
	}
	encoder.Width = cfg.Width
	encoder.Height = cfg.Height
	common.Log.Trace("DCT Encoder: %+v", encoder)

	return encoder, nil
}

// Create a new multi-filter encoder/decoder based on an inline image, getting all the encoding parameters
// from the filter specification and the DecodeParms (DP) dictionaries.
func newMultiEncoderFromInlineImage(inlineImage *ContentStreamInlineImage) (*core.MultiEncoder, error) {
	mencoder := core.NewMultiEncoder()

	// Prepare the decode params array (one for each filter type)
	// Optional, not always present.
	var decodeParamsDict *core.PdfObjectDictionary
	var decodeParamsArray []core.PdfObject
	if obj := inlineImage.DecodeParms; obj != nil {
		// If it is a dictionary, assume it applies to all
		dict, isDict := obj.(*core.PdfObjectDictionary)
		if isDict {
			decodeParamsDict = dict
		}

		// If it is an array, assume there is one for each
		arr, isArray := obj.(*core.PdfObjectArray)
		if isArray {
			for _, dictObj := range arr.Elements() {
				if dict, is := dictObj.(*core.PdfObjectDictionary); is {
					decodeParamsArray = append(decodeParamsArray, dict)
				} else {
					decodeParamsArray = append(decodeParamsArray, nil)
				}
			}
		}
	}

	obj := inlineImage.Filter
	if obj == nil {
		return nil, fmt.Errorf("filter missing")
	}

	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("multi filter can only be made from array")
	}

	for idx, obj := range array.Elements() {
		name, ok := obj.(*core.PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("multi filter array element not a name")
		}

		var dp core.PdfObject

		// If decode params dict is set, use it.  Otherwise take from array..
		if decodeParamsDict != nil {
			dp = decodeParamsDict
		} else {
			// Only get the dp if provided.  Oftentimes there is no decode params dict
			// provided.
			if len(decodeParamsArray) > 0 {
				if idx >= len(decodeParamsArray) {
					return nil, fmt.Errorf("missing elements in decode params array")
				}
				dp = decodeParamsArray[idx]
			}
		}

		var dParams *core.PdfObjectDictionary
		if dict, is := dp.(*core.PdfObjectDictionary); is {
			dParams = dict
		}

		if *name == core.StreamEncodingFilterNameFlate || *name == "Fl" {
			// TODO: need to separate out the DecodeParms..
			encoder, err := newFlateEncoderFromInlineImage(inlineImage, dParams)
			if err != nil {
				return nil, err
			}
			mencoder.AddEncoder(encoder)
		} else if *name == core.StreamEncodingFilterNameLZW {
			encoder, err := newLZWEncoderFromInlineImage(inlineImage, dParams)
			if err != nil {
				return nil, err
			}
			mencoder.AddEncoder(encoder)
		} else if *name == core.StreamEncodingFilterNameASCIIHex {
			encoder := core.NewASCIIHexEncoder()
			mencoder.AddEncoder(encoder)
		} else if *name == core.StreamEncodingFilterNameASCII85 || *name == "A85" {
			encoder := core.NewASCII85Encoder()
			mencoder.AddEncoder(encoder)
		} else {
			common.Log.Error("Unsupported filter %s", *name)
			return nil, fmt.Errorf("invalid filter in multi filter array")
		}
	}

	return mencoder, nil
}
