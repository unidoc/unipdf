/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
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

// String returns a string representation of what flags are set.
func (flag FieldFlag) String() string {
	s := ""
	if flag == FieldFlagClear {
		s = "Clear"
		return s
	}
	if flag&FieldFlagReadOnly > 0 {
		s += "|ReadOnly"
	}
	if flag&FieldFlagRequired > 0 {
		s += "|ReadOnly"
	}
	if flag&FieldFlagNoExport > 0 {
		s += "|NoExport"
	}
	if flag&FieldFlagNoToggleToOff > 0 {
		s += "|NoToggleToOff"
	}
	if flag&FieldFlagRadio > 0 {
		s += "|Radio"
	}
	if flag&FieldFlagPushbutton > 0 {
		s += "|Pushbutton"
	}
	if flag&FieldFlagRadiosInUnision > 0 {
		s += "|RadiosInUnision"
	}
	if flag&FieldFlagMultiline > 0 {
		s += "|Multiline"
	}
	if flag&FieldFlagPassword > 0 {
		s += "|Password"
	}
	if flag&FieldFlagFileSelect > 0 {
		s += "|FileSelect"
	}
	if flag&FieldFlagDoNotScroll > 0 {
		s += "|DoNotScroll"
	}
	if flag&FieldFlagComb > 0 {
		s += "|Comb"
	}
	if flag&FieldFlagRichText > 0 {
		s += "|RichText"
	}
	if flag&FieldFlagDoNotSpellCheck > 0 {
		s += "|DoNotSpellCheck"
	}
	if flag&FieldFlagCombo > 0 {
		s += "|Combo"
	}
	if flag&FieldFlagEdit > 0 {
		s += "|Edit"
	}
	if flag&FieldFlagSort > 0 {
		s += "|Sort"
	}
	if flag&FieldFlagMultiSelect > 0 {
		s += "|MultiSelect"
	}
	if flag&FieldFlagCommitOnSelChange > 0 {
		s += "|CommitOnSelChange"
	}

	return strings.Trim(s, "|")
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
	Annotations []*PdfAnnotationWidget
	Kids        []*PdfField

	FT *core.PdfObjectName
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

	if f.T == nil {
		return fn.String(), errors.New("field partial name (T) not specified")
	}
	parts := []string{f.T.Decoded()}

	// Avoid recursive loops by having a list of already traversed nodes.
	noscanMap := map[*PdfField]bool{}
	noscanMap[f] = true

	parent := f.Parent
	for parent != nil {
		if _, has := noscanMap[parent]; has {
			return fn.String(), errors.New("recursive traversal")
		}

		if parent.T == nil {
			return fn.String(), errors.New("field partial name (T) not specified")
		}
		parts = append(parts, parent.T.Decoded())

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
		partial = f.T.Decoded()
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
// Note: Call the more field context's ToPdfObject to set both the generic and
// non-generic information.
func (f *PdfField) ToPdfObject() core.PdfObject {
	container := f.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	// Create an array of the kids (fields or widgets).
	kids := core.MakeArray()
	for _, child := range f.Kids {
		kids.Append(child.ToPdfObject())
	}
	for _, annot := range f.Annotations {
		if annot.container != f.container {
			kids.Append(annot.GetContext().ToPdfObject())
		}
	}

	// Set fields.
	if f.Parent != nil {
		d.SetIfNotNil("Parent", f.Parent.GetContainingPdfObject())
	}
	if kids.Len() > 0 {
		d.Set("Kids", kids)
	}

	d.SetIfNotNil("FT", f.FT)
	d.SetIfNotNil("T", f.T)
	d.SetIfNotNil("TU", f.TU)
	d.SetIfNotNil("TM", f.TM)
	d.SetIfNotNil("Ff", f.Ff)
	d.SetIfNotNil("V", f.V)
	d.SetIfNotNil("DV", f.DV)
	d.SetIfNotNil("AA", f.AA)

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
	*PdfAnnotationWidget

	V    *PdfSignature
	Lock *core.PdfIndirectObject
	SV   *core.PdfIndirectObject
}

// NewPdfFieldSignature returns an initialized signature field.
func NewPdfFieldSignature(signature *PdfSignature) *PdfFieldSignature {
	field := &PdfFieldSignature{}
	field.PdfField = NewPdfField()
	field.PdfField.SetContext(field)

	field.PdfAnnotationWidget = NewPdfAnnotationWidget()
	field.PdfAnnotationWidget.SetContext(field)
	field.PdfAnnotationWidget.container = field.PdfField.container

	field.T = core.MakeString("")
	field.F = core.MakeInteger(132)
	field.V = signature
	return field
}

// ToPdfObject returns an indirect object containing the signature field dictionary.
func (sig *PdfFieldSignature) ToPdfObject() core.PdfObject {
	// Set general field attributes.
	if sig.PdfAnnotationWidget != nil {
		sig.PdfAnnotationWidget.ToPdfObject()
	}
	sig.PdfField.ToPdfObject()

	// Handle signature field specific attributes.
	container := sig.container

	d := container.PdfObject.(*core.PdfObjectDictionary)
	d.SetIfNotNil("FT", core.MakeName("Sig"))
	d.SetIfNotNil("Lock", sig.Lock)
	d.SetIfNotNil("SV", sig.SV)
	if sig.V != nil {
		d.SetIfNotNil("V", sig.V.ToPdfObject())
	}

	return container
}

// NewPdfField returns an initialized PdfField.
func NewPdfField() *PdfField {
	return &PdfField{
		container: core.MakeIndirectObject(core.MakeDict()),
	}
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
			return false, errors.New("recursive traversal")
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
	return len(f.Kids) == 0
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
		common.Log.Trace("No field flags found - assume clear")
	}

	return flags
}

// SetFlag sets the flag for the field.
func (f *PdfField) SetFlag(flag FieldFlag) {
	f.Ff = core.MakeInteger(int64(flag))
}

// newPdfFieldFromIndirectObject load a field from an indirect object containing the field dictionary.
func (r *PdfReader) newPdfFieldFromIndirectObject(container *core.PdfIndirectObject, parent *PdfField) (*PdfField, error) {
	// If already processed and cached - return processed model.
	if field, cached := r.modelManager.GetModelFromPrimitive(container).(*PdfField); cached {
		return field, nil
	}

	d, isDict := core.GetDict(container)
	if !isDict {
		return nil, fmt.Errorf("PdfField indirect object not containing a dictionary")
	}

	field := NewPdfField()
	field.container = container
	field.container.PdfObject = d

	// Field type (required in terminal fields).
	// Can be /Btn /Tx /Ch /Sig
	// Required for a terminal field (inheritable).

	isTerminal := false
	if name, has := core.GetName(d.Get("FT")); has {
		field.FT = name
		isTerminal = true
	}
	field.isTerminal = &isTerminal

	// Non-terminal field: One whose descendants are fields.  Not an actual field, just a container for inheritable
	// attributes for descendant terminal fields. Does not logically have a type of its own.
	// Terminal field: An actual field with a type.

	// Partial field name.
	field.T, _ = d.Get("T").(*core.PdfObjectString)

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

	// Load type specific fields.
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
		case "Sig":
			ctx, err := r.newPdfFieldSignatureFromDict(d)
			if err != nil {
				return nil, err
			}
			ctx.PdfField = field
			field.context = ctx
		default:
			common.Log.Debug("ERROR: Unsupported field type %s", *field.FT)
			return nil, errors.New("unsupported field type")
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

	field.Annotations = []*PdfAnnotationWidget{}

	// Has a merged-in widget annotation?
	if name, has := core.GetName(d.Get("Subtype")); has {
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
				return nil, errors.New("invalid widget annotation")
			}
			widget.parent = field
			widget.Parent = field.container

			field.Annotations = append(field.Annotations, widget)

			return field, nil
		}
	}

	// Kids can be field and/or widget annotations.
	if kids, has := core.GetArray(d.Get("Kids")); has {
		field.Kids = []*PdfField{}

		for _, obj := range kids.Elements() {
			container, isIndirect := core.GetIndirect(obj)
			if !isIndirect {
				stream, ok := core.GetStream(obj)
				if ok && stream.PdfObjectDictionary != nil {
					nodeType, ok := core.GetNameVal(stream.Get("Type"))
					if ok && nodeType == "Metadata" {
						common.Log.Debug("ERROR: form field Kids array contains invalid Metadata stream. Skipping.")
						continue
					}
				}

				return nil, errors.New("not an indirect object (form field)")
			}

			dict, ok := core.GetDict(container)
			if !ok {
				return nil, ErrTypeCheck
			}

			// Widget annotations contain key Subtype with value equal to /Widget.
			// Otherwise, fields are assumed. Also check for cases in which
			// a widget annotation is the single child of a field and is
			// embedded within the form field instead of being present in the
			// Kids array. In this case, first parse the field and then the
			// widget annotation.
			_, hasFT := core.GetName(dict.Get("FT"))
			if name, has := core.GetName(dict.Get("Subtype")); has && !hasFT && *name == "Widget" {
				annot, err := r.newPdfAnnotationFromIndirectObject(container)
				if err != nil {
					common.Log.Debug("Error loading widget annotation for field: %v", err)
					return nil, err
				}
				wa, ok := annot.context.(*PdfAnnotationWidget)
				if !ok {
					return nil, ErrTypeCheck
				}
				wa.parent = field
				field.Annotations = append(field.Annotations, wa)
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
	textf.DA, _ = core.GetString(d.Get("DA"))
	textf.Q, _ = core.GetInt(d.Get("Q"))
	textf.DS, _ = core.GetString(d.Get("DS"))
	textf.RV = d.Get("RV")
	// TODO: MaxLen should be loaded for other fields too?
	textf.MaxLen, _ = core.GetInt(d.Get("MaxLen"))
	return textf, nil
}

// newPdfFieldChoiceFromDict returns a new PdfFieldChoice (representing a choice field) loaded from a dictionary.
// This function loads only choice-field specific fields (called by a more generic field loader).
func newPdfFieldChoiceFromDict(d *core.PdfObjectDictionary) (*PdfFieldChoice, error) {
	choicef := &PdfFieldChoice{}
	choicef.Opt, _ = core.GetArray(d.Get("Opt"))
	choicef.TI, _ = core.GetInt(d.Get("TI"))
	choicef.I, _ = core.GetArray(d.Get("I"))
	return choicef, nil
}

// newPdfFieldButtonFromDict returns a new PdfFieldButton (representing a button field) loaded from a dictionary.
// This function loads only button-field specific fields (called by a more generic field loader).
func newPdfFieldButtonFromDict(d *core.PdfObjectDictionary) (*PdfFieldButton, error) {
	buttonf := &PdfFieldButton{}
	buttonf.Opt, _ = core.GetArray(d.Get("Opt"))
	return buttonf, nil
}

// newPdfFieldSignatureFromDict returns a new PdfFieldSignature (representing a signature field) loaded from a dictionary.
// This function loads only the signature-specific fields (called by a more generic field loader).
func (r *PdfReader) newPdfFieldSignatureFromDict(d *core.PdfObjectDictionary) (*PdfFieldSignature, error) {
	sigf := &PdfFieldSignature{}

	indobj, has := core.GetIndirect(d.Get("V"))
	if has {
		var err error
		sigf.V, err = r.newPdfSignatureFromIndirect(indobj)
		if err != nil {
			return nil, err
		}
	}

	sigf.Lock, _ = core.GetIndirect(d.Get("Lock"))
	sigf.SV, _ = core.GetIndirect(d.Get("SV"))
	return sigf, nil
}
