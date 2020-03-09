/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
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
	scaleX, scaleY float64

	// Text lines after wrapping to available width.
	textLines []string
}

// newParagraph create a new text paragraph. Uses default parameters: Helvetica, WinAnsiEncoding and
// wrap enabled with a wrap width of 100 points.
//
// Standard font may will have an encdoing set to WinAnsiEncoding. To set a different encoding, make a new font
// and use SetFont on the paragraph to override the defaut one.
func newParagraph(text string, style TextStyle) *Paragraph {
	p := &Paragraph{
		text:        text,
		textFont:    style.Font,
		fontSize:    style.FontSize,
		lineHeight:  1.0,
		enableWrap:  true,
		defaultWrap: true,
		alignment:   TextAlignmentLeft,
		angle:       0,
		scaleX:      1,
		scaleY:      1,
		positioning: positionRelative,
	}

	p.SetColor(style.Color)
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
	p.defaultWrap = false
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
	p.wrapText()
}

// Width returns the width of the Paragraph.
func (p *Paragraph) Width() float64 {
	if p.enableWrap && int(p.wrapWidth) > 0 {
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

// getTextWidth calculates the text width as if all in one line (not taking wrapping into account).
func (p *Paragraph) getTextWidth() float64 {
	w := 0.0

	for _, r := range p.text {
		// Ignore newline for this.. Handles as if all in one line.
		if r == '\u000A' { // LF
			continue
		}

		metrics, found := p.textFont.GetRuneMetrics(r)
		if !found {
			common.Log.Debug("ERROR: Rune char metrics not found! (rune 0x%04x=%c)", r, r)
			return -1 // FIXME: return error.
		}
		w += p.fontSize * metrics.Wx
	}

	return w
}

// getTextLineWidth calculates the text width of a provided line of text.
func (p *Paragraph) getTextLineWidth(line string) float64 {
	var width float64
	for _, r := range line {
		// Ignore newline for this.. Handles as if all in one line.
		if r == '\u000A' { // LF
			continue
		}

		metrics, found := p.textFont.GetRuneMetrics(r)
		if !found {
			common.Log.Debug("ERROR: Rune char metrics not found! (rune 0x%04x=%c)", r, r)
			return -1 // FIXME: return error.
		}

		width += p.fontSize * metrics.Wx
	}

	return width
}

// getMaxLineWidth returns the width of the longest line of text in the paragraph.
func (p *Paragraph) getMaxLineWidth() float64 {
	if p.textLines == nil || len(p.textLines) == 0 {
		p.wrapText()
	}

	var width float64
	for _, line := range p.textLines {
		w := p.getTextLineWidth(line)
		if w > width {
			width = w
		}
	}

	return width
}

// Simple algorithm to wrap the text into lines (greedy algorithm - fill the lines).
// TODO: Consider the Knuth/Plass algorithm or an alternative.
func (p *Paragraph) wrapText() error {
	if !p.enableWrap || int(p.wrapWidth) <= 0 {
		p.textLines = []string{p.text}
		return nil
	}

	chunk := NewTextChunk(p.text, TextStyle{
		Font:     p.textFont,
		FontSize: p.fontSize,
	})

	lines, err := chunk.Wrap(p.wrapWidth)
	if err != nil {
		return err
	}

	p.textLines = lines
	return nil
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages. Implements the Drawable interface.
func (p *Paragraph) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origContext := ctx
	var blocks []*Block

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
			// TODO: Handle case when Paragraph is larger than the Page...
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
		if int(p.wrapWidth) <= 0 {
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
	}
	// Absolute: not changing the context.
	return blocks, origContext, nil
}

// drawParagraphOnBlock draws Paragraph `p` on Block `blk` at the specified location on the page,
// adding it to the content stream.
func drawParagraphOnBlock(blk *Block, p *Paragraph, ctx DrawContext) (DrawContext, error) {
	// Find a free name for the font.
	num := 1
	fontName := core.PdfObjectName("Font" + strconv.Itoa(num))
	for blk.resources.HasFontByName(fontName) {
		num++
		fontName = core.PdfObjectName("Font" + strconv.Itoa(num))
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
			if r == ' ' {
				spaces++
				continue
			}
			if r == '\u000A' { // LF
				continue
			}
			metrics, found := p.textFont.GetRuneMetrics(r)
			if !found {
				common.Log.Debug("Unsupported rune i=%d rune=0x%04x=%c in font %s %s",
					i, r, r,
					p.textFont.BaseFont(), p.textFont.Subtype())
				return ctx, errors.New("unsupported text glyph")
			}

			w += p.fontSize * metrics.Wx
		}

		var objs []core.PdfObject

		spaceMetrics, found := p.textFont.GetRuneMetrics(' ')
		if !found {
			return ctx, errors.New("the font does not have a space glyph")
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
		enc := p.textFont.Encoder()

		var encoded []byte
		for _, r := range runes {
			if r == '\u000A' { // LF
				continue
			}
			if r == ' ' { // TODO: What about \t and other spaces.
				if len(encoded) > 0 {
					objs = append(objs, core.MakeStringFromBytes(encoded))
					encoded = nil
				}
				objs = append(objs, core.MakeFloat(-spaceWidth))
			} else {
				if _, ok := enc.RuneToCharcode(r); !ok {
					common.Log.Debug("unsupported rune in text encoding: %#x (%c)", r, r)
					continue
				}
				encoded = append(encoded, enc.Encode(string(r))...)
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
