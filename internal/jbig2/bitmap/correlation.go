/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// CorrelationScore computes the correlation score between the bitmaps: 'bm1' and 'bm2'.
// The correlation score is the ratio of the square of the number of pixels AND of the two bitmaps
// to the product of the number of ON pixels in each.
func CorrelationScore(bm1, bm2 *Bitmap, area1, area2 int, delX, delY float32, maxDiffW, maxDiffH int, tab []int) (score float64, err error) {
	const processName = "correlationScore"
	if bm1 == nil || bm2 == nil {
		return 0, errors.Error(processName, "provided nil bitmaps")
	}
	if tab == nil {
		return 0, errors.Error(processName, "'tab' not defined")
	}
	if area1 <= 0 || area2 <= 0 {
		return 0, errors.Error(processName, "areas must be greater than 0")
	}

	wi, hi := bm1.Width, bm1.Height
	wt, ht := bm2.Width, bm2.Height

	delW := abs(wi - wt)
	if delW > maxDiffW {
		return 0, nil
	}
	delH := abs(hi - ht)
	if delH > maxDiffH {
		return 0, nil
	}
	var idelX, idelY int
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

	// decide which rows need to be considered.
	loRow := max(idelY, 0)
	hiRow := min(ht+idelY, hi)

	row1 := bm1.RowStride * loRow
	row2 := bm2.RowStride * (loRow - idelY)

	// decide which columns of bm1 should be considered
	loCol := max(idelX, 0)
	hiCol := min(wt+idelX, wi)
	rowBytes2 := bm2.RowStride
	var pix1LSkip, pix2LSkip int
	if idelX >= 8 {
		pix1LSkip = idelX >> 3
		row1 += pix1LSkip
		loCol -= pix1LSkip << 3
		hiCol -= pix1LSkip << 3
		idelX &= 7
	} else if idelX <= -8 {
		pix2LSkip = -((idelX + 7) >> 3)
		row2 += pix2LSkip
		rowBytes2 -= pix2LSkip
		idelX += pix2LSkip << 3
	}

	if loCol >= hiCol || loRow >= hiRow {
		return 0, nil
	}

	rowBytes1 := (hiCol + 7) >> 3
	var (
		bt1, bt2, andByte byte
		count, x, y       int
	)

	switch {
	case idelX == 0:
		for y = loRow; y < hiRow; y, row1, row2 = y+1, row1+bm1.RowStride, row2+bm2.RowStride {
			for x = 0; x < rowBytes1; x++ {
				andByte = bm1.Data[row1+x] & bm2.Data[row2+x]
				count += tab[andByte]
			}
		}
	case idelX > 0:
		if rowBytes2 < rowBytes1 {
			for y = loRow; y < hiRow; y, row1, row2 = y+1, row1+bm1.RowStride, row2+bm2.RowStride {
				bt1, bt2 = bm1.Data[row1], bm2.Data[row2]>>uint(idelX)
				andByte = bt1 & bt2
				count += tab[andByte]

				for x = 1; x < rowBytes2; x++ {
					bt1, bt2 = bm1.Data[row1+x], (bm2.Data[row2+x]>>uint(idelX))|(bm2.Data[row2+x-1]<<uint(8-idelX))
					andByte = bt1 & bt2
					count += tab[andByte]
				}

				bt1 = bm1.Data[row1+x]
				bt2 = bm2.Data[row2+x-1] << uint(8-idelX)
				andByte = bt1 & bt2
				count += tab[andByte]
			}
		} else {
			for y = loRow; y < hiRow; y, row1, row2 = y+1, row1+bm1.RowStride, row2+bm2.RowStride {
				bt1, bt2 = bm1.Data[row1], bm2.Data[row2]>>uint(idelX)
				andByte = bt1 & bt2
				count += tab[andByte]
				for x = 1; x < rowBytes1; x++ {
					bt1 = bm1.Data[row1+x]
					bt2 = (bm2.Data[row2+x] >> uint(idelX)) | (bm2.Data[row2+x-1] << uint(8-idelX))
					andByte = bt1 & bt2
					count += tab[andByte]
				}
			}
		}
	default:
		if rowBytes1 < rowBytes2 {
			for y = loRow; y < hiRow; y, row1, row2 = y+1, row1+bm1.RowStride, row2+bm2.RowStride {
				for x = 0; x < rowBytes1; x++ {
					bt1 = bm1.Data[row1+x]
					bt2 = bm2.Data[row2+x] << uint(-idelX)
					bt2 |= bm2.Data[row2+x+1] >> uint(8+idelX)
					andByte = bt1 & bt2
					count += tab[andByte]
				}
			}
		} else {
			for y = loRow; y < hiRow; y, row1, row2 = y+1, row1+bm1.RowStride, row2+bm2.RowStride {
				for x = 0; x < rowBytes1-1; x++ {
					bt1 = bm1.Data[row1+x]
					bt2 = bm2.Data[row2+x] << uint(-idelX)
					bt2 |= bm2.Data[row2+x+1] >> uint(8+idelX)
					andByte = bt1 & bt2
					count += tab[andByte]
				}
				bt1 = bm1.Data[row1+x]
				bt2 = bm2.Data[row2+x] << uint(-idelX)
				andByte = bt1 & bt2
				count += tab[andByte]
			}
		}
	}
	score = float64(count) * float64(count) / (float64(area1) * float64(area2))

	return score, nil
}

