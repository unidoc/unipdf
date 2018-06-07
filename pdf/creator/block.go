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
)

// Block contains a portion of PDF Page contents. It has a width and a position and can
// be placed anywhere on a Page.  It can even contain a whole Page, and is used in the creator
// where each Drawable object can output one or more blocks, each representing content for separate pages
// (typically needed when Page breaks occur).
type Block struct {
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

	// Margins to be applied around the block when drawing on Page.
	margins margins
}

// NewBlock creates a new Block with specified width and height.
func NewBlock(width float64, height float64) *Block {
	b := &Block{}
	b.contents = &contentstream.ContentStreamOperations{}
	b.resources = model.NewPdfPageResources()
	b.width = width
	b.height = height
	return b
}

// NewBlockFromPage creates a Block from a PDF Page.  Useful for loading template pages as blocks from a PDF document
// and additional content with the creator.
func NewBlockFromPage(page *model.PdfPage) (*Block, error) {
	b := &Block{}

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

// SetAngle sets the rotation angle in degrees.
func (blk *Block) SetAngle(angleDeg float64) {
	blk.angle = angleDeg
}

// Duplicate the block with a new copy of the operations list.
func (blk *Block) duplicate() *Block {
	dup := &Block{}

	// Copy over.
	*dup = *blk

	dupContents := contentstream.ContentStreamOperations{}
	for _, op := range *blk.contents {
		dupContents = append(dupContents, op)
	}
	dup.contents = &dupContents

	return dup
}

// GeneratePageBlocks draws the block contents on a template Page block.
// Implements the Drawable interface.
func (blk *Block) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	blocks := []*Block{}

	if blk.positioning.isRelative() {
		// Draw at current ctx.X, ctx.Y position
		dup := blk.duplicate()
		cc := contentstream.NewContentCreator()
		cc.Translate(ctx.X, ctx.PageHeight-ctx.Y-blk.height)
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
		cc.Translate(blk.xPos, ctx.PageHeight-blk.yPos-blk.height)
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

// Height returns the Block's height.
func (blk *Block) Height() float64 {
	return blk.height
}

// Width returns the Block's width.
func (blk *Block) Width() float64 {
	return blk.width
}

// Add contents to a block.  Wrap both existing and new contents to ensure
// independence of content operations.
func (blk *Block) addContents(operations *contentstream.ContentStreamOperations) {
	blk.contents.WrapIfNeeded()
	operations.WrapIfNeeded()
	*blk.contents = append(*blk.contents, *operations...)
}

// Add contents to a block by contents string.
func (blk *Block) addContentsByString(contents string) error {
	cc := contentstream.NewContentStreamParser(contents)
	operations, err := cc.Parse()
	if err != nil {
		return err
	}

	blk.contents.WrapIfNeeded()
	operations.WrapIfNeeded()
	*blk.contents = append(*blk.contents, *operations...)

	return nil
}

// SetMargins sets the Block's left, right, top, bottom, margins.
func (blk *Block) SetMargins(left, right, top, bottom float64) {
	blk.margins.left = left
	blk.margins.right = right
	blk.margins.top = top
	blk.margins.bottom = bottom
}

// GetMargins returns the Block's margins: left, right, top, bottom.
func (blk *Block) GetMargins() (float64, float64, float64, float64) {
	return blk.margins.left, blk.margins.right, blk.margins.top, blk.margins.bottom
}

// SetPos sets the Block's positioning to absolute mode with the specified coordinates.
func (blk *Block) SetPos(x, y float64) {
	blk.positioning = positionAbsolute
	blk.xPos = x
	blk.yPos = y
}

// Scale block by specified factors in the x and y directions.
func (blk *Block) Scale(sx, sy float64) {
	ops := contentstream.NewContentCreator().
		Scale(sx, sy).
		Operations()

	*blk.contents = append(*ops, *blk.contents...)
	blk.contents.WrapIfNeeded()

	blk.width *= sx
	blk.height *= sy
}

// ScaleToWidth scales the Block to a specified width, maintaining the same aspect ratio.
func (blk *Block) ScaleToWidth(w float64) {
	ratio := w / blk.width
	blk.Scale(ratio, ratio)
}

// ScaleToHeight scales the Block to a specified height, maintaining the same aspect ratio.
func (blk *Block) ScaleToHeight(h float64) {
	ratio := h / blk.height
	blk.Scale(ratio, ratio)
}

// Internal function to apply translation to the block, moving block contents on the PDF.
func (blk *Block) translate(tx, ty float64) {
	ops := contentstream.NewContentCreator().
		Translate(tx, -ty).
		Operations()

	*blk.contents = append(*ops, *blk.contents...)
	blk.contents.WrapIfNeeded()
}

// Draw the block on a Page.
func (blk *Block) drawToPage(page *model.PdfPage) error {
	// Check if Page contents are wrapped - if not wrap it.
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
func (blk *Block) Draw(d Drawable) error {
	ctx := DrawContext{}
	ctx.Width = blk.width
	ctx.Height = blk.height
	ctx.PageWidth = blk.width
	ctx.PageHeight = blk.height
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

// DrawWithContext draws the Block using the specified drawing context.
func (blk *Block) DrawWithContext(d Drawable, ctx DrawContext) error {
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
func (blk *Block) mergeBlocks(toAdd *Block) error {
	err := mergeContents(blk.contents, blk.resources, toAdd.contents, toAdd.resources)
	return err
}

// Merge contents and content streams.
// Active in the sense that it modified the input contents and resources.
func mergeContents(contents *contentstream.ContentStreamOperations, resources *model.PdfPageResources,
	contentsToAdd *contentstream.ContentStreamOperations, resourcesToAdd *model.PdfPageResources) error {

	// To properly add contents from a block, we need to handle the resources that the block is
	// using and make sure it is accessible in the modified Page.
	//
	// Currently supporting: Font, XObject, Colormap, Pattern, Shading, GState resources
	// from the block.
	//

	xobjectMap := map[core.PdfObjectName]core.PdfObjectName{}
	fontMap := map[core.PdfObjectName]core.PdfObjectName{}
	csMap := map[core.PdfObjectName]core.PdfObjectName{}
	patternMap := map[core.PdfObjectName]core.PdfObjectName{}
	shadingMap := map[core.PdfObjectName]core.PdfObjectName{}
	gstateMap := map[core.PdfObjectName]core.PdfObjectName{}

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

							resources.SetColorspaceByName(useName, cs)
							csMap[*name] = useName
						} else {
							common.Log.Debug("Colorspace not found")
						}
					}

					if useName, has := csMap[*name]; has {
						op.Params[0] = &useName
					} else {
						common.Log.Debug("Error: Colorspace %s not found", *name)
					}
				}
			}
		case "SCN", "scn":
			if len(op.Params) == 1 {
				if name, ok := op.Params[0].(*core.PdfObjectName); ok {
					if _, processed := patternMap[*name]; !processed {
						var useName core.PdfObjectName
						p, found := resourcesToAdd.GetPatternByName(*name)
						if found {
							useName = *name
							for {
								p2, found := resources.GetPatternByName(useName)
								if !found || p2 == p {
									break
								}
								useName = useName + "0"
							}

							err := resources.SetPatternByName(useName, p.ToPdfObject())
							if err != nil {
								return err
							}

							patternMap[*name] = useName
						}
					}

					if useName, has := patternMap[*name]; has {
						op.Params[0] = &useName
					}
				}
			}
		case "sh":
			// Shading.
			if len(op.Params) == 1 {
				if name, ok := op.Params[0].(*core.PdfObjectName); ok {
					if _, processed := shadingMap[*name]; !processed {
						var useName core.PdfObjectName
						// Process if not already processed.
						sh, found := resourcesToAdd.GetShadingByName(*name)
						if found {
							useName = *name
							for {
								sh2, found := resources.GetShadingByName(useName)
								if !found || sh == sh2 {
									break
								}
								useName = useName + "0"
							}

							err := resources.SetShadingByName(useName, sh.ToPdfObject())
							if err != nil {
								common.Log.Debug("ERROR Set shading: %v", err)
								return err
							}

							shadingMap[*name] = useName
						} else {
							common.Log.Debug("Shading not found")
						}
					}

					if useName, has := shadingMap[*name]; has {
						op.Params[0] = &useName
					} else {
						common.Log.Debug("Error: Shading %s not found", *name)
					}
				}
			}
		case "gs":
			// ExtGState.
			if len(op.Params) == 1 {
				if name, ok := op.Params[0].(*core.PdfObjectName); ok {
					if _, processed := gstateMap[*name]; !processed {
						var useName core.PdfObjectName
						// Process if not already processed.
						gs, found := resourcesToAdd.GetExtGState(*name)
						if found {
							useName = *name
							i := 1
							for {
								gs2, found := resources.GetExtGState(useName)
								if !found || gs == gs2 {
									break
								}
								useName = core.PdfObjectName(fmt.Sprintf("GS%d", i))
								i++
							}
						}

						resources.AddExtGState(useName, gs)
						gstateMap[*name] = useName
					}

					useName := gstateMap[*name]
					op.Params[0] = &useName
				}
			}
		}

		*contents = append(*contents, op)
	}

	return nil
}
