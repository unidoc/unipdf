/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

var simpleFontDicts = []string{
	`<< /Type /Font
		/BaseFont /Helvetica
		/Subtype /Type1
		/Encoding /WinAnsiEncoding
		>>`,
	`<< /Type /Font
		/BaseFont /Helvetica-Oblique
		/Subtype /Type1
		/Encoding /WinAnsiEncoding
		>>`,
	`<< /Type /Font
		/Subtype /Type1
		/FirstChar 71
		/LastChar 79
		/Widths [ 778 722 278 500 667 556 833 722 778 ]
		/Encoding /WinAnsiEncoding
		/BaseFont /AOMFKK+Helvetica
		>>`,
	`<< /Type /Font
		/Subtype /Type1
		/FirstChar 71
		/LastChar 79
		/Widths [ 778 722 278 500 667 556 833 722 778 ]
		/Encoding /WinAnsiEncoding
		/BaseFont /PETER+Helvetica
		/FontDescriptor <<
			/Type /FontDescriptor
			/Ascent 718
			/CapHeight 718
			/Descent -207
			/Flags 32
			/FontBBox [ -166 -225 1000 931 ]
			/FontName /PETER+Helvetica
			/ItalicAngle 0
			/StemV 88
			/XHeight 523
			/StemH 88
			/CharSet (/G/O)
			%/FontFile3 19 0 R
			>>
		>>`,
}

var compositeFontDicts = []string{
	`<< /Type /Font
		/Subtype /Type0
		/Encoding /Identity-H
		/DescendantFonts [<<
			/Type /Font
			/Subtype /CIDFontType2
			/BaseFont /FLDOLC+PingFangSC-Regular
			/CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >>
			/W [ ]
			/DW 1000
			/FontDescriptor <<
				/Type /FontDescriptor
				/FontName /FLDOLC+PingFangSC-Regular
				/Flags 4
				/FontBBox [-123 -263 1177 1003]
				/ItalicAngle 0
				/Ascent 972
				/Descent -232
				/CapHeight 864
				/StemV 70
				/XHeight 648
				/StemH 64
				/AvgWidth 1000
				/MaxWidth 1300
				% /FontFile3 182 0 R
				>>
			>>]
		/BaseFont /FLDOLC+PingFangSC-Regular
		>>`,
}

