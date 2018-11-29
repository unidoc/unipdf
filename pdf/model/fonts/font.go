/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fonts

import (
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/textencoding"
)

// Font represents a font which is a series of glyphs. Character codes from PDF strings can be
// mapped to and from glyphs.  Each glyph has metrics.
type Font interface {
	Encoder() textencoding.TextEncoder
	// FIXME(dennwc): there is a bug with pointer receivers for this method
	//			      literally in every single font except pdfFontSimple
	//				  so this method does NOTHING for other fonts
	//				  if things still work, this means that only that font
	//				  needs the method and it shouldn't be in this interface
	SetEncoder(encoder textencoding.TextEncoder)
	GetGlyphCharMetrics(glyph textencoding.GlyphName) (CharMetrics, bool)
	ToPdfObject() core.PdfObject
}

// CharMetrics represents width and height metrics of a glyph.
type CharMetrics struct {
	GlyphName textencoding.GlyphName
	Wx        float64
	Wy        float64
}
