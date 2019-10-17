package context

import (
	"errors"

	"github.com/golang/freetype/truetype"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/transform"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/image/font"
)

type TextFont struct {
	Font *model.PdfFont
	Face font.Face
	Size float64
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
	}, nil
}

func NewTextFontFromPath(filePath string, size float64) (*TextFont, error) {
	font, err := model.NewPdfFontFromTTFFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewTextFont(font, size)
}

type TextState struct {
	Tc float64   // Character spacing.
	Tw float64   // Word spacing.
	Th float64   // Horizontal scaling.
	Tl float64   // Leading.
	Tf *TextFont // Font
	Ts float64   // Text rise.

	Tm  transform.Matrix // Text matrix.
	Tlm transform.Matrix // Text line matrix.
}

func NewTextState() *TextState {
	return &TextState{
		Th:  100,
		Tm:  transform.IdentityMatrix(),
		Tlm: transform.IdentityMatrix(),
	}
}

func (ts *TextState) DoTm(a, b, c, d, e, f float64) {
	ts.Tm = transform.NewMatrix(a, b, c, d, e, -f)
	ts.Tlm = ts.Tm.Clone()
}

func (ts *TextState) DoTd(tx, ty float64) {
	ts.Tlm.Concat(transform.NewMatrix(1, 0, 0, 1, tx, -ty))
	ts.Tm = ts.Tlm.Clone()
}

func (ts *TextState) DoTD(tx, ty float64) {
	ts.Tl = -ty
	ts.DoTd(tx, ty)
}

func (ts *TextState) DoTStar() {
	ts.DoTd(0, -ts.Tl)
}

func (ts *TextState) DoTj(text string, ctx Context) {
	// TODO:
	// Account for encoding.
	// Account for Tc.
	// Account for Tw.
	ctx.DrawString(text, 0, 0)
	w, _ := ctx.MeasureString(text)
	ts.Translate(w, 0)
}

func (ts *TextState) DoQuote(text string, ctx Context) {
	ts.DoTStar()
	ts.DoTj(text, ctx)
}

func (ts *TextState) DoQuotes(text string, aw, ac float64, ctx Context) {
	ts.Tw = aw
	ts.Tc = ac
	ts.DoQuote(text, ctx)
}

func (ts *TextState) DoTf(font *TextFont) {
	ts.Tf = font
}

func (ts *TextState) Translate(tx, ty float64) {
	ts.Tm.Concat(transform.TranslationMatrix(tx, ty))
}

func (ts *TextState) ResetTm() {
	ts.Tm = transform.IdentityMatrix()
	ts.Tlm = transform.IdentityMatrix()
}
