/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import "testing"

func TestNewCurve(t *testing.T) {
	creator := New()
	creator.NewPage()
	curve := NewCurve(20, 20, 15, 35, 40, 150)
	curve.SetWidth(3.0)
	curve.SetColor(ColorGreen)
	err := creator.Draw(curve)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	err = creator.WriteToFile("/tmp/curve.pdf")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
}

func CreateCurve(x1, y1, cx, cy, x2, y2 float64, color Color) *Curve {
	curve := NewCurve(x1, y1, cx, cy, x2, y2)
	curve.SetWidth(1)
	curve.SetColor(color)
	return curve
}

func CreateLine(x1, y1, x2, y2, width float64) *Line {
	line := NewLine(x1, y1, x2, y2)
	line.SetLineWidth(width)
	line.SetColor(ColorRed)
	return line
}

func TestNewCurveWithGlass(t *testing.T) {
	creator := New()
	creator.NewPage()

	// Width 200
	creator.Draw(CreateLine(30, 200, 270, 200, 1))

	// Curve up
	creator.Draw(CreateCurve(50, 200, 75, 145, 150, 150, ColorRed))
	creator.Draw(CreateCurve(150, 150, 205, 145, 250, 200, ColorGreen))

	// Curve down
	creator.Draw(CreateCurve(50, 200, 75, 245, 150, 250, ColorBlue))
	creator.Draw(CreateCurve(150, 250, 225, 245, 250, 200, ColorBlack))

	// Vertical line
	creator.Draw(CreateLine(50, 200, 51, 400, 1))
	creator.Draw(CreateLine(250, 200, 251, 400, 1))

	// Curve down
	creator.Draw(CreateCurve(51, 399, 75, 445, 150, 450, ColorRed))
	creator.Draw(CreateCurve(150, 450, 225, 445, 251, 399, ColorGreen))

	err := creator.WriteToFile("/tmp/curve_glass.pdf")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
}
