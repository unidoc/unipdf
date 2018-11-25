/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// ExtractText processes and extracts all text data in content streams and returns as a string.
// It takes into account character encodings in the PDF file, which are decoded by
// CharcodeBytesToUnicode.
// Characters that can't be decoded are replaced with MissingCodeRune ('\ufffd' = �).
func (e *Extractor) ExtractText() (string, error) {
	text, _, _, err := e.ExtractText2()
	return text, err
}

// ExtractText2 works like ExtractText but returns the number of characters in the output and the
// the number of characters that were not decoded.
func (e *Extractor) ExtractText2() (string, int, int, error) {
	textList, numChars, numMisses, err := e.ExtractXYText()
	if err != nil {
		return "", numChars, numMisses, err
	}
	return textList.ToText(), numChars, numMisses, nil
}

// ExtractXYText returns the text contents of `e` as a TextList.
func (e *Extractor) ExtractXYText() (*TextList, int, int, error) {
	textList := &TextList{}
	state := newTextState()
	fontStack := fontStacker{}
	var to *textObject

	cstreamParser := contentstream.NewContentStreamParser(e.contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		common.Log.Debug("ExtractXYText: parse failed. err=%v", err)
		return textList, state.numChars, state.numMisses, err
	}

	processor := contentstream.NewContentStreamProcessor(*operations)

	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState,
			resources *model.PdfPageResources) error {


			operand := op.Operand

			switch operand {
			case "q":
				if !fontStack.empty() {
					common.Log.Trace("Save font state: %s\n%s",
						fontStack.peek(), fontStack.String())
					fontStack.push(fontStack.peek())
				}
				if state.Tf != nil {
					common.Log.Trace("Save font state: %s\n->%s\n%s",
						fontStack.peek(), state.Tf, fontStack.String())
					fontStack.push(state.Tf)
				}
			case "Q":
				if !fontStack.empty() {
					common.Log.Trace("Restore font state: %s\n->%s\n%s",
						fontStack.peek(), fontStack.get(-2), fontStack.String())
					fontStack.pop()
				}
				if len(fontStack) >= 2 {
					common.Log.Trace("Restore font state: %s\n->%s\n%s",
						state.Tf, fontStack.peek(), fontStack.String())
					state.Tf = fontStack.pop()
				}
			case "BT": // Begin text
				// Begin a text object, initializing the text matrix, Tm, and the text line matrix,
				// Tlm, to the identity matrix. Text objects shall not be nested; a second BT shall
				// not appear before an ET.
				if to != nil {
					common.Log.Debug("BT called while in a text object")
				}
				to = newTextObject(e, gs, &state, &fontStack)
			case "ET": // End Text
				*textList = append(*textList, to.Texts...)
				to = nil
			case "T*": // Move to start of next text line
				to.nextLine()
			case "Td": // Move text location
				if ok, err := to.checkOp(op, 2, true); !ok {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				x, y, err := toFloatXY(op.Params)
				if err != nil {
					return err
				}
				to.moveText(x, y)
			case "TD": // Move text location and set leading.
				if ok, err := to.checkOp(op, 2, true); !ok {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				x, y, err := toFloatXY(op.Params)
				if err != nil {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				to.moveTextSetLeading(x, y)
			case "Tj": // Show text
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: Tj op=%s err=%v", op, err)
					return err
				}
				charcodes, ok := core.GetStringBytes(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: Tj op=%s GetStringBytes failed", op)
					return core.ErrTypeError
				}
				return to.showText(charcodes)
			case "TJ": // Show text with adjustable spacing
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: TJ err=%v", err)
					return err
				}
				args, ok := core.GetArray(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: Tj op=%s GetArrayVal failed", op)
					return err
				}
				return to.showTextAdjusted(args)
			case "'": // Move to next line and show text
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: ' err=%v", err)
					return err
				}
				charcodes, ok := core.GetStringBytes(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: ' op=%s GetStringBytes failed", op)
					return core.ErrTypeError
				}
				to.nextLine()
				return to.showText(charcodes)
			case `"`: // Set word and character spacing, move to next line, and show text
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: \" err=%v", err)
					return err
				}
				x, y, err := toFloatXY(op.Params[:2])
				if err != nil {
					return err
				}
				charcodes, ok := core.GetStringBytes(op.Params[2])
				if !ok {
					common.Log.Debug("ERROR: \" op=%s GetStringBytes failed", op)
					return core.ErrTypeError
				}
				to.setCharSpacing(x)
				to.setWordSpacing(y)
				to.nextLine()
				return to.showText(charcodes)
			case "TL": // Set text leading
				y, err := floatParam(op)
				if err != nil {
					common.Log.Debug("ERROR: TL err=%v", err)
					return err
				}
				to.setTextLeading(y)
			case "Tc": // Set character spacing
				y, err := floatParam(op)
				if err != nil {
					common.Log.Debug("ERROR: Tc err=%v", err)
					return err
				}
				to.setCharSpacing(y)
			case "Tf": // Set font
				if to == nil {
					// This is needed for ~/testdata/26-Hazard-Thermal-environment.pdf
					to = newTextObject(e, gs, &state, &fontStack)
				}
				if ok, err := to.checkOp(op, 2, true); !ok {
					common.Log.Debug("ERROR: Tf err=%v", err)
					return err
				}
				name, ok := core.GetNameVal(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: Tf op=%s GetNameVal failed", op)
					return core.ErrTypeError
				}
				size, err := core.GetNumberAsFloat(op.Params[1])
				if !ok {
					common.Log.Debug("ERROR: Tf op=%s GetFloatVal failed. err=%v", op, err)
					return err
				}
				err = to.setFont(name, size)
				if err != nil {
					return err
				}
			case "Tm": // Set text matrix
				if ok, err := to.checkOp(op, 6, true); !ok {
					common.Log.Debug("ERROR: Tm err=%v", err)
					return err
				}
				floats, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				to.setTextMatrix(floats)
			case "Tr": // Set text rendering mode
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: Tr err=%v", err)
					return err
				}
				mode, ok := core.GetIntVal(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: Tr op=%s GetIntVal failed", op)
					return core.ErrTypeError
				}
				to.setTextRenderMode(mode)
			case "Ts": // Set text rise
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: Ts err=%v", err)
					return err
				}
				y, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				to.setTextRise(y)
			case "Tw": // Set word spacing
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				y, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					common.Log.Debug("ERROR: err=%v", err)
					return err

				}
				to.setWordSpacing(y)
			case "Tz": // Set horizontal scaling
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				y, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					common.Log.Debug("ERROR: err=%v", err)
					return err
				}
				to.setHorizScaling(y)
			}

			return nil
		})

	err = processor.Process(e.resources)
	if err != nil {
		common.Log.Debug("ERROR: Processing: err=%v", err)
	}
	return textList, state.numChars, state.numMisses, err
}

