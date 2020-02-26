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
)

// BenchmarkGetComponents benchmarks the get components methods.
func BenchmarkGetComponents(b *testing.B) {
	// Having a bitmap 50x22 with few different components like letters.
	//
	//            1        2        3        4        5        6         7
	//
	//                111111 11112222 22222233 33333333 44444444 44
	//     01234567 89012345 67890123 45678901 23456789 01234567 89
	//
	//  0  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  1  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  2  00111110 01111000 00100111 11000010 00100111 10010001 00000000
	//  3  00100010 01001000 00100001 00000011 00100100 10010001 00000000
	//  4  00100010 01001000 00100001 00000010 10100100 10010101 00000000
	//  5  00100010 01001000 00100001 00000010 01100100 10011011 00000000
	//  6  00111100 01111000 00100001 00000010 00100111 10010001 00000000
	//  7  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  8  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  9  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 10  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 11  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 12  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 13  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 14  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 15  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 16  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 17  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 18  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 19  00000000 00000000 00010101 01010000 00000000 00000000 00000000
	// 20  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 21  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//
	//
	data := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x3E, 0x78, 0x27, 0xC2, 0x27, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x03, 0x24, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x02, 0xA4, 0x95, 0x00,
		0x22, 0x48, 0x21, 0x02, 0x64, 0x9B, 0x00,
		0x3C, 0x78, 0x21, 0x02, 0x27, 0x91, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x15, 0x50, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	bm, err := NewWithData(50, 22, data)
	require.NoError(b, err)

	type bmSizes struct {
		name string
		bm   *Bitmap
		size int
	}

	bm2, err := expandBinaryPower2(bm, 2)
	require.NoError(b, err)

	bm4, err := expandBinaryPower2(bm, 4)
	require.NoError(b, err)

	bm8, err := expandBinaryPower2(bm, 8)
	require.NoError(b, err)

	sizes := []*bmSizes{
		{"One", bm, 1},
		{"Two", bm2, 2},
		{"Four", bm4, 4},
		{"Eight", bm8, 8},
	}

	b.Run("ComponentConn", func(b *testing.B) {
		w, h := 12, 8
		for _, size := range sizes {
			b.Run(size.name, func(b *testing.B) {
				var bmB *Bitmap
				maxW, maxH := w*size.size, h*size.size
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					bmB, err = copyBitmap(nil, size.bm)
					require.NoError(b, err)
					b.StartTimer()

					_, _, err = bmB.GetComponents(ComponentConn, maxW, maxH)
					assert.NoError(b, err)
				}
			})
		}
	})

	b.Run("ComponentCharacters", func(b *testing.B) {
		w, h := 12, 8
		for _, size := range sizes {
			b.Run(size.name, func(b *testing.B) {
				var bmB *Bitmap
				maxW, maxH := w*size.size, h*size.size
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					bmB, err = copyBitmap(nil, size.bm)
					require.NoError(b, err)
					b.StartTimer()

					_, _, err = bmB.GetComponents(ComponentCharacters, maxW, maxH)
					assert.NoError(b, err)
				}
			})
		}
	})

	b.Run("ComponentWords", func(b *testing.B) {
		w, h := 20, 8
		for _, size := range sizes {
			b.Run(size.name, func(b *testing.B) {
				var bmB *Bitmap
				maxW, maxH := w*size.size, h*size.size
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					bmB, err = copyBitmap(nil, size.bm)
					require.NoError(b, err)
					b.StartTimer()

					_, _, err = bmB.GetComponents(ComponentWords, maxW, maxH)
					assert.NoError(b, err)
				}
			})
		}
	})
}

