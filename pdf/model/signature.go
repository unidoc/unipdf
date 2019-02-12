/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

var _ core.PdfObject = &pdfSignDictionary{}

// pdfSignDictionary is used as a wrapper around PdfSignature for digital checksum calculation
// and population of /Contents and /ByteRange.
// Implements interface core.PdfObject.
type pdfSignDictionary struct {
	*core.PdfObjectDictionary
	handler              *SignatureHandler
	signature            *PdfSignature
	fileOffset           int64
	contentsOffsetStart  int
	contentsOffsetEnd    int
	byteRangeOffsetStart int
	byteRangeOffsetEnd   int
}

// GetSubFilter returns SubFilter value or empty string.
func (d *pdfSignDictionary) GetSubFilter() string {
	obj := d.Get("SubFilter")
	if obj == nil {
		return ""
	}
	if sf, found := core.GetNameVal(obj); found {
		return sf
	}
	return ""
}

// WriteString outputs the object as it is to be written to file.
func (d *pdfSignDictionary) WriteString() string {
	d.contentsOffsetStart = 0
	d.contentsOffsetEnd = 0
	d.byteRangeOffsetStart = 0
	d.byteRangeOffsetEnd = 0
	out := bytes.NewBuffer(nil)
	out.WriteString("<<")
	for _, k := range d.Keys() {
		v := d.Get(k)
		switch k {
		case "ByteRange":
			out.WriteString(k.WriteString())
			out.WriteString(" ")
			d.byteRangeOffsetStart = out.Len()
			out.WriteString(v.WriteString())
			out.WriteString(" ")
			d.byteRangeOffsetEnd = out.Len() - 1
		case "Contents":
			out.WriteString(k.WriteString())
			out.WriteString(" ")
			d.contentsOffsetStart = out.Len()
			out.WriteString(v.WriteString())
			out.WriteString(" ")
			d.contentsOffsetEnd = out.Len() - 1
		default:
			out.WriteString(k.WriteString())
			out.WriteString(" ")
			out.WriteString(v.WriteString())
		}
	}
	out.WriteString(">>")
	return out.String()
}

// PdfSignature represents a PDF signature dictionary and is used for signing via form signature fields.
// (Section 12.8, Table 252 - Entries in a signature dictionary p. 475 in PDF32000_2008).
type PdfSignature struct {
	Handler   SignatureHandler
	container *core.PdfIndirectObject
	// Type: Sig/DocTimeStamp
	Type         *core.PdfObjectName
	Filter       *core.PdfObjectName
	SubFilter    *core.PdfObjectName
	Contents     *core.PdfObjectString
	Cert         core.PdfObject
	ByteRange    *core.PdfObjectArray
	Reference    *core.PdfObjectArray
	Changes      *core.PdfObjectArray
	Name         *core.PdfObjectString
	M            *core.PdfObjectString
	Location     *core.PdfObjectString
	Reason       *core.PdfObjectString
	ContactInfo  *core.PdfObjectString
	R            *core.PdfObjectInteger
	V            *core.PdfObjectInteger
	PropBuild    *core.PdfObjectDictionary
	PropAuthTime *core.PdfObjectInteger
	PropAuthType *core.PdfObjectName
}

// NewPdfSignature creates a new PdfSignature object.
func NewPdfSignature() *PdfSignature {
	sig := &PdfSignature{}
	sigDict := &pdfSignDictionary{
		PdfObjectDictionary: core.MakeDict(),
		handler:             &sig.Handler,
		signature:           sig,
	}
	sig.container = core.MakeIndirectObject(sigDict)
	return sig
}

// PdfSignatureReference represents a signature reference dictionary.
// (Table 253 - p. 477 in PDF32000_2008).
type PdfSignatureReference struct {
	// Type: SigRef
	TransformMethod *core.PdfObjectName
	TransformParams *core.PdfObjectDictionary
	Data            core.PdfObject
	DigestMethod    *core.PdfObjectName
}

// GetContainingPdfObject implements interface PdfModel.
func (sig *PdfSignature) GetContainingPdfObject() core.PdfObject {
	return sig.container
}

