/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/textencoding"
	"github.com/unidoc/unidoc/pdf/model/fonts"
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
	fontCommon
	container *core.PdfIndirectObject

	// These fields are specific to Type 0 fonts.
	encoder        textencoding.TextEncoder
	Encoding       core.PdfObject
	DescendantFont *PdfFont // Can be either CIDFontType0 or CIDFontType2 font.
}

// pdfFontType0FromSkeleton returns a pdfFontType0 with its common fields initalized.
func pdfFontType0FromSkeleton(base *fontCommon) *pdfFontType0 {
	return &pdfFontType0{
		fontCommon: *base,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfFontType0) baseFields() *fontCommon {
	return &font.fontCommon
}

// GetGlyphCharMetrics returns the character metrics for the specified glyph.  A bool flag is
// returned to indicate whether or not the entry was found in the glyph to charcode mapping.
func (font pdfFontType0) GetGlyphCharMetrics(glyph textencoding.GlyphName) (fonts.CharMetrics, bool) {
	if font.DescendantFont == nil {
		common.Log.Debug("ERROR: No descendant. font=%s", font)
		return fonts.CharMetrics{}, false
	}
	return font.DescendantFont.GetGlyphCharMetrics(glyph)
}

// GetCharMetrics returns the char metrics for character code `code`.
func (font pdfFontType0) GetCharMetrics(code uint16) (fonts.CharMetrics, bool) {
	if font.DescendantFont == nil {
		common.Log.Debug("ERROR: No descendant. font=%s", font)
		return fonts.CharMetrics{}, false
	}
	return font.DescendantFont.GetCharMetrics(code)
}

// Encoder returns the font's text encoder.
func (font pdfFontType0) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// SetEncoder sets the encoder for the font.
func (font pdfFontType0) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

// ToPdfObject converts the font to a PDF representation.
func (font *pdfFontType0) ToPdfObject() core.PdfObject {
	if font.container == nil {
		font.container = &core.PdfIndirectObject{}
	}
	d := font.baseFields().asPdfObjectDictionary("Type0")

	font.container.PdfObject = d

	if font.Encoding != nil {
		d.Set("Encoding", font.Encoding)
	} else if font.encoder != nil {
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
	arr, ok := core.GetArray(d.Get("DescendantFonts"))
	if !ok {
		common.Log.Debug("ERROR: Invalid DescendantFonts - not an array %s", base)
		return nil, core.ErrRangeError
	}
	if arr.Len() != 1 {
		common.Log.Debug("ERROR: Array length != 1 (%d)", arr.Len())
		return nil, core.ErrRangeError
	}
	df, err := newPdfFontFromPdfObject(arr.Get(0), false)
	if err != nil {
		common.Log.Debug("ERROR: Failed loading descendant font: err=%v %s", err, base)
		return nil, err
	}

	font := pdfFontType0FromSkeleton(base)
	font.DescendantFont = df

	encoderName, ok := core.GetNameVal(d.Get("Encoding"))
	if ok {
		if encoderName == "Identity-H" || encoderName == "Identity-V" {
			font.encoder = textencoding.NewIdentityTextEncoder(encoderName)
		} else {
			common.Log.Debug("Unhandled cmap %q", encoderName)
		}
	}
	return font, nil
}

// pdfCIDFontType0 represents a CIDFont Type0 font dictionary.
type pdfCIDFontType0 struct {
	fontCommon
	container *core.PdfIndirectObject

	// These fields are specific to Type 0 fonts.
	encoder textencoding.TextEncoder

	// Table 117 – Entries in a CIDFont dictionary (page 269)
	CIDSystemInfo *core.PdfObjectDictionary // (Required) Dictionary that defines the character
	// collection of the CIDFont. See Table 116.
}

// pdfCIDFontType0FromSkeleton returns a pdfCIDFontType0 with its common fields initalized.
func pdfCIDFontType0FromSkeleton(base *fontCommon) *pdfCIDFontType0 {
	return &pdfCIDFontType0{
		fontCommon: *base,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfCIDFontType0) baseFields() *fontCommon {
	return &font.fontCommon
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
func (font pdfCIDFontType0) GetGlyphCharMetrics(glyph textencoding.GlyphName) (fonts.CharMetrics, bool) {
	return fonts.CharMetrics{}, true
}

// GetCharMetrics returns the char metrics for character code `code`.
func (font pdfCIDFontType0) GetCharMetrics(code uint16) (fonts.CharMetrics, bool) {
	return fonts.CharMetrics{}, true
}

// ToPdfObject converts the pdfCIDFontType0 to a PDF representation.
func (font *pdfCIDFontType0) ToPdfObject() core.PdfObject {
	return core.MakeNull()
}

// newPdfCIDFontType0FromPdfObject creates a pdfCIDFontType0 object from a dictionary (either direct
// or via indirect object). If a problem occurs with loading an error is returned.
func newPdfCIDFontType0FromPdfObject(d *core.PdfObjectDictionary, base *fontCommon) (*pdfCIDFontType0, error) {
	if base.subtype != "CIDFontType0" {
		common.Log.Debug("ERROR: Font SubType != CIDFontType0. font=%s", base)
		return nil, core.ErrRangeError
	}

	font := pdfCIDFontType0FromSkeleton(base)

	// CIDSystemInfo.
	obj, ok := core.GetDict(d.Get("CIDSystemInfo"))
	if !ok {
		common.Log.Debug("ERROR: CIDSystemInfo (Required) missing. font=%s", base)
		return nil, ErrRequiredAttributeMissing
	}
	font.CIDSystemInfo = obj

	return font, nil
}

// pdfCIDFontType2 represents a CIDFont Type2 font dictionary.
type pdfCIDFontType2 struct {
	fontCommon
	container *core.PdfIndirectObject

	// These fields are specific to Type 0 fonts.
	encoder   textencoding.TextEncoder
	ttfParser *fonts.TtfType

	CIDSystemInfo *core.PdfObjectDictionary
	DW            core.PdfObject
	W             core.PdfObject
	DW2           core.PdfObject
	W2            core.PdfObject
	CIDToGIDMap   core.PdfObject

	widths       map[int]float64
	defaultWidth float64

	// Mapping between unicode runes to widths.
	runeToWidthMap map[rune]int

	// Also mapping between GIDs (glyph index) and width.
	gidToWidthMap map[fonts.GID]int

	// Cache for glyph to metrics.
	glyphToMetricsCache map[string]fonts.CharMetrics
}

// pdfCIDFontType2FromSkeleton returns a pdfCIDFontType2 with its common fields initalized.
func pdfCIDFontType2FromSkeleton(base *fontCommon) *pdfCIDFontType2 {
	return &pdfCIDFontType2{
		fontCommon: *base,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfCIDFontType2) baseFields() *fontCommon {
	return &font.fontCommon
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
func (font pdfCIDFontType2) GetGlyphCharMetrics(glyph textencoding.GlyphName) (fonts.CharMetrics, bool) {
	// Return cached value if cached.
	if font.glyphToMetricsCache == nil {
		font.glyphToMetricsCache = make(map[string]fonts.CharMetrics)
	}
	if metrics, cached := font.glyphToMetricsCache[glyph]; cached {
		return metrics, true
	}
	metrics := fonts.CharMetrics{}

	if font.ttfParser == nil {
		return metrics, false
	}

	enc := textencoding.NewTrueTypeFontEncoder(font.ttfParser.Chars)

	// Convert the glyph to character code.
	r, found := enc.GlyphToRune(glyph)
	if !found {
		common.Log.Debug("Unable to convert glyph %q to charcode (identity)", glyph)
		return metrics, false
	}

	w, found := font.runeToWidthMap[r]
	if !found {
		dw, ok := core.GetInt(font.DW)
		if !ok {
			return metrics, false
		}
		w = int(*dw)
	}
	metrics.GlyphName = glyph
	metrics.Wx = float64(w)

	font.glyphToMetricsCache[glyph] = metrics

	return metrics, true
}

// GetCharMetrics returns the char metrics for character code `code`.
func (font pdfCIDFontType2) GetCharMetrics(code uint16) (fonts.CharMetrics, bool) {
	if w, ok := font.widths[int(code)]; ok {
		return fonts.CharMetrics{Wx: float64(w)}, true
	}
	// TODO(peterwilliams97): The remainder of this function is pure guesswork. Explain it.
	// FIXME(gunnsth): Appears that we are assuming a code <-> rune identity mapping. Related PR #259 should solve this.
	w, ok := font.runeToWidthMap[code]
	if !ok {
		w = int(font.defaultWidth)
	}
	return fonts.CharMetrics{Wx: float64(w)}, true
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
	obj, ok := core.GetDict(d.Get("CIDSystemInfo"))
	if !ok {
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

	if arr2, ok := core.GetArray(font.W); ok {
		font.widths = map[int]float64{}
		for i := 0; i < arr2.Len()-1; i++ {
			obj0 := (*arr2).Get(i)
			n, ok0 := core.GetIntVal(obj0)
			if !ok0 {
				return nil, fmt.Errorf("Bad font W obj0: i=%d %#v", i, obj0)
			}
			i++
			if i > arr2.Len()-1 {
				return nil, fmt.Errorf("Bad font W array: arr2=%+v", arr2)
			}
			obj1 := (*arr2).Get(i)
			switch obj1.(type) {
			case *core.PdfObjectArray:
				arr, _ := core.GetArray(obj1)
				if widths, err := arr.ToFloat64Array(); err == nil {
					for j := 0; j < len(widths); j++ {
						font.widths[n+j] = widths[j]
					}
				} else {
					return nil, fmt.Errorf("Bad font W array obj1: i=%d %#v", i, obj1)
				}
			case *core.PdfObjectInteger:
				n1, ok1 := core.GetIntVal(obj1)
				if !ok1 {
					return nil, fmt.Errorf("Bad font W int obj1: i=%d %#v", i, obj1)
				}
				i++
				if i > arr2.Len()-1 {
					return nil, fmt.Errorf("Bad font W array: arr2=%+v", arr2)
				}
				obj2 := (*arr2).Get(i)
				v, err := core.GetNumberAsFloat(obj2)
				if err != nil {
					return nil, fmt.Errorf("Bad font W int obj2: i=%d %#v", i, obj2)
				}
				for j := n; j <= n1; j++ {
					font.widths[j] = v
				}
			default:
				return nil, fmt.Errorf("Bad font W obj1 type: i=%d %#v", i, obj1)
			}
		}
	}
	if defaultWidth, err := core.GetNumberAsFloat(font.DW); err == nil {
		font.defaultWidth = defaultWidth
	} else {
		font.defaultWidth = 1000.0
	}

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
	cidfont := &pdfCIDFontType2{
		fontCommon: fontCommon{
			subtype: "CIDFontType2",
		},
	}
	cidfont.ttfParser = &ttf

	// 2-byte character codes ➞ runes
	runes := make([]rune, 0, len(ttf.Chars))
	for r := range ttf.Chars {
		runes = append(runes, rune(r))
	}
	// make sure runes are sorted so PDF output is stable
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	k := 1000.0 / float64(ttf.UnitsPerEm)

	if len(ttf.Widths) <= 0 {
		return nil, errors.New("ERROR: Missing required attribute (Widths)")
	}

	missingWidth := k * float64(ttf.Widths[0])

	// Construct a rune ➞ width map.
	runeToWidthMap := make(map[rune]int)
	gidToWidthMap := map[fonts.GID]int{}
	for _, r := range runes {
		gid := ttf.Chars[r]

		w := k * float64(ttf.Widths[gid])
		runeToWidthMap[r] = int(w)
		gidToWidthMap[gid] = int(w)
	}
	cidfont.runeToWidthMap = runeToWidthMap
	cidfont.gidToWidthMap = gidToWidthMap

	// Default width.
	cidfont.DW = core.MakeInteger(int64(missingWidth))

	// Construct W array.  Stores character code to width mappings.
	wArr := &core.PdfObjectArray{}

	for i := uint16(0); int(i) < len(runes); {

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
	descriptor := &PdfFontDescriptor{
		FontName:  core.MakeName(ttf.PostScriptName),
		Ascent:    core.MakeFloat(k * float64(ttf.TypoAscender)),
		Descent:   core.MakeFloat(k * float64(ttf.TypoDescender)),
		CapHeight: core.MakeFloat(k * float64(ttf.CapHeight)),
		FontBBox: core.MakeArrayFromFloats([]float64{
			k * float64(ttf.Xmin),
			k * float64(ttf.Ymin),
			k * float64(ttf.Xmax),
			k * float64(ttf.Ymax),
		}),
		ItalicAngle:  core.MakeFloat(float64(ttf.ItalicAngle)),
		MissingWidth: core.MakeFloat(k * float64(ttf.Widths[0])),
	}

	// Embed the TrueType font program.
	ttfBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		common.Log.Debug("ERROR: Unable to read file contents: %v", err)
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

	cidfont.basefont = ttf.PostScriptName
	cidfont.fontDescriptor = descriptor

	// Make root Type0 font.
	type0 := pdfFontType0{
		fontCommon: fontCommon{
			subtype:  "Type0",
			basefont: ttf.PostScriptName,
		},
		DescendantFont: &PdfFont{
			context: cidfont,
		},
		Encoding: core.MakeName("Identity-H"),
		encoder:  textencoding.NewTrueTypeFontEncoder(ttf.Chars),
	}

	type0.toUnicodeCmap = ttf.MakeToUnicode()

	// Build Font.
	font := PdfFont{
		context: &type0,
	}

	return &font, nil
}
