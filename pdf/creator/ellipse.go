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
// Defines an ellipse with a center at (xc,yc) and a specified width and height.  The ellipse can have a colored
// fill and/or border with a specified width.
// Implements the Drawable interface and can be drawn on PDF using the Creator.
//
type ellipse struct {
	xc          float64
	yc          float64
	width       float64
	height      float64
	fillColor   *model.PdfColorDeviceRGB
	borderColor *model.PdfColorDeviceRGB
	borderWidth float64
}

// Generates a new ellipse centered at (xc,yc) with a width and height specified.
func NewEllipse(xc, yc, width, height float64) *ellipse {
	ell := &ellipse{}

	ell.xc = xc
	ell.yc = yc
	ell.width = width
	ell.height = height

	ell.borderColor = model.NewPdfColorDeviceRGB(0, 0, 0)
	ell.borderWidth = 1.0

	return ell
}

// Get the coordinates of the ellipse center (xc,yc).
func (ell *ellipse) GetCoords() (float64, float64) {
	return ell.xc, ell.yc
}

// Set border width.
func (ell *ellipse) SetBorderWidth(bw float64) {
	ell.borderWidth = bw
}

// Set border color.
func (ell *ellipse) SetBorderColor(color rgbColor) {
	ell.borderColor = model.NewPdfColorDeviceRGB(color.r, color.g, color.b)
}

// Set fill color.
func (ell *ellipse) SetFillColor(color rgbColor) {
	ell.fillColor = model.NewPdfColorDeviceRGB(color.r, color.g, color.b)
}

// Draws the rectangle on a new block representing the page.
func (ell *ellipse) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	drawell := draw.Circle{
		X:           ell.xc - ell.width/2,
		Y:           ctx.PageHeight - ell.yc - ell.height/2,
		Width:       ell.width,
		Height:      ell.height,
		Opacity:     1.0,
		BorderWidth: ell.borderWidth,
	}
	if ell.fillColor != nil {
		drawell.FillEnabled = true
		drawell.FillColor = ell.fillColor
	}
	if ell.borderColor != nil {
		drawell.BorderEnabled = true
		drawell.BorderColor = ell.borderColor
		drawell.BorderWidth = ell.borderWidth
	}

	contents, _, err := drawell.Draw("")
	if err != nil {
		return nil, ctx, err
	}

	err = block.addContentsByString(string(contents))
	if err != nil {
		return nil, ctx, err
	}

	return []*Block{block}, ctx, nil
}
