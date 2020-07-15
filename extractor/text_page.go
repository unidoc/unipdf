/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// makeTextPage builds a paraList from `marks`, the textMarks on a page.
// The paraList contains the page arranged as
//  - a list of texPara in reading order
//  - each textPara contains list of textLine (text lines or parts of text lines) in reading order
//  - each textLine contains a list of textWord (words or parts of words) in reading order
// The paraList is thus an ordering of words on a page.
//   - Users of the paraList are expected to work with words. This should be adequate for most uses
//     as words are the basic unit of meaning in written language.
//   - However we provide links back from the extracted text to the textMarks as follows.
//        * paraList.writeText() returns the extracted text for a page
//        * paras.toTextMarks() returns a TextMarkArray containing the marks
//        * TextMarkArray.RangeOffset(lo, hi) return the marks corresponding offsets [lo:hi] in the
//          extracted text.
// NOTE: The "parts of words" occur because of hyphenation. We do some weak coordinate based
//        dehypenation. Caller who need strong dehypenation should use NLP librarie.
//       The "parts of lines" are an implementation detail. Line fragments are combined in
//        paraList.writeText()
// ALGORITHM:
// 1) Group the textMarks into textWords based on their bounding boxes.
// 2) Group the textWords into textParas based on their bounding boxes.
// 3) Detect textParas arranged as cells in a table and convert each one to a textPara containing a
//    textTable.
// 4) Sort the textParas in reading order.
func makeTextPage(marks []*textMark, pageSize model.PdfRectangle) paraList {
	common.Log.Trace("makeTextPage: %d elements pageSize=%.2f", len(marks), pageSize)
	if len(marks) == 0 {
		return nil
	}

	// Group the marks into word fragments
	words := makeTextWords(marks, pageSize)
	if len(words) == 0 {
		return nil
	}

	// Put the word fragments into a container that facilitates the grouping of words into paragraphs.
	pageWords := makeWordBag(words, pageSize.Ury)

	// Divide the page into rectangular regions for each paragraph and creata a wordBag for each one.
	paraWords := dividePage(pageWords, pageSize.Ury)
	paraWords = mergeWordBags(paraWords)

	// Arrange the contents of each paragraph wordBag into lines and the lines into whole words.
	paras := make(paraList, 0, len(paraWords))
	for _, bag := range paraWords {
		para := bag.arrangeText()
		if para != nil {
			paras = append(paras, para)
		}
	}

	// Find paras that are cells in tables, convert the tables to paras and remove the cell paras.
	if len(paras) >= minTableParas {
		paras = paras.extractTables()
	}

	// Sort the paras into reading order.
	paras.sortReadingOrder()
	paras.log("sorted in reading order")

	return paras
}

