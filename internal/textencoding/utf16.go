/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/strutils"
)

// UTF16Encoder represents UTF-16 encoding.
type UTF16Encoder struct {
	baseName string
}

// NewUTF16TextEncoder returns a new UTF16Encoder based on the predefined
// encoding `baseName`.
func NewUTF16TextEncoder(baseName string) UTF16Encoder {
	return UTF16Encoder{baseName}
}

// String returns a string that describes `enc`.
func (enc UTF16Encoder) String() string {
	return enc.baseName
}

// Encode converts the Go unicode string to a PDF encoded string.
func (enc UTF16Encoder) Encode(str string) []byte {
	return []byte(strutils.StringToUTF16(str))
}

// Decode converts PDF encoded string to a Go unicode string.
func (enc UTF16Encoder) Decode(raw []byte) string {
	return strutils.UTF16ToString(raw)
}

// RuneToCharcode converts rune `r` to a PDF character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc UTF16Encoder) RuneToCharcode(r rune) (CharCode, bool) {
	return CharCode(r), true
}

// CharcodeToRune converts PDF character code `code` to a rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc UTF16Encoder) CharcodeToRune(code CharCode) (rune, bool) {
	return rune(code), true
}

// ToPdfObject returns a PDF Object that represents the encoding.
func (enc UTF16Encoder) ToPdfObject() core.PdfObject {
	if enc.baseName != "" {
		return core.MakeName(enc.baseName)
	}
	return core.MakeNull()
}
