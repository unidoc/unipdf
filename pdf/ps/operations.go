/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

import "math"

//
// Operation implementations for each supported operand.
//

// Absolute value.
func (op *PSOperand) abs(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if num, is := obj.(*PSReal); is {
		val := num.Val
		if val < 0 {
			err = stack.Push(MakeReal(-val))
		} else {
			err = stack.Push(MakeReal(val))
		}
	} else if num, is := obj.(*PSInteger); is {
		val := num.Val
		if val < 0 {
			err = stack.Push(MakeInteger(-val))
		} else {
			err = stack.Push(MakeInteger(val))
		}
	} else {
		return ErrTypeCheck
	}

	return err
}

// 5 27 add -> 32
func (op *PSOperand) add(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	real1, isReal1 := obj1.(*PSReal)
	int1, isInt1 := obj1.(*PSInteger)
	if !isReal1 && !isInt1 {
		return ErrTypeCheck
	}

	real2, isReal2 := obj2.(*PSReal)
	int2, isInt2 := obj2.(*PSInteger)
	if !isReal2 && !isInt2 {
		return ErrTypeCheck
	}

	// If both numbers integers -> integer output.
	if isInt1 && isInt2 {
		result := int1.Val + int2.Val
		err := stack.Push(MakeInteger(result))
		return err
	}

	// Otherwise -> real output.
	var result float64 = 0
	if isReal1 {
		result = real1.Val
	} else {
		result = float64(int1.Val)
	}

	if isReal2 {
		result += real2.Val
	} else {
		result += float64(int2.Val)
	}

	err = stack.Push(MakeReal(result))
	return err
}

// And operation.
// if bool: returns the logical "and" of the inputs
// bool1 bool2 and -> bool3
// if int: returns the bitwise "and" of the inputs
// int1 int2 and -> int3
func (op *PSOperand) and(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	// Boolean inputs.
	if bool1, is := obj1.(*PSBoolean); is {
		bool2, ok := obj2.(*PSBoolean)
		if !ok {
			return ErrTypeCheck
		}
		err = stack.Push(MakeBool(bool1.Val && bool2.Val)) // logical and
		return err
	}

	// Integer inputs
	if int1, is := obj1.(*PSInteger); is {
		int2, ok := obj2.(*PSInteger)
		if !ok {
			return ErrTypeCheck
		}
		err = stack.Push(MakeInteger(int1.Val & int2.Val)) // bitwise and
		return err
	}

	return ErrTypeCheck
}

// den num atan -> atan(num/den) in degrees.
// result is a real value.
func (op *PSOperand) atan(stack *PSStack) error {
	// Denominator
	den, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	// Numerator
	num, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	// special cases.
	// atan(inf) -> 90
	// atan(-inf) -> 270
	if den == 0 {
		var err error
		if num < 0 {
			err = stack.Push(MakeReal(270))
		} else {
			err = stack.Push(MakeReal(90))
		}
		return err
	}

	ratio := num / den
	angleDeg := math.Atan(ratio) * 180 / math.Pi

	err = stack.Push(MakeReal(angleDeg))
	return err
}

// bitshift
// int1 shift bitshift -> int2
func (op *PSOperand) bitshift(stack *PSStack) error {
	shift, err := stack.PopInteger()
	if err != nil {
		return err
	}

	int1, err := stack.PopInteger()
	if err != nil {
		return err
	}

	var result int

	if shift >= 0 {
		result = int1 << uint(shift)
	} else {
		result = int1 >> uint(-shift)
	}

	err = stack.Push(MakeInteger(result))
	return err
}

// Ceiling of number.
// num1 ceiling -> num2
// The type of the result is the same as of the operand.
func (op *PSOperand) ceiling(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if num, is := obj.(*PSReal); is {
		err = stack.Push(MakeReal(math.Ceil(num.Val)))
	} else if num, is := obj.(*PSInteger); is {
		err = stack.Push(MakeInteger(num.Val))
	} else {
		err = ErrTypeCheck
	}

	return err
}

// Copy
// any1 ... anyn n copy -> any1 ... anyn any1 ... anyn
func (op *PSOperand) copy(stack *PSStack) error {
	n, err := stack.PopInteger()
	if err != nil {
		return err
	}

	if n < 0 {
		return ErrRangeCheck
	}

	if n > len(*stack) {
		return ErrRangeCheck
	}

	*stack = append(*stack, (*stack)[len(*stack)-n:]...)
	return nil
}

