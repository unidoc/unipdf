/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

// There are 7 types of shading, indicated by the shading type variable:
// 1: Function-based shading.
// 2: Axial shading.
// 3: Radial shading.
// 4: Free-form Gouraud-shaded triangle mesh.
// 5: Lattice-form Gouraud-shaded triangle mesh.
// 6: Coons patch mesh.
// 7: Tensor-product patch mesh.
// types 4-7 are contained in a stream object, where the dictionary is given by the stream dictionary.
type PdfShading struct {
	ShadingType *PdfObjectInteger
	ColorSpace  PdfColorspace
	Background  *PdfObjectArray
	BBox        *PdfRectangle
	AntiAlias   *PdfObjectBool

	context   PdfModel  // The sub shading type entry (types 1-7).  Represented by PdfShadingType1-7.
	container PdfObject // The container.  Can be stream, indirect object, or dictionary.
}

func (this *PdfShading) GetContainingPdfObject() PdfObject {
	return this.container
}

// Context in this case is a reference to the subshading entry as represented by PdfShadingType1-7.
func (this *PdfShading) GetContext() PdfModel {
	return this.context
}

// Set the sub annotation (context).
func (this *PdfShading) SetContext(ctx PdfModel) {
	this.context = ctx
}

func (this *PdfShading) getShadingDict() (*PdfObjectDictionary, error) {
	obj := this.container

	if indObj, isInd := obj.(*PdfIndirectObject); isInd {
		d, ok := indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			return nil, ErrTypeError
		}

		return d, nil
	} else if streamObj, isStream := obj.(*PdfObjectStream); isStream {
		return streamObj.PdfObjectDictionary, nil
	} else if d, isDict := obj.(*PdfObjectDictionary); isDict {
		return d, nil
	} else {
		common.Log.Debug("Unable to access shading dictionary")
		return nil, ErrTypeError
	}
}

// Shading type 1: Function-based shading.
type PdfShadingType1 struct {
	*PdfShading
	Domain   *PdfObjectArray
	Matrix   *PdfObjectArray
	Function []PdfFunction
}

// Shading type 2: Axial shading.
type PdfShadingType2 struct {
	*PdfShading
	Coords   *PdfObjectArray
	Domain   *PdfObjectArray
	Function []PdfFunction
	Extend   *PdfObjectArray
}

// Shading type 3: Radial shading.
type PdfShadingType3 struct {
	*PdfShading
	Coords   *PdfObjectArray
	Domain   *PdfObjectArray
	Function []PdfFunction
	Extend   *PdfObjectArray
}

// Shading type 4: Free-form Gouraud-shaded triangle mesh.
type PdfShadingType4 struct {
	*PdfShading
	BitsPerCoordinate *PdfObjectInteger
	BitsPerComponent  *PdfObjectInteger
	BitsPerFlag       *PdfObjectInteger
	Decode            *PdfObjectArray
	Function          []PdfFunction
}

// Shading type 5: Lattice-form Gouraud-shaded triangle mesh.
type PdfShadingType5 struct {
	*PdfShading
	BitsPerCoordinate *PdfObjectInteger
	BitsPerComponent  *PdfObjectInteger
	VerticesPerRow    *PdfObjectInteger
	Decode            *PdfObjectArray
	Function          []PdfFunction
}

// Shading type 6: Coons patch mesh.
type PdfShadingType6 struct {
	*PdfShading
	BitsPerCoordinate *PdfObjectInteger
	BitsPerComponent  *PdfObjectInteger
	BitsPerFlag       *PdfObjectInteger
	Decode            *PdfObjectArray
	Function          []PdfFunction
}

// Shading type 7: Tensor-product patch mesh.
type PdfShadingType7 struct {
	*PdfShading
	BitsPerCoordinate *PdfObjectInteger
	BitsPerComponent  *PdfObjectInteger
	BitsPerFlag       *PdfObjectInteger
	Decode            *PdfObjectArray
	Function          []PdfFunction
}

