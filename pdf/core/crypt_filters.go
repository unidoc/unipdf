package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"fmt"
	"io"

	"github.com/unidoc/unidoc/common"
)

var (
	cryptMethods = make(map[string]cryptFilterMethod)
)

// registerCryptFilterMethod registers a CFM.
func registerCryptFilterMethod(m cryptFilterMethod) {
	cryptMethods[m.CFM()] = m
}

// getCryptFilterMethod check if a CFM with a specified name is supported an returns its implementation.
func getCryptFilterMethod(name string) (cryptFilterMethod, error) {
	f := cryptMethods[name]
	if f == nil {
		return nil, fmt.Errorf("unsupported crypt filter: %q", name)
	}
	return f, nil
}

func init() {
	// register supported crypt filter methods
	registerCryptFilterMethod(cryptFilterV2{})
	registerCryptFilterMethod(cryptFilterAESV2{})
	registerCryptFilterMethod(cryptFilterAESV3{})
}

// cryptFilterMethod is a common interface for crypt filter methods.
type cryptFilterMethod interface {
	// CFM returns a name of the filter that should be used in CFM field of Encrypt dictionary.
	CFM() string
	// MakeKey generates a object encryption key based on file encryption key and object numbers.
	// Used only for legacy filters - AESV3 doesn't change the key for each object.
	MakeKey(objNum, genNum uint32, fkey []byte) ([]byte, error)
	// EncryptBytes encrypts a buffer using object encryption key, as returned by MakeKey.
	// Implementation may reuse a buffer and encrypt data in-place.
	EncryptBytes(p []byte, okey []byte) ([]byte, error)
	// DecryptBytes decrypts a buffer using object encryption key, as returned by MakeKey.
	// Implementation may reuse a buffer and decrypt data in-place.
	DecryptBytes(p []byte, okey []byte) ([]byte, error)
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

// cryptFilterV2 is a RC4-based filter
type cryptFilterV2 struct{}

func (cryptFilterV2) CFM() string {
	return CryptFilterV2
}

func (f cryptFilterV2) MakeKey(objNum, genNum uint32, ekey []byte) ([]byte, error) {
	return makeKeyV2(objNum, genNum, ekey, false)
}

func (cryptFilterV2) EncryptBytes(buf []byte, okey []byte) ([]byte, error) {
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

func (cryptFilterV2) DecryptBytes(buf []byte, okey []byte) ([]byte, error) {
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

// cryptFilterAES implements a generic AES encryption and decryption algorithm used by AESV2 and AESV3 filter methods.
type cryptFilterAES struct{}

func (cryptFilterAES) EncryptBytes(buf []byte, okey []byte) ([]byte, error) {
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

func (cryptFilterAES) DecryptBytes(buf []byte, okey []byte) ([]byte, error) {
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
	if padLen >= len(buf) {
		common.Log.Debug("Illegal pad length")
		return buf, fmt.Errorf("Invalid pad length")
	}
	buf = buf[:len(buf)-padLen]

	return buf, nil
}

// cryptFilterAESV2 is an AES-based filter (128 bit key, PDF 1.6)
type cryptFilterAESV2 struct {
	cryptFilterAES
}

func (cryptFilterAESV2) CFM() string {
	return CryptFilterAESV2
}

func (cryptFilterAESV2) MakeKey(objNum, genNum uint32, ekey []byte) ([]byte, error) {
	return makeKeyV2(objNum, genNum, ekey, true)
}

// cryptFilterAESV3 is an AES-based filter (256 bit key, PDF 2.0)
type cryptFilterAESV3 struct {
	cryptFilterAES
}

func (cryptFilterAESV3) CFM() string {
	return CryptFilterAESV3
}

func (cryptFilterAESV3) MakeKey(_, _ uint32, ekey []byte) ([]byte, error) {
	return ekey, nil
}
