/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/unidoc/unipdf/v3/common"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
}

func quickEval(progText string) (PSObject, error) {
	parser := NewPSParser([]byte(progText))

	prog, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s\n", progText)
	fmt.Printf("-> Program: %s\n", prog.DebugString())

	exec := NewPSExecutor(prog)

	outputs, err := exec.Execute(nil)
	if err != nil {
		return nil, err
	}

	if len(outputs) != 1 {
		return nil, errors.New("stack result has too many values (>1)")
	}

	stack := PSStack(outputs)
	fmt.Printf("=> Result Stack: %s\n", stack.DebugString())

	return outputs[0], nil
}

func quickTest(progText string) (*PSStack, error) {
	parser := NewPSParser([]byte(progText))

	prog, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", progText)
	fmt.Printf("-> Program: %s\n", prog.DebugString())

	exec := NewPSExecutor(prog)

	outputs, err := exec.Execute(nil)
	if err != nil {
		return nil, err
	}
	stack := PSStack(outputs)

	return &stack, nil
}

func TestAdd1(t *testing.T) {
	progText := "{ 1 1 add }"

	obj, err := quickEval(progText)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	val, ok := obj.(*PSInteger)
	if !ok {
		t.Errorf("Wrong output type")
		return
	}

	if val.Val != 2 {
		t.Errorf("Wrong result")
		return
	}
}

func TestAdd2(t *testing.T) {
	progText := "{ 1.1 1 add 3 4 add add }"

	obj, err := quickEval(progText)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	val, ok := obj.(*PSReal)
	if !ok {
		t.Errorf("Wrong output type")
		return
	}
	if math.Abs(val.Val-9.1) > tolerance {
		t.Errorf("Wrong result")
		return
	}
}

//// 8.3 6.6 sub -> 1.7 (real)
// 8 6.3 sub -> 1.7 (real)
// 8 6 sub -> 2 (int)
func TestSub1(t *testing.T) {
	progText := "{ 8.3 6.6 sub }"

	obj, err := quickEval(progText)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	val, ok := obj.(*PSReal)
	if !ok {
		t.Errorf("Wrong output type")
		return
	}
	if math.Abs(val.Val-1.7) > tolerance {
		t.Errorf("Wrong result")
		return
	}
}

func TestSub2(t *testing.T) {
	progText := "{ 8 6.3 sub }"

	obj, err := quickEval(progText)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	val, ok := obj.(*PSReal)
	if !ok {
		t.Errorf("Wrong output type")
		return
	}
	if math.Abs(val.Val-1.7) > tolerance {
		t.Errorf("Wrong result")
		return
	}
}

func TestSub3(t *testing.T) {
	progText := "{ 8 6 sub }"

	obj, err := quickEval(progText)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	val, ok := obj.(*PSInteger)
	if !ok {
		t.Errorf("Wrong output type")
		return
	}
	if val.Val != 2 {
		t.Errorf("Wrong result")
		return
	}
}

// 6 + (3/8) -> 6.375
// 3 8 div 6 add
// 6 3 8 div add
//
// 8 - (7*3) -> -13
// 8 7 3 mul sub
// 7 3 mul 8 exch sub
// Simple test entry with a single expected PSObject output.
type SimpleTestEntry struct {
	progText string
	expected PSObject
}

func TestArithmetics(t *testing.T) {
	testcases := []SimpleTestEntry{
		{progText: "{ 3 8 div 6 add }", expected: MakeReal(6.375)},
		{progText: "{ 6 3 8 div add }", expected: MakeReal(6.375)},
		{progText: "{ 8 7 3 mul sub }", expected: MakeInteger(-13)},
		{progText: "{ 7 3 mul 8 exch sub }", expected: MakeInteger(-13)},
	}

	for _, testcase := range testcases {
		obj, err := quickEval(testcase.progText)
		if err != nil {
			t.Errorf("Error: %v", err)
			return
		}

		// Maybe not the most robust test (comparing the strings), but should do.
		if obj.DebugString() != testcase.expected.DebugString() {
			t.Errorf("Wrong result: %s != %s", obj.DebugString(), testcase.expected.DebugString())
			return
		}
	}
}