//
// Text operators
//

// moveText "Td" Moves start of text by `tx`,`ty`.
// Move to the start of the next line, offset from the start of the current line by (tx, ty).
// tx and ty are in unscaled text space units.
func (to *textObject) moveText(tx, ty float64) {
	to.moveTo(tx, ty)
}

// moveTextSetLeading "TD" Move text location and set leading.
// Move to the start of the next line, offset from the start of the current line by (tx, ty). As a
// side effect, this operator shall set the leading parameter in the text state. This operator shall
// have the same effect as this code:
//  −ty TL
//  tx ty Td
func (to *textObject) moveTextSetLeading(tx, ty float64) {
	to.State.Tl = -ty
	to.moveTo(tx, ty)
}

// nextLine "T*"" Moves start of text `Line` to next text line
// Move to the start of the next line. This operator has the same effect as the code
//    0 -Tl Td
// where Tl denotes the current leading parameter in the text state. The negative of Tl is used
// here because Tl is the text leading expressed as a positive number. Going to the next line
// entails decreasing the y coordinate. (page 250)
func (to *textObject) nextLine() {
	to.moveTo(0, -to.State.Tl)
}

// setTextMatrix "Tm".
// Set the text matrix, Tm, and the text line matrix, Tlm to the Matrix specified by the 6 numbers
// in `f`  (page 250)
func (to *textObject) setTextMatrix(f []float64) {
	a, b, c, d, tx, ty := f[0], f[1], f[2], f[3], f[4], f[5]
	to.Tm = contentstream.NewMatrix(a, b, c, d, tx, ty)
	to.Tlm = contentstream.NewMatrix(a, b, c, d, tx, ty)
	common.Log.Debug("setTextMatrix:Tm=%s", to.Tm)
}

// showText "Tj" Show a text string.
func (to *textObject) showText(charcodes []byte) error {
	return to.renderText(charcodes)
}

