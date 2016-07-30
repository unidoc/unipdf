/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"bytes"
	"compress/zlib"
	"testing"

	"github.com/unidoc/unidoc/common"
)

func init() {
	common.SetLogger(common.ConsoleLogger{})
}

// Compare the equality of content of two slices.
func compareSlices(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// This tests the TIFF predictor (Predictor 2) for PDF.
// Passes the test, but seems not to work on certain PDF files.
func TestFlateTiffPredictor(t *testing.T) {
	// 2 rows of data, 3 colors, 2 columns per row
	rawStream := []byte("\x01\x02\x01\x00\x03\x04\x05\xff\x01\xaf\x01\x02")
	expected := []byte("" +
		"\x01\x02\x01\x01\x05\x05" +
		"\x05\xff\x01\xb4\x00\x03")
	// \x01\x02\x01
	// \x00\x03\x04
	// \x05\xff\x01
	// \xaf\x01\x02

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(rawStream)
	w.Close()

	encoded := b.Bytes()
	common.Log.Debug("Compressed length: %d", len(encoded))

	rawText := `99 0 obj
<<
/DecodeParms << /Predictor 2
                /Colors 3
                /Columns 2
             >>
/Filter /FlateDecode
/Length 24
>>
stream
` + string(encoded) + `endstream
endobj`

	parser := PdfParser{}
	parser.reader = makeReaderForText(rawText)

	obj, err := parser.parseIndirectObject()
	if err != nil {
		t.Errorf("Invalid stream object (%s)", err)
		return
	}

	stream, ok := obj.(*PdfObjectStream)
	if !ok {
		t.Errorf("Not a valid pdf stream")
		return
	}

	common.Log.Debug("%q", stream)
	dict := stream.PdfObjectDictionary
	common.Log.Debug("dict: %q", dict)

	if len(stream.Stream) != len(encoded) {
		t.Errorf("Length not %d (%d)", len(encoded), len(stream.Stream))
		return
	}

	bdec, err := parser.decodeStream(stream)
	if err != nil {
		t.Errorf("Failed to decode stream (%s)", err)
		return
	}

	common.Log.Debug("Orig stream: % x\n", stream.Stream)
	common.Log.Debug("Decoded stream: % x\n", bdec)
	if !compareSlices(bdec, expected) {
		common.Log.Debug("Expected: % x\n", expected)
		t.Errorf("decoded != expected")
		return
	}
}
