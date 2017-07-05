/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"io/ioutil"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// The PdfFont structure represents an underlying font structure which can be of type:
// - Type0
// - Type1
// - TrueType
// etc.
type PdfFont struct {
	context interface{} // The underlying font: Type0, Type1, Truetype, etc..
}

// Set the encoding for the underlying font.
func (font PdfFont) SetEncoder(encoder textencoding.TextEncoder) {
	switch t := font.context.(type) {
	case *pdfFontTrueType:
		t.SetEncoder(encoder)
	}
}

func (font PdfFont) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	switch t := font.context.(type) {
	case *pdfFontTrueType:
		return t.GetGlyphCharMetrics(glyph)
	}

	return fonts.CharMetrics{}, false
}

func newPdfFontFromPdfObject(obj core.PdfObject) (*PdfFont, error) {
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

	if obj, isDefined := (*d)["Type"]; isDefined {
		oname, is := obj.(*core.PdfObjectName)
		if !is || string(*oname) != "Font" {
			common.Log.Debug("Incompatibility ERROR: Type (Required) defined but not Font name")
			return nil, errors.New("Range check error")
		}
	} else {
		common.Log.Debug("Incompatibility ERROR: Type (Required) missing")
		return nil, errors.New("Required attribute missing")
	}

	obj, isDefined := (*d)["Subtype"]
	if !isDefined {
		common.Log.Debug("Incompatibility ERROR: Subtype (Required) missing")
		return nil, errors.New("Required attribute missing")
	}

	subtype, ok := core.TraceToDirectObject(obj).(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("Incompatibility ERROR: subtype not a name (%T) ", obj)
		return nil, errors.New("Type check error")
	}

	switch subtype.String() {
	case "TrueType":
		truefont, err := newPdfFontTrueTypeFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error loading truetype font: %v", truefont)
			return nil, err
		}

		font.context = truefont
	default:
		common.Log.Debug("Unsupported font type: %s", subtype.String())
		return nil, errors.New("Unsupported font type")
	}

	return font, nil
}

func (font PdfFont) ToPdfObject() core.PdfObject {
	switch f := font.context.(type) {
	case *pdfFontTrueType:
		return f.ToPdfObject()
	}

	// If not supported, return null..
	common.Log.Debug("Unsupported font (%T) - returning null object", font.context)
	return core.MakeNull()
}

type pdfFontTrueType struct {
	Encoder textencoding.TextEncoder

	firstChar  int
	lastChar   int
	charWidths []float64

	// Subtype shall be TrueType.
	// Encoding is subject to limitations that are described in 9.6.6, "Character Encoding".
	// BaseFont is derived differently.
	BaseFont       core.PdfObject
	FirstChar      core.PdfObject
	LastChar       core.PdfObject
	Widths         core.PdfObject
	FontDescriptor *PdfFontDescriptor
	Encoding       core.PdfObject
	ToUnicode      core.PdfObject

	container *core.PdfIndirectObject
}

func (font pdfFontTrueType) SetEncoder(encoder textencoding.TextEncoder) {
	font.Encoder = encoder
}

func (font pdfFontTrueType) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	metrics := fonts.CharMetrics{}

	code, found := font.Encoder.GlyphNameToCharcode(glyph)
	if !found {
		return metrics, false
	}

	if int(code) < font.firstChar {
		common.Log.Debug("Code lower than firstchar (%d < %d)", code, font.firstChar)
		return metrics, false
	}

	if int(code) > font.lastChar {
		common.Log.Debug("Code higher than lastchar (%d < %d)", code, font.lastChar)
		return metrics, false
	}

	index := int(code) - font.firstChar
	if index >= len(font.charWidths) {
		common.Log.Debug("Code outside of widths range")
		return metrics, false
	}

	width := font.charWidths[index]
	metrics.Wx = width

	return metrics, true
}

