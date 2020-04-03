/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
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
	encoder.UpdateParams(img.GetParamsDict())

	inlineImage := ContentStreamInlineImage{}
	if img.ColorComponents == 1 {
		inlineImage.ColorSpace = core.MakeName("G") // G short for DeviceGray
	} else if img.ColorComponents == 3 {
		inlineImage.ColorSpace = core.MakeName("RGB") // RGB short for DeviceRGB
	} else if img.ColorComponents == 4 {
		inlineImage.ColorSpace = core.MakeName("CMYK") // CMYK short for DeviceCMYK
	} else {
		common.Log.Debug("Invalid number of color components for inline image: %d", img.ColorComponents)
		return nil, errors.New("invalid number of color components")
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
	// FIXME: Add decode params?

	return &inlineImage, nil
}

func (img *ContentStreamInlineImage) String() string {
	s := fmt.Sprintf("InlineImage(len=%d)\n", len(img.stream))
	if img.BitsPerComponent != nil {
		s += "- BPC " + img.BitsPerComponent.WriteString() + "\n"
	}
	if img.ColorSpace != nil {
		s += "- CS " + img.ColorSpace.WriteString() + "\n"
	}
	if img.Decode != nil {
		s += "- D " + img.Decode.WriteString() + "\n"
	}
	if img.DecodeParms != nil {
		s += "- DP " + img.DecodeParms.WriteString() + "\n"
	}
	if img.Filter != nil {
		s += "- F " + img.Filter.WriteString() + "\n"
	}
	if img.Height != nil {
		s += "- H " + img.Height.WriteString() + "\n"
	}
	if img.ImageMask != nil {
		s += "- IM " + img.ImageMask.WriteString() + "\n"
	}
	if img.Intent != nil {
		s += "- Intent " + img.Intent.WriteString() + "\n"
	}
	if img.Interpolate != nil {
		s += "- I " + img.Interpolate.WriteString() + "\n"
	}
	if img.Width != nil {
		s += "- W " + img.Width.WriteString() + "\n"
	}
	return s
}

// WriteString outputs the object as it is to be written to file.
func (img *ContentStreamInlineImage) WriteString() string {
	var output bytes.Buffer

	// We do not start with "BI" as that is the operand and is written out separately.
	// Write out the parameters
	s := ""

	if img.BitsPerComponent != nil {
		s += "/BPC " + img.BitsPerComponent.WriteString() + "\n"
	}
	if img.ColorSpace != nil {
		s += "/CS " + img.ColorSpace.WriteString() + "\n"
	}
	if img.Decode != nil {
		s += "/D " + img.Decode.WriteString() + "\n"
	}
	if img.DecodeParms != nil {
		s += "/DP " + img.DecodeParms.WriteString() + "\n"
	}
	if img.Filter != nil {
		s += "/F " + img.Filter.WriteString() + "\n"
	}
	if img.Height != nil {
		s += "/H " + img.Height.WriteString() + "\n"
	}
	if img.ImageMask != nil {
		s += "/IM " + img.ImageMask.WriteString() + "\n"
	}
	if img.Intent != nil {
		s += "/Intent " + img.Intent.WriteString() + "\n"
	}
	if img.Interpolate != nil {
		s += "/I " + img.Interpolate.WriteString() + "\n"
	}
	if img.Width != nil {
		s += "/W " + img.Width.WriteString() + "\n"
	}
	output.WriteString(s)

	output.WriteString("ID ")
	output.Write(img.stream)
	output.WriteString("\nEI\n")

	return output.String()
}

// GetColorSpace returns the colorspace of the inline image.
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
		return nil, errors.New("type check error")
	}

	if *name == "G" || *name == "DeviceGray" {
		return model.NewPdfColorspaceDeviceGray(), nil
	} else if *name == "RGB" || *name == "DeviceRGB" {
		return model.NewPdfColorspaceDeviceRGB(), nil
	} else if *name == "CMYK" || *name == "DeviceCMYK" {
		return model.NewPdfColorspaceDeviceCMYK(), nil
	} else if *name == "I" || *name == "Indexed" {
		return nil, errors.New("unsupported Index colorspace")
	} else {
		if resources.ColorSpace == nil {
			// Can also refer to a name in the PDF page resources...
			common.Log.Debug("Error, unsupported inline image colorspace: %s", *name)
			return nil, errors.New("unknown colorspace")
		}

		cs, has := resources.GetColorspaceByName(*name)
		if !has {
			// Can also refer to a name in the PDF page resources...
			common.Log.Debug("Error, unsupported inline image colorspace: %s", *name)
			return nil, errors.New("unknown colorspace")
		}

		return cs, nil
	}

}

// GetEncoder returns the encoder of the inline image.
func (img *ContentStreamInlineImage) GetEncoder() (core.StreamEncoder, error) {
	return newEncoderFromInlineImage(img)
}

// IsMask checks if an image is a mask.
// The image mask entry in the image dictionary specifies that the image data shall be used as a stencil
// mask for painting in the current color. The mask data is 1bpc, grayscale.
func (img *ContentStreamInlineImage) IsMask() (bool, error) {
	if img.ImageMask != nil {
		imMask, ok := img.ImageMask.(*core.PdfObjectBool)
		if !ok {
			common.Log.Debug("Image mask not a boolean")
			return false, errors.New("invalid object type")
		}

		return bool(*imMask), nil
	}

	return false, nil
}

