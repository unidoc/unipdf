/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"testing"

	"github.com/unidoc/unidoc/common"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

// cmap1Data represents a basic CMap.
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

// TestCMapParser tests basic loading of a simple CMap.
func TestCMapParser1(t *testing.T) {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

	cmap, err := LoadCmapFromDataCID([]byte(cmap1Data))
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

	if cmap.codespaces[0].Low != 0 {
		t.Errorf("code space low range != 0 (%d)", cmap.codespaces[0].Low)
		return
	}

	if cmap.codespaces[0].High != 0xFFFF {
		t.Errorf("code space high range != 0xffff (%d)", cmap.codespaces[0].High)
		return
	}

	expectedMappings := map[CharCode]rune{
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
	s, _ := cmap.CharcodeBytesToUnicode(charcodes)
	if s != " ," {
		t.Error("Incorrect charcode bytes -> string mapping")
		return
	}
}

const cmap2Data = `
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
	7 beginbfrange
	<0080> <00FF> <002C>
	<802F> <902F> <0038>
	endbfrange
	endcmap
	CMapName currentdict /CMap defineresource pop
	end
	end
`

// TestCMapParser2 tests a bug that came up when 2-byte character codes had the higher byte set to 0,
// e.g. 0x0080, and the character map was not taking the number of bytes of the input codemap into account.
func TestCMapParser2(t *testing.T) {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

	cmap, err := LoadCmapFromDataCID([]byte(cmap2Data))
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

	if cmap.codespaces[0].Low != 0 {
		t.Errorf("code space low range != 0 (%d)", cmap.codespaces[0].Low)
		return
	}

	if cmap.codespaces[0].High != 0xFFFF {
		t.Errorf("code space high range != 0xffff (%d)", cmap.codespaces[0].High)
		return
	}

	expectedMappings := map[CharCode]rune{
		0x0080: 0x002C,
		0x802F: 0x0038,
	}

	for k, expected := range expectedMappings {
		if v := cmap.CharcodeToUnicode(k); v != string(expected) {
			t.Errorf("incorrect mapping, expecting 0x%X -> 0x%X (got 0x%X)", k, expected, v)
			return
		}
	}

	// Check byte sequence mappings.
	expectedSequenceMappings := []struct {
		bytes    []byte
		expected string
	}{
		{[]byte{0x80, 0x2F, 0x00, 0x80}, string([]rune{0x0038, 0x002C})},
	}

	for _, exp := range expectedSequenceMappings {
		str, _ := cmap.CharcodeBytesToUnicode(exp.bytes)
		if str != exp.expected {
			t.Errorf("Incorrect byte sequence mapping % X -> % X (got % X)",
				exp.bytes, []rune(exp.expected), []rune(str))
			return
		}
	}
}

// cmapData3 is a CMap with a mixture of 1 and 2 byte codespaces.
const cmapData3 = `
	/CIDInit /ProcSet findresource begin
	12 dict begin begincmap
	/CIDSystemInfo
	3 dict dup begin
	/Registry (Adobe) def
	/Supplement 2 def
	end def

	/CMapName /test-1 def
	/CMapType 1 def

	4 begincodespacerange
	<00> <80>
	<8100> <9fff>
	<a0> <d0>
	<d140> <fbfc>
	endcodespacerange
	7 beginbfrange
	<00> <80> <10>
	<8100> <9f00> <1000>
	<a0> <d0> <90>
	<d140> <f000> <a000>
	endbfrange
	endcmap
`

// TestCMapParser3 test case of a CMap with mixed number of 1 and 2 bytes in the codespace range.
func TestCMapParser3(t *testing.T) {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

	cmap, err := LoadCmapFromDataCID([]byte(cmapData3))
	if err != nil {
		t.Error("Failed: ", err)
		return
	}

	if cmap.Name() != "test-1" {
		t.Errorf("CMap name incorrect (%s)", cmap.Name())
		return
	}

	if cmap.Type() != 1 {
		t.Errorf("CMap type incorrect")
		return
	}

	// Check codespaces.
	expectedCodespaces := []Codespace{
		{1, 0x00, 0x80},
		{1, 0xa0, 0xd0},
		{2, 0x8100, 0x9fff},
		{2, 0xd140, 0xfbfc},
	}

	if len(cmap.codespaces) != len(expectedCodespaces) {
		t.Errorf("len codespace != %d (%d)", len(expectedCodespaces), len(cmap.codespaces))
		return
	}

	for i, cs := range cmap.codespaces {
		exp := expectedCodespaces[i]
		if cs.NumBytes != exp.NumBytes {
			t.Errorf("code space number of bytes != %d (%d) %x", exp.NumBytes, cs.NumBytes, exp)
			return
		}

		if cs.Low != exp.Low {
			t.Errorf("code space low range != %d (%d) %x", exp.Low, cs.Low, exp)
			return
		}

		if cs.High != exp.High {
			t.Errorf("code space high range != 0x%X (0x%X) %x", exp.High, cs.High, exp)
			return
		}
	}

	// Check mappings.
	expectedMappings := map[CharCode]rune{
		0x80:   0x10 + 0x80,
		0x8100: 0x1000,
		0xa0:   0x90,
		0xd140: 0xa000,
	}
	for k, expected := range expectedMappings {
		if v := cmap.CharcodeToUnicode(k); v != string(expected) {
			t.Errorf("incorrect mapping: expecting 0x%02X -> 0x%02X (got 0x%02X)", k, expected, v)
			return
		}
	}

	// Check byte sequence mappings.
	expectedSequenceMappings := []struct {
		bytes    []byte
		expected string
	}{

		{[]byte{0x80, 0x81, 0x00, 0xa1, 0xd1, 0x80, 0x00},
			string([]rune{
				0x90,
				0x1000,
				0x91,
				0xa000 + 0x40,
				0x10})},
	}

	for _, exp := range expectedSequenceMappings {
		str, _ := cmap.CharcodeBytesToUnicode(exp.bytes)
		if str != exp.expected {
			t.Errorf("Incorrect byte sequence mapping: % 02X -> % 02X (got % 02X)",
				exp.bytes, []rune(exp.expected), []rune(str))
			return
		}
	}
}

// cmapData4 is a CMap with some utf16 encoded unicode strings that contain surrogates
const cmap4Data = `
    /CIDInit /ProcSet findresource begin
    11 dict begin
    begincmap
    /CIDSystemInfo
    << /Registry (Adobe)
    /Ordering (UCS)
    /Supplement 0
    >> def
    /CMapName /Adobe-Identity-UCS def
    /CMapType 2 def
    1 begincodespacerange
    <0000> <FFFF>
    endcodespacerange
    15 beginbfchar
    <01E1> <002C>
    <0201> <007C>
    <059C> <21D2>
    <05CA> <2200>
    <05CC> <2203>
    <05D0> <2208>
    <0652> <2295>
    <073F> <D835DC50>
    <0749> <D835DC5A>
    <0889> <D835DC84>
    <0893> <D835DC8E>
    <08DD> <D835DC9E>
    <08E5> <D835DCA6>
    <08E7> <2133>
    <0D52> <2265>
    endbfchar
    1 beginbfrange
    <0E36> <0E37> <27F5>
    endbfrange
    endcmap
`

// TestCMapParser4 checks that ut16 encoded unicode strings are interpreted correctly
func TestCMapParser4(t *testing.T) {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

	cmap, err := LoadCmapFromDataCID([]byte(cmap4Data))
	if err != nil {
		t.Error("Failed to load CMap: ", err)
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

	if cmap.codespaces[0].Low != 0 {
		t.Errorf("code space low range != 0 (%d)", cmap.codespaces[0].Low)
		return
	}

	if cmap.codespaces[0].High != 0xFFFF {
		t.Errorf("code space high range != 0xffff (%d)", cmap.codespaces[0].High)
		return
	}

	expectedMappings := map[CharCode]string{
		0x0889: "\U0001d484", // `𝒄`
		0x0893: "\U0001d48e", // `𝒎`
		0x08DD: "\U0001d49e", // `𝒞`
		0x08E5: "\U0001d4a6", // `𝒦
	}

	for k, expected := range expectedMappings {
		if v := cmap.CharcodeToUnicode(k); v != expected {
			t.Errorf("incorrect mapping, expecting 0x%04X -> %+q (got %+q)", k, expected, v)
			return
		}
	}

	// Check byte sequence mappings.
	expectedSequenceMappings := []struct {
		bytes    []byte
		expected string
	}{
		{[]byte{0x07, 0x3F, 0x07, 0x49}, "\U0001d450\U0001d45a"}, // `𝑐𝑚`
		{[]byte{0x08, 0x89, 0x08, 0x93}, "\U0001d484\U0001d48e"}, // `𝒄𝒎`
		{[]byte{0x08, 0xDD, 0x08, 0xE5}, "\U0001d49e\U0001d4a6"}, // `𝒞𝒦`
		{[]byte{0x08, 0xE7, 0x0D, 0x52}, "\u2133\u2265"},         // `ℳ≥`
	}

	for _, exp := range expectedSequenceMappings {
		str, _ := cmap.CharcodeBytesToUnicode(exp.bytes)
		if str != exp.expected {
			t.Errorf("Incorrect byte sequence mapping % 02X -> %+q (got %+q)",
				exp.bytes, exp.expected, str)
			return
		}
	}
}
