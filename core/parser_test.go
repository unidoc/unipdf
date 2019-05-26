/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
)

func makeReaderForText(txt string) (*bytes.Reader, *bufio.Reader, int64) {
	buf := []byte(txt)
	bufReader := bytes.NewReader(buf)
	bufferedReader := bufio.NewReader(bufReader)
	return bufReader, bufferedReader, int64(len(txt))
}

func makeParserForText(txt string) *PdfParser {
	rs, reader, fileSize := makeReaderForText(txt)
	return &PdfParser{rs: rs, reader: reader, fileSize: fileSize}
}

func BenchmarkSkipSpaces(b *testing.B) {
	parser := makeParserForText("       \t\t    \tABC")
	for n := 0; n < b.N; n++ {
		parser.skipSpaces()
		parser.SetFileOffset(0)
	}
}

var namePairs = map[string]string{
	"/Name1":                             "Name1",
	"/ASomewhatLongerName":               "ASomewhatLongerName",
	"/A;Name_With-Various***Characters?": "A;Name_With-Various***Characters?",
	"/1.2":                               "1.2",
	"/$$":                                "$$",
	"/@pattern":                          "@pattern",
	"/.notdef":                           ".notdef",
	"/Lime#20Green":                      "Lime Green",
	"/paired#28#29parentheses":           "paired()parentheses",
	"/The_Key_of_F#23_Minor":             "The_Key_of_F#_Minor",
	"/A#42":                              "AB",
	"/":                                  "",
	"/ ":                                 "",
	"/#3CBC88#3E#3CC5ED#3E#3CD544#3E#3CC694#3E": "<BC88><C5ED><D544><C694>",
}

func BenchmarkNameParsing(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for str, name := range namePairs {
			parser := makeParserForText(str)
			o, err := parser.parseName()
			if err != nil && err != io.EOF {
				b.Errorf("Unable to parse name string, error: %s", err)
			}
			if string(o) != name {
				b.Errorf("Mismatch %s != %s", o, name)
			}
		}
	}
}

func TestNameParsing(t *testing.T) {
	for str, name := range namePairs {
		parser := makeParserForText(str)
		o, err := parser.parseName()
		if err != nil && err != io.EOF {
			t.Errorf("Unable to parse name string, error: %s", err)
		}
		if string(o) != name {
			t.Errorf("Mismatch %s != %s", o, name)
		}
	}

	// Should fail (require starting with '/')
	parser := makeParserForText(" /Name")
	_, err := parser.parseName()
	if err == nil || err == io.EOF {
		t.Errorf("Should be invalid name")
	}
}

func TestBigDictParse(t *testing.T) {
	numObjects := 150000

	var buf bytes.Buffer
	buf.WriteString("<<")
	buf.WriteString("/ColorSpace <<")
	for i := 0; i < numObjects; i++ {
		buf.WriteString(fmt.Sprintf(`/Cs%d %d 0 R`, i, i))
	}
	buf.WriteString(">>")
	buf.WriteString("/Font <<>> ")
	buf.WriteString(">>")

	rs := bytes.NewReader(buf.Bytes())
	reader := bufio.NewReader(&buf)
	parser := &PdfParser{rs: rs, reader: reader, fileSize: int64(buf.Len())}

	val, err := parser.parseObject()
	require.NoError(t, err)
	require.NotNil(t, val)

	d, ok := GetDict(val)
	require.True(t, ok)
	require.Equal(t, 2, len(d.Keys()))

	d, ok = GetDict(d.Get("ColorSpace"))
	require.True(t, ok)
	require.Equal(t, numObjects, len(d.Keys()))
}

func BenchmarkStringParsing(b *testing.B) {
	entry := "(Strings may contain balanced parenthesis () and\nspecial characters (*!&}^% and so on).)"
	parser := makeParserForText(entry)
	for n := 0; n < b.N; n++ {
		_, err := parser.parseString()
		if err != nil && err != io.EOF {
			b.Errorf("Unable to parse string, error: %s", err)
		}
		parser.SetFileOffset(0)
	}
}

