package model

import (
	"errors"
	"testing"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
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

// TestSimpleFonts checks that we correctly recreate simple fonts that we parse.
func TestSimpleFonts(t *testing.T) {
	for _, d := range simpleFontDicts {
		objFontObj(t, d)
	}
}

// objFontObj parses `fontDict` to a make a Font, creates a PDF object from the Font and checks that
// the new PDF object is the same as the input object
func objFontObj(t *testing.T, fontDict string) error {

	parser := NewParserFromString(fontDict)
	obj, err := parser.ParseDict()
	if err != nil {
		t.Errorf("objFontObj: Failed to parse dict obj. fontDict=%q err=%v", fontDict, err)
		return err
	}
	font, err := NewPdfFontFromPdfObject(obj)
	if err != nil {
		t.Errorf("Failed to parse font object. obj=%s err=%v", obj, err)
		return err
	}

	// Resolve all the indirect references in the font objects so we can compare their contents.
	obj1 := FlattenObject(obj)
	obj2 := FlattenObject(font.ToPdfObject())

	// Check that the reconstituted font is the same as the original.
	if !EqualObjects(obj1, obj2) {
		t.Errorf("Different objects.\nobj1=%s\nobj2=%s\nfont=%s", obj1, obj2, font)
		return errors.New("different objects")
	}

	return nil
}