func TestNewStandard14Font(t *testing.T) {
	type expect struct {
		subtype  string
		basefont string
		fonts.CharMetrics
	}
	tests := map[string]expect{
		"Courier": expect{
			subtype:     "Type1",
			basefont:    "Courier",
			CharMetrics: fonts.CharMetrics{Wx: 600, Wy: 0}},
	}

	for in, expect := range tests {
		font, err := model.NewStandard14Font(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if font.Subtype() != expect.subtype || font.BaseFont() != expect.basefont {
			t.Fatalf("%s: expected BaseFont=%s SubType=%s, but got BaseFont=%s SubType=%s",
				in, expect.basefont, expect.subtype, font.BaseFont(), font.Subtype())
		}

		metrics, ok := font.GetGlyphCharMetrics("space")
		if !ok {
			t.Fatalf("%s: failed to get glyph metric", in)
		}
		if metrics.Wx != expect.Wx || metrics.Wy != expect.Wy {
			t.Fatalf("%s: expected glyph metrics is Wx=%f Wy=%f, but got Wx=%f Wy=%f",
				in, expect.Wx, expect.Wy, metrics.Wx, metrics.Wy)
		}
	}
}

// TestSimpleFonts checks that we correctly recreate simple fonts that we parse.
func TestSimpleFonts(t *testing.T) {
	for _, d := range simpleFontDicts {
		objFontObj(t, d)
	}
}

// TestCompositeFonts checks that we correctly recreate composite fonts that we parse.
func TestCompositeFonts(t *testing.T) {
	for _, d := range compositeFontDicts {
		objFontObj(t, d)
	}
}

// TestCharcodeBytesToUnicode checks that CharcodeBytesToUnicode is working for the tests in
// ToUnicode cmap.
func TestCharcodeBytesToUnicode(t *testing.T) {
	for _, test := range charcodeBytesToUnicodeTest {
		test.check(t)
	}
}

var charcodeBytesToUnicodeTest = []fontFragmentTest{
	fontFragmentTest{"Helvetica built-in",
		"../../contrib/testdata/font/simple.txt", 1,
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
			53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74,
			75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
			97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114,
			115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 128, 130, 131, 132, 133,
			134, 135, 136, 137, 138, 139, 140, 142, 145, 146, 147, 148, 149, 150, 151, 152, 153,
			154, 155, 156, 158, 159, 161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172,
			174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 187, 188, 189, 190,
			191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207,
			208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224,
			225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237, 238, 239, 240, 241,
			242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255},
		" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`" +
			"abcdefghijklmnopqrstuvwxyz{|}~€‚ƒ„…†‡ˆ‰Š‹OEŽ‘’“”•–—˜™š›oežŸ¡¢£¤¥¦§¨©ª«¬®¯°±²³´µ¶·" +
			"¸¹º»¼½¾¿ÀÁÂÃÄÅAEÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞfzàáâãäåaeçèéêëìíîïðñòóôõö÷øùúûüýþÿ",
	},
	fontFragmentTest{"Symbol built-in",
		"../../contrib/testdata/font/simple.txt", 3,
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
			53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74,
			75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 97,
			98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115,
			116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 160, 161, 162, 163, 164, 165, 166,
			167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184,
			185, 186, 187, 188, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203, 204,
			205, 206, 207, 208, 209, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224, 225,
			229, 241, 242, 243, 245},
		" !∀#∃%&∋()∗+,−./0123456789:;<=>?≅ΑΒΧΔΕΦΓΗΙϑΚΛΜΝΟΠΘΡΣΤΥςΩΞΨΖ[∴]⊥_αβχδεφγηιϕκλμνοπθρστυϖω" +
			"ξψζ{|}∼€ϒ′≤⁄∞ƒ♣♦♥♠↔←↑→↓°±″≥×∝∂•÷≠≡≈…↵ℵℑℜ℘⊗⊕∅∩∪⊃⊇⊄⊂⊆∈∉∠∇∏√⋅¬∧∨⇔⇐⇑⇒⇓◊〈∑〉∫⌠⌡",
	},
	fontFragmentTest{"ZapfDingbats built-in",
		"../../contrib/testdata/font/simple.txt", 4,
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
			53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74,
			75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
			97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114,
			115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 161, 162, 163, 164, 165,
			166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183,
			184, 185, 186, 187, 188, 189, 190, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201,
			202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219,
			220, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237,
			238, 239, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254},
		" ✁✂✃✄☎✆✇✈✉☛☞✌✍✎✏✐✑✒✓✔✕✖✗✘✙✚✛✜✝✞✟✠✡✢✣✤✥✦✧★✩✪✫✬✭✮✯✰✱✲✳✴✵✶✷✸✹✺✻✼✽✾✿❀❁❂❃❄❅❆❇❈❉❊❋●❍■❏❐❑❒▲▼◆❖◗" +
			"❘❙❚❛❜❝❞❡❢❣❤❥❦❧♣♦♥♠①②③④⑤⑥⑦⑧⑨⑩❶❷❸❹❺❻❼❽❾❿➀➁➂➃➄➅➆➇➈➉➊➋➌➍➎➏➐➑➒➓➔→↔↕" +
			"➘➙➚➛➜➝➞➟➠➡➢➣➤➥➦➧➨➩➪➫➬➭➮➯➱➲➳➴➵➶➷➸➹➺➻➼➽➾",
	},
	fontFragmentTest{"MacRoman encoding",
		"../../contrib/testdata/font/axes.txt", 10,
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
			53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74,
			75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
			97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114,
			115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 128, 129, 130, 131, 132, 133,
			134, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150, 151,
			152, 153, 154, 155, 156, 157, 158, 159, 160, 161, 162, 163, 164, 165, 166, 167, 168, 169,
			170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 187,
			188, 189, 190, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 203, 204, 205, 206,
			207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224,
			225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237, 238, 239, 241, 242, 243,
			244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255},
		" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`" +
			"abcdefghijklmnopqrstuvwxyz{|}~ÄÅÇÉÑÖÜáàâäãåçéèêëíìîïñóòôöõúùûü†°¢£§•¶fz®©™´¨≠AEØ∞" +
			"±≤≥¥µ∂∑∏π∫ªºΩaeø¿¡¬√ƒ≈∆«»…ÀÃÕOEoe–—“”‘’÷◊ÿŸ⁄€‹›fifl‡·‚„‰ÂÊÁËÈÍÎÏÌÓÔÒÚÛÙıˆ˜¯˘˙˚¸˝˛ˇ",
	},
	fontFragmentTest{"Test beginbfchar and beginbfrange cmap entries",
		"../../contrib/testdata/font/Yemeni.txt", 470,
		[]byte{0x1, 0xa8, 0x1, 0xb3, 0x1, 0xc2, 0x1, 0xcc, 0x1, 0xe7, 0x1, 0xef, 0x1, 0xf3, 0x0,
			0x20, 0x1, 0xa2, 0x1, 0xfc, 0x2, 0x8, 0x1, 0xa6, 0x1, 0xe7, 0x0, 0x20, 0x2, 0xb, 0x0,
			0x20, 0x2, 0xf, 0x0, 0x20, 0x0, 0x20, 0x1, 0xdd, 0x0, 0x20, 0x0, 0xcd, 0x0, 0xce, 0x0,
			0xcf, 0x0, 0xd0, 0x0, 0xd1, 0x1, 0xa1, 0x0, 0x20, 0x1, 0xa9, 0x2, 0x1},
		"ﺔﺟﺮﺸﻓﻛﻟ ﺎﻨﻴﺒﻓ ﻷ ﻻ  ﻉ ٠١٢٣٤ﺍ ﺕﻭ",
	},
	fontFragmentTest{"TrueType font with ToUnicode cmap",
		"../../contrib/testdata/font/print_alerts.txt", 9,
		[]byte{43, 40, 41, 34, 37, 42, 38, 49, 36, 38, 48, 34, 35, 36, 37, 35, 36, 58},
		"Alerts on printing",
	},
	fontFragmentTest{"Type0 font with ToUnicode cmap",
		"../../contrib/testdata/font/CollazoBio.txt", 7,
		[]byte{255, 50, 255, 65, 255, 78, 255, 68, 255, 79, 255, 77, 0, 32, 0, 32, 255, 77, 255, 65,
			255, 84, 255, 82, 255, 73, 255, 67, 255, 69, 255, 83, 0, 46},
		"Ｒａｎｄｏｍ  ｍａｔｒｉｃｅｓ.",
	},
	fontFragmentTest{"Type1 font with FontFile entry",
		"../../contrib/testdata/font/lm.txt", 7,
		[]byte{102, 65, 106, 66, 103},
		"{A|B}",
	},
	fontFragmentTest{"Type1 font with /Encoding with /Differences",
		"../../contrib/testdata/font/noise-invariant.txt", 102,
		[]byte{96, 247, 39, 32, 147, 231, 148, 32, 232, 32, 193, 111, 180, 32, 105, 116,
			169, 115, 32, 204, 195, 196, 197, 198, 199, 168, 202, 206, 226, 234, 172, 244, 173, 151,
			177, 151, 178, 179, 183, 185, 188, 205, 184, 189},
		"‘ł’ “Ł” Ø `o´ it's ˝ˆ˜¯˘˙¨˚ˇªº‹ı›—–—†‡•„…˛¸‰",
	},
	fontFragmentTest{"A potpourri of glyph naming conventions",
		"../../contrib/testdata/font/Ingmar.txt", 144,
		[]byte{18, 20, 10, 11, 13, 14, 15, 16, 21, 22, 23, 25, 26, 27, 28, 29, 30,
			31, 33, 12, 17, 19, 24},
		"ʼ8ČŽĆřćĐĭűőftffiflfffičž!fbfkffl\u00a0",
	},
	fontFragmentTest{"Zapf Dingbats",
		"../../contrib/testdata/font/estimation.txt", 122,
		[]byte{2, 3, 4, 5, 8, 9, 10, 11, 12, 13, 14},
		"✏✮✁☛❄❍❥❇◆✟✙",
	},
	fontFragmentTest{"Found these by trial and error",
		"../../contrib/testdata/font/helminths.txt", 19,
		[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
			32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
			53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74,
			75, 76, 77},
		" *ﺏﻁﻝﺍﺔﻴﻠﻜ،ﺕﺭﺘﻌﻤﺎﺠﻲﻨﻘﺩﻬ/ﻙﻭﻕﺃﻡﻋﻓﺴ٢٠٣ﻯﻥﺒﺸﺌﺱﻷ,ﺯﺤﺄﻀـﺓﺫ.)٤(٩ل٥٧٨ﻸﻰ%١ﺇ٦ﺡﻫﻱﻅﻐﺼﻑﺨﺀﻊLM",
	},
}

