/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

var (
	frameBitmap *Bitmap
	imageBitmap *Bitmap
)

// encodeDataBitmap is the data used to create bitmap for  the encoding process.
// The bitmap has 50 pix width and and 21 height.
//
//            1        2        3        4        5        6         7
//
//                111111 11112222 22222233 33333333 44444444 44
//     01234567 89012345 67890123 45678901 23456789 01234567 89
//
//  0  00000000 00000000 00000000 00000000 00000000 00000000 00000000
//  1  00000000 00000000 00000000 00000000 00000000 00000000 00000000
//  2  00111110 01111000 00100111 11000010 00100111 10010001 00000000
//  3  00100010 01001000 00100001 00000011 00100100 10010001 00000000
//  4  00100010 01001000 00100001 00000010 10100100 10010101 00000000
//  5  00100010 01001000 00100001 00000010 01100100 10011011 00000000
//  6  00111100 01111000 00100001 00000010 00100111 10010001 00000000
//  7  00000000 00000000 00000000 00000000 00000000 00000000 00000000
//  8  00000000 00000000 00000000 00000000 00000000 00000000 00000000
//  9  00000000 00000000 00000000 00000000 00000000 00000000 00000000
// 10  00000000 00000000 00000000 00000000 00000000 00000000 00000000
// 11  00000000 00000000 00000000 00000000 00000000 00000000 00000000
// 12  00000000 00000000 00000000 00000000 00000000 00000000 00000000
// 13  00000000 00000000 01111111 11111000 00000000 00000000 00000000
// 14  00000000 00000000 01111111 11111000 00000000 00000000 00000000
// 15  00000000 00000000 01100011 00011000 00000000 00000000 00000000
// 16  00000000 00000000 01100011 00011000 00000000 00000000 00000000
// 17  00000000 00000000 01100011 00011000 00000000 00000000 00000000
// 18  00000000 00000000 01111111 11111000 00000000 00000000 00000000
// 19  00000000 00000000 00010101 01010000 00000000 00000000 00000000
// 20  00000000 00000000 00000000 00000000 00000000 00000000 00000000
// 21  00000000 00000000 00000000 00000000 00000000 00000000 00000000/
var encodeBitmapData = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x3E, 0x78, 0x27, 0xC2, 0x27, 0x91, 0x00,
	0x22, 0x48, 0x21, 0x03, 0x24, 0x91, 0x00,
	0x22, 0x48, 0x21, 0x02, 0xA4, 0x95, 0x00,
	0x22, 0x48, 0x21, 0x02, 0x64, 0x9B, 0x00,
	0x3C, 0x78, 0x21, 0x02, 0x27, 0x91, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x63, 0x18, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x7F, 0xF8, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x15, 0x50, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func init() {
	const processName = "bitmaps.initialization"
	// prepare frame bitmap
	frameBitmap = New(50, 40)
	var err error
	frameBitmap, err = frameBitmap.AddBorder(2, 1)
	if err != nil {
		panic(errors.Wrap(err, processName, "frameBitmap"))
	}

	// prepare image bitmap
	imageBitmap, err = NewWithData(50, 22, encodeBitmapData)
	if err != nil {
		panic(errors.Wrap(err, processName, "imageBitmap"))
	}
}

// TstFrameBitmap gets the test frame bitmap
func TstFrameBitmap() *Bitmap {
	return frameBitmap.Copy()
}

// TstFrameBitmapData gets the test frame bitmap data
func TstFrameBitmapData() []byte {
	return frameBitmap.Data
}

// TstImageBitmap gets the test image bitmap
func TstImageBitmap() *Bitmap {
	return imageBitmap.Copy()
}

// TstImageBitmapData gets the test image bitmap data
func TstImageBitmapData() []byte {
	return imageBitmap.Data
}

// TstWordBitmap creates a bitmap with the words like:
// DO IT NOW
// OR NEVER
// without any boundaries.
func TstWordBitmap(t *testing.T, scale ...int) *Bitmap {
	// write following symbols:
	// DO IT NOW
	// OR NEVER
	// 414_115_41415 = 9 + space + 7 + space + 15
	// 414_41514 = 9 + space + 15
	sc := 1
	if len(scale) > 0 {
		sc = scale[0]
	}
	space := 3
	width := 9 + 7 + 15 + 2*space
	height := 5 + space + 5
	bm := New(width*sc, height*sc)
	bms := &Bitmaps{}
	var x *int
	space *= sc
	tmp := 0
	x = &tmp
	y := 0

	// D
	sym := TstDSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// O
	sym = TstOSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, space)
	// spaces

	// I
	sym = TstISymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// T
	sym = TstTSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, space)

	// N
	sym = TstNSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// O
	sym = TstOSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// W
	sym = TstWSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 0)

	*x = 0

	// next line - 8 symbol max size + space
	y = 5*sc + space
	// OR
	sym = TstOSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstRSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, space)

	// NEVER
	sym = TstNSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstESymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstVSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstESymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstRSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 0)

	TstWriteSymbols(t, bms, bm)
	return bm
}

