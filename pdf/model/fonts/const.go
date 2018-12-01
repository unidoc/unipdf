/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
/*
 * The embedded character metrics specified in this file are distributed under the terms listed in
 * ./afms/MustRead.html.
 */

package fonts

// FontWeight specified font weight.
type FontWeight int

// Font weights
const (
	FontWeightMedium FontWeight = iota // Medium
	FontWeightBold                     // Bold
	FontWeightRoman                    // Roman
)

// DescriptorLiteral describes geometric properties of a font.
type DescriptorLiteral struct {
	FontName   string
	FontFamily string
	FontWeight
	Flags       uint
	FontBBox    [4]float64
	ItalicAngle float64
	Ascent      float64
	Descent     float64
	CapHeight   float64
	XHeight     float64
	StemV       float64
	StemH       float64
}

var (
	Standard14Fonts = map[string]Font{
		"Courier":               NewFontCourier(),
		"Courier-Bold":          NewFontCourierBold(),
		"Courier-BoldOblique":   NewFontCourierBoldOblique(),
		"Courier-Oblique":       NewFontCourierOblique(),
		"Helvetica":             NewFontHelvetica(),
		"Helvetica-Bold":        NewFontHelveticaBold(),
		"Helvetica-BoldOblique": NewFontHelveticaBoldOblique(),
		"Helvetica-Oblique":     NewFontHelveticaOblique(),
		"Times-Roman":           NewFontTimesRoman(),
		"Times-Bold":            NewFontTimesBold(),
		"Times-BoldItalic":      NewFontTimesBoldItalic(),
		"Times-Italic":          NewFontTimesItalic(),
		"Symbol":                NewFontSymbol(),
		"ZapfDingbats":          NewFontZapfDingbats(),
	}

	Standard14Descriptors = map[string]DescriptorLiteral{
		"Courier": DescriptorLiteral{
			FontName:    "Courier",
			FontFamily:  "Courier",
			FontWeight:  FontWeightMedium,
			Flags:       0x0021,
			FontBBox:    [4]float64{-23, -250, 715, 805},
			ItalicAngle: 0,
			Ascent:      629,
			Descent:     -157,
			CapHeight:   562,
			XHeight:     426,
			StemV:       51,
			StemH:       51,
		},
		"Courier-Bold": DescriptorLiteral{
			FontName:    "Courier-Bold",
			FontFamily:  "Courier",
			FontWeight:  FontWeightBold,
			Flags:       0x0021,
			FontBBox:    [4]float64{-113, -250, 749, 801},
			ItalicAngle: 0,
			Ascent:      629,
			Descent:     -157,
			CapHeight:   562,
			XHeight:     439,
			StemV:       106,
			StemH:       84,
		},
		"Courier-BoldOblique": DescriptorLiteral{
			FontName:    "Courier-BoldOblique",
			FontFamily:  "Courier",
			FontWeight:  FontWeightBold,
			Flags:       0x0061,
			FontBBox:    [4]float64{-57, -250, 869, 801},
			ItalicAngle: -12,
			Ascent:      629,
			Descent:     -157,
			CapHeight:   562,
			XHeight:     439,
			StemV:       106,
			StemH:       84,
		},
		"Courier-Oblique": DescriptorLiteral{
			FontName:    "Courier-Oblique",
			FontFamily:  "Courier",
			FontWeight:  FontWeightMedium,
			Flags:       0x0061,
			FontBBox:    [4]float64{-27, -250, 849, 805},
			ItalicAngle: -12,
			Ascent:      629,
			Descent:     -157,
			CapHeight:   562,
			XHeight:     426,
			StemV:       51,
			StemH:       51,
		},
		"Helvetica": DescriptorLiteral{
			FontName:    "Helvetica",
			FontFamily:  "Helvetica",
			FontWeight:  FontWeightMedium,
			Flags:       0x0020,
			FontBBox:    [4]float64{-166, -225, 1000, 931},
			ItalicAngle: 0,
			Ascent:      718,
			Descent:     -207,
			CapHeight:   718,
			XHeight:     523,
			StemV:       88,
			StemH:       76,
		},
		"Helvetica-Bold": DescriptorLiteral{
			FontName:    "Helvetica-Bold",
			FontFamily:  "Helvetica",
			FontWeight:  FontWeightBold,
			Flags:       0x0020,
			FontBBox:    [4]float64{-170, -228, 1003, 962},
			ItalicAngle: 0,
			Ascent:      718,
			Descent:     -207,
			CapHeight:   718,
			XHeight:     532,
			StemV:       140,
			StemH:       118,
		},
		"Helvetica-BoldOblique": DescriptorLiteral{
			FontName:    "Helvetica-BoldOblique",
			FontFamily:  "Helvetica",
			FontWeight:  FontWeightBold,
			Flags:       0x0060,
			FontBBox:    [4]float64{-174, -228, 1114, 962},
			ItalicAngle: -12,
			Ascent:      718,
			Descent:     -207,
			CapHeight:   718,
			XHeight:     532,
			StemV:       140,
			StemH:       118,
		},
		"Helvetica-Oblique": DescriptorLiteral{
			FontName:    "Helvetica-Oblique",
			FontFamily:  "Helvetica",
			FontWeight:  FontWeightMedium,
			Flags:       0x0060,
			FontBBox:    [4]float64{-170, -225, 1116, 931},
			ItalicAngle: -12,
			Ascent:      718,
			Descent:     -207,
			CapHeight:   718,
			XHeight:     523,
			StemV:       88,
			StemH:       76,
		},
		"Times-Roman": DescriptorLiteral{
			FontName:    "Times-Roman",
			FontFamily:  "Times",
			FontWeight:  FontWeightRoman,
			Flags:       0x0020,
			FontBBox:    [4]float64{-168, -218, 1000, 898},
			ItalicAngle: 0,
			Ascent:      683,
			Descent:     -217,
			CapHeight:   662,
			XHeight:     450,
			StemV:       84,
			StemH:       28,
		},
		"Times-Bold": DescriptorLiteral{
			FontName:    "Times-Bold",
			FontFamily:  "Times",
			FontWeight:  FontWeightBold,
			Flags:       0x0020,
			FontBBox:    [4]float64{-168, -218, 1000, 935},
			ItalicAngle: 0,
			Ascent:      683,
			Descent:     -217,
			CapHeight:   676,
			XHeight:     461,
			StemV:       139,
			StemH:       44,
		},
		"Times-BoldItalic": DescriptorLiteral{
			FontName:    "Times-BoldItalic",
			FontFamily:  "Times",
			FontWeight:  FontWeightBold,
			Flags:       0x0060,
			FontBBox:    [4]float64{-200, -218, 996, 921},
			ItalicAngle: -15,
			Ascent:      683,
			Descent:     -217,
			CapHeight:   669,
			XHeight:     462,
			StemV:       121,
			StemH:       42,
		},
		"Times-Italic": DescriptorLiteral{
			FontName:    "Times-Italic",
			FontFamily:  "Times",
			FontWeight:  FontWeightMedium,
			Flags:       0x0060,
			FontBBox:    [4]float64{-169, -217, 1010, 883},
			ItalicAngle: -15.5,
			Ascent:      683,
			Descent:     -217,
			CapHeight:   653,
			XHeight:     441,
			StemV:       76,
			StemH:       32,
		},
		"Symbol": DescriptorLiteral{
			FontName:    "Symbol",
			FontFamily:  "Symbol",
			FontWeight:  FontWeightMedium,
			Flags:       0x0004,
			FontBBox:    [4]float64{-180, -293, 1090, 1010},
			ItalicAngle: 0,
			Ascent:      0,
			Descent:     0,
			CapHeight:   0,
			XHeight:     0,
			StemV:       85,
			StemH:       92,
		},
		"ZapfDingbats": DescriptorLiteral{
			FontName:    "ZapfDingbats",
			FontFamily:  "ZapfDingbats",
			FontWeight:  FontWeightMedium,
			Flags:       0x0004,
			FontBBox:    [4]float64{-1, -143, 981, 820},
			ItalicAngle: 0,
			Ascent:      0,
			Descent:     0,
			CapHeight:   0,
			XHeight:     0,
			StemV:       90,
			StemH:       28,
		},
	}
)
