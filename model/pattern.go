/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"fmt"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// A PdfPattern can represent a Pattern, either a tiling pattern or a shading pattern.
// Note that all patterns shall be treated as colours; a Pattern colour space shall be established with the CS or cs
// operator just like other colour spaces, and a particular pattern shall be installed as the current colour with the
// SCN or scn operator.
type PdfPattern struct {
	// Type: Pattern
	PatternType int64
	context     PdfModel // The sub pattern, either PdfTilingPattern (Type 1) or PdfShadingPattern (Type 2).

	container core.PdfObject
}

// GetContainingPdfObject returns the container of the pattern object (indirect object).
func (p *PdfPattern) GetContainingPdfObject() core.PdfObject {
	return p.container
}

// GetContext returns a reference to the subpattern entry: either PdfTilingPattern or PdfShadingPattern.
func (p *PdfPattern) GetContext() PdfModel {
	return p.context
}

// SetContext sets the sub pattern (context).  Either PdfTilingPattern or PdfShadingPattern.
func (p *PdfPattern) SetContext(ctx PdfModel) {
	p.context = ctx
}

// IsTiling specifies if the pattern is a tiling pattern.
func (p *PdfPattern) IsTiling() bool {
	return p.PatternType == 1
}

// IsShading specifies if the pattern is a shading pattern.
func (p *PdfPattern) IsShading() bool {
	return p.PatternType == 2
}

// GetAsTilingPattern returns a tiling pattern. Check with IsTiling() prior to using this.
func (p *PdfPattern) GetAsTilingPattern() *PdfTilingPattern {
	return p.context.(*PdfTilingPattern)
}

// GetAsShadingPattern returns a shading pattern. Check with IsShading() prior to using this.
func (p *PdfPattern) GetAsShadingPattern() *PdfShadingPattern {
	return p.context.(*PdfShadingPattern)
}

// PdfTilingPattern is a Tiling pattern that consists of repetitions of a pattern cell with defined intervals.
// It is a type 1 pattern. (PatternType = 1).
// A tiling pattern is represented by a stream object, where the stream content is
// a content stream that describes the pattern cell.
type PdfTilingPattern struct {
	*PdfPattern
	PaintType  *core.PdfObjectInteger // Colored or uncolored tiling pattern.
	TilingType *core.PdfObjectInteger // Constant spacing, no distortion or constant spacing/faster tiling.
	BBox       *PdfRectangle
	XStep      *core.PdfObjectFloat
	YStep      *core.PdfObjectFloat
	Resources  *PdfPageResources
	Matrix     *core.PdfObjectArray // Pattern matrix (6 numbers).
}

// IsColored specifies if the pattern is colored.
func (p *PdfTilingPattern) IsColored() bool {
	if p.PaintType != nil && *p.PaintType == 1 {
		return true
	}

	return false
}

// GetContentStream returns the pattern cell's content stream
func (p *PdfTilingPattern) GetContentStream() ([]byte, error) {
	decoded, _, err := p.GetContentStreamWithEncoder()
	return decoded, err
}

// GetContentStreamWithEncoder returns the pattern cell's content stream and its encoder
func (p *PdfTilingPattern) GetContentStreamWithEncoder() ([]byte, core.StreamEncoder, error) {
	streamObj, ok := p.container.(*core.PdfObjectStream)
	if !ok {
		common.Log.Debug("Tiling pattern container not a stream (got %T)", p.container)
		return nil, nil, core.ErrTypeError
	}

	decoded, err := core.DecodeStream(streamObj)
	if err != nil {
		common.Log.Debug("Failed decoding stream, err: %v", err)
		return nil, nil, err
	}

	encoder, err := core.NewEncoderFromStream(streamObj)
	if err != nil {
		common.Log.Debug("Failed finding decoding encoder: %v", err)
		return nil, nil, err
	}

	return decoded, encoder, nil
}

