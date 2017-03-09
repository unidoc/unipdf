/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package sampling

import (
	"fmt"
	"testing"
)

type TestResamplingTest1 struct {
	InputData     []byte
	BitsPerSample int
	Expected      []uint32
}

func samplesEqual(d1 []uint32, d2 []uint32) bool {
	if len(d1) != len(d2) {
		return false
	}

	for i := 0; i < len(d1); i++ {
		if d1[i] != d2[i] {
			return false
		}
	}

	return true
}

// Test resampling example data with different bits per sample.
// Input is in bytes.
func TestResamplingBytes(t *testing.T) {
	// 0xB5     0x5D     0x2A
	// 10110101 01011101 00101010
	// 2-bit resampling:
	// 10 11 01 01 01 01 11 01 00 10 10 10
	// 2  3  1  1  1  1  3  1  0  2  2  2
	// 3-bit resampling:
	// 101 101 010 101 110 100 101 010
	// 5   5   2   5   6   4   5   2
	// 4-bit resampling:
	// 1011 0101 0101 1101 0010 1010
	// 11   5    5    13   2    10
	// 5-bit resampling
	// 10110 10101 01110 10010 (1010)<- the remainder is dumped
	// 22    21    14    18
	// 12-bit resampling
	// 101101010101 110100101010
	// 2901         3370
	// 13-bit resampling
	// 1011010101011 (10100101010)
	// 16-bit resampling
	// 1011010101011101 (00101010)
	// 46429
	// 24-bit resampling
	// 101101010101110100101010
	//
	// 0xde 0xad 0xbe 0xef 0x15 0x13 0x37

	testcases := []TestResamplingTest1{
		{[]byte{0xB5, 0x5D, 0x2A}, 1, []uint32{1, 0, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 0, 1, 0, 1, 0, 1, 0}},
		{[]byte{0xB5, 0x5D, 0x2A}, 2, []uint32{2, 3, 1, 1, 1, 1, 3, 1, 0, 2, 2, 2}},
		{[]byte{0xB5, 0x5D, 0x2A}, 3, []uint32{5, 5, 2, 5, 6, 4, 5, 2}},
		{[]byte{0xB5, 0x5D, 0x2A}, 4, []uint32{11, 5, 5, 13, 2, 10}},
		{[]byte{0xB5, 0x5D, 0x2A}, 5, []uint32{22, 21, 14, 18}},
		{[]byte{0xB5, 0x5D, 0x2A}, 8, []uint32{0xB5, 0x5D, 0x2A}},
		{[]byte{0xB5, 0x5D, 0x2A}, 12, []uint32{2901, 3370}},
		{[]byte{0xB5, 0x5D, 0x2A}, 13, []uint32{5803}},
		{[]byte{0xB5, 0x5D, 0x2A}, 16, []uint32{0xB55D}},
		{[]byte{0xB5, 0x5D, 0x2A}, 24, []uint32{0xB55D2A}},
		{[]byte{0xde, 0xad, 0xbe, 0xef, 0x15, 0x13, 0x37, 0x20}, 24, []uint32{0xdeadbe, 0xef1513}},
		{[]byte{0xde, 0xad, 0xbe, 0xef, 0x15, 0x13, 0x37, 0x20, 0x21}, 32, []uint32{0xdeadbeef, 0x15133720}},
	}

	for _, testcase := range testcases {
		b := ResampleBytes(testcase.InputData, testcase.BitsPerSample)
		fmt.Println(b)
		if !samplesEqual(b, testcase.Expected) {
			t.Errorf("Test case failed. Got: % d, expected: % d", b, testcase.Expected)
			t.Errorf("Test case failed. Got: % X, expected: % X", b, testcase.Expected)
		}
	}
}

type TestResamplingTest2 struct {
	InputData           []uint32
	BitsPerInputSample  int
	BitsPerOutputSample int
	Expected            []uint32
}

