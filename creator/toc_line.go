/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"strings"

	"github.com/unidoc/unipdf/v3/model"
)

// TOCLine represents a line in a table of contents.
// The component can be used both in the context of a
// table of contents component and as a standalone component.
// The representation of a table of contents line is as follows:
//       [number] [title]      [separator] [page]
// e.g.: Chapter1 Introduction ........... 1
type TOCLine struct {
	// The underlyng styled paragraph used to render the TOC line.
	sp *StyledParagraph

	// Holds the text and style of the number part of the TOC line.
	Number TextChunk

	// Holds the text and style of the title part of the TOC line.
	Title TextChunk

	// Holds the text and style of the separator part of the TOC line.
	Separator TextChunk

	// Holds the text and style of the page part of the TOC line.
	Page TextChunk

	// The left margin of the TOC line.
	offset float64

	// The indentation level of the TOC line.
	level uint

	// The amount of space an indentation level occupies.
	levelOffset float64

	// Positioning: relative/absolute.
	positioning positioning

	// Page and position of the line link, if any.
	linkX    float64
	linkY    float64
	linkPage int64
}

// newTOCLine creates a new table of contents line with the default style.
func newTOCLine(number, title, page string, level uint, style TextStyle) *TOCLine {
	return newStyledTOCLine(
		TextChunk{
			Text:  number,
			Style: style,
		},
		TextChunk{
			Text:  title,
			Style: style,
		},
		TextChunk{
			Text:  page,
			Style: style,
		},
		level,
		style,
	)
}

// newStyledTOCLine creates a new table of contents line with the provided style.
func newStyledTOCLine(number, title, page TextChunk, level uint, style TextStyle) *TOCLine {
	sp := newStyledParagraph(style)
	sp.SetEnableWrap(true)
	sp.SetTextAlignment(TextAlignmentLeft)
	sp.SetMargins(0, 0, 2, 2)

	tl := &TOCLine{
		sp:     sp,
		Number: number,
		Title:  title,
		Page:   page,
		Separator: TextChunk{
			Text:  ".",
			Style: style,
		},
		offset:      0,
		level:       level,
		levelOffset: 10,
		positioning: positionRelative,
	}

	sp.margins.left = tl.offset + float64(tl.level-1)*tl.levelOffset
	sp.beforeRender = tl.prepareParagraph
	return tl
}

// SetStyle sets the style for all the line components: number, title,
// separator, page.
func (tl *TOCLine) SetStyle(style TextStyle) {
	tl.Number.Style = style
	tl.Title.Style = style
	tl.Separator.Style = style
	tl.Page.Style = style
}

// Level returns the indentation level of the TOC line.
func (tl *TOCLine) Level() uint {
	return tl.level
}

// SetLevel sets the indentation level of the TOC line.
func (tl *TOCLine) SetLevel(level uint) {
	tl.level = level
	tl.sp.margins.left = tl.offset + float64(tl.level-1)*tl.levelOffset
}

// LevelOffset returns the amount of space an indentation level occupies.
func (tl *TOCLine) LevelOffset() float64 {
	return tl.levelOffset
}

// SetLevelOffset sets the amount of space an indentation level occupies.
func (tl *TOCLine) SetLevelOffset(levelOffset float64) {
	tl.levelOffset = levelOffset
	tl.sp.margins.left = tl.offset + float64(tl.level-1)*tl.levelOffset
}

// GetMargins returns the margins of the TOC line: left, right, top, bottom.
func (tl *TOCLine) GetMargins() (float64, float64, float64, float64) {
	m := &tl.sp.margins
	return tl.offset, m.right, m.top, m.bottom
}

// SetMargins sets the margins TOC line.
func (tl *TOCLine) SetMargins(left, right, top, bottom float64) {
	tl.offset = left

	m := &tl.sp.margins
	m.left = tl.offset + float64(tl.level-1)*tl.levelOffset
	m.right = right
	m.top = top
	m.bottom = bottom
}

// SetLink makes the line an internal link.
// The text parameter represents the text that is displayed.
// The user is taken to the specified page, at the specified x and y
// coordinates. Position 0, 0 is at the top left of the page.
func (tl *TOCLine) SetLink(page int64, x, y float64) {
	tl.linkX = x
	tl.linkY = y
	tl.linkPage = page

	tl.SetStyle(tl.sp.defaultLinkStyle)
}

// getLineLink returns a new annotation if the line has a link set.
func (tl *TOCLine) getLineLink() *model.PdfAnnotation {
	if tl.linkPage <= 0 {
		return nil
	}

	return newInternalLinkAnnotation(tl.linkPage-1, tl.linkX, tl.linkY, 0)
}

// prepareParagraph generates and adds all the components of the TOC line
// to the underlying paragraph.
func (tl *TOCLine) prepareParagraph(sp *StyledParagraph, ctx DrawContext) {
	// Add text chunks to the paragraph.
	title := tl.Title.Text
	if tl.Number.Text != "" {
		title = " " + title
	}
	title += " "

	page := tl.Page.Text
	if page != "" {
		page = " " + page
	}

	sp.chunks = []*TextChunk{
		{
			Text:       tl.Number.Text,
			Style:      tl.Number.Style,
			annotation: tl.getLineLink(),
		},
		{
			Text:       title,
			Style:      tl.Title.Style,
			annotation: tl.getLineLink(),
		},
		{
			Text:       page,
			Style:      tl.Page.Style,
			annotation: tl.getLineLink(),
		},
	}

	sp.wrapText()

	// Insert separator.
	l := len(sp.lines)
	if l == 0 {
		return
	}

	availWidth := ctx.Width*1000 - sp.getTextLineWidth(sp.lines[l-1])
	sepWidth := sp.getTextLineWidth([]*TextChunk{&tl.Separator})
	sepCount := int(availWidth / sepWidth)
	sepText := strings.Repeat(tl.Separator.Text, sepCount)
	sepStyle := tl.Separator.Style

	chunk := sp.Insert(2, sepText)
	chunk.Style = sepStyle
	chunk.annotation = tl.getLineLink()

	// Push page numbers to the end of the line.
	availWidth = availWidth - float64(sepCount)*sepWidth
	if availWidth > 500 {
		spaceMetrics, found := sepStyle.Font.GetRuneMetrics(' ')
		if found && availWidth > spaceMetrics.Wx {
			spaces := int(availWidth / spaceMetrics.Wx)
			if spaces > 0 {
				style := sepStyle
				style.FontSize = 1

				chunk = sp.Insert(2, strings.Repeat(" ", spaces))
				chunk.Style = style
				chunk.annotation = tl.getLineLink()
			}
		}
	}
}

// GeneratePageBlocks generate the Page blocks. Multiple blocks are generated
// if the contents wrap over multiple pages.
func (tl *TOCLine) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	blocks, ctx, err := tl.sp.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}

	if tl.positioning.isRelative() {
		// Move back X to same start of line.
		ctx.X = origCtx.X
	}

	if tl.positioning.isAbsolute() {
		// If absolute: return original context.
		return blocks, origCtx, nil
	}

	return blocks, ctx, nil
}
