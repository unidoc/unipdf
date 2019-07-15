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

// TestWriteBit tests the WriteBit method of the Writer.
func TestWriteBit(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		// 10010011 11000111
		// 0x93 	0xC7
		bits := []int{1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1}
		for i := len(bits) - 1; i > -1; i-- {
			bit := bits[i]
			err := w.WriteBit(bit)
			require.NoError(t, err)
		}

		assert.Equal(t, byte(0xC7), data[0], "expected: %08b, is: %08b", 0xc7, data[0])
		assert.Equal(t, byte(0x93), data[1], "expected: %08b, is: %08b", 0x93, data[1])
	})

	t.Run("BitShifted", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			data := make([]byte, 4)
			w := New(data)
			w.bitIndex = 3

			// bits 11101
			bits := []int{1, 1, 1, 0, 1}
			for i := len(bits) - 1; i > -1; i-- {
				bit := bits[i]
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}

			// should be 11101000 - 0xe8
			assert.Equal(t, byte(0xe8), data[0])
		})

		t.Run("PreFilled", func(t *testing.T) {
			data := make([]byte, 4)
			w := New(data)

			w.bitIndex = 3
			// 0x01 - 00000001
			//             ^
			data[0] |= 0x01

			// bits 0,1,0
			bits := []int{0, 1, 0}
			for i := len(bits) - 1; i > -1; i-- {
				bit := bits[i]
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}

			// data should be 00010001
			assert.Equal(t, byte(0x11), data[0])
		})
	})

	t.Run("ByteShifted", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)
		w.byteIndex = 2

		// 11100011 - 0xe3
		bits := []int{1, 1, 1, 0, 0, 0, 1, 1}
		for i := len(bits) - 1; i > -1; i-- {
			bit := bits[i]
			err := w.WriteBit(bit)
			require.NoError(t, err)
		}

		assert.Equal(t, byte(0xe3), data[2])
	})

	t.Run("Overflow", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		w.byteIndex = 3

		err := w.WriteBit(1)
		require.NoError(t, err)

		w.bitIndex = 7

		// should flip the byte index
		err = w.WriteBit(0)
		require.NoError(t, err)

		err = w.WriteBit(1)
		require.Error(t, err)
	})

	t.Run("Inverse", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)

			// 	10010111 10101100
			//	0x97	 0xac
			bits := []int{1, 0, 0, 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 0, 0}

			// write all the bits
			for _, bit := range bits {
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}

			expected := byte(0x97)
			assert.Equal(t, expected, data[0], "expected: %08b is: %08b", expected, data[0])
			expected = byte(0xac)
			assert.Equal(t, expected, data[1], "expected: %08b is: %08b", expected, data[1])
		})

		t.Run("ByteShifted", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)
			w.byteIndex = 2

			// 11100011 - 0xe3
			bits := []int{1, 1, 1, 0, 0, 0, 1, 1}
			for _, bit := range bits {
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}

			assert.Equal(t, byte(0xe3), data[2], "expected: %08b, is: %08b", byte(0xe3), data[2])
		})

		t.Run("BitShifted", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)
			w.bitIndex = 5

			// 0xE0 - 11100000
			data[0] = 0xE0

			bits := []int{1, 0, 1, 0, 1}
			for _, bit := range bits {
				err := w.WriteBit(bit)
				require.NoError(t, err)
			}

			// should be 11100101 01000000 ...
			//			 0xE5	  0x40
			assert.Equal(t, byte(0xE5), data[0], "expected: %08b, is: %08b", byte(0xE5), data[0])
			assert.Equal(t, byte(0x40), data[1], "expected: %08b, is: %08b", byte(0x40), data[1])
		})

		t.Run("Overflow", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)

			w.byteIndex = 3

			err := w.WriteBit(1)
			require.NoError(t, err)

			w.bitIndex = 7

			// should flip the byte index
			err = w.WriteBit(0)
			require.NoError(t, err)

			err = w.WriteBit(1)
			require.Error(t, err)
		})
	})
}

