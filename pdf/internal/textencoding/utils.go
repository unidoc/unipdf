/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"fmt"
	"unicode"

	"github.com/unidoc/unidoc/common"
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

func splitWords(raw string, encoder TextEncoder) []string {
	runes := []rune(raw)

	var words []string

	startsAt := 0
	for idx, r := range runes {
		glyph, ok := encoder.RuneToGlyph(r)
		if !ok {
			common.Log.Debug("Glyph not found for rune %s", rs(r))
			continue
		}

		if glyph == "space" || glyph == "uni0020" {
			word := runes[startsAt:idx]
			words = append(words, string(word))
			startsAt = idx + 1
		}
	}

	word := runes[startsAt:]
	if len(word) > 0 {
		words = append(words, string(word))
	}

	return words
}

// rs returns a string describing rune `r`.
func rs(r rune) string {
	c := "unprintable"
	if unicode.IsPrint(r) {
		c = fmt.Sprintf("%#q", r)
	}
	return fmt.Sprintf("%+q (%s)", r, c)
}