// Test resampling example data with different bits per sample.
// Input is in uint32.
func TestResamplingUint32(t *testing.T) {
	// 0xB5     0x5D     0x2A     0x00
	// 10110101 01011101 00101010 00000000
	// 2-bit resampling:
	// 10 11 01 01 01 01 11 01 00 10 10 10 00 00 00 00
	// 2  3  1  1  1  1  3  1  0  2  2  2  0  0  0  0
	// 3-bit resampling:
	// 101 101 010 101 110 100 101 010 000 000 (00)
	// 5   5   2   5   6   4   5   2   0   0
	// 4-bit resampling:
	// 1011 0101 0101 1101 0010 1010 0000 0000
	// 11   5    5    13   2    10   0    0
	// 5-bit resampling
	// 10110 10101 01110 10010 10100 00000 (00)<- the remainder is dumped
	// 22    21    14    18    20    0
	// 12-bit resampling
	// 101101010101 110100101010 (00000000)
	// 2901         3370
	// 13-bit resampling
	// 1011010101011 1010010101000 (0000000)
	// 16-bit resampling
	// 1011010101011101 0010101000000000
	// 0xB55D           0x2A00
	// 24-bit resampling
	// 101101010101110100101010 (00000000)
	//
	// 0xde 0xad 0xbe 0xef 0x15 0x13 0x37

	testcases := []TestResamplingTest2{
		{[]uint32{0xB55D2A00}, 32, 1, []uint32{1, 0, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		{[]uint32{0xB55D2A00}, 32, 2, []uint32{2, 3, 1, 1, 1, 1, 3, 1, 0, 2, 2, 2, 0, 0, 0, 0}},
		{[]uint32{0xB55D2A00}, 32, 3, []uint32{5, 5, 2, 5, 6, 4, 5, 2, 0, 0}},
		{[]uint32{0xB55D2A00}, 32, 4, []uint32{11, 5, 5, 13, 2, 10, 0, 0}},
		{[]uint32{0xB55D2A00}, 32, 5, []uint32{22, 21, 14, 18, 20, 0}},
		{[]uint32{0xB55D2A00}, 32, 8, []uint32{0xB5, 0x5D, 0x2A, 0x00}},
		{[]uint32{0xB55D2A00}, 32, 12, []uint32{2901, 3370}},
		{[]uint32{0xB55D2A00}, 32, 13, []uint32{5803, 5288}},
		{[]uint32{0xB55D2A00}, 32, 16, []uint32{0xB55D, 0x2A00}},
		{[]uint32{0xB55D2A00}, 32, 24, []uint32{0xB55D2A}},
		{[]uint32{0xdeadbeef, 0x15133720}, 32, 24, []uint32{0xdeadbe, 0xef1513}},
		{[]uint32{0xdeadbeef, 0x15133720}, 32, 32, []uint32{0xdeadbeef, 0x15133720}},
	}

	for _, testcase := range testcases {
		b := ResampleUint32(testcase.InputData, testcase.BitsPerInputSample, testcase.BitsPerOutputSample)
		fmt.Println(b)
		if !samplesEqual(b, testcase.Expected) {
			//t.Errorf("Test case failed. Got: % d, expected: % d", b, testcase.Expected)
			t.Errorf("Test case failed. Got: % X, expected: % X", b, testcase.Expected)
		}
	}
}

// Test resampling example data with different bits per sample.
// Input is in uint32, certain number of bits.
func TestResamplingUint32xx(t *testing.T) {
	testcases := []TestResamplingTest2{
		{[]uint32{0, 0, 0}, 1, 8, []uint32{0}},
		{[]uint32{0, 1, 0}, 1, 8, []uint32{64}},
	}

	for _, testcase := range testcases {
		b := ResampleUint32(testcase.InputData, testcase.BitsPerInputSample, testcase.BitsPerOutputSample)
		fmt.Println(b)
		if !samplesEqual(b, testcase.Expected) {
			t.Errorf("Test case failed. Got: % d, expected: % d", b, testcase.Expected)
			t.Errorf("Test case failed. Got: % X, expected: % X", b, testcase.Expected)
		}
	}
}
