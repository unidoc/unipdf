/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"unicode/utf8"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"golang.org/x/text/encoding"
	xtransform "golang.org/x/text/transform"
)

// SimpleEncoder represents a 1 byte encoding.
type SimpleEncoder interface {
	TextEncoder
	BaseName() string
	Charcodes() []CharCode
}

// NewCustomSimpleTextEncoder returns a simpleEncoder based on map `encoding` and difference map
// `differences`.
func NewCustomSimpleTextEncoder(encoding, differences map[CharCode]GlyphName) (SimpleEncoder, error) {
	if len(encoding) == 0 {
		return nil, errors.New("empty custom encoding")
	}
	const baseName = "custom"
	baseEncoding := make(map[byte]rune)
	for code, glyph := range encoding {
		r, ok := GlyphToRune(glyph)
		if !ok {
			common.Log.Debug("ERROR: Unknown glyph. %q", glyph)
			continue
		}
		baseEncoding[byte(code)] = r
	}
	// TODO(dennwc): this seems to be incorrect - baseEncoding won't be saved when converting to PDF object
	enc := newSimpleEncoderFromMap(baseName, baseEncoding)
	if len(differences) != 0 {
		enc = ApplyDifferences(enc, differences)
	}
	return enc, nil
}

// NewSimpleTextEncoder returns a simpleEncoder based on predefined encoding `baseName` and
// difference map `differences`.
func NewSimpleTextEncoder(baseName string, differences map[CharCode]GlyphName) (SimpleEncoder, error) {
	fnc, ok := simple[baseName]
	if !ok {
		common.Log.Debug("ERROR: NewSimpleTextEncoder. Unknown encoding %q", baseName)
		return nil, fmt.Errorf("unsupported font encoding: %q (%v)", baseName, core.ErrNotSupported)
	}
	enc := fnc()
	if len(differences) != 0 {
		enc = ApplyDifferences(enc, differences)
	}
	return enc, nil
}

func newSimpleEncoderFromMap(name string, encoding map[byte]rune) SimpleEncoder {
	se := &simpleEncoding{
		baseName: name,
		decode:   encoding,
		encode:   make(map[rune]byte, len(encoding)),
	}

	// If more than one charcodes map to the same rune in `encoding` charcode->rune map then always
	// use the lower charcode in the `se.encode` rune->charcode map for consistency.
	for b, r := range se.decode {
		if b2, has := se.encode[r]; !has || b < b2 {
			se.encode[r] = b
		}
	}
	return se
}

var (
	simple = make(map[string]func() SimpleEncoder)
)

// RegisterSimpleEncoding registers a SimpleEncoder constructer by PDF encoding name.
func RegisterSimpleEncoding(name string, fnc func() SimpleEncoder) {
	if _, ok := simple[name]; ok {
		panic("already registered")
	}
	simple[name] = fnc
}

var (
	_ SimpleEncoder     = (*simpleEncoding)(nil)
	_ encoding.Encoding = (*simpleEncoding)(nil)
)

// simpleEncoding represents a 1 byte encoding.
type simpleEncoding struct {
	baseName string
	// one byte encoding: CharCode <-> byte
	encode map[rune]byte
	decode map[byte]rune

	// runes registered by encoder for tracking what runes are used for subsetting.
	registeredMap map[rune]struct{}
}

// Encode converts the Go unicode string to a PDF encoded string.
func (enc *simpleEncoding) Encode(str string) []byte {
	data, _ := enc.NewEncoder().Bytes([]byte(str))
	return data
}

// Decode converts PDF encoded string to a Go unicode string.
func (enc *simpleEncoding) Decode(raw []byte) string {
	data, _ := enc.NewDecoder().Bytes(raw)
	return string(data)
}

// NewDecoder implements encoding.Encoding.
func (enc *simpleEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: simpleDecoder{m: enc.decode}}
}

type simpleDecoder struct {
	m map[byte]rune
}

