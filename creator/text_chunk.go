/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// TextChunk represents a chunk of text along with a particular style.
type TextChunk struct {
	// The text that is being rendered in the PDF.
	Text string

	// The style of the text being rendered.
	Style TextStyle

	// Text chunk annotation.
	annotation *model.PdfAnnotation

	// Internally used in order to skip processing the annotation
	// if it has already been processed by the parent component.
	annotationProcessed bool
}

// SetAnnotation sets a annotation on a TextChunk
func (tc *TextChunk) SetAnnotation(annotation *model.PdfAnnotation) {
	tc.annotation = annotation
}

// newTextChunk returns a new text chunk instance.
func newTextChunk(text string, style TextStyle) *TextChunk {
	return &TextChunk{
		Text:  text,
		Style: style,
	}
}

// newExternalLinkAnnotation returns a new external link annotation.
func newExternalLinkAnnotation(url string) *model.PdfAnnotation {
	annotation := model.NewPdfAnnotationLink()

	// Set border style.
	bs := model.NewBorderStyle()
	bs.SetBorderWidth(0)
	annotation.BS = bs.ToPdfObject()

	// Set link destination.
	action := model.NewPdfActionURI()
	action.URI = core.MakeString(url)
	annotation.SetAction(action.PdfAction)

	return annotation.PdfAnnotation
}

// newExternalLinkAnnotation returns a new internal link annotation.
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

	return annotation.PdfAnnotation
}

// copyLinkAnnotation returns a new link annotation based on an existing one.
func copyLinkAnnotation(link *model.PdfAnnotationLink) *model.PdfAnnotationLink {
	if link == nil {
		return nil
	}

	annotation := model.NewPdfAnnotationLink()
	annotation.BS = link.BS
	annotation.A = link.A
	if action, err := link.GetAction(); err == nil && action != nil {
		annotation.SetAction(action)
	}

	if annotDest, ok := link.Dest.(*core.PdfObjectArray); ok {
		annotation.Dest = core.MakeArray(annotDest.Elements()...)
	}

	return annotation
}
