/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelectBitmapsBySize tests the select bitmaps by size function
func TestSelectBitmapsBySize(t *testing.T) {
	type size struct {
		width, height int
	}

	t.Run("WithBoxes", func(t *testing.T) {
		t.Run("AllWithin", func(t *testing.T) {
			sizes := []size{{5, 8}, {6, 9}, {5, 5}, {3, 7}}

			type testParam struct {
				name          string
				width, height int
				location      LocationFilter
				comparison    SizeComparison
			}

			tests := []testParam{
				// each width is within bounds, but height not,
				{"WidthLTE", 6, 2, LocSelectWidth, SizeSelectIfLTE},
				{"WidthLT", 7, 2, LocSelectWidth, SizeSelectIfLT},
				{"WidthGT", 2, 10, LocSelectWidth, SizeSelectIfGT},
				{"WidthGTE", 3, 10, LocSelectWidth, SizeSelectIfGTE},
				// each height is within bounds but width not.
				{"HeightLTE", 2, 9, LocSelectHeight, SizeSelectIfLTE},
				{"HeightLT", 2, 10, LocSelectHeight, SizeSelectIfLT},
				{"HeightGT", 10, 4, LocSelectHeight, SizeSelectIfGT},
				{"HeightGTE", 10, 5, LocSelectHeight, SizeSelectIfGTE},
				// either width or height fits the bounds.
				{"EitherLTEWidth", 6, 3, LocSelectIfEither, SizeSelectIfLTE},
				{"EitherLTEHeight", 4, 10, LocSelectIfEither, SizeSelectIfLTE},
				{"EitherGTEWidth", 3, 10, LocSelectIfEither, SizeSelectIfGTE},
				{"EitherGTEHeight", 2, 5, LocSelectIfEither, SizeSelectIfGTE},
				{"EitherGTWidth", 2, 10, LocSelectIfEither, SizeSelectIfGT},
				{"EitherGTHeight", 10, 4, LocSelectIfEither, SizeSelectIfGT},
				{"EitherLTWidth", 7, 2, LocSelectIfEither, SizeSelectIfLT},
				{"EitherLTHeight", 2, 10, LocSelectIfEither, SizeSelectIfLT},
				// both must fit the bounds
				{"BothLTE", 6, 9, LocSelectIfBoth, SizeSelectIfLTE},
				{"BothGTE", 3, 5, LocSelectIfBoth, SizeSelectIfGTE},
				{"BothLT", 7, 10, LocSelectIfBoth, SizeSelectIfLT},
				{"BothGT", 2, 4, LocSelectIfBoth, SizeSelectIfGT},
			}

			for _, testCase := range tests {
				t.Run(testCase.name, func(t *testing.T) {
					bitmaps := make([]*Bitmap, len(sizes))
					boxes := make([]*image.Rectangle, len(sizes))
					var rect image.Rectangle
					for i, sz := range sizes {
						bitmaps[i] = New(sz.width, sz.height)
						rect = image.Rect(0, 0, sz.width, sz.height)
						boxes[i] = &rect
					}

					bms := &Bitmaps{Values: bitmaps, Boxes: boxes}
					selected, err := bms.SelectBySize(testCase.width, testCase.height, testCase.location, testCase.comparison)
					require.NoError(t, err)

					assert.Equal(t, selected, bms)
					assert.Equal(t, selected.Values, bitmaps)
					assert.Equal(t, selected.Boxes, boxes)
				})
			}
		})

		t.Run("PartlyWithin", func(t *testing.T) {
			t.Run("BothLTE", func(t *testing.T) {
				sizes := []size{{7, 8}, {6, 9}, {5, 5}, {3, 7}}
				bitmaps := make([]*Bitmap, len(sizes))
				boxes := make([]*image.Rectangle, len(sizes))
				var rect image.Rectangle
				for i, sz := range sizes {
					bitmaps[i] = New(sz.width, sz.height)
					rect = image.Rect(0, 0, sz.width, sz.height)
					boxes[i] = &rect
				}

				bms := &Bitmaps{Values: bitmaps, Boxes: boxes}
				selected, err := bms.SelectBySize(6, 10, LocSelectIfBoth, SizeSelectIfLTE)
				require.NoError(t, err)

				assert.NotEqual(t, selected, bms)
				assert.Len(t, selected.Values, 3)
				assert.Len(t, selected.Boxes, 3)

				for i := 1; i < 4; i++ {
					assert.Equal(t, bitmaps[i], selected.Values[i-1])
					assert.Equal(t, boxes[i], selected.Boxes[i-1])
				}
			})
		})
	})
}

