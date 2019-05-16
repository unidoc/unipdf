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

// LineAnnotationDef defines a line between point 1 (X1,Y1) and point 2 (X2,Y2).  The line ending styles can be none
// (regular line), or arrows at either end.  The line also has a specified width, color and opacity.
type LineAnnotationDef struct {
	X1               float64
	Y1               float64
	X2               float64
	Y2               float64
	LineColor        *pdf.PdfColorDeviceRGB
	Opacity          float64 // Alpha value (0-1).
	LineWidth        float64
	LineEndingStyle1 draw.LineEndingStyle // Line ending style of point 1.
	LineEndingStyle2 draw.LineEndingStyle // Line ending style of point 2.
}

// CreateLineAnnotation creates a line annotation object that can be added to page PDF annotations.
func CreateLineAnnotation(lineDef LineAnnotationDef) (*pdf.PdfAnnotation, error) {
	// Line annotation.
	lineAnnotation := pdf.NewPdfAnnotationLine()

	// Line endpoint locations.
	lineAnnotation.L = pdfcore.MakeArrayFromFloats([]float64{lineDef.X1, lineDef.Y1, lineDef.X2, lineDef.Y2})

	// Line endings.
	le1 := pdfcore.MakeName("None")
	if lineDef.LineEndingStyle1 == draw.LineEndingStyleArrow {
		le1 = pdfcore.MakeName("ClosedArrow")
	}
	le2 := pdfcore.MakeName("None")
	if lineDef.LineEndingStyle2 == draw.LineEndingStyleArrow {
		le2 = pdfcore.MakeName("ClosedArrow")
	}
	lineAnnotation.LE = pdfcore.MakeArray(le1, le2)

	// Opacity.
	if lineDef.Opacity < 1.0 {
		lineAnnotation.CA = pdfcore.MakeFloat(lineDef.Opacity)
	}

	r, g, b := lineDef.LineColor.R(), lineDef.LineColor.G(), lineDef.LineColor.B()
	lineAnnotation.IC = pdfcore.MakeArrayFromFloats([]float64{r, g, b}) // fill color of line endings, rgb 0-1.
	lineAnnotation.C = pdfcore.MakeArrayFromFloats([]float64{r, g, b})  // line color, rgb 0-1.
	bs := pdf.NewBorderStyle()
	bs.SetBorderWidth(lineDef.LineWidth) // Line width: 3 points.
	lineAnnotation.BS = bs.ToPdfObject()

	// Make the appearance stream (for uniform appearance).
	apDict, bbox, err := makeLineAnnotationAppearanceStream(lineDef)
	if err != nil {
		return nil, err
	}
	lineAnnotation.AP = apDict

	// The rect specifies the location and dimensions of the annotation.  Technically if the annotation could not
	// be displayed if it goes outside these bounds, although rarely enforced.
	lineAnnotation.Rect = pdfcore.MakeArrayFromFloats([]float64{bbox.Llx, bbox.Lly, bbox.Urx, bbox.Ury})

	return lineAnnotation.PdfAnnotation, nil
}

func makeLineAnnotationAppearanceStream(lineDef LineAnnotationDef) (*pdfcore.PdfObjectDictionary, *pdf.PdfRectangle, error) {
	form := pdf.NewXObjectForm()
	form.Resources = pdf.NewPdfPageResources()

	gsName := ""
	if lineDef.Opacity < 1.0 {
		// Create graphics state with right opacity.
		gsState := pdfcore.MakeDict()
		gsState.Set("ca", pdfcore.MakeFloat(lineDef.Opacity))
		err := form.Resources.AddExtGState("gs1", gsState)
		if err != nil {
			common.Log.Debug("Unable to add extgstate gs1")
			return nil, nil, err
		}

		gsName = "gs1"
	}

	content, localBbox, globalBbox, err := drawPdfLine(lineDef, gsName)
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

func drawPdfLine(lineDef LineAnnotationDef, gsName string) ([]byte, *pdf.PdfRectangle, *pdf.PdfRectangle, error) {
	// The annotation is drawn locally in a relative coordinate system with 0,0 as the origin rather than an offset.
	line := draw.Line{
		X1:               0,
		Y1:               0,
		X2:               lineDef.X2 - lineDef.X1,
		Y2:               lineDef.Y2 - lineDef.Y1,
		LineColor:        lineDef.LineColor,
		Opacity:          lineDef.Opacity,
		LineWidth:        lineDef.LineWidth,
		LineEndingStyle1: lineDef.LineEndingStyle1,
		LineEndingStyle2: lineDef.LineEndingStyle2,
	}

	content, localBbox, err := line.Draw(gsName)
	if err != nil {
		return nil, nil, nil, err
	}

	// Bounding box - global page coordinate system (with offset).
	globalBbox := &pdf.PdfRectangle{}
	globalBbox.Llx = lineDef.X1 + localBbox.Llx
	globalBbox.Lly = lineDef.Y1 + localBbox.Lly
	globalBbox.Urx = lineDef.X1 + localBbox.Urx
	globalBbox.Ury = lineDef.Y1 + localBbox.Ury

	return content, localBbox, globalBbox, nil
}
