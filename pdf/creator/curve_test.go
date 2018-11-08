/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import "testing"

func TestNewCurve(t *testing.T) {
	creator := New()
	creator.NewPage()
	curve := creator.NewCurve(20, 20, 15, 35, 40, 150)
	curve.SetWidth(3.0)
	curve.SetColor(ColorGreen)
	err := creator.Draw(curve)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	err = creator.WriteToFile(tempFile("curve.pdf"))
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
}

func CreateCurve(c *Creator, x1, y1, cx, cy, x2, y2 float64, color Color) *Curve {
	curve := c.NewCurve(x1, y1, cx, cy, x2, y2)
	curve.SetWidth(1)
	curve.SetColor(color)
	return curve
}

func CreateLine(c *Creator, x1, y1, x2, y2, width float64) *Line {
	line := c.NewLine(x1, y1, x2, y2)
	line.SetLineWidth(width)
	line.SetColor(ColorRed)
	return line
}

func TestNewCurveWithGlass(t *testing.T) {
	creator := New()
	creator.NewPage()

	// Width 200
	creator.Draw(CreateLine(creator, 30, 200, 270, 200, 1))

	// Curve up
	creator.Draw(CreateCurve(creator, 50, 200, 75, 145, 150, 150, ColorRed))
	creator.Draw(CreateCurve(creator, 150, 150, 205, 145, 250, 200, ColorGreen))

	// Curve down
	creator.Draw(CreateCurve(creator, 50, 200, 75, 245, 150, 250, ColorBlue))
	creator.Draw(CreateCurve(creator, 150, 250, 225, 245, 250, 200, ColorBlack))

	// Vertical line
	creator.Draw(CreateLine(creator, 50, 200, 51, 400, 1))
	creator.Draw(CreateLine(creator, 250, 200, 251, 400, 1))

	// Curve down
	creator.Draw(CreateCurve(creator, 51, 399, 75, 445, 150, 450, ColorRed))
	creator.Draw(CreateCurve(creator, 150, 450, 225, 445, 251, 399, ColorGreen))

	err := creator.WriteToFile(tempFile("curve_glass.pdf"))
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
}
