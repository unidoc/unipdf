package bitmap

import (
	"errors"
	"github.com/unidoc/unidoc/common"
)

// ErrIndexOutOfRange is the error that returns if the bit index is out of range
var ErrIndexOutOfRange error = errors.New("Index out of range")

// "github.com/unidoc/unidoc/common"
// "github.com/unidoc/unidoc/pdf/internal/jbig2/bitset"

// Bitmap is the jbig2 bitmap representation
type Bitmap struct {
	Width, Height int

	// BitmapNumber is the number
	BitmapNumber int

	RowStride int

	// Data saves the bits data for the bitmap
	// Data *bitset.BitSet
	Data []byte
}

// New creates new bitmap with the parameters as provided in the arguments
func New(width, height int) *Bitmap {
	bm := &Bitmap{
		Width:     width,
		Height:    height,
		RowStride: (width + 7) >> 3,
		// Decoder: decoder,
		// Data: bitset.NewBitSet(width * height),
	}

	common.Log.Debug("Created bitmap - Width: %d, Height: %d", width, height)

	bm.Data = make([]byte, height*bm.RowStride)

	return bm
}

// Stringify the bitmap
func (b *Bitmap) String() string {

	var s string = "\n"
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

func (b *Bitmap) GetPixel(x, y int) bool {
	i := b.GetByteIndex(x, y)
	o := b.GetBitOffset(x)
	shift := uint(7 - o)
	if (b.Data[i]>>shift)&0x01 >= 1 {
		return true
	}
	return false
}

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

func (b *Bitmap) SetDefaultPixel() {
	for i := range b.Data {
		b.Data[i] = byte(0xff)
	}
}

func (b *Bitmap) GetByteIndex(x, y int) int {
	return y*b.RowStride + (x >> 3)
}

func (b *Bitmap) GetByte(index int) (byte, error) {
	if index > len(b.Data)-1 {
		return 0, ErrIndexOutOfRange
	}
	return b.Data[index], nil
}

func (b *Bitmap) SetByte(index int, v byte) error {
	if index > len(b.Data)-1 {
		return ErrIndexOutOfRange
	}

	// common.Log.Debug("SetByte: %08b at index: %d", v, index)
	b.Data[index] = v
	return nil
}

func (b *Bitmap) GetBitOffset(x int) int {
	return x & 0x07
}

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
