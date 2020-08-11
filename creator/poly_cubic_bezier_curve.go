/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

// PolyCubicBezierCurve represents a curve that is the result of joining several bezier curves
// Implements the Drawable interface and can be drawn on PDF using the Creator.
type PolyCubicBezierCurve struct {
	curves               []CubicBezierCurve
	fillColor            *model.PdfColorDeviceRGB
	fillOpacity          float64
	fillOpacityEnabled   bool
	borderColor          *model.PdfColorDeviceRGB
	borderWidth          float64
	borderOpacityEnabled bool
	borderOpacity        float64
}

// newPolyCubicBezierCurve creates a new PolyCubicBezierCurve with default parameters.
func newPolyCubicBezierCurve(curves []CubicBezierCurve) *PolyCubicBezierCurve {
	polyCubicBezierCurve := &PolyCubicBezierCurve{}
	polyCubicBezierCurve.curves = curves
	polyCubicBezierCurve.borderColor = model.NewPdfColorDeviceRGB(0, 0, 0)
	polyCubicBezierCurve.borderWidth = 1.0
	return polyCubicBezierCurve
}

// SetBorderWidth sets the border width.
func (polyCubicBezierCurve *PolyCubicBezierCurve) SetBorderWidth(bw float64) {
	polyCubicBezierCurve.borderWidth = bw
}

// SetBorderColor sets border color.
func (polyCubicBezierCurve *PolyCubicBezierCurve) SetBorderColor(col Color) {
	polyCubicBezierCurve.borderColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetBorderOpacity sets the border opacity.
func (polyCubicBezierCurve *PolyCubicBezierCurve) SetBorderOpacity(opacity float64) {
	polyCubicBezierCurve.borderOpacityEnabled = true
	polyCubicBezierCurve.borderOpacity = opacity
}

// SetFillColor sets the fill color.
func (polyCubicBezierCurve *PolyCubicBezierCurve) SetFillColor(col Color) {
	polyCubicBezierCurve.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetFillOpacity sets the fill opacity.
func (polyCubicBezierCurve *PolyCubicBezierCurve) SetFillOpacity(opacity float64) {
	polyCubicBezierCurve.fillOpacityEnabled = true
	polyCubicBezierCurve.fillOpacity = opacity
}

// GeneratePageBlocks draws the polyCubicBezierCurve on a new block representing the page. Implements the Drawable interface.
func (polyCubicBezierCurve *PolyCubicBezierCurve) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	drawPolyCubicBezierCurve := draw.PolyCubicBezierCurve{
		Curves: []draw.CubicBezierCurve{},
	}
	for _, c := range polyCubicBezierCurve.curves {
		drawPolyCubicBezierCurve.Curves = append(drawPolyCubicBezierCurve.Curves, draw.NewCubicBezierCurve(c.P0.X, c.P0.Y, c.P1.X, c.P1.Y, c.P2.X, c.P2.Y, c.P3.X, c.P3.Y))
	}

	if polyCubicBezierCurve.fillColor != nil {
		drawPolyCubicBezierCurve.FillEnabled = true
		drawPolyCubicBezierCurve.FillColor = polyCubicBezierCurve.fillColor
	}
	drawPolyCubicBezierCurve.BorderColor = polyCubicBezierCurve.borderColor
	drawPolyCubicBezierCurve.BorderWidth = polyCubicBezierCurve.borderWidth

	if !polyCubicBezierCurve.fillOpacityEnabled {
		polyCubicBezierCurve.fillOpacity = 1.0
	}
	if !polyCubicBezierCurve.borderOpacityEnabled {
		polyCubicBezierCurve.borderOpacity = 1.0
	}
	gsName, err := block.setOpacity(polyCubicBezierCurve.fillOpacity, polyCubicBezierCurve.borderOpacity)
	if err != nil {
		return nil, ctx, err
	}

	contents, _, err := drawPolyCubicBezierCurve.Draw(gsName)
	if err != nil {
		return nil, ctx, err
	}

	err = block.addContentsByString(string(contents))
	if err != nil {
		return nil, ctx, err
	}

	return []*Block{block}, ctx, nil
}
