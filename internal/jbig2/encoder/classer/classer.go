/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"image"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

var (
	debugCorrelationScore bool
)

// Classer holds all the data accumulated during the classifcation
// process that can be used for a compressed jbig2-type representation
// of a set of images.
type Classer struct {
	// input image names - 'safiles'
	ImageNames []string
	Method     Method
	Components bitmap.Component

	// MaxWidth is max component width allowed.
	MaxWidth int
	// MaxHeight is max component height allowed.
	MaxHeight int
	// PagesProcessed is the number of pages already processed - 'npages'.
	PagesProcessed int
	// BaseIndex is number of components already processed on fully processed pages.
	BaseIndex int

	// Number of components on each page - 'nacomps'.
	ComponentsNumber *basic.NumSlice

	// SizeHaus is the size of square struct elem for haus.
	SizeHaus int
	// Rank val of haus match
	RankHaus float32
	// Thresh is the thresh value for the correlation score.
	Thresh float32
	// Corrects thresh vaue for heaver components; 0 for no correction.
	WeightFactor float32

	// Width * Height of each template without extra border pixels - 'naarea'.
	TemplateAreas *basic.NumSlice

	// Width is max width of original src images.
	Width int
	// Height is max height of original src images.
	Height int
	// NumberOfClasses is the current number of classes - 'nclass'.
	NumberOfClasses int
	// KeepClassInstances is a flag that defines if the class instances should be stored
	// in the 'ClassInstances' BitmapsArray.
	KeepClassInstances bool
	// ClassInstances for each class. Unbordered - 'pixaa'.
	ClassInstances *bitmap.BitmapsArray
	// UndilatedTemplates for each class. Bordered and not dilated - 'pixat'.
	UndilatedTemplates *bitmap.Bitmaps
	// DilatedTemplates for each class. Bordered and dilated - 'pixatd'.
	DilatedTemplates *bitmap.Bitmaps

	// Hash table to find templates by their size - 'dahash'.
	TemplatesSize map[uint64]int
	// FgTemplates - foreground areas of undilated templates. Used for rank < 1.0 - 'nafgt'.
	FgTemplates *basic.NumSlice

	// CentroidPoints centroids of all bordered cc.
	CentroidPoints *bitmap.Points
	// CentroidPointsTemplates centroids of all bordered template cc.
	CentroidPointsTemplates *bitmap.Points
	// ClassIDs is the slice of class ids for each component - 'naclass'.
	ClassIDs *basic.NumSlice
	// PageNumbers it the page nums slice for each component - 'napage'.
	PageNumbers *basic.NumSlice
	// PtaUL is the slice of UL corners at which the template
	// is to be placed for each component.
	PtaUL *bitmap.Points
	// PtaLL is the slice of LL corners at which the template
	// is to be placed for each component.
	PtaLL *bitmap.Points
}

// New creates new Classer instance for provided 'method' and given 'components'.
func New(method Method, components bitmap.Component) (*Classer, error) {
	const processName = "classer.New"
	if method != RankHaus && method != Correlation {
		return nil, errors.Error(processName, "invalid classer method")
	}

	switch components {
	case bitmap.ComponentConn, bitmap.ComponentCharacters, bitmap.ComponentWords:
	default:
		return nil, errors.Error(processName, "invalid classer component")
	}
	return &Classer{Method: method, Components: components}, nil
}

// AddPage adds the 'inputPage' to the classer 'c'.
func (c *Classer) AddPage(inputPage *bitmap.Bitmap) error {
	// TODO: jbclass.c:486
	c.Width = inputPage.Width
	c.Height = inputPage.Height

	return nil
}

// AddPageComponents adds the components to the 'inputPage'.
func (c *Classer) AddPageComponents(inputPage *bitmap.Bitmap, boxas *bitmap.Boxes, components *bitmap.Bitmaps) error {
	const processName = "Classer.AddPageComponents"
	// TODO: jbclass.c:531
	if inputPage == nil {
		return errors.Error(processName, "nil input page")
	}
	if boxas == nil || components == nil || len(*boxas) == 0 {
		c.PagesProcessed++
		return nil
	}
	var err error
	switch c.Method {
	case RankHaus:
		err = c.ClassifyRankHaus(boxas, components)
	case Correlation:
		err = c.ClassifyCorrelation(boxas, components)
	default:
		common.Log.Debug("Unknown classify method: '%v'", c.Method)
		return errors.Error(processName, "unknown classify method")
	}
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	// TODO: jbGetULCorners(classer, pixs, boxax)
	c.GetULCorners(inputPage, boxas)
	n := len(*boxas)
	c.BaseIndex += n
	c.ComponentsNumber.AddInt(n)
	c.PagesProcessed++
	return nil
}

