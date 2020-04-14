/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"math"

	"github.com/unidoc/unipdf/v3/common"
)

var _ StdHandler = stdHandlerR6{}

// newAESCipher creates a new AES block cipher.
// The size of a buffer should be exactly 16, 24 or 32 bytes.
func newAESCipher(b []byte) (cipher.Block, error) {
	c, err := aes.NewCipher(b)
	if err != nil {
		common.Log.Error("ERROR: could not create AES cipher: %v", err)
		return nil, err
	}
	return c, nil
}

// NewHandlerR6 creates a new standard security handler for R=5 and R=6.
func NewHandlerR6() StdHandler {
	return stdHandlerR6{}
}

// stdHandlerR6 is an implementation of standard security handler with R=5 and R=6.
// Both revisions are expected to be used with AES encryption filters.
type stdHandlerR6 struct{}

// alg2a retrieves the encryption key from an encrypted document (R >= 5).
// 7.6.4.3.2 Algorithm 2.A (page 83)
func (sh stdHandlerR6) alg2a(d *StdEncryptDict, pass []byte) ([]byte, Permissions, error) {
	// O & U: 32 byte hash + 8 byte Validation Salt + 8 byte Key Salt
	if err := checkAtLeast("alg2a", "O", 48, d.O); err != nil {
		return nil, 0, err
	}
	if err := checkAtLeast("alg2a", "U", 48, d.U); err != nil {
		return nil, 0, err
	}

	// step a: Unicode normalization
	// TODO(dennwc): make sure that UTF-8 strings are normalized

	// step b: truncate to 127 bytes
	if len(pass) > 127 {
		pass = pass[:127]
	}

	// step c: test pass against the owner key
	h, err := sh.alg12(d, pass)
	if err != nil {
		return nil, 0, err
	}
	var (
		data []byte // data to hash
		ekey []byte // encrypted file key
		ukey []byte // user key; set only when using owner's password
	)
	var perm Permissions
	if len(h) != 0 {
		// owner password valid
		perm = PermOwner

		// step d: compute an intermediate owner key
		str := make([]byte, len(pass)+8+48)
		i := copy(str, pass)
		i += copy(str[i:], d.O[40:48]) // owner Key Salt
		i += copy(str[i:], d.U[0:48])

		data = str
		ekey = d.OE
		ukey = d.U[0:48]
	} else {
		// check user password
		h, err = sh.alg11(d, pass)
		if err == nil && len(h) == 0 {
			// try default password
			h, err = sh.alg11(d, []byte(""))
		}
		if err != nil {
			return nil, 0, err
		} else if len(h) == 0 {
			// wrong password
			return nil, 0, nil
		}
		perm = d.P
		// step e: compute an intermediate user key
		str := make([]byte, len(pass)+8)
		i := copy(str, pass)
		i += copy(str[i:], d.U[40:48]) // user Key Salt

		data = str
		ekey = d.UE
		ukey = nil
	}
	if err := checkAtLeast("alg2a", "Key", 32, ekey); err != nil {
		return nil, 0, err
	}
	ekey = ekey[:32]

	// intermediate key
	ikey, err := sh.alg2b(d.R, data, pass, ukey)
	if err != nil {
		return nil, 0, err
	}

	ac, err := aes.NewCipher(ikey[:32])
	if err != nil {
		return nil, 0, err
	}

	iv := make([]byte, aes.BlockSize)
	cbc := cipher.NewCBCDecrypter(ac, iv)
	fkey := make([]byte, 32)
	cbc.CryptBlocks(fkey, ekey)

	if d.R == 5 {
		return fkey, perm, nil
	}
	// validate permissions
	err = sh.alg13(d, fkey)
	if err != nil {
		return nil, 0, err
	}
	return fkey, perm, nil
}

// alg2bR5 computes a hash for R=5, used in a deprecated extension.
// It's used the same way as a hash described in Algorithm 2.B, but it doesn't use the original password
// and the user key to calculate the hash.
func alg2bR5(data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil), nil
}

