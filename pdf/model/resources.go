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

// Page resources model.
// Implements PdfModel.
type PdfPageResources struct {
	ExtGState  PdfObject
	ColorSpace *PdfPageResourcesColorspaces
	Pattern    PdfObject
	Shading    PdfObject
	XObject    PdfObject
	Font       PdfObject
	ProcSet    PdfObject
	Properties PdfObject
	// Primitive reource container.
	primitive *PdfObjectDictionary
}

func NewPdfPageResources() *PdfPageResources {
	r := &PdfPageResources{}
	r.primitive = MakeDict()
	return r
}

func NewPdfPageResourcesFromDict(dict *PdfObjectDictionary) (*PdfPageResources, error) {
	r := NewPdfPageResources()

	if obj := dict.Get("ExtGState"); obj != nil {
		r.ExtGState = obj
	}
	if obj := dict.Get("ColorSpace"); obj != nil && !isNullObject(obj) {
		colorspaces, err := newPdfPageResourcesColorspacesFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		r.ColorSpace = colorspaces
	}
	if obj := dict.Get("Pattern"); obj != nil {
		r.Pattern = obj
	}
	if obj := dict.Get("Shading"); obj != nil {
		r.Shading = obj
	}
	if obj := dict.Get("XObject"); obj != nil {
		r.XObject = obj
	}
	if obj := dict.Get("Font"); obj != nil {
		r.Font = obj
	}
	if obj := dict.Get("ProcSet"); obj != nil {
		r.ProcSet = obj
	}
	if obj := dict.Get("Properties"); obj != nil {
		r.Properties = obj
	}

	return r, nil
}

func (r *PdfPageResources) GetContainingPdfObject() PdfObject {
	return r.primitive
}

func (r *PdfPageResources) ToPdfObject() PdfObject {
	d := r.primitive
	d.SetIfNotNil("ExtGState", r.ExtGState)
	if r.ColorSpace != nil {
		d.SetIfNotNil("ColorSpace", r.ColorSpace.ToPdfObject())
	}
	d.SetIfNotNil("Pattern", r.Pattern)
	d.SetIfNotNil("Shading", r.Shading)
	d.SetIfNotNil("XObject", r.XObject)
	d.SetIfNotNil("Font", r.Font)
	d.SetIfNotNil("ProcSet", r.ProcSet)
	d.SetIfNotNil("Properties", r.Properties)

	return d
}

// Add External Graphics State (GState).  The gsDict can be specified either directly as a dictionary or an indirect
// object containing a dictionary.
func (r *PdfPageResources) AddExtGState(gsName PdfObjectName, gsDict PdfObject) error {
	if r.ExtGState == nil {
		r.ExtGState = MakeDict()
	}

	obj := r.ExtGState
	dict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ExtGState type error (got %T/%T)", obj, TraceToDirectObject(obj))
		return ErrTypeError
	}

	dict.Set(gsName, gsDict)
	return nil
}

// Get the ExtGState specified by keyName.  Returns a bool indicating whether it was found or not.
func (r *PdfPageResources) GetExtGState(keyName PdfObjectName) (PdfObject, bool) {
	if r.ExtGState == nil {
		return nil, false
	}

	dict, ok := TraceToDirectObject(r.ExtGState).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid ExtGState entry - not a dict (got %T)", r.ExtGState)
		return nil, false
	}

	if obj := dict.Get(keyName); obj != nil {
		return obj, true
	} else {
		return nil, false
	}
}

// Check whether a font is defined by the specified keyName.
func (r *PdfPageResources) HasExtGState(keyName PdfObjectName) bool {
	_, has := r.GetFontByName(keyName)
	return has
}

// Get the shading specified by keyName.  Returns nil if not existing. The bool flag indicated whether it was found
// or not.
func (r *PdfPageResources) GetShadingByName(keyName PdfObjectName) (*PdfShading, bool) {
	if r.Shading == nil {
		return nil, false
	}

	shadingDict, ok := TraceToDirectObject(r.Shading).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid Shading entry - not a dict (got %T)", r.Shading)
		return nil, false
	}

	if obj := shadingDict.Get(keyName); obj != nil {
		shading, err := newPdfShadingFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("ERROR: failed to load pdf shading: %v", err)
			return nil, false
		}
		return shading, true
	} else {
		return nil, false
	}
}

