/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// ExtractShapes returns the marked paths on the pdf page in `e` as a ShapeList.
func (e *Extractor) ExtractShapes() (*ShapeList, error) {
	shapeList := &ShapeList{}
	shape := Shape{}
	cp := Point{}
	inText := false

	cstreamParser := contentstream.NewContentStreamParser(e.contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		return shapeList, err
	}
	processor := contentstream.NewContentStreamProcessor(*operations)

	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState,
			resources *model.PdfPageResources) error {
			operand := op.Operand
			// common.Log.Debug("++Operand: %s", op.String())

			switch operand {
			case "BT":
				inText = true
			case "ET":
				inText = false

			case "m": // move to
				if inText {
					common.Log.Debug("m operand inside text")
					return nil
				}
				if len(op.Params) != 2 {
					return errors.New("m: Invalid number of inputs")
				}

				shapeList.AppendPath(shape)
				shape = NewShape()
				cp = Point{}

				cp, err = toPageCoords(gs, op.Params)
				if err != nil {
					return err
				}
				shape = NewShape()
				shape.AppendPoint(cp)

			case "l": // line to
				if inText {
					common.Log.Debug("l operand inside text")
					return nil
				}
				if len(op.Params) != 2 {
					return errors.New("l: Invalid number of inputs")
				}
				if shape.Empty() {
					common.Log.Debug("l operator with no cp. shape=%+v", shape)
				}
				cp, err = toPageCoords(gs, op.Params)
				if err != nil {
					return err
				}
				shape.AppendPoint(cp)

			case "c": // curve to  cp, p1, p2, p3
				if inText {
					common.Log.Debug("c operand inside text")
					return nil
				}
				if len(op.Params) != 6 {
					return errors.New("c: Invalid number of inputs")
				}
				if shape.Empty() {
					common.Log.Debug("c operator with no cp. shape=%+v", shape)
				}
				points, err := toPagePointList(gs, op.Params)
				if err != nil {
					return err
				}
				p1, p2, p3 := points[0], points[1], points[2]
				if shape.Empty() {
					common.Log.Debug("c operator with no cp. shape=%+v", shape)
					shape.AppendPoint(p3)
				} else {
					shape.AppendCurve(cp, p1, p2, p3)
				}
				cp = p3

			case "v": // curve to  cp, cp, p2, p3
				if inText {
					common.Log.Debug("v operand inside text")
					return nil
				}
				if len(op.Params) != 4 {
					return errors.New("v: Invalid number of inputs")
				}
				if shape.Empty() {
					common.Log.Debug("c operator with no cp. shape=%+v", shape)
				}
				points, err := toPagePointList(gs, op.Params)
				if err != nil {
					return err
				}
				p2, p3 := points[0], points[1]
				if shape.Empty() {
					common.Log.Debug("cv operator with no cp. shape=%+v", shape)
					shape.AppendPoint(p3)
				} else {
					shape.AppendCurve(cp, cp, p2, p3)
				}
				cp = p3

			case "y": // curve to  cp, p1, cp, p3
				if inText {
					common.Log.Debug("yv operand inside text")
					return nil
				}
				if len(op.Params) != 4 {
					return errors.New("v: Invalid number of inputs")
				}
				if shape.Empty() {
					common.Log.Debug("c operator with no cp. shape=%+v", shape)
				}
				points, err := toPagePointList(gs, op.Params)
				if err != nil {
					return err
				}
				p1, p3 := points[0], points[1]
				if shape.Empty() {
					common.Log.Debug("cv operator with no cp. shape=%+v", shape)
					shape.AppendPoint(cp)
				} else {
					shape.AppendCurve(cp, p1, cp, p3)
				}
				cp = p3

			case "re": // rectangle
				if inText {
					common.Log.Debug("re operand inside text")
					return nil
				}
				if len(op.Params) != 4 {
					return errors.New("re: Invalid number of inputs")
				}
				floats, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}
				x, y, w, h := floats[0], floats[1], floats[2], floats[3]
				p0 := toPagePoint(gs, x, y)
				p1 := toPagePoint(gs, x+w, y)
				p2 := toPagePoint(gs, x+w, y+h)
				p3 := toPagePoint(gs, x, y+h)

				shapeList.AppendPath(shape)
				shape = NewShape()
				shape.AppendPoint(p0)
				shape.AppendPoint(p1)
				shape.AppendPoint(p2)
				shape.AppendPoint(p3)
				shape.AppendPoint(p0)
				cp = p0

			case "h": // close path
				if inText {
					common.Log.Debug("h operand inside text")
					return nil
				}
				if !shape.Empty() {
					shape.AppendPoint(shape.Origin())
					shapeList.AppendPath(shape)
					shape = NewShape()
					cp = Point{}
				}

			case "S", "s", "f", "F", "f*", "B", "B*", "b", "b*", "n": // filling, stroking and closing paths
				if inText {
					common.Log.Debug("%s operand inside text", operand)
					return nil
				}

				lastPath := shapeList.LastPath(&shape)
				switch operand {
				case "s", "S":
					lastPath.ColorStroking = gs.ColorStroking
				case "f", "F": // close and fill path with winding rule fill
					lastPath.ColorNonStroking = gs.ColorNonStroking
					lastPath.FillType = FillRuleWinding
				case "f*": // close and fill path with odd-even fill
					lastPath.ColorNonStroking = gs.ColorNonStroking
					lastPath.FillType = FillRuleOddEven
				case "b", "B": // close, stroke and fill path with winding rule fill
					lastPath.ColorStroking = gs.ColorStroking
					lastPath.ColorNonStroking = gs.ColorNonStroking
					lastPath.FillType = FillRuleWinding
				case "b*", "B*": //// close, stroke and fill path with odd-even rule fill
					lastPath.ColorStroking = gs.ColorStroking
					lastPath.ColorNonStroking = gs.ColorNonStroking
					lastPath.FillType = FillRuleOddEven
				}

				switch operand {
				case "s", "f", "F", "b", "b*", "n":
					if !shape.Empty() {
						shape.AppendPoint(shape.Origin())
						shapeList.AppendPath(shape)
						shape = NewShape()
						cp = Point{}
					}
				}
			}
			return nil
		})

	err = processor.Process(e.resources)
	if err != nil {
		common.Log.Error("Error processing: %v", err)
		return nil, err
	}

	return shapeList, nil
}

