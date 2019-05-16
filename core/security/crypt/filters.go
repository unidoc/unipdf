/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package crypt

import (
	"fmt"

	"github.com/unidoc/unipdf/v3/core/security"
)

var (
	filterMethods = make(map[string]filterFunc)
)

// filterFunc is used to construct crypt filters from CryptFilter dictionary
type filterFunc func(d FilterDict) (Filter, error)

// Filter is a common interface for crypt filter methods.
type Filter interface {
	// Name returns a name of the filter that should be used in CFM field of Encrypt dictionary.
	Name() string
	// KeyLength returns a length of the encryption key in bytes.
	KeyLength() int
	// PDFVersion reports the minimal version of PDF document that introduced this filter.
	PDFVersion() [2]int
	// HandlerVersion reports V and R parameters that should be used for this filter.
	HandlerVersion() (V, R int)
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

// NewFilter creates CryptFilter from a corresponding dictionary.
func NewFilter(d FilterDict) (Filter, error) {
	fnc, err := getFilter(d.CFM)
	if err != nil {
		return nil, err
	}
	cf, err := fnc(d)
	if err != nil {
		return nil, err
	}
	return cf, nil
}

// NewIdentity creates an identity filter that bypasses all data without changes.
func NewIdentity() Filter {
	return filterIdentity{}
}

// FilterDict represents information from a CryptFilter dictionary.
type FilterDict struct {
	CFM       string // The method used, if any, by the PDF reader to decrypt data.
	AuthEvent security.AuthEvent
	Length    int // in bytes
}

// registerFilter register supported crypt filter methods.
// Table 25, CFM (page 92)
func registerFilter(name string, fnc filterFunc) {
	if _, ok := filterMethods[name]; ok {
		panic("already registered")
	}
	filterMethods[name] = fnc
}

// getFilter check if a CFM with a specified name is supported an returns its implementation.
func getFilter(name string) (filterFunc, error) {
	f := filterMethods[string(name)]
	if f == nil {
		return nil, fmt.Errorf("unsupported crypt filter: %q", name)
	}
	return f, nil
}

type filterIdentity struct{}

func (filterIdentity) PDFVersion() [2]int {
	return [2]int{}
}

func (filterIdentity) HandlerVersion() (V, R int) {
	return
}

func (filterIdentity) Name() string {
	return "Identity"
}

func (filterIdentity) KeyLength() int {
	return 0
}

func (filterIdentity) MakeKey(objNum, genNum uint32, fkey []byte) ([]byte, error) {
	return fkey, nil
}

func (filterIdentity) EncryptBytes(p []byte, okey []byte) ([]byte, error) {
	return p, nil
}

func (filterIdentity) DecryptBytes(p []byte, okey []byte) ([]byte, error) {
	return p, nil
}
