/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"image"
	"math"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// JbAddedPixels is the size of the border added around pix of each c.c. for further processing.
const JbAddedPixels = 6

// For PixHausTest, PixRankHausTest and PixCorrelationScore
// the values should be or greater.
const (
	MaxDiffWidth  = 2
	MaxDiffHeight = 2
)

const (
	// MaxConnCompWidth is the default max cc width.
	MaxConnCompWidth = 350
	// MaxCharCompWidth is the default max char width.
	MaxCharCompWidth = 350
	// MaxWordCompWidth is the default max word width.
	MaxWordCompWidth = 1000
	// MaxCompHeight is the default max component height.
	MaxCompHeight = 120
)

// finalAlignmentPositioning gets the best match position for the provided arguments.
func finalAlignmentPositioning(s *bitmap.Bitmap, x, y, iDelX, iDelY int, t *bitmap.Bitmap) (pt image.Point, err error) {
	const processName = "finalAlignmentPositioning"
	if s == nil {
		return pt, errors.Error(processName, "source not provided")
	}
	if t == nil {
		return pt, errors.Error(processName, "template not provided")
	}

	w, h := t.Width, t.Height
	bx, by := x-iDelX-JbAddedPixels, y-iDelY-JbAddedPixels

	common.Log.Trace("x: '%d', y: '%d', w: '%d', h: '%d', bx: '%d', by: '%d'", x, y, w, h, bx, by)
	box, err := bitmap.Rect(bx, by, w, h)
	if err != nil {
		return pt, errors.Wrap(err, processName, "")
	}
	d, _, err := s.ClipRectangle(box)
	if err != nil {
		common.Log.Error("Can't clip rectangle: %v", box)
		return pt, errors.Wrap(err, processName, "")
	}
	r := bitmap.New(d.Width, d.Height)
	minCount := math.MaxInt32
	var i, j, count, minX, minY int
	for i = -1; i <= 1; i++ {
		for j = -1; j <= 1; j++ {
			if _, err = bitmap.Copy(r, d); err != nil {
				return pt, errors.Wrap(err, processName, "")
			}
			if err = r.RasterOperation(j, i, w, h, bitmap.PixSrcXorDst, t, 0, 0); err != nil {
				return pt, errors.Wrap(err, processName, "")
			}
			count = r.CountPixels()
			if count < minCount {
				minX = j
				minY = i
				minCount = count
			}
		}
	}
	pt.X = minX
	pt.Y = minY
	return pt, nil
}

// Method is the encoding method used enum.
type Method int

// enum definitions of the encoding methods.
const (
	RankHaus Method = iota
	Correlation
)

// TwoByTwoWalk is the slice of values for classified encoding.
var TwoByTwoWalk = []int{
	0, 0,
	0, 1,
	-1, 0,
	0, -1,
	1, 0,
	-1, 1,
	1, 1,
	-1, -1,
	1, -1,
	0, -2,
	2, 0,
	0, 2,
	-2, 0,
	-1, -2,
	1, -2,
	2, -1,
	2, 1,
	1, 2,
	-1, 2,
	-2, 1,
	-2, -1,
	-2, -2,
	2, -2,
	2, 2,
	-2, 2,
}
