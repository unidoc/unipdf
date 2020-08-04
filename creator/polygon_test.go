/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPolygon(t *testing.T) {
	points := []Point{
		{X: 1.0, Y: 2.0},
	}
	polygon := newPolygon(points)
	assert.Equal(t, points, polygon.points)
}

func TestPolygonSetBorderWidth(t *testing.T) {
	polygon := Polygon{}
	polygon.SetBorderWidth(20.0)
	assert.Equal(t, 20.0, polygon.borderWidth)
}

func TestPolygonSetBorderColor(t *testing.T) {
	polygon := Polygon{}
	polygon.SetBorderColor(ColorRGBFromHex("#ffffff"))
	assert.Equal(t, 1.0, polygon.borderColor.R())
	assert.Equal(t, 1.0, polygon.borderColor.G())
	assert.Equal(t, 1.0, polygon.borderColor.B())
}

func TestPolygonSetBorderOpacity(t *testing.T) {
	polygon := Polygon{}
	polygon.SetBorderOpacity(0.5)
	assert.True(t, polygon.borderOpacityEnabled)
	assert.Equal(t, 0.5, polygon.borderOpacity)
}

func TestPolygonSetFillColor(t *testing.T) {
	polygon := Polygon{}
	polygon.SetFillColor(ColorRGBFromHex("#ffffff"))
	assert.Equal(t, 1.0, polygon.fillColor.R())
	assert.Equal(t, 1.0, polygon.fillColor.G())
	assert.Equal(t, 1.0, polygon.fillColor.B())
}

func TestPolygonSetFillOpacity(t *testing.T) {
	polygon := Polygon{}
	polygon.SetFillOpacity(0.5)
	assert.True(t, polygon.fillOpacityEnabled)
	assert.Equal(t, 0.5, polygon.fillOpacity)
}

func TestGeneratePageBlocksFromPolygonPoints(t *testing.T) {
	polygon := newPolygon([]Point{{X: 1.0, Y: 2.0}})

	drawContext := DrawContext{}

	blocks, ctx, err := polygon.GeneratePageBlocks(drawContext)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 2 m\nh\nQ\n", blocks[0].contents.String())
	assert.Equal(t, drawContext, ctx)
}

func TestGeneratePageBlocksFromPolygonWithFillColor(t *testing.T) {
	polygon := newPolygon([]Point{{X: 1.0, Y: 2.0}})
	polygon.SetFillColor(ColorRGBFromHex("#ffffff"))

	blocks, _, err := polygon.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 1 1 rg\n1 2 m\nh\nf\nQ\n", blocks[0].contents.String())
}

func TestGeneratePageBlocksFromPolygonWithBorder(t *testing.T) {
	polygon := newPolygon([]Point{{X: 1.0, Y: 2.0}})
	polygon.SetBorderColor(ColorRGBFromHex("#ffffff"))
	polygon.SetBorderWidth(20.0)

	blocks, _, err := polygon.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 1 1 RG\n20 w\n1 2 m\nh\nS\nQ\n", blocks[0].contents.String())
}

func TestGeneratePageBlocksFromPolygonWithOpacity(t *testing.T) {
	polygon := newPolygon([]Point{{X: 1.0, Y: 2.0}})
	polygon.SetFillOpacity(0.1)
	polygon.SetBorderOpacity(0.2)

	blocks, _, err := polygon.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n/GS0 gs\n1 2 m\nh\nQ\n", blocks[0].contents.String())
}
