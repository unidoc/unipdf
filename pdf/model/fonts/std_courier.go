/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
/*
 * The embedded character metrics specified in this file are distributed under the terms listed in
 * ./afms/MustRead.html.
 */

package fonts

const (
	// CourierName is a PDF name of the Courier font.
	CourierName = "Courier"
	// CourierBoldName is a PDF name of the Courier (bold) font.
	CourierBoldName = "Courier-Bold"
	// CourierObliqueName is a PDF name of the Courier (oblique) font.
	CourierObliqueName = "Courier-Oblique"
	// CourierBoldObliqueName is a PDF name of the Courier (bold, oblique) font.
	CourierBoldObliqueName = "Courier-BoldOblique"
)

// NewFontCourier returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontCourier() StdFont {
	return NewStdFont(CourierName, CourierCharMetrics)
}

// NewFontCourierBold returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontCourierBold() StdFont {
	return NewStdFont(CourierBoldName, CourierBoldCharMetrics)
}

// NewFontCourierOblique returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontCourierOblique() StdFont {
	return NewStdFont(CourierObliqueName, CourierObliqueCharMetrics)
}

// NewFontCourierBoldOblique returns a new instance of the font with a default encoder set
// (WinAnsiEncoding).
func NewFontCourierBoldOblique() StdFont {
	return NewStdFont(CourierBoldObliqueName, CourierBoldObliqueCharMetrics)
}

func init() {
	// the only font that has same metrics for all glyphs (fixed-width)
	// TODO(dennwc): once unexported, unpack on-demand (once)
	const wx = 600
	CourierCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	for _, glyph := range type1CommonGlyphs {
		CourierCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: wx}
	}
	// other font variant still have the same metrics
	CourierBoldCharMetrics = CourierCharMetrics
	CourierBoldObliqueCharMetrics = CourierCharMetrics
	CourierObliqueCharMetrics = CourierCharMetrics
}

// CourierCharMetrics are the font metrics loaded from afms/Courier.afm.  See afms/MustRead.html for
// license information.
//
// TODO(dennwc): unexport
var CourierCharMetrics map[GlyphName]CharMetrics

// Courier-Bold font metrics loaded from afms/Courier-Bold.afm.  See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var CourierBoldCharMetrics map[GlyphName]CharMetrics

// CourierBoldObliqueCharMetrics are the font metrics loaded from afms/Courier-BoldOblique.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var CourierBoldObliqueCharMetrics map[GlyphName]CharMetrics

// CourierObliqueCharMetrics are the font metrics loaded from afms/Courier-Oblique.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var CourierObliqueCharMetrics map[GlyphName]CharMetrics
