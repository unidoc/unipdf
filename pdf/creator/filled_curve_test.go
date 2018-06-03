/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/unidoc/unidoc/pdf/contentstream/draw"
)

func CreateFillCurve(x0, y0, x1, y1, x2, y2, x3, y3 float64) draw.CubicBezierCurve {
	return draw.NewCubicBezierCurve(x0, y0, x1, y1, x2, y2, x3, y3)
}

func TestNewFilledCurve(t *testing.T) {
	filledCurve := NewFilledCurve()
	filledCurve.FillEnabled = true
	filledCurve.BorderEnabled = true
	filledCurve.BorderWidth = 2
	filledCurve.SetFillColor(ColorGreen)
	filledCurve.SetBorderColor(ColorBlue)

	// Up Left
	filledCurve.AppendCurve(CreateFillCurve(300, 300, 230, 350, 200, 280, 220, 220))
	// Down Left
	filledCurve.AppendCurve(CreateFillCurve(225, 240, 240, 180, 260, 160, 300, 180))
	// Down Right
	filledCurve.AppendCurve(CreateFillCurve(305, 170, 335, 165, 350, 185, 365, 220))
	// Up Right
	filledCurve.AppendCurve(CreateFillCurve(365, 240, 385, 315, 350, 325, 300, 300))
	// Leaf
	filledCurve.AppendCurve(CreateFillCurve(300, 300, 290, 350, 295, 370, 300, 390))

	creator := New()
	creator.NewPage()
	creator.Draw(filledCurve)

	err := creator.WriteToFile("/tmp/filledCurve.pdf")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
}