// TstWordBitmapWithSpaces gets no space from the top, bottom, left and right edge.
func TstWordBitmapWithSpaces(t *testing.T, scale ...int) *Bitmap {
	// write following symbols:
	// DO IT NOW
	// OR NEVER
	// 414_115_41415 = 9 + space + 7 + space + 15
	// 414_41514 = 9 + space + 15
	sc := 1
	if len(scale) > 0 {
		sc = scale[0]
	}
	space := 3
	width := 9 + 7 + 15 + 2*space + 2*space
	height := 5 + space + 5 + 2*space
	bm := New(width*sc, height*sc)
	bms := &Bitmaps{}
	var x *int
	space *= sc
	tmp := space
	x = &tmp
	y := space

	// D
	sym := TstDSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// O
	sym = TstOSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, space)
	// spaces

	// I
	sym = TstISymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// T
	sym = TstTSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, space)

	// N
	sym = TstNSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// O
	sym = TstOSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	// W
	sym = TstWSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 0)

	*x = space

	// next line - 8 symbol max size + space
	y = 5*sc + space
	// OR
	sym = TstOSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstRSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, space)

	// NEVER
	sym = TstNSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstESymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstVSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstESymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 1*sc)

	sym = TstRSymbol(t, scale...)
	TstAddSymbol(t, bms, sym, x, y, 0)

	TstWriteSymbols(t, bms, bm)
	return bm
}

// TstAddSymbol is a helper function that adds 'sym' at the 'x' and 'y' position.
func TstAddSymbol(t *testing.T, bms *Bitmaps, sym *Bitmap, x *int, y int, space int) {
	bms.AddBitmap(sym)
	box := image.Rect(*x, y, *x+sym.Width, y+sym.Height)
	bms.AddBox(&box)
	*x += sym.Width + space
}

// TstWriteSymbols is a helper function to write given symbols from bitmaps into 'src' bitmap
// at the given 'x.
func TstWriteSymbols(t *testing.T, bms *Bitmaps, src *Bitmap) {
	for i := 0; i < bms.Size(); i++ {
		bm := bms.Values[i]
		box := bms.Boxes[i]
		err := src.RasterOperation(box.Min.X, box.Min.Y, bm.Width, bm.Height, PixSrc, bm, 0, 0)
		require.NoError(t, err)
	}
}

