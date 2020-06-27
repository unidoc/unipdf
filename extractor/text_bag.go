/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// wordBag is just a list of textWords in a rectangular region. It is needed for efficient
// comparison of the bounding boxes of the words to arrange them into paragraph regions.
// The implementation is not important as long as it implements the main function scanBand()
// efficiently.
// In the current implementation, wordBag is a list of word fragment bins arranged by their depth on
// a page with the word fragments  in each bin are sorted in reading order.
type wordBag struct {
	model.PdfRectangle         // Bounding box of all the textWord in the wordBag.
	fontsize           float64 // The size of the largest font in the wordBag.
	// The following fields are for the current bin based implementation
	pageHeight float64             // Used to calculate depths
	bins       map[int][]*textWord // bins[n] = w: n*depthBinPoints <= w.depth < (n+1)*depthBinPoints
}

// makeWordBag return a wordBag containg `words`
// In the current implementation, it does this by putting the words into the appropriate depth bins.
// Caller must check that `words` has at least one element.
func makeWordBag(words []*textWord, pageHeight float64) *wordBag {
	b := newWordBag(words[0], pageHeight)
	for _, w := range words[1:] {
		depthIdx := depthIndex(w.depth)
		b.bins[depthIdx] = append(b.bins[depthIdx], w)
	}
	b.sort()
	return b
}

// newWordBag returns a wordBag with page height `pageHeight` with the single word fragment `word`.
func newWordBag(word *textWord, pageHeight float64) *wordBag {
	depthIdx := depthIndex(word.depth)
	words := []*textWord{word}
	bag := wordBag{
		bins:         map[int][]*textWord{depthIdx: words},
		PdfRectangle: word.PdfRectangle,
		fontsize:     word.fontsize,
		pageHeight:   pageHeight,
	}
	return &bag
}

// String returns a description of `b`.
func (b *wordBag) String() string {
	var texts []string
	for _, depthIdx := range b.depthIndexes() {
		words, _ := b.bins[depthIdx]
		for _, w := range words {
			texts = append(texts, w.text)
		}
	}
	return fmt.Sprintf("%.2f fontsize=%.2f %d %q", b.PdfRectangle, b.fontsize, len(texts), texts)
}

// scanBand scans the bins for words w:
//     `minDepth` <= w.depth <= `maxDepth` &&  // in the depth diraction
//    `readingOverlap`(`para`, w) &&  // in the reading directon
//     math.Abs(w.fontsize-fontsize) > `fontTol`*fontsize // font size tolerance
// and applies `moveWord`(depthIdx, s,para w) to them.
// If `detectOnly` is true, moveWord is not applied.
// If `freezeDepth` is true, minDepth and maxDepth are not updated in scan as words are added.
func (b *wordBag) scanBand(title string, para *wordBag,
	readingOverlap func(para *wordBag, word *textWord) bool,
	minDepth, maxDepth, fontTol float64,
	detectOnly, freezeDepth bool) int {
	fontsize := para.fontsize
	lineDepth := lineDepthR * fontsize
	n := 0
	minDepth0, maxDepth0 := minDepth, maxDepth
	var newWords []*textWord
	for _, depthIdx := range b.depthBand(minDepth-lineDepth, maxDepth+lineDepth) {
		for _, word := range b.bins[depthIdx] {
			if !(minDepth-lineDepth <= word.depth && word.depth <= maxDepth+lineDepth) {
				continue
			}
			if !readingOverlap(para, word) {
				continue
			}
			fontRatio1 := math.Abs(word.fontsize-fontsize) / fontsize
			fontRatio2 := word.fontsize / fontsize
			fontRatio := math.Min(fontRatio1, fontRatio2)
			if fontTol > 0 {
				if fontRatio > fontTol {
					continue
				}
			}

			if !detectOnly {
				para.pullWord(b, word, depthIdx)
			}
			newWords = append(newWords, word)
			n++
			if !freezeDepth {
				if word.depth < minDepth {
					minDepth = word.depth
				}
				if word.depth > maxDepth {
					maxDepth = word.depth
				}
			}
			// Has no effect on results
			// fontsize = para.fontsize
			// lineDepth = lineDepthR * fontsize
			if detectOnly {
				break
			}
		}
	}
	if verbose {
		if len(title) > 0 {
			common.Log.Info("scanBand: %s [%.2f %.2f]->[%.2f %.2f] para=%.2f fontsize=%.2f %q",
				title,
				minDepth0, maxDepth0,
				minDepth, maxDepth,
				para.PdfRectangle, para.fontsize, truncate(para.text(), 20))
			for i, word := range newWords {
				fmt.Printf("  %q", word.text)
				if i >= 5 {
					break
				}
			}
			if len(newWords) > 0 {
				fmt.Println()
			}
		}
	}
	return n
}