func newPdfFontTrueTypeFromPdfObject(obj core.PdfObject) (*pdfFontTrueType, error) {
	font := &pdfFontTrueType{}

	if ind, is := obj.(*core.PdfIndirectObject); is {
		font.container = ind
		obj = ind.PdfObject
	}

	d, ok := obj.(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Font object invalid, not a dictionary (%T)", obj)
		return nil, errors.New("Type check error")
	}

	if obj, isDefined := (*d)["Type"]; isDefined {
		oname, is := obj.(*core.PdfObjectName)
		if !is || oname.String() != "Font" {
			common.Log.Debug("Incompatibility: Type defined but not Font")
		}
	}

	if obj, isDefined := (*d)["Subtype"]; isDefined {
		oname, is := obj.(*core.PdfObjectName)
		if !is || oname.String() != "TrueType" {
			common.Log.Debug("Incompatibility: Loading TrueType font but Subtype != TrueType")
		}
	}

	if obj, isDefined := (*d)["BaseFont"]; isDefined {
		font.BaseFont = obj
	}

	if obj, isDefined := (*d)["FirstChar"]; isDefined {
		font.FirstChar = obj

		intVal, ok := core.TraceToDirectObject(obj).(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("Invalid FirstChar type (%T)", obj)
			return nil, errors.New("Type check error")
		}
		font.firstChar = int(*intVal)
	} else {
		common.Log.Debug("ERROR: FirstChar attribute missing")
		return nil, errors.New("Required attribute missing")
	}

	if obj, isDefined := (*d)["LastChar"]; isDefined {
		font.LastChar = obj

		intVal, ok := core.TraceToDirectObject(obj).(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("Invalid LastChar type (%T)", obj)
			return nil, errors.New("Type check error")
		}
		font.lastChar = int(*intVal)
	} else {
		common.Log.Debug("ERROR: FirstChar attribute missing")
		return nil, errors.New("Required attribute missing")
	}

	font.charWidths = []float64{}
	if obj, isDefined := (*d)["Widths"]; isDefined {
		font.Widths = obj

		arr, ok := core.TraceToDirectObject(obj).(*core.PdfObjectArray)
		if !ok {
			common.Log.Debug("Widths attribute != array (%T)", arr)
			return nil, errors.New("Type check error")
		}

		widths, err := arr.ToFloat64Array()
		if err != nil {
			common.Log.Debug("Error converting widths to array")
			return nil, err
		}

		if len(widths) != (font.lastChar - font.firstChar + 1) {
			common.Log.Debug("Invalid widths length != %d (%d)", font.lastChar-font.firstChar+1, len(widths))
			return nil, errors.New("Range check error")
		}

		font.charWidths = widths
	} else {
		common.Log.Debug("Widths missing from font")
		return nil, errors.New("Required attribute missing")
	}

	if obj, isDefined := (*d)["FontDescriptor"]; isDefined {
		descriptor, err := newPdfFontDescriptorFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error loading font descriptor: %v", err)
			return nil, err
		}

		font.FontDescriptor = descriptor
	}

	if obj, isDefined := (*d)["Encoding"]; isDefined {
		font.Encoding = obj
	}

	if obj, isDefined := (*d)["ToUnicode"]; isDefined {
		font.ToUnicode = obj
	}

	return font, nil
}

func (this *pdfFontTrueType) ToPdfObject() core.PdfObject {
	if this.container == nil {
		this.container = &core.PdfIndirectObject{}
	}
	d := &core.PdfObjectDictionary{}
	this.container.PdfObject = d

	d.Set("Type", core.MakeName("Font"))
	d.Set("Subtype", core.MakeName("TrueType"))

	if this.BaseFont != nil {
		d.Set("BaseFont", this.BaseFont)
	}
	if this.FirstChar != nil {
		d.Set("FirstChar", this.FirstChar)
	}
	if this.LastChar != nil {
		d.Set("LastChar", this.LastChar)
	}
	if this.Widths != nil {
		d.Set("Widths", this.Widths)
	}
	if this.FontDescriptor != nil {
		d.Set("FontDescriptor", this.FontDescriptor.ToPdfObject())
	}
	if this.Encoding != nil {
		d.Set("Encoding", this.Encoding)
	}
	if this.ToUnicode != nil {
		d.Set("ToUnicode", this.ToUnicode)
	}

	return this.container
}

