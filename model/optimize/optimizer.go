/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"github.com/unidoc/unipdf/v3/core"
)

// New creates a optimizers chain from options.
func New(options Options) *Chain {
	chain := new(Chain)
	if options.ImageUpperPPI > 0 {
		imageOptimizer := new(ImagePPI)
		imageOptimizer.ImageUpperPPI = options.ImageUpperPPI
		chain.Append(imageOptimizer)
	}
	if options.ImageQuality > 0 {
		imageOptimizer := new(Image)
		imageOptimizer.ImageQuality = options.ImageQuality
		chain.Append(imageOptimizer)
	}
	if options.CombineDuplicateDirectObjects {
		chain.Append(new(CombineDuplicateDirectObjects))
	}
	if options.CombineDuplicateStreams {
		chain.Append(new(CombineDuplicateStreams))
	}
	if options.CombineIdenticalIndirectObjects {
		chain.Append(new(CombineIdenticalIndirectObjects))
	}
	if options.UseObjectStreams {
		chain.Append(new(ObjectStreams))
	}
	if options.CompressStreams {
		chain.Append(new(CompressStreams))
	}
	return chain
}

// replaceObjectsInPlace replaces objects. objTo will be modified by the process.
func replaceObjectsInPlace(objects []core.PdfObject, objTo map[core.PdfObject]core.PdfObject) {
	if objTo == nil || len(objTo) == 0 {
		return
	}
	for i, obj := range objects {
		if to, found := objTo[obj]; found {
			objects[i] = to
			continue
		}
		objTo[obj] = obj
		switch t := obj.(type) {
		case *core.PdfObjectArray:
			values := make([]core.PdfObject, t.Len())
			copy(values, t.Elements())
			replaceObjectsInPlace(values, objTo)
			for i, obj := range values {
				t.Set(i, obj)
			}
		case *core.PdfObjectStreams:
			replaceObjectsInPlace(t.Elements(), objTo)
		case *core.PdfObjectStream:
			values := []core.PdfObject{t.PdfObjectDictionary}
			replaceObjectsInPlace(values, objTo)
			t.PdfObjectDictionary = values[0].(*core.PdfObjectDictionary)
		case *core.PdfObjectDictionary:
			keys := t.Keys()
			values := make([]core.PdfObject, len(keys))
			for i, key := range keys {
				values[i] = t.Get(key)
			}
			replaceObjectsInPlace(values, objTo)
			for i, key := range keys {
				t.Set(key, values[i])
			}
		case *core.PdfIndirectObject:
			values := []core.PdfObject{t.PdfObject}
			replaceObjectsInPlace(values, objTo)
			t.PdfObject = values[0]
		}
	}
}

// Update all the object numbers prior to get hash of objects.
func updateObjectNumbers(objects []core.PdfObject) {
	// Update numbers
	for idx, obj := range objects {
		switch o := obj.(type) {
		case *core.PdfIndirectObject:
			o.ObjectNumber = int64(idx + 1)
			o.GenerationNumber = 0
		case *core.PdfObjectStream:
			o.ObjectNumber = int64(idx + 1)
			o.GenerationNumber = 0
		case *core.PdfObjectStreams:
			o.ObjectNumber = int64(idx + 1)
			o.GenerationNumber = 0
		}
	}
}
