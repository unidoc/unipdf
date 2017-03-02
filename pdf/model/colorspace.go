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
// Work is in progress to support all colorspaces.
//
type PdfColorspace interface {
	ToRGB(Image) (Image, error)
	GetNumComponents() int
	ToPdfObject() PdfObject
}

func newPdfColorspaceFromPdfObject(obj PdfObject) (PdfColorspace, error) {
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
			if *name == "CalGray" {
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
			}
		}
	}

	common.Log.Error("Colorspace type error: %s", obj.String())
	return nil, errors.New("Type error")
}

// Gray scale component.
// No specific parameters
type PdfColorspaceDeviceGray struct{}

func NewPdfColorspaceDeviceGray() *PdfColorspaceDeviceGray {
	return &PdfColorspaceDeviceGray{}
}

func (this *PdfColorspaceDeviceGray) GetNumComponents() int {
	return 1
}

func (this *PdfColorspaceDeviceGray) ToPdfObject() PdfObject {
	return MakeName("DeviceGray")
}

// Convert 1-component grayscale data to 3-component RGB.
func (this *PdfColorspaceDeviceGray) ToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()

	rgbSamples := []uint32{}
	for i := 0; i < len(samples); i++ {
		grayVal := samples[i]
		rgbSamples = append(rgbSamples, grayVal, grayVal, grayVal)
	}
	rgbImage.SetSamples(rgbSamples)

	return rgbImage, nil
}

// R, G, B components.
// No specific parameters
type PdfColorspaceDeviceRGB struct{}

func NewPdfColorspaceDeviceRGB() *PdfColorspaceDeviceRGB {
	return &PdfColorspaceDeviceRGB{}
}

func (this *PdfColorspaceDeviceRGB) GetNumComponents() int {
	return 3
}

func (this *PdfColorspaceDeviceRGB) ToPdfObject() PdfObject {
	return MakeName("DeviceRGB")
}

func (this *PdfColorspaceDeviceRGB) ToRGB(img Image) (Image, error) {
	return img, nil
}

func (this *PdfColorspaceDeviceRGB) ToGray(img Image) (Image, error) {
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

	return grayImage, nil
}

// C, M, Y, K components.
// No other parameters.
type PdfColorspaceDeviceCMYK struct{}

func NewPdfColorspaceDeviceCMYK() *PdfColorspaceDeviceCMYK {
	return &PdfColorspaceDeviceCMYK{}
}

func (this *PdfColorspaceDeviceCMYK) GetNumComponents() int {
	return 4
}

func (this *PdfColorspaceDeviceCMYK) ToPdfObject() PdfObject {
	return MakeName("DeviceCMYK")
}

func (this *PdfColorspaceDeviceCMYK) ToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()

	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1
	rgbSamples := []uint32{}
	for i := 0; i < len(samples); i += 4 {
		// Normalized c, m, y, k values.
		c := float64(samples[i]) / maxVal
		m := float64(samples[i+1]) / maxVal
		y := float64(samples[i+2]) / maxVal
		k := float64(samples[i+3]) / maxVal

		// K(ey) is black.
		r := 1.0 - math.Min(1.0, c+k)
		g := 1.0 - math.Min(1.0, m+k)
		b := 1.0 - math.Min(1.0, y+k)

		// Convert to uint32 format.
		R := uint32(r * maxVal)
		G := uint32(g * maxVal)
		B := uint32(b * maxVal)

		rgbSamples = append(rgbSamples, R, G, B)
	}
	rgbImage.SetSamples(rgbSamples)

	return rgbImage, nil
}

// CIE based gray level.
// Single component
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

