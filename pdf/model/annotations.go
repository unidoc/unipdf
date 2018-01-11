/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

// PDFAnnotation contains common attributes of an annotation.  The context object contains the subannotation,
// which can be a markup annotation or other types.
type PdfAnnotation struct {
	context      PdfModel // Sub-annotation.
	Rect         PdfObject
	Contents     PdfObject
	P            PdfObject // Reference to page object.
	NM           PdfObject
	M            PdfObject
	F            PdfObject
	AP           PdfObject
	AS           PdfObject
	Border       PdfObject
	C            PdfObject
	StructParent PdfObject
	OC           PdfObject

	primitive *PdfIndirectObject
}

// Context in this case is a reference to the subannotation.
func (this *PdfAnnotation) GetContext() PdfModel {
	return this.context
}

// Set the sub annotation (context).
func (this *PdfAnnotation) SetContext(ctx PdfModel) {
	this.context = ctx
}

func (this *PdfAnnotation) String() string {
	s := ""

	obj, ok := this.ToPdfObject().(*PdfIndirectObject)
	if ok {
		s = fmt.Sprintf("%T: %s", this.context, obj.PdfObject.String())
	}

	return s
}

// Additional elements for mark-up annotations.
type PdfAnnotationMarkup struct {
	T            PdfObject
	Popup        *PdfAnnotationPopup
	CA           PdfObject
	RC           PdfObject
	CreationDate PdfObject
	IRT          PdfObject
	Subj         PdfObject
	RT           PdfObject
	IT           PdfObject
	ExData       PdfObject
}

// Subtype: Text
type PdfAnnotationText struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Open       PdfObject
	Name       PdfObject
	State      PdfObject
	StateModel PdfObject
}

// Subtype: Link
type PdfAnnotationLink struct {
	*PdfAnnotation
	A          PdfObject
	Dest       PdfObject
	H          PdfObject
	PA         PdfObject
	QuadPoints PdfObject
	BS         PdfObject
}

// Subtype: FreeText
type PdfAnnotationFreeText struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	DA PdfObject
	Q  PdfObject
	RC PdfObject
	DS PdfObject
	CL PdfObject
	IT PdfObject
	BE PdfObject
	RD PdfObject
	BS PdfObject
	LE PdfObject
}

// Subtype: Line
type PdfAnnotationLine struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	L       PdfObject
	BS      PdfObject
	LE      PdfObject
	IC      PdfObject
	LL      PdfObject
	LLE     PdfObject
	Cap     PdfObject
	IT      PdfObject
	LLO     PdfObject
	CP      PdfObject
	Measure PdfObject
	CO      PdfObject
}

// Subtype: Square
type PdfAnnotationSquare struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	BS PdfObject
	IC PdfObject
	BE PdfObject
	RD PdfObject
}

// Subtype: Circle
type PdfAnnotationCircle struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	BS PdfObject
	IC PdfObject
	BE PdfObject
	RD PdfObject
}

// Subtype: Polygon
type PdfAnnotationPolygon struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Vertices PdfObject
	LE       PdfObject
	BS       PdfObject
	IC       PdfObject
	BE       PdfObject
	IT       PdfObject
	Measure  PdfObject
}

// Subtype: PolyLine
type PdfAnnotationPolyLine struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Vertices PdfObject
	LE       PdfObject
	BS       PdfObject
	IC       PdfObject
	BE       PdfObject
	IT       PdfObject
	Measure  PdfObject
}

// Subtype: Highlight
type PdfAnnotationHighlight struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints PdfObject
}

// Subtype: Underline
type PdfAnnotationUnderline struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints PdfObject
}

// Subtype: Squiggly
type PdfAnnotationSquiggly struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints PdfObject
}

// Subtype: StrikeOut
type PdfAnnotationStrikeOut struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints PdfObject
}

// Subtype: Caret
type PdfAnnotationCaret struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	RD PdfObject
	Sy PdfObject
}

// Subtype: Stamp
type PdfAnnotationStamp struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Name PdfObject
}

// Subtype: Ink
type PdfAnnotationInk struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	InkList PdfObject
	BS      PdfObject
}

// Subtype: Popup
type PdfAnnotationPopup struct {
	*PdfAnnotation
	Parent PdfObject
	Open   PdfObject
}

// Subtype: FileAttachment
type PdfAnnotationFileAttachment struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	FS   PdfObject
	Name PdfObject
}

// Subtype: Sound
type PdfAnnotationSound struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	Sound PdfObject
	Name  PdfObject
}

