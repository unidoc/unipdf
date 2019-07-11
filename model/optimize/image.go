/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// Image optimizes images by rewrite images into JPEG format with quality equals to ImageQuality.
// TODO(a5i): Add support for inline images.
// It implements interface model.Optimizer.
type Image struct {
	ImageQuality int
}

// imageInfo is information about an image.
type imageInfo struct {
	ColorSpace       core.PdfObjectName
	BitsPerComponent int
	ColorComponents  int
	Width            int
	Height           int
	Stream           *core.PdfObjectStream
	PPI              float64
}

// findImages returns images from objects.
func findImages(objects []core.PdfObject) []*imageInfo {
	subTypeKey := core.PdfObjectName("Subtype")
	streamProcessed := make(map[*core.PdfObjectStream]struct{})
	var err error
	var images []*imageInfo
	for _, obj := range objects {
		stream, ok := core.GetStream(obj)
		if !ok {
			continue
		}
		if _, found := streamProcessed[stream]; found {
			continue
		}
		streamProcessed[stream] = struct{}{}
		subTypeValue := stream.PdfObjectDictionary.Get(subTypeKey)
		subType, ok := core.GetName(subTypeValue)
		if !ok || string(*subType) != "Image" {
			continue
		}
		img := &imageInfo{BitsPerComponent: 8, Stream: stream}
		if img.ColorSpace, err = model.DetermineColorspaceNameFromPdfObject(stream.PdfObjectDictionary.Get("ColorSpace")); err != nil {
			common.Log.Error("Error determine color space %s", err)
			continue
		}
		if val, ok := core.GetIntVal(stream.PdfObjectDictionary.Get("BitsPerComponent")); ok {
			img.BitsPerComponent = val
		}
		if val, ok := core.GetIntVal(stream.PdfObjectDictionary.Get("Width")); ok {
			img.Width = val
		}
		if val, ok := core.GetIntVal(stream.PdfObjectDictionary.Get("Height")); ok {
			img.Height = val
		}

		switch img.ColorSpace {
		case "DeviceRGB":
			img.ColorComponents = 3
		case "DeviceGray":
			img.ColorComponents = 1
		default:
			common.Log.Warning("Optimization is not supported for color space %s", img.ColorSpace)
			continue
		}
		images = append(images, img)
	}
	return images
}

// Optimize optimizes PDF objects to decrease PDF size.
func (i *Image) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	if i.ImageQuality <= 0 {
		return objects, nil
	}
	images := findImages(objects)
	if len(images) == 0 {
		return objects, nil
	}

	replaceTable := make(map[core.PdfObject]core.PdfObject)
	imageMasks := make(map[core.PdfObject]struct{})
	for _, img := range images {
		obj := img.Stream.PdfObjectDictionary.Get(core.PdfObjectName("SMask"))
		imageMasks[obj] = struct{}{}
	}

	for index, img := range images {
		stream := img.Stream
		if _, isMask := imageMasks[stream]; isMask {
			continue
		}
		streamEncoder, err := core.NewEncoderFromStream(stream)
		if err != nil {
			common.Log.Warning("Error get encoder for the image stream %s")
			continue
		}
		data, err := streamEncoder.DecodeStream(stream)
		if err != nil {
			common.Log.Warning("Error decode the image stream %s")
			continue
		}
		encoder := core.NewDCTEncoder()
		encoder.ColorComponents = img.ColorComponents
		encoder.Quality = i.ImageQuality
		encoder.BitsPerComponent = img.BitsPerComponent
		encoder.Width = img.Width
		encoder.Height = img.Height
		streamData, err := encoder.EncodeBytes(data)
		if err != nil {
			return nil, err
		}
		originalSize := len(stream.Stream)
		if originalSize < len(streamData) {
			continue
		}
		newStream := &core.PdfObjectStream{Stream: streamData}
		newStream.PdfObjectReference = stream.PdfObjectReference
		newStream.PdfObjectDictionary = core.MakeDict()
		newStream.PdfObjectDictionary.Merge(stream.PdfObjectDictionary)
		fn := core.PdfObjectName(encoder.GetFilterName())
		newStream.PdfObjectDictionary.Set(core.PdfObjectName("Filter"), &fn)
		ln := core.PdfObjectInteger(int64(len(streamData)))
		newStream.PdfObjectDictionary.Set(core.PdfObjectName("Length"), &ln)
		replaceTable[stream] = newStream
		images[index].Stream = newStream
	}
	optimizedObjects = make([]core.PdfObject, len(objects))
	copy(optimizedObjects, objects)
	replaceObjectsInPlace(optimizedObjects, replaceTable)
	return optimizedObjects, nil
}
