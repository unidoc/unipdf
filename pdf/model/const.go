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
