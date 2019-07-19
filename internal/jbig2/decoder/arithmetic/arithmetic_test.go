/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// TestArithmeticDecoder tests the arithmetic decoder with the Apendix A methods.
func TestArithmeticDecoder(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	encoded := []byte{
		0x84, 0xC7, 0x3B, 0xFC, 0xE1, 0xA1, 0x43, 0x04, 0x02,
		0x20, 0x00, 0x00, 0x41, 0x0D, 0xBB, 0x86, 0xF4, 0x31,
		0x7F, 0xFF, 0x88, 0xFF, 0x37, 0x47, 0x1A, 0xDB, 0x6A,
		0xDF, 0xFF, 0xAC,
	}

	r := reader.New(encoded)
	a, err := New(r)
	if assert.NoError(t, err) {
		cx := NewStats(512, 0)
		for i := 0; i < 256; i++ {
			a.DecodeBit(cx)
		}
	}
}
