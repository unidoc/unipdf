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

// TestCentroid tests the centroid function.
func TestCentroid(t *testing.T) {
	// The centroid is based on the pixel number of the bitmap.
	// Having the bitmap 10×10 with 2 pix border.
	bm, err := New(8, 8).AddBorder(2, 1)
	require.NoError(t, err)

	pt, err := Centroid(bm, nil, nil)
	require.NoError(t, err)

	// the following for-loops computes in less efficient way the 'xsum', 'ysum' and 'pixsum'
	// Where:
	// 	point. X = float32(xsum)/float32(pixsum)
	//	point. Y = float32(ysum)/float32(pixsum)
	var xsum, ysum, pixsum int
	for i := 0; i < bm.Height; i++ {
		for j := 0; j < bm.Width; j++ {
			if bm.GetPixel(j, i) {
				xsum += j
				ysum += i
				pixsum++
			}
		}
	}
	ptCompare := Point{float32(xsum) / float32(pixsum), float32(ysum) / float32(pixsum)}

	// thus the 'pt' and 'ptCompare' should have equal 'X' and 'Y' values.
	assert.Equal(t, ptCompare, pt)

	bms := []*Bitmap{bm.Copy(), bm.Copy(), bm.Copy()}
	pts, err := Centroids(bms)
	require.NoError(t, err)

	for _, pt := range *pts {
		assert.Equal(t, ptCompare, pt)
	}
}

// TestMorphSequence tests the morph sequence functions.
func TestMorphSequence(t *testing.T) {
	t.Run("Verify", func(t *testing.T) {
		t.Run("SegmentMask", func(t *testing.T) {
			// leptonica message = "r11"
			var i, netRed, border int
			process := MorphProcess{Operation: MopRankBinaryReduction, Arguments: []int{1, 1}}
			err := process.verify(i, &netRed, &border)
			require.NoError(t, err)
		})

		t.Run("SegmentSeed", func(t *testing.T) {
			// leptonica message = "r1143 + o4.4 + x4"
			processes := []MorphProcess{
				{MopRankBinaryReduction, []int{1, 1, 4, 3}},
				{MopOpening, []int{4, 4}},
				{MopReplicativeBinaryExpansion, []int{4}},
			}
			err := verifyMorphProcesses(processes...)
			require.NoError(t, err)
		})

		t.Run("SegmentAddBorder", func(t *testing.T) {
			t.Run("Valid", func(t *testing.T) {
				processes := []MorphProcess{
					{MopAddBorder, []int{1}},
				}

				err := verifyMorphProcesses(processes...)
				assert.NoError(t, err)
			})

			t.Run("NotFirst", func(t *testing.T) {
				processes := []MorphProcess{
					{MopOpening, []int{1, 2}},
					{MopAddBorder, []int{2}},
				}
				err := verifyMorphProcesses(processes...)
				require.Error(t, err)
			})
		})
	})

	t.Run("Chain", func(t *testing.T) {
		processes := []MorphProcess{
			{MopAddBorder, []int{1}},
			{MopDilation, []int{1, 2}},
			{MopErosion, []int{1, 2}},
			{MopOpening, []int{2, 1}},
			{MopClosing, []int{1, 2}},
			{MopRankBinaryReduction, []int{2}},
			{MopReplicativeBinaryExpansion, []int{2}},
		}

		s, err := NewWithData(36, 8, []byte{
			0xFF, 0xFF, 0x0F, 0x0F, 0xF0,
			0xF0, 0x0F, 0x00, 0x0A, 0xC0,
			0xE0, 0xCC, 0xCF, 0xFC, 0xF0,
			0xEA, 0xFA, 0xFF, 0xFB, 0x80,
			0x88, 0X87, 0xF7, 0x78, 0x80,
			0xF8, 0x77, 0x73, 0x13, 0xC0,
			0xFF, 0xFF, 0xFF, 0xFF, 0xF0,
			0xF0, 0xC0, 0xE0, 0xB0, 0xF0,
		})
		require.NoError(t, err)

		d, err := MorphSequence(s, processes...)
		require.NoError(t, err)

		assert.NotNil(t, d)
	})
}

