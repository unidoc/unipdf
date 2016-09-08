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
func (this *ObjectManager) Register(primitive PdfObject, model PdfModel) {
	this.primitiveCache[model] = primitive
	this.modelCache[primitive] = model
}

func (this *ObjectManager) GetPrimitiveFromModel(model PdfModel) PdfObject {
	primitive, has := this.primitiveCache[model]
	if !has {
		return nil
	}
	return primitive
}

func (this *ObjectManager) GetModelFromPrimitive(primitive PdfObject) PdfObjectConvertible {
	model, has := this.modelCache[primitive]
	if !has {
		return nil
	}
	return model
}