// dividePage divides `pageWords`, the page wordBag, into a list of paragraph wordBags.
func dividePage(pageWords *wordBag, pageHeight float64) []*wordBag {
	var paraWordBags []*wordBag

	// We move words from `page` to paras until there no words left in page.
	// We do this by iterating through `page` in depth bin order and, for each surving bin (see
	// below),  creating a paragraph with seed word, `words[0]` in the code below.
	// We then move words from around the `para` region from `page` to `para` .
	// This may empty some page bins before we iterate to them
	// Some bins are emptied before they iterated to (seee "surving bin" above).
	// If a `page` survives until it is iterated to then at least one `para` will be built around it.

	for _, depthIdx := range pageWords.depthIndexes() {
		changed := false
		for !pageWords.empty(depthIdx) {
			// Start a new paragraph region `paraWords`.
			// Build `paraWords` out from the left-most (lowest in reading direction) word `words`[0],
			// in the bins in and below `depthIdx`.

			// `firstWord` is the left-most word from the bins in and a few lines below `depthIdx`. We
			// seed 'paraWords` with this word.
			firstReadingIdx := pageWords.firstReadingIndex(depthIdx)
			firstWord := pageWords.firstWord(firstReadingIdx)
			paraWords := newWordBag(firstWord, pageHeight)
			pageWords.removeWord(firstWord, firstReadingIdx)
			if verbosePage {
				common.Log.Info("words[0]=%s", firstWord.String())
			}

			// The following 3 numbers define whether words should be added to `paraWords`.
			minInterReadingGap := minInterReadingGapR * paraWords.fontsize
			maxIntraReadingGap := maxIntraReadingGapR * paraWords.fontsize
			maxIntraDepthGap := maxIntraDepthGapR * paraWords.fontsize

			// Add words to `paraWords` until we pass through the following loop without adding a
			// new word.
			for running := true; running; running = changed {
				changed = false

				// Add words that are within maxIntraDepthGap of `paraWords` in the depth direction.
				// i.e. Stretch paraWords in the depth direction, vertically for English text.
				if verbosePage {
					common.Log.Info("paraWords depth %.2f - %.2f maxIntraDepthGap=%.2f ",
						paraWords.minDepth(), paraWords.maxDepth(), maxIntraDepthGap)
				}
				if pageWords.scanBand("vertical", paraWords, partial(readingOverlapPlusGap, 0),
					paraWords.minDepth()-maxIntraDepthGap, paraWords.maxDepth()+maxIntraDepthGap,
					maxIntraDepthFontTolR, false, false) > 0 {
					changed = true
				}
				// Add words that are within maxIntraReadingGap of `paraWords` in the reading direction.
				// i.e. Stretch paraWords in the reading direction, horizontall for English text.
				if pageWords.scanBand("horizontal", paraWords, partial(readingOverlapPlusGap, maxIntraReadingGap),
					paraWords.minDepth(), paraWords.maxDepth(),
					maxIntraReadingFontTol, false, false) > 0 {
					changed = true
				}
				// The above stretching has got as far as it can go. Repeating it won't pull in more words.

				// Only try to combine other words if we can't grow paraWords in the simple way above.
				if changed {
					continue
				}

				// In the following cases, we don't expand `paraWords` while scanning. We look for words
				// around paraWords. If we find them, we add them then expand `paraWords` when we are done.
				// This pulls the numbers to the left of paraWords into paraWords
				// e.g. From
				// 		Regulatory compliance
				// 		Archiving
				// 		Document search
				// to
				// 		1. Regulatory compliance
				// 		2. Archiving
				// 		3. Document search

				// If there are words to the left of `paraWords`, add them.
				// We need to limit the number of words.
				n := pageWords.scanBand("", paraWords, partial(readingOverlapLeft, minInterReadingGap),
					paraWords.minDepth(), paraWords.maxDepth(),
					minInterReadingFontTol, true, false)
				if n > 0 {
					r := (paraWords.maxDepth() - paraWords.minDepth()) / paraWords.fontsize
					if (n > 1 && float64(n) > 0.3*r) || n <= 10 {
						if pageWords.scanBand("other", paraWords, partial(readingOverlapLeft, minInterReadingGap),
							paraWords.minDepth(), paraWords.maxDepth(),
							minInterReadingFontTol, false, true) > 0 {
							changed = true
						}
					}
				}
			}
			paraWordBags = append(paraWordBags, paraWords)
		}
	}

	return paraWordBags
}

// writeText writes the text in `paras` to `w`.
func (paras paraList) writeText(w io.Writer) {
	for ip, para := range paras {
		para.writeText(w)
		if ip != len(paras)-1 {
			if sameLine(para, paras[ip+1]) {
				w.Write([]byte(" "))
			} else {
				w.Write([]byte("\n"))
				w.Write([]byte("\n"))
			}
		}
	}
	w.Write([]byte("\n"))
	w.Write([]byte("\n"))
}

