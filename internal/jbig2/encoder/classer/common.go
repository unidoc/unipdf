/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"image"
	"math"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

const (
	// JbAddedPixels is the size of the border added around pix of each c.c. for further processing.
	JbAddedPixels = 6
)

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

// AccumulateComposites ...
// [in] 	'classes' 	one slice of bitmaps for each class.
// [out]	'samples' 	number of samples used to build each composite.
// [out]	'centroids'	centroids of bordered composites.
func AccumulateComposites(classes [][]*bitmap.Bitmap, samples *[]float64, centroids *[]bitmap.Point) ([]*bitmap.Bitmap, error) {
	// TODO: jbclass.c:1656
	return nil, nil
}

// CorrelationInit is the initialization function
// used for unsupervised classification of the collections
// of connected components. Uses correlation classification with components.
func CorrelationInit(components bitmap.Component, maxWidth, maxHeight int, thresh, weightFactor float32) (*Classer, error) {
	return correlationInitInternal(components, maxWidth, maxHeight, thresh, weightFactor, 1)
}

// CorrelationInitWithoutComponents is the initialization function
// used for unsupervised classification of the collections
// of connected components. Uses correlation classification without components.
func CorrelationInitWithoutComponents(components bitmap.Component, maxWidth, maxHeight int, thresh, weightFactor float32) (*Classer, error) {
	return correlationInitInternal(components, maxWidth, maxHeight, thresh, weightFactor, 0)
}

// finalAligmentPositioning gets the best match position for the provided arguments.
// NOTE: jbclass.c:2519
func finalAlignmentPositioning(s *bitmap.Bitmap, x, y, iDelX, iDelY int, t *bitmap.Bitmap, sumtab []int) (pt image.Point, err error) {
	const processName = "finalAligmentPositioning"
	if s == nil {
		return pt, errors.Error(processName, "source not provided")
	}
	if t == nil {
		return pt, errors.Error(processName, "template not provided")
	}

	w, h := t.Width, t.Height
	bx, by := x-iDelX-JbAddedPixels, y-iDelY-JbAddedPixels
	box := image.Rect(bx, by, bx+w, by+h)
	d, _, err := s.ClipRectangle(&box)
	if err != nil {
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

// TemplatesFromComposites ...
// returns 8 bpp templates for each class, or NULL on error
func TemplatesFromComposites(classCopmosits []*bitmap.Bitmap, samplesNumber []float64) ([]*bitmap.Bitmap, error) {
	// TODO: jbclass.c:1746
	return nil, nil
}

// Method is the encoding method used enum.
type Method int

// Enum definitions of the encoding methods.
const (
	RankHaus Method = iota
	Correlation
)

func correlationInitInternal(components bitmap.Component, maxWidth, maxHeight int, thresh, weightFactor float32, keepComponents int) (*Classer, error) {
	const processName = "correlationInitInternal"
	if components > bitmap.ComponentWords || components < 0 {
		return nil, errors.Error(processName, "invalid jbig2 component")
	}
	if thresh < 0.4 || thresh > 0.98 {
		return nil, errors.Error(processName, "jbig2 encoder thresh not in range [0.4 - 0.98]")
	}
	if weightFactor < 0.0 || weightFactor > 1.0 {
		return nil, errors.Error(processName, "jbig2 encoder weight factor not in range [0.0 - 1.0]")
	}
	// if max width is not defined get the value from the constants.
	if maxWidth == 0 {
		switch components {
		case bitmap.ComponentConn:
			maxWidth = MaxConnCompWidth
		case bitmap.ComponentCharacters:
			maxWidth = MaxCharCompWidth
		case bitmap.ComponentWords:
			maxWidth = MaxWordCompWidth
		default:
			return nil, errors.Errorf(processName, "invalid components provided: %v", components)
		}
	}
	// if max height is not defined take the 'MaxCompHeight' value.
	if maxHeight == 0 {
		maxHeight = MaxCompHeight
	}

	classer, err := New(Correlation, components)
	if err != nil {
		return nil, errors.Wrap(err, processName, "can't create classer")
	}
	classer.MaxWidth = maxWidth
	classer.MaxHeight = maxHeight
	classer.Thresh = thresh
	classer.WeightFactor = weightFactor
	classer.TemplatesSize = map[uint64]int{}
	classer.KeepClassInstances = keepComponents != 0
	return classer, nil
}

// TwoByTwoWalk ...
// TODO: jbclass.c:2364
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
