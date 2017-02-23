/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"fmt"

	"github.com/unidoc/unidoc/common"
)

// Creates the encoder from the stream's dictionary.
func NewEncoderFromStream(streamObj *PdfObjectStream) (StreamEncoder, error) {
	filterObj, hasFilter := (*(streamObj.PdfObjectDictionary))["Filter"]
	if !hasFilter {
		// No filter, return raw data back.
		return NewRawEncoder(), nil
	}

	// The filter should be a name or an array with a list of filter names.
	// Currently only supporting a single filter.
	method, ok := filterObj.(*PdfObjectName)
	if !ok {
		array, ok := filterObj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Filter not a Name or Array object")
		}
		if len(*array) == 0 {
			// Empty array -> indicates raw filter (no filter).
			return NewRawEncoder(), nil
		}

		if len(*array) != 1 {
			menc, err := newMultiEncoderFromStream(streamObj)
			if err != nil {
				common.Log.Error("Failed creating multi encoder: %v", err)
				return nil, err
			}

			common.Log.Debug("Multi enc: %s\n", menc)
			return menc, nil
		}

		// Single element.
		filterObj = (*array)[0]
		method, ok = filterObj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Filter array member not a Name object")
		}
	}

	if *method == StreamEncodingFilterNameFlate {
		return newFlateEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameASCIIHex {
		return NewASCIIHexEncoder(), nil
	} else if *method == StreamEncodingFilterNameLZW {
		return newLZWEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameASCII85 {
		return NewASCII85Encoder(), nil
	} else {
		common.Log.Debug("ERROR: Unsupported encoding method!")
		return nil, fmt.Errorf("Unsupported encoding method (%s)", *method)
	}
}

// Decodes the stream.
// Supports FlateDecode, ASCIIHexDecode, LZW.
func DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	common.Log.Debug("Decode stream")

	encoder, err := NewEncoderFromStream(streamObj)
	if err != nil {
		common.Log.Debug("Stream decoding failed: %v", err)
		return nil, err
	}

	common.Log.Debug("Encoder: %+v\n", encoder)

	decoded, err := encoder.DecodeStream(streamObj)
	if err != nil {
		common.Log.Debug("Stream decoding failed: %v", err)
		return nil, err
	}

	return decoded, nil
}
