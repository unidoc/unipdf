/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"
	"image/color"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
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

	// The XResolution and YResolution are the
	// image resolution parameters at width and height.
	XResolution, YResolution int
}

// New creates new bitmap with the parameters as provided in the arguments.
func New(width, height int) *Bitmap {
	bm := newBitmap(width, height)
	bm.Data = make([]byte, height*bm.RowStride)
	return bm
}

// Copy the bitmap 's' into bitmap 'd'. If 'd' is nil, it is created by the function.
func Copy(d, s *Bitmap) (*Bitmap, error) {
	return copyBitmap(d, s)
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
	const processName = "NewWithData"
	bm := newBitmap(width, height)
	bm.Data = data
	if len(data) < height*bm.RowStride {
		return nil, errors.Errorf(processName, "invalid data length: %d - should be: %d", len(data), height*bm.RowStride)
	}
	return bm, nil
}

// NewWithUnpaddedData creates new bitmap with provided 'width', 'height' and the byte slice 'data' that doesn't
// contains paddings on the last byte of each row.
// This function adds the padding to the bitmap data.
func NewWithUnpaddedData(width, height int, data []byte) (*Bitmap, error) {
	const processName = "NewWithUnpaddedData"
	bm := newBitmap(width, height)
	bm.Data = data
	if dataLen := ((width * height) + 7) >> 3; len(data) < dataLen {
		return nil, errors.Errorf(processName, "invalid data length: '%d'. The data should contain at least: '%d' bytes", len(data), dataLen)
	}
	if err := bm.addPadBits(); err != nil {
		return nil, errors.Wrap(err, processName, "")
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
	d, err := b.addBorderGeneral(borderSize, borderSize, borderSize, borderSize, val)
	if err != nil {
		return nil, errors.Wrap(err, "AddBorder", "")
	}
	return d, nil
}

// AddBorderGeneral creates new bitmap on the base of the bitmap 'b' with the border of size for each side
// 'left','right','top','bot'. The 'val' sets the border white (0) or black (1).
func (b *Bitmap) AddBorderGeneral(left, right, top, bot int, val int) (*Bitmap, error) {
	return b.addBorderGeneral(left, right, top, bot, val)
}

// And does the raster operation 'AND' on the provided bitmaps 'b' and 's'.
func (b *Bitmap) And(s *Bitmap) (d *Bitmap, err error) {
	const processName = "Bitmap.And"
	if b == nil {
		return nil, errors.Error(processName, "'bitmap 'b' is nil")
	}

	if s == nil {
		return nil, errors.Error(processName, "bitmap 's' is nil")
	}

	if !b.SizesEqual(s) {
		common.Log.Debug("%s - Bitmap 's' is not equal size with 'b'", processName)
	}

	if d, err = copyBitmap(d, b); err != nil {
		return nil, errors.Wrap(err, processName, "can't create 'd' bitmap")
	}

	if err = d.RasterOperation(0, 0, d.Width, d.Height, PixSrcAndDst, s, 0, 0); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return d, nil
}

// ClipRectangle clips the 'b' Bitmap to the 'box' with relatively defined coordinates.
// If the box is not contained within the pix the 'box' is clipped to the 'b' size.
func (b *Bitmap) ClipRectangle(box *image.Rectangle) (d *Bitmap, boxC *image.Rectangle, err error) {
	const processName = "ClipRectangle"
	if box == nil {
		return nil, nil, errors.Error(processName, "box is not defined")
	}
	w, h := b.Width, b.Height
	sRect := image.Rect(0, 0, w, h)
	if !box.Overlaps(sRect) {
		return nil, nil, errors.Error(processName, "box doesn't overlap b")
	}
	boxCT := box.Intersect(sRect)

	bx, by := boxCT.Min.X, boxCT.Min.Y
	bw, bh := boxCT.Dx(), boxCT.Dy()

	d = New(bw, bh)
	d.Text = b.Text
	if err = d.RasterOperation(0, 0, bw, bh, PixSrc, b, bx, by); err != nil {
		return nil, nil, errors.Wrap(err, processName, "PixSrc to clipped")
	}
	boxC = &boxCT
	return d, boxC, nil
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
	}
}

// CountPixels counts the pixels for the bitmap 'b'.
func (b *Bitmap) CountPixels() int {
	return b.countPixels()
}

// CreateTemplate creates a copy template bitmap on the base of the
// bitmap 'b'. A template has the same parameters as bitmap 'b', but contains empty Data.
func (b *Bitmap) CreateTemplate() *Bitmap {
	return b.createTemplate()
}

