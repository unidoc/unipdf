/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"testing"
)

func TestImageResampling(t *testing.T) {
	img := Image{}

	// Case 1:
	// Data:
	// 4x8bit: 00000001 11101000 01101110 00001010
	// Resample as 1bit:
	//
	// 4x8bit: 00000001 11101000 01101110 00001010
	// Downsample to 1bit
	// 4x8bit: 00000000 00000001 00000000 00000000
	// 4x1bit: 0100
	// Padding with 4x00
	// -> 01000000 = 64 decimal
	//
	img.BitsPerComponent = 8
	img.Data = []byte{1, 232, 110, 10}
	//int(this.Width) * int(this.Height) * this.ColorComponents
	img.Width = 4
	img.ColorComponents = 1
	img.Height = 1
	img.Resample(1)
	if len(img.Data) != 1 {
		t.Errorf("Incorrect length != 1 (%d)", len(img.Data))
		return
	}
	if img.Data[0] != 64 {
		t.Errorf("Value != 4 (%d)", img.Data[0])
	}

	// Case 2:
	// Data:
	// 4x8bit: 00000001 11101000 01101110 00001010 00000001 11101000 01101110 00001010 00000001 11101000 01101110 00001010
	//         0        1        0        0        0        1        0        0        0        1        0        0
	// 010001000100
	// -> 01000100 0100(0000)
	// -> 68 64
	img.BitsPerComponent = 8
	img.Data = []byte{1, 232, 110, 10, 1, 232, 110, 10, 1, 232, 110, 10}
	img.Width = 12
	img.ColorComponents = 1
	img.Height = 1
	img.Resample(1)

	if len(img.Data) != 2 {
		t.Errorf("Incorrect length != 2 (%d)", len(img.Data))
		return
	}
	if img.Data[0] != 68 {
		t.Errorf("Value != 68 (%d)", img.Data[0])
	}
	if img.Data[1] != 64 {
		t.Errorf("Value != 64 (%d)", img.Data[1])
	}
}