// TestBitmaps_SortByHeight tests the SortByHeight function for the bitmaps 'b'
func TestBitmaps_SortByHeight(t *testing.T) {
	bms := &Bitmaps{}
	bms.AddBitmap(New(10, 20))
	bms.AddBitmap(New(5, 18))
	bms.AddBitmap(New(2, 30))
	bms.AddBitmap(New(40, 20))

	bms.SortByHeight()
	assert.Equal(t, 18, bms.Values[0].Height)
	assert.Equal(t, 20, bms.Values[1].Height)
	assert.Equal(t, 20, bms.Values[2].Height)
	assert.Equal(t, 30, bms.Values[3].Height)
}

// TestBitmaps_SortByWidth tests the SortByHeight function for the bitmaps 'b'
func TestBitmaps_SortByWidth(t *testing.T) {
	bms := &Bitmaps{}
	bms.AddBitmap(New(10, 20))
	bms.AddBitmap(New(5, 18))
	bms.AddBitmap(New(2, 30))
	bms.AddBitmap(New(40, 20))

	bms.SortByWidth()
	assert.Equal(t, 2, bms.Values[0].Width)
	assert.Equal(t, 5, bms.Values[1].Width)
	assert.Equal(t, 10, bms.Values[2].Width)
	assert.Equal(t, 40, bms.Values[3].Width)
}

func TestBitmaps_GroupByHeight(t *testing.T) {
	bms := &Bitmaps{}
	bms.AddBitmap(New(10, 20))
	bms.AddBitmap(New(5, 18))
	bms.AddBitmap(New(2, 30))
	bms.AddBitmap(New(40, 20))
	bms.AddBitmap(New(2, 20))
	bms.AddBitmap(New(3, 20))
	bms.AddBitmap(New(3, 18))

	ba, err := bms.GroupByHeight()
	require.NoError(t, err)

	require.Len(t, ba.Values, 3)
	for i, bmst := range ba.Values {
		var height, length int
		switch i {
		case 0:
			height = 18
			length = 2
		case 1:
			height = 20
			length = 4
		case 2:
			height = 30
			length = 1
		}
		assert.Len(t, bmst.Values, length)
		assert.Equal(t, height, bmst.Values[0].Height)
	}
}

func TestBitmaps_GroupByWidth(t *testing.T) {
	bms := &Bitmaps{}
	bms.AddBitmap(New(10, 20))
	bms.AddBitmap(New(5, 18))
	bms.AddBitmap(New(2, 30))
	bms.AddBitmap(New(40, 20))
	bms.AddBitmap(New(2, 20))
	bms.AddBitmap(New(3, 20))
	bms.AddBitmap(New(3, 18))

	// 2 x 2, 2 x 3, 1 x 5, 1 x 10, 1 x 40
	// there should be 5 groups.
	ba, err := bms.GroupByWidth()
	require.NoError(t, err)

	require.Len(t, ba.Values, 5)
	for i, bmst := range ba.Values {
		var width, length int
		switch i {
		case 0:
			width = 2
			length = 2
		case 1:
			width = 3
			length = 2
		case 2:
			width = 5
			length = 1
		case 3:
			width = 10
			length = 1
		case 4:
			width = 40
			length = 1
		}
		assert.Len(t, bmst.Values, length)
		assert.Equal(t, width, bmst.Values[0].Width)
	}
}
