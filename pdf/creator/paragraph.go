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
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// Paragraph represents text drawn with a specified font and can wrap across lines and pages.
// By default it occupies the available width in the drawing context.
type Paragraph struct {
	// The input utf-8 text as a string (series of runes).
	text string

	// The font to be used to draw the text.
	textFont *model.PdfFont

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

// NewParagraph create a new text paragraph. Uses default parameters: Helvetica, WinAnsiEncoding and
// wrap enabled with a wrap width of 100 points.
func NewParagraph(text string) *Paragraph {
	p := &Paragraph{}
	p.text = text

	font, encoder, err := model.NewStandard14FontWithEncoding("Helvetica", model.GetAlphabet(text))
	if err != nil {
		common.Log.Debug("ERROR: NewStandard14FontWithEncoding failed err=%v. Falling back.", err)
		p.textFont = model.DefaultFont()
	}
	p.textFont = font
	p.SetEncoder(encoder)

	p.fontSize = 10
	p.lineHeight = 1.0

	// TODO: Can we wrap intellectually, only if given width is known?
	p.SetColor(ColorRGBFrom8bit(0, 0, 0))
	p.alignment = TextAlignmentLeft
	p.angle = 0

	p.scaleX = 1
	p.scaleY = 1

	p.positioning = positionRelative

	return p
}

// SetFont sets the Paragraph's font.
func (p *Paragraph) SetFont(font *model.PdfFont) {
	p.textFont = font
}

// SetFontSize sets the font size in document units (points).
func (p *Paragraph) SetFontSize(fontSize float64) {
	p.fontSize = fontSize
}

// SetTextAlignment sets the horizontal alignment of the text within the space provided.
func (p *Paragraph) SetTextAlignment(align TextAlignment) {
	p.alignment = align
}

// SetEncoder sets the text encoding.
func (p *Paragraph) SetEncoder(encoder textencoding.TextEncoder) {
	p.textFont.SetEncoder(encoder)
}

// SetLineHeight sets the line height (1.0 default).
func (p *Paragraph) SetLineHeight(lineheight float64) {
	p.lineHeight = lineheight
}

// SetText sets the text content of the Paragraph.
func (p *Paragraph) SetText(text string) {
	p.text = text
}

// Text sets the text content of the Paragraph.
func (p *Paragraph) Text() string {
	return p.text
}

// SetEnableWrap sets the line wrapping enabled flag.
func (p *Paragraph) SetEnableWrap(enableWrap bool) {
	p.enableWrap = enableWrap
}

// SetColor sets the color of the Paragraph text.
//
// Example:
// 1.   p := NewParagraph("Red paragraph")
//      // Set to red color with a hex code:
//      p.SetColor(creator.ColorRGBFromHex("#ff0000"))
//
// 2. Make Paragraph green with 8-bit rgb values (0-255 each component)
//      p.SetColor(creator.ColorRGBFrom8bit(0, 255, 0)
//
// 3. Make Paragraph blue with arithmetic (0-1) rgb components.
//      p.SetColor(creator.ColorRGBFromArithmetic(0, 0, 1.0)
//
func (p *Paragraph) SetColor(col Color) {
	pdfColor := model.NewPdfColorDeviceRGB(col.ToRGB())
	p.color = *pdfColor
}

// SetPos sets absolute positioning with specified coordinates.
func (p *Paragraph) SetPos(x, y float64) {
	p.positioning = positionAbsolute
	p.xPos = x
	p.yPos = y
}

// SetAngle sets the rotation angle of the text.
func (p *Paragraph) SetAngle(angle float64) {
	p.angle = angle
}

// SetMargins sets the Paragraph's margins.
func (p *Paragraph) SetMargins(left, right, top, bottom float64) {
	p.margins.left = left
	p.margins.right = right
	p.margins.top = top
	p.margins.bottom = bottom
}

// GetMargins returns the Paragraph's margins: left, right, top, bottom.
func (p *Paragraph) GetMargins() (float64, float64, float64, float64) {
	return p.margins.left, p.margins.right, p.margins.top, p.margins.bottom
}

// SetWidth sets the the Paragraph width. This is essentially the wrapping width, i.e. the width the
// text can extend to prior to wrapping over to next line.
func (p *Paragraph) SetWidth(width float64) {
	p.wrapWidth = width
	p.enableWrap = true
	p.wrapText()
}

// Width returns the width of the Paragraph.
func (p *Paragraph) Width() float64 {
	if p.enableWrap {
		return p.wrapWidth
	}
	return p.getTextWidth() / 1000.0
}

// Height returns the height of the Paragraph. The height is calculated based on the input text and
// how it is wrapped within the container. Does not include Margins.
func (p *Paragraph) Height() float64 {
	if p.textLines == nil || len(p.textLines) == 0 {
		p.wrapText()
	}

	return float64(len(p.textLines)) * p.lineHeight * p.fontSize
}

// Calculate the text width (if not wrapped).
func (p *Paragraph) getTextWidth() float64 {
	w := 0.0

	for _, r := range p.text {
		glyph, found := p.textFont.Encoder().RuneToGlyph(r)
		if !found {
			common.Log.Debug("ERROR: Glyph not found for rune: 0x%04x=%c", r, r)
			return -1 // XXX/FIXME: return error.
		}

		metrics, found := p.textFont.GetGlyphCharMetrics(glyph)
		if !found {
			common.Log.Debug("ERROR: Glyph char metrics not found! %q (rune 0x%04x=%c)", glyph, r, r)
			return -1 // XXX/FIXME: return error.
		}
		w += p.fontSize * metrics.Wx
	}

	return w
}

// Simple algorithm to wrap the text into lines (greedy algorithm - fill the lines).
// XXX/TODO: Consider the Knuth/Plass algorithm or an alternative.
func (p *Paragraph) wrapText() error {
	if !p.enableWrap {
		p.textLines = []string{p.text}
		return nil
	}

	line := []rune{}
	lineWidth := 0.0
	p.textLines = []string{}

	runes := []rune(p.text)
	glyphs := []string{}
	widths := []float64{}

	for _, val := range runes {
		glyph, found := p.textFont.Encoder().RuneToGlyph(val)
		if !found {
			common.Log.Debug("ERROR: Glyph not found for rune: %c", val)
			return errors.New("Glyph not found for rune")
		}

		metrics, found := p.textFont.GetGlyphCharMetrics(glyph)
		if !found {
			common.Log.Debug("ERROR: Glyph char metrics not found! %q rune=0x%04x=%c font=%s %#q",
				glyph, val, val, p.textFont.BaseFont(), p.textFont.Subtype())
			common.Log.Trace("Font: %#v", p.textFont)
			common.Log.Trace("Encoder: %#v", p.textFont.Encoder())
			return errors.New("Glyph char metrics missing")
		}

		w := p.fontSize * metrics.Wx
		if lineWidth+w > p.wrapWidth*1000.0 {
			// Goes out of bounds: Wrap.
			// Breaks on the character.
			idx := -1
			for i := len(glyphs) - 1; i >= 0; i-- {
				if glyphs[i] == "space" { // XXX: What about other space glyphs like controlHT?
					idx = i
					break
				}
			}
			if idx > 0 {
				// Back up to last space.
				p.textLines = append(p.textLines, string(line[0:idx+1]))

				// Remainder of line.
				line = append(line[idx+1:], val)
				glyphs = append(glyphs[idx+1:], glyph)
				widths = append(widths[idx+1:], w)
				lineWidth = sum(widths)

			} else {
				p.textLines = append(p.textLines, string(line))
				line = []rune{val}
				glyphs = []string{glyph}
				widths = []float64{w}
				lineWidth = w
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

// sum returns the sums of the elements in `widths`.
func sum(widths []float64) float64 {
	total := 0.0
	for _, w := range widths {
		total += w
	}
	return total
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages. Implements the Drawable interface.
func (p *Paragraph) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origContext := ctx
	blocks := []*Block{}

	blk := NewBlock(ctx.PageWidth, ctx.PageHeight)
	if p.positioning.isRelative() {
		// Account for Paragraph margins.
		ctx.X += p.margins.left
		ctx.Y += p.margins.top
		ctx.Width -= p.margins.left + p.margins.right
		ctx.Height -= p.margins.top + p.margins.bottom

		// Use available space.
		p.SetWidth(ctx.Width)

		if p.Height() > ctx.Height {
			// Goes out of the bounds.  Write on a new template instead and create a new context at
			// upper left corner.
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
	ctx, err := drawParagraphOnBlock(blk, p, ctx)
	if err != nil {
		common.Log.Debug("ERROR: %v", err)
		return nil, ctx, err
	}

	blocks = append(blocks, blk)
	if p.positioning.isRelative() {
		ctx.X -= p.margins.left // Move back.
		ctx.Width = origContext.Width
		return blocks, ctx, nil
	} else {
		// Absolute: not changing the context.
		return blocks, origContext, nil
	}
}

// drawParagraphOnBlock draws Paragraph `p` on Block `blk` at the specified location on the page,
// adding it to the content stream.
func drawParagraphOnBlock(blk *Block, p *Paragraph, ctx DrawContext) (DrawContext, error) {
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
		w := 0.0
		spaces := 0
		for i, r := range runes {
			glyph, found := p.textFont.Encoder().RuneToGlyph(r)
			if !found {
				common.Log.Debug("Rune 0x%x not supported by text encoder", r)
				return ctx, errors.New("Unsupported rune in text encoding")
			}
			if glyph == "space" {
				spaces++
				continue
			}
			metrics, found := p.textFont.GetGlyphCharMetrics(glyph)
			if !found {
				common.Log.Debug("Unsupported glyph %q i=%d rune=0x%04x=%c in font %s %s",
					glyph, i, r, r,
					p.textFont.BaseFont(), p.textFont.Subtype())
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
		switch p.alignment {
		case TextAlignmentJustify:
			if spaces > 0 && idx < len(p.textLines)-1 { // Not to justify last line.
				spaceWidth = (p.wrapWidth*1000.0 - w) / float64(spaces) / p.fontSize
			}
		case TextAlignmentCenter:
			// Start with a shift.
			textWidth := w + float64(spaces)*spaceWidth*p.fontSize
			shift := (p.wrapWidth*1000.0 - textWidth) / 2 / p.fontSize
			objs = append(objs, core.MakeFloat(-shift))
		case TextAlignmentRight:
			textWidth := w + float64(spaces)*spaceWidth*p.fontSize
			shift := (p.wrapWidth*1000.0 - textWidth) / p.fontSize
			objs = append(objs, core.MakeFloat(-shift))
		}

		encoded := []byte{}
		for _, r := range runes {
			glyph, ok := p.textFont.Encoder().RuneToGlyph(r)
			if !ok {
				common.Log.Debug("Rune 0x%x not supported by text encoder", r)
				return ctx, errors.New("Unsupported rune in text encoding")
			}

			if glyph == "space" { // XXX: What about \t and other spaces.
				if len(encoded) > 0 {
					objs = append(objs, core.MakeStringFromBytes(encoded))
					encoded = []byte{}
				}
				objs = append(objs, core.MakeFloat(-spaceWidth))
			} else {
				code, ok := p.textFont.Encoder().RuneToCharcode(r)
				if ok {
					encoded = append(encoded, byte(code))
				}
			}
		}
		if len(encoded) > 0 {
			objs = append(objs, core.MakeStringFromBytes(encoded))
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
