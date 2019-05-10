// +build integration
/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"archive/zip"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	pdf "github.com/unidoc/unidoc/pdf/model"
	"os"
	"path/filepath"
	"testing"
)

const (
	// EnvDirectory is the environment directory that defined the test files directory
	EnvDirectory = "JBIG2"
)

// TestDecodeJBIG2Files tries to decode the provided jbig2 files
// Requires environmental variable 'JBIG2' that contains the jbig2 testdata
func TestDecodeJBIG2Files(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
	}
	dirName := os.Getenv(EnvDirectory)
	if dirName == "" {
		return
	}

	require.NoError(t, os.RemoveAll(filepath.Join(dirName, jbig2ImagesDir)))
	require.NoError(t, os.Mkdir(filepath.Join(dirName, jbig2ImagesDir), os.ModePerm))

	filenames, err := readFileNames(dirName)
	require.NoError(t, err)

	// remove the jbig2imagesdir

	passwords := map[string]string{}

	for _, filename := range filenames {

		t.Run(rawFileName(filename), func(t *testing.T) {
			t.Logf("Getting file: %s", filepath.Join(dirName, filename))
			// get the file
			f, err := getFile(dirName, filename)
			require.NoError(t, err)

			defer f.Close()

			var reader *pdf.PdfReader
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

			w, err := os.Create(filepath.Join(dirName, jbig2ImagesDir, rawFileName(filename)+".zip"))
			require.NoError(t, err)

			defer w.Close()

			zw := zip.NewWriter(w)
			defer zw.Close()

			jbig2w, err := os.Create(filepath.Join(dirName, "jbig2_"+rawFileName(filename)+".zip"))
			require.NoError(t, err)

			defer jbig2w.Close()

			jbigZW := zip.NewWriter(jbig2w)
			defer jbigZW.Close()

			for pageNo := 1; pageNo <= numPages; pageNo++ {
				page, err := reader.GetPage(pageNo)
				require.NoError(t, err)

				images, jbig2Files, err := extractImagesOnPage(filepath.Join(dirName, rawFileName(filename)), page)
				require.NoError(t, err)

				writeImages(t, zw, dirName, filename, pageNo, images...)
				writeJBIG2Files(t, jbigZW, dirName, filename, pageNo, jbig2Files...)
			}

		})
	}

}
