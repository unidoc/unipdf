/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// Encoding for ZapfDingbats font.
type ZapfDingbatsEncoder struct {
}

func NewZapfDingbatsEncoder() ZapfDingbatsEncoder {
	encoder := ZapfDingbatsEncoder{}
	return encoder
}

// Convert a raw utf8 string (series of runes) to an encoded string (series of character codes) to be used in PDF.
func (enc ZapfDingbatsEncoder) Encode(raw string) string {
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
func (enc ZapfDingbatsEncoder) CharcodeToGlyph(code byte) (string, bool) {
	glyph, has := zapfDingbatsEncodingCharcodeToGlyphMap[code]
	if !has {
		common.Log.Debug("ZapfDingbats encoding error: unable to find charcode->glyph entry (%v)", code)
		return "", false
	}
	return glyph, true
}

// Conversion between glyph name and character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc ZapfDingbatsEncoder) GlyphToCharcode(glyph string) (byte, bool) {
	code, found := zapfDingbatsEncodingGlyphToCharcodeMap[glyph]
	if !found {
		common.Log.Debug("ZapfDingbats encoding error: unable to find glyph->charcode entry (%s)", glyph)
		return 0, false
	}

	return code, found
}

// Convert rune to character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc ZapfDingbatsEncoder) RuneToCharcode(val rune) (byte, bool) {
	glyph, found := enc.RuneToGlyph(val)
	if !found {
		common.Log.Debug("ZapfDingbats encoding error: unable to find rune->glyph entry (%v)", val)
		return 0, false
	}

	code, found := zapfDingbatsEncodingGlyphToCharcodeMap[glyph]
	if !found {
		common.Log.Debug("ZapfDingbats encoding error: unable to find glyph->charcode entry (%s)", glyph)
		return 0, false
	}

	return code, true
}

// Convert character code to rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc ZapfDingbatsEncoder) CharcodeToRune(charcode byte) (rune, bool) {
	glyph, found := zapfDingbatsEncodingCharcodeToGlyphMap[charcode]
	if !found {
		common.Log.Debug("ZapfDingbats encoding error: unable to find charcode->glyph entry (%d)", charcode)
		return 0, false
	}

	return enc.GlyphToRune(glyph)
}

// Convert rune to glyph name.
// The bool return flag is true if there was a match, and false otherwise.
func (enc ZapfDingbatsEncoder) RuneToGlyph(val rune) (string, bool) {
	// Seek in the zapfdingbats list first.
	glyph, found := runeToGlyph(val, zapfdingbatsRuneToGlyphMap)
	if !found {
		// Then revert to glyphlist if not found.
		glyph, found = runeToGlyph(val, glyphlistRuneToGlyphMap)
		if !found {
			common.Log.Debug("ZapfDingbats encoding error: unable to find rune->glyph entry (%v)", val)
			return "", false
		}
	}

	return glyph, true
}

// Convert glyph to rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc ZapfDingbatsEncoder) GlyphToRune(glyph string) (rune, bool) {
	// Seek in the zapfdingbats list first.
	val, found := glyphToRune(glyph, zapfdingbatsGlyphToRuneMap)
	if !found {
		// Then revert to glyphlist if not found.
		val, found = glyphToRune(glyph, glyphlistGlyphToRuneMap)
		if !found {
			common.Log.Debug("Symbol encoding error: unable to find glyph->rune entry (%v)", glyph)
			return 0, false
		}
	}

	return val, true
}

// Convert to PDF Object.
func (enc ZapfDingbatsEncoder) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()
	dict.Set("Type", core.MakeName("Encoding"))

	// Returning an empty Encoding object with no differences. Indicates that we are using the font's built-in
	// encoding.
	return core.MakeIndirectObject(dict)
}

