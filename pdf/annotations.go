/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"fmt"
	"github.com/unidoc/unidoc/common"
)

//type PdfAnnotation interface{}

type PdfAnnotation struct {
	context      interface{}
	Type         PdfObject // Annot
	Subtype      PdfObject
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
}

/*
Subtype
*/

// Subtype: Text
type PdfAnnotationText struct {
	PdfAnnotation
	Open       PdfObject
	Name       PdfObject
	State      PdfObject
	StateModel PdfObject
}

// Subtype: Link
type PdfAnnotationLink struct {
	PdfAnnotation
	A          PdfObject
	Dest       PdfObject
	H          PdfObject
	PA         PdfObject
	QuadPoints PdfObject
	BS         PdfObject
}

// Subtype: FreeText
type PdfAnnotationFreeText struct {
	PdfAnnotation
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
	PdfAnnotation
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
	PdfAnnotation
	BS PdfObject
	IC PdfObject
	BE PdfObject
	RD PdfObject
}

// Subtype: Circle
type PdfAnnotationCircle struct {
	PdfAnnotation
	BS PdfObject
	IC PdfObject
	BE PdfObject
	RD PdfObject
}

// Subtype: Polygon
type PdfAnnotationPolygon struct {
	PdfAnnotation
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
	PdfAnnotation
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
	PdfAnnotation
	QuadPoints PdfObject
}

// Subtype: Underline
type PdfAnnotationUnderline struct {
	PdfAnnotation
	QuadPoints PdfObject
}

// Subtype: Squiggly
type PdfAnnotationSquiggly struct {
	PdfAnnotation
	QuadPoints PdfObject
}

// Subtype: StrikeOut
type PdfAnnotationStrikeOut struct {
	PdfAnnotation
	QuadPoints PdfObject
}

// Subtype: Caret
type PdfAnnotationCaret struct {
	PdfAnnotation
	RD PdfObject
	Sy PdfObject
}

// Subtype: Stamp
type PdfAnnotationStamp struct {
	PdfAnnotation
	Name PdfObject
}

// Subtype: Ink
type PdfAnnotationInk struct {
	PdfAnnotation
	InkList PdfObject
	BS      PdfObject
}

// Subtype: Popup
type PdfAnnotationPopup struct {
	PdfAnnotation
	Parent PdfObject
	Open   PdfObject
}

// Subtype: FileAttachment
type PdfAnnotationFileAttachment struct {
	PdfAnnotation
	FS   PdfObject
	Name PdfObject
}

// Subtype: Sound
type PdfAnnotationSound struct {
	PdfAnnotation
	Sound PdfObject
	Name  PdfObject
}

// Subtype: Movie
type PdfAnnotationMovie struct {
	PdfAnnotation
	T     PdfObject
	Movie PdfObject
	A     PdfObject
}

// Subtype: Screen
type PdfAnnotationScreen struct {
	PdfAnnotation
	T  PdfObject
	MK PdfObject
	A  PdfObject
	AA PdfObject
}

// Subtype: Widget
type PdfAnnotationWidget struct {
	PdfAnnotation
	H      PdfObject
	MK     PdfObject
	A      PdfObject
	AA     PdfObject
	BS     PdfObject
	Parent PdfObject
}

// Subtype: Watermark
type PdfAnnotationWatermark struct {
	PdfAnnotation
	FixedPrint PdfObject
}

// Subtype: PrinterMark
type PdfAnnotationPrinterMark struct {
	PdfAnnotation
	MN PdfObject
}

// Subtype: TrapNet
type PdfAnnotationTrapNet struct {
	PdfAnnotation
}

// Subtype: 3D
type PdfAnnotation3D struct {
	PdfAnnotation
	T3DD PdfObject
	T3DV PdfObject
	T3DA PdfObject
	T3DI PdfObject
	T3DB PdfObject
}

type PdfAnnotationRedact struct {
	PdfAnnotation
	QuadPoints  PdfObject
	IC          PdfObject
	RO          PdfObject
	OverlayText PdfObject
	Repeat      PdfObject
	DA          PdfObject
	Q           PdfObject
}

