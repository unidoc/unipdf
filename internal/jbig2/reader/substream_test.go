/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package reader

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
)

// TestSubstream tests the SubstreamReader methods.
func TestSubstream(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}
	sampleData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	t.Run("Read", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			// get the base reader
			r := New(sampleData)
			var (
				offset uint64 = 3
				length uint64 = 2
			)

			substream, err := NewSubstreamReader(r, offset, length)
			require.NoError(t, err)

			var dataRead = make([]byte, 2)
			i, err := substream.Read(dataRead)
			if assert.NoError(t, err) {
				assert.Equal(t, 2, i)
				assert.Equal(t, sampleData[3], dataRead[0])
				assert.Equal(t, sampleData[4], dataRead[1])
			}
		})

		t.Run("ReadPart", func(t *testing.T) {
			// get the base reader
			r := New(sampleData)
			var (
				offset uint64 = 3
				length uint64 = 2
			)

			substream, err := NewSubstreamReader(r, offset, length)
			require.NoError(t, err)

			var dataRead = make([]byte, 3)
			i, err := substream.Read(dataRead)
			if assert.NoError(t, err) {
				assert.Equal(t, 2, i)
				assert.Equal(t, sampleData[3], dataRead[0])
				assert.Equal(t, sampleData[4], dataRead[1])
				assert.Equal(t, byte(0), dataRead[2])
			}
		})

		t.Run("EOF", func(t *testing.T) {
			// get the base reader
			r := New(sampleData)
			var (
				offset uint64 = 3
				length uint64 = 1
			)

			substream, err := NewSubstreamReader(r, offset, length)
			require.NoError(t, err)

			var dataRead = make([]byte, 3)

			b, err := substream.ReadByte()
			require.NoError(t, err)
			assert.Equal(t, sampleData[3], b)

			i, err := substream.Read(dataRead)
			if assert.Error(t, err) {
				assert.Zero(t, i)
			}
		})
	})

	t.Run("BaseOnReader", func(t *testing.T) {
		data := []byte{0x00, 3, 255, 0xcc, 0x1a, 0xbc, 0xde, 0x80, 0x01, 0x02, 0xf8, 0x08, 0xf0}
		br := New(data)
		r, err := NewSubstreamReader(br, 1, uint64(len(data)-1))
		require.NoError(t, err)

		b, err := r.ReadByte()
		require.NoError(t, err)
		assert.Equal(t, byte(3), b)
		assert.Equal(t, uint64(1), r.streamPos)

		bits, err := r.ReadBits(8)
		require.NoError(t, err)
		assert.Equal(t, uint64(255), bits)
		assert.Equal(t, uint64(2), r.streamPos)

		bits, err = r.ReadBits(4)
		require.NoError(t, err)
		assert.Equal(t, uint64(0xc), bits)
		assert.Equal(t, uint64(3), r.streamPos)

		bits, err = r.ReadBits(8)
		require.NoError(t, err)
		assert.Equal(t, uint64(0xc1), bits)
		assert.Equal(t, uint64(4), r.streamPos)

		bits, err = r.ReadBits(20)
		require.NoError(t, err)
		assert.Equal(t, uint64(0xabcde), bits)
		assert.Equal(t, uint64(6), r.streamPos)

		bl, err := r.ReadBool()
		require.NoError(t, err)
		assert.True(t, bl)
		assert.Equal(t, uint64(7), r.streamPos)

		bl, err = r.ReadBool()
		require.NoError(t, err)
		assert.False(t, bl)
		assert.Equal(t, uint64(7), r.streamPos)
		assert.Equal(t, byte(6), r.Align())

		s := make([]byte, 2)
		_, err = r.Read(s)
		assert.NoError(t, err, "Streampos: %d", r.streamPos)

		assert.True(t, bytes.Equal(s, []byte{0x01, 0x02}))

		bits, err = r.ReadBits(4)
		require.NoError(t, err)
		assert.Equal(t, uint64(0xf), bits)

		s = make([]byte, 2)
		_, err = r.Read(s)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(s, []byte{0x80, 0x8f}))
	})

	t.Run("SubstreamOfSubtream", func(t *testing.T) {
		data := []byte{0x00, 3, 255, 0xcc, 0x1a, 0xbc, 0xde, 0x80, 0x01, 0x02, 0xf8, 0x08, 0xf0}
		reader := New(data)

		sub1, err := NewSubstreamReader(reader, 1, 10)
		require.NoError(t, err)

		sub2, err := NewSubstreamReader(sub1, 2, 3)
		require.NoError(t, err)

		b, err := sub2.ReadByte()
		require.NoError(t, err)

		assert.Equal(t, data[3], b)
	})
}
