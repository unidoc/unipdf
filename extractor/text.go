/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/transform"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/text/unicode/norm"
)

// ExtractText processes and extracts all text data in content streams and returns as a string.
// It takes into account character encodings in the PDF file, which are decoded by
// CharcodeBytesToUnicode.
// Characters that can't be decoded are replaced with MissingCodeRune ('\ufffd' = �).
func (e *Extractor) ExtractText() (string, error) {
	text, _, _, err := e.ExtractTextWithStats()
	return text, err
}

// ExtractTextWithStats works like ExtractText but returns the number of characters in the output
// (`numChars`) and the number of characters that were not decoded (`numMisses`).
func (e *Extractor) ExtractTextWithStats() (extracted string, numChars int, numMisses int, err error) {
	pageText, numChars, numMisses, err := e.ExtractPageText()
	if err != nil {
		return "", numChars, numMisses, err
	}
	return pageText.Text(), numChars, numMisses, nil
}

// ExtractPageText returns the text contents of `e` (an Extractor for a page) as a PageText.
func (e *Extractor) ExtractPageText() (*PageText, int, int, error) {
	pt, numChars, numMisses, err := e.extractPageText(e.contents, e.resources, 0)
	if err != nil {
		return nil, numChars, numMisses, err
	}
	pt.computeViews()
	procBuf(pt)

	return pt, numChars, numMisses, err
}

// extractPageText returns the text contents of content stream `e` and resouces `resources` as a
// PageText.
// This can be called on a page or a form XObject.
func (e *Extractor) extractPageText(contents string, resources *model.PdfPageResources, level int) (
	*PageText, int, int, error) {
	common.Log.Trace("extractPageText: level=%d", level)
	pageText := &PageText{}
	state := newTextState()
	fontStack := fontStacker{}
	var to *textObject

	cstreamParser := contentstream.NewContentStreamParser(contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		common.Log.Debug("ERROR: extractPageText parse failed. err=%v", err)
		return pageText, state.numChars, state.numMisses, err
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
				if state.tfont != nil {
					common.Log.Trace("Save font state: %s\n->%s\n%s",
						fontStack.peek(), state.tfont, fontStack.String())
					fontStack.push(state.tfont)
				}
			case "Q":
				if !fontStack.empty() {
					common.Log.Trace("Restore font state: %s\n->%s\n%s",
						fontStack.peek(), fontStack.get(-2), fontStack.String())
					fontStack.pop()
				}
				if len(fontStack) >= 2 {
					common.Log.Trace("Restore font state: %s\n->%s\n%s",
						state.tfont, fontStack.peek(), fontStack.String())
					state.tfont = fontStack.pop()
				}
			case "BT": // Begin text
				// Begin a text object, initializing the text matrix, Tm, and the text line matrix,
				// Tlm, to the identity matrix. Text objects shall not be nested; a second BT shall
				// not appear before an ET.
				if to != nil {
					common.Log.Debug("BT called while in a text object")
				}
				to = newTextObject(e, resources, gs, &state, &fontStack)
			case "ET": // End Text
				pageText.marks = append(pageText.marks, to.marks...)
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
			case "Tj": // Show text.
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
			case "TJ": // Show text with adjustable spacing.
				if ok, err := to.checkOp(op, 1, true); !ok {
					common.Log.Debug("ERROR: TJ err=%v", err)
					return err
				}
				args, ok := core.GetArray(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: TJ op=%s GetArrayVal failed", op)
					return err
				}
				return to.showTextAdjusted(args)
			case "'": // Move to next line and show text.
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
			case `"`: // Set word and character spacing, move to next line, and show text.
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
			case "TL": // Set text leading.
				y, err := floatParam(op)
				if err != nil {
					common.Log.Debug("ERROR: TL err=%v", err)
					return err
				}
				to.setTextLeading(y)
			case "Tc": // Set character spacing.
				y, err := floatParam(op)
				if err != nil {
					common.Log.Debug("ERROR: Tc err=%v", err)
					return err
				}
				to.setCharSpacing(y)
			case "Tf": // Set font.
				if to == nil {
					// This is needed for 26-Hazard-Thermal-environment.pdf
					to = newTextObject(e, resources, gs, &state, &fontStack)
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
			case "Tm": // Set text matrix.
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
			case "Tr": // Set text rendering mode.
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
			case "Ts": // Set text rise.
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
			case "Tw": // Set word spacing.
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
			case "Tz": // Set horizontal scaling.
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

			case "Do":
				// Handle XObjects by recursing through form XObjects.
				name := *op.Params[0].(*core.PdfObjectName)
				_, xtype := resources.GetXObjectByName(name)
				if xtype != model.XObjectTypeForm {
					break
				}
				// Only process each form once.
				formResult, ok := e.formResults[string(name)]
				if !ok {
					xform, err := resources.GetXObjectFormByName(name)
					if err != nil {
						common.Log.Debug("ERROR: %v", err)
						return err
					}
					formContent, err := xform.GetContentStream()
					if err != nil {
						common.Log.Debug("ERROR: %v", err)
						return err
					}
					formResources := xform.Resources
					if formResources == nil {
						formResources = resources
					}
					tList, numChars, numMisses, err := e.extractPageText(string(formContent),
						formResources, level+1)
					if err != nil {
						common.Log.Debug("ERROR: %v", err)
						return err
					}
					formResult = textResult{*tList, numChars, numMisses}
					e.formResults[string(name)] = formResult
				}

				pageText.marks = append(pageText.marks, formResult.pageText.marks...)
				state.numChars += formResult.numChars
				state.numMisses += formResult.numMisses
			}
			return nil
		})

	err = processor.Process(resources)
	if err != nil {
		common.Log.Debug("ERROR: Processing: err=%v", err)
	}
	return pageText, state.numChars, state.numMisses, err
}

