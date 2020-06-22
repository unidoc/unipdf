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

// PdfAnnotation represents an annotation in PDF (section 12.5 p. 389).
type PdfAnnotation struct {
	// context contains the specific annotation fields.
	context PdfModel

	// Entries common to all annotation dictionaries (Table 164 p. 391).

	Rect         core.PdfObject
	Contents     core.PdfObject
	P            core.PdfObject // Reference to page object.
	NM           core.PdfObject
	M            core.PdfObject
	F            core.PdfObject
	AP           core.PdfObject
	AS           core.PdfObject
	Border       core.PdfObject
	C            core.PdfObject
	StructParent core.PdfObject
	OC           core.PdfObject

	container *core.PdfIndirectObject
}

// GetContext returns the annotation context which contains the specific type-dependent context.
// The context represents the subannotation.
func (a *PdfAnnotation) GetContext() PdfModel {
	if a == nil {
		return nil
	}
	return a.context
}

// SetContext sets the sub annotation (context).
func (a *PdfAnnotation) SetContext(ctx PdfModel) {
	a.context = ctx
}

func (a *PdfAnnotation) String() string {
	s := ""

	obj, ok := a.ToPdfObject().(*core.PdfIndirectObject)
	if ok {
		s = fmt.Sprintf("%T: %s", a.context, obj.PdfObject.String())
	}

	return s
}

// PdfAnnotationMarkup represents additional fields for mark-up annotations.
// (Section 12.5.6.2 p. 399).
type PdfAnnotationMarkup struct {
	T            core.PdfObject
	Popup        *PdfAnnotationPopup
	CA           core.PdfObject
	RC           core.PdfObject
	CreationDate core.PdfObject
	IRT          core.PdfObject
	Subj         core.PdfObject
	RT           core.PdfObject
	IT           core.PdfObject
	ExData       core.PdfObject
}

// PdfAnnotationText represents Text annotations.
// (Section 12.5.6.4 p. 402).
type PdfAnnotationText struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Open       core.PdfObject
	Name       core.PdfObject
	State      core.PdfObject
	StateModel core.PdfObject
}

// PdfAnnotationLink represents Link annotations.
// (Section 12.5.6.5 p. 403).
type PdfAnnotationLink struct {
	*PdfAnnotation
	A          core.PdfObject
	Dest       core.PdfObject
	H          core.PdfObject
	PA         core.PdfObject
	QuadPoints core.PdfObject
	BS         core.PdfObject

	action *PdfAction
	reader *PdfReader
}

// GetAction returns the PDF action for the annotation link.
func (link *PdfAnnotationLink) GetAction() (*PdfAction, error) {
	if link.action != nil {
		return link.action, nil
	}
	if link.A == nil {
		return nil, nil
	}
	if link.reader == nil {
		return nil, nil
	}

	action, err := link.reader.loadAction(link.A)
	if err != nil {
		return nil, err
	}
	link.action = action

	return link.action, nil
}

// SetAction sets the PDF action for the annotation link.
func (link *PdfAnnotationLink) SetAction(action *PdfAction) {
	link.action = action
	if action == nil {
		link.A = nil
	}
}

// PdfAnnotationFreeText represents FreeText annotations.
// (Section 12.5.6.6).
type PdfAnnotationFreeText struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	DA core.PdfObject
	Q  core.PdfObject
	RC core.PdfObject
	DS core.PdfObject
	CL core.PdfObject
	IT core.PdfObject
	BE core.PdfObject
	RD core.PdfObject
	BS core.PdfObject
	LE core.PdfObject
}

// PdfAnnotationLine represents Line annotations.
// (Section 12.5.6.7).
type PdfAnnotationLine struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	L       core.PdfObject
	BS      core.PdfObject
	LE      core.PdfObject
	IC      core.PdfObject
	LL      core.PdfObject
	LLE     core.PdfObject
	Cap     core.PdfObject
	IT      core.PdfObject
	LLO     core.PdfObject
	CP      core.PdfObject
	Measure core.PdfObject
	CO      core.PdfObject
}

// PdfAnnotationSquare represents Square annotations.
// (Section 12.5.6.8).
type PdfAnnotationSquare struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	BS core.PdfObject
	IC core.PdfObject
	BE core.PdfObject
	RD core.PdfObject
}

// PdfAnnotationCircle represents Circle annotations.
// (Section 12.5.6.8).
type PdfAnnotationCircle struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	BS core.PdfObject
	IC core.PdfObject
	BE core.PdfObject
	RD core.PdfObject
}

// PdfAnnotationPolygon represents Polygon annotations.
// (Section 12.5.6.9).
type PdfAnnotationPolygon struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Vertices core.PdfObject
	LE       core.PdfObject
	BS       core.PdfObject
	IC       core.PdfObject
	BE       core.PdfObject
	IT       core.PdfObject
	Measure  core.PdfObject
}

// PdfAnnotationPolyLine represents PolyLine annotations.
// (Section 12.5.6.9).
type PdfAnnotationPolyLine struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Vertices core.PdfObject
	LE       core.PdfObject
	BS       core.PdfObject
	IC       core.PdfObject
	BE       core.PdfObject
	IT       core.PdfObject
	Measure  core.PdfObject
}

// PdfAnnotationHighlight represents Highlight annotations.
// (Section 12.5.6.10).
type PdfAnnotationHighlight struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints core.PdfObject
}

// PdfAnnotationUnderline represents Underline annotations.
// (Section 12.5.6.10).
type PdfAnnotationUnderline struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints core.PdfObject
}

// PdfAnnotationSquiggly represents Squiggly annotations.
// (Section 12.5.6.10).
type PdfAnnotationSquiggly struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints core.PdfObject
}

// PdfAnnotationStrikeOut represents StrikeOut annotations.
// (Section 12.5.6.10).
type PdfAnnotationStrikeOut struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints core.PdfObject
}

// PdfAnnotationCaret represents Caret annotations.
// (Section 12.5.6.11).
type PdfAnnotationCaret struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	RD core.PdfObject
	Sy core.PdfObject
}

// PdfAnnotationStamp represents Stamp annotations.
// (Section 12.5.6.12).
type PdfAnnotationStamp struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Name core.PdfObject
}

// PdfAnnotationInk represents Ink annotations.
// (Section 12.5.6.13).
type PdfAnnotationInk struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	InkList core.PdfObject
	BS      core.PdfObject
}

// PdfAnnotationPopup represents Popup annotations.
// (Section 12.5.6.14).
type PdfAnnotationPopup struct {
	*PdfAnnotation
	Parent core.PdfObject
	Open   core.PdfObject
}

// PdfAnnotationFileAttachment represents FileAttachment annotations.
// (Section 12.5.6.15).
type PdfAnnotationFileAttachment struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	FS   core.PdfObject
	Name core.PdfObject
}

// PdfAnnotationSound represents Sound annotations.
// (Section 12.5.6.16).
type PdfAnnotationSound struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Sound core.PdfObject
	Name  core.PdfObject
}

