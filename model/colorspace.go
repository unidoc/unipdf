/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"
	"image/color"
	"math"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// PdfColorspace interface defines the common methods of a PDF colorspace.
// The colorspace defines the data storage format for each color and color representation.
//
// Device based colorspace, specified by name
// - /DeviceGray
// - /DeviceRGB
// - /DeviceCMYK
//
// CIE based colorspace specified by [name, dictionary]
// - [/CalGray dict]
// - [/CalRGB dict]
// - [/Lab dict]
// - [/ICCBased dict]
//
// Special colorspaces
// - /Pattern
// - /Indexed
// - /Separation
// - /DeviceN
//
// Work is in progress to support all colorspaces. At the moment ICCBased color spaces fall back to the alternate
// colorspace which works OK in most cases. For full color support, will need fully featured ICC support.
type PdfColorspace interface {
	// String returns the PdfColorspace's name.
	String() string
	// ImageToRGB converts an Image in a given PdfColorspace to an RGB image.
	ImageToRGB(Image) (Image, error)
	// ColorToRGB converts a single color in a given PdfColorspace to an RGB color.
	ColorToRGB(color PdfColor) (PdfColor, error)
	// GetNumComponents returns the number of components in the PdfColorspace.
	GetNumComponents() int
	// ToPdfObject returns a PdfObject representation of the PdfColorspace.
	ToPdfObject() core.PdfObject
	// ColorFromPdfObjects returns a PdfColor in the given PdfColorspace from an array of PdfObject where each
	// PdfObject represents a numeric value.
	ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error)
	// ColorFromFloats returns a new PdfColor based on input color components for a given PdfColorspace.
	ColorFromFloats(vals []float64) (PdfColor, error)
	// DecodeArray returns the Decode array for the PdfColorSpace, i.e. the range of each component.
	DecodeArray() []float64
}

// PdfColor interface represents a generic color in PDF.
type PdfColor interface {
}

// NewPdfColorspaceFromPdfObject loads a PdfColorspace from a PdfObject.  Returns an error if there is
// a failure in loading.
func NewPdfColorspaceFromPdfObject(obj core.PdfObject) (PdfColorspace, error) {
	var container *core.PdfIndirectObject
	var csName *core.PdfObjectName
	var csArray *core.PdfObjectArray

	if indObj, isInd := obj.(*core.PdfIndirectObject); isInd {
		container = indObj
	}

	// 8.6.3 p. 149 (PDF32000_2008):
	// A colour space shall be defined by an array object whose first element is a name object identifying the
	// colour space family. The remaining array elements, if any, are parameters that further characterize the
	// colour space; their number and types vary according to the particular family.
	//
	// For families that do not require parameters, the colour space may be specified simply by the family name
	// itself instead of an array.

	obj = core.TraceToDirectObject(obj)
	switch t := obj.(type) {
	case *core.PdfObjectArray:
		csArray = t
	case *core.PdfObjectName:
		csName = t
	}

	// If specified by a name directly: Device colorspace or Pattern.
	if csName != nil {
		switch *csName {
		case "DeviceGray":
			return NewPdfColorspaceDeviceGray(), nil
		case "DeviceRGB":
			return NewPdfColorspaceDeviceRGB(), nil
		case "DeviceCMYK":
			return NewPdfColorspaceDeviceCMYK(), nil
		case "Pattern":
			return NewPdfColorspaceSpecialPattern(), nil
		default:
			common.Log.Debug("ERROR: Unknown colorspace %s", *csName)
			return nil, errRangeError
		}
	}

	if csArray != nil && csArray.Len() > 0 {
		var csObject core.PdfObject = container
		if container == nil {
			csObject = csArray
		}
		if name, found := core.GetName(csArray.Get(0)); found {
			switch name.String() {
			case "DeviceGray":
				if csArray.Len() == 1 {
					return NewPdfColorspaceDeviceGray(), nil
				}
			case "DeviceRGB":
				if csArray.Len() == 1 {
					return NewPdfColorspaceDeviceRGB(), nil
				}
			case "DeviceCMYK":
				if csArray.Len() == 1 {
					return NewPdfColorspaceDeviceCMYK(), nil
				}
			case "CalGray":
				return newPdfColorspaceCalGrayFromPdfObject(csObject)
			case "CalRGB":
				return newPdfColorspaceCalRGBFromPdfObject(csObject)
			case "Lab":
				return newPdfColorspaceLabFromPdfObject(csObject)
			case "ICCBased":
				return newPdfColorspaceICCBasedFromPdfObject(csObject)
			case "Pattern":
				return newPdfColorspaceSpecialPatternFromPdfObject(csObject)
			case "Indexed":
				return newPdfColorspaceSpecialIndexedFromPdfObject(csObject)
			case "Separation":
				return newPdfColorspaceSpecialSeparationFromPdfObject(csObject)
			case "DeviceN":
				return newPdfColorspaceDeviceNFromPdfObject(csObject)
			default:
				common.Log.Debug("Array with invalid name: %s", *name)
			}
		}
	}

	common.Log.Debug("PDF File Error: Colorspace type error: %s", obj.String())
	return nil, ErrTypeCheck
}

// DetermineColorspaceNameFromPdfObject determines PDF colorspace from a PdfObject.  Returns the colorspace name and
// an error on failure. If the colorspace was not found, will return an empty string.
func DetermineColorspaceNameFromPdfObject(obj core.PdfObject) (core.PdfObjectName, error) {
	var csName *core.PdfObjectName
	var csArray *core.PdfObjectArray

	if indObj, is := obj.(*core.PdfIndirectObject); is {
		if array, is := indObj.PdfObject.(*core.PdfObjectArray); is {
			csArray = array
		} else if name, is := indObj.PdfObject.(*core.PdfObjectName); is {
			csName = name
		}
	} else if array, is := obj.(*core.PdfObjectArray); is {
		csArray = array
	} else if name, is := obj.(*core.PdfObjectName); is {
		csName = name
	}

	// If specified by a name directly: Device colorspace or Pattern.
	if csName != nil {
		switch *csName {
		case "DeviceGray", "DeviceRGB", "DeviceCMYK":
			return *csName, nil
		case "Pattern":
			return *csName, nil
		}
	}

	if csArray != nil && csArray.Len() > 0 {
		if name, is := csArray.Get(0).(*core.PdfObjectName); is {
			switch *name {
			case "DeviceGray", "DeviceRGB", "DeviceCMYK":
				if csArray.Len() == 1 {
					return *name, nil
				}
			case "CalGray", "CalRGB", "Lab":
				return *name, nil
			case "ICCBased", "Pattern", "Indexed":
				return *name, nil
			case "Separation", "DeviceN":
				return *name, nil
			}
		}
	}

	// Not found
	return "", nil
}

// PdfColorDeviceGray represents a grayscale color value that shall be represented by a single number in the
// range 0.0 to 1.0 where 0.0 corresponds to black and 1.0 to white.
type PdfColorDeviceGray float64

// NewPdfColorDeviceGray returns a new grayscale color based on an input grayscale float value in range [0-1].
func NewPdfColorDeviceGray(grayVal float64) *PdfColorDeviceGray {
	color := PdfColorDeviceGray(grayVal)
	return &color
}

// GetNumComponents returns the number of color components (1 for grayscale).
func (col *PdfColorDeviceGray) GetNumComponents() int {
	return 1
}

// Val returns the color value.
func (col *PdfColorDeviceGray) Val() float64 {
	return float64(*col)
}

// ToInteger convert to an integer format.
func (col *PdfColorDeviceGray) ToInteger(bits int) uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return uint32(maxVal * col.Val())
}

// PdfColorspaceDeviceGray represents a grayscale colorspace.
type PdfColorspaceDeviceGray struct{}

// NewPdfColorspaceDeviceGray returns a new grayscale colorspace.
func NewPdfColorspaceDeviceGray() *PdfColorspaceDeviceGray {
	return &PdfColorspaceDeviceGray{}
}

// GetNumComponents returns the number of color components of the colorspace device.
// Returns 1 for a grayscale device.
func (cs *PdfColorspaceDeviceGray) GetNumComponents() int {
	return 1
}

// DecodeArray returns the range of color component values in DeviceGray colorspace.
func (cs *PdfColorspaceDeviceGray) DecodeArray() []float64 {
	return []float64{0, 1.0}
}

// ToPdfObject returns the PDF representation of the colorspace.
func (cs *PdfColorspaceDeviceGray) ToPdfObject() core.PdfObject {
	return core.MakeName("DeviceGray")
}

func (cs *PdfColorspaceDeviceGray) String() string {
	return "DeviceGray"
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components. The slice should contain a single element between 0 and 1.
func (cs *PdfColorspaceDeviceGray) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("range check")
	}

	val := vals[0]

	if val < 0.0 || val > 1.0 {
		common.Log.Debug("Incompatibility: Range outside [0,1]")
	}

	// Needed for ~/testdata/acl2017_hllz.pdf
	if val < 0.0 {
		val = 0.0
	} else if val > 1.0 {
		val = 1.0
	}

	return NewPdfColorDeviceGray(val), nil
}

// ColorFromPdfObjects returns a new PdfColor based on the input slice of color
// components. The slice should contain a single PdfObjectFloat element in
// range 0-1.
func (cs *PdfColorspaceDeviceGray) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts gray -> rgb for a single color component.
func (cs *PdfColorspaceDeviceGray) ColorToRGB(color PdfColor) (PdfColor, error) {
	gray, ok := color.(*PdfColorDeviceGray)
	if !ok {
		common.Log.Debug("Input color not device gray %T", color)
		return nil, errors.New("type check error")
	}

	return NewPdfColorDeviceRGB(float64(*gray), float64(*gray), float64(*gray)), nil
}

// ImageToRGB convert 1-component grayscale data to 3-component RGB.
func (cs *PdfColorspaceDeviceGray) ImageToRGB(img Image) (Image, error) {
	data := make([]byte, 3*img.Width*img.Height)
	for y := 0; y < int(img.Height); y++ {
		for x := 0; x < int(img.Width); x++ {
			color, err := img.ColorAt(x, y)
			if err != nil {
				return img, err
			}
			r, g, b, _ := color.RGBA()

			idx := (y*int(img.Width) + x) * 3
			data[idx], data[idx+1], data[idx+2] = uint8(r>>8), uint8(g>>8), uint8(b>>8)
		}
	}

	rgbImage := img
	rgbImage.BitsPerComponent = 8
	rgbImage.ColorComponents = 3
	rgbImage.Data = data
	rgbImage.decode = nil

	common.Log.Trace("DeviceGray -> RGB")
	common.Log.Trace("samples: %v", img.Data)
	common.Log.Trace("RGB samples: %v", rgbImage.Data)
	common.Log.Trace("%v -> %v", img, rgbImage)

	return rgbImage, nil
}

// PdfColorDeviceRGB represents a color in DeviceRGB colorspace with R, G, B components, where component is
// defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.
type PdfColorDeviceRGB [3]float64

// NewPdfColorDeviceRGB returns a new PdfColorDeviceRGB based on the r,g,b component values.
func NewPdfColorDeviceRGB(r, g, b float64) *PdfColorDeviceRGB {
	color := PdfColorDeviceRGB{r, g, b}
	return &color
}

// GetNumComponents returns the number of color components (3 for RGB).
func (col *PdfColorDeviceRGB) GetNumComponents() int {
	return 3
}

// R returns the value of the red component of the color.
func (col *PdfColorDeviceRGB) R() float64 {
	return float64(col[0])
}

// G returns the value of the green component of the color.
func (col *PdfColorDeviceRGB) G() float64 {
	return float64(col[1])
}

// B returns the value of the blue component of the color.
func (col *PdfColorDeviceRGB) B() float64 {
	return float64(col[2])
}

