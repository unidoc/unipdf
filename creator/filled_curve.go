/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	pdfcontent "github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	pdfcore "github.com/unidoc/unipdf/v3/core"
	pdf "github.com/unidoc/unipdf/v3/model"
)

// FilledCurve represents a closed path of Bezier curves with a border and fill.
type FilledCurve struct {
	curves        []draw.CubicBezierCurve
	FillEnabled   bool // Show fill?
	fillColor     *pdf.PdfColorDeviceRGB
	BorderEnabled bool // Show border?
	BorderWidth   float64
	borderColor   *pdf.PdfColorDeviceRGB
}

// newFilledCurve returns a instance of filled curve.
func newFilledCurve() *FilledCurve {
	curve := FilledCurve{}
	curve.curves = []draw.CubicBezierCurve{}
	return &curve
}

// AppendCurve appends a Bezier curve to the filled curve.
func (fc *FilledCurve) AppendCurve(curve draw.CubicBezierCurve) *FilledCurve {
	fc.curves = append(fc.curves, curve)
	return fc
}

// SetFillColor sets the fill color for the path.
func (fc *FilledCurve) SetFillColor(color Color) {
	fc.fillColor = pdf.NewPdfColorDeviceRGB(color.ToRGB())
}

// SetBorderColor sets the border color for the path.
func (fc *FilledCurve) SetBorderColor(color Color) {
	fc.borderColor = pdf.NewPdfColorDeviceRGB(color.ToRGB())
}

// draw draws the filled curve. Can specify a graphics state (gsName) for setting opacity etc. Otherwise leave empty ("").
// Returns the content stream as a byte array, the bounding box and an error on failure.
func (fc *FilledCurve) draw(gsName string) ([]byte, *pdf.PdfRectangle, error) {
	bpath := draw.NewCubicBezierPath()
	for _, c := range fc.curves {
		bpath = bpath.AppendCurve(c)
	}

	creator := pdfcontent.NewContentCreator()
	creator.Add_q()

	if fc.FillEnabled {
		creator.Add_rg(fc.fillColor.R(), fc.fillColor.G(), fc.fillColor.B())
	}
	if fc.BorderEnabled {
		creator.Add_RG(fc.borderColor.R(), fc.borderColor.G(), fc.borderColor.B())
		creator.Add_w(fc.BorderWidth)
	}
	if len(gsName) > 1 {
		// If a graphics state is provided, use it. (can support transparency).
		creator.Add_gs(pdfcore.PdfObjectName(gsName))
	}

	draw.DrawBezierPathWithCreator(bpath, creator)
	creator.Add_h() // Close the path.

	if fc.FillEnabled && fc.BorderEnabled {
		creator.Add_B() // fill and stroke.
	} else if fc.FillEnabled {
		creator.Add_f() // Fill.
	} else if fc.BorderEnabled {
		creator.Add_S() // Stroke.
	}
	creator.Add_Q()

	// Get bounding box.
	pathBbox := bpath.GetBoundingBox()
	if fc.BorderEnabled {
		// Account for stroke width.
		pathBbox.Height += fc.BorderWidth
		pathBbox.Width += fc.BorderWidth
		pathBbox.X -= fc.BorderWidth / 2
		pathBbox.Y -= fc.BorderWidth / 2
	}

	// Bounding box - global coordinate system.
	bbox := &pdf.PdfRectangle{}
	bbox.Llx = pathBbox.X
	bbox.Lly = pathBbox.Y
	bbox.Urx = pathBbox.X + pathBbox.Width
	bbox.Ury = pathBbox.Y + pathBbox.Height
	return creator.Bytes(), bbox, nil
}

// GeneratePageBlocks draws the filled curve on page blocks.
func (fc *FilledCurve) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	contents, _, err := fc.draw("")
	err = block.addContentsByString(string(contents))
	if err != nil {
		return nil, ctx, err
	}
	return []*Block{block}, ctx, nil
}
