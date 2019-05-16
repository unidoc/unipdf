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
	RegisterStdFont(TimesRomanName, newFontTimesRoman, "TimesNewRoman", "Times")
	RegisterStdFont(TimesBoldName, newFontTimesBold, "TimesNewRoman,Bold", "Times,Bold")
	RegisterStdFont(TimesItalicName, newFontTimesItalic, "TimesNewRoman,Italic", "Times,Italic")
	RegisterStdFont(TimesBoldItalicName, newFontTimesBoldItalic, "TimesNewRoman,BoldItalic", "Times,BoldItalic")
}

const (
	// timesFamily is a PDF name of the Times font family.
	timesFamily = "Times"
	// TimesRomanName is a PDF name of the Times font.
	TimesRomanName = StdFontName("Times-Roman")
	// TimesBoldName is a PDF name of the Times (bold) font.
	TimesBoldName = StdFontName("Times-Bold")
	// TimesItalicName is a PDF name of the Times (italic) font.
	TimesItalicName = StdFontName("Times-Italic")
	// TimesBoldItalicName is a PDF name of the Times (bold, italic) font.
	TimesBoldItalicName = StdFontName("Times-BoldItalic")
)

// newFontTimesRoman returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontTimesRoman() StdFont {
	timesOnce.Do(initTimes)
	desc := Descriptor{
		Name:        TimesRomanName,
		Family:      timesFamily,
		Weight:      FontWeightRoman,
		Flags:       0x0020,
		BBox:        [4]float64{-168, -218, 1000, 898},
		ItalicAngle: 0,
		Ascent:      683,
		Descent:     -217,
		CapHeight:   662,
		XHeight:     450,
		StemV:       84,
		StemH:       28,
	}
	return NewStdFont(desc, timesRomanCharMetrics)
}

// newFontTimesBold returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontTimesBold() StdFont {
	timesOnce.Do(initTimes)
	desc := Descriptor{
		Name:        TimesBoldName,
		Family:      timesFamily,
		Weight:      FontWeightBold,
		Flags:       0x0020,
		BBox:        [4]float64{-168, -218, 1000, 935},
		ItalicAngle: 0,
		Ascent:      683,
		Descent:     -217,
		CapHeight:   676,
		XHeight:     461,
		StemV:       139,
		StemH:       44,
	}
	return NewStdFont(desc, timesBoldCharMetrics)
}

// newFontTimesItalic returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontTimesItalic() StdFont {
	timesOnce.Do(initTimes)
	desc := Descriptor{
		Name:        TimesItalicName,
		Family:      timesFamily,
		Weight:      FontWeightMedium,
		Flags:       0x0060,
		BBox:        [4]float64{-169, -217, 1010, 883},
		ItalicAngle: -15.5,
		Ascent:      683,
		Descent:     -217,
		CapHeight:   653,
		XHeight:     441,
		StemV:       76,
		StemH:       32,
	}
	return NewStdFont(desc, timesItalicCharMetrics)
}

// newFontTimesBoldItalic returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func newFontTimesBoldItalic() StdFont {
	timesOnce.Do(initTimes)
	desc := Descriptor{
		Name:        TimesBoldItalicName,
		Family:      timesFamily,
		Weight:      FontWeightBold,
		Flags:       0x0060,
		BBox:        [4]float64{-200, -218, 996, 921},
		ItalicAngle: -15,
		Ascent:      683,
		Descent:     -217,
		CapHeight:   669,
		XHeight:     462,
		StemV:       121,
		StemH:       42,
	}
	return NewStdFont(desc, timesBoldItalicCharMetrics)
}

var timesOnce sync.Once

func initTimes() {
	// unpack font metrics
	timesRomanCharMetrics = make(map[rune]CharMetrics, len(type1CommonRunes))
	timesBoldCharMetrics = make(map[rune]CharMetrics, len(type1CommonRunes))
	timesBoldItalicCharMetrics = make(map[rune]CharMetrics, len(type1CommonRunes))
	timesItalicCharMetrics = make(map[rune]CharMetrics, len(type1CommonRunes))
	for i, r := range type1CommonRunes {
		timesRomanCharMetrics[r] = CharMetrics{Wx: float64(timesRomanWx[i])}
		timesBoldCharMetrics[r] = CharMetrics{Wx: float64(timesBoldWx[i])}
		timesBoldItalicCharMetrics[r] = CharMetrics{Wx: float64(timesBoldItalicWx[i])}
		timesItalicCharMetrics[r] = CharMetrics{Wx: float64(timesItalicWx[i])}
	}
}

// timesRomanCharMetrics are the font metrics loaded from afms/Times-Roman.afm.
// See afms/MustRead.html for license information.
var timesRomanCharMetrics map[rune]CharMetrics

// timesBoldCharMetrics are the font metrics loaded from afms/Times-Bold.afm.
// See afms/MustRead.html for license information.
var timesBoldCharMetrics map[rune]CharMetrics

// timesBoldItalicCharMetrics are the font metrics loaded from afms/Times-BoldItalic.afm.
// See afms/MustRead.html for license information.
var timesBoldItalicCharMetrics map[rune]CharMetrics

// timesItalicCharMetrics font metrics loaded from afms/Times-Italic.afm.
// See afms/MustRead.html for license information.
var timesItalicCharMetrics map[rune]CharMetrics

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
