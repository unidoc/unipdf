/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 *
 * Based on pdf/contentstream/draw/path.go
 */

package extractor

import (
	"fmt"

	"github.com/unidoc/unidoc/pdf/contentstream"
)

// Path describes the straight line components of a pdf path.
// A path consists of straight line connections between each point defined in an array of points.
type Path struct {
	Points []Point
}

// NewPath returns an empty Path
func NewPath() Path {
	path := Path{}
	path.Points = []Point{}
	return path
}

// AppendPoint appends `point` to `path`
func (path *Path) AppendPoint(point Point) {
	path.Points = append(path.Points, point)
}

// Length returns the number of points in `path`
func (path *Path) Length() int {
	return len(path.Points)
}

// Copy returns a copy of `path`
func (path *Path) Copy() Path {
	pathcopy := NewPath()
	for _, p := range path.Points {
		pathcopy.Points = append(pathcopy.Points, p)
	}
	return pathcopy
}

// Transform transforms all point in `path` by the affine transformation a, b, c, d, tx, ty
func (path *Path) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	path.transformByMatrix(m)
}

// transformByMatrix transforms all point in `path` by the affine transformation `m`
func (path *Path) transformByMatrix(m contentstream.Matrix) {
	for _, p := range path.Points {
		p.transformByMatrix(m)
	}
}

// String returns a string describing `path`
func (path *Path) String() string {
	return fmt.Sprintf("%+v", path.Points)
}

// GetBoundingBox returns `path`'s bounding box
func (path *Path) GetBoundingBox() BoundingBox {
	if len(path.Points) == 0 {
		return BoundingBox{}
	}

	p := path.Points[0]
	minX, maxX := p.X, p.X
	minY, maxY := p.Y, p.Y
	for _, p := range path.Points[1:] {
		if p.X < minX {
			minX = p.X
		} else if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		} else if p.Y > maxY {
			maxY = p.Y
		}
	}

	return BoundingBox{
		Ll: Point{minX, minY},
		Ur: Point{maxX, maxY},
	}
}

// BoundingBox describes a bounding box.
type BoundingBox struct {
	Ll Point // lower,left  i.e. lowest coordinate values
	Ur Point // upper,right i.e. highest coordinate values
}

func (bbox *BoundingBox) String() string {
	return fmt.Sprintf("Lx:%s Ur:%s", bbox.Ll.String(), bbox.Ur.String())
}

// Transform transforms `bbox` by the affine transformation a, b, c, d, tx, ty
func (bbox *BoundingBox) Transform(a, b, c, d, tx, ty float64) {
	m := contentstream.NewMatrix(a, b, c, d, tx, ty)
	bbox.Ll.transformByMatrix(m)
	bbox.Ur.transformByMatrix(m)
}