// Complex test entry can have a more complex output.
type ComplexTestEntry struct {
	progText string
	expected string
}

func TestStackOperations(t *testing.T) {
	testcases := []ComplexTestEntry{
		{progText: "{ 7 8 9 3 1 roll }", expected: "[ int:9 int:7 int:8 ]"},
		{progText: "{ 7 8 9 3 -1 roll }", expected: "[ int:8 int:9 int:7 ]"},
		{progText: "{ 9 7 8 3 -1 roll }", expected: "[ int:7 int:8 int:9 ]"},
		{progText: "{ 1 1 0.2 7 8 9 3 1 roll }", expected: "[ int:1 int:1 real:0.20000 int:9 int:7 int:8 ]"},
	}

	for _, testcase := range testcases {
		stack, err := quickTest(testcase.progText)
		if err != nil {
			t.Errorf("Error: %v", err)
			return
		}

		// Maybe not the most robust test (comparing the strings), but should do.
		if stack.DebugString() != testcase.expected {
			t.Errorf("Wrong result: '%s' != '%s'", stack.DebugString(), testcase.expected)
			return
		}
	}
}

func TestFunctionOperations(t *testing.T) {
	testcases := []ComplexTestEntry{
		// atan
		{progText: "{ 0 1 atan }", expected: "[ real:0.00000 ]"},
		{progText: "{ 1 0 atan }", expected: "[ real:90.00000 ]"},
		{progText: "{ -100 0 atan }", expected: "[ real:270.00000 ]"},
		{progText: "{ 4 4 atan }", expected: "[ real:45.00000 ]"},
	}

	for _, testcase := range testcases {
		stack, err := quickTest(testcase.progText)
		if err != nil {
			t.Errorf("Error: %v", err)
			return
		}

		// Maybe not the most robust test (comparing the strings), but should do.
		if stack.DebugString() != testcase.expected {
			t.Errorf("Wrong result: '%s' != '%s'", stack.DebugString(), testcase.expected)
			return
		}
	}
}

