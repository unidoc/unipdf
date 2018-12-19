/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fonts

import (
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/textencoding"
)

var stdFonts = make(map[string]func() StdFont)

// NewStdFontByName creates a new StdFont by registered name. See RegisterStdFont.
func NewStdFontByName(name string) (StdFont, bool) {
	fnc, ok := stdFonts[name]
	if !ok {
		return StdFont{}, false
	}
	return fnc(), true
}

// RegisterStdFont registers a given StdFont constructor by font name. Font can then be created with NewStdFontByName.
func RegisterStdFont(name string, fnc func() StdFont) {
	if _, ok := stdFonts[name]; ok {
		panic("font already registered: " + name)
	}
	stdFonts[name] = fnc
}

var _ Font = StdFont{}

// StdFont represents one of the built-in fonts and it is assumed that every reader has access to it.
type StdFont struct {
	name    string
	metrics map[GlyphName]CharMetrics
	encoder textencoding.TextEncoder
}

// NewStdFont returns a new instance of the font with a default encoder set (WinAnsiEncoding).
func NewStdFont(name string, metrics map[GlyphName]CharMetrics) StdFont {
	enc := textencoding.NewWinAnsiTextEncoder() // Default
	return NewStdFontWithEncoding(name, metrics, enc)
}

// NewStdFontWithEncoding returns a new instance of the font with a specified encoder.
func NewStdFontWithEncoding(name string, metrics map[GlyphName]CharMetrics, encoder textencoding.TextEncoder) StdFont {
	return StdFont{
		name:    name,
		metrics: metrics,
		encoder: encoder,
	}
}

// Name returns a PDF name of the font.
func (font StdFont) Name() string {
	return font.name
}

// Encoder returns the font's text encoder.
func (font StdFont) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// GetGlyphCharMetrics returns character metrics for a given glyph.
func (font StdFont) GetGlyphCharMetrics(glyph GlyphName) (CharMetrics, bool) {
	metrics, has := font.metrics[glyph]
	if !has {
		return metrics, false
	}

	return metrics, true
}

// GetMetricsTable is a method specific to standard fonts. It returns the metrics table of all glyphs.
// Caller should not modify the table.
func (font StdFont) GetMetricsTable() map[GlyphName]CharMetrics {
	return font.metrics
}

// ToPdfObject returns a primitive PDF object representation of the font.
func (font StdFont) ToPdfObject() core.PdfObject {
	fontDict := core.MakeDict()
	fontDict.Set("Type", core.MakeName("Font"))
	fontDict.Set("Subtype", core.MakeName("Type1"))
	fontDict.Set("BaseFont", core.MakeName(font.name))
	fontDict.Set("Encoding", font.encoder.ToPdfObject())

	return core.MakeIndirectObject(fontDict)
}

// type1CommonGlyphs is list of common glyph names for some Type1. Used to unpack character metrics.
var type1CommonGlyphs = []textencoding.GlyphName{
	"A", "AE", "Aacute", "Abreve", "Acircumflex",
	"Adieresis", "Agrave", "Amacron", "Aogonek", "Aring",
	"Atilde", "B", "C", "Cacute", "Ccaron",
	"Ccedilla", "D", "Dcaron", "Dcroat", "Delta",
	"E", "Eacute", "Ecaron", "Ecircumflex", "Edieresis",
	"Edotaccent", "Egrave", "Emacron", "Eogonek", "Eth",
	"Euro", "F", "G", "Gbreve", "Gcommaaccent",
	"H", "I", "Iacute", "Icircumflex", "Idieresis",
	"Idotaccent", "Igrave", "Imacron", "Iogonek", "J",
	"K", "Kcommaaccent", "L", "Lacute", "Lcaron",
	"Lcommaaccent", "Lslash", "M", "N", "Nacute",
	"Ncaron", "Ncommaaccent", "Ntilde", "O", "OE",
	"Oacute", "Ocircumflex", "Odieresis", "Ograve", "Ohungarumlaut",
	"Omacron", "Oslash", "Otilde", "P", "Q",
	"R", "Racute", "Rcaron", "Rcommaaccent", "S",
	"Sacute", "Scaron", "Scedilla", "Scommaaccent", "T",
	"Tcaron", "Tcommaaccent", "Thorn", "U", "Uacute",
	"Ucircumflex", "Udieresis", "Ugrave", "Uhungarumlaut", "Umacron",
	"Uogonek", "Uring", "V", "W", "X",
	"Y", "Yacute", "Ydieresis", "Z", "Zacute",
	"Zcaron", "Zdotaccent", "a", "aacute", "abreve",
	"acircumflex", "acute", "adieresis", "ae", "agrave",
	"amacron", "ampersand", "aogonek", "aring", "asciicircum",
	"asciitilde", "asterisk", "at", "atilde", "b",
	"backslash", "bar", "braceleft", "braceright", "bracketleft",
	"bracketright", "breve", "brokenbar", "bullet", "c",
	"cacute", "caron", "ccaron", "ccedilla", "cedilla",
	"cent", "circumflex", "colon", "comma", "commaaccent",
	"copyright", "currency", "d", "dagger", "daggerdbl",
	"dcaron", "dcroat", "degree", "dieresis", "divide",
	"dollar", "dotaccent", "dotlessi", "e", "eacute",
	"ecaron", "ecircumflex", "edieresis", "edotaccent", "egrave",
	"eight", "ellipsis", "emacron", "emdash", "endash",
	"eogonek", "equal", "eth", "exclam", "exclamdown",
	"f", "fi", "five", "fl", "florin",
	"four", "fraction", "g", "gbreve", "gcommaaccent",
	"germandbls", "grave", "greater", "greaterequal", "guillemotleft",
	"guillemotright", "guilsinglleft", "guilsinglright", "h", "hungarumlaut",
	"hyphen", "i", "iacute", "icircumflex", "idieresis",
	"igrave", "imacron", "iogonek", "j", "k",
	"kcommaaccent", "l", "lacute", "lcaron", "lcommaaccent",
	"less", "lessequal", "logicalnot", "lozenge", "lslash",
	"m", "macron", "minus", "mu", "multiply",
	"n", "nacute", "ncaron", "ncommaaccent", "nine",
	"notequal", "ntilde", "numbersign", "o", "oacute",
	"ocircumflex", "odieresis", "oe", "ogonek", "ograve",
	"ohungarumlaut", "omacron", "one", "onehalf", "onequarter",
	"onesuperior", "ordfeminine", "ordmasculine", "oslash", "otilde",
	"p", "paragraph", "parenleft", "parenright", "partialdiff",
	"percent", "period", "periodcentered", "perthousand", "plus",
	"plusminus", "q", "question", "questiondown", "quotedbl",
	"quotedblbase", "quotedblleft", "quotedblright", "quoteleft", "quoteright",
	"quotesinglbase", "quotesingle", "r", "racute", "radical",
	"rcaron", "rcommaaccent", "registered", "ring", "s",
	"sacute", "scaron", "scedilla", "scommaaccent", "section",
	"semicolon", "seven", "six", "slash", "space",
	"sterling", "summation", "t", "tcaron", "tcommaaccent",
	"thorn", "three", "threequarters", "threesuperior", "tilde",
	"trademark", "two", "twosuperior", "u", "uacute",
	"ucircumflex", "udieresis", "ugrave", "uhungarumlaut", "umacron",
	"underscore", "uogonek", "uring", "v", "w",
	"x", "y", "yacute", "ydieresis", "yen",
	"z", "zacute", "zcaron", "zdotaccent", "zero",
}
