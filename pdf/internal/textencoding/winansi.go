/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"bytes"

	"github.com/unidoc/unidoc/pdf/core"
	"golang.org/x/text/encoding/charmap"
)

const baseWinAnsi = "WinAnsiEncoding"

// NewWinAnsiTextEncoder returns a simpleEncoder that implements WinAnsiEncoding.
func NewWinAnsiTextEncoder() SimpleEncoder {
	return &charmapEncoding{
		baseName: baseWinAnsi,
		charmap:  charmap.Windows1252,
	}
}

var _ SimpleEncoder = (*charmapEncoding)(nil)

type charmapEncoding struct {
	baseName string
	charmap  *charmap.Charmap
}

// String returns a text representation of encoding.
func (enc *charmapEncoding) String() string {
	return "charmapEncoding(" + enc.baseName + ")"
}

// BaseName returns a base name of the encoder, as specified in the PDF spec.
func (enc *charmapEncoding) BaseName() string {
	return enc.baseName
}

// Encode converts a Go unicode string `raw` to a PDF encoded string.
func (enc *charmapEncoding) Encode(raw string) []byte {
	runes := []rune(raw)
	buf := bytes.NewBuffer(nil)
	buf.Grow(len(runes))
	for _, r := range runes {
		b, ok := enc.charmap.EncodeRune(r)
		if !ok {
			b, _ = enc.charmap.EncodeRune(MissingCodeRune)
		}
		buf.WriteByte(b)
	}
	return buf.Bytes()
}

func (enc *charmapEncoding) Charcodes() []CharCode {
	codes := make([]CharCode, 0, 256)
	for i := 0; i < 256; i++ {
		code := CharCode(i)
		if _, ok := enc.CharcodeToRune(code); ok {
			codes = append(codes, code)
		}
	}
	return codes
}

func (enc *charmapEncoding) RuneToCharcode(r rune) (CharCode, bool) {
	b, ok := enc.charmap.EncodeRune(r)
	return CharCode(b), ok
}

func (enc *charmapEncoding) CharcodeToRune(code CharCode) (rune, bool) {
	if code > 0xff {
		return MissingCodeRune, false
	}
	switch enc.baseName {
	case "WinAnsiEncoding":
		// WinANSI in the old implementation remaps few characters

		// everything below 20 (space) is "missing"
		if code < 0x20 {
			return MissingCodeRune, false
		}

		const bullet = '•'
		switch code {

		// in WinAnsiEncoding all unused and non-visual codes map to the '•' character
		case 127: // DEL
			return bullet, true
		case 129, 141, 143, 144, 157: // unused in WinANSI
			return bullet, true

		// typographically similar
		case 160: // non-breaking space -> space
			return ' ', true
		case 173: // soft hyphen -> hyphen
			return '-', true
		}
	}
	r := enc.charmap.DecodeByte(byte(code))
	return r, r != MissingCodeRune
}

func (enc *charmapEncoding) CharcodeToGlyph(code CharCode) (GlyphName, bool) {
	// TODO(dennwc): only redirects the call - remove from the interface
	r, ok := enc.CharcodeToRune(code)
	if !ok {
		return "", false
	}
	return enc.RuneToGlyph(r)
}

func (enc *charmapEncoding) GlyphToCharcode(glyph GlyphName) (CharCode, bool) {
	// TODO(dennwc): only redirects the call - remove from the interface
	r, ok := GlyphToRune(glyph)
	if !ok {
		return MissingCodeRune, false
	}
	return enc.RuneToCharcode(r)
}

func (enc *charmapEncoding) RuneToGlyph(r rune) (GlyphName, bool) {
	// TODO(dennwc): should be in the font interface
	return runeToGlyph(r, glyphlistRuneToGlyphMap)
}

func (enc *charmapEncoding) GlyphToRune(glyph GlyphName) (rune, bool) {
	// TODO(dennwc): should be in the font interface
	return glyphToRune(glyph, glyphlistGlyphToRuneMap)
}

func (enc *charmapEncoding) ToPdfObject() core.PdfObject {
	return core.MakeName(enc.baseName)
}
