/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Component is the component definition enum.
type Component int

// Component enum definitions.
const (
	ComponentConn Component = iota
	ComponentCharacters
	ComponentWords
)

// ConnComponents is a top level call for decomposing the bitmap 'b' by creating components.
// Each component is a part of bitmap where each pixel is connected with at least one neighbor.
// The 'connectivity' is the number of possible directions where the pixel neighbors could be found. The only
// possible values it could take is 4 and 8.
// The connectivity 4 checks the neighbors at top, bottom, left and right, where connectivity 8 checks also upper left, upper right,
// bottom left and bottom right.
// The 'bms' is an optional argument. If it's not nil the components created by the function are added to the 'bms' Bitmaps.
// It sets up 2 temporary bitmaps and for each connected components that is located in raster order, it erase the c.c from one bitmap
// then uses the bounding box to extract component from the 'two' bitmap using XOR operation, and finally erase the component from the second bm.
// Returns bounding boxes of the components. If the bitmaps 'bms' are provided the boxes are added to it's Boxes variable.
func (b *Bitmap) ConnComponents(bms *Bitmaps, connectivity int) (boxa *Boxes, err error) {
	const processName = "Bitmap.ConnComponents"
	if b == nil {
		return nil, errors.Error(processName, "provided empty 'b' bitmap")
	}
	if connectivity != 4 && connectivity != 8 {
		return nil, errors.Error(processName, "connectivity not 4 or 8")
	}
	if bms == nil {
		if boxa, err = b.connComponentsBB(connectivity); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
	} else {
		if boxa, err = b.connComponentsBitmapsBB(bms, connectivity); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
	}
	return boxa, nil
}

// GetComponents determine the bitmap 'b' image components.
func (b *Bitmap) GetComponents(components Component, maxWidth, maxHeight int) (bitmaps *Bitmaps, boxes *Boxes, err error) {
	const processName = "Bitmap.GetComponents"
	if b == nil {
		return nil, nil, errors.Error(processName, "source Bitmap not defined.")
	}

	switch components {
	case ComponentConn, ComponentCharacters, ComponentWords:
	default:
		return nil, nil, errors.Error(processName, "invalid components parameter")
	}
	if b.Zero() {
		boxes = &Boxes{}
		bitmaps = &Bitmaps{}
		return bitmaps, boxes, nil
	}

	switch components {
	case ComponentConn:
		// no pre-processing
		bitmaps = &Bitmaps{}
		if boxes, err = b.ConnComponents(bitmaps, 8); err != nil {
			return nil, nil, errors.Wrap(err, processName, "no preprocessing")
		}
	case ComponentCharacters:
		temp, err := MorphSequence(b, MorphProcess{Operation: MopClosing, Arguments: []int{1, 6}})
		if err != nil {
			return nil, nil, errors.Wrap(err, processName, "characters preprocessing")
		}
		if common.Log.IsLogLevel(common.LogLevelTrace) {
			common.Log.Trace("ComponentCharacters bitmap after closing: %s", temp.String())
		}
		tempBitmaps := &Bitmaps{}
		boxes, err = temp.ConnComponents(tempBitmaps, 8)
		if err != nil {
			return nil, nil, errors.Wrap(err, processName, "characters preprocessing")
		}

		if common.Log.IsLogLevel(common.LogLevelTrace) {
			common.Log.Trace("ComponentCharacters bitmap after connectivity: %s", tempBitmaps.String())
		}

		if bitmaps, err = tempBitmaps.ClipToBitmap(b); err != nil {
			return nil, nil, errors.Wrap(err, processName, "characters preprocessing")
		}
	case ComponentWords:
		redFactor := 1
		var bm *Bitmap
		switch {
		case b.XResolution <= 200:
			bm = b
		case b.XResolution <= 400:
			redFactor = 2
			bm, err = reduceRankBinaryCascade(b, 1, 0, 0, 0)
			if err != nil {
				return nil, nil, errors.Wrap(err, processName, "word preprocess - xres<=400")
			}
		default:
			redFactor = 4
			bm, err = reduceRankBinaryCascade(b, 1, 1, 0, 0)
			if err != nil {
				return nil, nil, errors.Wrap(err, processName, "word preprocess - xres > 400")
			}
		}
		mask, _, err := wordMaskByDilation(bm)
		if err != nil {
			return nil, nil, errors.Wrap(err, processName, "word preprocess")
		}

		bm3, err := expandReplicate(mask, redFactor)
		if err != nil {
			return nil, nil, errors.Wrap(err, processName, "word preprocess")
		}
		tempBitmaps := &Bitmaps{}
		if boxes, err = bm3.ConnComponents(tempBitmaps, 4); err != nil {
			return nil, nil, errors.Wrap(err, processName, "word preprocess, connect expanded")
		}
		if bitmaps, err = tempBitmaps.ClipToBitmap(b); err != nil {
			return nil, nil, errors.Wrap(err, processName, "word preprocess")
		}
	}

	bitmaps, err = bitmaps.SelectBySize(maxWidth, maxHeight, LocSelectIfBoth, SizeSelectIfLTE)
	if err != nil {
		return nil, nil, errors.Wrap(err, processName, "")
	}

	boxes, err = boxes.SelectBySize(maxWidth, maxHeight, LocSelectIfBoth, SizeSelectIfLTE)
	if err != nil {
		return nil, nil, errors.Wrap(err, processName, "")
	}
	return bitmaps, boxes, nil
}

