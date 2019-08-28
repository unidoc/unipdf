/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"archive/zip"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// Extract images test writes out a zip file containing all images extracted
// from the subject PDF file and compares its hash with a known zip file hash.
// Also checks memory usage.
// Set environment variables:
//		UNIDOC_E2E_FORCE_TESTS to "1" to force the tests to execute.
//		UNIDOC_EXTRACT_IMAGES_TESTDATA to the path of the corpus folder.
var (
	extractImagesCorpusFolder = os.Getenv("UNIDOC_EXTRACT_IMAGES_TESTDATA")
)

// knownExtrImgsHashes defines a list of known output hashes to ensure that the output is constant.
// If there is a change in hash need to find out why and update only if the change is accepted.
var knownExtrImgsHashes = map[string]string{
	"1ecec6aa4abed1855fb88916d7feb8c9692daaf5.pdf": "64899eb2c683f2e0b1ce0e35b5377aed",
	"7eee345c983461d44ae939b3f800a97642817c8d.pdf": "5a4cd00537f8a1f39fff2d4c6dd5cc53",
	"52ab322c1697aca9bad37288f7c502e37fa657af.pdf": "2bddee02dff89a38c08322c9d2c779a6",
	"0edf09fd438db2f18c1bb08fccc1f81a7b280bf2.pdf": "583f755b3fb1bd5697036616442687ab",
	"cafe55316a45435c3817f4c1b6a19c9cd52db825.pdf": "b199badff0abb0311a2cbe35c7fce580",
	"6773e6aa5d8a2d26362cf3fca2874b3a81025bae.pdf": "f052e3e333839508a8bdd8d1a3ba1973",
	"d11a3ca55664828b69d7c39d83d5c0a63fcea89d.pdf": "29287cd44f009dce5aa9c2a0dc9a3c83",
	"483933bf73cc4fcc264eb69214ff763ccf299e49.pdf": "627dcf88805786d03b2e76d367b42642",
	"da1c5c4c4fe36f676dbca6ea01673c9fdf77c7a9.pdf": "7ffe076d1c7c88e07dfba7545b622ec6",
	"f856baf7ffcd96003b6bda800171cb0e5680f78e.pdf": "a9505d8c22f1fd063fbe0b05aa33a5fc",
	"201c20676fe8da14a8130852c91ed58b48cba8fb.pdf": "ffcb78d126c04be9ca2497bb43b6e964",
	"f0152456494aa09e5cf82c4afe9ecd2fdc2e8d72.pdf": "d0e68157aaa7f9f4406807512db3e676",
	"d95643acea1ec3f6215bda35e4cd89dbd8898c44.pdf": "1739aed3e1cbfa5e98f8d7fef17a614b",
	"110d793aeaa7accbe40b5ab9db249d5a103d3b50.pdf": "a57e347edddfd3f6032b85553b3537cd",
	"d15a0aa289524619a971188372dd05fb712f1b2c.pdf": "380907273bb8ea64943d986491d827ec",
	"932e0dfa52c20ffe83b8178fb98296a0dab177d1.pdf": "b44d8b073f99ac3db28d7951e3c7d970",
	"60a8c28da5c23081834adac4170755904d8c4166.pdf": "9167f381d5eed7a2e5fd10eca567c519",
	"e51296be2615b9389482c9c16505286619b6cf36.pdf": "ec6e1f6297dd1cbda6ccba39e0c7d3d2",
}

func TestExtractImages(t *testing.T) {
	if len(extractImagesCorpusFolder) == 0 {
		if forceTest {
			t.Fatalf("UNIDOC_EXTRACT_IMAGES_TESTDATA not set")
		}
	}

	files, err := ioutil.ReadDir(extractImagesCorpusFolder)
	if err != nil {
		if forceTest {
			t.Fatalf("Error opening %s: %v", extractImagesCorpusFolder, err)
		}
		t.Skipf("Skipping extract images bench - unable to open UNIDOC_EXTRACT_IMAGES_TESTDATA (%s)", extractImagesCorpusFolder)
		return
	}

	// Make a temporary folder and clean up after.
	tempdir, err := ioutil.TempDir("", "unidoc_extract_images")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	matchcount := 0
	for _, file := range files {
		basename := filepath.Base(file.Name())
		outName := strings.TrimSuffix(basename, filepath.Ext(basename)) + ".zip"

		t.Logf("%s", file.Name())
		fpath := filepath.Join(extractImagesCorpusFolder, file.Name())
		params := extractImagesParams{
			inputPath: fpath,
			outPath:   filepath.Join(tempdir, "extract_images_"+outName),
		}
		extractImagesSinglePdf(t, params)

		hash, err := hashFile(params.outPath)
		require.NoError(t, err)

		knownHash, has := knownExtrImgsHashes[file.Name()]
		if has {
			require.Equal(t, knownHash, hash)
			matchcount++
		} else {
			t.Logf("%s - hash: %s not in the list of known hashes", file.Name(), hash)
		}
	}

	// Ensure all the defined hashes were found.
	require.Equal(t, len(knownExtrImgsHashes), matchcount)

	t.Logf("Extract images benchmark complete for %d files in %s", len(files), extractImagesCorpusFolder)
}

type extractImagesParams struct {
	inputPath string
	outPath   string
}

func extractImagesSinglePdf(t *testing.T, params extractImagesParams) {
	measure := startMemoryMeasurement()

	// Create PDF reader.
	file, err := os.Open(params.inputPath)
	require.NoError(t, err)
	defer file.Close()

	reader, err := model.NewPdfReaderLazy(file)
	require.NoError(t, err)

	// Decrypt file, if necessary.
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

	// Create output zip file.
	outFile, err := os.Create(params.outPath)
	require.NoError(t, err)
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	for i := 0; i < numPages; i++ {
		page, err := reader.GetPage(i + 1)
		require.NoError(t, err)

		// Extract page images.
		imgExtractor, err := extractor.New(page)
		require.NoError(t, err)

		extracted, err := imgExtractor.ExtractPageImages(nil)
		require.NoError(t, err)

		for idx, img := range extracted.Images {
			// Convert extracted image to Go image.
			goImg, err := img.Image.ToGoImage()
			require.NoError(t, err)

			// Create zip file.
			imgFile, err := zipWriter.Create(fmt.Sprintf("p%d_%d.jpg", i+1, idx))
			require.NoError(t, err)

			// Write zip file.
			err = jpeg.Encode(imgFile, goImg, &jpeg.Options{Quality: 100})
			require.NoError(t, err)
		}
	}

	err = zipWriter.Close()
	require.NoError(t, err)

	measure.Stop()
	summary := measure.Summary()
	t.Logf("%s - summary %s", params.inputPath, summary)
}
