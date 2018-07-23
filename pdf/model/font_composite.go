package model

import (
	"errors"
	"io/ioutil"
	"sort"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/cmap"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

/*
   9.7.2 CID-Keyed Fonts Overview (page 267)
   The CID-keyed font architecture specifies the external representation of certain font programs,
   called *CMap* and *CIDFont* files, along with some conventions for combining and using those files.

   A *CMap* (character map) file shall specify the correspondence between character codes and the CID
   numbers used to identify glyphs. It is equivalent to the concept of an encoding in simple fonts.
   Whereas a simple font allows a maximum of 256 glyphs to be encoded and accessible at one time, a
   CMap can describe a mapping from multiple-byte codes to thousands of glyphs in a large CID-keyed
   font.

   9.7.4 CIDFonts (page 269)

   A CIDFont program contains glyph descriptions that are accessed using a CID as the character
   selector. There are two types of CIDFonts:
   • A Type 0 CIDFont contains glyph descriptions based on CFF
   • A Type 2 CIDFont contains glyph descriptions based on the TrueType font format

   A CIDFont dictionary is a PDF object that contains information about a CIDFont program. Although
   its Type value is Font, a CIDFont is not actually a font.
       It does not have an Encoding entry,
       it may not be listed in the Font subdictionary of a resource dictionary, and
       it may not be used as the operand of the Tf operator.
       It shall be used only as a descendant of a Type 0 font.
   The CMap in the Type 0 font shall be what defines the encoding that maps character codes to CIDs
   in  the CIDFont.

    9.7.6 Type 0 Font Dictionaries (page 279)

    Type      Font
    Subtype   Type0
    BaseFont  (Required) The name of the font. If the descendant is a Type 0 CIDFont, this name
              should be the concatenation of the CIDFont’s BaseFont name, a hyphen, and the CMap
              name given in the Encoding entry (or the CMapName entry in the CMap). If the
              descendant is a Type 2 CIDFont, this name should be the same as the CIDFont’s BaseFont
              name.
              NOTE In principle, this is an arbitrary name, since there is no font program
                   associated directly with a Type 0 font dictionary. The conventions described here
                   ensure maximum compatibility with existing readers.
    Encoding name or stream (Required)
             The name of a predefined CMap, or a stream containing a CMap that maps character codes
             to font numbers and CIDs. If the descendant is a Type 2 CIDFont whose associated
             TrueType font program is not embedded in the PDF file, the Encoding entry shall be a
             predefined CMap name (see 9.7.4.2, "Glyph Selection in CIDFonts").

    Type 0 font from 000046.pdf

    103 0 obj
    << /Type /Font /Subtype /Type0 /Encoding /Identity-H /DescendantFonts [179 0 R]
    /BaseFont /FLDOLC+PingFangSC-Regular >>
    endobj
    179 0 obj
    << /Type /Font /Subtype /CIDFontType0 /BaseFont /FLDOLC+PingFangSC-Regular
    /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >>
    /W 180 0 R /DW 1000 /FontDescriptor 181 0 R >>
    endobj
    180 0 obj
    [ ]
    endobj
    181 0 obj
    << /Type /FontDescriptor /FontName /FLDOLC+PingFangSC-Regular /Flags 4 /FontBBox
    [-123 -263 1177 1003] /ItalicAngle 0 /Ascent 972 /Descent -232 /CapHeight
    864 /StemV 70 /XHeight 648 /StemH 64 /AvgWidth 1000 /MaxWidth 1300 /FontFile3
    182 0 R >>
    endobj
    182 0 obj
    << /Length 183 0 R /Subtype /CIDFontType0C /Filter /FlateDecode >>
    stream
    ....
*/

// pdfFontType0 represents a Type0 font in PDF. Used for composite fonts which can encode multiple
// bytes for complex symbols (e.g. used in Asian languages). Represents the root font whereas the
// associated CIDFont is called its descendant.
type pdfFontType0 struct {
	container *core.PdfIndirectObject

	// These fields are common to all PDF fonts.

	basefont string // The font's "BaseFont" field.
	subtype  string // The font's "Subtype" field.

	// These are optional fields in the PDF font
	toUnicode core.PdfObject // The stream containing toUnicodeCmap. We keep it around for ToPdfObject.

	// These objects are computed from optional fields in the PDF font
	toUnicodeCmap  *cmap.CMap         // Computed from "ToUnicode"
	fontDescriptor *PdfFontDescriptor // Computed from "FontDescriptor"

	// objectNumber helps us find the font in the PDF being processed. This helps with debugging
	objectNumber int64

	// These fields are specific to Type 0 fonts.

	encoder        textencoding.TextEncoder
	Encoding       core.PdfObject
	DescendantFont *PdfFont // Can be either CIDFontType0 or CIDFontType2 font.
}

// pdfFontType0FromSkeleton returns a pdfFontType0 with its common fields initalized.
func pdfFontType0FromSkeleton(base *fontCommon) *pdfFontType0 {
	return &pdfFontType0{
		basefont:       base.basefont,
		subtype:        base.subtype,
		toUnicode:      base.toUnicode,
		toUnicodeCmap:  base.toUnicodeCmap,
		fontDescriptor: base.fontDescriptor,
		objectNumber:   base.objectNumber,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfFontType0) baseFields() *fontCommon {
	return &fontCommon{
		basefont:       font.basefont,
		subtype:        font.subtype,
		toUnicode:      font.toUnicode,
		toUnicodeCmap:  font.toUnicodeCmap,
		fontDescriptor: font.fontDescriptor,
		objectNumber:   font.objectNumber,
	}
}

// GetGlyphCharMetrics returns the character metrics for the specified glyph.  A bool flag is
// returned to indicate whether or not the entry was found in the glyph to charcode mapping.
func (font pdfFontType0) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	if font.DescendantFont == nil {
		common.Log.Debug("ERROR: No descendant. font=%s", font)
		return fonts.CharMetrics{}, false
	}
	return font.DescendantFont.GetGlyphCharMetrics(glyph)
}

// Encoder returns the font's text encoder.
func (font pdfFontType0) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// SetEncoder sets the encoder for the truetype font.
func (font pdfFontType0) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

// ToPdfObject converts the pdfFontType0 to a PDF representation.
func (font *pdfFontType0) ToPdfObject() core.PdfObject {
	if font.container == nil {
		font.container = &core.PdfIndirectObject{}
	}
	d := font.baseFields().asPdfObjectDictionary("Type0")

	font.container.PdfObject = d

	if font.encoder != nil {
		d.Set("Encoding", font.encoder.ToPdfObject())
	}
	if font.DescendantFont != nil {
		// Shall be 1 element array.
		d.Set("DescendantFonts", core.MakeArray(font.DescendantFont.ToPdfObject()))
	}

	return font.container
}

// newPdfFontType0FromPdfObject makes a pdfFontType0 based on the input `d` in base.
// If a problem is encountered, an error is returned.
func newPdfFontType0FromPdfObject(d *core.PdfObjectDictionary, base *fontCommon) (*pdfFontType0, error) {

	// DescendantFonts.
	arr, ok := core.GetArrayVal(core.TraceToDirectObject(d.Get("DescendantFonts")))
	if !ok {
		common.Log.Debug("ERROR: Invalid DescendantFonts - not an array %s", base)
		return nil, core.ErrRangeError
	}
	if len(arr) != 1 {
		common.Log.Debug("ERROR: Array length != 1 (%d)", len(arr))
		return nil, core.ErrRangeError
	}
	df, err := newPdfFontFromPdfObject(arr[0], false)
	if err != nil {
		common.Log.Debug("ERROR: Failed loading descendant font: err=%v %s", err, base)
		return nil, err
	}

	font := pdfFontType0FromSkeleton(base)
	font.DescendantFont = df

	encoderName, ok := core.GetNameVal(core.TraceToDirectObject(d.Get("Encoding")))
	// XXX: FIXME This is not valid if encoder is not Identity-H !@#$
	if ok /*&& encoderName == "Identity-H"*/ {
		font.encoder = textencoding.NewIdentityTextEncoder(encoderName)
	}
	return font, nil
}

// pdfCIDFontType0 represents a CIDFont Type0 font dictionary.
// XXX: This is a stub.
type pdfCIDFontType0 struct {
	container *core.PdfIndirectObject

	// These fields are common to all PDF fonts.

	basefont string // The font's "BaseFont" field.
	subtype  string // The font's "Subtype" field.

	// These are optional fields in the PDF font
	toUnicode core.PdfObject // The stream containing toUnicodeCmap. We keep it around for ToPdfObject.

	// These objects are computed from optional fields in the PDF font
	toUnicodeCmap  *cmap.CMap         // Computed from "ToUnicode"
	fontDescriptor *PdfFontDescriptor // Computed from "FontDescriptor"

	// objectNumber helps us find the font in the PDF being processed. This helps with debugging
	objectNumber int64

	// These fields are specific to Type 0 fonts.

	encoder textencoding.TextEncoder

	// Table 117 – Entries in a CIDFont dictionary (page 269)
	CIDSystemInfo  core.PdfObject // (Required) Dictionary that defines the character collection of the CIDFont. See Table 116.
	FontDescriptor core.PdfObject // (Required) Describes the CIDFont’s default metrics other than its glyph widths
}

// pdfCIDFontType0FromSkeleton returns a pdfCIDFontType0 with its common fields initalized.
func pdfCIDFontType0FromSkeleton(base *fontCommon) *pdfCIDFontType0 {
	return &pdfCIDFontType0{
		basefont:       base.basefont,
		subtype:        base.subtype,
		toUnicode:      base.toUnicode,
		toUnicodeCmap:  base.toUnicodeCmap,
		fontDescriptor: base.fontDescriptor,
		objectNumber:   base.objectNumber,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfCIDFontType0) baseFields() *fontCommon {
	return &fontCommon{
		basefont:       font.basefont,
		subtype:        font.subtype,
		toUnicode:      font.toUnicode,
		toUnicodeCmap:  font.toUnicodeCmap,
		fontDescriptor: font.fontDescriptor,
		objectNumber:   font.objectNumber,
	}
}

// Encoder returns the font's text encoder.
func (font pdfCIDFontType0) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// SetEncoder sets the encoder for the truetype font.
func (font pdfCIDFontType0) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

// GetGlyphCharMetrics returns the character metrics for the specified glyph.  A bool flag is
// returned to indicate whether or not the entry was found in the glyph to charcode mapping.
// XXX: This is a stub.
func (font pdfCIDFontType0) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	return fonts.CharMetrics{}, true
}

// ToPdfObject converts the pdfCIDFontType0 to a PDF representation.
// XXX: This is a stub.
func (font *pdfCIDFontType0) ToPdfObject() core.PdfObject {
	return core.MakeNull()
}

// newPdfCIDFontType0FromPdfObject creates a pdfCIDFontType0 object from a dictionary (either direct
// or via indirect object). If a problem occurs with loading an error is returned.
// XXX: This is a stub.
func newPdfCIDFontType0FromPdfObject(d *core.PdfObjectDictionary, base *fontCommon) (*pdfCIDFontType0, error) {
	if base.subtype != "CIDFontType0" {
		common.Log.Debug("ERROR: Font SubType != CIDFontType0. font=%s", base)
		return nil, core.ErrRangeError
	}

	font := pdfCIDFontType0FromSkeleton(base)

	// CIDSystemInfo.
	obj := core.TraceToDirectObject(d.Get("CIDSystemInfo"))
	if obj == nil {
		common.Log.Debug("ERROR: CIDSystemInfo (Required) missing. font=%s", base)
		return nil, ErrRequiredAttributeMissing
	}
	font.CIDSystemInfo = obj

	return font, nil
}

// pdfCIDFontType2 represents a CIDFont Type2 font dictionary.
type pdfCIDFontType2 struct {
	container *core.PdfIndirectObject

	// These fields are common to all PDF fonts.

	basefont string // The font's "BaseFont" field.
	subtype  string // The font's "Subtype" field.

	// These are optional fields in the PDF font
	toUnicode core.PdfObject // The stream containing toUnicodeCmap. We keep it around for ToPdfObject.

	// These objects are computed from optional fields in the PDF font
	toUnicodeCmap  *cmap.CMap         // Computed from "ToUnicode"
	fontDescriptor *PdfFontDescriptor // Computed from "FontDescriptor"

	// objectNumber helps us find the font in the PDF being processed. This helps with debugging
	objectNumber int64

	// These fields are specific to Type 0 fonts.

	encoder   textencoding.TextEncoder // !@#$ In base?
	ttfParser *fonts.TtfType

	CIDSystemInfo core.PdfObject
	DW            core.PdfObject
	W             core.PdfObject
	DW2           core.PdfObject
	W2            core.PdfObject
	CIDToGIDMap   core.PdfObject

	// Mapping between unicode runes to widths.
	runeToWidthMap map[uint16]int

	// Also mapping between GIDs (glyph index) and width.
	gidToWidthMap map[uint16]int
}

// pdfCIDFontType2FromSkeleton returns a pdfCIDFontType2 with its common fields initalized.
func pdfCIDFontType2FromSkeleton(base *fontCommon) *pdfCIDFontType2 {
	return &pdfCIDFontType2{
		basefont:       base.basefont,
		subtype:        base.subtype,
		toUnicode:      base.toUnicode,
		toUnicodeCmap:  base.toUnicodeCmap,
		fontDescriptor: base.fontDescriptor,
		objectNumber:   base.objectNumber,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfCIDFontType2) baseFields() *fontCommon {
	return &fontCommon{
		basefont:       font.basefont,
		subtype:        font.subtype,
		toUnicode:      font.toUnicode,
		toUnicodeCmap:  font.toUnicodeCmap,
		fontDescriptor: font.fontDescriptor,
		objectNumber:   font.objectNumber,
	}
}

// Encoder returns the font's text encoder.
func (font pdfCIDFontType2) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// SetEncoder sets the encoder for the truetype font.
func (font pdfCIDFontType2) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

// GetGlyphCharMetrics returns the character metrics for the specified glyph.  A bool flag is
// returned to indicate whether or not the entry was found in the glyph to charcode mapping.
func (font pdfCIDFontType2) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	metrics := fonts.CharMetrics{}

	enc := textencoding.NewTrueTypeFontEncoder(font.ttfParser.Chars)

	// Convert the glyph to character code.
	rune, found := enc.GlyphToRune(glyph)
	if !found {
		common.Log.Debug("Unable to convert glyph %q to charcode (identity)", glyph)
		return metrics, false
	}

	w, found := font.runeToWidthMap[uint16(rune)]
	if !found {
		return metrics, false
	}
	metrics.GlyphName = glyph
	metrics.Wx = float64(w)

	return metrics, true
}

// ToPdfObject converts the pdfCIDFontType2 to a PDF representation.
func (font *pdfCIDFontType2) ToPdfObject() core.PdfObject {
	if font.container == nil {
		font.container = &core.PdfIndirectObject{}
	}
	d := font.baseFields().asPdfObjectDictionary("CIDFontType2")
	font.container.PdfObject = d

	if font.CIDSystemInfo != nil {
		d.Set("CIDSystemInfo", font.CIDSystemInfo)
	}
	if font.DW != nil {
		d.Set("DW", font.DW)
	}
	if font.DW2 != nil {
		d.Set("DW2", font.DW2)
	}
	if font.W != nil {
		d.Set("W", font.W)
	}
	if font.W2 != nil {
		d.Set("W2", font.W2)
	}
	if font.CIDToGIDMap != nil {
		d.Set("CIDToGIDMap", font.CIDToGIDMap)
	}

	return font.container
}

// newPdfCIDFontType2FromPdfObject creates a pdfCIDFontType2 object from a dictionary (either direct
// or via indirect object). If a problem occurs with loading, an error is returned.
func newPdfCIDFontType2FromPdfObject(d *core.PdfObjectDictionary, base *fontCommon) (*pdfCIDFontType2, error) {
	if base.subtype != "CIDFontType2" {
		common.Log.Debug("ERROR: Font SubType != CIDFontType2. font=%s", base)
		return nil, core.ErrRangeError
	}

	font := pdfCIDFontType2FromSkeleton(base)

	// CIDSystemInfo.
	obj := d.Get("CIDSystemInfo")
	if obj == nil {
		common.Log.Debug("ERROR: CIDSystemInfo (Required) missing. font=%s", base)
		return nil, ErrRequiredAttributeMissing
	}
	font.CIDSystemInfo = obj

	// Optional attributes.
	font.DW = d.Get("DW")
	font.W = d.Get("W")
	font.DW2 = d.Get("DW2")
	font.W2 = d.Get("W2")
	font.CIDToGIDMap = d.Get("CIDToGIDMap")

	return font, nil
}

// NewCompositePdfFontFromTTFFile loads a composite font from a TTF font file. Composite fonts can
// be used to represent unicode fonts which can have multi-byte character codes, representing a wide
// range of values.
// It is represented by a Type0 Font with an underlying CIDFontType2 and an Identity-H encoding map.
// TODO: May be extended in the future to support a larger variety of CMaps and vertical fonts.
func NewCompositePdfFontFromTTFFile(filePath string) (*PdfFont, error) {
	// Load the truetype font data.
	ttf, err := fonts.TtfParse(filePath)
	if err != nil {
		common.Log.Debug("ERROR: while loading ttf font: %v", err)
		return nil, err
	}

	// Prepare the inner descendant font (CIDFontType2).
	cidfont := &pdfCIDFontType2{subtype: "CIDFontType2"}
	cidfont.ttfParser = &ttf

	// 2-byte character codes ➞ runes
	runes := make([]uint16, 0, len(ttf.Chars))
	for r := range ttf.Chars {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	base := fontCommon{
		subtype:  "Type0",
		basefont: ttf.PostScriptName,
	}

	k := 1000.0 / float64(ttf.UnitsPerEm)

	if len(ttf.Widths) <= 0 {
		return nil, errors.New("ERROR: Missing required attribute (Widths)")
	}

	missingWidth := k * float64(ttf.Widths[0])

	// Construct a rune ➞ width map.
	runeToWidthMap := map[uint16]int{}
	gidToWidthMap := map[uint16]int{}
	for _, r := range runes {
		glyphIndex := ttf.Chars[r]

		w := k * float64(ttf.Widths[glyphIndex])
		runeToWidthMap[r] = int(w)
		gidToWidthMap[glyphIndex] = int(w)
	}
	cidfont.runeToWidthMap = runeToWidthMap
	cidfont.gidToWidthMap = gidToWidthMap

	// Default width.
	cidfont.DW = core.MakeInteger(int64(missingWidth))

	// Construct W array.  Stores character code to width mappings.
	wArr := &core.PdfObjectArray{}
	i := uint16(0)
	for int(i) < len(runes) {

		j := i + 1
		for int(j) < len(runes) {
			if runeToWidthMap[runes[i]] != runeToWidthMap[runes[j]] {
				break
			}
			j++
		}

		// The W maps from CID to width, here CID = GID.
		gid1 := ttf.Chars[runes[i]]
		gid2 := ttf.Chars[runes[j-1]]

		wArr.Append(core.MakeInteger(int64(gid1)))
		wArr.Append(core.MakeInteger(int64(gid2)))
		wArr.Append(core.MakeInteger(int64(runeToWidthMap[runes[i]])))

		i = j
	}
	cidfont.W = core.MakeIndirectObject(wArr)

	// Use identity character id (CID) to glyph id (GID) mapping.
	cidfont.CIDToGIDMap = core.MakeName("Identity")

	d := core.MakeDict()
	d.Set("Ordering", core.MakeString("Identity"))
	d.Set("Registry", core.MakeString("Adobe"))
	d.Set("Supplement", core.MakeInteger(0))
	cidfont.CIDSystemInfo = d

	// Make the font descriptor.
	descriptor := &PdfFontDescriptor{}
	descriptor.Ascent = core.MakeFloat(k * float64(ttf.TypoAscender))
	descriptor.Descent = core.MakeFloat(k * float64(ttf.TypoDescender))
	descriptor.CapHeight = core.MakeFloat(k * float64(ttf.CapHeight))
	descriptor.FontBBox = core.MakeArrayFromFloats([]float64{k * float64(ttf.Xmin),
		k * float64(ttf.Ymin), k * float64(ttf.Xmax), k * float64(ttf.Ymax)})
	descriptor.ItalicAngle = core.MakeFloat(float64(ttf.ItalicAngle))
	descriptor.MissingWidth = core.MakeFloat(k * float64(ttf.Widths[0]))

	// Embed the TrueType font program.
	ttfBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		common.Log.Debug("ERROR: :Unable to read file contents: %v", err)
		return nil, err
	}

	stream, err := core.MakeStream(ttfBytes, core.NewFlateEncoder())
	if err != nil {
		common.Log.Debug("ERROR: Unable to make stream: %v", err)
		return nil, err
	}
	stream.PdfObjectDictionary.Set("Length1", core.MakeInteger(int64(len(ttfBytes))))
	descriptor.FontFile2 = stream

	if ttf.Bold {
		descriptor.StemV = core.MakeInteger(120)
	} else {
		descriptor.StemV = core.MakeInteger(70)
	}

	// Flags
	flags := fontFlagSymbolic // Symbolic.
	if ttf.IsFixedPitch {
		flags |= fontFlagFixedPitch
	}
	if ttf.ItalicAngle != 0 {
		flags |= fontFlagItalic
	}
	descriptor.Flags = core.MakeInteger(int64(flags))

	base.fontDescriptor = descriptor
	descendantFont := PdfFont{
		context: cidfont,
	}

	// Make root Type0 font.
	type0 := pdfFontType0{
		fontDescriptor: descriptor,
		DescendantFont: &descendantFont,
		Encoding:       core.MakeName("Identity-H"),
		encoder:        textencoding.NewTrueTypeFontEncoder(ttf.Chars),
	}

	// Build Font.
	font := PdfFont{
		context: &type0,
	}

	return &font, nil
}
