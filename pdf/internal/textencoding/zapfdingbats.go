/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

// NewZapfDingbatsEncoder returns a TextEncoder that implements ZapfDingbatsEncoding.
func NewZapfDingbatsEncoder() TextEncoder {
	enc, _ := NewSimpleTextEncoder("ZapfDingbatsEncoding", nil)
	return enc
}
