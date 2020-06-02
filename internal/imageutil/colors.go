package imageutil

import (
	"fmt"
	"image/color"
)

// BytesPerLine gets the number of bytes per line for given width, bits per color and color components number.
func BytesPerLine(width, bitsPerComponent, colorComponents int) int {
	return ((width*bitsPerComponent)*colorComponents + 7) >> 3
}

func ColorAt(x, y, width, bitsPerColor, colorComponents, bytesPerLine int, data []byte, alpha []byte) (color.Color, error) {
	switch colorComponents {
	case 1:

	case 3:
		return ColorAtRGB(x, y, width, bitsPerColor, data, alpha)
	case 4:
	default:
	}
}

// ColorAtGrayscale gets the color of the grayscale image with specific 'width', 'bitsPerColor' and 'data'
// at the 'x' and 'y' coordinates position.
func ColorAtGrayscale(x, y, bitsPerColor, bytesPerLine int, data []byte) (color.Color, error) {
	switch bitsPerColor {
	case 1:
		return ColorAtGray1BPC(x, y, bytesPerLine, data)
	case 2:
		return ColorAtGray2BPC(x, y, bytesPerLine, data)
	case 4:
		return ColorAtGray4BPC(x, y, bytesPerLine, data)
	case 8:
	case 16:
	default:
		return nil, fmt.Errorf("unsupported gray scale bits per component amount: '%d'", img.BitsPerComponent)
	}
}

// ColorAtGray1BPC gets the color of image in grayscale color space with one bit per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray1BPC(x, y, bytesPerLine int, data []byte) (color.Gray, error) {
	idx := y*bytesPerLine + x>>3
	if idx >= len(data) {
		return color.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	return color.Gray{Y: data[idx] >> uint(7-(x&7)) & 1 * 255}, nil
}

// ColorAtGray2BPC gets the color of image in grayscale color space with two bits per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray2BPC(x, y, bytesPerLine int, data []byte) (color.Gray, error) {
	idx := y*bytesPerLine + x>>2
	if idx >= len(data) {
		return color.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	return color.Gray{Y: data[idx] >> uint(6-(x&3)*2) & 3 * 85}, nil
}

// ColorAtGray4BPC gets the color of image in grayscale color space with four bits per component specific 'width',
// 'bytesPerLine' and 'data' at the 'x' and 'y' coordinates position.
func ColorAtGray4BPC(x, y, bytesPerLine int, data []byte) (color.Gray, error) {
	idx := y*bytesPerLine + x>>1
	if idx >= len(data) {
		return color.Gray{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	return color.Gray{Y: (data[idx] >> uint(4-(x&1)*4) & 15 * 17) & 0xff}, nil
}

// ColorAtRGB gets the color of the image with specific 'width' and 'bitsPerColor' at 'x' and 'y' coordinates.
// The 'data' defines image data and optional 'alpha' alpha component of the color.
func ColorAtRGB(x, y, width, bitsPerColor int, data, alpha []byte) (color.Color, error) {
	switch bitsPerColor {
	case 8:
		return ColorAtRGB8BPC(x, y, width, data, alpha)
	case 16:
		return ColorAtRGB16BPC(x, y, width, data, alpha)
	default:
	}
}

// ColorAtRGB8BPC gets the 8 bit per component RGBA color at 'x', 'y'.
func ColorAtRGB8BPC(x, y, width int, data, alpha []byte) (color.RGBA, error) {
	idx := y*width + x

	i := 3 * idx
	if i+2 >= len(data) {
		return color.RGBA{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	a := uint8(0xff)
	if alpha != nil && len(alpha) > idx {
		a = alpha[idx]
	}

	return color.RGBA{
		R: data[i],
		G: data[i+1],
		B: data[i+2],
		A: a,
	}, nil
}

// ColorAtRGB16BPC gets 16 bit per component RGBA64 color.
func ColorAtRGB16BPC(x, y, width int, data, alpha []byte) (color.RGBA64, error) {
	idx := (y*width + x) * 2

	i := idx * 3
	if i+5 >= len(data) {
		return color.RGBA64{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}

	a := uint16(0xffff)
	if alpha != nil && len(alpha) > idx+1 {
		a = uint16(alpha[idx])<<8 | uint16(alpha[idx+1])
	}

	return color.RGBA64{
		R: uint16(data[i])<<8 | uint16(data[i+1]),
		G: uint16(data[i+2])<<8 | uint16(data[i+3]),
		B: uint16(data[i+4])<<8 | uint16(data[i+5]),
		A: a,
	}, nil
}

// ColorAtCMYK gets the color at CMYK colorspace at 'x' and 'y' position for an image with specific 'width'.
func ColorAtCMYK(x, y, width int, data []byte) (color.CMYK, error) {
	idx := 4 * (y*width + x)
	if idx+3 >= len(data) {
		return color.CMYK{}, fmt.Errorf("image coordinates out of range (%d, %d)", x, y)
	}
	return color.CMYK{
		C: data[idx] & 0xff,
		M: data[idx+1] & 0xff,
		Y: data[idx+2] & 0xff,
		K: data[idx+3] & 0xff,
	}, nil
}
