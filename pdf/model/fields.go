/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// FieldFlag represents form field flags. Some of the flags can apply to all types of fields whereas other
// flags are specific.
type FieldFlag uint32

// The following constants define bitwise flag representing different attributes of a form field.
const (
	// FieldFlagClear has no flags.
	FieldFlagClear FieldFlag = 0

	// Flags for all field types.
	FieldFlagReadOnly FieldFlag = 1
	FieldFlagRequired FieldFlag = (1 << 1)
	FieldFlagNoExport FieldFlag = (2 << 1)

	// Flags for button fields only.
	FieldFlagNoToggleToOff   FieldFlag = (1 << 14)
	FieldFlagRadio           FieldFlag = (1 << 15)
	FieldFlagPushbutton      FieldFlag = (1 << 16)
	FieldFlagRadiosInUnision FieldFlag = (1 << 25)

	// Flags for text fields only.
	FieldFlagMultiline   FieldFlag = (1 << 12)
	FieldFlagPassword    FieldFlag = (1 << 13)
	FieldFlagFileSelect  FieldFlag = (1 << 20)
	FieldFlagDoNotScroll FieldFlag = (1 << 23)
	FieldFlagComb        FieldFlag = (1 << 24)
	FieldFlagRichText    FieldFlag = (1 << 25)

	// Flags for text and choice fields.
	FieldFlagDoNotSpellCheck FieldFlag = (1 << 22)

	// Flags for choice fields only.
	FieldFlagCombo             FieldFlag = (1 << 17)
	FieldFlagEdit              FieldFlag = (1 << 18)
	FieldFlagSort              FieldFlag = (1 << 19)
	FieldFlagMultiSelect       FieldFlag = (1 << 21)
	FieldFlagCommitOnSelChange FieldFlag = (1 << 26)
)

// Mask returns the uin32 bitmask for the specific flag.
func (flag FieldFlag) Mask() uint32 {
	return uint32(flag)
}

// Set applies flag fl to the flag's bitmask and returns the combined flag.
func (flag FieldFlag) Set(fl FieldFlag) FieldFlag {
	return FieldFlag(flag.Mask() | fl.Mask())
}

// Clear clears flag fl from the flag and returns the resulting flag.
func (flag FieldFlag) Clear(fl FieldFlag) FieldFlag {
	return FieldFlag(flag.Mask() &^ fl.Mask())
}

// Has checks if flag fl is set in flag and returns true if so, false otherwise.
func (flag FieldFlag) Has(fl FieldFlag) bool {
	return (flag.Mask() & fl.Mask()) > 0
}

// PdfField contains the common attributes of a form field. The context object contains the specific field data
// which can represent a button, text, choice or signature.
// The PdfField is typically not used directly, but is encapsulated by the more specific field types such as
// PdfFieldButton etc (i.e. the context attribute).
type PdfField struct {
	context    PdfModel                // Field data
	container  *core.PdfIndirectObject // Dictionary information stored inside an indirect object.
	isTerminal *bool                   // If set: indicates whether is a terminal field (if null, may not be determined yet).

	Parent      *PdfField
	Annotations []*PdfAnnotation
	Kids        []*PdfField

	FT *core.PdfObjectName
	//Kids   *core.PdfObjectArray
	T  *core.PdfObjectString
	TU *core.PdfObjectString
	TM *core.PdfObjectString
	Ff *core.PdfObjectInteger
	V  core.PdfObject
	DV core.PdfObject
	AA core.PdfObject
}

// FullName returns the full name of the field as in rootname.parentname.partialname.
func (f *PdfField) FullName() (string, error) {
	var fn bytes.Buffer

	parts := []string{}
	if f.T != nil {
		parts = append(parts, string(*f.T))
	} else {
		return fn.String(), errors.New("Field partial name (T) not specified")
	}

	// Avoid recursive loops by having a list of already traversed nodes.
	noscanMap := map[*PdfField]bool{}
	noscanMap[f] = true

	parent := f.Parent
	for parent != nil {
		if _, has := noscanMap[parent]; has {
			return fn.String(), errors.New("Recursive traversal")
		}

		if parent.T != nil {
			parts = append(parts, string(*f.T))
		} else {
			return fn.String(), errors.New("Field partial name (T) not specified")
		}

		noscanMap[parent] = true
		parent = parent.Parent
	}

	for i := len(parts) - 1; i >= 0; i-- {
		fn.WriteString(parts[i])
		if i > 0 {
			fn.WriteString(".")
		}
	}

	return fn.String(), nil
}

