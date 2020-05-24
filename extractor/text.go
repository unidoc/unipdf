/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"bytes"
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
	// procBuf(pt)

	return pt, numChars, numMisses, err
}

// extractPageText returns the text contents of content stream `e` and resouces `resources` as a
// PageText.
// This can be called on a page or a form XObject.
func (e *Extractor) extractPageText(contents string, resources *model.PdfPageResources, level int) (
	*PageText, int, int, error) {
	common.Log.Trace("extractPageText: level=%d", level)
	pageText := &PageText{pageSize: e.mediaBox}
	state := newTextState(e.mediaBox)
	fontStack := fontStacker{}
	to := newTextObject(e, resources, contentstream.GraphicsState{}, &state, &fontStack)
	var inTextObj bool

	// Uncomment the following 3 statements to log the content stream.
	// common.Log.Info("contents* %d -----------------------------", len(contents))
	// fmt.Println(contents)
	// common.Log.Info("contents+ -----------------------------")

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
					common.Log.Trace("Save font state: %s\n→%s\n%s",
						fontStack.peek(), state.tfont, fontStack.String())
					fontStack.push(state.tfont)
				}
			case "Q":
				if !fontStack.empty() {
					common.Log.Trace("Restore font state: %s\n→%s\n%s",
						fontStack.peek(), fontStack.get(-2), fontStack.String())
					fontStack.pop()
				}
				if len(fontStack) >= 2 {
					common.Log.Trace("Restore font state: %s\n→%s\n%s",
						state.tfont, fontStack.peek(), fontStack.String())
					state.tfont = fontStack.pop()
				}
			case "BT": // Begin text
				// Begin a text object, initializing the text matrix, Tm, and
				// the text line matrix, Tlm, to the identity matrix. Text
				// objects shall not be nested. A second BT shall not appear
				// before an ET. However, if that happens, all existing marks
				// are added to the  page marks, in order to avoid losing content.
				if inTextObj {
					common.Log.Debug("BT called while in a text object")
					pageText.marks = append(pageText.marks, to.marks...)
				}
				inTextObj = true
				to = newTextObject(e, resources, gs, &state, &fontStack)
			case "ET": // End Text
				// End text object, discarding text matrix. If the current
				// text object contains text marks, they are added to the
				// page text marks collection.
				// The ET operator should always have a matching BT operator.
				// However, if ET appears outside of a text object, the behavior
				// does not change: the text matrices are discarded and all
				// existing marks in the text object are added to the page marks.
				if !inTextObj {
					common.Log.Debug("ET called outside of a text object")
				}
				inTextObj = false
				pageText.marks = append(pageText.marks, to.marks...)
				to.reset()
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
				if ok, err := to.checkOp(op, 3, true); !ok {
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
				if len(op.Params) == 0 {
					common.Log.Debug("ERROR: expected XObject name operand for Do operator. Got %+v.", op.Params)
					return core.ErrRangeError
				}

				// Get XObject name.
				name, ok := core.GetName(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: invalid Do operator XObject name operand: %+v.", op.Params[0])
					return core.ErrTypeError
				}

				_, xtype := resources.GetXObjectByName(*name)
				if xtype != model.XObjectTypeForm {
					break
				}
				// Only process each form once.
				formResult, ok := e.formResults[name.String()]
				if !ok {
					xform, err := resources.GetXObjectFormByName(*name)
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
					e.formResults[name.String()] = formResult
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
	to.logCursor()
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
			to.logCursor()
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
	tc       float64        // Character spacing. Unscaled text space units.
	tw       float64        // Word spacing. Unscaled text space units.
	th       float64        // Horizontal scaling.
	tl       float64        // Leading. Unscaled text space units. Used by TD,T*,'," see Table 108.
	tfs      float64        // Text font size.
	tmode    RenderMode     // Text rendering mode.
	trise    float64        // Text rise. Unscaled text space units. Set by Ts.
	tfont    *model.PdfFont // Text font.
	mediaBox model.PdfRectangle
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
func newTextState(mediaBox model.PdfRectangle) textState {
	return textState{
		th:       100,
		tmode:    RenderModeFill,
		mediaBox: mediaBox,
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

// reset sets the text matrix `Tm` and the text line matrix `Tlm` of the text
// object to the identity matrix. In addition, the marks collection is cleared.
func (to *textObject) reset() {
	to.tm = transform.IdentityMatrix()
	to.tlm = transform.IdentityMatrix()
	to.marks = nil
	to.logCursor()
}

// logCursor is for debugging only. Remove !@#$
func (to *textObject) logCursor() {
	return
	state := to.state
	tfs := state.tfs
	th := state.th / 100.0
	stateMatrix := transform.NewMatrix(
		tfs*th, 0,
		0, tfs,
		0, state.trise)
	trm := to.gs.CTM.Mult(to.tm).Mult(stateMatrix)
	cur := translation(trm)
	common.Log.Info("showTrm: %s cur=%.2f tm=%.2f CTM=%.2f",
		fileLine(1, false), cur, to.tm, to.gs.CTM)
}

// renderText processes and renders byte array `data` for extraction purposes.
// It extracts textMarks based the charcodes in `data` and the currect text and graphics states
// are tracked in `to`.
func (to *textObject) renderText(data []byte) error {
	font := to.getCurrentFont()
	charcodes := font.BytesToCharcodes(data)
	runeSlices, numChars, numMisses := font.CharcodesToRuneSlices(charcodes)
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
	common.Log.Trace("spaceWidth=%.2f text=%q font=%s fontSize=%.1f", spaceWidth, runeSlices, font, tfs)

	stateMatrix := transform.NewMatrix(
		tfs*th, 0,
		0, tfs,
		0, state.trise)

	common.Log.Trace("renderText: %d codes=%+v runes=%q", len(charcodes), charcodes, runeSlices)

	for i, r := range runeSlices {
		if len(r) == 1 && r[0] == '\x00' {
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
		if len(r) == 1 && r[0] == 32 {
			w = state.tw
		}

		m, ok := font.GetCharMetrics(code)
		if !ok {
			common.Log.Debug("ERROR: No metric for code=%d r=0x%04x=%+q %s", code, r, r, font)
			return fmt.Errorf("no char metrics: font=%s code=%d", font.String(), code)
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
		end := to.gs.CTM.Mult(to.tm).Mult(td0)

		common.Log.Trace("end:\n\tCTM=%s\n\t tm=%s\n\ttd0=%s\n\t → %s xlat=%s",
			to.gs.CTM, to.tm, td0, end, translation(end))

		mark, onPage := to.newTextMark(
			string(r),
			trm,
			translation(end),
			math.Abs(spaceWidth*trm.ScalingFactorX()),
			font,
			to.state.tc)
		if !onPage {
			common.Log.Debug("Text mark outside page. Skipping")
			continue
		}
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
		if i != len(runeSlices)-1 {
			to.logCursor()
		}
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

// isTextSpace returns true if `text` contains nothing but space code points.
func isTextSpace(text string) bool {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// PageText represents the layout of text on a device page.
type PageText struct {
	marks     []textMark // Texts and their positions on a PDF page.
	viewText  string     // Extracted page text.
	viewMarks []TextMark // Public view of `marks`.
	pageSize  model.PdfRectangle
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

// computeViews processes the page TextMarks sorting by position and populates `pt.viewText` and
// `pt.viewMarks` which represent the text and marks in the order which it is read on the page.
// The comments above the TextMark definition describe how to use the []TextMark to
// maps substrings of the page text to locations on the PDF page.
func (pt *PageText) computeViews() {
	common.Log.Trace("ToTextLocation: %d elements", len(pt.marks))
	paras := makeTextPage(pt.marks, pt.pageSize, 0)
	b := new(bytes.Buffer)
	paras.writeText(b)
	pt.viewText = b.String()
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
	var bbox model.PdfRectangle
	found := false
	for _, tm := range ma.marks {
		if tm.Meta || isTextSpace(tm.Text) {
			continue
		}
		if found {
			bbox = rectUnion(bbox, tm.BBox)
		} else {
			bbox = tm.BBox
			found = true
		}
	}
	return bbox, found
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
	count int64
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
	return fmt.Sprintf("{@%04d TextMark: %d %q=%02x (%5.1f, %5.1f) (%5.1f, %5.1f) %s%s}",
		tm.count, tm.Offset, tm.Text, []rune(tm.Text), b.Llx, b.Lly, b.Urx, b.Ury, font, meta)
}

// spaceMark is a special TextMark used for spaces.
var spaceMark = TextMark{
	Text:     "[X]",
	Original: " ",
	Meta:     true,
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
