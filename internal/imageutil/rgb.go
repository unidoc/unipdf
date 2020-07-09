package imageutil

import (
	"errors"
	"fmt"
	"image"
	"image/color"
)

// NRGBA is the interface for all NRGBA images.
type NRGBA interface {
	// NRGBAAt gets the NRGBA color at the coordinates 'x', 'y'.
	NRGBAAt(x, y int) color.NRGBA
	// SetNRGBA sets the NRGBA color at the coordinates 'x', 'y'.
	SetNRGBA(x, y int, c color.NRGBA)
}

// RGBA is the interface that allows to get and set RGBA images.
type RGBA interface {
	// RGBAAt gets the RGBA color at the coordinates 'x', 'y'.
	RGBAAt(x, y int) color.RGBA
	// SetRGBA sets the RGBA color at the coordinates 'x', 'y'.
	SetRGBA(x, y int, c color.RGBA)
}

// NRGBA16 implements RGB image with 4 bits.
type NRGBA16 struct {
	ImageBase
}

//
// NRGBA16 Image interface methods.
//

// Compile time check if NRGBA16 implements Image interface.
var _ Image = &NRGBA16{}

// Base implements Image interface.
func (i *NRGBA16) Base() *ImageBase {
	return &i.ImageBase
}

// Copy implements Image interface.
func (i *NRGBA16) Copy() Image {
	return &NRGBA16{ImageBase: i.copy()}
}

// ColorAt implements Image interface.
func (i *NRGBA16) ColorAt(x, y int) (color.Color, error) {
	return ColorAtNRGBA16(x, y, i.Width, i.BytesPerLine, i.Data, i.Alpha, i.Decode)
}

// Set implements draw.Image interface.
func (i *NRGBA16) Set(x, y int, c color.Color) {
	idx := y*i.BytesPerLine + x*3/2
	if idx+1 >= len(i.Data) {
		return
	}
	cr := NRGBA16Model.Convert(c).(color.NRGBA)
	i.setNRGBA(x, y, idx, cr)
}

func (i *NRGBA16) setNRGBA(x, y, idx int, c color.NRGBA) {
	if x*3%2 == 0 {
		i.Data[idx] = (c.R>>4)<<4 | (c.G >> 4)
		i.Data[idx+1] = (c.B>>4)<<4 | (i.Data[idx+1] & 0xf)
	} else {
		i.Data[idx] = (i.Data[idx] & 0xf0) | (c.R >> 4)
		i.Data[idx+1] = (c.G>>4)<<4 | (c.B >> 4)
	}
	if i.Alpha != nil {
		aIdx := y * BytesPerLine(i.Width, 4, 1)
		if aIdx < len(i.Alpha) {
			if x%2 == 0 {
				i.Alpha[aIdx] = (c.A>>uint(4))<<uint(4) | (i.Alpha[idx] & 0xf)
			} else {
				i.Alpha[aIdx] = (i.Alpha[aIdx] & 0xf0) | (c.A >> uint(4))
			}
		}
	}
}

// Validate implements Image interface.
func (i *NRGBA16) Validate() error {
	if len(i.Data) != 3*i.Width*i.Height/2 {
		return errors.New("invalid image data size for provided dimensions")
	}
	return nil
}

//
// NRGBA16 image.Image interface methods.
//

// Compile time check if NRGBA16 implements image.Image interface.
var _ image.Image = &NRGBA16{}

// At implements Image interface.
func (i *NRGBA16) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

// ColorModel implements image.Image interface.
func (i *NRGBA16) ColorModel() color.Model {
	return NRGBA16Model
}

func nrgba16Model(c color.Color) color.Color {
	cr := color.NRGBAModel.Convert(c).(color.NRGBA)
	return nrgba16ModelNRGBA(cr)
}

func nrgba16ModelNRGBA(cr color.NRGBA) color.NRGBA {
	cr.R = cr.R>>4 | (cr.R>>4)<<4
	cr.G = cr.G>>4 | (cr.G>>4)<<4
	cr.B = cr.B>>4 | (cr.B>>4)<<4
	return cr
}

// Bounds implements image.Image interface.
func (i *NRGBA16) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

//
// NRGBA16 NRGBA interface methods.
//

// Compile time check if NRGBA16 implements NRGBA interface.
var _ NRGBA = &NRGBA16{}

// NRGBAAt implements NRGBA interface.
func (i *NRGBA16) NRGBAAt(x, y int) color.NRGBA {
	c, _ := ColorAtNRGBA16(x, y, i.Width, i.BytesPerLine, i.Data, i.Alpha, i.Decode)
	return c
}

