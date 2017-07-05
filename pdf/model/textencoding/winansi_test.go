/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "testing"

func TestWinAnsiEncoder(t *testing.T) {
	enc := NewWinAnsiTextEncoder()

	glyph, found := enc.CharcodeToGlyphName(32)
	if !found || glyph != "space" {
		t.Errorf("Glyph != space")
		return
	}

	glyph, found = enc.RuneToGlyphName('þ')
	if !found || glyph != "thorn" {
		t.Errorf("Glyph != thorn")
		return
	}
}
