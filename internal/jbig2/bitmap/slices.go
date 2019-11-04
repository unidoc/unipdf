/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"
	"strings"

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
	if pt == nil {
		return errors.Error("Points.Add", "pointer to 'pt' is nil")
	}
	*p = append(*p, (*pt)...)
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

// GetGeometry gets the geometry 'x' and 'y' of the point at the 'i' index.
// Returns error if the index is out of range.
func (p Points) GetGeometry(i int) (x, y float32, err error) {
	if i > len(p)-1 {
		return 0, 0, errors.Errorf("Points.Get", "index: '%d' out of range", i)
	}
	pt := p[i]
	return pt.X, pt.Y, nil
}

// Bitmaps is the structure that contains slice of the bitmaps and the bounding boxes.
// It allows to safely get the Bitmap and the bounding boxes.
type Bitmaps struct {
	Values []*Bitmap
	Boxes  []*image.Rectangle
}

// AddBitmap adds the bitmap 'bm' to the 'b' Bitmaps Values.
func (b *Bitmaps) AddBitmap(bm *Bitmap) {
	b.Values = append(b.Values, bm)
}

// AddBox adds the 'box' to the 'b' Bitmaps.
func (b *Bitmaps) AddBox(box *image.Rectangle) {
	b.Boxes = append(b.Boxes, box)
}

// ClipToBitmap returns a Bitmaps where each Bitmap is 'AND'ed with
// with the associated region stored in box with the 's' Bitmap.
func (b *Bitmaps) ClipToBitmap(s *Bitmap) (*Bitmaps, error) {
	const processName = "Bitmaps.ClipToBitmap"

	if b == nil {
		return nil, errors.Error(processName, "Bitmaps not defined")
	}
	if s == nil {
		return nil, errors.Error(processName, "source bitmap not defined")
	}

	n := len(b.Values)
	d := &Bitmaps{Values: make([]*Bitmap, n), Boxes: make([]*image.Rectangle, n)}

	var (
		bm, bmC *Bitmap
		box     *image.Rectangle
		err     error
	)

	for i := 0; i < n; i++ {
		if bm, err = b.GetBitmap(i); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if box, err = b.GetBox(i); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if bmC, err = s.clipRectangle(box, nil); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if bmC, err = bmC.And(bm); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		d.Values[i] = bmC
		d.Boxes[i] = box
	}
	return d, nil
}

// CountPixels counts the pixels for all the bitmaps and stores into *basic.NumSlice.
func (b *Bitmaps) CountPixels() *basic.NumSlice {
	ns := &basic.NumSlice{}
	for _, bm := range b.Values {
		ns.AddInt(bm.CountPixels())
	}
	return ns
}

// GetBitmap gets the bitmap at the 'i' index.
// If the index is out of possible range the function returns error.
func (b *Bitmaps) GetBitmap(i int) (*Bitmap, error) {
	const processName = "GetBitmap"
	if b == nil {
		return nil, errors.Error(processName, "provided nil Bitmaps")
	}
	if i > len(b.Values)-1 {
		return nil, errors.Errorf(processName, "index: '%d' out of range", i)
	}
	return b.Values[i], nil
}

// GetBox gets the Box at the 'i' index.
// If the index is out of range the function returns error.
func (b *Bitmaps) GetBox(i int) (*image.Rectangle, error) {
	const processName = "GetBox"
	if b == nil {
		return nil, errors.Error(processName, "provided nil 'Bitmaps'")
	}
	if i > len(b.Boxes)-1 {
		return nil, errors.Errorf(processName, "index: '%d' out of range", i)
	}
	return b.Boxes[i], nil
}

// SelectBySize selects provided bitmaps by provided 'width', 'height' location filter 'tp' and size comparison 'relation.
// Returns 'b' bitmap if it's empty or all the bitmaps matches the pattern.
func (b *Bitmaps) SelectBySize(width, height int, tp LocationFilter, relation SizeComparison) (d *Bitmaps, err error) {
	const processName = "Bitmaps.SelectBySize"
	if b == nil {
		return nil, errors.Error(processName, "'b' Bitmaps not defined")
	}
	// check the location filter type
	switch tp {
	case LocSelectWidth, LocSelectHeight, LocSelectIfEither, LocSelectIfBoth:
	default:
		return nil, errors.Errorf(processName, "provided invalid location filter type: %d", tp)
	}

	// check the relation value
	switch relation {
	case SizeSelectIfLT, SizeSelectIfGT, SizeSelectIfLTE, SizeSelectIfGTE:
	default:
		return nil, errors.Errorf(processName, "invalid relation: '%d'", relation)
	}

	na, err := b.makeSizeIndicator(width, height, tp, relation)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	d, err = b.selectByIndicator(na)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return d, nil
}

func (b *Bitmaps) String() string {
	sb := strings.Builder{}
	for _, bm := range b.Values {
		sb.WriteString(bm.String())
		sb.WriteRune('\n')
	}
	return sb.String()
}

