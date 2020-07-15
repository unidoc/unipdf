/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/unidoc/unipdf/v3/model"
)

// textLine repesents words on the same line within a textPara.
type textLine struct {
	model.PdfRectangle             // Bounding box (union of `marks` bounding boxes).
	depth              float64     // Distance from bottom of line to top of page.
	words              []*textWord // Words in this line.
	fontsize           float64     // Largest word font size.
}

// newTextLine creates a line with the font and bbox size of the first word in `b`, removes the word
// from `b` and adds it to the line.
func newTextLine(b *wordBag, depthIdx int) *textLine {
	word := b.firstWord(depthIdx)
	line := textLine{
		PdfRectangle: word.PdfRectangle,
		fontsize:     word.fontsize,
		depth:        word.depth,
	}
	line.pullWord(b, word, depthIdx)
	return &line
}

// String returns a description of `l`.
func (l *textLine) String() string {
	return fmt.Sprintf("%.2f %6.2f fontsize=%.2f \"%s\"",
		l.depth, l.PdfRectangle, l.fontsize, l.text())
}

// bbox makes textLine implement the `bounded` interface.
func (l *textLine) bbox() model.PdfRectangle {
	return l.PdfRectangle
}

// text returns the extracted text contained in line.
func (l *textLine) text() string {
	var words []string
	for _, w := range l.words {
		if w.newWord {
			words = append(words, " ")
		}
		words = append(words, w.text)
	}
	return strings.Join(words, "")
}

// toTextMarks returns the TextMarks contained in `l`.text().
// `offset` is used to give the TextMarks the correct Offset values.
func (l *textLine) toTextMarks(offset *int) []TextMark {
	var marks []TextMark
	for _, w := range l.words {
		if w.newWord {
			marks = appendSpaceMark(marks, offset, " ")
		}
		wordMarks := w.toTextMarks(offset)
		marks = append(marks, wordMarks...)
	}
	return marks
}

// pullWord removes `word` from bag and appends it to `l`.
func (l *textLine) pullWord(bag *wordBag, word *textWord, depthIdx int) {
	l.appendWord(word)
	bag.removeWord(word, depthIdx)
}

// appendWord appends `word` to `l`.
// `l.PdfRectangle` is increased to bound the new word.
// `l.fontsize` is the largest of the fontsizes of the words in line.
func (l *textLine) appendWord(word *textWord) {
	l.words = append(l.words, word)
	l.PdfRectangle = rectUnion(l.PdfRectangle, word.PdfRectangle)
	if word.fontsize > l.fontsize {
		l.fontsize = word.fontsize
	}
	if word.depth > l.depth {
		l.depth = word.depth
	}
}

// markWordBoundaries marks the word fragments that are the first fragments in whole words.
func (l *textLine) markWordBoundaries() {
	maxGap := maxIntraLineGapR * l.fontsize
	for i, w := range l.words[1:] {
		if gapReading(w, l.words[i]) >= maxGap {
			w.newWord = true
		}
	}
}

// endsInHyphen attempts to detect words that are split between lines
// IT currently returns true if `l` ends in a hyphen and its last minHyphenation runes don't coataib
// a space.
// TODO(peterwilliams97): Figure out a better heuristic
func (l *textLine) endsInHyphen() bool {
	// Computing l.text() is a little expensive so we filter out simple cases first.
	lastWord := l.words[len(l.words)-1]
	runes := []rune(lastWord.text)
	if !unicode.Is(unicode.Hyphen, runes[len(runes)-1]) {
		return false
	}
	if lastWord.newWord && endsInHyphen(runes) {
		return true
	}
	return endsInHyphen([]rune(l.text()))
}

// endsInHyphen returns true if `runes` ends with a hyphenated word.
func endsInHyphen(runes []rune) bool {
	return len(runes) >= minHyphenation &&
		unicode.Is(unicode.Hyphen, runes[len(runes)-1]) &&
		!unicode.IsSpace(runes[len(runes)-2])
}
