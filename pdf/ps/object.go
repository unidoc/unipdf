/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

import (
	"fmt"
)

// PSObject represents a postscript object.
type PSObject interface {
	// Duplicate makes a fresh copy of the PSObject.
	Duplicate() PSObject

	// DebugString returns a descriptive representation of the PSObject with more information than String()
	// for debugging purposes.
	DebugString() string

	// String returns a string representation of the PSObject.
	String() string
}

// PSInteger represents an integer.
type PSInteger struct {
	Val int
}

func (int *PSInteger) Duplicate() PSObject {
	obj := PSInteger{}
	obj.Val = int.Val
	return &obj
}

func (int *PSInteger) DebugString() string {
	return fmt.Sprintf("int:%d", int.Val)
}

func (int *PSInteger) String() string {
	return fmt.Sprintf("%d", int.Val)
}

// PSReal represents a real number.
type PSReal struct {
	Val float64
}

func (real *PSReal) DebugString() string {
	return fmt.Sprintf("real:%.5f", real.Val)
}

func (real *PSReal) String() string {
	return fmt.Sprintf("%.5f", real.Val)
}

func (real *PSReal) Duplicate() PSObject {
	obj := PSReal{}
	obj.Val = real.Val
	return &obj
}

// PSBoolean represents a boolean value.
type PSBoolean struct {
	Val bool
}

func (bool *PSBoolean) DebugString() string {
	return fmt.Sprintf("bool:%v", bool.Val)
}

func (bool *PSBoolean) String() string {
	return fmt.Sprintf("%v", bool.Val)
}

func (bool *PSBoolean) Duplicate() PSObject {
	obj := PSBoolean{}
	obj.Val = bool.Val
	return &obj
}

// PSProgram defines a Postscript program which is a series of PS objects (arguments, commands, programs etc).
type PSProgram []PSObject

// NewPSProgram returns an empty, initialized PSProgram.
func NewPSProgram() *PSProgram {
	return &PSProgram{}
}

// Append appends an object to the PSProgram.
func (prog *PSProgram) Append(obj PSObject) {
	*prog = append(*prog, obj)
}

func (prog *PSProgram) DebugString() string {
	s := "{ "
	for _, obj := range *prog {
		s += obj.DebugString()
		s += " "
	}
	s += "}"

	return s
}

func (prog *PSProgram) String() string {
	s := "{ "
	for _, obj := range *prog {
		s += obj.String()
		s += " "
	}
	s += "}"

	return s
}

func (prog *PSProgram) Duplicate() PSObject {
	prog2 := &PSProgram{}
	for _, obj := range *prog {
		prog2.Append(obj.Duplicate())
	}
	return prog2
}

// Exec executes the program, typically leaving output values on the stack.
func (prog *PSProgram) Exec(stack *PSStack) error {
	for _, obj := range *prog {
		var err error
		switch t := obj.(type) {
		case *PSInteger:
			number := t
			err = stack.Push(number)
		case *PSReal:
			number := t
			err = stack.Push(number)
		case *PSBoolean:
			val := t
			err = stack.Push(val)
		case *PSProgram:
			function := t
			err = stack.Push(function)
		case *PSOperand:
			op := t
			err = op.Exec(stack)
		default:
			return ErrTypeCheck
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// PSOperand represents a Postscript operand (text string).
type PSOperand string

func (op *PSOperand) DebugString() string {
	return fmt.Sprintf("op:'%s'", *op)
}

func (op *PSOperand) String() string {
	return fmt.Sprintf("%s", *op)
}

func (op *PSOperand) Duplicate() PSObject {
	s := *op
	return &s
}

// Exec executes the operand `op` in the state specified by `stack`.
func (op *PSOperand) Exec(stack *PSStack) error {
	err := ErrUnsupportedOperand
	switch *op {
	case "abs":
		err = op.abs(stack)
	case "add":
		err = op.add(stack)
	case "and":
		err = op.and(stack)
	case "atan":
		err = op.atan(stack)

	case "bitshift":
		err = op.bitshift(stack)

	case "ceiling":
		err = op.ceiling(stack)
	case "copy":
		err = op.copy(stack)
	case "cos":
		err = op.cos(stack)
	case "cvi":
		err = op.cvi(stack)
	case "cvr":
		err = op.cvr(stack)

	case "div":
		err = op.div(stack)
	case "dup":
		err = op.dup(stack)

	case "eq":
		err = op.eq(stack)
	case "exch":
		err = op.exch(stack)
	case "exp":
		err = op.exp(stack)

	case "floor":
		err = op.floor(stack)

	case "ge":
		err = op.ge(stack)
	case "gt":
		err = op.gt(stack)

	case "idiv":
		err = op.idiv(stack)
	case "if":
		err = op.ifCondition(stack)
	case "ifelse":
		err = op.ifelse(stack)
	case "index":
		err = op.index(stack)

	case "le":
		err = op.le(stack)
	case "log":
		err = op.log(stack)
	case "ln":
		err = op.ln(stack)
	case "lt":
		err = op.lt(stack)

	case "mod":
		err = op.mod(stack)
	case "mul":
		err = op.mul(stack)

	case "ne":
		err = op.ne(stack)
	case "neg":
		err = op.neg(stack)
	case "not":
		err = op.not(stack)

	case "or":
		err = op.or(stack)

	case "pop":
		err = op.pop(stack)

	case "round":
		err = op.round(stack)
	case "roll":
		err = op.roll(stack)

	case "sin":
		err = op.sin(stack)
	case "sqrt":
		err = op.sqrt(stack)
	case "sub":
		err = op.sub(stack)

	case "truncate":
		err = op.truncate(stack)

	case "xor":
		err = op.xor(stack)
	}

	return err
}
