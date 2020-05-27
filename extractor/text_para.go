/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/unidoc/unipdf/v3/model"
)

// textPara is a group of words in a rectangular region of a page that get read together.
// An peragraph in a document might span multiple pages. This is the paragraph framgent on one page.
// We start by finding paragraph regions on a page, then we break the words into the textPara into
// textLines.
type textPara struct {
	serial             int                // Sequence number for debugging.
	model.PdfRectangle                    // Bounding box.
	eBBox              model.PdfRectangle // Extented ounding box needed to compute reading order.
	lines              []*textLine        // Paragraph text gets broken into lines.
}

// newTextPara returns a textPara with the same bouding rectangle as `strata`.
func newTextPara(strata *textStrata) *textPara {
	para := textPara{
		serial:       serial.para,
		PdfRectangle: strata.PdfRectangle,
	}
	serial.para++
	return &para
}

// String returns a description of `p`.
func (p *textPara) String() string {
	return fmt.Sprintf("serial=%d %.2f %d lines\n%s\n-------------",
		p.serial, p.PdfRectangle, len(p.lines), p.text())
}

// text returns the text  of the lines in `p`.
func (p *textPara) text() string {
	parts := make([]string, len(p.lines))
	for i, line := range p.lines {
		parts[i] = line.text()
	}
	return strings.Join(parts, "\n")
}

// bbox makes textPara implement the `bounded` interface.
func (p *textPara) bbox() model.PdfRectangle {
	return p.PdfRectangle
}

// composePara builds a textPara from the words in `strata`.
// It does this by arranging the words in `strata` into lines.
func composePara(strata *textStrata) *textPara {
	para := newTextPara(strata)

	// build the lines
	for _, depthIdx := range strata.depthIndexes() {
		for !strata.empty(depthIdx) {

			// words[0] is the leftmost word from bins near `depthIdx`.
			firstReadingIdx := strata.firstReadingIndex(depthIdx)
			// create a new line
			words := strata.getStratum(firstReadingIdx)
			word0 := words[0]
			line := newTextLine(strata, firstReadingIdx)
			lastWord := words[0]

			// compute the search range
			// this is based on word0, the first word in the `firstReadingIdx` bin.
			fontSize := strata.fontsize
			minDepth := word0.depth - lineDepthR*fontSize
			maxDepth := word0.depth + lineDepthR*fontSize
			maxIntraWordGap := maxIntraWordGapR * fontSize

		remainingWords:
			// find the rest of the words in this line
			for {
				// Search for `leftWord`, the left-most word w: minDepth <= w.depth <= maxDepth.
				var leftWord *textWord
				leftDepthIdx := 0
				for _, depthIdx := range strata.depthBand(minDepth, maxDepth) {
					words := strata.stratumBand(depthIdx, minDepth, maxDepth)
					if len(words) == 0 {
						continue
					}
					word := words[0]
					gap := gapReading(word, lastWord)
					if gap < -maxIntraLineOverlapR*fontSize {
						break remainingWords
					}
					// No `leftWord` or `word` to the left of `leftWord`.
					if gap < maxIntraWordGap {
						if leftWord == nil || diffReading(word, leftWord) < 0 {
							leftDepthIdx = depthIdx
							leftWord = word
						}
					}
				}
				if leftWord == nil {
					break
				}

				// remove `leftWord` from `strata`[`leftDepthIdx`], and append it to `line`.
				line.moveWord(strata, leftDepthIdx, leftWord)
				lastWord = leftWord
				// // TODO(peterwilliams97): Replace lastWord with line.words[len(line.words)-1] ???
				// if lastWord != line.words[len(line.words)-1] {
				// 	panic("ddd")
				// }
			}

			line.mergeWordFragments()
			// add the line
			para.lines = append(para.lines, line)
		}
	}

	sort.Slice(para.lines, func(i, j int) bool {
		return diffDepthReading(para.lines[i], para.lines[j]) < 0
	})
	return para
}
