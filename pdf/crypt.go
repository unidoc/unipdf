/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

/*
 * The PDF standard supports encryption of strings and streams.
 * Section 7.6.
 */
package pdf

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"errors"
	"fmt"
	"io"

	"github.com/unidoc/unidoc/common"
)

type PdfCrypt struct {
	Filter           string
	Subfilter        string
	V                int
	length           int
	R                int
	O                []byte
	U                []byte
	P                int
	encryptMetadata  bool
	id0              string
	encryptionKey    []byte
	decryptedObjects map[PdfObject]bool
	encryptedObjects map[PdfObject]bool
	authenticated    bool
	// Crypt filters (V4).
	cryptFilters CryptFilters
	streamFilter string
	stringFilter string
}

type AccessPermissions struct {
	Printing          bool
	Modify            bool
	ExtractGraphics   bool
	Annotate          bool
	FillForms         bool
	DisabilityExtract bool // not clear what this means!
	RotateInsert      bool
	LimitPrintQuality bool
}

const padding = "\x28\xBF\x4E\x5E\x4E\x75\x8A\x41\x64\x00\x4E\x56\xFF" +
	"\xFA\x01\x08\x2E\x2E\x00\xB6\xD0\x68\x3E\x80\x2F\x0C" +
	"\xA9\xFE\x64\x53\x69\x7A"

type CryptFilter struct {
	cfm    string
	length int
}

type CryptFilters map[string]CryptFilter

// Load crypt filter information from the encryption dictionary (V4 only).
func (this *PdfCrypt) LoadCryptFilters(ed *PdfObjectDictionary) error {
	this.cryptFilters = CryptFilters{}

	cf, ok := (*ed)["CF"].(*PdfObjectDictionary)
	if !ok {
		return errors.New("Invalid CF")
	}

	for name, v := range *cf {
		dict, ok := v.(*PdfObjectDictionary)
		if !ok {
			return fmt.Errorf("Invalid dict in CF (name %s)", name)
		}

		if name == "Identity" {
			common.Log.Debug("ERROR - Cannot overwrite the identity filter - Trying next")
			continue
		}

		// If Type present, should be CryptFilter.
		if typename, ok := (*dict)["Type"].(*PdfObjectName); ok {
			if string(*typename) != "CryptFilter" {
				return fmt.Errorf("CF dict type != CryptFilter (%s)", typename)
			}
		}

		cf := CryptFilter{}

		// Method.
		cfMethod := "None" // Default.
		cfm, ok := (*dict)["CFM"].(*PdfObjectName)
		if ok {
			if *cfm == "V2" {
				cfMethod = "V2"
			} else if *cfm == "AESV2" {
				cfMethod = "AESV2"
			} else {
				return fmt.Errorf("Unsupported crypt filter (%s)", *cfm)
			}
		}
		if cfMethod != "V2" && cfMethod != "AESV2" {
			return fmt.Errorf("Unsupported crypt filter (%s)", cfMethod)
		}
		cf.cfm = cfMethod

		// Length.
		cf.length = 0
		length, ok := (*dict)["Length"].(*PdfObjectInteger)
		if ok {
			if *length%8 != 0 {
				return fmt.Errorf("Crypt filter length not multiple of 8 (%d)", *length)
			}

			// Standard security handler expresses the length in multiples of 8 (16 means 128)
			// We only deal with standard so far. (Public key not supported yet).
			if *length < 5 || *length > 16 {
				return fmt.Errorf("Crypt filter length not in range 40 - 128 bit (%d)", *length)
			}
			cf.length = int(*length)
		}

		this.cryptFilters[string(name)] = cf
	}
	// Cannot be overwritten.
	this.cryptFilters["Identity"] = CryptFilter{}

	// StrF strings filter.
	this.stringFilter = "Identity"
	if strf, ok := (*ed)["StrF"].(*PdfObjectName); ok {
		if _, exists := this.cryptFilters[string(*strf)]; !exists {
			return fmt.Errorf("Crypt filter for StrF not specified in CF dictionary (%s)", *strf)
		}
		this.stringFilter = string(*strf)
	}

	// StmF streams filter.
	this.streamFilter = "Identity"
	if stmf, ok := (*ed)["StmF"].(*PdfObjectName); ok {
		if _, exists := this.cryptFilters[string(*stmf)]; !exists {
			return fmt.Errorf("Crypt filter for StmF not specified in CF dictionary (%s)", *stmf)
		}
		this.streamFilter = string(*stmf)
	}

	return nil
}