// showTextAdjusted "TJ" Show text with adjustable spacing.
func (to *textObject) showTextAdjusted(args *core.PdfObjectArray) error {
	vertical := false
	for _, o := range args.Elements() {
		switch o.(type) {
		case *core.PdfObjectFloat, *core.PdfObjectInteger:
			// The following is supposed to be equivalent to the existing Unidoc implementation.
			x, err := core.GetNumberAsFloat(o)
			if err != nil {
				common.Log.Debug("showTextAdjusted: Bad numerical arg. o=%s args=%+v", o, args)
				return err
			}
			dx, dy := -x*0.001*to.State.Tfs, 0.0
			if vertical {
				dy, dx = dx, dy
			}
			td := translationMatrix(Point{X: dx, Y: dy})
			to.Tm = td.Mult(to.Tm)
			common.Log.Debug("showTextAdjusted: dx,dy=%3f,%.3f Tm=%s", dx, dy, to.Tm)
		case *core.PdfObjectString:
			charcodes, ok := core.GetStringBytes(o)
			if !ok {
				common.Log.Debug("showTextAdjusted: Bad string arg. o=%s args=%+v", o, args)
				return core.ErrTypeError
			}
			to.renderText(charcodes)
		default:
			common.Log.Debug("showTextAdjusted. Unexpected type (%T) args=%+v", o, args)
			return core.ErrTypeError
		}
	}
	return nil
}

// setTextLeading "TL" Set text leading.
func (to *textObject) setTextLeading(y float64) {
	if to == nil {
		return
	}
	to.State.Tl = y
}

// setCharSpacing "Tc" Set character spacing.
func (to *textObject) setCharSpacing(x float64) {
	if to == nil {
		return
	}
	to.State.Tc = x
}

// setFont "Tf" Set font.
func (to *textObject) setFont(name string, size float64) error {
	if to == nil {
		return nil
	}
	font, err := to.getFont(name)
	if err == nil {
		to.State.Tf = font
		if len(*to.fontStack) == 0 {
			to.fontStack.push(font)
		} else {
			(*to.fontStack)[len(*to.fontStack)-1] = font
		}
	} else if err == model.ErrFontNotSupported {
		// XXX: Do we need to handle this case in a special way?
		return err
	} else {
		return err
	}
	to.State.Tfs = size
	return nil
}

// setTextRenderMode "Tr" Set text rendering mode.
func (to *textObject) setTextRenderMode(mode int) {
	if to == nil {
		return
	}
	to.State.Tmode = RenderMode(mode)
}

// setTextRise "Ts" Set text rise.
func (to *textObject) setTextRise(y float64) {
	if to == nil {
		return
	}
	to.State.Trise = y
}

// setWordSpacing "Tw" Set word spacing.
func (to *textObject) setWordSpacing(y float64) {
	if to == nil {
		return
	}
	to.State.Tw = y
}

// setHorizScaling "Tz" Set horizontal scaling.
func (to *textObject) setHorizScaling(y float64) {
	if to == nil {
		return
	}
	to.State.Th = y
}

// floatParam returns the single float parameter of operatr `op`, or an error if it doesn't have
// a single float parameter or we aren't in a text stream.
func floatParam(op *contentstream.ContentStreamOperation) (float64, error) {
	if len(op.Params) != 1 {
		err := errors.New("incorrect parameter count")
		common.Log.Debug("ERROR: %#q should have %d input params, got %d %+v",
			op.Operand, 1, len(op.Params), op.Params)
		return 0.0, err
	}
	return core.GetNumberAsFloat(op.Params[0])
}

// checkOp returns true if we are in a text stream and `op` has `numParams` params.
// If `hard` is true and the number of params don't match, an error is returned.
func (to *textObject) checkOp(op *contentstream.ContentStreamOperation, numParams int,
	hard bool) (ok bool, err error) {
	if to == nil {
		params := []core.PdfObject{}
		if numParams > 0 {
			params = op.Params
			if len(params) > numParams {
				params = params[:numParams]
			}
		}
		common.Log.Debug("%#q operand outside text. params=%+v", op.Operand, params)
	}
	if numParams >= 0 {
		if len(op.Params) != numParams {
			if hard {
				err = errors.New("incorrect parameter count")
			}
			common.Log.Debug("ERROR: %#q should have %d input params, got %d %+v",
				op.Operand, numParams, len(op.Params), op.Params)
			return false, err
		}
	}
	return true, nil
}

