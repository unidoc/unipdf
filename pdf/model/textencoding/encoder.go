/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "github.com/unidoc/unidoc/pdf/core"

type TextEncoder interface {
	// Convert a raw utf8 string (series of runes) to an encoded string (series of character codes) to be used in PDF.
	Encode(raw string) string

	// Conversion between character code and glyph name.
	// The bool return flag is true if there was a match, and false otherwise.
	CharcodeToGlyph(code byte) (string, bool)

	// Conversion between glyph name and character code.
	// The bool return flag is true if there was a match, and false otherwise.
	GlyphToCharcode(glyph string) (byte, bool)

	// Convert rune to character code.
	// The bool return flag is true if there was a match, and false otherwise.
	RuneToCharcode(val rune) (byte, bool)

	// Convert character code to rune.
	// The bool return flag is true if there was a match, and false otherwise.
	CharcodeToRune(charcode byte) (rune, bool)

	// Convert rune to glyph name.
	// The bool return flag is true if there was a match, and false otherwise.
	RuneToGlyph(val rune) (string, bool)

	// Convert glyph to rune.
	// The bool return flag is true if there was a match, and false otherwise.
	GlyphToRune(glyph string) (rune, bool)

	ToPdfObject() core.PdfObject
}
