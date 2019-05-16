/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unipdf/v3/model"
)

// TextStyle is a collection of properties that can be assigned to a chunk of text.
type TextStyle struct {
	// The color of the text.
	Color Color

	// The font the text will use.
	Font *model.PdfFont

	// The size of the font.
	FontSize float64

	// The character spacing.
	CharSpacing float64

	// The rendering mode.
	RenderingMode TextRenderingMode
}

// newTextStyle creates a new text style object using the specified font.
func newTextStyle(font *model.PdfFont) TextStyle {
	return TextStyle{
		Color:    ColorRGBFrom8bit(0, 0, 0),
		Font:     font,
		FontSize: 10,
	}
}

// newLinkStyle creates a new text style object which can be
// used for link annotations.
func newLinkStyle(font *model.PdfFont) TextStyle {
	return TextStyle{
		Color:    ColorRGBFrom8bit(0, 0, 238),
		Font:     font,
		FontSize: 10,
	}
}
