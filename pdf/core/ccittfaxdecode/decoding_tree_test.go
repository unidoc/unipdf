package ccittfaxdecode

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
	for runLen, code := range WTerms {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range WMakeups {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range CommonMakeups {
		addNode(whiteTree, code, 0, runLen)
		addNode(blackTree, code, 0, runLen)
	}

	type testResult struct {
		RunLen int
		Code   Code
	}

	type testData struct {
		Codes []uint16
		Want  testResult
	}

	whiteTerminalTests := []testData{
		{
			Codes: formTestCodes(13568, 8),
			Want: testResult{
				RunLen: 0,
				Code:   WTerms[0],
			},
		},
		{
			Codes: formTestCodes(7168, 6),
			Want: testResult{
				RunLen: 1,
				Code:   WTerms[1],
			},
		},
		{
			Codes: formTestCodes(28672, 4),
			Want: testResult{
				RunLen: 2,
				Code:   WTerms[2],
			},
		},
		{
			Codes: formTestCodes(32768, 4),
			Want: testResult{
				RunLen: 3,
				Code:   WTerms[3],
			},
		},
		{
			Codes: formTestCodes(45056, 4),
			Want: testResult{
				RunLen: 4,
				Code:   WTerms[4],
			},
		},
		{
			Codes: formTestCodes(49152, 4),
			Want: testResult{
				RunLen: 5,
				Code:   WTerms[5],
			},
		},
		{
			Codes: formTestCodes(57344, 4),
			Want: testResult{
				RunLen: 6,
				Code:   WTerms[6],
			},
		},
		{
			Codes: formTestCodes(61440, 4),
			Want: testResult{
				RunLen: 7,
				Code:   WTerms[7],
			},
		},
		{
			Codes: formTestCodes(38912, 5),
			Want: testResult{
				RunLen: 8,
				Code:   WTerms[8],
			},
		},
		{
			Codes: formTestCodes(40960, 5),
			Want: testResult{
				RunLen: 9,
				Code:   WTerms[9],
			},
		},
		{
			Codes: formTestCodes(14336, 5),
			Want: testResult{
				RunLen: 10,
				Code:   WTerms[10],
			},
		},
		{
			Codes: formTestCodes(16384, 5),
			Want: testResult{
				RunLen: 11,
				Code:   WTerms[11],
			},
		},
		{
			Codes: formTestCodes(8192, 6),
			Want: testResult{
				RunLen: 12,
				Code:   WTerms[12],
			},
		},
		{
			Codes: formTestCodes(3072, 6),
			Want: testResult{
				RunLen: 13,
				Code:   WTerms[13],
			},
		},
		{
			Codes: formTestCodes(53248, 6),
			Want: testResult{
				RunLen: 14,
				Code:   WTerms[14],
			},
		},
		{
			Codes: formTestCodes(54272, 6),
			Want: testResult{
				RunLen: 15,
				Code:   WTerms[15],
			},
		},
		{
			Codes: formTestCodes(43008, 6),
			Want: testResult{
				RunLen: 16,
				Code:   WTerms[16],
			},
		},
		{
			Codes: formTestCodes(44032, 6),
			Want: testResult{
				RunLen: 17,
				Code:   WTerms[17],
			},
		},
		{
			Codes: formTestCodes(19968, 7),
			Want: testResult{
				RunLen: 18,
				Code:   WTerms[18],
			},
		},
		{
			Codes: formTestCodes(6144, 7),
			Want: testResult{
				RunLen: 19,
				Code:   WTerms[19],
			},
		},
		{
			Codes: formTestCodes(4096, 7),
			Want: testResult{
				RunLen: 20,
				Code:   WTerms[20],
			},
		},
		{
			Codes: formTestCodes(11776, 7),
			Want: testResult{
				RunLen: 21,
				Code:   WTerms[21],
			},
		},
		{
			Codes: formTestCodes(1536, 7),
			Want: testResult{
				RunLen: 22,
				Code:   WTerms[22],
			},
		},
		{
			Codes: formTestCodes(2048, 7),
			Want: testResult{
				RunLen: 23,
				Code:   WTerms[23],
			},
		},
		{
			Codes: formTestCodes(20480, 7),
			Want: testResult{
				RunLen: 24,
				Code:   WTerms[24],
			},
		},
		{
			Codes: formTestCodes(22016, 7),
			Want: testResult{
				RunLen: 25,
				Code:   WTerms[25],
			},
		},
		{
			Codes: formTestCodes(9728, 7),
			Want: testResult{
				RunLen: 26,
				Code:   WTerms[26],
			},
		},
		{
			Codes: formTestCodes(18432, 7),
			Want: testResult{
				RunLen: 27,
				Code:   WTerms[27],
			},
		},
		{
			Codes: formTestCodes(12288, 7),
			Want: testResult{
				RunLen: 28,
				Code:   WTerms[28],
			},
		},
		{
			Codes: formTestCodes(512, 8),
			Want: testResult{
				RunLen: 29,
				Code:   WTerms[29],
			},
		},
		{
			Codes: formTestCodes(768, 8),
			Want: testResult{
				RunLen: 30,
				Code:   WTerms[30],
			},
		},
		{
			Codes: formTestCodes(6656, 8),
			Want: testResult{
				RunLen: 31,
				Code:   WTerms[31],
			},
		},
		{
			Codes: formTestCodes(6912, 8),
			Want: testResult{
				RunLen: 32,
				Code:   WTerms[32],
			},
		},
		{
			Codes: formTestCodes(4608, 8),
			Want: testResult{
				RunLen: 33,
				Code:   WTerms[33],
			},
		},
		{
			Codes: formTestCodes(4864, 8),
			Want: testResult{
				RunLen: 34,
				Code:   WTerms[34],
			},
		},
		{
			Codes: formTestCodes(5120, 8),
			Want: testResult{
				RunLen: 35,
				Code:   WTerms[35],
			},
		},
		{
			Codes: formTestCodes(5376, 8),
			Want: testResult{
				RunLen: 36,
				Code:   WTerms[36],
			},
		},
		{
			Codes: formTestCodes(5632, 8),
			Want: testResult{
				RunLen: 37,
				Code:   WTerms[37],
			},
		},
		{
			Codes: formTestCodes(5888, 8),
			Want: testResult{
				RunLen: 38,
				Code:   WTerms[38],
			},
		},
		{
			Codes: formTestCodes(10240, 8),
			Want: testResult{
				RunLen: 39,
				Code:   WTerms[39],
			},
		},
		{
			Codes: formTestCodes(10496, 8),
			Want: testResult{
				RunLen: 40,
				Code:   WTerms[40],
			},
		},
		{
			Codes: formTestCodes(10752, 8),
			Want: testResult{
				RunLen: 41,
				Code:   WTerms[41],
			},
		},
		{
			Codes: formTestCodes(11008, 8),
			Want: testResult{
				RunLen: 42,
				Code:   WTerms[42],
			},
		},
		{
			Codes: formTestCodes(11264, 8),
			Want: testResult{
				RunLen: 43,
				Code:   WTerms[43],
			},
		},
		{
			Codes: formTestCodes(11520, 8),
			Want: testResult{
				RunLen: 44,
				Code:   WTerms[44],
			},
		},
		{
			Codes: formTestCodes(1024, 8),
			Want: testResult{
				RunLen: 45,
				Code:   WTerms[45],
			},
		},
		{
			Codes: formTestCodes(1280, 8),
			Want: testResult{
				RunLen: 46,
				Code:   WTerms[46],
			},
		},
		{
			Codes: formTestCodes(2560, 8),
			Want: testResult{
				RunLen: 47,
				Code:   WTerms[47],
			},
		},
		{
			Codes: formTestCodes(2816, 8),
			Want: testResult{
				RunLen: 48,
				Code:   WTerms[48],
			},
		},
		{
			Codes: formTestCodes(20992, 8),
			Want: testResult{
				RunLen: 49,
				Code:   WTerms[49],
			},
		},
		{
			Codes: formTestCodes(21248, 8),
			Want: testResult{
				RunLen: 50,
				Code:   WTerms[50],
			},
		},
		{
			Codes: formTestCodes(21504, 8),
			Want: testResult{
				RunLen: 51,
				Code:   WTerms[51],
			},
		},
		{
			Codes: formTestCodes(21760, 8),
			Want: testResult{
				RunLen: 52,
				Code:   WTerms[52],
			},
		},
		{
			Codes: formTestCodes(9216, 8),
			Want: testResult{
				RunLen: 53,
				Code:   WTerms[53],
			},
		},
		{
			Codes: formTestCodes(9472, 8),
			Want: testResult{
				RunLen: 54,
				Code:   WTerms[54],
			},
		},
		{
			Codes: formTestCodes(22528, 8),
			Want: testResult{
				RunLen: 55,
				Code:   WTerms[55],
			},
		},
		{
			Codes: formTestCodes(22784, 8),
			Want: testResult{
				RunLen: 56,
				Code:   WTerms[56],
			},
		},
		{
			Codes: formTestCodes(23040, 8),
			Want: testResult{
				RunLen: 57,
				Code:   WTerms[57],
			},
		},
		{
			Codes: formTestCodes(23296, 8),
			Want: testResult{
				RunLen: 58,
				Code:   WTerms[58],
			},
		},
		{
			Codes: formTestCodes(18944, 8),
			Want: testResult{
				RunLen: 59,
				Code:   WTerms[59],
			},
		},
		{
			Codes: formTestCodes(19200, 8),
			Want: testResult{
				RunLen: 60,
				Code:   WTerms[60],
			},
		},
		{
			Codes: formTestCodes(12800, 8),
			Want: testResult{
				RunLen: 61,
				Code:   WTerms[61],
			},
		},
		{
			Codes: formTestCodes(13056, 8),
			Want: testResult{
				RunLen: 62,
				Code:   WTerms[62],
			},
		},
		{
			Codes: formTestCodes(13312, 8),
			Want: testResult{
				RunLen: 63,
				Code:   WTerms[63],
			},
		},
	}

	for _, test := range whiteTerminalTests {
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

				if *gotCodePtr != test.Want.Code {
					t.Errorf("Wrong code for code: %v. Got %v, want %v\n",
						code, *gotCodePtr, test.Want.Code)
				}
			}
		}
	}
}
