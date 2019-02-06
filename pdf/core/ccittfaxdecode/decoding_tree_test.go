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

/*func TestFindRunLen(t *testing.T) {
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

	tests := []struct {
		Code uint16
		Want testResult
	}{
		{
			Code: 13568,
			Want: testResult{
				RunLen: 0,
				Code:   WTerms[0],
			},
		},
		{
			Code: 7168,
			Want: testResult{
				RunLen: 1,
				Code:   WTerms[1],
			},
		},
		{
			Code: 28672,
			Want: testResult{
				RunLen: 2,
				Code:   WTerms[2],
			},
		},
		{
			Code: 32768,
			Want: testResult{
				RunLen: 3,
				Code:   WTerms[3],
			},
		},
		{
			Code: 45056,
			Want: testResult{
				RunLen: 4,
				Code:   WTerms[4],
			},
		},
		{
			Code: 49152,
			Want: testResult{
				RunLen: 5,
				Code:   WTerms[5],
			},
		},
		{
			Code: 57344,
			Want: testResult{
				RunLen: 6,
				Code:   WTerms[6],
			},
		},
		{
			Code: 61440,
			Want: testResult{
				RunLen: 7,
				Code:   WTerms[7],
			},
		},
		{
			Code: 38912,
			Want: testResult{
				RunLen: 8,
				Code:   WTerms[8],
			},
		},
		{
			Code: 40960,
			Want: testResult{
				RunLen: 9,
				Code:   WTerms[9],
			},
		},
		{
			Code: 14336,
			Want: testResult{
				RunLen: 10,
				Code:   WTerms[10],
			},
		},
		{
			Code: 16384,
			Want: testResult{
				RunLen: 11,
				Code:   WTerms[11],
			},
		},
		{
			Code: 8192,
			Want: testResult{
				RunLen: 12,
				Code:   WTerms[12],
			},
		},
		{
			Code: 3072,
			Want: testResult{
				RunLen: 13,
				Code:   WTerms[13],
			},
		},
		{
			Code: 53248,
			Want: testResult{
				RunLen: 14,
				Code:   WTerms[14],
			},
		},
		{
			Code: 54272,
			Want: testResult{
				RunLen: 15,
				Code:   WTerms[15],
			},
		},
		{
			Code: 43008,
			Want: testResult{
				RunLen: 16,
				Code:   WTerms[16],
			},
		},
		{
			Code: 44032,
			Want: testResult{
				RunLen: 17,
				Code:   WTerms[17],
			},
		},
		{
			Code: 19968,
			Want: testResult{
				RunLen: 18,
				Code:   WTerms[18],
			},
		},
		{
			Code: 6144,
			Want: testResult{
				RunLen: 19,
				Code:   WTerms[19],
			},
		},
		{
			Code: 4096,
			Want: testResult{
				RunLen: 20,
				Code:   WTerms[20],
			},
		},
		{
			Code: 11776,
			Want: testResult{
				RunLen: 21,
				Code:   WTerms[21],
			},
		},
		{
			Code: 1536,
			Want: testResult{
				RunLen: 22,
				Code:   WTerms[22],
			},
		},
	}
}*/