// ToInteger convert to an integer format.
func (col *PdfColorDeviceRGB) ToInteger(bits int) [3]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [3]uint32{uint32(maxVal * col.R()), uint32(maxVal * col.G()), uint32(maxVal * col.B())}
}

// ToGray returns a PdfColorDeviceGray color based on the current RGB color.
func (col *PdfColorDeviceRGB) ToGray() *PdfColorDeviceGray {
	// Calculate grayValue [0-1]
	grayValue := 0.3*col.R() + 0.59*col.G() + 0.11*col.B()

	// Clip to [0-1]
	grayValue = math.Min(math.Max(grayValue, 0.0), 1.0)

	return NewPdfColorDeviceGray(grayValue)
}

// RGB colorspace.

// PdfColorspaceDeviceRGB represents an RGB colorspace.
type PdfColorspaceDeviceRGB struct{}

// NewPdfColorspaceDeviceRGB returns a new RGB colorspace object.
func NewPdfColorspaceDeviceRGB() *PdfColorspaceDeviceRGB {
	return &PdfColorspaceDeviceRGB{}
}

func (cs *PdfColorspaceDeviceRGB) String() string {
	return "DeviceRGB"
}

// GetNumComponents returns the number of color components of the colorspace device.
// Returns 3 for an RGB device.
func (cs *PdfColorspaceDeviceRGB) GetNumComponents() int {
	return 3
}

// DecodeArray returns the range of color component values in DeviceRGB colorspace.
func (cs *PdfColorspaceDeviceRGB) DecodeArray() []float64 {
	return []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
}

// ToPdfObject returns the PDF representation of the colorspace.
func (cs *PdfColorspaceDeviceRGB) ToPdfObject() core.PdfObject {
	return core.MakeName("DeviceRGB")
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components. The slice should contain three elements representing the
// red, green and blue components of the color. The values of the elements
// should be between 0 and 1.
func (cs *PdfColorspaceDeviceRGB) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 3 {
		return nil, errors.New("range check")
	}

	// Red.
	r := vals[0]
	if r < 0.0 || r > 1.0 {
		return nil, errors.New("range check")
	}

	// Green.
	g := vals[1]
	if g < 0.0 || g > 1.0 {
		return nil, errors.New("range check")
	}

	// Blue.
	b := vals[2]
	if b < 0.0 || b > 1.0 {
		return nil, errors.New("range check")
	}

	color := NewPdfColorDeviceRGB(r, g, b)
	return color, nil

}

// ColorFromPdfObjects gets the color from a series of pdf objects (3 for rgb).
func (cs *PdfColorspaceDeviceRGB) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 3 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB verifies that the input color is an RGB color. Method exists in
// order to satisfy the PdfColorspace interface.
func (cs *PdfColorspaceDeviceRGB) ColorToRGB(color PdfColor) (PdfColor, error) {
	rgb, ok := color.(*PdfColorDeviceRGB)
	if !ok {
		common.Log.Debug("Input color not device RGB")
		return nil, errors.New("type check error")
	}
	return rgb, nil
}

// ImageToRGB returns the passed in image. Method exists in order to satisfy
// the PdfColorspace interface.
func (cs *PdfColorspaceDeviceRGB) ImageToRGB(img Image) (Image, error) {
	return img, nil
}

// ImageToGray returns a new grayscale image based on the passed in RGB image.
func (cs *PdfColorspaceDeviceRGB) ImageToGray(img Image) (Image, error) {
	grayImage := img

	samples := img.GetSamples()

	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1
	var graySamples []uint32
	for i := 0; i < len(samples); i += 3 {
		// Normalized data, range 0-1.
		r := float64(samples[i]) / maxVal
		g := float64(samples[i+1]) / maxVal
		b := float64(samples[i+2]) / maxVal

		// Calculate grayValue [0-1]
		grayValue := 0.3*r + 0.59*g + 0.11*b

		// Clip to [0-1]
		grayValue = math.Min(math.Max(grayValue, 0.0), 1.0)

		// Convert to uint32
		val := uint32(grayValue * maxVal)
		graySamples = append(graySamples, val)
	}
	grayImage.SetSamples(graySamples)
	grayImage.ColorComponents = 1

	return grayImage, nil
}

//////////////////////
// DeviceCMYK
// C, M, Y, K components.
// No other parameters.

// PdfColorDeviceCMYK is a CMYK color, where each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.
type PdfColorDeviceCMYK [4]float64

// NewPdfColorDeviceCMYK returns a new CMYK color.
func NewPdfColorDeviceCMYK(c, m, y, k float64) *PdfColorDeviceCMYK {
	color := PdfColorDeviceCMYK{c, m, y, k}
	return &color
}

// GetNumComponents returns the number of color components (4 for CMYK).
func (col *PdfColorDeviceCMYK) GetNumComponents() int {
	return 4
}

// C returns the value of the cyan component of the color.
func (col *PdfColorDeviceCMYK) C() float64 {
	return float64(col[0])
}

// M returns the value of the magenta component of the color.
func (col *PdfColorDeviceCMYK) M() float64 {
	return float64(col[1])
}

// Y returns the value of the yellow component of the color.
func (col *PdfColorDeviceCMYK) Y() float64 {
	return float64(col[2])
}

// K returns the value of the key component of the color.
func (col *PdfColorDeviceCMYK) K() float64 {
	return float64(col[3])
}

// ToInteger convert to an integer format.
func (col *PdfColorDeviceCMYK) ToInteger(bits int) [4]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [4]uint32{uint32(maxVal * col.C()), uint32(maxVal * col.M()), uint32(maxVal * col.Y()), uint32(maxVal * col.K())}
}

// PdfColorspaceDeviceCMYK represents a CMYK colorspace.
type PdfColorspaceDeviceCMYK struct{}

// NewPdfColorspaceDeviceCMYK returns a new CMYK colorspace object.
func NewPdfColorspaceDeviceCMYK() *PdfColorspaceDeviceCMYK {
	return &PdfColorspaceDeviceCMYK{}
}

func (cs *PdfColorspaceDeviceCMYK) String() string {
	return "DeviceCMYK"
}

// GetNumComponents returns the number of color components of the colorspace device.
// Returns 4 for a CMYK device.
func (cs *PdfColorspaceDeviceCMYK) GetNumComponents() int {
	return 4
}

// DecodeArray returns the range of color component values in DeviceCMYK colorspace.
func (cs *PdfColorspaceDeviceCMYK) DecodeArray() []float64 {
	return []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
}

// ToPdfObject returns the PDF representation of the colorspace.
func (cs *PdfColorspaceDeviceCMYK) ToPdfObject() core.PdfObject {
	return core.MakeName("DeviceCMYK")
}

// ColorFromFloats returns a new PdfColorDevice based on the input slice of
// color components. The slice should contain four elements representing the
// cyan, magenta, yellow and key components of the color. The values of the
// elements should be between 0 and 1.
func (cs *PdfColorspaceDeviceCMYK) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 4 {
		return nil, errors.New("range check")
	}

	// Cyan
	c := vals[0]
	if c < 0.0 || c > 1.0 {
		return nil, errors.New("range check")
	}

	// Magenta
	m := vals[1]
	if m < 0.0 || m > 1.0 {
		return nil, errors.New("range check")
	}

	// Yellow.
	y := vals[2]
	if y < 0.0 || y > 1.0 {
		return nil, errors.New("range check")
	}

	// Key.
	k := vals[3]
	if k < 0.0 || k > 1.0 {
		return nil, errors.New("range check")
	}

	color := NewPdfColorDeviceCMYK(c, m, y, k)
	return color, nil
}

// ColorFromPdfObjects gets the color from a series of pdf objects (4 for cmyk).
func (cs *PdfColorspaceDeviceCMYK) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 4 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts a CMYK color to an RGB color.
func (cs *PdfColorspaceDeviceCMYK) ColorToRGB(color PdfColor) (PdfColor, error) {
	cmyk, ok := color.(*PdfColorDeviceCMYK)
	if !ok {
		common.Log.Debug("Input color not device cmyk")
		return nil, errors.New("type check error")
	}

	c := cmyk.C()
	m := cmyk.M()
	y := cmyk.Y()
	k := cmyk.K()

	c = c*(1-k) + k
	m = m*(1-k) + k
	y = y*(1-k) + k

	r := 1 - c
	g := 1 - m
	b := 1 - y

	return NewPdfColorDeviceRGB(r, g, b), nil
}

// ImageToRGB converts an image in CMYK colorspace to an RGB image.
func (cs *PdfColorspaceDeviceCMYK) ImageToRGB(img Image) (Image, error) {
	rgbImage := img

	common.Log.Trace("CMYK -> RGB")
	common.Log.Trace("Image BPC: %d, Color components: %d", img.BitsPerComponent, img.ColorComponents)
	common.Log.Trace("Len data: %d", len(img.Data))
	common.Log.Trace("Height: %d, Width: %d", img.Height, img.Width)

	decode := img.decode
	if decode == nil {
		decode = []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
	}
	if len(decode) != 8 {
		common.Log.Debug("Invalid decode array (%d): %.3f", len(decode), decode)
		return img, errors.New("invalid decode array")
	}
	common.Log.Trace("Decode array: %f", decode)

	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1
	common.Log.Trace("MaxVal: %f", maxVal)

	data := make([]byte, 3*img.Width*img.Height)
	for l := 0; l < int(img.Height); l++ {
		for x := 0; x < int(img.Width); x++ {
			col, err := img.ColorAt(x, l)
			if err != nil {
				return img, err
			}
			cmyk, ok := col.(color.CMYK)
			if !ok {
				return img, errors.New("")
			}

			// Normalized c, m, y, k values.
			c := interpolate(float64(cmyk.C), 0, maxVal, decode[0], decode[1])
			m := interpolate(float64(cmyk.M), 0, maxVal, decode[2], decode[3])
			y := interpolate(float64(cmyk.Y), 0, maxVal, decode[4], decode[5])
			k := interpolate(float64(cmyk.K), 0, maxVal, decode[6], decode[7])

			r := uint8(float64(1-(c*(1-k)+k)) * maxVal)
			g := uint8(float64(1-(m*(1-k)+k)) * maxVal)
			b := uint8(float64(1-(y*(1-k)+k)) * maxVal)

			idx := (l*int(img.Width) + x) * 3
			data[idx], data[idx+1], data[idx+2] = r, g, b
		}
	}

	rgbImage.BitsPerComponent = 8
	rgbImage.ColorComponents = 3
	rgbImage.Data = data

	return rgbImage, nil
}

//////////////////////
// CIE based gray level.
// Single component
// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.

// PdfColorCalGray represents a CalGray colorspace.
type PdfColorCalGray float64

// NewPdfColorCalGray returns a new CalGray color.
func NewPdfColorCalGray(grayVal float64) *PdfColorCalGray {
	color := PdfColorCalGray(grayVal)
	return &color
}

// GetNumComponents returns the number of color components (1 for CalGray).
func (col *PdfColorCalGray) GetNumComponents() int {
	return 1
}

// Val returns the value of the color.
func (col *PdfColorCalGray) Val() float64 {
	return float64(*col)
}

// ToInteger convert to an integer format.
func (col *PdfColorCalGray) ToInteger(bits int) uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return uint32(maxVal * col.Val())
}

// PdfColorspaceCalGray represents CalGray color space.
type PdfColorspaceCalGray struct {
	WhitePoint []float64 // [XW, YW, ZW]: Required
	BlackPoint []float64 // [XB, YB, ZB]
	Gamma      float64

	container *core.PdfIndirectObject
}

// NewPdfColorspaceCalGray returns a new CalGray colorspace object.
func NewPdfColorspaceCalGray() *PdfColorspaceCalGray {
	cs := &PdfColorspaceCalGray{}

	// Set optional parameters to default values.
	cs.BlackPoint = []float64{0.0, 0.0, 0.0}
	cs.Gamma = 1

	return cs
}

