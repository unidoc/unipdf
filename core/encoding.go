/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

// Implement encoders for PDF. Currently supported:
// - Raw (Identity)
// - FlateDecode
// - LZW
// - DCT Decode (JPEG)
// - RunLength
// - ASCII Hex
// - ASCII85
// - CCITT Fax (dummy)
// - JBIG2 (dummy)
// - JPX (dummy)

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"fmt"
	goimage "image"
	gocolor "image/color"
	"image/jpeg"
	"io"

	// Need two slightly different implementations of LZW (EarlyChange parameter).
	lzw0 "compress/lzw"

	lzw1 "golang.org/x/image/tiff/lzw"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/ccittfax"
)

// Stream encoding filter names.
const (
	StreamEncodingFilterNameFlate     = "FlateDecode"
	StreamEncodingFilterNameLZW       = "LZWDecode"
	StreamEncodingFilterNameDCT       = "DCTDecode"
	StreamEncodingFilterNameRunLength = "RunLengthDecode"
	StreamEncodingFilterNameASCIIHex  = "ASCIIHexDecode"
	StreamEncodingFilterNameASCII85   = "ASCII85Decode"
	StreamEncodingFilterNameCCITTFax  = "CCITTFaxDecode"
	StreamEncodingFilterNameJBIG2     = "JBIG2Decode"
	StreamEncodingFilterNameJPX       = "JPXDecode"
	StreamEncodingFilterNameRaw       = "Raw"
)

const (
	// DefaultJPEGQuality is the default quality produced by JPEG encoders.
	DefaultJPEGQuality = 75
)

// StreamEncoder represents the interface for all PDF stream encoders.
type StreamEncoder interface {
	GetFilterName() string
	MakeDecodeParams() PdfObject
	MakeStreamDict() *PdfObjectDictionary
	UpdateParams(params *PdfObjectDictionary)

	EncodeBytes(data []byte) ([]byte, error)
	DecodeBytes(encoded []byte) ([]byte, error)
	DecodeStream(streamObj *PdfObjectStream) ([]byte, error)
}

// FlateEncoder represents Flate encoding.
type FlateEncoder struct {
	Predictor        int
	BitsPerComponent int
	// For predictors
	Columns int
	Colors  int
}

// NewFlateEncoder makes a new flate encoder with default parameters, predictor 1 and bits per component 8.
func NewFlateEncoder() *FlateEncoder {
	encoder := &FlateEncoder{}

	// Default (No prediction)
	encoder.Predictor = 1

	// Currently only supporting 8.
	encoder.BitsPerComponent = 8

	encoder.Colors = 1
	encoder.Columns = 1

	return encoder
}

// SetPredictor sets the predictor function.  Specify the number of columns per row.
// The columns indicates the number of samples per row.
// Used for grouping data together for compression.
func (enc *FlateEncoder) SetPredictor(columns int) {
	// Only supporting PNG sub predictor for encoding.
	enc.Predictor = 11
	enc.Columns = columns
}

// GetFilterName returns the name of the encoding filter.
func (enc *FlateEncoder) GetFilterName() string {
	return StreamEncodingFilterNameFlate
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *FlateEncoder) MakeDecodeParams() PdfObject {
	if enc.Predictor > 1 {
		decodeParams := MakeDict()
		decodeParams.Set("Predictor", MakeInteger(int64(enc.Predictor)))

		// Only add if not default option.
		if enc.BitsPerComponent != 8 {
			decodeParams.Set("BitsPerComponent", MakeInteger(int64(enc.BitsPerComponent)))
		}
		if enc.Columns != 1 {
			decodeParams.Set("Columns", MakeInteger(int64(enc.Columns)))
		}
		if enc.Colors != 1 {
			decodeParams.Set("Colors", MakeInteger(int64(enc.Colors)))
		}
		return decodeParams
	}

	return nil
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
// Has the Filter set and the DecodeParms.
func (enc *FlateEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()
	dict.Set("Filter", MakeName(enc.GetFilterName()))

	decodeParams := enc.MakeDecodeParams()
	if decodeParams != nil {
		dict.Set("DecodeParms", decodeParams)
	}

	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *FlateEncoder) UpdateParams(params *PdfObjectDictionary) {
	predictor, err := GetNumberAsInt64(params.Get("Predictor"))
	if err == nil {
		enc.Predictor = int(predictor)
	}

	bpc, err := GetNumberAsInt64(params.Get("BitsPerComponent"))
	if err == nil {
		enc.BitsPerComponent = int(bpc)
	}

	columns, err := GetNumberAsInt64(params.Get("Width"))
	if err == nil {
		enc.Columns = int(columns)
	}

	colorComponents, err := GetNumberAsInt64(params.Get("ColorComponents"))
	if err == nil {
		enc.Colors = int(colorComponents)
	}
}

// Create a new flate decoder from a stream object, getting all the encoding parameters
// from the DecodeParms stream object dictionary entry.
func newFlateEncoderFromStream(streamObj *PdfObjectStream, decodeParams *PdfObjectDictionary) (*FlateEncoder, error) {
	encoder := NewFlateEncoder()

	encDict := streamObj.PdfObjectDictionary
	if encDict == nil {
		// No encoding dictionary.
		return encoder, nil
	}

	// If decodeParams not provided, see if we can get from the stream.
	if decodeParams == nil {
		obj := TraceToDirectObject(encDict.Get("DecodeParms"))
		switch t := obj.(type) {
		case *PdfObjectArray:
			arr := t
			if arr.Len() != 1 {
				common.Log.Debug("Error: DecodeParms array length != 1 (%d)", arr.Len())
				return nil, errors.New("range check error")
			}
			obj = TraceToDirectObject(arr.Get(0))
		case *PdfObjectDictionary:
			decodeParams = t
		case *PdfObjectNull, nil:
			// No decode params set.
		default:
			common.Log.Debug("Error: DecodeParms not a dictionary (%T)", obj)
			return nil, fmt.Errorf("invalid DecodeParms")
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
		predictor, ok := obj.(*PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: Predictor specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("invalid Predictor")
		}
		encoder.Predictor = int(*predictor)
	}

	// Bits per component.  Use default if not specified (8).
	obj = decodeParams.Get("BitsPerComponent")
	if obj != nil {
		bpc, ok := obj.(*PdfObjectInteger)
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
			columns, ok := obj.(*PdfObjectInteger)
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
			colors, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("predictor colors not an integer")
			}
			encoder.Colors = int(*colors)
		}
	}

	return encoder, nil
}

// DecodeBytes decodes a slice of Flate encoded bytes and returns the result.
func (enc *FlateEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	common.Log.Trace("FlateDecode bytes")
	if len(encoded) == 0 {
		common.Log.Debug("ERROR: empty Flate encoded buffer. Returning empty byte slice.")
		return []byte{}, nil
	}

	bufReader := bytes.NewReader(encoded)
	r, err := zlib.NewReader(bufReader)
	if err != nil {
		common.Log.Debug("Decoding error %v\n", err)
		common.Log.Debug("Stream (%d) % x", len(encoded), encoded)
		return nil, err
	}
	defer r.Close()

	var outBuf bytes.Buffer
	outBuf.ReadFrom(r)

	return outBuf.Bytes(), nil
}

// Prediction filters for PNG predictors.
const (
	pfNone  = 0 // No prediction (raw).
	pfSub   = 1 // Predicts same as left sample.
	pfUp    = 2 // Predicts same as sample above.
	pfAvg   = 3 // Predict based on left and above.
	pfPaeth = 4 // Paeth algorithm prediction.
)

