/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package annotator

import (
	"github.com/unidoc/unidoc/common"

	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

// A rectangle defined with a specified Width and Height and a lower left corner at (X,Y).  The rectangle can
// optionally have a border and a filling color.
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

// Creates a rectangle annotation object that can be added to page PDF annotations.
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
		gsState := &pdfcore.PdfObjectDictionary{}
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

	apDict := &pdfcore.PdfObjectDictionary{}
	apDict.Set("N", form.ToPdfObject())

	return apDict, globalBbox, nil
}

func drawPdfRectangle(rectDef RectangleAnnotationDef, gsName string) ([]byte, *pdf.PdfRectangle, *pdf.PdfRectangle, error) {
	path := draw.NewPath()

	path = path.AppendPoint(draw.NewPoint(0, 0))
	path = path.AppendPoint(draw.NewPoint(0, rectDef.Height))
	path = path.AppendPoint(draw.NewPoint(rectDef.Width, rectDef.Height))
	path = path.AppendPoint(draw.NewPoint(rectDef.Width, 0))
	path = path.AppendPoint(draw.NewPoint(0, 0))

	creator := pdfcontent.NewContentCreator()

	creator.Add_q()
	if rectDef.FillEnabled {
		creator.Add_rg(rectDef.FillColor.R(), rectDef.FillColor.G(), rectDef.FillColor.B())
	}
	if rectDef.BorderEnabled {
		creator.Add_RG(rectDef.BorderColor.R(), rectDef.BorderColor.G(), rectDef.BorderColor.B())
		creator.Add_w(rectDef.BorderWidth)
	}
	if len(gsName) > 1 {
		// If a graphics state is provided, use it. (Used for transparency settings here).
		creator.Add_gs(pdfcore.PdfObjectName(gsName))
	}
	drawPathWithCreator(path, creator)

	if rectDef.FillEnabled && rectDef.BorderEnabled {
		creator.Add_B() // fill and stroke.
	} else if rectDef.FillEnabled {
		creator.Add_f() // Fill.
	} else if rectDef.BorderEnabled {
		creator.Add_S() // Stroke.
	}
	creator.Add_Q()

	// Offsets (needed for placement of annotations bbox).
	offX := rectDef.X
	offY := rectDef.Y

	// Get bounding box.
	pathBbox := path.GetBoundingBox()

	// Bounding box - local coordinate system (without offset).
	localBbox := &pdf.PdfRectangle{}
	localBbox.Llx = pathBbox.X
	localBbox.Lly = pathBbox.Y
	localBbox.Urx = pathBbox.X + pathBbox.Width
	localBbox.Ury = pathBbox.Y + pathBbox.Height

	// Bounding box - global page coordinate system (with offset).
	globalBbox := &pdf.PdfRectangle{}
	globalBbox.Llx = offX + pathBbox.X
	globalBbox.Lly = offY + pathBbox.Y
	globalBbox.Urx = offX + pathBbox.X + pathBbox.Width
	globalBbox.Ury = offY + pathBbox.Y + pathBbox.Height

	return creator.Bytes(), localBbox, globalBbox, nil
}
