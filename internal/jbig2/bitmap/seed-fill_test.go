/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
)

func TestSeedfillBinary(t *testing.T) {
	// Having a test bitmap
	//
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	// 00000000 00111000 00000000
	// 00000000 00111000 00000000
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	// 00000000 00000000 00000000
	sData := []byte{
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x38, 0x00,
		0x00, 0x38, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
	}

	s, err := NewWithData(24, 12, sData)
	require.NoError(t, err)

	// The mask bitmap is a boundary for the pixels where they might
	// be filled in the source bitmap.
	// This means the whenever the mask bitmap has any 'ZERO' pixels it
	// can't be set in the result bitmap.
	// Our testing mask is divided into two parts joined together with a corner pixel
	// The upper one contains a 'hole' which should not be filled.
	// The lower one should be filled only if the connectivity is equal to 8.
	//
	// 00000000 00000000 00000000
	// 00000001 11111111 10000000
	// 00000001 11100011 10000000
	// 00000001 11111111 10000000
	// 00000001 11111111 10000000
	// 00000001 11111111 10000000
	// 00000001 11111111 10000000
	// 00000010 00000000 00000000
	// 00000001 11111111 10000000
	// 00000001 11111111 10000000
	// 00000001 11111111 10000000
	// 00000000 00000000 00000000
	mData := []byte{
		0x00, 0x00, 0x00,
		0x01, 0xFF, 0x80,
		0x01, 0xE3, 0x80,
		0x01, 0xFF, 0x80,
		0x01, 0xFF, 0x80,
		0x01, 0xFF, 0x80,
		0x01, 0xFF, 0x80,
		0x02, 0x00, 0x00,
		0x01, 0xFF, 0x80,
		0x01, 0xFF, 0x80,
		0x01, 0xFF, 0x80,
		0x00, 0x00, 0x00,
	}
	m, err := NewWithData(24, 12, mData)
	require.NoError(t, err)

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NilSource", func(t *testing.T) {
			_, err = seedfillBinary(nil, nil, m, 4)
			assert.Error(t, err)
		})

		t.Run("NilMask", func(t *testing.T) {
			_, err = seedfillBinary(nil, s, nil, 4)
			assert.Error(t, err)
		})

		t.Run("Connectivity", func(t *testing.T) {
			_, err = seedfillBinary(nil, s, m, 3)
			assert.Error(t, err)
		})
	})

	t.Run("Connectivity4", func(t *testing.T) {
		// The seedfillBinary with 4 - connectivity checks the connection
		// of the 'ONE' pixels in four directions only (above, bottom, left, right).
		// In our example that constraint doesn't allow the pixels to go to the lower part
		// of the mask as it is connected using corner direction.
		d, err := seedfillBinary(nil, s, m, 4)
		require.NoError(t, err)

		// The result would be a part of the 'mask' data
		// that is connected only in the four possible directions:
		//
		// 00000000 00000000 00000000
		// 00000001 11111111 10000000
		// 00000001 11100011 10000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000000 00000000 00000000
		// 00000000 00000000 00000000
		// 00000000 00000000 00000000
		// 00000000 00000000 00000000
		// 00000000 00000000 00000000
		expected := []byte{
			0x00, 0x00, 0x00,
			0x01, 0xFF, 0x80,
			0x01, 0xE3, 0x80,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x00, 0x00, 0x00,
			0x00, 0x00, 0x00,
			0x00, 0x00, 0x00,
			0x00, 0x00, 0x00,
			0x00, 0x00, 0x00,
		}
		assert.Equal(t, expected, d.Data, d.String())
	})

	t.Run("Connectivity8", func(t *testing.T) {
		// The 8 - connectivity seedfillBinary checks the connectivity
		// of the 'ONE' pixels in eight direction (all possible).
		// That affects the pixels on the corners too. Thus filling the
		// image with a limiting mask, should fill whole mask.
		d, err := seedfillBinary(nil, s, m, 8)
		require.NoError(t, err)

		// The result data should look like full mask:
		//
		// 00000000 00000000 00000000
		// 00000001 11111111 10000000
		// 00000001 11100011 10000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000010 00000000 00000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000001 11111111 10000000
		// 00000000 00000000 00000000
		expected := []byte{
			0x00, 0x00, 0x00,
			0x01, 0xFF, 0x80,
			0x01, 0xE3, 0x80,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x02, 0x00, 0x00,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x01, 0xFF, 0x80,
			0x00, 0x00, 0x00,
		}
		assert.Equal(t, expected, d.Data, d.String())
	})
}

