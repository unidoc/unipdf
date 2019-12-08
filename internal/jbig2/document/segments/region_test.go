/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// TestEncodeRegion tests the encode function for the region structure.
func TestEncodeRegion(t *testing.T) {
	regionSample := &RegionSegment{
		BitmapWidth:        30,
		BitmapHeight:       150,
		XLocation:          40,
		YLocation:          35,
		CombinaionOperator: bitmap.CmbOpXor,
	}
	w := writer.BufferedMSB()
	n, err := regionSample.Encode(w)
	require.NoError(t, err)

	// the encode should write exactly 17 bytes
	assert.Equal(t, 17, n)

	// create expected byte data slice.
	expected := []byte{
		// First four bytes should be the width - 30 uint32
		0x00, 0x00, 0x00, 0x1E,
		// Second four bytes should be the height - 150 uint32
		0x00, 0x00, 0x00, 0x96,
		// Next four bytes should be the x location of the region - 40 uint32
		0x00, 0x00, 0x00, 0x28,
		// Next four bytes should be the y location of the region - 35 uint32
		0x00, 0x00, 0x00, 0x23,
		// The last byte should define the region flags
		// 5 empty bits and 3 bits for the combination operator
		// The XOR enum has the enum value of '2' thus the byte would be: 00000011
		0x02,
	}
	assert.Equal(t, expected, w.Data())
}
