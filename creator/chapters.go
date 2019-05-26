/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// Chapter is used to arrange multiple drawables (paragraphs, images, etc) into a single section.
// The concept is the same as a book or a report chapter.
type Chapter struct {
	// The number of the chapter.
	number int

	// The title of the chapter.
	title string

	// The heading paragraph of the chapter.
	heading *Paragraph

	// The content components of the chapter.
	contents []Drawable

	// The number of subchapters the chapter has.
	subchapters int

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

	// Reference to the parent chapter the current chapter belongs to.
	parent *Chapter

	// Reference to the TOC of the creator.
	toc *TOC

	// Reference to the outline of the creator.
	outline *model.Outline

	// The item of the chapter in the outline.
	outlineItem *model.OutlineItem

	// The level of the chapter in the chapters hierarchy.
	level uint
}

// newChapter creates a new chapter with the specified title as the heading.
func newChapter(parent *Chapter, toc *TOC, outline *model.Outline, title string, number int, style TextStyle) *Chapter {
	var level uint = 1
	if parent != nil {
		level = parent.level + 1
	}

	chapter := &Chapter{
		number:        number,
		title:         title,
		showNumbering: true,
		includeInTOC:  true,
		parent:        parent,
		toc:           toc,
		outline:       outline,
		contents:      []Drawable{},
		level:         level,
	}

	p := newParagraph(chapter.headingText(), style)
	p.SetFont(style.Font)
	p.SetFontSize(style.FontSize)

	chapter.heading = p
	return chapter
}

// NewSubchapter creates a new child chapter with the specified title.
func (chap *Chapter) NewSubchapter(title string) *Chapter {
	style := newTextStyle(chap.heading.textFont)
	style.FontSize = 14

	chap.subchapters++
	subchapter := newChapter(chap, chap.toc, chap.outline, title, chap.subchapters, style)
	chap.Add(subchapter)

	return subchapter
}

// SetShowNumbering sets a flag to indicate whether or not to show chapter numbers as part of title.
func (chap *Chapter) SetShowNumbering(show bool) {
	chap.heading.SetText(chap.headingText())
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
		return errors.New("range check error")
	}

	switch d.(type) {
	case *Paragraph, *StyledParagraph, *Image, *Block, *Table, *PageBreak, *Chapter:
		chap.contents = append(chap.contents, d)
	default:
		common.Log.Debug("Unsupported: %T", d)
		return errors.New("type check error")
	}

	return nil
}

// headingNumber returns the chapter heading number based on the chapter
// hierarchy and the showNumbering property.
func (chap *Chapter) headingNumber() string {
	var chapNumber string
	if chap.showNumbering {
		if chap.number != 0 {
			chapNumber = strconv.Itoa(chap.number) + "."
		}

		if chap.parent != nil {
			parentChapNumber := chap.parent.headingNumber()
			if parentChapNumber != "" {
				chapNumber = parentChapNumber + chapNumber
			}
		}
	}

	return chapNumber
}

// headingText returns the chapter heading text content.
func (chap *Chapter) headingText() string {
	heading := chap.title
	if chapNumber := chap.headingNumber(); chapNumber != "" {
		heading = fmt.Sprintf("%s %s", chapNumber, heading)
	}

	return heading
}

// GeneratePageBlocks generate the Page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages.
func (chap *Chapter) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	if chap.positioning.isRelative() {
		// Update context.
		ctx.X += chap.margins.left
		ctx.Y += chap.margins.top
		ctx.Width -= chap.margins.left + chap.margins.right
		ctx.Height -= chap.margins.top
	}

	blocks, c, err := chap.heading.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}
	ctx = c

	// Generate chapter title and number.
	posX := ctx.X
	posY := ctx.Y - chap.heading.Height()
	page := int64(ctx.Page)

	chapNumber := chap.headingNumber()
	chapTitle := chap.headingText()

	// Add to TOC.
	if chap.includeInTOC {
		line := chap.toc.Add(chapNumber, chap.title, strconv.FormatInt(page, 10), chap.level)
		if chap.toc.showLinks {
			line.SetLink(page, posX, posY)
		}
	}

	// Add to outline.
	if chap.outlineItem == nil {
		chap.outlineItem = model.NewOutlineItem(
			chapTitle,
			model.NewOutlineDest(page-1, posX, posY),
		)

		if chap.parent != nil {
			chap.parent.outlineItem.Add(chap.outlineItem)
		} else {
			chap.outline.Add(chap.outlineItem)
		}
	} else {
		outlineDest := &chap.outlineItem.Dest
		outlineDest.Page = page - 1
		outlineDest.X = posX
		outlineDest.Y = posY
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
