/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/testutils"
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
	require.NoError(t, err)

	// Load the field from object number 1.
	obj, err := r.parser.LookupByNumber(1)
	require.NoError(t, err)

	ind, ok := obj.(*core.PdfIndirectObject)
	require.True(t, ok)

	field, err := r.newPdfFieldFromIndirectObject(ind, nil)
	require.NoError(t, err)

	// Check properties of the field.
	buttonf, ok := field.GetContext().(*PdfFieldButton)
	require.True(t, ok)
	require.NotNil(t, buttonf)

	if len(field.Kids) > 0 {
		t.Fatalf("Field should not have kids")
	}
	require.Len(t, field.Annotations, 1)

	// Field -> PDF object.  Regenerate the field dictionary and see if matches expectations.
	// Reset the dictionaries for both field and annotation to avoid re-use during re-generation of PDF object.
	field.container = core.MakeIndirectObject(core.MakeDict())
	field.Annotations[0].container = core.MakeIndirectObject(core.MakeDict())
	fieldPdfObj := field.ToPdfObject()
	fieldDict, ok := fieldPdfObj.(*core.PdfIndirectObject).PdfObject.(*core.PdfObjectDictionary)
	require.True(t, ok)

	// Load the expected field dictionary (output).  Slightly different than original as the input had
	// a merged-in annotation. Our output does not currently merge annotations.
	obj, err = r.parser.LookupByNumber(4)
	require.NoError(t, err)

	expDict, ok := obj.(*core.PdfIndirectObject).PdfObject.(*core.PdfObjectDictionary)
	require.True(t, ok)

	if !testutils.CompareDictionariesDeep(expDict, fieldDict) {
		t.Fatalf("Mismatch in expected and actual field dictionaries (deep)")
	}
}

func TestFormNil(t *testing.T) {
	var form *PdfAcroForm
	err := form.Fill(nil)
	require.NoError(t, err)
}

// TODO: Test loading and writing out of merged-in annotations.
func TestReadWriteMergedFieldAnnotation(t *testing.T) {
	raw := `
6 0 obj
<<
/Type /Annot
/Rect [0 0 0 0]
/P 99 0 R
/F 132
/Subtype /Widget
/T (Signature 1)
/FT /Sig
/V 7 0 R
>>
endobj
`
	_ = raw
	t.Skip("Not implemented yet")
}