// toTextMarks creates the TextMarkArray corresponding to the extracted text created by
// `paras`.writeText().
func (paras paraList) toTextMarks() []TextMark {
	offset := 0
	var marks []TextMark
	for ip, para := range paras {
		paraMarks := para.toTextMarks(&offset)
		marks = append(marks, paraMarks...)
		if ip != len(paras)-1 {
			if sameLine(para, paras[ip+1]) {
				marks = appendSpaceMark(marks, &offset, " ")
			} else {
				marks = appendSpaceMark(marks, &offset, "\n")
				marks = appendSpaceMark(marks, &offset, "\n")
			}
		}
	}
	marks = appendSpaceMark(marks, &offset, "\n")
	marks = appendSpaceMark(marks, &offset, "\n")
	return marks
}

// sameLine returms true if `para1` and `para2` are on the same line.
func sameLine(para1, para2 *textPara) bool {
	return isZero(para1.depth() - para2.depth())
}

// tables returns the tables from all the paras that contain them.
func (paras paraList) tables() []TextTable {
	var tables []TextTable
	for _, para := range paras {
		if para.table != nil {
			tables = append(tables, para.table.toTextTable())
		}
	}
	return tables
}

// sortReadingOrder sorts `paras` in reading order.
func (paras paraList) sortReadingOrder() {
	common.Log.Trace("sortReadingOrder: paras=%d ===========x=============", len(paras))
	if len(paras) <= 1 {
		return
	}
	paras.computeEBBoxes()
	sort.Slice(paras, func(i, j int) bool { return diffDepthReading(paras[i], paras[j]) <= 0 })
	order := paras.topoOrder()
	paras.reorder(order)
}

// topoOrder returns the ordering of the topological sort of `paras` using readBefore() to determine
// the incoming nodes to each node.
func (paras paraList) topoOrder() []int {
	if verbosePage {
		common.Log.Info("topoOrder:")
	}
	n := len(paras)
	visited := make([]bool, n)
	order := make([]int, 0, n)
	llyOrder := paras.llyOrdering()

	// sortNode recursively sorts below node `idx` in the adjacency matrix.
	var sortNode func(idx int)
	sortNode = func(idx int) {
		visited[idx] = true
		for i := 0; i < n; i++ {
			if !visited[i] {
				if paras.readBefore(llyOrder, idx, i) {
					sortNode(i)
				}
			}
		}
		order = append(order, idx) // Should prepend but it's cheaper to append and reverse later.
	}

	for idx := 0; idx < n; idx++ {
		if !visited[idx] {
			sortNode(idx)
		}
	}

	return reversed(order)
}

// readBefore returns true if paras[`i`] comes before paras[`j`].
// readBefore defines an ordering over `paras`.
// a = paras[i],  b= paras[j]
// 1. Line segment `a` comes before line segment `b` if their ranges of x-coordinates overlap and if
//    line segment `a` is above line segment `b` on the page.
// 2. Line segment `a` comes before line segment `b` if `a` is entirely to the left of `b` and if
//    there does not exist a line segment `c` whose y-coordinates are between `a` and `b` and whose
//    range of x coordinates overlaps both `a` and `b`.
// From Thomas M. Breuel "High Performance Document Layout Analysis"
func (paras paraList) readBefore(ordering []int, i, j int) bool {
	a, b := paras[i], paras[j]
	// Breuel's rule 1
	if overlappedXPara(a, b) && a.Lly > b.Lly {
		return true
	}

	// Breuel's rule 2
	if !(a.eBBox.Urx < b.eBBox.Llx) {
		return false
	}

	lo, hi := a.Lly, b.Lly
	if lo > hi {
		hi, lo = lo, hi
	}
	llx := math.Max(a.eBBox.Llx, b.eBBox.Llx)
	urx := math.Min(a.eBBox.Urx, b.eBBox.Urx)

	llyOrder := paras.llyRange(ordering, lo, hi)
	for _, k := range llyOrder {
		if k == i || k == j {
			continue
		}
		c := paras[k]
		if c.eBBox.Llx <= urx && llx <= c.eBBox.Urx {
			return false
		}
	}
	return true
}