type fontFragmentTest struct {
	description string
	filename    string
	objNum      int
	data        []byte
	expected    string
}

func (f *fontFragmentTest) String() string {
	return fmt.Sprintf("TEST{%q file=%q obj=%d}", f.description, f.filename, f.objNum)
}

// check loads the font in PDF fragment `filename`, object number `objNum`, runs
// CharcodeBytesToUnicode on `data` and checks that output equals `expected`.
func (f *fontFragmentTest) check(t *testing.T) {
	common.Log.Debug("fontFragmentTest: %s", f)
	numObj, err := parsePdfFragment(f.filename)
	if err != nil {
		t.Errorf("Failed to parse. %s err=%v", f, err)
		return
	}
	fontObj, ok := numObj[f.objNum]
	if !ok {
		t.Errorf("fontFragmentTest: %s. Unknown object. %d", f, f.objNum)
		return
	}
	font, err := model.NewPdfFontFromPdfObject(fontObj)
	if err != nil {
		t.Errorf("fontFragmentTest: %s. Failed to create font. err=%v", f, err)
		return
	}

	actualText, numChars, numMisses := font.CharcodeBytesToUnicode(f.data)
	if numMisses != 0 {
		t.Errorf("Some codes not decoded. numMisses=%d", numMisses)
		return
	}
	if actualText != f.expected {
		t.Errorf("Incorrect decoding. %s\nexpected=%q\n  actual=%q",
			f, f.expected, actualText)
	}
	if numChars != len([]rune(actualText)) {
		t.Errorf("Incorrect numChars. %s numChars=%d expected=%d\n%+v\n%c",
			f, numChars, len([]rune(actualText)), []rune(actualText), []rune(actualText))
	}
}

