package model

import (
	"errors"
	"io/ioutil"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/cmap"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

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

	// These fields are specific to simple PDF fonts.
	firstChar  int
	lastChar   int
	charWidths []float64
	encoder    textencoding.TextEncoder

	// Encoding is subject to limitations that are described in 9.6.6, "Character Encoding".
	// BaseFont is derived differently.
	FirstChar core.PdfObject
	LastChar  core.PdfObject
	Widths    core.PdfObject
	Encoding  core.PdfObject

	// Standard 14 fonts metrics
	fontMetrics map[string]fonts.CharMetrics
}

// pdfCIDFontType0FromSkeleton returns a pdfFontSimple with its common fields initalized.
func pdfFontSimpleFromSkeleton(base *fontCommon) *pdfFontSimple {
	return &pdfFontSimple{
		basefont:       base.basefont,
		subtype:        base.subtype,
		toUnicode:      base.toUnicode,
		toUnicodeCmap:  base.toUnicodeCmap,
		fontDescriptor: base.fontDescriptor,
		objectNumber:   base.objectNumber,
	}
}

// baseFields returns the fields of `font` that are common to all PDF fonts.
func (font *pdfFontSimple) baseFields() *fontCommon {
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
func (font *pdfFontSimple) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// SetEncoder sets the encoding for the underlying font.
func (font *pdfFontSimple) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

// GetGlyphCharMetrics returns the character metrics for the specified glyph.  A bool flag is
// returned to indicate whether or not the entry was found in the glyph to charcode mapping.
func (font pdfFontSimple) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	if font.fontMetrics != nil {
		metrics, ok := font.fontMetrics[glyph]
		return metrics, ok
	}

	metrics := fonts.CharMetrics{}

	code, found := font.encoder.GlyphToCharcode(glyph)
	if !found {
		return metrics, false
	}
	metrics.GlyphName = glyph

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

// newSimpleFontFromPdfObject creates a pdfFontSimple from dictionary `d`. Elements of `d` that
// are already parsed are contained in `base`.
// An error is returned if there is a problem with loading.
// !@#$ Just return a base 14 font, if obj is a base 14 font
//
// The value of Encoding is subject to limitations that are described in 9.6.6, "Character Encoding".
// • The value of BaseFont is derived differently.
//
// !@#$ 9.6.6.4 Encodings for TrueType Fonts (page 265)
//      Need to get TrueType font's cmap
func newSimpleFontFromPdfObject(d *core.PdfObjectDictionary, base *fontCommon, std14 bool) (*pdfFontSimple, error) {
	font := pdfFontSimpleFromSkeleton(base)

	// !@#$ Failing on ~/testdata/The-Byzantine-Generals-Problem.pdf
	// FirstChar is not defined in ~/testdata/shamirturing.pdf
	if !std14 {
		obj := d.Get("FirstChar")
		if obj == nil {
			obj = core.PdfObject(core.MakeInteger(0))
		}
		font.FirstChar = obj

		intVal, ok := core.TraceToDirectObject(obj).(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("ERROR: Invalid FirstChar type (%T)", obj)
			return nil, core.ErrTypeError
		}
		font.firstChar = int(*intVal)

		obj = d.Get("LastChar")
		if obj == nil {
			obj = core.PdfObject(core.MakeInteger(255))
		}
		font.LastChar = obj
		intVal, ok = core.TraceToDirectObject(obj).(*core.PdfObjectInteger)
		if !ok {
			common.Log.Debug("ERROR: Invalid LastChar type (%T)", obj)
			return nil, core.ErrTypeError
		}
		font.lastChar = int(*intVal)

		font.charWidths = []float64{}
		obj = d.Get("Widths")
		if obj != nil {
			font.Widths = obj

			arr, ok := core.TraceToDirectObject(obj).(*core.PdfObjectArray)
			if !ok {
				common.Log.Debug("ERROR: Widths attribute != array (%T)", obj)
				return nil, core.ErrTypeError
			}

			widths, err := arr.ToFloat64Array()
			if err != nil {
				common.Log.Debug("ERROR: converting widths to array")
				return nil, err
			}

			if len(widths) != (font.lastChar - font.firstChar + 1) {
				common.Log.Debug("ERROR: Invalid widths length != %d (%d)",
					font.lastChar-font.firstChar+1, len(widths))
				return nil, core.ErrRangeError
			}
			font.charWidths = widths
		}
	}

	font.Encoding = core.TraceToDirectObject(d.Get("Encoding"))
	return font, nil
}

// addEncoding adds the encoding to the font.
// The order of precedence is important
func (font *pdfFontSimple) addEncoding() error {
	var baseEncoder string
	var differences map[byte]string
	var err error
	if font.Encoding != nil {
		// !@#$ Stop setting default encoding in getFontEncoding XXX
		baseEncoder, differences, err = getFontEncoding(font.Encoding)
		if err != nil {
			common.Log.Debug("ERROR: BaseFont=%q Subtype=%q Encoding=%s (%T) err=%v", font.basefont,
				font.subtype, font.Encoding, font.Encoding, err)
			return err
		}
		base := font.baseFields()
		common.Log.Trace("addEncoding: BaseFont=%q Subtype=%q Encoding=%s (%T)", base.basefont,
			base.subtype, font.Encoding, font.Encoding)

		encoder, err := textencoding.NewSimpleTextEncoder(baseEncoder, differences)
		if err != nil {
			return err
		}
		font.SetEncoder(encoder)
	}

	if font.Encoder() == nil {
		descriptor := font.fontDescriptor
		if descriptor != nil {
			switch font.subtype {
			case "Type1":
				// XXX: !@#$ Is this the right order? Do the /Differences need to be reapplied?
				if descriptor.fontFile != nil && descriptor.fontFile.encoder != nil {
					common.Log.Debug("Using fontFile")
					font.SetEncoder(descriptor.fontFile.encoder)
				}
			case "TrueType":
				if descriptor.fontFile2 != nil {
					common.Log.Debug("Using FontFile2")
					encoder, err := descriptor.fontFile2.MakeEncoder()
					if err == nil {
						font.SetEncoder(encoder)
					}
				}
			}
		}
	}

	// At the end, apply the differences.
	if differences != nil {
		common.Log.Debug("differences=%+v font=%s", differences, font)
		if se, ok := font.Encoder().(textencoding.SimpleEncoder); ok {
			se.ApplyDifferences(differences)
			font.SetEncoder(se)
		}
	}
	return nil
}

// getFontEncoding returns font encoding of `obj` the "Encoding" entry in a font dict
// Table 114 – Entries in an encoding dictionary (page 263)
// 9.6.6.1 General (page 262)
// A font’s encoding is the association between character codes (obtained from text strings that
// are shown) and glyph descriptions. This sub-clause describes the character encoding scheme used
// with simple PDF fonts. Composite fonts (Type 0) use a different character mapping algorithm, as
// discussed in 9.7, "Composite Fonts".
// Except for Type 3 fonts, every font program shall have a built-in encoding. Under certain
// circumstances, a PDF font dictionary may change the encoding used with the font program to match
// the requirements of the conforming writer generating the text being shown.
func getFontEncoding(obj core.PdfObject) (string, map[byte]string, error) {
	baseName := "StandardEncoding"

	if obj == nil {
		// Fall back to StandardEncoding
		return baseName, nil, nil
	}

	switch encoding := obj.(type) {
	case *core.PdfObjectName:
		return string(*encoding), nil, nil
	case *core.PdfObjectDictionary:
		typ, err := core.GetName(core.TraceToDirectObject(encoding.Get("Type")))
		if err == nil && typ == "Encoding" {
			base, err := core.GetName(core.TraceToDirectObject(encoding.Get("BaseEncoding")))
			if err == nil {
				baseName = base
			}
		}
		diffList, err := core.GetArray(core.TraceToDirectObject(encoding.Get("Differences")))
		if err != nil {
			common.Log.Debug("ERROR: Bad font encoding dict=%+v err=%v", encoding, err)
			return "", nil, core.ErrTypeError
		}

		differences, err := textencoding.FromFontDifferences(diffList)
		return baseName, differences, err
	default:
		common.Log.Debug("ERROR: Encoding not a name or dict (%T) %s", obj, obj.String())
		return "", nil, core.ErrTypeError
	}
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
		d.Set("Encoding", font.encoder.ToPdfObject())
	}

	return font.container
}

