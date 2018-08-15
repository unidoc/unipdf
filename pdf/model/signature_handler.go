/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
package model

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"bytes"
	"os"
	"io"
)

// SignatureHandler interface defines the common functionality for PDF signature handlers, which
// need to be capable of validating digital signatures and signing PDF documents.
type SignatureHandler interface {
	Validate(reader *PdfReader, inputPath string) ([]SignatureValidationResult, error)
	Sign(privateKey []byte, writer *PdfWriter) error
}

// SignatureValidationResult defines the response from the signature validation handler.
type SignatureValidationResult struct {
	IsSigned    bool
	IsVerified  bool
	IsTrusted   bool
	Fields      []PdfField
	Name        string
	Date        PdfDate
	Reason      string
	Location    string
	ContactInfo string
}

var (
	currentSignatureHandler SignatureHandler = DefaultSignatureHandler{}
)

// SetSignatureHandler sets the current signature handler.
// Allows changing the default handler to a custom one.
func SetSignatureHandler(sighandler SignatureHandler) {
	currentSignatureHandler = sighandler
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
func (sh DefaultSignatureHandler) Sign(privateKey []byte, writer *pdf.PdfWriter) error {
	// Form creation...
	// annotation?  usually apply to a read document.
	// writer.AddSignature
	return errors.New("not implemented")
}

func (sh DefaultSignatureHandler) Validate(reader *PdfReader, inputPath string) ([]SignatureValidationResult, error) {
	results := []SignatureValidationResult{}

	acroForm := reader.AcroForm
	if acroForm == nil {
		fmt.Printf("No formdata present\n")
		return results, nil
	}

	sigdicts := acroForm.getSignatureDictionaries()
	for _, dict := range sigdicts {
		// Filter.
		filter, ok := core.GetName(dict.Get("Filter"))
		if !ok {
			return results, ErrTypeCheck
		}
		if *filter != "Adobe.PPKLite" && *filter != "Adobe.PPKMS" {
			common.Log.Debug("ERROR: Unsupported filter: %v\n", filter)
			//continue
			return nil, errors.New("unsupported filter")
		}

		// Type.
		ftype, ok := core.GetName(dict.Get("Type"))
		if !ok {
			fmt.Printf("Filter type not set\n")
			continue
		}
		if ftype.String() != "Sig" && ftype.String() != "DocTimeStamp" {
			fmt.Printf("Unsupported signature field type\n")
			continue
		}

		// ByteRange
		byteRangeArr, ok := core.GetArray(dict.Get("ByteRange")
		if !ok {
			common.Log.Debug("ERROR: ByteRange not specified")
			return nil, ErrRequiredAttributeMissing
		}
		byteRange, err := byteRangeArr.ToInt64Slice()
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

		// Contents.
		contents, ok := core.GetString(dict.Get("Contents"))
		if !ok {
			common.Log.Debug("ERROR: Contents invalid or missing")
			return nil, ErrInvalidAttribute
		}

		// SubFilter.
		subfilter, ok := core.GetName(dict.Get("SubFilter"))
		if !ok {
			common.Log.Debug("ERROR: SubFilter missing or invalid")
			return results, ErrInvalidAttribute
		}

		result := SignatureValidationResult{}
		verified := false
		trusted := false

		fmt.Println(*subfilter)
		switch *subfilter {
		case "adbe.pkcs7.detached":
			fallthrough
		case "ETSI.CAdES.detached":
			signPackage, err := pkcs7.Parse([]byte(contents.Str()))
			if err != nil {
				return results, err
			}

			// TODO: Add hash calculation into unidoc. (cannot access file descriptor from outside).
			fileContent, err := getContentForByteRange(inputPath, byteRange)
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
			signPackage, err := pkcs7.Parse([]byte(contents.Str()))
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
			fileContent, err := getContentForByteRange(inputPath, byteRange)
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
			fmt.Printf("Trusted? %v\n", trusted)

		case "adbe.x509.rsa_sha1":
			certString, ok := dict.Get("Cert").(*core.PdfObjectString)
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
			fileContent, err := getContentForByteRange(inputPath, byteRange)
			if err != nil {
				return results, err
			}

			// ASN1 decode the signature contents.
			signature := []byte(contents.Str())
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
			fmt.Printf("verifying...\n")
			verifyOptions := x509.VerifyOptions{}
			_, err = cert.Verify(verifyOptions)
			if err == nil {
				trusted = true
			}
			fmt.Printf("trusted: %v\n", trusted)

		case "ETSI.RFC3161":
			signPackage, err := pkcs7.Parse([]byte(contents.Str()))
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
			return results, errors.New("Unknown subfilter")
		}

		result.IsVerified = verified
		result.IsTrusted = trusted

		// Name
		signerName, has := core.GetName(dict.Get("Name"))
		if has {
			result.Name = signerName.String()
		}

		// Signature date
		signerDateStr, has := core.GetStringVal(dict.Get("M"))
		if has {
			signDate, err := pdf.NewPdfDate(signerDateStr)
			if err != nil {
				return results, err
			}
			result.Date = signDate
		}

		// Reason
		if reason, has := dict.Get("Reason").(*core.PdfObjectString); has {
			result.Reason = reason.Str()
		}

		// Location.
		if location, has := dict.Get("Location").(*core.PdfObjectString); has {
			result.Location = location.Str()
		}

		// ContactInfo.
		if contactInfo, has := dict.Get("ContactInfo").(*core.PdfObjectString); has {
			result.ContactInfo = contactInfo.Str()
		}

		results = append(results, result)
	}

	//fmt.Printf("%d Signature fields\n", numSigFields)
	fmt.Printf("Done\n")

	return results, nil
}


// getContentForByteRange returna the content in the specified byte range of input file specified by `inputPath`
// as a slice of bytes.
func getContentForByteRange(inputPath string, byteRange []int64) ([]byte, error) {
	buf := bytes.Buffer{}

	f, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	for i := 0; i < len(byteRange); i += 2 {
		fromOffset := byteRange[i]
		byteLen := byteRange[i+1]
		if byteLen < 0 {
			return nil, errors.New("byte length cannot be negative")
		}

		_, err := f.Seek(fromOffset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		bb := make([]byte, byteLen)
		_, err = io.ReadAtLeast(f, bb, int(byteLen))
		if err != nil {
			return nil, err
		}
		buf.Write(bb)
	}

	return buf.Bytes(), nil
}
