package fonts

import (
	"path/filepath"
	"testing"

	"github.com/unidoc/unipdf/v3/internal/textencoding"
)

const fontDir = `../../../creator/testdata`

type charCode = textencoding.CharCode

var casesTTFParse = []struct {
	path         string
	name         string
	bold         bool
	italicAngle  float32
	underlinePos int16
	underlineTh  int16
	isFixed      bool
	bbox         [4]int
	runes        map[rune]charCode
	widths       map[rune]int
}{
	{
		path:         "FreeSans.ttf",
		name:         "FreeSans",
		underlinePos: -151,
		underlineTh:  50,
		bbox:         [4]int{-631, 1632, -462, 1230},
		runes: map[rune]charCode{
			'x': 0x5d,
			'ё': 0x32a,
		},
		widths: map[rune]int{
			'x': 500,
			'ё': 556,
		},
	},
	{
		path:         "roboto/Roboto-Bold.ttf",
		name:         "Roboto-Bold",
		bold:         true,
		underlinePos: -150,
		underlineTh:  100,
		bbox:         [4]int{-1488, 2439, -555, 2163},
		runes: map[rune]charCode{
			'x': 0x5c,
			'ё': 0x3cb,
		},
		widths: map[rune]int{
			'x': 1042,
			'ё': 1107,
		},
	},
	{
		path:         "roboto/Roboto-BoldItalic.ttf",
		name:         "Roboto-BoldItalic",
		bold:         true,
		italicAngle:  -12,
		underlinePos: -150,
		underlineTh:  100,
		bbox:         [4]int{-1459, 2467, -555, 2163},
		runes: map[rune]charCode{
			'x': 0x5c,
			'ё': 0x3cb,
		},
		widths: map[rune]int{
			'x': 1021,
			'ё': 1084,
		},
	},
}

var testRunes = []rune{'x', 'ё'}

func TestTTFParse(t *testing.T) {
	for _, c := range casesTTFParse {
		t.Run(c.path, func(t *testing.T) {
			path := filepath.Join(fontDir, c.path)

			ft, err := TtfParseFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if ft.Bold != c.bold {
				t.Error(ft.Bold, c.bold)
			}
			if float32(ft.ItalicAngle) != c.italicAngle {
				t.Error(ft.ItalicAngle, c.italicAngle)
			}
			if ft.UnderlinePosition != c.underlinePos {
				t.Error(ft.UnderlinePosition, c.underlinePos)
			}
			if ft.UnderlineThickness != c.underlineTh {
				t.Error(ft.UnderlineThickness, c.underlineTh)
			}
			if ft.IsFixedPitch != c.isFixed {
				t.Error(ft.IsFixedPitch, c.isFixed)
			}
			if b := [4]int{int(ft.Xmin), int(ft.Xmax), int(ft.Ymin), int(ft.Ymax)}; b != c.bbox {
				t.Error(b, c.bbox)
			}
			if ft.PostScriptName != c.name {
				t.Errorf("%q %q", ft.PostScriptName, c.name)
			}

			enc := textencoding.NewTrueTypeFontEncoder(ft.Chars)

			for _, r := range testRunes {
				t.Run(string(r), func(t *testing.T) {
					ind, ok := enc.RuneToCharcode(r)
					if !ok {
						t.Fatal("no char")
					} else if ind != c.runes[r] {
						t.Fatalf("%x != %x", ind, c.runes[r])
					}
					w := ft.Widths[ft.Chars[r]]
					if int(w) != c.widths[r] {
						t.Errorf("%d != %d", int(w), c.widths[r])
					}
				})
			}
		})
	}
}
