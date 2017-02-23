/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

// Implement encoders for PDF. Currently supported:
// - Raw (Identity)
// - FlateDecode
// - LZW
// - ASCII Hex

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	// Need two slightly different implementations of LZW (EarlyChange parameter).
	lzw0 "compress/lzw"
	lzw1 "golang.org/x/image/tiff/lzw"

	"github.com/unidoc/unidoc/common"
)

const (
	StreamEncodingFilterNameFlate    = "FlateDecode"
	StreamEncodingFilterNameLZW      = "LZWDecode"
	StreamEncodingFilterNameASCIIHex = "ASCIIHexDecode"
	StreamEncodingFilterNameASCII85  = "ASCII85Decode"
)

type StreamEncoder interface {
	GetFilterName() string
	MakeDecodeParams() PdfObject
	MakeStreamDict() *PdfObjectDictionary

	EncodeBytes(data []byte) ([]byte, error)
	DecodeBytes(encoded []byte) ([]byte, error)
	DecodeStream(streamObj *PdfObjectStream) ([]byte, error)
}

// Flate encoding.
type FlateEncoder struct {
	Predictor        int
	BitsPerComponent int
	// For predictors
	Columns int
	Colors  int
}

// Make a new flate encoder with default parameters, predictor 1 and bits per component 8.
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

func (this *FlateEncoder) GetFilterName() string {
	return StreamEncodingFilterNameFlate
}

func (this *FlateEncoder) MakeDecodeParams() PdfObject {
	if this.Predictor > 1 {
		decodeParams := PdfObjectDictionary{}
		decodeParams["Predictor"] = MakeInteger(int64(this.Predictor))

		// Only add if not default option.
		if this.BitsPerComponent != 8 {
			decodeParams["BitsPerComponent"] = MakeInteger(int64(this.BitsPerComponent))
		}
		if this.Columns != 1 {
			decodeParams["Columns"] = MakeInteger(int64(this.Columns))
		}
		if this.Colors != 1 {
			decodeParams["Colors"] = MakeInteger(int64(this.Colors))
		}
		return &decodeParams
	}
	return nil
}

// Make a new instance of an encoding dictionary for a stream object.
// Has the Filter set and the DecodeParms.
func (this *FlateEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := PdfObjectDictionary{}

	dict["Filter"] = MakeName(this.GetFilterName())

	decodeParams := this.MakeDecodeParams()
	if decodeParams != nil {
		dict["DecodeParms"] = decodeParams
	}

	return &dict
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
		obj := (*encDict)["DecodeParms"]
		if obj != nil {
			dp, isDict := obj.(*PdfObjectDictionary)
			if !isDict {
				common.Log.Debug("Error: DecodeParms not a dictionary (%T)", obj)
				return nil, fmt.Errorf("Invalid DecodeParms")
			}
			decodeParams = dp
		}
	}
	if decodeParams == nil {
		// Can safely return here if no decode params, as the following depend on the decode params.
		return encoder, nil
	}

	common.Log.Debug("decode params: %s", decodeParams.String())
	obj, has := (*decodeParams)["Predictor"]
	if !has {
		common.Log.Debug("Error: Predictor missing from DecodeParms - Continue with default (1)")
	} else {
		predictor, ok := obj.(*PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: Predictor specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("Invalid Predictor")
		}
		encoder.Predictor = int(*predictor)
	}

	// Bits per component.  Use default if not specified (8).
	obj, has = (*decodeParams)["BitsPerComponent"]
	if has {
		bpc, ok := obj.(*PdfObjectInteger)
		if !ok {
			common.Log.Debug("ERROR: Invalid BitsPerComponent")
			return nil, fmt.Errorf("Invalid BitsPerComponent")
		}
		encoder.BitsPerComponent = int(*bpc)
	}

	if encoder.Predictor > 1 {
		// Columns.
		encoder.Columns = 1
		obj, has = (*decodeParams)["Columns"]
		if has {
			columns, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("Predictor column invalid")
			}

			encoder.Columns = int(*columns)
		}

		// Colors.
		// Number of interleaved color components per sample (Default 1 if not specified)
		encoder.Colors = 1
		obj, has = (*decodeParams)["Colors"]
		if has {
			colors, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("Predictor colors not an integer")
			}
			encoder.Colors = int(*colors)
		}
	}

	return encoder, nil
}

