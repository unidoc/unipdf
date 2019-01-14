/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"errors"

	"github.com/unidoc/unidoc/common/crypto/pkcs7"
	"github.com/unidoc/unidoc/pdf/core"
)

type adobePKCS7DetachedSignatureHandler struct {
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
}

// NewAdobePKCS7DetachedSignatureHandler creates a new Adobe.PPKMS/Adobe.PPKLite adbe.pkcs7.detached signature handler.
// The both parameters may be nil for the signature validation.
func NewAdobePKCS7DetachedSignatureHandler(privateKey *rsa.PrivateKey, certificate *x509.Certificate) (SignatureHandler, error) {
	return &adobePKCS7DetachedSignatureHandler{certificate: certificate, privateKey: privateKey}, nil
}

// InitSignature initialises the PdfSignature.
func (a *adobePKCS7DetachedSignatureHandler) InitSignature(sig *PdfSignature) error {
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
	sig.SubFilter = core.MakeName("adbe.pkcs7.detached")
	sig.Cert = core.MakeString(string(handler.certificate.Raw))

	sig.Reference = nil
	digest, err := handler.NewDigest(sig)
	if err != nil {
		return err
	}
	digest.Write([]byte("calculate the Contents field size"))
	return handler.Sign(sig, digest)
}

func (a *adobePKCS7DetachedSignatureHandler) getCertificate(sig *PdfSignature) (*x509.Certificate, error) {
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
func (a *adobePKCS7DetachedSignatureHandler) NewDigest(sig *PdfSignature) (Hasher, error) {
	return bytes.NewBuffer(nil), nil
}

// Validate validates PdfSignature.
func (a *adobePKCS7DetachedSignatureHandler) Validate(sig *PdfSignature, digest Hasher) (SignatureValidationResult, error) {
	signed := sig.Contents.Bytes()

	buffer := digest.(*bytes.Buffer)
	p7, err := pkcs7.Parse(signed)
	if err != nil {
		return SignatureValidationResult{}, err
	}
	p7.Content = buffer.Bytes()
	err = p7.Verify()
	if err != nil {
		return SignatureValidationResult{}, err
	}

	return SignatureValidationResult{IsSigned: true, IsVerified: true}, nil
}

// Sign sets the Contents fields.
func (a *adobePKCS7DetachedSignatureHandler) Sign(sig *PdfSignature, digest Hasher) error {

	buffer := digest.(*bytes.Buffer)
	signedData, err := pkcs7.NewSignedData(buffer.Bytes())
	if err != nil {
		return err
	}

	// Add the signing cert and private key
	if err := signedData.AddSigner(a.certificate, a.privateKey, pkcs7.SignerInfoConfig{}); err != nil {
		return err
	}

	// Call Detach() is you want to remove content from the signature
	// and generate an S/MIME detached signature
	signedData.Detach()
	// Finish() to obtain the signature bytes
	detachedSignature, err := signedData.Finish()
	if err != nil {
		return err
	}

	data := make([]byte, 8192)
	copy(data, detachedSignature)

	sig.Contents = core.MakeHexString(string(data))
	return nil
}

// IsApplicable returns true if the signature handler is applicable for the PdfSignature
func (a *adobePKCS7DetachedSignatureHandler) IsApplicable(sig *PdfSignature) bool {
	if sig == nil || sig.Filter == nil || sig.SubFilter == nil {
		return false
	}
	return (*sig.Filter == "Adobe.PPKMS" || *sig.Filter == "Adobe.PPKLite") && *sig.SubFilter == "adbe.pkcs7.detached"
}
