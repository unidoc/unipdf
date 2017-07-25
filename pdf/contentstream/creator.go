/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"math"

	. "github.com/unidoc/unidoc/pdf/core"
)

type ContentCreator struct {
	operands ContentStreamOperations
}

func NewContentCreator() *ContentCreator {
	creator := &ContentCreator{}
	creator.operands = ContentStreamOperations{}
	return creator
}

// Get the list of operations.
func (this *ContentCreator) Operations() *ContentStreamOperations {
	return &this.operands
}

// Convert a set of content stream operations to a content stream byte presentation, i.e. the kind that can be
// stored as a PDF stream or string format.
func (this *ContentCreator) Bytes() []byte {
	return this.operands.Bytes()
}

// Same as Bytes() except returns as a string for convenience.
func (this *ContentCreator) String() string {
	return string(this.operands.Bytes())
}

/* Graphics state operators. */

// Save the current graphics state on the stack - push.
func (this *ContentCreator) Add_q() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "q"
	this.operands = append(this.operands, &op)
	return this
}

// Restore the most recently stored state from the stack - pop.
func (this *ContentCreator) Add_Q() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Q"
	this.operands = append(this.operands, &op)
	return this
}

// Display XObject - image or form.
func (this *ContentCreator) Add_Do(name PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Do"
	op.Params = makeParamsFromNames([]PdfObjectName{name})
	this.operands = append(this.operands, &op)
	return this
}

// Modify the current transformation matrix (ctm).
func (this *ContentCreator) Add_cm(a, b, c, d, e, f float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "cm"
	op.Params = makeParamsFromFloats([]float64{a, b, c, d, e, f})
	this.operands = append(this.operands, &op)
	return this
}

// Convenience function for generating a cm operation to translate the transformation matrix.
func (this *ContentCreator) Translate(tx, ty float64) *ContentCreator {
	return this.Add_cm(1, 0, 0, 1, tx, ty)
}

// Convenience function for generating a cm command to scale the transformation matrix.
func (this *ContentCreator) Scale(sx, sy float64) *ContentCreator {
	return this.Add_cm(sx, 0, 0, sy, 0, 0)
}

// Convenience function for generating a cm command to rotate transformation matrix.
func (this *ContentCreator) RotateDeg(angle float64) *ContentCreator {
	u1 := math.Cos(angle * math.Pi / 180.0)
	u2 := math.Sin(angle * math.Pi / 180.0)
	u3 := -math.Sin(angle * math.Pi / 180.0)
	u4 := math.Cos(angle * math.Pi / 180.0)
	return this.Add_cm(u1, u2, u3, u4, 0, 0)
}

// Set the line width.
func (this *ContentCreator) Add_w(lineWidth float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "w"
	op.Params = makeParamsFromFloats([]float64{lineWidth})
	this.operands = append(this.operands, &op)
	return this
}

// Set the line cap style.
func (this *ContentCreator) Add_J(lineCapStyle string) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "J"
	op.Params = makeParamsFromNames([]PdfObjectName{PdfObjectName(lineCapStyle)})
	this.operands = append(this.operands, &op)
	return this
}

// Set the line join style.
func (this *ContentCreator) Add_j(lineJoinStyle string) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "j"
	op.Params = makeParamsFromNames([]PdfObjectName{PdfObjectName(lineJoinStyle)})
	this.operands = append(this.operands, &op)
	return this
}

// Set the miter limit.
func (this *ContentCreator) Add_M(miterlimit float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "M"
	op.Params = makeParamsFromFloats([]float64{miterlimit})
	this.operands = append(this.operands, &op)
	return this
}

// Set the line dash pattern.
func (this *ContentCreator) Add_d(dashArray []int64, dashPhase int64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "d"

	op.Params = []PdfObject{}
	op.Params = append(op.Params, MakeArrayFromIntegers64(dashArray))
	op.Params = append(op.Params, MakeInteger(dashPhase))
	this.operands = append(this.operands, &op)
	return this
}

// Set the color rendering intent.
func (this *ContentCreator) Add_ri(intent PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "ri"
	op.Params = makeParamsFromNames([]PdfObjectName{intent})
	this.operands = append(this.operands, &op)
	return this
}

