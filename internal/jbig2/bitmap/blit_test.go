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

// TestBlit tests the Blit function for the multiple CombinationOperators.
func TestBlit(t *testing.T) {
	t.Run("Unshifted", func(t *testing.T) {
		t.Run("SourceEqualSize", func(t *testing.T) {
			// Test Blit when the bytes are at their 0th position.
			t.Run("OR", func(t *testing.T) {
				dst := New(25, 25)
				src := New(25, 25)

				require.NoError(t, src.SetPixel(2, 2, 1))
				require.NoError(t, dst.SetPixel(3, 3, 1))
				op := CmbOpOr

				require.NoError(t, Blit(src, dst, 0, 0, op))

				// both value should be true at (2,2) and (3,3)
				assert.True(t, dst.GetPixel(2, 2))
				assert.True(t, dst.GetPixel(3, 3))
			})

			t.Run("AND", func(t *testing.T) {
				dst := New(25, 25)
				src := New(25, 25)

				require.NoError(t, src.SetPixel(3, 3, 1))
				require.NoError(t, src.SetPixel(2, 2, 1))
				require.NoError(t, dst.SetPixel(3, 3, 1))
				require.NoError(t, dst.SetPixel(4, 4, 1))
				op := CmbOpAnd

				require.NoError(t, Blit(src, dst, 0, 0, op))

				// both value should be true at (2,2) and (3,3)
				assert.True(t, dst.GetPixel(3, 3))
				assert.False(t, dst.GetPixel(2, 2))
				assert.False(t, dst.GetPixel(4, 4))
			})

			t.Run("XOR", func(t *testing.T) {
				dst := New(25, 25)
				src := New(25, 25)

				require.NoError(t, src.SetPixel(3, 3, 1))
				require.NoError(t, src.SetPixel(2, 2, 1))
				require.NoError(t, dst.SetPixel(3, 3, 1))
				require.NoError(t, dst.SetPixel(4, 4, 1))
				op := CmbOpXor

				require.NoError(t, Blit(src, dst, 0, 0, op))

				// both value should be true at (2,2) and (3,3)
				assert.False(t, dst.GetPixel(3, 3))
				assert.True(t, dst.GetPixel(2, 2))
				assert.True(t, dst.GetPixel(4, 4))
			})

			t.Run("XNOR", func(t *testing.T) {
				dst := New(25, 25)
				src := New(25, 25)

				require.NoError(t, src.SetPixel(3, 3, 1))
				require.NoError(t, src.SetPixel(2, 2, 1))
				require.NoError(t, dst.SetPixel(3, 3, 1))
				require.NoError(t, dst.SetPixel(4, 4, 1))
				op := CmbOpXNor

				require.NoError(t, Blit(src, dst, 0, 0, op))

				// both value should be true at (2,2) and (3,3)
				assert.True(t, dst.GetPixel(3, 3))
				assert.False(t, dst.GetPixel(2, 2))
				assert.False(t, dst.GetPixel(4, 4))
				assert.True(t, dst.GetPixel(5, 5))
			})

			t.Run("Replace", func(t *testing.T) {
				dst := New(25, 25)
				src := New(25, 25)

				require.NoError(t, src.SetPixel(3, 3, 1))
				require.NoError(t, src.SetPixel(2, 2, 1))
				require.NoError(t, dst.SetPixel(3, 3, 1))
				require.NoError(t, dst.SetPixel(4, 4, 1))
				op := CmbOpReplace

				require.NoError(t, Blit(src, dst, 0, 0, op))

				// both value should be true at (2,2) and (3,3)
				assert.True(t, dst.GetPixel(3, 3))
				assert.True(t, dst.GetPixel(2, 2))
				assert.False(t, dst.GetPixel(4, 4))
				assert.False(t, dst.GetPixel(5, 5))
			})

			t.Run("NOT", func(t *testing.T) {
				dst := New(25, 25)
				src := New(25, 25)

				require.NoError(t, src.SetPixel(3, 3, 1))
				require.NoError(t, src.SetPixel(2, 2, 1))
				require.NoError(t, dst.SetPixel(3, 3, 1))
				require.NoError(t, dst.SetPixel(4, 4, 1))
				op := CmbOpNot

				require.NoError(t, Blit(src, dst, 0, 0, op))

				// both value should be true at (2,2) and (3,3)
				assert.False(t, dst.GetPixel(3, 3))
				assert.False(t, dst.GetPixel(2, 2))
				assert.True(t, dst.GetPixel(4, 4))
				assert.True(t, dst.GetPixel(5, 5))
			})
		})

		t.Run("SourceSmallerSize", func(t *testing.T) {
			// Test Blit when the bytes are at their 0th position.
			dst := New(25, 25)
			src := New(5, 5)

			require.NoError(t, src.SetPixel(2, 2, 1))
			require.NoError(t, dst.SetPixel(3, 3, 1))
			op := CmbOpOr

			require.NoError(t, Blit(src, dst, 0, 0, op))

			// both value should be true at (2,2) and (3,3).
			assert.True(t, dst.GetPixel(2, 2))
			assert.True(t, dst.GetPixel(3, 3))
		})
	})

	t.Run("Shifted", func(t *testing.T) {
		t.Run("SourceEqualSize", func(t *testing.T) {
			// Test Blit when the bytes are at their 0th position.
			dst := New(25, 25)
			src := New(25, 25)

			require.NoError(t, src.SetPixel(2, 2, 1))
			require.NoError(t, dst.SetPixel(5, 5, 1))
			op := CmbOpOr

			require.NoError(t, Blit(src, dst, 1, 1, op))

			// both value should be true at (2,2) and (3,3)
			assert.True(t, dst.GetPixel(3, 3))
			assert.True(t, dst.GetPixel(5, 5), dst.String())
		})
	})

	t.Run("SpecialShifted", func(t *testing.T) {
		// Test Blit when the bytes are at their 0th position.
		dst := New(25, 25)
		src := New(5, 5)

		require.NoError(t, src.SetPixel(2, 2, 1))
		require.NoError(t, dst.SetPixel(5, 5, 1))
		op := CmbOpOr

		require.NoError(t, Blit(src, dst, 1, 1, op))

		// both value should be true at (2,2) and (3,3)
		assert.True(t, dst.GetPixel(3, 3))
		assert.True(t, dst.GetPixel(5, 5), dst.String())
	})
}
