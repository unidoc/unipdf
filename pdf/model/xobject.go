/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	. "github.com/unidoc/unidoc/pdf/core"
)

// XObjectImage (Table 89 in 8.9.5.1).
// Implements PdfModel interface.
type XObjectImage struct {
	//ColorSpace       PdfObject
	Width            *int64
	Height           *int64
	ColorSpace       PdfColorspace
	BitsPerComponent *int64
	Filter           *PdfObjectName
	Intent           PdfObject
	ImageMask        PdfObject
	Mask             PdfObject
	Decode           PdfObject
	Interpolate      PdfObject
	Alternatives     PdfObject
	SMask            PdfObject
	SMaskInData      PdfObject
	Name             PdfObject
	StructParent     PdfObject
	ID               PdfObject
	OPI              PdfObject
	Metadata         PdfObject
	OC               PdfObject
	Stream           []byte
	// Primitive
	primitive *PdfObjectStream
}

func NewXObjectImage() *XObjectImage {
	xobj := &XObjectImage{}
	stream := &PdfObjectStream{}
	stream.PdfObjectDictionary = &PdfObjectDictionary{}
	xobj.primitive = stream
	return xobj
}

// Creates a new XObject Image from an image object with default
// options.
func NewXObjectImageFromImage(name PdfObjectName, img *Image) (*XObjectImage, error) {
	xobj := NewXObjectImage()

	xobj.Name = &name
	xobj.Stream = img.Data

	// Width and height.
	imWidth := img.Width
	imHeight := img.Height
	xobj.Width = &imWidth
	xobj.Height = &imHeight

	// Bits.
	xobj.BitsPerComponent = &img.BitsPerComponent

	//xobj.ColorSpace = MakeName("DeviceRGB")
	xobj.ColorSpace = NewPdfColorspaceDeviceRGB()

	// define color space?

	return xobj, nil
}

// Build the image xobject from a stream object.
// An image dictionary is the dictionary portion of a stream object representing an image XObject.
func NewXObjectImageFromStream(stream *PdfObjectStream) (*XObjectImage, error) {
	img := &XObjectImage{}
	img.primitive = stream

	dict := *(stream.PdfObjectDictionary)

	if obj, isDefined := dict["Filter"]; isDefined {
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, errors.New("Invalid image Filter (not name)")
		}
		img.Filter = name
	}

	if obj, isDefined := dict["Width"]; isDefined {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image width object")
		}
		iVal := int64(*iObj)
		img.Width = &iVal
	} else {
		return nil, errors.New("Width missing")
	}

	if obj, isDefined := dict["Height"]; isDefined {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image height object")
		}
		iVal := int64(*iObj)
		img.Height = &iVal
	} else {
		return nil, errors.New("Height missing")
	}

	if obj, isDefined := dict["ColorSpace"]; isDefined {
		//img.ColorSpace = obj
		cs, err := newPdfColorspaceFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		img.ColorSpace = cs
	}

	if obj, isDefined := dict["BitsPerComponent"]; isDefined {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image height object")
		}
		iVal := int64(*iObj)
		img.BitsPerComponent = &iVal
	}

	if obj, isDefined := dict["Intent"]; isDefined {
		img.Intent = obj
	}
	if obj, isDefined := dict["ImageMask"]; isDefined {
		img.ImageMask = obj
	}
	if obj, isDefined := dict["Mask"]; isDefined {
		img.Mask = obj
	}
	if obj, isDefined := dict["Decode"]; isDefined {
		img.Decode = obj
	}
	if obj, isDefined := dict["Interpolate"]; isDefined {
		img.Interpolate = obj
	}
	if obj, isDefined := dict["Alternatives"]; isDefined {
		img.Alternatives = obj
	}
	if obj, isDefined := dict["SMask"]; isDefined {
		img.SMask = obj
	}
	if obj, isDefined := dict["SMaskInData"]; isDefined {
		img.SMaskInData = obj
	}
	if obj, isDefined := dict["Name"]; isDefined {
		img.Name = obj
	}
	if obj, isDefined := dict["StructParent"]; isDefined {
		img.StructParent = obj
	}
	if obj, isDefined := dict["ID"]; isDefined {
		img.ID = obj
	}
	if obj, isDefined := dict["OPI"]; isDefined {
		img.OPI = obj
	}
	if obj, isDefined := dict["Metadata"]; isDefined {
		img.Metadata = obj
	}
	if obj, isDefined := dict["OC"]; isDefined {
		img.OC = obj
	}

	img.Stream = stream.Stream

	return img, nil
}

// This will convert to an Image which can be transformed or saved out.
// The image is returned in RGB colormap.
func (ximg *XObjectImage) ToImage() (*Image, error) {
	image := &Image{}

	if ximg.Height == nil {
		return nil, errors.New("Height attribute missing")
	}
	image.Height = *ximg.Height

	if ximg.Width == nil {
		return nil, errors.New("Width attribute missing")
	}
	image.Width = *ximg.Width

	if ximg.BitsPerComponent == nil {
		return nil, errors.New("Bits per component missing")
	}
	image.BitsPerComponent = *ximg.BitsPerComponent

	decoded, err := DecodeStream(ximg.primitive)
	if err != nil {
		return nil, err
	}
	image.Data = decoded

	// Convert to RGB image representation.
	rgbImage, err := ximg.ColorSpace.ToRGB(*image)
	return &rgbImage, err
}

func (ximg *XObjectImage) GetContainingPdfObject() PdfObject {
	return ximg.primitive
}

// Return a stream object.
func (ximg *XObjectImage) ToPdfObject() PdfObject {
	stream := ximg.primitive
	stream.Stream = ximg.Stream

	dict := stream.PdfObjectDictionary

	dict.Set("Type", MakeName("XObject"))
	dict.Set("Subtype", MakeName("Image"))
	dict.Set("Width", MakeInteger(*(ximg.Width)))
	dict.Set("Height", MakeInteger(*(ximg.Height)))
	dict.Set("Filter", MakeName("DCTDecode"))

	if ximg.BitsPerComponent != nil {
		dict.Set("BitsPerComponent", MakeInteger(*(ximg.BitsPerComponent)))
	}

	if ximg.ColorSpace != nil {
		dict.SetIfNotNil("ColorSpace", ximg.ColorSpace.ToPdfObject())
	}
	dict.SetIfNotNil("Intent", ximg.Intent)
	dict.SetIfNotNil("ImageMask", ximg.ImageMask)
	dict.SetIfNotNil("Mask", ximg.Mask)
	dict.SetIfNotNil("Decode", ximg.Decode)
	dict.SetIfNotNil("Interpolate", ximg.Interpolate)
	dict.SetIfNotNil("Alternatives", ximg.Alternatives)
	dict.SetIfNotNil("SMask", ximg.SMask)
	dict.SetIfNotNil("SMaskInData", ximg.SMaskInData)
	dict.SetIfNotNil("Name", ximg.Name)
	dict.SetIfNotNil("StructParent", ximg.StructParent)
	dict.SetIfNotNil("ID", ximg.ID)
	dict.SetIfNotNil("OPI", ximg.OPI)
	dict.SetIfNotNil("Metadata", ximg.Metadata)
	dict.SetIfNotNil("OC", ximg.OC)

	dict.Set("Length", MakeInteger(int64(len(ximg.Stream))))
	stream.Stream = ximg.Stream

	return stream
}
