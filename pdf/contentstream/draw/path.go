/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

// A path consists of straight line connections between each point defined in an array of points.
type Path struct {
	Points []Point
}

func NewPath() Path {
	path := Path{}
	path.Points = []Point{}
	return path
}

func (this Path) AppendPoint(point Point) Path {
	this.Points = append(this.Points, point)
	return this
}

func (this Path) RemovePoint(number int) Path {
	if number < 1 || number > len(this.Points) {
		return this
	}

	idx := number - 1
	this.Points = append(this.Points[:idx], this.Points[idx+1:]...)
	return this
}

func (this Path) Length() int {
	return len(this.Points)
}

func (this Path) GetPointNumber(number int) Point {
	if number < 1 || number > len(this.Points) {
		return Point{}
	}
	return this.Points[number-1]
}

func (path Path) Copy() Path {
	pathcopy := Path{}
	pathcopy.Points = []Point{}
	for _, p := range path.Points {
		pathcopy.Points = append(pathcopy.Points, p)
	}
	return pathcopy
}

func (path Path) Offset(offX, offY float64) Path {
	for i, p := range path.Points {
		path.Points[i] = p.Add(offX, offY)
	}
	return path
}

func (path Path) GetBoundingBox() BoundingBox {
	bbox := BoundingBox{}

	minX := 0.0
	maxX := 0.0
	minY := 0.0
	maxY := 0.0
	for idx, p := range path.Points {
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

type BoundingBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}
