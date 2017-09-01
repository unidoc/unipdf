/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// PdfFont represents an underlying font structure which can be of type:
// - Type0
// - Type1
// - TrueType
// etc.
type PdfFont struct {
	context interface{} // The underlying font: Type0, Type1, Truetype, etc..
}

// Encoder returns the font's text encoder.
func (font PdfFont) Encoder() textencoding.TextEncoder {
	switch t := font.context.(type) {
	case *pdfFontTrueType:
		return t.Encoder()
	case *pdfFontType0:
		return t.Encoder()
	case *pdfCIDFontType2:
		return t.Encoder()
	}

	return nil
}

// SetEncoder sets the encoding for the underlying font.
func (font PdfFont) SetEncoder(encoder textencoding.TextEncoder) {
	switch t := font.context.(type) {
	case *pdfFontTrueType:
		t.SetEncoder(encoder)
	}
}

// GetGlyphCharMetrics returns the specified char metrics for a specified glyph name.
func (font PdfFont) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	switch t := font.context.(type) {
	case *pdfFontTrueType:
		return t.GetGlyphCharMetrics(glyph)
	case *pdfFontType0:
		return t.GetGlyphCharMetrics(glyph)
	case *pdfCIDFontType2:
		return t.GetGlyphCharMetrics(glyph)
	}
	common.Log.Debug("GetGlyphCharMetrics unsupported font type %T", font.context)

	return fonts.CharMetrics{}, false
}

// newPdfFontFromPdfObject loads a PdfFont from a dictionary.  If there is a problem an error is returned.
// The allowType0 indicates whether loading Type0 font should be supported.  Flag used to avoid
// cyclical loading.
func newPdfFontFromPdfObject(obj core.PdfObject, allowType0 bool) (*PdfFont, error) {
	font := &PdfFont{}

	dictObj := obj
	if ind, is := obj.(*core.PdfIndirectObject); is {
		dictObj = ind.PdfObject
	}

	d, ok := dictObj.(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Font not given by a dictionary (%T)", obj)
		return nil, errors.New("Type check error")
	}

	if obj := d.Get("Type"); obj != nil {
		oname, is := obj.(*core.PdfObjectName)
		if !is || string(*oname) != "Font" {
			common.Log.Debug("Incompatibility ERROR: Type (Required) defined but not Font name")
			return nil, errors.New("Range check error")
		}
	} else {
		common.Log.Debug("Incompatibility ERROR: Type (Required) missing")
		return nil, errors.New("Required attribute missing")
	}

	obj = d.Get("Subtype")
	if obj == nil {
		common.Log.Debug("Incompatibility ERROR: Subtype (Required) missing")
		return nil, errors.New("Required attribute missing")
	}

	subtype, ok := core.TraceToDirectObject(obj).(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("Incompatibility ERROR: subtype not a name (%T) ", obj)
		return nil, errors.New("Type check error")
	}

	switch subtype.String() {
	case "Type0":
		if !allowType0 {
			common.Log.Debug("Loading type0 not allowed")
			return nil, errors.New("Cyclical type0 loading error")
		}
		type0font, err := newPdfFontType0FromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error loading Type0 font: %v", err)
			return nil, err
		}
		font.context = type0font
	case "TrueType":
		truefont, err := newPdfFontTrueTypeFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error loading truetype font: %v", err)
			return nil, err
		}

		font.context = truefont
	case "CIDFontType2":
		cidfont, err := newPdfCIDFontType2FromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error loading cid font type2 font: %v", err)
			return nil, err
		}

		font.context = cidfont
	default:
		common.Log.Debug("Unsupported font type: %s", subtype.String())
		return nil, errors.New("Unsupported font type")
	}

	return font, nil
}

// ToPdfObject converts the PdfFont object to its PDF representation.
func (font PdfFont) ToPdfObject() core.PdfObject {
	switch f := font.context.(type) {
	case *pdfFontTrueType:
		return f.ToPdfObject()
	case *pdfFontType0:
		return f.ToPdfObject()
	case *pdfCIDFontType2:
		return f.ToPdfObject()
	}

	// If not supported, return null..
	common.Log.Debug("Unsupported font (%T) - returning null object", font.context)
	return core.MakeNull()
}

// PdfFontDescriptor specifies metrics and other attributes of a font and can refer to a FontFile for embedded fonts.
type PdfFontDescriptor struct {
	FontName     core.PdfObject
	FontFamily   core.PdfObject
	FontStretch  core.PdfObject
	FontWeight   core.PdfObject
	Flags        core.PdfObject
	FontBBox     core.PdfObject
	ItalicAngle  core.PdfObject
	Ascent       core.PdfObject
	Descent      core.PdfObject
	Leading      core.PdfObject
	CapHeight    core.PdfObject
	XHeight      core.PdfObject
	StemV        core.PdfObject
	StemH        core.PdfObject
	AvgWidth     core.PdfObject
	MaxWidth     core.PdfObject
	MissingWidth core.PdfObject
	FontFile     core.PdfObject
	FontFile2    core.PdfObject
	FontFile3    core.PdfObject
	CharSet      core.PdfObject

	// Additional entries for CIDFonts
	Style  core.PdfObject
	Lang   core.PdfObject
	FD     core.PdfObject
	CIDSet core.PdfObject

	// Container.
	container *core.PdfIndirectObject
}

