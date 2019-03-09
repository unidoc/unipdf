package reader

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	"testing"
)

func TestSubstream(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	var sampleData []byte = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

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

			var dataRead []byte = make([]byte, 2)
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

			var dataRead []byte = make([]byte, 3)
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

			var dataRead []byte = make([]byte, 3)

			b, err := substream.ReadByte()
			require.NoError(t, err)
			assert.Equal(t, sampleData[3], b)

			i, err := substream.Read(dataRead)
			if assert.Error(t, err) {
				assert.Zero(t, i)
			}

		})

	})

}