func (cs *PdfColorspaceCalGray) String() string {
	return "CalGray"
}

// GetNumComponents returns the number of color components of the colorspace device.
// Returns 1 for a CalGray device.
func (cs *PdfColorspaceCalGray) GetNumComponents() int {
	return 1
}

// DecodeArray returns the range of color component values in CalGray colorspace.
func (cs *PdfColorspaceCalGray) DecodeArray() []float64 {
	return []float64{0.0, 1.0}
}

func newPdfColorspaceCalGrayFromPdfObject(obj core.PdfObject) (*PdfColorspaceCalGray, error) {
	cs := NewPdfColorspaceCalGray()

	// If within an indirect object, then make a note of it.  If we write out the PdfObject later
	// we can reference the same container.  Otherwise is not within a container, but rather
	// a new array.
	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = core.TraceToDirectObject(obj)
	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("type error")
	}

	if array.Len() != 2 {
		return nil, fmt.Errorf("invalid CalGray colorspace")
	}

	// Name.
	obj = core.TraceToDirectObject(array.Get(0))
	name, ok := obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("CalGray name not a Name object")
	}
	if *name != "CalGray" {
		return nil, fmt.Errorf("not a CalGray colorspace")
	}

	// Dict.
	obj = core.TraceToDirectObject(array.Get(1))
	dict, ok := obj.(*core.PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("CalGray dict not a Dictionary object")
	}

	// WhitePoint (Required): [Xw, Yw, Zw]
	obj = dict.Get("WhitePoint")
	obj = core.TraceToDirectObject(obj)
	whitePointArray, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("CalGray: Invalid WhitePoint")
	}
	if whitePointArray.Len() != 3 {
		return nil, fmt.Errorf("CalGray: Invalid WhitePoint array")
	}
	whitePoint, err := whitePointArray.GetAsFloat64Slice()
	if err != nil {
		return nil, err
	}
	cs.WhitePoint = whitePoint

	// BlackPoint (Optional)
	obj = dict.Get("BlackPoint")
	if obj != nil {
		obj = core.TraceToDirectObject(obj)
		blackPointArray, ok := obj.(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalGray: Invalid BlackPoint")
		}
		if blackPointArray.Len() != 3 {
			return nil, fmt.Errorf("CalGray: Invalid BlackPoint array")
		}
		blackPoint, err := blackPointArray.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.BlackPoint = blackPoint
	}

	// Gamma (Optional)
	obj = dict.Get("Gamma")
	if obj != nil {
		obj = core.TraceToDirectObject(obj)
		gamma, err := core.GetNumberAsFloat(obj)
		if err != nil {
			return nil, fmt.Errorf("CalGray: gamma not a number")
		}
		cs.Gamma = gamma
	}

	return cs, nil
}

// ToPdfObject return the CalGray colorspace as a PDF object (name dictionary).
func (cs *PdfColorspaceCalGray) ToPdfObject() core.PdfObject {
	// CalGray color space dictionary..
	cspace := &core.PdfObjectArray{}

	cspace.Append(core.MakeName("CalGray"))

	dict := core.MakeDict()
	if cs.WhitePoint != nil {
		dict.Set("WhitePoint", core.MakeArray(core.MakeFloat(cs.WhitePoint[0]), core.MakeFloat(cs.WhitePoint[1]), core.MakeFloat(cs.WhitePoint[2])))
	} else {
		common.Log.Error("CalGray: Missing WhitePoint (Required)")
	}

	if cs.BlackPoint != nil {
		dict.Set("BlackPoint", core.MakeArray(core.MakeFloat(cs.BlackPoint[0]), core.MakeFloat(cs.BlackPoint[1]), core.MakeFloat(cs.BlackPoint[2])))
	}

	dict.Set("Gamma", core.MakeFloat(cs.Gamma))
	cspace.Append(dict)

	if cs.container != nil {
		cs.container.PdfObject = cspace
		return cs.container
	}

	return cspace
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components. The slice should contain a single element between 0 and 1.
func (cs *PdfColorspaceCalGray) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("range check")
	}

	val := vals[0]
	if val < 0.0 || val > 1.0 {
		return nil, errors.New("range check")
	}

	color := NewPdfColorCalGray(val)
	return color, nil
}

// ColorFromPdfObjects returns a new PdfColor based on the input slice of color
// components. The slice should contain a single PdfObjectFloat element in
// range 0-1.
func (cs *PdfColorspaceCalGray) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts a CalGray color to an RGB color.
func (cs *PdfColorspaceCalGray) ColorToRGB(color PdfColor) (PdfColor, error) {
	calgray, ok := color.(*PdfColorCalGray)
	if !ok {
		common.Log.Debug("Input color not cal gray")
		return nil, errors.New("type check error")
	}

	ANorm := calgray.Val()

	// A -> X,Y,Z
	X := cs.WhitePoint[0] * math.Pow(ANorm, cs.Gamma)
	Y := cs.WhitePoint[1] * math.Pow(ANorm, cs.Gamma)
	Z := cs.WhitePoint[2] * math.Pow(ANorm, cs.Gamma)

	// X,Y,Z -> rgb
	// http://stackoverflow.com/questions/21576719/how-to-convert-cie-color-space-into-rgb-or-hex-color-code-in-php
	r := 3.240479*X + -1.537150*Y + -0.498535*Z
	g := -0.969256*X + 1.875992*Y + 0.041556*Z
	b := 0.055648*X + -0.204043*Y + 1.057311*Z

	// Clip.
	r = math.Min(math.Max(r, 0), 1.0)
	g = math.Min(math.Max(g, 0), 1.0)
	b = math.Min(math.Max(b, 0), 1.0)

	return NewPdfColorDeviceRGB(r, g, b), nil
}

// ImageToRGB converts image in CalGray color space to RGB (A, B, C -> X, Y, Z).
func (cs *PdfColorspaceCalGray) ImageToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	var rgbSamples []uint32
	for i := 0; i < len(samples); i++ {
		// A represents the gray component of calibrated gray space.
		// It shall be in the range 0.0 - 1.0
		ANorm := float64(samples[i]) / maxVal

		// A -> X,Y,Z
		X := cs.WhitePoint[0] * math.Pow(ANorm, cs.Gamma)
		Y := cs.WhitePoint[1] * math.Pow(ANorm, cs.Gamma)
		Z := cs.WhitePoint[2] * math.Pow(ANorm, cs.Gamma)

		// X,Y,Z -> rgb
		// http://stackoverflow.com/questions/21576719/how-to-convert-cie-color-space-into-rgb-or-hex-color-code-in-php
		r := 3.240479*X + -1.537150*Y + -0.498535*Z
		g := -0.969256*X + 1.875992*Y + 0.041556*Z
		b := 0.055648*X + -0.204043*Y + 1.057311*Z

		// Clip.
		r = math.Min(math.Max(r, 0), 1.0)
		g = math.Min(math.Max(g, 0), 1.0)
		b = math.Min(math.Max(b, 0), 1.0)

		// Convert to uint32.
		R := uint32(r * maxVal)
		G := uint32(g * maxVal)
		B := uint32(b * maxVal)

		rgbSamples = append(rgbSamples, R, G, B)
	}
	rgbImage.SetSamples(rgbSamples)
	rgbImage.ColorComponents = 3

	return rgbImage, nil
}

// PdfColorCalRGB represents a color in the Colorimetric CIE RGB colorspace.
// A, B, C components
// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.
type PdfColorCalRGB [3]float64

// NewPdfColorCalRGB returns a new CalRBG color.
func NewPdfColorCalRGB(a, b, c float64) *PdfColorCalRGB {
	color := PdfColorCalRGB{a, b, c}
	return &color
}

// GetNumComponents returns the number of color components (3 for CalRGB).
func (col *PdfColorCalRGB) GetNumComponents() int {
	return 3
}

// A returns the value of the A component of the color.
func (col *PdfColorCalRGB) A() float64 {
	return float64(col[0])
}

// B returns the value of the B component of the color.
func (col *PdfColorCalRGB) B() float64 {
	return float64(col[1])
}

// C returns the value of the C component of the color.
func (col *PdfColorCalRGB) C() float64 {
	return float64(col[2])
}

// ToInteger convert to an integer format.
func (col *PdfColorCalRGB) ToInteger(bits int) [3]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [3]uint32{uint32(maxVal * col.A()), uint32(maxVal * col.B()), uint32(maxVal * col.C())}
}

// PdfColorspaceCalRGB stores A, B, C components
type PdfColorspaceCalRGB struct {
	WhitePoint []float64
	BlackPoint []float64
	Gamma      []float64
	Matrix     []float64 // [XA YA ZA XB YB ZB XC YC ZC] ; default value identity [1 0 0 0 1 0 0 0 1]
	dict       *core.PdfObjectDictionary

	container *core.PdfIndirectObject
}

// NewPdfColorspaceCalRGB returns a new CalRGB colorspace object.
func NewPdfColorspaceCalRGB() *PdfColorspaceCalRGB {
	// TODO: require parameters?
	cs := &PdfColorspaceCalRGB{}

	// Set optional parameters to default values.
	cs.BlackPoint = []float64{0.0, 0.0, 0.0}
	cs.Gamma = []float64{1.0, 1.0, 1.0}
	cs.Matrix = []float64{1, 0, 0, 0, 1, 0, 0, 0, 1} // Identity matrix.

	return cs
}

func (cs *PdfColorspaceCalRGB) String() string {
	return "CalRGB"
}

// GetNumComponents returns the number of color components of the colorspace device.
// Returns 3 for a CalRGB device.
func (cs *PdfColorspaceCalRGB) GetNumComponents() int {
	return 3
}

// DecodeArray returns the range of color component values in CalRGB colorspace.
func (cs *PdfColorspaceCalRGB) DecodeArray() []float64 {
	return []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
}

func newPdfColorspaceCalRGBFromPdfObject(obj core.PdfObject) (*PdfColorspaceCalRGB, error) {
	cs := NewPdfColorspaceCalRGB()

	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = core.TraceToDirectObject(obj)
	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("type error")
	}

	if array.Len() != 2 {
		return nil, fmt.Errorf("invalid CalRGB colorspace")
	}

	// Name.
	obj = core.TraceToDirectObject(array.Get(0))
	name, ok := obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("CalRGB name not a Name object")
	}
	if *name != "CalRGB" {
		return nil, fmt.Errorf("not a CalRGB colorspace")
	}

	// Dict.
	obj = core.TraceToDirectObject(array.Get(1))
	dict, ok := obj.(*core.PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("CalRGB name not a Name object")
	}

	// WhitePoint (Required): [Xw, Yw, Zw]
	obj = dict.Get("WhitePoint")
	obj = core.TraceToDirectObject(obj)
	whitePointArray, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("CalRGB: Invalid WhitePoint")
	}
	if whitePointArray.Len() != 3 {
		return nil, fmt.Errorf("CalRGB: Invalid WhitePoint array")
	}
	whitePoint, err := whitePointArray.GetAsFloat64Slice()
	if err != nil {
		return nil, err
	}
	cs.WhitePoint = whitePoint

	// BlackPoint (Optional)
	obj = dict.Get("BlackPoint")
	if obj != nil {
		obj = core.TraceToDirectObject(obj)
		blackPointArray, ok := obj.(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalRGB: Invalid BlackPoint")
		}
		if blackPointArray.Len() != 3 {
			return nil, fmt.Errorf("CalRGB: Invalid BlackPoint array")
		}
		blackPoint, err := blackPointArray.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.BlackPoint = blackPoint
	}

	// Gamma (Optional)
	obj = dict.Get("Gamma")
	if obj != nil {
		obj = core.TraceToDirectObject(obj)
		gammaArray, ok := obj.(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalRGB: Invalid Gamma")
		}
		if gammaArray.Len() != 3 {
			return nil, fmt.Errorf("CalRGB: Invalid Gamma array")
		}
		gamma, err := gammaArray.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.Gamma = gamma
	}

	// Matrix (Optional).
	obj = dict.Get("Matrix")
	if obj != nil {
		obj = core.TraceToDirectObject(obj)
		matrixArray, ok := obj.(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalRGB: Invalid Matrix")
		}
		if matrixArray.Len() != 9 {
			common.Log.Error("Matrix array: %s", matrixArray.String())
			return nil, fmt.Errorf("CalRGB: Invalid Matrix array")
		}
		matrix, err := matrixArray.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.Matrix = matrix
	}

	return cs, nil
}

