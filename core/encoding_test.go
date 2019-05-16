/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"encoding/base64"
	"testing"

	"github.com/unidoc/unipdf/v3/common"
)

func init() {
	common.SetLogger(common.ConsoleLogger{})
}

// Test flate encoding - Predictor 1.
func TestFlateEncodingPredictor1(t *testing.T) {
	rawStream := []byte("this is a dummy text with some \x01\x02\x03 binary data")

	encoder := NewFlateEncoder()
	encoder.Predictor = 1

	encoded, err := encoder.EncodeBytes(rawStream)
	if err != nil {
		t.Errorf("Failed to encode data: %v", err)
		return
	}

	decoded, err := encoder.DecodeBytes(encoded)
	if err != nil {
		t.Errorf("Failed to decode data: %v", err)
		return
	}

	if !compareSlices(decoded, rawStream) {
		t.Errorf("Slices not matching")
		t.Errorf("Decoded (%d): % x", len(encoded), encoded)
		t.Errorf("Raw     (%d): % x", len(rawStream), rawStream)
		return
	}
}

// Test post decoding predictors.
func TestPostDecodingPredictors(t *testing.T) {

	testcases := []struct {
		BitsPerComponent int
		Colors           int
		Columns          int
		Predictor        int
		Input            []byte
		Expected         []byte
	}{
		// BPC=8, Colors=3, PNG predictor = none.
		{
			BitsPerComponent: 8,
			Colors:           3,
			Columns:          3,
			Predictor:        15,
			Input: []byte{
				pfNone, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 1.
				pfNone, 3, 2, 1, 3, 2, 1, 3, 2, 1, // Row 2.
				pfNone, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 3.
			},
			Expected: []byte{
				1, 2, 3, 1, 2, 3, 1, 2, 3,
				3, 2, 1, 3, 2, 1, 3, 2, 1,
				1, 2, 3, 1, 2, 3, 1, 2, 3,
			},
		},
		// BPC=8, Colors=3, PNG predictor = sub (same as left).
		{
			BitsPerComponent: 8,
			Colors:           3,
			Columns:          3,
			Predictor:        15,
			Input: []byte{
				pfSub, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 1.
				pfSub, 3, 2, 1, 3, 2, 1, 3, 2, 1, // Row 2.
				pfSub, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 3.
			},
			Expected: []byte{
				1, 2, 3, 1 + 1, 2 + 2, 3 + 3, 1 + 1 + 1, 2 + 2 + 2, 3 + 3 + 3,
				3, 2, 1, 3 + 3, 2 + 2, 1 + 1, 3 + 3 + 3, 2 + 2 + 2, 1 + 1 + 1,
				1, 2, 3, 1 + 1, 2 + 2, 3 + 3, 1 + 1 + 1, 2 + 2 + 2, 3 + 3 + 3,
			},
		},
		// BPC=8, Colors=3, PNG predictor = up (same as above).
		{
			BitsPerComponent: 8,
			Colors:           3,
			Columns:          3,
			Predictor:        15,
			Input: []byte{
				pfUp, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 1.
				pfUp, 3, 2, 1, 3, 2, 1, 3, 2, 1, // Row 2.
				pfUp, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 3.
			},
			Expected: []byte{
				1, 2, 3, 1, 2, 3, 1, 2, 3,
				3 + 1, 2 + 2, 1 + 3, 3 + 1, 2 + 2, 1 + 3, 3 + 1, 2 + 2, 1 + 3,
				1 + 3 + 1, 2 + 2 + 2, 3 + 1 + 3, 1 + 3 + 1, 2 + 2 + 2, 3 + 1 + 3, 1 + 3 + 1, 2 + 2 + 2, 3 + 1 + 3,
			},
		},
		// BPC=8, Colors=3, PNG predictor = avg (average of left and above).
		// Use a spreadsheet to get expected values.
		{
			BitsPerComponent: 8,
			Colors:           3,
			Columns:          3,
			Predictor:        15,
			Input: []byte{
				pfAvg, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 1.
				pfAvg, 3, 2, 1, 3, 2, 1, 3, 2, 1, // Row 2.
				pfAvg, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 3.
			},
			Expected: []byte{
				1, 2, 3, 1, 3, 4, 1, 3, 5,
				3, 3, 2, 5, 5, 4, 6, 6, 5,
				2, 3, 4, 4, 6, 7, 6, 8, 9,
			},
		},
		// BPC=8, Colors=3, PNG predictor = Paeth.
		// Use a spreadsheet to get expected values.
		{
			BitsPerComponent: 8,
			Colors:           3,
			Columns:          3,
			Predictor:        15,
			Input: []byte{
				pfPaeth, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 1.
				pfPaeth, 3, 2, 1, 3, 2, 1, 3, 2, 1, // Row 2.
				pfPaeth, 1, 2, 3, 1, 2, 3, 1, 2, 3, // Row 3.
			},
			Expected: []byte{
				1, 2, 3, 2, 4, 6, 3, 6, 9,
				4, 4, 4, 7, 6, 7, 10, 8, 10,
				5, 6, 7, 8, 8, 10, 11, 10, 13,
			},
		},
	}

	for i, tcase := range testcases {
		encoder := &FlateEncoder{
			BitsPerComponent: tcase.BitsPerComponent,
			Colors:           tcase.Colors,
			Columns:          tcase.Columns,
			Predictor:        tcase.Predictor,
		}

		predicted, err := encoder.postDecodePredict(tcase.Input)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		t.Logf("%d: % d\n", i, predicted)
		if !compareSlices(predicted, tcase.Expected) {
			t.Errorf("Slices not matching (i = %d)", i)
			t.Errorf("Predicted (%d): % d", len(predicted), predicted)
			t.Fatalf("Expected  (%d): % d", len(tcase.Expected), tcase.Expected)
		}
	}
}

