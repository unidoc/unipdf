/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"math"
	"sort"

	"github.com/unidoc/unipdf/v3/model"
)

// textStrata is a list of word bings arranged by their depth on a page.
// The words in each bin are sorted in reading order.
type textStrata struct {
	serial             int                 // Sequence number for debugging.
	model.PdfRectangle                     // Bounding box (union of words' in bins bounding boxes).
	bins               map[int][]*textWord // bins[n] = w: (n-1)*depthBinPoints <= w.depth < (n-1)*depthBinPoints
	pageHeight         float64
	fontsize           float64
}

// makeTextStrata builds a textStrata from `words` but putting the words into the appropriate
// depth bins.
func makeTextStrata(words []*textWord, pageHeight float64) *textStrata {
	s := newTextStrata(pageHeight)
	for _, w := range words {
		depthIdx := depthIndex(w.depth)
		s.bins[depthIdx] = append(s.bins[depthIdx], w)
	}
	s.sort()
	return s
}

func newTextStrata(pageHeight float64) *textStrata {
	bins := textStrata{
		serial:       serial.bins,
		bins:         map[int][]*textWord{},
		PdfRectangle: model.PdfRectangle{Urx: -1.0, Ury: -1.0},
		pageHeight:   pageHeight,
	}
	serial.bins++
	return &bins
}

// String returns a description of `s`.
func (s *textStrata) String() string {
	var texts []string
	for _, depthIdx := range s.depthIndexes() {
		words, _ := s.bins[depthIdx]
		for _, w := range words {
			texts = append(texts, w.text())
		}
	}
	return fmt.Sprintf("serial=%d %d %q", s.serial, len(texts), texts)
}

// sort sorts the words in each in `s` in the reading direction.
func (s *textStrata) sort() {
	for _, bin := range s.bins {
		sort.Slice(bin, func(i, j int) bool { return diffReading(bin[i], bin[j]) < 0 })
	}
}

func (s *textStrata) minDepth() float64 {
	return s.pageHeight - s.Ury
}

