/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package writer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBufferFinishByte tests the FinishByte method of the Buffer.
func TestBufferFinishByte(t *testing.T) {
	for i := 0; i < 2; i++ {
		// Having a buffer with some bits already written.
		var (
			name string
			w    *Buffer
		)
		switch i {
		case 0:
			name = "LSB"
			w = &Buffer{}
		case 1:
			name = "MSB"
			w = BufferedMSB()
		}
		t.Run(name, func(t *testing.T) {
			bits := []int{1, 0, 1}
			for _, bit := range bits {
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}
			// the bit index is now at the third position
			assert.Equal(t, uint8(3), w.bitIndex)
			// and the byte index is at 0'th position.
			assert.Equal(t, 0, w.byteIndex)

			// by using the FinishByte method the bitIndex should be reset to 0
			w.FinishByte()
			assert.Equal(t, uint8(0), w.bitIndex)
			// and the byte index should point to the next byte
			assert.Equal(t, 1, w.byteIndex)
			// and the data should contain only a single byte
			assert.Equal(t, 1, len(w.data))
			// but the cap should be much equal to smallSize
			assert.Equal(t, smallSize, cap(w.data))
		})
	}
}

// TestBufferSkipBits tests the SkipBits method.
func TestBufferSkipBits(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		t.Run("OutOfRange", func(t *testing.T) {
			b := BufferedMSB()
			// make space for the 4 bytes
			b.expandIfNeeded(4)

			// assuming that there are 2 bytes and 2 bits
			b.byteIndex = 2
			b.bitIndex = 2
			// the index should be at position:
			// 00000000 00000000 00000000 00000000
			//                     ^
			// in order to get out of the range the function should skip:
			// -((2 * 8) + 2 + 1) bits
			skip := -(2*8 + 2 + 1)
			err := b.SkipBits(skip)
			assert.Error(t, err, "%v", b)
		})

		t.Run("WithinSingleByte", func(t *testing.T) {
			b := BufferedMSB()
			// make space for the 4 bytes.
			b.expandIfNeeded(4)

			// assumming that there are 2 bytes and 5 bits
			b.byteIndex = 2
			b.bitIndex = 5

			// the index should be at position:
			// 00000000 00000000 00000000 00000000
			//                        ^
			// skipping -3 bits
			err := b.SkipBits(-3)
			require.NoError(t, err)

			assert.Equal(t, uint8(2), b.bitIndex)
			assert.Equal(t, 2, b.byteIndex)
		})

		t.Run("ToFirstByte", func(t *testing.T) {
			b := BufferedMSB()
			// make space for the 4 bytes.
			b.expandIfNeeded(4)

			// assumming that there are 2 bytes and 5 bits
			b.byteIndex = 2
			b.bitIndex = 5

			// the index should be at position:
			// 00000000 00000000 00000000 00000000
			//                        ^
			// skipping -17 bits
			err := b.SkipBits(-17)
			require.NoError(t, err)

			// should result in indexes
			// 00000000 00000000 00000000 00000000
			//     ^
			assert.Equal(t, uint8(4), b.bitIndex)
			assert.Equal(t, 0, b.byteIndex)
		})
	})

	t.Run("Positive", func(t *testing.T) {
		t.Run("Overflow", func(t *testing.T) {
			b := BufferedMSB()

			// empty buffer should be expanded for the number of full bytes to write.
			// i.e. skipping 20 bits should result with a least 3 bytes length data.
			err := b.SkipBits(20)
			require.NoError(t, err)

			// there must be 3 empty bytes
			assert.Len(t, b.data, 3, "%v", b.data)
			// and the bit index should be at the '3' position.
			assert.Equal(t, uint8(4), b.bitIndex)
			assert.Equal(t, 2, b.byteIndex)
		})

		t.Run("WithinSingleByte", func(t *testing.T) {
			b := BufferedMSB()

			b.expandIfNeeded(1)

			b.bitIndex = 3
			err := b.SkipBits(2)
			require.NoError(t, err)

			// the byte index should not change.
			assert.Equal(t, 0, b.byteIndex)
			// the bit index should be moved forward by the nubmer of skipped bits = 3 + 2 = 5
			assert.Equal(t, uint8(3+2), b.bitIndex)
		})

		t.Run("NoUnnecesaryExpand", func(t *testing.T) {
			b := BufferedMSB()

			// lets create three bytes space.
			b.expandIfNeeded(3)
			// the data should now look like: []byte{0x00,0x00,0x00}
			require.Equal(t, []byte{0x00, 0x00, 0x00}, b.data)

			// let's assume that the byte index is at the '1' position
			b.byteIndex = 1

			// by skipping the bits by one full byte - 8
			// the byte index should be equal to 2 and bitIndex = 0
			// In this case the data slice should still contains only three bytes -> should not be expanded.
			err := b.SkipBits(8)
			require.NoError(t, err)

			assert.Equal(t, 2, b.byteIndex)
			assert.Equal(t, uint8(0), b.bitIndex)
			assert.Equal(t, []byte{0x00, 0x00, 0x00}, b.data)
		})
	})
}

