/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ccittfax

import (
	_ "image/png"
	"io/ioutil"
	"os"
	"testing"
)

func TestEncode(t *testing.T) {
	type testData struct {
		Encoder      *Encoder
		Pixels       [][]byte
		WantFilePath string
	}

	pixels, err := getPixels(testDataPath + "p3_0.png")
	if err != nil {
		t.Fatalf("Error getting pixels from image: %v\n", err)
	}

	tests := []testData{
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-Aligned-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:          0,
				Columns:    2560,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-EOL-Aligned-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:                0,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-EOL-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:          0,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-EOL-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:         0,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-EOL-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:       0,
				Columns: 2560,
				Rows:    3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K0-Columns2560-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-Aligned-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:          4,
				Columns:    2560,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-EOL-Aligned-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:                4,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-EOL-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:          4,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-EOL-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				K:         4,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-EOL-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:       4,
				Columns: 2560,
				Rows:    3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K4-Columns2560-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				K:                -1,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-Aligned-EOFB.gr4",
		},
		{
			Encoder: &Encoder{
				Columns:          2560,
				K:                -1,
				EndOfLine:        false,
				EncodedByteAlign: true,
				EndOfBlock:       false,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-Aligned-Rows3295.gr4",
		},
		{
			Encoder: &Encoder{
				K:          -1,
				Columns:    2560,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-EOFB.gr4",
		},
		{
			Encoder: &Encoder{
				K:                -1,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-EOL-Aligned-EOFB.gr4",
		},
		{
			Encoder: &Encoder{
				K:                -1,
				Columns:          2560,
				EndOfLine:        true,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-EOL-Aligned-Rows3295.gr4",
		},
		{
			Encoder: &Encoder{
				K:          -1,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-EOL-EOFB.gr4",
		},
		{
			Encoder: &Encoder{
				K:         -1,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-EOL-Rows3295.gr4",
		},
		{
			Encoder: &Encoder{
				K:       -1,
				Columns: 2560,
				Rows:    3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "K-1-Columns2560-Rows3295.gr4",
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-Aligned-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                0,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          0,
				Columns:    2560,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOFB.gr3",
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
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-Aligned-EOFB.gr3",
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
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          0,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:  true,
				K:         0,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-EOL-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1: true,
				K:        0,
				Columns:  2560,
				Rows:     3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K0-Columns2560-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-Aligned-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                4,
				Columns:          2560,
				EncodedByteAlign: true,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          4,
				Columns:    2560,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOFB.gr3",
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
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-Aligned-EOFB.gr3",
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
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-Aligned-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          4,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-EOFB.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:  true,
				K:         4,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-EOL-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1: true,
				K:        4,
				Columns:  2560,
				Rows:     3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K4-Columns2560-Rows3295.gr3",
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				K:                -1,
				Columns:          2560,
				EncodedByteAlign: true,
				EndOfBlock:       true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-Aligned-EOFB.gr4",
		},
		{
			Encoder: &Encoder{
				BlackIs1:         true,
				Columns:          2560,
				K:                -1,
				EndOfLine:        false,
				EncodedByteAlign: true,
				EndOfBlock:       false,
				Rows:             3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-Aligned-Rows3295.gr4",
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          -1,
				Columns:    2560,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOFB.gr4",
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
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-Aligned-EOFB.gr4",
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
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-Aligned-Rows3295.gr4",
		},
		{
			Encoder: &Encoder{
				BlackIs1:   true,
				K:          -1,
				Columns:    2560,
				EndOfLine:  true,
				EndOfBlock: true,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-EOFB.gr4",
		},
		{
			Encoder: &Encoder{
				BlackIs1:  true,
				K:         -1,
				Columns:   2560,
				EndOfLine: true,
				Rows:      3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-EOL-Rows3295.gr4",
		},
		{
			Encoder: &Encoder{
				BlackIs1: true,
				K:        -1,
				Columns:  2560,
				Rows:     3295,
			},
			Pixels:       pixels,
			WantFilePath: testDataPath + "BlackIs1-K-1-Columns2560-Rows3295.gr4",
		},
	}

	for _, test := range tests {
		f, err := os.Open(test.WantFilePath)
		if err != nil {
			t.Errorf("Error opening file with pixels wanted: %v\n", err)
		} else {
			wantEncoded, err := ioutil.ReadAll(f)
			if err != nil {
				f.Close()

				t.Errorf("Error reading pixels wanted from file: %v\n", err)
			} else {
				gotEncoded := test.Encoder.Encode(test.Pixels)

				if len(gotEncoded) != len(wantEncoded) {
					t.Errorf("Wrong encoded len. Got %v, want %v\n",
						len(gotEncoded), len(wantEncoded))
				} else {
					for i := range gotEncoded {
						if gotEncoded[i] != wantEncoded[i] {
							t.Errorf("Wrong pixel at %v for file %v. Got %v, want %v\n",
								i, test.WantFilePath, gotEncoded[i], wantEncoded[i])

							break
						}
					}
				}

				f.Close()
			}
		}
	}
}