// overlappedX returns true if `r0` and `r1` overlap on the x-axis.
func overlappedXPara(r0, r1 *textPara) bool {
	return intersectsX(r0.eBBox, r1.eBBox)
}

// llyOrdering is ordering over the indexes of `paras` sorted by Llx is increasing order.
func (paras paraList) llyOrdering() []int {
	ordering := make([]int, len(paras))
	for i := range paras {
		ordering[i] = i
	}
	sort.SliceStable(ordering, func(i, j int) bool {
		oi, oj := ordering[i], ordering[j]
		return paras[oi].Lly < paras[oj].Lly
	})
	return ordering
}

// llyRange returns the indexes in `paras` of paras p: lo <= p.Llx < hi
func (paras paraList) llyRange(ordering []int, lo, hi float64) []int {
	n := len(paras)
	if hi < paras[ordering[0]].Lly || lo > paras[ordering[n-1]].Lly {
		return nil
	}

	// i0 is the lowest i: lly(i) >= lo
	// i1 is the lowest i: lly(i) > hi
	i0 := sort.Search(n, func(i int) bool { return paras[ordering[i]].Lly >= lo })
	i1 := sort.Search(n, func(i int) bool { return paras[ordering[i]].Lly > hi })

	return ordering[i0:i1]
}

// computeEBBoxes computes the eBBox fields in the elements of `paras`.
// The EBBoxs are the regions around the paras that don't intersect paras in other columns.
// This is needed for sortReadingOrder to work with skinny paras in a column of fat paras. The
// sorting assumes the skinny para bounding box is as wide as the fat para bounding boxes.
func (paras paraList) computeEBBoxes() {
	if verbose {
		common.Log.Info("computeEBBoxes:")
	}

	for _, para := range paras {
		para.eBBox = para.PdfRectangle
	}
	paraYNeighbours := paras.yNeighbours()

	for i, aa := range paras {
		a := aa.eBBox
		// [llx, urx] is the reading direction interval for which no paras overlap `a`.
		llx, urx := -1.0e9, +1.0e9

		for _, j := range paraYNeighbours[aa] {
			b := paras[j].eBBox
			if b.Urx < a.Llx { // `b` to left of `a`. no x overlap.
				llx = math.Max(llx, b.Urx)
			} else if a.Urx < b.Llx { // `b` to right of `a`. no x overlap.
				urx = math.Min(urx, b.Llx)
			}
		}

		// llx extends left from `a` and overlaps no other paras.
		// urx extends right from `a` and overlaps no other paras.

		// Go through all paras below `a` within interval [llx, urx] in the reading direction and
		// expand `a` as far as possible to left and right without overlapping any of them.
		for j, bb := range paras {
			b := bb.eBBox
			if i == j || b.Ury > a.Lly {
				continue
			}

			if llx <= b.Llx && b.Llx < a.Llx {
				// If `b` is completely to right of `llx`, extend `a` left to `b`.
				a.Llx = b.Llx
			} else if b.Urx <= urx && a.Urx < b.Urx {
				// If `b` is completely to left of `urx`, extend `a` right to `b`.
				a.Urx = b.Urx
			}
		}
		if verbose {
			fmt.Printf("%4d: %6.2f->%6.2f %q\n", i, aa.eBBox, a, truncate(aa.text(), 50))
		}
		aa.eBBox = a
	}
	if useEBBox {
		for _, para := range paras {
			para.PdfRectangle = para.eBBox
		}
	}
}

// reversed return `order` reversed.
func reversed(order []int) []int {
	rev := make([]int, len(order))
	for i, v := range order {
		rev[len(order)-1-i] = v
	}
	return rev
}

// reorder reorders `para` to the order in `order`.
func (paras paraList) reorder(order []int) {
	sorted := make(paraList, len(paras))
	for i, k := range order {
		sorted[i] = paras[k]
	}
	copy(paras, sorted)
}
