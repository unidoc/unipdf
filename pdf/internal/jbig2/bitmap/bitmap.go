package bitmap

import (
	"errors"
	"github.com/unidoc/unidoc/common"
	"image"
	"image/color"
)

// ErrIndexOutOfRange is the error that returns if the bit index is out of range
var ErrIndexOutOfRange error = errors.New("Index out of range")

// Bitmap is the jbig2 bitmap representation
type Bitmap struct {
	Width, Height int

	// BitmapNumber is the number
	BitmapNumber int

	// RowStride is the number of bytes set per row
	RowStride int

	// Data saves the bits data for the bitmap
	Data []byte

	isVanilla bool
}

// New creates new bitmap with the parameters as provided in the arguments
func New(width, height int) *Bitmap {
	bm := &Bitmap{
		Width:     width,
		Height:    height,
		RowStride: (width + 7) >> 3,
	}

	common.Log.Debug("Created bitmap - Width: %d, Height: %d", width, height)

	bm.Data = make([]byte, height*bm.RowStride)

	return bm
}

// String implements the Stringer interface for the bitmap.
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

// GetPixel gets the pixel value at the coordinates 'x' and 'y'
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

// SetPixel sets the pixel at 'x', 'y' with the value of 'pixel'
// Returns an error if the index is out of range
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

// SetDefaultPixel sets all pixel to '1'
func (b *Bitmap) SetDefaultPixel() {
	for i := range b.Data {
		b.Data[i] = byte(0xff)
	}
}

// GetByteIndex gets the byte index from the provided data at index 'x','y'
func (b *Bitmap) GetByteIndex(x, y int) int {
	return y*b.RowStride + (x >> 3)
}

// GetByte gets the byte at 'index'
func (b *Bitmap) GetByte(index int) (byte, error) {
	if index > len(b.Data)-1 {
		return 0, ErrIndexOutOfRange
	}
	return b.Data[index], nil
}

// SetByte sets the byte at 'index' of value: 'v'
// Returns an error if the index is out of range
func (b *Bitmap) SetByte(index int, v byte) error {
	if index > len(b.Data)-1 {
		return ErrIndexOutOfRange
	}

	// common.Log.Debug("SetByte: %08b at index: %d", v, index)
	b.Data[index] = v
	return nil
}

// GetBitOffset gets the bit offset at the 'x' coordinate
func (b *Bitmap) GetBitOffset(x int) int {
	return x & 0x07
}

// Equals checks if all the pixels in the 'b' bitmap are equals to the 's' bitmap
// Used in testing purpose only
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

