/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"testing"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

func compareDictionariesDeep(d1, d2 *core.PdfObjectDictionary) bool {
	if len(d1.Keys()) != len(d2.Keys()) {
		common.Log.Debug("Dict entries mismatch (%d != %d)", len(d1.Keys()), len(d2.Keys()))
		common.Log.Debug("Was '%s' vs '%s'", d1.DefaultWriteString(), d2.DefaultWriteString())
		return false
	}

	for _, k := range d1.Keys() {
		if k == "Parent" {
			continue
		}
		v1 := core.TraceToDirectObject(d1.Get(k))
		v2 := core.TraceToDirectObject(d2.Get(k))

		if v1 == nil {
			common.Log.Debug("v1 is nil")
			return false
		}
		if v2 == nil {
			common.Log.Debug("v2 is nil")
			return false
		}

		switch t1 := v1.(type) {
		case *core.PdfObjectDictionary:
			t2, ok := v2.(*core.PdfObjectDictionary)
			if !ok {
				common.Log.Debug("Type mismatch %T vs %T", v1, v2)
				return false
			}
			if !compareDictionariesDeep(t1, t2) {
				return false
			}
			continue

		case *core.PdfObjectArray:
			t2, ok := v2.(*core.PdfObjectArray)
			if !ok {
				common.Log.Debug("v2 not an array")
				return false
			}
			if t1.Len() != t2.Len() {
				common.Log.Debug("array length mismatch (%d != %d)", t1.Len(), t2.Len())
				return false
			}
			for i := 0; i < t1.Len(); i++ {
				v1 := core.TraceToDirectObject(t1.Get(i))
				v2 := core.TraceToDirectObject(t2.Get(i))
				if d1, isD1 := v1.(*core.PdfObjectDictionary); isD1 {
					d2, isD2 := v2.(*core.PdfObjectDictionary)
					if !isD2 {
						return false
					}
					if !compareDictionariesDeep(d1, d2) {
						return false
					}
				} else {
					if v1.DefaultWriteString() != v2.DefaultWriteString() {
						common.Log.Debug("Mismatch '%s' != '%s'", v1.DefaultWriteString(), v2.DefaultWriteString())
						return false
					}
				}
			}
			continue
		}

		if v1.String() != v2.String() {
			common.Log.Debug("key=%s Mismatch! '%s' != '%s'", k, v1.String(), v2.String())
			common.Log.Debug("For '%T' - '%T'", v1, v2)
			common.Log.Debug("For '%+v' - '%+v'", v1, v2)
			return false
		}
	}

	return true
}

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

	if !compareDictionariesDeep(expDict, fieldDict) {
		t.Fatalf("Mismatch in expected and actual field dictionaries (deep)")
	}
}