// TestDilate tests the dilate morph function.
func TestDilate(t *testing.T) {
	// Let's have a test 36×8 bitmap with the following data:
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

	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	sel := selCreate(8, 36, "")

	// Smear the selection

	// First shifted 1 right:
	//
	// 010000000110000001100000011000000111
	// 000000000000000001000000011000000101
	// 001000000100000001000000011000000000
	// 000000000100000001100000011000000100
	// 000000000100000000100000011000000000
	// 010000000010000001100000011000000000
	// 000000000010000001100000011000000000
	// 001000000100000000100000011000000000
	sel.Data[0][1] = SelHit

	// Second shifter 2 right
	//
	// 011000000111000001110000011100000111
	// 000000000000000001100000011100000111
	// 001100000110000001100000011100000000
	// 000000000110000001110000011100000110
	// 000000000110000000110000011100000000
	// 011000000011000001110000011100000000
	// 000000000011000001110000011100000000
	// 001100000110000000110000011100000000
	sel.Data[0][2] = SelHit

	// Last shifted 1
	// 011000000111000001110000011100000111
	// 100000001100000011100000111100001111
	// 001100000110000011100000111100001010
	// 010000001110000011110000111100000110
	// 000000001110000011110000111100001000
	// 011000001011000001110000111100000000
	// 100000000111000011110000111100000000
	// 001100000110000011110000111100000000

	sel.Data[1][0] = SelHit

	expected := []byte{
		0x60, 0x70, 0x70, 0x70, 0x70, 0x80, 0xc0, 0xe0,
		0xf0, 0xf0, 0x30, 0x60, 0xe0, 0xf0, 0xa0, 0x40,
		0xe0, 0xf0, 0xf0, 0x60, 0x00, 0xe0, 0xf0, 0xf0,
		0x80, 0x60, 0xb0, 0x70, 0xf0, 0x00, 0x80, 0x70,
		0xf0, 0xf0, 0x00, 0x30, 0x60, 0xf0, 0xf0, 0x00,
	}

	t.Run("NilDest", func(t *testing.T) {
		d, err := Dilate(nil, s, sel)
		require.NoError(t, err)

		assert.Equal(t, expected, d.Data)
	})

	t.Run("WithDest", func(t *testing.T) {
		d := New(36, 8)

		res, err := Dilate(d, s, sel)
		require.NoError(t, err)

		assert.True(t, d == res)
		assert.Equal(t, expected, d.Data)
	})

	t.Run("NilSrc", func(t *testing.T) {
		_, err := Dilate(New(36, 8), nil, sel)
		assert.Error(t, err)
	})
}

// TestDilateBrick tests the dilate brick function.
func TestDilateBrick(t *testing.T) {
	// Let's have a test 36×8 bitmap with the following data:
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

	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	t.Run("NilDest", func(t *testing.T) {
		d, err := DilateBrick(nil, s, 3, 3)
		require.NoError(t, err)

		// Expected 'smear' result
		// 110000011110000111100001111000011111
		// 111000011110000111100001111000011111
		// 111000011100000111100001111000011111
		// 111000011100000111100001111000011100
		// 110000011110000111100001111000011100
		// 110000011110000111100001111000000000
		// 111000011110000111100001111000000000
		// 111000011110000111100001111000000000
		expected := []byte{
			0xc1, 0xe1, 0xe1, 0xe1, 0xf0,
			0xe1, 0xe1, 0xe1, 0xe1, 0xf0,
			0xe1, 0xc1, 0xe1, 0xe1, 0xf0,
			0xe1, 0xc1, 0xe1, 0xe1, 0xc0,
			0xc1, 0xe1, 0xe1, 0xe1, 0xc0,
			0xc1, 0xe1, 0xe1, 0xe0, 0x00,
			0xe1, 0xe1, 0xe1, 0xe0, 0x00,
			0xe1, 0xe1, 0xe1, 0xe0, 0x00,
		}
		assert.Equal(t, expected, d.Data)
	})
	t.Run("Size0", func(t *testing.T) {
		_, err := DilateBrick(New(36, 8), s, 0, 0)
		assert.Error(t, err)
	})

	t.Run("Size1", func(t *testing.T) {
		bm, err := DilateBrick(New(36, 8), s, 1, 1)
		require.NoError(t, err)

		assert.Equal(t, bm, s)
		assert.False(t, bm == s)
	})

	t.Run("VSize1", func(t *testing.T) {
		d := New(36, 8)

		bm, err := DilateBrick(d, s, 2, 1)
		require.NoError(t, err)
		// Expected result dilated with height 2
		//
		// 100000011100000111000001110000011110
		// 000000000000000110000001110000011110
		// 110000011000000110000001110000000000
		// 000000011000000111000001110000011000
		// 000000011000000011000001110000000000
		// 100000001100000111000001110000000000
		// 000000001100000111000001110000000000
		// 110000011000000011000001110000000000
		expected := []byte{
			0x81, 0xc1, 0xc1, 0xc1, 0xe0,
			0x00, 0x01, 0x81, 0xc1, 0xe0,
			0xc1, 0x81, 0x81, 0xc0, 0x00,
			0x01, 0x81, 0xc1, 0xc1, 0x80,
			0x01, 0x80, 0xc1, 0xc0, 0x00,
			0x80, 0xc1, 0xc1, 0xc0, 0x00,
			0x00, 0xc1, 0xc1, 0xc0, 0x00,
			0xc1, 0x80, 0xc1, 0xc0, 0x00,
		}
		assert.Equal(t, expected, bm.Data)
		assert.True(t, bm == d)
	})
	// common. SetLogger(common NewConsoleLoggerr(commo LogLevelDebugug))
	t.Run("HSize1", func(t *testing.T) {
		d := New(36, 8)

		bm, err := DilateBrick(d, s, 1, 2)
		require.NoError(t, err)

		// Expected:
		//
		// 100000001100000011000000110000001110
		// 010000001000000010000000110000001010
		// 010000001000000011000000110000001000
		// 000000001000000011000000110000001000
		// 100000001100000011000000110000000000
		// 100000000100000011000000110000000000
		// 010000001100000011000000110000000000
		// 010000001000000001000000110000000000
		expected := []byte{
			0x80, 0xc0, 0xc0, 0xc0, 0xe0,
			0x40, 0x80, 0x80, 0xc0, 0xa0,
			0x40, 0x80, 0xc0, 0xc0, 0x80,
			0x00, 0x80, 0xc0, 0xc0, 0x80,
			0x80, 0xc0, 0xc0, 0xc0, 0x00,
			0x80, 0x40, 0xc0, 0xc0, 0x00,
			0x40, 0xc0, 0xc0, 0xc0, 0x00,
			0x40, 0x80, 0x40, 0xc0, 0x00,
		}
		assert.Equal(t, expected, bm.Data)
		assert.True(t, bm == d)
	})

	t.Run("NilSrc", func(t *testing.T) {
		_, err := DilateBrick(New(36, 8), nil, 5, 5)
		assert.Error(t, err)
	})
}

