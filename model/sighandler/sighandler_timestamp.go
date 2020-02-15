<<<<<<< HEAD
/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

=======
>>>>>>> sync commit
package sighandler

import (
	"bytes"
	"crypto"
<<<<<<< HEAD
<<<<<<< HEAD
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/unidoc/pkcs7"
	"github.com/unidoc/timestamp"
=======
	"crypto/rand"
	"crypto/rsa"
=======
>>>>>>> add timestamp signature handler
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

<<<<<<< HEAD
>>>>>>> sync commit
=======
	"github.com/a5i/pkcs7"
	"github.com/digitorus/timestamp"
>>>>>>> add timestamp signature handler
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// docTimeStamp DocTimeStamp signature handler.
type docTimeStamp struct {
<<<<<<< HEAD
<<<<<<< HEAD
	timestampServerURL string
	hashAlgorithm      crypto.Hash
}

// NewDocTimeStamp creates a new DocTimeStamp signature handler.
// The timestampServerURL parameter can be empty string for the signature validation.
// The hashAlgorithm parameter can be crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512.
func NewDocTimeStamp(timestampServerURL string, hashAlgorithm crypto.Hash) (model.SignatureHandler, error) {
	return &docTimeStamp{
		timestampServerURL: timestampServerURL,
		hashAlgorithm:      hashAlgorithm,
	}, nil
=======
	signFunc SignFunc
=======
	timestampServerURL string
	signFunc           SignFunc
	hashAlgorithm      crypto.Hash
	emptySignatureLen  int
>>>>>>> add timestamp signature handler
}

// NewDocTimeStamp creates a new DocTimeStamp signature handler.
// Both parameters may be nil for the signature validation.
<<<<<<< HEAD
func NewDocTimeStamp() (model.SignatureHandler, error) {
	return &docTimeStamp{}, nil
>>>>>>> sync commit
=======
// The timestampServerURL parameter can be empty string for the signature validation.
// The signatureLen parameter can be 0 for the signature validation.
func NewDocTimeStamp(timestampServerURL string, signatureLen int) (model.SignatureHandler, error) {
	return &docTimeStamp{
		timestampServerURL: timestampServerURL,
		emptySignatureLen:  signatureLen,
		hashAlgorithm:      crypto.SHA512,
	}, nil
>>>>>>> add timestamp signature handler
}

// InitSignature initialises the PdfSignature.
func (a *docTimeStamp) InitSignature(sig *model.PdfSignature) error {
	handler := *a
	sig.Handler = &handler
	sig.Filter = core.MakeName("Adobe.PPKLite")
	sig.SubFilter = core.MakeName("ETSI.RFC3161")
	sig.Reference = nil
<<<<<<< HEAD
	digest, err := a.NewDigest(sig)
=======

	digest, err := handler.NewDigest(sig)
>>>>>>> sync commit
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

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> add timestamp signature handler
type timestampInfo struct {
	Version        int
	Policy         asn1.RawValue
	MessageImprint struct {
		HashAlgorithm pkix.AlgorithmIdentifier
		HashedMessage []byte
	}
<<<<<<< HEAD
	SerialNumber    asn1.RawValue
	GeneralizedTime time.Time
=======
>>>>>>> add timestamp signature handler
}

func getHashForOID(oid asn1.ObjectIdentifier) (crypto.Hash, error) {
	switch {
	case oid.Equal(pkcs7.OIDDigestAlgorithmSHA1), oid.Equal(pkcs7.OIDDigestAlgorithmECDSASHA1),
		oid.Equal(pkcs7.OIDDigestAlgorithmDSA), oid.Equal(pkcs7.OIDDigestAlgorithmDSASHA1),
		oid.Equal(pkcs7.OIDEncryptionAlgorithmRSA):
		return crypto.SHA1, nil
	case oid.Equal(pkcs7.OIDDigestAlgorithmSHA256), oid.Equal(pkcs7.OIDDigestAlgorithmECDSASHA256):
		return crypto.SHA256, nil
	case oid.Equal(pkcs7.OIDDigestAlgorithmSHA384), oid.Equal(pkcs7.OIDDigestAlgorithmECDSASHA384):
		return crypto.SHA384, nil
	case oid.Equal(pkcs7.OIDDigestAlgorithmSHA512), oid.Equal(pkcs7.OIDDigestAlgorithmECDSASHA512):
		return crypto.SHA512, nil
	}
	return crypto.Hash(0), pkcs7.ErrUnsupportedAlgorithm
}

<<<<<<< HEAD
// Validate validates PdfSignature.
func (a *docTimeStamp) Validate(sig *model.PdfSignature, digest model.Hasher) (model.SignatureValidationResult, error) {
	signed := sig.Contents.Bytes()
	p7, err := pkcs7.Parse(signed)
=======
// Validate validates PdfSignature.
func (a *docTimeStamp) Validate(sig *model.PdfSignature, digest model.Hasher) (model.SignatureValidationResult, error) {
	certificate, err := a.getCertificate(sig)
>>>>>>> sync commit
	if err != nil {
		return model.SignatureValidationResult{}, err
	}

<<<<<<< HEAD
	if err = p7.Verify(); err != nil {
		return model.SignatureValidationResult{}, err
	}

	var tsInfo timestampInfo

	_, err = asn1.Unmarshal(p7.Content, &tsInfo)
=======
// Validate validates PdfSignature.
func (a *docTimeStamp) Validate(sig *model.PdfSignature, digest model.Hasher) (model.SignatureValidationResult, error) {
	signed := sig.Contents.Bytes()
	p7, err := pkcs7.Parse(signed)
>>>>>>> add timestamp signature handler
	if err != nil {
		return model.SignatureValidationResult{}, err
	}

<<<<<<< HEAD
	hAlg, err := getHashForOID(tsInfo.MessageImprint.HashAlgorithm.Algorithm)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}
	h := hAlg.New()
	buffer := digest.(*bytes.Buffer)

	h.Write(buffer.Bytes())
	sm := h.Sum(nil)
	res := model.SignatureValidationResult{
		IsSigned:        true,
		IsVerified:      bytes.Equal(sm, tsInfo.MessageImprint.HashedMessage),
		GeneralizedTime: tsInfo.GeneralizedTime,
	}
	return res, nil
=======
	signed := sig.Contents.Bytes()
	var sigHash []byte
	if _, err := asn1.Unmarshal(signed, &sigHash); err != nil {
=======
	if err = p7.Verify(); err != nil {
>>>>>>> add timestamp signature handler
		return model.SignatureValidationResult{}, err
	}

	var tsInfo timestampInfo

	_, err = asn1.Unmarshal(p7.Content, &tsInfo)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}

	hAlg, err := getHashForOID(tsInfo.MessageImprint.HashAlgorithm.Algorithm)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}
<<<<<<< HEAD
	return model.SignatureValidationResult{IsSigned: true, IsVerified: true}, nil
>>>>>>> sync commit
=======
	h := hAlg.New()
	buffer := digest.(*bytes.Buffer)

	h.Write(buffer.Bytes())
	sm := h.Sum(nil)
	bytes.Equal(sm, tsInfo.MessageImprint.HashedMessage)

	return model.SignatureValidationResult{IsSigned: true, IsVerified: bytes.Equal(sm, tsInfo.MessageImprint.HashedMessage)}, nil
>>>>>>> add timestamp signature handler
}

// Sign sets the Contents fields for the PdfSignature.
func (a *docTimeStamp) Sign(sig *model.PdfSignature, digest model.Hasher) error {
<<<<<<< HEAD
<<<<<<< HEAD
	buffer := digest.(*bytes.Buffer)
	h := a.hashAlgorithm.New()

	if _, err := io.Copy(h, buffer); err != nil {
		return err
	}

	s := h.Sum(nil)
	r := timestamp.Request{
		HashAlgorithm:   a.hashAlgorithm,
		HashedMessage:   s,
		Certificates:    true,
		Extensions:      nil,
		ExtraExtensions: nil,
	}
	data, err := r.Marshal()
	if err != nil {
		return err
	}

	resp, err := http.Post(a.timestampServerURL, "application/timestamp-query", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code not ok (got %d)", resp.StatusCode)
	}

	var ci struct {
		Version asn1.RawValue
		Content asn1.RawValue
	}

	_, err = asn1.Unmarshal(body, &ci)
=======
	var data []byte
	var err error
=======
	if a.emptySignatureLen <= 0 {
		sig.Contents = core.MakeHexString(string(make([]byte, 8192)))
		return nil
	}
>>>>>>> add timestamp signature handler

	buffer := digest.(*bytes.Buffer)

	h := crypto.SHA512.New()
	io.Copy(h, buffer)
	//h.Write([]byte("test message"))
	s := h.Sum(nil)

	r := timestamp.Request{
		HashAlgorithm:   crypto.SHA512,
		HashedMessage:   s,
		Certificates:    true,
		Extensions:      nil,
		ExtraExtensions: nil,
	}
	data, err := r.Marshal()
	if err != nil {
		return err
	}

	resp, err := http.Post("https://freetsa.org/tsr", "application/timestamp-query", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code wiats 200 got %d", resp.StatusCode)
	}

	var ci struct {
		Version asn1.RawValue
		Content asn1.RawValue
	}

<<<<<<< HEAD
	data, err = asn1.Marshal(data)
>>>>>>> sync commit
=======
	_, err = asn1.Unmarshal(body, &ci)
>>>>>>> add timestamp signature handler
	if err != nil {
		return err
	}

<<<<<<< HEAD
<<<<<<< HEAD
	sig.Contents = core.MakeHexString(string(ci.Content.FullBytes))
=======
	sig.Contents = core.MakeHexString(string(data))
>>>>>>> sync commit
=======
	sig.Contents = core.MakeHexString(string(ci.Content.FullBytes))
>>>>>>> add timestamp signature handler
	return nil
}

// IsApplicable returns true if the signature handler is applicable for the PdfSignature.
func (a *docTimeStamp) IsApplicable(sig *model.PdfSignature) bool {
	if sig == nil || sig.Filter == nil || sig.SubFilter == nil {
		return false
	}
	return (*sig.Filter == "Adobe.PPKMS" || *sig.Filter == "Adobe.PPKLite") && *sig.SubFilter == "ETSI.RFC3161"
}