// TestBufferWrite tests the Write method of the Writer.
func TestBufferWrite(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		w := &Buffer{}

		toWrite := []byte{0x3f, 0x12, 0x86}

		n, err := w.Write(toWrite)
		require.NoError(t, err)

		assert.Equal(t, 3, n)

		n, err = w.Write([]byte{0xff})
		require.NoError(t, err)

		assert.Equal(t, 1, n)

		expected := append(toWrite, 0xff)
		for i, bt := range w.Data() {
			assert.Equal(t, expected[i], bt, "%d", i)
		}
	})

	t.Run("Shifted", func(t *testing.T) {
		w := &Buffer{}
		// write empty byte and reset it's byte index to 0.
		require.NoError(t, w.WriteByte(0x00))
		w.byteIndex = 0
		// assume that 3 '0' bits were already written.
		w.bitIndex = 3

		toWrite := []byte{0x3f, 0x12, 0x86}

		n, err := w.Write(toWrite)
		require.NoError(t, err)
		assert.Equal(t, 3, n)

		// 0x3f - 00111111
		// 00111111 << 3 = 11111000
		expected := byte(0xf8)
		assert.Equal(t, expected, w.data[0])

		// 0x12 - 00010010
		// 00111111 >> 5 = 00000001
		// 00010010 << 3 = 10010000
		// 				 | 10010101
		// 10010111 - 0x91
		expected = byte(0x91)
		assert.Equal(t, expected, w.data[1])

		// 0x86 - 10000110
		// 00010010 >> 5 = 	00000000
		// 10000110 << 3 = 	00110000
		// 				 |	00110000
		// 00110000 = 0x30
		expected = byte(0x30)
		assert.Equal(t, expected, w.data[2])
		assert.Len(t, w.Data(), 4)
	})

	t.Run("MSB", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			w := BufferedMSB()

			toWrite := []byte{0x3f, 0x12, 0x86}

			n, err := w.Write(toWrite)
			require.NoError(t, err)

			assert.Equal(t, 3, n)

			n, err = w.Write([]byte{0xff})
			require.NoError(t, err)

			assert.Equal(t, 1, n)

			expected := append(toWrite, 0xff)
			for i, bt := range w.Data() {
				assert.Equal(t, expected[i], bt, "%d", i)
			}
		})

		t.Run("Shifted", func(t *testing.T) {
			w := BufferedMSB()
			// write empty byte so the buffer data is initialized
			require.NoError(t, w.WriteByte(0x00))
			// reset it's byteindex
			w.byteIndex = 0
			// assume three '0' bits are already stored.
			w.bitIndex = 3

			toWrite := []byte{0x3f, 0x12, 0x86}

			n, err := w.Write(toWrite)
			require.NoError(t, err)
			assert.Equal(t, 3, n)

			// 0x3f - 00111111
			// 00111111 >> 3 = 00000111
			// 00000111 = 0x07
			expected := byte(0x07)
			assert.Equal(t, expected, w.data[0])

			// 0x12 - 00010010
			// 00111111 << 5 = 11100000
			// 00010010 >> 3 = 00000010
			// 				 | 11100010
			// 11100010 - 0xE2
			expected = byte(0xE2)
			assert.Equal(t, expected, w.data[1])

			// 0x86 - 10000110
			// 00010010 << 5 = 	01000000
			// 10000110 >> 3 = 	00010000
			// 				 |	01010000
			// 00110000 = 0x50
			expected = byte(0x50)
			assert.Equal(t, expected, w.data[2])

			// 0x86 - 10000110
			// 10000110 << 5 = 	11000000
			// 11000000 = 0xC0
			expected = byte(0xC0)
			assert.Equal(t, expected, w.data[3])
		})
	})
}