// AddPages adds the pages to the given classer.
func (c *Classer) AddPages(
// TODO: SARRAY - jbclass.c:445
) error {
	return nil
}

// ClassifyRankHaus is the classification using windowed rank hausdorff metric.
func (c *Classer) ClassifyRankHaus(boxa *bitmap.Boxes, pixa *bitmap.Bitmaps) error {
	const processName = "Classer.ClassifyRankHaus"
	if boxa == nil {
		return errors.Error(processName, "boxa not defined")
	}
	if pixa == nil {
		return errors.Error(processName, "pixa not defined")
	}

	n := len(pixa.Values)
	if n == 0 {
		return errors.Error(processName, "empty new components")
	}
	nafg := pixa.CountPixels()

	size := c.SizeHaus
	sel := bitmap.SelCreateBrick(size, size, size/2, size/2, bitmap.SelHit)
	bms1 := &bitmap.Bitmaps{Values: make([]*bitmap.Bitmap, n)}
	bms2 := &bitmap.Bitmaps{Values: make([]*bitmap.Bitmap, n)}
	var (
		bm, bm1, bm2 *bitmap.Bitmap
		err          error
	)
	for i := 0; i < n; i++ {
		bm, err = pixa.GetBitmap(i)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}

		bm1, err = bm.AddBorderGeneral(JbAddedPixels, JbAddedPixels, JbAddedPixels, JbAddedPixels, 0)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}
		bm2, err = bitmap.Dilate(nil, bm1, sel)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}
		bms1.Values[n] = bm1 // un-dilated
		bms2.Values[n] = bm2 // dilated
	}
	pta, err := bitmap.Centroids(bms1.Values)
	if err != nil {
		return errors.Wrap(err, processName, "")
	}
	if err = pta.Add(c.CentroidPoints); err != nil {
		common.Log.Trace("No centroids to add")
	}

	if c.RankHaus == 1.0 {
		err = c.classifyRankHouseOne(boxa, pixa, bms1, bms2, pta)
	} else {
		err = c.classifyRankHouseNonOne(boxa, pixa, bms1, bms2, pta, nafg)
	}
	if err != nil {
		return errors.Wrap(err, processName, "")
	}
	return nil
}

func (c *Classer) classifyRankHouseOne(boxa *bitmap.Boxes, pixa, bms1, bms2 *bitmap.Bitmaps, pta *bitmap.Points) (err error) {
	const processName = "Classer.classifyRankHouseOne"
	var (
		x1, y1, x2, y2         float32
		iClass                 int
		bm, bm1, bm2, bm3, bm4 *bitmap.Bitmap
		found, testVal         bool
	)
	for i := 0; i < len(pixa.Values); i++ {
		bm1 = bms1.Values[i]
		bm2 = bms2.Values[i]
		x1, y1, err = pta.GetGeometry(i)
		if err != nil {
			return errors.Wrapf(err, processName, "first geometry")
		}

		nt := len(c.UndilatedTemplates.Values)

		found = false
		findContext := c.FindSimilarSizedTemplatesInit(bm1)
		for iClass = findContext.Next(); iClass > -1; {
			bm3, err = c.UndilatedTemplates.GetBitmap(iClass)
			if err != nil {
				return errors.Wrap(err, processName, "bm3")
			}
			bm4, err = c.DilatedTemplates.GetBitmap(iClass)
			if err != nil {
				return errors.Wrap(err, processName, "bm4")
			}

			x2, y2, err = c.CentroidPointsTemplates.GetGeometry(iClass)
			if err != nil {
				return errors.Wrap(err, processName, "CentroidTemplates")
			}
			testVal, err = bitmap.Haustest(bm1, bm2, bm3, bm4, x1-x2, y1-y2, MaxDiffWidth, MaxDiffHeight)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			if testVal {
				found = true
				c.ClassIDs.AddInt(iClass)
				c.PageNumbers.AddInt(c.PagesProcessed)

				if c.KeepClassInstances {
					bitmaps, err := c.ClassInstances.GetBitmaps(iClass)
					if err != nil {
						return errors.Wrap(err, processName, "KeepPixaa")
					}
					bm, err = pixa.GetBitmap(i)
					if err != nil {
						return errors.Wrap(err, processName, "KeepPixaa")
					}
					bitmaps.AddBitmap(bm)

					box, err := boxa.Get(i)
					if err != nil {
						return errors.Wrap(err, processName, "KeepPixaa")
					}
					bitmaps.AddBox(box)
				}
				break
			}

			if !found {
				c.ClassIDs.AddInt(nt)
				c.PageNumbers.AddInt(c.PagesProcessed)
				bitmaps := &bitmap.Bitmaps{}
				bm, err = pixa.GetBitmap(i)
				if err != nil {
					return errors.Wrap(err, processName, "!found")
				}
				bitmaps.Values = append(bitmaps.Values, bm)
				wt, ht := bm.Width, bm.Height
				c.TemplatesSize[uint64(ht)*uint64(wt)] = nt
				box, err := boxa.Get(i)
				if err != nil {
					return errors.Wrap(err, processName, "!found")
				}
				bitmaps.AddBox(box)
				c.ClassInstances.AddBitmaps(bitmaps)
				c.CentroidPointsTemplates.AddPoint(x1, y1)
				c.UndilatedTemplates.AddBitmap(bm1)
				c.DilatedTemplates.AddBitmap(bm2)
			}
		}
	}
	return nil
}

