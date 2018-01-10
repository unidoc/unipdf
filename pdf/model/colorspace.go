/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"
	"math"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

//
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
//
type PdfColorspace interface {
	String() string
	ImageToRGB(Image) (Image, error)
	ColorToRGB(color PdfColor) (PdfColor, error)
	GetNumComponents() int
	ToPdfObject() PdfObject
	ColorFromPdfObjects(objects []PdfObject) (PdfColor, error)
	ColorFromFloats(vals []float64) (PdfColor, error)

	// Returns the decode array for the CS, i.e. the range of each component.
	DecodeArray() []float64
}

type PdfColor interface {
}

// NewPdfColorspaceFromPdfObject loads a PdfColorspace from a PdfObject.  Returns an error if there is
// a failure in loading.
func NewPdfColorspaceFromPdfObject(obj PdfObject) (PdfColorspace, error) {
	var container *PdfIndirectObject
	var csName *PdfObjectName
	var csArray *PdfObjectArray

	if indObj, is := obj.(*PdfIndirectObject); is {
		if array, is := indObj.PdfObject.(*PdfObjectArray); is {
			container = indObj
			csArray = array
		} else if name, is := indObj.PdfObject.(*PdfObjectName); is {
			container = indObj
			csName = name
		}
	} else if array, is := obj.(*PdfObjectArray); is {
		csArray = array
	} else if name, is := obj.(*PdfObjectName); is {
		csName = name
	}

	// If specified by a name directly: Device colorspace or Pattern.
	if csName != nil {
		if *csName == "DeviceGray" {
			cs := NewPdfColorspaceDeviceGray()
			return cs, nil
		} else if *csName == "DeviceRGB" {
			cs := NewPdfColorspaceDeviceRGB()
			return cs, nil
		} else if *csName == "DeviceCMYK" {
			cs := NewPdfColorspaceDeviceCMYK()
			return cs, nil
		} else if *csName == "Pattern" {
			cs := NewPdfColorspaceSpecialPattern()
			return cs, nil
		} else {
			common.Log.Error("Unknown colorspace %s", *csName)
			return nil, errors.New("Unknown colorspace")
		}
	}

	if csArray != nil && len(*csArray) > 0 {
		var csObject PdfObject = container
		if container == nil {
			csObject = csArray
		}
		if name, is := (*csArray)[0].(*PdfObjectName); is {
			if *name == "DeviceGray" && len(*csArray) == 1 {
				cs := NewPdfColorspaceDeviceGray()
				return cs, nil
			} else if *name == "DeviceRGB" && len(*csArray) == 1 {
				cs := NewPdfColorspaceDeviceRGB()
				return cs, nil
			} else if *name == "DeviceCMYK" && len(*csArray) == 1 {
				cs := NewPdfColorspaceDeviceCMYK()
				return cs, nil
			} else if *name == "CalGray" {
				cs, err := newPdfColorspaceCalGrayFromPdfObject(csObject)
				return cs, err
			} else if *name == "CalRGB" {
				cs, err := newPdfColorspaceCalRGBFromPdfObject(csObject)
				return cs, err
			} else if *name == "Lab" {
				cs, err := newPdfColorspaceLabFromPdfObject(csObject)
				return cs, err
			} else if *name == "ICCBased" {
				cs, err := newPdfColorspaceICCBasedFromPdfObject(csObject)
				return cs, err
			} else if *name == "Pattern" {
				cs, err := newPdfColorspaceSpecialPatternFromPdfObject(csObject)
				return cs, err
			} else if *name == "Indexed" {
				cs, err := newPdfColorspaceSpecialIndexedFromPdfObject(csObject)
				return cs, err
			} else if *name == "Separation" {
				cs, err := newPdfColorspaceSpecialSeparationFromPdfObject(csObject)
				return cs, err
			} else if *name == "DeviceN" {
				cs, err := newPdfColorspaceDeviceNFromPdfObject(csObject)
				return cs, err
			} else {
				common.Log.Debug("Array with invalid name: %s", *name)
			}
		}
	}

	common.Log.Debug("PDF File Error: Colorspace type error: %s", obj.String())
	return nil, errors.New("Type error")
}

// determine PDF colorspace from a PdfObject.  Returns the colorspace name and an error on failure.
// If the colorspace was not found, will return an empty string.
func determineColorspaceNameFromPdfObject(obj PdfObject) (PdfObjectName, error) {
	var csName *PdfObjectName
	var csArray *PdfObjectArray

	if indObj, is := obj.(*PdfIndirectObject); is {
		if array, is := indObj.PdfObject.(*PdfObjectArray); is {
			csArray = array
		} else if name, is := indObj.PdfObject.(*PdfObjectName); is {
			csName = name
		}
	} else if array, is := obj.(*PdfObjectArray); is {
		csArray = array
	} else if name, is := obj.(*PdfObjectName); is {
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

	if csArray != nil && len(*csArray) > 0 {
		if name, is := (*csArray)[0].(*PdfObjectName); is {
			switch *name {
			case "DeviceGray", "DeviceRGB", "DeviceCMYK":
				if len(*csArray) == 1 {
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

// Gray scale component.
// No specific parameters

// A grayscale value shall be represented by a single number in the range 0.0 to 1.0 where 0.0 corresponds to black
// and 1.0 to white.
type PdfColorDeviceGray float64

func NewPdfColorDeviceGray(grayVal float64) *PdfColorDeviceGray {
	color := PdfColorDeviceGray(grayVal)
	return &color
}

func (this *PdfColorDeviceGray) GetNumComponents() int {
	return 1
}

func (this *PdfColorDeviceGray) Val() float64 {
	return float64(*this)
}

// Convert to an integer format.
func (this *PdfColorDeviceGray) ToInteger(bits int) uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return uint32(maxVal * this.Val())
}

type PdfColorspaceDeviceGray struct{}

func NewPdfColorspaceDeviceGray() *PdfColorspaceDeviceGray {
	return &PdfColorspaceDeviceGray{}
}

func (this *PdfColorspaceDeviceGray) GetNumComponents() int {
	return 1
}

// DecodeArray returns the range of color component values in DeviceGray colorspace.
func (this *PdfColorspaceDeviceGray) DecodeArray() []float64 {
	return []float64{0, 1.0}
}

func (this *PdfColorspaceDeviceGray) ToPdfObject() PdfObject {
	return MakeName("DeviceGray")
}

func (this *PdfColorspaceDeviceGray) String() string {
	return "DeviceGray"
}

func (this *PdfColorspaceDeviceGray) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("Range check")
	}

	val := vals[0]

	if val < 0.0 || val > 1.0 {
		return nil, errors.New("Range check")
	}

	return NewPdfColorDeviceGray(val), nil
}

func (this *PdfColorspaceDeviceGray) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

// Convert gray -> rgb for a single color component.
func (this *PdfColorspaceDeviceGray) ColorToRGB(color PdfColor) (PdfColor, error) {
	gray, ok := color.(*PdfColorDeviceGray)
	if !ok {
		common.Log.Debug("Input color not device gray %T", color)
		return nil, errors.New("Type check error")
	}

	return NewPdfColorDeviceRGB(float64(*gray), float64(*gray), float64(*gray)), nil
}

// Convert 1-component grayscale data to 3-component RGB.
func (this *PdfColorspaceDeviceGray) ImageToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()
	common.Log.Trace("DeviceGray-ToRGB Samples: % d", samples)

	rgbSamples := []uint32{}
	for i := 0; i < len(samples); i++ {
		grayVal := samples[i]
		rgbSamples = append(rgbSamples, grayVal, grayVal, grayVal)
	}
	rgbImage.BitsPerComponent = 8
	rgbImage.ColorComponents = 3
	rgbImage.SetSamples(rgbSamples)

	common.Log.Trace("DeviceGray -> RGB")
	common.Log.Trace("samples: %v", samples)
	common.Log.Trace("RGB samples: %v", rgbSamples)
	common.Log.Trace("%v -> %v", img, rgbImage)

	return rgbImage, nil
}

//////////////////////
// Device RGB
// R, G, B components.
// No specific parameters

// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.
type PdfColorDeviceRGB [3]float64

func NewPdfColorDeviceRGB(r, g, b float64) *PdfColorDeviceRGB {
	color := PdfColorDeviceRGB{r, g, b}
	return &color
}

func (this *PdfColorDeviceRGB) GetNumComponents() int {
	return 3
}

func (this *PdfColorDeviceRGB) R() float64 {
	return float64(this[0])
}

func (this *PdfColorDeviceRGB) G() float64 {
	return float64(this[1])
}

func (this *PdfColorDeviceRGB) B() float64 {
	return float64(this[2])
}

// Convert to an integer format.
func (this *PdfColorDeviceRGB) ToInteger(bits int) [3]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [3]uint32{uint32(maxVal * this.R()), uint32(maxVal * this.G()), uint32(maxVal * this.B())}
}

func (this *PdfColorDeviceRGB) ToGray() *PdfColorDeviceGray {
	// Calculate grayValue [0-1]
	grayValue := 0.3*this.R() + 0.59*this.G() + 0.11*this.B()

	// Clip to [0-1]
	grayValue = math.Min(math.Max(grayValue, 0.0), 1.0)

	return NewPdfColorDeviceGray(grayValue)
}

// RGB colorspace.

type PdfColorspaceDeviceRGB struct{}

func NewPdfColorspaceDeviceRGB() *PdfColorspaceDeviceRGB {
	return &PdfColorspaceDeviceRGB{}
}

func (this *PdfColorspaceDeviceRGB) String() string {
	return "DeviceRGB"
}

func (this *PdfColorspaceDeviceRGB) GetNumComponents() int {
	return 3
}

// DecodeArray returns the range of color component values in DeviceRGB colorspace.
func (this *PdfColorspaceDeviceRGB) DecodeArray() []float64 {
	return []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
}

func (this *PdfColorspaceDeviceRGB) ToPdfObject() PdfObject {
	return MakeName("DeviceRGB")
}

func (this *PdfColorspaceDeviceRGB) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 3 {
		return nil, errors.New("Range check")
	}

	// Red.
	r := vals[0]
	if r < 0.0 || r > 1.0 {
		return nil, errors.New("Range check")
	}

	// Green.
	g := vals[1]
	if g < 0.0 || g > 1.0 {
		return nil, errors.New("Range check")
	}

	// Blue.
	b := vals[2]
	if b < 0.0 || b > 1.0 {
		return nil, errors.New("Range check")
	}

	color := NewPdfColorDeviceRGB(r, g, b)
	return color, nil

}

