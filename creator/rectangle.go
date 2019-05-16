/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

// Rectangle defines a rectangle with upper left corner at (x,y) and a specified width and height.  The rectangle
// can have a colored fill and/or border with a specified width.
// Implements the Drawable interface and can be drawn on PDF using the Creator.
type Rectangle struct {
	x           float64 // Upper left corner
	y           float64
	width       float64
	height      float64
	fillColor   *model.PdfColorDeviceRGB
	borderColor *model.PdfColorDeviceRGB
	borderWidth float64
}

// newRectangle creates a new Rectangle with default parameters with left corner at (x,y) and width, height as specified.
func newRectangle(x, y, width, height float64) *Rectangle {
	rect := &Rectangle{}

	rect.x = x
	rect.y = y
	rect.width = width
	rect.height = height

	rect.borderColor = model.NewPdfColorDeviceRGB(0, 0, 0)
	rect.borderWidth = 1.0

	return rect
}

// GetCoords returns coordinates of the Rectangle's upper left corner (x,y).
func (rect *Rectangle) GetCoords() (float64, float64) {
	return rect.x, rect.y
}

// SetBorderWidth sets the border width.
func (rect *Rectangle) SetBorderWidth(bw float64) {
	rect.borderWidth = bw
}

// SetBorderColor sets border color.
func (rect *Rectangle) SetBorderColor(col Color) {
	rect.borderColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetFillColor sets the fill color.
func (rect *Rectangle) SetFillColor(col Color) {
	rect.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// GeneratePageBlocks draws the rectangle on a new block representing the page. Implements the Drawable interface.
func (rect *Rectangle) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
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
	if rect.borderColor != nil && rect.borderWidth > 0 {
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
