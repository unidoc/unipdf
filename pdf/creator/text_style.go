/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// TextStyle is a collection of properties that can be assigned to a chunk of text.
type TextStyle struct {
	// The color of the text.
	Color Color

	// The font the text will use.
	Font fonts.Font

	// The size of the font.
	FontSize float64
}

func NewTextStyle() TextStyle {
	font := fonts.NewFontHelvetica()
	font.SetEncoder(textencoding.NewWinAnsiTextEncoder())

	return TextStyle{
		Color:    ColorRGBFrom8bit(0, 0, 0),
		Font:     font,
		FontSize: 10,
	}
}

type TextChunk struct {
	Text  string
	Style TextStyle
}