// Get the color from a series of pdf objects (3 for rgb).
func (this *PdfColorspaceDeviceRGB) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 3 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceDeviceRGB) ColorToRGB(color PdfColor) (PdfColor, error) {
	rgb, ok := color.(*PdfColorDeviceRGB)
	if !ok {
		common.Log.Debug("Input color not device RGB")
		return nil, errors.New("Type check error")
	}
	return rgb, nil
}

func (this *PdfColorspaceDeviceRGB) ImageToRGB(img Image) (Image, error) {
	return img, nil
}

func (this *PdfColorspaceDeviceRGB) ImageToGray(img Image) (Image, error) {
	grayImage := img

	samples := img.GetSamples()

	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1
	graySamples := []uint32{}
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

// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.
type PdfColorDeviceCMYK [4]float64

func NewPdfColorDeviceCMYK(c, m, y, k float64) *PdfColorDeviceCMYK {
	color := PdfColorDeviceCMYK{c, m, y, k}
	return &color
}

func (this *PdfColorDeviceCMYK) GetNumComponents() int {
	return 4
}

func (this *PdfColorDeviceCMYK) C() float64 {
	return float64(this[0])
}

func (this *PdfColorDeviceCMYK) M() float64 {
	return float64(this[1])
}

func (this *PdfColorDeviceCMYK) Y() float64 {
	return float64(this[2])
}

func (this *PdfColorDeviceCMYK) K() float64 {
	return float64(this[3])
}

// Convert to an integer format.
func (this *PdfColorDeviceCMYK) ToInteger(bits int) [4]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [4]uint32{uint32(maxVal * this.C()), uint32(maxVal * this.M()), uint32(maxVal * this.Y()), uint32(maxVal * this.K())}
}

type PdfColorspaceDeviceCMYK struct{}

func NewPdfColorspaceDeviceCMYK() *PdfColorspaceDeviceCMYK {
	return &PdfColorspaceDeviceCMYK{}
}

func (this *PdfColorspaceDeviceCMYK) String() string {
	return "DeviceCMYK"
}

func (this *PdfColorspaceDeviceCMYK) GetNumComponents() int {
	return 4
}

// DecodeArray returns the range of color component values in DeviceCMYK colorspace.
func (this *PdfColorspaceDeviceCMYK) DecodeArray() []float64 {
	return []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
}

func (this *PdfColorspaceDeviceCMYK) ToPdfObject() PdfObject {
	return MakeName("DeviceCMYK")
}

func (this *PdfColorspaceDeviceCMYK) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 4 {
		return nil, errors.New("Range check")
	}

	// Cyan
	c := vals[0]
	if c < 0.0 || c > 1.0 {
		return nil, errors.New("Range check")
	}

	// Magenta
	m := vals[1]
	if m < 0.0 || m > 1.0 {
		return nil, errors.New("Range check")
	}

	// Yellow.
	y := vals[2]
	if y < 0.0 || y > 1.0 {
		return nil, errors.New("Range check")
	}

	// Key.
	k := vals[3]
	if k < 0.0 || k > 1.0 {
		return nil, errors.New("Range check")
	}

	color := NewPdfColorDeviceCMYK(c, m, y, k)
	return color, nil
}

// Get the color from a series of pdf objects (4 for cmyk).
func (this *PdfColorspaceDeviceCMYK) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 4 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceDeviceCMYK) ColorToRGB(color PdfColor) (PdfColor, error) {
	cmyk, ok := color.(*PdfColorDeviceCMYK)
	if !ok {
		common.Log.Debug("Input color not device cmyk")
		return nil, errors.New("Type check error")
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

func (this *PdfColorspaceDeviceCMYK) ImageToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()

	common.Log.Trace("CMYK -> RGB")
	common.Log.Trace("image bpc: %d, color comps: %d", img.BitsPerComponent, img.ColorComponents)
	common.Log.Trace("Len data: %d, len samples: %d", len(img.Data), len(samples))
	common.Log.Trace("Height: %d, Width: %d", img.Height, img.Width)
	if len(samples)%4 != 0 {
		//common.Log.Debug("samples: % d", samples)
		common.Log.Debug("Input image: %#v", img)
		common.Log.Debug("CMYK -> RGB fail, len samples: %d", len(samples))
		return img, errors.New("CMYK data not a multiple of 4")
	}

	decode := img.decode
	if decode == nil {
		decode = []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
	}
	if len(decode) != 8 {
		common.Log.Debug("Invalid decode array (%d): % .3f", len(decode), decode)
		return img, errors.New("Invalid decode array")
	}
	common.Log.Trace("Decode array: % f", decode)

	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1
	common.Log.Trace("MaxVal: %f", maxVal)
	rgbSamples := []uint32{}
	for i := 0; i < len(samples); i += 4 {
		// Normalized c, m, y, k values.
		c := interpolate(float64(samples[i]), 0, maxVal, decode[0], decode[1])
		m := interpolate(float64(samples[i+1]), 0, maxVal, decode[2], decode[3])
		y := interpolate(float64(samples[i+2]), 0, maxVal, decode[4], decode[5])
		k := interpolate(float64(samples[i+3]), 0, maxVal, decode[6], decode[7])

		c = c*(1-k) + k
		m = m*(1-k) + k
		y = y*(1-k) + k

		r := 1 - c
		g := 1 - m
		b := 1 - y

		// Convert to uint32 format.
		R := uint32(r * maxVal)
		G := uint32(g * maxVal)
		B := uint32(b * maxVal)
		//common.Log.Trace("(%f,%f,%f,%f) -> (%f,%f,%f) [%d,%d,%d]", c, m, y, k, r, g, b, R, G, B)

		rgbSamples = append(rgbSamples, R, G, B)
	}
	rgbImage.SetSamples(rgbSamples)
	rgbImage.ColorComponents = 3

	return rgbImage, nil
}

//////////////////////
// CIE based gray level.
// Single component
// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.

type PdfColorCalGray float64

func NewPdfColorCalGray(grayVal float64) *PdfColorCalGray {
	color := PdfColorCalGray(grayVal)
	return &color
}

func (this *PdfColorCalGray) GetNumComponents() int {
	return 1
}

func (this *PdfColorCalGray) Val() float64 {
	return float64(*this)
}

// Convert to an integer format.
func (this *PdfColorCalGray) ToInteger(bits int) uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return uint32(maxVal * this.Val())
}

// CalGray color space.
type PdfColorspaceCalGray struct {
	WhitePoint []float64 // [XW, YW, ZW]: Required
	BlackPoint []float64 // [XB, YB, ZB]
	Gamma      float64

	container *PdfIndirectObject
}

