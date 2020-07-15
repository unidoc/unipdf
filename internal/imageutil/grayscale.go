/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package imageutil

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

func init() {
	makePixelSumAtByte()
}

// Gray is an interface that allows to get and set color.Gray for grayscale images.
type Gray interface {
	GrayAt(x, y int) color.Gray
	SetGray(x, y int, g color.Gray)
}

// MonochromeThresholdConverter creates a new monochrome colorConverter.
func MonochromeThresholdConverter(threshold uint8) ColorConverter {
	return &monochromeThresholdConverter{Threshold: threshold}
}

type monochromeThresholdConverter struct {
	Threshold uint8
}

func monochromeTriangleConverter(src image.Image) (Image, error) {
	if m, ok := src.(*Monochrome); ok {
		return m, nil
	}

	graySrc, ok := src.(Gray)
	if !ok {
		// Convert an image to the basic gray image and get it's histogramer
		g, err := GrayConverter.Convert(src)
		if err != nil {
			return nil, err
		}
		graySrc = g.(Gray)
	}
	bounds := src.Bounds()
	m, err := NewImage(bounds.Max.X, bounds.Max.Y, 1, 1, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	dst := m.(*Monochrome)
	threshold := AutoThresholdTriangle(GrayHistogram(graySrc))
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			gc := monochromeModelFunc(graySrc.GrayAt(x, y), monochromeModel(threshold))
			dst.SetGray(x, y, gc)
		}
	}
	return m, nil
}

// Convert implements Palette interface.
func (m *monochromeThresholdConverter) Convert(img image.Image) (Image, error) {
	if mi, ok := img.(*Monochrome); ok {
		return mi.Copy(), nil
	}
	bounds := img.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 1, 1, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	g.(*Monochrome).ModelThreshold = m.Threshold

	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			g.Set(x, y, c)
		}
	}
	return g, nil
}

// Monochrome is a grayscale image with a 1-bit per pixel.
type Monochrome struct {
	ImageBase
	ModelThreshold uint8
}

// Base implements Image interface.
func (m *Monochrome) Base() *ImageBase {
	return &m.ImageBase
}

// Set implements draw.Image interface.
func (m *Monochrome) Set(x, y int, c color.Color) {
	index := y*m.BytesPerLine + x>>3
	if index > len(m.Data)-1 {
		return
	}
	g := m.ColorModel().Convert(c).(color.Gray)
	m.setGray(x, g, index)
}

// Copy implements Image interface.
func (m *Monochrome) Copy() Image {
	return &Monochrome{ImageBase: m.ImageBase.copy(), ModelThreshold: m.ModelThreshold}
}

//
// Monochrome image.Image.
//

// Compile time check if Monochrome implements image.Image interface.
var _ image.Image = &Monochrome{}

// At implements image.Image interface.
func (m *Monochrome) At(x, y int) color.Color {
	c, _ := m.ColorAt(x, y)
	return c
}

// Bounds implements image.Image.
func (m *Monochrome) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: m.Width, Y: m.Height}}
}

type monochromeModel uint8

// Convert implements color.Model interface.
func (m monochromeModel) Convert(c color.Color) color.Color {
	g := color.GrayModel.Convert(c).(color.Gray)
	return monochromeModelFunc(g, m)
}

func monochromeModelFunc(g color.Gray, m monochromeModel) color.Gray {
	if g.Y > uint8(m) {
		return color.Gray{Y: math.MaxUint8}
	}
	return color.Gray{}
}

// MonochromeModel is the monochrome color Model that relates on given threshold.
func MonochromeModel(threshold uint8) color.Model {
	return monochromeModel(threshold)
}

// ColorModel implements image.Image interface.
func (m *Monochrome) ColorModel() color.Model {
	return MonochromeModel(m.ModelThreshold)
}

//
// Monochrome Image.
//

// Compile time check if Monochrome implements Image interface.
var _ Image = &Monochrome{}

// ColorAt implements Image interface.
func (m *Monochrome) ColorAt(x, y int) (color.Color, error) {
	return ColorAtGray1BPC(x, y, m.BytesPerLine, m.Data, m.Decode)
}

// Validate implements Image interface.
func (m *Monochrome) Validate() error {
	if len(m.Data) != m.Height*m.BytesPerLine {
		return ErrInvalidImage
	}
	return nil
}

//
// Gray interface.
//
// Compile time check if Monochrome implements Gray interface.
var _ Gray = &Monochrome{}