// Apply predictor to decoded `outData` to get final output data.
func (enc *FlateEncoder) postDecodePredict(outData []byte) ([]byte, error) {
	if enc.Predictor > 1 {
		if enc.Predictor == 2 { // TIFF encoding.
			// TODO(gunnsth): Needs test cases.
			common.Log.Trace("Tiff encoding")
			common.Log.Trace("Colors: %d", enc.Colors)

			rowLength := int(enc.Columns) * enc.Colors
			if rowLength < 1 {
				// No data. Return empty set.
				return []byte{}, nil
			}
			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				common.Log.Debug("ERROR: TIFF encoding: Invalid row length...")
				return nil, fmt.Errorf("invalid row length (%d/%d)", len(outData), rowLength)
			}
			if rowLength%enc.Colors != 0 {
				return nil, fmt.Errorf("invalid row length (%d) for colors %d", rowLength, enc.Colors)
			}
			if rowLength > len(outData) {
				common.Log.Debug("Row length cannot be longer than data length (%d/%d)", rowLength, len(outData))
				return nil, errors.New("range check error")
			}
			common.Log.Trace("inp outData (%d): % x", len(outData), outData)

			pOutBuffer := bytes.NewBuffer(nil)

			// 0-255  -255 255 ; 0-255=-255;
			for i := 0; i < rows; i++ {
				rowData := outData[rowLength*i : rowLength*(i+1)]
				// Predicts the same as the sample to the left.
				// Interleaved by colors.
				for j := enc.Colors; j < rowLength; j++ {
					rowData[j] += rowData[j-enc.Colors]
				}
				pOutBuffer.Write(rowData)
			}
			pOutData := pOutBuffer.Bytes()
			common.Log.Trace("POutData (%d): % x", len(pOutData), pOutData)
			return pOutData, nil
		} else if enc.Predictor >= 10 && enc.Predictor <= 15 {
			common.Log.Trace("PNG Encoding")
			// Columns represents the number of samples per row; Each sample can contain multiple color
			// components.
			rowLength := int(enc.Columns*enc.Colors + 1) // 1 byte to specify predictor algorithms per row.
			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				return nil, fmt.Errorf("invalid row length (%d/%d)", len(outData), rowLength)
			}
			if rowLength > len(outData) {
				common.Log.Debug("Row length cannot be longer than data length (%d/%d)", rowLength, len(outData))
				return nil, errors.New("range check error")
			}

			pOutBuffer := bytes.NewBuffer(nil)

			common.Log.Trace("Predictor columns: %d", enc.Columns)
			common.Log.Trace("Length: %d / %d = %d rows", len(outData), rowLength, rows)
			prevRowData := make([]byte, rowLength)
			for i := 0; i < rowLength; i++ {
				prevRowData[i] = 0
			}

			bytesPerPixel := enc.Colors // Assuming BPC = 8.

			for i := 0; i < rows; i++ {
				rowData := outData[rowLength*i : rowLength*(i+1)]

				fb := rowData[0]
				switch fb {
				case pfNone:
				case pfSub:
					for j := 1 + bytesPerPixel; j < rowLength; j++ {
						rowData[j] += rowData[j-bytesPerPixel]
					}
				case pfUp:
					// Up: Predicts the same as the sample above
					for j := 1; j < rowLength; j++ {
						rowData[j] += prevRowData[j]
					}
				case pfAvg:
					// Avg: Predicts the same as the average of the sample to the left and above.
					for j := 1; j < bytesPerPixel+1; j++ {
						rowData[j] += prevRowData[j] / 2
					}
					for j := bytesPerPixel + 1; j < rowLength; j++ {
						rowData[j] += byte((int(rowData[j-bytesPerPixel]) + int(prevRowData[j])) / 2)
					}
				case pfPaeth:
					// Paeth: a nonlinear function of the sample to the left (a), sample above (b)
					// and the upper left (c).
					for j := 1; j < rowLength; j++ {
						var a, b, c byte
						b = prevRowData[j] // above.
						if j >= bytesPerPixel+1 {
							a = rowData[j-bytesPerPixel]
							c = prevRowData[j-bytesPerPixel]
						}
						rowData[j] += paeth(a, b, c)
					}
				default:
					common.Log.Debug("ERROR: Invalid filter byte (%d) @row %d", fb, i)
					return nil, fmt.Errorf("invalid filter byte (%d)", fb)
				}

				copy(prevRowData, rowData)
				pOutBuffer.Write(rowData[1:])
			}
			pOutData := pOutBuffer.Bytes()
			return pOutData, nil
		} else {
			common.Log.Debug("ERROR: Unsupported predictor (%d)", enc.Predictor)
			return nil, fmt.Errorf("unsupported predictor (%d)", enc.Predictor)
		}
	}

	return outData, nil
}

// DecodeStream decodes a FlateEncoded stream object and give back decoded bytes.
func (enc *FlateEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	// TODO: Handle more filter bytes and support more values of BitsPerComponent.

	common.Log.Trace("FlateDecode stream")
	common.Log.Trace("Predictor: %d", enc.Predictor)
	if enc.BitsPerComponent != 8 {
		return nil, fmt.Errorf("invalid BitsPerComponent=%d (only 8 supported)", enc.BitsPerComponent)
	}

	outData, err := enc.DecodeBytes(streamObj.Stream)
	if err != nil {
		return nil, err
	}

	return enc.postDecodePredict(outData)
}

