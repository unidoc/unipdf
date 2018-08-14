/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import "github.com/unidoc/unidoc/pdf/core"

// PdfSignature represents a PDF signature dictionary and is used for signing via form signature fields.
// (Section 12.8, Table 252 - Entries in a signature dictionary p. 475).
type PdfSignature struct {
	container *core.PdfIndirectObject
	// Type: Sig
	Filter        *core.PdfObjectName
	SubFilter     *core.PdfObjectName
	Contents      *core.PdfObjectString
	Cert          core.PdfObject
	ByteRange     *core.PdfObjectArray
	Reference     *core.PdfObjectArray
	Changes       *core.PdfObjectArray
	Name          *core.PdfObjectString
	M             *core.PdfObjectString
	Location      *core.PdfObjectString
	Reason        *core.PdfObjectString
	ContactInfo   *core.PdfObjectString
	R             *core.PdfObjectInteger
	V             *core.PdfObjectInteger
	Prop_Build    *core.PdfObjectDictionary
	Prop_AuthTime *core.PdfObjectInteger
	Prop_AuthType *core.PdfObjectName
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

// NewPdfSignature returns an initialized PdfSignature.
// TODO: Make more flexible?
func NewPdfSignature() *PdfSignature {
	sig := &PdfSignature{}
	sig.container = core.MakeIndirectObject(core.MakeDict())

	// FIXME/TODO: Replace with generic signature handler (provide default).
	sig.Filter = core.MakeName("Adobe.PPKLite")
	//sig.Type = core.MakeName("Sig")

	return sig
}

// ToPdfObject implements interface PdfModel.
func (sig *PdfSignature) ToPdfObject() core.PdfObject {
	container := sig.container
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	dict.Set("Type", core.MakeName("Sig"))

	if sig.Filter != nil {
		dict.Set("Filter", sig.Filter)
	}
	if sig.SubFilter != nil {
		dict.Set("SubFilter", sig.SubFilter)
	}
	if sig.Contents != nil {
		dict.Set("Contents", sig.Contents)
	}
	if sig.Cert != nil {
		dict.Set("Cert", sig.Cert)
	}
	if sig.ByteRange != nil {
		dict.Set("ByteRange", sig.ByteRange)
	}
	if sig.Reference != nil {
		dict.Set("Reference", sig.Reference)
	}
	if sig.Changes != nil {
		dict.Set("Changes", sig.Changes)
	}
	if sig.Name != nil {
		dict.Set("Name", sig.Name)
	}
	if sig.M != nil {
		dict.Set("M", sig.M)
	}
	if sig.Reason != nil {
		dict.Set("Reason", sig.Reason)
	}
	if sig.ContactInfo != nil {
		dict.Set("ContactInfo", sig.ContactInfo)
	}

	// XXX/FIXME: ByteRange and Contents need to be updated dynamically.
	return container
}

// PdfSignatureField represents a form field that contains a digital signature.
// (12.7.4.5 - Signature Fields p. 454 in PDF32000_2008.PDF).
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

	V *PdfSignature
	// Type: /Sig
	// V: *PdfSignature...
	Lock *core.PdfIndirectObject // Shall be an indirect reference.
	SV   *core.PdfIndirectObject // Shall be an indirect reference.
}

// NewPdfSignatureField prepares a PdfSignatureField from a PdfSignature.
func NewPdfSignatureField(signature *PdfSignature) *PdfSignatureField {
	sf := &PdfSignatureField{}
	sf.container = &core.PdfIndirectObject{}
	sf.container.PdfObject = core.MakeDict()

	sf.V = signature

	return sf
}

// ToPdfObject implements interface PdfModel.
func (sf *PdfSignatureField) ToPdfObject() core.PdfObject {
	container := sf.container
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	dict.Set("FT", core.MakeName("Sig"))

	if sf.V != nil {
		dict.Set("V", sf.V.ToPdfObject())
	}
	if sf.Lock != nil {
		dict.Set("Lock", sf.Lock)
	}
	if sf.SV != nil {
		dict.Set("SV", sf.SV)
	}
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
	// Type
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

// PdfCertificateSeed represents certificate seed value dictionary.
// (Table 235 - p. 457 in PDF32000_2008).
type PdfCertificateSeed struct {
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

// TODO: Add signature model.  PdfSignature...
// getSignatureDictionaries returns a slice of signature dictionaries from form fields.
func (form *PdfAcroForm) getSignatureDictionaries() []*core.PdfObjectDictionary {
	sigdicts := []*core.PdfObjectDictionary{}

	sigfields := form.getSignatureFields()
	for _, f := range sigfields {
		dict, has := core.GetDict(f.V)
		if has {
			sigdicts = append(sigdicts, dict)
		}

	}

	return sigdicts
}

// getSignatureFields returns a slice of signature fields.
func (form *PdfAcroForm) getSignatureFields() []*PdfField {
	sigfields := []*PdfField{}

	if form.Fields != nil {
		for _, f := range *form.Fields {
			if f.FT != nil && f.FT.String() == "Sig" {
				sigfields = append(sigfields, f)
			}
		}
	}

	return sigfields
}
