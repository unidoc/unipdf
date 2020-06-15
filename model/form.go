/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"fmt"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

/*
FT = Btn, Tx, Ch, Sig
Btn = button
Tx = text
Ch = choice
Sig = signature
*/

// PdfAcroForm represents the AcroForm dictionary used for representation of form data in PDF.
type PdfAcroForm struct {
	Fields          *[]*PdfField
	NeedAppearances *core.PdfObjectBool
	SigFlags        *core.PdfObjectInteger
	CO              *core.PdfObjectArray
	DR              *PdfPageResources
	DA              *core.PdfObjectString
	Q               *core.PdfObjectInteger
	XFA             core.PdfObject

	container *core.PdfIndirectObject
}

// NewPdfAcroForm returns a new PdfAcroForm with an intialized container (indirect object).
func NewPdfAcroForm() *PdfAcroForm {
	return &PdfAcroForm{
		Fields:    &[]*PdfField{},
		container: core.MakeIndirectObject(core.MakeDict()),
	}
}

// flattenFields returns a flattened list of field hierarchy.
func flattenFields(field *PdfField) []*PdfField {
	list := []*PdfField{field}
	for _, k := range field.Kids {
		list = append(list, flattenFields(k)...)
	}
	return list
}

// AllFields returns a flattened list of all fields in the form.
func (form *PdfAcroForm) AllFields() []*PdfField {
	if form == nil {
		return nil
	}
	var fields []*PdfField
	if form.Fields != nil {
		for _, field := range *form.Fields {
			fields = append(fields, flattenFields(field)...)
		}
	}
	return fields
}

// signatureFields returns a slice of all signature fields in the form.
func (form *PdfAcroForm) signatureFields() []*PdfFieldSignature {
	var sigfields []*PdfFieldSignature

	for _, f := range form.AllFields() {
		switch t := f.GetContext().(type) {
		case *PdfFieldSignature:
			sigf := t
			sigfields = append(sigfields, sigf)
		}
	}

	return sigfields
}

// newPdfAcroFormFromDict is used when loading forms from PDF files.
func (r *PdfReader) newPdfAcroFormFromDict(container *core.PdfIndirectObject, d *core.PdfObjectDictionary) (*PdfAcroForm, error) {
	acroForm := NewPdfAcroForm()
	if container != nil {
		acroForm.container = container
		container.PdfObject = core.MakeDict()
	}

	if obj := d.Get("Fields"); obj != nil {
		fieldArray, ok := core.GetArray(obj)
		if !ok {
			return nil, fmt.Errorf("fields not an array (%T)", obj)
		}

		var fields []*PdfField
		for _, obj := range fieldArray.Elements() {
			container, isIndirect := core.GetIndirect(obj)
			if !isIndirect {
				if _, isNull := obj.(*core.PdfObjectNull); isNull {
					common.Log.Trace("Skipping over null field")
					continue
				}
				common.Log.Debug("Field not contained in indirect object %T", obj)
				return nil, fmt.Errorf("field not in an indirect object")
			}
			field, err := r.newPdfFieldFromIndirectObject(container, nil)
			if err != nil {
				return nil, err
			}
			common.Log.Trace("AcroForm Field: %+v", *field)
			fields = append(fields, field)
		}
		acroForm.Fields = &fields
	}

	if obj := d.Get("NeedAppearances"); obj != nil {
		val, ok := core.GetBool(obj)
		if ok {
			acroForm.NeedAppearances = val
		} else {
			common.Log.Debug("ERROR: NeedAppearances invalid (got %T)", obj)
		}
	}

	if obj := d.Get("SigFlags"); obj != nil {
		val, ok := core.GetInt(obj)
		if ok {
			acroForm.SigFlags = val
		} else {
			common.Log.Debug("ERROR: SigFlags invalid (got %T)", obj)
		}
	}

	if obj := d.Get("CO"); obj != nil {
		arr, ok := core.GetArray(obj)
		if ok {
			acroForm.CO = arr
		} else {
			common.Log.Debug("ERROR: CO invalid (got %T)", obj)
		}
	}

	if obj := d.Get("DR"); obj != nil {
		if d, ok := core.GetDict(obj); ok {
			resources, err := NewPdfPageResourcesFromDict(d)
			if err != nil {
				common.Log.Error("Invalid DR: %v", err)
				return nil, err
			}

			acroForm.DR = resources
		} else {
			common.Log.Debug("ERROR: DR invalid (got %T)", obj)
		}
	}

	if obj := d.Get("DA"); obj != nil {
		str, ok := core.GetString(obj)
		if ok {
			acroForm.DA = str
		} else {
			common.Log.Debug("ERROR: DA invalid (got %T)", obj)
		}
	}

	if obj := d.Get("Q"); obj != nil {
		val, ok := core.GetInt(obj)
		if ok {
			acroForm.Q = val
		} else {
			common.Log.Debug("ERROR: Q invalid (got %T)", obj)
		}
	}

	if obj := d.Get("XFA"); obj != nil {
		acroForm.XFA = obj
	}

	return acroForm, nil
}

// GetContainingPdfObject returns the container of the PdfAcroForm (indirect object).
func (form *PdfAcroForm) GetContainingPdfObject() core.PdfObject {
	return form.container
}