// CorrelationScoreSimple computes the correlation score value which should be the same as the result of the 'CorrelationScore' function.
// This function uses raster operations and is about 2-3x slower. This function makes it easier to understand how is the correlation computed.
func CorrelationScoreSimple(bm1, bm2 *Bitmap, area1, area2 int, delX, delY float32, maxDiffW, maxDiffH int, tab []int) (score float64, err error) {
	const processName = "CorrelationScoreSimple"
	if bm1 == nil || bm2 == nil {
		return score, errors.Error(processName, "nil bitmaps provided")
	}
	if tab == nil {
		return score, errors.Error(processName, "tab undefined")
	}
	if area1 == 0 || area2 == 0 {
		return score, errors.Error(processName, "provided areas must be > 0")
	}

	wi, hi := bm1.Width, bm1.Height
	wt, ht := bm2.Width, bm2.Height
	if abs(wi-wt) > maxDiffW {
		return 0, nil
	}
	if abs(hi-ht) > maxDiffH {
		return 0, nil
	}
	var idelX, idelY int
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
	bmT := bm1.createTemplate()
	if err = bmT.RasterOperation(idelX, idelY, wt, ht, PixSrc, bm2, 0, 0); err != nil {
		return score, errors.Wrap(err, processName, "bm2 to Template")
	}
	if err = bmT.RasterOperation(0, 0, wi, hi, PixSrcAndDst, bm1, 0, 0); err != nil {
		return score, errors.Wrap(err, processName, "bm1 and bmT")
	}
	count := bmT.countPixels()
	score = float64(count) * float64(count) / (float64(area1) * float64(area2))
	return score, nil
}

