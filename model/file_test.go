/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unipdf/v3/core"
	"strings"
	"testing"
)

func TestUrlFileSpec(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Filespec
/FS /URL
/F (ftp://www.beatles.com/Movies/AbbeyRoad.mov)
>>
endobj
`

	r := NewReaderForText(rawText)

	err := r.ParseIndObjSeries()
	require.NoError(t, err)

	// Load the field from object number 1.
	obj, err := r.parser.LookupByNumber(1)
	require.NoError(t, err)

	ind, ok := obj.(*core.PdfIndirectObject)
	require.True(t, ok)

	fileSpec, err := NewPdfFilespecFromObj(ind)
	require.NoError(t, err)

	require.Equal(t, "URL", fileSpec.FS.String())
	require.Equal(t, "ftp://www.beatles.com/Movies/AbbeyRoad.mov", fileSpec.F.String())

	outDict, ok := core.GetDict(fileSpec.ToPdfObject())
	if !ok {
		t.Fatalf("error")
	}

	contains := strings.Contains(
		strings.Replace(rawText, "\n", "", -1),
		outDict.WriteString())
	require.True(t, contains, "generated output doesn't match the expected output")
}

func TestPdfFileSpec(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Filespec
/F (VideoIssue1.mov)
/UF (VideoIssue2.mov)
/DOS (VIDEOISSUE.MOV)
/Mac (VideoIssue3.mov)
/Unix (VideoIssue4.mov)
>>
endobj
`

	r := NewReaderForText(rawText)

	err := r.ParseIndObjSeries()
	require.NoError(t, err)

	// Load the field from object number 1.
	obj, err := r.parser.LookupByNumber(1)
	require.NoError(t, err)

	ind, ok := obj.(*core.PdfIndirectObject)
	require.True(t, ok)

	fileSpec, err := NewPdfFilespecFromObj(ind)
	require.NoError(t, err)

	require.Equal(t, "VideoIssue1.mov", fileSpec.F.String())
	require.Equal(t, "VideoIssue2.mov", fileSpec.UF.String())
	require.Equal(t, "VIDEOISSUE.MOV", fileSpec.DOS.String())
	require.Equal(t, "VideoIssue3.mov", fileSpec.Mac.String())
	require.Equal(t, "VideoIssue4.mov", fileSpec.Unix.String())

	outDict, ok := core.GetDict(fileSpec.ToPdfObject())
	if !ok {
		t.Fatalf("error")
	}

	contains := strings.Contains(
		strings.Replace(rawText, "\n", "", -1),
		outDict.WriteString())

	require.True(t, contains, "generated output doesn't match the expected output")
}
