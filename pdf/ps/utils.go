/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

// MakeReal returns a new PSReal object initialized with `val`.
func MakeReal(val float64) *PSReal {
	obj := PSReal{}
	obj.Val = val
	return &obj
}

// MakeInteger returns a new PSInteger object initialized with `val`.
func MakeInteger(val int) *PSInteger {
	obj := PSInteger{}
	obj.Val = val
	return &obj
}

// MakeBool returns a new PSBoolean object initialized with `val`.
func MakeBool(val bool) *PSBoolean {
	obj := PSBoolean{}
	obj.Val = val
	return &obj
}

// MakeOperand returns a new PSOperand object based on string `val`.
func MakeOperand(val string) *PSOperand {
	obj := PSOperand(val)
	return &obj
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
