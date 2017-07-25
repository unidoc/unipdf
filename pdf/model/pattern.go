/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"fmt"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

// A PdfPattern can represent a Pattern, either a tiling pattern or a shading pattern.
// Note that all patterns shall be treated as colours; a Pattern colour space shall be established with the CS or cs
// operator just like other colour spaces, and a particular pattern shall be installed as the current colour with the
// SCN or scn operator.
type PdfPattern struct {
	// Type: Pattern
	PatternType int64
	context     PdfModel // The sub pattern, either PdfTilingPattern (Type 1) or PdfShadingPattern (Type 2).

	container PdfObject
}

func (this *PdfPattern) GetContainingPdfObject() PdfObject {
	return this.container
}

// Context in this case is a reference to the subpattern entry: either PdfTilingPattern or PdfShadingPattern.
func (this *PdfPattern) GetContext() PdfModel {
	return this.context
}

// Set the sub pattern (context).  Either PdfTilingPattern or PdfShadingPattern.
func (this *PdfPattern) SetContext(ctx PdfModel) {
	this.context = ctx
}

func (this *PdfPattern) IsTiling() bool {
	return this.PatternType == 1
}

func (this *PdfPattern) IsShading() bool {
	return this.PatternType == 2
}

// Check with IsTiling() prior to using this to ensure is a tiling pattern.
func (this *PdfPattern) GetAsTilingPattern() *PdfTilingPattern {
	return this.context.(*PdfTilingPattern)
}

// Check with IsShading() prior to using this, to ensure is a shading pattern.
func (this *PdfPattern) GetAsShadingPattern() *PdfShadingPattern {
	return this.context.(*PdfShadingPattern)
}

// A Tiling pattern consists of repetitions of a pattern cell with defined intervals.
// It is a type 1 pattern. (PatternType = 1).
// A tiling pattern is represented by a stream object, where the stream content is
// a content stream that describes the pattern cell.
type PdfTilingPattern struct {
	*PdfPattern
	PaintType  *PdfObjectInteger // Colored or uncolored tiling pattern.
	TilingType *PdfObjectInteger // Constant spacing, no distortion or constant spacing/faster tiling.
	BBox       *PdfRectangle
	XStep      *PdfObjectFloat
	YStep      *PdfObjectFloat
	Resources  *PdfPageResources
	Matrix     *PdfObjectArray // Pattern matrix (6 numbers).
}

func (this *PdfTilingPattern) IsColored() bool {
	if this.PaintType != nil && *this.PaintType == 1 {
		return true
	} else {
		return false
	}
}

// Get the pattern cell's content stream.
func (this *PdfTilingPattern) GetContentStream() ([]byte, error) {
	streamObj, ok := this.container.(*PdfObjectStream)
	if !ok {
		common.Log.Debug("Tiling pattern container not a stream (got %T)", this.container)
		return nil, ErrTypeError
	}

	decoded, err := DecodeStream(streamObj)
	if err != nil {
		common.Log.Debug("Failed decoding stream, err: %v", err)
		return nil, err
	}

	return decoded, nil
}

// Set the pattern cell's content stream.
func (this *PdfTilingPattern) SetContentStream(content []byte, encoder StreamEncoder) error {
	streamObj, ok := this.container.(*PdfObjectStream)
	if !ok {
		common.Log.Debug("Tiling pattern container not a stream (got %T)", this.container)
		return ErrTypeError
	}

	// If encoding is not set, use raw encoder.
	if encoder == nil {
		encoder = NewRawEncoder()
	}

	streamDict := streamObj.PdfObjectDictionary

	// Make a new stream dict based on the encoding parameters.
	encDict := encoder.MakeStreamDict()
	// Merge the encoding dict into the stream dict.
	streamDict.Merge(encDict)

	encoded, err := encoder.EncodeBytes(content)
	if err != nil {
		return err
	}

	// Update length.
	streamDict.Set("Length", MakeInteger(int64(len(encoded))))

	streamObj.Stream = []byte(encoded)

	return nil
}