var stringPairs = map[string]string{
	"(This is a string)":                        "This is a string",
	"(Strings may contain\n newlines and such)": "Strings may contain\n newlines and such",
	"(Strings may contain balanced parenthesis () and\nspecial characters (*!&}^% and so on).)": "Strings may contain balanced parenthesis () and\nspecial characters (*!&}^% and so on).",
	"(These \\\ntwo strings \\\nare the same.)":                                                 "These two strings are the same.",
	"(These two strings are the same.)":                                                         "These two strings are the same.",
	"(\\\\)":                                                                                    "\\",
	"(This string has an end-of-line at the end of it.\n)":                                      "This string has an end-of-line at the end of it.\n",
	"(So does this one.\\n)":                                                                    "So does this one.\n",
	"(\\0053)":                                                                                  "\0053",
	"(\\53)":                                                                                    "\053",
	"(\\053)":                                                                                   "+",
	"(\\53\\101)":                                                                               "+A",
}

func TestStringParsing(t *testing.T) {
	for raw, expected := range stringPairs {
		parser := makeParserForText(raw)
		o, err := parser.parseString()
		if err != nil && err != io.EOF {
			t.Errorf("Unable to parse string, error: %s", err)
		}
		if o.Str() != expected {
			t.Errorf("String Mismatch %s: \"%s\" != \"%s\"", raw, o, expected)
		}
	}
}

func TestReadTextLine(t *testing.T) {
	// reading text ling + rewinding should be idempotent, that is:
	// if we rewind back len(str) bytes after reading string str we should arrive at beginning of str
	rawText := "abc\xb0cde"
	parser := makeParserForText(rawText)
	s, err := parser.readTextLine()
	if err != nil && err != io.EOF {
		t.Errorf("Unable to parse string, error: %s", err)
	}
	if parser.GetFileOffset() != int64(len(s)) {
		t.Errorf("File Offset after reading string of length %d is %d", len(s), parser.GetFileOffset())
	}
}

func TestBinStringParsing(t *testing.T) {
	// From an example O entry in Encrypt dictionary.
	rawText1 := "(\xE6\x00\xEC\xC2\x02\x88\xAD\x8B\\r\x64\xA9" +
		"\\)\xC6\xA8\x3E\xE2\x51\x76\x79\xAA\x02\x18\xBE\xCE\xEA" +
		"\x8B\x79\x86\x72\x6A\x8C\xDB)"

	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText1)
	o, err := parser.parseString()
	if err != nil && err != io.EOF {
		t.Errorf("Unable to parse string, error: %s", err)
	}
	if len(o.Str()) != 32 {
		t.Errorf("Wrong length, should be 32 (got %d)", len(o.Str()))
	}
}

// Main challenge in the text is "\\278A" which is "\\27" octal and 8A
func TestStringParsing2(t *testing.T) {
	rawText := "[(\\227\\224`\\274\\31W\\216\\276\\23\\231\\246U\\33\\317\\6-)(\\210S\\377:\\322\\278A\\200$*/e]\\371|)]"

	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	list, err := parser.parseArray()
	require.NoError(t, err)
	require.Equal(t, 2, list.Len())
}

func TestBoolParsing(t *testing.T) {
	// 7.3.2
	testEntries := map[string]bool{}
	testEntries["false"] = false
	testEntries["true"] = true

	for key, expected := range testEntries {
		parser := PdfParser{}
		parser.rs, parser.reader, parser.fileSize = makeReaderForText(key)
		val, err := parser.parseBool()
		require.NoError(t, err)
		require.Equal(t, expected, bool(val))
	}
}

func BenchmarkNumericParsing(b *testing.B) {
	txt1 := "[34.5 -3.62 1 +123.6 4. -.002 0.0]"
	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(txt1)

	for n := 0; n < b.N; n++ {
		_, err := parser.parseArray()
		require.NoError(b, err)
		parser.SetFileOffset(0)
	}
}

func TestNumericParsing1(t *testing.T) {
	// 7.3.3
	txt1 := "[34.5 -3.62 1 +123.6 4. -.002 0.0]"
	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(txt1)
	list, err := parser.parseArray()
	require.NoError(t, err)
	require.Equal(t, 7, list.Len())

	expectedFloats := map[int]float32{
		0: 34.5,
		1: -3.62,
		3: 123.6,
		4: 4.0,
		5: -0.002,
		6: 0.0,
	}

	for idx, val := range expectedFloats {
		num, ok := list.Get(idx).(*PdfObjectFloat)
		require.True(t, ok)
		require.Equal(t, val, float32(*num))
	}

	inum, ok := list.Get(2).(*PdfObjectInteger)
	require.True(t, ok)
	require.Equal(t, 1, int(*inum))
}

