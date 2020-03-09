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

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// StyledParagraph represents text drawn with a specified font and can wrap across lines and pages.
// By default occupies the available width in the drawing context.
type StyledParagraph struct {
	// Text chunks with styles that compose the paragraph.
	chunks []*TextChunk

	// Style used for the paragraph for spacing and offsets.
	defaultStyle TextStyle

	// Style used for the paragraph link annotations.
	defaultLinkStyle TextStyle

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
	lines [][]*TextChunk

	// Before render callback.
	beforeRender func(p *StyledParagraph, ctx DrawContext)
}

// newStyledParagraph creates a new styled paragraph.
func newStyledParagraph(style TextStyle) *StyledParagraph {
	// TODO: Can we wrap intellectually, only if given width is known?
	return &StyledParagraph{
		chunks:           []*TextChunk{},
		defaultStyle:     style,
		defaultLinkStyle: newLinkStyle(style.Font),
		lineHeight:       1.0,
		alignment:        TextAlignmentLeft,
		enableWrap:       true,
		defaultWrap:      true,
		angle:            0,
		scaleX:           1,
		scaleY:           1,
		positioning:      positionRelative,
	}
}

// appendChunk adds the provided text chunk to the paragraph.
func (p *StyledParagraph) appendChunk(chunk *TextChunk) *TextChunk {
	p.chunks = append(p.chunks, chunk)
	p.wrapText()
	return chunk
}

// Append adds a new text chunk to the paragraph.
func (p *StyledParagraph) Append(text string) *TextChunk {
	chunk := NewTextChunk(text, p.defaultStyle)
	return p.appendChunk(chunk)
}

// Insert adds a new text chunk at the specified position in the paragraph.
func (p *StyledParagraph) Insert(index uint, text string) *TextChunk {
	l := uint(len(p.chunks))
	if index > l {
		index = l
	}

	chunk := NewTextChunk(text, p.defaultStyle)
	p.chunks = append(p.chunks[:index], append([]*TextChunk{chunk}, p.chunks[index:]...)...)
	p.wrapText()

	return chunk
}

// AddExternalLink adds a new external link to the paragraph.
// The text parameter represents the text that is displayed and the url
// parameter sets the destionation of the link.
func (p *StyledParagraph) AddExternalLink(text, url string) *TextChunk {
	chunk := NewTextChunk(text, p.defaultLinkStyle)
	chunk.annotation = newExternalLinkAnnotation(url)
	return p.appendChunk(chunk)
}

// AddInternalLink adds a new internal link to the paragraph.
// The text parameter represents the text that is displayed.
// The user is taken to the specified page, at the specified x and y
// coordinates. Position 0, 0 is at the top left of the page.
// The zoom of the destination page is controlled with the zoom
// parameter. Pass in 0 to keep the current zoom value.
func (p *StyledParagraph) AddInternalLink(text string, page int64, x, y, zoom float64) *TextChunk {
	chunk := NewTextChunk(text, p.defaultLinkStyle)
	chunk.annotation = newInternalLinkAnnotation(page-1, x, y, zoom)
	return p.appendChunk(chunk)
}

// Reset removes all the text chunks the paragraph contains.
func (p *StyledParagraph) Reset() {
	p.chunks = []*TextChunk{}
}

// SetText replaces all the text of the paragraph with the specified one.
func (p *StyledParagraph) SetText(text string) *TextChunk {
	p.Reset()
	return p.Append(text)
}