// TstDSymbol is a helper function to get 'D' symbol.
func TstDSymbol(t *testing.T, scale ...int) *Bitmap {
	// 11110000
	// 10010000
	// 10010000
	// 10010000
	// 11100000
	bm, err := NewWithData(4, 5, []byte{0xf0, 0x90, 0x90, 0x90, 0xE0})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstVSymbol is a helper function to get 'V' symbol.
func TstVSymbol(t *testing.T, scale ...int) *Bitmap {
	// 10001000
	// 10001000
	// 10001000
	// 01010000
	// 00100000
	bm, err := NewWithData(5, 5, []byte{0x88, 0x88, 0x88, 0x50, 0x20})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstOSymbol is a helper function to get 'O' symbol.
func TstOSymbol(t *testing.T, scale ...int) *Bitmap {
	// 11110000
	// 10010000
	// 10010000
	// 10010000
	// 11110000
	bm, err := NewWithData(4, 5, []byte{0xF0, 0x90, 0x90, 0x90, 0xF0})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstISymbol is a helper function to get 'I' symbol.
func TstISymbol(t *testing.T, scale ...int) *Bitmap {
	// 10000000
	// 10000000
	// 10000000
	// 10000000
	// 10000000
	bm, err := NewWithData(1, 5, []byte{0x80, 0x80, 0x80, 0x80, 0x80})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstTSymbol is a helper function to write 'T' letter
func TstTSymbol(t *testing.T, scale ...int) *Bitmap {
	// 11111000
	// 00100000
	// 00100000
	// 00100000
	// 00100000
	bm, err := NewWithData(5, 5, []byte{0xF8, 0x20, 0x20, 0x20, 0x20})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstNSymbol is a helper function to write 'N' letter.
func TstNSymbol(t *testing.T, scale ...int) *Bitmap {
	// 1001000
	// 1101000
	// 1011000
	// 1001000
	// 1001000
	bm, err := NewWithData(4, 5, []byte{0x90, 0xD0, 0xB0, 0x90, 0x90})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstWSymbol is a helper function to write 'W' letter.
func TstWSymbol(t *testing.T, scale ...int) *Bitmap {
	// 10001000
	// 10001000
	// 10101000
	// 11011000
	// 10001000
	bm, err := NewWithData(5, 5, []byte{0x88, 0x88, 0xA8, 0xD8, 0x88})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstRSymbol is a helper function to write 'R' letter.
func TstRSymbol(t *testing.T, scale ...int) *Bitmap {
	//  11110000
	//  10010000
	//  11110000
	//  10100000
	//  10010000
	bm, err := NewWithData(4, 5, []byte{0xF0, 0x90, 0xF0, 0xA0, 0x90})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstESymbol is a helper function to write 'E' letter.
func TstESymbol(t *testing.T, scale ...int) *Bitmap {
	// 11110000
	// 10000000
	// 11100000
	// 10000000
	// 11110000
	bm, err := NewWithData(4, 5, []byte{0xF0, 0x80, 0xE0, 0x80, 0xF0})
	require.NoError(t, err)
	return TstGetScaledSymbol(t, bm, scale...)
}

// TstGetScaledSymbol is a helper function to get scaled bitmap.
func TstGetScaledSymbol(t *testing.T, sm *Bitmap, scale ...int) *Bitmap {
	if len(scale) == 0 {
		return sm
	}

	if scale[0] == 1 {
		return sm
	}

	bm, err := MorphSequence(sm, MorphProcess{Operation: MopReplicativeBinaryExpansion, Arguments: scale})
	require.NoError(t, err)
	return bm
}

// TstPSymbol is a helper function to get 'P' symbol.
func TstPSymbol(t *testing.T) *Bitmap {
	t.Helper()
	symbol := New(5, 8)
	require.NoError(t, symbol.SetPixel(0, 0, 1))
	require.NoError(t, symbol.SetPixel(1, 0, 1))
	require.NoError(t, symbol.SetPixel(2, 0, 1))
	require.NoError(t, symbol.SetPixel(3, 0, 1))
	require.NoError(t, symbol.SetPixel(4, 1, 1))
	require.NoError(t, symbol.SetPixel(0, 1, 1))
	require.NoError(t, symbol.SetPixel(4, 2, 1))
	require.NoError(t, symbol.SetPixel(0, 2, 1))
	require.NoError(t, symbol.SetPixel(4, 3, 1))
	require.NoError(t, symbol.SetPixel(0, 3, 1))
	require.NoError(t, symbol.SetPixel(0, 4, 1))
	require.NoError(t, symbol.SetPixel(1, 4, 1))
	require.NoError(t, symbol.SetPixel(2, 4, 1))
	require.NoError(t, symbol.SetPixel(3, 4, 1))
	require.NoError(t, symbol.SetPixel(0, 5, 1))
	require.NoError(t, symbol.SetPixel(0, 6, 1))
	require.NoError(t, symbol.SetPixel(0, 7, 1))
	return symbol
}

// TstASymbol is a helper function to get 'A' symbol.
func TstASymbol(t *testing.T) *Bitmap {
	t.Helper()
	a := New(6, 6)
	require.NoError(t, a.SetPixel(1, 0, 1))
	require.NoError(t, a.SetPixel(2, 0, 1))
	require.NoError(t, a.SetPixel(3, 0, 1))
	require.NoError(t, a.SetPixel(4, 0, 1))
	require.NoError(t, a.SetPixel(5, 1, 1))
	require.NoError(t, a.SetPixel(1, 2, 1))
	require.NoError(t, a.SetPixel(2, 2, 1))
	require.NoError(t, a.SetPixel(3, 2, 1))
	require.NoError(t, a.SetPixel(4, 2, 1))
	require.NoError(t, a.SetPixel(5, 2, 1))
	require.NoError(t, a.SetPixel(0, 3, 1))
	require.NoError(t, a.SetPixel(5, 3, 1))
	require.NoError(t, a.SetPixel(0, 4, 1))
	require.NoError(t, a.SetPixel(5, 4, 1))
	require.NoError(t, a.SetPixel(1, 5, 1))
	require.NoError(t, a.SetPixel(2, 5, 1))
	require.NoError(t, a.SetPixel(3, 5, 1))
	require.NoError(t, a.SetPixel(4, 5, 1))
	require.NoError(t, a.SetPixel(5, 5, 1))
	return a
}

// TstCSymbol is a helper function to get 'C' symbol.
func TstCSymbol(t *testing.T) *Bitmap {
	t.Helper()
	c := New(6, 6)
	require.NoError(t, c.SetPixel(1, 0, 1))
	require.NoError(t, c.SetPixel(2, 0, 1))
	require.NoError(t, c.SetPixel(3, 0, 1))
	require.NoError(t, c.SetPixel(4, 0, 1))
	require.NoError(t, c.SetPixel(0, 1, 1))
	require.NoError(t, c.SetPixel(5, 1, 1))
	require.NoError(t, c.SetPixel(0, 2, 1))
	require.NoError(t, c.SetPixel(0, 3, 1))
	require.NoError(t, c.SetPixel(0, 4, 1))
	require.NoError(t, c.SetPixel(5, 4, 1))
	require.NoError(t, c.SetPixel(1, 5, 1))
	require.NoError(t, c.SetPixel(2, 5, 1))
	require.NoError(t, c.SetPixel(3, 5, 1))
	require.NoError(t, c.SetPixel(4, 5, 1))
	return c
}
