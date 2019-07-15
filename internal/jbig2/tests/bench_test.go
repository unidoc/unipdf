/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2"
)

// BenchmarkDecodeJBIG2Files benchmarks the decoding process of  jbig2 encoded images stored within pdf files.
// The function reads pdf files located in the directory provided as `UNIDOC_JBIG2_TESTDATA` environmental variable.
// Then the function extracts the images and starts subBenchmarks for each image.
func BenchmarkDecodeJBIG2Files(b *testing.B) {
	b.Helper()
	dirName := os.Getenv(EnvDirectory)
	require.NotEmpty(b, dirName, "No Environment variable 'UNIDOC_JBIG2_TESTDATA' found")

	filenames, err := readFileNames(dirName)
	require.NoError(b, err)
	require.NotEmpty(b, filenames, "no files found within provided directory")

	for _, filename := range filenames {
		b.Run(rawFileName(filename), func(b *testing.B) {
			images, err := extractImages(dirName, filename)
			require.NoError(b, err)

			for _, image := range images {
				b.Run(fmt.Sprintf("Page#%d/Image#%d-%d", image.pageNo, image.idx, len(image.jbig2Data)), func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						d, err := jbig2.NewDocumentWithGlobals(image.jbig2Data, image.globals)
						require.NoError(b, err)

						p, err := d.GetPage(1)
						require.NoError(b, err)

						_, err = p.GetBitmap()
						require.NoError(b, err)
					}
				})
			}
		})
	}
}