// ToPdfObject returns colorspace in a PDF object format [name dictionary]
func (cs *PdfColorspaceCalRGB) ToPdfObject() core.PdfObject {
	// CalRGB color space dictionary..
	cspace := &core.PdfObjectArray{}

	cspace.Append(core.MakeName("CalRGB"))

	dict := core.MakeDict()
	if cs.WhitePoint != nil {
		wp := core.MakeArray(core.MakeFloat(cs.WhitePoint[0]), core.MakeFloat(cs.WhitePoint[1]), core.MakeFloat(cs.WhitePoint[2]))
		dict.Set("WhitePoint", wp)
	} else {
		common.Log.Error("CalRGB: Missing WhitePoint (Required)")
	}

	if cs.BlackPoint != nil {
		bp := core.MakeArray(core.MakeFloat(cs.BlackPoint[0]), core.MakeFloat(cs.BlackPoint[1]), core.MakeFloat(cs.BlackPoint[2]))
		dict.Set("BlackPoint", bp)
	}
	if cs.Gamma != nil {
		g := core.MakeArray(core.MakeFloat(cs.Gamma[0]), core.MakeFloat(cs.Gamma[1]), core.MakeFloat(cs.Gamma[2]))
		dict.Set("Gamma", g)
	}
	if cs.Matrix != nil {
		matrix := core.MakeArray(core.MakeFloat(cs.Matrix[0]), core.MakeFloat(cs.Matrix[1]), core.MakeFloat(cs.Matrix[2]),
			core.MakeFloat(cs.Matrix[3]), core.MakeFloat(cs.Matrix[4]), core.MakeFloat(cs.Matrix[5]),
			core.MakeFloat(cs.Matrix[6]), core.MakeFloat(cs.Matrix[7]), core.MakeFloat(cs.Matrix[8]))
		dict.Set("Matrix", matrix)
	}
	cspace.Append(dict)

	if cs.container != nil {
		cs.container.PdfObject = cspace
		return cs.container
	}

	return cspace
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components. The slice should contain three elements representing the
// A, B and C components of the color. The values of the elements should be
// between 0 and 1.
func (cs *PdfColorspaceCalRGB) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 3 {
		return nil, errors.New("range check")
	}

	// A
	a := vals[0]
	if a < 0.0 || a > 1.0 {
		return nil, errors.New("range check")
	}

	// B
	b := vals[1]
	if b < 0.0 || b > 1.0 {
		return nil, errors.New("range check")
	}

	// C.
	c := vals[2]
	if c < 0.0 || c > 1.0 {
		return nil, errors.New("range check")
	}

	color := NewPdfColorCalRGB(a, b, c)
	return color, nil
}

// ColorFromPdfObjects returns a new PdfColor based on the input slice of color
// components. The slice should contain three PdfObjectFloat elements representing
// the A, B and C components of the color.
func (cs *PdfColorspaceCalRGB) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 3 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts a CalRGB color to an RGB color.
func (cs *PdfColorspaceCalRGB) ColorToRGB(color PdfColor) (PdfColor, error) {
	calrgb, ok := color.(*PdfColorCalRGB)
	if !ok {
		common.Log.Debug("Input color not cal rgb")
		return nil, errors.New("type check error")
	}

	// A, B, C in range 0.0 to 1.0
	aVal := calrgb.A()
	bVal := calrgb.B()
	cVal := calrgb.C()

	// A, B, C -> X,Y,Z
	// Gamma [GR GC GB]
	// Matrix [XA YA ZA XB YB ZB XC YC ZC]
	X := cs.Matrix[0]*math.Pow(aVal, cs.Gamma[0]) + cs.Matrix[3]*math.Pow(bVal, cs.Gamma[1]) + cs.Matrix[6]*math.Pow(cVal, cs.Gamma[2])
	Y := cs.Matrix[1]*math.Pow(aVal, cs.Gamma[0]) + cs.Matrix[4]*math.Pow(bVal, cs.Gamma[1]) + cs.Matrix[7]*math.Pow(cVal, cs.Gamma[2])
	Z := cs.Matrix[2]*math.Pow(aVal, cs.Gamma[0]) + cs.Matrix[5]*math.Pow(bVal, cs.Gamma[1]) + cs.Matrix[8]*math.Pow(cVal, cs.Gamma[2])

	// X, Y, Z -> R, G, B
	// http://stackoverflow.com/questions/21576719/how-to-convert-cie-color-space-into-rgb-or-hex-color-code-in-php
	r := 3.240479*X + -1.537150*Y + -0.498535*Z
	g := -0.969256*X + 1.875992*Y + 0.041556*Z
	b := 0.055648*X + -0.204043*Y + 1.057311*Z

	// Clip.
	r = math.Min(math.Max(r, 0), 1.0)
	g = math.Min(math.Max(g, 0), 1.0)
	b = math.Min(math.Max(b, 0), 1.0)

	return NewPdfColorDeviceRGB(r, g, b), nil
}

// ImageToRGB converts CalRGB colorspace image to RGB and returns the result.
func (cs *PdfColorspaceCalRGB) ImageToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	var rgbSamples []uint32
	for i := 0; i < len(samples)-2; i += 3 {
		// A, B, C in range 0.0 to 1.0
		aVal := float64(samples[i]) / maxVal
		bVal := float64(samples[i+1]) / maxVal
		cVal := float64(samples[i+2]) / maxVal

		// A, B, C -> X,Y,Z
		// Gamma [GR GC GB]
		// Matrix [XA YA ZA XB YB ZB XC YC ZC]
		X := cs.Matrix[0]*math.Pow(aVal, cs.Gamma[0]) + cs.Matrix[3]*math.Pow(bVal, cs.Gamma[1]) + cs.Matrix[6]*math.Pow(cVal, cs.Gamma[2])
		Y := cs.Matrix[1]*math.Pow(aVal, cs.Gamma[0]) + cs.Matrix[4]*math.Pow(bVal, cs.Gamma[1]) + cs.Matrix[7]*math.Pow(cVal, cs.Gamma[2])
		Z := cs.Matrix[2]*math.Pow(aVal, cs.Gamma[0]) + cs.Matrix[5]*math.Pow(bVal, cs.Gamma[1]) + cs.Matrix[8]*math.Pow(cVal, cs.Gamma[2])

		// X, Y, Z -> R, G, B
		// http://stackoverflow.com/questions/21576719/how-to-convert-cie-color-space-into-rgb-or-hex-color-code-in-php
		r := 3.240479*X + -1.537150*Y + -0.498535*Z
		g := -0.969256*X + 1.875992*Y + 0.041556*Z
		b := 0.055648*X + -0.204043*Y + 1.057311*Z

		// Clip.
		r = math.Min(math.Max(r, 0), 1.0)
		g = math.Min(math.Max(g, 0), 1.0)
		b = math.Min(math.Max(b, 0), 1.0)

		// Convert to uint32.
		R := uint32(r * maxVal)
		G := uint32(g * maxVal)
		B := uint32(b * maxVal)

		rgbSamples = append(rgbSamples, R, G, B)
	}
	rgbImage.SetSamples(rgbSamples)
	rgbImage.ColorComponents = 3

	return rgbImage, nil
}

// PdfColorLab represents a color in the L*, a*, b* 3 component colorspace.
// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.
type PdfColorLab [3]float64

// NewPdfColorLab returns a new Lab color.
func NewPdfColorLab(l, a, b float64) *PdfColorLab {
	color := PdfColorLab{l, a, b}
	return &color
}

// GetNumComponents returns the number of color components (3 for Lab).
func (col *PdfColorLab) GetNumComponents() int {
	return 3
}

// L returns the value of the L component of the color.
func (col *PdfColorLab) L() float64 {
	return float64(col[0])
}

// A returns the value of the A component of the color.
func (col *PdfColorLab) A() float64 {
	return float64(col[1])
}

// B returns the value of the B component of the color.
func (col *PdfColorLab) B() float64 {
	return float64(col[2])
}

// ToInteger convert to an integer format.
func (col *PdfColorLab) ToInteger(bits int) [3]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [3]uint32{uint32(maxVal * col.L()), uint32(maxVal * col.A()), uint32(maxVal * col.B())}
}

// PdfColorspaceLab is a L*, a*, b* 3 component colorspace.
type PdfColorspaceLab struct {
	WhitePoint []float64 // Required.
	BlackPoint []float64
	Range      []float64 // [amin amax bmin bmax]

	container *core.PdfIndirectObject
}

func (cs *PdfColorspaceLab) String() string {
	return "Lab"
}

// GetNumComponents returns the number of color components of the colorspace device.
// Returns 3 for a Lab device.
func (cs *PdfColorspaceLab) GetNumComponents() int {
	return 3
}

// DecodeArray returns the range of color component values in the Lab colorspace.
func (cs *PdfColorspaceLab) DecodeArray() []float64 {
	// Range for L
	decode := []float64{0, 100}

	// Range for A,B specified by range or default
	if cs.Range != nil && len(cs.Range) == 4 {
		decode = append(decode, cs.Range...)
	} else {
		decode = append(decode, -100, 100, -100, 100)
	}

	return decode
}

// NewPdfColorspaceLab returns a new Lab colorspace object.
func NewPdfColorspaceLab() *PdfColorspaceLab {
	// TODO: require parameters?
	cs := &PdfColorspaceLab{}

	// Set optional parameters to default values.
	cs.BlackPoint = []float64{0.0, 0.0, 0.0}
	cs.Range = []float64{-100, 100, -100, 100} // Identity matrix.

	return cs
}

func newPdfColorspaceLabFromPdfObject(obj core.PdfObject) (*PdfColorspaceLab, error) {
	cs := NewPdfColorspaceLab()

	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = core.TraceToDirectObject(obj)
	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("type error")
	}

	if array.Len() != 2 {
		return nil, fmt.Errorf("invalid CalRGB colorspace")
	}

	// Name.
	obj = core.TraceToDirectObject(array.Get(0))
	name, ok := obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("lab name not a Name object")
	}
	if *name != "Lab" {
		return nil, fmt.Errorf("not a Lab colorspace")
	}

	// Dict.
	obj = core.TraceToDirectObject(array.Get(1))
	dict, ok := obj.(*core.PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("colorspace dictionary missing or invalid")
	}

	// WhitePoint (Required): [Xw, Yw, Zw]
	obj = dict.Get("WhitePoint")
	obj = core.TraceToDirectObject(obj)
	whitePointArray, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Lab Invalid WhitePoint")
	}
	if whitePointArray.Len() != 3 {
		return nil, fmt.Errorf("Lab: Invalid WhitePoint array")
	}
	whitePoint, err := whitePointArray.GetAsFloat64Slice()
	if err != nil {
		return nil, err
	}
	cs.WhitePoint = whitePoint

	// BlackPoint (Optional)
	obj = dict.Get("BlackPoint")
	if obj != nil {
		obj = core.TraceToDirectObject(obj)
		blackPointArray, ok := obj.(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Lab: Invalid BlackPoint")
		}
		if blackPointArray.Len() != 3 {
			return nil, fmt.Errorf("Lab: Invalid BlackPoint array")
		}
		blackPoint, err := blackPointArray.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.BlackPoint = blackPoint
	}

	// Range (Optional)
	obj = dict.Get("Range")
	if obj != nil {
		obj = core.TraceToDirectObject(obj)
		rangeArray, ok := obj.(*core.PdfObjectArray)
		if !ok {
			common.Log.Error("Range type error")
			return nil, fmt.Errorf("Lab: Type error")
		}
		if rangeArray.Len() != 4 {
			common.Log.Error("Range range error")
			return nil, fmt.Errorf("Lab: Range error")
		}
		rang, err := rangeArray.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.Range = rang
	}

	return cs, nil
}

