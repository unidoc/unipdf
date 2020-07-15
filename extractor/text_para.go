/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// paraList is a sequence of textPara. We use it so often that it is convenient to have its own
// type so we can have methods on it.
type paraList []*textPara

// textPara is a group of words in a rectangular region of a page that get read together.
// A paragraph in a document might span multiple pages. This is a paragraph fragment on one page.
// textParas can be tables in which case the content is in `table`, otherwise the content is in `lines`.
// textTable cells are textParas so this gives one level of recursion
type textPara struct {
	model.PdfRectangle                    // Bounding box.
	eBBox              model.PdfRectangle // Extended bounding box needed to compute reading order.
	lines              []*textLine        // The lines in the paragraph. (nil for the table case)
	table              *textTable         // The table contained in this region if there is one. nil otherwise
	// The following fields are used for detecting and extracting tables.
	isCell bool // Is this para a cell in a textTable?
	// The unique highest para completely to the left of this that overlaps it in the y-direction, if one exists..
	left *textPara
	// The unique highest para completely to the right of this that overlaps it in the y-direction, if one exists.
	right *textPara
	// The unique highest para completely above this that overlaps it in the x-direction, if one exists.
	above *textPara
	// The unique highest para completely below this that overlaps it in the x-direction, if one exists.
	below *textPara
}

// makeTextPara returns a textPara with bounding rectangle `bbox`.
func makeTextPara(bbox model.PdfRectangle, lines []*textLine) *textPara {
	return &textPara{PdfRectangle: bbox, lines: lines}
}

// String returns a description of `p`.
func (p *textPara) String() string {
	table := ""
	if p.table != nil {
		table = fmt.Sprintf("[%dx%d] ", p.table.w, p.table.h)
	}
	return fmt.Sprintf("%6.2f %s%d lines %q",
		p.PdfRectangle, table, len(p.lines), truncate(p.text(), 50))
}

// depth returns the paragraph's depth. which is the depth of its top line.
// We return the top line depth because textPara depth is used to tell if 2 paras have the same
// depth. English readers compare paragraph depths by their top lines.
func (p *textPara) depth() float64 {
	if len(p.lines) > 0 {
		return p.lines[0].depth
	}
	// Use the top left cell of the table if there is one
	return p.table.get(0, 0).depth()
}

// text is a convenience function that returns the text `p` including tables.
func (p *textPara) text() string {
	w := new(bytes.Buffer)
	p.writeText(w)
	return w.String()
}

// writeText writes the text of `p` including tables to `w`.
func (p *textPara) writeText(w io.Writer) {
	if p.table == nil {
		p.writeCellText(w)
		return
	}
	for y := 0; y < p.table.h; y++ {
		for x := 0; x < p.table.w; x++ {
			cell := p.table.get(x, y)
			if cell == nil {
				w.Write([]byte("\t"))
			} else {
				cell.writeCellText(w)
			}
			w.Write([]byte(" "))
		}
		if y < p.table.h-1 {
			w.Write([]byte("\n"))
		}
	}
}

// toTextMarks creates the TextMarkArray corresponding to the extracted text created by
// paras `p`.writeText().
func (p *textPara) toTextMarks(offset *int) []TextMark {
	if p.table == nil {
		return p.toCellTextMarks(offset)
	}
	var marks []TextMark
	for y := 0; y < p.table.h; y++ {
		for x := 0; x < p.table.w; x++ {
			cell := p.table.get(x, y)
			if cell == nil {
				marks = appendSpaceMark(marks, offset, "\t")
			} else {
				cellMarks := cell.toCellTextMarks(offset)
				marks = append(marks, cellMarks...)
			}
			marks = appendSpaceMark(marks, offset, " ")
		}
		if y < p.table.h-1 {
			marks = appendSpaceMark(marks, offset, "\n")
		}
	}
	return marks
}

