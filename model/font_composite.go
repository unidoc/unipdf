/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/cmap"
	"github.com/unidoc/unipdf/v3/internal/textencoding"
	"github.com/unidoc/unipdf/v3/model/internal/fonts"
	"github.com/unidoc/unitype"
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

// pdfFontType0 implements pdfFont
var _ pdfFont = (*pdfFontType0)(nil)

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
	codeToCID      *cmap.CMap
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

func (font *pdfFontType0) getFontDescriptor() *PdfFontDescriptor {
	if font.fontDescriptor == nil && font.DescendantFont != nil {
		return font.DescendantFont.FontDescriptor()
	}
	return font.fontDescriptor
}

// GetRuneMetrics returns the character metrics for the specified rune.
// A bool flag is returned to indicate whether or not the entry was found.
func (font pdfFontType0) GetRuneMetrics(r rune) (fonts.CharMetrics, bool) {
	if font.DescendantFont == nil {
		common.Log.Debug("ERROR: No descendant. font=%s", font)
		return fonts.CharMetrics{}, false
	}
	return font.DescendantFont.GetRuneMetrics(r)
}

// GetCharMetrics returns the char metrics for character code `code`.
func (font pdfFontType0) GetCharMetrics(code textencoding.CharCode) (fonts.CharMetrics, bool) {
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

// bytesToCharcodes attempts to convert the specified byte slice to charcodes,
// based on the font's charcode to CID CMap.
func (font *pdfFontType0) bytesToCharcodes(data []byte) ([]textencoding.CharCode, bool) {
	if font.codeToCID == nil {
		return nil, false
	}

	codes, ok := font.codeToCID.BytesToCharcodes(data)
	if !ok {
		return nil, false
	}

	charcodes := make([]textencoding.CharCode, len(codes))
	for i, code := range codes {
		charcodes[i] = textencoding.CharCode(code)
	}

	return charcodes, true
}

// Subset name is `tag+name`.
func makeSubsetName(name, tag string) string {
	if strings.Contains(name, "+") {
		parts := strings.Split(name, "+")
		if len(parts) == 2 {
			name = parts[1]
		}
	}
	return tag + "+" + name
}

// Generates tag for subsetting with 6 random uppercase letters.
func genSubsetTag() string {
	letters := "QWERTYUIOPASDFGHJKLZXCVBNM"
	var buf bytes.Buffer
	for i := 0; i < 6; i++ {
		buf.WriteRune(rune(letters[rand.Intn(len(letters))]))
	}
	return buf.String()
}

// subsetRegistered subsets the `font` to only the glyphs that have been registered by encoder.
// NOTE: Only works for CIDFontType2 (TrueType CID font), no-op otherwise.
func (font *pdfFontType0) subsetRegistered() error {
	cidfnt, ok := font.DescendantFont.context.(*pdfCIDFontType2)
	if !ok {
		common.Log.Debug("Font not supported for subsetting %T", font.DescendantFont)
		return nil
	}
	if cidfnt == nil {
		return nil
	}
	if cidfnt.fontDescriptor == nil {
		common.Log.Debug("Missing font descriptor")
		return nil
	}
	if font.encoder == nil {
		common.Log.Debug("No encoder - subsetting ignored")
		return nil
	}

	stream, ok := core.GetStream(cidfnt.fontDescriptor.FontFile2)
	if !ok {
		common.Log.Debug("Embedded font object not found -- ABORT subsetting")
		return errors.New("fontfile2 not found")
	}
	decoded, err := core.DecodeStream(stream)
	if err != nil {
		common.Log.Debug("Decode error: %v", err)
		return err
	}

	fnt, err := unitype.Parse(bytes.NewReader(decoded))
	if err != nil {
		common.Log.Debug("Error parsing %d byte font", len(stream.Stream))
		return err
	}

	var runes []rune
	var subset *unitype.Font
	switch tenc := font.encoder.(type) {
	case *textencoding.TrueTypeFontEncoder:
		// Means the font has been loaded from TTF file.
		runes = tenc.RegisteredRunes()
		subset, err = fnt.SubsetKeepRunes(runes)
		if err != nil {
			common.Log.Debug("ERROR: %v", err)
			return err
		}
		// Reduce the encoder also.
		tenc.SubsetRegistered()
	case *textencoding.IdentityEncoder:
		// IdentityEncoder typically means font was parsed from PDF file.
		// TODO: These are not actual runes... but glyph ids ? Very confusing.
		runes = tenc.RegisteredRunes()
		indices := make([]unitype.GlyphIndex, len(runes))
		for i, r := range runes {
			indices[i] = unitype.GlyphIndex(r)
		}

		subset, err = fnt.SubsetKeepIndices(indices)
		if err != nil {
			common.Log.Debug("ERROR: %v", err)
			return err
		}
	case textencoding.SimpleEncoder:
		// Simple encoding, bytes are 0-255
		charcodes := tenc.Charcodes()
		for _, c := range charcodes {
			r, ok := tenc.CharcodeToRune(c)
			if !ok {
				common.Log.Debug("ERROR: unable convert charcode to rune: %d", c)
				continue
			}
			runes = append(runes, r)
		}
	default:
		return fmt.Errorf("unsupported encoder for subsetting: %T", font.encoder)
	}

	var buf bytes.Buffer
	err = subset.Write(&buf)
	if err != nil {
		common.Log.Debug("ERROR: %v", err)
		return err
	}

	// Update info for ToUnicode CMap entry.
	if font.toUnicodeCmap != nil {
		codeToUnicode := make(map[cmap.CharCode]rune, len(runes))
		for _, r := range runes {
			cc, ok := font.encoder.RuneToCharcode(r)
			if !ok {
				continue
			}
			codeToUnicode[cmap.CharCode(cc)] = r
		}
		font.toUnicodeCmap = cmap.NewToUnicodeCMap(codeToUnicode)
	}

	stream, err = core.MakeStream(buf.Bytes(), core.NewFlateEncoder())
	if err != nil {
		common.Log.Debug("ERROR: %v", err)
		return err
	}
	stream.Set("Length1", core.MakeInteger(int64(buf.Len())))
	if curstr, ok := core.GetStream(cidfnt.fontDescriptor.FontFile2); ok {
		// Replace the current stream (keep same object).
		*curstr = *stream
	} else {
		cidfnt.fontDescriptor.FontFile2 = stream
	}

	// Set subset name.
	tag := genSubsetTag()

	if len(font.basefont) > 0 {
		font.basefont = makeSubsetName(font.basefont, tag)
	}
	if len(cidfnt.basefont) > 0 {
		cidfnt.basefont = makeSubsetName(cidfnt.basefont, tag)
	}
	if len(font.name) > 0 {
		font.name = makeSubsetName(font.name, tag)
	}
	if cidfnt.fontDescriptor != nil {
		fname, ok := core.GetName(cidfnt.fontDescriptor.FontName)
		if ok && len(fname.String()) > 0 {
			fname := makeSubsetName(fname.String(), tag)
			cidfnt.fontDescriptor.FontName = core.MakeName(fname)
		}
	}

	return nil
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
		// TODO: Identity-H maps 16-bit character codes straight to glyph index (don't need actual runes).
		if encoderName == "Identity-H" || encoderName == "Identity-V" {
			font.encoder = textencoding.NewIdentityTextEncoder(encoderName)
		} else if cmap.IsPredefinedCMap(encoderName) {
			font.codeToCID, err = cmap.LoadPredefinedCMap(encoderName)
			if err != nil {
				common.Log.Debug("WARN: could not load predefined CMap %s: %v", encoderName, err)
			}
		} else {
			common.Log.Debug("Unhandled cmap %q", encoderName)
		}
	}

	if cidToUnicode := df.baseFields().toUnicodeCmap; cidToUnicode != nil {
		if dfn := cidToUnicode.Name(); dfn == "Adobe-CNS1-UCS2" || dfn == "Adobe-GB1-UCS2" ||
			dfn == "Adobe-Japan1-UCS2" || dfn == "Adobe-Korea1-UCS2" {
			font.encoder = textencoding.NewCMapEncoder(encoderName, font.codeToCID, cidToUnicode)
		}
	}

	return font, nil
}

