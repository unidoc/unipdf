/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package strutils provides convenient functions for string processing in unidoc internally.
package strutils

import (
	"bytes"
	"unicode/utf16"

	"github.com/unidoc/unipdf/v3/common"
)

var pdfdocEncodingRuneMap map[rune]byte

func init() {
	pdfdocEncodingRuneMap = map[rune]byte{}
	for b, r := range pdfDocEncoding {
		pdfdocEncodingRuneMap[r] = b
	}
}

// UTF16ToRunes decodes the UTF-16BE encoded byte slice `b` to unicode runes.
func UTF16ToRunes(b []byte) []rune {
	if len(b) == 1 {
		return []rune{rune(b[0])}
	}
	if len(b)%2 != 0 {
		b = append(b, 0)
		common.Log.Debug("ERROR: UTF16ToRunes. Padding with zeros.")
	}
	n := len(b) >> 1
	chars := make([]uint16, n)
	for i := 0; i < n; i++ {
		chars[i] = uint16(b[i<<1])<<8 + uint16(b[i<<1+1])
	}
	runes := utf16.Decode(chars)
	return runes
}

// UTF16ToString decodes the UTF-16BE encoded byte slice `b` to a unicode go string.
func UTF16ToString(b []byte) string {
	return string(UTF16ToRunes(b))
}

// StringToUTF16 encoded `s` to UTF16 and returns a string containing UTF16 runes.
func StringToUTF16(s string) string {
	encoded := utf16.Encode([]rune(s))

	var buf bytes.Buffer
	for _, code := range encoded {
		buf.WriteByte(byte((code >> 8) & 0xff))
		buf.WriteByte(byte(code & 0xff))
	}

	return buf.String()
}

// PDFDocEncodingToRunes decodes PDFDocEncoded byte slice `b` to unicode runes.
func PDFDocEncodingToRunes(b []byte) []rune {
	var runes []rune
	for _, bval := range b {
		rune, has := pdfDocEncoding[bval]
		if !has {
			common.Log.Debug("Error: PDFDocEncoding input mapping error %d - skipping", bval)
			continue
		}

		runes = append(runes, rune)
	}

	return runes
}

// PDFDocEncodingToString decodes PDFDocEncoded byte slice `b` to unicode go string.
func PDFDocEncodingToString(b []byte) string {
	return string(PDFDocEncodingToRunes(b))
}

// StringToPDFDocEncoding encoded go string `s` to PdfDocEncoding.
func StringToPDFDocEncoding(s string) []byte {
	var buf bytes.Buffer
	for _, r := range s {
		b, has := pdfdocEncodingRuneMap[r]
		if !has {
			common.Log.Debug("ERROR: PDFDocEncoding rune mapping missing %c/%X - skipping", r, r)
			continue
		}
		buf.WriteByte(b)
	}

	return buf.Bytes()
}
