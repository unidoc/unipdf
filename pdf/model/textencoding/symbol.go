/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// Encoding for Symbol font.
type SymbolEncoder struct {
}

func NewSymbolEncoder() SymbolEncoder {
	encoder := SymbolEncoder{}
	return encoder
}

// Convert a raw utf8 string (series of runes) to an encoded string (series of character codes) to be used in PDF.
func (enc SymbolEncoder) Encode(raw string) string {
	encoded := []byte{}
	for _, rune := range raw {
		code, found := enc.RuneToCharcode(rune)
		if !found {
			continue
		}

		encoded = append(encoded, code)
	}

	return string(encoded)
}

// Conversion between character code and glyph name.
// The bool return flag is true if there was a match, and false otherwise.
func (enc SymbolEncoder) CharcodeToGlyph(code byte) (string, bool) {
	glyph, has := symbolEncodingCharcodeToGlyphMap[code]
	if !has {
		common.Log.Debug("Symbol encoding error: unable to find charcode->glyph entry (%v)", code)
		return "", false
	}
	return glyph, true
}

// Conversion between glyph name and character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc SymbolEncoder) GlyphToCharcode(glyph string) (byte, bool) {
	code, found := symbolEncodingGlyphToCharcodeMap[glyph]
	if !found {
		common.Log.Debug("Symbol encoding error: unable to find glyph->charcode entry (%s)", glyph)
		return 0, false
	}

	return code, found
}

// Convert rune to character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc SymbolEncoder) RuneToCharcode(val rune) (byte, bool) {
	glyph, found := runeToGlyph(val, glyphlistRuneToGlyphMap)
	if !found {
		common.Log.Debug("Symbol encoding error: unable to find rune->glyph entry (%v)", val)
		return 0, false
	}

	code, found := symbolEncodingGlyphToCharcodeMap[glyph]
	if !found {
		common.Log.Debug("Symbol encoding error: unable to find glyph->charcode entry (%s)", glyph)
		return 0, false
	}

	return code, true
}

// Convert character code to rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc SymbolEncoder) CharcodeToRune(charcode byte) (rune, bool) {
	glyph, found := symbolEncodingCharcodeToGlyphMap[charcode]
	if !found {
		common.Log.Debug("Symbol encoding error: unable to find charcode->glyph entry (%d)", charcode)
		return 0, false
	}

	val, found := glyphToRune(glyph, glyphlistGlyphToRuneMap)
	if !found {
		return 0, false
	}

	return val, true
}

// Convert rune to glyph name.
// The bool return flag is true if there was a match, and false otherwise.
func (enc SymbolEncoder) RuneToGlyph(val rune) (string, bool) {
	return runeToGlyph(val, glyphlistRuneToGlyphMap)
}

// Convert glyph to rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc SymbolEncoder) GlyphToRune(glyph string) (rune, bool) {
	return glyphToRune(glyph, glyphlistGlyphToRuneMap)
}

// Convert to PDF Object.
func (enc SymbolEncoder) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()
	dict.Set("Type", core.MakeName("Encoding"))

	// Returning an empty Encoding object with no differences. Indicates that we are using the font's built-in
	// encoding.
	return core.MakeIndirectObject(dict)
}

