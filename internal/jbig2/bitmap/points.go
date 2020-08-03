/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"sort"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Point is the basic structure that contains x, y float32 values.
// In compare with image.Point the x and y are floats not integers.
type Point struct {
	X, Y float32
}

// Points is the slice of the float Points that has panic safe methods for getting and adding new Points.
type Points []Point

// Add adds the points 'pt' to the slice of points.
func (p *Points) Add(pt *Points) error {
	const processName = "Points.Add"
	if p == nil {
		return errors.Error(processName, "points not defined")
	}
	if pt == nil {
		return errors.Error(processName, "argument points not defined")
	}
	*p = append(*p, *pt...)
	return nil
}

// AddPoint adds the Point{'x', 'y'} to the 'p' Points.
func (p *Points) AddPoint(x, y float32) {
	*p = append(*p, Point{x, y})
}

// Get gets the point at 'i' index.
// Returns error if the 'i' index is out of range.
func (p Points) Get(i int) (Point, error) {
	if i > len(p)-1 {
		return Point{}, errors.Errorf("Points.Get", "index: '%d' out of range", i)
	}
	return p[i], nil
}

// GetIntX gets integer value of x coordinate for the point at index 'i'.
func (p Points) GetIntX(i int) (int, error) {
	if i >= len(p) {
		return 0, errors.Errorf("Points.GetIntX", "index: '%d' out of range", i)
	}
	return int(p[i].X), nil
}

// GetIntY gets integer value of y coordinate for the point at index 'i'.
func (p Points) GetIntY(i int) (int, error) {
	if i >= len(p) {
		return 0, errors.Errorf("Points.GetIntY", "index: '%d' out of range", i)
	}
	return int(p[i].Y), nil
}

// GetGeometry gets the geometry 'x' and 'y' of the point at the 'i' index.
// Returns error if the index is out of range.
func (p Points) GetGeometry(i int) (x, y float32, err error) {
	if i > len(p)-1 {
		return 0, 0, errors.Errorf("Points.Get", "index: '%d' out of range", i)
	}
	pt := p[i]
	return pt.X, pt.Y, nil
}

// Size returns the size of the points slice.
func (p Points) Size() int {
	return len(p)
}

// XSorter is the sorter function based on the points 'x' coordinates.
func (p Points) XSorter() func(i, j int) bool {
	return func(i, j int) bool {
		return p[i].X < p[j].X
	}
}

// YSorter is the sorter function based on the points 'y' coordinates.
func (p Points) YSorter() func(i, j int) bool {
	return func(i, j int) bool {
		return p[i].Y < p[j].Y
	}
}

// ClassedPoints is the wrapper that contains both points and integer slice, where
// each index in point relates to the index of int slice.
// Even though the type implements sort.Interface, don't use it directly - as this would panic.
// Use SortByX, or SortByY function.
type ClassedPoints struct {
	*Points
	basic.IntSlice
	currentSortFunction func(i, j int) bool
}

// NewClassedPoints creates and validates PointsBitmaps based on provided 'points'
// and 'classes'. The classes contains id's of the classes related to the index of the 'points'.
func NewClassedPoints(points *Points, classes basic.IntSlice) (*ClassedPoints, error) {
	const processName = "NewClassedPoints"
	if points == nil {
		return nil, errors.Error(processName, "provided nil points")
	}
	if classes == nil {
		return nil, errors.Error(processName, "provided nil classes")
	}
	// create ClassedPoints
	c := &ClassedPoints{Points: points, IntSlice: classes}
	// validate all the input classes.
	if err := c.validateIntSlice(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return c, nil
}

// GetIntXByClass gets the integer of point.Y by the index of the 'IntSlice'.
func (c *ClassedPoints) GetIntXByClass(i int) (int, error) {
	const processName = "ClassedPoints.GetIntYByClass"
	if i >= c.IntSlice.Size() {
		return 0, errors.Errorf(processName, "i: '%d' is out of the range of the IntSlice", i)
	}
	return int(c.XAtIndex(i)), nil
}

// GetIntYByClass gets the integer of point.Y by the index of the 'IntSlice'.
func (c *ClassedPoints) GetIntYByClass(i int) (int, error) {
	const processName = "ClassedPoints.GetIntYByClass"
	if i >= c.IntSlice.Size() {
		return 0, errors.Errorf(processName, "i: '%d' is out of the range of the IntSlice", i)
	}
	return int(c.YAtIndex(i)), nil
}

// GroupByY groups provided intSlice into ClassedPoints based on their 'Y'.
func (c *ClassedPoints) GroupByY() ([]*ClassedPoints, error) {
	const processName = "ClassedPoints.GroupByY"
	if err := c.validateIntSlice(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	if c.IntSlice.Size() == 0 {
		return nil, errors.Error(processName, "No classes provided")
	}
	// sort the classes by Y
	c.SortByY()
	// define the variables
	var (
		res   []*ClassedPoints
		tempY int
	)
	y := -1
	var currentPoints *ClassedPoints
	for i := 0; i < len(c.IntSlice); i++ {
		tempY = int(c.YAtIndex(i))
		if tempY != y {
			currentPoints = &ClassedPoints{Points: c.Points}
			y = tempY
			res = append(res, currentPoints)
		}
		currentPoints.IntSlice = append(currentPoints.IntSlice, c.IntSlice[i])
	}
	// sort each layer of the points by the 'x'
	for _, pts := range res {
		pts.SortByX()
	}
	return res, nil
}

// SortByY sorts classed points by y coordinate.
func (c *ClassedPoints) SortByY() {
	c.currentSortFunction = c.ySortFunction()
	sort.Sort(c)
}

// SortByX sorts classed points by x coordinate.
func (c *ClassedPoints) SortByX() {
	c.currentSortFunction = c.xSortFunction()
	sort.Sort(c)
}

// compile time check for the sort.Interface implementation.
var _ sort.Interface = &ClassedPoints{}

// Less implements sort.Interface.
func (c *ClassedPoints) Less(i, j int) bool {
	return c.currentSortFunction(i, j)
}

// Swap implements sort.Interface interface.
func (c *ClassedPoints) Swap(i, j int) {
	// swap only the slices of the indexes.
	c.IntSlice[i], c.IntSlice[j] = c.IntSlice[j], c.IntSlice[i]
}

// Len implements sort.Interface interface.
func (c *ClassedPoints) Len() int {
	return c.IntSlice.Size()
}

// XAtIndex gets the 'x' coordinate from the points where the 'i' is the index
// of the IntSlice.
func (c *ClassedPoints) XAtIndex(i int) float32 {
	return (*c.Points)[c.IntSlice[i]].X
}

func (c *ClassedPoints) xSortFunction() func(i int, j int) bool {
	return func(i, j int) bool {
		return c.XAtIndex(i) < c.XAtIndex(j)
	}
}

func (c *ClassedPoints) validateIntSlice() error {
	const processName = "validateIntSlice"
	// check if all classes are within the context of the 'points'.
	for _, idx := range c.IntSlice {
		// the index must be within the range of the points.
		if idx >= (c.Points.Size()) {
			return errors.Errorf(processName, "class id: '%d' is not a valid index in the points of size: %d", idx, c.Points.Size())
		}
	}
	return nil
}

// YAtIndex gets the 'y' coordinate from the points where the 'i' is the index
// of the IntSlice.
func (c *ClassedPoints) YAtIndex(i int) float32 {
	return (*c.Points)[c.IntSlice[i]].Y
}

func (c *ClassedPoints) ySortFunction() func(i int, j int) bool {
	return func(i, j int) bool {
		return c.YAtIndex(i) < c.YAtIndex(j)
	}
}
