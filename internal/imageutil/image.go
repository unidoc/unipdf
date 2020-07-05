package imageutil

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
)

// ErrInvalidImage is an error used when provided image is invalid.
var ErrInvalidImage = errors.New("invalid image data size for provided dimensions")

// Image is an interface used that allows to do image operations.
type Image interface {
	draw.Image
	Base() *ImageBase
	Copy() Image
	Pix() []byte
	ColorAt(x, y int) (color.Color, error)
	Validate() error
}

// SMasker is an interface used to get and check the alpha mask from an image.
type SMasker interface {
	HasAlpha() bool
	GetAlpha() []byte
	MakeAlpha()
}

// ImageBase is a structure that represents an bitmap image.
type ImageBase struct {
	Width, Height                     int
	BitsPerComponent, ColorComponents int
	Data, Alpha                       []byte
	Decode                            []float64
	BytesPerLine                      int
}

// Pix implements Image interface.
func (i *ImageBase) Pix() []byte {
	return i.Data
}

// HasAlpha implements SMasker interface.
func (i *ImageBase) HasAlpha() bool {
	return i.Alpha == nil
}

// GetAlpha implements SMasker interface.
func (i *ImageBase) GetAlpha() []byte {
	return i.Alpha
}

// MakeAlpha implements SMasker interface.
func (i *ImageBase) MakeAlpha() {
	i.newAlpha()
}

func (i *ImageBase) newAlpha() {
	// The alpha component would always have new 1 color component while calculating bytes per line.
	bytesPerLine := BytesPerLine(i.Width, i.BitsPerComponent, 1)
	i.Alpha = make([]byte, i.Height*bytesPerLine)
}

func (i *ImageBase) copy() ImageBase {
	cp := *i
	cp.Data = make([]byte, len(i.Data))
	copy(cp.Data, i.Data)
	return cp
}

// NewImage creates new image for provided image parameters and image data byte slice.
func NewImage(width, height, bitsPerComponent, colorComponents int, data, alpha []byte, decode []float64) (Image, error) {
	base := ImageBase{
		Width:            width,
		Height:           height,
		BitsPerComponent: bitsPerComponent,
		ColorComponents:  colorComponents,
		Data:             data,
		Alpha:            alpha,
		Decode:           decode,
		BytesPerLine:     BytesPerLine(width, bitsPerComponent, colorComponents),
	}
	if data == nil {
		base.Data = make([]byte, height*base.BytesPerLine)
	}

	var img Image
	switch colorComponents {
	case 1:
		switch bitsPerComponent {
		case 1:
			img = &Monochrome{ImageBase: base, ModelThreshold: 0x0f}
		case 2:
			img = &Gray2{ImageBase: base}
		case 4:
			img = &Gray4{ImageBase: base}
		case 8:
			img = &Gray8{ImageBase: base}
		case 16:
			img = &Gray16{ImageBase: base}
		}
	case 3:
		switch bitsPerComponent {
		case 4:
			img = &NRGBA16{ImageBase: base}
		case 8:
			img = &NRGBA32{ImageBase: base}
		case 16:
			img = &NRGBA64{ImageBase: base}
		}
	case 4:
		img = &CMYK32{ImageBase: base}
	}
	if img == nil {
		return nil, ErrInvalidImage
	}
	// if err := img.Validate(); err != nil {
	// 	return nil, err
	// }
	return img, nil
}

// BytesPerLine gets the number of bytes per line for given width, bits per color and color components number.
func BytesPerLine(width, bitsPerComponent, colorComponents int) int {
	return ((width*bitsPerComponent)*colorComponents + 7) >> 3
}

// FromGoImage creates a new Image from provided image 'i'.
func FromGoImage(i image.Image) (Image, error) {
	switch img := i.(type) {
	case Image:
		return img.Copy(), nil
	case Gray:
		return GrayConverter.Convert(i)
	case CMYK:
		return CMYKConverter.Convert(i)
	default:
		return NRGBAConverter.Convert(i)
	}
}
