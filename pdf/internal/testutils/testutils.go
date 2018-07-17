/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package testutils provides test methods that are not intended to be exported.
package testutils

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

func CompareDictionariesDeep(d1, d2 *core.PdfObjectDictionary) bool {
	if len(d1.Keys()) != len(d2.Keys()) {
		common.Log.Debug("Dict entries mismatch (%d != %d)", len(d1.Keys()), len(d2.Keys()))
		common.Log.Debug("Was '%s' vs '%s'", d1.DefaultWriteString(), d2.DefaultWriteString())
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
					if v1.DefaultWriteString() != v2.DefaultWriteString() {
						common.Log.Debug("Mismatch '%s' != '%s'", v1.DefaultWriteString(), v2.DefaultWriteString())
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
