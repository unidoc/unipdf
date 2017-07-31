/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"bytes"
	"fmt"
	goimage "image"

	"io/ioutil"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// The Image type is used to draw an image onto PDF.
type Image struct {
	xobj *model.XObjectImage

	// Rotation angle.
	angle float64

	// The dimensions of the image. As to be placed on the PDF.
	width, height float64

	// The original dimensions of the image (pixel based).
	origWidth, origHeight float64

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos float64
	yPos float64

	// Opacity (alpha value).
	opacity float64

	// Margins to be applied around the block when drawing on Page.
	margins margins

	// Rotional origin.  Default (0,0 - upper left corner of block).
	rotOriginX, rotOriginY float64
}

// NewImage create a new image from a unidoc image (model.Image).
func NewImage(img *model.Image) (*Image, error) {
	image := &Image{}

	// Create the XObject image.
	ximg, err := model.NewXObjectImageFromImage(img, nil, core.NewFlateEncoder())
	if err != nil {
		common.Log.Error("Failed to create xobject image: %s", err)
		return nil, err
	}

	image.xobj = ximg

	// Image original size in points = pixel size.
	image.origWidth = float64(img.Width)
	image.origHeight = float64(img.Height)
	image.width = image.origWidth
	image.height = image.origHeight
	image.angle = 0
	image.opacity = 1.0

	image.positioning = positionRelative

	return image, nil
}

// NewImageFromData creates an Image from image data.
func NewImageFromData(data []byte) (*Image, error) {
	imgReader := bytes.NewReader(data)

	// Load the image with default handler.
	img, err := model.ImageHandling.Read(imgReader)
	if err != nil {
		common.Log.Error("Error loading image: %s", err)
		return nil, err
	}

	return NewImage(img)
}

// NewImageFromFile creates an Image from a file.
func NewImageFromFile(path string) (*Image, error) {
	imgData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	img, err := NewImageFromData(imgData)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// NewImageFromGoImage creates an Image from a go image.Image datastructure.
func NewImageFromGoImage(goimg goimage.Image) (*Image, error) {
	img, err := model.ImageHandling.NewImageFromGoImage(goimg)
	if err != nil {
		return nil, err
	}

	return NewImage(img)
}

// Height returns Image's document height.
func (img *Image) Height() float64 {
	return img.height
}

// Width returns Image's document width.
func (img *Image) Width() float64 {
	return img.width
}

// SetOpacity sets opacity for Image.
func (img *Image) SetOpacity(opacity float64) {
	img.opacity = opacity
}

// SetMargins sets the margins for the Image (in relative mode): left, right, top, bottom.
func (img *Image) SetMargins(left, right, top, bottom float64) {
	img.margins.left = left
	img.margins.right = right
	img.margins.top = top
	img.margins.bottom = bottom
}

// GetMargins returns the Image's margins: left, right, top, bottom.
func (img *Image) GetMargins() (float64, float64, float64, float64) {
	return img.margins.left, img.margins.right, img.margins.top, img.margins.bottom
}

// GeneratePageBlocks generate the Page blocks. Draws the Image on a block, implementing the Drawable interface.
func (img *Image) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	blocks := []*Block{}
	origCtx := ctx

	blk := NewBlock(ctx.PageWidth, ctx.PageHeight)
	if img.positioning.isRelative() {
		if img.height > ctx.Height {
			// Goes out of the bounds.  Write on a new template instead and create a new context at upper
			// left corner.

			blocks = append(blocks, blk)
			blk = NewBlock(ctx.PageWidth, ctx.PageHeight)

			// New Page.
			ctx.Page++
			newContext := ctx
			newContext.Y = ctx.Margins.top // + p.Margins.top
			newContext.X = ctx.Margins.left + img.margins.left
			newContext.Height = ctx.PageHeight - ctx.Margins.top - ctx.Margins.bottom - img.margins.bottom
			newContext.Width = ctx.PageWidth - ctx.Margins.left - ctx.Margins.right - img.margins.left - img.margins.right
			ctx = newContext
		} else {
			ctx.Y += img.margins.top
			ctx.Height -= img.margins.top + img.margins.bottom
			ctx.X += img.margins.left
			ctx.Width -= img.margins.left + img.margins.right
		}
	} else {
		// Absolute.
		ctx.X = img.xPos
		ctx.Y = img.yPos
	}

	// Place the Image on the template at position (x,y) based on the ctx.
	ctx, err := drawImageOnBlock(blk, img, ctx)
	if err != nil {
		return nil, ctx, err
	}

	blocks = append(blocks, blk)

	if img.positioning.isAbsolute() {
		// Absolute drawing should not affect context.
		ctx = origCtx
	} else {
		// XXX/TODO: Use projected height.
		ctx.Y += img.margins.bottom
		ctx.Height -= img.margins.bottom
	}

	return blocks, ctx, nil
}

