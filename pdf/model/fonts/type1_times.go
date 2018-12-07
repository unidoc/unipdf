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
	// TimesRomanName is a PDF name of the Times font.
	TimesRomanName = "Times-Roman"
	// TimesBoldName is a PDF name of the Times (bold) font.
	TimesBoldName = "Times-Bold"
	// TimesItalicName is a PDF name of the Times (italic) font.
	TimesItalicName = "Times-Italic"
	// TimesBoldItalicName is a PDF name of the Times (bold, italic) font.
	TimesBoldItalicName = "Times-BoldItalic"
)

// NewFontTimesRoman returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontTimesRoman() Type1Font {
	return NewType1Font(TimesRomanName, TimesRomanCharMetrics)
}

// NewFontTimesBold returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontTimesBold() Type1Font {
	return NewType1Font(TimesBoldName, TimesBoldCharMetrics)
}

// NewFontTimesItalic returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontTimesItalic() Type1Font {
	return NewType1Font(TimesItalicName, TimesItalicCharMetrics)
}

// NewFontTimesBoldItalic returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewFontTimesBoldItalic() Type1Font {
	return NewType1Font(TimesBoldItalicName, TimesBoldItalicCharMetrics)
}

func init() {
	// unpack font metrics
	// TODO(dennwc): once unexported, unpack on-demand (once)
	TimesRomanCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	TimesBoldCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	TimesBoldItalicCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	TimesItalicCharMetrics = make(map[GlyphName]CharMetrics, len(type1CommonGlyphs))
	for i, glyph := range type1CommonGlyphs {
		TimesRomanCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: float64(timesRomanWx[i])}
		TimesBoldCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: float64(timesBoldWx[i])}
		TimesBoldItalicCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: float64(timesBoldItalicWx[i])}
		TimesItalicCharMetrics[glyph] = CharMetrics{GlyphName: glyph, Wx: float64(timesItalicWx[i])}
	}
}

// TimesRomanCharMetrics are the font metrics loaded from afms/Times-Roman.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var TimesRomanCharMetrics map[GlyphName]CharMetrics

// TimesBoldCharMetrics are the font metrics loaded from afms/Times-Bold.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var TimesBoldCharMetrics map[GlyphName]CharMetrics

// TimesBoldItalicCharMetrics are the font metrics loaded from afms/Times-BoldItalic.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var TimesBoldItalicCharMetrics map[GlyphName]CharMetrics

// TimesItalicCharMetrics font metrics loaded from afms/Times-Italic.afm.
// See afms/MustRead.html for license information.
//
// TODO(dennwc): unexport
var TimesItalicCharMetrics map[GlyphName]CharMetrics

// timesRomanWx are the font metrics loaded from afms/Times-Roman.afm.
// See afms/MustRead.html for license information.
var timesRomanWx = []int16{
	722, 889, 722, 722, 722, 722, 722, 722, 722, 722,
	722, 667, 667, 667, 667, 667, 722, 722, 722, 612,
	611, 611, 611, 611, 611, 611, 611, 611, 611, 722,
	500, 556, 722, 722, 722, 722, 333, 333, 333, 333,
	333, 333, 333, 333, 389, 722, 722, 611, 611, 611,
	611, 611, 889, 722, 722, 722, 722, 722, 722, 889,
	722, 722, 722, 722, 722, 722, 722, 722, 556, 722,
	667, 667, 667, 667, 556, 556, 556, 556, 556, 611,
	611, 611, 556, 722, 722, 722, 722, 722, 722, 722,
	722, 722, 722, 944, 722, 722, 722, 722, 611, 611,
	611, 611, 444, 444, 444, 444, 333, 444, 667, 444,
	444, 778, 444, 444, 469, 541, 500, 921, 444, 500,
	278, 200, 480, 480, 333, 333, 333, 200, 350, 444,
	444, 333, 444, 444, 333, 500, 333, 278, 250, 250,
	760, 500, 500, 500, 500, 588, 500, 400, 333, 564,
	500, 333, 278, 444, 444, 444, 444, 444, 444, 444,
	500, 1000, 444, 1000, 500, 444, 564, 500, 333, 333,
	333, 556, 500, 556, 500, 500, 167, 500, 500, 500,
	500, 333, 564, 549, 500, 500, 333, 333, 500, 333,
	333, 278, 278, 278, 278, 278, 278, 278, 278, 500,
	500, 278, 278, 344, 278, 564, 549, 564, 471, 278,
	778, 333, 564, 500, 564, 500, 500, 500, 500, 500,
	549, 500, 500, 500, 500, 500, 500, 722, 333, 500,
	500, 500, 500, 750, 750, 300, 276, 310, 500, 500,
	500, 453, 333, 333, 476, 833, 250, 250, 1000, 564,
	564, 500, 444, 444, 408, 444, 444, 444, 333, 333,
	333, 180, 333, 333, 453, 333, 333, 760, 333, 389,
	389, 389, 389, 389, 500, 278, 500, 500, 278, 250,
	500, 600, 278, 326, 278, 500, 500, 750, 300, 333,
	980, 500, 300, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 500, 500, 722, 500, 500, 500, 500, 500,
	444, 444, 444, 444, 500,
}

