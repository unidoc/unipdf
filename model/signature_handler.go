/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// Hasher is the interface that wraps the basic Write method.
type Hasher interface {
	Write(p []byte) (n int, err error)
}

// SignatureHandler interface defines the common functionality for PDF signature handlers, which
// need to be capable of validating digital signatures and signing PDF documents.
type SignatureHandler interface {
	IsApplicable(sig *PdfSignature) bool
	Validate(sig *PdfSignature, digest Hasher) (SignatureValidationResult, error)
	// InitSignature sets the PdfSignature parameters.
	InitSignature(*PdfSignature) error
	NewDigest(sig *PdfSignature) (Hasher, error)
	Sign(sig *PdfSignature, digest Hasher) error
}

// SignatureValidationResult defines the response from the signature validation handler.
type SignatureValidationResult struct {
	// List of errors when validating the signature.
	Errors      []string
	IsSigned    bool
	IsVerified  bool
	IsTrusted   bool
	Fields      []*PdfField
	Name        string
	Date        PdfDate
	Reason      string
	Location    string
	ContactInfo string

	// TODO(gunnsth): Add more fields such as ability to access the certificate information (name, CN, etc).
	// TODO: Also add flags to indicate whether the signature covers the entire file, or the entire portion of
	// a revision (if incremental updates used).

	// GeneralizedTime is the time at which the time-stamp token has been created by the TSA (RFC 3161).
	GeneralizedTime time.Time
}

func (v SignatureValidationResult) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Name: %s\n", v.Name))
	if v.Date.year > 0 {
		buf.WriteString(fmt.Sprintf("Date: %s\n", v.Date.ToGoTime().String()))
	} else {
		buf.WriteString("Date not specified\n")
	}
	if len(v.Reason) > 0 {
		buf.WriteString(fmt.Sprintf("Reason: %s\n", v.Reason))
	} else {
		buf.WriteString("No reason specified\n")
	}
	if len(v.Location) > 0 {
		buf.WriteString(fmt.Sprintf("Location: %s\n", v.Location))
	} else {
		buf.WriteString("Location not specified\n")
	}
	if len(v.ContactInfo) > 0 {
		buf.WriteString(fmt.Sprintf("Contact Info: %s\n", v.ContactInfo))
	} else {
		buf.WriteString("Contact info not specified\n")
	}
	buf.WriteString(fmt.Sprintf("Fields: %d\n", len(v.Fields)))
	if v.IsSigned {
		buf.WriteString("Signed: Document is signed\n")
	} else {
		buf.WriteString("Signed: Not signed\n")
	}
	if v.IsVerified {
		buf.WriteString("Signature validation: Is valid\n")
	} else {
		buf.WriteString("Signature validation: Is invalid\n")
	}
	if v.IsTrusted {
		buf.WriteString("Trusted: Certificate is trusted\n")
	} else {
		buf.WriteString("Trusted: Untrusted certificate\n")
	}
	if !v.GeneralizedTime.IsZero() {
		buf.WriteString(fmt.Sprintf("GeneralizedTime: %s\n", v.GeneralizedTime.String()))
	}
	return buf.String()
}

// ValidateSignatures validates digital signatures in the document.
func (r *PdfReader) ValidateSignatures(handlers []SignatureHandler) ([]SignatureValidationResult, error) {
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
	for _, f := range r.AcroForm.AllFields() {
		if f.V == nil {
			continue
		}
		if d, found := core.GetDict(f.V); found {
			if name, ok := core.GetNameVal(d.Get("Type")); ok && name == "Sig" {
				ind, found := core.GetIndirect(f.V)
				if !found {
					common.Log.Debug("ERROR: Signature container is nil")
					return nil, ErrTypeCheck
				}

				sig, err := r.newPdfSignatureFromIndirect(ind)
				if err != nil {
					return nil, err
				}

				// Search for an appropriate handler.
				var sigHandler SignatureHandler
				for _, handler := range handlers {
					if handler.IsApplicable(sig) {
						sigHandler = handler
						break
					}
				}

				pairs = append(pairs, &sigFieldPair{
					sig:     sig,
					field:   f,
					handler: sigHandler,
				})
			}
		}
	}

	var results []SignatureValidationResult
	for _, pair := range pairs {
		defaultResult := SignatureValidationResult{
			IsSigned: true,
			Fields:   []*PdfField{pair.field},
		}
		if pair.handler == nil {
			defaultResult.Errors = append(defaultResult.Errors, "handler not set")
			results = append(results, defaultResult)
			continue
		}
		digest, err := pair.handler.NewDigest(pair.sig)
		if err != nil {
			defaultResult.Errors = append(defaultResult.Errors, "digest error", err.Error())
			results = append(results, defaultResult)
			continue
		}
		byteRange := pair.sig.ByteRange
		if byteRange == nil {
			defaultResult.Errors = append(defaultResult.Errors, "ByteRange not set")
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

		result.Name = pair.sig.Name.Decoded()
		result.Reason = pair.sig.Reason.Decoded()
		if pair.sig.M != nil {
			sigDate, err := NewPdfDate(pair.sig.M.String())
			if err != nil {
				result.Errors = append(result.Errors, err.Error())
				continue
			}
			result.Date = sigDate
		}
		result.ContactInfo = pair.sig.ContactInfo.Decoded()
		result.Location = pair.sig.Location.Decoded()

		result.Fields = defaultResult.Fields
		results = append(results, result)
	}
	return results, nil
}
