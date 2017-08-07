/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

// A limited postscript parser for PDF function type 4.

import (
	"fmt"

	"github.com/unidoc/unidoc/common"
)

// A PSExecutor has its own execution stack and is used to executre a PS routine (program).
type PSExecutor struct {
	Stack   *PSStack
	program *PSProgram
}

func NewPSExecutor(program *PSProgram) *PSExecutor {
	executor := &PSExecutor{}
	executor.Stack = NewPSStack()
	executor.program = program
	return executor
}

func PSObjectArrayToFloat64Array(objects []PSObject) ([]float64, error) {
	vals := []float64{}

	for _, obj := range objects {
		if number, is := obj.(*PSInteger); is {
			vals = append(vals, float64(number.Val))
		} else if number, is := obj.(*PSReal); is {
			vals = append(vals, number.Val)
		} else {
			return nil, fmt.Errorf("Type error")
		}
	}

	return vals, nil
}

func (this *PSExecutor) Execute(objects []PSObject) ([]PSObject, error) {
	// Add the arguments on stack
	// [obj1 obj2 ...]
	for _, obj := range objects {
		err := this.Stack.Push(obj)
		if err != nil {
			return nil, err
		}
	}

	err := this.program.Exec(this.Stack)
	if err != nil {
		common.Log.Debug("Exec failed: %v", err)
		return nil, err
	}

	result := []PSObject(*this.Stack)
	this.Stack.Empty()

	return result, nil
}
