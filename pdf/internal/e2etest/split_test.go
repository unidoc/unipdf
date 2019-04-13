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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model"
)

// Split tests splits a single page from a PDF, writes out and performs a sanity check on the output with ghostscript.
// Also checks memory use.
// Set environment variables:
//		UNIDOC_E2E_FORCE_TESTS to "1" to force the tests to execute.
//		UNIDOC_SPLIT_TESTDATA to the path of the corpus folder.
//		UNIDOC_GS_BIN_PATH to the path of the ghostscript binary (gs) for validation.
var (
	splitCorpusFolder = os.Getenv("UNIDOC_SPLIT_TESTDATA")
)

// knownHashes defines a list of known output hashes to ensure that the output is constant.
// If there is a change in hash need to find out why and update only if the change is accepted.
var knownHashes = map[string]string{
	"bf7c9d5dabc7e7ec2fc0cf9db2d9c8e7aa456fca.pdf": "0858340ec31e869d7e65a67f5e7b97f5",
	"371dce2c2720581a3eef3f123e5741dd3566ef87.pdf": "2343e102e15c5a4ac6ea8e95dab032cc",
	"e815311526b50036db6e89c54af2b9626edecf30.pdf": "b3843bf3f9df85040df075646e5ffc13",
	"3bf64014e0c9e4a56f1a9363f1b34fd707bd9fa0.pdf": "be6dde4d78d1b8f07cac800b374dd5e5",
	"004feecd47e2da4f2ed5cdbbf4791a77dd59ce20.pdf": "066c6931543cc52c3df7d11592b7dda9",
	"30c0a5cff80870cd58c2738d622f5d63e37dc90c.pdf": "d1a25c9f5dd1cfc2b665eb4999a1135c",
	"8f8ce400b9d66656cd09260035aa0cc3f7e46c82.pdf": "c94cd56024c724b370ecc61fb67b0fa5",
	"a35d386af4828b7221591343761191e8f9a28bc0.pdf": "c8858dffbfc091fbb4e32b26bfb33bde",
	"e815699a5234540fda89ea3a2ece055349a0d535.pdf": "024763e941869dd195438277fa187c30",
}

func TestSplitting(t *testing.T) {
	if len(splitCorpusFolder) == 0 {
		if forceTest {
			t.Fatalf("uNIDOC_SPLIT_TESTDATA not set")
		}
	}

	files, err := ioutil.ReadDir(splitCorpusFolder)
	if err != nil {
		if forceTest {
			t.Fatalf("Error opening %s: %v", splitCorpusFolder, err)
		}
		t.Skipf("Skipping split bench - unable to open UNIDOC_SPLIT_TESTDATA (%s)", splitCorpusFolder)
		return
	}

	// Make a temporary folder and clean up after.
	tempdir, err := ioutil.TempDir("", "unidoc_split")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	matchcount := 0
	for _, file := range files {
		// Ensure memory is garbage collected prior to running for consistency.
		debug.FreeOSMemory()

		t.Logf("%s", file.Name())
		fpath := filepath.Join(splitCorpusFolder, file.Name())
		params := splitParams{
			inputPath:    fpath,
			outPath:      filepath.Join(tempdir, "1.pdf"),
			gsValidation: len(ghostscriptBinPath) > 0,
		}
		splitSinglePdf(t, params)

		hash, err := hashFile(params.outPath)
		require.NoError(t, err)

		knownHash, has := knownHashes[file.Name()]
		if has {
			require.Equal(t, knownHash, hash)
			matchcount++
		} else {
			t.Logf("%s - hash: %s not in the list of known hashes", file.Name(), hash)
		}
	}

	// Ensure all the defined hashes were found.
	require.Equal(t, len(knownHashes), matchcount)

	t.Logf("Split benchmark complete for %d files in %s", len(files), splitCorpusFolder)
}

type splitParams struct {
	inputPath    string
	outPath      string
	gsValidation bool
}

func splitSinglePdf(t *testing.T, params splitParams) {
	measure := startMemoryMeasurement()

	file, err := os.Open(params.inputPath)
	require.NoError(t, err)
	defer file.Close()

	reader, err := model.NewPdfReader(file)
	require.NoError(t, err)

	isEncrypted, err := reader.IsEncrypted()
	require.NoError(t, err)
	if isEncrypted {
		auth, err := reader.Decrypt([]byte(""))
		require.NoError(t, err)
		require.True(t, auth)
	}

	numPages, err := reader.GetNumPages()
	require.NoError(t, err)

	if numPages < 1 {
		common.Log.Debug("Empty pdf - nothing to be done!")
		return
	}

	writer := model.NewPdfWriter()

	// Split the first page.
	page, err := reader.GetPage(1)
	require.NoError(t, err)

	err = writer.AddPage(page)
	require.NoError(t, err)

	of, err := os.Create(params.outPath)
	require.NoError(t, err)
	defer of.Close()

	err = writer.Write(of)
	require.NoError(t, err)

	measure.Stop()
	summary := measure.Summary()
	t.Logf("%s - summary %s", params.inputPath, summary)

	// GS validation of input, output pdfs.
	if params.gsValidation {
		common.Log.Debug("Validating input file")
		inputWarnings, err := validatePdf(params.inputPath, "")
		require.NoError(t, err)

		common.Log.Debug("Validating output file")

		warnings, err := validatePdf(params.outPath, "")
		if err != nil && warnings > inputWarnings {
			common.Log.Debug("Input warnings %d vs output %d", inputWarnings, warnings)
			t.Fatalf("Invalid PDF input %d/ output %d warnings", inputWarnings, warnings)
		}
		common.Log.Debug("Valid PDF!")
	}
}