// CorrelationScoreThresholded checks whether the correlation score is >= scoreThreshold.
// 'area1' 				- number of 'ON' pixels in 'bm1'.
// 'area2' 				- number of 'ON' pixels in 'bm2'.
// 'delX' 				- x comp of centroid difference.
// 'delY'				- y comp of centroid difference.
// 'maxDiffW'			- max width difference between 'bm1' and 'bm2'.
// 'maxDiffH'			- max height difference between 'bm1' and 'bm2'.
// 'tab'				- is a sum tab for the and byte (created by MakeSumTab8 function).
// 'downcount' 			- is the number of 'ON' pixels below each row of bitmap 'bm1'.
// 'score_threshold'	- is the correlation score that the bitmaps should have at least to return false.
func CorrelationScoreThresholded(bm1, bm2 *Bitmap, area1, area2 int, delX, delY float32, maxDiffW, maxDiffH int, tab, downcount []int, scoreThreshold float32) (bool, error) {
	const processName = "CorrelationScoreThresholded"
	if bm1 == nil {
		return false, errors.Error(processName, "correlationScoreThresholded bm1 is nil")
	}
	if bm2 == nil {
		return false, errors.Error(processName, "correlationScoreThresholded bm2 is nil")
	}

	if area1 <= 0 || area2 <= 0 {
		return false, errors.Error(processName, "correlationScoreThresholded - areas must be > 0")
	}
	if downcount == nil {
		return false, errors.Error(processName, "provided no 'downcount'")
	}
	if tab == nil {
		return false, errors.Error(processName, "provided nil 'sumtab'")
	}

	wi, hi := bm1.Width, bm1.Height
	wt, ht := bm2.Width, bm2.Height

	if basic.Abs(wi-wt) > maxDiffW {
		return false, nil
	}
	if basic.Abs(hi-ht) > maxDiffH {
		return false, nil
	}

	idelX := int(delX + basic.Sign(delX)*0.5)
	idelY := int(delY + basic.Sign(delY)*0.5)

	// compute the correlation count.
	threshold := int(math.Ceil(math.Sqrt(float64(scoreThreshold) * float64(area1) * float64(area2))))
	rowBytes2 := bm2.RowStride

	// only the rows underlying the shifted bm2 need to be considered
	loRow := max(idelY, 0)
	hiRow := min(ht+idelY, hi)

	row1Index := bm1.RowStride * loRow
	row2Index := bm2.RowStride * (loRow - idelY)

	var untouchable int
	if hiRow <= hi {
		// some rows of bm1 would never contribute
		untouchable = downcount[hiRow-1]
	}

	loCol := max(idelX, 0)
	hiCol := min(wt+idelX, wi)
	var bm1LSkip, bm2LSkip int
	if idelX >= 8 {
		bm1LSkip = idelX >> 3
		row1Index += bm1LSkip
		loCol -= bm1LSkip << 3
		hiCol -= bm1LSkip << 3
		idelX &= 7
	} else if idelX <= -8 {
		bm2LSkip = -((idelX + 7) >> 3)
		row2Index += bm2LSkip
		rowBytes2 -= bm2LSkip
		idelX += bm2LSkip << 3
	}

	var (
		count, y, x           int
		andByte, byte1, byte2 byte
	)
	if loCol >= hiCol || loRow >= hiRow {
		// there is no overlap
		return false, nil
	}
	rowBytes1 := (hiCol + 7) >> 3

	switch {
	case idelX == 0:
		for y = loRow; y < hiRow; y, row1Index, row2Index = y+1, row1Index+bm1.RowStride, row2Index+bm2.RowStride {
			for x = 0; x < rowBytes1; x++ {
				andByte = bm1.Data[row1Index+x] & bm2.Data[row2Index+x]
				count += tab[andByte]
			}
			if count >= threshold {
				return true, nil
			}
			if v := count + downcount[y] - untouchable; v < threshold {
				return false, nil
			}
		}
	case idelX > 0 && rowBytes2 < rowBytes1:
		for y = loRow; y < hiRow; y, row1Index, row2Index = y+1, row1Index+bm1.RowStride, row2Index+bm2.RowStride {
			byte1 = bm1.Data[row1Index]
			byte2 = bm2.Data[row2Index] >> uint(idelX)
			andByte = byte1 & byte2
			count += tab[andByte]

			for x = 1; x < rowBytes2; x++ {
				byte1 = bm1.Data[row1Index+x]
				byte2 = bm2.Data[row2Index+x]>>uint(idelX) | bm2.Data[row2Index+x-1]<<uint(8-idelX)
				andByte = byte1 & byte2
				count += tab[andByte]
			}

			byte1 = bm1.Data[row1Index+x]
			byte2 = bm2.Data[row2Index+x-1] << uint(8-idelX)
			andByte = byte1 & byte2
			count += tab[andByte]

			if count >= threshold {
				return true, nil
			} else if count+downcount[y]-untouchable < threshold {
				return false, nil
			}
		}
	case idelX > 0 && rowBytes2 >= rowBytes1:
		for y = loRow; y < hiRow; y, row1Index, row2Index = y+1, row1Index+bm1.RowStride, row2Index+bm2.RowStride {
			byte1 = bm1.Data[row1Index]
			byte2 = bm2.Data[row2Index] >> uint(idelX)

			andByte = byte1 & byte2
			count += tab[andByte]

			for x = 1; x < rowBytes1; x++ {
				byte1 = bm1.Data[row1Index+x]
				byte2 = bm2.Data[row2Index+x] >> uint(idelX)
				byte2 |= bm2.Data[row2Index+x-1] << uint(8-idelX)
				andByte = byte1 & byte2
				count += tab[andByte]
			}
			if count >= threshold {
				return true, nil
			} else if count+downcount[y]-untouchable < threshold {
				return false, nil
			}
		}
	case rowBytes1 < rowBytes2:
		for y = loRow; y < hiRow; y, row1Index, row2Index = y+1, row1Index+bm1.RowStride, row2Index+bm2.RowStride {
			for x = 0; x < rowBytes1; x++ {
				byte1 = bm1.Data[row1Index+x]
				byte2 = bm2.Data[row2Index+x] << uint(-idelX)
				byte2 |= bm2.Data[row2Index+x+1] >> uint(8+idelX)
				andByte = byte1 & byte2
				count += tab[andByte]
			}

			if count >= threshold {
				return true, nil
			} else if left := count + downcount[y] - untouchable; left < threshold {
				return false, nil
			}
		}
	case rowBytes2 >= rowBytes1:
		for y = loRow; y < hiRow; y, row1Index, row2Index = y+1, row1Index+bm1.RowStride, row2Index+bm2.RowStride {
			for x = 0; x < rowBytes1; x++ {
				byte1 = bm1.Data[row1Index+x]
				byte2 = bm2.Data[row2Index+x] << uint(-idelX)
				byte2 |= bm2.Data[row2Index+x+1] >> uint(8+idelX)
				andByte = byte1 & byte2
				count += tab[andByte]
			}

			byte1 = bm1.Data[row1Index+x]
			byte2 = bm2.Data[row2Index+x] << uint(-idelX)
			andByte = byte1 & byte2
			count += tab[andByte]

			if count >= threshold {
				return true, nil
			} else if count+downcount[y]-untouchable < threshold {
				return false, nil
			}
		}
	}

	score := float32(count) * float32(count) / (float32(area1) * float32(area2))
	if score >= scoreThreshold {
		common.Log.Trace("count: %d < threshold %d but score %f >= scoreThreshold %f", count, threshold, score, scoreThreshold)
	}
	return false, nil
}

