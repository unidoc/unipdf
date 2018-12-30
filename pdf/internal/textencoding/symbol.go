/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

// NewSymbolEncoder returns a TextEncoder that implements SymbolEncoding.
func NewSymbolEncoder() TextEncoder {
	enc, _ := NewSimpleTextEncoder("SymbolEncoding", nil)
	return enc
}