// repeat repeats first n bytes of buf until the end of the buffer.
// It assumes that the length of buf is a multiple of n.
func repeat(buf []byte, n int) {
	bp := n
	for bp < len(buf) {
		copy(buf[bp:], buf[:bp])
		bp *= 2
	}
}

// alg2b computes a hash for R=6.
// 7.6.4.3.3 Algorithm 2.B (page 83)
func alg2b(data, pwd, userKey []byte) ([]byte, error) {
	var (
		s256, s384, s512 hash.Hash
	)
	s256 = sha256.New()
	hbuf := make([]byte, 64)

	h := s256
	h.Write(data)
	K := h.Sum(hbuf[:0])

	buf := make([]byte, 64*(127+64+48))

	round := func(rnd int) ([]byte, error) {
		// step a: repeat pass+K 64 times
		n := len(pwd) + len(K) + len(userKey)
		part := buf[:n]
		i := copy(part, pwd)
		i += copy(part[i:], K[:])
		i += copy(part[i:], userKey)
		if i != n {
			common.Log.Error("ERROR: unexpected round input size.")
			return nil, errors.New("wrong size")
		}
		K1 := buf[:n*64]
		repeat(K1, n)

		// step b: encrypt K1 with AES-128 CBC
		ac, err := newAESCipher(K[0:16])
		if err != nil {
			return nil, err
		}

		cbc := cipher.NewCBCEncrypter(ac, K[16:32])
		cbc.CryptBlocks(K1, K1)
		E := K1

		// step c: use 16 bytes of E as big-endian int, select the next hash
		b := 0
		for i := 0; i < 16; i++ {
			b += int(E[i] % 3)
		}
		var h hash.Hash
		switch b % 3 {
		case 0:
			h = s256
		case 1:
			if s384 == nil {
				s384 = sha512.New384()
			}
			h = s384
		case 2:
			if s512 == nil {
				s512 = sha512.New()
			}
			h = s512
		}

		// step d: take the hash of E, use as a new K
		h.Reset()
		h.Write(E)
		K = h.Sum(hbuf[:0])

		return E, nil
	}

	for i := 0; ; {
		E, err := round(i)
		if err != nil {
			return nil, err
		}

		b := uint8(E[len(E)-1])
		// from the spec, it appears that i should be incremented after
		// the test, but that doesn't match what Adobe does
		i++
		if i >= 64 && b <= uint8(i-32) {
			break
		}
	}
	return K[:32], nil
}

// alg2b computes a hash for R=5 and R=6.
func (sh stdHandlerR6) alg2b(R int, data, pwd, userKey []byte) ([]byte, error) {
	if R == 5 {
		return alg2bR5(data)
	}
	return alg2b(data, pwd, userKey)
}

// alg8 computes the encryption dictionary's U (user password) and UE (user encryption) values (R>=5).
// 7.6.4.4.6 Algorithm 8 (page 86)
func (sh stdHandlerR6) alg8(d *StdEncryptDict, ekey []byte, upass []byte) error {
	if err := checkAtLeast("alg8", "Key", 32, ekey); err != nil {
		return err
	}
	// step a: compute U (user password)
	var rbuf [16]byte
	if _, err := io.ReadFull(rand.Reader, rbuf[:]); err != nil {
		return err
	}
	valSalt := rbuf[0:8]
	keySalt := rbuf[8:16]

	str := make([]byte, len(upass)+len(valSalt))
	i := copy(str, upass)
	i += copy(str[i:], valSalt)

	h, err := sh.alg2b(d.R, str, upass, nil)
	if err != nil {
		return err
	}

	U := make([]byte, len(h)+len(valSalt)+len(keySalt))
	i = copy(U, h[:32])
	i += copy(U[i:], valSalt)
	i += copy(U[i:], keySalt)

	d.U = U

	// step b: compute UE (user encryption)

	// str still contains a password, reuse it
	i = len(upass)
	i += copy(str[i:], keySalt)

	h, err = sh.alg2b(d.R, str, upass, nil)
	if err != nil {
		return err
	}

	ac, err := newAESCipher(h[:32])
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	cbc := cipher.NewCBCEncrypter(ac, iv)
	UE := make([]byte, 32)
	cbc.CryptBlocks(UE, ekey[:32])
	d.UE = UE

	return nil
}

