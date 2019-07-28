/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"image"

	"github.com/unidoc/unipdf/internal/jbig2/bitmap"
)

// Classer holds all the data accumulated during the classifcation
// process that can be used for a compressed jbig2-type representation
// of a set of images.
type Classer struct {
	// input image names - 'safiles'
	ImageNames []string
	Method     Method
	Components Component

	// MaxWidth is max component width allowed.
	MaxWidth int
	// MaxHeight is max component height allowed.
	MaxHeight int
	// NPages is the number of pages already processed.
	NPages int
	// BaseIndex is number of components already processed on fully processed pages.
	BaseIndex int

	// Number of components on each page.
	NAComponents [][]int

	// SizeHaus is the size of square struct elem for haus.
	SizeHaus int
	// Rank val of haus match
	RankHaus float32
	// Thresh is the thresh value for the correlation score.
	Thresh float32
	// Corrects thresh vaue for heaver components; 0 for no correction.
	WeightFactor float32

	// Width * Height of each template without extra border pixels.
	NAArea [][]int

	// Width is max width of original src images.
	Width int
	// Height is max height of original src images.
	Height int
	// NClass is the current number of classes.
	NClass int
	// if 0 pixa isn't filled.
	KeepPixaa int
	// Instances for each class. Unbordered.
	Pixaa *bitmap.Pixaa
	// Templates for each class. Bordered and not dilated.
	Pixat *bitmap.Pixa
	// Templates for each class. Bordered and dilated.
	Pixatd *bitmap.Pixa

	// Hash table to find templates by their size.
	TemplatesSize map[uint64][]float64
	// FgTemplates Fg areas of undilated templates. Used for rank < 1.0.
	FgTemplates []int

	// Ptac centroids of all bordered cc.
	Ptac []image.Point
	// Ptact centroids of all bordered template cc.
	Ptact []image.Point
	// ClassIDs is the slice of class ids for each component.
	ClassIDs []int
	// PageNumbers it the page nums slice for each component.
	PageNumbers []int
	// PtaUL is the slice of UL corners at which the template
	// is to be placed for each component.
	PtaUL []image.Point
	// PtaLL is the slice of LL corners at which the template
	// is to be placed for each component.
	PtaLL []image.Point
}

// New creates new Classer instance for provided 'method' and given 'components'.
func New(method Method, components Component) (*Classer, error) {
	// TODO: jbclass.c:1791
	return &Classer{}, nil
}

// AddPage adds the 'inputPage' to the classer 'c'.
func (c *Classer) AddPage(inputPage *bitmap.Pix) error {
	// TODO: jbclass.c:486
	return nil
}

// AddPageComponents adds the components to the 'inputPage'.
func (c *Classer) AddPageComponents(inputPage *bitmap.Pix, boundingBoxes *bitmap.Boxa, components *bitmap.Pixa) error {
	// TODO: jbclass.c:531
	return nil
}

// AddPages adds the pages to the given classer.
func (c *Classer) AddPages(
// TODO: SARRAY - jbclass.c:445
) error {
	return nil
}

// ClassifyRankHaus is the classification using windowed rank hausdorff metric.
func (c *Classer) ClassifyRankHaus(newCompontentsBB *bitmap.Boxa, newComponentsPix *bitmap.Pixa) error {
	return nil
}

// ClassifyCorrelation is the classification using windowed correlation score.
func (c *Classer) ClassifyCorrelation(newCompontentsBB *bitmap.Boxa, newComponentsPix *bitmap.Pixa) error {
	// TODO: jbclass.c:1031
	return nil
}

// DataSave ...
func (c *Classer) DataSave() *Data {
	// TODO: jbclass.c:1879
	return nil
}

// GetLLCorners get the ll corners.
func (c *Classer) GetLLCorners() error {
	// TODO: jbclass.c:2225
	return nil
}

// GetULCorners get the ul corners.
func (c *Classer) GetULCorners(img *bitmap.Pix, bounds *bitmap.Boxa) error {
	// TODO: jbclass.c:2225
	return nil
}

// FindSimilarSizedTemplatesInit ...
func (c *Classer) FindSimilarSizedTemplatesInit(toMatch *bitmap.Pix) *FindTemplatesState {
	// TODO: jbclass.c:2401
	return nil
}

// FindTemplatesState stores the state of a state machine which fetches similar sized templates.
type FindTemplatesState struct {
	Classer *Classer
	// Desired width
	Width int
	// Desired height
	Height int
	// Index into two_by_two step array
	Index int
	// Current number array
	CurrentNumbers []float64
	// Current element of the array
	N int
}

// Next ...
func (f *FindTemplatesState) Next() int {
	// TODO: jbclass.c:2452
	return 0
}

// Dna is the double number array structure.
type Dna struct {
	Array    []float64
	RefCount int
}

// Clone returns pointer to the same 'Dna' or nil.
func (d *Dna) Clone() *Dna {
	if d == nil {
		return nil
	}
	d.RefCount++
	return d
}

// DnaHash is the hash map of the float64 arrays (Dna's).
type DnaHash map[uint64]*Dna

// GetDna gets the Dna at given 'key' and with given 'flag' ObjectAccess.
func (d DnaHash) GetDna(key uint64, flag ObjectAccess) (*Dna, error) {
	dna, ok := d[key]
	if !ok {
		return nil, nil
	}
	switch flag {
	case OANoCopy:
		return dna, nil
	case OACopy:
	case OAClone:
	default:
		return dna, nil
	}
}
