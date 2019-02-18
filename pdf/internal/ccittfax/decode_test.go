/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ccittfax

import (
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"testing"
)

const (
	testDataPath = "./testdata/"
)

func TestDecodeNextRunLen(t *testing.T) {
	type testResult struct {
		PixelsRow []byte
		BitPos    int
		Err       error
	}

	type testData struct {
		Encoded   []byte
		PixelsRow []byte
		BitPos    int
		IsWhite   bool
		Want      testResult
	}

	// Calculated test data based on the reference documentation.
	tests := []testData{
		{
			Encoded:   []byte{1, 28, 2},
			PixelsRow: []byte{},
			BitPos:    16,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white},
				BitPos: 24,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 2, 0, 0, 0},
			PixelsRow: []byte{},
			BitPos:    16,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white},
				BitPos: 24,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 56},
			PixelsRow: []byte{},
			BitPos:    21,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: []byte{black, black, black, black, black, black, black, black, black, black, black,
					black, black, black},
				BitPos: 29,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 56, 0, 0, 0},
			PixelsRow: []byte{},
			BitPos:    21,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: []byte{black, black, black, black, black, black, black, black, black, black, black,
					black, black, black},
				BitPos: 29,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 154, 80},
			PixelsRow: []byte{},
			BitPos:    23,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: drawPixels(nil, true, 1624),
				BitPos:    39,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 154, 80, 0, 0},
			PixelsRow: []byte{},
			BitPos:    23,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: drawPixels(nil, true, 1624),
				BitPos:    39,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 29, 67, 128},
			PixelsRow: nil,
			BitPos:    21,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: drawPixels(nil, false, 1100),
				BitPos:    41,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 29, 67, 128, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    21,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: drawPixels(nil, false, 1100),
				BitPos:    41,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 9, 166, 188},
			PixelsRow: nil,
			BitPos:    21,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: drawPixels(nil, true, 3655),
				BitPos:    46,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 9, 166, 188, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    21,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: drawPixels(nil, true, 3655),
				BitPos:    46,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 15, 129, 200, 21, 128},
			PixelsRow: nil,
			BitPos:    21,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: drawPixels(nil, false, 3502),
				BitPos:    58,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 15, 129, 200, 21, 128, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    21,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: drawPixels(nil, false, 3502),
				BitPos:    58,
				Err:       nil,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   true,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   false,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
	}

	for _, test := range tests {
		gotPixelsRow, gotBitPos, gotErr := decodeNextRunLen(test.Encoded, test.PixelsRow, test.BitPos, test.IsWhite)

		if len(gotPixelsRow) != len(test.Want.PixelsRow) {
			t.Errorf("Wrong pixels row len. Got %v, want %v\n", len(gotPixelsRow), len(test.Want.PixelsRow))
		} else {
			for i := range gotPixelsRow {
				if gotPixelsRow[i] != test.Want.PixelsRow[i] {
					t.Errorf("Wrong pixel at %v. Got %v, want %v\n",
						i, gotPixelsRow[i], test.Want.PixelsRow[i])
				}
			}
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}

		if gotErr != test.Want.Err {
			t.Errorf("Wrong err. Got %v, want %v\n", gotErr, test.Want.Err)
		}
	}
}