// Histogram implements Histogramer interface.
func (m *Monochrome) Histogram() (histogram [256]int) {
	for _, bt := range m.Data {
		histogram[0xff] += int(tab8[m.Data[bt]])
	}
	return histogram
}

// GrayAt implements Gray interface.
func (m *Monochrome) GrayAt(x, y int) color.Gray {
	c, _ := ColorAtGray1BPC(x, y, m.BytesPerLine, m.Data, m.Decode)
	return c
}

// SetGray implements Gray interface.
func (m *Monochrome) SetGray(x, y int, g color.Gray) {
	index := y*m.BytesPerLine + x>>3
	if index > len(m.Data)-1 {
		return
	}
	g = monochromeModelFunc(g, monochromeModel(m.ModelThreshold))
	m.setGray(x, g, index)
}

func (m *Monochrome) setGray(x int, g color.Gray, index int) {
	if g.Y == 0 {
		m.clearBit(index, x)
	} else {
		m.setBit(index, x)
	}
}

func (m *Monochrome) setBit(index, x int) {
	m.Data[index] |= 0x80 >> uint(x&7)
}

func (m *Monochrome) clearBit(index, x int) {
	m.Data[index] &= ^(0x80 >> uint(x&7))
}

// ColorAtGray1BPC gets the color of image in grayscale color space with one bit per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray1BPC(x, y, bytesPerLine int, data []byte, decode []float64) (color.Gray, error) {
	idx := y*bytesPerLine + x>>3
	if idx >= len(data) {
		return color.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	// Get the 0,1 value from the byte at 'idx' position.
	c := data[idx] >> uint(7-(x&7)) & 1
	if len(decode) == 2 {
		// Use the decode matrix for the values interpolation.
		c = uint8(LinearInterpolate(float64(c), 0.0, 1.0, decode[0], decode[1])) & 1
	}
	return color.Gray{Y: c * 255}, nil
}

// Gray2 is a 2-bit base grayscale image. It implements image.Image, draw.Image, Image and Gray interfaces.
type Gray2 struct {
	ImageBase
}

//
// Gray2 Image methods.
//

// Compile time check if Gray2 implements Image interface.
var _ Image = &Gray2{}

// Base implements Image interface.
func (i *Gray2) Base() *ImageBase {
	return &i.ImageBase
}

// ColorAt implements Image interface.
func (i *Gray2) ColorAt(x, y int) (color.Color, error) {
	return ColorAtGray2BPC(x, y, i.BytesPerLine, i.Data, i.Decode)
}

// Copy implements Image interface.
func (i *Gray2) Copy() Image {
	return &Gray2{ImageBase: i.copy()}
}

// Set implements draw.Image interface.
func (i *Gray2) Set(x, y int, c color.Color) {
	if x >= i.Width || y >= i.Height {
		return
	}
	g := Gray2Model.Convert(c).(color.Gray)
	line := y * i.BytesPerLine
	index := line + (x >> 2)
	value := g.Y >> 6
	i.Data[index] = (i.Data[index] & (^(0xc0 >> uint(2*((x)&3))))) | (value << uint(6-2*(x&3)))
}

// Validate implements Image interface.
func (i *Gray2) Validate() error {
	if len(i.Data) != i.Height*i.BytesPerLine {
		return ErrInvalidImage
	}
	return nil
}

//
// Gray2 image.Image methods.
//

// Compile time check if Gray2 implements image.Image interface.
var _ image.Image = &Gray2{}

// At implements image.Image interface.
func (i *Gray2) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

// Bounds implements
func (i *Gray2) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

// ColorModel implements image.Image interface.
func (i *Gray2) ColorModel() color.Model {
	return Gray2Model
}

func gray2Model(c color.Color) color.Color {
	g := color.GrayModel.Convert(c).(color.Gray)
	return gray2ModelGray(g)
}

func gray2ModelGray(g color.Gray) color.Gray {
	tmp := g.Y >> 6
	tmp |= tmp << 2
	g.Y = tmp | tmp<<4
	return g
}

//
// Gray2 - Gray interface.
//

// Compile time check if Gray2 implements Gray interface.
var _ Gray = &Gray2{}

// GrayAt implements Gray interface.
func (i *Gray2) GrayAt(x, y int) color.Gray {
	c, _ := ColorAtGray2BPC(x, y, i.BytesPerLine, i.Data, i.Decode)
	return c
}

