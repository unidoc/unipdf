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
	"hash"

	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// Adobe X509 RSA SHA1 signature handler.
type adobeX509RSASHA1 struct {
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
}

// NewAdobeX509RSASHA1 creates a new Adobe.PPKMS/Adobe.PPKLite adbe.x509.rsa_sha1 signature handler.
// The both parameters may be nil for the signature validation.
func NewAdobeX509RSASHA1(privateKey *rsa.PrivateKey, certificate *x509.Certificate) (model.SignatureHandler, error) {
	return &adobeX509RSASHA1{certificate: certificate, privateKey: privateKey}, nil
}

// InitSignature initialises the PdfSignature.
func (a *adobeX509RSASHA1) InitSignature(sig *model.PdfSignature) error {
	if a.certificate == nil {
		return errors.New("certificate must not be nil")
	}
	if a.privateKey == nil {
		return errors.New("privateKey must not be nil")
	}

	handler := *a
	sig.Handler = &handler
	sig.Filter = core.MakeName("Adobe.PPKLite")
	//sig.Filter = core.MakeName("Adobe.PPKMS")
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
	/*
		switch sa {
		case x509.SHA1WithRSA:
			return crypto.SHA1, true
		case x509.SHA256WithRSA:
			return crypto.SHA256, true
		case x509.SHA512WithRSA:
			return crypto.SHA512, true
		}
		return crypto.MD5, false
	*/
}

func (a *adobeX509RSASHA1) getCertificate(sig *model.PdfSignature) (*x509.Certificate, error) {
	certificate := a.certificate
	if certificate == nil {
		certData := sig.Cert.(*core.PdfObjectString).Bytes()
		certs, err := x509.ParseCertificates(certData)
		if err != nil {
			return nil, err
		}
		certificate = certs[0]
	}
	return certificate, nil
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
	certData := sig.Cert.(*core.PdfObjectString).Bytes()
	certs, err := x509.ParseCertificates(certData)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}
	if len(certs) == 0 {
		return model.SignatureValidationResult{}, errors.New("certificate not found")
	}
	cert := certs[0]
	signed := sig.Contents.Bytes()
	var sigHash []byte
	if _, err := asn1.Unmarshal(signed, &sigHash); err != nil {
		return model.SignatureValidationResult{}, err
	}
	h, ok := digest.(hash.Hash)
	if !ok {
		return model.SignatureValidationResult{}, errors.New("hash type error")
	}
	certificate, err := a.getCertificate(sig)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}
	ha, _ := getHashFromSignatureAlgorithm(certificate.SignatureAlgorithm)
	if err := rsa.VerifyPKCS1v15(cert.PublicKey.(*rsa.PublicKey), ha, h.Sum(nil), sigHash); err != nil {
		return model.SignatureValidationResult{}, err
	}
	return model.SignatureValidationResult{IsSigned: true, IsVerified: true}, nil
}

// Sign sets the Contents fields for the PdfSignature.
func (a *adobeX509RSASHA1) Sign(sig *model.PdfSignature, digest model.Hasher) error {
	h, ok := digest.(hash.Hash)
	if !ok {
		return errors.New("hash type error")
	}
	ha, _ := getHashFromSignatureAlgorithm(a.certificate.SignatureAlgorithm)
	data, err := rsa.SignPKCS1v15(rand.Reader, a.privateKey, ha, h.Sum(nil))
	if err != nil {
		return err
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
