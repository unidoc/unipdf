/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

// NewZapfDingbatsEncoder returns a SimpleEncoder that implements ZapfDingbatsEncoding.
func NewZapfDingbatsEncoder() SimpleEncoder {
	enc, _ := NewSimpleTextEncoder("ZapfDingbatsEncoding", nil)
	return enc
}
