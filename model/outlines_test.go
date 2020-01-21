/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetOutlines(t *testing.T) {
	// Open reader.
	f, err := os.Open(`./testdata/pages3.pdf`)
	require.NoError(t, err)

	reader, err := NewPdfReader(f)
	require.NoError(t, err)
	defer f.Close()

	pageNum, err := reader.GetNumPages()
	require.NoError(t, err)

	// Add reader pages to writer.
	writer := NewPdfWriter()
	for i := 1; i < pageNum; i++ {
		page, err := reader.GetPage(i)
		require.NoError(t, err)

		writer.AddPage(page)
	}

	// Generate outline.
	srcOutline := NewOutline()
	for i := 0; i < 3; i++ {
		item := NewOutlineItem(fmt.Sprintf("Outline %d", i+1),
			NewOutlineDest(int64(i), float64(i), float64(i)))
		srcOutline.Add(item)

		for j := 0; j < i; j++ {
			childItem := NewOutlineItem(fmt.Sprintf("%s.%d", item.Title, j+1),
				NewOutlineDest(int64(i), float64(i*j), float64(i*j)))
			item.Add(childItem)
			item = childItem
		}
	}
	writer.AddOutlineTree(srcOutline.ToOutlineTree())

	// Write file to buffer.
	var buf bytes.Buffer
	err = writer.Write(&buf)
	require.NoError(t, err)

	// Read file from buffer.
	reader, err = NewPdfReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)

	// Compare generated outline with the one read from the buffer.
	dstOutline, err := reader.GetOutlines()
	require.NoError(t, err)

	srcJson, err := json.Marshal(srcOutline)
	require.NoError(t, err)
	dstJson, err := json.Marshal(dstOutline)
	require.NoError(t, err)
	require.Equal(t, srcJson, dstJson)
}