// ToPdfObject returns colorspace in a PDF object format [name dictionary]
func (cs *PdfColorspaceLab) ToPdfObject() core.PdfObject {
	// CalRGB color space dictionary..
	csObj := core.MakeArray()

	csObj.Append(core.MakeName("Lab"))

	dict := core.MakeDict()
	if cs.WhitePoint != nil {
		wp := core.MakeArray(core.MakeFloat(cs.WhitePoint[0]), core.MakeFloat(cs.WhitePoint[1]), core.MakeFloat(cs.WhitePoint[2]))
		dict.Set("WhitePoint", wp)
	} else {
		common.Log.Error("Lab: Missing WhitePoint (Required)")
	}

	if cs.BlackPoint != nil {
		bp := core.MakeArray(core.MakeFloat(cs.BlackPoint[0]), core.MakeFloat(cs.BlackPoint[1]), core.MakeFloat(cs.BlackPoint[2]))
		dict.Set("BlackPoint", bp)
	}

	if cs.Range != nil {
		val := core.MakeArray(core.MakeFloat(cs.Range[0]), core.MakeFloat(cs.Range[1]), core.MakeFloat(cs.Range[2]), core.MakeFloat(cs.Range[3]))
		dict.Set("Range", val)
	}
	csObj.Append(dict)

	if cs.container != nil {
		cs.container.PdfObject = csObj
		return cs.container
	}

	return csObj
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components. The slice should contain three elements representing the
// L (range 0-100), A (range -100-100) and B (range -100-100) components of
// the color.
func (cs *PdfColorspaceLab) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 3 {
		return nil, errors.New("range check")
	}

	// L
	l := vals[0]
	if l < 0.0 || l > 100.0 {
		common.Log.Debug("L out of range (got %v should be 0-100)", l)
		return nil, errors.New("range check")
	}

	// A
	a := vals[1]
	aMin := float64(-100)
	aMax := float64(100)
	if len(cs.Range) > 1 {
		aMin = cs.Range[0]
		aMax = cs.Range[1]
	}
	if a < aMin || a > aMax {
		common.Log.Debug("A out of range (got %v; range %v to %v)", a, aMin, aMax)
		return nil, errors.New("range check")
	}

	// B.
	b := vals[2]
	bMin := float64(-100)
	bMax := float64(100)
	if len(cs.Range) > 3 {
		bMin = cs.Range[2]
		bMax = cs.Range[3]
	}
	if b < bMin || b > bMax {
		common.Log.Debug("b out of range (got %v; range %v to %v)", b, bMin, bMax)
		return nil, errors.New("range check")
	}

	color := NewPdfColorLab(l, a, b)
	return color, nil
}

// ColorFromPdfObjects returns a new PdfColor based on the input slice of color
// components. The slice should contain three PdfObjectFloat elements representing
// the L, A and B components of the color.
func (cs *PdfColorspaceLab) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 3 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts a Lab color to an RGB color.
func (cs *PdfColorspaceLab) ColorToRGB(color PdfColor) (PdfColor, error) {
	gFunc := func(x float64) float64 {
		if x >= 6.0/29 {
			return x * x * x
		}
		return 108.0 / 841 * (x - 4/29)
	}

	lab, ok := color.(*PdfColorLab)
	if !ok {
		common.Log.Debug("input color not lab")
		return nil, errors.New("type check error")
	}

	// Get L*, a*, b* values.
	LStar := lab.L()
	AStar := lab.A()
	BStar := lab.B()

	// Convert L*,a*,b* -> L, M, N
	L := (LStar+16)/116 + AStar/500
	M := (LStar + 16) / 116
	N := (LStar+16)/116 - BStar/200

	// L, M, N -> X,Y,Z
	X := cs.WhitePoint[0] * gFunc(L)
	Y := cs.WhitePoint[1] * gFunc(M)
	Z := cs.WhitePoint[2] * gFunc(N)

	// Convert to RGB.
	// X, Y, Z -> R, G, B
	// http://stackoverflow.com/questions/21576719/how-to-convert-cie-color-space-into-rgb-or-hex-color-code-in-php
	r := 3.240479*X + -1.537150*Y + -0.498535*Z
	g := -0.969256*X + 1.875992*Y + 0.041556*Z
	b := 0.055648*X + -0.204043*Y + 1.057311*Z

	// Clip.
	r = math.Min(math.Max(r, 0), 1.0)
	g = math.Min(math.Max(g, 0), 1.0)
	b = math.Min(math.Max(b, 0), 1.0)

	return NewPdfColorDeviceRGB(r, g, b), nil
}

// ImageToRGB converts Lab colorspace image to RGB and returns the result.
func (cs *PdfColorspaceLab) ImageToRGB(img Image) (Image, error) {
	g := func(x float64) float64 {
		if x >= 6.0/29 {
			return x * x * x
		}
		return 108.0 / 841 * (x - 4/29)
	}

	rgbImage := img

	// Each n-bit unit within the bit stream shall be interpreted as an unsigned integer in the range 0 to 2n- 1,
	// with the high-order bit first.
	// The image dictionary’s Decode entry maps this integer to a colour component value, equivalent to what could be
	// used with colour operators such as sc or g.

	componentRanges := img.decode
	if len(componentRanges) != 6 {
		// If image's Decode not appropriate, fall back to default decode array.
		common.Log.Trace("Image - Lab Decode range != 6... use [0 100 amin amax bmin bmax] default decode array")
		componentRanges = cs.DecodeArray()
	}

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	var rgbSamples []uint32
	for i := 0; i < len(samples); i += 3 {
		// Get normalized L*, a*, b* values. [0-1]
		LNorm := float64(samples[i]) / maxVal
		ANorm := float64(samples[i+1]) / maxVal
		BNorm := float64(samples[i+2]) / maxVal

		LStar := interpolate(LNorm, 0.0, 1.0, componentRanges[0], componentRanges[1])
		AStar := interpolate(ANorm, 0.0, 1.0, componentRanges[2], componentRanges[3])
		BStar := interpolate(BNorm, 0.0, 1.0, componentRanges[4], componentRanges[5])

		// Convert L*,a*,b* -> L, M, N
		L := (LStar+16)/116 + AStar/500
		M := (LStar + 16) / 116
		N := (LStar+16)/116 - BStar/200

		// L, M, N -> X,Y,Z
		X := cs.WhitePoint[0] * g(L)
		Y := cs.WhitePoint[1] * g(M)
		Z := cs.WhitePoint[2] * g(N)

		// Convert to RGB.
		// X, Y, Z -> R, G, B
		// http://stackoverflow.com/questions/21576719/how-to-convert-cie-color-space-into-rgb-or-hex-color-code-in-php
		r := 3.240479*X + -1.537150*Y + -0.498535*Z
		g := -0.969256*X + 1.875992*Y + 0.041556*Z
		b := 0.055648*X + -0.204043*Y + 1.057311*Z

		// Clip.
		r = math.Min(math.Max(r, 0), 1.0)
		g = math.Min(math.Max(g, 0), 1.0)
		b = math.Min(math.Max(b, 0), 1.0)

		// Convert to uint32.
		R := uint32(r * maxVal)
		G := uint32(g * maxVal)
		B := uint32(b * maxVal)

		rgbSamples = append(rgbSamples, R, G, B)
	}
	rgbImage.SetSamples(rgbSamples)
	rgbImage.ColorComponents = 3

	return rgbImage, nil
}

//////////////////////
// ICC Based colors.
// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.

/*
type PdfColorICCBased []float64

func NewPdfColorICCBased(vals []float64) *PdfColorICCBased {
	color := PdfColorICCBased{}
	for _, val := range vals {
		color = append(color, val)
	}
	return &color
}

func (this *PdfColorICCBased) GetNumComponents() int {
	return len(*this)
}

// Convert to an integer format.
func (this *PdfColorICCBased) ToInteger(bits int) []uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	ints := []uint32{}
	for _, val := range *this {
		ints = append(ints, uint32(maxVal*val))
	}

	return ints

}
*/
// See p. 157 for calculations...

// PdfColorspaceICCBased format [/ICCBased stream]
//
// The stream shall contain the ICC profile.
// A conforming reader shall support ICC.1:2004:10 as required by PDF 1.7, which will enable it
// to properly render all embedded ICC profiles regardless of the PDF version
//
// In the current implementation, we rely on the alternative colormap provided.
type PdfColorspaceICCBased struct {
	N         int           // Number of color components (Required). Can be 1,3, or 4.
	Alternate PdfColorspace // Alternate colorspace for non-conforming readers.
	// If omitted ICC not supported: then use DeviceGray,
	// DeviceRGB or DeviceCMYK for N=1,3,4 respectively.
	Range    []float64             // Array of 2xN numbers, specifying range of each color component.
	Metadata *core.PdfObjectStream // Metadata stream.
	Data     []byte                // ICC colormap data.

	container *core.PdfIndirectObject
	stream    *core.PdfObjectStream
}

// GetNumComponents returns the number of color components.
func (cs *PdfColorspaceICCBased) GetNumComponents() int {
	return cs.N
}

// DecodeArray returns the range of color component values in the ICCBased colorspace.
func (cs *PdfColorspaceICCBased) DecodeArray() []float64 {
	return cs.Range
}

func (cs *PdfColorspaceICCBased) String() string {
	return "ICCBased"
}

// NewPdfColorspaceICCBased returns a new ICCBased colorspace object.
func NewPdfColorspaceICCBased(N int) (*PdfColorspaceICCBased, error) {
	cs := &PdfColorspaceICCBased{}

	if N != 1 && N != 3 && N != 4 {
		return nil, fmt.Errorf("invalid N (1/3/4)")
	}

	cs.N = N

	return cs, nil
}

// Input format [/ICCBased stream]
func newPdfColorspaceICCBasedFromPdfObject(obj core.PdfObject) (*PdfColorspaceICCBased, error) {
	cs := &PdfColorspaceICCBased{}
	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = core.TraceToDirectObject(obj)

	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("type error")
	}

	if array.Len() != 2 {
		return nil, fmt.Errorf("invalid ICCBased colorspace")
	}

	// Name.
	obj = core.TraceToDirectObject(array.Get(0))
	name, ok := obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("ICCBased name not a Name object")
	}
	if *name != "ICCBased" {
		return nil, fmt.Errorf("not an ICCBased colorspace")
	}

	// Stream
	obj = array.Get(1)
	stream, ok := core.GetStream(obj)
	if !ok {
		common.Log.Error("ICCBased not pointing to stream: %T", obj)
		return nil, fmt.Errorf("ICCBased stream invalid")
	}

	dict := stream.PdfObjectDictionary

	n, ok := dict.Get("N").(*core.PdfObjectInteger)
	if !ok {
		return nil, fmt.Errorf("ICCBased missing N from stream dict")
	}
	if *n != 1 && *n != 3 && *n != 4 {
		return nil, fmt.Errorf("ICCBased colorspace invalid N (not 1,3,4)")
	}
	cs.N = int(*n)

	if obj := dict.Get("Alternate"); obj != nil {
		alternate, err := NewPdfColorspaceFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		cs.Alternate = alternate
	}

	if obj := dict.Get("Range"); obj != nil {
		obj = core.TraceToDirectObject(obj)
		array, ok := obj.(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("ICCBased Range not an array")
		}
		if array.Len() != 2*cs.N {
			return nil, fmt.Errorf("ICCBased Range wrong number of elements")
		}
		r, err := array.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.Range = r
	} else {
		// Set defaults
		cs.Range = make([]float64, 2*cs.N)
		for i := 0; i < cs.N; i++ {
			cs.Range[2*i] = 0.0
			cs.Range[2*i+1] = 1.0
		}
	}

	if obj := dict.Get("Metadata"); obj != nil {
		stream, ok := obj.(*core.PdfObjectStream)
		if !ok {
			return nil, fmt.Errorf("ICCBased Metadata not a stream")
		}
		cs.Metadata = stream
	}

	data, err := core.DecodeStream(stream)
	if err != nil {
		return nil, err
	}
	cs.Data = data
	cs.stream = stream

	return cs, nil
}

