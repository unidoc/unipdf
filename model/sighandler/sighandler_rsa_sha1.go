/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package sighandler

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"
	"hash"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// SignFunc represents a custom signing function. The function should return
// the computed signature.
type SignFunc func(sig *model.PdfSignature, digest model.Hasher) ([]byte, error)

// Adobe X509 RSA SHA1 signature handler.
type adobeX509RSASHA1 struct {
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
	signFunc    SignFunc
}

// NewAdobeX509RSASHA1Custom creates a new Adobe.PPKMS/Adobe.PPKLite adbe.x509.rsa_sha1 signature handler
// with a custom signing function. Both parameters may be nil for the signature validation.
func NewAdobeX509RSASHA1Custom(certificate *x509.Certificate, signFunc SignFunc) (model.SignatureHandler, error) {
	return &adobeX509RSASHA1{certificate: certificate, signFunc: signFunc}, nil
}

// NewAdobeX509RSASHA1 creates a new Adobe.PPKMS/Adobe.PPKLite adbe.x509.rsa_sha1 signature handler.
// Both parameters may be nil for the signature validation.
func NewAdobeX509RSASHA1(privateKey *rsa.PrivateKey, certificate *x509.Certificate) (model.SignatureHandler, error) {
	return &adobeX509RSASHA1{certificate: certificate, privateKey: privateKey}, nil
}

// InitSignature initialises the PdfSignature.
func (a *adobeX509RSASHA1) InitSignature(sig *model.PdfSignature) error {
	if a.certificate == nil {
		return errors.New("certificate must not be nil")
	}
	if a.privateKey == nil && a.signFunc == nil {
		return errors.New("must provide either a private key or a signing function")
	}

	handler := *a
	sig.Handler = &handler
	sig.Filter = core.MakeName("Adobe.PPKLite")
	sig.SubFilter = core.MakeName("adbe.x509.rsa_sha1")
	sig.Cert = core.MakeString(string(handler.certificate.Raw))
	sig.Reference = nil

	digest, err := handler.NewDigest(sig)
	if err != nil {
		return err
	}
	digest.Write([]byte("calculate the Contents field size"))
	return handler.Sign(sig, digest)
}

func getHashFromSignatureAlgorithm(sa x509.SignatureAlgorithm) (crypto.Hash, bool) {
	return crypto.SHA1, true
}

func (a *adobeX509RSASHA1) getCertificate(sig *model.PdfSignature) (*x509.Certificate, error) {
	if a.certificate != nil {
		return a.certificate, nil
	}

	var certData []byte
	switch certObj := sig.Cert.(type) {
	case *core.PdfObjectString:
		certData = certObj.Bytes()
	case *core.PdfObjectArray:
		if certObj.Len() == 0 {
			return nil, errors.New("no signature certificates found")
		}
		for _, obj := range certObj.Elements() {
			certStr, ok := core.GetString(obj)
			if !ok {
				return nil, fmt.Errorf("invalid certificate object type in signature certificate chain: %T", obj)
			}
			certData = append(certData, certStr.Bytes()...)
		}
	default:
		return nil, fmt.Errorf("invalid signature certificate object type: %T", certObj)
	}

	certs, err := x509.ParseCertificates(certData)
	if err != nil {
		return nil, err
	}

	return certs[0], nil
}

// NewDigest creates a new digest.
func (a *adobeX509RSASHA1) NewDigest(sig *model.PdfSignature) (model.Hasher, error) {
	certificate, err := a.getCertificate(sig)
	if err != nil {
		return nil, err
	}
	h, _ := getHashFromSignatureAlgorithm(certificate.SignatureAlgorithm)
	return h.New(), nil
}

// Validate validates PdfSignature.
func (a *adobeX509RSASHA1) Validate(sig *model.PdfSignature, digest model.Hasher) (model.SignatureValidationResult, error) {
	certificate, err := a.getCertificate(sig)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}

	signed := sig.Contents.Bytes()
	var sigHash []byte
	if _, err := asn1.Unmarshal(signed, &sigHash); err != nil {
		return model.SignatureValidationResult{}, err
	}
	h, ok := digest.(hash.Hash)
	if !ok {
		return model.SignatureValidationResult{}, errors.New("hash type error")
	}
	ha, _ := getHashFromSignatureAlgorithm(certificate.SignatureAlgorithm)
	if err := rsa.VerifyPKCS1v15(certificate.PublicKey.(*rsa.PublicKey), ha, h.Sum(nil), sigHash); err != nil {
		return model.SignatureValidationResult{}, err
	}
	return model.SignatureValidationResult{IsSigned: true, IsVerified: true}, nil
}

// Sign sets the Contents fields for the PdfSignature.
func (a *adobeX509RSASHA1) Sign(sig *model.PdfSignature, digest model.Hasher) error {
	var data []byte
	var err error

	if a.signFunc != nil {
		data, err = a.signFunc(sig, digest)
		if err != nil {
			return err
		}
	} else {
		h, ok := digest.(hash.Hash)
		if !ok {
			return errors.New("hash type error")
		}
		ha, _ := getHashFromSignatureAlgorithm(a.certificate.SignatureAlgorithm)

		data, err = rsa.SignPKCS1v15(rand.Reader, a.privateKey, ha, h.Sum(nil))
		if err != nil {
			return err
		}
	}

	data, err = asn1.Marshal(data)
	if err != nil {
		return err
	}

	sig.Contents = core.MakeHexString(string(data))
	return nil
}

// IsApplicable returns true if the signature handler is applicable for the PdfSignature.
func (a *adobeX509RSASHA1) IsApplicable(sig *model.PdfSignature) bool {
	if sig == nil || sig.Filter == nil || sig.SubFilter == nil {
		return false
	}
	return (*sig.Filter == "Adobe.PPKMS" || *sig.Filter == "Adobe.PPKLite") && *sig.SubFilter == "adbe.x509.rsa_sha1"
}