func (this *FlateEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
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

// Decode a FlateEncoded stream object and give back decoded bytes.
func (this *FlateEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	// TODO: Revamp this support to handle TIFF predictor (2).
	// Also handle more filter bytes and support more values of BitsPerComponent.

	common.Log.Debug("FlateDecode")
	common.Log.Debug("Predictor: %d", this.Predictor)
	if this.BitsPerComponent != 8 {
		return nil, fmt.Errorf("Invalid BitsPerComponent (only 8 supported)")
	}

	outData, err := this.DecodeBytes(streamObj.Stream)
	if err != nil {
		return nil, err
	}
	common.Log.Debug("En: % x\n", streamObj.Stream)
	common.Log.Debug("De: % x\n", outData)

	if this.Predictor > 1 {
		if this.Predictor == 2 { // TIFF encoding: Needs some tests.
			common.Log.Debug("Tiff encoding")
			common.Log.Debug("Colors: %d", this.Colors)

			rowLength := int(this.Columns) * this.Colors
			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				common.Log.Debug("ERROR: TIFF encoding: Invalid row length...")
				return nil, fmt.Errorf("Invalid row length (%d/%d)", len(outData), rowLength)
			}

			if rowLength%this.Colors != 0 {
				return nil, fmt.Errorf("Invalid row length (%d) for colors %d", rowLength, this.Colors)
			}
			common.Log.Debug("inp outData (%d): % x", len(outData), outData)

			pOutBuffer := bytes.NewBuffer(nil)

			// 0-255  -255 255 ; 0-255=-255;
			for i := 0; i < rows; i++ {
				rowData := outData[rowLength*i : rowLength*(i+1)]
				//common.Log.Debug("RowData before: % d", rowData)
				// Predicts the same as the sample to the left.
				// Interleaved by colors.
				for j := this.Colors; j < rowLength; j++ {
					rowData[j] = byte(int(rowData[j]+rowData[j-this.Colors]) % 256)
				}
				pOutBuffer.Write(rowData)
			}
			pOutData := pOutBuffer.Bytes()
			common.Log.Debug("POutData (%d): % x", len(pOutData), pOutData)
			return pOutData, nil
		} else if this.Predictor >= 10 && this.Predictor <= 15 {
			common.Log.Debug("PNG Encoding")
			rowLength := int(this.Columns + 1) // 1 byte to specify predictor algorithms per row.
			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				common.Log.Debug("ERROR: Invalid row length...")
				return nil, fmt.Errorf("Invalid row length (%d/%d)", len(outData), rowLength)
			}

			pOutBuffer := bytes.NewBuffer(nil)

			common.Log.Debug("Predictor columns: %d", this.Columns)
			common.Log.Debug("Length: %d / %d = %d rows", len(outData), rowLength, rows)
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
					return nil, fmt.Errorf("Invalid filter byte (%d)", fb)
				}

				for i := 0; i < rowLength; i++ {
					prevRowData[i] = rowData[i]
				}
				pOutBuffer.Write(rowData[1:])
			}
			pOutData := pOutBuffer.Bytes()
			return pOutData, nil
		} else {
			common.Log.Debug("ERROR: Unsupported predictor (%d)", this.Predictor)
			return nil, fmt.Errorf("Unsupported predictor (%d)", this.Predictor)
		}
	}

	return outData, nil
}

