/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// TextEncoder defines the common methods that a text encoder implementation must have in UniDoc.
type TextEncoder interface {
	// String returns a string that describes the TextEncoder instance.
	String() string

	// Encode converts the Go unicode string `raw` to a PDF encoded string.
	Encode(raw string) []byte

	// CharcodeToGlyph returns the glyph name for character code `code`.
	// The bool return flag is true if there was a match, and false otherwise.
	CharcodeToGlyph(code uint16) (string, bool)

	// GlyphToCharcode returns the PDF character code corresponding to glyph name `glyph`.
	// The bool return flag is true if there was a match, and false otherwise.
	GlyphToCharcode(glyph string) (uint16, bool)

	// RuneToCharcode returns the PDF character code corresponding to rune `r`.
	// The bool return flag is true if there was a match, and false otherwise.
	// This is usually implemented as RuneToGlyph->GlyphToCharcode
	RuneToCharcode(r rune) (uint16, bool)

	// CharcodeToRune returns the rune corresponding to character code `code`.
	// The bool return flag is true if there was a match, and false otherwise.
	// This is usually implemented as CharcodeToGlyph->GlyphToRune
	CharcodeToRune(code uint16) (rune, bool)

	// RuneToGlyph returns the glyph name for rune `r`.
	// The bool return flag is true if there was a match, and false otherwise.
	RuneToGlyph(r rune) (string, bool)

	// GlyphToRune returns the rune corresponding to glyph name `glyph`.
	// The bool return flag is true if there was a match, and false otherwise.
	GlyphToRune(glyph string) (rune, bool)

	// ToPdfObject returns a PDF Object that represents the encoding.
	ToPdfObject() core.PdfObject
}

// Convenience functions

// doEncode converts a Go unicode string `raw` to a PDF encoded string using the encoder `enc`.
func doEncode(enc TextEncoder, raw string) []byte {
	encoded := []byte{}
	for _, r := range raw {
		code, found := enc.RuneToCharcode(r)
		if !found {
			common.Log.Debug("Failed to map rune to charcode for rune 0x%04x", r)
			continue
		}
		encoded = append(encoded, byte(code))
	}
	return encoded
}

// doRuneToCharcode converts rune `r` to a PDF character code.
// The bool return flag is true if there was a match, and false otherwise.
func doRuneToCharcode(enc TextEncoder, r rune) (uint16, bool) {
	g, ok := enc.RuneToGlyph(r)
	if !ok {
		return 0, false
	}
	return enc.GlyphToCharcode(g)
}

// doCharcodeToRune converts PDF character code `code` to a rune.
// The bool return flag is true if there was a match, and false otherwise.
func doCharcodeToRune(enc TextEncoder, code uint16) (rune, bool) {
	g, ok := enc.CharcodeToGlyph(code)
	if !ok {
		return 0, false
	}
	return enc.GlyphToRune(g)
}