// TestErode tests the erode morph function.
func TestErode(t *testing.T) {
	// Let's have a test 36×8 bitmap with the following data:
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

	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	sel := SelCreateBrick(1, 1, 1, 0, SelHit)

	t.Run("Symmetric", func(t *testing.T) {
		MorphBC = SymmetricMorphBC

		t.Run("NilDst", func(t *testing.T) {
			d, err := erode(nil, s, sel)
			require.NoError(t, err)

			assert.NotNil(t, d)

			// Expected result:
			//  111111111111111111111111111111111111
			//  100000001100000011000000110000001110
			//  000000000000000010000000110000001010
			//  010000001000000010000000110000000000
			//  000000001000000011000000110000001000
			//  000000001000000001000000110000000000
			//  100000000100000011000000110000000000
			//  000000000100000011000000110000000000
			expected := []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xF0,
				0x80, 0xC0, 0xC0, 0xC0, 0xE0,
				0x00, 0x00, 0x80, 0xC0, 0xA0,
				0x40, 0x80, 0x80, 0xC0, 0x00,
				0x00, 0x80, 0xC0, 0xC0, 0x80,
				0x00, 0x80, 0x40, 0xC0, 0x00,
				0x80, 0x40, 0xC0, 0xC0, 0x00,
				0x00, 0x40, 0xC0, 0xC0, 0x00,
			}
			assert.Equal(t, expected, d.Data)
		})

		t.Run("Single", func(t *testing.T) {
			data := []byte{
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
			}
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111

			s, err := NewWithData(36, 8, data)
			require.NoError(t, err)

			sel := SelCreateBrick(3, 3, 1, 1, SelHit)

			d, err := erode(nil, s, sel)
			require.NoError(t, err)

			// Expected result:
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111

			expected := []byte{
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
			}
			assert.Equal(t, expected, d.Data)
		})
	})

	t.Run("Asymmetric", func(t *testing.T) {
		MorphBC = AsymmetricMorphBC

		t.Run("NilDst", func(t *testing.T) {
			d, err := erode(nil, s, sel)
			require.NoError(t, err)

			assert.NotNil(t, d)

			// Expected result:
			//  000000000000000000000000000000000000
			//  100000001100000011000000110000001110
			//  000000000000000010000000110000001010
			//  010000001000000010000000110000000000
			//  000000001000000011000000110000001000
			//  000000001000000001000000110000000000
			//  100000000100000011000000110000000000
			//  000000000100000011000000110000000000
			expected := []byte{
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x80, 0xC0, 0xC0, 0xC0, 0xE0,
				0x00, 0x00, 0x80, 0xC0, 0xA0,
				0x40, 0x80, 0x80, 0xC0, 0x00,
				0x00, 0x80, 0xC0, 0xC0, 0x80,
				0x00, 0x80, 0x40, 0xC0, 0x00,
				0x80, 0x40, 0xC0, 0xC0, 0x00,
				0x00, 0x40, 0xC0, 0xC0, 0x00,
			}
			assert.Equal(t, expected, d.Data)
		})

		t.Run("Single", func(t *testing.T) {
			data := []byte{
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
			}
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111

			s, err := NewWithData(36, 8, data)
			require.NoError(t, err)

			sel := SelCreateBrick(3, 3, 1, 1, SelHit)

			d, err := erode(nil, s, sel)
			require.NoError(t, err)

			// Expected result:
			// 000000000000000000000000000000000000
			// 011111111111111111111111111111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111111111111111110
			// 011111111111111111111111111111111110
			// 000000000000000000000000000000000000
			//
			expected := []byte{
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x00, 0x00, 0x00, 0x00, 0x00,
			}
			assert.Equal(t, expected, d.Data, d.String())
		})
	})

	t.Run("NilSrc", func(t *testing.T) {
		d := New(36, 8)
		_, err := erode(d, nil, sel)
		require.Error(t, err)
	})
}

