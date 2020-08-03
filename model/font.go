/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"

	"github.com/unidoc/unipdf/v3/internal/cmap"
	"github.com/unidoc/unipdf/v3/internal/textencoding"
	"github.com/unidoc/unipdf/v3/model/internal/fonts"
)

// pdfFont is an internal interface for fonts that can be stored in PDF documents.
type pdfFont interface {
	fonts.Font
	// ToPdfObject returns a PDF representation of the font and implements interface Model.
	ToPdfObject() core.PdfObject
	// getFontDescriptor returns the font descriptor of the font.
	getFontDescriptor() *PdfFontDescriptor
	// baseFields returns fields that are common for PDF fonts.
	baseFields() *fontCommon
}

// PdfFont represents an underlying font structure which can be of type:
// - Type0
// - Type1
// - TrueType
// etc.
type PdfFont struct {
	context pdfFont // The underlying font: Type0, Type1, Truetype, etc..
}

// SubsetRegistered subsets the font to only the glyphs that have been registered by the encoder.
// NOTE: This only works on fonts that support subsetting. For unsupported fonts this is a no-op, although a debug
//   message is emitted.  Currently supported fonts are embedded Truetype CID fonts (type 0).
// NOTE: Make sure to call this soon before writing (once all needed runes have been registered).
// If using package creator, use its EnableFontSubsetting method instead.
func (font *PdfFont) SubsetRegistered() error {
	switch t := font.context.(type) {
	case *pdfFontType0:
		err := t.subsetRegistered()
		if err != nil {
			common.Log.Debug("Subset error: %v", err)
			return err
		}
		if t.container != nil {
			if t.encoder != nil {
				t.encoder.ToPdfObject() // Forced update of encoder object.
			}
			t.ToPdfObject() // Forced update of object.
		}
	default:
		common.Log.Debug("Font %T does not support subsetting", t)
	}
	return nil
}

// GetFontDescriptor returns the font descriptor for `font`.
func (font PdfFont) GetFontDescriptor() (*PdfFontDescriptor, error) {
	return font.context.getFontDescriptor(), nil
}

// String returns a string that describes `font`.
func (font *PdfFont) String() string {
	enc := ""
	if font.context.Encoder() != nil {
		enc = font.context.Encoder().String()
	}
	return fmt.Sprintf("FONT{%T %s %s}", font.context, font.baseFields().coreString(), enc)
}

// BaseFont returns the font's "BaseFont" field.
func (font *PdfFont) BaseFont() string {
	return font.baseFields().basefont
}

// Subtype returns the font's "Subtype" field.
func (font *PdfFont) Subtype() string {
	subtype := font.baseFields().subtype
	if t, ok := font.context.(*pdfFontType0); ok {
		subtype = subtype + ":" + t.DescendantFont.Subtype()
	}
	return subtype
}

// IsCID returns true if the underlying font is CID.
func (font *PdfFont) IsCID() bool {
	return font.baseFields().isCIDFont()
}

// FontDescriptor returns font's PdfFontDescriptor. This may be a builtin descriptor for standard 14
// fonts but must be an explicit descriptor for other fonts.
func (font *PdfFont) FontDescriptor() *PdfFontDescriptor {
	if font.baseFields().fontDescriptor != nil {
		return font.baseFields().fontDescriptor
	}
	if d := font.context.getFontDescriptor(); d != nil {
		return d
	}
	common.Log.Error("All fonts have a Descriptor. font=%s", font)
	return nil
}

// ToUnicode returns the name of the font's "ToUnicode" field if there is one, or "" if there isn't.
func (font *PdfFont) ToUnicode() string {
	if font.baseFields().toUnicodeCmap == nil {
		return ""
	}
	return font.baseFields().toUnicodeCmap.Name()
}

// DefaultFont returns the default font, which is currently the built in Helvetica.
func DefaultFont() *PdfFont {
	helvetica, ok := fonts.NewStdFontByName(HelveticaName)
	if !ok {
		panic("Helvetica should always be available")
	}
	std := stdFontToSimpleFont(helvetica)
	return &PdfFont{context: &std}
}

func newStandard14Font(basefont StdFontName) (pdfFontSimple, error) {
	fnt, ok := fonts.NewStdFontByName(basefont)
	if !ok {
		return pdfFontSimple{}, ErrFontNotSupported
	}
	std := stdFontToSimpleFont(fnt)
	return std, nil
}

// StdFontName represents name of a standard font.
type StdFontName = fonts.StdFontName

// Names of the standard 14 fonts.
var (
	CourierName              = fonts.CourierName
	CourierBoldName          = fonts.CourierBoldName
	CourierObliqueName       = fonts.CourierObliqueName
	CourierBoldObliqueName   = fonts.CourierBoldObliqueName
	HelveticaName            = fonts.HelveticaName
	HelveticaBoldName        = fonts.HelveticaBoldName
	HelveticaObliqueName     = fonts.HelveticaObliqueName
	HelveticaBoldObliqueName = fonts.HelveticaBoldObliqueName
	SymbolName               = fonts.SymbolName
	ZapfDingbatsName         = fonts.ZapfDingbatsName
	TimesRomanName           = fonts.TimesRomanName
	TimesBoldName            = fonts.TimesBoldName
	TimesItalicName          = fonts.TimesItalicName
	TimesBoldItalicName      = fonts.TimesBoldItalicName
)

