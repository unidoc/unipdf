package creator

import (
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	"github.com/unidoc/unidoc/pdf/model"
)

const (
	gapBetweenDoubleBorder = 2
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

	border.borderWidthTop = 0
	border.borderWidthBottom = 0
	border.borderWidthLeft = 0
	border.borderWidthRight = 0

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

	if border.fillColor != nil {
		drawrect := draw.Rectangle{
			Opacity: 1.0,
			X:       border.x + (border.borderWidthLeft * 3) - 0.5,
			Y:       (ctx.PageHeight - border.y - border.height) + (border.borderWidthLeft * 2),
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
		if border.StyleTop == CellBorderStyleDoubleTop {
			x := startX
			y := startY + (border.borderWidthTop * gapBetweenDoubleBorder)

			lineTop := draw.BasicLine{
				LineWidth: border.borderWidthTop,
				Opacity:   1.0,
				LineColor: border.borderColorTop,
				X1:        x - border.borderWidthLeft*gapBetweenDoubleBorder,
				Y1:        y,
				X2:        x + border.width + (border.borderWidthRight * gapBetweenDoubleBorder) + border.borderWidthRight,
				Y2:        y,
				LineStyle: border.LineStyle,
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

		// Line Top
		lineTop := draw.BasicLine{
			LineWidth: border.borderWidthTop,
			Opacity:   1.0,
			LineColor: border.borderColorTop,
			X1:        startX,
			Y1:        startY,
			X2:        startX + border.width + border.borderWidthRight,
			Y2:        startY,
			LineStyle: border.LineStyle,
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

	if border.borderWidthBottom != 0 {
		x := startX
		y := startY - border.height

		if border.StyleBottom == CellBorderStyleDoubleBottom {
			dx := x
			dy := y - (border.borderWidthBottom * gapBetweenDoubleBorder)

			lineBottom := draw.BasicLine{
				LineWidth: border.borderWidthBottom,
				Opacity:   1.0,
				LineColor: border.borderColorBottom,
				X1:        dx - border.borderWidthLeft*gapBetweenDoubleBorder,
				Y1:        dy,
				X2:        dx + border.width + (border.borderWidthRight * gapBetweenDoubleBorder) + border.borderWidthRight,
				Y2:        dy,
				LineStyle: border.LineStyle,
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

		lineBottom := draw.BasicLine{
			LineWidth: border.borderWidthBottom,
			Opacity:   1.0,
			LineColor: border.borderColorBottom,
			X1:        x,
			Y1:        y,
			X2:        x + border.width + border.borderWidthRight,
			Y2:        y,
			LineStyle: border.LineStyle,
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

	if border.borderWidthLeft != 0 {
		x := startX
		y := startY

		if border.StyleLeft == CellBorderStyleDoubleLeft {
			// Line Left
			lineLeft := draw.BasicLine{
				LineWidth: border.borderWidthLeft,
				Opacity:   1.0,
				LineColor: border.borderColorLeft,
				X1:        x - border.borderWidthLeft*gapBetweenDoubleBorder,
				Y1:        y + border.borderWidthTop + (border.borderWidthTop * 2),
				X2:        x - border.borderWidthLeft*gapBetweenDoubleBorder,
				Y2:        y - border.height - (border.borderWidthTop * 2),
				LineStyle: border.LineStyle,
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

		// Line Left
		lineLeft := draw.BasicLine{
			LineWidth: border.borderWidthLeft,
			Opacity:   1.0,
			LineColor: border.borderColorLeft,
			X1:        x,
			Y1:        y + border.borderWidthTop,
			X2:        x,
			Y2:        y - border.height,
			LineStyle: border.LineStyle,
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
		x := startX + border.width
		y := startY

		if border.StyleRight == CellBorderStyleDoubleRight {
			// Line Right
			lineRight := draw.BasicLine{
				LineWidth: border.borderWidthRight,
				Opacity:   1.0,
				LineColor: border.borderColorRight,
				X1:        x + border.borderWidthRight*gapBetweenDoubleBorder,
				Y1:        y + border.borderWidthTop + (border.borderWidthTop * gapBetweenDoubleBorder),
				X2:        x + border.borderWidthRight*gapBetweenDoubleBorder,
				Y2:        y - border.height - (border.borderWidthTop * gapBetweenDoubleBorder),
				LineStyle: border.LineStyle,
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

		// Line Right
		lineRight := draw.BasicLine{
			LineWidth: border.borderWidthRight,
			Opacity:   1.0,
			LineColor: border.borderColorRight,
			X1:        x,
			Y1:        y + border.borderWidthTop,
			X2:        x,
			Y2:        y - border.height,
			LineStyle: border.LineStyle,
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

	return []*Block{block}, ctx, nil
}