// Shading patterns provide a smooth transition between colors across an area to be painted, i.e.
// color(x,y) = f(x,y) at each point.
// It is a type 2 pattern (PatternType = 2).
type PdfShadingPattern struct {
	*PdfPattern
	Shading   *PdfShading
	Matrix    *PdfObjectArray
	ExtGState PdfObject
}

// Load a pdf pattern from an indirect object. Used in parsing/loading PDFs.
func newPdfPatternFromPdfObject(container PdfObject) (*PdfPattern, error) {
	pattern := &PdfPattern{}

	var dict *PdfObjectDictionary
	if indObj, is := container.(*PdfIndirectObject); is {
		pattern.container = indObj
		d, ok := indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			common.Log.Debug("Pattern indirect object not containing dictionary (got %T)", indObj.PdfObject)
			return nil, ErrTypeError
		}
		dict = d
	} else if streamObj, is := container.(*PdfObjectStream); is {
		pattern.container = streamObj
		dict = streamObj.PdfObjectDictionary
	} else {
		common.Log.Debug("Pattern not an indirect object or stream")
		return nil, ErrTypeError
	}

	// PatternType.
	obj := dict.Get("PatternType")
	if obj == nil {
		common.Log.Debug("Pdf Pattern not containing PatternType")
		return nil, ErrRequiredAttributeMissing
	}
	patternType, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("Pattern type not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	if *patternType != 1 && *patternType != 2 {
		common.Log.Debug("Pattern type != 1/2 (got %d)", *patternType)
		return nil, ErrRangeError
	}
	pattern.PatternType = int64(*patternType)

	switch *patternType {
	case 1: // Tiling pattern.
		ctx, err := newPdfTilingPatternFromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfPattern = pattern
		pattern.context = ctx
		return pattern, nil
	case 2: // Shading pattern.
		ctx, err := newPdfShadingPatternFromDictionary(dict)
		if err != nil {
			return nil, err
		}
		ctx.PdfPattern = pattern
		pattern.context = ctx
		return pattern, nil
	}

	return nil, errors.New("Unknown pattern")
}

// Load entries specific to a pdf tiling pattern from a dictionary. Used in parsing/loading PDFs.
func newPdfTilingPatternFromDictionary(dict *PdfObjectDictionary) (*PdfTilingPattern, error) {
	pattern := &PdfTilingPattern{}

	// PaintType (required).
	obj := dict.Get("PaintType")
	if obj == nil {
		common.Log.Debug("PaintType missing")
		return nil, ErrRequiredAttributeMissing
	}
	paintType, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("PaintType not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	pattern.PaintType = paintType

	// TilingType (required).
	obj = dict.Get("TilingType")
	if obj == nil {
		common.Log.Debug("TilingType missing")
		return nil, ErrRequiredAttributeMissing
	}
	tilingType, ok := obj.(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("TilingType not an integer (got %T)", obj)
		return nil, ErrTypeError
	}
	pattern.TilingType = tilingType

	// BBox (required).
	obj = dict.Get("BBox")
	if obj == nil {
		common.Log.Debug("BBox missing")
		return nil, ErrRequiredAttributeMissing
	}
	obj = TraceToDirectObject(obj)
	arr, ok := obj.(*PdfObjectArray)
	if !ok {
		common.Log.Debug("BBox should be specified by an array (got %T)", obj)
		return nil, ErrTypeError
	}
	rect, err := NewPdfRectangle(*arr)
	if err != nil {
		common.Log.Debug("BBox error: %v", err)
		return nil, err
	}
	pattern.BBox = rect

	// XStep (required).
	obj = dict.Get("XStep")
	if obj == nil {
		common.Log.Debug("XStep missing")
		return nil, ErrRequiredAttributeMissing
	}
	xStep, err := getNumberAsFloat(obj)
	if err != nil {
		common.Log.Debug("Error getting XStep as float: %v", xStep)
		return nil, err
	}
	pattern.XStep = MakeFloat(xStep)

	// YStep (required).
	obj = dict.Get("YStep")
	if obj == nil {
		common.Log.Debug("YStep missing")
		return nil, ErrRequiredAttributeMissing
	}
	yStep, err := getNumberAsFloat(obj)
	if err != nil {
		common.Log.Debug("Error getting YStep as float: %v", yStep)
		return nil, err
	}
	pattern.YStep = MakeFloat(yStep)

	// Resources (required).
	obj = dict.Get("Resources")
	if obj == nil {
		common.Log.Debug("Resources missing")
		return nil, ErrRequiredAttributeMissing
	}
	dict, ok = TraceToDirectObject(obj).(*PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("Invalid resource dictionary (%T)", obj)
	}
	resources, err := NewPdfPageResourcesFromDict(dict)
	if err != nil {
		return nil, err
	}
	pattern.Resources = resources

	// Matrix (optional).
	if obj := dict.Get("Matrix"); obj != nil {
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Matrix not an array (got %T)", obj)
			return nil, ErrTypeError
		}
		pattern.Matrix = arr
	}

	return pattern, nil
}

