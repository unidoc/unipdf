/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/unidoc/unidoc/common"
)

type Version struct {
	Major int
	Minor int
}

type EncryptInfo struct {
	Version
	Encrypt  *PdfObjectDictionary
	ID0, ID1 string
}

// PdfCryptNewEncrypt makes the document crypt handler based on a specified crypt filter.
func PdfCryptNewEncrypt(cf CryptFilter, userPass, ownerPass []byte, perm AccessPermissions) (*PdfCrypt, *EncryptInfo, error) {
	crypter := &PdfCrypt{
		encryptedObjects: make(map[PdfObject]bool),
		cryptFilters:     make(cryptFilters),
		encryptStd: stdEncryptDict{
			P:               perm,
			EncryptMetadata: true,
		},
	}
	// TODO(dennwc): define it in the CF interface
	var vers Version
	switch cf.(type) {
	case cryptFilterV2:
		crypter.encrypt.V = 2
		crypter.encryptStd.R = 3
	case cryptFilterAESV2:
		vers.Major, vers.Minor = 1, 5
		crypter.encrypt.V = 4
		crypter.encryptStd.R = 4
	case cryptFilterAESV3:
		vers.Major, vers.Minor = 2, 0
		crypter.encrypt.V = 5
		crypter.encryptStd.R = 6 // TODO(dennwc): a way to set R=5?
	}
	if cf != nil {
		crypter.encrypt.Length = cf.KeyLength() * 8
	}
	const (
		defaultFilter = StandardCryptFilter
	)
	crypter.cryptFilters[defaultFilter] = cf
	if crypter.encrypt.V >= 4 {
		crypter.streamFilter = defaultFilter
		crypter.stringFilter = defaultFilter
	}
	ed := crypter.newEncyptDict()

	// Prepare the ID object for the trailer.
	hashcode := md5.Sum([]byte(time.Now().Format(time.RFC850)))
	id0 := string(hashcode[:])
	b := make([]byte, 100)
	rand.Read(b)
	hashcode = md5.Sum(b)
	id1 := string(hashcode[:])
	common.Log.Trace("Random b: % x", b)

	common.Log.Trace("Gen Id 0: % x", id0)

	crypter.id0 = string(id0)

	err := crypter.generateParams(userPass, ownerPass)
	if err != nil {
		return nil, nil, err
	}
	// Generate encryption parameters
	if crypter.encryptStd.R < 5 {
		ed.Set("O", MakeString(string(crypter.encryptStd.O)))
		ed.Set("U", MakeString(string(crypter.encryptStd.U)))
	} else { // R >= 5
		ed.Set("O", MakeString(string(crypter.encryptStd.O)))
		ed.Set("U", MakeString(string(crypter.encryptStd.U)))
		ed.Set("OE", MakeString(string(crypter.encryptStd.OE)))
		ed.Set("UE", MakeString(string(crypter.encryptStd.UE)))
		ed.Set("EncryptMetadata", MakeBool(crypter.encryptStd.EncryptMetadata))
		if crypter.encryptStd.R > 5 {
			ed.Set("Perms", MakeString(string(crypter.encryptStd.Perms)))
		}
	}
	if crypter.encrypt.V >= 4 {
		if err := crypter.saveCryptFilters(ed); err != nil {
			return nil, nil, err
		}
	}

	return crypter, &EncryptInfo{
		Version: vers,
		Encrypt: ed,
		ID0:     id0, ID1: id1,
	}, nil
}

// PdfCrypt provides PDF encryption/decryption support.
// The PDF standard supports encryption of strings and streams (Section 7.6).
// TODO (v3): Consider unexporting.
type PdfCrypt struct {
	encrypt    encryptDict
	encryptStd stdEncryptDict

	id0              string
	encryptionKey    []byte
	decryptedObjects map[PdfObject]bool
	encryptedObjects map[PdfObject]bool
	authenticated    bool
	// Crypt filters (V4).
	cryptFilters cryptFilters
	streamFilter string
	stringFilter string

	parser *PdfParser

	decryptedObjNum map[int]struct{}
	ivAESZero       []byte // a zero buffer used as an initialization vector for AES
}

