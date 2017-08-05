/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"errors"
	"fmt"
	"sort"

	"github.com/unidoc/unidoc/common"
)

// Check slice range to make sure within bounds for accessing:
//    slice[a:b] where sliceLen=len(slice).
func checkBounds(sliceLen, a, b int) error {
	if a < 0 || a > sliceLen {
		return errors.New("Slice index a out of bounds")
	}
	if b < a {
		return errors.New("Invalid slice index b < a")
	}
	if b > sliceLen {
		return errors.New("Slice index b out of bounds")
	}

	return nil
}

// Inspect analyzes the document object structure.
func (parser *PdfParser) Inspect() (map[string]int, error) {
	return parser.inspect()
}

// GetObjectNums returns a sorted list of object numbers of the PDF objects in the file.
func (parser *PdfParser) GetObjectNums() []int {
	objNums := []int{}
	for _, x := range parser.xrefs {
		objNums = append(objNums, x.objectNumber)
	}

	// Sort the object numbers to give consistent ordering of PDF objects in output.
	// Needed since parser.xrefs is a map.
	sort.Ints(objNums)

	return objNums
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

	keys := []int{}
	for k := range parser.xrefs {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	i := 0
	for _, k := range keys {
		xref := parser.xrefs[k]
		if xref.objectNumber == 0 {
			continue
		}
		objCount++
		common.Log.Trace("==========")
		common.Log.Trace("Looking up object number: %d", xref.objectNumber)
		o, err := parser.LookupByNumber(xref.objectNumber)
		if err != nil {
			common.Log.Trace("ERROR: Fail to lookup obj %d (%s)", xref.objectNumber, err)
			failedCount++
			continue
		}

		common.Log.Trace("obj: %s", o)

		iobj, isIndirect := o.(*PdfIndirectObject)
		if isIndirect {
			common.Log.Trace("IND OOBJ %d: %s", xref.objectNumber, iobj)
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
			common.Log.Trace("DIRECT OBJ %d: %s", xref.objectNumber, o)
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

	if len(parser.xrefs) < 1 {
		common.Log.Debug("ERROR: This document is invalid (xref table missing!)")
		return nil, fmt.Errorf("Invalid document (xref table missing)")
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
	} else {
		return x
	}
}