func (this *PdfAnnotation) ToPdfObject() PdfObject {
	switch t := this.context.(type) {
	case *PdfAnnotationText:
		return t.ToPdfObject()
	case *PdfAnnotationLink:
		return t.ToPdfObject()
	case *PdfAnnotationFreeText:
		return t.ToPdfObject()
	case *PdfAnnotationLine:
		return t.ToPdfObject()
	case *PdfAnnotationSquare:
		return t.ToPdfObject()
	case *PdfAnnotationCircle:
		return t.ToPdfObject()
	case *PdfAnnotationPolygon:
		return t.ToPdfObject()
	case *PdfAnnotationPolyLine:
		return t.ToPdfObject()
	case *PdfAnnotationHighlight:
		return t.ToPdfObject()
	case *PdfAnnotationCaret:
		return t.ToPdfObject()
	case *PdfAnnotationStamp:
		return t.ToPdfObject()
	case *PdfAnnotationInk:
		return t.ToPdfObject()
	case *PdfAnnotationPopup:
		return t.ToPdfObject()
	case *PdfAnnotationFileAttachment:
		return t.ToPdfObject()
	case *PdfAnnotationSound:
		return t.ToPdfObject()
	case *PdfAnnotationMovie:
		return t.ToPdfObject()
	case *PdfAnnotationScreen:
		return t.ToPdfObject()
	case *PdfAnnotationWidget:
		return t.ToPdfObject(false)
	case *PdfAnnotationPrinterMark:
		return t.ToPdfObject()
	case *PdfAnnotationTrapNet:
		return t.ToPdfObject()
	case *PdfAnnotationWatermark:
		return t.ToPdfObject()
	case *PdfAnnotation3D:
		return t.ToPdfObject()
	case *PdfAnnotationRedact:
		return t.ToPdfObject()
	}
	common.Log.Error("Unknown type %T", this.context)
	return nil
}

