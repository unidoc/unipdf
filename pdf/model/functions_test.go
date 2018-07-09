// Test functions
// XXX: We need tests of type 0, type 2, type 3, type 4 functions.  Particularly type 0 is complex and
// needs comprehensive tests.

package model

import (
	"fmt"
	"math"
	"testing"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

func init() {
	//common.SetLogger(common.ConsoleLogger{})
	common.SetLogger(common.DummyLogger{})
}

type Type4TestCase struct {
	Inputs   []float64
	Expected []float64
}

// TODO/XXX: Implement example 2 from page 167.

func TestType4Function1(t *testing.T) {
	rawText := `
10 0 obj
<<
	/FunctionType 4
	/Domain [ -1.0 1.0 -1.0 1.0]
	/Range [ -1.0 1.0 ]
	/Length 48
>>
stream
{ 360 mul sin
2 div
exch 360 mul
sin 2 div
add
} endstream
endobj
`
	/*
	 * Inputs: [x y] where -1<x<1, -1<y<1
	 * Outputs: z where -1 < z < 1
	 * z = sin(y * 360)/2 + sin(x * 360)/2
	 */

	parser := NewParserFromString(rawText)

	obj, err := parser.ParseIndirectObject()
	if err != nil {
		t.Fatalf("Failed to parse indirect obj (%s)", err)
	}

	_, ok := obj.(*PdfObjectStream)
	if !ok {
		t.Fatalf("Invalid object type (%q)", obj)
	}

	fun, err := newPdfFunctionFromPdfObject(obj)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// z = sin(360*x)/2 + sin(360*y)/2
	testcases := []Type4TestCase{
		{[]float64{0.5, 0.5}, []float64{0}},
		{[]float64{0.25, 0.25}, []float64{1.0}},
		{[]float64{0.25, 0.5}, []float64{0.5}},
		{[]float64{0.5, 0.25}, []float64{0.5}},
		{[]float64{-0.5, -0.25}, []float64{-0.50}},
	}

	for _, testcase := range testcases {
		outputs, err := fun.Evaluate(testcase.Inputs)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}
		fmt.Println(testcase)
		fmt.Println(outputs)

		if len(outputs) != len(testcase.Expected) {
			t.Fatalf("Failed, output length mismatch")
		}
		for i := 0; i < len(outputs); i++ {
			if math.Abs(outputs[i]-testcase.Expected[i]) > 0.000001 {
				t.Fatalf("Failed, output and expected mismatch")
			}
		}
	}
}
