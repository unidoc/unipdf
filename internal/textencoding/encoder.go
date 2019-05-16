/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"encoding/binary"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// CharCode is a character code used in the specific encoding.
type CharCode uint16

// GlyphName is a name of a glyph.
type GlyphName string

// TextEncoder defines the common methods that a text encoder implementation must have in UniDoc.
type TextEncoder interface {
	// String returns a string that describes the TextEncoder instance.
	String() string

	// Encode converts the Go unicode string to a PDF encoded string.
	Encode(str string) []byte

	// Decode converts PDF encoded string to a Go unicode string.
	Decode(raw []byte) string

	// RuneToCharcode returns the PDF character code corresponding to rune `r`.
	// The bool return flag is true if there was a match, and false otherwise.
	// This is usually implemented as RuneToGlyph->GlyphToCharcode
	RuneToCharcode(r rune) (CharCode, bool)

	// CharcodeToRune returns the rune corresponding to character code `code`.
	// The bool return flag is true if there was a match, and false otherwise.
	// This is usually implemented as CharcodeToGlyph->GlyphToRune
	CharcodeToRune(code CharCode) (rune, bool)

	// ToPdfObject returns a PDF Object that represents the encoding.
	ToPdfObject() core.PdfObject
}

// Convenience functions

// encodeString8bit converts a Go unicode string `raw` to a PDF encoded string using the encoder `enc`.
// It expects that character codes will fit into a single byte.
func encodeString8bit(enc TextEncoder, raw string) []byte {
	encoded := make([]byte, 0, len(raw))
	for _, r := range raw {
		code, found := enc.RuneToCharcode(r)
		if !found || code > 0xff {
			common.Log.Debug("Failed to map rune to charcode for rune 0x%04x", r)
			continue
		}
		encoded = append(encoded, byte(code))
	}
	return encoded
}

// encodeString16bit converts a Go unicode string `raw` to a PDF encoded string using the encoder `enc`.
// Each character will be encoded as two bytes.
func encodeString16bit(enc TextEncoder, raw string) []byte {
	// runes -> character codes -> bytes
	runes := []rune(raw)
	encoded := make([]byte, 0, len(runes)*2)
	for _, r := range runes {
		code, ok := enc.RuneToCharcode(r)
		if !ok {
			common.Log.Debug("Failed to map rune to charcode. rune=%+q", r)
			continue
		}

		// Each entry represented by 2 bytes.
		var v [2]byte
		binary.BigEndian.PutUint16(v[:], uint16(code))
		encoded = append(encoded, v[:]...)
	}
	return encoded
}

// decodeString16bit converts PDF encoded string to a Go unicode string using the encoder `enc`.
// Each character will be decoded from two bytes.
func decodeString16bit(enc TextEncoder, raw []byte) string {
	// bytes -> character codes -> runes
	runes := make([]rune, 0, len(raw)/2+len(raw)%2)

	for len(raw) > 0 {
		if len(raw) == 1 {
			raw = []byte{raw[0], 0}
		}
		// Each entry represented by 2 bytes.
		code := CharCode(binary.BigEndian.Uint16(raw[:]))
		raw = raw[2:]

		r, ok := enc.CharcodeToRune(code)
		if !ok {
			common.Log.Debug("Failed to map charcode to rune. charcode=%#x", code)
			continue
		}
		runes = append(runes, r)
	}
	return string(runes)
}
