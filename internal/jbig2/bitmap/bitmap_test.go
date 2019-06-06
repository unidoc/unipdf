/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
)

// TestBitmap tests the bitmap methods and constructors.
func TestBitmap(t *testing.T) {

	t.Run("New", func(t *testing.T) {
		// tests the creator of the bitmap
		t.Run("SingleBytePerRow", func(t *testing.T) {
			bm := New(5, 5)
			assert.Equal(t, 5, bm.Height)
			assert.Equal(t, 5, bm.Width)
			assert.Equal(t, 1, bm.RowStride)
			assert.Equal(t, 5, len(bm.Data))
		})

		t.Run("MultipleBytesPerRow", func(t *testing.T) {
			bm := New(25, 25)
			assert.Equal(t, 25, bm.Height)
			assert.Equal(t, 25, bm.Width)

			// 3 * 8 < 25 => RowStride = 4
			assert.Equal(t, 4, bm.RowStride)
			// 4 * 25
			assert.Equal(t, 100, len(bm.Data))
		})
	})

	t.Run("GetPixel", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			t.Run("Small", func(t *testing.T) {
				bm := New(5, 5)

				bm.Data[0] = 0x80
				assert.True(t, bm.GetPixel(0, 0), bm.String())
			})

			t.Run("Big", func(t *testing.T) {
				bm := New(25, 25)

				bm.Data[len(bm.Data)-1] = 0xF0
				assert.True(t, bm.GetPixel(24, 24), bm.String())
			})
		})

		t.Run("Invalid", func(t *testing.T) {
			bm := New(5, 5)

			assert.False(t, bm.GetPixel(100, 100))
		})
	})

	t.Run("SetPixel", func(t *testing.T) {

		t.Run("Valid", func(t *testing.T) {
			bm := New(5, 5)

			require.NoError(t, bm.SetPixel(0, 0, 1))
			assert.Equal(t, uint8(0x80), bm.Data[0])
		})

		t.Run("Invalid", func(t *testing.T) {
			bm := New(5, 5)

			require.Error(t, bm.SetPixel(100, 100, 1))
		})
	})

	t.Run("SetDefaultPixel", func(t *testing.T) {
		bm := New(5, 5)
		bm.SetDefaultPixel()

		for _, b := range bm.Data {
			assert.Equal(t, uint8(0xff), b)
		}
	})

	t.Run("GetByte", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			bm := New(5, 5)
			bm.Data[0] = 0xff

			b, err := bm.GetByte(0)
			require.NoError(t, err)

			assert.Equal(t, byte(0xff), b)
		})

		t.Run("Invalid", func(t *testing.T) {
			bm := New(5, 5)

			_, err := bm.GetByte(5)
			require.Error(t, err)

		})

	})

	t.Run("SetByte", func(t *testing.T) {
		t.Run("OutOfRange", func(t *testing.T) {
			bm := New(5, 5)
			require.Error(t, bm.SetByte(5, 0xff))
		})

		t.Run("Valid", func(t *testing.T) {
			bm := New(5, 5)
			require.NoError(t, bm.SetByte(0, 0xff))
			assert.Equal(t, uint8(0xff), bm.Data[0])
		})
	})

	t.Run("Equals", func(t *testing.T) {
		src := New(5, 5)
		src.Data[0] = 0xff
		src.Data[1] = 0xff

		t.Run("SameSizeDifferentData", func(t *testing.T) {
			tc1 := New(5, 5)
			tc1.Data[0] = 0xff
			tc1.Data[1] = 0xf0
			assert.False(t, src.Equals(tc1))
		})

		t.Run("SameDataDifferentSize", func(t *testing.T) {
			tc2 := New(10, 10)
			tc2.Data[0] = 0xff
			tc2.Data[1] = 0xff
			assert.False(t, src.Equals(tc2))
		})

		t.Run("SameSizeSameData", func(t *testing.T) {
			tc3 := New(5, 5)
			tc3.Data[0] = 0xff
			tc3.Data[1] = 0xff
			assert.True(t, src.Equals(tc3))
		})

	})

	t.Run("GetUnpadded", func(t *testing.T) {

		t.Run("EqualPadding", func(t *testing.T) {
			if testing.Verbose() {
				common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
			}
			// the widht of 20 would have some padding
			bm := New(20, 2)

			bm.SetPixel(17, 0, 1)
			bm.SetPixel(19, 0, 1)

			bm.SetPixel(3, 1, 1)
			bm.SetPixel(4, 1, 1)

			bm.SetPixel(9, 1, 1)
			bm.SetPixel(12, 1, 1)
			bm.SetPixel(19, 1, 1)

			unpadded := bm.GetUnpaddedData()

			assert.Len(t, unpadded, 5)

			for i, bt := range unpadded {
				switch i {
				case 0, 1:
					// 00000000
					assert.Equal(t, byte(0x00), bt)
				case 2:
					// 01010001
					assert.Equal(t, byte(0x51), bt)
				case 3:
					// 10000100
					assert.Equal(t, byte(0x84), bt)
				case 4:
					// 10000001
					assert.Equal(t, byte(0x81), bt, fmt.Sprintf("Should be: %08b is: %08b", 0x81, bt))
				}
				t.Logf("%d, %08b", i, bt)
			}

			t.Logf("Bitmap: %s", bm)
		})

		t.Run("NotEqualPadding", func(t *testing.T) {
			if testing.Verbose() {
				common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
			}
			// the widht of 20 would have some padding
			bm := New(19, 2)

			bm.SetPixel(16, 0, 1)
			bm.SetPixel(18, 0, 1)

			bm.SetPixel(3, 1, 1)
			bm.SetPixel(4, 1, 1)

			bm.SetPixel(9, 1, 1)
			bm.SetPixel(12, 1, 1)
			bm.SetPixel(18, 1, 1)

			unpadded := bm.GetUnpaddedData()

			for i, bt := range unpadded {
				switch i {
				case 0, 1:
					assert.Equal(t, byte(0x00), bt)
				case 2:
					assert.Equal(t, byte(0xa3), bt, fmt.Sprintf("Should be: %08b is: %08b", 0xa3, bt))
				case 3:
					assert.Equal(t, byte(0x09), bt, fmt.Sprintf("Should be: %08b is: %08b", 0x09, bt))
				case 4:
					assert.Equal(t, byte(0x04), bt, fmt.Sprintf("Should be: %08b is: %08b", 0x04, bt))

				}
				t.Logf("%d, %08b", i, bt)
			}
			t.Logf("Bitmap: %s", bm)
		})

	})
}
