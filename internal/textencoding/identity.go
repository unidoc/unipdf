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

// IdentityEncoder represents an 2-byte identity encoding
type IdentityEncoder struct {
	baseName string
}

// NewIdentityTextEncoder returns a new IdentityEncoder based on predefined
// encoding `baseName` and difference map `differences`.
func NewIdentityTextEncoder(baseName string) IdentityEncoder {
	return IdentityEncoder{baseName}
}

// String returns a string that describes `enc`.
func (enc IdentityEncoder) String() string {
	return enc.baseName
}

// Encode converts the Go unicode string to a PDF encoded string.
func (enc IdentityEncoder) Encode(str string) []byte {
	return encodeString16bit(enc, str)
}

// Decode converts PDF encoded string to a Go unicode string.
func (enc IdentityEncoder) Decode(raw []byte) string {
	return decodeString16bit(enc, raw)
}

// RuneToCharcode converts rune `r` to a PDF character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc IdentityEncoder) RuneToCharcode(r rune) (CharCode, bool) {
	return CharCode(r), true
}

// CharcodeToRune converts PDF character code `code` to a rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc IdentityEncoder) CharcodeToRune(code CharCode) (rune, bool) {
	return rune(code), true
}

// RuneToGlyph returns the glyph name for rune `r`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc IdentityEncoder) RuneToGlyph(r rune) (GlyphName, bool) {
	if r == ' ' {
		return "space", true
	}
	glyph := GlyphName(fmt.Sprintf("uni%.4X", r))
	return glyph, true
}

// GlyphToRune returns the rune corresponding to glyph name `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc IdentityEncoder) GlyphToRune(glyph GlyphName) (rune, bool) {
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
func (enc IdentityEncoder) ToPdfObject() core.PdfObject {
	if enc.baseName != "" {
		return core.MakeName(enc.baseName)
	}
	return core.MakeNull()
}