// NewPdfFontFromTTFFile loads a TTF font and returns a PdfFont type that can be used in text
// styling functions.
// Uses a WinAnsiTextEncoder and loads only character codes 32-255.
func NewPdfFontFromTTFFile(filePath string) (*PdfFont, error) {
	const minCode = 32
	const maxCode = 255

	ttf, err := fonts.TtfParse(filePath)
	if err != nil {
		common.Log.Debug("ERROR: loading ttf font: %v", err)
		return nil, err
	}

	truefont := &pdfFontSimple{subtype: "TrueType"}

	// TODO: Make more generic to allow customization... Need to know which glyphs are to be used,
	// then can derive
	// TODO: Subsetting fonts.
	truefont.encoder = textencoding.NewWinAnsiTextEncoder()
	truefont.firstChar = minCode
	truefont.lastChar = maxCode

	truefont.basefont = ttf.PostScriptName
	truefont.FirstChar = core.MakeInteger(minCode)
	truefont.LastChar = core.MakeInteger(maxCode)

	k := 1000.0 / float64(ttf.UnitsPerEm)
	if len(ttf.Widths) <= 0 {
		return nil, errors.New("ERROR: Missing required attribute (Widths)")
	}

	missingWidth := k * float64(ttf.Widths[0])

	vals := make([]float64, 0, maxCode-minCode+1)
	for code := minCode; code <= maxCode; code++ {
		r, found := truefont.Encoder().CharcodeToRune(uint16(code))
		if !found {
			common.Log.Debug("Rune not found (code: %d)", code)
			vals = append(vals, missingWidth)
			continue
		}

		pos, ok := ttf.Chars[uint16(r)]
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
		common.Log.Debug("ERROR: Invalid length of widths, %d < %d", len(vals), 255-32+1)
		return nil, core.ErrRangeError
	}

	truefont.charWidths = vals[:255-32+1]

	// Use WinAnsiEncoding by default.
	truefont.Encoding = core.MakeName("WinAnsiEncoding")

	descriptor := &PdfFontDescriptor{}
	descriptor.Ascent = core.MakeFloat(k * float64(ttf.TypoAscender))
	descriptor.Descent = core.MakeFloat(k * float64(ttf.TypoDescender))
	descriptor.CapHeight = core.MakeFloat(k * float64(ttf.CapHeight))
	descriptor.FontBBox = core.MakeArrayFromFloats([]float64{k * float64(ttf.Xmin),
		k * float64(ttf.Ymin), k * float64(ttf.Xmax), k * float64(ttf.Ymax)})
	descriptor.ItalicAngle = core.MakeFloat(float64(ttf.ItalicAngle))
	descriptor.MissingWidth = core.MakeFloat(k * float64(ttf.Widths[0]))

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
	truefont.fontDescriptor = descriptor

	font := &PdfFont{
		context: truefont,
	}

	return font, nil
}

