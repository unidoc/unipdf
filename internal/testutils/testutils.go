/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package testutils provides test methods that are not intended to be exported.
package testutils

import (
	"errors"
	"io"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// ParseIndirectObjects parses a sequence of indirect/stream objects sequentially from a `rawpdf` text.
// Great for testing.
func ParseIndirectObjects(rawpdf string) (map[int64]core.PdfObject, error) {
	p := core.NewParserFromString(rawpdf)

	objmap := map[int64]core.PdfObject{}
	for {
		obj, err := p.ParseIndirectObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := obj.(type) {
		case *core.PdfIndirectObject:
			objmap[t.ObjectNumber] = obj
		case *core.PdfObjectStream:
			objmap[t.ObjectNumber] = obj
		}
	}

	for _, obj := range objmap {
		resolveReferences(obj, objmap)
	}

	return objmap, nil
}

// resolveReferences traverses `obj` structure and replaces references with corresponding PdfObject from `objmap`.
func resolveReferences(obj core.PdfObject, objmap map[int64]core.PdfObject) error {
	switch t := obj.(type) {
	case *core.PdfIndirectObject:
		indobj := t
		resolveReferences(indobj.PdfObject, objmap)
	case *core.PdfObjectDictionary:
		dict := t
		for _, key := range dict.Keys() {
			val := dict.Get(key)
			if ref, isref := val.(*core.PdfObjectReference); isref {
				replace, ok := objmap[ref.ObjectNumber]
				if !ok {
					return errors.New("reference to outside object")
				}
				dict.Set(key, replace)
			} else {
				resolveReferences(val, objmap)
			}
		}
	case *core.PdfObjectArray:
		array := t
		for index, val := range array.Elements() {
			if ref, isref := val.(*core.PdfObjectReference); isref {
				replace, ok := objmap[ref.ObjectNumber]
				if !ok {
					return errors.New("reference to outside object")
				}
				array.Set(index, replace)
			} else {
				resolveReferences(val, objmap)
			}
		}
	}

	return nil
}

// CompareDictionariesDeep does a deep comparison of `d1` and `d2` and returns true if equal.
func CompareDictionariesDeep(d1, d2 *core.PdfObjectDictionary) bool {
	if len(d1.Keys()) != len(d2.Keys()) {
		common.Log.Debug("Dict entries mismatch (%d != %d)", len(d1.Keys()), len(d2.Keys()))
		common.Log.Debug("Was '%s' vs '%s'", d1.WriteString(), d2.WriteString())
		return false
	}

	for _, k := range d1.Keys() {
		if k == "Parent" {
			continue
		}
		v1 := core.TraceToDirectObject(d1.Get(k))
		v2 := core.TraceToDirectObject(d2.Get(k))

		if v1 == nil {
			common.Log.Debug("v1 is nil")
			return false
		}
		if v2 == nil {
			common.Log.Debug("v2 is nil")
			return false
		}

		switch t1 := v1.(type) {
		case *core.PdfObjectDictionary:
			t2, ok := v2.(*core.PdfObjectDictionary)
			if !ok {
				common.Log.Debug("Type mismatch %T vs %T", v1, v2)
				return false
			}
			if !CompareDictionariesDeep(t1, t2) {
				return false
			}
			continue

		case *core.PdfObjectArray:
			t2, ok := v2.(*core.PdfObjectArray)
			if !ok {
				common.Log.Debug("v2 not an array")
				return false
			}
			if t1.Len() != t2.Len() {
				common.Log.Debug("array length mismatch (%d != %d)", t1.Len(), t2.Len())
				return false
			}
			for i := 0; i < t1.Len(); i++ {
				v1 := core.TraceToDirectObject(t1.Get(i))
				v2 := core.TraceToDirectObject(t2.Get(i))
				if d1, isD1 := v1.(*core.PdfObjectDictionary); isD1 {
					d2, isD2 := v2.(*core.PdfObjectDictionary)
					if !isD2 {
						return false
					}
					if !CompareDictionariesDeep(d1, d2) {
						return false
					}
				} else {
					if v1.WriteString() != v2.WriteString() {
						common.Log.Debug("Mismatch '%s' != '%s'", v1.WriteString(), v2.WriteString())
						return false
					}
				}
			}
			continue
		}

		if v1.String() != v2.String() {
			common.Log.Debug("key=%s Mismatch! '%s' != '%s'", k, v1.String(), v2.String())
			common.Log.Debug("For '%T' - '%T'", v1, v2)
			common.Log.Debug("For '%+v' - '%+v'", v1, v2)
			return false
		}
	}

	return true
}