// HausTest does the Hausdorff 2-way check for the provided bitmaps.
// Parameters:
//	p1			- new not dilated bitmap
//	p2			- new dilated bitmap
//  p3			- exemplar not dilated bitmap
//	p4			- exemplar dilated bitmap
//	delX, delY	- component centroid difference for 'x' and 'y' coordinates.
//	maxDiffW	- maximum width difference of 'p1' and 'p2'
// 	maxDiffH	- maximum height difference of 'p1' and 'p2'
// The centroid difference is used to align two images to the nearest integer for each check.
// It checks if the dilated image of one contains all the pixels of the undilated image of the other.
func HausTest(p1, p2, p3, p4 *Bitmap, delX, delY float32, maxDiffW, maxDiffH int) (bool, error) {
	const processName = "HausTest"

	// do a short check if the size is out of possible difference.
	wi, hi := p1.Width, p1.Height
	wt, ht := p3.Width, p3.Height

	if basic.Abs(wi-wt) > maxDiffW {
		return false, nil
	}
	if basic.Abs(hi-ht) > maxDiffH {
		return false, nil
	}

	// round difference in centroid location to nearest integer.
	idelX := int(delX + basic.Sign(delX)*0.5)
	idelY := int(delY + basic.Sign(delY)*0.5)

	var err error
	// do 1-direction hausdorff
	pt := p1.CreateTemplate()
	if err = pt.RasterOperation(0, 0, wi, hi, PixSrc, p1, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "p1 -SRC-> t")
	}
	if err = pt.RasterOperation(idelX, idelY, wi, hi, PixNotSrcAndDst, p4, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "!p4 & t")
	}

	if pt.Zero() {
		return false, nil
	}

	if err = pt.RasterOperation(idelX, idelY, wt, ht, PixSrc, p3, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "p3 -SRC-> t")
	}
	if err = pt.RasterOperation(0, 0, wt, ht, PixNotSrcAndDst, p2, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "!p2 & t")
	}
	return pt.Zero(), nil
}

