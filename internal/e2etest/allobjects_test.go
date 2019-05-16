/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// The allobjects test probes through all objects in PDF files of a specified corpus, decoding streams.
// Set environment variables:
//		UNIDOC_E2E_FORCE_TESTS to "1" to force the tests to execute.
//		UNIDOC_ALLOBJECTS_TESTDATA to the path of the corpus folder.
var (
	allObjectsCorpusFolder = os.Getenv("UNIDOC_ALLOBJECTS_TESTDATA")
)

func TestAllObjects(t *testing.T) {
	if len(allObjectsCorpusFolder) == 0 {
		if forceTest {
			t.Fatalf("UNIDOC_ALLOBJECTS_TESTDATA not set")
		}
	}

	files, err := ioutil.ReadDir(allObjectsCorpusFolder)
	if err != nil {
		if forceTest {
			t.Fatalf("Error opening %s: %v", allObjectsCorpusFolder, err)
		}
		t.Skipf("Skipping allobjects test - unable to open UNIDOC_ALLOBJECTS_TESTDATA (%s)", allObjectsCorpusFolder)
		return
	}

	for _, file := range files {
		fpath := filepath.Join(allObjectsCorpusFolder, file.Name())
		t.Logf("%s", fpath)
		err := probeAllObjectsSinglePdf(fpath)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	}
	t.Logf("allObjects test complete for %d files in %s", len(files), allObjectsCorpusFolder)
}

func probeAllObjectsSinglePdf(inputPath string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	// Try decrypting with an empty one.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return err
		}

		if !auth {
			return errors.New("unauthorized read")
		}
	}

	_, err = pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	objNums := pdfReader.GetObjectNums()

	// Output.
	for _, objNum := range objNums {
		obj, err := pdfReader.GetIndirectObjectByNumber(objNum)
		if err != nil {
			return err
		}
		if stream, is := obj.(*core.PdfObjectStream); is {
			_, err := core.DecodeStream(stream)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
