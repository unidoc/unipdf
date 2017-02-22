/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// The content stream parser provides functionality to parse the content stream into a list of
// operands that can then be processed further for rendering or extraction of information.

package model

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

// Content stream parser.
type ContentStreamParser struct {
	reader *bufio.Reader
}

type ContentStreamOperation struct {
	Params  []PdfObject
	Operand string
}

// Create a new instance of the content stream parser from an input content
// stream string.
func NewContentStreamParser(contentStr string) *ContentStreamParser {
	// Each command has parameters and an operand (command).

	parser := ContentStreamParser{}

	buffer := bytes.NewBufferString(contentStr)
	parser.reader = bufio.NewReader(buffer)

	return &parser
}

// A representation of an inline image in a Content stream.
// Everything between the BI and EI operands.
// ContentStreamInlineImage implements the PdfObject interface
// although strictly it is not a PDF object.
type ContentStreamInlineImage struct {
	BitsPerComponent PdfObject
	ColorSpace       PdfObject
	Decode           PdfObject
	DecodeParms      PdfObject
	Filter           PdfObject
	Height           PdfObject
	ImageMask        PdfObject
	Intent           PdfObject
	Interpolate      PdfObject
	Width            PdfObject
	stream           []byte
}

func (this *ContentStreamInlineImage) String() string {
	str := fmt.Sprintf("InlineImage(len=%d)", len(this.stream))
	return str
}

func (this *ContentStreamInlineImage) DefaultWriteString() string {
	var output bytes.Buffer

	// We do not start with "BI" as that is the operand and is written out separately.
	// Write out the parameters
	s := "BPC " + this.BitsPerComponent.DefaultWriteString() + "\n"
	s += "CS " + this.ColorSpace.DefaultWriteString() + "\n"
	s += "D " + this.Decode.DefaultWriteString() + "\n"
	s += "DP " + this.DecodeParms.DefaultWriteString() + "\n"
	s += "F " + this.Filter.DefaultWriteString() + "\n"
	s += "H " + this.Height.DefaultWriteString() + "\n"
	s += "IM " + this.ImageMask.DefaultWriteString() + "\n"
	s += "Intent " + this.Intent.DefaultWriteString() + "\n"
	s += "I " + this.Interpolate.DefaultWriteString() + "\n"
	s += "W " + this.Width.DefaultWriteString() + "\n"
	output.WriteString(s)

	output.WriteString("ID ")
	output.Write(this.stream)

	return output.String()
}

