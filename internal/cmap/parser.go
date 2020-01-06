/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// cMapParser parses CMap character to unicode mapping files.
type cMapParser struct {
	reader *bufio.Reader
}

// cMapParser creates a new instance of the PDF CMap parser from input data.
func newCMapParser(content []byte) *cMapParser {
	parser := cMapParser{}

	buffer := bytes.NewBuffer(content)
	parser.reader = bufio.NewReader(buffer)

	return &parser
}

// parseObject detects the signature at the current file position and parses the corresponding object.
func (p *cMapParser) parseObject() (cmapObject, error) {
	p.skipSpaces()
	for {
		bb, err := p.reader.Peek(2)
		if err != nil {
			return nil, err
		}

		if bb[0] == '%' {
			p.parseComment()
			p.skipSpaces()
			continue
		} else if bb[0] == '/' {
			name, err := p.parseName()
			return name, err
		} else if bb[0] == '(' {
			str, err := p.parseString()
			return str, err
		} else if bb[0] == '[' {
			arr, err := p.parseArray()
			return arr, err
		} else if (bb[0] == '<') && (bb[1] == '<') {
			dict, err := p.parseDict()
			return dict, err
		} else if bb[0] == '<' {
			shex, err := p.parseHexString()
			return shex, err
		} else if core.IsDecimalDigit(bb[0]) || (bb[0] == '-' && core.IsDecimalDigit(bb[1])) {
			number, err := p.parseNumber()
			if err != nil {
				return nil, err
			}
			return number, nil
		} else {
			// Operand?
			operand, err := p.parseOperand()
			if err != nil {
				return nil, err
			}

			return operand, nil
		}
	}
}

// skipSpaces skips over any spaces.  Returns the number of spaces skipped and an error if any.
func (p *cMapParser) skipSpaces() (int, error) {
	cnt := 0
	for {
		bb, err := p.reader.Peek(1)
		if err != nil {
			return 0, err
		}
		if core.IsWhiteSpace(bb[0]) {
			p.reader.ReadByte()
			cnt++
		} else {
			break
		}
	}

	return cnt, nil
}

// parseComment reads a comment line starting with '%'.
func (p *cMapParser) parseComment() (string, error) {
	var r bytes.Buffer

	_, err := p.skipSpaces()
	if err != nil {
		return r.String(), err
	}

	isFirst := true
	for {
		bb, err := p.reader.Peek(1)
		if err != nil {
			common.Log.Debug("parseComment: err=%v", err)
			return r.String(), err
		}
		if isFirst && bb[0] != '%' {
			return r.String(), ErrBadCMapComment
		}
		isFirst = false
		if (bb[0] != '\r') && (bb[0] != '\n') {
			b, _ := p.reader.ReadByte()
			r.WriteByte(b)
		} else {
			break
		}
	}
	return r.String(), nil
}

// parseName parses a name starting with '/'.
func (p *cMapParser) parseName() (cmapName, error) {
	name := ""
	nameStarted := false
	for {
		bb, err := p.reader.Peek(1)
		if err == io.EOF {
			break // Can happen when loading from object stream.
		}
		if err != nil {
			return cmapName{name}, err
		}

		if !nameStarted {
			// Should always start with '/', otherwise not valid.
			if bb[0] == '/' {
				nameStarted = true
				p.reader.ReadByte()
			} else {
				common.Log.Debug("ERROR: Name starting with %s (% x)", bb, bb)
				return cmapName{name}, fmt.Errorf("invalid name: (%c)", bb[0])
			}
		} else {
			if core.IsWhiteSpace(bb[0]) {
				break
			} else if (bb[0] == '/') || (bb[0] == '[') || (bb[0] == '(') || (bb[0] == ']') || (bb[0] == '<') || (bb[0] == '>') {
				break // Looks like start of next statement.
			} else if bb[0] == '#' {
				hexcode, err := p.reader.Peek(3)
				if err != nil {
					return cmapName{name}, err
				}
				p.reader.Discard(3)

				code, err := hex.DecodeString(string(hexcode[1:3]))
				if err != nil {
					return cmapName{name}, err
				}
				name += string(code)
			} else {
				b, _ := p.reader.ReadByte()
				name += string(b)
			}
		}
	}

	return cmapName{name}, nil
}

