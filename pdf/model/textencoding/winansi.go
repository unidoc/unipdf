/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

func NewWinAnsiTextEncoder() SimpleEncoder {
	enc, _ := NewSimpleTextEncoder("WinAnsiEncoding", nil)
	return enc
}
