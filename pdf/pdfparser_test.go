/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/unidoc/unidoc/common"
)

func init() {
	common.SetLogger(common.ConsoleLogger{})
}

func makeReaderForText(txt string) *bufio.Reader {
	buf := []byte(txt)
	bufReader := bytes.NewReader(buf)
	bufferedReader := bufio.NewReader(bufReader)
	return bufferedReader
}

func TestNameParsing(t *testing.T) {
	namePairs := map[string]string{}

	namePairs["/Name1"] = "Name1"
	namePairs["/ASomewhatLongerName"] = "ASomewhatLongerName"
	namePairs["/A;Name_With-Various***Characters?"] = "A;Name_With-Various***Characters?"
	namePairs["/1.2"] = "1.2"
	namePairs["/$$"] = "$$"
	namePairs["/@pattern"] = "@pattern"
	namePairs["/.notdef"] = ".notdef"
	namePairs["/Lime#20Green"] = "Lime Green"
	namePairs["/paired#28#29parentheses"] = "paired()parentheses"
	namePairs["/The_Key_of_F#23_Minor"] = "The_Key_of_F#_Minor"
	namePairs["/A#42"] = "AB"
	namePairs["/"] = ""
	namePairs["/ "] = ""
	namePairs["/#3CBC88#3E#3CC5ED#3E#3CD544#3E#3CC694#3E"] = "<BC88><C5ED><D544><C694>"

	for str, name := range namePairs {
		parser := PdfParser{}
		parser.reader = makeReaderForText(str)
		o, err := parser.parseName()
		if err != nil && err != io.EOF {
			t.Errorf("Unable to parse name string, error: %s", err)
		}
		if string(o) != name {
			t.Errorf("Mismatch %s != %s", o, name)
		}
	}

	// Should fail (require starting with '/')
	parser := PdfParser{}
	parser.reader = makeReaderForText(" /Name")
	_, err := parser.parseName()
	if err == nil || err == io.EOF {
		t.Errorf("Should be invalid name")
	}
}

type testStringEntry struct {
	raw      string
	expected string
}

func TestStringParsing(t *testing.T) {
	testEntries := []testStringEntry{
		{"(This is a string)", "This is a string"},
		{"(Strings may contain\n newlines and such)", "Strings may contain\n newlines and such"},
		{"(Strings may contain balanced parenthesis () and\nspecial characters (*!&}^% and so on).)",
			"Strings may contain balanced parenthesis () and\nspecial characters (*!&}^% and so on)."},
		{"(These \\\ntwo strings \\\nare the same.)", "These two strings are the same."},
		{"(These two strings are the same.)", "These two strings are the same."},
		{"(\\\\)", "\\"},
		{"(This string has an end-of-line at the end of it.\n)",
			"This string has an end-of-line at the end of it.\n"},
		{"(So does this one.\\n)", "So does this one.\n"},
		{"(\\0053)", "\0053"},
		{"(\\053)", "\053"},
		{"(\\53)", "\053"},
		{"(\\053)", "+"},
		{"(\\53\\101)", "+A"},
	}
	for _, entry := range testEntries {
		parser := PdfParser{}
		parser.reader = makeReaderForText(entry.raw)
		o, err := parser.parseString()
		if err != nil && err != io.EOF {
			t.Errorf("Unable to parse string, error: %s", err)
		}
		if string(o) != entry.expected {
			t.Errorf("String Mismatch %s: \"%s\" != \"%s\"", entry.raw, o, entry.expected)
		}
	}
}

func TestBinStringParsing(t *testing.T) {
	// From an example O entry in Encrypt dictionary.
	rawText1 := "(\xE6\x00\xEC\xC2\x02\x88\xAD\x8B\\r\x64\xA9" +
		"\\)\xC6\xA8\x3E\xE2\x51\x76\x79\xAA\x02\x18\xBE\xCE\xEA" +
		"\x8B\x79\x86\x72\x6A\x8C\xDB)"

	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText1)
	o, err := parser.parseString()
	if err != nil && err != io.EOF {
		t.Errorf("Unable to parse string, error: %s", err)
	}
	if len(string(o)) != 32 {
		t.Errorf("Wrong length, should be 32 (got %d)", len(string(o)))
	}
}