// classifyRankHouseNonOne is a helper that classifies when the rank < 1.0.
func (c *Classer) classifyRankHouseNonOne(boxa *bitmap.Boxes, pixa, bms1, bms2 *bitmap.Bitmaps, pta *bitmap.Points, nafg *basic.NumSlice) (err error) {
	const processName = "Classer.classifyRankHouseOne"
	var (
		x1, y1, x2, y2         float32
		area1, area3, iClass   int
		bm, bm1, bm2, bm3, bm4 *bitmap.Bitmap
		found, testVal         bool
	)
	tab8 := bitmap.MakePixelSumTab8()
	for i := 0; i < len(pixa.Values); i++ {
		if bm1, err = bms1.GetBitmap(i); err != nil {
			return errors.Wrap(err, processName, "bms1.Get(i)")
		}

		if area1, err = nafg.GetInt(i); err != nil {
			common.Log.Trace("Getting FGTemplates at: %d failed: %v", i, err)
		}

		if bm2, err = bms2.GetBitmap(i); err != nil {
			return errors.Wrap(err, processName, "bms2.Get(i)")
		}
		if x1, y1, err = pta.GetGeometry(i); err != nil {
			return errors.Wrapf(err, processName, "pta[i].Geometry")
		}

		nt := len(c.UndilatedTemplates.Values)
		found = false
		findContext := c.FindSimilarSizedTemplatesInit(bm1)
		for iClass = findContext.Next(); iClass > -1; {
			if bm3, err = c.UndilatedTemplates.GetBitmap(iClass); err != nil {
				return errors.Wrap(err, processName, "pixat.[iClass]")
			}
			if area3, err = c.FgTemplates.GetInt(iClass); err != nil {
				common.Log.Trace("Getting FGTemplate[%d] failed: %v", iClass, err)
			}

			if bm4, err = c.DilatedTemplates.GetBitmap(iClass); err != nil {
				return errors.Wrap(err, processName, "pixatd[iClass]")
			}
			if x2, y2, err = c.CentroidPointsTemplates.GetGeometry(iClass); err != nil {
				return errors.Wrap(err, processName, "CentroidPointsTemplates[iClass]")
			}
			testVal, err = bitmap.RankHausTest(bm1, bm2, bm3, bm4, x1-x2, y1-y2, MaxDiffWidth, MaxDiffHeight, area1, area3, c.RankHaus, tab8)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			if testVal {
				found = true
				c.ClassIDs.AddInt(iClass)
				c.PageNumbers.AddInt(c.PagesProcessed)
				if c.KeepClassInstances {
					bitmaps, err := c.ClassInstances.GetBitmaps(iClass)
					if err != nil {
						return errors.Wrap(err, processName, "c.Pixaa.GetBitmaps(iClass)")
					}
					if bm, err = pixa.GetBitmap(i); err != nil {
						return errors.Wrap(err, processName, "pixa[i]")
					}
					bitmaps.Values = append(bitmaps.Values, bm)
					box, err := boxa.Get(i)
					if err != nil {
						return errors.Wrap(err, processName, "boxa.Get(i)")
					}
					bitmaps.Boxes = append(bitmaps.Boxes, box)
				}
				break
			}
		}
		if !found {
			c.ClassIDs.AddInt(nt)
			c.PageNumbers.AddInt(c.PagesProcessed)

			bitmaps := &bitmap.Bitmaps{}
			bm = pixa.Values[i]
			bitmaps.AddBitmap(bm)

			wt, ht := bm.Width, bm.Height
			c.TemplatesSize[uint64(wt)*uint64(ht)] = nt
			box, err := boxa.Get(i)
			if err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			bitmaps.AddBox(box)
			c.ClassInstances.AddBitmaps(bitmaps)
			c.CentroidPointsTemplates.AddPoint(x1, y1)
			c.UndilatedTemplates.AddBitmap(bm1)
			c.DilatedTemplates.AddBitmap(bm2)
			c.FgTemplates.AddInt(area1)
		}
	}
	c.NumberOfClasses = len(c.UndilatedTemplates.Values)
	return nil
}

