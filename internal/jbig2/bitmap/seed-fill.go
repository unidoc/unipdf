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

// maxIterations is a constant used to prevent infitite loops.
const maxIterations = 5000

// seedFillBinary is an algorithm that fills the resultant 'd' bitmap
// with the values from 's' bitmap, expanded in the boundaries of the 'm' bitmap.
// The connectivity defines on how many directions should the 'ON' pixel check
// if it has any other 'ONE' neighbor.
// The connectivity 4 allows to check at: top, bottom, left, right, while the
// connectivity 8 checks also the corners.
// See more at: http://www.vincent-net.com/luc/papers/93ieeeip_recons.pdf.
// If the 'd' is not provided, the function creates one.
func seedFillBinary(d, s, m *Bitmap, connectivity int) (*Bitmap, error) {
	const processName = "seedFillBinary"

	if s == nil {
		return nil, errors.Error(processName, "source bitmap is nil")
	}
	if m == nil {
		return nil, errors.Error(processName, "'mask' bitmap is nil")
	}
	if connectivity != 4 && connectivity != 8 {
		return nil, errors.Error(processName, "connectivity not in range {4,8}")
	}

	var err error
	d, err = copyBitmap(d, s)
	if err != nil {
		return nil, errors.Wrap(err, processName, "copy source to 'd'")
	}

	t := s.createTemplate()
	m.setPadBits(0)

	for i := 0; i < maxIterations; i++ {
		t, err = copyBitmap(t, d)
		if err != nil {
			return nil, errors.Wrapf(err, processName, "iteration: %d", i)
		}
		if err = seedfillBinaryLow(d, m, connectivity); err != nil {
			return nil, errors.Wrapf(err, processName, "iteration: %d", i)
		}
		if t.Equals(d) {
			// binary seed fill converged.
			break
		}
	}
	return d, nil
}

func seedfillBinaryLow(s *Bitmap, m *Bitmap, connectivity int) (err error) {
	const processName = "seedfillBinaryLow"
	h := min(s.Height, m.Height)
	wpl := min(s.RowStride, m.RowStride)

	switch connectivity {
	case 4:
		err = seedfillBinaryLow4(s, m, h, wpl)
	case 8:
		err = seedfillBinaryLow8(s, m, h, wpl)
	default:
		return errors.Errorf(processName, "connectivity must be 4 or 8 - is: '%d'", connectivity)
	}
	if err != nil {
		return errors.Wrap(err, processName, "")
	}
	return nil
}

func seedfillBinaryLow4(s, m *Bitmap, h, wpl int) (err error) {
	const processName = "seedfillBinaryLow4"
	var (
		i, j, lineS, lineM                                  int
		bt, mask, btAbove, btLeft, btBelow, btRight, btPrev byte
	)
	for i = 0; i < h; i++ {
		lineS = i * s.RowStride
		lineM = i * m.RowStride

		for j = 0; j < wpl; j++ {
			bt, err = s.GetByte(lineS + j)
			if err != nil {
				return errors.Wrap(err, processName, "first get")
			}
			mask, err = m.GetByte(lineM + j)
			if err != nil {
				return errors.Wrap(err, processName, "second get")
			}

			if i > 0 {
				btAbove, err = s.GetByte(lineS - s.RowStride + j)
				if err != nil {
					return errors.Wrap(err, processName, "i > 0")
				}
				bt |= btAbove
			}
			if j > 0 {
				btLeft, err = s.GetByte(lineS + j - 1)
				if err != nil {
					return errors.Wrap(err, processName, "j > 0")
				}
				bt |= btLeft << 7
			}
			bt &= mask

			if bt == 0 || (^bt) == 0 {
				if err = s.SetByte(lineS+j, bt); err != nil {
					return errors.Wrap(err, processName, "bt == 0 || (^bt) == 0")
				}
				continue
			}

			for {
				btPrev = bt
				bt = (bt | (bt >> 1) | (bt << 1)) & mask
				if (bt ^ btPrev) == 0 {
					if err = s.SetByte(lineS+j, bt); err != nil {
						return errors.Wrap(err, processName, "setting prev byte")
					}
					break
				}
			}
		}
	}

	for i = h - 1; i >= 0; i-- {
		lineS = i * s.RowStride
		lineM = i * m.RowStride

		for j = wpl - 1; j >= 0; j-- {
			if bt, err = s.GetByte(lineS + j); err != nil {
				return errors.Wrap(err, processName, "reverse first get")
			}
			if mask, err = m.GetByte(lineM + j); err != nil {
				return errors.Wrap(err, processName, "reverse get mask byte")
			}

			if i < h-1 {
				if btBelow, err = s.GetByte(lineS + s.RowStride + j); err != nil {
					return errors.Wrap(err, processName, "reverse i < h -1")
				}
				bt |= btBelow
			}
			if j < wpl-1 {
				if btRight, err = s.GetByte(lineS + j + 1); err != nil {
					return errors.Wrap(err, processName, "reverse j < wpl - 1")
				}
				bt |= btRight >> 7
			}
			bt &= mask

			if bt == 0 || (^bt) == 0 {
				if err = s.SetByte(lineS+j, bt); err != nil {
					return errors.Wrap(err, processName, "setting masked byte failed")
				}
				continue
			}

			for {
				btPrev = bt
				bt = (bt | (bt >> 1) | (bt << 1)) & mask
				if (bt ^ btPrev) == 0 {
					if err = s.SetByte(lineS+j, bt); err != nil {
						return errors.Wrap(err, processName, "reverse setting prev byte")
					}
					break
				}
			}
		}
	}
	return nil
}

