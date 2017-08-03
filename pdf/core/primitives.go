/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bytes"
	"fmt"

	"github.com/unidoc/unidoc/common"
)

// PdfObject is an interface which all primitive PDF objects must implement.
type PdfObject interface {
	// Output a string representation of the primitive (for debugging).
	String() string

	// Output the PDF primitive as written to file as expected by the standard.
	DefaultWriteString() string
}

// PdfObjectBool represents the primitive PDF boolean object.
type PdfObjectBool bool

// PdfObjectInteger represents the primitive PDF integer numerical object.
type PdfObjectInteger int64

// PdfObjectFloat represents the primitive PDF floating point numerical object.
type PdfObjectFloat float64

// PdfObjectString represents the primitive PDF string object.
// TODO (v3): Change to a struct and add a flag for hex/plaintext.
type PdfObjectString string

// PdfObjectName represents the primitive PDF name object.
type PdfObjectName string

// PdfObjectArray represents the primitive PDF array object.
type PdfObjectArray []PdfObject

// PdfObjectDictionary represents the primitive PDF dictionary/map object.
type PdfObjectDictionary struct {
	dict map[PdfObjectName]PdfObject
	keys []PdfObjectName
}

// PdfObjectNull represents the primitive PDF null object.
type PdfObjectNull struct{}

// PdfObjectReference represents the primitive PDF reference object.
type PdfObjectReference struct {
	ObjectNumber     int64
	GenerationNumber int64
}

// PdfIndirectObject represents the primitive PDF indirect object.
type PdfIndirectObject struct {
	PdfObjectReference
	PdfObject
}

// PdfObjectStream represents the primitive PDF Object stream.
type PdfObjectStream struct {
	PdfObjectReference
	*PdfObjectDictionary
	Stream []byte
}

// MakeDict creates and returns an empty PdfObjectDictionary.
func MakeDict() *PdfObjectDictionary {
	d := &PdfObjectDictionary{}
	d.dict = map[PdfObjectName]PdfObject{}
	d.keys = []PdfObjectName{}
	return d
}

// MakeName creates a PdfObjectName from a string.
func MakeName(s string) *PdfObjectName {
	name := PdfObjectName(s)
	return &name
}

// MakeInteger creates a PdfObjectInteger from an int64.
func MakeInteger(val int64) *PdfObjectInteger {
	num := PdfObjectInteger(val)
	return &num
}

// MakeArray creates an PdfObjectArray from a list of PdfObjects.
func MakeArray(objects ...PdfObject) *PdfObjectArray {
	array := PdfObjectArray{}
	for _, obj := range objects {
		array = append(array, obj)
	}
	return &array
}

// MakeArrayFromIntegers creates an PdfObjectArray from a slice of ints, where each array element is
// an PdfObjectInteger.
func MakeArrayFromIntegers(vals []int) *PdfObjectArray {
	array := PdfObjectArray{}
	for _, val := range vals {
		array = append(array, MakeInteger(int64(val)))
	}
	return &array
}

// MakeArrayFromIntegers64 creates an PdfObjectArray from a slice of int64s, where each array element
// is an PdfObjectInteger.
func MakeArrayFromIntegers64(vals []int64) *PdfObjectArray {
	array := PdfObjectArray{}
	for _, val := range vals {
		array = append(array, MakeInteger(val))
	}
	return &array
}

// MakeArrayFromFloats creates an PdfObjectArray from a slice of float64s, where each array element is an
// PdfObjectFloat.
func MakeArrayFromFloats(vals []float64) *PdfObjectArray {
	array := PdfObjectArray{}
	for _, val := range vals {
		array = append(array, MakeFloat(val))
	}
	return &array
}

// MakeFloat creates an PdfObjectFloat from a float64.
func MakeFloat(val float64) *PdfObjectFloat {
	num := PdfObjectFloat(val)
	return &num
}

// MakeString creates an PdfObjectString from a string.
func MakeString(s string) *PdfObjectString {
	str := PdfObjectString(s)
	return &str
}

// MakeNull creates an PdfObjectNull.
func MakeNull() *PdfObjectNull {
	null := PdfObjectNull{}
	return &null
}

// MakeIndirectObject creates an PdfIndirectObject with a specified direct object PdfObject.
func MakeIndirectObject(obj PdfObject) *PdfIndirectObject {
	ind := &PdfIndirectObject{}
	ind.PdfObject = obj
	return ind
}

// MakeStream creates an PdfObjectStream with specified contents and encoding. If encoding is nil, then raw encoding
// will be used (i.e. no encoding applied).
func MakeStream(contents []byte, encoder StreamEncoder) (*PdfObjectStream, error) {
	stream := &PdfObjectStream{}

	if encoder == nil {
		encoder = NewRawEncoder()
	}

	stream.PdfObjectDictionary = encoder.MakeStreamDict()

	encoded, err := encoder.EncodeBytes(contents)
	if err != nil {
		return nil, err
	}
	stream.PdfObjectDictionary.Set("Length", MakeInteger(int64(len(encoded))))

	stream.Stream = encoded
	return stream, nil
}