func (r *PdfReader) newPdfAnnotationFromDict(d PdfObjectDictionary) (*PdfAnnotation, error) {
	annot := PdfAnnotation{}

	if obj, has := d["Type"]; has {
		str, ok := obj.(*PdfObjectString)
		if !ok {
			return nil, fmt.Errorf("Invalid type of Type (%T)", obj)
		}
		if *str != "Annot" {
			// Log a debug message.
			// Not returning an error on this.
			common.Log.Debug("Unsuspected Type != Annot (%s)", *str)
		}
	}

	if obj, has := d["Rect"]; has {
		annot.Rect = obj
	}

	if obj, has := d["Contents"]; has {
		annot.Contents = obj
	}

	if obj, has := d["P"]; has {
		annot.P = obj
	}

	if obj, has := d["NM"]; has {
		annot.NM = obj
	}

	if obj, has := d["M"]; has {
		annot.M = obj
	}

	if obj, has := d["F"]; has {
		annot.F = obj
	}

	if obj, has := d["AP"]; has {
		annot.AP = obj
	}

	if obj, has := d["AS"]; has {
		annot.AS = obj
	}

	if obj, has := d["Border"]; has {
		annot.Border = obj
	}

	if obj, has := d["C"]; has {
		annot.C = obj
	}

	if obj, has := d["StructParent"]; has {
		annot.StructParent = obj
	}

	if obj, has := d["OC"]; has {
		annot.OC = obj
	}

	subtypeObj, has := d["Subtype"]
	if !has {
		return nil, fmt.Errorf("Missing Subtype")
	}
	subtype, ok := subtypeObj.(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("Invalid Subtype object type != name (%T)", subtypeObj)
	}
	switch *subtype {
	case "Text":
		ctx, err := newPdfAnnotationTextFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return &annot, nil
	case "Link":
		ctx, err := newPdfAnnotationLinkFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAnnotation = annot
		annot.context = ctx
		return &annot, nil
	case "FreeText":
		ctx, err = newPdfAnnotationFreeTextFromDict(d)
	case "Line":
		ctx, err = newPdfAnnotationLineFromDict(d)
	case "Square":
		ctx, err = newPdfAnnotationSquareFromDict(d)
	case "Circle":
		ctx, err = newPdfAnnotationCircleFromDict(d)
	case "Polygon":
		ctx, err = newPdfAnnotationPolygonFromDict(d)
	case "PolyLine":
		ctx, err = newPdfAnnotationPolyLineFromDict(d)
	case "TextMarkup":
		ctx, err = newPdfAnnotationTextMarkupFromDict(d)
	case "Caret":
		ctx, err = newPdfAnnotationCaretFromDict(d)
	case "Stamp":
		ctx, err = newPdfAnnotationStampFromDict(d)
	case "Ink":
		ctx, err = newPdfAnnotationInkFromDict(d)
	case "Popup":
		ctx, err = newPdfAnnotationPopupFromDict(d)
	case "FileAttachment":
		ctx, err = newPdfAnnotationFileAttachmentFromDict(d)
	case "Sound":
		ctx, err = newPdfAnnotationSoundFromDict(d)
	case "Movie":
		ctx, err = newPdfAnnotationMovieFromDict(d)
	case "Screen":
		ctx, err = newPdfAnnotationScreenFromDict(d)
	case "Widget":
		ctx, err = newPdfAnnotationWidgetFromDict(d)
	case "PrinterMark":
		ctx, err = newPdfAnnotationPrinterMarkFromDict(d)
	case "TrapNet":
		ctx, err = newPdfAnnotationTrapNetFromDict(d)
	case "Watermark":
		ctx, err = newPdfAnnotationWatermarkFromDict(d)
	case "3D":
		ctx, err = newPdfAnnotation3DFromDict(d)
	case "Redact":
		ctx, err = newPdfAnnotationRedactFromDict(d)
	default:
		err = fmt.Errorf("Unknown annotation (%s)", *subtype)
	}
	if err != nil {
		return nil, err
	}
	annot.context = ctx
	return &annot, nil
}

func newPdfAnnotationTextFromDict(d PdfObjectDictionary) (*PdfAnnotationText, error) {
	annot := PdfAnnotationText{}

	if obj, has := d["Open"]; has {
		annot.Open = obj
	}

	if obj, has := d["Name"]; has {
		annot.Name = obj
	}

	if obj, has := d["State"]; has {
		annot.State = obj
	}

	if obj, has := d["StateModel"]; has {
		annot.StateModel = obj
	}

	return &annot, nil
}


func newPdfAnnotationLinkFromDict(d PdfObjectDictionary) (*PdfAnnotationLink, error) {
	annot := PdfAnnotationLink{}

	if obj, has := d["A"]; has {
		annot.A = obj
	}
	if obj, has := d["Dest"]; has {
		annot.Dest = obj
	}
	if obj, has := d["H"]; has {
		annot.H = obj
	}
	if obj, has := d["PA"]; has {
		annot.PA = obj
	}
	if obj, has := d["QuadPoints"]; has {
		annot.QuadPoints = obj
	}
	if obj, has := d["BS"]; has {
		annot.BS = obj
	}

	return &annot, nil
}

func newPdfAnnotationFreeTextFromDict(d PdfObjectDictionary) (*PdfAnnotationFreeText, error) {
	annot := PdfAnnotationFreeText{}

	if obj, has := d["DA"]; has {
		annot.DA = obj
	}
	if obj, has := d["Q"]; has {
		annot.Q = obj
	}
	if obj, has := d["RC"]; has {
		annot.RC = obj
	}
	if obj, has := d["DS"]; has {
		annot.DS = obj
	}
	if obj, has := d["CL"]; has {
		annot.CL = obj
	}
	if obj, has := d["IT"]; has {
		annot.IT = obj
	}
	if obj, has := d["BE"]; has {
		annot.BE = obj
	}
	if obj, has := d["RD"]; has {
		annot.RD = obj
	}
	if obj, has := d["BS"]; has {
		annot.BS = obj
	}
	if obj, has := d["LE"]; has {
		annot.LE = obj
	}

	return &annot, nil
}