func (crypt *PdfCrypt) newEncyptDict() *PdfObjectDictionary {
	// Generate the encryption dictionary.
	ed := MakeDict()
	ed.Set("Filter", MakeName("Standard"))
	ed.Set("V", MakeInteger(int64(crypt.encrypt.V)))
	ed.Set("Length", MakeInteger(int64(crypt.encrypt.Length)))
	ed.Set("P", MakeInteger(int64(crypt.encryptStd.P)))
	ed.Set("R", MakeInteger(int64(crypt.encryptStd.R)))
	return ed
}

// String returns a descriptive information string about the encryption method used.
func (crypt *PdfCrypt) String() string {
	if crypt == nil {
		return ""
	}
	// TODO(dennwc): define a String method on CF
	str := crypt.encrypt.Filter + " - "

	if crypt.encrypt.V == 0 {
		str += "Undocumented algorithm"
	} else if crypt.encrypt.V == 1 {
		// RC4 or AES (bits: 40)
		str += "RC4: 40 bits"
	} else if crypt.encrypt.V == 2 {
		str += fmt.Sprintf("RC4: %d bits", crypt.encrypt.Length)
	} else if crypt.encrypt.V == 3 {
		str += "Unpublished algorithm"
	} else if crypt.encrypt.V >= 4 {
		// Look at CF, StmF, StrF
		str += fmt.Sprintf("Stream filter: %s - String filter: %s", crypt.streamFilter, crypt.stringFilter)
		str += "; Crypt filters:"
		for name, cf := range crypt.cryptFilters {
			str += fmt.Sprintf(" - %s: %s (%d)", name, cf.Name(), cf.KeyLength())
		}
	}
	perms := crypt.GetAccessPermissions()
	str += fmt.Sprintf(" - %#v", perms)

	return str
}

type authEvent string

const (
	authEventDocOpen = authEvent("DocOpen")
	authEventEFOpen  = authEvent("EFOpen")
)

type cryptFiltersDict map[string]cryptFilterDict

// encryptDict is a set of field common to all encryption dictionaries.
type encryptDict struct {
	Filter    string           // (Required) The name of the preferred security handler for this document.
	V         int              // (Required) A code specifying the algorithm to be used in encrypting and decrypting the document.
	SubFilter string           // Completely specifies the format and interpretation of the encryption dictionary.
	Length    int              // The length of the encryption key, in bits.
	CF        cryptFiltersDict // Crypt filters dictionary.
	StmF      string           // The filter that shall be used by default when decrypting streams.
	StrF      string           // The filter that shall be used when decrypting all strings in the document.
	EFF       string           // The filter that shall be used when decrypting embedded file streams.
}

// stdEncryptDict is a set of additional fields used in standard encryption dictionary.
type stdEncryptDict struct {
	R      int // (Required) A number specifying which revision of the standard security handler shall be used.
	O, U   []byte
	OE, UE []byte // R=6

	P               AccessPermissions
	Perms           []byte // An encrypted copy of P (16 bytes). Used to verify permissions. R=6
	EncryptMetadata bool   // Indicates whether the document-level metadata stream shall be encrypted.
}

// AccessPermissions is a bitmask of access permissions for a PDF file.
type AccessPermissions uint32

const (
	// PermOwner grants all permissions.
	PermOwner = AccessPermissions(math.MaxUint32)

	PermPrinting        = AccessPermissions(1 << 2) // bit 3
	PermModify          = AccessPermissions(1 << 3) // bit 4
	PermExtractGraphics = AccessPermissions(1 << 4) // bit 5
	PermAnnotate        = AccessPermissions(1 << 5) // bit 6
	// PermFillForms allow form filling, if annotation is disabled?  If annotation enabled, is not looked at.
	PermFillForms         = AccessPermissions(1 << 8) // bit 9
	PermDisabilityExtract = AccessPermissions(1 << 9) // bit 10 // TODO: not clear what this means!
	// PermRotateInsert allows rotating, editing page order.
	PermRotateInsert = AccessPermissions(1 << 10) // bit 11
	// PermFullPrintQuality limits print quality (lowres), assuming Printing bit is set.
	PermFullPrintQuality = AccessPermissions(1 << 11) // bit 12
)

// Allowed checks if a set of permissions can be granted.
func (p AccessPermissions) Allowed(p2 AccessPermissions) bool {
	return p&p2 == p2
}