// TestBufferWriteBit tests the WriteBit method of the Writer.
func TestBufferWriteBit(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		buf := &Buffer{}
		// 10010011 11000111
		// 0x93 	0xC7
		bits := []int{1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1}
		for i := len(bits) - 1; i > -1; i-- {
			bit := bits[i]
			err := buf.WriteBit(bit)
			require.NoError(t, err)
		}

		assert.Equal(t, byte(0xC7), buf.data[0], "expected: %08b, is: %08b", 0xc7, buf.data[0])
		assert.Equal(t, byte(0x93), buf.data[1], "expected: %08b, is: %08b", 0x93, buf.data[1])
	})

	t.Run("BitShifted", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			buf := &Buffer{}
			// fill thee buffer with 3 bits
			for i := 0; i < 3; i++ {
				err := buf.WriteBit(int(0))
				require.NoError(t, err)
			}

			// bits 11101
			bits := []int{1, 1, 1, 0, 1}
			for i := len(bits) - 1; i > -1; i-- {
				bit := bits[i]
				err := buf.WriteBit(bit)
				require.NoError(t, err)
			}

			// should be 11101000 - 0xe8
			assert.Equal(t, byte(0xe8), buf.data[0])
		})
	})

	t.Run("ByteShifted", func(t *testing.T) {
		buf := &Buffer{}
		require.NoError(t, buf.WriteByte(0x00))

		// write 8 bits that should look like a byte 0xe3
		// 11100011 - 0xe3
		bits := []int{1, 1, 1, 0, 0, 0, 1, 1}
		for i := len(bits) - 1; i > -1; i-- {
			bit := bits[i]
			err := buf.WriteBit(bit)
			require.NoError(t, err)
		}
		assert.Equal(t, 2, len(buf.data))
		assert.Equal(t, byte(0xe3), buf.data[1])

		// there should be no error on writing additional byte.
		assert.NoError(t, buf.WriteByte(0x00))
	})

	t.Run("Finished", func(t *testing.T) {
		buf := &Buffer{}

		// write some bits to the first byte.
		firstBits := []int{1, 0, 1}
		for _, bit := range firstBits {
			err := buf.WriteBit(bit)
			require.NoError(t, err)
		}

		// finish this byte
		buf.FinishByte()
		secondBits := []int{1, 0, 1}

		// write some bits to the second byte.
		for _, bit := range secondBits {
			err := buf.WriteBit(bit)
			require.NoError(t, err)
		}

		if assert.Len(t, buf.data, 2) {
			// 00000101 - 0x05
			assert.Equal(t, byte(0x05), buf.Data()[0])
			assert.Equal(t, byte(0x05), buf.Data()[1])
		}
	})

	t.Run("Inverse", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			w := BufferedMSB()

			// 	10010111 10101100
			//	0x97	 0xac
			bits := []int{1, 0, 0, 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 0, 0}

			// write all the bits
			for _, bit := range bits {
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}

			expected := byte(0x97)
			assert.Equal(t, expected, w.data[0], "expected: %08b is: %08b", expected, w.data[0])
			expected = byte(0xac)
			assert.Equal(t, expected, w.data[1], "expected: %08b is: %08b", expected, w.data[1])
		})

		t.Run("ByteShifted", func(t *testing.T) {
			buf := BufferedMSB()
			err := buf.WriteByte(0x00)
			require.NoError(t, err)

			// 11100011 - 0xe3
			bits := []int{1, 1, 1, 0, 0, 0, 1, 1}
			for _, bit := range bits {
				err := buf.WriteBit(bit)
				require.NoError(t, err)
			}

			assert.Equal(t, byte(0xe3), buf.data[1], "expected: %08b, is: %08b", byte(0xe3), buf.data[1])
		})

		t.Run("BitShifted", func(t *testing.T) {
			w := BufferedMSB()

			// 0xE0 - 11100000
			err := w.WriteByte(0xE0)
			require.NoError(t, err)

			w.bitIndex = 5
			w.byteIndex = 0

			bits := []int{1, 0, 1, 0, 1}
			for _, bit := range bits {
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}

			// should be 11100101 01000000 ...
			//			 0xE5	  0x40
			assert.Equal(t, byte(0xE5), w.data[0], "expected: %08b, is: %08b", byte(0xE5), w.data[0])
			assert.Equal(t, byte(0x40), w.data[1], "expected: %08b, is: %08b", byte(0x40), w.data[1])
		})

		t.Run("Finished", func(t *testing.T) {
			buf := BufferedMSB()

			// write some bits to the first byte
			firstBits := []int{1, 0, 1}
			for _, bit := range firstBits {
				require.NoError(t, buf.WriteBit(bit))
			}
			// finish the byte
			buf.FinishByte()

			// write bits to the second byte.
			secondBits := []int{1, 0, 1}
			for _, bit := range secondBits {
				require.NoError(t, buf.WriteBit(bit))
			}

			if assert.Len(t, buf.Data(), 2) {
				// 10100000 - 0xa0
				assert.Equal(t, byte(0xa0), buf.Data()[0])
				assert.Equal(t, byte(0xa0), buf.Data()[1])
			}
		})
	})
}

