/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

// PSStack defines a stack of PSObjects. PSObjects can be pushed on or pull from the stack.
type PSStack []PSObject

// NewPSStack returns an initialized PSStack.
func NewPSStack() *PSStack {
	return &PSStack{}
}

// Empty empties the stack.
func (stack *PSStack) Empty() {
	*stack = []PSObject{}
}

// Push pushes an object on top of the stack.
func (stack *PSStack) Push(obj PSObject) error {
	if len(*stack) > 100 {
		return ErrStackOverflow
	}

	*stack = append(*stack, obj)
	return nil
}

// Pop pops an object from the top of the stack.
func (stack *PSStack) Pop() (PSObject, error) {
	if len(*stack) < 1 {
		return nil, ErrStackUnderflow
	}

	obj := (*stack)[len(*stack)-1]
	*stack = (*stack)[0 : len(*stack)-1]

	return obj, nil
}

// PopInteger specificially pops an integer from the top of the stack, returning the value as an int.
func (stack *PSStack) PopInteger() (int, error) {
	obj, err := stack.Pop()
	if err != nil {
		return 0, err
	}

	if number, is := obj.(*PSInteger); is {
		return number.Val, nil
	}
	return 0, ErrTypeCheck
}

// PopNumberAsFloat64 pops and return the numeric value of the top of the stack as a float64.
// Real or integer only.
func (stack *PSStack) PopNumberAsFloat64() (float64, error) {
	obj, err := stack.Pop()
	if err != nil {
		return 0, err
	}

	if number, is := obj.(*PSReal); is {
		return number.Val, nil
	} else if number, is := obj.(*PSInteger); is {
		return float64(number.Val), nil
	} else {
		return 0, ErrTypeCheck
	}
}

// String returns a string representation of the stack.
func (stack *PSStack) String() string {
	s := "[ "
	for _, obj := range *stack {
		s += obj.String()
		s += " "
	}
	s += "]"

	return s
}

// DebugString returns a descriptive string representation of the stack - intended for debugging.
func (stack *PSStack) DebugString() string {
	s := "[ "
	for _, obj := range *stack {
		s += obj.DebugString()
		s += " "
	}
	s += "]"

	return s
}