// SetTextAlignment sets the horizontal alignment of the text within the space provided.
func (p *StyledParagraph) SetTextAlignment(align TextAlignment) {
	p.alignment = align
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
	if p.enableWrap && int(p.wrapWidth) > 0 {
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

// getLineHeight returns both the capheight and the font size based height of
// the line with the specified index.
func (p *StyledParagraph) getLineHeight(idx int) (capHeight, height float64) {
	if p.lines == nil || len(p.lines) == 0 {
		p.wrapText()
	}
	if idx < 0 || idx > len(p.lines)-1 {
		common.Log.Debug("ERROR: invalid paragraph line index %d. Returning 0, 0", idx)
		return 0, 0
	}

	line := p.lines[idx]
	for _, chunk := range line {
		descriptor, err := chunk.Style.Font.GetFontDescriptor()
		if err != nil {
			common.Log.Debug("ERROR: Unable to get font descriptor")
		}

		var fontCapHeight float64
		if descriptor != nil {
			if fontCapHeight, err = descriptor.GetCapHeight(); err != nil {
				common.Log.Debug("ERROR: Unable to get font CapHeight: %v", err)
			}
		}
		if int(fontCapHeight) <= 0 {
			common.Log.Debug("WARN: CapHeight not available - setting to 1000")
			fontCapHeight = 1000
		}

		h := fontCapHeight / 1000.0 * chunk.Style.FontSize * p.lineHeight
		if h > capHeight {
			capHeight = h
		}

		h = p.lineHeight * chunk.Style.FontSize
		if h > height {
			height = h
		}
	}

	return capHeight, height
}

// getTextWidth calculates the text width as if all in one line (not taking
// wrapping into account).
func (p *StyledParagraph) getTextWidth() float64 {
	var width float64
	lenChunks := len(p.chunks)

	for i, chunk := range p.chunks {
		style := &chunk.Style
		lenRunes := len(chunk.Text)

		for j, r := range chunk.Text {
			// Ignore newline for this. Handles as if all in one line.
			if r == '\u000A' { // LF
				continue
			}

			metrics, found := style.Font.GetRuneMetrics(r)
			if !found {
				common.Log.Debug("Rune char metrics not found! %v\n", r)

				// FIXME: return error.
				return -1
			}

			width += style.FontSize * metrics.Wx

			// Do not add character spacing for the last character of the line.
			if i != lenChunks-1 || j != lenRunes-1 {
				width += style.CharSpacing * 1000.0
			}
		}
	}

	return width
}

// getTextLineWidth calculates the text width of a provided collection of text chunks.
func (p *StyledParagraph) getTextLineWidth(line []*TextChunk) float64 {
	var width float64
	lenChunks := len(line)

	for i, chunk := range line {
		style := &chunk.Style
		lenRunes := len(chunk.Text)

		for j, r := range chunk.Text {
			// Ignore newline for this. Handles as if all in one line.
			if r == '\u000A' { // LF
				continue
			}

			metrics, found := style.Font.GetRuneMetrics(r)
			if !found {
				common.Log.Debug("Rune char metrics not found! %v\n", r)

				// FIXME: return error.
				return -1
			}

			width += style.FontSize * metrics.Wx

			// Do not add character spacing for the last character of the line.
			if i != lenChunks-1 || j != lenRunes-1 {
				width += style.CharSpacing * 1000.0
			}
		}
	}

	return width
}

// getMaxLineWidth returns the width of the longest line of text in the paragraph.
func (p *StyledParagraph) getMaxLineWidth() float64 {
	if p.lines == nil || len(p.lines) == 0 {
		p.wrapText()
	}

	var width float64
	for _, line := range p.lines {
		w := p.getTextLineWidth(line)
		if w > width {
			width = w
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
// TODO: Consider the Knuth/Plass algorithm or an alternative.
func (p *StyledParagraph) wrapText() error {
	if !p.enableWrap || int(p.wrapWidth) <= 0 {
		p.lines = [][]*TextChunk{p.chunks}
		return nil
	}

	p.lines = [][]*TextChunk{}
	var line []*TextChunk
	var lineWidth float64

	copyAnnotation := func(src *model.PdfAnnotation) *model.PdfAnnotation {
		if src == nil {
			return nil
		}

		var annotation *model.PdfAnnotation
		switch t := src.GetContext().(type) {
		case *model.PdfAnnotationLink:
			if annot := copyLinkAnnotation(t); annot != nil {
				annotation = annot.PdfAnnotation
			}
		}

		return annotation
	}

	for _, chunk := range p.chunks {
		style := chunk.Style
		annotation := chunk.annotation

		var (
			part   []rune
			widths []float64
		)

		for _, r := range chunk.Text {
			// newline wrapping.
			if r == '\u000A' { // LF
				// moves to next line.
				line = append(line, &TextChunk{
					Text:       strings.TrimRightFunc(string(part), unicode.IsSpace),
					Style:      style,
					annotation: copyAnnotation(annotation),
				})
				p.lines = append(p.lines, line)
				line = nil

				lineWidth = 0
				part = nil
				widths = nil
				continue
			}

			metrics, found := style.Font.GetRuneMetrics(r)
			if !found {
				common.Log.Debug("Rune char metrics not found! %v\n", r)
				return errors.New("glyph char metrics missing")
			}

			w := style.FontSize * metrics.Wx
			charWidth := w + style.CharSpacing*1000.0

			if lineWidth+w > p.wrapWidth*1000.0 {
				// Goes out of bounds: Wrap.
				// Breaks on the character.
				// TODO: when goes outside: back up to next space,
				// otherwise break on the character.
				idx := -1
				for j := len(part) - 1; j >= 0; j-- {
					if part[j] == ' ' {
						idx = j
						break
					}
				}

				text := string(part)
				if idx >= 0 {
					text = string(part[0 : idx+1])

					part = part[idx+1:]
					part = append(part, r)
					widths = widths[idx+1:]
					widths = append(widths, charWidth)

					lineWidth = 0
					for _, width := range widths {
						lineWidth += width
					}
				} else {
					lineWidth = charWidth
					part = []rune{r}
					widths = []float64{charWidth}
				}

				line = append(line, &TextChunk{
					Text:       strings.TrimRightFunc(string(text), unicode.IsSpace),
					Style:      style,
					annotation: copyAnnotation(annotation),
				})
				p.lines = append(p.lines, line)
				line = []*TextChunk{}
			} else {
				lineWidth += charWidth
				part = append(part, r)
				widths = append(widths, charWidth)
			}
		}

		if len(part) > 0 {
			line = append(line, &TextChunk{
				Text:       string(part),
				Style:      style,
				annotation: copyAnnotation(annotation),
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
	var blocks []*Block

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

	if p.beforeRender != nil {
		p.beforeRender(p, ctx)
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
	// Find first free index for the font resources of the paragraph.
	num := 1
	fontName := core.PdfObjectName(fmt.Sprintf("Font%d", num))
	for blk.resources.HasFontByName(fontName) {
		num++
		fontName = core.PdfObjectName(fmt.Sprintf("Font%d", num))
	}

	// Add default font to the page resources.
	err := blk.resources.SetFontByName(fontName, p.defaultStyle.Font.ToPdfObject())
	if err != nil {
		return ctx, err
	}
	num++

	defaultFontName := fontName
	defaultFontSize := p.defaultStyle.FontSize

	// Wrap the text into lines.
	p.wrapText()

	// Add the fonts of all chunks to the page resources.
	var fonts [][]core.PdfObjectName

	for _, line := range p.lines {
		var fontLine []core.PdfObjectName

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

	currY := yPos
	for idx, line := range p.lines {
		currX := ctx.X

		if idx != 0 {
			// Move to next line if not first.
			cc.Add_Tstar()
		}

		isLastLine := idx == len(p.lines)-1

		// Get width of the line (excluding spaces).
		var (
			width      float64
			height     float64
			spaceWidth float64
			spaces     uint
		)

		var chunkWidths []float64
		for _, chunk := range line {
			style := &chunk.Style

			if style.FontSize > height {
				height = style.FontSize
			}

			spaceMetrics, found := style.Font.GetRuneMetrics(' ')
			if !found {
				return ctx, errors.New("the font does not have a space glyph")
			}

			var chunkSpaces uint
			var chunkWidth float64
			lenChunk := len(chunk.Text)
			for i, r := range chunk.Text {
				if r == ' ' {
					chunkSpaces++
					continue
				}
				if r == '\u000A' { // LF
					continue
				}

				metrics, found := style.Font.GetRuneMetrics(r)
				if !found {
					common.Log.Debug("Unsupported rune %v in font\n", r)
					return ctx, errors.New("unsupported text glyph")
				}

				chunkWidth += style.FontSize * metrics.Wx

				// Do not add character spacing for the last character of the line.
				if i != lenChunk-1 {
					chunkWidth += style.CharSpacing * 1000.0
				}
			}

			chunkWidths = append(chunkWidths, chunkWidth)
			width += chunkWidth

			spaceWidth += float64(chunkSpaces) * spaceMetrics.Wx * style.FontSize
			spaces += chunkSpaces
		}
		height *= p.lineHeight

		// Add line shifts.
		var objs []core.PdfObject

		wrapWidth := p.wrapWidth * 1000.0
		if p.alignment == TextAlignmentJustify {
			// Do not justify last line.
			if spaces > 0 && !isLastLine {
				spaceWidth = (wrapWidth - width) / float64(spaces) / defaultFontSize
			}
		} else if p.alignment == TextAlignmentCenter {
			// Start with an offset of half of the remaining line space.
			offset := (wrapWidth - width - spaceWidth) / 2
			shift := offset / defaultFontSize
			objs = append(objs, core.MakeFloat(-shift))

			currX += offset / 1000.0
		} else if p.alignment == TextAlignmentRight {
			// Push the text at the end of the line.
			offset := (wrapWidth - width - spaceWidth)
			shift := offset / defaultFontSize
			objs = append(objs, core.MakeFloat(-shift))

			currX += offset / 1000.0
		}

		if len(objs) > 0 {
			cc.Add_Tf(defaultFontName, defaultFontSize).
				Add_TL(defaultFontSize * p.lineHeight).
				Add_TJ(objs...)
		}

		// Render line text chunks.
		for k, chunk := range line {
			style := &chunk.Style

			r, g, b := style.Color.ToRGB()
			fontName := defaultFontName
			fontSize := defaultFontSize

			// Set chunk rendering mode.
			cc.Add_Tr(int64(style.RenderingMode))

			// Set chunk character spacing.
			cc.Add_Tc(style.CharSpacing)

			if p.alignment != TextAlignmentJustify || isLastLine {
				spaceMetrics, found := style.Font.GetRuneMetrics(' ')
				if !found {
					return ctx, errors.New("the font does not have a space glyph")
				}

				fontName = fonts[idx][k]
				fontSize = style.FontSize
				spaceWidth = spaceMetrics.Wx
			}
			enc := style.Font.Encoder()

			var encStr []byte
			for _, rn := range chunk.Text {
				if r == '\u000A' { // LF
					continue
				}
				if rn == ' ' {
					if len(encStr) > 0 {
						cc.Add_rg(r, g, b).
							Add_Tf(fonts[idx][k], style.FontSize).
							Add_TL(style.FontSize * p.lineHeight).
							Add_TJ([]core.PdfObject{core.MakeStringFromBytes(encStr)}...)

						encStr = nil
					}

					cc.Add_Tf(fontName, fontSize).
						Add_TL(fontSize * p.lineHeight).
						Add_TJ([]core.PdfObject{core.MakeFloat(-spaceWidth)}...)

					chunkWidths[k] += spaceWidth * fontSize
				} else {
					if _, ok := enc.RuneToCharcode(rn); !ok {
						common.Log.Debug("unsupported rune in text encoding: %#x (%c)", rn, rn)
						continue
					}
					encStr = append(encStr, enc.Encode(string(rn))...)
				}
			}

			if len(encStr) > 0 {
				cc.Add_rg(r, g, b).
					Add_Tf(fonts[idx][k], style.FontSize).
					Add_TL(style.FontSize * p.lineHeight).
					Add_TJ([]core.PdfObject{core.MakeStringFromBytes(encStr)}...)
			}

			chunkWidth := chunkWidths[k] / 1000.0

			// Add annotations.
			if chunk.annotation != nil {
				var annotRect *core.PdfObjectArray

				// Process annotation.
				if !chunk.annotationProcessed {
					switch t := chunk.annotation.GetContext().(type) {
					case *model.PdfAnnotationLink:
						// Initialize annotation rectangle.
						annotRect = core.MakeArray()
						t.Rect = annotRect

						// Reverse the Y axis of the destination coordinates.
						// The user passes in the annotation coordinates as if
						// position 0, 0 is at the top left of the page.
						// However, position 0, 0 in the PDF is at the bottom
						// left of the page.
						annotDest, ok := t.Dest.(*core.PdfObjectArray)
						if ok && annotDest.Len() == 5 {
							t, ok := annotDest.Get(1).(*core.PdfObjectName)
							if ok && t.String() == "XYZ" {
								y, err := core.GetNumberAsFloat(annotDest.Get(3))
								if err == nil {
									annotDest.Set(3, core.MakeFloat(ctx.PageHeight-y))
								}
							}
						}
					}

					chunk.annotationProcessed = true
				}

				// Set the coordinates of the annotation.
				if annotRect != nil {
					// Calculate rotated annotation position.
					annotPos := draw.NewPoint(currX-ctx.X, currY-yPos).Rotate(p.angle)
					annotPos.X += ctx.X
					annotPos.Y += yPos

					// Calculate rotated annotation bounding box.
					offX, offY, annotW, annotH := rotateRect(chunkWidth, height, p.angle)
					annotPos.X += offX
					annotPos.Y += offY

					annotRect.Clear()
					annotRect.Append(core.MakeFloat(annotPos.X))
					annotRect.Append(core.MakeFloat(annotPos.Y))
					annotRect.Append(core.MakeFloat(annotPos.X + annotW))
					annotRect.Append(core.MakeFloat(annotPos.Y + annotH))
				}

				blk.AddAnnotation(chunk.annotation)
			}

			currX += chunkWidth

			// Reset rendering mode.
			cc.Add_Tr(int64(TextRenderingModeFill))

			// Reset character spacing.
			cc.Add_Tc(0)
		}

		currY -= height
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