// TestErodeBrick tests the erode brick morph function.
func TestErodeBrick(t *testing.T) {
	// Having a test data for the erode brick method:
	//
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111110111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	data := []byte{
		0xff, 0xff, 0xff, 0xff, 0xf0,
		0xff, 0xff, 0xff, 0xff, 0xf0,
		0xff, 0xff, 0xff, 0xff, 0xf0,
		0xff, 0xff, 0xfe, 0xff, 0xf0,
		0xff, 0xff, 0xff, 0xff, 0xf0,
		0xff, 0xff, 0xff, 0xff, 0xf0,
		0xff, 0xff, 0xff, 0xff, 0xf0,
		0xff, 0xff, 0xff, 0xff, 0xf0,
	}

	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	t.Run("Size0", func(t *testing.T) {
		bm := New(36, 8)
		_, err := erodeBrick(bm, s, 0, 1)
		assert.Error(t, err)
	})

	t.Run("BothSize1", func(t *testing.T) {
		bm := New(36, 8)
		d, err := erodeBrick(bm, s, 1, 1)
		require.NoError(t, err)

		assert.False(t, d == s)
		assert.Equal(t, d, s)
	})

	t.Run("Symmetric", func(t *testing.T) {
		MorphBC = SymmetricMorphBC

		t.Run("NilDst", func(t *testing.T) {
			d, err := erodeBrick(nil, s, 3, 3)
			assert.NoError(t, err)

			// Expected result:
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			//
			expected := []byte{
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
			}
			assert.Equal(t, expected, d.Data)
		})

		t.Run("WithDest", func(t *testing.T) {
			bm := New(36, 8)
			d, err := erodeBrick(bm, s, 3, 3)
			assert.NoError(t, err)

			// Expected result:
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111100011111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			//
			expected := []byte{
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xfc, 0x7f, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
			}
			assert.Equal(t, expected, d.Data)
		})

		t.Run("OneSize1", func(t *testing.T) {
			bm := New(36, 8)
			d, err := erodeBrick(bm, s, 1, 3)
			require.NoError(t, err)

			// Expected result:
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			//
			expected := []byte{
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
			}
			assert.Equal(t, expected, d.Data, d.String())
		})
	})

	t.Run("Asymmetric", func(t *testing.T) {
		MorphBC = AsymmetricMorphBC

		t.Run("NilDst", func(t *testing.T) {
			d, err := erodeBrick(nil, s, 3, 3)
			assert.NoError(t, err)

			// Expected result:
			// 000000000000000000000000000000000000
			// 011111111111111111111111111111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111111111111111110
			// 011111111111111111111111111111111110
			// 000000000000000000000000000000000000
			//
			expected := []byte{
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x00, 0x00, 0x00, 0x00, 0x00,
			}
			assert.Equal(t, expected, d.Data)
		})

		t.Run("WithDest", func(t *testing.T) {
			bm := New(36, 8)
			d, err := erodeBrick(bm, s, 3, 3)
			assert.NoError(t, err)

			// Expected result:
			// 000000000000000000000000000000000000
			// 011111111111111111111111111111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111100011111111110
			// 011111111111111111111111111111111110
			// 011111111111111111111111111111111110
			// 000000000000000000000000000000000000
			//
			expected := []byte{
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xfc, 0x7f, 0xe0,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x7f, 0xff, 0xff, 0xff, 0xe0,
				0x00, 0x00, 0x00, 0x00, 0x00,
			}
			assert.Equal(t, expected, d.Data)
		})

		t.Run("OneSize1", func(t *testing.T) {
			bm := New(36, 8)
			d, err := erodeBrick(bm, s, 1, 3)
			require.NoError(t, err)

			// Expected result:
			// 000000000000000000000000000000000000
			// 111111111111111111111111111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111110111111111111
			// 111111111111111111111111111111111111
			// 111111111111111111111111111111111111
			// 000000000000000000000000000000000000
			//
			expected := []byte{
				0x00, 0x00, 0x00, 0x00, 0x00,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xfe, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0xff, 0xff, 0xff, 0xf0,
				0x00, 0x00, 0x00, 0x00, 0x00,
			}
			assert.Equal(t, expected, d.Data, d.String())
		})
	})
}

