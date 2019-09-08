/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
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
	"bf7c9d5dabc7e7ec2fc0cf9db2d9c8e7aa456fca.pdf": "fdd638603c6f655babbc90358de66107",
	"371dce2c2720581a3eef3f123e5741dd3566ef87.pdf": "4c5356ac623a96004d80315f24613fff",
	"e815311526b50036db6e89c54af2b9626edecf30.pdf": "97dcfdde59a2f3a6eb105d0c31ebd3fb",
	"3bf64014e0c9e4a56f1a9363f1b34fd707bd9fa0.pdf": "6f310c9fdd44d49766d3cc32d3053b89",
	"004feecd47e2da4f2ed5cdbbf4791a77dd59ce20.pdf": "309a072a97d0566aa3f85edae504bb53",
	"30c0a5cff80870cd58c2738d622f5d63e37dc90c.pdf": "67d7c2fbf21dd9d65c8bb9ab29dfec60",
	"8f8ce400b9d66656cd09260035aa0cc3f7e46c82.pdf": "679650c27697a7b83ee792692daaff18",
	"a35d386af4828b7221591343761191e8f9a28bc0.pdf": "1955d6cf29715652bea999bcbadc818b",
	"e815699a5234540fda89ea3a2ece055349a0d535.pdf": "5a1d97ee1aabc5dcacbbf3cd164b964d",
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
		t.Logf("%s", file.Name())
		fpath := filepath.Join(splitCorpusFolder, file.Name())
		params := splitParams{
			inputPath:    fpath,
			outPath:      filepath.Join(tempdir, "split_1_"+file.Name()),
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

	reader, err := model.NewPdfReaderLazy(file)
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

	model.SetPdfProducer("UniDoc")
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