// writeCellText writes the text of `p` not including tables to `w`.
func (p *textPara) writeCellText(w io.Writer) {
	for il, line := range p.lines {
		lineText := line.text()
		reduced := doHyphens && line.endsInHyphen() && il != len(p.lines)-1
		if reduced { // Line ending with hyphen. Remove it.
			lineText = removeLastRune(lineText)
		}
		w.Write([]byte(lineText))
		if !(reduced || il == len(p.lines)-1) {
			w.Write([]byte(getSpace(line.depth, p.lines[il+1].depth)))
		}
	}
}

// toCellTextMarks creates the TextMarkArray corresponding to the extracted text created by
// paras `p`.writeCellText().
func (p *textPara) toCellTextMarks(offset *int) []TextMark {
	var marks []TextMark
	for il, line := range p.lines {
		lineMarks := line.toTextMarks(offset)
		reduced := doHyphens && line.endsInHyphen() && il != len(p.lines)-1
		if reduced { // Line ending with hyphen. Remove it.
			lineMarks = removeLastTextMarkRune(lineMarks, offset)
		}
		marks = append(marks, lineMarks...)
		if !(reduced || il == len(p.lines)-1) {
			marks = appendSpaceMark(marks, offset, getSpace(line.depth, p.lines[il+1].depth))
		}
	}
	return marks
}

// removeLastTextMarkRune removes the last rune from `marks`.
func removeLastTextMarkRune(marks []TextMark, offset *int) []TextMark {
	tm := marks[len(marks)-1]
	runes := []rune(tm.Text)
	if len(runes) == 1 {
		marks = marks[:len(marks)-1]
		tm1 := marks[len(marks)-1]
		*offset = tm1.Offset + len(tm1.Text)
	} else {
		text := removeLastRune(tm.Text)
		*offset += len(text) - len(tm.Text)
		tm.Text = text
	}
	return marks
}

// removeLastRune removes the last run from `text`.
func removeLastRune(text string) string {
	runes := []rune(text)
	return string(runes[:len(runes)-1])
}

// getSpace returns the space to insert between lines of depth `depth1` and `depth2`.
// Next line is the same depth so it's the same line as this one in the extracted text
func getSpace(depth1, depth2 float64) string {
	eol := !isZero(depth1 - depth2)
	if eol {
		return "\n"
	}
	return " "
}

// bbox makes textPara implement the `bounded` interface.
func (p *textPara) bbox() model.PdfRectangle {
	return p.PdfRectangle
}

// fontsize return the para's fontsize which we take to be the first line's fontsize.
// Caller must check that `p` has at least one line.
func (p *textPara) fontsize() float64 {
	return p.lines[0].fontsize
}

// removeDuplicates removes duplicate word fragments such as those used for bolding.
func (b *wordBag) removeDuplicates() {
	for _, depthIdx := range b.depthIndexes() {
		if len(b.bins[depthIdx]) == 0 {
			continue
		}
		word := b.bins[depthIdx][0]
		delta := maxDuplicateWordR * word.fontsize
		minDepth := word.depth
		for _, idx := range b.depthBand(minDepth, minDepth+delta) {
			duplicates := map[*textWord]struct{}{}
			words := b.bins[idx]
			for _, w := range words {
				if w != word && w.text == word.text &&
					math.Abs(w.Llx-word.Llx) < delta &&
					math.Abs(w.Urx-word.Urx) < delta &&
					math.Abs(w.Lly-word.Lly) < delta &&
					math.Abs(w.Ury-word.Ury) < delta {
					duplicates[w] = struct{}{}
				}
			}
			if len(duplicates) > 0 {
				i := 0
				for _, w := range words {
					if _, ok := duplicates[w]; !ok {
						words[i] = w
						i++
					}
				}
				b.bins[idx] = words[:len(words)-len(duplicates)]
				if len(b.bins[idx]) == 0 {
					delete(b.bins, idx)
				}
			}
		}
	}
}

