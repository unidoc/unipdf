/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/core"
)

func TestReaderLazy(t *testing.T) {
	f, err := os.Open(`./testdata/minimal.pdf`)
	require.NoError(t, err)

	reader, err := NewPdfReaderLazy(f)
	require.NoError(t, err)
	defer f.Close()

	require.Equal(t, 1, len(reader.PageList))

	page, err := reader.GetPage(1)
	require.NoError(t, err)

	ref, isRef := page.Contents.(*core.PdfObjectReference)
	require.True(t, isRef)

	obj := ref.Resolve()
	_, isStream := obj.(*core.PdfObjectStream)
	require.True(t, isStream)

	str, err := page.GetAllContentStreams()
	require.NoError(t, err)
	require.Equal(t, 55, len(str))

	var buf bytes.Buffer
	writer := NewPdfWriter()
	err = writer.AddPage(page)
	require.NoError(t, err)

	err = writer.Write(&buf)
	require.NoError(t, err)
}