// SetGray implements Gray interface.
func (i *Gray2) SetGray(x, y int, gray color.Gray) {
	g := gray2ModelGray(gray)
	line := y * i.BytesPerLine
	index := line + (x >> 2)
	if index >= len(i.Data) {
		return
	}
	value := g.Y >> 6
	i.Data[index] = (i.Data[index] & (^(0xc0 >> uint(2*((x)&3))))) | (value << uint(6-2*(x&3)))
}

// Histogram implements Histogramer interface.
func (i *Gray2) Histogram() (histogram [256]int) {
	for x := 0; x < i.Width; x++ {
		for y := 0; y < i.Height; y++ {
			histogram[i.GrayAt(x, y).Y]++
		}
	}
	return histogram
}

// ColorAtGray2BPC gets the color of image in grayscale color space with two bits per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray2BPC(x, y, bytesPerLine int, data []byte, decode []float64) (color.Gray, error) {
	idx := y*bytesPerLine + x>>2
	if idx >= len(data) {
		return color.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	c := data[idx] >> uint(6-(x&3)*2) & 3
	if len(decode) == 2 {
		c = uint8(uint32(LinearInterpolate(float64(c), 0, 3.0, decode[0], decode[1])) & 3)
	}
	return color.Gray{Y: c * 85}, nil
}

