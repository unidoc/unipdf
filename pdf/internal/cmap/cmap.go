/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"fmt"
	"sort"
	"strings"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// CharCode is a character code or Unicode
// rune is int32 https://golang.org/doc/go1#rune
type CharCode uint32

// Maximum number of possible bytes per code.
const maxCodeLen = 4

// Codespace represents a single codespace range used in the CMap.
type Codespace struct {
	NumBytes int
	Low      CharCode
	High     CharCode
}

// CIDSystemInfo=Dict("Registry": Adobe, "Ordering": Korea1, "Supplement": 0, )
type CIDSystemInfo struct {
	Registry   string
	Ordering   string
	Supplement int
}

// CMap represents a character code to unicode mapping used in PDF files.
//
// 9.7.5 CMaps (Page 272)
//
// Page 278
// c) The beginbfchar and endbfchar shall not appear in a CMap that is used as the Encoding entry of
// a Type 0 font; however, they may appear in the definition of a ToUnicode CMap
//
// https://www.adobe.com/content/dam/acom/en/devnet/acrobat/pdfs/5411.ToUnicode.pdf
// https://github.com/adobe-type-tools/cmap-resources/releases ***
type CMap struct {
	*cMapParser

	name       string
	nbits      int // 8 bits for simple fonts, 16 bits for CID fonts.
	ctype      int
	version    string
	usecmap    string // Base this cmap on `usecmap` if `usecmap` is not empty
	systemInfo CIDSystemInfo

	// For regular cmaps
	codespaces []Codespace

	// Text encoder to look up runes from input glyph names. !@#$ Not used
	// encoder textencoding.TextEncoder

	// For ToUnicode (ctype 2) cmaps
	codeToUnicode     map[CharCode]string
	toUnicodeIdentity bool
}

// String retuns a human readable description of `cmap`
func (cmap *CMap) String() string {
	si := cmap.systemInfo
	parts := []string{
		fmt.Sprintf("nbits:%d", cmap.nbits),
		fmt.Sprintf("type:%d", cmap.ctype),
	}
	if cmap.version != "" {
		parts = append(parts, fmt.Sprintf("version:%s", cmap.version))
	}
	if cmap.usecmap != "" {
		parts = append(parts, fmt.Sprintf("usecmap:%#q", cmap.usecmap))
	}
	parts = append(parts, fmt.Sprintf("systemInfo:%s", si.String()))
	if len(cmap.codespaces) > 0 {
		parts = append(parts, fmt.Sprintf("codespaces:%d", len(cmap.codespaces)))
	}
	if len(cmap.codeToUnicode) > 0 {
		parts = append(parts, fmt.Sprintf("codeToUnicode:%d", len(cmap.codeToUnicode)))
	}
	return fmt.Sprintf("CMAP{%#q %s}", cmap.name, strings.Join(parts, " "))
}

// newCMap returns an initialized CMap.
func newCMap(isSimple bool) *CMap {
	nbits := 16
	if isSimple {
		nbits = 8
	}
	cmap := &CMap{
		nbits:         nbits,
		codeToUnicode: map[CharCode]string{},
	}
	return cmap
}

// String returns a human readable description of `info`
// It looks like "Adobe-Japan2-000".
func (info *CIDSystemInfo) String() string {
	return fmt.Sprintf("%s-%s-%03d", info.Registry, info.Ordering, info.Supplement)
}

// NewCIDSystemInfo returns the CIDSystemInfo encoded in PDFObject `obj`
func NewCIDSystemInfo(obj PdfObject) (info CIDSystemInfo, err error) {
	obj = TraceToDirectObject(obj)
	d := *obj.(*PdfObjectDictionary)
	registry, err := GetString(d.Get("Registry"))
	if err != nil {
		return
	}
	ordering, err := GetString(d.Get("Ordering"))
	if err != nil {
		return
	}
	supplement, err := GetInteger(d.Get("Supplement"))
	if err != nil {
		return
	}
	info = CIDSystemInfo{
		Registry:   registry,
		Ordering:   ordering,
		Supplement: supplement,
	}
	return
}

// Name returns the name of the CMap.
func (cmap *CMap) Name() string {
	return cmap.name
}

// Type the CMap type.
func (cmap *CMap) Type() int {
	return cmap.ctype
}

const (
	// MissingCodeRune replaces runes that can't be decoded.
	// MissingCodeRune = '?'
	MissingCodeRune = textencoding.MissingCodeRune // '\ufffd' // �
	// MissingCodeString replaces strings that can't be decoded.
	// MissingCodeString = "?"
)

var (
	MissingCodeString = string(MissingCodeRune) // �
)

// CharcodeBytesToUnicode converts a byte array of charcodes to a unicode string representation.
// It also returns a bool flag to tell if the conversion was successful.
// NOTE: This only works for ToUnicode cmaps
func (cmap *CMap) CharcodeBytesToUnicode(data []byte) (string, int) {
	charcodes, matched := cmap.bytesToCharcodes(data)
	if !matched {
		common.Log.Debug("ERROR: CharcodeBytesToUnicode. Not in codespaces. data=[% 02x] cmap=%s",
			data, cmap)
		return "", 0
	}

	parts := []string{}
	missing := []CharCode{}
	for _, code := range charcodes {
		s, ok := cmap.codeToUnicode[code]
		if !ok {
			missing = append(missing, code)
			s = MissingCodeString
		}
		parts = append(parts, s)
	}
	unicode := strings.Join(parts, "")
	if len(missing) > 0 {
		common.Log.Debug("ERROR: CharcodeBytesToUnicode. Not in map.\n"+
			"\tdata=[% 02x]=%#q\n"+
			"\tcharcodes=%02x\n"+
			"\tmissing=%d %02x\n"+
			"\tunicode=`%s`\n"+
			"\tcmap=%s",
			data, string(data), charcodes, len(missing), missing, unicode, cmap)
	}
	return unicode, len(missing)
}

