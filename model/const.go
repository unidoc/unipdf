/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"

	"github.com/unidoc/unipdf/v3/core"
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
	ErrFontNotSupported         = fmt.Errorf("unsupported font (%v)", core.ErrNotSupported)
	ErrType1CFontNotSupported   = fmt.Errorf("Type1C fonts are not currently supported (%v)", core.ErrNotSupported)
	ErrType3FontNotSupported    = fmt.Errorf("Type3 fonts are not currently supported (%v)", core.ErrNotSupported)
	ErrTTCmapNotSupported       = fmt.Errorf("unsupported TrueType cmap format (%v)", core.ErrNotSupported)
)
