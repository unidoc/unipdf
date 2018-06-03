/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"
	"strings"

	"github.com/unidoc/unidoc/pdf/model"
)

// NewCurve returns new instance of Curve between points (x1,y1) and (x2, y2) with control point (cx,cy).
func NewCurve(x1, y1, cx, cy, x2, y2 float64) *Curve {
	c := &Curve{}

	c.x1 = x1
	c.y1 = y1

	c.cx = cx
	c.cy = cy

	c.x2 = x2
	c.y2 = y2

	c.lineColor = model.NewPdfColorDeviceRGB(0, 0, 0)
	c.lineWidth = 1.0
	return c
}

// Curve represents a cubic Bezier curve with a control point.
type Curve struct {
	x1 float64
	y1 float64
	cx float64 // control point
	cy float64
	x2 float64
	y2 float64

	lineColor *model.PdfColorDeviceRGB
	lineWidth float64
}

// SetWidth sets line width.
func (c *Curve) SetWidth(width float64) {
	c.lineWidth = width
}

// SetColor sets the line color.
func (c *Curve) SetColor(col Color) {
	c.lineColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// GeneratePageBlocks draws the curve onto page blocks.
func (c *Curve) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	var ops []string
	ops = append(ops, fmt.Sprintf("%.2f w", c.lineWidth))                                               // line widtdh
	ops = append(ops, fmt.Sprintf("%.3f %.3f %.3f RG", c.lineColor[0], c.lineColor[1], c.lineColor[2])) // line color
	ops = append(ops, fmt.Sprintf("%.2f %.2f m", c.x1, ctx.PageHeight-c.y1))                            // move to
	ops = append(ops, fmt.Sprintf("%.5f %.5f %.5f %.5f v S", c.cx, ctx.PageHeight-c.cy, c.x2, ctx.PageHeight-c.y2))

	err := block.addContentsByString(strings.Join(ops, "\n"))
	if err != nil {
		return nil, ctx, err
	}
	return []*Block{block}, ctx, nil
}
