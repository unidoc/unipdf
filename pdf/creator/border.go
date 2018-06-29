package creator

import (
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	"github.com/unidoc/unidoc/pdf/model"
)

// Border represents cell border
type Border struct {
	x                 float64 // Upper left corner
	y                 float64
	width             float64
	height            float64
	fillColor         *model.PdfColorDeviceRGB
	borderColorLeft   *model.PdfColorDeviceRGB
	borderWidthLeft   float64
	borderColorBottom *model.PdfColorDeviceRGB
	borderWidthBottom float64
	borderColorRight  *model.PdfColorDeviceRGB
	borderWidthRight  float64
	borderColorTop    *model.PdfColorDeviceRGB
	borderWidthTop    float64
	LineStyle         draw.LineStyle
}

// NewBorder returns and instance of border
func NewBorder(x, y, width, height float64) *Border {
	border := &Border{}

	border.x = x
	border.y = y
	border.width = width
	border.height = height

	border.borderColorTop = model.NewPdfColorDeviceRGB(0, 0, 0)
	border.borderColorBottom = model.NewPdfColorDeviceRGB(0, 0, 0)
	border.borderColorLeft = model.NewPdfColorDeviceRGB(0, 0, 0)
	border.borderColorRight = model.NewPdfColorDeviceRGB(0, 0, 0)

	border.borderWidthTop = 1
	border.borderWidthBottom = 1
	border.borderWidthLeft = 1
	border.borderWidthRight = 1

	border.LineStyle = draw.LineStyleDefault
	return border
}

// GetCoords returns coordinates of border
func (border *Border) GetCoords() (float64, float64) {
	return border.x, border.y
}

// SetWidthLeft sets border width for left
func (border *Border) SetWidthLeft(bw float64) {
	border.borderWidthLeft = bw
}

// SetColorLeft sets border color for left
func (border *Border) SetColorLeft(col Color) {
	border.borderColorLeft = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthBottom sets border width for bottom
func (border *Border) SetWidthBottom(bw float64) {
	border.borderWidthBottom = bw
}

// SetColorBottom sets border color for bottom
func (border *Border) SetColorBottom(col Color) {
	border.borderColorBottom = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthRight sets border width for right
func (border *Border) SetWidthRight(bw float64) {
	border.borderWidthRight = bw
}

// SetColorRight sets border color for right
func (border *Border) SetColorRight(col Color) {
	border.borderColorRight = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthTop sets border width for top
func (border *Border) SetWidthTop(bw float64) {
	border.borderWidthTop = bw
}

// SetColorTop sets border color for top
func (border *Border) SetColorTop(col Color) {
	border.borderColorTop = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetFillColor sets background color for border
func (border *Border) SetFillColor(col Color) {
	border.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// GeneratePageBlocks creates drawable
func (border *Border) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)
	startX := border.x
	startY := ctx.PageHeight - border.y

	if border.fillColor != nil {
		drawrect := draw.Rectangle{
			Opacity: 1.0,
			X:       border.x,
			Y:       ctx.PageHeight - border.y - border.height,
			Height:  border.height,
			Width:   border.width,
		}
		drawrect.FillEnabled = true
		drawrect.FillColor = border.fillColor
		drawrect.BorderEnabled = false

		contents, _, err := drawrect.Draw("")
		if err != nil {
			return nil, ctx, err
		}

		err = block.addContentsByString(string(contents))
		if err != nil {
			return nil, ctx, err
		}
	}

	if border.borderWidthTop != 0 {
		// Line Top
		lineTop := draw.Line{
			LineWidth:        border.borderWidthTop,
			Opacity:          1.0,
			LineColor:        border.borderColorTop,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX,
			Y1:               startY,
			X2:               startX + border.width,
			Y2:               startY,
			LineStyle:        border.LineStyle,
		}
		contentsTop, _, err := lineTop.Draw("")
		if err != nil {
			return nil, ctx, err
		}
		err = block.addContentsByString(string(contentsTop))
		if err != nil {
			return nil, ctx, err
		}
	}

	if border.borderWidthLeft != 0 {
		// Line Left
		lineLeft := draw.Line{
			LineWidth:        border.borderWidthLeft,
			Opacity:          1.0,
			LineColor:        border.borderColorLeft,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX,
			Y1:               startY,
			X2:               startX,
			Y2:               startY - border.height,
			LineStyle:        border.LineStyle,
		}
		contentsLeft, _, err := lineLeft.Draw("")
		if err != nil {
			return nil, ctx, err
		}
		err = block.addContentsByString(string(contentsLeft))
		if err != nil {
			return nil, ctx, err
		}
	}

	if border.borderWidthRight != 0 {
		// Line Right
		lineRight := draw.Line{
			LineWidth:        border.borderWidthRight,
			Opacity:          1.0,
			LineColor:        border.borderColorRight,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + border.width,
			Y1:               startY,
			X2:               startX + border.width,
			Y2:               startY - border.height,
			LineStyle:        border.LineStyle,
		}
		contentsRight, _, err := lineRight.Draw("")
		if err != nil {
			return nil, ctx, err
		}
		err = block.addContentsByString(string(contentsRight))
		if err != nil {
			return nil, ctx, err
		}
	}

	if border.borderWidthBottom != 0 {
		// Line Bottom
		lineBottom := draw.Line{
			LineWidth:        border.borderWidthBottom,
			Opacity:          1.0,
			LineColor:        border.borderColorBottom,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + border.width,
			Y1:               startY - border.height,
			X2:               startX,
			Y2:               startY - border.height,
			LineStyle:        border.LineStyle,
		}
		contentsBottom, _, err := lineBottom.Draw("")
		if err != nil {
			return nil, ctx, err
		}
		err = block.addContentsByString(string(contentsBottom))
		if err != nil {
			return nil, ctx, err
		}
	}

	return []*Block{block}, ctx, nil
}
