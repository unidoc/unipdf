/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/unidoc/unipdf/v3/common"
)

func init() {
	registerFilter("AESV3", newFilterAESV3)
}

// NewFilterAESV3 creates an AES-based filter with a 256 bit key (AESV3).
func NewFilterAESV3() Filter {
	f, err := newFilterAESV3(FilterDict{})
	if err != nil {
		common.Log.Error("ERROR: could not create AES V3 crypt filter: %v", err)
		return filterAESV3{}
	}
	return f
}

func newFilterAESV3(d FilterDict) (Filter, error) {
	if d.Length == 256 {
		common.Log.Debug("AESV3 crypt filter length appears to be in bits rather than bytes - assuming bits (%d)", d.Length)
		d.Length /= 8
	}
	if d.Length != 0 && d.Length != 32 {
		return nil, fmt.Errorf("invalid AESV3 crypt filter length (%d)", d.Length)
	}
	return filterAESV3{}, nil
}

// filterAES implements a generic AES encryption and decryption algorithm used by AESV2 and AESV3 filter methods.
type filterAES struct{}

func (filterAES) EncryptBytes(buf []byte, okey []byte) ([]byte, error) {
	// Strings and streams encrypted with AES shall use a padding
	// scheme that is described in Internet RFC 2898, PKCS #5:
	// Password-Based Cryptography Specification Version 2.0; see
	// the Bibliography. For an original message length of M,
	// the pad shall consist of 16 - (M mod 16) bytes whose value
	// shall also be 16 - (M mod 16).
	//
	// A 9-byte message has a pad of 7 bytes, each with the value
	// 0x07. The pad can be unambiguously removed to determine the
	// original message length when decrypting. Note that the pad is
	// present when M is evenly divisible by 16; it contains 16 bytes
	// of 0x10.

	ciph, err := aes.NewCipher(okey)
	if err != nil {
		return nil, err
	}

	common.Log.Trace("AES Encrypt (%d): % x", len(buf), buf)

	// If using the AES algorithm, the Cipher Block Chaining (CBC)
	// mode, which requires an initialization vector, is used. The
	// block size parameter is set to 16 bytes, and the initialization
	// vector is a 16-byte random number that is stored as the first
	// 16 bytes of the encrypted stream or string.

	const block = aes.BlockSize // 16

	pad := block - len(buf)%block
	for i := 0; i < pad; i++ {
		buf = append(buf, byte(pad))
	}
	common.Log.Trace("Padded to %d bytes", len(buf))

	// Generate random 16 bytes, place in beginning of buffer.
	ciphertext := make([]byte, block+len(buf))
	iv := ciphertext[:block]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(ciph, iv)
	mode.CryptBlocks(ciphertext[block:], buf)

	buf = ciphertext
	common.Log.Trace("to (%d): % x", len(buf), buf)

	return buf, nil
}

func (filterAES) DecryptBytes(buf []byte, okey []byte) ([]byte, error) {
	// Strings and streams encrypted with AES shall use a padding
	// scheme that is described in Internet RFC 2898, PKCS #5:
	// Password-Based Cryptography Specification Version 2.0; see
	// the Bibliography. For an original message length of M,
	// the pad shall consist of 16 - (M mod 16) bytes whose value
	// shall also be 16 - (M mod 16).
	//
	// A 9-byte message has a pad of 7 bytes, each with the value
	// 0x07. The pad can be unambiguously removed to determine the
	// original message length when decrypting. Note that the pad is
	// present when M is evenly divisible by 16; it contains 16 bytes
	// of 0x10.

	ciph, err := aes.NewCipher(okey)
	if err != nil {
		return nil, err
	}

	// If using the AES algorithm, the Cipher Block Chaining (CBC)
	// mode, which requires an initialization vector, is used. The
	// block size parameter is set to 16 bytes, and the initialization
	// vector is a 16-byte random number that is stored as the first
	// 16 bytes of the encrypted stream or string.
	if len(buf) < 16 {
		common.Log.Debug("ERROR AES invalid buf %s", buf)
		return buf, fmt.Errorf("AES: Buf len < 16 (%d)", len(buf))
	}

	iv := buf[:16]
	buf = buf[16:]

	if len(buf)%16 != 0 {
		common.Log.Debug(" iv (%d): % x", len(iv), iv)
		common.Log.Debug("buf (%d): % x", len(buf), buf)
		return buf, fmt.Errorf("AES buf length not multiple of 16 (%d)", len(buf))
	}

	mode := cipher.NewCBCDecrypter(ciph, iv)

	common.Log.Trace("AES Decrypt (%d): % x", len(buf), buf)
	common.Log.Trace("chop AES Decrypt (%d): % x", len(buf), buf)
	mode.CryptBlocks(buf, buf)
	common.Log.Trace("to (%d): % x", len(buf), buf)

	if len(buf) == 0 {
		common.Log.Trace("Empty buf, returning empty string")
		return buf, nil
	}

	// The padded length is indicated by the last values.  Remove those.
	padLen := int(buf[len(buf)-1])
	if padLen > len(buf) {
		common.Log.Debug("Illegal pad length (%d > %d)", padLen, len(buf))
		return buf, fmt.Errorf("invalid pad length")
	}
	buf = buf[:len(buf)-padLen]

	return buf, nil
}

var _ Filter = filterAESV3{}

// filterAESV3 is an AES-based filter (256 bit key, PDF 2.0)
type filterAESV3 struct {
	filterAES
}

// PDFVersion implements Filter interface.
func (filterAESV3) PDFVersion() [2]int {
	return [2]int{2, 0}
}

// HandlerVersion implements Filter interface.
func (filterAESV3) HandlerVersion() (V, R int) {
	V, R = 5, 6
	return
}

// Name implements Filter interface.
func (filterAESV3) Name() string {
	return "AESV3"
}

// KeyLength implements Filter interface.
func (filterAESV3) KeyLength() int {
	return 256 / 8
}

// MakeKey implements Filter interface.
func (filterAESV3) MakeKey(_, _ uint32, ekey []byte) ([]byte, error) {
	return ekey, nil // document encryption key == object encryption key
}