// TestOpen tests the opening functions for the bitmap.
func TestOpen(t *testing.T) {
	// Let's have a test 36×8 bitmap with the following data:
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

	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NilSrc", func(t *testing.T) {
			_, err := open(nil, nil, SelCreateBrick(1, 1, 1, 1, SelHit))
			assert.Error(t, err)
		})

		t.Run("Selection", func(t *testing.T) {
			_, err := open(nil, s, SelCreateBrick(0, 1, 1, 1, SelHit))
			assert.Error(t, err)
		})
	})

	sel := SelCreateBrick(1, 1, 3, 0, SelHit)

	t.Run("Asymmetric", func(t *testing.T) {
		MorphBC = AsymmetricMorphBC

		t.Run("Valid", func(t *testing.T) {
			d, err := open(nil, s, sel)
			require.NoError(t, err)

			// Expected result should look like:
			//
			// 100000001100000011000000110000001110
			// 000000000000000010000000110000001010
			// 010000001000000010000000110000000000
			// 000000001000000011000000110000001000
			// 000000001000000001000000110000000000
			// 000000000000000000000000000000000000
			// 000000000000000000000000000000000000
			// 000000000000000000000000000000000000
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X00, 0X00, 0X80, 0XC0, 0XA0,
				0X40, 0X80, 0X80, 0XC0, 0X00,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0X40, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0X00, 0X00,
				0X00, 0X00, 0X00, 0X00, 0X00,
				0X00, 0X00, 0X00, 0X00, 0X00,
			}
			assert.Equal(t, expected, d.Data)
		})
	})

	t.Run("Symmetric", func(t *testing.T) {
		MorphBC = SymmetricMorphBC

		t.Run("Valid", func(t *testing.T) {
			d, err := open(nil, s, sel)
			require.NoError(t, err)

			// Expected result should look like:
			//
			// 100000001100000011000000110000001110
			// 000000000000000010000000110000001010
			// 010000001000000010000000110000000000
			// 000000001000000011000000110000001000
			// 000000001000000001000000110000000000
			// 000000000000000000000000000000000000
			// 000000000000000000000000000000000000
			// 000000000000000000000000000000000000
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X00, 0X00, 0X80, 0XC0, 0XA0,
				0X40, 0X80, 0X80, 0XC0, 0X00,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0X40, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0X00, 0X00,
				0X00, 0X00, 0X00, 0X00, 0X00,
				0X00, 0X00, 0X00, 0X00, 0X00,
			}
			assert.Equal(t, expected, d.Data)
		})
	})
}

