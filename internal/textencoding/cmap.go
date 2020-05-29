/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"bytes"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/cmap"
)

// CMapEncoder encodes/decodes strings based on CMap mappings.
type CMapEncoder struct {
	baseName     string
	codeToCID    *cmap.CMap
	cidToUnicode *cmap.CMap
}

// NewCMapEncoder returns a new CMapEncoder based on the predefined
// encoding `baseName`. If `codeToCID` is nil, Identity encoding is assumed.
// `cidToUnicode` must not be nil.
func NewCMapEncoder(baseName string, codeToCID, cidToUnicode *cmap.CMap) CMapEncoder {
	return CMapEncoder{
		baseName:     baseName,
		codeToCID:    codeToCID,
		cidToUnicode: cidToUnicode,
	}
}

// Encode converts the Go unicode string to a PDF encoded string.
func (enc CMapEncoder) Encode(str string) []byte {
	if enc.cidToUnicode == nil {
		return []byte{}
	}

	if enc.cidToUnicode.NBits() == 8 {
		return encodeString8bit(enc, str)
	}
	return encodeString16bit(enc, str)
}

// Decode converts PDF encoded string to a Go unicode string.
func (enc CMapEncoder) Decode(raw []byte) string {
	if enc.codeToCID != nil {
		if codes, ok := enc.codeToCID.BytesToCharcodes(raw); ok {
			var buf bytes.Buffer
			for _, code := range codes {
				s, _ := enc.charcodeToString(CharCode(code))
				buf.WriteString(s)
			}

			return buf.String()
		}
	}

	return decodeString16bit(enc, raw)
}

// RuneToCharcode converts rune `r` to a PDF character code.
// The bool return flag is true if there was a match, and false otherwise.
func (enc CMapEncoder) RuneToCharcode(r rune) (CharCode, bool) {
	if enc.cidToUnicode == nil {
		return 0, false
	}

	// Map rune to CID.
	cid, ok := enc.cidToUnicode.StringToCID(string(r))
	if !ok {
		return 0, false
	}

	// Map CID to charcode. If charcode to CID CMap is nil, assume Identity encoding.
	if enc.codeToCID != nil {
		code, ok := enc.codeToCID.CIDToCharcode(cid)
		if !ok {
			return 0, false
		}
		return CharCode(code), true
	}

	return CharCode(cid), true
}

// CharcodeToRune converts PDF character code `code` to a rune.
// The bool return flag is true if there was a match, and false otherwise.
func (enc CMapEncoder) CharcodeToRune(code CharCode) (rune, bool) {
	s, ok := enc.charcodeToString(code)
	return ([]rune(s))[0], ok
}

func (enc CMapEncoder) charcodeToString(code CharCode) (string, bool) {
	if enc.cidToUnicode == nil {
		return MissingCodeString, false
	}

	// Map charcode to CID. If charcode to CID CMap is nil, assume Identity encoding.
	cid := cmap.CharCode(code)
	if enc.codeToCID != nil {
		var ok bool
		if cid, ok = enc.codeToCID.CharcodeToCID(cmap.CharCode(code)); !ok {
			return MissingCodeString, false
		}
	}

	// Map CID to rune.
	return enc.cidToUnicode.CharcodeToUnicode(cid)
}

// String returns a string that describes `enc`.
func (enc CMapEncoder) String() string {
	return enc.baseName
}

// ToPdfObject returns a PDF Object that represents the encoding.
func (enc CMapEncoder) ToPdfObject() core.PdfObject {
	if enc.baseName != "" {
		return core.MakeName(enc.baseName)
	}
	return core.MakeNull()
}