func NewPdfColorspaceCalGray() *PdfColorspaceCalGray {
	cs := &PdfColorspaceCalGray{}

	// Set optional parameters to default values.
	cs.BlackPoint = []float64{0.0, 0.0, 0.0}
	cs.Gamma = 1

	return cs
}

func (this *PdfColorspaceCalGray) String() string {
	return "CalGray"
}

func (this *PdfColorspaceCalGray) GetNumComponents() int {
	return 1
}

// DecodeArray returns the range of color component values in CalGray colorspace.
func (this *PdfColorspaceCalGray) DecodeArray() []float64 {
	return []float64{0.0, 1.0}
}

func newPdfColorspaceCalGrayFromPdfObject(obj PdfObject) (*PdfColorspaceCalGray, error) {
	cs := NewPdfColorspaceCalGray()

	// If within an indirect object, then make a note of it.  If we write out the PdfObject later
	// we can reference the same container.  Otherwise is not within a container, but rather
	// a new array.
	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = TraceToDirectObject(obj)
	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Type error")
	}

	if len(*array) != 2 {
		return nil, fmt.Errorf("Invalid CalGray colorspace")
	}

	// Name.
	obj = TraceToDirectObject((*array)[0])
	name, ok := obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("CalGray name not a Name object")
	}
	if *name != "CalGray" {
		return nil, fmt.Errorf("Not a CalGray colorspace")
	}

	// Dict.
	obj = TraceToDirectObject((*array)[1])
	dict, ok := obj.(*PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("CalGray dict not a Dictionary object")
	}

	// WhitePoint (Required): [Xw, Yw, Zw]
	obj = dict.Get("WhitePoint")
	obj = TraceToDirectObject(obj)
	whitePointArray, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("CalGray: Invalid WhitePoint")
	}
	if len(*whitePointArray) != 3 {
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
		obj = TraceToDirectObject(obj)
		blackPointArray, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalGray: Invalid BlackPoint")
		}
		if len(*blackPointArray) != 3 {
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
		obj = TraceToDirectObject(obj)
		gamma, err := getNumberAsFloat(obj)
		if err != nil {
			return nil, fmt.Errorf("CalGray: gamma not a number")
		}
		cs.Gamma = gamma
	}

	return cs, nil
}

// Return as PDF object format [name dictionary]
func (this *PdfColorspaceCalGray) ToPdfObject() PdfObject {
	// CalGray color space dictionary..
	cspace := &PdfObjectArray{}

	cspace.Append(MakeName("CalGray"))

	dict := MakeDict()
	if this.WhitePoint != nil {
		dict.Set("WhitePoint", MakeArray(MakeFloat(this.WhitePoint[0]), MakeFloat(this.WhitePoint[1]), MakeFloat(this.WhitePoint[2])))
	} else {
		common.Log.Error("CalGray: Missing WhitePoint (Required)")
	}

	if this.BlackPoint != nil {
		dict.Set("BlackPoint", MakeArray(MakeFloat(this.BlackPoint[0]), MakeFloat(this.BlackPoint[1]), MakeFloat(this.BlackPoint[2])))
	}

	dict.Set("Gamma", MakeFloat(this.Gamma))
	cspace.Append(dict)

	if this.container != nil {
		this.container.PdfObject = cspace
		return this.container
	}

	return cspace
}

func (this *PdfColorspaceCalGray) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("Range check")
	}

	val := vals[0]
	if val < 0.0 || val > 1.0 {
		return nil, errors.New("Range check")
	}

	color := NewPdfColorCalGray(val)
	return color, nil
}

func (this *PdfColorspaceCalGray) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceCalGray) ColorToRGB(color PdfColor) (PdfColor, error) {
	calgray, ok := color.(*PdfColorCalGray)
	if !ok {
		common.Log.Debug("Input color not cal gray")
		return nil, errors.New("Type check error")
	}

	ANorm := calgray.Val()

	// A -> X,Y,Z
	X := this.WhitePoint[0] * math.Pow(ANorm, this.Gamma)
	Y := this.WhitePoint[1] * math.Pow(ANorm, this.Gamma)
	Z := this.WhitePoint[2] * math.Pow(ANorm, this.Gamma)

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

// A, B, C -> X, Y, Z
func (this *PdfColorspaceCalGray) ImageToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	rgbSamples := []uint32{}
	for i := 0; i < len(samples); i++ {
		// A represents the gray component of calibrated gray space.
		// It shall be in the range 0.0 - 1.0
		ANorm := float64(samples[i]) / maxVal

		// A -> X,Y,Z
		X := this.WhitePoint[0] * math.Pow(ANorm, this.Gamma)
		Y := this.WhitePoint[1] * math.Pow(ANorm, this.Gamma)
		Z := this.WhitePoint[2] * math.Pow(ANorm, this.Gamma)

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

//////////////////////
// Colorimetric CIE RGB colorspace.
// A, B, C components
// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.

type PdfColorCalRGB [3]float64

func NewPdfColorCalRGB(a, b, c float64) *PdfColorCalRGB {
	color := PdfColorCalRGB{a, b, c}
	return &color
}

func (this *PdfColorCalRGB) GetNumComponents() int {
	return 3
}

func (this *PdfColorCalRGB) A() float64 {
	return float64(this[0])
}

func (this *PdfColorCalRGB) B() float64 {
	return float64(this[1])
}

func (this *PdfColorCalRGB) C() float64 {
	return float64(this[2])
}

// Convert to an integer format.
func (this *PdfColorCalRGB) ToInteger(bits int) [3]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [3]uint32{uint32(maxVal * this.A()), uint32(maxVal * this.B()), uint32(maxVal * this.C())}
}

// A, B, C components
type PdfColorspaceCalRGB struct {
	WhitePoint []float64
	BlackPoint []float64
	Gamma      []float64
	Matrix     []float64 // [XA YA ZA XB YB ZB XC YC ZC] ; default value identity [1 0 0 0 1 0 0 0 1]
	dict       *PdfObjectDictionary

	container *PdfIndirectObject
}

// require parameters?
func NewPdfColorspaceCalRGB() *PdfColorspaceCalRGB {
	cs := &PdfColorspaceCalRGB{}

	// Set optional parameters to default values.
	cs.BlackPoint = []float64{0.0, 0.0, 0.0}
	cs.Gamma = []float64{1.0, 1.0, 1.0}
	cs.Matrix = []float64{1, 0, 0, 0, 1, 0, 0, 0, 1} // Identity matrix.

	return cs
}

func (this *PdfColorspaceCalRGB) String() string {
	return "CalRGB"
}

func (this *PdfColorspaceCalRGB) GetNumComponents() int {
	return 3
}

// DecodeArray returns the range of color component values in CalRGB colorspace.
func (this *PdfColorspaceCalRGB) DecodeArray() []float64 {
	return []float64{0.0, 1.0, 0.0, 1.0, 0.0, 1.0}
}

func newPdfColorspaceCalRGBFromPdfObject(obj PdfObject) (*PdfColorspaceCalRGB, error) {
	cs := NewPdfColorspaceCalRGB()

	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = TraceToDirectObject(obj)
	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Type error")
	}

	if len(*array) != 2 {
		return nil, fmt.Errorf("Invalid CalRGB colorspace")
	}

	// Name.
	obj = TraceToDirectObject((*array)[0])
	name, ok := obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("CalRGB name not a Name object")
	}
	if *name != "CalRGB" {
		return nil, fmt.Errorf("Not a CalRGB colorspace")
	}

	// Dict.
	obj = TraceToDirectObject((*array)[1])
	dict, ok := obj.(*PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("CalRGB name not a Name object")
	}

	// WhitePoint (Required): [Xw, Yw, Zw]
	obj = dict.Get("WhitePoint")
	obj = TraceToDirectObject(obj)
	whitePointArray, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("CalRGB: Invalid WhitePoint")
	}
	if len(*whitePointArray) != 3 {
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
		obj = TraceToDirectObject(obj)
		blackPointArray, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalRGB: Invalid BlackPoint")
		}
		if len(*blackPointArray) != 3 {
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
		obj = TraceToDirectObject(obj)
		gammaArray, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalRGB: Invalid Gamma")
		}
		if len(*gammaArray) != 3 {
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
		obj = TraceToDirectObject(obj)
		matrixArray, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("CalRGB: Invalid Matrix")
		}
		if len(*matrixArray) != 9 {
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

// Return as PDF object format [name dictionary]
func (this *PdfColorspaceCalRGB) ToPdfObject() PdfObject {
	// CalRGB color space dictionary..
	cspace := &PdfObjectArray{}

	cspace.Append(MakeName("CalRGB"))

	dict := MakeDict()
	if this.WhitePoint != nil {
		wp := MakeArray(MakeFloat(this.WhitePoint[0]), MakeFloat(this.WhitePoint[1]), MakeFloat(this.WhitePoint[2]))
		dict.Set("WhitePoint", wp)
	} else {
		common.Log.Error("CalRGB: Missing WhitePoint (Required)")
	}

	if this.BlackPoint != nil {
		bp := MakeArray(MakeFloat(this.BlackPoint[0]), MakeFloat(this.BlackPoint[1]), MakeFloat(this.BlackPoint[2]))
		dict.Set("BlackPoint", bp)
	}
	if this.Gamma != nil {
		g := MakeArray(MakeFloat(this.Gamma[0]), MakeFloat(this.Gamma[1]), MakeFloat(this.Gamma[2]))
		dict.Set("Gamma", g)
	}
	if this.Matrix != nil {
		matrix := MakeArray(MakeFloat(this.Matrix[0]), MakeFloat(this.Matrix[1]), MakeFloat(this.Matrix[2]),
			MakeFloat(this.Matrix[3]), MakeFloat(this.Matrix[4]), MakeFloat(this.Matrix[5]),
			MakeFloat(this.Matrix[6]), MakeFloat(this.Matrix[7]), MakeFloat(this.Matrix[8]))
		dict.Set("Matrix", matrix)
	}
	cspace.Append(dict)

	if this.container != nil {
		this.container.PdfObject = cspace
		return this.container
	}

	return cspace
}

func (this *PdfColorspaceCalRGB) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 3 {
		return nil, errors.New("Range check")
	}

	// A
	a := vals[0]
	if a < 0.0 || a > 1.0 {
		return nil, errors.New("Range check")
	}

	// B
	b := vals[1]
	if b < 0.0 || b > 1.0 {
		return nil, errors.New("Range check")
	}

	// C.
	c := vals[2]
	if c < 0.0 || c > 1.0 {
		return nil, errors.New("Range check")
	}

	color := NewPdfColorCalRGB(a, b, c)
	return color, nil
}

