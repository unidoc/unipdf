/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
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
	container.PdfObject = &PdfObjectDictionary{}

	annot.primitive = container
	return annot
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

	if obj, has := (*d)["Type"]; has {
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

	if obj, has := (*d)["Rect"]; has {
		annot.Rect = obj
	}

	if obj, has := (*d)["Contents"]; has {
		annot.Contents = obj
	}

	if obj, has := (*d)["P"]; has {
		annot.P = obj
	}

	if obj, has := (*d)["NM"]; has {
		annot.NM = obj
	}

	if obj, has := (*d)["M"]; has {
		annot.M = obj
	}

	if obj, has := (*d)["F"]; has {
		annot.F = obj
	}

	if obj, has := (*d)["AP"]; has {
		annot.AP = obj
	}

	if obj, has := (*d)["AS"]; has {
		annot.AS = obj
	}

	if obj, has := (*d)["Border"]; has {
		annot.Border = obj
	}

	if obj, has := (*d)["C"]; has {
		annot.C = obj
	}

	if obj, has := (*d)["StructParent"]; has {
		annot.StructParent = obj
	}

	if obj, has := (*d)["OC"]; has {
		annot.OC = obj
	}

	subtypeObj, has := (*d)["Subtype"]
	if !has {
		return nil, fmt.Errorf("Missing Subtype")
	}
	subtype, ok := subtypeObj.(*PdfObjectName)
	if !ok {
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

	if obj, has := (*d)["T"]; has {
		annot.T = obj
	}

	if obj, has := (*d)["Popup"]; has {
		indObj, isIndirect := obj.(*PdfIndirectObject)
		if !isIndirect {
			if _, isNull := obj.(*PdfObjectNull); !isNull {
				return nil, fmt.Errorf("Popup should point to an indirect object")
			}
		}

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

	if obj, has := (*d)["CA"]; has {
		annot.CA = obj
	}
	if obj, has := (*d)["RC"]; has {
		annot.RC = obj
	}
	if obj, has := (*d)["CreationDate"]; has {
		annot.CreationDate = obj
	}
	if obj, has := (*d)["IRT"]; has {
		annot.IRT = obj
	}
	if obj, has := (*d)["Subj"]; has {
		annot.Subj = obj
	}
	if obj, has := (*d)["RT"]; has {
		annot.RT = obj
	}
	if obj, has := (*d)["IT"]; has {
		annot.IT = obj
	}
	if obj, has := (*d)["ExData"]; has {
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

	if obj, has := (*d)["Open"]; has {
		annot.Open = obj
	}

	if obj, has := (*d)["Name"]; has {
		annot.Name = obj
	}

	if obj, has := (*d)["State"]; has {
		annot.State = obj
	}

	if obj, has := (*d)["StateModel"]; has {
		annot.StateModel = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationLinkFromDict(d *PdfObjectDictionary) (*PdfAnnotationLink, error) {
	annot := PdfAnnotationLink{}

	if obj, has := (*d)["A"]; has {
		annot.A = obj
	}
	if obj, has := (*d)["Dest"]; has {
		annot.Dest = obj
	}
	if obj, has := (*d)["H"]; has {
		annot.H = obj
	}
	if obj, has := (*d)["PA"]; has {
		annot.PA = obj
	}
	if obj, has := (*d)["QuadPoints"]; has {
		annot.QuadPoints = obj
	}
	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationFreeTextFromDict(d *PdfObjectDictionary) (*PdfAnnotationFreeText, error) {
	annot := PdfAnnotationFreeText{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["DA"]; has {
		annot.DA = obj
	}
	if obj, has := (*d)["Q"]; has {
		annot.Q = obj
	}
	if obj, has := (*d)["RC"]; has {
		annot.RC = obj
	}
	if obj, has := (*d)["DS"]; has {
		annot.DS = obj
	}
	if obj, has := (*d)["CL"]; has {
		annot.CL = obj
	}
	if obj, has := (*d)["IT"]; has {
		annot.IT = obj
	}
	if obj, has := (*d)["BE"]; has {
		annot.BE = obj
	}
	if obj, has := (*d)["RD"]; has {
		annot.RD = obj
	}
	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}
	if obj, has := (*d)["LE"]; has {
		annot.LE = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationLineFromDict(d *PdfObjectDictionary) (*PdfAnnotationLine, error) {
	annot := PdfAnnotationLine{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["L"]; has {
		annot.L = obj
	}
	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}
	if obj, has := (*d)["LE"]; has {
		annot.LE = obj
	}
	if obj, has := (*d)["IC"]; has {
		annot.IC = obj
	}
	if obj, has := (*d)["LL"]; has {
		annot.LL = obj
	}
	if obj, has := (*d)["LLE"]; has {
		annot.LLE = obj
	}
	if obj, has := (*d)["Cap"]; has {
		annot.Cap = obj
	}
	if obj, has := (*d)["IT"]; has {
		annot.IT = obj
	}
	if obj, has := (*d)["LLO"]; has {
		annot.LLO = obj
	}
	if obj, has := (*d)["CP"]; has {
		annot.CP = obj
	}
	if obj, has := (*d)["Measure"]; has {
		annot.Measure = obj
	}
	if obj, has := (*d)["CO"]; has {
		annot.CO = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationSquareFromDict(d *PdfObjectDictionary) (*PdfAnnotationSquare, error) {
	annot := PdfAnnotationSquare{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}
	if obj, has := (*d)["IC"]; has {
		annot.IC = obj
	}
	if obj, has := (*d)["BE"]; has {
		annot.BE = obj
	}
	if obj, has := (*d)["RD"]; has {
		annot.RD = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationCircleFromDict(d *PdfObjectDictionary) (*PdfAnnotationCircle, error) {
	annot := PdfAnnotationCircle{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}
	if obj, has := (*d)["IC"]; has {
		annot.IC = obj
	}
	if obj, has := (*d)["BE"]; has {
		annot.BE = obj
	}
	if obj, has := (*d)["RD"]; has {
		annot.RD = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPolygonFromDict(d *PdfObjectDictionary) (*PdfAnnotationPolygon, error) {
	annot := PdfAnnotationPolygon{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["Vertices"]; has {
		annot.Vertices = obj
	}
	if obj, has := (*d)["LE"]; has {
		annot.LE = obj
	}
	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}
	if obj, has := (*d)["IC"]; has {
		annot.IC = obj
	}
	if obj, has := (*d)["BE"]; has {
		annot.BE = obj
	}
	if obj, has := (*d)["IT"]; has {
		annot.IT = obj
	}
	if obj, has := (*d)["Measure"]; has {
		annot.Measure = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPolyLineFromDict(d *PdfObjectDictionary) (*PdfAnnotationPolyLine, error) {
	annot := PdfAnnotationPolyLine{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["Vertices"]; has {
		annot.Vertices = obj
	}
	if obj, has := (*d)["LE"]; has {
		annot.LE = obj
	}
	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}
	if obj, has := (*d)["IC"]; has {
		annot.IC = obj
	}
	if obj, has := (*d)["BE"]; has {
		annot.BE = obj
	}
	if obj, has := (*d)["IT"]; has {
		annot.IT = obj
	}
	if obj, has := (*d)["Measure"]; has {
		annot.Measure = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationHighlightFromDict(d *PdfObjectDictionary) (*PdfAnnotationHighlight, error) {
	annot := PdfAnnotationHighlight{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationUnderlineFromDict(d *PdfObjectDictionary) (*PdfAnnotationUnderline, error) {
	annot := PdfAnnotationUnderline{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationSquigglyFromDict(d *PdfObjectDictionary) (*PdfAnnotationSquiggly, error) {
	annot := PdfAnnotationSquiggly{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationStrikeOut(d *PdfObjectDictionary) (*PdfAnnotationStrikeOut, error) {
	annot := PdfAnnotationStrikeOut{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationCaretFromDict(d *PdfObjectDictionary) (*PdfAnnotationCaret, error) {
	annot := PdfAnnotationCaret{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["RD"]; has {
		annot.RD = obj
	}

	if obj, has := (*d)["Sy"]; has {
		annot.Sy = obj
	}

	return &annot, nil
}
func (r *PdfReader) newPdfAnnotationStampFromDict(d *PdfObjectDictionary) (*PdfAnnotationStamp, error) {
	annot := PdfAnnotationStamp{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["Name"]; has {
		annot.Name = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationInkFromDict(d *PdfObjectDictionary) (*PdfAnnotationInk, error) {
	annot := PdfAnnotationInk{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["InkList"]; has {
		annot.InkList = obj
	}

	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPopupFromDict(d *PdfObjectDictionary) (*PdfAnnotationPopup, error) {
	annot := PdfAnnotationPopup{}

	if obj, has := (*d)["Parent"]; has {
		annot.Parent = obj
	}

	if obj, has := (*d)["Open"]; has {
		annot.Open = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationFileAttachmentFromDict(d *PdfObjectDictionary) (*PdfAnnotationFileAttachment, error) {
	annot := PdfAnnotationFileAttachment{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["FS"]; has {
		annot.FS = obj
	}

	if obj, has := (*d)["Name"]; has {
		annot.Name = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationSoundFromDict(d *PdfObjectDictionary) (*PdfAnnotationSound, error) {
	annot := PdfAnnotationSound{}

	markup, err := r.newPdfAnnotationMarkupFromDict(d)
	if err != nil {
		return nil, err
	}
	annot.PdfAnnotationMarkup = markup

	if obj, has := (*d)["Name"]; has {
		annot.Name = obj
	}

	if obj, has := (*d)["Sound"]; has {
		annot.Sound = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationRichMediaFromDict(d *PdfObjectDictionary) (*PdfAnnotationRichMedia, error) {
	annot := &PdfAnnotationRichMedia{}

	if obj, has := (*d)["RichMediaSettings"]; has {
		annot.RichMediaSettings = obj
	}

	if obj, has := (*d)["RichMediaContent"]; has {
		annot.RichMediaContent = obj
	}

	return annot, nil
}

func (r *PdfReader) newPdfAnnotationMovieFromDict(d *PdfObjectDictionary) (*PdfAnnotationMovie, error) {
	annot := PdfAnnotationMovie{}

	if obj, has := (*d)["T"]; has {
		annot.T = obj
	}

	if obj, has := (*d)["Movie"]; has {
		annot.Movie = obj
	}

	if obj, has := (*d)["A"]; has {
		annot.A = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationScreenFromDict(d *PdfObjectDictionary) (*PdfAnnotationScreen, error) {
	annot := PdfAnnotationScreen{}

	if obj, has := (*d)["T"]; has {
		annot.T = obj
	}

	if obj, has := (*d)["MK"]; has {
		annot.MK = obj
	}

	if obj, has := (*d)["A"]; has {
		annot.A = obj
	}

	if obj, has := (*d)["AA"]; has {
		annot.AA = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationWidgetFromDict(d *PdfObjectDictionary) (*PdfAnnotationWidget, error) {
	annot := PdfAnnotationWidget{}

	if obj, has := (*d)["H"]; has {
		annot.H = obj
	}

	if obj, has := (*d)["MK"]; has {
		annot.MK = obj
		// MK can be an indirect object...
		// Expected to be a dictionary.
	}

	if obj, has := (*d)["A"]; has {
		annot.A = obj
	}

	if obj, has := (*d)["AA"]; has {
		annot.AA = obj
	}

	if obj, has := (*d)["BS"]; has {
		annot.BS = obj
	}

	if obj, has := (*d)["Parent"]; has {
		annot.Parent = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationPrinterMarkFromDict(d *PdfObjectDictionary) (*PdfAnnotationPrinterMark, error) {
	annot := PdfAnnotationPrinterMark{}

	if obj, has := (*d)["MN"]; has {
		annot.MN = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationTrapNetFromDict(d *PdfObjectDictionary) (*PdfAnnotationTrapNet, error) {
	annot := PdfAnnotationTrapNet{}
	// empty?e
	return &annot, nil
}

func (r *PdfReader) newPdfAnnotationWatermarkFromDict(d *PdfObjectDictionary) (*PdfAnnotationWatermark, error) {
	annot := PdfAnnotationWatermark{}

	if obj, has := (*d)["FixedPrint"]; has {
		annot.FixedPrint = obj
	}

	return &annot, nil
}

func (r *PdfReader) newPdfAnnotation3DFromDict(d *PdfObjectDictionary) (*PdfAnnotation3D, error) {
	annot := PdfAnnotation3D{}

	if obj, has := (*d)["3DD"]; has {
		annot.T3DD = obj
	}
	if obj, has := (*d)["3DV"]; has {
		annot.T3DV = obj
	}
	if obj, has := (*d)["3DA"]; has {
		annot.T3DA = obj
	}
	if obj, has := (*d)["3DI"]; has {
		annot.T3DI = obj
	}
	if obj, has := (*d)["3DB"]; has {
		annot.T3DB = obj
	}

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

	if obj, has := (*d)["QuadPoints"]; has {
		annot.QuadPoints = obj
	}
	if obj, has := (*d)["IC"]; has {
		annot.IC = obj
	}
	if obj, has := (*d)["RO"]; has {
		annot.RO = obj
	}
	if obj, has := (*d)["OverlayText"]; has {
		annot.OverlayText = obj
	}
	if obj, has := (*d)["Repeat"]; has {
		annot.Repeat = obj
	}
	if obj, has := (*d)["DA"]; has {
		annot.DA = obj
	}
	if obj, has := (*d)["Q"]; has {
		annot.Q = obj
	}

	return &annot, nil
}

func (this *PdfAnnotation) GetContainingPdfObject() PdfObject {
	return this.primitive
}

func (this *PdfAnnotation) ToPdfObject() PdfObject {
	container := this.primitive
	d := container.PdfObject.(*PdfObjectDictionary)

	d.SetIfNotNil("Type", MakeName("Annot"))
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
	this.PdfAnnotationMarkup.appendToPdfDictionary(d)

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