// pdfCIDFontType0 implements pdfFont
var _ pdfFont = (*pdfCIDFontType0)(nil)

// pdfCIDFontType0 represents a CIDFont Type0 font dictionary.
type pdfCIDFontType0 struct {
	fontCommon
	container *core.PdfIndirectObject

	// These fields are specific to Type 0 fonts.
	encoder textencoding.TextEncoder

	// Table 117 – Entries in a CIDFont dictionary (page 269)
	// (Required) Dictionary that defines the character collection of the CIDFont.
	// See Table 116.
	CIDSystemInfo *core.PdfObjectDictionary

	// Glyph metrics fields (optional).
	DW  core.PdfObject // default glyph width.
	W   core.PdfObject // glyph widths array.
	DW2 core.PdfObject // default glyph metrics for CID fonts used for vertical writing.
	W2  core.PdfObject // glyph metrics for CID fonts used for vertical writing.

	widths       map[textencoding.CharCode]float64
	defaultWidth float64
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

func (font *pdfCIDFontType0) getFontDescriptor() *PdfFontDescriptor {
	return font.fontDescriptor
}

// Encoder returns the font's text encoder.
func (font pdfCIDFontType0) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// GetRuneMetrics returns the character metrics for the specified rune.
// A bool flag is returned to indicate whether or not the entry was found.
func (font pdfCIDFontType0) GetRuneMetrics(r rune) (fonts.CharMetrics, bool) {
	return fonts.CharMetrics{Wx: font.defaultWidth}, true
}

// GetCharMetrics returns the char metrics for character code `code`.
func (font pdfCIDFontType0) GetCharMetrics(code textencoding.CharCode) (fonts.CharMetrics, bool) {
	width := font.defaultWidth
	if w, ok := font.widths[code]; ok {
		width = w
	}

	return fonts.CharMetrics{Wx: width}, true
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

	// Optional attributes.
	font.DW = d.Get("DW")
	font.W = d.Get("W")
	font.DW2 = d.Get("DW2")
	font.W2 = d.Get("W2")

	// Get font default glyph width.
	font.defaultWidth = 1000.0
	if dw, err := core.GetNumberAsFloat(font.DW); err == nil {
		font.defaultWidth = dw
	}

	// Parse glyph widths array, if one is present.
	fontWidths, err := parseCIDFontWidthsArray(font.W)
	if err != nil {
		return nil, err
	}
	if fontWidths == nil {
		fontWidths = map[textencoding.CharCode]float64{}
	}
	font.widths = fontWidths

	return font, nil
}

// pdfCIDFontType2 implements pdfFont
var _ pdfFont = (*pdfCIDFontType2)(nil)

// pdfCIDFontType2 represents a CIDFont Type2 font dictionary.
type pdfCIDFontType2 struct {
	fontCommon
	container *core.PdfIndirectObject

	// These fields are specific to Type 0 fonts.
	encoder textencoding.TextEncoder

	// Table 117 – Entries in a CIDFont dictionary (page 269)
	// Dictionary that defines the character collection of the CIDFont (required).
	// See Table 116.
	CIDSystemInfo *core.PdfObjectDictionary

	// Glyph metrics fields (optional).
	DW  core.PdfObject // default glyph width.
	W   core.PdfObject // glyph widths array.
	DW2 core.PdfObject // default glyph metrics for CID fonts used for vertical writing.
	W2  core.PdfObject // glyph metrics for CID fonts used for vertical writing.

	// CIDs to glyph indices mapping (optional).
	CIDToGIDMap core.PdfObject

	widths       map[textencoding.CharCode]float64
	defaultWidth float64

	// Mapping between unicode runes to widths.
	// TODO(dennwc): it is used only in GetGlyphCharMetrics
	//  			 we can precompute metrics and drop it
	runeToWidthMap map[rune]int
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

func (font *pdfCIDFontType2) getFontDescriptor() *PdfFontDescriptor {
	return font.fontDescriptor
}

// Encoder returns the font's text encoder.
func (font pdfCIDFontType2) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// GetRuneMetrics returns the character metrics for the specified rune.
// A bool flag is returned to indicate whether or not the entry was found.
func (font pdfCIDFontType2) GetRuneMetrics(r rune) (fonts.CharMetrics, bool) {
	w, found := font.runeToWidthMap[r]
	if !found {
		dw, ok := core.GetInt(font.DW)
		if !ok {
			return fonts.CharMetrics{}, false
		}
		w = int(*dw)
	}
	return fonts.CharMetrics{Wx: float64(w)}, true
}

// GetCharMetrics returns the char metrics for character code `code`.
func (font pdfCIDFontType2) GetCharMetrics(code textencoding.CharCode) (fonts.CharMetrics, bool) {
	if w, ok := font.widths[code]; ok {
		return fonts.CharMetrics{Wx: float64(w)}, true
	}
	// TODO(peterwilliams97): The remainder of this function is pure guesswork. Explain it.
	// FIXME(gunnsth): Appears that we are assuming a code <-> rune identity mapping.
	r := rune(code)
	w, ok := font.runeToWidthMap[r]
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

	// Get font default glyph width.
	font.defaultWidth = 1000.0
	if dw, err := core.GetNumberAsFloat(font.DW); err == nil {
		font.defaultWidth = dw
	}

	// Parse glyph widths array, if one is present.
	fontWidths, err := parseCIDFontWidthsArray(font.W)
	if err != nil {
		return nil, err
	}
	if fontWidths == nil {
		fontWidths = map[textencoding.CharCode]float64{}
	}
	font.widths = fontWidths

	return font, nil
}

func parseCIDFontWidthsArray(w core.PdfObject) (map[textencoding.CharCode]float64, error) {
	if w == nil {
		return nil, nil
	}

	wArr, ok := core.GetArray(w)
	if !ok {
		return nil, nil
	}

	fontWidths := map[textencoding.CharCode]float64{}
	wArrLen := wArr.Len()
	for i := 0; i < wArrLen-1; i++ {
		obj0 := core.TraceToDirectObject(wArr.Get(i))
		n, ok0 := core.GetIntVal(obj0)
		if !ok0 {
			return nil, fmt.Errorf("Bad font W obj0: i=%d %#v", i, obj0)
		}
		i++
		if i > wArrLen-1 {
			return nil, fmt.Errorf("Bad font W array: arr2=%+v", wArr)
		}

		obj1 := core.TraceToDirectObject(wArr.Get(i))
		switch obj1.(type) {
		case *core.PdfObjectArray:
			arr, _ := core.GetArray(obj1)
			if widths, err := arr.ToFloat64Array(); err == nil {
				for j := 0; j < len(widths); j++ {
					fontWidths[textencoding.CharCode(n+j)] = widths[j]
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
			if i > wArrLen-1 {
				return nil, fmt.Errorf("Bad font W array: arr2=%+v", wArr)
			}

			obj2 := wArr.Get(i)
			v, err := core.GetNumberAsFloat(obj2)
			if err != nil {
				return nil, fmt.Errorf("Bad font W int obj2: i=%d %#v", i, obj2)
			}

			for j := n; j <= n1; j++ {
				fontWidths[textencoding.CharCode(j)] = v
			}
		default:
			return nil, fmt.Errorf("Bad font W obj1 type: i=%d %#v", i, obj1)
		}
	}

	return fontWidths, nil
}

// NewCompositePdfFontFromTTFFile loads a composite font from a TTF font file. Composite fonts can
// be used to represent unicode fonts which can have multi-byte character codes, representing a wide
// range of values. They are often used for symbolic languages, including Chinese, Japanese and Korean.
// It is represented by a Type0 Font with an underlying CIDFontType2 and an Identity-H encoding map.
// TODO: May be extended in the future to support a larger variety of CMaps and vertical fonts.
// NOTE: For simple fonts, use NewPdfFontFromTTFFile.
func NewCompositePdfFontFromTTFFile(filePath string) (*PdfFont, error) {
	f, err := os.Open(filePath)
	if err != nil {
		common.Log.Debug("ERROR: opening file: %v", err)
		return nil, err
	}
	defer f.Close()
	return NewCompositePdfFontFromTTF(f)
}

// NewCompositePdfFontFromTTF loads a composite TTF font. Composite fonts can
// be used to represent unicode fonts which can have multi-byte character codes, representing a wide
// range of values. They are often used for symbolic languages, including Chinese, Japanese and Korean.
// It is represented by a Type0 Font with an underlying CIDFontType2 and an Identity-H encoding map.
// TODO: May be extended in the future to support a larger variety of CMaps and vertical fonts.
// NOTE: For simple fonts, use NewPdfFontFromTTF.
func NewCompositePdfFontFromTTF(r io.ReadSeeker) (*PdfFont, error) {
	ttfBytes, err := ioutil.ReadAll(r)
	if err != nil {
		common.Log.Debug("ERROR: Unable to read font contents: %v", err)
		return nil, err
	}

	ttf, err := fonts.TtfParse(bytes.NewReader(ttfBytes))
	if err != nil {
		common.Log.Debug("ERROR: while loading ttf font: %v", err)
		return nil, err
	}

	// Prepare the inner descendant font (CIDFontType2).
	cidfont := &pdfCIDFontType2{
		fontCommon: fontCommon{
			subtype: "CIDFontType2",
		},

		// Use identity character id (CID) to glyph id (GID) mapping.
		// Code below relies on the fact that identity mapping is used.
		CIDToGIDMap: core.MakeName("Identity"),
	}

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
	for _, r := range runes {
		gid := ttf.Chars[r]

		w := k * float64(ttf.Widths[gid])
		runeToWidthMap[r] = int(w)
	}
	cidfont.runeToWidthMap = runeToWidthMap

	// Default width.
	cidfont.DW = core.MakeInteger(int64(missingWidth))

	// Construct W array.  Stores character code to width mappings.
	wArr := makeCIDWidthArr(runes, runeToWidthMap, ttf.Chars)
	cidfont.W = core.MakeIndirectObject(wArr)

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
		encoder:  ttf.NewEncoder(),
	}

	// Generate CMap for the Type 0 font, which is the inverse of ttf.Chars.
	if len(ttf.Chars) > 0 {
		codeToUnicode := make(map[cmap.CharCode]rune, len(ttf.Chars))
		for r, gid := range ttf.Chars {
			cid := cmap.CharCode(gid)
			if rn, ok := codeToUnicode[cid]; !ok || (ok && rn > r) {
				codeToUnicode[cid] = r
			}
		}
		type0.toUnicodeCmap = cmap.NewToUnicodeCMap(codeToUnicode)
	}

	// Build Font.
	font := PdfFont{
		context: &type0,
	}

	return &font, nil
}

func makeCIDWidthArr(runes []rune, widths map[rune]int, gids map[rune]fonts.GID) *core.PdfObjectArray {
	// Construct W array. Stores character code to width mappings.
	arr := &core.PdfObjectArray{}

	// 9.7.4.3 Glyph metrics in CIDFonts

	// In the first format, c shall be an integer specifying a starting CID value; it shall be followed by an array of
	// n numbers that shall specify the widths for n consecutive CIDs, starting with c.
	// The second format shall define the same width, w, as a number, for all CIDs in the range c_first to c_last.

	// We always use the second format.

	// TODO(dennwc): this can be done on GIDs instead of runes
	for i := 0; i < len(runes); {
		w := widths[runes[i]]

		li := i
		for j := i + 1; j < len(runes); j++ {
			w2 := widths[runes[j]]
			if w == w2 {
				li = j
			} else {
				break
			}
		}

		// The W maps from CID to width, here CID = GID.
		gid1 := gids[runes[i]]
		gid2 := gids[runes[li]]

		arr.Append(core.MakeInteger(int64(gid1)))
		arr.Append(core.MakeInteger(int64(gid2)))
		arr.Append(core.MakeInteger(int64(w)))

		i = li + 1
	}
	return arr
}
