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

// PdfPageResources is a Page resources model.
// Implements PdfModel.
type PdfPageResources struct {
	ExtGState  core.PdfObject
	ColorSpace core.PdfObject
	Pattern    core.PdfObject
	Shading    core.PdfObject
	XObject    core.PdfObject
	Font       core.PdfObject
	ProcSet    core.PdfObject
	Properties core.PdfObject
	// Primitive resource container.
	primitive *core.PdfObjectDictionary

	// Loaded objects.
	colorspace *PdfPageResourcesColorspaces
}

// NewPdfPageResources returns a new PdfPageResources object.
func NewPdfPageResources() *PdfPageResources {
	r := &PdfPageResources{}
	r.primitive = core.MakeDict()
	return r
}

// NewPdfPageResourcesFromDict creates and returns a new PdfPageResources object
// from the input dictionary.
func NewPdfPageResourcesFromDict(dict *core.PdfObjectDictionary) (*PdfPageResources, error) {
	r := NewPdfPageResources()

	if obj := dict.Get("ExtGState"); obj != nil {
		r.ExtGState = obj
	}
	if obj := dict.Get("ColorSpace"); obj != nil && !core.IsNullObject(obj) {
		r.ColorSpace = obj
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
	if obj := core.ResolveReference(dict.Get("Font")); obj != nil {
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

// GetColorspaces loads PdfPageResourcesColorspaces from `r.ColorSpace` and returns an error if there
// is a problem loading. Once loaded, the same object is returned on multiple calls.
func (r *PdfPageResources) GetColorspaces() (*PdfPageResourcesColorspaces, error) {
	if r.colorspace != nil {
		return r.colorspace, nil
	}
	if r.ColorSpace == nil {
		return nil, nil
	}

	colorspaces, err := newPdfPageResourcesColorspacesFromPdfObject(r.ColorSpace)
	if err != nil {
		return nil, err
	}
	r.colorspace = colorspaces
	return r.colorspace, nil
}

// SetColorSpace sets `r` colorspace object to `colorspace`.
func (r *PdfPageResources) SetColorSpace(colorspace *PdfPageResourcesColorspaces) {
	r.colorspace = colorspace
}

// GetContainingPdfObject returns the container of the resources object (indirect object).
func (r *PdfPageResources) GetContainingPdfObject() core.PdfObject {
	return r.primitive
}

// ToPdfObject returns the PDF representation of the page resources.
func (r *PdfPageResources) ToPdfObject() core.PdfObject {
	d := r.primitive
	d.SetIfNotNil("ExtGState", r.ExtGState)
	if r.colorspace != nil {
		r.ColorSpace = r.colorspace.ToPdfObject()
	}
	d.SetIfNotNil("ColorSpace", r.ColorSpace)
	d.SetIfNotNil("Pattern", r.Pattern)
	d.SetIfNotNil("Shading", r.Shading)
	d.SetIfNotNil("XObject", r.XObject)
	d.SetIfNotNil("Font", r.Font)
	d.SetIfNotNil("ProcSet", r.ProcSet)
	d.SetIfNotNil("Properties", r.Properties)

	return d
}

// AddExtGState add External Graphics State (GState). The gsDict can be specified
// either directly as a dictionary or an indirect object containing a dictionary.
func (r *PdfPageResources) AddExtGState(gsName core.PdfObjectName, gsDict core.PdfObject) error {
	if r.ExtGState == nil {
		r.ExtGState = core.MakeDict()
	}

	obj := r.ExtGState
	dict, ok := core.TraceToDirectObject(obj).(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ExtGState type error (got %T/%T)", obj, core.TraceToDirectObject(obj))
		return core.ErrTypeError
	}

	dict.Set(gsName, gsDict)
	return nil
}

// GetExtGState gets the ExtGState specified by keyName. Returns a bool
// indicating whether it was found or not.
func (r *PdfPageResources) GetExtGState(keyName core.PdfObjectName) (core.PdfObject, bool) {
	if r.ExtGState == nil {
		return nil, false
	}

	dict, ok := core.TraceToDirectObject(r.ExtGState).(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid ExtGState entry - not a dict (got %T)", r.ExtGState)
		return nil, false
	}
	if obj := dict.Get(keyName); obj != nil {
		return obj, true
	}

	return nil, false
}

// HasExtGState checks whether a font is defined by the specified keyName.
func (r *PdfPageResources) HasExtGState(keyName core.PdfObjectName) bool {
	_, has := r.GetFontByName(keyName)
	return has
}

// GetShadingByName gets the shading specified by keyName. Returns nil if not existing.
// The bool flag indicated whether it was found or not.
func (r *PdfPageResources) GetShadingByName(keyName core.PdfObjectName) (*PdfShading, bool) {
	if r.Shading == nil {
		return nil, false
	}

	shadingDict, ok := core.TraceToDirectObject(r.Shading).(*core.PdfObjectDictionary)
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
	}

	return nil, false
}

// SetShadingByName sets a shading resource specified by keyName.
func (r *PdfPageResources) SetShadingByName(keyName core.PdfObjectName, shadingObj core.PdfObject) error {
	if r.Shading == nil {
		r.Shading = core.MakeDict()
	}

	shadingDict, has := r.Shading.(*core.PdfObjectDictionary)
	if !has {
		return core.ErrTypeError
	}

	shadingDict.Set(keyName, shadingObj)
	return nil
}

// GetPatternByName gets the pattern specified by keyName. Returns nil if not existing.
// The bool flag indicated whether it was found or not.
func (r *PdfPageResources) GetPatternByName(keyName core.PdfObjectName) (*PdfPattern, bool) {
	if r.Pattern == nil {
		return nil, false
	}

	patternDict, ok := core.TraceToDirectObject(r.Pattern).(*core.PdfObjectDictionary)
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
	}

	return nil, false
}

// SetPatternByName sets a pattern resource specified by keyName.
func (r *PdfPageResources) SetPatternByName(keyName core.PdfObjectName, pattern core.PdfObject) error {
	if r.Pattern == nil {
		r.Pattern = core.MakeDict()
	}

	patternDict, has := r.Pattern.(*core.PdfObjectDictionary)
	if !has {
		return core.ErrTypeError
	}

	patternDict.Set(keyName, pattern)
	return nil
}

// GetFontByName gets the font specified by keyName. Returns the PdfObject which
// the entry refers to. Returns a bool value indicating whether or not the entry was found.
func (r *PdfPageResources) GetFontByName(keyName core.PdfObjectName) (core.PdfObject, bool) {
	if r.Font == nil {
		return nil, false
	}

	fontDict, has := core.TraceToDirectObject(r.Font).(*core.PdfObjectDictionary)
	if !has {
		common.Log.Debug("ERROR: Font not a dictionary! (got %T)", core.TraceToDirectObject(r.Font))
		return nil, false
	}
	if obj := fontDict.Get(keyName); obj != nil {
		return obj, true
	}

	return nil, false
}

// HasFontByName checks whether a font is defined by the specified keyName.
func (r *PdfPageResources) HasFontByName(keyName core.PdfObjectName) bool {
	_, has := r.GetFontByName(keyName)
	return has
}

// SetFontByName sets the font specified by keyName to the given object.
func (r *PdfPageResources) SetFontByName(keyName core.PdfObjectName, obj core.PdfObject) error {
	if r.Font == nil {
		// Create if not existing.
		r.Font = core.MakeDict()
	}

	fontDict, has := core.TraceToDirectObject(r.Font).(*core.PdfObjectDictionary)
	if !has {
		common.Log.Debug("ERROR: Font not a dictionary! (got %T)", core.TraceToDirectObject(r.Font))
		return core.ErrTypeError
	}

	fontDict.Set(keyName, obj)
	return nil
}

// GetColorspaceByName returns the colorspace with the specified name from the page resources.
func (r *PdfPageResources) GetColorspaceByName(keyName core.PdfObjectName) (PdfColorspace, bool) {
	colorspace, err := r.GetColorspaces()
	if err != nil {
		common.Log.Debug("ERROR getting colorsprace: %v", err)
		return nil, false
	}

	if colorspace == nil {
		return nil, false
	}

	cs, has := colorspace.Colorspaces[string(keyName)]
	if !has {
		return nil, false
	}

	return cs, true
}

// HasColorspaceByName checks if the colorspace with the specified name exists in the page resources.
func (r *PdfPageResources) HasColorspaceByName(keyName core.PdfObjectName) bool {
	colorspace, err := r.GetColorspaces()
	if err != nil {
		common.Log.Debug("ERROR getting colorsprace: %v", err)
		return false
	}
	if colorspace == nil {
		return false
	}

	_, has := colorspace.Colorspaces[string(keyName)]
	return has
}

// SetColorspaceByName adds the provided colorspace to the page resources.
func (r *PdfPageResources) SetColorspaceByName(keyName core.PdfObjectName, cs PdfColorspace) error {
	colorspace, err := r.GetColorspaces()
	if err != nil {
		common.Log.Debug("ERROR getting colorsprace: %v", err)
		return err
	}
	if colorspace == nil {
		colorspace = NewPdfPageResourcesColorspaces()
		r.SetColorSpace(colorspace)
	}

	colorspace.Set(keyName, cs)
	return nil
}

// HasXObjectByName checks if an XObject with a specified keyName is defined.
func (r *PdfPageResources) HasXObjectByName(keyName core.PdfObjectName) bool {
	obj, _ := r.GetXObjectByName(keyName)
	return obj != nil
}

// GenerateXObjectName generates an unused XObject name that can be used for
// adding new XObjects. Uses format XObj1, XObj2, ...
func (r *PdfPageResources) GenerateXObjectName() core.PdfObjectName {
	num := 1
	for {
		name := core.MakeName(fmt.Sprintf("XObj%d", num))
		if !r.HasXObjectByName(*name) {
			return *name
		}
		num++
	}
	// Not reached.
}

// XObjectType represents the type of an XObject.
type XObjectType int

// XObject types.
const (
	XObjectTypeUndefined XObjectType = iota
	XObjectTypeImage
	XObjectTypeForm
	XObjectTypePS
	XObjectTypeUnknown
)

// GetXObjectByName returns the XObject with the specified keyName and the object type.
func (r *PdfPageResources) GetXObjectByName(keyName core.PdfObjectName) (*core.PdfObjectStream, XObjectType) {
	if r.XObject == nil {
		return nil, XObjectTypeUndefined
	}

	xresDict, has := core.TraceToDirectObject(r.XObject).(*core.PdfObjectDictionary)
	if !has {
		common.Log.Debug("ERROR: XObject not a dictionary! (got %T)", core.TraceToDirectObject(r.XObject))
		return nil, XObjectTypeUndefined
	}

	if obj := xresDict.Get(keyName); obj != nil {
		stream, ok := core.GetStream(obj)
		if !ok {
			common.Log.Debug("XObject not pointing to a stream %T", obj)
			return nil, XObjectTypeUndefined
		}
		dict := stream.PdfObjectDictionary

		name, ok := core.TraceToDirectObject(dict.Get("Subtype")).(*core.PdfObjectName)
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

// SetXObjectByName adds the XObject from the passed in stream to the page resources.
// The added XObject is identified by the specified name.
func (r *PdfPageResources) SetXObjectByName(keyName core.PdfObjectName, stream *core.PdfObjectStream) error {
	if r.XObject == nil {
		r.XObject = core.MakeDict()
	}

	obj := core.TraceToDirectObject(r.XObject)
	xresDict, has := obj.(*core.PdfObjectDictionary)
	if !has {
		common.Log.Debug("Invalid XObject, got %T/%T", r.XObject, obj)
		return errors.New("type check error")
	}

	xresDict.Set(keyName, stream)
	return nil
}

// GetXObjectImageByName returns the XObjectImage with the specified name from the
// page resources, if it exists.
func (r *PdfPageResources) GetXObjectImageByName(keyName core.PdfObjectName) (*XObjectImage, error) {
	stream, xtype := r.GetXObjectByName(keyName)
	if stream == nil {
		return nil, nil
	}
	if xtype != XObjectTypeImage {
		return nil, errors.New("not an image")
	}

	ximg, err := NewXObjectImageFromStream(stream)
	if err != nil {
		return nil, err
	}

	return ximg, nil
}

// SetXObjectImageByName adds the provided XObjectImage to the page resources.
// The added XObjectImage is identified by the specified name.
func (r *PdfPageResources) SetXObjectImageByName(keyName core.PdfObjectName, ximg *XObjectImage) error {
	stream := ximg.ToPdfObject().(*core.PdfObjectStream)
	err := r.SetXObjectByName(keyName, stream)
	return err
}

// GetXObjectFormByName returns the XObjectForm with the specified name from the
// page resources, if it exists.
func (r *PdfPageResources) GetXObjectFormByName(keyName core.PdfObjectName) (*XObjectForm, error) {
	stream, xtype := r.GetXObjectByName(keyName)
	if stream == nil {
		return nil, nil
	}
	if xtype != XObjectTypeForm {
		return nil, errors.New("not a form")
	}

	xform, err := NewXObjectFormFromStream(stream)
	if err != nil {
		return nil, err
	}

	return xform, nil
}

// SetXObjectFormByName adds the provided XObjectForm to the page resources.
// The added XObjectForm is identified by the specified name.
func (r *PdfPageResources) SetXObjectFormByName(keyName core.PdfObjectName, xform *XObjectForm) error {
	stream := xform.ToPdfObject().(*core.PdfObjectStream)
	err := r.SetXObjectByName(keyName, stream)
	return err
}
