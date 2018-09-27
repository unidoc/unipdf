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

	number    TextChunk
	title     TextChunk
	page      TextChunk
	separator TextChunk

	level       uint
	levelOffset float64
}

func NewTOCLine(number, title, page string, level uint) *TOCLine {
	style := NewTextStyle()

	return NewStyledTOCLine(
		TextChunk{Text: number, Style: style},
		TextChunk{Text: title, Style: style},
		TextChunk{Text: page, Style: style},
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
		number: number,
		title:  title,
		page:   page,
		separator: TextChunk{
			Text:  ".",
			Style: style,
		},
		level:       level,
		levelOffset: 10,
	}

	sp.margins.left = float64(level) * tl.levelOffset
	sp.beforeRender = tl.prepareParagraph
	return tl
}

func (tl *TOCLine) prepareParagraph(sp *StyledParagraph, ctx DrawContext) {
	if tl.number.Text != "" {
		tl.title.Text = " " + tl.title.Text
	}
	tl.title.Text += " "

	if tl.page.Text != "" {
		tl.page.Text = " " + tl.page.Text
	}

	sp.chunks = []TextChunk{
		tl.number,
		tl.title,
		tl.page,
	}
	sp.SetEncoder(sp.encoder)
	sp.wrapText()

	l := len(sp.lines)
	if l == 0 {
		return
	}

	// Insert separator
	availWidth := ctx.Width*1000 - sp.getTextLineWidth(sp.lines[l-1])
	sepWidth := sp.getTextLineWidth([]TextChunk{tl.separator})
	sepCount := int(availWidth / sepWidth)
	sepText := strings.Repeat(tl.separator.Text, sepCount)

	sp.Insert(2, sepText, tl.separator.Style)

	// Push page numbers to the end of the line
	availWidth = availWidth - float64(sepCount)*sepWidth
	if availWidth > 500 {
		spaceMetrics, found := tl.separator.Style.Font.GetGlyphCharMetrics("space")
		if found && availWidth > spaceMetrics.Wx {
			spaces := int(availWidth / spaceMetrics.Wx)
			if spaces > 0 {
				style := tl.separator.Style
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
