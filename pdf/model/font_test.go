package model_test

import (
	"errors"
	"fmt"
	"io"
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

// TestTrueTypeToUnicode checks that CharcodeBytesToUnicode is working for a TrueType font with a
// ToUnicode cmap.
func TestTrueTypeToUnicode(t *testing.T) {
	numObj, err := parsePdfObjects(ttToUnicode)
	if err != nil {
		t.Errorf("Failed to parse ttToUnicode object. err=%v", err)
		return
	}
	fontObj := numObj[9]
	font, err := model.NewPdfFontFromPdfObject(fontObj)
	if err != nil {
		t.Errorf("Failed to create font. err=%v", err)
		return
	}

	data := []byte{43, 40, 41, 34, 37, 42, 38, 49, 36, 38, 48, 34, 35, 36, 37, 35, 36, 58}
	expectedText := "Alerts on printing"
	actualText, numChars, numMisses := font.CharcodeBytesToUnicode(data)
	if numMisses != 0 {
		t.Errorf("Some codes not decoded. numMisses=%d", numMisses)
		return
	}
	if actualText != expectedText {
		t.Errorf("Incorrect decoding.\nexpected=%q\n  actual=%q", expectedText, actualText)
	}
	if numChars != len(actualText) {
		t.Errorf("Incorrect numChars=%d expected=%d", numChars, len(actualText))
	}
}

// ttToUnicode is a TrueType font object and its ToUnicode cmap.
// The stream data in obj 26 (the ToUnicode cmap) is Sprintf'd to avoid binary data in the `` string.
var ttToUnicode = fmt.Sprintf(`9 0 obj
<< /Type /Font /Subtype /TrueType /BaseFont /AHSHJL+.SFUIText /ToUnicode 26 0 R /FirstChar 33 /LastChar 79 /Widths [ 635 381 246
583 363 282 609 252 571 523 674 560 594 542 543 609 591 637 584 874 614 614
362 246 268 609 742 870 644 596 604 771 716 653 297 525 657 543 297 774 539
382 382 726 968 295 538 ] >>
endobj
26 0 obj
<< /Length 497 /Filter /FlateDecode >>
stream
%s
endstream
endobj
`,
	"\x78\x01\x5d\x93\xcd\x8a\xdb\x30\x14\x46\xf7\x7e\x0a\x2d\xa7\x8b\xc1\x8a\xe5\x24\x33\x60\x0c\xc3\x94\x81\x2c\xfa\x43\xd3\x3e\x80\x6d\xc9\xc1\xd0\xd8\xc6\x71\x16\x79\xfb\x9e\xef\x66\x3a\x85\x2e\xbe\xc5\xf1\xd5\x55\xee\x51\xa4\xfc\xf5\xf0\xf9\x30\x0e\xab\xcb\xbf\x2f\x53\x77\x4c\xab\xeb\x87\x31\x2e\xe9\x32\x5d\x97\x2e\xb9\x36\x9d\x86\x31\xdb\x14\x2e\x0e\xdd\xfa\x4e\xf6\xad\x3b\x37\x73\x96\xd3\x7c\xbc\x5d\xd6\x74\x3e\x8c\xfd\xe4\xaa\x2a\x73\x2e\xff\x41\xcb\x65\x5d\x6e\xee\xe1\x25\x4e\x6d\xfa\xa4\x6f\xdf\x96\x98\x96\x61\x3c\xb9\x87\x5f\xaf\x47\xfb\x72\xbc\xce\xf3\xef\x74\x4e\xe3\xea\x7c\x56\xd7\x2e\xa6\x9e\xed\xbe\x34\xf3\xd7\xe6\x9c\x5c\x6e\xad\x8f\x87\x48\x7d\x58\x6f\x8f\x74\xfd\x5b\xf1\xf3\x36\x27\xc7\x44\x74\x6c\xee\x23\x75\x53\x4c\x97\xb9\xe9\xd2\xd2\x8c\xa7\x94\x55\xde\xd7\xd5\xdb\x5b\x9d\xa5\x31\xfe\x57\x2a\xb7\xf7\x8e\xb6\x7f\x5f\x5a\x6c\xea\x4a\xf1\x7e\xeb\xeb\xac\x2a\x0a\x90\x78\xbf\x2f\x84\x01\x24\xde\xef\x9e\x85\x25\x48\xc0\x24\xdc\x82\x84\xc5\xa5\x70\x07\x12\xef\x0b\xdb\xea\x09\x24\x2c\xee\x54\x7d\x06\x09\xb8\x15\x36\x20\xa1\x37\x08\x5b\x90\x78\x5f\x6e\x84\x1d\x48\x58\x6c\xd5\x08\x12\xf0\x49\xd5\x04\x12\x7a\x77\xc2\x1e\x24\xa0\x86\x0c\xc8\x2b\xa0\xc6\x08\xc8\x29\xf4\xf6\x42\xe4\x14\x7c\xb5\x73\x40\x4e\x61\xb1\xa6\x0a\xc8\x29\x8c\x11\x85\xc8\x29\xf4\xea\x34\x02\x72\x0a\x28\xdf\xb0\x07\x09\xa8\x31\x02\xae\x0a\xd8\x08\x71\x55\xd8\xca\xa6\xc2\x35\x98\xef\x6e\xaf\x2a\xae\x0a\x55\x9d\x64\xc0\x55\xa1\xd7\x7e\x17\xd7\x60\xbe\x0c\x43\x15\x57\x85\xc5\x36\x24\xae\xc1\x7c\x11\xc9\xaa\x12\x57\x85\xaa\x04\x39\x3f\x0b\x28\xc1\x12\x57\x05\x5f\x5b\x8c\x2b\xdf\x41\x0e\x90\x2a\xae\x0a\xfa\x3a\x58\xb6\xb7\xd0\x6b\x8b\x71\x2d\xef\xbe\xad\xaa\xb8\x2a\xf4\xea\x0f\x2d\x71\x55\xe8\x95\x11\x96\x16\x50\xfa\x25\xae\xa5\x09\x72\x07\x40\xe4\x14\x76\x96\x11\xa7\x6b\x61\x2a\xeb\x45\x8e\x73\xa8\x7c\xd1\xda\xce\xc8\xe1\xa2\xc5\x6c\xc5\x25\xfe\x7b\x5b\x75\x9f\xf5\xee\x3e\xde\x49\x77\x5d\x16\x9e\x88\x3d\x4e\x7b\x3d\x7a\x15\xc3\x98\x3e\xde\xef\x3c\xcd\xda\xc0\xf2\x07\xc5\xfd\x00\x4f")

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

// parsePdfObjects parses a fragment of a PDF `text` (e.g. ttToUnicode above) and returns a map of
// {object number: object} with indirect objects replaced by their values if they are in `text`.
func parsePdfObjects(text string) (map[int64]core.PdfObject, error) {
	numObj := map[int64]core.PdfObject{}
	parser := core.NewParserFromString(text)
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
			numObj[t.ObjectNumber] = obj
		case *core.PdfObjectStream:
			numObj[t.ObjectNumber] = obj
		}
	}

	for _, obj := range numObj {
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
				if o, ok := numObj[ref.ObjectNumber]; ok {
					dict.Set(k, o)
				}
			}
		}
	}
	return numObj, nil
}

// func isFontObject(obj core.PdfObject) bool {
// 	var dict *core.PdfObjectDictionary
// 	switch t := obj.(type) {
// 	case *core.PdfIndirectObject:
// 		dict = t.PdfObject.(*core.PdfObjectDictionary)
// 	case *core.PdfObjectDictionary:
// 		dict = t
// 	default:
// 		return false
// 	}
// 	name, err := core.GetName(dict.Get("Type"))
// 	return err == nil && name == "Font"
// }

// func showDict(dict *core.PdfObjectDictionary) string {
// 	parts := []string{}
// 	for _, k := range dict.Keys() {
// 		parts = append(parts, fmt.Sprintf("%s: %T", k, dict.Get(k)))
// 	}
// 	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
// }
