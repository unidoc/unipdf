/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

package pdf

import (
	"testing"
)

// Test PDF date parsing from string.
func TestDateParse(t *testing.T) {
	// Case 1. Test everything.
	str := "D:20080313232937+01'00'"
	date, err := NewPdfDate(str)
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if date.year != 2008 {
		t.Errorf("Year != 2008")
		return
	}
	if date.month != 3 {
		t.Errorf("month != 3")
		return
	}
	if date.day != 13 {
		t.Errorf("Day != 13")
		return
	}
	if date.hour != 23 {
		t.Errorf("Hour != 23 (%d)", date.hour)
		return
	}
	if date.minute != 29 {
		t.Errorf("Minute != 29 (%d)", date.minute)
	}
	if date.second != 37 {
		t.Errorf("Second != 37 (%d)", date.second)
		return
	}
	if date.utOffsetSign != '+' {
		t.Errorf("Invalid offset sign")
		return
	}
	if date.utOffsetHours != 1 {
		t.Errorf("Invalid offset hours")
		return
	}
	if date.utOffsetMins != 0 {
		t.Errorf("Invalid offset minutes")
		return
	}

	// Case 2: Negative sign.
	str = "D:20150811050933-07'00'"
	date, err = NewPdfDate(str)
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if date.utOffsetSign != '-' {
		t.Errorf("Invalid offset sign")
		return
	}
	if date.utOffsetHours != 7 {
		t.Errorf("Invalid offset hours")
		return
	}

	// Case 3. Offset minutes.
	str = "D:20110807220047+09'30'"
	date, err = NewPdfDate(str)
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if date.utOffsetMins != 30 {
		t.Errorf("Offset mins != 30")
		return
	}
}

// Test parsing and building the date.
func TestPdfDateBuild(t *testing.T) {
	// Case 1. Test everything.
	dateStr1 := "D:20080313232937+01'00'"
	date, err := NewPdfDate(dateStr1)
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}

	obj := date.ToPdfObject()
	strObj, ok := obj.(*PdfObjectString)
	if !ok {
		t.Errorf("Date PDF object should be a string")
		return
	}
	if string(*strObj) != dateStr1 {
		t.Errorf("Built date string does not match original (%s)", strObj)
		return
	}
}

// Test page loading.
func TestPdfPage1(t *testing.T) {
	rawText := `
9 0 obj
<<
  /Type /Page
  /Parent 3 0 R
  /MediaBox [0 0 612 459]
  /Contents 13 0 R
  /Resources <<
    /ProcSet 11 0 R
    /ExtGState <<
      /GS0 << /BM /Normal >>
    >>
    /XObject <</Im0 12 0 R>>
  >>
>>
endobj
    `
	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText)

	obj, err := parser.parseIndirectObject()
	if err != nil {
		t.Errorf("Failed to parse indirect obj (%s)", err)
		return
	}

	pageObj, ok := obj.(*PdfIndirectObject)
	if !ok {
		t.Errorf("Invalid page object type != dictionary (%q)", obj)
		return
	}
	pageDict, ok := pageObj.PdfObject.(*PdfObjectDictionary)
	if !ok {
		t.Errorf("Page object != dictionary")
		return
	}

	page, err := NewPdfPage(*pageDict)
	if err != nil {
		t.Errorf("Unable to load page (%s)", err)
		return
	}

	if page.MediaBox.Llx != 0 || page.MediaBox.Lly != 0 {
		t.Errorf("llx, lly != 0,0")
		return
	}

	if page.MediaBox.Urx != 612 || page.MediaBox.Ury != 459 {
		t.Errorf("urx, ury!= 612 (%f), 459 (%f)", page.MediaBox.Urx, page.MediaBox.Ury)
		return
	}
}

// Test rectangle parsing and loading.
func TestRect(t *testing.T) {
	rawText := `<< /MediaBox [0 0 613.644043 802.772034] >>`

	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText)

	dict, err := parser.parseDict()
	if err != nil {
		t.Errorf("Failed to parse dict obj (%s)", err)
		return
	}

	obj, _ := (*dict)["MediaBox"]
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		t.Errorf("Type != Array")
		return
	}

	rect, err := NewPdfRectangle(*arr)
	if err != nil {
		t.Errorf("Failed to create rectangle (%s)", err)
		return
	}

	if rect.Llx != 0 {
		t.Errorf("rect.llx != 0 (%f)", rect.Llx)
		return
	}

	if rect.Lly != 0 {
		t.Errorf("rect.lly != 0 (%f)", rect.Lly)
		return
	}

	if rect.Urx != 613.644043 {
		t.Errorf("rect.urx != 613.644043 (%f)", rect.Urx)
		return
	}
	if rect.Ury != 802.772034 {
		t.Errorf("rect.urx != 802.772034 (%f)", rect.Ury)
		return
	}
}