// Charcode to Glyph map (Symbol encoding)
var symbolEncodingCharcodeToGlyphMap map[byte]string = map[byte]string{
	32:  "space",
	33:  "exclam",
	34:  "universal",
	35:  "numbersign",
	36:  "existential",
	37:  "percent",
	38:  "ampersand",
	39:  "suchthat",
	40:  "parenleft",
	41:  "parenright",
	42:  "asteriskmath",
	43:  "plus",
	44:  "comma",
	45:  "minus",
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
	64:  "congruent",
	65:  "Alpha",
	66:  "Beta",
	67:  "Chi",
	68:  "Delta",
	69:  "Epsilon",
	70:  "Phi",
	71:  "Gamma",
	72:  "Eta",
	73:  "Iota",
	74:  "theta1",
	75:  "Kappa",
	76:  "Lambda",
	77:  "Mu",
	78:  "Nu",
	79:  "Omicron",
	80:  "Pi",
	81:  "Theta",
	82:  "Rho",
	83:  "Sigma",
	84:  "Tau",
	85:  "Upsilon",
	86:  "sigma1",
	87:  "Omega",
	88:  "Xi",
	89:  "Psi",
	90:  "Zeta",
	91:  "bracketleft",
	92:  "therefore",
	93:  "bracketright",
	94:  "perpendicular",
	95:  "underscore",
	96:  "radicalex",
	97:  "alpha",
	98:  "beta",
	99:  "chi",
	100: "delta",
	101: "epsilon",
	102: "phi",
	103: "gamma",
	104: "eta",
	105: "iota",
	106: "phi1",
	107: "kappa",
	108: "lambda",
	109: "mu",
	110: "nu",
	111: "omicron",
	112: "pi",
	113: "theta",
	114: "rho",
	115: "sigma",
	116: "tau",
	117: "upsilon",
	118: "omega1",
	119: "omega",
	120: "xi",
	121: "psi",
	122: "zeta",
	123: "braceleft",
	124: "bar",
	125: "braceright",
	126: "similar",
	160: "Euro",
	161: "Upsilon1",
	162: "minute",
	163: "lessequal",
	164: "fraction",
	165: "infinity",
	166: "florin",
	167: "club",
	168: "diamond",
	169: "heart",
	170: "spade",
	171: "arrowboth",
	172: "arrowleft",
	173: "arrowup",
	174: "arrowright",
	175: "arrowdown",
	176: "degree",
	177: "plusminus",
	178: "second",
	179: "greaterequal",
	180: "multiply",
	181: "proportional",
	182: "partialdiff",
	183: "bullet",
	184: "divide",
	185: "notequal",
	186: "equivalence",
	187: "approxequal",
	188: "ellipsis",
	189: "arrowvertex",
	190: "arrowhorizex",
	191: "carriagereturn",
	192: "aleph",
	193: "Ifraktur",
	194: "Rfraktur",
	195: "weierstrass",
	196: "circlemultiply",
	197: "circleplus",
	198: "emptyset",
	199: "intersection",
	200: "union",
	201: "propersuperset",
	202: "reflexsuperset",
	203: "notsubset",
	204: "propersubset",
	205: "reflexsubset",
	206: "element",
	207: "notelement",
	208: "angle",
	209: "gradient",
	210: "registerserif",
	211: "copyrightserif",
	212: "trademarkserif",
	213: "product",
	214: "radical",
	215: "dotmath",
	216: "logicalnot",
	217: "logicaland",
	218: "logicalor",
	219: "arrowdblboth",
	220: "arrowdblleft",
	221: "arrowdblup",
	222: "arrowdblright",
	223: "arrowdbldown",
	224: "lozenge",
	225: "angleleft",
	226: "registersans",
	227: "copyrightsans",
	228: "trademarksans",
	229: "summation",
	230: "parenlefttp",
	231: "parenleftex",
	232: "parenleftbt",
	233: "bracketlefttp",
	234: "bracketleftex",
	235: "bracketleftbt",
	236: "bracelefttp",
	237: "braceleftmid",
	238: "braceleftbt",
	239: "braceex",
	241: "angleright",
	242: "integral",
	243: "integraltp",
	244: "integralex",
	245: "integralbt",
	246: "parenrighttp",
	247: "parenrightex",
	248: "parenrightbt",
	249: "bracketrighttp",
	250: "bracketrightex",
	251: "bracketrightbt",
	252: "bracerighttp",
	253: "bracerightmid",
	254: "bracerightbt",
}

