/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReduceRankBinary2 tests the reduce rank binary 2 function.
func TestReduceRankBinary2(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		t.Run("NilSource", func(t *testing.T) {
			_, err := reduceRankBinary2(nil, 1, nil)
			require.Error(t, err)
		})
		t.Run("InvalidLevel", func(t *testing.T) {
			_, err := reduceRankBinary2(New(20, 20), 0, nil)
			require.Error(t, err)
		})
		t.Run("Height1", func(t *testing.T) {
			_, err := reduceRankBinary2(New(10, 1), 1, nil)
			require.Error(t, err)
		})
	})
	// Let's have a test 36x8 bitmap with the following data:
	//
	// 10000000 11000000 11000000 11000000 11100000
	// 00000000 00000000 10000000 11000000 10100000
	//
	// 01000000 10000000 10000000 11000000 00000000
	// 00000000 10000000 11000000 11000000 10000000
	//
	// 00000000 10000000 01000000 11000000 00000000
	// 10000000 01000000 11000000 11000000 00000000
	//
	// 00000000 01000000 11000000 11000000 00000000
	// 01000000 10000000 01000000 11000000 00000000
	//										   ^
	//						Here starts the padding

	// the data is the byte slice used as the source of the bitmap.
	data := []byte{
		0x80, 0xC0, 0xC0, 0xC0, 0xE0,
		0x00, 0x00, 0x80, 0xC0, 0xA0,
		0x40, 0x80, 0x80, 0xC0, 0x00,
		0x00, 0x80, 0xC0, 0xC0, 0x80,
		0x00, 0x80, 0x40, 0xC0, 0x00,
		0x80, 0x40, 0xC0, 0xC0, 0x00,
		0x00, 0x40, 0xC0, 0xC0, 0x00,
		0x40, 0x80, 0x40, 0xC0, 0x00,
	}

	t.Run("Level1", func(t *testing.T) {
		// The level1 reduce should reduce all 2x2 regions that has at least
		// one pixel 'ON' starting from the UL corner.
		// Thus the result data should be:
		//
		// 10001000 10001000 11000000
		// 10001000 10001000 10000000
		// 10001000 10001000 00000000
		// 10001000 10001000 00000000
		s, err := NewWithData(36, 8, data)
		require.NoError(t, err)

		d, err := reduceRankBinary2(s, 1, nil)
		require.NoError(t, err)

		assert.Equal(t, 18, d.Width)
		assert.Equal(t, 4, d.Height)

		assert.Equal(t, []byte{0x88, 0x88, 0xC0, 0x88, 0x88, 0x80, 0x88, 0x88, 0x00, 0x88, 0x88, 0x00}, d.Data)
	})

	t.Run("Level2", func(t *testing.T) {
		// The level2 reduce should recude all 2x2 regions that has at least
		// two pixels 'ON' starting form the UL corenr.
		// The result data should be:
		//
		// 00001000 10001000 11000000
		// 00001000 10001000 00000000
		// 00001000 10001000 00000000
		// 00001000 10001000 00000000
		s, err := NewWithData(36, 8, data)
		require.NoError(t, err)

		d, err := reduceRankBinary2(s, 2, nil)
		require.NoError(t, err)

		assert.Equal(t, 18, d.Width)
		assert.Equal(t, 4, d.Height)

		assert.Equal(t, []byte{
			0x08, 0x88, 0xC0,
			0x08, 0x88, 0x00,
			0x08, 0x88, 0x00,
			0x08, 0x88, 0x00,
		}, d.Data)
	})

	t.Run("Level3", func(t *testing.T) {
		// The level3 reduce should recude all 2x2 regions that has at least
		// three pixels 'ON' starting form the UL corenr.
		// The result data should be:
		//
		// 00000000 10001000 11000000
		// 00000000 10001000 00000000
		// 00000000 10001000 00000000
		// 00000000 10001000 00000000
		s, err := NewWithData(36, 8, data)
		require.NoError(t, err)

		d, err := reduceRankBinary2(s, 3, nil)
		require.NoError(t, err)

		assert.Equal(t, 18, d.Width)
		assert.Equal(t, 4, d.Height)

		assert.Equal(t, []byte{
			0x00, 0x88, 0x80,
			0x00, 0x88, 0x00,
			0x00, 0x88, 0x00,
			0x00, 0x88, 0x00,
		}, d.Data)
	})

	t.Run("Level4", func(t *testing.T) {
		// The level4 reduce should recude all 2x2 regions that has at least
		// four pixels 'ON' starting form the UL corenr.
		// The result data should be:
		//
		// 00000000 00001000 00000000
		// 00000000 00001000 00000000
		// 00000000 00001000 00000000
		// 00000000 00001000 00000000
		s, err := NewWithData(36, 8, data)
		require.NoError(t, err)

		d, err := reduceRankBinary2(s, 4, nil)
		require.NoError(t, err)

		assert.Equal(t, 18, d.Width)
		assert.Equal(t, 4, d.Height)

		assert.Equal(t, []byte{
			0x00, 0x08, 0x00,
			0x00, 0x08, 0x00,
			0x00, 0x08, 0x00,
			0x00, 0x08, 0x00,
		}, d.Data)
	})
}

func TestReduceRankBinaryCascade(t *testing.T) {
	t.Run("NilSource", func(t *testing.T) {
		_, err := reduceRankBinaryCascade(nil)
		require.Error(t, err)
	})
	t.Run("NoLevels", func(t *testing.T) {
		s := New(80, 80)
		_, err := reduceRankBinaryCascade(s)
		require.Error(t, err)
	})
	t.Run("Level0", func(t *testing.T) {
		s := New(10, 10)
		d, err := reduceRankBinaryCascade(s, 0)
		require.NoError(t, err)

		assert.False(t, s == d)
		assert.Equal(t, s, d)
	})
}
