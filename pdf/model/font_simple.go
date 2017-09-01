package model

import (
	"errors"
	"io/ioutil"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// pdfFontTrueType represents a simple TrueType font.
type pdfFontTrueType struct {
	encoder textencoding.TextEncoder

	firstChar  int
	lastChar   int
	charWidths []float64

	BaseFont       core.PdfObject
	FirstChar      core.PdfObject
	LastChar       core.PdfObject
	Widths         core.PdfObject
	FontDescriptor *PdfFontDescriptor
	Encoding       core.PdfObject
	ToUnicode      core.PdfObject

	container *core.PdfIndirectObject
}

// Encoder returns the font's text encoder.
func (font pdfFontTrueType) Encoder() textencoding.TextEncoder {
	return font.encoder
}

// SetEncoder sets the encoder for the truetype font.
func (font pdfFontTrueType) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

// GetGlyphCharMetrics returns the character metrics for the specified glyph.  A bool flag is returned to
// indicate whether or not the entry was found in the glyph to charcode mapping.
func (font pdfFontTrueType) GetGlyphCharMetrics(glyph string) (fonts.CharMetrics, bool) {
	metrics := fonts.CharMetrics{}

	code, found := font.Encoder().GlyphToCharcode(glyph)
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

// newPdfFontTrueTypeFromPdfObject creates a pdfFontTrueType from a dictionary. An error is returned if there is
// a problem with loading.
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

	if obj := d.Get("Type"); obj != nil {
		oname, is := obj.(*core.PdfObjectName)
		if !is || oname.String() != "Font" {
			common.Log.Debug("Incompatibility: Type defined but not Font")
		}
	}

	if obj := d.Get("Subtype"); obj != nil {
		oname, is := obj.(*core.PdfObjectName)
		if !is || oname.String() != "TrueType" {
			common.Log.Debug("Incompatibility: Loading TrueType font but Subtype != TrueType")
		}
	}

	font.BaseFont = d.Get("BaseFont")

	if obj := d.Get("FirstChar"); obj != nil {
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

	if obj := d.Get("LastChar"); obj != nil {
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
	if obj := d.Get("Widths"); obj != nil {
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

	if obj := d.Get("FontDescriptor"); obj != nil {
		descriptor, err := newPdfFontDescriptorFromPdfObject(obj)
		if err != nil {
			common.Log.Debug("Error loading font descriptor: %v", err)
			return nil, err
		}

		font.FontDescriptor = descriptor
	}

	font.Encoding = d.Get("Encoding")
	font.ToUnicode = d.Get("ToUnicode")

	return font, nil
}

// ToPdfObject converts the pdfFontTrueType to its PDF representation for outputting.
func (this *pdfFontTrueType) ToPdfObject() core.PdfObject {
	if this.container == nil {
		this.container = &core.PdfIndirectObject{}
	}
	d := core.MakeDict()
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

// NewPdfFontFromTTFFile loads a TTF font and returns a PdfFont type that can be used in text styling functions.
// Uses a WinAnsiTextEncoder and loads only character codes 32-255.
func NewPdfFontFromTTFFile(filePath string) (*PdfFont, error) {
	ttf, err := fonts.TtfParse(filePath)
	if err != nil {
		common.Log.Debug("Error loading ttf font: %v", err)
		return nil, err
	}

	truefont := &pdfFontTrueType{}

	// TODO: Make more generic to allow customization... Need to know which glyphs are to be used, then can derive
	// TODO: needed encoding via a BaseEncoding and a Differences entry if needed.
	// TODO: Subsetting fonts.
	truefont.encoder = textencoding.NewWinAnsiTextEncoder()
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
		runeVal, found := truefont.Encoder().CharcodeToRune(uint16(charcode))
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

	// Use WinAnsiEncoding by default.
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
