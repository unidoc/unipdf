/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"hash"
	"io"

	"github.com/unidoc/unidoc/pdf/core"
)

// Digest is the interface that wraps the basic Write method.
type Digest interface {
	Write(p []byte) (n int, err error)
}

// SignatureHandler interface defines the common functionality for PDF signature handlers, which
// need to be capable of validating digital signatures and signing PDF documents.
type SignatureHandler interface {
	IsApplicable(sig *PdfSignature) bool
	Validate(sig *PdfSignature, digest Digest) (SignatureValidationResult, error)
	// InitSignature sets the PdfSignature parameters.
	InitSignature(*PdfSignature) error
	NewDigest(sig *PdfSignature) (Digest, error)
	Sign(sig *PdfSignature, digest Digest) error
}

// SignatureValidationResult defines the response from the signature validation handler.
type SignatureValidationResult struct {
	IsSigned    bool
	IsVerified  bool
	IsTrusted   bool
	Fields      []*PdfField
	Name        string
	Date        PdfDate
	Reason      string
	Location    string
	ContactInfo string
}

type adobeX509RSASHA1SignatureHandler struct {
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
}

// NewAdobeX509RSASHA1SignatureHandler creates a new Adobe.PPKMS/Adobe.PPKLite adbe.x509.rsa_sha1 signature handler.
// The both parameters may be nil for the signature validation.
func NewAdobeX509RSASHA1SignatureHandler(privateKey *rsa.PrivateKey, certificate *x509.Certificate) (SignatureHandler, error) {
	return &adobeX509RSASHA1SignatureHandler{certificate: certificate, privateKey: privateKey}, nil
}

// InitSignature initialises the PdfSignature.
func (a *adobeX509RSASHA1SignatureHandler) InitSignature(sig *PdfSignature) error {
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

func (a *adobeX509RSASHA1SignatureHandler) getCertificate(sig *PdfSignature) (*x509.Certificate, error) {
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
func (a *adobeX509RSASHA1SignatureHandler) NewDigest(sig *PdfSignature) (Digest, error) {
	certificate, err := a.getCertificate(sig)
	if err != nil {
		return nil, err
	}
	h, _ := getHashFromSignatureAlgorithm(certificate.SignatureAlgorithm)
	return h.New(), nil
}

// Validate validates PdfSignature.
func (a *adobeX509RSASHA1SignatureHandler) Validate(sig *PdfSignature, digest Digest) (SignatureValidationResult, error) {
	certData := sig.Cert.(*core.PdfObjectString).Bytes()
	certs, err := x509.ParseCertificates(certData)
	if err != nil {
		return SignatureValidationResult{}, err
	}
	if len(certs) == 0 {
		return SignatureValidationResult{}, errors.New("certificate not found")
	}
	cert := certs[0]
	signed := sig.Contents.Bytes()
	var sigHash []byte
	if _, err := asn1.Unmarshal(signed, &sigHash); err != nil {
		return SignatureValidationResult{}, err
	}
	h, ok := digest.(hash.Hash)
	if !ok {
		return SignatureValidationResult{}, errors.New("hash type error")
	}
	certificate, err := a.getCertificate(sig)
	if err != nil {
		return SignatureValidationResult{}, err
	}
	ha, _ := getHashFromSignatureAlgorithm(certificate.SignatureAlgorithm)
	if err := rsa.VerifyPKCS1v15(cert.PublicKey.(*rsa.PublicKey), ha, h.Sum(nil), sigHash); err != nil {
		return SignatureValidationResult{}, err
	}
	return SignatureValidationResult{IsSigned: true, IsVerified: true}, nil
}

// Sign sets the Contents fields.
func (a *adobeX509RSASHA1SignatureHandler) Sign(sig *PdfSignature, digest Digest) error {
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

// IsApplicable returns true if the signature handler is applicable for the PdfSignature
func (a *adobeX509RSASHA1SignatureHandler) IsApplicable(sig *PdfSignature) bool {
	if sig == nil || sig.Filter == nil || sig.SubFilter == nil {
		return false
	}
	return (*sig.Filter == "Adobe.PPKMS" || *sig.Filter == "Adobe.PPKLite") && *sig.SubFilter == "adbe.x509.rsa_sha1"
}

// Validate validates signatures.
func (r *PdfReader) Validate(handlers []SignatureHandler) ([]SignatureValidationResult, error) {
	if r.AcroForm == nil {
		return nil, nil
	}
	if r.AcroForm.Fields == nil {
		return nil, nil
	}
	type sigFieldPair struct {
		sig     *PdfSignature
		field   *PdfField
		handler SignatureHandler
	}
	var pairs []*sigFieldPair

	for _, f := range *r.AcroForm.Fields {
		if f.V == nil {
			continue
		}
		if d, found := core.GetDict(f.V); found {
			if name, ok := core.GetNameVal(d.Get("Type")); ok && name == "Sig" {
				ind, _ := core.GetIndirect(f.V) // TODO refactoring
				sig, err := r.newPdfSignatureFromIndirect(ind)
				if err != nil {
					return nil, err
				}
				pairs = append(pairs, &sigFieldPair{sig: sig, field: f})
			}
		}
	}

	for _, pair := range pairs {
		for _, handler := range handlers {
			if handler.IsApplicable(pair.sig) {
				pair.handler = handler
				break
			}
		}
	}

	var results []SignatureValidationResult

	for _, pair := range pairs {
		defaultResult := SignatureValidationResult{IsSigned: true, Fields: []*PdfField{pair.field}}
		if pair.handler == nil {
			// TODO think about variants
			//  to throw an error
			//  to skip the field and add error message to the result
			results = append(results, defaultResult)
			continue
		}
		digest, err := pair.handler.NewDigest(pair.sig)
		if err != nil {
			// TODO think about variants
			//  to throw an error
			//  to skip the field and add error message to the result
			results = append(results, defaultResult)
			continue
		}
		byteRange := pair.sig.ByteRange
		if byteRange == nil {
			// TODO think about variants
			//  to throw an error
			//  to skip the field and add error message to the result
			results = append(results, defaultResult)
			continue
		}
		for i := 0; i < byteRange.Len(); i = i + 2 {
			start, _ := core.GetNumberAsInt64(byteRange.Get(i))
			ln, _ := core.GetIntVal(byteRange.Get(i + 1))
			if _, err := r.rs.Seek(start, io.SeekStart); err != nil {
				return nil, err
			}
			data := make([]byte, ln)
			if _, err := r.rs.Read(data); err != nil {
				return nil, err
			}
			digest.Write(data)
		}

		result, err := pair.handler.Validate(pair.sig, digest)
		if err != nil {
			return nil, err
		}
		result.Fields = defaultResult.Fields
		results = append(results, result)
	}
	return results, nil
}
