/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 *
 * Based on pdf/contentstream/draw/point.go
 */

// !@#$ Functional, not updating
package extractor

import (
	"fmt"

	"github.com/unidoc/unidoc/pdf/contentstream"
)

// Point defines a point in Cartesian coordinates
type Point struct {
	X float64
	Y float64
}

// NewPoint returns a Point at 'x', 'y'.
func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Set sets `p` to `x`, `y`.
func (p *Point) Set(x, y float64) {
	p.X, p.Y = x, y
}

// Transform transforms `p` by the affine transformation a, b, c, d, tx, ty.
func (p *Point) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	p.transformByMatrix(m)
}

// Displace returns `p` displaced by `delta`.
func (p Point) Displace(delta Point) Point {
	return Point{p.X + delta.X, p.Y + delta.Y}
}

// transformByMatrix transforms `p` by the affine transformation `m`.
func (p *Point) transformByMatrix(m contentstream.Matrix) {
	p.X, p.Y = m.Transform(p.X, p.Y)
}

// String returns a string describing `p`.
func (p *Point) String() string {
	return fmt.Sprintf("(%.1f,%.1f)", p.X, p.Y)
}