func (this *PdfColorspaceCalGray) GetNumComponents() int {
	return 1
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
	obj = (*dict)["WhitePoint"]
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
	obj = (*dict)["BlackPoint"]
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
	obj = (*dict)["Gamma"]
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

	dict := PdfObjectDictionary{}
	if this.WhitePoint != nil {
		dict["WhitePoint"] = MakeArray(MakeFloat(this.WhitePoint[0]), MakeFloat(this.WhitePoint[1]), MakeFloat(this.WhitePoint[2]))
	} else {
		common.Log.Error("CalGray: Missing WhitePoint (Required)")
	}

	if this.BlackPoint != nil {
		dict["BlackPoint"] = MakeArray(MakeFloat(this.BlackPoint[0]), MakeFloat(this.BlackPoint[1]), MakeFloat(this.BlackPoint[2]))
	}

	dict["Gamma"] = MakeFloat(this.Gamma)
	cspace.Append(&dict)

	if this.container != nil {
		this.container.PdfObject = cspace
		return this.container
	}

	return cspace
}

// A, B, C -> X, Y, Z
func (this *PdfColorspaceCalGray) ToRGB(img Image) (Image, error) {
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

	return rgbImage, nil
}

// Colorimetric CIE RGB colorspace.
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

func (this *PdfColorspaceCalRGB) GetNumComponents() int {
	return 1
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
	obj = (*dict)["WhitePoint"]
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
	obj = (*dict)["BlackPoint"]
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
	obj = (*dict)["Gamma"]
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
	obj = (*dict)["Matrix"]
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

	dict := PdfObjectDictionary{}
	if this.WhitePoint != nil {
		dict["WhitePoint"] = MakeArray(MakeFloat(this.WhitePoint[0]), MakeFloat(this.WhitePoint[1]), MakeFloat(this.WhitePoint[2]))
	} else {
		common.Log.Error("CalRGB: Missing WhitePoint (Required)")
	}

	if this.BlackPoint != nil {
		dict["BlackPoint"] = MakeArray(MakeFloat(this.BlackPoint[0]), MakeFloat(this.BlackPoint[1]), MakeFloat(this.BlackPoint[2]))
	}
	if this.Gamma != nil {
		dict["Gamma"] = MakeArray(MakeFloat(this.Gamma[0]), MakeFloat(this.Gamma[1]), MakeFloat(this.Gamma[2]))
	}
	if this.Matrix != nil {
		dict["Matrix"] = MakeArray(MakeFloat(this.Matrix[0]), MakeFloat(this.Matrix[1]), MakeFloat(this.Matrix[2]),
			MakeFloat(this.Matrix[3]), MakeFloat(this.Matrix[4]), MakeFloat(this.Matrix[5]),
			MakeFloat(this.Matrix[6]), MakeFloat(this.Matrix[7]), MakeFloat(this.Matrix[8]))
	}
	cspace.Append(&dict)

	if this.container != nil {
		this.container.PdfObject = cspace
		return this.container
	}

	return cspace
}

