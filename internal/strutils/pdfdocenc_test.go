/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package strutils

import (
	"bytes"
	"testing"
)

func TestPDFDocEncodingDecode(t *testing.T) {
	testcases := []struct {
		Encoded  []byte
		Expected string
	}{
		{[]byte{0x47, 0x65, 0x72, 0xfe, 0x72, 0xfa, 0xf0, 0x75, 0x72}, "Gerþrúður"},
		{[]byte("Ger\xfer\xfa\xf0ur"), "Gerþrúður"},
	}

	for _, testcase := range testcases {
		str := PDFDocEncodingToString(testcase.Encoded)
		if str != testcase.Expected {
			t.Fatalf("Mismatch %s != %s", str, testcase.Expected)
		}

		enc := StringToPDFDocEncoding(str)
		if !bytes.Equal(enc, testcase.Encoded) {
			t.Fatalf("Encode mismatch %s (%X) != %s (%X)", enc, enc, testcase.Encoded, testcase.Encoded)
		}
	}

	return
}
