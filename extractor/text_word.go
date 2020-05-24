/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/unidoc/unipdf/v3/internal/textencoding"
	"github.com/unidoc/unipdf/v3/model"
)

// textWord represents a word. It's a sequence of textMarks that are close enough toghether in the
// reading direction and doesn't have any space textMarks.
type textWord struct {
	serial             int        // Sequence number for debugging.
	model.PdfRectangle            // Bounding box (union of `marks` bounding boxes).
	depth              float64    // Distance from bottom of word to top of page.
	marks              []textMark // Marks in this word.
	fontsize           float64    // Largest fontsize in `marks` w
	spaceAfter         bool
}

// makeTextPage builds a word list from `marks`, the textMarks on a page.
// `pageSize` is used to calculate the words` depths depth on the page
func makeTextWords(marks []textMark, pageSize model.PdfRectangle) []*textWord {
	var words []*textWord
	var cursor *textWord

	// addWord adds `cursor` to `words` and resets it to nil
	addWord := func() {
		if cursor != nil {
			if !isTextSpace(cursor.text()) {
				words = append(words, cursor)
			}
			cursor = nil
		}
	}

	for _, tm := range marks {
		isSpace := isTextSpace(tm.text)
		if cursor == nil && !isSpace {
			cursor = newTextWord([]textMark{tm}, pageSize)
			continue
		}
		if isSpace {
			addWord()
			continue
		}

		depthGap := pageSize.Ury - tm.Lly - cursor.depth
		readingGap := tm.Llx - cursor.Urx
		fontsize := cursor.fontsize

		// These are the conditions for `tm` to be from a new word.
		// - Change in reading position is larger than a space which we guess to be 0.11*fontsize.
		// - Change in reading position is too negative to be just a kerning adjustment.
		// - Change in depth is too large to be just a leading adjustment.
		sameWord := -0.19*fontsize <= readingGap && readingGap <= 0.11*fontsize &&
			math.Abs(depthGap) <= 0.04*fontsize
		if !sameWord {
			addWord()
			cursor = newTextWord([]textMark{tm}, pageSize)
			continue
		}

		cursor.addMark(tm, pageSize)
	}
	addWord()
	return words
}

// newTextWord creates a textWords containing `marks`.
// `pageSize` is used to calculate the word's depth on the page.
func newTextWord(marks []textMark, pageSize model.PdfRectangle) *textWord {
	r := marks[0].PdfRectangle
	fontsize := marks[0].fontsize
	for _, tm := range marks[1:] {
		r = rectUnion(r, tm.PdfRectangle)
		if tm.fontsize > fontsize {
			fontsize = tm.fontsize
		}
	}
	depth := pageSize.Ury - r.Lly

	word := textWord{
		serial:       serial.word,
		PdfRectangle: r,
		marks:        marks,
		depth:        depth,
		fontsize:     fontsize,
	}
	serial.word++
	return &word
}

// String returns a description of `w.
func (w *textWord) String() string {
	return fmt.Sprintf("serial=%d base=%.2f %.2f fontsize=%.2f \"%s\"",
		w.serial, w.depth, w.PdfRectangle, w.fontsize, w.text())
}

func (w *textWord) bbox() model.PdfRectangle {
	return w.PdfRectangle
}

// addMark adds textMark `tm` to word `w`.
// `pageSize` is used to calculate the word's depth on the page.
func (w *textWord) addMark(tm textMark, pageSize model.PdfRectangle) {
	w.marks = append(w.marks, tm)
	w.PdfRectangle = rectUnion(w.PdfRectangle, tm.PdfRectangle)
	if tm.fontsize > w.fontsize {
		w.fontsize = tm.fontsize
	}
	w.depth = pageSize.Ury - w.PdfRectangle.Lly
	if w.depth < 0 {
		panic(w.depth)
	}
}

// len returns the number of runes in `w`.
func (w *textWord) len() int {
	return utf8.RuneCountInString(w.text())
}

func (w *textWord) merge(word *textWord) {
	w.PdfRectangle = rectUnion(w.PdfRectangle, word.PdfRectangle)
	w.marks = append(w.marks, word.marks...)
}

func (w *textWord) text() string {
	var parts []string
	for _, tm := range w.marks {
		for _, r := range tm.text {
			parts = append(parts, textencoding.RuneToString(r))
		}
	}
	return strings.Join(parts, "")
}

// font returns the fontID of the `idx`th rune in text.
// compute on creation? !@#$
func (w *textWord) font(idx int) string {
	numChars := 0
	for _, tm := range w.marks {
		for _, r := range tm.text {
			numChars += len(textencoding.RuneToString(r))
			if numChars > idx {
				return fmt.Sprintf("%s:%.3f", tm.font, tm.fontsize)
			}
		}
	}
	panic("no match")
}

func baseRange(words []*textWord) (minDepth, maxDepth float64) {
	for i, w := range words {
		depth := w.depth
		if i == 0 {
			minDepth = depth
			maxDepth = depth
		} else if depth < minDepth {
			minDepth = depth
		} else if depth > maxDepth {
			maxDepth = depth
		}
	}
	return
}

func removeWord(words []*textWord, word *textWord) []*textWord {
	for i, w := range words {
		if w == word {
			return removeWordAt(words, i)
		}
	}
	panic("word not in words")
}

func removeWordAt(words []*textWord, idx int) []*textWord {
	n := len(words)
	copy(words[idx:], words[idx+1:])
	return words[:n-1]
}
