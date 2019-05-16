/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"github.com/unidoc/unipdf/v3/core"
)

// ObjectStreams groups PDF objects to object streams.
// It implements interface model.Optimizer.
type ObjectStreams struct {
}

// Optimize optimizes PDF objects to decrease PDF size.
func (o *ObjectStreams) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	objStream := &core.PdfObjectStreams{}
	skippedObjects := make([]core.PdfObject, 0, len(objects))
	for _, obj := range objects {
		if io, isIndirectObj := obj.(*core.PdfIndirectObject); isIndirectObj && io.GenerationNumber == 0 {
			objStream.Append(obj)
		} else {
			skippedObjects = append(skippedObjects, obj)
		}
	}
	if objStream.Len() == 0 {
		return skippedObjects, nil
	}

	optimizedObjects = make([]core.PdfObject, 0, len(skippedObjects)+objStream.Len()+1)
	if objStream.Len() > 1 {
		optimizedObjects = append(optimizedObjects, objStream)
	}
	optimizedObjects = append(optimizedObjects, objStream.Elements()...)
	optimizedObjects = append(optimizedObjects, skippedObjects...)

	return optimizedObjects, nil
}
