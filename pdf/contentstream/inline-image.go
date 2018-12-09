/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// ContentStreamInlineImage is a representation of an inline image in a Content stream. Everything between the BI and EI operands.
// ContentStreamInlineImage implements the core.PdfObject interface although strictly it is not a PDF object.
type ContentStreamInlineImage struct {
	BitsPerComponent core.PdfObject
	ColorSpace       core.PdfObject
	Decode           core.PdfObject
	DecodeParms      core.PdfObject
	Filter           core.PdfObject
	Height           core.PdfObject
	ImageMask        core.PdfObject
	Intent           core.PdfObject
	Interpolate      core.PdfObject
	Width            core.PdfObject
	stream           []byte
}

// NewInlineImageFromImage makes a new content stream inline image object from an image.
func NewInlineImageFromImage(img model.Image, encoder core.StreamEncoder) (*ContentStreamInlineImage, error) {
	if encoder == nil {
		encoder = core.NewRawEncoder()
	}

	inlineImage := ContentStreamInlineImage{}
	if img.ColorComponents == 1 {
		inlineImage.ColorSpace = core.MakeName("G") // G short for DeviceGray
	} else if img.ColorComponents == 3 {
		inlineImage.ColorSpace = core.MakeName("RGB") // RGB short for DeviceRGB
	} else if img.ColorComponents == 4 {
		inlineImage.ColorSpace = core.MakeName("CMYK") // CMYK short for DeviceCMYK
	} else {
		common.Log.Debug("Invalid number of color components for inline image: %d", img.ColorComponents)
		return nil, errors.New("Invalid number of color components")
	}
	inlineImage.BitsPerComponent = core.MakeInteger(img.BitsPerComponent)
	inlineImage.Width = core.MakeInteger(img.Width)
	inlineImage.Height = core.MakeInteger(img.Height)

	encoded, err := encoder.EncodeBytes(img.Data)
	if err != nil {
		return nil, err
	}

	inlineImage.stream = encoded

	filterName := encoder.GetFilterName()
	if filterName != core.StreamEncodingFilterNameRaw {
		inlineImage.Filter = core.MakeName(filterName)
	}
	// XXX/FIXME: Add decode params?

	return &inlineImage, nil
}

func (img *ContentStreamInlineImage) String() string {
	s := fmt.Sprintf("InlineImage(len=%d)\n", len(img.stream))
	if img.BitsPerComponent != nil {
		s += "- BPC " + img.BitsPerComponent.DefaultWriteString() + "\n"
	}
	if img.ColorSpace != nil {
		s += "- CS " + img.ColorSpace.DefaultWriteString() + "\n"
	}
	if img.Decode != nil {
		s += "- D " + img.Decode.DefaultWriteString() + "\n"
	}
	if img.DecodeParms != nil {
		s += "- DP " + img.DecodeParms.DefaultWriteString() + "\n"
	}
	if img.Filter != nil {
		s += "- F " + img.Filter.DefaultWriteString() + "\n"
	}
	if img.Height != nil {
		s += "- H " + img.Height.DefaultWriteString() + "\n"
	}
	if img.ImageMask != nil {
		s += "- IM " + img.ImageMask.DefaultWriteString() + "\n"
	}
	if img.Intent != nil {
		s += "- Intent " + img.Intent.DefaultWriteString() + "\n"
	}
	if img.Interpolate != nil {
		s += "- I " + img.Interpolate.DefaultWriteString() + "\n"
	}
	if img.Width != nil {
		s += "- W " + img.Width.DefaultWriteString() + "\n"
	}
	return s
}

func (img *ContentStreamInlineImage) DefaultWriteString() string {
	var output bytes.Buffer

	// We do not start with "BI" as that is the operand and is written out separately.
	// Write out the parameters
	s := ""

	if img.BitsPerComponent != nil {
		s += "/BPC " + img.BitsPerComponent.DefaultWriteString() + "\n"
	}
	if img.ColorSpace != nil {
		s += "/CS " + img.ColorSpace.DefaultWriteString() + "\n"
	}
	if img.Decode != nil {
		s += "/D " + img.Decode.DefaultWriteString() + "\n"
	}
	if img.DecodeParms != nil {
		s += "/DP " + img.DecodeParms.DefaultWriteString() + "\n"
	}
	if img.Filter != nil {
		s += "/F " + img.Filter.DefaultWriteString() + "\n"
	}
	if img.Height != nil {
		s += "/H " + img.Height.DefaultWriteString() + "\n"
	}
	if img.ImageMask != nil {
		s += "/IM " + img.ImageMask.DefaultWriteString() + "\n"
	}
	if img.Intent != nil {
		s += "/Intent " + img.Intent.DefaultWriteString() + "\n"
	}
	if img.Interpolate != nil {
		s += "/I " + img.Interpolate.DefaultWriteString() + "\n"
	}
	if img.Width != nil {
		s += "/W " + img.Width.DefaultWriteString() + "\n"
	}
	output.WriteString(s)

	output.WriteString("ID ")
	output.Write(img.stream)
	output.WriteString("\nEI\n")

	return output.String()
}