// TestOpenBrick tests the open brick morph function.
func TestOpenBrick(t *testing.T) {
	// Having a test data for the erode brick method:
	//
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111110111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	// 111111111111111111111111111111111111
	//
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

	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	t.Run("NilSource", func(t *testing.T) {
		_, err = openBrick(nil, nil, 1, 1)
		assert.Error(t, err)
	})

	t.Run("Hsize&&Vsize < 0", func(t *testing.T) {
		_, err = openBrick(nil, s, 0, 0)
		assert.Error(t, err)
	})

	t.Run("Hsize==Vsize==1", func(t *testing.T) {
		d, err := openBrick(nil, s, 1, 1)
		require.NoError(t, err)

		assert.Equal(t, d, s)
		assert.False(t, d == s)
	})

	t.Run("Symmetric", func(t *testing.T) {
		MorphBC = SymmetricMorphBC

		t.Run("Hsize||vSize==1", func(t *testing.T) {
			d, err := openBrick(nil, s, 1, 2)
			require.NoError(t, err)

			// Expected result should look like:
			//
			// 100000001100000011000000110000001110
			// 000000000000000010000000110000001010
			// 000000001000000010000000110000000000
			// 000000001000000011000000110000000000
			// 000000001000000001000000110000000000
			// 000000000100000011000000110000000000
			// 000000000100000011000000110000000000
			// 000000000000000001000000110000000000
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X00, 0X00, 0X80, 0XC0, 0XA0,
				0X00, 0X80, 0X80, 0XC0, 0X00,
				0X00, 0X80, 0XC0, 0XC0, 0X00,
				0X00, 0X80, 0X40, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X00, 0X00, 0X40, 0XC0, 0X00,
			}
			assert.Equal(t, expected, d.Data, d.String())
		})

		t.Run("Valid", func(t *testing.T) {
			d, err := openBrick(nil, s, 2, 2)
			require.NoError(t, err)

			// Expected result should look like:
			//
			// 100000001100000011000000110000001110
			// 000000000000000000000000110000000000
			// 000000000000000000000000110000000000
			// 000000000000000000000000110000000000
			// 000000000000000000000000110000000000
			// 000000000000000011000000110000000000
			// 000000000000000011000000110000000000
			// 000000000000000000000000110000000000
			//
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0XC0, 0XC0, 0X00,
				0X00, 0X00, 0XC0, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
			}
			assert.Equal(t, expected, d.Data, d.String())
		})
	})

	t.Run("Asymmetric", func(t *testing.T) {
		MorphBC = AsymmetricMorphBC

		t.Run("Hsize||vSize==1", func(t *testing.T) {
			d, err := openBrick(New(36, 8), s, 1, 2)
			require.NoError(t, err)

			// Expected result should look like:
			//
			// 000000000000000010000000110000001010
			// 000000000000000010000000110000001010
			// 000000001000000010000000110000000000
			// 000000001000000011000000110000000000
			// 000000001000000001000000110000000000
			// 000000000100000011000000110000000000
			// 000000000100000011000000110000000000
			// 000000000000000001000000110000000000
			expected := []byte{
				0x00, 0x00, 0x80, 0xc0, 0xa0,
				0x00, 0x00, 0x80, 0xc0, 0xa0,
				0x00, 0x80, 0x80, 0xc0, 0x00,
				0x00, 0x80, 0xc0, 0xc0, 0x00,
				0x00, 0x80, 0x40, 0xc0, 0x00,
				0x00, 0x40, 0xc0, 0xc0, 0x00,
				0x00, 0x40, 0xc0, 0xc0, 0x00,
				0x00, 0x00, 0x40, 0xc0, 0x00,
			}
			assert.Equal(t, expected, d.Data, d.String())
		})

		t.Run("Valid", func(t *testing.T) {
			d, err := openBrick(nil, s, 2, 2)
			require.NoError(t, err)

			// Expected result should look like:
			//
			// 000000000000000000000000110000000000
			// 000000000000000000000000110000000000
			// 000000000000000000000000110000000000
			// 000000000000000000000000110000000000
			// 000000000000000000000000110000000000
			// 000000000000000011000000110000000000
			// 000000000000000011000000110000000000
			// 000000000000000000000000110000000000
			//
			expected := []byte{
				0x00, 0x00, 0x00, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
				0X00, 0X00, 0XC0, 0XC0, 0X00,
				0X00, 0X00, 0XC0, 0XC0, 0X00,
				0X00, 0X00, 0X00, 0XC0, 0X00,
			}
			assert.Equal(t, expected, d.Data, d.String())
		})
	})
}

// TestClose tests the close morph functions.
func TestClose(t *testing.T) {
	// Let's have a test 36×8 bitmap with the following data:
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

	// define the source bitmap
	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	// create selection
	sel := SelCreateBrick(1, 1, 2, 2, SelHit)

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NilSource", func(t *testing.T) {
			sel := SelCreateBrick(1, 1, 2, 2, SelHit)

			_, err := closeBitmap(nil, nil, sel)
			assert.Error(t, err)
		})

		t.Run("Selelction", func(t *testing.T) {
			_, err = closeBitmap(nil, s, SelCreateBrick(0, 1, 1, 1, SelHit))
			assert.Error(t, err)
		})
	})

	t.Run("Symmetric", func(t *testing.T) {
		MorphBC = SymmetricMorphBC

		d, err := closeBitmap(nil, s, sel)
		require.NoError(t, err)

		// Expected result should look like:
		//
		// 111111111111111111111111111111111111
		// 111111111111111111111111111111111111
		// 110000001000000010000000110000000000
		// 110000001000000011000000110000001000
		// 110000001000000001000000110000000000
		// 110000000100000011000000110000000000
		// 110000000100000011000000110000000000
		// 110000001000000001000000110000000000

		expected := []byte{
			0XFF, 0XFF, 0XFF, 0XFF, 0XF0,
			0XFF, 0XFF, 0XFF, 0XFF, 0XF0,
			0XC0, 0X80, 0X80, 0XC0, 0X00,
			0XC0, 0X80, 0XC0, 0XC0, 0X80,
			0XC0, 0X80, 0X40, 0XC0, 0X00,
			0XC0, 0X40, 0XC0, 0XC0, 0X00,
			0XC0, 0X40, 0XC0, 0XC0, 0X00,
			0XC0, 0X80, 0X40, 0XC0, 0X00,
		}
		assert.Equal(t, expected, d.Data, "Bytes: % #02X\n %s", d.Data, d.String())
	})

	t.Run("Asymmetric", func(t *testing.T) {
		MorphBC = AsymmetricMorphBC
		d, err := closeBitmap(nil, s, sel)
		require.NoError(t, err)

		// Expected result should look like:
		//
		//  000000000000000000000000000000000000
		//  000000000000000000000000000000000000
		//  000000001000000010000000110000000000
		//  000000001000000011000000110000001000
		//  000000001000000001000000110000000000
		//  000000000100000011000000110000000000
		//  000000000100000011000000110000000000
		//  000000001000000001000000110000000000
		expected := []byte{
			0X00, 0X00, 0X00, 0X00, 0X00,
			0X00, 0X00, 0X00, 0X00, 0X00,
			0X00, 0X80, 0X80, 0XC0, 0X00,
			0X00, 0X80, 0XC0, 0XC0, 0X80,
			0X00, 0X80, 0X40, 0XC0, 0X00,
			0X00, 0X40, 0XC0, 0XC0, 0X00,
			0X00, 0X40, 0XC0, 0XC0, 0X00,
			0X00, 0X80, 0X40, 0XC0, 0X00,
		}
		assert.Equal(t, expected, d.Data, "Bytes: % #02X\n %s", d.Data, d.String())
	})
}

