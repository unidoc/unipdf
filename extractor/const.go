/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import "errors"

var isTesting = false

var (
	errTypeCheck = errors.New("type check error")
)