// Set a shading resource specified by keyName.
func (r *PdfPageResources) SetShadingByName(keyName PdfObjectName, shadingObj PdfObject) error {
	if r.Shading == nil {
		r.Shading = MakeDict()
	}

	shadingDict, has := r.Shading.(*PdfObjectDictionary)
	if !has {
		return ErrTypeError
	}

	shadingDict.Set(keyName, shadingObj)
	return nil
}

// Get the pattern specified by keyName.  Returns nil if not existing. The bool flag indicated whether it was found
// or not.
func (r *PdfPageResources) GetPatternByName(keyName PdfObjectName) (*PdfPattern, bool) {
	if r.Pattern == nil {
		return nil, false
	}

	patternDict, ok := TraceToDirectObject(r.Pattern).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid Pattern entry - not a dict (got %T)", r.Pattern)
		return nil, false
	}

	if obj := patternDict.Get(keyName); obj != nil {
		pattern, err := newPdfPatternFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("ERROR: failed to load pdf pattern: %v", err)
			return nil, false
		}

		return pattern, true
	} else {
		return nil, false
	}
}

// Set a pattern resource specified by keyName.
func (r *PdfPageResources) SetPatternByName(keyName PdfObjectName, pattern PdfObject) error {
	if r.Pattern == nil {
		r.Pattern = MakeDict()
	}

	patternDict, has := r.Pattern.(*PdfObjectDictionary)
	if !has {
		return ErrTypeError
	}

	patternDict.Set(keyName, pattern)
	return nil
}

// Get the font specified by keyName.  Returns the PdfObject which the entry refers to.
// Returns a bool value indicating whether or not the entry was found.
func (r *PdfPageResources) GetFontByName(keyName PdfObjectName) (PdfObject, bool) {
	if r.Font == nil {
		return nil, false
	}

	fontDict, has := TraceToDirectObject(r.Font).(*PdfObjectDictionary)
	if !has {
		common.Log.Debug("ERROR: Font not a dictionary! (got %T)", TraceToDirectObject(r.Font))
		return nil, false
	}

	if obj := fontDict.Get(keyName); obj != nil {
		return obj, true
	} else {
		return nil, false
	}
}

// Check whether a font is defined by the specified keyName.
func (r *PdfPageResources) HasFontByName(keyName PdfObjectName) bool {
	_, has := r.GetFontByName(keyName)
	return has
}

// Set the font specified by keyName to the given object.
func (r *PdfPageResources) SetFontByName(keyName PdfObjectName, obj PdfObject) error {
	if r.Font == nil {
		// Create if not existing.
		r.Font = MakeDict()
	}

	fontDict, has := TraceToDirectObject(r.Font).(*PdfObjectDictionary)
	if !has {
		common.Log.Debug("ERROR: Font not a dictionary! (got %T)", TraceToDirectObject(r.Font))
		return ErrTypeError
	}

	fontDict.Set(keyName, obj)
	return nil
}

func (r *PdfPageResources) GetColorspaceByName(keyName PdfObjectName) (PdfColorspace, bool) {
	if r.ColorSpace == nil {
		return nil, false
	}

	cs, has := r.ColorSpace.Colorspaces[string(keyName)]
	if !has {
		return nil, false
	}

	return cs, true
}

func (r *PdfPageResources) HasColorspaceByName(keyName PdfObjectName) bool {
	if r.ColorSpace == nil {
		return false
	}

	_, has := r.ColorSpace.Colorspaces[string(keyName)]
	return has
}

func (r *PdfPageResources) SetColorspaceByName(keyName PdfObjectName, cs PdfColorspace) error {
	if r.ColorSpace == nil {
		r.ColorSpace = NewPdfPageResourcesColorspaces()
	}

	r.ColorSpace.Set(keyName, cs)
	return nil
}

