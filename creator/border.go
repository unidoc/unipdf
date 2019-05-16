package creator

import (
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

// border represents cell border.
type border struct {
	x                 float64 // Upper left corner.
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
	styleLeft         CellBorderStyle
	styleRight        CellBorderStyle
	styleTop          CellBorderStyle
	styleBottom       CellBorderStyle
}

// newBorder returns and instance of border.
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

// GetCoords returns coordinates of border.
func (border *border) GetCoords() (float64, float64) {
	return border.x, border.y
}

// SetWidthLeft sets border width for left.
func (border *border) SetWidthLeft(bw float64) {
	border.borderWidthLeft = bw
}

// SetColorLeft sets border color for left.
func (border *border) SetColorLeft(col Color) {
	border.borderColorLeft = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthBottom sets border width for bottom.
func (border *border) SetWidthBottom(bw float64) {
	border.borderWidthBottom = bw
}

// SetColorBottom sets border color for bottom.
func (border *border) SetColorBottom(col Color) {
	border.borderColorBottom = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthRight sets border width for right.
func (border *border) SetWidthRight(bw float64) {
	border.borderWidthRight = bw
}

// SetColorRight sets border color for right.
func (border *border) SetColorRight(col Color) {
	border.borderColorRight = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetWidthTop sets border width for top.
func (border *border) SetWidthTop(bw float64) {
	border.borderWidthTop = bw
}

// SetColorTop sets border color for top.
func (border *border) SetColorTop(col Color) {
	border.borderColorTop = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetFillColor sets background color for border.
func (border *border) SetFillColor(col Color) {
	border.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetStyleLeft sets border style for left side.
func (border *border) SetStyleLeft(style CellBorderStyle) {
	border.styleLeft = style
}

// SetStyleRight sets border style for right side.
func (border *border) SetStyleRight(style CellBorderStyle) {
	border.styleRight = style
}

// SetStyleTop sets border style for top side.
func (border *border) SetStyleTop(style CellBorderStyle) {
	border.styleTop = style
}

// SetStyleBottom sets border style for bottom side.
func (border *border) SetStyleBottom(style CellBorderStyle) {
	border.styleBottom = style
}

// GeneratePageBlocks implements drawable interface.
func (border *border) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)
	// Start points is in upper left corner.
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

	// a is the spacing between inner and outer line centers (double border only).
	aTop := border.borderWidthTop
	aBottom := border.borderWidthBottom
	aLeft := border.borderWidthLeft
	aRight := border.borderWidthRight

	// wb represents the effective width of border (including gap and double lines in double border).
	wbTop := border.borderWidthTop
	if border.styleTop == CellBorderStyleDouble {
		wbTop += 2 * aTop
	}
	wbBottom := border.borderWidthBottom
	if border.styleBottom == CellBorderStyleDouble {
		wbBottom += 2 * aBottom
	}
	wbLeft := border.borderWidthLeft
	if border.styleLeft == CellBorderStyleDouble {
		wbLeft += 2 * aLeft
	}
	wbRight := border.borderWidthRight
	if border.styleRight == CellBorderStyleDouble {
		wbRight += 2 * aRight
	}

	// Left border.
	if border.borderWidthTop != 0 {
		x := startX
		y := startY

		if border.styleTop == CellBorderStyleDouble {
			y -= aTop

			// Double - Outer line.
			lineTop := draw.BasicLine{}
			lineTop.X1 = x - wbTop/2
			lineTop.Y1 = y + 2*aTop
			lineTop.X2 = x + border.width + wbTop/2
			lineTop.Y2 = y + 2*aTop
			lineTop.LineColor = border.borderColorTop
			lineTop.LineWidth = border.borderWidthTop
			lineTop.LineStyle = border.LineStyle
			contentsTop, _, err := lineTop.Draw("")
			if err != nil {
				return nil, ctx, err
			}
			err = block.addContentsByString(string(contentsTop))
			if err != nil {
				return nil, ctx, err
			}
		}

		lineTop := draw.BasicLine{
			LineWidth: border.borderWidthTop,
			Opacity:   1.0,
			LineColor: border.borderColorTop,
			X1:        x - wbTop/2 + (wbLeft - border.borderWidthLeft),
			Y1:        y,
			X2:        x + border.width + wbTop/2 - (wbRight - border.borderWidthRight),
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

	// Bottom border.
	if border.borderWidthBottom != 0 {
		x := startX
		y := startY - border.height

		if border.styleBottom == CellBorderStyleDouble {
			y += aBottom
			// Double border - Outer line.
			lineBottom := draw.BasicLine{
				LineWidth: border.borderWidthBottom,
				Opacity:   1.0,
				LineColor: border.borderColorBottom,
				X1:        x - wbBottom/2,
				Y1:        y - 2*aBottom,
				X2:        x + border.width + wbBottom/2,
				Y2:        y - 2*aBottom,
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
			X1:        x - wbBottom/2 + (wbLeft - border.borderWidthLeft),
			Y1:        y,
			X2:        x + border.width + wbBottom/2 - (wbRight - border.borderWidthRight),
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

	// Left border.
	if border.borderWidthLeft != 0 {
		x := startX
		y := startY

		if border.styleLeft == CellBorderStyleDouble {
			x += aLeft

			// Double border - outer line.
			lineLeft := draw.BasicLine{
				LineWidth: border.borderWidthLeft,
				Opacity:   1.0,
				LineColor: border.borderColorLeft,
				X1:        x - 2*aLeft,
				Y1:        y + wbLeft/2,
				X2:        x - 2*aLeft,
				Y2:        y - border.height - wbLeft/2,
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

		// Line Left.
		lineLeft := draw.BasicLine{
			LineWidth: border.borderWidthLeft,
			Opacity:   1.0,
			LineColor: border.borderColorLeft,
			X1:        x,
			Y1:        y + wbLeft/2 - (wbTop - border.borderWidthTop),
			X2:        x,
			Y2:        y - border.height - wbLeft/2 + (wbBottom - border.borderWidthBottom),
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

	// Right border.
	if border.borderWidthRight != 0 {
		x := startX + border.width
		y := startY

		if border.styleRight == CellBorderStyleDouble {
			x -= aRight

			// Double border - Outer line.
			lineRight := draw.BasicLine{
				LineWidth: border.borderWidthRight,
				Opacity:   1.0,
				LineColor: border.borderColorRight,
				X1:        x + 2*aRight,
				Y1:        y + wbRight/2,
				X2:        x + 2*aRight,
				Y2:        y - border.height - wbRight/2,
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

		lineRight := draw.BasicLine{
			LineWidth: border.borderWidthRight,
			Opacity:   1.0,
			LineColor: border.borderColorRight,
			X1:        x,
			Y1:        y + wbRight/2 - (wbTop - border.borderWidthTop),
			X2:        x,
			Y2:        y - border.height - wbRight/2 + (wbBottom - border.borderWidthBottom),
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