func (this *PdfColorspaceCalRGB) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 3 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceCalRGB) ColorToRGB(color PdfColor) (PdfColor, error) {
	calrgb, ok := color.(*PdfColorCalRGB)
	if !ok {
		common.Log.Debug("Input color not cal rgb")
		return nil, errors.New("Type check error")
	}

	// A, B, C in range 0.0 to 1.0
	aVal := calrgb.A()
	bVal := calrgb.B()
	cVal := calrgb.C()

	// A, B, C -> X,Y,Z
	// Gamma [GR GC GB]
	// Matrix [XA YA ZA XB YB ZB XC YC ZC]
	X := this.Matrix[0]*math.Pow(aVal, this.Gamma[0]) + this.Matrix[3]*math.Pow(bVal, this.Gamma[1]) + this.Matrix[6]*math.Pow(cVal, this.Gamma[2])
	Y := this.Matrix[1]*math.Pow(aVal, this.Gamma[0]) + this.Matrix[4]*math.Pow(bVal, this.Gamma[1]) + this.Matrix[7]*math.Pow(cVal, this.Gamma[2])
	Z := this.Matrix[2]*math.Pow(aVal, this.Gamma[0]) + this.Matrix[5]*math.Pow(bVal, this.Gamma[1]) + this.Matrix[8]*math.Pow(cVal, this.Gamma[2])

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

func (this *PdfColorspaceCalRGB) ImageToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	rgbSamples := []uint32{}
	for i := 0; i < len(samples)-2; i++ {
		// A, B, C in range 0.0 to 1.0
		aVal := float64(samples[i]) / maxVal
		bVal := float64(samples[i+1]) / maxVal
		cVal := float64(samples[i+2]) / maxVal

		// A, B, C -> X,Y,Z
		// Gamma [GR GC GB]
		// Matrix [XA YA ZA XB YB ZB XC YC ZC]
		X := this.Matrix[0]*math.Pow(aVal, this.Gamma[0]) + this.Matrix[3]*math.Pow(bVal, this.Gamma[1]) + this.Matrix[6]*math.Pow(cVal, this.Gamma[2])
		Y := this.Matrix[1]*math.Pow(aVal, this.Gamma[0]) + this.Matrix[4]*math.Pow(bVal, this.Gamma[1]) + this.Matrix[7]*math.Pow(cVal, this.Gamma[2])
		Z := this.Matrix[2]*math.Pow(aVal, this.Gamma[0]) + this.Matrix[5]*math.Pow(bVal, this.Gamma[1]) + this.Matrix[8]*math.Pow(cVal, this.Gamma[2])

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
// L*, a*, b* 3 component colorspace.
// Each component is defined in the range 0.0 - 1.0 where 1.0 is the primary intensity.

type PdfColorLab [3]float64

func NewPdfColorLab(l, a, b float64) *PdfColorLab {
	color := PdfColorLab{l, a, b}
	return &color
}

func (this *PdfColorLab) GetNumComponents() int {
	return 3
}

func (this *PdfColorLab) L() float64 {
	return float64(this[0])
}

func (this *PdfColorLab) A() float64 {
	return float64(this[1])
}

func (this *PdfColorLab) B() float64 {
	return float64(this[2])
}

// Convert to an integer format.
func (this *PdfColorLab) ToInteger(bits int) [3]uint32 {
	maxVal := math.Pow(2, float64(bits)) - 1
	return [3]uint32{uint32(maxVal * this.L()), uint32(maxVal * this.A()), uint32(maxVal * this.B())}
}

// L*, a*, b* 3 component colorspace.
type PdfColorspaceLab struct {
	WhitePoint []float64 // Required.
	BlackPoint []float64
	Range      []float64 // [amin amax bmin bmax]

	container *PdfIndirectObject
}

func (this *PdfColorspaceLab) String() string {
	return "Lab"
}

func (this *PdfColorspaceLab) GetNumComponents() int {
	return 3
}

// DecodeArray returns the range of color component values in the Lab colorspace.
func (this *PdfColorspaceLab) DecodeArray() []float64 {
	// Range for L
	decode := []float64{0, 100}

	// Range for A,B specified by range or default
	if this.Range != nil && len(this.Range) == 4 {
		decode = append(decode, this.Range...)
	} else {
		decode = append(decode, -100, 100, -100, 100)
	}

	return decode
}

// require parameters?
func NewPdfColorspaceLab() *PdfColorspaceLab {
	cs := &PdfColorspaceLab{}

	// Set optional parameters to default values.
	cs.BlackPoint = []float64{0.0, 0.0, 0.0}
	cs.Range = []float64{-100, 100, -100, 100} // Identity matrix.

	return cs
}

func newPdfColorspaceLabFromPdfObject(obj PdfObject) (*PdfColorspaceLab, error) {
	cs := NewPdfColorspaceLab()

	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = TraceToDirectObject(obj)
	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Type error")
	}

	if len(*array) != 2 {
		return nil, fmt.Errorf("Invalid CalRGB colorspace")
	}

	// Name.
	obj = TraceToDirectObject((*array)[0])
	name, ok := obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("Lab name not a Name object")
	}
	if *name != "Lab" {
		return nil, fmt.Errorf("Not a Lab colorspace")
	}

	// Dict.
	obj = TraceToDirectObject((*array)[1])
	dict, ok := obj.(*PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("Colorspace dictionary missing or invalid")
	}

	// WhitePoint (Required): [Xw, Yw, Zw]
	obj = dict.Get("WhitePoint")
	obj = TraceToDirectObject(obj)
	whitePointArray, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Lab Invalid WhitePoint")
	}
	if len(*whitePointArray) != 3 {
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
		obj = TraceToDirectObject(obj)
		blackPointArray, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Lab: Invalid BlackPoint")
		}
		if len(*blackPointArray) != 3 {
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
		obj = TraceToDirectObject(obj)
		rangeArray, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Error("Range type error")
			return nil, fmt.Errorf("Lab: Type error")
		}
		if len(*rangeArray) != 4 {
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

// Return as PDF object format [name dictionary]
func (this *PdfColorspaceLab) ToPdfObject() PdfObject {
	// CalRGB color space dictionary..
	csObj := &PdfObjectArray{}

	csObj.Append(MakeName("Lab"))

	dict := MakeDict()
	if this.WhitePoint != nil {
		wp := MakeArray(MakeFloat(this.WhitePoint[0]), MakeFloat(this.WhitePoint[1]), MakeFloat(this.WhitePoint[2]))
		dict.Set("WhitePoint", wp)
	} else {
		common.Log.Error("Lab: Missing WhitePoint (Required)")
	}

	if this.BlackPoint != nil {
		bp := MakeArray(MakeFloat(this.BlackPoint[0]), MakeFloat(this.BlackPoint[1]), MakeFloat(this.BlackPoint[2]))
		dict.Set("BlackPoint", bp)
	}

	if this.Range != nil {
		val := MakeArray(MakeFloat(this.Range[0]), MakeFloat(this.Range[1]), MakeFloat(this.Range[2]), MakeFloat(this.Range[3]))
		dict.Set("Range", val)
	}
	csObj.Append(dict)

	if this.container != nil {
		this.container.PdfObject = csObj
		return this.container
	}

	return csObj
}

func (this *PdfColorspaceLab) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 3 {
		return nil, errors.New("Range check")
	}

	// L
	l := vals[0]
	if l < 0.0 || l > 100.0 {
		common.Log.Debug("L out of range (got %v should be 0-100)", l)
		return nil, errors.New("Range check")
	}

	// A
	a := vals[1]
	aMin := float64(-100)
	aMax := float64(100)
	if len(this.Range) > 1 {
		aMin = this.Range[0]
		aMax = this.Range[1]
	}
	if a < aMin || a > aMax {
		common.Log.Debug("A out of range (got %v; range %v to %v)", a, aMin, aMax)
		return nil, errors.New("Range check")
	}

	// B.
	b := vals[2]
	bMin := float64(-100)
	bMax := float64(100)
	if len(this.Range) > 3 {
		bMin = this.Range[2]
		bMax = this.Range[3]
	}
	if b < bMin || b > bMax {
		common.Log.Debug("b out of range (got %v; range %v to %v)", b, bMin, bMax)
		return nil, errors.New("Range check")
	}

	color := NewPdfColorLab(l, a, b)
	return color, nil
}

