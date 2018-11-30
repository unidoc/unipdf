/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package transform

import (
	"fmt"
	"math"

	"github.com/unidoc/unidoc/common"
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

// String returns a string describing `m`.
func (m Matrix) String() string {
	a, b, c, d, tx, ty := m[0], m[1], m[3], m[4], m[6], m[7]
	return fmt.Sprintf("[%.4f,%.4f,%.4f,%.4f:%.4f,%.4f]", a, b, c, d, tx, ty)
}

// Set sets `m` to affine transform a,b,c,d,tx,ty.
func (m *Matrix) Set(a, b, c, d, tx, ty float64) {
	m[0], m[1] = a, b
	m[3], m[4] = c, d
	m[6], m[7] = tx, ty
	m.clampRange()
}

// Concat sets `m` to `m` × `b`.
// `b` needs to be created by newMatrix. i.e. It must be an affine transform.
//    m00 m01 0     b00 b01 0     m00*b00 + m01*b01        m00*b10 + m01*b11        0
//    m10 m11 0  ×  b10 b11 0  =  m10*b00 + m11*b01        m10*b10 + m11*b11        0
//    m20 m21 1     b20 b21 1     m20*b00 + m21*b10 + b20  m20*b01 + m21*b11 + b21  1
func (m *Matrix) Concat(b Matrix) {
	*m = Matrix{
		m[0]*b[0] + m[1]*b[3], m[0]*b[1] + m[1]*b[4], 0,
		m[3]*b[0] + m[4]*b[3], m[3]*b[1] + m[4]*b[4], 0,
		m[6]*b[0] + m[7]*b[3] + b[6], m[6]*b[1] + m[7]*b[4] + b[7], 1,
	}
	m.clampRange()
}

// Mult returns `m` × `b`.
func (m Matrix) Mult(b Matrix) Matrix {
	m.Concat(b)
	return m
}

// Translate appends a translation of `dx`,`dy` to `m`.
// m.Translate(dx, dy) is equivalent to m.Concat(NewMatrix(1, 0, 0, 1, dx, dy))
func (m *Matrix) Translate(dx, dy float64) {
	m[6] += dx
	m[7] += dy
	m.clampRange()
}

// Translation returns the translation part of `m`.
func (m *Matrix) Translation() (float64, float64) {
	return m[6], m[7]
}

// Translation returns the translation part of `m`.
func (m *Matrix) ScalingX() float64 {
	return math.Hypot(m[0], m[1])
}

// Transform returns coordinates `x`,`y` transformed by `m`.
func (m *Matrix) Transform(x, y float64) (float64, float64) {
	xp := x*m[0] + y*m[1] + m[6]
	yp := x*m[3] + y*m[4] + m[7]
	return xp, yp
}

// ScalingFactorX returns X scaling of  the affine transform.
func (m *Matrix) ScalingFactorX() float64 {
	return math.Sqrt(m[0]*m[0] + m[1]*m[1])
}

// ScalingFactorY returns X scaling of  the affine transform.
func (m *Matrix) ScalingFactorY() float64 {
	return math.Sqrt(m[3]*m[3] + m[4]*m[4])
}

// Angle returns the angle of the affine transform.
// For simplicity, we assume the transform is a multiple of 90 degrees.
func (m *Matrix) Angle() int {
	a, b, c, d := m[0], m[1], m[3], m[4]
	// We are returning θ for
	// a b    cos θ  -sin θ
	// c d =  sin θ   cos θ
	if a > 0 && d > 0 {
		//  1  0
		//  0  1
		return 0
	} else if b < 0 && c > 0 {
		//  0  1
		// -1  0
		return 90
	} else if a < 0 && d < 0 {
		// -1  0
		//  0 -1
		return 180
	} else if b > 0 && c < 0 {
		// 0 -1
		// 1  0
		return 270
	}
	common.Log.Debug("ERROR: Angle not a multiple of 90°. m=%s", m)
	return 0
}

// clampRange forces `m` to have reasonable values. It is a guard against crazy values in corrupt PDF files.
// Currently it clamps elements to [-maxAbsNumber, -maxAbsNumber] to avoid floating point exceptions.
func (m *Matrix) clampRange() {
	for i, x := range m {
		if x > maxAbsNumber {
			common.Log.Debug("CLAMP: %d -> %d", x, maxAbsNumber)
			m[i] = maxAbsNumber
		} else if x < -maxAbsNumber {
			common.Log.Debug("CLAMP: %d -> %d", x, -maxAbsNumber)
			m[i] = -maxAbsNumber
		}
	}
}

// maxAbsNumber defines the maximum absolute value of allowed practical matrix element values as needed
// to avoid floating point exceptions.
// TODO(gunnsth): Add reference or point to a specific example PDF that validates this.
const maxAbsNumber = 1e9
