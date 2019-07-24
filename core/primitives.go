/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/strutils"
)

// PdfObject is an interface which all primitive PDF objects must implement.
type PdfObject interface {
	// String outputs a string representation of the primitive (for debugging).
	String() string

	// WriteString outputs the PDF primitive as written to file as expected by the standard.
	// TODO(dennwc): it should return a byte slice, or accept a writer
	WriteString() string
}

// PdfObjectBool represents the primitive PDF boolean object.
type PdfObjectBool bool

// PdfObjectInteger represents the primitive PDF integer numerical object.
type PdfObjectInteger int64

// PdfObjectFloat represents the primitive PDF floating point numerical object.
type PdfObjectFloat float64

// PdfObjectString represents the primitive PDF string object.
type PdfObjectString struct {
	val   string
	isHex bool
}

// PdfObjectName represents the primitive PDF name object.
type PdfObjectName string

// PdfObjectArray represents the primitive PDF array object.
type PdfObjectArray struct {
	vec []PdfObject
}

// PdfObjectDictionary represents the primitive PDF dictionary/map object.
type PdfObjectDictionary struct {
	dict map[PdfObjectName]PdfObject
	keys []PdfObjectName

	// For lazy-loading, need access to the parser (and the cross reference table for object access).
	parser *PdfParser
}

// PdfObjectNull represents the primitive PDF null object.
type PdfObjectNull struct{}

