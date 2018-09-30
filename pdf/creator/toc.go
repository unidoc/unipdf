/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import "github.com/unidoc/unidoc/pdf/model/fonts"

// TOC represents a table of contents component. It consists of a paragraph
// heading and a collection of table of contents lines.
type TOC struct {
	// The heading of the table of contents.
	heading *StyledParagraph

	// The lines of the table of contents.
	lines []*TOCLine

	// The style of the number part of new lines.
	lineNumberStyle TextStyle

	// The style of the title part of new lines.
	lineTitleStyle TextStyle

	// The style of the separator part of new lines.
	lineSeparatorStyle TextStyle

	// The style of the page part of new lines.
	linePageStyle TextStyle

	// Positioning: relative/absolute.
	positioning positioning
}

// NewTOC creates a new table of contents.
func NewTOC(title string) *TOC {
	headingStyle := NewTextStyle()
	headingStyle.Font = fonts.NewFontHelveticaBold()
	headingStyle.FontSize = 14

	heading := NewStyledParagraph(title, headingStyle)
	heading.SetEnableWrap(true)
	heading.SetTextAlignment(TextAlignmentLeft)
	heading.SetMargins(0, 0, 0, 5)

	lineStyle := NewTextStyle()

	return &TOC{
		heading:            heading,
		lines:              []*TOCLine{},
		lineNumberStyle:    lineStyle,
		lineTitleStyle:     lineStyle,
		lineSeparatorStyle: lineStyle,
		linePageStyle:      lineStyle,
		positioning:        positionRelative,
	}
}

// Heading returns the heading component of the table of contents.
func (t *TOC) Heading() *StyledParagraph {
	return t.heading
}

// SetHeadings sets the text and the style of the heading
// of the table of contents.
func (t *TOC) SetHeading(text string, style TextStyle) {
	t.heading.Reset(text, style)
}

// Add adds a new line with the default style to the table of contents.
func (t *TOC) Add(number, title, page string, level uint) *TOCLine {
	tl := t.AddLine(NewStyledTOCLine(
		TextChunk{
			Text:  number,
			Style: t.lineNumberStyle,
		},
		TextChunk{
			Text:  title,
			Style: t.lineTitleStyle,
		},
		TextChunk{
			Text:  page,
			Style: t.linePageStyle,
		},
		level,
	))

	tl.Separator.Style = t.lineSeparatorStyle
	return tl
}

// Add adds a new line with the provided style to the table of contents.
func (t *TOC) AddLine(line *TOCLine) *TOCLine {
	if line == nil {
		return nil
	}

	t.lines = append(t.lines, line)
	return line
}

// SetSeparator sets the separator for all the lines of the table of contents.
func (t *TOC) SetLineSeparator(separator string) {
	for _, line := range t.lines {
		line.Separator.Text = separator
	}
}

// SetMargins sets the margins for all the lines of the table of contents.
func (t *TOC) SetLineMargins(left, right, top, bottom float64) {
	for _, line := range t.lines {
		line.SetMargins(left, right, top, bottom)
	}
}

// SetLineNumberStyle sets the style for numbers part of  all the lines
// the table of contents has.
func (t *TOC) SetLineNumberStyle(style TextStyle) {
	t.lineNumberStyle = style
	for _, line := range t.lines {
		line.Number.Style = style
	}
}

// SetLineTitleStyle sets the style for page part of  all the lines
// the table of contents has.
func (t *TOC) SetLineTitleStyle(style TextStyle) {
	t.lineTitleStyle = style
	for _, line := range t.lines {
		line.Title.Style = style
	}
}

// SetLineSeparatorStyle sets the style for separator part of  all the lines
// the table of contents has.
func (t *TOC) SetLineSeparatorStyle(style TextStyle) {
	t.lineSeparatorStyle = style
	for _, line := range t.lines {
		line.Separator.Style = style
	}
}

// SetLinePageStyle sets the style for the page for all the lines
// the table of contents has.
func (t *TOC) SetLinePageStyle(style TextStyle) {
	t.linePageStyle = style
	for _, line := range t.lines {
		line.Page.Style = style
	}
}

// SetLineLevelOffset sets the amount of space an indentation level occupies
// for all the lines the table of contents has.
func (t *TOC) SetLineLevelOffset(levelOffset float64) {
	for _, line := range t.lines {
		line.SetLevelOffset(levelOffset)
	}
}

// GeneratePageBlocks generate the Page blocks.  Multiple blocks are generated if the contents wrap over
// multiple pages.
func (t *TOC) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	// Generate heading blocks
	blocks, ctx, err := t.heading.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}

	// Generate blocks for the table of contents lines
	for _, line := range t.lines {
		newBlocks, c, err := line.GeneratePageBlocks(ctx)
		if err != nil {
			return blocks, ctx, err
		}
		if len(newBlocks) < 1 {
			continue
		}

		// The first block is always appended to the last..
		blocks[len(blocks)-1].mergeBlocks(newBlocks[0])
		blocks = append(blocks, newBlocks[1:]...)

		ctx = c
	}

	if t.positioning.isRelative() {
		// Move back X to same start of line.
		ctx.X = origCtx.X
	}

	if t.positioning.isAbsolute() {
		// If absolute: return original context.
		return blocks, origCtx, nil

	}

	return blocks, ctx, nil
}
