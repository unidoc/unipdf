/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"bytes"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/common"
)

var coderDebugFlag bool

func init() {
	flag.BoolVar(&coderDebugFlag, "encoder-debug", false, "Shows debug logs for the jbig2 arithmetic encoder")
}

// TestEncoder tests the encoder using the standard H.2 test sequence.
func TestEncoder(t *testing.T) {
	testSequence := []byte{0x00, 0x02, 0x00, 0x51, 0x00, 0x00, 0x00, 0xC0, 0x03, 0x52, 0x87, 0x2A, 0xAA, 0xAA, 0xAA, 0xAA, 0x82, 0xC0, 0x20, 0x00, 0xFC, 0xD7, 0x9E, 0xF6, 0xBF, 0x7F, 0xED, 0x90, 0x4F, 0x46, 0xA3, 0xBF}
	require.Len(t, testSequence, 32)

	shouldBe := []byte{0x84, 0xC7, 0x3B, 0xFC, 0xE1, 0xA1, 0x43, 0x04, 0x02, 0x20, 0x00, 0x00, 0x41, 0x0D, 0xBB, 0x86, 0xF4, 0x31, 0x7F, 0xFF, 0x88, 0xFF, 0x37, 0x47, 0x1A, 0xDB, 0x6A, 0xDF, 0xFF, 0xAC}
	require.Len(t, shouldBe, 30)

	e := &Encoder{}
	e.Init()

	if coderDebugFlag {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

		CoderDebugging = true
	}

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
	n, err := e.toBuffer(buf)
	assert.Equal(t, 30, n)

	require.NoError(t, err)

	result := buf.Bytes()
	assert.Len(t, result, 30)
	if assert.True(t, len(result) <= len(shouldBe)) {
		for i := 0; i < len(result); i++ {
			assert.Equal(t, shouldBe[i], result[i], "At index: '%d'", i)
		}
	}

}