// ToPdfObject returns colorspace in a PDF object format [name stream]
func (cs *PdfColorspaceICCBased) ToPdfObject() core.PdfObject {
	csObj := &core.PdfObjectArray{}

	csObj.Append(core.MakeName("ICCBased"))

	var stream *core.PdfObjectStream
	if cs.stream != nil {
		stream = cs.stream
	} else {
		stream = &core.PdfObjectStream{}
	}
	dict := core.MakeDict()

	dict.Set("N", core.MakeInteger(int64(cs.N)))

	if cs.Alternate != nil {
		dict.Set("Alternate", cs.Alternate.ToPdfObject())
	}

	if cs.Metadata != nil {
		dict.Set("Metadata", cs.Metadata)
	}
	if cs.Range != nil {
		var ranges []core.PdfObject
		for _, r := range cs.Range {
			ranges = append(ranges, core.MakeFloat(r))
		}
		dict.Set("Range", core.MakeArray(ranges...))
	}

	// Encode with a default encoder?
	dict.Set("Length", core.MakeInteger(int64(len(cs.Data))))
	// Need to have a representation of the stream...
	stream.Stream = cs.Data
	stream.PdfObjectDictionary = dict

	csObj.Append(stream)

	if cs.container != nil {
		cs.container.PdfObject = csObj
		return cs.container
	}

	return csObj
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components.
func (cs *PdfColorspaceICCBased) ColorFromFloats(vals []float64) (PdfColor, error) {
	if cs.Alternate == nil {
		if cs.N == 1 {
			cs := NewPdfColorspaceDeviceGray()
			return cs.ColorFromFloats(vals)
		} else if cs.N == 3 {
			cs := NewPdfColorspaceDeviceRGB()
			return cs.ColorFromFloats(vals)
		} else if cs.N == 4 {
			cs := NewPdfColorspaceDeviceCMYK()
			return cs.ColorFromFloats(vals)
		} else {
			return nil, errors.New("ICC Based colorspace missing alternative")
		}
	}

	return cs.Alternate.ColorFromFloats(vals)
}

// ColorFromPdfObjects returns a new PdfColor based on the input slice of color
// component PDF objects.
func (cs *PdfColorspaceICCBased) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if cs.Alternate == nil {
		if cs.N == 1 {
			cs := NewPdfColorspaceDeviceGray()
			return cs.ColorFromPdfObjects(objects)
		} else if cs.N == 3 {
			cs := NewPdfColorspaceDeviceRGB()
			return cs.ColorFromPdfObjects(objects)
		} else if cs.N == 4 {
			cs := NewPdfColorspaceDeviceCMYK()
			return cs.ColorFromPdfObjects(objects)
		} else {
			return nil, errors.New("ICC Based colorspace missing alternative")
		}
	}

	return cs.Alternate.ColorFromPdfObjects(objects)
}

// ColorToRGB converts a ICCBased color to an RGB color.
func (cs *PdfColorspaceICCBased) ColorToRGB(color PdfColor) (PdfColor, error) {
	if cs.Alternate == nil {
		common.Log.Debug("ICC Based colorspace missing alternative")
		if cs.N == 1 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceGray (N=1)")
			grayCS := NewPdfColorspaceDeviceGray()
			return grayCS.ColorToRGB(color)
		} else if cs.N == 3 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceRGB (N=3)")
			// Already in RGB.
			return color, nil
		} else if cs.N == 4 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceCMYK (N=4)")
			// CMYK
			cmykCS := NewPdfColorspaceDeviceCMYK()
			return cmykCS.ColorToRGB(color)
		} else {
			return nil, errors.New("ICC Based colorspace missing alternative")
		}
	}

	common.Log.Trace("ICC Based colorspace with alternative: %#v", cs)
	return cs.Alternate.ColorToRGB(color)
}

// ImageToRGB converts ICCBased colorspace image to RGB and returns the result.
func (cs *PdfColorspaceICCBased) ImageToRGB(img Image) (Image, error) {
	if cs.Alternate == nil {
		common.Log.Debug("ICC Based colorspace missing alternative")
		if cs.N == 1 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceGray (N=1)")
			grayCS := NewPdfColorspaceDeviceGray()
			return grayCS.ImageToRGB(img)
		} else if cs.N == 3 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceRGB (N=3)")
			// Already in RGB.
			return img, nil
		} else if cs.N == 4 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceCMYK (N=4)")
			// CMYK
			cmykCS := NewPdfColorspaceDeviceCMYK()
			return cmykCS.ImageToRGB(img)
		} else {
			return img, errors.New("ICC Based colorspace missing alternative")
		}
	}
	common.Log.Trace("ICC Based colorspace with alternative: %#v", cs)

	output, err := cs.Alternate.ImageToRGB(img)
	common.Log.Trace("ICC Input image: %+v", img)
	common.Log.Trace("ICC Output image: %+v", output)
	return output, err //cs.Alternate.ImageToRGB(img)
}

// PdfColorPattern represents a pattern color.
type PdfColorPattern struct {
	Color       PdfColor           // Color defined in underlying colorspace.
	PatternName core.PdfObjectName // Name of the pattern (reference via resource dicts).
}

// PdfColorspaceSpecialPattern is a Pattern colorspace.
// Can be defined either as /Pattern or with an underlying colorspace [/Pattern cs].
type PdfColorspaceSpecialPattern struct {
	UnderlyingCS PdfColorspace

	container *core.PdfIndirectObject
}

// NewPdfColorspaceSpecialPattern returns a new pattern color.
func NewPdfColorspaceSpecialPattern() *PdfColorspaceSpecialPattern {
	return &PdfColorspaceSpecialPattern{}
}

func (cs *PdfColorspaceSpecialPattern) String() string {
	return "Pattern"
}

// GetNumComponents returns the number of color components of the underlying
// colorspace device.
func (cs *PdfColorspaceSpecialPattern) GetNumComponents() int {
	return cs.UnderlyingCS.GetNumComponents()
}

// DecodeArray returns an empty slice as there are no components associated with pattern colorspace.
func (cs *PdfColorspaceSpecialPattern) DecodeArray() []float64 {
	return []float64{}
}

func newPdfColorspaceSpecialPatternFromPdfObject(obj core.PdfObject) (*PdfColorspaceSpecialPattern, error) {
	common.Log.Trace("New Pattern CS from obj: %s %T", obj.String(), obj)
	cs := NewPdfColorspaceSpecialPattern()

	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = core.TraceToDirectObject(obj)
	if name, isName := obj.(*core.PdfObjectName); isName {
		if *name != "Pattern" {
			return nil, fmt.Errorf("invalid name")
		}

		return cs, nil
	}

	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		common.Log.Error("Invalid Pattern CS Object: %#v", obj)
		return nil, fmt.Errorf("invalid Pattern CS object")
	}
	if array.Len() != 1 && array.Len() != 2 {
		common.Log.Error("Invalid Pattern CS array: %#v", array)
		return nil, fmt.Errorf("invalid Pattern CS array")
	}

	obj = array.Get(0)
	if name, isName := obj.(*core.PdfObjectName); isName {
		if *name != "Pattern" {
			common.Log.Error("Invalid Pattern CS array name: %#v", name)
			return nil, fmt.Errorf("invalid name")
		}
	}

	// Has an underlying color space.
	if array.Len() > 1 {
		obj = array.Get(1)
		obj = core.TraceToDirectObject(obj)
		baseCS, err := NewPdfColorspaceFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		cs.UnderlyingCS = baseCS
	}

	common.Log.Trace("Returning Pattern with underlying cs: %T", cs.UnderlyingCS)
	return cs, nil
}

// ToPdfObject returns the PDF representation of the colorspace.
func (cs *PdfColorspaceSpecialPattern) ToPdfObject() core.PdfObject {
	if cs.UnderlyingCS == nil {
		return core.MakeName("Pattern")
	}

	csObj := core.MakeArray(core.MakeName("Pattern"))
	csObj.Append(cs.UnderlyingCS.ToPdfObject())

	if cs.container != nil {
		cs.container.PdfObject = csObj
		return cs.container
	}

	return csObj
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components.
func (cs *PdfColorspaceSpecialPattern) ColorFromFloats(vals []float64) (PdfColor, error) {
	if cs.UnderlyingCS == nil {
		return nil, errors.New("underlying CS not specified")
	}
	return cs.UnderlyingCS.ColorFromFloats(vals)
}

// ColorFromPdfObjects loads the color from PDF objects.
// The first objects (if present) represent the color in underlying colorspace.  The last one represents
// the name of the pattern.
func (cs *PdfColorspaceSpecialPattern) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) < 1 {
		return nil, errors.New("invalid number of parameters")
	}
	patternColor := &PdfColorPattern{}

	// Pattern name.
	pname, ok := objects[len(objects)-1].(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("Pattern name not a name (got %T)", objects[len(objects)-1])
		return nil, ErrTypeCheck
	}
	patternColor.PatternName = *pname

	// Pattern color if specified.
	if len(objects) > 1 {
		colorObjs := objects[0 : len(objects)-1]
		if cs.UnderlyingCS == nil {
			common.Log.Debug("Pattern color with defined color components but underlying cs missing")
			return nil, errors.New("underlying CS not defined")
		}
		color, err := cs.UnderlyingCS.ColorFromPdfObjects(colorObjs)
		if err != nil {
			common.Log.Debug("ERROR: Unable to convert color via underlying cs: %v", err)
			return nil, err
		}
		patternColor.Color = color
	}

	return patternColor, nil
}

// ColorToRGB only converts color used with uncolored patterns (defined in underlying colorspace).  Does not go into the
// pattern objects and convert those.  If that is desired, needs to be done separately.  See for example
// grayscale conversion example in unidoc-examples repo.
func (cs *PdfColorspaceSpecialPattern) ColorToRGB(color PdfColor) (PdfColor, error) {
	patternColor, ok := color.(*PdfColorPattern)
	if !ok {
		common.Log.Debug("Color not pattern (got %T)", color)
		return nil, ErrTypeCheck
	}

	if patternColor.Color == nil {
		// No color defined, can return same back.  No transform needed.
		return color, nil
	}

	if cs.UnderlyingCS == nil {
		return nil, errors.New("underlying CS not defined")
	}

	return cs.UnderlyingCS.ColorToRGB(patternColor.Color)
}

// ImageToRGB returns an error since an image cannot be defined in a pattern colorspace.
func (cs *PdfColorspaceSpecialPattern) ImageToRGB(img Image) (Image, error) {
	common.Log.Debug("Error: Image cannot be specified in Pattern colorspace")
	return img, errors.New("invalid colorspace for image (pattern)")
}