var zapfDingbatsEncodingCharcodeToGlyphMap = map[byte]string{
	32:  "space",
	33:  "a1",
	34:  "a2",
	35:  "a202",
	36:  "a3",
	37:  "a4",
	38:  "a5",
	39:  "a119",
	40:  "a118",
	41:  "a117",
	42:  "a11",
	43:  "a12",
	44:  "a13",
	45:  "a14",
	46:  "a15",
	47:  "a16",
	48:  "a105",
	49:  "a17",
	50:  "a18",
	51:  "a19",
	52:  "a20",
	53:  "a21",
	54:  "a22",
	55:  "a23",
	56:  "a24",
	57:  "a25",
	58:  "a26",
	59:  "a27",
	60:  "a28",
	61:  "a6",
	62:  "a7",
	63:  "a8",
	64:  "a9",
	65:  "a10",
	66:  "a29",
	67:  "a30",
	68:  "a31",
	69:  "a32",
	70:  "a33",
	71:  "a34",
	72:  "a35",
	73:  "a36",
	74:  "a37",
	75:  "a38",
	76:  "a39",
	77:  "a40",
	78:  "a41",
	79:  "a42",
	80:  "a43",
	81:  "a44",
	82:  "a45",
	83:  "a46",
	84:  "a47",
	85:  "a48",
	86:  "a49",
	87:  "a50",
	88:  "a51",
	89:  "a52",
	90:  "a53",
	91:  "a54",
	92:  "a55",
	93:  "a56",
	94:  "a57",
	95:  "a58",
	96:  "a59",
	97:  "a60",
	98:  "a61",
	99:  "a62",
	100: "a63",
	101: "a64",
	102: "a65",
	103: "a66",
	104: "a67",
	105: "a68",
	106: "a69",
	107: "a70",
	108: "a71",
	109: "a72",
	110: "a73",
	111: "a74",
	112: "a203",
	113: "a75",
	114: "a204",
	115: "a76",
	116: "a77",
	117: "a78",
	118: "a79",
	119: "a81",
	120: "a82",
	121: "a83",
	122: "a84",
	123: "a97",
	124: "a98",
	125: "a99",
	126: "a100",
	128: "a89",
	129: "a90",
	130: "a93",
	131: "a94",
	132: "a91",
	133: "a92",
	134: "a205",
	135: "a85",
	136: "a206",
	137: "a86",
	138: "a87",
	139: "a88",
	140: "a95",
	141: "a96",
	161: "a101",
	162: "a102",
	163: "a103",
	164: "a104",
	165: "a106",
	166: "a107",
	167: "a108",
	168: "a112",
	169: "a111",
	170: "a110",
	171: "a109",
	172: "a120",
	173: "a121",
	174: "a122",
	175: "a123",
	176: "a124",
	177: "a125",
	178: "a126",
	179: "a127",
	180: "a128",
	181: "a129",
	182: "a130",
	183: "a131",
	184: "a132",
	185: "a133",
	186: "a134",
	187: "a135",
	188: "a136",
	189: "a137",
	190: "a138",
	191: "a139",
	192: "a140",
	193: "a141",
	194: "a142",
	195: "a143",
	196: "a144",
	197: "a145",
	198: "a146",
	199: "a147",
	200: "a148",
	201: "a149",
	202: "a150",
	203: "a151",
	204: "a152",
	205: "a153",
	206: "a154",
	207: "a155",
	208: "a156",
	209: "a157",
	210: "a158",
	211: "a159",
	212: "a160",
	213: "a161",
	214: "a163",
	215: "a164",
	216: "a196",
	217: "a165",
	218: "a192",
	219: "a166",
	220: "a167",
	221: "a168",
	222: "a169",
	223: "a170",
	224: "a171",
	225: "a172",
	226: "a173",
	227: "a162",
	228: "a174",
	229: "a175",
	230: "a176",
	231: "a177",
	232: "a178",
	233: "a179",
	234: "a193",
	235: "a180",
	236: "a199",
	237: "a181",
	238: "a200",
	239: "a182",
	241: "a201",
	242: "a183",
	243: "a184",
	244: "a197",
	245: "a185",
	246: "a194",
	247: "a198",
	248: "a186",
	249: "a195",
	250: "a187",
	251: "a188",
	252: "a189",
	253: "a190",
	254: "a191",
}

