/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

import (
	"errors"
)

// Tolerance for comparing real values.
const TOLERANCE = 0.000001

// Common errors.
var ErrStackUnderflow = errors.New("Stack underflow")
var ErrStackOverflow = errors.New("Stack overflow")
var ErrTypeCheck = errors.New("Type check error")
var ErrRangeCheck = errors.New("Range check error")
var ErrUndefinedResult = errors.New("Undefined result error")
