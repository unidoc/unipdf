/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package security

// StdHandler is an interface for standard security handlers.
type StdHandler interface {
	// GenerateParams uses owner and user passwords to set encryption parameters and generate an encryption key.
	// It assumes that R, P and EncryptMetadata are already set.
	GenerateParams(d *StdEncryptDict, ownerPass, userPass []byte) ([]byte, error)

	// Authenticate uses encryption dictionary parameters and the password to calculate
	// the document encryption key. It also returns permissions that should be granted to a user.
	// In case of failed authentication, it returns empty key and zero permissions with no error.
	Authenticate(d *StdEncryptDict, pass []byte) ([]byte, Permissions, error)
}

// StdEncryptDict is a set of additional fields used in standard encryption dictionary.
type StdEncryptDict struct {
	R int // (Required) A number specifying which revision of the standard security handler shall be used.

	P               Permissions
	EncryptMetadata bool // Indicates whether the document-level metadata stream shall be encrypted.

	// set by security handlers:

	O, U   []byte
	OE, UE []byte // R=6
	Perms  []byte // An encrypted copy of P (16 bytes). Used to verify permissions. R=6
}
