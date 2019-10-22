package context

import (
	"github.com/unidoc/unipdf/v3/internal/transform"
)

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

func (ts *TextState) ProcTm(a, b, c, d, e, f float64) {
	ts.Tm = transform.NewMatrix(a, b, c, d, e, -f)
	ts.Tlm = ts.Tm.Clone()
}

func (ts *TextState) ProcTd(tx, ty float64) {
	ts.Tlm.Concat(transform.NewMatrix(1, 0, 0, 1, tx, -ty))
	ts.Tm = ts.Tlm.Clone()
}

func (ts *TextState) ProcTD(tx, ty float64) {
	ts.Tl = -ty
	ts.ProcTd(tx, ty)
}

func (ts *TextState) ProcTStar() {
	ts.ProcTd(0, -ts.Tl)
}

func (ts *TextState) ProcTj(data []byte, ctx Context) {
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

func (ts *TextState) ProcQ(data []byte, ctx Context) {
	ts.ProcTStar()
	ts.ProcTj(data, ctx)
}

func (ts *TextState) ProcDQ(data []byte, aw, ac float64, ctx Context) {
	ts.Tw = aw
	ts.Tc = ac
	ts.ProcQ(data, ctx)
}

func (ts *TextState) ProcTf(font *TextFont) {
	ts.Tf = font
}

func (ts *TextState) Translate(tx, ty float64) {
	ts.Tm.Concat(transform.TranslationMatrix(tx, ty))
}

func (ts *TextState) ResetTm() {
	ts.Tm = transform.IdentityMatrix()
	ts.Tlm = transform.IdentityMatrix()
}
