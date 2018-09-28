/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"strings"
)

type TOCLine struct {
	sp *StyledParagraph

	Number    TextChunk
	Title     TextChunk
	Page      TextChunk
	Separator TextChunk

	level       uint
	offset      float64
	levelOffset float64
}

func NewTOCLine(number, title, page string, level uint) *TOCLine {
	style := NewTextStyle()

	return NewStyledTOCLine(
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
	)
}

func NewStyledTOCLine(number, title, page TextChunk, level uint) *TOCLine {
	style := NewTextStyle()

	sp := NewStyledParagraph("", style)
	sp.SetEnableWrap(true)
	sp.SetTextAlignment(TextAlignmentLeft)

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
	}

	sp.margins.left = tl.offset + float64(tl.level-1)*tl.levelOffset
	sp.beforeRender = tl.prepareParagraph
	return tl
}

func (tl *TOCLine) Level() uint {
	return tl.level
}

func (tl *TOCLine) SetLevel(level uint) {
	tl.level = level
	tl.sp.margins.left = tl.offset + float64(tl.level-1)*tl.levelOffset
}

func (tl *TOCLine) LevelOffset() float64 {
	return tl.levelOffset
}

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

func (tl *TOCLine) prepareParagraph(sp *StyledParagraph, ctx DrawContext) {
	// Add text chunks to the paragraph
	title := tl.Title.Text
	if tl.Number.Text != "" {
		title = " " + title
	}
	title += " "

	page := tl.Page.Text
	if page != "" {
		page = " " + page
	}

	sp.chunks = []TextChunk{
		tl.Number,
		TextChunk{
			Text:  title,
			Style: tl.Title.Style,
		},
		TextChunk{
			Text:  page,
			Style: tl.Page.Style,
		},
	}

	sp.SetEncoder(sp.encoder)
	sp.wrapText()

	// Insert separator
	l := len(sp.lines)
	if l == 0 {
		return
	}

	availWidth := ctx.Width*1000 - sp.getTextLineWidth(sp.lines[l-1])
	sepWidth := sp.getTextLineWidth([]TextChunk{tl.Separator})
	sepCount := int(availWidth / sepWidth)
	sepText := strings.Repeat(tl.Separator.Text, sepCount)
	sepStyle := tl.Separator.Style

	sp.Insert(2, sepText, sepStyle)

	// Push page numbers to the end of the line
	availWidth = availWidth - float64(sepCount)*sepWidth
	if availWidth > 500 {
		spaceMetrics, found := sepStyle.Font.GetGlyphCharMetrics("space")
		if found && availWidth > spaceMetrics.Wx {
			spaces := int(availWidth / spaceMetrics.Wx)
			if spaces > 0 {
				style := sepStyle
				style.FontSize = 1
				sp.Insert(2, strings.Repeat(" ", spaces), style)
			}
		}
	}
}

func (tl *TOCLine) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	blocks, ctx, err := tl.sp.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}
	if len(blocks) > 1 {
		// Did not fit, moved to new Page block.
		ctx.Page++
	}

	if tl.sp.positioning.isRelative() {
		// Move back X to same start of line.
		ctx.X = origCtx.X
	}

	if tl.sp.positioning.isAbsolute() {
		// If absolute: return original context.
		return blocks, origCtx, nil
	}

	return blocks, ctx, nil
}