func (this *PdfColorspaceLab) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 3 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceLab) ColorToRGB(color PdfColor) (PdfColor, error) {
	gFunc := func(x float64) float64 {
		if x >= 6.0/29 {
			return x * x * x
		} else {
			return 108.0 / 841 * (x - 4/29)
		}
	}

	lab, ok := color.(*PdfColorLab)
	if !ok {
		common.Log.Debug("input color not lab")
		return nil, errors.New("Type check error")
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
	X := this.WhitePoint[0] * gFunc(L)
	Y := this.WhitePoint[1] * gFunc(M)
	Z := this.WhitePoint[2] * gFunc(N)

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

func (this *PdfColorspaceLab) ImageToRGB(img Image) (Image, error) {
	g := func(x float64) float64 {
		if x >= 6.0/29 {
			return x * x * x
		} else {
			return 108.0 / 841 * (x - 4/29)
		}
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
		componentRanges = this.DecodeArray()
	}

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	rgbSamples := []uint32{}
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
		X := this.WhitePoint[0] * g(L)
		Y := this.WhitePoint[1] * g(M)
		Z := this.WhitePoint[2] * g(N)

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

// format [/ICCBased stream]
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
	Range    []float64        // Array of 2xN numbers, specifying range of each color component.
	Metadata *PdfObjectStream // Metadata stream.
	Data     []byte           // ICC colormap data.

	container *PdfIndirectObject
	stream    *PdfObjectStream
}

func (this *PdfColorspaceICCBased) GetNumComponents() int {
	return this.N
}

// DecodeArray returns the range of color component values in the ICCBased colorspace.
func (this *PdfColorspaceICCBased) DecodeArray() []float64 {
	return this.Range
}

func (this *PdfColorspaceICCBased) String() string {
	return "ICCBased"
}

func NewPdfColorspaceICCBased(N int) (*PdfColorspaceICCBased, error) {
	cs := &PdfColorspaceICCBased{}

	if N != 1 && N != 3 && N != 4 {
		return nil, fmt.Errorf("Invalid N (1/3/4)")
	}

	cs.N = N

	return cs, nil
}

// Input format [/ICCBased stream]
func newPdfColorspaceICCBasedFromPdfObject(obj PdfObject) (*PdfColorspaceICCBased, error) {
	cs := &PdfColorspaceICCBased{}
	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = TraceToDirectObject(obj)
	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Type error")
	}

	if len(*array) != 2 {
		return nil, fmt.Errorf("Invalid ICCBased colorspace")
	}

	// Name.
	obj = TraceToDirectObject((*array)[0])
	name, ok := obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("ICCBased name not a Name object")
	}
	if *name != "ICCBased" {
		return nil, fmt.Errorf("Not an ICCBased colorspace")
	}

	// Stream
	obj = (*array)[1]
	stream, ok := obj.(*PdfObjectStream)
	if !ok {
		common.Log.Error("ICCBased not pointing to stream: %T", obj)
		return nil, fmt.Errorf("ICCBased stream invalid")
	}

	dict := stream.PdfObjectDictionary

	n, ok := dict.Get("N").(*PdfObjectInteger)
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
		obj = TraceToDirectObject(obj)
		array, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("ICCBased Range not an array")
		}
		if len(*array) != 2*cs.N {
			return nil, fmt.Errorf("ICCBased Range wrong number of elements")
		}
		r, err := array.GetAsFloat64Slice()
		if err != nil {
			return nil, err
		}
		cs.Range = r
	}

	if obj := dict.Get("Metadata"); obj != nil {
		stream, ok := obj.(*PdfObjectStream)
		if !ok {
			return nil, fmt.Errorf("ICCBased Metadata not a stream")
		}
		cs.Metadata = stream
	}

	data, err := DecodeStream(stream)
	if err != nil {
		return nil, err
	}
	cs.Data = data
	cs.stream = stream

	return cs, nil
}

// Return as PDF object format [name stream]
func (this *PdfColorspaceICCBased) ToPdfObject() PdfObject {
	csObj := &PdfObjectArray{}

	csObj.Append(MakeName("ICCBased"))

	var stream *PdfObjectStream
	if this.stream != nil {
		stream = this.stream
	} else {
		stream = &PdfObjectStream{}
	}
	dict := MakeDict()

	dict.Set("N", MakeInteger(int64(this.N)))

	if this.Alternate != nil {
		dict.Set("Alternate", this.Alternate.ToPdfObject())
	}

	if this.Metadata != nil {
		dict.Set("Metadata", this.Metadata)
	}
	if this.Range != nil {
		ranges := []PdfObject{}
		for _, r := range this.Range {
			ranges = append(ranges, MakeFloat(r))
		}
		dict.Set("Range", MakeArray(ranges...))
	}

	// Encode with a default encoder?
	dict.Set("Length", MakeInteger(int64(len(this.Data))))
	// Need to have a representation of the stream...
	stream.Stream = this.Data
	stream.PdfObjectDictionary = dict

	csObj.Append(stream)

	if this.container != nil {
		this.container.PdfObject = csObj
		return this.container
	}

	return csObj
}

func (this *PdfColorspaceICCBased) ColorFromFloats(vals []float64) (PdfColor, error) {
	if this.Alternate == nil {
		if this.N == 1 {
			cs := NewPdfColorspaceDeviceGray()
			return cs.ColorFromFloats(vals)
		} else if this.N == 3 {
			cs := NewPdfColorspaceDeviceRGB()
			return cs.ColorFromFloats(vals)
		} else if this.N == 4 {
			cs := NewPdfColorspaceDeviceCMYK()
			return cs.ColorFromFloats(vals)
		} else {
			return nil, errors.New("ICC Based colorspace missing alternative")
		}
	}

	return this.Alternate.ColorFromFloats(vals)
}

func (this *PdfColorspaceICCBased) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if this.Alternate == nil {
		if this.N == 1 {
			cs := NewPdfColorspaceDeviceGray()
			return cs.ColorFromPdfObjects(objects)
		} else if this.N == 3 {
			cs := NewPdfColorspaceDeviceRGB()
			return cs.ColorFromPdfObjects(objects)
		} else if this.N == 4 {
			cs := NewPdfColorspaceDeviceCMYK()
			return cs.ColorFromPdfObjects(objects)
		} else {
			return nil, errors.New("ICC Based colorspace missing alternative")
		}
	}

	return this.Alternate.ColorFromPdfObjects(objects)
}

func (this *PdfColorspaceICCBased) ColorToRGB(color PdfColor) (PdfColor, error) {
	/*
		_, ok := color.(*PdfColorICCBased)
		if !ok {
			common.Log.Debug("ICC Based color error, type: %T", color)
			return nil, errors.New("Type check error")
		}
	*/

	if this.Alternate == nil {
		common.Log.Debug("ICC Based colorspace missing alternative")
		if this.N == 1 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceGray (N=1)")
			grayCS := NewPdfColorspaceDeviceGray()
			return grayCS.ColorToRGB(color)
		} else if this.N == 3 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceRGB (N=3)")
			// Already in RGB.
			return color, nil
		} else if this.N == 4 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceCMYK (N=4)")
			// CMYK
			cmykCS := NewPdfColorspaceDeviceCMYK()
			return cmykCS.ColorToRGB(color)
		} else {
			return nil, errors.New("ICC Based colorspace missing alternative")
		}
	}

	common.Log.Trace("ICC Based colorspace with alternative: %#v", this)
	return this.Alternate.ColorToRGB(color)
}

