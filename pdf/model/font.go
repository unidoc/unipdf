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
	"github.com/unidoc/unidoc/pdf/internal/cmap"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// PdfFont represents an underlying font structure which can be of type:
// - Type0
// - Type1
// - TrueType
// etc.
type PdfFont struct {
	fontSkeleton
	context fonts.Font // The underlying font: Type0, Type1, Truetype, etc..
	dict    *PdfObjectDictionary
}

// fontSkeleton represents the fields that are common to all PDF fonts
type fontSkeleton struct {
	dict *PdfObjectDictionary

	subtype  string
	basefont string

	BaseFont  PdfObject
	Subtype   PdfObject
	ToUnicode PdfObject

	// ToUnicode cmap
	UCMap          *cmap.CMap
	fontDescriptor *PdfFontDescriptor
}

// toFont returns a PdfObjectDictionary for `font`.
// It is for use in font ToPdfObject functions.
// NOTE: The returned dict's SubType is set to `subtype` if font doesn't have a subtype
func (skel fontSkeleton) toDict(subtype string) *PdfObjectDictionary {

	if subtype != "" && skel.subtype != "" {
		common.Log.Debug("ERROR: toDict. Overriding subtype to %#q %s", subtype, skel)
	} else if subtype == "" && skel.subtype == "" {
		common.Log.Debug("ERROR: toDict no subtype. font=%s", skel)
	} else if skel.subtype == "" {
		skel.subtype = subtype
	}

	d := MakeDict()
	d.Set("Type", MakeName("Font"))
	d.Set("Subtype", MakeName(skel.subtype))
	if skel.BaseFont != nil {
		d.Set("BaseFont", skel.BaseFont)
	}
	if skel.fontDescriptor != nil {
		d.Set("FontDescriptor", skel.fontDescriptor.ToPdfObject())
	}
	if skel.ToUnicode != nil {
		d.Set("ToUnicode", skel.ToUnicode)
	}
	return d
}

// String returns a string that describes `font`.
func (font PdfFont) String() string {
	return fmt.Sprintf("%T %s", font.context, font.fontSkeleton.String())
}

// String returns a string that describes `skel`.
func (skel fontSkeleton) String() string {
	descriptor := ""
	if skel.fontDescriptor != nil {
		descriptor = "(has descriptor)"
	}
	return fmt.Sprintf("%#q %#q %s", skel.subtype, skel.basefont, descriptor)
}

// CharcodeBytesToUnicode converts PDF character codes `data` to a Go unicode string.
// (page 291)
// The process of finding glyph descriptions in OpenType fonts by a conforming reader shall be the following:
// • For Type 1 fonts using “CFF” tables, the process shall be as described in 9.6.6.2, "Encodings
//   for Type 1 Fonts".
// • For TrueType fonts using “glyf” tables, the process shall be as described in 9.6.6.4,
//   "Encodings for TrueType Fonts". Since this process sometimes produces ambiguous results,
//   conforming writers, instead of using a simple font, shall use a Type 0 font with an Identity-H
//   encoding and use the glyph indices as character codes, as described following Table 118.
func (font PdfFont) CharcodeBytesToUnicode(data []byte) string {
	if font.UCMap != nil {
		unicode, ok := font.UCMap.CharcodeBytesToUnicode(data)
		if ok {
			return unicode
		}
	}
	// Fall back to encoding
	charcodes := []uint16{}
	if font.isCIDFont() {
		if len(data) == 1 {
			data = []byte{0, data[0]}
		}
		if len(data)%2 != 0 {
			common.Log.Debug("ERROR: Padding data=%+v to even length", data)
			data = append(data, 0)
		}
		for i := 0; i < len(data); i += 2 {
			b := uint16(data[i])<<8 | uint16(data[i+1])
			charcodes = append(charcodes, b)
		}
	} else {
		for _, b := range data {
			charcodes = append(charcodes, uint16(b))
		}
	}
	if encoder := font.Encoder(); encoder != nil {
		runes := []rune{}
		for _, code := range charcodes {
			r, ok := encoder.CharcodeToRune(code)
			if !ok {
				common.Log.Debug("ERROR: No rune. code=0x%04x font=%s encoding=%s data = [% 02x]=%#q",
					code, font, encoder, data, data)
				r = cmap.MissingCodeRune
			}
			runes = append(runes, r)
		}
		return string(runes)
	}
	common.Log.Debug("ERROR: Couldn't convert to unicode. Using input. data=%#q=[% 02x] font=%s",
		string(data), data, font)
	return string(data)
}