// TestCloseBrick tests the close brick morph function.
func TestCloseBrick(t *testing.T) {
	// Let's have a test 36×8 bitmap with the following data:
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

	// define the source bitmap
	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NilSource", func(t *testing.T) {
			_, err := closeBrick(nil, nil, 2, 3)
			assert.Error(t, err)
		})

		t.Run("Hsize||Vsize<1", func(t *testing.T) {
			_, err = closeBrick(nil, s, 0, 1)
			assert.Error(t, err)
		})
	})

	t.Run("Hsize&&Vsize==1", func(t *testing.T) {
		d, err := closeBrick(nil, s, 1, 1)
		require.NoError(t, err)

		assert.Equal(t, d, s)
		assert.False(t, s == d)
	})

	t.Run("Symmetric", func(t *testing.T) {
		MorphBC = SymmetricMorphBC

		t.Run("Hsize||Vsize==1", func(t *testing.T) {
			d, err := closeBrick(nil, s, 1, 2)
			require.NoError(t, err)

			assert.NotNil(t, d)

			// 100000001100000011000000110000001110
			// 000000001000000010000000110000001010
			// 010000001000000010000000110000001000
			// 000000001000000011000000110000001000
			// 000000001000000011000000110000000000
			// 100000000100000011000000110000000000
			// 000000000100000011000000110000000000
			// 010000001000000001000000110000000000
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X00, 0X80, 0X80, 0XC0, 0XA0,
				0X40, 0X80, 0X80, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X00,
				0X80, 0X40, 0XC0, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X40, 0X80, 0X40, 0XC0, 0X00,
			}
			assert.Equal(t, expected, d.Data)
		})
		t.Run("Regular", func(t *testing.T) {
			d, err := closeBrick(nil, s, 2, 2)
			require.NoError(t, err)

			// 100000001100000011000000110000001110
			// 100000001000000010000000110000001110
			// 110000001000000010000000110000001000
			// 000000001000000011000000110000001000
			// 000000001000000011000000110000000000
			// 100000000100000011000000110000000000
			// 100000000100000011000000110000000000
			// 110000001000000001000000110000000000
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X80, 0X80, 0X80, 0XC0, 0XE0,
				0XC0, 0X80, 0X80, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X00,
				0X80, 0X40, 0XC0, 0XC0, 0X00,
				0X80, 0X40, 0XC0, 0XC0, 0X00,
				0XC0, 0X80, 0X40, 0XC0, 0X00,
			}
			assert.Equal(t, expected, d.Data)
		})
	})

	t.Run("Asymmetric", func(t *testing.T) {
		MorphBC = AsymmetricMorphBC

		t.Run("Hsize||Vsize==1", func(t *testing.T) {
			d, err := closeBrick(nil, s, 1, 2)
			require.NoError(t, err)

			// 000000000000000000000000000000000000
			// 000000001000000010000000110000001010
			// 010000001000000010000000110000001000
			// 000000001000000011000000110000001000
			// 000000001000000011000000110000000000
			// 100000000100000011000000110000000000
			// 000000000100000011000000110000000000
			// 010000001000000001000000110000000000
			expected := []byte{
				0X00, 0X00, 0X00, 0X00, 0X00,
				0X00, 0X80, 0X80, 0XC0, 0XA0,
				0X40, 0X80, 0X80, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X00,
				0X80, 0X40, 0XC0, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X40, 0X80, 0X40, 0XC0, 0X00,
			}
			assert.Equal(t, expected, d.Data)
		})

		t.Run("Regular", func(t *testing.T) {
			d, err := closeBrick(nil, s, 2, 2)
			require.NoError(t, err)
			//  000000000000000000000000000000000000
			//  000000001000000010000000110000001110
			//  010000001000000010000000110000001000
			//  000000001000000011000000110000001000
			//  000000001000000011000000110000000000
			//  000000000100000011000000110000000000
			//  000000000100000011000000110000000000
			//  010000001000000001000000110000000000
			expected := []byte{
				0X00, 0X00, 0X00, 0X00, 0X00,
				0X00, 0X80, 0X80, 0XC0, 0XE0,
				0X40, 0X80, 0X80, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X40, 0X80, 0X40, 0XC0, 0X00,
			}
			assert.Equal(t, expected, d.Data)
		})
	})
}