// Cosine
// angle cos -> real
// Angle is in degrees
func (op *PSOperand) cos(stack *PSStack) error {
	angle, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	result := math.Cos(angle * math.Pi / 180.0)
	err = stack.Push(MakeReal(result))
	return err
}

// Convert to integer
func (op *PSOperand) cvi(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if num, is := obj.(*PSReal); is {
		val := int(num.Val)
		err = stack.Push(MakeInteger(val))
	} else if num, is := obj.(*PSInteger); is {
		val := num.Val
		err = stack.Push(MakeInteger(val))
	} else {
		return ErrTypeCheck
	}

	return err
}

// Convert number tor real
func (op *PSOperand) cvr(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if num, is := obj.(*PSReal); is {
		err = stack.Push(MakeReal(num.Val))
	} else if num, is := obj.(*PSInteger); is {
		err = stack.Push(MakeReal(float64(num.Val)))
	} else {
		return ErrTypeCheck
	}

	return err
}

func (op *PSOperand) div(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	real1, isReal1 := obj1.(*PSReal)
	int1, isInt1 := obj1.(*PSInteger)
	if !isReal1 && !isInt1 {
		return ErrTypeCheck
	}
	// Cannot be 0.
	if isReal1 && real1.Val == 0 {
		return ErrUndefinedResult
	}
	if isInt1 && int1.Val == 0 {
		return ErrUndefinedResult
	}

	real2, isReal2 := obj2.(*PSReal)
	int2, isInt2 := obj2.(*PSInteger)
	if !isReal2 && !isInt2 {
		return ErrTypeCheck
	}

	// Float output.
	var result float64 = 0
	if isReal2 {
		result = real2.Val
	} else {
		result = float64(int2.Val)
	}

	if isReal1 {
		result /= real1.Val
	} else {
		result /= float64(int1.Val)
	}

	err = stack.Push(MakeReal(result))
	return err
}

// Duplicates the top object on the stack (dup)
func (op *PSOperand) dup(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	// Push it back.
	err = stack.Push(obj)
	if err != nil {
		return err
	}
	// Push the duplicate.
	err = stack.Push(obj.Duplicate())
	return err
}

// Check for equality.
// any1 any2 eq bool
func (op *PSOperand) eq(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	// bool, real, int
	// if bool, both must be bool
	bool1, isBool1 := obj1.(*PSBoolean)
	bool2, isBool2 := obj2.(*PSBoolean)
	if isBool1 || isBool2 {
		var err error
		if isBool1 && isBool2 {
			err = stack.Push(MakeBool(bool1.Val == bool2.Val))
		} else {
			// Type mismatch -> false
			err = stack.Push(MakeBool(false))
		}
		return err
	}

	var val1 float64
	var val2 float64

	if number, is := obj1.(*PSInteger); is {
		val1 = float64(number.Val)
	} else if number, is := obj1.(*PSReal); is {
		val1 = number.Val
	} else {
		return ErrTypeCheck
	}

	if number, is := obj2.(*PSInteger); is {
		val2 = float64(number.Val)
	} else if number, is := obj2.(*PSReal); is {
		val2 = number.Val
	} else {
		return ErrTypeCheck
	}

	if math.Abs(val2-val1) < tolerance {
		err = stack.Push(MakeBool(true))
	} else {
		err = stack.Push(MakeBool(false))
	}

	return err
}

// Exchange the top two elements of the stack (exch)
func (op *PSOperand) exch(stack *PSStack) error {
	top, err := stack.Pop()
	if err != nil {
		return err
	}

	next, err := stack.Pop()
	if err != nil {
		return err
	}

	err = stack.Push(top)
	if err != nil {
		return err
	}
	err = stack.Push(next)

	return err
}

// base exponent exp -> base^exp
// Raises base to exponent power.
// The result is a real number.
func (op *PSOperand) exp(stack *PSStack) error {
	exponent, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	base, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	if math.Abs(exponent) < 1 && base < 0 {
		return ErrUndefinedResult
	}

	result := math.Pow(base, exponent)
	err = stack.Push(MakeReal(result))

	return err
}

// Floor of number.
func (op *PSOperand) floor(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if num, is := obj.(*PSReal); is {
		err = stack.Push(MakeReal(math.Floor(num.Val)))
	} else if num, is := obj.(*PSInteger); is {
		err = stack.Push(MakeInteger(num.Val))
	} else {
		return ErrTypeCheck
	}

	return err
}