// isCIDFont returns true if `skel` is a CID font.
func (skel fontSkeleton) isCIDFont() bool {
	if skel.subtype == "" {
		common.Log.Debug("ERROR: isCIDFont. context is nil. font=%s", skel)
	}
	isCID := false
	switch skel.subtype {
	case "Type0", "CIDFontType0", "CIDFontType2":
		isCID = true
	}
	common.Log.Trace("isCIDFont: isCID=%t font=%s", isCID, skel)
	return isCID
}

// actualFont returns the Font in font.context
func (font PdfFont) actualFont() fonts.Font {
	if font.context == nil {
		common.Log.Debug("ERROR: actualFont. context is nil. font=%s", font)
	}
	switch t := font.context.(type) {
	case *pdfFontSimple:
		return t
	case *pdfFontType0:
		return t
	case *pdfCIDFontType0:
		return t
	case *pdfCIDFontType2:
		return t
	case fonts.FontCourier:
		return t
	case fonts.FontCourierBold:
		return t
	case fonts.FontCourierBoldOblique:
		return t
	case fonts.FontCourierOblique:
		return t
	case fonts.FontHelvetica:
		return t
	case fonts.FontHelveticaBold:
		return t
	case fonts.FontHelveticaBoldOblique:
		return t
	case fonts.FontHelveticaOblique:
		return t
	case fonts.FontTimesRoman:
		return t
	case fonts.FontTimesBold:
		return t
	case fonts.FontTimesBoldItalic:
		return t
	case fonts.FontTimesItalic:
		return t
	case fonts.FontSymbol:
		return t
	case fonts.FontZapfDingbats:
		return t
	default:
		common.Log.Debug("ERROR: actualFont. Unknown font type %t. font=%s", t, font)
		return nil
	}
}

// Encoder returns the font's text encoder.
func (font PdfFont) Encoder() textencoding.TextEncoder {
	t := font.actualFont()
	if t == nil {
		common.Log.Debug("ERROR: Encoder not implemented for font type=%#T", font.context)
		// XXX: Should we return a default encoding?
		return nil
	}
	return t.Encoder()
}

// SetEncoder sets the encoding for the underlying font.
// !@#$ Is this only possible for simple fonts?
func (font PdfFont) SetEncoder(encoder textencoding.TextEncoder) {
	t := font.actualFont()
	if t == nil {
		common.Log.Debug("ERROR: SetEncoder. Not implemented for font type=%#T", font.context)
		return
	}
	t.SetEncoder(encoder)
}

// GetGlyphCharMetrics returns the specified char metrics for a specified glyph name.
func (font PdfFont) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	t := font.actualFont()
	if t == nil {
		common.Log.Debug("ERROR: GetGlyphCharMetrics Not implemented for font type=%#T", font.context)
		return fonts.CharMetrics{}, false
	}
	return t.GetGlyphCharMetrics(glyph)

}

// NewPdfFontFromPdfObject loads a PdfFont from a dictionary.  If there is a problem an error is
// returned.
func NewPdfFontFromPdfObject(fontObj PdfObject) (*PdfFont, error) {
	return newPdfFontFromPdfObject(fontObj, true)
}