// ClassifyCorrelation is the classification using windowed correlation score.
func (c *Classer) ClassifyCorrelation(boxa *bitmap.Boxes, pixas *bitmap.Bitmaps) error {
	// TODO: jbclass.c:1031
	const processName = "Classer.ClassifyCorrelation"
	if boxa == nil {
		return errors.Error(processName, "newComponents bounding boxes not found")
	}
	if pixas == nil {
		return errors.Error(processName, "newComponents bitmap array not found")
	}

	n := len(pixas.Values)
	if n == 0 {
		common.Log.Debug("ClassifyCorrelation - provided pixas is empty")
		return nil
	}

	var (
		bm, bm1 *bitmap.Bitmap
		err     error
	)
	bms1 := &bitmap.Bitmaps{Values: make([]*bitmap.Bitmap, n)}

	for i, bm := range pixas.Values {
		bm1, err = bm.AddBorderGeneral(JbAddedPixels, JbAddedPixels, JbAddedPixels, JbAddedPixels, 0)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}
		bms1.Values[i] = bm1
	}

	nafgt := c.FgTemplates
	sumtab := bitmap.MakePixelSumTab8()
	centtab := bitmap.MakePixelCentroidTab8()

	pixcts := make([]int, n)
	pixRowCts := make([][]int, n)
	pts := bitmap.Points(make([]bitmap.Point, n))
	pta := &pts

	var (
		xsum, ysum                    int
		downcount, rowIndex, rowCount int
		x, y                          int
		bt                            byte
	)

	for i, bm := range bms1.Values {
		pixRowCts[i] = make([]int, bm.Height)

		xsum = 0
		ysum = 0

		rowIndex = (bm.Height - 1) * bm.RowStride
		downcount = 0

		for y = bm.Height - 1; y >= 0; y, rowIndex = y-1, rowIndex-bm.RowStride {
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
			(*pta)[i] = bitmap.Point{X: float32(xsum) / float32(downcount), Y: float32(ysum) / float32(downcount)}
		} else {
			// no pixels
			(*pta)[i] = bitmap.Point{X: float32(bm.Width) / float32(2), Y: float32(bm.Height) / float32(2)}
		}
	}
	c.CentroidPoints.Add(pta)
	// GOTO: jbclass.c:1143
	var (
		area, area1, area2 int
		threshold          float32
		x1, y1, x2, y2     float32
		ct1, ct2           bitmap.Point
		found              bool
		findContext        *FindTemplatesState
		i                  int
		bm2                *bitmap.Bitmap
		box                *image.Rectangle
		bitmaps            *bitmap.Bitmaps
	)

	for i, bm1 = range bms1.Values {
		area1 = pixcts[i]
		if x1, y1, err = pta.GetGeometry(i); err != nil {
			return errors.Wrap(err, processName, "pta - i")
		}

		found = false
		nt := len(c.UndilatedTemplates.Values)
		findContext = c.FindSimilarSizedTemplatesInit(bm1)
		for iclass := findContext.Next(); iclass > -1; {
			// get the template
			if bm2, err = c.UndilatedTemplates.GetBitmap(iclass); err != nil {
				return errors.Wrap(err, processName, "unidlated[iclass] = bm2")
			}

			if area2, err = nafgt.GetInt(iclass); err != nil {
				common.Log.Trace("FG Template [iclass] failed: %v", err)
			}

			if x2, y2, err = c.CentroidPointsTemplates.GetGeometry(iclass); err != nil {
				return errors.Wrap(err, processName, "CentroidPointTemplates[iclass] = x2,y2 ")
			}
			if c.WeightFactor > 0.0 {
				if area, err = c.TemplateAreas.GetInt(iclass); err != nil {
					common.Log.Trace("TemplateAreas[iclass] = area %v", err)
				}
				threshold = c.Thresh + (1.0-c.Thresh)*c.WeightFactor*float32(area2)/float32(area)
			} else {
				threshold = c.Thresh
			}

			overThreshold, err := bitmap.CorrelationScoreThresholded(bm1, bm2, area1, area2, ct1.X-ct2.X, ct1.Y-ct2.Y, MaxDiffWidth, MaxDiffHeight, sumtab, pixRowCts[i], threshold)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			if debugCorrelationScore {
				// debug the correlation score values
				var (
					score, testScore float64
					count, testCount int
				)
				score, err = bitmap.CorrelationScore(bm1, bm2, area1, area2, x1-x2, y1-y2, MaxDiffWidth, MaxDiffHeight, sumtab)
				if err != nil {
					return errors.Wrap(err, processName, "debugCorrelationScore")
				}

				testScore, err = bitmap.CorrelationScoreSimple(bm1, bm2, area1, area2, x1-x2, y1-y2, MaxDiffWidth, MaxDiffHeight, sumtab)
				if err != nil {
					return errors.Wrap(err, processName, "debugCorrelationScore")
				}
				count = int(math.Sqrt(float64(score) * float64(area1) * float64(area2)))
				testCount = int(math.Sqrt(float64(testScore) * float64(area1) * float64(area2)))
				if (score >= float64(threshold)) != (testScore >= float64(threshold)) {
					return errors.Errorf(processName, "debug Correlation score mismatch - %d(%0.4f, %v) vs %d(%0.4f, %v) %0.4f", count, score, score >= float64(threshold), testCount, testScore, testScore >= float64(threshold), score-testScore)
				}

				if score >= float64(threshold) != overThreshold {
					return errors.Errorf(processName, "debug Correlation score Mismatch between correlation / threshold. Comparison: %0.4f(%0.4f, %d) >= %0.4f(%0.4f) vs %b",
						score, score*float64(area1)*float64(area2), count, threshold, threshold*float32(area1)*float32(area2), overThreshold)
				}
			}

			if overThreshold {
				found = true
				c.ClassIDs.AddInt(iclass)
				c.PageNumbers.AddInt(c.PagesProcessed)

				if c.KeepClassInstances {
					if bm, err = pixas.GetBitmap(i); err != nil {
						return errors.Wrap(err, processName, "KeepClassInstances - i")
					}
					if bitmaps, err = c.ClassInstances.GetBitmaps(iclass); err != nil {
						return errors.Wrap(err, processName, "KeepClassInstances - iclass")
					}
					bitmaps.AddBitmap(bm)
					if box, err = boxa.Get(i); err != nil {
						return errors.Wrap(err, processName, "KeepClassInstances")
					}
					bitmaps.AddBox(box)
				}
				break
			}
		}
		if !found {
			c.ClassIDs.AddInt(nt)
			c.PageNumbers.AddInt(c.PagesProcessed)

			bitmaps = &bitmap.Bitmaps{}
			if bm, err = pixas.GetBitmap(i); err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			bitmaps.AddBitmap(bm)

			wt, ht := bm.Width, bm.Height
			key := uint64(ht) * uint64(wt)
			c.TemplatesSize[key] = nt
			if box, err = boxa.Get(i); err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			bitmaps.AddBox(box)
			c.ClassInstances.AddBitmaps(bitmaps)
			c.CentroidPointsTemplates.AddPoint(x1, y1)
			c.FgTemplates.AddInt(area1)

			area = (bm1.Width - 2*JbAddedPixels) * (bm1.Height - 2*JbAddedPixels)
			c.TemplateAreas.AddInt(area)
		}
	}
	c.NumberOfClasses = len(c.UndilatedTemplates.Values)
	return nil
}

