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
	// The aliases seen for the standard 14 font names.
	// Most of these are from table 5.5.1 in
	// https://www.adobe.com/content/dam/acom/en/devnet/acrobat/pdfs/adobe_supplement_iso32000.pdf
	RegisterStdFont(CourierName, newFontCourier, "CourierCourierNew", "CourierNew")
	RegisterStdFont(CourierBoldName, newFontCourierBold, "CourierNew,Bold")
	RegisterStdFont(CourierObliqueName, newFontCourierOblique, "CourierNew,Italic")
	RegisterStdFont(CourierBoldObliqueName, newFontCourierBoldOblique, "CourierNew,BoldItalic")
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

// newFontCourier returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontCourier() StdFont {
	courierOnce.Do(initCourier)
	desc := Descriptor{
		Name:        CourierName,
		Family:      string(CourierName),
		Weight:      FontWeightMedium,
		Flags:       0x0021,
		BBox:        [4]float64{-23, -250, 715, 805},
		ItalicAngle: 0,
		Ascent:      629,
		Descent:     -157,
		CapHeight:   562,
		XHeight:     426,
		StemV:       51,
		StemH:       51,
	}
	return NewStdFont(desc, courierCharMetrics)
}

// newFontCourierBold returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontCourierBold() StdFont {
	courierOnce.Do(initCourier)
	desc := Descriptor{
		Name:        CourierBoldName,
		Family:      string(CourierName),
		Weight:      FontWeightBold,
		Flags:       0x0021,
		BBox:        [4]float64{-113, -250, 749, 801},
		ItalicAngle: 0,
		Ascent:      629,
		Descent:     -157,
		CapHeight:   562,
		XHeight:     439,
		StemV:       106,
		StemH:       84,
	}
	return NewStdFont(desc, courierBoldCharMetrics)
}

// NewFontCourierOblique returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontCourierOblique() StdFont {
	courierOnce.Do(initCourier)
	desc := Descriptor{
		Name:        CourierObliqueName,
		Family:      string(CourierName),
		Weight:      FontWeightMedium,
		Flags:       0x0061,
		BBox:        [4]float64{-27, -250, 849, 805},
		ItalicAngle: -12,
		Ascent:      629,
		Descent:     -157,
		CapHeight:   562,
		XHeight:     426,
		StemV:       51,
		StemH:       51,
	}
	return NewStdFont(desc, courierObliqueCharMetrics)
}

// NewFontCourierBoldOblique returns a new instance of the font with a default encoder set
// (WinAnsiEncoding).
func newFontCourierBoldOblique() StdFont {
	courierOnce.Do(initCourier)
	desc := Descriptor{
		Name:        CourierBoldObliqueName,
		Family:      string(CourierName),
		Weight:      FontWeightBold,
		Flags:       0x0061,
		BBox:        [4]float64{-57, -250, 869, 801},
		ItalicAngle: -12,
		Ascent:      629,
		Descent:     -157,
		CapHeight:   562,
		XHeight:     439,
		StemV:       106,
		StemH:       84,
	}
	return NewStdFont(desc, courierBoldObliqueCharMetrics)
}

var courierOnce sync.Once

func initCourier() {
	// the only font that has same metrics for all glyphs (fixed-width)
	const wx = 600
	courierCharMetrics = make(map[rune]CharMetrics, len(type1CommonRunes))
	for _, r := range type1CommonRunes {
		courierCharMetrics[r] = CharMetrics{Wx: wx}
	}
	// other font variant still have the same metrics
	courierBoldCharMetrics = courierCharMetrics
	courierBoldObliqueCharMetrics = courierCharMetrics
	courierObliqueCharMetrics = courierCharMetrics
}

// courierCharMetrics are the font metrics loaded from afms/Courier.afm.  See afms/MustRead.html for
// license information.
var courierCharMetrics map[rune]CharMetrics

// Courier-Bold font metrics loaded from afms/Courier-Bold.afm.  See afms/MustRead.html for license information.
var courierBoldCharMetrics map[rune]CharMetrics

// courierBoldObliqueCharMetrics are the font metrics loaded from afms/Courier-BoldOblique.afm.
// See afms/MustRead.html for license information.
var courierBoldObliqueCharMetrics map[rune]CharMetrics

// courierObliqueCharMetrics are the font metrics loaded from afms/Courier-Oblique.afm.
// See afms/MustRead.html for license information.
var courierObliqueCharMetrics map[rune]CharMetrics
