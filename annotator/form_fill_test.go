/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package annotator

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/fjson"
	"github.com/unidoc/unipdf/v3/model"

	"github.com/unidoc/unipdf/v3/internal/testutils"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

// TestFormFillRender tests the form fill/flatten process using the test data
// in a provided corpus directory. The corpus directory should contain
// (name.pdf, name.json) pairs of files. The filled output files are rendered
// to PNG images and compared to counterpart golden images found in a provided
// baseline directory.
// The test input parameters are specified through environment variables:
//  - UNIDOC_RENDERTEST_FORMFILL_TESTDATA
//    The test corpus directory. If not provided, the test is skipped.
//  - UNIDOC_RENDERTEST_FORMFILL_BASELINE
//    The baseline corpus directory. If not provided, the test is skipped.
//  - UNIDOC_RENDERTEST_FORMFILL_FORCETEST
//    Set to "1" to force the test to run. If enabled, the test data and
//    baseline corpus directories are required.
//  - UNIDOC_RENDERTEST_FORMFILL_SAVE_BASELINE
//    Set to "1" to save rendered images of new input files to the
//    baseline directory.
func TestFormFillRender(t *testing.T) {
	// Read environment variables.
	forceRun := os.Getenv("UNIDOC_RENDERTEST_FORMFILL_FORCETEST") == "1"

	renderInputPath := os.Getenv("UNIDOC_RENDERTEST_FORMFILL_TESTDATA")
	if renderInputPath == "" {
		if forceRun {
			t.Fatalf("UNIDOC_RENDERTEST_FORMFILL_TESTDATA not set")
		}
		t.Skip("skipping render tests; set UNIDOC_RENDERTEST_FORMFILL_TESTDATA to run")
	}

	renderBaselinePath := os.Getenv("UNIDOC_RENDERTEST_FORMFILL_BASELINE")
	if renderBaselinePath == "" {
		if forceRun {
			t.Fatalf("UNIDOC_RENDERTEST_FORMFILL_BASELINE not set")
		}
		t.Skip("skipping render tests; set UNIDOC_RENDERTEST_FORMFILL_BASELINE to run")
	}
	saveBaseline := os.Getenv("UNIDOC_RENDERTEST_FORMFILL_SAVE_BASELINE") == "1"

	// Get input file list.
	type formFillInput struct {
		filename string
		pdfPath  string
		jsonPath string
	}

	files, err := ioutil.ReadDir(renderInputPath)
	require.NoError(t, err)

	var fillInputs []*formFillInput
	for _, file := range files {
		basename := file.Name()

		ext := filepath.Ext(basename)
		if ext != ".json" {
			continue
		}
		filename := strings.TrimSuffix(basename, ext)

		fillInputs = append(fillInputs, &formFillInput{
			filename: filename,
			pdfPath:  filepath.Join(renderInputPath, filename+".pdf"),
			jsonPath: filepath.Join(renderInputPath, basename),
		})
	}

	lenFillInputs := len(fillInputs)
	if lenFillInputs == 0 {
		t.Skip("skipping render tests; no input files found")
	}

	// Create a temporary folder and clean up after the tests finish.
	tempDir, err := ioutil.TempDir("", "unidoc_form_fill")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Fill and render input files.
	appearance := FieldAppearance{RegenerateTextFields: true}
	style := appearance.Style()
	style.BorderSize = 1
	style.AutoFontSizeFraction = 0.70
	appearance.SetStyle(style)

	for i, fi := range fillInputs {
		t.Logf("(%d/%d) Running render tests for file %s", i+1, lenFillInputs, fi.pdfPath)

		fillAndRender := func(flatten, regenerate bool) {
			outputPath := fi.filename + "_fill"
			if flatten {
				outputPath += "_flatten"
			}
			if regenerate {
				outputPath += "_regenerate"
			}
			outputPath = filepath.Join(tempDir, outputPath+".pdf")
			appearance.OnlyIfMissing = !regenerate

			testWriteFilledForm(t, fi.pdfPath, fi.jsonPath, outputPath, flatten, appearance)
			testutils.RunRenderTest(t, outputPath, tempDir, renderBaselinePath, saveBaseline)
		}

		// Fill, regenerate appearance only if missing.
		fillAndRender(false, false)

		// Fill, flatten, regenerate appearance only if missing.
		fillAndRender(true, false)

		// Fill, regenerate appearance.
		fillAndRender(false, true)

		// Fill, flatten, regenerate appearance.
		fillAndRender(true, true)
	}
}

func testWriteFilledForm(t *testing.T, pdfPath, jsonPath, outputPath string,
	flatten bool, appearance model.FieldAppearanceGenerator) {
	// Load JSON template.
	jsonData, err := fjson.LoadFromJSONFile(jsonPath)
	require.NoError(t, err)

	// Open input file.
	inputFile, err := os.Open(pdfPath)
	require.NoError(t, err)
	defer inputFile.Close()

	// Create reader.
	reader, err := model.NewPdfReader(inputFile)
	require.NoError(t, err)
	require.NotNil(t, reader.AcroForm)

	// Fill form fields.
	require.NoError(t, reader.AcroForm.FillWithAppearance(jsonData, appearance))

	// Flatten form fields.
	if flatten {
		require.NoError(t, reader.FlattenFields(true, appearance))
	}

	// Create writer.
	writer := model.NewPdfWriter()
	writer.SetForms(reader.AcroForm)

	for _, page := range reader.PageList {
		require.NoError(t, writer.AddPage(page))
	}

	// Write output file.
	outputFile, err := os.Create(outputPath)
	require.NoError(t, err)
	defer outputFile.Close()

	require.NoError(t, writer.Write(outputFile))
}
