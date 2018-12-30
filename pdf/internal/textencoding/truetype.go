/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"fmt"
	"sort"
	"strings"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// GID is a glyph index.
type GID uint16

// TODO(dennwc): should not mix Identity-H CMap and Encoding in the same object

// TrueTypeFontEncoder handles text encoding for composite TrueType fonts.
// It performs mapping between character ids and glyph ids.
// It has a preloaded rune (unicode code point) to glyph index map that has been loaded from a font.
// Corresponds to Identity-H CMap and Identity encoding.
type TrueTypeFontEncoder struct {
	runeToGIDMap map[rune]GID
}

// NewTrueTypeFontEncoder creates a new text encoder for TTF fonts with a runeToGlyphIndexMap that
// has been preloaded from the font file.
// The new instance is preloaded with a CMapIdentityH (Identity-H) CMap which maps 2-byte charcodes
// to CIDs (glyph index).
func NewTrueTypeFontEncoder(runeToGIDMap map[rune]GID) TrueTypeFontEncoder {
	return TrueTypeFontEncoder{
		runeToGIDMap: runeToGIDMap,
	}
}

// ttEncoderMaxNumEntries is the maximum number of encoding entries shown in simpleEncoder.String().
const ttEncoderMaxNumEntries = 10

// String returns a string that describes `enc`.
func (enc TrueTypeFontEncoder) String() string {
	parts := []string{
		fmt.Sprintf("%d entries", len(enc.runeToGIDMap)),
	}

	runes := make([]rune, 0, len(enc.runeToGIDMap))
	for r := range enc.runeToGIDMap {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	n := len(runes)
	if n > ttEncoderMaxNumEntries {
		n = ttEncoderMaxNumEntries
	}

	for i := 0; i < n; i++ {
		r := runes[i]
		parts = append(parts, fmt.Sprintf("%d=0x%02x: %q",
			r, r, enc.runeToGIDMap[r]))
	}
	return fmt.Sprintf("TRUETYPE_ENCODER{%s}", strings.Join(parts, ", "))
}

// Encode converts the Go unicode string `raw` to a PDF encoded string.
func (enc TrueTypeFontEncoder) Encode(raw string) []byte {
	return encodeString16bit(enc, raw)
}

// CharcodeToGlyph returns the glyph name matching character code `code`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) CharcodeToGlyph(code CharCode) (GlyphName, bool) {
	r, found := enc.CharcodeToRune(code)
	if found && r == ' ' {
		return "space", true
	}

	// Returns "uniXXXX" format where XXXX is the code in hex format.
	glyph := GlyphName(fmt.Sprintf("uni%.4X", code))
	return glyph, true
}

// GlyphToCharcode returns character code matching the glyph name `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) GlyphToCharcode(glyph GlyphName) (CharCode, bool) {
	// String with "uniXXXX" format where XXXX is the hexcode.
	if len(glyph) == 7 && glyph[0:3] == "uni" {
		var unicode uint16
		n, err := fmt.Sscanf(string(glyph), "uni%X", &unicode)
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
func (enc TrueTypeFontEncoder) RuneToCharcode(r rune) (CharCode, bool) {
	glyphIndex, ok := enc.runeToGIDMap[r]
	if !ok {
		common.Log.Debug("Missing rune %d (%+q) from encoding", r, r)
		return 0, false
	}
	// Identity : charcode <-> glyphIndex
	// TODO(dennwc): Here charcode is probably the same as CID.
	// TODO(dennwc): Find out what are the alternative mappings (enc.cmap?).
	charcode := CharCode(glyphIndex)

	return charcode, true
}

// CharcodeToRune converts PDF character code `code` to a rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) CharcodeToRune(code CharCode) (rune, bool) {
	// TODO: Make a reverse map stored.
	for r, gid := range enc.runeToGIDMap {
		// Identity : glyphIndex <-> charcode
		charcode := CharCode(gid)
		if charcode == code {
			return r, true
		}
	}
	common.Log.Debug("CharcodeToRune: No match. code=0x%04x enc=%s", code, enc)
	return 0, false
}

// RuneToGlyph returns the glyph name for rune `r`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) RuneToGlyph(r rune) (GlyphName, bool) {
	if r == ' ' {
		return "space", true
	}
	// TODO(dennwc): this is wrong; font may override this with a "post" table that specifies glyph names
	glyph := GlyphName(fmt.Sprintf("uni%.4X", r))
	return glyph, true
}

// GlyphToRune returns the rune corresponding to glyph name `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc TrueTypeFontEncoder) GlyphToRune(glyph GlyphName) (rune, bool) {
	// TODO(dennwc): this is wrong; font may override this with a "post" table that specifies glyph names
	// String with "uniXXXX" format where XXXX is the hexcode.
	if len(glyph) == 7 && glyph[0:3] == "uni" {
		unicode := uint16(0)
		n, err := fmt.Sscanf(string(glyph), "uni%X", &unicode)
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
	// TODO(dennwc): reasonable question: why it have to implement this interface then?
	return core.MakeNull()
}
