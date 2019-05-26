/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package transform

import (
	"math"
	"testing"

	"github.com/unidoc/unipdf/v3/common"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

// TestAngle tests the Matrix.Angle() function.
func TestAngle(t *testing.T) {
	extraTests := []angleCase{}
	for theta := 0.01; theta <= 360.0; theta *= 1.1 {
		extraTests = append(extraTests, makeAngleCase(2.0, theta))
	}

	const angleTol = 1.0e-10

	for _, test := range append(angleTests, extraTests...) {
		p := test.params
		m := NewMatrix(p.a, p.b, p.c, p.d, p.tx, p.ty)
		theta := m.Angle()
		if math.Abs(theta-test.theta) > angleTol {
			t.Fatalf("Bad angle: m=%s expected=%g° actual=%g°", m, test.theta, theta)
		}
	}
}

type params struct{ a, b, c, d, tx, ty float64 }
type angleCase struct {
	params         // Affine transform.
	theta  float64 // Rotation of affine transform in degrees.
}

var angleTests = []angleCase{
	{params: params{1, 0, 0, 1, 0, 0}, theta: 0},
	{params: params{0, -1, 1, 0, 0, 0}, theta: 90},
	{params: params{-1, 0, 0, -1, 0, 0}, theta: 180},
	{params: params{0, 1, -1, 0, 0, 0}, theta: 270},
	{params: params{1, -1, 1, 1, 0, 0}, theta: 45},
	{params: params{-1, -1, 1, -1, 0, 0}, theta: 135},
	{params: params{-1, 1, -1, -1, 0, 0}, theta: 225},
	{params: params{1, 1, -1, 1, 0, 0}, theta: 315},
}

// makeAngleCase makes an angleCase for a Matrix with scale `r` and angle `theta` degrees.
func makeAngleCase(r, theta float64) angleCase {
	radians := theta / 180.0 * math.Pi
	a := r * math.Cos(radians)
	b := -r * math.Sin(radians)
	c := -b
	d := a
	return angleCase{params{a, b, c, d, 0, 0}, theta}
}
