/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

type Polygon struct {
	points               []Point
	fillColor            *model.PdfColorDeviceRGB
	fillOpacity          float64
	fillOpacityEnabled   bool
	borderColor          *model.PdfColorDeviceRGB
	borderWidth          float64
	borderOpacityEnabled bool
	borderOpacity        float64
}

// newPolygon creates a new Polygon with default parameters.
func newPolygon(points []Point) *Polygon {
	polygon := &Polygon{}
	polygon.points = points
	return polygon
}

// SetBorderWidth sets the border width.
func (polygon *Polygon) SetBorderWidth(bw float64) {
	polygon.borderWidth = bw
}

// SetBorderColor sets border color.
func (polygon *Polygon) SetBorderColor(col Color) {
	polygon.borderColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetBorderOpacity sets the border opacity.
func (polygon *Polygon) SetBorderOpacity(opacity float64) {
	polygon.borderOpacityEnabled = true
	polygon.borderOpacity = opacity
}

// SetFillColor sets the fill color.
func (polygon *Polygon) SetFillColor(col Color) {
	polygon.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetFillOpacity sets the fill opacity.
func (polygon *Polygon) SetFillOpacity(opacity float64) {
	polygon.fillOpacityEnabled = true
	polygon.fillOpacity = opacity
}

// GeneratePageBlocks draws the polygon on a new block representing the page. Implements the Drawable interface.
func (polygon *Polygon) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	drawPolygon := draw.Polygon{
		Points: []draw.Point{},
	}
	for _, p := range polygon.points {
		drawPolygon.Points = append(drawPolygon.Points, draw.NewPoint(p.X, p.Y))
	}
	if polygon.fillColor != nil {
		drawPolygon.FillEnabled = true
		drawPolygon.FillColor = polygon.fillColor
	}
	if polygon.borderColor != nil && polygon.borderWidth > 0 {
		drawPolygon.BorderEnabled = true
		drawPolygon.BorderColor = polygon.borderColor
		drawPolygon.BorderWidth = polygon.borderWidth
	}

	if !polygon.fillOpacityEnabled {
		polygon.fillOpacity = 1.0
	}
	if !polygon.borderOpacityEnabled {
		polygon.borderOpacity = 1.0
	}
	gsName, err := block.setOpacity(polygon.fillOpacity, polygon.borderOpacity)
	if err != nil {
		return nil, ctx, err
	}

	contents, _, err := drawPolygon.Draw(gsName)
	if err != nil {
		return nil, ctx, err
	}

	err = block.addContentsByString(string(contents))
	if err != nil {
		return nil, ctx, err
	}

	return []*Block{block}, ctx, nil
}