// Subtype: Rich Media
type PdfAnnotationRichMedia struct {
	*PdfAnnotation
	RichMediaSettings PdfObject
	RichMediaContent  PdfObject
}

// Subtype: Movie
type PdfAnnotationMovie struct {
	*PdfAnnotation
	T     PdfObject
	Movie PdfObject
	A     PdfObject
}

// Subtype: Screen
type PdfAnnotationScreen struct {
	*PdfAnnotation
	T  PdfObject
	MK PdfObject
	A  PdfObject
	AA PdfObject
}

// Subtype: Widget
type PdfAnnotationWidget struct {
	*PdfAnnotation
	H      PdfObject
	MK     PdfObject
	A      PdfObject
	AA     PdfObject
	BS     PdfObject
	Parent PdfObject
}

// Subtype: Watermark
type PdfAnnotationWatermark struct {
	*PdfAnnotation
	FixedPrint PdfObject
}

// Subtype: PrinterMark
type PdfAnnotationPrinterMark struct {
	*PdfAnnotation
	MN PdfObject
}

// Subtype: TrapNet
type PdfAnnotationTrapNet struct {
	*PdfAnnotation
}

// Subtype: 3D
type PdfAnnotation3D struct {
	*PdfAnnotation
	T3DD PdfObject
	T3DV PdfObject
	T3DA PdfObject
	T3DI PdfObject
	T3DB PdfObject
}

// Subtype: Projection
type PdfAnnotationProjection struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
}

// Subtype: Redact
type PdfAnnotationRedact struct {
	*PdfAnnotation
	*PdfAnnotationMarkup
	QuadPoints  PdfObject
	IC          PdfObject
	RO          PdfObject
	OverlayText PdfObject
	Repeat      PdfObject
	DA          PdfObject
	Q           PdfObject
}

// Construct a new PDF annotation model and initializes the underlying PDF primitive.
func NewPdfAnnotation() *PdfAnnotation {
	annot := &PdfAnnotation{}

	container := &PdfIndirectObject{}
	container.PdfObject = MakeDict()

	annot.primitive = container
	return annot
}

// Create a new text annotation.
func NewPdfAnnotationText() *PdfAnnotationText {
	annotation := NewPdfAnnotation()
	textAnnotation := &PdfAnnotationText{}
	textAnnotation.PdfAnnotation = annotation
	textAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(textAnnotation)
	return textAnnotation
}

// Create a new link annotation.
func NewPdfAnnotationLink() *PdfAnnotationLink {
	annotation := NewPdfAnnotation()
	linkAnnotation := &PdfAnnotationLink{}
	linkAnnotation.PdfAnnotation = annotation
	annotation.SetContext(linkAnnotation)
	return linkAnnotation
}

// Create a new free text annotation.
func NewPdfAnnotationFreeText() *PdfAnnotationFreeText {
	annotation := NewPdfAnnotation()
	freetextAnnotation := &PdfAnnotationFreeText{}
	freetextAnnotation.PdfAnnotation = annotation
	freetextAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(freetextAnnotation)
	return freetextAnnotation
}

// Create a new line annotation.
func NewPdfAnnotationLine() *PdfAnnotationLine {
	annotation := NewPdfAnnotation()
	lineAnnotation := &PdfAnnotationLine{}
	lineAnnotation.PdfAnnotation = annotation
	lineAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(lineAnnotation)
	return lineAnnotation
}

// Create a new square annotation.
func NewPdfAnnotationSquare() *PdfAnnotationSquare {
	annotation := NewPdfAnnotation()
	rectAnnotation := &PdfAnnotationSquare{}
	rectAnnotation.PdfAnnotation = annotation
	rectAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(rectAnnotation)
	return rectAnnotation
}

// Create a new circle annotation.
func NewPdfAnnotationCircle() *PdfAnnotationCircle {
	annotation := NewPdfAnnotation()
	circAnnotation := &PdfAnnotationCircle{}
	circAnnotation.PdfAnnotation = annotation
	circAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(circAnnotation)
	return circAnnotation
}

// Create a new polygon annotation.
func NewPdfAnnotationPolygon() *PdfAnnotationPolygon {
	annotation := NewPdfAnnotation()
	polygonAnnotation := &PdfAnnotationPolygon{}
	polygonAnnotation.PdfAnnotation = annotation
	polygonAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(polygonAnnotation)
	return polygonAnnotation
}

// Create a new polyline annotation.
func NewPdfAnnotationPolyLine() *PdfAnnotationPolyLine {
	annotation := NewPdfAnnotation()
	polylineAnnotation := &PdfAnnotationPolyLine{}
	polylineAnnotation.PdfAnnotation = annotation
	polylineAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(polylineAnnotation)
	return polylineAnnotation
}