func seedfillBinaryLow8(s, m *Bitmap, h, wpl int) (err error) {
	const processName = "seedfillBinaryLow8"
	var (
		i, j, lineS, lineM                                        int
		bt, mask, btAbove, btLeft, btBelow, btRight, btPrev, temp byte
	)
	for i = 0; i < h; i++ {
		lineS = i * s.RowStride
		lineM = i * m.RowStride

		for j = 0; j < wpl; j++ {
			if bt, err = s.GetByte(lineS + j); err != nil {
				return errors.Wrap(err, processName, "get source byte")
			}
			if mask, err = m.GetByte(lineM + j); err != nil {
				return errors.Wrap(err, processName, "get mask byte")
			}

			if i > 0 {
				if btAbove, err = s.GetByte(lineS - s.RowStride + j); err != nil {
					return errors.Wrap(err, processName, "i > 0 byte")
				}
				bt |= btAbove | (btAbove << 1) | (btAbove >> 1)
				if j > 0 {
					if temp, err = s.GetByte(lineS - s.RowStride + j - 1); err != nil {
						return errors.Wrap(err, processName, "i > 0 && j > 0 byte")
					}
					bt |= temp << 7
				}
				if j < wpl-1 {
					if temp, err = s.GetByte(lineS - s.RowStride + j + 1); err != nil {
						return errors.Wrap(err, processName, "j < wpl - 1 byte")
					}
					bt |= temp >> 7
				}
			}
			if j > 0 {
				if btLeft, err = s.GetByte(lineS + j - 1); err != nil {
					return errors.Wrap(err, processName, "j > 0")
				}
				bt |= btLeft << 7
			}
			bt &= mask

			if bt == 0 || ^bt == 0 {
				if err = s.SetByte(lineS+j, bt); err != nil {
					return errors.Wrap(err, processName, "setting empty byte")
				}
			}

			for {
				btPrev = bt
				bt = (bt | (bt >> 1) | (bt << 1)) & mask
				if (bt ^ btPrev) == 0 {
					if err = s.SetByte(lineS+j, bt); err != nil {
						return errors.Wrap(err, processName, "setting prev byte")
					}
					break
				}
			}
		}
	}

	for i = h - 1; i >= 0; i-- {
		lineS = i * s.RowStride
		lineM = i * m.RowStride

		for j = wpl - 1; j >= 0; j-- {
			if bt, err = s.GetByte(lineS + j); err != nil {
				return errors.Wrap(err, processName, "reverse get source byte")
			}
			if mask, err = m.GetByte(lineM + j); err != nil {
				return errors.Wrap(err, processName, "reverse get mask byte")
			}

			if i < h-1 {
				if btBelow, err = s.GetByte(lineS + s.RowStride + j); err != nil {
					return errors.Wrap(err, processName, "i < h - 1 -> get source byte")
				}
				bt |= btBelow | (btBelow << 1) | btBelow>>1
				if j > 0 {
					if temp, err = s.GetByte(lineS + s.RowStride + j - 1); err != nil {
						return errors.Wrap(err, processName, "i < h-1 & j > 0 -> get source byte")
					}
					bt |= temp << 7
				}
				if j < wpl-1 {
					if temp, err = s.GetByte(lineS + s.RowStride + j + 1); err != nil {
						return errors.Wrap(err, processName, "i < h-1 && j <wpl-1 -> get source byte")
					}
					bt |= temp >> 7
				}
			}
			if j < wpl-1 {
				if btRight, err = s.GetByte(lineS + j + 1); err != nil {
					return errors.Wrap(err, processName, "j < wpl -1 -> get source byte")
				}
				bt |= btRight >> 7
			}
			bt &= mask

			if bt == 0 || (^bt) == 0 {
				if err = s.SetByte(lineS+j, bt); err != nil {
					return errors.Wrap(err, processName, "set masked byte")
				}
			}

			for {
				btPrev = bt
				bt = (bt | (bt >> 1) | (bt << 1)) & mask
				if (bt ^ btPrev) == 0 {
					if err = s.SetByte(lineS+j, bt); err != nil {
						return errors.Wrap(err, processName, "reverse set prev byte")
					}
					break
				}
			}
		}
	}
	return nil
}