func (b *Bitmap) connComponentsBB(connectivity int) (boxa *Boxes, err error) {
	const processName = "Bitmap.connComponentsBB"
	if connectivity != 4 && connectivity != 8 {
		return nil, errors.Error(processName, "connectivity must be a '4' or '8'")
	}
	if b.Zero() {
		return &Boxes{}, nil
	}
	b.setPadBits(0)

	bm1, err := copyBitmap(nil, b)
	if err != nil {
		return nil, errors.Wrap(err, processName, "bm1")
	}

	stack := &basic.Stack{}
	stack.Aux = &basic.Stack{}
	boxa = &Boxes{}
	var (
		yStart, xStart int
		pt             image.Point
		ok             bool
		box            *image.Rectangle
	)
	for {
		if pt, ok, err = bm1.nextOnPixel(xStart, yStart); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if !ok {
			break
		}
		if box, err = seedFillStackBB(bm1, stack, pt.X, pt.Y, connectivity); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if err = boxa.Add(box); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		xStart = pt.X
		yStart = pt.Y
	}
	return boxa, nil
}

func (b *Bitmap) connComponentsBitmapsBB(bma *Bitmaps, connectivity int) (boxa *Boxes, err error) {
	const processName = "connComponentsBitmapsBB"

	if connectivity != 4 && connectivity != 8 {
		return nil, errors.Error(processName, "connectivity must be a '4' or '8'")
	}
	if bma == nil {
		return nil, errors.Error(processName, "provided nil Bitmaps")
	}
	if len(bma.Values) > 0 {
		return nil, errors.Error(processName, "provided non-empty Bitmaps")
	}
	if b.Zero() {
		return &Boxes{}, nil
	}
	var (
		bm1, bm2, bm3, bm4 *Bitmap
	)
	b.setPadBits(0)
	if bm1, err = copyBitmap(nil, b); err != nil {
		return nil, errors.Wrap(err, processName, "bm1")
	}
	if bm2, err = copyBitmap(nil, b); err != nil {
		return nil, errors.Wrap(err, processName, "bm2")
	}
	stack := &basic.Stack{}
	stack.Aux = &basic.Stack{}
	boxa = &Boxes{}

	var (
		xStart, yStart int
		pt             image.Point
		ok             bool
		box            *image.Rectangle
	)
	for {
		if pt, ok, err = bm1.nextOnPixel(xStart, yStart); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if !ok {
			break
		}

		// fill the stackBB
		if box, err = seedFillStackBB(bm1, stack, pt.X, pt.Y, connectivity); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		if err = boxa.Add(box); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}

		// save the component classes and clear it from bm2
		if bm3, err = bm1.clipRectangle(box, nil); err != nil {
			return nil, errors.Wrap(err, processName, "bm3")
		}
		if bm4, err = bm2.clipRectangle(box, nil); err != nil {
			return nil, errors.Wrap(err, processName, "bm4")
		}

		if _, err = xor(bm3, bm3, bm4); err != nil {
			return nil, errors.Wrap(err, processName, "bm3 ^ bm4")
		}

		if err = bm2.RasterOperation(box.Min.X, box.Min.Y, box.Dx(), box.Dy(), PixSrcXorDst, bm3, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "bm2 -XOR-> bm3")
		}
		bma.AddBitmap(bm3)
		xStart = pt.X
		yStart = pt.Y
	}
	bma.Boxes = *boxa
	return boxa, nil
}