// highestWord returns the hight word in b.bins[depthIdx] w: minDepth <= w.depth <= maxDepth.
func (b *wordBag) highestWord(depthIdx int, minDepth, maxDepth float64) *textWord {
	for _, word := range b.bins[depthIdx] {
		if minDepth <= word.depth && word.depth <= maxDepth {
			return word
		}
	}
	return nil
}

// depthBand returns the indexes of the bins with depth: `minDepth` <= depth <= `maxDepth`.
func (b *wordBag) depthBand(minDepth, maxDepth float64) []int {
	if len(b.bins) == 0 {
		return nil
	}
	return b.depthRange(b.getDepthIdx(minDepth), b.getDepthIdx(maxDepth))
}

// depthRange returns the sorted keys of b.bins for depths indexes [`minDepth`,`maxDepth`).
func (b *wordBag) depthRange(minDepthIdx, maxDepthIdx int) []int {
	indexes := b.depthIndexes()
	var rangeIndexes []int
	for _, depthIdx := range indexes {
		if minDepthIdx <= depthIdx && depthIdx <= maxDepthIdx {
			rangeIndexes = append(rangeIndexes, depthIdx)
		}
	}
	return rangeIndexes
}

// firstReadingIndex returns the index of the bin containing the left-most word near the top of `b`.
// Precisely, this is the index of the depth bin that starts with that word with the smallest
// reading direction value in the depth region `minDepthIndex` < depth <= minDepthIndex+ 4*fontsize
// The point of this function is to find the top-most left-most word in `b` that is not a superscript.
func (b *wordBag) firstReadingIndex(minDepthIdx int) int {
	fontsize := b.firstWord(minDepthIdx).fontsize
	minDepth := float64(minDepthIdx+1) * depthBinPoints
	maxDepth := minDepth + topWordRangeR*fontsize
	firstReadingIdx := minDepthIdx
	for _, depthIdx := range b.depthBand(minDepth, maxDepth) {
		if diffReading(b.firstWord(depthIdx), b.firstWord(firstReadingIdx)) < 0 {
			firstReadingIdx = depthIdx
		}
	}
	return firstReadingIdx
}

// getDepthIdx returns the index into `b.bins` for depth axis value `depth`.
// Caller must check that len(b.bins) > 0.
func (b *wordBag) getDepthIdx(depth float64) int {
	indexes := b.depthIndexes()
	depthIdx := depthIndex(depth)
	if depthIdx < indexes[0] {
		return indexes[0]
	}
	if depthIdx > indexes[len(indexes)-1] {
		return indexes[len(indexes)-1]
	}
	return depthIdx
}

// empty returns true if the depth bin with index `depthIdx` is empty.
// NOTE: We delete bins as soon as they become empty so we just have to check for the bin's existence.
func (b *wordBag) empty(depthIdx int) bool {
	_, ok := b.bins[depthIdx]
	return !ok
}

// firstWord returns the first word in reading order in bin `depthIdx`.
func (b *wordBag) firstWord(depthIdx int) *textWord {
	return b.bins[depthIdx][0]
}

// stratum returns a copy of `b`.bins[`depthIdx`].
// stratum is guaranteed to return a non-nil value. It must be called with a valid depth index.
// NOTE: We need to return a copy because remove() and other functions manipulate the array
// underlying the slice.
func (b *wordBag) stratum(depthIdx int) []*textWord {
	words := b.bins[depthIdx]
	dup := make([]*textWord, len(words))
	copy(dup, words)
	return dup
}

// pullWord adds `word` to `b` and removes it from `bag`.
// `depthIdx` is the depth index of `word` in all wordBags.
// TODO(peterwilliams97): Compute depthIdx from `word` instead of passing it around.
func (b *wordBag) pullWord(bag *wordBag, word *textWord, depthIdx int) {
	b.PdfRectangle = rectUnion(b.PdfRectangle, word.PdfRectangle)
	if word.fontsize > b.fontsize {
		b.fontsize = word.fontsize
	}
	b.bins[depthIdx] = append(b.bins[depthIdx], word)
	bag.removeWord(word, depthIdx)
}

