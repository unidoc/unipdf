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

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// ContentStreamParser represents a content stream parser for parsing content streams in PDFs.
type ContentStreamParser struct {
	reader *bufio.Reader
}

// NewContentStreamParser creates a new instance of the content stream parser from an input content
// stream string.
func NewContentStreamParser(contentStr string) *ContentStreamParser {
	// Each command has parameters and an operand (command).
	parser := ContentStreamParser{}

	buffer := bytes.NewBufferString(contentStr + "\n") // Add newline at end to get last operand without EOF error.
	parser.reader = bufio.NewReader(buffer)

	return &parser
}

// Parse parses all commands in content stream, returning a list of operation data.
func (csp *ContentStreamParser) Parse() (*ContentStreamOperations, error) {
	operations := ContentStreamOperations{}

	for {
		operation := ContentStreamOperation{}

		for {
			obj, isOperand, err := csp.parseObject()
			if err != nil {
				if err == io.EOF {
					// End of data. Successful exit point.
					return &operations, nil
				}
				return &operations, err
			}
			if isOperand {
				operation.Operand, _ = core.GetStringVal(obj)
				operations = append(operations, &operation)
				break
			} else {
				operation.Params = append(operation.Params, obj)
			}
		}

		if operation.Operand == "BI" {
			// Parse an inline image, reads everything between the "BI" and "EI".
			// The image is stored as the parameter.
			im, err := csp.ParseInlineImage()
			if err != nil {
				return &operations, err
			}
			operation.Params = append(operation.Params, im)
		}
	}
}

// Skip over any spaces.  Returns the number of spaces skipped and
// an error if any.
func (csp *ContentStreamParser) skipSpaces() (int, error) {
	cnt := 0
	for {
		bb, err := csp.reader.Peek(1)
		if err != nil {
			return 0, err
		}
		if core.IsWhiteSpace(bb[0]) {
			csp.reader.ReadByte()
			cnt++
		} else {
			break
		}
	}

	return cnt, nil
}

// Skip over comments and spaces. Can handle multi-line comments.
func (csp *ContentStreamParser) skipComments() error {
	if _, err := csp.skipSpaces(); err != nil {
		return err
	}

	isFirst := true
	for {
		bb, err := csp.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return err
		}
		if isFirst && bb[0] != '%' {
			// Not a comment clearly.
			return nil
		}
		isFirst = false

		if (bb[0] != '\r') && (bb[0] != '\n') {
			csp.reader.ReadByte()
		} else {
			break
		}
	}

	// Call recursively to handle multiline comments.
	return csp.skipComments()
}