func TestNumericParsing2(t *testing.T) {
	// 7.3.3
	txt1 := "[+4.-.002]" // 4.0 and -0.002
	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(txt1)
	list, err := parser.parseArray()
	if err != nil {
		t.Errorf("Error parsing array")
		return
	}
	if list.Len() != 2 {
		t.Errorf("Len list != 2 (%d)", list.Len())
		return
	}

	expectedFloats := map[int]float32{
		0: 4.0,
		1: -0.002,
	}

	for idx, val := range expectedFloats {
		num, ok := list.Get(idx).(*PdfObjectFloat)
		if !ok {
			t.Errorf("Idx %d not float (%f)", idx, val)
			return
		}
		if float32(*num) != val {
			t.Errorf("Idx %d, value incorrect (%f)", idx, val)
		}
	}
}

func TestNumericParsingExponentials(t *testing.T) {
	testcases := []struct {
		RawObj   string
		Expected []float64
	}{
		{"[+4.-.002+3e-2-2e0]", []float64{4.0, -0.002, 0.03, -2.0}}, // 7.3.3.
		{"[-1E+35 1E+35]", []float64{-1e35, 1e35}},
	}

	for _, tcase := range testcases {
		t.Run(tcase.RawObj, func(t *testing.T) {
			parser := PdfParser{}
			parser.rs, parser.reader, parser.fileSize = makeReaderForText(tcase.RawObj)
			list, err := parser.parseArray()
			require.NoError(t, err)

			floats, err := list.ToFloat64Array()
			require.NoError(t, err)
			require.Equal(t, tcase.Expected, floats)
		})
	}
}

func BenchmarkHexStringParsing(b *testing.B) {
	var ref bytes.Buffer
	for i := 0; i < 0xff; i++ {
		ref.WriteByte(byte(i))
	}
	parser := makeParserForText("<" + hex.EncodeToString(ref.Bytes()) + ">")
	for n := 0; n < b.N; n++ {
		hs, err := parser.parseHexString()
		if err != nil {
			b.Errorf("Error parsing hex string: %s", err.Error())
			return
		}
		if hs.Str() != ref.String() {
			b.Errorf("Reference and parsed hex strings mismatch")
		}
		parser.SetFileOffset(0)
	}
}

func TestHexStringParsing(t *testing.T) {
	// 7.3.4.3
}

// TODO.
// Test reference to object outside of cross-ref table - should be 0
// Test xref object with offset 0, should be treated as 'f'ree.
// (compatibility with malformed writers).

func TestDictParsing1(t *testing.T) {
	txt1 := "<<\n\t/Name /Game /key/val/data\t[0 1 2 3.14 5]\t\n\n>>"
	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(txt1)
	dict, err := parser.ParseDict()
	if err != nil {
		t.Errorf("Error parsing dict")
	}

	if len(dict.Keys()) != 3 {
		t.Errorf("Length of dict != 3")
	}

	name, ok := dict.Get("Name").(*PdfObjectName)
	if !ok || *name != "Game" {
		t.Errorf("Value error")
	}

	key, ok := dict.Get("key").(*PdfObjectName)
	if !ok || *key != "val" {
		t.Errorf("Value error")
	}

	data, ok := dict.Get("data").(*PdfObjectArray)
	if !ok {
		t.Errorf("Invalid data")
	}
	integer, ok := data.Get(2).(*PdfObjectInteger)
	if !ok || *integer != 2 {
		t.Errorf("Wrong data")
	}

	float, ok := data.Get(3).(*PdfObjectFloat)
	if !ok || *float != 3.14 {
		t.Error("Wrong data")
	}
}

func TestDictParsing2(t *testing.T) {
	rawText := "<< /Type /Example\n" +
		"/Subtype /DictionaryExample /Version 0.01\n" +
		"/IntegerItem 12 \n" +
		"/StringItem (a string) /Subdictionary << /Item1 0.4\n" +
		"/Item2 true /LastItem (not!) /VeryLastItem (OK)\n" +
		">>\n >>"

	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	dict, err := parser.ParseDict()
	if err != nil {
		t.Errorf("Error parsing dict")
	}

	if len(dict.Keys()) != 6 {
		t.Errorf("Length of dict != 6")
	}

	typeName, ok := dict.Get("Type").(*PdfObjectName)
	if !ok || *typeName != "Example" {
		t.Errorf("Wrong type")
	}

	str, ok := dict.Get("StringItem").(*PdfObjectString)
	if !ok || str.Str() != "a string" {
		t.Errorf("Invalid string item")
	}

	subDict, ok := dict.Get("Subdictionary").(*PdfObjectDictionary)
	if !ok {
		t.Errorf("Invalid sub dictionary")
	}
	item2, ok := subDict.Get("Item2").(*PdfObjectBool)
	if !ok || *item2 != true {
		t.Errorf("Invalid bool item")
	}
	realnum, ok := subDict.Get("Item1").(*PdfObjectFloat)
	if !ok || *realnum != 0.4 {
		t.Errorf("Invalid real number")
	}
}

