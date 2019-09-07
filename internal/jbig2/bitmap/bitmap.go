/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// tab8 contains number of '1' bits in each possible 8 bit value stored at it's index.
var tab8 [256]uint8

func init() {
	for i := 0; i < 256; i++ {
		tab8[i] = uint8(i&0x1) +
			(uint8(i>>1) & 0x1) +
			(uint8(i>>2) & 0x1) +
			(uint8(i>>3) & 0x1) +
			(uint8(i>>4) & 0x1) +
			(uint8(i>>5) & 0x1) +
			(uint8(i>>6) & 0x1) +
			(uint8(i>>7) & 0x1)
	}
}

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
	// Color is the bitmap's color interpretation.
	Color Color

	// Special instructions for I/O.
	Special int

	// Text string associated with the pix.
	Text string
	Box  image.Rectangle
}

// New creates new bitmap with the parameters as provided in the arguments.
func New(width, height int) *Bitmap {
	bm := newBitmap(width, height)
	bm.Data = make([]byte, height*bm.RowStride)
	return bm
}

func newBitmap(width, height int) *Bitmap {
	return &Bitmap{
		Width:     width,
		Height:    height,
		RowStride: (width + 7) >> 3,
	}
}

// NewWithData creates new bitmap with the provided 'width', 'height' and the byte slice 'data'.
func NewWithData(width, height int, data []byte) (*Bitmap, error) {
	bm := newBitmap(width, height)
	bm.Data = data
	if len(data) < height*bm.RowStride {
		return nil, fmt.Errorf("invalid data length: %d - should be: %d", len(data), height*bm.RowStride)
	}
	return bm, nil
}

// AddBorder creates a new bitmap with the border of size 'borderSize'. If the 'borderSize' is different than zero
// the resultant bitmap dimensions would increase by width += 2* borderSize, height += 2*borderSize.
// The value 'val' represents the binary bit 'value' of the border - '0' and '1'
func (b *Bitmap) AddBorder(borderSize, val int) (*Bitmap, error) {
	if borderSize == 0 {
		return b.Copy(), nil
	}
	return b.addBorderGeneral(borderSize, borderSize, borderSize, borderSize, val)
}

// AddBorderGeneral creates new bitmap on the base of the bitmap 'b' with the border of size for each side
// 'left','right','top','bot'. The 'val' sets the border white (0) or black (1).
func (b *Bitmap) AddBorderGeneral(left, right, top, bot int, val int) (*Bitmap, error) {
	return b.addBorderGeneral(left, right, top, bot, val)
}

// Copy gets a copy of the 'b' bitmap.
func (b *Bitmap) Copy() *Bitmap {
	data := make([]byte, len(b.Data))
	copy(data, b.Data)
	return &Bitmap{
		Width:        b.Width,
		Height:       b.Height,
		RowStride:    b.RowStride,
		Data:         data,
		Color:        b.Color,
		Text:         b.Text,
		BitmapNumber: b.BitmapNumber,
		Special:      b.Special,
		Box:          b.Box,
	}
}

