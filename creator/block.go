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
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
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

	// Block annotations.
	annotations []*model.PdfAnnotation
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

// NewBlockFromPage creates a Block from a PDF Page.  Useful for loading template pages as blocks
// from a PDF document and additional content with the creator.
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

	// Inherit page rotation angle.
	if page.Rotate != nil {
		b.angle = -float64(*page.Rotate)
	}

	return b, nil
}

// Angle returns the block rotation angle in degrees.
func (blk *Block) Angle() float64 {
	return blk.angle
}

// SetAngle sets the rotation angle in degrees.
func (blk *Block) SetAngle(angleDeg float64) {
	blk.angle = angleDeg
}

// AddAnnotation adds an annotation to the current block.
// The annotation will be added to the page the block will be rendered on.
func (blk *Block) AddAnnotation(annotation *model.PdfAnnotation) {
	for _, annot := range blk.annotations {
		if annot == annotation {
			return
		}
	}

	blk.annotations = append(blk.annotations, annotation)
}

// duplicate duplicates the block with a new copy of the operations list.
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
	cc := contentstream.NewContentCreator()

	// Position block.
	blkWidth, blkHeight := blk.Width(), blk.Height()
	if blk.positioning.isRelative() {
		// Relative. Draw at current ctx.X, ctx.Y position.
		cc.Translate(ctx.X, ctx.PageHeight-ctx.Y-blkHeight)
	} else {
		// Absolute. Draw at blk.xPos, blk.yPos position.
		cc.Translate(blk.xPos, ctx.PageHeight-blk.yPos-blkHeight)
	}

	// Rotate block.
	rotatedHeight := blkHeight
	if blk.angle != 0 {
		// Make the rotation about the center of the block.
		cc.Translate(blkWidth/2, blkHeight/2)
		cc.RotateDeg(blk.angle)
		cc.Translate(-blkWidth/2, -blkHeight/2)

		_, rotatedHeight = blk.RotatedSize()
	}

	if blk.positioning.isRelative() {
		ctx.Y += rotatedHeight
	}

	dup := blk.duplicate()
	contents := append(*cc.Operations(), *dup.contents...)
	contents.WrapIfNeeded()
	dup.contents = &contents

	return []*Block{dup}, ctx, nil
}

// Height returns the Block's height.
func (blk *Block) Height() float64 {
	return blk.height
}

// Width returns the Block's width.
func (blk *Block) Width() float64 {
	return blk.width
}

// RotatedSize returns the width and height of the rotated block.
func (blk *Block) RotatedSize() (float64, float64) {
	_, _, w, h := rotateRect(blk.width, blk.height, blk.angle)
	return w, h
}

// addContents adds contents to a block.  Wrap both existing and new contents to ensure
// independence of content operations.
func (blk *Block) addContents(operations *contentstream.ContentStreamOperations) {
	blk.contents.WrapIfNeeded()
	operations.WrapIfNeeded()
	*blk.contents = append(*blk.contents, *operations...)
}

// addContentsByString adds contents to a block by contents string.
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

// translate translates the block, moving block contents on the PDF. For internal use.
func (blk *Block) translate(tx, ty float64) {
	ops := contentstream.NewContentCreator().
		Translate(tx, -ty).
		Operations()

	*blk.contents = append(*ops, *blk.contents...)
	blk.contents.WrapIfNeeded()
}

// drawToPage draws the block on a PdfPage. Generates the content streams and appends to the PdfPage's content
// stream and links needed resources.
func (blk *Block) drawToPage(page *model.PdfPage) error {

	// TODO(gunnsth): Appears very wasteful to do this all the time.
	//        Possibly create another wrapper around model.PdfPage (creator.page) which can keep track of whether
	//        this has already been done.

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

	// Merge resources for blocks which were created from pages.
	// Necessary for adding resources which do not appear in the block contents.
	if err = mergeResources(blk.resources, page.Resources); err != nil {
		return err
	}

	err = page.SetContentStreams([]string{string(ops.Bytes())}, core.NewFlateEncoder())
	if err != nil {
		return err
	}

	// Add block annotations to the page.
	for _, annotation := range blk.annotations {
		page.AddAnnotation(annotation)
	}

	return nil
}

// Draw draws the drawable d on the block.
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
		return errors.New("too many output blocks")
	}

	for _, newBlock := range blocks {
		if err := blk.mergeBlocks(newBlock); err != nil {
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
		return errors.New("too many output blocks")
	}

	for _, newBlock := range blocks {
		if err := blk.mergeBlocks(newBlock); err != nil {
			return err
		}
	}

	return nil
}

// mergeBlocks appends another block onto the block.
func (blk *Block) mergeBlocks(toAdd *Block) error {
	err := mergeContents(blk.contents, blk.resources, toAdd.contents, toAdd.resources)
	if err != nil {
		return err
	}

	// Merge annotations.
	for _, annot := range toAdd.annotations {
		blk.AddAnnotation(annot)
	}

	return nil
}

// mergeContents merges contents and content streams.
// Active in the sense that it modified the input contents and resources.
func mergeContents(contents *contentstream.ContentStreamOperations, resources *model.PdfPageResources,
	contentsToAdd *contentstream.ContentStreamOperations, resourcesToAdd *model.PdfPageResources) error {

	// TODO(gunnsth): It seems rather expensive to mergeContents all the time. A lot of repetition.
	//    It would be more efficient to perform the merge at the very and when we have all the "blocks"
	//    for each page.

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
						// Process if not already processed.
						obj, found := resourcesToAdd.GetFontByName(*name)

						useName := *name
						if found && obj != nil {
							useName = resourcesNextUnusedFontName(name.String(), obj, resources)
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

// mergeResources adds all resources from src which are missing from dst.
// For now, the method only merges colorspaces.
func mergeResources(src, dst *model.PdfPageResources) error {
	// Merge colorspaces.
	colorspaces, _ := src.GetColorspaces()
	if colorspaces != nil && len(colorspaces.Colorspaces) > 0 {
		for name, colorspace := range colorspaces.Colorspaces {
			colorspaceName := *core.MakeName(name)
			if dst.HasColorspaceByName(colorspaceName) {
				continue
			}

			err := dst.SetColorspaceByName(colorspaceName, colorspace)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func resourcesNextUnusedFontName(name string, font core.PdfObject, resources *model.PdfPageResources) core.PdfObjectName {
	prefix := strings.TrimRightFunc(strings.TrimSpace(name), func(r rune) bool {
		return unicode.IsNumber(r)
	})
	if prefix == "" {
		prefix = "Font"
	}

	num := 0
	fontName := core.PdfObjectName(name)

	for {
		f, found := resources.GetFontByName(fontName)
		if !found || f == font {
			break
		}

		num++
		fontName = core.PdfObjectName(fmt.Sprintf("%s%d", prefix, num))
	}

	return fontName
}
