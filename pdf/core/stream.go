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

	if _, isNull := filterObj.(*PdfObjectNull); isNull {
		// Filter is null -> raw data.
		return NewRawEncoder(), nil
	}

	// The filter should be a name or an array with a list of filter names.
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

			common.Log.Trace("Multi enc: %s\n", menc)
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
	} else if *method == StreamEncodingFilterNameLZW {
		return newLZWEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameDCT {
		return newDCTEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameASCIIHex {
		return NewASCIIHexEncoder(), nil
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
	common.Log.Trace("Decode stream")

	encoder, err := NewEncoderFromStream(streamObj)
	if err != nil {
		common.Log.Debug("Stream decoding failed: %v", err)
		return nil, err
	}
	common.Log.Trace("Encoder: %+v\n", encoder)

	decoded, err := encoder.DecodeStream(streamObj)
	if err != nil {
		common.Log.Debug("Stream decoding failed: %v", err)
		return nil, err
	}

	return decoded, nil
}

// Encodes the stream.
// Uses the encoding specified by the object.
func EncodeStream(streamObj *PdfObjectStream) error {
	common.Log.Trace("Encode stream")

	encoder, err := NewEncoderFromStream(streamObj)
	if err != nil {
		common.Log.Debug("Stream decoding failed: %v", err)
		return err
	}

	if lzwenc, is := encoder.(*LZWEncoder); is {
		// If LZW:
		// Make sure to use EarlyChange 0.. We do not have write support for 1 yet.
		lzwenc.EarlyChange = 0
		(*streamObj.PdfObjectDictionary)["EarlyChange"] = MakeInteger(0)
	}

	common.Log.Trace("Encoder: %+v\n", encoder)
	encoded, err := encoder.EncodeBytes(streamObj.Stream)
	if err != nil {
		common.Log.Debug("Stream encoding failed: %v", err)
		return err
	}

	streamObj.Stream = encoded

	// Update length
	(*streamObj.PdfObjectDictionary)["Length"] = MakeInteger(int64(len(encoded)))

	return nil
}