// SetPos sets the absolute position. Changes object positioning to absolute.
func (img *Image) SetPos(x, y float64) {
	img.positioning = positionAbsolute
	img.xPos = x
	img.yPos = y
}

// Scale scales Image by a constant factor, both width and height.
func (img *Image) Scale(xFactor, yFactor float64) {
	img.width = xFactor * img.width
	img.height = yFactor * img.height
}

// ScaleToWidth scale Image to a specified width w, maintaining the aspect ratio.
func (img *Image) ScaleToWidth(w float64) {
	ratio := img.height / img.width
	img.width = w
	img.height = w * ratio
}

// ScaleToHeight scale Image to a specified height h, maintaining the aspect ratio.
func (img *Image) ScaleToHeight(h float64) {
	ratio := img.width / img.height
	img.height = h
	img.width = h * ratio
}

// SetWidth set the Image's document width to specified w. This does not change the raw image data, i.e.
// no actual scaling of data is performed. That is handled by the PDF viewer.
func (img *Image) SetWidth(w float64) {
	img.width = w
}

// SetHeight sets the Image's document height to specified h.
func (img *Image) SetHeight(h float64) {
	img.height = h
}

// SetAngle sets Image rotation angle in degrees.
func (img *Image) SetAngle(angle float64) {
	img.angle = angle
}

// Draw the image onto the specified blk.
func drawImageOnBlock(blk *Block, img *Image, ctx DrawContext) (DrawContext, error) {
	origCtx := ctx

	// Find a free name for the image.
	num := 1
	imgName := core.PdfObjectName(fmt.Sprintf("Img%d", num))
	for blk.resources.HasXObjectByName(imgName) {
		num++
		imgName = core.PdfObjectName(fmt.Sprintf("Img%d", num))
	}

	// Add to the Page resources.
	err := blk.resources.SetXObjectImageByName(imgName, img.xobj)
	if err != nil {
		return ctx, err
	}

	// Find an available GS name.
	i := 0
	gsName := core.PdfObjectName(fmt.Sprintf("GS%d", i))
	for blk.resources.HasExtGState(gsName) {
		i++
		gsName = core.PdfObjectName(fmt.Sprintf("GS%d", i))
	}

	// Graphics state with normal blend mode.
	gs0 := core.MakeDict()
	gs0.Set("BM", core.MakeName("Normal"))
	if img.opacity < 1.0 {
		gs0.Set("CA", core.MakeFloat(img.opacity))
		gs0.Set("ca", core.MakeFloat(img.opacity))
	}

	err = blk.resources.AddExtGState(gsName, core.MakeIndirectObject(gs0))
	if err != nil {
		return ctx, err
	}

	xPos := ctx.X
	yPos := ctx.PageHeight - ctx.Y - img.Height()
	angle := img.angle

	// Create content stream to add to the Page contents.
	contentCreator := contentstream.NewContentCreator()

	contentCreator.Add_gs(gsName) // Set graphics state.

	contentCreator.Translate(xPos, yPos)
	if angle != 0 {
		// Make the rotation about the upper left corner.
		contentCreator.Translate(0, img.Height())
		contentCreator.RotateDeg(angle)
		contentCreator.Translate(0, -img.Height())
	}

	contentCreator.
		Scale(img.Width(), img.Height()).
		Add_Do(imgName) // Draw the image.

	ops := contentCreator.Operations()
	ops.WrapIfNeeded()

	blk.addContents(ops)

	if img.positioning.isRelative() {
		ctx.Y += img.Height()
		ctx.Height -= img.Height()
		return ctx, nil
	} else {
		// Absolute positioning - return original context.
		return origCtx, nil
	}
}
