/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

func NewSymbolEncoder() SimpleEncoder {
	enc, _ := NewSimpleTextEncoder("SymbolEncoding", nil)
	return enc
}