// Create a new text highlight annotation.
func NewPdfAnnotationHighlight() *PdfAnnotationHighlight {
	annotation := NewPdfAnnotation()
	highlightAnnotation := &PdfAnnotationHighlight{}
	highlightAnnotation.PdfAnnotation = annotation
	highlightAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(highlightAnnotation)
	return highlightAnnotation
}

// Create a new text underline annotation.
func NewPdfAnnotationUnderline() *PdfAnnotationUnderline {
	annotation := NewPdfAnnotation()
	underlineAnnotation := &PdfAnnotationUnderline{}
	underlineAnnotation.PdfAnnotation = annotation
	underlineAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(underlineAnnotation)
	return underlineAnnotation
}

// Create a new text squiggly annotation.
func NewPdfAnnotationSquiggly() *PdfAnnotationSquiggly {
	annotation := NewPdfAnnotation()
	squigglyAnnotation := &PdfAnnotationSquiggly{}
	squigglyAnnotation.PdfAnnotation = annotation
	squigglyAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(squigglyAnnotation)
	return squigglyAnnotation
}

// Create a new text strikeout annotation.
func NewPdfAnnotationStrikeOut() *PdfAnnotationStrikeOut {
	annotation := NewPdfAnnotation()
	strikeoutAnnotation := &PdfAnnotationStrikeOut{}
	strikeoutAnnotation.PdfAnnotation = annotation
	strikeoutAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(strikeoutAnnotation)
	return strikeoutAnnotation
}

// Create a new caret annotation.
func NewPdfAnnotationCaret() *PdfAnnotationCaret {
	annotation := NewPdfAnnotation()
	caretAnnotation := &PdfAnnotationCaret{}
	caretAnnotation.PdfAnnotation = annotation
	caretAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(caretAnnotation)
	return caretAnnotation
}

// Create a new stamp annotation.
func NewPdfAnnotationStamp() *PdfAnnotationStamp {
	annotation := NewPdfAnnotation()
	stampAnnotation := &PdfAnnotationStamp{}
	stampAnnotation.PdfAnnotation = annotation
	stampAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(stampAnnotation)
	return stampAnnotation
}

// Create a new ink annotation.
func NewPdfAnnotationInk() *PdfAnnotationInk {
	annotation := NewPdfAnnotation()
	inkAnnotation := &PdfAnnotationInk{}
	inkAnnotation.PdfAnnotation = annotation
	inkAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(inkAnnotation)
	return inkAnnotation
}

// Create a new popup annotation.
func NewPdfAnnotationPopup() *PdfAnnotationPopup {
	annotation := NewPdfAnnotation()
	popupAnnotation := &PdfAnnotationPopup{}
	popupAnnotation.PdfAnnotation = annotation
	annotation.SetContext(popupAnnotation)
	return popupAnnotation
}

// Create a new file attachment annotation.
func NewPdfAnnotationFileAttachment() *PdfAnnotationFileAttachment {
	annotation := NewPdfAnnotation()
	fileAnnotation := &PdfAnnotationFileAttachment{}
	fileAnnotation.PdfAnnotation = annotation
	fileAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(fileAnnotation)
	return fileAnnotation
}

// Create a new sound annotation.
func NewPdfAnnotationSound() *PdfAnnotationSound {
	annotation := NewPdfAnnotation()
	soundAnnotation := &PdfAnnotationSound{}
	soundAnnotation.PdfAnnotation = annotation
	soundAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(soundAnnotation)
	return soundAnnotation
}

// Create a new rich media annotation.
func NewPdfAnnotationRichMedia() *PdfAnnotationRichMedia {
	annotation := NewPdfAnnotation()
	richmediaAnnotation := &PdfAnnotationRichMedia{}
	richmediaAnnotation.PdfAnnotation = annotation
	annotation.SetContext(richmediaAnnotation)
	return richmediaAnnotation
}

// Create a new movie annotation.
func NewPdfAnnotationMovie() *PdfAnnotationMovie {
	annotation := NewPdfAnnotation()
	movieAnnotation := &PdfAnnotationMovie{}
	movieAnnotation.PdfAnnotation = annotation
	annotation.SetContext(movieAnnotation)
	return movieAnnotation
}

// Create a new screen annotation.
func NewPdfAnnotationScreen() *PdfAnnotationScreen {
	annotation := NewPdfAnnotation()
	screenAnnotation := &PdfAnnotationScreen{}
	screenAnnotation.PdfAnnotation = annotation
	annotation.SetContext(screenAnnotation)
	return screenAnnotation
}