var zapfDingbatsEncodingGlyphToCharcodeMap = map[string]byte{
	"space": 32,
	"a1":    33,
	"a2":    34,
	"a202":  35,
	"a3":    36,
	"a4":    37,
	"a5":    38,
	"a119":  39,
	"a118":  40,
	"a117":  41,
	"a11":   42,
	"a12":   43,
	"a13":   44,
	"a14":   45,
	"a15":   46,
	"a16":   47,
	"a105":  48,
	"a17":   49,
	"a18":   50,
	"a19":   51,
	"a20":   52,
	"a21":   53,
	"a22":   54,
	"a23":   55,
	"a24":   56,
	"a25":   57,
	"a26":   58,
	"a27":   59,
	"a28":   60,
	"a6":    61,
	"a7":    62,
	"a8":    63,
	"a9":    64,
	"a10":   65,
	"a29":   66,
	"a30":   67,
	"a31":   68,
	"a32":   69,
	"a33":   70,
	"a34":   71,
	"a35":   72,
	"a36":   73,
	"a37":   74,
	"a38":   75,
	"a39":   76,
	"a40":   77,
	"a41":   78,
	"a42":   79,
	"a43":   80,
	"a44":   81,
	"a45":   82,
	"a46":   83,
	"a47":   84,
	"a48":   85,
	"a49":   86,
	"a50":   87,
	"a51":   88,
	"a52":   89,
	"a53":   90,
	"a54":   91,
	"a55":   92,
	"a56":   93,
	"a57":   94,
	"a58":   95,
	"a59":   96,
	"a60":   97,
	"a61":   98,
	"a62":   99,
	"a63":   100,
	"a64":   101,
	"a65":   102,
	"a66":   103,
	"a67":   104,
	"a68":   105,
	"a69":   106,
	"a70":   107,
	"a71":   108,
	"a72":   109,
	"a73":   110,
	"a74":   111,
	"a203":  112,
	"a75":   113,
	"a204":  114,
	"a76":   115,
	"a77":   116,
	"a78":   117,
	"a79":   118,
	"a81":   119,
	"a82":   120,
	"a83":   121,
	"a84":   122,
	"a97":   123,
	"a98":   124,
	"a99":   125,
	"a100":  126,
	"a89":   128,
	"a90":   129,
	"a93":   130,
	"a94":   131,
	"a91":   132,
	"a92":   133,
	"a205":  134,
	"a85":   135,
	"a206":  136,
	"a86":   137,
	"a87":   138,
	"a88":   139,
	"a95":   140,
	"a96":   141,
	"a101":  161,
	"a102":  162,
	"a103":  163,
	"a104":  164,
	"a106":  165,
	"a107":  166,
	"a108":  167,
	"a112":  168,
	"a111":  169,
	"a110":  170,
	"a109":  171,
	"a120":  172,
	"a121":  173,
	"a122":  174,
	"a123":  175,
	"a124":  176,
	"a125":  177,
	"a126":  178,
	"a127":  179,
	"a128":  180,
	"a129":  181,
	"a130":  182,
	"a131":  183,
	"a132":  184,
	"a133":  185,
	"a134":  186,
	"a135":  187,
	"a136":  188,
	"a137":  189,
	"a138":  190,
	"a139":  191,
	"a140":  192,
	"a141":  193,
	"a142":  194,
	"a143":  195,
	"a144":  196,
	"a145":  197,
	"a146":  198,
	"a147":  199,
	"a148":  200,
	"a149":  201,
	"a150":  202,
	"a151":  203,
	"a152":  204,
	"a153":  205,
	"a154":  206,
	"a155":  207,
	"a156":  208,
	"a157":  209,
	"a158":  210,
	"a159":  211,
	"a160":  212,
	"a161":  213,
	"a163":  214,
	"a164":  215,
	"a196":  216,
	"a165":  217,
	"a192":  218,
	"a166":  219,
	"a167":  220,
	"a168":  221,
	"a169":  222,
	"a170":  223,
	"a171":  224,
	"a172":  225,
	"a173":  226,
	"a162":  227,
	"a174":  228,
	"a175":  229,
	"a176":  230,
	"a177":  231,
	"a178":  232,
	"a179":  233,
	"a193":  234,
	"a180":  235,
	"a199":  236,
	"a181":  237,
	"a200":  238,
	"a182":  239,
	"a201":  241,
	"a183":  242,
	"a184":  243,
	"a197":  244,
	"a185":  245,
	"a194":  246,
	"a198":  247,
	"a186":  248,
	"a195":  249,
	"a187":  250,
	"a188":  251,
	"a189":  252,
	"a190":  253,
	"a191":  254,
}
