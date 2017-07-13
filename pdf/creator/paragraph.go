/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// XXX/TODO: Under consideration.  Should allow paragraph to scale?  Makes more sense to change font size.
// Alternatively can draw to a block and scale the block, if need to fit into a specific slot.

// A paragraph represents text drawn with a specified font and can wrap across lines and pages.
// By default occupies the available width in the drawing context.
type paragraph struct {
	// The input utf-8 text as a string (series of runes).
	text string

	// The text encoder which can convert the text (as runes) into a series of glyphs and get character metrics.
	encoder textencoding.TextEncoder

	// The font to be used to draw the text.
	textFont fonts.Font

	// The font size (points).
	fontSize float64

	// The line relative height (default 1).
	lineHeight float64

	// The text color.
	color model.PdfColorDeviceRGB

	// Text alignment: Align left/right/center/justify.
	alignment TextAlignment

	// Wrapping properties.
	enableWrap bool
	wrapWidth  float64

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
	scaleX, scaleY float64

	// Text lines after wrapping to available width.
	textLines []string
}

// Create a new text block.  Uses default parameters: Helvetica, WinAnsiEncoding and wrap enabled
// with a wrap width of 100 points.
func NewParagraph(text string) *paragraph {
	p := &paragraph{}
	p.text = text
	p.textFont = fonts.NewFontHelvetica()
	p.SetEncoder(textencoding.NewWinAnsiTextEncoder())
	p.fontSize = 10
	p.lineHeight = 1.0
	p.enableWrap = true
	p.SetColor(0, 0, 0)
	p.alignment = TextAlignmentLeft
	p.angle = 0

	p.scaleX = 1
	p.scaleY = 1

	p.positioning = positionRelative

	return p
}

func (p *paragraph) SetFont(font fonts.Font) {
	p.textFont = font
}

func (p *paragraph) SetFontSize(fontSize float64) {
	p.fontSize = fontSize
}

// Alignment of the text within the width provided.
func (p *paragraph) SetTextAlignment(align TextAlignment) {
	p.alignment = align
}

// Set text encoding.
func (p *paragraph) SetEncoder(encoder textencoding.TextEncoder) {
	p.encoder = encoder
	// Sync with the text font too.
	// XXX/FIXME: Keep in 1 place only.
	p.textFont.SetEncoder(encoder)
}

func (p *paragraph) SetLineHeight(lineheight float64) {
	p.lineHeight = lineheight
}

func (p *paragraph) SetText(text string) {
	p.text = text
}

// Set line wrapping enabled flag.
func (p *paragraph) SetEnableWrap(enableWrap bool) {
	p.enableWrap = enableWrap
}

// Set RGB color.
func (p *paragraph) SetColor(r, g, b float64) {
	color := model.NewPdfColorDeviceRGB(r, g, b)
	p.color = *color
}

// Drawable interface implementations.

// Set absolute positioning with specified coordinates.
func (p *paragraph) SetPos(x, y float64) {
	p.positioning = positionAbsolute
	p.xPos = x
	p.yPos = y
}

// Set rotation angle.
func (p *paragraph) SetAngle(angle float64) {
	p.angle = angle
}

// Set paragraph Margins.
func (p *paragraph) SetMargins(left, right, top, bottom float64) {
	p.margins.left = left
	p.margins.right = right
	p.margins.top = top
	p.margins.bottom = bottom
}

// Get paragraph Margins: left, right, top, bottom.
func (p *paragraph) GetMargins() (float64, float64, float64, float64) {
	return p.margins.left, p.margins.right, p.margins.top, p.margins.bottom
}

// Set the paragraph width. Esentially the wrapping width, the width the text can extend to prior to wrapping.
func (p *paragraph) SetWidth(width float64) {
	p.wrapWidth = width
	p.wrapText()
}

func (p *paragraph) Width() float64 {
	return p.wrapWidth
}

// The height is calculated based on the input text and how it is wrapped within the container.
// Height does not include Margins.
func (p *paragraph) Height() float64 {
	if p.textLines == nil || len(p.textLines) == 0 {
		p.wrapText()
	}

	h := float64(len(p.textLines)) * p.lineHeight * p.fontSize
	return h
}

func (p *paragraph) Scale(sx, sy float64) {
	p.scaleX = sx
	p.scaleY = sy
}

func (p *paragraph) ScaleToHeight(h float64) {
	ratio := h / p.Height()
	p.Scale(ratio, ratio)
}

