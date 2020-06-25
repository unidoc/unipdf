/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// ImageExtractOptions contains options for controlling image extraction from
// PDF pages.
type ImageExtractOptions struct {
	IncludeInlineStencilMasks bool
}

// ExtractPageImages returns the image contents of the page extractor, including data
// and position, size information for each image.
// A set of options to control page image extraction can be passed in. The options
// parameter can be nil for the default options. By default, inline stencil masks
// are not extracted.
func (e *Extractor) ExtractPageImages(options *ImageExtractOptions) (*PageImages, error) {
	ctx := &imageExtractContext{
		options: options,
	}

	err := ctx.extractContentStreamImages(e.contents, e.resources)
	if err != nil {
		return nil, err
	}

	return &PageImages{
		Images: ctx.extractedImages,
	}, nil
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

	// Extract options.
	options *ImageExtractOptions
}

type cachedImage struct {
	image *model.Image
	cs    model.PdfColorspace
}

func (ctx *imageExtractContext) extractContentStreamImages(contents string, resources *model.PdfPageResources) error {
	cstreamParser := contentstream.NewContentStreamParser(contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		return err
	}

	if ctx.cacheXObjectImages == nil {
		ctx.cacheXObjectImages = map[*core.PdfObjectStream]*cachedImage{}
	}
	if ctx.options == nil {
		ctx.options = &ImageExtractOptions{}
	}

	processor := contentstream.NewContentStreamProcessor(*operations)
	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
			return ctx.processOperand(op, gs, resources)
		})

	return processor.Process(resources)
}

// Process individual content stream operands for image extraction.
func (ctx *imageExtractContext) processOperand(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
	if op.Operand == "BI" && len(op.Params) == 1 {
		// BI: Inline image.
		iimg, ok := op.Params[0].(*contentstream.ContentStreamInlineImage)
		if !ok {
			return nil
		}

		if isImageMask, ok := core.GetBoolVal(iimg.ImageMask); ok {
			if isImageMask && !ctx.options.IncludeInlineStencilMasks {
				return nil
			}
		}

		return ctx.extractInlineImage(iimg, gs, resources)
	} else if op.Operand == "Do" && len(op.Params) == 1 {
		// Do: XObject.
		name, ok := core.GetName(op.Params[0])
		if !ok {
			common.Log.Debug("ERROR: Type")
			return errTypeCheck
		}

		_, xtype := resources.GetXObjectByName(*name)
		switch xtype {
		case model.XObjectTypeImage:
			return ctx.extractXObjectImage(name, gs, resources)
		case model.XObjectTypeForm:
			return ctx.extractFormImages(name, gs, resources)
		}
	}
	return nil
}

func (ctx *imageExtractContext) extractInlineImage(iimg *contentstream.ContentStreamInlineImage, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
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

	imgMark := ImageMark{
		Image:  &rgbImg,
		Width:  gs.CTM.ScalingFactorX(),
		Height: gs.CTM.ScalingFactorY(),
		Angle:  gs.CTM.Angle(),
	}
	imgMark.X, imgMark.Y = gs.CTM.Translation()

	ctx.extractedImages = append(ctx.extractedImages, imgMark)
	ctx.inlineImages++
	return nil
}

func (ctx *imageExtractContext) extractXObjectImage(name *core.PdfObjectName, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
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

	rgbImg, err := cs.ImageToRGB(*img)
	if err != nil {
		return err
	}

	common.Log.Debug("@Do CTM: %s", gs.CTM.String())
	imgMark := ImageMark{
		Image:  &rgbImg,
		Width:  gs.CTM.ScalingFactorX(),
		Height: gs.CTM.ScalingFactorY(),
		Angle:  gs.CTM.Angle(),
	}
	imgMark.X, imgMark.Y = gs.CTM.Translation()

	ctx.extractedImages = append(ctx.extractedImages, imgMark)
	ctx.xObjectImages++
	return nil
}

// Go through the XObject Form content stream (recursive processing).
func (ctx *imageExtractContext) extractFormImages(name *core.PdfObjectName, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
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
	err = ctx.extractContentStreamImages(string(formContent), formResources)
	if err != nil {
		return err
	}
	ctx.xObjectForms++
	return nil
}
