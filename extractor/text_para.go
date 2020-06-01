/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"unicode"

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
	table              *textTable
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
	w := new(bytes.Buffer)
	p.writeText(w)
	return w.String()
}

// writeText writes the text of `p` including tables to `w`.
func (p *textPara) writeText(w io.Writer) {
	if p.table != nil {
		for y := 0; y < p.table.h; y++ {
			for x := 0; x < p.table.w; x++ {
				cell := p.table.cells[y*p.table.w+x]
				cell.writeCellText(w)
				w.Write([]byte(" "))
			}
			w.Write([]byte("\n"))
		}
	} else {
		p.writeCellText(w)
		w.Write([]byte("\n"))
	}
}

// writeCellText writes the text of `p` not including tables to `w`.
func (p *textPara) writeCellText(w io.Writer) {
	// w := new(bytes.Buffer)
	para := p
	for il, line := range para.lines {
		s := line.text()
		reduced := false
		if doHyphens {
			if line.hyphenated && il != len(para.lines)-1 {
				// Line ending with hyphen. Remove it.
				runes := []rune(s)
				s = string(runes[:len(runes)-1])
				reduced = true
			}
		}
		w.Write([]byte(s))
		if reduced {
			// We removed the hyphen from the end of the line so we don't need a line ending.
			continue
		}
		if il < len(para.lines)-1 && isZero(line.depth-para.lines[il+1].depth) {
			// Next line is the same depth so it's the same line as this one in the extracted text
			w.Write([]byte(" "))
			continue
		}
		if il < len(para.lines)-1 {
			w.Write([]byte("\n"))
		}
	}
}

// toTextMarks creates the TextMarkArray corresponding to the extracted text created by
// paras `p`.writeText().
func (p *textPara) toTextMarks(offset *int) []TextMark {
	var marks []TextMark
	addMark := func(mark TextMark) {
		mark.Offset = *offset
		marks = append(marks, mark)
		*offset += len(mark.Text)
	}
	addSpaceMark := func(spaceChar string) {
		mark := spaceMark
		mark.Text = spaceChar
		addMark(mark)
	}
	if p.table != nil {
		for y := 0; y < p.table.h; y++ {
			for x := 0; x < p.table.w; x++ {
				cell := p.table.cells[y*p.table.w+x]
				cellMarks := cell.toCellTextMarks(offset)
				marks = append(marks, cellMarks...)
				addSpaceMark(" ")
			}
			addSpaceMark("\n")
		}
	} else {
		marks = p.toCellTextMarks(offset)
		addSpaceMark("\n")
	}
	return marks
}

// toTextMarks creates the TextMarkArray corresponding to the extracted text created by
// paras `paras`.writeCellText().
func (p *textPara) toCellTextMarks(offset *int) []TextMark {
	var marks []TextMark
	addMark := func(mark TextMark) {
		mark.Offset = *offset
		marks = append(marks, mark)
		*offset += len(mark.Text)
	}
	addSpaceMark := func(spaceChar string) {
		mark := spaceMark
		mark.Text = spaceChar
		addMark(mark)
	}
	para := p

	for il, line := range para.lines {
		lineMarks := line.toTextMarks(offset)
		marks = append(marks, lineMarks...)
		reduced := false
		if doHyphens {
			if line.hyphenated && il != len(para.lines)-1 {
				tm := marks[len(marks)-1]
				r := []rune(tm.Text)
				if unicode.IsSpace(r[len(r)-1]) {
					panic(tm)
				}
				if len(r) == 1 {
					marks = marks[:len(marks)-1]
					*offset = marks[len(marks)-1].Offset + len(marks[len(marks)-1].Text)
				} else {
					s := string(r[:len(r)-1])
					*offset += len(s) - len(tm.Text)
					tm.Text = s
				}
				reduced = true
			}
		}
		if reduced {
			continue
		}
		if il < len(para.lines)-1 && isZero(line.depth-para.lines[il+1].depth) {
			// Next line is the same depth so it's the same line as this one in the extracted text
			addSpaceMark(" ")
			continue
		}
		if il < len(para.lines)-1 {
			addSpaceMark("\n")
		}
	}

	addSpaceMark("\n")

	return marks
}

// bbox makes textPara implement the `bounded` interface.
func (p *textPara) bbox() model.PdfRectangle {
	return p.PdfRectangle
}

// fontsize return the para's fontsize which we take to be the first line's fontsize
func (p *textPara) fontsize() float64 {
	if len(p.lines) == 0 {
		panic(p)
	}
	return p.lines[0].fontsize
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
	if len(para.lines) == 0 {
		panic(para)
	}
	return para
}
