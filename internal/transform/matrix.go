/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package transform

import (
	"fmt"
	"math"

	"github.com/unidoc/unipdf/v3/common"
)

// Matrix is a linear transform matrix in homogenous coordinates.
// PDF coordinate transforms are always affine so we only need 6 of these. See newMatrix.
type Matrix [9]float64

// IdentityMatrix returns the identity transform.
func IdentityMatrix() Matrix {
	return NewMatrix(1, 0, 0, 1, 0, 0)
}

// TranslationMatrix returns a matrix that translates by `tx`,`ty`.
func TranslationMatrix(tx, ty float64) Matrix {
	return NewMatrix(1, 0, 0, 1, tx, ty)
}

// ScaleMatrix returns a matrix that scales by `x`,`y`.
func ScaleMatrix(x, y float64) Matrix {
	return NewMatrix(x, 0, 0, y, 0, 0)
}

// RotationMatrix returns a matrix that rotates by angle `angle`, specified in radians.
func RotationMatrix(angle float64) Matrix {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return NewMatrix(c, s, -s, c, 0, 0)
}

// ShearMatrix returns a matrix that shears `x`,`y`.
func ShearMatrix(x, y float64) Matrix {
	return NewMatrix(1, y, x, 1, 0, 0)
}

// NewMatrix returns an affine transform matrix laid out in homogenous coordinates as
//      a  b  0
//      c  d  0
//      tx ty 1
func NewMatrix(a, b, c, d, tx, ty float64) Matrix {
	m := Matrix{
		a, b, 0,
		c, d, 0,
		tx, ty, 1,
	}
	m.clampRange()
	return m
}

// String returns a string describing `m`.
func (m Matrix) String() string {
	a, b, c, d, tx, ty := m[0], m[1], m[3], m[4], m[6], m[7]
	return fmt.Sprintf("[%7.4f,%7.4f,%7.4f,%7.4f:%7.4f,%7.4f]", a, b, c, d, tx, ty)
}

// Set sets `m` to affine transform a,b,c,d,tx,ty.
func (m *Matrix) Set(a, b, c, d, tx, ty float64) {
	m[0], m[1] = a, b
	m[3], m[4] = c, d
	m[6], m[7] = tx, ty
	m.clampRange()
}

// Concat sets `m` to `b` × `m`.
// `b` needs to be created by newMatrix. i.e. It must be an affine transform.
//    b00 b01 0     m00 m01 0     b00*m00 + b01*m01        b00*m10 + b01*m11        0
//    b10 b11 0  ×  m10 m11 0  ➔  b10*m00 + b11*m01        b10*m10 + b11*m11        0
//    b20 b21 1     m20 m21 1     b20*m00 + b21*m10 + m20  b20*m01 + b21*m11 + m21  1
func (m *Matrix) Concat(b Matrix) {
	*m = Matrix{
		b[0]*m[0] + b[1]*m[3], b[0]*m[1] + b[1]*m[4], 0,
		b[3]*m[0] + b[4]*m[3], b[3]*m[1] + b[4]*m[4], 0,
		b[6]*m[0] + b[7]*m[3] + m[6], b[6]*m[1] + b[7]*m[4] + m[7], 1,
	}
	m.clampRange()
}

// Mult returns `b` × `m`.
func (m Matrix) Mult(b Matrix) Matrix {
	m.Concat(b)
	return m
}

// Translate appends a translation of `x`,`y` to `m`.
// m.Translate(dx, dy) is equivalent to m.Concat(NewMatrix(1, 0, 0, 1, dx, dy))
func (m *Matrix) Translate(x, y float64) {
	m.Concat(TranslationMatrix(x, y))
}

// Translation returns the translation part of `m`.
func (m *Matrix) Translation() (float64, float64) {
	return m[6], m[7]
}

// Scale scales the current matrix by `x`,`y`.
func (m *Matrix) Scale(x, y float64) {
	m.Concat(ScaleMatrix(x, y))
}

// Rotate rotates the current matrix by angle `angle`, specified in radians.
func (m *Matrix) Rotate(angle float64) {
	m.Concat(RotationMatrix(angle))
}

// Shear shears the current matrix by `x',`y`.
func (m *Matrix) Shear(x, y float64) {
	m.Concat(ShearMatrix(x, y))
}

// Clone returns a copy of the current matrix.
func (m *Matrix) Clone() Matrix {
	return NewMatrix(m[0], m[1], m[3], m[4], m[6], m[7])
}

// Transform returns coordinates `x`,`y` transformed by `m`.
func (m *Matrix) Transform(x, y float64) (float64, float64) {
	xp := x*m[0] + y*m[1] + m[6]
	yp := x*m[3] + y*m[4] + m[7]
	return xp, yp
}

// ScalingFactorX returns the X scaling of the affine transform.
func (m *Matrix) ScalingFactorX() float64 {
	return math.Hypot(m[0], m[1])
}

// ScalingFactorY returns the Y scaling of the affine transform.
func (m *Matrix) ScalingFactorY() float64 {
	return math.Hypot(m[3], m[4])
}

// Angle returns the angle of the affine transform in `m` in degrees.
func (m *Matrix) Angle() float64 {
	theta := math.Atan2(-m[1], m[0])
	if theta < 0.0 {
		theta += 2 * math.Pi
	}
	return theta / math.Pi * 180.0

}

// clampRange forces `m` to have reasonable values. It is a guard against crazy values in corrupt PDF files.
// Currently it clamps elements to [-maxAbsNumber, -maxAbsNumber] to avoid floating point exceptions.
func (m *Matrix) clampRange() {
	for i, x := range m {
		if x > maxAbsNumber {
			common.Log.Debug("CLAMP: %g -> %g", x, maxAbsNumber)
			m[i] = maxAbsNumber
		} else if x < -maxAbsNumber {
			common.Log.Debug("CLAMP: %g -> %g", x, -maxAbsNumber)
			m[i] = -maxAbsNumber
		}
	}
}

// Unrealistic returns true if `m` is too small to have been created intentionally.
// If it returns true then `m` probably contains junk values, due to some processing error in the
// PDF generator or our code.
func (m *Matrix) Unrealistic() bool {
	xx, xy, yx, yy := math.Abs(m[0]), math.Abs(m[1]), math.Abs(m[3]), math.Abs(m[4])
	goodXxYy := xx > minSafeScale && yy > minSafeScale
	goodXyYx := xy > minSafeScale && yx > minSafeScale
	return !(goodXxYy || goodXyYx)
}

// minSafeScale is the minimum matrix scale that is expected to occur in a valid PDF file.
const minSafeScale = 1e-6

// maxAbsNumber defines the maximum absolute value of allowed practical matrix element values as needed
// to avoid floating point exceptions.
// TODO(gunnsth): Add reference or point to a specific example PDF that validates this.
const maxAbsNumber = 1e9

// minDeterminant is the smallest matrix determinant we are prepared to deal with.
// Smaller determinants may lead to rounding errors.
const minDeterminant = 1.0e-6
