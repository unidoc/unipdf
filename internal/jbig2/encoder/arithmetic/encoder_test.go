/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/bitwise"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/arithmetic"
)

// TestEncoder tests the encoder using the standard H.2 test sequence.
func TestEncoder(t *testing.T) {
	testSequence := []byte{0x00, 0x02, 0x00, 0x51, 0x00, 0x00, 0x00, 0xC0, 0x03, 0x52, 0x87, 0x2A, 0xAA, 0xAA, 0xAA, 0xAA, 0x82, 0xC0, 0x20, 0x00, 0xFC, 0xD7, 0x9E, 0xF6, 0xBF, 0x7F, 0xED, 0x90, 0x4F, 0x46, 0xA3, 0xBF}
	require.Len(t, testSequence, 32)

	shouldBe := []byte{0x84, 0xC7, 0x3B, 0xFC, 0xE1, 0xA1, 0x43, 0x04, 0x02, 0x20, 0x00, 0x00, 0x41, 0x0D, 0xBB, 0x86, 0xF4, 0x31, 0x7F, 0xFF, 0x88, 0xFF, 0x37, 0x47, 0x1A, 0xDB, 0x6A, 0xDF, 0xFF, 0xAC}
	require.Len(t, shouldBe, 30)

	e := &Encoder{}
	e.Init()

	prev := uint32(0)
	for _, b := range testSequence {
		for i := 7; i >= 0; i-- {
			bit := (b >> uint(i)) & 0x1
			// fmt.Printf("Byte: %08b, Bit: %d, Value: %01b\n", b, i, bit)

			err := e.encodeBit(e.context, prev, bit)
			require.NoError(t, err)
		}
	}

	e.flush()

	assert.Equal(t, 30, e.outbufUsed)
	buf := &bytes.Buffer{}
	n, err := e.WriteTo(buf)
	assert.Equal(t, int64(30), n)

	require.NoError(t, err)

	result := buf.Bytes()
	assert.Len(t, result, 30)
	if assert.True(t, len(result) <= len(shouldBe)) {
		for i := 0; i < len(result); i++ {
			assert.Equal(t, shouldBe[i], result[i], "At index: '%d'", i)
		}
	}

}

// TestEncodeInteger tests the encode integer function.
func TestEncodeInteger(t *testing.T) {
	e := New()
	err := e.EncodeInteger(IADT, 5)
	require.NoError(t, err)

	err = e.EncodeInteger(IADH, 10)
	require.NoError(t, err)

	err = e.EncodeOOB(IADH)
	require.NoError(t, err)

	err = e.EncodeOOB(IADT)
	require.NoError(t, err)
	e.Final()

	buf := &bytes.Buffer{}
	_, err = e.WriteTo(buf)
	require.NoError(t, err)

	r := bitwise.NewReader(buf.Bytes())

	dec, err := arithmetic.New(r)
	require.NoError(t, err)

	dt, err := dec.DecodeInt(arithmetic.NewStats(512, 1))
	require.NoError(t, err)

	dh, err := dec.DecodeInt(arithmetic.NewStats(512, 1))
	require.NoError(t, err)

	assert.Equal(t, 5, int(dt))
	assert.Equal(t, 10, int(dh))
}

func TestEncoder_EncodeIAID(t *testing.T) {
	e := New()
	err := e.EncodeIAID(3, 4)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	e.Final()
	// write to buffer
	_, err = e.WriteTo(buf)
	require.NoError(t, err)

	r := bitwise.NewReader(buf.Bytes())
	d, err := arithmetic.New(r)
	require.NoError(t, err)

	v, err := d.DecodeIAID(3, arithmetic.NewStats(512, 0))
	require.NoError(t, err)

	assert.Equal(t, int64(4), v)
}