// TestSeedFillStack tests the seedFillStackBB functions.
func TestSeedFillStackBB(t *testing.T) {
	// Having a test bitmap
	//
	//                111111 11112222
	//     01234567 89012345 67890123
	//
	// 0:  00000000 00000000 00000000
	// 1:  00100000 00000000 00000000
	// 2:  01110011 11111111 11100000
	// 3:  00100111 11111111 11110000
	// 4:  00000110 00000000 00110000
	// 5:  00000110 00011100 00001000
	// 6:  00000111 00011110 00000000
	// 7:  00000111 11111100 00000000
	// 8:  00000001 11110000 00000000
	// 9:  11100000 00001111 00000000
	// 10: 01100000 00000111 10000000
	// 11: 00100000 00000010 00000000
	//
	sData := []byte{
		0x00, 0x00, 0x00,
		0x20, 0x00, 0x00,
		0x73, 0xFF, 0xE0,
		0x27, 0xFF, 0xF0,
		0x06, 0x00, 0x30,
		0x06, 0x1C, 0x08,
		0x07, 0x1E, 0x00,
		0x07, 0xFC, 0x00,
		0x01, 0xF0, 0x00,
		0xE0, 0x0F, 0x00,
		0x60, 0x07, 0x80,
		0x20, 0x02, 0x00,
	}

	s, err := NewWithData(24, 12, sData)
	require.NoError(t, err)

	t.Run("General", func(t *testing.T) {
		t.Run("Invalid", func(t *testing.T) {
			t.Run("NoSource", func(t *testing.T) {
				stack := basic.Stack{}
				_, err := seedFillStackBB(nil, &stack, 11, 7, 8)
				assert.Error(t, err)
			})
			t.Run("NoStack", func(t *testing.T) {
				s, err := copyBitmap(nil, s)
				require.NoError(t, err)

				_, err = seedFillStackBB(s, nil, 11, 7, 8)
				assert.Error(t, err)
			})
			t.Run("NoAuxStack", func(t *testing.T) {
				s, err := copyBitmap(nil, s)
				require.NoError(t, err)

				stack := &basic.Stack{}
				_, err = seedFillStackBB(s, stack, 11, 7, 4)
				assert.Error(t, err)

				_, err = seedFillStackBB(s, stack, 11, 7, 8)
				assert.Error(t, err)
			})
			t.Run("Connectivity", func(t *testing.T) {
				s, err := copyBitmap(nil, s)
				require.NoError(t, err)

				stack := &basic.Stack{Aux: &basic.Stack{}}
				_, err = seedFillStackBB(s, stack, 11, 7, 30)
				assert.Error(t, err)
			})
		})
		t.Run("Connectivity4", func(t *testing.T) {
			s, err := copyBitmap(nil, s)
			require.NoError(t, err)

			stack := &basic.Stack{Aux: &basic.Stack{}}
			box, err := seedFillStackBB(s, stack, 11, 7, 4)
			require.NoError(t, err)

			require.NotNil(t, box)

			assert.Equal(t, box.Min.X, 5)
			assert.Equal(t, box.Min.Y, 2)
			assert.Equal(t, box.Max.X, 20)
			assert.Equal(t, box.Max.Y, 9)
		})
		t.Run("Connectivity8", func(t *testing.T) {
			s, err := copyBitmap(nil, s)
			require.NoError(t, err)

			stack := &basic.Stack{Aux: &basic.Stack{}}
			box, err := seedFillStackBB(s, stack, 11, 7, 8)
			require.NoError(t, err)

			require.NotNil(t, box)

			assert.Equal(t, box.Min.X, 5)
			assert.Equal(t, box.Min.Y, 2)
			assert.Equal(t, box.Max.X, 21)
			assert.Equal(t, box.Max.Y, 12)
		})
	})

	t.Run("Connectivity4", func(t *testing.T) {
		t.Run("Invalid", func(t *testing.T) {
			t.Run("NoSource", func(t *testing.T) {
				stack := basic.Stack{}
				_, err := seedFillStack4BB(nil, &stack, 11, 7)
				assert.Error(t, err)
			})
			t.Run("NoStack", func(t *testing.T) {
				s, err := copyBitmap(nil, s)
				require.NoError(t, err)

				_, err = seedFillStack4BB(s, nil, 11, 7)
				assert.Error(t, err)
			})
			t.Run("NoAuxStack", func(t *testing.T) {
				s, err := copyBitmap(nil, s)
				require.NoError(t, err)

				stack := &basic.Stack{}
				_, err = seedFillStack4BB(s, stack, 11, 7)
				assert.Error(t, err)
			})
		})
		t.Run("OutOfRange", func(t *testing.T) {
			s, err := copyBitmap(nil, s)
			require.NoError(t, err)

			stack := &basic.Stack{Aux: &basic.Stack{}}
			box, err := seedFillStack4BB(s, stack, 50, 30)
			require.NoError(t, err)
			assert.Nil(t, box)
		})
		s, err := copyBitmap(nil, s)
		require.NoError(t, err)

		stack := &basic.Stack{Aux: &basic.Stack{}}
		box, err := seedFillStack4BB(s, stack, 11, 7)
		require.NoError(t, err)
		require.NotNil(t, box)

		assert.False(t, box.Empty())
		assert.Equal(t, box.Min.X, 5)
		assert.Equal(t, box.Min.Y, 2)
		assert.Equal(t, box.Max.X, 20)
		assert.Equal(t, box.Max.Y, 9)
	})

	t.Run("Connectivity8", func(t *testing.T) {
		t.Run("Invalid", func(t *testing.T) {
			t.Run("NoSource", func(t *testing.T) {
				stack := basic.Stack{}
				_, err := seedFillStack8BB(nil, &stack, 11, 7)
				assert.Error(t, err)
			})
			t.Run("NoStack", func(t *testing.T) {
				s, err := copyBitmap(nil, s)
				require.NoError(t, err)

				_, err = seedFillStack8BB(s, nil, 11, 7)
				assert.Error(t, err)
			})
			t.Run("NoAuxStack", func(t *testing.T) {
				s, err := copyBitmap(nil, s)
				require.NoError(t, err)

				stack := &basic.Stack{}
				_, err = seedFillStack8BB(s, stack, 11, 7)
				assert.Error(t, err)
			})
		})
		t.Run("OutOfRange", func(t *testing.T) {
			s, err := copyBitmap(nil, s)
			require.NoError(t, err)

			stack := &basic.Stack{Aux: &basic.Stack{}}
			box, err := seedFillStack8BB(s, stack, 50, 30)
			require.NoError(t, err)
			assert.Nil(t, box)
		})
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
		// create a copy of the bitmap 's'
		s, err := copyBitmap(nil, s)
		require.NoError(t, err)

		stack := &basic.Stack{Aux: &basic.Stack{}}

		box, err := seedFillStack8BB(s, stack, 11, 7)
		require.NoError(t, err)
		require.NotNil(t, box)

		assert.Equal(t, box.Min.X, 5)
		assert.Equal(t, box.Min.Y, 2)
		assert.Equal(t, box.Max.X, 21)
		assert.Equal(t, box.Max.Y, 12)
	})
}
