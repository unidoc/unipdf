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
	r.primitive = &PdfObjectDictionary{}
	return r
}

func NewPdfPageResourcesFromDict(dict *PdfObjectDictionary) (*PdfPageResources, error) {
	r := NewPdfPageResources()

	if obj, isDefined := (*dict)["ExtGState"]; isDefined {
		r.ExtGState = obj
	}
	if obj, isDefined := (*dict)["ColorSpace"]; isDefined && !isNullObject(obj) {
		colorspaces, err := newPdfPageResourcesColorspacesFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		r.ColorSpace = colorspaces
	}
	if obj, isDefined := (*dict)["Pattern"]; isDefined {
		r.Pattern = obj
	}
	if obj, isDefined := (*dict)["Shading"]; isDefined {
		r.Shading = obj
	}
	if obj, isDefined := (*dict)["XObject"]; isDefined {
		r.XObject = obj
	}
	if obj, isDefined := (*dict)["Font"]; isDefined {
		r.Font = obj
	}
	if obj, isDefined := (*dict)["ProcSet"]; isDefined {
		r.ProcSet = obj
	}
	if obj, isDefined := (*dict)["Properties"]; isDefined {
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
		r.ExtGState = &PdfObjectDictionary{}
	}

	obj := r.ExtGState
	dict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ExtGState type error (got %T/%T)", obj, TraceToDirectObject(obj))
		return ErrTypeError
	}

	(*dict)[gsName] = gsDict
	return nil
}

// Get the shading specified by keyName.  Returns nil if not existing. The bool flag indicated whether it was found
// or not.
func (r *PdfPageResources) GetShadingByName(keyName string) (*PdfShading, bool) {
	if r.Shading == nil {
		return nil, false
	}

	shadingDict, ok := r.Shading.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid Shading entry - not a dict (got %T)", r.Shading)
		return nil, false
	}

	if obj, has := (*shadingDict)[PdfObjectName(keyName)]; has {
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
func (r *PdfPageResources) SetShadingByName(keyName string, shadingObj PdfObject) error {
	if r.Shading == nil {
		r.Shading = &PdfObjectDictionary{}
	}

	shadingDict, has := r.Shading.(*PdfObjectDictionary)
	if !has {
		return ErrTypeError
	}

	(*shadingDict)[PdfObjectName(keyName)] = shadingObj
	return nil
}

// Get the pattern specified by keyName.  Returns nil if not existing. The bool flag indicated whether it was found
// or not.
func (r *PdfPageResources) GetPatternByName(keyName string) (*PdfPattern, bool) {
	if r.Pattern == nil {
		return nil, false
	}

	patternDict, ok := r.Pattern.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid Pattern entry - not a dict (got %T)", r.Pattern)
		return nil, false
	}

	if obj, has := (*patternDict)[PdfObjectName(keyName)]; has {
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
func (r *PdfPageResources) SetPatternByName(keyName string, pattern PdfObject) error {
	if r.Pattern == nil {
		r.Pattern = &PdfObjectDictionary{}
	}

	patternDict, has := r.Pattern.(*PdfObjectDictionary)
	if !has {
		return ErrTypeError
	}

	(*patternDict)[PdfObjectName(keyName)] = pattern
	return nil
}

// Check if an XObject with a specified keyName is defined.
func (r *PdfPageResources) HasXObjectByName(keyName string) bool {
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
func (r *PdfPageResources) GetXObjectByName(keyName string) (*PdfObjectStream, XObjectType) {
	if r.XObject == nil {
		return nil, XObjectTypeUndefined
	}

	xresDict, has := TraceToDirectObject(r.XObject).(*PdfObjectDictionary)
	if !has {
		common.Log.Debug("ERROR: XObject not a dictionary! (got %T)", TraceToDirectObject(r.XObject))
		return nil, XObjectTypeUndefined
	}

	if obj, has := (*xresDict)[PdfObjectName(keyName)]; has {
		stream, ok := obj.(*PdfObjectStream)
		if !ok {
			common.Log.Debug("XObject not pointing to a stream %T", obj)
			return nil, XObjectTypeUndefined
		}
		dict := stream.PdfObjectDictionary

		name, ok := (*dict)["Subtype"].(*PdfObjectName)
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

func (r *PdfPageResources) setXObjectByName(keyName string, stream *PdfObjectStream) error {
	if r.XObject == nil {
		r.XObject = &PdfObjectDictionary{}
	}

	obj := TraceToDirectObject(r.XObject)
	xresDict, has := obj.(*PdfObjectDictionary)
	if !has {
		common.Log.Debug("Invalid XObject, got %T/%T", r.XObject, obj)
		return errors.New("Type check error")
	}

	(*xresDict)[PdfObjectName(keyName)] = stream
	return nil
}

func (r *PdfPageResources) GetXObjectImageByName(keyName string) (*XObjectImage, error) {
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

func (r *PdfPageResources) SetXObjectImageByName(keyName string, ximg *XObjectImage) error {
	stream := ximg.ToPdfObject().(*PdfObjectStream)
	err := r.setXObjectByName(keyName, stream)
	return err
}

func (r *PdfPageResources) GetXObjectFormByName(keyName string) (*XObjectForm, error) {
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

func (r *PdfPageResources) SetXObjectFormByName(keyName string, xform *XObjectForm) error {
	stream := xform.ToPdfObject().(*PdfObjectStream)
	err := r.setXObjectByName(keyName, stream)
	return err
}