func (p *paragraph) ScaleToWidth(w float64) {
	ratio := w / p.Width()
	p.Scale(ratio, ratio)
}

// Calculate the text width (if not wrapped).
func (p *paragraph) getTextWidth() float64 {
	w := float64(0.0)

	for _, rune := range p.text {
		glyph, found := p.encoder.RuneToGlyph(rune)
		if !found {
			common.Log.Debug("Error! Glyph not found for rune: %s\n", rune)
			return -1 // XXX/FIXME: return error.
		}

		metrics, found := p.textFont.GetGlyphCharMetrics(glyph)
		if !found {
			common.Log.Debug("Glyph char metrics not found! %s\n", glyph)
			return -1 // XXX/FIXME: return error.
		}
		w += p.fontSize * metrics.Wx
	}

	return w
}

// Simple algorithm to wrap the text into lines (greedy algorithm - fill the lines).
// XXX/TODO: Consider the Knuth/Plass algorithm or an alternative.
func (p *paragraph) wrapText() error {
	if !p.enableWrap {
		p.textLines = []string{p.encoder.Encode(p.text)}
		return nil
	}

	line := []rune{}
	lineWidth := float64(0.0)
	p.textLines = []string{}

	runes := []rune(p.text)
	glyphs := []string{}
	widths := []float64{}

	for _, val := range runes {
		glyph, found := p.encoder.RuneToGlyph(val)
		if !found {
			common.Log.Debug("Error! Glyph not found for rune: %v\n", val)
			return errors.New("Glyph not found for rune") // XXX/FIXME: return error.
		}

		metrics, found := p.textFont.GetGlyphCharMetrics(glyph)
		if !found {
			common.Log.Debug("Glyph char metrics not found! %s\n", glyph)
			return errors.New("Glyph char metrics missing") // XXX/FIXME: return error.
		}

		w := p.fontSize * metrics.Wx
		if lineWidth+w > p.wrapWidth*1000.0 {
			// Goes out of bounds: Wrap.
			// Breaks on the character.
			// XXX/TODO: when goes outside: back up to next space, otherwise break on the character.
			idx := -1
			for i := len(glyphs) - 1; i >= 0; i-- {
				if glyphs[i] == "space" {
					idx = i
					break
				}
			}
			if idx > 0 {
				p.textLines = append(p.textLines, string(line[0:idx+1]))

				line = line[idx+1:]
				line = append(line, val)

				glyphs = glyphs[idx+1:]
				glyphs = append(glyphs, glyph)
				widths = widths[idx+1:]
				widths = append(widths, w)

				lineWidth = 0
				for _, width := range widths {
					lineWidth += width
				}

			} else {
				p.textLines = append(p.textLines, string(line))
				line = []rune{val}
				lineWidth = w
				widths = []float64{w}
				glyphs = []string{glyph}
			}
		} else {
			line = append(line, val)
			lineWidth += w
			glyphs = append(glyphs, glyph)
			widths = append(widths, w)
		}
	}
	if len(line) > 0 {
		p.textLines = append(p.textLines, string(line))
	}

	return nil
}