func newPdfAnnotationLineFromDict(d PdfObjectDictionary) (*PdfAnnotationLine, error) {
	annot := PdfAnnotationLine{}

	if obj, has := d["L"]; has {
		annot.L = obj
	}
	if obj, has := d["BS"]; has {
		annot.BS = obj
	}
	if obj, has := d["LE"]; has {
		annot.LE = obj
	}
	if obj, has := d["IC"]; has {
		annot.IC = obj
	}
	if obj, has := d["LL"]; has {
		annot.LL = obj
	}
	if obj, has := d["LLE"]; has {
		annot.LLE = obj
	}
	if obj, has := d["Cap"]; has {
		annot.Cap = obj
	}
	if obj, has := d["IT"]; has {
		annot.IT = obj
	}
	if obj, has := d["LLO"]; has {
		annot.LLO = obj
	}
	if obj, has := d["CP"]; has {
		annot.CP = obj
	}
	if obj, has := d["Measure"]; has {
		annot.Measure = obj
	}
	if obj, has := d["CO"]; has {
		annot.CO = obj
	}

	return &annot, nil
}

func newPdfAnnotationSquareFromDict(d PdfObjectDictionary) (*PdfAnnotationSquare, error) {
	annot := PdfAnnotationSquare{}

	if obj, has := d["BS"]; has {
		annot.BS = obj
	}
	if obj, has := d["IC"]; has {
		annot.IC = obj
	}
	if obj, has := d["BE"]; has {
		annot.BE = obj
	}
	if obj, has := d["RD"]; has {
		annot.RD = obj
	}

	return &annot, nil
}

func newPdfAnnotationCircleFromDict(d PdfObjectDictionary) (*PdfAnnotationCircle, error) {
	annot := PdfAnnotationCircle{}

	if obj, has := d["BS"]; has {
		annot.BS = obj
	}
	if obj, has := d["IC"]; has {
		annot.IC = obj
	}
	if obj, has := d["BE"]; has {
		annot.BE = obj
	}
	if obj, has := d["RD"]; has {
		annot.RD = obj
	}

	return &annot, nil
}

func newPdfAnnotationPolygonFromDict(d PdfObjectDictionary) (*PdfAnnotationPolygon, error) {
	annot := PdfAnnotationPolygon{}
	if obj, has := d["Vertices"]; has {
		annot.Vertices = obj
	}
	if obj, has := d["LE"]; has {
		annot.LE = obj
	}
	if obj, has := d["BS"]; has {
		annot.BS = obj
	}
	if obj, has := d["IC"]; has {
		annot.IC = obj
	}
	if obj, has := d["BE"]; has {
		annot.BE = obj
	}
	if obj, has := d["IT"]; has {
		annot.IT = obj
	}
	if obj, has := d["Measure"]; has {
		annot.Measure = obj
	}

	return &annot, nil
}

func newPdfAnnotationPolyLineFromDict(d PdfObjectDictionary) (*PdfAnnotationPolyLine, error) {
	annot := PdfAnnotationPolyLine{}
	if obj, has := d["Vertices"]; has {
		annot.Vertices = obj
	}
	if obj, has := d["LE"]; has {
		annot.LE = obj
	}
	if obj, has := d["BS"]; has {
		annot.BS = obj
	}
	if obj, has := d["IC"]; has {
		annot.IC = obj
	}
	if obj, has := d["BE"]; has {
		annot.BE = obj
	}
	if obj, has := d["IT"]; has {
		annot.IT = obj
	}
	if obj, has := d["Measure"]; has {
		annot.Measure = obj
	}

	return &annot, nil
}