// ShapeList is a list of pdf paths
type ShapeList struct {
	Shapes []Shape
}

// Shape describes a pdf path
type Shape struct {
	Lines            Path            // Line segmnents
	Curves           CubicBezierPath // Curve segments
	Segments         []PathSegment   // All segments
	ColorStroking    model.PdfColor  // Colour that shape is stroked with, if any
	ColorNonStroking model.PdfColor  // Colour that shape is filled with, if any
	FillType         FillRule        // Filling rule of filled shaped
}

type PathSegment struct {
	Index  int
	Curved bool
}

type FillRule int

const (
	FillRuleWinding FillRule = iota
	FillRuleOddEven
)

func (shape *Shape) String() string {
	return fmt.Sprintf("stroke:%+v fill:%+v lines:%d curves:%d",
		shape.ColorStroking, shape.ColorNonStroking, shape.Lines.Length(), shape.Curves.Length())
}

// NewShape returns an empty Shape
func NewShape() Shape {
	return Shape{}
}

// AppendPoint appends `point` to `shape`
// This can be used to move the current pointer or to add a line segment
// point is assumed to be in page coordinates
func (shape *Shape) AppendPoint(point Point) {
	n := shape.Lines.Length()
	shape.Lines.AppendPoint(point)
	shape.Segments = append(shape.Segments, PathSegment{n, false})
	common.Log.Debug("AppendPath: point=%s shape=%d", point.String(), shape.Length())
}