// ToPdfObject implements interface PdfModel.
func (sig *PdfSignature) ToPdfObject() core.PdfObject {
	container := sig.container

	var dict *core.PdfObjectDictionary
	if sigDict, ok := container.PdfObject.(*pdfSignDictionary); ok {
		dict = sigDict.PdfObjectDictionary
	} else {
		dict = container.PdfObject.(*core.PdfObjectDictionary)
	}

	dict.Set("Type", sig.Type)
	dict.SetIfNotNil("Filter", sig.Filter)
	dict.SetIfNotNil("SubFilter", sig.SubFilter)
	dict.SetIfNotNil("Contents", sig.Contents)
	dict.SetIfNotNil("Cert", sig.Cert)
	dict.SetIfNotNil("ByteRange", sig.ByteRange)
	dict.SetIfNotNil("Reference", sig.Reference)
	dict.SetIfNotNil("Changes", sig.Changes)
	dict.SetIfNotNil("Name", sig.Name)
	dict.SetIfNotNil("M", sig.M)
	dict.SetIfNotNil("Reason", sig.Reason)
	dict.SetIfNotNil("ContactInfo", sig.ContactInfo)
	dict.SetIfNotNil("ByteRange", sig.ByteRange)
	dict.SetIfNotNil("Contents", sig.Contents)

	// NOTE: ByteRange and Contents need to be updated dynamically.
	// TODO: Currently dynamic update is only in the appender, need to support
	// in the PdfWriter too for the initial signature on document creation.
	return container
}

// newPdfSignatureFromIndirect loads a PdfSignature from an indirect object containing the signature dictionary.
func (r *PdfReader) newPdfSignatureFromIndirect(container *core.PdfIndirectObject) (*PdfSignature, error) {
	dict, ok := container.PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Signature container not containing a dictionary")
		return nil, ErrTypeCheck
	}

	// If PdfSignature already processed and cached, return the cached entry.
	if sig, cached := r.modelManager.GetModelFromPrimitive(container).(*PdfSignature); cached {
		return sig, nil
	}

	sig := &PdfSignature{}
	sig.container = container

	sig.Type, ok = core.GetName(dict.Get("Type"))
	if !ok {
		common.Log.Error("ERROR: Signature Type attribute invalid or missing")
		return nil, ErrInvalidAttribute
	}

	sig.Filter, ok = core.GetName(dict.Get("Filter"))
	if !ok {
		common.Log.Error("ERROR: Signature Filter attribute invalid or missing")
		return nil, ErrInvalidAttribute
	}

	sig.SubFilter, _ = core.GetName(dict.Get("SubFilter"))

	sig.Contents, ok = core.GetString(dict.Get("Contents"))
	if !ok {
		common.Log.Error("ERROR: Signature contents missing")
		return nil, ErrInvalidAttribute
	}

	sig.Cert = dict.Get("Cert")
	sig.ByteRange, _ = core.GetArray(dict.Get("ByteRange"))
	sig.Reference, _ = core.GetArray(dict.Get("Reference"))
	sig.Changes, _ = core.GetArray(dict.Get("Changes"))
	sig.Name, _ = core.GetString(dict.Get("Name"))
	sig.M, _ = core.GetString(dict.Get("M"))
	sig.Location, _ = core.GetString(dict.Get("Location"))
	sig.Reason, _ = core.GetString(dict.Get("Reason"))
	sig.ContactInfo, _ = core.GetString(dict.Get("ContactInfo"))
	sig.R, _ = core.GetInt(dict.Get("R"))
	sig.V, _ = core.GetInt(dict.Get("V"))
	sig.PropBuild, _ = core.GetDict(dict.Get("Prop_Build"))
	sig.PropAuthTime, _ = core.GetInt(dict.Get("Prop_AuthTime"))
	sig.PropAuthType, _ = core.GetName(dict.Get("Prop_AuthType"))

	return sig, nil
}

// PdfSignatureField represents a form field that contains a digital signature.
// (12.7.4.5 - Signature Fields p. 454 in PDF32000_2008).
//
// The signature form field serves two primary purposes. 1. Define the form field that will provide the
// visual signing properties for display but may also hold information needed when the actual signing
// takes place such as signature method. This carries information from the author of the document to the
// software that later does signing.
//
// Filling in (signing) the signature field entails updating at least the V entry and usually the AP entry of the
// associated widget annotation. (Exporting a signature field exports the T, V, AP entries)
//
// The annotation rectangle (Rect) in such a dictionary shall give the position of the field on its page. Signature
// fields that are not intended to be visible shall have an annotation rectangle that has zero height and width. PDF
// processors shall treat such signatures as not visible. PDF processors shall also treat signatures as not
// visible if either the Hidden bit or the NoView bit of the F entry is true
//
// The location of a signature within a document can have a bearing on its legal meaning. For this reason,
// signature fields shall never refer to more than one annotation.
type PdfSignatureField struct {
	container *core.PdfIndirectObject

	V    *PdfSignature
	Lock *core.PdfIndirectObject
	SV   *core.PdfIndirectObject
	Kids *core.PdfObjectArray
}

// NewPdfSignatureField prepares a PdfSignatureField from a PdfSignature.
func NewPdfSignatureField(signature *PdfSignature) *PdfSignatureField {
	return &PdfSignatureField{
		V:         signature,
		container: core.MakeIndirectObject(core.MakeDict()),
	}
}

