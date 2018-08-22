/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 *
 * Based on pdf/contentstream/draw/bezier_curve.go
 */

package extractor

import (
	"fmt"
	"math"

	"github.com/unidoc/unidoc/pdf/contentstream"
)

// CubicBezierCurve describes a cubic Bézier curve which is defined by:
// R(t) = P0*(1-t)^3 + P1*3*t*(1-t)^2 + P2*3*t^2*(1-t) + P3*t^3
// where P0 is the current point, P1, P2 control points and P3 the final point.
type CubicBezierCurve struct {
	P0 Point // Starting point.
	P1 Point // Control point 1.
	P2 Point // Control point 2.
	P3 Point // Final point.
}

// NewCubicBezierCurve returns a CubicBezierCurve with points (xi, yi) i=0..3
func NewCubicBezierCurve(x0, y0, x1, y1, x2, y2, x3, y3 float64) CubicBezierCurve {
	return CubicBezierCurve{
		P0: NewPoint(x0, y0),
		P1: NewPoint(x1, y1),
		P2: NewPoint(x2, y2),
		P3: NewPoint(x3, y3),
	}
}

// Transform transforms all control points in `curve` by the affine transformation a, b, c, d, tx, ty
func (curve *CubicBezierCurve) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	curve.transformByMatrix(m)
}

// transformByMatrix transforms all control points in `curve` by the affine transformation `m`
func (curve *CubicBezierCurve) transformByMatrix(m contentstream.Matrix) {
	curve.P0.transformByMatrix(m)
	curve.P1.transformByMatrix(m)
	curve.P2.transformByMatrix(m)
	curve.P3.transformByMatrix(m)
}

// String returns a string describing `curve`
func (curve *CubicBezierCurve) String() string {
	return fmt.Sprintf("P0:%s,P1:%s,P2:%s,P3:%s",
		curve.P0.String(), curve.P1.String(), curve.P2.String(), curve.P3.String())
}

// GetBoundingBox returns the bounding box of `curve`.
func (curve *CubicBezierCurve) GetBoundingBox() BoundingBox {
	minX, maxX := curve.P0.X, curve.P0.X
	minY, maxY := curve.P0.Y, curve.P0.Y

	// Sample 1000 points.
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
		} else if Rx > maxX {
			maxX = Rx
		}
		if Ry < minY {
			minY = Ry
		} else if Ry > maxY {
			maxY = Ry
		}
	}

	return BoundingBox{
		Ll: Point{minX, minY},
		Ur: Point{maxX, maxY},
	}
}

// CubicBezierPath represents a pdf path composed of cubic Bézier curves
type CubicBezierPath struct {
	Curves []CubicBezierCurve
}

// NewCubicBezierPath returns a CubicBezierPath with no curves
func NewCubicBezierPath() CubicBezierPath {
	return CubicBezierPath{}
}

// AppendCurve appends `curve` to `bpath`
func (bpath *CubicBezierPath) AppendCurve(curve CubicBezierCurve) {
	bpath.Curves = append(bpath.Curves, curve)
}

// Copy returns a copy of `bpath`
func (bpath *CubicBezierPath) Copy() CubicBezierPath {
	bpathcopy := CubicBezierPath{}
	bpathcopy.Curves = []CubicBezierCurve{}
	for _, c := range bpath.Curves {
		bpathcopy.Curves = append(bpathcopy.Curves, c)
	}
	return bpathcopy
}

// Transform transforms all curves in `bpath` by the affine transformation a, b, c, d, tx, ty
func (bpath *CubicBezierPath) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	bpath.transformByMatrix(m)
}

// transformByMatrix transforms all curves in `bpath` by the affine transformation `m`
func (bpath *CubicBezierPath) transformByMatrix(m contentstream.Matrix) {
	for _, curve := range bpath.Curves {
		curve.transformByMatrix(m)
	}
}

// String returns a string describing `bpath`
func (bpath *CubicBezierPath) String() string {
	return fmt.Sprintf("%+v", bpath.Curves)
}

// Length returns the number of curves in `bpath`.
func (bpath *CubicBezierPath) Length() int {
	return len(bpath.Curves)
}

// GetBoundingBox returns the bounding box of `bpath`
func (bpath *CubicBezierPath) GetBoundingBox() BoundingBox {
	if len(bpath.Curves) == 0 {
		return BoundingBox{}
	}

	bbox := bpath.Curves[0].GetBoundingBox()
	minX, minY := bbox.Ll.X, bbox.Ll.Y
	maxX, maxY := bbox.Ur.X, bbox.Ur.Y

	for _, curve := range bpath.Curves {
		bbox := curve.GetBoundingBox()
		if bbox.Ll.X < minX {
			minX = bbox.Ll.X
		}
		if bbox.Ur.X > maxX {
			maxX = bbox.Ur.X
		}
		if bbox.Ll.Y < minY {
			minY = bbox.Ll.Y
		}
		if bbox.Ur.Y > maxY {
			maxY = bbox.Ur.Y
		}
	}

	return BoundingBox{
		Ll: Point{minX, minY},
		Ur: Point{maxX, maxY},
	}
}