// CountPixels counts the pixels for the bitmap 'b'.
func (b *Bitmap) CountPixels() int {
	return b.countPixels()
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

// Equivalent checks if the bitmaps 'b' and 's' are equivalent
// from the visual point of view.
func (b *Bitmap) Equivalent(s *Bitmap) bool {
	return b.equivalent(s)
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
	if b.Color == Vanilla {
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
	if b.Color == Chocolate {
		b.inverseData()
	}
	return b.Data
}

// InverseData inverses the data color interpretation.
func (b *Bitmap) InverseData() {
	b.inverseData()
}

// RemoveBorder create a new bitmap based on the 'b' with removed border of size 'borderSize'.
func (b *Bitmap) RemoveBorder(borderSize int) (*Bitmap, error) {
	if borderSize == 0 {
		return b.Copy(), nil
	}
	return b.removeBorderGeneral(borderSize, borderSize, borderSize, borderSize)
}

// RemoveBorderGeneral creates a new bitmap with removed border of size 'left', 'right', 'top', 'bot'.
// The resultant bitmap dimensions would be smaller by the value of border size.
func (b *Bitmap) RemoveBorderGeneral(left, right, top, bot int) (*Bitmap, error) {
	return b.removeBorderGeneral(left, right, top, bot)
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

// SizesEqual checks if the bitmaps are of the same size.
func (b *Bitmap) SizesEqual(s *Bitmap) bool {
	if b == s {
		return true
	}

	if b.Width != s.Width || b.Height != s.Height {
		return false
	}
	return true
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

func (b *Bitmap) addBorderGeneral(left, right, top, bot int, val int) (*Bitmap, error) {
	if left < 0 || right < 0 || top < 0 || bot < 0 {
		return nil, errors.New("negative border added")
	}

	ws, hs := b.Width, b.Height
	wd := ws + left + right
	hd := hs + top + bot

	bd := New(wd, hd)
	bd.Color = b.Color

	op := PixClr
	if val > 0 {
		op = PixSet
	}

	err := bd.RasterOperation(0, 0, left, hd, op, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	err = bd.RasterOperation(wd-right, 0, right, hd, op, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	err = bd.RasterOperation(0, 0, wd, top, op, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	err = bd.RasterOperation(0, hd-bot, wd, bot, op, nil, 0, 0)
	if err != nil {
		return nil, err
	}

	// copy the pixels into the interior
	err = bd.RasterOperation(left, top, ws, hs, PixSrc, b, 0, 0)
	if err != nil {
		return nil, err
	}
	return bd, nil
}

func (b *Bitmap) clearAll() error {
	return b.RasterOperation(0, 0, b.Width, b.Height, PixClr, nil, 0, 0)
}

func (b *Bitmap) countPixels() int {
	var (
		sum     int
		endmask uint8
		bt      byte
		btIndex int
	)
	fullBytes := b.RowStride
	padding := uint(b.Width & 0x07)
	if padding != 0 {
		endmask = uint8((0xff << (8 - padding)) & 0xff)
		fullBytes--
	}

	for y := 0; y < b.Height; y++ {
		for btIndex = 0; btIndex < fullBytes; btIndex++ {
			bt = b.Data[y*b.RowStride+btIndex]
			sum += int(tab8[bt])
		}
		if padding != 0 {
			sum += int(tab8[b.Data[y*b.RowStride+btIndex]&endmask])
		}
	}
	return sum
}

func (b *Bitmap) equivalent(s *Bitmap) bool {
	if b == s {
		return true
	}

	if !b.SizesEqual(s) {
		return false
	}

	// get the XOR of the bitmaps 'b' and 's'
	xor := combineBitmap(b, s, CmbOpXor)

	// count the pixels for the first bitmap.
	pixelCount := b.countPixels()
	thresh := int(0.25 * float32(pixelCount))

	// if the symbols are significantly different end fast.
	if xor.thresholdPixelSum(thresh) {
		return false
	}

	var (
		parsedPixCounts     [9][9]int
		horizontalParsedPix [18][9]int
		verticalParsedPix   [9][18]int
		horizontalModuleCtr int
		verticalModuleCtr   int
	)
	divider := 9
	verticalPart := b.Height / divider
	horizontalPart := b.Width / divider
	vp, hp := verticalPart/2, horizontalPart/2
	if verticalPart < horizontalPart {
		vp = horizontalPart / 2
		hp = verticalPart / 2
	}

	pointThresh := float64(vp) * float64(hp) * math.Pi
	vLineThresh := int(float64(verticalPart*horizontalPart/2) * 0.9)
	hLineThresh := int(float64(horizontalPart*verticalPart/2) * 0.9)

	for horizontalPosition := 0; horizontalPosition < divider; horizontalPosition++ {
		horizontalStart := horizontalPart*horizontalPosition + horizontalModuleCtr
		var horizontalEnd int

		if horizontalPosition == divider-1 {
			horizontalModuleCtr = 0
			horizontalEnd = b.Width
		} else {
			horizontalEnd = horizontalStart + horizontalPart
			if ((b.Width - horizontalModuleCtr) % divider) > 0 {
				horizontalModuleCtr++
				horizontalEnd++
			}
		}

		for verticalPosition := 0; verticalPosition < divider; verticalPosition++ {
			verticalStart := verticalPart*verticalPosition + verticalModuleCtr
			var verticalEnd int

			if verticalPosition == divider-1 {
				verticalModuleCtr = 0
				verticalEnd = b.Height
			} else {
				verticalEnd = verticalStart + verticalPart
				if (b.Height-verticalModuleCtr)%divider > 0 {
					verticalModuleCtr++
					verticalEnd++
				}
			}
			var leftCount, rightCount, downCount, upCount int

			horizontalCenter := (horizontalStart + horizontalEnd) / 2
			verticalCenter := (verticalStart + verticalEnd) / 2

			for h := horizontalStart; h < horizontalEnd; h++ {
				for v := verticalStart; v < verticalEnd; v++ {
					if xor.GetPixel(h, v) {
						if h < horizontalCenter {
							leftCount++
						} else {
							rightCount++
						}

						if v < verticalCenter {
							upCount++
						} else {
							downCount++
						}
					}
				}
			}

			parsedPixCounts[horizontalPosition][verticalPosition] = leftCount + rightCount

			horizontalParsedPix[horizontalPosition*2][verticalPosition] = leftCount
			horizontalParsedPix[horizontalPosition*2+1][verticalPosition] = rightCount

			verticalParsedPix[horizontalPosition][verticalPosition*2] = upCount
			verticalParsedPix[horizontalPosition][verticalPosition*2+1] = downCount
		}
	}

	for i := 0; i < divider*2-1; i++ {
		for j := 0; j < (divider - 1); j++ {
			var horizontalSum int

			for x := 0; x < 2; x++ {
				for y := 0; y < 2; y++ {
					horizontalSum += horizontalParsedPix[i+x][j+y]
				}
			}

			if horizontalSum > hLineThresh {
				return false
			}
		}
	}

	for i := 0; i < (divider - 1); i++ {
		for j := 0; j < ((divider * 2) - 1); j++ {
			var verticalSum int

			for x := 0; x < 2; x++ {
				for y := 0; y < 2; y++ {
					verticalSum += verticalParsedPix[i+x][j+y]
				}
			}

			if verticalSum > vLineThresh {
				return false
			}
		}
	}

	// check for cross lines

	for i := 0; i < (divider - 2); i++ {
		for j := 0; j < (divider - 2); j++ {
			var leftCross, rightCross int

			for x := 0; x < 3; x++ {
				for y := 0; y < 3; y++ {
					if x == y {
						leftCross += parsedPixCounts[i+x][j+y]
					}

					if (2 - x) == y {
						rightCross += parsedPixCounts[i+x][j+y]
					}
				}
			}

			if leftCross > hLineThresh || rightCross > hLineThresh {
				return false
			}
		}
	}

	for i := 0; i < (divider - 1); i++ {
		for j := 0; j < (divider - 1); j++ {
			var sum int

			for x := 0; x < 2; x++ {
				for y := 0; y < 2; y++ {
					sum += parsedPixCounts[i+x][j+y]
				}
			}

			if float64(sum) > pointThresh {
				return false
			}
		}
	}
	return true
}

func (b *Bitmap) inverseData() {
	b.RasterOperation(0, 0, b.Width, b.Height, PixNotDst, nil, 0, 0)
	// flip the color interpretation
	if b.Color == Chocolate {
		b.Color = Vanilla
	} else {
		b.Color = Chocolate
	}
}

func (b *Bitmap) removeBorderGeneral(left, right, top, bot int) (*Bitmap, error) {
	if left < 0 || right < 0 || top < 0 || bot < 0 {
		return nil, errors.New("negative broder remove values")
	}

	ws, hs := b.Width, b.Height
	wd := ws - left - right
	hd := hs - top - bot

	if wd <= 0 {
		return nil, errors.New("width must be > 0")
	}

	if hd <= 0 {
		return nil, errors.New("height must be > 0")
	}

	bm := New(wd, hd)
	bm.Color = b.Color

	err := bm.RasterOperation(0, 0, wd, hd, PixSrc, b, left, top)
	if err != nil {
		return nil, err
	}
	return bm, nil
}

func (b *Bitmap) setPadBits(val int) {
	endbits := 8 - b.Width%8
	if endbits == 0 {
		// no partial words
		return
	}
	fullBytes := b.Width / 8
	mask := rmaskByte[endbits]
	if val == 0 {
		mask ^= mask
	}

	var bIndex int
	for i := 0; i < b.Height; i++ {
		bIndex = i*b.RowStride + fullBytes
		if val == 0 {
			b.Data[bIndex] &= mask
		} else {
			b.Data[bIndex] |= mask
		}
	}
}

// thresholdPixelSum this function sums the '1' pixels and returns immediately
// if the count goes above threshold.
// Returns false if the sum is greater then 'thresh' threshold.
func (b *Bitmap) thresholdPixelSum(thresh int) bool {
	var (
		sum     int
		endmask uint8
		bt      byte
		btIndex int
	)
	fullBytes := b.RowStride
	padding := uint(b.Width & 0x07)
	if padding != 0 {
		endmask = uint8((0xff << (8 - padding)) & 0xff)
		fullBytes--
	}

	for y := 0; y < b.Height; y++ {
		for btIndex = 0; btIndex < fullBytes; btIndex++ {
			bt = b.Data[y*b.RowStride+btIndex]
			sum += int(tab8[bt])
		}

		if padding != 0 {
			bt = b.Data[y*b.RowStride+btIndex] & endmask
			sum += int(tab8[bt])
		}

		if sum > thresh {
			return true
		}
	}
	return false
}

// CorrelationScoreThresholded checks whether the correlation score is >= scoreThreshold.
func CorrelationScoreThresholded(bm1, bm2 *Bitmap, area1, area2 int, delX, delY float32, maxDiffW, maxDiffH int, tab, downcount []int, scoreThreshold float32) (bool, error) {
	if bm1 == nil {
		return false, errors.New("correlationScoreThresholded bm1 is nil")
	}
	if bm2 == nil {
		return false, errors.New("correlationScoreThresholded bm2 is nil")
	}

	if area1 <= 0 || area2 <= 0 {
		return false, errors.New("correlationScoreThresholded - areas must be > 0")
	}

	wi, hi := bm1.Width, bm1.Height
	wt, ht := bm2.Width, bm2.Height

	if abs(wi-wt) > maxDiffW {
		return false, nil
	}
	if abs(hi-ht) > maxDiffH {
		return false, nil
	}
	var idelX, idelY int
	if delX >= 0 {
		idelX = int(delX + 0.5)
	} else {
		idelX = int(delX - 0.5)
	}
	if delY >= 0 {
		idelY = int(delY + 0.5)
	} else {
		idelY = int(delY + 0.5)
	}

	// compute the correlation count.
	threshold := int(math.Ceil(math.Sqrt(float64(scoreThreshold) * float64(area1) * float64(area2))))
	var count int
	var rowBytes1 int
	rowBytes2 := bm2.RowStride

	// only the rows underlying the shifted bm2 need to be considered
	loRow := max(idelY, 0)
	hiRow := min(ht+idelY, hi)

	row1Index := bm1.RowStride * loRow
	row2Index := bm2.RowStride * (loRow - idelY)

	var untouchable int
	if hiRow <= hi {
		// some rows of bm1 would never contribute
		untouchable = downcount[hiRow-1]
	}

	loCol := max(idelX, 0)
	hiCol := min(wt+idelX, wi)
	var bm1LSkip, bm2LSkip int
	if idelX >= 8 {
		bm1LSkip = idelX >> 3
		row1Index += bm1LSkip
		loCol -= bm1LSkip << 3
		hiCol -= bm1LSkip << 3
		idelX &= 7
	} else if idelX <= -8 {
		bm2LSkip = -((idelX + 7) >> 3)
		row2Index += bm2LSkip
		rowBytes2 -= bm2LSkip
		idelX += bm2LSkip << 3
	}
	var (
		y, x                  int
		andByte, byte1, byte2 byte
	)
	if !(loCol >= hiCol || loRow >= hiRow) {
		rowBytes1 = (hiCol + 7) >> 3

		switch {
		case idelX == 0:
			for y = loRow; y < hiRow; y++ {
				for x = 0; x < rowBytes1; x++ {
					andByte = bm1.Data[row1Index+x] & bm2.Data[row2Index+x]
					count += tab[andByte]
				}
				if count >= threshold {
					return true, nil
				}
				if count+downcount[y]-untouchable < threshold {
					return false, nil
				}
				row1Index += bm1.RowStride
				row2Index += bm2.RowStride
			}
		case idelX > 0 && rowBytes2 < rowBytes1:
			for y = loRow; y < hiRow; y++ {
				byte1 = bm1.Data[row1Index]
				byte2 = bm2.Data[row2Index]
				andByte = byte1 & byte2
				count += tab[andByte]

				for x = 1; x < rowBytes2; x++ {
					byte1 = bm1.Data[row1Index+x]
					byte2 = (bm2.Data[row2Index+x]>>uint(idelX) | bm2.Data[row2Index+x-1]<<uint(8-delX))
					andByte = byte1 & byte2
					count += tab[andByte]
				}

				byte1 = bm1.Data[row1Index+x]
				byte2 = bm2.Data[row2Index-1] << uint(8-idelX)
				andByte = byte1 & byte2
				count += tab[andByte]

				if count >= threshold {
					return true, nil
				} else if count+downcount[y]-untouchable < threshold {
					return false, nil
				}
				row1Index += bm1.RowStride
				row2Index += bm2.RowStride
			}
		case idelX > 0 && rowBytes1 >= rowBytes2:
			for y = loRow; y < hiRow; y++ {
				for x = 0; x < rowBytes1; x++ {
					byte1 = bm1.Data[row1Index+x]
					byte2 = bm2.Data[row2Index+x] << uint(8-idelX)
					byte2 |= bm2.Data[row2Index+x+1] >> uint(8+idelX)
					andByte = byte1 & byte2
					count += tab[andByte]
				}
				if count >= threshold {
					return true, nil
				} else if count+downcount[y]-untouchable < threshold {
					return false, nil
				}
				row1Index += bm1.RowStride
				row2Index += bm2.RowStride
			}
		case rowBytes1 < rowBytes2:
			for y = loRow; y < hiRow; y++ {
				for x = 0; x < rowBytes1; x++ {
					byte1 = bm1.Data[row1Index+x]
					byte2 = bm2.Data[row2Index+x] << uint(8-idelX)
					byte2 |= bm2.Data[row2Index+x+1] >> uint(8+idelX)
					andByte = byte1 & byte2
					count += tab[andByte]
				}

				if count >= threshold {
					return true, nil
				} else if count+downcount[y]-untouchable < threshold {
					return false, nil
				}
				row1Index += bm1.RowStride
				row2Index += bm2.RowStride
			}
		case rowBytes2 >= rowBytes1:
			for y = loRow; y < hiRow; y++ {
				for x = 0; x < row1Index-1; x++ {
					byte1 = bm1.Data[row1Index+x]
					byte2 = bm2.Data[row2Index+x] << uint(8-idelX)
					byte2 |= bm2.Data[row2Index+x+1] >> uint(8+idelX)
					andByte = byte1 & byte2
					count += tab[andByte]
				}

				byte1 = bm1.Data[row1Index+x]
				byte2 = bm2.Data[row2Index+x] << uint(8-idelX)
				andByte = byte1 & byte2
				count += tab[andByte]

				if count >= threshold {
					return true, nil
				} else if count+downcount[y]-untouchable < threshold {
					return false, nil
				}

				row1Index += bm1.RowStride
				row2Index += bm2.RowStride
			}
		}
	}
	score := float32(count) * float32(count) / (float32(area1) * float32(area2))
	if score >= scoreThreshold {
		common.Log.Debug("count: %d < threshold %d but score %f >= scoreThreshold %f", count, threshold, score, scoreThreshold)
	}
	return false, nil

}

func abs(input int) int {
	if input < 0 {
		return -input
	}
	return input
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