// alg9 computes the encryption dictionary's O (owner password) and OE (owner encryption) values (R>=5).
// 7.6.4.4.7 Algorithm 9 (page 86)
func (sh stdHandlerR6) alg9(d *StdEncryptDict, ekey []byte, opass []byte) error {
	if err := checkAtLeast("alg9", "Key", 32, ekey); err != nil {
		return err
	}
	if err := checkAtLeast("alg9", "U", 48, d.U); err != nil {
		return err
	}
	// step a: compute O (owner password)
	var rbuf [16]byte
	if _, err := io.ReadFull(rand.Reader, rbuf[:]); err != nil {
		return err
	}
	valSalt := rbuf[0:8]
	keySalt := rbuf[8:16]
	userKey := d.U[:48]

	str := make([]byte, len(opass)+len(valSalt)+len(userKey))
	i := copy(str, opass)
	i += copy(str[i:], valSalt)
	i += copy(str[i:], userKey)

	h, err := sh.alg2b(d.R, str, opass, userKey)
	if err != nil {
		return err
	}

	O := make([]byte, len(h)+len(valSalt)+len(keySalt))
	i = copy(O, h[:32])
	i += copy(O[i:], valSalt)
	i += copy(O[i:], keySalt)

	d.O = O

	// step b: compute OE (owner encryption)

	// str still contains a password and a user key - reuse both, but overwrite the salt
	i = len(opass)
	i += copy(str[i:], keySalt)
	// i += len(userKey)

	h, err = sh.alg2b(d.R, str, opass, userKey)
	if err != nil {
		return err
	}

	ac, err := newAESCipher(h[:32])
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	cbc := cipher.NewCBCEncrypter(ac, iv)
	OE := make([]byte, 32)
	cbc.CryptBlocks(OE, ekey[:32])
	d.OE = OE

	return nil
}

// alg10 computes the encryption dictionary's Perms (permissions) value (R=6).
// 7.6.4.4.8 Algorithm 10 (page 87)
func (sh stdHandlerR6) alg10(d *StdEncryptDict, ekey []byte) error {
	if err := checkAtLeast("alg10", "Key", 32, ekey); err != nil {
		return err
	}
	// step a: extend permissions to 64 bits
	perms := uint64(uint32(d.P)) | (math.MaxUint32 << 32)

	// step b: record permissions
	Perms := make([]byte, 16)
	binary.LittleEndian.PutUint64(Perms[:8], perms)

	// step c: record EncryptMetadata
	if d.EncryptMetadata {
		Perms[8] = 'T'
	} else {
		Perms[8] = 'F'
	}

	// step d: write "adb" magic
	copy(Perms[9:12], "adb")

	// step e: write 4 bytes of random data

	// spec doesn't specify them as generated "from a strong random source",
	// but we will use the cryptographic random generator anyway
	if _, err := io.ReadFull(rand.Reader, Perms[12:16]); err != nil {
		return err
	}

	// step f: encrypt permissions
	ac, err := newAESCipher(ekey[:32])
	if err != nil {
		return err
	}

	ecb := newECBEncrypter(ac)
	ecb.CryptBlocks(Perms, Perms)

	d.Perms = Perms[:16]
	return nil
}

// alg11 authenticates the user password (R >= 5) and returns the hash.
func (sh stdHandlerR6) alg11(d *StdEncryptDict, upass []byte) ([]byte, error) {
	if err := checkAtLeast("alg11", "U", 48, d.U); err != nil {
		return nil, err
	}
	str := make([]byte, len(upass)+8)
	i := copy(str, upass)
	i += copy(str[i:], d.U[32:40]) // user Validation Salt

	h, err := sh.alg2b(d.R, str, upass, nil)
	if err != nil {
		return nil, err
	}

	h = h[:32]
	if !bytes.Equal(h, d.U[:32]) {
		return nil, nil
	}
	return h, nil
}

