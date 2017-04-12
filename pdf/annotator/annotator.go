package annotator

import (
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

// The currently supported line ending styles are None, Arrow (ClosedArrow) and Butt.
type LineEndingStyle int

const (
	LineEndingStyleNone  LineEndingStyle = 0
	LineEndingStyleArrow LineEndingStyle = 1
	LineEndingStyleButt  LineEndingStyle = 2
)

// Defines a line between point 1 (X1,Y1) and point 2 (X2,Y2).  The line ending styles can be none (regular line),
// or arrows at either end.  The line also has a specified width, color and opacity.
type LineAnnotationDef struct {
	X1               float64
	Y1               float64
	X2               float64
	Y2               float64
	LineColor        *pdf.PdfColorDeviceRGB
	Opacity          float64 // Alpha value (0-1).
	LineWidth        float64
	LineEndingStyle1 LineEndingStyle // Line ending style of point 1.
	LineEndingStyle2 LineEndingStyle // Line ending style of point 2.
}

// Creates a line annotation object that can be added to page PDF annotations.
func CreateLineAnnotation(lineDef LineAnnotationDef) (*pdf.PdfAnnotation, error) {
	// Line annotation.
	lineAnnotation := pdf.NewPdfAnnotationLine()

	// Line endpoint locations.
	lineAnnotation.L = pdfcore.MakeArrayFromFloats([]float64{lineDef.X1, lineDef.Y1, lineDef.X2, lineDef.Y2})

	// Line endings.
	le1 := pdfcore.MakeName("None")
	if lineDef.LineEndingStyle1 == LineEndingStyleArrow {
		le1 = pdfcore.MakeName("ClosedArrow")
	}
	le2 := pdfcore.MakeName("None")
	if lineDef.LineEndingStyle2 == LineEndingStyleArrow {
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

type CircleAnnotationDef struct {
}
