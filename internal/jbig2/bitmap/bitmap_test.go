/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"fmt"
	"math/rand"
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

	t.Run("SetEightBytes", func(t *testing.T) {
		t.Run("Partial", func(t *testing.T) {
			s := New(70, 2)
			err := s.setEightBytes(4, uint64(0xffffffffffffffff))
			require.NoError(t, err)

			assert.Equal(t, byte(0xff), s.Data[4])
			assert.Equal(t, byte(0xff), s.Data[5])
			assert.Equal(t, byte(0xff), s.Data[6])
			assert.Equal(t, byte(0xff), s.Data[7])
			// the eight - partial byte should contain only bits set up to the
			// possible width: 11111100 -> 72 - 70 = 2 padding bits
			assert.Equal(t, byte(0xfc), s.Data[8])
		})
		t.Run("NoPadding", func(t *testing.T) {
			s := New(64, 2)
			err := s.setEightBytes(4, uint64(0xffffffffffffffff))
			require.NoError(t, err)

			assert.Equal(t, byte(0xff), s.Data[4])
			assert.Equal(t, byte(0xff), s.Data[5])
			assert.Equal(t, byte(0xff), s.Data[6])
			assert.Equal(t, byte(0xff), s.Data[7])
			assert.Equal(t, byte(0x00), s.Data[8])
		})
		t.Run("Full", func(t *testing.T) {
			s := New(128, 2)
			err := s.setEightBytes(4, uint64(0xffffffffffffffff))
			require.NoError(t, err)

			assert.Equal(t, byte(0x00), s.Data[3])
			assert.Equal(t, byte(0xff), s.Data[4])
			assert.Equal(t, byte(0xff), s.Data[5])
			assert.Equal(t, byte(0xff), s.Data[6])
			assert.Equal(t, byte(0xff), s.Data[7])
			assert.Equal(t, byte(0xff), s.Data[8])
			assert.Equal(t, byte(0xff), s.Data[9])
			assert.Equal(t, byte(0xff), s.Data[10])
			assert.Equal(t, byte(0xff), s.Data[11])
			assert.Equal(t, byte(0x00), s.Data[12])
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
			// the width of 20 would have some padding
			bm := New(20, 2)

			bm.SetPixel(17, 0, 1)
			bm.SetPixel(19, 0, 1)

			bm.SetPixel(3, 1, 1)
			bm.SetPixel(4, 1, 1)

			bm.SetPixel(9, 1, 1)
			bm.SetPixel(12, 1, 1)
			bm.SetPixel(19, 1, 1)

			// row stride should be 3
			// padding at last byte of row is 4
			// 00000000	00000000 01010000
			// 00011000	01001000 00010000
			//
			// 0x00		0x00	 0xA0
			// 0x18		0x48	 0x10

			// 00000100

			assert.Equal(t, 6, len(bm.Data))
			assert.Equal(t, byte(0x00), bm.Data[0], "expected: %08b, is: %08b", byte(0x00), bm.Data[0])
			assert.Equal(t, byte(0x00), bm.Data[1], "expected: %08b, is: %08b", byte(0x00), bm.Data[1])
			assert.Equal(t, byte(0x50), bm.Data[2], "expected: %08b, is: %08b", byte(0x50), bm.Data[2])
			assert.Equal(t, byte(0x18), bm.Data[3], "expected: %08b, is: %08b", byte(0x18), bm.Data[3])
			assert.Equal(t, byte(0x48), bm.Data[4], "expected: %08b, is: %08b", byte(0x48), bm.Data[4])
			assert.Equal(t, byte(0x10), bm.Data[5], "expected: %08b, is: %08b", byte(0x10), bm.Data[5])

			unpadded, err := bm.GetUnpaddedData()
			require.NoError(t, err)

			// unpadded data should be:
			// 00000000 00000000 01010001
			// 10000100 10000001

			assert.Len(t, unpadded, 5)
			assert.Equal(t, byte(0x00), unpadded[0], "expected: %08b, is: %08b", byte(0x00), unpadded[0])
			assert.Equal(t, byte(0x00), unpadded[1], "expected: %08b, is: %08b", byte(0x00), unpadded[1])
			assert.Equal(t, byte(0x51), unpadded[2], "expected: %08b, is: %08b", byte(0x51), unpadded[2])
			assert.Equal(t, byte(0x84), unpadded[3], "expected: %08b, is: %08b", byte(0x84), unpadded[3])
			assert.Equal(t, byte(0x81), unpadded[4], "expected: %08b, is: %08b", byte(0x81), unpadded[4])
		})

		t.Run("NotEqualPadding", func(t *testing.T) {
			if testing.Verbose() {
				common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
			}
			t.Run("AllMarked", func(t *testing.T) {
				// bitmap width - 2196 % 8 = 4
				bm := New(2196, 3)
				for x := 0; x < bm.Width; x++ {
					for y := 0; y < bm.Height; y++ {
						bm.SetPixel(x, y, 1)
					}
				}

				unpadded, err := bm.GetUnpaddedData()
				require.NoError(t, err)

				for i := range unpadded {
					if i == len(unpadded)-1 {
						assert.Equal(t, byte(0xF0), unpadded[i])
						continue
					}
					assert.Equal(t, byte(0xFF), unpadded[i])
				}
			})

			t.Run("SomeMarked", func(t *testing.T) {
				// the width of 20 would have some padding
				bm := New(19, 2)

				bm.SetPixel(16, 0, 1)
				bm.SetPixel(18, 0, 1)

				bm.SetPixel(3, 1, 1)
				bm.SetPixel(4, 1, 1)

				bm.SetPixel(9, 1, 1)
				bm.SetPixel(12, 1, 1)
				bm.SetPixel(18, 1, 1)

				unpadded, err := bm.GetUnpaddedData()
				require.NoError(t, err)

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
				}
			})
		})
	})

	t.Run("Inverse", func(t *testing.T) {
		// having a bitmap of width: 20 and height: 2 with the data:
		//	  11110101	  10101100	  10110000	  11010011 	  10110001	  11100000 - total 25 '1' bits.
		//			F5			AC			B0			D3			B1			E0
		data := []byte{0xF5, 0xAC, 0xB0, 0xD3, 0xB1, 0xE0}

		t.Run("Chocolate", func(t *testing.T) {
			cdata := make([]byte, 6)
			copy(cdata, data)

			bm, err := NewWithData(20, 2, cdata)
			require.NoError(t, err)
			bm.Color = Chocolate

			bm.InverseData()
			assert.Equal(t, Vanilla, bm.Color)
			// The result should be:
			// 	00001010	01010011	01000000	00101100	01001110	00010000
			//		0x0A		0x53		0x40		0x2C		0x4E		0x10
			shouldBe := []byte{0x0A, 0x53, 0x40, 0x2C, 0x4E, 0x10}
			assert.Equal(t, shouldBe, bm.Data)
		})

		t.Run("Vanilla", func(t *testing.T) {
			cdata := make([]byte, 6)
			copy(cdata, data)

			bm, err := NewWithData(20, 2, cdata)
			require.NoError(t, err)
			bm.Color = Vanilla

			bm.InverseData()
			assert.Equal(t, Chocolate, bm.Color)
			// The result should be:
			// 	00001010	01010011	01000000	00101100	01001110	00010000
			//		0x0A		0x53		0x40		0x2C		0x4E		0x10
			shouldBe := []byte{0x0A, 0x53, 0x40, 0x2C, 0x4E, 0x10}
			assert.Equal(t, shouldBe, bm.Data)
		})

	})

	t.Run("Equivalent", func(t *testing.T) {
		t.Run("SmallSize", func(t *testing.T) {
			// the Equivalent function take into consideration the bitmaps
			// The root data to compare is the bitmap of width: 20 and height: 6 with the data:
			//	  11110101	  10101100	  10110000
			//			F5			AC			B0
			// 	  11010011 	  10110001	  11100000
			//			D3			B1			E0
			//	  01101001	  01001010    10000000
			//			69			4A			80
			//    11101111    10101111	  01000000
			//    		EF			AF			40
			//	  00001000	  01011101    11010000
			//			08			5D			D0
			//    11101001	  11111111	  01010000
			// 			E9			FF			50
			data := []byte{
				0xF5, 0xAC, 0xB0,
				0xD3, 0xB1, 0xE0,
				0x69, 0x4A, 0x80,
				0xEF, 0xAF, 0x40,
				0x08, 0x5D, 0xD0,
				0xE9, 0xFF, 0x50,
			}
			bm, err := NewWithData(20, 6, data)
			require.NoError(t, err)

			// the first data contains the same byte data as the root 'bm'.
			firstData := make([]byte, 18)
			copy(firstData, data)

			first, err := NewWithData(20, 6, firstData)
			require.NoError(t, err)

			// the result of the Equivalent should be true as the bitmaps are mostly equivalent.
			assert.True(t, bm.Equivalent(first))

			// as the size of the bitmap is relatively small changing even a bit would result in false.
			secondData := make([]byte, 18)
			copy(secondData, data)
			secondData[0] = 0xF4

			second, err := NewWithData(20, 6, secondData)
			require.NoError(t, err)

			assert.False(t, bm.Equivalent(second))
		})

		t.Run("BigSize", func(t *testing.T) {
			rd := rand.New(rand.NewSource(356))

			// Let's have a 1024x768 bitmap with random data bytes.
			data := make([]byte, (1024*768)>>3)
			n, err := rd.Read(data)
			require.NoError(t, err)

			// 1024*768 >> 3 = 98304
			assert.Equal(t, 98304, n)

			bm, err := NewWithData(1024, 768, data)
			require.NoError(t, err)

			// let's create a second bitmap with flipped one bit only.
			firstData := make([]byte, (1024*768)>>3)
			copy(firstData, data)

			if firstData[0]&0x01 == 0 {
				// set the first bit to '1'
				firstData[0] |= 0x01
			} else {
				// clear the first bit to '0'
				firstData[0] &^= 1
			}
			first, err := NewWithData(1024, 768, firstData)
			require.NoError(t, err)

			// Even though the data differs with a single bit, the bitmaps are treated as equivalent
			// due to it's size.
			assert.True(t, bm.Equivalent(first))
		})
	})

	t.Run("Copy", func(t *testing.T) {
		// Having some bitmap of size 40x50 with some random data.
		data := make([]byte, (40*50)>>3)
		n, err := rand.Read(data)
		require.NoError(t, err)

		// 40x50 >> 3 = 250
		assert.Equal(t, 250, n)

		bm, err := NewWithData(40, 50, data)
		require.NoError(t, err)

		copied := bm.Copy()
		// 'assert.Equal' checks if all fields are equal for the structures.
		assert.Equal(t, bm, copied)
		// but while comparing the pointers they're not pointing to the same structure.
		assert.False(t, bm == copied)
	})

	t.Run("CountPixels", func(t *testing.T) {
		// having a bitmap of width: 20 and height: 2 with the data:
		//	  11110101	  10101100	  10110000	  11010011 	  10110001	  11100000 - total 25 '1' bits.
		//			F5			AC			B0			D3			B1			E0
		data := []byte{0xF5, 0xAC, 0xB0, 0xD3, 0xB1, 0xE0}
		bm, err := NewWithData(20, 2, data)
		require.NoError(t, err)

		oneBitsNumber := 25

		assert.Equal(t, oneBitsNumber, bm.CountPixels())
	})

	t.Run("AddBorder", func(t *testing.T) {
		t.Run("Border>0", func(t *testing.T) {
			// The root data to add the border is the bitmap of width: 20 and height: 6 with the data:
			//	  11110101	  10101100	  10110000
			//			F5			AC			B0
			// 	  11010011 	  10110001	  11100000
			//			D3			B1			E0
			//	  01101001	  01001010    10000000
			//			69			4A			80
			//    11101111    10101111	  01000000
			//    		EF			AF			40
			//	  00001000	  01011101    11010000
			//			08			5D			D0
			//    11101001	  11111111	  01010000
			// 			E9			FF			50
			data := []byte{
				0xF5, 0xAC, 0xB0,
				0xD3, 0xB1, 0xE0,
				0x69, 0x4A, 0x80,
				0xEF, 0xAF, 0x40,
				0x08, 0x5D, 0xD0,
				0xE9, 0xFF, 0x50,
			}
			bm, err := NewWithData(20, 6, data)
			require.NoError(t, err)

			// add the border of size 1 to each side of the bitmap
			bmWithBorder, err := bm.AddBorder(1, 1)
			require.NoError(t, err)

			// the result should be as follows
			//	  11111111	11111111  11111100	- 0xFF, 0xFF, 0xFC
			//	  11111010  11010110  01011100	- 0xFA, 0xD6, 0x5C
			// 	  11101001  11011000  11110100	- 0xE9, 0xD8, 0xF4
			//	  10110100  10100101  01000100	- 0xB4, 0xA5, 0x44
			//    11110111  11010111  10100100	- 0xF7, 0xD7, 0xA4
			//	  10000100  00101110  11101100	- 0x84, 0x2E, 0xEC
			//    11110100  11111111  10101100	- 0xF4, 0xFF, 0xAC
			//	  11111111  11111111  11111100	- 0xFF, 0xFF, 0xFC
			shouldBe := []byte{0xFF, 0xFF, 0xFC, 0xFA, 0xD6, 0x5C, 0xE9, 0xD8, 0xF4, 0xB4, 0xA5, 0x44, 0xF7, 0xD7, 0xA4, 0x84, 0x2E, 0xEC, 0xF4, 0xFF, 0xAC, 0xFF, 0xFF, 0xFC}
			assert.Equal(t, bmWithBorder.Data, shouldBe)
			assert.Equal(t, bmWithBorder.Width, bm.Width+2)
			assert.Equal(t, bmWithBorder.Height, bm.Height+2)
		})
		t.Run("Border=0", func(t *testing.T) {
			rd := rand.New(rand.NewSource(12345))

			data := make([]byte, 6)
			n, err := rd.Read(data)
			require.NoError(t, err)

			assert.Equal(t, 6, n)

			bm, err := NewWithData(18, 2, data)
			require.NoError(t, err)

			copied, err := bm.AddBorder(0, 1)
			require.NoError(t, err)

			assert.Equal(t, bm, copied)
			assert.False(t, bm == copied)
		})

		t.Run("General", func(t *testing.T) {
			// The root data to add the border is the bitmap of width: 20 and height: 6 with the data:
			//	  11110101	  10101100	  10110000
			//			F5			AC			B0
			// 	  11010011 	  10110001	  11100000
			//			D3			B1			E0
			//	  01101001	  01001010    10000000
			//			69			4A			80
			//    11101111    10101111	  01000000
			//    		EF			AF			40
			//	  00001000	  01011101    11010000
			//			08			5D			D0
			//    11101001	  11111111	  01010000
			// 			E9			FF			50
			data := []byte{
				0xF5, 0xAC, 0xB0,
				0xD3, 0xB1, 0xE0,
				0x69, 0x4A, 0x80,
				0xEF, 0xAF, 0x40,
				0x08, 0x5D, 0xD0,
				0xE9, 0xFF, 0x50,
			}
			bm, err := NewWithData(20, 6, data)
			require.NoError(t, err)

			t.Run("Negative", func(t *testing.T) {
				nbm, err := bm.AddBorderGeneral(-1, 0, 0, 0, 0)
				require.Error(t, err)
				assert.Nil(t, nbm)
			})

			t.Run("Left", func(t *testing.T) {
				lbm, err := bm.AddBorderGeneral(8, 0, 0, 0, 1)
				require.NoError(t, err)

				// The data with 8 bits left border should be:
				//	11111111  11110101	  10101100	  10110000
				//		  FF		F5			AC			B0
				// 	11111111  11010011 	  10110001	  11100000
				//		  FF		D3			B1			E0
				//	11111111  01101001	  01001010    10000000
				//		  FF		69			4A			80
				//  11111111  11101111    10101111	  01000000
				//		  FF   		EF			AF			40
				//	11111111  00001000	  01011101    11010000
				//		  FF		08			5D			D0
				//  11111111  11101001	  11111111	  01010000
				//		  FF 		E9			FF			50
				data := []byte{
					0xFF, 0xF5, 0xAC, 0xB0,
					0xFF, 0xD3, 0xB1, 0xE0,
					0xFF, 0x69, 0x4A, 0x80,
					0xFF, 0xEF, 0xAF, 0x40,
					0xFF, 0x08, 0x5D, 0xD0,
					0xFF, 0xE9, 0xFF, 0x50,
				}

				assert.Equal(t, data, lbm.Data)
			})

			t.Run("Right", func(t *testing.T) {
				rbm, err := bm.AddBorderGeneral(0, 4, 0, 0, 1)
				require.NoError(t, err)

				// The data with '4' bits right border should be:
				//	  11110101	  10101100	  10111111
				//			F5			AC			BF
				// 	  11010011 	  10110001	  11101111
				//			D3			B1			EF
				//	  01101001	  01001010    10001111
				//			69			4A			8F
				//    11101111    10101111	  01001111
				//    		EF			AF			4F
				//	  00001000	  01011101    11011111
				//			08			5D			DF
				//    11101001	  11111111	  01011111
				// 			E9			FF			5F

				data := []byte{
					0xF5, 0xAC, 0xBF,
					0xD3, 0xB1, 0xEF,
					0x69, 0x4A, 0x8F,
					0xEF, 0xAF, 0x4F,
					0x08, 0x5D, 0xDF,
					0xE9, 0xFF, 0x5F,
				}
				assert.Equal(t, rbm.Data, data)
			})

			t.Run("Top", func(t *testing.T) {
				tbm, err := bm.AddBorderGeneral(0, 0, 1, 0, 1)
				require.NoError(t, err)

				// The data with '1 bit' size top border should be:
				//	  11111111	  11111111	  11110000
				//			FF			FF			F0
				//	  11110101	  10101100	  10110000
				//			F5			AC			B0
				// 	  11010011 	  10110001	  11100000
				//			D3			B1			E0
				//	  01101001	  01001010    10000000
				//			69			4A			80
				//    11101111    10101111	  01000000
				//    		EF			AF			40
				//	  00001000	  01011101    11010000
				//			08			5D			D0
				//    11101001	  11111111	  01010000
				// 			E9			FF			50
				data := []byte{
					0xFF, 0xFF, 0xF0,
					0xF5, 0xAC, 0xB0,
					0xD3, 0xB1, 0xE0,
					0x69, 0x4A, 0x80,
					0xEF, 0xAF, 0x40,
					0x08, 0x5D, 0xD0,
					0xE9, 0xFF, 0x50,
				}
				assert.Equal(t, tbm.Data, data)
			})

			t.Run("Bottom", func(t *testing.T) {
				bbm, err := bm.AddBorderGeneral(0, 0, 0, 1, 1)
				require.NoError(t, err)

				// The data with '1bit' size bottom data should be:
				//	  11110101	  10101100	  10110000
				//			F5			AC			B0
				// 	  11010011 	  10110001	  11100000
				//			D3			B1			E0
				//	  01101001	  01001010    10000000
				//			69			4A			80
				//    11101111    10101111	  01000000
				//    		EF			AF			40
				//	  00001000	  01011101    11010000
				//			08			5D			D0
				//    11101001	  11111111	  01010000
				// 			E9			FF			50
				//	  11111111	  11111111	  11110000
				//			FF			FF			F0
				data := []byte{
					0xF5, 0xAC, 0xB0,
					0xD3, 0xB1, 0xE0,
					0x69, 0x4A, 0x80,
					0xEF, 0xAF, 0x40,
					0x08, 0x5D, 0xD0,
					0xE9, 0xFF, 0x50,
					0xFF, 0xFF, 0xF0,
				}
				assert.Equal(t, bbm.Data, data)
			})
		})
	})

	t.Run("RemoveBorder", func(t *testing.T) {
		t.Run("Border>0", func(t *testing.T) {
			// having a test data of width: 22, height: 8 with a border of '1pix' size
			// with the data:
			//	  11111111	11111111  11111100	- 0xFF, 0xFF, 0xFC
			//	  11111010  11010110  01011100	- 0xFA, 0xD6, 0x5C
			// 	  11101001  11011000  11110100	- 0xE9, 0xD8, 0xF4
			//	  10110100  10100101  01000100	- 0xB4, 0xA5, 0x44
			//    11110111  11010111  10100100	- 0xF7, 0xD7, 0xA4
			//	  10000100  00101110  11101100	- 0x84, 0x2E, 0xEC
			//    11110100  11111111  10101100	- 0xF4, 0xFF, 0xAC
			//	  11111111  11111111  11111100	- 0xFF, 0xFF, 0xFC
			data := []byte{0xFF, 0xFF, 0xFC, 0xFA, 0xD6, 0x5C, 0xE9, 0xD8, 0xF4, 0xB4, 0xA5, 0x44, 0xF7, 0xD7, 0xA4, 0x84, 0x2E, 0xEC, 0xF4, 0xFF, 0xAC, 0xFF, 0xFF, 0xFC}
			bm, err := NewWithData(22, 8, data)
			require.NoError(t, err)

			bmNoBorder, err := bm.RemoveBorder(1)
			require.NoError(t, err)

			// the data without border should look like
			//	  11110101	  10101100	  10110000
			//			F5			AC			B0
			// 	  11010011 	  10110001	  11100000
			//			D3			B1			E0
			//	  01101001	  01001010    10000000
			//			69			4A			80
			//    11101111    10101111	  01000000
			//    		EF			AF			40
			//	  00001000	  01011101    11010000
			//			08			5D			D0
			//    11101001	  11111111	  01010000
			// 			E9			FF			50
			shouldBe := []byte{0xF5, 0xAC, 0xB0, 0xD3, 0xB1, 0xE0, 0x69, 0x4A, 0x80, 0xEF, 0xAF, 0x40, 0x08, 0x5D, 0xD0, 0xE9, 0xFF, 0x50}

			assert.Equal(t, bm.Width-2, bmNoBorder.Width)
			assert.Equal(t, bm.Height-2, bmNoBorder.Height)
			assert.Equal(t, shouldBe, bmNoBorder.Data)
		})

		t.Run("Border=0", func(t *testing.T) {
			rd := rand.New(rand.NewSource(12345))

			data := make([]byte, 6)
			n, err := rd.Read(data)
			require.NoError(t, err)

			assert.Equal(t, 6, n)

			bm, err := NewWithData(18, 2, data)
			require.NoError(t, err)

			copied, err := bm.RemoveBorder(0)
			require.NoError(t, err)

			assert.Equal(t, bm, copied)
			assert.False(t, bm == copied)
		})

		t.Run("General", func(t *testing.T) {
			// The root data to add the border is the bitmap of width: 20 and height: 6 with the data:
			//	  11110101	  10101100	  10110000
			//			F5			AC			B0
			// 	  11010011 	  10110001	  11100000
			//			D3			B1			E0
			//	  01101001	  01001010    10000000
			//			69			4A			80
			//    11101111    10101111	  01000000
			//    		EF			AF			40
			//	  00001000	  01011101    11010000
			//			08			5D			D0
			//    11101001	  11111111	  01010000
			// 			E9			FF			50
			generalData := []byte{
				0xF5, 0xAC, 0xB0,
				0xD3, 0xB1, 0xE0,
				0x69, 0x4A, 0x80,
				0xEF, 0xAF, 0x40,
				0x08, 0x5D, 0xD0,
				0xE9, 0xFF, 0x50,
			}

			t.Run("Negative", func(t *testing.T) {
				t.Run("BorderSize", func(t *testing.T) {
					bm, err := NewWithData(20, 2, generalData)
					require.NoError(t, err)

					nbm, err := bm.RemoveBorderGeneral(-1, 0, 0, 0)
					require.Error(t, err)
					assert.Nil(t, nbm)
				})

				t.Run("Resultant", func(t *testing.T) {
					t.Run("Width", func(t *testing.T) {
						bm, err := NewWithData(20, 2, generalData)
						require.NoError(t, err)
						// remove border of left + right >= 20
						nbm, err := bm.RemoveBorderGeneral(14, 6, 0, 0)
						require.Error(t, err)
						assert.Nil(t, nbm)
					})
					t.Run("Height", func(t *testing.T) {
						bm, err := NewWithData(20, 2, generalData)
						require.NoError(t, err)

						// remove border of top + bottom > 2
						nbm, err := bm.RemoveBorderGeneral(0, 0, 1, 2)
						require.Error(t, err)
						assert.Nil(t, nbm)
					})
				})
			})

			t.Run("Left", func(t *testing.T) {
				// The data with 8 bits left border should be:
				//	11111111  11110101	  10101100	  10110000
				//		  FF		F5			AC			B0
				// 	11111111  11010011 	  10110001	  11100000
				//		  FF		D3			B1			E0
				//	11111111  01101001	  01001010    10000000
				//		  FF		69			4A			80
				//  11111111  11101111    10101111	  01000000
				//		  FF   		EF			AF			40
				//	11111111  00001000	  01011101    11010000
				//		  FF		08			5D			D0
				//  11111111  11101001	  11111111	  01010000
				//		  FF 		E9			FF			50
				data := []byte{
					0xFF, 0xF5, 0xAC, 0xB0,
					0xFF, 0xD3, 0xB1, 0xE0,
					0xFF, 0x69, 0x4A, 0x80,
					0xFF, 0xEF, 0xAF, 0x40,
					0xFF, 0x08, 0x5D, 0xD0,
					0xFF, 0xE9, 0xFF, 0x50,
				}
				bm, err := NewWithData(28, 6, data)
				require.NoError(t, err)

				lbm, err := bm.RemoveBorderGeneral(8, 0, 0, 0)
				require.NoError(t, err)

				assert.Equal(t, generalData, lbm.Data)
			})

			t.Run("Right", func(t *testing.T) {
				// The data with '4' bits right border should be:
				//	  11110101	  10101100	  10111111
				//			F5			AC			BF
				// 	  11010011 	  10110001	  11101111
				//			D3			B1			EF
				//	  01101001	  01001010    10001111
				//			69			4A			8F
				//    11101111    10101111	  01001111
				//    		EF			AF			4F
				//	  00001000	  01011101    11011111
				//			08			5D			DF
				//    11101001	  11111111	  01011111
				// 			E9			FF			5F
				data := []byte{
					0xF5, 0xAC, 0xBF,
					0xD3, 0xB1, 0xEF,
					0x69, 0x4A, 0x8F,
					0xEF, 0xAF, 0x4F,
					0x08, 0x5D, 0xDF,
					0xE9, 0xFF, 0x5F,
				}

				bm, err := NewWithData(24, 6, data)
				require.NoError(t, err)

				rbm, err := bm.RemoveBorderGeneral(0, 4, 0, 0)
				require.NoError(t, err)

				assert.Equal(t, generalData, rbm.Data)
			})

			t.Run("Top", func(t *testing.T) {
				// The data with '1 bit' size top border should be:
				//	  11111111	  11111111	  11110000
				//			FF			FF			F0
				//	  11110101	  10101100	  10110000
				//			F5			AC			B0
				// 	  11010011 	  10110001	  11100000
				//			D3			B1			E0
				//	  01101001	  01001010    10000000
				//			69			4A			80
				//    11101111    10101111	  01000000
				//    		EF			AF			40
				//	  00001000	  01011101    11010000
				//			08			5D			D0
				//    11101001	  11111111	  01010000
				// 			E9			FF			50
				data := []byte{
					0xFF, 0xFF, 0xF0,
					0xF5, 0xAC, 0xB0,
					0xD3, 0xB1, 0xE0,
					0x69, 0x4A, 0x80,
					0xEF, 0xAF, 0x40,
					0x08, 0x5D, 0xD0,
					0xE9, 0xFF, 0x50,
				}

				bm, err := NewWithData(20, 7, data)
				require.NoError(t, err)

				tbm, err := bm.RemoveBorderGeneral(0, 0, 1, 0)
				require.NoError(t, err)

				assert.Equal(t, generalData, tbm.Data)
			})

			t.Run("Bottom", func(t *testing.T) {
				// The data with '1bit' size bottom data should be:
				//	  11110101	  10101100	  10110000
				//			F5			AC			B0
				// 	  11010011 	  10110001	  11100000
				//			D3			B1			E0
				//	  01101001	  01001010    10000000
				//			69			4A			80
				//    11101111    10101111	  01000000
				//    		EF			AF			40
				//	  00001000	  01011101    11010000
				//			08			5D			D0
				//    11101001	  11111111	  01010000
				// 			E9			FF			50
				//	  11111111	  11111111	  11110000
				//			FF			FF			F0
				data := []byte{
					0xF5, 0xAC, 0xB0,
					0xD3, 0xB1, 0xE0,
					0x69, 0x4A, 0x80,
					0xEF, 0xAF, 0x40,
					0x08, 0x5D, 0xD0,
					0xE9, 0xFF, 0x50,
					0xFF, 0xFF, 0xF0,
				}
				bm, err := NewWithData(20, 7, data)
				require.NoError(t, err)

				bbm, err := bm.RemoveBorderGeneral(0, 0, 0, 1)
				require.NoError(t, err)

				assert.Equal(t, generalData, bbm.Data)
			})
		})
	})

	t.Run("NextOnPixel", func(t *testing.T) {
		// Having a bitmap with given data:
		//
		// 00001000 00100000
		// 00000000 00000010
		data := []byte{0x08, 0x20, 0x00, 0x02}

		bm, err := NewWithData(16, 2, data)
		require.NoError(t, err)

		// First should be at Pt(4,0)
		pt, ok, err := bm.nextOnPixel(0, 0)
		require.NoError(t, err)

		assert.True(t, ok)
		assert.Equal(t, pt.X, 4)
		assert.Equal(t, pt.Y, 0)

		// The second should be at Pt(10, 0)
		pt, ok, err = bm.nextOnPixel(5, 0)
		require.NoError(t, err)

		assert.True(t, ok)
		assert.Equal(t, pt.X, 10)
		assert.Equal(t, pt.Y, 0)

		// The third should be on another line at Pt(14,1)
		pt, ok, err = bm.nextOnPixel(11, 0)
		require.NoError(t, err)

		assert.True(t, ok)
		assert.Equal(t, pt.X, 14)
		assert.Equal(t, pt.Y, 1)

		// There should be no more 'ON' pixels.
		_, ok, err = bm.nextOnPixel(15, 1)
		require.NoError(t, err)

		assert.False(t, ok)

		// providing 'x' that is out of possible index range returns error.
		_, _, err = bm.nextOnPixel(50, 0)
		assert.Error(t, err)

		// providing 'y' that is out of possible index range returns error.
		_, _, err = bm.nextOnPixel(3, 40)
		assert.Error(t, err)
	})

	t.Run("Zero", func(t *testing.T) {
		t.Run("Full", func(t *testing.T) {
			// having a bitmap of size 20,2 with filled bytes
			data := []byte{0xFF, 0xFF, 0xF0, 0xFF, 0xFF, 0xF0}
			bm, err := NewWithData(20, 2, data)
			require.NoError(t, err)

			// the Zero function would return false
			assert.False(t, bm.Zero())
		})

		t.Run("Empty", func(t *testing.T) {
			bm := New(20, 2)
			// now the Zero function should return true.
			assert.True(t, bm.Zero())
		})

		t.Run("PartlySet", func(t *testing.T) {
			data := []byte{0x00, 0x00, 0xF0, 0x00, 0x00, 0x00}
			bm, err := NewWithData(20, 2, data)
			require.NoError(t, err)

			assert.False(t, bm.Zero())
		})

	})

	t.Run("ConnComponents", func(t *testing.T) {

	})

	t.Run("ThresholdPixelSum", func(t *testing.T) {
		// Having a bitmap 100x100 with randomly distributed 10 pixel per row
		bm := New(100, 100)
		mp := [100][100]bool{}
		var x, count int
		for y := 0; y < bm.Height; y++ {
			for i := 0; i < 10; i++ {
				for {
					x = rand.Intn(bm.Width)
					if !mp[y][x] {
						break
					}
				}
				mp[y][x] = true
				count++
				require.NoError(t, bm.SetPixel(x, y, 1))
			}
		}
		require.Equal(t, 1000, count)
		tab8 := makePixelSumTab8()
		t.Run("Above", func(t *testing.T) {
			// for count > threshold the function returns true.
			above, err := bm.ThresholdPixelSum(500, tab8)
			require.NoError(t, err)

			assert.True(t, above)
		})

		t.Run("Equal", func(t *testing.T) {
			// In count > threshold  the inequality is false for count = threshold
			above, err := bm.ThresholdPixelSum(1000, tab8)
			require.NoError(t, err)

			// count > threshold = false
			assert.False(t, above)
		})

		t.Run("NotAbove", func(t *testing.T) {
			// if threshold > count then the inequality 'count > threshold' is false.
			above, err := bm.ThresholdPixelSum(1001, tab8)
			require.NoError(t, err)

			assert.False(t, above)
		})
	})
}

