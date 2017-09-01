/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"fmt"

	"bytes"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// TrueTypeFontEncoder handles text encoding for composite TrueType fonts.
// It performs mapping between character ids and glyph ids.
// It has a preloaded rune (unicode code point) to glyph index map that has been loaded from a font.
// Corresponds to Identity-H.
type TrueTypeFontEncoder struct {
	runeToGlyphIndexMap map[uint16]uint16
	cmap                CMap
}

// NewTrueTypeFontEncoder creates a new text encoder for TTF fonts with a pre-loaded runeToGlyphIndexMap,
// that has been pre-loaded from the font file.
// The new instance is preloaded with a CMapIdentityH (Identity-H) CMap which maps 2-byte charcodes to CIDs (glyph index).
func NewTrueTypeFontEncoder(runeToGlyphIndexMap map[uint16]uint16) TrueTypeFontEncoder {
	encoder := TrueTypeFontEncoder{}
	encoder.runeToGlyphIndexMap = runeToGlyphIndexMap
	encoder.cmap = CMapIdentityH{}
	return encoder
}

// Convert a raw utf8 string (encoded runes) to an encoded string (series of character codes) to be used in PDF.
func (enc TrueTypeFontEncoder) Encode(utf8 string) string {
	// runes -> character codes -> bytes
	var encoded bytes.Buffer
	for _, r := range utf8 {
		charcode, found := enc.RuneToCharcode(r)
		if !found {
			common.Log.Debug("Failed to map rune to charcode for rune 0x%X", r)
			continue
		}

		// Each entry represented by 2 bytes.
		encoded.WriteByte(byte((charcode & 0xff00) >> 8))
		encoded.WriteByte(byte(charcode & 0xff))
	}
	return encoded.String()
}

// Conversion between character code and glyph name.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) CharcodeToGlyph(code uint16) (string, bool) {
	rune, found := enc.CharcodeToRune(code)
	if found && rune == 0x20 {
		return "space", true
	}

	// Returns "uniXXXX" format where XXXX is the code in hex format.
	glyph := fmt.Sprintf("uni%X", code)
	return glyph, true
}

// Conversion between glyph name and character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) GlyphToCharcode(glyph string) (uint16, bool) {
	// String with "uniXXXX" format where XXXX is the hexcode.
	if len(glyph) == 7 && glyph[0:3] == "uni" {
		var unicode uint16
		n, err := fmt.Sscanf(glyph, "uni%X", &unicode)
		if n == 1 && err == nil {
			return enc.RuneToCharcode(rune(unicode))
		}
	}

	// Look in glyphlist.
	if rune, found := glyphlistGlyphToRuneMap[glyph]; found {
		return enc.RuneToCharcode(rune)
	}

	common.Log.Debug("Symbol encoding error: unable to find glyph->charcode entry (%s)", glyph)
	return 0, false
}

// Convert rune to character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) RuneToCharcode(r rune) (uint16, bool) {
	glyphIndex, has := enc.runeToGlyphIndexMap[uint16(r)]
	if !has {
		common.Log.Debug("Missing rune %d (0x%X) from encoding", r, r)
		return 0, false
	}
	// Identity : charcode <-> glyphIndex
	charcode := glyphIndex

	return uint16(charcode), true
}

// Convert character code to rune.
// The bool return flag is true if there was a match found, and false otherwise.
func (enc TrueTypeFontEncoder) CharcodeToRune(charcode uint16) (rune, bool) {
	// TODO: Make a reverse map stored.
	for c, glyphIndex := range enc.runeToGlyphIndexMap {
		if glyphIndex == charcode {
			return rune(c), true
		}
	}

	return 0, false
}

// Convert rune to glyph name.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) RuneToGlyph(val rune) (string, bool) {
	if val == 0x20 {
		return "space", true
	}
	glyph := fmt.Sprintf("uni%.4X", val)
	return glyph, true
}

// Convert glyph to rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) GlyphToRune(glyph string) (rune, bool) {
	// String with "uniXXXX" format where XXXX is the hexcode.
	if len(glyph) == 7 && glyph[0:3] == "uni" {
		var unicode uint16
		n, err := fmt.Sscanf(glyph, "uni%X", &unicode)
		if n == 1 && err == nil {
			return rune(unicode), true
		}
	}

	// Look in glyphlist.
	if rune, found := glyphlistGlyphToRuneMap[glyph]; found {
		return rune, true
	}

	return 0, false
}

// ToPdfObject returns a nil as it is not truly a PDF object and should not be attempted to store in file.
func (enc TrueTypeFontEncoder) ToPdfObject() core.PdfObject {
	return nil
}