// TestBufferWriteByte tests the WriteByte method of the Writer.
func TestBufferWriteByte(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		w := &Buffer{}

		input := []byte{0x4f, 0xff, 0x13, 0x2}

		for _, b := range input {
			err := w.WriteByte(b)
			require.NoError(t, err)
		}

		assert.Equal(t, 4, len(w.data))
		for i := 0; i < len(w.data); i++ {
			assert.Equal(t, input[i], w.data[i])
		}
	})

	t.Run("ByteShifted", func(t *testing.T) {
		w := &Buffer{}
		err := w.WriteByte(0xff)
		require.NoError(t, err)
		err = w.WriteByte(0x00)
		require.NoError(t, err)

		err = w.WriteByte(0x23)
		require.NoError(t, err)

		assert.Equal(t, byte(0xff), w.data[0])
		assert.Equal(t, byte(0x00), w.data[1])
		assert.Equal(t, byte(0x23), w.data[2])
	})

	t.Run("BitShifted", func(t *testing.T) {
		w := &Buffer{}

		// insert empty byte and reset the byte index to 0
		require.NoError(t, w.WriteByte(0x00))
		w.byteIndex = 0
		// assume there are 5 bits already set
		w.bitIndex = 5

		input := []byte{0x4f, 0xff, 0x13}

		for _, b := range input {
			err := w.WriteByte(b)
			require.NoError(t, err)
		}

		// 	00000000
		//     ^
		//
		// 0x4f - 01001111
		// 01001111 << 5 = 11100000
		//				 | 11100000
		// 11100000 - 0xE0
		expected := byte(0xe0)
		assert.Equal(t, expected, w.data[0], "expected: %08b is %08b", expected, w.data[0])

		// 0xff - 11111111
		// 01001111 >> 3 = 00001001
		// 11111111 << 5 = 11100000
		// 				 | 11101001
		// 11101001 = 0xe9
		expected = byte(0xe9)
		assert.Equal(t, expected, w.data[1], "expected %08b is %08b", expected, w.data[1])

		// 0x13 - 00010011
		// 11111111 >> 3 = 00011111
		// 00010011 << 5 = 01100000
		//			     | 01111111
		// 01111111 = 0x7F
		expected = byte(0x7F)
		assert.Equal(t, expected, w.data[2], "expected %08b is %08b", expected, w.data[2])

		// 00010011 >> 3 = 00000010
		// 00000010 = 0x02
		expected = byte(0x02)
		assert.Equal(t, expected, w.data[3])
	})

	t.Run("MSB", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			w := BufferedMSB()

			input := []byte{0x4f, 0xff, 0x13, 0x2}

			for _, b := range input {
				err := w.WriteByte(b)
				require.NoError(t, err)
			}

			assert.Len(t, w.data, 4)
			for i := 0; i < len(w.data); i++ {
				assert.Equal(t, input[i], w.data[i])
			}
		})

		t.Run("BitShifted", func(t *testing.T) {
			w := BufferedMSB()
			// write empty byte and reset it's byte index to 0.
			require.NoError(t, w.WriteByte(0x00))
			w.byteIndex = 0
			// assume that 5 empty bits are already written
			w.bitIndex = 5

			input := []byte{0x4f, 0xff, 0x13}

			for _, b := range input {
				err := w.WriteByte(b)
				require.NoError(t, err)
			}

			// 0x4f - 01001111
			// 01001111 >> 5 = 00000010
			// 				   0x02
			expected := byte(0x02)
			assert.Equal(t, expected, w.data[0], "expected: %08b is %08b", expected, w.data[0])

			// 0xff - 11111111
			// 01001111 << 3 = 01111000
			// 11111111 >> 5 = 00000111
			//				 | 01111111
			// 01111111 = 0x7F
			expected = byte(0x7F)
			assert.Equal(t, expected, w.data[1], "expected %08b is %08b", expected, w.data[1])

			// 0x13 - 00010011
			// 11111111 << 3 = 11111000
			// 00010011 >> 5 = 00000000
			// 				 | 11111000
			// 11111000 = 0xF8
			expected = byte(0xf8)
			assert.Equal(t, expected, w.data[2], "expected %08b is %08b", expected, w.data[2])

			// 00010011 << 3 = 10011000
			// 10011000 - 0x98
			expected = byte(0x98)
			assert.Equal(t, expected, w.data[3])
		})
	})
}