func TestDictParsing3(t *testing.T) {
	rawText := "<<>>"

	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	dict, err := parser.ParseDict()
	if err != nil {
		t.Errorf("Error parsing dict")
	}

	if len(dict.Keys()) != 0 {
		t.Errorf("Length of dict != 0")
	}
}

/*
func TestDictParsing4(t *testing.T) {
	rawText := "<</Key>>"

	parser := PdfParser{}
	parser.rs, parser.reader = makeReaderForText(rawText)
	dict, err := parser.ParseDict()
	if err != nil {
		t.Errorf("Error parsing dict (%s)", err)
		return
	}

	if len(*dict) != 1 {
		t.Errorf("Length of dict != 1")
		return
	}

	_, ok := (*dict)["Key"].(*PdfObjectNull)
	if !ok {
		t.Errorf("Invalid object (should be PDF null)")
		return
	}
}
*/

func TestArrayParsing(t *testing.T) {
	// 7.3.7.
}

func TestReferenceParsing(t *testing.T) {
	// TODO
}

func TestNullParsing(t *testing.T) {
	// TODO
}

func TestStreamParsing(t *testing.T) {
	// TODO
}

func TestIndirectObjParsing1(t *testing.T) {
	testcases := []struct {
		description string
		rawPDF      string
		checkFunc   func(obj PdfObject)
	}{
		{"Typical case",
			`1 0 obj
<<
/Names 2 0 R
/Pages 3 0 R
/Metadata 4 0 R
/ViewerPreferences
<<
/Rights
<<
/Document [/FullSave]
/TimeOfUbiquitization (D:20071210131309Z)
/RightsID [(x\\Ä-z<80><83>ã[W< b<99>\rhvèC©ðFüE^TN£^\jó]ç=çø\n<8f>:Ë¹\(<9a>\r=§^\~CÌÁxîÚð^V/=Î|Q\r<99>¢ ) (#$ÐJ^C<98>^ZX­<86>^TÞ¿ø¸^N]ú<8f>^N×2<9f>§ø±D^Q\r!'¡<8a>dp°,l¿<9d>É<82>«eæ§B­}«Ç8p·<97>\fl¿²G/x¹>) (kc2²µ^?-©¸þ$åiØ.Aé7^P½ÒÏð^S^^Y×rùç^OÌµ¶¿Hp^?*NËwóúËo§ü1ª<97>îFÜ\\<8f>OÚ^P[¸<93>0^)]
/Version 1
/Msg (This form has document rights applied to it.  These rights allow anyone completing this form, with the free Adobe Reader, to save their filled-in form locally.)
/Form [/Import /Export /SubmitStandalone /SpawnTemplate]
>>
>>
/AcroForm 5 0 R
/Type /Catalog
>>
endobj
3 0 obj
`,
			func(obj PdfObject) {
				indirect, ok := GetIndirect(obj)
				require.True(t, ok)
				require.NotNil(t, indirect)
				require.NotNil(t, indirect.PdfObject)
				require.Equal(t, int64(1), indirect.ObjectNumber)
				require.Equal(t, int64(0), indirect.GenerationNumber)

				dict, isDict := GetDict(indirect)
				require.True(t, isDict)

				dict, isDict = GetDict(dict.Get("ViewerPreferences"))
				require.True(t, isDict)
				require.Len(t, dict.Keys(), 1)

				dict, isDict = GetDict(dict.Get("Rights"))
				require.True(t, isDict)

				version, ok := GetIntVal(dict.Get("Version"))
				require.True(t, ok)
				require.Equal(t, 1, version)
			},
		},
		{
			"Basic object with short inner string",
			`1 0 obj
(a)
endobj
`, func(obj PdfObject) {
				indirect, ok := GetIndirect(obj)
				require.True(t, ok)
				require.NotNil(t, indirect)
				require.NotNil(t, indirect.PdfObject)
				str, ok := GetString(obj)
				require.True(t, ok)
				require.Equal(t, "a", str.String())
			},
		},
		{"Empty indirect object interpreted as containing null object",
			`1 0 obj
endobj
`,
			func(obj PdfObject) {
				indirect, ok := GetIndirect(obj)
				require.True(t, ok)
				require.NotNil(t, indirect)
				require.NotNil(t, indirect.PdfObject)
				require.True(t, IsNullObject(indirect.PdfObject))
			},
		},
	}

	for _, tcase := range testcases {
		t.Logf("%s", tcase.description)
		parser := PdfParser{}
		parser.rs, parser.reader, parser.fileSize = makeReaderForText(tcase.rawPDF)

		obj, err := parser.ParseIndirectObject()
		if err != nil && err != io.EOF {
			t.Errorf("Failed to parse indirect obj (%s)", err)
			return
		}

		tcase.checkFunc(obj)

		common.Log.Debug("Parsed obj: %s", obj)
	}
}

