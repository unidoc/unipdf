/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	. "github.com/unidoc/unidoc/pdf/core"
)

// A PDFModel is a higher level PDF construct which can be collapsed into a PDF primitive.
// Each PDFModel has an underlying Primitive and vice versa.
// Copies can be made, but care must be taken to do it properly.
type PdfModel interface {
	ToPdfObject() PdfObject
	GetContainingPdfObject() PdfObject
}

// The model manager is used to cache Primitive <-> Model mappings where needed.
// In many cases only Model -> Primitive mapping is needed and only a reference to the Primitive
// is stored in the Model.  In some cases, the Model needs to be found from the Primitive,
// and that is where the ModelManager can be used (in both directions).
//
// Note that it is not always used, the Primitive <-> Model mapping needs to be registered
// for each time it is used.  Thus, it is only used for special cases, commonly where the same
// object is used by two higher level objects. (Example PDF Widgets owned by both Page Annotations,
// and the interactive form - AcroForm).
type ModelManager struct {
	primitiveCache map[PdfModel]PdfObject
	modelCache     map[PdfObject]PdfModel
}

func NewModelManager() *ModelManager {
	mm := ModelManager{}
	mm.primitiveCache = map[PdfModel]PdfObject{}
	mm.modelCache = map[PdfObject]PdfModel{}
	return &mm
}

// Register (cache) a model to primitive relationship.
func (this *ModelManager) Register(primitive PdfObject, model PdfModel) {
	this.primitiveCache[model] = primitive
	this.modelCache[primitive] = model
}

func (this *ModelManager) GetPrimitiveFromModel(model PdfModel) PdfObject {
	primitive, has := this.primitiveCache[model]
	if !has {
		return nil
	}
	return primitive
}

func (this *ModelManager) GetModelFromPrimitive(primitive PdfObject) PdfModel {
	model, has := this.modelCache[primitive]
	if !has {
		return nil
	}
	return model
}
