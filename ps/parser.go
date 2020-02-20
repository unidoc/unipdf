/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ps

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/unidoc/unipdf/v3/common"
	pdfcore "github.com/unidoc/unipdf/v3/core"
)

// PSParser is a basic Postscript parser.
type PSParser struct {
	reader *bufio.Reader
}

// NewPSParser returns a new instance of the PDF Postscript parser from input data.
func NewPSParser(content []byte) *PSParser {
	parser := PSParser{}

	buffer := bytes.NewBuffer(content)
	parser.reader = bufio.NewReader(buffer)

	return &parser
}

// Parse parses the postscript and store as a program that can be executed.
func (p *PSParser) Parse() (*PSProgram, error) {
	p.skipSpaces()
	bb, err := p.reader.Peek(2)
	if err != nil {
		return nil, err
	}
	if bb[0] != '{' {
		return nil, errors.New("invalid PS Program not starting with {")
	}

	program, err := p.parseFunction()
	if err != nil && err != io.EOF {
		return nil, err
	}

	return program, err
}

// Detect the signature at the current parse position and parse
// the corresponding object.
func (p *PSParser) parseFunction() (*PSProgram, error) {
	c, _ := p.reader.ReadByte()
	if c != '{' {
		return nil, errors.New("invalid function")
	}

	function := NewPSProgram()

	for {
		p.skipSpaces()
		bb, err := p.reader.Peek(2)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		common.Log.Trace("Peek string: %s", string(bb))
		// Determine type.
		if bb[0] == '}' {
			common.Log.Trace("EOF function")
			p.reader.ReadByte()
			break
		} else if bb[0] == '{' {
			common.Log.Trace("Function!")
			inlineF, err := p.parseFunction()
			if err != nil {
				return nil, err
			}
			function.Append(inlineF)
		} else if pdfcore.IsDecimalDigit(bb[0]) || (bb[0] == '-' && pdfcore.IsDecimalDigit(bb[1])) {
			common.Log.Trace("->Number!")
			number, err := p.parseNumber()
			if err != nil {
				return nil, err
			}
			function.Append(number)
		} else {
			common.Log.Trace("->Operand or bool?")
			// Let's peek farther to find out.
			bb, _ = p.reader.Peek(5)
			peekStr := string(bb)
			common.Log.Trace("Peek str: %s", peekStr)

			if (len(peekStr) > 4) && (peekStr[:5] == "false") {
				b, err := p.parseBool()
				if err != nil {
					return nil, err
				}
				function.Append(b)
			} else if (len(peekStr) > 3) && (peekStr[:4] == "true") {
				b, err := p.parseBool()
				if err != nil {
					return nil, err
				}
				function.Append(b)
			} else {
				operand, err := p.parseOperand()
				if err != nil {
					return nil, err
				}
				function.Append(operand)
			}
		}
	}

	return function, nil
}

// Skip over any spaces.  Returns the number of spaces skipped and
// an error if any.
func (p *PSParser) skipSpaces() (int, error) {
	cnt := 0
	for {
		bb, err := p.reader.Peek(1)
		if err != nil {
			return 0, err
		}
		if pdfcore.IsWhiteSpace(bb[0]) {
			p.reader.ReadByte()
			cnt++
		} else {
			break
		}
	}

	return cnt, nil
}

// Numeric objects.
// Integer or Real numbers.
func (p *PSParser) parseNumber() (PSObject, error) {
	o, err := pdfcore.ParseNumber(p.reader)
	if err != nil {
		return nil, err
	}

	switch o := o.(type) {
	case *pdfcore.PdfObjectFloat:
		return MakeReal(float64(*o)), nil
	case *pdfcore.PdfObjectInteger:
		return MakeInteger(int(*o)), nil
	}

	return nil, fmt.Errorf("unhandled number type %T", o)
}

// Parse bool object.
func (p *PSParser) parseBool() (*PSBoolean, error) {
	bb, err := p.reader.Peek(4)
	if err != nil {
		return MakeBool(false), err
	}
	if (len(bb) >= 4) && (string(bb[:4]) == "true") {
		p.reader.Discard(4)
		return MakeBool(true), nil
	}

	bb, err = p.reader.Peek(5)
	if err != nil {
		return MakeBool(false), err
	}
	if (len(bb) >= 5) && (string(bb[:5]) == "false") {
		p.reader.Discard(5)
		return MakeBool(false), nil
	}

	return MakeBool(false), errors.New("unexpected boolean string")
}

// An operand is a text command represented by a word.
func (p *PSParser) parseOperand() (*PSOperand, error) {
	var bytes []byte
	for {
		bb, err := p.reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if pdfcore.IsDelimiter(bb[0]) {
			break
		}
		if pdfcore.IsWhiteSpace(bb[0]) {
			break
		}

		b, _ := p.reader.ReadByte()
		bytes = append(bytes, b)
	}

	if len(bytes) == 0 {
		return nil, errors.New("invalid operand (empty)")
	}

	return MakeOperand(string(bytes)), nil
}
