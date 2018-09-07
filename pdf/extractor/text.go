/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// The current version of this file is a halfway step from the old UniDoc text extractor to a
// full PDF font parser.
// We will soon implement all the functions marked as `Not implemented yet`.

package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

// ExtractText processes and extracts all text data in content streams and returns as a string.
// It takes into account character encoding in the PDF file, which is decoded by
// CharcodeBytesToUnicode.
// The text is processed linearly e.g. in the order in which it appears. A best effort is done to
// add spaces and newlines.
func (e *Extractor) ExtractText() (string, error) {
	text, _, _, err := e.ExtractText2()
	return text, err
}

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
				to.renderRawText("\n")
			case "TD": // Move text location and set leading
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
				charcodes, ok := core.GetStringBytes(op.Params[0])
				if !ok {
					common.Log.Debug("ERROR: \" op=%s GetStringBytes failed", op)
					return core.ErrTypeError
				}
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
		common.Log.Error("ERROR: Processing: err=%v", err)
	}
	return textList, state.numChars, state.numMisses, err
}

//
// Text operators
//

// moveText "Td" Moves start of text by `tx`,`ty`
// Move to the start of the next line, offset from the start of the current line by (tx, ty).
// tx and ty are in unscaled text space units.
func (to *textObject) moveText(tx, ty float64) {
	// Not implemented yet
}

// moveTextSetLeading "TD" Move text location and set leading
// Move to the start of the next line, offset from the start of the current line by (tx, ty). As a
// side effect, this operator shall set the leading parameter in the text state. This operator shall
// have the same effect as this code:
//  −ty TL
//  tx ty Td
func (to *textObject) moveTextSetLeading(tx, ty float64) {
	// Not implemented yet
	// The following is supposed to be equivalent to the existing Unidoc implementation.
	if tx > 0 {
		to.renderRawText(" ")
	}
	if ty < 0 {
		// TODO: More flexible space characters?
		to.renderRawText("\n")
	}
}

// nextLine "T*"" Moves start of text `Line` to next text line
// Move to the start of the next line. This operator has the same effect as the code
//    0 -Tl Td
// where Tl denotes the current leading parameter in the text state. The negative of Tl is used
// here because Tl is the text leading expressed as a positive number. Going to the next line
// entails decreasing the y coordinate. (page 250)
func (to *textObject) nextLine() {
	// Not implemented yet
}

// setTextMatrix "Tm"
// Set the text matrix, Tm, and the text line matrix, Tlm to the Matrix specified by the 6 numbers
// in `f`  (page 250)
func (to *textObject) setTextMatrix(f []float64) {
	// Not implemented yet
	// The following is supposed to be equivalent to the existing Unidoc implementation.
	tx, ty := f[4], f[5]
	if to.yPos == -1 {
		to.yPos = tx
	} else if to.yPos > ty {
		to.renderRawText("\n")
		to.xPos, to.yPos = tx, ty
		return
	}
	if to.xPos == -1 {
		to.xPos = tx
	} else if to.xPos < ty {
		to.renderRawText("\t")
		to.xPos = tx
	}
}

// showText "Tj" Show a text string
func (to *textObject) showText(charcodes []byte) error {
	return to.renderText(charcodes)
}

// showTextAdjusted "TJ" Show text with adjustable spacing
func (to *textObject) showTextAdjusted(args *core.PdfObjectArray) error {
	for _, o := range args.Elements() {
		switch o.(type) {
		case *core.PdfObjectFloat, *core.PdfObjectInteger:
			// Not implemented yet
			// The following is supposed to be equivalent to the existing Unidoc implementation.
			v, _ := core.GetNumberAsFloat(o)
			if v < -100 {
				to.renderRawText("\n")
			}
		case *core.PdfObjectString:
			charcodes, ok := core.GetStringBytes(o)
			if !ok {
				common.Log.Debug("ERROR: showTextAdjusted: GetStringBytes failed. args=%+v", args)
				return core.ErrTypeError
			}
			err := to.renderText(charcodes)
			if err != nil {
				common.Log.Debug("showTextAdjusted: renderText failed. args=%+v err=%v", args, err)
				return err
			}
		default:
			common.Log.Debug("showTextAdjusted. Unexpected type args=%+v", args)
			return core.ErrTypeError
		}
	}
	return nil
}

// setTextLeading "TL" Set text leading
func (to *textObject) setTextLeading(y float64) {
	// Not implemented yet
}

// setCharSpacing "Tc" Set character spacing
func (to *textObject) setCharSpacing(x float64) {
	// Not implemented yet
}

// setFont "Tf" Set font
func (to *textObject) setFont(name string, size float64) error {
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
	// to.State.Tfs = size
	return nil
}

// setTextRenderMode "Tr" Set text rendering mode
func (to *textObject) setTextRenderMode(mode int) {
	// Not implemented yet
}

// setTextRise "Ts" Set text rise
func (to *textObject) setTextRise(y float64) {
	// Not implemented yet
}

// setWordSpacing "Tw" Set word spacing
func (to *textObject) setWordSpacing(y float64) {
	// Not implemented yet
}

// setHorizScaling "Tz" Set horizontal scaling
func (to *textObject) setHorizScaling(y float64) {
	// Not implemented yet
}

