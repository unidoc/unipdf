/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
)

// Errors when parsing/loading data in PDF.
// TODO(gunnsth): Unexport errors except if there is a clear use case.
var (
	ErrRequiredAttributeMissing = errors.New("required attribute missing")
	ErrInvalidAttribute         = errors.New("invalid attribute")
	ErrTypeCheck                = errors.New("type check")
	errRangeError               = errors.New("range check error")
	ErrEncrypted                = errors.New("file needs to be decrypted first")
	ErrNoFont                   = errors.New("font not defined")
	ErrFontNotSupported         = errors.New("unsupported font")
	ErrType1CFontNotSupported   = errors.New("Type1C fonts are not currently supported")
	ErrType3FontNotSupported    = errors.New("Type3 fonts are not currently supported")
	ErrTTCmapNotSupported       = errors.New("unsupported TrueType cmap format")
)