func TestDecodeHorizontalMode(t *testing.T) {
	type testResult struct {
		PixelsRow []byte
		BitPos    int
		A0        int
		Err       error
	}

	type testData struct {
		Encoded   []byte
		PixelsRow []byte
		BitPos    int
		IsWhite   bool
		A0        int
		Want      testResult
	}

	tests := []testData{
		{
			Encoded:   []byte{1, 28, 2, 192},
			PixelsRow: []byte{},
			BitPos:    16,
			IsWhite:   true,
			A0:        0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white, black, black},
				BitPos: 26,
				A0:     31,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 2, 192, 0, 0, 0},
			PixelsRow: []byte{},
			BitPos:    16,
			IsWhite:   true,
			A0:        7,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white, black, black},
				BitPos: 26,
				A0:     31,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 56, 224},
			PixelsRow: []byte{},
			BitPos:    21,
			IsWhite:   false,
			A0:        0,
			Want: testResult{
				PixelsRow: []byte{black, black, black, black, black, black, black, black, black, black, black,
					black, black, black, white},
				BitPos: 35,
				A0:     15,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{1, 28, 40, 56, 224, 0, 0, 0},
			PixelsRow: []byte{},
			BitPos:    21,
			IsWhite:   false,
			A0:        10,
			Want: testResult{
				PixelsRow: []byte{black, black, black, black, black, black, black, black, black, black, black,
					black, black, black, white},
				BitPos: 35,
				A0:     15,
				Err:    nil,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   true,
			A0:        0,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				A0:        0,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   true,
			A0:        120,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				A0:        120,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   false,
			A0:        10,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				A0:        10,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   false,
			A0:        13,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				A0:        13,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   true,
			A0:        23,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				A0:        23,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   true,
			A0:        34,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				A0:        34,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   false,
			A0:        134,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				A0:        134,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{},
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   false,
			A0:        35,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				A0:        35,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   true,
			A0:        876,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				A0:        876,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   true,
			A0:        738,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				A0:        738,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    0,
			IsWhite:   false,
			A0:        283,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
				A0:        283,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   nil,
			PixelsRow: nil,
			BitPos:    2,
			IsWhite:   false,
			A0:        29,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
				A0:        29,
				Err:       errWrongCodeInHorizontalMode,
			},
		},
		{
			Encoded:   []byte{1, 28, 2, 0, 0, 0, 0, 0},
			PixelsRow: nil,
			BitPos:    16,
			IsWhite:   true,
			A0:        0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white, white, white, white,
					white, white, white, white, white, white, white, white, white, white, white, white, white, white,
					white, white, white},
				BitPos: 16,
				A0:     0,
				Err:    errWrongCodeInHorizontalMode,
			},
		},
	}

	resultingPixelsRow1 := drawPixels(nil, true, 1624)
	resultingPixelsRow1 = drawPixels(resultingPixelsRow1, false, 1728)

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 154, 80, 6, 80, 220},
		PixelsRow: []byte{},
		BitPos:    23,
		IsWhite:   true,
		A0:        0,
		Want: testResult{
			PixelsRow: resultingPixelsRow1,
			BitPos:    62,
			A0:        3352,
			Err:       nil,
		},
	})

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 154, 80, 6, 80, 220, 0, 0, 0},
		PixelsRow: []byte{},
		BitPos:    23,
		IsWhite:   true,
		A0:        112,
		Want: testResult{
			PixelsRow: resultingPixelsRow1,
			BitPos:    62,
			A0:        3352,
			Err:       nil,
		},
	})

	resultingPixelsRow2 := drawPixels(nil, false, 1100)
	resultingPixelsRow2 = drawPixels(resultingPixelsRow2, true, 1027)

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 29, 67, 181, 96},
		PixelsRow: nil,
		BitPos:    21,
		IsWhite:   false,
		A0:        0,
		Want: testResult{
			PixelsRow: resultingPixelsRow2,
			BitPos:    54,
			A0:        2127,
			Err:       nil,
		},
	})

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 29, 67, 181, 96, 0, 0, 0},
		PixelsRow: nil,
		BitPos:    21,
		IsWhite:   false,
		A0:        109,
		Want: testResult{
			PixelsRow: resultingPixelsRow2,
			BitPos:    54,
			A0:        2127,
			Err:       nil,
		},
	})

	resultingPixelsRow3 := drawPixels(nil, true, 3655)
	resultingPixelsRow3 = drawPixels(resultingPixelsRow3, false, 3547)

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 9, 166, 188, 7, 192, 230, 25, 96},
		PixelsRow: nil,
		BitPos:    21,
		IsWhite:   true,
		A0:        0,
		Want: testResult{
			PixelsRow: resultingPixelsRow3,
			BitPos:    83,
			A0:        7202,
			Err:       nil,
		},
	})

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 9, 166, 188, 7, 192, 230, 25, 96, 0, 0, 0},
		PixelsRow: nil,
		BitPos:    21,
		IsWhite:   true,
		A0:        1121,
		Want: testResult{
			PixelsRow: resultingPixelsRow3,
			BitPos:    83,
			A0:        7202,
			Err:       nil,
		},
	})

	resultingPixelsRow4 := drawPixels(nil, false, 3502)
	resultingPixelsRow4 = drawPixels(resultingPixelsRow4, true, 3488)

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 15, 129, 200, 21, 128, 125, 166, 54},
		PixelsRow: nil,
		BitPos:    21,
		IsWhite:   false,
		A0:        0,
		Want: testResult{
			PixelsRow: resultingPixelsRow4,
			BitPos:    87,
			A0:        6990,
			Err:       nil,
		},
	})

	tests = append(tests, testData{
		Encoded:   []byte{1, 28, 40, 15, 129, 200, 21, 128, 125, 166, 54, 0, 0, 0},
		PixelsRow: nil,
		BitPos:    21,
		IsWhite:   false,
		A0:        2148,
		Want: testResult{
			PixelsRow: resultingPixelsRow4,
			BitPos:    87,
			A0:        6990,
			Err:       nil,
		},
	})

	for _, test := range tests {
		gotPixelsRow, gotBitPos, gotA0, gotErr := decodeHorizontalMode(test.Encoded, test.PixelsRow,
			test.BitPos, test.IsWhite, test.A0)

		if len(gotPixelsRow) != len(test.Want.PixelsRow) {
			t.Errorf("Wrong pixels row len. Got %v, want %v\n", len(gotPixelsRow), len(test.Want.PixelsRow))
		} else {
			for i := range gotPixelsRow {
				if gotPixelsRow[i] != test.Want.PixelsRow[i] {
					t.Errorf("Wrong pixel at %v. Got %v, want %v\n",
						i, gotPixelsRow[i], test.Want.PixelsRow[i])
				}
			}
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}

		if gotA0 != test.Want.A0 {
			t.Errorf("Wrong A0. Got %v, want %v\n", gotA0, test.Want.A0)
		}

		if gotErr != test.Want.Err {
			t.Errorf("Wrong err. Got %v, want %v\n", gotErr, test.Want.Err)
		}
	}
}