// SetContentStream sets the pattern cell's content stream.
func (p *PdfTilingPattern) SetContentStream(content []byte, encoder core.StreamEncoder) error {
	streamObj, ok := p.container.(*core.PdfObjectStream)
	if !ok {
		common.Log.Debug("Tiling pattern container not a stream (got %T)", p.container)
		return core.ErrTypeError
	}

	// If encoding is not set, use raw encoder.
	if encoder == nil {
		encoder = core.NewRawEncoder()
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
	streamDict.Set("Length", core.MakeInteger(int64(len(encoded))))

	streamObj.Stream = []byte(encoded)

	return nil
}

// PdfShadingPattern is a Shading patterns that provide a smooth transition between colors across an area to be painted,
// i.e. color(x,y) = f(x,y) at each point.
// It is a type 2 pattern (PatternType = 2).
type PdfShadingPattern struct {
	*PdfPattern
	Shading   *PdfShading
	Matrix    *core.PdfObjectArray
	ExtGState core.PdfObject
}

// Load a pdf pattern from an indirect object. Used in parsing/loading PDFs.
func newPdfPatternFromPdfObject(container core.PdfObject) (*PdfPattern, error) {
	pattern := &PdfPattern{}

	var dict *core.PdfObjectDictionary
	if indObj, is := core.GetIndirect(container); is {
		pattern.container = indObj
		d, ok := indObj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			common.Log.Debug("Pattern indirect object not containing dictionary (got %T)", indObj.PdfObject)
			return nil, core.ErrTypeError
		}
		dict = d
	} else if streamObj, is := core.GetStream(container); is {
		pattern.container = streamObj
		dict = streamObj.PdfObjectDictionary
	} else {
		common.Log.Debug("Pattern not an indirect object or stream. %T", container)
		return nil, core.ErrTypeError
	}

	// PatternType.
	obj := dict.Get("PatternType")
	if obj == nil {
		common.Log.Debug("Pdf Pattern not containing PatternType")
		return nil, ErrRequiredAttributeMissing
	}
	patternType, ok := obj.(*core.PdfObjectInteger)
	if !ok {
		common.Log.Debug("Pattern type not an integer (got %T)", obj)
		return nil, core.ErrTypeError
	}
	if *patternType != 1 && *patternType != 2 {
		common.Log.Debug("Pattern type != 1/2 (got %d)", *patternType)
		return nil, core.ErrRangeError
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

	return nil, errors.New("unknown pattern")
}

// Load entries specific to a pdf tiling pattern from a dictionary.
// Used in parsing/loading PDFs.
func newPdfTilingPatternFromDictionary(dict *core.PdfObjectDictionary) (*PdfTilingPattern, error) {
	pattern := &PdfTilingPattern{}

	// PaintType (required).
	obj := dict.Get("PaintType")
	if obj == nil {
		common.Log.Debug("PaintType missing")
		return nil, ErrRequiredAttributeMissing
	}
	paintType, ok := obj.(*core.PdfObjectInteger)
	if !ok {
		common.Log.Debug("PaintType not an integer (got %T)", obj)
		return nil, core.ErrTypeError
	}
	pattern.PaintType = paintType

	// TilingType (required).
	obj = dict.Get("TilingType")
	if obj == nil {
		common.Log.Debug("TilingType missing")
		return nil, ErrRequiredAttributeMissing
	}
	tilingType, ok := obj.(*core.PdfObjectInteger)
	if !ok {
		common.Log.Debug("TilingType not an integer (got %T)", obj)
		return nil, core.ErrTypeError
	}
	pattern.TilingType = tilingType

	// BBox (required).
	obj = dict.Get("BBox")
	if obj == nil {
		common.Log.Debug("BBox missing")
		return nil, ErrRequiredAttributeMissing
	}
	obj = core.TraceToDirectObject(obj)
	arr, ok := obj.(*core.PdfObjectArray)
	if !ok {
		common.Log.Debug("BBox should be specified by an array (got %T)", obj)
		return nil, core.ErrTypeError
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
	xStep, err := core.GetNumberAsFloat(obj)
	if err != nil {
		common.Log.Debug("Error getting XStep as float: %v", xStep)
		return nil, err
	}
	pattern.XStep = core.MakeFloat(xStep)

	// YStep (required).
	obj = dict.Get("YStep")
	if obj == nil {
		common.Log.Debug("YStep missing")
		return nil, ErrRequiredAttributeMissing
	}
	yStep, err := core.GetNumberAsFloat(obj)
	if err != nil {
		common.Log.Debug("Error getting YStep as float: %v", yStep)
		return nil, err
	}
	pattern.YStep = core.MakeFloat(yStep)

	// Resources (required).
	obj = dict.Get("Resources")
	if obj == nil {
		common.Log.Debug("Resources missing")
		return nil, ErrRequiredAttributeMissing
	}
	dict, ok = core.TraceToDirectObject(obj).(*core.PdfObjectDictionary)
	if !ok {
		return nil, fmt.Errorf("invalid resource dictionary (%T)", obj)
	}
	resources, err := NewPdfPageResourcesFromDict(dict)
	if err != nil {
		return nil, err
	}
	pattern.Resources = resources

	// Matrix (optional).
	if obj := dict.Get("Matrix"); obj != nil {
		arr, ok := obj.(*core.PdfObjectArray)
		if !ok {
			common.Log.Debug("Matrix not an array (got %T)", obj)
			return nil, core.ErrTypeError
		}
		pattern.Matrix = arr
	}

	return pattern, nil
}

// Load entries specific to a pdf shading pattern from a dictionary.
// Used in parsing/loading PDFs.
func newPdfShadingPatternFromDictionary(dict *core.PdfObjectDictionary) (*PdfShadingPattern, error) {
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
		arr, ok := obj.(*core.PdfObjectArray)
		if !ok {
			common.Log.Debug("Matrix not an array (got %T)", obj)
			return nil, core.ErrTypeError
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

func (p *PdfPattern) getDict() *core.PdfObjectDictionary {
	if indObj, is := p.container.(*core.PdfIndirectObject); is {
		dict, ok := indObj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			return nil
		}
		return dict
	} else if streamObj, is := p.container.(*core.PdfObjectStream); is {
		return streamObj.PdfObjectDictionary
	} else {
		common.Log.Debug("Trying to access pattern dictionary of invalid object type (%T)", p.container)
		return nil
	}
}

// ToPdfObject returns the PDF representation of the pattern.
func (p *PdfPattern) ToPdfObject() core.PdfObject {
	d := p.getDict()
	d.Set("Type", core.MakeName("Pattern"))
	d.Set("PatternType", core.MakeInteger(p.PatternType))

	return p.container
}

// ToPdfObject returns the PDF representation of the tiling pattern.
func (p *PdfTilingPattern) ToPdfObject() core.PdfObject {
	p.PdfPattern.ToPdfObject()

	d := p.getDict()
	if p.PaintType != nil {
		d.Set("PaintType", p.PaintType)
	}
	if p.TilingType != nil {
		d.Set("TilingType", p.TilingType)
	}
	if p.BBox != nil {
		d.Set("BBox", p.BBox.ToPdfObject())
	}
	if p.XStep != nil {
		d.Set("XStep", p.XStep)
	}
	if p.YStep != nil {
		d.Set("YStep", p.YStep)
	}
	if p.Resources != nil {
		d.Set("Resources", p.Resources.ToPdfObject())
	}
	if p.Matrix != nil {
		d.Set("Matrix", p.Matrix)
	}

	return p.container
}

// ToPdfObject returns the PDF representation of the shading pattern.
func (p *PdfShadingPattern) ToPdfObject() core.PdfObject {
	p.PdfPattern.ToPdfObject()
	d := p.getDict()

	if p.Shading != nil {
		d.Set("Shading", p.Shading.ToPdfObject())
	}
	if p.Matrix != nil {
		d.Set("Matrix", p.Matrix)
	}
	if p.ExtGState != nil {
		d.Set("ExtGState", p.ExtGState)
	}

	return p.container
}
