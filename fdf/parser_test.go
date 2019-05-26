/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fdf

import (
	"fmt"
	"testing"

	"github.com/unidoc/unipdf/v3/core"
)

const fdfExample1 = `
%FDF-1.4
%âãÏÓ
1 0 obj
<</FDF<</Fields[<</T(Field1)/V(Test1)>><</T(Field2)/V(Test2)>>]>>>>
endobj
trailer
<</Root 1 0 R>>
%%EOF
`

func TestFdfExample1(t *testing.T) {
	fdfDoc, err := newParserFromString(fdfExample1)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	fdfDict, err := fdfDoc.Root()
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	fmt.Printf("P: %v\n", fdfDoc)
	fmt.Printf("FDF: %v\n", fdfDict)
	fmt.Printf("Keys: %v\n", fdfDict.Keys())

	fields, ok := fdfDict.Get("Fields").(*core.PdfObjectArray)
	if !ok {
		t.Errorf("Incorrect type (%T)", fdfDict.Get("Fields"))
		return
	}

	expectedFields := 2
	expectedT := []string{"Field1", "Field2"}
	expectedV := []string{"Test1", "Test2"}

	if fields.Len() != expectedFields {
		t.Errorf("Incorrect number of fields (got %d)", fields.Len())
		return
	}
	for i := 0; i < expectedFields; i++ {
		fd, ok := fields.Get(i).(*core.PdfObjectDictionary)
		if !ok {
			t.Errorf("Incorrect field type")
			return
		}

		ts, ok := fd.Get("T").(*core.PdfObjectString)
		if !ok {
			t.Errorf("Type error")
			return
		}
		if ts.Str() != expectedT[i] {
			t.Errorf("Incorrect value")
			return
		}

		vs, ok := fd.Get("V").(*core.PdfObjectString)
		if !ok {
			t.Errorf("Type error")
			return
		}
		if vs.Str() != expectedV[i] {
			t.Errorf("Incorrect value")
			return
		}
	}

	for i, k := range fields.Elements() {
		fmt.Printf("Field %v : %v\n", i, k)
	}

}