// Test LZW encoding.
func TestLZWEncoding(t *testing.T) {
	rawStream := []byte("this is a dummy text with some \x01\x02\x03 binary data")

	encoder := NewLZWEncoder()
	// Only supporitng early change 0 for encoding at the moment.
	encoder.EarlyChange = 0

	encoded, err := encoder.EncodeBytes(rawStream)
	if err != nil {
		t.Errorf("Failed to encode data: %v", err)
		return
	}

	decoded, err := encoder.DecodeBytes(encoded)
	if err != nil {
		t.Errorf("Failed to decode data: %v", err)
		return
	}

	if !compareSlices(decoded, rawStream) {
		t.Errorf("Slices not matching")
		t.Errorf("Decoded (%d): % x", len(encoded), encoded)
		t.Errorf("Raw     (%d): % x", len(rawStream), rawStream)
		return
	}
}

// Test run length encoding.
func TestRunLengthEncoding(t *testing.T) {
	rawStream := []byte("this is a dummy text with some \x01\x02\x03 binary data")
	encoder := NewRunLengthEncoder()
	encoded, err := encoder.EncodeBytes(rawStream)
	if err != nil {
		t.Errorf("Failed to RunLength encode data: %v", err)
		return
	}
	decoded, err := encoder.DecodeBytes(encoded)
	if err != nil {
		t.Errorf("Failed to RunLength decode data: %v", err)
		return
	}
	if !compareSlices(decoded, rawStream) {
		t.Errorf("Slices not matching. RunLength")
		t.Errorf("Decoded (%d): % x", len(encoded), encoded)
		t.Errorf("Raw     (%d): % x", len(rawStream), rawStream)
		return
	}
}

// Test ASCII hex encoding.
func TestASCIIHexEncoding(t *testing.T) {
	byteData := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	expected := []byte("DE AD BE EF >")

	encoder := NewASCIIHexEncoder()
	encoded, err := encoder.EncodeBytes(byteData)
	if err != nil {
		t.Errorf("Failed to encode data: %v", err)
		return
	}

	if !compareSlices(encoded, expected) {
		t.Errorf("Slices not matching")
		t.Errorf("Expected (%d): %s", len(expected), expected)
		t.Errorf("Encoded  (%d): %s", len(encoded), encoded)
		return
	}
}

// ASCII85.
func TestASCII85EncodingWikipediaExample(t *testing.T) {
	expected := `Man is distinguished, not only by his reason, but by this singular passion from other animals, which is a lust of the mind, that by a perseverance of delight in the continued and indefatigable generation of knowledge, exceeds the short vehemence of any carnal pleasure.`
	// Base 64 encoded, Ascii85 encoded version (wikipedia).
	encodedInBase64 := `OWpxb15CbGJELUJsZUIxREorKitGKGYscS8wSmhLRjxHTD5DakAuNEdwJGQ3RiEsTDdAPDZAKS8wSkRFRjxHJTwrRVY6MkYhLE88REorKi5APCpLMEA8NkwoRGYtXDBFYzVlO0RmZlooRVplZS5CbC45cEYiQUdYQlBDc2krREdtPkAzQkIvRiomT0NBZnUyL0FLWWkoREliOkBGRCwqKStDXVU9QDNCTiNFY1lmOEFURDNzQHE/ZCRBZnRWcUNoW05xRjxHOjgrRVY6LitDZj4tRkQ1VzhBUmxvbERJYWwoRElkPGpAPD8zckA6RiVhK0Q1OCdBVEQ0JEJsQGwzRGU6LC1ESnNgOEFSb0ZiLzBKTUtAcUI0XkYhLFI8QUtaJi1EZlRxQkclRz51RC5SVHBBS1lvJytDVC81K0NlaSNESUk/KEUsOSlvRioyTTcvY34+`
	encoded, _ := base64.StdEncoding.DecodeString(encodedInBase64)

	encoder := NewASCII85Encoder()
	enc1, err := encoder.EncodeBytes([]byte(expected))
	if err != nil {
		t.Errorf("Fail")
		return
	}
	if string(enc1) != string(encoded) {
		t.Errorf("ASCII85 encoding wiki example fail")
		return
	}

	decoded, err := encoder.DecodeBytes([]byte(encoded))
	if err != nil {
		t.Errorf("Fail, error: %v", err)
		return
	}
	if expected != string(decoded) {
		t.Errorf("Mismatch! '%s' vs '%s'", decoded, expected)
		return
	}
}