// Test /Prev and xref tables.  Check if the priority order is right.
// Test recovering xref tables. Refactor to recovery.go ?

func TestXrefStreamParse(t *testing.T) {
	rawText := `99 0 obj
<<  /Type /XRef
    /Index [0 5]
    /W [1 2 2]
    /Filter /ASCIIHexDecode
    /Size 5
    /Length 65
>>
stream
00 0000 FFFF
02 000F 0000
02 000F 0001
02 000F 0002
01 BA5E 0000>
endstream
endobj`
	parser := PdfParser{}
	parser.xrefs.ObjectMap = make(map[int]XrefObject)
	parser.objstms = make(objectStreams)
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)

	xrefDict, err := parser.parseXrefStream(nil)
	if err != nil {
		t.Errorf("Invalid xref stream object (%s)", err)
		return
	}

	typeName, ok := xrefDict.Get("Type").(*PdfObjectName)
	if !ok || *typeName != "XRef" {
		t.Errorf("Invalid Type != XRef")
		return
	}

	if len(parser.xrefs.ObjectMap) != 4 {
		t.Errorf("Wrong length (%d)", len(parser.xrefs.ObjectMap))
		return
	}

	if parser.xrefs.ObjectMap[3].XType != XrefTypeObjectStream {
		t.Errorf("Invalid type")
		return
	}
	if parser.xrefs.ObjectMap[3].OsObjNumber != 15 {
		t.Errorf("Wrong object stream obj number")
		return
	}
	if parser.xrefs.ObjectMap[3].OsObjIndex != 2 {
		t.Errorf("Wrong object stream obj index")
		return
	}

	common.Log.Debug("Xref dict: %s", xrefDict)
}

// TODO(gunnsth): Clear up. Should define clear inputs and expectation data and then run it.
func TestObjectParse(t *testing.T) {
	parser := PdfParser{}

	// Test object detection.
	// Invalid object type.
	rawText := " \t9 0 false"
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	obj, err := parser.parseObject()
	if err != nil {
		t.Error("Should ignore tab/space")
		return
	}

	// Integer
	rawText = "0"
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	obj, err = parser.parseObject()
	if err != nil {
		t.Errorf("Error parsing object: %v", err)
		return
	}
	nump, ok := obj.(*PdfObjectInteger)
	if !ok {
		t.Errorf("Unable to identify integer")
		return
	}
	if *nump != 0 {
		t.Errorf("Wrong value, expecting 9 (%d)", *nump)
		return
	}

	// Integer
	rawText = "9 0 false"
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	obj, err = parser.parseObject()

	if err != nil {
		t.Errorf("Error parsing object")
		return
	}
	nump, ok = obj.(*PdfObjectInteger)
	if !ok {
		t.Errorf("Unable to identify integer")
		return
	}
	if *nump != 9 {
		t.Errorf("Wrong value, expecting 9 (%d)", *nump)
		return
	}

	// Reference
	rawText = "9 0 R false"
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	obj, err = parser.parseObject()
	if err != nil {
		t.Errorf("Error parsing object")
		return
	}
	refp, ok := obj.(*PdfObjectReference)
	if !ok {
		t.Errorf("Unable to identify reference")
		return
	}
	if (*refp).ObjectNumber != 9 {
		t.Errorf("Wrong value, expecting object number 9")
		return
	}

	// Reference
	rawText = "909 0 R false"
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	obj, err = parser.parseObject()
	if err != nil {
		t.Errorf("Error parsing object")
		return
	}
	refp, ok = obj.(*PdfObjectReference)
	if !ok {
		t.Errorf("Unable to identify reference")
		return
	}
	if (*refp).ObjectNumber != 909 {
		t.Errorf("Wrong value, expecting object number 9")
		return
	}

	// Bool
	rawText = "false 9 0 R"
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	obj, err = parser.parseObject()
	if err != nil {
		t.Errorf("Error parsing object")
		return
	}
	boolp, ok := obj.(*PdfObjectBool)
	if !ok {
		t.Errorf("Unable to identify bool object")
		return
	}
	if *boolp != false {
		t.Errorf("Wrong value, expecting false")
		return
	}
}

