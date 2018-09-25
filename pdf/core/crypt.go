/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"

	"github.com/unidoc/unidoc/common"
)

// PdfCrypt provides PDF encryption/decryption support.
// The PDF standard supports encryption of strings and streams (Section 7.6).
// TODO (v3): Consider unexporting.
type PdfCrypt struct {
	Filter           string
	Subfilter        string
	V                int
	Length           int
	R                int
	O                []byte
	U                []byte
	OE               []byte // R=6
	UE               []byte // R=6
	P                int    // TODO (v3): uint32
	Perms            []byte // R=6
	EncryptMetadata  bool
	Id0              string
	EncryptionKey    []byte
	DecryptedObjects map[PdfObject]bool
	EncryptedObjects map[PdfObject]bool
	Authenticated    bool
	// Crypt filters (V4).
	CryptFilters CryptFilters
	StreamFilter string
	StringFilter string

	parser *PdfParser

	ivAESZero []byte // a zero buffer used as an initialization vector for AES
}

// AccessPermissions is a list of access permissions for a PDF file.
type AccessPermissions struct {
	Printing        bool
	Modify          bool
	ExtractGraphics bool
	Annotate        bool

	// Allow form filling, if annotation is disabled?  If annotation enabled, is not looked at.
	FillForms         bool
	DisabilityExtract bool // not clear what this means!

	// Allow rotating, editing page order.
	RotateInsert bool

	// Limit print quality (lowres), assuming Printing is true.
	FullPrintQuality bool
}

const padding = "\x28\xBF\x4E\x5E\x4E\x75\x8A\x41\x64\x00\x4E\x56\xFF" +
	"\xFA\x01\x08\x2E\x2E\x00\xB6\xD0\x68\x3E\x80\x2F\x0C" +
	"\xA9\xFE\x64\x53\x69\x7A"

// StandardCryptFilter is a default name for a standard crypt filter.
const StandardCryptFilter = "StdCF"

// CryptFilter represents information from a CryptFilter dictionary.
// TODO (v3): Replace with cryptFilterMethod interface.
type CryptFilter struct {
	Cfm    string
	Length int
	cfm    cryptFilterMethod
}

func (cf CryptFilter) getCFM() (cryptFilterMethod, error) {
	// TODO (v3): remove this method and access cf.cfm directly
	if cf.cfm != nil {
		return cf.cfm, nil
	}
	// There is a non-zero chance that someone relies on the ability to
	// add crypt filters manually using the library.
	// So if we hit such case - be nice and find a filter by name.
	return getCryptFilterMethod(cf.Cfm)
}

// Encryption filters names.
// Table 25, CFM (page 92)
const (
	CryptFilterNone  = "None"  // do not decrypt data
	CryptFilterV2    = "V2"    // RC4-based filter
	CryptFilterAESV2 = "AESV2" // AES-based filter (128 bit key, PDF 1.6)
	CryptFilterAESV3 = "AESV3" // AES-based filter (256 bit key, PDF 2.0)
)

func newCryptFiltersV2(length int) CryptFilters {
	return CryptFilters{
		StandardCryptFilter: NewCryptFilterV2(length),
	}
}

// NewCryptFilterV2 creates a RC4-based filter with a specified key length (in bytes).
func NewCryptFilterV2(length int) CryptFilter {
	// TODO (v3): Unexport.
	return CryptFilter{
		Cfm:    CryptFilterV2,
		Length: length,
		cfm:    cryptFilterV2{},
	}
}

// NewCryptFilterAESV2 creates an AES-based filter with a 128 bit key (AESV2).
func NewCryptFilterAESV2() CryptFilter {
	// TODO (v3): Unexport.
	return CryptFilter{
		Cfm:    CryptFilterAESV2,
		Length: 16,
		cfm:    cryptFilterAESV2{},
	}
}

// NewCryptFilterAESV3 creates an AES-based filter with a 256 bit key (AESV3).
func NewCryptFilterAESV3() CryptFilter {
	// TODO (v3): Unexport.
	return CryptFilter{
		Cfm:    CryptFilterAESV3,
		Length: 32,
		cfm:    cryptFilterAESV3{},
	}
}

// CryptFilters is a map of crypt filter name and underlying CryptFilter info.
// TODO (v3): Unexport.
type CryptFilters map[string]CryptFilter

func (m CryptFilters) byName(cfm string) (cryptFilterMethod, error) {
	cf, ok := m[cfm]
	if !ok {
		err := fmt.Errorf("Unsupported crypt filter (%s)", cfm)
		common.Log.Debug("%s", err)
		return nil, err
	}
	f, err := cf.getCFM()
	if err != nil {
		common.Log.Debug("%s", err)
		return nil, err
	}
	return f, nil
}