// EncodeBytes encodes a bytes array and return the encoded value based on the encoder parameters.
func (enc *FlateEncoder) EncodeBytes(data []byte) ([]byte, error) {
	if enc.Predictor != 1 && enc.Predictor != 11 {
		common.Log.Debug("Encoding error: FlateEncoder Predictor = 1, 11 only supported")
		return nil, ErrUnsupportedEncodingParameters
	}

	if enc.Predictor == 11 {
		// The length of each output row in number of samples.
		// N.B. Each output row has one extra sample as compared to the input to indicate the
		// predictor type.
		rowLength := int(enc.Columns)
		rows := len(data) / rowLength
		if len(data)%rowLength != 0 {
			common.Log.Error("Invalid row length")
			return nil, errors.New("invalid row length")
		}

		pOutBuffer := bytes.NewBuffer(nil)

		tmpData := make([]byte, rowLength)

		for i := 0; i < rows; i++ {
			rowData := data[rowLength*i : rowLength*(i+1)]

			// PNG SUB method.
			// Sub: Predicts the same as the sample to the left.
			tmpData[0] = rowData[0]
			for j := 1; j < rowLength; j++ {
				tmpData[j] = byte(int(rowData[j]-rowData[j-1]) % 256)
			}

			pOutBuffer.WriteByte(1) // sub method
			pOutBuffer.Write(tmpData)
		}

		data = pOutBuffer.Bytes()
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()

	return b.Bytes(), nil
}

// LZWEncoder provides LZW encoding/decoding functionality.
type LZWEncoder struct {
	Predictor        int
	BitsPerComponent int
	// For predictors
	Columns int
	Colors  int
	// LZW algorithm setting.
	EarlyChange int
}

// NewLZWEncoder makes a new LZW encoder with default parameters.
func NewLZWEncoder() *LZWEncoder {
	encoder := &LZWEncoder{}

	// Default (No prediction)
	encoder.Predictor = 1

	// Currently only supporting 8.
	encoder.BitsPerComponent = 8

	encoder.Colors = 1
	encoder.Columns = 1
	encoder.EarlyChange = 1

	return encoder
}

// GetFilterName returns the name of the encoding filter.
func (enc *LZWEncoder) GetFilterName() string {
	return StreamEncodingFilterNameLZW
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *LZWEncoder) MakeDecodeParams() PdfObject {
	if enc.Predictor > 1 {
		decodeParams := MakeDict()
		decodeParams.Set("Predictor", MakeInteger(int64(enc.Predictor)))

		// Only add if not default option.
		if enc.BitsPerComponent != 8 {
			decodeParams.Set("BitsPerComponent", MakeInteger(int64(enc.BitsPerComponent)))
		}
		if enc.Columns != 1 {
			decodeParams.Set("Columns", MakeInteger(int64(enc.Columns)))
		}
		if enc.Colors != 1 {
			decodeParams.Set("Colors", MakeInteger(int64(enc.Colors)))
		}
		return decodeParams
	}
	return nil
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
// Has the Filter set and the DecodeParms.
func (enc *LZWEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()

	dict.Set("Filter", MakeName(enc.GetFilterName()))

	decodeParams := enc.MakeDecodeParams()
	if decodeParams != nil {
		dict.Set("DecodeParms", decodeParams)
	}

	dict.Set("EarlyChange", MakeInteger(int64(enc.EarlyChange)))

	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *LZWEncoder) UpdateParams(params *PdfObjectDictionary) {
	predictor, err := GetNumberAsInt64(params.Get("Predictor"))
	if err == nil {
		enc.Predictor = int(predictor)
	}

	bpc, err := GetNumberAsInt64(params.Get("BitsPerComponent"))
	if err == nil {
		enc.BitsPerComponent = int(bpc)
	}

	columns, err := GetNumberAsInt64(params.Get("Width"))
	if err == nil {
		enc.Columns = int(columns)
	}

	colorComponents, err := GetNumberAsInt64(params.Get("ColorComponents"))
	if err == nil {
		enc.Colors = int(colorComponents)
	}

	earlyChange, err := GetNumberAsInt64(params.Get("EarlyChange"))
	if err == nil {
		enc.EarlyChange = int(earlyChange)
	}
}

// Create a new LZW encoder/decoder from a stream object, getting all the encoding parameters
// from the DecodeParms stream object dictionary entry.
func newLZWEncoderFromStream(streamObj *PdfObjectStream, decodeParams *PdfObjectDictionary) (*LZWEncoder, error) {
	// Start with default settings.
	encoder := NewLZWEncoder()

	encDict := streamObj.PdfObjectDictionary
	if encDict == nil {
		// No encoding dictionary.
		return encoder, nil
	}

	// If decodeParams not provided, see if we can get from the stream.
	if decodeParams == nil {
		obj := TraceToDirectObject(encDict.Get("DecodeParms"))
		if obj != nil {
			if dp, isDict := obj.(*PdfObjectDictionary); isDict {
				decodeParams = dp
			} else if a, isArr := obj.(*PdfObjectArray); isArr {
				if a.Len() == 1 {
					if dp, isDict := GetDict(a.Get(0)); isDict {
						decodeParams = dp
					}
				}
			}
			if decodeParams == nil {
				common.Log.Error("DecodeParms not a dictionary %#v", obj)
				return nil, fmt.Errorf("invalid DecodeParms")
			}
		}
	}

	// The EarlyChange indicates when to increase code length, as different
	// implementations use a different mechanisms. Essentially this chooses
	// which LZW implementation to use.
	// The default is 1 (one code early)
	obj := encDict.Get("EarlyChange")
	if obj != nil {
		earlyChange, ok := obj.(*PdfObjectInteger)
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

	if decodeParams == nil {
		// No decode parameters. Can safely return here if not set as the following options
		// are related to the decode Params.
		return encoder, nil
	}

	obj = decodeParams.Get("Predictor")
	if obj != nil {
		predictor, ok := obj.(*PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: Predictor specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("invalid Predictor")
		}
		encoder.Predictor = int(*predictor)
	}

	// Bits per component.  Use default if not specified (8).
	obj = decodeParams.Get("BitsPerComponent")
	if obj != nil {
		bpc, ok := obj.(*PdfObjectInteger)
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
			columns, ok := obj.(*PdfObjectInteger)
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
			colors, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("predictor colors not an integer")
			}
			encoder.Colors = int(*colors)
		}
	}

	common.Log.Trace("decode params: %s", decodeParams.String())
	return encoder, nil
}

// DecodeBytes decodes a slice of LZW encoded bytes and returns the result.
func (enc *LZWEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	var outBuf bytes.Buffer
	bufReader := bytes.NewReader(encoded)

	var r io.ReadCloser
	if enc.EarlyChange == 1 {
		// LZW implementation with code length increases one code early (1).
		r = lzw1.NewReader(bufReader, lzw1.MSB, 8)
	} else {
		// 0: LZW implementation with postponed code length increases (0).
		r = lzw0.NewReader(bufReader, lzw0.MSB, 8)
	}
	defer r.Close()

	_, err := outBuf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return outBuf.Bytes(), nil
}

// DecodeStream decodes a LZW encoded stream and returns the result as a
// slice of bytes.
func (enc *LZWEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	// Revamp this support to handle TIFF predictor (2).
	// Also handle more filter bytes and check
	// BitsPerComponent.  Default value is 8, currently we are only
	// supporting that one.

	common.Log.Trace("LZW Decoding")
	common.Log.Trace("Predictor: %d", enc.Predictor)

	outData, err := enc.DecodeBytes(streamObj.Stream)
	if err != nil {
		return nil, err
	}

	common.Log.Trace(" IN: (%d) % x", len(streamObj.Stream), streamObj.Stream)
	common.Log.Trace("OUT: (%d) % x", len(outData), outData)

	if enc.Predictor > 1 {
		if enc.Predictor == 2 { // TIFF encoding: Needs some tests.
			common.Log.Trace("Tiff encoding")

			rowLength := int(enc.Columns) * enc.Colors
			if rowLength < 1 {
				// No data. Return empty set.
				return []byte{}, nil
			}

			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				common.Log.Debug("ERROR: TIFF encoding: Invalid row length...")
				return nil, fmt.Errorf("invalid row length (%d/%d)", len(outData), rowLength)
			}

			if rowLength%enc.Colors != 0 {
				return nil, fmt.Errorf("invalid row length (%d) for colors %d", rowLength, enc.Colors)
			}

			if rowLength > len(outData) {
				common.Log.Debug("Row length cannot be longer than data length (%d/%d)", rowLength, len(outData))
				return nil, errors.New("range check error")
			}
			common.Log.Trace("inp outData (%d): % x", len(outData), outData)

			pOutBuffer := bytes.NewBuffer(nil)

			// 0-255  -255 255 ; 0-255=-255;
			for i := 0; i < rows; i++ {
				rowData := outData[rowLength*i : rowLength*(i+1)]
				// Predicts the same as the sample to the left.
				// Interleaved by colors.
				for j := enc.Colors; j < rowLength; j++ {
					rowData[j] = byte(int(rowData[j]+rowData[j-enc.Colors]) % 256)
				}
				// GH: Appears that this is not working as expected...

				pOutBuffer.Write(rowData)
			}
			pOutData := pOutBuffer.Bytes()
			common.Log.Trace("POutData (%d): % x", len(pOutData), pOutData)
			return pOutData, nil
		} else if enc.Predictor >= 10 && enc.Predictor <= 15 {
			common.Log.Trace("PNG Encoding")
			// Columns represents the number of samples per row; Each sample can contain multiple color
			// components.
			rowLength := int(enc.Columns*enc.Colors + 1) // 1 byte to specify predictor algorithms per row.
			if rowLength < 1 {
				// No data. Return empty set.
				return []byte{}, nil
			}
			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				return nil, fmt.Errorf("invalid row length (%d/%d)", len(outData), rowLength)
			}
			if rowLength > len(outData) {
				common.Log.Debug("Row length cannot be longer than data length (%d/%d)", rowLength, len(outData))
				return nil, errors.New("range check error")
			}

			pOutBuffer := bytes.NewBuffer(nil)

			common.Log.Trace("Predictor columns: %d", enc.Columns)
			common.Log.Trace("Length: %d / %d = %d rows", len(outData), rowLength, rows)
			prevRowData := make([]byte, rowLength)
			for i := 0; i < rowLength; i++ {
				prevRowData[i] = 0
			}

			for i := 0; i < rows; i++ {
				rowData := outData[rowLength*i : rowLength*(i+1)]

				fb := rowData[0]
				switch fb {
				case 0:
					// No prediction. (No operation).
				case 1:
					// Sub: Predicts the same as the sample to the left.
					for j := 2; j < rowLength; j++ {
						rowData[j] = byte(int(rowData[j]+rowData[j-1]) % 256)
					}
				case 2:
					// Up: Predicts the same as the sample above
					for j := 1; j < rowLength; j++ {
						rowData[j] = byte(int(rowData[j]+prevRowData[j]) % 256)
					}
				default:
					common.Log.Debug("ERROR: Invalid filter byte (%d)", fb)
					return nil, fmt.Errorf("invalid filter byte (%d)", fb)
				}

				for i := 0; i < rowLength; i++ {
					prevRowData[i] = rowData[i]
				}
				pOutBuffer.Write(rowData[1:])
			}
			pOutData := pOutBuffer.Bytes()
			return pOutData, nil
		} else {
			common.Log.Debug("ERROR: Unsupported predictor (%d)", enc.Predictor)
			return nil, fmt.Errorf("unsupported predictor (%d)", enc.Predictor)
		}
	}

	return outData, nil
}

// EncodeBytes implements support for LZW encoding.  Currently not supporting predictors (raw compressed data only).
// Only supports the Early change = 1 algorithm (compress/lzw) as the other implementation
// does not have a write method.
// TODO: Consider refactoring compress/lzw to allow both.
func (enc *LZWEncoder) EncodeBytes(data []byte) ([]byte, error) {
	if enc.Predictor != 1 {
		return nil, fmt.Errorf("LZW Predictor = 1 only supported yet")
	}

	if enc.EarlyChange == 1 {
		return nil, fmt.Errorf("LZW Early Change = 0 only supported yet")
	}

	var b bytes.Buffer
	w := lzw0.NewWriter(&b, lzw0.MSB, 8)
	w.Write(data)
	w.Close()

	return b.Bytes(), nil
}

