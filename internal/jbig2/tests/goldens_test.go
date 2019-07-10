/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Goldens is a model for goldens file json document.
type Goldens map[string]*GoldenObject

// GoldenObject is a row of the golden file.
type GoldenObject struct {
	IsValid          bool   `json:"is_valid"`
	Hash             string `json:"hash"`
	MatchBitmapImage bool   `json:"match_bitmap_image"`
	IsBitmapValid    bool   `json:"is_bitmap_valid"`
}

func checkGoldenFiles(t testing.TB, dirname, filename string, readHashes ...fileHash) {
	goldens, err := readGoldenFile(dirname, filename)
	require.NoError(t, err)

	if len(goldens) == 0 {
		// copy all the file hashes into Goldens map.
		for _, fh := range readHashes {
			row := &GoldenObject{Hash: fh.hash}
			goldens[fh.fileName] = row
		}

		err = writeGoldenFile(dirname, filename, goldens)
		require.NoError(t, err)
		return
	}

	for _, fh := range readHashes {
		single, exist := goldens[fh.fileName]
		switch {
		case !exist:
			goldens[fh.fileName] = &GoldenObject{Hash: fh.hash}
			continue
		case single.IsValid:
			continue
		}

		// if the single raw is not valid then udate it's hash
		single.Hash = fh.hash
	}
	err = writeGoldenFile(dirname, filename, goldens)
	require.NoError(t, err)
}

func readGoldenFile(dirname, filename string) (Goldens, error) {
	// prepare golden files directory name
	goldenDir := filepath.Join(dirname, "goldens")

	// check if the directory exists.
	if _, err := os.Stat(goldenDir); err != nil {
		if err = os.Mkdir(goldenDir, 0666); err != nil {
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
