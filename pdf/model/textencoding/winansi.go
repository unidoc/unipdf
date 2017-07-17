/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// WinAnsiEncoding.
type WinAnsiEncoder struct {
}

func NewWinAnsiTextEncoder() WinAnsiEncoder {
	encoder := WinAnsiEncoder{}
	return encoder
}

func (winenc WinAnsiEncoder) ToPdfObject() core.PdfObject {
	return core.MakeName("WinAnsiEncoding")
}

// Convert a raw utf8 string (series of runes) to an encoded string (series of character codes) to be used in PDF.
func (winenc WinAnsiEncoder) Encode(raw string) string {
	encoded := []byte{}
	for _, rune := range raw {
		code, has := winenc.RuneToCharcode(rune)
		if has {
			encoded = append(encoded, code)
		}
	}

	return string(encoded)
}

// Conversion between character code and glyph name.
// The bool return flag is true if there was a match, and false otherwise.
func (winenc WinAnsiEncoder) CharcodeToGlyph(code byte) (string, bool) {
	glyph, has := winansiEncodingCharcodeToGlyphMap[code]
	if !has {
		common.Log.Debug("Charcode -> Glyph error: charcode not found: %d\n", code)
		return "", false
	}
	return glyph, true
}

// Conversion between glyph name and character code.
// The bool return flag is true if there was a match, and false otherwise.
func (winenc WinAnsiEncoder) GlyphToCharcode(glyph string) (byte, bool) {
	code, found := winansiEncodingGlyphToCharcodeMap[glyph]
	if !found {
		common.Log.Debug("Glyph -> Charcode error: glyph not found: %s\n", glyph)
		return 0, false
	}

	return code, true
}

// Convert rune to character code.
// The bool return flag is true if there was a match, and false otherwise.
func (winenc WinAnsiEncoder) RuneToCharcode(val rune) (byte, bool) {
	glyph, found := winenc.RuneToGlyph(val)
	if !found {
		return 0, false
	}

	code, found := winansiEncodingGlyphToCharcodeMap[glyph]
	if !found {
		common.Log.Debug("Glyph -> Charcode error: glyph not found %s\n", glyph)
		return 0, false
	}

	return code, true
}

// Convert character code to rune.
// The bool return flag is true if there was a match, and false otherwise.
func (winenc WinAnsiEncoder) CharcodeToRune(charcode byte) (rune, bool) {
	glyph, found := winansiEncodingCharcodeToGlyphMap[charcode]
	if !found {
		common.Log.Debug("Charcode -> Glyph error: charcode not found: %d\n", charcode)
		return 0, false
	}

	ucode, found := glyphToRune(glyph, glyphlistGlyphToRuneMap)
	if !found {
		return 0, false
	}

	return ucode, true
}

// Convert rune to glyph name.
// The bool return flag is true if there was a match, and false otherwise.
func (winenc WinAnsiEncoder) RuneToGlyph(val rune) (string, bool) {
	return runeToGlyph(val, glyphlistRuneToGlyphMap)
}

// Convert glyph to rune.
// The bool return flag is true if there was a match, and false otherwise.
func (winenc WinAnsiEncoder) GlyphToRune(glyph string) (rune, bool) {
	return glyphToRune(glyph, glyphlistGlyphToRuneMap)
}