func (bool *PdfObjectBool) String() string {
	if *bool {
		return "true"
	} else {
		return "false"
	}
}

// DefaultWriteString outputs the object as it is to be written to file.
func (bool *PdfObjectBool) DefaultWriteString() string {
	if *bool {
		return "true"
	} else {
		return "false"
	}
}

func (int *PdfObjectInteger) String() string {
	return fmt.Sprintf("%d", *int)
}

// DefaultWriteString outputs the object as it is to be written to file.
func (int *PdfObjectInteger) DefaultWriteString() string {
	return fmt.Sprintf("%d", *int)
}

func (float *PdfObjectFloat) String() string {
	return fmt.Sprintf("%f", *float)
}

// DefaultWriteString outputs the object as it is to be written to file.
func (float *PdfObjectFloat) DefaultWriteString() string {
	return fmt.Sprintf("%f", *float)
}

func (str *PdfObjectString) String() string {
	return fmt.Sprintf("%s", string(*str))
}

// DefaultWriteString outputs the object as it is to be written to file.
func (str *PdfObjectString) DefaultWriteString() string {
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
	for i := 0; i < len(*str); i++ {
		char := (*str)[i]
		if escStr, useEsc := escapeSequences[char]; useEsc {
			output.WriteString(escStr)
		} else {
			output.WriteByte(char)
		}
	}
	output.WriteString(")")

	return output.String()
}

func (name *PdfObjectName) String() string {
	return fmt.Sprintf("%s", string(*name))
}

// DefaultWriteString outputs the object as it is to be written to file.
func (name *PdfObjectName) DefaultWriteString() string {
	var output bytes.Buffer

	if len(*name) > 127 {
		common.Log.Debug("ERROR: Name too long (%s)", *name)
	}

	output.WriteString("/")
	for i := 0; i < len(*name); i++ {
		char := (*name)[i]
		if !IsPrintable(char) || char == '#' || IsDelimiter(char) {
			output.WriteString(fmt.Sprintf("#%.2x", char))
		} else {
			output.WriteByte(char)
		}
	}

	return output.String()
}

