/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

package pdf

import (
	"bytes"
	"fmt"
)

type PdfObject interface {
	String() string
	DefaultWriteString() string
	// Make a recursive traverse function too with a handler function?
}

type PdfObjectBool bool
type PdfObjectInteger int64
type PdfObjectFloat float64
type PdfObjectString string
type PdfObjectName string
type PdfObjectArray []PdfObject
type PdfObjectDictionary map[PdfObjectName]PdfObject
type PdfObjectNull struct{}

type PdfObjectReference struct {
	ObjectNumber     int64
	GenerationNumber int64
}

type PdfIndirectObject struct {
	PdfObjectReference
	PdfObject
}

type PdfObjectStream struct {
	PdfObjectReference
	*PdfObjectDictionary
	Stream []byte
}

func makeName(s string) *PdfObjectName {
	name := PdfObjectName(s)
	return &name
}

func makeInteger(val int64) *PdfObjectInteger {
	num := PdfObjectInteger(val)
	return &num
}

func makeString(s string) *PdfObjectString {
	str := PdfObjectString(s)
	return &str
}

func (this *PdfObjectBool) String() string {
	if *this {
		return "true"
	} else {
		return "false"
	}
}

func (this *PdfObjectBool) DefaultWriteString() string {
	if *this {
		return "true"
	} else {
		return "false"
	}
}

func (this *PdfObjectInteger) String() string {
	return fmt.Sprintf("%d", *this)
}

func (this *PdfObjectInteger) DefaultWriteString() string {
	return fmt.Sprintf("%d", *this)
}

func (this *PdfObjectFloat) String() string {
	return fmt.Sprintf("%f", *this)
}

func (this *PdfObjectFloat) DefaultWriteString() string {
	return fmt.Sprintf("%f", *this)
}

func (this *PdfObjectString) String() string {
	return fmt.Sprintf("%s", string(*this))
}

func (this *PdfObjectString) DefaultWriteString() string {
	var output bytes.Buffer

	escapeSequences := map[byte]string{
		'\n': "\\n",
		'\r': "\\r",
		'\t': "\\t",
		'\b': "\\b",
		'\f': "\\f",
		'(':  "\\(",
		')':  "\\)",
		'\\': "\\\\",
	}

	output.WriteString("(")
	for i := 0; i < len(*this); i++ {
		char := (*this)[i]
		if escStr, useEsc := escapeSequences[char]; useEsc {
			output.WriteString(escStr)
		} else {
			output.WriteByte(char)
		}
	}
	output.WriteString(")")

	return output.String()
}

func (this *PdfObjectName) String() string {
	return fmt.Sprintf("%s", string(*this))
}

// Regular characters that are outside the range EXCLAMATION MARK(21h)
// (!) to TILDE (7Eh) (~) should be written using the hexadecimal notation.
func isPrintable(char byte) bool {
	if char < 0x21 || char > 0x7E {
		return false
	}
	return true
}

func isDelimiter(char byte) bool {
	if char == '(' || char == ')' {
		return true
	}
	if char == '<' || char == '>' {
		return true
	}
	if char == '[' || char == ']' {
		return true
	}
	if char == '{' || char == '}' {
		return true
	}
	if char == '/' {
		return true
	}
	if char == '%' {
		return true
	}

	return false
}

func (this *PdfObjectName) DefaultWriteString() string {
	var output bytes.Buffer

	if len(*this) > 127 {
		log.Error("Name too long (%s)", *this)
	}

	output.WriteString("/")
	for i := 0; i < len(*this); i++ {
		char := (*this)[i]
		if !isPrintable(char) || char == '#' || isDelimiter(char) {
			output.WriteString(fmt.Sprintf("#%.2x", char))
		} else {
			output.WriteByte(char)
		}
	}

	return output.String()
}

func (this *PdfObjectArray) String() string {
	outStr := "["
	for ind, o := range *this {
		outStr += o.String()
		if ind < (len(*this) - 1) {
			outStr += ", "
		}
	}
	outStr += "]"
	return outStr
}

func (this *PdfObjectArray) DefaultWriteString() string {
	outStr := "["
	for ind, o := range *this {
		outStr += o.DefaultWriteString()
		if ind < (len(*this) - 1) {
			outStr += " "
		}
	}
	outStr += "]"
	return outStr
}

func (this *PdfObjectDictionary) String() string {
	outStr := "Dict("
	for k, v := range *this {
		outStr += fmt.Sprintf("\"%s\": %s, ", k, v.String())
	}
	outStr += ")"
	return outStr
}

func (this *PdfObjectDictionary) DefaultWriteString() string {
	outStr := "<<"
	for k, v := range *this {
		log.Debug("Writing k: %s %T", k, v)
		outStr += k.DefaultWriteString()
		outStr += " "
		outStr += v.DefaultWriteString()
	}
	outStr += ">>"
	return outStr
}

func (this *PdfObjectReference) String() string {
	return fmt.Sprintf("Ref(%d %d)", this.ObjectNumber, this.GenerationNumber)
}

func (this *PdfObjectReference) DefaultWriteString() string {
	return fmt.Sprintf("%d %d R", this.ObjectNumber, this.GenerationNumber)
}

func (this *PdfIndirectObject) String() string {
	// Avoid printing out the object, can cause problems with circular
	// references.
	return fmt.Sprintf("IObject:%d", (*this).ObjectNumber)
}

func (this *PdfIndirectObject) DefaultWriteString() string {
	outStr := fmt.Sprintf("%d 0 R", (*this).ObjectNumber)
	return outStr
}

func (this *PdfObjectStream) String() string {
	return fmt.Sprintf("Object stream %d: %s", this.ObjectNumber, this.PdfObjectDictionary)
}

func (this *PdfObjectStream) DefaultWriteString() string {
	outStr := fmt.Sprintf("%d 0 R", (*this).ObjectNumber)
	return outStr
}

func (this *PdfObjectNull) String() string {
	return "null"
}

func (this *PdfObjectNull) DefaultWriteString() string {
	return "null"
}