func (img *ContentStreamInlineImage) GetColorSpace(resources *model.PdfPageResources) (model.PdfColorspace, error) {
	if img.ColorSpace == nil {
		// Default.
		common.Log.Debug("Inline image not having specified colorspace, assuming Gray")
		return model.NewPdfColorspaceDeviceGray(), nil
	}

	// If is an array, then could be an indexed colorspace.
	if arr, isArr := img.ColorSpace.(*core.PdfObjectArray); isArr {
		return newIndexedColorspaceFromPdfObject(arr)
	}

	name, ok := img.ColorSpace.(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("Error: Invalid object type (%T;%+v)", img.ColorSpace, img.ColorSpace)
		return nil, errors.New("Type check error")
	}

	if *name == "G" || *name == "DeviceGray" {
		return model.NewPdfColorspaceDeviceGray(), nil
	} else if *name == "RGB" || *name == "DeviceRGB" {
		return model.NewPdfColorspaceDeviceRGB(), nil
	} else if *name == "CMYK" || *name == "DeviceCMYK" {
		return model.NewPdfColorspaceDeviceCMYK(), nil
	} else if *name == "I" {
		return nil, errors.New("Unsupported Index colorspace")
	} else {
		if resources.ColorSpace == nil {
			// Can also refer to a name in the PDF page resources...
			common.Log.Debug("Error, unsupported inline image colorspace: %s", *name)
			return nil, errors.New("Unknown colorspace")
		}

		cs, has := resources.ColorSpace.Colorspaces[string(*name)]
		if !has {
			// Can also refer to a name in the PDF page resources...
			common.Log.Debug("Error, unsupported inline image colorspace: %s", *name)
			return nil, errors.New("Unknown colorspace")
		}

		return cs, nil
	}

}

func (img *ContentStreamInlineImage) GetEncoder() (core.StreamEncoder, error) {
	return newEncoderFromInlineImage(img)
}

// IsMask check if an image is a mask.
// The image mask entry in the image dictionary specifies that the image data shall be used as a stencil
// mask for painting in the current color. The mask data is 1bpc, grayscale.
func (img *ContentStreamInlineImage) IsMask() (bool, error) {
	if img.ImageMask != nil {
		imMask, ok := img.ImageMask.(*core.PdfObjectBool)
		if !ok {
			common.Log.Debug("Image mask not a boolean")
			return false, errors.New("Invalid object type")
		}

		return bool(*imMask), nil
	} else {
		return false, nil
	}

}

// ToImage export the inline image to Image which can be transformed or exported easily.
// Page resources are needed to look up colorspace information.
func (img *ContentStreamInlineImage) ToImage(resources *model.PdfPageResources) (*model.Image, error) {
	// Decode the imaging data if encoded.
	encoder, err := newEncoderFromInlineImage(img)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("encoder: %+v %T", encoder, encoder)
	common.Log.Trace("inline image: %+v", img)

	decoded, err := encoder.DecodeBytes(img.stream)
	if err != nil {
		return nil, err
	}

	image := &model.Image{}

	// Height.
	if img.Height == nil {
		return nil, errors.New("Height attribute missing")
	}
	height, ok := img.Height.(*core.PdfObjectInteger)
	if !ok {
		return nil, errors.New("Invalid height")
	}
	image.Height = int64(*height)

	// Width.
	if img.Width == nil {
		return nil, errors.New("Width attribute missing")
	}
	width, ok := img.Width.(*core.PdfObjectInteger)
	if !ok {
		return nil, errors.New("Invalid width")
	}
	image.Width = int64(*width)

	// Image mask?
	isMask, err := img.IsMask()
	if err != nil {
		return nil, err
	}

	if isMask {
		// Masks are grayscale 1bpc.
		image.BitsPerComponent = 1
		image.ColorComponents = 1
	} else {
		// BPC.
		if img.BitsPerComponent == nil {
			common.Log.Debug("Inline Bits per component missing - assuming 8")
			image.BitsPerComponent = 8
		} else {
			bpc, ok := img.BitsPerComponent.(*core.PdfObjectInteger)
			if !ok {
				common.Log.Debug("Error invalid bits per component value, type %T", img.BitsPerComponent)
				return nil, errors.New("BPC Type error")
			}
			image.BitsPerComponent = int64(*bpc)
		}

		// Color components.
		if img.ColorSpace != nil {
			cs, err := img.GetColorSpace(resources)
			if err != nil {
				return nil, err
			}
			image.ColorComponents = cs.GetNumComponents()
		} else {
			// Default gray if not specified.
			common.Log.Debug("Inline Image colorspace not specified - assuming 1 color component")
			image.ColorComponents = 1
		}
	}

	image.Data = decoded

	return image, nil
}

