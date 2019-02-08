package ccittfaxdecode

import "testing"

/*func TestDecodeRow1D(t *testing.T) {
	type testResult struct {
		PixelsRow []byte
		BitPos    int
	}

	tests := []struct {
		Encoder *Encoder
		Encoded []byte
		BitPos  int
	}{
		{
			Encoder: &Encoder{}
		}
	}
}*/

func TestTryFetchEOFB(t *testing.T) {
	type testResult struct {
		OK     bool
		BitPos int
		Err    error
	}

	tests := []struct {
		Encoded []byte
		BitPos  int
		Want    testResult
	}{
		{
			Encoded: []byte{0, 16, 1},
			BitPos:  0,
			Want: testResult{
				OK:     true,
				BitPos: 24,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 16, 1, 0},
			BitPos:  0,
			Want: testResult{
				OK:     true,
				BitPos: 24,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 4, 0, 64},
			BitPos:  2,
			Want: testResult{
				OK:     true,
				BitPos: 26,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 4, 0, 64, 0},
			BitPos:  2,
			Want: testResult{
				OK:     true,
				BitPos: 26,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 16, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    ErrEOFBCorrupt,
			},
		},
		{
			Encoded: []byte{0, 1, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 1, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    nil,
			},
		},
	}

	for _, test := range tests {
		gotOK, gotBitPos, gotErr := tryFetchEOFB(test.Encoded, test.BitPos)

		if gotOK != test.Want.OK {
			t.Errorf("Wrong ok. Got %v, want %v\n", gotOK, test.Want.OK)
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}

		if gotErr != test.Want.Err {
			t.Errorf("Wrong err. Got %v, want %v\n", gotErr, test.Want.Err)
		}
	}
}

func TestTryFetchRTC2D(t *testing.T) {
	type testResult struct {
		OK     bool
		BitPos int
		Err    error
	}

	tests := []struct {
		Encoded []byte
		BitPos  int
		Want    testResult
	}{
		{
			Encoded: []byte{0, 24, 0, 192, 6, 0, 48, 1, 128, 12},
			BitPos:  0,
			Want: testResult{
				OK:     true,
				BitPos: 78,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 1, 128, 12, 0, 96, 3},
			BitPos:  2,
			Want: testResult{
				OK:     true,
				BitPos: 80,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 192, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 192, 6, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 1, 128, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 192, 6, 0, 48, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 1, 128, 12, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 192, 6, 0, 48, 1, 128, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 1, 128, 12, 0, 96, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    ErrRTCCorrupt,
			},
		},
		{
			Encoded: []byte{},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    nil,
			},
		},
		{
			Encoded: []byte{},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    nil,
			},
		},
		{
			Encoded: nil,
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    nil,
			},
		},
		{
			Encoded: nil,
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    nil,
			},
		},
	}

	for _, test := range tests {
		gotOK, gotBitPos, gotErr := tryFetchRTC2D(test.Encoded, test.BitPos)

		if gotOK != test.Want.OK {
			t.Errorf("Wrong ok. Got %v, want %v\n", gotOK, test.Want.OK)
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}

		if gotErr != test.Want.Err {
			t.Errorf("Wrong err. Got %v, want %v\n", gotErr, test.Want.Err)
		}
	}
}

