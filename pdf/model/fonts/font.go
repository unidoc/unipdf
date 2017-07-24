/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fonts

import (
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

type Font interface {
	SetEncoder(encoder textencoding.TextEncoder)
	GetGlyphCharMetrics(glyph string) (CharMetrics, bool)
	ToPdfObject() core.PdfObject
}

type CharMetrics struct {
	GlyphName string
	Wx        float64
	Wy        float64
}