// floatParam returns the single float parameter of operatr `op`, or an error if it doesn't have
// a single float parameter or we aren't in a text stream.
func floatParam(op *contentstream.ContentStreamOperation) (float64, error) {
	if len(op.Params) != 1 {
		err := errors.New("Incorrect parameter count")
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
		common.Log.Debug("%#q operand outside text", op.Operand)
		return false, nil
	}
	if numParams >= 0 {
		if len(op.Params) != numParams {
			if hard {
				err = errors.New("Incorrect parameter count")
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

// pop pops and returns the element on the top of the font stack if there is one, or nil  if there isn't.
func (fontStack *fontStacker) pop() *model.PdfFont {
	if fontStack.empty() {
		return nil
	}
	font := (*fontStack)[len(*fontStack)-1]
	*fontStack = (*fontStack)[:len(*fontStack)-1]
	return font
}

// peek returns the element on the top of the font stack if there is one, or nil if there isn't.
func (fontStack *fontStacker) peek() *model.PdfFont {
	if fontStack.empty() {
		return nil
	}
	return (*fontStack)[len(*fontStack)-1]
}

// get returns the `idx`'th element of the font stack if there is one, or nil if there isn't.
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
	// Tc    float64        // Character spacing. Unscaled text space units.
	// Tw    float64        // Word spacing. Unscaled text space units.
	// Th    float64        // Horizontal scaling
	// Tl    float64        // Leading. Unscaled text space units. Used by TD,T*,'," see Table 108
	// Tfs   float64        // Text font size
	// Tmode RenderMode     // Text rendering mode
	// Trise float64        // Text rise. Unscaled text space units. Set by Ts
	Tf *model.PdfFont // Text font
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
//        | Tfs x Th   0      0 |
// Trm  = | 0         Tfs     0 | × Tm × CTM
//        | 0         Trise   1 |

// textObject represents a PDF text object.
type textObject struct {
	e         *Extractor
	gs        contentstream.GraphicsState
	fontStack *fontStacker
	State     *textState
	// Tm    contentstream.Matrix // Text matrix. For the character pointer.
	// Tlm   contentstream.Matrix // Text line matrix. For the start of line pointer.
	Texts []XYText // Text gets written here.

	// These fields are used to implement existing UniDoc behaviour.
	xPos, yPos float64
}

// newTextState returns a default textState
func newTextState() textState {
	// Not implemented yet
	return textState{}
}

// newTextObject returns a default textObject
func newTextObject(e *Extractor, gs contentstream.GraphicsState, state *textState,
	fontStack *fontStacker) *textObject {
	return &textObject{
		e:         e,
		gs:        gs,
		fontStack: fontStack,
		State:     state,
		// Tm:    contentstream.IdentityMatrix(),
		// Tlm:   contentstream.IdentityMatrix(),
	}
}

// renderRawText writes `text` directly to the extracted text
func (to *textObject) renderRawText(text string) {
	to.Texts = append(to.Texts, XYText{text})
}

// renderText emits byte array `data` to the calling program
func (to *textObject) renderText(data []byte) error {
	text := ""
	if len(*to.fontStack) == 0 {
		common.Log.Debug("ERROR: No font defined. data=%#q", string(data))
		text = string(data)
		return model.ErrNoFont
	}
	font := to.fontStack.peek()
	var numChars, numMisses int
	text, numChars, numMisses = font.CharcodeBytesToUnicode(data)
	to.State.numChars += numChars
	to.State.numMisses += numMisses

	to.Texts = append(to.Texts, XYText{text})
	return nil
}

// XYText represents text and its position in device coordinates
type XYText struct {
	Text string
	// Position and rendering fields. Not implemented yet
}

// String returns a string describing `t`
func (t *XYText) String() string {
	return truncate(t.Text, 100)
}

// TextList is a list of texts and their position on a pdf page
type TextList []XYText

func (tl *TextList) Length() int {
	return len(*tl)
}

// ToText returns the contents of `tl` as a single string
func (tl *TextList) ToText() string {
	var buf bytes.Buffer
	for _, t := range *tl {
		buf.WriteString(t.Text)
	}
	procBuf(&buf)
	return buf.String()
}

// getFont returns the font named `name` if it exists in the page's resources or an error if it
// doesn't.
func (to *textObject) getFont(name string) (*model.PdfFont, error) {

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
		common.Log.Debug("getFont: NewPdfFontFromPdfObject failed. name=%#q err=%v", name, err)
	}
	return font, err
}

// getFontDict returns the font object called `name` if it exists in the page's Font resources or
// an error if it doesn't.
// XXX: TODO: Can we cache font values?
func (to *textObject) getFontDict(name string) (fontObj core.PdfObject, err error) {
	resources := to.e.resources
	if resources == nil {
		common.Log.Debug("getFontDict. No resources. name=%#q", name)
		return nil, nil
	}
	fontObj, found := resources.GetFontByName(core.PdfObjectName(name))
	if !found {
		common.Log.Debug("ERROR: getFontDict: Font not found: name=%#q", name)
		return nil, errors.New("Font not in resources")
	}
	return fontObj, nil
}

// getCharMetrics returns the character metrics for the code points in `text1` for font `font`.
func getCharMetrics(font *model.PdfFont, text string) (metrics []fonts.CharMetrics, err error) {
	encoder := font.Encoder()
	if encoder == nil {
		return nil, errors.New("No font encoder")
	}
	for _, r := range text {
		glyph, found := encoder.RuneToGlyph(r)
		if !found {
			common.Log.Debug("Error! Glyph not found for rune=%s", r)
			glyph = "space"
		}
		m, ok := font.GetGlyphCharMetrics(glyph)
		if !ok {
			common.Log.Debug("ERROR: Metrics not found for rune=%+v glyph=%#q", r, glyph)
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}