// Greater than or equal
// num1 num2 ge -> bool; num1 >= num2
func (op *PSOperand) ge(stack *PSStack) error {
	num2, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	num1, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	// Check equlity.
	if math.Abs(num1-num2) < tolerance {
		err := stack.Push(MakeBool(true))
		return err
	} else if num1 > num2 {
		err := stack.Push(MakeBool(true))
		return err
	} else {
		err := stack.Push(MakeBool(false))
		return err
	}
}

// Greater than
// num1 num2 gt -> bool; num1 > num2
func (op *PSOperand) gt(stack *PSStack) error {
	num2, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	num1, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	// Check equlity.
	if math.Abs(num1-num2) < tolerance {
		err := stack.Push(MakeBool(false))
		return err
	} else if num1 > num2 {
		err := stack.Push(MakeBool(true))
		return err
	} else {
		err := stack.Push(MakeBool(false))
		return err
	}
}

// Integral division
// 25 3 div -> 8
func (op *PSOperand) idiv(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	int1, ok := obj1.(*PSInteger)
	if !ok {
		return ErrTypeCheck
	}
	if int1.Val == 0 {
		return ErrUndefinedResult
	}

	int2, ok := obj2.(*PSInteger)
	if !ok {
		return ErrTypeCheck
	}

	result := int2.Val / int1.Val
	err = stack.Push(MakeInteger(result))

	return err
}

// If conditional
// bool proc if -> run proc() if bool is true
func (op *PSOperand) ifCondition(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}
	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	// Type checks.
	proc, ok := obj1.(*PSProgram)
	if !ok {
		return ErrTypeCheck
	}
	condition, ok := obj2.(*PSBoolean)
	if !ok {
		return ErrTypeCheck
	}

	// Run proc if condition is true.
	if condition.Val {
		err := proc.Exec(stack)
		return err
	}

	return nil
}

// If else conditional
// bool proc1 proc2 ifelse -> execute proc1() if bool is true, otherwise proc2()
func (op *PSOperand) ifelse(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}
	obj2, err := stack.Pop()
	if err != nil {
		return err
	}
	obj3, err := stack.Pop()
	if err != nil {
		return err
	}

	// Type checks.
	proc2, ok := obj1.(*PSProgram)
	if !ok {
		return ErrTypeCheck
	}
	proc1, ok := obj2.(*PSProgram)
	if !ok {
		return ErrTypeCheck
	}
	condition, ok := obj3.(*PSBoolean)
	if !ok {
		return ErrTypeCheck
	}

	// Run proc if condition is true.
	if condition.Val {
		err := proc1.Exec(stack)
		return err
	}
	err = proc2.Exec(stack)
	return err
}

// Add a copy of the nth object in the stack to the top.
// any_n ... any_0 n index -> any_n ... any_0 any_n
// index from 0
func (op *PSOperand) index(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	n, ok := obj.(*PSInteger)
	if !ok {
		return ErrTypeCheck
	}

	if n.Val < 0 {
		return ErrRangeCheck
	}

	if n.Val > len(*stack)-1 {
		return ErrStackUnderflow
	}

	objN := (*stack)[len(*stack)-1-n.Val]

	err = stack.Push(objN.Duplicate())
	return err
}

// Less or equal
// num1 num2 le -> bool; num1 <= num2
func (op *PSOperand) le(stack *PSStack) error {
	num2, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	num1, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	// Check equlity.
	if math.Abs(num1-num2) < tolerance {
		err := stack.Push(MakeBool(true))
		return err
	} else if num1 < num2 {
		err := stack.Push(MakeBool(true))
		return err
	} else {
		err := stack.Push(MakeBool(false))
		return err
	}
}

// num log -> real
func (op *PSOperand) log(stack *PSStack) error {
	// Value
	val, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	result := math.Log10(val)
	err = stack.Push(MakeReal(result))
	return err
}

// num ln -> ln(num)
// The result is a real number.
func (op *PSOperand) ln(stack *PSStack) error {
	// Value
	val, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	result := math.Log(val)
	err = stack.Push(MakeReal(result))
	return err
}

// Less than
// num1 num2 lt -> bool; num1 < num2
func (op *PSOperand) lt(stack *PSStack) error {
	num2, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	num1, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	// Check equlity.
	if math.Abs(num1-num2) < tolerance {
		err := stack.Push(MakeBool(false))
		return err
	} else if num1 < num2 {
		err := stack.Push(MakeBool(true))
		return err
	} else {
		err := stack.Push(MakeBool(false))
		return err
	}
}