func (b *Bitmaps) makeSizeIndicator(width, height int, tp LocationFilter, relation SizeComparison) (na *basic.NumSlice, err error) {
	const processName = "Bitmaps.makeSizeIndicator"
	if b == nil {
		return nil, errors.Error(processName, "bitmaps 'b' not defined")
	}
	switch tp {
	case LocSelectWidth, LocSelectHeight, LocSelectIfEither, LocSelectIfBoth:
	default:
		return nil, errors.Errorf(processName, "provided invalid location filter type: %d", tp)
	}

	switch relation {
	case SizeSelectIfLT, SizeSelectIfGT, SizeSelectIfLTE, SizeSelectIfGTE:
	default:
		return nil, errors.Errorf(processName, "invalid relation: '%d'", relation)
	}
	na = &basic.NumSlice{}
	var (
		ival, w, h int
		bm         *Bitmap
	)
	for _, bm = range b.Values {
		ival = 0
		w, h = bm.Width, bm.Height
		switch tp {
		case LocSelectWidth:
			if (relation == SizeSelectIfLT && w < width) ||
				(relation == SizeSelectIfGT && w > width) ||
				(relation == SizeSelectIfLTE && w <= width) ||
				(relation == SizeSelectIfGTE && w >= width) {
				ival = 1
			}
		case LocSelectHeight:
			if (relation == SizeSelectIfLT && h < height) ||
				(relation == SizeSelectIfGT && h > height) ||
				(relation == SizeSelectIfLTE && h <= height) ||
				(relation == SizeSelectIfGTE && h >= height) {
				ival = 1
			}
		case LocSelectIfEither:
			if (relation == SizeSelectIfLT && (w < width || h < height)) ||
				(relation == SizeSelectIfGT && (w > width || h > height)) ||
				(relation == SizeSelectIfLTE && (w <= width || h <= height)) ||
				(relation == SizeSelectIfGTE && (w >= width || h >= height)) {
				ival = 1
			}
		case LocSelectIfBoth:
			if (relation == SizeSelectIfLT && (w < width && h < height)) ||
				(relation == SizeSelectIfGT && (w > width && h > height)) ||
				(relation == SizeSelectIfLTE && (w <= width && h <= height)) ||
				(relation == SizeSelectIfGTE && (w >= width && h >= height)) {
				ival = 1
			}
		}
		na.AddInt(ival)
	}
	return na, nil
}

func (b *Bitmaps) selectByIndicator(na *basic.NumSlice) (d *Bitmaps, err error) {
	const processName = "Bitmaps.selectByIndicator"
	if b == nil {
		return nil, errors.Error(processName, "'b' bitmaps not defined")
	}
	if na == nil {
		return nil, errors.Error(processName, "'na' indicators not defined")
	}
	if len(b.Values) == 0 {
		return b, nil
	}

	if len(*na) != len(b.Values) {
		return nil, errors.Errorf(processName, "na length: %d, is different than bitmaps: %d", len(*na), len(b.Values))
	}
	var ival, i, changeCount int
	for i = 0; i < len(*na); i++ {
		if ival, err = na.GetInt(i); err != nil {
			return nil, errors.Wrap(err, processName, "first check")
		}
		if ival == 1 {
			changeCount++
		}
	}
	if changeCount == len(b.Values) {
		return b, nil
	}
	d = &Bitmaps{}
	hasBoxes := len(b.Values) == len(b.Boxes)
	for i = 0; i < len(*na); i++ {
		if ival = int((*na)[i]); ival == 0 {
			continue
		}
		d.Values = append(d.Values, b.Values[i])
		if hasBoxes {
			d.Boxes = append(d.Boxes, b.Boxes[i])
		}
	}
	return d, nil
}

// BitmapsArray is the struct that contains slice of the 'Bitmaps',
// with the bounding boxes around each 'Bitmaps'.
type BitmapsArray struct {
	Values []*Bitmaps
	Boxes  []*image.Rectangle
}

// AddBitmaps adds the 'bm' Bitmaps to the 'b'.Values.
func (b *BitmapsArray) AddBitmaps(bm *Bitmaps) {
	b.Values = append(b.Values, bm)
}

// AddBox adds the 'box' to the 'b' Boxes.
func (b *BitmapsArray) AddBox(box *image.Rectangle) {
	b.Boxes = append(b.Boxes, box)
}

// GetBitmaps gets the 'Bitmaps' at the 'i' position.
// Returns error if the index is out of range.
func (b *BitmapsArray) GetBitmaps(i int) (*Bitmaps, error) {
	const processName = "BitmapsArray.GetBitmaps"
	if b == nil {
		return nil, errors.Error(processName, "provided nil 'BitmapsArray'")
	}
	if i > len(b.Values)-1 {
		return nil, errors.Errorf(processName, "index: '%d' out of range", i)
	}
	return b.Values[i], nil
}

// GetBox gets the boundary box at 'i' position.
// Returns error if the index is out of range.
func (b *BitmapsArray) GetBox(i int) (*image.Rectangle, error) {
	const processName = "BitmapsArray.GetBox"
	if b == nil {
		return nil, errors.Error(processName, "provided nil 'BitmapsArray'")
	}
	if i > len(b.Boxes)-1 {
		return nil, errors.Errorf(processName, "index: '%d' out of range", i)
	}
	return b.Boxes[i], nil
}