const padding = "\x28\xBF\x4E\x5E\x4E\x75\x8A\x41\x64\x00\x4E\x56\xFF" +
	"\xFA\x01\x08\x2E\x2E\x00\xB6\xD0\x68\x3E\x80\x2F\x0C" +
	"\xA9\xFE\x64\x53\x69\x7A"

// StandardCryptFilter is a default name for a standard crypt filter.
const StandardCryptFilter = "StdCF"

func newCryptFiltersV2(length int) cryptFilters {
	return cryptFilters{
		StandardCryptFilter: NewCryptFilterV2(length),
	}
}

// NewCryptFilterV2 creates a RC4-based filter with a specified key length (in bytes).
func NewCryptFilterV2(length int) CryptFilter {
	f, err := newCryptFilterV2(cryptFilterDict{Length: length})
	if err != nil {
		panic(err)
	}
	return f
}

// NewCryptFilterAESV2 creates an AES-based filter with a 128 bit key (AESV2).
func NewCryptFilterAESV2() CryptFilter {
	f, err := newCryptFilterAESV2(cryptFilterDict{})
	if err != nil {
		panic(err)
	}
	return f
}

// NewCryptFilterAESV3 creates an AES-based filter with a 256 bit key (AESV3).
func NewCryptFilterAESV3() CryptFilter {
	f, err := newCryptFilterAESV3(cryptFilterDict{})
	if err != nil {
		panic(err)
	}
	return f
}

// cryptFilters is a map of crypt filter name and underlying CryptFilter info.
type cryptFilters map[string]CryptFilter

// loadCryptFilters loads crypt filter information from the encryption dictionary (V>=4).
func (crypt *PdfCrypt) loadCryptFilters(ed *PdfObjectDictionary) error {
	crypt.cryptFilters = cryptFilters{}

	obj := ed.Get("CF")
	obj = TraceToDirectObject(obj) // XXX may need to resolve reference...
	if ref, isRef := obj.(*PdfObjectReference); isRef {
		o, err := crypt.parser.LookupByReference(*ref)
		if err != nil {
			common.Log.Debug("Error looking up CF reference")
			return err
		}
		obj = TraceToDirectObject(o)
	}

	cf, ok := obj.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Invalid CF, type: %T", obj)
		return errors.New("Invalid CF")
	}

	for _, name := range cf.Keys() {
		v := cf.Get(name)

		if ref, isRef := v.(*PdfObjectReference); isRef {
			o, err := crypt.parser.LookupByReference(*ref)
			if err != nil {
				common.Log.Debug("Error lookup up dictionary reference")
				return err
			}
			v = TraceToDirectObject(o)
		}

		dict, ok := v.(*PdfObjectDictionary)
		if !ok {
			return fmt.Errorf("Invalid dict in CF (name %s) - not a dictionary but %T", name, v)
		}

		if name == "Identity" {
			common.Log.Debug("ERROR - Cannot overwrite the identity filter - Trying next")
			continue
		}

		var cfd cryptFilterDict
		if err := cfd.ReadFrom(dict); err != nil {
			return err
		}
		fnc, err := getCryptFilterMethod(cfd.CFM)
		if err != nil {
			return err
		}
		cf, err := fnc(cfd)
		if err != nil {
			return err
		}
		crypt.cryptFilters[string(name)] = cf
	}
	// Cannot be overwritten.
	crypt.cryptFilters["Identity"] = cryptFilteridentity{}

	// StrF strings filter.
	crypt.stringFilter = "Identity"
	if strf, ok := ed.Get("StrF").(*PdfObjectName); ok {
		if _, exists := crypt.cryptFilters[string(*strf)]; !exists {
			return fmt.Errorf("Crypt filter for StrF not specified in CF dictionary (%s)", *strf)
		}
		crypt.stringFilter = string(*strf)
	}

	// StmF streams filter.
	crypt.streamFilter = "Identity"
	if stmf, ok := ed.Get("StmF").(*PdfObjectName); ok {
		if _, exists := crypt.cryptFilters[string(*stmf)]; !exists {
			return fmt.Errorf("Crypt filter for StmF not specified in CF dictionary (%s)", *stmf)
		}
		crypt.streamFilter = string(*stmf)
	}

	return nil
}

