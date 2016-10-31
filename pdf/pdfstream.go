/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"

	"github.com/unidoc/unidoc/common"
)

// Decodes the stream.
// Supports FlateDecode, ASCIIHexDecode.
func (this *PdfParser) decodeStream(obj *PdfObjectStream) ([]byte, error) {
	common.Log.Debug("Decode stream")
	common.Log.Debug("filter %s", (*obj).PdfObjectDictionary)

	filterObj, hasFilter := (*(obj.PdfObjectDictionary))["Filter"]
	if !hasFilter {
		// No filter, return raw data back.
		return obj.Stream, nil
	}

	// The filter should be a name or an array with a list of filter names.
	// Currently only supporting a single filter.
	method, ok := filterObj.(*PdfObjectName)
	if !ok {
		array, ok := filterObj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Filter not a Name or Array object")
		}
		if len(*array) != 1 {
			return nil, fmt.Errorf("Currently not supporting serial multi filter decoding")
		}
		filterObj = (*array)[0]
		method, ok = filterObj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Filter array member not a Name object")
		}
	}

	if *method == "FlateDecode" {
		// Refactor to a separate function.
		// Revamp this support to handle TIFF predictor (2).
		// Also handle more filter bytes and check
		// BitsPerComponent.  Default value is 8, currently we are only
		// supporting that one.
		predictor := 1

		decodeParams, hasDecodeParams := (*(obj.PdfObjectDictionary))["DecodeParms"].(*PdfObjectDictionary)
		if hasDecodeParams {
			common.Log.Debug("decode params: %s", decodeParams.String())
			predictor = int(*((*decodeParams)["Predictor"].(*PdfObjectInteger)))

			obits, hasbits := (*decodeParams)["BitsPerComponent"]
			if hasbits {
				pbits, ok := obits.(*PdfObjectInteger)
				if !ok {
					common.Log.Debug("ERROR: Invalid BitsPerComponent")
					return nil, fmt.Errorf("Invalid BitsPerComponent")
				}
				if *pbits != 8 {
					return nil, fmt.Errorf("Currently only 8 bits for flatedecode supported")
				}
			}
		}
		common.Log.Debug("Predictor: %d", predictor)

		common.Log.Debug("Encoding method: %s", method)

		bufReader := bytes.NewReader(obj.Stream)
		r, err := zlib.NewReader(bufReader)
		if err != nil {
			common.Log.Debug("Decoding error %v\n", err)
			common.Log.Debug("Stream (%d) % x", len(obj.Stream), obj.Stream)
			return nil, err
		}
		defer r.Close()

		var outBuf bytes.Buffer
		outBuf.ReadFrom(r)
		outData := outBuf.Bytes()

		if hasDecodeParams && predictor != 1 {
			if predictor == 2 { // TIFF encoding: Needs some tests.
				common.Log.Debug("Tiff encoding")

				columns, ok := (*decodeParams)["Columns"].(*PdfObjectInteger)
				if !ok {
					common.Log.Debug("ERROR: Predictor Column missing\n")
					return nil, fmt.Errorf("Predictor column missing")
				}

				colors := 1
				pcolors, hascolors := (*decodeParams)["Colors"].(*PdfObjectInteger)
				if hascolors {
					// Number of interleaved color components per sample
					colors = int(*pcolors)
				}
				common.Log.Debug("colors: %d", colors)

				rowLength := int(*columns) * colors
				rows := len(outData) / rowLength
				if len(outData)%rowLength != 0 {
					common.Log.Debug("ERROR: TIFF encoding: Invalid row length...")
					return nil, fmt.Errorf("Invalid row length (%d/%d)", len(outData), rowLength)
				}

				if rowLength%colors != 0 {
					return nil, fmt.Errorf("Invalid row length (%d) for colors %d", rowLength, colors)
				}
				common.Log.Debug("inp outData (%d): % x", len(outData), outData)

				pOutBuffer := bytes.NewBuffer(nil)

				// 0-255  -255 255 ; 0-255=-255;
				for i := 0; i < rows; i++ {
					rowData := outData[rowLength*i : rowLength*(i+1)]
					//common.Log.Debug("RowData before: % d", rowData)
					// Predicts the same as the sample to the left.
					// Interleaved by colors.
					for j := colors; j < rowLength; j++ {
						rowData[j] = byte(int(rowData[j]+rowData[j-colors]) % 256)
					}
					// GH: Appears that this is not working as expected...
					//common.Log.Debug("RowData after: % d", rowData)

					pOutBuffer.Write(rowData)
				}
				pOutData := pOutBuffer.Bytes()
				common.Log.Debug("POutData (%d): % x", len(pOutData), pOutData)
				return pOutData, nil
			} else if predictor >= 10 && predictor <= 15 {
				common.Log.Debug("PNG Encoding")
				columns, ok := (*decodeParams)["Columns"].(*PdfObjectInteger)
				if !ok {
					common.Log.Debug("ERROR: Predictor Column missing\n")
					return nil, fmt.Errorf("Predictor column missing")
				}
				rowLength := int(*columns + 1) // 1 byte to specify predictor algorithms per row.
				rows := len(outData) / rowLength
				if len(outData)%rowLength != 0 {
					common.Log.Debug("ERROR: Invalid row length...")
					return nil, fmt.Errorf("Invalid row length (%d/%d)", len(outData), rowLength)
				}

				pOutBuffer := bytes.NewBuffer(nil)

				common.Log.Debug("Predictor columns: %d", columns)
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
				common.Log.Debug("ERROR: Unsupported predictor (%d)", predictor)
				return nil, fmt.Errorf("Unsupported predictor (%d)", predictor)
			}
		}

		return outData, nil
	} else if *method == "ASCIIHexDecode" {
		bufReader := bytes.NewReader(obj.Stream)
		inb := []byte{}
		for {
			b, err := bufReader.ReadByte()
			if err != nil {
				return nil, err
			}
			if b == '>' {
				break
			}
			if isWhiteSpace(b) {
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

	common.Log.Debug("ERROR: Unsupported encoding method!")
	return nil, fmt.Errorf("Unsupported encoding method")
}