// TestMinimalPDFFile test basic parsing of a minimal pdf file.
func TestMinimalPDFFile(t *testing.T) {
	file, err := os.Open("./testdata/minimal.pdf")
	require.NoError(t, err)
	defer file.Close()

	parser, err := NewParser(file)
	require.NoError(t, err)

	require.Len(t, parser.xrefs.ObjectMap, 4)
	require.Equal(t, 1, parser.xrefs.ObjectMap[1].ObjectNumber)
	require.Equal(t, int64(18), parser.xrefs.ObjectMap[1].Offset)
	require.Equal(t, XrefTypeTableEntry, parser.xrefs.ObjectMap[1].XType)
	require.Equal(t, 3, parser.xrefs.ObjectMap[3].ObjectNumber)
	require.Equal(t, int64(178), parser.xrefs.ObjectMap[3].Offset)
	require.Equal(t, XrefTypeTableEntry, parser.xrefs.ObjectMap[3].XType)

	// Check catalog object.
	catalogObj, err := parser.LookupByNumber(1)
	require.NoError(t, err)

	catalog, ok := catalogObj.(*PdfIndirectObject)
	require.True(t, ok)

	catalogDict, ok := catalog.PdfObject.(*PdfObjectDictionary)
	require.True(t, ok)

	typename, ok := catalogDict.Get("Type").(*PdfObjectName)
	require.True(t, ok)
	require.Equal(t, "Catalog", typename.String())

	// Check Page object.
	pageObj, err := parser.LookupByNumber(3)
	require.NoError(t, err)

	page, ok := pageObj.(*PdfIndirectObject)
	require.True(t, ok)
	pageDict, ok := page.PdfObject.(*PdfObjectDictionary)
	require.True(t, ok)
	require.Len(t, pageDict.Keys(), 4)

	resourcesDict, ok := pageDict.Get("Resources").(*PdfObjectDictionary)
	require.True(t, ok)
	require.Len(t, resourcesDict.Keys(), 1)

	fontDict, ok := resourcesDict.Get("Font").(*PdfObjectDictionary)
	require.True(t, ok)

	f1Dict, ok := fontDict.Get("F1").(*PdfObjectDictionary)
	require.True(t, ok)
	require.Len(t, f1Dict.Keys(), 3)

	baseFont, ok := f1Dict.Get("BaseFont").(*PdfObjectName)
	require.True(t, ok)
	require.Equal(t, "Times-Roman", baseFont.String())
}

// Test PDF version parsing.
func TestPDFVersionParse(t *testing.T) {
	// Test parsing when the version is at the start of the file.
	f1, err := os.Open("./testdata/minimal.pdf")
	require.NoError(t, err)
	defer f1.Close()

	parser := &PdfParser{
		rs:                                    f1,
		ObjCache:                              make(objectCache),
		streamLengthReferenceLookupInProgress: map[int64]bool{},
	}

	// Test parsed version.
	majorVersion, minorVersion, err := parser.parsePdfVersion()
	require.NoError(t, err)
	require.Equal(t, majorVersion, 1)
	require.Equal(t, minorVersion, 1)

	// Test file offset position.
	expected := "%PDF-1.1"
	b := make([]byte, len(expected))
	_, err = parser.reader.Read(b)
	require.NoError(t, err)
	require.Equal(t, string(b), expected)

	// Test parsing when the file has invalid data before the version.
	f2, err := os.Open("./testdata/invalidstart.pdf")
	require.NoError(t, err)
	defer f2.Close()

	parser = &PdfParser{
		rs:                                    f2,
		ObjCache:                              make(objectCache),
		streamLengthReferenceLookupInProgress: map[int64]bool{},
	}

	// Test parsed version.
	majorVersion, minorVersion, err = parser.parsePdfVersion()
	require.NoError(t, err)
	require.Equal(t, majorVersion, 1)
	require.Equal(t, minorVersion, 3)

	// Test file offset position.
	expected = "%PDF-1.3"
	b = make([]byte, len(expected))
	_, err = parser.reader.Read(b)
	require.NoError(t, err)
	require.Equal(t, string(b), expected)
}