// NewStandard14Font returns the standard 14 font named `basefont` as a *PdfFont, or an error if it
// `basefont` is not one of the standard 14 font names.
func NewStandard14Font(basefont StdFontName) (*PdfFont, error) {
	std, err := newStandard14Font(basefont)
	if err != nil {
		return nil, err
	}

	if basefont != SymbolName && basefont != ZapfDingbatsName {
		// Default to using WinAnsiEncoder for text generation as it spans a large number of symbols.
		std.encoder = textencoding.NewWinAnsiEncoder()
	}

	return &PdfFont{context: &std}, nil
}

// NewStandard14FontMustCompile returns the standard 14 font named `basefont` as a *PdfFont.
// If `basefont` is one of the 14 Standard14Font values defined above then NewStandard14FontMustCompile
// is guaranteed to succeed.
func NewStandard14FontMustCompile(basefont StdFontName) *PdfFont {
	font, err := NewStandard14Font(basefont)
	if err != nil {
		panic(fmt.Errorf("invalid Standard14Font %#q", basefont))
	}
	return font
}

// NewStandard14FontWithEncoding returns the standard 14 font named `basefont` as a *PdfFont and
// a TextEncoder that encodes all the runes in `alphabet`, or an error if this is not possible.
// An error can occur if `basefont` is not one the standard 14 font names.
func NewStandard14FontWithEncoding(basefont StdFontName, alphabet map[rune]int) (*PdfFont,
	textencoding.SimpleEncoder, error) {
	std, err := newStandard14Font(basefont)
	if err != nil {
		return nil, nil, err
	}
	enc, ok := std.Encoder().(textencoding.SimpleEncoder)
	if !ok {
		return nil, nil, fmt.Errorf("only simple encoding is supported, got %T", std.Encoder())
	}

	// collect all runes from alphabet that are missing in the encoding
	// and find corresponding glyph names
	missing := make(map[rune]textencoding.GlyphName)
	for r := range alphabet {
		if _, ok := enc.RuneToCharcode(r); !ok {
			_, ok := std.fontMetrics[r]
			if !ok {
				common.Log.Trace("rune %#x=%q not in the font", r, r)
				continue
			}
			glyph, ok := textencoding.RuneToGlyph(r)
			if !ok {
				common.Log.Debug("no glyph for rune %#x=%q", r, r)
				continue
			}
			if len(missing) >= 255 {
				return nil, nil, errors.New("too many characters for simple encoding")
			}
			missing[r] = glyph
		}
	}

	// collect the list of empty indexes in the encoding that can be filed
	// and join the list of runes unused in the alphabet to overwrite, if necessary
	var (
		gaps   []textencoding.CharCode
		unused []textencoding.CharCode
	)
	// note, that this loop will become endless if CharCode becomes a byte
	for code := textencoding.CharCode(1); code <= 0xff; code++ {
		r, ok := enc.CharcodeToRune(code)
		if !ok {
			gaps = append(gaps, code)
			continue
		}
		if _, ok = alphabet[r]; !ok {
			unused = append(unused, code)
		}
	}
	// join into a single list of replacable charcodes, gaps first
	replacable := append(gaps, unused...)

	if len(replacable) < len(missing) {
		return nil, nil, fmt.Errorf("need to encode %d runes, but have only %d slots",
			len(missing), len(replacable))
	}

	// sort, make an order predictable
	runes := make([]rune, 0, len(missing))
	for r := range missing {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	// build a map of replacements
	differences := make(map[textencoding.CharCode]textencoding.GlyphName, len(runes))
	for _, r := range runes {
		code := replacable[0]
		replacable = replacable[1:]

		differences[code] = missing[r]
	}
	enc = textencoding.ApplyDifferences(enc, differences)
	std.SetEncoder(enc)

	return &PdfFont{context: &std}, enc, nil
}

// GetAlphabet returns a map of the runes in `text` and their frequencies.
func GetAlphabet(text string) map[rune]int {
	alphabet := map[rune]int{}
	for _, r := range text {
		alphabet[r]++
	}
	return alphabet
}

// NewPdfFontFromPdfObject loads a PdfFont from the dictionary `fontObj`.  If there is a problem an
// error is returned.
func NewPdfFontFromPdfObject(fontObj core.PdfObject) (*PdfFont, error) {
	return newPdfFontFromPdfObject(fontObj, true)
}

// newPdfFontFromPdfObject loads a PdfFont from the dictionary `fontObj`.  If there is a problem an
// error is returned.
// The allowType0 flag indicates whether loading Type0 font should be supported.  This is used to
// avoid cyclical loading.
func newPdfFontFromPdfObject(fontObj core.PdfObject, allowType0 bool) (*PdfFont, error) {
	d, base, err := newFontBaseFieldsFromPdfObject(fontObj)
	if err != nil {
		// In the case of not yet supported fonts, we attempt to return enough information in the
		// font for the caller to see some font properties.
		// TODO(peterwilliams97): Add support for these fonts and remove this special error handling.
		if err == ErrType3FontNotSupported || err == ErrType1CFontNotSupported {
			simplefont, err2 := newSimpleFontFromPdfObject(d, base, nil)
			if err2 != nil {
				common.Log.Debug("ERROR: While loading simple font: font=%s err=%v", base, err2)
				return nil, err
			}
			return &PdfFont{context: simplefont}, err
		}

		return nil, err
	}

	font := &PdfFont{}
	switch base.subtype {
	case "Type0":
		if !allowType0 {
			common.Log.Debug("ERROR: Loading type0 not allowed. font=%s", base)
			return nil, errors.New("cyclical type0 loading")
		}
		type0font, err := newPdfFontType0FromPdfObject(d, base)
		if err != nil {
			common.Log.Debug("ERROR: While loading Type0 font. font=%s err=%v", base, err)
			return nil, err
		}
		font.context = type0font
	case "Type1", "Type3", "MMType1", "TrueType":
		var simplefont *pdfFontSimple
		fnt, builtin := fonts.NewStdFontByName(fonts.StdFontName(base.basefont))
		if builtin {
			std := stdFontToSimpleFont(fnt)
			font.context = &std

			stdObj := core.TraceToDirectObject(std.ToPdfObject())
			d14, stdBase, err := newFontBaseFieldsFromPdfObject(stdObj)

			if err != nil {
				common.Log.Debug("ERROR: Bad Standard14\n\tfont=%s\n\tstd=%+v", base, std)
				return nil, err
			}

			for _, k := range d.Keys() {
				d14.Set(k, d.Get(k))
			}
			simplefont, err = newSimpleFontFromPdfObject(d14, stdBase, std.std14Encoder)
			if err != nil {
				common.Log.Debug("ERROR: Bad Standard14\n\tfont=%s\n\tstd=%+v", base, std)
				return nil, err
			}

			simplefont.charWidths = std.charWidths
			simplefont.fontMetrics = std.fontMetrics
		} else {
			simplefont, err = newSimpleFontFromPdfObject(d, base, nil)
			if err != nil {
				common.Log.Debug("ERROR: While loading simple font: font=%s err=%v", base, err)
				return nil, err
			}
		}
		err = simplefont.addEncoding()
		if err != nil {
			return nil, err
		}
		if builtin {
			simplefont.updateStandard14Font()
		}
		if builtin && simplefont.encoder == nil && simplefont.std14Encoder == nil {
			// This is not possible.
			common.Log.Error("simplefont=%s", simplefont)
			common.Log.Error("fnt=%+v", fnt)
		}
		if len(simplefont.charWidths) == 0 {
			common.Log.Debug("ERROR: No widths. font=%s", simplefont)
		}
		font.context = simplefont
	case "CIDFontType0":
		cidfont, err := newPdfCIDFontType0FromPdfObject(d, base)
		if err != nil {
			common.Log.Debug("ERROR: While loading cid font type0 font: %v", err)
			return nil, err
		}
		font.context = cidfont
	case "CIDFontType2":
		cidfont, err := newPdfCIDFontType2FromPdfObject(d, base)
		if err != nil {
			common.Log.Debug("ERROR: While loading cid font type2 font. font=%s err=%v", base, err)
			return nil, err
		}
		font.context = cidfont
	default:
		common.Log.Debug("ERROR: Unsupported font type: font=%s", base)
		return nil, fmt.Errorf("unsupported font type: font=%s", base)
	}

	return font, nil
}

// BytesToCharcodes converts the bytes in a PDF string to character codes.
func (font *PdfFont) BytesToCharcodes(data []byte) []textencoding.CharCode {
	common.Log.Trace("BytesToCharcodes: data=[% 02x]=%#q", data, data)
	if type0, ok := font.context.(*pdfFontType0); ok && type0.codeToCID != nil {
		if charcodes, ok := type0.bytesToCharcodes(data); ok {
			return charcodes
		}
	}

	charcodes := make([]textencoding.CharCode, 0, len(data)+len(data)%2)
	if font.baseFields().isCIDFont() {
		// Identity only?
		if len(data) == 1 {
			data = []byte{0, data[0]}
		}
		if len(data)%2 != 0 {
			common.Log.Debug("ERROR: Padding data=%+v to even length", data)
			data = append(data, 0)
		}
		for i := 0; i < len(data); i += 2 {
			b := uint16(data[i])<<8 | uint16(data[i+1])
			charcodes = append(charcodes, textencoding.CharCode(b))
		}
	} else {
		// Simple font: byte -> charcode.
		for _, b := range data {
			charcodes = append(charcodes, textencoding.CharCode(b))
		}
	}
	return charcodes
}

// CharcodesToUnicodeWithStats is identical to CharcodesToUnicode except it returns more statistical
// information about hits and misses from the reverse mapping process.
// NOTE: The number of runes returned may be greater than the number of charcodes.
// TODO(peterwilliams97): Deprecate in v4 and use only CharcodesToStrings()
func (font *PdfFont) CharcodesToUnicodeWithStats(charcodes []textencoding.CharCode) (runelist []rune, numHits, numMisses int) {
	texts, numHits, numMisses := font.CharcodesToStrings(charcodes)
	return []rune(strings.Join(texts, "")), numHits, numMisses
}

// CharcodesToStrings returns the unicode strings corresponding to `charcodes`.
// The int returns are the number of strings and the number of unconvereted codes.
// NOTE: The number of strings returned is equal to the number of charcodes
func (font *PdfFont) CharcodesToStrings(charcodes []textencoding.CharCode) ([]string, int, int) {
	fontBase := font.baseFields()
	texts := make([]string, 0, len(charcodes))
	numMisses := 0
	for _, code := range charcodes {
		if fontBase.toUnicodeCmap != nil {
			if s, ok := fontBase.toUnicodeCmap.CharcodeToUnicode(cmap.CharCode(code)); ok {
				texts = append(texts, s)
				continue
			}
		}

		// Fall back to encoding.
		encoder := font.Encoder()
		if encoder != nil {
			if r, ok := encoder.CharcodeToRune(code); ok {
				texts = append(texts, string(r))
				continue
			}
		}

		common.Log.Debug("ERROR: No rune. code=0x%04x charcodes=[% 04x] CID=%t\n"+
			"\tfont=%s\n\tencoding=%s",
			code, charcodes, fontBase.isCIDFont(), font, encoder)
		numMisses++
		texts = append(texts, cmap.MissingCodeString)
	}

	if numMisses != 0 {
		common.Log.Debug("ERROR: Couldn't convert to unicode. Using input.\n"+
			"\tnumChars=%d numMisses=%d\n"+
			"\tfont=%s",
			len(charcodes), numMisses, font)
	}

	return texts, len(texts), numMisses
}

// CharcodeBytesToUnicode converts PDF character codes `data` to a Go unicode string.
//
// 9.10 Extraction of Text Content (page 292)
// The process of finding glyph descriptions in OpenType fonts by a conforming reader shall be the following:
// • For Type 1 fonts using “CFF” tables, the process shall be as described in 9.6.6.2, "Encodings
//   for Type 1 Fonts".
// • For TrueType fonts using “glyf” tables, the process shall be as described in 9.6.6.4,
//   "Encodings for TrueType Fonts". Since this process sometimes produces ambiguous results,
//   conforming writers, instead of using a simple font, shall use a Type 0 font with an Identity-H
//   encoding and use the glyph indices as character codes, as described following Table 118.
func (font *PdfFont) CharcodeBytesToUnicode(data []byte) (string, int, int) {
	runes, _, numMisses := font.CharcodesToUnicodeWithStats(font.BytesToCharcodes(data))
	str := textencoding.ExpandLigatures(runes)
	return str, utf8.RuneCountInString(str), numMisses
}

// CharcodesToUnicode converts the character codes `charcodes` to a slice of runes.
// How it works:
//  1) Use the ToUnicode CMap if there is one.
//  2) Use the underlying font's encoding.
func (font *PdfFont) CharcodesToUnicode(charcodes []textencoding.CharCode) []rune {
	runes, _, _ := font.CharcodesToUnicodeWithStats(charcodes)
	return runes
}

// RunesToCharcodeBytes maps the provided runes to charcode bytes and it
// returns the resulting slice of bytes, along with the number of runes which
// could not be converted. If the number of misses is 0, all runes were
// successfully converted.
func (font *PdfFont) RunesToCharcodeBytes(data []rune) ([]byte, int) {
	// Create collection of encoders used for rune to charcode mapping:
	// - if the font has a to Unicode CMap, use it first.
	// - if the font has an encoder, use it as a fallback.
	var encoders []textencoding.TextEncoder
	if toUnicode := font.baseFields().toUnicodeCmap; toUnicode != nil {
		encoders = append(encoders, textencoding.NewCMapEncoder("", nil, toUnicode))
	}
	if encoder := font.Encoder(); encoder != nil {
		encoders = append(encoders, encoder)
	}

	var buffer bytes.Buffer
	var numMisses int
	for _, r := range data {
		// Attempt to encode the current rune using each of the encoders,
		// falling back to the next one in case of failure.
		var encoded bool
		for _, encoder := range encoders {
			if encBytes := encoder.Encode(string(r)); len(encBytes) > 0 {
				buffer.Write(encBytes)
				encoded = true
				break
			}
		}

		if !encoded {
			common.Log.Debug("ERROR: failed to map rune `%+q` to charcode", r)
			numMisses++
		}
	}

	if numMisses != 0 {
		common.Log.Debug("ERROR: could not convert all runes to charcodes.\n"+
			"\tnumRunes=%d numMisses=%d\n"+
			"\tfont=%s encoders=%+v", len(data), numMisses, font, encoders)
	}

	return buffer.Bytes(), numMisses
}

// StringToCharcodeBytes maps the provided string runes to charcode bytes and
// it returns the resulting slice of bytes, along with the number of runes
// which could not be converted. If the number of misses is 0, all string runes
// were successfully converted.
func (font *PdfFont) StringToCharcodeBytes(str string) ([]byte, int) {
	return font.RunesToCharcodeBytes([]rune(str))
}

// ToPdfObject converts the PdfFont object to its PDF representation.
func (font *PdfFont) ToPdfObject() core.PdfObject {
	if font.context == nil {
		common.Log.Debug("ERROR: font context is nil")
		return core.MakeNull()
	}
	return font.context.ToPdfObject()
}

// Encoder returns the font's text encoder.
func (font *PdfFont) Encoder() textencoding.TextEncoder {
	t := font.actualFont()
	if t == nil {
		common.Log.Debug("ERROR: Encoder not implemented for font type=%#T", font.context)
		// TODO: Should we return a default encoding?
		return nil
	}
	return t.Encoder()
}

// CharMetrics represents width and height metrics of a glyph.
type CharMetrics = fonts.CharMetrics

// GetRuneMetrics returns the char metrics for a rune.
// TODO(peterwilliams97) There is nothing callers can do if no CharMetrics are found so we might as
//                       well give them 0 width. There is no need for the bool return.
func (font *PdfFont) GetRuneMetrics(r rune) (CharMetrics, bool) {
	t := font.actualFont()
	if t == nil {
		common.Log.Debug("ERROR: GetGlyphCharMetrics Not implemented for font type=%#T", font.context)
		return fonts.CharMetrics{}, false
	}
	if m, ok := t.GetRuneMetrics(r); ok {
		return m, true
	}
	if desc, err := font.GetFontDescriptor(); err == nil && desc != nil {
		return fonts.CharMetrics{Wx: desc.missingWidth}, true
	}

	common.Log.Debug("GetGlyphCharMetrics: No metrics for font=%s", font)
	return fonts.CharMetrics{}, false
}

// GetCharMetrics returns the char metrics for character code `code`.
// How it works:
//  1) It calls the GetCharMetrics function for the underlying font, either a simple font or
//     a Type0 font. The underlying font GetCharMetrics() functions do direct charcode ➞  metrics
//     mappings.
//  2) If the underlying font's GetCharMetrics() doesn't have a CharMetrics for `code` then a
//     a CharMetrics with the FontDescriptor's /MissingWidth is returned.
//  3) If there is no /MissingWidth then a failure is returned.
// TODO(peterwilliams97) There is nothing callers can do if no CharMetrics are found so we might as
//                       well give them 0 width. There is no need for the bool return.
// TODO(gunnsth): Reconsider whether needed or if can map via GlyphName.
func (font *PdfFont) GetCharMetrics(code textencoding.CharCode) (CharMetrics, bool) {
	var nometrics fonts.CharMetrics

	// TODO(peterwilliams97): pdfFontType0.GetCharMetrics() calls pdfCIDFontType2.GetCharMetrics()
	// 						  through this function. Would it be more straightforward for
	// 						  pdfFontType0.GetCharMetrics() to call pdfCIDFontType0.GetCharMetrics()
	// 						  and pdfCIDFontType2.GetCharMetrics() directly?

	switch t := font.context.(type) {
	case *pdfFontSimple:
		if m, ok := t.GetCharMetrics(code); ok {
			return m, ok
		}
	case *pdfFontType0:
		if m, ok := t.GetCharMetrics(code); ok {
			return m, ok
		}
	case *pdfCIDFontType0:
		if m, ok := t.GetCharMetrics(code); ok {
			return m, ok
		}
	case *pdfCIDFontType2:
		if m, ok := t.GetCharMetrics(code); ok {
			return m, ok
		}
	default:
		common.Log.Debug("ERROR: GetCharMetrics not implemented for font type=%T.", font.context)
		return nometrics, false
	}

	if descriptor, err := font.GetFontDescriptor(); err == nil && descriptor != nil {
		return fonts.CharMetrics{Wx: descriptor.missingWidth}, true
	}

	common.Log.Debug("GetCharMetrics: No metrics for font=%s", font)
	return nometrics, false
}

// actualFont returns the Font in font.context
func (font PdfFont) actualFont() pdfFont {
	if font.context == nil {
		common.Log.Debug("ERROR: actualFont. context is nil. font=%s", font)
	}
	return font.context
}

// baseFields returns the fields of `font`.context that are common to all PDF fonts.
func (font *PdfFont) baseFields() *fontCommon {
	if font.context == nil {
		common.Log.Debug("ERROR: baseFields. context is nil.")
		return nil
	}
	return font.context.baseFields()
}

// fontCommon represents the fields that are common to all PDF fonts.
type fontCommon struct {
	// All fonts have these fields.
	basefont string // The font's "BaseFont" field.
	subtype  string // The font's "Subtype" field.
	name     string

	// These are optional fields in the PDF font.
	toUnicode core.PdfObject // The stream containing toUnicodeCmap. We keep it around for ToPdfObject.

	// These objects are computed from optional fields in the PDF font.
	toUnicodeCmap  *cmap.CMap         // Computed from "ToUnicode".
	fontDescriptor *PdfFontDescriptor // Computed from "FontDescriptor".

	// objectNumber helps us find the font in the PDF being processed. This helps with debugging.
	objectNumber int64
}

// asPdfObjectDictionary returns `base` as a core.PdfObjectDictionary.
// It is for use in font ToPdfObject functions.
// NOTE: The returned dict's "Subtype" field is set to `subtype` if `base` doesn't have a subtype.
func (base fontCommon) asPdfObjectDictionary(subtype string) *core.PdfObjectDictionary {
	if subtype != "" && base.subtype != "" && subtype != base.subtype {
		common.Log.Debug("ERROR: asPdfObjectDictionary. Overriding subtype to %#q %s", subtype, base)
	} else if subtype == "" && base.subtype == "" {
		common.Log.Debug("ERROR: asPdfObjectDictionary no subtype. font=%s", base)
	} else if base.subtype == "" {
		base.subtype = subtype
	}

	d := core.MakeDict()
	d.Set("Type", core.MakeName("Font"))
	d.Set("BaseFont", core.MakeName(base.basefont))
	d.Set("Subtype", core.MakeName(base.subtype))

	if base.fontDescriptor != nil {
		d.Set("FontDescriptor", base.fontDescriptor.ToPdfObject())
	}
	if base.toUnicode != nil {
		d.Set("ToUnicode", base.toUnicode)
	} else if base.toUnicodeCmap != nil {
		o, err := base.toUnicodeCmap.Stream()
		if err != nil {
			common.Log.Debug("WARN: could not get CMap stream. err=%v", err)
		} else {
			d.Set("ToUnicode", o)
		}
	}
	return d
}

// String returns a string that describes `base`.
func (base fontCommon) String() string {
	return fmt.Sprintf("FONT{%s}", base.coreString())
}

// coreString returns the contents of fontCommon.String() without the FONT{} wrapper.
func (base fontCommon) coreString() string {
	descriptor := ""
	if base.fontDescriptor != nil {
		descriptor = base.fontDescriptor.String()
	}
	return fmt.Sprintf("%#q %#q %q obj=%d ToUnicode=%t flags=0x%0x %s",
		base.subtype, base.basefont, base.name, base.objectNumber, base.toUnicode != nil,
		base.fontFlags(), descriptor)
}

func (base fontCommon) fontFlags() int {
	if base.fontDescriptor == nil {
		return 0
	}
	return base.fontDescriptor.flags
}

// isCIDFont returns true if `base` is a CID font.
func (base fontCommon) isCIDFont() bool {
	if base.subtype == "" {
		common.Log.Debug("ERROR: isCIDFont. context is nil. font=%s", base)
	}
	isCID := false
	switch base.subtype {
	case "Type0", "CIDFontType0", "CIDFontType2":
		isCID = true
	}
	common.Log.Trace("isCIDFont: isCID=%t font=%s", isCID, base)
	return isCID
}

// newFontBaseFieldsFromPdfObject returns `fontObj` as a dictionary the common fields from that
// dictionary in the fontCommon return.  If there is a problem an error is returned.
// The fontCommon is the group of fields common to all PDF fonts.
func newFontBaseFieldsFromPdfObject(fontObj core.PdfObject) (*core.PdfObjectDictionary, *fontCommon, error) {
	font := &fontCommon{}

	if obj, ok := fontObj.(*core.PdfIndirectObject); ok {
		font.objectNumber = obj.ObjectNumber
	}

	d, ok := core.GetDict(fontObj)
	if !ok {
		common.Log.Debug("ERROR: Font not given by a dictionary (%T)", fontObj)
		return nil, nil, ErrFontNotSupported
	}

	objtype, ok := core.GetNameVal(d.Get("Type"))
	if !ok {
		common.Log.Debug("ERROR: Font Incompatibility. Type (Required) missing")
		return nil, nil, ErrRequiredAttributeMissing
	}
	if objtype != "Font" {
		common.Log.Debug("ERROR: Font Incompatibility. Type=%q. Should be %q.", objtype, "Font")
		return nil, nil, core.ErrTypeError
	}

	subtype, ok := core.GetNameVal(d.Get("Subtype"))
	if !ok {
		common.Log.Debug("ERROR: Font Incompatibility. Subtype (Required) missing")
		return nil, nil, ErrRequiredAttributeMissing
	}
	font.subtype = subtype

	name, ok := core.GetNameVal(d.Get("Name"))
	if ok {
		font.name = name
	}

	if subtype == "Type3" {
		common.Log.Debug("ERROR: Type 3 font not supported. d=%s", d)
		return d, font, ErrType3FontNotSupported
	}

	basefont, ok := core.GetNameVal(d.Get("BaseFont"))
	if !ok {
		common.Log.Debug("ERROR: Font Incompatibility. BaseFont (Required) missing")
		return d, font, ErrRequiredAttributeMissing
	}
	font.basefont = basefont

	obj := d.Get("FontDescriptor")
	if obj != nil {
		fontDescriptor, err := newPdfFontDescriptorFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("ERROR: Bad font descriptor. err=%v", err)
			return d, font, err
		}
		font.fontDescriptor = fontDescriptor
	}

	toUnicode := d.Get("ToUnicode")
	if toUnicode != nil {
		font.toUnicode = core.TraceToDirectObject(toUnicode)
		codemap, err := toUnicodeToCmap(font.toUnicode, font)
		if err != nil {
			return d, font, err
		}
		font.toUnicodeCmap = codemap
	} else if subtype == "CIDFontType0" || subtype == "CIDFontType2" {
		si, err := cmap.NewCIDSystemInfo(d.Get("CIDSystemInfo"))
		if err != nil {
			return d, font, err
		}

		cmapName := fmt.Sprintf("%s-%s-UCS2", si.Registry, si.Ordering)
		if cmap.IsPredefinedCMap(cmapName) {
			font.toUnicodeCmap, err = cmap.LoadPredefinedCMap(cmapName)
			if err != nil {
				common.Log.Debug("WARN: could not load predefined CMap %s: %v", cmapName, err)
			}
		}
	}

	return d, font, nil
}

