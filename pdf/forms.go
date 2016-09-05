/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

//
// Higher level manipulation of forms (AcroForm).
//

package pdf

import (
	"fmt"
)

// Higher level object convertible to a PDF primitive.
// type PdfObjectConvertible interface {
//	ToPdfObject(updateIfExists bool) PdfObject
//}

type PdfAcroForm struct {
	Fields          *[]*PdfField
	NeedAppearances PdfObject
	SigFlags        PdfObject
	CO              PdfObject
	DR              PdfObject
	DA              PdfObject
	Q               PdfObject
	XFA             PdfObject
}

// Used when loading forms from PDF files.
func (r *PdfReader) newPdfAcroFormFromDict(d PdfObjectDictionary) (*PdfAcroForm, error) {
	acroForm := PdfAcroForm{}

	if obj, has := d["Fields"]; has {
		obj, err := r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		fieldArray, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Fields not an array (%T)", obj)
		}

		fields := []*PdfField{}
		for _, obj := range *fieldArray {
			obj, err := r.traceToObject(obj)
			if err != nil {
				return nil, err
			}
			fDict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
			if !ok {
				return nil, fmt.Errorf("Invalid Fields entry: %T", obj)
			}
			field, err := r.newPdfFieldFromDict(*fDict, nil)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
		acroForm.Fields = &fields
	}

	if obj, has := d["NeedAppearances"]; has {
		acroForm.NeedAppearances = obj
	}
	if obj, has := d["SigFlags"]; has {
		acroForm.SigFlags = obj
	}
	if obj, has := d["CO"]; has {
		acroForm.CO = obj
	}
	if obj, has := d["DR"]; has {
		acroForm.DR = obj
	}
	if obj, has := d["DA"]; has {
		acroForm.DA = obj
	}
	if obj, has := d["Q"]; has {
		acroForm.Q = obj
	}
	if obj, has := d["XFA"]; has {
		acroForm.XFA = obj
	}

	return &acroForm, nil
}

func (this *PdfAcroForm) ToPdfObject(updateIfExists bool) PdfObject {
	var container PdfIndirectObject

	if cachedObj, isCached := PdfObjectConverterCache[this]; isCached {
		if !updateIfExists {
			return cachedObj
		}
		obj := cachedObj.(*PdfIndirectObject)
		container = *obj
	}

	container = PdfIndirectObject{}
	dict := PdfObjectDictionary{}
	container.PdfObject = &dict

	if this.Fields != nil {
		arr := PdfObjectArray{}
		for _, field := range *this.Fields {
			arr = append(arr, field.ToPdfObject(false))
		}
		dict["Fields"] = &arr
	}

	if this.NeedAppearances != nil {
		dict["NeedAppearances"] = this.NeedAppearances
	}
	if this.SigFlags != nil {
		dict["SigFlags"] = this.SigFlags
	}
	if this.CO != nil {
		dict["CO"] = this.CO
	}
	if this.DR != nil {
		dict["DR"] = this.DR
	}
	if this.DA != nil {
		dict["DA"] = this.DA
	}
	if this.Q != nil {
		dict["Q"] = this.Q
	}
	if this.XFA != nil {
		dict["XFA"] = this.XFA
	}

	PdfObjectConverterCache[this] = &container
	return &container
}

type PdfField struct {
	FT     *PdfObjectName // field type
	Parent *PdfField
	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.
	Kids []PdfObjectConverter // Kids can be array of other fields or widgets (PdfObjectConverter)
	T    PdfObject
	TU   PdfObject
	TM   PdfObject
	Ff   PdfObject // field flag
	V    PdfObject //value
	DV   PdfObject
	AA   PdfObject
	// Widget annotation can be merged in.
	// PdfAnnotationWidget
}

