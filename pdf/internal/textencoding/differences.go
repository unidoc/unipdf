package textencoding

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// ApplyDifferences modifies or wraps the base encoding and overlays differences over it.
func ApplyDifferences(base SimpleEncoder, differences map[CharCode]GlyphName) SimpleEncoder {
	if enc, ok := base.(*simpleEncoder); ok {
		enc.applyDifferences(differences)
		return enc
	}
	return newDifferencesEncoding(base, differences)
}

func newDifferencesEncoding(base SimpleEncoder, differences map[CharCode]GlyphName) SimpleEncoder {
	// TODO(dennwc): check if it's a differencesEncoding, and merge the mapping
	d := &differencesEncoding{
		base:        base,
		differences: differences,
		code2rune:   make(map[CharCode]rune),
		rune2code:   make(map[rune]CharCode),
	}
	for code, glyph := range differences {
		r, ok := GlyphToRune(glyph)
		if ok {
			d.rune2code[r] = code
		} else {
			common.Log.Debug("ERROR: No match for glyph=%q differences=%+v", glyph, differences)
		}
		d.code2rune[code] = r
	}
	return d
}

// differencesEncoding remaps characters of a base encoding and act as a pass-trough for other characters.
type differencesEncoding struct {
	base SimpleEncoder

	// original mapping to encode to PDF
	differences map[CharCode]GlyphName
	// overlayed on top of base encoding
	code2rune map[CharCode]rune
	rune2code map[rune]CharCode
}

// BaseName returns base encoding name.
func (enc *differencesEncoding) BaseName() string {
	return enc.base.BaseName()
}

// String returns a string that describes the encoding.
func (enc *differencesEncoding) String() string {
	return fmt.Sprintf("differences(%s, %v)", enc.base.String(), enc.differences)
}

// Charcodes returns a slice of all charcodes in this encoding.
func (enc *differencesEncoding) Charcodes() []CharCode {
	codes := enc.base.Charcodes()
	sorted := true
	for _, code := range codes {
		if _, ok := enc.code2rune[code]; !ok {
			codes = append(codes, code)
			sorted = false
		}
	}
	if !sorted {
		sort.Slice(codes, func(i, j int) bool {
			return codes[i] < codes[j]
		})
	}
	return codes
}

// Encode converts a Go unicode string `raw` to a PDF encoded string.
func (enc *differencesEncoding) Encode(raw string) []byte {
	runes := []rune(raw)
	buf := bytes.NewBuffer(nil)
	buf.Grow(len(runes))
	for _, r := range runes {
		code, _ := enc.RuneToCharcode(r)
		// relies on the fact that underlying encoding is 8 bit
		buf.WriteByte(byte(code))
	}
	return buf.Bytes()
}

// RuneToCharcode returns the PDF character code corresponding to rune `r`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) RuneToCharcode(r rune) (CharCode, bool) {
	if code, ok := enc.rune2code[r]; ok {
		return code, true
	}
	return enc.base.RuneToCharcode(r)
}

// CharcodeToRune returns the rune corresponding to character code `code`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) CharcodeToRune(code CharCode) (rune, bool) {
	if r, ok := enc.code2rune[code]; ok {
		return r, true
	}
	return enc.base.CharcodeToRune(code)
}

// CharcodeToGlyph returns the glyph name for character code `code`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) CharcodeToGlyph(code CharCode) (GlyphName, bool) {
	if glyph, ok := enc.differences[code]; ok {
		return glyph, true
	}
	// TODO(dennwc): only redirects the call - remove from the interface
	r, ok := enc.CharcodeToRune(code)
	if !ok {
		return "", false
	}
	return enc.RuneToGlyph(r)
}

// GlyphToCharcode returns character code for glyph `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) GlyphToCharcode(glyph GlyphName) (CharCode, bool) {
	// TODO: store reverse map?
	for code, glyph2 := range enc.differences {
		if glyph2 == glyph {
			return code, true
		}
	}
	// TODO(dennwc): only redirects the call - remove from the interface
	r, ok := GlyphToRune(glyph)
	if !ok {
		return MissingCodeRune, false
	}
	return enc.RuneToCharcode(r)
}

// RuneToGlyph returns the glyph corresponding to rune `r`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) RuneToGlyph(r rune) (GlyphName, bool) {
	// TODO(dennwc): should be in the font interface
	return runeToGlyph(r, glyphlistRuneToGlyphMap)
}

// GlyphToRune returns the rune corresponding to glyph `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) GlyphToRune(glyph GlyphName) (rune, bool) {
	// TODO(dennwc): should be in the font interface
	return glyphToRune(glyph, glyphlistGlyphToRuneMap)
}

// ToPdfObject returns the encoding as a PdfObject.
func (enc *differencesEncoding) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()
	dict.Set("Type", core.MakeName("Encoding"))
	dict.Set("BaseEncoding", enc.base.ToPdfObject())
	dict.Set("Differences", toFontDifferences(enc.differences))
	return core.MakeIndirectObject(dict)
}