// TestCloseSafeBrick tests closeSafeBrick function.
func TestCloseSafeBrick(t *testing.T) {
	// Let's have a test 36×8 bitmap with the following data:
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

	// define the source bitmap
	s, err := NewWithData(36, 8, data)
	require.NoError(t, err)

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NilSrc", func(t *testing.T) {

		})

		t.Run("HSize||VSize<1", func(t *testing.T) {

		})
	})

	t.Run("HSize&&VSize==1", func(t *testing.T) {
		d, err := closeSafeBrick(nil, s, 1, 1)
		require.NoError(t, err)

		// Check if the result bitmap is a copy
		assert.Equal(t, d, s)
		assert.False(t, d == s)
	})

	t.Run("Symmetric", func(t *testing.T) {
		MorphBC = SymmetricMorphBC

		d, err := closeSafeBrick(nil, s, 2, 2)
		require.NoError(t, err)

		// 100000001100000011000000110000001110
		// 100000001000000010000000110000001110
		// 110000001000000010000000110000001000
		// 000000001000000011000000110000001000
		// 000000001000000011000000110000000000
		// 100000000100000011000000110000000000
		// 100000000100000011000000110000000000
		// 110000001000000001000000110000000000
		expected := []byte{
			0X80, 0XC0, 0XC0, 0XC0, 0XE0,
			0X80, 0X80, 0X80, 0XC0, 0XE0,
			0XC0, 0X80, 0X80, 0XC0, 0X80,
			0X00, 0X80, 0XC0, 0XC0, 0X80,
			0X00, 0X80, 0XC0, 0XC0, 0X00,
			0X80, 0X40, 0XC0, 0XC0, 0X00,
			0X80, 0X40, 0XC0, 0XC0, 0X00,
			0XC0, 0X80, 0X40, 0XC0, 0X00,
		}
		assert.Equal(t, expected, d.Data)
	})

	t.Run("Asymmetric", func(t *testing.T) {
		MorphBC = AsymmetricMorphBC

		t.Run("Hsize||Vsize==1", func(t *testing.T) {
			d, err := closeSafeBrick(nil, s, 2, 2)
			require.NoError(t, err)

			// 100000001100000011000000110000001110
			// 000000001000000010000000110000001110
			// 010000001000000010000000110000001000
			// 000000001000000011000000110000001000
			// 000000001000000011000000110000000000
			// 100000000100000011000000110000000000
			// 000000000100000011000000110000000000
			// 010000001000000001000000110000000000
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X00, 0X80, 0X80, 0XC0, 0XE0,
				0X40, 0X80, 0X80, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X00,
				0X80, 0X40, 0XC0, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X40, 0X80, 0X40, 0XC0, 0X00,
			}

			assert.Equal(t, expected, d.Data)
		})

		t.Run("Regular", func(t *testing.T) {
			d, err := closeSafeBrick(nil, s, 2, 2)
			require.NoError(t, err)
			// 100000001100000011000000110000001110
			// 000000001000000010000000110000001110
			// 010000001000000010000000110000001000
			// 000000001000000011000000110000001000
			// 000000001000000011000000110000000000
			// 100000000100000011000000110000000000
			// 000000000100000011000000110000000000
			// 010000001000000001000000110000000000
			expected := []byte{
				0X80, 0XC0, 0XC0, 0XC0, 0XE0,
				0X00, 0X80, 0X80, 0XC0, 0XE0,
				0X40, 0X80, 0X80, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X80,
				0X00, 0X80, 0XC0, 0XC0, 0X00,
				0X80, 0X40, 0XC0, 0XC0, 0X00,
				0X00, 0X40, 0XC0, 0XC0, 0X00,
				0X40, 0X80, 0X40, 0XC0, 0X00,
			}

			assert.Equal(t, expected, d.Data)
			// t.Errorf("%s, % #02X", d.String(), d.Data)
		})
	})
}
