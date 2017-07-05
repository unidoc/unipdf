/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "github.com/unidoc/unidoc/pdf/core"

type TextEncoder interface {
	Encode(raw string) string
	CharcodeToGlyphName(code byte) (string, bool)
	RuneToCharcode(val rune) (byte, bool)
	RuneToGlyphName(val rune) (string, bool)
	GlyphNameToCharcode(glyph string) (byte, bool)
	CharcodeToRune(charcode byte) (rune, bool)
	ToPdfObject() core.PdfObject
}
