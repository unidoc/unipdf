package textencoding

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// FromFontDifferences converts `diffList` (a /Differences array from an /Encoding object) to a map
// representing character code to glyph mappings.
func FromFontDifferences(diffList *core.PdfObjectArray) (map[CharCode]GlyphName, error) {
	differences := make(map[CharCode]GlyphName)
	var n CharCode
	for _, obj := range diffList.Elements() {
		switch v := obj.(type) {
		case *core.PdfObjectInteger:
			n = CharCode(*v)
		case *core.PdfObjectName:
			s := string(*v)
			differences[n] = GlyphName(s)
			n++
		default:
			common.Log.Debug("ERROR: Bad type. obj=%s", obj)
			return nil, core.ErrTypeError
		}
	}
	return differences, nil
}

// toFontDifferences converts `differences` (a map representing character code to glyph mappings)
// to a /Differences array for an /Encoding object.
func toFontDifferences(differences map[CharCode]GlyphName) *core.PdfObjectArray {
	if len(differences) == 0 {
		return nil
	}

	codes := make([]CharCode, 0, len(differences))
	for c := range differences {
		codes = append(codes, c)
	}
	sort.Slice(codes, func(i, j int) bool {
		return codes[i] < codes[j]
	})

	n := codes[0]
	diffList := []core.PdfObject{core.MakeInteger(int64(n)), core.MakeName(string(differences[n]))}
	for _, c := range codes[1:] {
		if c == n+1 {
			diffList = append(diffList, core.MakeName(string(differences[c])))
		} else {
			diffList = append(diffList, core.MakeInteger(int64(c)))
		}
		n = c
	}
	return core.MakeArray(diffList...)
}

// ApplyDifferences modifies or wraps the base encoding and overlays differences over it.
func ApplyDifferences(base SimpleEncoder, differences map[CharCode]GlyphName) SimpleEncoder {
	// TODO(dennwc): check if it's a differencesEncoding, and merge the mapping
	d := &differencesEncoding{
		base:        base,
		differences: differences,
		decode:      make(map[byte]rune),
		encode:      make(map[rune]byte),
	}
	for code, glyph := range differences {
		b := byte(code)
		r, ok := GlyphToRune(glyph)
		if ok {
			d.encode[r] = b
		} else {
			common.Log.Debug("ERROR: No match for glyph=%q differences=%+v", glyph, differences)
		}
		d.decode[b] = r
	}
	return d
}

// differencesEncoding remaps characters of a base encoding and act as a pass-trough for other characters.
// Assumes that an underlying encoding is 8 bit.
type differencesEncoding struct {
	base SimpleEncoder

	// original mapping to encode to PDF
	differences map[CharCode]GlyphName

	// overlayed on top of base encoding (8 bit)
	decode map[byte]rune
	encode map[rune]byte
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
	seen := make(map[CharCode]struct{}, len(codes))
	for _, code := range codes {
		seen[code] = struct{}{}
	}
	for b := range enc.decode {
		code := CharCode(b)
		if _, ok := seen[code]; !ok {
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
	if b, ok := enc.encode[r]; ok {
		return CharCode(b), true
	}
	return enc.base.RuneToCharcode(r)
}

// CharcodeToRune returns the rune corresponding to character code `code`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) CharcodeToRune(code CharCode) (rune, bool) {
	if code > 0xff {
		return MissingCodeRune, false
	}
	b := byte(code)
	if r, ok := enc.decode[b]; ok {
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
	return enc.base.CharcodeToGlyph(code)
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
	code, ok := enc.RuneToCharcode(r)
	if !ok {
		return "", false
	}
	if glyph, ok := enc.differences[code]; ok {
		return glyph, true
	}
	return enc.base.RuneToGlyph(r)
}

// GlyphToRune returns the rune corresponding to glyph `glyph`.
// The bool return flag is true if there was a match, and false otherwise.
func (enc *differencesEncoding) GlyphToRune(glyph GlyphName) (rune, bool) {
	// TODO(dennwc): should be in the font interface
	code, ok := enc.GlyphToCharcode(glyph)
	if !ok {
		return MissingCodeRune, false
	}
	return enc.CharcodeToRune(code)
}

// ToPdfObject returns the encoding as a PdfObject.
func (enc *differencesEncoding) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()
	dict.Set("Type", core.MakeName("Encoding"))
	dict.Set("BaseEncoding", enc.base.ToPdfObject())
	dict.Set("Differences", toFontDifferences(enc.differences))
	return core.MakeIndirectObject(dict)
}
