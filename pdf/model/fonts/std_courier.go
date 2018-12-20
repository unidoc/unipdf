/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
/*
 * The embedded character metrics specified in this file are distributed under the terms listed in
 * ./afms/MustRead.html.
 */

package fonts

import "sync"

func init() {
	RegisterStdFont(CourierName, NewFontCourier)
	RegisterStdFont(CourierBoldName, NewFontCourierBold)
	RegisterStdFont(CourierObliqueName, NewFontCourierOblique)
	RegisterStdFont(CourierBoldObliqueName, NewFontCourierBoldOblique)
}

const (
	// CourierName is a PDF name of the Courier font.
	CourierName = StdFontName("Courier")
	// CourierBoldName is a PDF name of the Courier (bold) font.
	CourierBoldName = StdFontName("Courier-Bold")
	// CourierObliqueName is a PDF name of the Courier (oblique) font.
	CourierObliqueName = StdFontName("Courier-Oblique")
	// CourierBoldObliqueName is a PDF name of the Courier (bold, oblique) font.
	CourierBoldObliqueName = StdFontName("Courier-BoldOblique")
)

// NewFontCourier returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontCourier() StdFont {
	courierOnce.Do(initCourier)
	return NewStdFont(CourierName, courierCharMetrics)
}

// NewFontCourierBold returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontCourierBold() StdFont {
	courierOnce.Do(initCourier)
	return NewStdFont(CourierBoldName, courierBoldCharMetrics)
}

// NewFontCourierOblique returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontCourierOblique() StdFont {
	courierOnce.Do(initCourier)
	return NewStdFont(CourierObliqueName, courierObliqueCharMetrics)
}

// NewFontCourierBoldOblique returns a new instance of the font with a default encoder set
// (WinAnsiEncoding).
func NewFontCourierBoldOblique() StdFont {
	courierOnce.Do(initCourier)
	return NewStdFont(CourierBoldObliqueName, courierBoldObliqueCharMetrics)
}

var courierOnce sync.Once

func initCourier() {
	// the only font that has same metrics for all glyphs (fixed-width)
	const wx = 600
	courierCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	for _, glyph := range type1CommonGlyphs {
		courierCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: wx}
	}
	// other font variant still have the same metrics
	courierBoldCharMetrics = courierCharMetrics
	courierBoldObliqueCharMetrics = courierCharMetrics
	courierObliqueCharMetrics = courierCharMetrics
}

// courierCharMetrics are the font metrics loaded from afms/Courier.afm.  See afms/MustRead.html for
// license information.
var courierCharMetrics map[GlyphName]CharMetrics

// Courier-Bold font metrics loaded from afms/Courier-Bold.afm.  See afms/MustRead.html for license information.
var courierBoldCharMetrics map[GlyphName]CharMetrics

// courierBoldObliqueCharMetrics are the font metrics loaded from afms/Courier-BoldOblique.afm.
// See afms/MustRead.html for license information.
var courierBoldObliqueCharMetrics map[GlyphName]CharMetrics

// courierObliqueCharMetrics are the font metrics loaded from afms/Courier-Oblique.afm.
// See afms/MustRead.html for license information.
var courierObliqueCharMetrics map[GlyphName]CharMetrics
