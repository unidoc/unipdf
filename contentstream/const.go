/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import "errors"

var (
	// ErrInvalidOperand specifies that invalid operands have been encountered
	// while parsing the content stream.
	ErrInvalidOperand = errors.New("invalid operand")
)