// TestWriteBits tests the WriteBits function.
func TestWriteBits(t *testing.T) {
	t.Run("NonMSB", func(t *testing.T) {
		b := &Buffer{}

		// having empty buffered MSB.
		n, err := b.WriteBits(0xb, 4)
		require.NoError(t, err)
		assert.Zero(t, n)

		assert.Len(t, b.data, 1)
		assert.Equal(t, byte(0xb), b.data[0])

		n, err = b.WriteBits(0xdf, 8)
		require.NoError(t, err)
		assert.Equal(t, 1, n)

		if assert.Len(t, b.data, 2) {
			assert.Equal(t, byte(0xd), b.data[1])
			assert.Equal(t, byte(0xfb), b.data[0])
		}
	})

	t.Run("MSB", func(t *testing.T) {
		b := BufferedMSB()

		n, err := b.WriteBits(0xf, 4)
		require.NoError(t, err)

		assert.Zero(t, n)

		// the output now should be
		// 11110000
		//     ^
		if assert.Len(t, b.data, 1) {
			assert.Equal(t, byte(0xf0), b.data[0], "%08b", b.data[0])
		}

		// write 10111 = 0x17, 5
		n, err = b.WriteBits(0x17, 5)
		require.NoError(t, err)

		// current output should be
		// 11111011 10000000
		//           ^
		if assert.Len(t, b.data, 2) {
			assert.Equal(t, byte(0xfb), b.data[0])
			assert.Equal(t, byte(0x80), b.data[1])
			assert.Equal(t, uint8(1), b.bitIndex)
		}
	})
}