// Prepare the document crypt handler based on the encryption dictionary
// and trailer dictionary.
func PdfCryptMakeNew(ed, trailer *PdfObjectDictionary) (PdfCrypt, error) {
	crypter := PdfCrypt{}
	crypter.decryptedObjects = map[PdfObject]bool{}
	crypter.encryptedObjects = map[PdfObject]bool{}
	crypter.authenticated = false

	filter, ok := (*ed)["Filter"].(*PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR Crypt dictionary missing required Filter field!")
		return crypter, errors.New("Required crypt field Filter missing")
	}
	if *filter != "Standard" {
		common.Log.Debug("ERROR Unsupported filter (%s)", *filter)
		return crypter, errors.New("Unsupported Filter")
	}
	crypter.Filter = string(*filter)

	subfilter, ok := (*ed)["SubFilter"].(*PdfObjectString)
	if ok {
		crypter.Subfilter = string(*subfilter)
		common.Log.Debug("Using subfilter %s", subfilter)
	}

	if L, ok := (*ed)["Length"].(*PdfObjectInteger); ok {
		if (*L % 8) != 0 {
			common.Log.Debug("ERROR Invalid encryption length")
			return crypter, errors.New("Invalid encryption length")
		}
		crypter.length = int(*L)
	} else {
		crypter.length = 40
	}

	V, ok := (*ed)["V"].(*PdfObjectInteger)
	if ok {
		if *V >= 1 && *V <= 2 {
			crypter.V = int(*V)
			// Default algorithm is V2.
			crypter.cryptFilters = CryptFilters{}
			crypter.cryptFilters["Default"] = CryptFilter{cfm: "V2", length: crypter.length}
		} else if *V == 4 {
			crypter.V = int(*V)
			if err := crypter.LoadCryptFilters(ed); err != nil {
				return crypter, err
			}
		} else {
			common.Log.Debug("ERROR Unsupported encryption algo V = %d", *V)
			return crypter, errors.New("Unsupported algorithm")
		}
	} else {
		crypter.V = 0
	}

	R, ok := (*ed)["R"].(*PdfObjectInteger)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing R")
	}
	if *R < 2 || *R > 4 {
		return crypter, errors.New("Invalid R")
	}
	crypter.R = int(*R)

	O, ok := (*ed)["O"].(*PdfObjectString)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing O")
	}
	if len(*O) != 32 {
		return crypter, fmt.Errorf("Length(O) != 32 (%d)", len(*O))
	}
	crypter.O = []byte(*O)

	U, ok := (*ed)["U"].(*PdfObjectString)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing U")
	}
	if len(*U) != 32 {
		// Strictly this does not cause an error.
		// If O is OK and others then can still read the file.
		common.Log.Debug("Warning: Length(U) != 32 (%d)", len(*U))
		//return crypter, errors.New("Length(U) != 32")
	}
	crypter.U = []byte(*U)

	P, ok := (*ed)["P"].(*PdfObjectInteger)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing permissions attr")
	}
	crypter.P = int(*P)

	em, ok := (*ed)["EncryptMetadata"].(*PdfObjectBool)
	if ok {
		crypter.encryptMetadata = bool(*em)
	} else {
		crypter.encryptMetadata = true // True by default.
	}

	// Default: empty ID.
	// Strictly, if file is encrypted, the ID should always be specified
	// but clearly not everyone is following the specification.
	id0 := PdfObjectString("")
	if idArray, ok := (*trailer)["ID"].(*PdfObjectArray); ok {
		common.Log.Debug("Trailer ID array missing!")
		id0obj, ok := (*idArray)[0].(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Invalid trailer ID")
		}
		id0 = *id0obj
	}
	crypter.id0 = string(id0)

	return crypter, nil
}

