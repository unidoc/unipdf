/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
)

var (
	ErrRequiredAttributeMissing = errors.New("required attribute missing")
	ErrInvalidAttribute         = errors.New("invalid attribute")
	ErrTypeError                = errors.New("type check error")
	ErrRangeError               = errors.New("range check error")
	ErrEncrypted                = errors.New("file needs to be decrypted first")
	ErrBadText                  = errors.New("could not decode text")
	ErrNoFont                   = errors.New("font not defined")
)
