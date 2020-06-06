/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package sighandler

import (
	"bytes"
	"crypto"
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
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// docTimeStamp DocTimeStamp signature handler.
type docTimeStamp struct {
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
}

// InitSignature initialises the PdfSignature.
func (a *docTimeStamp) InitSignature(sig *model.PdfSignature) error {
	handler := *a
	sig.Handler = &handler
	sig.Filter = core.MakeName("Adobe.PPKLite")
	sig.SubFilter = core.MakeName("ETSI.RFC3161")
	sig.Reference = nil
	digest, err := a.NewDigest(sig)
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

type timestampInfo struct {
	Version        int
	Policy         asn1.RawValue
	MessageImprint struct {
		HashAlgorithm pkix.AlgorithmIdentifier
		HashedMessage []byte
	}
	SerialNumber    asn1.RawValue
	GeneralizedTime asn1.RawValue
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

// parseGeneralizedTime parses the GeneralizedTime from the given byte slice
// and returns the resulting time.
func parseGeneralizedTime(bytes []byte) (ret time.Time, err error) {
	const formatStr = "20060102150405.000Z0700"
	const formatStr2 = "20060102150405Z0700"
	s := string(bytes)
	ret, err = time.Parse(formatStr, s)

	//if ret, err = time.Parse(formatStr, s); err != nil {
	//	return
	//}

	if err != nil {
		if ret, err = time.Parse(formatStr2, s); err != nil {
			return
		}
	}
	return
}

// Validate validates PdfSignature.
func (a *docTimeStamp) Validate(sig *model.PdfSignature, digest model.Hasher) (model.SignatureValidationResult, error) {
	signed := sig.Contents.Bytes()
	p7, err := pkcs7.Parse(signed)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}

	if err = p7.Verify(); err != nil {
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
	h := hAlg.New()
	buffer := digest.(*bytes.Buffer)

	h.Write(buffer.Bytes())
	sm := h.Sum(nil)

	generalizedTime, err := parseGeneralizedTime(tsInfo.GeneralizedTime.Bytes)
	if err != nil {
		return model.SignatureValidationResult{}, err
	}

	res := model.SignatureValidationResult{
		IsSigned:        true,
		IsVerified:      bytes.Equal(sm, tsInfo.MessageImprint.HashedMessage),
		GeneralizedTime: generalizedTime,
	}
	return res, nil
}

// Sign sets the Contents fields for the PdfSignature.
func (a *docTimeStamp) Sign(sig *model.PdfSignature, digest model.Hasher) error {
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
	if err != nil {
		return err
	}

	sig.Contents = core.MakeHexString(string(ci.Content.FullBytes))
	return nil
}

// IsApplicable returns true if the signature handler is applicable for the PdfSignature.
func (a *docTimeStamp) IsApplicable(sig *model.PdfSignature) bool {
	if sig == nil || sig.Filter == nil || sig.SubFilter == nil {
		return false
	}
	return (*sig.Filter == "Adobe.PPKMS" || *sig.Filter == "Adobe.PPKLite") && *sig.SubFilter == "ETSI.RFC3161"
}
