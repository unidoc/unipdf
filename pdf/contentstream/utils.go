/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"errors"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

func makeParamsFromFloats(vals []float64) []core.PdfObject {
	params := []core.PdfObject{}
	for _, val := range vals {
		params = append(params, core.MakeFloat(val))
	}
	return params
}

func makeParamsFromNames(vals []core.PdfObjectName) []core.PdfObject {
	params := []core.PdfObject{}
	for _, val := range vals {
		params = append(params, core.MakeName(string(val)))
	}
	return params
}

func makeParamsFromStrings(vals []core.PdfObjectString) []core.PdfObject {
	params := []core.PdfObject{}
	for _, val := range vals {
		params = append(params, core.MakeString(string(val)))
	}
	return params
}

func makeParamsFromInts(vals []int64) []core.PdfObject {
	params := []core.PdfObject{}
	for _, val := range vals {
		params = append(params, core.MakeInteger(val))
	}
	return params
}

func newIndexedColorspaceFromPdfObject(obj core.PdfObject) (model.PdfColorspace, error) {
	arr, ok := obj.(*core.PdfObjectArray)
	if !ok {
		common.Log.Debug("Error: Invalid indexed cs not in array (%#v)", obj)
		return nil, errors.New("Type check error")
	}

	if len(*arr) != 4 {
		common.Log.Debug("Error: Invalid cs array, length != 4 (%d)", len(*arr))
		return nil, errors.New("Range check error")
	}

	// Format is [/I base 255 bytes], where base = /G,/RGB,/CMYK
	name, ok := (*arr)[0].(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("Error: Invalid cs array first element not a name (array: %#v)", *arr)
		return nil, errors.New("Type check error")
	}
	if *name != "I" && *name != "Indexed" {
		common.Log.Debug("Error: Invalid cs array first element != I (got: %v)", *name)
		return nil, errors.New("Range check error")
	}

	// Check base
	name, ok = (*arr)[1].(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("Error: Invalid cs array 2nd element not a name (array: %#v)", *arr)
		return nil, errors.New("Type check error")
	}
	if *name != "G" && *name != "RGB" && *name != "CMYK" && *name != "DeviceGray" && *name != "DeviceRGB" && *name != "DeviceCMYK" {
		common.Log.Debug("Error: Invalid cs array 2nd element != G/RGB/CMYK (got: %v)", *name)
		return nil, errors.New("Range check error")
	}
	basename := ""
	switch *name {
	case "G", "DeviceGray":
		basename = "DeviceGray"
	case "RGB", "DeviceRGB":
		basename = "DeviceRGB"
	case "CMYK", "DeviceCMYK":
		basename = "DeviceCMYK"
	}

	// Prepare to a format that can be loaded by model's newPdfColorspaceFromPdfObject.
	csArr := core.MakeArray(core.MakeName("Indexed"), core.MakeName(basename), (*arr)[2], (*arr)[3])

	return model.NewPdfColorspaceFromPdfObject(csArr)
}
