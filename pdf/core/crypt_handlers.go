/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

type stdSecurityHandler interface {
	// GenerateParams uses owner and user passwords to set encryption parameters and generate an encryption key.
	// It assumes that R, P and EncryptMetadata are already set.
	GenerateParams(d *stdEncryptDict, ownerPass, userPass []byte) ([]byte, error)

	// Authenticate uses encryption dictionary parameters and the password to calculate
	// the document encryption key. It also returns permissions that should be granted to a user.
	// In case of failed authentication, it returns empty key and zero permissions with no error.
	Authenticate(d *stdEncryptDict, pass []byte) ([]byte, AccessPermissions, error)
}
