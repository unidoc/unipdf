/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// Passthrough benchmark loads a PDF, writes back out and performs a sanity check on the output with ghostscript.
// Set environment variables:
//		UNIDOC_E2E_FORCE_TESTS to "1" to force the tests to execute.
//		UNIDOC_PASSTHROUGH_TESTDATA to the path of the corpus folder.
//		UNIDOC_GS_BIN_PATH to the path of the ghostscript binary (gs).
var (
	forceTest               = os.Getenv("UNIDOC_E2E_FORCE_TESTS") == "1"
	passthroughCorpusFolder = os.Getenv("UNIDOC_PASSTHROUGH_TESTDATA")
)

func TestPassthrough(t *testing.T) {
	if len(passthroughCorpusFolder) == 0 {
		if forceTest {
			t.Fatalf("UNIDOC_PASSTHROUGH_TESTDATA not set")
		}
	}

	files, err := ioutil.ReadDir(passthroughCorpusFolder)
	if err != nil {
		if forceTest {
			t.Fatalf("Error opening %s: %v", passthroughCorpusFolder, err)
		}
		t.Skipf("Skipping passthrough bench - unable to open UNIDOC_PASSTHROUGH_TESTDATA (%s)", passthroughCorpusFolder)
		return
	}

	// Make a temporary folder and clean up after.
	tempdir, err := ioutil.TempDir("", "unidoc_passthrough")
	if err != nil {
		t.Fatalf("Failed to create temporary folder")
	}
	defer os.RemoveAll(tempdir)

	for _, file := range files {
		t.Logf("%s", file.Name())
		fpath := filepath.Join(passthroughCorpusFolder, file.Name())
		params := passthroughParams{
			inputPath:    fpath,
			outPath:      filepath.Join(tempdir, "1.pdf"),
			gsValidation: len(ghostscriptBinPath) > 0,
		}
		err := passthroughSinglePdf(params)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	}
	t.Logf("Passthrough benchmark complete for %d files in %s", len(files), passthroughCorpusFolder)
}

type passthroughParams struct {
	inputPath    string
	outPath      string
	gsValidation bool
}

func passthroughSinglePdf(params passthroughParams) error {
	file, err := os.Open(params.inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader, err := model.NewPdfReader(file)
	if err != nil {
		common.Log.Debug("Reader create error %s\n", err)
		return err
	}

	isEncrypted, err := reader.IsEncrypted()
	if err != nil {
		return err
	}
	if isEncrypted {
		valid, err := reader.Decrypt([]byte(""))
		if err != nil {
			common.Log.Debug("Fail to decrypt: %v", err)
			return err
		}

		if !valid {
			return fmt.Errorf("Unable to access, encrypted")
		}
	}

	numPages, err := reader.GetNumPages()
	if err != nil {
		common.Log.Debug("Failed to get number of pages")
		return err
	}

	if numPages < 1 {
		common.Log.Debug("Empty pdf - nothing to be done!")
		return nil
	}

	writer := model.NewPdfWriter()

	// Optional content.
	ocProps, err := reader.GetOCProperties()
	if err != nil {
		return err
	}
	writer.SetOCProperties(ocProps)

	for j := 0; j < numPages; j++ {
		page, err := reader.GetPage(j + 1)
		if err != nil {
			common.Log.Debug("Get page error %s", err)
			return err
		}

		// Load and set outlines (table of contents).
		outlineTree := reader.GetOutlineTree()

		err = writer.AddPage(page)
		if err != nil {
			common.Log.Debug("Add page error %s", err)
			return err
		}

		writer.AddOutlineTree(outlineTree)
	}

	// Copy the forms over to the new document also.
	writer.SetForms(reader.AcroForm)

	of, err := os.Create(params.outPath)
	if err != nil {
		common.Log.Debug("Failed to create file (%s)", err)
		return err
	}
	defer of.Close()

	err = writer.Write(of)
	if err != nil {
		common.Log.Debug("WriteFile error")
		return err
	}

	// GS validation of input, output pdfs.
	if params.gsValidation {
		common.Log.Debug("Validating input file")
		inputWarnings, err := validatePdf(params.inputPath, "")
		if err != nil {
			return err
		}

		common.Log.Debug("Validating output file")

		warnings, err := validatePdf(params.outPath, "")
		if err != nil && warnings > inputWarnings {
			common.Log.Debug("Input warnings %d vs output %d", inputWarnings, warnings)
			return fmt.Errorf("Invalid PDF input %d/ output %d warnings", inputWarnings, warnings)
		}
		common.Log.Debug("Valid PDF!")
	}

	return nil
}
