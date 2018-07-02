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

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

// ExtractText processes and extracts all text data in content streams and returns as a string.
// Takes into account character encoding via CMaps in the PDF file.
// The text is processed linearly e.g. in the order in which it appears. A best effort is done to
// add spaces and newlines.
func (e *Extractor) ExtractText() (string, error) {
	textList, err := e.ExtractXYText()
	if err != nil {
		return "", err
	}
	return textList.ToText(), nil
}

// ExtractXYText returns the text contents of `e` as a TextList.
func (e *Extractor) ExtractXYText() (*TextList, error) {
	textList := &TextList{}
	state := newTextState()
	var to *TextObject

	cstreamParser := contentstream.NewContentStreamParser(e.contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		common.Log.Debug("ExtractXYText: parse failed. err=%v", err)
		return textList, err
	}

	// fmt.Println("========================= xxx =========================")
	// fmt.Printf("%s\n", e.contents)
	// fmt.Println("========================= ||| =========================")
	processor := contentstream.NewContentStreamProcessor(*operations)

	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState,
			resources *model.PdfPageResources) error {

			operand := op.Operand
			// common.Log.Debug("++Operand: %s", op.String())

			switch operand {
			case "BT": // Begin text
				// Begin a text object, initializing the text matrix, Tm, and the text line matrix, Tlm,
				// to the identity matrix. Text objects shall not be nested; a second BT shall not appear
				// before an ET.
				if to != nil {
					common.Log.Debug("BT called while in a text object")
				}
				to = newTextObject(e, gs, &state)
			case "ET": // End Text
				*textList = append(*textList, to.Texts...)
				to = nil
			case "T*": // Move to start of next text line
				to.nextLine()
			case "Td": // Move text location
				if ok, err := checkOp(op, to, 2, true); !ok {
					return err
				}
				to.renderRawText("\n")
			case "TD": // Move text location and set leading
				if ok, err := checkOp(op, to, 2, true); !ok {
					return err
				}
				x, y, err := toFloatXY(op.Params)
				if err != nil {
					return err
				}
				to.moveTextSetLeading(x, y)
			case "Tj": // Show text
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				charcodes, err := GetStringBytes(op.Params[0])
				if err != nil {
					return err
				}
				return to.showText(charcodes)
			case "TJ": // Show text with adjustable spacing
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				args, err := GetArray(op.Params[0])
				if err != nil {
					return err
				}
				return to.showTextAdjusted(args)
			case "'": // Move to next line and show text
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				charcodes, err := GetStringBytes(op.Params[0])
				if err != nil {
					return err
				}
				to.nextLine()
				return to.showText(charcodes)
			case `"`: // Set word and character spacing, move to next line, and show text
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				charcodes, err := GetStringBytes(op.Params[0])
				if err != nil {
					return err
				}
				to.nextLine()
				return to.showText(charcodes)
			case "TL": // Set text leading
				ok, y, err := checkOpFloat(op, to)
				if !ok || err != nil {
					return err
				}
				to.setTextLeading(y)
			case "Tc": // Set character spacing
				ok, y, err := checkOpFloat(op, to)
				if !ok || err != nil {
					return err
				}
				to.setCharSpacing(y)
			case "Tf": // Set font
				if ok, err := checkOp(op, to, 2, true); !ok {
					return err
				}
				name, err := GetName(op.Params[0])
				if err != nil {
					return err
				}
				size, err := GetNumberAsFloat(op.Params[1])
				if err != nil {
					return err
				}
				err = to.setFont(name, size)
				if err == model.ErrUnsupportedFont {
					common.Log.Debug("Swallow error. err=%v", err)
					err = nil
				}
				if err != nil {
					return err
				}
			case "Tm": // Set text matrix
				if ok, err := checkOp(op, to, 6, true); !ok {
					return err
				}
				floats, err := model.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}
				to.setTextMatrix(floats)
			case "Tr": // Set text rendering mode
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				mode, err := GetInteger(op.Params[0])
				if err != nil {
					return err
				}
				to.setTextRenderMode(mode)
			case "Ts": // Set text rise
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				y, err := GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}
				to.setTextRise(y)
			case "Tw": // Set word spacing
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				y, err := GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}
				to.setWordSpacing(y)
			case "Tz": // Set horizontal scaling
				if ok, err := checkOp(op, to, 1, true); !ok {
					return err
				}
				y, err := GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}
				to.setHorizScaling(y)
			}

			return nil
		})

	err = processor.Process(e.resources)
	if err == model.ErrUnsupportedFont {
		common.Log.Debug("Swallow error. err=%v", err)
		err = nil
	}
	if err != nil {
		common.Log.Error("ERROR: Processing: err=%v", err)
		return textList, err
	}

	return textList, nil
}