// arrangeText arranges the word fragments (textWords) in `b` into lines and words.
// The lines are groups of textWords of similar depths.
// The textWords in each line are sorted in reading order and those that start whole words (as
// opposed to word fragments) have their `newWord` flag set to true.
func (b *wordBag) arrangeText() *textPara {
	b.sort() // Sort the words in `b`'s bins in the reading direction.

	if doRemoveDuplicates {
		b.removeDuplicates()
	}

	var lines []*textLine

	// Build the lines by iterating through the words from top to bottom.
	// In the current implementation, we do this by emptying the word bins in increasing depth order.
	for _, depthIdx := range b.depthIndexes() {
		for !b.empty(depthIdx) {

			// firstWord is the left-most word near the top of the bin with index `depthIdx`. As we
			// are scanning down `b`, this is the  left-most word near the top of the `b`
			firstReadingIdx := b.firstReadingIndex(depthIdx)
			firstWord := b.firstWord(firstReadingIdx)
			// Create a new line.
			line := newTextLine(b, firstReadingIdx)

			// Compute the search range based on `b` first word fontsize.
			fontsize := firstWord.fontsize
			minDepth := firstWord.depth - lineDepthR*fontsize
			maxDepth := firstWord.depth + lineDepthR*fontsize
			maxIntraWordGap := maxIntraWordGapR * fontsize
			maxIntraLineOverlap := maxIntraLineOverlapR * fontsize

			// Find the rest of the words in the line that starts with `firstWord`
			// Search down from `minDepth`, half a line above `firstWord` to `maxDepth`, half a line
			// below `firstWord` for the leftmost word to the right of the last word in `line`.
		remainingWords:
			for {
				var nextWord *textWord // The next word to add to `line` if there is one.
				nextDepthIdx := 0      // nextWord's depthIndex
				// We start with this highest remaining word
				for _, depthIdx := range b.depthBand(minDepth, maxDepth) {
					word := b.highestWord(depthIdx, minDepth, maxDepth)
					if word == nil {
						continue
					}
					gap := gapReading(word, line.words[len(line.words)-1])
					if gap < -maxIntraLineOverlap { // Reverted too far to left. Can't be same line.
						break remainingWords
					}
					if gap > maxIntraWordGap { // Advanced too far too right. Might not be same line.
						continue
					}
					if nextWord != nil && diffReading(word, nextWord) >= 0 { // Not leftmost world
						continue
					}
					nextWord = word
					nextDepthIdx = depthIdx
				}
				if nextWord == nil { // No more words in this line.
					break
				}
				// remove `nextWord` from `b` and append it to `line`.
				line.pullWord(b, nextWord, nextDepthIdx)
			}

			line.markWordBoundaries()
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	sort.Slice(lines, func(i, j int) bool {
		return diffDepthReading(lines[i], lines[j]) < 0
	})

	para := makeTextPara(b.PdfRectangle, lines)

	if verbosePara {
		common.Log.Info("arrangeText !!! para=%s", para.String())
		if verboseParaLine {
			for i, line := range para.lines {
				fmt.Printf("%4d: %s\n", i, line.String())
				if verboseParaWord {
					for j, word := range line.words {
						fmt.Printf("%8d: %s\n", j, word.String())
						for k, mark := range word.marks {
							fmt.Printf("%12d: %s\n", k, mark.String())
						}
					}
				}
			}
		}
	}
	return para
}

// log logs the contents of `paras`.
func (paras paraList) log(title string) {
	if !verbosePage {
		return
	}
	common.Log.Info("%8s: %d paras =======-------=======", title, len(paras))
	for i, para := range paras {
		if para == nil {
			continue
		}
		text := para.text()
		tabl := "  "
		if para.table != nil {
			tabl = fmt.Sprintf("[%dx%d]", para.table.w, para.table.h)
		}
		fmt.Printf("%4d: %6.2f %s %q\n", i, para.PdfRectangle, tabl, truncate(text, 50))
	}
}
