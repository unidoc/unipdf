/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"time"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
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
func NewPdfSignature(handler SignatureHandler) *PdfSignature {
	sig := &PdfSignature{
		Type:    core.MakeName("Sig"),
		Handler: handler,
	}

	dict := &pdfSignDictionary{
		PdfObjectDictionary: core.MakeDict(),
		handler:             &handler,
		signature:           sig,
	}

	sig.container = core.MakeIndirectObject(dict)
	return sig
}

// GetContainingPdfObject implements interface PdfModel.
func (sig *PdfSignature) GetContainingPdfObject() core.PdfObject {
	return sig.container
}

// SetName sets the `Name` field of the signature.
func (sig *PdfSignature) SetName(name string) {
	sig.Name = core.MakeString(name)
}

// SetDate sets the `M` field of the signature.
func (sig *PdfSignature) SetDate(date time.Time, format string) {
	if format == "" {
		format = "D:20060102150405-07'00'"
	}

	sig.M = core.MakeString(date.Format(format))
}

// SetReason sets the `Reason` field of the signature.
func (sig *PdfSignature) SetReason(reason string) {
	sig.Reason = core.MakeString(reason)
}

// SetLocation sets the `Location` field of the signature.
func (sig *PdfSignature) SetLocation(location string) {
	sig.Location = core.MakeString(location)
}

// Initialize initializes the PdfSignature.
func (sig *PdfSignature) Initialize() error {
	if sig.Handler == nil {
		return errors.New("signature handler cannot be nil")
	}

	return sig.Handler.InitSignature(sig)
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

	dict.SetIfNotNil("Type", sig.Type)
	dict.SetIfNotNil("Filter", sig.Filter)
	dict.SetIfNotNil("SubFilter", sig.SubFilter)
	dict.SetIfNotNil("ByteRange", sig.ByteRange)
	dict.SetIfNotNil("Contents", sig.Contents)
	dict.SetIfNotNil("Cert", sig.Cert)
	dict.SetIfNotNil("Name", sig.Name)
	dict.SetIfNotNil("Reason", sig.Reason)
	dict.SetIfNotNil("M", sig.M)
	dict.SetIfNotNil("Reference", sig.Reference)
	dict.SetIfNotNil("Changes", sig.Changes)
	dict.SetIfNotNil("ContactInfo", sig.ContactInfo)

	// NOTE: ByteRange and Contents are updated dynamically (appender).
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

	sig.Type, _ = core.GetName(dict.Get("Type"))

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
