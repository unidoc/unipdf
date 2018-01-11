/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

// Chapter is used to arrange multiple drawables (paragraphs, images, etc) into a single section. The concept is
// the same as a book or a report chapter.
type Chapter struct {
	number  int
	title   string
	heading *Paragraph

	subchapters int

	contents []Drawable

	// Show chapter numbering
	showNumbering bool

	// Include in TOC.
	includeInTOC bool

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos, yPos float64

	// Margins to be applied around the block when drawing on Page.
	margins margins

	// Reference to the creator's TOC.
	toc *TableOfContents
}

// NewChapter creates a new chapter with the specified title as the heading.
func (c *Creator) NewChapter(title string) *Chapter {
	chap := &Chapter{}

	c.chapters++
	chap.number = c.chapters
	chap.title = title

	chap.showNumbering = true
	chap.includeInTOC = true

	heading := fmt.Sprintf("%d. %s", c.chapters, title)
	p := NewParagraph(heading)
	p.SetFontSize(16)
	p.SetFont(fonts.NewFontHelvetica()) // bold?

	chap.heading = p
	chap.contents = []Drawable{}

	// Keep a reference for toc.
	chap.toc = c.toc

	return chap
}

// SetShowNumbering sets a flag to indicate whether or not to show chapter numbers as part of title.
func (chap *Chapter) SetShowNumbering(show bool) {
	if show {
		heading := fmt.Sprintf("%d. %s", chap.number, chap.title)
		chap.heading.SetText(heading)
	} else {
		heading := fmt.Sprintf("%s", chap.title)
		chap.heading.SetText(heading)
	}
	chap.showNumbering = show
}

// SetIncludeInTOC sets a flag to indicate whether or not to include in tOC.
func (chap *Chapter) SetIncludeInTOC(includeInTOC bool) {
	chap.includeInTOC = includeInTOC
}

// GetHeading returns the chapter heading paragraph. Used to give access to address style: font, sizing etc.
func (chap *Chapter) GetHeading() *Paragraph {
	return chap.heading
}

// SetMargins sets the Chapter margins: left, right, top, bottom.
// Typically not needed as the creator's page margins are used.
func (chap *Chapter) SetMargins(left, right, top, bottom float64) {
	chap.margins.left = left
	chap.margins.right = right
	chap.margins.top = top
	chap.margins.bottom = bottom
}

// GetMargins returns the Chapter's margin: left, right, top, bottom.
func (chap *Chapter) GetMargins() (float64, float64, float64, float64) {
	return chap.margins.left, chap.margins.right, chap.margins.top, chap.margins.bottom
}

// Add adds a new Drawable to the chapter.
func (chap *Chapter) Add(d Drawable) error {
	if Drawable(chap) == d {
		common.Log.Debug("ERROR: Cannot add itself")
		return errors.New("Range check error")
	}

	switch d.(type) {
	case *Chapter:
		common.Log.Debug("Error: Cannot add chapter to a chapter")
		return errors.New("Type check error")
	case *Paragraph, *Image, *Block, *Subchapter, *Table:
		chap.contents = append(chap.contents, d)
	default:
		common.Log.Debug("Unsupported: %T", d)
		return errors.New("Type check error")
	}

	return nil
}

// GeneratePageBlocks generate the Page blocks.  Multiple blocks are generated if the contents wrap over
// multiple pages.
func (chap *Chapter) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	if chap.positioning.isRelative() {
		// Update context.
		ctx.X += chap.margins.left
		ctx.Y += chap.margins.top
		ctx.Width -= chap.margins.left + chap.margins.right
		ctx.Height -= chap.margins.top
	}

	blocks, ctx, err := chap.heading.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}
	if len(blocks) > 1 {
		ctx.Page++ // Did not fit, moved to new Page block.
	}

	if chap.includeInTOC {
		// Add to TOC.
		chap.toc.add(chap.title, chap.number, 0, ctx.Page)
	}

	for _, d := range chap.contents {
		newBlocks, c, err := d.GeneratePageBlocks(ctx)
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

	if chap.positioning.isRelative() {
		// Move back X to same start of line.
		ctx.X = origCtx.X
	}

	if chap.positioning.isAbsolute() {
		// If absolute: return original context.
		return blocks, origCtx, nil

	}

	return blocks, ctx, nil
}
