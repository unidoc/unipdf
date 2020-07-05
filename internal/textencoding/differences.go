package textencoding

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
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
	if len(differences) == 0 {
		return base
	}
	d := &differencesEncoding{
		base:        base,
		differences: differences,
		decode:      make(map[byte]rune),
		encode:      make(map[rune]byte),
	}
	if d2, ok := base.(*differencesEncoding); ok {
		// merge differences
		diff := make(map[CharCode]GlyphName)
		for code, glyph := range d2.differences {
			diff[code] = glyph
		}
		for code, glyph := range differences {
			diff[code] = glyph
		}
		differences = diff
		base = d2.base
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

// Encode converts a Go unicode string to a PDF encoded string.
func (enc *differencesEncoding) Encode(str string) []byte {
	runes := []rune(str)
	buf := bytes.NewBuffer(nil)
	buf.Grow(len(runes))
	for _, r := range runes {
		code, _ := enc.RuneToCharcode(r)
		// relies on the fact that underlying encoding is 8 bit
		buf.WriteByte(byte(code))
	}
	return buf.Bytes()
}

// Decode converts PDF encoded string to a Go unicode string.
func (enc *differencesEncoding) Decode(raw []byte) string {
	runes := make([]rune, 0, len(raw))
	// relies on the fact that underlying encoding is 8 bit
	for _, b := range raw {
		r, _ := enc.CharcodeToRune(CharCode(b))
		runes = append(runes, r)
	}
	return string(runes)
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

// ToPdfObject returns the encoding as a PdfObject.
func (enc *differencesEncoding) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()
	dict.Set("Type", core.MakeName("Encoding"))
	dict.Set("BaseEncoding", enc.base.ToPdfObject())

	if diff := toFontDifferences(enc.differences); diff != nil {
		dict.Set("Differences", diff)
	} else {
		common.Log.Debug("WARN: font Differences array is nil. Output may be incorrect.")
	}

	return core.MakeIndirectObject(dict)
}
