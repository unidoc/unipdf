package imageutil

import (
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonochrome(t *testing.T) {
	g, err := NewImage(10, 10, 1, 1, nil, nil, nil)
	require.NoError(t, err)

	c := g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))

	g.Set(6, 6, color.White)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(math.MaxUint8))

	g.Set(6, 6, color.Black)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))
}

func TestGray2(t *testing.T) {
	g, err := NewImage(10, 10, 2, 1, nil, nil, nil)
	require.NoError(t, err)

	c := g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))

	g.Set(6, 6, color.White)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(math.MaxUint8))

	g.Set(6, 6, color.Black)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))
}

func TestGray4(t *testing.T) {
	g, err := NewImage(10, 10, 4, 1, nil, nil, nil)
	require.NoError(t, err)

	c := g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))

	g.Set(6, 6, color.White)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(math.MaxUint8))

	g.Set(6, 6, color.Black)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))
}

func TestGray8(t *testing.T) {
	g, err := NewImage(10, 10, 8, 1, nil, nil, nil)
	require.NoError(t, err)

	c := g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))

	g.Set(6, 6, color.White)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(math.MaxUint8))

	g.Set(6, 6, color.Black)
	c = g.At(6, 6).(color.Gray)
	assert.Equal(t, c.Y, uint8(0))
}

func TestGray16(t *testing.T) {
	g, err := NewImage(10, 10, 16, 1, nil, nil, nil)
	require.NoError(t, err)

	c := g.At(6, 6).(color.Gray16)
	assert.Equal(t, c.Y, uint16(0))

	g.Set(6, 6, color.White)
	c = g.At(6, 6).(color.Gray16)
	assert.Equal(t, c.Y, uint16(math.MaxUint16))

	g.Set(6, 6, color.Black)
	c = g.At(6, 6).(color.Gray16)
	assert.Equal(t, c.Y, uint16(0))
}
