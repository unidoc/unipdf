/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// XObjectForm (Table 95 in 8.10.2).
type XObjectForm struct {
	Filter core.StreamEncoder

	FormType      core.PdfObject
	BBox          core.PdfObject
	Matrix        core.PdfObject
	Resources     *PdfPageResources
	Group         core.PdfObject
	Ref           core.PdfObject
	MetaData      core.PdfObject
	PieceInfo     core.PdfObject
	LastModified  core.PdfObject
	StructParent  core.PdfObject
	StructParents core.PdfObject
	OPI           core.PdfObject
	OC            core.PdfObject
	Name          core.PdfObject

	// Stream data.
	Stream []byte
	// Primitive
	primitive *core.PdfObjectStream
}

// NewXObjectForm creates a brand new XObject Form. Creates a new underlying PDF object stream primitive.
func NewXObjectForm() *XObjectForm {
	xobj := &XObjectForm{}
	stream := &core.PdfObjectStream{}
	stream.PdfObjectDictionary = core.MakeDict()
	xobj.primitive = stream
	return xobj
}

// NewXObjectFormFromStream builds the Form XObject from a stream object.
// TODO: Should this be exposed? Consider different access points.
func NewXObjectFormFromStream(stream *core.PdfObjectStream) (*XObjectForm, error) {
	form := &XObjectForm{}
	form.primitive = stream

	dict := *(stream.PdfObjectDictionary)

	encoder, err := core.NewEncoderFromStream(stream)
	if err != nil {
		return nil, err
	}
	form.Filter = encoder

	if obj := dict.Get("Subtype"); obj != nil {
		name, ok := obj.(*core.PdfObjectName)
		if !ok {
			return nil, errors.New("type error")
		}
		if *name != "Form" {
			common.Log.Debug("Invalid form subtype")
			return nil, errors.New("invalid form subtype")
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
		obj = core.TraceToDirectObject(obj)
		d, ok := obj.(*core.PdfObjectDictionary)
		if !ok {
			common.Log.Debug("Invalid XObject Form Resources object, pointing to non-dictionary")
			return nil, core.ErrTypeError
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

// GetContainingPdfObject returns the XObject Form's containing object (indirect object).
func (xform *XObjectForm) GetContainingPdfObject() core.PdfObject {
	return xform.primitive
}

// GetContentStream returns the XObject Form's content stream.
func (xform *XObjectForm) GetContentStream() ([]byte, error) {
	decoded, err := core.DecodeStream(xform.primitive)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

// SetContentStream updates the content stream with specified encoding.
// If encoding is null, will use the xform.Filter object or Raw encoding if not set.
func (xform *XObjectForm) SetContentStream(content []byte, encoder core.StreamEncoder) error {
	encoded := content

	if encoder == nil {
		if xform.Filter != nil {
			encoder = xform.Filter
		} else {
			encoder = core.NewRawEncoder()
		}
	}

	enc, err := encoder.EncodeBytes(encoded)
	if err != nil {
		return err
	}
	encoded = enc

	xform.Stream = encoded
	xform.Filter = encoder

	return nil
}

// ToPdfObject returns a stream object.
func (xform *XObjectForm) ToPdfObject() core.PdfObject {
	stream := xform.primitive

	dict := stream.PdfObjectDictionary
	if xform.Filter != nil {
		// Pre-populate the stream dictionary with the encoding related fields.
		dict = xform.Filter.MakeStreamDict()
		stream.PdfObjectDictionary = dict
	}
	dict.Set("Type", core.MakeName("XObject"))
	dict.Set("Subtype", core.MakeName("Form"))

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

	dict.Set("Length", core.MakeInteger(int64(len(xform.Stream))))
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
	Filter           core.StreamEncoder

	Intent       core.PdfObject
	ImageMask    core.PdfObject
	Mask         core.PdfObject
	Matte        core.PdfObject
	Decode       core.PdfObject
	Interpolate  core.PdfObject
	Alternatives core.PdfObject
	SMask        core.PdfObject
	SMaskInData  core.PdfObject
	Name         core.PdfObject // Obsolete. Currently read if available and write if available. Not setting on new created files.
	StructParent core.PdfObject
	ID           core.PdfObject
	OPI          core.PdfObject
	Metadata     core.PdfObject
	OC           core.PdfObject
	Stream       []byte
	// Primitive
	primitive *core.PdfObjectStream
}

// NewXObjectImage returns a new XObjectImage.
func NewXObjectImage() *XObjectImage {
	xobj := &XObjectImage{}
	stream := &core.PdfObjectStream{}
	stream.PdfObjectDictionary = core.MakeDict()
	xobj.primitive = stream
	return xobj
}

// NewXObjectImageFromImage creates a new XObject Image from an image object
// with default options. If encoder is nil, uses raw encoding (none).
func NewXObjectImageFromImage(img *Image, cs PdfColorspace, encoder core.StreamEncoder) (*XObjectImage, error) {
	xobj := NewXObjectImage()
	return UpdateXObjectImageFromImage(xobj, img, cs, encoder)
}

// UpdateXObjectImageFromImage creates a new XObject Image from an
// Image object `img` and default masks from xobjIn.
// The default masks are overriden if img.hasAlpha
// If `encoder` is nil, uses raw encoding (none).
func UpdateXObjectImageFromImage(xobjIn *XObjectImage, img *Image, cs PdfColorspace,
	encoder core.StreamEncoder) (*XObjectImage, error) {
	if encoder == nil {
		encoder = core.NewRawEncoder()
	}
	encoder.UpdateParams(img.GetParamsDict())

	encoded, err := encoder.EncodeBytes(img.Data)
	if err != nil {
		common.Log.Debug("Error with encoding: %v", err)
		return nil, err
	}
	xobj := NewXObjectImage()

	// Width and height.
	imWidth := img.Width
	imHeight := img.Height
	xobj.Width = &imWidth
	xobj.Height = &imHeight

	// Bits per Component.
	imBPC := img.BitsPerComponent
	xobj.BitsPerComponent = &imBPC

	xobj.Filter = encoder
	xobj.Stream = encoded

	// Guess colorspace if not explicitly set.
	if cs == nil {
		if img.ColorComponents == 1 {
			xobj.ColorSpace = NewPdfColorspaceDeviceGray()
		} else if img.ColorComponents == 3 {
			xobj.ColorSpace = NewPdfColorspaceDeviceRGB()
		} else if img.ColorComponents == 4 {
			xobj.ColorSpace = NewPdfColorspaceDeviceCMYK()
		} else {
			return nil, errors.New("colorspace undefined")
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
		smask.BitsPerComponent = xobj.BitsPerComponent
		smask.Width = &img.Width
		smask.Height = &img.Height
		smask.ColorSpace = NewPdfColorspaceDeviceGray()
		xobj.SMask = smask.ToPdfObject()
	} else {
		xobj.SMask = xobjIn.SMask
		xobj.ImageMask = xobjIn.ImageMask
		if xobj.ColorSpace.GetNumComponents() == 1 {
			smaskMatteToGray(xobj)
		}
	}

	return xobj, nil
}

// smaskMatteToGray converts to gray the Matte value in the SMask image referenced by `xobj` (if
// there is one)
func smaskMatteToGray(xobj *XObjectImage) error {
	if xobj.SMask == nil {
		return nil
	}
	stream, ok := xobj.SMask.(*core.PdfObjectStream)
	if !ok {
		common.Log.Debug("SMask is not *PdfObjectStream")
		return core.ErrTypeError
	}
	dict := stream.PdfObjectDictionary
	matte := dict.Get("Matte")
	if matte == nil {
		return nil
	}

	gray, err := toGray(matte.(*core.PdfObjectArray))
	if err != nil {
		return err
	}
	grayMatte := core.MakeArrayFromFloats([]float64{gray})
	dict.SetIfNotNil("Matte", grayMatte)
	return nil
}

// toGray converts a 1, 3 or 4 dimensional color `matte` to gray
// If `matte` is not a 1, 3 or 4 dimensional color then an error is returned
func toGray(matte *core.PdfObjectArray) (float64, error) {
	colors, err := matte.ToFloat64Array()
	if err != nil {
		common.Log.Debug("Bad Matte array: matte=%s err=%v", matte, err)
	}
	switch len(colors) {
	case 1:
		return colors[0], nil
	case 3:
		cs := PdfColorspaceDeviceRGB{}
		rgbColor, err := cs.ColorFromFloats(colors)
		if err != nil {
			return 0.0, err
		}
		return rgbColor.(*PdfColorDeviceRGB).ToGray().Val(), nil

	case 4:
		cs := PdfColorspaceDeviceCMYK{}
		cmykColor, err := cs.ColorFromFloats(colors)
		if err != nil {
			return 0.0, err
		}
		rgbColor, err := cs.ColorToRGB(cmykColor.(*PdfColorDeviceCMYK))
		if err != nil {
			return 0.0, err
		}
		return rgbColor.(*PdfColorDeviceRGB).ToGray().Val(), nil
	}
	err = errors.New("bad Matte color")
	common.Log.Error("toGray: matte=%s err=%v", matte, err)
	return 0.0, err
}

// NewXObjectImageFromStream builds the image xobject from a stream object.
// An image dictionary is the dictionary portion of a stream object representing an image XObject.
func NewXObjectImageFromStream(stream *core.PdfObjectStream) (*XObjectImage, error) {
	img := &XObjectImage{}
	img.primitive = stream

	dict := *(stream.PdfObjectDictionary)

	encoder, err := core.NewEncoderFromStream(stream)
	if err != nil {
		return nil, err
	}
	img.Filter = encoder

	if obj := core.TraceToDirectObject(dict.Get("Width")); obj != nil {
		iObj, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			return nil, errors.New("invalid image width object")
		}
		iVal := int64(*iObj)
		img.Width = &iVal
	} else {
		return nil, errors.New("width missing")
	}

	if obj := core.TraceToDirectObject(dict.Get("Height")); obj != nil {
		iObj, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			return nil, errors.New("invalid image height object")
		}
		iVal := int64(*iObj)
		img.Height = &iVal
	} else {
		return nil, errors.New("height missing")
	}

	if obj := core.TraceToDirectObject(dict.Get("ColorSpace")); obj != nil {
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

	if obj := core.TraceToDirectObject(dict.Get("BitsPerComponent")); obj != nil {
		iObj, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			return nil, errors.New("invalid image height object")
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
	img.Matte = dict.Get("Matte")
	img.Name = dict.Get("Name")
	img.StructParent = dict.Get("StructParent")
	img.ID = dict.Get("ID")
	img.OPI = dict.Get("OPI")
	img.Metadata = dict.Get("Metadata")
	img.OC = dict.Get("OC")

	img.Stream = stream.Stream

	return img, nil
}

// SetImage updates XObject Image with new image data.
func (ximg *XObjectImage) SetImage(img *Image, cs PdfColorspace) error {
	// update image parameters of the filter encoder.
	ximg.Filter.UpdateParams(img.GetParamsDict())
	encoded, err := ximg.Filter.EncodeBytes(img.Data)
	if err != nil {
		return err
	}

	ximg.Stream = encoded

	// Width, height and bits.
	w := img.Width
	ximg.Width = &w

	h := img.Height
	ximg.Height = &h

	bpc := img.BitsPerComponent
	ximg.BitsPerComponent = &bpc

	// Guess colorspace if not explicitly set.
	if cs == nil {
		if img.ColorComponents == 1 {
			ximg.ColorSpace = NewPdfColorspaceDeviceGray()
		} else if img.ColorComponents == 3 {
			ximg.ColorSpace = NewPdfColorspaceDeviceRGB()
		} else if img.ColorComponents == 4 {
			ximg.ColorSpace = NewPdfColorspaceDeviceCMYK()
		} else {
			return errors.New("colorspace undefined")
		}
	} else {
		ximg.ColorSpace = cs
	}

	return nil
}

// SetFilter sets compression filter. Decodes with current filter sets and
// encodes the data with the new filter.
func (ximg *XObjectImage) SetFilter(encoder core.StreamEncoder) error {
	encoded := ximg.Stream
	decoded, err := ximg.Filter.DecodeBytes(encoded)
	if err != nil {
		return err
	}

	ximg.Filter = encoder
	encoder.UpdateParams(ximg.getParamsDict())
	encoded, err = encoder.EncodeBytes(decoded)
	if err != nil {
		return err
	}

	ximg.Stream = encoded
	return nil
}

// ToImage converts an object to an Image which can be transformed or saved out.
// The image data is decoded and the Image returned.
func (ximg *XObjectImage) ToImage() (*Image, error) {
	image := &Image{}

	if ximg.Height == nil {
		return nil, errors.New("height attribute missing")
	}
	image.Height = *ximg.Height

	if ximg.Width == nil {
		return nil, errors.New("width attribute missing")
	}
	image.Width = *ximg.Width

	if ximg.BitsPerComponent == nil {
		return nil, errors.New("bits per component missing")
	}
	image.BitsPerComponent = *ximg.BitsPerComponent

	image.ColorComponents = ximg.ColorSpace.GetNumComponents()

	decoded, err := core.DecodeStream(ximg.primitive)
	if err != nil {
		return nil, err
	}
	image.Data = decoded

	if ximg.Decode != nil {
		darr, ok := ximg.Decode.(*core.PdfObjectArray)
		if !ok {
			common.Log.Debug("Invalid Decode object")
			return nil, errors.New("invalid type")
		}
		decode, err := darr.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		image.decode = decode
	}

	return image, nil
}

// GetContainingPdfObject returns the container of the image object (indirect object).
func (ximg *XObjectImage) GetContainingPdfObject() core.PdfObject {
	return ximg.primitive
}

// ToPdfObject returns a stream object.
func (ximg *XObjectImage) ToPdfObject() core.PdfObject {
	stream := ximg.primitive

	dict := stream.PdfObjectDictionary
	if ximg.Filter != nil {
		//dict.Set("Filter", ximg.Filter)
		// Pre-populate the stream dictionary with the
		// encoding related fields.
		dict = ximg.Filter.MakeStreamDict()
		stream.PdfObjectDictionary = dict
	}
	dict.Set("Type", core.MakeName("XObject"))
	dict.Set("Subtype", core.MakeName("Image"))
	dict.Set("Width", core.MakeInteger(*(ximg.Width)))
	dict.Set("Height", core.MakeInteger(*(ximg.Height)))

	if ximg.BitsPerComponent != nil {
		dict.Set("BitsPerComponent", core.MakeInteger(*(ximg.BitsPerComponent)))
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
	dict.SetIfNotNil("Matte", ximg.Matte)
	dict.SetIfNotNil("Name", ximg.Name)
	dict.SetIfNotNil("StructParent", ximg.StructParent)
	dict.SetIfNotNil("ID", ximg.ID)
	dict.SetIfNotNil("OPI", ximg.OPI)
	dict.SetIfNotNil("Metadata", ximg.Metadata)
	dict.SetIfNotNil("OC", ximg.OC)

	dict.Set("Length", core.MakeInteger(int64(len(ximg.Stream))))
	stream.Stream = ximg.Stream

	return stream
}

// getParamsDict returns *core.PdfObjectDictionary with a set of basic image parameters.
func (ximg *XObjectImage) getParamsDict() *core.PdfObjectDictionary {
	params := core.MakeDict()
	params.Set("Width", core.MakeInteger(*ximg.Width))
	params.Set("Height", core.MakeInteger(*ximg.Height))
	params.Set("ColorComponents", core.MakeInteger(int64(ximg.ColorSpace.GetNumComponents())))
	params.Set("BitsPerComponent", core.MakeInteger(*ximg.BitsPerComponent))
	return params
}
