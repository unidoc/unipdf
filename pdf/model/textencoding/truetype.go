/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

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

// NewTrueTypeFontEncoder creates a new text encoder for TTF fonts with a runeToGlyphIndexMap that
// has been preloaded from the font file.
// The new instance is preloaded with a CMapIdentityH (Identity-H) CMap which maps 2-byte charcodes
// to CIDs (glyph index).
func NewTrueTypeFontEncoder(runeToGlyphIndexMap map[uint16]uint16) TrueTypeFontEncoder {
	return TrueTypeFontEncoder{
		runeToGlyphIndexMap: runeToGlyphIndexMap,
		cmap:                CMapIdentityH{},
	}
}

// ttEncoderMaxNumEntries is the maximum number of encoding entries shown in SimpleEncoder.String().
const ttEncoderMaxNumEntries = 10

// String returns a string that describes `enc`.
func (enc TrueTypeFontEncoder) String() string {
	parts := []string{
		fmt.Sprintf("%d entries", len(enc.runeToGlyphIndexMap)),
	}

	codes := []int{}
	for c := range enc.runeToGlyphIndexMap {
		codes = append(codes, int(c))
	}
	sort.Ints(codes)
	numCodes := len(codes)
	if numCodes > ttEncoderMaxNumEntries {
		numCodes = ttEncoderMaxNumEntries
	}

	for i := 0; i < numCodes; i++ {
		c := codes[i]
		parts = append(parts, fmt.Sprintf("%d=0x%02x: %q",
			c, c, enc.runeToGlyphIndexMap[uint16(c)]))
	}
	return fmt.Sprintf("TRUETYPE_ENCODER{%s}", strings.Join(parts, ", "))
}

// Encode converts the Go unicode string `raw` to a PDF encoded string.
func (enc TrueTypeFontEncoder) Encode(raw string) []byte {
	// runes -> character codes -> bytes
	var encoded bytes.Buffer
	for _, r := range raw {
		code, ok := enc.RuneToCharcode(r)
		if !ok {
			common.Log.Debug("Failed to map rune to charcode. rune=%+q", r)
			continue
		}

		// Each entry represented by 2 bytes.
		encoded.WriteByte(byte((code & 0xff00) >> 8))
		encoded.WriteByte(byte(code & 0xff))
	}
	return encoded.Bytes()
}

// CharcodeToGlyph returns the glyph name matching character code `code`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) CharcodeToGlyph(code uint16) (string, bool) {
	r, found := enc.CharcodeToRune(code)
	if found && r == 0x20 {
		return "space", true
	}

	// Returns "uniXXXX" format where XXXX is the code in hex format.
	glyph := fmt.Sprintf("uni%.4X", code)
	return glyph, true
}

// GlyphToCharcode returns character code matching the glyph name `glyph`.
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

// RuneToCharcode converts rune `r` to a PDF character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) RuneToCharcode(r rune) (uint16, bool) {
	glyphIndex, ok := enc.runeToGlyphIndexMap[uint16(r)]
	if !ok {
		common.Log.Debug("Missing rune %d (%+q) from encoding", r, r)
		return 0, false
	}
	// Identity : charcode <-> glyphIndex
	charcode := glyphIndex

	return uint16(charcode), true
}

// CharcodeToRune converts PDF character code `code` to a rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) CharcodeToRune(code uint16) (rune, bool) {
	// TODO: Make a reverse map stored.
	for code, glyphIndex := range enc.runeToGlyphIndexMap {
		if glyphIndex == code {
			return rune(code), true
		}
	}
	common.Log.Debug("CharcodeToRune: No match. code=0x%04x enc=%s", code, enc)
	return 0, false
}

// RuneToGlyph returns the glyph name for rune `r`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) RuneToGlyph(r rune) (string, bool) {
	if r == 0x20 {
		return "space", true
	}
	glyph := fmt.Sprintf("uni%.4X", r)
	return glyph, true
}

// GlyphToRune returns the rune corresponding to glyph name `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) GlyphToRune(glyph string) (rune, bool) {
	// String with "uniXXXX" format where XXXX is the hexcode.
	if len(glyph) == 7 && glyph[0:3] == "uni" {
		unicode := uint16(0)
		n, err := fmt.Sscanf(glyph, "uni%X", &unicode)
		if n == 1 && err == nil {
			return rune(unicode), true
		}
	}

	// Look in glyphlist.
	if r, ok := glyphlistGlyphToRuneMap[glyph]; ok {
		return r, true
	}

	return 0, false
}

// ToPdfObject returns a nil as it is not truly a PDF object and should not be attempted to store in file.
func (enc TrueTypeFontEncoder) ToPdfObject() core.PdfObject {
	return core.MakeNull()
}
