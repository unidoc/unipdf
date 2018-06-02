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

// PdfField represents a field of an interactive form.
// Implements PdfModel interface.
type PdfField struct {
	FT     *core.PdfObjectName // field type
	Parent *PdfField
	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.
	KidsF []PdfModel // Kids can be array of other fields or widgets (PdfModel).
	KidsA []*PdfAnnotation
	T     core.PdfObject
	TU    core.PdfObject
	TM    core.PdfObject
	Ff    core.PdfObject // field flag
	V     core.PdfObject //value
	DV    core.PdfObject
	AA    core.PdfObject

	// Variable Text:
	DA core.PdfObject
	Q  core.PdfObject
	DS core.PdfObject
	RV core.PdfObject

	primitive *core.PdfIndirectObject
}

func NewPdfField() *PdfField {
	field := &PdfField{}
	container := core.MakeIndirectObject(core.MakeDict())
	field.primitive = container
	return field
}

// newPdfFieldFromIndirectObject is used to load fields from PDF files.
func (r *PdfReader) newPdfFieldFromIndirectObject(container *core.PdfIndirectObject, parent *PdfField) (*PdfField, error) {
	d, isDict := container.PdfObject.(*core.PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("Pdf Field indirect object not containing a dictionary")
	}

	field := NewPdfField()

	// Field type (required in terminal fields).
	// Can be /Btn /Tx /Ch /Sig
	// Required for a terminal field (inheritable).
	var err error
	if obj := d.Get("FT"); obj != nil {
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		name, ok := obj.(*core.PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Invalid type of FT field (%T)", obj)
		}

		field.FT = name
	}

	// Partial field name (Optional)
	field.T = d.Get("T")
	// Alternate description (Optional)
	field.TU = d.Get("TU")
	// Mapping name (Optional)
	field.TM = d.Get("TM")
	// Field flag. (Optional; inheritable)
	field.Ff = d.Get("Ff")
	// Value (Optional; inheritable) - Various types depending on the field type.
	field.V = d.Get("V")
	// Default value for reset (Optional; inheritable)
	field.DV = d.Get("DV")
	// Additional actions dictionary (Optional)
	field.AA = d.Get("AA")

	// Variable text:
	field.DA = d.Get("DA")
	field.Q = d.Get("Q")
	field.DS = d.Get("DS")
	field.RV = d.Get("RV")

	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.

	// Set ourself?
	if parent != nil {
		field.Parent = parent
	}

	// Has a merged-in widget annotation?
	if obj := d.Get("Subtype"); obj != nil {
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		common.Log.Trace("Merged in annotation (%T)", obj)
		name, ok := obj.(*core.PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Invalid type of Subtype (%T)", obj)
		}
		if *name == "Widget" {
			// Is a merged field / widget dict.

			// Check if the annotation has already been loaded?
			// Most likely referenced to by a page...  Could be in either direction.
			// r.newPdfAnnotationFromIndirectObject acts as a caching mechanism.
			annot, err := r.newPdfAnnotationFromIndirectObject(container)
			if err != nil {
				return nil, err
			}
			widget, ok := annot.GetContext().(*PdfAnnotationWidget)
			if !ok {
				return nil, fmt.Errorf("Invalid widget")
			}

			widget.Parent = field.GetContainingPdfObject()
			field.KidsA = append(field.KidsA, annot)
			return field, nil
		}
	}

	// Has kids (can be field and/or widget annotations).
	if obj := d.Get("Kids"); obj != nil {
		obj, err := r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		fieldArray, ok := core.TraceToDirectObject(obj).(*core.PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Kids not an array (%T)", obj)
		}

		field.KidsF = []PdfModel{}
		for _, obj := range *fieldArray {
			obj, err := r.traceToObject(obj)
			if err != nil {
				return nil, err
			}

			container, isIndirect := obj.(*core.PdfIndirectObject)
			if !isIndirect {
				return nil, fmt.Errorf("Not an indirect object (form field)")
			}

			childField, err := r.newPdfFieldFromIndirectObject(container, field)
			if err != nil {
				return nil, err
			}

			field.KidsF = append(field.KidsF, childField)
		}
	}

	return field, nil
}

// GetContainingPdfObject returns the containing object for the PdfField, i.e. an indirect object
// containing the field dictionary.
func (this *PdfField) GetContainingPdfObject() core.PdfObject {
	return this.primitive
}

// ToPdfObject returns the Field as a field dictionary within an indirect object as container.
// If Kids refer only to a single pdf widget annotation widget, then can merge it in.
// Currently not merging it in.
func (this *PdfField) ToPdfObject() core.PdfObject {
	container := this.primitive
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	if this.Parent != nil {
		dict.Set("Parent", this.Parent.GetContainingPdfObject())
	}

	if this.KidsF != nil {
		// Create an array of the kids (fields or widgets).
		common.Log.Trace("KidsF: %+v", this.KidsF)
		arr := core.PdfObjectArray{}
		for _, child := range this.KidsF {
			arr = append(arr, child.ToPdfObject())
		}
		dict.Set("Kids", &arr)
	}
	if this.KidsA != nil {
		common.Log.Trace("KidsA: %+v", this.KidsA)
		_, hasKids := dict.Get("Kids").(*core.PdfObjectArray)
		if !hasKids {
			dict.Set("Kids", &core.PdfObjectArray{})
		}
		arr := dict.Get("Kids").(*core.PdfObjectArray)
		for _, child := range this.KidsA {
			*arr = append(*arr, child.GetContext().ToPdfObject())
		}
	}

	if this.FT != nil {
		dict.Set("FT", this.FT)
	}

	if this.T != nil {
		dict.Set("T", this.T)
	}
	if this.TU != nil {
		dict.Set("TU", this.TU)
	}
	if this.TM != nil {
		dict.Set("TM", this.TM)
	}
	if this.Ff != nil {
		dict.Set("Ff", this.Ff)
	}
	if this.V != nil {
		dict.Set("V", this.V)
	}
	if this.DV != nil {
		dict.Set("DV", this.DV)
	}
	if this.AA != nil {
		dict.Set("AA", this.AA)
	}

	// Variable text:
	dict.SetIfNotNil("DA", this.DA)
	dict.SetIfNotNil("Q", this.Q)
	dict.SetIfNotNil("DS", this.DS)
	dict.SetIfNotNil("RV", this.RV)

	return container
}
