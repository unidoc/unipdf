/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"math"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// ContentCreator is a builder for PDF content streams.
type ContentCreator struct {
	operands ContentStreamOperations
}

// NewContentCreator returns a new initialized ContentCreator.
func NewContentCreator() *ContentCreator {
	creator := &ContentCreator{}
	creator.operands = ContentStreamOperations{}
	return creator
}

// Operations returns the list of operations.
func (cc *ContentCreator) Operations() *ContentStreamOperations {
	return &cc.operands
}

// Bytes converts the content stream operations to a content stream byte presentation, i.e. the kind that can be
// stored as a PDF stream or string format.
func (cc *ContentCreator) Bytes() []byte {
	return cc.operands.Bytes()
}

// String is same as Bytes() except returns as a string for convenience.
func (cc *ContentCreator) String() string {
	return string(cc.operands.Bytes())
}

// Wrap ensures that the contentstream is wrapped within a balanced q ... Q expression.
func (cc *ContentCreator) Wrap() {
	cc.operands.WrapIfNeeded()
}

// AddOperand adds a specified operand.
func (cc *ContentCreator) AddOperand(op ContentStreamOperation) *ContentCreator {
	cc.operands = append(cc.operands, &op)
	return cc
}

// Graphics state operators.

// Add_q adds 'q' operand to the content stream: Pushes the current graphics state on the stack.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_q() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "q"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Q adds 'Q' operand to the content stream: Pops the most recently stored state from the stack.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_Q() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Q"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_cm adds 'cm' operation to the content stream: Modifies the current transformation matrix (ctm)
// of the graphics state.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_cm(a, b, c, d, e, f float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "cm"
	op.Params = makeParamsFromFloats([]float64{a, b, c, d, e, f})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Translate applies a simple x-y translation to the transformation matrix.
func (cc *ContentCreator) Translate(tx, ty float64) *ContentCreator {
	return cc.Add_cm(1, 0, 0, 1, tx, ty)
}

// Scale applies x-y scaling to the transformation matrix.
func (cc *ContentCreator) Scale(sx, sy float64) *ContentCreator {
	return cc.Add_cm(sx, 0, 0, sy, 0, 0)
}

// RotateDeg applies a rotation to the transformation matrix.
func (cc *ContentCreator) RotateDeg(angle float64) *ContentCreator {
	u1 := math.Cos(angle * math.Pi / 180.0)
	u2 := math.Sin(angle * math.Pi / 180.0)
	u3 := -math.Sin(angle * math.Pi / 180.0)
	u4 := math.Cos(angle * math.Pi / 180.0)
	return cc.Add_cm(u1, u2, u3, u4, 0, 0)
}

