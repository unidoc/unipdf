/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

import "math"

type Vector struct {
	Dx float64
	Dy float64
}

func NewVector(dx, dy float64) Vector {
	v := Vector{}
	v.Dx = dx
	v.Dy = dy
	return v
}

func NewVectorBetween(a Point, b Point) Vector {
	v := Vector{}
	v.Dx = b.X - a.X
	v.Dy = b.Y - a.Y
	return v
}

func NewVectorPolar(length float64, theta float64) Vector {
	v := Vector{}

	v.Dx = length * math.Cos(theta)
	v.Dy = length * math.Sin(theta)
	return v
}

func (v Vector) Add(other Vector) Vector {
	v.Dx += other.Dx
	v.Dy += other.Dy
	return v
}

func (v Vector) Rotate(phi float64) Vector {
	mag := v.Magnitude()
	angle := v.GetPolarAngle()

	return NewVectorPolar(mag, angle+phi)
}

// Change the sign of the vector: -vector.
func (this Vector) Flip() Vector {
	mag := this.Magnitude()
	theta := this.GetPolarAngle()

	this.Dx = mag * math.Cos(theta+math.Pi)
	this.Dy = mag * math.Sin(theta+math.Pi)
	return this
}

func (v Vector) FlipY() Vector {
	v.Dy = -v.Dy
	return v
}

func (v Vector) FlipX() Vector {
	v.Dx = -v.Dx
	return v
}

func (this Vector) Scale(factor float64) Vector {
	mag := this.Magnitude()
	theta := this.GetPolarAngle()

	this.Dx = factor * mag * math.Cos(theta)
	this.Dy = factor * mag * math.Sin(theta)
	return this
}

func (this Vector) Magnitude() float64 {
	return math.Sqrt(math.Pow(this.Dx, 2.0) + math.Pow(this.Dy, 2.0))
}

func (this Vector) GetPolarAngle() float64 {
	return math.Atan2(this.Dy, this.Dx)
}
