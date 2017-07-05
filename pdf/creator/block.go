/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"

	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// A block can contain a portion of PDF page contents. It has a width and a position and can
// be placed anywhere on a page.  It can even contain a whole page, and is used in the creator
// where each Drawable object can output one or more blocks, each representing content for separate pages
// (typically needed when page breaks occur).
type block struct {
	// Block contents and resources.
	contents  *contentstream.ContentStreamOperations
	resources *model.PdfPageResources

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos, yPos float64

	// The bounding box for the block.
	width  float64
	height float64

	// Rotation angle.
	angle float64

	// Margins to be applied around the block when drawing on page.
	margins margins
}

// Create a new block with specified width and height.
func NewBlock(width float64, height float64) *block {
	b := &block{}
	b.contents = &contentstream.ContentStreamOperations{}
	b.resources = model.NewPdfPageResources()
	b.width = width
	b.height = height
	return b
}

// Create a block from a PDF page.  Useful for loading template pages as blocks from a PDF document and additional
// content with the creator.
func NewBlockFromPage(page *model.PdfPage) (*block, error) {
	b := &block{}

	content, err := page.GetAllContentStreams()
	if err != nil {
		return nil, err
	}

	contentParser := contentstream.NewContentStreamParser(content)
	operations, err := contentParser.Parse()
	if err != nil {
		return nil, err
	}
	operations.WrapIfNeeded()

	b.contents = operations

	if page.Resources != nil {
		b.resources = page.Resources
	} else {
		b.resources = model.NewPdfPageResources()
	}

	mbox, err := page.GetMediaBox()
	if err != nil {
		return nil, err
	}

	if mbox.Llx != 0 || mbox.Lly != 0 {
		// Account for media box offset if any.
		b.translate(-mbox.Llx, mbox.Lly)
	}
	b.width = mbox.Urx - mbox.Llx
	b.height = mbox.Ury - mbox.Lly

	return b, nil
}

// Block sizing is always based on specified size.  Returns SizingSpecifiedSize.
func (blk *block) GetSizingMechanism() Sizing {
	return SizingSpecifiedSize
}

// Set the rotation angle in degrees.
func (blk *block) SetAngle(angleDeg float64) {
	blk.angle = angleDeg
}

// Duplicate the block with a new copy of the operations list.
func (blk *block) duplicate() *block {
	dup := &block{}

	// Copy over.
	*dup = *blk

	dupContents := contentstream.ContentStreamOperations{}
	for _, op := range *blk.contents {
		dupContents = append(dupContents, op)
	}
	dup.contents = &dupContents

	return dup
}

// Draws the block contents on a template page block.
func (blk *block) GeneratePageBlocks(ctx drawContext) ([]*block, drawContext, error) {
	blocks := []*block{}

	if blk.positioning.isRelative() {
		// Draw at current ctx.X, ctx.Y position
		dup := blk.duplicate()
		cc := contentstream.NewContentCreator()
		cc.Translate(ctx.X, ctx.pageHeight-ctx.Y-blk.height)
		if blk.angle != 0 {
			// Make the rotation about the upper left corner.
			// XXX/TODO: Account for rotation origin. (Consider).
			cc.Translate(0, blk.Height())
			cc.RotateDeg(blk.angle)
			cc.Translate(0, -blk.Height())
		}
		contents := append(*cc.Operations(), *dup.contents...)
		dup.contents = &contents

		blocks = append(blocks, dup)

		ctx.Y += blk.height
	} else {
		// Absolute. Draw at blk.xPos, blk.yPos position
		dup := blk.duplicate()
		cc := contentstream.NewContentCreator()
		cc.Translate(blk.xPos, ctx.pageHeight-blk.yPos-blk.height)
		if blk.angle != 0 {
			// Make the rotation about the upper left corner.
			// XXX/TODO: Consider supporting specification of rotation origin.
			cc.Translate(0, blk.Height())
			cc.RotateDeg(blk.angle)
			cc.Translate(0, -blk.Height())
		}
		contents := append(*cc.Operations(), *dup.contents...)
		contents.WrapIfNeeded()
		dup.contents = &contents

		blocks = append(blocks, dup)
	}

	return blocks, ctx, nil
}

// Get block height.
func (blk *block) Height() float64 {
	return blk.height
}

// Get block width.
func (blk *block) Width() float64 {
	return blk.width
}

// Add contents to a block.  Wrap both existing and new contents to ensure
// independence of content operations.
func (blk *block) addContents(operations *contentstream.ContentStreamOperations) {
	blk.contents.WrapIfNeeded()
	operations.WrapIfNeeded()
	*blk.contents = append(*blk.contents, *operations...)
}

// Set block margins.
func (blk *block) SetMargins(left, right, top, bottom float64) {
	blk.margins.left = left
	blk.margins.right = right
	blk.margins.top = top
	blk.margins.bottom = bottom
}

// Return block margins: left, right, top, bottom margins.
func (blk *block) GetMargins() (float64, float64, float64, float64) {
	return blk.margins.left, blk.margins.right, blk.margins.top, blk.margins.bottom
}

// Set block positioning to absolute and set the absolute position coordinates as specified.
func (blk *block) SetPos(x, y float64) {
	blk.positioning = positionAbsolute
	blk.xPos = x
	blk.yPos = y
}

// Scale block by specified factors in the x and y directions.
func (blk *block) Scale(sx, sy float64) {
	ops := contentstream.NewContentCreator().
		Scale(sx, sy).
		Operations()

	*blk.contents = append(*ops, *blk.contents...)
	blk.contents.WrapIfNeeded()

	blk.width *= sx
	blk.height *= sy
}