// DCTEncoder provides a DCT (JPG) encoding/decoding functionality for images.
type DCTEncoder struct {
	ColorComponents  int // 1 (gray), 3 (rgb), 4 (cmyk)
	BitsPerComponent int // 8 or 16 bit
	Width            int
	Height           int
	Quality          int
}

// NewDCTEncoder makes a new DCT encoder with default parameters.
func NewDCTEncoder() *DCTEncoder {
	encoder := &DCTEncoder{}

	encoder.ColorComponents = 3
	encoder.BitsPerComponent = 8

	encoder.Quality = DefaultJPEGQuality

	return encoder
}

// GetFilterName returns the name of the encoding filter.
func (enc *DCTEncoder) GetFilterName() string {
	return StreamEncodingFilterNameDCT
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *DCTEncoder) MakeDecodeParams() PdfObject {
	// Does not have decode params.
	return nil
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
// Has the Filter set.  Some other parameters are generated elsewhere.
func (enc *DCTEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()

	dict.Set("Filter", MakeName(enc.GetFilterName()))

	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *DCTEncoder) UpdateParams(params *PdfObjectDictionary) {
	colorComponents, err := GetNumberAsInt64(params.Get("ColorComponents"))
	if err == nil {
		enc.ColorComponents = int(colorComponents)
	}

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

	quality, err := GetNumberAsInt64(params.Get("Quality"))
	if err == nil {
		enc.Quality = int(quality)
	}
}

// Create a new DCT encoder/decoder from a stream object, getting all the encoding parameters
// from the stream object dictionary entry and the image data itself.
// TODO: Support if used with other filters [ASCII85Decode FlateDecode DCTDecode]...
// need to apply the other filters prior to this one...
func newDCTEncoderFromStream(streamObj *PdfObjectStream, multiEnc *MultiEncoder) (*DCTEncoder, error) {
	// Start with default settings.
	encoder := NewDCTEncoder()

	encDict := streamObj.PdfObjectDictionary
	if encDict == nil {
		// No encoding dictionary.
		return encoder, nil
	}

	// If using DCTDecode in combination with other filters, make sure to decode that first...
	encoded := streamObj.Stream
	if multiEnc != nil {
		e, err := multiEnc.DecodeBytes(encoded)
		if err != nil {
			return nil, err
		}
		encoded = e
	}

	bufReader := bytes.NewReader(encoded)

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
	encoder.Quality = DefaultJPEGQuality

	return encoder, nil
}

// DecodeBytes decodes a slice of DCT encoded bytes and returns the result.
func (enc *DCTEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	bufReader := bytes.NewReader(encoded)
	//img, _, err := goimage.Decode(bufReader)
	img, err := jpeg.Decode(bufReader)
	if err != nil {
		common.Log.Debug("Error decoding image: %s", err)
		return nil, err
	}
	bounds := img.Bounds()

	var decoded = make([]byte, bounds.Dx()*bounds.Dy()*enc.ColorComponents*enc.BitsPerComponent/8)
	index := 0

	for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			color := img.At(i, j)

			// Gray scale.
			if enc.ColorComponents == 1 {
				if enc.BitsPerComponent == 16 {
					// Gray - 16 bit.
					val, ok := color.(gocolor.Gray16)
					if !ok {
						return nil, errors.New("color type error")
					}
					decoded[index] = byte((val.Y >> 8) & 0xff)
					index++
					decoded[index] = byte(val.Y & 0xff)
					index++
				} else {
					// Gray - 8 bit.
					val, ok := color.(gocolor.Gray)
					if !ok {
						return nil, errors.New("color type error")
					}
					decoded[index] = byte(val.Y & 0xff)
					index++
				}
			} else if enc.ColorComponents == 3 {
				if enc.BitsPerComponent == 16 {
					val, ok := color.(gocolor.RGBA64)
					if !ok {
						return nil, errors.New("color type error")
					}
					decoded[index] = byte((val.R >> 8) & 0xff)
					index++
					decoded[index] = byte(val.R & 0xff)
					index++
					decoded[index] = byte((val.G >> 8) & 0xff)
					index++
					decoded[index] = byte(val.G & 0xff)
					index++
					decoded[index] = byte((val.B >> 8) & 0xff)
					index++
					decoded[index] = byte(val.B & 0xff)
					index++
				} else {
					// RGB - 8 bit.
					val, isRGB := color.(gocolor.RGBA)
					if isRGB {
						decoded[index] = val.R & 0xff
						index++
						decoded[index] = val.G & 0xff
						index++
						decoded[index] = val.B & 0xff
						index++
					} else {
						// Hack around YCbCr from go jpeg package.
						val, ok := color.(gocolor.YCbCr)
						if !ok {
							return nil, errors.New("color type error")
						}
						r, g, b, _ := val.RGBA()
						// The fact that we cannot use the Y, Cb, Cr values directly,
						// indicates that either the jpeg package is converting the raw
						// data into YCbCr with some kind of mapping, or that the original
						// data is not in R,G,B...
						// TODO: This is not good as it means we end up with R, G, B... even
						// if the original colormap was different.  Unless calling the RGBA()
						// call exactly reverses the previous conversion to YCbCr (even if
						// real data is not rgb)... ?
						// TODO: Test more. Consider whether we need to implement our own jpeg filter.
						decoded[index] = byte(r >> 8) //byte(val.Y & 0xff)
						index++
						decoded[index] = byte(g >> 8) //val.Cb & 0xff)
						index++
						decoded[index] = byte(b >> 8) //val.Cr & 0xff)
						index++
					}
				}
			} else if enc.ColorComponents == 4 {
				// CMYK - 8 bit.
				val, ok := color.(gocolor.CMYK)
				if !ok {
					return nil, errors.New("color type error")
				}
				// TODO: Is the inversion not handled right in the JPEG package for APP14?
				// Should not need to invert here...
				decoded[index] = 255 - val.C&0xff
				index++
				decoded[index] = 255 - val.M&0xff
				index++
				decoded[index] = 255 - val.Y&0xff
				index++
				decoded[index] = 255 - val.K&0xff
				index++
			}
		}
	}

	return decoded, nil
}

// DecodeStream decodes a DCT encoded stream and returns the result as a
// slice of bytes.
func (enc *DCTEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return enc.DecodeBytes(streamObj.Stream)
}

// DrawableImage is same as golang image/draw's Image interface that allow drawing images.
type DrawableImage interface {
	ColorModel() gocolor.Model
	Bounds() goimage.Rectangle
	At(x, y int) gocolor.Color
	Set(x, y int, c gocolor.Color)
}

// EncodeBytes DCT encodes the passed in slice of bytes.
func (enc *DCTEncoder) EncodeBytes(data []byte) ([]byte, error) {
	bounds := goimage.Rect(0, 0, enc.Width, enc.Height)
	var img DrawableImage
	if enc.ColorComponents == 1 {
		if enc.BitsPerComponent == 16 {
			img = goimage.NewGray16(bounds)
		} else {
			img = goimage.NewGray(bounds)
		}
	} else if enc.ColorComponents == 3 {
		if enc.BitsPerComponent == 16 {
			img = goimage.NewRGBA64(bounds)
		} else {
			img = goimage.NewRGBA(bounds)
		}
	} else if enc.ColorComponents == 4 {
		img = goimage.NewCMYK(bounds)
	} else {
		return nil, errors.New("unsupported")
	}

	// Draw the data on the image..
	if enc.BitsPerComponent < 8 {
		enc.BitsPerComponent = 8
	}

	bytesPerColor := enc.ColorComponents * enc.BitsPerComponent / 8
	if bytesPerColor < 1 {
		bytesPerColor = 1
	}

	x := 0
	y := 0
	for i := 0; i+bytesPerColor-1 < len(data); i += bytesPerColor {
		var c gocolor.Color
		if enc.ColorComponents == 1 {
			if enc.BitsPerComponent == 16 {
				val := uint16(data[i])<<8 | uint16(data[i+1])
				c = gocolor.Gray16{val}
			} else {
				val := uint8(data[i] & 0xff)
				c = gocolor.Gray{val}
			}
		} else if enc.ColorComponents == 3 {
			if enc.BitsPerComponent == 16 {
				r := uint16(data[i])<<8 | uint16(data[i+1])
				g := uint16(data[i+2])<<8 | uint16(data[i+3])
				b := uint16(data[i+4])<<8 | uint16(data[i+5])
				c = gocolor.RGBA64{R: r, G: g, B: b, A: 0}
			} else {
				r := uint8(data[i] & 0xff)
				g := uint8(data[i+1] & 0xff)
				b := uint8(data[i+2] & 0xff)
				c = gocolor.RGBA{R: r, G: g, B: b, A: 0}
			}
		} else if enc.ColorComponents == 4 {
			c1 := uint8(data[i] & 0xff)
			m1 := uint8(data[i+1] & 0xff)
			y1 := uint8(data[i+2] & 0xff)
			k1 := uint8(data[i+3] & 0xff)
			c = gocolor.CMYK{C: c1, M: m1, Y: y1, K: k1}
		}

		img.Set(x, y, c)
		x++
		if x == enc.Width {
			x = 0
			y++
		}
	}

	// The quality is specified from 0-100 (with 100 being the best quality) in the DCT structure.
	// N.B. even 100 is lossy, as still is transformed, but as good as it gets for DCT.
	// This is not related to the DPI, but rather inherent transformation losses.

	opt := jpeg.Options{}
	opt.Quality = enc.Quality

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &opt)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RunLengthEncoder represents Run length encoding.
type RunLengthEncoder struct {
}