// TestWriteByte tests the WriteByte method of the Writer.
func TestWriteByte(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		input := []byte{0x4f, 0xff, 0x13, 0x2}

		for _, b := range input {
			err := w.WriteByte(b)
			require.NoError(t, err)
		}

		for i := 0; i < len(data); i++ {
			assert.Equal(t, input[i], data[i])
		}
	})

	t.Run("ByteShifted", func(t *testing.T) {
		data := make([]byte, 4)
		data[0] = 0xff

		w := New(data)
		w.byteIndex = 2

		err := w.WriteByte(0x23)
		require.NoError(t, err)

		assert.Equal(t, byte(0xff), data[0])
		assert.Equal(t, byte(0x00), data[1])
		assert.Equal(t, byte(0x23), data[2])
	})

	t.Run("BitShifted", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		w.bitIndex = 5

		input := []byte{0x4f, 0xff, 0x13}

		for _, b := range input {
			err := w.writeByte(b)
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
		assert.Equal(t, expected, data[0], "expected: %08b is %08b", expected, data[0])

		// 0xff - 11111111
		// 01001111 >> 3 = 00001001
		// 11111111 << 5 = 11100000
		// 				 | 11101001
		// 11101001 = 0xe9
		expected = byte(0xe9)
		assert.Equal(t, expected, data[1], "expected %08b is %08b", expected, data[1])

		// 0x13 - 00010011
		// 11111111 >> 3 = 00011111
		// 00010011 << 5 = 01100000
		//			     | 01111111
		// 01111111 = 0x7F
		expected = byte(0x7F)
		assert.Equal(t, expected, data[2], "expected %08b is %08b", expected, data[2])

		// 00010011 >> 3 = 00000010
		// 00000010 = 0x02
		expected = byte(0x02)
		assert.Equal(t, expected, data[3])
	})

	t.Run("Overflow", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		input := []byte{0x4f, 0xff, 0x13, 0xff, 0x12}

		for i, b := range input {
			err := w.writeByte(b)
			switch i {
			case 4:
				require.Error(t, err)
			default:
				require.NoError(t, err)
			}
		}
	})

	t.Run("BitOverflow", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		w.bitIndex = 5

		input := []byte{0x4f, 0xff, 0x13, 0xff}

		for i, b := range input {
			err := w.WriteByte(b)
			switch i {
			case 3:
				require.Error(t, err)
			default:
				require.NoError(t, err)
			}
		}
	})

	t.Run("MSB", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)

			input := []byte{0x4f, 0xff, 0x13, 0x2}

			for _, b := range input {
				err := w.WriteByte(b)
				require.NoError(t, err)
			}

			for i := 0; i < len(data); i++ {
				assert.Equal(t, input[i], data[i])
			}
		})

		t.Run("ByteShifted", func(t *testing.T) {
			data := make([]byte, 4)
			data[0] = 0xff

			w := NewMSB(data)
			w.byteIndex = 2

			err := w.WriteByte(0x23)
			require.NoError(t, err)

			assert.Equal(t, byte(0xff), data[0])
			assert.Equal(t, byte(0x00), data[1])
			assert.Equal(t, byte(0x23), data[2])
		})

		t.Run("BitShifted", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)

			w.bitIndex = 5

			input := []byte{0x4f, 0xff, 0x13}

			for _, b := range input {
				err := w.writeByte(b)
				require.NoError(t, err)
			}

			// 0x4f - 01001111
			// 01001111 >> 5 = 00000010
			// 				   0x02
			expected := byte(0x02)
			assert.Equal(t, expected, data[0], "expected: %08b is %08b", expected, data[0])

			// 0xff - 11111111
			// 01001111 << 3 = 01111000
			// 11111111 >> 5 = 00000111
			//				 | 01111111
			// 01111111 = 0x7F
			expected = byte(0x7F)
			assert.Equal(t, expected, data[1], "expected %08b is %08b", expected, data[1])

			// 0x13 - 00010011
			// 11111111 << 3 = 11111000
			// 00010011 >> 5 = 00000000
			// 				 | 11111000
			// 11111000 = 0xF8
			expected = byte(0xf8)
			assert.Equal(t, expected, data[2], "expected %08b is %08b", expected, data[2])

			// 00010011 << 3 = 10011000
			// 10011000 - 0x98
			expected = byte(0x98)
			assert.Equal(t, expected, data[3])
		})
	})
}