func TestFetchNext2DCode(t *testing.T) {
	type testResult struct {
		Code   Code
		BitPos int
		OK     bool
	}

	tests := []struct {
		Data   []byte
		BitPos int
		Want   testResult
	}{
		{
			Data:   []byte{16},
			BitPos: 0,
			Want: testResult{
				Code:   P,
				BitPos: 4,
				OK:     true,
			},
		},
		{
			Data:   []byte{4},
			BitPos: 2,
			Want: testResult{
				Code:   P,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{16, 0},
			BitPos: 0,
			Want: testResult{
				Code:   P,
				BitPos: 4,
				OK:     true,
			},
		},
		{
			Data:   []byte{4, 0},
			BitPos: 2,
			Want: testResult{
				Code:   P,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{32},
			BitPos: 0,
			Want: testResult{
				Code:   H,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{8},
			BitPos: 2,
			Want: testResult{
				Code:   H,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{32, 0},
			BitPos: 0,
			Want: testResult{
				Code:   H,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{8, 0},
			BitPos: 2,
			Want: testResult{
				Code:   H,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{128},
			BitPos: 0,
			Want: testResult{
				Code:   V0,
				BitPos: 1,
				OK:     true,
			},
		},
		{
			Data:   []byte{32},
			BitPos: 2,
			Want: testResult{
				Code:   V0,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{128, 0},
			BitPos: 0,
			Want: testResult{
				Code:   V0,
				BitPos: 1,
				OK:     true,
			},
		},
		{
			Data:   []byte{32, 0},
			BitPos: 2,
			Want: testResult{
				Code:   V0,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{96},
			BitPos: 0,
			Want: testResult{
				Code:   V1R,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{24},
			BitPos: 2,
			Want: testResult{
				Code:   V1R,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{96, 0},
			BitPos: 0,
			Want: testResult{
				Code:   V1R,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{24, 0},
			BitPos: 2,
			Want: testResult{
				Code:   V1R,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{12},
			BitPos: 0,
			Want: testResult{
				Code:   V2R,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{3},
			BitPos: 2,
			Want: testResult{
				Code:   V2R,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{12, 0},
			BitPos: 0,
			Want: testResult{
				Code:   V2R,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{3, 0},
			BitPos: 2,
			Want: testResult{
				Code:   V2R,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{6},
			BitPos: 0,
			Want: testResult{
				Code:   V3R,
				BitPos: 7,
				OK:     true,
			},
		},
		{
			Data:   []byte{1, 128},
			BitPos: 2,
			Want: testResult{
				Code:   V3R,
				BitPos: 9,
				OK:     true,
			},
		},
		{
			Data:   []byte{6, 0},
			BitPos: 0,
			Want: testResult{
				Code:   V3R,
				BitPos: 7,
				OK:     true,
			},
		},
		{
			Data:   []byte{1, 128, 0},
			BitPos: 2,
			Want: testResult{
				Code:   V3R,
				BitPos: 9,
				OK:     true,
			},
		},
		{
			Data:   []byte{64},
			BitPos: 0,
			Want: testResult{
				Code:   V1L,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{16},
			BitPos: 2,
			Want: testResult{
				Code:   V1L,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{64, 0},
			BitPos: 0,
			Want: testResult{
				Code:   V1L,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{16, 0},
			BitPos: 2,
			Want: testResult{
				Code:   V1L,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{8},
			BitPos: 0,
			Want: testResult{
				Code:   V2L,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{2},
			BitPos: 2,
			Want: testResult{
				Code:   V2L,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{8, 0},
			BitPos: 0,
			Want: testResult{
				Code:   V2L,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{2, 0},
			BitPos: 2,
			Want: testResult{
				Code:   V2L,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{4},
			BitPos: 0,
			Want: testResult{
				Code:   V3L,
				BitPos: 7,
				OK:     true,
			},
		},
		{
			Data:   []byte{1, 0},
			BitPos: 2,
			Want: testResult{
				Code:   V3L,
				BitPos: 9,
				OK:     true,
			},
		},
		{
			Data:   []byte{0, 0, 0},
			BitPos: 0,
			Want: testResult{
				Code:   Code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   []byte{0, 0, 0},
			BitPos: 2,
			Want: testResult{
				Code:   Code{},
				BitPos: 2,
				OK:     false,
			},
		},
		{
			Data:   []byte{0},
			BitPos: 0,
			Want: testResult{
				Code:   Code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   []byte{0},
			BitPos: 2,
			Want: testResult{
				Code:   Code{},
				BitPos: 2,
				OK:     false,
			},
		},
		{
			Data:   []byte{},
			BitPos: 0,
			Want: testResult{
				Code:   Code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   []byte{},
			BitPos: 2,
			Want: testResult{
				Code:   Code{},
				BitPos: 2,
				OK:     false,
			},
		},
		{
			Data:   nil,
			BitPos: 0,
			Want: testResult{
				Code:   Code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   nil,
			BitPos: 2,
			Want: testResult{
				Code:   Code{},
				BitPos: 2,
				OK:     false,
			},
		},
	}

	for _, test := range tests {
		gotCode, gotBitPos, gotOK := fetchNext2DCode(test.Data, test.BitPos)

		if gotCode != test.Want.Code {
			t.Errorf("Wrong code. Got %v, want %v\n", gotCode, test.Want.Code)
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}

		if gotOK != test.Want.OK {
			t.Errorf("Wrong ok. Got %v, want %v\n", gotOK, test.Want.OK)
		}
	}
}

func TestGet2DCodeFromUint16(t *testing.T) {
	type testResult struct {
		Code Code
		OK   bool
	}

	tests := []struct {
		Codes  []uint16
		BitPos int
		Want   testResult
	}{
		{
			Codes:  formTestCodes(4096, 4),
			BitPos: 0,
			Want: testResult{
				Code: P,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(8192, 3),
			BitPos: 0,
			Want: testResult{
				Code: H,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(32768, 1),
			BitPos: 0,
			Want: testResult{
				Code: V0,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(24576, 3),
			BitPos: 0,
			Want: testResult{
				Code: V1R,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(3072, 6),
			BitPos: 0,
			Want: testResult{
				Code: V2R,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(1536, 7),
			BitPos: 0,
			Want: testResult{
				Code: V3R,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(16384, 3),
			BitPos: 0,
			Want: testResult{
				Code: V1L,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(2048, 6),
			BitPos: 0,
			Want: testResult{
				Code: V2L,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(1024, 7),
			BitPos: 0,
			Want: testResult{
				Code: V3L,
				OK:   true,
			},
		},
		{
			Codes:  []uint16{0, 0, 0},
			BitPos: 0,
			Want: testResult{
				Code: Code{},
				OK:   false,
			},
		},
	}

	for _, test := range tests {
		for _, code := range test.Codes {
			gotCode, gotOK := get2DCodeFromUint16(code, test.BitPos)

			if gotCode != test.Want.Code {
				t.Errorf("Wrong code. Got %v, want %v\n", gotCode, test.Want.Code)
			}

			if gotOK != test.Want.OK {
				t.Errorf("Wrong ok. Got %v, want %v\n", gotOK, test.Want.OK)
			}
		}
	}
}

func TestFetchNextRunLen(t *testing.T) {
	type testResult struct {
		RunLen int
		BitPos int
	}

	tests := []struct {
		Data    []byte
		BitPos  int
		IsWhite bool
		Want    testResult
	}{
		{
			Data:    []byte{0, 0, 0},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: -1,
				BitPos: 0,
			},
		},
		{
			Data:    []byte{0, 0, 0},
			BitPos:  2,
			IsWhite: true,
			Want: testResult{
				RunLen: -1,
				BitPos: 2,
			},
		},
		{
			Data:    []byte{0, 0, 0},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: -1,
				BitPos: 0,
			},
		},
		{
			Data:    []byte{0, 0, 0},
			BitPos:  2,
			IsWhite: false,
			Want: testResult{
				RunLen: -1,
				BitPos: 2,
			},
		},
		{
			Data:    []byte{},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: -1,
				BitPos: 0,
			},
		},
		{
			Data:    []byte{},
			BitPos:  2,
			IsWhite: true,
			Want: testResult{
				RunLen: -1,
				BitPos: 2,
			},
		},
		{
			Data:    []byte{},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: -1,
				BitPos: 0,
			},
		},
		{
			Data:    []byte{},
			BitPos:  2,
			IsWhite: false,
			Want: testResult{
				RunLen: -1,
				BitPos: 2,
			},
		},
		{
			Data:    nil,
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: -1,
				BitPos: 0,
			},
		},
		{
			Data:    nil,
			BitPos:  2,
			IsWhite: true,
			Want: testResult{
				RunLen: -1,
				BitPos: 2,
			},
		},
		{
			Data:    nil,
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: -1,
				BitPos: 0,
			},
		},
		{
			Data:    nil,
			BitPos:  2,
			IsWhite: false,
			Want: testResult{
				RunLen: -1,
				BitPos: 2,
			},
		},
		{
			Data:    []byte{0, 29, 255},
			BitPos:  4,
			IsWhite: true,
			Want: testResult{
				RunLen: 2432,
				BitPos: 16,
			},
		},
		{
			Data:    []byte{0, 29, 255},
			BitPos:  4,
			IsWhite: false,
			Want: testResult{
				RunLen: 2432,
				BitPos: 16,
			},
		},
		{
			Data:    []byte{0, 29, 0},
			BitPos:  4,
			IsWhite: true,
			Want: testResult{
				RunLen: 2432,
				BitPos: 16,
			},
		},
		{
			Data:    []byte{0, 29, 0},
			BitPos:  4,
			IsWhite: false,
			Want: testResult{
				RunLen: 2432,
				BitPos: 16,
			},
		},
		{
			Data:    []byte{1, 223, 255},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: 2432,
				BitPos: 12,
			},
		},
		{
			Data:    []byte{1, 223, 255},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: 2432,
				BitPos: 12,
			},
		},
		{
			Data:    []byte{1, 223, 0},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: 2432,
				BitPos: 12,
			},
		},
		{
			Data:    []byte{1, 223, 0},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: 2432,
				BitPos: 12,
			},
		},
		{
			Data:    []byte{168},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: 16,
				BitPos: 6,
			},
		},
		{
			Data:    []byte{168, 0},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: 16,
				BitPos: 6,
			},
		},
		{
			Data:    []byte{84},
			BitPos:  1,
			IsWhite: true,
			Want: testResult{
				RunLen: 16,
				BitPos: 7,
			},
		},
		{
			Data:    []byte{12, 192},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: 28,
				BitPos: 12,
			},
		},
		{
			Data:    []byte{12, 192, 0},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: 28,
				BitPos: 12,
			},
		},
		{
			Data:    []byte{3, 48},
			BitPos:  2,
			IsWhite: false,
			Want: testResult{
				RunLen: 28,
				BitPos: 14,
			},
		},
		{
			Data:    []byte{3, 48, 0},
			BitPos:  2,
			IsWhite: false,
			Want: testResult{
				RunLen: 28,
				BitPos: 14,
			},
		},
		{
			Data:    []byte{107, 128},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: 1152,
				BitPos: 9,
			},
		},
		{
			Data:    []byte{107, 128, 0},
			BitPos:  0,
			IsWhite: true,
			Want: testResult{
				RunLen: 1152,
				BitPos: 9,
			},
		},
		{
			Data:    []byte{26, 224},
			BitPos:  2,
			IsWhite: true,
			Want: testResult{
				RunLen: 1152,
				BitPos: 11,
			},
		},
		{
			Data:    []byte{26, 224, 0},
			BitPos:  2,
			IsWhite: true,
			Want: testResult{
				RunLen: 1152,
				BitPos: 11,
			},
		},
		{
			Data:    []byte{2, 104},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: 832,
				BitPos: 13,
			},
		},
		{
			Data:    []byte{2, 104, 0},
			BitPos:  0,
			IsWhite: false,
			Want: testResult{
				RunLen: 832,
				BitPos: 13,
			},
		},
		{
			Data:    []byte{0, 154},
			BitPos:  2,
			IsWhite: false,
			Want: testResult{
				RunLen: 832,
				BitPos: 15,
			},
		},
		{
			Data:    []byte{0, 154, 0},
			BitPos:  2,
			IsWhite: false,
			Want: testResult{
				RunLen: 832,
				BitPos: 15,
			},
		},
	}

	for _, test := range tests {
		gotRunLen, gotBitPos := fetchNextRunLen(test.Data, test.BitPos, test.IsWhite)

		if gotRunLen != test.Want.RunLen {
			t.Errorf("Wrong run len. Got %v, want %v\n", gotRunLen, test.Want.RunLen)
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}
	}
}

func TestDrawPixels(t *testing.T) {
	tests := []struct {
		Row     []byte
		IsWhite bool
		Length  int
		Want    []byte
	}{
		{
			Row:     nil,
			IsWhite: true,
			Length:  2,
			Want:    []byte{white, white},
		},
		{
			Row:     nil,
			IsWhite: false,
			Length:  2,
			Want:    []byte{black, black},
		},
		{
			Row:     nil,
			IsWhite: true,
			Length:  3,
			Want:    []byte{white, white, white},
		},
		{
			Row:     nil,
			IsWhite: false,
			Length:  3,
			Want:    []byte{black, black, black},
		},
		{
			Row:     nil,
			IsWhite: true,
			Length:  0,
			Want:    nil,
		},
		{
			Row:     nil,
			IsWhite: false,
			Length:  0,
			Want:    nil,
		},
		{
			Row:     nil,
			IsWhite: true,
			Length:  -1,
			Want:    nil,
		},
		{
			Row:     nil,
			IsWhite: false,
			Length:  -1,
			Want:    nil,
		},
		{
			Row:     []byte{},
			IsWhite: true,
			Length:  2,
			Want:    []byte{white, white},
		},
		{
			Row:     []byte{},
			IsWhite: false,
			Length:  2,
			Want:    []byte{black, black},
		},
		{
			Row:     []byte{},
			IsWhite: true,
			Length:  3,
			Want:    []byte{white, white, white},
		},
		{
			Row:     []byte{},
			IsWhite: false,
			Length:  3,
			Want:    []byte{black, black, black},
		},
		{
			Row:     []byte{},
			IsWhite: true,
			Length:  0,
			Want:    []byte{},
		},
		{
			Row:     []byte{},
			IsWhite: false,
			Length:  0,
			Want:    []byte{},
		},
		{
			Row:     []byte{},
			IsWhite: true,
			Length:  -1,
			Want:    []byte{},
		},
		{
			Row:     []byte{},
			IsWhite: false,
			Length:  -1,
			Want:    []byte{},
		},
		{
			Row:     []byte{white, white},
			IsWhite: true,
			Length:  2,
			Want:    []byte{white, white, white, white},
		},
		{
			Row:     []byte{white, white},
			IsWhite: false,
			Length:  2,
			Want:    []byte{white, white, black, black},
		},
		{
			Row:     []byte{white, white, black},
			IsWhite: true,
			Length:  3,
			Want:    []byte{white, white, black, white, white, white},
		},
		{
			Row:     []byte{black, black, white},
			IsWhite: false,
			Length:  3,
			Want:    []byte{black, black, white, black, black, black},
		},
		{
			Row:     []byte{white, black},
			IsWhite: true,
			Length:  0,
			Want:    []byte{white, black},
		},
		{
			Row:     []byte{white, black},
			IsWhite: false,
			Length:  0,
			Want:    []byte{white, black},
		},
		{
			Row:     []byte{black, white},
			IsWhite: true,
			Length:  -1,
			Want:    []byte{black, white},
		},
		{
			Row:     []byte{black, white},
			IsWhite: false,
			Length:  -1,
			Want:    []byte{black, white},
		},
	}

	for _, test := range tests {
		got := drawPixels(test.Row, test.IsWhite, test.Length)

		if len(got) != len(test.Want) {
			t.Errorf("Wrong len. Got %v, want %v\n", len(got), len(test.Want))
		} else {
			for i := range got {
				if got[i] != test.Want[i] {
					t.Errorf("Wron pixel at %v. Got %v, want %v\n", i, got[i], test.Want[i])
				}
			}
		}
	}
}

func TestTryFetchEOL(t *testing.T) {
	type testResult struct {
		OK     bool
		BitPos int
	}

	tests := []struct {
		Encoded []byte
		BitPos  int
		Want    testResult
	}{
		{
			Encoded: []byte{0, 16},
			BitPos:  0,
			Want: testResult{
				OK:     true,
				BitPos: 12,
			},
		},
		{
			Encoded: []byte{0, 16},
			BitPos:  1,
			Want: testResult{
				OK:     false,
				BitPos: 1,
			},
		},
		{
			Encoded: []byte{128, 8},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: []byte{128, 8},
			BitPos:  1,
			Want: testResult{
				OK:     true,
				BitPos: 13,
			},
		},
		{
			Encoded: []byte{0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: []byte{},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: nil,
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
	}

	for _, test := range tests {
		gotOK, gotBitPos := tryFetchEOL(test.Encoded, test.BitPos)

		if gotOK != test.Want.OK {
			t.Errorf("Wrong ok. Got %v, want %v\n", gotOK, test.Want.OK)
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}
	}
}

func TestTryFetchEOL0(t *testing.T) {
	type testResult struct {
		OK     bool
		BitPos int
	}

	tests := []struct {
		Encoded []byte
		BitPos  int
		Want    testResult
	}{
		{
			Encoded: []byte{0, 16},
			BitPos:  0,
			Want: testResult{
				OK:     true,
				BitPos: 13,
			},
		},
		{
			Encoded: []byte{0, 16},
			BitPos:  1,
			Want: testResult{
				OK:     false,
				BitPos: 1,
			},
		},
		{
			Encoded: []byte{128, 8},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: []byte{128, 8},
			BitPos:  1,
			Want: testResult{
				OK:     true,
				BitPos: 14,
			},
		},
		{
			Encoded: []byte{0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: []byte{},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: nil,
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
	}

	for _, test := range tests {
		gotOK, gotBitPos := tryFetchEOL0(test.Encoded, test.BitPos)

		if gotOK != test.Want.OK {
			t.Errorf("Wrong ok. Got %v, want %v\n", gotOK, test.Want.OK)
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}
	}
}

func TestTryFetchEOL1(t *testing.T) {
	type testResult struct {
		OK     bool
		BitPos int
	}

	tests := []struct {
		Encoded []byte
		BitPos  int
		Want    testResult
	}{
		{
			Encoded: []byte{0, 24},
			BitPos:  0,
			Want: testResult{
				OK:     true,
				BitPos: 13,
			},
		},
		{
			Encoded: []byte{0, 24},
			BitPos:  1,
			Want: testResult{
				OK:     false,
				BitPos: 1,
			},
		},
		{
			Encoded: []byte{128, 12},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: []byte{128, 12},
			BitPos:  1,
			Want: testResult{
				OK:     true,
				BitPos: 14,
			},
		},
		{
			Encoded: []byte{0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: []byte{},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
		{
			Encoded: nil,
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
			},
		},
	}

	for _, test := range tests {
		gotOK, gotBitPos := tryFetchEOL1(test.Encoded, test.BitPos)

		if gotOK != test.Want.OK {
			t.Errorf("Wrong ok. Got %v, want %v\n", gotOK, test.Want.OK)
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}
	}
}

func TestBitFromUint16(t *testing.T) {
	type testData struct {
		Num    uint16
		BitPos int
		Want   byte
	}
	var tests []testData

	var num uint16 = 43690
	for i := 0; i < 16; i++ {
		var want byte = 1
		if (i % 2) != 0 {
			want = 0
		}

		tests = append(tests, testData{
			Num:    num,
			BitPos: i,
			Want:   want,
		})
	}

	for _, test := range tests {
		bit := bitFromUint16(test.Num, test.BitPos)

		if bit != test.Want {
			t.Errorf("Wrong bit. Got %v want %v\n", bit, test.Want)
		}
	}
}

func TestFetchNextCode(t *testing.T) {
	testData := []byte{95, 21, 197}

	type testResult struct {
		Code       uint16
		CodeBitPos int
		DataBitPos int
	}

	tests := []struct {
		Data   []byte
		BitPos int
		Want   testResult
	}{
		{
			Data:   testData,
			BitPos: 0,
			Want: testResult{
				Code:       24341,
				CodeBitPos: 0,
				DataBitPos: 16,
			},
		},
		{
			Data:   testData,
			BitPos: 1,
			Want: testResult{
				Code:       48683,
				CodeBitPos: 0,
				DataBitPos: 17,
			},
		},
		{
			Data:   testData,
			BitPos: 2,
			Want: testResult{
				Code:       31831,
				CodeBitPos: 0,
				DataBitPos: 18,
			},
		},
		{
			Data:   testData,
			BitPos: 3,
			Want: testResult{
				Code:       63662,
				CodeBitPos: 0,
				DataBitPos: 19,
			},
		},
		{
			Data:   testData,
			BitPos: 4,
			Want: testResult{
				Code:       61788,
				CodeBitPos: 0,
				DataBitPos: 20,
			},
		},
		{
			Data:   testData,
			BitPos: 5,
			Want: testResult{
				Code:       58040,
				CodeBitPos: 0,
				DataBitPos: 21,
			},
		},
		{
			Data:   testData,
			BitPos: 6,
			Want: testResult{
				Code:       50545,
				CodeBitPos: 0,
				DataBitPos: 22,
			},
		},
		{
			Data:   testData,
			BitPos: 7,
			Want: testResult{
				Code:       35554,
				CodeBitPos: 0,
				DataBitPos: 23,
			},
		},
		{
			Data:   testData,
			BitPos: 8,
			Want: testResult{
				Code:       5573,
				CodeBitPos: 0,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 9,
			Want: testResult{
				Code:       5573,
				CodeBitPos: 1,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 10,
			Want: testResult{
				Code:       5573,
				CodeBitPos: 2,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 11,
			Want: testResult{
				Code:       5573,
				CodeBitPos: 3,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 12,
			Want: testResult{
				Code:       1477,
				CodeBitPos: 4,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 13,
			Want: testResult{
				Code:       1477,
				CodeBitPos: 5,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 14,
			Want: testResult{
				Code:       453,
				CodeBitPos: 6,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 15,
			Want: testResult{
				Code:       453,
				CodeBitPos: 7,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 16,
			Want: testResult{
				Code:       197,
				CodeBitPos: 8,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 17,
			Want: testResult{
				Code:       69,
				CodeBitPos: 9,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 18,
			Want: testResult{
				Code:       5,
				CodeBitPos: 10,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 19,
			Want: testResult{
				Code:       5,
				CodeBitPos: 11,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 20,
			Want: testResult{
				Code:       5,
				CodeBitPos: 12,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 21,
			Want: testResult{
				Code:       5,
				CodeBitPos: 13,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 22,
			Want: testResult{
				Code:       1,
				CodeBitPos: 14,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 23,
			Want: testResult{
				Code:       1,
				CodeBitPos: 15,
				DataBitPos: 24,
			},
		},
		{
			Data:   testData,
			BitPos: 24,
			Want: testResult{
				Code:       0,
				CodeBitPos: 16,
				DataBitPos: 24,
			},
		},
	}

	for _, test := range tests {
		gotCode, gotCodeBitPos, gotDataBitPos := fetchNextCode(test.Data, test.BitPos)

		if gotCode != test.Want.Code {
			t.Errorf("Wrong code. Got %v, want %v\n", gotCode, test.Want.Code)
		}

		if gotCodeBitPos != test.Want.CodeBitPos {
			t.Errorf("Wrong code bit pos. Got %v, want %v\n", gotCodeBitPos, test.Want.CodeBitPos)
		}

		if gotDataBitPos != test.Want.DataBitPos {
			t.Errorf("Wrong data bit pos. Got %v, want %v\n", gotDataBitPos, test.Want.DataBitPos)
		}
	}
}