//
// Text operators
//

// moveText "Td" Moves start of text by `tx`,`ty`
// Move to the start of the next line, offset from the start of the current line by (tx, ty).
// tx and ty are in unscaled text space units.
func (to *TextObject) moveText(tx, ty float64) {
	// Not implemented yet
}

// moveTextSetLeading "TD" Move text location and set leading
// Move to the start of the next line, offset from the start of the current line by (tx, ty). As a
// side effect, this operator shall set the leading parameter in the text state. This operator shall
// have the same effect as this code:
//  −ty TL
//  tx ty Td
func (to *TextObject) moveTextSetLeading(tx, ty float64) {
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
func (to *TextObject) nextLine() {
	// Not implemented yet
}

// setTextMatrix "Tm"
// Set the text matrix, Tm, and the text line matrix, Tlm to the Matrix specified by the 6 numbers
// in `f`  (page 250)
func (to *TextObject) setTextMatrix(f []float64) {
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
func (to *TextObject) showText(charcodes []byte) error {
	return to.renderText(charcodes)
}

// showTextAdjusted "TJ" Show text with adjustable spacing
func (to *TextObject) showTextAdjusted(args []PdfObject) error {
	for _, o := range args {
		switch o.(type) {
		case *PdfObjectFloat, *PdfObjectInteger:
			// Not implemented yet
			// The following is supposed to be equivalent to the existing Unidoc implementation.
			v, _ := GetNumberAsFloat(o)
			if v < -100 {
				to.renderRawText("\n")
			}
		case *PdfObjectString:
			charcodes, err := GetStringBytes(o)
			if err != nil {
				common.Log.Debug("showTextAdjusted args=%+v err=%v", args, err)
				return err
			}
			err = to.renderText(charcodes)
			if err != nil {
				return err
			}
		default:
			common.Log.Debug("showTextAdjusted. Unexpected type args=%+v", args)
			return ErrTypeCheck
		}
	}
	return nil
}

// setTextLeading "TL" Set text leading
func (to *TextObject) setTextLeading(y float64) {
	// Not implemented yet
}

// setCharSpacing "Tc" Set character spacing
func (to *TextObject) setCharSpacing(x float64) {
	// Not implemented yet
}

// setFont "Tf" Set font
func (to *TextObject) setFont(name string, size float64) error {
	font, err := to.getFont(name)
	if err != nil {
		return err
	}
	to.State.Tf = font
	// to.State.Tfs = size
	return nil
}

// setTextRenderMode "Tr" Set text rendering mode
func (to *TextObject) setTextRenderMode(mode int) {
	// Not implemented yet
}

// setTextRise "Ts" Set text rise
func (to *TextObject) setTextRise(y float64) {
	// Not implemented yet
}

// setWordSpacing "Tw" Set word spacing
func (to *TextObject) setWordSpacing(y float64) {
	// Not implemented yet
}

// setHorizScaling "Tz" Set horizontal scaling
func (to *TextObject) setHorizScaling(y float64) {
	// Not implemented yet
}

// Operator validation
func checkOpFloat(op *contentstream.ContentStreamOperation, to *TextObject) (ok bool, x float64, err error) {
	if ok, err = checkOp(op, to, 1, true); !ok {
		return
	}
	x, err = GetNumberAsFloat(op.Params[0])
	return
}

// checkOp returns true if we are in a text stream and `op` has `numParams` params
// If `hard` is true and the number of params don't match then an error is returned
func checkOp(op *contentstream.ContentStreamOperation, to *TextObject, numParams int, hard bool) (ok bool, err error) {
	if to == nil {
		common.Log.Debug("%#q operand outside text", op.Operand)
		return
	}
	if numParams >= 0 {
		if len(op.Params) != numParams {
			if hard {
				err = errors.New("Incorrect parameter count")
			}
			common.Log.Debug("Error: %#q should have %d input params, got %d %+v",
				op.Operand, numParams, len(op.Params), op.Params)
			return
		}
	}
	ok = true
	return
}

// 9.3 Text State Parameters and Operators (page 243)
// Some of these parameters are expressed in unscaled text space units. This means that they shall
// be specified in a coordinate system that shall be defined by the text matrix, Tm but shall not be
// scaled by the font size parameter, Tfs.
type TextState struct {
	// Tc    float64        // Character spacing. Unscaled text space units.
	// Tw    float64        // Word spacing. Unscaled text space units.
	// Th    float64        // Horizontal scaling
	// Tl    float64        // Leading. Unscaled text space units. Used by TD,T*,'," see Table 108
	// Tfs   float64        // Text font size
	// Tmode RenderMode     // Text rendering mode
	// Trise float64        // Text rise. Unscaled text space units. Set by Ts
	Tf *model.PdfFont // Text font
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
//
type TextObject struct {
	e     *Extractor
	gs    contentstream.GraphicsState
	State *TextState
	// Tm    contentstream.Matrix // Text matrix. For the character pointer.
	// Tlm   contentstream.Matrix // Text line matrix. For the start of line pointer.
	Texts []XYText // Text gets written here.

	// These fields are used to implement existing UniDoc behaviour.
	xPos, yPos float64
}

// newTextState returns a default TextState
func newTextState() TextState {
	// Not implemented yet
	return TextState{}
}

// newTextObject returns a default TextObject
func newTextObject(e *Extractor, gs contentstream.GraphicsState, state *TextState) *TextObject {
	return &TextObject{
		e:     e,
		gs:    gs,
		State: state,
		// Tm:    contentstream.IdentityMatrix(),
		// Tlm:   contentstream.IdentityMatrix(),
	}
}

// renderRawText writes `text` directly to the extracted text
func (to *TextObject) renderRawText(text string) {
	to.Texts = append(to.Texts, XYText{text})
}

// renderText emits byte array `data` to the calling program
func (to *TextObject) renderText(data []byte) (err error) {
	text := ""
	if to.State.Tf == nil {
		common.Log.Debug("ERROR: No font defined. data=%#q", string(data))
		text = string(data)
		err = model.ErrBadText
	} else {
		text, err = to.State.Tf.CharcodeBytesToUnicode(data)
	}
	to.Texts = append(to.Texts, XYText{text})
	return
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
// doesn't
func (to *TextObject) getFont(name string) (*model.PdfFont, error) {
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
// an error if it doesn't
// XXX: TODO: Can we cache font values?
func (to *TextObject) getFontDict(name string) (fontObj PdfObject, err error) {
	resources := to.e.resources
	if resources == nil {
		common.Log.Debug("getFontDict. No resources. name=%#q", name)
		return
	}

	fontObj, found := resources.GetFontByName(PdfObjectName(name))
	if !found {
		err = errors.New("Font not in resources")
		common.Log.Debug("ERROR: getFontDict: Font not found: name=%#q err=%v", name, err)
		return
	}
	fontObj = TraceToDirectObject(fontObj)
	return
}

// getCharMetrics returns the character metrics for the code points in `text1` for font `font`
func getCharMetrics(font *model.PdfFont, text string) (metrics []fonts.CharMetrics, err error) {
	encoder := font.Encoder()
	if encoder == nil {
		err = errors.New("No font encoder")
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
	return
}