func (this *PdfColorspaceICCBased) ImageToRGB(img Image) (Image, error) {
	if this.Alternate == nil {
		common.Log.Debug("ICC Based colorspace missing alternative")
		if this.N == 1 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceGray (N=1)")
			grayCS := NewPdfColorspaceDeviceGray()
			return grayCS.ImageToRGB(img)
		} else if this.N == 3 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceRGB (N=3)")
			// Already in RGB.
			return img, nil
		} else if this.N == 4 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceCMYK (N=4)")
			// CMYK
			cmykCS := NewPdfColorspaceDeviceCMYK()
			return cmykCS.ImageToRGB(img)
		} else {
			return img, errors.New("ICC Based colorspace missing alternative")
		}
	}
	common.Log.Trace("ICC Based colorspace with alternative: %#v", this)

	output, err := this.Alternate.ImageToRGB(img)
	common.Log.Trace("ICC Input image: %+v", img)
	common.Log.Trace("ICC Output image: %+v", output)
	return output, err //this.Alternate.ImageToRGB(img)
}

//////////////////////
// Pattern color.

type PdfColorPattern struct {
	Color       PdfColor      // Color defined in underlying colorspace.
	PatternName PdfObjectName // Name of the pattern (reference via resource dicts).
}

// Pattern colorspace.
// Can be defined either as /Pattern or with an underlying colorspace [/Pattern cs].
type PdfColorspaceSpecialPattern struct {
	UnderlyingCS PdfColorspace

	container *PdfIndirectObject
}

func NewPdfColorspaceSpecialPattern() *PdfColorspaceSpecialPattern {
	return &PdfColorspaceSpecialPattern{}
}

func (this *PdfColorspaceSpecialPattern) String() string {
	return "Pattern"
}

func (this *PdfColorspaceSpecialPattern) GetNumComponents() int {
	return this.UnderlyingCS.GetNumComponents()
}

// DecodeArray returns an empty slice as there are no components associated with pattern colorspace.
func (this *PdfColorspaceSpecialPattern) DecodeArray() []float64 {
	return []float64{}
}

func newPdfColorspaceSpecialPatternFromPdfObject(obj PdfObject) (*PdfColorspaceSpecialPattern, error) {
	common.Log.Trace("New Pattern CS from obj: %s %T", obj.String(), obj)
	cs := NewPdfColorspaceSpecialPattern()

	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = TraceToDirectObject(obj)
	if name, isName := obj.(*PdfObjectName); isName {
		if *name != "Pattern" {
			return nil, fmt.Errorf("Invalid name")
		}

		return cs, nil
	}

	array, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Error("Invalid Pattern CS Object: %#v", obj)
		return nil, fmt.Errorf("Invalid Pattern CS object")
	}
	if len(*array) != 1 && len(*array) != 2 {
		common.Log.Error("Invalid Pattern CS array: %#v", array)
		return nil, fmt.Errorf("Invalid Pattern CS array")
	}

	obj = (*array)[0]
	if name, isName := obj.(*PdfObjectName); isName {
		if *name != "Pattern" {
			common.Log.Error("Invalid Pattern CS array name: %#v", name)
			return nil, fmt.Errorf("Invalid name")
		}
	}

	// Has an underlying color space.
	if len(*array) > 1 {
		obj = (*array)[1]
		obj = TraceToDirectObject(obj)
		baseCS, err := NewPdfColorspaceFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		cs.UnderlyingCS = baseCS
	}

	common.Log.Trace("Returning Pattern with underlying cs: %T", cs.UnderlyingCS)
	return cs, nil
}

func (this *PdfColorspaceSpecialPattern) ToPdfObject() PdfObject {
	if this.UnderlyingCS == nil {
		return MakeName("Pattern")
	}

	csObj := MakeArray(MakeName("Pattern"))
	csObj.Append(this.UnderlyingCS.ToPdfObject())

	if this.container != nil {
		this.container.PdfObject = csObj
		return this.container
	}

	return csObj
}

func (this *PdfColorspaceSpecialPattern) ColorFromFloats(vals []float64) (PdfColor, error) {
	if this.UnderlyingCS == nil {
		return nil, errors.New("Underlying CS not specified")
	}
	return this.UnderlyingCS.ColorFromFloats(vals)
}

// The first objects (if present) represent the color in underlying colorspace.  The last one represents
// the name of the pattern.
func (this *PdfColorspaceSpecialPattern) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) < 1 {
		return nil, errors.New("Invalid number of parameters")
	}
	patternColor := &PdfColorPattern{}

	// Pattern name.
	pname, ok := objects[len(objects)-1].(*PdfObjectName)
	if !ok {
		common.Log.Debug("Pattern name not a name (got %T)", objects[len(objects)-1])
		return nil, ErrTypeError
	}
	patternColor.PatternName = *pname

	// Pattern color if specified.
	if len(objects) > 1 {
		colorObjs := objects[0 : len(objects)-1]
		if this.UnderlyingCS == nil {
			common.Log.Debug("Pattern color with defined color components but underlying cs missing")
			return nil, errors.New("Underlying CS not defined")
		}
		color, err := this.UnderlyingCS.ColorFromPdfObjects(colorObjs)
		if err != nil {
			common.Log.Debug("ERROR: Unable to convert color via underlying cs: %v", err)
			return nil, err
		}
		patternColor.Color = color
	}

	return patternColor, nil
}

// Only converts color used with uncolored patterns (defined in underlying colorspace).  Does not go into the
// pattern objects and convert those.  If that is desired, needs to be done separately.  See for example
// grayscale conversion example in unidoc-examples repo.
func (this *PdfColorspaceSpecialPattern) ColorToRGB(color PdfColor) (PdfColor, error) {
	patternColor, ok := color.(*PdfColorPattern)
	if !ok {
		common.Log.Debug("Color not pattern (got %T)", color)
		return nil, ErrTypeError
	}

	if patternColor.Color == nil {
		// No color defined, can return same back.  No transform needed.
		return color, nil
	}

	if this.UnderlyingCS == nil {
		return nil, errors.New("Underlying CS not defined.")
	}

	return this.UnderlyingCS.ColorToRGB(patternColor.Color)
}

// An image cannot be defined in a pattern colorspace, returns an error.
func (this *PdfColorspaceSpecialPattern) ImageToRGB(img Image) (Image, error) {
	common.Log.Debug("Error: Image cannot be specified in Pattern colorspace")
	return img, errors.New("Invalid colorspace for image (pattern)")
}

//////////////////////
// Indexed colorspace. An indexed color space is a lookup table, where the input element is an index to the lookup
// table and the output is a color defined in the lookup table in the Base colorspace.
// [/Indexed base hival lookup]
type PdfColorspaceSpecialIndexed struct {
	Base   PdfColorspace
	HiVal  int
	Lookup PdfObject

	colorLookup []byte // m*(hival+1); m is number of components in Base colorspace

	container *PdfIndirectObject
}

func NewPdfColorspaceSpecialIndexed() *PdfColorspaceSpecialIndexed {
	cs := &PdfColorspaceSpecialIndexed{}
	cs.HiVal = 255
	return cs
}

func (this *PdfColorspaceSpecialIndexed) String() string {
	return "Indexed"
}

func (this *PdfColorspaceSpecialIndexed) GetNumComponents() int {
	return 1
}

// DecodeArray returns the component range values for the Indexed colorspace.
func (this *PdfColorspaceSpecialIndexed) DecodeArray() []float64 {
	return []float64{0, float64(this.HiVal)}
}

