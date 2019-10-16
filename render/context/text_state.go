package context

import (
	"github.com/unidoc/unipdf/v3/internal/transform"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

type TextState struct {
	Tc  float64   // Character spacing.
	Tw  float64   // Word spacing.
	Th  float64   // Horizontal scaling.
	Tl  float64   // Leading.
	Tf  font.Face // Font.
	Tfs float64   // Font size.
	Ts  float64   // Text rise.

	Tm  transform.Matrix // Text matrix.
	Tlm transform.Matrix // Text line matrix.
}

func NewTextState() *TextState {
	return &TextState{
		Th:  100,
		Tm:  transform.IdentityMatrix(),
		Tlm: transform.IdentityMatrix(),
		Tf:  basicfont.Face7x13,
		Tfs: 13,
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

func (ts *TextState) DoTf(tf font.Face, tfs float64) {
	if tfs <= 1 {
		tfs = 13
	}
	ts.Tfs = tfs
	ts.Tf = tf
}

func (ts *TextState) Translate(tx, ty float64) {
	ts.Tm.Concat(transform.TranslationMatrix(tx, ty))
}

func (ts *TextState) ResetTm() {
	ts.Tm = transform.IdentityMatrix()
	ts.Tlm = transform.IdentityMatrix()
}