func TestDecodePassMode(t *testing.T) {
	type testResult struct {
		PixelsRow []byte
		A0        int
	}

	type testData struct {
		Pixels    [][]byte
		PixelsRow []byte
		IsWhite   bool
		A0        int
		Want      testResult
	}

	tests := []testData{
		{
			Pixels: [][]byte{
				{white, white, white, white, white},
			},
			PixelsRow: nil,
			IsWhite:   true,
			A0:        -1,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white},
				A0:        5,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black},
				{white, white, white, white, white},
			},
			PixelsRow: []byte{black},
			IsWhite:   true,
			A0:        1,
			Want: testResult{
				PixelsRow: []byte{black, white, white, white, white},
				A0:        5,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black},
				{white, white, white, white, white},
			},
			PixelsRow: []byte{},
			IsWhite:   false,
			A0:        0,
			Want: testResult{
				PixelsRow: []byte{black, black, black, black, black},
				A0:        5,
			},
		},
	}

	for _, test := range tests {
		gotPixelsRow, gotA0 := decodePassMode(test.Pixels, test.PixelsRow, test.IsWhite, test.A0)

		if len(gotPixelsRow) != len(test.Want.PixelsRow) {
			t.Errorf("Wrong pixels row len. Got %v, want %v\n", len(gotPixelsRow), len(test.Want.PixelsRow))
		} else {
			for i := range gotPixelsRow {
				if gotPixelsRow[i] != test.Want.PixelsRow[i] {
					t.Errorf("Wrong pixel at %v. Got %v, want %v\n",
						i, gotPixelsRow[i], test.Want.PixelsRow[i])
				}
			}
		}

		if gotA0 != test.Want.A0 {
			t.Errorf("Wrong a0. Got %v, want %v\n", gotA0, test.Want.A0)
		}
	}
}