func newPdfColorspaceSpecialIndexedFromPdfObject(obj PdfObject) (*PdfColorspaceSpecialIndexed, error) {
	cs := NewPdfColorspaceSpecialIndexed()

	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = TraceToDirectObject(obj)
	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Type error")
	}

	if len(*array) != 4 {
		return nil, fmt.Errorf("Indexed CS: invalid array length")
	}

	// Check name.
	obj = (*array)[0]
	name, ok := obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("Indexed CS: invalid name")
	}
	if *name != "Indexed" {
		return nil, fmt.Errorf("Indexed CS: wrong name")
	}

	// Get base colormap.
	obj = (*array)[1]

	// Base cs cannot be another /Indexed or /Pattern space.
	baseName, err := determineColorspaceNameFromPdfObject(obj)
	if baseName == "Indexed" || baseName == "Pattern" {
		common.Log.Debug("Error: Indexed colorspace cannot have Indexed/Pattern CS as base (%v)", baseName)
		return nil, ErrRangeError
	}

	baseCs, err := NewPdfColorspaceFromPdfObject(obj)
	if err != nil {
		return nil, err
	}
	cs.Base = baseCs

	// Get hi val.
	obj = (*array)[2]
	val, err := getNumberAsInt64(obj)
	if err != nil {
		return nil, err
	}
	if val > 255 {
		return nil, fmt.Errorf("Indexed CS: Invalid hival")
	}
	cs.HiVal = int(val)

	// Index table.
	obj = (*array)[3]
	cs.Lookup = obj
	obj = TraceToDirectObject(obj)
	var data []byte
	if str, ok := obj.(*PdfObjectString); ok {
		data = []byte(*str)
		common.Log.Trace("Indexed string color data: % d", data)
	} else if stream, ok := obj.(*PdfObjectStream); ok {
		common.Log.Trace("Indexed stream: %s", obj.String())
		common.Log.Trace("Encoded (%d) : %# x", len(stream.Stream), stream.Stream)
		decoded, err := DecodeStream(stream)
		if err != nil {
			return nil, err
		}
		common.Log.Trace("Decoded (%d) : % X", len(decoded), decoded)
		data = decoded
	} else {
		return nil, fmt.Errorf("Indexed CS: Invalid table format")
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

func (this *PdfColorspaceSpecialIndexed) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("Range check")
	}

	N := this.Base.GetNumComponents()

	index := int(vals[0]) * N
	if index < 0 || (index+N-1) >= len(this.colorLookup) {
		return nil, errors.New("Outside range")
	}

	cvals := this.colorLookup[index : index+N]
	floats := []float64{}
	for _, val := range cvals {
		floats = append(floats, float64(val)/255.0)
	}
	color, err := this.Base.ColorFromFloats(floats)
	if err != nil {
		return nil, err
	}

	return color, nil
}

func (this *PdfColorspaceSpecialIndexed) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceSpecialIndexed) ColorToRGB(color PdfColor) (PdfColor, error) {
	if this.Base == nil {
		return nil, errors.New("Indexed base colorspace undefined")
	}

	return this.Base.ColorToRGB(color)
}

// Convert an indexed image to RGB.
func (this *PdfColorspaceSpecialIndexed) ImageToRGB(img Image) (Image, error) {
	//baseImage := img
	// Make a new representation of the image to be converted with the base colorspace.
	baseImage := Image{}
	baseImage.Height = img.Height
	baseImage.Width = img.Width
	baseImage.alphaData = img.alphaData
	baseImage.BitsPerComponent = img.BitsPerComponent
	baseImage.hasAlpha = img.hasAlpha
	baseImage.ColorComponents = img.ColorComponents

	samples := img.GetSamples()
	N := this.Base.GetNumComponents()

	baseSamples := []uint32{}
	// Convert the indexed data to base color map data.
	for i := 0; i < len(samples); i++ {
		// Each data point represents an index location.
		// For each entry there are N values.
		index := int(samples[i]) * N
		common.Log.Trace("Indexed Index: %d", index)
		// Ensure does not go out of bounds.
		if index+N-1 >= len(this.colorLookup) {
			// Clip to the end value.
			index = len(this.colorLookup) - N - 1
			common.Log.Trace("Clipping to index: %d", index)
		}

		cvals := this.colorLookup[index : index+N]
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
	return this.Base.ImageToRGB(baseImage)
}

// [/Indexed base hival lookup]
func (this *PdfColorspaceSpecialIndexed) ToPdfObject() PdfObject {
	csObj := MakeArray(MakeName("Indexed"))
	csObj.Append(this.Base.ToPdfObject())
	csObj.Append(MakeInteger(int64(this.HiVal)))
	csObj.Append(this.Lookup)

	if this.container != nil {
		this.container.PdfObject = csObj
		return this.container
	}

	return csObj
}

//////////////////////
// Separation colorspace.
// At the moment the colour space is set to a Separation space, the conforming reader shall determine whether the
// device has an available colorant (e.g. dye) corresponding to the name of the requested space. If so, the conforming
// reader shall ignore the alternateSpace and tintTransform parameters; subsequent painting operations within the
// space shall apply the designated colorant directly, according to the tint values supplied.
//
// Format: [/Separation name alternateSpace tintTransform]
type PdfColorspaceSpecialSeparation struct {
	ColorantName   *PdfObjectName
	AlternateSpace PdfColorspace
	TintTransform  PdfFunction

	// Container, if when parsing CS array is inside a container.
	container *PdfIndirectObject
}

func NewPdfColorspaceSpecialSeparation() *PdfColorspaceSpecialSeparation {
	cs := &PdfColorspaceSpecialSeparation{}
	return cs
}

func (this *PdfColorspaceSpecialSeparation) String() string {
	return "Separation"
}

func (this *PdfColorspaceSpecialSeparation) GetNumComponents() int {
	return 1
}

// DecodeArray returns the component range values for the Separation colorspace.
func (this *PdfColorspaceSpecialSeparation) DecodeArray() []float64 {
	return []float64{0, 1.0}
}

// Object is an array or indirect object containing the array.
func newPdfColorspaceSpecialSeparationFromPdfObject(obj PdfObject) (*PdfColorspaceSpecialSeparation, error) {
	cs := NewPdfColorspaceSpecialSeparation()

	// If within an indirect object, then make a note of it.  If we write out the PdfObject later
	// we can reference the same container.  Otherwise is not within a container, but rather
	// a new array.
	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	obj = TraceToDirectObject(obj)
	array, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Separation CS: Invalid object")
	}

	if len(*array) != 4 {
		return nil, fmt.Errorf("Separation CS: Incorrect array length")
	}

	// Check name.
	obj = (*array)[0]
	name, ok := obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("Separation CS: invalid family name")
	}
	if *name != "Separation" {
		return nil, fmt.Errorf("Separation CS: wrong family name")
	}

	// Get colorant name.
	obj = (*array)[1]
	name, ok = obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("Separation CS: Invalid colorant name")
	}
	cs.ColorantName = name

	// Get base colormap.
	obj = (*array)[2]
	alternativeCs, err := NewPdfColorspaceFromPdfObject(obj)
	if err != nil {
		return nil, err
	}
	cs.AlternateSpace = alternativeCs

	// Tint transform is specified by a PDF function.
	tintTransform, err := newPdfFunctionFromPdfObject((*array)[3])
	if err != nil {
		return nil, err
	}

	cs.TintTransform = tintTransform

	return cs, nil
}

func (this *PdfColorspaceSpecialSeparation) ToPdfObject() PdfObject {
	csArray := MakeArray(MakeName("Separation"))

	csArray.Append(this.ColorantName)
	csArray.Append(this.AlternateSpace.ToPdfObject())
	csArray.Append(this.TintTransform.ToPdfObject())

	// If in a container, replace the contents and return back.
	// Helps not getting too many duplicates of the same objects.
	if this.container != nil {
		this.container.PdfObject = csArray
		return this.container
	}

	return csArray
}

func (this *PdfColorspaceSpecialSeparation) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != 1 {
		return nil, errors.New("Range check")
	}

	tint := vals[0]
	input := []float64{tint}
	output, err := this.TintTransform.Evaluate(input)
	if err != nil {
		common.Log.Debug("Error, failed to evaluate: %v", err)
		common.Log.Trace("Tint transform: %+v", this.TintTransform)
		return nil, err
	}

	common.Log.Trace("Processing ColorFromFloats(%+v) on AlternateSpace: %#v", output, this.AlternateSpace)
	color, err := this.AlternateSpace.ColorFromFloats(output)
	if err != nil {
		common.Log.Debug("Error, failed to evaluate in alternate space: %v", err)
		return nil, err
	}

	return color, nil
}

func (this *PdfColorspaceSpecialSeparation) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != 1 {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceSpecialSeparation) ColorToRGB(color PdfColor) (PdfColor, error) {
	if this.AlternateSpace == nil {
		return nil, errors.New("Alternate colorspace undefined")
	}

	return this.AlternateSpace.ColorToRGB(color)
}

