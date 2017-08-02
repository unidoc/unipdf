/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import "errors"

var (
	// ErrUnsupportedEncodingParameters error indicates that encoding/decoding was attempted with unsupported
	// encoding parameters.
	// For example when trying to encode with an unsupported Predictor (flate).
	ErrUnsupportedEncodingParameters = errors.New("Unsupported encoding parameters")
)