// Export the inline image to Image which can be transformed or exported easily.
func (this *ContentStreamInlineImage) ToImage() (*Image, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// Parse an inline image from a content stream, both read its properties and
// binary data.
// When called, "BI" has already been read from the stream.  This function
// finishes reading through "EI" and then returns the ContentStreamInlineImage.
func (this *ContentStreamParser) ParseInlineImage() (*ContentStreamInlineImage, error) {
	// Reading parameters.
	im := ContentStreamInlineImage{}

	for {
		this.skipSpaces()
		obj, err, isOperand := this.parseObject()
		if err != nil {
			return nil, err
		}

		if !isOperand {
			// Not an operand.. Read key value properties..
			param, ok := obj.(*PdfObjectName)
			if !ok {
				return nil, fmt.Errorf("Invalid inline image property (expecting name) - %T", obj)
			}

			valueObj, err, isOperand := this.parseObject()
			if err != nil {
				return nil, err
			}
			if isOperand {
				return nil, fmt.Errorf("Not expecting an operand")
			}

			if *param == "BPC" {
				im.BitsPerComponent = valueObj
			} else if *param == "CS" {
				im.ColorSpace = valueObj
			} else if *param == "D" {
				im.Decode = valueObj
			} else if *param == "DP" {
				im.DecodeParms = valueObj
			} else if *param == "F" {
				im.Filter = valueObj
			} else if *param == "H" {
				im.Height = valueObj
			} else if *param == "IM" {
				im.ImageMask = valueObj
			} else if *param == "Intent" {
				im.Intent = valueObj
			} else if *param == "I" {
				im.Interpolate = valueObj
			} else if *param == "W" {
				im.Width = valueObj
			} else {
				return nil, fmt.Errorf("Unknown inline image parameter %s", *param)
			}
		}

		if isOperand {
			operand, ok := obj.(*PdfObjectString)
			if !ok {
				return nil, fmt.Errorf("Failed to read inline image - invalid operand")
			}

			if *operand == "EI" {
				// Image fully defined
				common.Log.Debug("Inline image finished...")
				return &im, nil
			} else if *operand == "ID" {
				// Inline image data.
				// Should get a single space (0x20) followed by the data and then EI.
				common.Log.Debug("ID start")

				// Skip the space if its there.
				b, err := this.reader.Peek(1)
				if err != nil {
					return nil, err
				}
				if IsWhiteSpace(b[0]) {
					this.reader.Discard(1)
				}

				// Unfortunately there is no good way to know how many bytes to read since it
				// depends on the Filter and encoding etc.
				// Therefore we will simply read until we find "<ws>EI<ws>" where <ws> is whitespace
				// although of course that could be a part of the data (even if unlikely).
				im.stream = []byte{}
				state := 0
				var skipBytes []byte
				for {
					c, err := this.reader.ReadByte()
					if err != nil {
						common.Log.Debug("Unable to find end of image EI in inline image data")
						return nil, err
					}

					if state == 0 {
						if IsWhiteSpace(c) {
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
							// Need an extra check to decide if we fall back to state 0 or 1.
							if IsWhiteSpace(c) {
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
							state = 0
						}
					} else if state == 3 {
						skipBytes = append(skipBytes, c)
						if IsWhiteSpace(c) {
							// image data finished.
							common.Log.Debug("Image stream (%d): % x", len(im.stream), im.stream)
							// Exit point.
							return &im, nil
						} else {
							// Seems like "<ws>EI" was part of the data.
							im.stream = append(im.stream, skipBytes...)
							state = 0
						}
					}
				}
				// Never reached (exit point is at end of EI).
			}
		}
	}
}

// Parses all commands in content stream, returning a list of operation data.
func (this *ContentStreamParser) Parse() ([]*ContentStreamOperation, error) {
	operations := []*ContentStreamOperation{}

	for {
		operation := ContentStreamOperation{}

		for {
			obj, err, isOperand := this.parseObject()
			if err != nil {
				if err == io.EOF {
					return operations, nil
				}
				return nil, err
			}
			if isOperand {
				operation.Operand = string(*obj.(*PdfObjectString))
				operations = append(operations, &operation)
				break
			} else {
				operation.Params = append(operation.Params, obj)
			}
		}

		if operation.Operand == "BI" {
			// Parse an inline image, reads everything between the "BI" and "EI".
			// The image is stored as the parameter.
			im, err := this.ParseInlineImage()
			if err != nil {
				return nil, err
			}
			operation.Params = append(operation.Params, im)
		}
	}

	common.Log.Debug("Operation list: %v\n", operations)
	return operations, nil
}

// Parses and extracts all text data in content streams and returns as a string.
// Does not take into account Encoding table, the output is simply the character codes.
func (this *ContentStreamParser) ExtractText() (string, error) {
	operations, err := this.Parse()
	if err != nil {
		return "", err
	}
	inText := false
	txt := ""
	for _, op := range operations {
		if op.Operand == "BT" {
			inText = true
		} else if op.Operand == "ET" {
			inText = false
		}
		if op.Operand == "Td" || op.Operand == "TD" || op.Operand == "T*" {
			// Move to next line...
			txt += "\n"
		}
		if inText && op.Operand == "TJ" {
			if len(op.Params) < 1 {
				continue
			}
			paramList, ok := op.Params[0].(*PdfObjectArray)
			if !ok {
				return "", fmt.Errorf("Invalid parameter type, no array (%T)", op.Params[0])
			}
			for _, obj := range *paramList {
				if strObj, ok := obj.(*PdfObjectString); ok {
					txt += string(*strObj)
				}
			}
		} else if inText && op.Operand == "Tj" {
			if len(op.Params) < 1 {
				continue
			}
			param, ok := op.Params[0].(*PdfObjectString)
			if !ok {
				return "", fmt.Errorf("Invalid parameter type, not string (%T)", op.Params[0])
			}
			txt += string(*param)
		}
	}

	return txt, nil
}

// Skip over any spaces.  Returns the number of spaces skipped and
// an error if any.
func (this *ContentStreamParser) skipSpaces() (int, error) {
	cnt := 0
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			return 0, err
		}
		if IsWhiteSpace(bb[0]) {
			this.reader.ReadByte()
			cnt++
		} else {
			break
		}
	}

	return cnt, nil
}

