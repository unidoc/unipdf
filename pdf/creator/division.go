/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"

	"github.com/unidoc/unidoc/common"
)

// Division is a container component which can wrap across multiple pages (unlike Block).
// It can contain multiple Drawable components (currently supporting Paragraph and Image).
//
// The component stacking behavior is vertical, where the Drawables are drawn on top of each other.
// TODO: Add inline mode (horizontal stacking).
type Division struct {
	components []VectorDrawable

	// Positioning: relative / absolute.
	positioning positioning

	// Margins to be applied around the block when drawing on Page.
	margins margins

	// Controls whether the components are stacked horizontally
	inline bool
}

// NewDivision returns a new Division container component.
func NewDivision() *Division {
	return &Division{
		components: []VectorDrawable{},
	}
}

func (div *Division) IsInline() bool {
	return div.inline
}

func (div *Division) SetInline(inline bool) {
	div.inline = inline
}

// Add adds a VectorDrawable to the Division container.
// Currently supported VectorDrawables: *Paragraph, *Image.
func (div *Division) Add(d VectorDrawable) error {
	supported := false

	switch d.(type) {
	case *Paragraph:
		supported = true
	case *Image:
		supported = true
	}

	if !supported {
		return errors.New("Unsupported type in Division")
	}

	div.components = append(div.components, d)

	return nil
}

// Height returns the height for the Division component assuming all stacked on top of each other.
func (div *Division) Height() float64 {
	y := 0.0
	yMax := 0.0
	for _, component := range div.components {
		compWidth, compHeight := component.Width(), component.Height()
		switch t := component.(type) {
		case *Paragraph:
			p := t
			compWidth += p.margins.left + p.margins.right
			compHeight += p.margins.top + p.margins.bottom
		}

		// Vertical stacking.
		y += compHeight
		yMax = y
	}

	return yMax
}

// Width is not used. Not used as a Division element is designed to fill into available width depending on
// context.  Returns 0.
func (div *Division) Width() float64 {
	return 0
}

// GeneratePageBlocks generates the page blocks for the Division component.
// Multiple blocks are generated if the contents wrap over multiple pages.
func (div *Division) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	pageblocks := []*Block{}

	origCtx := ctx

	if div.positioning.isRelative() {
		// Update context.
		ctx.X += div.margins.left
		ctx.Y += div.margins.top
		ctx.Width -= div.margins.left + div.margins.right
		ctx.Height -= div.margins.top
	}

	divCtx := ctx
	tmpCtx := ctx
	var lineHeight float64

	// Draw.
	for _, component := range div.components {
		if div.inline {
			if (ctx.X-divCtx.X)+component.Width() <= ctx.Width {
				ctx.Y = tmpCtx.Y
				ctx.Height = tmpCtx.Height
			} else {
				ctx.X = divCtx.X
				ctx.Width = divCtx.Width

				tmpCtx.Y += lineHeight
				tmpCtx.Height -= lineHeight
				lineHeight = 0
			}
		}

		newblocks, updCtx, err := component.GeneratePageBlocks(ctx)
		if err != nil {
			common.Log.Debug("Error generating page blocks: %v", err)
			return nil, ctx, err
		}

		if len(newblocks) < 1 {
			continue
		}

		if len(pageblocks) > 0 {
			// If there are pageblocks already in place.
			// merge the first block in with current Block and append the rest.
			pageblocks[len(pageblocks)-1].mergeBlocks(newblocks[0])
			pageblocks = append(pageblocks, newblocks[1:]...)
		} else {
			pageblocks = append(pageblocks, newblocks[0:]...)
		}

		// Apply padding/margins.
		if !div.inline {
			updCtx.X = ctx.X
		} else {
			if dl := ctx.Height - updCtx.Height; dl > lineHeight {
				lineHeight = dl
			}

			if ctx.Page != updCtx.Page {
				tmpCtx.Y = divCtx.Y
				tmpCtx.Height = divCtx.Height
				lineHeight = 0
			}
		}

		ctx = updCtx
	}

	if div.positioning.isRelative() {
		// Move back X to same start of line.
		ctx.X = origCtx.X
	}

	if div.positioning.isAbsolute() {
		// If absolute: return original context.
		return pageblocks, origCtx, nil
	}

	return pageblocks, ctx, nil
}