// SetNRGBA implements NRGBA interface.
func (i *NRGBA16) SetNRGBA(x, y int, c color.NRGBA) {
	idx := y*i.BytesPerLine + x*3/2
	if idx+1 >= len(i.Data) {
		return
	}
	c = nrgba16ModelNRGBA(c)
	i.setNRGBA(x, y, idx, c)
}

// ColorAtNRGBA16 gets the 4 bit per component NRGBA32 color at 'x', 'y' coordinates.
func ColorAtNRGBA16(x, y, width, bytesPerLine int, data, alpha []byte, decode []float64) (color.NRGBA, error) {
	// Index in the data is equal to rowIndex (y*img.BytesPerLine) + the numbers of bytes to the right.
	idx := y*bytesPerLine + x*3/2
	if idx+1 >= len(data) {
		return color.NRGBA{}, errInvalidCoordinates(x, y)
	}

	const (
		max4bitValue = 0xf
		max8BitValue = uint8(0xff)
	)
	a := max8BitValue
	if alpha != nil {
		aIdx := y * BytesPerLine(width, 4, 1)
		if aIdx < len(alpha) {
			if x%2 == 0 {
				a = (alpha[aIdx] >> uint(4)) & max4bitValue
			} else {
				a = alpha[aIdx] & max4bitValue
			}
			a |= a << 4
		}
	}
	// Calculate bit position at which the color data starts.
	var r, g, b uint8
	if x*3%2 == 0 {
		// The R and G components are contained by the current byte
		// and the B component is contained by the next byte.
		r = (data[idx] >> uint(4)) & max4bitValue
		g = data[idx] & max4bitValue
		b = (data[idx+1] >> uint(4)) & max4bitValue
	} else {
		// The R component is contained by the current byte and the
		// G and B components are contained by the next byte.
		r = data[idx] & max4bitValue
		g = (data[idx+1] >> uint(4)) & max4bitValue
		b = data[idx+1] & max4bitValue
	}

	if len(decode) == 6 {
		r = uint8(uint32(LinearInterpolate(float64(r), 0, 15, decode[0], decode[1])) & 0xf)
		g = uint8(uint32(LinearInterpolate(float64(g), 0, 15, decode[2], decode[3])) & 0xf)
		b = uint8(uint32(LinearInterpolate(float64(b), 0, 15, decode[4], decode[5])) & 0xf)
	}

	return color.NRGBA{
		R: (r << 4) | (r & 0xf),
		G: (g << 4) | (g & 0xf),
		B: (b << 4) | (b & 0xf),
		A: a,
	}, nil
}

func errInvalidCoordinates(x int, y int) error {
	return fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
}