// LoadCryptFilters loads crypt filter information from the encryption dictionary (V>=4).
// TODO (v3): Unexport.
func (crypt *PdfCrypt) LoadCryptFilters(ed *PdfObjectDictionary) error {
	crypt.CryptFilters = CryptFilters{}

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

		// If Type present, should be CryptFilter.
		if typename, ok := dict.Get("Type").(*PdfObjectName); ok {
			if string(*typename) != "CryptFilter" {
				return fmt.Errorf("CF dict type != CryptFilter (%s)", typename)
			}
		}

		cf := CryptFilter{}

		// Method.
		cfmName, ok := dict.Get("CFM").(*PdfObjectName)
		if !ok {
			return fmt.Errorf("Unsupported crypt filter (None)")
		}
		cf.Cfm = string(*cfmName)

		cfm, err := getCryptFilterMethod(cf.Cfm)
		if err != nil {
			return err
		}
		cf.cfm = cfm

		// Length.
		cf.Length = 0
		length, ok := dict.Get("Length").(*PdfObjectInteger)
		if ok {
			// TODO(dennwc): pass length to getCryptFilterMethod and allow filter to validate it
			if *length%8 != 0 {
				return fmt.Errorf("Crypt filter length not multiple of 8 (%d)", *length)
			}

			// Standard security handler expresses the length in multiples of 8 (16 means 128)
			// We only deal with standard so far. (Public key not supported yet).
			if *length < 5 || *length > 16 {
				if *length == 64 || *length == 128 {
					common.Log.Debug("STANDARD VIOLATION: Crypt Length appears to be in bits rather than bytes - assuming bits (%d)", *length)
					*length /= 8
				} else if !(*length == 32 && cf.Cfm == CryptFilterAESV3) {
					return fmt.Errorf("Crypt filter length not in range 40 - 128 bit (%d)", *length)
				}
			}
			cf.Length = int(*length)
		}

		crypt.CryptFilters[string(name)] = cf
	}
	// Cannot be overwritten.
	crypt.CryptFilters["Identity"] = CryptFilter{}

	// StrF strings filter.
	crypt.StringFilter = "Identity"
	if strf, ok := ed.Get("StrF").(*PdfObjectName); ok {
		if _, exists := crypt.CryptFilters[string(*strf)]; !exists {
			return fmt.Errorf("Crypt filter for StrF not specified in CF dictionary (%s)", *strf)
		}
		crypt.StringFilter = string(*strf)
	}

	// StmF streams filter.
	crypt.StreamFilter = "Identity"
	if stmf, ok := ed.Get("StmF").(*PdfObjectName); ok {
		if _, exists := crypt.CryptFilters[string(*stmf)]; !exists {
			return fmt.Errorf("Crypt filter for StmF not specified in CF dictionary (%s)", *stmf)
		}
		crypt.StreamFilter = string(*stmf)
	}

	return nil
}

// SaveCryptFilters saves crypt filter information to the encryption dictionary (V>=4).
// TODO (v3): Unexport.
func (crypt *PdfCrypt) SaveCryptFilters(ed *PdfObjectDictionary) error {
	if crypt.V < 4 {
		return errors.New("can only be used with V>=4")
	}
	cf := MakeDict()
	ed.Set("CF", cf)

	for name, filter := range crypt.CryptFilters {
		if name == "Identity" {
			continue
		}
		v := MakeDict()
		cf.Set(PdfObjectName(name), v)

		v.Set("Type", MakeName("CryptFilter"))
		v.Set("AuthEvent", MakeName("DocOpen"))
		v.Set("CFM", MakeName(string(filter.Cfm)))
		v.Set("Length", MakeInteger(int64(filter.Length)))
	}
	ed.Set("StrF", MakeName(crypt.StringFilter))
	ed.Set("StmF", MakeName(crypt.StreamFilter))
	return nil
}

