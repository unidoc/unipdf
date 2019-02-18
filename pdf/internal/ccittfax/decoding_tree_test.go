/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ccittfax

import (
	"testing"
)

func formTestCodes(code uint16, bitsWritten int) []uint16 {
	var codes []uint16

	currentCombination := code
	for {
		codes = append(codes, currentCombination)

		i := 15
		for i >= bitsWritten {
			bit := bitFromUint16(currentCombination, i)

			if bit == 0 {
				currentCombination = setBitToUint16(currentCombination, 1, i)

				break
			} else {
				currentCombination = setBitToUint16(currentCombination, 0, i)
			}

			i--
		}

		if i < bitsWritten {
			break
		}
	}

	return codes
}

func TestFormCodes(t *testing.T) {
	tests := []struct {
		Code        uint16
		BitsWritten int
		Want        []uint16
	}{
		{
			Code:        65520,
			BitsWritten: 12,
			Want: []uint16{
				65520,
				65521,
				65522,
				65523,
				65524,
				65525,
				65526,
				65527,
				65528,
				65529,
				65530,
				65531,
				65532,
				65533,
				65534,
				65535,
			},
		},
		{
			Code:        65529,
			BitsWritten: 12,
			Want: []uint16{
				65529,
				65530,
				65531,
				65532,
				65533,
				65534,
				65535,
			},
		},
		{
			Code:        65531,
			BitsWritten: 12,
			Want: []uint16{
				65531,
				65532,
				65533,
				65534,
				65535,
			},
		},
	}

	for _, test := range tests {
		gotCodes := formTestCodes(test.Code, test.BitsWritten)

		if len(gotCodes) != len(test.Want) {
			t.Errorf("Wrong codes len. Got %v, want %v\n", len(gotCodes), len(test.Want))
		} else {
			for i := range gotCodes {
				if gotCodes[i] != test.Want[i] {
					t.Errorf("Slices differ in %v. Got %v, want %v\n", i, gotCodes[i], test.Want[i])

					break
				}
			}
		}
	}
}

func setBitToUint16(num uint16, val byte, pos int) uint16 {
	mask := uint16(0xFFFE)<<uint8(15-pos) | uint16(0x7FFF)>>uint(pos)

	shiftedVal := uint16(val) << uint8(15-pos)

	return num&mask | shiftedVal
}

func TestSetBitToUint16(t *testing.T) {
	tests := []struct {
		Num  uint16
		Want uint16
	}{
		{
			Num:  0,
			Want: 1,
		},
		{
			Num:  1,
			Want: 3,
		},
		{
			Num:  3,
			Want: 7,
		},
		{
			Num:  7,
			Want: 15,
		},
		{
			Num:  15,
			Want: 31,
		},
		{
			Num:  31,
			Want: 63,
		},
		{
			Num:  63,
			Want: 127,
		},
		{
			Num:  127,
			Want: 255,
		},
		{
			Num:  255,
			Want: 511,
		},
		{
			Num:  511,
			Want: 1023,
		},
		{
			Num:  1023,
			Want: 2047,
		},
		{
			Num:  2047,
			Want: 4095,
		},
		{
			Num:  4095,
			Want: 8191,
		},
		{
			Num:  8191,
			Want: 16383,
		},
		{
			Num:  16383,
			Want: 32767,
		},
		{
			Num:  32767,
			Want: 65535,
		},
	}

	for i, test := range tests {
		gotNum := setBitToUint16(test.Num, 1, 15-i)

		if gotNum != test.Want {
			t.Errorf("Wrong value. Got %v, want %v\n", gotNum, test.Want)
		}
	}

	for i := len(tests) - 1; i >= 0; i-- {
		gotNum := setBitToUint16(tests[i].Want, 0, 15-i)

		if gotNum != tests[i].Num {
			t.Errorf("Wrong value. Got %v, want %v\n", gotNum, tests[i].Want)
		}
	}
}

