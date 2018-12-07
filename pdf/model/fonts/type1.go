/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fonts

import (
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/textencoding"
)

var _ Font = Type1Font{}

// Type1Font represents one of the built-in fonts and it is assumed that every reader has access to it.
type Type1Font struct {
	name    string
	metrics map[GlyphName]CharMetrics
	encoder textencoding.TextEncoder
}

// NewType1Font returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewType1Font(name string, metrics map[GlyphName]CharMetrics) Type1Font {
	enc := textencoding.NewWinAnsiTextEncoder() // Default
	return NewType1FontWithEncoding(name, metrics, enc)
}

// NewType1FontWithEncoding returns a new instance of the font with a specified encoder.
func NewType1FontWithEncoding(name string, metrics map[GlyphName]CharMetrics, encoder textencoding.TextEncoder) Type1Font {
	return Type1Font{
		name:    name,
		metrics: metrics,
		encoder: encoder,
	}
}

// Name returns a PDF name of the font.
func (font Type1Font) Name() string {
	return font.name
}

// Encoder returns the font's text encoder.
func (font Type1Font) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// GetGlyphCharMetrics returns character metrics for a given glyph.
func (font Type1Font) GetGlyphCharMetrics(glyph GlyphName) (CharMetrics, bool) {
	metrics, has := font.metrics[glyph]
	if !has {
		return metrics, false
	}

	return metrics, true
}

// ToPdfObject returns a primitive PDF object representation of the font.
func (font Type1Font) ToPdfObject() core.PdfObject {
	fontDict := core.MakeDict()
	fontDict.Set("Type", core.MakeName("Font"))
	fontDict.Set("Subtype", core.MakeName("Type1"))
	fontDict.Set("BaseFont", core.MakeName(font.name))
	fontDict.Set("Encoding", font.encoder.ToPdfObject())

	return core.MakeIndirectObject(fontDict)
}