// Equals checks if all the pixels in the 'b' bitmap are equals to the 's' bitmap.
func (b *Bitmap) Equals(s *Bitmap) bool {
	if len(b.Data) != len(s.Data) || b.Width != s.Width || b.Height != s.Height {
		return false
	}

	for y := 0; y < b.Height; y++ {
		lineIndex := y * b.RowStride
		for i := 0; i < b.RowStride; i++ {
			if b.Data[lineIndex+i] != s.Data[lineIndex+i] {
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
		return 0, errors.Errorf("GetByte", "index: %d out of range", index)
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

	const processName = "GetUnpaddedData"
	for y := 0; y < b.Height; y++ {
		// btIndex is the byte index per row.
		for btIndex := 0; btIndex < b.RowStride; btIndex++ {
			bt := b.Data[y*b.RowStride+btIndex]
			if btIndex != b.RowStride-1 {
				err := w.WriteByte(bt)
				if err != nil {
					return nil, errors.Wrap(err, processName, "")
				}
				continue
			}

			for i := uint(0); i < padding; i++ {
				err := w.WriteBit(int(bt >> (7 - i) & 0x01))
				if err != nil {
					return nil, errors.Wrap(err, processName, "")
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
	d, err := b.removeBorderGeneral(borderSize, borderSize, borderSize, borderSize)
	if err != nil {
		return nil, errors.Wrap(err, "RemoveBorder", "")
	}
	return d, nil
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
		return errors.Errorf("SetPixel", "index out of range: %d", i)
	}
	o := b.GetBitOffset(x)

	shift := uint(7 - o)
	src := b.Data[i]
	var result byte
	if pixel == 1 {
		result = src | (pixel & 0x01 << shift)
	} else {
		result = src & ^(1 << shift)
	}
	b.Data[i] = result

	return nil
}

// SetByte sets the byte at 'index' with value 'v'.
// Returns an error if the index is out of range.
func (b *Bitmap) SetByte(index int, v byte) error {
	if index > len(b.Data)-1 || index < 0 {
		return errors.Errorf("SetByte", "index out of range: %d", index)
	}

	b.Data[index] = v
	return nil
}

// SetDefaultPixel sets all bits within bitmap to '1'.
func (b *Bitmap) SetDefaultPixel() {
	for i := range b.Data {
		b.Data[i] = byte(0xff)
	}
}

// SetPadBits sets the pad bits for the current bitmap.
func (b *Bitmap) SetPadBits(value int) {
	b.setPadBits(value)
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

// ThresholdPixelSum checks if the number of the 'ON' pixels is above the provided 'thresh' threshold.
// If the on pixel count > thresh the function returns quckly.
func (b *Bitmap) ThresholdPixelSum(thresh int, tab8 []int) (above bool, err error) {
	const processName = "Bitmap.ThresholdPixelSum"
	if tab8 == nil {
		tab8 = makePixelSumTab8()
	}

	fullBytes := b.Width >> 3
	endBits := b.Width & 7
	endMask := byte(0xff << uint(8-endBits))
	var (
		i, j, lineIndex, count int
		bt                     byte
	)
	for i = 0; i < b.Height; i++ {
		lineIndex = b.RowStride * i
		for j = 0; j < fullBytes; j++ {
			bt, err = b.GetByte(lineIndex + j)
			if err != nil {
				return false, errors.Wrap(err, processName, "fullByte")
			}
			count += tab8[bt]
		}
		if endBits != 0 {
			bt, err = b.GetByte(lineIndex + j)
			if err != nil {
				return false, errors.Wrap(err, processName, "partialByte")
			}
			bt &= endMask
			count += tab8[bt]
		}
		if count > thresh {
			return true, nil
		}
	}
	return above, nil
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

// Zero check if there is no 'ONE' pixels.
func (b *Bitmap) Zero() bool {
	fullBytes := b.Width / 8
	endBits := b.Width & 7
	var endMask byte
	if endBits != 0 {
		endMask = byte(0xff << uint(8-endBits))
	}

	var line, i, j int
	for i = 0; i < b.Height; i++ {
		line = b.RowStride * i
		for j = 0; j < fullBytes; j, line = j+1, line+1 {
			if b.Data[line] != 0 {
				return false
			}
		}
		if endBits > 0 {
			if b.Data[line]&endMask != 0 {
				return false
			}
		}
	}
	return true
}

func (b *Bitmap) addBorderGeneral(left, right, top, bot int, val int) (*Bitmap, error) {
	const processName = "addBorderGeneral"
	if left < 0 || right < 0 || top < 0 || bot < 0 {
		return nil, errors.Error(processName, "negative border added")
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
		return nil, errors.Wrap(err, processName, "left")
	}
	err = bd.RasterOperation(wd-right, 0, right, hd, op, nil, 0, 0)
	if err != nil {
		return nil, errors.Wrap(err, processName, "right")
	}
	err = bd.RasterOperation(0, 0, wd, top, op, nil, 0, 0)
	if err != nil {
		return nil, errors.Wrap(err, processName, "top")
	}
	err = bd.RasterOperation(0, hd-bot, wd, bot, op, nil, 0, 0)
	if err != nil {
		return nil, errors.Wrap(err, processName, "bottom")
	}

	// copy the pixels into the interior
	err = bd.RasterOperation(left, top, ws, hs, PixSrc, b, 0, 0)
	if err != nil {
		return nil, errors.Wrap(err, processName, "copy")
	}
	return bd, nil
}

// addPadBits creates new data byte slice that contains extra padding on the last byte for each row.
func (b *Bitmap) addPadBits() (err error) {
	const processName = "bitmap.addPadBits"
	endbits := b.Width % 8
	if endbits == 0 {
		// no partial words
		return nil
	}
	fullBytes := b.Width / 8
	//	mask := rmaskByte[endbits]

	r := reader.New(b.Data)
	data := make([]byte, b.Height*b.RowStride)
	w := writer.NewMSB(data)
	temp := make([]byte, fullBytes)
	var (
		i    int
		bits uint64
	)
	for i = 0; i < b.Height; i++ {
		// iterate over full bytes
		if _, err = r.Read(temp); err != nil {
			return errors.Wrap(err, processName, "full byte")
		}
		if _, err = w.Write(temp); err != nil {
			return errors.Wrap(err, processName, "full bytes")
		}
		// read unused bits
		if bits, err = r.ReadBits(byte(endbits)); err != nil {
			return errors.Wrap(err, processName, "skipping bits")
		}

		if err = w.WriteByte(byte(bits) << uint(8-endbits)); err != nil {
			return errors.Wrap(err, processName, "last byte")
		}
	}
	b.Data = w.Data()
	return nil
}

func (b *Bitmap) clearAll() error {
	return b.RasterOperation(0, 0, b.Width, b.Height, PixClr, nil, 0, 0)
}

// clipRectangle clips the bitmap 'b' with the bounds provided by the 'box' argument.
// The optional 'clipped' image.Rectangle argument would be changed for the actual box of clipped bitmap.
// The function returns the clipped bitmap. If the 'box' doesn't intersect with the 'b' bitmap the function returns nil bitmap.
func (b *Bitmap) clipRectangle(box, boxClipped *image.Rectangle) (clipped *Bitmap, err error) {
	const processName = "clipRectangle"
	if box == nil {
		return nil, errors.Error(processName, "provided nil 'box'")
	}
	w, h := b.Width, b.Height
	clippedTemp, err := ClipBoxToRectangle(box, w, h)
	if err != nil {
		common.Log.Warning("'box' doesn't overlap bitmap 'b': %v", err)
		return nil, nil
	}
	bx, by := clippedTemp.Min.X, clippedTemp.Min.Y
	bw, bh := clippedTemp.Max.X-clippedTemp.Min.X, clippedTemp.Max.Y-clippedTemp.Min.Y

	clipped = New(bw, bh)

	clipped.Text = b.Text
	if err = clipped.RasterOperation(0, 0, bw, bh, PixSrc, b, bx, by); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	if boxClipped != nil {
		*boxClipped = *clippedTemp
	}
	return clipped, nil
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

func (b *Bitmap) createTemplate() *Bitmap {
	return &Bitmap{
		Width:        b.Width,
		Height:       b.Height,
		RowStride:    b.RowStride,
		Color:        b.Color,
		Text:         b.Text,
		BitmapNumber: b.BitmapNumber,
		Special:      b.Special,
		Data:         make([]byte, len(b.Data)),
	}
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
	if err := b.RasterOperation(0, 0, b.Width, b.Height, PixNotDst, nil, 0, 0); err != nil {
		common.Log.Debug("Inverse data failed: '%v'", err)
	}
	// flip the color interpretation
	if b.Color == Chocolate {
		b.Color = Vanilla
	} else {
		b.Color = Chocolate
	}
}

// nextOnPixel
func (b *Bitmap) nextOnPixel(xStart, yStart int) (pt image.Point, ok bool, err error) {
	const processName = "nextOnPixel"
	pt, ok, err = b.nextOnPixelLow(b.Width, b.Height, b.RowStride, xStart, yStart)
	if err != nil {
		return pt, false, errors.Wrap(err, processName, "")
	}
	return pt, ok, nil
}

func (b *Bitmap) nextOnPixelLow(w, h, wpl, xStart, yStart int) (pt image.Point, ok bool, err error) {
	const processName = "Bitmap.nextOnPixelLow"
	var (
		x  int
		bt byte
	)
	line := yStart * wpl
	index := line + (xStart / 8)

	if bt, err = b.GetByte(index); err != nil {
		return pt, false, errors.Wrap(err, processName, "xStart and yStart out of range")
	}

	// if the 'bt' byte is different than 0, it must contain 'ON' pixel.
	// search the byte place.
	if bt != 0 {
		xEnd := xStart - (xStart % 8) + 7
		for x = xStart; x <= xEnd && x < w; x++ {
			if b.GetPixel(x, yStart) {
				pt.X = x
				pt.Y = yStart
				return pt, true, nil
			}
		}
	}

	// continue with the rest of the line.
	startByte := (xStart / 8) + 1
	x = 8 * startByte
	var i int
	for index = line + startByte; x < w; index, x = index+1, x+8 {
		if bt, err = b.GetByte(index); err != nil {
			return pt, false, errors.Wrap(err, processName, "rest of the line byte")
		}
		if bt == 0 {
			continue
		}
		for i = 0; i < 8 && x < w; i, x = i+1, x+1 {
			if b.GetPixel(x, yStart) {
				pt.X = x
				pt.Y = yStart
				return pt, true, nil
			}
		}
	}
	// search till the end of the bitmap data.
	for y := yStart + 1; y < h; y++ {
		line = y * wpl
		for index, x = line, 0; x < w; index, x = index+1, x+8 {
			if bt, err = b.GetByte(index); err != nil {
				return pt, false, errors.Wrap(err, processName, "following lines")
			}

			// if the byte is different than 0x00 it must contain at least one 'ON' bit.
			if bt == 0 {
				continue
			}
			for i = 0; i < 8 && x < w; i, x = i+1, x+1 {
				if b.GetPixel(x, y) {
					pt.X = x
					pt.Y = y
					return pt, true, nil
				}
			}
		}
	}
	return pt, false, nil
}

func (b *Bitmap) removeBorderGeneral(left, right, top, bot int) (*Bitmap, error) {
	const processName = "removeBorderGeneral"
	if left < 0 || right < 0 || top < 0 || bot < 0 {
		return nil, errors.Error(processName, "negative broder remove values")
	}

	ws, hs := b.Width, b.Height
	wd := ws - left - right
	hd := hs - top - bot

	if wd <= 0 {
		return nil, errors.Errorf(processName, "width: %d must be > 0", wd)
	}

	if hd <= 0 {
		return nil, errors.Errorf(processName, "height: %d must be > 0", hd)
	}

	bm := New(wd, hd)
	bm.Color = b.Color

	err := bm.RasterOperation(0, 0, wd, hd, PixSrc, b, left, top)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return bm, nil
}

func (b *Bitmap) resizeImageData(s *Bitmap) error {
	if s == nil {
		return errors.Error("resizeImageData", "src is not defined")
	}
	if b.SizesEqual(s) {
		return nil
	}

	b.Data = make([]byte, len(s.Data))
	b.Width = s.Width
	b.Height = s.Height
	b.RowStride = s.RowStride
	// NOTE: if resolution included, set also resolution
	return nil
}

func (b *Bitmap) setAll() error {
	err := rasterOperation(b, 0, 0, b.Width, b.Height, PixSet, nil, 0, 0)
	if err != nil {
		return errors.Wrap(err, "setAll", "")
	}
	return nil
}

func (b *Bitmap) setBit(index int) {
	b.Data[(index >> 3)] |= 0x80 >> uint(index&7)
}

func (b *Bitmap) setPadBits(val int) {
	endbits := 8 - b.Width%8
	if endbits == 8 {
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

func (b *Bitmap) setTwoBytes(index int, tb uint16) error {
	if index+1 > len(b.Data)-1 {
		return errors.Errorf("setTwoBytes", "index: '%d' out of range", index)
	}
	b.Data[index] = byte((tb & 0xff00) >> 8)
	b.Data[index+1] = byte(tb & 0xff)
	return nil
}

func (b *Bitmap) setFourBytes(index int, fb uint32) error {
	if index+3 > len(b.Data)-1 {
		return errors.Errorf("setFourBytes", "index: '%d' out of range", index)
	}
	b.Data[index] = byte((fb & 0xff000000) >> 24)
	b.Data[index+1] = byte((fb & 0xff0000) >> 16)
	b.Data[index+2] = byte((fb & 0xff00) >> 8)
	b.Data[index+3] = byte(fb & 0xff)
	return nil
}

func (b *Bitmap) setEightBytes(index int, eb uint64) error {
	fullBytesNumber := b.RowStride - (index % b.RowStride)
	if b.RowStride != b.Width>>3 {
		fullBytesNumber--
	}
	if fullBytesNumber >= 8 {
		return b.setEightFullBytes(index, eb)
	}
	return b.setEightPartlyBytes(index, fullBytesNumber, eb)
}

func (b *Bitmap) setEightPartlyBytes(index, fullBytesNumber int, eb uint64) (err error) {
	var (
		temp  byte
		shift int
	)
	const processName = "setEightPartlyBytes"
	for i := 1; i <= fullBytesNumber; i++ {
		// eb >> 7 * 8 = 56
		// 4 + 7 - 7
		shift = 64 - i*8
		temp = byte(eb >> uint(shift) & 0xff)
		common.Log.Trace("temp: %08b, index: %d, idx: %d, fullBytesNumber: %d, shift: %d", temp, index, index+i-1, fullBytesNumber, shift)
		if err = b.SetByte(index+i-1, temp); err != nil {
			return errors.Wrap(err, processName, "fullByte")
		}
	}
	padding := b.RowStride*8 - b.Width
	if padding == 0 {
		return nil
	}
	shift -= 8
	temp = byte(eb>>uint(shift)&0xff) << uint(padding)
	if err = b.SetByte(index+fullBytesNumber, temp); err != nil {
		return errors.Wrap(err, processName, "padded")
	}
	return nil
}

func (b *Bitmap) setEightFullBytes(index int, eb uint64) error {
	if index+7 > len(b.Data)-1 {
		return errors.Error("setEightBytes", "index out of range")
	}

	b.Data[index] = byte((eb & 0xff00000000000000) >> 56)
	b.Data[index+1] = byte((eb & 0xff000000000000) >> 48)
	b.Data[index+2] = byte((eb & 0xff0000000000) >> 40)
	b.Data[index+3] = byte((eb & 0xff00000000) >> 32)
	b.Data[index+4] = byte((eb & 0xff000000) >> 24)
	b.Data[index+5] = byte((eb & 0xff0000) >> 16)
	b.Data[index+6] = byte((eb & 0xff00) >> 8)
	b.Data[index+7] = byte(eb & 0xff)
	return nil
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

func copyBitmap(d, s *Bitmap) (*Bitmap, error) {
	if s == nil {
		return nil, errors.Error("copyBitmap", "source not defined")
	}
	if s == d {
		return d, nil
	}

	if d == nil {
		// create new bitmap if 'd' is nil.
		d = s.createTemplate()
		copy(d.Data, s.Data)
		return d, nil
	}
	err := d.resizeImageData(s)
	if err != nil {
		return nil, errors.Wrap(err, "copyBitmap", "")
	}
	d.Text = s.Text
	copy(d.Data, s.Data)
	return d, nil
}

func xor(d, b1, b2 *Bitmap) (*Bitmap, error) {
	const processName = "bitmap.xor"

	if b1 == nil {
		return nil, errors.Error(processName, "'b1' is nil")
	}
	if b2 == nil {
		return nil, errors.Error(processName, "'b2' is nil")
	}

	if d == b2 {
		return nil, errors.Error(processName, "'d' == 'b2'")
	}

	if !b1.SizesEqual(b2) {
		common.Log.Debug("%s - Bitmap 'b1' is not equal size with 'b2'", processName)
	}
	var err error
	if d, err = copyBitmap(d, b1); err != nil {
		return nil, errors.Wrap(err, processName, "can't create 'd'")
	}

	if err = d.RasterOperation(0, 0, d.Width, d.Height, PixSrcXorDst, b2, 0, 0); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return d, nil
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

func subtract(d, s1, s2 *Bitmap) (*Bitmap, error) {
	const processName = "subtract"
	if s1 == nil {
		return nil, errors.Error(processName, "'s1' is nil")
	}
	if s2 == nil {
		return nil, errors.Error(processName, "'s2' is nil")
	}

	var err error
	switch {
	case d == s1:
		if err = d.RasterOperation(0, 0, s1.Width, s1.Height, PixNotSrcAndDst, s2, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "d == s1")
		}
	case d == s2:
		if err = d.RasterOperation(0, 0, s1.Width, s1.Height, PixNotSrcAndDst, s1, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "d == s2")
		}
	default:
		d, err = copyBitmap(d, s1)
		if err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if err = d.RasterOperation(0, 0, s1.Width, s1.Height, PixNotSrcAndDst, s2, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "default")
		}
	}
	return d, nil
}
