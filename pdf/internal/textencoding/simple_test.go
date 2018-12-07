/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"fmt"
	"testing"

	"github.com/unidoc/unidoc/common"
)

// This test covers all the standard encodings in simple.go

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

// TestBasicEncodings checks for known glyph->rune mappings in the standard encodings.
func TestBasicEncodings(t *testing.T) {
	for _, test := range testCases {
		test.check(t)
	}
}

var testCases = []encodingTest{
	{"MacExpertEncoding", "₂₃₄₅", []GlyphName{"twoinferior", "threeinferior", "fourinferior", "fiveinferior"}},
	{"MacRomanEncoding", "◊ﬂ˝ˇ", []GlyphName{"lozenge", "fl", "hungarumlaut", "caron"}},
	{"PdfDocEncoding", "¾Ðí©", []GlyphName{"threequarters", "Eth", "iacute", "copyright"}},
	{"StandardEncoding", "ºªı„", []GlyphName{"ordmasculine", "ordfeminine", "dotlessi", "quotedblbase"}},
	{"SymbolEncoding", "δ∂ℵ⌡", []GlyphName{"delta", "partialdiff", "aleph", "integralbt"}},
	{"WinAnsiEncoding", "×÷®Ï", []GlyphName{"multiply", "divide", "registered", "Idieresis"}},
	{"ZapfDingbatsEncoding", "☎①➔➨", []GlyphName{"a4", "a120", "a160", "a178"}},
}

type encodingTest struct {
	encoding string
	runes    string
	glyphs   []GlyphName
}

func (f *encodingTest) String() string {
	return fmt.Sprintf("ENCODING_TEST{%#q}", f.encoding)
}

func (f *encodingTest) check(t *testing.T) {
	common.Log.Debug("encodingTest: %s", f)
	runes := []rune(f.runes)
	if len(runes) != len(f.glyphs) {
		t.Fatalf("Bad test %s runes=%d glyphs=%d", f, len(runes), len(f.glyphs))
	}
	enc, err := NewSimpleTextEncoder(f.encoding, nil)
	if err != nil {
		t.Fatalf("NewSimpleTextEncoder(%#q) failed. err=%v", f.encoding, err)
	}
	for i, glyph := range f.glyphs {
		expected := runes[i]
		r, ok := enc.GlyphToRune(glyph)
		if !ok {
			t.Fatalf("Encoding %#q has no glyph %q", f.encoding, glyph)
		}
		if r != expected {
			t.Fatalf("%s: Expected 0x%04x=%c. Got 0x%04x=%c", f, r, r, expected, expected)
		}
	}
}
