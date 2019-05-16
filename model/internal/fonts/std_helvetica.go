/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
/*
 * The embedded character metrics specified in this file are distributed under the terms listed in
 * ./testdata/afms/MustRead.html.
 */

package fonts

import "sync"

func init() {
	// The aliases seen for the standard 14 font names.
	// Most of these are from table 5.5.1 in
	// https://www.adobe.com/content/dam/acom/en/devnet/acrobat/pdfs/adobe_supplement_iso32000.pdf
	RegisterStdFont(HelveticaName, newFontHelvetica, "Arial")
	RegisterStdFont(HelveticaBoldName, newFontHelveticaBold, "Arial,Bold")
	RegisterStdFont(HelveticaObliqueName, newFontHelveticaOblique, "Arial,Italic")
	RegisterStdFont(HelveticaBoldObliqueName, newFontHelveticaBoldOblique, "Arial,BoldItalic")
}

const (
	// HelveticaName is a PDF name of the Helvetica font.
	HelveticaName = StdFontName("Helvetica")
	// HelveticaBoldName is a PDF name of the Helvetica (bold) font.
	HelveticaBoldName = StdFontName("Helvetica-Bold")
	// HelveticaObliqueName is a PDF name of the Helvetica (oblique) font.
	HelveticaObliqueName = StdFontName("Helvetica-Oblique")
	// HelveticaBoldObliqueName is a PDF name of the Helvetica (bold, oblique) font.
	HelveticaBoldObliqueName = StdFontName("Helvetica-BoldOblique")
)

// newFontHelvetica returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontHelvetica() StdFont {
	helveticaOnce.Do(initHelvetica)
	desc := Descriptor{
		Name:        HelveticaName,
		Family:      string(HelveticaName),
		Weight:      FontWeightMedium,
		Flags:       0x0020,
		BBox:        [4]float64{-166, -225, 1000, 931},
		ItalicAngle: 0,
		Ascent:      718,
		Descent:     -207,
		CapHeight:   718,
		XHeight:     523,
		StemV:       88,
		StemH:       76,
	}
	return NewStdFont(desc, helveticaCharMetrics)
}

// newFontHelveticaBold returns a new instance of the font with a default encoder set
// (WinAnsiEncoding).
func newFontHelveticaBold() StdFont {
	helveticaOnce.Do(initHelvetica)
	desc := Descriptor{
		Name:        HelveticaBoldName,
		Family:      string(HelveticaName),
		Weight:      FontWeightBold,
		Flags:       0x0020,
		BBox:        [4]float64{-170, -228, 1003, 962},
		ItalicAngle: 0,
		Ascent:      718,
		Descent:     -207,
		CapHeight:   718,
		XHeight:     532,
		StemV:       140,
		StemH:       118,
	}
	return NewStdFont(desc, helveticaBoldCharMetrics)
}

// newFontHelveticaOblique returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontHelveticaOblique() StdFont {
	helveticaOnce.Do(initHelvetica)
	desc := Descriptor{
		Name:        HelveticaObliqueName,
		Family:      string(HelveticaName),
		Weight:      FontWeightMedium,
		Flags:       0x0060,
		BBox:        [4]float64{-170, -225, 1116, 931},
		ItalicAngle: -12,
		Ascent:      718,
		Descent:     -207,
		CapHeight:   718,
		XHeight:     523,
		StemV:       88,
		StemH:       76,
	}
	return NewStdFont(desc, helveticaObliqueCharMetrics)
}

// newFontHelveticaBoldOblique returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontHelveticaBoldOblique() StdFont {
	helveticaOnce.Do(initHelvetica)
	desc := Descriptor{
		Name:        HelveticaBoldObliqueName,
		Family:      string(HelveticaName),
		Weight:      FontWeightBold,
		Flags:       0x0060,
		BBox:        [4]float64{-174, -228, 1114, 962},
		ItalicAngle: -12,
		Ascent:      718,
		Descent:     -207,
		CapHeight:   718,
		XHeight:     532,
		StemV:       140,
		StemH:       118,
	}
	return NewStdFont(desc, helveticaBoldObliqueCharMetrics)
}

var helveticaOnce sync.Once

func initHelvetica() {
	// unpack font metrics
	helveticaCharMetrics = make(map[rune]CharMetrics, len(type1CommonRunes))
	helveticaBoldCharMetrics = make(map[rune]CharMetrics, len(type1CommonRunes))
	for i, r := range type1CommonRunes {
		helveticaCharMetrics[r] = CharMetrics{Wx: float64(helveticaWx[i])}
		helveticaBoldCharMetrics[r] = CharMetrics{Wx: float64(helveticaBoldWx[i])}
	}
	helveticaObliqueCharMetrics = helveticaCharMetrics
	helveticaBoldObliqueCharMetrics = helveticaBoldCharMetrics
}

// helveticaCharMetrics are the font metrics loaded from afms/Helvetica.afm.
// See afms/MustRead.html for license information.
var helveticaCharMetrics map[rune]CharMetrics

// helveticaBoldCharMetrics are the font metrics loaded from afms/Helvetica-Bold.afm.
// See afms/MustRead.html for license information.
var helveticaBoldCharMetrics map[rune]CharMetrics

// helveticaBoldObliqueCharMetrics are the font metrics loaded from afms/Helvetica-BoldOblique.afm.
// See afms/MustRead.html for license information.
var helveticaBoldObliqueCharMetrics map[rune]CharMetrics

// helveticaObliqueCharMetrics are the font metrics loaded from afms/Helvetica-Oblique.afm.
// See afms/MustRead.html for license information.
var helveticaObliqueCharMetrics map[rune]CharMetrics

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
