/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fjson

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// FieldData represents form field data loaded from JSON file.
type FieldData struct {
	values []fieldValue
}

// fieldValue represents a field name and value for a PDF form field.  Options lists
// a list of allowed values if present.
type fieldValue struct {
	Name    string   `json:"name"`
	Value   string   `json:"value"`
	Options []string `json:"options,omitempty"`
}

// LoadJSON loads JSON form data from `r`.
func LoadJSON(r io.Reader) (*FieldData, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var fdata FieldData
	err = json.Unmarshal(data, &fdata.values)
	if err != nil {
		return nil, err
	}

	return &fdata, nil
}

// LoadJSONFromPath loads form field data from a JSON file.
func LoadJSONFromPath(jsonPath string) (*FieldData, error) {
	f, err := os.Open(jsonPath)
	if err != nil {
		return nil, err
	}

	return LoadJSON(f)
}

// LoadPDF loads form field data from a PDF.
func LoadPDF(rs io.ReadSeeker) (*FieldData, error) {
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
		var val string
		switch t := f.V.(type) {
		case *core.PdfObjectString:
			val = t.Decoded()
		default:
			common.Log.Debug("%s: Unsupported %T", name, t)
			if len(f.Annotations) > 0 {
				for _, wa := range f.Annotations {
					state, found := core.GetName(wa.AS)
					if found {
						val = state.String()
					}

					// Options are the keys in the N/D dictionaries in the AP appearance dict.
					apDict, has := core.GetDict(wa.AP)
					if has {
						nDict, has := core.GetDict(apDict.Get("N"))
						if has {
							for _, key := range nDict.Keys() {
								keystr := key.String()
								if _, has := optMap[keystr]; !has {
									options = append(options, keystr)
									optMap[keystr] = struct{}{}
								}
							}
						}
						dDict, has := core.GetDict(apDict.Get("D"))
						if has {
							for _, key := range dDict.Keys() {
								keystr := key.String()
								if _, has := optMap[keystr]; !has {
									options = append(options, keystr)
									optMap[keystr] = struct{}{}
								}
							}
						}
					}
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

// LoadPDFfromPath loads form field data from a PDF file.
func LoadPDFFromPath(pdfPath string) (*FieldData, error) {
	f, err := os.Open(pdfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return LoadPDF(f)
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