func TestVariousCases(t *testing.T) {
	testcases := []ComplexTestEntry{
		// dup
		{progText: "{ 99 dup }", expected: "[ int:99 int:99 ]"},
		// ceiling
		{progText: "{ 3.2 ceiling }", expected: "[ real:4.00000 ]"},
		{progText: "{ -4.8 ceiling }", expected: "[ real:-4.00000 ]"},
		{progText: "{ 99 ceiling }", expected: "[ int:99 ]"},
		// floor
		{progText: "{ 3.2 floor }", expected: "[ real:3.00000 ]"},
		{progText: "{ -4.8 floor }", expected: "[ real:-5.00000 ]"},
		{progText: "{ 99 floor }", expected: "[ int:99 ]"},
		// exp
		{progText: "{ 9 0.5 exp }", expected: "[ real:3.00000 ]"},
		{progText: "{ -9 -1 exp }", expected: "[ real:-0.11111 ]"},
		// and
		{progText: "{ true true and }", expected: "[ bool:true ]"},
		{progText: "{ true false and }", expected: "[ bool:false ]"},
		{progText: "{ false true and }", expected: "[ bool:false ]"},
		{progText: "{ false false and }", expected: "[ bool:false ]"},
		{progText: "{ 99 1 and }", expected: "[ int:1 ]"},
		{progText: "{ 52 7 and }", expected: "[ int:4 ]"},
		// bitshift
		{progText: "{ 7 3 bitshift }", expected: "[ int:56 ]"},
		{progText: "{ 142 -3 bitshift }", expected: "[ int:17 ]"},
		// copy
		{progText: "{ 7 3 2 copy }", expected: "[ int:7 int:3 int:7 int:3 ]"},
		{progText: "{ 7 3 0 copy }", expected: "[ int:7 int:3 ]"},
		// cos
		{progText: "{ 0 cos }", expected: "[ real:1.00000 ]"},
		{progText: "{ 90 cos }", expected: "[ real:0.00000 ]"},
		// eq.
		{progText: "{ 4.0 4 eq }", expected: "[ bool:true ]"},
		{progText: "{ 4 4.0 eq }", expected: "[ bool:true ]"},
		{progText: "{ 4.0 4.0 eq }", expected: "[ bool:true ]"},
		{progText: "{ 4 4 eq }", expected: "[ bool:true ]"},
		{progText: "{ -4 4 eq }", expected: "[ bool:false ]"},
		{progText: "{ false false eq }", expected: "[ bool:true ]"},
		{progText: "{ true false eq }", expected: "[ bool:false ]"},
		{progText: "{ true 4 eq }", expected: "[ bool:false ]"},
		// ge
		{progText: "{ 4.2 4 ge }", expected: "[ bool:true ]"},
		{progText: "{ 4 4 ge }", expected: "[ bool:true ]"},
		{progText: "{ 3.9 4 ge }", expected: "[ bool:false ]"},
		// gt
		{progText: "{ 4.2 4 gt }", expected: "[ bool:true ]"},
		{progText: "{ 4 4 gt }", expected: "[ bool:false ]"},
		{progText: "{ 3.9 4 gt }", expected: "[ bool:false ]"},
		// if
		{progText: "{ 4.2 4 gt {5} if }", expected: "[ int:5 ]"},
		{progText: "{ 4.2 4 gt {4.0 4.0 ge {3} if} if}", expected: "[ int:3 ]"},
		{progText: "{ 4.0 4.0 gt {5} if }", expected: "[ ]"},
		// ifelse
		{progText: "{ 4.2 4 gt {5} {4} ifelse }", expected: "[ int:5 ]"},
		{progText: "{ 3 4 gt {5} {4} ifelse }", expected: "[ int:4 ]"},
		// index
		{progText: "{ 0 1 2 3 4 5 2 index }", expected: "[ int:0 int:1 int:2 int:3 int:4 int:5 int:3 ]"},
		{progText: "{ 9 8 7 2 index }", expected: "[ int:9 int:8 int:7 int:9 ]"},
		// le
		{progText: "{ 4.2 4 le }", expected: "[ bool:false ]"},
		{progText: "{ 4 4 le }", expected: "[ bool:true ]"},
		{progText: "{ 3.9 4 le }", expected: "[ bool:true ]"},
		// ln
		{progText: "{ 10 ln }", expected: "[ real:2.30259 ]"},
		{progText: "{ 100 ln }", expected: "[ real:4.60517 ]"},
		// log
		{progText: "{ 10 log }", expected: "[ real:1.00000 ]"},
		{progText: "{ 100 log }", expected: "[ real:2.00000 ]"},
		// lt
		{progText: "{ 4.2 4 lt }", expected: "[ bool:false ]"},
		{progText: "{ 4 4 lt }", expected: "[ bool:false ]"},
		{progText: "{ 3.9 4 lt }", expected: "[ bool:true ]"},
		// ne
		{progText: "{ 4.0 4 ne }", expected: "[ bool:false ]"},
		{progText: "{ 4 4.0 ne }", expected: "[ bool:false ]"},
		{progText: "{ 4.0 4.0 ne }", expected: "[ bool:false ]"},
		{progText: "{ 4 4 ne }", expected: "[ bool:false ]"},
		{progText: "{ -4 4 ne }", expected: "[ bool:true ]"},
		{progText: "{ false false ne }", expected: "[ bool:false ]"},
		{progText: "{ true false ne }", expected: "[ bool:true ]"},
		{progText: "{ true 4 ne }", expected: "[ bool:true ]"},
		// neg
		// not
		{progText: "{ true not }", expected: "[ bool:false ]"},
		{progText: "{ false not }", expected: "[ bool:true ]"},
		{progText: "{ 52 not }", expected: "[ int:-53 ]"},
		// or
		{progText: "{ true true or }", expected: "[ bool:true ]"},
		{progText: "{ true false or }", expected: "[ bool:true ]"},
		{progText: "{ false true or }", expected: "[ bool:true ]"},
		{progText: "{ false false or }", expected: "[ bool:false ]"},
		{progText: "{ 17 5 or }", expected: "[ int:21 ]"},
		// pop
		{progText: "{ 1 2 3 pop }", expected: "[ int:1 int:2 ]"},
		{progText: "{ 1 2 pop }", expected: "[ int:1 ]"},
		{progText: "{ 1 pop }", expected: "[ ]"},
		// round
		{progText: "{ 3.2 round }", expected: "[ real:3.00000 ]"},
		{progText: "{ 6.5 round }", expected: "[ real:7.00000 ]"},
		{progText: "{ -4.8 round }", expected: "[ real:-5.00000 ]"},
		{progText: "{ -6.5 round }", expected: "[ real:-6.00000 ]"},
		{progText: "{ 99 round }", expected: "[ int:99 ]"},
		// roll
		{progText: "{ 1 2 3 3 -1 roll }", expected: "[ int:2 int:3 int:1 ]"},
		{progText: "{ 1 2 3 3 1 roll }", expected: "[ int:3 int:1 int:2 ]"},
		{progText: "{ 1 2 3 3 0 roll }", expected: "[ int:1 int:2 int:3 ]"},
		// sin
		{progText: "{ 0 sin }", expected: "[ real:0.00000 ]"},
		{progText: "{ 90 sin }", expected: "[ real:1.00000 ]"},
		// sqrt
		{progText: "{ 4 sqrt }", expected: "[ real:2.00000 ]"},
		{progText: "{ 2 sqrt }", expected: "[ real:1.41421 ]"},
		// truncate
		{progText: "{ 3.2 truncate }", expected: "[ real:3.00000 ]"},
		{progText: "{ -4.8 truncate }", expected: "[ real:-4.00000 ]"},
		{progText: "{ 99 truncate }", expected: "[ int:99 ]"},
		// xor
		{progText: "{ true true xor }", expected: "[ bool:false ]"},
		{progText: "{ true false xor }", expected: "[ bool:true ]"},
		{progText: "{ false true xor }", expected: "[ bool:true ]"},
		{progText: "{ false false xor }", expected: "[ bool:false ]"},
		{progText: "{ 7 3 xor }", expected: "[ int:4 ]"},
		{progText: "{ 12 3 xor }", expected: "[ int:15 ]"},
	}

	for _, testcase := range testcases {
		stack, err := quickTest(testcase.progText)
		if err != nil {
			t.Errorf("Error: %v", err)
			return
		}

		// Maybe not the most robust test (comparing the strings), but should do.
		if stack.DebugString() != testcase.expected {
			t.Errorf("Wrong result: '%s' != '%s'", stack.DebugString(), testcase.expected)
			return
		}
	}
}

func TestTintTransform1(t *testing.T) {
	testcases := []ComplexTestEntry{
		// from corpus epson_pages3_color_pages1.pdf.
		{progText: "{ 0.0000 dup 0 mul exch dup 0 mul exch dup 0 mul exch 1 mul }", expected: "[ real:0.00000 real:0.00000 real:0.00000 real:0.00000 ]"},
	}

	for _, testcase := range testcases {
		stack, err := quickTest(testcase.progText)
		if err != nil {
			t.Errorf("Error: %v", err)
			return
		}

		// Maybe not the most robust test (comparing the strings), but should do.
		if stack.DebugString() != testcase.expected {
			t.Errorf("Wrong result: '%s' != '%s'", stack.DebugString(), testcase.expected)
			return
		}
	}
}
