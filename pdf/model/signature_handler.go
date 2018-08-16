/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
package model

import (
	"bytes"
	"crypto/sha1"
	"errors"

	"github.com/unidoc/unidoc/pdf/internal/crypto/asn1"
	"github.com/unidoc/unidoc/pdf/internal/crypto/pkcs7"
	"github.com/unidoc/unidoc/pdf/internal/crypto/x509"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// SignatureHandler interface defines the common functionality for PDF signature handlers, which
// need to be capable of validating digital signatures and signing PDF documents.
type SignatureHandler interface {
	Validate(reader *PdfReader) ([]SignatureValidationResult, error)
	Sign(privateKey []byte, writer *PdfWriter) error
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

// DefaultSignatureHandler implements the default signature handling in UniDoc.
type DefaultSignatureHandler struct{}

// Sign signs the output PDF document using PKCS#7 with ETSI.CAdES.detached subfilter.
// TODO: Need to 1. Define signature fields, 2. Apply the signature to those fields... 3. Handle writing.
// 1. Create the signature.
// signature := NewPdfSignature()
// sigfield := forms.NewDigitalSignatureField(signature)
// sigfield.SetText("Blabla")
// sigfield.SetImage(fromFile("x.jpg"))
// X. set position?
// page.AddFormField(sigfield)
func (sh DefaultSignatureHandler) Sign(privateKey []byte, writer *PdfWriter) error {
	// Form creation...
	// annotation?  usually apply to a read document.
	// writer.AddSignature
	return errors.New("not implemented")
}

// Validate validates the PDF signatures in a PDF document loaded in `reader`.
func (sh DefaultSignatureHandler) Validate(reader *PdfReader) ([]SignatureValidationResult, error) {
	results := []SignatureValidationResult{}

	acroForm := reader.AcroForm
	if acroForm == nil {
		return results, nil
	}

	sigfields := acroForm.signatureFields()

	sigToFieldMap := map[*PdfSignature][]*PdfField{}
	for _, sigfield := range sigfields {
		sig := sigfield.V
		if sig == nil {
			continue
		}

		fields, has := sigToFieldMap[sig]
		if has {
			fields = append(fields, sigfield.PdfField)
		} else {
			fields = []*PdfField{sigfield.PdfField}
		}
		sigToFieldMap[sig] = fields
	}

	for sig, fields := range sigToFieldMap {
		if sig.Filter.String() != "Adobe.PPKLite" && sig.Filter.String() != "Adobe.PPKMS" {
			common.Log.Debug("ERROR: Unsupported filter: %v\n", sig.Filter)
			return nil, errors.New("unsupported filter")
		}

		// Type.
		if sig.Type.String() != "Sig" && sig.Type.String() != "DocTimeStamp" {
			common.Log.Debug("WARN: Unsupported signature field type - skipping over\n")
			continue
		}

		// ByteRange
		if sig.ByteRange == nil {
			common.Log.Debug("ERROR: ByteRange not specified")
			return nil, ErrRequiredAttributeMissing
		}
		byteRange, err := sig.ByteRange.ToInt64Slice()
		if err != nil {
			common.Log.Debug("ERROR: %v", err)
			return nil, err
		}
		if len(byteRange) == 0 {
			return results, errors.New("Byte range empty")
		}
		if len(byteRange)%2 != 0 {
			return results, errors.New("Byte range not a multiple of 2")
		}

		// Contents required.
		if sig.Contents == nil {
			common.Log.Debug("ERROR: Contents invalid or missing")
			return nil, ErrInvalidAttribute
		}

		// SubFilter.
		if sig.SubFilter == nil {
			common.Log.Debug("ERROR: SubFilter missing or invalid")
			return results, ErrInvalidAttribute
		}

		result := SignatureValidationResult{}
		verified := false
		trusted := false

		switch sig.SubFilter.String() {
		case "adbe.pkcs7.detached":
			fallthrough
		case "ETSI.CAdES.detached":
			signPackage, err := pkcs7.Parse([]byte(sig.Contents.Str()))
			if err != nil {
				return results, err
			}

			// TODO: Add hash calculation into unidoc. (cannot access file descriptor from outside).
			fileContent, err := getContentForByteRange(reader, byteRange)
			if err != nil {
				return results, err
			}

			// Set the content (as detached).
			signPackage.Content = fileContent

			err = signPackage.Verify()
			if err != nil {
				verified = false
			} else {
				verified = true
			}

			// Check trust of certificates.
			trusted = true
			for _, cert := range signPackage.Certificates {
				verifyOptions := x509.VerifyOptions{}
				_, err := cert.Verify(verifyOptions)
				if err != nil {
					trusted = false
					break
				}
			}

		case "adbe.pkcs7.sha1":
			signPackage, err := pkcs7.Parse([]byte(sig.Contents.Str()))
			if err != nil {
				return results, err
			}

			err = signPackage.Verify()
			if err != nil {
				verified = false
			} else {
				verified = true
			}

			// TODO: Not to require input path.
			fileContent, err := getContentForByteRange(reader, byteRange)
			if err != nil {
				return results, err
			}

			h := sha1.New()
			h.Write(fileContent)
			d := h.Sum(nil)

			if len(d) != len(signPackage.Content) {
				verified = false
			}
			for i := range d {
				if d[i] != signPackage.Content[i] {
					verified = false
					break
				}
			}

			// Check trust of certificates.
			trusted = true
			for _, cert := range signPackage.Certificates {
				verifyOptions := x509.VerifyOptions{}
				_, err := cert.Verify(verifyOptions)
				if err != nil {
					trusted = false
					break
				}
			}

		case "adbe.x509.rsa_sha1":
			certString, ok := sig.Cert.(*core.PdfObjectString)
			if !ok {
				return results, errors.New("Cert missing")
			}
			certs, err := x509.ParseCertificates([]byte(certString.Str()))
			if err != nil {
				return results, err
			}
			if len(certs) == 0 {
				return results, errors.New("No certificates")
			}
			// Use first certificate.
			cert := certs[0]

			// TODO: Add hash calculation into unidoc. (cannot access file descriptor from outside).
			fileContent, err := getContentForByteRange(reader, byteRange)
			if err != nil {
				return results, err
			}

			// ASN1 decode the signature contents.
			signature := []byte(sig.Contents.Str())
			var asn1Sig asn1.RawContent
			_, err = asn1.Unmarshal(signature, &asn1Sig)
			if err != nil {
				return results, err
			}

			err = cert.CheckSignature(cert.SignatureAlgorithm, fileContent, asn1Sig)
			if err == nil {
				verified = true
			} else {
				verified = false
			}

			// Check trust of certificate.
			common.Log.Trace("verifying...")
			verifyOptions := x509.VerifyOptions{}
			_, err = cert.Verify(verifyOptions)
			if err == nil {
				trusted = true
			}
			common.Log.Trace("trusted: %v", trusted)

		case "ETSI.RFC3161":
			signPackage, err := pkcs7.Parse([]byte(sig.Contents.Str()))
			if err != nil {
				return results, err
			}

			err = signPackage.Verify()
			if err == nil {
				verified = true
			} else {
				verified = false
			}

			// TODO: Chain validation.
			// Check trust of certificates.
			trusted = true
			for _, cert := range signPackage.Certificates {
				verifyOptions := x509.VerifyOptions{}
				_, err := cert.Verify(verifyOptions)
				if err != nil {
					trusted = false
					break
				}
			}
		default:
			return results, errors.New("unknown subfilter")
		}

		result.IsVerified = verified
		result.IsTrusted = trusted
		result.Fields = fields

		// Name
		if sig.Name != nil {
			result.Name = sig.Name.String()
		}

		// Signature date
		if sig.M != nil {
			signerDateStr, has := core.GetStringVal(sig.M)
			if has {
				signDate, err := NewPdfDate(signerDateStr)
				if err != nil {
					return results, err
				}
				result.Date = signDate
			}
		}

		if sig.Reason != nil {
			result.Reason = sig.Reason.String()
		}

		if sig.Location != nil {
			result.Location = sig.Location.String()
		}

		if sig.ContactInfo != nil {
			result.ContactInfo = sig.ContactInfo.String()
		}

		results = append(results, result)
	}

	return results, nil
}

// getContentForByteRange returns the content in the specified byte ranges of input PDF loaded in `reader`.
// Used for signature verification purposes.
func getContentForByteRange(reader *PdfReader, byteRange []int64) ([]byte, error) {
	buf := bytes.Buffer{}

	for i := 0; i < len(byteRange); i += 2 {
		fromOffset := byteRange[i]
		byteLen := byteRange[i+1]
		if byteLen < 0 {
			return nil, errors.New("byte length cannot be negative")
		}

		bb, err := reader.parser.ReadBytesAt(fromOffset, byteLen)
		if err != nil {
			return nil, err
		}

		buf.Write(bb)
	}

	return buf.Bytes(), nil
}