func gray2Converter(src image.Image) (Image, error) {
	if g, ok := src.(*Gray2); ok {
		return g.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 2, 1, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	grayConvertImage(src, g, bounds)
	return g, nil
}

// Gray4 is a grayscale image implementation where the gray pixel is stored in 4 bits.
type Gray4 struct {
	ImageBase
}

//
// Image interface.
//

// Compile time check if Gray4 implements Image interface.
var _ Image = &Gray4{}

// Base implements Image interface.
func (i *Gray4) Base() *ImageBase {
	return &i.ImageBase
}

// ColorAt implements Image interface.
func (i *Gray4) ColorAt(x, y int) (color.Color, error) {
	return ColorAtGray4BPC(x, y, i.BytesPerLine, i.Data, i.Decode)
}

// Copy creates a copy of given image.
func (i *Gray4) Copy() Image {
	return &Gray4{ImageBase: i.copy()}
}

// Set for given image sets color 'c' at coordinates 'x' and 'y'.
func (i *Gray4) Set(x, y int, c color.Color) {
	if x >= i.Width || y >= i.Height {
		return
	}
	g := Gray4Model.Convert(c).(color.Gray)
	i.setGray(x, y, g)
}

// Validate implements Image interface.
func (i *Gray4) Validate() error {
	if len(i.Data) != i.Height*i.BytesPerLine {
		return ErrInvalidImage
	}
	return nil
}

//
// Gray4 image.Image interface methods.
//

// Compile time check if Gray4 implements image.Image interface.
var _ image.Image = &Gray4{}

// At implements image.Image interface.
func (i *Gray4) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

// Bounds implements image.Image interface.
func (i *Gray4) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

// ColorModel implements image.Image interface.
func (i *Gray4) ColorModel() color.Model {
	return Gray4Model
}

//
// Gray4 Gray interface methods.
//

// Compile time check if Gray4 implements Gray interface.
var _ Gray = &Gray4{}

// GrayAt implements Gray interface.
func (i *Gray4) GrayAt(x, y int) color.Gray {
	c, _ := ColorAtGray4BPC(x, y, i.BytesPerLine, i.Data, i.Decode)
	return c
}

// Histogram implements Histogramer interface.
func (i *Gray4) Histogram() (histogram [256]int) {
	for x := 0; x < i.Width; x++ {
		for y := 0; y < i.Height; y++ {
			histogram[i.GrayAt(x, y).Y]++
		}
	}
	return histogram
}

// SetGray implements Gray interface.
func (i *Gray4) SetGray(x, y int, g color.Gray) {
	if x >= i.Width || y >= i.Height {
		return
	}
	g = gray4ModelFunc(g)
	i.setGray(x, y, g)
}

func (i *Gray4) setGray(x int, y int, g color.Gray) {
	line := y * i.BytesPerLine
	index := line + (x >> 1)
	if index >= len(i.Data) {
		return
	}
	value := g.Y >> 4
	i.Data[index] = (i.Data[index] & (^(0xf0 >> uint(4*(x&1))))) | (value << uint(4-4*(x&1)))
}

// ColorAtGray4BPC gets the color of image in grayscale color space with 4 bits per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray4BPC(x, y, bytesPerLine int, data []byte, decode []float64) (color.Gray, error) {
	idx := y*bytesPerLine + x>>1
	if idx >= len(data) {
		return color.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	c := data[idx] >> uint(4-(x&1)*4) & 0xf
	if len(decode) == 2 {
		c = uint8(uint32(LinearInterpolate(float64(c), 0, 15, decode[0], decode[1])) & 0xf)
	}
	return color.Gray{Y: c * 17 & 0xff}, nil
}

func gray4Model(c color.Color) color.Color {
	g := color.GrayModel.Convert(c).(color.Gray)
	return gray4ModelFunc(g)
}

func gray4ModelFunc(g color.Gray) color.Gray {
	g.Y >>= 4
	g.Y |= g.Y << 4
	return g
}

func gray4Converter(src image.Image) (Image, error) {
	if g, ok := src.(*Gray4); ok {
		return g.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 4, 1, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	grayConvertImage(src, g, bounds)
	return g, nil
}

// Gray8 is a grayscale image with the gray pixel size of 8-bit.
type Gray8 struct {
	ImageBase
}

//
// Gray8 Image interface methods.
//

// Compile time check if Gray8 implements Image interface.
var _ Image = &Gray8{}

// Base implements image interface.
func (i *Gray8) Base() *ImageBase {
	return &i.ImageBase
}

// ColorAt implements Image interface.
func (i *Gray8) ColorAt(x, y int) (color.Color, error) {
	return ColorAtGray8BPC(x, y, i.BytesPerLine, i.Data, i.Decode)
}

// Copy implements Image interface.
func (i *Gray8) Copy() Image {
	return &Gray8{ImageBase: i.copy()}
}

// Set implements draw.Image interface.
func (i *Gray8) Set(x, y int, c color.Color) {
	idx := y*i.BytesPerLine + x
	if idx > len(i.Data)-1 {
		return
	}
	g := color.GrayModel.Convert(c)
	i.Data[idx] = g.(color.Gray).Y
}

// Validate implements Image interface.
func (i *Gray8) Validate() error {
	if len(i.Data) != i.Height*i.BytesPerLine {
		return ErrInvalidImage
	}
	return nil
}

//
// Gray8 image.Image methods.
//

var _ image.Image = &Gray8{}

// At implements image.Image interface.
func (i *Gray8) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

// Bounds implements image.Image interface.
func (i *Gray8) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

// ColorModel implements image.Image interface.`
func (i *Gray8) ColorModel() color.Model {
	return color.GrayModel
}

//
// Gray8 Gray interface methods.
//

// Compile time check if Gray8 implements Gray interface.
var _ Gray = &Gray8{}

// Histogram implements Histogramer interface.
func (i *Gray8) Histogram() (histogram [256]int) {
	for j := 0; j < len(i.Data); j++ {
		histogram[i.Data[j]]++
	}
	return histogram
}

// GrayAt implements Gray interface.
func (i *Gray8) GrayAt(x, y int) color.Gray {
	c, _ := ColorAtGray8BPC(x, y, i.BytesPerLine, i.Data, i.Decode)
	return c
}

// SetGray implements Gray interface.
func (i *Gray8) SetGray(x, y int, g color.Gray) {
	idx := y*i.BytesPerLine + x
	if idx > len(i.Data)-1 {
		return
	}
	i.Data[idx] = g.Y
}

// ColorAtGray8BPC gets the color of image in grayscale color space with 8 bits per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray8BPC(x, y, bytesPerLine int, data []byte, decode []float64) (color.Gray, error) {
	idx := y*bytesPerLine + x
	if idx >= len(data) {
		return color.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	c := data[idx]
	if len(decode) == 2 {
		c = uint8(uint32(LinearInterpolate(float64(c), 0, 255, decode[0], decode[1])) & 0xff)
	}
	return color.Gray{Y: c}, nil
}

func grayConverter(src image.Image) (Image, error) {
	if g, ok := src.(*Gray8); ok {
		return g.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 8, 1, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	grayConvertImage(src, g, bounds)
	return g, nil
}

// Gray16 is a grayscale image where gray pixel is of 16-bit size.
type Gray16 struct {
	ImageBase
}

//
// Gray16 Image interface methods.
//

// Compile time check if Gray16 implements Image interface.
var _ Image = &Gray16{}

// Base implements Image interface.
func (i *Gray16) Base() *ImageBase {
	return &i.ImageBase
}

// ColorAt implements Image interface.
func (i *Gray16) ColorAt(x, y int) (color.Color, error) {
	return ColorAtGray16BPC(x, y, i.BytesPerLine, i.Data, i.Decode)
}

// Copy implements Image interface.
func (i *Gray16) Copy() Image {
	return &Gray16{ImageBase: i.copy()}
}

// Set implements draw.Image interface.
func (i *Gray16) Set(x, y int, c color.Color) {
	idx := (y*i.BytesPerLine/2 + x) * 2
	if idx+1 >= len(i.Data) {
		return
	}
	g := color.Gray16Model.Convert(c).(color.Gray16)

	i.Data[idx], i.Data[idx+1] = uint8(g.Y>>8), uint8(g.Y&0xff)
}

// Validate implements Image interface.
func (i *Gray16) Validate() error {
	if len(i.Data) != i.Height*i.BytesPerLine {
		return ErrInvalidImage
	}
	return nil
}

//
// Gray16 image.Image interface methods.
//

// Compile time check if Gray16 implements image.Image interface.
var _ image.Image = &Gray16{}

// At implements image.Image interface.
func (i *Gray16) At(x, y int) color.Color {
	c, _ := i.ColorAt(x, y)
	return c
}

// Bounds implements image.Image interface.
func (i *Gray16) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.Width, Y: i.Height}}
}

