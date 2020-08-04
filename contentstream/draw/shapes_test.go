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
		Points: []Point{
			{X: 0.0, Y: 1.0},
			{X: 2.0, Y: 1.0},
			{X: 2.0, Y: 3.0},
			{X: 0.0, Y: 3.0},
			{X: 0.0, Y: 1.0},
		},
	}
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
	assert.Equal(t, []byte{0x71, 0xa, 0x32, 0x35, 0x35, 0x20, 0x31, 0x32, 0x38, 0x20, 0x30, 0x20, 0x72, 0x67, 0xa, 0x68, 0xa, 0x66, 0xa, 0x51, 0xa}, bytes)
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
	assert.Equal(t, []byte{0x71, 0xa, 0x32, 0x35, 0x35, 0x20, 0x31, 0x32, 0x38, 0x20, 0x30, 0x20, 0x52, 0x47, 0xa, 0x31, 0x30, 0x20, 0x77, 0xa, 0x68, 0xa, 0x53, 0xa, 0x51, 0xa}, bytes)
}

func TestPolygonWithGsName(t *testing.T) {
	polygon := Polygon{}
	bytes, _, err := polygon.Draw("foo")
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.Equal(t, []byte{0x71, 0xa, 0x2f, 0x66, 0x6f, 0x6f, 0x20, 0x67, 0x73, 0xa, 0x68, 0xa, 0x51, 0xa}, bytes)
}