func (s *textStrata) maxDepth() float64 {
	return s.pageHeight - s.Lly
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

func depthBand(depthIdx int) (float64, float64) {
	minDepth := float64(depthIdx) * depthBinPoints
	maxDepth := float64(depthIdx+1) * depthBinPoints
	return minDepth, maxDepth
}

// depthIndexes returns the sorted keys of s.bins.
func (s *textStrata) depthIndexes() []int {
	indexes := make([]int, len(s.bins))
	i := 0
	for idx := range s.bins {
		indexes[i] = idx
		i++
	}
	sort.Ints(indexes)
	return indexes
}

// Variation in line depth as a fraction of font size. +lineDepthR for subscripts, -lineDepthR for
// superscripts
const lineDepthR = 0.5

// scanBand scans the bins for words
// w: `minDepth` <= w.depth <= `maxDepth` &&  // in the depth diraction
//    `readingOverlap`(`para`, w) &&  in the reading directon
//     math.Abs(w.fontsize-fontsize) > `fontTol`*fontsize // font size tolerance
// and applies `moveWord`(depthIdx, s,para w) to them.
// If `detectOnly` is true, don't appy moveWord.
// If `freezeDepth` is trus, don't update minDepth and maxDepth in scan as words are added/
func (s *textStrata) scanBand(para *textStrata,
	readingOverlap func(para *textStrata, word *textWord) bool,
	minDepth, maxDepth, fontTol float64,
	detectOnly, freezeDepth bool) int {
	fontsize := para.fontsize
	lineDepth := lineDepthR * fontsize
	n := 0
	for _, depthIdx := range s.depthBand(minDepth-lineDepth, maxDepth+lineDepth) {
		for _, word := range s.bins[depthIdx] {
			if !(minDepth-lineDepth <= word.depth && word.depth <= maxDepth+lineDepth) {
				continue
			}
			if !readingOverlap(para, word) {
				continue
			}
			if fontTol > 0 && math.Abs(word.fontsize-fontsize) > fontTol*fontsize {
				continue
			}
			if !detectOnly {
				moveWord(depthIdx, s, para, word)
			}
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
	return n
}

// stratumBand returns the words in s.bins[depthIdx] w: minDepth <= w.depth <= maxDepth.
func (s *textStrata) stratumBand(depthIdx int, minDepth, maxDepth float64) []*textWord {
	var words []*textWord
	for _, word := range s.bins[depthIdx] {
		if minDepth <= word.depth && word.depth <= maxDepth {
			words = append(words, word)
		}
	}
	return words
}

// depthBand returns the indexes of the bins with depth: `minDepth` <= depth <= `maxDepth`.
func (s *textStrata) depthBand(minDepth, maxDepth float64) []int {
	return s.depthRange(s.getDepthIdx(minDepth), s.getDepthIdx(maxDepth))
}

// depthRange returns the sorted keys of s.bins for depths indexes [`minDepth`,`maxDepth`).
func (s *textStrata) depthRange(minDepthIdx, maxDepthIdx int) []int {
	indexes := s.depthIndexes()
	var rangeIndexes []int
	for _, depthIdx := range indexes {
		if minDepthIdx <= depthIdx && depthIdx <= maxDepthIdx {
			rangeIndexes = append(rangeIndexes, depthIdx)
		}
	}
	return rangeIndexes
}

// firstReadingIndex returns the index of the depth bin that starts with that word with the smallest
// reading direction value in the depth region `minDepthIndex` < depth <= minDepthIndex+ 4*fontsize
// This avoids choosing a bin that starts with a superscript word.
func (s *textStrata) firstReadingIndex(minDepthIdx int) int {
	firstReadingIdx := minDepthIdx
	firstReadingWords := s.getStratum(firstReadingIdx)
	fontsize := firstReadingWords[0].fontsize
	minDepth := float64(minDepthIdx+1) * depthBinPoints
	for _, depthIdx := range s.depthBand(minDepth, minDepth+4*fontsize) {
		words := s.getStratum(depthIdx)
		if diffReading(words[0], firstReadingWords[0]) < 0 {
			firstReadingIdx = depthIdx
			firstReadingWords = s.getStratum(firstReadingIdx)
		}
	}
	return firstReadingIdx
}

// getDepthIdx returns the index into `s.bins` for non-reading axis value `depth`.
func (s *textStrata) getDepthIdx(depth float64) int {
	depthIdx, minIdx, maxIdx := -101, -101, -101
	indexes := s.depthIndexes()
	if len(indexes) > 0 {
		depthIdx = depthIndex(depth)
		minIdx = indexes[0]
		maxIdx = indexes[len(indexes)-1]
		if depthIdx < minIdx {
			depthIdx = minIdx
		}
		if depthIdx > maxIdx {
			depthIdx = maxIdx
		}
	}
	return depthIdx
}

func (s *textStrata) empty(depthIdx int) bool {
	_, ok := s.bins[depthIdx]
	return !ok
}

// getStratum returns a copy of `p`.bins[`depthIdx`].
// getStratum is guaranteed to return a non-nil value (!@#$ Will need to check it is called with valid index)
// NOTE: We need to return a copy because remove() and other functions manipulate the array
// underlying the slice.
func (s *textStrata) getStratum(depthIdx int) []*textWord {
	words := s.bins[depthIdx]
	if words == nil {
		panic(depthIdx)
	}
	dup := make([]*textWord, len(words))
	copy(dup, words)
	return dup
}

// moveWord moves `word` from 'page'[`depthIdx`] to 'para'[`depthIdx`].
func moveWord(depthIdx int, page, para *textStrata, word *textWord) {
	if para.Llx > para.Urx {
		para.PdfRectangle = word.PdfRectangle
	} else {
		para.PdfRectangle = rectUnion(para.PdfRectangle, word.PdfRectangle)
	}
	if word.fontsize > para.fontsize {
		para.fontsize = word.fontsize
	}
	para.bins[depthIdx] = append(para.bins[depthIdx], word)
	page.removeWord(depthIdx, word)
}

// removeWord removes `word`from `s`.bins[`depthIdx`].
// !@#$ Find a more efficient way of doing this.
func (s *textStrata) removeWord(depthIdx int, word *textWord) {
	words := removeWord(s.getStratum(depthIdx), word)
	if len(words) == 0 {
		delete(s.bins, depthIdx)
	} else {
		s.bins[depthIdx] = words
	}
}
