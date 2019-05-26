/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"bytes"
	"fmt"

	"github.com/unidoc/unipdf/v3/core"
)

// ContentStreamOperation represents an operation in PDF contentstream which consists of
// an operand and parameters.
type ContentStreamOperation struct {
	Params  []core.PdfObject
	Operand string
}

// ContentStreamOperations is a slice of ContentStreamOperations.
type ContentStreamOperations []*ContentStreamOperation

// Check if the content stream operations are fully wrapped (within q ... Q)
func (ops *ContentStreamOperations) isWrapped() bool {
	if len(*ops) < 2 {
		return false
	}

	depth := 0
	for _, op := range *ops {
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

// WrapIfNeeded wraps the entire contents within q ... Q.  If unbalanced, then adds extra Qs at the end.
// Only does if needed. Ensures that when adding new content, one start with all states
// in the default condition.
func (ops *ContentStreamOperations) WrapIfNeeded() *ContentStreamOperations {
	if len(*ops) == 0 {
		// No need to wrap if empty.
		return ops
	}
	if ops.isWrapped() {
		return ops
	}

	*ops = append([]*ContentStreamOperation{{Operand: "q"}}, *ops...)

	depth := 0
	for _, op := range *ops {
		if op.Operand == "q" {
			depth++
		} else if op.Operand == "Q" {
			depth--
		}
	}

	for depth > 0 {
		*ops = append(*ops, &ContentStreamOperation{Operand: "Q"})
		depth--
	}

	return ops
}

// Bytes converts a set of content stream operations to a content stream byte presentation,
// i.e. the kind that can be stored as a PDF stream or string format.
func (ops *ContentStreamOperations) Bytes() []byte {
	var buf bytes.Buffer

	for _, op := range *ops {
		if op == nil {
			continue
		}

		if op.Operand == "BI" {
			// Inline image requires special handling.
			buf.WriteString(op.Operand + "\n")
			buf.WriteString(op.Params[0].WriteString())

		} else {
			// Default handler.
			for _, param := range op.Params {
				buf.WriteString(param.WriteString())
				buf.WriteString(" ")

			}

			buf.WriteString(op.Operand + "\n")
		}
	}

	return buf.Bytes()
}

// String returns `ops.Bytes()` as a string.
func (ops *ContentStreamOperations) String() string {
	return string(ops.Bytes())
}

// ExtractText parses and extracts all text data in content streams and returns as a string.
// Does not take into account Encoding table, the output is simply the character codes.
//
// Deprecated: More advanced text extraction is offered in package extractor with character encoding support.
func (csp *ContentStreamParser) ExtractText() (string, error) {
	operations, err := csp.Parse()
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
			xfloat, ok := op.Params[4].(*core.PdfObjectFloat)
			if !ok {
				xint, ok := op.Params[4].(*core.PdfObjectInteger)
				if !ok {
					continue
				}
				xfloat = core.MakeFloat(float64(*xint))
			}
			yfloat, ok := op.Params[5].(*core.PdfObjectFloat)
			if !ok {
				yint, ok := op.Params[5].(*core.PdfObjectInteger)
				if !ok {
					continue
				}
				yfloat = core.MakeFloat(float64(*yint))
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
			paramList, ok := op.Params[0].(*core.PdfObjectArray)
			if !ok {
				return "", fmt.Errorf("invalid parameter type, no array (%T)", op.Params[0])
			}
			for _, obj := range paramList.Elements() {
				switch v := obj.(type) {
				case *core.PdfObjectString:
					txt += v.Str()
				case *core.PdfObjectFloat:
					if *v < -100 {
						txt += " "
					}
				case *core.PdfObjectInteger:
					if *v < -100 {
						txt += " "
					}
				}
			}
		} else if inText && op.Operand == "Tj" {
			if len(op.Params) < 1 {
				continue
			}
			param, ok := op.Params[0].(*core.PdfObjectString)
			if !ok {
				return "", fmt.Errorf("invalid parameter type, not string (%T)", op.Params[0])
			}
			txt += param.Str()
		}
	}

	return txt, nil
}