// ColorModel implements image.Image interface.
func (i *Gray16) ColorModel() color.Model {
	return color.Gray16Model
}

//
// Gray16 Gray interface methods.
//
var _ Gray = &Gray16{}

// GrayAt implements Gray interface.
func (i *Gray16) GrayAt(x, y int) color.Gray {
	c, _ := i.ColorAt(x, y)
	return color.Gray{Y: uint8(c.(color.Gray16).Y >> 8)}
}

// Histogram implements Histogramer interface.
func (i *Gray16) Histogram() (histogram [256]int) {
	for x := 0; x < i.Width; x++ {
		for y := 0; y < i.Height; y++ {
			histogram[i.GrayAt(x, y).Y]++
		}
	}
	return histogram
}

// SetGray implements Gray interface.
func (i *Gray16) SetGray(x, y int, g color.Gray) {
	idx := (y*i.BytesPerLine/2 + x) * 2
	if idx+1 >= len(i.Data) {
		return
	}
	i.Data[idx] = g.Y
	i.Data[idx+1] = g.Y
}

// ColorAtGray16BPC gets the color of image in grayscale color space with 8 bits per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray16BPC(x, y, bytesPerLine int, data []byte, decode []float64) (color.Gray16, error) {
	idx := (y*bytesPerLine/2 + x) * 2
	// For 16 BPC we need two bytes, thus idx + 1 must be greater equal than the length of data.
	if idx+1 >= len(data) {
		return color.Gray16{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	c := uint16(data[idx])<<8 | uint16(data[idx+1])
	if len(decode) == 2 {
		c = uint16(uint64(LinearInterpolate(float64(c), 0, 65535, decode[0], decode[1])))
	}
	return color.Gray16{Y: c}, nil
}

func gray16Converter(src image.Image) (Image, error) {
	if g, ok := src.(*Gray16); ok {
		return g.Copy(), nil
	}
	bounds := src.Bounds()
	g, err := NewImage(bounds.Max.X, bounds.Max.Y, 16, 1, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	grayConvertImage(src, g, bounds)
	return g, nil
}

func grayConvertImage(src image.Image, dst Image, bounds image.Rectangle) {
	switch srcImg := src.(type) {
	case Gray:
		grayConvertFromGray(srcImg, dst.(Gray), bounds)
	case NRGBA:
		grayConvertFromNRGBA(srcImg, dst.(Gray), bounds)
	case CMYK:
		grayConvertFromCMYK(srcImg, dst.(Gray), bounds)
	case RGBA:
		grayConvertFromRGBA(srcImg, dst.(Gray), bounds)
	default:
		convertImages(src, dst.(Image), bounds)
	}
}

func grayConvertFromGray(src, dst Gray, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			dst.SetGray(x, y, src.GrayAt(x, y))
		}
	}
}

func grayConvertFromNRGBA(src NRGBA, dst Gray, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			g := colorsNRGBAToGray(src.NRGBAAt(x, y))
			dst.SetGray(x, y, g)
		}
	}
}

func grayConvertFromRGBA(src RGBA, dst Gray, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			g := colorsRGBAToGray(src.RGBAAt(x, y))
			dst.SetGray(x, y, g)
		}
	}
}

func grayConvertFromCMYK(src CMYK, dst Gray, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			g := colorsCMYKToGray(src.CMYKAt(x, y))
			dst.SetGray(x, y, g)
		}
	}
}

// tab8 contains number of '1' bits in each possible 8 bit value stored at it's index.
var tab8 [256]uint8

func makePixelSumAtByte() {
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