// Used for PDF parsing. Loads the PDF shading from a PDF object.
// Can be either an indirect object (types 1-3) containing the dictionary, or a stream object with the stream
// dictionary containing the shading dictionary (types 4-7).
func newPdfShadingFromPdfObject(obj PdfObject) (*PdfShading, error) {
	shading := &PdfShading{}

	var dict *PdfObjectDictionary
	if indObj, isInd := obj.(*PdfIndirectObject); isInd {
		shading.container = indObj

		d, ok := indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			common.Log.Debug("Object not a dictionary type")
			return nil, ErrTypeError
		}

		dict = d
	} else if streamObj, isStream := obj.(*PdfObjectStream); isStream {
		shading.container = streamObj
		dict = streamObj.PdfObjectDictionary
	} else if d, isDict := obj.(*PdfObjectDictionary); isDict {
		shading.container = d
		dict = d
	} else {
		common.Log.Debug("Object type unexpected (%T)", obj)
		return nil, ErrTypeError
	}

	if dict == nil {
		common.Log.Debug("Dictionary missing")
		return nil, errors.New("Dict missing")
	}

	// Shading type (required).
	obj = dict.Get("ShadingType")
	if obj == nil {
		common.Log.Debug("Required shading type missing")
		return nil, ErrRequiredAttributeMissing
	}
	obj = TraceToDirectObject(obj)
	shadingType, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("Invalid type for shading type (%T)", obj)
		return nil, ErrTypeError
	}
	if *shadingType < 1 || *shadingType > 7 {
		common.Log.Debug("Invalid shading type, not 1-7 (got %d)", *shadingType)
		return nil, ErrTypeError
	}
	shading.ShadingType = shadingType

	// Color space (required).
	obj = dict.Get("ColorSpace")
	if obj == nil {
		common.Log.Debug("Required ColorSpace entry missing")
		return nil, ErrRequiredAttributeMissing
	}
	cs, err := NewPdfColorspaceFromPdfObject(obj)
	if err != nil {
		common.Log.Debug("Failed loading colorspace: %v", err)
		return nil, err
	}
	shading.ColorSpace = cs

	// Background (optional). Array of color components.
	obj = dict.Get("Background")
	if obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Background should be specified by an array (got %T)", obj)
			return nil, ErrTypeError
		}
		shading.Background = arr
	}

	// BBox.
	obj = dict.Get("BBox")
	if obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Background should be specified by an array (got %T)", obj)
			return nil, ErrTypeError
		}
		rect, err := NewPdfRectangle(*arr)
		if err != nil {
			common.Log.Debug("BBox error: %v", err)
			return nil, err
		}
		shading.BBox = rect
	}

	// AntiAlias.
	obj = dict.Get("AntiAlias")
	if obj != nil {
		obj = TraceToDirectObject(obj)
		val, ok := obj.(*PdfObjectBool)
		if !ok {
			common.Log.Debug("AntiAlias invalid type, should be bool (got %T)", obj)
			return nil, ErrTypeError
		}
		shading.AntiAlias = val
	}

	// Load specific shading type specific entries.
	switch *shadingType {
	case 1:
		ctx, err := newPdfShadingType1FromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfShading = shading
		shading.context = ctx
		return shading, nil
	case 2:
		ctx, err := newPdfShadingType2FromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfShading = shading
		shading.context = ctx
		return shading, nil
	case 3:
		ctx, err := newPdfShadingType3FromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfShading = shading
		shading.context = ctx
		return shading, nil
	case 4:
		ctx, err := newPdfShadingType4FromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfShading = shading
		shading.context = ctx
		return shading, nil
	case 5:
		ctx, err := newPdfShadingType5FromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfShading = shading
		shading.context = ctx
		return shading, nil
	case 6:
		ctx, err := newPdfShadingType6FromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfShading = shading
		shading.context = ctx
		return shading, nil
	case 7:
		ctx, err := newPdfShadingType7FromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfShading = shading
		shading.context = ctx
		return shading, nil
	}

	return nil, errors.New("Unknown shading type")
}