// RankHausTest does the test of the Hausdorff ranked check.
// Parameters:
//	p1			- new bitmap, not dilated
//	p2			- new bitmap, dilated
//	p3			-u exemplar bitmap, not dilated
//	p4			- exemplar bitmap, dilated
//	delX, delY	- component centroid difference for 'x' and 'y' coordinates
//	maxDiffW	- maximum Width difference of 'p1' and 'p2'
// 	maxDiffH	- maximum Height difference of 'p1' and 'p2'
//	area1		- 'ON' - fg - pixels area of the 'p1' bitmap
// 	area3		- 'ON' - fg - pixels area of the 'p3' bitmap
//	rank		- rank value of the test
//	tab8		- table of the pixel sums for a single byte.
// The 'rank' value is being converted to a number of pixels by multiplication with the number of undilated images.
// The centroid difference is used for alignment of the images.
// The rank Hausdorff checks if dilated image of one contains the rank fraction pixels of the undilated image of the other in both directions.
func RankHausTest(p1, p2, p3, p4 *Bitmap, delX, delY float32, maxDiffW, maxDiffH, area1, area3 int, rank float32, tab8 []int) (match bool, err error) {
	const processName = "RankHausTest"

	// at first try to eliminate possible matches based on the size difference.
	wi, hi := p1.Width, p1.Height
	wt, ht := p3.Width, p3.Height

	if basic.Abs(wi-wt) > maxDiffW {
		return false, nil
	}
	if basic.Abs(hi-ht) > maxDiffH {
		return false, nil
	}

	// convert the rank value into pixel threshold for 'p1' and 'p3' 'ON' pixels areas.
	thresh1 := int(float32(area1)*(1.0-rank) + 0.5)
	thresh3 := int(float32(area3)*(1.0-rank) + 0.5)

	// round the difference of the centroid locations.
	var iDelX, iDelY int
	if delX >= 0 {
		iDelX = int(delX + 0.5)
	} else {
		iDelX = int(delX - 0.5)
	}

	if delY >= 0 {
		iDelY = int(delY + 0.5)
	} else {
		iDelY = int(delY - 0.5)
	}

	// do the 1-direction hausdorff check, if every pixel in 'p1' is within a dilation distance of some pixel in 'p3'.
	t := p1.CreateTemplate()
	if err = t.RasterOperation(0, 0, wi, hi, PixSrc, p1, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "p1 -SRC-> t")
	}
	if err = t.RasterOperation(iDelX, iDelY, wi, hi, PixNotSrcAndDst, p4, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "t & !p4")
	}

	// check if this hausdorff test check is within a threshold of 'p1' bitmap.
	match, err = t.ThresholdPixelSum(thresh1, tab8)
	if err != nil {
		return false, errors.Wrap(err, processName, "t->thresh1")
	}
	if match {
		return false, nil
	}

	// now do a 1-direction hausdorff checking that every pixel of 'p3' is within dilation distance of some pixel in 'p1'.
	// (p2 entirely covers p3)
	if err = t.RasterOperation(iDelX, iDelY, wt, ht, PixSrc, p3, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "p3 -SRC-> t")
	}
	if err = t.RasterOperation(0, 0, wt, ht, PixNotSrcAndDst, p2, 0, 0); err != nil {
		return false, errors.Wrap(err, processName, "t & !p2")
	}

	// check if the pixel sum of the 't' bitmap is above the provided threshold
	match, err = t.ThresholdPixelSum(thresh3, tab8)
	if err != nil {
		return false, errors.Wrap(err, processName, "t->thresh3")
	}
	return !match, nil
}