// fontStacker is the PDF font stack implementation.
type fontStacker []*model.PdfFont

// String returns a string describing the current state of the font stack.
func (fontStack *fontStacker) String() string {
	parts := []string{"---- font stack"}
	for i, font := range *fontStack {
		s := "<nil>"
		if font != nil {
			s = font.String()
		}
		parts = append(parts, fmt.Sprintf("\t%2d: %s", i, s))
	}
	return strings.Join(parts, "\n")
}

// push pushes `font` onto the font stack.
func (fontStack *fontStacker) push(font *model.PdfFont) {
	*fontStack = append(*fontStack, font)
}

// pop pops and returns the element on the top of the font stack if there is one or nil if there isn't.
func (fontStack *fontStacker) pop() *model.PdfFont {
	if fontStack.empty() {
		return nil
	}
	font := (*fontStack)[len(*fontStack)-1]
	*fontStack = (*fontStack)[:len(*fontStack)-1]
	return font
}

// peek returns the element on the top of the font stack if there is one or nil if there isn't.
func (fontStack *fontStacker) peek() *model.PdfFont {
	if fontStack.empty() {
		return nil
	}
	return (*fontStack)[len(*fontStack)-1]
}

// get returns the `idx`'th element of the font stack if there is one or nil if there isn't.
//  idx = 0: bottom of font stack
//  idx = len(fontstack) - 1: top of font stack
//  idx = -n is same as dx = len(fontstack) - n, so fontstack.get(-1) is same as fontstack.peek()
func (fontStack *fontStacker) get(idx int) *model.PdfFont {
	if idx < 0 {
		idx += fontStack.size()
	}
	if idx < 0 || idx > fontStack.size()-1 {
		return nil
	}
	return (*fontStack)[idx]
}

// empty returns true if the font stack is empty.
func (fontStack *fontStacker) empty() bool {
	return len(*fontStack) == 0
}

// size returns the number of elements in the font stack.
func (fontStack *fontStacker) size() int {
	return len(*fontStack)
}

// 9.3 Text State Parameters and Operators (page 243)
// Some of these parameters are expressed in unscaled text space units. This means that they shall
// be specified in a coordinate system that shall be defined by the text matrix, Tm but shall not be
// scaled by the font size parameter, Tfs.

// textState represents the text state.
type textState struct {
	Tc    float64        // Character spacing. Unscaled text space units.
	Tw    float64        // Word spacing. Unscaled text space units.
	Th    float64        // Horizontal scaling.
	Tl    float64        // Leading. Unscaled text space units. Used by TD,T*,'," see Table 108.
	Tfs   float64        // Text font size.
	Tmode RenderMode     // Text rendering mode.
	Trise float64        // Text rise. Unscaled text space units. Set by Ts.
	Tf    *model.PdfFont // Text font.
	// For debugging
	numChars  int
	numMisses int
}

// 9.4.1 General (page 248)
// A PDF text object consists of operators that may show text strings, move the text position, and
// set text state and certain other parameters. In addition, two parameters may be specified only
// within a text object and shall not persist from one text object to the next:
//   • Tm, the text matrix
//   • Tlm, the text line matrix
//
// Text space is converted to device space by this transform (page 252)
// Trm is the text rendering matrix
//        | Tfs x Th   0      0 |
// Trm  = | 0         Tfs     0 | × Tm × CTM
//        | 0         Trise   1 |
// This corresponds to the following code in renderText()
//  trm := stateMatrix.Mult(to.Tm).Mult(to.gs.CTM))

// textObject represents a PDF text object.
type textObject struct {
	e         *Extractor
	gs        contentstream.GraphicsState
	fontStack *fontStacker
	State     *textState
	Tm        contentstream.Matrix // Text matrix. For the character pointer.
	Tlm       contentstream.Matrix // Text line matrix. For the start of line pointer.
	Texts     []XYText             // Text gets written here.

	// These fields are used to implement existing UniDoc behaviour.
	xPos, yPos float64
}

// newTextState returns a default textState.
func newTextState() textState {
	return textState{
		Th:    100,
		Tmode: RenderModeFill,
	}
}

// newTextObject returns a default textObject.
func newTextObject(e *Extractor, gs contentstream.GraphicsState, state *textState,
	fontStack *fontStacker) *textObject {
	return &textObject{
		e:         e,
		gs:        gs,
		fontStack: fontStack,
		State:     state,
		Tm:        contentstream.IdentityMatrix(),
		Tlm:       contentstream.IdentityMatrix(),
	}
}

