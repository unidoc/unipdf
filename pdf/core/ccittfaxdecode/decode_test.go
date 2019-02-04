package ccittfaxdecode

import "testing"

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