// saveCryptFilters saves crypt filter information to the encryption dictionary (V>=4).
func (crypt *PdfCrypt) saveCryptFilters(ed *PdfObjectDictionary) error {
	if crypt.encrypt.V < 4 {
		return errors.New("can only be used with V>=4")
	}
	cf := MakeDict()
	ed.Set("CF", cf)

	for name, filter := range crypt.cryptFilters {
		if name == "Identity" {
			continue
		}
		v := cryptFilterToDict(filter, "")
		cf.Set(PdfObjectName(name), v)
	}
	ed.Set("StrF", MakeName(crypt.stringFilter))
	ed.Set("StmF", MakeName(crypt.streamFilter))
	return nil
}

// PdfCryptNewDecrypt makes the document crypt handler based on the encryption dictionary
// and trailer dictionary. Returns an error on failure to process.
func PdfCryptNewDecrypt(parser *PdfParser, ed, trailer *PdfObjectDictionary) (*PdfCrypt, error) {
	crypter := &PdfCrypt{
		authenticated:    false,
		decryptedObjects: make(map[PdfObject]bool),
		encryptedObjects: make(map[PdfObject]bool),
		decryptedObjNum:  make(map[int]struct{}),
		parser:           parser,
	}

	filter, ok := ed.Get("Filter").(*PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR Crypt dictionary missing required Filter field!")
		return crypter, errors.New("Required crypt field Filter missing")
	}
	if *filter != "Standard" {
		common.Log.Debug("ERROR Unsupported filter (%s)", *filter)
		return crypter, errors.New("Unsupported Filter")
	}
	crypter.encrypt.Filter = string(*filter)

	if subfilter, ok := ed.Get("SubFilter").(*PdfObjectString); ok {
		crypter.encrypt.SubFilter = subfilter.Str()
		common.Log.Debug("Using subfilter %s", subfilter)
	}

	if L, ok := ed.Get("Length").(*PdfObjectInteger); ok {
		if (*L % 8) != 0 {
			common.Log.Debug("ERROR Invalid encryption length")
			return crypter, errors.New("Invalid encryption length")
		}
		crypter.encrypt.Length = int(*L)
	} else {
		crypter.encrypt.Length = 40
	}

	crypter.encrypt.V = 0
	if v, ok := ed.Get("V").(*PdfObjectInteger); ok {
		V := int(*v)
		crypter.encrypt.V = V
		if V >= 1 && V <= 2 {
			// Default algorithm is V2.
			crypter.cryptFilters = newCryptFiltersV2(crypter.encrypt.Length)
		} else if V >= 4 && V <= 5 {
			if err := crypter.loadCryptFilters(ed); err != nil {
				return crypter, err
			}
		} else {
			common.Log.Debug("ERROR Unsupported encryption algo V = %d", V)
			return crypter, errors.New("Unsupported algorithm")
		}
	}

	R, ok := ed.Get("R").(*PdfObjectInteger)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing R")
	}
	// TODO(dennwc): according to spec, R should be validated according to V value
	if *R < 2 || *R > 6 {
		return crypter, fmt.Errorf("Invalid R (%d)", *R)
	}
	crypter.encryptStd.R = int(*R)

	O, ok := ed.Get("O").(*PdfObjectString)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing O")
	}
	if crypter.encryptStd.R == 5 || crypter.encryptStd.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(O.Str()) < 48 {
			return crypter, fmt.Errorf("Length(O) < 48 (%d)", len(O.Str()))
		}
	} else if len(O.Str()) != 32 {
		return crypter, fmt.Errorf("Length(O) != 32 (%d)", len(O.Str()))
	}
	crypter.encryptStd.O = O.Bytes()

	U, ok := ed.Get("U").(*PdfObjectString)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing U")
	}
	if crypter.encryptStd.R == 5 || crypter.encryptStd.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(U.Str()) < 48 {
			return crypter, fmt.Errorf("Length(U) < 48 (%d)", len(U.Str()))
		}
	} else if len(U.Str()) != 32 {
		// Strictly this does not cause an error.
		// If O is OK and others then can still read the file.
		common.Log.Debug("Warning: Length(U) != 32 (%d)", len(U.Str()))
		//return crypter, errors.New("Length(U) != 32")
	}
	crypter.encryptStd.U = U.Bytes()

	if crypter.encryptStd.R >= 5 {
		OE, ok := ed.Get("OE").(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Encrypt dictionary missing OE")
		}
		if len(OE.Str()) != 32 {
			return crypter, fmt.Errorf("Length(OE) != 32 (%d)", len(OE.Str()))
		}
		crypter.encryptStd.OE = OE.Bytes()

		UE, ok := ed.Get("UE").(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Encrypt dictionary missing UE")
		}
		if len(UE.Str()) != 32 {
			return crypter, fmt.Errorf("Length(UE) != 32 (%d)", len(UE.Str()))
		}
		crypter.encryptStd.UE = UE.Bytes()
	}

	P, ok := ed.Get("P").(*PdfObjectInteger)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing permissions attr")
	}
	crypter.encryptStd.P = AccessPermissions(*P)

	if crypter.encryptStd.R == 6 {
		Perms, ok := ed.Get("Perms").(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Encrypt dictionary missing Perms")
		}
		if len(Perms.Str()) != 16 {
			return crypter, fmt.Errorf("Length(Perms) != 16 (%d)", len(Perms.Str()))
		}
		crypter.encryptStd.Perms = Perms.Bytes()
	}

	em, ok := ed.Get("EncryptMetadata").(*PdfObjectBool)
	if ok {
		crypter.encryptStd.EncryptMetadata = bool(*em)
	} else {
		crypter.encryptStd.EncryptMetadata = true // True by default.
	}

	// Default: empty ID.
	// Strictly, if file is encrypted, the ID should always be specified
	// but clearly not everyone is following the specification.
	id0 := ""
	if idArray, ok := trailer.Get("ID").(*PdfObjectArray); ok && idArray.Len() >= 1 {
		id0obj, ok := GetString(idArray.Get(0))
		if !ok {
			return crypter, errors.New("Invalid trailer ID")
		}
		id0 = id0obj.Str()
	} else {
		common.Log.Debug("Trailer ID array missing or invalid!")
	}
	crypter.id0 = id0

	return crypter, nil
}