// Main challenge in the text is "\\278A" which is "\\27" octal and 8A
func TestStringParsing2(t *testing.T) {
	rawText := "[(\\227\\224`\\274\\31W\\216\\276\\23\\231\\246U\\33\\317\\6-)(\\210S\\377:\\322\\278A\\200$*/e]\\371|)]"

	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText)
	list, err := parser.parseArray()
	if err != nil {
		t.Errorf("Failed to parse string list (%s)", err)
		return
	}
	if len(list) != 2 {
		t.Errorf("Length of list should be 2 (%d)", len(list))
		return
	}
}

func TestBoolParsing(t *testing.T) {
	// 7.3.2
	testEntries := map[string]bool{}
	testEntries["false"] = false
	testEntries["true"] = true

	for key, expected := range testEntries {
		parser := PdfParser{}
		parser.reader = makeReaderForText(key)
		val, err := parser.parseBool()
		if err != nil {
			t.Errorf("Error parsing bool: %s", err)
			return
		}
		if bool(val) != expected {
			t.Errorf("bool not as expected (%b)", val)
			return
		}
	}
}

func TestNumericParsing1(t *testing.T) {
	// 7.3.3
	txt1 := "[34.5 -3.62 1 +123.6 4. -.002 0.0]"
	parser := PdfParser{}
	parser.reader = makeReaderForText(txt1)
	list, err := parser.parseArray()
	if err != nil {
		t.Errorf("Error parsing array")
		return
	}
	if len(list) != 7 {
		t.Errorf("Len list != 7 (%d)", len(list))
		return
	}

	expectedFloats := map[int]float32{
		0: 34.5,
		1: -3.62,
		3: 123.6,
		4: 4.0,
		5: -0.002,
		6: 0.0,
	}

	for idx, val := range expectedFloats {
		num, ok := list[idx].(*PdfObjectFloat)
		if !ok {
			t.Errorf("Idx %d not float (%f)", idx, val)
			return
		}
		if float32(*num) != val {
			t.Errorf("Idx %d, value incorrect (%f)", idx)
		}
	}

	inum, ok := list[2].(*PdfObjectInteger)
	if !ok {
		t.Errorf("Number 3 not int")
		return
	}
	if *inum != 1 {
		t.Errorf("Number 3, val != 1")
		return
	}
}

func TestNumericParsing2(t *testing.T) {
	// 7.3.3
	txt1 := "[+4.-.002]" // 4.0 and -0.002
	parser := PdfParser{}
	parser.reader = makeReaderForText(txt1)
	list, err := parser.parseArray()
	if err != nil {
		t.Errorf("Error parsing array")
		return
	}
	if len(list) != 2 {
		t.Errorf("Len list != 2 (%d)", len(list))
		return
	}

	expectedFloats := map[int]float32{
		0: 4.0,
		1: -0.002,
	}

	for idx, val := range expectedFloats {
		num, ok := list[idx].(*PdfObjectFloat)
		if !ok {
			t.Errorf("Idx %d not float (%f)", idx, val)
			return
		}
		if float32(*num) != val {
			t.Errorf("Idx %d, value incorrect (%f)", idx)
		}
	}
}

