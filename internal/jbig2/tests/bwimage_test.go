/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// TestGoImageToJBIG2Image tests the core.GoImageToJBIG2Image function.
// This tests requires 'jbig2-store-images' flag. The test images could be found
// within the directory provided in the environment variable: 'UNIDOC_JBIG2_TEST_IMAGES'.
// If no files would be found in the directory the test will be skipped.
func TestGoImageToJBIG2Image(t *testing.T) {
	dirName := os.Getenv(EnvImageDirectory)
	if dirName == "" {
		t.Skipf("no environment variable: '%s' provided", EnvImageDirectory)
	}

	// get the file names within given directory
	fileNames, err := readFileNames(dirName, "jpg")
	require.NoError(t, err)

	if len(fileNames) == 0 {
		t.Skipf("no files found in the '%s' directory", dirName)
	}

	// prepare temporary directory
	tempDir := filepath.Join(os.TempDir(), "unipdf", "jbig2", "images")
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

	defer func() {
		if !keepImageFiles {
			os.RemoveAll(tempDir)
		}
	}()

	thresholds := []float64{0.0, 0.25, 0.5, 0.75}
	names := []string{"auto", "63", "127", "191"}

	buf := &bytes.Buffer{}
	h := md5.New()
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

			gvp := []goldenValuePair{}
			for i, th := range thresholds {
				t.Run(names[i], func(t *testing.T) {
					jb2, err := core.GoImageToJBIG2(img, th)
					require.NoError(t, err)

					// convert the image to black/white
					bwImage, err := jb2.ToGoImage()
					require.NoError(t, err)

					buf.Reset()

					err = jpeg.Encode(buf, bwImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
					require.NoError(t, err)

					h.Reset()
					_, err = h.Write(buf.Bytes())
					require.NoError(t, err)

					// get the destination file name
					gvp = append(gvp, goldenValuePair{
						Filename: names[i],
						Hash:     h.Sum(nil),
					})

					if keepImageFiles {
						df, err := zw.Create(names[i] + ".jpg")
						require.NoError(t, err)

						_, err = buf.WriteTo(df)
						require.NoError(t, err)
					}
				})
			}
			checkGoldenValuePairs(t, dirName, rawName, gvp...)
		})
	}
}
