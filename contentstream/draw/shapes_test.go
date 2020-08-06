/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unidoc/unipdf/v3/model"
)

func TestPolygonBoundingBox(t *testing.T) {
	polygon := Polygon{
		Points: [][]Point{{
			{X: 0.0, Y: 1.0},
			{X: 2.0, Y: 1.0},
			{X: 2.0, Y: 3.0},
			{X: 0.0, Y: 3.0},
			{X: 0.0, Y: 1.0},
		},
		}}
	bytes, boundingBox, err := polygon.Draw("")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.NotNil(t, bytes)
	assert.Equal(t, boundingBox.Llx, 0.0)
	assert.Equal(t, boundingBox.Lly, 1.0)
	assert.Equal(t, boundingBox.Urx, 2.0)
	assert.Equal(t, boundingBox.Ury, 3.0)

	assert.Equal(t, "q\n0 1 m\n2 1 l\n2 3 l\n0 3 l\n0 1 l\nh\nQ\n", string(bytes))
}

func TestPolygonWithCutout(t *testing.T) {
	polygon := Polygon{
		Points: [][]Point{
			{
				{X: 1.0, Y: 1.0},
				{X: 4.0, Y: 1.0},
				{X: 4.0, Y: 4.0},
				{X: 1.0, Y: 4.0},
				{X: 1.0, Y: 1.0},
			},
			{
				{X: 2.0, Y: 2.0},
				{X: 3.0, Y: 2.0},
				{X: 3.0, Y: 3.0},
				{X: 2.0, Y: 3.0},
				{X: 2.0, Y: 2.0},
			},
		}}
	bytes, boundingBox, err := polygon.Draw("")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.NotNil(t, bytes)
	assert.Equal(t, boundingBox.Llx, 1.0)
	assert.Equal(t, boundingBox.Lly, 1.0)
	assert.Equal(t, boundingBox.Urx, 4.0)
	assert.Equal(t, boundingBox.Ury, 4.0)

	assert.Equal(t, "q\n1 1 m\n4 1 l\n4 4 l\n1 4 l\n1 1 l\nh\n2 2 m\n3 2 l\n3 3 l\n2 3 l\n2 2 l\nh\nQ\n", string(bytes))
}

func TestPolygonWithFill(t *testing.T) {
	polygon := Polygon{
		FillEnabled: true,
		FillColor:   model.NewPdfColorDeviceRGB(255, 128, 0),
	}
	bytes, _, err := polygon.Draw("")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.Equal(t, "q\n255 128 0 rg\nf\nQ\n", string(bytes))
}

func TestPolygonWithBorder(t *testing.T) {
	polygon := Polygon{
		BorderEnabled: true,
		BorderColor:   model.NewPdfColorDeviceRGB(255, 128, 0),
		BorderWidth:   10.0,
	}
	bytes, _, err := polygon.Draw("")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.Equal(t, "q\n255 128 0 RG\n10 w\nS\nQ\n", string(bytes))
}

func TestPolygonWithGsName(t *testing.T) {
	polygon := Polygon{}
	bytes, _, err := polygon.Draw("foo")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.Equal(t, "q\n/foo gs\nQ\n", string(bytes))
}

func TestPolylineBoundingBox(t *testing.T) {
	polyline := Polyline{
		Points: []Point{
			{X: 0.0, Y: 1.0},
			{X: 2.0, Y: 3.0},
			{X: 4.0, Y: 5.0},
		},
		LineColor: model.NewPdfColorDeviceRGB(255, 128, 0),
		LineWidth: 10.0,
	}
	bytes, boundingBox, err := polyline.Draw("")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.NotNil(t, bytes)
	assert.Equal(t, boundingBox.Llx, 0.0)
	assert.Equal(t, boundingBox.Lly, 1.0)
	assert.Equal(t, boundingBox.Urx, 4.0)
	assert.Equal(t, boundingBox.Ury, 5.0)

	assert.Equal(t, "q\n0 1 m\n2 3 l\n4 5 l\n255 128 0 RG\n10 w\nS\nQ\n", string(bytes))
}

func TestPolylineWithGsName(t *testing.T) {
	polyline := Polyline{LineColor: model.NewPdfColorDeviceRGB(0, 0, 0)}
	bytes, _, err := polyline.Draw("foo")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.Equal(t, "q\n0 0 0 RG\n0 w\n/foo gs\nS\nQ\n", string(bytes))
}
