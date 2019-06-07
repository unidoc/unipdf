/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package reader

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReader tests the reader Read methods.
func TestReader(t *testing.T) {
	data := []byte{3, 255, 0xcc, 0x1a, 0xbc, 0xde, 0x80, 0x01, 0x02, 0xf8, 0x08, 0xf0}

	r := New(data)
	b, err := r.ReadByte()
	require.NoError(t, err)
	assert.Equal(t, byte(3), b)

	bits, err := r.ReadBits(8)
	require.NoError(t, err)
	assert.Equal(t, uint64(255), bits)

	bits, err = r.ReadBits(4)
	require.NoError(t, err)
	assert.Equal(t, uint64(0xc), bits)

	bits, err = r.ReadBits(8)
	require.NoError(t, err)
	assert.Equal(t, uint64(0xc1), bits)

	bits, err = r.ReadBits(20)
	require.NoError(t, err)
	assert.Equal(t, uint64(0xabcde), bits)

	bl, err := r.ReadBool()
	require.NoError(t, err)
	assert.True(t, bl)

	bl, err = r.ReadBool()
	require.NoError(t, err)
	assert.False(t, bl)

	assert.Equal(t, byte(6), r.Align())

	s := make([]byte, 2)
	_, err = r.Read(s)
	require.NoError(t, err)

	assert.True(t, bytes.Equal(s, []byte{0x01, 0x02}))

	bits, err = r.ReadBits(4)
	require.NoError(t, err)
	assert.Equal(t, uint64(0xf), bits)

	s = make([]byte, 2)
	_, err = r.Read(s)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(s, []byte{0x80, 0x8f}))
}

// TestSeeker test the Reader Seek methods.
func TestSeeker(t *testing.T) {
	data := []byte{3, 255, 0xcc, 0x1a, 0xbc, 0xde, 0x80, 0x01, 0x02, 0xf8, 0x08, 0xf0}

	t.Run("SeekStart", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			r := New(data)
			r.Seek(3, io.SeekStart)
			b, err := r.ReadByte()
			assert.Equal(t, nil, err)
			assert.Equal(t, data[3], b)
		})
		t.Run("Invalid", func(t *testing.T) {
			r := New(data)
			_, err := r.Seek(int64(len(data)+1), io.SeekStart)
			assert.Equal(t, nil, err)

			_, err = r.ReadByte()
			assert.Equal(t, io.EOF, err)
		})
	})
	t.Run("SeekCurrent", func(t *testing.T) {
		r := New(data)

		_, err := r.ReadByte()
		assert.Equal(t, nil, err)

		_, err = r.Seek(int64(2), io.SeekCurrent)
		assert.Equal(t, nil, err)

		b, err := r.ReadByte()
		assert.Equal(t, nil, err)

		assert.Equal(t, data[2+1], b)
	})
	t.Run("SeekEnd", func(t *testing.T) {
		r := New(data)

		_, err := r.Seek(int64(-1), io.SeekEnd)
		assert.Equal(t, nil, err)

		b, err := r.ReadByte()
		assert.Equal(t, nil, err)
		assert.Equal(t, data[len(data)-1], b)
	})

}
