/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"archive/zip"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// TestGoImageToJBIG2Image tests the core.GoImageToJBIG2Image function.
// This tests requires 'jbig2-store-images' flag. The test images could be found
// within the directory provided in the environment variable: 'UNIDOC_JBIG2_TEST_IMAGES'.
// If no files would be found in the directory the test will be skipped.
func TestGoImageToJBIG2Image(t *testing.T) {
	if !keepImageFiles {
		t.Skip("flag -jbig2-store-images not set")
	}

	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	dirName := os.Getenv(EnvImageDirectory)
	if dirName == "" {
		t.Skipf("no environment variable: '%s' provided", EnvImageDirectory)
	}

	// get the file names within given directory
	fileNames, err := readFileNames(dirName, "")
	require.NoError(t, err)

	if len(fileNames) == 0 {
		t.Skipf("no files found in the '%s' directory", dirName)
	}

	// prepare temporary directory
	tempDir := filepath.Join(os.TempDir(), "unipdf", "jbig2", "images")
	err = os.MkdirAll(tempDir, 0700)
	require.NoError(t, err)

	thresholds := []float64{0.0, 0.25, 0.5, 0.75}
	names := []string{"auto", "63", "127", "191"}
	// iterate over all files
	for _, fileName := range fileNames {
		f, err := getFile(dirName, fileName)
		require.NoError(t, err)
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			common.Log.Debug("File: '%s' couldn't be read as an image")
			continue
		}

		rawName := rawFileName(fileName)

		t.Run(rawName, func(t *testing.T) {
			zipFileName := filepath.Join(tempDir, rawName+".zip")
			// create zip file containing encoded images
			zf, err := os.Create(zipFileName)
			require.NoError(t, err)
			defer zf.Close()

			zw := zip.NewWriter(zf)
			defer zw.Close()

			for i, th := range thresholds {
				t.Run(names[i], func(t *testing.T) {
					jb2, err := core.GoImageToJBIG2(img, th)
					require.NoError(t, err)

					// convert the image to black/white
					bwImage, err := jb2.ToGoImage()
					require.NoError(t, err)

					// get the destination file name
					df, err := zw.Create(names[i] + ".jpg")
					require.NoError(t, err)

					err = jpeg.Encode(df, bwImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
					require.NoError(t, err)
				})
			}
		})
	}
}
