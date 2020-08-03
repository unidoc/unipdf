/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrelationThreshold(t *testing.T) {
	t.Run("InThreshold", func(t *testing.T) {
		// Having a 100x100 all filled bitmap
		s := New(100, 100)
		err := s.RasterOperation(0, 0, s.Width, s.Height, PixNotDst, nil, 0, 0)
		require.NoError(t, err)

		// in order to have 95% threshold
		// the 9500 bits should be at least set to 'ON'.
		tp := s.Copy()
		// set randomly 500 bits to false.
		count := 499
		match := map[int]map[int]struct{}{}

		for count != 0 {
			col := rand.Intn(100)
			bitIndex := rand.Intn(100)
			_, ok := match[col]
			if !ok {
				match[col] = map[int]struct{}{}
			} else {
				_, ok = match[col][bitIndex]
				if ok {
					continue
				}
			}
			match[col][bitIndex] = struct{}{}
			assert.NoError(t, tp.SetPixel(bitIndex, col, 0))
			count--
		}
		// downcount contains the amount of the pixels
		downcount := make([]int, 100)
		var dc int
		for y := s.Height - 1; y >= 0; y-- {
			downcount[y] = dc
			dc += 100
		}
		sumtab := MakePixelSumTab8()

		// the correlation score should be 0.9501 as there are 9501 common pixels for these images.
		score, err := CorrelationScore(s, tp, 100*100, 100*100-499, 0, 0, 20, 20, sumtab)
		require.NoError(t, err)

		assert.InDelta(t, 0.9501, score, 0.0001)

		simpleScore, err := CorrelationScoreSimple(s, tp, 100*100, 100*100-499, 0, 0, 20, 20, sumtab)
		require.NoError(t, err)

		assert.InDelta(t, 0.9501, simpleScore, 0.0001)

		// the result of the function with  '0.9500' score_threshold should be true.
		ok, err := CorrelationScoreThresholded(s, tp, 100*100, 100*100-499, 0, 0, 20, 20, sumtab, downcount, 0.9500)
		require.NoError(t, err)

		assert.True(t, ok, tp.String())

		// this function should not pass the score of 0.9502
		ok, err = CorrelationScoreThresholded(s, tp, 100*100, 100*100-499, 0, 0, 0, 0, sumtab, downcount, 0.9502)
		require.NoError(t, err)

		assert.False(t, ok)
	})

	t.Run("ShiftedLeft", func(t *testing.T) {
		t.Run("Minor", func(t *testing.T) {
			// Let's createa a 100x100 bitmap with an internal 80x80 block.
			//
			// at first create a 80x80 bitmap.
			s := New(80, 80)
			// fill it with the bits
			err := s.RasterOperation(0, 0, s.Width, s.Height, PixSet, nil, 0, 0)
			require.NoError(t, err)

			// add the border of size '10' and value 'OFF'
			s, err = s.AddBorder(10, 0)
			require.NoError(t, err)

			// create a bitmap 100x100 with internal bitmap 80x80 that is located 5 bits further
			// to the UL corner than the 's' bitmap.
			d := New(80, 80)
			err = d.RasterOperation(0, 0, d.Width, d.Height, PixSet, nil, 0, 0)
			require.NoError(t, err)

			// add the border directed more within UL direction (left, top= 5; right, bottom =15)
			d, err = d.AddBorderGeneral(5, 15, 5, 15, 0)
			require.NoError(t, err)

			require.Equal(t, s.Width, d.Width)
			require.Equal(t, s.Height, d.Height)

			centroids, err := Centroids([]*Bitmap{s, d})
			require.NoError(t, err)

			// create the downcount for the 's' bitmap.
			downcount := make([]int, 100)
			var count int
			for y := s.Height - 1; y >= 0; y-- {
				switch y {
				case 99, 98, 1, 0:
				default:
					count += 80
				}
				downcount[y] = count
			}

			sumtab := MakePixelSumTab8()

			var dx, dy float32
			dx = (*centroids)[0].X - (*centroids)[1].X
			dy = (*centroids)[0].Y - (*centroids)[1].Y

			score, err := CorrelationScore(s, d, 80*80, 80*80, dx, dy, 10, 10, sumtab)
			require.NoError(t, err)

			simpleScore, err := CorrelationScoreSimple(s, d, 80*80, 80*80, dx, dy, 10, 10, sumtab)
			require.NoError(t, err)

			assert.InDelta(t, score, simpleScore, 0.00001)

			inThreshold, err := CorrelationScoreThresholded(s, d, 80*80, 80*80, dx, dy, 10, 10, sumtab, downcount, 0.99)
			require.NoError(t, err)

			assert.True(t, inThreshold)
		})

		t.Run("Major", func(t *testing.T) {
			// Let's createa a 1000x1000 bitmap with an internal 80x80 block.
			//
			// at first create a 800x800 bitmap.
			s := New(800, 800)
			// fill it with the bits
			err := s.RasterOperation(0, 0, s.Width, s.Height, PixSet, nil, 0, 0)
			require.NoError(t, err)

			// add the border of size '100' and value 'OFF'
			s, err = s.AddBorder(100, 0)
			require.NoError(t, err)

			// create a bitmap 1000x1000 with internal bitmap 800x800 that is located 50 bits further
			// to the BR corner than the 's' bitmap.
			d := New(800, 800)
			err = d.RasterOperation(0, 0, d.Width, d.Height, PixSet, nil, 0, 0)
			require.NoError(t, err)

			// add the border directed more within UL direction (left, top= 50; right, bottom =150)
			d, err = d.AddBorderGeneral(50, 150, 50, 150, 0)
			require.NoError(t, err)

			require.Equal(t, s.Width, d.Width)
			require.Equal(t, s.Height, d.Height)

			centroids, err := Centroids([]*Bitmap{s, d})
			require.NoError(t, err)

			// create the downcount for the 's' bitmap.
			downcount := make([]int, 1000)
			var count int
			for y := s.Height - 1; y >= 0; y-- {
				if y >= 100 || y < 900 {
					count += 800
				}
				downcount[y] = count
			}

			sumtab := MakePixelSumTab8()

			var dx, dy float32
			dx = (*centroids)[0].X - (*centroids)[1].X
			dy = (*centroids)[0].Y - (*centroids)[1].Y

			score, err := CorrelationScore(s, d, 800*800, 800*800, dx, dy, 10, 10, sumtab)
			require.NoError(t, err)

			simpleScore, err := CorrelationScoreSimple(s, d, 800*800, 800*800, dx, dy, 10, 10, sumtab)
			require.NoError(t, err)

			assert.InDelta(t, score, simpleScore, 0.00001)

			inThreshold, err := CorrelationScoreThresholded(s, d, 800*800, 800*800, dx, dy, 10, 10, sumtab, downcount, 0.99)
			require.NoError(t, err)

			assert.True(t, inThreshold)
		})
	})

	t.Run("ShiftedRight", func(t *testing.T) {
		t.Run("Minor", func(t *testing.T) {
			// Let's createa a 100x100 bitmap with an internal 80x80 block.
			//
			// at first create a 80x80 bitmap.
			s := New(80, 80)
			// fill it with the bits
			err := s.RasterOperation(0, 0, s.Width, s.Height, PixSet, nil, 0, 0)
			require.NoError(t, err)

			// add the border of size '10' and value 'OFF'
			s, err = s.AddBorder(10, 0)
			require.NoError(t, err)

			// create a bitmap 100x100 with internal bitmap 80x80 that is located 5 bits further
			// to the BR corner than the 's' bitmap.
			d := New(80, 80)
			err = d.RasterOperation(0, 0, d.Width, d.Height, PixSet, nil, 0, 0)
			require.NoError(t, err)

			// add the border directed more within UL direction (left, top= 15; right, bottom =5)
			d, err = d.AddBorderGeneral(15, 5, 15, 5, 0)
			require.NoError(t, err)

			require.Equal(t, s.Width, d.Width)
			require.Equal(t, s.Height, d.Height)

			centroids, err := Centroids([]*Bitmap{s, d})
			require.NoError(t, err)

			// create the downcount for the 's' bitmap.
			downcount := make([]int, 100)
			var count int
			for y := s.Height - 1; y >= 0; y-- {
				switch y {
				case 99, 98, 1, 0:
				default:
					count += 80
				}
				downcount[y] = count
			}

			sumtab := MakePixelSumTab8()

			var dx, dy float32
			dx = (*centroids)[0].X - (*centroids)[1].X
			dy = (*centroids)[0].Y - (*centroids)[1].Y

			score, err := CorrelationScore(s, d, 80*80, 80*80, dx, dy, 10, 10, sumtab)
			require.NoError(t, err)

			simpleScore, err := CorrelationScoreSimple(s, d, 80*80, 80*80, dx, dy, 10, 10, sumtab)
			require.NoError(t, err)

			assert.InDelta(t, score, simpleScore, 0.00001)

			inThreshold, err := CorrelationScoreThresholded(s, d, 80*80, 80*80, dx, dy, 10, 10, sumtab, downcount, 0.99)
			require.NoError(t, err)

			assert.True(t, inThreshold)
		})

		t.Run("Major", func(t *testing.T) {
			// Let's createa a 1000x1000 bitmap with an internal 80x80 block.
			//
			// at first create a 800x800 bitmap.
			s := New(800, 800)
			// fill it with the bits
			err := s.RasterOperation(0, 0, s.Width, s.Height, PixSet, nil, 0, 0)
			require.NoError(t, err)

			// add the border of size '100' and value 'OFF'
			s, err = s.AddBorder(100, 0)
			require.NoError(t, err)
			for i := 0; i <= 200; i += 5 {
				leftTop := i
				rightBot := 200 - i
				t.Run(strconv.Itoa(leftTop)+"x"+strconv.Itoa(rightBot), func(t *testing.T) {
					// create a bitmap 1000x1000 with internal bitmap 800x800 that is located to the BR corner than the 's' bitmap.
					d := New(800, 800)
					err = d.RasterOperation(0, 0, d.Width, d.Height, PixSet, nil, 0, 0)
					require.NoError(t, err)

					// add the border directed more within UL direction (left, top= 150; right, bottom =50)
					d, err = d.AddBorderGeneral(leftTop, rightBot, leftTop, rightBot, 0)
					require.NoError(t, err)

					require.Equal(t, s.Width, d.Width)
					require.Equal(t, s.Height, d.Height)

					centroids, err := Centroids([]*Bitmap{s, d})
					require.NoError(t, err)

					// create the downcount for the 's' bitmap.
					downcount := make([]int, 1000)
					var count int
					for y := s.Height - 1; y >= 0; y-- {
						if y >= 100 || y < 900 {
							count += 800
						}
						downcount[y] = count
					}

					sumtab := MakePixelSumTab8()

					var dx, dy float32
					dx = (*centroids)[0].X - (*centroids)[1].X
					dy = (*centroids)[0].Y - (*centroids)[1].Y

					score, err := CorrelationScore(s, d, 800*800, 800*800, dx, dy, 10, 10, sumtab)
					require.NoError(t, err)

					simpleScore, err := CorrelationScoreSimple(s, d, 800*800, 800*800, dx, dy, 10, 10, sumtab)
					require.NoError(t, err)

					assert.InDelta(t, score, simpleScore, 0.00001)

					inThreshold, err := CorrelationScoreThresholded(s, d, 800*800, 800*800, dx, dy, 10, 10, sumtab, downcount, 0.99)
					require.NoError(t, err)

					assert.True(t, inThreshold)
				})
			}
		})
	})
}

// TestHausdorffChecks checks the HausTest and RankHausTest functions.
func TestHausdorffChecks(t *testing.T) {

}
