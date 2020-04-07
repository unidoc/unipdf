/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/model/optimize"
)

var (
	envOptimizePDFFiles = "UNIDOC_OPTIMIZE_TESTDATA"
	// updateGoldens is the runtime flag that states that the md5 hashes
	// for each decoded test case image should be updated.
	updateGoldens      bool
	keepOptimizedFiles bool
)

func init() {
	flag.BoolVar(&updateGoldens, "optimize-update-goldens", false, "updates the golden file hashes on the run")
	flag.BoolVar(&keepOptimizedFiles, "keep-optimized-files", false, "stores optimized files in the temp directory")
}

func TestImagePPIOptimize(t *testing.T) {
	dirName := os.Getenv(envOptimizePDFFiles)
	if dirName == "" {
		t.Skipf("No env: '%s' provided", envOptimizePDFFiles)
	}

	filenames, err := readFileNames(dirName, ".pdf")
	require.NoError(t, err)

	tempDir := filepath.Join(os.TempDir(), "unipdf", "optimizer")
	err = os.MkdirAll(tempDir, 0700)
	require.NoError(t, err)

	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	h := md5.New()

	goldens := []goldenValuePair{}
	for _, filename := range filenames {
		rawName := rawFileName(filename)
		t.Run(rawName, func(t *testing.T) {
			var closers []io.Closer
			f, err := os.Open(filepath.Join(dirName, filename))
			require.NoError(t, err)

			closers = append(closers, f)

			r, err := readPDF(f, "")
			require.NoError(t, err)

			w := model.NewPdfWriter()
			err = readerToWriter(r, &w, nil)
			require.NoError(t, err)

			w.SetOptimizer(optimize.New(optimize.Options{
				CombineDuplicateDirectObjects:   true,
				CombineIdenticalIndirectObjects: true,
				CombineDuplicateStreams:         true,
				CompressStreams:                 true,
				UseObjectStreams:                true,
				ImageQuality:                    100,
				ImageUpperPPI:                   100,
			}))

			writers := []io.Writer{h}
			if keepOptimizedFiles {
				f, err := os.Create(filepath.Join(tempDir, filename+"_optimized.pdf"))
				require.NoError(t, err)

				closers = append(closers, f)
				writers = append(writers, f)
			}

			err = w.Write(io.MultiWriter(writers...))
			require.NoError(t, err)

			hashEncoded := hex.EncodeToString(h.Sum(nil))
			h.Reset()

			goldens = append(goldens, goldenValuePair{
				Filename: filename,
				Hash:     []byte(hashEncoded),
			})

			for _, closer := range closers {
				closer.Close()
			}
		})
	}
	checkGoldenValuePairs(t, dirName, "optimized-goldens", goldens...)
}

// Goldens is a model used to store the jbig2 test case 'golden files'.
// The golden files stores the md5 'hash' value for each 'filename' key.
// It is used to check if the decoded jbig2 image had changed using it's md5 hash.
type Goldens map[string]string

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

type goldenValuePair struct {
	Filename string
	Hash     []byte
}

func checkGoldenValuePairs(t *testing.T, dirname, goldenFileName string, results ...goldenValuePair) {
	goldens, err := readGoldenFile(dirname, goldenFileName)
	require.NoError(t, err)

	if updateGoldens {
		for _, result := range results {
			goldens[result.Filename] = hex.EncodeToString(result.Hash)
		}
		err = writeGoldenFile(dirname, goldenFileName, goldens)
		require.NoError(t, err)
		return
	}

	for _, result := range results {
		t.Run(fmt.Sprintf("%s/Golden", result.Filename), func(t *testing.T) {
			goldenValue, exist := goldens[result.Filename]
			if assert.True(t, exist, "hash doesn't exists") {
				// check if the md5 hash equals with the given fh.hash
				hexValue := hex.EncodeToString(result.Hash)
				assert.Equal(t, goldenValue, hexValue, "hash: '%s' doesn't match the golden stored hash: '%s'", hexValue, goldenValue)
			}
		})
	}
}

func readPDF(f *os.File, password ...string) (*model.PdfReader, error) {
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, err
	}

	// check if is encrypted
	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}

	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, err
		}

		if !auth {
			if len(password) > 0 {
				auth, err = pdfReader.Decrypt([]byte(password[0]))
				if err != nil {
					return nil, err
				}
			}
			if !auth {
				return nil, fmt.Errorf("reading the file: '%s' failed. Invalid password provided", f.Name())
			}
		}
	}
	return pdfReader, nil
}

func readFileNames(dirname, suffix string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if suffix != "" && !strings.HasSuffix(strings.ToLower(info.Name()), suffix) {
				return nil
			}
			files = append(files, info.Name())
		}
		return nil
	})
	return files, err
}

func rawFileName(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func readerToWriter(r *model.PdfReader, w *model.PdfWriter, pages []int) error {
	if r == nil {
		return errors.New("source PDF cannot be null")
	}
	if w == nil {
		return errors.New("destination PDF cannot be null")
	}

	// Get number of pages.
	pageCount, err := r.GetNumPages()
	if err != nil {
		return err
	}

	// Add optional properties
	if ocProps, err := r.GetOCProperties(); err == nil {
		if err = w.SetOCProperties(ocProps); err != nil {
			return err
		}
	}

	// Add pages.
	if len(pages) == 0 {
		pages = createPageRange(pageCount)
	}

	for _, numPage := range pages {
		if numPage < 1 || numPage > pageCount {
			continue
		}

		page, err := r.GetPage(numPage)
		if err != nil {
			return err
		}

		if err = w.AddPage(page); err != nil {
			return err
		}
	}

	// Add forms.
	if r.AcroForm != nil {
		if err = w.SetForms(r.AcroForm); err != nil {
			return err
		}
	}

	return nil
}

func createPageRange(count int) []int {
	if count <= 0 {
		return []int{}
	}

	var pages []int
	for i := 0; i < count; i++ {
		pages = append(pages, i+1)
	}

	return pages
}