func TestDecode(t *testing.T) {
	type testResult struct {
		Pixels [][]byte
		Err    error
	}

	type testData struct {
		Encoder       *Encoder
		InputFilePath string
		Want          testResult
	}

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	wantPixels, err := getPixels(testDataPath + "p3_0.png")
	if err != nil {
		t.Fatalf("Error reading image: %v\n", err)
	}

	tests := []testData{
		{
			Encoder: &Encoder{
				K:       0,
				Columns: 2560,
				Rows:    3295,
			},
			InputFilePath: testDataPath + "K0-Columns2560-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:         0,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			InputFilePath: testDataPath + "K0-Columns2560-EOL-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:          0,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "K0-Columns2560-EOL-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "K0-Columns2560-EOL-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "K0-Columns2560-EOL-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:          0,
				Columns:    2560,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "K0-Columns2560-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "K0-Columns2560-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "K0-Columns2560-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:       4,
				Columns: 2560,
				Rows:    3295,
			},
			InputFilePath: testDataPath + "K4-Columns2560-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:         4,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			InputFilePath: testDataPath + "K4-Columns2560-EOL-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:          4,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "K4-Columns2560-EOL-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "K4-Columns2560-EOL-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "K4-Columns2560-EOL-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:          4,
				Columns:    2560,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "K4-Columns2560-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "K4-Columns2560-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "K4-Columns2560-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:       -1,
				Columns: 2560,
				Rows:    3295,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:         -1,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-EOL-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:          -1,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-EOL-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                -1,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-EOL-Aligned-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                -1,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-EOL-Aligned-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:          -1,
				Columns:    2560,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                -1,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-Aligned-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				K:                -1,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "K-1-Columns2560-Aligned-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1: true,
				K:        0,
				Columns:  2560,
				Rows:     3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:  true,
				K:         0,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          0,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                0,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                0,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          0,
				Columns:    2560,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "BlackIs1-K0-Columns2560-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1: true,
				K:        4,
				Columns:  2560,
				Rows:     3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:  true,
				K:         4,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          4,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                4,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                4,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          4,
				Columns:    2560,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-Aligned-Rows3295.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "BlackIs1-K4-Columns2560-Aligned-EOFB.gr3",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1: true,
				K:        -1,
				Columns:  2560,
				Rows:     3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:  true,
				K:         -1,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          -1,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                -1,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-Aligned-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                -1,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-Aligned-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          -1,
				Columns:    2560,
				EndOfBlock: true,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                -1,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-Aligned-Rows3295.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                -1,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			InputFilePath: testDataPath + "BlackIs1-K-1-Columns2560-Aligned-EOFB.gr4",
			Want: testResult{
				Pixels: wantPixels,
				Err:    nil,
			},
		},
	}

	for _, test := range tests {
		f, err := os.Open(test.InputFilePath)
		if err != nil {
			t.Fatalf("Error opening encoded file: %v\n", err)
		}

		encodedData, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatalf("Error reading encoded data from file: %v\n", err)
		}

		gotPixels, gotErr := test.Encoder.Decode(encodedData)
		if gotErr != test.Want.Err {
			t.Errorf("Wrong err. Got %v, want %v\n", gotErr, test.Want.Err)
		} else {
			if len(gotPixels) != len(test.Want.Pixels) {
				t.Errorf("Wrong pixels len. Got %v, want %v\n",
					len(gotPixels), len(test.Want.Pixels))
			} else {
				for i := range gotPixels {
					for j := range gotPixels[i] {
						if gotPixels[i][j] != test.Want.Pixels[i][j] {
							t.Errorf("Wrong pixel at %v:%v. Got %v, want %v\n",
								i, j, gotPixels[i][j], test.Want.Pixels[i][j])
						}
					}
				}
			}
		}
	}
}