// newPdfFontDescriptorFromPdfObject loads the font descriptor from a PdfObject.  Can either be a *PdfIndirectObject or
// a *PdfObjectDictionary.
func newPdfFontDescriptorFromPdfObject(obj core.PdfObject) (*PdfFontDescriptor, error) {
	descriptor := &PdfFontDescriptor{}

	if ind, is := obj.(*core.PdfIndirectObject); is {
		descriptor.container = ind
		obj = ind.PdfObject
	}

	d, ok := obj.(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("FontDescriptor not given by a dictionary (%T)", obj)
		return nil, errors.New("Type check error")
	}

	if obj := d.Get("Type"); obj != nil {
		oname, is := obj.(*core.PdfObjectName)
		if !is || string(*oname) != "FontDescriptor" {
			common.Log.Debug("Incompatibility: Font descriptor Type invalid (%T)", obj)
		}
	} else {
		common.Log.Debug("Incompatibility: Type (Required) missing")
	}

	if obj := d.Get("FontName"); obj != nil {
		descriptor.FontName = obj
	} else {
		common.Log.Debug("Incompatibility: FontName (Required) missing")
	}

	descriptor.FontFamily = d.Get("FontFamily")
	descriptor.FontStretch = d.Get("FontStretch")
	descriptor.FontWeight = d.Get("FontWeight")
	descriptor.Flags = d.Get("Flags")
	descriptor.FontBBox = d.Get("FontBBox")
	descriptor.ItalicAngle = d.Get("ItalicAngle")
	descriptor.Ascent = d.Get("Ascent")
	descriptor.Descent = d.Get("Descent")
	descriptor.Leading = d.Get("Leading")
	descriptor.CapHeight = d.Get("CapHeight")
	descriptor.XHeight = d.Get("XHeight")
	descriptor.StemV = d.Get("StemV")
	descriptor.StemH = d.Get("StemH")
	descriptor.AvgWidth = d.Get("AvgWidth")
	descriptor.MaxWidth = d.Get("MaxWidth")
	descriptor.MissingWidth = d.Get("MissingWidth")
	descriptor.FontFile = d.Get("FontFile")
	descriptor.FontFile2 = d.Get("FontFile2")
	descriptor.FontFile3 = d.Get("FontFile3")
	descriptor.CharSet = d.Get("CharSet")
	descriptor.Style = d.Get("Style")
	descriptor.Lang = d.Get("Lang")
	descriptor.FD = d.Get("FD")
	descriptor.CIDSet = d.Get("CIDSet")

	return descriptor, nil
}

// ToPdfObject returns the PdfFontDescriptor as a PDF dictionary inside an indirect object.
func (this *PdfFontDescriptor) ToPdfObject() core.PdfObject {
	d := core.MakeDict()
	if this.container == nil {
		this.container = &core.PdfIndirectObject{}
	}
	this.container.PdfObject = d

	d.Set("Type", core.MakeName("FontDescriptor"))

	if this.FontName != nil {
		d.Set("FontName", this.FontName)
	}

	if this.FontFamily != nil {
		d.Set("FontFamily", this.FontFamily)
	}

	if this.FontStretch != nil {
		d.Set("FontStretch", this.FontStretch)
	}

	if this.FontWeight != nil {
		d.Set("FontWeight", this.FontWeight)
	}

	if this.Flags != nil {
		d.Set("Flags", this.Flags)
	}

	if this.FontBBox != nil {
		d.Set("FontBBox", this.FontBBox)
	}

	if this.ItalicAngle != nil {
		d.Set("ItalicAngle", this.ItalicAngle)
	}

	if this.Ascent != nil {
		d.Set("Ascent", this.Ascent)
	}

	if this.Descent != nil {
		d.Set("Descent", this.Descent)
	}

	if this.Leading != nil {
		d.Set("Leading", this.Leading)
	}

	if this.CapHeight != nil {
		d.Set("CapHeight", this.CapHeight)
	}

	if this.XHeight != nil {
		d.Set("XHeight", this.XHeight)
	}

	if this.StemV != nil {
		d.Set("StemV", this.StemV)
	}

	if this.StemH != nil {
		d.Set("StemH", this.StemH)
	}

	if this.AvgWidth != nil {
		d.Set("AvgWidth", this.AvgWidth)
	}

	if this.MaxWidth != nil {
		d.Set("MaxWidth", this.MaxWidth)
	}

	if this.MissingWidth != nil {
		d.Set("MissingWidth", this.MissingWidth)
	}

	if this.FontFile != nil {
		d.Set("FontFile", this.FontFile)
	}

	if this.FontFile2 != nil {
		d.Set("FontFile2", this.FontFile2)
	}

	if this.FontFile3 != nil {
		d.Set("FontFile3", this.FontFile3)
	}

	if this.CharSet != nil {
		d.Set("CharSet", this.CharSet)
	}

	if this.Style != nil {
		d.Set("FontName", this.FontName)
	}

	if this.Lang != nil {
		d.Set("Lang", this.Lang)
	}

	if this.FD != nil {
		d.Set("FD", this.FD)
	}

	if this.CIDSet != nil {
		d.Set("CIDSet", this.CIDSet)
	}

	return this.container
}
