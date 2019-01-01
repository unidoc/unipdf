/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "testing"

func TestWinAnsiEncoder(t *testing.T) {
	enc := NewWinAnsiEncoder()

	r, found := enc.CharcodeToRune(32)
	if !found || r != ' ' {
		t.Errorf("rune != space")
		return
	}
	code, found := enc.RuneToCharcode('þ')
	if !found || code != 254 {
		t.Errorf("code != 254")
		return
	}

	glyph, found := RuneToGlyph('þ')
	if !found || glyph != "thorn" {
		t.Errorf("Glyph != thorn")
		return
	}
}