func (this *PdfCrypt) GetAccessPermissions() AccessPermissions {
	perms := AccessPermissions{}

	P := this.P
	if P&(1<<2) > 0 {
		perms.Printing = true
	}
	if P&(1<<3) > 0 {
		perms.Modify = true
	}
	if P&(1<<4) > 0 {
		perms.ExtractGraphics = true
	}
	if P&(1<<5) > 0 {
		perms.Annotate = true
	}
	if P&(1<<8) > 0 {
		perms.FillForms = true
	}
	if P&(1<<9) > 0 {
		perms.DisabilityExtract = true
	}
	if P&(1<<10) > 0 {
		perms.RotateInsert = true
	}
	if P&(1<<11) > 0 {
		perms.LimitPrintQuality = true
	}
	return perms
}

func (perms AccessPermissions) GetP() int32 {
	var P int32 = 0

	if perms.Printing { // bit 3
		P |= (1 << 2)
	}
	if perms.Modify { // bit 4
		P |= (1 << 3)
	}
	if perms.ExtractGraphics { // bit 5
		P |= (1 << 4)
	}
	if perms.Annotate { // bit 6
		P |= (1 << 5)
	}
	if perms.FillForms {
		P |= (1 << 8) // bit 9
	}
	if perms.DisabilityExtract {
		P |= (1 << 9) // bit 10, what means?
	}
	if perms.RotateInsert {
		P |= (1 << 10) // bit 11
	}
	if perms.LimitPrintQuality {
		P |= (1 << 11) // bit 12
	}
	return P
}

// Check whether the specified password can be used to decrypt the
// document.
func (this *PdfCrypt) authenticate(password []byte) (bool, error) {
	// Also build the encryption/decryption key.

	this.authenticated = false

	// Try user password.
	common.Log.Debug("Debugging authentication - user pass")
	authenticated, err := this.alg6(password)
	if err != nil {
		return false, err
	}
	if authenticated {
		common.Log.Debug("this.authenticated = True")
		this.authenticated = true
		return true, nil
	}

	// Try owner password also.
	// May not be necessary if only want to get all contents.
	// (user pass needs to be known or empty).
	common.Log.Debug("Debugging authentication - owner pass")
	authenticated, err = this.alg7(password, password)
	if err != nil {
		return false, err
	}
	if authenticated {
		common.Log.Debug("this.authenticated = True")
		this.authenticated = true
		return true, nil
	}

	return false, nil
}

func (this *PdfCrypt) paddedPass(pass []byte) []byte {
	key := make([]byte, 32)
	if len(pass) >= 32 {
		for i := 0; i < 32; i++ {
			key[i] = pass[i]
		}
	} else {
		for i := 0; i < len(pass); i++ {
			key[i] = pass[i]
		}
		for i := len(pass); i < 32; i++ {
			key[i] = padding[i-len(pass)]
		}
	}
	return key
}

// Generates a key for encrypting a specific object based on the
// object and generation number, as well as the document encryption key.
func (this *PdfCrypt) makeKey(filter string, objNum, genNum uint32, ekey []byte) ([]byte, error) {
	cf, ok := this.cryptFilters[filter]
	if !ok {
		common.Log.Debug("ERROR Unsupported crypt filter (%s)", filter)
		return nil, fmt.Errorf("Unsupported crypt filter (%s)", filter)
	}
	isAES := false
	if cf.cfm == "AESV2" {
		isAES = true
	}

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
	} else {
		return hashb, nil
	}
}

// Check if object has already been processed.
func (this *PdfCrypt) isDecrypted(obj PdfObject) bool {
	_, ok := this.decryptedObjects[obj]
	if ok {
		common.Log.Debug("Already decrypted")
		return true
	} else {
		common.Log.Debug("Not decrypted yet")
		return false
	}
}