// renderText emits byte array `data` to the calling program.
func (to *textObject) renderText(data []byte) error {
	font := to.getCurrentFont()

	charcodes := font.BytesToCharcodes(data)

	runes, numChars, numMisses := font.CharcodesToUnicode(charcodes)

	to.State.numChars += numChars
	to.State.numMisses += numMisses

	state := to.State
	tfs := state.Tfs
	th := state.Th / 100.0
	spaceMetrics, err := font.GetRuneCharMetrics(' ')
	if err != nil {
		spaceMetrics, _ = model.DefaultFont().GetRuneCharMetrics(' ')
	}
	spaceWidth := spaceMetrics.Wx * glyphTextRatio
	common.Log.Trace("spaceWidth=%.2f text=%q font=%s fontSize=%.1f", spaceWidth, runes, font, tfs)

	stateMatrix := contentstream.NewMatrix(
		tfs*th, 0,
		0, tfs,
		0, state.Trise)

	common.Log.Debug("==========================================")
	common.Log.Debug("%d codes=%+v runes=%q", len(charcodes), charcodes, runes)

	for i, r := range runes {

		code := charcodes[i]
		// The location of the text on the page in device coordinates is given by trm, the text
		// rendering matrix.
		trm := stateMatrix.Mult(to.Tm).Mult(to.gs.CTM)

		// calculate the text location displacement due to writing `r`. We will use this to update
		// to.Tm

		// w is the unscaled movement at the end of a word.
		w := 0.0
		if r == " " {
			w = state.Tw
		}

		m, ok := font.GetCharMetrics(code)
		if !ok {
			common.Log.Debug("ERROR: No metric for code=%d r=0x%04x=%+q %s", code, r, r, font)
			return errors.New("no char metrics")
		}

		// c is the character size in unscaled text units.
		c := Point{X: m.Wx * glyphTextRatio, Y: m.Wy * glyphTextRatio}

		// t is the displacement of the text cursor when the character is rendered.
		// float tx = displacementX * fontSize * horizontalScaling;
		// w = 0
		t0 := Point{X: (c.X*tfs + w) * th}
		t := Point{X: (c.X*tfs + state.Tc + w) * th}

		// td, td0 are t, t0 in matrix form.
		// td0 is where this char ends. td is where the next char stats.
		td0 := translationMatrix(t0)
		td := translationMatrix(t)

		nextTm := td.Mult(to.Tm)

		xyt := newXYText(
			string(r),
			translation(trm),
			translation(td0.Mult(to.Tm).Mult(to.gs.CTM)),
			trm.Orientation(),
			spaceWidth*trm.ScalingFactorX())
		common.Log.Debug("i=%d code=%d, xyt=%s", i, code, xyt)
		to.Texts = append(to.Texts, xyt)

		// update the text matrix by the displacement of the text location.
		to.Tm = nextTm
		common.Log.Trace("to.Tm=%s", to.Tm)
	}

	return nil
}

// glyphTextRatio converts Glyph metrics units to unscaled text space units.
const glyphTextRatio = 1.0 / 1000.0

// translation returns the translation part of `m`.
func translation(m contentstream.Matrix) Point {
	tx, ty := m.Translation()
	return Point{tx, ty}
}

// translationMatrix returns a matrix that translates by `p`.
func translationMatrix(p Point) contentstream.Matrix {
	return contentstream.TranslationMatrix(p.X, p.Y)
}

// moveTo moves the start of line pointer by `tx`,`ty` and sets the text pointer to the
// start of line pointer.
// Move to the start of the next line, offset from the start of the current line by (tx, ty).
// `tx` and `ty` are in unscaled text space units.
func (to *textObject) moveTo(tx, ty float64) {
	common.Log.Debug("moveTo: tx,ty=%.3f,%.3f Tm=%s Tlm=%s", tx, ty, to.Tm, to.Tlm)
	to.Tlm = contentstream.NewMatrix(1, 0, 0, 1, tx, ty).Mult(to.Tlm)
	to.Tm = to.Tlm
	common.Log.Debug("          ===> Tm=%s", to.Tm)
}

