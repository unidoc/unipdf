/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

type PSStack []PSObject

func NewPSStack() *PSStack {
	return &PSStack{}
}

func (stack *PSStack) Empty() {
	*stack = []PSObject{}
}

func (stack *PSStack) Push(obj PSObject) error {
	if len(*stack) > 100 {
		return ErrStackOverflow
	}

	*stack = append(*stack, obj)
	return nil
}

func (stack *PSStack) Pop() (PSObject, error) {
	if len(*stack) < 1 {
		return nil, ErrStackUnderflow
	}

	obj := (*stack)[len(*stack)-1]
	*stack = (*stack)[0 : len(*stack)-1]

	return obj, nil
}

func (stack *PSStack) PopInteger() (int, error) {
	obj, err := stack.Pop()
	if err != nil {
		return 0, err
	}

	if number, is := obj.(*PSInteger); is {
		return number.Val, nil
	} else {
		return 0, ErrTypeCheck
	}
}

// Pop and return the numeric value of the top of the stack as a float64.
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

func (this *PSStack) String() string {
	s := "[ "
	for _, obj := range *this {
		s += obj.String()
		s += " "
	}
	s += "]"

	return s
}

func (this *PSStack) DebugString() string {
	s := "[ "
	for _, obj := range *this {
		s += obj.DebugString()
		s += " "
	}
	s += "]"

	return s
}
