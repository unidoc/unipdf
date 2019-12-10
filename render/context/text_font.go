package context

import (
	"errors"

	"github.com/golang/freetype/truetype"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/textencoding"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/image/font"
)

// TextFont represents a font used to draw text to a target, through a
// rendering context.
type TextFont struct {
	Font *model.PdfFont
	Face font.Face
	Size float64

	ttf      *truetype.Font
	origFont *model.PdfFont
}

// NewTextFont returns a new text font instance based on the specified PDF font
// and the specified font size.
func NewTextFont(font *model.PdfFont, size float64) (*TextFont, error) {
	descriptor := font.FontDescriptor()
	if descriptor == nil {
		return nil, errors.New("could not get font descriptor")
	}

	fontStream, ok := core.GetStream(descriptor.FontFile2)
	if !ok {
		return nil, errors.New("missing font file stream")
	}

	fontData, err := core.DecodeStream(fontStream)
	if err != nil {
		return nil, err
	}

	ttfFont, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	if size <= 1 {
		size = 10
	}

	return &TextFont{
		Font: font,
		Face: truetype.NewFace(ttfFont, &truetype.Options{Size: size}),
		Size: size,
		ttf:  ttfFont,
	}, nil
}

// NewTextFontFromPath returns a new text font instance based on the specified
// font file and the specified font size.
func NewTextFontFromPath(filePath string, size float64) (*TextFont, error) {
	font, err := model.NewPdfFontFromTTFFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewTextFont(font, size)
}

// WithSize returns a new text font instance based on the current text font,
// with the specified font size.
func (tf *TextFont) WithSize(size float64, originalFont *model.PdfFont) *TextFont {
	if size <= 1 {
		size = 10
	}

	return &TextFont{
		Font:     tf.Font,
		Face:     truetype.NewFace(tf.ttf, &truetype.Options{Size: size}),
		Size:     size,
		ttf:      tf.ttf,
		origFont: originalFont,
	}
}

// BytesToCharcodes converts the specified byte data to character codes, using
// the encapsulated PDF font instance.
func (tf *TextFont) BytesToCharcodes(data []byte) []textencoding.CharCode {
	if tf.origFont != nil {
		return tf.origFont.BytesToCharcodes(data)
	}

	return tf.Font.BytesToCharcodes(data)
}

// CharcodesToUnicode converts the specified character codes to a slice of
// runes, using the encapsulated PDF font instance.
func (tf *TextFont) CharcodesToUnicode(charcodes []textencoding.CharCode) []rune {
	if tf.origFont != nil {
		return tf.origFont.CharcodesToUnicode(charcodes)
	}

	return tf.Font.CharcodesToUnicode(charcodes)
}

// GetCharMetrics returns the metrics of the specified character code. The
// character metrics are calculated by the internal PDF font.
func (tf *TextFont) GetCharMetrics(code textencoding.CharCode) (float64, float64, bool) {
	if tf.origFont != nil {
		if metrics, ok := tf.origFont.GetCharMetrics(code); ok && metrics.Wx != 0 {
			return metrics.Wx, metrics.Wy, true
		}
	}

	metrics, ok := tf.Font.GetCharMetrics(code)
	return metrics.Wx, metrics.Wy, ok
}