// NewRunLengthEncoder makes a new run length encoder
func NewRunLengthEncoder() *RunLengthEncoder {
	return &RunLengthEncoder{}
}

// GetFilterName returns the name of the encoding filter.
func (enc *RunLengthEncoder) GetFilterName() string {
	return StreamEncodingFilterNameRunLength
}

// Create a new run length decoder from a stream object.
func newRunLengthEncoderFromStream(streamObj *PdfObjectStream, decodeParams *PdfObjectDictionary) (*RunLengthEncoder, error) {
	// TODO(dennwc): unused paramaters; check if it can have any in PDF spec
	return NewRunLengthEncoder(), nil
}

// DecodeBytes decodes a byte slice from Run length encoding.
//
// 7.4.5 RunLengthDecode Filter
// The RunLengthDecode filter decodes data that has been encoded in a simple byte-oriented format based on run length.
// The encoded data shall be a sequence of runs, where each run shall consist of a length byte followed by 1 to 128
// bytes of data. If the length byte is in the range 0 to 127, the following length + 1 (1 to 128) bytes shall be
// copied literally during decompression. If length is in the range 129 to 255, the following single byte shall be
// copied 257 - length (2 to 128) times during decompression. A length value of 128 shall denote EOD.
func (enc *RunLengthEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	// TODO(dennwc): use encoded slice directly, instead of wrapping it into a Reader
	bufReader := bytes.NewReader(encoded)
	var inb []byte
	for {
		b, err := bufReader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b > 128 {
			v, err := bufReader.ReadByte()
			if err != nil {
				return nil, err
			}
			for i := 0; i < 257-int(b); i++ {
				inb = append(inb, v)
			}
		} else if b < 128 {
			for i := 0; i < int(b)+1; i++ {
				v, err := bufReader.ReadByte()
				if err != nil {
					return nil, err
				}
				inb = append(inb, v)
			}
		} else {
			break
		}
	}

	return inb, nil
}

// DecodeStream decodes RunLengthEncoded stream object and give back decoded bytes.
func (enc *RunLengthEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return enc.DecodeBytes(streamObj.Stream)
}