// timesBoldWx are the font metrics loaded from afms/Times-Bold.afm.
// See afms/MustRead.html for license information.
var timesBoldWx = []int16{
	722, 1000, 722, 722, 722, 722, 722, 722, 722, 722,
	722, 667, 722, 722, 722, 722, 722, 722, 722, 612,
	667, 667, 667, 667, 667, 667, 667, 667, 667, 722,
	500, 611, 778, 778, 778, 778, 389, 389, 389, 389,
	389, 389, 389, 389, 500, 778, 778, 667, 667, 667,
	667, 667, 944, 722, 722, 722, 722, 722, 778, 1000,
	778, 778, 778, 778, 778, 778, 778, 778, 611, 778,
	722, 722, 722, 722, 556, 556, 556, 556, 556, 667,
	667, 667, 611, 722, 722, 722, 722, 722, 722, 722,
	722, 722, 722, 1000, 722, 722, 722, 722, 667, 667,
	667, 667, 500, 500, 500, 500, 333, 500, 722, 500,
	500, 833, 500, 500, 581, 520, 500, 930, 500, 556,
	278, 220, 394, 394, 333, 333, 333, 220, 350, 444,
	444, 333, 444, 444, 333, 500, 333, 333, 250, 250,
	747, 500, 556, 500, 500, 672, 556, 400, 333, 570,
	500, 333, 278, 444, 444, 444, 444, 444, 444, 444,
	500, 1000, 444, 1000, 500, 444, 570, 500, 333, 333,
	333, 556, 500, 556, 500, 500, 167, 500, 500, 500,
	556, 333, 570, 549, 500, 500, 333, 333, 556, 333,
	333, 278, 278, 278, 278, 278, 278, 278, 333, 556,
	556, 278, 278, 394, 278, 570, 549, 570, 494, 278,
	833, 333, 570, 556, 570, 556, 556, 556, 556, 500,
	549, 556, 500, 500, 500, 500, 500, 722, 333, 500,
	500, 500, 500, 750, 750, 300, 300, 330, 500, 500,
	556, 540, 333, 333, 494, 1000, 250, 250, 1000, 570,
	570, 556, 500, 500, 555, 500, 500, 500, 333, 333,
	333, 278, 444, 444, 549, 444, 444, 747, 333, 389,
	389, 389, 389, 389, 500, 333, 500, 500, 278, 250,
	500, 600, 333, 416, 333, 556, 500, 750, 300, 333,
	1000, 500, 300, 556, 556, 556, 556, 556, 556, 556,
	500, 556, 556, 500, 722, 500, 500, 500, 500, 500,
	444, 444, 444, 444, 500,
}

