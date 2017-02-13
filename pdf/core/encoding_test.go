/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
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
