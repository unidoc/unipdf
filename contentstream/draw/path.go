/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

// Path consists of straight line connections between each point defined in an array of points.
type Path struct {
	Points []Point
}

// NewPath returns a new empty path.
func NewPath() Path {
	return Path{}
}

// AppendPoint adds the specified point to the path.
func (p Path) AppendPoint(point Point) Path {
	p.Points = append(p.Points, point)
	return p
}

// RemovePoint removes the point at the index specified by number from the
// path. The index is 1-based.
func (p Path) RemovePoint(number int) Path {
	if number < 1 || number > len(p.Points) {
		return p
	}

	idx := number - 1
	p.Points = append(p.Points[:idx], p.Points[idx+1:]...)
	return p
}

// Length returns the number of points in the path.
func (p Path) Length() int {
	return len(p.Points)
}

// GetPointNumber returns the path point at the index specified by number.
// The index is 1-based.
func (p Path) GetPointNumber(number int) Point {
	if number < 1 || number > len(p.Points) {
		return Point{}
	}
	return p.Points[number-1]
}

// Copy returns a clone of the path.
func (p Path) Copy() Path {
	pathcopy := Path{}
	pathcopy.Points = []Point{}
	for _, p := range p.Points {
		pathcopy.Points = append(pathcopy.Points, p)
	}
	return pathcopy
}

// Offset shifts the path with the specified offsets.
func (p Path) Offset(offX, offY float64) Path {
	for i, pt := range p.Points {
		p.Points[i] = pt.Add(offX, offY)
	}
	return p
}

// GetBoundingBox returns the bounding box of the path.
func (p Path) GetBoundingBox() BoundingBox {
	bbox := BoundingBox{}

	minX := 0.0
	maxX := 0.0
	minY := 0.0
	maxY := 0.0
	for idx, p := range p.Points {
		if idx == 0 {
			minX = p.X
			maxX = p.X
			minY = p.Y
			maxY = p.Y
			continue
		}

		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	bbox.X = minX
	bbox.Y = minY
	bbox.Width = maxX - minX
	bbox.Height = maxY - minY
	return bbox
}

// BoundingBox represents the smallest rectangular area that encapsulates an object.
type BoundingBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}
