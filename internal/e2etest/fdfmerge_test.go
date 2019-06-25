/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
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
	"NW_null_Business_V04.fdf":      "45ccf325025b366d6e9c90bde53da1aa",
	"NW_null_Business_V05.fdf":      "49c0c89cd2384c8a75f9cb20a778698f",
	"NW_null_Business_V05.v1.2.fdf": "49c0c89cd2384c8a75f9cb20a778698f",
	"NW_null_Contract_V04.fdf":      "80ebec761eba106cda38a4613819634e",
	"N_null_Contract.fdf":           "557e9d5788ba418e3e5f6ffdf710a3b9",
	"Network_Contract_V01.fdf":      "3bf058c9e4cefae222c92caa28fca603",
	"checkmark_check.fdf":           "b95c8f8c0673e5541d28f212c0b25b5b",
	"checkmark_circle.fdf":          "06bda9e3539e63aebdfc20f8fe3d83e9",
	"checkmark_cross.fdf":           "34dc015cf122bffcef8c62c559fc0ac7",
	"checkmark_diamond.fdf":         "5a3c2951da0aa2943e9007d4baed82bf",
	"checkmark_square.fdf":          "83d97592cd75c2c62a2e6ae2962379db",
	"checkmark_star.fdf":            "2e460f069e474714573724255fcdffda",
	"test_fail.fdf":                 "d7eb6071341f823a64f7234a20830d74",
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

		// Ensure memory is garbage collected prior to running for consistency.
		debug.FreeOSMemory()

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
		// FIXME: Hack needed to ensure that annotations are loaded.
		// TODO: Remove.  Resolved in PR#93.
		{
			_, err := p.GetAnnotations()
			require.NoError(t, err)
		}

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
