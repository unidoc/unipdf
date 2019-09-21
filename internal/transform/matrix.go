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

// TranslationMatrix returns a matrix that translates by `tx`, `ty`.
func TranslationMatrix(tx, ty float64) Matrix {
	return NewMatrix(1, 0, 0, 1, tx, ty)
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

/*
 * Partial 2D affine matrix decomposition.
 *  | a b | ➡ | cosθ -sinθ | × | sX  0 |
 *  | c d |   | sinθ  cosθ |   |  0 sY |
 *
 * Scale, Rotate and Translate convert from scales, angles and translations to affine transforms.
 * ScalingFactor<X,Y>, Angle and Translation convert from affine transforms to scales, angles and
 * translations.
 *
 * Transforms don't have unique angles in this scheme.
 *  e.g. | 1  0 | maps (1,0)➡(1,0)  0° and (0,1)➡(0,-1) 180°
 *       | 0 -1 |
 *
 * TODO(peterwilliams97): Define a unique decomposition of a 2D affine transform into rotation,
 *  shear, anisotropic scaling and translation.
 *
 * See https://math.stackexchange.com/questions/78137/decomposition-of-a-nonsquare-affine-matrix/
 *
 *  A = | a b | ➡  | cosθ -sinθ | × | 1 0 | × |sX  0 |
 *      | c d |    | sinθ  cosθ |   | q 1 |   | 0 sY |
 *
 *  sX = sqrt(a^2 + b^2)
 *  sY = det(A)/sX = (ad - bc)/sqrt(a^2 + b^2)
 *   q = (ac + bd)/det(A) = (ac + bd)/(ad - bc)
 *   θ = atan(-b, a)
 */

// NewMatrix returns an affine transform matrix that
//   scales by `xScale`, `yScale`,
//   rotated by `theta` degrees, and
//   translates by `tx`, `ty`.
func NewMatrixFromTransforms(xScale, yScale, theta, tx, ty float64) Matrix {
	return IdentityMatrix().Scale(xScale, yScale).Rotate(theta).Translate(tx, ty)
}

// String returns a string describing `m`.
func (m Matrix) String() string {
	a, b, c, d, tx, ty := m[0], m[1], m[3], m[4], m[6], m[7]
	return fmt.Sprintf("[%7.4f,%7.4f,%7.4f,%7.4f:%7.4f,%7.4f]", a, b, c, d, tx, ty)
}

// Scale returns `m` with an extra  scaling of `xScale`,`yScale` to `m`.
// NOTE: This scaling pre-multiplies `m` so it will be scaled and rotated by `m`.
func (m Matrix) Scale(xScale, yScale float64) Matrix {
	return m.Mult(NewMatrix(xScale, 0, 0, yScale, 0, 0))
}

// Rotate returns `m` with an extra rotation of `theta` degrees.
// NOTE: This rotation pre-multiplies `m` so it will be scaled and rotated by `m`.
func (m Matrix) Rotate(theta float64) Matrix {
	sin, cos := math.Sincos(theta / 180.0 * math.Pi)
	return m.Mult(NewMatrix(cos, -sin, sin, cos, 0, 0))
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

// Translate returns `m` with an extra translation of `tx`,`ty`.
func (m Matrix) Translate(tx, ty float64) Matrix {
	return NewMatrix(m[0], m[1], m[3], m[4], m[6]+tx, m[7]+ty)
}

// Translation returns the translation part of `m`.
func (m Matrix) Translation() (float64, float64) {
	return m[6], m[7]
}

// Transform returns coordinates `x`,`y` transformed by `m`.
func (m Matrix) Transform(x, y float64) (float64, float64) {
	xp := x*m[0] + y*m[1] + m[6]
	yp := x*m[3] + y*m[4] + m[7]
	return xp, yp
}

// ScalingFactorX returns the X scaling of the affine transform.
func (m Matrix) ScalingFactorX() float64 {
	return math.Hypot(m[0], m[1])
}

// ScalingFactorY returns the Y scaling of the affine transform.
func (m Matrix) ScalingFactorY() float64 {
	return math.Hypot(m[3], m[4])
}

// Angle returns the angle of the affine transform in `m` in degrees.
func (m Matrix) Angle() float64 {
	theta := math.Atan2(-m[1], m[0])
	if theta < 0.0 {
		theta += 2 * math.Pi
	}
	return theta / math.Pi * 180.0
}

// Inverse returns the inverse of `m` and a boolean to indicate whether the inverse exists.
func (m Matrix) Inverse() (Matrix, bool) {
	a, b := m[0], m[1]
	c, d := m[3], m[4]
	tx, ty := m[6], m[7]
	det := a*d - b*c
	if math.Abs(det) < minDeterminant {
		return Matrix{}, false
	}
	aI, bI := d/det, -b/det
	cI, dI := -c/det, a/det
	txI := -(aI*tx + cI*ty)
	tyI := -(bI*tx + dI*ty)
	return NewMatrix(aI, bI, cI, dI, txI, tyI), true
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
func (m Matrix) Unrealistic() bool {
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