// Used when loading fields from PDF files.
func (r *PdfReader) newPdfFieldFromDict(d PdfObjectDictionary, parent *PdfField) (*PdfField, error) {
	field := PdfField{}

	// Field type (required in terminal fields).
	// Can be /Btn /Tx /Ch /Sig
	// Required for a terminal field (inheritable).
	var err error
	if obj, has := d["FT"]; has {
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Invalid type of FT field (%T)", obj)
		}

		field.FT = name
	}

	// Partial field name (Optional)
	if obj, has := d["T"]; has {
		field.T = obj
	}
	// Alternate description (Optional)
	if obj, has := d["TU"]; has {
		field.TU = obj
	}
	// Mapping name (Optional)
	if obj, has := d["TM"]; has {
		field.TM = obj
	}
	// Field flag. (Optional; inheritable)
	if obj, has := d["Ff"]; has {
		field.Ff = obj
	}
	// Value (Optional; inheritable) - Various types depending on the field type.
	if obj, has := d["V"]; has {
		field.V = obj
	}
	// Default value for reset (Optional; inheritable)
	if obj, has := d["DV"]; has {
		field.DV = obj
	}
	// Additional actions dictionary (Optional)
	if obj, has := d["AA"]; has {
		field.AA = obj
	}

	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.

	// Set ourself?
	if parent != nil {
		field.Parent = parent
	}

	// Has a merged-in widget annotation?
	if obj, has := d["Subtype"]; has {
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Invalid type of Subtype (%T)", obj)
		}
		if *name == "Widget" {
			// Is a merged field / widget dict.
			widget := r.newPdfAnnotationWidgetFromDict(d, &field)
			field.Kids = append(field.Kids, PdfObjectConverter(widget))
			return &field, nil
		}
	}

	if obj, has := d["Kids"]; has {
		obj, err := r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		fieldArray, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Fields not an array (%T)", obj)
		}

		field.Kids = []PdfObjectConverter{}
		for _, obj := range *fieldArray {
			obj, err := r.traceToObject(obj)
			if err != nil {
				return nil, err
			}
			fDict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
			if !ok {
				return nil, fmt.Errorf("Invalid Fields entry: %T", obj)
			}

			childField, err := r.newPdfFieldFromDict(*fDict, &field)
			if err != nil {
				return nil, err
			}

			field.Kids = append(field.Kids, childField)
		}
	}

	return &field, nil
}

// If Kids refer only to a single pdf widget annotation widget, then can merge it in.
// Currently not merging it in.
func (this *PdfField) ToPdfObject(updateIfExists bool) PdfObject {
	var container PdfIndirectObject

	if cachedObj, isCached := PdfObjectConverterCache[this]; isCached {
		if !updateIfExists {
			return cachedObj
		}
		obj := cachedObj.(*PdfIndirectObject)
		container = *obj
	}

	container = PdfIndirectObject{}
	dict := PdfObjectDictionary{}
	container.PdfObject = &dict

	if this.Parent != nil {
		dict["Parent"] = this.Parent.ToPdfObject(false)
	}

	if this.Kids != nil {
		// Create an array of the kids (fields or widgets).
		arr := PdfObjectArray{}
		for _, child := range this.Kids {
			arr = append(arr, child.ToPdfObject(false))
		}
		dict["Kids"] = &arr
	}

	if this.T != nil {
		dict["T"] = this.T
	}
	if this.TU != nil {
		dict["TU"] = this.TU
	}
	if this.TM != nil {
		dict["TM"] = this.TM
	}
	if this.Ff != nil {
		dict["Ff"] = this.Ff
	}
	if this.V != nil {
		dict["V"] = this.V
	}
	if this.DV != nil {
		dict["DV"] = this.DV
	}
	if this.AA != nil {
		dict["AA"] = this.AA
	}

	PdfObjectConverterCache[this] = &container
	return &container
}

/*
type PdfAnnotationWidget struct {
	Subtype PdfObject // Widget (required)
	H       PdfObject
	MK      PdfObject
	A       PdfObject
	AA      PdfObject
	BS      PdfObject
	Parent  PdfObjectConverter
}*/

func (r *PdfReader) newPdfAnnotationWidgetFromDict(d PdfObjectDictionary, parent *PdfField) *PdfAnnotationWidget {
	annotation := PdfAnnotationWidget{}

	if obj, has := d["Subtype"]; has {
		annotation.Subtype = obj
	}
	if obj, has := d["H"]; has {
		annotation.H = obj
	}
	if obj, has := d["MK"]; has {
		annotation.MK = obj
	}
	if obj, has := d["A"]; has {
		annotation.A = obj
	}
	if obj, has := d["AA"]; has {
		annotation.AA = obj
	}
	if obj, has := d["BS"]; has {
		annotation.BS = obj
	}

	annotation.Parent = parent
	return &annotation
}

func (this *PdfAnnotationWidget) ToPdfObject(updateIfExists bool) PdfObject {
	var container PdfIndirectObject

	if cachedObj, isCached := PdfObjectConverterCache[this]; isCached {
		if !updateIfExists {
			return cachedObj
		}
		obj := cachedObj.(*PdfIndirectObject)
		container = *obj
	}

	container = PdfIndirectObject{}
	dict := PdfObjectDictionary{}
	container.PdfObject = &dict
	d := PdfObjectDictionary{}

	d.SetIfNotNil("Subtype", this.Subtype)
	d.SetIfNotNil("H", this.H)
	d.SetIfNotNil("MK", this.MK)
	d.SetIfNotNil("A", this.A)
	d.SetIfNotNil("AA", this.AA)
	d.SetIfNotNil("BS", this.BS)
	if this.Parent != nil {
		d["Parent"] = this.Parent.ToPdfObject(false)
	}

	PdfObjectConverterCache[this] = &container
	return &container
}
