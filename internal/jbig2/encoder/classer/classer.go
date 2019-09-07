/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"errors"
	"image"

	"github.com/unidoc/unipdf/common"
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
	ComponentsNumber []int

	// SizeHaus is the size of square struct elem for haus.
	SizeHaus int
	// Rank val of haus match
	RankHaus float32
	// Thresh is the thresh value for the correlation score.
	Thresh float32
	// Corrects thresh vaue for heaver components; 0 for no correction.
	WeightFactor float32

	// Width * Height of each template without extra border pixels.
	NAArea []int

	// Width is max width of original src images.
	Width int
	// Height is max height of original src images.
	Height int
	// NClass is the current number of classes.
	NClass int
	// if 0 pixa isn't filled.
	KeepPixaa int
	// Instances for each class. Unbordered.
	Pixaa [][]*bitmap.Bitmap
	// Templates for each class. Bordered and not dilated.
	Pixat []*bitmap.Bitmap
	// Templates for each class. Bordered and dilated.
	Pixatd []*bitmap.Bitmap

	// Hash table to find templates by their size.
	TemplatesSize map[uint64][]float64
	// FgTemplates Fg areas of undilated templates. Used for rank < 1.0.
	FgTemplates []int

	// CentroidPoints centroids of all bordered cc.
	CentroidPoints []bitmap.Point
	// CentroidPointsTemplates centroids of all bordered template cc.
	CentroidPointsTemplates []bitmap.Point
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
	if method != RankHaus && method != Correlation {
		return nil, errors.New("jbig2 encoder invalid classer method")
	}

	switch components {
	case ConnComps, Characters, Words:
	default:
		return nil, errors.New("jbig2 encoder invalid classer component")
	}
	return &Classer{Method: method, Components: components}, nil
}

// AddPage adds the 'inputPage' to the classer 'c'.
func (c *Classer) AddPage(inputPage *bitmap.Bitmap) error {
	// TODO: jbclass.c:486
	c.Width = inputPage.Width
	c.Height = inputPage.Height
	c.AddPageComponents(inputPage, boundingBoxes, components)

	return nil
}

// AddPageComponents adds the components to the 'inputPage'.
func (c *Classer) AddPageComponents(inputPage *bitmap.Bitmap, boundingBoxes []image.Rectangle, components []*bitmap.Bitmap) error {
	// TODO: jbclass.c:531
	if boundingBoxes == nil || components == nil || len(boundingBoxes) == 0 {
		return nil
	}
	var err error
	switch c.Method {
	case RankHaus:
		err = c.ClassifyRankHaus(boundingBoxes, components)
	case Correlation:
		err = c.ClassifyCorrelation(boundingBoxes, components)
	default:
		common.Log.Debug("Unknown classify method: '%v'", c.Method)
		err = errors.New("unknown classify method")
	}
	if err != nil {
		return err
	}

	c.BaseIndex += len(boundingBoxes)
	c.ComponentsNumber = append(c.ComponentsNumber, len(boundingBoxes))

	return nil
}

// AddPages adds the pages to the given classer.
func (c *Classer) AddPages(
// TODO: SARRAY - jbclass.c:445
) error {
	return nil
}

// ClassifyRankHaus is the classification using windowed rank hausdorff metric.
func (c *Classer) ClassifyRankHaus(newCompontentsBB []image.Rectangle, newComponents []*bitmap.Bitmap) error {
	var n, nafg int
	if n = len(newComponents); n == 0 {
		return errors.New("empty new components")
	}

	for _, bm := range newComponents {
		nafg += bm.CountPixels()
	}

	size := c.SizeHaus
	sel := bitmap.SelCreateBrick(size, size, size/2, size/2, bitmap.SelHit)

	bms1 := make([]*bitmap.Bitmap, n)
	bms2 := make([]*bitmap.Bitmap, n)
	var (
		bm, bm1, bm2 *bitmap.Bitmap
		err          error
	)
	for i := 0; i < n; i++ {
		bm := bms1[i]

		bm1, err = bm.AddBorderGeneral(JbAddedPixels, JbAddedPixels, JbAddedPixels, JbAddedPixels, 0)
		if err != nil {
			return err
		}
		bm2, err = bitmap.Dilate(nil, bm1, sel)
		if err != nil {
			return err
		}
		bms1[n] = bm1 // un-dilated
		bms2[n] = bm2 // dilated
	}

	return nil
}

