/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
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
	acroForm := &PdfAcroForm{}

	container := &core.PdfIndirectObject{}
	container.PdfObject = core.MakeDict()
	acroForm.container = container

	acroForm.Fields = &[]*PdfField{}

	return acroForm
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
	fields := []*PdfField{}
	if form.Fields != nil {
		for _, field := range *form.Fields {
			fields = append(fields, flattenFields(field)...)
		}
	}
	return fields
}

// signatureFields returns a slice of all signature fields in the form.
func (form *PdfAcroForm) signatureFields() []*PdfFieldSignature {
	sigfields := []*PdfFieldSignature{}

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
func (r *PdfReader) newPdfAcroFormFromDict(d *core.PdfObjectDictionary) (*PdfAcroForm, error) {
	acroForm := NewPdfAcroForm()

	if obj := d.Get("Fields"); obj != nil {
		obj, err := r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		fieldArray, ok := core.TraceToDirectObject(obj).(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Fields not an array (%T)", obj)
		}

		fields := []*PdfField{}
		for _, obj := range fieldArray.Elements() {
			obj, err := r.traceToObject(obj)
			if err != nil {
				return nil, err
			}
			container, isIndirect := obj.(*core.PdfIndirectObject)
			if !isIndirect {
				if _, isNull := obj.(*core.PdfObjectNull); isNull {
					common.Log.Trace("Skipping over null field")
					continue
				}
				common.Log.Debug("Field not contained in indirect object %T", obj)
				return nil, fmt.Errorf("Field not in an indirect object")
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
		val, ok := obj.(*core.PdfObjectBool)
		if ok {
			acroForm.NeedAppearances = val
		} else {
			common.Log.Debug("ERROR: NeedAppearances invalid (got %T)", obj)
		}
	}

	if obj := d.Get("SigFlags"); obj != nil {
		val, ok := obj.(*core.PdfObjectInteger)
		if ok {
			acroForm.SigFlags = val
		} else {
			common.Log.Debug("ERROR: SigFlags invalid (got %T)", obj)
		}
	}

	if obj := d.Get("CO"); obj != nil {
		obj = core.TraceToDirectObject(obj)
		arr, ok := obj.(*core.PdfObjectArray)
		if ok {
			acroForm.CO = arr
		} else {
			common.Log.Debug("ERROR: CO invalid (got %T)", obj)
		}
	}

	if obj := d.Get("DR"); obj != nil {
		obj = core.TraceToDirectObject(obj)
		if d, ok := obj.(*core.PdfObjectDictionary); ok {
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
		str, ok := obj.(*core.PdfObjectString)
		if ok {
			acroForm.DA = str
		} else {
			common.Log.Debug("ERROR: DA invalid (got %T)", obj)
		}
	}

	if obj := d.Get("Q"); obj != nil {
		val, ok := obj.(*core.PdfObjectInteger)
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
func (this *PdfAcroForm) GetContainingPdfObject() core.PdfObject {
	return this.container
}

// ToPdfObject converts PdfAcroForm to a PdfObject, i.e. an indirect object containing the
// AcroForm dictionary.
func (this *PdfAcroForm) ToPdfObject() core.PdfObject {
	container := this.container
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	if this.Fields != nil {
		arr := core.PdfObjectArray{}
		for _, field := range *this.Fields {
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

	if this.NeedAppearances != nil {
		dict.Set("NeedAppearances", this.NeedAppearances)
	}
	if this.SigFlags != nil {
		dict.Set("SigFlags", this.SigFlags)
	}
	if this.CO != nil {
		dict.Set("CO", this.CO)
	}
	if this.DR != nil {
		dict.Set("DR", this.DR.ToPdfObject())
	}
	if this.DA != nil {
		dict.Set("DA", this.DA)
	}
	if this.Q != nil {
		dict.Set("Q", this.Q)
	}
	if this.XFA != nil {
		dict.Set("XFA", this.XFA)
	}

	return container
}
