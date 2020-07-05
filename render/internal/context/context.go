/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package context

import (
	"image"

	"github.com/unidoc/unipdf/v3/internal/transform"
)

// Context defines operations for rendering to a particular target.
type Context interface {
	//
	// Graphics state operations
	//

	// Push adds the current context state on the stack.
	Push()

	// Pop removes the most recent context state from the stack.
	Pop()

	//
	// Matrix operations
	//

	// Matrix returns the current transformation matrix.
	Matrix() transform.Matrix

	// SetMatrix modifies the transformation matrix.
	SetMatrix(m transform.Matrix)

	// Translate updates the current matrix with a translation.
	Translate(x, y float64)

	// Scale updates the current matrix with a scaling factor.
	// Scaling occurs about the origin.
	Scale(x, y float64)

	// Rotate updates the current matrix with a anticlockwise rotation.
	// Rotation occurs about the origin. Angle is specified in radians.
	Rotate(angle float64)

	//
	// Path operations
	//

	// MoveTo starts a new subpath within the current path starting at
	// the specified point.
	MoveTo(x, y float64)

	// LineTo adds a line segment to the current path starting at the current
	// point.
	LineTo(x, y float64)

	// CubicTo adds a cubic bezier curve to the current path starting at the
	// current point.
	CubicTo(x1, y1, x2, y2, x3, y3 float64)

	// QuadraticTo adds a quadratic bezier curve to the current path starting
	// at the current point.
	QuadraticTo(x1, y1, x2, y2 float64)

	// NewSubPath starts a new subpath within the current path.
	NewSubPath()

	// ClosePath adds a line segment from the current point to the beginning
	// of the current subpath.
	ClosePath()

	// ClearPath clears the current path.
	ClearPath()

	// Clip updates the clipping region by intersecting the current
	// clipping region with the current path as it would be filled by Fill().
	// The path is cleared after this operation.
	Clip()

	// ClipPreserve updates the clipping region by intersecting the current
	// clipping region with the current path as it would be filled by Fill().
	// The path is preserved after this operation.
	ClipPreserve()

	// ResetClip clears the clipping region.
	ResetClip()

	//
	// Line style operations
	//

	// LineWidth returns the current line width.
	LineWidth() float64

	// SetLineWidth sets the line width.
	SetLineWidth(lineWidth float64)

	// SetLineCap sets the line cap style.
	SetLineCap(lineCap LineCap)

	// SetLineJoin sets the line join style.
	SetLineJoin(lineJoin LineJoin)

	// SetDash sets the line dash pattern.
	SetDash(dashes ...float64)

	// SetDashOffset sets the initial offset into the dash pattern to use when
	// stroking dashed paths.
	SetDashOffset(offset float64)

	//
	// Fill and stroke operations
	//

	// Fill fills the current path with the current color. Open subpaths
	// are implicity closed.
	Fill()

	// FillPreserve fills the current path with the current color. Open subpaths
	// are implicity closed. The path is preserved after this operation.
	FillPreserve()

	// Stroke strokes the current path with the current color, line width,
	// line cap, line join and dash settings. The path is cleared after this
	// operation.
	Stroke()

	// StrokePreserve strokes the current path with the current color,
	// line width, line cap, line join and dash settings. The path is preserved
	// after this operation.
	StrokePreserve()

	// SetNRGBA sets the both the fill and stroke colors.
	// r, g, b, a values should be in range 0-1.
	SetRGBA(r, g, b, a float64)

	// SetNRGBA sets the fill color.
	// r, g, b, a values should be in range 0-1.
	SetFillRGBA(r, g, b, a float64)

	// SetStrokeStyle sets current fill pattern.
	SetFillStyle(pattern Pattern)

	// SetFillRule sets the fill rule.
	SetFillRule(fillRule FillRule)

	// SetNRGBA sets the stroke color.
	// r, g, b, a values should be in range 0-1.
	SetStrokeRGBA(r, g, b, a float64)

	// SetStrokeStyle sets current stroke pattern.
	SetStrokeStyle(pattern Pattern)

	//
	// Text operations
	//

	// TextState returns the current text state.
	TextState() *TextState

	// DrawString renders the specified string and the specified position.
	DrawString(s string, x, y float64)

	// Measure string returns the width and height of the specified string.
	MeasureString(s string) (w, h float64)

	//
	// Draw operations
	//

	// DrawRectangle draws the specified rectangle.
	DrawRectangle(x, y, w, h float64)

	// DrawImage draws the specified image at the specified point.
	DrawImage(image image.Image, x, y int)

	// DrawImageAnchored draws the specified image at the specified anchor point.
	// The anchor point is x - w * ax, y - h * ay, where w, h is the size of the
	// image. Use ax=0.5, ay=0.5 to center the image at the specified point.
	DrawImageAnchored(image image.Image, x, y int, ax, ay float64)

	//
	// Misc operations
	//

	// Width returns the width of the rendering area.
	Height() int

	// Height returns the height of the rendering area.
	Width() int
}