// AppendCurve appends Bézier curve with control points p0,p1,p2,p3 to `shape`
// This can be used to move the current pointer or to add a line segmebnt
func (shape *Shape) AppendCurve(p0, p1, p2, p3 Point) {
	n := shape.Lines.Length()
	curve := CubicBezierCurve{
		P0: p0,
		P1: p1,
		P2: p2,
		P3: p3,
	}
	shape.Curves.AppendCurve(curve)
	shape.Segments = append(shape.Segments, PathSegment{n, true})
	// common.Log.Debug("AppendCurve: point=%s shape=%s", point.String(), shape.String())
}

// Origin returns the first point in `shape`
// Do NOT call Origin with an empty shape
func (shape *Shape) Origin() Point {
	if shape.Empty() {
		common.Log.Error("Not allowed! Shape.Origin: No points")
		return Point{}
	}
	i := shape.Segments[0].Index
	if shape.Segments[0].Curved {
		return shape.Curves.Curves[i].P0
	}
	return shape.Lines.Points[i]
}

// Length returns the number of segments in `shape`
func (shape *Shape) Length() int {
	numLines := shape.Lines.Length() - 1
	if numLines < 0 {
		numLines = 0
	}
	return numLines + shape.Curves.Length()
}

// Empty returns true if no points or curves have been added to `shape`
func (shape *Shape) Empty() bool {
	return len(shape.Segments) == 0
}

// Copy returns a copy of `shape`
func (shape *Shape) Copy() Shape {
	shape2 := NewShape()
	shape2.Lines = shape.Lines.Copy()
	shape2.Curves = shape.Curves.Copy()
	for _, s := range shape.Segments {
		shape2.Segments = append(shape2.Segments, s)
	}
	return shape2
}

// Transform transforms `shape` by the affine transformation a, b, c, d, tx, ty
func (shape *Shape) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	shape.transformByMatrix(m)
}

// transformByMatrix transforms `shape` by the affine transformation `m`
func (shape *Shape) transformByMatrix(m contentstream.Matrix) {
	shape.Lines.transformByMatrix(m)
	shape.Curves.transformByMatrix(m)
}

// GetBoundingBox returns `shape`s bounding box
func (shape *Shape) GetBoundingBox() BoundingBox {
	bboxL := shape.Lines.GetBoundingBox()
	bboxC := shape.Curves.GetBoundingBox()
	if shape.Lines.Length() == 0 && shape.Curves.Length() == 0 {
		return BoundingBox{}
	} else if shape.Lines.Length() == 0 {
		return bboxC
	} else if shape.Curves.Length() == 0 {
		return bboxL
	}

	return BoundingBox{
		Ll: Point{minFloat(bboxL.Ll.X, bboxC.Ll.X), minFloat(bboxL.Ll.Y, bboxC.Ll.Y)},
		Ur: Point{maxFloat(bboxL.Ur.X, bboxC.Ur.X), maxFloat(bboxL.Ur.Y, bboxC.Ur.Y)},
	}
}

func (sl *ShapeList) Length() int {
	return len(sl.Shapes)
}

func (sl *ShapeList) LastPath(currentPath *Shape) *Shape {
	if len(sl.Shapes) > 0 {
		currentPath = &sl.Shapes[len(sl.Shapes)-1]
	}
	return currentPath
}

// add appends a Shape to the path list
func (sl *ShapeList) AppendPath(s Shape) {
	if s.Length() > 0 {
		sl.Shapes = append(sl.Shapes, s)
	}
}

// Transform transforms all shapes of  `sl` by the affine transformation a, b, c, d, tx, ty
func (sl *ShapeList) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	sl.transformByMatrix(m)
}

// transformByMatrix transforms `shape` by the affine transformation `m`
func (sl *ShapeList) transformByMatrix(m contentstream.Matrix) {
	for _, shape := range sl.Shapes {
		shape.transformByMatrix(m)
	}
}
