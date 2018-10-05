/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// StyledParagraph represents text drawn with a specified font and can wrap across lines and pages.
// By default occupies the available width in the drawing context.
type StyledParagraph struct {
	// Text chunks with styles that compose the paragraph
	chunks []TextChunk

	// Style used for the paragraph for spacing and offsets
	defaultStyle TextStyle

	// The text encoder which can convert the text (as runes) into a series of glyphs and get character metrics.
	encoder textencoding.TextEncoder

	// Text alignment: Align left/right/center/justify.
	alignment TextAlignment

	// The line relative height (default 1).
	lineHeight float64

	// Wrapping properties.
	enableWrap bool
	wrapWidth  float64

	// defaultWrap defines whether wrapping has been defined explictly or whether default behavior should
	// be observed. Default behavior depends on context: normally wrap is expected, except for example in
	// table cells wrapping is off by default.
	defaultWrap bool

	// Rotation angle (degrees).
	angle float64

	// Margins to be applied around the block when drawing on Page.
	margins margins

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos float64
	yPos float64

	// Scaling factors (1 default).
	scaleX float64
	scaleY float64

	// Text chunk lines after wrapping to available width.
	lines [][]TextChunk
}

// NewStyledParagraph creates a new styled paragraph.
// Uses default parameters: Helvetica, WinAnsiEncoding and wrap enabled
// with a wrap width of 100 points.
func NewStyledParagraph(text string, style TextStyle) *StyledParagraph {
	// TODO: Can we wrap intellectually, only if given width is known?
	p := &StyledParagraph{
		chunks: []TextChunk{
			TextChunk{
				Text:  text,
				Style: style,
			},
		},
		defaultStyle: NewTextStyle(),
		lineHeight:   1.0,
		alignment:    TextAlignmentLeft,
		enableWrap:   true,
		defaultWrap:  true,
		angle:        0,
		scaleX:       1,
		scaleY:       1,
		positioning:  positionRelative,
	}

	p.SetEncoder(textencoding.NewWinAnsiTextEncoder())
	return p
}

// Append adds a new text chunk with a specified style to the paragraph.
func (p *StyledParagraph) Append(text string, style TextStyle) {
	chunk := TextChunk{
		Text:  text,
		Style: style,
	}
	chunk.Style.Font.SetEncoder(p.encoder)

	p.chunks = append(p.chunks, chunk)
	p.wrapText()
}

// Reset sets the entire text and also the style of the paragraph
// to those specified. It behaves as if the paragraph was a new one.
func (p *StyledParagraph) Reset(text string, style TextStyle) {
	p.chunks = []TextChunk{}
	p.Append(text, style)
}

// SetTextAlignment sets the horizontal alignment of the text within the space provided.
func (p *StyledParagraph) SetTextAlignment(align TextAlignment) {
	p.alignment = align
}

// SetEncoder sets the text encoding.
func (p *StyledParagraph) SetEncoder(encoder textencoding.TextEncoder) {
	p.encoder = encoder
	p.defaultStyle.Font.SetEncoder(encoder)

	// Sync with the text font too.
	// XXX/FIXME: Keep in 1 place only.
	for _, chunk := range p.chunks {
		chunk.Style.Font.SetEncoder(encoder)
	}
}

// SetLineHeight sets the line height (1.0 default).
func (p *StyledParagraph) SetLineHeight(lineheight float64) {
	p.lineHeight = lineheight
}

// SetEnableWrap sets the line wrapping enabled flag.
func (p *StyledParagraph) SetEnableWrap(enableWrap bool) {
	p.enableWrap = enableWrap
	p.defaultWrap = false
}

// SetPos sets absolute positioning with specified coordinates.
func (p *StyledParagraph) SetPos(x, y float64) {
	p.positioning = positionAbsolute
	p.xPos = x
	p.yPos = y
}

// SetAngle sets the rotation angle of the text.
func (p *StyledParagraph) SetAngle(angle float64) {
	p.angle = angle
}

// SetMargins sets the Paragraph's margins.
func (p *StyledParagraph) SetMargins(left, right, top, bottom float64) {
	p.margins.left = left
	p.margins.right = right
	p.margins.top = top
	p.margins.bottom = bottom
}

// GetMargins returns the Paragraph's margins: left, right, top, bottom.
func (p *StyledParagraph) GetMargins() (float64, float64, float64, float64) {
	return p.margins.left, p.margins.right, p.margins.top, p.margins.bottom
}

// SetWidth sets the the Paragraph width. This is essentially the wrapping width,
// i.e. the width the text can extend to prior to wrapping over to next line.
func (p *StyledParagraph) SetWidth(width float64) {
	p.wrapWidth = width
	p.wrapText()
}

