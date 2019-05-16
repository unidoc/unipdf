/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"crypto/md5"

	"github.com/unidoc/unipdf/v3/core"
)

// CombineDuplicateDirectObjects combines duplicated direct objects by its data hash.
// It implements interface model.Optimizer.
type CombineDuplicateDirectObjects struct {
}

// Optimize optimizes PDF objects to decrease PDF size.
func (dup *CombineDuplicateDirectObjects) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	updateObjectNumbers(objects)
	dictsByHash := make(map[string][]*core.PdfObjectDictionary)
	var processDict func(pDict *core.PdfObjectDictionary)

	processDict = func(pDict *core.PdfObjectDictionary) {
		for _, key := range pDict.Keys() {
			obj := pDict.Get(key)
			if dict, isDictObj := obj.(*core.PdfObjectDictionary); isDictObj {
				hasher := md5.New()
				hasher.Write([]byte(dict.WriteString()))
				hash := string(hasher.Sum(nil))
				dictsByHash[hash] = append(dictsByHash[hash], dict)
				processDict(dict)
			}
		}
	}

	for _, obj := range objects {
		ind, isIndirectObj := obj.(*core.PdfIndirectObject)
		if !isIndirectObj {
			continue
		}
		if dict, isDictObj := ind.PdfObject.(*core.PdfObjectDictionary); isDictObj {
			processDict(dict)
		}
	}

	indirects := make([]core.PdfObject, 0, len(dictsByHash))
	replaceTable := make(map[core.PdfObject]core.PdfObject)

	for _, dicts := range dictsByHash {
		if len(dicts) < 2 {
			continue
		}
		dict := core.MakeDict()
		dict.Merge(dicts[0])
		ind := core.MakeIndirectObject(dict)
		indirects = append(indirects, ind)
		for i := 0; i < len(dicts); i++ {
			dict := dicts[i]
			replaceTable[dict] = ind
		}
	}

	optimizedObjects = make([]core.PdfObject, len(objects))
	copy(optimizedObjects, objects)
	optimizedObjects = append(indirects, optimizedObjects...)
	replaceObjectsInPlace(optimizedObjects, replaceTable)
	return optimizedObjects, nil
}