func TestASCII85Encoding(t *testing.T) {
	encoded := `FD,B0+EVmJAKYo'+D#G#De*R"B-:o0+E_a:A0>T(+AbuZ@;]Tu:ddbqAnc'mEr~>`
	expected := "this type of encoding is used in PS and PDF files"

	encoder := NewASCII85Encoder()

	enc1, err := encoder.EncodeBytes([]byte(expected))
	if err != nil {
		t.Errorf("Fail")
		return
	}
	if encoded != string(enc1) {
		t.Errorf("Encoding error")
		return
	}

	decoded, err := encoder.DecodeBytes([]byte(encoded))
	if err != nil {
		t.Errorf("Fail, error: %v", err)
		return
	}
	if expected != string(decoded) {
		t.Errorf("Mismatch! '%s' vs '%s'", decoded, expected)
		return
	}
}

type TestASCII85DecodingTestCase struct {
	Encoded  string
	Expected string
}

func TestASCII85Decoding(t *testing.T) {
	// Map encoded -> Decoded
	testcases := []TestASCII85DecodingTestCase{
		{"z~>", "\x00\x00\x00\x00"},
		{"z ~>", "\x00\x00\x00\x00"},
		{"zz~>", "\x00\x00\x00\x00\x00\x00\x00\x00"},
		{" zz~>", "\x00\x00\x00\x00\x00\x00\x00\x00"},
		{" z z~>", "\x00\x00\x00\x00\x00\x00\x00\x00"},
		{" z z ~>", "\x00\x00\x00\x00\x00\x00\x00\x00"},
		{"+T~>", `!`},
		{"+`d~>", `!s`},
		{"+`hr~>", `!sz`},
		{"+`hsS~>", `!szx`},
		{"+`hsS+T~>", `!szx!`},
		{"+ `hs S +T ~>", `!szx!`},
	}

	encoder := NewASCII85Encoder()

	for _, testcase := range testcases {
		encoded := testcase.Encoded
		expected := testcase.Expected
		decoded, err := encoder.DecodeBytes([]byte(encoded))
		if err != nil {
			t.Errorf("Fail, error: %v", err)
			return
		}
		if expected != string(decoded) {
			t.Errorf("Mismatch! '%s' vs '%s'", decoded, expected)
			return
		}
	}
}

// Test multi encoder with FlateDecode and ASCIIHexDecode.
func TestMultiEncoder(t *testing.T) {
	rawStream := []byte("this is a dummy text with some \x01\x02\x03 binary data")

	encoder := NewMultiEncoder()

	enc1 := NewFlateEncoder()
	enc1.Predictor = 1
	encoder.AddEncoder(enc1)

	enc2 := NewASCIIHexEncoder()
	encoder.AddEncoder(enc2)

	encoded, err := encoder.EncodeBytes(rawStream)
	if err != nil {
		t.Errorf("Failed to encode data: %v", err)
		return
	}
	common.Log.Debug("Multi Encoded: %s", encoded)
	// Multi Encoded: 78 9C 2A C9 C8 2C 56 00 A2 44 85 94 D2 DC DC 4A  85 92 D4 8A 12 85 F2
	// CC 92 0C 85 E2 FC DC 54 05 46 26 66 85 A4CC BC C4 A2 4A 85 94 C4 92 44 40 00 00 00
	// FF FF 78 87 0F 9C >
	// Looks fine..

	decoded, err := encoder.DecodeBytes(encoded)
	if err != nil {
		t.Errorf("Failed to decode data: %v", err)
		return
	}

	if !compareSlices(decoded, rawStream) {
		t.Errorf("Slices not matching")
		t.Errorf("Decoded (%d): % x", len(encoded), encoded)
		t.Errorf("Raw     (%d): % x", len(rawStream), rawStream)
		return
	}
}