// Encode a bytes array and return the encoded value based on the encoder parameters.
func (this *FlateEncoder) EncodeBytes(data []byte) ([]byte, error) {
	if this.Predictor != 1 {
		return nil, fmt.Errorf("FlateEncoder Predictor = 1 only supported yet")
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()

	return b.Bytes(), nil
}

// LZW encoding/decoding functionality.
type LZWEncoder struct {
	Predictor        int
	BitsPerComponent int
	// For predictors
	Columns int
	Colors  int
	// LZW algorithm setting.
	EarlyChange int
}

// Make a new LZW encoder with default parameters.
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

func (this *LZWEncoder) GetFilterName() string {
	return StreamEncodingFilterNameLZW
}

func (this *LZWEncoder) MakeDecodeParams() PdfObject {
	if this.Predictor > 1 {
		decodeParams := PdfObjectDictionary{}
		decodeParams["Predictor"] = MakeInteger(int64(this.Predictor))

		// Only add if not default option.
		if this.BitsPerComponent != 8 {
			decodeParams["BitsPerComponent"] = MakeInteger(int64(this.BitsPerComponent))
		}
		if this.Columns != 1 {
			decodeParams["Columns"] = MakeInteger(int64(this.Columns))
		}
		if this.Colors != 1 {
			decodeParams["Colors"] = MakeInteger(int64(this.Colors))
		}
		return &decodeParams
	}
	return nil
}

// Make a new instance of an encoding dictionary for a stream object.
// Has the Filter set and the DecodeParms.
func (this *LZWEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := PdfObjectDictionary{}

	dict["Filter"] = MakeName(this.GetFilterName())

	decodeParams := this.MakeDecodeParams()
	if decodeParams != nil {
		dict["DecodeParms"] = decodeParams
	}

	dict["EarlyChange"] = MakeInteger(int64(this.EarlyChange))

	return &dict
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
		obj := (*encDict)["DecodeParms"]
		if obj != nil {
			dp, isDict := obj.(*PdfObjectDictionary)
			if !isDict {
				common.Log.Debug("Error: DecodeParms not a dictionary (%T)", obj)
				return nil, fmt.Errorf("Invalid DecodeParms")
			}
			decodeParams = dp
		}
	}

	// The EarlyChange indicates when to increase code length, as different
	// implementations use a different mechanisms. Essentially this chooses
	// which LZW implementation to use.
	// The default is 1 (one code early)
	obj, has := (*decodeParams)["EarlyChange"]
	if has {
		earlyChange, ok := obj.(*PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: EarlyChange specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("Invalid EarlyChange")
		}
		if *earlyChange != 0 && *earlyChange != 1 {
			return nil, fmt.Errorf("Invalid EarlyChange value (not 0 or 1)")
		}

		encoder.EarlyChange = int(*earlyChange)
	}

	if decodeParams == nil {
		// No decode parameters. Can safely return here if not set as the following options
		// are related to the decode Params.
		return encoder, nil
	}

	obj, has = (*decodeParams)["Predictor"]
	if has {
		predictor, ok := obj.(*PdfObjectInteger)
		if !ok {
			common.Log.Debug("Error: Predictor specified but not numeric (%T)", obj)
			return nil, fmt.Errorf("Invalid Predictor")
		}
		encoder.Predictor = int(*predictor)
	}

	// Bits per component.  Use default if not specified (8).
	obj, has = (*decodeParams)["BitsPerComponent"]
	if has {
		bpc, ok := obj.(*PdfObjectInteger)
		if !ok {
			common.Log.Debug("ERROR: Invalid BitsPerComponent")
			return nil, fmt.Errorf("Invalid BitsPerComponent")
		}
		encoder.BitsPerComponent = int(*bpc)
	}

	if encoder.Predictor > 1 {
		// Columns.
		encoder.Columns = 1
		obj, has = (*decodeParams)["Columns"]
		if has {
			columns, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("Predictor column invalid")
			}

			encoder.Columns = int(*columns)
		}

		// Colors.
		// Number of interleaved color components per sample (Default 1 if not specified)
		encoder.Colors = 1
		obj, has = (*decodeParams)["Colors"]
		if has {
			colors, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, fmt.Errorf("Predictor colors not an integer")
			}
			encoder.Colors = int(*colors)
		}
	}

	common.Log.Debug("decode params: %s", decodeParams.String())
	return encoder, nil
}