// Width returns the width of the Paragraph.
func (p *StyledParagraph) Width() float64 {
	if p.enableWrap {
		return p.wrapWidth
	}

	return p.getTextWidth() / 1000.0
}

// Height returns the height of the Paragraph. The height is calculated based on the input text and how it is wrapped
// within the container. Does not include Margins.
func (p *StyledParagraph) Height() float64 {
	if p.lines == nil || len(p.lines) == 0 {
		p.wrapText()
	}

	var height float64
	for _, line := range p.lines {
		var lineHeight float64
		for _, chunk := range line {
			h := p.lineHeight * chunk.Style.FontSize
			if h > lineHeight {
				lineHeight = h
			}
		}

		height += lineHeight
	}

	return height
}

// getTextWidth calculates the text width as if all in one line (not taking wrapping into account).
func (p *StyledParagraph) getTextWidth() float64 {
	var width float64
	for _, chunk := range p.chunks {
		style := &chunk.Style

		for _, rune := range chunk.Text {
			glyph, found := p.encoder.RuneToGlyph(rune)
			if !found {
				common.Log.Debug("Error! Glyph not found for rune: %s\n", rune)

				// XXX/FIXME: return error.
				return -1
			}

			// Ignore newline for this.. Handles as if all in one line.
			if glyph == "controlLF" {
				continue
			}

			metrics, found := style.Font.GetGlyphCharMetrics(glyph)
			if !found {
				common.Log.Debug("Glyph char metrics not found! %s\n", glyph)

				// XXX/FIXME: return error.
				return -1
			}

			width += style.FontSize * metrics.Wx
		}
	}

	return width
}

// getTextHeight calculates the text height as if all in one line (not taking wrapping into account).
func (p *StyledParagraph) getTextHeight() float64 {
	var height float64
	for _, chunk := range p.chunks {
		h := chunk.Style.FontSize * p.lineHeight
		if h > height {
			height = h
		}
	}

	return height
}

