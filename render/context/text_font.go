package context

import (
	"errors"

	"github.com/golang/freetype/truetype"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/image/font"
)

type TextFont struct {
	Font *model.PdfFont
	Face font.Face
	Size float64

	ttf *truetype.Font
}

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

func NewTextFontFromPath(filePath string, size float64) (*TextFont, error) {
	font, err := model.NewPdfFontFromTTFFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewTextFont(font, size)
}

func (tf *TextFont) WithSize(size float64) *TextFont {
	return &TextFont{
		Font: tf.Font,
		Face: truetype.NewFace(tf.ttf, &truetype.Options{Size: size}),
		Size: size,
		ttf:  tf.ttf,
	}
}
