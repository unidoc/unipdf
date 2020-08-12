/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParagraphWithWrappedText(t *testing.T) {
	c := New()
	p := c.NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetPos(0, 0)
	p.SetWidth(100)
	p.wrapText()
	assert.Equal(t, 29, len(p.textLines))
	assert.Equal(t, "Lorem ipsum dolor sit", p.textLines[0])
}

func TestParagraphWithWrappedTextAndMaxWrapLines(t *testing.T) {
	c := New()
	p := c.NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetPos(0, 0)
	p.SetWidth(100)
	p.SetMaxWrapLines(10)
	p.wrapText()
	assert.Equal(t, 10, len(p.textLines))
	assert.Equal(t, "Lorem ipsum dolor sit", p.textLines[0])
}

func TestParagraphWithWrappedTextAndLinesDoNotExceedMaxWrapLines(t *testing.T) {
	c := New()
	p := c.NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetPos(0, 0)
	p.SetWidth(100)
	p.SetMaxWrapLines(100)
	p.wrapText()
	assert.Equal(t, 29, len(p.textLines))
	assert.Equal(t, "Lorem ipsum dolor sit", p.textLines[0])
}

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