// PartialName returns the partial name of the field.
func (f *PdfField) PartialName() string {
	partial := ""
	if f.T != nil {
		partial = string(*f.T)
	} else {
		common.Log.Debug("Field missing T field (incompatible)")
	}
	return partial
}

// GetContext returns the PdfField context which is the more specific field data type, e.g. PdfFieldButton
// for a button field.
func (f *PdfField) GetContext() PdfModel {
	return f.context
}

// SetContext sets the specific fielddata type, e.g. would be PdfFieldButton for a button field.
func (f *PdfField) SetContext(ctx PdfModel) {
	f.context = ctx
}

// GetContainingPdfObject returns the containing object for the PdfField, i.e. an indirect object
// containing the field dictionary.
func (f *PdfField) GetContainingPdfObject() core.PdfObject {
	return f.container
}

// String returns a string representation of the field.
func (f *PdfField) String() string {
	if obj, ok := f.ToPdfObject().(*core.PdfIndirectObject); ok {
		return fmt.Sprintf("%T: %s", f.context, obj.PdfObject.String())
	}
	return ""
}

// ToPdfObject sets the common field elements.
// Note: Call the more field context's ToPdfObject to set both the generic and non-generic information.
func (f *PdfField) ToPdfObject() core.PdfObject {
	container := f.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("FT", f.FT)
	if f.Parent != nil {
		d.Set("Parent", f.Parent.GetContainingPdfObject())
	}

	if f.Kids != nil {
		// Create an array of the kids (fields or widgets).
		kids := core.MakeArray()
		for _, child := range f.Kids {
			kids.Append(child.ToPdfObject())
		}
		d.Set("Kids", kids)
	}

	if f.Annotations != nil {
		_, hasKids := d.Get("Kids").(*core.PdfObjectArray)
		if !hasKids {
			d.Set("Kids", &core.PdfObjectArray{})
		}
		// TODO: If only 1 widget annotation, it can be merged in.
		kids := d.Get("Kids").(*core.PdfObjectArray)
		for _, annot := range f.Annotations {
			kids.Append(annot.GetContext().ToPdfObject())
		}
	}

	d.SetIfNotNil("T", f.T)
	d.SetIfNotNil("TU", f.TU)
	d.SetIfNotNil("TM", f.TM)
	d.SetIfNotNil("Ff", f.Ff)
	if f.V != nil {
		d.Set("V", f.V)
	}
	if f.DV != nil {
		d.Set("DV", f.DV)
	}
	if f.AA != nil {
		d.Set("AA", f.DV)
	}

	return container
}

// PdfFieldButton represents a button field which includes push buttons, checkboxes, and radio buttons.
type PdfFieldButton struct {
	*PdfField
	Opt *core.PdfObjectArray
}

// ButtonType represents the subtype of a button field, can be one of:
// - Checkbox (ButtonTypeCheckbox)
// - PushButton (ButtonTypePushButton)
// - RadioButton (ButtonTypeRadioButton)
type ButtonType int

// Definitions for field button types
const (
	ButtonTypeCheckbox ButtonType = iota
	ButtonTypePush     ButtonType = iota
	ButtonTypeRadio    ButtonType = iota
)

// GetType returns the button field type which returns one of the following
// - PdfFieldButtonPush for push button fields
// - PdfFieldButtonCheckbox for checkbox fields
// - PdfFieldButtonRadio for radio button fields
func (fb *PdfFieldButton) GetType() ButtonType {
	btype := ButtonTypeCheckbox
	if fb.Ff != nil {
		if (uint32(*fb.Ff) & FieldFlagPushbutton.Mask()) > 0 {
			btype = ButtonTypePush
		} else if (uint32(*fb.Ff) & FieldFlagRadio.Mask()) > 0 {
			btype = ButtonTypeRadio
		}
	}

	return btype
}

// IsPush returns true if the button field represents a push button, false otherwise.
func (fb *PdfFieldButton) IsPush() bool {
	return fb.GetType() == ButtonTypePush
}