// ToFloat64Array returns a slice of all elements in the array as a float64 slice.  An error is returned if the array
// contains non-numeric objects (each element can be either PdfObjectInteger or PdfObjectFloat).
func (array *PdfObjectArray) ToFloat64Array() ([]float64, error) {
	vals := []float64{}

	for _, obj := range *array {
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

// ToIntegerArray returns a slice of all array elements as an int slice. An error is returned if the array contains
// non-integer objects. Each element can only be PdfObjectInteger.
func (array *PdfObjectArray) ToIntegerArray() ([]int, error) {
	vals := []int{}

	for _, obj := range *array {
		if number, is := obj.(*PdfObjectInteger); is {
			vals = append(vals, int(*number))
		} else {
			return nil, fmt.Errorf("Type error")
		}
	}

	return vals, nil
}

func (array *PdfObjectArray) String() string {
	outStr := "["
	for ind, o := range *array {
		outStr += o.String()
		if ind < (len(*array) - 1) {
			outStr += ", "
		}
	}
	outStr += "]"
	return outStr
}

// DefaultWriteString outputs the object as it is to be written to file.
func (array *PdfObjectArray) DefaultWriteString() string {
	outStr := "["
	for ind, o := range *array {
		outStr += o.DefaultWriteString()
		if ind < (len(*array) - 1) {
			outStr += " "
		}
	}
	outStr += "]"
	return outStr
}

// Append adds an PdfObject to the array.
func (array *PdfObjectArray) Append(obj PdfObject) {
	*array = append(*array, obj)
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

// GetAsFloat64Slice returns the array as []float64 slice.
// Returns an error if not entirely numeric (only PdfObjectIntegers, PdfObjectFloats).
func (array *PdfObjectArray) GetAsFloat64Slice() ([]float64, error) {
	slice := []float64{}

	for _, obj := range *array {
		obj := TraceToDirectObject(obj)
		number, err := getNumberAsFloat(obj)
		if err != nil {
			return nil, fmt.Errorf("Array element not a number")
		}
		slice = append(slice, number)
	}

	return slice, nil
}

// Merge merges in key/values from another dictionary. Overwriting if has same keys.
func (d *PdfObjectDictionary) Merge(another *PdfObjectDictionary) {
	if another != nil {
		for _, key := range another.Keys() {
			val := another.Get(key)
			d.Set(key, val)
		}
	}
}

func (d *PdfObjectDictionary) String() string {
	outStr := "Dict("
	for _, k := range d.keys {
		v := d.dict[k]
		outStr += fmt.Sprintf("\"%s\": %s, ", k, v.String())
	}
	outStr += ")"
	return outStr
}

// DefaultWriteString outputs the object as it is to be written to file.
func (d *PdfObjectDictionary) DefaultWriteString() string {
	outStr := "<<"
	for _, k := range d.keys {
		v := d.dict[k]
		common.Log.Trace("Writing k: %s %T %v %v", k, v, k, v)
		outStr += k.DefaultWriteString()
		outStr += " "
		outStr += v.DefaultWriteString()
	}
	outStr += ">>"
	return outStr
}

// Set sets the dictionary's key -> val mapping entry. Overwrites if key already set.
func (d *PdfObjectDictionary) Set(key PdfObjectName, val PdfObject) {
	found := false
	for _, k := range d.keys {
		if k == key {
			found = true
			break
		}
	}

	if !found {
		d.keys = append(d.keys, key)
	}

	d.dict[key] = val
}

// Get returns the PdfObject corresponding to the specified key.
// Returns a nil value if the key is not set.
//
// The design is such that we only return 1 value.
// The reason is that, it will be easy to do type casts such as
// name, ok := dict.Get("mykey").(*PdfObjectName)
// if !ok ....
func (d *PdfObjectDictionary) Get(key PdfObjectName) PdfObject {
	val, has := d.dict[key]
	if !has {
		return nil
	}
	return val
}

// Keys returns the list of keys in the dictionary.
func (d *PdfObjectDictionary) Keys() []PdfObjectName {
	return d.keys
}

// Remove removes an element specified by key.
func (d *PdfObjectDictionary) Remove(key PdfObjectName) {
	idx := -1
	for i, k := range d.keys {
		if k == key {
			idx = i
			break
		}
	}

	if idx >= 0 {
		// Found. Remove from key list and map.
		d.keys = append(d.keys[:idx], d.keys[idx+1:]...)
		delete(d.dict, key)
	}
}

// SetIfNotNil sets the dictionary's key -> val mapping entry -IF- val is not nil.
// Note that we take care to perform a type switch.  Otherwise if we would supply a nil value
// of another type, e.g. (PdfObjectArray*)(nil), then it would not be a PdfObject(nil) and thus
// would get set.
//
func (d *PdfObjectDictionary) SetIfNotNil(key PdfObjectName, val PdfObject) {
	if val != nil {
		switch t := val.(type) {
		case *PdfObjectName:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectDictionary:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectStream:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectString:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectNull:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectInteger:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectArray:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectBool:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectFloat:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfObjectReference:
			if t != nil {
				d.Set(key, val)
			}
		case *PdfIndirectObject:
			if t != nil {
				d.Set(key, val)
			}
		default:
			common.Log.Error("ERROR: Unknown type: %T - should never happen!", val)
		}
	}
}

func (ref *PdfObjectReference) String() string {
	return fmt.Sprintf("Ref(%d %d)", ref.ObjectNumber, ref.GenerationNumber)
}

// DefaultWriteString outputs the object as it is to be written to file.
func (ref *PdfObjectReference) DefaultWriteString() string {
	return fmt.Sprintf("%d %d R", ref.ObjectNumber, ref.GenerationNumber)
}

func (ind *PdfIndirectObject) String() string {
	// Avoid printing out the object, can cause problems with circular
	// references.
	return fmt.Sprintf("IObject:%d", (*ind).ObjectNumber)
}

// DefaultWriteString outputs the object as it is to be written to file.
func (ind *PdfIndirectObject) DefaultWriteString() string {
	outStr := fmt.Sprintf("%d 0 R", (*ind).ObjectNumber)
	return outStr
}

func (stream *PdfObjectStream) String() string {
	return fmt.Sprintf("Object stream %d: %s", stream.ObjectNumber, stream.PdfObjectDictionary)
}

// DefaultWriteString outputs the object as it is to be written to file.
func (stream *PdfObjectStream) DefaultWriteString() string {
	outStr := fmt.Sprintf("%d 0 R", (*stream).ObjectNumber)
	return outStr
}

func (null *PdfObjectNull) String() string {
	return "null"
}

// DefaultWriteString outputs the object as it is to be written to file.
func (null *PdfObjectNull) DefaultWriteString() string {
	return "null"
}

// Handy functions to work with primitive objects.

// TraceMaxDepth specifies the maximum recursion depth allowed.
const TraceMaxDepth = 20

// TraceToDirectObject traces a PdfObject to a direct object.  For example direct objects contained
// in indirect objects (can be double referenced even).
//
// Note: This function does not trace/resolve references. That needs to be done beforehand.
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
