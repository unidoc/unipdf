/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"

	"github.com/unidoc/unipdf/v3/internal/textencoding"
	"github.com/unidoc/unipdf/v3/model/internal/fonts"
)

// pdfFontSimple implements pdfFont
var _ pdfFont = (*pdfFontSimple)(nil)

// pdfFontSimple describes a Simple Font
//
// 9.6 Simple Fonts (page 254)
// 9.6.1 General
// There are several types of simple fonts, all of which have these properties:
// • Glyphs in the font shall be selected by single-byte character codes obtained from a string that
//   is shown by the text-showing operators. Logically, these codes index into a table of 256 glyphs;
//   the mapping from codes to glyphs is called the font’s encoding. Under some circumstances, the
//   encoding may be altered by means described in 9.6.6, "Character Encoding".
// • Each glyph shall have a single set of metrics, including a horizontal displacement or width,
//   as described in 9.2.4, "Glyph Positioning and Metrics"; that is, simple fonts support only
//   horizontal writing mode.
// • Except for Type 0 fonts, Type 3 fonts in non-Tagged PDF documents, and certain standard Type 1
//   fonts, every font dictionary shall contain a subsidiary dictionary, the font descriptor,
//   containing font-wide metrics and other attributes of the font.
//   Among those attributes is an optional font filestream containing the font program.
type pdfFontSimple struct {
	fontCommon
	container *core.PdfIndirectObject

	// These fields are specific to simple PDF fonts.

	charWidths map[textencoding.CharCode]float64
	// encoder is the encoder specified by the /Encoding entry in the font dict.
	encoder textencoding.TextEncoder
	// std14Encoder is used for Standard 14 fonts where no /Encoding is specified in the font dict.
	std14Encoder textencoding.TextEncoder

	// std14Descriptor is used for Standard 14 fonts where no /FontDescriptor is specified in the font dict.
	std14Descriptor *PdfFontDescriptor

	// Encoding is subject to limitations that are described in 9.6.6, "Character Encoding".
	// BaseFont is derived differently.
	FirstChar core.PdfObject
	LastChar  core.PdfObject
	Widths    core.PdfObject
	Encoding  core.PdfObject

	// Standard 14 fonts metrics
	fontMetrics map[rune]fonts.CharMetrics
}