// wrapText splits text into lines. It uses a simple greedy algorithm to wrap
// fill the lines.
// XXX/TODO: Consider the Knuth/Plass algorithm or an alternative.
func (p *StyledParagraph) wrapText() error {
	if !p.enableWrap {
		p.lines = [][]TextChunk{p.chunks}
		return nil
	}

	p.lines = [][]TextChunk{}
	var line []TextChunk
	var lineWidth float64

	for _, chunk := range p.chunks {
		style := chunk.Style

		var part []rune
		var glyphs []string
		var widths []float64

		for _, r := range chunk.Text {
			glyph, found := p.encoder.RuneToGlyph(r)
			if !found {
				common.Log.Debug("Error! Glyph not found for rune: %v\n", r)

				// XXX/FIXME: return error.
				return errors.New("Glyph not found for rune")
			}

			// newline wrapping.
			if glyph == "controlLF" {
				// moves to next line.
				line = append(line, TextChunk{
					Text:  strings.TrimRightFunc(string(part), unicode.IsSpace),
					Style: style,
				})
				p.lines = append(p.lines, line)
				line = []TextChunk{}

				lineWidth = 0
				part = []rune{}
				widths = []float64{}
				glyphs = []string{}
				continue
			}

			metrics, found := style.Font.GetGlyphCharMetrics(glyph)
			if !found {
				common.Log.Debug("Glyph char metrics not found! %s\n", glyph)

				// XXX/FIXME: return error.
				return errors.New("Glyph char metrics missing")
			}

			w := style.FontSize * metrics.Wx
			if lineWidth+w > p.wrapWidth*1000.0 {
				// Goes out of bounds: Wrap.
				// Breaks on the character.
				// XXX/TODO: when goes outside: back up to next space,
				// otherwise break on the character.
				idx := -1
				for j := len(glyphs) - 1; j >= 0; j-- {
					if glyphs[j] == "space" {
						idx = j
						break
					}
				}

				text := string(part)
				if idx >= 0 {
					text = string(part[0 : idx+1])

					part = part[idx+1:]
					part = append(part, r)
					glyphs = glyphs[idx+1:]
					glyphs = append(glyphs, glyph)
					widths = widths[idx+1:]
					widths = append(widths, w)

					lineWidth = 0
					for _, width := range widths {
						lineWidth += width
					}
				} else {
					lineWidth = w
					part = []rune{r}
					glyphs = []string{glyph}
					widths = []float64{w}
				}

				line = append(line, TextChunk{
					Text:  strings.TrimRightFunc(string(text), unicode.IsSpace),
					Style: style,
				})
				p.lines = append(p.lines, line)
				line = []TextChunk{}
			} else {
				lineWidth += w
				part = append(part, r)
				glyphs = append(glyphs, glyph)
				widths = append(widths, w)
			}
		}

		if len(part) > 0 {
			line = append(line, TextChunk{
				Text:  string(part),
				Style: style,
			})
		}
	}

	if len(line) > 0 {
		p.lines = append(p.lines, line)
	}

	return nil
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated
// if the contents wrap over multiple pages. Implements the Drawable interface.
func (p *StyledParagraph) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origContext := ctx
	blocks := []*Block{}

	blk := NewBlock(ctx.PageWidth, ctx.PageHeight)
	if p.positioning.isRelative() {
		// Account for Paragraph Margins.
		ctx.X += p.margins.left
		ctx.Y += p.margins.top
		ctx.Width -= p.margins.left + p.margins.right
		ctx.Height -= p.margins.top + p.margins.bottom

		// Use available space.
		p.SetWidth(ctx.Width)

		if p.Height() > ctx.Height {
			// Goes out of the bounds.  Write on a new template instead and create a new context at upper
			// left corner.
			// XXX/TODO: Handle case when Paragraph is larger than the Page...
			// Should be fine if we just break on the paragraph, i.e. splitting it up over 2+ pages

			blocks = append(blocks, blk)
			blk = NewBlock(ctx.PageWidth, ctx.PageHeight)

			// New Page.
			ctx.Page++
			newContext := ctx
			newContext.Y = ctx.Margins.top // + p.Margins.top
			newContext.X = ctx.Margins.left + p.margins.left
			newContext.Height = ctx.PageHeight - ctx.Margins.top - ctx.Margins.bottom - p.margins.bottom
			newContext.Width = ctx.PageWidth - ctx.Margins.left - ctx.Margins.right - p.margins.left - p.margins.right
			ctx = newContext
		}
	} else {
		// Absolute.
		if p.wrapWidth == 0 {
			// Use necessary space.
			p.SetWidth(p.getTextWidth())
		}
		ctx.X = p.xPos
		ctx.Y = p.yPos
	}

	// Place the Paragraph on the template at position (x,y) based on the ctx.
	ctx, err := drawStyledParagraphOnBlock(blk, p, ctx)
	if err != nil {
		common.Log.Debug("ERROR: %v", err)
		return nil, ctx, err
	}

	blocks = append(blocks, blk)
	if p.positioning.isRelative() {
		ctx.X -= p.margins.left // Move back.
		ctx.Width = origContext.Width
		return blocks, ctx, nil
	}
	// Absolute: not changing the context.
	return blocks, origContext, nil
}