// seedFillStack4BB does the Paul Heckbert's stack based on 4 and 8 connectivity seedfill algorithm.
// The function takes 's' Bitmap and removes pixel at 'x' and 'y' coordinates as well as all the 4 or 8-connected 'ON' pixels.
// Returns the bounding box image.Rectangle of the removed 4 or 8-connected component.
func seedFillStackBB(s *Bitmap, stack *basic.Stack, x, y, connectivity int) (box *image.Rectangle, err error) {
	const processName = "seedFillStackBB"
	if s == nil {
		return nil, errors.Error(processName, "provided nil 's' Bitmap")
	}
	if stack == nil {
		return nil, errors.Error(processName, "provided nil 'stack'")
	}

	switch connectivity {
	case 4:
		if box, err = seedFillStack4BB(s, stack, x, y); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		return box, nil
	case 8:
		if box, err = seedFillStack8BB(s, stack, x, y); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		return box, nil
	default:
		return nil, errors.Errorf(processName, "connectivity is not 4 or 8: '%d'", connectivity)
	}
}

// seedFillStack4BB does the Paul Heckbert's stack based 4 connectivity seedfill algorithm.
// The function takes 's' Bitmap and removes pixel at 'x' and 'y' coordinates as well as all the 4-connected 'ON' pixels.
// Returns the bounding box image.Rectangle of the removed 4-connected component.
func seedFillStack4BB(s *Bitmap, stack *basic.Stack, x, y int) (box *image.Rectangle, err error) {
	const processName = "seedFillStackBB"
	if s == nil {
		return nil, errors.Error(processName, "provided nil 's' Bitmap")
	}
	if stack == nil {
		return nil, errors.Error(processName, "provided nil 'stack'")
	}

	w, h := s.Width, s.Height
	xMax := w - 1
	yMax := h - 1
	// check if the provided 'x' and 'y' are in the given image boundaries.
	// Also check if the bit value at the 'x' and 'y' position is  'ON'.
	if x < 0 || x > xMax || y < 0 || y > yMax || !s.GetPixel(x, y) {
		return nil, nil
	}

	// define the initial value limits for the 'x' and 'y' in the 'rect' bounding box.
	var rect *image.Rectangle
	rect, err = Rect(100000, 100000, 0, 0)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	if err = pushFillSegmentBoundingBox(stack, x, x, y, 1, yMax, rect); err != nil {
		return nil, errors.Wrap(err, processName, "initial push")
	}
	if err = pushFillSegmentBoundingBox(stack, x, x, y+1, -1, yMax, rect); err != nil {
		return nil, errors.Wrap(err, processName, "2nd initial push")
	}

	rect.Min.X, rect.Max.X = x, x
	rect.Min.Y, rect.Max.Y = y, y

	var (
		out    *fillSegment
		xStart int
	)

	// pop all the segments off the stack and fill a neighboring scan lines.
	for stack.Len() > 0 {
		if out, err = popFillSegment(stack); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		y = out.y

		// Segment scan line 'out.y - out.dy' for each 'x'
		// in range out.xLeft <= x <= out.xRight.
		// Explore the adjacent pixels in scan line 'y'.
		// There are three possible regions:
		// - to the left of out.xLeft - 1
		// - between out.xLeft and out.xRight
		// - to the right of the out.xRight + 1
		for x = out.xLeft; x >= 0 && s.GetPixel(x, y); x-- {
			if err = s.SetPixel(x, y, 0); err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		}

		// check if the out.xLeft pixel was off and was not cleared.
		if x >= out.xLeft {
			// skip the bits to the next 'ON' bit in the provided bounds
			for x++; x <= out.xRight && x <= xMax && !s.GetPixel(x, y); x++ {
			}
			xStart = x

			// check if the x is still in the bounds.
			if !(x <= out.xRight && x <= xMax) {
				continue
			}
		} else {
			xStart = x + 1
			// check if there is a leak on left side
			if xStart < out.xLeft-1 {
				if err = pushFillSegmentBoundingBox(stack, xStart, out.xLeft-1, out.y, -out.dy, yMax, rect); err != nil {
					return nil, errors.Wrap(err, processName, "leak on left side")
				}
			}
			x = out.xLeft + 1
		}

		for {
			for ; x <= xMax && s.GetPixel(x, y); x++ {
				if err = s.SetPixel(x, y, 0); err != nil {
					return nil, errors.Wrap(err, processName, "2nd set")
				}
			}

			// push the x,y bb
			if err = pushFillSegmentBoundingBox(stack, xStart, x-1, out.y, out.dy, yMax, rect); err != nil {
				return nil, errors.Wrap(err, processName, "normal push")
			}

			// check if there is a leak on the right side
			if x > out.xRight+1 {
				if err = pushFillSegmentBoundingBox(stack, out.xRight+1, x-1, out.y, -out.dy, yMax, rect); err != nil {
					return nil, errors.Wrap(err, processName, "leak on right side")
				}
			}

			// skip to the next 'OFF' bit.
			for x++; x <= out.xRight && x <= xMax && !s.GetPixel(x, y); x++ {
			}
			xStart = x

			if !(x <= out.xRight && x <= xMax) {
				break
			}
		}
	}
	rect.Max.X++
	rect.Max.Y++
	return rect, nil
}

