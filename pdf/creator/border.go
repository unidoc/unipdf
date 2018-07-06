package creator

import (
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	"github.com/unidoc/unidoc/pdf/model"
)

const (
	doubleBorderAdjustment = 1
)

// border represents cell border
type border struct {
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
	StyleLeft         CellBorderStyle
	StyleRight        CellBorderStyle
	StyleTop          CellBorderStyle
	StyleBottom       CellBorderStyle
}

// newBorder returns and instance of border
func newBorder(x, y, width, height float64) *border {
	border := &border{}

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

	border.LineStyle = draw.LineStyleSolid
	return border
}

// GetCoords returns coordinates of border
func (border *border) GetCoords() (float64, float64) {
	return border.x, border.y
}

// SetWidthLeft sets border width for left
func (border *border) SetWidthLeft(bw float64) {
	border.borderWidthLeft = bw
}

// SetColorLeft sets border color for left
func (border *border) SetColorLeft(col Color) {
	border.borderColorLeft = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthBottom sets border width for bottom
func (border *border) SetWidthBottom(bw float64) {
	border.borderWidthBottom = bw
}

// SetColorBottom sets border color for bottom
func (border *border) SetColorBottom(col Color) {
	border.borderColorBottom = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthRight sets border width for right
func (border *border) SetWidthRight(bw float64) {
	border.borderWidthRight = bw
}

// SetColorRight sets border color for right
func (border *border) SetColorRight(col Color) {
	border.borderColorRight = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthTop sets border width for top
func (border *border) SetWidthTop(bw float64) {
	border.borderWidthTop = bw
}

// SetColorTop sets border color for top
func (border *border) SetColorTop(col Color) {
	border.borderColorTop = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetFillColor sets background color for border
func (border *border) SetFillColor(col Color) {
	border.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// GeneratePageBlocks creates drawable
func (border *border) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)
	startX := border.x
	startY := ctx.PageHeight - border.y

	// Width height adjustment for double border
	autoTopAdjustmentOnLeft := border.borderWidthLeft * doubleBorderAdjustment
	autoTopAdjustmentOnRight := border.borderWidthRight * doubleBorderAdjustment
	autoRightAdjustmentOnTop := border.borderWidthTop * doubleBorderAdjustment
	autoRightAdjustmentOnBottom := border.borderWidthBottom * doubleBorderAdjustment
	autoLeftAdjustmentOnTop := border.borderWidthTop * doubleBorderAdjustment
	autoLeftAdjustmentOnBottom := border.borderWidthBottom * doubleBorderAdjustment
	autoBottomAdjustmentOnLeft := border.borderWidthLeft * doubleBorderAdjustment
	autoBottomAdjustmentOnRight := border.borderWidthRight * doubleBorderAdjustment

	if border.fillColor != nil {
		drawrect := draw.Rectangle{
			Opacity: 1.0,
			X:       border.x + (border.borderWidthLeft * 3) - 0.5,
			Y:       (ctx.PageHeight - border.y - border.height) + (border.borderWidthLeft * 2) - 0.5,
			Height:  border.height - (border.borderWidthBottom * 3),
			Width:   border.width - (border.borderWidthRight * 3),
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
		y1 := startY
		y2 := startY

		if border.StyleTop == CellBorderStyleDoubleTop {
			// Line Top
			lineTop := draw.Line{
				LineWidth:        border.borderWidthTop,
				Opacity:          1.0,
				LineColor:        border.borderColorTop,
				LineEndingStyle1: draw.LineEndingStyleNone,
				LineEndingStyle2: draw.LineEndingStyleNone,
				X1:               startX + autoTopAdjustmentOnLeft + 1, // +1 for corner adjustment
				Y1:               y1 - (doubleBorderAdjustment * border.borderWidthTop),
				X2:               ((startX + border.width) - autoTopAdjustmentOnRight) + 0.5, // +0.5 for corner adjustment
				Y2:               y2 - (doubleBorderAdjustment * border.borderWidthTop),
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

			y1 = y1 + (doubleBorderAdjustment * border.borderWidthTop)
			y2 = y2 + (doubleBorderAdjustment * border.borderWidthTop)
		}

		// Line Top
		lineTop := draw.Line{
			LineWidth:        border.borderWidthTop,
			Opacity:          1.0,
			LineColor:        border.borderColorTop,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + 0.5,
			Y1:               y1,
			X2:               startX + border.width + autoTopAdjustmentOnRight + 0.5,
			Y2:               y2,
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
		x1 := startX + 1
		x2 := startX + 1

		if border.StyleLeft == CellBorderStyleDoubleLeft {
			// Line Left
			lineLeft := draw.Line{
				LineWidth:        border.borderWidthLeft,
				Opacity:          1.0,
				LineColor:        border.borderColorLeft,
				LineEndingStyle1: draw.LineEndingStyleNone,
				LineEndingStyle2: draw.LineEndingStyleNone,
				X1:               x1 - (doubleBorderAdjustment * border.borderWidthLeft),
				Y1:               startY + autoLeftAdjustmentOnTop,
				X2:               x2 - (doubleBorderAdjustment * border.borderWidthLeft),
				Y2:               (startY - border.height) - autoLeftAdjustmentOnBottom,
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

			x1 = x1 + (doubleBorderAdjustment * border.borderWidthLeft)
			x2 = x2 + (doubleBorderAdjustment * border.borderWidthLeft)
		}

		// Line Left
		lineLeft := draw.Line{
			LineWidth:        border.borderWidthLeft,
			Opacity:          1.0,
			LineColor:        border.borderColorLeft,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               x1,
			Y1:               startY,
			X2:               x2,
			Y2:               (startY - border.height) + autoLeftAdjustmentOnBottom,
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
		x1 := startX + border.width
		x2 := startX + border.width

		if border.StyleRight == CellBorderStyleDoubleRight {
			// Line Right
			lineRight := draw.Line{
				LineWidth:        border.borderWidthRight,
				Opacity:          1.0,
				LineColor:        border.borderColorRight,
				LineEndingStyle1: draw.LineEndingStyleNone,
				LineEndingStyle2: draw.LineEndingStyleNone,
				X1:               x1 - (doubleBorderAdjustment * border.borderWidthRight),
				Y1:               startY - autoRightAdjustmentOnTop,
				X2:               x2 - (doubleBorderAdjustment * border.borderWidthRight),
				Y2:               (startY - border.height) + autoRightAdjustmentOnBottom,
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

			x1 = x1 + (doubleBorderAdjustment * border.borderWidthRight)
			x2 = x2 + (doubleBorderAdjustment * border.borderWidthRight)
		}

		// Line Right
		lineRight := draw.Line{
			LineWidth:        border.borderWidthRight,
			Opacity:          1.0,
			LineColor:        border.borderColorRight,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               x1,
			Y1:               startY + autoRightAdjustmentOnTop,
			X2:               x2,
			Y2:               (startY - border.height) - autoRightAdjustmentOnBottom,
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

		y1 := startY - border.height
		y2 := startY - border.height

		if border.StyleBottom == CellBorderStyleDoubleBottom {
			lineBottom := draw.Line{
				LineWidth:        border.borderWidthBottom,
				Opacity:          1.0,
				LineColor:        border.borderColorBottom,
				LineEndingStyle1: draw.LineEndingStyleNone,
				LineEndingStyle2: draw.LineEndingStyleNone,
				X1:               (startX + border.width) - autoBottomAdjustmentOnRight,
				Y1:               y1 + (doubleBorderAdjustment * border.borderWidthBottom),
				X2:               startX + autoBottomAdjustmentOnLeft + 1, // +1 for corner adjustment
				Y2:               y2 + (doubleBorderAdjustment * border.borderWidthBottom),
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

			y1 = y1 - (doubleBorderAdjustment * border.borderWidthBottom)
			y2 = y2 - (doubleBorderAdjustment * border.borderWidthBottom)
		}

		lineBottom := draw.Line{
			LineWidth:        border.borderWidthBottom,
			Opacity:          1.0,
			LineColor:        border.borderColorBottom,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + border.width + autoBottomAdjustmentOnRight,
			Y1:               y1,
			X2:               startX + autoBottomAdjustmentOnRight,
			Y2:               y2,
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
