/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"

	"github.com/unidoc/unipdf/v3/internal/testutils"
	"github.com/unidoc/unipdf/v3/internal/textencoding"
	"github.com/unidoc/unipdf/v3/model/internal/fonts"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

var simpleFontDicts = []string{
	`<< /Type /Font
		/BaseFont /Courier-BoldOblique
		/Subtype /Type1
		>>`,
	`<< /Type /Font
		/BaseFont /Helvetica
		/Subtype /Type1
		/Encoding /WinAnsiEncoding
		>>`,
	`<< /Type /Font
		/BaseFont /Courier
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
	tests := map[fonts.StdFontName]expect{
		"Courier": {
			subtype:     "Type1",
			basefont:    "Courier",
			CharMetrics: fonts.CharMetrics{Wx: 600},
		},
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

		// Test space character metrics.
		metrics, ok := font.GetRuneMetrics(' ')
		if !ok {
			t.Fatalf("%s: failed to get glyph metric", in)
		}
		if metrics.Wx != expect.Wx || metrics.Wy != expect.Wy {
			t.Fatalf("%s: expected glyph metrics is Wx=%f Wy=%f, but got Wx=%f Wy=%f",
				in, expect.Wx, expect.Wy, metrics.Wx, metrics.Wy)
		}

		// Test no-break space character metrics.
		// Values should be the same as the metrics for regular space.
		metrics, ok = font.GetRuneMetrics(0xA0)
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
		t.Run("", func(t *testing.T) {
			objFontObj(t, d)
		})
	}
}

// TestStandardFontDict tests PDF object output of standard font.
// Importantly, this test makes sure that the output dictionary does not have an `Encoding`
// key and uses the encoding of the standard font (ZapfEncoding in this case).
func TestStandardFontOutputDict(t *testing.T) {
	zapfdb, err := model.NewStandard14Font(fonts.ZapfDingbatsName)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	dict, ok := core.GetDict(zapfdb.ToPdfObject())
	if !ok {
		t.Fatalf("not a dict")
	}

	if len(dict.Keys()) != 3 {
		t.Fatalf("Incorrect number of keys (%d): %v", len(dict.Keys()), dict.Keys())
	}

	ntype, ok := core.GetName(dict.Get("Type"))
	if !ok {
		t.Fatalf("invalid Type")
	}
	if ntype.String() != "Font" {
		t.Fatalf("Type != Font (%s)", ntype.String())
	}

	basef, ok := core.GetName(dict.Get("BaseFont"))
	if !ok {
		t.Fatalf("Invalid BaseFont")
	}
	if basef.String() != "ZapfDingbats" {
		t.Fatalf("BaseFont != ZapfDingbats (%s)", basef.String())
	}

	subtype, ok := core.GetName(dict.Get("Subtype"))
	if !ok {
		t.Fatalf("Invalid Subtype")
	}
	if subtype.String() != "Type1" {
		t.Fatalf("Subtype != Type1 (%s)", subtype.String())
	}
}

// Test loading a standard font from object and check the encoding and glyph metrics.
func TestLoadStandardFontEncodings(t *testing.T) {
	raw := `
	1 0 obj
	<< /Type /Font
		/BaseFont /Courier
		/Subtype /Type1
	>>
	endobj
	`

	r := model.NewReaderForText(raw)

	err := r.ParseIndObjSeries()
	if err != nil {
		t.Fatalf("Failed loading indirect object series: %v", err)
	}

	// Load the field from object number 1.
	obj, err := r.GetIndirectObjectByNumber(1)
	if err != nil {
		t.Fatalf("Failed to parse indirect obj (%s)", err)
	}

	font, err := model.NewPdfFontFromPdfObject(obj)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	str := "Aabcdefg0123456790*"
	for _, r := range str {
		_, has := font.GetRuneMetrics(r)
		if !has {
			t.Fatalf("Loaded simple font not having glyph char metrics for %v", r)
		}
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
		t.Run(test.description, func(t *testing.T) {
			test.check(t)
		})
	}
}

// TestFontDescriptor checks that the builtin standard 14 font descriptors are working.
func TestFontDescriptor(t *testing.T) {
	type params struct {
		FontName   string
		FontFamily string
		Flags      uint
		FontBBox   [4]float64
		CapHeight  float64
		XHeight    float64
	}

	tests := map[fonts.StdFontName]params{
		"Courier": params{
			FontName:   "Courier",
			FontFamily: "Courier",
			Flags:      0x0021,
			FontBBox:   [4]float64{-23, -250, 715, 805},
			CapHeight:  562,
			XHeight:    426,
		},
		"ZapfDingbats": params{
			FontName:   "ZapfDingbats",
			FontFamily: "ZapfDingbats",
			Flags:      0x0004,
			FontBBox:   [4]float64{-1, -143, 981, 820},
			CapHeight:  0,
			XHeight:    0,
		},
	}

	for fontName, expect := range tests {
		t.Run(string(fontName), func(t *testing.T) {
			font := model.NewStandard14FontMustCompile(fontName)

			descriptor := font.FontDescriptor()
			if descriptor == nil {
				t.Fatalf("%#q: No descriptor.", fontName)
			}
			actualFontName, ok := core.GetNameVal(descriptor.FontName)
			if !ok {
				t.Fatalf("%#q: No FontName. descriptor=%+v", fontName, descriptor)
			}
			fontFamily, ok := core.GetNameVal(descriptor.FontFamily)
			if !ok {
				t.Fatalf("%#q: No FontFamily. descriptor=%+v", fontName, descriptor)
			}
			flags, ok := core.GetIntVal(descriptor.Flags)
			if !ok {
				t.Fatalf("%#q: No Flags. descriptor=%+v", fontName, descriptor)
			}
			arr, ok := core.GetArray(descriptor.FontBBox)
			if !ok {
				t.Fatalf("%#q: No FontBBox. descriptor=%+v", fontName, descriptor)
			}
			fontBBox := [4]float64{}
			for i := 0; i < 4; i++ {
				x, ok := core.GetFloatVal(arr.Get(i))
				if !ok {
					t.Fatalf("%#q: Bad FontBBox. descriptor=%+v", fontName, descriptor)
				}
				fontBBox[i] = x
			}

			capHeight, ok := core.GetFloatVal(descriptor.CapHeight)
			if !ok {
				t.Fatalf("%#q: No CapHeight. descriptor=%+v", fontName, descriptor)
			}
			xHeight, ok := core.GetFloatVal(descriptor.XHeight)
			if !ok {
				t.Fatalf("%#q: No XHeight. descriptor=%+v", fontName, descriptor)
			}

			actual := params{
				FontName:   actualFontName,
				FontFamily: fontFamily,
				Flags:      uint(flags),
				FontBBox:   fontBBox,
				CapHeight:  capHeight,
				XHeight:    xHeight,
			}

			if actual.FontName != expect.FontName ||
				actual.FontFamily != expect.FontFamily ||
				actual.Flags != expect.Flags ||
				actual.FontBBox != expect.FontBBox ||
				actual.CapHeight != expect.CapHeight {
				t.Fatalf("%s:\n\texpect=%+v\n\tactual=%+v", fontName, expect, actual)
			}
		})
	}
}

var charcodeBytesToUnicodeTest = []fontFragmentTest{
	{"Helvetica WinAnsiEncoding",
		"./testdata/font/simple.txt", 1,
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
			"¸¹º»¼½¾¿ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞfzàáâãäåæçèéêëìíîïðñòóôõö÷øùúûüýþÿ",
	},
	{"Helvetica built-in",
		"./testdata/font/simple.txt", 5,
		[]byte{
			32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58,
			59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85,
			86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
			110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126,
			161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 177, 178, 179, 180, 182, 183,
			184, 185, 186, 187, 188, 189, 191, 193, 194, 195, 196, 197, 198, 199, 225, 227, 232, 241, 245, 248, 249,
			250, 251},
		` !"#$%&’()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_‘abcdefghijklmnopqrstuvwxyz{|}~` +
			`¡¢£⁄¥ƒ§¤'“«‹›fifl–†‡·¶•‚„”»…‰¿` + "`" + `´ˆ˜¯˘˙ÆªŁæıłøoefz`,
	},
	{"Symbol built-in",
		"./testdata/font/simple.txt", 3,
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
			53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74,
			75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 97,
			98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115,
			116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 160, 161, 162, 163, 164, 165, 166,
			167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184,
			185, 186, 187, 188, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203, 204,
			205, 206, 207, 208, 209, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224, 225,
			229, 241, 242, 243, 245},
		" !∀#∃%&∋()∗+,−./0123456789:;<=>?≅ΑΒΧ∆ΕΦΓΗΙϑΚΛΜΝΟΠΘΡΣΤΥςΩΞΨΖ[∴]⊥_αβχδεφγηιϕκλµνοπθρστυϖω" +
			"ξψζ{|}∼€ϒ′≤⁄∞ƒ♣♦♥♠↔←↑→↓°±″≥×∝∂•÷≠≡≈…↵ℵℑℜ℘⊗⊕∅∩∪⊃⊇⊄⊂⊆∈∉∠∇∏√⋅¬∧∨⇔⇐⇑⇒⇓◊〈∑〉∫⌠⌡",
	},
	{"ZapfDingbats built-in",
		"./testdata/font/simple.txt", 4,
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
	{"MacRoman encoding",
		"./testdata/font/axes.txt", 10,
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
			"abcdefghijklmnopqrstuvwxyz{|}~ÄÅÇÉÑÖÜáàâäãåçéèêëíìîïñóòôöõúùûü†°¢£§•¶fz®©™´¨≠ÆØ∞" +
			"±≤≥¥µ∂∑∏π∫ªºΩæø¿¡¬√ƒ≈∆«»…ÀÃÕOEoe–—“”‘’÷◊ÿŸ⁄€‹›fifl‡·‚„‰ÂÊÁËÈÍÎÏÌÓÔÒÚÛÙıˆ˜¯˘˙˚¸˝˛ˇ",
	},
	{"Test beginbfchar and beginbfrange cmap entries",
		"./testdata/font/Yemeni.txt", 470,
		[]byte{0x1, 0xa8, 0x1, 0xb3, 0x1, 0xc2, 0x1, 0xcc, 0x1, 0xe7, 0x1, 0xef, 0x1, 0xf3, 0x0,
			0x20, 0x1, 0xa2, 0x1, 0xfc, 0x2, 0x8, 0x1, 0xa6, 0x1, 0xe7, 0x0, 0x20, 0x2, 0xb, 0x0,
			0x20, 0x2, 0xf, 0x0, 0x20, 0x0, 0x20, 0x1, 0xdd, 0x0, 0x20, 0x0, 0xcd, 0x0, 0xce, 0x0,
			0xcf, 0x0, 0xd0, 0x0, 0xd1, 0x1, 0xa1, 0x0, 0x20, 0x1, 0xa9, 0x2, 0x1},
		"ﺔﺟﺮﺸﻓﻛﻟ ﺎﻨﻴﺒﻓ ﻷ ﻻ  ﻉ ٠١٢٣٤ﺍ ﺕﻭ",
	},
	{"TrueType font with ToUnicode cmap",
		"./testdata/font/print_alerts.txt", 9,
		[]byte{43, 40, 41, 34, 37, 42, 38, 49, 36, 38, 48, 34, 35, 36, 37, 35, 36, 58},
		"Alerts on printing",
	},
	{"Type0 font with ToUnicode cmap",
		"./testdata/font/CollazoBio.txt", 7,
		[]byte{255, 50, 255, 65, 255, 78, 255, 68, 255, 79, 255, 77, 0, 32, 0, 32, 255, 77, 255, 65,
			255, 84, 255, 82, 255, 73, 255, 67, 255, 69, 255, 83, 0, 46},
		"Ｒａｎｄｏｍ  ｍａｔｒｉｃｅｓ.",
	},
	{"Type1 font with FontFile entry",
		"./testdata/font/lm.txt", 7,
		[]byte{102, 65, 106, 66, 103},
		"{A|B}",
	},
	{"Type1 font with /Encoding with /Differences",
		"./testdata/font/noise-invariant.txt", 102,
		[]byte{96, 248, 39, 32, 147, 232, 148, 32, 233, 32, 193, 111, 180, 32, 105, 116,
			169, 115, 32, 204, 195, 196, 197, 198, 199, 168, 202, 206, 227, 235, 172, 245, 173, 151,
			177, 151, 178, 179, 183, 185, 188, 205, 184, 189},
		"‘ł’ “Ł” Ø `o´ it's ˝ˆ˜¯˘˙¨˚ˇªº‹ı›—–—†‡•„…˛¸‰",
	},
	{"base glyphs′",
		"./testdata/font/cover.txt", 11,
		[]byte{44, 45, 46, 48, 49, 50, 51, 53, 54, 55, 56, 58, 59,
			65, 66, 67, 68, 69, 70, 71, 72,
			84, 85,
			97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 108, 109, 110, 111,
			114, 115, 116, 117},
		",-.01235678:;ABCDEFGHTUabcdefghijlmnorstu",
	},
	{"tex glyphs 48->′",
		"./testdata/font/noise-contrast.txt", 36,
		[]byte{33, 48, 65, 104, 149, 253},
		"!′Ah•ý",
	},
	{"tex2 glyphs ",
		"./testdata/font/Weil.txt", 30,
		[]byte{55, 0, 1, 2, 20, 24, 33, 50, 102, 103, 104, 105},
		"↦−·×≤∼→∈{}⟨⟩",
	},
	{"additional glyphs",
		"./testdata/font/noise-contrast.txt", 34,
		[]byte{32, 40, 48, 64, 80, 88, 65, 104, 149, 253},
		"({∑∑h•ý",
	},
	{".notdef glyphs",
		"./testdata/font/lec10.txt", 6,
		[]byte{59, 66},
		string([]rune{textencoding.MissingCodeRune, textencoding.MissingCodeRune}),
	},
	{"Numbered glyphs pattern 1",
		"./testdata/font/v14.txt", 14,
		[]byte{24, 25, 26, 27, 29},
		" ffifflfffi",
	},
	{"Glyph aliases",
		"./testdata/font/townes.txt", 10,
		[]byte{2, 3, 4, 5, 6, 7, 1, 8, 9, 5, 1, 10, 9, 5, 48},
		"Townes van Zan…",
	},
	{"Glyph `.` extensions. e.g. `integral.disp`",
		"./testdata/font/preview.txt", 156,
		[]byte{83, 0, 4, 67, 62, 64, 100, 65},
		"∫=≈≥∈<d>",
	},
	{"A potpourri of glyph naming conventions",
		"./testdata/font/Ingmar.txt", 144,
		[]byte{18, 20, 10, 11, 13, 14, 15, 16, 21, 22, 23, 25, 26, 27, 28, 29, 30,
			31, 33, 12, 17, 19, 24},
		"ʼ8ČŽĆřćĐĭűőftffiflfffičž!fbfkffl\u00a0",
	},
	{"Zapf Dingbats",
		"./testdata/font/estimation.txt", 122,
		[]byte{2, 3, 4, 5, 8, 9, 10, 11, 12, 13, 14},
		"✏✮✁☛❄❍❥❇◆✟✙",
	},
	{"Found these by trial and error",
		"./testdata/font/helminths.txt", 19,
		[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
			32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
			53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74,
			75, 76, 77},
		" *ﺏﻁﻝﺍﺔﻴﻠﻜ،ﺕﺭﺘﻌﻤﺎﺠﻲﻨﻘﺩﻬ/ﻙﻭﻕﺃﻡﻋﻓﺴ٢٠٣ﻯﻥﺒﺸﺌﺱﻷ,ﺯﺤﺄﻀـﺓﺫ.)٤(٩ل٥٧٨ﻸﻰ%١ﺇ٦ﺡﻫﻱﻅﻐﺼﻑﺨﺀﻊLM",
	},
	{"Tesseract",
		"./testdata/font/tesseract.txt", 3,
		[]byte{0, 65, 0, 97,
			1, 2, 1, 65, 1, 97,
			12, 2, 12, 65, 12, 97,
			20, 65, 20, 97, 20, 255,
			42, 2, 42, 65, 42, 97,
			65, 66, 67, 255},
		"AaĂŁšంుౡᑁᑡᓿ⨂⩁⩡䅂䏿",
	},
}

