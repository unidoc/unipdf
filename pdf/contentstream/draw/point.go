/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

import (
	"fmt"

	"github.com/unidoc/unidoc/pdf/internal/transform"
)

type Point struct {
	X float64
	Y float64
}

func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

func (p Point) Add(dx, dy float64) Point {
	p.X += dx
	p.Y += dy
	return p
}

// AddVector adds vector to a point.
func (p Point) AddVector(v Vector) Point {
	p.X += v.Dx
	p.Y += v.Dy
	return p
}

// Rotate returns a new Point at `p` rotated by `theta` degrees.
func (p Point) Rotate(theta float64) Point {
	r := transform.NewPoint(p.X, p.Y).Rotate(theta)
	return NewPoint(r.X, r.Y)
}

func (p Point) String() string {
	return fmt.Sprintf("(%.1f,%.1f)", p.X, p.Y)
}
