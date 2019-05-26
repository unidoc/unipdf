/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"testing"

	"github.com/unidoc/unipdf/v3/common"
)

func init() {
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
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
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)

	obj, err := parser.ParseIndirectObject()
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

	bdec, err := DecodeStream(stream)
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

// Tests a stream with multi encoded.
func TestMultiEncodedStream(t *testing.T) {
	// 2 rows of data, 3 colors, 2 columns per row
	encoded := []byte("78 9C 2A C9 C8 2C 56 00 A2 44 85 94 D2 DC DC 4A 85 92 D4 8A 12 85 F2 CC 92 0C 85 E2 FC DC 54 05 46 26 66 85 A4 CC " +
		"BC C4 A2 4A 85 94 C4 92 44 40 00 00 00 FF FF 78 87 0F 9C >")

	expected := []byte("this is a dummy text with some \x01\x02\x03 binary data")

	common.Log.Debug("Compressed length: %d", len(encoded))

	rawText := `99 0 obj
<<
/DecodeParms << /Predictor 1 /Colors 3 /Columns 2 >>
/Filter [/ASCIIHexDecode /FlateDecode]
/Length ` + fmt.Sprintf("%d", len(encoded)) + `
>>
stream
` + string(encoded) + `endstream
endobj`

	parser := PdfParser{}
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)

	obj, err := parser.ParseIndirectObject()
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

	bdec, err := DecodeStream(stream)
	if err != nil {
		t.Errorf("Failed to decode stream (%s)", err)
		return
	}

	common.Log.Debug("Stream: %s", stream.Stream)
	common.Log.Debug("Decoded stream: % x", bdec)

	if !compareSlices(bdec, expected) {
		common.Log.Debug("Expected: % x", expected)
		t.Errorf("decoded != expected")
		return
	}

}
