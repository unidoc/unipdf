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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// TestImageEncodeDecodeJBIG2 tests the encode and decode process for the JBIG2 encoder.
func TestImageEncodeDecodeJBIG2(t *testing.T) {
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

	// prepare temporary directory where the jbig2 files would be stored
	tempDir := filepath.Join(os.TempDir(), "unipdf", "jbig2", "encoded-decoded")
	err = os.MkdirAll(tempDir, 0700)
	require.NoError(t, err)

	defer func() {
		if !keepImageFiles {
			os.RemoveAll(tempDir)
		}
	}()

	for _, fileName := range fileNames {
		// read the file
		f, err := getFile(dirName, fileName)
		require.NoError(t, err)
		defer f.Close()

		// try to read the file as image.
		img, _, err := image.Decode(f)
		if err != nil {
			// if the image is of unknown decoding or is not an image skip the test.
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

			// wrap zip writer over the file writer.
			zw := zip.NewWriter(zf)
			defer zw.Close()

			// convert the input image into jbig2 1bpp acceptable binary image.
			jimg, err := core.GoImageToJBIG2(img, 0.0)
			require.NoError(t, err)

			// create the encoder
			e := &core.JBIG2Encoder{}
			e.FileMode = true
			err = e.AddPageImage(jimg, core.JBIG2PageSettings{DuplicatedLinesRemoval: true})
			require.NoError(t, err)

			data, err := e.Encode()
			require.NoError(t, err)

			// create .jbig2 file within the zip file
			jbf, err := zw.Create(rawName + ".jbig2")
			require.NoError(t, err)

			// write encoded data into jb2 file
			n, err := jbf.Write(data)
			require.NoError(t, err)
			assert.Equal(t, n, len(data))

			// create a jpeg file to show the black white image.
			df, err := zw.Create(rawName + ".jpeg")
			require.NoError(t, err)

			// create golang image and store it within zip file
			bwImage, err := jimg.ToGoImage()
			require.NoError(t, err)

			// store the binary image in the 'jpeg' format.
			err = jpeg.Encode(df, bwImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
			require.NoError(t, err)

			// decode the encoded data and store it's results in the zipped file.
			d := &core.JBIG2Encoder{}
			decoded, err := d.DecodeImages(data)
			require.NoError(t, err)
			require.Len(t, decoded, 1)

			// create decoded image file within the zip file.
			dimg, err := zw.Create(rawName + "_decoded.jpeg")
			require.NoError(t, err)

			// write the decoded image
			err = jpeg.Encode(dimg, decoded[0], &jpeg.Options{Quality: jpeg.DefaultQuality})
			require.NoError(t, err)
		})
	}
}
