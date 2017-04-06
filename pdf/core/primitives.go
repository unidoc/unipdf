/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Defines PDF primitive objects as per the standard. Also defines a PdfObject
// interface allowing to universally work with these objects. It allows
// recursive writing of the objects to file as well and stringifying for
// debug purposes.

package core

import (
	"bytes"
	"fmt"

	"github.com/unidoc/unidoc/common"
)

// PDF Primitives implement the PdfObject interface.
type PdfObject interface {
	String() string             // Output a string representation of the primitive (for debugging).
	DefaultWriteString() string // Output the PDF primitive as expected by the standard.
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

// Quick functions to make pdf objects form primitive objects.
func MakeName(s string) *PdfObjectName {
	name := PdfObjectName(s)
	return &name
}

func MakeInteger(val int64) *PdfObjectInteger {
	num := PdfObjectInteger(val)
	return &num
}

func MakeArray(objects ...PdfObject) *PdfObjectArray {
	array := PdfObjectArray{}
	for _, obj := range objects {
		array = append(array, obj)
	}
	return &array
}

func MakeArrayFromIntegers(vals []int) *PdfObjectArray {
	array := PdfObjectArray{}
	for _, val := range vals {
		array = append(array, MakeInteger(int64(val)))
	}
	return &array
}

func MakeArrayFromFloats(vals []float64) *PdfObjectArray {
	array := PdfObjectArray{}
	for _, val := range vals {
		array = append(array, MakeFloat(val))
	}
	return &array
}

func MakeFloat(val float64) *PdfObjectFloat {
	num := PdfObjectFloat(val)
	return &num
}

func MakeString(s string) *PdfObjectString {
	str := PdfObjectString(s)
	return &str
}

func MakeNull() *PdfObjectNull {
	null := PdfObjectNull{}
	return &null
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

func (this *PdfObjectName) DefaultWriteString() string {
	var output bytes.Buffer

	if len(*this) > 127 {
		common.Log.Debug("ERROR: Name too long (%s)", *this)
	}

	output.WriteString("/")
	for i := 0; i < len(*this); i++ {
		char := (*this)[i]
		if !IsPrintable(char) || char == '#' || IsDelimiter(char) {
			output.WriteString(fmt.Sprintf("#%.2x", char))
		} else {
			output.WriteByte(char)
		}
	}

	return output.String()
}

func (this *PdfObjectArray) ToFloat64Array() ([]float64, error) {
	vals := []float64{}

	for _, obj := range *this {
		if number, is := obj.(*PdfObjectInteger); is {
			vals = append(vals, float64(*number))
		} else if number, is := obj.(*PdfObjectFloat); is {
			vals = append(vals, float64(*number))
		} else {
			return nil, fmt.Errorf("Type error")
		}
	}

	return vals, nil
}

func (this *PdfObjectArray) ToIntegerArray() ([]int, error) {
	vals := []int{}

	for _, obj := range *this {
		if number, is := obj.(*PdfObjectInteger); is {
			vals = append(vals, int(*number))
		} else {
			return nil, fmt.Errorf("Type error")
		}
	}

	return vals, nil
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

func (this *PdfObjectArray) Append(obj PdfObject) {
	*this = append(*this, obj)
}

func getNumberAsFloat(obj PdfObject) (float64, error) {
	if fObj, ok := obj.(*PdfObjectFloat); ok {
		return float64(*fObj), nil
	}

	if iObj, ok := obj.(*PdfObjectInteger); ok {
		return float64(*iObj), nil
	}

	return 0, fmt.Errorf("Not a number")
}

// For numeric array: Get the array in []float64 slice representation.
// Will return error if not entirely numeric.
func (this *PdfObjectArray) GetAsFloat64Slice() ([]float64, error) {
	slice := []float64{}

	for _, obj := range *this {
		obj := TraceToDirectObject(obj)
		number, err := getNumberAsFloat(obj)
		if err != nil {
			return nil, fmt.Errorf("Array element not a number")
		}
		slice = append(slice, number)
	}

	return slice, nil
}

// Merge in key/values from another dictionary.  Overwriting if has same keys.
func (this *PdfObjectDictionary) Merge(another *PdfObjectDictionary) {
	if another != nil {
		for key, val := range *another {
			(*this)[key] = val
		}
	}
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
		common.Log.Trace("Writing k: %s %T %v %v", k, v, k, v)
		outStr += k.DefaultWriteString()
		outStr += " "
		outStr += v.DefaultWriteString()
	}
	outStr += ">>"
	return outStr
}

func (d *PdfObjectDictionary) Set(key PdfObjectName, val PdfObject) {
	(*d)[key] = val
}

// Only use if the original value is a PdfObject.  If for example using *PdfObjectArray or other primitives
// then the nil check will not fail, will be a new interface referring the nil (not a nil PdfObject).
// TODO: Consider removing. Better to avoid the casting and nil check before calling.
func (d *PdfObjectDictionary) SetIfNotNil(key PdfObjectName, val PdfObject) {
	if val != nil {
		(*d)[key] = val
	}
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

// Handy functions to work with primitive objects.
// Traces a pdf object to a direct object.  For example contained
// in indirect objects (can be double referenced even).
//
// Note: This function does not trace/resolve references.
// That needs to be done beforehand.
const TraceMaxDepth = 20

func TraceToDirectObject(obj PdfObject) PdfObject {
	iobj, isIndirectObj := obj.(*PdfIndirectObject)
	depth := 0
	for isIndirectObj == true {
		obj = iobj.PdfObject
		iobj, isIndirectObj = obj.(*PdfIndirectObject)
		depth++
		if depth > TraceMaxDepth {
			common.Log.Error("Trace depth level beyond 20 - error!")
			return nil
		}
	}
	return obj
}
