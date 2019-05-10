/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// TOC represents a table of contents component.
// It consists of a paragraph heading and a collection of
// table of contents lines.
// The representation of a table of contents line is as follows:
//       [number] [title]      [separator] [page]
// e.g.: Chapter1 Introduction ........... 1
type TOC struct {
	// The heading of the table of contents.
	heading *StyledParagraph

	// The lines of the table of contents.
	lines []*TOCLine

	// The style of the number part of new TOC lines.
	lineNumberStyle TextStyle

	// The style of the title part of new TOC lines.
	lineTitleStyle TextStyle

	// The style of the separator part of new TOC lines.
	lineSeparatorStyle TextStyle

	// The style of the page part of new TOC lines.
	linePageStyle TextStyle

	// The separator for new TOC lines.
	lineSeparator string

	// The amount of space an indentation level occupies in a TOC line.
	lineLevelOffset float64

	// The margins of new TOC lines.
	lineMargins margins

	// Positioning: relative/absolute.
	positioning positioning

	// Default style used for internal operations.
	defaultStyle TextStyle

	// Specifies if the TOC displays links.
	showLinks bool
}

// newTOC creates a new table of contents.
func newTOC(title string, style, styleHeading TextStyle) *TOC {
	headingStyle := styleHeading
	headingStyle.FontSize = 14

	heading := newStyledParagraph(headingStyle)
	heading.SetEnableWrap(true)
	heading.SetTextAlignment(TextAlignmentLeft)
	heading.SetMargins(0, 0, 0, 5)

	chunk := heading.Append(title)
	chunk.Style = headingStyle

	return &TOC{
		heading:            heading,
		lines:              []*TOCLine{},
		lineNumberStyle:    style,
		lineTitleStyle:     style,
		lineSeparatorStyle: style,
		linePageStyle:      style,
		lineSeparator:      ".",
		lineLevelOffset:    10,
		lineMargins:        margins{0, 0, 2, 2},
		positioning:        positionRelative,
		defaultStyle:       style,
		showLinks:          true,
	}
}

// Heading returns the heading component of the table of contents.
func (t *TOC) Heading() *StyledParagraph {
	return t.heading
}

// Lines returns all the lines the table of contents has.
func (t *TOC) Lines() []*TOCLine {
	return t.lines
}

// SetHeading sets the text and the style of the heading of the TOC component.
func (t *TOC) SetHeading(text string, style TextStyle) {
	h := t.Heading()

	h.Reset()
	chunk := h.Append(text)
	chunk.Style = style
}

// Add adds a new line with the default style to the table of contents.
func (t *TOC) Add(number, title, page string, level uint) *TOCLine {
	tl := t.AddLine(newStyledTOCLine(
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
		t.defaultStyle,
	))

	if tl == nil {
		return nil
	}

	// Set line margins.
	m := &t.lineMargins
	tl.SetMargins(m.left, m.right, m.top, m.bottom)

	// Set line level offset.
	tl.SetLevelOffset(t.lineLevelOffset)

	// Set line separator text and style.
	tl.Separator.Text = t.lineSeparator
	tl.Separator.Style = t.lineSeparatorStyle

	return tl
}

// AddLine adds a new line with the provided style to the table of contents.
func (t *TOC) AddLine(line *TOCLine) *TOCLine {
	if line == nil {
		return nil
	}

	t.lines = append(t.lines, line)
	return line
}

// SetLineSeparator sets the separator for all new lines of the table of contents.
func (t *TOC) SetLineSeparator(separator string) {
	t.lineSeparator = separator
}

// SetLineMargins sets the margins for all new lines of the table of contents.
func (t *TOC) SetLineMargins(left, right, top, bottom float64) {
	m := &t.lineMargins

	m.left = left
	m.right = right
	m.top = top
	m.bottom = bottom
}

// SetLineStyle sets the style for all the line components: number, title,
// separator, page. The style is applied only for new lines added to the
// TOC component.
func (t *TOC) SetLineStyle(style TextStyle) {
	t.SetLineNumberStyle(style)
	t.SetLineTitleStyle(style)
	t.SetLineSeparatorStyle(style)
	t.SetLinePageStyle(style)
}

// SetLineNumberStyle sets the style for the numbers part of all new lines
// of the table of contents.
func (t *TOC) SetLineNumberStyle(style TextStyle) {
	t.lineNumberStyle = style
}

// SetLineTitleStyle sets the style for the title part of all new lines
// of the table of contents.
func (t *TOC) SetLineTitleStyle(style TextStyle) {
	t.lineTitleStyle = style
}

// SetLineSeparatorStyle sets the style for the separator part of all new
// lines of the table of contents.
func (t *TOC) SetLineSeparatorStyle(style TextStyle) {
	t.lineSeparatorStyle = style
}

// SetLinePageStyle sets the style for the page part of all new lines
// of the table of contents.
func (t *TOC) SetLinePageStyle(style TextStyle) {
	t.linePageStyle = style
}

// SetLineLevelOffset sets the amount of space an indentation level occupies
// for all new lines of the table of contents.
func (t *TOC) SetLineLevelOffset(levelOffset float64) {
	t.lineLevelOffset = levelOffset
}

// SetShowLinks sets visibility of links for the TOC lines.
func (t *TOC) SetShowLinks(showLinks bool) {
	t.showLinks = showLinks
}

// GeneratePageBlocks generate the Page blocks. Multiple blocks are generated
// if the contents wrap over multiple pages.
func (t *TOC) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	// Generate heading blocks.
	blocks, ctx, err := t.heading.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}

	// Generate blocks for the table of contents lines.
	for _, line := range t.lines {
		linkPage := line.linkPage
		if !t.showLinks {
			line.linkPage = 0
		}

		newBlocks, c, err := line.GeneratePageBlocks(ctx)
		line.linkPage = linkPage

		if err != nil {
			return blocks, ctx, err
		}
		if len(newBlocks) < 1 {
			continue
		}

		// The first block is always appended to the last.
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
