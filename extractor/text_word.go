/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/text/unicode/norm"
)

// textWord represents a word fragment.
// makeTextWords() shows how textWords are created.
// We don't see whole words until textWords are eventually sorted into textLines  in
// wordBag.arrangeText(). textLines are slices of textWord that define whole words by the
//  newWord marker on those fragments that start whole words.
//  - A textLine is the textWords at similar depths sorted in reading order.
//  - All textWords, w, in the textLine that start whole words have w.newWord = true
type textWord struct {
	model.PdfRectangle             // Bounding box (union of `marks` bounding boxes).
	depth              float64     // Distance from bottom of this word to the top of the page.
	text               string      // The word fragment text.
	marks              []*textMark // Marks in this word.
	fontsize           float64     // Largest fontsize in the word.
	newWord            bool        // Is this word fragment the start of  a new word?
}

// makeTextPage combines `marks`, the textMarks on a page, into word fragments.
// `pageSize` is used to calculate the words` depths depth on the page.
// Algorithm:
//  1. `marks` are in the order they were rendered in the PDF.
//  2. Successive marks are combined into a word fragment unless
//      One mark is a space character.
//      They are separated by more than maxWordAdvanceR*fontsize in the reading direction
//      They are not within the location allowed by horizontal and vertical variations allowed by
//       reasonable kerning and leading.
// TODO(peterwilliams97): Check for overlapping textWords for cases such as diacritics, bolding by
//                       repeating and others.
func makeTextWords(marks []*textMark, pageSize model.PdfRectangle) []*textWord {
	var words []*textWord // The words.
	var newWord *textWord // The word being built.

	// addNewWord adds `newWord` to `words` and resets `newWord` to nil.
	addNewWord := func() {
		if newWord != nil {
			text := newWord.computeText()
			if !isTextSpace(text) {
				newWord.text = text
				words = append(words, newWord)
			}
			newWord = nil
		}
	}

	for _, tm := range marks {
		if doCombineDiacritics && newWord != nil && len(newWord.marks) > 0 {
			// Combine diacritic marks into neighbourimg non-diacritics marks.
			prev := newWord.marks[len(newWord.marks)-1]
			text, isDiacritic := combiningDiacritic(tm.text)
			prevText, prevDiacritic := combiningDiacritic(prev.text)
			if isDiacritic && !prevDiacritic && prev.inDiacriticArea(tm) {
				newWord.addDiacritic(text)
				continue
			}
			if prevDiacritic && !isDiacritic && tm.inDiacriticArea(prev) {
				// If the previous mark was the diacritic, merge it into this mark and re-append it
				newWord.marks = newWord.marks[:len(newWord.marks)-1]
				newWord.appendMark(tm, pageSize)
				newWord.addDiacritic(prevText)
				continue
			}
		}

		// Check for spaces between words.
		isSpace := isTextSpace(tm.text)
		if isSpace {
			addNewWord()
			continue
		}

		if newWord == nil && !isSpace {
			newWord = newTextWord([]*textMark{tm}, pageSize)
			continue
		}

		fontsize := newWord.fontsize
		depthGap := math.Abs(getDepth(pageSize, tm)-newWord.depth) / fontsize
		readingGap := gapReading(tm, newWord) / fontsize

		// These are the conditions for `tm` to be from a new word.
		// - Gap between words in reading position is larger than a space.
		// - Change in reading position is too negative to be just a kerning adjustment.
		// - Change in depth is too large to be just a leading adjustment.
		if readingGap >= maxWordAdvanceR || !(-maxKerningR <= readingGap && depthGap <= maxLeadingR) {
			addNewWord()
			newWord = newTextWord([]*textMark{tm}, pageSize)
			continue
		}
		newWord.appendMark(tm, pageSize)
	}
	addNewWord()

	return words
}

// newTextWord creates a textWords containing `marks`.
// `pageSize` is used to calculate the word's depth on the page.
func newTextWord(marks []*textMark, pageSize model.PdfRectangle) *textWord {
	r := marks[0].PdfRectangle
	fontsize := marks[0].fontsize
	for _, tm := range marks[1:] {
		r = rectUnion(r, tm.PdfRectangle)
		if tm.fontsize > fontsize {
			fontsize = tm.fontsize
		}
	}

	return &textWord{
		PdfRectangle: r,
		marks:        marks,
		depth:        pageSize.Ury - r.Lly,
		fontsize:     fontsize,
	}
}

// String returns a description of `w`.
func (w *textWord) String() string {
	return fmt.Sprintf("%.2f %6.2f fontsize=%.2f \"%s\"",
		w.depth, w.PdfRectangle, w.fontsize, w.text)
}

// bbox makes textWord implement the `bounded` interface.
func (w *textWord) bbox() model.PdfRectangle {
	return w.PdfRectangle
}

// appendMark adds textMark `tm` to  `w`.
// `pageSize` is used to calculate the word's depth on the page.
func (w *textWord) appendMark(tm *textMark, pageSize model.PdfRectangle) {
	w.marks = append(w.marks, tm)
	w.PdfRectangle = rectUnion(w.PdfRectangle, tm.PdfRectangle)
	if tm.fontsize > w.fontsize {
		w.fontsize = tm.fontsize
	}
	w.depth = pageSize.Ury - w.PdfRectangle.Lly
}

// addDiacritic adds combining diacritic `text` `tm` to `w`.
// It adds the diacritic to the last mark and doesn't update the size
func (w *textWord) addDiacritic(text string) {
	lastMark := w.marks[len(w.marks)-1]
	lastMark.text = lastMark.text + text
	lastMark.text = norm.NFKC.String(lastMark.text)
}

// absorb combines `word` into `w`.
func (w *textWord) absorb(word *textWord) {
	w.PdfRectangle = rectUnion(w.PdfRectangle, word.PdfRectangle)
	w.marks = append(w.marks, word.marks...)
}

// text returns the text in `w`.
func (w *textWord) computeText() string {
	texts := make([]string, len(w.marks))
	for i, tm := range w.marks {
		texts[i] = tm.text
	}
	return strings.Join(texts, "")
}

// toTextMarks returns the TextMarks contained in `w`.text().
// `offset` is used to give the TextMarks the correct Offset values.
func (w *textWord) toTextMarks(offset *int) []TextMark {
	var marks []TextMark
	for _, tm := range w.marks {
		marks = appendTextMark(marks, offset, tm.ToTextMark())
	}
	return marks
}

// removeWord returns `words` with `word` removed.
// Caller must check that `words` contains `word`,
// TODO(peterwilliams97): Optimize
func removeWord(words []*textWord, word *textWord) []*textWord {
	for i, w := range words {
		if w == word {
			return removeWordAt(words, i)
		}
	}
	common.Log.Error("removeWord: words doesn't contain word=%s", word)
	return nil
}

// removeWord returns `words` with `words[idx]` removed.
func removeWordAt(words []*textWord, idx int) []*textWord {
	n := len(words)
	copy(words[idx:], words[idx+1:])
	return words[:n-1]
}
