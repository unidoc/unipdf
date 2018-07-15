/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	. "github.com/unidoc/unidoc/pdf/core"
)

// PdfModel represents a high level PDF type which can be collapsed into a PDF primitive (typically a dictionary
// contained within an indirect object).
type PdfModel interface {
	ToPdfObject() PdfObject
	GetContainingPdfObject() PdfObject
}

// modelManager is used to cache PdfObject <-> Model mappings where needed.
// In many cases only Model -> PdfObject mapping is needed and only a reference to the PdfObject
// is stored in the Model.  In some cases, the Model needs to be found given the PdfObject,
// and that is where the modelManager can be used (in both directions).
//
// Note that it is not always used, the Primitive <-> Model mapping needs to be registered
// for each time it is used.  Thus, it is only used for special cases, commonly where the same
// object is used by two higher level objects.
//
// Example use case: PDF Annotation Widgets can be referenced by both Page Annotations, and the interactive
// form - AcroForm. With the cache, can check if already loaded and get the underlying model without duplication.
type modelManager struct {
	primitiveCache map[PdfModel]PdfObject
	modelCache     map[PdfObject]PdfModel
}

func newModelManager() *modelManager {
	mm := modelManager{}
	mm.primitiveCache = map[PdfModel]PdfObject{}
	mm.modelCache = map[PdfObject]PdfModel{}
	return &mm
}

// Register registers (caches) a model to primitive relationship, i.e. that a certain PdfModel is represented by
// the specific PdfObject.
func (mm *modelManager) Register(primitive PdfObject, model PdfModel) {
	mm.primitiveCache[model] = primitive
	mm.modelCache[primitive] = model
}

// GetPrimitiveFromModel looks up the PdfObject that represents the PdfModel (nil if nothing found).
func (mm *modelManager) GetPrimitiveFromModel(model PdfModel) PdfObject {
	primitive, has := mm.primitiveCache[model]
	if !has {
		return nil
	}
	return primitive
}

// GetModelFromPrimitive looks up the PdfModel represented by the PdfObject (nil if nothing found).
func (mm *modelManager) GetModelFromPrimitive(primitive PdfObject) PdfModel {
	model, has := mm.modelCache[primitive]
	if !has {
		return nil
	}
	return model
}
