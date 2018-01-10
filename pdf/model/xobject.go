/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

// XObjectForm (Table 95 in 8.10.2).
type XObjectForm struct {
	Filter StreamEncoder

	FormType      PdfObject
	BBox          PdfObject
	Matrix        PdfObject
	Resources     *PdfPageResources
	Group         PdfObject
	Ref           PdfObject
	MetaData      PdfObject
	PieceInfo     PdfObject
	LastModified  PdfObject
	StructParent  PdfObject
	StructParents PdfObject
	OPI           PdfObject
	OC            PdfObject
	Name          PdfObject

	// Stream data.
	Stream []byte
	// Primitive
	primitive *PdfObjectStream
}

// Create a brand new XObject Form. Creates a new underlying PDF object stream primitive.
func NewXObjectForm() *XObjectForm {
	xobj := &XObjectForm{}
	stream := &PdfObjectStream{}
	stream.PdfObjectDictionary = MakeDict()
	xobj.primitive = stream
	return xobj
}

// Build the Form XObject from a stream object.
// XXX: Should this be exposed? Consider different access points.
func NewXObjectFormFromStream(stream *PdfObjectStream) (*XObjectForm, error) {
	form := &XObjectForm{}
	form.primitive = stream

	dict := *(stream.PdfObjectDictionary)

	encoder, err := NewEncoderFromStream(stream)
	if err != nil {
		return nil, err
	}
	form.Filter = encoder

	if obj := dict.Get("Subtype"); obj != nil {
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, errors.New("Type error")
		}
		if *name != "Form" {
			common.Log.Debug("Invalid form subtype")
			return nil, errors.New("Invalid form subtype")
		}
	}

	if obj := dict.Get("FormType"); obj != nil {
		form.FormType = obj
	}
	if obj := dict.Get("BBox"); obj != nil {
		form.BBox = obj
	}
	if obj := dict.Get("Matrix"); obj != nil {
		form.Matrix = obj
	}
	if obj := dict.Get("Resources"); obj != nil {
		obj = TraceToDirectObject(obj)
		d, ok := obj.(*PdfObjectDictionary)
		if !ok {
			common.Log.Debug("Invalid XObject Form Resources object, pointing to non-dictionary")
			return nil, errors.New("Type check error")
		}
		res, err := NewPdfPageResourcesFromDict(d)
		if err != nil {
			common.Log.Debug("Failed getting form resources")
			return nil, err
		}
		form.Resources = res
		common.Log.Trace("Form resources: %#v", form.Resources)
	}

	form.Group = dict.Get("Group")
	form.Ref = dict.Get("Ref")
	form.MetaData = dict.Get("MetaData")
	form.PieceInfo = dict.Get("PieceInfo")
	form.LastModified = dict.Get("LastModified")
	form.StructParent = dict.Get("StructParent")
	form.StructParents = dict.Get("StructParents")
	form.OPI = dict.Get("OPI")
	form.OC = dict.Get("OC")
	form.Name = dict.Get("Name")

	form.Stream = stream.Stream

	return form, nil
}

func (xform *XObjectForm) GetContainingPdfObject() PdfObject {
	return xform.primitive
}