// Transform implements xtransform.Transformer.
func (enc simpleDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, _ error) {
	for len(src) != 0 {
		b := src[0]
		src = src[1:]

		r, ok := enc.m[b]
		if !ok {
			r = MissingCodeRune
		}
		if utf8.RuneLen(r) > len(dst) {
			return nDst, nSrc, xtransform.ErrShortDst
		}
		n := utf8.EncodeRune(dst, r)
		dst = dst[n:]

		nSrc++
		nDst += n
	}
	return nDst, nSrc, nil
}

// Reset implements xtransform.Transformer.
func (enc simpleDecoder) Reset() {}

// NewEncoder implements encoding.Encoding.
func (enc *simpleEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: simpleEncoder{m: enc.encode}}
}

type simpleEncoder struct {
	m map[rune]byte
}

// Transform implements xtransform.Transformer.
func (enc simpleEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, _ error) {
	for len(src) != 0 {
		if !utf8.FullRune(src) && !atEOF {
			return nDst, nSrc, xtransform.ErrShortSrc
		} else if len(dst) == 0 {
			return nDst, nSrc, xtransform.ErrShortDst
		}
		r, n := utf8.DecodeRune(src)
		if r == utf8.RuneError {
			r = MissingCodeRune
		}
		src = src[n:]
		nSrc += n

		b, ok := enc.m[r]
		if !ok {
			b, _ = enc.m[MissingCodeRune]
		}
		dst[0] = b

		dst = dst[1:]
		nDst++
	}
	return nDst, nSrc, nil
}

// Reset implements xtransform.Transformer.
func (enc simpleEncoder) Reset() {}

// String returns a text representation of encoding.
func (enc *simpleEncoding) String() string {
	return "simpleEncoding(" + enc.baseName + ")"
}

// BaseName returns a base name of the encoder, as specified in the PDF spec.
func (enc *simpleEncoding) BaseName() string {
	return enc.baseName
}

func (enc *simpleEncoding) Charcodes() []CharCode {
	codes := make([]CharCode, 0, len(enc.decode))
	for b := range enc.decode {
		codes = append(codes, CharCode(b))
	}
	sort.Slice(codes, func(i, j int) bool {
		return codes[i] < codes[j]
	})
	return codes
}

func (enc *simpleEncoding) RuneToCharcode(r rune) (CharCode, bool) {
	b, ok := enc.encode[r]
	if enc.registeredMap == nil {
		enc.registeredMap = map[rune]struct{}{}
	}
	enc.registeredMap[r] = struct{}{} // Register use (subsetting).
	return CharCode(b), ok
}

func (enc *simpleEncoding) CharcodeToRune(code CharCode) (rune, bool) {
	if code > 0xff {
		return MissingCodeRune, false
	}
	b := byte(code)
	r, ok := enc.decode[b]
	if enc.registeredMap == nil {
		enc.registeredMap = map[rune]struct{}{}
	}
	enc.registeredMap[r] = struct{}{} // Register use (subsetting).
	return r, ok
}

func (enc *simpleEncoding) ToPdfObject() core.PdfObject {
	return core.MakeName(enc.baseName)
}

// newSimpleMapping creates a byte-to-rune mapping that can be used to create simple encodings.
// An implementation will build reverse map only once when the encoding is first used.
func newSimpleMapping(name string, m map[byte]rune) *simpleMapping {
	return &simpleMapping{
		baseName: name,
		decode:   m,
	}
}

type simpleMapping struct {
	baseName string
	once     sync.Once
	decode   map[byte]rune
	encode   map[rune]byte
}

func (m *simpleMapping) init() {
	m.encode = make(map[rune]byte, len(m.decode))
	// If more than one charcodes map to the same rune in encoding charcode->rune map then always
	// use the lower charcode in the `se.encode` rune->charcode map for consistency.
	for b, r := range m.decode {
		if b2, has := m.encode[r]; !has || b < b2 {
			m.encode[r] = b
		}
	}
}

// NewEncoder creates a new SimpleEncoding from the byte-to-rune mapping.
func (m *simpleMapping) NewEncoder() SimpleEncoder {
	m.once.Do(m.init)
	return &simpleEncoding{
		baseName: m.baseName,
		encode:   m.encode,
		decode:   m.decode,
	}
}