// PdfAnnotationRichMedia represents Rich Media annotations.
type PdfAnnotationRichMedia struct {
	*PdfAnnotation
	RichMediaSettings core.PdfObject
	RichMediaContent  core.PdfObject
}

// PdfAnnotationMovie represents Movie annotations.
// (Section 12.5.6.17).
type PdfAnnotationMovie struct {
	*PdfAnnotation
	T     core.PdfObject
	Movie core.PdfObject
	A     core.PdfObject
}

// PdfAnnotationScreen represents Screen annotations.
// (Section 12.5.6.18).
type PdfAnnotationScreen struct {
	*PdfAnnotation
	T  core.PdfObject
	MK core.PdfObject
	A  core.PdfObject
	AA core.PdfObject
}

// PdfAnnotationWidget represents Widget annotations.
// Note: Widget annotations are used to display form fields.
// (Section 12.5.6.19).
type PdfAnnotationWidget struct {
	*PdfAnnotation
	H      core.PdfObject
	MK     core.PdfObject
	A      core.PdfObject
	AA     core.PdfObject
	BS     core.PdfObject
	Parent core.PdfObject

	parent     *PdfField
	processing bool // Used in ToPdfObject serialization to avoid infinite loops for merged-in annots.
}

// Field returns the parent form field of the widget annotation, if one exists.
// NOTE: the method returns nil if the parent form field has not been parsed.
func (widget *PdfAnnotationWidget) Field() *PdfField {
	return widget.parent
}

// PdfAnnotationWatermark represents Watermark annotations.
// (Section 12.5.6.22).
type PdfAnnotationWatermark struct {
	*PdfAnnotation
	FixedPrint core.PdfObject
}

// PdfAnnotationPrinterMark represents PrinterMark annotations.
// (Section 12.5.6.20).
type PdfAnnotationPrinterMark struct {
	*PdfAnnotation
	MN core.PdfObject
}

// PdfAnnotationTrapNet represents TrapNet annotations.
// (Section 12.5.6.21).
type PdfAnnotationTrapNet struct {
	*PdfAnnotation
}

// PdfAnnotation3D represents 3D annotations.
// (Section 13.6.2).
type PdfAnnotation3D struct {
	*PdfAnnotation
	T3DD core.PdfObject
	T3DV core.PdfObject
	T3DA core.PdfObject
	T3DI core.PdfObject
	T3DB core.PdfObject
}

// PdfAnnotationProjection represents Projection annotations.
type PdfAnnotationProjection struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
}

// PdfAnnotationRedact represents Redact annotations.
// (Section 12.5.6.23).
type PdfAnnotationRedact struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints  core.PdfObject
	IC          core.PdfObject
	RO          core.PdfObject
	OverlayText core.PdfObject
	Repeat      core.PdfObject
	DA          core.PdfObject
	Q           core.PdfObject
}

// NewPdfAnnotation returns an initialized generic PDF annotation model.
func NewPdfAnnotation() *PdfAnnotation {
	annot := &PdfAnnotation{}
	annot.container = core.MakeIndirectObject(core.MakeDict())
	return annot
}

// NewPdfAnnotationText returns a new text annotation.
func NewPdfAnnotationText() *PdfAnnotationText {
	annotation := NewPdfAnnotation()
	textAnnotation := &PdfAnnotationText{}
	textAnnotation.PdfAnnotation = annotation
	textAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(textAnnotation)
	return textAnnotation
}

// NewPdfAnnotationLink returns a new link annotation.
func NewPdfAnnotationLink() *PdfAnnotationLink {
	annotation := NewPdfAnnotation()
	linkAnnotation := &PdfAnnotationLink{}
	linkAnnotation.PdfAnnotation = annotation
	annotation.SetContext(linkAnnotation)
	return linkAnnotation
}

// NewPdfAnnotationFreeText returns a new free text annotation.
func NewPdfAnnotationFreeText() *PdfAnnotationFreeText {
	annotation := NewPdfAnnotation()
	freetextAnnotation := &PdfAnnotationFreeText{}
	freetextAnnotation.PdfAnnotation = annotation
	freetextAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(freetextAnnotation)
	return freetextAnnotation
}

// NewPdfAnnotationLine returns a new line annotation.
func NewPdfAnnotationLine() *PdfAnnotationLine {
	annotation := NewPdfAnnotation()
	lineAnnotation := &PdfAnnotationLine{}
	lineAnnotation.PdfAnnotation = annotation
	lineAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(lineAnnotation)
	return lineAnnotation
}

// NewPdfAnnotationSquare returns a new square annotation.
func NewPdfAnnotationSquare() *PdfAnnotationSquare {
	annotation := NewPdfAnnotation()
	rectAnnotation := &PdfAnnotationSquare{}
	rectAnnotation.PdfAnnotation = annotation
	rectAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(rectAnnotation)
	return rectAnnotation
}

// NewPdfAnnotationCircle returns a new circle annotation.
func NewPdfAnnotationCircle() *PdfAnnotationCircle {
	annotation := NewPdfAnnotation()
	circAnnotation := &PdfAnnotationCircle{}
	circAnnotation.PdfAnnotation = annotation
	circAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(circAnnotation)
	return circAnnotation
}

// NewPdfAnnotationPolygon returns a new polygon annotation.
func NewPdfAnnotationPolygon() *PdfAnnotationPolygon {
	annotation := NewPdfAnnotation()
	polygonAnnotation := &PdfAnnotationPolygon{}
	polygonAnnotation.PdfAnnotation = annotation
	polygonAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(polygonAnnotation)
	return polygonAnnotation
}

// NewPdfAnnotationPolyLine returns a new polyline annotation.
func NewPdfAnnotationPolyLine() *PdfAnnotationPolyLine {
	annotation := NewPdfAnnotation()
	polylineAnnotation := &PdfAnnotationPolyLine{}
	polylineAnnotation.PdfAnnotation = annotation
	polylineAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(polylineAnnotation)
	return polylineAnnotation
}

// NewPdfAnnotationHighlight returns a new text highlight annotation.
func NewPdfAnnotationHighlight() *PdfAnnotationHighlight {
	annotation := NewPdfAnnotation()
	highlightAnnotation := &PdfAnnotationHighlight{}
	highlightAnnotation.PdfAnnotation = annotation
	highlightAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(highlightAnnotation)
	return highlightAnnotation
}

// NewPdfAnnotationUnderline returns a new text underline annotation.
func NewPdfAnnotationUnderline() *PdfAnnotationUnderline {
	annotation := NewPdfAnnotation()
	underlineAnnotation := &PdfAnnotationUnderline{}
	underlineAnnotation.PdfAnnotation = annotation
	underlineAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(underlineAnnotation)
	return underlineAnnotation
}

// NewPdfAnnotationSquiggly returns a new text squiggly annotation.
func NewPdfAnnotationSquiggly() *PdfAnnotationSquiggly {
	annotation := NewPdfAnnotation()
	squigglyAnnotation := &PdfAnnotationSquiggly{}
	squigglyAnnotation.PdfAnnotation = annotation
	squigglyAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(squigglyAnnotation)
	return squigglyAnnotation
}