// GetAccessPermissions returns the PDF access permissions as an AccessPermissions object.
func (crypt *PdfCrypt) GetAccessPermissions() AccessPermissions {
	return crypt.encryptStd.P
}

func (crypt *PdfCrypt) securityHandler() stdSecurityHandler {
	if crypt.encryptStd.R >= 5 {
		return stdHandlerR6{}
	}
	return stdHandlerR4{
		ID0:    crypt.id0,
		Length: crypt.encrypt.Length,
	}
}

// Check whether the specified password can be used to decrypt the document.
// Also build the encryption/decryption key.
func (crypt *PdfCrypt) authenticate(password []byte) (bool, error) {
	crypt.authenticated = false
	h := crypt.securityHandler()
	fkey, perm, err := h.Authenticate(&crypt.encryptStd, password)
	if err != nil {
		return false, err
	} else if perm == 0 || len(fkey) == 0 {
		return false, nil
	}
	crypt.authenticated = true
	crypt.encryptionKey = fkey
	return true, nil
}

// Check access rights and permissions for a specified password.  If either user/owner password is specified,
// full rights are granted, otherwise the access rights are specified by the Permissions flag.
//
// The bool flag indicates that the user can access and can view the file.
// The AccessPermissions shows what access the user has for editing etc.
// An error is returned if there was a problem performing the authentication.
func (crypt *PdfCrypt) checkAccessRights(password []byte) (bool, AccessPermissions, error) {
	h := crypt.securityHandler()
	// TODO(dennwc): it computes an encryption key as well; if necessary, define a new interface method to optimize this
	fkey, perm, err := h.Authenticate(&crypt.encryptStd, password)
	if err != nil {
		return false, 0, err
	} else if perm == 0 || len(fkey) == 0 {
		return false, 0, nil
	}
	return true, perm, nil
}

// Generates a key for encrypting a specific object based on the
// object and generation number, as well as the document encryption key.
func (crypt *PdfCrypt) makeKey(filter string, objNum, genNum uint32, ekey []byte) ([]byte, error) {
	f, ok := crypt.cryptFilters[filter]
	if !ok {
		return nil, fmt.Errorf("Unknown crypt filter (%s)", filter)
	}
	return f.MakeKey(objNum, genNum, ekey)
}

// encryptDictKeys list all required field for "Encrypt" dictionary.
// It is used as a fingerprint to detect old copies of this dictionary.
var encryptDictKeys = []PdfObjectName{
	"V", "R", "O", "U", "P",
}