type textResult struct {
	pageText  PageText
	numChars  int
	numMisses int
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
	to.state.tl = -ty
	to.moveTo(tx, ty)
}

// nextLine "T*"" Moves start of text line to next text line
// Move to the start of the next line. This operator has the same effect as the code
//    0 -Tl Td
// where Tl denotes the current leading parameter in the text state. The negative of Tl is used
// here because Tl is the text leading expressed as a positive number. Going to the next line
// entails decreasing the y coordinate. (page 250)
func (to *textObject) nextLine() {
	to.moveTo(0, -to.state.tl)
}

// setTextMatrix "Tm".
// Set the text matrix, Tm, and the text line matrix, Tlm to the Matrix specified by the 6 numbers
// in `f` (page 250).
func (to *textObject) setTextMatrix(f []float64) {
	if len(f) != 6 {
		common.Log.Debug("ERROR: len(f) != 6 (%d)", len(f))
		return
	}
	a, b, c, d, tx, ty := f[0], f[1], f[2], f[3], f[4], f[5]
	to.tm = transform.NewMatrix(a, b, c, d, tx, ty)
	to.tlm = to.tm
}

// showText "Tj". Show a text string.
func (to *textObject) showText(charcodes []byte) error {
	return to.renderText(charcodes)
}

// showTextAdjusted "TJ". Show text with adjustable spacing.
func (to *textObject) showTextAdjusted(args *core.PdfObjectArray) error {
	vertical := false
	for _, o := range args.Elements() {
		switch o.(type) {
		case *core.PdfObjectFloat, *core.PdfObjectInteger:
			x, err := core.GetNumberAsFloat(o)
			if err != nil {
				common.Log.Debug("ERROR: showTextAdjusted. Bad numerical arg. o=%s args=%+v", o, args)
				return err
			}
			dx, dy := -x*0.001*to.state.tfs, 0.0
			if vertical {
				dy, dx = dx, dy
			}
			td := translationMatrix(transform.Point{X: dx, Y: dy})
			to.tm.Concat(td)
			common.Log.Trace("showTextAdjusted: dx,dy=%3f,%.3f Tm=%s", dx, dy, to.tm)
		case *core.PdfObjectString:
			charcodes, ok := core.GetStringBytes(o)
			if !ok {
				common.Log.Trace("showTextAdjusted: Bad string arg. o=%s args=%+v", o, args)
				return core.ErrTypeError
			}
			to.renderText(charcodes)
		default:
			common.Log.Debug("ERROR: showTextAdjusted. Unexpected type (%T) args=%+v", o, args)
			return core.ErrTypeError
		}
	}
	return nil
}

// setTextLeading "TL". Set text leading.
func (to *textObject) setTextLeading(y float64) {
	if to == nil || to.state == nil {
		return
	}
	to.state.tl = y
}

// setCharSpacing "Tc". Set character spacing.
func (to *textObject) setCharSpacing(x float64) {
	if to == nil {
		return
	}
	to.state.tc = x
}

// setFont "Tf". Set font.
func (to *textObject) setFont(name string, size float64) error {
	if to == nil {
		return nil
	}
	font, err := to.getFont(name)
	if err == nil {
		to.state.tfont = font
		if len(*to.fontStack) == 0 {
			to.fontStack.push(font)
		} else {
			(*to.fontStack)[len(*to.fontStack)-1] = font
		}
	} else if err == model.ErrFontNotSupported {
		// TODO(peterwilliams97): Do we need to handle this case in a special way?
		return err
	} else {
		return err
	}
	to.state.tfs = size
	return nil
}

// setTextRenderMode "Tr". Set text rendering mode.
func (to *textObject) setTextRenderMode(mode int) {
	if to == nil {
		return
	}
	to.state.tmode = RenderMode(mode)
}

// setTextRise "Ts". Set text rise.
func (to *textObject) setTextRise(y float64) {
	if to == nil {
		return
	}
	to.state.trise = y
}

// setWordSpacing "Tw". Set word spacing.
func (to *textObject) setWordSpacing(y float64) {
	if to == nil {
		return
	}
	to.state.tw = y
}

// setHorizScaling "Tz". Set horizontal scaling.
func (to *textObject) setHorizScaling(y float64) {
	if to == nil {
		return
	}
	to.state.th = y
}

