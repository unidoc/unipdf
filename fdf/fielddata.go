/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fdf

import (
	"errors"
	"io"
	"os"
	"sort"

	"github.com/unidoc/unipdf/v3/core"
)

// Data represents forms data format (FDF) file data.
type Data struct {
	root   *core.PdfObjectDictionary
	fields *core.PdfObjectArray
}

// Load loads FDF form data from `r`.
func Load(r io.ReadSeeker) (*Data, error) {
	p, err := newParser(r)
	if err != nil {
		return nil, err
	}

	fdfDict, err := p.Root()
	if err != nil {
		return nil, err
	}

	fields, found := core.GetArray(fdfDict.Get("Fields"))
	if !found {
		return nil, errors.New("fields missing")
	}

	return &Data{
		fields: fields,
		root:   fdfDict,
	}, nil
}

// LoadFromPath loads FDF form data from file path `fdfPath`.
func LoadFromPath(fdfPath string) (*Data, error) {
	f, err := os.Open(fdfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Load(f)
}

// FieldDictionaries returns a map of field names to field dictionaries.
func (fdf *Data) FieldDictionaries() (map[string]*core.PdfObjectDictionary, error) {
	fieldDataMap := map[string]*core.PdfObjectDictionary{}

	for i := 0; i < fdf.fields.Len(); i++ {
		fieldDict, has := core.GetDict(fdf.fields.Get(i))
		if has {
			// Key value field data.
			t, _ := core.GetString(fieldDict.Get("T"))
			if t != nil {
				fieldDataMap[t.Str()] = fieldDict
			}
		}
	}

	return fieldDataMap, nil
}

// FieldValues implements interface model.FieldValueProvider.
// Returns a map of field names to values (PdfObjects).
func (fdf *Data) FieldValues() (map[string]core.PdfObject, error) {
	fieldDictMap, err := fdf.FieldDictionaries()
	if err != nil {
		return nil, err
	}

	var keys []string
	for fieldName := range fieldDictMap {
		keys = append(keys, fieldName)
	}
	sort.Strings(keys)

	fieldValMap := map[string]core.PdfObject{}
	for _, fieldName := range keys {
		fieldDict := fieldDictMap[fieldName]
		val := core.TraceToDirectObject(fieldDict.Get("V"))
		fieldValMap[fieldName] = val
	}

	return fieldValMap, nil
}
