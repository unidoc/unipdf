/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

func MakeReal(val float64) PSObject {
	obj := PSReal{}
	obj.Val = val
	return &obj
}

func MakeInteger(val int) PSObject {
	obj := PSInteger{}
	obj.Val = val
	return &obj
}

func MakeBool(val bool) *PSBoolean {
	obj := PSBoolean{}
	obj.Val = val
	return &obj
}

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