// Skip over comments and spaces. Can handle multi-line comments.
func (this *ContentStreamParser) skipComments() error {
	if _, err := this.skipSpaces(); err != nil {
		return err
	}

	isFirst := true
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return err
		}
		if isFirst && bb[0] != '%' {
			// Not a comment clearly.
			return nil
		} else {
			isFirst = false
		}
		if (bb[0] != '\r') && (bb[0] != '\n') {
			this.reader.ReadByte()
		} else {
			break
		}
	}

	// Call recursively to handle multiline comments.
	return this.skipComments()
}

// Parse a name starting with '/'.
func (this *ContentStreamParser) parseName() (PdfObjectName, error) {
	name := ""
	nameStarted := false
	for {
		bb, err := this.reader.Peek(1)
		if err == io.EOF {
			break // Can happen when loading from object stream.
		}
		if err != nil {
			return PdfObjectName(name), err
		}

		if !nameStarted {
			// Should always start with '/', otherwise not valid.
			if bb[0] == '/' {
				nameStarted = true
				this.reader.ReadByte()
			} else {
				common.Log.Error("Name starting with %s (% x)", bb, bb)
				return PdfObjectName(name), fmt.Errorf("Invalid name: (%c)", bb[0])
			}
		} else {
			if IsWhiteSpace(bb[0]) {
				break
			} else if (bb[0] == '/') || (bb[0] == '[') || (bb[0] == '(') || (bb[0] == ']') || (bb[0] == '<') || (bb[0] == '>') {
				break // Looks like start of next statement.
			} else if bb[0] == '#' {
				hexcode, err := this.reader.Peek(3)
				if err != nil {
					return PdfObjectName(name), err
				}
				this.reader.Discard(3)

				code, err := hex.DecodeString(string(hexcode[1:3]))
				if err != nil {
					return PdfObjectName(name), err
				}
				name += string(code)
			} else {
				b, _ := this.reader.ReadByte()
				name += string(b)
			}
		}
	}
	return PdfObjectName(name), nil
}

// Numeric objects.
// Section 7.3.3.
// Integer or Float.
//
// An integer shall be written as one or more decimal digits optionally
// preceded by a sign. The value shall be interpreted as a signed
// decimal integer and shall be converted to an integer object.
//
// A real value shall be written as one or more decimal digits with an
// optional sign and a leading, trailing, or embedded PERIOD (2Eh)
// (decimal point). The value shall be interpreted as a real number
// and shall be converted to a real object.
//
// Regarding exponential numbers: 7.3.3 Numeric Objects:
// A conforming writer shall not use the PostScript syntax for numbers
// with non-decimal radices (such as 16#FFFE) or in exponential format
// (such as 6.02E23).
// Nontheless, we sometimes get numbers with exponential format, so
// we will support it in the reader (no confusion with other types, so
// no compromise).
func (this *ContentStreamParser) parseNumber() (PdfObject, error) {
	isFloat := false
	allowSigns := true
	numStr := ""
	for {
		common.Log.Debug("Parsing number \"%s\"", numStr)
		bb, err := this.reader.Peek(1)
		if err == io.EOF {
			// GH: EOF handling.  Handle EOF like end of line.  Can happen with
			// encoded object streams that the object is at the end.
			// In other cases, we will get the EOF error elsewhere at any rate.
			break // Handle like EOF
		}
		if err != nil {
			common.Log.Error("ERROR %s", err)
			return nil, err
		}
		if allowSigns && (bb[0] == '-' || bb[0] == '+') {
			// Only appear in the beginning, otherwise serves as a delimiter.
			b, _ := this.reader.ReadByte()
			numStr += string(b)
			allowSigns = false // Only allowed in beginning, and after e (exponential).
		} else if IsDecimalDigit(bb[0]) {
			b, _ := this.reader.ReadByte()
			numStr += string(b)
		} else if bb[0] == '.' {
			b, _ := this.reader.ReadByte()
			numStr += string(b)
			isFloat = true
		} else if bb[0] == 'e' {
			// Exponential number format.
			b, _ := this.reader.ReadByte()
			numStr += string(b)
			isFloat = true
			allowSigns = true
		} else {
			break
		}
	}

	if isFloat {
		fVal, err := strconv.ParseFloat(numStr, 64)
		o := PdfObjectFloat(fVal)
		return &o, err
	} else {
		intVal, err := strconv.ParseInt(numStr, 10, 64)
		o := PdfObjectInteger(intVal)
		return &o, err
	}
}