func wordMaskByDilation(s *Bitmap) (mask *Bitmap, dilationSize int, err error) {
	const processName = "Bitmap.wordMaskByDilation"
	if s == nil {
		return nil, 0, errors.Errorf(processName, "'s' bitmap not defined")
	}

	var bm1, bm2 *Bitmap
	if bm1, err = copyBitmap(nil, s); err != nil {
		return nil, 0, errors.Wrap(err, processName, "copy 's'")
	}

	var (
		ncc         [13]int
		total, diff int
	)
	// appropriate for 75 to 150 ppi
	nDil := 12
	naCC := basic.NewNumSlice(nDil + 1)
	naDiff := basic.NewNumSlice(nDil + 1)

	var boxes *Boxes
	for i := 0; i <= nDil; i++ {
		if i == 0 {
			if bm2, err = copyBitmap(nil, bm1); err != nil {
				return nil, 0, errors.Wrap(err, processName, "first bm2")
			}
		} else {
			// dilate the bitmap by sel 2h
			if bm2, err = morphSequence(bm1, MorphProcess{Operation: MopDilation, Arguments: []int{2, 1}}); err != nil {
				return nil, 0, errors.Wrap(err, processName, "dilation bm2")
			}
		}
		if boxes, err = bm2.connComponentsBB(4); err != nil {
			return nil, 0, errors.Wrap(err, processName, "")
		}
		ncc[i] = len(*boxes)
		naCC.AddInt(ncc[i])
		switch i {
		case 0:
			total = ncc[0]
		default:
			diff = ncc[i-1] - ncc[i]
			naDiff.AddInt(diff)
		}
		bm1 = bm2
	}

	check := true
	iBest := 2
	var count, maxDiff int
	for i := 1; i < len(*naDiff); i++ {
		if count, err = naCC.GetInt(i); err != nil {
			return nil, 0, errors.Wrap(err, processName, "Checking best dilation")
		}
		if check && count < int(0.3*float32(total)) {
			iBest = i + 1
			check = false
		}
		if diff, err = naDiff.GetInt(i); err != nil {
			return nil, 0, errors.Wrap(err, processName, "getting naDiff")
		}
		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// NOTE: the resolution here is assumed to be the 'Width' of the bitmap.
	xResolution := s.XResolution
	if xResolution == 0 {
		xResolution = 150
	}
	if xResolution > 110 {
		iBest++
	}
	if iBest < 2 {
		common.Log.Trace("JBIG2 setting iBest to minimum allowable")
		iBest = 2
	}

	dilationSize = iBest + 1
	if mask, err = closeBrick(nil, s, iBest+1, 1); err != nil {
		return nil, 0, errors.Wrap(err, processName, "getting mask failed")
	}
	return mask, dilationSize, nil
}
