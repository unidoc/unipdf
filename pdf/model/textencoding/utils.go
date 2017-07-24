/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "github.com/unidoc/unidoc/common"

func glyphToRune(glyph string, glyphToRuneMap map[string]rune) (rune, bool) {
	ucode, found := glyphToRuneMap[glyph]
	if found {
		return ucode, true
	}

	//common.Log.Debug("Glyph->Rune ERROR: Unable to find glyph %s", glyph)
	return 0, false
}

func runeToGlyph(ucode rune, runeToGlyphMap map[rune]string) (string, bool) {
	glyph, found := runeToGlyphMap[ucode]
	if found {
		return glyph, true
	}

	//common.Log.Debug("Rune->Glyph ERROR: Unable to find rune %v", ucode)
	return "", false
}

func splitWords(raw string, encoder TextEncoder) []string {
	runes := []rune(raw)

	words := []string{}

	startsAt := 0
	for idx, code := range runes {
		glyph, found := encoder.RuneToGlyph(code)
		if !found {
			common.Log.Debug("Glyph not found for code: %s\n", string(code))
			continue
		}

		if glyph == "space" {
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