// alg12 authenticates the owner password (R >= 5) and returns the hash.
// 7.6.4.4.10 Algorithm 12 (page 87)
func (sh stdHandlerR6) alg12(d *StdEncryptDict, opass []byte) ([]byte, error) {
	if err := checkAtLeast("alg12", "U", 48, d.U); err != nil {
		return nil, err
	}
	if err := checkAtLeast("alg12", "O", 48, d.O); err != nil {
		return nil, err
	}
	str := make([]byte, len(opass)+8+48)
	i := copy(str, opass)
	i += copy(str[i:], d.O[32:40]) // owner Validation Salt
	i += copy(str[i:], d.U[0:48])

	h, err := sh.alg2b(d.R, str, opass, d.U[0:48])
	if err != nil {
		return nil, err
	}

	h = h[:32]
	if !bytes.Equal(h, d.O[:32]) {
		return nil, nil
	}
	return h, nil
}

// alg13 validates user permissions (P+EncryptMetadata vs Perms) for R=6.
// 7.6.4.4.11 Algorithm 13 (page 87)
func (sh stdHandlerR6) alg13(d *StdEncryptDict, fkey []byte) error {
	if err := checkAtLeast("alg13", "Key", 32, fkey); err != nil {
		return err
	}
	if err := checkAtLeast("alg13", "Perms", 16, d.Perms); err != nil {
		return err
	}
	perms := make([]byte, 16)
	copy(perms, d.Perms[:16])

	ac, err := aes.NewCipher(fkey[:32])
	if err != nil {
		return err
	}

	ecb := newECBDecrypter(ac)
	ecb.CryptBlocks(perms, perms)

	if !bytes.Equal(perms[9:12], []byte("adb")) {
		return errors.New("decoded permissions are invalid")
	}
	p := Permissions(binary.LittleEndian.Uint32(perms[0:4]))
	if p != d.P {
		return errors.New("permissions validation failed")
	}
	encMeta := true
	if perms[8] == 'T' {
		encMeta = true
	} else if perms[8] == 'F' {
		encMeta = false
	} else {
		return errors.New("decoded metadata encryption flag is invalid")
	}
	if encMeta != d.EncryptMetadata {
		return errors.New("metadata encryption validation failed")
	}
	return nil
}

// GenerateParams is the algorithm opposite to alg2a (R>=5).
// It generates U,O,UE,OE,Perms fields using AESv3 encryption.
// There is no algorithm number assigned to this function in the spec.
// It expects R, P and EncryptMetadata fields to be set.
func (sh stdHandlerR6) GenerateParams(d *StdEncryptDict, opass, upass []byte) ([]byte, error) {
	ekey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, ekey); err != nil {
		return nil, err
	}
	// all these field will be populated by functions below
	d.U = nil
	d.O = nil
	d.UE = nil
	d.OE = nil
	d.Perms = nil // populated only for R=6

	if len(upass) > 127 {
		upass = upass[:127]
	}
	if len(opass) > 127 {
		opass = opass[:127]
	}
	// generate U and UE
	if err := sh.alg8(d, ekey, upass); err != nil {
		return nil, err
	}
	// generate O and OE
	if err := sh.alg9(d, ekey, opass); err != nil {
		return nil, err
	}
	if d.R == 5 {
		return ekey, nil
	}
	// generate Perms
	if err := sh.alg10(d, ekey); err != nil {
		return nil, err
	}
	return ekey, nil
}

// Authenticate implements StdHandler interface.
func (sh stdHandlerR6) Authenticate(d *StdEncryptDict, pass []byte) ([]byte, Permissions, error) {
	return sh.alg2a(d, pass)
}
