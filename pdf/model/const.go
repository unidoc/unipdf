/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
)

var (
	ErrRequiredAttributeMissing = errors.New("Required attribute missing")
	ErrInvalidAttribute         = errors.New("Invalid attribute")
	ErrTypeError                = errors.New("Type check error")
	ErrRangeError               = errors.New("Range check error")
)
