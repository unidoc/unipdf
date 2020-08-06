/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPolyline(t *testing.T) {
	points := []Point{{X: 1.0, Y: 2.0}}
	polyline := newPolyline(points)
	assert.Equal(t, points, polyline.points)
}

func TestPolylineSetLineWidth(t *testing.T) {
	polyline := Polyline{}
	polyline.SetLineWidth(20.0)
	assert.Equal(t, 20.0, polyline.lineWidth)
}

func TestPolylineSetLineColor(t *testing.T) {
	polyline := Polyline{}
	polyline.SetLineColor(ColorRGBFromHex("#ffffff"))
	assert.Equal(t, 1.0, polyline.lineColor.R())
	assert.Equal(t, 1.0, polyline.lineColor.G())
	assert.Equal(t, 1.0, polyline.lineColor.B())
}

func TestPolylineSetLineOpacity(t *testing.T) {
	polyline := Polyline{}
	polyline.SetLineOpacity(0.5)
	assert.True(t, polyline.lineOpacityEnabled)
	assert.Equal(t, 0.5, polyline.lineOpacity)
}

func TestGeneratePageBlocksFromPolylinePoints(t *testing.T) {
	polyline := newPolyline([]Point{{X: 1.0, Y: 2.0}})
	polyline.SetLineColor(ColorRGBFromHex("#ffffff"))
	polyline.SetLineWidth(20.0)

	drawContext := DrawContext{}

	blocks, ctx, err := polyline.GeneratePageBlocks(drawContext)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 2 m\n1 1 1 RG\n20 w\nS\nQ\n", blocks[0].contents.String())
	assert.Equal(t, drawContext, ctx)
}

func TestGeneratePageBlocksFromPolylineWithOpacity(t *testing.T) {
	polyline := newPolyline([]Point{{X: 1.0, Y: 2.0}})
	polyline.SetLineOpacity(0.1)

	blocks, _, err := polyline.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 2 m\n0 0 0 RG\n1 w\n/GS0 gs\nS\nQ\n", blocks[0].contents.String())
}