// Add_w adds 'w' operand to the content stream, which sets the line width.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_w(lineWidth float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "w"
	op.Params = makeParamsFromFloats([]float64{lineWidth})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_J adds 'J' operand to the content stream: Set the line cap style (graphics state).
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_J(lineCapStyle string) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "J"
	op.Params = makeParamsFromNames([]core.PdfObjectName{core.PdfObjectName(lineCapStyle)})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_j adds 'j' operand to the content stream: Set the line join style (graphics state).
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_j(lineJoinStyle string) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "j"
	op.Params = makeParamsFromNames([]core.PdfObjectName{core.PdfObjectName(lineJoinStyle)})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_M adds 'M' operand to the content stream: Set the miter limit (graphics state).
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_M(miterlimit float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "M"
	op.Params = makeParamsFromFloats([]float64{miterlimit})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_d adds 'd' operand to the content stream: Set the line dash pattern.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_d(dashArray []int64, dashPhase int64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "d"

	op.Params = []core.PdfObject{}
	op.Params = append(op.Params, core.MakeArrayFromIntegers64(dashArray))
	op.Params = append(op.Params, core.MakeInteger(dashPhase))
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_ri adds 'ri' operand to the content stream, which sets the color rendering intent.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_ri(intent core.PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "ri"
	op.Params = makeParamsFromNames([]core.PdfObjectName{intent})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_i adds 'i' operand to the content stream: Set the flatness tolerance in the graphics state.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_i(flatness float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "i"
	op.Params = makeParamsFromFloats([]float64{flatness})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_gs adds 'gs' operand to the content stream: Set the graphics state.
//
// See section 8.4.4 "Graphic State Operators" and Table 57 (pp. 135-136 PDF32000_2008).
func (cc *ContentCreator) Add_gs(dictName core.PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "gs"
	op.Params = makeParamsFromNames([]core.PdfObjectName{dictName})
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Path construction operators (8.5.2) */

// Add_m adds 'm' operand to the content stream: Move the current point to (x,y).
//
// See section 8.5.2 "Path Construction Operators" and Table 59 (pp. 140-141 PDF32000_2008).
func (cc *ContentCreator) Add_m(x, y float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "m"
	op.Params = makeParamsFromFloats([]float64{x, y})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_l adds 'l' operand to the content stream:
// Append a straight line segment from the current point to (x,y).
//
// See section 8.5.2 "Path Construction Operators" and Table 59 (pp. 140-141 PDF32000_2008).
func (cc *ContentCreator) Add_l(x, y float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "l"
	op.Params = makeParamsFromFloats([]float64{x, y})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_c adds 'c' operand to the content stream: Append a Bezier curve to the current path from
// the current point to (x3,y3) with (x1,x1) and (x2,y2) as control points.
//
// See section 8.5.2 "Path Construction Operators" and Table 59 (pp. 140-141 PDF32000_2008).
func (cc *ContentCreator) Add_c(x1, y1, x2, y2, x3, y3 float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "c"
	op.Params = makeParamsFromFloats([]float64{x1, y1, x2, y2, x3, y3})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_v appends 'v' operand to the content stream: Append a Bezier curve to the current path from the
// current point to (x3,y3) with the current point and (x2,y2) as control points.
//
// See section 8.5.2 "Path Construction Operators" and Table 59 (pp. 140-141 PDF32000_2008).
func (cc *ContentCreator) Add_v(x2, y2, x3, y3 float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "v"
	op.Params = makeParamsFromFloats([]float64{x2, y2, x3, y3})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_y appends 'y' operand to the content stream: Append a Bezier curve to the current path from the
// current point to (x3,y3) with (x1, y1) and (x3,y3) as control points.
//
// See section 8.5.2 "Path Construction Operators" and Table 59 (pp. 140-141 PDF32000_2008).
func (cc *ContentCreator) Add_y(x1, y1, x3, y3 float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "y"
	op.Params = makeParamsFromFloats([]float64{x1, y1, x3, y3})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_h appends 'h' operand to the content stream:
// Close the current subpath by adding a line between the current position and the starting position.
//
// See section 8.5.2 "Path Construction Operators" and Table 59 (pp. 140-141 PDF32000_2008).
func (cc *ContentCreator) Add_h() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "h"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_re appends 're' operand to the content stream:
// Append a rectangle to the current path as a complete subpath, with lower left corner (x,y).
//
// See section 8.5.2 "Path Construction Operators" and Table 59 (pp. 140-141 PDF32000_2008).
func (cc *ContentCreator) Add_re(x, y, width, height float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "re"
	op.Params = makeParamsFromFloats([]float64{x, y, width, height})
	cc.operands = append(cc.operands, &op)
	return cc
}

/* XObject operators. */

// Add_Do adds 'Do' operation to the content stream:
// Displays an XObject (image or form) specified by `name`.
//
// See section 8.8 "External Objects" and Table 87 (pp. 209-220 PDF32000_2008).
func (cc *ContentCreator) Add_Do(name core.PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Do"
	op.Params = makeParamsFromNames([]core.PdfObjectName{name})
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Path painting operators (8.5.3 p. 142 PDF32000_2008). */

// Add_S appends 'S' operand to the content stream: Stroke the path.
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_S() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "S"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_s appends 's' operand to the content stream: Close and stroke the path.
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_s() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "s"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_f appends 'f' operand to the content stream:
// Fill the path using the nonzero winding number rule to determine fill region.
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_f() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "f"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_f_starred appends 'f*' operand to the content stream.
// f*: Fill the path using the even-odd rule to determine fill region.
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_f_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "f*"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_B appends 'B' operand to the content stream:
// Fill and then stroke the path (nonzero winding number rule).
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_B() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "B"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_B_starred appends 'B*' operand to the content stream:
// Fill and then stroke the path (even-odd rule).
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_B_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "B*"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_b appends 'b' operand to the content stream:
// Close, fill and then stroke the path (nonzero winding number rule).
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_b() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "b"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_b_starred appends 'b*' operand to the content stream:
// Close, fill and then stroke the path (even-odd winding number rule).
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_b_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "b*"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_n appends 'n' operand to the content stream:
// End the path without filling or stroking.
//
// See section 8.5.3 "Path Painting Operators" and Table 60 (p. 143 PDF32000_2008).
func (cc *ContentCreator) Add_n() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "n"
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Clipping path operators (8.5.4 p. 145 PDF32000_2008). */

// Add_W appends 'W' operand to the content stream:
// Modify the current clipping path by intersecting with the current path (nonzero winding rule).
//
// See section 8.5.4 "Clipping Path Operators" and Table 61 (p. 146 PDF32000_2008).
func (cc *ContentCreator) Add_W() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "W"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_W_starred appends 'W*' operand to the content stream:
// Modify the current clipping path by intersecting with the current path (even odd rule).
//
// See section 8.5.4 "Clipping Path Operators" and Table 61 (p. 146 PDF32000_2008).
func (cc *ContentCreator) Add_W_starred() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "W*"
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Color operators (8.6.8 p. 179 PDF32000_2008). */

// Add_CS appends 'CS' operand to the content stream:
// Set the current colorspace for stroking operations.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_CS(name core.PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "CS"
	op.Params = makeParamsFromNames([]core.PdfObjectName{name})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_cs appends 'cs' operand to the content stream:
// Same as CS but for non-stroking operations.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_cs(name core.PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "cs"
	op.Params = makeParamsFromNames([]core.PdfObjectName{name})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_SC appends 'SC' operand to the content stream:
// Set color for stroking operations.  Input: c1, ..., cn.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_SC(c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "SC"
	op.Params = makeParamsFromFloats(c)
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_SCN appends 'SCN' operand to the content stream:
// Same as SC but supports more colorspaces.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_SCN(c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "SCN"
	op.Params = makeParamsFromFloats(c)
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_SCN_pattern appends 'SCN' operand to the content stream for pattern `name`:
// SCN with name attribute (for pattern). Syntax: c1 ... cn name SCN.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_SCN_pattern(name core.PdfObjectName, c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "SCN"
	op.Params = makeParamsFromFloats(c)
	op.Params = append(op.Params, core.MakeName(string(name)))
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_scn appends 'scn' operand to the content stream:
// Same as SC but for nonstroking operations.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_scn(c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "scn"
	op.Params = makeParamsFromFloats(c)
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_scn_pattern appends 'scn' operand to the content stream for pattern `name`:
// scn with name attribute (for pattern). Syntax: c1 ... cn name scn.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_scn_pattern(name core.PdfObjectName, c ...float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "scn"
	op.Params = makeParamsFromFloats(c)
	op.Params = append(op.Params, core.MakeName(string(name)))
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_G appends 'G' operand to the content stream:
// Set the stroking colorspace to DeviceGray and sets the gray level (0-1).
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_G(gray float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "G"
	op.Params = makeParamsFromFloats([]float64{gray})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_g appends 'g' operand to the content stream:
// Same as G but used for nonstroking operations.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_g(gray float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "g"
	op.Params = makeParamsFromFloats([]float64{gray})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_RG appends 'RG' operand to the content stream:
// Set the stroking colorspace to DeviceRGB and sets the r,g,b colors (0-1 each).
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_RG(r, g, b float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "RG"
	op.Params = makeParamsFromFloats([]float64{r, g, b})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_rg appends 'rg' operand to the content stream:
// Same as RG but used for nonstroking operations.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_rg(r, g, b float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "rg"
	op.Params = makeParamsFromFloats([]float64{r, g, b})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_K appends 'K' operand to the content stream:
// Set the stroking colorspace to DeviceCMYK and sets the c,m,y,k color (0-1 each component).
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_K(c, m, y, k float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "K"
	op.Params = makeParamsFromFloats([]float64{c, m, y, k})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_k appends 'k' operand to the content stream:
// Same as K but used for nonstroking operations.
//
// See section 8.6.8 "Colour Operators" and Table 74 (p. 179-180 PDF32000_2008).
func (cc *ContentCreator) Add_k(c, m, y, k float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "k"
	op.Params = makeParamsFromFloats([]float64{c, m, y, k})
	cc.operands = append(cc.operands, &op)
	return cc
}

// SetStrokingColor sets the stroking `color` where color can be one of
// PdfColorDeviceGray, PdfColorDeviceRGB, or PdfColorDeviceCMYK.
func (cc *ContentCreator) SetStrokingColor(color model.PdfColor) *ContentCreator {
	switch t := color.(type) {
	case *model.PdfColorDeviceGray:
		cc.Add_G(t.Val())
	case *model.PdfColorDeviceRGB:
		cc.Add_RG(t.R(), t.G(), t.B())
	case *model.PdfColorDeviceCMYK:
		cc.Add_K(t.C(), t.M(), t.Y(), t.K())
	default:
		common.Log.Debug("SetStrokingColor: unsupported color: %T", t)
	}
	return cc
}

// SetNonStrokingColor sets the non-stroking `color` where color can be one of
// PdfColorDeviceGray, PdfColorDeviceRGB, or PdfColorDeviceCMYK.
func (cc *ContentCreator) SetNonStrokingColor(color model.PdfColor) *ContentCreator {
	switch t := color.(type) {
	case *model.PdfColorDeviceGray:
		cc.Add_g(t.Val())
	case *model.PdfColorDeviceRGB:
		cc.Add_rg(t.R(), t.G(), t.B())
	case *model.PdfColorDeviceCMYK:
		cc.Add_k(t.C(), t.M(), t.Y(), t.K())
	default:
		common.Log.Debug("SetNonStrokingColor: unsupported color: %T", t)
	}
	return cc
}

/* Shading operator (8.7.4.2 p. 189 PDF32000_2008). */

// Add_sh appends 'sh' operand to the content stream:
// Paints the shape and colour shading described by a shading dictionary specified by `name`,
// subject to the current clipping path
//
// See section 8.7.4 "Shading Patterns" and Table 77 (p. 190 PDF32000_2008).
func (cc *ContentCreator) Add_sh(name core.PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "sh"
	op.Params = makeParamsFromNames([]core.PdfObjectName{name})
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Text related operators. */

/* Text object operators (9.4 p. 256 PDF32000_2008). */

// Add_BT appends 'BT' operand to the content stream:
// Begin text.
//
// See section 9.4 "Text Objects" and Table 107 (p. 256 PDF32000_2008).
func (cc *ContentCreator) Add_BT() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "BT"
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_ET appends 'ET' operand to the content stream:
// End text.
//
// See section 9.4 "Text Objects" and Table 107 (p. 256 PDF32000_2008).
func (cc *ContentCreator) Add_ET() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "ET"
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Text state operators (9.3 p. 251 PDF32000_2008). */

// Add_Tc appends 'Tc' operand to the content stream:
// Set character spacing.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (cc *ContentCreator) Add_Tc(charSpace float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tc"
	op.Params = makeParamsFromFloats([]float64{charSpace})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Tw appends 'Tw' operand to the content stream:
// Set word spacing.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (cc *ContentCreator) Add_Tw(wordSpace float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tw"
	op.Params = makeParamsFromFloats([]float64{wordSpace})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Tz appends 'Tz' operand to the content stream:
// Set horizontal scaling.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (cc *ContentCreator) Add_Tz(scale float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tz"
	op.Params = makeParamsFromFloats([]float64{scale})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_TL appends 'TL' operand to the content stream:
// Set leading.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (cc *ContentCreator) Add_TL(leading float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "TL"
	op.Params = makeParamsFromFloats([]float64{leading})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Tf appends 'Tf' operand to the content stream:
// Set font and font size specified by font resource `fontName` and `fontSize`.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (cc *ContentCreator) Add_Tf(fontName core.PdfObjectName, fontSize float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tf"
	op.Params = makeParamsFromNames([]core.PdfObjectName{fontName})
	op.Params = append(op.Params, makeParamsFromFloats([]float64{fontSize})...)
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Tr appends 'Tr' operand to the content stream:
// Set text rendering mode.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (cc *ContentCreator) Add_Tr(render int64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tr"
	op.Params = makeParamsFromInts([]int64{render})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Ts appends 'Ts' operand to the content stream:
// Set text rise.
//
// See section 9.3 "Text State Parameters and Operators" and
// Table 105 (pp. 251-252 PDF32000_2008).
func (cc *ContentCreator) Add_Ts(rise float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Ts"
	op.Params = makeParamsFromFloats([]float64{rise})
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Text positioning operators (9.4.2 p. 257 PDF32000_2008). */

// Add_Td appends 'Td' operand to the content stream:
// Move to start of next line with offset (`tx`, `ty`).
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (cc *ContentCreator) Add_Td(tx, ty float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Td"
	op.Params = makeParamsFromFloats([]float64{tx, ty})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_TD appends 'TD' operand to the content stream:
// Move to start of next line with offset (`tx`, `ty`).
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (cc *ContentCreator) Add_TD(tx, ty float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "TD"
	op.Params = makeParamsFromFloats([]float64{tx, ty})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Tm appends 'Tm' operand to the content stream:
// Set the text line matrix.
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (cc *ContentCreator) Add_Tm(a, b, c, d, e, f float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tm"
	op.Params = makeParamsFromFloats([]float64{a, b, c, d, e, f})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_Tstar appends 'T*' operand to the content stream:
// Move to the start of next line.
//
// See section 9.4.2 "Text Positioning Operators" and
// Table 108 (pp. 257-258 PDF32000_2008).
func (cc *ContentCreator) Add_Tstar() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "T*"
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Text showing operators (9.4.3 p. 258 PDF32000_2008). */

// Add_Tj appends 'Tj' operand to the content stream:
// Show a text string.
//
// See section 9.4.3 "Text Showing Operators" and
// Table 209 (pp. 258-259 PDF32000_2008).
func (cc *ContentCreator) Add_Tj(textstr core.PdfObjectString) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "Tj"
	op.Params = makeParamsFromStrings([]core.PdfObjectString{textstr})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_quote appends "'" operand to the content stream:
// Move to next line and show a string.
//
// See section 9.4.3 "Text Showing Operators" and
// Table 209 (pp. 258-259 PDF32000_2008).
func (cc *ContentCreator) Add_quote(textstr core.PdfObjectString) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "'"
	op.Params = makeParamsFromStrings([]core.PdfObjectString{textstr})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_quotes appends `"` operand to the content stream:
// Move to next line and show a string, using `aw` and `ac` as word
// and character spacing respectively.
//
// See section 9.4.3 "Text Showing Operators" and
// Table 209 (pp. 258-259 PDF32000_2008).
func (cc *ContentCreator) Add_quotes(textstr core.PdfObjectString, aw, ac float64) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = `"`
	op.Params = makeParamsFromFloats([]float64{aw, ac})
	op.Params = append(op.Params, makeParamsFromStrings([]core.PdfObjectString{textstr})...)
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_TJ appends 'TJ' operand to the content stream:
// Show one or more text string. Array of numbers (displacement) and strings.
//
// See section 9.4.3 "Text Showing Operators" and
// Table 209 (pp. 258-259 PDF32000_2008).
func (cc *ContentCreator) Add_TJ(vals ...core.PdfObject) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "TJ"
	op.Params = []core.PdfObject{core.MakeArray(vals...)}
	cc.operands = append(cc.operands, &op)
	return cc
}

/* Marked content operators (14.6 p. 560 PDF32000_2008). */

// Add_BMC appends 'BMC' operand to the content stream:
// Begins a marked-content sequence terminated by a balancing EMC operator.
// `tag` shall be a name object indicating the role or significance of
// the sequence.
//
// See section 14.6 "Marked Content" and Table 320 (p. 561 PDF32000_2008).
func (cc *ContentCreator) Add_BMC(tag core.PdfObjectName) *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "BMC"
	op.Params = makeParamsFromNames([]core.PdfObjectName{tag})
	cc.operands = append(cc.operands, &op)
	return cc
}

// Add_EMC appends 'EMC' operand to the content stream:
// Ends a marked-content sequence.
//
// See section 14.6 "Marked Content" and Table 320 (p. 561 PDF32000_2008).
func (cc *ContentCreator) Add_EMC() *ContentCreator {
	op := ContentStreamOperation{}
	op.Operand = "EMC"
	cc.operands = append(cc.operands, &op)
	return cc
}
