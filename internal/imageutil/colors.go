package imageutil

import (
	"fmt"
	"image"
	"image/color"
)

// ColorConverter is an enum that defines image color colorConverter.
type ColorConverter interface {
	// Convert converts images to given color colorConverter.
	Convert(src image.Image) (Image, error)
}

func convertImages(src image.Image, dst Image, bounds image.Rectangle) {
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := src.At(x, y)
			dst.Set(x, y, c)
		}
	}
}

// Converter implementations.
var (
	MonochromeConverter = ConverterFunc(monochromeTriangleConverter)
	Gray2Converter      = ConverterFunc(gray2Converter)
	Gray4Converter      = ConverterFunc(gray4Converter)
	GrayConverter       = ConverterFunc(grayConverter)
	Gray16Converter     = ConverterFunc(gray16Converter)
	NRGBA16Converter    = ConverterFunc(nrgba16Converter)
	NRGBAConverter      = ConverterFunc(nrgbaConverter)
	NRGBA64Converter    = ConverterFunc(nrgba64Converter)
	CMYKConverter       = ConverterFunc(cmykConverter)
)

// Models
var (
	Gray2Model   = color.ModelFunc(gray2Model)
	Gray4Model   = color.ModelFunc(gray4Model)
	NRGBA16Model = color.ModelFunc(nrgba16Model)
)

type colorConverter struct {
	f func(src image.Image) (Image, error)
}

// Convert implements color colorConverter interface.
func (p colorConverter) Convert(src image.Image) (Image, error) {
	return p.f(src)
}

// ConverterFunc creates new colorConverter based on the provided 'paletteFunc'.
func ConverterFunc(converterFunc func(img image.Image) (Image, error)) ColorConverter {
	return colorConverter{f: converterFunc}
}

// GetConverter gets the color converter for given bits per component (bit depth) and color components parameters.
func GetConverter(bitsPerComponent, colorComponents int) (ColorConverter, error) {
	switch colorComponents {
	case 1:
		switch bitsPerComponent {
		case 1:
			return MonochromeConverter, nil
		case 2:
			return Gray2Converter, nil
		case 4:
			return Gray4Converter, nil
		case 8:
			return GrayConverter, nil
		case 16:
			return Gray16Converter, nil
		}
	case 3:
		switch bitsPerComponent {
		case 4:
			return NRGBA16Converter, nil
		case 8:
			return NRGBAConverter, nil
		case 16:
			return NRGBA64Converter, nil
		}
	case 4:
		return CMYKConverter, nil
	}
	return nil, fmt.Errorf("provided invalid colorConverter parameters. BitsPerComponent: %d, ColorComponents: %d", bitsPerComponent, colorComponents)
}

// ColorAtFunc is a type of function that gets the
type ColorAtFunc func(x, y, width, bitsPerColor, colorComponents, bytesPerLine int, data []byte, alpha []byte) (color.Color, error)

// ColorAt gets the color of the image with provided parameters.
// 	Parameters:
//		x				- horizontal coordinate of the pixel.
//		y				- vertical coordinate of the pixel.
//		width			- the width of the image.
//		bitsPerColor	- number of bits per color component.
//		colorComponents	- number of color components per pixel.
//		bytesPerLine	- number of bytes per line.
//		data			- byte slice data for given image.
//		alpha			- (optional) the alpha part data slice of the image.
func ColorAt(x, y, width, bitsPerColor, colorComponents, bytesPerLine int, data, alpha []byte, decode []float64) (color.Color, error) {
	switch colorComponents {
	case 1:
		return ColorAtGrayscale(x, y, bitsPerColor, bytesPerLine, data, decode)
	case 3:
		return ColorAtNRGBA(x, y, width, bytesPerLine, bitsPerColor, data, alpha, decode)
	case 4:
		return ColorAtCMYK(x, y, width, data, decode)
	default:
		return nil, fmt.Errorf("provided invalid color component for the image: %d", colorComponents)
	}
}