// PdfColorspaceSpecialIndexed is an indexed color space is a lookup table, where the input element
// is an index to the lookup table and the output is a color defined in the lookup table in the Base
// colorspace.
// [/Indexed base hival lookup]
type PdfColorspaceSpecialIndexed struct {
	Base   PdfColorspace
	HiVal  int
	Lookup core.PdfObject

	colorLookup []byte // m*(hival+1); m is number of components in Base colorspace

	container *core.PdfIndirectObject
}

// NewPdfColorspaceSpecialIndexed returns a new Indexed color.
func NewPdfColorspaceSpecialIndexed() *PdfColorspaceSpecialIndexed {
	return &PdfColorspaceSpecialIndexed{HiVal: 255}
}

func (cs *PdfColorspaceSpecialIndexed) String() string {
	return "Indexed"
}

// GetNumComponents returns the number of color components (1 for Indexed).
func (cs *PdfColorspaceSpecialIndexed) GetNumComponents() int {
	return 1
}

// DecodeArray returns the component range values for the Indexed colorspace.
func (cs *PdfColorspaceSpecialIndexed) DecodeArray() []float64 {
	return []float64{0, float64(cs.HiVal)}
}

func newPdfColorspaceSpecialIndexedFromPdfObject(obj core.PdfObject) (*PdfColorspaceSpecialIndexed, error) {
	cs := NewPdfColorspaceSpecialIndexed()

	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = core.TraceToDirectObject(obj)
	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("type error")
	}

	if array.Len() != 4 {
		return nil, fmt.Errorf("indexed CS: invalid array length")
	}

	// Check name.
	obj = array.Get(0)
	name, ok := obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("indexed CS: invalid name")
	}
	if *name != "Indexed" {
		return nil, fmt.Errorf("indexed CS: wrong name")
	}

	// Get base colormap.
	obj = array.Get(1)

	// Base cs cannot be another /Indexed or /Pattern space.
	baseName, err := DetermineColorspaceNameFromPdfObject(obj)
	if baseName == "Indexed" || baseName == "Pattern" {
		common.Log.Debug("Error: Indexed colorspace cannot have Indexed/Pattern CS as base (%v)", baseName)
		return nil, errRangeError
	}

	baseCs, err := NewPdfColorspaceFromPdfObject(obj)
	if err != nil {
		return nil, err
	}
	cs.Base = baseCs

	// Get hi val.
	obj = array.Get(2)
	val, err := core.GetNumberAsInt64(obj)
	if err != nil {
		return nil, err
	}
	if val > 255 {
		return nil, fmt.Errorf("indexed CS: Invalid hival")
	}
	cs.HiVal = int(val)

	// Index table.
	obj = array.Get(3)
	cs.Lookup = obj
	obj = core.TraceToDirectObject(obj)
	var data []byte
	if str, ok := obj.(*core.PdfObjectString); ok {
		data = str.Bytes()
		common.Log.Trace("Indexed string color data: % d", data)
	} else if stream, ok := obj.(*core.PdfObjectStream); ok {
		common.Log.Trace("Indexed stream: %s", obj.String())
		common.Log.Trace("Encoded (%d) : %# x", len(stream.Stream), stream.Stream)
		decoded, err := core.DecodeStream(stream)
		if err != nil {
			return nil, err
		}
		common.Log.Trace("Decoded (%d) : % X", len(decoded), decoded)
		data = decoded
	} else {
		common.Log.Debug("Type: %T", obj)
		return nil, fmt.Errorf("indexed CS: Invalid table format")
	}

	if len(data) < cs.Base.GetNumComponents()*(cs.HiVal+1) {
		// Sometimes the table length is too short.  In this case we need to
		// note what absolute maximum index is.
		common.Log.Debug("PDF Incompatibility: Index stream too short")
		common.Log.Debug("Fail, len(data): %d, components: %d, hiVal: %d", len(data), cs.Base.GetNumComponents(), cs.HiVal)
	} else {
		// trim
		data = data[:cs.Base.GetNumComponents()*(cs.HiVal+1)]
	}

	cs.colorLookup = data

	return cs, nil
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components. The slice should contain a single element.
func (cs *PdfColorspaceSpecialIndexed) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("range check")
	}

	N := cs.Base.GetNumComponents()

	index := int(vals[0]) * N
	if index < 0 || (index+N-1) >= len(cs.colorLookup) {
		return nil, errors.New("outside range")
	}

	cvals := cs.colorLookup[index : index+N]
	var floats []float64
	for _, val := range cvals {
		floats = append(floats, float64(val)/255.0)
	}
	color, err := cs.Base.ColorFromFloats(floats)
	if err != nil {
		return nil, err
	}

	return color, nil
}

// ColorFromPdfObjects returns a new PdfColor based on the input slice of color
// components. The slice should contain a single PdfObjectFloat element.
func (cs *PdfColorspaceSpecialIndexed) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts an Indexed color to an RGB color.
func (cs *PdfColorspaceSpecialIndexed) ColorToRGB(color PdfColor) (PdfColor, error) {
	if cs.Base == nil {
		return nil, errors.New("indexed base colorspace undefined")
	}

	return cs.Base.ColorToRGB(color)
}

// ImageToRGB convert an indexed image to RGB.
func (cs *PdfColorspaceSpecialIndexed) ImageToRGB(img Image) (Image, error) {
	//baseImage := img
	// Make a new representation of the image to be converted with the base colorspace.
	baseImage := Image{}
	baseImage.Height = img.Height
	baseImage.Width = img.Width
	baseImage.alphaData = img.alphaData
	// TODO(peterwilliams97): Add support for other BitsPerComponent values.
	// See https://github.com/unidoc/unipdf/issues/260
	baseImage.BitsPerComponent = 8
	baseImage.hasAlpha = img.hasAlpha
	baseImage.ColorComponents = cs.Base.GetNumComponents()

	samples := img.GetSamples()
	N := cs.Base.GetNumComponents()

	if N < 1 {
		return Image{}, fmt.Errorf("bad base colorspace NumComponents=%d", N)
	}

	var baseSamples []uint32
	// Convert the indexed data to base color map data.
	for i := 0; i < len(samples); i++ {
		// Each data point represents an index location.
		// For each entry there are N values.
		index := int(samples[i])
		common.Log.Trace("Indexed: index=%d N=%d lut=%d", index, N, len(cs.colorLookup))
		// Ensure does not go out of bounds.
		if (index+1)*N > len(cs.colorLookup) {
			// Clip to the end value.
			index = len(cs.colorLookup)/N - 1
			common.Log.Trace("Clipping to index: %d", index)
			if index < 0 {
				common.Log.Debug("ERROR: Can't clip index. Is PDF file damaged?")
				break
			}
		}

		cvals := cs.colorLookup[index*N : (index+1)*N]
		common.Log.Trace("C Vals: % d", cvals)
		for _, val := range cvals {
			baseSamples = append(baseSamples, uint32(val))
		}
	}
	baseImage.SetSamples(baseSamples)
	baseImage.ColorComponents = N

	common.Log.Trace("Input samples: %d", samples)
	common.Log.Trace("-> Output samples: %d", baseSamples)

	// Convert to rgb.
	return cs.Base.ImageToRGB(baseImage)
}

// ToPdfObject converts colorspace to a PDF object. [/Indexed base hival lookup]
func (cs *PdfColorspaceSpecialIndexed) ToPdfObject() core.PdfObject {
	csObj := core.MakeArray(core.MakeName("Indexed"))
	csObj.Append(cs.Base.ToPdfObject())
	csObj.Append(core.MakeInteger(int64(cs.HiVal)))
	csObj.Append(cs.Lookup)

	if cs.container != nil {
		cs.container.PdfObject = csObj
		return cs.container
	}

	return csObj
}

// PdfColorspaceSpecialSeparation is a Separation colorspace.
// At the moment the colour space is set to a Separation space, the conforming reader shall determine whether the
// device has an available colorant (e.g. dye) corresponding to the name of the requested space. If so, the conforming
// reader shall ignore the alternateSpace and tintTransform parameters; subsequent painting operations within the
// space shall apply the designated colorant directly, according to the tint values supplied.
//
// Format: [/Separation name alternateSpace tintTransform]
type PdfColorspaceSpecialSeparation struct {
	ColorantName   *core.PdfObjectName
	AlternateSpace PdfColorspace
	TintTransform  PdfFunction

	// Container, if when parsing CS array is inside a container.
	container *core.PdfIndirectObject
}

// NewPdfColorspaceSpecialSeparation returns a new separation color.
func NewPdfColorspaceSpecialSeparation() *PdfColorspaceSpecialSeparation {
	cs := &PdfColorspaceSpecialSeparation{}
	return cs
}

func (cs *PdfColorspaceSpecialSeparation) String() string {
	return "Separation"
}

// GetNumComponents returns the number of color components (1 for Separation).
func (cs *PdfColorspaceSpecialSeparation) GetNumComponents() int {
	return 1
}

// DecodeArray returns the component range values for the Separation colorspace.
func (cs *PdfColorspaceSpecialSeparation) DecodeArray() []float64 {
	return []float64{0, 1.0}
}

// Object is an array or indirect object containing the array.
func newPdfColorspaceSpecialSeparationFromPdfObject(obj core.PdfObject) (*PdfColorspaceSpecialSeparation, error) {
	cs := NewPdfColorspaceSpecialSeparation()

	// If within an indirect object, then make a note of it.  If we write out the PdfObject later
	// we can reference the same container.  Otherwise is not within a container, but rather
	// a new array.
	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = core.TraceToDirectObject(obj)
	array, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("separation CS: Invalid object")
	}

	if array.Len() != 4 {
		return nil, fmt.Errorf("separation CS: Incorrect array length")
	}

	// Check name.
	obj = array.Get(0)
	name, ok := obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("separation CS: invalid family name")
	}
	if *name != "Separation" {
		return nil, fmt.Errorf("separation CS: wrong family name")
	}

	// Get colorant name.
	obj = array.Get(1)
	name, ok = obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("separation CS: Invalid colorant name")
	}
	cs.ColorantName = name

	// Get base colormap.
	obj = array.Get(2)
	alternativeCs, err := NewPdfColorspaceFromPdfObject(obj)
	if err != nil {
		return nil, err
	}
	cs.AlternateSpace = alternativeCs

	// Tint transform is specified by a PDF function.
	tintTransform, err := newPdfFunctionFromPdfObject(array.Get(3))
	if err != nil {
		return nil, err
	}

	cs.TintTransform = tintTransform

	return cs, nil
}

// ToPdfObject returns the PDF representation of the colorspace.
func (cs *PdfColorspaceSpecialSeparation) ToPdfObject() core.PdfObject {
	csArray := core.MakeArray(core.MakeName("Separation"))

	csArray.Append(cs.ColorantName)
	csArray.Append(cs.AlternateSpace.ToPdfObject())
	csArray.Append(cs.TintTransform.ToPdfObject())

	// If in a container, replace the contents and return back.
	// Helps not getting too many duplicates of the same objects.
	if cs.container != nil {
		cs.container.PdfObject = csArray
		return cs.container
	}

	return csArray
}

// ColorFromFloats returns a new PdfColor based on the input slice of color
// components. The slice should contain a single element.
func (cs *PdfColorspaceSpecialSeparation) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("range check")
	}

	tint := vals[0]
	input := []float64{tint}
	output, err := cs.TintTransform.Evaluate(input)
	if err != nil {
		common.Log.Debug("Error, failed to evaluate: %v", err)
		common.Log.Trace("Tint transform: %+v", cs.TintTransform)
		return nil, err
	}

	common.Log.Trace("Processing ColorFromFloats(%+v) on AlternateSpace: %#v", output, cs.AlternateSpace)
	color, err := cs.AlternateSpace.ColorFromFloats(output)
	if err != nil {
		common.Log.Debug("Error, failed to evaluate in alternate space: %v", err)
		return nil, err
	}

	return color, nil
}

