/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package encoder

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
)

// Encoder is the jbig2 encoder structure used for encoding the image into the
type Encoder struct {
	// some paramete
	doc *document.Document
}

// EncodeBytes encodes input 'data' and encoding 'parameters' into jbig2 encoding.
func (e *Encoder) EncodeBytes(data []byte, parameters Parameters) ([]byte, error) {
	return nil, nil
}
