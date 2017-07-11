/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	"github.com/unidoc/unidoc/pdf/model"
)

//
// Defines a rectangle with upper left corner at (x,y) and a specified width and height.  The rectangle
// can have a colored fill and/or border with a specified width.
// Implements the Drawable interface and can be drawn on PDF using the Creator.
//
type rectangle struct {
	x           float64 // Upper left corner
	y           float64
	width       float64
	height      float64
	fillColor   *model.PdfColorDeviceRGB
	borderColor *model.PdfColorDeviceRGB
	borderWidth float64
}

// Generate a new line with default parameters between (x1,y1) to (x2,y2).
func NewRectangle(x, y, width, height float64) *rectangle {
	rect := &rectangle{}

	rect.x = x
	rect.y = y
	rect.width = width
	rect.height = height

	rect.borderColor = model.NewPdfColorDeviceRGB(0, 0, 0)
	rect.borderWidth = 1.0

	return rect
}

// Get the coords of the upper left corner (x,y).
func (rect *rectangle) GetCoords() (float64, float64) {
	return rect.x, rect.y
}

// Set border width.
func (rect *rectangle) SetBorderWidth(bw float64) {
	rect.borderWidth = bw
}

// Set border color: r,g,b values from [0-1].
func (rect *rectangle) SetBorderColorRGB(r, g, b float64) {
	rect.borderColor = model.NewPdfColorDeviceRGB(r, g, b)
}

// Set fill color: r,g,b values from [0-1].
func (rect *rectangle) SetFillColorRGB(r, g, b float64) {
	rect.fillColor = model.NewPdfColorDeviceRGB(r, g, b)
}

// Draws the rectangle on a new block representing the page.
func (rect *rectangle) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	drawrect := draw.Rectangle{
		Opacity: 1.0,
		X:       rect.x,
		Y:       ctx.PageHeight - rect.y - rect.height,
		Height:  rect.height,
		Width:   rect.width,
	}
	if rect.fillColor != nil {
		drawrect.FillEnabled = true
		drawrect.FillColor = rect.fillColor
	}
	if rect.borderColor != nil {
		drawrect.BorderEnabled = true
		drawrect.BorderColor = rect.borderColor
		drawrect.BorderWidth = rect.borderWidth
	}

	contents, _, err := drawrect.Draw("")
	if err != nil {
		return nil, ctx, err
	}

	err = block.addContentsByString(string(contents))
	if err != nil {
		return nil, ctx, err
	}

	return []*Block{block}, ctx, nil
}