// ImageToRGB converts an image with samples in Separation CS to an image with samples specified in
// DeviceRGB CS.
func (this *PdfColorspaceSpecialSeparation) ImageToRGB(img Image) (Image, error) {
	altImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	common.Log.Trace("Separation color space -> ToRGB conversion")
	common.Log.Trace("samples in: %d", len(samples))
	common.Log.Trace("TintTransform: %+v", this.TintTransform)

	altDecode := this.AlternateSpace.DecodeArray()

	altSamples := []uint32{}
	// Convert tints to color data in the alternate colorspace.
	for i := 0; i < len(samples); i++ {
		// A single tint component is in the range 0.0 - 1.0
		tint := float64(samples[i]) / maxVal

		// Convert the tint value to the alternate space value.
		outputs, err := this.TintTransform.Evaluate([]float64{tint})
		//common.Log.Trace("%v Converting tint value: %f -> [% f]", this.AlternateSpace, tint, outputs)

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
	altImage.ColorComponents = this.AlternateSpace.GetNumComponents()

	// Set the image's decode parameters for interpretation in the alternative CS.
	altImage.decode = altDecode

	// Convert to RGB via the alternate colorspace.
	return this.AlternateSpace.ImageToRGB(altImage)
}

//////////////////////
// DeviceN color spaces are similar to Separation color spaces, except they can contain an arbitrary
// number of color components.
//
// Format: [/DeviceN names alternateSpace tintTransform]
//     or: [/DeviceN names alternateSpace tintTransform attributes]
type PdfColorspaceDeviceN struct {
	ColorantNames  *PdfObjectArray
	AlternateSpace PdfColorspace
	TintTransform  PdfFunction
	Attributes     *PdfColorspaceDeviceNAttributes

	// Optional
	container *PdfIndirectObject
}

func NewPdfColorspaceDeviceN() *PdfColorspaceDeviceN {
	cs := &PdfColorspaceDeviceN{}
	return cs
}

func (this *PdfColorspaceDeviceN) String() string {
	return "DeviceN"
}

// GetNumComponents returns the number of input color components, i.e. that are input to the tint transform.
func (this *PdfColorspaceDeviceN) GetNumComponents() int {
	return len(*this.ColorantNames)
}

// DecodeArray returns the component range values for the DeviceN colorspace.
// [0 1.0 0 1.0 ...] for each color component.
func (this *PdfColorspaceDeviceN) DecodeArray() []float64 {
	decode := []float64{}
	for i := 0; i < this.GetNumComponents(); i++ {
		decode = append(decode, 0.0, 1.0)
	}
	return decode
}

func newPdfColorspaceDeviceNFromPdfObject(obj PdfObject) (*PdfColorspaceDeviceN, error) {
	cs := NewPdfColorspaceDeviceN()

	// If within an indirect object, then make a note of it.  If we write out the PdfObject later
	// we can reference the same container.  Otherwise is not within a container, but rather
	// a new array.
	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		cs.container = indObj
	}

	// Check the CS array.
	obj = TraceToDirectObject(obj)
	csArray, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("DeviceN CS: Invalid object")
	}

	if len(*csArray) != 4 && len(*csArray) != 5 {
		return nil, fmt.Errorf("DeviceN CS: Incorrect array length")
	}

	// Check name.
	obj = (*csArray)[0]
	name, ok := obj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("DeviceN CS: invalid family name")
	}
	if *name != "DeviceN" {
		return nil, fmt.Errorf("DeviceN CS: wrong family name")
	}

	// Get colorant names.  Specifies the number of components too.
	obj = (*csArray)[1]
	obj = TraceToDirectObject(obj)
	nameArray, ok := obj.(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("DeviceN CS: Invalid names array")
	}
	cs.ColorantNames = nameArray

	// Get base colormap.
	obj = (*csArray)[2]
	alternativeCs, err := NewPdfColorspaceFromPdfObject(obj)
	if err != nil {
		return nil, err
	}
	cs.AlternateSpace = alternativeCs

	// Tint transform is specified by a PDF function.
	tintTransform, err := newPdfFunctionFromPdfObject((*csArray)[3])
	if err != nil {
		return nil, err
	}
	cs.TintTransform = tintTransform

	// Attributes.
	if len(*csArray) == 5 {
		attr, err := newPdfColorspaceDeviceNAttributesFromPdfObject((*csArray)[4])
		if err != nil {
			return nil, err
		}
		cs.Attributes = attr
	}

	return cs, nil
}

// Format: [/DeviceN names alternateSpace tintTransform]
//     or: [/DeviceN names alternateSpace tintTransform attributes]

func (this *PdfColorspaceDeviceN) ToPdfObject() PdfObject {
	csArray := MakeArray(MakeName("DeviceN"))
	csArray.Append(this.ColorantNames)
	csArray.Append(this.AlternateSpace.ToPdfObject())
	csArray.Append(this.TintTransform.ToPdfObject())
	if this.Attributes != nil {
		csArray.Append(this.Attributes.ToPdfObject())
	}

	if this.container != nil {
		this.container.PdfObject = csArray
		return this.container
	}

	return csArray
}

func (this *PdfColorspaceDeviceN) ColorFromFloats(vals []float64) (PdfColor, error) {
	if len(vals) != this.GetNumComponents() {
		return nil, errors.New("Range check")
	}

	output, err := this.TintTransform.Evaluate(vals)
	if err != nil {
		return nil, err
	}

	color, err := this.AlternateSpace.ColorFromFloats(output)
	if err != nil {
		return nil, err
	}
	return color, nil
}

func (this *PdfColorspaceDeviceN) ColorFromPdfObjects(objects []PdfObject) (PdfColor, error) {
	if len(objects) != this.GetNumComponents() {
		return nil, errors.New("Range check")
	}

	floats, err := getNumbersAsFloat(objects)
	if err != nil {
		return nil, err
	}

	return this.ColorFromFloats(floats)
}

func (this *PdfColorspaceDeviceN) ColorToRGB(color PdfColor) (PdfColor, error) {
	if this.AlternateSpace == nil {
		return nil, errors.New("DeviceN alternate space undefined")
	}
	return this.AlternateSpace.ColorToRGB(color)
}

func (this *PdfColorspaceDeviceN) ImageToRGB(img Image) (Image, error) {
	altImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	// Convert tints to color data in the alternate colorspace.
	altSamples := []uint32{}
	for i := 0; i < len(samples); i += this.GetNumComponents() {
		// The input to the tint transformation is the tint
		// for each color component.
		//
		// A single tint component is in the range 0.0 - 1.0
		inputs := []float64{}
		for j := 0; j < this.GetNumComponents(); j++ {
			tint := float64(samples[i+j]) / maxVal
			inputs = append(inputs, tint)
		}

		// Transform the tints to the alternate colorspace.
		// (scaled units).
		outputs, err := this.TintTransform.Evaluate(inputs)
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
	return this.AlternateSpace.ImageToRGB(altImage)
}

// Additional information about the components of colour space that conforming readers may use.
// Conforming readers need not use the alternateSpace and tintTransform parameters, and may
// instead use custom blending algorithms, along with other information provided in the attributes
// dictionary if present.
type PdfColorspaceDeviceNAttributes struct {
	Subtype     *PdfObjectName // DeviceN or NChannel (DeviceN default)
	Colorants   PdfObject
	Process     PdfObject
	MixingHints PdfObject

	// Optional
	container *PdfIndirectObject
}

func newPdfColorspaceDeviceNAttributesFromPdfObject(obj PdfObject) (*PdfColorspaceDeviceNAttributes, error) {
	attr := &PdfColorspaceDeviceNAttributes{}

	var dict *PdfObjectDictionary
	if indObj, isInd := obj.(*PdfIndirectObject); isInd {
		attr.container = indObj
		var ok bool
		dict, ok = indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			common.Log.Error("DeviceN attribute type error")
			return nil, errors.New("Type error")
		}
	} else if d, isDict := obj.(*PdfObjectDictionary); isDict {
		dict = d
	} else {
		common.Log.Error("DeviceN attribute type error")
		return nil, errors.New("Type error")
	}

	if obj := dict.Get("Subtype"); obj != nil {
		name, ok := TraceToDirectObject(obj).(*PdfObjectName)
		if !ok {
			common.Log.Error("DeviceN attribute Subtype type error")
			return nil, errors.New("Type error")
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

func (this *PdfColorspaceDeviceNAttributes) ToPdfObject() PdfObject {
	dict := MakeDict()

	if this.Subtype != nil {
		dict.Set("Subtype", this.Subtype)
	}
	dict.SetIfNotNil("Colorants", this.Colorants)
	dict.SetIfNotNil("Process", this.Process)
	dict.SetIfNotNil("MixingHints", this.MixingHints)

	if this.container != nil {
		this.container.PdfObject = dict
		return this.container
	}

	return dict
}