// DataSave ...
func (c *Classer) DataSave() *Data {
	// TODO: jbclass.c:1879
	return nil
}

// GetLLCorners get the ll corners.
func (c *Classer) GetLLCorners() (err error) {
	// TODO: jbclass.c:2319
	const processName = "Classer.GetLLCorners"
	if c.PtaUL == nil {
		return errors.Error(processName, "UL Corners not defined")
	}
	n := len(*c.PtaUL)

	c.PtaLL = &bitmap.Points{}
	*c.PtaLL = bitmap.Points(make([]bitmap.Point, n))

	var (
		x1, y1    float32
		iClass, h int
		bm        *bitmap.Bitmap
	)
	for i := 0; i < n; i++ {
		x1, y1, err = c.PtaUL.GetGeometry(i)
		if err != nil {
			// NOTE: check if the error is required.
			common.Log.Debug("Getting PtaUL failed: %v", err)
			return errors.Wrap(err, processName, "PtaUL Geometry")
		}
		iClass, err = c.ClassIDs.GetInt(i)
		if err != nil {
			common.Log.Debug("Getting ClassID failed: %v", err)
			return errors.Wrap(err, processName, "ClassID")
		}
		bm, err = c.UndilatedTemplates.GetBitmap(iClass)
		if err != nil {
			common.Log.Debug("Getting UndilatedTemplates failed: %v", err)
			return errors.Wrap(err, processName, "Undilated Templates")
		}
		h = bm.Height
		// Add the global LL corner point.
		c.PtaLL.AddPoint(x1, y1+float32(h)-1-2*float32(JbAddedPixels))
	}

	return nil
}