func getPixels(imagePath string) ([][]byte, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	var pixels [][]byte
	for y := 0; y < h; y++ {
		var row []byte
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 65535 && g == 65535 && b == 65535 {
				// append white
				row = append(row, 1)
			} else {
				row = append(row, 0)
			}
		}

		pixels = append(pixels, row)
	}

	return pixels, nil
}

func TestDecodeVerticalMode(t *testing.T) {
	type testResult struct {
		PixelsRow []byte
		A0        int
	}

	type testData struct {
		Pixels    [][]byte
		PixelsRow []byte
		IsWhite   bool
		A0        int
		Shift     int
		Want      testResult
	}

	tests := []testData{
		{
			Pixels: [][]byte{
				{white, white, white, white, white},
			},
			PixelsRow: nil,
			IsWhite:   true,
			A0:        -1,
			Shift:     0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white},
				A0:        5,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black},
				{white, white, white, white, white},
			},
			PixelsRow: []byte{white},
			IsWhite:   false,
			A0:        1,
			Shift:     -1,
			Want: testResult{
				PixelsRow: []byte{white, black, black, black},
				A0:        4,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black},
				{white, white, white, white, white},
			},
			PixelsRow: []byte{},
			IsWhite:   true,
			A0:        -1,
			Shift:     -2,
			Want: testResult{
				PixelsRow: []byte{white, white, white},
				A0:        3,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black},
				{white, white, white, white, white},
			},
			PixelsRow: nil,
			IsWhite:   true,
			A0:        -1,
			Shift:     -3,
			Want: testResult{
				PixelsRow: []byte{white, white},
				A0:        2,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black, black, black, black},
				{white, black, white, black, black, black, black, black},
			},
			PixelsRow: []byte{black},
			IsWhite:   true,
			A0:        1,
			Shift:     -1,
			Want: testResult{
				PixelsRow: []byte{black, white},
				A0:        2,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black, black, black, black},
				{white, black, white, black, black, black, black, black},
			},
			PixelsRow: []byte{black},
			IsWhite:   true,
			A0:        1,
			Shift:     0,
			Want: testResult{
				PixelsRow: []byte{black, white, white},
				A0:        3,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black, black, black, black},
				{white, black, white, black, black, black, black, black},
			},
			PixelsRow: []byte{black},
			IsWhite:   true,
			A0:        1,
			Shift:     1,
			Want: testResult{
				PixelsRow: []byte{black, white, white, white},
				A0:        4,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black, black, black, black},
				{white, black, white, black, black, black, black, black},
			},
			PixelsRow: []byte{black},
			IsWhite:   true,
			A0:        1,
			Shift:     2,
			Want: testResult{
				PixelsRow: []byte{black, white, white, white, white},
				A0:        5,
			},
		},
		{
			Pixels: [][]byte{
				{black, black, black, black, black, black, black, black},
				{white, black, white, black, black, black, black, black},
			},
			PixelsRow: []byte{black},
			IsWhite:   true,
			A0:        1,
			Shift:     3,
			Want: testResult{
				PixelsRow: []byte{black, white, white, white, white, white},
				A0:        6,
			},
		},
	}

	for _, test := range tests {
		gotPixelsRow, gotA0 := decodeVerticalMode(test.Pixels, test.PixelsRow, test.IsWhite, test.A0, test.Shift)

		if len(gotPixelsRow) != len(test.Want.PixelsRow) {
			t.Errorf("Wrong pixels row len. Got %v, want %v\n",
				len(gotPixelsRow), len(test.Want.PixelsRow))
		} else {
			for i := range gotPixelsRow {
				if gotPixelsRow[i] != test.Want.PixelsRow[i] {
					t.Errorf("Wrong pixel at %v. Got %v, want %v\n",
						i, gotPixelsRow[i], test.Want.PixelsRow[i])
				}
			}
		}

		if gotA0 != test.Want.A0 {
			t.Errorf("Wrong a0. Got %v, want %v\n", gotA0, test.Want.A0)
		}
	}
}

