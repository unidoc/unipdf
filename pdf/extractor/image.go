/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// ExtractPageImages returns the image contents of the page extractor, including data
// and position, size information for each image.
func (e *Extractor) ExtractPageImages() (*PageImages, error) {
	ctx := &imageExtractContext{}

	err := ctx.extractImagesInContentStream(e.contents, e.resources)
	if err != nil {
		return nil, err
	}

	images := &PageImages{
		Images: ctx.extractedImages,
	}
	return images, nil
}

// PageImages represents extracted images on a PDF page with spatial information:
// display position and size.
type PageImages struct {
	Images []ImageMark
}

// ImageMark represents an image drawn on a page and its position in device coordinates.
// All coordinates are in device coordinates.
type ImageMark struct {
	Image *model.Image

	// Dimensions of the image as displayed in the PDF.
	Width  float64
	Height float64

	// Position of the image in PDF coordinates (lower left corner).
	X float64
	Y float64

	// Angle in degrees, if rotated.
	Angle float64
}

// Provide context for image extraction content stream processing.
type imageExtractContext struct {
	extractedImages []ImageMark
	inlineImages    int
	xObjectImages   int
	xObjectForms    int

	// Cache to avoid processing same image many times.
	cacheXObjectImages map[*core.PdfObjectStream]*cachedImage
}

type cachedImage struct {
	image *model.Image
	cs    model.PdfColorspace
}

func (ctx *imageExtractContext) extractImagesInContentStream(contents string, resources *model.PdfPageResources) error {
	cstreamParser := contentstream.NewContentStreamParser(contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		return err
	}

	if ctx.cacheXObjectImages == nil {
		ctx.cacheXObjectImages = map[*core.PdfObjectStream]*cachedImage{}
	}

	processor := contentstream.NewContentStreamProcessor(*operations)
	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
			return ctx.processOperand(op, gs, resources)
		})

	err = processor.Process(resources)
	if err != nil {
		return err
	}

	return nil
}

// Process individual content stream operands for image extraction.
func (ctx *imageExtractContext) processOperand(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
	if op.Operand == "BI" && len(op.Params) == 1 {
		// BI: Inline image.

		iimg, ok := op.Params[0].(*contentstream.ContentStreamInlineImage)
		if !ok {
			return nil
		}

		img, err := iimg.ToImage(resources)
		if err != nil {
			return err
		}

		cs, err := iimg.GetColorSpace(resources)
		if err != nil {
			return err
		}
		if cs == nil {
			// Default if not specified?
			cs = model.NewPdfColorspaceDeviceGray()
		}

		rgbImg, err := cs.ImageToRGB(*img)
		if err != nil {
			return err
		}
		xDim := gs.CTM.ScalingFactorX()
		yDim := gs.CTM.ScalingFactorY()
		xPos, yPos := gs.CTM.Translation()
		angle := gs.CTM.Angle()

		imgMark := ImageMark{
			Image:  &rgbImg,
			X:      xPos,
			Y:      yPos,
			Width:  xDim,
			Height: yDim,
			Angle:  angle,
		}

		ctx.extractedImages = append(ctx.extractedImages, imgMark)
		ctx.inlineImages++
	} else if op.Operand == "Do" && len(op.Params) == 1 {
		// Do: XObject.
		name, ok := core.GetName(op.Params[0])
		if !ok {
			return errTypeCheck
		}

		_, xtype := resources.GetXObjectByName(*name)
		if xtype == model.XObjectTypeImage {
			common.Log.Debug(" XObject Image: %s", *name)

			stream, _ := resources.GetXObjectByName(*name)
			if stream == nil {
				return nil
			}

			// Cache on stream pointer so can ensure that it is the same object (better than using name).
			cimg, cached := ctx.cacheXObjectImages[stream]
			if !cached {
				ximg, err := resources.GetXObjectImageByName(*name)
				if err != nil {
					return err
				}
				if ximg == nil {
					return nil
				}

				img, err := ximg.ToImage()
				if err != nil {
					return err
				}

				cimg = &cachedImage{
					image: img,
					cs:    ximg.ColorSpace,
				}
				ctx.cacheXObjectImages[stream] = cimg
			}
			img := cimg.image
			cs := cimg.cs

			common.Log.Debug("@Do CTM: %s", gs.CTM.String())
			xDim := gs.CTM.ScalingFactorX()
			yDim := gs.CTM.ScalingFactorY()
			xPos, yPos := gs.CTM.Translation()
			angle := gs.CTM.Angle()

			rgbImg, err := cs.ImageToRGB(*img)
			if err != nil {
				return err
			}
			imgMark := ImageMark{
				Image:  &rgbImg,
				X:      xPos,
				Y:      yPos,
				Width:  xDim,
				Height: yDim,
				Angle:  angle,
			}

			ctx.extractedImages = append(ctx.extractedImages, imgMark)
			ctx.xObjectImages++
		} else if xtype == model.XObjectTypeForm {
			// Go through the XObject Form content stream.
			xform, err := resources.GetXObjectFormByName(*name)
			if err != nil {
				return err
			}
			if xform == nil {
				return nil
			}

			formContent, err := xform.GetContentStream()
			if err != nil {
				return err
			}

			// Process the content stream in the Form object too:
			formResources := xform.Resources
			if formResources == nil {
				formResources = resources
			}

			// Process the content stream in the Form object too:
			err = ctx.extractImagesInContentStream(string(formContent), formResources)
			if err != nil {
				return err
			}
			ctx.xObjectForms++
		}
	}
	return nil
}
