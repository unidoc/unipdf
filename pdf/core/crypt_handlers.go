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
		return errors.New("encrypt dictionary missing R")
	}
	// TODO(dennwc): according to spec, R should be validated according to V value
	if *R < 2 || *R > 6 {
		return fmt.Errorf("invalid R (%d)", *R)
	}
	d.R = int(*R)

	O, ok := ed.GetString("O")
	if !ok {
		return errors.New("encrypt dictionary missing O")
	}
	if d.R == 5 || d.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(O) < 48 {
			return fmt.Errorf("Length(O) < 48 (%d)", len(O))
		}
	} else if len(O) != 32 {
		return fmt.Errorf("Length(O) != 32 (%d)", len(O))
	}
	d.O = []byte(O)

	U, ok := ed.GetString("U")
	if !ok {
		return errors.New("encrypt dictionary missing U")
	}
	if d.R == 5 || d.R == 6 {
		// the spec says =48 bytes, but Acrobat pads them out longer
		if len(U) < 48 {
			return fmt.Errorf("Length(U) < 48 (%d)", len(U))
		}
	} else if len(U) != 32 {
		// Strictly this does not cause an error.
		// If O is OK and others then can still read the file.
		common.Log.Debug("Warning: Length(U) != 32 (%d)", len(U))
		//return crypter, errors.New("Length(U) != 32")
	}
	d.U = []byte(U)

	if d.R >= 5 {
		OE, ok := ed.GetString("OE")
		if !ok {
			return errors.New("encrypt dictionary missing OE")
		} else if len(OE) != 32 {
			return fmt.Errorf("Length(OE) != 32 (%d)", len(OE))
		}
		d.OE = []byte(OE)

		UE, ok := ed.GetString("UE")
		if !ok {
			return errors.New("encrypt dictionary missing UE")
		} else if len(UE) != 32 {
			return fmt.Errorf("Length(UE) != 32 (%d)", len(UE))
		}
		d.UE = []byte(UE)
	}

	P, ok := ed.Get("P").(*PdfObjectInteger)
	if !ok {
		return errors.New("encrypt dictionary missing permissions attr")
	}
	d.P = AccessPermissions(*P)

	if d.R == 6 {
		Perms, ok := ed.GetString("Perms")
		if !ok {
			return errors.New("encrypt dictionary missing Perms")
		} else if len(Perms) != 16 {
			return fmt.Errorf("Length(Perms) != 16 (%d)", len(Perms))
		}
		d.Perms = []byte(Perms)
	}

	if em, ok := ed.Get("EncryptMetadata").(*PdfObjectBool); ok {
		d.EncryptMetadata = bool(*em)
	} else {
		d.EncryptMetadata = true // True by default.
	}
	return nil
}