// IsCheckbox returns true if the button field represents a checkbox, false otherwise.
func (fb *PdfFieldButton) IsCheckbox() bool {
	return fb.GetType() == ButtonTypeCheckbox
}

// IsRadio returns true if the button field represents a radio button, false otherwise.
func (fb *PdfFieldButton) IsRadio() bool {
	return fb.GetType() == ButtonTypeRadio
}

// SetType sets the field button's type.  Can be one of:
// - PdfFieldButtonPush for push button fields
// - PdfFieldButtonCheckbox for checkbox fields
// - PdfFieldButtonRadio for radio button fields
// This sets the field's flag appropriately.
func (fb *PdfFieldButton) SetType(btype ButtonType) {
	flag := uint32(0)
	if fb.Ff != nil {
		flag = uint32(*fb.Ff)
	}

	switch btype {
	case ButtonTypePush:
		flag |= FieldFlagPushbutton.Mask()
	case ButtonTypeRadio:
		flag |= FieldFlagRadio.Mask()
	}

	fb.Ff = core.MakeInteger(int64(flag))
}

// ToPdfObject returns the button field dictionary within an indirect object.
func (fb *PdfFieldButton) ToPdfObject() core.PdfObject {
	fb.PdfField.ToPdfObject()
	container := fb.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	d.Set("FT", core.MakeName("Btn"))
	if fb.Opt != nil {
		d.Set("Opt", fb.Opt)
	}
	return container
}

// PdfFieldText represents a text field where user can enter text.
type PdfFieldText struct {
	*PdfField
	DA     *core.PdfObjectString
	Q      *core.PdfObjectInteger
	DS     *core.PdfObjectString
	RV     core.PdfObject
	MaxLen *core.PdfObjectInteger
}

// ToPdfObject returns the text field dictionary within an indirect object (container).
func (ft *PdfFieldText) ToPdfObject() core.PdfObject {
	ft.PdfField.ToPdfObject()
	container := ft.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	d.Set("FT", core.MakeName("Tx"))
	if ft.DA != nil {
		d.Set("DA", ft.DA)
	}
	if ft.Q != nil {
		d.Set("Q", ft.Q)
	}
	if ft.DS != nil {
		d.Set("DS", ft.DS)
	}
	if ft.RV != nil {
		d.Set("RV", ft.RV)
	}
	if ft.MaxLen != nil {
		d.Set("MaxLen", ft.MaxLen)
	}

	return container
}

// PdfFieldChoice represents a choice field which includes scrollable list boxes and combo boxes.
type PdfFieldChoice struct {
	*PdfField
	Opt *core.PdfObjectArray
	TI  *core.PdfObjectInteger
	I   *core.PdfObjectArray
}

// ToPdfObject returns the choice field dictionary within an indirect object (container).
func (ch *PdfFieldChoice) ToPdfObject() core.PdfObject {
	// Set general field attributes
	ch.PdfField.ToPdfObject()
	container := ch.container

	// Handle choice specific attributes
	d := container.PdfObject.(*core.PdfObjectDictionary)
	d.Set("FT", core.MakeName("Ch"))
	if ch.Opt != nil {
		d.Set("Opt", ch.Opt)
	}
	if ch.TI != nil {
		d.Set("TI", ch.TI)
	}
	if ch.I != nil {
		d.Set("I", ch.I)
	}

	return container
}

// PdfFieldSignature signature field represents digital signatures and optional data for authenticating
// the name of the signer and verifying document contents.
type PdfFieldSignature struct {
	*PdfField
	V    *core.PdfIndirectObject
	Lock *core.PdfIndirectObject
	SV   *core.PdfIndirectObject
}

// ToPdfObject returns an indirect object containing the signature field dictionary.
func (sig *PdfFieldSignature) ToPdfObject() *core.PdfIndirectObject {
	// Set general field attributes
	sig.PdfField.ToPdfObject()
	container := sig.container

	// Handle signature field specific attributes
	d := container.PdfObject.(*core.PdfObjectDictionary)
	d.Set("FT", core.MakeName("Sig"))
	if sig.V != nil {
		d.Set("V", sig.V)
	}
	if sig.Lock != nil {
		d.Set("Lock", sig.Lock)
	}
	if sig.SV != nil {
		d.Set("SV", sig.SV)
	}

	return container
}