// ToImage exports the inline image to Image which can be transformed or exported easily.
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
		return nil, errors.New("height attribute missing")
	}
	height, ok := img.Height.(*core.PdfObjectInteger)
	if !ok {
		return nil, errors.New("invalid height")
	}
	image.Height = int64(*height)

	// Width.
	if img.Width == nil {
		return nil, errors.New("width attribute missing")
	}
	width, ok := img.Width.(*core.PdfObjectInteger)
	if !ok {
		return nil, errors.New("invalid width")
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

// ParseInlineImage parses an inline image from a content stream, both reading its properties and binary data.
// When called, "BI" has already been read from the stream.  This function
// finishes reading through "EI" and then returns the ContentStreamInlineImage.
func (csp *ContentStreamParser) ParseInlineImage() (*ContentStreamInlineImage, error) {
	// Reading parameters.
	im := ContentStreamInlineImage{}

	for {
		csp.skipSpaces()
		obj, isOperand, err := csp.parseObject()
		if err != nil {
			return nil, err
		}

		if !isOperand {
			// Not an operand.. Read key value properties..
			param, ok := core.GetName(obj)
			if !ok {
				common.Log.Debug("Invalid inline image property (expecting name) - %T", obj)
				return nil, fmt.Errorf("invalid inline image property (expecting name) - %T", obj)
			}

			valueObj, isOperand, err := csp.parseObject()
			if err != nil {
				return nil, err
			}
			if isOperand {
				return nil, fmt.Errorf("not expecting an operand")
			}

			// From 8.9.7 "Inline Images" p. 223 (PDF32000_2008):
			// The key-value pairs appearing between the BI and ID operators are analogous to those in the dictionary
			// portion of an image XObject (though the syntax is different).
			// Table 93 shows the entries that are valid for an inline image, all of which shall have the same meanings
			// as in a stream dictionary (see Table 5) or an image dictionary (see Table 89).
			// Entries other than those listed shall be ignored; in particular, the Type, Subtype, and Length
			// entries normally found in a stream or image dictionary are unnecessary.
			// For convenience, the abbreviations shown in the table may be used in place of the fully spelled-out keys.
			// Table 94 shows additional abbreviations that can be used for the names of colour spaces and filters.

			switch *param {
			case "BPC", "BitsPerComponent":
				im.BitsPerComponent = valueObj
			case "CS", "ColorSpace":
				im.ColorSpace = valueObj
			case "D", "Decode":
				im.Decode = valueObj
			case "DP", "DecodeParms":
				im.DecodeParms = valueObj
			case "F", "Filter":
				im.Filter = valueObj
			case "H", "Height":
				im.Height = valueObj
			case "IM", "ImageMask":
				im.ImageMask = valueObj
			case "Intent":
				im.Intent = valueObj
			case "I", "Interpolate":
				im.Interpolate = valueObj
			case "W", "Width":
				im.Width = valueObj
			case "Length", "Subtype", "Type":
				common.Log.Debug("Ignoring inline parameter %s", *param)
			default:
				return nil, fmt.Errorf("unknown inline image parameter %s", *param)
			}
		}

		if isOperand {
			operand, ok := obj.(*core.PdfObjectString)
			if !ok {
				return nil, fmt.Errorf("failed to read inline image - invalid operand")
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
						} else if c == 'E' {
							// Allow cases where EI is not preceded by whitespace.
							// The extra parsing after EI<ws> should be sufficient
							// in order to decide if the image stream ended.
							skipBytes = append(skipBytes, c)
							state = 2
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
							// Whitspace after EI.
							// To ensure that is not a part of encoded image data: Peek up to 20 bytes ahead
							// and check that the following data is valid objects/operands.
							peekbytes, err := csp.reader.Peek(20)
							if err != nil && err != io.EOF {
								return nil, err
							}
							dummyParser := NewContentStreamParser(string(peekbytes))

							// Assume is done, check that the following 3 objects/operands are valid.
							isDone := true
							for i := 0; i < 3; i++ {
								op, isOp, err := dummyParser.parseObject()
								if err != nil {
									if err == io.EOF {
										break
									}
									continue
								}
								if isOp && !isValidOperand(op.String()) {
									isDone = false
									break
								}
							}

							if isDone {
								// Valid object or operand found, i.e. the EI marks the end of the data.
								// -> image data finished.
								if len(im.stream) > 100 {
									common.Log.Trace("Image stream (%d): % x ...", len(im.stream), im.stream[:100])
								} else {
									common.Log.Trace("Image stream (%d): % x", len(im.stream), im.stream)
								}
								// Exit point.
								return &im, nil
							}
						}

						// Seems like "<ws>EI" was part of the data.
						im.stream = append(im.stream, skipBytes...)
						skipBytes = []byte{} // Clear.
						state = 0
					}
				}
				// Never reached (exit point is at end of EI).
			}
		}
	}
}
