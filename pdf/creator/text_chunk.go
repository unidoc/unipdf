/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// TextChunk represents a chunk of text along with a particular style.
type TextChunk struct {
	// The text that is being rendered in the PDF.
	Text string

	// The style of the text being rendered.
	Style TextStyle

	annotation *model.PdfAnnotation
}

func newTextChunk(text string, style TextStyle) *TextChunk {
	return &TextChunk{
		Text:  text,
		Style: style,
	}
}

func newExternalLinkAnnotation(location string) *model.PdfAnnotation {
	annotation := model.NewPdfAnnotationLink()

	// Set border style.
	bs := model.NewBorderStyle()
	bs.SetBorderWidth(0)
	annotation.BS = bs.ToPdfObject()

	// Set link destination.
	action := core.MakeDict()
	action.Set(core.PdfObjectName("S"), core.MakeName("URI"))
	action.Set(core.PdfObjectName("URI"), core.MakeString(location))
	annotation.A = action

	// Create default annotation rectangle.
	annotation.Rect = core.MakeArray()
	annotation.PdfAnnotation.Rect = annotation.Rect

	return annotation.PdfAnnotation
}

func newInternalLinkAnnotation(page int64, x, y, zoom float64) *model.PdfAnnotation {
	annotation := model.NewPdfAnnotationLink()

	// Set border style.
	bs := model.NewBorderStyle()
	bs.SetBorderWidth(0)
	annotation.BS = bs.ToPdfObject()

	// Set link destination.
	if page < 0 {
		page = 0
	}

	annotation.Dest = core.MakeArray(
		core.MakeInteger(page),
		core.MakeName("XYZ"),
		core.MakeFloat(x),
		core.MakeFloat(y),
		core.MakeFloat(zoom),
	)

	// Create default annotation rectangle.
	annotation.Rect = core.MakeArray()
	annotation.PdfAnnotation.Rect = annotation.Rect

	return annotation.PdfAnnotation
}
