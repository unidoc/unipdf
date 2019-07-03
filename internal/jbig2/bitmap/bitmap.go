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

// GetChocolateData gets bitmap data as a byte sice with Chocolate bit intepretation.
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
func (b *Bitmap) GetUnpaddedData() []byte {
	padding := uint(b.Width & 0x07)

	if padding == 0 {
		return b.Data
	}

	size := b.Width * b.Height
	if size%8 != 0 {
		size >>= 3
		size++
	} else {
		size >>= 3
	}

	padding = 8 - padding
	data := make([]byte, size)
	var (
		// currentIndex is the index at which the byte is currently in the 'data' byte array.
		currentIndex int
		// siginificantBits are the bits in the current byte that are already set and significant.
		significantBits uint
	)
	for line := 0; line < b.Height; line++ {
		// iterate over all rowstrides within the line.
		for i := 0; i < b.RowStride; i++ {
			// get the byte at x, y.
			bt := b.Data[line*b.RowStride+i]

			if line == 0 {
				// copy the line.
				data[currentIndex] = bt
				if i == b.RowStride-1 {
					significantBits = padding
				} else {
					currentIndex++
				}
				continue
			}
			// last byte i.e. 11010000 with padding 3
			// and 'bt' is 10101111
			// then last byte should be 11010 | 101 and
			// current byte at index should be 01111
			// the padding should be 7 - padding
			lastByte := data[currentIndex]

			if i != b.RowStride-1 {
				// byte bt: 00011000 significantBits: 3 lastByte is 10100000
				// the lastByte should be 10100011
				// so it should be 10100000 | 00011000 >> 3
				lastByte |= (bt >> (8 - significantBits))

				data[currentIndex] = lastByte

				currentIndex++

				lastByte = bt << significantBits
				data[currentIndex] = lastByte
				continue
			}

			// if the line is the last in the row add the default padding to the current padding
			// i.e.
			// source byte: 0100100 with padding = 2 the data to take should be 01001000
			// the only significant bits are 010010xx
			// then the lastbyte should join the data with current padding i.e.

			// significantBits is 3 and byte is 01110000 shoule be now 01110000 | (010010xx >> (8-(8-3)))
			// the data which still needs to be added is 0100000 with significantBits
			dif := significantBits + (8 - padding)

			if dif > 8 {
				// if the current significant bits number and the padded bits number greater than 8
				// write as many bits as possible to the current byte and add the currentIndex
				// 5 + (8 - 2) = 11
				// get 8-5 bits at first and store it on the index

				lastByte = lastByte | (bt >> (8 - significantBits))
				data[currentIndex] = lastByte

				// increase the index
				currentIndex++

				// get the rest bits 11 - (8 - 5) =  (dif - (8 - significantBits))
				// which would be the significant bits
				significantBits = dif - (8 - significantBits)
				lastByte = bt << significantBits

				data[currentIndex] = lastByte
			} else if dif == 8 {
				lastByte = lastByte | (bt >> (8 - significantBits))
				data[currentIndex] = lastByte
			} else {
				// if the difference is smaller or equal to 8
				lastByte = lastByte | bt>>(8-dif)
				significantBits = dif

				data[currentIndex] = lastByte

				if dif == 8 {
					currentIndex++
				}
			}
		}
	}
	return data
}

// GetVanillaData gets bitmap data as a byte sice with Vanilla bit intepretation.
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
			c := color.White
			if b.GetPixel(x, y) {
				c = color.Black
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func (b *Bitmap) inverseData() {
	for i := 0; i < len(b.Data); i++ {
		b.Data[i] = ^b.Data[i]
	}
	b.isVanilla = !b.isVanilla
}