// PdfCryptMakeNew makes the document crypt handler based on the encryption dictionary
// and trailer dictionary. Returns an error on failure to process.
func PdfCryptMakeNew(parser *PdfParser, ed, trailer *PdfObjectDictionary) (PdfCrypt, error) {
	crypter := PdfCrypt{}
	crypter.DecryptedObjects = map[PdfObject]bool{}
	crypter.EncryptedObjects = map[PdfObject]bool{}
	crypter.Authenticated = false
	crypter.parser = parser

	filter, ok := ed.Get("Filter").(*PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR Crypt dictionary missing required Filter field!")
		return crypter, errors.New("Required crypt field Filter missing")
	}
	if *filter != "Standard" {
		common.Log.Debug("ERROR Unsupported filter (%s)", *filter)
		return crypter, errors.New("Unsupported Filter")
	}
	crypter.Filter = string(*filter)

	subfilter, ok := ed.Get("SubFilter").(*PdfObjectString)
	if ok {
		crypter.Subfilter = string(*subfilter)
		common.Log.Debug("Using subfilter %s", subfilter)
	}

	if L, ok := ed.Get("Length").(*PdfObjectInteger); ok {
		if (*L % 8) != 0 {
			common.Log.Debug("ERROR Invalid encryption length")
			return crypter, errors.New("Invalid encryption length")
		}
		crypter.Length = int(*L)
	} else {
		crypter.Length = 40
	}

	crypter.V = 0
	if v, ok := ed.Get("V").(*PdfObjectInteger); ok {
		V := int(*v)
		crypter.V = V
		if V >= 1 && V <= 2 {
			// Default algorithm is V2.
			crypter.CryptFilters = newCryptFiltersV2(crypter.Length)
		} else if V >= 4 && V <= 5 {
			if err := crypter.LoadCryptFilters(ed); err != nil {
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
	crypter.R = int(*R)

	O, ok := ed.Get("O").(*PdfObjectString)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing O")
	}
	if crypter.R == 5 || crypter.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(*O) < 48 {
			return crypter, fmt.Errorf("Length(O) < 48 (%d)", len(*O))
		}
	} else if len(*O) != 32 {
		return crypter, fmt.Errorf("Length(O) != 32 (%d)", len(*O))
	}
	crypter.O = []byte(*O)

	U, ok := ed.Get("U").(*PdfObjectString)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing U")
	}
	if crypter.R == 5 || crypter.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(*U) < 48 {
			return crypter, fmt.Errorf("Length(U) < 48 (%d)", len(*U))
		}
	} else if len(*U) != 32 {
		// Strictly this does not cause an error.
		// If O is OK and others then can still read the file.
		common.Log.Debug("Warning: Length(U) != 32 (%d)", len(*U))
		//return crypter, errors.New("Length(U) != 32")
	}
	crypter.U = []byte(*U)

	if crypter.R >= 5 {
		OE, ok := ed.Get("OE").(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Encrypt dictionary missing OE")
		}
		if len(*OE) != 32 {
			return crypter, fmt.Errorf("Length(OE) != 32 (%d)", len(*OE))
		}
		crypter.OE = []byte(*OE)

		UE, ok := ed.Get("UE").(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Encrypt dictionary missing UE")
		}
		if len(*UE) != 32 {
			return crypter, fmt.Errorf("Length(UE) != 32 (%d)", len(*UE))
		}
		crypter.UE = []byte(*UE)
	}

	P, ok := ed.Get("P").(*PdfObjectInteger)
	if !ok {
		return crypter, errors.New("Encrypt dictionary missing permissions attr")
	}
	crypter.P = int(*P)

	if crypter.R == 6 {
		Perms, ok := ed.Get("Perms").(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Encrypt dictionary missing Perms")
		}
		if len(*Perms) != 16 {
			return crypter, fmt.Errorf("Length(Perms) != 16 (%d)", len(*Perms))
		}
		crypter.Perms = []byte(*Perms)
	}

	em, ok := ed.Get("EncryptMetadata").(*PdfObjectBool)
	if ok {
		crypter.EncryptMetadata = bool(*em)
	} else {
		crypter.EncryptMetadata = true // True by default.
	}

	// Default: empty ID.
	// Strictly, if file is encrypted, the ID should always be specified
	// but clearly not everyone is following the specification.
	id0 := PdfObjectString("")
	if idArray, ok := trailer.Get("ID").(*PdfObjectArray); ok && len(*idArray) >= 1 {
		id0obj, ok := (*idArray)[0].(*PdfObjectString)
		if !ok {
			return crypter, errors.New("Invalid trailer ID")
		}
		id0 = *id0obj
	} else {
		common.Log.Debug("Trailer ID array missing or invalid!")
	}
	crypter.Id0 = string(id0)

	return crypter, nil
}

// GetAccessPermissions returns the PDF access permissions as an AccessPermissions object.
func (crypt *PdfCrypt) GetAccessPermissions() AccessPermissions {
	perms := AccessPermissions{}

	P := crypt.P
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
		perms.FullPrintQuality = true
	}
	return perms
}

// GetP returns the P entry to be used in Encrypt dictionary based on AccessPermissions settings.
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
		P |= (1 << 9) // bit 10
	}
	if perms.RotateInsert {
		P |= (1 << 10) // bit 11
	}
	if perms.FullPrintQuality {
		P |= (1 << 11) // bit 12
	}
	return P
}

// Check whether the specified password can be used to decrypt the document.
func (crypt *PdfCrypt) authenticate(password []byte) (bool, error) {
	// Also build the encryption/decryption key.

	crypt.Authenticated = false
	if crypt.R >= 5 {
		authenticated, err := crypt.alg2a(password)
		if err != nil {
			return false, err
		}
		crypt.Authenticated = authenticated
		return authenticated, err
	}

	// Try user password.
	common.Log.Trace("Debugging authentication - user pass")
	authenticated, err := crypt.Alg6(password)
	if err != nil {
		return false, err
	}
	if authenticated {
		common.Log.Trace("this.Authenticated = True")
		crypt.Authenticated = true
		return true, nil
	}

	// Try owner password also.
	// May not be necessary if only want to get all contents.
	// (user pass needs to be known or empty).
	common.Log.Trace("Debugging authentication - owner pass")
	authenticated, err = crypt.Alg7(password)
	if err != nil {
		return false, err
	}
	if authenticated {
		common.Log.Trace("this.Authenticated = True")
		crypt.Authenticated = true
		return true, nil
	}

	return false, nil
}