// Decrypt a buffer with a selected crypt filter.
func (this *PdfCrypt) decryptBytes(buf []byte, filter string, okey []byte) ([]byte, error) {
	common.Log.Debug("Decrypt bytes")
	cf, ok := this.cryptFilters[filter]
	if !ok {
		common.Log.Debug("ERROR Unsupported crypt filter (%s)", filter)
		return nil, fmt.Errorf("Unsupported crypt filter (%s)", filter)
	}

	cfMethod := cf.cfm
	if cfMethod == "V2" {
		// Standard RC4 algorithm.
		ciph, err := rc4.NewCipher(okey)
		if err != nil {
			return nil, err
		}
		common.Log.Debug("RC4 Decrypt: % x", buf)
		ciph.XORKeyStream(buf, buf)
		common.Log.Debug("to: % x", buf)
		return buf, nil
	} else if cfMethod == "AESV2" {
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

		common.Log.Debug("AES Decrypt (%d): % x", len(buf), buf)
		common.Log.Debug("chop AES Decrypt (%d): % x", len(buf), buf)
		mode.CryptBlocks(buf, buf)
		common.Log.Debug("to (%d): % x", len(buf), buf)
		//copy(buf[0:], buf[16:])
		//common.Log.Debug("chop to (%d): % x", len(buf), buf)
		return buf, nil
	}
	return nil, fmt.Errorf("Unsupported crypt filter method (%s)", cfMethod)
}

