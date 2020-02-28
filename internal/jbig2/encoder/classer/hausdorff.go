/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// classifyRankHaus is the classification using windowed rank hausdorff metric.
func (c *Classer) classifyRankHaus(boxa *bitmap.Boxes, pixa *bitmap.Bitmaps, pageNumber int) error {
	const processName = "classifyRankHaus"
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

	size := c.Settings.SizeHaus
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

	if c.Settings.RankHaus == 1.0 {
		err = c.classifyRankHouseOne(boxa, pixa, bms1, bms2, pta, pageNumber)
	} else {
		err = c.classifyRankHouseNonOne(boxa, pixa, bms1, bms2, pta, nafg, pageNumber)
	}
	if err != nil {
		return errors.Wrap(err, processName, "")
	}
	return nil
}

func (c *Classer) classifyRankHouseOne(boxa *bitmap.Boxes, pixa, bms1, bms2 *bitmap.Bitmaps, pta *bitmap.Points, pageNumber int) (err error) {
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
		findContext := initSimilarTemplatesFinder(c, bm1)
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
			testVal, err = bitmap.HausTest(bm1, bm2, bm3, bm4, x1-x2, y1-y2, MaxDiffWidth, MaxDiffHeight)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			if testVal {
				found = true
				// add the class index in the slice
				if err = c.ClassIDs.Add(iClass); err != nil {
					return errors.Wrap(err, processName, "")
				}
				// add the page number for the class index.
				if err = c.ComponentPageNumbers.Add(pageNumber); err != nil {
					return errors.Wrap(err, processName, "")
				}

				if c.Settings.KeepClassInstances {
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
		}

		if !found {
			if err = c.ClassIDs.Add(nt); err != nil {
				return errors.Wrap(err, processName, "")
			}
			if err = c.ComponentPageNumbers.Add(pageNumber); err != nil {
				return errors.Wrap(err, processName, "")
			}
			bitmaps := &bitmap.Bitmaps{}
			bm, err = pixa.GetBitmap(i)
			if err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			bitmaps.Values = append(bitmaps.Values, bm)
			wt, ht := bm.Width, bm.Height
			c.TemplatesSize.Add(uint64(ht)*uint64(wt), nt)
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
	return nil
}

// classifyRankHouseNonOne is a helper that classifies when the rank < 1.0.
func (c *Classer) classifyRankHouseNonOne(boxa *bitmap.Boxes, pixa, bms1, bms2 *bitmap.Bitmaps, pta *bitmap.Points, nafg *basic.NumSlice, pageNumber int) (err error) {
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
		findContext := initSimilarTemplatesFinder(c, bm1)
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
			testVal, err = bitmap.RankHausTest(bm1, bm2, bm3, bm4, x1-x2, y1-y2, MaxDiffWidth, MaxDiffHeight, area1, area3, float32(c.Settings.RankHaus), tab8)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			if testVal {
				found = true
				if err = c.ClassIDs.Add(iClass); err != nil {
					return errors.Wrap(err, processName, "")
				}
				if err = c.ComponentPageNumbers.Add(pageNumber); err != nil {
					return errors.Wrap(err, processName, "")
				}
				if c.Settings.KeepClassInstances {
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
			if err = c.ClassIDs.Add(nt); err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			if err = c.ComponentPageNumbers.Add(pageNumber); err != nil {
				return errors.Wrap(err, processName, "!found")
			}

			bitmaps := &bitmap.Bitmaps{}
			bm = pixa.Values[i]
			bitmaps.AddBitmap(bm)

			wt, ht := bm.Width, bm.Height
			c.TemplatesSize.Add(uint64(wt)*uint64(ht), nt)
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
