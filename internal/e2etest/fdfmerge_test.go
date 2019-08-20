/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/annotator"
	"github.com/unidoc/unipdf/v3/fdf"
	"github.com/unidoc/unipdf/v3/model"
)

// FDF merge tests merge FDF data into template PDF data and flattens to an output PDF file.
// Output files are checked with ghostscript and memory consumption is measured.
// Set environment variables:
//		UNIDOC_E2E_FORCE_TESTS to "1" to force the tests to execute.
//		UNIDOC_FDFMERGE_TESTDATA to the path of the corpus folder.
//		UNIDOC_GS_BIN_PATH to the path of the ghostscript binary (gs) for validation.
var (
	fdfMergeCorpusFolder = os.Getenv("UNIDOC_FDFMERGE_TESTDATA")
)

// fdfMergeHashes defines a list of known output hashes to ensure that the output is constant.
// If there is a change in hash need to find out why and update only if the change is accepted.
var fdfMergeHashes = map[string]string{
	"NW_null_Business_V04.fdf":      "6e33f219994e4b9ee1e1843c976504df",
	"NW_null_Business_V05.fdf":      "ff1f8bd39f9be9844a6d85bafe07c790",
	"NW_null_Business_V05.v1.2.fdf": "ff1f8bd39f9be9844a6d85bafe07c790",
	"NW_null_Contract_V04.fdf":      "a54f4b42dc34997cfb701ef647cdbdfe",
	"N_null_Contract.fdf":           "c173340d6492984532cf51a4f5ceb4b6",
	"Network_Contract_V01.fdf":      "0ae2537bf8a8366aa97c1ca965b88d1f",
	"checkmark_check.fdf":           "8892cdb01318421f8d198233b80ab8e3",
	"checkmark_circle.fdf":          "3b1e6ef6aae2a7497b090e0960d2c163",
	"checkmark_cross.fdf":           "6b16b6d7437a3f59a7e9e72c1ecfd59b",
	"checkmark_diamond.fdf":         "123488e428914832f21e213339ed74f1",
	"checkmark_square.fdf":          "d0ac69dac7a933e440a5005b1712edeb",
	"checkmark_star.fdf":            "1326f152fb8158dffc08e5bb51cba1bc",
	"test_fail.fdf":                 "9a90cef679d6b4c13017c73c2528ca75",
}

// Test filling (fdf merge) and flattening form data and annotations.
func TestFdfMerging(t *testing.T) {
	if len(fdfMergeCorpusFolder) == 0 {
		if forceTest {
			t.Fatalf("UNIDOC_FDFMERGE_TESTDATA not set")
		}
		t.Skipf("UNIDOC_FDFMERGE_TESTDATA not set")
	}

	files, err := ioutil.ReadDir(fdfMergeCorpusFolder)
	if err != nil {
		if forceTest {
			t.Fatalf("Error opening %s: %v", fdfMergeCorpusFolder, err)
		}
		t.Skipf("Skipping flatten bench - unable to open UNIDOC_FDFMERGE_TESTDATA (%s)", fdfMergeCorpusFolder)
	}

	// Make a temporary folder and clean up after.
	tempdir, err := ioutil.TempDir("", "unidoc_fdfmerge")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	matchcount := 0
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file.Name())) != ".fdf" {
			continue
		}
		fdfPath := filepath.Join(fdfMergeCorpusFolder, file.Name())
		bareName := strings.TrimSuffix(file.Name(), ".fdf")
		pdfPath := filepath.Join(fdfMergeCorpusFolder, bareName+".pdf")

		t.Logf("%s", file.Name())
		params := fdfMergeParams{
			templatePath: pdfPath,
			fdfPath:      fdfPath,
			outPath:      filepath.Join(tempdir, "filled_flatten_1_"+bareName+".pdf"),
			gsValidation: len(ghostscriptBinPath) > 0,
		}
		fdfMergeSingle(t, params)

		hash, err := hashFile(params.outPath)
		require.NoError(t, err)

		knownHash, has := fdfMergeHashes[file.Name()]
		if has {
			require.Equal(t, knownHash, hash)
			matchcount++
		} else {
			t.Logf("Output: %s", params.outPath)
			t.Logf("%s - hash: %s not in the list of known hashes", file.Name(), hash)
		}
	}

	// Ensure all the defined hashes were found.
	require.Equal(t, len(fdfMergeHashes), matchcount)

	t.Logf("FDF merge benchmark complete for %d cases in %s", matchcount, fdfMergeCorpusFolder)
}

type fdfMergeParams struct {
	templatePath string // template PDF file.
	fdfPath      string // form data FDF file.
	outPath      string
	gsValidation bool
}

func fdfMergeSingle(t *testing.T, params fdfMergeParams) {
	measure := startMemoryMeasurement()

	fdfData, err := fdf.LoadFromPath(params.fdfPath)
	require.NoError(t, err)

	f, err := os.Open(params.templatePath)
	require.NoError(t, err)
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	require.NoError(t, err)

	// Populate the form data.
	err = pdfReader.AcroForm.Fill(fdfData)
	require.NoError(t, err)

	// Flatten form.
	fieldAppearance := annotator.FieldAppearance{OnlyIfMissing: true, RegenerateTextFields: true}

	// NOTE: To customize certain styles try:
	style := fieldAppearance.Style()
	style.CheckmarkRune = '✖'
	style.AutoFontSizeFraction = 0.70
	fieldAppearance.SetStyle(style)

	err = pdfReader.FlattenFields(true, fieldAppearance)
	require.NoError(t, err)

	// Write out.
	model.SetPdfProducer("UniDoc")
	pdfWriter := model.NewPdfWriter()
	pdfWriter.SetForms(nil)

	for _, p := range pdfReader.PageList {
		err = pdfWriter.AddPage(p)
		require.NoError(t, err)
	}

	fout, err := os.Create(params.outPath)
	require.NoError(t, err)
	defer fout.Close()

	err = pdfWriter.Write(fout)
	require.NoError(t, err)

	measure.Stop()
	summary := measure.Summary()
	t.Logf("%s - summary %s", params.templatePath, summary)
}