// XYText represents text drawn on a page and its position in device coordinates.
type XYText struct {
	Point                           // Position of text. Left-bottom?
	End              Point          // End of text. Right-top?
	ColorStroking    model.PdfColor // Colour that text is stroked with, if any.
	ColorNonStroking model.PdfColor // Colour that text is filled with, if any.
	Orient           contentstream.Orientation
	Text             string
	SpaceWidth       float64
	Font             string
	FontSize         float64
	counter          int
}

var counter int

func newXYText(text string, point, end Point, orient contentstream.Orientation, spaceWidth float64) XYText {
	counter++
	return XYText{
		Text:       text,
		Point:      point,
		End:        end,
		Orient:     orient,
		SpaceWidth: spaceWidth,
		counter:    counter,
	}
}

// String returns a string describing `t`.
func (t XYText) String() string {
	return fmt.Sprintf("@@%d %s,%s %.1f %q", t.counter,
		t.Point.String(), t.End.String(), t.End.X-t.X, truncate(t.Text, 100))
}

// Width returns the width of `t`.Text in its orientation.
func (t XYText) Width() float64 {
	var w float64
	switch t.Orient {
	case contentstream.OrientationLandscape:
		w = math.Abs(t.End.Y - t.Y)
	default:
		w = math.Abs(t.End.X - t.X)
	}
	common.Log.Debug("      Width %q (%s %s) -> %.1f", t.Text, t.Point.String(), t.End.String(), w)
	return w
}

// TextList is a list of texts and their positions on a PDF page.
type TextList []XYText

// Length returns the number of elements in `tl`.
func (tl *TextList) Length() int {
	return len(*tl)
}

// // AppendText appends the location and contents of `text` to a text list.
// func (tl *TextList) AppendText(gs contentstream.GraphicsState, p, e Point, text string, spaceWidth float64) {
// 	t := XYText{
// 		Point:            p,
// 		End:              e,
// 		ColorStroking:    gs.ColorStroking,
// 		ColorNonStroking: gs.ColorNonStroking,
// 		Orient:           gs.PageOrientation(),
// 		Text:             text,
// 		SpaceWidth:       spaceWidth,
// 	}
// 	common.Log.Debug("AppendText: %s", t.String())
// 	*tl = append(*tl, t)
// }

// ToText returns the contents of `tl` as a single string.
func (tl *TextList) ToText() string {
	tl.printTexts("ToText: before sorting")
	tl.SortPosition()

	lines := tl.toLines()
	texts := []string{}
	for _, l := range lines {
		texts = append(texts, l.Text)
	}
	return strings.Join(texts, "\n")
}

// SortPosition sorts a text list by its elements' position on a page.
// Sorting is by orientation then top to bottom, left to right when page is orientated so that text
// is horizontal.
func (tl *TextList) SortPosition() {
	sort.SliceStable(*tl, func(i, j int) bool {
		ti, tj := (*tl)[i], (*tl)[j]
		if ti.Orient != tj.Orient {
			return ti.Orient > tj.Orient
		}
		// x, y is orientated so text is horizontal.
		xi, xj := ti.X, tj.X
		yi, yj := ti.Y, tj.Y
		if ti.Orient == contentstream.OrientationLandscape {
			xi, yi = yi, -xi
			xj, yj = yj, -xj
		}

		if yi != yj {
			return yi > yj
		}
		return xi < xj
	})
}

// Line represents a line of text on a page.
type Line struct {
	Y     float64   // y position of line.
	Dx    []float64 // x distance between successive words in line.
	Text  string    // text in the line.
	Words []string  // words in the line
}

// toLines return the text and positions in `tl` as a slice of Line.
// NOTE: Caller must sort the text list by top-to-bottom, left-to-write (for orientation adjusted so
// that text is horizontal) before calling this function.
func (tl *TextList) toLines() []Line {
	portText, landText := TextList{}, TextList{}
	for _, t := range *tl {
		if t.Orient == contentstream.OrientationPortrait {
			portText = append(portText, t)
		} else {
			t.X, t.Y = t.Y, -t.X
			t.End.X, t.End.Y = t.End.Y, -t.End.X
			t.Orient = contentstream.OrientationPortrait
			landText = append(landText, t)
	}
}
	common.Log.Debug("toLines: portrait  ^^^^^^^")
	portLines := portText.toLinesOrient()
	common.Log.Debug("toLines: landscape &&&&&&&")
	landLines := landText.toLinesOrient()
	common.Log.Debug("portText=%d landText=%d", len(portText), len(landText))
	return append(portLines, landLines...)
}

