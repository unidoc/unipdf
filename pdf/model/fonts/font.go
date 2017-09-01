/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fonts

import (
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// Font represents a font which is a series of glyphs. Character codes from PDF strings can be mapped to and from
// glyphs.  Each glyph has metrics.
type Font interface {
	Encoder() textencoding.TextEncoder
	SetEncoder(encoder textencoding.TextEncoder)
	GetGlyphCharMetrics(glyph string) (CharMetrics, bool)
	ToPdfObject() core.PdfObject
}

// CharMetrics represents width and height metrics of a glyph.
type CharMetrics struct {
	GlyphName string
	Wx        float64
	Wy        float64
}