// Load shading type 1 specific attributes from pdf object.  Used in parsing/loading PDFs.
func newPdfShadingType1FromDictionary(dict *PdfObjectDictionary) (*PdfShadingType1, error) {
	shading := PdfShadingType1{}

	// Domain (optional).
	if obj := dict.Get("Domain"); obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Domain not an array (got %T)", obj)
			return nil, errors.New("Type check error")
		}
		shading.Domain = arr
	}

	// Matrix (optional).
	if obj := dict.Get("Matrix"); obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Matrix not an array (got %T)", obj)
			return nil, errors.New("Type check error")
		}
		shading.Matrix = arr
	}

	// Function (required).
	obj := dict.Get("Function")
	if obj == nil {
		common.Log.Debug("Required attribute missing:  Function")
		return nil, ErrRequiredAttributeMissing
	}
	shading.Function = []PdfFunction{}
	if array, is := obj.(*PdfObjectArray); is {
		for _, obj := range *array {
			function, err := newPdfFunctionFromPdfObject(obj)
			if err != nil {
				common.Log.Debug("Error parsing function: %v", err)
				return nil, err
			}
			shading.Function = append(shading.Function, function)
		}
	} else {
		function, err := newPdfFunctionFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error parsing function: %v", err)
			return nil, err
		}
		shading.Function = append(shading.Function, function)
	}

	return &shading, nil
}

// Load shading type 2 specific attributes from pdf object.  Used in parsing/loading PDFs.
func newPdfShadingType2FromDictionary(dict *PdfObjectDictionary) (*PdfShadingType2, error) {
	shading := PdfShadingType2{}

	// Coords (required).
	obj := dict.Get("Coords")
	if obj == nil {
		common.Log.Debug("Required attribute missing:  Coords")
		return nil, ErrRequiredAttributeMissing
	}
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("Coords not an array (got %T)", obj)
		return nil, errors.New("Type check error")
	}
	if len(*arr) != 4 {
		common.Log.Debug("Coords length not 4 (got %d)", len(*arr))
		return nil, errors.New("Invalid attribute")
	}
	shading.Coords = arr

	// Domain (optional).
	if obj := dict.Get("Domain"); obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Domain not an array (got %T)", obj)
			return nil, errors.New("Type check error")
		}
		shading.Domain = arr
	}

	// Function (required).
	obj = dict.Get("Function")
	if obj == nil {
		common.Log.Debug("Required attribute missing:  Function")
		return nil, ErrRequiredAttributeMissing
	}
	shading.Function = []PdfFunction{}
	if array, is := obj.(*PdfObjectArray); is {
		for _, obj := range *array {
			function, err := newPdfFunctionFromPdfObject(obj)
			if err != nil {
				common.Log.Debug("Error parsing function: %v", err)
				return nil, err
			}
			shading.Function = append(shading.Function, function)
		}
	} else {
		function, err := newPdfFunctionFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error parsing function: %v", err)
			return nil, err
		}
		shading.Function = append(shading.Function, function)
	}

	// Extend (optional).
	if obj := dict.Get("Extend"); obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Matrix not an array (got %T)", obj)
			return nil, ErrTypeCheck
		}
		if len(*arr) != 2 {
			common.Log.Debug("Extend length not 2 (got %d)", len(*arr))
			return nil, ErrInvalidAttribute
		}
		shading.Extend = arr
	}

	return &shading, nil
}

// Load shading type 3 specific attributes from pdf object.  Used in parsing/loading PDFs.
func newPdfShadingType3FromDictionary(dict *PdfObjectDictionary) (*PdfShadingType3, error) {
	shading := PdfShadingType3{}

	// Coords (required).
	obj := dict.Get("Coords")
	if obj == nil {
		common.Log.Debug("Required attribute missing: Coords")
		return nil, ErrRequiredAttributeMissing
	}
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("Coords not an array (got %T)", obj)
		return nil, ErrTypeError
	}
	if len(*arr) != 6 {
		common.Log.Debug("Coords length not 6 (got %d)", len(*arr))
		return nil, ErrInvalidAttribute
	}
	shading.Coords = arr

	// Domain (optional).
	if obj := dict.Get("Domain"); obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Domain not an array (got %T)", obj)
			return nil, ErrTypeError
		}
		shading.Domain = arr
	}

	// Function (required).
	obj = dict.Get("Function")
	if obj == nil {
		common.Log.Debug("Required attribute missing:  Function")
		return nil, ErrRequiredAttributeMissing
	}
	shading.Function = []PdfFunction{}
	if array, is := obj.(*PdfObjectArray); is {
		for _, obj := range *array {
			function, err := newPdfFunctionFromPdfObject(obj)
			if err != nil {
				common.Log.Debug("Error parsing function: %v", err)
				return nil, err
			}
			shading.Function = append(shading.Function, function)
		}
	} else {
		function, err := newPdfFunctionFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error parsing function: %v", err)
			return nil, err
		}
		shading.Function = append(shading.Function, function)
	}

	// Extend (optional).
	if obj := dict.Get("Extend"); obj != nil {
		obj = TraceToDirectObject(obj)
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Matrix not an array (got %T)", obj)
			return nil, ErrTypeError
		}
		if len(*arr) != 2 {
			common.Log.Debug("Extend length not 2 (got %d)", len(*arr))
			return nil, ErrInvalidAttribute
		}
		shading.Extend = arr
	}

	return &shading, nil
}

