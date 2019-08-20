package model

import (
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unipdf/v3/core"
	"testing"
)

func TestUrlFileSpec(t *testing.T) {
	rawText := `
1 0 obj
<< /FS /URL
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
}