type fontFragmentTest struct {
	description string
	filename    string
	objNum      int64
	data        []byte
	expected    string
}

func (f *fontFragmentTest) String() string {
	return fmt.Sprintf("TEST{%q file=%q obj=%d}", f.description, f.filename, f.objNum)
}

// check loads the font in PDF fragment `filename`, object number `objNum`, runs
// CharcodeBytesToUnicode on `data` and checks that output equals `expected`.
func (f *fontFragmentTest) check(t *testing.T) {
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
		t.Errorf("Some codes not decoded %s. font=%s numMisses=%d", f, font, numMisses)
		return
	}
	if actualText != f.expected {
		t.Errorf("Incorrect decoding. %s encoding=%s\nexpected=%q\n  actual=%q",
			f, font.Encoder(), f.expected, actualText)
		act, exp := []rune(actualText), []rune(f.expected)
		if len(act) != len(exp) {
			t.Errorf("\texpected=%d actual=%d", len(exp), len(act))
		} else {
			for i, a := range act {
				e := exp[i]
				if a != e {
					t.Errorf("\ti=%d expected=0x%04x=%c actual=0x%04x=%c", i, e, e, a, a)
				}
			}
		}
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
func parsePdfFragment(filename string) (map[int64]core.PdfObject, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return testutils.ParseIndirectObjects(string(data))
}

// TestLoadedSimpleFontEncoding tests loading a simple font with a Differences encoding.
// It checks if the loaded font encoding has the expected characteristics.
func TestLoadedSimpleFontEncoding(t *testing.T) {
	rawpdf := `
59 0 obj
<</BaseFont /Helvetica/Encoding 60 0 R/Name /Helv/Subtype /Type1/Type /Font>>
endobj
60 0 obj
<</Differences [24 /breve /caron /circumflex /dotaccent /hungarumlaut /ogonek /ring /tilde 39 /quotesingle 96 /grave 128 /bullet /dagger /daggerdbl /ellipsis /emdash /endash /florin /fraction /guilsinglleft /guilsinglright /minus /perthousand /quotedblbase /quotedblleft /quotedblright /quoteleft /quoteright /quotesinglbase /trademark /fi /fl /Lslash /OE /Scaron /Ydieresis /Zcaron /dotlessi /lslash /oe /scaron /zcaron 160 /Euro 164 /currency 166 /brokenbar 168 /dieresis /copyright /ordfeminine 172 /logicalnot /.notdef /registered /macron /degree /plusminus /twosuperior /threesuperior /acute /mu 183 /periodcentered /cedilla /onesuperior /ordmasculine 188 /onequarter /onehalf /threequarters 192 /Agrave /Aacute /Acircumflex /Atilde /Adieresis /Aring /AE /Ccedilla /Egrave /Eacute /Ecircumflex /Edieresis /Igrave /Iacute /Icircumflex /Idieresis /Eth /Ntilde /Ograve /Oacute /Ocircumflex /Otilde /Odieresis /multiply /Oslash /Ugrave /Uacute /Ucircumflex /Udieresis /Yacute /Thorn /germandbls /agrave /aacute /acircumflex /atilde /adieresis /aring /ae /ccedilla /egrave /eacute /ecircumflex /edieresis /igrave /iacute /icircumflex /idieresis /eth /ntilde /ograve /oacute /ocircumflex /otilde /odieresis /divide /oslash /ugrave /uacute /ucircumflex /udieresis /yacute /thorn /ydieresis]/Type /Encoding>>
endobj
`

	objects, err := testutils.ParseIndirectObjects(rawpdf)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	font, err := model.NewPdfFontFromPdfObject(objects[59])
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// The expected encoding is StandardEncoding with the applied differences.
	baseEncoding := newStandandTextEncoder(t)

	differencesMap := map[textencoding.CharCode]rune{
		24:  '˘',
		25:  'ˇ',
		26:  'ˆ',
		27:  '˙',
		28:  '˝',
		29:  '˛',
		30:  '˚',
		31:  '˜',
		39:  '\'',
		96:  '`',
		128: '•',
		129: '†',
		130: '‡',
		131: '…',
		132: '—',
		133: '–',
		134: 'ƒ',
		135: '⁄',
		136: '‹',
		137: '›',
		138: '−',
		139: '‰',
		140: '„',
		141: '“',
		142: '”',
		143: '‘',
		144: '’',
		145: '‚',
		146: '™',
		147: 'ﬁ',
		148: 'ﬂ',
		149: 'Ł',
		150: 'Œ',
		151: 'Š',
		152: 'Ÿ',
		153: 'Ž',
		154: 'ı',
		155: 'ł',
		156: 'œ',
		157: 'š',
		158: 'ž',
		160: '€',
		164: '¤',
		166: '¦',
		168: '¨',
		169: '©',
		170: 'ª',
		172: '¬',
		173: '�',
		174: '®',
		175: '¯',
		176: '°',
		177: '±',
		178: '²',
		179: '³',
		180: '´',
		181: 'µ',
		183: '·',
		184: '¸',
		185: '¹',
		186: 'º',
		188: '¼',
		189: '½',
		190: '¾',
		192: 'À',
		193: 'Á',
		194: 'Â',
		195: 'Ã',
		196: 'Ä',
		197: 'Å',
		198: 'Æ',
		199: 'Ç',
		200: 'È',
		201: 'É',
		202: 'Ê',
		203: 'Ë',
		204: 'Ì',
		205: 'Í',
		206: 'Î',
		207: 'Ï',
		208: 'Ð',
		209: 'Ñ',
		210: 'Ò',
		211: 'Ó',
		212: 'Ô',
		213: 'Õ',
		214: 'Ö',
		215: '×',
		216: 'Ø',
		217: 'Ù',
		218: 'Ú',
		219: 'Û',
		220: 'Ü',
		221: 'Ý',
		222: 'Þ',
		223: 'ß',
		224: 'à',
		225: 'á',
		226: 'â',
		227: 'ã',
		228: 'ä',
		229: 'å',
		230: 'æ',
		231: 'ç',
		232: 'è',
		233: 'é',
		234: 'ê',
		235: 'ë',
		236: 'ì',
		237: 'í',
		238: 'î',
		239: 'ï',
		240: 'ð',
		241: 'ñ',
		242: 'ò',
		243: 'ó',
		244: 'ô',
		245: 'õ',
		246: 'ö',
		247: '÷',
		248: 'ø',
		249: 'ù',
		250: 'ú',
		251: 'û',
		252: 'ü',
		253: 'ý',
		254: 'þ',
		255: 'ÿ',
	}

	enc := font.Encoder()
	for code := textencoding.CharCode(32); code < 255; code++ {
		fontrune, has := enc.CharcodeToRune(code)
		if !has {
			baserune, bad := baseEncoding.CharcodeToRune(code)
			if bad {
				t.Fatalf("font not having glyph for char code %d - whereas base encoding had %q", code, baserune)
			}
		}

		// Check if in differencesmap first.
		rune, has := differencesMap[code]
		if has {
			if rune != fontrune {
				t.Fatalf("Mismatch for char code %d, font has: %q and expected is: %q (differences)", code, fontrune, rune)
			}

			continue
		}

		// If not in differences, should be according to StandardEncoding (base).
		rune, has = baseEncoding.CharcodeToRune(code)
		if has && rune != fontrune {
			t.Fatalf("Mismatch for char code %d (%X), font has: %q and expected is: %q (StandardEncoding)", code, code, fontrune, rune)
		}
	}
}

func TestLoadSimpleFontWithDifferences(t *testing.T) {
	testcases := []struct {
		Path             string
		FontObjNumber    int64
		BaseEncodingName string
	}{
		{"./testdata/font/diff1.obj", 53, "WinAnsiEncoding"},
	}

	for _, tcase := range testcases {
		data, err := ioutil.ReadFile(tcase.Path)
		require.NoError(t, err)

		objects, err := testutils.ParseIndirectObjects(string(data))
		require.NoError(t, err)

		font, err := model.NewPdfFontFromPdfObject(objects[tcase.FontObjNumber])
		require.NoError(t, err)
		require.NotNil(t, font)

		encoder := font.Encoder()
		require.NotNil(t, encoder)

		stdEncoder, ok := encoder.(textencoding.SimpleEncoder)
		require.True(t, ok)

		require.Equal(t, tcase.BaseEncodingName, stdEncoder.BaseName())
	}
}

// newStandandTextEncoder returns a simpleEncoder that implements StandardEncoding.
// The non-symbolic standard 14 fonts have StandardEncoding.
func newStandandTextEncoder(t *testing.T) textencoding.SimpleEncoder {
	enc, err := textencoding.NewSimpleTextEncoder("StandardEncoding", nil)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	return enc
}