// NewPdfAnnotationStrikeOut returns a new text strikeout annotation.
func NewPdfAnnotationStrikeOut() *PdfAnnotationStrikeOut {
	annotation := NewPdfAnnotation()
	strikeoutAnnotation := &PdfAnnotationStrikeOut{}
	strikeoutAnnotation.PdfAnnotation = annotation
	strikeoutAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(strikeoutAnnotation)
	return strikeoutAnnotation
}

// NewPdfAnnotationCaret returns a new caret annotation.
func NewPdfAnnotationCaret() *PdfAnnotationCaret {
	annotation := NewPdfAnnotation()
	caretAnnotation := &PdfAnnotationCaret{}
	caretAnnotation.PdfAnnotation = annotation
	caretAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(caretAnnotation)
	return caretAnnotation
}

// NewPdfAnnotationStamp returns a new stamp annotation.
func NewPdfAnnotationStamp() *PdfAnnotationStamp {
	annotation := NewPdfAnnotation()
	stampAnnotation := &PdfAnnotationStamp{}
	stampAnnotation.PdfAnnotation = annotation
	stampAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(stampAnnotation)
	return stampAnnotation
}

// NewPdfAnnotationInk returns a new ink annotation.
func NewPdfAnnotationInk() *PdfAnnotationInk {
	annotation := NewPdfAnnotation()
	inkAnnotation := &PdfAnnotationInk{}
	inkAnnotation.PdfAnnotation = annotation
	inkAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(inkAnnotation)
	return inkAnnotation
}

// NewPdfAnnotationPopup returns a new popup annotation.
func NewPdfAnnotationPopup() *PdfAnnotationPopup {
	annotation := NewPdfAnnotation()
	popupAnnotation := &PdfAnnotationPopup{}
	popupAnnotation.PdfAnnotation = annotation
	annotation.SetContext(popupAnnotation)
	return popupAnnotation
}

// NewPdfAnnotationFileAttachment returns a new file attachment annotation.
func NewPdfAnnotationFileAttachment() *PdfAnnotationFileAttachment {
	annotation := NewPdfAnnotation()
	fileAnnotation := &PdfAnnotationFileAttachment{}
	fileAnnotation.PdfAnnotation = annotation
	fileAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(fileAnnotation)
	return fileAnnotation
}

// NewPdfAnnotationSound returns a new sound annotation.
func NewPdfAnnotationSound() *PdfAnnotationSound {
	annotation := NewPdfAnnotation()
	soundAnnotation := &PdfAnnotationSound{}
	soundAnnotation.PdfAnnotation = annotation
	soundAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(soundAnnotation)
	return soundAnnotation
}

// NewPdfAnnotationRichMedia returns a new rich media annotation.
func NewPdfAnnotationRichMedia() *PdfAnnotationRichMedia {
	annotation := NewPdfAnnotation()
	richmediaAnnotation := &PdfAnnotationRichMedia{}
	richmediaAnnotation.PdfAnnotation = annotation
	annotation.SetContext(richmediaAnnotation)
	return richmediaAnnotation
}

// NewPdfAnnotationMovie returns a new movie annotation.
func NewPdfAnnotationMovie() *PdfAnnotationMovie {
	annotation := NewPdfAnnotation()
	movieAnnotation := &PdfAnnotationMovie{}
	movieAnnotation.PdfAnnotation = annotation
	annotation.SetContext(movieAnnotation)
	return movieAnnotation
}

// NewPdfAnnotationScreen returns a new screen annotation.
func NewPdfAnnotationScreen() *PdfAnnotationScreen {
	annotation := NewPdfAnnotation()
	screenAnnotation := &PdfAnnotationScreen{}
	screenAnnotation.PdfAnnotation = annotation
	annotation.SetContext(screenAnnotation)
	return screenAnnotation
}

// NewPdfAnnotationWatermark returns a new watermark annotation.
func NewPdfAnnotationWatermark() *PdfAnnotationWatermark {
	annotation := NewPdfAnnotation()
	watermarkAnnotation := &PdfAnnotationWatermark{}
	watermarkAnnotation.PdfAnnotation = annotation
	annotation.SetContext(watermarkAnnotation)
	return watermarkAnnotation
}

// NewPdfAnnotationPrinterMark returns a new printermark annotation.
func NewPdfAnnotationPrinterMark() *PdfAnnotationPrinterMark {
	annotation := NewPdfAnnotation()
	printermarkAnnotation := &PdfAnnotationPrinterMark{}
	printermarkAnnotation.PdfAnnotation = annotation
	annotation.SetContext(printermarkAnnotation)
	return printermarkAnnotation
}

// NewPdfAnnotationTrapNet returns a new trapnet annotation.
func NewPdfAnnotationTrapNet() *PdfAnnotationTrapNet {
	annotation := NewPdfAnnotation()
	trapnetAnnotation := &PdfAnnotationTrapNet{}
	trapnetAnnotation.PdfAnnotation = annotation
	annotation.SetContext(trapnetAnnotation)
	return trapnetAnnotation
}

// NewPdfAnnotation3D returns a new 3d annotation.
func NewPdfAnnotation3D() *PdfAnnotation3D {
	annotation := NewPdfAnnotation()
	x3dAnnotation := &PdfAnnotation3D{}
	x3dAnnotation.PdfAnnotation = annotation
	annotation.SetContext(x3dAnnotation)
	return x3dAnnotation
}

// NewPdfAnnotationProjection returns a new projection annotation.
func NewPdfAnnotationProjection() *PdfAnnotationProjection {
	annotation := NewPdfAnnotation()
	projectionAnnotation := &PdfAnnotationProjection{}
	projectionAnnotation.PdfAnnotation = annotation
	projectionAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(projectionAnnotation)
	return projectionAnnotation
}

// NewPdfAnnotationRedact returns a new redact annotation.
func NewPdfAnnotationRedact() *PdfAnnotationRedact {
	annotation := NewPdfAnnotation()
	redactAnnotation := &PdfAnnotationRedact{}
	redactAnnotation.PdfAnnotation = annotation
	redactAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(redactAnnotation)
	return redactAnnotation
}

// NewPdfAnnotationWidget returns an initialized annotation widget.
func NewPdfAnnotationWidget() *PdfAnnotationWidget {
	annotation := NewPdfAnnotation()
	annotationWidget := &PdfAnnotationWidget{}
	annotationWidget.PdfAnnotation = annotation
	annotation.SetContext(annotationWidget)
	return annotationWidget
}

