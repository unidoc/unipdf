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

// PdfShading represents a shading dictionary. There are 7 types of shading,
// indicatedby the shading type variable:
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

func (s *PdfShading) GetContainingPdfObject() PdfObject {
	return s.container
}

// GetContext returns a reference to the subshading entry as represented by PdfShadingType1-7.
func (s *PdfShading) GetContext() PdfModel {
	return s.context
}

// SetContext set the sub annotation (context).
func (s *PdfShading) SetContext(ctx PdfModel) {
	s.context = ctx
}

func (s *PdfShading) getShadingDict() (*PdfObjectDictionary, error) {
	obj := s.container

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

// PdfShadingType1 is a Function-based shading.
type PdfShadingType1 struct {
	*PdfShading
	Domain   *PdfObjectArray
	Matrix   *PdfObjectArray
	Function []PdfFunction
}

// PdfShadingType2 is a Axial shading.
type PdfShadingType2 struct {
	*PdfShading
	Coords   *PdfObjectArray
	Domain   *PdfObjectArray
	Function []PdfFunction
	Extend   *PdfObjectArray
}

// PdfShadingType3 is a Radial shading.
type PdfShadingType3 struct {
	*PdfShading
	Coords   *PdfObjectArray
	Domain   *PdfObjectArray
	Function []PdfFunction
	Extend   *PdfObjectArray
}

// PdfShadingType4 is a Free-form Gouraud-shaded triangle mesh.
type PdfShadingType4 struct {
	*PdfShading
	BitsPerCoordinate *PdfObjectInteger
	BitsPerComponent  *PdfObjectInteger
	BitsPerFlag       *PdfObjectInteger
	Decode            *PdfObjectArray
	Function          []PdfFunction
}

// PdfShadingType5 is a Lattice-form Gouraud-shaded triangle mesh.
type PdfShadingType5 struct {
	*PdfShading
	BitsPerCoordinate *PdfObjectInteger
	BitsPerComponent  *PdfObjectInteger
	VerticesPerRow    *PdfObjectInteger
	Decode            *PdfObjectArray
	Function          []PdfFunction
}

// PdfShadingType6 is a Coons patch mesh.
type PdfShadingType6 struct {
	*PdfShading
	BitsPerCoordinate *PdfObjectInteger
	BitsPerComponent  *PdfObjectInteger
	BitsPerFlag       *PdfObjectInteger
	Decode            *PdfObjectArray
	Function          []PdfFunction
}

// PdfShadingType7 is a Tensor-product patch mesh.
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
		for _, obj := range array.Elements() {
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
	if arr.Len() != 4 {
		common.Log.Debug("Coords length not 4 (got %d)", arr.Len())
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
		for _, obj := range array.Elements() {
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
		if arr.Len() != 2 {
			common.Log.Debug("Extend length not 2 (got %d)", arr.Len())
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
	if arr.Len() != 6 {
		common.Log.Debug("Coords length not 6 (got %d)", arr.Len())
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
		for _, obj := range array.Elements() {
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
		if arr.Len() != 2 {
			common.Log.Debug("Extend length not 2 (got %d)", arr.Len())
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
		for _, obj := range array.Elements() {
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
			for _, obj := range array.Elements() {
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
			for _, obj := range array.Elements() {
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
			for _, obj := range array.Elements() {
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

func (s *PdfShading) ToPdfObject() PdfObject {
	container := s.container

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if s.ShadingType != nil {
		d.Set("ShadingType", s.ShadingType)
	}
	if s.ColorSpace != nil {
		d.Set("ColorSpace", s.ColorSpace.ToPdfObject())
	}
	if s.Background != nil {
		d.Set("Background", s.Background)
	}
	if s.BBox != nil {
		d.Set("BBox", s.BBox.ToPdfObject())
	}
	if s.AntiAlias != nil {
		d.Set("AntiAlias", s.AntiAlias)
	}

	return container
}

func (s *PdfShadingType1) ToPdfObject() PdfObject {
	s.PdfShading.ToPdfObject()

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if s.Domain != nil {
		d.Set("Domain", s.Domain)
	}
	if s.Matrix != nil {
		d.Set("Matrix", s.Matrix)
	}
	if s.Function != nil {
		if len(s.Function) == 1 {
			d.Set("Function", s.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range s.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return s.container
}

func (s *PdfShadingType2) ToPdfObject() PdfObject {
	s.PdfShading.ToPdfObject()

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}
	if d == nil {
		common.Log.Error("Shading dict is nil")
		return nil
	}
	if s.Coords != nil {
		d.Set("Coords", s.Coords)
	}
	if s.Domain != nil {
		d.Set("Domain", s.Domain)
	}
	if s.Function != nil {
		if len(s.Function) == 1 {
			d.Set("Function", s.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range s.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}
	if s.Extend != nil {
		d.Set("Extend", s.Extend)
	}

	return s.container
}

func (s *PdfShadingType3) ToPdfObject() PdfObject {
	s.PdfShading.ToPdfObject()

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if s.Coords != nil {
		d.Set("Coords", s.Coords)
	}
	if s.Domain != nil {
		d.Set("Domain", s.Domain)
	}
	if s.Function != nil {
		if len(s.Function) == 1 {
			d.Set("Function", s.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range s.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}
	if s.Extend != nil {
		d.Set("Extend", s.Extend)
	}

	return s.container
}

func (s *PdfShadingType4) ToPdfObject() PdfObject {
	s.PdfShading.ToPdfObject()

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if s.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", s.BitsPerCoordinate)
	}
	if s.BitsPerComponent != nil {
		d.Set("BitsPerComponent", s.BitsPerComponent)
	}
	if s.BitsPerFlag != nil {
		d.Set("BitsPerFlag", s.BitsPerFlag)
	}
	if s.Decode != nil {
		d.Set("Decode", s.Decode)
	}
	if s.Function != nil {
		if len(s.Function) == 1 {
			d.Set("Function", s.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range s.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return s.container
}

func (s *PdfShadingType5) ToPdfObject() PdfObject {
	s.PdfShading.ToPdfObject()

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if s.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", s.BitsPerCoordinate)
	}
	if s.BitsPerComponent != nil {
		d.Set("BitsPerComponent", s.BitsPerComponent)
	}
	if s.VerticesPerRow != nil {
		d.Set("VerticesPerRow", s.VerticesPerRow)
	}
	if s.Decode != nil {
		d.Set("Decode", s.Decode)
	}
	if s.Function != nil {
		if len(s.Function) == 1 {
			d.Set("Function", s.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range s.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return s.container
}

func (s *PdfShadingType6) ToPdfObject() PdfObject {
	s.PdfShading.ToPdfObject()

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if s.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", s.BitsPerCoordinate)
	}
	if s.BitsPerComponent != nil {
		d.Set("BitsPerComponent", s.BitsPerComponent)
	}
	if s.BitsPerFlag != nil {
		d.Set("BitsPerFlag", s.BitsPerFlag)
	}
	if s.Decode != nil {
		d.Set("Decode", s.Decode)
	}
	if s.Function != nil {
		if len(s.Function) == 1 {
			d.Set("Function", s.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range s.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return s.container
}

func (s *PdfShadingType7) ToPdfObject() PdfObject {
	s.PdfShading.ToPdfObject()

	d, err := s.getShadingDict()
	if err != nil {
		common.Log.Error("Unable to access shading dict")
		return nil
	}

	if s.BitsPerCoordinate != nil {
		d.Set("BitsPerCoordinate", s.BitsPerCoordinate)
	}
	if s.BitsPerComponent != nil {
		d.Set("BitsPerComponent", s.BitsPerComponent)
	}
	if s.BitsPerFlag != nil {
		d.Set("BitsPerFlag", s.BitsPerFlag)
	}
	if s.Decode != nil {
		d.Set("Decode", s.Decode)
	}
	if s.Function != nil {
		if len(s.Function) == 1 {
			d.Set("Function", s.Function[0].ToPdfObject())
		} else {
			farr := MakeArray()
			for _, f := range s.Function {
				farr.Append(f.ToPdfObject())
			}
			d.Set("Function", farr)
		}
	}

	return s.container
}