// objFontObj parses `fontDict` to a make a Font, creates a PDF object from the Font and checks that
// the new PDF object is the same as the input object
func objFontObj(t *testing.T, fontDict string) error {

	parser := core.NewParserFromString(fontDict)
	obj, err := parser.ParseDict()
	if err != nil {
		t.Errorf("objFontObj: Failed to parse dict obj. fontDict=%q err=%v", fontDict, err)
		return err
	}
	font, err := model.NewPdfFontFromPdfObject(obj)
	if err != nil {
		t.Errorf("Failed to parse font object. obj=%s err=%v", obj, err)
		return err
	}

	// Resolve all the indirect references in the font objects so we can compare their contents.
	obj1 := core.FlattenObject(obj)
	obj2 := core.FlattenObject(font.ToPdfObject())

	// Check that the reconstituted font is the same as the original.
	if !core.EqualObjects(obj1, obj2) {
		t.Errorf("Different objects.\nobj1=%s\nobj2=%s\nfont=%s", obj1, obj2, font)
		return errors.New("different objects")
	}

	return nil
}

// parsePdfFragment parses a file containing fragments of a PDF `filename` (see
// charcodeBytesToUnicodeTest) and returns a map of {object number: object} with indirect objects
// replaced by their values if they are in `filename`.
func parsePdfFragment(filename string) (map[int]core.PdfObject, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parsePdfObjects(string(data))
}

// parsePdfObjects parses a fragment of a PDF `text` and returns a map of {object number: object}
// with indirect objects replaced by their values if they are in `text`.
func parsePdfObjects(text string) (map[int]core.PdfObject, error) {
	numObj := map[int]core.PdfObject{}
	parser := core.NewParserFromString(text)
	common.Log.Debug("parsePdfObjects")

	// Build the numObj {object number: object} map
	nums := []int{}
	for {
		obj, err := parser.ParseIndirectObject()
		common.Log.Debug("parsePdfObjects:  %T %v", obj, err)
		if err != nil {
			if err == io.EOF {
				break
			}
			common.Log.Debug("parsePdfObjects:  err=%v", err)
			return numObj, err
		}
		common.Log.Debug("parsePdfObjects: %d %T", len(numObj), obj)
		switch t := obj.(type) {
		case *core.PdfIndirectObject:
			numObj[int(t.ObjectNumber)] = obj
			nums = append(nums, int(t.ObjectNumber))
		case *core.PdfObjectStream:
			numObj[int(t.ObjectNumber)] = obj
			nums = append(nums, int(t.ObjectNumber))
		}
	}

	common.Log.Debug("parsePdfObjects: Parsed %d objects %+v", len(numObj), nums)

	// Replace the indirect objects in all dicts and arrays with their values, if they are in numObj.
	for n, obj := range numObj {
		common.Log.Debug("-- 0 %d obj %T", n, obj)
		iobj, ok := obj.(*core.PdfIndirectObject)
		if !ok {
			continue
		}
		common.Log.Debug("   -- %T", iobj.PdfObject)
		iobj.PdfObject, ok = replaceReferences(numObj, iobj.PdfObject)
		if !ok {
			common.Log.Debug("ERROR: unresolved reference")
		}
	}
	return numObj, nil
}

// replaceReferences replaces the object references in all dicts and arrays with their values, if
// they are in numObj. The boolean return is true if all object references were successfuly
// replaced.
func replaceReferences(numObj map[int]core.PdfObject, obj core.PdfObject) (core.PdfObject, bool) {
	var ok bool
	switch t := obj.(type) {
	case *core.PdfObjectReference:
		o, ok := numObj[int(t.ObjectNumber)]
		common.Log.Debug("    %d 0 R  %t ", t.ObjectNumber, ok)
		return o, ok
	case *core.PdfObjectDictionary:
		for _, k := range t.Keys() {
			o := t.Get(k)
			o, ok = replaceReferences(numObj, o)
			if !ok {
				return o, ok
			}
			t.Set(k, o)
		}
	case *core.PdfObjectArray:
		for i, o := range t.Elements() {
			o, ok = replaceReferences(numObj, o)
			if !ok {
				return o, ok
			}
			t.Set(i, o)
		}
	}
	return obj, true
}
