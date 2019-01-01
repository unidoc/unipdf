/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import (
	"sync"

	"golang.org/x/text/encoding/charmap"
)

const baseWinAnsi = "WinAnsiEncoding"

func init() {
	RegisterSimpleEncoding(baseWinAnsi, NewWinAnsiEncoder)
}

var (
	winAnsiOnce       sync.Once
	winAnsiCharToRune map[byte]rune
	winAnsiRuneToChar map[rune]byte
)

// NewWinAnsiEncoder returns a simpleEncoder that implements WinAnsiEncoding.
func NewWinAnsiEncoder() SimpleEncoder {
	winAnsiOnce.Do(initWinAnsi)
	return &simpleEncoding{
		baseName: baseWinAnsi,
		encode:   winAnsiRuneToChar,
		decode:   winAnsiCharToRune,
	}
}

func initWinAnsi() {
	winAnsiCharToRune = make(map[byte]rune, 256)
	winAnsiRuneToChar = make(map[rune]byte, 256)

	// WinAnsiEncoding is also known as CP1252
	enc := charmap.Windows1252

	// in WinAnsiEncoding, comparing to CP1252, all unused and
	// non-visual codes are replaced with '•' character
	const bullet = '•'
	replace := map[byte]rune{
		127: bullet, // DEL

		// unused
		129: bullet,
		141: bullet,
		143: bullet,
		144: bullet,
		157: bullet,

		// typographically similar
		160: ' ', // non-breaking space -> space
		173: '-', // soft hyphen -> hyphen
	}

	for i := int(' '); i < 256; i++ {
		b := byte(i)
		r := enc.DecodeByte(b)

		// don't use replace map. since it creates duplicates
		winAnsiRuneToChar[r] = b

		if rp, ok := replace[b]; ok {
			r = rp
		}
		winAnsiCharToRune[b] = r
	}
}
