/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

import "math"

// Vector represents a two-dimensional vector.
type Vector struct {
	Dx float64
	Dy float64
}

// NewVector returns a new vector with the direction specified by dx and dy.
func NewVector(dx, dy float64) Vector {
	v := Vector{}
	v.Dx = dx
	v.Dy = dy
	return v
}

// NewVectorBetween returns a new vector with the direction specified by
// the subtraction of point a from point b (b-a).
func NewVectorBetween(a Point, b Point) Vector {
	v := Vector{}
	v.Dx = b.X - a.X
	v.Dy = b.Y - a.Y
	return v
}

// NewVectorPolar returns a new vector calculated from the specified
// magnitude and angle.
func NewVectorPolar(length float64, theta float64) Vector {
	v := Vector{}

	v.Dx = length * math.Cos(theta)
	v.Dy = length * math.Sin(theta)
	return v
}

// Add adds the specified vector to the current one and returns the result.
func (v Vector) Add(other Vector) Vector {
	v.Dx += other.Dx
	v.Dy += other.Dy
	return v
}

// Rotate rotates the vector by the specified angle.
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

// FlipY flips the sign of the Dy component of the vector.
func (v Vector) FlipY() Vector {
	v.Dy = -v.Dy
	return v
}

// FlipX flips the sign of the Dx component of the vector.
func (v Vector) FlipX() Vector {
	v.Dx = -v.Dx
	return v
}

// Scale scales the vector by the specified factor.
func (v Vector) Scale(factor float64) Vector {
	mag := v.Magnitude()
	theta := v.GetPolarAngle()

	v.Dx = factor * mag * math.Cos(theta)
	v.Dy = factor * mag * math.Sin(theta)
	return v
}

// Magnitude returns the magnitude of the vector.
func (v Vector) Magnitude() float64 {
	return math.Sqrt(math.Pow(v.Dx, 2.0) + math.Pow(v.Dy, 2.0))
}

// GetPolarAngle returns the angle the magnitude of the vector forms with the
// positive X-axis going counterclockwise.
func (v Vector) GetPolarAngle() float64 {
	return math.Atan2(v.Dy, v.Dx)
}
