/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/testutils"
)

func TestSeparationCS1(t *testing.T) {
	rawObject := `
% Colour space
5 0 obj
[ /Separation /LogoGreen /DeviceCMYK 12 0 R ]
endobj
% Tint transformation function
12 0 obj
<<
	/FunctionType 4
	/Domain [0.0 1.0]
	/Range [ 0.0 1.0 0.0 1.0 0.0 1.0 0.0 1.0 ]
	/Length 65
>>
stream
{ dup 0.84 mul
exch 0.00 exch dup 0.44 mul exch 0.21 mul
}
endstream endobj
	`

	// Test a few lookups and see if it is accurate.
	// Test rgb conversion for a few specific values also.

	fmt.Println(rawObject)

	//t.Errorf("Test not implemented yet")
}

func TestDeviceNCS1(t *testing.T) {
	// Implement Example 3 on p. 172

	//t.Errorf("Test not implemented yet")
}

// Bug with outputing Separation colorspaces using a Function0 function.
func TestColorspaceLoading(t *testing.T) {
	rawpdf := []byte(`
10 0 obj
<<
	/ColorSpace <<
        /R99 99 0 R
		/R86 86 0 R
		/R88 88 0 R
	>>
>>
endobj
99 0 obj
[/Separation /All 86 0 R 87 0 R]
endobj
86 0 obj
[/ICCBased 85 0 R]
endobj
88 0 obj
[/Separation /Black 86 0 R 87 0 R]
endobj
85 0 obj
<<
	/N 3
	/Alternate /DeviceRGB
	/Length 3144
>>
stream
!!STREAMDATA2!!
endstream endobj
87 0 obj
<<
	/Size [255]
	/BitsPerSample 8
	/Domain [0 1]
	/Length 765
	/Encode [0 254]
	/Decode [0 1 0 1 0 1]
	/Range [0 1 0 1 0 1]
	/FunctionType 0
>>
stream
!!STREAMDATA1!!
endstream endobj
`)
	functype0, err := ioutil.ReadFile(`testdata/cs_functype0.bin`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	iccdata, err := ioutil.ReadFile(`testdata/iccstream.bin`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	rawpdf = bytes.Replace(rawpdf, []byte("!!STREAMDATA1!!"), functype0, 1)
	rawpdf = bytes.Replace(rawpdf, []byte("!!STREAMDATA2!!"), iccdata, 1)
	//fmt.Printf("%s\n", rawpdf)

	objMap, err := testutils.ParseIndirectObjects(string(rawpdf))
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(objMap) != 6 {
		t.Fatalf("len(objMap) != 6 (%d)", len(objMap))
	}

	resourceDict, ok := core.GetDict(objMap[10])
	if !ok {
		t.Fatalf("Resource object missing dictionary")
	}

	resources, err := NewPdfPageResourcesFromDict(resourceDict)
	if err != nil {
		t.Fatalf("Error loading resources: %v", err)
	}

	fmt.Printf("Out\n")
	fmt.Printf("%s\n", resources.ToPdfObject())
	outDict, ok := core.GetDict(resources.ToPdfObject())
	if !ok {
		t.Fatalf("error")
	}
	fmt.Printf("%s\n", outDict.WriteString())
	r99, ok := resources.GetColorspaceByName("R99")
	if !ok {
		t.Fatalf("error")
	}

	array, ok := core.GetArray(r99.ToPdfObject())
	if !ok {
		t.Fatalf("error")
	}

	if array.Len() != 4 {
		t.Fatalf("len != 4 (got %d)", array.Len())
	}

	name, ok := core.GetName(array.Get(0))
	if !ok || *name != "Separation" {
		t.Fatalf("Unexpected value")
	}

	name, ok = core.GetName(array.Get(1))
	if !ok || *name != "All" {
		t.Fatalf("Unexpected value")
	}

	ind, ok := core.GetIndirect(array.Get(2))
	if !ok {
		t.Fatalf("error")
	}
	if ind.ObjectNumber != 86 {
		t.Fatalf("Incorrect obj num")
	}

	f, ok := core.GetStream(array.Get(3))
	if !ok {
		t.Fatalf("Stream get error")
	}
	if f.ObjectNumber != 87 {
		t.Fatalf("Incorrect function obj number (got %d)", f.ObjectNumber)
	}
}
