/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/core"
)

type dictKeyVal struct {
	Key   core.PdfObjectName
	Value core.PdfObject
}

func makeKeyVal(key core.PdfObjectName, val core.PdfObject) dictKeyVal {
	return dictKeyVal{Key: key, Value: val}
}

func makeDict(args ...dictKeyVal) *core.PdfObjectDictionary {
	dict := core.MakeDict()
	for _, kv := range args {
		dict.Set(kv.Key, kv.Value)
	}
	return dict
}

func TestInlineImageParsing(t *testing.T) {
	testcases := []struct {
		Content  string
		Expected ContentStreamOperations
	}{
		// Case 1. Typical inline image.
		{
			`
q Q
q 3608 0 0 4 604 5048 cm
BI
/CS/RGB
/W 902
/H 1
/BPC 8
/F/Fl
/DP<</Predictor 15
/Columns 902
/Colors 3>>
ID x<9a>í<94>$áµ½<81> ¼
ôh<8e>
EI Q
q 3608 0 0 4 604 5044 cm
`,
			ContentStreamOperations{
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "Q"},
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "cm", Params: makeParamsFromInts([]int64{3608, 0, 0, 4, 604, 5048})},
				&ContentStreamOperation{Operand: "BI",
					Params: []core.PdfObject{&ContentStreamInlineImage{
						ColorSpace:       core.MakeName("RGB"),
						Width:            core.MakeInteger(902),
						Height:           core.MakeInteger(1),
						BitsPerComponent: core.MakeInteger(8),
						Filter:           core.MakeName("Fl"),
						DecodeParms: makeDict(
							makeKeyVal("Predictor", core.MakeInteger(15)),
							makeKeyVal("Columns", core.MakeInteger(902)),
							makeKeyVal("Colors", core.MakeInteger(3)),
						),
						stream: []byte("x<9a>\u00ed<94>$\u00e1\u00b5\u00bd<81> \u00bc\n\u00f4h<8e>"),
					}},
				},
				&ContentStreamOperation{Operand: "Q"},
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "cm", Params: makeParamsFromInts([]int64{3608, 0, 0, 4, 604, 5044})},
			},
		},
		// Case 2. Inline image with EI inside ID ... EI
		{
			`
q 3608 0 0 4 604 5048 cm
BI
/CS/RGB
/W 902
/H 1
/BPC 8
/F/Fl
/DP<</Predictor 15
/Columns 902
/Colors 3>>
ID x<9a>í<94>[@^L EI   $áµ½<81> ¼
ôh<8e>
EI Q
q 3608 0 0 4 604 5044 cm
`,
			ContentStreamOperations{
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "cm", Params: makeParamsFromInts([]int64{3608, 0, 0, 4, 604, 5048})},
				&ContentStreamOperation{Operand: "BI",
					Params: []core.PdfObject{&ContentStreamInlineImage{
						ColorSpace:       core.MakeName("RGB"),
						Width:            core.MakeInteger(902),
						Height:           core.MakeInteger(1),
						BitsPerComponent: core.MakeInteger(8),
						Filter:           core.MakeName("Fl"),
						DecodeParms: makeDict(
							makeKeyVal("Predictor", core.MakeInteger(15)),
							makeKeyVal("Columns", core.MakeInteger(902)),
							makeKeyVal("Colors", core.MakeInteger(3)),
						),
						stream: []byte("x<9a>\u00ed<94>[@^L EI   $\u00e1\u00b5\u00bd<81> \u00bc\n\u00f4h<8e>"),
					}},
				},
				&ContentStreamOperation{Operand: "Q"},
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "cm", Params: makeParamsFromInts([]int64{3608, 0, 0, 4, 604, 5044})},
			},
		},
		// Case 3. Inline image data with EI inside data and content ending on the EI operand.
		{
			`
q Q
BI
/CS/RGB
/W 902
/H 1
/BPC 8
/F/Fl
ID x<9a>í<94>[@^L EI   $áµ½<81> ¼
ôh<8e>
EI`,
			ContentStreamOperations{
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "Q"},
				&ContentStreamOperation{Operand: "BI",
					Params: []core.PdfObject{&ContentStreamInlineImage{
						ColorSpace:       core.MakeName("RGB"),
						Width:            core.MakeInteger(902),
						Height:           core.MakeInteger(1),
						BitsPerComponent: core.MakeInteger(8),
						Filter:           core.MakeName("Fl"),
						stream:           []byte("x<9a>\u00ed<94>[@^L EI   $\u00e1\u00b5\u00bd<81> \u00bc\n\u00f4h<8e>"),
					}},
				},
			},
		},
		// Case 4. Inline image with one operand after EI.
		{
			`
q Q
BI
/CS/RGB
/W 902
/H 1
/BPC 8
/F/Fl
ID x<9a>
EI q`,
			ContentStreamOperations{
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "Q"},
				&ContentStreamOperation{Operand: "BI",
					Params: []core.PdfObject{&ContentStreamInlineImage{
						ColorSpace:       core.MakeName("RGB"),
						Width:            core.MakeInteger(902),
						Height:           core.MakeInteger(1),
						BitsPerComponent: core.MakeInteger(8),
						Filter:           core.MakeName("Fl"),
						stream:           []byte("x<9a>"),
					}},
				},
				&ContentStreamOperation{Operand: "q"},
			},
		},
		// Case 5. Inline image with " EI  q 1 " inside ID .. EI image data.
		// Note: would fail if more objects were in place or invalid data was in the contents.
		{
			`
q Q
BI
/CS/RGB
/W 902
/H 1
/BPC 8
/F/Fl
ID x<9a>í<94>[@^L EI  q 1 $áµ½<81> ¼
ôh<8e>
EI`,
			ContentStreamOperations{
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "Q"},
				&ContentStreamOperation{Operand: "BI",
					Params: []core.PdfObject{&ContentStreamInlineImage{
						ColorSpace:       core.MakeName("RGB"),
						Width:            core.MakeInteger(902),
						Height:           core.MakeInteger(1),
						BitsPerComponent: core.MakeInteger(8),
						Filter:           core.MakeName("Fl"),
						stream:           []byte("x<9a>\u00ed<94>[@^L EI  q 1 $\u00e1\u00b5\u00bd<81> \u00bc\n\u00f4h<8e>"),
					}},
				},
			},
		},
		// Case 6. Very long object after EI, check that handles as image data finished.
		{
			`
q Q
BI
/CS/RGB
/W 902
/H 1
/BPC 8
/F/Fl
ID x<9a>
EI (Very long string that is going to be only partially parser 123456789012345678901234567890) Tj`,
			ContentStreamOperations{
				&ContentStreamOperation{Operand: "q"},
				&ContentStreamOperation{Operand: "Q"},
				&ContentStreamOperation{Operand: "BI",
					Params: []core.PdfObject{&ContentStreamInlineImage{
						ColorSpace:       core.MakeName("RGB"),
						Width:            core.MakeInteger(902),
						Height:           core.MakeInteger(1),
						BitsPerComponent: core.MakeInteger(8),
						Filter:           core.MakeName("Fl"),
						stream:           []byte("x<9a>"),
					}},
				},
				&ContentStreamOperation{Operand: "Tj",
					Params: []core.PdfObject{
						core.MakeString("Very long string that is going to be only partially parser 123456789012345678901234567890"),
					},
				},
			},
		},
	}

	for i, tcase := range testcases {
		t.Logf("Case %d", i+1)
		parser := NewContentStreamParser(tcase.Content)
		ops, err := parser.Parse()
		require.NoError(t, err)
		require.NotNil(t, ops)
		require.Equal(t, tcase.Expected, *ops)
	}
}
