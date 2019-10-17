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

func (tf *TextFont) Clone(size float64) *TextFont {
	return &TextFont{
		Font: tf.Font,
		Face: truetype.NewFace(tf.ttf, &truetype.Options{Size: size}),
		Size: size,
		ttf:  tf.ttf,
	}
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

func (ts *TextState) DoTj(data []byte, ctx Context) {
	font := ts.Tf.Font
	tfs := ts.Tf.Size
	th := ts.Th / 100.0
	stateMatrix := transform.NewMatrix(tfs*th, 0, 0, tfs, 0, ts.Ts)

	charcodes := font.BytesToCharcodes(data)
	runes := font.CharcodesToUnicode(charcodes)
	for i, r := range runes {
		if r == '\x00' {
			continue
		}

		// Calculate text rendering matrix.
		tm := ts.Tm.Clone()
		ts.Tm.Concat(stateMatrix)

		// Draw rune.
		x, y := ts.Tm.Transform(0, 0)
		ctx.DrawString(string(r), x, y)

		// Calculate word spacing.
		tw := 0.0
		if r == ' ' {
			tw = ts.Tw
		}

		// Calculate rune spacing.
		var w float64
		if m, ok := font.GetCharMetrics(charcodes[i]); ok {
			w = m.Wx * 0.001 * tfs
		} else {
			w, _ = ctx.MeasureString(string(r))
		}

		// Calculate displacement offset.
		tx := (w + ts.Tc + tw) * th

		// Generate new text matrix.
		ts.Tm = tm
		ts.Translate(tx, 0)
	}
}

func (ts *TextState) DoQuote(data []byte, ctx Context) {
	ts.DoTStar()
	ts.DoTj(data, ctx)
}

func (ts *TextState) DoQuotes(data []byte, aw, ac float64, ctx Context) {
	ts.Tw = aw
	ts.Tc = ac
	ts.DoQuote(data, ctx)
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