// Set the flatness tolerance.
func (this *ContentCreator) Add_i(flatness float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "i"
	op.Params = makeParamsFromFloats([]float64{flatness})
	this.operands = append(this.operands, &op)
	return this
}

// Set the graphics state.
func (this *ContentCreator) Add_gs(dictName PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "gs"
	op.Params = makeParamsFromNames([]PdfObjectName{dictName})
	this.operands = append(this.operands, &op)
	return this
}

/* Path construction operators. */

// m: Move the current point to (x,y).
func (this *ContentCreator) Add_m(x, y float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "m"
	op.Params = makeParamsFromFloats([]float64{x, y})
	this.operands = append(this.operands, &op)
	return this
}

// l: Append a straight line segment from the current point to (x,y).
func (this *ContentCreator) Add_l(x, y float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "l"
	op.Params = makeParamsFromFloats([]float64{x, y})
	this.operands = append(this.operands, &op)
	return this
}

// c: Append a Bezier curve to the current path from the current point to (x3,y3) with (x1,x1) and (x2,y2) as control
// points.
func (this *ContentCreator) Add_c(x1, y1, x2, y2, x3, y3 float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "c"
	op.Params = makeParamsFromFloats([]float64{x1, y1, x2, y2, x3, y3})
	this.operands = append(this.operands, &op)
	return this
}

// v: Append a Bezier curve to the current path from the current point to (x3,y3) with the current point and (x2,y2) as
// control points.
func (this *ContentCreator) Add_v(x2, y2, x3, y3 float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "v"
	op.Params = makeParamsFromFloats([]float64{x2, y2, x3, y3})
	this.operands = append(this.operands, &op)
	return this
}

// y: Append a Bezier curve to the current path from the current point to (x3,y3) with (x1, y1) and (x3,y3) as
// control points.
func (this *ContentCreator) Add_y(x1, y1, x3, y3 float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "y"
	op.Params = makeParamsFromFloats([]float64{x1, y1, x3, y3})
	this.operands = append(this.operands, &op)
	return this
}

// h: Close the current subpath by adding a line between the current position and the starting position.
func (this *ContentCreator) Add_h() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "h"
	this.operands = append(this.operands, &op)
	return this
}

// re: Append a rectangle to the current path as a complete subpath, with lower left corner (x,y).
func (this *ContentCreator) Add_re(x, y, width, height float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "re"
	op.Params = makeParamsFromFloats([]float64{x, y, width, height})
	this.operands = append(this.operands, &op)
	return this
}

/* Path painting operators. */

// S: stroke the path.
func (this *ContentCreator) Add_S() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "S"
	this.operands = append(this.operands, &op)
	return this
}

// s: Close and stroke the path.
func (this *ContentCreator) Add_s() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "s"
	this.operands = append(this.operands, &op)
	return this
}

// f: Fill the path using the nonzero winding number rule to determine fill region.
func (this *ContentCreator) Add_f() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "f"
	this.operands = append(this.operands, &op)
	return this
}

// f*: Fill the path using the even-odd rule to determine fill region.
func (this *ContentCreator) Add_f_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "f*"
	this.operands = append(this.operands, &op)
	return this
}

// B: Fill and then stroke the path (nonzero winding number rule).
func (this *ContentCreator) Add_B() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "B"
	this.operands = append(this.operands, &op)
	return this
}

// B*: Fill and then stroke the path (even-odd rule).
func (this *ContentCreator) Add_B_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "B*"
	this.operands = append(this.operands, &op)
	return this
}

// b: Close, fill and then stroke the path (nonzero winding number rule).
func (this *ContentCreator) Add_b() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "b"
	this.operands = append(this.operands, &op)
	return this
}

// b*: Close, fill and then stroke the path (even-odd winding number rule).
func (this *ContentCreator) Add_b_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "b*"
	this.operands = append(this.operands, &op)
	return this
}

// n: End the path without filling or stroking.
func (this *ContentCreator) Add_n() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "n"
	this.operands = append(this.operands, &op)
	return this
}

/* Clipping path operators. */

