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

var debugCorrelationScore bool

// ClassifyCorrelation is the classification using windowed correlation score.
func (c *Classer) classifyCorrelation(boxa *bitmap.Boxes, pixas *bitmap.Bitmaps, pageNumber int) error {
	const processName = "classifyCorrelation"
	if boxa == nil {
		return errors.Error(processName, "newComponents bounding boxes not found")
	}
	if pixas == nil {
		return errors.Error(processName, "newComponents bitmap array not found")
	}

	n := len(pixas.Values)
	if n == 0 {
		common.Log.Debug("classifyCorrelation - provided pixas is empty")
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
	if err = c.CentroidPoints.Add(pta); err != nil {
		return errors.Wrap(err, processName, "centroid add")
	}

	var (
		area, area1, area2 int
		threshold          float64
		x1, y1, x2, y2     float32
		ct1, ct2           bitmap.Point
		found              bool
		findContext        *similarTemplatesFinder
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
		findContext = initSimilarTemplatesFinder(c, bm1)
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
			if c.Settings.WeightFactor > 0.0 {
				if area, err = c.TemplateAreas.Get(iclass); err != nil {
					common.Log.Trace("TemplateAreas[iclass] = area %v", err)
				}
				threshold = c.Settings.Thresh + (1.0-c.Settings.Thresh)*c.Settings.WeightFactor*float64(area2)/float64(area)
			} else {
				threshold = c.Settings.Thresh
			}

			overThreshold, err := bitmap.CorrelationScoreThresholded(bm1, bm2, area1, area2, ct1.X-ct2.X, ct1.Y-ct2.Y, MaxDiffWidth, MaxDiffHeight, sumtab, pixRowCts[i], float32(threshold))
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
				count = int(math.Sqrt(score * float64(area1) * float64(area2)))
				testCount = int(math.Sqrt(testScore * float64(area1) * float64(area2)))
				if (score >= threshold) != (testScore >= threshold) {
					return errors.Errorf(processName, "debug Correlation score mismatch - %d(%0.4f, %v) vs %d(%0.4f, %v) %0.4f", count, score, score >= float64(threshold), testCount, testScore, testScore >= float64(threshold), score-testScore)
				}

				if score >= threshold != overThreshold {
					return errors.Errorf(processName, "debug Correlation score Mismatch between correlation / threshold. Comparison: %0.4f(%0.4f, %d) >= %0.4f(%0.4f) vs %v",
						score, score*float64(area1)*float64(area2), count, threshold, float32(threshold)*float32(area1)*float32(area2), overThreshold)
				}
			}

			if overThreshold {
				found = true
				if err = c.ClassIDs.Add(iclass); err != nil {
					return errors.Wrap(err, processName, "overThreshold")
				}
				if err = c.ComponentPageNumbers.Add(pageNumber); err != nil {
					return errors.Wrap(err, processName, "overThreshold")
				}

				if c.Settings.KeepClassInstances {
					if bm, err = pixas.GetBitmap(i); err != nil {
						return errors.Wrap(err, processName, "KeepClassInstances - i")
					}
					if bitmaps, err = c.ClassInstances.GetBitmaps(iclass); err != nil {
						return errors.Wrap(err, processName, "KeepClassInstances - iClass")
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
			if err = c.ClassIDs.Add(nt); err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			if err = c.ComponentPageNumbers.Add(pageNumber); err != nil {
				return errors.Wrap(err, processName, "!found")
			}

			bitmaps = &bitmap.Bitmaps{}
			if bm, err = pixas.GetBitmap(i); err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			bitmaps.AddBitmap(bm)

			wt, ht := bm.Width, bm.Height
			key := uint64(ht) * uint64(wt)
			c.TemplatesSize.Add(key, nt)
			if box, err = boxa.Get(i); err != nil {
				return errors.Wrap(err, processName, "!found")
			}
			bitmaps.AddBox(box)
			c.ClassInstances.AddBitmaps(bitmaps)
			c.CentroidPointsTemplates.AddPoint(x1, y1)
			c.FgTemplates.AddInt(area1)
			c.UndilatedTemplates.AddBitmap(bm)

			area = (bm1.Width - 2*JbAddedPixels) * (bm1.Height - 2*JbAddedPixels)
			if err = c.TemplateAreas.Add(area); err != nil {
				return errors.Wrap(err, processName, "!found")
			}
		}
	}
	c.NumberOfClasses = len(c.UndilatedTemplates.Values)
	return nil
}