// toUnicodeToCmap returns a CMap of `toUnicode` if it exists.
func toUnicodeToCmap(toUnicode core.PdfObject, font *fontCommon) (*cmap.CMap, error) {
	toUnicodeStream, ok := core.GetStream(toUnicode)
	if !ok {
		common.Log.Debug("ERROR: toUnicodeToCmap: Not a stream (%T)", toUnicode)
		return nil, core.ErrTypeError
	}
	data, err := core.DecodeStream(toUnicodeStream)
	if err != nil {
		return nil, err
	}

	cm, err := cmap.LoadCmapFromData(data, !font.isCIDFont())
	if err != nil {
		// Show the object number of the bad cmap to help with debugging.
		common.Log.Debug("ERROR: ObjectNumber=%d err=%v", toUnicodeStream.ObjectNumber, err)
	}
	return cm, err
}

// 9.8.2 Font Descriptor Flags (page 283)
const (
	fontFlagFixedPitch  = 0x00001
	fontFlagSerif       = 0x00002
	fontFlagSymbolic    = 0x00004
	fontFlagScript      = 0x00008
	fontFlagNonsymbolic = 0x00020
	fontFlagItalic      = 0x00040
	fontFlagAllCap      = 0x10000
	fontFlagSmallCap    = 0x20000
	fontFlagForceBold   = 0x40000
)