// NewPdfField returns an initialized PdfField.
func NewPdfField() *PdfField {
	field := &PdfField{}
	container := core.MakeIndirectObject(core.MakeDict())
	field.container = container
	return field
}

// inherit traverses through a field and its ancestry for calculation of inherited attributes. The process is
// down-up with lower nodes being assessed first. Typically values are inherited as-is (without merging), so
// the first non-nil entry is generally the one to use.
// The custom provided eval function evaluates field nodes and returns true once a match was found.
// A bool flag is returned to indicate whether there was a match.
func (f *PdfField) inherit(eval func(*PdfField) bool) (bool, error) {
	nodeMap := map[*PdfField]bool{}
	found := false

	// Traverse from the node up to the root.
	node := f
	for node != nil {
		if _, has := nodeMap[node]; has {
			return false, errors.New("Recursive traversal")
		}

		stop := eval(node)
		if stop {
			found = true
			break
		}

		nodeMap[node] = true
		node = node.Parent
	}

	return found, nil
}

// IsTerminal returns true for terminal fields, false otherwise.
// Terminal fields are fields whose descendants are only widget annotations.
func (f *PdfField) IsTerminal() bool {
	if len(f.Kids) == 0 {
		return true
	}
	return false
}

// Flags returns the field flags for the field accounting for any inherited flags.
func (f *PdfField) Flags() FieldFlag {
	var flags FieldFlag
	found, err := f.inherit(func(node *PdfField) bool {
		if node.Ff != nil {
			flags = FieldFlag(*f.Ff)
			return true
		}
		return false
	})
	if err != nil {
		common.Log.Debug("Error evaluating flags via inheritance: %v", err)
	}
	if !found {
		common.Log.Debug("No field flags found")
	}

	return flags
}

// newPdfFieldFromIndirectObject load a field from an indirect object containing the field dictionary.
func (r *PdfReader) newPdfFieldFromIndirectObject(container *core.PdfIndirectObject, parent *PdfField) (*PdfField, error) {
	d, isDict := container.PdfObject.(*core.PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("PdfField indirect object not containing a dictionary")
	}

	field := NewPdfField()

	// Field type (required in terminal fields).
	// Can be /Btn /Tx /Ch /Sig
	// Required for a terminal field (inheritable).
	if obj := d.GetDirect("FT"); obj != nil {
		name, ok := obj.(*core.PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Invalid type of FT field (%T)", obj)
		}

		field.FT = name
		isTerminal := true
		field.isTerminal = &isTerminal
	} else {
		isTerminal := false
		field.isTerminal = &isTerminal
	}

	// Non-terminal field: One whose descendants are fields.  Not an actual field, just a container for inheritable
	// attributes for descendant terminal fields. Does not logically have a type of its own.
	// Terminal field: An actual field with a type.

	// Partial field name (Required)
	if s, has := d.Get("T").(*core.PdfObjectString); has {
		field.T = s
	} else {
		common.Log.Debug("Invalid - T field missing (required)")
		return nil, ErrTypeCheck
	}

	// Alternate description (Optional)
	field.TU, _ = d.Get("TU").(*core.PdfObjectString)

	// Mapping name (Optional)
	field.TM, _ = d.Get("TM").(*core.PdfObjectString)

	// Field flag. (Optional; inheritable)
	field.Ff, _ = d.Get("Ff").(*core.PdfObjectInteger)

	// Value (Optional; inheritable) - Various types depending on the field type.
	field.V = d.Get("V")

	// Default value for reset (Optional; inheritable)
	field.DV = d.Get("DV")

	// Additional actions dictionary (Optional)
	field.AA = d.Get("AA")

	// Type specific types.
	if field.FT != nil {
		switch *field.FT {
		case "Tx":
			ctx, err := newPdfFieldTextFromDict(d)
			if err != nil {
				return nil, err
			}
			ctx.PdfField = field
			field.context = ctx
		case "Ch":
			ctx, err := newPdfFieldChoiceFromDict(d)
			if err != nil {
				return nil, err
			}
			ctx.PdfField = field
			field.context = ctx
		case "Btn":
			ctx, err := newPdfFieldButtonFromDict(d)
			if err != nil {
				return nil, err
			}
			ctx.PdfField = field
			field.context = ctx
		default:
			return nil, errors.New("Unsupported field type")
		}
	}

	// Type specific:

	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.

	// Set ourself?
	if parent != nil {
		field.Parent = parent
	}

	field.Annotations = []*PdfAnnotation{}

	// Has a merged-in widget annotation?
	if name := d.GetDirect("Subtype").(*core.PdfObjectName); name != nil {
		if *name == "Widget" {
			// Is a merged field / widget dict.

			// Note that r.newPdfAnnotationFromIndirectObject acts as a caching mechanism if the annotation
			// has been loaded elsewhere already.
			annot, err := r.newPdfAnnotationFromIndirectObject(container)
			if err != nil {
				return nil, err
			}
			widget, ok := annot.GetContext().(*PdfAnnotationWidget)
			if !ok {
				return nil, fmt.Errorf("Invalid widget annotation")
			}
			widget.Parent = field.GetContainingPdfObject()

			field.Annotations = append(field.Annotations, annot)

			return field, nil
		}
	}

	// Kids can be field and/or widget annotations.
	if kids, has := d.GetDirect("Kids").(*core.PdfObjectArray); has {
		field.Kids = []*PdfField{}

		for _, obj := range *kids {
			obj, err := r.traceToObject(obj)
			if err != nil {
				return nil, err
			}

			container, isIndirect := obj.(*core.PdfIndirectObject)
			if !isIndirect {
				return nil, fmt.Errorf("Not an indirect object (form field)")
			}

			dict, ok := container.Direct().(*core.PdfObjectDictionary)
			if !ok {
				return nil, ErrTypeCheck
			}

			// Widget annotations contain key Subtype with value equal to /Widget. Otherwise are assumed to be fields.
			if name, has := dict.GetDirect("Subtype").(*core.PdfObjectName); has && *name == "Widget" {
				widg, err := r.newPdfAnnotationFromIndirectObject(container)
				if err != nil {
					common.Log.Debug("Error loading widget annotation for field: %v", err)
					return nil, err
				}
				field.Annotations = append(field.Annotations, widg)
			} else {
				childf, err := r.newPdfFieldFromIndirectObject(container, field)
				if err != nil {
					common.Log.Debug("Error loading child field: %v", err)
					return nil, err
				}
				field.Kids = append(field.Kids, childf)
			}
		}
	}

	return field, nil
}

