/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
)

// TestDecodeJBIG2Files tries to decode the provided jbig2 files.
// Requires environmental variable 'UNIDOC_JBIG2_TESTDATA' that contains the jbig2 testdata.
// Decoded images are stored within zipped archive files - that has the same name as the pdf file.
// In order to check the decoded images this function creates also the directory 'goldens'
// which would have json files for each 'pdf' input, containing valid image hashes.
// If the 'jbig2-update-goldens' runtime flag is provided, the test function updates all the 'hashes'
// for the decoded jbig2 images in related 'golden' files.
// In order to check the decoded images use 'jbig2-store-images' flag, then the function would store them
// within zipped files in the os.TempDir()/unipdf/jbig2 directory.
func TestDecodeJBIG2Files(t *testing.T) {
	dirName := os.Getenv(EnvJBIG2Directory)
	if dirName == "" {
		return
	}

	filenames, err := readFileNames(dirName, "pdf")
	require.NoError(t, err)

	if len(filenames) == 0 {
		return
	}

	tempDir := filepath.Join(os.TempDir(), "unipdf", "jbig2", "decoded")
	err = os.MkdirAll(tempDir, 0700)
	require.NoError(t, err)

	var f *os.File
	switch {
	case logToFile:
		fileName := filepath.Join(tempDir, fmt.Sprintf("log_%s.txt", time.Now().Format("20060102")))
		f, err = os.Create(fileName)
		require.NoError(t, err)
		common.SetLogger(common.NewWriterLogger(common.LogLevelTrace, f))
	case testing.Verbose():
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	// clear all the temporary files
	defer func() {
		if f != nil {
			f.Close()
		}

		switch {
		case !keepImageFiles && !logToFile:
			err = os.RemoveAll(filepath.Join(tempDir))
		case !keepImageFiles:
			err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if strings.HasSuffix(info.Name(), "zip") {
					return os.Remove(path)
				}
				return nil
			})
		}
		if err != nil {
			common.Log.Error(err.Error())
		}
	}()

	for _, filename := range filenames {
		rawName := rawFileName(filename)
		t.Run(rawName, func(t *testing.T) {
			images, err := extractImages(dirName, filename)
			require.NoError(t, err)

			// create zipped file
			fileName := filepath.Join(tempDir, rawName+".zip")

			w, err := os.Create(fileName)
			require.NoError(t, err)
			defer w.Close()

			zw := zip.NewWriter(w)
			defer zw.Close()

			err = writeExtractedImages(zw, images...)
			require.NoError(t, err)

			checkImageGoldenFiles(t, dirName, rawName, images...)
		})
	}
}
