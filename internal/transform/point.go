/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 *
 * Based on pdf/contentstream/draw/point.go
 */

package transform

import (
	"fmt"
	"math"
)

// Point defines a point (X,Y) in Cartesian coordinates.
type Point struct {
	X float64
	Y float64
}

// NewPoint returns a Point at `(x,y)`.
func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Set mutates `p` and sets to coordinates `(x, y)`.
func (p *Point) Set(x, y float64) {
	p.X, p.Y = x, y
}

// Transform mutates and transforms `p` by the affine transformation a, b, c, d, tx, ty.
func (p *Point) Transform(a, b, c, d, tx, ty float64) {
	m := NewMatrix(a, b, c, d, tx, ty)
	p.transformByMatrix(m)
}

// Displace returns a new Point at location `p` + `delta`.
func (p Point) Displace(delta Point) Point {
	return Point{p.X + delta.X, p.Y + delta.Y}
}

// Rotate returns a new Point at `p` rotated by `theta` degrees.
func (p Point) Rotate(theta float64) Point {
	r := math.Hypot(p.X, p.Y)
	t := math.Atan2(p.Y, p.X)
	sin, cos := math.Sincos(t + theta/180.0*math.Pi)
	return Point{r * cos, r * sin}
}

// Distance returns the distance between `a` and `b`.
func (a Point) Distance(b Point) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

// Interpolate does linear interpolation between point `a` and `b` for value `t`.
func (a Point) Interpolate(b Point, t float64) Point {
	return Point{
		X: (1-t)*a.X + t*b.X,
		Y: (1-t)*a.Y + t*b.Y,
	}
}

// String returns a string describing `p`.
func (p Point) String() string {
	return fmt.Sprintf("(%.2f,%.2f)", p.X, p.Y)
}

// transformByMatrix mutates and transforms `p` by the affine transformation `m`.
func (p *Point) transformByMatrix(m Matrix) {
	p.X, p.Y = m.Transform(p.X, p.Y)
}