// newPdfFieldTextFromDict returns a new PdfFieldText (representing a variable text field) loaded from a dictionary.
// This function loads only text-field specific fields (called by a more generic field loader).
func newPdfFieldTextFromDict(d *core.PdfObjectDictionary) (*PdfFieldText, error) {
	textf := &PdfFieldText{}
	textf.DA, _ = d.GetDirect("DA").(*core.PdfObjectString)
	textf.Q, _ = d.GetDirect("Q").(*core.PdfObjectInteger)
	textf.DS, _ = d.GetDirect("DS").(*core.PdfObjectString)
	textf.RV = d.Get("RV")
	// TODO: MaxLen should be loaded for other fields too?
	textf.MaxLen = d.Get("MaxLen").(*core.PdfObjectInteger)
	return textf, nil
}

// newPdfFieldChoiceFromDict returns a new PdfFieldChoice (representing a choice field) loaded from a dictionary.
// This function loads only choice-field specific fields (called by a more generic field loader).
func newPdfFieldChoiceFromDict(d *core.PdfObjectDictionary) (*PdfFieldChoice, error) {
	choicef := &PdfFieldChoice{}
	choicef.Opt, _ = d.GetDirect("Opt").(*core.PdfObjectArray)
	choicef.TI, _ = d.GetDirect("TI").(*core.PdfObjectInteger)
	choicef.I, _ = d.GetDirect("I").(*core.PdfObjectArray)
	return choicef, nil
}

// newPdfFieldButtonFromDict returns a new PdfFieldButton (representing a button field) loaded from a dictionary.
// This function loads only button-field specific fields (called by a more generic field loader).
func newPdfFieldButtonFromDict(d *core.PdfObjectDictionary) (*PdfFieldButton, error) {
	buttonf := &PdfFieldButton{}
	buttonf.Opt, _ = d.GetDirect("Opt").(*core.PdfObjectArray)
	return buttonf, nil
}