// PdfObjectReference represents the primitive PDF reference object.
type PdfObjectReference struct {
	// For PdfAppender, need access to the parser (and the cross reference table for object access).
	parser           *PdfParser
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

// PdfObjectStreams represents the primitive PDF object streams.
// 7.5.7 Object Streams (page 45).
type PdfObjectStreams struct {
	PdfObjectReference
	vec []PdfObject
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

// MakeBool creates a PdfObjectBool from a bool value.
func MakeBool(val bool) *PdfObjectBool {
	bval := PdfObjectBool(val)
	return &bval
}

// MakeArray creates an PdfObjectArray from a list of PdfObjects.
func MakeArray(objects ...PdfObject) *PdfObjectArray {
	array := &PdfObjectArray{}
	array.vec = []PdfObject{}
	for _, obj := range objects {
		array.vec = append(array.vec, obj)
	}
	return array
}

// MakeArrayFromIntegers creates an PdfObjectArray from a slice of ints, where each array element is
// an PdfObjectInteger.
func MakeArrayFromIntegers(vals []int) *PdfObjectArray {
	array := MakeArray()
	for _, val := range vals {
		array.Append(MakeInteger(int64(val)))
	}
	return array
}

// MakeArrayFromIntegers64 creates an PdfObjectArray from a slice of int64s, where each array element
// is an PdfObjectInteger.
func MakeArrayFromIntegers64(vals []int64) *PdfObjectArray {
	array := MakeArray()
	for _, val := range vals {
		array.Append(MakeInteger(val))
	}
	return array
}

// MakeArrayFromFloats creates an PdfObjectArray from a slice of float64s, where each array element is an
// PdfObjectFloat.
func MakeArrayFromFloats(vals []float64) *PdfObjectArray {
	array := MakeArray()
	for _, val := range vals {
		array.Append(MakeFloat(val))
	}
	return array
}

// MakeFloat creates an PdfObjectFloat from a float64.
func MakeFloat(val float64) *PdfObjectFloat {
	num := PdfObjectFloat(val)
	return &num
}

// MakeString creates an PdfObjectString from a string.
// NOTE: PDF does not use utf-8 string encoding like Go so `s` will often not be a utf-8 encoded
// string.
func MakeString(s string) *PdfObjectString {
	str := PdfObjectString{val: s}
	return &str
}

// MakeStringFromBytes creates an PdfObjectString from a byte array.
// This is more natural than MakeString as `data` is usually not utf-8 encoded.
func MakeStringFromBytes(data []byte) *PdfObjectString {
	return MakeString(string(data))
}

// MakeHexString creates an PdfObjectString from a string intended for output as a hexadecimal string.
func MakeHexString(s string) *PdfObjectString {
	str := PdfObjectString{val: s, isHex: true}
	return &str
}

// MakeEncodedString creates a PdfObjectString with encoded content, which can be either
// UTF-16BE or PDFDocEncoding depending on whether `utf16BE` is true or false respectively.
func MakeEncodedString(s string, utf16BE bool) *PdfObjectString {
	if utf16BE {
		var buf bytes.Buffer
		buf.Write([]byte{0xFE, 0xFF})
		buf.WriteString(strutils.StringToUTF16(s))
		return &PdfObjectString{val: buf.String(), isHex: true}
	}

	return &PdfObjectString{val: string(strutils.StringToPDFDocEncoding(s)), isHex: false}
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

// MakeObjectStreams creates an PdfObjectStreams from a list of PdfObjects.
func MakeObjectStreams(objects ...PdfObject) *PdfObjectStreams {
	streams := &PdfObjectStreams{}
	streams.vec = []PdfObject{}
	for _, obj := range objects {
		streams.vec = append(streams.vec, obj)
	}
	return streams
}

// GetParser returns the parser for lazy-loading or compare references.
func (ref *PdfObjectReference) GetParser() *PdfParser {
	return ref.parser
}

// Resolve resolves the reference and returns the indirect or stream object.
// If the reference cannot be resolved, a *PdfObjectNull object is returned.
func (ref *PdfObjectReference) Resolve() PdfObject {
	if ref.parser == nil {
		return MakeNull()
	}
	obj, _, err := ref.parser.resolveReference(ref)
	if err != nil {
		common.Log.Debug("ERROR resolving reference: %v - returning null object", err)
		return MakeNull()
	}
	if obj == nil {
		common.Log.Debug("ERROR resolving reference: nil object - returning a null object")
		return MakeNull()
	}
	return obj
}

// String returns the state of the bool as "true" or "false".
func (bool *PdfObjectBool) String() string {
	if *bool {
		return "true"
	}
	return "false"
}

// WriteString outputs the object as it is to be written to file.
func (bool *PdfObjectBool) WriteString() string {
	if *bool {
		return "true"
	}
	return "false"
}

func (int *PdfObjectInteger) String() string {
	return fmt.Sprintf("%d", *int)
}

// WriteString outputs the object as it is to be written to file.
func (int *PdfObjectInteger) WriteString() string {
	return strconv.FormatInt(int64(*int), 10)
}

func (float *PdfObjectFloat) String() string {
	return fmt.Sprintf("%f", *float)
}

// WriteString outputs the object as it is to be written to file.
func (float *PdfObjectFloat) WriteString() string {
	return strconv.FormatFloat(float64(*float), 'f', -1, 64)
}

// String returns a string representation of the *PdfObjectString.
func (str *PdfObjectString) String() string {
	return str.val
}

// Str returns the string value of the PdfObjectString. Defined in addition to String() function to clarify that
// this function returns the underlying string directly, whereas the String function technically could include
// debug info.
func (str *PdfObjectString) Str() string {
	return str.val
}

// Decoded returns the PDFDocEncoding or UTF-16BE decoded string contents.
// UTF-16BE is applied when the first two bytes are 0xFE, 0XFF, otherwise decoding of
// PDFDocEncoding is performed.
func (str *PdfObjectString) Decoded() string {
	if str == nil {
		return ""
	}
	b := []byte(str.val)
	if len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF {
		// UTF16BE.
		return strutils.UTF16ToString(b[2:])
	}

	return strutils.PDFDocEncodingToString(b)
}

// Bytes returns the PdfObjectString content as a []byte array.
func (str *PdfObjectString) Bytes() []byte {
	return []byte(str.val)
}

// WriteString outputs the object as it is to be written to file.
func (str *PdfObjectString) WriteString() string {
	var output bytes.Buffer

	// Handle hex representation.
	if str.isHex {
		shex := hex.EncodeToString(str.Bytes())
		output.WriteString("<")
		output.WriteString(shex)
		output.WriteString(">")
		return output.String()
	}

	// Otherwise regular string.
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
	for i := 0; i < len(str.val); i++ {
		char := str.val[i]
		if escStr, useEsc := escapeSequences[char]; useEsc {
			output.WriteString(escStr)
		} else {
			output.WriteByte(char)
		}
	}
	output.WriteString(")")
	return output.String()
}

// String returns a string representation of `name`.
func (name *PdfObjectName) String() string {
	return string(*name)
}

// WriteString outputs the object as it is to be written to file.
func (name *PdfObjectName) WriteString() string {
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

// Elements returns a slice of the PdfObject elements in the array.
func (array *PdfObjectArray) Elements() []PdfObject {
	if array == nil {
		return nil
	}
	return array.vec
}

// Len returns the number of elements in the array.
func (array *PdfObjectArray) Len() int {
	if array == nil {
		return 0
	}
	return len(array.vec)
}

// Get returns the i-th element of the array or nil if out of bounds (by index).
func (array *PdfObjectArray) Get(i int) PdfObject {
	if array == nil || i >= len(array.vec) || i < 0 {
		return nil
	}
	return array.vec[i]
}

// Set sets the PdfObject at index i of the array. An error is returned if the index is outside bounds.
func (array *PdfObjectArray) Set(i int, obj PdfObject) error {
	if i < 0 || i >= len(array.vec) {
		return errors.New("outside bounds")
	}
	array.vec[i] = obj
	return nil
}

// Append appends PdfObject(s) to the array.
func (array *PdfObjectArray) Append(objects ...PdfObject) {
	if array == nil {
		common.Log.Debug("Warn - Attempt to append to a nil array")
		return
	}
	if array.vec == nil {
		array.vec = []PdfObject{}
	}

	for _, obj := range objects {
		array.vec = append(array.vec, obj)
	}
}

// Clear resets the array to an empty state.
func (array *PdfObjectArray) Clear() {
	array.vec = []PdfObject{}
}

// ToFloat64Array returns a slice of all elements in the array as a float64 slice.  An error is
// returned if the array contains non-numeric objects (each element can be either PdfObjectInteger
// or PdfObjectFloat).
func (array *PdfObjectArray) ToFloat64Array() ([]float64, error) {
	var vals []float64

	for _, obj := range array.Elements() {
		switch t := obj.(type) {
		case *PdfObjectInteger:
			vals = append(vals, float64(*t))
		case *PdfObjectFloat:
			vals = append(vals, float64(*t))
		default:
			return nil, ErrTypeError
		}
	}

	return vals, nil
}

// ToIntegerArray returns a slice of all array elements as an int slice. An error is returned if the
// array non-integer objects. Each element can only be PdfObjectInteger.
func (array *PdfObjectArray) ToIntegerArray() ([]int, error) {
	var vals []int

	for _, obj := range array.Elements() {
		if number, is := obj.(*PdfObjectInteger); is {
			vals = append(vals, int(*number))
		} else {
			return nil, ErrTypeError
		}
	}

	return vals, nil
}

// ToInt64Slice returns a slice of all array elements as an int64 slice. An error is returned if the
// array non-integer objects. Each element can only be PdfObjectInteger.
func (array *PdfObjectArray) ToInt64Slice() ([]int64, error) {
	var vals []int64

	for _, obj := range array.Elements() {
		if number, is := obj.(*PdfObjectInteger); is {
			vals = append(vals, int64(*number))
		} else {
			return nil, ErrTypeError
		}
	}

	return vals, nil
}

// String returns a string describing `array`.
func (array *PdfObjectArray) String() string {
	outStr := "["
	for ind, o := range array.Elements() {
		outStr += o.String()
		if ind < (array.Len() - 1) {
			outStr += ", "
		}
	}
	outStr += "]"
	return outStr
}

// WriteString outputs the object as it is to be written to file.
func (array *PdfObjectArray) WriteString() string {
	var b strings.Builder
	b.WriteString("[")

	for ind, o := range array.Elements() {
		b.WriteString(o.WriteString())
		if ind < (array.Len() - 1) {
			b.WriteString(" ")
		}
	}

	b.WriteString("]")
	return b.String()
}

// GetNumberAsFloat returns the contents of `obj` as a float if it is an integer or float, or an
// error if it isn't.
func GetNumberAsFloat(obj PdfObject) (float64, error) {
	switch t := obj.(type) {
	case *PdfObjectFloat:
		return float64(*t), nil
	case *PdfObjectInteger:
		return float64(*t), nil
	}
	return 0, ErrNotANumber
}

// IsNullObject returns true if `obj` is a PdfObjectNull.
func IsNullObject(obj PdfObject) bool {
	_, isNull := TraceToDirectObject(obj).(*PdfObjectNull)
	return isNull
}

// GetNumbersAsFloat converts a list of pdf objects representing floats or integers to a slice of
// float64 values.
func GetNumbersAsFloat(objects []PdfObject) (floats []float64, err error) {
	for _, obj := range objects {
		val, err := GetNumberAsFloat(obj)
		if err != nil {
			return nil, err
		}
		floats = append(floats, val)
	}
	return floats, nil
}

// GetNumberAsInt64 returns the contents of `obj` as an int64 if it is an integer or float, or an
// error if it isn't. This is for cases where expecting an integer, but some implementations
// actually store the number in a floating point format.
func GetNumberAsInt64(obj PdfObject) (int64, error) {
	switch t := obj.(type) {
	case *PdfObjectFloat:
		common.Log.Debug("Number expected as integer was stored as float (type casting used)")
		return int64(*t), nil
	case *PdfObjectInteger:
		return int64(*t), nil
	}
	return 0, ErrNotANumber
}

// getNumberAsFloatOrNull returns the contents of `obj` as a *float if it is an integer or float,
// or nil if it `obj` is nil. In other cases an error is returned.
func getNumberAsFloatOrNull(obj PdfObject) (*float64, error) {
	switch t := obj.(type) {
	case *PdfObjectFloat:
		val := float64(*t)
		return &val, nil
	case *PdfObjectInteger:
		val := float64(*t)
		return &val, nil
	case *PdfObjectNull:
		return nil, nil
	}
	return nil, ErrNotANumber
}

// GetAsFloat64Slice returns the array as []float64 slice.
// Returns an error if not entirely numeric (only PdfObjectIntegers, PdfObjectFloats).
func (array *PdfObjectArray) GetAsFloat64Slice() ([]float64, error) {
	var slice []float64

	for _, obj := range array.Elements() {
		number, err := GetNumberAsFloat(TraceToDirectObject(obj))
		if err != nil {
			return nil, fmt.Errorf("array element not a number")
		}
		slice = append(slice, number)
	}

	return slice, nil
}

// Merge merges in key/values from another dictionary. Overwriting if has same keys.
// The mutated dictionary (d) is returned in order to allow method chaining.
func (d *PdfObjectDictionary) Merge(another *PdfObjectDictionary) *PdfObjectDictionary {
	if another != nil {
		for _, key := range another.Keys() {
			val := another.Get(key)
			d.Set(key, val)
		}
	}

	return d
}

// String returns a string describing `d`.
func (d *PdfObjectDictionary) String() string {
	var b strings.Builder
	b.WriteString("Dict(")
	for _, k := range d.keys {
		v := d.dict[k]
		b.WriteString(`"` + k.String() + `": `)
		b.WriteString(v.String())
		b.WriteString(`, `)
	}
	b.WriteString(")")
	return b.String()
}

// WriteString outputs the object as it is to be written to file.
func (d *PdfObjectDictionary) WriteString() string {
	var b strings.Builder

	b.WriteString("<<")
	for _, k := range d.keys {
		v := d.dict[k]
		b.WriteString(k.WriteString())
		b.WriteString(" ")
		b.WriteString(v.WriteString())
	}

	b.WriteString(">>")
	return b.String()
}

// Set sets the dictionary's key -> val mapping entry. Overwrites if key already set.
func (d *PdfObjectDictionary) Set(key PdfObjectName, val PdfObject) {
	_, found := d.dict[key]
	if !found {
		d.keys = append(d.keys, key)
	}
	d.dict[key] = val
}

// Get returns the PdfObject corresponding to the specified key.
// Returns a nil value if the key is not set.
func (d *PdfObjectDictionary) Get(key PdfObjectName) PdfObject {
	val, has := d.dict[key]
	if !has {
		return nil
	}
	return val
}

// GetString is a helper for Get that returns a string value.
// Returns false if the key is missing or a value is not a string.
func (d *PdfObjectDictionary) GetString(key PdfObjectName) (string, bool) {
	val, ok := d.dict[key].(*PdfObjectString)
	if !ok {
		return "", false
	}
	return val.Str(), true
}

// Keys returns the list of keys in the dictionary.
// If `d` is nil returns a nil slice.
func (d *PdfObjectDictionary) Keys() []PdfObjectName {
	if d == nil {
		return nil
	}
	return d.keys
}

// Clear resets the dictionary to an empty state.
func (d *PdfObjectDictionary) Clear() {
	d.keys = []PdfObjectName{}
	d.dict = map[PdfObjectName]PdfObject{}
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

// String returns a string describing `ref`.
func (ref *PdfObjectReference) String() string {
	return fmt.Sprintf("Ref(%d %d)", ref.ObjectNumber, ref.GenerationNumber)
}

// WriteString outputs the object as it is to be written to file.
func (ref *PdfObjectReference) WriteString() string {
	var b strings.Builder
	b.WriteString(strconv.FormatInt(ref.ObjectNumber, 10))
	b.WriteString(" ")
	b.WriteString(strconv.FormatInt(ref.GenerationNumber, 10))
	b.WriteString(" R")
	return b.String()
}

// String returns a string describing `ind`.
func (ind *PdfIndirectObject) String() string {
	// Avoid printing out the object, can cause problems with circular
	// references.
	return fmt.Sprintf("IObject:%d", (*ind).ObjectNumber)
}

// WriteString outputs the object as it is to be written to file.
func (ind *PdfIndirectObject) WriteString() string {
	var b strings.Builder
	b.WriteString(strconv.FormatInt(ind.ObjectNumber, 10))
	b.WriteString(" 0 R")
	return b.String()
}

// String returns a string describing `stream`.
func (stream *PdfObjectStream) String() string {
	return fmt.Sprintf("Object stream %d: %s", stream.ObjectNumber, stream.PdfObjectDictionary)
}

// WriteString outputs the object as it is to be written to file.
func (stream *PdfObjectStream) WriteString() string {
	var b strings.Builder
	b.WriteString(strconv.FormatInt(stream.ObjectNumber, 10))
	b.WriteString(" 0 R")
	return b.String()
}

// String returns a string describing `null`.
func (null *PdfObjectNull) String() string {
	return "null"
}

// WriteString outputs the object as it is to be written to file.
func (null *PdfObjectNull) WriteString() string {
	return "null"
}

// Handy functions to work with primitive objects.

// TraceMaxDepth specifies the maximum recursion depth allowed.
const traceMaxDepth = 10

// TraceToDirectObject traces a PdfObject to a direct object.  For example direct objects contained
// in indirect objects (can be double referenced even).
func TraceToDirectObject(obj PdfObject) PdfObject {
	if ref, isRef := obj.(*PdfObjectReference); isRef {
		obj = ref.Resolve()
	}

	iobj, isIndirectObj := obj.(*PdfIndirectObject)
	depth := 0
	for isIndirectObj {
		obj = iobj.PdfObject
		iobj, isIndirectObj = GetIndirect(obj)
		depth++
		if depth > traceMaxDepth {
			common.Log.Error("ERROR: Trace depth level beyond %d - not going deeper!", traceMaxDepth)
			return nil
		}
	}
	return obj
}

// Convenience methods for converting PdfObject to underlying types.

// GetBool returns the *PdfObjectBool object that is represented by a PdfObject directly or indirectly
// within an indirect object. The bool flag indicates whether a match was found.
func GetBool(obj PdfObject) (bo *PdfObjectBool, found bool) {
	bo, found = TraceToDirectObject(obj).(*PdfObjectBool)
	return bo, found
}

// GetBoolVal returns the bool value within a *PdObjectBool represented by an PdfObject interface directly or indirectly.
// If the PdfObject does not represent a bool value, a default value of false is returned (found = false also).
func GetBoolVal(obj PdfObject) (b bool, found bool) {
	bo, found := TraceToDirectObject(obj).(*PdfObjectBool)
	if found {
		return bool(*bo), true
	}
	return false, false
}

// GetInt returns the *PdfObjectBool object that is represented by a PdfObject either directly or indirectly
// within an indirect object. The bool flag indicates whether a match was found.
func GetInt(obj PdfObject) (into *PdfObjectInteger, found bool) {
	into, found = TraceToDirectObject(obj).(*PdfObjectInteger)
	return into, found
}

// GetIntVal returns the int value represented by the PdfObject directly or indirectly if contained within an
// indirect object. On type mismatch the found bool flag returned is false and a nil pointer is returned.
func GetIntVal(obj PdfObject) (val int, found bool) {
	into, found := TraceToDirectObject(obj).(*PdfObjectInteger)
	if found && into != nil {
		return int(*into), true
	}
	return 0, false
}

// GetFloat returns the *PdfObjectFloat represented by the PdfObject directly or indirectly within an indirect
// object. On type mismatch the found bool flag is false and a nil pointer is returned.
func GetFloat(obj PdfObject) (fo *PdfObjectFloat, found bool) {
	fo, found = TraceToDirectObject(obj).(*PdfObjectFloat)
	return fo, found
}

// GetFloatVal returns the float64 value represented by the PdfObject directly or indirectly if contained within an
// indirect object. On type mismatch the found bool flag returned is false and a nil pointer is returned.
func GetFloatVal(obj PdfObject) (val float64, found bool) {
	fo, found := TraceToDirectObject(obj).(*PdfObjectFloat)
	if found {
		return float64(*fo), true
	}
	return 0, false
}

// GetString returns the *PdfObjectString represented by the PdfObject directly or indirectly within an indirect
// object. On type mismatch the found bool flag is false and a nil pointer is returned.
func GetString(obj PdfObject) (so *PdfObjectString, found bool) {
	so, found = TraceToDirectObject(obj).(*PdfObjectString)
	return so, found
}

// GetStringVal returns the string value represented by the PdfObject directly or indirectly if
// contained within an indirect object. On type mismatch the found bool flag returned is false and
// an empty string is returned.
func GetStringVal(obj PdfObject) (val string, found bool) {
	so, found := TraceToDirectObject(obj).(*PdfObjectString)
	if found {
		return so.Str(), true
	}
	return
}

// GetStringBytes is like GetStringVal except that it returns the string as a []byte.
// It is for convenience.
func GetStringBytes(obj PdfObject) (val []byte, found bool) {
	so, found := TraceToDirectObject(obj).(*PdfObjectString)
	if found {
		return so.Bytes(), true
	}
	return
}

// GetName returns the *PdfObjectName represented by the PdfObject directly or indirectly within an indirect
// object. On type mismatch the found bool flag is false and a nil pointer is returned.
func GetName(obj PdfObject) (name *PdfObjectName, found bool) {
	name, found = TraceToDirectObject(obj).(*PdfObjectName)
	return name, found
}

// GetNameVal returns the string value represented by the PdfObject directly or indirectly if
// contained within an indirect object. On type mismatch the found bool flag returned is false and
// an empty string is returned.
func GetNameVal(obj PdfObject) (val string, found bool) {
	name, found := TraceToDirectObject(obj).(*PdfObjectName)
	if found {
		return string(*name), true
	}
	return
}

// GetArray returns the *PdfObjectArray represented by the PdfObject directly or indirectly within an indirect
// object. On type mismatch the found bool flag is false and a nil pointer is returned.
func GetArray(obj PdfObject) (arr *PdfObjectArray, found bool) {
	arr, found = TraceToDirectObject(obj).(*PdfObjectArray)
	return arr, found
}

// GetDict returns the *PdfObjectDictionary represented by the PdfObject directly or indirectly within an indirect
// object. On type mismatch the found bool flag is false and a nil pointer is returned.
func GetDict(obj PdfObject) (dict *PdfObjectDictionary, found bool) {
	dict, found = TraceToDirectObject(obj).(*PdfObjectDictionary)
	return dict, found
}

// GetIndirect returns the *PdfIndirectObject represented by the PdfObject. On type mismatch the found bool flag is
// false and a nil pointer is returned.
func GetIndirect(obj PdfObject) (ind *PdfIndirectObject, found bool) {
	obj = ResolveReference(obj)
	ind, found = obj.(*PdfIndirectObject)
	return ind, found
}

// GetStream returns the *PdfObjectStream represented by the PdfObject. On type mismatch the found bool flag is
// false and a nil pointer is returned.
func GetStream(obj PdfObject) (stream *PdfObjectStream, found bool) {
	obj = ResolveReference(obj)
	stream, found = obj.(*PdfObjectStream)
	return stream, found
}

// GetObjectStreams returns the *PdfObjectStreams represented by the PdfObject. On type mismatch the found bool flag is
// false and a nil pointer is returned.
func GetObjectStreams(obj PdfObject) (objStream *PdfObjectStreams, found bool) {
	objStream, found = obj.(*PdfObjectStreams)
	return objStream, found
}

// Append appends PdfObject(s) to the streams.
func (streams *PdfObjectStreams) Append(objects ...PdfObject) {
	if streams == nil {
		common.Log.Debug("Warn - Attempt to append to a nil streams")
		return
	}
	if streams.vec == nil {
		streams.vec = []PdfObject{}
	}

	for _, obj := range objects {
		streams.vec = append(streams.vec, obj)
	}
}

// Set sets the PdfObject at index i of the streams. An error is returned if the index is outside bounds.
func (streams *PdfObjectStreams) Set(i int, obj PdfObject) error {
	if i < 0 || i >= len(streams.vec) {
		return errors.New("Outside bounds")
	}
	streams.vec[i] = obj
	return nil
}

// Elements returns a slice of the PdfObject elements in the array.
// Preferred over accessing the array directly as type may be changed in future major versions (v3).
func (streams *PdfObjectStreams) Elements() []PdfObject {
	if streams == nil {
		return nil
	}
	return streams.vec
}

// String returns a string describing `streams`.
func (streams *PdfObjectStreams) String() string {
	return fmt.Sprintf("Object stream %d", streams.ObjectNumber)
}

// Len returns the number of elements in the streams.
func (streams *PdfObjectStreams) Len() int {
	if streams == nil {
		return 0
	}
	return len(streams.vec)
}

// WriteString outputs the object as it is to be written to file.
func (streams *PdfObjectStreams) WriteString() string {
	var b strings.Builder
	b.WriteString(strconv.FormatInt(streams.ObjectNumber, 10))
	b.WriteString(" 0 R")
	return b.String()
}