// CharcodeToUnicode converts a single character code `code ` to a unicode string.
// If `code` is not in the unicode map, "?" is returned
// Note that CharcodeBytesToUnicode is typically more efficient.
func (cmap *CMap) CharcodeToUnicode(code CharCode) string {
	s, _ := cmap.CharcodeToUnicode2(code)
	return s
}

// CharcodeToUnicode2 converts a single character code `code` to a unicode string.
// The bool value is set to true if `code` is in the unicode map,
func (cmap *CMap) CharcodeToUnicode2(code CharCode) (string, bool) {
	if s, ok := cmap.codeToUnicode[code]; ok {
		return s, true
	}
	common.Log.Debug("ERROR: CharcodeToUnicode could not convert code=0x%04x. cmap=%s. Returning %q",
		code, cmap, MissingCodeString)
	return MissingCodeString, false
}

// bytesToCharcodes attempts to convert the entire byte array `data` to a list of character codes
// from the ranges specified by `cmap`'s codespaces.
// Returns:
//      character code sequence (if there is a match complete match)
//      matched?
// NOTE: A partial list of character codes will be returned if a complete match is not possible.
func (cmap *CMap) bytesToCharcodes(data []byte) ([]CharCode, bool) {
	charcodes := []CharCode{}
	if cmap.nbits == 8 {
		for _, b := range data {
			charcodes = append(charcodes, CharCode(b))
		}
		return charcodes, true
	}
	for i := 0; i < len(data); {
		code, n, matched := cmap.matchCode(data[i:])
		if !matched {
			common.Log.Debug("ERROR: No code match at i=%d bytes=[% 02x]=%#q", i, data, string(data))
			return charcodes, false
		}
		charcodes = append(charcodes, code)
		i += n
	}
	return charcodes, true
}

// matchCode attempts to match the byte array `data` with a character code in `cmap`'s codespaces
// Returns:
//      character code (if there is a match) of
//      number of bytes read (if there is a match)
//      matched?
func (cmap *CMap) matchCode(data []byte) (code CharCode, n int, matched bool) {
	for j := 0; j < maxCodeLen; j++ {
		if j < len(data) {
			code = code<<8 | CharCode(data[j])
			n++
		}
		matched = cmap.inCodespace(code, j+1)
		if matched {
			return
		}
	}
	// No codespace matched data. This is a serious problem.
	common.Log.Debug("ERROR: No codespace matches bytes=[% 02x]=%#q cmap=%s",
		data, string(data), cmap)
	return
}

// inCodespace returns true if `code` in `numBytes` byte codespace
func (cmap *CMap) inCodespace(code CharCode, numBytes int) bool {
	for _, cs := range cmap.codespaces {
		if cs.Low <= code && code <= cs.High && numBytes == cs.NumBytes {
			return true
		}
	}
	return false
}

// LoadCmapFromDataCID parses in-memory cmap `data` and returns the resulting CMap
// Convenience function
func LoadCmapFromDataCID(data []byte) (*CMap, error) {
	return LoadCmapFromData(data, false)
}

// LoadCmapFromData parses in-memory cmap `data` and returns the resulting CMap
// If isCID is true then use 1-byte encodings, otherwise use codespaces in the cmap
//
// 9.10.3 ToUnicode CMaps (page 293)
// The CMap defined in the ToUnicode entry of the font dictionary shall follow the syntax for CMaps
// • The only pertinent entry in the CMap stream dictionary is UseCMap,
// • The CMap file shall contain begincodespacerange and endcodespacerange operators that are
//   consistent with the encoding that the font uses. In particular, for a simple font, the
//   codespace shall be one byte long.
// • It shall use the beginbfchar, endbfchar, beginbfrange, and endbfrange operators to define the
//    mapping from character codes to Unicode character sequences expressed in UTF-16BE encoding
func LoadCmapFromData(data []byte, isSimple bool) (*CMap, error) {
	common.Log.Trace("LoadCmapFromData: isSimple=%t", isSimple)

	cmap := newCMap(isSimple)
	cmap.cMapParser = newCMapParser(data)

	// In debugging it may help to see the data being parsed.
	// fmt.Println("===============*******===========")
	// fmt.Printf("%s\n", string(data))
	// fmt.Println("===============&&&&&&&===========")

	err := cmap.parse()
	if err != nil {
		return nil, err
	}
	if len(cmap.codespaces) == 0 {
		common.Log.Debug("ERROR: No codespaces. cmap=%s", cmap)
		return nil, ErrBadCMap
	}
	// We need to sort codespaces so that we check shorter codes first
	sort.Slice(cmap.codespaces, func(i, j int) bool {
		return cmap.codespaces[i].Low < cmap.codespaces[j].Low
	})
	return cmap, nil
}