// Load shading type 4 specific attributes from pdf object.  Used in parsing/loading PDFs.
func newPdfShadingType4FromDictionary(dict *PdfObjectDictionary) (*PdfShadingType4, error) {
	shading := PdfShadingType4{}

	// BitsPerCoordinate (required).
	obj := dict.Get("BitsPerCoordinate")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerCoordinate")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerCoordinate not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerCoordinate = integer

	// BitsPerComponent (required).
	obj = dict.Get("BitsPerComponent")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerComponent")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerComponent not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerComponent = integer

	// BitsPerFlag (required).
	obj = dict.Get("BitsPerFlag")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerFlag")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerFlag not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerComponent = integer

	// Decode (required).
	obj = dict.Get("Decode")
	if obj == nil {
		common.Log.Debug("Required attribute missing: Decode")
		return nil, ErrRequiredAttributeMissing
	}
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("Decode not an array (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.Decode = arr

	// Function (required).
	obj = dict.Get("Function")
	if obj == nil {
		common.Log.Debug("Required attribute missing:  Function")
		return nil, ErrRequiredAttributeMissing
	}
	shading.Function = []PdfFunction{}
	if array, is := obj.(*PdfObjectArray); is {
		for _, obj := range *array {
			function, err := newPdfFunctionFromPdfObject(obj)
			if err != nil {
				common.Log.Debug("Error parsing function: %v", err)
				return nil, err
			}
			shading.Function = append(shading.Function, function)
		}
	} else {
		function, err := newPdfFunctionFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error parsing function: %v", err)
			return nil, err
		}
		shading.Function = append(shading.Function, function)
	}

	return &shading, nil
}

