/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import "github.com/unidoc/unidoc/pdf/core"

// PdfSignatureAppearance defines a signature with a specified form field and
// annotation widget for appearance styling.
type PdfSignatureAppearance struct {
	*PdfField
	*PdfAnnotationWidget
	
	Signature *PdfSignature
}

// NewPdfSignatureAppearance returns an initialized signature field appearance.
func NewPdfSignatureAppearance() *PdfSignatureAppearance {
	app := &PdfSignatureAppearance{}
	app.PdfField = NewPdfField()
	app.PdfAnnotationWidget = NewPdfAnnotationWidget()
	app.PdfField.SetContext(app)
	app.PdfAnnotationWidget.SetContext(app)
	app.PdfAnnotationWidget.container = app.PdfField.container
	return app
}

// ToPdfObject implements interface PdfModel.
func (app *PdfSignatureAppearance) ToPdfObject() core.PdfObject {
	if app.Signature != nil {
		app.V = app.Signature.ToPdfObject()
	}
	app.PdfAnnotation.ToPdfObject()
	app.PdfField.ToPdfObject()

	container := app.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("Widget"))
	d.SetIfNotNil("H", app.H)
	d.SetIfNotNil("MK", app.MK)
	d.SetIfNotNil("A", app.A)
	d.SetIfNotNil("AA", app.PdfAnnotationWidget.AA)
	d.SetIfNotNil("BS", app.BS)
	d.SetIfNotNil("Parent", app.PdfAnnotationWidget.Parent)

	return container
}