func TestDecodeRow1D(t *testing.T) {
	type testResult struct {
		PixelsRow []byte
		BitPos    int
	}

	type testData struct {
		Encoder *Encoder
		Encoded []byte
		BitPos  int
		Want    testResult
	}

	tests := []testData{
		{
			Encoder: &Encoder{
				Columns: 17,
			},
			Encoded: []byte{160, 199},
			BitPos:  0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white},
				BitPos: 16,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 17,
			},
			Encoded: []byte{40, 49, 192},
			BitPos:  2,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white},
				BitPos: 18,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 17,
			},
			Encoded: []byte{40, 49, 192, 0},
			BitPos:  2,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white},
				BitPos: 18,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 17,
			},
			Encoded: []byte{160, 199, 0},
			BitPos:  0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white},
				BitPos: 16,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 17,
			},
			Encoded: []byte{160, 199, 7},
			BitPos:  0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white},
				BitPos: 16,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 17,
			},
			Encoded: []byte{40, 49, 193, 192},
			BitPos:  2,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white},
				BitPos: 18,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 31,
			},
			Encoded: []byte{40, 49, 193, 192},
			BitPos:  2,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white, black, black, black,
					black, black, black, black, black, black, black, black, black, black, black},
				BitPos: 26,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 31,
			},
			Encoded: []byte{160, 199, 7},
			BitPos:  0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white, black, black, black,
					black, black, black, black, black, black, black, black, black, black, black},
				BitPos: 24,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 31,
			},
			Encoded: []byte{160, 199, 7, 0},
			BitPos:  0,
			Want: testResult{
				PixelsRow: []byte{white, white, white, white, white, white, white, white, white,
					black, black, black, black, black, black, black, white, black, black, black,
					black, black, black, black, black, black, black, black, black, black, black},
				BitPos: 24,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 40,
			},
			Encoded: []byte{},
			BitPos:  0,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 40,
			},
			Encoded: []byte{},
			BitPos:  2,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 40,
			},
			Encoded: nil,
			BitPos:  0,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    0,
			},
		},
		{
			Encoder: &Encoder{
				Columns: 40,
			},
			Encoded: nil,
			BitPos:  2,
			Want: testResult{
				PixelsRow: nil,
				BitPos:    2,
			},
		},
	}

	resultsPixels := drawPixels(nil, true, 0)
	resultsPixels = drawPixels(resultsPixels, false, 2486)
	resultsPixels = drawPixels(resultsPixels, true, 6338)
	resultsPixels = drawPixels(resultsPixels, false, 1)

	tests = append(tests, testData{
		Encoder: &Encoder{
			Columns: 8825,
		},
		Encoded: []byte{53, 1, 208, 56, 1, 240, 31, 108, 58},
		BitPos:  0,
		Want: testResult{
			PixelsRow: resultsPixels,
			BitPos:    72,
		},
	})

	tests = append(tests, testData{
		Encoder: &Encoder{
			Columns: 8825,
		},
		Encoded: []byte{53, 1, 208, 56, 1, 240, 31, 108, 58, 0},
		BitPos:  0,
		Want: testResult{
			PixelsRow: resultsPixels,
			BitPos:    72,
		},
	})

	tests = append(tests, testData{
		Encoder: &Encoder{
			Columns: 8825,
		},
		Encoded: []byte{13, 64, 116, 14, 0, 124, 7, 219, 14, 128},
		BitPos:  2,
		Want: testResult{
			PixelsRow: resultsPixels,
			BitPos:    74,
		},
	})

	tests = append(tests, testData{
		Encoder: &Encoder{
			Columns: 8825,
		},
		Encoded: []byte{13, 64, 116, 14, 0, 124, 7, 219, 14, 128, 0},
		BitPos:  2,
		Want: testResult{
			PixelsRow: resultsPixels,
			BitPos:    74,
		},
	})

	for _, test := range tests {
		gotPixelsRow, gotBitPos := test.Encoder.decodeRow1D(test.Encoded, test.BitPos)

		if len(gotPixelsRow) != len(test.Want.PixelsRow) {
			t.Errorf("Wrong pixels row len. Got %v, want %v\n",
				len(gotPixelsRow), len(test.Want.PixelsRow))
		} else {
			for i := range gotPixelsRow {
				if gotPixelsRow[i] != test.Want.PixelsRow[i] {
					t.Errorf("Wrong value at %v. Got %v, want %v\n",
						i, gotPixelsRow[i], test.Want.PixelsRow[i])
				}
			}
		}

		if gotBitPos != test.Want.BitPos {
			t.Errorf("Wrong bit pos. Got %v, want %v\n", gotBitPos, test.Want.BitPos)
		}
	}
}

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
				Err:    errEOFBCorrupt,
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
				Err:    errRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    errRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 192, 6, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    errRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 1, 128, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    errRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 192, 6, 0, 48, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    errRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 1, 128, 12, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    errRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 24, 0, 192, 6, 0, 48, 1, 128, 0, 0, 0},
			BitPos:  0,
			Want: testResult{
				OK:     false,
				BitPos: 0,
				Err:    errRTCCorrupt,
			},
		},
		{
			Encoded: []byte{0, 6, 0, 48, 1, 128, 12, 0, 96, 0, 0, 0},
			BitPos:  2,
			Want: testResult{
				OK:     false,
				BitPos: 2,
				Err:    errRTCCorrupt,
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
		Code   code
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
				Code:   p,
				BitPos: 4,
				OK:     true,
			},
		},
		{
			Data:   []byte{4},
			BitPos: 2,
			Want: testResult{
				Code:   p,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{16, 0},
			BitPos: 0,
			Want: testResult{
				Code:   p,
				BitPos: 4,
				OK:     true,
			},
		},
		{
			Data:   []byte{4, 0},
			BitPos: 2,
			Want: testResult{
				Code:   p,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{32},
			BitPos: 0,
			Want: testResult{
				Code:   h,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{8},
			BitPos: 2,
			Want: testResult{
				Code:   h,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{32, 0},
			BitPos: 0,
			Want: testResult{
				Code:   h,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{8, 0},
			BitPos: 2,
			Want: testResult{
				Code:   h,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{128},
			BitPos: 0,
			Want: testResult{
				Code:   v0,
				BitPos: 1,
				OK:     true,
			},
		},
		{
			Data:   []byte{32},
			BitPos: 2,
			Want: testResult{
				Code:   v0,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{128, 0},
			BitPos: 0,
			Want: testResult{
				Code:   v0,
				BitPos: 1,
				OK:     true,
			},
		},
		{
			Data:   []byte{32, 0},
			BitPos: 2,
			Want: testResult{
				Code:   v0,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{96},
			BitPos: 0,
			Want: testResult{
				Code:   v1r,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{24},
			BitPos: 2,
			Want: testResult{
				Code:   v1r,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{96, 0},
			BitPos: 0,
			Want: testResult{
				Code:   v1r,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{24, 0},
			BitPos: 2,
			Want: testResult{
				Code:   v1r,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{12},
			BitPos: 0,
			Want: testResult{
				Code:   v2r,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{3},
			BitPos: 2,
			Want: testResult{
				Code:   v2r,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{12, 0},
			BitPos: 0,
			Want: testResult{
				Code:   v2r,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{3, 0},
			BitPos: 2,
			Want: testResult{
				Code:   v2r,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{6},
			BitPos: 0,
			Want: testResult{
				Code:   v3r,
				BitPos: 7,
				OK:     true,
			},
		},
		{
			Data:   []byte{1, 128},
			BitPos: 2,
			Want: testResult{
				Code:   v3r,
				BitPos: 9,
				OK:     true,
			},
		},
		{
			Data:   []byte{6, 0},
			BitPos: 0,
			Want: testResult{
				Code:   v3r,
				BitPos: 7,
				OK:     true,
			},
		},
		{
			Data:   []byte{1, 128, 0},
			BitPos: 2,
			Want: testResult{
				Code:   v3r,
				BitPos: 9,
				OK:     true,
			},
		},
		{
			Data:   []byte{64},
			BitPos: 0,
			Want: testResult{
				Code:   v1l,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{16},
			BitPos: 2,
			Want: testResult{
				Code:   v1l,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{64, 0},
			BitPos: 0,
			Want: testResult{
				Code:   v1l,
				BitPos: 3,
				OK:     true,
			},
		},
		{
			Data:   []byte{16, 0},
			BitPos: 2,
			Want: testResult{
				Code:   v1l,
				BitPos: 5,
				OK:     true,
			},
		},
		{
			Data:   []byte{8},
			BitPos: 0,
			Want: testResult{
				Code:   v2l,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{2},
			BitPos: 2,
			Want: testResult{
				Code:   v2l,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{8, 0},
			BitPos: 0,
			Want: testResult{
				Code:   v2l,
				BitPos: 6,
				OK:     true,
			},
		},
		{
			Data:   []byte{2, 0},
			BitPos: 2,
			Want: testResult{
				Code:   v2l,
				BitPos: 8,
				OK:     true,
			},
		},
		{
			Data:   []byte{4},
			BitPos: 0,
			Want: testResult{
				Code:   v3l,
				BitPos: 7,
				OK:     true,
			},
		},
		{
			Data:   []byte{1, 0},
			BitPos: 2,
			Want: testResult{
				Code:   v3l,
				BitPos: 9,
				OK:     true,
			},
		},
		{
			Data:   []byte{0, 0, 0},
			BitPos: 0,
			Want: testResult{
				Code:   code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   []byte{0, 0, 0},
			BitPos: 2,
			Want: testResult{
				Code:   code{},
				BitPos: 2,
				OK:     false,
			},
		},
		{
			Data:   []byte{0},
			BitPos: 0,
			Want: testResult{
				Code:   code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   []byte{0},
			BitPos: 2,
			Want: testResult{
				Code:   code{},
				BitPos: 2,
				OK:     false,
			},
		},
		{
			Data:   []byte{},
			BitPos: 0,
			Want: testResult{
				Code:   code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   []byte{},
			BitPos: 2,
			Want: testResult{
				Code:   code{},
				BitPos: 2,
				OK:     false,
			},
		},
		{
			Data:   nil,
			BitPos: 0,
			Want: testResult{
				Code:   code{},
				BitPos: 0,
				OK:     false,
			},
		},
		{
			Data:   nil,
			BitPos: 2,
			Want: testResult{
				Code:   code{},
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
		Code code
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
				Code: p,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(8192, 3),
			BitPos: 0,
			Want: testResult{
				Code: h,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(32768, 1),
			BitPos: 0,
			Want: testResult{
				Code: v0,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(24576, 3),
			BitPos: 0,
			Want: testResult{
				Code: v1r,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(3072, 6),
			BitPos: 0,
			Want: testResult{
				Code: v2r,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(1536, 7),
			BitPos: 0,
			Want: testResult{
				Code: v3r,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(16384, 3),
			BitPos: 0,
			Want: testResult{
				Code: v1l,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(2048, 6),
			BitPos: 0,
			Want: testResult{
				Code: v2l,
				OK:   true,
			},
		},
		{
			Codes:  formTestCodes(1024, 7),
			BitPos: 0,
			Want: testResult{
				Code: v3l,
				OK:   true,
			},
		},
		{
			Codes:  []uint16{0, 0, 0},
			BitPos: 0,
			Want: testResult{
				Code: code{},
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
