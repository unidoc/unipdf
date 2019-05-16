/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

import (
	"math"

	"github.com/unidoc/unipdf/v3/model"
)

// CubicBezierCurve is defined by:
// R(t) = P0*(1-t)^3 + P1*3*t*(1-t)^2 + P2*3*t^2*(1-t) + P3*t^3
// where P0 is the current point, P1, P2 control points and P3 the final point.
type CubicBezierCurve struct {
	P0 Point // Starting point.
	P1 Point // Control point 1.
	P2 Point // Control point 2.
	P3 Point // Final point.
}

// NewCubicBezierCurve returns a new cubic Bezier curve.
func NewCubicBezierCurve(x0, y0, x1, y1, x2, y2, x3, y3 float64) CubicBezierCurve {
	curve := CubicBezierCurve{}
	curve.P0 = NewPoint(x0, y0)
	curve.P1 = NewPoint(x1, y1)
	curve.P2 = NewPoint(x2, y2)
	curve.P3 = NewPoint(x3, y3)
	return curve
}

// AddOffsetXY adds X,Y offset to all points on a curve.
func (curve CubicBezierCurve) AddOffsetXY(offX, offY float64) CubicBezierCurve {
	curve.P0.X += offX
	curve.P1.X += offX
	curve.P2.X += offX
	curve.P3.X += offX

	curve.P0.Y += offY
	curve.P1.Y += offY
	curve.P2.Y += offY
	curve.P3.Y += offY

	return curve
}

// GetBounds returns the bounding box of the Bezier curve.
func (curve CubicBezierCurve) GetBounds() model.PdfRectangle {
	minX := curve.P0.X
	maxX := curve.P0.X
	minY := curve.P0.Y
	maxY := curve.P0.Y

	// 1000 points.
	for t := 0.0; t <= 1.0; t += 0.001 {
		Rx := curve.P0.X*math.Pow(1-t, 3) +
			curve.P1.X*3*t*math.Pow(1-t, 2) +
			curve.P2.X*3*math.Pow(t, 2)*(1-t) +
			curve.P3.X*math.Pow(t, 3)
		Ry := curve.P0.Y*math.Pow(1-t, 3) +
			curve.P1.Y*3*t*math.Pow(1-t, 2) +
			curve.P2.Y*3*math.Pow(t, 2)*(1-t) +
			curve.P3.Y*math.Pow(t, 3)

		if Rx < minX {
			minX = Rx
		}
		if Rx > maxX {
			maxX = Rx
		}
		if Ry < minY {
			minY = Ry
		}
		if Ry > maxY {
			maxY = Ry
		}
	}

	bounds := model.PdfRectangle{}
	bounds.Llx = minX
	bounds.Lly = minY
	bounds.Urx = maxX
	bounds.Ury = maxY
	return bounds
}

// CubicBezierPath represents a collection of cubic Bezier curves.
type CubicBezierPath struct {
	Curves []CubicBezierCurve
}

// NewCubicBezierPath returns a new empty cubic Bezier path.
func NewCubicBezierPath() CubicBezierPath {
	bpath := CubicBezierPath{}
	bpath.Curves = []CubicBezierCurve{}
	return bpath
}

// AppendCurve appends the specified Bezier curve to the path.
func (p CubicBezierPath) AppendCurve(curve CubicBezierCurve) CubicBezierPath {
	p.Curves = append(p.Curves, curve)
	return p
}

// Copy returns a clone of the Bezier path.
func (p CubicBezierPath) Copy() CubicBezierPath {
	bpathcopy := CubicBezierPath{}
	bpathcopy.Curves = []CubicBezierCurve{}
	for _, c := range p.Curves {
		bpathcopy.Curves = append(bpathcopy.Curves, c)
	}
	return bpathcopy
}

// Offset shifts the Bezier path with the specified offsets.
func (p CubicBezierPath) Offset(offX, offY float64) CubicBezierPath {
	for i, c := range p.Curves {
		p.Curves[i] = c.AddOffsetXY(offX, offY)
	}
	return p
}

// GetBoundingBox returns the bounding box of the Bezier path.
func (p CubicBezierPath) GetBoundingBox() Rectangle {
	bbox := Rectangle{}

	minX := 0.0
	maxX := 0.0
	minY := 0.0
	maxY := 0.0
	for idx, c := range p.Curves {
		curveBounds := c.GetBounds()
		if idx == 0 {
			minX = curveBounds.Llx
			maxX = curveBounds.Urx
			minY = curveBounds.Lly
			maxY = curveBounds.Ury
			continue
		}

		if curveBounds.Llx < minX {
			minX = curveBounds.Llx
		}
		if curveBounds.Urx > maxX {
			maxX = curveBounds.Urx
		}
		if curveBounds.Lly < minY {
			minY = curveBounds.Lly
		}
		if curveBounds.Ury > maxY {
			maxY = curveBounds.Ury
		}
	}

	bbox.X = minX
	bbox.Y = minY
	bbox.Width = maxX - minX
	bbox.Height = maxY - minY
	return bbox
}
