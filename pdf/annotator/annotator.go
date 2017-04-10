package annotator

import (
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

type LineEndingStyle int

const (
	LineEndingStyleNone  LineEndingStyle = 0
	LineEndingStyleArrow LineEndingStyle = 1
	LineEndingStyleButt  LineEndingStyle = 2
)

type LineAnnotationDef struct {
	X1               float64
	Y1               float64
	X2               float64
	Y2               float64
	LineColor        *pdf.PdfColorDeviceRGB
	Opacity          float64
	LineWidth        float64
	LineEndingStyle1 LineEndingStyle
	LineEndingStyle2 LineEndingStyle
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