func NewPdfFontFromTTFFile(filePath string) (*PdfFont, error) {
	ttf, err := fonts.TtfParse(filePath)
	if err != nil {
		common.Log.Debug("Error loading ttf font: %v", err)
		return nil, err
	}

	truefont := &pdfFontTrueType{}

	truefont.Encoder = textencoding.NewWinAnsiTextEncoder()
	truefont.firstChar = 32
	truefont.lastChar = 255

	truefont.BaseFont = core.MakeName(ttf.PostScriptName)
	truefont.FirstChar = core.MakeInteger(32)
	truefont.LastChar = core.MakeInteger(255)

	k := 1000.0 / float64(ttf.UnitsPerEm)

	if len(ttf.Widths) <= 0 {
		return nil, errors.New("Missing required attribute (Widths)")
	}

	missingWidth := k * float64(ttf.Widths[0])
	vals := []float64{}

	for charcode := 32; charcode <= 255; charcode++ {
		runeVal, found := truefont.Encoder.CharcodeToRune(byte(charcode))
		if !found {
			common.Log.Debug("Rune not found (charcode: %d)", charcode)
			vals = append(vals, missingWidth)
			continue
		}

		pos, ok := ttf.Chars[uint16(runeVal)]
		if !ok {
			common.Log.Debug("Rune not in TTF Chars")
			vals = append(vals, missingWidth)
			continue
		}

		w := k * float64(ttf.Widths[pos])

		vals = append(vals, w)
	}

	truefont.Widths = &core.PdfIndirectObject{PdfObject: core.MakeArrayFromFloats(vals)}

	if len(vals) < (255 - 32 + 1) {
		common.Log.Debug("Invalid length of widths, %d < %d", len(vals), 255-32+1)
		return nil, errors.New("Range check error")
	}

	truefont.charWidths = vals[:255-32+1]

	// Default.
	// XXX/FIXME TODO: Only use the encoder object.

	truefont.Encoding = core.MakeName("WinAnsiEncoding")

	descriptor := &PdfFontDescriptor{}
	descriptor.Ascent = core.MakeFloat(k * float64(ttf.TypoAscender))
	descriptor.Descent = core.MakeFloat(k * float64(ttf.TypoDescender))
	descriptor.CapHeight = core.MakeFloat(k * float64(ttf.CapHeight))
	descriptor.FontBBox = core.MakeArrayFromFloats([]float64{k * float64(ttf.Xmin), k * float64(ttf.Ymin), k * float64(ttf.Xmax), k * float64(ttf.Ymax)})
	descriptor.ItalicAngle = core.MakeFloat(float64(ttf.ItalicAngle))
	descriptor.MissingWidth = core.MakeFloat(k * float64(ttf.Widths[0]))

	ttfBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		common.Log.Debug("Unable to read file contents: %v", err)
		return nil, err
	}

	// XXX/TODO: Encode the file...
	stream, err := core.MakeStream(ttfBytes, core.NewFlateEncoder())
	if err != nil {
		common.Log.Debug("Unable to make stream: %v", err)
		return nil, err
	}
	stream.PdfObjectDictionary.Set("Length1", core.MakeInteger(int64(len(ttfBytes))))
	descriptor.FontFile2 = stream

	if ttf.Bold {
		descriptor.StemV = core.MakeInteger(120)
	} else {
		descriptor.StemV = core.MakeInteger(70)
	}

	// Flags.
	flags := 1 << 5
	if ttf.IsFixedPitch {
		flags |= 1
	}
	if ttf.ItalicAngle != 0 {
		flags |= 1 << 6
	}
	descriptor.Flags = core.MakeInteger(int64(flags))

	// Build Font.
	truefont.FontDescriptor = descriptor

	font := &PdfFont{}
	font.context = truefont

	return font, nil
}

