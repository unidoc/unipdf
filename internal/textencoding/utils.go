/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"fmt"
	"unicode"

	"github.com/unidoc/unipdf/v3/common"
)

func glyphToRune(glyph GlyphName, glyphToRuneMap map[GlyphName]rune) (rune, bool) {
	r, ok := glyphToRuneMap[glyph]
	if ok {
		return r, true
	}

	common.Log.Debug("ERROR: glyphToRune unable to find glyph %q", glyph)
	return 0, false
}

func runeToGlyph(r rune, runeToGlyphMap map[rune]GlyphName) (GlyphName, bool) {
	glyph, ok := runeToGlyphMap[r]
	if ok {
		return glyph, true
	}
	common.Log.Debug("ERROR: runeToGlyph unable to find glyph for rune %s", rs(r))
	return "", false
}

// rs returns a string describing rune `r`.
func rs(r rune) string {
	c := "unprintable"
	if unicode.IsPrint(r) {
		c = fmt.Sprintf("%#q", r)
	}
	return fmt.Sprintf("%+q (%s)", r, c)
}