// Check if an XObject with a specified keyName is defined.
func (r *PdfPageResources) HasXObjectByName(keyName PdfObjectName) bool {
	obj, _ := r.GetXObjectByName(keyName)
	if obj != nil {
		return true
	} else {
		return false
	}
}

type XObjectType int

const (
	XObjectTypeUndefined XObjectType = iota
	XObjectTypeImage     XObjectType = iota
	XObjectTypeForm      XObjectType = iota
	XObjectTypePS        XObjectType = iota
	XObjectTypeUnknown   XObjectType = iota
)

// Returns the XObject with the specified keyName and the object type.
func (r *PdfPageResources) GetXObjectByName(keyName PdfObjectName) (*PdfObjectStream, XObjectType) {
	if r.XObject == nil {
		return nil, XObjectTypeUndefined
	}

	xresDict, has := TraceToDirectObject(r.XObject).(*PdfObjectDictionary)
	if !has {
		common.Log.Debug("ERROR: XObject not a dictionary! (got %T)", TraceToDirectObject(r.XObject))
		return nil, XObjectTypeUndefined
	}

	if obj := xresDict.Get(keyName); obj != nil {
		stream, ok := obj.(*PdfObjectStream)
		if !ok {
			common.Log.Debug("XObject not pointing to a stream %T", obj)
			return nil, XObjectTypeUndefined
		}
		dict := stream.PdfObjectDictionary

		name, ok := TraceToDirectObject(dict.Get("Subtype")).(*PdfObjectName)
		if !ok {
			common.Log.Debug("XObject Subtype not a Name, dict: %s", dict.String())
			return nil, XObjectTypeUndefined
		}

		if *name == "Image" {
			return stream, XObjectTypeImage
		} else if *name == "Form" {
			return stream, XObjectTypeForm
		} else if *name == "PS" {
			return stream, XObjectTypePS
		} else {
			common.Log.Debug("XObject Subtype not known (%s)", *name)
			return nil, XObjectTypeUndefined
		}
	} else {
		return nil, XObjectTypeUndefined
	}
}

func (r *PdfPageResources) SetXObjectByName(keyName PdfObjectName, stream *PdfObjectStream) error {
	if r.XObject == nil {
		r.XObject = MakeDict()
	}

	obj := TraceToDirectObject(r.XObject)
	xresDict, has := obj.(*PdfObjectDictionary)
	if !has {
		common.Log.Debug("Invalid XObject, got %T/%T", r.XObject, obj)
		return errors.New("Type check error")
	}

	xresDict.Set(keyName, stream)
	return nil
}

func (r *PdfPageResources) GetXObjectImageByName(keyName PdfObjectName) (*XObjectImage, error) {
	stream, xtype := r.GetXObjectByName(keyName)
	if stream == nil {
		return nil, nil
	}
	if xtype != XObjectTypeImage {
		return nil, errors.New("Not an image")
	}

	ximg, err := NewXObjectImageFromStream(stream)
	if err != nil {
		return nil, err
	}

	return ximg, nil
}

func (r *PdfPageResources) SetXObjectImageByName(keyName PdfObjectName, ximg *XObjectImage) error {
	stream := ximg.ToPdfObject().(*PdfObjectStream)
	err := r.SetXObjectByName(keyName, stream)
	return err
}

func (r *PdfPageResources) GetXObjectFormByName(keyName PdfObjectName) (*XObjectForm, error) {
	stream, xtype := r.GetXObjectByName(keyName)
	if stream == nil {
		return nil, nil
	}
	if xtype != XObjectTypeForm {
		return nil, errors.New("Not a form")
	}

	xform, err := NewXObjectFormFromStream(stream)
	if err != nil {
		return nil, err
	}

	return xform, nil
}

func (r *PdfPageResources) SetXObjectFormByName(keyName PdfObjectName, xform *XObjectForm) error {
	stream := xform.ToPdfObject().(*PdfObjectStream)
	err := r.SetXObjectByName(keyName, stream)
	return err
}
