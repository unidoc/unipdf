/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"bytes"

	"github.com/unidoc/unipdf/v3/core"
)

type objectStructure struct {
	catalogDict *core.PdfObjectDictionary
	pagesDict   *core.PdfObjectDictionary
	pages       []*core.PdfIndirectObject
}

// getObjectStructure identifies the Catalog and Pages dictionary and finds a list of pages.
func getObjectStructure(objects []core.PdfObject) objectStructure {
	objstr := objectStructure{}
	found := false
	for _, obj := range objects {
		switch t := obj.(type) {
		case *core.PdfIndirectObject:
			dict, is := core.GetDict(t)
			if !is {
				continue
			}
			kind, is := core.GetName(dict.Get("Type"))
			if !is {
				continue
			}

			switch kind.String() {
			case "Catalog":
				objstr.catalogDict = dict
				found = true
			}
		}
		if found {
			break
		}
	}

	if !found {
		return objstr
	}

	pagesDict, ok := core.GetDict(objstr.catalogDict.Get("Pages"))
	if !ok {
		return objstr
	}
	objstr.pagesDict = pagesDict

	kids, ok := core.GetArray(pagesDict.Get("Kids"))
	if !ok {
		return objstr
	}
	for _, obj := range kids.Elements() {
		pobj, ok := core.GetIndirect(obj)
		if !ok {
			break
		}
		objstr.pages = append(objstr.pages, pobj)
	}

	return objstr
}

// getPageContents loads the page content stream as a string from a /Contents entry.
// Either a single stream, or an array of streams. Returns the list of objects that
// can be used if need to replace.
func getPageContents(contentsObj core.PdfObject) (contents string, objs []core.PdfObject) {
	var buf bytes.Buffer

	switch t := contentsObj.(type) {
	case *core.PdfIndirectObject:
		objs = append(objs, t)
		contentsObj = t.PdfObject
	}

	switch t := contentsObj.(type) {
	case *core.PdfObjectStream:
		if decoded, err := core.DecodeStream(t); err == nil {
			buf.Write(decoded)
			objs = append(objs, t)
		}
	case *core.PdfObjectArray:
		for _, elobj := range t.Elements() {
			switch el := elobj.(type) {
			case *core.PdfObjectStream:
				if decoded, err := core.DecodeStream(el); err == nil {
					buf.Write(decoded)
					objs = append(objs, el)
				}
			}
		}
	}
	return buf.String(), objs
}