// Create a new watermark annotation.
func NewPdfAnnotationWatermark() *PdfAnnotationWatermark {
	annotation := NewPdfAnnotation()
	watermarkAnnotation := &PdfAnnotationWatermark{}
	watermarkAnnotation.PdfAnnotation = annotation
	annotation.SetContext(watermarkAnnotation)
	return watermarkAnnotation
}

// Create a new printermark annotation.
func NewPdfAnnotationPrinterMark() *PdfAnnotationPrinterMark {
	annotation := NewPdfAnnotation()
	printermarkAnnotation := &PdfAnnotationPrinterMark{}
	printermarkAnnotation.PdfAnnotation = annotation
	annotation.SetContext(printermarkAnnotation)
	return printermarkAnnotation
}

// Create a new trapnet annotation.
func NewPdfAnnotationTrapNet() *PdfAnnotationTrapNet {
	annotation := NewPdfAnnotation()
	trapnetAnnotation := &PdfAnnotationTrapNet{}
	trapnetAnnotation.PdfAnnotation = annotation
	annotation.SetContext(trapnetAnnotation)
	return trapnetAnnotation
}

// Create a new 3d annotation.
func NewPdfAnnotation3D() *PdfAnnotation3D {
	annotation := NewPdfAnnotation()
	x3dAnnotation := &PdfAnnotation3D{}
	x3dAnnotation.PdfAnnotation = annotation
	annotation.SetContext(x3dAnnotation)
	return x3dAnnotation
}

// Create a new projection annotation.
func NewPdfAnnotationProjection() *PdfAnnotationProjection {
	annotation := NewPdfAnnotation()
	projectionAnnotation := &PdfAnnotationProjection{}
	projectionAnnotation.PdfAnnotation = annotation
	projectionAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(projectionAnnotation)
	return projectionAnnotation
}

// Create a new redact annotation.
func NewPdfAnnotationRedact() *PdfAnnotationRedact {
	annotation := NewPdfAnnotation()
	redactAnnotation := &PdfAnnotationRedact{}
	redactAnnotation.PdfAnnotation = annotation
	redactAnnotation.PdfAnnotationMarkup = &PdfAnnotationMarkup{}
	annotation.SetContext(redactAnnotation)
	return redactAnnotation
}

