/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"io"
	"testing"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// ParseIndObjSeries loads a series of indirect objects until it runs into an error or EOF.
// Fully loads the objects and traverses resolving references to *PdfIndirectObjects.
// For use in testing where a series of indirect objects can be defined sequentially.
func (r *PdfReader) ParseIndObjSeries() error {
	for {
		obj, err := r.parser.ParseIndirectObject()
		if err != nil {
			if err != io.EOF {
				common.Log.Debug("Error parsing indirect object: %v", err)
				return err
			}
			break
		}

		switch t := obj.(type) {
		case *core.PdfObjectStream:
			r.parser.ObjCache[int(t.ObjectNumber)] = t
		case *core.PdfIndirectObject:
			r.parser.ObjCache[int(t.ObjectNumber)] = t
		default:
			common.Log.Debug("Incorrect type for ind obj: %T", obj)
			return ErrTypeCheck
		}
	}

	// Traverse the objects, resolving references to instances to PdfIndirectObject pointers.
	for _, obj := range r.parser.ObjCache {
		err := r.traverseObjectData(obj)
		if err != nil {
			common.Log.Debug("ERROR: Unable to traverse(%s)", err)
			return err
		}
	}

	return nil
}

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

	dateFromTime, err := NewPdfDateFromTime(date.ToGoTime())
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if dateFromTime.ToPdfObject().String() != date.ToPdfObject().String() {
		t.Errorf("Convert to and from time failed")
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
	dateFromTime, err = NewPdfDateFromTime(date.ToGoTime())
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if dateFromTime.ToPdfObject().String() != date.ToPdfObject().String() {
		t.Errorf("Convert to and from time failed")
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
	dateFromTime, err = NewPdfDateFromTime(date.ToGoTime())
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if dateFromTime.ToPdfObject().String() != date.ToPdfObject().String() {
		t.Errorf("Convert to and from time failed")
		return
	}

	// Case 4.  Another test from failed file.
	// Minutes not specified at end (assume is 0).
	str = "D:20061023115457-04'"
	date, err = NewPdfDate(str)
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if date.year != 2006 {
		t.Errorf("Year != 2006")
		return
	}
	if date.month != 10 {
		t.Errorf("month != 10")
		return
	}
	if date.day != 23 {
		t.Errorf("Day != 23")
		return
	}
	if date.hour != 11 {
		t.Errorf("Hour != 11 (%d)", date.hour)
		return
	}
	if date.minute != 54 {
		t.Errorf("Minute != 29 (%d)", date.minute)
	}
	if date.second != 57 {
		t.Errorf("Second != 37 (%d)", date.second)
		return
	}
	if date.utOffsetSign != '-' {
		t.Errorf("Invalid offset sign")
		return
	}
	if date.utOffsetHours != 4 {
		t.Errorf("Invalid offset hours")
		return
	}
	if date.utOffsetMins != 0 {
		t.Errorf("Invalid offset minutes")
		return
	}
	dateFromTime, err = NewPdfDateFromTime(date.ToGoTime())
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if dateFromTime.ToPdfObject().String() != date.ToPdfObject().String() {
		t.Errorf("Convert to and from time failed")
		return
	}

	// Case 5: Missing some more parameters.
	// Seems that many implementations consider some stuff optional...
	// Not following the standard, but we need to handle it.
	// D:20050823042205
	str = "D:20050823042205"
	date, err = NewPdfDate(str)
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if date.year != 2005 {
		t.Errorf("Year != 2005")
		return
	}
	if date.month != 8 {
		t.Errorf("month != 8")
		return
	}
	if date.day != 23 {
		t.Errorf("Day != 23")
		return
	}
	if date.hour != 04 {
		t.Errorf("Hour != 11 (%d)", date.hour)
		return
	}
	if date.minute != 22 {
		t.Errorf("Minute != 29 (%d)", date.minute)
	}
	if date.second != 05 {
		t.Errorf("Second != 37 (%d)", date.second)
		return
	}
	if date.utOffsetHours != 0 {
		t.Errorf("Invalid offset hours")
		return
	}
	if date.utOffsetMins != 0 {
		t.Errorf("Invalid offset minutes")
		return
	}
	dateFromTime, err = NewPdfDateFromTime(date.ToGoTime())
	if err != nil {
		t.Errorf("Fail: %s", err)
		return
	}
	if dateFromTime.ToPdfObject().String() != date.ToPdfObject().String() {
		t.Errorf("Convert to and from time failed")
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
	strObj, ok := obj.(*core.PdfObjectString)
	if !ok {
		t.Errorf("Date PDF object should be a string")
		return
	}
	if strObj.Str() != dateStr1 {
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
	parser := core.NewParserFromString(rawText)

	obj, err := parser.ParseIndirectObject()
	if err != nil {
		t.Errorf("Failed to parse indirect obj (%s)", err)
		return
	}

	pageObj, ok := obj.(*core.PdfIndirectObject)
	if !ok {
		t.Errorf("Invalid page object type != dictionary (%q)", obj)
		return
	}
	pageDict, ok := pageObj.PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		t.Errorf("Page object != dictionary")
		return
	}

	// Bit of a hacky way to do this.  PDF reader is needed if need to resolve external references,
	// but none in this case, so can just use a dummy instance.
	dummyPdfReader := PdfReader{}
	page, err := dummyPdfReader.newPdfPageFromDict(pageDict)
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

	//parser := PdfParser{}
	//parser.reader = makeReaderForText(rawText)
	parser := core.NewParserFromString(rawText)

	dict, err := parser.ParseDict()
	if err != nil {
		t.Errorf("Failed to parse dict obj (%s)", err)
		return
	}

	obj := dict.Get("MediaBox")
	arr, ok := obj.(*core.PdfObjectArray)
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