// PdfFontDescriptor specifies metrics and other attributes of a font and can refer to a FontFile
// for embedded fonts.
// 9.8 Font Descriptors (page 281)
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
	FontFile     core.PdfObject // PFB
	FontFile2    core.PdfObject // TTF
	FontFile3    core.PdfObject // OTF / CFF
	CharSet      core.PdfObject

	flags        int
	missingWidth float64
	*fontFile
	fontFile2 *fonts.TtfType

	// Additional entries for CIDFonts
	Style  core.PdfObject
	Lang   core.PdfObject
	FD     core.PdfObject
	CIDSet core.PdfObject

	// Container.
	container *core.PdfIndirectObject
}

// GetDescent returns the Descent of the font `descriptor`.
func (desc *PdfFontDescriptor) GetDescent() (float64, error) {
	return core.GetNumberAsFloat(desc.Descent)
}

// GetAscent returns the Ascent of the font `descriptor`.
func (desc *PdfFontDescriptor) GetAscent() (float64, error) {
	return core.GetNumberAsFloat(desc.Ascent)
}

// GetCapHeight returns the CapHeight of the font `descriptor`.
func (desc *PdfFontDescriptor) GetCapHeight() (float64, error) {
	return core.GetNumberAsFloat(desc.CapHeight)
}