var standard14Fonts = map[string]pdfFontSimple{
	"Courier": pdfFontSimple{subtype: "Type1",
		basefont:    "Courier",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.CourierCharMetrics,
	},
	"Courier-Bold": pdfFontSimple{subtype: "Type1",
		basefont:    "Courier-Bold",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.CourierBoldCharMetrics,
	},
	"Courier-BoldOblique": pdfFontSimple{subtype: "Type1",
		basefont:    "Courier-BoldOblique",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.CourierBoldObliqueCharMetrics,
	},
	"Courier-Oblique": pdfFontSimple{subtype: "Type1",
		basefont:    "Courier-Oblique",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.CourierObliqueCharMetrics,
	},
	"Helvetica": pdfFontSimple{subtype: "Type1",
		basefont:    "Helvetica",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.HelveticaCharMetrics,
	},
	"Helvetica-Bold": pdfFontSimple{subtype: "Type1",
		basefont:    "Helvetica-Bold",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.HelveticaBoldCharMetrics,
	},
	"Helvetica-BoldOblique": pdfFontSimple{subtype: "Type1",
		basefont:    "Helvetica-BoldOblique",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.HelveticaBoldObliqueCharMetrics,
	},
	"Helvetica-Oblique": pdfFontSimple{subtype: "Type1",
		basefont:    "Helvetica-Oblique",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.HelveticaObliqueCharMetrics,
	},
	"Times-Roman": pdfFontSimple{subtype: "Type1",
		basefont:    "Times-Roman",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.TimesRomanCharMetrics,
	},
	"Times-Bold": pdfFontSimple{subtype: "Type1",
		basefont:    "Times-Bold",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.TimesBoldCharMetrics,
	},
	"Times-BoldItalic": pdfFontSimple{subtype: "Type1",
		basefont:    "Times-BoldItalic",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.TimesBoldItalicCharMetrics,
	},
	"Times-Italic": pdfFontSimple{subtype: "Type1",
		basefont:    "Times-Italic",
		encoder:     textencoding.NewWinAnsiTextEncoder(),
		fontMetrics: fonts.TimesItalicCharMetrics,
	},
	"Symbol": pdfFontSimple{subtype: "Type1",
		basefont:    "Symbol",
		encoder:     textencoding.NewSymbolEncoder(),
		fontMetrics: fonts.SymbolCharMetrics,
	},
	"ZapfDingbats": pdfFontSimple{subtype: "Type1",
		basefont:    "ZapfDingbats",
		encoder:     textencoding.NewZapfDingbatsEncoder(),
		fontMetrics: fonts.ZapfDingbatsCharMetrics,
	},
}