// Scale to a specified width, maintaining aspect ratio.
func (blk *block) ScaleToWidth(w float64) {
	ratio := w / blk.width
	blk.Scale(ratio, ratio)
}

// Scale to a specified height, maintaining aspect ratio.
func (blk *block) ScaleToHeight(h float64) {
	ratio := h / blk.height
	blk.Scale(ratio, ratio)
}

// Internal function to apply translation to the block, moving block contents on the PDF.
func (blk *block) translate(tx, ty float64) {
	ops := contentstream.NewContentCreator().
		Translate(tx, -ty).
		Operations()

	*blk.contents = append(*ops, *blk.contents...)
	blk.contents.WrapIfNeeded()
}

// Draw the block on a page.
func (blk *block) drawToPage(page *model.PdfPage) error {
	// Check if page contents are wrapped - if not wrap it.
	content, err := page.GetAllContentStreams()
	if err != nil {
		return err
	}

	contentParser := contentstream.NewContentStreamParser(content)
	ops, err := contentParser.Parse()
	if err != nil {
		return err
	}
	ops.WrapIfNeeded()

	// Ensure resource dictionaries are available.
	if page.Resources == nil {
		page.Resources = model.NewPdfPageResources()
	}

	// Merge the contents into ops.
	err = mergeContents(ops, page.Resources, blk.contents, blk.resources)
	if err != nil {
		return err
	}

	err = page.SetContentStreams([]string{string(ops.Bytes())}, core.NewFlateEncoder())
	if err != nil {
		return err
	}

	return nil
}

// Draw the drawable d on the block.
// Note that the drawable must not wrap, i.e. only return one block. Otherwise an error is returned.
func (blk *block) Draw(d Drawable) error {
	ctx := drawContext{}
	ctx.Width = blk.width
	ctx.Height = blk.height
	ctx.pageWidth = blk.width
	ctx.pageHeight = blk.height
	ctx.X = 0 // Upper left corner of block
	ctx.Y = 0

	blocks, _, err := d.GeneratePageBlocks(ctx)
	if err != nil {
		return err
	}

	if len(blocks) != 1 {
		return errors.New("Too many output blocks")
	}

	for _, newBlock := range blocks {
		err := mergeContents(blk.contents, blk.resources, newBlock.contents, newBlock.resources)
		if err != nil {
			return err
		}
	}

	return nil
}

// Append another block onto the block.
func (blk *block) mergeBlocks(toAdd *block) error {
	err := mergeContents(blk.contents, blk.resources, toAdd.contents, toAdd.resources)
	return err
}

// Merge contents and content streams.
// Active in the sense that it modified the input contents and resources.
func mergeContents(contents *contentstream.ContentStreamOperations, resources *model.PdfPageResources,
	contentsToAdd *contentstream.ContentStreamOperations, resourcesToAdd *model.PdfPageResources) error {

	// To properly add contents from a block, we need to handle the resources that the block is
	// using and make sure it is accessible in the modified page.
	//
	// Currently only supporting: Font, XObject, Colormap resources
	// from the block.
	//

	xobjectMap := map[core.PdfObjectName]core.PdfObjectName{}
	fontMap := map[core.PdfObjectName]core.PdfObjectName{}
	csMap := map[core.PdfObjectName]core.PdfObjectName{}

	for _, op := range *contentsToAdd {
		switch op.Operand {
		case "Do":
			// XObject.
			if len(op.Params) == 1 {
				if name, ok := op.Params[0].(*core.PdfObjectName); ok {
					if _, processed := xobjectMap[*name]; !processed {
						var useName core.PdfObjectName
						// Process if not already processed..
						obj, _ := resourcesToAdd.GetXObjectByName(*name)
						if obj != nil {
							useName = *name
							for {
								obj2, _ := resources.GetXObjectByName(useName)
								if obj2 == nil || obj2 == obj {
									break
								}
								// If there is a conflict... then append "0" to the name..
								useName = useName + "0"
							}
						}

						resources.SetXObjectByName(useName, obj)
						xobjectMap[*name] = useName
					}
					useName := xobjectMap[*name]
					op.Params[0] = &useName
				}
			}
		case "Tf":
			// Font.
			if len(op.Params) == 2 {
				if name, ok := op.Params[0].(*core.PdfObjectName); ok {
					if _, processed := fontMap[*name]; !processed {
						var useName core.PdfObjectName
						// Process if not already processed.
						obj, found := resourcesToAdd.GetFontByName(*name)
						if found {
							useName = *name
							for {
								obj2, found := resources.GetFontByName(useName)
								if !found || obj2 == obj {
									break
								}
								useName = useName + "0"
							}
						}

						resources.SetFontByName(useName, obj)
						fontMap[*name] = useName
					}

					useName := fontMap[*name]
					op.Params[0] = &useName
				}
			}
		case "CS", "cs":
			// Colorspace.
			if len(op.Params) == 1 {
				if name, ok := op.Params[0].(*core.PdfObjectName); ok {
					if _, processed := csMap[*name]; !processed {
						var useName core.PdfObjectName
						// Process if not already processed.
						cs, found := resourcesToAdd.GetColorspaceByName(*name)
						if found {
							useName = *name
							for {
								cs2, found := resources.GetColorspaceByName(useName)
								if !found || cs == cs2 {
									break
								}
								useName = useName + "0"
							}
						}

						resources.SetColorspaceByName(useName, cs)
						csMap[*name] = useName
					}

					useName := csMap[*name]
					op.Params[0] = &useName
				}
			}
		}

		*contents = append(*contents, op)
	}

	return nil
}
