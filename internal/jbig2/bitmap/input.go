package bitmap

import (
	goimage "image"
	gocolor "image/color"
	"image/draw"
	"math"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// CreateInput is the input for of the Bitmap.
type CreateInput struct {
	Width, Height       int
	BitsPerComponent    int
	ColorComponents     int
	Data                []byte
	BlackWhiteThreshold float64

	// Transparency data: alpha channel.
	// Stored in same bits per component as original data with 1 color component.
	alphaData []byte    // Alpha channel data.
	hasAlpha  bool      // Indicates whether the alpha channel data is available.
	decode    []float64 // [Dmin Dmax ... values for each color component]
}

// ColorAt get the color of the image at.
func (c *CreateInput) ColorAt(x, y int) (gocolor.Color, error) {
	const processName = "bitmap.CreateInput.ColorAt"
	data := c.Data
	lenData := len(c.Data)
	maxVal := uint32(1<<uint32(c.BitsPerComponent)) - 1

	switch c.ColorComponents {
	case 1:
		// Grayscale image.
		switch c.BitsPerComponent {
		case 1, 2, 4:
			// 1, 2 or 4 bit grayscale image.
			bpc := c.BitsPerComponent
			divider := 8 / bpc

			// Calculate index of byte containing the gray value
			// in the image data, based on the specified x,y coordinates.
			idx := (y*c.Width + x) / divider
			if idx >= lenData {
				return nil, errors.Errorf(processName, "image coordinates out of range (%d, %d)", x, y)
			}

			// Calculate bit position at which the color data starts.
			pos := 8 - uint(((y*c.Width+x)%divider)*bpc+bpc)

			// Extract gray color value starting at the calculated position.
			val := float64(((1 << uint(c.BitsPerComponent)) - 1) & (data[idx] >> pos))
			if len(c.decode) == 2 {
				dMin, dMax := c.decode[0], c.decode[1]
				val = interpolate(val, 0, float64(maxVal), dMin, dMax)
			}

			return gocolor.Gray{
				Y: uint8(uint32(val) * 255 / maxVal & 0xff),
			}, nil
		case 16:
			// 16 bit grayscale image.
			idx := (y*c.Width + x) * 2
			if idx+1 >= lenData {
				return nil, errors.Errorf(processName, "image coordinates out of range (%d, %d)", x, y)
			}

			return gocolor.Gray16{
				Y: uint16(data[idx])<<8 | uint16(data[idx+1]),
			}, nil
		default:
			// Assuming 8 bit grayscale image.
			idx := y*c.Width + x
			if idx >= lenData {
				return nil, errors.Errorf(processName, "image coordinates out of range (%d, %d)", x, y)
			}
			val := float64(data[idx])
			if len(c.decode) == 2 {
				dMin, dMax := c.decode[0], c.decode[1]
				val = interpolate(val, 0, float64(maxVal), dMin, dMax)
			}

			return gocolor.Gray{
				Y: uint8(uint32(val) * 255 / maxVal & 0xff),
			}, nil
		}
	case 3:
		// RGB image.
		switch c.BitsPerComponent {
		case 4:
			// 4 bit per component RGB image.
			idx := (y*c.Width + x) * 3 / 2
			if idx+1 >= lenData {
				return nil, errors.Errorf(processName, "image coordinates out of range (%d, %d)", x, y)
			}

			// Calculate bit position at which the color data starts.
			pos := (y*c.Width + x) * 3 % 2

			var r, g, b uint8
			if pos == 0 {
				// The R and G components are contained by the current byte
				// and the B component is contained by the next byte.
				r = ((1 << uint(c.BitsPerComponent)) - 1) & (data[idx] >> uint(4))
				g = ((1 << uint(c.BitsPerComponent)) - 1) & (data[idx] >> uint(0))
				b = ((1 << uint(c.BitsPerComponent)) - 1) & (data[idx+1] >> uint(4))
			} else {
				// The R component is contained by the current byte and the
				// G and B components are contained by the next byte.
				r = ((1 << uint(c.BitsPerComponent)) - 1) & (data[idx] >> uint(0))
				g = ((1 << uint(c.BitsPerComponent)) - 1) & (data[idx+1] >> uint(4))
				b = ((1 << uint(c.BitsPerComponent)) - 1) & (data[idx+1] >> uint(0))
			}

			return gocolor.RGBA{
				R: uint8(uint32(r) * 255 / maxVal & 0xff),
				G: uint8(uint32(g) * 255 / maxVal & 0xff),
				B: uint8(uint32(b) * 255 / maxVal & 0xff),
				A: uint8(0xff),
			}, nil
		case 16:
			// 16 bit per component RGB image.
			idx := (y*c.Width + x) * 2

			i := idx * 3
			if i+5 >= lenData {
				return nil, errors.Errorf(processName, "image coordinates out of range (%d, %d)", x, y)
			}

			a := uint16(0xffff)
			if c.alphaData != nil && len(c.alphaData) > idx+1 {
				a = uint16(c.alphaData[idx])<<8 | uint16(c.alphaData[idx+1])
			}

			return gocolor.RGBA64{
				R: uint16(data[i])<<8 | uint16(data[i+1]),
				G: uint16(data[i+2])<<8 | uint16(data[i+3]),
				B: uint16(data[i+4])<<8 | uint16(data[i+5]),
				A: a,
			}, nil
		default:
			// Assuming 8 bit per component RGB image.
			idx := y*c.Width + x

			i := 3 * idx
			if i+2 >= lenData {
				return nil, errors.Errorf(processName, "image coordinates out of range (%d, %d)", x, y)
			}

			a := uint8(0xff)
			if c.alphaData != nil && len(c.alphaData) > idx {
				a = c.alphaData[idx]
			}

			return gocolor.RGBA{
				R: data[i] & 0xff,
				G: data[i+1] & 0xff,
				B: data[i+2] & 0xff,
				A: a,
			}, nil
		}
	case 4:
		// CMYK image.
		idx := 4 * (y*c.Width + x)
		if idx+3 >= lenData {
			return nil, errors.Errorf(processName, "image coordinates out of range (%d, %d)", x, y)
		}

		return gocolor.CMYK{
			C: data[idx] & 0xff,
			M: data[idx+1] & 0xff,
			Y: data[idx+2] & 0xff,
			K: data[idx+3] & 0xff,
		}, nil
	}
	common.Log.Debug("ERROR: unsupported image. %d components, %d bits per component", c.ColorComponents, c.BitsPerComponent)
	return nil, errors.Error(processName, "unsupported image colorspace")
}

// ToGoImage gets go image from the given input.
func (c *CreateInput) ToGoImage() (goimage.Image, error) {
	const processName = "bitmap.CreateInput.ToGoImage"
	common.Log.Trace("Converting to go image")
	bounds := goimage.Rect(0, 0, c.Width, c.Height)

	var imgOut draw.Image
	switch c.ColorComponents {
	case 1:
		if c.BitsPerComponent == 16 {
			imgOut = goimage.NewGray16(bounds)
		} else {
			imgOut = goimage.NewGray(bounds)
		}
	case 3:
		if c.BitsPerComponent == 16 {
			imgOut = goimage.NewRGBA64(bounds)
		} else {
			imgOut = goimage.NewRGBA(bounds)
		}
	case 4:
		imgOut = goimage.NewCMYK(bounds)
	default:
		common.Log.Debug("Unsupported number of colors components per sample: %d", c.ColorComponents)
		return nil, errors.Error(processName, "unsupported colors")
	}

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			color, err := c.ColorAt(x, y)
			if err != nil {
				common.Log.Debug("ERROR: %v. Image details: %d components, %d bits per component, %dx%d dimensions, %d data length",
					err, c.ColorComponents, c.BitsPerComponent, c.Width, c.Height, len(c.Data))
				continue
			}
			imgOut.Set(x, y, color)
		}
	}
	return imgOut, nil
}