// pdfCIDFontType0FromSkeleton returns a pdfFontSimple with its common fields initalized.
func pdfFontSimpleFromSkeleton(base *fontCommon) *pdfFontSimple {
	return &pdfFontSimple{
		fontCommon: *base,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfFontSimple) baseFields() *fontCommon {
	return &font.fontCommon
}

func (font *pdfFontSimple) getFontDescriptor() *PdfFontDescriptor {
	if d := font.fontDescriptor; d != nil {
		return d
	}
	return font.std14Descriptor
}

// Encoder returns the font's text encoder.
func (font *pdfFontSimple) Encoder() textencoding.TextEncoder {
	// TODO(peterwilliams97): Need to make font.Encoder()==nil test work for
	// font.std14=Encoder=font.encoder=nil See https://golang.org/doc/faq#nil_error
	if font.encoder != nil {
		return font.encoder
	}

	// Standard 14 fonts have builtin encoders that we fall back to when no /Encoding is specified
	// in the font dict.
	if font.std14Encoder != nil {
		return font.std14Encoder
	}

	// Default to StandardEncoding
	enc, _ := textencoding.NewSimpleTextEncoder("StandardEncoding", nil)
	return enc
}

// SetEncoder sets the encoding for the underlying font.
// TODO(peterwilliams97): Change function signature to SetEncoder(encoder *textencoding.simpleEncoder).
// TODO(gunnsth): Makes sense if SetEncoder is removed from the interface fonts.Font as proposed in PR #260.
func (font *pdfFontSimple) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

// GetRuneMetrics returns the character metrics for the rune.
// A bool flag is returned to indicate whether or not the entry was found.
func (font pdfFontSimple) GetRuneMetrics(r rune) (fonts.CharMetrics, bool) {
	if font.fontMetrics != nil {
		metrics, has := font.fontMetrics[r]
		if has {
			return metrics, true
		}
	}
	encoder := font.Encoder()
	if encoder == nil {
		common.Log.Debug("No encoder for fonts=%s", font)
		return fonts.CharMetrics{}, false
	}
	code, found := encoder.RuneToCharcode(r)
	if !found {
		if r != ' ' {
			common.Log.Trace("No charcode for rune=%v font=%s", r, font)
		}
		return fonts.CharMetrics{}, false
	}
	metrics, ok := font.GetCharMetrics(code)
	return metrics, ok
}

// GetCharMetrics returns the character metrics for the specified character code.  A bool flag is
// returned to indicate whether or not the entry was found in the glyph to charcode mapping.
// How it works:
//  1) Return a value the /Widths array (charWidths) if there is one.
//  2) If the font has the same name as a standard 14 font then return width=250.
//  3) Otherwise return no match and let the caller substitute a default.
func (font pdfFontSimple) GetCharMetrics(code textencoding.CharCode) (fonts.CharMetrics, bool) {
	if width, ok := font.charWidths[code]; ok {
		return fonts.CharMetrics{Wx: width}, true
	}
	if fonts.IsStdFont(fonts.StdFontName(font.basefont)) {
		// PdfBox says this is what Acrobat does. Their reference is PDFBOX-2334.
		return fonts.CharMetrics{Wx: 250}, true
	}
	return fonts.CharMetrics{}, false
}

// newSimpleFontFromPdfObject creates a pdfFontSimple from dictionary `d`. Elements of `d` that
// are already parsed are contained in `base`.
// Standard 14 fonts need to to specify their builtin encoders in the `std14Encoder` parameter.
// An error is returned if there is a problem with loading.
//
// The value of Encoding is subject to limitations that are described in 9.6.6, "Character Encoding".
// • The value of BaseFont is derived differently.
//
func newSimpleFontFromPdfObject(d *core.PdfObjectDictionary, base *fontCommon,
	std14Encoder textencoding.TextEncoder) (*pdfFontSimple, error) {
	font := pdfFontSimpleFromSkeleton(base)
	font.std14Encoder = std14Encoder

	// FirstChar is not defined in ~/testdata/shamirturing.pdf
	if std14Encoder == nil {
		obj := d.Get("FirstChar")
		if obj == nil {
			obj = core.MakeInteger(0)
		}
		font.FirstChar = obj

		intVal, ok := core.GetIntVal(obj)
		if !ok {
			common.Log.Debug("ERROR: Invalid FirstChar type (%T)", obj)
			return nil, core.ErrTypeError
		}
		firstChar := textencoding.CharCode(intVal)

		obj = d.Get("LastChar")
		if obj == nil {
			obj = core.MakeInteger(255)
		}
		font.LastChar = obj
		intVal, ok = core.GetIntVal(obj)
		if !ok {
			common.Log.Debug("ERROR: Invalid LastChar type (%T)", obj)
			return nil, core.ErrTypeError
		}
		lastChar := textencoding.CharCode(intVal)

		font.charWidths = make(map[textencoding.CharCode]float64)
		obj = d.Get("Widths")
		if obj != nil {
			font.Widths = obj

			arr, ok := core.GetArray(obj)
			if !ok {
				common.Log.Debug("ERROR: Widths attribute != array (%T)", obj)
				return nil, core.ErrTypeError
			}

			widths, err := arr.ToFloat64Array()
			if err != nil {
				common.Log.Debug("ERROR: converting widths to array")
				return nil, err
			}

			if len(widths) != int(lastChar-firstChar+1) {
				common.Log.Debug("ERROR: Invalid widths length != %d (%d)",
					lastChar-firstChar+1, len(widths))
				return nil, core.ErrRangeError
			}
			for i, w := range widths {
				font.charWidths[firstChar+textencoding.CharCode(i)] = w
			}
		}
	}

	font.Encoding = core.TraceToDirectObject(d.Get("Encoding"))
	return font, nil
}

// addEncoding adds the encoding to the font and sets the `font.encoder` field.
// The order of precedence is important:
// 1. If encoder already set, load it initially (with subsequent steps potentially overwriting).
// 2. Attempts to construct the encoder from the Encoding dictionary.
// 3. If no encoder loaded, attempt to load from the font file.
// 4. Apply differences map and set as the `font`'s encoder.
func (font *pdfFontSimple) addEncoding() error {
	var (
		baseEncoder string
		differences map[textencoding.CharCode]textencoding.GlyphName
		encoder     textencoding.SimpleEncoder
	)

	if font.Encoder() != nil {
		encoder, ok := font.Encoder().(textencoding.SimpleEncoder)
		if ok && encoder != nil {
			baseEncoder = encoder.BaseName()
		}
	}

	if font.Encoding != nil {
		baseEncoderName, differences, err := font.getFontEncoding()
		if err != nil {
			common.Log.Debug("ERROR: BaseFont=%q Subtype=%q Encoding=%s (%T) err=%v", font.basefont,
				font.subtype, font.Encoding, font.Encoding, err)
			return err
		}
		if baseEncoderName != "" {
			baseEncoder = baseEncoderName
		}

		encoder, err = textencoding.NewSimpleTextEncoder(baseEncoder, differences)
		if err != nil {
			return err
		}
	}

	if encoder == nil {
		descriptor := font.fontDescriptor
		if descriptor != nil {
			switch font.subtype {
			case "Type1":
				if descriptor.fontFile != nil && descriptor.fontFile.encoder != nil {
					common.Log.Debug("Using fontFile")
					encoder = descriptor.fontFile.encoder
				}
			case "TrueType":
				if descriptor.fontFile2 != nil {
					common.Log.Debug("Using FontFile2")
					enc, err := descriptor.fontFile2.MakeEncoder()
					if err == nil {
						encoder = enc
					}
				}
			}
		}
	}

	if encoder != nil {
		// At the end, apply the differences.
		if differences != nil {
			common.Log.Trace("differences=%+v font=%s", differences, font.baseFields())
			encoder = textencoding.ApplyDifferences(encoder, differences)
		}
		font.SetEncoder(encoder)
	}

	return nil
}

// getFontEncoding returns font encoding of `obj` the "Encoding" entry in a font dict.
// Table 114 – Entries in an encoding dictionary (page 263)
// 9.6.6.1 General (page 262)
// A font’s encoding is the association between character codes (obtained from text strings that
// are shown) and glyph descriptions. This sub-clause describes the character encoding scheme used
// with simple PDF fonts. Composite fonts (Type 0) use a different character mapping algorithm, as
// discussed in 9.7, "Composite Fonts".
// Except for Type 3 fonts, every font program shall have a built-in encoding. Under certain
// circumstances, a PDF font dictionary may change the encoding used with the font program to match
// the requirements of the conforming writer generating the text being shown.
func (font *pdfFontSimple) getFontEncoding() (baseName string, differences map[textencoding.CharCode]textencoding.GlyphName, err error) {
	baseName = "StandardEncoding"
	if name, ok := builtinEncodings[font.basefont]; ok {
		baseName = name
	} else if font.fontFlags()&fontFlagSymbolic != 0 {
		for base, name := range builtinEncodings {
			if strings.Contains(font.basefont, base) {
				baseName = name
				break
			}
		}
	}

	if font.Encoding == nil {
		// Fall back to StandardEncoding | SymbolEncoding | ZapfDingbatsEncoding
		// This works because the only way BaseEncoding can get overridden is by FontFile entries
		// and the only encoding names we have seen in FontFile's are StandardEncoding or no entry.
		return baseName, nil, nil
	}

	switch encoding := font.Encoding.(type) {
	case *core.PdfObjectName:
		return string(*encoding), nil, nil
	case *core.PdfObjectDictionary:
		baseenc, ok := core.GetName(encoding.Get("BaseEncoding"))
		if ok {
			baseName = baseenc.String()
		}
		if diffObj := encoding.Get("Differences"); diffObj != nil {
			diffList, ok := core.GetArray(diffObj)
			if !ok {
				common.Log.Debug("ERROR: Bad font encoding dict=%+v Differences=%T",
					encoding, encoding.Get("Differences"))
				return "", nil, core.ErrTypeError
			}
			differences, err = textencoding.FromFontDifferences(diffList)
		}
		return baseName, differences, err
	default:
		common.Log.Debug("ERROR: Encoding not a name or dict (%T) %s", font.Encoding, font.Encoding)
		return "", nil, core.ErrTypeError
	}
}

var builtinEncodings = map[string]string{
	"Symbol":       "SymbolEncoding",
	"ZapfDingbats": "ZapfDingbatsEncoding",
}

// ToPdfObject converts the pdfFontSimple to its PDF representation for outputting.
func (font *pdfFontSimple) ToPdfObject() core.PdfObject {
	if font.container == nil {
		font.container = &core.PdfIndirectObject{}
	}
	d := font.baseFields().asPdfObjectDictionary("")
	font.container.PdfObject = d

	if font.FirstChar != nil {
		d.Set("FirstChar", font.FirstChar)
	}
	if font.LastChar != nil {
		d.Set("LastChar", font.LastChar)
	}
	if font.Widths != nil {
		d.Set("Widths", font.Widths)
	}
	if font.Encoding != nil {
		d.Set("Encoding", font.Encoding)
	} else if font.encoder != nil {
		encObj := font.encoder.ToPdfObject()
		if encObj != nil {
			d.Set("Encoding", encObj)
		}
	}

	return font.container
}

// NewPdfFontFromTTFFile loads a TTF font file and returns a PdfFont type
// that can be used in text styling functions.
// Uses a WinAnsiTextEncoder and loads only character codes 32-255.
// NOTE: For composite fonts such as used in symbolic languages, use NewCompositePdfFontFromTTFFile.
func NewPdfFontFromTTFFile(filePath string) (*PdfFont, error) {
	f, err := os.Open(filePath)
	if err != nil {
		common.Log.Debug("ERROR: reading TTF font file: %v", err)
		return nil, err
	}
	defer f.Close()

	return NewPdfFontFromTTF(f)
}

// NewPdfFontFromTTF loads a TTF font and returns a PdfFont type that can be
// used in text styling functions.
// Uses a WinAnsiTextEncoder and loads only character codes 32-255.
// NOTE: For composite fonts such as used in symbolic languages, use NewCompositePdfFontFromTTF.
func NewPdfFontFromTTF(r io.ReadSeeker) (*PdfFont, error) {
	const minCode = textencoding.CharCode(32)
	const maxCode = textencoding.CharCode(255)

	ttfBytes, err := ioutil.ReadAll(r)
	if err != nil {
		common.Log.Debug("ERROR: Unable to read font contents: %v", err)
		return nil, err
	}
	ttf, err := fonts.TtfParse(bytes.NewReader(ttfBytes))
	if err != nil {
		common.Log.Debug("ERROR: loading TTF font: %v", err)
		return nil, err
	}

	truefont := &pdfFontSimple{
		charWidths: make(map[textencoding.CharCode]float64),
		fontCommon: fontCommon{
			subtype: "TrueType",
		},
	}

	truefont.encoder = textencoding.NewWinAnsiEncoder()

	truefont.basefont = ttf.PostScriptName
	truefont.FirstChar = core.MakeInteger(int64(minCode))
	truefont.LastChar = core.MakeInteger(int64(maxCode))

	k := 1000.0 / float64(ttf.UnitsPerEm)
	if len(ttf.Widths) <= 0 {
		return nil, errors.New("ERROR: Missing required attribute (Widths)")
	}

	missingWidth := k * float64(ttf.Widths[0])

	vals := make([]float64, 0, maxCode-minCode+1)
	for code := minCode; code <= maxCode; code++ {
		r, found := truefont.Encoder().CharcodeToRune(code)
		if !found {
			common.Log.Debug("Rune not found (code: %d)", code)
			vals = append(vals, missingWidth)
			continue
		}

		gid, ok := ttf.Chars[r]
		if !ok {
			common.Log.Debug("Rune not in TTF Chars")
			vals = append(vals, missingWidth)
			continue
		}

		w := k * float64(ttf.Widths[gid])

		vals = append(vals, w)
	}

	truefont.Widths = core.MakeIndirectObject(core.MakeArrayFromFloats(vals))

	if len(vals) < int(maxCode-minCode+1) {
		common.Log.Debug("ERROR: Invalid length of widths, %d < %d", len(vals), 255-32+1)
		return nil, core.ErrRangeError
	}

	for i := textencoding.CharCode(minCode); i <= maxCode; i++ {
		truefont.charWidths[i] = vals[i-minCode]
	}

	// Use WinAnsiEncoding by default.
	truefont.Encoding = core.MakeName("WinAnsiEncoding")

	descriptor := &PdfFontDescriptor{}
	descriptor.FontName = core.MakeName(ttf.PostScriptName)
	descriptor.Ascent = core.MakeFloat(k * float64(ttf.TypoAscender))
	descriptor.Descent = core.MakeFloat(k * float64(ttf.TypoDescender))
	descriptor.CapHeight = core.MakeFloat(k * float64(ttf.CapHeight))
	descriptor.FontBBox = core.MakeArrayFromFloats([]float64{k * float64(ttf.Xmin),
		k * float64(ttf.Ymin), k * float64(ttf.Xmax), k * float64(ttf.Ymax)})
	descriptor.ItalicAngle = core.MakeFloat(float64(ttf.ItalicAngle))
	descriptor.MissingWidth = core.MakeFloat(k * float64(ttf.Widths[0]))

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

	flags := fontFlagNonsymbolic
	if ttf.IsFixedPitch {
		flags |= fontFlagFixedPitch
	}
	if ttf.ItalicAngle != 0 {
		flags |= fontFlagItalic
	}
	descriptor.Flags = core.MakeInteger(int64(flags))

	// Build Font.
	truefont.fontDescriptor = descriptor

	font := &PdfFont{
		context: truefont,
	}

	return font, nil
}

// updateStandard14Font fills the font.charWidths for standard 14 fonts.
// Don't call this function with a font that is not in the standard 14.
func (font *pdfFontSimple) updateStandard14Font() {
	se, ok := font.Encoder().(textencoding.SimpleEncoder)
	if !ok {
		// This can't happen.
		common.Log.Error("Wrong encoder type: %T. font=%s.", font.Encoder(), font)
		return
	}

	codes := se.Charcodes()
	font.charWidths = make(map[textencoding.CharCode]float64, len(codes))
	for _, code := range codes {
		// codes was built from the same mapping mapping, so each should have a rune
		r, _ := se.CharcodeToRune(code)
		font.charWidths[code] = font.fontMetrics[r].Wx
	}
}

func stdFontToSimpleFont(f fonts.StdFont) pdfFontSimple {
	l := f.Descriptor()
	return pdfFontSimple{
		fontCommon: fontCommon{
			subtype:  "Type1",
			basefont: f.Name(),
		},
		fontMetrics: f.GetMetricsTable(),
		std14Descriptor: &PdfFontDescriptor{
			FontName:    core.MakeName(string(l.Name)),
			FontFamily:  core.MakeName(l.Family),
			FontWeight:  core.MakeFloat(float64(l.Weight)),
			Flags:       core.MakeInteger(int64(l.Flags)),
			FontBBox:    core.MakeArrayFromFloats(l.BBox[:]),
			ItalicAngle: core.MakeFloat(l.ItalicAngle),
			Ascent:      core.MakeFloat(l.Ascent),
			Descent:     core.MakeFloat(l.Descent),
			CapHeight:   core.MakeFloat(l.CapHeight),
			XHeight:     core.MakeFloat(l.XHeight),
			StemV:       core.MakeFloat(l.StemV),
			StemH:       core.MakeFloat(l.StemH),
		},
		std14Encoder: f.Encoder(),
	}
}
