/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"testing"

	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/testutils"
)

// Test loading of a basic checkbox field with a merged-in annotation.
func TestCheckboxField1(t *testing.T) {
	rawText := `
1 0 obj
<<
/Type /Annot
/Subtype /Widget
/Rect [100 100 120 120]
/FT /Btn
/T (Urgent)
/V /Yes
/AS /Yes
/AP <</N <</Yes 2 0 R /Off 3 0 R>> >>
>>
endobj

2 0 obj
<</Type /XObject
/Subtype /Form
/BBox [0 0 20 20]
/Resources 20 0 R
/Length 44
>>
stream
q
0 0 1 rg
BT
/ZaDb 12 Tf
0 0 Td
(4) Tj
ET
Q
endstream
endobj

3 0 obj
<</Type /XObject
/Subtype /Form
/BBox [0 0 20 20]
/Resources 20 0 R
/Length 51
>>
stream
q
0 0 1 rg
BT
/ZaDb 12 Tf
0 0 Td
(8) Tj
ET
Q
endstream
endobj

4 0 obj
% Copy of obj 1 except not with merged-in annotation
<<
/FT /Btn
/T (Urgent)
/V /Yes
/Kids [5 0 R]
>>
endobj

5 0 obj
<<
/Type /Annot
/Subtype /Widget
/Rect [100 100 120 120]
/AS /Yes
/AP <</N <</Yes 2 0 R /Off 3 0 R>> >>
/Parent 4 0 R
>>
endobj
`
	r := NewReaderForText(rawText)

	err := r.ParseIndObjSeries()
	if err != nil {
		t.Fatalf("Failed loading indirect object series: %v", err)
	}

	// Load the field from object number 1.
	obj, err := r.parser.LookupByNumber(1)
	if err != nil {
		t.Fatalf("Failed to parse indirect obj (%s)", err)
	}

	ind, ok := obj.(*core.PdfIndirectObject)
	if !ok {
		t.Fatalf("Incorrect type (%T)", obj)
	}

	field, err := r.newPdfFieldFromIndirectObject(ind, nil)
	if err != nil {
		t.Fatalf("Unable to load field (%v)", err)
		return
	}

	// Check properties of the field.
	buttonf, ok := field.GetContext().(*PdfFieldButton)
	if !ok {
		t.Errorf("Field content incorrect (%T)", field.GetContext())
		return
	}
	if buttonf == nil {
		t.Fatalf("buttonf is nil")
	}

	if len(field.Kids) > 0 {
		t.Fatalf("Field should not have kids")
	}

	if len(field.Annotations) != 1 {
		t.Fatalf("Field should have a single annotation")
	}

	// Field -> PDF object.  Regenerate the field dictionary and see if matches expectations.
	// Reset the dictionaries for both field and annotation to avoid re-use during re-generation of PDF object.
	field.container = core.MakeIndirectObject(core.MakeDict())
	field.Annotations[0].primitive = core.MakeIndirectObject(core.MakeDict())
	fieldPdfObj := field.ToPdfObject()
	fieldDict, ok := fieldPdfObj.(*core.PdfIndirectObject).PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		t.Fatalf("Type error")
	}

	// Load the expected field dictionary (output).  Slightly different than original as the input had
	// a merged-in annotation. Our output does not currently merge annotations.
	obj, err = r.parser.LookupByNumber(4)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	expDict, ok := obj.(*core.PdfIndirectObject).PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		t.Fatalf("Unable to load expected dict")
	}

	if !testutils.CompareDictionariesDeep(expDict, fieldDict) {
		t.Fatalf("Mismatch in expected and actual field dictionaries (deep)")
	}
}