// Load entries specific to a pdf shading pattern from a dictionary. Used in parsing/loading PDFs.
func newPdfShadingPatternFromDictionary(dict *PdfObjectDictionary) (*PdfShadingPattern, error) {
	pattern := &PdfShadingPattern{}

	// Shading (required).
	obj := dict.Get("Shading")
	if obj == nil {
		common.Log.Debug("Shading missing")
		return nil, ErrRequiredAttributeMissing
	}
	shading, err := newPdfShadingFromPdfObject(obj)
	if err != nil {
		common.Log.Debug("Error loading shading: %v", err)
		return nil, err
	}
	pattern.Shading = shading

	// Matrix (optional).
	if obj := dict.Get("Matrix"); obj != nil {
		arr, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Matrix not an array (got %T)", obj)
			return nil, ErrTypeError
		}
		pattern.Matrix = arr
	}

	// ExtGState (optional).
	if obj := dict.Get("ExtGState"); obj != nil {
		pattern.ExtGState = obj
	}

	return pattern, nil
}

/* Conversions to pdf objects. */

func (this *PdfPattern) getDict() *PdfObjectDictionary {
	if indObj, is := this.container.(*PdfIndirectObject); is {
		dict, ok := indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			return nil
		}
		return dict
	} else if streamObj, is := this.container.(*PdfObjectStream); is {
		return streamObj.PdfObjectDictionary
	} else {
		common.Log.Debug("Trying to access pattern dictionary of invalid object type (%T)", this.container)
		return nil
	}
}

func (this *PdfPattern) ToPdfObject() PdfObject {
	d := this.getDict()
	d.Set("Type", MakeName("Pattern"))
	d.Set("PatternType", MakeInteger(this.PatternType))

	return this.container
}

func (this *PdfTilingPattern) ToPdfObject() PdfObject {
	this.PdfPattern.ToPdfObject()

	d := this.getDict()
	if this.PaintType != nil {
		d.Set("PaintType", this.PaintType)
	}
	if this.TilingType != nil {
		d.Set("TilingType", this.TilingType)
	}
	if this.BBox != nil {
		d.Set("BBox", this.BBox.ToPdfObject())
	}
	if this.XStep != nil {
		d.Set("XStep", this.XStep)
	}
	if this.YStep != nil {
		d.Set("YStep", this.YStep)
	}
	if this.Resources != nil {
		d.Set("Resources", this.Resources.ToPdfObject())
	}
	if this.Matrix != nil {
		d.Set("Matrix", this.Matrix)
	}

	return this.container
}

func (this *PdfShadingPattern) ToPdfObject() PdfObject {
	this.PdfPattern.ToPdfObject()
	d := this.getDict()

	if this.Shading != nil {
		d.Set("Shading", this.Shading.ToPdfObject())
	}
	if this.Matrix != nil {
		d.Set("Matrix", this.Matrix)
	}
	if this.ExtGState != nil {
		d.Set("ExtGState", this.ExtGState)
	}

	return this.container
}
