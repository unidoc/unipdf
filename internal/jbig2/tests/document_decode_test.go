/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"archive/zip"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// EnvDirectory is the environment variable that should contain directory path
// to the jbig2 encoded test files.
const EnvDirectory = "UNIDOC_JBIG2_TESTDATA"

// jbig2Validate is the runtime flag which defines if the golden files
// should be validated.
var jbig2Validate bool

func init() {
	flag.BoolVar(&jbig2Validate, "jbig2-validate", false, "checks if all the jbig2 golden files are valid")
}

// TestDecodeJBIG2Files tries to decode the provided jbig2 files.
// Requires environmental variable 'UNIDOC_JBIG2_TESTDATA' that contains the jbig2 testdata.
// Decoded images are stored within zipped archive files - that has the same name as the pdf file.
// In order to check the decoded images this function creates also the directory 'goldens'
// which would have json files for each 'pdf' input, containing valid flags.
// If the 'jbig2-validate' runtime flag is provided, the test function would
// check if all the golden files has been marked as valid.
func TestDecodeJBIG2Files(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	dirName := os.Getenv(EnvDirectory)
	if dirName == "" {
		return
	}
	filenames, err := readFileNames(dirName)
	require.NoError(t, err)

	if len(filenames) > 0 {
		_, err = os.Stat(filepath.Join(dirName, jbig2DecodedDirectory))
		if err != nil {
			err = os.Mkdir(filepath.Join(dirName, jbig2DecodedDirectory), 0666)
			require.NoError(t, err)
		}
	}
	passwords := make(map[string]string)

	for _, filename := range filenames {
		rawName := rawFileName(filename)
		t.Run(rawName, func(t *testing.T) {
			// get the file
			f, err := getFile(dirName, filename)
			require.NoError(t, err)
			defer f.Close()

			var reader *model.PdfReader
			password, ok := passwords[filename]
			if ok {
				// read the pdf with the password
				reader, err = readPDF(f, password)
			} else {
				reader, err = readPDF(f)
			}
			if err != nil {
				if err.Error() != "EOF not found" {
					require.NoError(t, err)
				}
			}

			numPages, err := reader.GetNumPages()
			require.NoError(t, err)

			// create zipped file
			fileName := filepath.Join(dirName, jbig2DecodedDirectory, rawName+".zip")

			w, err := os.Create(fileName)
			require.NoError(t, err)
			defer w.Close()

			zw := zip.NewWriter(w)
			defer zw.Close()

			var allHashes []fileHash

			for pageNo := 1; pageNo <= numPages; pageNo++ {
				page, err := reader.GetPage(pageNo)
				require.NoError(t, err)

				images, err := extractImagesOnPage(filepath.Join(dirName, rawName), page)
				require.NoError(t, err)

				hashes, err := writeExtractedImages(zw, rawName, pageNo, images...)
				require.NoError(t, err)

				allHashes = append(allHashes, hashes...)
			}
			checkGoldenFiles(t, dirName, rawName, allHashes...)
		})
	}
}
