/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
/*
 * The embedded character metrics specified in this file are distributed under the terms listed in
 * ./testdata/afms/MustRead.html.
 */

package fonts

const (
	// HelveticaName is a PDF name of the Helvetica font.
	HelveticaName = "Helvetica"
	// HelveticaBoldName is a PDF name of the Helvetica (bold) font.
	HelveticaBoldName = "Helvetica-Bold"
	// HelveticaObliqueName is a PDF name of the Helvetica (oblique) font.
	HelveticaObliqueName = "Helvetica-Oblique"
	// HelveticaBoldObliqueName is a PDF name of the Helvetica (bold, oblique) font.
	HelveticaBoldObliqueName = "Helvetica-BoldOblique"
)

// NewFontHelvetica returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontHelvetica() Type1Font {
	return NewType1Font(HelveticaName, HelveticaCharMetrics)
}

// NewFontHelveticaBold returns a new instance of the font with a default encoder set
// (WinAnsiEncoding).
func NewFontHelveticaBold() Type1Font {
	return NewType1Font(HelveticaBoldName, HelveticaBoldCharMetrics)
}

// NewFontHelveticaOblique returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontHelveticaOblique() Type1Font {
	return NewType1Font(HelveticaObliqueName, HelveticaObliqueCharMetrics)
}

// NewFontHelveticaBoldOblique returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontHelveticaBoldOblique() Type1Font {
	return NewType1Font(HelveticaBoldObliqueName, HelveticaBoldObliqueCharMetrics)
}

func init() {
	// unpack font metrics
	// TODO(dennwc): once unexported, unpack on-demand (once)
	HelveticaCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	HelveticaBoldCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	for i, glyph := range type1CommonGlyphs {
		HelveticaCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: float64(helveticaWx[i])}
		HelveticaBoldCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: float64(helveticaBoldWx[i])}
	}
	HelveticaObliqueCharMetrics = HelveticaCharMetrics
	HelveticaBoldObliqueCharMetrics = HelveticaBoldCharMetrics
}

// HelveticaCharMetrics are the font metrics loaded from afms/Helvetica.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var HelveticaCharMetrics map[GlyphName]CharMetrics

// HelveticaBoldCharMetrics are the font metrics loaded from afms/Helvetica-Bold.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var HelveticaBoldCharMetrics map[GlyphName]CharMetrics

// HelveticaBoldObliqueCharMetrics are the font metrics loaded from afms/Helvetica-BoldOblique.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var HelveticaBoldObliqueCharMetrics map[GlyphName]CharMetrics

// HelveticaObliqueCharMetrics are the font metrics loaded from afms/Helvetica-Oblique.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var HelveticaObliqueCharMetrics map[GlyphName]CharMetrics

