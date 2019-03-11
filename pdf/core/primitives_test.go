/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"strconv"
	"strings"
	"testing"
)

func TestHexStringWriteBasic(t *testing.T) {
	testcases := map[string]string{
		" ": "<20>",
	}

	for src, expected := range testcases {
		strObj := MakeHexString(src)
		ws := strObj.WriteString()

		if ws != expected {
			t.Fatalf("%s: '%s' != '%s'\n", src, ws, expected)
		}
	}
}

// Test writing and parsing back of hexadecimal and regular strings.
func TestHexStringMulti(t *testing.T) {
	testcases := []string{
		"This is a string",
		"Strings may contain\n newlines and such",
		string([]byte{0x50, 0x01, 0x00, 0x90, 0xff, 0x49, 0xdf, 0x20, 0x32}),
		"",
	}

	for _, testcase := range testcases {
		// Make *PdfObject representations for regular and hexadecimal strings.
		s := MakeString(testcase)
		shex := MakeHexString(testcase)

		// Write out.
		writestr := s.WriteString()
		writestrhex := shex.WriteString()

		// Parse back.
		parser1 := makeParserForText(writestr)
		parser2 := makeParserForText(writestrhex)

		// Check that representation is correct.
		obj1, err := parser1.parseObject()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		strObj1, ok := obj1.(*PdfObjectString)
		if !ok {
			t.Fatalf("Type incorrect")
		}
		if strObj1.isHex != false {
			t.Fatalf("Should not be hex")
		}
		if strObj1.Str() != testcase {
			t.Fatalf("String mismatch")
		}

		obj2, err := parser2.parseObject()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		strObj2, ok := obj2.(*PdfObjectString)
		if !ok {
			t.Fatalf("Type incorrect")
		}
		if strObj2.isHex != true {
			t.Fatalf("Should be hex")
		}
		if strObj2.Str() != testcase {
			t.Fatalf("String mismatch")
		}
	}
}

func TestPdfDocEncodingDecode(t *testing.T) {
	testcases := []struct {
		Encoded  PdfObjectString
		Expected string
	}{
		{PdfObjectString{val: "Ger\xfer\xfa\xf0ur", isHex: false}, "Gerþrúður"},
	}

	for _, testcase := range testcases {
		dec := testcase.Encoded.Decoded()
		if dec != testcase.Expected {
			t.Fatalf("%s != %s", dec, testcase.Expected)
		}

		str := MakeEncodedString(dec, false)
		if str.Decoded() != dec {
			t.Fatalf("%s (%X) != %s (%X)", str.Decoded(), str.Decoded(), dec, dec)
		}
	}
}

func TestUTF16StringEncodeDecode(t *testing.T) {
	testcases := []string{"漢字", `Testing «ταБЬℓσ»: 1<2 & 4+1>3, now 20% off!`}

	for _, tc := range testcases {
		// UTF16-BE.
		str := MakeEncodedString(tc, true)
		if str.Decoded() != tc {
			t.Fatalf("% X != % X (%s)", str.Decoded(), tc, tc)
		}

	}
}

func BenchmarkPdfObjectIntegerWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		i := MakeInteger(int64(n))
		i.WriteString()
	}
}

func BenchmarkPdfObjectFloatWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		f := MakeFloat(float64(n))
		f.WriteString()
	}
}

func BenchmarkPdfObjectStringWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := MakeString(strings.Repeat("a", n))
		s.WriteString()
	}
}

func BenchmarkPdfObjectNameWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		c := n
		if n > 127 {
			c = 127
		}

		o := MakeString(strings.Repeat("z", c))
		o.WriteString()
	}
}

func BenchmarkPdfObjectArrayWriteString(b *testing.B) {
	a := MakeArray()
	for i := 0; i < 10000; i++ {
		a.Append(MakeInteger(int64(i)))
	}

	for n := 0; n < b.N; n++ {
		a.WriteString()
	}
}

func BenchmarkPdfObjectDictionaryWriteString(b *testing.B) {
	a := MakeDict()
	for i := 0; i < 10000; i++ {
		o := MakeName(strconv.Itoa(i))
		a.Set(*o, MakeInteger(int64(i)))
	}

	for n := 0; n < b.N; n++ {
		a.WriteString()
	}
}

func BenchmarkPdfObjectReferenceWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		r := &PdfObjectReference{
			ObjectNumber:     int64(n),
			GenerationNumber: int64(n),
		}
		r.WriteString()
	}
}

func BenchmarkPdfIndirectObjectWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		i := &PdfIndirectObject{}
		i.ObjectNumber = int64(n)
		i.WriteString()
	}
}

func BenchmarkPdfObjectStreamWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := &PdfObjectStream{}
		s.ObjectNumber = int64(n)
		s.WriteString()
	}
}

func BenchmarkPdfObjectStreamsWriteString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := &PdfObjectStreams{}
		s.ObjectNumber = int64(n)
		s.WriteString()
	}
}
