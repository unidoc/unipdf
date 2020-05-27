/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/unidoc/unipdf/v3/model"
)

// textLine repesents words on the same line within a textPara.
type textLine struct {
	serial             int         // Sequence number for debugging.
	model.PdfRectangle             // Bounding box (union of `marks` bounding boxes).
	depth              float64     // Distance from bottom of line to top of page.
	words              []*textWord // Words in this line.
	fontsize           float64
	hyphenated         bool
}

// newTextLine creates a line with font and bbox size of `w`, removes `w` from p.bins[bestWordDepthIdx] and adds it to the line
func newTextLine(p *textStrata, depthIdx int) *textLine {
	words := p.getStratum(depthIdx)
	word := words[0]
	line := textLine{
		serial:       serial.line,
		PdfRectangle: word.PdfRectangle,
		fontsize:     word.fontsize,
		depth:        word.depth,
	}
	serial.line++
	line.moveWord(p, depthIdx, word)
	return &line
}

// String returns a description of `l`.
func (l *textLine) String() string {
	return fmt.Sprintf("serial=%d %.2f %.2f fontsize=%.2f \"%s\"",
		l.serial, l.depth, l.PdfRectangle, l.fontsize, l.text())
}

// bbox makes textLine implementethe `bounded` interface.
func (l *textLine) bbox() model.PdfRectangle {
	return l.PdfRectangle
}

// text returns the extracted text contained in line..
func (l *textLine) text() string {
	var words []string
	for _, w := range l.words {
		words = append(words, w.text())
		if w.spaceAfter {
			words = append(words, " ")
		}
	}
	return strings.Join(words, "")

}

// toTextMarks returns the TextMarks contained in `l`.text().
// `offset` is used to give the TextMarks the correct Offset values.
func (l *textLine) toTextMarks(offset *int) []TextMark {
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
	for _, word := range l.words {
		for _, tm := range word.marks {
			addMark(tm.ToTextMark())
		}
		if word.spaceAfter {
			addSpaceMark(" ")
		}
	}
	return marks
}

// moveWord removes `word` from p.bins[bestWordDepthIdx] and adds it to `l`.
// `l.PdfRectangle` is increased to bound the new word
// `l.fontsize` is the largest of the fontsizes of the words in line
func (l *textLine) moveWord(s *textStrata, depthIdx int, word *textWord) {
	l.words = append(l.words, word)
	l.PdfRectangle = rectUnion(l.PdfRectangle, word.PdfRectangle)
	if word.fontsize > l.fontsize {
		l.fontsize = word.fontsize
	}
	if word.depth > l.depth {
		l.depth = word.depth
	}
	s.removeWord(depthIdx, word)
}

// mergeWordFragments merges the word fragments in the words in `l`.
func (l *textLine) mergeWordFragments() {
	fontsize := l.fontsize
	if len(l.words) > 1 {
		maxGap := maxIntraLineGapR * fontsize
		fontTol := maxIntraWordFontTolR * fontsize
		merged := []*textWord{l.words[0]}

		for _, word := range l.words[1:] {
			lastMerged := merged[len(merged)-1]
			doMerge := false
			if gapReading(word, lastMerged) >= maxGap {
				lastMerged.spaceAfter = true
			} else if lastMerged.font(lastMerged.len()-1) == word.font(0) &&
				math.Abs(lastMerged.fontsize-word.fontsize) < fontTol {
				doMerge = true
			}
			if doMerge {
				lastMerged.absorb(word)
			} else {
				merged = append(merged, word)
			}
		}
		l.words = merged
	}

	// check for hyphen at end of line
	runes := []rune(l.text())
	l.hyphenated = len(runes) >= 4 &&
		unicode.Is(unicode.Hyphen, runes[len(runes)-1]) &&
		!unicode.IsSpace(runes[len(runes)-2])
	// if l.hyphenated {
	// 	// fmt.Fprintf(os.Stderr, "\n%q ", l.text())
	// 	common.Log.Info("### %d %q\n\t%q:%t\n\t%q:%t",
	// 		len(runes), l.text(),
	// 		runes[len(runes)-1], unicode.Is(unicode.Hyphen, runes[len(runes)-1]),
	// 		runes[len(runes)-2], !unicode.IsSpace(runes[len(runes)-2]),
	// 	)
	// }
}