// Check access rights and permissions for a specified password.  If either user/owner password is specified,
// full rights are granted, otherwise the access rights are specified by the Permissions flag.
//
// The bool flag indicates that the user can access and can view the file.
// The AccessPermissions shows what access the user has for editing etc.
// An error is returned if there was a problem performing the authentication.
func (crypt *PdfCrypt) checkAccessRights(password []byte) (bool, AccessPermissions, error) {
	perms := AccessPermissions{}

	// Try owner password -> full rights.
	var (
		isOwner bool
		err     error
	)
	if crypt.R >= 5 {
		var h []byte
		h, err = crypt.alg12(password)
		if err != nil {
			return false, perms, err
		}
		isOwner = len(h) != 0
	} else {
		isOwner, err = crypt.Alg7(password)
	}
	if err != nil {
		return false, perms, err
	}
	if isOwner {
		// owner -> full rights.
		perms.Annotate = true
		perms.DisabilityExtract = true
		perms.ExtractGraphics = true
		perms.FillForms = true
		perms.FullPrintQuality = true
		perms.Modify = true
		perms.Printing = true
		perms.RotateInsert = true
		return true, perms, nil
	}

	// Try user password.
	var isUser bool
	if crypt.R >= 5 {
		var h []byte
		h, err = crypt.alg11(password)
		if err != nil {
			return false, perms, err
		}
		isUser = len(h) != 0
	} else {
		isUser, err = crypt.Alg6(password)
	}
	if err != nil {
		return false, perms, err
	}
	if isUser {
		// User password specified correctly -> access granted with specified permissions.
		return true, crypt.GetAccessPermissions(), nil
	}

	// Cannot even view the file.
	return false, perms, nil
}