// toLinesOrient returns the text and positions in `tl` as a slice of Line.
// NOTE: Caller must sort the text list top-to-bottom, left-to-write before calling this function.
func (tl *TextList) toLinesOrient() []Line {
	tl.printTexts("toLines: before")
	if len(*tl) == 0 {
		return []Line{}
	}
	lines := []Line{}
	words := []string{}
	x := []float64{}
	y := (*tl)[0].Y

	scanning := false

	averageCharWidth := ExponAve{}
	wordSpacing := ExponAve{}
	lastEndX := 0.0 // (*tl)[i-1).End.X

	for _, t := range *tl {
		// common.Log.Debug("%d --------------------------", i)
		if t.Y < y {
			if len(words) > 0 {
				line := newLine(y, x, words)
				if averageCharWidth.running {
					line = removeDuplicates(line, averageCharWidth.ave)
				}
				lines = append(lines, line)
			}
			words = []string{}
			x = []float64{}
			y = t.Y
			scanning = false
		}

		// Detect text movements that represent spaces on the printed page.
		// We use a heuristic from PdfBox: If the next character starts to the right of where a
		// character after a space at "normal spacing" would start, then there is a space before it.
		// The tricky thing to guess here is the width of a space at normal spacing.
		// We follow PdfBox and use min(deltaSpace, deltaCharWidth).
		deltaSpace := 0.0
		if t.SpaceWidth == 0 {
			deltaSpace = math.MaxFloat64
		} else {
			wordSpacing.update(t.SpaceWidth)
			deltaSpace = wordSpacing.ave * 0.5
		}
		averageCharWidth.update(t.Width())
		deltaCharWidth := averageCharWidth.ave * 0.3

		isSpace := false
		nextWordX := lastEndX + min(deltaSpace, deltaCharWidth)
		if scanning && t.Text != " " {
			isSpace = nextWordX < t.X
		}
		common.Log.Debug("t=%s", t)
		common.Log.Debug("width=%.2f delta=%.2f deltaSpace=%.2g deltaCharWidth=%.2g",
			t.Width(), min(deltaSpace, deltaCharWidth), deltaSpace, deltaCharWidth)

		common.Log.Debug("%+q [%.1f, %.1f] lastEndX=%.2f nextWordX=%.2f (%.2f) isSpace=%t",
			t.Text, t.X, t.Y, lastEndX, nextWordX, nextWordX-t.X, isSpace)
		if isSpace {
			words = append(words, " ")
			x = append(x, (lastEndX+t.X)*0.5)
		}

		// Add the text to the line.
		lastEndX = t.End.X
		words = append(words, t.Text)
		x = append(x, t.X)
		scanning = true
		common.Log.Debug("lastEndX=%.2f", lastEndX)
	}
	if len(words) > 0 {
		line := newLine(y, x, words)
		if averageCharWidth.running {
			line = removeDuplicates(line, averageCharWidth.ave)
		}
		lines = append(lines, line)
	}
	return lines
}

// min returns the lesser of `a` and `b`.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ExponAve implements an exponential average.
type ExponAve struct {
	ave     float64 // Current average value.
	running bool    // Has `ave` been set?
}

// update updates the exponential average `exp.ave` and returns it
func (exp *ExponAve) update(x float64) float64 {
	if !exp.running {
		exp.ave = x
		exp.running = true
	} else {
		exp.ave = (exp.ave + x) * 0.5
	}
	return exp.ave
}