// EncodeBytes encodes a bytes array and return the encoded value based on the encoder parameters.
func (enc *RunLengthEncoder) EncodeBytes(data []byte) ([]byte, error) {
	bufReader := bytes.NewReader(data)
	var inb []byte
	var literal []byte

	b0, err := bufReader.ReadByte()
	if err == io.EOF {
		return []byte{}, nil
	} else if err != nil {
		return nil, err
	}
	runLen := 1

	for {
		b, err := bufReader.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if b == b0 {
			if len(literal) > 0 {
				literal = literal[:len(literal)-1]
				if len(literal) > 0 {
					inb = append(inb, byte(len(literal)-1))
					inb = append(inb, literal...)
				}
				runLen = 1
				literal = []byte{}
			}
			runLen++
			if runLen >= 127 {
				inb = append(inb, byte(257-runLen), b0)
				runLen = 0
			}

		} else {
			if runLen > 0 {
				if runLen == 1 {
					literal = []byte{b0}
				} else {
					inb = append(inb, byte(257-runLen), b0)
				}

				runLen = 0
			}
			literal = append(literal, b)
			if len(literal) >= 127 {
				inb = append(inb, byte(len(literal)-1))
				inb = append(inb, literal...)
				literal = []byte{}
			}
		}
		b0 = b
	}

	if len(literal) > 0 {
		inb = append(inb, byte(len(literal)-1))
		inb = append(inb, literal...)
	} else if runLen > 0 {
		inb = append(inb, byte(257-runLen), b0)
	}
	inb = append(inb, 128)
	return inb, nil
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *RunLengthEncoder) MakeDecodeParams() PdfObject {
	return nil
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
func (enc *RunLengthEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()
	dict.Set("Filter", MakeName(enc.GetFilterName()))
	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *RunLengthEncoder) UpdateParams(params *PdfObjectDictionary) {
}

// ASCIIHexEncoder implements ASCII hex encoder/decoder.
type ASCIIHexEncoder struct {
}

// NewASCIIHexEncoder makes a new ASCII hex encoder.
func NewASCIIHexEncoder() *ASCIIHexEncoder {
	encoder := &ASCIIHexEncoder{}
	return encoder
}

// GetFilterName returns the name of the encoding filter.
func (enc *ASCIIHexEncoder) GetFilterName() string {
	return StreamEncodingFilterNameASCIIHex
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *ASCIIHexEncoder) MakeDecodeParams() PdfObject {
	return nil
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
func (enc *ASCIIHexEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()
	dict.Set("Filter", MakeName(enc.GetFilterName()))
	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *ASCIIHexEncoder) UpdateParams(params *PdfObjectDictionary) {
}

// DecodeBytes decodes a slice of ASCII encoded bytes and returns the result.
func (enc *ASCIIHexEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	bufReader := bytes.NewReader(encoded)
	var inb []byte
	for {
		b, err := bufReader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == '>' {
			break
		}
		if IsWhiteSpace(b) {
			continue
		}
		if (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') || (b >= '0' && b <= '9') {
			inb = append(inb, b)
		} else {
			common.Log.Debug("ERROR: Invalid ascii hex character (%c)", b)
			return nil, fmt.Errorf("invalid ascii hex character (%c)", b)
		}
	}
	if len(inb)%2 == 1 {
		inb = append(inb, '0')
	}
	common.Log.Trace("Inbound %s", inb)
	outb := make([]byte, hex.DecodedLen(len(inb)))
	_, err := hex.Decode(outb, inb)
	if err != nil {
		return nil, err
	}
	return outb, nil
}

// DecodeStream implements ASCII hex decoding.
func (enc *ASCIIHexEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return enc.DecodeBytes(streamObj.Stream)
}

// EncodeBytes ASCII encodes the passed in slice of bytes.
func (enc *ASCIIHexEncoder) EncodeBytes(data []byte) ([]byte, error) {
	var encoded bytes.Buffer

	for _, b := range data {
		encoded.WriteString(fmt.Sprintf("%.2X ", b))
	}
	encoded.WriteByte('>')

	return encoded.Bytes(), nil
}

// ASCII85Encoder implements ASCII85 encoder/decoder.
type ASCII85Encoder struct {
}

// NewASCII85Encoder makes a new ASCII85 encoder.
func NewASCII85Encoder() *ASCII85Encoder {
	encoder := &ASCII85Encoder{}
	return encoder
}

// GetFilterName returns the name of the encoding filter.
func (enc *ASCII85Encoder) GetFilterName() string {
	return StreamEncodingFilterNameASCII85
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *ASCII85Encoder) MakeDecodeParams() PdfObject {
	return nil
}

// MakeStreamDict make a new instance of an encoding dictionary for a stream object.
func (enc *ASCII85Encoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()
	dict.Set("Filter", MakeName(enc.GetFilterName()))
	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *ASCII85Encoder) UpdateParams(params *PdfObjectDictionary) {
}

// DecodeBytes decodes byte array with ASCII85. 5 ASCII characters -> 4 raw binary bytes
func (enc *ASCII85Encoder) DecodeBytes(encoded []byte) ([]byte, error) {
	var decoded []byte

	common.Log.Trace("ASCII85 Decode")

	i := 0
	eod := false

	for i < len(encoded) && !eod {
		codes := [5]byte{0, 0, 0, 0, 0}
		spaces := 0 // offset due to whitespace.
		j := 0
		toWrite := 4
		for j < 5+spaces {
			if i+j == len(encoded) {
				break
			}
			code := encoded[i+j]
			if IsWhiteSpace(code) {
				// Skip whitespace.
				spaces++
				j++
				continue
			} else if code == '~' && i+j+1 < len(encoded) && encoded[i+j+1] == '>' {
				toWrite = (j - spaces) - 1
				if toWrite < 0 {
					toWrite = 0
				}
				// EOD marker.  Marks end of data.
				eod = true
				break
			} else if code >= '!' && code <= 'u' {
				// Valid code.
				code -= '!'
			} else if code == 'z' && j-spaces == 0 {
				// 'z' in beginning of the byte sequence means that all 5 codes are 0.
				// Already all 0 initialized, so can break here.
				toWrite = 4
				j++
				break
			} else {
				common.Log.Error("Failed decoding, invalid code")
				return nil, errors.New("invalid code encountered")
			}

			codes[j-spaces] = code
			j++
		}
		i += j

		// Pad with 'u' 84 (unused ones)
		// Takes care of issues at ends for input data that is not a multiple of 4-bytes.
		for m := toWrite + 1; m < 5; m++ {
			codes[m] = 84
		}

		// Convert to a uint32 value.
		value := uint32(codes[0])*85*85*85*85 + uint32(codes[1])*85*85*85 + uint32(codes[2])*85*85 + uint32(codes[3])*85 + uint32(codes[4])

		// Convert to 4 bytes.
		decodedBytes := []byte{
			byte((value >> 24) & 0xff),
			byte((value >> 16) & 0xff),
			byte((value >> 8) & 0xff),
			byte(value & 0xff)}

		// This accounts for the end of data, where the original data length is not a multiple of 4.
		// In that case, 0 bytes are assumed but only
		decoded = append(decoded, decodedBytes[:toWrite]...)
	}

	common.Log.Trace("ASCII85, encoded: % X", encoded)
	common.Log.Trace("ASCII85, decoded: % X", decoded)

	return decoded, nil
}

// DecodeStream implements ASCII85 stream decoding.
func (enc *ASCII85Encoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return enc.DecodeBytes(streamObj.Stream)
}

// Convert a base 256 number to a series of base 85 values (5 codes).
//  85^5 = 4437053125 > 256^4 = 4294967296
// So 5 base-85 numbers will always be enough to cover 4 base-256 numbers.
// The base 256 value is already converted to an uint32 value.
func (enc *ASCII85Encoder) base256Tobase85(base256val uint32) [5]byte {
	base85 := [5]byte{0, 0, 0, 0, 0}
	remainder := base256val
	for i := 0; i < 5; i++ {
		divider := uint32(1)
		for j := 0; j < 4-i; j++ {
			divider *= 85
		}
		val := remainder / divider
		remainder = remainder % divider
		base85[i] = byte(val)
	}
	return base85
}

// EncodeBytes encodes data into ASCII85 encoded format.
func (enc *ASCII85Encoder) EncodeBytes(data []byte) ([]byte, error) {
	var encoded bytes.Buffer

	for i := 0; i < len(data); i += 4 {
		b1 := data[i]
		n := 1

		b2 := byte(0)
		if i+1 < len(data) {
			b2 = data[i+1]
			n++
		}

		b3 := byte(0)
		if i+2 < len(data) {
			b3 = data[i+2]
			n++
		}

		b4 := byte(0)
		if i+3 < len(data) {
			b4 = data[i+3]
			n++
		}

		// Convert to a uint32 number.
		base256 := (uint32(b1) << 24) | (uint32(b2) << 16) | (uint32(b3) << 8) | uint32(b4)
		if base256 == 0 {
			encoded.WriteByte('z')
		} else {
			base85vals := enc.base256Tobase85(base256)
			for _, val := range base85vals[:n+1] {
				encoded.WriteByte(val + '!')
			}
		}
	}

	// EOD.
	encoded.WriteString("~>")
	return encoded.Bytes(), nil
}

// RawEncoder implements Raw encoder/decoder (no encoding, pass through)
type RawEncoder struct{}

// NewRawEncoder returns a new instace of RawEncoder.
func NewRawEncoder() *RawEncoder {
	return &RawEncoder{}
}

// GetFilterName returns the name of the encoding filter.
func (enc *RawEncoder) GetFilterName() string {
	return StreamEncodingFilterNameRaw
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *RawEncoder) MakeDecodeParams() PdfObject {
	return nil
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
func (enc *RawEncoder) MakeStreamDict() *PdfObjectDictionary {
	return MakeDict()
}

// UpdateParams updates the parameter values of the encoder.
func (enc *RawEncoder) UpdateParams(params *PdfObjectDictionary) {
}

// DecodeBytes returns the passed in slice of bytes.
// The purpose of the method is to satisfy the StreamEncoder interface.
func (enc *RawEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	return encoded, nil
}

// DecodeStream returns the passed in stream as a slice of bytes.
// The purpose of the method is to satisfy the StreamEncoder interface.
func (enc *RawEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return streamObj.Stream, nil
}

// EncodeBytes returns the passed in slice of bytes.
// The purpose of the method is to satisfy the StreamEncoder interface.
func (enc *RawEncoder) EncodeBytes(data []byte) ([]byte, error) {
	return data, nil
}

// CCITTFaxEncoder implements Group3 and Group4 facsimile (fax) encoder/decoder.
type CCITTFaxEncoder struct {
	K                      int
	EndOfLine              bool
	EncodedByteAlign       bool
	Columns                int
	Rows                   int
	EndOfBlock             bool
	BlackIs1               bool
	DamagedRowsBeforeError int
}

// NewCCITTFaxEncoder makes a new CCITTFax encoder.
func NewCCITTFaxEncoder() *CCITTFaxEncoder {
	return &CCITTFaxEncoder{
		Columns:    1728,
		EndOfBlock: true,
	}
}

// GetFilterName returns the name of the encoding filter.
func (enc *CCITTFaxEncoder) GetFilterName() string {
	return StreamEncodingFilterNameCCITTFax
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *CCITTFaxEncoder) MakeDecodeParams() PdfObject {
	decodeParams := MakeDict()
	decodeParams.Set("K", MakeInteger(int64(enc.K)))
	decodeParams.Set("Columns", MakeInteger(int64(enc.Columns)))

	if enc.BlackIs1 {
		decodeParams.Set("BlackIs1", MakeBool(enc.BlackIs1))
	}

	if enc.EncodedByteAlign {
		decodeParams.Set("EncodedByteAlign", MakeBool(enc.EncodedByteAlign))
	}

	if enc.EndOfLine && enc.K >= 0 {
		decodeParams.Set("EndOfLine", MakeBool(enc.EndOfLine))
	}

	if enc.Rows != 0 && !enc.EndOfBlock {
		decodeParams.Set("Rows", MakeInteger(int64(enc.Rows)))
	}

	if !enc.EndOfBlock {
		decodeParams.Set("EndOfBlock", MakeBool(enc.EndOfBlock))
	}

	if enc.DamagedRowsBeforeError != 0 {
		decodeParams.Set("DamagedRowsBeforeError", MakeInteger(int64(enc.DamagedRowsBeforeError)))
	}

	return decodeParams
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
func (enc *CCITTFaxEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()
	dict.Set("Filter", MakeName(enc.GetFilterName()))

	decodeParams := enc.MakeDecodeParams()
	if decodeParams != nil {
		dict.Set("DecodeParms", decodeParams)
	}

	return dict
}

// newCCITTFaxEncoderFromStream creates a new CCITTFax decoder from a stream object, getting all the encoding parameters
// from the DecodeParms stream object dictionary entry.
func newCCITTFaxEncoderFromStream(streamObj *PdfObjectStream, decodeParams *PdfObjectDictionary) (*CCITTFaxEncoder, error) {
	encoder := NewCCITTFaxEncoder()

	encDict := streamObj.PdfObjectDictionary
	if encDict == nil {
		// No encoding dictionary.
		return encoder, nil
	}

	// If decodeParams not provided, see if we can get from the stream.
	if decodeParams == nil {
		obj := TraceToDirectObject(encDict.Get("DecodeParms"))
		if obj != nil {
			switch t := obj.(type) {
			case *PdfObjectDictionary:
				decodeParams = t
				break
			case *PdfObjectArray:
				if t.Len() == 1 {
					if dp, ok := GetDict(t.Get(0)); ok {
						decodeParams = dp
					}
				}
			default:
				common.Log.Error("DecodeParms not a dictionary %#v", obj)
				return nil, errors.New("invalid DecodeParms")
			}
		}
		if decodeParams == nil {
			common.Log.Error("DecodeParms not specified %#v", obj)
			return nil, errors.New("invalid DecodeParms")
		}
	}

	if k, err := GetNumberAsInt64(decodeParams.Get("K")); err == nil {
		encoder.K = int(k)
	}

	if columns, err := GetNumberAsInt64(decodeParams.Get("Columns")); err == nil {
		encoder.Columns = int(columns)
	} else {
		encoder.Columns = 1728
	}

	if blackIs1, err := GetNumberAsInt64(decodeParams.Get("BlackIs1")); err == nil {
		encoder.BlackIs1 = blackIs1 > 0
	} else {
		if blackIs1, ok := GetBoolVal(decodeParams.Get("BlackIs1")); ok {
			encoder.BlackIs1 = blackIs1
		} else {
			if decodeArr, ok := GetArray(decodeParams.Get("Decode")); ok {
				intArr, err := decodeArr.ToIntegerArray()
				if err == nil {
					encoder.BlackIs1 = intArr[0] == 1 && intArr[1] == 0
				}
			}
		}
	}

	if align, err := GetNumberAsInt64(decodeParams.Get("EncodedByteAlign")); err == nil {
		encoder.EncodedByteAlign = align > 0
	} else {
		if align, ok := GetBoolVal(decodeParams.Get("EncodedByteAlign")); ok {
			encoder.EncodedByteAlign = align
		}
	}

	if eol, err := GetNumberAsInt64(decodeParams.Get("EndOfLine")); err == nil {
		encoder.EndOfLine = eol > 0
	} else {
		if eol, ok := GetBoolVal(decodeParams.Get("EndOfLine")); ok {
			encoder.EndOfLine = eol
		}
	}

	if rows, err := GetNumberAsInt64(decodeParams.Get("Rows")); err == nil {
		encoder.Rows = int(rows)
	}

	encoder.EndOfBlock = true
	if eofb, err := GetNumberAsInt64(decodeParams.Get("EndOfBlock")); err == nil {
		encoder.EndOfBlock = eofb > 0
	} else {
		if eofb, ok := GetBoolVal(decodeParams.Get("EndOfBlock")); ok {
			encoder.EndOfBlock = eofb
		}
	}

	if rows, err := GetNumberAsInt64(decodeParams.Get("DamagedRowsBeforeError")); err != nil {
		encoder.DamagedRowsBeforeError = int(rows)
	}

	common.Log.Trace("decode params: %s", decodeParams.String())
	return encoder, nil
}

// UpdateParams updates the parameter values of the encoder.
func (enc *CCITTFaxEncoder) UpdateParams(params *PdfObjectDictionary) {
	if k, err := GetNumberAsInt64(params.Get("K")); err == nil {
		enc.K = int(k)
	}

	if columns, err := GetNumberAsInt64(params.Get("Columns")); err == nil {
		enc.Columns = int(columns)
	} else if columns, err = GetNumberAsInt64(params.Get("Width")); err == nil {
		enc.Columns = int(columns)
	}

	if blackIs1, err := GetNumberAsInt64(params.Get("BlackIs1")); err == nil {
		enc.BlackIs1 = blackIs1 > 0
	} else {
		if blackIs1, ok := GetBoolVal(params.Get("BlackIs1")); ok {
			enc.BlackIs1 = blackIs1
		} else {
			if decodeArr, ok := GetArray(params.Get("Decode")); ok {
				intArr, err := decodeArr.ToIntegerArray()
				if err == nil {
					enc.BlackIs1 = intArr[0] == 1 && intArr[1] == 0
				}
			}
		}
	}

	if align, err := GetNumberAsInt64(params.Get("EncodedByteAlign")); err == nil {
		enc.EncodedByteAlign = align > 0
	} else {
		if align, ok := GetBoolVal(params.Get("EncodedByteAlign")); ok {
			enc.EncodedByteAlign = align
		}
	}

	if eol, err := GetNumberAsInt64(params.Get("EndOfLine")); err == nil {
		enc.EndOfLine = eol > 0
	} else {
		if eol, ok := GetBoolVal(params.Get("EndOfLine")); ok {
			enc.EndOfLine = eol
		}
	}

	if rows, err := GetNumberAsInt64(params.Get("Rows")); err == nil {
		enc.Rows = int(rows)
	} else if rows, err = GetNumberAsInt64(params.Get("Height")); err == nil {
		enc.Rows = int(rows)
	}

	if eofb, err := GetNumberAsInt64(params.Get("EndOfBlock")); err == nil {
		enc.EndOfBlock = eofb > 0
	} else {
		if eofb, ok := GetBoolVal(params.Get("EndOfBlock")); ok {
			enc.EndOfBlock = eofb
		}
	}

	if rows, err := GetNumberAsInt64(params.Get("DamagedRowsBeforeError")); err != nil {
		enc.DamagedRowsBeforeError = int(rows)
	}
}

// DecodeBytes decodes the CCITTFax encoded image data.
func (enc *CCITTFaxEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	encoder := &ccittfax.Encoder{
		K:                      enc.K,
		Columns:                enc.Columns,
		EndOfLine:              enc.EndOfLine,
		EndOfBlock:             enc.EndOfBlock,
		BlackIs1:               enc.BlackIs1,
		DamagedRowsBeforeError: enc.DamagedRowsBeforeError,
		Rows:                   enc.Rows,
		EncodedByteAlign:       enc.EncodedByteAlign,
	}

	pixels, err := encoder.Decode(encoded)
	if err != nil {
		return nil, err
	}

	// reassemble image
	var decoded []byte
	decodedIdx := 0
	var bitPos byte
	var currentByte byte
	for i := range pixels {
		for j := range pixels[i] {
			currentByte |= pixels[i][j] << (7 - bitPos)

			bitPos++

			if bitPos == 8 {
				decoded = append(decoded, currentByte)
				currentByte = 0

				decodedIdx++

				bitPos = 0
			}
		}
	}

	if bitPos > 0 {
		decoded = append(decoded, currentByte)
	}

	return decoded, nil
}

// DecodeStream decodes the stream containing CCITTFax encoded image data.
func (enc *CCITTFaxEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return enc.DecodeBytes(streamObj.Stream)
}

// EncodeBytes encodes the image data using either Group3 or Group4 CCITT facsimile (fax) encoding.
// `data` is expected to be 1 color component, 1 byte per component.
func (enc *CCITTFaxEncoder) EncodeBytes(data []byte) ([]byte, error) {
	var pixels [][]byte

	for i := 0; i < len(data); i += enc.Columns {
		pixelsRow := make([]byte, enc.Columns)

		pixel := 0
		for j := 0; j < enc.Columns; j++ {
			if data[i+j] == 255 {
				pixelsRow[pixel] = 1
			} else {
				pixelsRow[pixel] = 0
			}

			pixel++
		}

		pixels = append(pixels, pixelsRow)
	}

	encoder := &ccittfax.Encoder{
		K:                      enc.K,
		Columns:                enc.Columns,
		EndOfLine:              enc.EndOfLine,
		EndOfBlock:             enc.EndOfBlock,
		BlackIs1:               enc.BlackIs1,
		DamagedRowsBeforeError: enc.DamagedRowsBeforeError,
		Rows:                   enc.Rows,
		EncodedByteAlign:       enc.EncodedByteAlign,
	}

	return encoder.Encode(pixels), nil
}

// JPXEncoder implements JPX encoder/decoder (dummy, for now)
// FIXME: implement
type JPXEncoder struct{}

// NewJPXEncoder returns a new instance of JPXEncoder.
func NewJPXEncoder() *JPXEncoder {
	return &JPXEncoder{}
}

// GetFilterName returns the name of the encoding filter.
func (enc *JPXEncoder) GetFilterName() string {
	return StreamEncodingFilterNameJPX
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *JPXEncoder) MakeDecodeParams() PdfObject {
	return nil
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
func (enc *JPXEncoder) MakeStreamDict() *PdfObjectDictionary {
	return MakeDict()
}

// UpdateParams updates the parameter values of the encoder.
func (enc *JPXEncoder) UpdateParams(params *PdfObjectDictionary) {
}

// DecodeBytes decodes a slice of JPX encoded bytes and returns the result.
func (enc *JPXEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	common.Log.Debug("Error: Attempting to use unsupported encoding %s", enc.GetFilterName())
	return encoded, ErrNoJPXDecode
}

// DecodeStream decodes a JPX encoded stream and returns the result as a
// slice of bytes.
func (enc *JPXEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	common.Log.Debug("Error: Attempting to use unsupported encoding %s", enc.GetFilterName())
	return streamObj.Stream, ErrNoJPXDecode
}

// EncodeBytes JPX encodes the passed in slice of bytes.
func (enc *JPXEncoder) EncodeBytes(data []byte) ([]byte, error) {
	common.Log.Debug("Error: Attempting to use unsupported encoding %s", enc.GetFilterName())
	return data, ErrNoJPXDecode
}

// MultiEncoder supports serial encoding.
type MultiEncoder struct {
	// Encoders in the order that they are to be applied.
	encoders []StreamEncoder
}

// NewMultiEncoder returns a new instance of MultiEncoder.
func NewMultiEncoder() *MultiEncoder {
	encoder := MultiEncoder{}
	encoder.encoders = []StreamEncoder{}

	return &encoder
}

func newMultiEncoderFromStream(streamObj *PdfObjectStream) (*MultiEncoder, error) {
	mencoder := NewMultiEncoder()

	encDict := streamObj.PdfObjectDictionary
	if encDict == nil {
		// No encoding dictionary.
		return mencoder, nil
	}

	// Prepare the decode params array (one for each filter type)
	// Optional, not always present.
	var decodeParamsDict *PdfObjectDictionary
	var decodeParamsArray []PdfObject
	obj := encDict.Get("DecodeParms")
	if obj != nil {
		// If it is a dictionary, assume it applies to all
		dict, isDict := obj.(*PdfObjectDictionary)
		if isDict {
			decodeParamsDict = dict
		}

		// If it is an array, assume there is one for each
		arr, isArray := obj.(*PdfObjectArray)
		if isArray {
			for _, dictObj := range arr.Elements() {
				dictObj = TraceToDirectObject(dictObj)
				if dict, is := dictObj.(*PdfObjectDictionary); is {
					decodeParamsArray = append(decodeParamsArray, dict)
				} else {
					decodeParamsArray = append(decodeParamsArray, MakeDict())
				}
			}
		}
	}

	obj = encDict.Get("Filter")
	if obj == nil {
		return nil, fmt.Errorf("filter missing")
	}

	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("multi filter can only be made from array")
	}

	for idx, obj := range array.Elements() {
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("multi filter array element not a name")
		}

		var dp PdfObject

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

		var dParams *PdfObjectDictionary
		if dict, is := dp.(*PdfObjectDictionary); is {
			dParams = dict
		}

		common.Log.Trace("Next name: %s, dp: %v, dParams: %v", *name, dp, dParams)
		if *name == StreamEncodingFilterNameFlate {
			// TODO: need to separate out the DecodeParms..
			encoder, err := newFlateEncoderFromStream(streamObj, dParams)
			if err != nil {
				return nil, err
			}
			mencoder.AddEncoder(encoder)
		} else if *name == StreamEncodingFilterNameLZW {
			encoder, err := newLZWEncoderFromStream(streamObj, dParams)
			if err != nil {
				return nil, err
			}
			mencoder.AddEncoder(encoder)
		} else if *name == StreamEncodingFilterNameASCIIHex {
			encoder := NewASCIIHexEncoder()
			mencoder.AddEncoder(encoder)
		} else if *name == StreamEncodingFilterNameASCII85 {
			encoder := NewASCII85Encoder()
			mencoder.AddEncoder(encoder)
		} else if *name == StreamEncodingFilterNameDCT {
			encoder, err := newDCTEncoderFromStream(streamObj, mencoder)
			if err != nil {
				return nil, err
			}
			mencoder.AddEncoder(encoder)
			common.Log.Trace("Added DCT encoder...")
			common.Log.Trace("Multi encoder: %#v", mencoder)
		} else {
			common.Log.Error("Unsupported filter %s", *name)
			return nil, fmt.Errorf("invalid filter in multi filter array")
		}
	}

	return mencoder, nil
}

// GetFilterName returns the names of the underlying encoding filters,
// separated by spaces.
// Note: This is just a string, should not be used in /Filter dictionary entry. Use GetFilterArray for that.
// TODO(v4): Refactor to GetFilter() which can be used for /Filter (either Name or Array), this can be
//  renamed to String() as a pretty string to use in debugging etc.
func (enc *MultiEncoder) GetFilterName() string {
	name := ""
	for idx, encoder := range enc.encoders {
		name += encoder.GetFilterName()
		if idx < len(enc.encoders)-1 {
			name += " "
		}
	}
	return name
}

// GetFilterArray returns the names of the underlying encoding filters in an array that
// can be used as /Filter entry.
func (enc *MultiEncoder) GetFilterArray() *PdfObjectArray {
	names := make([]PdfObject, len(enc.encoders))
	for i, e := range enc.encoders {
		names[i] = MakeName(e.GetFilterName())
	}
	return MakeArray(names...)
}

// MakeDecodeParams makes a new instance of an encoding dictionary based on
// the current encoder settings.
func (enc *MultiEncoder) MakeDecodeParams() PdfObject {
	if len(enc.encoders) == 0 {
		return nil
	}

	if len(enc.encoders) == 1 {
		return enc.encoders[0].MakeDecodeParams()
	}

	array := MakeArray()
	for _, encoder := range enc.encoders {
		decodeParams := encoder.MakeDecodeParams()
		if decodeParams == nil {
			array.Append(MakeNull())
		} else {
			array.Append(decodeParams)
		}
	}

	return array
}

// AddEncoder adds the passed in encoder to the underlying encoder slice.
func (enc *MultiEncoder) AddEncoder(encoder StreamEncoder) {
	enc.encoders = append(enc.encoders, encoder)
}

// MakeStreamDict makes a new instance of an encoding dictionary for a stream object.
func (enc *MultiEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := MakeDict()
	dict.Set("Filter", enc.GetFilterArray())

	// Pass all values from children, except Filter and DecodeParms.
	for _, encoder := range enc.encoders {
		encDict := encoder.MakeStreamDict()
		for _, key := range encDict.Keys() {
			val := encDict.Get(key)
			if key != "Filter" && key != "DecodeParms" {
				dict.Set(key, val)
			}
		}
	}

	// Make the decode params array or dict.
	decodeParams := enc.MakeDecodeParams()
	if decodeParams != nil {
		dict.Set("DecodeParms", decodeParams)
	}

	return dict
}

// UpdateParams updates the parameter values of the encoder.
func (enc *MultiEncoder) UpdateParams(params *PdfObjectDictionary) {
	for _, encoder := range enc.encoders {
		encoder.UpdateParams(params)
	}
}

// DecodeBytes decodes a multi-encoded slice of bytes by passing it through the
// DecodeBytes method of the underlying encoders.
func (enc *MultiEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	decoded := encoded
	var err error
	// Apply in forward order.
	for _, encoder := range enc.encoders {
		common.Log.Trace("Multi Encoder Decode: Applying Filter: %v %T", encoder, encoder)

		decoded, err = encoder.DecodeBytes(decoded)
		if err != nil {
			return nil, err
		}
	}

	return decoded, nil
}

// DecodeStream decodes a multi-encoded stream by passing it through the
// DecodeStream method of the underlying encoders.
func (enc *MultiEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return enc.DecodeBytes(streamObj.Stream)
}

// EncodeBytes encodes the passed in slice of bytes by passing it through the
// EncodeBytes method of the underlying encoders.
func (enc *MultiEncoder) EncodeBytes(data []byte) ([]byte, error) {
	encoded := data
	var err error

	// Apply in inverse order.
	for i := len(enc.encoders) - 1; i >= 0; i-- {
		encoder := enc.encoders[i]
		encoded, err = encoder.EncodeBytes(encoded)
		if err != nil {
			return nil, err
		}
	}

	return encoded, nil
}