// Parse a name starting with '/'.
func (csp *ContentStreamParser) parseName() (core.PdfObjectName, error) {
	name := ""
	nameStarted := false
	for {
		bb, err := csp.reader.Peek(1)
		if err == io.EOF {
			break // Can happen when loading from object stream.
		}
		if err != nil {
			return core.PdfObjectName(name), err
		}

		if !nameStarted {
			// Should always start with '/', otherwise not valid.
			if bb[0] == '/' {
				nameStarted = true
				csp.reader.ReadByte()
			} else {
				common.Log.Error("Name starting with %s (% x)", bb, bb)
				return core.PdfObjectName(name), fmt.Errorf("invalid name: (%c)", bb[0])
			}
		} else {
			if core.IsWhiteSpace(bb[0]) {
				break
			} else if (bb[0] == '/') || (bb[0] == '[') || (bb[0] == '(') || (bb[0] == ']') || (bb[0] == '<') || (bb[0] == '>') {
				break // Looks like start of next statement.
			} else if bb[0] == '#' {
				hexcode, err := csp.reader.Peek(3)
				if err != nil {
					return core.PdfObjectName(name), err
				}
				csp.reader.Discard(3)

				code, err := hex.DecodeString(string(hexcode[1:3]))
				if err != nil {
					return core.PdfObjectName(name), err
				}
				name += string(code)
			} else {
				b, _ := csp.reader.ReadByte()
				name += string(b)
			}
		}
	}
	return core.PdfObjectName(name), nil
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
func (csp *ContentStreamParser) parseNumber() (core.PdfObject, error) {
	return core.ParseNumber(csp.reader)
}

// A string starts with '(' and ends with ')'.
func (csp *ContentStreamParser) parseString() (*core.PdfObjectString, error) {
	csp.reader.ReadByte()

	var bytes []byte
	count := 1
	for {
		bb, err := csp.reader.Peek(1)
		if err != nil {
			return core.MakeString(string(bytes)), err
		}

		if bb[0] == '\\' { // Escape sequence.
			csp.reader.ReadByte() // Skip the escape \ byte.
			b, err := csp.reader.ReadByte()
			if err != nil {
				return core.MakeString(string(bytes)), err
			}

			// Octal '\ddd' number (base 8).
			if core.IsOctalDigit(b) {
				bb, err := csp.reader.Peek(2)
				if err != nil {
					return core.MakeString(string(bytes)), err
				}

				var numeric []byte
				numeric = append(numeric, b)
				for _, val := range bb {
					if core.IsOctalDigit(val) {
						numeric = append(numeric, val)
					} else {
						break
					}
				}
				csp.reader.Discard(len(numeric) - 1)

				common.Log.Trace("Numeric string \"%s\"", numeric)
				code, err := strconv.ParseUint(string(numeric), 8, 32)
				if err != nil {
					return core.MakeString(string(bytes)), err
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
				csp.reader.ReadByte()
				break
			}
		}

		b, _ := csp.reader.ReadByte()
		bytes = append(bytes, b)
	}

	return core.MakeString(string(bytes)), nil
}

// Starts with '<' ends with '>'.
func (csp *ContentStreamParser) parseHexString() (*core.PdfObjectString, error) {
	csp.reader.ReadByte()

	hextable := []byte("0123456789abcdefABCDEF")

	var tmp []byte
	for {
		csp.skipSpaces()

		bb, err := csp.reader.Peek(1)
		if err != nil {
			return core.MakeString(""), err
		}

		if bb[0] == '>' {
			csp.reader.ReadByte()
			break
		}

		b, _ := csp.reader.ReadByte()
		if bytes.IndexByte(hextable, b) >= 0 {
			tmp = append(tmp, b)
		}
	}

	if len(tmp)%2 == 1 {
		tmp = append(tmp, '0')
	}

	buf, _ := hex.DecodeString(string(tmp))
	return core.MakeHexString(string(buf)), nil
}

// Starts with '[' ends with ']'.  Can contain any kinds of direct objects.
func (csp *ContentStreamParser) parseArray() (*core.PdfObjectArray, error) {
	arr := core.MakeArray()

	csp.reader.ReadByte()

	for {
		csp.skipSpaces()

		bb, err := csp.reader.Peek(1)
		if err != nil {
			return arr, err
		}

		if bb[0] == ']' {
			csp.reader.ReadByte()
			break
		}

		obj, _, err := csp.parseObject()
		if err != nil {
			return arr, err
		}
		arr.Append(obj)
	}

	return arr, nil
}

// Parse bool object.
func (csp *ContentStreamParser) parseBool() (core.PdfObjectBool, error) {
	bb, err := csp.reader.Peek(4)
	if err != nil {
		return core.PdfObjectBool(false), err
	}
	if (len(bb) >= 4) && (string(bb[:4]) == "true") {
		csp.reader.Discard(4)
		return core.PdfObjectBool(true), nil
	}

	bb, err = csp.reader.Peek(5)
	if err != nil {
		return core.PdfObjectBool(false), err
	}
	if (len(bb) >= 5) && (string(bb[:5]) == "false") {
		csp.reader.Discard(5)
		return core.PdfObjectBool(false), nil
	}

	return core.PdfObjectBool(false), errors.New("unexpected boolean string")
}

// Parse null object.
func (csp *ContentStreamParser) parseNull() (core.PdfObjectNull, error) {
	_, err := csp.reader.Discard(4)
	return core.PdfObjectNull{}, err
}

func (csp *ContentStreamParser) parseDict() (*core.PdfObjectDictionary, error) {
	common.Log.Trace("Reading content stream dict!")

	dict := core.MakeDict()

	// Pass the '<<'
	c, _ := csp.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("invalid dict")
	}
	c, _ = csp.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("invalid dict")
	}

	for {
		csp.skipSpaces()

		bb, err := csp.reader.Peek(2)
		if err != nil {
			return nil, err
		}

		common.Log.Trace("Dict peek: %s (% x)!", string(bb), string(bb))
		if (bb[0] == '>') && (bb[1] == '>') {
			common.Log.Trace("EOF dictionary")
			csp.reader.ReadByte()
			csp.reader.ReadByte()
			break
		}
		common.Log.Trace("Parse the name!")

		keyName, err := csp.parseName()
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
			csp.skipSpaces()
			bb, _ := csp.reader.Peek(1)
			if bb[0] == '/' {
				dict.Set(newKey, core.MakeNull())
				continue
			}
		}

		csp.skipSpaces()

		val, _, err := csp.parseObject()
		if err != nil {
			return nil, err
		}
		dict.Set(keyName, val)

		common.Log.Trace("dict[%s] = %s", keyName, val.String())
	}

	return dict, nil
}