// 12 10 mod -> 2
func (op *PSOperand) mod(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	int1, ok := obj1.(*PSInteger)
	if !ok {
		return ErrTypeCheck
	}
	if int1.Val == 0 {
		return ErrUndefinedResult
	}

	int2, ok := obj2.(*PSInteger)
	if !ok {
		return ErrTypeCheck
	}

	result := int2.Val % int1.Val
	err = stack.Push(MakeInteger(result))
	return err
}

// 6 8 mul -> 48
func (op *PSOperand) mul(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	real1, isReal1 := obj1.(*PSReal)
	int1, isInt1 := obj1.(*PSInteger)
	if !isReal1 && !isInt1 {
		return ErrTypeCheck
	}

	real2, isReal2 := obj2.(*PSReal)
	int2, isInt2 := obj2.(*PSInteger)
	if !isReal2 && !isInt2 {
		return ErrTypeCheck
	}

	// If both numbers integers -> integer output.
	if isInt1 && isInt2 {
		result := int1.Val * int2.Val
		err := stack.Push(MakeInteger(result))
		return err
	}

	// Otherwise -> real output.
	var result float64 = 0
	if isReal1 {
		result = real1.Val
	} else {
		result = float64(int1.Val)
	}

	if isReal2 {
		result *= real2.Val
	} else {
		result *= float64(int2.Val)
	}

	err = stack.Push(MakeReal(result))
	return err
}

// Not equal (inverse of eq)
// any1 any2 ne -> bool
func (op *PSOperand) ne(stack *PSStack) error {
	// Simply call equate and then negate the result.
	// Implementing directly could be more efficient, but probably not a big deal in most cases.
	err := op.eq(stack)
	if err != nil {
		return err
	}

	err = op.not(stack)
	return err
}

// Negate
// 6 neg -> -6
func (op *PSOperand) neg(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if real, isReal := obj.(*PSReal); isReal {
		err = stack.Push(MakeReal(-real.Val))
		return err
	} else if inum, isInt := obj.(*PSInteger); isInt {
		err = stack.Push(MakeInteger(-inum.Val))
		return err

	} else {
		return ErrTypeCheck
	}
}

// Logical/bitwise negation
// bool1 not -> bool2 (logical)
// int1 not ->  int2 (bitwise)
func (op *PSOperand) not(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if bool1, is := obj.(*PSBoolean); is {
		err = stack.Push(MakeBool(!bool1.Val))
		return err
	} else if int1, isInt := obj.(*PSInteger); isInt {
		err = stack.Push(MakeInteger(^int1.Val))
		return err
	} else {
		return ErrTypeCheck
	}
}

// OR logical/bitwise operation.
// bool1 bool2 or -> bool3 (logical or)
// int1 int2 or -> int3 (bitwise or)
func (op *PSOperand) or(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	// Boolean inputs (logical).
	if bool1, is := obj1.(*PSBoolean); is {
		bool2, ok := obj2.(*PSBoolean)
		if !ok {
			return ErrTypeCheck
		}
		err = stack.Push(MakeBool(bool1.Val || bool2.Val))
		return err
	}

	// Integer inputs (bitwise).
	if int1, is := obj1.(*PSInteger); is {
		int2, ok := obj2.(*PSInteger)
		if !ok {
			return ErrTypeCheck
		}
		err = stack.Push(MakeInteger(int1.Val | int2.Val))
		return err
	}

	return ErrTypeCheck
}

// Remove the top element on the stack (pop)
func (op *PSOperand) pop(stack *PSStack) error {
	_, err := stack.Pop()
	if err != nil {
		return err
	}
	return nil
}

// Round number off.
// num1 round -> num2
func (op *PSOperand) round(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if num, is := obj.(*PSReal); is {
		err = stack.Push(MakeReal(math.Floor(num.Val + 0.5)))
	} else if num, is := obj.(*PSInteger); is {
		err = stack.Push(MakeInteger(num.Val))
	} else {
		return ErrTypeCheck
	}

	return err
}

