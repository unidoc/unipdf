/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Boxes is the wrapper over the slice of image.Rectangles that allows to
// get and add the image.Rectangle safely.
type Boxes []*image.Rectangle

// Add adds the 'box' to the provided 'Boxes'.
func (b *Boxes) Add(box *image.Rectangle) error {
	if b == nil {
		return errors.Error("Boxes.Add", "'Boxes' not defined")
	}
	*b = append(*b, box)
	return nil
}

// Get gets the box at 'i' index. Returns error if the index 'i' is out of range.
func (b *Boxes) Get(i int) (*image.Rectangle, error) {
	const processName = "Boxes.Get"
	if b == nil {
		return nil, errors.Error(processName, "'Boxes' not defined")
	}
	if i > len(*b)-1 {
		return nil, errors.Errorf(processName, "index: '%d' out of range", i)
	}
	return (*b)[i], nil
}

// SelectBySize select sthe boxes 'b' by provided 'width' 'height', location filter 'tp' and 'relation'.
// If nothing changes the function returns the boxes 'b' by itself.
func (b *Boxes) SelectBySize(width, height int, tp LocationFilter, relation SizeComparison) (result *Boxes, err error) {
	const processName = "Boxes.SelectBySize"
	if b == nil {
		return nil, errors.Error(processName, "boxes 'b' not defined")
	}
	// return boxes 'b' if its' empty.
	if len(*b) == 0 {
		return b, nil
	}

	switch tp {
	case LocSelectWidth, LocSelectHeight, LocSelectIfEither, LocSelectIfBoth:
	default:
		return nil, errors.Errorf(processName, "invalid filter type: %d", tp)
	}
	switch relation {
	case SizeSelectIfLT, SizeSelectIfGT, SizeSelectIfLTE, SizeSelectIfGTE:
	default:
		return nil, errors.Errorf(processName, "invalid relation type: '%d'", tp)
	}

	// get the indexes of the boxes that matches given arguments.
	na := b.makeSizeIndicator(width, height, tp, relation)
	// select the boxes from the 'b' that matches indexes found in the 'na'.
	d, err := b.selectWithIndicator(na)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return d, nil
}

func (b *Boxes) makeSizeIndicator(width, height int, tp LocationFilter, relation SizeComparison) *basic.NumSlice {
	na := &basic.NumSlice{}

	var ival, w, h int
	for _, box := range *b {
		ival = 0
		w, h = box.Dx(), box.Dy()
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
			if (relation == SizeSelectIfLT && (h < height || w < width)) ||
				(relation == SizeSelectIfGT && (h > height || w > width)) ||
				(relation == SizeSelectIfLTE && (h <= height || w <= width)) ||
				(relation == SizeSelectIfGTE && (h >= height || w >= width)) {
				ival = 1
			}
		case LocSelectIfBoth:
			if (relation == SizeSelectIfLT && (h < height && w < width)) ||
				(relation == SizeSelectIfGT && (h > height && w > width)) ||
				(relation == SizeSelectIfLTE && (h <= height && w <= width)) ||
				(relation == SizeSelectIfGTE && (h >= height && w >= width)) {
				ival = 1
			}
		}
		na.AddInt(ival)
	}
	return na
}

// selectWithIndicator selects the boxes based on the 'na' NumSlice that has indexes of
// the boxes that should be selected.
func (b *Boxes) selectWithIndicator(na *basic.NumSlice) (d *Boxes, err error) {
	const processName = "Boxes.selectWithIndicator"
	if b == nil {
		return nil, errors.Error(processName, "boxes 'b' not defined")
	}
	if na == nil {
		return nil, errors.Error(processName, "'na' not defined")
	}

	if len(*na) != len(*b) {
		return nil, errors.Error(processName, "boxes 'b' has different size than 'na'")
	}

	var ival, changeCount int
	for i := 0; i < len(*na); i++ {
		if ival, err = na.GetInt(i); err != nil {
			return nil, errors.Wrap(err, processName, "checking count")
		}
		if ival == 1 {
			changeCount++
		}
	}

	if changeCount == len(*b) {
		return b, nil
	}
	boxes := Boxes{}

	for i := 0; i < len(*na); i++ {
		// we don't have to use the 'GetInt' now - it was already checked on the first iteration.
		ival = int((*na)[i])
		if ival == 0 {
			continue
		}
		// as in the 'na' we don't have to check if the boxes 'b' has the i'th box - it was already checked
		// with the boxes size.
		boxes = append(boxes, (*b)[i])
	}
	d = &boxes
	return d, nil
}

// ClipBoxToRectangle clips the image.Rectangle 'box' for the provided wi, hi which are rectangle representing image.
// The UL corner of the 'box' is assumed to be at (0,0) and LR at ('wi' - 1, 'hi' - 1)
func ClipBoxToRectangle(box *image.Rectangle, wi, hi int) (out *image.Rectangle, err error) {
	const processName = "ClipBoxToRectangle"
	if box == nil {
		return nil, errors.Error(processName, "'box' not defined")
	}
	if box.Min.X >= wi || box.Min.Y >= hi || box.Max.X <= 0 || box.Max.Y <= 0 {
		return nil, errors.Error(processName, "'box' outside rectangle")
	}
	out = &(*box)
	if out.Min.X < 0 {
		out.Max.X += out.Min.X
		out.Min.X = 0
	}
	if out.Min.Y < 0 {
		out.Max.Y += out.Min.Y
		out.Min.Y = 0
	}

	// if w > wi
	if out.Max.X > wi {
		// max - min = wi - min
		out.Max.X = wi
	}

	// if h > hi
	if out.Max.Y > hi {
		// max - min = hi - min
		out.Max.Y = hi
	}
	return out, nil
}

// Rect create new rectangle box with the negative coordinates rules.
func Rect(x, y, w, h int) (*image.Rectangle, error) {
	const processName = "bitmap.Rect"
	if x < 0 {
		w += x
		x = 0
		if w <= 0 {
			return nil, errors.Errorf(processName, "x:'%d' < 0 and w: '%d' <= 0", x, w)
		}
	}
	if y < 0 {
		h += y
		y = 0
		if h <= 0 {
			return nil, errors.Error(processName, "y < 0 and box off +quad")
		}
	}
	box := image.Rect(x, y, x+w, y+h)
	return &box, nil
}
