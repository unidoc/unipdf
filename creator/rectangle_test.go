/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRectangle(t *testing.T) {
	rectangle := newRectangle(1.0, 2.0, 3.0, 4.0)
	assert.Equal(t, 1.0, rectangle.x)
	assert.Equal(t, 2.0, rectangle.y)
	assert.Equal(t, 3.0, rectangle.width)
	assert.Equal(t, 4.0, rectangle.height)
	assert.Equal(t, 0.0, rectangle.borderColor.R())
	assert.Equal(t, 0.0, rectangle.borderColor.G())
	assert.Equal(t, 0.0, rectangle.borderColor.B())
	assert.Equal(t, 1.0, rectangle.borderWidth)
}

func TestRectangleGetCoords(t *testing.T) {
	rectangle := newRectangle(1.0, 2.0, 3.0, 4.0)
	x, y := rectangle.GetCoords()
	assert.Equal(t, 1.0, x)
	assert.Equal(t, 2.0, y)
}

func TestRectangleSetBorderWidth(t *testing.T) {
	rectangle := Rectangle{}
	rectangle.SetBorderWidth(20.0)
	assert.Equal(t, 20.0, rectangle.borderWidth)
}

func TestRectangleSetBorderColor(t *testing.T) {
	rectangle := Rectangle{}
	rectangle.SetBorderColor(ColorRGBFromHex("#ffffff"))
	assert.Equal(t, 1.0, rectangle.borderColor.R())
	assert.Equal(t, 1.0, rectangle.borderColor.G())
	assert.Equal(t, 1.0, rectangle.borderColor.B())
}

func TestRectangleSetBorderOpacity(t *testing.T) {
	rectangle := Rectangle{}
	rectangle.SetBorderOpacity(0.5)
	assert.True(t, rectangle.borderOpacityEnabled)
	assert.Equal(t, 0.5, rectangle.borderOpacity)
}

func TestRectangleSetFillColor(t *testing.T) {
	rectangle := Rectangle{}
	rectangle.SetFillColor(ColorRGBFromHex("#ffffff"))
	assert.Equal(t, 1.0, rectangle.fillColor.R())
	assert.Equal(t, 1.0, rectangle.fillColor.G())
	assert.Equal(t, 1.0, rectangle.fillColor.B())
}

func TestRectangleSetFillOpacity(t *testing.T) {
	rectangle := Rectangle{}
	rectangle.SetFillOpacity(0.5)
	assert.True(t, rectangle.fillOpacityEnabled)
	assert.Equal(t, 0.5, rectangle.fillOpacity)
}

func TestGeneratePageBlocksFromRectangle(t *testing.T) {
	rectangle := newRectangle(1.0, 2.0, 3.0, 4.0)

	drawContext := DrawContext{}

	blocks, ctx, err := rectangle.GeneratePageBlocks(drawContext)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n0 0 0 RG\n1 w\n1 -6 m\n1 -2 l\n4 -2 l\n4 -6 l\n1 -6 l\nh\nS\nQ\n", blocks[0].contents.String())
	assert.Equal(t, drawContext, ctx)
}

func TestGeneratePageBlocksFromRectangleWithFillColor(t *testing.T) {
	rectangle := newRectangle(1.0, 2.0, 3.0, 4.0)
	rectangle.SetFillColor(ColorRGBFromHex("#ffffff"))

	blocks, _, err := rectangle.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 1 1 rg\n0 0 0 RG\n1 w\n1 -6 m\n1 -2 l\n4 -2 l\n4 -6 l\n1 -6 l\nh\nB\nQ\n", blocks[0].contents.String())
}

func TestGeneratePageBlocksFromRectangleWithBorder(t *testing.T) {
	rectangle := newRectangle(1.0, 2.0, 3.0, 4.0)
	rectangle.SetBorderColor(ColorRGBFromHex("#ffffff"))
	rectangle.SetBorderWidth(20.0)

	blocks, _, err := rectangle.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n1 1 1 RG\n20 w\n1 -6 m\n1 -2 l\n4 -2 l\n4 -6 l\n1 -6 l\nh\nS\nQ\n", blocks[0].contents.String())
}

func TestGeneratePageBlocksFromRectangleWithOpacity(t *testing.T) {
	rectangle := newRectangle(1.0, 2.0, 3.0, 4.0)
	rectangle.SetFillOpacity(0.1)
	rectangle.SetBorderOpacity(0.2)

	blocks, _, err := rectangle.GeneratePageBlocks(DrawContext{})
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}

	assert.Equal(t, "q\n0 0 0 RG\n1 w\n/GS0 gs\n1 -6 m\n1 -2 l\n4 -2 l\n4 -6 l\n1 -6 l\nh\nS\nQ\n", blocks[0].contents.String())
}