// ColorAtGrayscale gets the color of the grayscale image with specific 'width', 'bitsPerColor' and 'data'
// at the 'x' and 'y' coordinates position.
func ColorAtGrayscale(x, y, bitsPerColor, bytesPerLine int, data []byte, decode []float64) (color.Color, error) {
	switch bitsPerColor {
	case 1:
		return ColorAtGray1BPC(x, y, bytesPerLine, data, decode)
	case 2:
		return ColorAtGray2BPC(x, y, bytesPerLine, data, decode)
	case 4:
		return ColorAtGray4BPC(x, y, bytesPerLine, data, decode)
	case 8:
		return ColorAtGray8BPC(x, y, bytesPerLine, data, decode)
	case 16:
		return ColorAtGray16BPC(x, y, bytesPerLine, data, decode)
	default:
		return nil, fmt.Errorf("unsupported gray scale bits per color amount: '%d'", bitsPerColor)
	}
}

// ColorAtNRGBA gets the color of the image with specific 'width' and 'bitsPerColor' at 'x' and 'y' coordinates.
// The 'data' defines image data and optional 'alpha' alpha component of the color.
func ColorAtNRGBA(x, y, width, bytesPerLine, bitsPerColor int, data, alpha []byte, decode []float64) (color.Color, error) {
	switch bitsPerColor {
	case 4:
		return ColorAtNRGBA16(x, y, width, bytesPerLine, data, alpha, decode)
	case 8:
		return ColorAtNRGBA32(x, y, width, data, alpha, decode)
	case 16:
		return ColorAtNRGBA64(x, y, width, data, alpha, decode)
	default:
		return nil, fmt.Errorf("unsupported rgb bits per color amount: '%d'", bitsPerColor)
	}
}

func colorsGrayToRGBA(g color.Gray) color.NRGBA {
	return color.NRGBA{
		R: g.Y,
		G: g.Y,
		B: g.Y,
		A: 0xff,
	}
}

func colorsNRGBAToGray(c color.NRGBA) color.Gray {
	r, g, b, _ := c.RGBA()
	y := (19595*r + 38470*g + 7471*b + 1<<15) >> 24
	return color.Gray{Y: uint8(y)}
}

func colorsCMYKtoNRGBA(c color.CMYK) color.NRGBA {
	r, g, b := color.CMYKToRGB(c.C, c.M, c.Y, c.K)
	return color.NRGBA{
		R: r,
		G: g,
		B: b,
		A: 0xff,
	}
}

func colorsNRGBAToCMYK(cr color.NRGBA) color.CMYK {
	r, g, b, _ := cr.RGBA()

	c, m, y, k := color.RGBToCMYK(uint8(r>>8), uint8(g>>8), uint8(b>>8))
	return color.CMYK{
		C: c,
		M: m,
		Y: y,
		K: k,
	}
}

func colorsCMYKToGray(c color.CMYK) color.Gray {
	r, g, b := color.CMYKToRGB(c.C, c.M, c.Y, c.K)
	y := (19595*uint32(r) + 38470*uint32(g) + 7471*uint32(b) + 1<<7) >> 16
	return color.Gray{Y: uint8(y)}
}

func colorsGrayToCMYK(g color.Gray) color.CMYK {
	return color.CMYK{K: 0xff - g.Y}
}

func colorsRGBAToGray(c color.RGBA) color.Gray {
	y := (19595*uint32(c.R) + 38470*uint32(c.G) + 7471*uint32(c.B) + 1<<7) >> 16
	return color.Gray{Y: uint8(y)}
}

func colorsRGBAToNRGBA(c color.RGBA) color.NRGBA {
	switch c.A {
	case 0xff:
		return color.NRGBA{R: c.R, G: c.G, B: c.B, A: 0xff}
	case 0x00:
		return color.NRGBA{}
	default:
		r, g, b, a := c.RGBA()
		r = (r * 0xffff) / a
		g = (g * 0xffff) / a
		b = (b * 0xffff) / a
		return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	}
}

func colorsRGBAToCMYK(cr color.RGBA) color.CMYK {
	c, m, y, k := color.RGBToCMYK(cr.R, cr.G, cr.B)
	return color.CMYK{
		C: c,
		M: m,
		Y: y,
		K: k,
	}
}
