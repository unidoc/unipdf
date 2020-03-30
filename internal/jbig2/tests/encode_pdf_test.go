/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/creator"
)

// TestImageEncodeJBIG2PDF tests the encode process for the JBIG2 encoder into PDF file.
func TestImageEncodeJBIG2PDF(t *testing.T) {
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

	// prepare temporary directory where the jbig2 files would be stored
	tempDir := filepath.Join(os.TempDir(), "unipdf", "jbig2", "encoded-pdf")
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
		case !keepEncodedFile && !logToFile:
			err = os.RemoveAll(filepath.Join(tempDir))
		case !keepEncodedFile:
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
		if !keepEncodedFile {
			os.RemoveAll(tempDir)
		}
	}()

	buf := &bytes.Buffer{}
	h := md5.New()
	edp := []goldenValuePair{}

	for _, fileName := range fileNames {
		var duplicateLinesRemoval bool
		duplicateLinesName := "NoDuplicateLinesRemoval"
		for i := 0; i < 2; i++ {
			if i == 1 {
				duplicateLinesRemoval = true
				duplicateLinesName = duplicateLinesName[2:]
			}
			t.Run(rawFileName(fileName)+duplicateLinesName, func(t *testing.T) {
				// read the file
				c := creator.New()

				img, err := c.NewImageFromFile(filepath.Join(dirName, fileName))
				require.NoError(t, err)

				// conver an image to binary image
				err = img.ToBinaryImage()
				require.NoError(t, err)

				img.ScaleToWidth(612.0)

				e := core.NewJBIG2Encoder()
				if duplicateLinesRemoval {
					e.DefaultPageSettings.DuplicatedLinesRemoval = true
				}
				img.SetEncoder(e)

				// Use page width of 612 points, and calculate the height proportionally based on the image.
				// Standard PPI is 72 points per inch, thus a width of 8.5"
				height := 612.0 * img.Height() / img.Width()
				c.NewPage()
				c.SetPageSize(creator.PageSize{612, height})
				// c.SetPageSize(creator.PageSize{img.Width() * 1.2, img.Height() * 1.2})
				img.SetPos(0, 0)

				err = c.Draw(img)
				require.NoError(t, err)

				err = c.Write(buf)
				require.NoError(t, err)

				_, err = h.Write(buf.Bytes())
				require.NoError(t, err)

				if keepEncodedFile {
					f, err := os.Create(filepath.Join(tempDir, rawFileName(fileName)+duplicateLinesName+".pdf"))
					require.NoError(t, err)
					defer f.Close()

					_, err = f.Write(buf.Bytes())
					require.NoError(t, err)
				}
				hashEncoded := h.Sum(nil)
				buf.Reset()

				edp = append(edp, goldenValuePair{
					Filename: rawFileName(fileName) + duplicateLinesName,
					Hash:     hashEncoded,
				})
			})
		}
	}
	const goldenFileName = "encoded-pdf"
	checkGoldenValuePairs(t, dirName, goldenFileName, edp...)
}
