/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package context

import (
	"github.com/unidoc/unipdf/v3/internal/transform"
)

// TextState holds a representation of a PDF text state. The text state
// processes different text related operations which may occur in PDF content
// streams. It is used as a part of a renderding context in order to manipulate
// and display text.
type TextState struct {
	Tc  float64          // Character spacing.
	Tw  float64          // Word spacing.
	Th  float64          // Horizontal scaling.
	Tl  float64          // Leading.
	Tf  *TextFont        // Font
	Ts  float64          // Text rise.
	Tm  transform.Matrix // Text matrix.
	Tlm transform.Matrix // Text line matrix.
}

// NewTextState returns a new TextState instance.
func NewTextState() *TextState {
	return &TextState{
		Th:  100,
		Tm:  transform.IdentityMatrix(),
		Tlm: transform.IdentityMatrix(),
	}
}

// ProcTm processes a `Tm` operation, which sets the current text matrix.
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (ts *TextState) ProcTm(a, b, c, d, e, f float64) {
	ts.Tm = transform.NewMatrix(a, b, c, d, e, -f)
	ts.Tlm = ts.Tm.Clone()
}

// ProcTd processes a `Td` operation, which advances the text state to a new
// line with offsets `tx`,`ty`.
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (ts *TextState) ProcTd(tx, ty float64) {
	ts.Tlm.Concat(transform.TranslationMatrix(tx, -ty))
	ts.Tm = ts.Tlm.Clone()
}

// ProcTD processes a `TD` operation, which advances the text state to a new
// line with offsets `tx`,`ty`.
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (ts *TextState) ProcTD(tx, ty float64) {
	ts.Tl = -ty
	ts.ProcTd(tx, ty)
}

// ProcTStar processes a `T*` operation, which advances the text state to a
// new line.
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (ts *TextState) ProcTStar() {
	ts.ProcTd(0, -ts.Tl)
}

// ProcTj processes a `Tj` operation, which displays a text string.
//
// See section 9.4.3 "Text Showing Operators" and
// Table 209 (pp. 258-259 PDF32000_2008).
func (ts *TextState) ProcTj(data []byte, ctx Context) {
	tfs := ts.Tf.Size
	th := ts.Th / 100.0
	stateMatrix := transform.NewMatrix(tfs*th, 0, 0, tfs, 0, ts.Ts)

	runes := ts.Tf.CharcodesToUnicode(ts.Tf.BytesToCharcodes(data))
	for _, r := range runes {
		if r == '\x00' {
			continue
		}

		// Calculate text rendering matrix.
		tm := ts.Tm.Clone()
		ts.Tm.Concat(stateMatrix)

		// Draw rune.
		x, y := ts.Tm.Transform(0, 0)
		ctx.Scale(1, -1)
		ctx.DrawString(string(r), x, y)
		ctx.Scale(1, -1)

		// Calculate word spacing.
		tw := 0.0
		if r == ' ' {
			tw = ts.Tw
		}

		// Calculate rune spacing.
		var w float64
		if wX, _, ok := ts.Tf.GetRuneMetrics(r); ok {
			w = wX * 0.001 * tfs
		} else {
			w, _ = ctx.MeasureString(string(r))
		}

		// Calculate displacement offset.
		tx := (w + ts.Tc + tw) * th

		// Generate new text matrix.
		ts.Tm = transform.TranslationMatrix(tx, 0).Mult(tm)
	}
}

// ProcQ processes a `'` operation, which advances the text state to a new line
// and then displays a text string.
//
// See section 9.4.3 "Text Showing Operators" and
// Table 209 (pp. 258-259 PDF32000_2008).
func (ts *TextState) ProcQ(data []byte, ctx Context) {
	ts.ProcTStar()
	ts.ProcTj(data, ctx)
}

// ProcDQ processes a `''` operation, which advances the text state to a new
// line and then displays a text string using aw and ac as word and character
// spacing.
//
// See section 9.4.3 "Text Showing Operators" and
// Table 209 (pp. 258-259 PDF32000_2008).
func (ts *TextState) ProcDQ(data []byte, aw, ac float64, ctx Context) {
	ts.Tw = aw
	ts.Tc = ac
	ts.ProcQ(data, ctx)
}

// ProcTf processes a `Tf` operation which sets the font and its size.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (ts *TextState) ProcTf(font *TextFont) {
	ts.Tf = font
}

// Translate translates the current text matrix with `tx`,`ty`.
func (ts *TextState) Translate(tx, ty float64) {
	ts.Tm = transform.TranslationMatrix(tx, ty).Mult(ts.Tm)
}

// Reset resets both the text matrix and the line matrix.
func (ts *TextState) Reset() {
	ts.Tm = transform.IdentityMatrix()
	ts.Tlm = transform.IdentityMatrix()
}
