/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"

	"github.com/unidoc/unipdf/v3/common"
)

// Division is a container component which can wrap across multiple pages (unlike Block).
// It can contain multiple Drawable components (currently supporting Paragraph and Image).
//
// The component stacking behavior is vertical, where the Drawables are drawn on top of each other.
// Also supports horizontal stacking by activating the inline mode.
type Division struct {
	components []VectorDrawable

	// Positioning: relative / absolute.
	positioning positioning

	// Margins to be applied around the block when drawing on Page.
	margins margins

	// Controls whether the components are stacked horizontally
	inline bool
}

// newDivision returns a new Division container component.
func newDivision() *Division {
	return &Division{
		components: []VectorDrawable{},
	}
}

// Inline returns whether the inline mode of the division is active.
func (div *Division) Inline() bool {
	return div.inline
}

// SetInline sets the inline mode of the division.
func (div *Division) SetInline(inline bool) {
	div.inline = inline
}

// Add adds a VectorDrawable to the Division container.
// Currently supported VectorDrawables: *Paragraph, *StyledParagraph, *Image.
func (div *Division) Add(d VectorDrawable) error {
	supported := false

	switch d.(type) {
	case *Paragraph:
		supported = true
	case *StyledParagraph:
		supported = true
	case *Image:
		supported = true
	}

	if !supported {
		return errors.New("unsupported type in Division")
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
		case *StyledParagraph:
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
	var pageblocks []*Block

	origCtx := ctx

	if div.positioning.isRelative() {
		// Update context.
		ctx.X += div.margins.left
		ctx.Y += div.margins.top
		ctx.Width -= div.margins.left + div.margins.right
		ctx.Height -= div.margins.top
	}

	// Set the inline mode of the division to the context.
	ctx.Inline = div.inline

	// Draw.
	divCtx := ctx
	tmpCtx := ctx
	var lineHeight float64

	for _, component := range div.components {
		if ctx.Inline {
			// Check whether the component fits on the current line.
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
		if ctx.Inline {
			// Recalculate positions on page change.
			if ctx.Page != updCtx.Page {
				divCtx.Y = ctx.Margins.top
				divCtx.Height = ctx.PageHeight - ctx.Margins.top

				tmpCtx.Y = divCtx.Y
				tmpCtx.Height = divCtx.Height
				lineHeight = updCtx.Height - divCtx.Height
			} else {
				// Calculate current line max height.
				if dl := ctx.Height - updCtx.Height; dl > lineHeight {
					lineHeight = dl
				}
			}
		} else {
			updCtx.X = ctx.X
		}

		ctx = updCtx
	}

	// Restore the original inline mode of the context.
	ctx.Inline = origCtx.Inline

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
