/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
)

// Check slice range to make sure within bounds for accessing:
//    slice[a:b] where sliceLen=len(slice).
func checkBounds(sliceLen, a, b int) error {
	if a < 0 || a > sliceLen {
		return errors.New("slice index a out of bounds")
	}
	if b < a {
		return errors.New("invalid slice index b < a")
	}
	if b > sliceLen {
		return errors.New("slice index b out of bounds")
	}

	return nil
}

// Inspect analyzes the document object structure. Returns a map of object types (by name) with the instance count
// as value.
func (parser *PdfParser) Inspect() (map[string]int, error) {
	return parser.inspect()
}

// GetObjectNums returns a sorted list of object numbers of the PDF objects in the file.
func (parser *PdfParser) GetObjectNums() []int {
	var objNums []int
	for _, x := range parser.xrefs.ObjectMap {
		objNums = append(objNums, x.ObjectNumber)
	}

	// Sort the object numbers to give consistent ordering of PDF objects in output.
	// Needed since parser.xrefs is a map.
	sort.Ints(objNums)

	return objNums
}

// ResolveReference resolves reference if `o` is a *PdfObjectReference and returns the object referenced to.
// Otherwise returns back `o`.
func ResolveReference(obj PdfObject) PdfObject {
	if ref, isRef := obj.(*PdfObjectReference); isRef {
		return ref.Resolve()
	}
	return obj
}

// ResolveReferencesDeep recursively traverses through object `o`, looking up and replacing
// references with indirect objects.
// Optionally a map of already deep-resolved objects can be provided via `traversed`. The `traversed` map
// is updated while traversing the objects to avoid traversing same objects multiple times.
func ResolveReferencesDeep(o PdfObject, traversed map[PdfObject]struct{}) error {
	if traversed == nil {
		traversed = map[PdfObject]struct{}{}
	}
	return resolveReferencesDeep(o, 0, traversed)
}

func resolveReferencesDeep(o PdfObject, depth int, traversed map[PdfObject]struct{}) error {
	common.Log.Trace("Traverse object data (depth = %d)", depth)
	if _, isTraversed := traversed[o]; isTraversed {
		common.Log.Trace("-Already traversed...")
		return nil
	}
	traversed[o] = struct{}{}

	switch t := o.(type) {
	case *PdfIndirectObject:
		io := t
		common.Log.Trace("io: %s", io)
		common.Log.Trace("- %s", io.PdfObject)
		return resolveReferencesDeep(io.PdfObject, depth+1, traversed)
	case *PdfObjectStream:
		so := t
		return resolveReferencesDeep(so.PdfObjectDictionary, depth+1, traversed)
	case *PdfObjectDictionary:
		dict := t
		common.Log.Trace("- dict: %s", dict)
		for _, name := range dict.Keys() {
			v := dict.Get(name)
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj := ref.Resolve()
				dict.Set(name, resolvedObj)
				err := resolveReferencesDeep(resolvedObj, depth+1, traversed)
				if err != nil {
					return err
				}
			} else {
				err := resolveReferencesDeep(v, depth+1, traversed)
				if err != nil {
					return err
				}
			}
		}
		return nil
	case *PdfObjectArray:
		arr := t
		common.Log.Trace("- array: %s", arr)
		for idx, v := range arr.Elements() {
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj := ref.Resolve()
				arr.Set(idx, resolvedObj)
				err := resolveReferencesDeep(resolvedObj, depth+1, traversed)
				if err != nil {
					return err
				}
			} else {
				err := resolveReferencesDeep(v, depth+1, traversed)
				if err != nil {
					return err
				}
			}
		}
		return nil
	case *PdfObjectReference:
		common.Log.Debug("ERROR: Tracing a reference!")
		return errors.New("error tracing a reference")
	}

	return nil
}

func getUniDocVersion() string {
	return common.Version
}

/*
 * Inspect object types.
 * Go through all objects in the cross ref table and detect the types.
 * Mostly for debugging purposes and inspecting odd PDF files.
 */