// Glyph to charcode map (Symbol encoding).
var symbolEncodingGlyphToCharcodeMap map[string]byte = map[string]byte{
	"space":          32,
	"exclam":         33,
	"universal":      34,
	"numbersign":     35,
	"existential":    36,
	"percent":        37,
	"ampersand":      38,
	"suchthat":       39,
	"parenleft":      40,
	"parenright":     41,
	"asteriskmath":   42,
	"plus":           43,
	"comma":          44,
	"minus":          45,
	"period":         46,
	"slash":          47,
	"zero":           48,
	"one":            49,
	"two":            50,
	"three":          51,
	"four":           52,
	"five":           53,
	"six":            54,
	"seven":          55,
	"eight":          56,
	"nine":           57,
	"colon":          58,
	"semicolon":      59,
	"less":           60,
	"equal":          61,
	"greater":        62,
	"question":       63,
	"congruent":      64,
	"Alpha":          65,
	"Beta":           66,
	"Chi":            67,
	"Delta":          68,
	"Epsilon":        69,
	"Phi":            70,
	"Gamma":          71,
	"Eta":            72,
	"Iota":           73,
	"theta1":         74,
	"Kappa":          75,
	"Lambda":         76,
	"Mu":             77,
	"Nu":             78,
	"Omicron":        79,
	"Pi":             80,
	"Theta":          81,
	"Rho":            82,
	"Sigma":          83,
	"Tau":            84,
	"Upsilon":        85,
	"sigma1":         86,
	"Omega":          87,
	"Xi":             88,
	"Psi":            89,
	"Zeta":           90,
	"bracketleft":    91,
	"therefore":      92,
	"bracketright":   93,
	"perpendicular":  94,
	"underscore":     95,
	"radicalex":      96,
	"alpha":          97,
	"beta":           98,
	"chi":            99,
	"delta":          100,
	"epsilon":        101,
	"phi":            102,
	"gamma":          103,
	"eta":            104,
	"iota":           105,
	"phi1":           106,
	"kappa":          107,
	"lambda":         108,
	"mu":             109,
	"nu":             110,
	"omicron":        111,
	"pi":             112,
	"theta":          113,
	"rho":            114,
	"sigma":          115,
	"tau":            116,
	"upsilon":        117,
	"omega1":         118,
	"omega":          119,
	"xi":             120,
	"psi":            121,
	"zeta":           122,
	"braceleft":      123,
	"bar":            124,
	"braceright":     125,
	"similar":        126,
	"Euro":           160,
	"Upsilon1":       161,
	"minute":         162,
	"lessequal":      163,
	"fraction":       164,
	"infinity":       165,
	"florin":         166,
	"club":           167,
	"diamond":        168,
	"heart":          169,
	"spade":          170,
	"arrowboth":      171,
	"arrowleft":      172,
	"arrowup":        173,
	"arrowright":     174,
	"arrowdown":      175,
	"degree":         176,
	"plusminus":      177,
	"second":         178,
	"greaterequal":   179,
	"multiply":       180,
	"proportional":   181,
	"partialdiff":    182,
	"bullet":         183,
	"divide":         184,
	"notequal":       185,
	"equivalence":    186,
	"approxequal":    187,
	"ellipsis":       188,
	"arrowvertex":    189,
	"arrowhorizex":   190,
	"carriagereturn": 191,
	"aleph":          192,
	"Ifraktur":       193,
	"Rfraktur":       194,
	"weierstrass":    195,
	"circlemultiply": 196,
	"circleplus":     197,
	"emptyset":       198,
	"intersection":   199,
	"union":          200,
	"propersuperset": 201,
	"reflexsuperset": 202,
	"notsubset":      203,
	"propersubset":   204,
	"reflexsubset":   205,
	"element":        206,
	"notelement":     207,
	"angle":          208,
	"gradient":       209,
	"registerserif":  210,
	"copyrightserif": 211,
	"trademarkserif": 212,
	"product":        213,
	"radical":        214,
	"dotmath":        215,
	"logicalnot":     216,
	"logicaland":     217,
	"logicalor":      218,
	"arrowdblboth":   219,
	"arrowdblleft":   220,
	"arrowdblup":     221,
	"arrowdblright":  222,
	"arrowdbldown":   223,
	"lozenge":        224,
	"angleleft":      225,
	"registersans":   226,
	"copyrightsans":  227,
	"trademarksans":  228,
	"summation":      229,
	"parenlefttp":    230,
	"parenleftex":    231,
	"parenleftbt":    232,
	"bracketlefttp":  233,
	"bracketleftex":  234,
	"bracketleftbt":  235,
	"bracelefttp":    236,
	"braceleftmid":   237,
	"braceleftbt":    238,
	"braceex":        239,
	"angleright":     241,
	"integral":       242,
	"integraltp":     243,
	"integralex":     244,
	"integralbt":     245,
	"parenrighttp":   246,
	"parenrightex":   247,
	"parenrightbt":   248,
	"bracketrighttp": 249,
	"bracketrightex": 250,
	"bracketrightbt": 251,
	"bracerighttp":   252,
	"bracerightmid":  253,
	"bracerightbt":   254,
}
