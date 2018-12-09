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

// Flip changes the sign of the vector: -vector.
func (v Vector) Flip() Vector {
	mag := v.Magnitude()
	theta := v.GetPolarAngle()

	v.Dx = mag * math.Cos(theta+math.Pi)
	v.Dy = mag * math.Sin(theta+math.Pi)
	return v
}

func (v Vector) FlipY() Vector {
	v.Dy = -v.Dy
	return v
}

func (v Vector) FlipX() Vector {
	v.Dx = -v.Dx
	return v
}

func (v Vector) Scale(factor float64) Vector {
	mag := v.Magnitude()
	theta := v.GetPolarAngle()

	v.Dx = factor * mag * math.Cos(theta)
	v.Dy = factor * mag * math.Sin(theta)
	return v
}

func (v Vector) Magnitude() float64 {
	return math.Sqrt(math.Pow(v.Dx, 2.0) + math.Pow(v.Dy, 2.0))
}

func (v Vector) GetPolarAngle() float64 {
	return math.Atan2(v.Dy, v.Dx)
}