// Used for PDF parsing.  Loads a PDF annotation model from a PDF primitive dictionary object.
// Loads the common PDF annotation dictionary, and anything needed for the annotation subtype.
func (r *PdfReader) newPdfAnnotationFromIndirectObject(container *PdfIndirectObject) (*PdfAnnotation, error) {
	d, isDict := container.PdfObject.(*PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("Annotation indirect object not containing a dictionary")
	}

	// Check if cached, return cached model if exists.
	if model := r.modelManager.GetModelFromPrimitive(d); model != nil {
		annot, ok := model.(*PdfAnnotation)
		if !ok {
			return nil, fmt.Errorf("Cached model not a PDF annotation")
		}
		return annot, nil
	}

	annot := &PdfAnnotation{}
	annot.primitive = container
	r.modelManager.Register(d, annot)

	if obj := d.Get("Type"); obj != nil {
		str, ok := obj.(*PdfObjectName)
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
	subtype, ok := subtypeObj.(*PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR: Invalid Subtype object type != name (%T)", subtypeObj)
		return nil, fmt.Errorf("Invalid Subtype object type != name (%T)", subtypeObj)
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

	err := fmt.Errorf("Unknown annotation (%s)", *subtype)
	return nil, err
}

// Load data for markup annotation subtypes.
func (r *PdfReader) newPdfAnnotationMarkupFromDict(d *PdfObjectDictionary) (*PdfAnnotationMarkup, error) {
	annot := &PdfAnnotationMarkup{}

	if obj := d.Get("T"); obj != nil {
		annot.T = obj
	}

	if obj := d.Get("Popup"); obj != nil {
		indObj, isIndirect := obj.(*PdfIndirectObject)
		if !isIndirect {
			if _, isNull := obj.(*PdfObjectNull); !isNull {
				return nil, fmt.Errorf("Popup should point to an indirect object")
			}
		} else {
			popupAnnotObj, err := r.newPdfAnnotationFromIndirectObject(indObj)
			if err != nil {
				return nil, err
			}
			popupAnnot, isPopupAnnot := popupAnnotObj.context.(*PdfAnnotationPopup)
			if !isPopupAnnot {
				return nil, fmt.Errorf("Popup not referring to a popup annotation!")
			}

			annot.Popup = popupAnnot
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

func (r *PdfReader) newPdfAnnotationTextFromDict(d *PdfObjectDictionary) (*PdfAnnotationText, error) {
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

func (r *PdfReader) newPdfAnnotationLinkFromDict(d *PdfObjectDictionary) (*PdfAnnotationLink, error) {
	annot := PdfAnnotationLink{}

	annot.A = d.Get("A")
	annot.Dest = d.Get("Dest")
	annot.H = d.Get("H")
	annot.PA = d.Get("PA")
	annot.QuadPoints = d.Get("QuadPoints")
	annot.BS = d.Get("BS")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationFreeTextFromDict(d *PdfObjectDictionary) (*PdfAnnotationFreeText, error) {
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

func (r *PdfReader) newPdfAnnotationLineFromDict(d *PdfObjectDictionary) (*PdfAnnotationLine, error) {
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

func (r *PdfReader) newPdfAnnotationSquareFromDict(d *PdfObjectDictionary) (*PdfAnnotationSquare, error) {
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

func (r *PdfReader) newPdfAnnotationCircleFromDict(d *PdfObjectDictionary) (*PdfAnnotationCircle, error) {
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

func (r *PdfReader) newPdfAnnotationPolygonFromDict(d *PdfObjectDictionary) (*PdfAnnotationPolygon, error) {
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

func (r *PdfReader) newPdfAnnotationPolyLineFromDict(d *PdfObjectDictionary) (*PdfAnnotationPolyLine, error) {
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

func (r *PdfReader) newPdfAnnotationHighlightFromDict(d *PdfObjectDictionary) (*PdfAnnotationHighlight, error) {
	annot := PdfAnnotationHighlight{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationUnderlineFromDict(d *PdfObjectDictionary) (*PdfAnnotationUnderline, error) {
	annot := PdfAnnotationUnderline{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationSquigglyFromDict(d *PdfObjectDictionary) (*PdfAnnotationSquiggly, error) {
	annot := PdfAnnotationSquiggly{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationStrikeOut(d *PdfObjectDictionary) (*PdfAnnotationStrikeOut, error) {
	annot := PdfAnnotationStrikeOut{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.QuadPoints = d.Get("QuadPoints")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationCaretFromDict(d *PdfObjectDictionary) (*PdfAnnotationCaret, error) {
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
func (r *PdfReader) newPdfAnnotationStampFromDict(d *PdfObjectDictionary) (*PdfAnnotationStamp, error) {
	annot := PdfAnnotationStamp{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	annot.Name = d.Get("Name")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationInkFromDict(d *PdfObjectDictionary) (*PdfAnnotationInk, error) {
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

func (r *PdfReader) newPdfAnnotationPopupFromDict(d *PdfObjectDictionary) (*PdfAnnotationPopup, error) {
	annot := PdfAnnotationPopup{}

	annot.Parent = d.Get("Parent")
	annot.Open = d.Get("Open")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationFileAttachmentFromDict(d *PdfObjectDictionary) (*PdfAnnotationFileAttachment, error) {
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

func (r *PdfReader) newPdfAnnotationSoundFromDict(d *PdfObjectDictionary) (*PdfAnnotationSound, error) {
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

func (r *PdfReader) newPdfAnnotationRichMediaFromDict(d *PdfObjectDictionary) (*PdfAnnotationRichMedia, error) {
	annot := &PdfAnnotationRichMedia{}

	annot.RichMediaSettings = d.Get("RichMediaSettings")
	annot.RichMediaContent = d.Get("RichMediaContent")

	return annot, nil
}

func (r *PdfReader) newPdfAnnotationMovieFromDict(d *PdfObjectDictionary) (*PdfAnnotationMovie, error) {
	annot := PdfAnnotationMovie{}

	annot.T = d.Get("T")
	annot.Movie = d.Get("Movie")
	annot.A = d.Get("A")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationScreenFromDict(d *PdfObjectDictionary) (*PdfAnnotationScreen, error) {
	annot := PdfAnnotationScreen{}

	annot.T = d.Get("T")
	annot.MK = d.Get("MK")
	annot.A = d.Get("A")
	annot.AA = d.Get("AA")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationWidgetFromDict(d *PdfObjectDictionary) (*PdfAnnotationWidget, error) {
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

func (r *PdfReader) newPdfAnnotationPrinterMarkFromDict(d *PdfObjectDictionary) (*PdfAnnotationPrinterMark, error) {
	annot := PdfAnnotationPrinterMark{}

	annot.MN = d.Get("MN")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationTrapNetFromDict(d *PdfObjectDictionary) (*PdfAnnotationTrapNet, error) {
	annot := PdfAnnotationTrapNet{}
	// empty?e
	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationWatermarkFromDict(d *PdfObjectDictionary) (*PdfAnnotationWatermark, error) {
	annot := PdfAnnotationWatermark{}

	annot.FixedPrint = d.Get("FixedPrint")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotation3DFromDict(d *PdfObjectDictionary) (*PdfAnnotation3D, error) {
	annot := PdfAnnotation3D{}

	annot.T3DD = d.Get("3DD")
	annot.T3DV = d.Get("3DV")
	annot.T3DA = d.Get("3DA")
	annot.T3DI = d.Get("3DI")
	annot.T3DB = d.Get("3DB")

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationProjectionFromDict(d *PdfObjectDictionary) (*PdfAnnotationProjection, error) {
	annot := &PdfAnnotationProjection{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	return annot, nil
}

func (r *PdfReader) newPdfAnnotationRedactFromDict(d *PdfObjectDictionary) (*PdfAnnotationRedact, error) {
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

func (this *PdfAnnotation) GetContainingPdfObject() PdfObject {
	return this.primitive
}

// Note: Call the sub-annotation's ToPdfObject to set both the generic and non-generic information.
// TODO/FIXME: Consider doing it here instead.
func (this *PdfAnnotation) ToPdfObject() PdfObject {
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.Set("Type", MakeName("Annot"))
	d.SetIfNotNil("Rect", this.Rect)
	d.SetIfNotNil("Contents", this.Contents)
	d.SetIfNotNil("P", this.P)
	d.SetIfNotNil("NM", this.NM)
	d.SetIfNotNil("M", this.M)
	d.SetIfNotNil("F", this.F)
	d.SetIfNotNil("AP", this.AP)
	d.SetIfNotNil("AS", this.AS)
	d.SetIfNotNil("Border", this.Border)
	d.SetIfNotNil("C", this.C)
	d.SetIfNotNil("StructParent", this.StructParent)
	d.SetIfNotNil("OC", this.OC)

	return container
}

// Markup portion of the annotation.
func (this *PdfAnnotationMarkup) appendToPdfDictionary(d *PdfObjectDictionary) {
	d.SetIfNotNil("T", this.T)
	if this.Popup != nil {
		d.Set("Popup", this.Popup.ToPdfObject())
	}
	d.SetIfNotNil("CA", this.CA)
	d.SetIfNotNil("RC", this.RC)
	d.SetIfNotNil("CreationDate", this.CreationDate)
	d.SetIfNotNil("IRT", this.IRT)
	d.SetIfNotNil("Subj", this.Subj)
	d.SetIfNotNil("RT", this.RT)
	d.SetIfNotNil("IT", this.IT)
	d.SetIfNotNil("ExData", this.ExData)
}

func (this *PdfAnnotationText) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	if this.PdfAnnotationMarkup != nil {
		this.PdfAnnotationMarkup.appendToPdfDictionary(d)
	}

	d.SetIfNotNil("Subtype", MakeName("Text"))
	d.SetIfNotNil("Open", this.Open)
	d.SetIfNotNil("Name", this.Name)
	d.SetIfNotNil("State", this.State)
	d.SetIfNotNil("StateModel", this.StateModel)
	return container
}

func (this *PdfAnnotationLink) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("Link"))
	d.SetIfNotNil("A", this.A)
	d.SetIfNotNil("Dest", this.Dest)
	d.SetIfNotNil("H", this.H)
	d.SetIfNotNil("PA", this.PA)
	d.SetIfNotNil("QuadPoints", this.QuadPoints)
	d.SetIfNotNil("BS", this.BS)
	return container
}

func (this *PdfAnnotationFreeText) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("FreeText"))
	d.SetIfNotNil("DA", this.DA)
	d.SetIfNotNil("Q", this.Q)
	d.SetIfNotNil("RC", this.RC)
	d.SetIfNotNil("DS", this.DS)
	d.SetIfNotNil("CL", this.CL)
	d.SetIfNotNil("IT", this.IT)
	d.SetIfNotNil("BE", this.BE)
	d.SetIfNotNil("RD", this.RD)
	d.SetIfNotNil("BS", this.BS)
	d.SetIfNotNil("LE", this.LE)

	return container
}
func (this *PdfAnnotationLine) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Line"))
	d.SetIfNotNil("L", this.L)
	d.SetIfNotNil("BS", this.BS)
	d.SetIfNotNil("LE", this.LE)
	d.SetIfNotNil("IC", this.IC)
	d.SetIfNotNil("LL", this.LL)
	d.SetIfNotNil("LLE", this.LLE)
	d.SetIfNotNil("Cap", this.Cap)
	d.SetIfNotNil("IT", this.IT)
	d.SetIfNotNil("LLO", this.LLO)
	d.SetIfNotNil("CP", this.CP)
	d.SetIfNotNil("Measure", this.Measure)
	d.SetIfNotNil("CO", this.CO)

	return container
}

func (this *PdfAnnotationSquare) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	if this.PdfAnnotationMarkup != nil {
		this.PdfAnnotationMarkup.appendToPdfDictionary(d)
	}

	d.SetIfNotNil("Subtype", MakeName("Square"))
	d.SetIfNotNil("BS", this.BS)
	d.SetIfNotNil("IC", this.IC)
	d.SetIfNotNil("BE", this.BE)
	d.SetIfNotNil("RD", this.RD)

	return container
}

func (this *PdfAnnotationCircle) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Circle"))
	d.SetIfNotNil("BS", this.BS)
	d.SetIfNotNil("IC", this.IC)
	d.SetIfNotNil("BE", this.BE)
	d.SetIfNotNil("RD", this.RD)

	return container
}

func (this *PdfAnnotationPolygon) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Polygon"))
	d.SetIfNotNil("Vertices", this.Vertices)
	d.SetIfNotNil("LE", this.LE)
	d.SetIfNotNil("BS", this.BS)
	d.SetIfNotNil("IC", this.IC)
	d.SetIfNotNil("BE", this.BE)
	d.SetIfNotNil("IT", this.IT)
	d.SetIfNotNil("Measure", this.Measure)

	return container
}

func (this *PdfAnnotationPolyLine) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("PolyLine"))
	d.SetIfNotNil("Vertices", this.Vertices)
	d.SetIfNotNil("LE", this.LE)
	d.SetIfNotNil("BS", this.BS)
	d.SetIfNotNil("IC", this.IC)
	d.SetIfNotNil("BE", this.BE)
	d.SetIfNotNil("IT", this.IT)
	d.SetIfNotNil("Measure", this.Measure)

	return container
}

func (this *PdfAnnotationHighlight) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Highlight"))
	d.SetIfNotNil("QuadPoints", this.QuadPoints)
	return container
}

func (this *PdfAnnotationUnderline) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Underline"))
	d.SetIfNotNil("QuadPoints", this.QuadPoints)
	return container
}

func (this *PdfAnnotationSquiggly) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Squiggly"))
	d.SetIfNotNil("QuadPoints", this.QuadPoints)
	return container
}

func (this *PdfAnnotationStrikeOut) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("StrikeOut"))
	d.SetIfNotNil("QuadPoints", this.QuadPoints)
	return container
}

func (this *PdfAnnotationCaret) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Caret"))
	d.SetIfNotNil("RD", this.RD)
	d.SetIfNotNil("Sy", this.Sy)
	return container
}

func (this *PdfAnnotationStamp) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Stamp"))
	d.SetIfNotNil("Name", this.Name)
	return container
}

func (this *PdfAnnotationInk) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Ink"))
	d.SetIfNotNil("InkList", this.InkList)
	d.SetIfNotNil("BS", this.BS)
	return container
}

func (this *PdfAnnotationPopup) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("Popup"))
	d.SetIfNotNil("Parent", this.Parent)
	d.SetIfNotNil("Open", this.Open)
	return container
}

func (this *PdfAnnotationFileAttachment) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("FileAttachment"))
	d.SetIfNotNil("FS", this.FS)
	d.SetIfNotNil("Name", this.Name)
	return container
}

func (this *PdfAnnotationRichMedia) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("RichMedia"))
	d.SetIfNotNil("RichMediaSettings", this.RichMediaSettings)
	d.SetIfNotNil("RichMediaContent", this.RichMediaContent)
	return container
}

func (this *PdfAnnotationSound) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Sound"))
	d.SetIfNotNil("Sound", this.Sound)
	d.SetIfNotNil("Name", this.Name)
	return container
}

func (this *PdfAnnotationMovie) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("Movie"))
	d.SetIfNotNil("T", this.T)
	d.SetIfNotNil("Movie", this.Movie)
	d.SetIfNotNil("A", this.A)
	return container
}

func (this *PdfAnnotationScreen) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("Screen"))
	d.SetIfNotNil("T", this.T)
	d.SetIfNotNil("MK", this.MK)
	d.SetIfNotNil("A", this.A)
	d.SetIfNotNil("AA", this.AA)
	return container
}

func (this *PdfAnnotationWidget) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("Widget"))
	d.SetIfNotNil("H", this.H)
	d.SetIfNotNil("MK", this.MK)
	d.SetIfNotNil("A", this.A)
	d.SetIfNotNil("AA", this.AA)
	d.SetIfNotNil("BS", this.BS)
	d.SetIfNotNil("Parent", this.Parent)

	return container
}

func (this *PdfAnnotationPrinterMark) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("PrinterMark"))
	d.SetIfNotNil("MN", this.MN)
	return container
}

func (this *PdfAnnotationTrapNet) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("TrapNet"))
	return container
}

func (this *PdfAnnotationWatermark) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Subtype", MakeName("Watermark"))
	d.SetIfNotNil("FixedPrint", this.FixedPrint)

	return container
}

