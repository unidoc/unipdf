/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"bytes"
	"strings"
	"testing"
)

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

	// Should encode hyphen to 0x2D consistently (not alternative 0xAD)
	b := enc.Encode(strings.Repeat("-", 100))
	if !bytes.Equal(b, bytes.Repeat([]byte{0x2D}, 100)) {
		t.Fatalf("Incorrect encoding of hyphen")
	}

	s := enc.Decode([]byte{0xAD, 0x2D})
	if s != "--" {
		t.Fatalf("Incorrect decoding of hyphen")
	}
}
