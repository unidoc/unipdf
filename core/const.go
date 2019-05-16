/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import "errors"

// Common errors that may occur on PDF parsing/writing.
var (
	// ErrUnsupportedEncodingParameters error indicates that encoding/decoding was attempted with unsupported
	// encoding parameters.
	// For example when trying to encode with an unsupported Predictor (flate).
	ErrUnsupportedEncodingParameters = errors.New("unsupported encoding parameters")
	ErrNoCCITTFaxDecode              = errors.New("CCITTFaxDecode encoding is not yet implemented")
	ErrNoJBIG2Decode                 = errors.New("JBIG2Decode encoding is not yet implemented")
	ErrNoJPXDecode                   = errors.New("JPXDecode encoding is not yet implemented")
	ErrNoPdfVersion                  = errors.New("version not found")
	ErrTypeError                     = errors.New("type check error")
	ErrRangeError                    = errors.New("range check error")
	ErrNotSupported                  = errors.New("feature not currently supported")
	ErrNotANumber                    = errors.New("not a number")
)
