/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
)

type stdSecurityHandler interface {
	// GenerateParams uses owner and user passwords to set encryption parameters and generate an encryption key.
	// It assumes that R, P and EncryptMetadata are already set.
	GenerateParams(d *stdEncryptDict, ownerPass, userPass []byte) ([]byte, error)

	// Authenticate uses encryption dictionary parameters and the password to calculate
	// the document encryption key. It also returns permissions that should be granted to a user.
	// In case of failed authentication, it returns empty key and zero permissions with no error.
	Authenticate(d *stdEncryptDict, pass []byte) ([]byte, AccessPermissions, error)
}

// stdEncryptDict is a set of additional fields used in standard encryption dictionary.
type stdEncryptDict struct {
	R int // (Required) A number specifying which revision of the standard security handler shall be used.

	P               AccessPermissions
	EncryptMetadata bool // Indicates whether the document-level metadata stream shall be encrypted.

	// set by security handlers:

	O, U   []byte
	OE, UE []byte // R=6
	Perms  []byte // An encrypted copy of P (16 bytes). Used to verify permissions. R=6
}

func (d *stdEncryptDict) EncodeTo(ed *PdfObjectDictionary) {
	ed.Set("R", MakeInteger(int64(d.R)))
	ed.Set("P", MakeInteger(int64(d.P)))

	ed.Set("O", MakeStringFromBytes(d.O))
	ed.Set("U", MakeStringFromBytes(d.U))
	if d.R >= 5 {
		ed.Set("OE", MakeStringFromBytes(d.OE))
		ed.Set("UE", MakeStringFromBytes(d.UE))
		ed.Set("EncryptMetadata", MakeBool(d.EncryptMetadata))
		if d.R > 5 {
			ed.Set("Perms", MakeStringFromBytes(d.Perms))
		}
	}
}

func (d *stdEncryptDict) DecodeFrom(ed *PdfObjectDictionary) error {
	// TODO(dennwc): this code is too verbose; maybe use reflection to populate fields and validate afterwards?
	R, ok := ed.Get("R").(*PdfObjectInteger)
	if !ok {
		return errors.New("Encrypt dictionary missing R")
	}
	// TODO(dennwc): according to spec, R should be validated according to V value
	if *R < 2 || *R > 6 {
		return fmt.Errorf("Invalid R (%d)", *R)
	}
	d.R = int(*R)

	O, ok := ed.Get("O").(*PdfObjectString)
	if !ok {
		return errors.New("Encrypt dictionary missing O")
	}
	if d.R == 5 || d.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(O.Str()) < 48 {
			return fmt.Errorf("Length(O) < 48 (%d)", len(O.Str()))
		}
	} else if len(O.Str()) != 32 {
		return fmt.Errorf("Length(O) != 32 (%d)", len(O.Str()))
	}
	d.O = O.Bytes()

	U, ok := ed.Get("U").(*PdfObjectString)
	if !ok {
		return errors.New("Encrypt dictionary missing U")
	}
	if d.R == 5 || d.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(U.Str()) < 48 {
			return fmt.Errorf("Length(U) < 48 (%d)", len(U.Str()))
		}
	} else if len(U.Str()) != 32 {
		// Strictly this does not cause an error.
		// If O is OK and others then can still read the file.
		common.Log.Debug("Warning: Length(U) != 32 (%d)", len(U.Str()))
		//return crypter, errors.New("Length(U) != 32")
	}
	d.U = U.Bytes()

	if d.R >= 5 {
		OE, ok := ed.Get("OE").(*PdfObjectString)
		if !ok {
			return errors.New("Encrypt dictionary missing OE")
		}
		if len(OE.Str()) != 32 {
			return fmt.Errorf("Length(OE) != 32 (%d)", len(OE.Str()))
		}
		d.OE = OE.Bytes()

		UE, ok := ed.Get("UE").(*PdfObjectString)
		if !ok {
			return errors.New("Encrypt dictionary missing UE")
		}
		if len(UE.Str()) != 32 {
			return fmt.Errorf("Length(UE) != 32 (%d)", len(UE.Str()))
		}
		d.UE = UE.Bytes()
	}

	P, ok := ed.Get("P").(*PdfObjectInteger)
	if !ok {
		return errors.New("Encrypt dictionary missing permissions attr")
	}
	d.P = AccessPermissions(*P)

	if d.R == 6 {
		Perms, ok := ed.Get("Perms").(*PdfObjectString)
		if !ok {
			return errors.New("Encrypt dictionary missing Perms")
		}
		if len(Perms.Str()) != 16 {
			return fmt.Errorf("Length(Perms) != 16 (%d)", len(Perms.Str()))
		}
		d.Perms = Perms.Bytes()
	}

	em, ok := ed.Get("EncryptMetadata").(*PdfObjectBool)
	if ok {
		d.EncryptMetadata = bool(*em)
	} else {
		d.EncryptMetadata = true // True by default.
	}
	return nil
}