func (crypt *PdfCrypt) paddedPass(pass []byte) []byte {
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
func (crypt *PdfCrypt) makeKey(filter string, objNum, genNum uint32, ekey []byte) ([]byte, error) {
	f, err := crypt.CryptFilters.byName(filter)
	if err != nil {
		return nil, err
	}
	return f.MakeKey(objNum, genNum, ekey)
}

// Check if object has already been processed.
func (crypt *PdfCrypt) isDecrypted(obj PdfObject) bool {
	_, ok := crypt.DecryptedObjects[obj]
	if ok {
		common.Log.Trace("Already decrypted")
		return true
	}

	common.Log.Trace("Not decrypted yet")
	return false
}

// Decrypt a buffer with a selected crypt filter.
func (crypt *PdfCrypt) decryptBytes(buf []byte, filter string, okey []byte) ([]byte, error) {
	common.Log.Trace("Decrypt bytes")
	f, err := crypt.CryptFilters.byName(filter)
	if err != nil {
		return nil, err
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
		crypt.DecryptedObjects[obj] = true

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
		crypt.DecryptedObjects[obj] = true
		dict := obj.PdfObjectDictionary

		if s, ok := dict.Get("Type").(*PdfObjectName); ok && *s == "XRef" {
			return nil // Cross-reference streams should not be encrypted
		}

		objNum := obj.ObjectNumber
		genNum := obj.GenerationNumber
		common.Log.Trace("Decrypting stream %d %d !", objNum, genNum)

		// TODO: Check for crypt filter (V4).
		// The Crypt filter shall be the first filter in the Filter array entry.

		streamFilter := StandardCryptFilter // Default RC4.
		if crypt.V >= 4 {
			streamFilter = crypt.StreamFilter
			common.Log.Trace("this.StreamFilter = %s", crypt.StreamFilter)

			if filters, ok := dict.Get("Filter").(*PdfObjectArray); ok {
				// Crypt filter can only be the first entry.
				if firstFilter, ok := (*filters)[0].(*PdfObjectName); ok {
					if *firstFilter == "Crypt" {
						// Crypt filter overriding the default.
						// Default option is Identity.
						streamFilter = "Identity"

						// Check if valid crypt filter specified in the decode params.
						if decodeParams, ok := dict.Get("DecodeParms").(*PdfObjectDictionary); ok {
							if filterName, ok := decodeParams.Get("Name").(*PdfObjectName); ok {
								if _, ok := crypt.CryptFilters[string(*filterName)]; ok {
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

		okey, err := crypt.makeKey(streamFilter, uint32(objNum), uint32(genNum), crypt.EncryptionKey)
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
		if crypt.V >= 4 {
			// Currently only support Identity / RC4.
			common.Log.Trace("with %s filter", crypt.StringFilter)
			if crypt.StringFilter == "Identity" {
				// Identity: pass unchanged: No action.
				return nil
			}
			stringFilter = crypt.StringFilter
		}

		key, err := crypt.makeKey(stringFilter, uint32(parentObjNum), uint32(parentGenNum), crypt.EncryptionKey)
		if err != nil {
			return err
		}

		// Overwrite the encrypted with decrypted string.
		decrypted := make([]byte, len(*obj))
		for i := 0; i < len(*obj); i++ {
			decrypted[i] = (*obj)[i]
		}
		common.Log.Trace("Decrypt string: %s : % x", decrypted, decrypted)
		decrypted, err = crypt.decryptBytes(decrypted, stringFilter, key)
		if err != nil {
			return err
		}
		*obj = PdfObjectString(decrypted)

		return nil
	case *PdfObjectArray:
		for _, o := range *obj {
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
	_, ok := crypt.EncryptedObjects[obj]
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
	f, err := crypt.CryptFilters.byName(filter)
	if err != nil {
		return nil, err
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
		crypt.EncryptedObjects[obj] = true

		common.Log.Trace("Encrypting indirect %d %d obj!", obj.ObjectNumber, obj.GenerationNumber)

		objNum := obj.ObjectNumber
		genNum := obj.GenerationNumber

		err := crypt.Encrypt(obj.PdfObject, objNum, genNum)
		if err != nil {
			return err
		}
		return nil
	case *PdfObjectStream:
		crypt.EncryptedObjects[obj] = true
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
		if crypt.V >= 4 {
			// For now.  Need to change when we add support for more than
			// Identity / RC4.
			streamFilter = crypt.StreamFilter
			common.Log.Trace("this.StreamFilter = %s", crypt.StreamFilter)

			if filters, ok := dict.Get("Filter").(*PdfObjectArray); ok {
				// Crypt filter can only be the first entry.
				if firstFilter, ok := (*filters)[0].(*PdfObjectName); ok {
					if *firstFilter == "Crypt" {
						// Crypt filter overriding the default.
						// Default option is Identity.
						streamFilter = "Identity"

						// Check if valid crypt filter specified in the decode params.
						if decodeParams, ok := dict.Get("DecodeParms").(*PdfObjectDictionary); ok {
							if filterName, ok := decodeParams.Get("Name").(*PdfObjectName); ok {
								if _, ok := crypt.CryptFilters[string(*filterName)]; ok {
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

		okey, err := crypt.makeKey(streamFilter, uint32(objNum), uint32(genNum), crypt.EncryptionKey)
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
		if crypt.V >= 4 {
			common.Log.Trace("with %s filter", crypt.StringFilter)
			if crypt.StringFilter == "Identity" {
				// Identity: pass unchanged: No action.
				return nil
			}
			stringFilter = crypt.StringFilter
		}

		key, err := crypt.makeKey(stringFilter, uint32(parentObjNum), uint32(parentGenNum), crypt.EncryptionKey)
		if err != nil {
			return err
		}

		encrypted := make([]byte, len(*obj))
		for i := 0; i < len(*obj); i++ {
			encrypted[i] = (*obj)[i]
		}
		common.Log.Trace("Encrypt string: %s : % x", encrypted, encrypted)
		encrypted, err = crypt.encryptBytes(encrypted, stringFilter, key)
		if err != nil {
			return err
		}
		*obj = PdfObjectString(encrypted)

		return nil
	case *PdfObjectArray:
		for _, o := range *obj {
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

// aesZeroIV allocates a zero-filled buffer that serves as an initialization vector for AESv3.
func (crypt *PdfCrypt) aesZeroIV() []byte {
	if crypt.ivAESZero == nil {
		crypt.ivAESZero = make([]byte, aes.BlockSize)
	}
	return crypt.ivAESZero
}

// alg2a retrieves the encryption key from an encrypted document (R >= 5).
// It returns false if the password was wrong.
// 7.6.4.3.2 Algorithm 2.A (page 83)
func (crypt *PdfCrypt) alg2a(pass []byte) (bool, error) {
	// O & U: 32 byte hash + 8 byte Validation Salt + 8 byte Key Salt

	// step a: Unicode normalization
	// TODO(dennwc): make sure that UTF-8 strings are normalized

	// step b: truncate to 127 bytes
	if len(pass) > 127 {
		pass = pass[:127]
	}

	// step c: test pass against the owner key
	h, err := crypt.alg12(pass)
	if err != nil {
		return false, err
	}
	var (
		data []byte // data to hash
		ekey []byte // encrypted file key
		ukey []byte // user key; set only when using owner's password
	)
	if len(h) != 0 {
		// owner password valid

		// step d: compute an intermediate owner key
		str := make([]byte, len(pass)+8+48)
		i := copy(str, pass)
		i += copy(str[i:], crypt.O[40:48]) // owner Key Salt
		i += copy(str[i:], crypt.U[0:48])

		data = str
		ekey = crypt.OE
		ukey = crypt.U[0:48]
	} else {
		// check user password
		h, err = crypt.alg11(pass)
		if err == nil && len(h) == 0 {
			// try default password
			h, err = crypt.alg11([]byte(""))
		}
		if err != nil {
			return false, err
		} else if len(h) == 0 {
			// wrong password
			return false, nil
		}
		// step e: compute an intermediate user key
		str := make([]byte, len(pass)+8)
		i := copy(str, pass)
		i += copy(str[i:], crypt.U[40:48]) // user Key Salt

		data = str
		ekey = crypt.UE
		ukey = nil
	}
	ekey = ekey[:32]

	// intermediate key
	ikey := crypt.alg2b(data, pass, ukey)

	ac, err := aes.NewCipher(ikey[:32])
	if err != nil {
		panic(err)
	}

	iv := crypt.aesZeroIV()
	cbc := cipher.NewCBCDecrypter(ac, iv)
	fkey := make([]byte, 32)
	cbc.CryptBlocks(fkey, ekey)

	crypt.EncryptionKey = fkey

	if crypt.R == 5 {
		return true, nil
	}

	return crypt.alg13(fkey)
}

// alg2b computes a hash for R=5 and R=6.
func (crypt *PdfCrypt) alg2b(data, pwd, userKey []byte) []byte {
	if crypt.R == 5 {
		return alg2b_R5(data)
	}
	return alg2b(data, pwd, userKey)
}

// alg2b_R5 computes a hash for R=5, used in a deprecated extension.
// It's used the same way as a hash described in Algorithm 2.B, but it doesn't use the original password
// and the user key to calculate the hash.
func alg2b_R5(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
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
func alg2b(data, pwd, userKey []byte) []byte {
	var (
		s256, s384, s512 hash.Hash
	)
	s256 = sha256.New()
	hbuf := make([]byte, 64)

	h := s256
	h.Write(data)
	K := h.Sum(hbuf[:0])

	buf := make([]byte, 64*(127+64+48))

	round := func(rnd int) (E []byte) {
		// step a: repeat pass+K 64 times
		n := len(pwd) + len(K) + len(userKey)
		part := buf[:n]
		i := copy(part, pwd)
		i += copy(part[i:], K[:])
		i += copy(part[i:], userKey)
		if i != n {
			panic("wrong size")
		}
		K1 := buf[:n*64]
		repeat(K1, n)

		// step b: encrypt K1 with AES-128 CBC
		ac, err := aes.NewCipher(K[0:16])
		if err != nil {
			panic(err)
		}
		cbc := cipher.NewCBCEncrypter(ac, K[16:32])
		cbc.CryptBlocks(K1, K1)
		E = K1

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

		return E
	}

	for i := 0; ; {
		E := round(i)
		b := uint8(E[len(E)-1])
		// from the spec, it appears that i should be incremented after
		// the test, but that doesn't match what Adobe does
		i++
		if i >= 64 && b <= uint8(i-32) {
			break
		}
	}
	return K[:32]
}

// Alg2 computes an encryption key.
// TODO (v3): Unexport.
func (crypt *PdfCrypt) Alg2(pass []byte) []byte {
	common.Log.Trace("Alg2")
	key := crypt.paddedPass(pass)

	h := md5.New()
	h.Write(key)

	// Pass O.
	h.Write(crypt.O)

	// Pass P (Lower order byte first).
	var p uint32 = uint32(crypt.P)
	var pb = []byte{}
	for i := 0; i < 4; i++ {
		pb = append(pb, byte(((p >> uint(8*i)) & 0xff)))
	}
	h.Write(pb)
	common.Log.Trace("go P: % x", pb)

	// Pass ID[0] from the trailer
	h.Write([]byte(crypt.Id0))

	common.Log.Trace("this.R = %d encryptMetadata %v", crypt.R, crypt.EncryptMetadata)
	if (crypt.R >= 4) && !crypt.EncryptMetadata {
		h.Write([]byte{0xff, 0xff, 0xff, 0xff})
	}
	hashb := h.Sum(nil)

	if crypt.R >= 3 {
		for i := 0; i < 50; i++ {
			h = md5.New()
			h.Write(hashb[0 : crypt.Length/8])
			hashb = h.Sum(nil)
		}
	}

	if crypt.R >= 3 {
		return hashb[0 : crypt.Length/8]
	}

	return hashb[0:5]
}

// Create the RC4 encryption key.
func (crypt *PdfCrypt) alg3Key(pass []byte) []byte {
	h := md5.New()
	okey := crypt.paddedPass(pass)
	h.Write(okey)

	if crypt.R >= 3 {
		for i := 0; i < 50; i++ {
			hashb := h.Sum(nil)
			h = md5.New()
			h.Write(hashb)
		}
	}

	encKey := h.Sum(nil)
	if crypt.R == 2 {
		encKey = encKey[0:5]
	} else {
		encKey = encKey[0 : crypt.Length/8]
	}
	return encKey
}

// Alg3 computes the encryption dictionary’s O (owner password) value.
// TODO (v3): Unexport.
func (crypt *PdfCrypt) Alg3(upass, opass []byte) (PdfObjectString, error) {
	// Return O string val.
	O := PdfObjectString("")

	var encKey []byte
	if len(opass) > 0 {
		encKey = crypt.alg3Key(opass)
	} else {
		encKey = crypt.alg3Key(upass)
	}

	ociph, err := rc4.NewCipher(encKey)
	if err != nil {
		return O, errors.New("Failed rc4 ciph")
	}

	ukey := crypt.paddedPass(upass)
	encrypted := make([]byte, len(ukey))
	ociph.XORKeyStream(encrypted, ukey)

	if crypt.R >= 3 {
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

// Alg4 computes the encryption dictionary’s U (user password) value (Security handlers of revision 2).
// TODO (v3): Unexport.
func (crypt *PdfCrypt) Alg4(upass []byte) (PdfObjectString, []byte, error) {
	U := PdfObjectString("")

	ekey := crypt.Alg2(upass)
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

// Alg5 computes the encryption dictionary’s U (user password) value (Security handlers of revision 3 or greater).
// TODO (v3): Unexport.
func (crypt *PdfCrypt) Alg5(upass []byte) (PdfObjectString, []byte, error) {
	U := PdfObjectString("")

	ekey := crypt.Alg2(upass)

	h := md5.New()
	h.Write([]byte(padding))
	h.Write([]byte(crypt.Id0))
	hash := h.Sum(nil)

	common.Log.Trace("Alg5")
	common.Log.Trace("ekey: % x", ekey)
	common.Log.Trace("ID: % x", crypt.Id0)

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
		common.Log.Trace("i = %d, ekey: % x", i, ekey2)
		common.Log.Trace("i = %d -> % x", i, encrypted)
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

// Alg6 authenticates the user password.
// TODO (v3): Unexport.
func (crypt *PdfCrypt) Alg6(upass []byte) (bool, error) {
	var uo PdfObjectString
	var err error
	var key []byte
	if crypt.R == 2 {
		uo, key, err = crypt.Alg4(upass)
	} else if crypt.R >= 3 {
		uo, key, err = crypt.Alg5(upass)
	} else {
		return false, errors.New("invalid R")
	}

	if err != nil {
		return false, err
	}

	common.Log.Trace("check: % x == % x ?", string(uo), string(crypt.U))

	uGen := string(uo)      // Generated U from specified pass.
	uDoc := string(crypt.U) // U from the document.
	if crypt.R >= 3 {
		// comparing on the first 16 bytes in the case of security
		// handlers of revision 3 or greater),
		if len(uGen) > 16 {
			uGen = uGen[0:16]
		}
		if len(uDoc) > 16 {
			uDoc = uDoc[0:16]
		}
	}

	if uGen == uDoc {
		crypt.EncryptionKey = key
		return true, nil
	}

	return false, nil
}

// Alg7 authenticates the owner password.
// TODO (v3): Unexport.
func (crypt *PdfCrypt) Alg7(opass []byte) (bool, error) {
	encKey := crypt.alg3Key(opass)

	decrypted := make([]byte, len(crypt.O))
	if crypt.R == 2 {
		ciph, err := rc4.NewCipher(encKey)
		if err != nil {
			return false, errors.New("Failed cipher")
		}
		ciph.XORKeyStream(decrypted, crypt.O)
	} else if crypt.R >= 3 {
		s := append([]byte{}, crypt.O...)
		for i := 0; i < 20; i++ {
			//newKey := encKey
			newKey := append([]byte{}, encKey...)
			for j := 0; j < len(encKey); j++ {
				newKey[j] ^= byte(19 - i)
			}
			ciph, err := rc4.NewCipher(newKey)
			if err != nil {
				return false, errors.New("Failed cipher")
			}
			ciph.XORKeyStream(decrypted, s)
			s = append([]byte{}, decrypted...)
		}
	} else {
		return false, errors.New("invalid R")
	}

	auth, err := crypt.Alg6(decrypted)
	if err != nil {
		return false, nil
	}

	return auth, nil
}

// GenerateParams generates encryption parameters for specified passwords.
// Can be called only for R>=5.
func (crypt *PdfCrypt) GenerateParams(upass, opass []byte) error {
	if crypt.R < 5 {
		// TODO(dennwc): move code for R<5 from PdfWriter.Encrypt
		return errors.New("can be used only for R>=5")
	}
	crypt.EncryptionKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, crypt.EncryptionKey); err != nil {
		return err
	}
	return crypt.generateR6(upass, opass)
}

// generateR6 is the algorithm opposite to alg2a (R>=5).
// It generates U,O,UE,OE,Perms fields using AESv3 encryption.
// There is no algorithm number assigned to this function in the spec.
func (crypt *PdfCrypt) generateR6(upass, opass []byte) error {
	// all these field will be populated by functions below
	crypt.U = nil
	crypt.O = nil
	crypt.UE = nil
	crypt.OE = nil
	crypt.Perms = nil // populated only for R=6

	if len(upass) > 127 {
		upass = upass[:127]
	}
	if len(opass) > 127 {
		opass = opass[:127]
	}
	// generate U and UE
	if err := crypt.alg8(upass); err != nil {
		return err
	}
	// generate O and OE
	if err := crypt.alg9(opass); err != nil {
		return err
	}
	if crypt.R == 5 {
		return nil
	}
	// generate Perms
	return crypt.alg10()
}

// alg8 computes the encryption dictionary's U (user password) and UE (user encryption) values (R>=5).
// 7.6.4.4.6 Algorithm 8 (page 86)
func (crypt *PdfCrypt) alg8(upass []byte) error {
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

	h := crypt.alg2b(str, upass, nil)

	U := make([]byte, len(h)+len(valSalt)+len(keySalt))
	i = copy(U, h[:32])
	i += copy(U[i:], valSalt)
	i += copy(U[i:], keySalt)

	crypt.U = U

	// step b: compute UE (user encryption)

	// str still contains a password, reuse it
	i = len(upass)
	i += copy(str[i:], keySalt)

	h = crypt.alg2b(str, upass, nil)

	ac, err := aes.NewCipher(h[:32])
	if err != nil {
		panic(err)
	}

	iv := crypt.aesZeroIV()
	cbc := cipher.NewCBCEncrypter(ac, iv)
	UE := make([]byte, 32)
	cbc.CryptBlocks(UE, crypt.EncryptionKey[:32])
	crypt.UE = UE

	return nil
}

// alg9 computes the encryption dictionary's O (owner password) and OE (owner encryption) values (R>=5).
// 7.6.4.4.7 Algorithm 9 (page 86)
func (crypt *PdfCrypt) alg9(opass []byte) error {
	// step a: compute O (owner password)
	var rbuf [16]byte
	if _, err := io.ReadFull(rand.Reader, rbuf[:]); err != nil {
		return err
	}
	valSalt := rbuf[0:8]
	keySalt := rbuf[8:16]
	userKey := crypt.U[:48]

	str := make([]byte, len(opass)+len(valSalt)+len(userKey))
	i := copy(str, opass)
	i += copy(str[i:], valSalt)
	i += copy(str[i:], userKey)

	h := crypt.alg2b(str, opass, userKey)

	O := make([]byte, len(h)+len(valSalt)+len(keySalt))
	i = copy(O, h[:32])
	i += copy(O[i:], valSalt)
	i += copy(O[i:], keySalt)

	crypt.O = O

	// step b: compute OE (owner encryption)

	// str still contains a password and a user key - reuse both, but overwrite the salt
	i = len(opass)
	i += copy(str[i:], keySalt)
	// i += len(userKey)

	h = crypt.alg2b(str, opass, userKey)

	ac, err := aes.NewCipher(h[:32])
	if err != nil {
		panic(err)
	}

	iv := crypt.aesZeroIV()
	cbc := cipher.NewCBCEncrypter(ac, iv)
	OE := make([]byte, 32)
	cbc.CryptBlocks(OE, crypt.EncryptionKey[:32])
	crypt.OE = OE

	return nil
}

// alg10 computes the encryption dictionary's Perms (permissions) value (R=6).
// 7.6.4.4.8 Algorithm 10 (page 87)
func (crypt *PdfCrypt) alg10() error {
	// step a: extend permissions to 64 bits
	perms := uint64(uint32(crypt.P)) | (math.MaxUint32 << 32)

	// step b: record permissions
	Perms := make([]byte, 16)
	binary.LittleEndian.PutUint64(Perms[:8], perms)

	// step c: record EncryptMetadata
	if crypt.EncryptMetadata {
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
	ac, err := aes.NewCipher(crypt.EncryptionKey[:32])
	if err != nil {
		panic(err)
	}

	ecb := newECBEncrypter(ac)
	ecb.CryptBlocks(Perms, Perms)

	crypt.Perms = Perms[:16]
	return nil
}

// alg11 authenticates the user password (R >= 5) and returns the hash.
func (crypt *PdfCrypt) alg11(upass []byte) ([]byte, error) {
	str := make([]byte, len(upass)+8)
	i := copy(str, upass)
	i += copy(str[i:], crypt.U[32:40]) // user Validation Salt

	h := crypt.alg2b(str, upass, nil)
	h = h[:32]
	if !bytes.Equal(h, crypt.U[:32]) {
		return nil, nil
	}
	return h, nil
}

// alg12 authenticates the owner password (R >= 5) and returns the hash.
// 7.6.4.4.10 Algorithm 12 (page 87)
func (crypt *PdfCrypt) alg12(opass []byte) ([]byte, error) {
	str := make([]byte, len(opass)+8+48)
	i := copy(str, opass)
	i += copy(str[i:], crypt.O[32:40]) // owner Validation Salt
	i += copy(str[i:], crypt.U[0:48])

	h := crypt.alg2b(str, opass, crypt.U[0:48])
	h = h[:32]
	if !bytes.Equal(h, crypt.O[:32]) {
		return nil, nil
	}
	return h, nil
}

// alg13 validates user permissions (P+EncryptMetadata vs Perms) for R=6.
// 7.6.4.4.11 Algorithm 13 (page 87)
func (crypt *PdfCrypt) alg13(fkey []byte) (bool, error) {
	perms := make([]byte, 16)
	copy(perms, crypt.Perms[:16])

	ac, err := aes.NewCipher(fkey[:32])
	if err != nil {
		panic(err)
	}

	ecb := newECBDecrypter(ac)
	ecb.CryptBlocks(perms, perms)

	if !bytes.Equal(perms[9:12], []byte("adb")) {
		return false, errors.New("decoded permissions are invalid")
	}
	p := int(int32(binary.LittleEndian.Uint32(perms[0:4])))
	if p != crypt.P {
		return false, errors.New("permissions validation failed")
	}
	encMeta := true
	if perms[8] == 'T' {
		encMeta = true
	} else if perms[8] == 'F' {
		encMeta = false
	} else {
		return false, errors.New("decoded metadata encryption flag is invalid")
	}
	if encMeta != crypt.EncryptMetadata {
		return false, errors.New("metadata encryption validation failed")
	}
	return true, nil
}
