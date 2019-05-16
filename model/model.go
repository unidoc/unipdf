/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"github.com/unidoc/unipdf/v3/core"
)

// PdfModel is a higher level PDF construct which can be collapsed into a PdfObject.
// Each PdfModel has an underlying PdfObject and vice versa (one-to-one).
// Under normal circumstances there should only be one copy of each.
// Copies can be made, but care must be taken to do it properly.
type PdfModel interface {
	ToPdfObject() core.PdfObject
	GetContainingPdfObject() core.PdfObject
}

// modelManager is used to cache primitive PdfObject <-> Model mappings where needed (one-to-one).
// In many cases only Model -> PdfObject mapping is needed and only a reference to the PdfObject
// is stored in the Model.  In some cases, the Model needs to be found from the PdfObject,
// and that is where the modelManager can be used (in both directions).
//
// Note that it is not always used, the PdfObject <-> Model mapping needs to be registered
// for each time it is used.  Thus, it is only used for special cases, commonly where the same
// object is used by two higher level objects. (Example PDF Widgets owned by both Page Annotations,
// and the interactive form - AcroForm).
type modelManager struct {
	primitiveCache map[PdfModel]core.PdfObject
	modelCache     map[core.PdfObject]PdfModel
}

// newModelManager returns a new initialized modelManager.
func newModelManager() *modelManager {
	mm := modelManager{}
	mm.primitiveCache = map[PdfModel]core.PdfObject{}
	mm.modelCache = map[core.PdfObject]PdfModel{}
	return &mm
}

// Register registers (caches) a model to primitive object relationship.
func (mm *modelManager) Register(primitive core.PdfObject, model PdfModel) {
	mm.primitiveCache[model] = primitive
	mm.modelCache[primitive] = model
}

// GetPrimitiveFromModel returns the primitive object corresponding to the input `model`.
func (mm *modelManager) GetPrimitiveFromModel(model PdfModel) core.PdfObject {
	primitive, has := mm.primitiveCache[model]
	if !has {
		return nil
	}
	return primitive
}

// GetModelFromPrimitive returns the model corresponding to the `primitive` PdfObject.
func (mm *modelManager) GetModelFromPrimitive(primitive core.PdfObject) PdfModel {
	model, has := mm.modelCache[primitive]
	if !has {
		return nil
	}
	return model
}
