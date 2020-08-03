/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/unidoc/unipdf/v3/core"
)

// IdentityEncoder represents an 2-byte identity encoding.
// NOTE: In many cases this is just used to encode/decode to glyph index and does not have a unicode
//  meaning, except via the ToUnicode maps.
// TODO: The use of runes as indicators for glyph indices and not-utf8 runes is not good and confusing.
//  Might be better to combine the Identity encoder with a ToUnicode map and keep track of the actual
//  runes and character codes, CMaps together.
type IdentityEncoder struct {
	baseName string

	// runes registered by encoder for tracking what runes are used for subsetting.
	registeredMap map[rune]struct{}
}

// NewIdentityTextEncoder returns a new IdentityEncoder based on predefined
// encoding `baseName` and difference map `differences`.
func NewIdentityTextEncoder(baseName string) *IdentityEncoder {
	return &IdentityEncoder{
		baseName: baseName,
	}
}

// RegisteredRunes returns the slice of runes that have been registered as used by the encoder.
func (enc *IdentityEncoder) RegisteredRunes() []rune {
	runes := make([]rune, len(enc.registeredMap))
	i := 0
	for r := range enc.registeredMap {
		runes[i] = r
		i++
	}
	return runes
}

// String returns a string that describes `enc`.
func (enc *IdentityEncoder) String() string {
	return enc.baseName
}

// Encode converts the Go unicode string to a PDF encoded string.
func (enc *IdentityEncoder) Encode(str string) []byte {
	return encodeString16bit(enc, str)
}

// Decode converts PDF encoded string to a Go unicode string.
func (enc *IdentityEncoder) Decode(raw []byte) string {
	return decodeString16bit(enc, raw)
}

// RuneToCharcode converts rune `r` to a PDF character code.
// The bool return flag is true if there was a match, and false otherwise.
// TODO: Here the `r` is an actual rune.
func (enc *IdentityEncoder) RuneToCharcode(r rune) (CharCode, bool) {
	if enc.registeredMap == nil {
		enc.registeredMap = map[rune]struct{}{}
	}
	enc.registeredMap[r] = struct{}{} // Register use (subsetting).

	return CharCode(r), true
}

// CharcodeToRune converts PDF character code `code` to a rune.
// The bool return flag is true if there was a match, and false otherwise.
// TODO: Here the `r` is not necessarily an actual rune but a glyph index (unless both).
func (enc *IdentityEncoder) CharcodeToRune(code CharCode) (rune, bool) {
	if enc.registeredMap == nil {
		enc.registeredMap = map[rune]struct{}{}
	}

	// TODO: The rune(code) is confusing and is not an actual utf8 rune.
	enc.registeredMap[rune(code)] = struct{}{}
	return rune(code), true
}

// RuneToGlyph returns the glyph name for rune `r`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *IdentityEncoder) RuneToGlyph(r rune) (GlyphName, bool) {
	if r == ' ' {
		return "space", true
	}
	glyph := GlyphName(fmt.Sprintf("uni%.4X", r))
	return glyph, true
}

// GlyphToRune returns the rune corresponding to glyph name `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *IdentityEncoder) GlyphToRune(glyph GlyphName) (rune, bool) {
	// String with "uniXXXX" format where XXXX is the hexcode.
	if glyph == "space" {
		return ' ', true
	} else if !strings.HasPrefix(string(glyph), "uni") || len(glyph) != 7 {
		return 0, false
	}
	r, err := strconv.ParseUint(string(glyph[3:]), 16, 16)
	if err != nil {
		return 0, false
	}
	return rune(r), true
}

// ToPdfObject returns a nil as it is not truly a PDF object and should not be attempted to store in file.
func (enc *IdentityEncoder) ToPdfObject() core.PdfObject {
	if enc.baseName != "" {
		return core.MakeName(enc.baseName)
	}
	return core.MakeNull()
}