// Charcode to glyph name map (WinAnsiEncoding).
var winansiEncodingCharcodeToGlyphMap = map[byte]string{
	32:  "space",
	33:  "exclam",
	34:  "quotedbl",
	35:  "numbersign",
	36:  "dollar",
	37:  "percent",
	38:  "ampersand",
	39:  "quotesingle",
	40:  "parenleft",
	41:  "parenright",
	42:  "asterisk",
	43:  "plus",
	44:  "comma",
	45:  "hyphen",
	46:  "period",
	47:  "slash",
	48:  "zero",
	49:  "one",
	50:  "two",
	51:  "three",
	52:  "four",
	53:  "five",
	54:  "six",
	55:  "seven",
	56:  "eight",
	57:  "nine",
	58:  "colon",
	59:  "semicolon",
	60:  "less",
	61:  "equal",
	62:  "greater",
	63:  "question",
	64:  "at",
	65:  "A",
	66:  "B",
	67:  "C",
	68:  "D",
	69:  "E",
	70:  "F",
	71:  "G",
	72:  "H",
	73:  "I",
	74:  "J",
	75:  "K",
	76:  "L",
	77:  "M",
	78:  "N",
	79:  "O",
	80:  "P",
	81:  "Q",
	82:  "R",
	83:  "S",
	84:  "T",
	85:  "U",
	86:  "V",
	87:  "W",
	88:  "X",
	89:  "Y",
	90:  "Z",
	91:  "bracketleft",
	92:  "backslash",
	93:  "bracketright",
	94:  "asciicircum",
	95:  "underscore",
	96:  "grave",
	97:  "a",
	98:  "b",
	99:  "c",
	100: "d",
	101: "e",
	102: "f",
	103: "g",
	104: "h",
	105: "i",
	106: "j",
	107: "k",
	108: "l",
	109: "m",
	110: "n",
	111: "o",
	112: "p",
	113: "q",
	114: "r",
	115: "s",
	116: "t",
	117: "u",
	118: "v",
	119: "w",
	120: "x",
	121: "y",
	122: "z",
	123: "braceleft",
	124: "bar",
	125: "braceright",
	126: "asciitilde",
	127: "bullet",
	128: "Euro",
	129: "bullet",
	130: "quotesinglbase",
	131: "florin",
	132: "quotedblbase",
	133: "ellipsis",
	134: "dagger",
	135: "daggerdbl",
	136: "circumflex",
	137: "perthousand",
	138: "Scaron",
	139: "guilsinglleft",
	140: "OE",
	141: "bullet",
	142: "Zcaron",
	143: "bullet",
	144: "bullet",
	145: "quoteleft",
	146: "quoteright",
	147: "quotedblleft",
	148: "quotedblright",
	149: "bullet",
	150: "endash",
	151: "emdash",
	152: "tilde",
	153: "trademark",
	154: "scaron",
	155: "guilsinglright",
	156: "oe",
	157: "bullet",
	158: "zcaron",
	159: "Ydieresis",
	160: "space",
	161: "exclamdown",
	162: "cent",
	163: "sterling",
	164: "currency",
	165: "yen",
	166: "brokenbar",
	167: "section",
	168: "dieresis",
	169: "copyright",
	170: "ordfeminine",
	171: "guillemotleft",
	172: "logicalnot",
	173: "hyphen",
	174: "registered",
	175: "macron",
	176: "degree",
	177: "plusminus",
	178: "twosuperior",
	179: "threesuperior",
	180: "acute",
	181: "mu",
	182: "paragraph",
	183: "periodcentered",
	184: "cedilla",
	185: "onesuperior",
	186: "ordmasculine",
	187: "guillemotright",
	188: "onequarter",
	189: "onehalf",
	190: "threequarters",
	191: "questiondown",
	192: "Agrave",
	193: "Aacute",
	194: "Acircumflex",
	195: "Atilde",
	196: "Adieresis",
	197: "Aring",
	198: "AE",
	199: "Ccedilla",
	200: "Egrave",
	201: "Eacute",
	202: "Ecircumflex",
	203: "Edieresis",
	204: "Igrave",
	205: "Iacute",
	206: "Icircumflex",
	207: "Idieresis",
	208: "Eth",
	209: "Ntilde",
	210: "Ograve",
	211: "Oacute",
	212: "Ocircumflex",
	213: "Otilde",
	214: "Odieresis",
	215: "multiply",
	216: "Oslash",
	217: "Ugrave",
	218: "Uacute",
	219: "Ucircumflex",
	220: "Udieresis",
	221: "Yacute",
	222: "Thorn",
	223: "germandbls",
	224: "agrave",
	225: "aacute",
	226: "acircumflex",
	227: "atilde",
	228: "adieresis",
	229: "aring",
	230: "ae",
	231: "ccedilla",
	232: "egrave",
	233: "eacute",
	234: "ecircumflex",
	235: "edieresis",
	236: "igrave",
	237: "iacute",
	238: "icircumflex",
	239: "idieresis",
	240: "eth",
	241: "ntilde",
	242: "ograve",
	243: "oacute",
	244: "ocircumflex",
	245: "otilde",
	246: "odieresis",
	247: "divide",
	248: "oslash",
	249: "ugrave",
	250: "uacute",
	251: "ucircumflex",
	252: "udieresis",
	253: "yacute",
	254: "thorn",
	255: "ydieresis",
}