// TestWrite tests the Write method of the Writer.
func TestWrite(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		toWrite := []byte{0x3f, 0x12, 0x86}

		n, err := w.Write(toWrite)
		require.NoError(t, err)

		assert.Equal(t, 3, n)

		n, err = w.Write([]byte{0xff})
		require.NoError(t, err)

		assert.Equal(t, 1, n)
	})

	t.Run("Shifted", func(t *testing.T) {
		data := make([]byte, 4)

		w := New(data)
		w.bitIndex = 3

		toWrite := []byte{0x3f, 0x12, 0x86}

		n, err := w.Write(toWrite)
		require.NoError(t, err)
		assert.Equal(t, 3, n)

		// 0x3f - 00111111
		// 00111111 << 3 = 11111000
		expected := byte(0xf8)
		assert.Equal(t, expected, data[0])

		// 0x12 - 00010010
		// 00111111 >> 5 = 00000001
		// 00010010 << 3 = 10010000
		// 				 | 10010101
		// 10010111 - 0x91
		expected = byte(0x91)
		assert.Equal(t, expected, data[1])

		// 0x86 - 10000110
		// 00010010 >> 5 = 	00000000
		// 10000110 << 3 = 	00110000
		// 				 |	00110000
		// 00110000 = 0x30
		expected = byte(0x30)
		assert.Equal(t, expected, data[2])
	})

	t.Run("OutOfRange", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)

		// 5 byte values into 4 byte size data.
		overflow := []byte{0x21, 0x66, 0x14, 0xff, 0x12}

		_, err := w.Write(overflow)
		require.Error(t, err)
	})

	t.Run("BitIndexOverflow", func(t *testing.T) {
		data := make([]byte, 4)
		w := New(data)
		// set bit index greater than 0 - some bits are already written.
		w.bitIndex = 3

		// even thought there are 4 bytes to write and the
		// size of the writer data is 4 an error should be thrown
		// due to the bitIndex != 0
		overflow := []byte{0x21, 0x66, 0x14, 0xff}

		_, err := w.Write(overflow)
		require.Error(t, err)
	})

	t.Run("MSB", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)

			toWrite := []byte{0x3f, 0x12, 0x86}

			n, err := w.Write(toWrite)
			require.NoError(t, err)

			assert.Equal(t, 3, n)

			n, err = w.Write([]byte{0xff})
			require.NoError(t, err)

			assert.Equal(t, 1, n)
		})

		t.Run("Overflow", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)

			toWrite := []byte{0x3f, 0x12, 0x86}

			n, err := w.Write(toWrite)
			require.NoError(t, err)

			assert.Equal(t, 3, n)

			n, err = w.Write([]byte{0xff})
			require.NoError(t, err)

			assert.Equal(t, 1, n)
		})

		t.Run("Shifted", func(t *testing.T) {
			data := make([]byte, 4)

			w := NewMSB(data)
			w.bitIndex = 3

			toWrite := []byte{0x3f, 0x12, 0x86}

			n, err := w.Write(toWrite)
			require.NoError(t, err)
			assert.Equal(t, 3, n)

			// 0x3f - 00111111
			// 00111111 >> 3 = 00000111
			// 00000111 = 0x07
			expected := byte(0x07)
			assert.Equal(t, expected, data[0])

			// 0x12 - 00010010
			// 00111111 << 5 = 11100000
			// 00010010 >> 3 = 00000010
			// 				 | 11100010
			// 11100010 - 0xE2
			expected = byte(0xE2)
			assert.Equal(t, expected, data[1])

			// 0x86 - 10000110
			// 00010010 << 5 = 	01000000
			// 10000110 >> 3 = 	00010000
			// 				 |	01010000
			// 00110000 = 0x50
			expected = byte(0x50)
			assert.Equal(t, expected, data[2])

			// 0x86 - 10000110
			// 10000110 << 5 = 	11000000
			// 11000000 = 0xC0
			expected = byte(0xC0)
			assert.Equal(t, expected, data[3])
		})

		t.Run("BitIndexOverflow", func(t *testing.T) {
			data := make([]byte, 4)
			w := NewMSB(data)
			// set bit index greater than 0 - some bits are already written.
			w.bitIndex = 3

			// even thought there are 4 bytes to write and the
			// size of the writer data is 4 an error should be thrown
			// due to the bitIndex != 0
			overflow := []byte{0x21, 0x66, 0x14, 0xff}

			_, err := w.Write(overflow)
			require.Error(t, err)
		})
	})
}
