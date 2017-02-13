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
