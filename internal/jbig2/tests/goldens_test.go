/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Goldens is a model used to store the jbig2 testcase 'golden files'.
// The golden files stores the md5 'hash' value for each 'filename' key.
// It is used to check if the decoded jbig2 image had changed using it's md5 hash.
type Goldens map[string]string

func checkGoldenFiles(t *testing.T, dirname, filename string, images ...*extractedImage) {
	goldens, err := readGoldenFile(dirname, filename)
	require.NoError(t, err)

	if jbig2UpdateGoldens {
		// copy all the file hashes into Goldens map.
		for _, img := range images {
			goldens[img.fullName()] = img.hash
		}

		err = writeGoldenFile(dirname, filename, goldens)
		require.NoError(t, err)
		return
	}

	for _, img := range images {
		t.Run(fmt.Sprintf("Page#%d/Image#%d", img.pageNo, img.idx), func(t *testing.T) {
			single, exist := goldens[img.fullName()]
			// check if the 'filename' key exists.
			if assert.True(t, exist, "hash doesn't exists") {
				// check if the md5 hash equals with the given fh.hash
				assert.Equal(t, img.hash, single, "hash: '%s' doesn't match the golden stored hash: '%s'", img.hash, single)
			}
		})
	}
}

func readGoldenFile(dirname, filename string) (Goldens, error) {
	// prepare golden files directory name
	goldenDir := filepath.Join(dirname, "goldens")

	// check if the directory exists.
	if _, err := os.Stat(goldenDir); err != nil {
		if err = os.Mkdir(goldenDir, 0700); err != nil {
			return nil, err
		}
		return Goldens{}, nil
	}

	// create if not exists the golden file
	f, err := os.OpenFile(filepath.Join(goldenDir, filename+"_golden.json"), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	goldens := Goldens{}
	err = json.NewDecoder(f).Decode(&goldens)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return goldens, nil
}

func writeGoldenFile(dirname, filename string, goldens Goldens) error {
	// create if not exists the golden file
	f, err := os.Create(filepath.Join(dirname, "goldens", filename+"_golden.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	e := json.NewEncoder(f)
	e.SetIndent("", "\t")
	if err = e.Encode(&goldens); err != nil {
		return err
	}
	return nil
}
