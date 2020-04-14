/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package crypt

import (
	"crypto/md5"
	"crypto/rc4"
	"fmt"

	"github.com/unidoc/unipdf/v3/common"
)

func init() {
	registerFilter("V2", newFilterV2)
}

// NewFilterV2 creates a RC4-based filter with a specified key length (in bytes).
func NewFilterV2(length int) Filter {
	f, err := newFilterV2(FilterDict{Length: length})
	if err != nil {
		common.Log.Error("ERROR: could not create RC4 V2 crypt filter: %v", err)
		return filterV2{length: length}
	}
	return f
}

// newFilterV2 creates a RC4-based filter from a Filter dictionary.
func newFilterV2(d FilterDict) (Filter, error) {
	if d.Length%8 != 0 {
		return nil, fmt.Errorf("crypt filter length not multiple of 8 (%d)", d.Length)
	}
	// Standard security handler expresses the length in multiples of 8 (16 means 128)
	// We only deal with standard so far. (Public key not supported yet).
	if d.Length < 5 || d.Length > 16 {
		if d.Length == 40 || d.Length == 64 || d.Length == 128 {
			common.Log.Debug("STANDARD VIOLATION: Crypt Length appears to be in bits rather than bytes - assuming bits (%d)", d.Length)
			d.Length /= 8
		} else {
			return nil, fmt.Errorf("crypt filter length not in range 40 - 128 bit (%d)", d.Length)
		}
	}
	return filterV2{length: d.Length}, nil
}

// makeKeyV2 is a common object key generation shared by V2 and AESV2 crypt filters.
func makeKeyV2(objNum, genNum uint32, ekey []byte, isAES bool) ([]byte, error) {
	key := make([]byte, len(ekey)+5)
	for i := 0; i < len(ekey); i++ {
		key[i] = ekey[i]
	}
	for i := 0; i < 3; i++ {
		b := byte((objNum >> uint32(8*i)) & 0xff)
		key[i+len(ekey)] = b
	}
	for i := 0; i < 2; i++ {
		b := byte((genNum >> uint32(8*i)) & 0xff)
		key[i+len(ekey)+3] = b
	}
	if isAES {
		// If using the AES algorithm, extend the encryption key an
		// additional 4 bytes by adding the value “sAlT”, which
		// corresponds to the hexadecimal values 0x73, 0x41, 0x6C, 0x54.
		key = append(key, 0x73)
		key = append(key, 0x41)
		key = append(key, 0x6C)
		key = append(key, 0x54)
	}

	// Take the MD5.
	h := md5.New()
	h.Write(key)
	hashb := h.Sum(nil)

	if len(ekey)+5 < 16 {
		return hashb[0 : len(ekey)+5], nil
	}

	return hashb, nil
}

var _ Filter = filterV2{}

// filterV2 is a RC4-based filter
type filterV2 struct {
	length int
}

// PDFVersion implements Filter interface.
func (f filterV2) PDFVersion() [2]int {
	return [2]int{} // TODO(dennwc): unspecified; check what it should be
}

// HandlerVersion implements Filter interface.
func (f filterV2) HandlerVersion() (V, R int) {
	V, R = 2, 3
	return
}

// Name implements Filter interface.
func (filterV2) Name() string {
	return "V2"
}

// KeyLength implements Filter interface.
func (f filterV2) KeyLength() int {
	return f.length
}

// MakeKey implements Filter interface.
func (f filterV2) MakeKey(objNum, genNum uint32, ekey []byte) ([]byte, error) {
	return makeKeyV2(objNum, genNum, ekey, false)
}

// EncryptBytes implements Filter interface.
func (filterV2) EncryptBytes(buf []byte, okey []byte) ([]byte, error) {
	// Standard RC4 algorithm.
	ciph, err := rc4.NewCipher(okey)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("RC4 Encrypt: % x", buf)
	ciph.XORKeyStream(buf, buf)
	common.Log.Trace("to: % x", buf)
	return buf, nil
}

// DecryptBytes implements Filter interface.
func (filterV2) DecryptBytes(buf []byte, okey []byte) ([]byte, error) {
	// Standard RC4 algorithm.
	ciph, err := rc4.NewCipher(okey)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("RC4 Decrypt: % x", buf)
	ciph.XORKeyStream(buf, buf)
	common.Log.Trace("to: % x", buf)
	return buf, nil
}