func (this *PdfColorspaceCalRGB) ToRGB(img Image) (Image, error) {
	rgbImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	rgbSamples := []uint32{}
	for i := 0; i < len(samples); i++ {
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

	return rgbImage, nil
}

// L*, a*, b* 3 component colorspace.
type PdfColorspaceLab struct {
	WhitePoint []float64 // Required.
	BlackPoint []float64
	Range      []float64 // [amin amax bmin bmax]

	container *PdfIndirectObject
}

func (this *PdfColorspaceLab) GetNumComponents() int {
	return 3
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
	obj = (*dict)["WhitePoint"]
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
	obj = (*dict)["BlackPoint"]
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
	obj = (*dict)["Range"]
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

	dict := PdfObjectDictionary{}
	if this.WhitePoint != nil {
		dict["WhitePoint"] = MakeArray(MakeFloat(this.WhitePoint[0]), MakeFloat(this.WhitePoint[1]), MakeFloat(this.WhitePoint[2]))
	} else {
		common.Log.Error("Lab: Missing WhitePoint (Required)")
	}

	if this.BlackPoint != nil {
		dict["BlackPoint"] = MakeArray(MakeFloat(this.BlackPoint[0]), MakeFloat(this.BlackPoint[1]), MakeFloat(this.BlackPoint[2]))
	}

	if this.Range != nil {
		dict["Range"] = MakeArray(MakeFloat(this.Range[0]), MakeFloat(this.Range[1]), MakeFloat(this.Range[2]), MakeFloat(this.Range[3]))
	}
	csObj.Append(&dict)

	if this.container != nil {
		this.container.PdfObject = csObj
		return this.container
	}

	return csObj
}

func (this *PdfColorspaceLab) ToRGB(img Image) (Image, error) {
	g := func(x float64) float64 {
		if x >= 6.0/29 {
			return x * x * x
		} else {
			return 108.0 / 841 * (x - 4/29)
		}
	}

	rgbImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	rgbSamples := []uint32{}
	for i := 0; i < len(samples); i += 3 {
		// Get normalized L*, a*, b* values. [0-1]
		LNorm := float64(samples[i]) / maxVal
		ANorm := float64(samples[i+1]) / maxVal
		BNorm := float64(samples[i+2]) / maxVal

		// Rescale them.  Not very clearly specified.
		// L: [0,1] -> [0, 100]
		// a*: [0,1] -> [-128, 127] or Range if specified
		// b*: [0,1] -> [-100, 100] or Range if specified
		LStar := float64(LNorm * 100.0)
		rng := []float64{-128.0, 127.0}
		if this.Range != nil {
			rng = this.Range
		}
		AStar := interpolate(ANorm, 0.0, 1.0, rng[0], rng[1])
		BStar := interpolate(BNorm, 0.0, 1.0, rng[0], rng[1])

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

	return rgbImage, nil
}

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

	n, ok := (*dict)["N"].(*PdfObjectInteger)
	if !ok {
		return nil, fmt.Errorf("ICCBased missing N from stream dict")
	}
	if *n != 1 && *n != 3 && *n != 4 {
		return nil, fmt.Errorf("ICCBased colorspace invalid N (not 1,3,4)")
	}
	cs.N = int(*n)

	if obj, has := (*dict)["Alternate"]; has {
		alternate, err := newPdfColorspaceFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		cs.Alternate = alternate
	}

	if obj, has := (*dict)["Range"]; has {
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

	if obj, has := (*dict)["Metadata"]; has {
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
	dict := &PdfObjectDictionary{}

	(*dict)["N"] = MakeInteger(int64(this.N))

	if this.Alternate != nil {
		(*dict)["Alternate"] = this.Alternate.ToPdfObject()
	}

	if this.Metadata != nil {
		(*dict)["Metadata"] = this.Metadata
	}
	if this.Range != nil {
		(*dict)["Range"] = MakeArray(MakeFloat(this.Range[0]), MakeFloat(this.Range[1]), MakeFloat(this.Range[2]), MakeFloat(this.Range[3]))
	}

	// Encode with a default encoder?
	(*dict)["Length"] = MakeInteger(int64(len(this.Data)))
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

func (this *PdfColorspaceICCBased) ToRGB(img Image) (Image, error) {
	if this.Alternate == nil {
		common.Log.Debug("ICC Based colorspace missing alternative")
		if this.N == 1 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceGray (N=1)")
			grayCS := NewPdfColorspaceDeviceGray()
			return grayCS.ToRGB(img)
		} else if this.N == 3 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceRGB (N=3)")
			// Already in RGB.
			return img, nil
		} else if this.N == 4 {
			common.Log.Debug("ICC Based colorspace missing alternative - using DeviceCMYK (N=4)")
			// CMYK
			cmykCS := NewPdfColorspaceDeviceCMYK()
			return cmykCS.ToRGB(img)
		} else {
			return img, errors.New("ICC Based colorspace missing alternative")
		}
	}

	return this.Alternate.ToRGB(img)
}

// Pattern color.
// See 8.6 for info about color spaces and 11.6.7 patterns and transparency.
// Simply /Pattern ?
// Can be [/Pattern /DeviceRGB] too?
// Need to investigate more, specs not clear... XXX
type PdfColorspaceSpecialPattern struct {
	UnderlyingCS PdfColorspace

	container *PdfIndirectObject
}

func NewPdfColorspaceSpecialPattern() *PdfColorspaceSpecialPattern {
	return &PdfColorspaceSpecialPattern{}
}

func (this *PdfColorspaceSpecialPattern) GetNumComponents() int {
	return this.UnderlyingCS.GetNumComponents()
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
		baseCS, err := newPdfColorspaceFromPdfObject(obj)
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

func (this *PdfColorspaceSpecialPattern) ToRGB(img Image) (Image, error) {
	return img, errors.New("Not implemented")
}

//
// Indexed colorspace.
// [/Indexed base hival lookup]
//
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

func (this *PdfColorspaceSpecialIndexed) GetNumComponents() int {
	if this.Base == nil {
		common.Log.Error("Base colorspace not set!")
		return 0
	}

	return this.Base.GetNumComponents()
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
	baseCs, err := newPdfColorspaceFromPdfObject(obj)
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
	} else if stream, ok := obj.(*PdfObjectStream); ok {
		common.Log.Debug("Indexed stream: %s", obj.String())
		common.Log.Debug("Encoded (%d) : %# x", len(stream.Stream), stream.Stream)
		decoded, err := DecodeStream(stream)
		if err != nil {
			return nil, err
		}
		common.Log.Debug("Decoded (%d) : % X", len(decoded), decoded)
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

// Convert to RGB image.
func (this *PdfColorspaceSpecialIndexed) ToRGB(img Image) (Image, error) {
	baseImage := img

	samples := img.GetSamples()
	N := this.Base.GetNumComponents()

	baseSamples := []uint32{}
	// Convert the indexed data to base color map data.
	for i := 0; i < len(samples); i++ {
		// Each data point represents an index location.
		// For each entry there are N values.
		index := int(samples[i]) * N
		// Ensure does not go out of bounds.
		if index+N-1 >= len(this.colorLookup) {
			// Clip to the end value.
			index = len(this.colorLookup) - N - 1
		}

		cvals := this.colorLookup[index : index+N]
		for _, val := range cvals {
			baseSamples = append(baseSamples, uint32(val))
		}
	}
	baseImage.SetSamples(baseSamples)

	// Convert to rgb.
	return this.Base.ToRGB(baseImage)
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

//
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

func (this *PdfColorspaceSpecialSeparation) GetNumComponents() int {
	return 1
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
	alternativeCs, err := newPdfColorspaceFromPdfObject(obj)
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

func (this *PdfColorspaceSpecialSeparation) ToRGB(img Image) (Image, error) {
	altImage := img

	samples := img.GetSamples()
	maxVal := math.Pow(2, float64(img.BitsPerComponent)) - 1

	altSamples := []uint32{}
	// Convert tints to color data in the alternate colorspace.
	for i := 0; i < len(samples); i++ {
		// A single tint component is in the range 0.0 - 1.0
		tint := float64(samples[i]) / maxVal

		// Convert the tint value to the alternate space value.
		outputs, err := this.TintTransform.Evaluate([]float64{tint})
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
	return this.AlternateSpace.ToRGB(altImage)
}

//
// DeviceN color spaces may contain an arbitrary number of color components.
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

func (this *PdfColorspaceDeviceN) GetNumComponents() int {
	return len(*this.ColorantNames)
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
	alternativeCs, err := newPdfColorspaceFromPdfObject(obj)
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

func (this *PdfColorspaceDeviceN) ToRGB(img Image) (Image, error) {
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
	return this.AlternateSpace.ToRGB(altImage)
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

	if obj, has := (*dict)["Subtype"]; has {
		name, ok := TraceToDirectObject(obj).(*PdfObjectName)
		if !ok {
			common.Log.Error("DeviceN attribute Subtype type error")
			return nil, errors.New("Type error")
		}

		attr.Subtype = name
	}

	if obj, has := (*dict)["Colorants"]; has {
		attr.Colorants = obj
	}

	if obj, has := (*dict)["Process"]; has {
		attr.Process = obj
	}

	if obj, has := (*dict)["MixingHints"]; has {
		attr.MixingHints = obj
	}

	return attr, nil
}

func (this *PdfColorspaceDeviceNAttributes) ToPdfObject() PdfObject {
	dict := &PdfObjectDictionary{}

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