// Used for PDF parsing.  Loads a PDF annotation model from a PDF dictionary.
// Loads the common PDF annotation dictionary, and anything needed for the annotation subtype.
func (r *PdfReader) newPdfAnnotationFromIndirectObject(container *core.PdfIndirectObject) (*PdfAnnotation, error) {
	d, isDict := container.PdfObject.(*core.PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("annotation indirect object not containing a dictionary")
	}

	// Check if cached, return cached model if exists.
	if model := r.modelManager.GetModelFromPrimitive(d); model != nil {
		annot, ok := model.(*PdfAnnotation)
		if !ok {
			return nil, fmt.Errorf("cached model not a PDF annotation")
		}
		return annot, nil
	}

	annot := &PdfAnnotation{}
	annot.container = container
	r.modelManager.Register(d, annot)

	if obj := d.Get("Type"); obj != nil {
		str, ok := obj.(*core.PdfObjectName)
		if !ok {
			common.Log.Trace("Incompatibility! Invalid type of Type (%T) - should be Name", obj)
		} else {
			if *str != "Annot" {
				// Log a debug message.
				// Not returning an error on this.
				common.Log.Trace("Unsuspected Type != Annot (%s)", *str)
			}
		}
	}

	if obj := d.Get("Rect"); obj != nil {
		annot.Rect = obj
	}

	if obj := d.Get("Contents"); obj != nil {
		annot.Contents = obj
	}

	if obj := d.Get("P"); obj != nil {
		annot.P = obj
	}

	if obj := d.Get("NM"); obj != nil {
		annot.NM = obj
	}

	if obj := d.Get("M"); obj != nil {
		annot.M = obj
	}

	if obj := d.Get("F"); obj != nil {
		annot.F = obj
	}

	if obj := d.Get("AP"); obj != nil {
		annot.AP = obj
	}

	if obj := d.Get("AS"); obj != nil {
		annot.AS = obj
	}

	if obj := d.Get("Border"); obj != nil {
		annot.Border = obj
	}

	if obj := d.Get("C"); obj != nil {
		annot.C = obj
	}

	if obj := d.Get("StructParent"); obj != nil {
		annot.StructParent = obj
	}

	if obj := d.Get("OC"); obj != nil {
		annot.OC = obj
	}

	subtypeObj := d.Get("Subtype")
	if subtypeObj == nil {
		common.Log.Debug("WARNING: Compatibility issue - annotation Subtype missing - assuming no subtype")
		annot.context = nil
		return annot, nil
	}
	subtype, ok := subtypeObj.(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR: Invalid Subtype object type != name (%T)", subtypeObj)
		return nil, fmt.Errorf("invalid Subtype object type != name (%T)", subtypeObj)
	}
	switch *subtype {
	case "Text":
		ctx, err := r.newPdfAnnotationTextFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Link":
		ctx, err := r.newPdfAnnotationLinkFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "FreeText":
		ctx, err := r.newPdfAnnotationFreeTextFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Line":
		ctx, err := r.newPdfAnnotationLineFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		common.Log.Trace("LINE ANNOTATION: annot (%T): %+v\n", annot, annot)
		common.Log.Trace("LINE ANNOTATION: ctx (%T): %+v\n", ctx, ctx)
		common.Log.Trace("LINE ANNOTATION Markup: ctx (%T): %+v\n", ctx.PdfAnnotationMarkup, ctx.PdfAnnotationMarkup)

		return annot, nil
	case "Square":
		ctx, err := r.newPdfAnnotationSquareFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Circle":
		ctx, err := r.newPdfAnnotationCircleFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Polygon":
		ctx, err := r.newPdfAnnotationPolygonFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "PolyLine":
		ctx, err := r.newPdfAnnotationPolyLineFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Highlight":
		ctx, err := r.newPdfAnnotationHighlightFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Underline":
		ctx, err := r.newPdfAnnotationUnderlineFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Squiggly":
		ctx, err := r.newPdfAnnotationSquigglyFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "StrikeOut":
		ctx, err := r.newPdfAnnotationStrikeOut(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Caret":
		ctx, err := r.newPdfAnnotationCaretFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Stamp":
		ctx, err := r.newPdfAnnotationStampFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Ink":
		ctx, err := r.newPdfAnnotationInkFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Popup":
		ctx, err := r.newPdfAnnotationPopupFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "FileAttachment":
		ctx, err := r.newPdfAnnotationFileAttachmentFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Sound":
		ctx, err := r.newPdfAnnotationSoundFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "RichMedia":
		ctx, err := r.newPdfAnnotationRichMediaFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Movie":
		ctx, err := r.newPdfAnnotationMovieFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Screen":
		ctx, err := r.newPdfAnnotationScreenFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Widget":
		ctx, err := r.newPdfAnnotationWidgetFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "PrinterMark":
		ctx, err := r.newPdfAnnotationPrinterMarkFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "TrapNet":
		ctx, err := r.newPdfAnnotationTrapNetFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Watermark":
		ctx, err := r.newPdfAnnotationWatermarkFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "3D":
		ctx, err := r.newPdfAnnotation3DFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Projection":
		ctx, err := r.newPdfAnnotationProjectionFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	case "Redact":
		ctx, err := r.newPdfAnnotationRedactFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return annot, nil
	}

	common.Log.Debug("ERROR: Ignoring unknown annotation: %s", *subtype)
	return nil, nil
}

// Load data for markup annotation subtypes.
func (r *PdfReader) newPdfAnnotationMarkupFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationMarkup, error) {
	annot := &PdfAnnotationMarkup{}

	if obj := d.Get("T"); obj != nil {
		annot.T = obj
	}

	if obj := d.Get("Popup"); obj != nil {
		indObj, isIndirect := obj.(*core.PdfIndirectObject)
		if !isIndirect {
			if _, isNull := obj.(*core.PdfObjectNull); !isNull {
				return nil, errors.New("popup should point to an indirect object")
			}
		} else {
			popupAnnotObj, err := r.newPdfAnnotationFromIndirectObject(indObj)
			if err != nil {
				return nil, err
			}
			if popupAnnotObj != nil {
				popupAnnot, isPopupAnnot := popupAnnotObj.context.(*PdfAnnotationPopup)
				if !isPopupAnnot {
					return nil, errors.New("object not referring to a popup annotation")
				}
				annot.Popup = popupAnnot
			}
		}
	}

	if obj := d.Get("CA"); obj != nil {
		annot.CA = obj
	}
	if obj := d.Get("RC"); obj != nil {
		annot.RC = obj
	}
	if obj := d.Get("CreationDate"); obj != nil {
		annot.CreationDate = obj
	}
	if obj := d.Get("IRT"); obj != nil {
		annot.IRT = obj
	}
	if obj := d.Get("Subj"); obj != nil {
		annot.Subj = obj
	}
	if obj := d.Get("RT"); obj != nil {
		annot.RT = obj
	}
	if obj := d.Get("IT"); obj != nil {
		annot.IT = obj
	}
	if obj := d.Get("ExData"); obj != nil {
		annot.ExData = obj
	}

	return annot, nil
}

