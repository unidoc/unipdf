/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"math"

	"github.com/unidoc/unipdf/v3/model"
)

/*
 * Sorting functions.
 *
 * There are two directions:
 *  - reading. Left to right in English
 *  - depth (aka non-reading).  Top to botttom in English.
 *
 * Text is read in reading then depth order.
 *
 * TODO(peterwilliams97): Add support for other reading orders and page rotations
 */

// bounded is an object with a bounding box. A mark, word, line or para.
type bounded interface {
	bbox() model.PdfRectangle
}

// getDepth returns the depth of `a` on a page of size `pageSize`.
func getDepth(pageSize model.PdfRectangle, a bounded) float64 {
	return pageSize.Ury - a.bbox().Lly
}

// diffReading returns `a` - `b` in the reading direction.
func diffReading(a, b bounded) float64 {
	return a.bbox().Llx - b.bbox().Llx
}

// rectContainsRect returns true if `a` contains `b`.
func rectContainsRect(a, b model.PdfRectangle) bool {
	return a.Llx <= b.Llx && b.Urx <= a.Urx && a.Lly <= b.Lly && b.Ury <= a.Ury
}

// diffDepth returns `a` - `b` in the depth direction.
func diffDepth(a, b bounded) float64 {
	return bboxDepth(a) - bboxDepth(b)
}

// diffReadingDepth returns `a` - `b` in the reading then depth direction..
func diffReadingDepth(a, b bounded) float64 {
	diff := diffReading(a, b)
	if !isZero(diff) {
		return diff
	}
	return diffDepth(a, b)
}

// diffDepthReading returns `a` - `b` in the depth then reading directions
func diffDepthReading(a, b bounded) float64 {
	cmp := diffDepth(a, b)
	if !isZero(cmp) {
		return cmp
	}
	return diffReading(a, b)
}

// gapReading returns the reading direction gap between `a` and the following object `b` in the
// reading direction.
func gapReading(a, b bounded) float64 {
	return a.bbox().Llx - b.bbox().Urx
}

// bboxDepth returns the relative depth of `b`. Depth is only used for comparison so we don't care
// about its absolute value
func bboxDepth(b bounded) float64 {
	return -b.bbox().Lly
}

// readingOverlapLeft returns true is the left of `word` is in within `para` or delta to its right
func readingOverlapLeft(para *wordBag, word *textWord, delta float64) bool {
	return para.Urx <= word.Llx && word.Llx < para.Urx+delta
}

// readingOverlapPlusGap returns true if `word` overlaps [para.Llx-maxIntraReadingGap, para.Urx+maxIntraReadingGap]
// in the reading direction.
func readingOverlapPlusGap(para *wordBag, word *textWord, maxIntraReadingGap float64) bool {
	return word.Llx < para.Urx+maxIntraReadingGap && para.Llx-maxIntraReadingGap < word.Urx
}

// partial return 'overlap`(*wordBag, *textWord, `param`) bool.
func partial(overlap func(*wordBag, *textWord, float64) bool,
	param float64) func(*wordBag, *textWord) bool {
	return func(para *wordBag, word *textWord) bool {
		return overlap(para, word, param)
	}
}

// rectUnion returns the smallest axis-aligned rectangle that contains `b1` and `b2`.
func rectUnion(b1, b2 model.PdfRectangle) model.PdfRectangle {
	return model.PdfRectangle{
		Llx: math.Min(b1.Llx, b2.Llx),
		Lly: math.Min(b1.Lly, b2.Lly),
		Urx: math.Max(b1.Urx, b2.Urx),
		Ury: math.Max(b1.Ury, b2.Ury),
	}
}

// rectIntersection returns the largest axis-aligned rectangle that is contained by `b1` and `b2`.
func rectIntersection(b1, b2 model.PdfRectangle) (model.PdfRectangle, bool) {
	if !intersects(b1, b2) {
		return model.PdfRectangle{}, false
	}
	return model.PdfRectangle{
		Llx: math.Max(b1.Llx, b2.Llx),
		Urx: math.Min(b1.Urx, b2.Urx),
		Lly: math.Max(b1.Lly, b2.Lly),
		Ury: math.Min(b1.Ury, b2.Ury),
	}, true
}

// intersects returns true if `r0` and `r1` overlap in the x and y axes.
func intersects(b1, b2 model.PdfRectangle) bool {
	return intersectsX(b1, b2) && intersectsY(b1, b2)
}

// intersectsX returns true if `r0` and `r1` overlap in the x axis.
func intersectsX(r0, r1 model.PdfRectangle) bool {
	return r1.Llx <= r0.Urx && r0.Llx <= r1.Urx
}

// intersectsY returns true if `r0` and `r1` overlap in the y axis.
func intersectsY(r0, r1 model.PdfRectangle) bool {
	return r0.Lly <= r1.Ury && r1.Lly <= r0.Ury
}
