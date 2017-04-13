/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

import "fmt"

type Point struct {
	X float64
	Y float64
}

func NewPoint(x, y float64) Point {
	point := Point{}
	point.X = x
	point.Y = y
	return point
}

func (p Point) Add(dx, dy float64) Point {
	p.X += dx
	p.Y += dy
	return p
}

// Add vector to a point.
func (this Point) AddVector(v Vector) Point {
	this.X += v.Dx
	this.Y += v.Dy
	return this
}

func (p Point) String() string {
	return fmt.Sprintf("(%.1f,%.1f)", p.X, p.Y)
}