func (this *LZWEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	var outBuf bytes.Buffer
	bufReader := bytes.NewReader(encoded)

	var r io.ReadCloser
	if this.EarlyChange == 1 {
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

func (this *LZWEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	// Revamp this support to handle TIFF predictor (2).
	// Also handle more filter bytes and check
	// BitsPerComponent.  Default value is 8, currently we are only
	// supporting that one.

	common.Log.Debug("LZW Decoding")
	common.Log.Debug("Predictor: %d", this.Predictor)

	outData, err := this.DecodeBytes(streamObj.Stream)
	if err != nil {
		return nil, err
	}

	common.Log.Debug(" IN: (%d) % x", len(streamObj.Stream), streamObj.Stream)
	common.Log.Debug("OUT: (%d) % x", len(outData), outData)

	if this.Predictor > 1 {
		if this.Predictor == 2 { // TIFF encoding: Needs some tests.
			common.Log.Debug("Tiff encoding")

			rowLength := int(this.Columns) * this.Colors
			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				common.Log.Debug("ERROR: TIFF encoding: Invalid row length...")
				return nil, fmt.Errorf("Invalid row length (%d/%d)", len(outData), rowLength)
			}

			if rowLength%this.Colors != 0 {
				return nil, fmt.Errorf("Invalid row length (%d) for colors %d", rowLength, this.Colors)
			}
			common.Log.Debug("inp outData (%d): % x", len(outData), outData)

			pOutBuffer := bytes.NewBuffer(nil)

			// 0-255  -255 255 ; 0-255=-255;
			for i := 0; i < rows; i++ {
				rowData := outData[rowLength*i : rowLength*(i+1)]
				//common.Log.Debug("RowData before: % d", rowData)
				// Predicts the same as the sample to the left.
				// Interleaved by colors.
				for j := this.Colors; j < rowLength; j++ {
					rowData[j] = byte(int(rowData[j]+rowData[j-this.Colors]) % 256)
				}
				// GH: Appears that this is not working as expected...

				pOutBuffer.Write(rowData)
			}
			pOutData := pOutBuffer.Bytes()
			common.Log.Debug("POutData (%d): % x", len(pOutData), pOutData)
			return pOutData, nil
		} else if this.Predictor >= 10 && this.Predictor <= 15 {
			common.Log.Debug("PNG Encoding")
			rowLength := int(this.Columns + 1) // 1 byte to specify predictor algorithms per row.
			rows := len(outData) / rowLength
			if len(outData)%rowLength != 0 {
				common.Log.Debug("ERROR: Invalid row length...")
				return nil, fmt.Errorf("Invalid row length (%d/%d)", len(outData), rowLength)
			}

			pOutBuffer := bytes.NewBuffer(nil)

			common.Log.Debug("Predictor columns: %d", this.Columns)
			common.Log.Debug("Length: %d / %d = %d rows", len(outData), rowLength, rows)
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
					return nil, fmt.Errorf("Invalid filter byte (%d)", fb)
				}

				for i := 0; i < rowLength; i++ {
					prevRowData[i] = rowData[i]
				}
				pOutBuffer.Write(rowData[1:])
			}
			pOutData := pOutBuffer.Bytes()
			return pOutData, nil
		} else {
			common.Log.Debug("ERROR: Unsupported predictor (%d)", this.Predictor)
			return nil, fmt.Errorf("Unsupported predictor (%d)", this.Predictor)
		}
	}

	return outData, nil
}

// Support for encoding LZW.  Currently not supporting predictors (raw compressed data only).
// Only supports the Early change = 1 algorithm (compress/lzw) as the other implementation
// does not have a write method.
// TODO: Consider refactoring compress/lzw to allow both.
func (this *LZWEncoder) EncodeBytes(data []byte) ([]byte, error) {
	if this.Predictor != 1 {
		return nil, fmt.Errorf("LZW Predictor = 1 only supported yet")
	}

	if this.EarlyChange == 1 {
		return nil, fmt.Errorf("LZW Early Change = 0 only supported yet")
	}

	var b bytes.Buffer
	w := lzw0.NewWriter(&b, lzw0.MSB, 8)
	w.Write(data)
	w.Close()

	return b.Bytes(), nil
}

/////
// ASCII hex encoder/decoder.
type ASCIIHexEncoder struct {
}