func seedFillStack8BB(s *Bitmap, stack *basic.Stack, x, y int) (box *image.Rectangle, err error) {
	const processName = "seedFillStackBB"
	if s == nil {
		return nil, errors.Error(processName, "provided nil 's' Bitmap")
	}
	if stack == nil {
		return nil, errors.Error(processName, "provided nil 'stack'")
	}

	w, h := s.Width, s.Height
	xMax := w - 1
	yMax := h - 1
	// check if the provided 'x' and 'y' are in the given image boundaries.
	// Also check if the bit value at the 'x' and 'y' position is  'ON'.
	if x < 0 || x > xMax || y < 0 || y > yMax || !s.GetPixel(x, y) {
		return nil, nil
	}

	// define the initial value limits for the 'x' and 'y' in the 'rect' bounding box.
	rect := image.Rect(100000, 100000, 0, 0)

	if err = pushFillSegmentBoundingBox(stack, x, x, y, 1, yMax, &rect); err != nil {
		return nil, errors.Wrap(err, processName, "initial push")
	}
	if err = pushFillSegmentBoundingBox(stack, x, x, y+1, -1, yMax, &rect); err != nil {
		return nil, errors.Wrap(err, processName, "2nd initial push")
	}

	rect.Min.X, rect.Max.X = x, x
	rect.Min.Y, rect.Max.Y = y, y

	var (
		out    *fillSegment
		xStart int
	)
	for stack.Len() > 0 {
		// pop the fillSegment from the stack and use it's values in the line scanning process.
		if out, err = popFillSegment(stack); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		y = out.y
		// explore adjacent pixels in scan line 'y'. There are three possible regions:
		// - to the left of out.xLeft - 1
		// - between out.xLeft and out.xRight
		// - to the right of the out.xRight +1
		for x = out.xLeft - 1; x >= 0 && s.GetPixel(x, y); x-- {
			if err = s.SetPixel(x, y, 0); err != nil {
				return nil, errors.Wrap(err, processName, "1st set")
			}
		}

		// check if the pixel at 'out.xLeft' was off and wasn't cleared.
		if x >= out.xLeft-1 {
			for {
				// skip the bits until 'ON' pixel is found.
				for x++; x <= out.xRight+1 && x <= xMax && !s.GetPixel(x, y); x++ {
				}
				xStart = x

				// if x > out.xRight+1 || x > xMax {
				if !(x <= out.xRight+1 && x <= xMax) {
					break
				}

				for ; x <= xMax && s.GetPixel(x, y); x++ {
					if err = s.SetPixel(x, y, 0); err != nil {
						return nil, errors.Wrap(err, processName, "2nd set")
					}
				}

				// push the x,y bb
				if err = pushFillSegmentBoundingBox(stack, xStart, x-1, out.y, out.dy, yMax, &rect); err != nil {
					return nil, errors.Wrap(err, processName, "normal push")
				}

				// check if there is a leak on the right side
				if x > out.xRight {
					if err = pushFillSegmentBoundingBox(stack, out.xRight+1, x-1, out.y, -out.dy, yMax, &rect); err != nil {
						return nil, errors.Wrap(err, processName, "leak on right side")
					}
				}
			}
			continue
		}
		xStart = x + 1
		// check if there is a leak on left side
		// possible when the 'x' was < out.xLeft
		if xStart < out.xLeft {
			if err = pushFillSegmentBoundingBox(stack, xStart, out.xLeft-1, out.y, -out.dy, yMax, &rect); err != nil {
				return nil, errors.Wrap(err, processName, "leak on left side")
			}
		}
		x = out.xLeft

		for {
			for ; x <= xMax && s.GetPixel(x, y); x++ {
				if err = s.SetPixel(x, y, 0); err != nil {
					return nil, errors.Wrap(err, processName, "2nd set")
				}
			}

			// push the x,y bb
			if err = pushFillSegmentBoundingBox(stack, xStart, x-1, out.y, out.dy, yMax, &rect); err != nil {
				return nil, errors.Wrap(err, processName, "normal push")
			}

			// check if there is a leak on the right side
			if x > out.xRight {
				if err = pushFillSegmentBoundingBox(stack, out.xRight+1, x-1, out.y, -out.dy, yMax, &rect); err != nil {
					return nil, errors.Wrap(err, processName, "leak on right side")
				}
			}

			// skip the bits until 'ON' pixel is found.
			for x++; x <= out.xRight+1 && x <= xMax && !s.GetPixel(x, y); x++ {
			}
			xStart = x

			// if x > out.xRight+1 || x > xMax {
			if !(x <= out.xRight+1 && x <= xMax) {
				break
			}
		}
	}
	rect.Max.X++
	rect.Max.Y++
	return &rect, nil
}