// Roll stack contents (num dir roll)
// num: number of elements, dir: direction
// 7 8 9 3  1 roll -> 9 7 8
// 7 8 9 3 -1 roll -> 8 9 7
// n j roll
func (op *PSOperand) roll(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	j, ok := obj1.(*PSInteger)
	if !ok {
		return ErrTypeCheck
	}

	n, ok := obj2.(*PSInteger)
	if !ok {
		return ErrTypeCheck
	}
	if n.Val < 0 {
		return ErrRangeCheck
	}
	if n.Val == 0 || n.Val == 1 {
		// Do nothing..
		return nil
	}
	if n.Val > len(*stack) {
		return ErrStackUnderflow
	}

	for i := 0; i < abs(j.Val); i++ {
		var substack []PSObject

		substack = (*stack)[len(*stack)-(n.Val) : len(*stack)]
		if j.Val > 0 {
			// if j > 0; put the top element on bottom of the substack
			top := substack[len(substack)-1]
			substack = append([]PSObject{top}, substack[0:len(substack)-1]...)
		} else {
			// if j < 0: put the bottom element on top
			bottom := substack[len(substack)-n.Val]
			substack = append(substack[1:], bottom)
		}

		s := append((*stack)[0:len(*stack)-n.Val], substack...)
		stack = &s
	}

	return nil
}

// Sine.
// angle sin -> real
// Angle is in degrees
func (op *PSOperand) sin(stack *PSStack) error {
	angle, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	result := math.Sin(angle * math.Pi / 180.0)
	err = stack.Push(MakeReal(result))
	return err
}

// Square root.
// num sqrt -> real; real=sqrt(num)
// The result is a real number.
func (op *PSOperand) sqrt(stack *PSStack) error {
	val, err := stack.PopNumberAsFloat64()
	if err != nil {
		return err
	}

	if val < 0 {
		return ErrRangeCheck
	}

	result := math.Sqrt(val)
	err = stack.Push(MakeReal(result))
	return err
}

// 8.3 6.6 sub -> 1.7 (real)
// 8 6.3 sub -> 1.7 (real)
// 8 6 sub -> 2 (int)
func (op *PSOperand) sub(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	real1, isReal1 := obj1.(*PSReal)
	int1, isInt1 := obj1.(*PSInteger)
	if !isReal1 && !isInt1 {
		return ErrTypeCheck
	}

	real2, isReal2 := obj2.(*PSReal)
	int2, isInt2 := obj2.(*PSInteger)
	if !isReal2 && !isInt2 {
		return ErrTypeCheck
	}

	// If both numbers integers -> integer output.
	if isInt1 && isInt2 {
		result := int2.Val - int1.Val
		err := stack.Push(MakeInteger(result))
		return err
	}

	// Otherwise -> real output.
	var result float64 = 0
	if isReal2 {
		result = real2.Val
	} else {
		result = float64(int2.Val)
	}

	if isReal1 {
		result -= real1.Val
	} else {
		result -= float64(int1.Val)
	}

	err = stack.Push(MakeReal(result))
	return err
}

// Truncate number.
// num1 truncate -> num2
// The resulting number is the same type as the input.
func (op *PSOperand) truncate(stack *PSStack) error {
	obj, err := stack.Pop()
	if err != nil {
		return err
	}

	if num, is := obj.(*PSReal); is {
		truncated := int(num.Val)
		err = stack.Push(MakeReal(float64(truncated)))
	} else if num, is := obj.(*PSInteger); is {
		err = stack.Push(MakeInteger(num.Val))
	} else {
		return ErrTypeCheck
	}

	return err
}

// XOR logical/bitwise operation.
// bool1 bool2 xor -> bool3 (logical xor)
// int1 int2 xor -> int3 (bitwise xor)
func (op *PSOperand) xor(stack *PSStack) error {
	obj1, err := stack.Pop()
	if err != nil {
		return err
	}

	obj2, err := stack.Pop()
	if err != nil {
		return err
	}

	// Boolean inputs (logical).
	if bool1, is := obj1.(*PSBoolean); is {
		bool2, ok := obj2.(*PSBoolean)
		if !ok {
			return ErrTypeCheck
		}
		err = stack.Push(MakeBool(bool1.Val != bool2.Val))
		return err
	}

	// Integer inputs (bitwise).
	if int1, is := obj1.(*PSInteger); is {
		int2, ok := obj2.(*PSInteger)
		if !ok {
			return ErrTypeCheck
		}
		err = stack.Push(MakeInteger(int1.Val ^ int2.Val))
		return err
	}

	return ErrTypeCheck
}