// A string starts with '(' and ends with ')'.
func (this *ContentStreamParser) parseString() (PdfObjectString, error) {
	this.reader.ReadByte()

	bytes := []byte{}
	count := 1
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			return PdfObjectString(bytes), err
		}

		if bb[0] == '\\' { // Escape sequence.
			this.reader.ReadByte() // Skip the escape \ byte.
			b, err := this.reader.ReadByte()
			if err != nil {
				return PdfObjectString(bytes), err
			}

			// Octal '\ddd' number (base 8).
			if IsOctalDigit(b) {
				bb, err := this.reader.Peek(2)
				if err != nil {
					return PdfObjectString(bytes), err
				}

				numeric := []byte{}
				numeric = append(numeric, b)
				for _, val := range bb {
					if IsOctalDigit(val) {
						numeric = append(numeric, val)
					} else {
						break
					}
				}
				this.reader.Discard(len(numeric) - 1)

				common.Log.Debug("Numeric string \"%s\"", numeric)
				code, err := strconv.ParseUint(string(numeric), 8, 32)
				if err != nil {
					return PdfObjectString(bytes), err
				}
				bytes = append(bytes, byte(code))
				continue
			}

			switch b {
			case 'n':
				bytes = append(bytes, '\n')
			case 'r':
				bytes = append(bytes, '\r')
			case 't':
				bytes = append(bytes, '\t')
			case 'b':
				bytes = append(bytes, '\b')
			case 'f':
				bytes = append(bytes, '\f')
			case '(':
				bytes = append(bytes, '(')
			case ')':
				bytes = append(bytes, ')')
			case '\\':
				bytes = append(bytes, '\\')
			}

			continue
		} else if bb[0] == '(' {
			count++
		} else if bb[0] == ')' {
			count--
			if count == 0 {
				this.reader.ReadByte()
				break
			}
		}

		b, _ := this.reader.ReadByte()
		bytes = append(bytes, b)
	}

	return PdfObjectString(bytes), nil
}

// Starts with '<' ends with '>'.
func (this *ContentStreamParser) parseHexString() (PdfObjectString, error) {
	this.reader.ReadByte()

	hextable := []byte("0123456789abcdefABCDEF")

	tmp := []byte{}
	for {
		this.skipSpaces()

		bb, err := this.reader.Peek(1)
		if err != nil {
			return PdfObjectString(""), err
		}

		if bb[0] == '>' {
			this.reader.ReadByte()
			break
		}

		b, _ := this.reader.ReadByte()
		if bytes.IndexByte(hextable, b) >= 0 {
			tmp = append(tmp, b)
		}
	}

	if len(tmp)%2 == 1 {
		tmp = append(tmp, '0')
	}

	buf, _ := hex.DecodeString(string(tmp))
	return PdfObjectString(buf), nil
}

// Starts with '[' ends with ']'.  Can contain any kinds of direct objects.
func (this *ContentStreamParser) parseArray() (PdfObjectArray, error) {
	arr := make(PdfObjectArray, 0)

	this.reader.ReadByte()

	for {
		this.skipSpaces()

		bb, err := this.reader.Peek(1)
		if err != nil {
			return arr, err
		}

		if bb[0] == ']' {
			this.reader.ReadByte()
			break
		}

		obj, err, _ := this.parseObject()
		if err != nil {
			return arr, err
		}
		arr = append(arr, obj)
	}

	return arr, nil
}

// Parse bool object.
func (this *ContentStreamParser) parseBool() (PdfObjectBool, error) {
	bb, err := this.reader.Peek(4)
	if err != nil {
		return PdfObjectBool(false), err
	}
	if (len(bb) >= 4) && (string(bb[:4]) == "true") {
		this.reader.Discard(4)
		return PdfObjectBool(true), nil
	}

	bb, err = this.reader.Peek(5)
	if err != nil {
		return PdfObjectBool(false), err
	}
	if (len(bb) >= 5) && (string(bb[:5]) == "false") {
		this.reader.Discard(5)
		return PdfObjectBool(false), nil
	}

	return PdfObjectBool(false), errors.New("Unexpected boolean string")
}

// Parse null object.
func (this *ContentStreamParser) parseNull() (PdfObjectNull, error) {
	_, err := this.reader.Discard(4)
	return PdfObjectNull{}, err
}

