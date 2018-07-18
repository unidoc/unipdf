package model_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func init() {
	// common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
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
	fontFragmentTest{
		"TrueType font with ToUnicode cmap",
		"testdata/print_alerts.txt", 9,
		[]byte{43, 40, 41, 34, 37, 42, 38, 49, 36, 38, 48, 34, 35, 36, 37, 35, 36, 58},
		"Alerts on printing",
	},
	fontFragmentTest{
		"Type1 font with FontFile entry",
		"testdata/lm.txt", 7,
		[]byte{102, 65, 106, 66, 103},
		"{A|B}",
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
	return fmt.Sprintf("TEST{%q file=%q obj=%d|", f.description, f.filename, f.objNum)
}

// check loads the font in PDF fragment `filename`, object number `objNum`, runs
// CharcodeBytesToUnicode on `data` and checks that output equals `expected`.
func (f *fontFragmentTest) check(t *testing.T) {
	numObj, err := parsePdfFragment(f.filename)
	if err != nil {
		t.Errorf("Failed to parse. %s err=%v", f, err)
		return
	}
	fontObj := numObj[f.objNum]
	font, err := model.NewPdfFontFromPdfObject(fontObj)
	if err != nil {
		t.Errorf("Failed to create font. %s err=%v", f, err)
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
	if numChars != len(actualText) {
		t.Errorf("Incorrect numChars. %s numChars=%d expected=%d",
			f, numChars, len(actualText))
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

func parsePdfFragment(filename string) (map[int]core.PdfObject, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parsePdfObjects(string(data))
}

// parsePdfObjects parses a fragment of a PDF `text` (e.g. ttToUnicode above) and returns a map of
// {object number: object} with indirect objects replaced by their values if they are in `text`.
func parsePdfObjects(text string) (map[int]core.PdfObject, error) {
	numObj := map[int]core.PdfObject{}
	parser := core.NewParserFromString(text)

	// Build the numObj {object number: object} map
	for {
		obj, err := parser.ParseIndirectObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return numObj, err
		}
		switch t := obj.(type) {
		case *core.PdfIndirectObject:
			numObj[int(t.ObjectNumber)] = obj
		case *core.PdfObjectStream:
			numObj[int(t.ObjectNumber)] = obj
		}
	}

	// Replace the indirect objects in all dicts with their values, if they are in numObj.
	for i, obj := range numObj {
		iobj, ok := obj.(*core.PdfIndirectObject)
		if !ok {
			continue
		}
		dict, ok := iobj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			continue
		}
		for _, k := range dict.Keys() {
			if ref, ok := dict.Get(k).(*core.PdfObjectReference); ok {
				if o, ok := numObj[int(ref.ObjectNumber)]; ok {
					dict.Set(k, o)
				}
			}
		}
		fmt.Printf("%d 0 obj %s\n", i, showDict(dict))
	}

	return numObj, nil
}

func showDict(dict *core.PdfObjectDictionary) string {
	parts := []string{}
	for _, k := range dict.Keys() {
		parts = append(parts, fmt.Sprintf("%s: %T", k, dict.Get(k)))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}
