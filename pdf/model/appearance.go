/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import "github.com/unidoc/unidoc/pdf/core"

// PdfAppearance contains the common attributes of an appearance form field.
type PdfAppearance struct {
	*PdfField
	*PdfAnnotationWidget
	// TODO(gunnsth): Signature is not really part of an appearance.
	Signature *PdfSignature
}

// NewPdfAppearance returns an initialized annotation widget.
func NewPdfAppearance() *PdfAppearance {
	app := &PdfAppearance{}
	app.PdfField = NewPdfField()
	app.PdfAnnotationWidget = NewPdfAnnotationWidget()
	app.PdfField.SetContext(app)
	app.PdfAnnotationWidget.SetContext(app)
	app.PdfAnnotationWidget.container = app.PdfField.container
	return app
}

// ToPdfObject implements interface PdfModel.
func (app *PdfAppearance) ToPdfObject() core.PdfObject {
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