func (this *ContentStreamParser) parseDict() (*PdfObjectDictionary, error) {
	common.Log.Debug("Reading content stream dict!")

	dict := make(PdfObjectDictionary)

	// Pass the '<<'
	c, _ := this.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("Invalid dict")
	}
	c, _ = this.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("Invalid dict")
	}

	for {
		this.skipSpaces()

		bb, err := this.reader.Peek(2)
		if err != nil {
			return nil, err
		}

		common.Log.Debug("Dict peek: %s (% x)!", string(bb), string(bb))
		if (bb[0] == '>') && (bb[1] == '>') {
			common.Log.Debug("EOF dictionary")
			this.reader.ReadByte()
			this.reader.ReadByte()
			break
		}
		common.Log.Debug("Parse the name!")

		keyName, err := this.parseName()
		common.Log.Debug("Key: %s", keyName)
		if err != nil {
			common.Log.Debug("ERROR Returning name err %s", err)
			return nil, err
		}

		if len(keyName) > 4 && keyName[len(keyName)-4:] == "null" {
			// Some writers have a bug where the null is appended without
			// space.  For example "\Boundsnull"
			newKey := keyName[0 : len(keyName)-4]
			common.Log.Debug("Taking care of null bug (%s)", keyName)
			common.Log.Debug("New key \"%s\" = null", newKey)
			this.skipSpaces()
			bb, _ := this.reader.Peek(1)
			if bb[0] == '/' {
				var nullObj PdfObjectNull
				dict[newKey] = &nullObj
				continue
			}
		}

		this.skipSpaces()

		val, err, _ := this.parseObject()
		if err != nil {
			return nil, err
		}
		dict[keyName] = val

		common.Log.Debug("dict[%s] = %s", keyName, val.String())
	}

	return &dict, nil
}

// An operand is a text command represented by a word.
func (this *ContentStreamParser) parseOperand() (PdfObjectString, error) {
	bytes := []byte{}
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			return PdfObjectString(bytes), err
		}
		if IsDelimiter(bb[0]) {
			break
		}
		if IsWhiteSpace(bb[0]) {
			break
		}

		b, _ := this.reader.ReadByte()
		bytes = append(bytes, b)
	}

	return PdfObjectString(bytes), nil
}

// Parse a generic object.  Returns the object, an error code, and a bool
// value indicating whether the object is an operand.  An operand
// is contained in a pdf string object.
func (this *ContentStreamParser) parseObject() (PdfObject, error, bool) {
	// Determine the kind of object.
	// parse it!
	// make a list of operands, then once operand arrives put into a package.

	this.skipSpaces()
	for {
		bb, err := this.reader.Peek(2)
		if err != nil {
			return nil, err, false
		}

		common.Log.Debug("Peek string: %s", string(bb))
		// Determine type.
		if bb[0] == '%' {
			this.skipComments()
			continue
		} else if bb[0] == '/' {
			name, err := this.parseName()
			common.Log.Debug("->Name: '%s'", name)
			return &name, err, false
		} else if bb[0] == '(' {
			common.Log.Debug("->String!")
			str, err := this.parseString()
			return &str, err, false
		} else if bb[0] == '<' && bb[1] != '<' {
			common.Log.Debug("->Hex String!")
			str, err := this.parseHexString()
			return &str, err, false
		} else if bb[0] == '[' {
			common.Log.Debug("->Array!")
			arr, err := this.parseArray()
			return &arr, err, false
		} else if IsDecimalDigit(bb[0]) || (bb[0] == '-' && IsDecimalDigit(bb[1])) {
			common.Log.Debug("->Number!")
			number, err := this.parseNumber()
			return number, err, false
		} else if bb[0] == '<' && bb[1] == '<' {
			dict, err := this.parseDict()
			return dict, err, false
		} else {
			common.Log.Debug("->Operand or bool?")
			// Let's peek farther to find out.
			bb, _ = this.reader.Peek(5)
			peekStr := string(bb)
			common.Log.Debug("Peek str: %s", peekStr)

			if (len(peekStr) > 3) && (peekStr[:4] == "null") {
				null, err := this.parseNull()
				return &null, err, false
			} else if (len(peekStr) > 4) && (peekStr[:5] == "false") {
				b, err := this.parseBool()
				return &b, err, false
			} else if (len(peekStr) > 3) && (peekStr[:4] == "true") {
				b, err := this.parseBool()
				return &b, err, false
			}

			operand, err := this.parseOperand()
			return &operand, err, true
		}
	}
}
