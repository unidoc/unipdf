/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package strutils

import (
	"testing"
)

var utf16enc = []byte{
	/* 0xfe, 0xff, */ 0x00, 0x62, 0x00, 0x75, 0x00, 0x74, 0x00, 0x74, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x41,
	0x00, 0x72, 0x00, 0x65, 0x00, 0x61, 0x00, 0x53, 0x00, 0x75, 0x00, 0x62, 0x00, 0x66, 0x00, 0x6f,
	0x00, 0x72, 0x00, 0x6d, 0x00, 0x5b, 0x00, 0x30, 0x00, 0x5d,
}

func TestUTF16Encoding(t *testing.T) {
	b := utf16enc
	exp := "buttonAreaSubform[0]"
	v := UTF16ToString(b)

	if v != exp {
		t.Errorf("'%s' != '%s'\n", v, exp)
	}
}

func TestUTF16EncodeDecode(t *testing.T) {
	testcases := []string{"þráður321", "áþðurfyrr \n", "⌘⌘⌘⺃⺓$", "€€$£"}

	for _, tcase := range testcases {
		encoded := StringToUTF16(tcase)
		decoded := UTF16ToString([]byte(encoded))
		if decoded != tcase {
			t.Fatalf("'% X' != '% X'\n", decoded, tcase)
		}
	}
}