// Draw block on specified location on Page, adding to the content stream.
func drawStyledParagraphOnBlock(blk *Block, p *StyledParagraph, ctx DrawContext) (DrawContext, error) {
	// Find first free index for the font resources of the paragraph
	num := 1
	fontName := core.PdfObjectName(fmt.Sprintf("Font%d", num))
	for blk.resources.HasFontByName(fontName) {
		num++
		fontName = core.PdfObjectName(fmt.Sprintf("Font%d", num))
	}

	// Add default font to the page resources
	err := blk.resources.SetFontByName(fontName, p.defaultStyle.Font.ToPdfObject())
	if err != nil {
		return ctx, err
	}
	num++

	defaultFontName := fontName
	defaultFontSize := p.defaultStyle.FontSize

	// Wrap the text into lines.
	p.wrapText()

	// Add the fonts of all chunks to the page resources
	fonts := [][]core.PdfObjectName{}

	for _, line := range p.lines {
		fontLine := []core.PdfObjectName{}

		for _, chunk := range line {
			fontName = core.PdfObjectName(fmt.Sprintf("Font%d", num))

			err := blk.resources.SetFontByName(fontName, chunk.Style.Font.ToPdfObject())
			if err != nil {
				return ctx, err
			}

			fontLine = append(fontLine, fontName)
			num++
		}

		fonts = append(fonts, fontLine)
	}

	// Create the content stream.
	cc := contentstream.NewContentCreator()
	cc.Add_q()

	yPos := ctx.PageHeight - ctx.Y - defaultFontSize*p.lineHeight
	cc.Translate(ctx.X, yPos)

	if p.angle != 0 {
		cc.RotateDeg(p.angle)
	}

	cc.Add_BT()

	for idx, line := range p.lines {
		if idx != 0 {
			// Move to next line if not first.
			cc.Add_Tstar()
		}

		isLastLine := idx == len(p.lines)-1

		// Get width of the line (excluding spaces).
		var width float64
		var spaceWidth float64
		var spaces uint

		for _, chunk := range line {
			style := &chunk.Style

			spaceMetrics, found := style.Font.GetGlyphCharMetrics("space")
			if !found {
				return ctx, errors.New("The font does not have a space glyph")
			}

			var chunkSpaces uint
			for _, r := range chunk.Text {
				glyph, found := p.encoder.RuneToGlyph(r)
				if !found {
					common.Log.Debug("Rune 0x%x not supported by text encoder", r)
					return ctx, errors.New("Unsupported rune in text encoding")
				}

				if glyph == "space" {
					chunkSpaces++
					continue
				}
				if glyph == "controlLF" {
					continue
				}

				metrics, found := style.Font.GetGlyphCharMetrics(glyph)
				if !found {
					common.Log.Debug("Unsupported glyph %s in font\n", glyph)
					return ctx, errors.New("Unsupported text glyph")
				}

				width += style.FontSize * metrics.Wx
			}

			spaceWidth += float64(chunkSpaces) * spaceMetrics.Wx * style.FontSize
			spaces += chunkSpaces
		}

		// Add line shifts
		objs := []core.PdfObject{}
		if p.alignment == TextAlignmentJustify {
			// Not to justify last line.
			if spaces > 0 && !isLastLine {
				spaceWidth = (p.wrapWidth*1000.0 - width) / float64(spaces) / defaultFontSize
			}
		} else if p.alignment == TextAlignmentCenter {
			// Start with a shift.
			shift := (p.wrapWidth*1000.0 - width - spaceWidth) / 2 / defaultFontSize
			objs = append(objs, core.MakeFloat(-shift))
		} else if p.alignment == TextAlignmentRight {
			shift := (p.wrapWidth*1000.0 - width - spaceWidth) / defaultFontSize
			objs = append(objs, core.MakeFloat(-shift))
		}

		if len(objs) > 0 {
			cc.Add_Tf(defaultFontName, defaultFontSize).
				Add_TL(defaultFontSize * p.lineHeight).
				Add_TJ(objs...)
		}

		// Render line text chunks
		for k, chunk := range line {
			style := &chunk.Style

			r, g, b := style.Color.ToRGB()
			fontName := defaultFontName
			fontSize := defaultFontSize

			if p.alignment != TextAlignmentJustify || isLastLine {
				spaceMetrics, found := style.Font.GetGlyphCharMetrics("space")
				if !found {
					return ctx, errors.New("The font does not have a space glyph")
				}

				fontName = fonts[idx][k]
				fontSize = style.FontSize
				spaceWidth = spaceMetrics.Wx
			}

			encStr := ""
			for _, rn := range chunk.Text {
				glyph, found := p.encoder.RuneToGlyph(rn)
				if !found {
					common.Log.Debug("Rune 0x%x not supported by text encoder", r)
					return ctx, errors.New("Unsupported rune in text encoding")
				}

				if glyph == "space" {
					if !found {
						common.Log.Debug("Unsupported glyph %s in font\n", glyph)
						return ctx, errors.New("Unsupported text glyph")
					}

					if len(encStr) > 0 {
						cc.Add_rg(r, g, b).
							Add_Tf(fonts[idx][k], style.FontSize).
							Add_TL(style.FontSize * p.lineHeight).
							Add_TJ([]core.PdfObject{core.MakeString(encStr)}...)

						encStr = ""
					}

					cc.Add_Tf(fontName, fontSize).
						Add_TL(fontSize * p.lineHeight).
						Add_TJ([]core.PdfObject{core.MakeFloat(-spaceWidth)}...)
				} else {
					encStr += p.encoder.Encode(string(rn))
				}
			}

			if len(encStr) > 0 {
				cc.Add_rg(r, g, b).
					Add_Tf(fonts[idx][k], style.FontSize).
					Add_TL(style.FontSize * p.lineHeight).
					Add_TJ([]core.PdfObject{core.MakeString(encStr)}...)
			}
		}
	}
	cc.Add_ET()
	cc.Add_Q()

	ops := cc.Operations()
	ops.WrapIfNeeded()

	blk.addContents(ops)

	if p.positioning.isRelative() {
		pHeight := p.Height() + p.margins.bottom
		ctx.Y += pHeight
		ctx.Height -= pHeight

		// If the division is inline, calculate context new X coordinate.
		if ctx.Inline {
			ctx.X += p.Width() + p.margins.right
		}
	}

	return ctx, nil
}
