/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"strings"
	"unicode"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// TextChunk represents a chunk of text along with a particular style.
type TextChunk struct {
	// The text that is being rendered in the PDF.
	Text string

	// The style of the text being rendered.
	Style TextStyle

	// Text chunk annotation.
	annotation *model.PdfAnnotation

	// Internally used in order to skip processing the annotation
	// if it has already been processed by the parent component.
	annotationProcessed bool
}

// NewTextChunk returns a new text chunk instance.
func NewTextChunk(text string, style TextStyle) *TextChunk {
	return &TextChunk{
		Text:  text,
		Style: style,
	}
}

// SetAnnotation sets a annotation on a TextChunk.
func (tc *TextChunk) SetAnnotation(annotation *model.PdfAnnotation) {
	tc.annotation = annotation
}

// Wrap wraps the text of the chunk into lines based on its style and the
// specified width.
func (tc *TextChunk) Wrap(width float64) ([]string, error) {
	if int(width) <= 0 {
		return []string{tc.Text}, nil
	}

	var lines []string
	var line []rune
	var lineWidth float64
	var widths []float64

	style := tc.Style
	runes := []rune(tc.Text)

	for _, r := range runes {
		// Move to the next line due to newline wrapping (LF).
		if r == '\u000A' {
			lines = append(lines, strings.TrimRightFunc(string(line), unicode.IsSpace)+string(r))
			line = nil
			lineWidth = 0
			widths = nil
			continue
		}

		metrics, found := style.Font.GetRuneMetrics(r)
		if !found {
			common.Log.Debug("ERROR: Rune char metrics not found! rune=0x%04x=%c font=%s %#q",
				r, r, style.Font.BaseFont(), style.Font.Subtype())
			common.Log.Trace("Font: %#v", style.Font)
			common.Log.Trace("Encoder: %#v", style.Font.Encoder())
			return nil, errors.New("glyph char metrics missing")
		}

		w := style.FontSize * metrics.Wx
		charWidth := w + style.CharSpacing*1000.0
		if lineWidth+w > width*1000.0 {
			// Goes out of bounds. Break on the character.
			idx := -1
			for i := len(line) - 1; i >= 0; i-- {
				if line[i] == ' ' {
					idx = i
					break
				}
			}
			if idx > 0 {
				// Back up to last space.
				lines = append(lines, strings.TrimRightFunc(string(line[0:idx+1]), unicode.IsSpace))

				// Remainder of line.
				line = append(line[idx+1:], r)
				widths = append(widths[idx+1:], charWidth)

				lineWidth = 0
				for _, width := range widths {
					lineWidth += width
				}
			} else {
				lines = append(lines, strings.TrimRightFunc(string(line), unicode.IsSpace))
				line = []rune{r}
				widths = []float64{charWidth}
				lineWidth = charWidth
			}
		} else {
			line = append(line, r)
			lineWidth += charWidth
			widths = append(widths, charWidth)
		}
	}
	if len(line) > 0 {
		lines = append(lines, string(line))
	}

	return lines, nil
}

// Fit fits the chunk into the specified bounding box, cropping off the
// remainder in a new chunk, if it exceeds the specified dimensions.
// NOTE: The method assumes a line height of 1.0. In order to account for other
// line height values, the passed in height must be divided by the line height:
// height = height / lineHeight
func (tc *TextChunk) Fit(width, height float64) (*TextChunk, error) {
	lines, err := tc.Wrap(width)
	if err != nil {
		return nil, err
	}

	fit := int(height / tc.Style.FontSize)
	if fit >= len(lines) {
		return nil, nil
	}
	lf := "\u000A"
	tc.Text = strings.Replace(strings.Join(lines[:fit], " "), lf+" ", lf, -1)

	remainder := strings.Replace(strings.Join(lines[fit:], " "), lf+" ", lf, -1)
	return NewTextChunk(remainder, tc.Style), nil
}

// newExternalLinkAnnotation returns a new external link annotation.
func newExternalLinkAnnotation(url string) *model.PdfAnnotation {
	annotation := model.NewPdfAnnotationLink()

	// Set border style.
	bs := model.NewBorderStyle()
	bs.SetBorderWidth(0)
	annotation.BS = bs.ToPdfObject()

	// Set link destination.
	action := model.NewPdfActionURI()
	action.URI = core.MakeString(url)
	annotation.SetAction(action.PdfAction)

	return annotation.PdfAnnotation
}

// newExternalLinkAnnotation returns a new internal link annotation.
func newInternalLinkAnnotation(page int64, x, y, zoom float64) *model.PdfAnnotation {
	annotation := model.NewPdfAnnotationLink()

	// Set border style.
	bs := model.NewBorderStyle()
	bs.SetBorderWidth(0)
	annotation.BS = bs.ToPdfObject()

	// Set link destination.
	if page < 0 {
		page = 0
	}

	annotation.Dest = core.MakeArray(
		core.MakeInteger(page),
		core.MakeName("XYZ"),
		core.MakeFloat(x),
		core.MakeFloat(y),
		core.MakeFloat(zoom),
	)

	return annotation.PdfAnnotation
}

// copyLinkAnnotation returns a new link annotation based on an existing one.
func copyLinkAnnotation(link *model.PdfAnnotationLink) *model.PdfAnnotationLink {
	if link == nil {
		return nil
	}

	annotation := model.NewPdfAnnotationLink()
	annotation.BS = link.BS
	annotation.A = link.A
	if action, err := link.GetAction(); err == nil && action != nil {
		annotation.SetAction(action)
	}

	if annotDest, ok := link.Dest.(*core.PdfObjectArray); ok {
		annotation.Dest = core.MakeArray(annotDest.Elements()...)
	}

	return annotation
}