// Glyph to charcode map (WinAnsiEncoding).
var winansiEncodingGlyphToCharcodeMap = map[string]byte{
	"space":        32,
	"exclam":       33,
	"quotedbl":     34,
	"numbersign":   35,
	"dollar":       36,
	"percent":      37,
	"ampersand":    38,
	"quotesingle":  39,
	"parenleft":    40,
	"parenright":   41,
	"asterisk":     42,
	"plus":         43,
	"comma":        44,
	"hyphen":       45,
	"period":       46,
	"slash":        47,
	"zero":         48,
	"one":          49,
	"two":          50,
	"three":        51,
	"four":         52,
	"five":         53,
	"six":          54,
	"seven":        55,
	"eight":        56,
	"nine":         57,
	"colon":        58,
	"semicolon":    59,
	"less":         60,
	"equal":        61,
	"greater":      62,
	"question":     63,
	"at":           64,
	"A":            65,
	"B":            66,
	"C":            67,
	"D":            68,
	"E":            69,
	"F":            70,
	"G":            71,
	"H":            72,
	"I":            73,
	"J":            74,
	"K":            75,
	"L":            76,
	"M":            77,
	"N":            78,
	"O":            79,
	"P":            80,
	"Q":            81,
	"R":            82,
	"S":            83,
	"T":            84,
	"U":            85,
	"V":            86,
	"W":            87,
	"X":            88,
	"Y":            89,
	"Z":            90,
	"bracketleft":  91,
	"backslash":    92,
	"bracketright": 93,
	"asciicircum":  94,
	"underscore":   95,
	"grave":        96,
	"a":            97,
	"b":            98,
	"c":            99,
	"d":            100,
	"e":            101,
	"f":            102,
	"g":            103,
	"h":            104,
	"i":            105,
	"j":            106,
	"k":            107,
	"l":            108,
	"m":            109,
	"n":            110,
	"o":            111,
	"p":            112,
	"q":            113,
	"r":            114,
	"s":            115,
	"t":            116,
	"u":            117,
	"v":            118,
	"w":            119,
	"x":            120,
	"y":            121,
	"z":            122,
	"braceleft":    123,
	"bar":          124,
	"braceright":   125,
	"asciitilde":   126,
	"bullet":       127,
	"Euro":         128,
	//"bullet":         129,
	"quotesinglbase": 130,
	"florin":         131,
	"quotedblbase":   132,
	"ellipsis":       133,
	"dagger":         134,
	"daggerdbl":      135,
	"circumflex":     136,
	"perthousand":    137,
	"Scaron":         138,
	"guilsinglleft":  139,
	"OE":             140,
	//"bullet":         141,
	"Zcaron": 142,
	//"bullet":         143,
	//"bullet":         144,
	"quoteleft":     145,
	"quoteright":    146,
	"quotedblleft":  147,
	"quotedblright": 148,
	//"bullet":         149,
	"endash":         150,
	"emdash":         151,
	"tilde":          152,
	"trademark":      153,
	"scaron":         154,
	"guilsinglright": 155,
	"oe":             156,
	//"bullet":         157,
	"zcaron":    158,
	"Ydieresis": 159,
	//"space":          160,
	"exclamdown":    161,
	"cent":          162,
	"sterling":      163,
	"currency":      164,
	"yen":           165,
	"brokenbar":     166,
	"section":       167,
	"dieresis":      168,
	"copyright":     169,
	"ordfeminine":   170,
	"guillemotleft": 171,
	"logicalnot":    172,
	//"hyphen":         173,
	"registered":     174,
	"macron":         175,
	"degree":         176,
	"plusminus":      177,
	"twosuperior":    178,
	"threesuperior":  179,
	"acute":          180,
	"mu":             181,
	"paragraph":      182,
	"periodcentered": 183,
	"cedilla":        184,
	"onesuperior":    185,
	"ordmasculine":   186,
	"guillemotright": 187,
	"onequarter":     188,
	"onehalf":        189,
	"threequarters":  190,
	"questiondown":   191,
	"Agrave":         192,
	"Aacute":         193,
	"Acircumflex":    194,
	"Atilde":         195,
	"Adieresis":      196,
	"Aring":          197,
	"AE":             198,
	"Ccedilla":       199,
	"Egrave":         200,
	"Eacute":         201,
	"Ecircumflex":    202,
	"Edieresis":      203,
	"Igrave":         204,
	"Iacute":         205,
	"Icircumflex":    206,
	"Idieresis":      207,
	"Eth":            208,
	"Ntilde":         209,
	"Ograve":         210,
	"Oacute":         211,
	"Ocircumflex":    212,
	"Otilde":         213,
	"Odieresis":      214,
	"multiply":       215,
	"Oslash":         216,
	"Ugrave":         217,
	"Uacute":         218,
	"Ucircumflex":    219,
	"Udieresis":      220,
	"Yacute":         221,
	"Thorn":          222,
	"germandbls":     223,
	"agrave":         224,
	"aacute":         225,
	"acircumflex":    226,
	"atilde":         227,
	"adieresis":      228,
	"aring":          229,
	"ae":             230,
	"ccedilla":       231,
	"egrave":         232,
	"eacute":         233,
	"ecircumflex":    234,
	"edieresis":      235,
	"igrave":         236,
	"iacute":         237,
	"icircumflex":    238,
	"idieresis":      239,
	"eth":            240,
	"ntilde":         241,
	"ograve":         242,
	"oacute":         243,
	"ocircumflex":    244,
	"otilde":         245,
	"odieresis":      246,
	"divide":         247,
	"oslash":         248,
	"ugrave":         249,
	"uacute":         250,
	"ucircumflex":    251,
	"udieresis":      252,
	"yacute":         253,
	"thorn":          254,
	"ydieresis":      255,
}