// String returns a string describing the font descriptor.
func (desc *PdfFontDescriptor) String() string {
	var parts []string
	if desc.FontName != nil {
		parts = append(parts, desc.FontName.String())
	}
	if desc.FontFamily != nil {
		parts = append(parts, desc.FontFamily.String())
	}
	if desc.fontFile != nil {
		parts = append(parts, desc.fontFile.String())
	}
	if desc.fontFile2 != nil {
		parts = append(parts, desc.fontFile2.String())
	}
	parts = append(parts, fmt.Sprintf("FontFile3=%t", desc.FontFile3 != nil))

	return fmt.Sprintf("FONT_DESCRIPTOR{%s}", strings.Join(parts, ", "))
}

// newPdfFontDescriptorFromPdfObject loads the font descriptor from a core.PdfObject.  Can either be a
// *PdfIndirectObject or a *core.PdfObjectDictionary.
func newPdfFontDescriptorFromPdfObject(obj core.PdfObject) (*PdfFontDescriptor, error) {
	descriptor := &PdfFontDescriptor{}

	obj = core.ResolveReference(obj)
	if ind, is := obj.(*core.PdfIndirectObject); is {
		descriptor.container = ind
		obj = ind.PdfObject
	}

	d, ok := core.GetDict(obj)
	if !ok {
		common.Log.Debug("ERROR: FontDescriptor not given by a dictionary (%T)", obj)
		return nil, core.ErrTypeError
	}

	if obj := d.Get("FontName"); obj != nil {
		descriptor.FontName = obj
	} else {
		common.Log.Debug("Incompatibility: FontName (Required) missing")
	}
	fontname, _ := core.GetName(descriptor.FontName)

	if obj := d.Get("Type"); obj != nil {
		oname, is := obj.(*core.PdfObjectName)
		if !is || string(*oname) != "FontDescriptor" {
			common.Log.Debug("Incompatibility: Font descriptor Type invalid (%T) font=%q %T",
				obj, fontname, descriptor.FontName)
		}
	} else {
		common.Log.Trace("Incompatibility: Type (Required) missing. font=%q %T",
			fontname, descriptor.FontName)
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

	if descriptor.Flags != nil {
		if flags, ok := core.GetIntVal(descriptor.Flags); ok {
			descriptor.flags = flags
		}
	}
	if descriptor.MissingWidth != nil {
		if missingWidth, err := core.GetNumberAsFloat(descriptor.MissingWidth); err == nil {
			descriptor.missingWidth = missingWidth
		}
	}

	if descriptor.FontFile != nil {
		fontFile, err := newFontFileFromPdfObject(descriptor.FontFile)
		if err != nil {
			return descriptor, err
		}
		common.Log.Trace("fontFile=%s", fontFile)
		descriptor.fontFile = fontFile
	}
	if descriptor.FontFile2 != nil {
		fontFile2, err := fonts.NewFontFile2FromPdfObject(descriptor.FontFile2)
		if err != nil {
			return descriptor, err
		}
		common.Log.Trace("fontFile2=%s", fontFile2.String())
		descriptor.fontFile2 = &fontFile2
	}
	return descriptor, nil
}

// ToPdfObject returns the PdfFontDescriptor as a PDF dictionary inside an indirect object.
func (desc *PdfFontDescriptor) ToPdfObject() core.PdfObject {
	d := core.MakeDict()
	if desc.container == nil {
		desc.container = &core.PdfIndirectObject{}
	}
	desc.container.PdfObject = d

	d.Set("Type", core.MakeName("FontDescriptor"))

	if desc.FontName != nil {
		d.Set("FontName", desc.FontName)
	}

	if desc.FontFamily != nil {
		d.Set("FontFamily", desc.FontFamily)
	}

	if desc.FontStretch != nil {
		d.Set("FontStretch", desc.FontStretch)
	}

	if desc.FontWeight != nil {
		d.Set("FontWeight", desc.FontWeight)
	}

	if desc.Flags != nil {
		d.Set("Flags", desc.Flags)
	}

	if desc.FontBBox != nil {
		d.Set("FontBBox", desc.FontBBox)
	}

	if desc.ItalicAngle != nil {
		d.Set("ItalicAngle", desc.ItalicAngle)
	}

	if desc.Ascent != nil {
		d.Set("Ascent", desc.Ascent)
	}

	if desc.Descent != nil {
		d.Set("Descent", desc.Descent)
	}

	if desc.Leading != nil {
		d.Set("Leading", desc.Leading)
	}

	if desc.CapHeight != nil {
		d.Set("CapHeight", desc.CapHeight)
	}

	if desc.XHeight != nil {
		d.Set("XHeight", desc.XHeight)
	}

	if desc.StemV != nil {
		d.Set("StemV", desc.StemV)
	}

	if desc.StemH != nil {
		d.Set("StemH", desc.StemH)
	}

	if desc.AvgWidth != nil {
		d.Set("AvgWidth", desc.AvgWidth)
	}

	if desc.MaxWidth != nil {
		d.Set("MaxWidth", desc.MaxWidth)
	}

	if desc.MissingWidth != nil {
		d.Set("MissingWidth", desc.MissingWidth)
	}

	if desc.FontFile != nil {
		d.Set("FontFile", desc.FontFile)
	}

	if desc.FontFile2 != nil {
		d.Set("FontFile2", desc.FontFile2)
	}

	if desc.FontFile3 != nil {
		d.Set("FontFile3", desc.FontFile3)
	}

	if desc.CharSet != nil {
		d.Set("CharSet", desc.CharSet)
	}

	if desc.Style != nil {
		d.Set("FontName", desc.FontName)
	}

	if desc.Lang != nil {
		d.Set("Lang", desc.Lang)
	}

	if desc.FD != nil {
		d.Set("FD", desc.FD)
	}

	if desc.CIDSet != nil {
		d.Set("CIDSet", desc.CIDSet)
	}

	return desc.container
}