// GetULCorners get the ul corners.
func (c *Classer) GetULCorners(s *bitmap.Bitmap, boxa *bitmap.Boxes) error {
	// TODO: jbclass.c:2225
	const processName = "Classer.GetULCorners"
	if s == nil {
		return errors.Error(processName, "nil image bitmap")
	}
	if boxa == nil {
		return errors.Error(processName, "nil bounds")
	}

	n := len(*boxa)
	sumTab := bitmap.MakePixelSumTab8()
	var (
		index, iClass, idelX, idelY int
		x1, y1, x2, y2              float32
		err                         error
		box                         *image.Rectangle
		t                           *bitmap.Bitmap
		pt                          image.Point
	)
	for i := 0; i < n; i++ {
		index = c.BaseIndex + i
		if x1, y1, err = c.CentroidPoints.GetGeometry(i); err != nil {
			return errors.Wrap(err, processName, "CentroidPoints")
		}
		if iClass, err = c.ClassIDs.GetInt(index); err != nil {
			return errors.Wrap(err, processName, "")
		}
		if x2, y2, err = c.CentroidPointsTemplates.GetGeometry(iClass); err != nil {
			return errors.Wrap(err, processName, "CentroidPointsTemplates")
		}
		delX := x2 - x1
		delY := y2 - y1
		if delX >= 0 {
			idelX = int(delX + 0.5)
		} else {
			idelX = int(delX - 0.5)
		}
		if delY >= 0 {
			idelY = int(delY + 0.5)
		} else {
			idelY = int(delY - 0.5)
		}
		if box, err = boxa.Get(i); err != nil {
			return errors.Wrap(err, processName, "")
		}
		x, y := box.Min.X, box.Min.Y

		// finalPositionForAligment()
		t, err = c.UndilatedTemplates.GetBitmap(iClass)
		if err != nil {
			return errors.Wrap(err, processName, "UndilatedTemplates.Get(iClass)")
		}

		pt, err = finalAlignmentPositioning(s, x, y, idelX, idelY, t, sumTab)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}

		c.PtaUL.AddPoint(float32(x-idelX+pt.X), float32(y-idelY+pt.Y))

	}
	return nil
}

// FindSimilarSizedTemplatesInit initializes the templatesState context.
func (c *Classer) FindSimilarSizedTemplatesInit(bms *bitmap.Bitmap) *FindTemplatesState {
	return &FindTemplatesState{
		Width:   bms.Width,
		Height:  bms.Height,
		Classer: c,
	}
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
