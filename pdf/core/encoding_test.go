/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"encoding/base64"
	"testing"

	"github.com/unidoc/unidoc/common"
)

func init() {
	common.SetLogger(common.ConsoleLogger{})
}

// Test flate encoding.
func TestFlateEncoding(t *testing.T) {
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
