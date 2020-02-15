package sighandler

import (
	"bytes"
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

// docTimeStamp DocTimeStamp signature handler.
type docTimeStamp struct {
	signFunc SignFunc
}

// NewDocTimeStamp creates a new DocTimeStamp signature handler.
// Both parameters may be nil for the signature validation.
func NewDocTimeStamp() (model.SignatureHandler, error) {
	return &docTimeStamp{}, nil
}

// InitSignature initialises the PdfSignature.
func (a *docTimeStamp) InitSignature(sig *model.PdfSignature) error {
	handler := *a
	sig.Handler = &handler
	sig.Filter = core.MakeName("Adobe.PPKLite")
	sig.SubFilter = core.MakeName("ETSI.RFC3161")
	sig.Reference = nil

	digest, err := handler.NewDigest(sig)
	if err != nil {
		return err
	}
	digest.Write([]byte("calculate the Contents field size"))
	return handler.Sign(sig, digest)
}

func (a *docTimeStamp) getCertificate(sig *model.PdfSignature) (*x509.Certificate, error) {
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
func (a *docTimeStamp) NewDigest(sig *model.PdfSignature) (model.Hasher, error) {
	return bytes.NewBuffer(nil), nil
}

// Validate validates PdfSignature.
func (a *docTimeStamp) Validate(sig *model.PdfSignature, digest model.Hasher) (model.SignatureValidationResult, error) {
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
func (a *docTimeStamp) Sign(sig *model.PdfSignature, digest model.Hasher) error {
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
func (a *docTimeStamp) IsApplicable(sig *model.PdfSignature) bool {
	if sig == nil || sig.Filter == nil || sig.SubFilter == nil {
		return false
	}
	return (*sig.Filter == "Adobe.PPKMS" || *sig.Filter == "Adobe.PPKLite") && *sig.SubFilter == "ETSI.RFC3161"
}