// TestConnComponents is the function that tests the connectivity components methods.
func TestConnComponents(t *testing.T) {
	// Having a bitmap 50x18 with few different components like letters
	//
	//            1        2        3        4        5        6         7
	//
	//                111111 11112222 22222233 33333333 44444444 44
	//     01234567 89012345 67890123 45678901 23456789 01234567 89
	//
	//  0  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  1  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  2  00111110 01111000 00100111 11000010 00100111 10010001 00000000
	//  3  00100010 01001000 00100001 00000011 00100100 10010001 00000000
	//  4  00100010 01001000 00100001 00000010 10100100 10010101 00000000
	//  5  00100010 01001000 00100001 00000010 01100100 10011011 00000000
	//  6  00111100 01111000 00100001 00000010 00100111 10010001 00000000
	//  7  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  8  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  9  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 10  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 11  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 12  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 13  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 14  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 15  00000000 00000000 00010101 01010000 00000000 00000000 00000000
	// 16  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 17  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//
	//
	data := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x3E, 0x78, 0x27, 0xC2, 0x27, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x03, 0x24, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x02, 0xA4, 0x95, 0x00,
		0x22, 0x48, 0x21, 0x02, 0x64, 0x9B, 0x00,
		0x3C, 0x78, 0x21, 0x02, 0x27, 0x91, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x15, 0x50, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	bm, err := NewWithData(50, 18, data)
	require.NoError(t, err)

	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

	t.Run("Connectivity8", func(t *testing.T) {
		// for connectivity 8 there should be 8 components:
		// letters: d, o, i, t, n, o, w
		// and a skull image.
		t.Run("WithBitmaps", func(t *testing.T) {
			bitmaps := &Bitmaps{}

			bm, err := copyBitmap(nil, bm)
			require.NoError(t, err)

			boxes, err := bm.ConnComponents(bitmaps, 8)
			require.NoError(t, err)

			if assert.Len(t, bitmaps.Values, 8) && assert.Len(t, *boxes, 8) {
				for i, bm := range bitmaps.Values {
					box := bitmaps.Boxes[i]
					switch i {
					case 0:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'D' - expected data:
						//
						// 11111000  0xF8
						// 10001000  0x88
						// 10001000  0x88
						// 10001000  0x88
						// 11110000  0xF0
						assert.Equal(t, []byte{0xF8, 0x88, 0x88, 0x88, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 2)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 6+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 1:
						assert.Equal(t, 4, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'O' - expected data:
						//
						// 11110000 - 0xF0
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 11110000 - 0xF0
						assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 9)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 12+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 2:
						assert.Equal(t, 1, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'I' - expected data:
						//
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						assert.Equal(t, []byte{0x80, 0x80, 0x80, 0x80, 0x80}, bm.Data)
						assert.Equal(t, box.Min.X, 18)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 18+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 3:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Width)
						// Letter 'T' - expected data:
						//
						// 11111000 - 0xF8
						// 00100000 - 0x20
						// 00100000 - 0x20
						// 00100000 - 0x20
						// 00100000 - 0x20
						assert.Equal(t, []byte{0xF8, 0x20, 0x20, 0x20, 0x20}, bm.Data)
						assert.Equal(t, box.Min.X, 21)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 25+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 4:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'N' - expected data:
						//
						// 10001000 - 0x88
						// 11001000 - 0xC8
						// 10101000 - 0xA8
						// 10011000 - 0x98
						// 10001000 - 0x88
						assert.Equal(t, []byte{0x88, 0xC8, 0xA8, 0x98, 0x88}, bm.Data)
						assert.Equal(t, box.Min.X, 30)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 34+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 5:
						assert.Equal(t, 4, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'O' - expected data:
						//
						// 11110000 - 0xF0
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 11110000 - 0xF0
						assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 37)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 40+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 6:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'W' - expected data:
						//
						// 10001000 - 0x88
						// 10001000 - 0x88
						// 10101000 - 0xA8
						// 11011000 - 0xD8
						// 10001000 - 0x88
						assert.Equal(t, []byte{0x88, 0x88, 0xA8, 0xD8, 0x88}, bm.Data)
						assert.Equal(t, box.Min.X, 43)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 47+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 7:
						// the last one is the the image
						//
						// 11111111 11110000 - 0xFF, 0xF0
						// 11111111 11110000 - 0xFF, 0xF0
						// 11000110 00110000 - 0xC6, 0x30
						// 11000110 00110000 - 0xC6, 0x30
						// 11000110 00110000 - 0xC6, 0x30
						// 11111111 11110000 - 0xFF, 0xF0
						// 00101010 10100000 - 0x2A, 0xA0
						assert.Equal(t, 12, bm.Width)
						assert.Equal(t, 7, bm.Height)
						assert.Equal(t, []byte{0xFF, 0xF0, 0xFF, 0xF0, 0xC6, 0x30, 0xC6, 0x30, 0xC6, 0x30, 0xFF, 0xF0, 0x2A, 0xA0}, bm.Data)
						assert.Equal(t, box.Min.X, 17)
						assert.Equal(t, box.Min.Y, 9)
						assert.Equal(t, box.Max.X, 28+1)
						assert.Equal(t, box.Max.Y, 15+1)
					}
				}
			}
		})

		t.Run("OnlyBoxes", func(t *testing.T) {
			bm, err := copyBitmap(nil, bm)
			require.NoError(t, err)

			boxes, err := bm.ConnComponents(nil, 8)
			require.NoError(t, err)

			require.Len(t, *boxes, 8)
			for i, box := range *boxes {
				switch i {
				case 0:
					assert.Equal(t, box.Min.X, 2)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 6+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 1:
					assert.Equal(t, box.Min.X, 9)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 12+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 2:
					assert.Equal(t, box.Min.X, 18)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 18+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 3:
					assert.Equal(t, box.Min.X, 21)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 25+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 4:
					assert.Equal(t, box.Min.X, 30)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 34+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 5:
					assert.Equal(t, box.Min.X, 37)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 40+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 6:
					assert.Equal(t, box.Min.X, 43)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 47+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 7:
					assert.Equal(t, box.Min.X, 17)
					assert.Equal(t, box.Min.Y, 9)
					assert.Equal(t, box.Max.X, 28+1)
					assert.Equal(t, box.Max.Y, 15+1)
				}
			}
		})
	})

	t.Run("Connectivity4", func(t *testing.T) {
		// Having the connectivity of '4' the letters that
		// are connected using their corners are not treated
		// as single letter.
		// Thus when there were '8' symbols in '8' connectivity
		// the letter 'N' and 'W' both disassembles into 3 copmonents.
		// This gives us 8-2+3+3 = 12 classes.
		t.Run("WithBitmaps", func(t *testing.T) {
			bitmaps := &Bitmaps{}

			bm, err := copyBitmap(nil, bm)
			require.NoError(t, err)

			boxes, err := bm.ConnComponents(bitmaps, 4)
			require.NoError(t, err)

			require.NotNil(t, boxes)

			cond := assert.Len(t, bitmaps.Values, 12) && assert.Len(t, *boxes, 12) && assert.Len(t, bitmaps.Boxes, 12)
			require.True(t, cond)

			for i, bm := range bitmaps.Values {
				// The box in bitmaps must be equal to box as an output.
				box := bitmaps.Boxes[i]
				assert.Equal(t, box, (*boxes)[i])
				switch i {
				case 0:
					assert.Equal(t, 5, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'D' - expected data:
					//
					// 11111000  0xF8
					// 10001000  0x88
					// 10001000  0x88
					// 10001000  0x88
					// 11110000  0xF0
					assert.Equal(t, []byte{0xF8, 0x88, 0x88, 0x88, 0xF0}, bm.Data)
					assert.Equal(t, box.Min.X, 2)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 6+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 1:
					assert.Equal(t, 4, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'O' - expected data:
					//
					// 11110000 - 0xF0
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 11110000 - 0xF0
					assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
					assert.Equal(t, box.Min.X, 9)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 12+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 2:
					assert.Equal(t, 1, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'I' - expected data:
					//
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					assert.Equal(t, []byte{0x80, 0x80, 0x80, 0x80, 0x80}, bm.Data)
					assert.Equal(t, box.Min.X, 18)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 18+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 3:
					assert.Equal(t, 5, bm.Width)
					assert.Equal(t, 5, bm.Width)
					// Letter 'T' - expected data:
					//
					// 11111000 - 0xF8
					// 00100000 - 0x20
					// 00100000 - 0x20
					// 00100000 - 0x20
					// 00100000 - 0x20
					assert.Equal(t, []byte{0xF8, 0x20, 0x20, 0x20, 0x20}, bm.Data)
					assert.Equal(t, box.Min.X, 21)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 25+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 4:
					assert.Equal(t, 2, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Left part of letter 'N' - expected data:
					//
					// 10000000 - 0x80
					// 11000000 - 0xC0
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					assert.Equal(t, []byte{0x80, 0xC0, 0x80, 0x80, 0x80}, bm.Data)
					assert.Equal(t, box.Min.X, 30)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 31+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 5:
					// A right side of the letter 'N' - expected data:
					//
					// 01000000 - 0x40
					// 01000000 - 0x40
					// 01000000 - 0x40
					// 11000000 - 0xC0
					// 01000000 - 0x40
					assert.Equal(t, 2, bm.Width)
					assert.Equal(t, 5, bm.Height)
					assert.Equal(t, []byte{0x40, 0x40, 0x40, 0xC0, 0x40}, bm.Data)
					assert.Equal(t, box.Min.X, 33)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 34+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 6:
					assert.Equal(t, 4, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'O' - expected data:
					//
					// 11110000 - 0xF0
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 11110000 - 0xF0
					assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
					assert.Equal(t, box.Min.X, 37)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 40+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 7:
					// Left part of the letter 'W' - expected data:
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 11000000 - 0xC0
					// 10000000 - 0x80
					assert.Equal(t, 2, bm.Width)
					assert.Equal(t, 5, bm.Height)
					assert.Equal(t, []byte{0x80, 0x80, 0x80, 0xC0, 0x80}, bm.Data)
					assert.Equal(t, box.Min.X, 43)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 44+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 8:
					// The right part of the letter 'W' - expected data:
					//
					// 01000000 - 0x40
					// 01000000 - 0x40
					// 01000000 - 0x40
					// 11000000 - 0xC0
					// 01000000 - 0x40
					assert.Equal(t, 2, bm.Width)
					assert.Equal(t, 5, bm.Height)
					assert.Equal(t, []byte{0x40, 0x40, 0x40, 0xC0, 0x40}, bm.Data)
					assert.Equal(t, box.Min.X, 46)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 47+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 9:
					// A single dot in the middle of 'N' letter - expected data:
					// 10000000 - 0x80
					assert.Equal(t, 1, bm.Width, bm)
					assert.Equal(t, 1, bm.Height)
					assert.Equal(t, []byte{0x80}, bm.Data)
					assert.Equal(t, box.Min.X, 32)
					assert.Equal(t, box.Min.Y, 4)
					assert.Equal(t, box.Max.X, 32+1)
					assert.Equal(t, box.Max.Y, 4+1)
				case 10:
					// The dot in the middle of the 'W' letter
					// 100000000 - 0x80
					assert.Equal(t, 1, bm.Width)
					assert.Equal(t, 1, bm.Height)
					assert.Equal(t, []byte{0x80}, bm.Data)
					assert.Equal(t, box.Min.X, 45)
					assert.Equal(t, box.Min.Y, 4)
					assert.Equal(t, box.Max.X, 45+1)
					assert.Equal(t, box.Max.Y, 4+1)
				case 11:
					// the last one is the the image
					//
					// 11111111 11110000 - 0xFF, 0xF0
					// 11111111 11110000 - 0xFF, 0xF0
					// 11000110 00110000 - 0xC6, 0x30
					// 11000110 00110000 - 0xC6, 0x30
					// 11000110 00110000 - 0xC6, 0x30
					// 11111111 11110000 - 0xFF, 0xF0
					// 00101010 10100000 - 0x2A, 0xA0
					assert.Equal(t, 12, bm.Width)
					assert.Equal(t, 7, bm.Height)
					assert.Equal(t, []byte{0xFF, 0xF0, 0xFF, 0xF0, 0xC6, 0x30, 0xC6, 0x30, 0xC6, 0x30, 0xFF, 0xF0, 0x2A, 0xA0}, bm.Data)
					assert.Equal(t, box.Min.X, 17)
					assert.Equal(t, box.Min.Y, 9)
					assert.Equal(t, box.Max.X, 28+1)
					assert.Equal(t, box.Max.Y, 15+1)
				}
			}
		})

		t.Run("OnlyBoxes", func(t *testing.T) {
			bm, err := copyBitmap(nil, bm)
			require.NoError(t, err)

			boxes, err := bm.ConnComponents(nil, 4)
			require.NoError(t, err)

			require.Len(t, *boxes, 12)
			for i, box := range *boxes {
				switch i {
				case 0:
					assert.Equal(t, box.Min.X, 2)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 6+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 1:
					assert.Equal(t, box.Min.X, 9)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 12+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 2:
					assert.Equal(t, box.Min.X, 18)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 18+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 3:
					assert.Equal(t, box.Min.X, 21)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 25+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 4:
					assert.Equal(t, box.Min.X, 30)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 31+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 5:
					assert.Equal(t, box.Min.X, 33)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 34+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 6:
					assert.Equal(t, box.Min.X, 37)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 40+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 7:
					assert.Equal(t, box.Min.X, 43)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 44+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 8:
					assert.Equal(t, box.Min.X, 46)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 47+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 9:
					assert.Equal(t, box.Min.X, 32)
					assert.Equal(t, box.Min.Y, 4)
					assert.Equal(t, box.Max.X, 32+1)
					assert.Equal(t, box.Max.Y, 4+1)
				case 10:
					assert.Equal(t, box.Min.X, 45)
					assert.Equal(t, box.Min.Y, 4)
					assert.Equal(t, box.Max.X, 45+1)
					assert.Equal(t, box.Max.Y, 4+1)
				case 11:
					assert.Equal(t, box.Min.X, 17)
					assert.Equal(t, box.Min.Y, 9)
					assert.Equal(t, box.Max.X, 28+1)
					assert.Equal(t, box.Max.Y, 15+1)
				}
			}
		})
	})

	t.Run("Invalid", func(t *testing.T) {
		t.Run("NoBitmap", func(t *testing.T) {
			var bm *Bitmap
			_, err := bm.ConnComponents(nil, 8)
			require.Error(t, err)
		})

		t.Run("Connectivity", func(t *testing.T) {
			// having any bitmap with some connectivity different
			// then '4' and '8' - i.e. '2' should return error.
			bm := New(10, 10)
			_, err := bm.ConnComponents(nil, 2)
			require.Error(t, err)
		})
	})
}

// TestGetComponents tests the GetComponents function.
func TestGetComponents(t *testing.T) {
	// Having a bitmap 50x22 with few different components like letters.
	//
	//            1        2        3        4        5        6         7
	//
	//                111111 11112222 22222233 33333333 44444444 44
	//     01234567 89012345 67890123 45678901 23456789 01234567 89
	//
	//  0  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  1  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  2  00111110 01111000 00100111 11000010 00100111 10010001 00000000
	//  3  00100010 01001000 00100001 00000011 00100100 10010001 00000000
	//  4  00100010 01001000 00100001 00000010 10100100 10010101 00000000
	//  5  00100010 01001000 00100001 00000010 01100100 10011011 00000000
	//  6  00111100 01111000 00100001 00000010 00100111 10010001 00000000
	//  7  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  8  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  9  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 10  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 11  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 12  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 13  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 14  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 15  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 16  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 17  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 18  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 19  00000000 00000000 00010101 01010000 00000000 00000000 00000000
	// 20  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 21  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//
	//
	data := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x3E, 0x78, 0x27, 0xC2, 0x27, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x03, 0x24, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x02, 0xA4, 0x95, 0x00,
		0x22, 0x48, 0x21, 0x02, 0x64, 0x9B, 0x00,
		0x3C, 0x78, 0x21, 0x02, 0x27, 0x91, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x15, 0x50, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	bm, err := NewWithData(50, 22, data)
	require.NoError(t, err)

	common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))

	t.Run("ComponentConn", func(t *testing.T) {
		t.Run("Max12-8", func(t *testing.T) {
			// the highest component is 12 bits wide and 8bits tall.
			// this should get all possible components out of the given input bitmap.
			bm, err = copyBitmap(nil, bm)
			require.NoError(t, err)

			bitmaps, boxes, err := bm.GetComponents(ComponentConn, 12, 8)
			require.NoError(t, err)

			require.NotNil(t, bitmaps)
			require.NotNil(t, boxes)

			if assert.Len(t, bitmaps.Values, 8) && assert.Len(t, *boxes, 8) {
				for i, bm := range bitmaps.Values {
					box := bitmaps.Boxes[i]
					switch i {
					case 0:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'D' - expected data:
						//
						// 11111000  0xF8
						// 10001000  0x88
						// 10001000  0x88
						// 10001000  0x88
						// 11110000  0xF0
						assert.Equal(t, []byte{0xF8, 0x88, 0x88, 0x88, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 2)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 6+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 1:
						assert.Equal(t, 4, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'O' - expected data:
						//
						// 11110000 - 0xF0
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 11110000 - 0xF0
						assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 9)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 12+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 2:
						assert.Equal(t, 1, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'I' - expected data:
						//
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						assert.Equal(t, []byte{0x80, 0x80, 0x80, 0x80, 0x80}, bm.Data)
						assert.Equal(t, box.Min.X, 18)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 18+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 3:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Width)
						// Letter 'T' - expected data:
						//
						// 11111000 - 0xF8
						// 00100000 - 0x20
						// 00100000 - 0x20
						// 00100000 - 0x20
						// 00100000 - 0x20
						assert.Equal(t, []byte{0xF8, 0x20, 0x20, 0x20, 0x20}, bm.Data)
						assert.Equal(t, box.Min.X, 21)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 25+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 4:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'N' - expected data:
						//
						// 10001000 - 0x88
						// 11001000 - 0xC8
						// 10101000 - 0xA8
						// 10011000 - 0x98
						// 10001000 - 0x88
						assert.Equal(t, []byte{0x88, 0xC8, 0xA8, 0x98, 0x88}, bm.Data)
						assert.Equal(t, box.Min.X, 30)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 34+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 5:
						assert.Equal(t, 4, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'O' - expected data:
						//
						// 11110000 - 0xF0
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 11110000 - 0xF0
						assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 37)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 40+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 6:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'W' - expected data:
						//
						// 10001000 - 0x88
						// 10001000 - 0x88
						// 10101000 - 0xA8
						// 11011000 - 0xD8
						// 10001000 - 0x88
						assert.Equal(t, []byte{0x88, 0x88, 0xA8, 0xD8, 0x88}, bm.Data)
						assert.Equal(t, box.Min.X, 43)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 47+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 7:
						// the last one is the the image
						//
						// 11111111 11110000 - 0xFF, 0xF0
						// 11111111 11110000 - 0xFF, 0xF0
						// 11000110 00110000 - 0xC6, 0x30
						// 11000110 00110000 - 0xC6, 0x30
						// 11000110 00110000 - 0xC6, 0x30
						// 11111111 11110000 - 0xFF, 0xF0
						// 00101010 10100000 - 0x2A, 0xA0
						assert.Equal(t, 12, bm.Width)
						assert.Equal(t, 7, bm.Height)
						assert.Equal(t, []byte{0xFF, 0xF0, 0xFF, 0xF0, 0xC6, 0x30, 0xC6, 0x30, 0xC6, 0x30, 0xFF, 0xF0, 0x2A, 0xA0}, bm.Data)
						assert.Equal(t, box.Min.X, 17)
						assert.Equal(t, box.Min.Y, 13)
						assert.Equal(t, box.Max.X, 28+1)
						assert.Equal(t, box.Max.Y, 19+1)
					}
				}
			}
		})

		t.Run("Max10-6", func(t *testing.T) {
			bm, err := copyBitmap(nil, bm)
			require.NoError(t, err)

			bitmaps, boxes, err := bm.GetComponents(ComponentConn, 10, 6)
			require.NoError(t, err)

			if assert.Len(t, bitmaps.Values, 7) && assert.Len(t, *boxes, 7) {
				for i, bm := range bitmaps.Values {
					box := bitmaps.Boxes[i]
					switch i {
					case 0:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'D' - expected data:
						//
						// 11111000  0xF8
						// 10001000  0x88
						// 10001000  0x88
						// 10001000  0x88
						// 11110000  0xF0
						assert.Equal(t, []byte{0xF8, 0x88, 0x88, 0x88, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 2)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 6+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 1:
						assert.Equal(t, 4, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'O' - expected data:
						//
						// 11110000 - 0xF0
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 11110000 - 0xF0
						assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 9)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 12+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 2:
						assert.Equal(t, 1, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'I' - expected data:
						//
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						// 10000000 - 0x80
						assert.Equal(t, []byte{0x80, 0x80, 0x80, 0x80, 0x80}, bm.Data)
						assert.Equal(t, box.Min.X, 18)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 18+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 3:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Width)
						// Letter 'T' - expected data:
						//
						// 11111000 - 0xF8
						// 00100000 - 0x20
						// 00100000 - 0x20
						// 00100000 - 0x20
						// 00100000 - 0x20
						assert.Equal(t, []byte{0xF8, 0x20, 0x20, 0x20, 0x20}, bm.Data)
						assert.Equal(t, box.Min.X, 21)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 25+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 4:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'N' - expected data:
						//
						// 10001000 - 0x88
						// 11001000 - 0xC8
						// 10101000 - 0xA8
						// 10011000 - 0x98
						// 10001000 - 0x88
						assert.Equal(t, []byte{0x88, 0xC8, 0xA8, 0x98, 0x88}, bm.Data)
						assert.Equal(t, box.Min.X, 30)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 34+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 5:
						assert.Equal(t, 4, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'O' - expected data:
						//
						// 11110000 - 0xF0
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 10010000 - 0x90
						// 11110000 - 0xF0
						assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
						assert.Equal(t, box.Min.X, 37)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 40+1)
						assert.Equal(t, box.Max.Y, 6+1)
					case 6:
						assert.Equal(t, 5, bm.Width)
						assert.Equal(t, 5, bm.Height)
						// Letter 'W' - expected data:
						//
						// 10001000 - 0x88
						// 10001000 - 0x88
						// 10101000 - 0xA8
						// 11011000 - 0xD8
						// 10001000 - 0x88
						assert.Equal(t, []byte{0x88, 0x88, 0xA8, 0xD8, 0x88}, bm.Data)
						assert.Equal(t, box.Min.X, 43)
						assert.Equal(t, box.Min.Y, 2)
						assert.Equal(t, box.Max.X, 47+1)
						assert.Equal(t, box.Max.Y, 6+1)
					}
				}
			}
		})
	})

	t.Run("Char", func(t *testing.T) {
		// the highest component is 12 bits wide and 8bits tall.
		// this should get all possible components out of the given input bitmap.
		bm, err = copyBitmap(nil, bm)
		require.NoError(t, err)

		bitmaps, boxes, err := bm.GetComponents(ComponentCharacters, 12, 8)
		require.NoError(t, err)

		require.NotNil(t, bitmaps)
		require.NotNil(t, boxes)

		if assert.Len(t, bitmaps.Values, 8) && assert.Len(t, *boxes, 8) {
			for i, bm := range bitmaps.Values {
				box := bitmaps.Boxes[i]
				switch i {
				case 0:
					assert.Equal(t, 5, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'D' - expected data:
					//
					// 11111000  0xF8
					// 10001000  0x88
					// 10001000  0x88
					// 10001000  0x88
					// 11110000  0xF0
					assert.Equal(t, []byte{0xF8, 0x88, 0x88, 0x88, 0xF0}, bm.Data)
					assert.Equal(t, box.Min.X, 2)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 6+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 1:
					assert.Equal(t, 4, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'O' - expected data:
					//
					// 11110000 - 0xF0
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 11110000 - 0xF0
					assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
					assert.Equal(t, box.Min.X, 9)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 12+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 2:
					assert.Equal(t, 1, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'I' - expected data:
					//
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					// 10000000 - 0x80
					assert.Equal(t, []byte{0x80, 0x80, 0x80, 0x80, 0x80}, bm.Data)
					assert.Equal(t, box.Min.X, 18)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 18+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 3:
					assert.Equal(t, 5, bm.Width)
					assert.Equal(t, 5, bm.Width)
					// Letter 'T' - expected data:
					//
					// 11111000 - 0xF8
					// 00100000 - 0x20
					// 00100000 - 0x20
					// 00100000 - 0x20
					// 00100000 - 0x20
					assert.Equal(t, []byte{0xF8, 0x20, 0x20, 0x20, 0x20}, bm.Data)
					assert.Equal(t, box.Min.X, 21)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 25+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 4:
					assert.Equal(t, 5, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'N' - expected data:
					//
					// 10001000 - 0x88
					// 11001000 - 0xC8
					// 10101000 - 0xA8
					// 10011000 - 0x98
					// 10001000 - 0x88
					assert.Equal(t, []byte{0x88, 0xC8, 0xA8, 0x98, 0x88}, bm.Data)
					assert.Equal(t, box.Min.X, 30)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 34+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 5:
					assert.Equal(t, 4, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'O' - expected data:
					//
					// 11110000 - 0xF0
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 10010000 - 0x90
					// 11110000 - 0xF0
					assert.Equal(t, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0}, bm.Data)
					assert.Equal(t, box.Min.X, 37)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 40+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 6:
					assert.Equal(t, 5, bm.Width)
					assert.Equal(t, 5, bm.Height)
					// Letter 'W' - expected data:
					//
					// 10001000 - 0x88
					// 10001000 - 0x88
					// 10101000 - 0xA8
					// 11011000 - 0xD8
					// 10001000 - 0x88
					assert.Equal(t, []byte{0x88, 0x88, 0xA8, 0xD8, 0x88}, bm.Data)
					assert.Equal(t, box.Min.X, 43)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 47+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 7:
					// the last one is the the image
					//
					// 11111111 11110000 - 0xFF, 0xF0
					// 11111111 11110000 - 0xFF, 0xF0
					// 11000110 00110000 - 0xC6, 0x30
					// 11000110 00110000 - 0xC6, 0x30
					// 11000110 00110000 - 0xC6, 0x30
					// 11111111 11110000 - 0xFF, 0xF0
					// 00101010 10100000 - 0x2A, 0xA0
					assert.Equal(t, 12, bm.Width)
					assert.Equal(t, 7, bm.Height)
					assert.Equal(t, []byte{0xFF, 0xF0, 0xFF, 0xF0, 0xC6, 0x30, 0xC6, 0x30, 0xC6, 0x30, 0xFF, 0xF0, 0x2A, 0xA0}, bm.Data)
					assert.Equal(t, box.Min.X, 17)
					assert.Equal(t, box.Min.Y, 13)
					assert.Equal(t, box.Max.X, 28+1)
					assert.Equal(t, box.Max.Y, 19+1)
				}
			}
		}
	})

	t.Run("Word", func(t *testing.T) {
		// the highest component is 20 bits wide and 8bits tall.
		bm, err = copyBitmap(nil, bm)
		require.NoError(t, err)

		bitmaps, boxes, err := bm.GetComponents(ComponentWords, 20, 8)
		require.NoError(t, err)

		require.NotNil(t, bitmaps)
		require.NotNil(t, boxes)

		// this should result in a three words and a image components:
		if assert.Len(t, bitmaps.Values, 4) && assert.Len(t, *boxes, 4) {
			for i, word := range bitmaps.Values {
				box := (*boxes)[i]
				switch i {
				case 0:
					// the first word should look like 'DO'.
					//
					// 11111001 11100000	- 0xF9, 0xE0
					// 10001001 00100000	- 0x89, 0x20
					// 10001001 00100000	- 0x89, 0x20
					// 10001001 00100000	- 0x89, 0x20
					// 11110001 11100000	- 0xF1, 0xE0
					data := []byte{0xF9, 0xE0, 0x89, 0x20, 0x89, 0x20, 0x89, 0x20, 0xF1, 0xE0}
					assert.Equal(t, data, word.Data)
					assert.Equal(t, box.Min.X, 2)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 12+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 1:
					// the second word should be like 'IT'
					//
					// 10011111	- 0x9F
					// 10000100	- 0x84
					// 10000100	- 0x84
					// 10000100	- 0x84
					// 10000100	- 0x84
					data := []byte{0x9F, 0x84, 0x84, 0x84, 0x84}
					assert.Equal(t, data, word.Data)
					assert.Equal(t, box.Min.X, 18)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 25+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 2:
					// the third word should look like 'NOW'
					//
					// 10001001 11100100 01000000	- 0x89, 0xE4, 0x40
					// 11001001 00100100 01000000	- 0xC9, 0x24, 0x40
					// 10101001 00100101 01000000	- 0xA9, 0x25, 0x40
					// 10011001 00100110 11000000	- 0x99, 0x26, 0xC0
					// 10001001 11100100 01000000	- 0x89,	0xE4, 0x40
					data := []byte{
						0x89, 0xE4, 0x40,
						0xC9, 0x24, 0x40,
						0xA9, 0x25, 0x40,
						0x99, 0x26, 0xC0,
						0x89, 0xE4, 0x40,
					}
					assert.Equal(t, data, word.Data)
					assert.Equal(t, box.Min.X, 30)
					assert.Equal(t, box.Min.Y, 2)
					assert.Equal(t, box.Max.X, 47+1)
					assert.Equal(t, box.Max.Y, 6+1)
				case 3:
					//  the last component is the image of the 'skull'i
					//
					// 11111111 11110000 - 0xFF, 0xF0
					// 11111111 11110000 - 0xFF, 0xF0
					// 11000110 00110000 - 0xC6, 0x30
					// 11000110 00110000 - 0xC6, 0x30
					// 11000110 00110000 - 0xC6, 0x30
					// 11111111 11110000 - 0xFF, 0xF0
					// 00101010 10100000 - 0x2A, 0xA0
					data := []byte{
						0xFF, 0xF0,
						0xFF, 0xF0,
						0xC6, 0x30,
						0xC6, 0x30,
						0xC6, 0x30,
						0xFF, 0xF0,
						0x2A, 0xA0,
					}

					assert.Equal(t, 12, word.Width)
					assert.Equal(t, 7, word.Height)
					assert.Equal(t, data, word.Data, word.String())
					assert.Equal(t, box.Min.X, 17)
					assert.Equal(t, box.Min.Y, 13)
					assert.Equal(t, box.Max.X, 28+1)
					assert.Equal(t, box.Max.Y, 19+1)
				}
			}
		}
	})

	t.Run("Zero", func(t *testing.T) {
		// Create a zero like bitmap - it should not create any component.
		bm := New(bm.Width, bm.Height)

		bitmaps, boxes, err := bm.GetComponents(ComponentConn, 12, 8)
		require.NoError(t, err)

		if assert.NotNil(t, bitmaps) {
			assert.Empty(t, bitmaps.Values)
		}

		if assert.NotNil(t, boxes) {
			assert.Empty(t, *boxes)
		}
	})

	t.Run("EmptyBitmap", func(t *testing.T) {
		var b *Bitmap
		_, _, err := b.GetComponents(ComponentConn, 12, 8)
		assert.Error(t, err)
	})
	// TODO: Get a bitmap with the symbols on the edge. Check if

	t.Run("NoEmptyFrame", func(t *testing.T) {
		// In this scenario the bitmap is filled with the letters
		// directly from the left edge to the right edge.
		bm := TstWordBitmap(t, 2)

		sm, _, err := bm.GetComponents(ComponentConn, 100, 100)
		require.NoError(t, err)

		for i, symbol := range sm.Values {
			switch i {
			case 0:
				// letter D
				d := TstDSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 1:
				// letter O
				d := TstOSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 2:
				// letter I
				d := TstISymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 3:
				// letter T
				d := TstTSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 4:
				// letter N
				d := TstNSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 5:
				// letter O
				d := TstOSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 6:
				// letter W
				d := TstWSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 7:
				// letter O
				d := TstOSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 8:
				// letter R
				d := TstRSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 9:
				// letter N
				d := TstNSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 10:
				// letter E
				d := TstESymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 11:
				// letter V
				d := TstVSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 12:
				// letter E
				d := TstESymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			case 13:
				// letter R
				d := TstRSymbol(t, 2)
				assert.Equal(t, d.Data, symbol.Data)
			}
		}

	})
}

// TestWordMaskByDilation tests the WordMaskByDilation function.
func TestWordMaskByDilation(t *testing.T) {
	// Having a bitmap 50x18 with few different components like letters
	//
	//            1        2        3        4        5        6         7
	//
	//                111111 11112222 22222233 33333333 44444444 44
	//     01234567 89012345 67890123 45678901 23456789 01234567 89
	//
	//  0  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  1  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  2  00111110 01111000 00100111 11000010 00100111 10010001 00000000
	//  3  00100010 01001000 00100001 00000011 00100100 10010001 00000000
	//  4  00100010 01001000 00100001 00000010 10100100 10010101 00000000
	//  5  00100010 01001000 00100001 00000010 01100100 10011011 00000000
	//  6  00111100 01111000 00100001 00000010 00100111 10010001 00000000
	//  7  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  8  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  9  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 10  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 11  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 12  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 13  00000000 00000000 01100011 00011000 00000000 00000000 00000000
	// 14  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 15  00000000 00000000 00010101 01010000 00000000 00000000 00000000
	// 16  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 17  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//
	//
	data := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x3E, 0x78, 0x27, 0xC2, 0x27, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x03, 0x24, 0x91, 0x00,
		0x22, 0x48, 0x21, 0x02, 0xA4, 0x95, 0x00,
		0x22, 0x48, 0x21, 0x02, 0x64, 0x9B, 0x00,
		0x3C, 0x78, 0x21, 0x02, 0x27, 0x91, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x15, 0x50, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	bm, err := NewWithData(50, 18, data)
	require.NoError(t, err)

	masks, _, err := wordMaskByDilation(bm)
	require.NoError(t, err)

	// the function wordMaskByDilation should create a mask bitmap that
	// would set all the box areas for each copmonent of the input
	// with all 'ONE' by using dilation.
	//
	//            1        2        3        4        5        6         7
	//
	//                111111 11112222 22222233 33333333 44444444 44
	//     01234567 89012345 67890123 45678901 23456789 01234567 89
	//
	//  0  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  1  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  2  00111111 11111000 00111111 11000011 11111111 11111111 00000000
	//  3  00111111 11111000 00100001 00000011 11111111 11111111 00000000
	//  4  00111111 11111000 00100001 00000011 11111111 11111111 00000000
	//  5  00111111 11111000 00100001 00000011 11111111 11111111 00000000
	//  6  00111111 11111000 00100001 00000011 11111111 11111111 00000000
	//  7  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  8  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//  9  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 10  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 11  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 12  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 13  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 14  00000000 00000000 01111111 11111000 00000000 00000000 00000000
	// 15  00000000 00000000 00011111 11110000 00000000 00000000 00000000
	// 16  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	// 17  00000000 00000000 00000000 00000000 00000000 00000000 00000000
	//
	result := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x3F, 0xF8, 0x3F, 0xC3, 0xFF, 0xFF, 0x00,
		0x3F, 0xF8, 0x21, 0x03, 0xFF, 0xFF, 0x00,
		0x3F, 0xF8, 0x21, 0x03, 0xFF, 0xFF, 0x00,
		0x3F, 0xF8, 0x21, 0x03, 0xFF, 0xFF, 0x00,
		0x3F, 0xF8, 0x21, 0x03, 0xFF, 0xFF, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x1F, 0xF0, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	assert.Equal(t, result, masks.Data)
}