func nrgba16Converter(src image.Image) (Image, error) {
	if i, ok := src.(*NRGBA16); ok {
		return i.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 4, 3, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	nrgbaConvertImage(src, g, bounds)
	return g, nil
}

//
// NRGBA32 image.
//

// Compile time check if NRGBA32 implements Image interface.
var _ Image = &NRGBA32{}

// NRGBA32 implements RGB image with 4 bits.
type NRGBA32 struct {
	ImageBase
}

// Base implements Image interface.
func (i *NRGBA32) Base() *ImageBase {
	return &i.ImageBase
}

// ColorAt implements Image interface.
func (i *NRGBA32) ColorAt(x, y int) (color.Color, error) {
	return ColorAtNRGBA32(x, y, i.Width, i.Data, i.Alpha, i.Decode)
}

// Copy implements Image interface.
func (i *NRGBA32) Copy() Image {
	return &NRGBA32{ImageBase: i.copy()}
}

// Set implements draw.Image interface.
func (i *NRGBA32) Set(x, y int, c color.Color) {
	idx := y*i.Width + x
	// We need three consecutive bytes in order to get the right components.
	index := 3 * idx
	if index+2 >= len(i.Data) {
		return
	}
	cr := color.NRGBAModel.Convert(c).(color.NRGBA)
	i.setRGBA(idx, cr)
}

// Validate implements Image interface.
func (i *NRGBA32) Validate() error {
	if len(i.Data) != 3*i.Width*i.Height {
		return errors.New("invalid image data size for provided dimensions")
	}
	return nil
}

//
// NRGBA32 image.Image interface methods.
//

// Compile time check if NRGBA32 implements image.Image interface.
var _ image.Image = &NRGBA32{}

// ColorModel implements image.Image interface.
func (i *NRGBA32) ColorModel() color.Model {
	return color.NRGBAModel
}

// Bounds implements image.Image interface.
func (i *NRGBA32) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

// At implements Image interface.
func (i *NRGBA32) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

//
// NRGBA32 NRGBA interface methods.
//

// Compile time check if NRGBA32 implements NRGBA interface.
var _ NRGBA = &NRGBA32{}

// NRGBAAt implements NRGBA implements.
func (i *NRGBA32) NRGBAAt(x, y int) color.NRGBA {
	c, _ := ColorAtNRGBA32(x, y, i.Width, i.Data, i.Alpha, i.Decode)
	return c
}

// SetNRGBA implements NRGBA interface.
func (i *NRGBA32) SetNRGBA(x, y int, c color.NRGBA) {
	idx := y*i.Width + x
	// We need three consecutive bytes in order to get the right components.
	index := 3 * idx
	if index+2 >= len(i.Data) {
		return
	}
	i.setRGBA(idx, c)
	return
}

func (i *NRGBA32) setRGBA(idx int, c color.NRGBA) {
	index := 3 * idx
	i.Data[index] = c.R
	i.Data[index+1] = c.G
	i.Data[index+2] = c.B
	if idx < len(i.Alpha) {
		i.Alpha[idx] = c.A
	}
}

// ColorAtNRGBA32 gets the 8 bit per component NRGBA32 color at 'x', 'y'.
func ColorAtNRGBA32(x, y, width int, data, alpha []byte, decode []float64) (color.NRGBA, error) {
	idx := y*width + x
	// We need three consecutive bytes in order to get the right components.
	i := 3 * idx
	if i+2 >= len(data) {
		return color.NRGBA{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	a := uint8(0xff)
	if alpha != nil && len(alpha) > idx {
		a = alpha[idx]
	}

	r, g, b := data[i], data[i+1], data[i+2]
	if len(decode) == 6 {
		r = uint8(uint32(LinearInterpolate(float64(r), 0, 255, decode[0], decode[1])) & 0xff)
		g = uint8(uint32(LinearInterpolate(float64(g), 0, 255, decode[2], decode[3])) & 0xff)
		b = uint8(uint32(LinearInterpolate(float64(b), 0, 255, decode[4], decode[5])) & 0xff)
	}
	return color.NRGBA{R: r, G: g, B: b, A: a}, nil
}

func nrgbaConverter(src image.Image) (Image, error) {
	if i, ok := src.(*NRGBA32); ok {
		return i.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 8, 3, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	nrgbaConvertImage(src, g, bounds)
	return g, nil
}

// Compile time check if NRGBA64 implements Image interface.
var _ Image = &NRGBA64{}

// NRGBA64 implements RGB image with 4 bits.
type NRGBA64 struct {
	ImageBase
}

//
// NRGBA64 Image interface methods.
//

// Base implements Image interface.
func (i *NRGBA64) Base() *ImageBase {
	return &i.ImageBase
}

// Set implements draw.Image interface.
func (i *NRGBA64) Set(x, y int, c color.Color) {
	idx := (y*i.Width + x) * 2
	index := idx * 3
	if index+5 >= len(i.Data) {
		return
	}
	cr := color.NRGBA64Model.Convert(c).(color.NRGBA64)
	i.setNRGBA64(index, cr, idx)
}

func (i *NRGBA64) setNRGBA64(index int, cr color.NRGBA64, idx int) {
	i.Data[index] = uint8(cr.R >> 8)
	i.Data[index+1] = uint8(cr.R & 0xff)
	i.Data[index+2] = uint8(cr.G >> 8)
	i.Data[index+3] = uint8(cr.G & 0xff)
	i.Data[index+4] = uint8(cr.B >> 8)
	i.Data[index+5] = uint8(cr.B & 0xff)

	if idx+1 < len(i.Alpha) {
		i.Alpha[idx] = uint8(cr.A >> 8)
		i.Alpha[idx+1] = uint8(cr.A & 0xff)
	}
}

// SetNRGBA64 sets the 'NRGBA64' color at 'x' and 'y' coordinates.
func (i *NRGBA64) SetNRGBA64(x, y int, c color.NRGBA64) {
	idx := (y*i.Width + x) * 2
	index := idx * 3
	if index+5 >= len(i.Data) {
		return
	}

	i.setNRGBA64(index, c, idx)
}

// NRGBA64At gets color.NRGBA64 at 'x' and 'y'.
func (i *NRGBA64) NRGBA64At(x, y int) color.NRGBA64 {
	c, _ := ColorAtNRGBA64(x, y, i.Width, i.Data, i.Alpha, i.Decode)
	return c
}

// Copy implements Image interface.
func (i *NRGBA64) Copy() Image {
	return &NRGBA64{ImageBase: i.copy()}
}

// ColorAt implements Image interface.
func (i *NRGBA64) ColorAt(x, y int) (color.Color, error) {
	return ColorAtNRGBA64(x, y, i.Width, i.Data, i.Alpha, i.Decode)
}

// Validate implements Image interface.
func (i *NRGBA64) Validate() error {
	if len(i.Data) != 3*2*i.Width*i.Height {
		return errors.New("invalid image data size for provided dimensions")
	}
	return nil
}

//
// NRGBA64 image.Image interface methods.
//

// Compile time check if NRGBA64 implements image.Image interface.
var _ image.Image = &NRGBA64{}

// ColorModel implements image.Image interface.
func (i *NRGBA64) ColorModel() color.Model {
	return color.NRGBA64Model
}

// Bounds implements image.Image interface.
func (i *NRGBA64) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

// At implements image.Image interface.
func (i *NRGBA64) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

// ColorAtNRGBA64 gets 16 bit per component NRGBA64 color.
func ColorAtNRGBA64(x, y, width int, data, alpha []byte, decode []float64) (color.NRGBA64, error) {
	idx := (y*width + x) * 2

	index := idx * 3
	if index+5 >= len(data) {
		return color.NRGBA64{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	const max16Bit = 0xffff
	a := uint16(max16Bit)
	if alpha != nil && len(alpha) > idx+1 {
		a = uint16(alpha[idx])<<8 | uint16(alpha[idx+1])
	}

	r := uint16(data[index])<<8 | uint16(data[index+1])
	g := uint16(data[index+2])<<8 | uint16(data[index+3])
	b := uint16(data[index+4])<<8 | uint16(data[index+5])
	if len(decode) == 6 {
		r = uint16(uint64(LinearInterpolate(float64(r), 0, 65535, decode[0], decode[1])) & max16Bit)
		g = uint16(uint64(LinearInterpolate(float64(g), 0, 65535, decode[2], decode[3])) & max16Bit)
		b = uint16(uint64(LinearInterpolate(float64(b), 0, 65535, decode[4], decode[5])) & max16Bit)
	}
	return color.NRGBA64{R: r, G: g, B: b, A: a}, nil
}

func nrgba64Converter(src image.Image) (Image, error) {
	if i, ok := src.(*NRGBA64); ok {
		return i.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 16, 3, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	nrgbaConvertImage(src, g, bounds)
	return g, nil
}

func nrgbaConvertImage(src image.Image, dst Image, bounds image.Rectangle) {
	if masker, ok := src.(SMasker); ok && masker.HasAlpha() {
		// If the source image had some mask, set new size of the mask here.
		dst.(SMasker).MakeAlpha()
	}
	switch imgSrc := src.(type) {
	case Gray:
		nrgbaConvertGrayImage(imgSrc, dst.(NRGBA), bounds)
	case NRGBA:
		nrgbaConvertNRGBAImage(imgSrc, dst.(NRGBA), bounds)
	case CMYK:
		nrgbaConvertCMYKImage(imgSrc, dst.(NRGBA), bounds)
	case RGBA:
		nrgbaConvertRGBAImage(imgSrc, dst.(NRGBA), bounds)
	default:
		convertImages(src, dst, bounds)
	}
}

func nrgbaConvertNRGBAImage(src, dst NRGBA, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			dst.SetNRGBA(x, y, src.NRGBAAt(x, y))
		}
	}
}

func nrgbaConvertGrayImage(src Gray, dst NRGBA, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			g := src.GrayAt(x, y)
			dst.SetNRGBA(x, y, colorsGrayToRGBA(g))
		}
	}
}

func nrgbaConvertCMYKImage(src CMYK, dst NRGBA, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := src.CMYKAt(x, y)
			dst.SetNRGBA(x, y, colorsCMYKtoNRGBA(c))
		}
	}
}

func nrgbaConvertRGBAImage(src RGBA, dst NRGBA, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := src.RGBAAt(x, y)
			dst.SetNRGBA(x, y, colorsRGBAToNRGBA(c))
		}
	}
}
