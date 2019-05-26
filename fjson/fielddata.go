/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fjson

import (
	"encoding/json"
	"io"
	"os"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// FieldData represents form field data loaded from JSON file.
type FieldData struct {
	values []fieldValue
}

// fieldValue represents a field name and value for a PDF form field.
type fieldValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`

	// Options lists allowed values if present.
	Options []string `json:"options,omitempty"`
}

// LoadFromJSON loads JSON form data from `r`.
func LoadFromJSON(r io.Reader) (*FieldData, error) {
	var fdata FieldData
	err := json.NewDecoder(r).Decode(&fdata.values)
	if err != nil {
		return nil, err
	}
	return &fdata, nil
}

// LoadFromJSONFile loads form field data from a JSON file.
func LoadFromJSONFile(filePath string) (*FieldData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return LoadFromJSON(f)
}

// LoadFromPDF loads form field data from a PDF.
func LoadFromPDF(rs io.ReadSeeker) (*FieldData, error) {
	pdfReader, err := model.NewPdfReader(rs)
	if err != nil {
		return nil, err
	}

	if pdfReader.AcroForm == nil {
		return nil, nil
	}

	var fieldvals []fieldValue
	fields := pdfReader.AcroForm.AllFields()
	for _, f := range fields {
		var options []string
		optMap := make(map[string]struct{})

		name, err := f.FullName()
		if err != nil {
			return nil, err
		}

		if t, ok := f.V.(*core.PdfObjectString); ok {
			fieldvals = append(fieldvals, fieldValue{
				Name:  name,
				Value: t.Decoded(),
			})
			continue
		}

		var val string
		for _, wa := range f.Annotations {
			state, found := core.GetName(wa.AS)
			if found {
				val = state.String()
			}

			// Options are the keys in the N/D dictionaries in the AP appearance dict.
			apDict, has := core.GetDict(wa.AP)
			if !has {
				continue
			}
			nDict, _ := core.GetDict(apDict.Get("N"))
			for _, key := range nDict.Keys() {
				keystr := key.String()
				if _, has := optMap[keystr]; !has {
					options = append(options, keystr)
					optMap[keystr] = struct{}{}
				}
			}
			dDict, _ := core.GetDict(apDict.Get("D"))
			for _, key := range dDict.Keys() {
				keystr := key.String()
				if _, has := optMap[keystr]; !has {
					options = append(options, keystr)
					optMap[keystr] = struct{}{}
				}
			}
		}

		fval := fieldValue{
			Name:    name,
			Value:   val,
			Options: options,
		}
		fieldvals = append(fieldvals, fval)
	}

	fdata := FieldData{
		values: fieldvals,
	}

	return &fdata, nil
}

// LoadFromPDFFile loads form field data from a PDF file.
func LoadFromPDFFile(filePath string) (*FieldData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return LoadFromPDF(f)
}

// JSON returns the field data as a string in JSON format.
func (fd FieldData) JSON() (string, error) {
	data, err := json.MarshalIndent(fd.values, "", "    ")
	return string(data), err
}

// FieldValues implements model.FieldValueProvider interface.
func (fd *FieldData) FieldValues() (map[string]core.PdfObject, error) {
	fvalMap := make(map[string]core.PdfObject)
	for _, fval := range fd.values {
		if len(fval.Value) > 0 {
			fvalMap[fval.Name] = core.MakeString(fval.Value)
		}
	}
	return fvalMap, nil
}
