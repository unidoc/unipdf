/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package ps implements various functionalities needed for handling Postscript for PDF uses, in particular
// for PDF function type 4.
package ps

import (
	"github.com/unidoc/unipdf/v3/common"
)

// PSExecutor has its own execution stack and is used to executre a PS routine (program).
type PSExecutor struct {
	Stack   *PSStack
	program *PSProgram
}

// NewPSExecutor returns an initialized PSExecutor for an input `program`.
func NewPSExecutor(program *PSProgram) *PSExecutor {
	executor := &PSExecutor{}
	executor.Stack = NewPSStack()
	executor.program = program
	return executor
}

// PSObjectArrayToFloat64Array converts []PSObject into a []float64 array. Each PSObject must represent a number,
// otherwise a ErrTypeCheck error occurs.
func PSObjectArrayToFloat64Array(objects []PSObject) ([]float64, error) {
	var vals []float64

	for _, obj := range objects {
		if number, is := obj.(*PSInteger); is {
			vals = append(vals, float64(number.Val))
		} else if number, is := obj.(*PSReal); is {
			vals = append(vals, number.Val)
		} else {
			return nil, ErrTypeCheck
		}
	}

	return vals, nil
}

// Execute executes the program for an input parameters `objects` and returns a slice of output objects.
func (exec *PSExecutor) Execute(objects []PSObject) ([]PSObject, error) {
	// Add the arguments on stack
	// [obj1 obj2 ...]
	for _, obj := range objects {
		err := exec.Stack.Push(obj)
		if err != nil {
			return nil, err
		}
	}

	err := exec.program.Exec(exec.Stack)
	if err != nil {
		common.Log.Debug("Exec failed: %v", err)
		return nil, err
	}

	result := []PSObject(*exec.Stack)
	exec.Stack.Empty()

	return result, nil
}
