/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package annotator

import (
	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/contentstream/draw"
	pdfcore "github.com/unidoc/unipdf/v3/core"
	pdf "github.com/unidoc/unipdf/v3/model"
)

// RectangleAnnotationDef is a rectangle defined with a specified Width and Height and a lower left corner at (X,Y).
// The rectangle can optionally have a border and a filling color.
// The Width/Height includes the border (if any specified).
type RectangleAnnotationDef struct {
	X             float64
	Y             float64
	Width         float64
	Height        float64
	FillEnabled   bool // Show fill?
	FillColor     *pdf.PdfColorDeviceRGB
	BorderEnabled bool // Show border?
	BorderWidth   float64
	BorderColor   *pdf.PdfColorDeviceRGB
	Opacity       float64 // Alpha value (0-1).
}

// CreateRectangleAnnotation creates a rectangle annotation object that can be added to page PDF annotations.
func CreateRectangleAnnotation(rectDef RectangleAnnotationDef) (*pdf.PdfAnnotation, error) {
	rectAnnotation := pdf.NewPdfAnnotationSquare()

	if rectDef.BorderEnabled {
		r, g, b := rectDef.BorderColor.R(), rectDef.BorderColor.G(), rectDef.BorderColor.B()
		rectAnnotation.C = pdfcore.MakeArrayFromFloats([]float64{r, g, b})
		bs := pdf.NewBorderStyle()
		bs.SetBorderWidth(rectDef.BorderWidth)
		rectAnnotation.BS = bs.ToPdfObject()
	}

	if rectDef.FillEnabled {
		r, g, b := rectDef.FillColor.R(), rectDef.FillColor.G(), rectDef.FillColor.B()
		rectAnnotation.IC = pdfcore.MakeArrayFromFloats([]float64{r, g, b})
	} else {
		rectAnnotation.IC = pdfcore.MakeArrayFromIntegers([]int{}) // No fill.
	}

	if rectDef.Opacity < 1.0 {
		rectAnnotation.CA = pdfcore.MakeFloat(rectDef.Opacity)
	}

	// Make the appearance stream (for uniform appearance).
	apDict, bbox, err := makeRectangleAnnotationAppearanceStream(rectDef)
	if err != nil {
		return nil, err
	}

	rectAnnotation.AP = apDict
	rectAnnotation.Rect = pdfcore.MakeArrayFromFloats([]float64{bbox.Llx, bbox.Lly, bbox.Urx, bbox.Ury})

	return rectAnnotation.PdfAnnotation, nil

}

func makeRectangleAnnotationAppearanceStream(rectDef RectangleAnnotationDef) (*pdfcore.PdfObjectDictionary, *pdf.PdfRectangle, error) {
	form := pdf.NewXObjectForm()
	form.Resources = pdf.NewPdfPageResources()

	gsName := ""
	if rectDef.Opacity < 1.0 {
		// Create graphics state with right opacity.
		gsState := pdfcore.MakeDict()
		gsState.Set("ca", pdfcore.MakeFloat(rectDef.Opacity))
		gsState.Set("CA", pdfcore.MakeFloat(rectDef.Opacity))
		err := form.Resources.AddExtGState("gs1", gsState)
		if err != nil {
			common.Log.Debug("Unable to add extgstate gs1")
			return nil, nil, err
		}

		gsName = "gs1"
	}

	content, localBbox, globalBbox, err := drawPdfRectangle(rectDef, gsName)
	if err != nil {
		return nil, nil, err
	}

	err = form.SetContentStream(content, nil)
	if err != nil {
		return nil, nil, err
	}

	// Local bounding box for the XObject Form.
	form.BBox = localBbox.ToPdfObject()

	apDict := pdfcore.MakeDict()
	apDict.Set("N", form.ToPdfObject())

	return apDict, globalBbox, nil
}

func drawPdfRectangle(rectDef RectangleAnnotationDef, gsName string) ([]byte, *pdf.PdfRectangle, *pdf.PdfRectangle, error) {
	// The annotation is drawn locally in a relative coordinate system with 0,0 as the origin rather than an offset.
	rect := draw.Rectangle{
		X:             0,
		Y:             0,
		Width:         rectDef.Width,
		Height:        rectDef.Height,
		FillEnabled:   rectDef.FillEnabled,
		FillColor:     rectDef.FillColor,
		BorderEnabled: rectDef.BorderEnabled,
		BorderWidth:   2 * rectDef.BorderWidth,
		BorderColor:   rectDef.BorderColor,
		Opacity:       rectDef.Opacity,
	}

	content, localBbox, err := rect.Draw(gsName)
	if err != nil {
		return nil, nil, nil, err
	}

	// Bounding box - global page coordinate system (with offset).
	globalBbox := &pdf.PdfRectangle{}
	globalBbox.Llx = rectDef.X + localBbox.Llx
	globalBbox.Lly = rectDef.Y + localBbox.Lly
	globalBbox.Urx = rectDef.X + localBbox.Urx
	globalBbox.Ury = rectDef.Y + localBbox.Ury

	return content, localBbox, globalBbox, nil
}