// TestSubtract tests the subtract function
func TestSubtract(t *testing.T) {
	// Having a 8x6 src1 bitmap.
	//
	// 11111000
	// 11111100
	// 11111000
	// 11110000
	// 11100000
	// 11000000
	d1 := []byte{0xF8, 0xFC, 0xF8, 0xF0, 0xE0, 0xC0}
	src1, err := NewWithData(8, 6, d1)
	require.NoError(t, err)

	// and a 8x6 src2 bitmap
	//
	// 00011100
	// 00111100
	// 00111000
	// 00110000
	// 00000000
	// 00000000
	d2 := []byte{0x1C, 0x3C, 0x38, 0x30, 0x00, 0x00}
	src2, err := NewWithData(8, 6, d2)
	require.NoError(t, err)

	t.Run("NilDest", func(t *testing.T) {
		d, err := subtract(nil, src1, src2)
		require.NoError(t, err)

		// the result should be:
		// 11100000
		// 11000000
		// 11000000
		// 11000000
		// 11100000
		// 11000000
		expected := []byte{0xE0, 0xC0, 0xC0, 0xC0, 0xE0, 0xC0}
		assert.Equal(t, expected, d.Data)
	})

	t.Run("DestEqualSrc1", func(t *testing.T) {
		tm := src1.Copy()
		_, err := subtract(tm, tm, src2)
		require.NoError(t, err)

		// the result should be:
		// 11100000
		// 11000000
		// 11000000
		// 11000000
		// 11100000
		// 11000000
		expected := []byte{0xE0, 0xC0, 0xC0, 0xC0, 0xE0, 0xC0}
		assert.Equal(t, expected, tm.Data)
	})

	t.Run("DestEqualSrc2", func(t *testing.T) {
		tm := src2.Copy()
		_, err := subtract(tm, src1, tm)
		require.NoError(t, err)

		// The result should be:
		// 00000100
		// 00000000
		// 00000000
		// 00000000
		// 00000000
		// 00000000
		expected := []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00}
		assert.Equal(t, expected, tm.Data)
	})

	t.Run("SomeDest", func(t *testing.T) {
		tm := New(10, 6)
		d, err := subtract(tm, src1, src2)
		require.NoError(t, err)

		// the result should be:
		// 11100000
		// 11000000
		// 11000000
		// 11000000
		// 11100000
		// 11000000
		expected := []byte{0xE0, 0xC0, 0xC0, 0xC0, 0xE0, 0xC0}
		assert.Equal(t, expected, d.Data)
	})
}
