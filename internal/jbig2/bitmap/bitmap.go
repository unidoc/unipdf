/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"errors"
	"image"
	"image/color"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// ErrIndexOutOfRange is the error that returns if the bitmap byte index is out of range.
var ErrIndexOutOfRange = errors.New("bitmap byte index out of range")

// Bitmap is the jbig2 bitmap representation.
type Bitmap struct {
	// Width and Height represents bitmap dimensions.
	Width, Height int

	// BitmapNumber is the bitmap's id number.
	BitmapNumber int

	// RowStride is the number of bytes set per row.
	RowStride int

	// Data saves the bits data for the bitmap.
	Data []byte

	isVanilla bool
}

// New creates new bitmap with the parameters as provided in the arguments.
func New(width, height int) *Bitmap {
	bm := &Bitmap{
		Width:     width,
		Height:    height,
		RowStride: (width + 7) >> 3,
	}
	bm.Data = make([]byte, height*bm.RowStride)

	return bm
}

// NewWithData creates new bitmap with the provided 'width', 'height' and the byte slice 'data'.
func NewWithData(width, height int, data []byte) *Bitmap {
	bm := New(width, height)
	bm.Data = data
	return bm
}

// Equals checks if all the pixels in the 'b' bitmap are equals to the 's' bitmap.
func (b *Bitmap) Equals(s *Bitmap) bool {
	if len(b.Data) != len(s.Data) {
		return false
	}

	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			if b.GetPixel(x, y) != s.GetPixel(x, y) {
				return false
			}
		}
	}

	return true
}

// GetBitOffset gets the bit offset at the 'x' coordinate.
func (b *Bitmap) GetBitOffset(x int) int {
	return x & 0x07
}

// GetByte gets and returns the byte at given 'index'.
func (b *Bitmap) GetByte(index int) (byte, error) {
	if index > len(b.Data)-1 || index < 0 {
		return 0, ErrIndexOutOfRange
	}
	return b.Data[index], nil
}

// GetByteIndex gets the byte index from the bitmap at coordinates 'x','y'.
func (b *Bitmap) GetByteIndex(x, y int) int {
	return y*b.RowStride + (x >> 3)
}

// GetChocolateData gets bitmap data as a byte slice with Chocolate bit interpretation.
// 'Chocolate' data is the bit interpretation where the 0'th bit means white and the 1'th bit means black.
// The naming convention based on the: `https://en.wikipedia.org/wiki/Binary_image#Interpretation` page.
func (b *Bitmap) GetChocolateData() []byte {
	if b.isVanilla {
		b.inverseData()
	}
	return b.Data
}

// GetPixel gets the pixel value at the coordinates 'x', 'y'.
func (b *Bitmap) GetPixel(x, y int) bool {
	i := b.GetByteIndex(x, y)
	o := b.GetBitOffset(x)
	shift := uint(7 - o)
	if i > len(b.Data)-1 {
		common.Log.Debug("Trying to get pixel out of the data range. x: '%d', y:'%d', bm: '%s'", x, y, b)
		return false
	}
	if (b.Data[i]>>shift)&0x01 >= 1 {
		return true
	}
	return false
}

// GetUnpaddedData gets the data without row stride padding.
// The unpadded data contains bitmap.Height * bitmap.Width bits with
// optional last byte padding.
func (b *Bitmap) GetUnpaddedData() ([]byte, error) {
	padding := uint(b.Width & 0x07)
	if padding == 0 {
		return b.Data, nil
	}

	size := b.Width * b.Height
	if size%8 != 0 {
		size >>= 3
		size++
	} else {
		size >>= 3
	}

	data := make([]byte, size)
	w := writer.NewMSB(data)

	for y := 0; y < b.Height; y++ {
		// btIndex is the byte index per row.
		for btIndex := 0; btIndex < b.RowStride; btIndex++ {
			bt := b.Data[y*b.RowStride+btIndex]
			if btIndex != b.RowStride-1 {
				err := w.WriteByte(bt)
				if err != nil {
					return nil, err
				}
				continue
			}

			for i := uint(0); i < padding; i++ {
				err := w.WriteBit(int(bt >> (7 - i) & 0x01))
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return data, nil
}

// GetVanillaData gets bitmap data as a byte slice with Vanilla bit interpretation.
// 'Vanilla' is the bit interpretation where the 0'th bit means black and 1'th bit means white.
// The naming convention based on the `https://en.wikipedia.org/wiki/Binary_image#Interpretation` page.
func (b *Bitmap) GetVanillaData() []byte {
	if !b.isVanilla {
		b.inverseData()
	}
	return b.Data
}

// SetPixel sets the pixel at 'x', 'y' coordinates with the value of 'pixel'.
// Returns an error if the index is out of range.
func (b *Bitmap) SetPixel(x, y int, pixel byte) error {
	i := b.GetByteIndex(x, y)
	if i > len(b.Data)-1 {
		return ErrIndexOutOfRange
	}
	o := b.GetBitOffset(x)

	shift := uint(7 - o)
	src := b.Data[i]

	result := src | (pixel & 0x01 << shift)
	b.Data[i] = result

	return nil
}

// SetDefaultPixel sets all bits within bitmap to '1'.
func (b *Bitmap) SetDefaultPixel() {
	for i := range b.Data {
		b.Data[i] = byte(0xff)
	}
}

// SetByte sets the byte at 'index' with value 'v'.
// Returns an error if the index is out of range.
func (b *Bitmap) SetByte(index int, v byte) error {
	if index > len(b.Data)-1 || index < 0 {
		return ErrIndexOutOfRange
	}

	b.Data[index] = v
	return nil
}

// String implements the Stringer interface.
func (b *Bitmap) String() string {
	var s = "\n"
	for y := 0; y < b.Height; y++ {
		var row string
		for x := 0; x < b.Width; x++ {
			pix := b.GetPixel(x, y)
			if pix {
				row += "1"
			} else {
				row += "0"
			}
		}
		s += row + "\n"
	}
	return s
}

// ToImage gets the bitmap data and store in the image.Image.
func (b *Bitmap) ToImage() image.Image {
	img := image.NewGray(image.Rect(0, 0, b.Width-1, b.Height-1))
	for x := 0; x < b.Width; x++ {
		for y := 0; y < b.Height; y++ {
			c := color.Black
			if b.GetPixel(x, y) {
				c = color.White
			}
			img.Set(x, y, c)
		}
	}
	return img
}

// InverseData inverses the data if the 'isChocolate' flag matches
// current bitmap 'isVanilla' state.
func (b *Bitmap) InverseData(isChocolate bool) {
	if b.isVanilla != !isChocolate {
		b.inverseData()
	}
}

func (b *Bitmap) inverseData() {
	for i := 0; i < len(b.Data); i++ {
		b.Data[i] = ^b.Data[i]
	}
	b.isVanilla = !b.isVanilla
}
