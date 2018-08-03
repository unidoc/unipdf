/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package strutils provides convenient functions for string processing in unidoc internally.
package strutils

import (
	"unicode/utf16"

	"github.com/unidoc/unidoc/common"
)

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

// PDFDocEncodingToRunes decodes PDFDocEncoded byte slice `b` to unicode runes.
func PDFDocEncodingToRunes(b []byte) []rune {
	runes := []rune{}
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