// newPdfFontFromPdfObject loads a PdfFont from a dictionary.  If there is a problem an error is
// returned.
// The allowType0 indicates whether loading Type0 font should be supported.  Flag used to avoid
// cyclical loading.
func newPdfFontFromPdfObject(fontObj PdfObject, allowType0 bool) (*PdfFont, error) {
	skeleton, err := newFontSkeletonFromPdfObject(fontObj)
	if err != nil {
		return nil, err
	}
	font := &PdfFont{fontSkeleton: *skeleton}

	switch skeleton.subtype {
	case "Type0":
		if !allowType0 {
			common.Log.Debug("ERROR: Loading type0 not allowed. font=%s", font)
			return nil, errors.New("Cyclical type0 loading")
		}
		type0font, err := newPdfFontType0FromPdfObject(fontObj, skeleton)
		if err != nil {
			common.Log.Debug("ERROR: While loading Type0 font. font=%s err=%v", font, err)
			return nil, err
		}
		font.context = type0font
	case "Type1", "Type3", "MMType1", "TrueType": // !@#$
		if std, ok := fonts.Standard14Fonts[font.basefont]; ok && font.subtype == "Type1" {
			font.context = std
		} else {
			simplefont, err := newSimpleFontFromPdfObject(fontObj, skeleton)
			if err != nil {
				common.Log.Debug("ERROR: While loading simple font: font=%s err=%v", font, err)
				return nil, err
			}
			font.context = simplefont
		}
	case "CIDFontType0":
		cidfont, err := newPdfCIDFontType0FromPdfObject(fontObj, skeleton)
		if err != nil {
			common.Log.Debug("ERROR: While loading cid font type0 font: %v", err)
			return nil, err
		}
		font.context = cidfont
	case "CIDFontType2":
		cidfont, err := newPdfCIDFontType2FromPdfObject(fontObj, skeleton)
		if err != nil {
			common.Log.Debug("ERROR: While loading cid font type2 font. font=%s err=%v", font, err)
			return nil, err
		}
		font.context = cidfont
	default:
		common.Log.Debug("ERROR: Unsupported font type: font=%s", font)
		return nil, ErrUnsupportedFont
	}

	return font, nil
}

// newFontSkeletonFromPdfObject loads a fontSkeleton from a dictionary.  If there is a problem an error is
// returned.
// The allowType0 indicates whether loading Type0 font should be supported.  Flag used to avoid
// cyclical loading.
func newFontSkeletonFromPdfObject(fontObj PdfObject) (*fontSkeleton, error) {
	font := &fontSkeleton{}

	dictObj := fontObj
	if ind, is := fontObj.(*PdfIndirectObject); is {
		dictObj = ind.PdfObject
	}

	d, ok := dictObj.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Font not given by a dictionary (%T)", fontObj)
		return nil, ErrUnsupportedFont
	}
	font.dict = d

	basefont, err := GetName(d.Get("BaseFont"))
	if err == nil {
		font.basefont = basefont
		font.BaseFont = d.Get("BaseFont")
	}

	if obj := d.Get("Type"); obj != nil {
		oname, is := obj.(*PdfObjectName)
		if !is || string(*oname) != "Font" {
			common.Log.Debug("ERROR: Font Incompatibility. Type=%q Should be %q", string(*oname), "Font")
			return nil, ErrRangeError
		}
	} else {
		common.Log.Debug("ERROR: Font Incompatibility. Type (Required) missing")
		return nil, ErrRequiredAttributeMissing
	}

	obj := d.Get("Subtype")
	if obj == nil {
		common.Log.Debug("ERROR: Font Incompatibility. Subtype (Required) missing")
		return nil, ErrRequiredAttributeMissing
	}
	subtype, err := GetName(TraceToDirectObject(obj))
	if err != nil {
		common.Log.Debug("ERROR: Font Incompatibility. subtype not a name (%T) font=%s", obj, font)
		return nil, ErrTypeError
	}
	font.subtype = subtype

	obj = d.Get("FontDescriptor")
	if obj != nil {
		fontDescriptor, err := newPdfFontDescriptorFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("ERROR: Bad font descriptor")
			return nil, ErrRequiredAttributeMissing
		}
		if err == nil {
			font.fontDescriptor = fontDescriptor
		}
	}

	font.ToUnicode = TraceToDirectObject(d.Get("ToUnicode"))

	if font.ToUnicode != nil {
		codemap, err := toUnicodeToCmap(font.ToUnicode, font.isCIDFont())
		if err != nil {
			return nil, err
		}
		font.UCMap = codemap
	}

	return font, nil
}

// ToPdfObject converts the PdfFont object to its PDF representation.
func (font PdfFont) ToPdfObject() PdfObject {
	if t := font.actualFont(); t != nil {
		return t.ToPdfObject()
	}
	common.Log.Debug("ERROR: ToPdfObject Not implemented for font type=%#T. Returning null object",
		font.context)
	return MakeNull()
}