// Make a new ASCII hex encoder.
func NewASCIIHexEncoder() *ASCIIHexEncoder {
	encoder := &ASCIIHexEncoder{}
	return encoder
}

func (this *ASCIIHexEncoder) GetFilterName() string {
	return StreamEncodingFilterNameASCIIHex
}

func (this *ASCIIHexEncoder) MakeDecodeParams() PdfObject {
	return nil
}

// Make a new instance of an encoding dictionary for a stream object.
func (this *ASCIIHexEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := PdfObjectDictionary{}

	dict["Filter"] = MakeName(this.GetFilterName())
	return &dict
}

func (this *ASCIIHexEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	bufReader := bytes.NewReader(encoded)
	inb := []byte{}
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
			return nil, fmt.Errorf("Invalid ascii hex character (%c)", b)
		}
	}
	if len(inb)%2 == 1 {
		inb = append(inb, '0')
	}
	common.Log.Debug("Inbound %s", inb)
	outb := make([]byte, hex.DecodedLen(len(inb)))
	_, err := hex.Decode(outb, inb)
	if err != nil {
		return nil, err
	}
	return outb, nil
}

// ASCII hex decoding.
func (this *ASCIIHexEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return this.DecodeBytes(streamObj.Stream)
}

func (this *ASCIIHexEncoder) EncodeBytes(data []byte) ([]byte, error) {
	var encoded bytes.Buffer

	for _, b := range data {
		encoded.WriteString(fmt.Sprintf("%.2X ", b))
	}
	encoded.WriteByte('>')

	return encoded.Bytes(), nil
}

//
// ASCII85 encoder/decoder.
//
type ASCII85Encoder struct {
}

// Make a new ASCII85 encoder.
func NewASCII85Encoder() *ASCII85Encoder {
	encoder := &ASCII85Encoder{}
	return encoder
}

func (this *ASCII85Encoder) GetFilterName() string {
	return StreamEncodingFilterNameASCII85
}

func (this *ASCII85Encoder) MakeDecodeParams() PdfObject {
	return nil
}

// Make a new instance of an encoding dictionary for a stream object.
func (this *ASCII85Encoder) MakeStreamDict() *PdfObjectDictionary {
	dict := PdfObjectDictionary{}

	dict["Filter"] = MakeName(this.GetFilterName())
	return &dict
}

// 5 ASCII characters -> 4 raw binary bytes
func (this *ASCII85Encoder) DecodeBytes(encoded []byte) ([]byte, error) {
	decoded := []byte{}

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
				return nil, errors.New("Invalid code encountered")
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

	return decoded, nil
}

// ASCII85 stream decoding.
func (this *ASCII85Encoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return this.DecodeBytes(streamObj.Stream)
}

// Convert a base 256 number to a series of base 85 values (5 codes).
//  85^5 = 4437053125 > 256^4 = 4294967296
// So 5 base-85 numbers will always be enough to cover 4 base-256 numbers.
// The base 256 value is already converted to an uint32 value.
func (this *ASCII85Encoder) base256Tobase85(base256val uint32) [5]byte {
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

// Encode data into ASCII85 encoded format.
func (this *ASCII85Encoder) EncodeBytes(data []byte) ([]byte, error) {
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
			base85vals := this.base256Tobase85(base256)
			for _, val := range base85vals[:n+1] {
				encoded.WriteByte(val + '!')
			}
		}
	}

	// EOD.
	encoded.WriteString("~>")
	return encoded.Bytes(), nil
}

//
// Raw encoder/decoder (no encoding, pass through)
//
type RawEncoder struct{}

func NewRawEncoder() *RawEncoder {
	return &RawEncoder{}
}

func (this *RawEncoder) GetFilterName() string {
	return ""
}

func (this *RawEncoder) MakeDecodeParams() PdfObject {
	return nil
}

// Make a new instance of an encoding dictionary for a stream object.
func (this *RawEncoder) MakeStreamDict() *PdfObjectDictionary {
	return &PdfObjectDictionary{}
}

func (this *RawEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	return encoded, nil
}

func (this *RawEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return streamObj.Stream, nil
}