// ClassifyCorrelation is the classification using windowed correlation score.
func (c *Classer) ClassifyCorrelation(newCompontentsBB []image.Rectangle, newComponents []*bitmap.Bitmap) error {
	// TODO: jbclass.c:1031
	boxa, pixas := newCompontentsBB, newComponents
	if boxa == nil {
		return errors.New("ClassifyCorrelation - newComponents bounding boxes not found")
	}
	if pixas == nil {
		return errors.New("ClassifyCorrelation - newComponents bitmap array not found")
	}

	if len(pixas) == 0 {
		common.Log.Debug("ClassifyCorrelation - provided pixas is empty")
		return nil
	}

	var (
		bm, bm1 *bitmap.Bitmap
		err     error
	)
	pixa1 := make([]*bitmap.Bitmap, len(pixas))
	for i, bm := range pixas {
		bm1, err = bm.AddBorderGeneral(JbAddedPixels, JbAddedPixels, JbAddedPixels, JbAddedPixels, 0)
		if err != nil {
			return err
		}
		pixa1[i] = bm1
	}

	nafgt := c.FgTemplates
	sumtab := bitmap.MakePixelSumTab8()
	centtab := bitmap.MakePixelCentroidTab8()
	pixcts := make([]int, len(pixas))
	pixRowCts := make([][]int, len(pixas))
	pta := make([]bitmap.Point, len(pixas))

	var (
		xsum, ysum                    int
		downcount, rowIndex, rowCount int
		x, y                          int
		bt                            byte
	)
	for i, bm := range pixa1 {
		pixRowCts[i] = make([]int, bm.Height)

		xsum = 0
		ysum = 0

		rowIndex = (bm.Height - 1) * bm.RowStride
		downcount = 0
		increaser := func() {
			y--
			rowIndex -= bm.RowStride
		}

		for y = bm.Height - 1; y >= 0; increaser() {
			pixRowCts[i][y] = downcount
			rowCount = 0
			for x = 0; x < bm.RowStride; x++ {
				bt = bm.Data[rowIndex+x]
				rowCount += sumtab[bt]
				xsum += centtab[bt] + x*8*sumtab[bt]
			}
			downcount += rowCount
			ysum += rowCount * y
		}
		pixcts[i] = downcount
		if downcount > 0 {
			pta[i] = bitmap.Point{X: float32(xsum) / float32(downcount), Y: float32(ysum) / float32(downcount)}
		} else {
			// no pixels
			pta[i] = bitmap.Point{X: float32(bm.Width) / float32(2), Y: float32(bm.Height) / float32(2)}
		}
	}

	c.CentroidPoints = append(c.CentroidPoints, pta...)
	// GOTO: jbclass.c:1143
	var (
		area, area1, area2 int
		threshold          float32
		ct1, ct2           bitmap.Point
		found              bool
		findContext        *FindTemplatesState
		i                  int
		bm2                *bitmap.Bitmap
	)

	for i, bm1 = range pixa1 {
		area1 = pixcts[i]
		ct1 = pta[i]
		found = false
		nt := len(c.Pixat)
		findContext = c.FindSimilarSizedTemplatesInit(bm1)
		for iclass := findContext.Next(); iclass > -1; {
			// get the template
			bm2 = c.Pixat[iclass]
			area2 = nafgt[iclass]
			ct2 = c.CentroidPointsTemplates[iclass]
			if c.WeightFactor > 0.0 {
				area = c.NAArea[iclass]
				threshold = c.Thresh + (1.0-c.Thresh)*c.WeightFactor*float32(area2)/float32(area)
			} else {
				threshold = c.Thresh
			}

			overThreshold, err := bitmap.CorrelationScoreThresholded(bm1, bm2, area1, area2, ct1.X-ct2.X, ct1.Y-ct2.Y, MaxDiffWidth, MaxDiffHeight, sumtab, pixRowCts[i], threshold)
			if err != nil {
				return err
			}

			if overThreshold {
				found = true
				c.ClassIDs = append(c.ClassIDs, iclass)
				c.PageNumbers = append(c.PageNumbers, c.NPages)

				if c.KeepPixaa != 0 {
					bm = pixas[i]
					c.Pixaa[iclass] = append(c.Pixaa[iclass], bm)
					bm.Box = boxa[i]
				}
				break
			}
		}
		if !found {
			c.ClassIDs = append(c.ClassIDs, nt)
			c.PageNumbers = append(c.PageNumbers, c.NPages)
			pixa := []*bitmap.Bitmap{}
			pix := pixas[i]
			pixa = append(pixa, pix)

			wt, ht := pix.Width, pix.Height
			key := uint64(ht * wt)
			templatesSizes := c.TemplatesSize[key]
			templatesSizes = append(templatesSizes, float64(nt))
			c.TemplatesSize[key] = templatesSizes
			pix.Box = boxa[i]
			c.Pixaa = append(c.Pixaa, pixa)
			c.CentroidPointsTemplates = append(c.CentroidPoints, ct1)
			c.FgTemplates = append(c.FgTemplates, area1)
			area = (bm1.Width - 2*JbAddedPixels) * (bm1.Height - 2*JbAddedPixels)
			c.NAArea = append(c.NAArea, area)
		}
	}
	c.NClass = len(c.Pixat)
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
func (c *Classer) FindSimilarSizedTemplatesInit(toMatch *bitmap.Bitmap) *FindTemplatesState {
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
	// TODO: handle this return
	return nil, nil
}