// GetUnpaddedData gets the data without padding for the rowstride
func (b *Bitmap) GetUnpaddedData() []byte {

	var (
		padding  = uint(b.Width & 0x07)
		useShift = padding != 0
	)

	common.Log.Debug("Padding: %d, rowStride: %d, Width: %d", padding, b.RowStride, b.Width)

	if !useShift {
		return b.Data
	}

	size := b.Width * b.Height
	common.Log.Debug("Size before: %d", size)
	if size%8 != 0 {
		size >>= 3
		size++
	} else {
		size >>= 3
	}
	padding = 8 - padding

	common.Log.Debug("Size: %d, Width: %d Height: %d", size, b.Width, b.Height)

	var data = make([]byte, size)

	var (
		// currentIndex is the index at which the byte is currently in the 'data' byte array
		currentIndex int

		// siginificantBits are the bits in the current byte that are already set and significant
		significantBits uint
	)

	for line := 0; line < b.Height; line++ {

		// iterate over all rowstrides within the line
		for i := 0; i < b.RowStride; i++ {
			// get the byte at x, y
			bt := b.Data[line*b.RowStride+i]

			if line == 0 {
				// copy the line
				data[currentIndex] = bt

				if i == b.RowStride-1 {
					significantBits = padding
				} else {
					currentIndex++
				}
			} else {

				// last byte i.e. 11010000 with padding 3
				// and 'bt' is 10101111
				// then last byte should be 11010 | 101 and
				// current byte at index should be 01111
				// the padding should be 7 - padding
				lastByte := data[currentIndex]

				common.Log.Debug("Line: %d, byteNo: %d, byteValue: %08b currentIndex: %d", line, i, bt, currentIndex)

				if i == b.RowStride-1 {
					common.Log.Debug("LastRow")
					// if the line is the last in the row add the default padding to the current padding
					// i.e.
					// source byte: 0100100 with padding = 2 the data to take should be 01001000
					// the only significant bits are 010010xx
					// then the lastbyte should join the data with current padding i.e.

					// significantBits is 3 and byte is 01110000 shoule be now 01110000 | (010010xx >> (8-(8-3)))
					// the data which still needs to be added is 0100000 with significantBits
					dif := significantBits + (8 - padding)
					common.Log.Debug("Significant: %d, Padding: %d, Dif: %d", significantBits, padding, dif)
					if dif > 8 {
						common.Log.Debug("more significant bits.")
						// if the current significant bits number and the padded bits number greater than 8
						// write as many bits as possible to the current byte and add the currentIndex
						// 5 + (8 - 2) = 11
						// get 8-5 bits at first and store it on the index
						common.Log.Debug("LastByte before: %08b", lastByte)
						common.Log.Debug("Byte: %08b", bt)
						lastByte = lastByte | (bt >> (8 - significantBits))
						data[currentIndex] = lastByte
						common.Log.Debug("LastByte after first: %08b", lastByte)

						// increase the index
						currentIndex++
						common.Log.Debug("Current Index: %d", currentIndex)

						// get the rest bits 11 - (8 - 5) =  (dif - (8 - significantBits))
						// which would be the significant bits
						significantBits = dif - (8 - significantBits)
						lastByte = bt << significantBits

						data[currentIndex] = lastByte
					} else if dif == 8 {
						lastByte = lastByte | (bt >> (8 - significantBits))
						data[currentIndex] = lastByte
					} else {
						common.Log.Debug("Dif: %d", dif)
						// if the difference is smaller or equal to 8
						lastByte = lastByte | bt>>(8-dif)
						significantBits = dif

						data[currentIndex] = lastByte

						if dif == 8 {
							currentIndex++
						}
					}
				} else {
					// byte bt: 00011000 significantBits: 3 lastByte is 10100000
					// the lastByte should be 10100011
					// so it should be 10100000 | 00011000 >> 3
					common.Log.Debug("Normal - Current LastByte: %08b, sigBits: %d", lastByte, significantBits)
					lastByte |= (bt >> (8 - significantBits))

					data[currentIndex] = lastByte
					common.Log.Debug("The LastByte value: %08b", lastByte)

					currentIndex++

					lastByte = bt << significantBits
					data[currentIndex] = lastByte

					common.Log.Debug("Normal - 2 - LastByte: %08b Significant bits: %d", lastByte, significantBits)

				}

			}

		}
	}

	return data
}

// ToImage gets the bitmap data and store in the image.Image
func (b *Bitmap) ToImage() image.Image {

	img := image.NewGray(image.Rect(0, 0, b.Width-1, b.Height-1))
	for x := 0; x < b.Width; x++ {
		for y := 0; y < b.Height; y++ {
			var c color.Color
			if b.GetPixel(x, y) {
				c = color.Black
			} else {
				c = color.White
			}
			img.Set(x, y, c)
		}
	}
	return img
}

// GetVanillaData vanilla is the bit interpretation where the 0'th bit means black and 1'th bit means white
func (b *Bitmap) GetVanillaData() []byte {
	if b.isVanilla {
		return b.Data
	}
	b.inverseData()
	return b.Data
}

// GetChocolateData 'chocolate' data is the bit interpretation where the 0'th bit means white and the 1'th bit means black
func (b *Bitmap) GetChocolateData() []byte {
	if !b.isVanilla {
		return b.Data
	}
	b.inverseData()
	return b.Data
}

func (b *Bitmap) inverseData() {
	for i := 0; i < len(b.Data); i++ {
		b.Data[i] = ^b.Data[i]
	}
	b.isVanilla = !b.isVanilla
}