func (r *PdfReader) newPdfAnnotationTextFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationText, error) {
	annot := PdfAnnotationText{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.Open = d.Get("Open")
	annot.Name = d.Get("Name")
	annot.State = d.Get("State")
	annot.StateModel = d.Get("StateModel")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationLinkFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationLink, error) {
	annot := PdfAnnotationLink{}

	annot.A = d.Get("A")
	annot.Dest = d.Get("Dest")
	annot.H = d.Get("H")
	annot.PA = d.Get("PA")
	annot.QuadPoints = d.Get("QuadPoints")
	annot.BS = d.Get("BS")

	return &annot, nil
}

func (r *PdfReader) loadAction(obj core.PdfObject) (*PdfAction, error) {
	if indObj, isIndirect := core.GetIndirect(obj); isIndirect {
		actionObj, err := r.newPdfActionFromIndirectObject(indObj)
		if err != nil {
			return nil, err
		}
		return actionObj, nil
	} else if !core.IsNullObject(obj) {
		return nil, errors.New("action should point to an indirect object")
	}
	return nil, nil
}

func (r *PdfReader) newPdfAnnotationFreeTextFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationFreeText, error) {
	annot := PdfAnnotationFreeText{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.DA = d.Get("DA")
	annot.Q = d.Get("Q")
	annot.RC = d.Get("RC")
	annot.DS = d.Get("DS")
	annot.CL = d.Get("CL")
	annot.IT = d.Get("IT")
	annot.BE = d.Get("BE")
	annot.RD = d.Get("RD")
	annot.BS = d.Get("BS")
	annot.LE = d.Get("LE")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationLineFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationLine, error) {
	annot := PdfAnnotationLine{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.L = d.Get("L")
	annot.BS = d.Get("BS")
	annot.LE = d.Get("LE")
	annot.IC = d.Get("IC")
	annot.LL = d.Get("LL")
	annot.LLE = d.Get("LLE")
	annot.Cap = d.Get("Cap")
	annot.IT = d.Get("IT")
	annot.LLO = d.Get("LLO")
	annot.CP = d.Get("CP")
	annot.Measure = d.Get("Measure")
	annot.CO = d.Get("CO")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationSquareFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationSquare, error) {
	annot := PdfAnnotationSquare{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.BS = d.Get("BS")
	annot.IC = d.Get("IC")
	annot.BE = d.Get("BE")
	annot.RD = d.Get("RD")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationCircleFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationCircle, error) {
	annot := PdfAnnotationCircle{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.BS = d.Get("BS")
	annot.IC = d.Get("IC")
	annot.BE = d.Get("BE")
	annot.RD = d.Get("RD")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPolygonFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationPolygon, error) {
	annot := PdfAnnotationPolygon{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.Vertices = d.Get("Vertices")
	annot.LE = d.Get("LE")
	annot.BS = d.Get("BS")
	annot.IC = d.Get("IC")
	annot.BE = d.Get("BE")
	annot.IT = d.Get("IT")
	annot.Measure = d.Get("Measure")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPolyLineFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationPolyLine, error) {
	annot := PdfAnnotationPolyLine{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.Vertices = d.Get("Vertices")
	annot.LE = d.Get("LE")
	annot.BS = d.Get("BS")
	annot.IC = d.Get("IC")
	annot.BE = d.Get("BE")
	annot.IT = d.Get("IT")
	annot.Measure = d.Get("Measure")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationHighlightFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationHighlight, error) {
	annot := PdfAnnotationHighlight{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationUnderlineFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationUnderline, error) {
	annot := PdfAnnotationUnderline{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationSquigglyFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationSquiggly, error) {
	annot := PdfAnnotationSquiggly{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationStrikeOut(d *core.PdfObjectDictionary) (*PdfAnnotationStrikeOut, error) {
	annot := PdfAnnotationStrikeOut{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationCaretFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationCaret, error) {
	annot := PdfAnnotationCaret{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.RD = d.Get("RD")
	annot.Sy = d.Get("Sy")

	return &annot, nil
}
func (r *PdfReader) newPdfAnnotationStampFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationStamp, error) {
	annot := PdfAnnotationStamp{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.Name = d.Get("Name")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationInkFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationInk, error) {
	annot := PdfAnnotationInk{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.InkList = d.Get("InkList")
	annot.BS = d.Get("BS")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPopupFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationPopup, error) {
	annot := PdfAnnotationPopup{}

	annot.Parent = d.Get("Parent")
	annot.Open = d.Get("Open")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationFileAttachmentFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationFileAttachment, error) {
	annot := PdfAnnotationFileAttachment{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.FS = d.Get("FS")
	annot.Name = d.Get("Name")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationSoundFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationSound, error) {
	annot := PdfAnnotationSound{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.Name = d.Get("Name")
	annot.Sound = d.Get("Sound")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationRichMediaFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationRichMedia, error) {
	annot := &PdfAnnotationRichMedia{}

	annot.RichMediaSettings = d.Get("RichMediaSettings")
	annot.RichMediaContent = d.Get("RichMediaContent")

	return annot, nil
}

func (r *PdfReader) newPdfAnnotationMovieFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationMovie, error) {
	annot := PdfAnnotationMovie{}

	annot.T = d.Get("T")
	annot.Movie = d.Get("Movie")
	annot.A = d.Get("A")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationScreenFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationScreen, error) {
	annot := PdfAnnotationScreen{}

	annot.T = d.Get("T")
	annot.MK = d.Get("MK")
	annot.A = d.Get("A")
	annot.AA = d.Get("AA")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationWidgetFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationWidget, error) {
	annot := PdfAnnotationWidget{}

	annot.H = d.Get("H")

	annot.MK = d.Get("MK")
	// MK can be an indirect object...
	// Expected to be a dictionary.

	annot.A = d.Get("A")
	annot.AA = d.Get("AA")
	annot.BS = d.Get("BS")
	annot.Parent = d.Get("Parent")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPrinterMarkFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationPrinterMark, error) {
	annot := PdfAnnotationPrinterMark{}

	annot.MN = d.Get("MN")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationTrapNetFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationTrapNet, error) {
	annot := PdfAnnotationTrapNet{}
	// empty?e
	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationWatermarkFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationWatermark, error) {
	annot := PdfAnnotationWatermark{}

	annot.FixedPrint = d.Get("FixedPrint")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotation3DFromDict(d *core.PdfObjectDictionary) (*PdfAnnotation3D, error) {
	annot := PdfAnnotation3D{}

	annot.T3DD = d.Get("3DD")
	annot.T3DV = d.Get("3DV")
	annot.T3DA = d.Get("3DA")
	annot.T3DI = d.Get("3DI")
	annot.T3DB = d.Get("3DB")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationProjectionFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationProjection, error) {
	annot := &PdfAnnotationProjection{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	return annot, nil
}

func (r *PdfReader) newPdfAnnotationRedactFromDict(d *core.PdfObjectDictionary) (*PdfAnnotationRedact, error) {
	annot := PdfAnnotationRedact{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")
	annot.IC = d.Get("IC")
	annot.RO = d.Get("RO")
	annot.OverlayText = d.Get("OverlayText")
	annot.Repeat = d.Get("Repeat")
	annot.DA = d.Get("DA")
	annot.Q = d.Get("Q")

	return &annot, nil
}

// GetContainingPdfObject implements interface PdfModel.
func (a *PdfAnnotation) GetContainingPdfObject() core.PdfObject {
	return a.container
}

// ToPdfObject implements interface PdfModel.
// Note: Call the sub-annotation's ToPdfObject to set both the generic and non-generic information.
func (a *PdfAnnotation) ToPdfObject() core.PdfObject {
	container := a.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.Clear()

	d.Set("Type", core.MakeName("Annot"))
	d.SetIfNotNil("Rect", a.Rect)
	d.SetIfNotNil("Contents", a.Contents)
	d.SetIfNotNil("P", a.P)
	d.SetIfNotNil("NM", a.NM)
	d.SetIfNotNil("M", a.M)
	d.SetIfNotNil("F", a.F)
	d.SetIfNotNil("AP", a.AP)
	d.SetIfNotNil("AS", a.AS)
	d.SetIfNotNil("Border", a.Border)
	d.SetIfNotNil("C", a.C)
	d.SetIfNotNil("StructParent", a.StructParent)
	d.SetIfNotNil("OC", a.OC)

	return container
}

// appendToPdfDictionary appends the markup portion of the annotation to the dictionary `d`.
func (markup *PdfAnnotationMarkup) appendToPdfDictionary(d *core.PdfObjectDictionary) {
	d.SetIfNotNil("T", markup.T)
	if markup.Popup != nil {
		d.Set("Popup", markup.Popup.ToPdfObject())
	}
	d.SetIfNotNil("CA", markup.CA)
	d.SetIfNotNil("RC", markup.RC)
	d.SetIfNotNil("CreationDate", markup.CreationDate)
	d.SetIfNotNil("IRT", markup.IRT)
	d.SetIfNotNil("Subj", markup.Subj)
	d.SetIfNotNil("RT", markup.RT)
	d.SetIfNotNil("IT", markup.IT)
	d.SetIfNotNil("ExData", markup.ExData)
}

// ToPdfObject implements interface PdfModel.
func (text *PdfAnnotationText) ToPdfObject() core.PdfObject {
	text.PdfAnnotation.ToPdfObject()
	container := text.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	if text.PdfAnnotationMarkup != nil {
		text.PdfAnnotationMarkup.appendToPdfDictionary(d)
	}

	d.SetIfNotNil("Subtype", core.MakeName("Text"))
	d.SetIfNotNil("Open", text.Open)
	d.SetIfNotNil("Name", text.Name)
	d.SetIfNotNil("State", text.State)
	d.SetIfNotNil("StateModel", text.StateModel)
	return container
}

// ToPdfObject implements interface PdfModel.
func (link *PdfAnnotationLink) ToPdfObject() core.PdfObject {
	link.PdfAnnotation.ToPdfObject()
	container := link.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("Link"))

	if link.action != nil && link.action.context != nil {
		d.Set("A", link.action.context.ToPdfObject())
	} else if link.A != nil {
		d.Set("A", link.A)
	}

	d.SetIfNotNil("Dest", link.Dest)
	d.SetIfNotNil("H", link.H)
	d.SetIfNotNil("PA", link.PA)
	d.SetIfNotNil("QuadPoints", link.QuadPoints)
	d.SetIfNotNil("BS", link.BS)
	return container
}

// ToPdfObject implements interface PdfModel.
func (ft *PdfAnnotationFreeText) ToPdfObject() core.PdfObject {
	ft.PdfAnnotation.ToPdfObject()
	container := ft.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	ft.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("FreeText"))
	d.SetIfNotNil("DA", ft.DA)
	d.SetIfNotNil("Q", ft.Q)
	d.SetIfNotNil("RC", ft.RC)
	d.SetIfNotNil("DS", ft.DS)
	d.SetIfNotNil("CL", ft.CL)
	d.SetIfNotNil("IT", ft.IT)
	d.SetIfNotNil("BE", ft.BE)
	d.SetIfNotNil("RD", ft.RD)
	d.SetIfNotNil("BS", ft.BS)
	d.SetIfNotNil("LE", ft.LE)

	return container
}

// ToPdfObject implements interface PdfModel.
func (line *PdfAnnotationLine) ToPdfObject() core.PdfObject {
	line.PdfAnnotation.ToPdfObject()
	container := line.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	line.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Line"))
	d.SetIfNotNil("L", line.L)
	d.SetIfNotNil("BS", line.BS)
	d.SetIfNotNil("LE", line.LE)
	d.SetIfNotNil("IC", line.IC)
	d.SetIfNotNil("LL", line.LL)
	d.SetIfNotNil("LLE", line.LLE)
	d.SetIfNotNil("Cap", line.Cap)
	d.SetIfNotNil("IT", line.IT)
	d.SetIfNotNil("LLO", line.LLO)
	d.SetIfNotNil("CP", line.CP)
	d.SetIfNotNil("Measure", line.Measure)
	d.SetIfNotNil("CO", line.CO)

	return container
}

// ToPdfObject implements interface PdfModel.
func (sq *PdfAnnotationSquare) ToPdfObject() core.PdfObject {
	sq.PdfAnnotation.ToPdfObject()
	container := sq.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	if sq.PdfAnnotationMarkup != nil {
		sq.PdfAnnotationMarkup.appendToPdfDictionary(d)
	}

	d.SetIfNotNil("Subtype", core.MakeName("Square"))
	d.SetIfNotNil("BS", sq.BS)
	d.SetIfNotNil("IC", sq.IC)
	d.SetIfNotNil("BE", sq.BE)
	d.SetIfNotNil("RD", sq.RD)

	return container
}

// ToPdfObject implements interface PdfModel.
func (circle *PdfAnnotationCircle) ToPdfObject() core.PdfObject {
	circle.PdfAnnotation.ToPdfObject()
	container := circle.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	circle.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Circle"))
	d.SetIfNotNil("BS", circle.BS)
	d.SetIfNotNil("IC", circle.IC)
	d.SetIfNotNil("BE", circle.BE)
	d.SetIfNotNil("RD", circle.RD)

	return container
}

// ToPdfObject implements interface PdfModel.
func (poly *PdfAnnotationPolygon) ToPdfObject() core.PdfObject {
	poly.PdfAnnotation.ToPdfObject()
	container := poly.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	poly.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Polygon"))
	d.SetIfNotNil("Vertices", poly.Vertices)
	d.SetIfNotNil("LE", poly.LE)
	d.SetIfNotNil("BS", poly.BS)
	d.SetIfNotNil("IC", poly.IC)
	d.SetIfNotNil("BE", poly.BE)
	d.SetIfNotNil("IT", poly.IT)
	d.SetIfNotNil("Measure", poly.Measure)

	return container
}

// ToPdfObject implements interface PdfModel.
func (polyl *PdfAnnotationPolyLine) ToPdfObject() core.PdfObject {
	polyl.PdfAnnotation.ToPdfObject()
	container := polyl.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	polyl.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("PolyLine"))
	d.SetIfNotNil("Vertices", polyl.Vertices)
	d.SetIfNotNil("LE", polyl.LE)
	d.SetIfNotNil("BS", polyl.BS)
	d.SetIfNotNil("IC", polyl.IC)
	d.SetIfNotNil("BE", polyl.BE)
	d.SetIfNotNil("IT", polyl.IT)
	d.SetIfNotNil("Measure", polyl.Measure)

	return container
}

// ToPdfObject implements interface PdfModel.
func (hl *PdfAnnotationHighlight) ToPdfObject() core.PdfObject {
	hl.PdfAnnotation.ToPdfObject()
	container := hl.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	hl.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Highlight"))
	d.SetIfNotNil("QuadPoints", hl.QuadPoints)
	return container
}

// ToPdfObject implements interface PdfModel.
func (underline *PdfAnnotationUnderline) ToPdfObject() core.PdfObject {
	underline.PdfAnnotation.ToPdfObject()
	container := underline.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	underline.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Underline"))
	d.SetIfNotNil("QuadPoints", underline.QuadPoints)
	return container
}

// ToPdfObject implements interface PdfModel.
func (squiggly *PdfAnnotationSquiggly) ToPdfObject() core.PdfObject {
	squiggly.PdfAnnotation.ToPdfObject()
	container := squiggly.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	squiggly.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Squiggly"))
	d.SetIfNotNil("QuadPoints", squiggly.QuadPoints)
	return container
}

// ToPdfObject implements interface PdfModel.
func (strikeo *PdfAnnotationStrikeOut) ToPdfObject() core.PdfObject {
	strikeo.PdfAnnotation.ToPdfObject()
	container := strikeo.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	strikeo.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("StrikeOut"))
	d.SetIfNotNil("QuadPoints", strikeo.QuadPoints)
	return container
}

// ToPdfObject implements interface PdfModel.
func (caret *PdfAnnotationCaret) ToPdfObject() core.PdfObject {
	caret.PdfAnnotation.ToPdfObject()
	container := caret.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	caret.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Caret"))
	d.SetIfNotNil("RD", caret.RD)
	d.SetIfNotNil("Sy", caret.Sy)
	return container
}

// ToPdfObject implements interface PdfModel.
func (stamp *PdfAnnotationStamp) ToPdfObject() core.PdfObject {
	stamp.PdfAnnotation.ToPdfObject()
	container := stamp.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	stamp.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Stamp"))
	d.SetIfNotNil("Name", stamp.Name)
	return container
}

// ToPdfObject implements interface PdfModel.
func (ink *PdfAnnotationInk) ToPdfObject() core.PdfObject {
	ink.PdfAnnotation.ToPdfObject()
	container := ink.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	ink.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Ink"))
	d.SetIfNotNil("InkList", ink.InkList)
	d.SetIfNotNil("BS", ink.BS)
	return container
}

// ToPdfObject implements interface PdfModel.
func (popup *PdfAnnotationPopup) ToPdfObject() core.PdfObject {
	popup.PdfAnnotation.ToPdfObject()
	container := popup.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("Popup"))
	d.SetIfNotNil("Parent", popup.Parent)
	d.SetIfNotNil("Open", popup.Open)
	return container
}

// ToPdfObject implements interface PdfModel.
func (file *PdfAnnotationFileAttachment) ToPdfObject() core.PdfObject {
	file.PdfAnnotation.ToPdfObject()
	container := file.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	file.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("FileAttachment"))
	d.SetIfNotNil("FS", file.FS)
	d.SetIfNotNil("Name", file.Name)
	return container
}

// ToPdfObject implements interface PdfModel.
func (rm *PdfAnnotationRichMedia) ToPdfObject() core.PdfObject {
	rm.PdfAnnotation.ToPdfObject()
	container := rm.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("RichMedia"))
	d.SetIfNotNil("RichMediaSettings", rm.RichMediaSettings)
	d.SetIfNotNil("RichMediaContent", rm.RichMediaContent)
	return container
}

// ToPdfObject implements interface PdfModel.
func (snd *PdfAnnotationSound) ToPdfObject() core.PdfObject {
	snd.PdfAnnotation.ToPdfObject()
	container := snd.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	snd.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Sound"))
	d.SetIfNotNil("Sound", snd.Sound)
	d.SetIfNotNil("Name", snd.Name)
	return container
}

// ToPdfObject implements interface PdfModel.
func (mov *PdfAnnotationMovie) ToPdfObject() core.PdfObject {
	mov.PdfAnnotation.ToPdfObject()
	container := mov.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("Movie"))
	d.SetIfNotNil("T", mov.T)
	d.SetIfNotNil("Movie", mov.Movie)
	d.SetIfNotNil("A", mov.A)
	return container
}

// ToPdfObject implements interface PdfModel.
func (scr *PdfAnnotationScreen) ToPdfObject() core.PdfObject {
	scr.PdfAnnotation.ToPdfObject()
	container := scr.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("Screen"))
	d.SetIfNotNil("T", scr.T)
	d.SetIfNotNil("MK", scr.MK)
	d.SetIfNotNil("A", scr.A)
	d.SetIfNotNil("AA", scr.AA)
	return container
}

// ToPdfObject implements interface PdfModel.
func (widget *PdfAnnotationWidget) ToPdfObject() core.PdfObject {
	widget.PdfAnnotation.ToPdfObject()
	container := widget.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	if widget.processing {
		// Avoid recursion for merged-in annotations (calling ToPdfObject from both field and annotation).
		return container
	}
	widget.processing = true

	d.SetIfNotNil("Subtype", core.MakeName("Widget"))
	d.SetIfNotNil("H", widget.H)
	d.SetIfNotNil("MK", widget.MK)
	d.SetIfNotNil("A", widget.A)
	d.SetIfNotNil("AA", widget.AA)
	d.SetIfNotNil("BS", widget.BS)

	parentObj := widget.Parent
	if widget.parent != nil {
		if widget.parent.container == widget.container {
			// Populate the part from the field.
			widget.parent.ToPdfObject()
		}
		parentObj = widget.parent.GetContainingPdfObject()
	}
	if parentObj != container {
		// Accommodate merged-in field/annotations.
		d.SetIfNotNil("Parent", parentObj)
	}

	widget.processing = false

	return container
}

// ToPdfObject implements interface PdfModel.
func (pm *PdfAnnotationPrinterMark) ToPdfObject() core.PdfObject {
	pm.PdfAnnotation.ToPdfObject()
	container := pm.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("PrinterMark"))
	d.SetIfNotNil("MN", pm.MN)
	return container
}

// ToPdfObject implements interface PdfModel.
func (trapn *PdfAnnotationTrapNet) ToPdfObject() core.PdfObject {
	trapn.PdfAnnotation.ToPdfObject()
	container := trapn.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("TrapNet"))
	return container
}

// ToPdfObject implements interface PdfModel.
func (wm *PdfAnnotationWatermark) ToPdfObject() core.PdfObject {
	wm.PdfAnnotation.ToPdfObject()
	container := wm.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.SetIfNotNil("Subtype", core.MakeName("Watermark"))
	d.SetIfNotNil("FixedPrint", wm.FixedPrint)

	return container
}

// ToPdfObject implements interface PdfModel.
func (a3d *PdfAnnotation3D) ToPdfObject() core.PdfObject {
	a3d.PdfAnnotation.ToPdfObject()
	container := a3d.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	d.SetIfNotNil("Subtype", core.MakeName("3D"))
	d.SetIfNotNil("3DD", a3d.T3DD)
	d.SetIfNotNil("3DV", a3d.T3DV)
	d.SetIfNotNil("3DA", a3d.T3DA)
	d.SetIfNotNil("3DI", a3d.T3DI)
	d.SetIfNotNil("3DB", a3d.T3DB)
	return container
}

// ToPdfObject implements interface PdfModel.
func (proj *PdfAnnotationProjection) ToPdfObject() core.PdfObject {
	proj.PdfAnnotation.ToPdfObject()
	container := proj.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	proj.PdfAnnotationMarkup.appendToPdfDictionary(d)
	return container
}

// ToPdfObject implements interface PdfModel.
func (redact *PdfAnnotationRedact) ToPdfObject() core.PdfObject {
	redact.PdfAnnotation.ToPdfObject()
	container := redact.container
	d := container.PdfObject.(*core.PdfObjectDictionary)
	redact.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", core.MakeName("Redact"))
	d.SetIfNotNil("QuadPoints", redact.QuadPoints)
	d.SetIfNotNil("IC", redact.IC)
	d.SetIfNotNil("RO", redact.RO)
	d.SetIfNotNil("OverlayText", redact.OverlayText)
	d.SetIfNotNil("Repeat", redact.Repeat)
	d.SetIfNotNil("DA", redact.DA)
	d.SetIfNotNil("Q", redact.Q)
	return container
}

// Border definitions.

// BorderStyle defines border type, typically used for annotations.
type BorderStyle int

const (
	// BorderStyleSolid represents a solid border.
	BorderStyleSolid BorderStyle = iota

	// BorderStyleDashed represents a dashed border.
	BorderStyleDashed BorderStyle = iota

	// BorderStyleBeveled represents a beveled border.
	BorderStyleBeveled BorderStyle = iota

	// BorderStyleInset represents an inset border.
	BorderStyleInset BorderStyle = iota

	// BorderStyleUnderline represents an underline border.
	BorderStyleUnderline BorderStyle = iota
)

// GetPdfName returns the PDF name used to indicate the border style.
// (Table 166 p. 395).
func (bs *BorderStyle) GetPdfName() string {
	switch *bs {
	case BorderStyleSolid:
		return "S"
	case BorderStyleDashed:
		return "D"
	case BorderStyleBeveled:
		return "B"
	case BorderStyleInset:
		return "I"
	case BorderStyleUnderline:
		return "U"
	}

	return "" // Should not happen.
}

// PdfBorderStyle represents a border style dictionary (12.5.4 Border Styles p. 394).
type PdfBorderStyle struct {
	W *float64     // Border width
	S *BorderStyle // Border style
	D *[]int       // Dash array.

	container core.PdfObject
}

// NewBorderStyle returns an initialized PdfBorderStyle.
func NewBorderStyle() *PdfBorderStyle {
	bs := &PdfBorderStyle{}
	return bs
}

// SetBorderWidth sets the style's border width.
func (bs *PdfBorderStyle) SetBorderWidth(width float64) {
	bs.W = &width
}

// GetBorderWidth returns the border style's width.
func (bs *PdfBorderStyle) GetBorderWidth() float64 {
	if bs.W == nil {
		return 1 // Default.
	}
	return *bs.W
}

// newPdfBorderStyleFromPdfObject is used to load a PdfBorderStyle from a dictionary.
func newPdfBorderStyleFromPdfObject(obj core.PdfObject) (*PdfBorderStyle, error) {
	bs := &PdfBorderStyle{}
	bs.container = obj

	var d *core.PdfObjectDictionary
	obj = core.TraceToDirectObject(obj)
	d, ok := obj.(*core.PdfObjectDictionary)
	if !ok {
		return nil, errors.New("type check")
	}

	// Type.
	if obj := d.Get("Type"); obj != nil {
		name, ok := obj.(*core.PdfObjectName)
		if !ok {
			common.Log.Debug("Incompatibility with Type not a name object: %T", obj)
		} else {
			if *name != "Border" {
				common.Log.Debug("Warning, Type != Border: %s", *name)
			}
		}
	}

	// Border width.
	if obj := d.Get("W"); obj != nil {
		val, err := core.GetNumberAsFloat(obj)
		if err != nil {
			common.Log.Debug("Error retrieving W: %v", err)
			return nil, err
		}
		bs.W = &val
	}

	// Border style.
	if obj := d.Get("S"); obj != nil {
		name, ok := obj.(*core.PdfObjectName)
		if !ok {
			return nil, errors.New("border S not a name object")
		}

		var style BorderStyle
		switch *name {
		case "S":
			style = BorderStyleSolid
		case "D":
			style = BorderStyleDashed
		case "B":
			style = BorderStyleBeveled
		case "I":
			style = BorderStyleInset
		case "U":
			style = BorderStyleUnderline
		default:
			common.Log.Debug("Invalid style name %s", *name)
			return nil, errors.New("style type range check")
		}

		bs.S = &style
	}

	// Dash array.
	if obj := d.Get("D"); obj != nil {
		vec, ok := obj.(*core.PdfObjectArray)
		if !ok {
			common.Log.Debug("Border D dash not an array: %T", obj)
			return nil, errors.New("border D type check error")
		}

		vals, err := vec.ToIntegerArray()
		if err != nil {
			common.Log.Debug("Border D Problem converting to integer array: %v", err)
			return nil, err
		}

		bs.D = &vals
	}

	return bs, nil
}

// ToPdfObject implements interface PdfModel.
func (bs *PdfBorderStyle) ToPdfObject() core.PdfObject {
	d := core.MakeDict()
	if bs.container != nil {
		if indObj, is := bs.container.(*core.PdfIndirectObject); is {
			indObj.PdfObject = d
		}
	}

	d.Set("Subtype", core.MakeName("Border"))
	if bs.W != nil {
		d.Set("W", core.MakeFloat(*bs.W))
	}
	if bs.S != nil {
		d.Set("S", core.MakeName(bs.S.GetPdfName()))
	}
	if bs.D != nil {
		d.Set("D", core.MakeArrayFromIntegers(*bs.D))
	}

	if bs.container != nil {
		return bs.container
	}
	return d
}

// BorderEffect represents a border effect (Table 167 p. 395).
type BorderEffect int

const (
	// BorderEffectNoEffect represents no border effect.
	BorderEffectNoEffect BorderEffect = iota

	// BorderEffectCloudy represents a cloudy border effect.
	BorderEffectCloudy BorderEffect = iota
)

// PdfBorderEffect represents a PDF border effect.
type PdfBorderEffect struct {
	S *BorderEffect // Border effect type
	I *float64      // Intensity of the effect
}
