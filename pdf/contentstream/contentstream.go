/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"bytes"
	"fmt"

	. "github.com/unidoc/unidoc/pdf/core"
)

type ContentStreamOperation struct {
	Params  []PdfObject
	Operand string
}

type ContentStreamOperations []*ContentStreamOperation

// Check if the content stream operations are fully wrapped (within q ... Q)
func (this *ContentStreamOperations) isWrapped() bool {
	if len(*this) < 2 {
		return false
	}

	depth := 0
	for _, op := range *this {
		if op.Operand == "q" {
			depth++
		} else if op.Operand == "Q" {
			depth--
		} else {
			if depth < 1 {
				return false
			}
		}
	}

	// Should end at depth == 0
	return depth == 0
}

// Wrap entire contents within q ... Q.  If unbalanced, then adds extra Qs at the end.
// Only does if needed. Ensures that when adding new content, one start with all states
// in the default condition.
func (this *ContentStreamOperations) WrapIfNeeded() *ContentStreamOperations {
	if len(*this) == 0 {
		// No need to wrap if empty.
		return this
	}
	if this.isWrapped() {
		return this
	}

	*this = append([]*ContentStreamOperation{{Operand: "q"}}, *this...)

	depth := 0
	for _, op := range *this {
		if op.Operand == "q" {
			depth++
		} else if op.Operand == "Q" {
			depth--
		}
	}

	for depth > 0 {
		*this = append(*this, &ContentStreamOperation{Operand: "Q"})
		depth--
	}

	return this
}

// Convert a set of content stream operations to a content stream byte presentation, i.e. the kind that can be
// stored as a PDF stream or string format.
func (this *ContentStreamOperations) Bytes() []byte {
	var buf bytes.Buffer

	for _, op := range *this {
		if op == nil {
			continue
		}

		if op.Operand == "BI" {
			// Inline image requires special handling.
			buf.WriteString(op.Operand + "\n")
			buf.WriteString(op.Params[0].DefaultWriteString())

		} else {
			// Default handler.
			for _, param := range op.Params {
				buf.WriteString(param.DefaultWriteString())
				buf.WriteString(" ")

			}

			buf.WriteString(op.Operand + "\n")
		}
	}

	return buf.Bytes()
}

// ExtractText parses and extracts all text data in content streams and returns as a string.
// Does not take into account Encoding table, the output is simply the character codes.
//
// Deprecated: More advanced text extraction is offered in package extractor with character encoding support.
func (this *ContentStreamParser) ExtractText() (string, error) {
	operations, err := this.Parse()
	if err != nil {
		return "", err
	}
	inText := false
	xPos, yPos := float64(-1), float64(-1)
	txt := ""
	for _, op := range *operations {
		if op.Operand == "BT" {
			inText = true
		} else if op.Operand == "ET" {
			inText = false
		}
		if op.Operand == "Td" || op.Operand == "TD" || op.Operand == "T*" {
			// Move to next line...
			txt += "\n"
		}
		if op.Operand == "Tm" {
			if len(op.Params) != 6 {
				continue
			}
			xfloat, ok := op.Params[4].(*PdfObjectFloat)
			if !ok {
				xint, ok := op.Params[4].(*PdfObjectInteger)
				if !ok {
					continue
				}
				xfloat = MakeFloat(float64(*xint))
			}
			yfloat, ok := op.Params[5].(*PdfObjectFloat)
			if !ok {
				yint, ok := op.Params[5].(*PdfObjectInteger)
				if !ok {
					continue
				}
				yfloat = MakeFloat(float64(*yint))
			}
			if yPos == -1 {
				yPos = float64(*yfloat)
			} else if yPos > float64(*yfloat) {
				txt += "\n"
				xPos = float64(*xfloat)
				yPos = float64(*yfloat)
				continue
			}
			if xPos == -1 {
				xPos = float64(*xfloat)
			} else if xPos < float64(*xfloat) {
				txt += "\t"
				xPos = float64(*xfloat)
			}
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
				switch v := obj.(type) {
				case *PdfObjectString:
					txt += string(*v)
				case *PdfObjectFloat:
					if *v < -100 {
						txt += " "
					}
				case *PdfObjectInteger:
					if *v < -100 {
						txt += " "
					}
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
