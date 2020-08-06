/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

// Polyline defines a slice of points that are connected as straight lines.
// Implements the Drawable interface and can be drawn on PDF using the Creator.
type Polyline struct {
	points             []Point
	lineColor          *model.PdfColorDeviceRGB
	lineWidth          float64
	lineOpacityEnabled bool
	lineOpacity        float64
}

// newPolyline creates a new Polyline with default parameters.
func newPolyline(points []Point) *Polyline {
	polyline := &Polyline{}
	polyline.points = points
	polyline.lineColor = model.NewPdfColorDeviceRGB(0, 0, 0)
	polyline.lineWidth = 1.0
	return polyline
}

// SetLineWidth sets the line width.
func (polyline *Polyline) SetLineWidth(bw float64) {
	polyline.lineWidth = bw
}

// SetLineColor sets line color.
func (polyline *Polyline) SetLineColor(col Color) {
	polyline.lineColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetLineOpacity sets the line opacity.
func (polyline *Polyline) SetLineOpacity(opacity float64) {
	polyline.lineOpacityEnabled = true
	polyline.lineOpacity = opacity
}

// GeneratePageBlocks draws the polyline on a new block representing the page. Implements the Drawable interface.
func (polyline *Polyline) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	drawPolyline := draw.Polyline{
		Points: []draw.Point{},
	}
	for _, p := range polyline.points {
		drawPolyline.Points = append(drawPolyline.Points, draw.NewPoint(p.X, p.Y))
	}
	drawPolyline.LineColor = polyline.lineColor
	drawPolyline.LineWidth = polyline.lineWidth
	if !polyline.lineOpacityEnabled {
		polyline.lineOpacity = 1.0
	}
	gsName, err := block.setOpacity(polyline.lineOpacity, polyline.lineOpacity)
	if err != nil {
		return nil, ctx, err
	}

	contents, _, err := drawPolyline.Draw(gsName)
	if err != nil {
		return nil, ctx, err
	}

	err = block.addContentsByString(string(contents))
	if err != nil {
		return nil, ctx, err
	}

	return []*Block{block}, ctx, nil
}