// ColorFromPdfObjects returns a new PdfColor based on the input slice of color
// components. The slice should contain a single PdfObjectFloat element.
func (cs *PdfColorspaceSpecialSeparation) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts a color in Separation colorspace to RGB colorspace.
func (cs *PdfColorspaceSpecialSeparation) ColorToRGB(color PdfColor) (PdfColor, error) {
	if cs.AlternateSpace == nil {
		return nil, errors.New("alternate colorspace undefined")
	}

	return cs.AlternateSpace.ColorToRGB(color)
}

// ImageToRGB converts an image with samples in Separation CS to an image with samples specified in
// DeviceRGB CS.
func (cs *PdfColorspaceSpecialSeparation) ImageToRGB(img Image) (Image, error) {
	altImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	common.Log.Trace("Separation color space -> ToRGB conversion")
	common.Log.Trace("samples in: %d", len(samples))
	common.Log.Trace("TintTransform: %+v", cs.TintTransform)

	altDecode := cs.AlternateSpace.DecodeArray()

	var altSamples []uint32
	// Convert tints to color data in the alternate colorspace.
	for _, sample := range samples {
		// A single tint component is in the range 0.0 - 1.0
		tint := float64(sample) / maxVal

		// Convert the tint value to the alternate space value.
		outputs, err := cs.TintTransform.Evaluate([]float64{tint})
		//common.Log.Trace("%v Converting tint value: %f -> [% f]", cs.AlternateSpace, tint, outputs)

		if err != nil {
			return img, err
		}

		for i, val := range outputs {
			// Convert component value to 0-1 range.
			altVal := interpolate(val, altDecode[i*2], altDecode[i*2+1], 0, 1)

			// Rescale to [0, maxVal]
			altComponent := uint32(altVal * maxVal)

			altSamples = append(altSamples, altComponent)
		}
	}
	common.Log.Trace("Samples out: %d", len(altSamples))
	altImage.SetSamples(altSamples)
	altImage.ColorComponents = cs.AlternateSpace.GetNumComponents()

	// Set the image's decode parameters for interpretation in the alternative CS.
	altImage.decode = altDecode

	// Convert to RGB via the alternate colorspace.
	return cs.AlternateSpace.ImageToRGB(altImage)
}

// PdfColorspaceDeviceN represents a DeviceN color space. DeviceN color spaces are similar to Separation color
// spaces, except they can contain an arbitrary number of color components.
//
// Format: [/DeviceN names alternateSpace tintTransform]
//     or: [/DeviceN names alternateSpace tintTransform attributes]
type PdfColorspaceDeviceN struct {
	ColorantNames  *core.PdfObjectArray
	AlternateSpace PdfColorspace
	TintTransform  PdfFunction
	Attributes     *PdfColorspaceDeviceNAttributes

	// Optional
	container *core.PdfIndirectObject
}

// NewPdfColorspaceDeviceN returns an initialized PdfColorspaceDeviceN.
func NewPdfColorspaceDeviceN() *PdfColorspaceDeviceN {
	cs := &PdfColorspaceDeviceN{}
	return cs
}

// String returns the name of the colorspace (DeviceN).
func (cs *PdfColorspaceDeviceN) String() string {
	return "DeviceN"
}

// GetNumComponents returns the number of input color components, i.e. that are input to the tint transform.
func (cs *PdfColorspaceDeviceN) GetNumComponents() int {
	return cs.ColorantNames.Len()
}

// DecodeArray returns the component range values for the DeviceN colorspace.
// [0 1.0 0 1.0 ...] for each color component.
func (cs *PdfColorspaceDeviceN) DecodeArray() []float64 {
	var decode []float64
	for i := 0; i < cs.GetNumComponents(); i++ {
		decode = append(decode, 0.0, 1.0)
	}
	return decode
}

// newPdfColorspaceDeviceNFromPdfObject loads a DeviceN colorspace from a PdfObjectArray which can be
// contained within an indirect object.
func newPdfColorspaceDeviceNFromPdfObject(obj core.PdfObject) (*PdfColorspaceDeviceN, error) {
	cs := NewPdfColorspaceDeviceN()

	// If within an indirect object, then make a note of it.  If we write out the PdfObject later
	// we can reference the same container.  Otherwise is not within a container, but rather
	// a new array.
	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	// Check the CS array.
	obj = core.TraceToDirectObject(obj)
	csArray, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("deviceN CS: Invalid object")
	}

	if csArray.Len() != 4 && csArray.Len() != 5 {
		return nil, fmt.Errorf("deviceN CS: Incorrect array length")
	}

	// Check name.
	obj = csArray.Get(0)
	name, ok := obj.(*core.PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("deviceN CS: invalid family name")
	}
	if *name != "DeviceN" {
		return nil, fmt.Errorf("deviceN CS: wrong family name")
	}

	// Get colorant names.  Specifies the number of components too.
	obj = csArray.Get(1)
	obj = core.TraceToDirectObject(obj)
	nameArray, ok := obj.(*core.PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("deviceN CS: Invalid names array")
	}
	cs.ColorantNames = nameArray

	// Get base colormap.
	obj = csArray.Get(2)
	alternativeCs, err := NewPdfColorspaceFromPdfObject(obj)
	if err != nil {
		return nil, err
	}
	cs.AlternateSpace = alternativeCs

	// Tint transform is specified by a PDF function.
	tintTransform, err := newPdfFunctionFromPdfObject(csArray.Get(3))
	if err != nil {
		return nil, err
	}
	cs.TintTransform = tintTransform

	// Attributes.
	if csArray.Len() == 5 {
		attr, err := newPdfColorspaceDeviceNAttributesFromPdfObject(csArray.Get(4))
		if err != nil {
			return nil, err
		}
		cs.Attributes = attr
	}

	return cs, nil
}

// ToPdfObject returns a *PdfIndirectObject containing a *PdfObjectArray representation of the DeviceN colorspace.
// Format: [/DeviceN names alternateSpace tintTransform]
//     or: [/DeviceN names alternateSpace tintTransform attributes]
func (cs *PdfColorspaceDeviceN) ToPdfObject() core.PdfObject {
	csArray := core.MakeArray(core.MakeName("DeviceN"))
	csArray.Append(cs.ColorantNames)
	csArray.Append(cs.AlternateSpace.ToPdfObject())
	csArray.Append(cs.TintTransform.ToPdfObject())
	if cs.Attributes != nil {
		csArray.Append(cs.Attributes.ToPdfObject())
	}

	if cs.container != nil {
		cs.container.PdfObject = csArray
		return cs.container
	}

	return csArray
}

// ColorFromFloats returns a new PdfColor based on input color components.
func (cs *PdfColorspaceDeviceN) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != cs.GetNumComponents() {
		return nil, errors.New("range check")
	}

	output, err := cs.TintTransform.Evaluate(vals)
	if err != nil {
		return nil, err
	}

	color, err := cs.AlternateSpace.ColorFromFloats(output)
	if err != nil {
		return nil, err
	}
	return color, nil
}

// ColorFromPdfObjects returns a new PdfColor based on input color components. The input PdfObjects should
// be numeric.
func (cs *PdfColorspaceDeviceN) ColorFromPdfObjects(objects []core.PdfObject) (PdfColor, error) {
	if len(objects) != cs.GetNumComponents() {
		return nil, errors.New("range check")
	}

	floats, err := core.GetNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return cs.ColorFromFloats(floats)
}

// ColorToRGB converts a DeviceN color to an RGB color.
func (cs *PdfColorspaceDeviceN) ColorToRGB(color PdfColor) (PdfColor, error) {
	if cs.AlternateSpace == nil {
		return nil, errors.New("DeviceN alternate space undefined")
	}
	return cs.AlternateSpace.ColorToRGB(color)
}

// ImageToRGB converts an Image in a given PdfColorspace to an RGB image.
func (cs *PdfColorspaceDeviceN) ImageToRGB(img Image) (Image, error) {
	altImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	// Convert tints to color data in the alternate colorspace.
	var altSamples []uint32
	for i := 0; i < len(samples); i += cs.GetNumComponents() {
		// The input to the tint transformation is the tint
		// for each color component.
		//
		// A single tint component is in the range 0.0 - 1.0
		var inputs []float64
		for j := 0; j < cs.GetNumComponents(); j++ {
			tint := float64(samples[i+j]) / maxVal
			inputs = append(inputs, tint)
		}

		// Transform the tints to the alternate colorspace.
		// (scaled units).
		outputs, err := cs.TintTransform.Evaluate(inputs)
		if err != nil {
			return img, err
		}

		for _, val := range outputs {
			// Clip.
			val = math.Min(math.Max(0, val), 1.0)
			// Rescale to [0, maxVal]
			altComponent := uint32(val * maxVal)
			altSamples = append(altSamples, altComponent)
		}
	}
	altImage.SetSamples(altSamples)

	// Convert to RGB via the alternate colorspace.
	return cs.AlternateSpace.ImageToRGB(altImage)
}

// PdfColorspaceDeviceNAttributes contains additional information about the components of colour space that
// conforming readers may use. Conforming readers need not use the alternateSpace and tintTransform parameters,
// and may instead use a custom blending algorithms, along with other information provided in the attributes
// dictionary if present.
type PdfColorspaceDeviceNAttributes struct {
	Subtype     *core.PdfObjectName // DeviceN or NChannel (DeviceN default)
	Colorants   core.PdfObject
	Process     core.PdfObject
	MixingHints core.PdfObject

	// Optional
	container *core.PdfIndirectObject
}

// newPdfColorspaceDeviceNAttributesFromPdfObject loads a PdfColorspaceDeviceNAttributes from an input
// PdfObjectDictionary (direct/indirect).
func newPdfColorspaceDeviceNAttributesFromPdfObject(obj core.PdfObject) (*PdfColorspaceDeviceNAttributes, error) {
	attr := &PdfColorspaceDeviceNAttributes{}

	var dict *core.PdfObjectDictionary
	if indObj, isInd := obj.(*core.PdfIndirectObject); isInd {
		attr.container = indObj
		var ok bool
		dict, ok = indObj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			common.Log.Error("DeviceN attribute type error")
			return nil, errors.New("type error")
		}
	} else if d, isDict := obj.(*core.PdfObjectDictionary); isDict {
		dict = d
	} else {
		common.Log.Error("DeviceN attribute type error")
		return nil, errors.New("type error")
	}

	if obj := dict.Get("Subtype"); obj != nil {
		name, ok := core.TraceToDirectObject(obj).(*core.PdfObjectName)
		if !ok {
			common.Log.Error("DeviceN attribute Subtype type error")
			return nil, errors.New("type error")
		}

		attr.Subtype = name
	}

	if obj := dict.Get("Colorants"); obj != nil {
		attr.Colorants = obj
	}

	if obj := dict.Get("Process"); obj != nil {
		attr.Process = obj
	}

	if obj := dict.Get("MixingHints"); obj != nil {
		attr.MixingHints = obj
	}

	return attr, nil
}

// ToPdfObject returns a PdfObject representation of PdfColorspaceDeviceNAttributes as a PdfObjectDictionary directly
// or indirectly within an indirect object container.
func (cs *PdfColorspaceDeviceNAttributes) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()

	if cs.Subtype != nil {
		dict.Set("Subtype", cs.Subtype)
	}
	dict.SetIfNotNil("Colorants", cs.Colorants)
	dict.SetIfNotNil("Process", cs.Process)
	dict.SetIfNotNil("MixingHints", cs.MixingHints)

	if cs.container != nil {
		cs.container.PdfObject = dict
		return cs.container
	}

	return dict
}