// Check if object has already been processed.
func (crypt *PdfCrypt) isDecrypted(obj PdfObject) bool {
	_, ok := crypt.decryptedObjects[obj]
	if ok {
		common.Log.Trace("Already decrypted")
		return true
	}
	switch obj := obj.(type) {
	case *PdfObjectStream:
		if crypt.encryptStd.R != 5 {
			if name, ok := obj.Get("Type").(*PdfObjectName); ok && *name == "XRef" {
				return true // Cross-reference streams should not be encrypted
			}
		}
	case *PdfIndirectObject:
		if _, ok = crypt.decryptedObjNum[int(obj.ObjectNumber)]; ok {
			return true
		}
		switch obj := obj.PdfObject.(type) {
		case *PdfObjectDictionary:
			// detect old copies of "Encrypt" dictionary
			// TODO: find a better way to do it
			ok := true
			for _, key := range encryptDictKeys {
				if obj.Get(key) == nil {
					ok = false
					break
				}
			}
			if ok {
				return true
			}
		}
	}

	common.Log.Trace("Not decrypted yet")
	return false
}

// Decrypt a buffer with a selected crypt filter.
func (crypt *PdfCrypt) decryptBytes(buf []byte, filter string, okey []byte) ([]byte, error) {
	common.Log.Trace("Decrypt bytes")
	f, ok := crypt.cryptFilters[filter]
	if !ok {
		return nil, fmt.Errorf("Unknown crypt filter (%s)", filter)
	}
	return f.DecryptBytes(buf, okey)
}