func (this *PdfAnnotation3D) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	d.SetIfNotNil("Subtype", MakeName("3D"))
	d.SetIfNotNil("3DD", this.T3DD)
	d.SetIfNotNil("3DV", this.T3DV)
	d.SetIfNotNil("3DA", this.T3DA)
	d.SetIfNotNil("3DI", this.T3DI)
	d.SetIfNotNil("3DB", this.T3DB)
	return container
}

func (this *PdfAnnotationProjection) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)
	return container
}

func (this *PdfAnnotationRedact) ToPdfObject() PdfObject {
	this.PdfAnnotation.ToPdfObject()
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

	d.SetIfNotNil("Subtype", MakeName("Redact"))
	d.SetIfNotNil("QuadPoints", this.QuadPoints)
	d.SetIfNotNil("IC", this.IC)
	d.SetIfNotNil("RO", this.RO)
	d.SetIfNotNil("OverlayText", this.OverlayText)
	d.SetIfNotNil("Repeat", this.Repeat)
	d.SetIfNotNil("DA", this.DA)
	d.SetIfNotNil("Q", this.Q)
	return container
}

// Border definitions.

type BorderStyle int

const (
	BorderStyleSolid     BorderStyle = iota
	BorderStyleDashed    BorderStyle = iota
	BorderStyleBeveled   BorderStyle = iota
	BorderStyleInset     BorderStyle = iota
	BorderStyleUnderline BorderStyle = iota
)

