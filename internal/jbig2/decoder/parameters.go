/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package decoder

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
)

// Parameters are the paramters used by the jbig2 decoder.
type Parameters struct {
	UnpaddedData bool
	Color        bitmap.Color
}
