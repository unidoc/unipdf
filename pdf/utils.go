/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

package pdf

import (
	"sort"

	"github.com/unidoc/unidoc/common"
)

func (this *PdfReader) Inspect() {
	this.parser.inspect()
}

var log = common.GetLogger()

func getUniDocVersion() string {
	return common.Version
}

/*
 * Inspect object types.
 * Go through all objects in the cross ref table and detect the types.
 */
func (this *PdfParser) inspect() {
	log.Debug("--------INSPECT ----------")
	log.Debug("Xref table:")

	objTypes := map[string]int{}
	objCount := 0
	failedCount := 0

	keys := []int{}
	for k, _ := range this.xrefs {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	i := 0
	for _, k := range keys {
		xref := this.xrefs[k]
		if xref.objectNumber == 0 {
			continue
		}
		objCount++
		log.Debug("==========")
		log.Debug("Looking up object number: %d", xref.objectNumber)
		o, err := this.LookupByNumber(xref.objectNumber)
		if err != nil {
			log.Error("Fail to lookup obj %d (%s)", xref.objectNumber, err)
			failedCount++
			continue
		}

		log.Debug("obj: %s", o)

		iobj, isIndirect := o.(*PdfIndirectObject)
		if isIndirect {
			log.Debug("IND OOBJ %d: %s", xref.objectNumber, iobj)
			dict, isDict := iobj.PdfObject.(*PdfObjectDictionary)
			if isDict {
				ot, ok := (*dict)["Type"].(*PdfObjectName)

				if ok {
					otype := string(*ot)
					log.Debug("---> Obj type: %s", otype)
					_, isDefined := objTypes[otype]
					if isDefined {
						objTypes[otype]++
					} else {
						objTypes[otype] = 1
					}
				}
			}
		} else if sobj, isStream := o.(*PdfObjectStream); isStream {
			if otype, ok := (*(sobj.PdfObjectDictionary))["Type"].(*PdfObjectName); ok {
				log.Debug("--> Stream object type: %s", *otype)
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
				ot, isName := (*dict)["Type"].(*PdfObjectName)
				if isName {
					otype := string(*ot)
					log.Debug("AAA obj type %s", otype)
					objTypes[otype]++
				}
			}
			log.Debug("DIR OOBJ %d: %s", xref.objectNumber, o)
		}

		i++
	}
	log.Debug("--------EOF INSPECT ----------")
	log.Debug("=======")
	log.Debug("Object count: %d", objCount)
	log.Debug("Failed lookup: %d", failedCount)
	for t, c := range objTypes {
		log.Debug("%s: %d", t, c)
	}
	log.Debug("=======")

	if len(this.xrefs) < 1 {
		log.Error("This document is invalid (xref table missing!)")
		return
	}
	fontObjs, ok := objTypes["Font"]
	if !ok || fontObjs < 2 {
		log.Debug("This document is probably scanned!")
	} else {
		log.Debug("This document is valid for extraction!")
	}
}