// printTexts is a debugging function. XXX(peterwilliams97) Remove this.
func (tl *TextList) printTexts(message string) {
	return
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}
	prefix := fmt.Sprintf("[%s:%d]", file, line)

	common.Log.Debug("=====================================")
	common.Log.Debug("printTexts %s %s", prefix, message)
	common.Log.Debug("%d texts", len(*tl))
	parts := []string{}
	for i, t := range *tl {
		fmt.Printf("%5d: %s\n", i, t.String())
		parts = append(parts, t.Text)
	}
	common.Log.Debug("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Printf("%s\n", strings.Join(parts, ""))
	common.Log.Debug("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
}

// newLine returns the Line representation of strings `words` with y coordinate `y` and x
// coordinates `x`.
func newLine(y float64, x []float64, words []string) Line {
	dx := []float64{}
	for i := 1; i < len(x); i++ {
		dx = append(dx, x[i]-x[i-1])
	}
	return Line{Y: y, Dx: dx, Text: strings.Join(words, ""), Words: words}
}

// removeDuplicates returns `line` with duplicate characters removed. `charWidth` is the average
// character width for the line.
func removeDuplicates(line Line, charWidth float64) Line {
	if len(line.Dx) == 0 {
		return line
	}

	tol := charWidth * 0.3
	words := []string{line.Words[0]}
	dxList := []float64{}

	w0 := line.Words[0]
	for i, dx := range line.Dx {
		w := line.Words[i+1]
		if w != w0 || dx > tol {
			words = append(words, w)
			dxList = append(dxList, dx)
		}
		w0 = w
	}
	return Line{Y: line.Y, Dx: dxList, Text: strings.Join(words, ""), Words: words}
}

// PageOrientation is a heuristic for the orientation of a page.
// XXX TODO: Use Page's Rotate flag instead.
func (tl *TextList) PageOrientation() contentstream.Orientation {
	landscapeCount := 0
	for _, t := range *tl {
		if t.Orient == contentstream.OrientationLandscape {
			landscapeCount++
		}
	}
	portraitCount := len(*tl) - landscapeCount
	if landscapeCount > portraitCount {
		return contentstream.OrientationLandscape
	}
	return contentstream.OrientationPortrait
}

// Transform transforms all points in `tl` by the affine transformation a, b, c, d, tx, ty.
func (tl *TextList) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	for _, t := range *tl {
		t.X, t.Y = m.Transform(t.X, t.Y)
	}
}

// getCurrentFont returns the font on top of the font stack, or DefaultFont if the font stack is
// empty.
func (to *textObject) getCurrentFont() *model.PdfFont {
	if to.fontStack.empty() {
		common.Log.Debug("ERROR: No font defined. Using default.")
		return model.DefaultFont()
	}
	return to.fontStack.peek()
}

// getFont returns the font named `name` if it exists in the page's resources or an error if it
// doesn't. It caches the returned fonts.
func (to *textObject) getFont(name string) (*model.PdfFont, error) {
	if to.e.fontCache != nil {
		to.e.accessCount++
		entry, ok := to.e.fontCache[name]
		if ok {
			entry.access = to.e.accessCount
			return entry.font, nil
		}
	}

	// Font not in cache. Load it.
	font, err := to.getFontDirect(name)
	if err != nil {
		return nil, err
	}

	if to.e.fontCache != nil {
		entry := fontEntry{font, to.e.accessCount}

		// Eject a victim if the cache is full.
		if len(to.e.fontCache) >= maxFontCache {
			names := []string{}
			for name := range to.e.fontCache {
				names = append(names, name)
			}
			sort.Slice(names, func(i, j int) bool {
				return to.e.fontCache[names[i]].access < to.e.fontCache[names[j]].access
			})
			delete(to.e.fontCache, names[0])
		}
		to.e.fontCache[name] = entry
	}

	return font, nil
}

// fontEntry is a entry in the font cache.
type fontEntry struct {
	font   *model.PdfFont // The font being cached.
	access int64          // Last access. Used to determine LRU cache victims.
}

// maxFontCache is the maximum number of PdfFont's in fontCache.
const maxFontCache = 10

// getFontDirect returns the font named `name` if it exists in the page's resources or an error if
// is doesn't.
// This is a direct (uncached access).
func (to *textObject) getFontDirect(name string) (*model.PdfFont, error) {

	// This is a hack for testing.
	if name == "UniDocCourier" {
		return model.NewStandard14FontMustCompile(model.Courier), nil
	}

	fontObj, err := to.getFontDict(name)
	if err != nil {
		return nil, err
	}
	font, err := model.NewPdfFontFromPdfObject(fontObj)
	if err != nil {
		common.Log.Debug("getFontDirect: NewPdfFontFromPdfObject failed. name=%#q err=%v", name, err)
	}
	return font, err
}

// getFontDict returns the font dict with key `name` if it exists in the page's Font resources or
// an error if it doesn't.
func (to *textObject) getFontDict(name string) (fontObj core.PdfObject, err error) {
	resources := to.e.resources
	if resources == nil {
		common.Log.Debug("getFontDict. No resources. name=%#q", name)
		return nil, nil
	}
	fontObj, found := resources.GetFontByName(core.PdfObjectName(name))
	if !found {
		common.Log.Debug("ERROR: getFontDict: Font not found: name=%#q", name)
		return nil, errors.New("font not in resources")
	}
	return fontObj, nil
}
