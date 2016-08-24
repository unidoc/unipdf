/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

//
// Higher level manipulation of forms (AcroForm).
//

package pdf

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/unidoc/unidoc/common"
)

type PdfAcroForm struct {
	Fields          []*PdfField
	NeedAppearances PdfObject
	SigFlags        PdfObject
	CO              PdfObject
	DR              PdfObject
	DA              PdfObject
	Q               PdfObject
	XFA             PdfObject
}

type PdfWidgetAnnotation {
	Subtype PdfObject
	H PdfObject
	MK PdfObject
	A PdfObject
	AA PdfObject
	BS PdfObject
	Parent *PdfIndirectObject // Max 1 parent; Gets tricky for both form and annotation refs?  Seems to usually refer to the page one.
}

type PdfField struct {
	Parent *PdfField
	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.
	Kids []*PdfField
	FT   *PdfObjectString // field type
	T    PdfObject
	TU   PdfObject
	TM   PdfObject
	Ff   PdfObject // field flag
	V    PdfObject //value
	DV   PdfObject
	AA   PdfObject
	// Widget annotation can be merged in.

	// Variable text fields.
	DA PdfObject
	Q  PdfObject
	DS PdfObject
	RV PdfObject
	// Text field
	MaxLen PdfObject // inheritable
	// Choice fields.
	Opt PdfObject
	TI  PdfObject
	I   PdfObject
	// Signature fields (Table 232).
	Lock PdfObject
	SV   PdfObject
	// Signature field lock dict (Table 233).
	Type   PdfObject // SigFieldLock
	Action *PdfObjectName
	Fields PdfObject
	// Signature field seed value dictionary (Table 234)
	//Type //SV
	Ff
	Filter
	SubFilter
	DigestMethod
	V
	Cert
	Reasons
	MDP
	TimeStamp
	LegalAttestation
	AddRevInfo
	// Certificate seed value dictionary (Table 235).
	// Type //SVCert
	// Ff
	Subject
	SubjectDN
	KeyUsage
	Issuer
	OID
	URL
	URLType
	// S
}