// ToPdfObject implements interface PdfModel.
func (sf *PdfSignatureField) ToPdfObject() core.PdfObject {
	container := sf.container
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	// Set fields.
	if sf.V != nil {
		dict.Set("V", sf.V.ToPdfObject())
	}

	dict.Set("FT", core.MakeName("Sig"))
	dict.SetIfNotNil("Lock", sf.Lock)
	dict.SetIfNotNil("SV", sf.SV)
	dict.SetIfNotNil("Kids", sf.Kids)
	// Other standard fields...

	return container
}

// PdfSignatureFieldLock represents signature field lock dictionary.
// (Table 233 - p. 455 in PDF32000_2008).
type PdfSignatureFieldLock struct {
	Type   core.PdfObject
	Action *core.PdfObjectName
	Fields *core.PdfObjectArray
	P      *core.PdfObjectInteger
}

// PdfSignatureFieldSeed represents signature field seed value dictionary.
// (Table 234 - p. 455 in PDF32000_2008).
type PdfSignatureFieldSeed struct {
	container *core.PdfIndirectObject

	Ff               *core.PdfObjectInteger
	Filter           *core.PdfObjectName
	SubFilter        *core.PdfObjectArray
	DigestMethod     *core.PdfObjectArray
	V                *core.PdfObjectInteger
	Cert             core.PdfObject
	Reasons          *core.PdfObjectArray
	MDP              *core.PdfObjectDictionary
	TimeStamp        *core.PdfObjectDictionary
	LegalAttestation *core.PdfObjectArray
	AddRevInfo       *core.PdfObjectBool
	LockDocument     *core.PdfObjectName
	AppearanceFilter *core.PdfObjectString
}

func (pss *PdfSignatureFieldSeed) ToPdfObject() core.PdfObject {
	container := pss.container
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	dict.SetIfNotNil("Ff", pss.Ff)
	dict.SetIfNotNil("Filter", pss.Filter)
	dict.SetIfNotNil("SubFilter", pss.SubFilter)
	dict.SetIfNotNil("DigestMethod", pss.DigestMethod)
	dict.SetIfNotNil("V", pss.V)
	dict.SetIfNotNil("Cert", pss.Cert)
	dict.SetIfNotNil("Reasons", pss.Reasons)
	dict.SetIfNotNil("MDP", pss.MDP)
	dict.SetIfNotNil("TimeStamp", pss.TimeStamp)
	dict.SetIfNotNil("LegalAttestation", pss.LegalAttestation)
	dict.SetIfNotNil("AddRevInfo", pss.AddRevInfo)
	dict.SetIfNotNil("LockDocument", pss.LockDocument)
	dict.SetIfNotNil("AppearanceFilter", pss.AppearanceFilter)

	return container
}

// PdfCertificateSeed represents certificate seed value dictionary.
// (Table 235 - p. 457 in PDF32000_2008).
type PdfCertificateSeed struct {
	container *core.PdfIndirectObject
	// Type
	Ff                            *core.PdfObjectInteger
	Subject                       *core.PdfObjectArray
	SignaturePolicyOID            *core.PdfObjectString
	SignaturePolicyHashValue      *core.PdfObjectString
	SignaturePolicyHashAlgorithm  *core.PdfObjectName
	SignaturePolicyCommitmentType *core.PdfObjectArray
	SubjectDN                     *core.PdfObjectArray
	KeyUsage                      *core.PdfObjectArray
	Issuer                        *core.PdfObjectArray
	OID                           *core.PdfObjectArray
	URL                           *core.PdfObjectString
	URLType                       *core.PdfObjectName
}

func (pcs *PdfCertificateSeed) ToPdfObject() core.PdfObject {
	container := pcs.container
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	dict.SetIfNotNil("Ff", pcs.Ff)
	dict.SetIfNotNil("Subject", pcs.Subject)
	dict.SetIfNotNil("SignaturePolicyOID", pcs.SignaturePolicyOID)
	dict.SetIfNotNil("SignaturePolicyHashValue", pcs.SignaturePolicyHashValue)
	dict.SetIfNotNil("SignaturePolicyHashAlgorithm", pcs.SignaturePolicyHashAlgorithm)
	dict.SetIfNotNil("SignaturePolicyCommitmentType", pcs.SignaturePolicyCommitmentType)
	dict.SetIfNotNil("SubjectDN", pcs.SubjectDN)
	dict.SetIfNotNil("KeyUsage", pcs.KeyUsage)
	dict.SetIfNotNil("Issuer", pcs.Issuer)
	dict.SetIfNotNil("OID", pcs.OID)
	dict.SetIfNotNil("URL", pcs.URL)
	dict.SetIfNotNil("URLType", pcs.URLType)

	return container
}
