/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"

	"math"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

type chapter struct {
	number  int
	title   string
	heading *paragraph

	subchapters int

	contents []Drawable

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos, yPos float64

	// Margins to be applied around the block when drawing on page.
	margins margins

	// Chapter sizing is set to occupy available space.
	sizing Sizing
}

func (c *Creator) NewChapter(title string) *chapter {
	chap := &chapter{}

	c.chapters++
	chap.number = c.chapters
	chap.title = title

	heading := fmt.Sprintf("%d. %s", c.chapters, title)
	p := NewParagraph(heading)

	p.SetFontSize(16)
	p.SetFont(fonts.NewFontHelvetica()) // bold?

	chap.heading = p
	chap.contents = []Drawable{}

	// Chapter sizing is fixed to occupy available size.
	chap.sizing = SizingOccupyAvailableSpace

	return chap
}

// Chapter sizing is fixed to occupy available space in the drawing context.
func (chap *chapter) GetSizingMechanism() Sizing {
	return chap.sizing
}

// Chapter height is a sum of the content heights.
func (chap *chapter) Height() float64 {
	h := float64(0)
	for _, d := range chap.contents {
		h += d.Height()
	}
	return h
}

// Chapter width is the maximum of the content widths.
func (chap *chapter) Width() float64 {
	maxW := float64(0)
	for _, d := range chap.contents {
		maxW = math.Max(maxW, d.Width())
	}
	return maxW
}

// Set absolute coordinates.
func (chap *chapter) SetPos(x, y float64) {
	chap.positioning = positionAbsolute
	chap.xPos = x
	chap.yPos = y
}

// Set chapter margins.  Typically not needed as the page margins are used.
func (chap *chapter) SetMargins(left, right, top, bottom float64) {
	chap.margins.left = left
	chap.margins.right = right
	chap.margins.top = top
	chap.margins.bottom = bottom
}

// Get chapter margins: left, right, top, bottom.
func (chap *chapter) GetMargins() (float64, float64, float64, float64) {
	return chap.margins.left, chap.margins.right, chap.margins.top, chap.margins.bottom
}

// Add a new drawable to the chapter.
func (chap *chapter) Add(d Drawable) {
	if Drawable(chap) == d {
		common.Log.Debug("ERROR: Cannot add itself")
		return
	}

	switch d.(type) {
	case *chapter:
		common.Log.Debug("Error: Cannot add chapter to a chapter")
	case *paragraph, *image, *block, *subchapter:
		chap.contents = append(chap.contents, d)
	default:
		common.Log.Debug("Unsupported: %T", d)
	}
}

// XXX/FIXME: Need to know actual page numbers to keep track for TOC.
// TODO: Add page number to context... ?
func (chap *chapter) GeneratePageBlocks(ctx drawContext) ([]*block, drawContext, error) {
	blocks, ctx, err := chap.heading.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
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

	return blocks, ctx, nil
}