// ToPdfObject converts PdfAcroForm to a PdfObject, i.e. an indirect object containing the
// AcroForm dictionary.
func (form *PdfAcroForm) ToPdfObject() core.PdfObject {
	container := form.container
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	if form.Fields != nil {
		arr := core.PdfObjectArray{}
		for _, field := range *form.Fields {
			ctx := field.GetContext()
			if ctx != nil {
				// Call subtype's ToPdfObject directly to get the entire field data.
				arr.Append(ctx.ToPdfObject())
			} else {
				arr.Append(field.ToPdfObject())
			}
		}
		dict.Set("Fields", &arr)
	}

	if form.NeedAppearances != nil {
		dict.Set("NeedAppearances", form.NeedAppearances)
	}
	if form.SigFlags != nil {
		dict.Set("SigFlags", form.SigFlags)
	}
	if form.CO != nil {
		dict.Set("CO", form.CO)
	}
	if form.DR != nil {
		dict.Set("DR", form.DR.ToPdfObject())
	}
	if form.DA != nil {
		dict.Set("DA", form.DA)
	}
	if form.Q != nil {
		dict.Set("Q", form.Q)
	}
	if form.XFA != nil {
		dict.Set("XFA", form.XFA)
	}

	return container
}

// FieldValueProvider provides field values from a data source such as FDF, JSON or any other.
type FieldValueProvider interface {
	FieldValues() (map[string]core.PdfObject, error)
}

// Fill populates `form` with values provided by `provider`.
func (form *PdfAcroForm) Fill(provider FieldValueProvider) error {
	return form.fill(provider, nil)
}

// FillWithAppearance populates `form` with values provided by `provider`.
// If not nil, `appGen` is used to generate appearance dictionaries for the
// field annotations, based on the specified settings. Otherwise, appearance
// generation is skipped.
// e.g.: appGen := annotator.FieldAppearance{OnlyIfMissing: true, RegenerateTextFields: true}
// NOTE: In next major version this functionality will be part of Fill. (v4)
func (form *PdfAcroForm) FillWithAppearance(provider FieldValueProvider, appGen FieldAppearanceGenerator) error {
	return form.fill(provider, appGen)
}

// fill populates `form` with values provided by `provider`. If `appGen` is
// not nil, field appearances are also generated.
func (form *PdfAcroForm) fill(provider FieldValueProvider, appGen FieldAppearanceGenerator) error {
	if form == nil {
		return nil
	}
	objMap, err := provider.FieldValues()
	if err != nil {
		return err
	}

	for _, field := range form.AllFields() {
		// Try finding the field in the provider field map using its partial
		// name. If not found, try finding it by its full name.
		partialName := field.PartialName()
		valObj, found := objMap[partialName]
		if !found {
			if fullName, err := field.FullName(); err == nil {
				valObj, found = objMap[fullName]
			}
		}
		if !found {
			common.Log.Debug("WARN: form field %s not found in the provider. Skipping.", partialName)
			continue
		}

		// Fill field with the provided value.
		if err := fillFieldValue(field, valObj); err != nil {
			return err
		}

		// Generate field appearance based on the specified settings.
		if appGen == nil {
			continue
		}

		for _, annot := range field.Annotations {
			// appGen generates the appearance based on the form/field/annotation and other settings
			// depending on the implementation (for example may only generate appearance if none set).
			apDict, err := appGen.GenerateAppearanceDict(form, field, annot)
			if err != nil {
				return err
			}

			annot.AP = apDict
			annot.ToPdfObject()
		}
	}

	return nil
}

// fillFieldValue populates form field `f` with value represented by `v`.
func fillFieldValue(f *PdfField, val core.PdfObject) error {
	switch f.GetContext().(type) {
	case *PdfFieldText:
		switch t := val.(type) {
		case *core.PdfObjectName:
			name := t
			common.Log.Debug("Unexpected: Got V as name -> converting to string '%s'", name.String())
			f.V = core.MakeEncodedString(t.String(), true)
		case *core.PdfObjectString:
			f.V = core.MakeEncodedString(t.String(), true)
		default:
			common.Log.Debug("ERROR: Unsupported text field V type: %T (%#v)", t, t)
		}
	case *PdfFieldButton:
		// See section 12.7.4.2.3 "Check Boxes" (pp. 440-441 PDF32000_2008).
		switch val.(type) {
		case *core.PdfObjectName:
			if len(val.String()) > 0 {
				f.V = val
				setFieldAnnotAS(f, val)
			}
		case *core.PdfObjectString:
			if len(val.String()) > 0 {
				f.V = core.MakeName(val.String())
				setFieldAnnotAS(f, f.V)
			}
		default:
			common.Log.Debug("ERROR: UNEXPECTED %s -> %v", f.PartialName(), val)
			f.V = val
		}
	case *PdfFieldChoice:
		// See section 12.7.4.4 "Choice Fields" (pp. 444-446 PDF32000_2008).
		switch val.(type) {
		case *core.PdfObjectName:
			if len(val.String()) > 0 {
				f.V = core.MakeString(val.String())
				setFieldAnnotAS(f, val)
			}
		case *core.PdfObjectString:
			if len(val.String()) > 0 {
				f.V = val
				setFieldAnnotAS(f, core.MakeName(val.String()))
			}
		default:
			common.Log.Debug("ERROR: UNEXPECTED %s -> %v", f.PartialName(), val)
			f.V = val
		}
	case *PdfFieldSignature:
		common.Log.Debug("TODO: Signature appearance not supported yet: %s/%v", f.PartialName(), val)
	}

	return nil
}

// setFieldAnnotAS sets the appearance stream of the field annotations to `val`.
func setFieldAnnotAS(f *PdfField, val core.PdfObject) {
	for _, wa := range f.Annotations {
		wa.AS = val
		wa.ToPdfObject()
	}
}