func (this *BorderStyle) GetPdfName() string {
	switch *this {
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

// Border style
type PdfBorderStyle struct {
	W *float64     // Border width
	S *BorderStyle // Border style
	D *[]int       // Dash array.

	container PdfObject
}

func NewBorderStyle() *PdfBorderStyle {
	bs := &PdfBorderStyle{}
	return bs
}

func (this *PdfBorderStyle) SetBorderWidth(width float64) {
	this.W = &width
}

func (this *PdfBorderStyle) GetBorderWidth() float64 {
	if this.W == nil {
		return 1 // Default.
	}
	return *this.W
}

func newPdfBorderStyleFromPdfObject(obj PdfObject) (*PdfBorderStyle, error) {
	bs := &PdfBorderStyle{}
	bs.container = obj

	var d *PdfObjectDictionary
	obj = TraceToDirectObject(obj)
	d, ok := obj.(*PdfObjectDictionary)
	if !ok {
		return nil, errors.New("Type check")
	}

	// Type.
	if obj := d.Get("Type"); obj != nil {
		name, ok := obj.(*PdfObjectName)
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
		val, err := getNumberAsFloat(obj)
		if err != nil {
			common.Log.Debug("Error retrieving W: %v", err)
			return nil, err
		}
		bs.W = &val
	}

	// Border style.
	if obj := d.Get("S"); obj != nil {
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, errors.New("Border S not a name object")
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
			return nil, errors.New("Style type range check")
		}

		bs.S = &style
	}

	// Dash array.
	if obj := d.Get("D"); obj != nil {
		vec, ok := obj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Border D dash not an array: %T", obj)
			return nil, errors.New("Border D type check error")
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

func (this *PdfBorderStyle) ToPdfObject() PdfObject {
	d := MakeDict()
	if this.container != nil {
		if indObj, is := this.container.(*PdfIndirectObject); is {
			indObj.PdfObject = d
		}
	}

	d.Set("Subtype", MakeName("Border"))
	if this.W != nil {
		d.Set("W", MakeFloat(*this.W))
	}
	if this.S != nil {
		d.Set("S", MakeName(this.S.GetPdfName()))
	}
	if this.D != nil {
		d.Set("D", MakeArrayFromIntegers(*this.D))
	}

	if this.container != nil {
		return this.container
	} else {
		return d
	}
}

// Border effect
type BorderEffect int

const (
	BorderEffectNoEffect BorderEffect = iota
	BorderEffectCloudy   BorderEffect = iota
)

type PdfBorderEffect struct {
	S *BorderEffect // Border effect type
	I *float64      // Intensity of the effect
}