// parseString parses a string starts with '(' and ends with ')'.
func (p *cMapParser) parseString() (cmapString, error) {
	p.reader.ReadByte()

	buf := bytes.Buffer{}

	count := 1
	for {
		bb, err := p.reader.Peek(1)
		if err != nil {
			return cmapString{buf.String()}, err
		}

		if bb[0] == '\\' { // Escape sequence.
			p.reader.ReadByte() // Skip the escape \ byte.
			b, err := p.reader.ReadByte()
			if err != nil {
				return cmapString{buf.String()}, err
			}

			// Octal '\ddd' number (base 8).
			if core.IsOctalDigit(b) {
				bb, err := p.reader.Peek(2)
				if err != nil {
					return cmapString{buf.String()}, err
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
				p.reader.Discard(len(numeric) - 1)

				common.Log.Trace("Numeric string \"%s\"", numeric)
				code, err := strconv.ParseUint(string(numeric), 8, 32)
				if err != nil {
					return cmapString{buf.String()}, err
				}
				buf.WriteByte(byte(code))
				continue
			}

			switch b {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case '(':
				buf.WriteByte('(')
			case ')':
				buf.WriteByte(')')
			case '\\':
				buf.WriteByte('\\')
			}

			continue
		} else if bb[0] == '(' {
			count++
		} else if bb[0] == ')' {
			count--
			if count == 0 {
				p.reader.ReadByte()
				break
			}
		}

		b, _ := p.reader.ReadByte()
		buf.WriteByte(b)
	}

	return cmapString{buf.String()}, nil
}

// parseHexString parses a PostScript hex string.
// Hex strings start with '<' ends with '>'.
// Currently not converting the hex codes to characters.
func (p *cMapParser) parseHexString() (cmapHexString, error) {
	p.reader.ReadByte()

	hextable := []byte("0123456789abcdefABCDEF")

	buf := bytes.Buffer{}

	for {
		p.skipSpaces()

		bb, err := p.reader.Peek(1)
		if err != nil {
			return cmapHexString{}, err
		}

		if bb[0] == '>' {
			p.reader.ReadByte()
			break
		}

		b, _ := p.reader.ReadByte()
		if bytes.IndexByte(hextable, b) >= 0 {
			buf.WriteByte(b)
		}
	}

	if buf.Len()%2 == 1 {
		common.Log.Debug("parseHexString: appending '0' to %#q", buf.String())
		buf.WriteByte('0')
	}

	numBytes := buf.Len() / 2
	hexb, _ := hex.DecodeString(buf.String())
	return cmapHexString{numBytes: numBytes, b: hexb}, nil
}

// parseArray parses a PDF array, which starts with '[', ends with ']'and can contain any kinds of
// direct objects.
func (p *cMapParser) parseArray() (cmapArray, error) {
	arr := cmapArray{}
	arr.Array = []cmapObject{}

	p.reader.ReadByte()

	for {
		p.skipSpaces()

		bb, err := p.reader.Peek(1)
		if err != nil {
			return arr, err
		}

		if bb[0] == ']' {
			p.reader.ReadByte()
			break
		}

		obj, err := p.parseObject()
		if err != nil {
			return arr, err
		}
		arr.Array = append(arr.Array, obj)
	}

	return arr, nil
}

// parseDict parses a PDF dictionary object, which starts with with '<<' and ends with '>>'.
func (p *cMapParser) parseDict() (cmapDict, error) {
	common.Log.Trace("Reading PDF Dict!")

	dict := makeDict()

	// Pass the '<<'
	c, _ := p.reader.ReadByte()
	if c != '<' {
		return dict, ErrBadCMapDict
	}
	c, _ = p.reader.ReadByte()
	if c != '<' {
		return dict, ErrBadCMapDict
	}

	for {
		p.skipSpaces()

		bb, err := p.reader.Peek(2)
		if err != nil {
			return dict, err
		}

		if (bb[0] == '>') && (bb[1] == '>') {
			p.reader.ReadByte()
			p.reader.ReadByte()
			break
		}

		key, err := p.parseName()
		common.Log.Trace("Key: %s", key.Name)
		if err != nil {
			common.Log.Debug("ERROR: Returning name. err=%v", err)
			return dict, err
		}

		p.skipSpaces()

		val, err := p.parseObject()
		if err != nil {
			return dict, err
		}
		dict.Dict[key.Name] = val

		// Skip "def" which optionally follows key value dict definitions in CMaps.
		p.skipSpaces()
		bb, err = p.reader.Peek(3)
		if err != nil {
			return dict, err
		}
		if string(bb) == "def" {
			p.reader.Discard(3)
		}

	}

	return dict, nil
}

// parseDict parseNumber a PDF number.
func (p *cMapParser) parseNumber() (cmapObject, error) {
	o, err := core.ParseNumber(p.reader)
	if err != nil {
		return nil, err
	}

	switch o := o.(type) {
	case *core.PdfObjectFloat:
		return cmapFloat{float64(*o)}, nil
	case *core.PdfObjectInteger:
		return cmapInt{int64(*o)}, nil
	}

	return nil, fmt.Errorf("unhandled number type %T", o)
}

// parseOperand parses an operand, which is a text command represented by a word.
func (p *cMapParser) parseOperand() (cmapOperand, error) {
	op := cmapOperand{}

	buf := bytes.Buffer{}
	for {
		bb, err := p.reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			return op, err
		}
		if core.IsDelimiter(bb[0]) {
			break
		}
		if core.IsWhiteSpace(bb[0]) {
			break
		}

		b, _ := p.reader.ReadByte()
		buf.WriteByte(b)
	}

	if buf.Len() == 0 {
		return op, fmt.Errorf("invalid operand (empty)")
	}

	op.Operand = buf.String()

	return op, nil
}
