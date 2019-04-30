/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func benchmarkParagraphAdding(b *testing.B, loops int) {
	for n := 0; n < b.N; n++ {
		c := New()
		for i := 0; i < loops; i++ {
			p := c.NewParagraph(fmt.Sprintf("Paragraph %d - %d", n, i))
			err := c.Draw(p)
			require.NoError(b, err)
		}

		var buf bytes.Buffer
		err := c.Write(&buf)
		require.NoError(b, err)
	}
}

// Benchmark adding multiple paragraphs and writing out.
// Check the effects of varying number of paragraphs written out.
func BenchmarkParagraphAdding1(b *testing.B)   { benchmarkParagraphAdding(b, 1) }
func BenchmarkParagraphAdding10(b *testing.B)  { benchmarkParagraphAdding(b, 10) }
func BenchmarkParagraphAdding100(b *testing.B) { benchmarkParagraphAdding(b, 100) }
