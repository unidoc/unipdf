/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"testing"
)

// TestGlypRune tests that glyphlistGlyphToRuneMap and glyphlistRuneToGlyphMap match
func TestGlypRune(t *testing.T) {
	for r, g := range glyphlistRuneToGlyphMap {
		r2, ok := glyphlistGlyphToRuneMap[g]
		if !ok {
			t.Errorf("rune→glyph→rune mismatch: %s → %q → %s", rs(r), g, rs(r2))
		}
	}

	for g, r := range glyphlistGlyphToRuneMap {
		g2, ok := glyphlistRuneToGlyphMap[r]
		if !ok {
			t.Errorf("glyph→rune→glyph mismatch: %q → %s → %q", g, rs(r), g2)
		}
	}
}

// TestRuneToGlyph checks for known glyph->rune mappings.
func TestRuneToGlyph(t *testing.T) {
	runes := []rune("₂₃₄₅◊ﬂ˝ˇ¾Ðí©ºªı„δ∂ℵ⌡×÷®Ï☎①➔➨")
	glyphs := []GlyphName{
		"twoinferior", "threeinferior", "fourinferior", "fiveinferior",
		"lozenge", "fl", "hungarumlaut", "caron",
		"threequarters", "Eth", "iacute", "copyright",
		"ordmasculine", "ordfeminine", "dotlessi", "quotedblbase",
		"delta", "partialdiff", "aleph", "integralbt",
		"multiply", "divide", "registered", "Idieresis",
		"a4", "a120", "a160", "a178",
	}

	if len(runes) != len(glyphs) {
		t.Fatalf("Bad test: runes=%d glyphs=%d", len(runes), len(glyphs))
	}
	for i, glyph := range glyphs {
		t.Run(string(glyph), func(t *testing.T) {
			expected := runes[i]
			r, ok := GlyphToRune(glyph)
			if !ok {
				t.Fatalf("no glyph %q", glyph)
			}
			if r != expected {
				t.Fatalf("Expected 0x%04x=%c. Got 0x%04x=%c", r, r, expected, expected)
			}
		})
	}
}