// Font descriptors specifies metrics and other attributes of a font.
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

// Load the font descriptor from a PdfObject.  Can either be a *PdfIndirectObject or
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

	if obj, isDefined := (*d)["Type"]; isDefined {
		oname, is := obj.(*core.PdfObjectName)
		if !is || string(*oname) != "FontDescriptor" {
			common.Log.Debug("Incompatibility: Font descriptor Type invalid (%T)", obj)
		}
	} else {
		common.Log.Debug("Incompatibility: Type (Required) missing")
	}

	if obj, isDefined := (*d)["FontName"]; isDefined {
		descriptor.FontName = obj
	} else {
		common.Log.Debug("Incompatibility: FontName (Required) missing")
	}

	if obj, isDefined := (*d)["FontFamily"]; isDefined {
		descriptor.FontFamily = obj
	}

	if obj, isDefined := (*d)["FontStretch"]; isDefined {
		descriptor.FontStretch = obj
	}

	if obj, isDefined := (*d)["FontWeight"]; isDefined {
		descriptor.FontWeight = obj
	}

	if obj, isDefined := (*d)["Flags"]; isDefined {
		descriptor.Flags = obj
	}

	if obj, isDefined := (*d)["FontBBox"]; isDefined {
		descriptor.FontBBox = obj
	}

	if obj, isDefined := (*d)["ItalicAngle"]; isDefined {
		descriptor.ItalicAngle = obj
	}

	if obj, isDefined := (*d)["Ascent"]; isDefined {
		descriptor.Ascent = obj
	}

	if obj, isDefined := (*d)["Descent"]; isDefined {
		descriptor.Descent = obj
	}

	if obj, isDefined := (*d)["Leading"]; isDefined {
		descriptor.Leading = obj
	}

	if obj, isDefined := (*d)["CapHeight"]; isDefined {
		descriptor.CapHeight = obj
	}

	if obj, isDefined := (*d)["XHeight"]; isDefined {
		descriptor.XHeight = obj
	}

	if obj, isDefined := (*d)["StemV"]; isDefined {
		descriptor.StemV = obj
	}

	if obj, isDefined := (*d)["StemH"]; isDefined {
		descriptor.StemH = obj
	}

	if obj, isDefined := (*d)["AvgWidth"]; isDefined {
		descriptor.AvgWidth = obj
	}

	if obj, isDefined := (*d)["MaxWidth"]; isDefined {
		descriptor.MaxWidth = obj
	}

	if obj, isDefined := (*d)["MissingWidth"]; isDefined {
		descriptor.MissingWidth = obj
	}

	if obj, isDefined := (*d)["FontFile"]; isDefined {
		descriptor.FontFile = obj
	}

	if obj, isDefined := (*d)["FontFile2"]; isDefined {
		descriptor.FontFile2 = obj
	}

	if obj, isDefined := (*d)["FontFile3"]; isDefined {
		descriptor.FontFile3 = obj
	}

	if obj, isDefined := (*d)["CharSet"]; isDefined {
		descriptor.CharSet = obj
	}

	if obj, isDefined := (*d)["Style"]; isDefined {
		descriptor.Style = obj
	}

	if obj, isDefined := (*d)["Lang"]; isDefined {
		descriptor.Lang = obj
	}

	if obj, isDefined := (*d)["FD"]; isDefined {
		descriptor.FD = obj
	}

	if obj, isDefined := (*d)["CIDSet"]; isDefined {
		descriptor.CIDSet = obj
	}

	return descriptor, nil
}

// Convert to a PDF dictionary inside an indirect object.
func (this *PdfFontDescriptor) ToPdfObject() core.PdfObject {
	d := &core.PdfObjectDictionary{}
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