// Includes exponential numbers.
func TestNumericParsing3(t *testing.T) {
	// 7.3.3
	txt1 := "[+4.-.002+3e-2-2e0]" // 4.0, -0.002, 1e-2, -2.0
	parser := PdfParser{}
	parser.reader = makeReaderForText(txt1)
	list, err := parser.parseArray()
	if err != nil {
		t.Errorf("Error parsing array (%s)", err)
		return
	}
	if len(list) != 4 {
		t.Errorf("Len list != 2 (%d)", len(list))
		return
	}

	expectedFloats := map[int]float32{
		0: 4.0,
		1: -0.002,
		2: 0.03,
		3: -2.0,
	}

	for idx, val := range expectedFloats {
		num, ok := list[idx].(*PdfObjectFloat)
		if !ok {
			t.Errorf("Idx %d not float (%f)", idx, val)
			return
		}
		if float32(*num) != val {
			t.Errorf("Idx %d, value incorrect (%f)", idx)
		}
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
	parser.reader = makeReaderForText(txt1)
	dict, err := parser.parseDict()
	if err != nil {
		t.Errorf("Error parsing dict")
	}

	if len(*dict) != 3 {
		t.Errorf("Length of dict != 3")
	}

	name, ok := (*dict)["Name"].(*PdfObjectName)
	if !ok || *name != "Game" {
		t.Errorf("Value error")
	}

	key, ok := (*dict)["key"].(*PdfObjectName)
	if !ok || *key != "val" {
		t.Errorf("Value error")
	}

	data, ok := (*dict)["data"].(*PdfObjectArray)
	if !ok {
		t.Errorf("Invalid data")
	}
	integer, ok := (*data)[2].(*PdfObjectInteger)
	if !ok || *integer != 2 {
		t.Errorf("Wrong data")
	}

	float, ok := (*data)[3].(*PdfObjectFloat)
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
	parser.reader = makeReaderForText(rawText)
	dict, err := parser.parseDict()
	if err != nil {
		t.Errorf("Error parsing dict")
	}

	if len(*dict) != 6 {
		t.Errorf("Length of dict != 6")
	}

	typeName, ok := (*dict)["Type"].(*PdfObjectName)
	if !ok || *typeName != "Example" {
		t.Errorf("Wrong type")
	}

	str, ok := (*dict)["StringItem"].(*PdfObjectString)
	if !ok || *str != "a string" {
		t.Errorf("Invalid string item")
	}

	subDict, ok := (*dict)["Subdictionary"].(*PdfObjectDictionary)
	if !ok {
		t.Errorf("Invalid sub dictionary")
	}
	item2, ok := (*subDict)["Item2"].(*PdfObjectBool)
	if !ok || *item2 != true {
		t.Errorf("Invalid bool item")
	}
	realnum, ok := (*subDict)["Item1"].(*PdfObjectFloat)
	if !ok || *realnum != 0.4 {
		t.Errorf("Invalid real number")
	}
}

func TestDictParsing3(t *testing.T) {
	rawText := "<<>>"

	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText)
	dict, err := parser.parseDict()
	if err != nil {
		t.Errorf("Error parsing dict")
	}

	if len(*dict) != 0 {
		t.Errorf("Length of dict != 0")
	}
}

