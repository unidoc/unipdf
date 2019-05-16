/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fjson

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

func TestLoadPDFFormData(t *testing.T) {
	fdata, err := LoadFromPDFFile(`./testdata/basicform.pdf`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	data, err := fdata.JSON()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	var fields []struct {
		Name    string   `json:"name"`
		Value   string   `json:"value"`
		Options []string `json:"options"`
	}

	err = json.Unmarshal([]byte(data), &fields)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if len(fields) != 9 {
		t.Fatalf("Should have 9 fields")
	}

	// Check first field.
	if fields[0].Name != "full_name" {
		t.Fatalf("Incorrect field name (got %s)", fields[0].Name)
	}
	if fields[0].Value != "" {
		t.Fatalf("Value not empty")
	}
	if len(fields[0].Options) != 0 {
		t.Fatalf("Options not empty")
	}

	// Check another field.
	if fields[7].Name != "female" {
		t.Fatalf("Incorrect field name (got %s)", fields[7].Name)
	}
	if fields[7].Value != "Off" {
		t.Fatalf("Value not Off (got %s)", fields[7].Value)
	}
	if strings.Join(fields[7].Options, ", ") != "Off, Yes" {
		t.Fatalf("Wrong options (got %#v)", fields[7].Options)
	}
}

// Tests loading JSON form data, filling in a form and loading the PDF form data and comparing the results.
func TestFillPDFForm(t *testing.T) {
	fdata, err := LoadFromJSONFile(`./testdata/formdata.json`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	data, err := fdata.JSON()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Open pdf to fill.
	var pdfReader *model.PdfReader
	{
		f, err := os.Open(`./testdata/basicform.pdf`)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		defer f.Close()
		pdfReader, err = model.NewPdfReader(f)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	}

	err = pdfReader.AcroForm.Fill(fdata)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Write to buffer.
	var buf bytes.Buffer
	{
		// TODO(gunnsth): Implement a simpler method for populating all pages from a reader.
		pdfWriter := model.NewPdfWriter()
		for i := range pdfReader.PageList {
			err := pdfWriter.AddPage(pdfReader.PageList[i])
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
		}
		err := pdfWriter.SetForms(pdfReader.AcroForm)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		err = pdfWriter.Write(&buf)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	}

	bufReader := bytes.NewReader(buf.Bytes())

	// Read back.
	fdata2, err := LoadFromPDF(bufReader)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	data2, err := fdata2.JSON()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if data != data2 {
		t.Fatalf("%s != %s", data, data2)
	}
}