// toUnicodeToCmap returns a CMap of `toUnicode` if it exists
// 9.10.3 ToUnicode CMaps (page 29)
// The CMap defined in the ToUnicode entry of the font dictionary shall follow the syntax for CMaps
// This CMap differs from an ordinary one in these ways:
// • The only pertinent entry in the CMap stream dictionary (see Table 120) is UseCMap, which may be
//   used if the CMap is based on another ToUnicode CMap.
// • The CMap file shall contain begincodespacerange and endcodespacerange operators that are
//   consistent with the encoding that the font uses. In particular, for a simple font, the
//   codespace shall be one byte long.
// • It shall use the beginbfchar, endbfchar, beginbfrange, and endbfrange operators to define the
//   mapping from character codes to Unicode character sequences expressed in UTF-16BE encoding
func toUnicodeToCmap(toUnicode PdfObject, isCID bool) (*cmap.CMap, error) {
	toUnicodeStream, ok := toUnicode.(*PdfObjectStream)
	if !ok {
		common.Log.Debug("ERROR: toUnicodeToCmap: Not a stream (%T)", toUnicode)
		return nil, errors.New("Invalid ToUnicode entry - not a stream")
	}
	data, err := DecodeStream(toUnicodeStream)
	if err != nil {
		return nil, err
	}
	return cmap.LoadCmapFromData(data, !isCID)
}

// PdfFontDescriptor specifies metrics and other attributes of a font and can refer to a FontFile
// for embedded fonts.
// 9.8 Font Descriptors (page 281)
type PdfFontDescriptor struct {
	FontName     PdfObject
	FontFamily   PdfObject
	FontStretch  PdfObject
	FontWeight   PdfObject
	Flags        PdfObject
	FontBBox     PdfObject
	ItalicAngle  PdfObject
	Ascent       PdfObject
	Descent      PdfObject
	Leading      PdfObject
	CapHeight    PdfObject
	XHeight      PdfObject
	StemV        PdfObject
	StemH        PdfObject
	AvgWidth     PdfObject
	MaxWidth     PdfObject
	MissingWidth PdfObject
	FontFile     PdfObject
	FontFile2    PdfObject
	FontFile3    PdfObject
	CharSet      PdfObject

	// Additional entries for CIDFonts
	Style  PdfObject
	Lang   PdfObject
	FD     PdfObject
	CIDSet PdfObject

	// Container.
	container *PdfIndirectObject
}

// newPdfFontDescriptorFromPdfObject loads the font descriptor from a PdfObject.  Can either be a
// *PdfIndirectObject or a *PdfObjectDictionary.
func newPdfFontDescriptorFromPdfObject(obj PdfObject) (*PdfFontDescriptor, error) {
	descriptor := &PdfFontDescriptor{}

	if ind, is := obj.(*PdfIndirectObject); is {
		descriptor.container = ind
		obj = ind.PdfObject
	}

	d, ok := obj.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: FontDescriptor not given by a dictionary (%T)", obj)
		return nil, ErrTypeError
	}

	if obj := d.Get("FontName"); obj != nil {
		descriptor.FontName = obj
	} else {
		common.Log.Debug("Incompatibility: FontName (Required) missing")
	}
	fontname, _ := GetName(descriptor.FontName)

	if obj := d.Get("Type"); obj != nil {
		oname, is := obj.(*PdfObjectName)
		if !is || string(*oname) != "FontDescriptor" {
			common.Log.Debug("Incompatibility: Font descriptor Type invalid (%T) font=%q %T",
				obj, fontname, descriptor.FontName)
		}
	} else {
		common.Log.Trace("Incompatibility: Type (Required) missing. font=%q %T",
			fontname, descriptor.FontName)
		// return nil, errors.New("$$$$$")
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
func (this *PdfFontDescriptor) ToPdfObject() PdfObject {
	d := MakeDict()
	if this.container == nil {
		this.container = &PdfIndirectObject{}
	}
	this.container.PdfObject = d

	d.Set("Type", MakeName("FontDescriptor"))

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