// An operand is a text command represented by a word.
func (csp *ContentStreamParser) parseOperand() (*core.PdfObjectString, error) {
	var bytes []byte
	for {
		bb, err := csp.reader.Peek(1)
		if err != nil {
			return core.MakeString(string(bytes)), err
		}
		if core.IsDelimiter(bb[0]) {
			break
		}
		if core.IsWhiteSpace(bb[0]) {
			break
		}

		b, _ := csp.reader.ReadByte()
		bytes = append(bytes, b)
	}

	return core.MakeString(string(bytes)), nil
}

// Parse a generic object.  Returns the object, an error code, and a bool
// value indicating whether the object is an operand.  An operand
// is contained in a pdf string object.
func (csp *ContentStreamParser) parseObject() (obj core.PdfObject, isop bool, err error) {
	// Determine the kind of object.
	// parse it!
	// make a list of operands, then once operand arrives put into a package.

	csp.skipSpaces()
	for {
		bb, err := csp.reader.Peek(2)
		if err != nil {
			return nil, false, err
		}

		common.Log.Trace("Peek string: %s", string(bb))
		// Determine type.
		if bb[0] == '%' {
			csp.skipComments()
			continue
		} else if bb[0] == '/' {
			name, err := csp.parseName()
			common.Log.Trace("->Name: '%s'", name)
			return &name, false, err
		} else if bb[0] == '(' {
			common.Log.Trace("->String!")
			str, err := csp.parseString()
			return str, false, err
		} else if bb[0] == '<' && bb[1] != '<' {
			common.Log.Trace("->Hex String!")
			str, err := csp.parseHexString()
			return str, false, err
		} else if bb[0] == '[' {
			common.Log.Trace("->Array!")
			arr, err := csp.parseArray()
			return arr, false, err
		} else if core.IsFloatDigit(bb[0]) || (bb[0] == '-' && core.IsFloatDigit(bb[1])) {
			common.Log.Trace("->Number!")
			number, err := csp.parseNumber()
			return number, false, err
		} else if bb[0] == '<' && bb[1] == '<' {
			dict, err := csp.parseDict()
			return dict, false, err
		} else {
			// Otherwise, can be: keyword such as "null", "false", "true" or an operand...
			common.Log.Trace("->Operand or bool?")
			// Let's peek farther to find out.
			bb, _ = csp.reader.Peek(5)
			peekStr := string(bb)
			common.Log.Trace("cont Peek str: %s", peekStr)

			if (len(peekStr) > 3) && (peekStr[:4] == "null") {
				null, err := csp.parseNull()
				return &null, false, err
			} else if (len(peekStr) > 4) && (peekStr[:5] == "false") {
				b, err := csp.parseBool()
				return &b, false, err
			} else if (len(peekStr) > 3) && (peekStr[:4] == "true") {
				b, err := csp.parseBool()
				return &b, false, err
			}

			operand, err := csp.parseOperand()
			if err != nil {
				return operand, false, err
			}
			if len(operand.String()) < 1 {
				return operand, false, ErrInvalidOperand
			}
			return operand, true, nil
		}
	}
}
