/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"errors"
	"image"

	"github.com/unidoc/unipdf/internal/jbig2/bitmap"
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
// [in] 	'classes' 	one pixa for each class.
// [out]	'samples' 	number of samples used to build each composite.
// [out]	'centroids'	centroids of bordered composites.
func AccumulateComposites(classes *bitmap.Pixaa, samples *[]float64, centroids *[]image.Point) (*bitmap.Pixa, error) {
	// TODO: jbclass.c:1656
	return nil, nil
}

// CorrelationInit is the initialization function
// used for unsupervised classification of the collections
// of connected components. Uses correlation classification with components.
func CorrelationInit(components Component, maxWidth, maxHeight int, thresh, weightFactor float32) (*Classer, error) {
	return correlationInitInternal(components, maxWidth, maxHeight, thresh, weightFactor, 1)
}

// CorrelationInitWithoutComponents is the initialization function
// used for unsupervised classification of the collections
// of connected components. Uses correlation classification without components.
func CorrelationInitWithoutComponents(components Component, maxWidth, maxHeight int, thresh, weightFactor float32) (*Classer, error) {
	return correlationInitInternal(components, maxWidth, maxHeight, thresh, weightFactor, 0)
}

// FinalAlignmentPositioning ...
func FinalAlignmentPositioning(
	inputPageImage *bitmap.Pix,
	x, y, iDelX, iDelY int,
	template *bitmap.Pix,
	sumtab []int,
	pdX, pdY []int,
) error {
	// TODO: 2519
	return nil
}

// GetComponents determines the image components.
// The 'img' image argument must be of 1bpp format.
func GetComponents(
	img *bitmap.Pix,
	components Component,
	maxWidth, maxHeight int,
	boundingBoxes []*bitmap.Boxa,
	componentItems []*bitmap.Pixa,
) error {
	// TODO: jbclass.c:1313
	return nil
}

// PixHaustest ...
func PixHaustest(p1, p2, p3, p4 *bitmap.Pix, delX, delY float32, maxDiffW, maxDiffH int) int {
	// TODO: jbclass.c:846
	return 0
}

// PixRankHausTest ...
func PixRankHausTest(
	p1, p2, p3, p4 *bitmap.Pix,
	delX, delY float32,
	maxDiffW, maxDiffH, area1, area3 int,
	rank float32,
	// tab8 []int,
) int {
	// TODO: jbclass.c:944
	return 0
}

// PixWordMaskByDilation ...
func PixWordMaskByDilation(
	pixS *bitmap.Pix, // 1bpp
	// TODO: determine output arguments
) error {
	// TODO: jbclass.c:1455
	return nil
}

// PixWordBoxesByDilation ...
func PixWordBoxesByDilation(
	pixs *bitmap.Bitmap,
	minWidth, minHeight, maxWidth, maxHeight int,
	// TODO: determine output arguments
	// pBoxa []*bitmap.Boxa, // [required]
	// pSize *int,
	// pixAdb *bitmap.Pixa,
) error {
	// TODO: jbclass.c:1594
	return nil
}

// TemplatesFromComposites ...
// returns 8 bpp templates for each class, or NULL on error
func TemplatesFromComposites(classCopmosits *bitmap.Pixa, samplesNumber []float64) (*bitmap.Pixa, error) {
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

// Component is the jbig2 classification components enums.
type Component int

// Enum definitions of the Components.
const (
	ConnComps Component = iota
	Characters
	Words
)

// ObjectAccess is the enum flag that determines if the related
// access to the object.
type ObjectAccess int

// ObjectAccess enum constants
const (
	OANoCopy ObjectAccess = iota
	OACopy
	OAClone
	OACopyClone
)

const (
	// OAInsert one of the ObjectAccess enums.
	OAInsert ObjectAccess = OANoCopy
)

func correlationInitInternal(components Component, maxWidth, maxHeight int, thresh, weightFactor float32, keepComponents int) (*Classer, error) {
	if components > Words || components < 0 {
		return nil, errors.New("invalid jbig2 component")
	}

	if thresh < 0.4 || thresh > 0.98 {
		return nil, errors.New("jbig2 encoder thresh not in range [0.4 - 0.98]")
	}

	if weightFactor < 0.0 || weightFactor > 1.0 {
		return nil, errors.New("jbig2 encoder weight factor not in range [0.0 - 1.0]")
	}

	if maxWidth == 0 {
		switch components {
		case ConnComps:
			maxWidth = MaxConnCompWidth
		case Characters:
			maxWidth = MaxCharCompWidth
		default:
			maxWidth = MaxWordCompWidth
		}
	}
	if maxHeight == 0 {
		maxHeight = MaxCompHeight
	}

	classer, err := New(Correlation, components)
	if err != nil {
		return nil, err
	}
	classer.MaxWidth = maxWidth
	classer.MaxHeight = maxHeight
	classer.Thresh = thresh
	classer.WeightFactor = weightFactor
	// TODO: setDnaHash (double hash table) ? map[int][]float64
	classer.KeepPixaa = keepComponents

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
