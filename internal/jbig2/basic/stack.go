/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package basic

// Stack is the LIFO data structure implementation
type Stack struct {
	// Data keeps the stack's values.
	Data []interface{}
	// Aux is the auxiliary additional stack use for some helpers.
	Aux *Stack
}

// Len returns the size of the stack.
func (s *Stack) Len() int {
	return len(s.Data)
}

// Peek returns the top element of the stack 's'.
// returns false if the stack is zero length.
func (s *Stack) Peek() (v interface{}, ok bool) {
	return s.peek()
}

// Pop the top element of the slack and returns it.
// Returns false if the stack is 'zero' length.
func (s *Stack) Pop() (v interface{}, ok bool) {
	v, ok = s.peek()
	if !ok {
		return nil, ok
	}

	// remove it from the stack.
	s.Data = s.Data[:s.top()]
	return v, true
}

// Push adds the 'v' element to the top of the stack.
func (s *Stack) Push(v interface{}) {
	s.Data = append(s.Data, v)
}

func (s *Stack) peek() (interface{}, bool) {
	top := s.top()
	// check if the stack is zero size.
	if top == -1 {
		return nil, false
	}

	// get the last element.
	return s.Data[top], true
}

func (s *Stack) top() int {
	return len(s.Data) - 1
}