// Load shading type 5 specific attributes from pdf object.  Used in parsing/loading PDFs.
func newPdfShadingType5FromDictionary(dict *PdfObjectDictionary) (*PdfShadingType5, error) {
	shading := PdfShadingType5{}

	// BitsPerCoordinate (required).
	obj := dict.Get("BitsPerCoordinate")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerCoordinate")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerCoordinate not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerCoordinate = integer

	// BitsPerComponent (required).
	obj = dict.Get("BitsPerComponent")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerComponent")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerComponent not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerComponent = integer

	// VerticesPerRow (required).
	obj = dict.Get("VerticesPerRow")
	if obj == nil {
		common.Log.Debug("Required attribute missing: VerticesPerRow")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("VerticesPerRow not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.VerticesPerRow = integer

	// Decode (required).
	obj = dict.Get("Decode")
	if obj == nil {
		common.Log.Debug("Required attribute missing: Decode")
		return nil, ErrRequiredAttributeMissing
	}
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("Decode not an array (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.Decode = arr

	// Function (optional).
	if obj := dict.Get("Function"); obj != nil {
		// Function (required).
		shading.Function = []PdfFunction{}
		if array, is := obj.(*PdfObjectArray); is {
			for _, obj := range *array {
				function, err := newPdfFunctionFromPdfObject(obj)
				if err != nil {
					common.Log.Debug("Error parsing function: %v", err)
					return nil, err
				}
				shading.Function = append(shading.Function, function)
			}
		} else {
			function, err := newPdfFunctionFromPdfObject(obj)
			if err != nil {
				common.Log.Debug("Error parsing function: %v", err)
				return nil, err
			}
			shading.Function = append(shading.Function, function)
		}
	}

	return &shading, nil
}

// Load shading type 6 specific attributes from pdf object.  Used in parsing/loading PDFs.
func newPdfShadingType6FromDictionary(dict *PdfObjectDictionary) (*PdfShadingType6, error) {
	shading := PdfShadingType6{}

	// BitsPerCoordinate (required).
	obj := dict.Get("BitsPerCoordinate")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerCoordinate")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerCoordinate not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerCoordinate = integer

	// BitsPerComponent (required).
	obj = dict.Get("BitsPerComponent")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerComponent")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerComponent not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerComponent = integer

	// BitsPerFlag (required).
	obj = dict.Get("BitsPerFlag")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerFlag")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerFlag not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerComponent = integer

	// Decode (required).
	obj = dict.Get("Decode")
	if obj == nil {
		common.Log.Debug("Required attribute missing: Decode")
		return nil, ErrRequiredAttributeMissing
	}
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("Decode not an array (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.Decode = arr

	// Function (optional).
	if obj := dict.Get("Function"); obj != nil {
		shading.Function = []PdfFunction{}
		if array, is := obj.(*PdfObjectArray); is {
			for _, obj := range *array {
				function, err := newPdfFunctionFromPdfObject(obj)
				if err != nil {
					common.Log.Debug("Error parsing function: %v", err)
					return nil, err
				}
				shading.Function = append(shading.Function, function)
			}
		} else {
			function, err := newPdfFunctionFromPdfObject(obj)
			if err != nil {
				common.Log.Debug("Error parsing function: %v", err)
				return nil, err
			}
			shading.Function = append(shading.Function, function)
		}
	}

	return &shading, nil
}

// Load shading type 7 specific attributes from pdf object.  Used in parsing/loading PDFs.
func newPdfShadingType7FromDictionary(dict *PdfObjectDictionary) (*PdfShadingType7, error) {
	shading := PdfShadingType7{}

	// BitsPerCoordinate (required).
	obj := dict.Get("BitsPerCoordinate")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerCoordinate")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerCoordinate not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerCoordinate = integer

	// BitsPerComponent (required).
	obj = dict.Get("BitsPerComponent")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerComponent")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerComponent not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerComponent = integer

	// BitsPerFlag (required).
	obj = dict.Get("BitsPerFlag")
	if obj == nil {
		common.Log.Debug("Required attribute missing: BitsPerFlag")
		return nil, ErrRequiredAttributeMissing
	}
	integer, ok = obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("BitsPerFlag not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.BitsPerComponent = integer

	// Decode (required).
	obj = dict.Get("Decode")
	if obj == nil {
		common.Log.Debug("Required attribute missing: Decode")
		return nil, ErrRequiredAttributeMissing
	}
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("Decode not an array (got %T)", obj)
		return nil, ErrTypeError
	}
	shading.Decode = arr

	// Function (optional).
	if obj := dict.Get("Function"); obj != nil {
		shading.Function = []PdfFunction{}
		if array, is := obj.(*PdfObjectArray); is {
			for _, obj := range *array {
				function, err := newPdfFunctionFromPdfObject(obj)
				if err != nil {
					common.Log.Debug("Error parsing function: %v", err)
					return nil, err
				}
				shading.Function = append(shading.Function, function)
			}
		} else {
			function, err := newPdfFunctionFromPdfObject(obj)
			if err != nil {
				common.Log.Debug("Error parsing function: %v", err)
				return nil, err
			}
			shading.Function = append(shading.Function, function)
		}
	}

	return &shading, nil
}

/* Conversion to pdf objects. */

func (this *PdfShading) ToPdfObject() PdfObject {
	container := this.container

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if this.ShadingType != nil {
		d.Set("ShadingType", this.ShadingType)
	}
	if this.ColorSpace != nil {
		d.Set("ColorSpace", this.ColorSpace.ToPdfObject())
	}
	if this.Background != nil {
		d.Set("Background", this.Background)
	}
	if this.BBox != nil {
		d.Set("BBox", this.BBox.ToPdfObject())
	}
	if this.AntiAlias != nil {
		d.Set("AntiAlias", this.AntiAlias)
	}

	return container
}

func (this *PdfShadingType1) ToPdfObject() PdfObject {
	this.PdfShading.ToPdfObject()

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if this.Domain != nil {
		d.Set("Domain", this.Domain)
	}
	if this.Matrix != nil {
		d.Set("Matrix", this.Matrix)
	}
	if this.Function != nil {
		if len(this.Function) == 1 {
			d.Set("Function", this.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range this.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return this.container
}

func (this *PdfShadingType2) ToPdfObject() PdfObject {
	this.PdfShading.ToPdfObject()

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}
	if d == nil {
		common.Log.Error("Shading dict is nil")
		return nil
	}
	if this.Coords != nil {
		d.Set("Coords", this.Coords)
	}
	if this.Domain != nil {
		d.Set("Domain", this.Domain)
	}
	if this.Function != nil {
		if len(this.Function) == 1 {
			d.Set("Function", this.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range this.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}
	if this.Extend != nil {
		d.Set("Extend", this.Extend)
	}

	return this.container
}

func (this *PdfShadingType3) ToPdfObject() PdfObject {
	this.PdfShading.ToPdfObject()

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if this.Coords != nil {
		d.Set("Coords", this.Coords)
	}
	if this.Domain != nil {
		d.Set("Domain", this.Domain)
	}
	if this.Function != nil {
		if len(this.Function) == 1 {
			d.Set("Function", this.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range this.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}
	if this.Extend != nil {
		d.Set("Extend", this.Extend)
	}

	return this.container
}

func (this *PdfShadingType4) ToPdfObject() PdfObject {
	this.PdfShading.ToPdfObject()

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if this.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", this.BitsPerCoordinate)
	}
	if this.BitsPerComponent != nil {
		d.Set("BitsPerComponent", this.BitsPerComponent)
	}
	if this.BitsPerFlag != nil {
		d.Set("BitsPerFlag", this.BitsPerFlag)
	}
	if this.Decode != nil {
		d.Set("Decode", this.Decode)
	}
	if this.Function != nil {
		if len(this.Function) == 1 {
			d.Set("Function", this.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range this.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return this.container
}

func (this *PdfShadingType5) ToPdfObject() PdfObject {
	this.PdfShading.ToPdfObject()

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if this.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", this.BitsPerCoordinate)
	}
	if this.BitsPerComponent != nil {
		d.Set("BitsPerComponent", this.BitsPerComponent)
	}
	if this.VerticesPerRow != nil {
		d.Set("VerticesPerRow", this.VerticesPerRow)
	}
	if this.Decode != nil {
		d.Set("Decode", this.Decode)
	}
	if this.Function != nil {
		if len(this.Function) == 1 {
			d.Set("Function", this.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range this.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return this.container
}

func (this *PdfShadingType6) ToPdfObject() PdfObject {
	this.PdfShading.ToPdfObject()

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if this.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", this.BitsPerCoordinate)
	}
	if this.BitsPerComponent != nil {
		d.Set("BitsPerComponent", this.BitsPerComponent)
	}
	if this.BitsPerFlag != nil {
		d.Set("BitsPerFlag", this.BitsPerFlag)
	}
	if this.Decode != nil {
		d.Set("Decode", this.Decode)
	}
	if this.Function != nil {
		if len(this.Function) == 1 {
			d.Set("Function", this.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range this.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return this.container
}

func (this *PdfShadingType7) ToPdfObject() PdfObject {
	this.PdfShading.ToPdfObject()

	d, err := this.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if this.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", this.BitsPerCoordinate)
	}
	if this.BitsPerComponent != nil {
		d.Set("BitsPerComponent", this.BitsPerComponent)
	}
	if this.BitsPerFlag != nil {
		d.Set("BitsPerFlag", this.BitsPerFlag)
	}
	if this.Decode != nil {
		d.Set("Decode", this.Decode)
	}
	if this.Function != nil {
		if len(this.Function) == 1 {
			d.Set("Function", this.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range this.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return this.container
}