func (this *RawEncoder) EncodeBytes(data []byte) ([]byte, error) {
	return data, nil
}

//
// Multi encoder: support serial encoding.
//
type MultiEncoder struct {
	// Encoders in the order that they are to be applied.
	encoders []StreamEncoder
}

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
	decodeParamsArray := []PdfObject{}
	obj := (*encDict)["DecodeParms"]
	if obj != nil {
		// If it is a dictionary, assume it applies to all
		dict, isDict := obj.(*PdfObjectDictionary)
		if isDict {
			decodeParamsDict = dict
		}

		// If it is an array, assume there is one for each
		arr, isArray := obj.(*PdfObjectArray)
		if isArray {
			for _, dictObj := range *arr {
				if dict, is := dictObj.(*PdfObjectDictionary); is {
					decodeParamsArray = append(decodeParamsArray, dict)
				} else {
					decodeParamsArray = append(decodeParamsArray, nil)
				}
			}
		}
	}

	obj = (*encDict)["Filter"]
	if obj == nil {
		return nil, fmt.Errorf("Filter missing")
	}

	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Multi filter can only be made from array")
	}

	for idx, obj := range *array {
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Multi filter array element not a name")
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
					return nil, fmt.Errorf("Missing elements in decode params array")
				}
				dp = decodeParamsArray[idx]
			}
		}

		var dParams *PdfObjectDictionary
		if dict, is := dp.(*PdfObjectDictionary); is {
			dParams = dict
		}

		if *name == StreamEncodingFilterNameFlate {
			// XXX: need to separate out the DecodeParms..
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
		} else {
			common.Log.Error("Unsupported filter %s", *name)
			return nil, fmt.Errorf("Invalid filter in multi filter array")
		}
	}

	return mencoder, nil
}

func (this *MultiEncoder) GetFilterName() string {
	name := ""
	for idx, encoder := range this.encoders {
		name += encoder.GetFilterName()
		if idx < len(this.encoders)-1 {
			name += " "
		}
	}
	return name
}

func (this *MultiEncoder) MakeDecodeParams() PdfObject {
	if len(this.encoders) == 0 {
		return nil
	}

	if len(this.encoders) == 1 {
		return this.encoders[0].MakeDecodeParams()
	}

	array := PdfObjectArray{}
	for _, encoder := range this.encoders {
		decodeParams := encoder.MakeDecodeParams()
		if decodeParams == nil {
			array = append(array, MakeNull())
		} else {
			array = append(array, decodeParams)
		}
	}

	return &array
}

func (this *MultiEncoder) AddEncoder(encoder StreamEncoder) {
	this.encoders = append(this.encoders, encoder)
}

func (this *MultiEncoder) MakeStreamDict() *PdfObjectDictionary {
	dict := PdfObjectDictionary{}

	dict["Filter"] = MakeName(this.GetFilterName())

	// Pass all values from children, except Filter and DecodeParms.
	for _, encoder := range this.encoders {
		encDict := encoder.MakeStreamDict()
		for key, val := range *encDict {
			if key != "Filter" && key != "DecodeParms" {
				dict[key] = val
			}
		}
	}

	// Make the decode params array or dict.
	decodeParams := this.MakeDecodeParams()
	if decodeParams != nil {
		dict["DecodeParms"] = decodeParams
	}

	return &dict
}

func (this *MultiEncoder) DecodeBytes(encoded []byte) ([]byte, error) {
	decoded := encoded
	var err error
	// Apply in inverse order.
	for i := len(this.encoders) - 1; i >= 0; i-- {
		encoder := this.encoders[i]

		decoded, err = encoder.DecodeBytes(decoded)
		if err != nil {
			return nil, err
		}
	}

	return decoded, nil
}

func (this *MultiEncoder) DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	return this.DecodeBytes(streamObj.Stream)
}

func (this *MultiEncoder) EncodeBytes(data []byte) ([]byte, error) {
	encoded := data
	var err error

	// Apply in inverse order.
	for _, encoder := range this.encoders {
		encoded, err = encoder.EncodeBytes(encoded)
		if err != nil {
			return nil, err
		}
	}

	return encoded, nil
}
