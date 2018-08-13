/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

import (
	"errors"
)

// tolerance specifies the tolerance for equivalence when comparing real values.
const tolerance = 0.000001

// ErrStackUnderflow is due to a stack underflow.
var ErrStackUnderflow = errors.New("stack underflow")

// ErrStackOverflow is due to a stack overflow.
var ErrStackOverflow = errors.New("stack overflow")

// ErrTypeCheck is due to a type mismatch, typically when a type is expected as input but a different type
// is received instead.
var ErrTypeCheck = errors.New("type check error")

// ErrRangeCheck occurs when an input value is incorrect or within valid boundaeries.
var ErrRangeCheck = errors.New("range check error")

// ErrUnsupportedOperand occurs when an unsupported operand is encountered.
var ErrUnsupportedOperand = errors.New("unsupported operand")

// ErrUndefinedResult occurs when the function does not have a result for given input parameters. An example
// is division by 0.
var ErrUndefinedResult = errors.New("undefined result error")