// W: Modify the current clipping path by intersecting with the current path (nonzero winding rule).
func (this *ContentCreator) Add_W() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "W"
	this.operands = append(this.operands, &op)
	return this
}

// W*: Modify the current clipping path by intersecting with the current path (even odd rule).
func (this *ContentCreator) Add_W_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "W*"
	this.operands = append(this.operands, &op)
	return this
}

/* Color operators. */

// CS: Set the current colorspace for stroking operations.
func (this *ContentCreator) Add_CS(name PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "CS"
	op.Params = makeParamsFromNames([]PdfObjectName{name})
	this.operands = append(this.operands, &op)
	return this
}

// cs: Same as CS but for non-stroking operations.
func (this *ContentCreator) Add_cs(name PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "cs"
	op.Params = makeParamsFromNames([]PdfObjectName{name})
	this.operands = append(this.operands, &op)
	return this
}

// SC: Set color for stroking operations.  Input: c1, ..., cn.
func (this *ContentCreator) Add_SC(c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "SC"
	op.Params = makeParamsFromFloats(c)
	this.operands = append(this.operands, &op)
	return this
}

// SCN: Same as SC but supports more colorspaces.
func (this *ContentCreator) Add_SCN(c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "SCN"
	op.Params = makeParamsFromFloats(c)
	this.operands = append(this.operands, &op)
	return this
}

// SCN with name attribute (for pattern). Syntax: c1 ... cn name SCN.
func (this *ContentCreator) Add_SCN_pattern(name PdfObjectName, c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "SCN"
	op.Params = makeParamsFromFloats(c)
	op.Params = append(op.Params, MakeName(string(name)))
	this.operands = append(this.operands, &op)
	return this
}

// scn: Same as SC but for nonstroking operations.
func (this *ContentCreator) Add_scn(c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "scn"
	op.Params = makeParamsFromFloats(c)
	this.operands = append(this.operands, &op)
	return this
}

// scn with name attribute (for pattern). Syntax: c1 ... cn name scn.
func (this *ContentCreator) Add_scn_pattern(name PdfObjectName, c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "scn"
	op.Params = makeParamsFromFloats(c)
	op.Params = append(op.Params, MakeName(string(name)))
	this.operands = append(this.operands, &op)
	return this
}

// G: Set the stroking colorspace to DeviceGray and sets the gray level (0-1).
func (this *ContentCreator) Add_G(gray float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "G"
	op.Params = makeParamsFromFloats([]float64{gray})
	this.operands = append(this.operands, &op)
	return this
}

// g: Same as G but used for nonstroking operations.
func (this *ContentCreator) Add_g(gray float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "g"
	op.Params = makeParamsFromFloats([]float64{gray})
	this.operands = append(this.operands, &op)
	return this
}

// RG: Set the stroking colorspace to DeviceRGB and sets the r,g,b colors (0-1 each).
func (this *ContentCreator) Add_RG(r, g, b float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "RG"
	op.Params = makeParamsFromFloats([]float64{r, g, b})
	this.operands = append(this.operands, &op)
	return this
}

// rg: Same as RG but used for nonstroking operations.
func (this *ContentCreator) Add_rg(r, g, b float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "rg"
	op.Params = makeParamsFromFloats([]float64{r, g, b})
	this.operands = append(this.operands, &op)
	return this
}

// K: Set the stroking colorspace to DeviceCMYK and sets the c,m,y,k color (0-1 each component).
func (this *ContentCreator) Add_K(c, m, y, k float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "K"
	op.Params = makeParamsFromFloats([]float64{c, m, y, k})
	this.operands = append(this.operands, &op)
	return this
}

// k: Same as K but used for nonstroking operations.
func (this *ContentCreator) Add_k(c, m, y, k float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "k"
	op.Params = makeParamsFromFloats([]float64{c, m, y, k})
	this.operands = append(this.operands, &op)
	return this
}

/* Shading operators. */

func (this *ContentCreator) Add_sh(name PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "sh"
	op.Params = makeParamsFromNames([]PdfObjectName{name})
	this.operands = append(this.operands, &op)
	return this
}

/* Text related operators */

/* Text state operators */