/*
func TestDictParsing4(t *testing.T) {
	rawText := "<</Key>>"

	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText)
	dict, err := parser.parseDict()
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
	rawText := `1 0 obj
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
`
	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText)

	obj, err := parser.parseIndirectObject()
	if err != nil {
		t.Errorf("Failed to parse indirect obj (%s)", err)
		return
	}

	common.Log.Debug("Parsed obj: %s", obj)
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
	parser.xrefs = make(XrefTable)
	parser.objstms = make(ObjectStreams)
	parser.reader = makeReaderForText(rawText)

	xrefDict, err := parser.parseXrefStream(nil)
	if err != nil {
		t.Errorf("Invalid xref stream object (%s)", err)
		return
	}

	typeName, ok := (*xrefDict)["Type"].(*PdfObjectName)
	if !ok || *typeName != "XRef" {
		t.Errorf("Invalid Type != XRef")
		return
	}

	if len(parser.xrefs) != 4 {
		t.Errorf("Wrong length (%d)", len(parser.xrefs))
		return
	}

	if parser.xrefs[3].xtype != XREF_OBJECT_STREAM {
		t.Errorf("Invalid type")
		return
	}
	if parser.xrefs[3].osObjNumber != 15 {
		t.Errorf("Wrong object stream obj number")
		return
	}
	if parser.xrefs[3].osObjIndex != 2 {
		t.Errorf("Wrong object stream obj index")
		return
	}

	common.Log.Debug("Xref dict: %s", xrefDict)
}

func TestObjectParse(t *testing.T) {
	parser := PdfParser{}

	// Test object detection.
	// Invalid object type.
	rawText := " \t9 0 false"
	parser.reader = makeReaderForText(rawText)
	obj, err := parser.parseObject()
	if err != nil {
		t.Error("Should ignore tab/space")
		return
	}

	// Integer
	rawText = "9 0 false"
	parser.reader = makeReaderForText(rawText)
	obj, err = parser.parseObject()

	if err != nil {
		t.Errorf("Error parsing object")
		return
	}
	nump, ok := obj.(*PdfObjectInteger)
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
	parser.reader = makeReaderForText(rawText)
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
	parser.reader = makeReaderForText(rawText)
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
	parser.reader = makeReaderForText(rawText)
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

var file1 = "../testfiles/minimal.pdf"

func TestMinimalPDFFile(t *testing.T) {
	file, err := os.Open(file1)
	if err != nil {
		t.Errorf("Unable to open minimal test file (%s)", err)
		return
	}
	defer file.Close()

	reader, err := NewPdfReader(file)
	if err != nil {
		t.Errorf("Unable to read test file (%s)", err)
		return
	}

	numPages, err := reader.GetNumPages()
	if err != nil {
		t.Error("Unable to get number of pages")
	}

	fmt.Printf("Num pages: %d\n", numPages)
	if numPages != 1 {
		t.Error("Wrong number of pages")
	}

	parser := reader.parser
	if len(parser.xrefs) != 4 {
		t.Errorf("Wrong number of xrefs %d != 4", len(parser.xrefs))
	}

	if parser.xrefs[1].objectNumber != 1 {
		t.Errorf("Invalid xref0 object number != 1 (%d)", parser.xrefs[0].objectNumber)
	}
	if parser.xrefs[1].offset != 18 {
		t.Errorf("Invalid offset != 18 (%d)", parser.xrefs[0].offset)
	}
	if parser.xrefs[1].xtype != XREF_TABLE_ENTRY {
		t.Errorf("Invalid xref type")
	}
	if parser.xrefs[3].objectNumber != 3 {
		t.Errorf("Invalid xref object number != 3 (%d)", parser.xrefs[2].objectNumber)
	}
	if parser.xrefs[3].offset != 178 {
		t.Errorf("Invalid offset != 178")
	}
	if parser.xrefs[3].xtype != XREF_TABLE_ENTRY {
		t.Errorf("Invalid xref type")
	}

	// Check catalog object.
	catalogObj, err := parser.LookupByNumber(1)
	if err != nil {
		t.Error("Unable to look up catalog object")
	}
	catalog, ok := catalogObj.(*PdfIndirectObject)
	if !ok {
		t.Error("Unable to look up catalog object")
	}
	catalogDict, ok := catalog.PdfObject.(*PdfObjectDictionary)
	if !ok {
		t.Error("Unable to find dictionary")
	}
	typename, ok := (*catalogDict)["Type"].(*PdfObjectName)
	if !ok {
		t.Error("Unable to check type")
	}
	if *typename != "Catalog" {
		t.Error("Wrong type name (%s != Catalog)", *typename)
	}

	// Check Page object.
	pageObj, err := parser.LookupByNumber(3)
	if err != nil {
		t.Error("Unable to look up Page")
	}
	page, ok := pageObj.(*PdfIndirectObject)
	if !ok {
		t.Error("Unable to look up Page")
	}
	pageDict, ok := page.PdfObject.(*PdfObjectDictionary)
	if !ok {
		t.Error("Unable to load Page dictionary")
	}
	if len(*pageDict) != 4 {
		t.Error("Page dict should have 4 objects (%d)", len(*pageDict))
	}
	resourcesDict, ok := (*pageDict)["Resources"].(*PdfObjectDictionary)
	if !ok {
		t.Error("Unable to load Resources dictionary")
	}
	if len(*resourcesDict) != 1 {
		t.Error("Page Resources dict should have 1 member (%d)", len(*resourcesDict))
	}
	fontDict, ok := (*resourcesDict)["Font"].(*PdfObjectDictionary)
	if !ok {
		t.Error("Unable to load font")
	}
	f1Dict, ok := (*fontDict)["F1"].(*PdfObjectDictionary)
	if !ok {
		t.Error("Unable to load F1 dict")
	}
	if len(*f1Dict) != 3 {
		t.Error("Invalid F1 dict length 3 != %d", len(*f1Dict))
	}
	baseFont, ok := (*f1Dict)["BaseFont"].(*PdfObjectName)
	if !ok {
		t.Error("Unable to load base font")
	}
	if *baseFont != "Times-Roman" {
		t.Error("Invalid base font (should be Times-Roman not %s)", *baseFont)
	}
}