// fillSegment is the Heckber seedfill algorithm helper that holds the current run segment information.
type fillSegment struct {
	// xLeft is the left edge of run
	xLeft int
	// xRight is the right edge of run
	xRight int
	// y is the currount run 'y'
	y int
	// direction of the parent segment: 1 above, -1 below.
	dy int
}

// pushFillSegmentBoundingBox is a helper function for push and poping fillSegs.
func pushFillSegmentBoundingBox(stack *basic.Stack, xLeft, xRight, y, dy, yMax int, rect *image.Rectangle) (err error) {
	const processName = "pushFillSegmentBoundingBox"
	if stack == nil {
		return errors.Error(processName, "nil stack provided")
	}
	if rect == nil {
		return errors.Error(processName, "provided nil image.Rectangle")
	}

	rect.Min.X = basic.Min(rect.Min.X, xLeft)
	rect.Max.X = basic.Max(rect.Max.X, xRight)
	rect.Min.Y = basic.Min(rect.Min.Y, y)
	rect.Max.Y = basic.Max(rect.Max.Y, y)
	if !(y+dy >= 0 && y+dy <= yMax) {
		return nil
	}

	if stack.Aux == nil {
		return errors.Error(processName, "auxStack not defined")
	}

	var fSeg *fillSegment
	fv, ok := stack.Aux.Pop()
	if ok {
		if fSeg, ok = fv.(*fillSegment); !ok {
			return errors.Error(processName, "auxStack data is not a *fillSegment")
		}
	} else {
		fSeg = &fillSegment{}
	}
	fSeg.xLeft = xLeft
	fSeg.xRight = xRight
	fSeg.y = y
	fSeg.dy = dy
	stack.Push(fSeg)
	return nil
}

// popFillSegment is a helper function for the Heckber algorithm seedFills.
func popFillSegment(stack *basic.Stack) (out *fillSegment, err error) {
	const processName = "popFillSegment"
	if stack == nil {
		return nil, errors.Error(processName, "nil stack provided")
	}
	if stack.Aux == nil {
		return nil, errors.Error(processName, "auxStack not defined")
	}
	fv, ok := stack.Pop()
	if !ok {
		return nil, nil
	}
	fSeg, ok := fv.(*fillSegment)
	if !ok {
		return nil, errors.Error(processName, "stack doesn't contain *fillSegment")
	}
	out = &fillSegment{fSeg.xLeft, fSeg.xRight, fSeg.y + fSeg.dy, fSeg.dy}
	stack.Aux.Push(fSeg)
	return out, nil
}
