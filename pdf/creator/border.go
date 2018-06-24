package creator

import (
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	"github.com/unidoc/unidoc/pdf/model"
)

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
	LineStyle         CellBorderLineStyle
}

func NewBorder(x, y, width, height float64, lineStyle CellBorderLineStyle) *Border {
	border := &Border{}

	border.x = x
	border.y = y
	border.width = width
	border.height = height

	border.borderColorTop = model.NewPdfColorDeviceRGB(0, 0, 0)
	border.borderColorBottom = model.NewPdfColorDeviceRGB(0, 0, 0)
	border.borderColorLeft = model.NewPdfColorDeviceRGB(0, 0, 0)
	border.borderColorRight = model.NewPdfColorDeviceRGB(0, 0, 0)
	border.LineStyle = lineStyle
	return border
}

func (border *Border) GetCoords() (float64, float64) {
	return border.x, border.y
}

func (border *Border) SetBorderWidthLeft(bw float64) {
	border.borderWidthLeft = bw
}

func (border *Border) SetBorderColorLeft(col Color) {
	border.borderColorLeft = model.NewPdfColorDeviceRGB(col.ToRGB())
}

func (border *Border) SetBorderWidthBottom(bw float64) {
	border.borderWidthBottom = bw
}

func (border *Border) SetBorderColorBottom(col Color) {
	border.borderColorBottom = model.NewPdfColorDeviceRGB(col.ToRGB())
}

func (border *Border) SetBorderWidthRight(bw float64) {
	border.borderWidthRight = bw
}

func (border *Border) SetBorderColorRight(col Color) {
	border.borderColorRight = model.NewPdfColorDeviceRGB(col.ToRGB())
}

func (border *Border) SetBorderWidthTop(bw float64) {
	border.borderWidthTop = bw
}

func (border *Border) SetBorderColorTop(col Color) {
	border.borderColorTop = model.NewPdfColorDeviceRGB(col.ToRGB())
}

func (border *Border) SetFillColor(col Color) {
	border.fillColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

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

	// Line Top
	var lineTop draw.StraightPath
	if border.LineStyle == CellBorderLineStyleDashed {
		lineTop = draw.DashedLine{
			LineWidth:        border.borderWidthTop,
			Opacity:          1.0,
			LineColor:        border.borderColorTop,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX,
			Y1:               startY,
			X2:               startX + border.width,
			Y2:               startY,
		}
	} else {
		lineTop = draw.Line{
			LineWidth:        border.borderWidthTop,
			Opacity:          1.0,
			LineColor:        border.borderColorTop,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX,
			Y1:               startY,
			X2:               startX + border.width,
			Y2:               startY,
		}
	}
	contentsTop, _, err := lineTop.Draw("")
	if err != nil {
		return nil, ctx, err
	}
	err = block.addContentsByString(string(contentsTop))
	if err != nil {
		return nil, ctx, err
	}

	// Line Left
	var lineLeft draw.StraightPath
	if border.LineStyle == CellBorderLineStyleDashed {
		lineLeft = draw.DashedLine{
			LineWidth:        border.borderWidthLeft,
			Opacity:          1.0,
			LineColor:        border.borderColorLeft,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX,
			Y1:               startY,
			X2:               startX,
			Y2:               startY - border.height,
		}
	} else {
		lineLeft = draw.Line{
			LineWidth:        border.borderWidthLeft,
			Opacity:          1.0,
			LineColor:        border.borderColorLeft,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX,
			Y1:               startY,
			X2:               startX,
			Y2:               startY - border.height,
		}
	}
	contentsLeft, _, err := lineLeft.Draw("")
	if err != nil {
		return nil, ctx, err
	}
	err = block.addContentsByString(string(contentsLeft))
	if err != nil {
		return nil, ctx, err
	}

	// Line Right
	var lineRight draw.StraightPath
	if border.LineStyle == CellBorderLineStyleDashed {
		lineRight = draw.DashedLine{
			LineWidth:        border.borderWidthRight,
			Opacity:          1.0,
			LineColor:        border.borderColorRight,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + border.width,
			Y1:               startY,
			X2:               startX + border.width,
			Y2:               startY - border.height,
		}
	} else {
		lineRight = draw.Line{
			LineWidth:        border.borderWidthRight,
			Opacity:          1.0,
			LineColor:        border.borderColorRight,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + border.width,
			Y1:               startY,
			X2:               startX + border.width,
			Y2:               startY - border.height,
		}
	}
	contentsRight, _, err := lineRight.Draw("")
	if err != nil {
		return nil, ctx, err
	}
	err = block.addContentsByString(string(contentsRight))
	if err != nil {
		return nil, ctx, err
	}

	// Line Bottom
	var lineBottom draw.StraightPath
	if border.LineStyle == CellBorderLineStyleDashed {
		lineBottom = draw.DashedLine{
			LineWidth:        border.borderWidthBottom,
			Opacity:          1.0,
			LineColor:        border.borderColorBottom,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + border.width,
			Y1:               startY - border.height,
			X2:               startX,
			Y2:               startY - border.height,
		}
	} else {
		lineBottom = draw.Line{
			LineWidth:        border.borderWidthBottom,
			Opacity:          1.0,
			LineColor:        border.borderColorBottom,
			LineEndingStyle1: draw.LineEndingStyleNone,
			LineEndingStyle2: draw.LineEndingStyleNone,
			X1:               startX + border.width,
			Y1:               startY - border.height,
			X2:               startX,
			Y2:               startY - border.height,
		}
	}
	contentsBottom, _, err := lineBottom.Draw("")
	if err != nil {
		return nil, ctx, err
	}
	err = block.addContentsByString(string(contentsBottom))
	if err != nil {
		return nil, ctx, err
	}

	return []*Block{block}, ctx, nil
}
