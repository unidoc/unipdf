/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

// Subchapter simply represents a sub chapter pertaining to a specific Chapter.  It can contain multiple
// Drawables, just like a chapter.
type Subchapter struct {
	chapterNum    int
	subchapterNum int
	title         string
	heading       *Paragraph

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

// NewSubchapter creates a new Subchapter under Chapter ch with specified title.
// All other parameters are set to their defaults.
func (c *Creator) NewSubchapter(ch *Chapter, title string) *Subchapter {
	subchap := &Subchapter{}

	ch.subchapters++
	subchap.subchapterNum = ch.subchapters

	subchap.chapterNum = ch.number
	subchap.title = title

	heading := fmt.Sprintf("%d.%d %s", subchap.chapterNum, subchap.subchapterNum, title)
	p := NewParagraph(heading)

	p.SetFontSize(14)
	p.SetFont(fonts.NewFontHelvetica()) // bold?

	subchap.showNumbering = true
	subchap.includeInTOC = true

	subchap.heading = p
	subchap.contents = []Drawable{}

	// Add subchapter to ch.
	ch.Add(subchap)

	// Keep a reference for toc.
	subchap.toc = c.toc

	return subchap
}

// SetShowNumbering sets a flag to indicate whether or not to show chapter numbers as part of title.
func (subchap *Subchapter) SetShowNumbering(show bool) {
	if show {
		heading := fmt.Sprintf("%d.%d. %s", subchap.chapterNum, subchap.subchapterNum, subchap.title)
		subchap.heading.SetText(heading)
	} else {
		heading := fmt.Sprintf("%s", subchap.title)
		subchap.heading.SetText(heading)
	}
	subchap.showNumbering = show
}

// SetIncludeInTOC sets a flag to indicate whether or not to include in the table of contents.
func (subchap *Subchapter) SetIncludeInTOC(includeInTOC bool) {
	subchap.includeInTOC = includeInTOC
}

// GetHeading returns the Subchapter's heading Paragraph to address style (font type, size, etc).
func (subchap *Subchapter) GetHeading() *Paragraph {
	return subchap.heading
}

// Set absolute coordinates.
/*
func (subchap *subchapter) SetPos(x, y float64) {
	subchap.positioning = positionAbsolute
	subchap.xPos = x
	subchap.yPos = y
}
*/

// SetMargins sets the Subchapter's margins (left, right, top, bottom).
// These margins are typically not needed as the Creator's page margins are used preferably.
func (subchap *Subchapter) SetMargins(left, right, top, bottom float64) {
	subchap.margins.left = left
	subchap.margins.right = right
	subchap.margins.top = top
	subchap.margins.bottom = bottom
}

// GetMargins returns the Subchapter's margins: left, right, top, bottom.
func (subchap *Subchapter) GetMargins() (float64, float64, float64, float64) {
	return subchap.margins.left, subchap.margins.right, subchap.margins.top, subchap.margins.bottom
}

// Add adds a new Drawable to the chapter.
// The currently supported Drawables are: *Paragraph, *Image, *Block, *Table.
func (subchap *Subchapter) Add(d Drawable) {
	switch d.(type) {
	case *Chapter, *Subchapter:
		common.Log.Debug("Error: Cannot add chapter or subchapter to a subchapter")
	case *Paragraph, *Image, *Block, *Table:
		subchap.contents = append(subchap.contents, d)
	default:
		common.Log.Debug("Unsupported: %T", d)
	}
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated if the contents wrap over
// multiple pages. Implements the Drawable interface.
func (subchap *Subchapter) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	if subchap.positioning.isRelative() {
		// Update context.
		ctx.X += subchap.margins.left
		ctx.Y += subchap.margins.top
		ctx.Width -= subchap.margins.left + subchap.margins.right
		ctx.Height -= subchap.margins.top
	}

	blocks, ctx, err := subchap.heading.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}
	if len(blocks) > 1 {
		ctx.Page++ // did not fit - moved to next Page.
	}
	if subchap.includeInTOC {
		// Add to TOC.
		subchap.toc.add(subchap.title, subchap.chapterNum, subchap.subchapterNum, ctx.Page)
	}

	for _, d := range subchap.contents {
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

	if subchap.positioning.isRelative() {
		// Move back X to same start of line.
		ctx.X = origCtx.X
	}

	if subchap.positioning.isAbsolute() {
		// If absolute: return original context.
		return blocks, origCtx, nil

	}

	return blocks, ctx, nil
}
