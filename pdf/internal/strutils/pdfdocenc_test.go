package strutils

import (
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
	v := []byte{0x47, 0x65, 0x72, 0xfe, 0x72, 0xfa, 0xf0, 0x75, 0x72}

	for _, testcase := range testcases {
		str := PDFDocEncodingToString(testcase.Encoded)
		if str != testcase.Expected {
			t.Fatalf("Mismatch %s != %s", str, testcase.Expected)
		}
	}

	return
}