// BT: Begin text.
func (this *ContentCreator) Add_BT() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "BT"
	this.operands = append(this.operands, &op)
	return this
}

// ET: End text.
func (this *ContentCreator) Add_ET() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "ET"
	this.operands = append(this.operands, &op)
	return this
}

// Tc: Set character spacing.
func (this *ContentCreator) Add_Tc(charSpace float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tc"
	op.Params = makeParamsFromFloats([]float64{charSpace})
	this.operands = append(this.operands, &op)
	return this
}

// Tw: Set word spacing.
func (this *ContentCreator) Add_Tw(wordSpace float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tw"
	op.Params = makeParamsFromFloats([]float64{wordSpace})
	this.operands = append(this.operands, &op)
	return this
}

// Tz: Set horizontal scaling.
func (this *ContentCreator) Add_Tz(scale float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tz"
	op.Params = makeParamsFromFloats([]float64{scale})
	this.operands = append(this.operands, &op)
	return this
}

// TL: Set leading.
func (this *ContentCreator) Add_TL(leading float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "TL"
	op.Params = makeParamsFromFloats([]float64{leading})
	this.operands = append(this.operands, &op)
	return this
}

// Tf: Set font and font size.
func (this *ContentCreator) Add_Tf(fontName PdfObjectName, fontSize float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tf"
	op.Params = makeParamsFromNames([]PdfObjectName{fontName})
	op.Params = append(op.Params, makeParamsFromFloats([]float64{fontSize})...)
	this.operands = append(this.operands, &op)
	return this
}

// Tr: Set text rendering mode.
func (this *ContentCreator) Add_Tr(render int64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tr"
	op.Params = makeParamsFromInts([]int64{render})
	this.operands = append(this.operands, &op)
	return this
}

// Ts: Set text rise.
func (this *ContentCreator) Add_Ts(rise float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Ts"
	op.Params = makeParamsFromFloats([]float64{rise})
	this.operands = append(this.operands, &op)
	return this
}

/* Text positioning operators. */

// Td: Move to start of next line with offset (tx, ty).
func (this *ContentCreator) Add_Td(tx, ty float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Td"
	op.Params = makeParamsFromFloats([]float64{tx, ty})
	this.operands = append(this.operands, &op)
	return this
}

// TD: Move to start of next line with offset (tx, ty).
func (this *ContentCreator) Add_TD(tx, ty float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "TD"
	op.Params = makeParamsFromFloats([]float64{tx, ty})
	this.operands = append(this.operands, &op)
	return this
}

// Tm: Set the text line matrix.
func (this *ContentCreator) Add_Tm(a, b, c, d, e, f float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tm"
	op.Params = makeParamsFromFloats([]float64{a, b, c, d, e, f})
	this.operands = append(this.operands, &op)
	return this
}

// T*: Move to the start of next line.
func (this *ContentCreator) Add_Tstar() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "T*"
	this.operands = append(this.operands, &op)
	return this
}

/* Text showing operators */

// Tj: Show a text string.
func (this *ContentCreator) Add_Tj(textstr PdfObjectString) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tj"
	op.Params = makeParamsFromStrings([]PdfObjectString{textstr})
	this.operands = append(this.operands, &op)
	return this
}

// ': Move to next line and show a string.
func (this *ContentCreator) Add_quote(textstr PdfObjectString) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "'"
	op.Params = makeParamsFromStrings([]PdfObjectString{textstr})
	this.operands = append(this.operands, &op)
	return this
}

// '': Move to next line and show a string, using aw and ac as word and character spacing respectively.
func (this *ContentCreator) Add_quotes(textstr PdfObjectString, aw, ac float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "''"
	op.Params = makeParamsFromFloats([]float64{aw, ac})
	op.Params = append(op.Params, makeParamsFromStrings([]PdfObjectString{textstr})...)
	this.operands = append(this.operands, &op)
	return this
}

// TJ. Show one or more text string.  Array of numbers (displacement) and strings.
func (this *ContentCreator) Add_TJ(vals ...PdfObject) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "TJ"
	op.Params = []PdfObject{MakeArray(vals...)}
	this.operands = append(this.operands, &op)
	return this
}