// Decrypt an object with specified key. For numbered objects,
// the key argument is not used and a new one is generated based
// on the object and generation number.
// Traverses through all the subobjects (recursive).
//
// Does not look up references..  That should be done prior to calling.
func (this *PdfCrypt) Decrypt(obj PdfObject, parentObjNum, parentGenNum int64) error {
	if this.isDecrypted(obj) {
		return nil
	}

	if io, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		this.decryptedObjects[io] = true

		common.Log.Debug("Decrypting indirect %d %d obj!", io.ObjectNumber, io.GenerationNumber)

		objNum := (*io).ObjectNumber
		genNum := (*io).GenerationNumber

		err := this.Decrypt(io.PdfObject, objNum, genNum)
		if err != nil {
			return err
		}

		return nil
	}

	if so, isStream := obj.(*PdfObjectStream); isStream {
		// Mark as decrypted first to avoid recursive issues.
		this.decryptedObjects[so] = true
		objNum := (*so).ObjectNumber
		genNum := (*so).GenerationNumber
		common.Log.Debug("Decrypting stream %d %d !", objNum, genNum)

		// TODO: Check for crypt filter (V4).
		// The Crypt filter shall be the first filter in the Filter array entry.

		dict := so.PdfObjectDictionary

		streamFilter := "Default" // Default RC4.
		if this.V >= 4 {
			streamFilter = this.streamFilter
			common.Log.Debug("this.streamFilter = %s", this.streamFilter)

			if filters, ok := (*dict)["Filter"].(*PdfObjectArray); ok {
				// Crypt filter can only be the first entry.
				if firstFilter, ok := (*filters)[0].(*PdfObjectName); ok {
					if *firstFilter == "Crypt" {
						// Crypt filter overriding the default.
						// Default option is Identity.
						streamFilter = "Identity"

						// Check if valid crypt filter specified in the decode params.
						if decodeParams, ok := (*dict)["DecodeParms"].(*PdfObjectDictionary); ok {
							if filterName, ok := (*decodeParams)["Name"].(*PdfObjectName); ok {
								if _, ok := this.cryptFilters[string(*filterName)]; ok {
									common.Log.Debug("Using stream filter %s", *filterName)
									streamFilter = string(*filterName)
								}
							}
						}
					}
				}
			}

			common.Log.Debug("with %s filter", streamFilter)
			if streamFilter == "Identity" {
				// Identity: pass unchanged.
				return nil
			}
		}

		err := this.Decrypt(so.PdfObjectDictionary, objNum, genNum)
		if err != nil {
			return err
		}

		okey, err := this.makeKey(streamFilter, uint32(objNum), uint32(genNum), this.encryptionKey)
		if err != nil {
			return err
		}

		so.Stream, err = this.decryptBytes(so.Stream, streamFilter, okey)
		if err != nil {
			return err
		}
		// Update the length based on the decrypted stream.
		(*dict)["Length"] = MakeInteger(int64(len(so.Stream)))

		return nil
	}
	if s, isString := obj.(*PdfObjectString); isString {
		common.Log.Debug("Decrypting string!")

		stringFilter := "Default"
		if this.V >= 4 {
			// Currently only support Identity / RC4.
			common.Log.Debug("with %s filter", this.stringFilter)
			if this.stringFilter == "Identity" {
				// Identity: pass unchanged: No action.
				return nil
			} else {
				stringFilter = this.stringFilter
			}
		}

		key, err := this.makeKey(stringFilter, uint32(parentObjNum), uint32(parentGenNum), this.encryptionKey)
		if err != nil {
			return err
		}

		// Overwrite the encrypted with decrypted string.
		decrypted := make([]byte, len(*s))
		for i := 0; i < len(*s); i++ {
			decrypted[i] = (*s)[i]
		}
		common.Log.Debug("Decrypt string: %s : % x", decrypted, decrypted)
		decrypted, err = this.decryptBytes(decrypted, stringFilter, key)
		if err != nil {
			return err
		}
		*s = PdfObjectString(decrypted)

		return nil
	}

	if a, isArray := obj.(*PdfObjectArray); isArray {
		for _, o := range *a {
			err := this.Decrypt(o, parentObjNum, parentGenNum)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if d, isDict := obj.(*PdfObjectDictionary); isDict {
		isSig := false
		if t, hasType := (*d)["Type"]; hasType {
			typeStr, ok := t.(*PdfObjectName)
			if ok && *typeStr == "Sig" {
				isSig = true
			}
		}
		for keyidx, o := range *d {
			// How can we avoid this check, i.e. implement a more smart
			// traversal system?
			if isSig && string(keyidx) == "Contents" {
				// Leave the Contents of a Signature dictionary.
				continue
			}

			if string(keyidx) != "Parent" && string(keyidx) != "Prev" && string(keyidx) != "Last" { // Check not needed?
				err := this.Decrypt(o, parentObjNum, parentGenNum)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	return nil
}

// Check if object has already been processed.
func (this *PdfCrypt) isEncrypted(obj PdfObject) bool {
	_, ok := this.encryptedObjects[obj]
	if ok {
		common.Log.Debug("Already encrypted")
		return true
	} else {
		common.Log.Debug("Not encrypted yet")
		return false
	}
}

// Encrypt a buffer with the specified crypt filter and key.
func (this *PdfCrypt) encryptBytes(buf []byte, filter string, okey []byte) ([]byte, error) {
	common.Log.Debug("Encrypt bytes")
	cf, ok := this.cryptFilters[filter]
	if !ok {
		common.Log.Debug("ERROR Unsupported crypt filter (%s)", filter)
		return nil, fmt.Errorf("Unsupported crypt filter (%s)", filter)
	}

	cfMethod := cf.cfm
	if cfMethod == "V2" {
		// Standard RC4 algorithm.
		ciph, err := rc4.NewCipher(okey)
		if err != nil {
			return nil, err
		}
		common.Log.Debug("RC4 Encrypt: % x", buf)
		ciph.XORKeyStream(buf, buf)
		common.Log.Debug("to: % x", buf)
		return buf, nil
	} else if cfMethod == "AESV2" {
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

		common.Log.Debug("AES Encrypt (%d): % x", len(buf), buf)

		// If using the AES algorithm, the Cipher Block Chaining (CBC)
		// mode, which requires an initialization vector, is used. The
		// block size parameter is set to 16 bytes, and the initialization
		// vector is a 16-byte random number that is stored as the first
		// 16 bytes of the encrypted stream or string.
		pad := 16 - len(buf)%16
		for i := 0; i < pad; i++ {
			buf = append(buf, byte(pad))
		}
		common.Log.Debug("Padded to %d bytes", len(buf))

		// Generate random 16 bytes, place in beginning of buffer.
		ciphertext := make([]byte, 16+len(buf))
		iv := ciphertext[:16]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, err
		}

		mode := cipher.NewCBCEncrypter(ciph, iv)
		mode.CryptBlocks(ciphertext[aes.BlockSize:], buf)

		buf = ciphertext
		common.Log.Debug("to (%d): % x", len(buf), buf)

		return buf, nil
	}
	return nil, fmt.Errorf("Unsupported crypt filter method (%s)", cfMethod)
}

// Encrypt an object with specified key. For numbered objects,
// the key argument is not used and a new one is generated based
// on the object and generation number.
// Traverses through all the subobjects (recursive).
//
// Does not look up references..  That should be done prior to calling.
func (this *PdfCrypt) Encrypt(obj PdfObject, parentObjNum, parentGenNum int64) error {
	if this.isEncrypted(obj) {
		return nil
	}

	if io, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		this.encryptedObjects[io] = true

		common.Log.Debug("Encrypting indirect %d %d obj!", io.ObjectNumber, io.GenerationNumber)

		objNum := (*io).ObjectNumber
		genNum := (*io).GenerationNumber

		err := this.Encrypt(io.PdfObject, objNum, genNum)
		if err != nil {
			return err
		}

		return nil
	}

	if so, isStream := obj.(*PdfObjectStream); isStream {
		this.encryptedObjects[so] = true
		objNum := (*so).ObjectNumber
		genNum := (*so).GenerationNumber
		common.Log.Debug("Encrypting stream %d %d !", objNum, genNum)

		// TODO: Check for crypt filter (V4).
		// The Crypt filter shall be the first filter in the Filter array entry.

		dict := so.PdfObjectDictionary

		streamFilter := "Default" // Default RC4.
		if this.V >= 4 {
			// For now.  Need to change when we add support for more than
			// Identity / RC4.
			streamFilter = this.streamFilter
			common.Log.Debug("this.streamFilter = %s", this.streamFilter)

			if filters, ok := (*dict)["Filter"].(*PdfObjectArray); ok {
				// Crypt filter can only be the first entry.
				if firstFilter, ok := (*filters)[0].(*PdfObjectName); ok {
					if *firstFilter == "Crypt" {
						// Crypt filter overriding the default.
						// Default option is Identity.
						streamFilter = "Identity"

						// Check if valid crypt filter specified in the decode params.
						if decodeParams, ok := (*dict)["DecodeParms"].(*PdfObjectDictionary); ok {
							if filterName, ok := (*decodeParams)["Name"].(*PdfObjectName); ok {
								if _, ok := this.cryptFilters[string(*filterName)]; ok {
									common.Log.Debug("Using stream filter %s", *filterName)
									streamFilter = string(*filterName)
								}
							}
						}
					}
				}
			}

			common.Log.Debug("with %s filter", streamFilter)
			if streamFilter == "Identity" {
				// Identity: pass unchanged.
				return nil
			}
		}

		err := this.Encrypt(so.PdfObjectDictionary, objNum, genNum)
		if err != nil {
			return err
		}

		okey, err := this.makeKey(streamFilter, uint32(objNum), uint32(genNum), this.encryptionKey)
		if err != nil {
			return err
		}

		so.Stream, err = this.encryptBytes(so.Stream, streamFilter, okey)
		if err != nil {
			return err
		}
		// Update the length based on the encrypted stream.
		(*dict)["Length"] = MakeInteger(int64(len(so.Stream)))

		return nil
	}
	if s, isString := obj.(*PdfObjectString); isString {
		common.Log.Debug("Encrypting string!")

		stringFilter := "Default"
		if this.V >= 4 {
			common.Log.Debug("with %s filter", this.stringFilter)
			if this.stringFilter == "Identity" {
				// Identity: pass unchanged: No action.
				return nil
			} else {
				stringFilter = this.stringFilter
			}
		}

		key, err := this.makeKey(stringFilter, uint32(parentObjNum), uint32(parentGenNum), this.encryptionKey)
		if err != nil {
			return err
		}

		encrypted := make([]byte, len(*s))
		for i := 0; i < len(*s); i++ {
			encrypted[i] = (*s)[i]
		}
		common.Log.Debug("Encrypt string: %s : % x", encrypted, encrypted)
		encrypted, err = this.encryptBytes(encrypted, stringFilter, key)
		if err != nil {
			return err
		}
		*s = PdfObjectString(encrypted)

		return nil
	}

	if a, isArray := obj.(*PdfObjectArray); isArray {
		for _, o := range *a {
			err := this.Encrypt(o, parentObjNum, parentGenNum)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if d, isDict := obj.(*PdfObjectDictionary); isDict {
		isSig := false
		if t, hasType := (*d)["Type"]; hasType {
			typeStr, ok := t.(*PdfObjectName)
			if ok && *typeStr == "Sig" {
				isSig = true
			}
		}

		for keyidx, o := range *d {
			// How can we avoid this check, i.e. implement a more smart
			// traversal system?
			if isSig && string(keyidx) == "Contents" {
				// Leave the Contents of a Signature dictionary.
				continue
			}
			if string(keyidx) != "Parent" && string(keyidx) != "Prev" && string(keyidx) != "Last" { // Check not needed?
				err := this.Encrypt(o, parentObjNum, parentGenNum)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	return nil
}

// Algorithm 2: Computing an encryption key.
func (this *PdfCrypt) alg2(pass []byte) []byte {
	common.Log.Debug("Alg2")
	key := this.paddedPass(pass)

	h := md5.New()
	h.Write(key)

	// Pass O.
	h.Write(this.O)

	// Pass P (Lower order byte first).
	var p uint32 = uint32(this.P)
	var pb = []byte{}
	for i := 0; i < 4; i++ {
		pb = append(pb, byte(((p >> uint(8*i)) & 0xff)))
	}
	h.Write(pb)
	common.Log.Debug("go P: % x", pb)

	// Pass ID[0] from the trailer
	h.Write([]byte(this.id0))

	common.Log.Debug("this.R = %d encryptMetadata %v", this.R, this.encryptMetadata)
	if (this.R >= 4) && !this.encryptMetadata {
		h.Write([]byte{0xff, 0xff, 0xff, 0xff})
	}
	hashb := h.Sum(nil)

	if this.R >= 3 {
		for i := 0; i < 50; i++ {
			h = md5.New()
			h.Write(hashb[0 : this.length/8])
			hashb = h.Sum(nil)
		}
	}

	if this.R >= 3 {
		return hashb[0 : this.length/8]
	}

	return hashb[0:5]
}

// Create the RC4 encryption key.
func (this *PdfCrypt) alg3_key(pass []byte) []byte {
	h := md5.New()
	okey := this.paddedPass(pass)
	h.Write(okey)

	if this.R >= 3 {
		for i := 0; i < 50; i++ {
			hashb := h.Sum(nil)
			h = md5.New()
			h.Write(hashb)
		}
	}

	encKey := h.Sum(nil)
	if this.R == 2 {
		encKey = encKey[0:5]
	} else {
		encKey = encKey[0 : this.length/8]
	}
	return encKey
}

// Algorithm 3: Computing the encryption dictionary’s O
// (owner password) value.
func (this *PdfCrypt) alg3(upass, opass []byte) (PdfObjectString, error) {
	// Return O string val.
	O := PdfObjectString("")

	var encKey []byte
	if len(opass) > 0 {
		encKey = this.alg3_key(opass)
	} else {
		encKey = this.alg3_key(upass)
	}

	ociph, err := rc4.NewCipher(encKey)
	if err != nil {
		return O, errors.New("Failed rc4 ciph")
	}

	ukey := this.paddedPass(upass)
	encrypted := make([]byte, len(ukey))
	ociph.XORKeyStream(encrypted, ukey)

	if this.R >= 3 {
		encKey2 := make([]byte, len(encKey))
		for i := 0; i < 19; i++ {
			for j := 0; j < len(encKey); j++ {
				encKey2[j] = encKey[j] ^ byte(i+1)
			}
			ciph, err := rc4.NewCipher(encKey2)
			if err != nil {
				return O, errors.New("Failed rc4 ciph")
			}
			ciph.XORKeyStream(encrypted, encrypted)
		}
	}

	O = PdfObjectString(encrypted)
	return O, nil
}

// Algorithm 4: Computing the encryption dictionary’s U (user password)
// value (Security handlers of revision 2).
func (this *PdfCrypt) alg4(upass []byte) (PdfObjectString, []byte, error) {
	U := PdfObjectString("")

	ekey := this.alg2(upass)
	ciph, err := rc4.NewCipher(ekey)
	if err != nil {
		return U, ekey, errors.New("Failed rc4 ciph")
	}

	s := []byte(padding)
	encrypted := make([]byte, len(s))
	ciph.XORKeyStream(encrypted, s)

	U = PdfObjectString(encrypted)
	return U, ekey, nil
}

// Algorithm 5: Computing the encryption dictionary’s U (user password)
// value (Security handlers of revision 3 or greater).
func (this *PdfCrypt) alg5(upass []byte) (PdfObjectString, []byte, error) {
	U := PdfObjectString("")

	ekey := this.alg2(upass)

	h := md5.New()
	h.Write([]byte(padding))
	h.Write([]byte(this.id0))
	hash := h.Sum(nil)

	common.Log.Debug("Alg5")
	common.Log.Debug("ekey: % x", ekey)
	common.Log.Debug("ID: % x", this.id0)

	if len(hash) != 16 {
		return U, ekey, errors.New("Hash length not 16 bytes")
	}

	ciph, err := rc4.NewCipher(ekey)
	if err != nil {
		return U, ekey, errors.New("Failed rc4 ciph")
	}
	encrypted := make([]byte, 16)
	ciph.XORKeyStream(encrypted, hash)

	// Do the following 19 times: Take the output from the previous
	// invocation of the RC4 function and pass it as input to a new
	// invocation of the function; use an encryption key generated by
	// taking each byte of the original encryption key obtained in step
	// (a) and performing an XOR (exclusive or) operation between that
	// byte and the single-byte value of the iteration counter (from 1 to 19).
	ekey2 := make([]byte, len(ekey))
	for i := 0; i < 19; i++ {
		for j := 0; j < len(ekey); j++ {
			ekey2[j] = ekey[j] ^ byte(i+1)
		}
		ciph, err = rc4.NewCipher(ekey2)
		if err != nil {
			return U, ekey, errors.New("Failed rc4 ciph")
		}
		ciph.XORKeyStream(encrypted, encrypted)
		common.Log.Debug("i = %d, ekey: % x", i, ekey2)
		common.Log.Debug("i = %d -> % x", i, encrypted)
	}

	bb := make([]byte, 32)
	for i := 0; i < 16; i++ {
		bb[i] = encrypted[i]
	}

	// Append 16 bytes of arbitrary padding to the output from the final
	// invocation of the RC4 function and store the 32-byte result as
	// the value of the U entry in the encryption dictionary.
	_, err = rand.Read(bb[16:32])
	if err != nil {
		return U, ekey, errors.New("Failed to gen rand number")
	}

	U = PdfObjectString(bb)
	return U, ekey, nil
}

// Algorithm 6: Authenticating the user password
func (this *PdfCrypt) alg6(upass []byte) (bool, error) {
	var uo PdfObjectString
	var err error
	var key []byte
	if this.R == 2 {
		uo, key, err = this.alg4(upass)
	} else if this.R >= 3 {
		uo, key, err = this.alg5(upass)
	} else {
		return false, errors.New("invalid R")
	}

	if err != nil {
		return false, err
	}

	common.Log.Debug("check: % x == % x ?", string(uo), string(this.U))

	uGen := string(uo)     // Generated U from specified pass.
	uDoc := string(this.U) // U from the document.
	if this.R >= 3 {
		// comparing on the first 16 bytes in the case of security
		// handlers of revision 3 or greater),
		uGen = uGen[0:16]
		uDoc = uDoc[0:16]
	}
	if uGen == uDoc {
		this.encryptionKey = key
		return true, nil
	} else {
		return false, nil
	}
}

// Algorithm 7: Authenticating the owner password.
func (this *PdfCrypt) alg7(upass, opass []byte) (bool, error) {
	encKey := this.alg3_key(opass)

	decrypted := make([]byte, len(this.O))
	if this.R == 2 {
		ciph, err := rc4.NewCipher(encKey)
		if err != nil {
			return false, errors.New("Failed cipher")
		}
		ciph.XORKeyStream(decrypted, this.O)
	} else if this.R >= 3 {
		s := this.O
		newKey := encKey
		for i := 0; i < 20; i++ {
			for j := 0; j < len(encKey); j++ {
				newKey[j] ^= byte(i)
			}
			ciph, err := rc4.NewCipher(newKey)
			if err != nil {
				return false, errors.New("Failed cipher")
			}
			ciph.XORKeyStream(decrypted, s)
			s = decrypted
		}
	} else {
		return false, errors.New("invalid R")
	}

	if string(decrypted) == string(upass) {
		// Correct.
		return true, nil
	} else {
		return false, nil
	}
}
