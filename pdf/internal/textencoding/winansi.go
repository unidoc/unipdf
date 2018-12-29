/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "golang.org/x/text/encoding/charmap"

var winAnsiEncoding = make(map[CharCode]rune, 256)

func init() {
	// WinAnsiEncoding is also known as CP1252
	enc := charmap.Windows1252

	// in WinAnsiEncoding all unused and non-visual codes map to the '•' character
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
		if rp, ok := replace[b]; ok {
			r = rp
		}
		winAnsiEncoding[CharCode(b)] = r
	}
}

// NewWinAnsiTextEncoder returns a SimpleEncoder that implements WinAnsiEncoding.
func NewWinAnsiTextEncoder() *SimpleEncoder {
	const baseName = "WinAnsiEncoding"
	enc := newSimpleTextEncoder(winAnsiEncoding, baseName, nil)
	return enc
}