func newPdfAnnotationHighlightFromDict(d PdfObjectDictionary) (*PdfAnnotationHighlight, error) {
	annot := PdfAnnotationHighlight{}

	if obj, has := d["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}

func newPdfAnnotationUnderlineFromDict(d PdfObjectDictionary) (*PdfAnnotationUnderline, error) {
	annot := PdfAnnotationUnderline{}

	if obj, has := d["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}

func newPdfAnnotationSquigglyFromDict(d PdfObjectDictionary) (*PdfAnnotationSquiggly, error) {
	annot := PdfAnnotationSquiggly{}

	if obj, has := d["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}

func newPdfAnnotationStrikeOut(d PdfObjectDictionary) (*PdfAnnotationStrikeOut, error) {
	annot := PdfAnnotationStrikeOut{}

	if obj, has := d["QuadPoints"]; has {
		annot.QuadPoints = obj
	}

	return &annot, nil
}


func newPdfAnnotationCaretFromDict(d PdfObjectDictionary) (*PdfAnnotationCaret, error) {
	annot := PdfAnnotationCaret{}
	if obj, has := d["RD"]; has {
		annot.RD = obj
	}

	if obj, has := d["Sy"]; has {
		annot.Sy = obj
	}

	return &annot, nil
}
func newPdfAnnotationStampFromDict(d PdfObjectDictionary) (*PdfAnnotationStamp, error) {
	annot := PdfAnnotationStamp{}

	if obj, has := d["Name"]; has {
		annot.Name = obj
	}

	return &annot, nil
}

func newPdfAnnotationInkFromDict(d PdfObjectDictionary) (*PdfAnnotationInk, error) {
	annot := PdfAnnotationInk{}

	if obj, has := d["InkList"]; has {
		annot.InkList = obj
	}

	if obj, has := d["BS"]; has {
		annot.BS = obj
	}

	return &annot, nil
}

func newPdfAnnotationPopupFromDict(d PdfObjectDictionary) (*PdfAnnotationPopup, error) {
	Parent PdfObject
	Open   PdfObject
}

func newPdfAnnotationFileAttachmentFromDict(d PdfObjectDictionary) (*PdfAnnotationFileAttachment, error) {
	FS   PdfObject
	Name PdfObject
}

func newPdfAnnotationSoundFromDict(d PdfObjectDictionary) (*PdfAnnotationSound, error) {
	Sound PdfObject
	Name  PdfObject
}

func newPdfAnnotationMovieFromDict(d PdfObjectDictionary) (*PdfAnnotationMovie, error) {
	T     PdfObject
	Movie PdfObject
	A     PdfObject
}

func newPdfAnnotationScreenFromDict(d PdfObjectDictionary) (*PdfAnnotationScreen, error) {
	T  PdfObject
	MK PdfObject
	A  PdfObject
	AA PdfObject
}

func newPdfAnnotationWidgetFromDict(d PdfObjectDictionary) (*PdfAnnotationWidget, error) {
	H      PdfObject
	MK     PdfObject
	A      PdfObject
	AA     PdfObject
	BS     PdfObject
	Parent PdfObject
}

func newPdfAnnotationPrinterMarkFromDict(d PdfObjectDictionary) (*PdfAnnotationPrinterMark, error) {
		MN PdfObject
}

func newPdfAnnotationTrapNetFromDict(d PdfObjectDictionary) (*PdfAnnotationTrapNet, error) {

}

func newPdfAnnotationWatermarkFromDict(d PdfObjectDictionary) (*PdfAnnotationWatermark, error) {
FixedPrint PdfObject
}

func newPdfAnnotation3DFromDict(d PdfObjectDictionary) (*PdfAnnotation3D, error) {
	T3DD PdfObject
	T3DV PdfObject
	T3DA PdfObject
	T3DI PdfObject
	T3DB PdfObject
}

func newPdfAnnotationRedactFromDict(d PdfObjectDictionary) (*PdfAnnotationRedact, error) {
	QuadPoints  PdfObject
	IC          PdfObject
	RO          PdfObject
	OverlayText PdfObject
	Repeat      PdfObject
	DA          PdfObject
	Q           PdfObject
}


/////////


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