func (parser *PdfParser) inspect() (map[string]int, error) {
	common.Log.Trace("--------INSPECT ----------")
	common.Log.Trace("Xref table:")

	objTypes := map[string]int{}
	objCount := 0
	failedCount := 0

	var keys []int
	for k := range parser.xrefs.ObjectMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	i := 0
	for _, k := range keys {
		xref := parser.xrefs.ObjectMap[k]
		if xref.ObjectNumber == 0 {
			continue
		}
		objCount++
		common.Log.Trace("==========")
		common.Log.Trace("Looking up object number: %d", xref.ObjectNumber)
		o, err := parser.LookupByNumber(xref.ObjectNumber)
		if err != nil {
			common.Log.Trace("ERROR: Fail to lookup obj %d (%s)", xref.ObjectNumber, err)
			failedCount++
			continue
		}

		common.Log.Trace("obj: %s", o)

		iobj, isIndirect := o.(*PdfIndirectObject)
		if isIndirect {
			common.Log.Trace("IND OOBJ %d: %s", xref.ObjectNumber, iobj)
			dict, isDict := iobj.PdfObject.(*PdfObjectDictionary)
			if isDict {
				// Check if has Type parameter.
				if ot, has := dict.Get("Type").(*PdfObjectName); has {
					otype := string(*ot)
					common.Log.Trace("---> Obj type: %s", otype)
					_, isDefined := objTypes[otype]
					if isDefined {
						objTypes[otype]++
					} else {
						objTypes[otype] = 1
					}
				} else if ot, has := dict.Get("Subtype").(*PdfObjectName); has {
					// Check if subtype
					otype := string(*ot)
					common.Log.Trace("---> Obj subtype: %s", otype)
					_, isDefined := objTypes[otype]
					if isDefined {
						objTypes[otype]++
					} else {
						objTypes[otype] = 1
					}
				}
				if val, has := dict.Get("S").(*PdfObjectName); has && *val == "JavaScript" {
					// Check if Javascript.
					_, isDefined := objTypes["JavaScript"]
					if isDefined {
						objTypes["JavaScript"]++
					} else {
						objTypes["JavaScript"] = 1
					}
				}

			}
		} else if sobj, isStream := o.(*PdfObjectStream); isStream {
			if otype, ok := sobj.PdfObjectDictionary.Get("Type").(*PdfObjectName); ok {
				common.Log.Trace("--> Stream object type: %s", *otype)
				k := string(*otype)
				if _, isDefined := objTypes[k]; isDefined {
					objTypes[k]++
				} else {
					objTypes[k] = 1
				}
			}
		} else { // Direct.
			dict, isDict := o.(*PdfObjectDictionary)
			if isDict {
				ot, isName := dict.Get("Type").(*PdfObjectName)
				if isName {
					otype := string(*ot)
					common.Log.Trace("--- obj type %s", otype)
					objTypes[otype]++
				}
			}
			common.Log.Trace("DIRECT OBJ %d: %s", xref.ObjectNumber, o)
		}

		i++
	}
	common.Log.Trace("--------EOF INSPECT ----------")
	common.Log.Trace("=======")
	common.Log.Trace("Object count: %d", objCount)
	common.Log.Trace("Failed lookup: %d", failedCount)
	for t, c := range objTypes {
		common.Log.Trace("%s: %d", t, c)
	}
	common.Log.Trace("=======")

	if len(parser.xrefs.ObjectMap) < 1 {
		common.Log.Debug("ERROR: This document is invalid (xref table missing!)")
		return nil, fmt.Errorf("invalid document (xref table missing)")
	}

	fontObjs, ok := objTypes["Font"]
	if !ok || fontObjs < 2 {
		common.Log.Trace("This document is probably scanned!")
	} else {
		common.Log.Trace("This document is valid for extraction!")
	}

	return objTypes, nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// EqualObjects returns true if `obj1` and `obj2` have the same contents.
//
// NOTE: It is a good idea to flatten obj1 and obj2 with FlattenObject before calling this function
// so that contents, rather than references, can be compared.
func EqualObjects(obj1, obj2 PdfObject) bool {
	return equalObjects(obj1, obj2, 0)
}

// equalObjects returns true if `obj1` and `obj2` have the same contents.
// It recursively checks the contents of indirect objects, arrays and dicts to a depth of
// TraceMaxDepth. `depth` is the current recusion depth.
func equalObjects(obj1, obj2 PdfObject, depth int) bool {
	if depth > traceMaxDepth {
		common.Log.Error("Trace depth level beyond %d - error!", traceMaxDepth)
		return false
	}

	if obj1 == nil && obj2 == nil {
		return true
	} else if obj1 == nil || obj2 == nil {
		return false
	}
	if reflect.TypeOf(obj1) != reflect.TypeOf(obj2) {
		return false
	}

	// obj1 and obj2 are non-nil and of the same type
	switch t1 := obj1.(type) {
	case *PdfObjectNull, *PdfObjectReference:
		return true
	case *PdfObjectName:
		return *t1 == *(obj2.(*PdfObjectName))
	case *PdfObjectString:
		return *t1 == *(obj2.(*PdfObjectString))
	case *PdfObjectInteger:
		return *t1 == *(obj2.(*PdfObjectInteger))
	case *PdfObjectBool:
		return *t1 == *(obj2.(*PdfObjectBool))
	case *PdfObjectFloat:
		return *t1 == *(obj2.(*PdfObjectFloat))
	case *PdfIndirectObject:
		return equalObjects(TraceToDirectObject(obj1), TraceToDirectObject(obj2), depth+1)
	case *PdfObjectArray:
		t2 := obj2.(*PdfObjectArray)
		if len((*t1).vec) != len((*t2).vec) {
			return false
		}
		for i, o1 := range (*t1).vec {
			if !equalObjects(o1, (*t2).vec[i], depth+1) {
				return false
			}
		}
		return true
	case *PdfObjectDictionary:
		t2 := obj2.(*PdfObjectDictionary)
		d1, d2 := (*t1).dict, (*t2).dict
		if len(d1) != len(d2) {
			return false
		}
		for k, o1 := range d1 {
			o2, ok := d2[k]
			if !ok || !equalObjects(o1, o2, depth+1) {
				return false
			}
		}
		return true
	case *PdfObjectStream:
		t2 := obj2.(*PdfObjectStream)
		return equalObjects((*t1).PdfObjectDictionary, (*t2).PdfObjectDictionary, depth+1)
	default:
		common.Log.Error("ERROR: Unknown type: %T - should never happen!", obj1)
	}

	return false
}

// FlattenObject returns the contents of `obj`. In other words, `obj` with indirect objects replaced
// by their values.
// The replacements are made recursively to a depth of traceMaxDepth.
// NOTE: Dicts are sorted to make objects with same contents have the same PDF object strings.
func FlattenObject(obj PdfObject) PdfObject {
	return flattenObject(obj, 0)
}

// flattenObject returns `obj` with indirect objects recursively replaced by their values.
// `depth` is the recursion depth.
func flattenObject(obj PdfObject, depth int) PdfObject {
	if depth > traceMaxDepth {
		common.Log.Error("Trace depth level beyond %d - error!", traceMaxDepth)
		return MakeNull()
	}
	switch t := obj.(type) {
	case *PdfIndirectObject:
		obj = flattenObject((*t).PdfObject, depth+1)
	case *PdfObjectArray:
		for i, o := range (*t).vec {
			(*t).vec[i] = flattenObject(o, depth+1)
		}
	case *PdfObjectDictionary:
		for k, o := range (*t).dict {
			(*t).dict[k] = flattenObject(o, depth+1)
		}
		sort.Slice((*t).keys, func(i, j int) bool { return (*t).keys[i] < (*t).keys[j] })
	}
	return obj
}

// ParseNumber parses a numeric objects from a buffered stream.
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
func ParseNumber(buf *bufio.Reader) (PdfObject, error) {
	isFloat := false
	allowSigns := true
	var r bytes.Buffer
	for {
		if common.Log.IsLogLevel(common.LogLevelTrace) {
			common.Log.Trace("Parsing number \"%s\"", r.String())
		}
		bb, err := buf.Peek(1)
		if err == io.EOF {
			// GH: EOF handling.  Handle EOF like end of line.  Can happen with
			// encoded object streams that the object is at the end.
			// In other cases, we will get the EOF error elsewhere at any rate.
			break // Handle like EOF
		}
		if err != nil {
			common.Log.Debug("ERROR %s", err)
			return nil, err
		}
		if allowSigns && (bb[0] == '-' || bb[0] == '+') {
			// Only appear in the beginning, otherwise serves as a delimiter.
			b, _ := buf.ReadByte()
			r.WriteByte(b)
			allowSigns = false // Only allowed in beginning, and after e (exponential).
		} else if IsDecimalDigit(bb[0]) {
			b, _ := buf.ReadByte()
			r.WriteByte(b)
		} else if bb[0] == '.' {
			b, _ := buf.ReadByte()
			r.WriteByte(b)
			isFloat = true
		} else if bb[0] == 'e' || bb[0] == 'E' {
			// Exponential number format.
			b, _ := buf.ReadByte()
			r.WriteByte(b)
			isFloat = true
			allowSigns = true
		} else {
			break
		}
	}

	var o PdfObject
	if isFloat {
		fVal, err := strconv.ParseFloat(r.String(), 64)
		if err != nil {
			common.Log.Debug("Error parsing number %v err=%v. Using 0.0. Output may be incorrect", r.String(), err)
			fVal = 0.0
			err = nil
		}

		objFloat := PdfObjectFloat(fVal)
		o = &objFloat
	} else {
		intVal, err := strconv.ParseInt(r.String(), 10, 64)
		if err != nil {
			common.Log.Debug("Error parsing number %v err=%v. Using 0. Output may be incorrect", r.String(), err)
			intVal = 0
			err = nil
		}

		objInt := PdfObjectInteger(intVal)
		o = &objInt
	}

	return o, nil
}
