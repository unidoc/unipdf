/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var curves []CubicBezierCurve

func init() {
	curves = []CubicBezierCurve{
		{
			P0: Point{X: 0.0, Y: 0.0},
			P1: Point{X: 1.0, Y: 2.0},
			P2: Point{X: 2.0, Y: 2.0},
			P3: Point{X: 3.0, Y: 0.0},
		},
		{
			P0: Point{X: 3.0, Y: 0.0},
			P1: Point{X: 4.0, Y: 2.0},
			P2: Point{X: 5.0, Y: 2.0},
			P3: Point{X: 6.0, Y: 0.0},
		},
	}
}

func TestNewPolyCubicBezierCurve(t *testing.T) {
	curve := newPolyCubicBezierCurve(curves)
	assert.Equal(t, curves, curve.curves)
}

func TestPolyCubicBezierCurveSetBorderWidth(t *testing.T) {
	curve := PolyCubicBezierCurve{}
	curve.SetBorderWidth(20.0)
	assert.Equal(t, 20.0, curve.borderWidth)
}

func TestPolyCubicBezierCurveSetBorderColor(t *testing.T) {
	curve := PolyCubicBezierCurve{}
	curve.SetBorderColor(ColorRGBFromHex("#ffffff"))
	assert.Equal(t, 1.0, curve.borderColor.R())
	assert.Equal(t, 1.0, curve.borderColor.G())
	assert.Equal(t, 1.0, curve.borderColor.B())
}

func TestPolyCubicBezierCurveSetBorderOpacity(t *testing.T) {
	curve := PolyCubicBezierCurve{}
	curve.SetBorderOpacity(0.5)
	assert.True(t, curve.borderOpacityEnabled)
	assert.Equal(t, 0.5, curve.borderOpacity)
}

func TestPolyCubicBezierCurveSetFillColor(t *testing.T) {
	curve := PolyCubicBezierCurve{}
	curve.SetFillColor(ColorRGBFromHex("#ffffff"))
	assert.Equal(t, 1.0, curve.fillColor.R())
	assert.Equal(t, 1.0, curve.fillColor.G())
	assert.Equal(t, 1.0, curve.fillColor.B())
}

func TestPolyCubicBezierCurveSetFillOpacity(t *testing.T) {
	curve := PolyCubicBezierCurve{}
	curve.SetFillOpacity(0.5)
	assert.True(t, curve.fillOpacityEnabled)
	assert.Equal(t, 0.5, curve.fillOpacity)
}

func TestGeneratePageBlocksFromPolyCubicBezierCurvePoints(t *testing.T) {
	curve := newPolyCubicBezierCurve(curves)

	drawContext := DrawContext{}

	blocks, ctx, err := curve.GeneratePageBlocks(drawContext)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n0 0 0 RG\n1 w\n0 0 m\n1 2 2 2 3 0 c\n3 0 m\n4 2 5 2 6 0 c\nS\nQ\n", blocks[0].contents.String())
	assert.Equal(t, drawContext, ctx)
}

func TestGeneratePageBlocksFromPolyCubicBezierCurveWithFillColor(t *testing.T) {
	curve := newPolyCubicBezierCurve(curves)
	curve.SetFillColor(ColorRGBFromHex("#ffffff"))

	blocks, _, err := curve.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 1 1 rg\n0 0 0 RG\n1 w\n0 0 m\n1 2 2 2 3 0 c\n3 0 m\n4 2 5 2 6 0 c\nh\nB\nQ\n", blocks[0].contents.String())
}

func TestGeneratePageBlocksFromPolyCubicBezierCurveWithBorder(t *testing.T) {
	curve := newPolyCubicBezierCurve(curves)
	curve.SetBorderColor(ColorRGBFromHex("#ffffff"))
	curve.SetBorderWidth(20.0)

	blocks, _, err := curve.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 1 1 RG\n20 w\n0 0 m\n1 2 2 2 3 0 c\n3 0 m\n4 2 5 2 6 0 c\nS\nQ\n", blocks[0].contents.String())
}

func TestGeneratePageBlocksFromPolyCubicBezierCurveWithOpacity(t *testing.T) {
	curve := newPolyCubicBezierCurve(curves)
	curve.SetFillOpacity(0.1)
	curve.SetBorderOpacity(0.2)

	blocks, _, err := curve.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n0 0 0 RG\n1 w\n/GS0 gs\n0 0 m\n1 2 2 2 3 0 c\n3 0 m\n4 2 5 2 6 0 c\nS\nQ\n", blocks[0].contents.String())
}