// removeWord removes `word`from `b`.
// In the current implementation it  removes `word`from `b`.bins[`depthIdx`].
// NOTE: We delete bins as soon as they become empty to save code that calls other wordBag
// functions from having to check for empty bins.
// TODO(peterwilliams97): Find a more efficient way of doing this.
func (b *wordBag) removeWord(word *textWord, depthIdx int) {
	words := removeWord(b.stratum(depthIdx), word)
	if len(words) == 0 {
		delete(b.bins, depthIdx)
	} else {
		b.bins[depthIdx] = words
	}
}

// mergeWordBags merges the bags less than a character width to the left of a bag into that bag.
func mergeWordBags(paraWords []*wordBag) []*wordBag {
	if len(paraWords) <= 1 {
		return paraWords
	}
	if verbose {
		common.Log.Info("mergeWordBags:")
	}
	sort.Slice(paraWords, func(i, j int) bool {
		pi, pj := paraWords[i], paraWords[j]
		ai := pi.Width() * pi.Height()
		aj := pj.Width() * pj.Height()
		if ai != aj {
			return ai > aj
		}
		if pi.Height() != pj.Height() {
			return pi.Height() > pj.Height()
		}
		return i < j
	})
	var merged []*wordBag
	absorbed := map[int]struct{}{}
	for i0 := 0; i0 < len(paraWords); i0++ {
		if _, ok := absorbed[i0]; ok {
			continue
		}
		para0 := paraWords[i0]
		for i1 := i0 + 1; i1 < len(paraWords); i1++ {
			if _, ok := absorbed[i0]; ok {
				continue
			}
			para1 := paraWords[i1]
			r := para0.PdfRectangle
			r.Llx -= para0.fontsize
			if rectContainsRect(r, para1.PdfRectangle) {
				para0.absorb(para1)
				absorbed[i1] = struct{}{}
			}
		}
		merged = append(merged, para0)
	}

	if len(paraWords) != len(merged)+len(absorbed) {
		common.Log.Error("mergeWordBags: %d->%d absorbed=%d",
			len(paraWords), len(merged), len(absorbed))
	}
	return merged
}

// absorb combines the words from `bag` into `b`.
func (b *wordBag) absorb(bag *wordBag) {
	for depthIdx, words := range bag.bins {
		for _, word := range words {
			b.pullWord(bag, word, depthIdx)
		}
	}
}

// depthIndex returns a bin index for depth `depth`.
// The returned depthIdx obeys the following rule.
// depthIdx * depthBinPoints <= depth <= (depthIdx+1) * depthBinPoint
func depthIndex(depth float64) int {
	var depthIdx int
	if depth >= 0 {
		depthIdx = int(depth / depthBinPoints)
	} else {
		depthIdx = int(depth/depthBinPoints) - 1
	}
	return depthIdx
}

// depthIndexes returns the sorted keys of b.bins.
func (b *wordBag) depthIndexes() []int {
	if len(b.bins) == 0 {
		return nil
	}
	indexes := make([]int, len(b.bins))
	i := 0
	for idx := range b.bins {
		indexes[i] = idx
		i++
	}
	sort.Ints(indexes)
	return indexes
}

// sort sorts the word fragments in each bin in `b` in the reading direction.
func (b *wordBag) sort() {
	for _, bin := range b.bins {
		sort.Slice(bin, func(i, j int) bool { return diffReading(bin[i], bin[j]) < 0 })
	}
}

// minDepth returns the minimum depth that word fragments in `b` touch.
func (b *wordBag) minDepth() float64 {
	return b.pageHeight - (b.Ury - b.fontsize)
}

// maxDepth returns the maximum depth that word fragments in `b` touch.
func (b *wordBag) maxDepth() float64 {
	return b.pageHeight - b.Lly
}

// The following functions are used only for logging.

func (b *wordBag) text() string {
	words := b.allWords()
	texts := make([]string, len(words))
	for i, w := range words {
		texts[i] = w.text
	}
	return strings.Join(texts, " ")
}

func (b *wordBag) allWords() []*textWord {
	var wordList []*textWord
	for _, words := range b.bins {
		wordList = append(wordList, words...)
	}
	return wordList
}
