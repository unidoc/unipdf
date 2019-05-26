/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

// Ellipse defines an ellipse with a center at (xc,yc) and a specified width and height.  The ellipse can have a colored
// fill and/or border with a specified width.
// Implements the Drawable interface and can be drawn on PDF using the Creator.
type Ellipse struct {
	xc          float64
	yc          float64
	width       float64
	height      float64
	fillColor   *model.PdfColorDeviceRGB
	borderColor *model.PdfColorDeviceRGB
	borderWidth float64
}

// newEllipse creates a new ellipse centered at (xc,yc) with a width and height specified.
func newEllipse(xc, yc, width, height float64) *Ellipse {
	ell := &Ellipse{}

	ell.xc = xc
	ell.yc = yc
	ell.width = width
	ell.height = height

	ell.borderColor = model.NewPdfColorDeviceRGB(0, 0, 0)
	ell.borderWidth = 1.0

	return ell
}

// GetCoords returns the coordinates of the Ellipse's center (xc,yc).
func (ell *Ellipse) GetCoords() (float64, float64) {
	return ell.xc, ell.yc
}

// SetBorderWidth sets the border width.
func (ell *Ellipse) SetBorderWidth(bw float64) {
	ell.borderWidth = bw
}

// SetBorderColor sets the border color.
func (ell *Ellipse) SetBorderColor(col Color) {
	ell.borderColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetFillColor sets the fill color.
func (ell *Ellipse) SetFillColor(col Color) {
	ell.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// GeneratePageBlocks draws the rectangle on a new block representing the page.
func (ell *Ellipse) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
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
