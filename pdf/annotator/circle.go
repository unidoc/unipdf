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

// Make the bezier path with the content creator.
func drawBezierPathWithCreator(bpath draw.CubicBezierPath, creator *pdfcontent.ContentCreator) {
	for idx, c := range bpath.Curves {
		if idx == 0 {
			creator.Add_m(c.P0.X, c.P0.Y)
		}
		creator.Add_c(c.P1.X, c.P1.Y, c.P2.X, c.P2.Y, c.P3.X, c.P3.Y)
	}
}

func makeCircleAnnotationAppearanceStream(circDef CircleAnnotationDef) (*pdfcore.PdfObjectDictionary, *pdf.PdfRectangle, error) {
	form := pdf.NewXObjectForm()
	form.Resources = pdf.NewPdfPageResources()

	gsName := ""
	if circDef.Opacity < 1.0 {
		// Create graphics state with right opacity.
		gsState := &pdfcore.PdfObjectDictionary{}
		gsState.Set("ca", pdfcore.MakeFloat(circDef.Opacity))
		gsState.Set("CA", pdfcore.MakeFloat(circDef.Opacity))
		err := form.Resources.AddExtGState("gs1", gsState)
		if err != nil {
			common.Log.Debug("Unable to add extgstate gs1")
			return nil, nil, err
		}

		gsName = "gs1"
	}

	content, localBbox, globalBbox, err := drawPdfCircle(circDef, gsName)
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

func drawPdfCircle(circDef CircleAnnotationDef, gsName string) ([]byte, *pdf.PdfRectangle, *pdf.PdfRectangle, error) {
	xRad := circDef.Width / 2
	yRad := circDef.Height / 2
	if circDef.BorderEnabled {
		xRad -= circDef.BorderWidth / 2
		yRad -= circDef.BorderWidth / 2
	}

	magic := 0.551784
	xMagic := xRad * magic
	yMagic := yRad * magic

	bpath := draw.NewCubicBezierPath()
	bpath = bpath.AppendCurve(draw.NewCubicBezierCurve(-xRad, 0, -xRad, yMagic, -xMagic, yRad, 0, yRad))
	bpath = bpath.AppendCurve(draw.NewCubicBezierCurve(0, yRad, xMagic, yRad, xRad, yMagic, xRad, 0))
	bpath = bpath.AppendCurve(draw.NewCubicBezierCurve(xRad, 0, xRad, -yMagic, xMagic, -yRad, 0, -yRad))
	bpath = bpath.AppendCurve(draw.NewCubicBezierCurve(0, -yRad, -xMagic, -yRad, -xRad, -yMagic, -xRad, 0))
	bpath = bpath.Offset(xRad, yRad)
	if circDef.BorderEnabled {
		bpath = bpath.Offset(circDef.BorderWidth/2, circDef.BorderWidth/2)
	}

	creator := pdfcontent.NewContentCreator()

	creator.Add_q()

	if circDef.FillEnabled {
		creator.Add_rg(circDef.FillColor.R(), circDef.FillColor.G(), circDef.FillColor.B())
	}
	if circDef.BorderEnabled {
		creator.Add_RG(circDef.BorderColor.R(), circDef.BorderColor.G(), circDef.BorderColor.B())
		creator.Add_w(circDef.BorderWidth)
	}
	if len(gsName) > 1 {
		// If a graphics state is provided, use it. (Used for transparency settings here).
		creator.Add_gs(pdfcore.PdfObjectName(gsName))
	}

	drawBezierPathWithCreator(bpath, creator)

	if circDef.FillEnabled && circDef.BorderEnabled {
		creator.Add_B() // fill and stroke.
	} else if circDef.FillEnabled {
		creator.Add_f() // Fill.
	} else if circDef.BorderEnabled {
		creator.Add_S() // Stroke.
	}
	creator.Add_Q()

	// Offsets (needed for placement of annotations bbox).
	offX := circDef.X
	offY := circDef.Y

	// Get bounding box.
	pathBbox := bpath.GetBoundingBox()
	if circDef.BorderEnabled {
		// Account for stroke width.
		pathBbox.Height += circDef.BorderWidth
		pathBbox.Width += circDef.BorderWidth
		pathBbox.X -= circDef.BorderWidth / 2
		pathBbox.Y -= circDef.BorderWidth / 2
	}

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
