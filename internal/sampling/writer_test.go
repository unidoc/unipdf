/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package sampling

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/imageutil"
)

// TestSampleWriter tests SampleWriter interface.
func TestSampleWriter(t *testing.T) {
	t.Run("8BitGray", func(t *testing.T) {
		// Set the images as follows:
		//
		// 00000111 00111000 10101010
		// 11111111 00000000 11110000
		input := []byte{0x7, 0x38, 0xAA, 0xFF, 0x00, 0xF0}
		img := imageutil.NewImageBase(3, 2, 8, 1, nil, nil, nil)

		w := NewWriter(img)

		err := w.WriteSample(uint32(input[0]))
		require.NoError(t, err)

		samples := make([]uint32, 5)
		for i := 1; i < len(input); i++ {
			samples[i-1] = uint32(input[i])
		}

		err = w.WriteSamples(samples)
		require.NoError(t, err)

		assert.EqualValues(t, input, img.Data)

		err = w.WriteSample(0x0)
		assert.Error(t, err)
	})

	t.Run("16BitGray", func(t *testing.T) {
		// Set the images as follows:
		//
		// 00000111 00111000 10101010 00001111
		// 11111111 00000000 11110000 11001100
		input := []byte{0x7, 0x38, 0xaa, 0x0f, 0xff, 0x00, 0xf0, 0xcc}
		img := imageutil.NewImageBase(2, 2, 16, 1, nil, nil, nil)

		w := NewWriter(img)

		err := w.WriteSample((uint32(input[0]) << 8) | uint32(input[1]))
		require.NoError(t, err)

		samples := []uint32{0xaa0f, 0xff00, 0xf0cc}
		err = w.WriteSamples(samples)
		assert.NoError(t, err)

		assert.EqualValues(t, input, img.Data)

		err = w.WriteSample(0x0)
		assert.Error(t, err)
	})

	t.Run("4BitGray", func(t *testing.T) {
		t.Run("WithPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 00000111 00111000 10101010 11110000
			// 11111111 00000000 11110000 11010000
			input := []byte{0x7, 0x38, 0xaa, 0xf0, 0xff, 0x00, 0xf0, 0xd0}
			img := imageutil.NewImageBase(7, 2, 4, 1, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x0)
			require.NoError(t, err)

			err = w.WriteSample(0x7)
			require.NoError(t, err)

			samples := []uint32{0x3, 0x8, 0xa, 0xa, 0xf, 0xf, 0xf, 0x0, 0x0, 0xf, 0x0, 0xd}
			err = w.WriteSamples(samples)
			require.NoError(t, err)

			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(20)
			assert.Error(t, err)
		})

		t.Run("NoPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 00000111 00111000 10101010 11110000
			// 11111111 00000000 11110000 11010000
			input := []byte{0x7, 0x38, 0xaa, 0xf0, 0xff, 0x00, 0xf0, 0xd0}
			img := imageutil.NewImageBase(8, 2, 4, 1, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x0)
			require.NoError(t, err)

			err = w.WriteSample(0x7)
			require.NoError(t, err)

			samples := []uint32{0x3, 0x8, 0xa, 0xa, 0xf, 0x0, 0xf, 0xf, 0x0, 0x0, 0xf, 0x0, 0xd, 0x0}
			err = w.WriteSamples(samples)
			require.NoError(t, err)

			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(2)
			assert.Error(t, err)
		})
	})

	t.Run("2BitGray", func(t *testing.T) {
		t.Run("WithPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 10000111 00111000
			// 11111111 11110000
			input := []byte{0x87, 0x38, 0xff, 0xf0}
			img := imageutil.NewImageBase(7, 2, 2, 1, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x2)
			require.NoError(t, err)

			samples := []uint32{0x0, 0x1, 0x3, 0x0, 0x3, 0x2, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x0}
			err = w.WriteSamples(samples)
			require.NoError(t, err)

			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(0x0)
			assert.Error(t, err)
		})

		t.Run("NoPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 10000111 00111000
			// 11111111 11110000
			input := []byte{0x87, 0x38, 0xff, 0xf0}
			img := imageutil.NewImageBase(8, 2, 2, 1, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x2)
			require.NoError(t, err)

			samples := []uint32{0x0, 0x1, 0x3, 0x0, 0x3, 0x2, 0x0, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x0, 0x0}
			err = w.WriteSamples(samples)
			require.NoError(t, err)

			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(0x0)
			assert.Error(t, err)
		})
	})

	t.Run("Monochrome", func(t *testing.T) {
		t.Run("WithPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 10000111 00111000
			// 11111111 11110000
			input := []byte{0x87, 0x38, 0xff, 0xf0}
			img := imageutil.NewImageBase(13, 2, 1, 1, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x1)
			require.NoError(t, err)

			samples := []uint32{
				0, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0,
			}
			err = w.WriteSamples(samples)
			require.NoError(t, err)

			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(0)
			require.Error(t, err)
		})

		t.Run("NoPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 10000111 00111000
			// 11111111 11110000
			input := []byte{0x87, 0x38, 0xff, 0xf0}
			img := imageutil.NewImageBase(16, 2, 1, 1, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x1)
			require.NoError(t, err)

			samples := []uint32{
				0, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 0,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0,
			}
			err = w.WriteSamples(samples)
			require.NoError(t, err)

			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(1)
			assert.Error(t, err)
		})
	})

	t.Run("4BitRGB", func(t *testing.T) {
		t.Run("WithPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 10000111 00111000 10101010 00001111 11000000
			// 11111111 00000000 11110000 11001100 10100000
			input := []byte{
				0x87, 0x38, 0xaa, 0x0f, 0xc0,
				0xff, 0x00, 0xf0, 0xcc, 0xa0,
			}
			img := imageutil.NewImageBase(3, 2, 4, 3, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x8)
			require.NoError(t, err)

			samples := []uint32{
				0x7, 0x3, 0x8, 0xa, 0xa, 0x0, 0xf, 0xc,
				0xf, 0xf, 0x0, 0x0, 0xf, 0x0, 0xc, 0xc, 0xa,
			}
			err = w.WriteSamples(samples)
			require.NoError(t, err)

			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(1)
			assert.Error(t, err)
		})

		t.Run("NoPadding", func(t *testing.T) {
			// Set the images as follows:
			//
			// 10000111 00111000 10101010 00001111 11000000 00110011
			// 11111111 00000000 11110000 11001100 10100000 11001100
			input := []byte{
				0x87, 0x38, 0xaa, 0x0f, 0xc0, 0x33,
				0xff, 0x00, 0xf0, 0xcc, 0xa0, 0xcc,
			}
			img := imageutil.NewImageBase(4, 2, 4, 3, nil, nil, nil)

			w := NewWriter(img)

			err := w.WriteSample(0x8)
			require.NoError(t, err)

			samples := []uint32{
				0x7, 0x3, 0x8, 0xa, 0xa, 0x0, 0xf, 0xc, 0x0, 0x3, 0x3,
				0xf, 0xf, 0x0, 0x0, 0xf, 0x0, 0xc, 0xc, 0xa, 0x0, 0xc, 0xc,
			}
			err = w.WriteSamples(samples)
			require.NoError(t, err)
			assert.EqualValues(t, input, img.Data)

			err = w.WriteSample(1)
			assert.Error(t, err)
		})
	})

	t.Run("8BitRGB", func(t *testing.T) {
		// Set the images as follows:
		//
		// 10000111 00111000 10101010 00001111 11000000 00110011
		// 11111111 00000000 11110000 11001100 10100000 11001100
		input := []byte{
			0x87, 0x38, 0xaa, 0x0f, 0xc0, 0x33,
			0xff, 0x00, 0xf0, 0xcc, 0xa0, 0xcc,
		}
		img := imageutil.NewImageBase(2, 2, 8, 3, nil, nil, nil)

		w := NewWriter(img)

		err := w.WriteSample(0x87)
		require.NoError(t, err)

		samples := []uint32{
			0x38, 0xaa, 0x0f, 0xc0, 0x33,
			0xff, 0x00, 0xf0, 0xcc, 0xa0, 0xcc,
		}
		err = w.WriteSamples(samples)
		require.NoError(t, err)

		assert.EqualValues(t, input, img.Data)

		err = w.WriteSample(1)
		assert.Error(t, err)
	})

	t.Run("16BitRGB", func(t *testing.T) {
		// Set the images as follows:
		//
		// 10000111 00111000 10101010 00001111 11000000 00110011 00001111 11000000 00110011 10100000 11001100 10000001
		// 11111111 00000000 11110000 11001100 10100000 11001100 10000111 00111000 00001111 11000000 10101010 11001100
		input := []byte{
			0x87, 0x38, 0xaa, 0x0f, 0xc0, 0x33, 0x0f, 0xc0, 0x33, 0xa0, 0xcc, 0x81,
			0xff, 0x00, 0xf0, 0xcc, 0xa0, 0xcc, 0x87, 0x38, 0x0f, 0xc0, 0xaa, 0xcc,
		}
		img := imageutil.NewImageBase(2, 2, 16, 3, nil, nil, nil)

		w := NewWriter(img)

		err := w.WriteSample(0x8738)
		require.NoError(t, err)

		samples := []uint32{
			0xaa0f, 0xc033, 0x0fc0, 0x33a0, 0xcc81,
			0xff00, 0xf0cc, 0xa0cc, 0x8738, 0x0fc0, 0xaacc,
		}
		err = w.WriteSamples(samples)
		require.NoError(t, err)

		assert.EqualValues(t, input, img.Data)

		err = w.WriteSample(1)
		assert.Error(t, err)
	})

	t.Run("8BitCMYK", func(t *testing.T) {
		// Set the images as follows:
		//
		// 10000111 00111000 10101010 00001111 11000000 00110011 00001111 11000000 00110011 10100000 11001100 10000001
		// 11111111 00000000 11110000 11001100 10100000 11001100 10000111 00111000 00001111 11000000 10101010 11001100
		input := []byte{
			0x87, 0x38, 0xaa, 0x0f, 0xc0, 0x33, 0x0f, 0xc0, 0x33, 0xa0, 0xcc, 0x81,
			0xff, 0x00, 0xf0, 0xcc, 0xa0, 0xcc, 0x87, 0x38, 0x0f, 0xc0, 0xaa, 0xcc,
		}
		img := imageutil.NewImageBase(3, 2, 8, 4, nil, nil, nil)

		w := NewWriter(img)

		err := w.WriteSample(0x87)
		require.NoError(t, err)

		samples := []uint32{
			0x38, 0xaa, 0x0f, 0xc0, 0x33, 0x0f, 0xc0, 0x33, 0xa0, 0xcc, 0x81,
			0xff, 0x00, 0xf0, 0xcc, 0xa0, 0xcc, 0x87, 0x38, 0x0f, 0xc0, 0xaa, 0xcc,
		}
		err = w.WriteSamples(samples)
		require.NoError(t, err)
		assert.EqualValues(t, input, img.Data)

		err = w.WriteSample(1)
		assert.Error(t, err)
	})
}
