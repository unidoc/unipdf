package cmap

import (
	"testing"

	"github.com/unidoc/unidoc/common"
)

func init() {
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
}

const cmap1Data = `
/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo
<<  /Registry (Adobe)
/Ordering (UCS)
/Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
8 beginbfchar
<0003> <0020>
<0007> <0024>
<0033> <0050>
<0035> <0052>
<0037> <0054>
<005A> <0077>
<005C> <0079>
<005F> <007C>
endbfchar
7 beginbfrange
<000F> <0017> <002C>
<001B> <001D> <0038>
<0025> <0026> <0042>
<002F> <0031> <004C>
<0044> <004C> <0061>
<004F> <0053> <006C>
<0055> <0057> <0072>
endbfrange
endcmap
CMapName currentdict /CMap defineresource pop
end
end
`

func TestCMapParser1(t *testing.T) {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))

	cmap, err := LoadCmapFromData([]byte(cmap1Data))
	if err != nil {
		t.Error("Failed: ", err)
		return
	}

	if cmap.Name() != "Adobe-Identity-UCS" {
		t.Errorf("CMap name incorrect (%s)", cmap.Name())
		return
	}

	if cmap.Type() != 2 {
		t.Errorf("CMap type incorrect")
		return
	}

	if len(cmap.codespaces) != 1 {
		t.Errorf("len codespace != 1 (%d)", len(cmap.codespaces))
		return
	}

	if cmap.codespaces[0].low != 0 {
		t.Errorf("code space low range != 0 (%d)", cmap.codespaces[0].low)
		return
	}

	if cmap.codespaces[0].high != 0xFFFF {
		t.Errorf("code space high range != 0xffff (%d)", cmap.codespaces[0].high)
		return
	}

	expectedMappings := map[uint64]rune{
		0x0003:     0x0020,
		0x005F:     0x007C,
		0x000F:     0x002C,
		0x000F + 5: 0x002C + 5,
		0x001B:     0x0038,
		0x001B + 2: 0x0038 + 2,
		0x002F:     0x004C,
		0x0044:     0x0061,
		0x004F:     0x006C,
		0x0055:     0x0072,
	}

	for k, expected := range expectedMappings {
		if v := cmap.CharcodeToUnicode(k); v != string(expected) {
			t.Errorf("incorrect mapping, expecting 0x%X -> 0x%X (%#v)", k, expected, v)
			return
		}
	}

	v := cmap.CharcodeToUnicode(0x99)
	if v != "?" { //!= "notdef" {
		t.Errorf("Unmapped code, expected to map to undefined")
		return
	}

	charcodes := []byte{0x00, 0x03, 0x00, 0x0F}
	s := cmap.CharcodeBytesToUnicode(charcodes)
	if s != " ," {
		t.Error("Incorrect charcode bytes -> string mapping")
		return
	}
}
