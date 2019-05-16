/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"fmt"
	"image"
	"math"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/image/draw"
)

// ImagePPI optimizes images by scaling images such that the PPI (pixels per inch) is never higher than ImageUpperPPI.
// TODO(a5i): Add support for inline images.
// It implements interface model.Optimizer.
type ImagePPI struct {
	ImageUpperPPI float64
}

func scaleImage(stream *core.PdfObjectStream, scale float64) error {
	xImg, err := model.NewXObjectImageFromStream(stream)
	if err != nil {
		return err
	}
	i, err := xImg.ToImage()
	if err != nil {
		return err
	}
	goimg, err := i.ToGoImage()
	if err != nil {
		return err
	}

	newW := int(math.RoundToEven(float64(i.Width) * scale))
	newH := int(math.RoundToEven(float64(i.Height) * scale))
	rect := image.Rect(0, 0, newW, newH)

	var newImage draw.Image
	var imageHandler func(image.Image) (*model.Image, error)

	switch xImg.ColorSpace.String() {
	case "DeviceRGB":
		newImage = image.NewRGBA(rect)
		imageHandler = model.ImageHandling.NewImageFromGoImage
	case "DeviceGray":
		newImage = image.NewGray(rect)
		imageHandler = model.ImageHandling.NewGrayImageFromGoImage
	default:
		return fmt.Errorf("optimization is not supported for color space %s", xImg.ColorSpace.String())
	}

	draw.CatmullRom.Scale(newImage, newImage.Bounds(), goimg, goimg.Bounds(), draw.Over, &draw.Options{})
	if i, err = imageHandler(newImage); err != nil {
		return err
	}

	// Update image encoder
	encoderParams := core.MakeDict()
	encoderParams.Set("ColorComponents", core.MakeInteger(int64(i.ColorComponents)))
	encoderParams.Set("BitsPerComponent", core.MakeInteger(i.BitsPerComponent))
	encoderParams.Set("Width", core.MakeInteger(i.Width))
	encoderParams.Set("Height", core.MakeInteger(i.Height))
	encoderParams.Set("Quality", core.MakeInteger(100))
	encoderParams.Set("Predictor", core.MakeInteger(1))

	xImg.Filter.UpdateParams(encoderParams)

	// Update image
	if err := xImg.SetImage(i, nil); err != nil {
		return err
	}
	xImg.ToPdfObject()
	return nil
}

// Optimize optimizes PDF objects to decrease PDF size.
func (i *ImagePPI) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	if i.ImageUpperPPI <= 0 {
		return objects, nil
	}
	images := findImages(objects)
	if len(images) == 0 {
		return objects, nil
	}
	imageMasks := make(map[core.PdfObject]struct{})
	for _, img := range images {
		obj := img.Stream.PdfObjectDictionary.Get(core.PdfObjectName("SMask"))
		imageMasks[obj] = struct{}{}
	}
	imageByStream := make(map[*core.PdfObjectStream]*imageInfo)
	for _, img := range images {
		imageByStream[img.Stream] = img
	}
	var catalog *core.PdfObjectDictionary
	for _, obj := range objects {
		if dict, isDict := core.GetDict(obj); catalog == nil && isDict {
			if tp, ok := core.GetName(dict.Get(core.PdfObjectName("Type"))); ok && *tp == "Catalog" {
				catalog = dict
			}
		}
	}
	if catalog == nil {
		return objects, nil
	}
	pages, hasPages := core.GetDict(catalog.Get(core.PdfObjectName("Pages")))
	if !hasPages {
		return objects, nil
	}
	kids, hasKids := core.GetArray(pages.Get(core.PdfObjectName("Kids")))
	if !hasKids {
		return objects, nil
	}
	imageByName := make(map[string]*imageInfo)

	for _, pageObj := range kids.Elements() {
		page, ok := core.GetDict(pageObj)
		if !ok {
			continue
		}
		contents, hasContents := core.GetArray(page.Get("Contents"))
		if !hasContents {
			continue
		}
		resources, hasResources := core.GetDict(page.Get("Resources"))
		if !hasResources {
			continue
		}
		xObject, hasXObject := core.GetDict(resources.Get("XObject"))
		if !hasXObject {
			continue
		}
		xObjectKeys := xObject.Keys()
		for _, key := range xObjectKeys {
			if stream, isStream := core.GetStream(xObject.Get(key)); isStream {
				if img, found := imageByStream[stream]; found {
					imageByName[string(key)] = img
				}
			}
		}
		for _, obj := range contents.Elements() {
			if stream, isStream := core.GetStream(obj); isStream {
				streamEncoder, err := core.NewEncoderFromStream(stream)
				if err != nil {
					return nil, err
				}
				data, err := streamEncoder.DecodeStream(stream)
				if err != nil {
					return nil, err
				}

				p := contentstream.NewContentStreamParser(string(data))
				operations, err := p.Parse()
				if err != nil {
					return nil, err
				}
				scaleX, scaleY := 1.0, 1.0
				for _, operation := range *operations {
					if operation.Operand == "Q" {
						scaleX, scaleY = 1.0, 1.0
					}
					if operation.Operand == "cm" && len(operation.Params) == 6 {
						if sx, ok := core.GetFloatVal(operation.Params[0]); ok {
							scaleX = scaleX * sx
						}
						if sy, ok := core.GetFloatVal(operation.Params[3]); ok {
							scaleY = scaleY * sy
						}
						if sx, ok := core.GetIntVal(operation.Params[0]); ok {
							scaleX = scaleX * float64(sx)
						}
						if sy, ok := core.GetIntVal(operation.Params[3]); ok {
							scaleY = scaleY * float64(sy)
						}
					}
					if operation.Operand == "Do" && len(operation.Params) == 1 {
						name, ok := core.GetName(operation.Params[0])
						if !ok {
							continue
						}
						if img, found := imageByName[string(*name)]; found {
							wInch, hInch := scaleX/72.0, scaleY/72.0
							xPPI, yPPI := float64(img.Width)/wInch, float64(img.Height)/hInch
							if wInch == 0 || hInch == 0 {
								xPPI = 72.0
								yPPI = 72.0
							}
							img.PPI = math.Max(img.PPI, xPPI)
							img.PPI = math.Max(img.PPI, yPPI)
						}
					}
				}
			}
		}
	}

	for _, img := range images {
		if _, isMask := imageMasks[img.Stream]; isMask {
			continue
		}
		if img.PPI <= i.ImageUpperPPI {
			continue
		}
		scale := i.ImageUpperPPI / img.PPI
		if err := scaleImage(img.Stream, scale); err != nil {
			common.Log.Debug("Error scale image keep original image: %s", err)
		} else {
			if mask, hasMask := core.GetStream(img.Stream.PdfObjectDictionary.Get(core.PdfObjectName("SMask"))); hasMask {
				if err := scaleImage(mask, scale); err != nil {
					return nil, err
				}
			}
		}
	}

	return objects, nil
}
