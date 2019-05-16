/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"crypto/md5"

	"github.com/unidoc/unipdf/v3/core"
)

// CombineIdenticalIndirectObjects combines identical indirect objects.
// It implements interface model.Optimizer.
type CombineIdenticalIndirectObjects struct {
}

// Optimize optimizes PDF objects to decrease PDF size.
func (c *CombineIdenticalIndirectObjects) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	updateObjectNumbers(objects)
	replaceTable := make(map[core.PdfObject]core.PdfObject)
	toDelete := make(map[core.PdfObject]struct{})

	indWithDictByHash := make(map[string][]*core.PdfIndirectObject)

	for _, obj := range objects {
		ind, isIndirectObj := obj.(*core.PdfIndirectObject)
		if !isIndirectObj {
			continue
		}
		if dict, isDictObj := ind.PdfObject.(*core.PdfObjectDictionary); isDictObj {
			if name, isName := dict.Get("Type").(*core.PdfObjectName); isName && *name == "Page" {
				continue
			}
			hasher := md5.New()
			hasher.Write([]byte(dict.WriteString()))

			hash := string(hasher.Sum(nil))
			indWithDictByHash[hash] = append(indWithDictByHash[hash], ind)
		}
	}

	for _, dicts := range indWithDictByHash {
		if len(dicts) < 2 {
			continue
		}
		firstDict := dicts[0]
		for i := 1; i < len(dicts); i++ {
			dict := dicts[i]
			replaceTable[dict] = firstDict
			toDelete[dict] = struct{}{}
		}
	}

	optimizedObjects = make([]core.PdfObject, 0, len(objects)-len(toDelete))
	for _, obj := range objects {
		if _, found := toDelete[obj]; found {
			continue
		}
		optimizedObjects = append(optimizedObjects, obj)
	}
	replaceObjectsInPlace(optimizedObjects, replaceTable)
	return optimizedObjects, nil
}