// helveticaWx are the font metrics loaded from afms/Helvetica.afm.
// See afms/MustRead.html for license information.
var helveticaWx = []int16{
	667, 1000, 667, 667, 667, 667, 667, 667, 667, 667,
	667, 667, 722, 722, 722, 722, 722, 722, 722, 612,
	667, 667, 667, 667, 667, 667, 667, 667, 667, 722,
	556, 611, 778, 778, 778, 722, 278, 278, 278, 278,
	278, 278, 278, 278, 500, 667, 667, 556, 556, 556,
	556, 556, 833, 722, 722, 722, 722, 722, 778, 1000,
	778, 778, 778, 778, 778, 778, 778, 778, 667, 778,
	722, 722, 722, 722, 667, 667, 667, 667, 667, 611,
	611, 611, 667, 722, 722, 722, 722, 722, 722, 722,
	722, 722, 667, 944, 667, 667, 667, 667, 611, 611,
	611, 611, 556, 556, 556, 556, 333, 556, 889, 556,
	556, 667, 556, 556, 469, 584, 389, 1015, 556, 556,
	278, 260, 334, 334, 278, 278, 333, 260, 350, 500,
	500, 333, 500, 500, 333, 556, 333, 278, 278, 250,
	737, 556, 556, 556, 556, 643, 556, 400, 333, 584,
	556, 333, 278, 556, 556, 556, 556, 556, 556, 556,
	556, 1000, 556, 1000, 556, 556, 584, 556, 278, 333,
	278, 500, 556, 500, 556, 556, 167, 556, 556, 556,
	611, 333, 584, 549, 556, 556, 333, 333, 556, 333,
	333, 222, 278, 278, 278, 278, 278, 222, 222, 500,
	500, 222, 222, 299, 222, 584, 549, 584, 471, 222,
	833, 333, 584, 556, 584, 556, 556, 556, 556, 556,
	549, 556, 556, 556, 556, 556, 556, 944, 333, 556,
	556, 556, 556, 834, 834, 333, 370, 365, 611, 556,
	556, 537, 333, 333, 476, 889, 278, 278, 1000, 584,
	584, 556, 556, 611, 355, 333, 333, 333, 222, 222,
	222, 191, 333, 333, 453, 333, 333, 737, 333, 500,
	500, 500, 500, 500, 556, 278, 556, 556, 278, 278,
	556, 600, 278, 317, 278, 556, 556, 834, 333, 333,
	1000, 556, 333, 556, 556, 556, 556, 556, 556, 556,
	556, 556, 556, 500, 722, 500, 500, 500, 500, 556,
	500, 500, 500, 500, 556,
}

// helveticaBoldWx are the font metrics loaded from afms/Helvetica-Bold.afm.
// See afms/MustRead.html for license information.
var helveticaBoldWx = []int16{
	722, 1000, 722, 722, 722, 722, 722, 722, 722, 722,
	722, 722, 722, 722, 722, 722, 722, 722, 722, 612,
	667, 667, 667, 667, 667, 667, 667, 667, 667, 722,
	556, 611, 778, 778, 778, 722, 278, 278, 278, 278,
	278, 278, 278, 278, 556, 722, 722, 611, 611, 611,
	611, 611, 833, 722, 722, 722, 722, 722, 778, 1000,
	778, 778, 778, 778, 778, 778, 778, 778, 667, 778,
	722, 722, 722, 722, 667, 667, 667, 667, 667, 611,
	611, 611, 667, 722, 722, 722, 722, 722, 722, 722,
	722, 722, 667, 944, 667, 667, 667, 667, 611, 611,
	611, 611, 556, 556, 556, 556, 333, 556, 889, 556,
	556, 722, 556, 556, 584, 584, 389, 975, 556, 611,
	278, 280, 389, 389, 333, 333, 333, 280, 350, 556,
	556, 333, 556, 556, 333, 556, 333, 333, 278, 250,
	737, 556, 611, 556, 556, 743, 611, 400, 333, 584,
	556, 333, 278, 556, 556, 556, 556, 556, 556, 556,
	556, 1000, 556, 1000, 556, 556, 584, 611, 333, 333,
	333, 611, 556, 611, 556, 556, 167, 611, 611, 611,
	611, 333, 584, 549, 556, 556, 333, 333, 611, 333,
	333, 278, 278, 278, 278, 278, 278, 278, 278, 556,
	556, 278, 278, 400, 278, 584, 549, 584, 494, 278,
	889, 333, 584, 611, 584, 611, 611, 611, 611, 556,
	549, 611, 556, 611, 611, 611, 611, 944, 333, 611,
	611, 611, 556, 834, 834, 333, 370, 365, 611, 611,
	611, 556, 333, 333, 494, 889, 278, 278, 1000, 584,
	584, 611, 611, 611, 474, 500, 500, 500, 278, 278,
	278, 238, 389, 389, 549, 389, 389, 737, 333, 556,
	556, 556, 556, 556, 556, 333, 556, 556, 278, 278,
	556, 600, 333, 389, 333, 611, 556, 834, 333, 333,
	1000, 556, 333, 611, 611, 611, 611, 611, 611, 611,
	556, 611, 611, 556, 778, 556, 556, 556, 556, 556,
	500, 500, 500, 500, 556,
}