// Generate the Page blocks.  Multiple blocks are generated if the contents wrap over
// multiple pages.
func (p *paragraph) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origContext := ctx
	blocks := []*Block{}

	blk := NewBlock(ctx.PageWidth, ctx.PageHeight)
	if p.positioning.isRelative() {
		// Account for paragraph Margins.
		ctx.X += p.margins.left
		ctx.Y += p.margins.top
		ctx.Width -= p.margins.left + p.margins.right
		ctx.Height -= p.margins.top + p.margins.bottom

		// Use available space.
		p.SetWidth(ctx.Width)

		if p.Height() > ctx.Height {
			// Goes out of the bounds.  Write on a new template instead and create a new context at upper
			// left corner.
			// XXX/TODO: Handle case when paragraph is larger than the Page...
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

	// Place the paragraph on the template at position (x,y) based on the ctx.
	ctx, err := drawParagraphOnBlock(blk, p, ctx)
	if err != nil {
		common.Log.Debug("ERROR: %v", err)
		return nil, ctx, err
	}

	blocks = append(blocks, blk)
	if p.positioning.isRelative() {
		ctx.X -= p.margins.left // Move back.
		return blocks, ctx, nil
	} else {
		// Absolute: not changing the context.
		return blocks, origContext, nil
	}
}

// Draw block on specified location on Page, adding to the content stream.
func drawParagraphOnBlock(blk *Block, p *paragraph, ctx DrawContext) (DrawContext, error) {
	// Find a free name for the font.
	num := 1
	fontName := core.PdfObjectName(fmt.Sprintf("Font%d", num))
	for blk.resources.HasFontByName(fontName) {
		num++
		fontName = core.PdfObjectName(fmt.Sprintf("Font%d", num))
	}

	// Add to the Page resources.
	err := blk.resources.SetFontByName(fontName, p.textFont.ToPdfObject())
	if err != nil {
		return ctx, err
	}

	// Wrap the text into lines.
	p.wrapText()

	// Create the content stream.
	cc := contentstream.NewContentCreator()
	cc.Add_q()

	yPos := ctx.PageHeight - ctx.Y - p.fontSize*p.lineHeight

	cc.Translate(ctx.X, yPos)
	if p.angle != 0 {
		cc.RotateDeg(p.angle)
	}

	cc.Add_BT().
		Add_rg(p.color.R(), p.color.G(), p.color.B()).
		Add_Tf(fontName, p.fontSize).
		Add_TL(p.fontSize * p.lineHeight)

	for idx, line := range p.textLines {
		if idx != 0 {
			// Move to next line if not first.
			cc.Add_Tstar()
		}

		runes := []rune(line)

		// Get width of the line (excluding spaces).
		w := float64(0)
		spaces := 0
		for _, runeVal := range runes {
			glyph, found := p.encoder.RuneToGlyph(runeVal)
			if !found {
				common.Log.Debug("Rune 0x%x not supported by text encoder", runeVal)
				return ctx, errors.New("Unsupported rune in text encoding")
			}
			if glyph == "space" {
				spaces++
				continue
			}
			metrics, found := p.textFont.GetGlyphCharMetrics(glyph)
			if !found {
				common.Log.Debug("Unsupported glyph %s in font\n", glyph)
				return ctx, errors.New("Unsupported text glyph")
			}

			w += p.fontSize * metrics.Wx
		}

		objs := []core.PdfObject{}

		spaceMetrics, found := p.textFont.GetGlyphCharMetrics("space")
		if !found {
			return ctx, errors.New("The font does not have a space glyph")
		}
		spaceWidth := spaceMetrics.Wx
		if p.alignment == TextAlignmentJustify {
			if spaces > 0 && idx < len(p.textLines)-1 { // Not to justify last line.
				spaceWidth = (p.wrapWidth*1000.0 - w) / float64(spaces) / p.fontSize
			}
		} else if p.alignment == TextAlignmentCenter {
			// Start with a shift.
			textWidth := w + float64(spaces)*spaceWidth*p.fontSize
			shift := (p.wrapWidth*1000.0 - textWidth) / 2 / p.fontSize
			objs = append(objs, core.MakeFloat(-shift))
		} else if p.alignment == TextAlignmentRight {
			textWidth := w + float64(spaces)*spaceWidth*p.fontSize
			shift := (p.wrapWidth*1000.0 - textWidth) / p.fontSize
			objs = append(objs, core.MakeFloat(-shift))
		}

		encStr := ""
		for _, runeVal := range runes {
			//creator.Add_Tj(core.PdfObjectString(tb.Encoder.Encode(line)))
			glyph, found := p.encoder.RuneToGlyph(runeVal)
			if !found {
				common.Log.Debug("Rune 0x%x not supported by text encoder", runeVal)
				return ctx, errors.New("Unsupported rune in text encoding")
			}

			if glyph == "space" {
				if !found {
					common.Log.Debug("Unsupported glyph %s in font\n", glyph)
					return ctx, errors.New("Unsupported text glyph")
				}

				if len(encStr) > 0 {
					objs = append(objs, core.MakeString(encStr))
					encStr = ""
				}
				objs = append(objs, core.MakeFloat(-spaceWidth))
			} else {
				encStr += string(p.encoder.Encode(string(runeVal)))
			}
		}
		if len(encStr) > 0 {
			objs = append(objs, core.MakeString(encStr))
		}

		cc.Add_TJ(objs...)
	}
	cc.Add_ET()
	cc.Add_Q()

	ops := cc.Operations()
	ops.WrapIfNeeded()

	blk.addContents(ops)

	if p.positioning.isRelative() {
		ctx.Y += p.Height() + p.margins.bottom
		ctx.Height -= p.Height() + p.margins.bottom
	}

	return ctx, nil
}