func (xform *XObjectForm) GetContentStream() ([]byte, error) {
	decoded, err := DecodeStream(xform.primitive)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

// Update the content stream with specified encoding.  If encoding is null, will use the xform.Filter object
// or Raw encoding if not set.
func (xform *XObjectForm) SetContentStream(content []byte, encoder StreamEncoder) error {
	encoded := content

	if encoder == nil {
		if xform.Filter != nil {
			encoder = xform.Filter
		} else {
			encoder = NewRawEncoder()
		}
	}

	enc, err := encoder.EncodeBytes(encoded)
	if err != nil {
		return err
	}
	encoded = enc

	xform.Stream = encoded

	return nil
}

// Return a stream object.
func (xform *XObjectForm) ToPdfObject() PdfObject {
	stream := xform.primitive

	dict := stream.PdfObjectDictionary
	if xform.Filter != nil {
		// Pre-populate the stream dictionary with the encoding related fields.
		dict = xform.Filter.MakeStreamDict()
		stream.PdfObjectDictionary = dict
	}
	dict.Set("Type", MakeName("XObject"))
	dict.Set("Subtype", MakeName("Form"))

	dict.SetIfNotNil("FormType", xform.FormType)
	dict.SetIfNotNil("BBox", xform.BBox)
	dict.SetIfNotNil("Matrix", xform.Matrix)
	if xform.Resources != nil {
		dict.SetIfNotNil("Resources", xform.Resources.ToPdfObject())
	}
	dict.SetIfNotNil("Group", xform.Group)
	dict.SetIfNotNil("Ref", xform.Ref)
	dict.SetIfNotNil("MetaData", xform.MetaData)
	dict.SetIfNotNil("PieceInfo", xform.PieceInfo)
	dict.SetIfNotNil("LastModified", xform.LastModified)
	dict.SetIfNotNil("StructParent", xform.StructParent)
	dict.SetIfNotNil("StructParents", xform.StructParents)
	dict.SetIfNotNil("OPI", xform.OPI)
	dict.SetIfNotNil("OC", xform.OC)
	dict.SetIfNotNil("Name", xform.Name)

	dict.Set("Length", MakeInteger(int64(len(xform.Stream))))
	stream.Stream = xform.Stream

	return stream
}

// XObjectImage (Table 89 in 8.9.5.1).
// Implements PdfModel interface.
type XObjectImage struct {
	//ColorSpace       PdfObject
	Width            *int64
	Height           *int64
	ColorSpace       PdfColorspace
	BitsPerComponent *int64
	Filter           StreamEncoder

	Intent       PdfObject
	ImageMask    PdfObject
	Mask         PdfObject
	Decode       PdfObject
	Interpolate  PdfObject
	Alternatives PdfObject
	SMask        PdfObject
	SMaskInData  PdfObject
	Name         PdfObject // Obsolete. Currently read if available and write if available. Not setting on new created files.
	StructParent PdfObject
	ID           PdfObject
	OPI          PdfObject
	Metadata     PdfObject
	OC           PdfObject
	Stream       []byte
	// Primitive
	primitive *PdfObjectStream
}

func NewXObjectImage() *XObjectImage {
	xobj := &XObjectImage{}
	stream := &PdfObjectStream{}
	stream.PdfObjectDictionary = MakeDict()
	xobj.primitive = stream
	return xobj
}

// Creates a new XObject Image from an image object with default options.
// If encoder is nil, uses raw encoding (none).
func NewXObjectImageFromImage(img *Image, cs PdfColorspace, encoder StreamEncoder) (*XObjectImage, error) {
	xobj := NewXObjectImage()

	if encoder == nil {
		encoder = NewRawEncoder()
	}

	encoded, err := encoder.EncodeBytes(img.Data)
	if err != nil {
		common.Log.Debug("Error with encoding: %v", err)
		return nil, err
	}

	xobj.Filter = encoder
	xobj.Stream = encoded

	// Width and height.
	imWidth := img.Width
	imHeight := img.Height
	xobj.Width = &imWidth
	xobj.Height = &imHeight

	// Bits.
	xobj.BitsPerComponent = &img.BitsPerComponent

	// Guess colorspace if not explicitly set.
	if cs == nil {
		if img.ColorComponents == 1 {
			xobj.ColorSpace = NewPdfColorspaceDeviceGray()
		} else if img.ColorComponents == 3 {
			xobj.ColorSpace = NewPdfColorspaceDeviceRGB()
		} else if img.ColorComponents == 4 {
			xobj.ColorSpace = NewPdfColorspaceDeviceCMYK()
		} else {
			return nil, errors.New("Colorspace undefined")
		}
	} else {
		xobj.ColorSpace = cs
	}

	if img.hasAlpha {
		// Add the alpha channel information as a stencil mask (SMask).
		// Has same width and height as original and stored in same
		// bits per component (1 component, hence the DeviceGray channel).
		smask := NewXObjectImage()
		smask.Filter = encoder
		encoded, err := encoder.EncodeBytes(img.alphaData)
		if err != nil {
			common.Log.Debug("Error with encoding: %v", err)
			return nil, err
		}
		smask.Stream = encoded
		smask.BitsPerComponent = &img.BitsPerComponent
		smask.Width = &img.Width
		smask.Height = &img.Height
		smask.ColorSpace = NewPdfColorspaceDeviceGray()
		xobj.SMask = smask.ToPdfObject()
	}

	return xobj, nil
}

// Build the image xobject from a stream object.
// An image dictionary is the dictionary portion of a stream object representing an image XObject.
func NewXObjectImageFromStream(stream *PdfObjectStream) (*XObjectImage, error) {
	img := &XObjectImage{}
	img.primitive = stream

	dict := *(stream.PdfObjectDictionary)

	encoder, err := NewEncoderFromStream(stream)
	if err != nil {
		return nil, err
	}
	img.Filter = encoder

	if obj := TraceToDirectObject(dict.Get("Width")); obj != nil {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image width object")
		}
		iVal := int64(*iObj)
		img.Width = &iVal
	} else {
		return nil, errors.New("Width missing")
	}

	if obj := TraceToDirectObject(dict.Get("Height")); obj != nil {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image height object")
		}
		iVal := int64(*iObj)
		img.Height = &iVal
	} else {
		return nil, errors.New("Height missing")
	}

	if obj := TraceToDirectObject(dict.Get("ColorSpace")); obj != nil {
		cs, err := NewPdfColorspaceFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		img.ColorSpace = cs
	} else {
		// If not specified, assume gray..
		common.Log.Debug("XObject Image colorspace not specified - assuming 1 color component")
		img.ColorSpace = NewPdfColorspaceDeviceGray()
	}

	if obj := TraceToDirectObject(dict.Get("BitsPerComponent")); obj != nil {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image height object")
		}
		iVal := int64(*iObj)
		img.BitsPerComponent = &iVal
	}

	img.Intent = dict.Get("Intent")
	img.ImageMask = dict.Get("ImageMask")
	img.Mask = dict.Get("Mask")
	img.Decode = dict.Get("Decode")
	img.Interpolate = dict.Get("Interpolate")
	img.Alternatives = dict.Get("Alternatives")
	img.SMask = dict.Get("SMask")
	img.SMaskInData = dict.Get("SMaskInData")
	img.Name = dict.Get("Name")
	img.StructParent = dict.Get("StructParent")
	img.ID = dict.Get("ID")
	img.OPI = dict.Get("OPI")
	img.Metadata = dict.Get("Metadata")
	img.OC = dict.Get("OC")

	img.Stream = stream.Stream

	return img, nil
}