// floatParam returns the single float parameter of operator `op`, or an error if it doesn't have
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
func (to *textObject) checkOp(op *contentstream.ContentStreamOperation, numParams int, hard bool) (
	ok bool, err error) {
	if to == nil {
		var params []core.PdfObject
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
	tc    float64        // Character spacing. Unscaled text space units.
	tw    float64        // Word spacing. Unscaled text space units.
	th    float64        // Horizontal scaling.
	tl    float64        // Leading. Unscaled text space units. Used by TD,T*,'," see Table 108.
	tfs   float64        // Text font size.
	tmode RenderMode     // Text rendering mode.
	trise float64        // Text rise. Unscaled text space units. Set by Ts.
	tfont *model.PdfFont // Text font.
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
//  trm := to.gs.CTM.Mult(stateMatrix).Mult(to.tm)

// textObject represents a PDF text object.
type textObject struct {
	e         *Extractor
	resources *model.PdfPageResources
	gs        contentstream.GraphicsState
	fontStack *fontStacker
	state     *textState
	tm        transform.Matrix // Text matrix. For the character pointer.
	tlm       transform.Matrix // Text line matrix. For the start of line pointer.
	marks     []textMark       // Text marks get written here.
}

// newTextState returns a default textState.
func newTextState() textState {
	return textState{
		th:    100,
		tmode: RenderModeFill,
	}
}

// newTextObject returns a default textObject.
func newTextObject(e *Extractor, resources *model.PdfPageResources, gs contentstream.GraphicsState,
	state *textState, fontStack *fontStacker) *textObject {
	return &textObject{
		e:         e,
		resources: resources,
		gs:        gs,
		fontStack: fontStack,
		state:     state,
		tm:        transform.IdentityMatrix(),
		tlm:       transform.IdentityMatrix(),
	}
}

// renderText processes and renders byte array `data` for extraction purposes.
func (to *textObject) renderText(data []byte) error {
	font := to.getCurrentFont()
	charcodes := font.BytesToCharcodes(data)
	runes, numChars, numMisses := font.CharcodesToUnicodeWithStats(charcodes)
	if numMisses > 0 {
		common.Log.Debug("renderText: numChars=%d numMisses=%d", numChars, numMisses)
	}

	to.state.numChars += numChars
	to.state.numMisses += numMisses

	state := to.state
	tfs := state.tfs
	th := state.th / 100.0
	spaceMetrics, ok := font.GetRuneMetrics(' ')
	if !ok {
		spaceMetrics, ok = font.GetCharMetrics(32)
	}
	if !ok {
		spaceMetrics, _ = model.DefaultFont().GetRuneMetrics(' ')
	}
	spaceWidth := spaceMetrics.Wx * glyphTextRatio
	common.Log.Trace("spaceWidth=%.2f text=%q font=%s fontSize=%.1f", spaceWidth, runes, font, tfs)

	stateMatrix := transform.NewMatrix(
		tfs*th, 0,
		0, tfs,
		0, state.trise)

	common.Log.Trace("renderText: %d codes=%+v runes=%q", len(charcodes), charcodes, runes)

	for i, r := range runes {
		// TODO(peterwilliams97): Need to find and fix cases where this happens.
		if r == '\x00' {
			continue
		}

		code := charcodes[i]
		// The location of the text on the page in device coordinates is given by trm, the text
		// rendering matrix.
		trm := to.gs.CTM.Mult(to.tm).Mult(stateMatrix)

		// calculate the text location displacement due to writing `r`. We will use this to update
		// to.tm

		// w is the unscaled movement at the end of a word.
		w := 0.0
		if r == ' ' {
			w = state.tw
		}

		m, ok := font.GetCharMetrics(code)
		if !ok {
			common.Log.Debug("ERROR: No metric for code=%d r=0x%04x=%+q %s", code, r, r, font)
			return errors.New("no char metrics")
		}

		// c is the character size in unscaled text units.
		c := transform.Point{X: m.Wx * glyphTextRatio, Y: m.Wy * glyphTextRatio}

		// t0 is the end of this character.
		// t is the displacement of the text cursor when the character is rendered.
		t0 := transform.Point{X: (c.X*tfs + w) * th}
		t := transform.Point{X: (c.X*tfs + state.tc + w) * th}

		// td, td0 are t, t0 in matrix form.
		// td0 is where this character ends. td is where the next character starts.
		td0 := translationMatrix(t0)
		td := translationMatrix(t)

		common.Log.Trace("\"%c\" stateMatrix=%s CTM=%s Tm=%s", r, stateMatrix, to.gs.CTM, to.tm)
		common.Log.Trace("tfs=%.3f th=%.3f Tc=%.3f w=%.3f (Tw=%.3f)", tfs, th, state.tc, w, state.tw)
		common.Log.Trace("m=%s c=%+v t0=%+v td0=%s trm0=%s", m, c, t0, td0, td0.Mult(to.tm).Mult(to.gs.CTM))

		mark := to.newTextMark(
			string(r),
			trm,
			translation(to.gs.CTM.Mult(to.tm).Mult(td0)),
			math.Abs(spaceWidth*trm.ScalingFactorX()),
			font,
			to.state.tc)
		if font == nil {
			common.Log.Debug("ERROR: No font.")
		} else if font.Encoder() == nil {
			common.Log.Debug("ERROR: No encoding. font=%s", font)
		} else {
			original, ok := font.Encoder().CharcodeToRune(code)
			if ok {
				mark.original = string(original)
			}
		}
		common.Log.Trace("i=%d code=%d mark=%s trm=%s", i, code, mark, trm)
		to.marks = append(to.marks, mark)

		// update the text matrix by the displacement of the text location.
		to.tm.Concat(td)
		common.Log.Trace("to.tm=%s", to.tm)
	}

	return nil
}

// glyphTextRatio converts Glyph metrics units to unscaled text space units.
const glyphTextRatio = 1.0 / 1000.0

// translation returns the translation part of `m`.
func translation(m transform.Matrix) transform.Point {
	tx, ty := m.Translation()
	return transform.Point{X: tx, Y: ty}
}

// translationMatrix returns a matrix that translates by `p`.
func translationMatrix(p transform.Point) transform.Matrix {
	return transform.TranslationMatrix(p.X, p.Y)
}

// moveTo moves the start of line pointer by `tx`,`ty` and sets the text pointer to the
// start of line pointer.
// Move to the start of the next line, offset from the start of the current line by (tx, ty).
// `tx` and `ty` are in unscaled text space units.
func (to *textObject) moveTo(tx, ty float64) {
	to.tlm.Concat(transform.NewMatrix(1, 0, 0, 1, tx, ty))
	to.tm = to.tlm
}

// textMark represents text drawn on a page and its position in device coordinates.
// All dimensions are in device coordinates.
type textMark struct {
	text          string             // The text (decoded via ToUnicode).
	original      string             // Original text (decoded).
	bbox          model.PdfRectangle // Text bounding box.
	orient        int                // The text orientation in degrees. This is the current TRM rounded to 10°.
	orientedStart transform.Point    // Left of text in orientation where text is horizontal.
	orientedEnd   transform.Point    // Right of text in orientation where text is horizontal.
	height        float64            // Text height.
	spaceWidth    float64            // Best guess at the width of a space in the font the text was rendered with.
	font          *model.PdfFont     // The font the mark was drawn with.
	fontsize      float64            // The font size the mark was drawn with.
	charspacing   float64            // TODO (peterwilliams97: Should this be exposed in TextMark?
	trm           transform.Matrix   // The current text rendering matrix (TRM above).
	end           transform.Point    // The end of character device coordinates.
	count         int64              // To help with reading debug logs.
}

// newTextMark returns a textMark for text `text` rendered with text rendering matrix (TRM) `trm`
// and end of character device coordinates `end`. `spaceWidth` is our best guess at the width of a
// space in the font the text is rendered in device coordinates.
func (to *textObject) newTextMark(text string, trm transform.Matrix, end transform.Point,
	spaceWidth float64, font *model.PdfFont, charspacing float64) textMark {
	to.e.textCount++
	theta := trm.Angle()
	orient := nearestMultiple(theta, 10)
	var height float64
	if orient%180 != 90 {
		height = trm.ScalingFactorY()
	} else {
		height = trm.ScalingFactorX()
	}

	start := translation(trm)
	bbox := model.PdfRectangle{Llx: start.X, Lly: start.Y, Urx: end.X, Ury: end.Y}
	switch orient % 360 {
	case 90:
		bbox.Urx -= height
	case 180:
		bbox.Ury -= height
	case 270:
		bbox.Urx += height
	default:
		bbox.Ury += height
	}
	tm := textMark{
		text:          text,
		orient:        orient,
		bbox:          bbox,
		orientedStart: start.Rotate(theta),
		orientedEnd:   end.Rotate(theta),
		height:        math.Abs(height),
		spaceWidth:    spaceWidth,
		font:          font,
		fontsize:      to.state.tfs,
		charspacing:   charspacing,
		trm:           trm,
		end:           end,
		count:         to.e.textCount,
	}
	if !isTextSpace(tm.text) && tm.Width() == 0.0 {
		common.Log.Debug("ERROR: Zero width text. tm=%s\n\tm=%#v", tm, tm)
	}
	return tm
}

// isTextSpace returns true if `text` contains nothing but space code points.
func isTextSpace(text string) bool {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// nearestMultiple return the integer multiple of `m` that is closest to `x`.
func nearestMultiple(x float64, m int) int {
	if m == 0 {
		m = 1
	}
	fac := float64(m)
	return int(math.Round(x/fac) * fac)
}

// String returns a string describing `tm`.
func (tm textMark) String() string {
	return fmt.Sprintf("textMark{@%03d [%.3f,%.3f] w=%.1f %d° %q}",
		tm.count, tm.orientedStart.X, tm.orientedStart.Y, tm.Width(), tm.orient,
		truncate(tm.text, 100))
}

// Width returns the width of `tm`.text in the text direction.
func (tm textMark) Width() float64 {
	return math.Abs(tm.orientedStart.X - tm.orientedEnd.X)
}

// ToTextMark returns the public view of `tm`.
func (tm textMark) ToTextMark() TextMark {
	return TextMark{
		Text:     tm.text,
		Original: tm.original,
		BBox:     tm.bbox,
		Font:     tm.font,
		FontSize: tm.fontsize,
	}
}

// PageText represents the layout of text on a device page.
type PageText struct {
	marks     []textMark // Texts and their positions on a PDF page.
	viewText  string     // Extracted page text.
	viewMarks []TextMark // Public view of `marks`.
}

// String returns a string describing `pt`.
func (pt PageText) String() string {
	summary := fmt.Sprintf("PageText: %d elements", len(pt.marks))
	parts := []string{"-" + summary}
	for _, tm := range pt.marks {
		parts = append(parts, tm.String())
	}
	parts = append(parts, "+"+summary)
	return strings.Join(parts, "\n")
}

// length returns the number of elements in `pt.marks`.
func (pt PageText) length() int {
	return len(pt.marks)
}

// Text returns the extracted page text.
func (pt PageText) Text() string {
	return pt.viewText
}

// ToText returns the page text as a single string.
// Deprecated: This function is deprecated and will be removed in a future major version. Please use
// Text() instead.
func (pt PageText) ToText() string {
	return pt.Text()
}

// Marks returns the TextMark collection for a page. It represents all the text on the page.
func (pt PageText) Marks() *TextMarkArray {
	return &TextMarkArray{marks: pt.viewMarks}
}

// TextMarkArray is a collection of TextMarks.
type TextMarkArray struct {
	marks []TextMark
}

// Append appends `mark` to the mark array.
func (ma *TextMarkArray) Append(mark TextMark) {
	ma.marks = append(ma.marks, mark)
}

// String returns a string describing `ma`.
func (ma TextMarkArray) String() string {
	n := len(ma.marks)
	if n == 0 {
		return "EMPTY"
	}
	m0 := ma.marks[0]
	m1 := ma.marks[n-1]
	return fmt.Sprintf("{TEXTMARKARRAY: %d elements\n\tfirst=%s\n\t last=%s}", n, m0, m1)
}

// Elements returns the TextMarks in `ma`.
func (ma *TextMarkArray) Elements() []TextMark {
	return ma.marks
}

// Len returns the number of TextMarks in `ma`.
func (ma *TextMarkArray) Len() int {
	if ma == nil {
		return 0
	}
	return len(ma.marks)
}

// RangeOffset returns the TextMarks in `ma` that have `start` <= TextMark.Offset < `end`.
func (ma *TextMarkArray) RangeOffset(start, end int) (*TextMarkArray, error) {
	if ma == nil {
		return nil, errors.New("ma==nil")
	}
	if end < start {
		return nil, fmt.Errorf("end < start. RangeOffset not defined. start=%d end=%d ", start, end)
	}
	n := len(ma.marks)
	if n == 0 {
		return ma, nil
	}
	if start < ma.marks[0].Offset {
		start = ma.marks[0].Offset
	}
	if end > ma.marks[n-1].Offset+1 {
		end = ma.marks[n-1].Offset + 1
	}

	iStart := sort.Search(n, func(i int) bool { return ma.marks[i].Offset >= start })
	if !(0 <= iStart && iStart < n) {
		err := fmt.Errorf("Out of range. start=%d iStart=%d len=%d\n\tfirst=%v\n\t last=%v",
			start, iStart, n, ma.marks[0], ma.marks[n-1])
		return nil, err
	}
	iEnd := sort.Search(n, func(i int) bool { return ma.marks[i].Offset > end-1 })
	if !(0 <= iEnd && iEnd < n) {
		err := fmt.Errorf("Out of range. end=%d iEnd=%d len=%d\n\tfirst=%v\n\t last=%v",
			end, iEnd, n, ma.marks[0], ma.marks[n-1])
		return nil, err
	}
	if iEnd <= iStart {
		// This should never happen.
		return nil, fmt.Errorf("start=%d end=%d iStart=%d iEnd=%d", start, end, iStart, iEnd)
	}
	return &TextMarkArray{marks: ma.marks[iStart:iEnd]}, nil
}

// BBox returns the smallest axis-aligned rectangle that encloses all the TextMarks in `ma`.
func (ma *TextMarkArray) BBox() (model.PdfRectangle, bool) {
	if len(ma.marks) == 0 {
		return model.PdfRectangle{}, false
	}
	bbox := ma.marks[0].BBox
	for _, tm := range ma.marks[1:] {
		if isTextSpace(tm.Text) {
			continue
		}
		bbox = rectUnion(bbox, tm.BBox)
	}
	return bbox, true
}

// rectUnion returns the smallest axis-aligned rectangle that contains `b1` and `b2`.
func rectUnion(b1, b2 model.PdfRectangle) model.PdfRectangle {
	return model.PdfRectangle{
		Llx: math.Min(b1.Llx, b2.Llx),
		Lly: math.Min(b1.Lly, b2.Lly),
		Urx: math.Max(b1.Urx, b2.Urx),
		Ury: math.Max(b1.Ury, b2.Ury),
	}
}

// TextMark represents extracted text on a page with information regarding both textual content,
// formatting (font and size) and positioning.
// It is the smallest unit of text on a PDF page, typically a single character.
//
// getBBox() in test_text.go shows how to compute bounding boxes of substrings of extracted text.
// The following code extracts the text on PDF page `page` into `text` then finds the bounding box
// `bbox` of substring `term` in `text`.
//
//     ex, _ := New(page)
//     // handle errors
//     pageText, _, _, err := ex.ExtractPageText()
//     // handle errors
//     text := pageText.Text()
//     textMarks := pageText.Marks()
//
//     	start := strings.Index(text, term)
//      end := start + len(term)
//      spanMarks, err := textMarks.RangeOffset(start, end)
//      // handle errors
//      bbox, ok := spanMarks.BBox()
//      // handle errors
type TextMark struct {
	// Text is the extracted text. It has been decoded to Unicode via ToUnicode().
	Text string
	// Original is the text in the PDF. It has not been decoded like `Text`.
	Original string
	// BBox is the bounding box of the text.
	BBox model.PdfRectangle
	// Font is the font the text was drawn with.
	Font *model.PdfFont
	// FontSize is the font size the text was drawn with.
	FontSize float64
	// Offset is the offset of the start of TextMark.Text in the extracted text. If you do this
	//   text, textMarks := pageText.Text(), pageText.Marks()
	//   marks := textMarks.Elements()
	// then marks[i].Offset is the offset of marks[i].Text in text.
	Offset int
	// Meta is set true for spaces and line breaks that we insert in the extracted text. We insert
	// spaces (line breaks) when we see characters that are over a threshold horizontal (vertical)
	//  distance  apart. See wordJoiner (lineJoiner) in PageText.computeViews().
	Meta bool
}

// String returns a string describing `tm`.
func (tm TextMark) String() string {
	b := tm.BBox
	var font string
	if tm.Font != nil {
		font = tm.Font.String()
		if len(font) > 50 {
			font = font[:50] + "..."
		}
	}
	var meta string
	if tm.Meta {
		meta = " *M*"
	}
	return fmt.Sprintf("{TextMark: %d %q=%02x (%5.1f, %5.1f) (%5.1f, %5.1f) %s%s}",
		tm.Offset, tm.Text, []rune(tm.Text), b.Llx, b.Lly, b.Urx, b.Ury, font, meta)
}

// computeViews processes the page TextMarks sorting by position and populates `pt.viewText` and
// `pt.viewMarks` which represent the text and marks in the order which it is read on the page.
// The comments above the TextMark definition describe how to use the []TextMark to
// maps substrings of the page text to locations on the PDF page.
func (pt *PageText) computeViews() {
	fontHeight := pt.height()
	// We sort with a y tolerance to allow for subscripts, diacritics etc.
	tol := minFloat(fontHeight*0.2, 5.0)
	common.Log.Trace("ToTextLocation: %d elements fontHeight=%.1f tol=%.1f", len(pt.marks), fontHeight, tol)
	// Uncomment the 2 following Debug statements to see the effects of sorting.
	// common.Log.Debug("computeViews: Before sorting %s", pt)
	pt.sortPosition(tol)
	// common.Log.Debug("computeViews: After sorting %s", pt)
	lines := pt.toLines(tol)
	texts := make([]string, len(lines))
	for i, l := range lines {
		texts[i] = strings.Join(l.words(), wordJoiner)
	}
	text := strings.Join(texts, lineJoiner)
	var marks []TextMark
	offset := 0
	for i, l := range lines {
		for j, tm := range l.marks {
			tm.Offset = offset
			marks = append(marks, tm)
			offset += len(tm.Text)
			if j == len(l.marks)-1 {
				break
			}
			if wordJoinerLen > 0 {
				tm := TextMark{
					Offset: offset,
					Text:   wordJoiner,
					Meta:   true,
				}
				marks = append(marks, tm)
				offset += wordJoinerLen
			}
		}
		if i == len(lines)-1 {
			break
		}
		if lineJoinerLen > 0 {
			tm := TextMark{
				Offset: offset,
				Text:   lineJoiner,
				Meta:   true,
			}
			marks = append(marks, tm)
			offset += lineJoinerLen
		}
	}
	pt.viewText = text
	pt.viewMarks = marks
}

// height returns the max height of the elements in `pt.marks`.
func (pt PageText) height() float64 {
	fontHeight := 0.0
	for _, tm := range pt.marks {
		if tm.height > fontHeight {
			fontHeight = tm.height
		}
	}
	return fontHeight
}

const (
	// wordJoiner is added between text marks in extracted text.
	wordJoiner = ""
	// lineJoiner is added between lines in extracted text.
	lineJoiner = "\n"
)

var (
	wordJoinerLen = len(wordJoiner)
	lineJoinerLen = len(lineJoiner)
	// spaceMark is a special TextMark used for spaces.
	spaceMark = TextMark{
		Text:     " ",
		Original: " ",
		Meta:     true,
	}
)

// sortPosition sorts a text list by its elements' positions on a page.
// Sorting is by orientation then top to bottom, left to right when page is orientated so that text
// is horizontal.
// Text is considered to be on different lines if the lines' orientedStart.Y differs by more than `tol`.
func (pt *PageText) sortPosition(tol float64) {
	if len(pt.marks) == 0 {
		return
	}

	// For grouping data vertically into lines, it is necessary to have the data presorted by
	// descending y position.
	sort.SliceStable(pt.marks, func(i, j int) bool {
		ti, tj := pt.marks[i], pt.marks[j]
		if ti.orient != tj.orient {
			return ti.orient < tj.orient
		}
		return ti.orientedStart.Y >= tj.orientedStart.Y
	})

	// Cluster the marks into y-clusters by relative y proximity. Each cluster is our guess of what
	// makes up a line of text.
	clusters := make([]int, len(pt.marks))
	cluster := 0
	clusters[0] = cluster
	for i := 1; i < len(pt.marks); i++ {
		if pt.marks[i-1].orient != pt.marks[i].orient {
			cluster++
		} else {
			if pt.marks[i-1].orientedStart.Y - pt.marks[i].orientedStart.Y > tol {
				cluster++
			}
		}
		clusters[i] = cluster
	}

	// Sort by y-cluster and x.
	sort.SliceStable(pt.marks, func(i, j int) bool {
		ti, tj := pt.marks[i], pt.marks[j]
		if ti.orient != tj.orient {
			return ti.orient < tj.orient
		}
		if clusters[i] != clusters[j] {
			return clusters[i] < clusters[j]
		}
		return ti.orientedStart.X < tj.orientedStart.X
	})
}

// textLine represents a line of text on a page.
type textLine struct {
	x      float64    // x position of line.
	y      float64    // y position of line.
	h      float64    // height of line text.
	dxList []float64  // x distance between successive words in line.
	marks  []TextMark // TextMarks in the line.
}

// words returns the texts in `tl`.
func (tl textLine) words() []string {
	var texts []string
	for _, tm := range tl.marks {
		texts = append(texts, tm.Text)
	}
	return texts
}

// toLines returns the text and positions in `pt.marks` as a slice of textLine.
// NOTE: Caller must sort the text list top-to-bottom, left-to-right (for orientation adjusted so
// that text is horizontal) before calling this function.
func (pt PageText) toLines(tol float64) []textLine {
	// We divide `pt.marks` into slices which contain texts with the same orientation, extract the
	// lines for each orientation then return the concatenation of these lines sorted by orientation.
	tlOrient := make(map[int][]textMark, len(pt.marks))
	for _, tm := range pt.marks {
		tlOrient[tm.orient] = append(tlOrient[tm.orient], tm)
	}
	var lines []textLine
	for _, o := range orientKeys(tlOrient) {
		lns := PageText{marks: tlOrient[o]}.toLinesOrient(tol)
		lines = append(lines, lns...)
	}
	return lines
}

// toLinesOrient returns the text and positions in `pt.marks` as a slice of textLine.
// NOTE: This function only works on text lists where all text is the same orientation so it should
// only be called from toLines.
// Caller must sort the text list top-to-bottom, left-to-right (for orientation adjusted so
// that text is horizontal) before calling this function.
func (pt PageText) toLinesOrient(tol float64) []textLine {
	if len(pt.marks) == 0 {
		return []textLine{}
	}
	var marks []TextMark
	var lines []textLine
	var xx []float64
	y := pt.marks[0].orientedStart.Y

	scanning := false

	averageCharWidth := exponAve{}
	wordSpacing := exponAve{}
	lastEndX := 0.0 // lastEndX is pt.marks[i-1].orientedEnd.X

	for _, tm := range pt.marks {
		if tm.orientedStart.Y+tol < y {
			if len(marks) > 0 {
				tl := newLine(y, xx, marks)
				if averageCharWidth.running {
					// FIXME(peterwilliams97): Fix and reinstate combineDiacritics.
					// tl = combineDiacritics(tl, averageCharWidth.ave)
					tl = removeDuplicates(tl, averageCharWidth.ave)
				}
				lines = append(lines, tl)
			}
			marks = []TextMark{}
			xx = []float64{}
			y = tm.orientedStart.Y
			scanning = false
		}

		// Detect text movements that represent spaces on the printed page.
		// We use a heuristic from PdfBox: If the next character starts to the right of where a
		// character after a space at "normal spacing" would start, then there is a space before it.
		// The tricky thing to guess here is the width of a space at normal spacing.
		// We follow PdfBox and use min(deltaSpace, deltaCharWidth).
		deltaSpace := 0.0
		if tm.spaceWidth == 0 {
			deltaSpace = math.MaxFloat64
		} else {
			wordSpacing.update(tm.spaceWidth)
			deltaSpace = wordSpacing.ave * 0.5
		}
		averageCharWidth.update(tm.Width())
		deltaCharWidth := averageCharWidth.ave * 0.3

		isSpace := false
		nextWordX := lastEndX + minFloat(deltaSpace, deltaCharWidth)
		if scanning && !isTextSpace(tm.text) {
			isSpace = nextWordX < tm.orientedStart.X
		}
		common.Log.Trace("tm=%s", tm)
		common.Log.Trace("width=%.2f delta=%.2f deltaSpace=%.2g deltaCharWidth=%.2g",
			tm.Width(), minFloat(deltaSpace, deltaCharWidth), deltaSpace, deltaCharWidth)
		common.Log.Trace("%+q [%.1f, %.1f] lastEndX=%.2f nextWordX=%.2f (%.2f) isSpace=%t",
			tm.text, tm.orientedStart.X, tm.orientedStart.Y, lastEndX, nextWordX,
			nextWordX-tm.orientedStart.X, isSpace)

		if isSpace {
			marks = append(marks, spaceMark)
			xx = append(xx, (lastEndX+tm.orientedStart.X)*0.5)
		}

		// Add the text to the line.
		lastEndX = tm.orientedEnd.X
		marks = append(marks, tm.ToTextMark())
		xx = append(xx, tm.orientedStart.X)
		scanning = true
		common.Log.Trace("lastEndX=%.2f", lastEndX)
	}
	if len(marks) > 0 {
		tl := newLine(y, xx, marks)
		if averageCharWidth.running {
			tl = removeDuplicates(tl, averageCharWidth.ave)
		}
		lines = append(lines, tl)
	}
	return lines
}

// orientKeys returns the keys of `tlOrient` as a sorted slice.
func orientKeys(tlOrient map[int][]textMark) []int {
	keys := []int{}
	for k := range tlOrient {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// exponAve implements an exponential average.
type exponAve struct {
	ave     float64 // Current average value.
	running bool    // Has `ave` been set?
}

// update updates the exponential average `exp`.ave with latest value `x` and returns `exp`.ave.
func (exp *exponAve) update(x float64) float64 {
	if !exp.running {
		exp.ave = x
		exp.running = true
	} else {
		// NOTE(peterwilliams97): 0.5 is a guess. It may be possible to improve average character
		// and space width estimation by tuning this value. It may be that different exponents
		// would work better for character and space estimation.
		exp.ave = (exp.ave + x) * 0.5
	}
	return exp.ave
}

// newLine returns the textLine representation of strings `words` with y coordinate `y` and x
// coordinates `xx` and height `h`.
func newLine(y float64, xx []float64, marks []TextMark) textLine {
	dxList := make([]float64, len(xx)-1)
	for i := 1; i < len(xx); i++ {
		dxList[i-1] = xx[i] - xx[i-1]
	}
	return textLine{
		x:      xx[0],
		y:      y,
		dxList: dxList,
		marks:  marks,
	}
}

// removeDuplicates returns `tl` with duplicate characters removed. `charWidth` is the average
// character width for the line.
func removeDuplicates(tl textLine, charWidth float64) textLine {
	if len(tl.dxList) == 0 || len(tl.marks) == 0 {
		return tl
	}
	// NOTE(peterwilliams97) 0.3 is a guess. It may be possible to tune this to a better value.
	tol := charWidth * 0.3
	marks := []TextMark{tl.marks[0]}
	var dxList []float64

	tm0 := tl.marks[0]
	for i, dx := range tl.dxList {
		tm := tl.marks[i+1]
		if tm.Text != tm0.Text || dx > tol {
			marks = append(marks, tm)
			dxList = append(dxList, dx)
		}
		tm0 = tm
	}
	return textLine{
		x:      tl.x,
		y:      tl.y,
		dxList: dxList,
		marks:  marks,
	}
}

// combineDiacritics returns `line` with diacritics close to characters combined with the characters.
// `charWidth` is the average character width for the line.
// We have to do this because PDF can render diacritics separately to the characters they attach to
// in extracted text.
func combineDiacritics(tl textLine, charWidth float64) textLine {
	if len(tl.dxList) == 0 || len(tl.marks) == 0 {
		return tl
	}
	// NOTE(peterwilliams97) 0.2 is a guess. It may be possible to tune this to a better value.
	tol := charWidth * 0.2
	common.Log.Trace("combineDiacritics: charWidth=%.2f tol=%.2f", charWidth, tol)

	var marks []TextMark
	var dxList []float64
	tm := marks[0]
	w, c := countDiacritic(tm.Text)
	delta := 0.0
	dx0 := 0.0
	parts := []string{w}
	numChars := c

	for i, dx := range tl.dxList {
		tm = marks[i+1]
		w, c := countDiacritic(tm.Text)
		if numChars+c <= 1 && delta+dx <= tol {
			if len(parts) == 0 {
				dx0 = dx
			} else {
				delta += dx
			}
			parts = append(parts, w)
			numChars += c
		} else {
			if len(parts) > 0 {
				if len(marks) > 0 {
					dxList = append(dxList, dx0)
				}
				tm.Text = combine(parts)
				marks = append(marks, tm)
			}
			parts = []string{w}
			numChars = c
			dx0 = dx
			delta = 0.0
		}
	}
	if len(parts) > 0 {
		if len(marks) > 0 {
			dxList = append(dxList, dx0)
		}
		tm.Text = combine(parts)
		marks = append(marks, tm)
	}
	if len(marks) != len(dxList)+1 {
		common.Log.Error("Inconsistent: \nwords=%d \ndxList=%d %.2f",
			len(marks), len(dxList), dxList)
		return tl
	}
	return textLine{
		x:      tl.x,
		y:      tl.y,
		dxList: dxList,
		marks:  marks,
	}
}

// combine combines any diacritics in `parts` with the single non-diacritic character in `parts`.
func combine(parts []string) string {
	if len(parts) == 1 {
		// Must be a non-diacritic.
		return parts[0]
	}

	// We need to put the diacritics before the non-diacritic for NFKC normalization to work.
	diacritic := map[string]bool{}
	for _, w := range parts {
		r := []rune(w)[0]
		diacritic[w] = unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Sk, r)
	}
	sort.SliceStable(parts, func(i, j int) bool { return !diacritic[parts[i]] && diacritic[parts[j]] })

	// Construct the NFKC-normalized concatenation of the diacritics and the non-diacritic.
	for i, w := range parts {
		parts[i] = strings.TrimSpace(norm.NFKC.String(w))
	}
	return strings.Join(parts, "")
}

// countDiacritic returns the combining diacritic version of `w` (usually itself) and the number of
// non-diacritics in `w` (0 or 1).
func countDiacritic(w string) (string, int) {
	runes := []rune(w)
	if len(runes) != 1 {
		return w, 1
	}
	r := runes[0]
	c := 1
	if (unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Sk, r)) &&
		r != '\'' && r != '"' && r != '`' {
		c = 0
	}
	if w2, ok := diacritics[r]; ok {
		c = 0
		w = w2
	}
	return w, c
}

// diacritics is a map of diacritic characters that are not classified as unicode.Mn or unicode.Sk
// and the corresponding unicode.Mn or unicode.Sk characters. This map was copied from PdfBox.
// (https://svn.apache.org/repos/asf/pdfbox/trunk/pdfbox/src/main/java/org/apache/pdfbox/text/TextPosition.java)
var diacritics = map[rune]string{
	0x0060: "\u0300",
	0x02CB: "\u0300",
	0x0027: "\u0301",
	0x02B9: "\u0301",
	0x02CA: "\u0301",
	0x005e: "\u0302",
	0x02C6: "\u0302",
	0x007E: "\u0303",
	0x02C9: "\u0304",
	0x00B0: "\u030A",
	0x02BA: "\u030B",
	0x02C7: "\u030C",
	0x02C8: "\u030D",
	0x0022: "\u030E",
	0x02BB: "\u0312",
	0x02BC: "\u0313",
	0x0486: "\u0313",
	0x055A: "\u0313",
	0x02BD: "\u0314",
	0x0485: "\u0314",
	0x0559: "\u0314",
	0x02D4: "\u031D",
	0x02D5: "\u031E",
	0x02D6: "\u031F",
	0x02D7: "\u0320",
	0x02B2: "\u0321",
	0x02CC: "\u0329",
	0x02B7: "\u032B",
	0x02CD: "\u0331",
	0x005F: "\u0332",
	0x204E: "\u0359",
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
			var names []string
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
// it doesn't. Accesses page resources directly (not cached).
func (to *textObject) getFontDirect(name string) (*model.PdfFont, error) {
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

// getFontDict returns the font dict with key `name` if it exists in the page's or form's Font
// resources or an error if it doesn't.
func (to *textObject) getFontDict(name string) (fontObj core.PdfObject, err error) {
	resources := to.resources
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