// NewWithInput creates new bitmap for provided 'i' CreateInput.
func NewWithInput(input *CreateInput) (*Bitmap, error) {
	const processName = "bitmap.NewWithInput"
	i, err := input.ToGoImage()
	if err != nil {
		return nil, err
	}
	var th uint8
	if input.BlackWhiteThreshold <= 0 {
		gray := ImgToGray(i)
		histogram := GrayImageHistogram(gray)
		th = AutoThresholdTriangle(histogram)
		i = gray
	} else if input.BlackWhiteThreshold > 1.0 {
		// check if input.BlackWhiteThreshold is unknown - set to 0.0 is not in the allowed range.
		return nil, errors.Error(processName, "provided threshold is not in a range {0.0, 1.0}")
	} else {
		th = uint8(255 * input.BlackWhiteThreshold)
	}
	return binaryToBitmap(ImgToBinary(i, th)), nil
}

func binaryToBitmap(i *goimage.Gray) *Bitmap {
	bounds := i.Bounds()
	// compute the rowStride - number of bytes in the row.
	bm := New(bounds.Dx(), bounds.Dy())
	// allocate the byte slice data
	var pix gocolor.Gray
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			pix = i.GrayAt(x, y)
			// check if the pixel is black or white
			// where black pixel would be stored as '1' bit
			// and the white as '0' bit.

			// the pix is color.Black if it's Y value is '0'.
			if pix.Y != 0 {
				if err := bm.SetPixel(x, y, 1); err != nil {
					common.Log.Debug("can't set pixel at bitmap: %v", bm)
				}
			}
		}
	}
	return bm
}

// Simple linear interpolation from the PDF manual.
func interpolate(x, xMin, xMax, yMin, yMax float64) float64 {
	if math.Abs(xMax-xMin) < 0.000001 {
		return yMin
	}

	y := yMin + (x-xMin)*(yMax-yMin)/(xMax-xMin)
	return y
}