// Update XObject Image with new image data.
func (ximg *XObjectImage) SetImage(img *Image, cs PdfColorspace) error {
	encoded, err := ximg.Filter.EncodeBytes(img.Data)
	if err != nil {
		return err
	}

	ximg.Stream = encoded

	// Width, height and bits.
	ximg.Width = &img.Width
	ximg.Height = &img.Height
	ximg.BitsPerComponent = &img.BitsPerComponent

	// Guess colorspace if not explicitly set.
	if cs == nil {
		if img.ColorComponents == 1 {
			ximg.ColorSpace = NewPdfColorspaceDeviceGray()
		} else if img.ColorComponents == 3 {
			ximg.ColorSpace = NewPdfColorspaceDeviceRGB()
		} else if img.ColorComponents == 4 {
			ximg.ColorSpace = NewPdfColorspaceDeviceCMYK()
		} else {
			return errors.New("Colorspace undefined")
		}
	} else {
		ximg.ColorSpace = cs
	}

	return nil
}

// Set compression filter.  Decodes with current filter sets and encodes the data with the new filter.
func (ximg *XObjectImage) SetFilter(encoder StreamEncoder) error {
	encoded := ximg.Stream
	decoded, err := ximg.Filter.DecodeBytes(encoded)
	if err != nil {
		return err
	}

	ximg.Filter = encoder
	encoded, err = encoder.EncodeBytes(decoded)
	if err != nil {
		return err
	}

	ximg.Stream = encoded
	return nil
}

// This will convert to an Image which can be transformed or saved out.
// The image data is decoded and the Image returned.
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

	image.ColorComponents = ximg.ColorSpace.GetNumComponents()

	decoded, err := DecodeStream(ximg.primitive)
	if err != nil {
		return nil, err
	}
	image.Data = decoded

	if ximg.Decode != nil {
		darr, ok := ximg.Decode.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Invalid Decode object")
			return nil, errors.New("Invalid type")
		}
		decode, err := darr.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		image.decode = decode
	}

	return image, nil
}

func (ximg *XObjectImage) GetContainingPdfObject() PdfObject {
	return ximg.primitive
}

// Return a stream object.
func (ximg *XObjectImage) ToPdfObject() PdfObject {
	stream := ximg.primitive

	dict := stream.PdfObjectDictionary
	if ximg.Filter != nil {
		//dict.Set("Filter", ximg.Filter)
		// Pre-populate the stream dictionary with the
		// encoding related fields.
		dict = ximg.Filter.MakeStreamDict()
		stream.PdfObjectDictionary = dict
	}
	dict.Set("Type", MakeName("XObject"))
	dict.Set("Subtype", MakeName("Image"))
	dict.Set("Width", MakeInteger(*(ximg.Width)))
	dict.Set("Height", MakeInteger(*(ximg.Height)))

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
