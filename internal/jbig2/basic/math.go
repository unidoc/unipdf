/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package basic

// Abs get the absolute value of the integer 'v'.
func Abs(v int) int {
	if v > 0 {
		return v
	}
	return -v
}

// Ceil gets the 'ceil' value for the provided 'numerator' and 'denominator'.
func Ceil(numerator, denominator int) int {
	if numerator%denominator == 0 {
		return numerator / denominator
	}
	return (numerator / denominator) + 1
}

// Max gets the maximum value from the provided 'x', 'y' arguments.
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Min gets the minimal value from the provided 'x' and 'y' arguments.
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Sign gets the float32 sign of the 'v' value.
// If the value 'v' is greater or equal to 0.0 the function returns 1.0.
// Otherwise it returns '-1.0'.
func Sign(v float32) float32 {
	if v >= 0.0 {
		return 1.0
	}
	return -1.0
}