// Decrypt an object with specified key. For numbered objects,
// the key argument is not used and a new one is generated based
// on the object and generation number.
// Traverses through all the subobjects (recursive).
//
// Does not look up references..  That should be done prior to calling.
func (crypt *PdfCrypt) Decrypt(obj PdfObject, parentObjNum, parentGenNum int64) error {
	if crypt.isDecrypted(obj) {
		return nil
	}

	switch obj := obj.(type) {
	case *PdfIndirectObject:
		crypt.decryptedObjects[obj] = true

		common.Log.Trace("Decrypting indirect %d %d obj!", obj.ObjectNumber, obj.GenerationNumber)

		objNum := obj.ObjectNumber
		genNum := obj.GenerationNumber

		err := crypt.Decrypt(obj.PdfObject, objNum, genNum)
		if err != nil {
			return err
		}
		return nil
	case *PdfObjectStream:
		// Mark as decrypted first to avoid recursive issues.
		crypt.decryptedObjects[obj] = true
		dict := obj.PdfObjectDictionary

		if crypt.encryptStd.R != 5 {
			if s, ok := dict.Get("Type").(*PdfObjectName); ok && *s == "XRef" {
				return nil // Cross-reference streams should not be encrypted
			}
		}

		objNum := obj.ObjectNumber
		genNum := obj.GenerationNumber
		common.Log.Trace("Decrypting stream %d %d !", objNum, genNum)

		// TODO: Check for crypt filter (V4).
		// The Crypt filter shall be the first filter in the Filter array entry.

		streamFilter := StandardCryptFilter // Default RC4.
		if crypt.encrypt.V >= 4 {
			streamFilter = crypt.streamFilter
			common.Log.Trace("this.streamFilter = %s", crypt.streamFilter)

			if filters, ok := dict.Get("Filter").(*PdfObjectArray); ok {
				// Crypt filter can only be the first entry.
				if firstFilter, ok := GetName(filters.Get(0)); ok {
					if *firstFilter == "Crypt" {
						// Crypt filter overriding the default.
						// Default option is Identity.
						streamFilter = "Identity"

						// Check if valid crypt filter specified in the decode params.
						if decodeParams, ok := dict.Get("DecodeParms").(*PdfObjectDictionary); ok {
							if filterName, ok := decodeParams.Get("Name").(*PdfObjectName); ok {
								if _, ok := crypt.cryptFilters[string(*filterName)]; ok {
									common.Log.Trace("Using stream filter %s", *filterName)
									streamFilter = string(*filterName)
								}
							}
						}
					}
				}
			}

			common.Log.Trace("with %s filter", streamFilter)
			if streamFilter == "Identity" {
				// Identity: pass unchanged.
				return nil
			}
		}

		err := crypt.Decrypt(dict, objNum, genNum)
		if err != nil {
			return err
		}

		okey, err := crypt.makeKey(streamFilter, uint32(objNum), uint32(genNum), crypt.encryptionKey)
		if err != nil {
			return err
		}

		obj.Stream, err = crypt.decryptBytes(obj.Stream, streamFilter, okey)
		if err != nil {
			return err
		}
		// Update the length based on the decrypted stream.
		dict.Set("Length", MakeInteger(int64(len(obj.Stream))))

		return nil
	case *PdfObjectString:
		common.Log.Trace("Decrypting string!")

		stringFilter := StandardCryptFilter
		if crypt.encrypt.V >= 4 {
			// Currently only support Identity / RC4.
			common.Log.Trace("with %s filter", crypt.stringFilter)
			if crypt.stringFilter == "Identity" {
				// Identity: pass unchanged: No action.
				return nil
			}
			stringFilter = crypt.stringFilter
		}

		key, err := crypt.makeKey(stringFilter, uint32(parentObjNum), uint32(parentGenNum), crypt.encryptionKey)
		if err != nil {
			return err
		}

		// Overwrite the encrypted with decrypted string.
		str := obj.Str()
		decrypted := make([]byte, len(str))
		for i := 0; i < len(str); i++ {
			decrypted[i] = str[i]
		}
		common.Log.Trace("Decrypt string: %s : % x", decrypted, decrypted)
		decrypted, err = crypt.decryptBytes(decrypted, stringFilter, key)
		if err != nil {
			return err
		}
		obj.val = string(decrypted)

		return nil
	case *PdfObjectArray:
		for _, o := range obj.Elements() {
			err := crypt.Decrypt(o, parentObjNum, parentGenNum)
			if err != nil {
				return err
			}
		}
		return nil
	case *PdfObjectDictionary:
		isSig := false
		if t := obj.Get("Type"); t != nil {
			typeStr, ok := t.(*PdfObjectName)
			if ok && *typeStr == "Sig" {
				isSig = true
			}
		}
		for _, keyidx := range obj.Keys() {
			o := obj.Get(keyidx)
			// How can we avoid this check, i.e. implement a more smart
			// traversal system?
			if isSig && string(keyidx) == "Contents" {
				// Leave the Contents of a Signature dictionary.
				continue
			}

			if string(keyidx) != "Parent" && string(keyidx) != "Prev" && string(keyidx) != "Last" { // Check not needed?
				err := crypt.Decrypt(o, parentObjNum, parentGenNum)
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
func (crypt *PdfCrypt) isEncrypted(obj PdfObject) bool {
	_, ok := crypt.encryptedObjects[obj]
	if ok {
		common.Log.Trace("Already encrypted")
		return true
	}

	common.Log.Trace("Not encrypted yet")
	return false
}

// Encrypt a buffer with the specified crypt filter and key.
func (crypt *PdfCrypt) encryptBytes(buf []byte, filter string, okey []byte) ([]byte, error) {
	common.Log.Trace("Encrypt bytes")
	f, ok := crypt.cryptFilters[filter]
	if !ok {
		return nil, fmt.Errorf("Unknown crypt filter (%s)", filter)
	}
	return f.EncryptBytes(buf, okey)
}

// Encrypt an object with specified key. For numbered objects,
// the key argument is not used and a new one is generated based
// on the object and generation number.
// Traverses through all the subobjects (recursive).
//
// Does not look up references..  That should be done prior to calling.
func (crypt *PdfCrypt) Encrypt(obj PdfObject, parentObjNum, parentGenNum int64) error {
	if crypt.isEncrypted(obj) {
		return nil
	}
	switch obj := obj.(type) {
	case *PdfIndirectObject:
		crypt.encryptedObjects[obj] = true

		common.Log.Trace("Encrypting indirect %d %d obj!", obj.ObjectNumber, obj.GenerationNumber)

		objNum := obj.ObjectNumber
		genNum := obj.GenerationNumber

		err := crypt.Encrypt(obj.PdfObject, objNum, genNum)
		if err != nil {
			return err
		}
		return nil
	case *PdfObjectStream:
		crypt.encryptedObjects[obj] = true
		dict := obj.PdfObjectDictionary

		if s, ok := dict.Get("Type").(*PdfObjectName); ok && *s == "XRef" {
			return nil // Cross-reference streams should not be encrypted
		}

		objNum := obj.ObjectNumber
		genNum := obj.GenerationNumber
		common.Log.Trace("Encrypting stream %d %d !", objNum, genNum)

		// TODO: Check for crypt filter (V4).
		// The Crypt filter shall be the first filter in the Filter array entry.

		streamFilter := StandardCryptFilter // Default RC4.
		if crypt.encrypt.V >= 4 {
			// For now.  Need to change when we add support for more than
			// Identity / RC4.
			streamFilter = crypt.streamFilter
			common.Log.Trace("this.streamFilter = %s", crypt.streamFilter)

			if filters, ok := dict.Get("Filter").(*PdfObjectArray); ok {
				// Crypt filter can only be the first entry.
				if firstFilter, ok := GetName(filters.Get(0)); ok {
					if *firstFilter == "Crypt" {
						// Crypt filter overriding the default.
						// Default option is Identity.
						streamFilter = "Identity"

						// Check if valid crypt filter specified in the decode params.
						if decodeParams, ok := dict.Get("DecodeParms").(*PdfObjectDictionary); ok {
							if filterName, ok := decodeParams.Get("Name").(*PdfObjectName); ok {
								if _, ok := crypt.cryptFilters[string(*filterName)]; ok {
									common.Log.Trace("Using stream filter %s", *filterName)
									streamFilter = string(*filterName)
								}
							}
						}
					}
				}
			}

			common.Log.Trace("with %s filter", streamFilter)
			if streamFilter == "Identity" {
				// Identity: pass unchanged.
				return nil
			}
		}

		err := crypt.Encrypt(obj.PdfObjectDictionary, objNum, genNum)
		if err != nil {
			return err
		}

		okey, err := crypt.makeKey(streamFilter, uint32(objNum), uint32(genNum), crypt.encryptionKey)
		if err != nil {
			return err
		}

		obj.Stream, err = crypt.encryptBytes(obj.Stream, streamFilter, okey)
		if err != nil {
			return err
		}
		// Update the length based on the encrypted stream.
		dict.Set("Length", MakeInteger(int64(len(obj.Stream))))

		return nil
	case *PdfObjectString:
		common.Log.Trace("Encrypting string!")

		stringFilter := StandardCryptFilter
		if crypt.encrypt.V >= 4 {
			common.Log.Trace("with %s filter", crypt.stringFilter)
			if crypt.stringFilter == "Identity" {
				// Identity: pass unchanged: No action.
				return nil
			}
			stringFilter = crypt.stringFilter
		}

		key, err := crypt.makeKey(stringFilter, uint32(parentObjNum), uint32(parentGenNum), crypt.encryptionKey)
		if err != nil {
			return err
		}

		str := obj.Str()
		encrypted := make([]byte, len(str))
		for i := 0; i < len(str); i++ {
			encrypted[i] = str[i]
		}
		common.Log.Trace("Encrypt string: %s : % x", encrypted, encrypted)
		encrypted, err = crypt.encryptBytes(encrypted, stringFilter, key)
		if err != nil {
			return err
		}
		obj.val = string(encrypted)

		return nil
	case *PdfObjectArray:
		for _, o := range obj.Elements() {
			err := crypt.Encrypt(o, parentObjNum, parentGenNum)
			if err != nil {
				return err
			}
		}
		return nil
	case *PdfObjectDictionary:
		isSig := false
		if t := obj.Get("Type"); t != nil {
			typeStr, ok := t.(*PdfObjectName)
			if ok && *typeStr == "Sig" {
				isSig = true
			}
		}

		for _, keyidx := range obj.Keys() {
			o := obj.Get(keyidx)
			// How can we avoid this check, i.e. implement a more smart
			// traversal system?
			if isSig && string(keyidx) == "Contents" {
				// Leave the Contents of a Signature dictionary.
				continue
			}
			if string(keyidx) != "Parent" && string(keyidx) != "Prev" && string(keyidx) != "Last" { // Check not needed?
				err := crypt.Encrypt(o, parentObjNum, parentGenNum)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	return nil
}

// generateParams generates encryption parameters for specified passwords.
func (crypt *PdfCrypt) generateParams(upass, opass []byte) error {
	h := crypt.securityHandler()
	ekey, err := h.GenerateParams(&crypt.encryptStd, opass, upass)
	if err != nil {
		return err
	}
	crypt.encryptionKey = ekey
	return nil
}
