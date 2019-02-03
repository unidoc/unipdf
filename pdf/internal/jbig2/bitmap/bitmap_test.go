package bitmap

import (
	"github.com/stretchr/testify/assert"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"testing"
)

func TestBitmapCombine(t *testing.T) {

	t.Run("Or", func(t *testing.T) {
		t.Run("SameSize", func(t *testing.T) {
			t.Run("XY0", func(t *testing.T) {
				// the root is
				// 0 1 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				root := New(5, 5, container.New())
				root.SetPixel(1, 0, 1)

				// src is
				// 0 0 0 0 0
				// 0 1 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				src := New(5, 5, container.New())
				src.SetPixel(1, 1, 1)

				err := root.Combine(src, 0, 0, int64(CmbOpOr))
				if assert.NoError(t, err) {
					// The result should be a union of both bitmaps
					// 0 1 0 0 0
					// 0 1 0 0 0
					// 0 0 0 0 0
					// 0 0 0 0 0
					// 0 0 0 0 0

					for i := 0; i < 25; i++ {
						row, col := i/5, i%5
						pix := root.GetPixel(col, row)
						if i == 1 || i == 6 {
							assert.True(t, pix)
						} else {
							assert.False(t, pix)
						}
					}

				}
			})
			t.Run("XYGreaterThan0", func(t *testing.T) {
				// the root is
				// 0 1 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				root := New(5, 5, container.New())
				root.SetPixel(1, 0, 1)

				// src is
				// 1 0 0 0 0
				// 0 1 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				// 0 0 0 0 0
				src := New(5, 5, container.New())
				src.SetPixel(0, 0, 1)
				src.SetPixel(1, 1, 1)

				if assert.NoError(t, root.Combine(src, 0, 1, int64(CmbOpOr))) {

					// The result should paste the value of src into root
					// starting at x = 0, y = 1
					// 0 1 0 0 0
					// 1 0 0 0 0
					// 0 1 0 0 0
					// 0 0 0 0 0
					// 0 0 0 0 0
					for i := 0; i < 25; i++ {
						pix := root.GetPixel(i%5, i/5)
						switch i {
						case 1, 5, 11:
							assert.True(t, pix, "i: %d", i)
						default:
							assert.False(t, pix, "i: %d", i)
						}

					}
				}
			})
		})
		t.Run("SrcSmaller", func(t *testing.T) {
			// the root is
			// 0 1 0 0 0
			// 0 0 0 0 0
			// 0 0 1 0 0
			// 0 0 0 0 0
			// 0 0 0 0 0
			root := New(5, 5, container.New())
			root.SetPixel(1, 0, 1)
			root.SetPixel(2, 2, 1)

			// src is
			// 1 1 1
			// 1 0 1
			// 1 1 1
			src := New(3, 3, container.New())
			src.Data.SetAll(true)
			src.SetPixel(1, 1, 0)

			if assert.NoError(t, root.Combine(src, 1, 1, int64(CmbOpOr))) {
				// The result should union with 'src' bitmap at x: 1, y: 1
				// 0 1 0 0 0
				// 0 1 1 1 0
				// 0 1 1 1 0
				// 0 1 1 1 0
				// 0 0 0 0 0

				for i := 0; i < 25; i++ {
					col, row := i%5, i/5
					pix := root.GetPixel(col, row)
					switch i {
					case 1, 6, 7, 8, 11, 12, 13, 16, 17, 18:
						assert.True(t, pix)
					default:
						assert.False(t, pix)
					}
				}

			}
		})

	})

	t.Run("And", func(t *testing.T) {

	})

	t.Run("Xor", func(t *testing.T) {

	})

	t.Run("Xnor", func(t *testing.T) {

	})
}