func TestFindRunLen(t *testing.T) {
	for runLen, code := range wTerms {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range wMakeups {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range bTerms {
		addNode(blackTree, code, 0, runLen)
	}

	for runLen, code := range bMakeups {
		addNode(blackTree, code, 0, runLen)
	}

	for runLen, code := range commonMakeups {
		addNode(whiteTree, code, 0, runLen)
		addNode(blackTree, code, 0, runLen)
	}

	addNode(twoDimTree, p, 0, 0)
	addNode(twoDimTree, h, 0, 0)
	addNode(twoDimTree, v0, 0, 0)
	addNode(twoDimTree, v1r, 0, 0)
	addNode(twoDimTree, v2r, 0, 0)
	addNode(twoDimTree, v3r, 0, 0)
	addNode(twoDimTree, v1l, 0, 0)
	addNode(twoDimTree, v2l, 0, 0)
	addNode(twoDimTree, v3l, 0, 0)

	type testResult struct {
		RunLen      int
		Code        uint16
		BitsWritten int
	}

	type testData struct {
		Codes []uint16
		Want  testResult
	}

	whiteTests := []testData{
		{
			Codes: formTestCodes(13568, 8),
			Want: testResult{
				RunLen:      0,
				Code:        13568,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(7168, 6),
			Want: testResult{
				RunLen:      1,
				Code:        7168,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(28672, 4),
			Want: testResult{
				RunLen:      2,
				Code:        28672,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(32768, 4),
			Want: testResult{
				RunLen:      3,
				Code:        32768,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(45056, 4),
			Want: testResult{
				RunLen:      4,
				Code:        45056,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(49152, 4),
			Want: testResult{
				RunLen:      5,
				Code:        49152,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(57344, 4),
			Want: testResult{
				RunLen:      6,
				Code:        57344,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(61440, 4),
			Want: testResult{
				RunLen:      7,
				Code:        61440,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(38912, 5),
			Want: testResult{
				RunLen:      8,
				Code:        38912,
				BitsWritten: 5,
			},
		},
		{
			Codes: formTestCodes(40960, 5),
			Want: testResult{
				RunLen:      9,
				Code:        40960,
				BitsWritten: 5,
			},
		},
		{
			Codes: formTestCodes(14336, 5),
			Want: testResult{
				RunLen:      10,
				Code:        14336,
				BitsWritten: 5,
			},
		},
		{
			Codes: formTestCodes(16384, 5),
			Want: testResult{
				RunLen:      11,
				Code:        16384,
				BitsWritten: 5,
			},
		},
		{
			Codes: formTestCodes(8192, 6),
			Want: testResult{
				RunLen:      12,
				Code:        8192,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(3072, 6),
			Want: testResult{
				RunLen:      13,
				Code:        3072,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(53248, 6),
			Want: testResult{
				RunLen:      14,
				Code:        53248,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(54272, 6),
			Want: testResult{
				RunLen:      15,
				Code:        54272,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(43008, 6),
			Want: testResult{
				RunLen:      16,
				Code:        43008,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(44032, 6),
			Want: testResult{
				RunLen:      17,
				Code:        44032,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(19968, 7),
			Want: testResult{
				RunLen:      18,
				Code:        19968,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(6144, 7),
			Want: testResult{
				RunLen:      19,
				Code:        6144,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(4096, 7),
			Want: testResult{
				RunLen:      20,
				Code:        4096,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(11776, 7),
			Want: testResult{
				RunLen:      21,
				Code:        11776,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(1536, 7),
			Want: testResult{
				RunLen:      22,
				Code:        1536,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(2048, 7),
			Want: testResult{
				RunLen:      23,
				Code:        2048,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(20480, 7),
			Want: testResult{
				RunLen:      24,
				Code:        20480,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(22016, 7),
			Want: testResult{
				RunLen:      25,
				Code:        22016,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(9728, 7),
			Want: testResult{
				RunLen:      26,
				Code:        9728,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(18432, 7),
			Want: testResult{
				RunLen:      27,
				Code:        18432,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(12288, 7),
			Want: testResult{
				RunLen:      28,
				Code:        12288,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(512, 8),
			Want: testResult{
				RunLen:      29,
				Code:        512,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(768, 8),
			Want: testResult{
				RunLen:      30,
				Code:        768,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(6656, 8),
			Want: testResult{
				RunLen:      31,
				Code:        6656,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(6912, 8),
			Want: testResult{
				RunLen:      32,
				Code:        6912,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(4608, 8),
			Want: testResult{
				RunLen:      33,
				Code:        4608,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(4864, 8),
			Want: testResult{
				RunLen:      34,
				Code:        4864,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(5120, 8),
			Want: testResult{
				RunLen:      35,
				Code:        5120,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(5376, 8),
			Want: testResult{
				RunLen:      36,
				Code:        5376,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(5632, 8),
			Want: testResult{
				RunLen:      37,
				Code:        5632,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(5888, 8),
			Want: testResult{
				RunLen:      38,
				Code:        5888,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(10240, 8),
			Want: testResult{
				RunLen:      39,
				Code:        10240,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(10496, 8),
			Want: testResult{
				RunLen:      40,
				Code:        10496,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(10752, 8),
			Want: testResult{
				RunLen:      41,
				Code:        10752,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(11008, 8),
			Want: testResult{
				RunLen:      42,
				Code:        11008,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(11264, 8),
			Want: testResult{
				RunLen:      43,
				Code:        11264,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(11520, 8),
			Want: testResult{
				RunLen:      44,
				Code:        11520,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(1024, 8),
			Want: testResult{
				RunLen:      45,
				Code:        1024,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(1280, 8),
			Want: testResult{
				RunLen:      46,
				Code:        1280,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(2560, 8),
			Want: testResult{
				RunLen:      47,
				Code:        2560,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(2816, 8),
			Want: testResult{
				RunLen:      48,
				Code:        2816,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(20992, 8),
			Want: testResult{
				RunLen:      49,
				Code:        20992,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(21248, 8),
			Want: testResult{
				RunLen:      50,
				Code:        21248,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(21504, 8),
			Want: testResult{
				RunLen:      51,
				Code:        21504,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(21760, 8),
			Want: testResult{
				RunLen:      52,
				Code:        21760,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(9216, 8),
			Want: testResult{
				RunLen:      53,
				Code:        9216,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(9472, 8),
			Want: testResult{
				RunLen:      54,
				Code:        9472,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(22528, 8),
			Want: testResult{
				RunLen:      55,
				Code:        22528,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(22784, 8),
			Want: testResult{
				RunLen:      56,
				Code:        22784,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(23040, 8),
			Want: testResult{
				RunLen:      57,
				Code:        23040,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(23296, 8),
			Want: testResult{
				RunLen:      58,
				Code:        23296,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(18944, 8),
			Want: testResult{
				RunLen:      59,
				Code:        18944,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(19200, 8),
			Want: testResult{
				RunLen:      60,
				Code:        19200,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(12800, 8),
			Want: testResult{
				RunLen:      61,
				Code:        12800,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(13056, 8),
			Want: testResult{
				RunLen:      62,
				Code:        13056,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(13312, 8),
			Want: testResult{
				RunLen:      63,
				Code:        13312,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(55296, 5),
			Want: testResult{
				RunLen:      64,
				Code:        55296,
				BitsWritten: 5,
			},
		},
		{
			Codes: formTestCodes(36864, 5),
			Want: testResult{
				RunLen:      128,
				Code:        36864,
				BitsWritten: 5,
			},
		},
		{
			Codes: formTestCodes(23552, 6),
			Want: testResult{
				RunLen:      192,
				Code:        23552,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(28160, 7),
			Want: testResult{
				RunLen:      256,
				Code:        28160,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(13824, 8),
			Want: testResult{
				RunLen:      320,
				Code:        13824,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(14080, 8),
			Want: testResult{
				RunLen:      384,
				Code:        14080,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(25600, 8),
			Want: testResult{
				RunLen:      448,
				Code:        25600,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(25856, 8),
			Want: testResult{
				RunLen:      512,
				Code:        25856,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(26624, 8),
			Want: testResult{
				RunLen:      576,
				Code:        26624,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(26368, 8),
			Want: testResult{
				RunLen:      640,
				Code:        26368,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(26112, 9),
			Want: testResult{
				RunLen:      704,
				Code:        26112,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(26240, 9),
			Want: testResult{
				RunLen:      768,
				Code:        26240,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(26880, 9),
			Want: testResult{
				RunLen:      832,
				Code:        26880,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27008, 9),
			Want: testResult{
				RunLen:      896,
				Code:        27008,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27136, 9),
			Want: testResult{
				RunLen:      960,
				Code:        27136,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27264, 9),
			Want: testResult{
				RunLen:      1024,
				Code:        27264,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27392, 9),
			Want: testResult{
				RunLen:      1088,
				Code:        27392,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27520, 9),
			Want: testResult{
				RunLen:      1152,
				Code:        27520,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27648, 9),
			Want: testResult{
				RunLen:      1216,
				Code:        27648,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27776, 9),
			Want: testResult{
				RunLen:      1280,
				Code:        27776,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(27904, 9),
			Want: testResult{
				RunLen:      1344,
				Code:        27904,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(28032, 9),
			Want: testResult{
				RunLen:      1408,
				Code:        28032,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(19456, 9),
			Want: testResult{
				RunLen:      1472,
				Code:        19456,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(19584, 9),
			Want: testResult{
				RunLen:      1536,
				Code:        19584,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(19712, 9),
			Want: testResult{
				RunLen:      1600,
				Code:        19712,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(24576, 6),
			Want: testResult{
				RunLen:      1664,
				Code:        24576,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(19840, 9),
			Want: testResult{
				RunLen:      1728,
				Code:        19840,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(256, 11),
			Want: testResult{
				RunLen:      1792,
				Code:        256,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(384, 11),
			Want: testResult{
				RunLen:      1856,
				Code:        384,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(416, 11),
			Want: testResult{
				RunLen:      1920,
				Code:        416,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(288, 12),
			Want: testResult{
				RunLen:      1984,
				Code:        288,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(304, 12),
			Want: testResult{
				RunLen:      2048,
				Code:        304,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(320, 12),
			Want: testResult{
				RunLen:      2112,
				Code:        320,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(336, 12),
			Want: testResult{
				RunLen:      2176,
				Code:        336,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(352, 12),
			Want: testResult{
				RunLen:      2240,
				Code:        352,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(368, 12),
			Want: testResult{
				RunLen:      2304,
				Code:        368,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(448, 12),
			Want: testResult{
				RunLen:      2368,
				Code:        448,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(464, 12),
			Want: testResult{
				RunLen:      2432,
				Code:        464,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(480, 12),
			Want: testResult{
				RunLen:      2496,
				Code:        480,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(496, 12),
			Want: testResult{
				RunLen:      2560,
				Code:        496,
				BitsWritten: 12,
			},
		},
	}

	blackTests := []testData{
		{
			Codes: formTestCodes(3520, 10),
			Want: testResult{
				RunLen:      0,
				Code:        3520,
				BitsWritten: 10,
			},
		},
		{
			Codes: formTestCodes(16384, 3),
			Want: testResult{
				RunLen:      1,
				Code:        16384,
				BitsWritten: 3,
			},
		},
		{
			Codes: formTestCodes(49152, 2),
			Want: testResult{
				RunLen:      2,
				Code:        49152,
				BitsWritten: 2,
			},
		},
		{
			Codes: formTestCodes(32768, 2),
			Want: testResult{
				RunLen:      3,
				Code:        32768,
				BitsWritten: 2,
			},
		},
		{
			Codes: formTestCodes(24576, 3),
			Want: testResult{
				RunLen:      4,
				Code:        24576,
				BitsWritten: 3,
			},
		},
		{
			Codes: formTestCodes(12288, 4),
			Want: testResult{
				RunLen:      5,
				Code:        12288,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(8192, 4),
			Want: testResult{
				RunLen:      6,
				Code:        8192,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(6144, 5),
			Want: testResult{
				RunLen:      7,
				Code:        6144,
				BitsWritten: 5,
			},
		},
		{
			Codes: formTestCodes(5120, 6),
			Want: testResult{
				RunLen:      8,
				Code:        5120,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(4096, 6),
			Want: testResult{
				RunLen:      9,
				Code:        4096,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(2048, 7),
			Want: testResult{
				RunLen:      10,
				Code:        2048,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(2560, 7),
			Want: testResult{
				RunLen:      11,
				Code:        2560,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(3584, 7),
			Want: testResult{
				RunLen:      12,
				Code:        3584,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(1024, 8),
			Want: testResult{
				RunLen:      13,
				Code:        1024,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(1792, 8),
			Want: testResult{
				RunLen:      14,
				Code:        1792,
				BitsWritten: 8,
			},
		},
		{
			Codes: formTestCodes(3072, 9),
			Want: testResult{
				RunLen:      15,
				Code:        3072,
				BitsWritten: 9,
			},
		},
		{
			Codes: formTestCodes(1472, 10),
			Want: testResult{
				RunLen:      16,
				Code:        1472,
				BitsWritten: 10,
			},
		},
		{
			Codes: formTestCodes(1536, 10),
			Want: testResult{
				RunLen:      17,
				Code:        1536,
				BitsWritten: 10,
			},
		},
		{
			Codes: formTestCodes(512, 10),
			Want: testResult{
				RunLen:      18,
				Code:        512,
				BitsWritten: 10,
			},
		},
		{
			Codes: formTestCodes(3296, 11),
			Want: testResult{
				RunLen:      19,
				Code:        3296,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(3328, 11),
			Want: testResult{
				RunLen:      20,
				Code:        3328,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(3456, 11),
			Want: testResult{
				RunLen:      21,
				Code:        3456,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(1760, 11),
			Want: testResult{
				RunLen:      22,
				Code:        1760,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(1280, 11),
			Want: testResult{
				RunLen:      23,
				Code:        1280,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(736, 11),
			Want: testResult{
				RunLen:      24,
				Code:        736,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(768, 11),
			Want: testResult{
				RunLen:      25,
				Code:        768,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(3232, 12),
			Want: testResult{
				RunLen:      26,
				Code:        3232,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3248, 12),
			Want: testResult{
				RunLen:      27,
				Code:        3248,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3264, 12),
			Want: testResult{
				RunLen:      28,
				Code:        3264,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3280, 12),
			Want: testResult{
				RunLen:      29,
				Code:        3280,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1664, 12),
			Want: testResult{
				RunLen:      30,
				Code:        1664,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1680, 12),
			Want: testResult{
				RunLen:      31,
				Code:        1680,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1696, 12),
			Want: testResult{
				RunLen:      32,
				Code:        1696,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1712, 12),
			Want: testResult{
				RunLen:      33,
				Code:        1712,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3360, 12),
			Want: testResult{
				RunLen:      34,
				Code:        3360,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3376, 12),
			Want: testResult{
				RunLen:      35,
				Code:        3376,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3392, 12),
			Want: testResult{
				RunLen:      36,
				Code:        3392,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3408, 12),
			Want: testResult{
				RunLen:      37,
				Code:        3408,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3424, 12),
			Want: testResult{
				RunLen:      38,
				Code:        3424,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3440, 12),
			Want: testResult{
				RunLen:      39,
				Code:        3440,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1728, 12),
			Want: testResult{
				RunLen:      40,
				Code:        1728,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1744, 12),
			Want: testResult{
				RunLen:      41,
				Code:        1744,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3488, 12),
			Want: testResult{
				RunLen:      42,
				Code:        3488,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3504, 12),
			Want: testResult{
				RunLen:      43,
				Code:        3504,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1344, 12),
			Want: testResult{
				RunLen:      44,
				Code:        1344,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1360, 12),
			Want: testResult{
				RunLen:      45,
				Code:        1360,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1376, 12),
			Want: testResult{
				RunLen:      46,
				Code:        1376,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1392, 12),
			Want: testResult{
				RunLen:      47,
				Code:        1392,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1600, 12),
			Want: testResult{
				RunLen:      48,
				Code:        1600,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1616, 12),
			Want: testResult{
				RunLen:      49,
				Code:        1616,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1312, 12),
			Want: testResult{
				RunLen:      50,
				Code:        1312,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1328, 12),
			Want: testResult{
				RunLen:      51,
				Code:        1328,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(576, 12),
			Want: testResult{
				RunLen:      52,
				Code:        576,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(880, 12),
			Want: testResult{
				RunLen:      53,
				Code:        880,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(896, 12),
			Want: testResult{
				RunLen:      54,
				Code:        896,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(624, 12),
			Want: testResult{
				RunLen:      55,
				Code:        624,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(640, 12),
			Want: testResult{
				RunLen:      56,
				Code:        640,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1408, 12),
			Want: testResult{
				RunLen:      57,
				Code:        1408,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1424, 12),
			Want: testResult{
				RunLen:      58,
				Code:        1424,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(688, 12),
			Want: testResult{
				RunLen:      59,
				Code:        688,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(704, 12),
			Want: testResult{
				RunLen:      60,
				Code:        704,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1440, 12),
			Want: testResult{
				RunLen:      61,
				Code:        1440,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1632, 12),
			Want: testResult{
				RunLen:      62,
				Code:        1632,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1648, 12),
			Want: testResult{
				RunLen:      63,
				Code:        1648,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(960, 10),
			Want: testResult{
				RunLen:      64,
				Code:        960,
				BitsWritten: 10,
			},
		},
		{
			Codes: formTestCodes(3200, 12),
			Want: testResult{
				RunLen:      128,
				Code:        3200,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(3216, 12),
			Want: testResult{
				RunLen:      192,
				Code:        3216,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(1456, 12),
			Want: testResult{
				RunLen:      256,
				Code:        1456,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(816, 12),
			Want: testResult{
				RunLen:      320,
				Code:        816,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(832, 12),
			Want: testResult{
				RunLen:      384,
				Code:        832,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(848, 12),
			Want: testResult{
				RunLen:      448,
				Code:        848,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(864, 13),
			Want: testResult{
				RunLen:      512,
				Code:        864,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(872, 13),
			Want: testResult{
				RunLen:      576,
				Code:        872,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(592, 13),
			Want: testResult{
				RunLen:      640,
				Code:        592,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(600, 13),
			Want: testResult{
				RunLen:      704,
				Code:        600,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(608, 13),
			Want: testResult{
				RunLen:      768,
				Code:        608,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(616, 13),
			Want: testResult{
				RunLen:      832,
				Code:        616,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(912, 13),
			Want: testResult{
				RunLen:      896,
				Code:        912,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(920, 13),
			Want: testResult{
				RunLen:      960,
				Code:        920,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(928, 13),
			Want: testResult{
				RunLen:      1024,
				Code:        928,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(936, 13),
			Want: testResult{
				RunLen:      1088,
				Code:        936,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(944, 13),
			Want: testResult{
				RunLen:      1152,
				Code:        944,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(952, 13),
			Want: testResult{
				RunLen:      1216,
				Code:        952,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(656, 13),
			Want: testResult{
				RunLen:      1280,
				Code:        656,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(664, 13),
			Want: testResult{
				RunLen:      1344,
				Code:        664,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(672, 13),
			Want: testResult{
				RunLen:      1408,
				Code:        672,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(680, 13),
			Want: testResult{
				RunLen:      1472,
				Code:        680,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(720, 13),
			Want: testResult{
				RunLen:      1536,
				Code:        720,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(728, 13),
			Want: testResult{
				RunLen:      1600,
				Code:        728,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(800, 13),
			Want: testResult{
				RunLen:      1664,
				Code:        800,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(808, 13),
			Want: testResult{
				RunLen:      1728,
				Code:        808,
				BitsWritten: 13,
			},
		},
		{
			Codes: formTestCodes(256, 11),
			Want: testResult{
				RunLen:      1792,
				Code:        256,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(384, 11),
			Want: testResult{
				RunLen:      1856,
				Code:        384,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(416, 11),
			Want: testResult{
				RunLen:      1920,
				Code:        416,
				BitsWritten: 11,
			},
		},
		{
			Codes: formTestCodes(288, 12),
			Want: testResult{
				RunLen:      1984,
				Code:        288,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(304, 12),
			Want: testResult{
				RunLen:      2048,
				Code:        304,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(320, 12),
			Want: testResult{
				RunLen:      2112,
				Code:        320,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(336, 12),
			Want: testResult{
				RunLen:      2176,
				Code:        336,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(352, 12),
			Want: testResult{
				RunLen:      2240,
				Code:        352,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(368, 12),
			Want: testResult{
				RunLen:      2304,
				Code:        368,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(448, 12),
			Want: testResult{
				RunLen:      2368,
				Code:        448,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(464, 12),
			Want: testResult{
				RunLen:      2432,
				Code:        464,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(480, 12),
			Want: testResult{
				RunLen:      2496,
				Code:        480,
				BitsWritten: 12,
			},
		},
		{
			Codes: formTestCodes(496, 12),
			Want: testResult{
				RunLen:      2560,
				Code:        496,
				BitsWritten: 12,
			},
		},
	}

	twoDimTests := []struct {
		Codes []uint16
		Want  testResult
	}{
		{
			Codes: formTestCodes(4096, 4),
			Want: testResult{
				RunLen:      0,
				Code:        4096,
				BitsWritten: 4,
			},
		},
		{
			Codes: formTestCodes(8192, 3),
			Want: testResult{
				RunLen:      0,
				Code:        8192,
				BitsWritten: 3,
			},
		},
		{
			Codes: formTestCodes(32768, 1),
			Want: testResult{
				RunLen:      0,
				Code:        32768,
				BitsWritten: 1,
			},
		},
		{
			Codes: formTestCodes(24576, 3),
			Want: testResult{
				RunLen:      0,
				Code:        24576,
				BitsWritten: 3,
			},
		},
		{
			Codes: formTestCodes(3072, 6),
			Want: testResult{
				RunLen:      0,
				Code:        3072,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(1536, 7),
			Want: testResult{
				RunLen:      0,
				Code:        1536,
				BitsWritten: 7,
			},
		},
		{
			Codes: formTestCodes(16384, 3),
			Want: testResult{
				RunLen:      0,
				Code:        16384,
				BitsWritten: 3,
			},
		},
		{
			Codes: formTestCodes(2048, 6),
			Want: testResult{
				RunLen:      0,
				Code:        2048,
				BitsWritten: 6,
			},
		},
		{
			Codes: formTestCodes(1024, 7),
			Want: testResult{
				RunLen:      0,
				Code:        1024,
				BitsWritten: 7,
			},
		},
	}

	for _, test := range whiteTests {
		for _, code := range test.Codes {
			gotRunLenPtr, gotCodePtr := findRunLen(whiteTree, code, 0)

			if gotRunLenPtr == nil {
				t.Errorf("Got nil value for run len for code %v\n", code)
			} else if gotCodePtr == nil {
				t.Errorf("Got nil value for code for code: %v\n", code)
			} else {
				if *gotRunLenPtr != test.Want.RunLen {
					t.Errorf("Wrong run len for code: %v. Got %v, want %v\n",
						code, *gotRunLenPtr, test.Want.RunLen)
				}

				if gotCodePtr.Code != test.Want.Code {
					t.Errorf("Wrong code for code: %v. Got %v, want %v\n",
						code, *gotCodePtr, test.Want.Code)
				}

				if gotCodePtr.BitsWritten != test.Want.BitsWritten {
					t.Errorf("Wrong bits written for code: %v. Got %v, want %v\n",
						code, gotCodePtr.BitsWritten, test.Want.BitsWritten)
				}
			}
		}
	}

	for _, test := range blackTests {
		for _, code := range test.Codes {
			gotRunLenPtr, gotCodePtr := findRunLen(blackTree, code, 0)

			if gotRunLenPtr == nil {
				t.Errorf("Got nil value for run len for code %v\n", code)
			} else if gotCodePtr == nil {
				t.Errorf("Got nil value for code for code: %v\n", code)
			} else {
				if *gotRunLenPtr != test.Want.RunLen {
					t.Errorf("Wrong run len for code: %v. Got %v, want %v\n",
						code, *gotRunLenPtr, test.Want.RunLen)
				}

				if gotCodePtr.Code != test.Want.Code {
					t.Errorf("Wrong code for code: %v. Got %v, want %v\n",
						code, *gotCodePtr, test.Want.Code)
				}

				if gotCodePtr.BitsWritten != test.Want.BitsWritten {
					t.Errorf("Wrong bits written for code: %v. Got %v, want %v\n",
						code, gotCodePtr.BitsWritten, test.Want.BitsWritten)
				}
			}
		}
	}

	for _, test := range twoDimTests {
		for _, code := range test.Codes {
			gotRunLenPtr, gotCodePtr := findRunLen(twoDimTree, code, 0)

			if gotRunLenPtr == nil {
				t.Errorf("Got nil value for run len for code %v\n", code)
			} else if gotCodePtr == nil {
				t.Errorf("Got nil value for code for code: %v\n", code)
			} else {
				if *gotRunLenPtr != test.Want.RunLen {
					t.Errorf("Wrong run len for code: %v. Got %v, want %v\n",
						code, *gotRunLenPtr, test.Want.RunLen)
				}

				if gotCodePtr.Code != test.Want.Code {
					t.Errorf("Wrong code for code: %v. Got %v, want %v\n",
						code, *gotCodePtr, test.Want.Code)
				}

				if gotCodePtr.BitsWritten != test.Want.BitsWritten {
					t.Errorf("Wrong bits written for code: %v. Got %v, want %v\n",
						code, gotCodePtr.BitsWritten, test.Want.BitsWritten)
				}
			}
		}
	}
}
