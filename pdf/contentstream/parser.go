/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

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

// Create a new instance of the content stream parser from an input content
// stream string.
func NewContentStreamParser(contentStr string) *ContentStreamParser {
	// Each command has parameters and an operand (command).
	parser := ContentStreamParser{}

	buffer := bytes.NewBufferString(contentStr + "\n") // Add newline at end to get last operand without EOF error.
	parser.reader = bufio.NewReader(buffer)

	return &parser
}

// Parses all commands in content stream, returning a list of operation data.
func (this *ContentStreamParser) Parse() (*ContentStreamOperations, error) {
	operations := ContentStreamOperations{}

	for {
		operation := ContentStreamOperation{}

		for {
			obj, err, isOperand := this.parseObject()
			if err != nil {
				if err == io.EOF {
					// End of data. Successful exit point.
					return &operations, nil
				}
				return &operations, err
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
				return &operations, err
			}
			operation.Params = append(operation.Params, im)
		}
	}
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
// Nonetheless, we sometimes get numbers with exponential format, so
// we will support it in the reader (no confusion with other types, so
// no compromise).
func (this *ContentStreamParser) parseNumber() (PdfObject, error) {
	isFloat := false
	allowSigns := true
	numStr := ""
	for {
		common.Log.Trace("Parsing number \"%s\"", numStr)
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
		if err != nil {
			common.Log.Debug("Error parsing number %q err=%v. Using 0.0. Output may be incorrect", numStr, err)
			fVal = 0.0
			err = nil
		}
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

				common.Log.Trace("Numeric string \"%s\"", numeric)
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
	common.Log.Trace("Reading content stream dict!")

	dict := MakeDict()

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

		common.Log.Trace("Dict peek: %s (% x)!", string(bb), string(bb))
		if (bb[0] == '>') && (bb[1] == '>') {
			common.Log.Trace("EOF dictionary")
			this.reader.ReadByte()
			this.reader.ReadByte()
			break
		}
		common.Log.Trace("Parse the name!")

		keyName, err := this.parseName()
		common.Log.Trace("Key: %s", keyName)
		if err != nil {
			common.Log.Debug("ERROR Returning name err %s", err)
			return nil, err
		}

		if len(keyName) > 4 && keyName[len(keyName)-4:] == "null" {
			// Some writers have a bug where the null is appended without
			// space.  For example "\Boundsnull"
			newKey := keyName[0 : len(keyName)-4]
			common.Log.Trace("Taking care of null bug (%s)", keyName)
			common.Log.Trace("New key \"%s\" = null", newKey)
			this.skipSpaces()
			bb, _ := this.reader.Peek(1)
			if bb[0] == '/' {
				dict.Set(newKey, MakeNull())
				continue
			}
		}

		this.skipSpaces()

		val, err, _ := this.parseObject()
		if err != nil {
			return nil, err
		}
		dict.Set(keyName, val)

		common.Log.Trace("dict[%s] = %s", keyName, val.String())
	}

	return dict, nil
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

		common.Log.Trace("Peek string: %s", string(bb))
		// Determine type.
		if bb[0] == '%' {
			this.skipComments()
			continue
		} else if bb[0] == '/' {
			name, err := this.parseName()
			common.Log.Trace("->Name: '%s'", name)
			return &name, err, false
		} else if bb[0] == '(' {
			common.Log.Trace("->String!")
			str, err := this.parseString()
			return &str, err, false
		} else if bb[0] == '<' && bb[1] != '<' {
			common.Log.Trace("->Hex String!")
			str, err := this.parseHexString()
			return &str, err, false
		} else if bb[0] == '[' {
			common.Log.Trace("->Array!")
			arr, err := this.parseArray()
			return &arr, err, false
		} else if IsFloatDigit(bb[0]) || (bb[0] == '-' && IsFloatDigit(bb[1])) {
			common.Log.Trace("->Number!")
			number, err := this.parseNumber()
			return number, err, false
		} else if bb[0] == '<' && bb[1] == '<' {
			dict, err := this.parseDict()
			return dict, err, false
		} else {
			common.Log.Trace("->Operand or bool?")
			// Let's peek farther to find out.
			bb, _ = this.reader.Peek(5)
			peekStr := string(bb)
			common.Log.Trace("Peek str: %s", peekStr)

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
