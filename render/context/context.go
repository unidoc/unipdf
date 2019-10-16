package context

import (
	"image"
	"image/color"

	"github.com/unidoc/unipdf/v3/internal/transform"
)

type Context interface {
	AsMask() *image.Alpha
	Clear()
	ClearPath()
	Clip()
	ClipPreserve()
	ClosePath()
	Matrix() transform.Matrix
	SetMatrix(m transform.Matrix)
	CubicTo(x1, y1, x2, y2, x3, y3 float64)
	DrawArc(x, y, r, angle1, angle2 float64)
	DrawCircle(x, y, r float64)
	DrawEllipse(x, y, rx, ry float64)
	DrawEllipticalArc(x, y, rx, ry, angle1, angle2 float64)
	DrawImage(im image.Image, x, y int)
	DrawImageAnchored(im image.Image, x, y int, ax, ay float64)
	DrawLine(x1, y1, x2, y2 float64)
	DrawPoint(x, y, r float64)
	DrawRectangle(x, y, w, h float64)
	DrawRegularPolygon(n int, x, y, r, rotation float64)
	DrawRoundedRectangle(x, y, w, h, r float64)

	// Text functions.
	DrawString(s string, x, y float64)
	DrawStringAnchored(s string, x, y, ax, ay float64)
	MeasureString(s string) (w, h float64)
	TextState() *TextState

	Fill()
	FillPreserve()
	Height() int
	Identity()
	InvertMask()
	InvertY()
	LineTo(x, y float64)
	LineWidth() float64
	MoveTo(x, y float64)
	NewSubPath()
	Pop()
	Push()
	QuadraticTo(x1, y1, x2, y2 float64)
	ResetClip()
	Rotate(angle float64)
	RotateAbout(angle, x, y float64)
	Scale(x, y float64)
	ScaleAbout(sx, sy, x, y float64)
	SetColor(c color.Color)
	SetDash(dashes ...float64)
	SetFillRule(fillRule FillRule)
	SetFillRuleEvenOdd()
	SetFillRuleWinding()
	SetFillStyle(pattern Pattern)
	SetHexColor(x string)
	SetLineCap(lineCap LineCap)
	SetLineCapButt()
	SetLineCapRound()
	SetLineCapSquare()
	SetLineJoin(lineJoin LineJoin)
	SetLineJoinBevel()
	SetLineJoinRound()
	SetLineWidth(lineWidth float64)
	SetMask(mask *image.Alpha) error
	SetRGB(r, g, b float64)
	SetRGB255(r, g, b int)
	SetRGBA(r, g, b, a float64)
	SetRGBA255(r, g, b, a int)
	SetStrokeRGBA(r, g, b, a float64)
	SetFillRGBA(r, g, b, a float64)
	SetStrokeStyle(pattern Pattern)
	Shear(x, y float64)
	ShearAbout(sx, sy, x, y float64)
	Stroke()
	StrokePreserve()
	Transform(x, y float64) (tx, ty float64)
	Translate(x, y float64)
	Width() int
}