// ParseInlineImage parses an inline image from a content stream, both read its properties and binary data.
// When called, "BI" has already been read from the stream.  This function
// finishes reading through "EI" and then returns the ContentStreamInlineImage.
func (csp *ContentStreamParser) ParseInlineImage() (*ContentStreamInlineImage, error) {
	// Reading parameters.
	im := ContentStreamInlineImage{}

	for {
		csp.skipSpaces()
		obj, err, isOperand := csp.parseObject()
		if err != nil {
			return nil, err
		}

		if !isOperand {
			// Not an operand.. Read key value properties..
			param, ok := obj.(*core.PdfObjectName)
			if !ok {
				common.Log.Debug("Invalid inline image property (expecting name) - %T", obj)
				return nil, fmt.Errorf("Invalid inline image property (expecting name) - %T", obj)
			}

			valueObj, err, isOperand := csp.parseObject()
			if err != nil {
				return nil, err
			}
			if isOperand {
				return nil, fmt.Errorf("Not expecting an operand")
			}

			if *param == "BPC" || *param == "BitsPerComponent" {
				im.BitsPerComponent = valueObj
			} else if *param == "CS" || *param == "ColorSpace" {
				im.ColorSpace = valueObj
			} else if *param == "D" || *param == "Decode" {
				im.Decode = valueObj
			} else if *param == "DP" || *param == "DecodeParms" {
				im.DecodeParms = valueObj
			} else if *param == "F" || *param == "Filter" {
				im.Filter = valueObj
			} else if *param == "H" || *param == "Height" {
				im.Height = valueObj
			} else if *param == "IM" {
				im.ImageMask = valueObj
			} else if *param == "Intent" {
				im.Intent = valueObj
			} else if *param == "I" {
				im.Interpolate = valueObj
			} else if *param == "W" || *param == "Width" {
				im.Width = valueObj
			} else {
				return nil, fmt.Errorf("Unknown inline image parameter %s", *param)
			}
		}

		if isOperand {
			operand, ok := obj.(*core.PdfObjectString)
			if !ok {
				return nil, fmt.Errorf("Failed to read inline image - invalid operand")
			}

			if operand.Str() == "EI" {
				// Image fully defined
				common.Log.Trace("Inline image finished...")
				return &im, nil
			} else if operand.Str() == "ID" {
				// Inline image data.
				// Should get a single space (0x20) followed by the data and then EI.
				common.Log.Trace("ID start")

				// Skip the space if its there.
				b, err := csp.reader.Peek(1)
				if err != nil {
					return nil, err
				}
				if core.IsWhiteSpace(b[0]) {
					csp.reader.Discard(1)
				}

				// Unfortunately there is no good way to know how many bytes to read since it
				// depends on the Filter and encoding etc.
				// Therefore we will simply read until we find "<ws>EI<ws>" where <ws> is whitespace
				// although of course that could be a part of the data (even if unlikely).
				im.stream = []byte{}
				state := 0
				var skipBytes []byte
				for {
					c, err := csp.reader.ReadByte()
					if err != nil {
						common.Log.Debug("Unable to find end of image EI in inline image data")
						return nil, err
					}

					if state == 0 {
						if core.IsWhiteSpace(c) {
							skipBytes = []byte{}
							skipBytes = append(skipBytes, c)
							state = 1
						} else {
							im.stream = append(im.stream, c)
						}
					} else if state == 1 {
						skipBytes = append(skipBytes, c)
						if c == 'E' {
							state = 2
						} else {
							im.stream = append(im.stream, skipBytes...)
							skipBytes = []byte{} // Clear.
							// Need an extra check to decide if we fall back to state 0 or 1.
							if core.IsWhiteSpace(c) {
								state = 1
							} else {
								state = 0
							}
						}
					} else if state == 2 {
						skipBytes = append(skipBytes, c)
						if c == 'I' {
							state = 3
						} else {
							im.stream = append(im.stream, skipBytes...)
							skipBytes = []byte{} // Clear.
							state = 0
						}
					} else if state == 3 {
						skipBytes = append(skipBytes, c)
						if core.IsWhiteSpace(c) {
							// image data finished.
							if len(im.stream) > 100 {
								common.Log.Trace("Image stream (%d): % x ...", len(im.stream), im.stream[:100])
							} else {
								common.Log.Trace("Image stream (%d): % x", len(im.stream), im.stream)
							}
							// Exit point.
							return &im, nil
						} else {
							// Seems like "<ws>EI" was part of the data.
							im.stream = append(im.stream, skipBytes...)
							skipBytes = []byte{} // Clear.
							state = 0
						}
					}
				}
				// Never reached (exit point is at end of EI).
			}
		}
	}
}
