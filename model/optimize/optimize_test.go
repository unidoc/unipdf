/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model/optimize"
)

// parseIndirectObjects parses a sequence of indirect/stream objects sequentially from a `rawpdf` text.
func parseIndirectObjects(rawpdf string) ([]core.PdfObject, error) {
	p := core.NewParserFromString(rawpdf)
	var indirects []core.PdfObject
	for {
		obj, err := p.ParseIndirectObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		indirects = append(indirects, obj)
	}

	return indirects, nil
}

// debugObjects prints objects in a readable fashion, convenient when debugging.
func debugObjects(objects []core.PdfObject) string {
	var buf bytes.Buffer

	for _, obj := range objects {
		switch t := obj.(type) {
		case *core.PdfIndirectObject:
			buf.WriteString(fmt.Sprintf("%d 0 obj\n", t.ObjectNumber))
			buf.WriteString(fmt.Sprintf("  %s\n", t.PdfObject.String()))
		}
	}
	return buf.String()
}

func TestOptimizeIdenticalIndirects1(t *testing.T) {
	rawpdf := `
1 0 obj
<<
  /Name (1234)
>>
endobj
2 0 obj
<< /Name (1234) >>
endobj
`
	objects, err := parseIndirectObjects(rawpdf)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if len(objects) != 2 {
		t.Fatalf("len(objects) != 2 (%d)", len(objects))
	}

	// Combine duplicate direct objects - Expect unchanged results.
	{
		opt := optimize.CombineDuplicateDirectObjects{}
		optObjects, err := opt.Optimize(objects)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(optObjects) != 2 {
			t.Fatalf("len(optObjects1) != 2 (%d)", len(optObjects))
		}
	}

	// Combine indirect objects should go from 2 to 1.
	{
		opt := optimize.CombineIdenticalIndirectObjects{}
		optObjects, err := opt.Optimize(objects)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(optObjects) != 1 {
			t.Fatalf("len(optObjects1) != 1 (%d)", len(optObjects))
		}
	}
}

// More complex case, where has a reference, where as the other does not.
// Expecting this NOT to work as we don't currently support this case.
// TODO: Add support for this.
func TestOptimizeIdenticalIndirectsUnsupported1(t *testing.T) {
	rawpdf := `
1 0 obj
(1234)
endobj
2 0 obj
<<
  /Name (1234)
>>
endobj
3 0 obj
<< /Name 1 0 R >>
endobj
`
	objects, err := parseIndirectObjects(rawpdf)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if len(objects) != 3 {
		t.Fatalf("len(objects) != 2 (%d)", len(objects))
	}

	// Combine duplicate direct objects - Expect unchanged results.
	{
		opt := optimize.CombineDuplicateDirectObjects{}
		optObjects, err := opt.Optimize(objects)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(optObjects) != 3 {
			t.Fatalf("len(optObjects1) != 2 (%d)", len(optObjects))
		}
	}

	// Combine indirect objects should go from 3 to 2.
	{
		opt := optimize.CombineIdenticalIndirectObjects{}
		optObjects, err := opt.Optimize(objects)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(optObjects) != 3 { // TODO: Add support. IF IDEAL: would be 2.
			t.Fatalf("len(optObjects1) != 2 (%d)", len(optObjects))
		}
	}
}

// Showcases problem with sequence of CombineDuplicateDirectObjects followed by CombineIdenticalIndirectObjects
// if object numbers are not updated between steps (due to non-unique object numbering and reference strings).
func TestOptimizationSequence1(t *testing.T) {
	rawpdf := `
1 0 obj
<<
  /Inner << /Color (red) >>
>>
endobj
2 0 obj
<<
 /Inner << /Color (red) >>
 /Other (abc)
>>
endobj
3 0 obj
<<
 /Inner << /Color (blue) >>
 /Other (abc)
>>
endobj
4 0 obj
<<
 /Inner << /Color (blue) >>
>>
endobj
`
	objects, err := parseIndirectObjects(rawpdf)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(objects) != 4 {
		t.Fatalf("len(objects) != 4 (%d)", len(objects))
	}
	debugstr1 := debugObjects(objects)

	// 1. Combine duplicate direct objects.
	// Expect that 2 new indirect objects will be added, as two of the inner dictionaries are identical.
	opt := optimize.CombineDuplicateDirectObjects{}
	optObjects, err := opt.Optimize(objects)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(optObjects) != 6 {
		t.Fatalf("len(optObjects) != 6 (%d)", len(optObjects))
	}
	debugstr2 := debugObjects(optObjects)

	// 2. Combine indirect objects.
	// Should not make any difference here unless there was a problem.
	opt2 := optimize.CombineIdenticalIndirectObjects{}
	optObjects, err = opt2.Optimize(optObjects)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	debugstr3 := debugObjects(optObjects)
	fmt.Println("==Original")
	fmt.Println(debugstr1)
	fmt.Println("==After CombineDuplicateDirectObjects")
	fmt.Println(debugstr2)
	fmt.Println("==After CombineIdenticalIndirectObjects")
	fmt.Println(debugstr3)
	if len(optObjects) != 6 {
		t.Fatalf("len(optObjects) != 6 (%d)", len(optObjects))
	}
}