// timesBoldItalicWx are the font metrics loaded from afms/Times-BoldItalic.afm.
// See afms/MustRead.html for license information.
var timesBoldItalicWx = []int16{
	667, 944, 667, 667, 667, 667, 667, 667, 667, 667,
	667, 667, 667, 667, 667, 667, 722, 722, 722, 612,
	667, 667, 667, 667, 667, 667, 667, 667, 667, 722,
	500, 667, 722, 722, 722, 778, 389, 389, 389, 389,
	389, 389, 389, 389, 500, 667, 667, 611, 611, 611,
	611, 611, 889, 722, 722, 722, 722, 722, 722, 944,
	722, 722, 722, 722, 722, 722, 722, 722, 611, 722,
	667, 667, 667, 667, 556, 556, 556, 556, 556, 611,
	611, 611, 611, 722, 722, 722, 722, 722, 722, 722,
	722, 722, 667, 889, 667, 611, 611, 611, 611, 611,
	611, 611, 500, 500, 500, 500, 333, 500, 722, 500,
	500, 778, 500, 500, 570, 570, 500, 832, 500, 500,
	278, 220, 348, 348, 333, 333, 333, 220, 350, 444,
	444, 333, 444, 444, 333, 500, 333, 333, 250, 250,
	747, 500, 500, 500, 500, 608, 500, 400, 333, 570,
	500, 333, 278, 444, 444, 444, 444, 444, 444, 444,
	500, 1000, 444, 1000, 500, 444, 570, 500, 389, 389,
	333, 556, 500, 556, 500, 500, 167, 500, 500, 500,
	500, 333, 570, 549, 500, 500, 333, 333, 556, 333,
	333, 278, 278, 278, 278, 278, 278, 278, 278, 500,
	500, 278, 278, 382, 278, 570, 549, 606, 494, 278,
	778, 333, 606, 576, 570, 556, 556, 556, 556, 500,
	549, 556, 500, 500, 500, 500, 500, 722, 333, 500,
	500, 500, 500, 750, 750, 300, 266, 300, 500, 500,
	500, 500, 333, 333, 494, 833, 250, 250, 1000, 570,
	570, 500, 500, 500, 555, 500, 500, 500, 333, 333,
	333, 278, 389, 389, 549, 389, 389, 747, 333, 389,
	389, 389, 389, 389, 500, 333, 500, 500, 278, 250,
	500, 600, 278, 366, 278, 500, 500, 750, 300, 333,
	1000, 500, 300, 556, 556, 556, 556, 556, 556, 556,
	500, 556, 556, 444, 667, 500, 444, 444, 444, 500,
	389, 389, 389, 389, 500,
}

// timesItalicWx font metrics loaded from afms/Times-Italic.afm.
// See afms/MustRead.html for license information.
var timesItalicWx = []int16{
	611, 889, 611, 611, 611, 611, 611, 611, 611, 611,
	611, 611, 667, 667, 667, 667, 722, 722, 722, 612,
	611, 611, 611, 611, 611, 611, 611, 611, 611, 722,
	500, 611, 722, 722, 722, 722, 333, 333, 333, 333,
	333, 333, 333, 333, 444, 667, 667, 556, 556, 611,
	556, 556, 833, 667, 667, 667, 667, 667, 722, 944,
	722, 722, 722, 722, 722, 722, 722, 722, 611, 722,
	611, 611, 611, 611, 500, 500, 500, 500, 500, 556,
	556, 556, 611, 722, 722, 722, 722, 722, 722, 722,
	722, 722, 611, 833, 611, 556, 556, 556, 556, 556,
	556, 556, 500, 500, 500, 500, 333, 500, 667, 500,
	500, 778, 500, 500, 422, 541, 500, 920, 500, 500,
	278, 275, 400, 400, 389, 389, 333, 275, 350, 444,
	444, 333, 444, 444, 333, 500, 333, 333, 250, 250,
	760, 500, 500, 500, 500, 544, 500, 400, 333, 675,
	500, 333, 278, 444, 444, 444, 444, 444, 444, 444,
	500, 889, 444, 889, 500, 444, 675, 500, 333, 389,
	278, 500, 500, 500, 500, 500, 167, 500, 500, 500,
	500, 333, 675, 549, 500, 500, 333, 333, 500, 333,
	333, 278, 278, 278, 278, 278, 278, 278, 278, 444,
	444, 278, 278, 300, 278, 675, 549, 675, 471, 278,
	722, 333, 675, 500, 675, 500, 500, 500, 500, 500,
	549, 500, 500, 500, 500, 500, 500, 667, 333, 500,
	500, 500, 500, 750, 750, 300, 276, 310, 500, 500,
	500, 523, 333, 333, 476, 833, 250, 250, 1000, 675,
	675, 500, 500, 500, 420, 556, 556, 556, 333, 333,
	333, 214, 389, 389, 453, 389, 389, 760, 333, 389,
	389, 389, 389, 389, 500, 333, 500, 500, 278, 250,
	500, 600, 278, 300, 278, 500, 500, 750, 300, 333,
	980, 500, 300, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 500, 444, 667, 444, 444, 444, 444, 500,
	389, 389, 389, 389, 500,
}
