package model

import (
	"strings"

	"github.com/unidoc/unipdf/v3/core"
)

// DSS describes PDF DSS certificates info.
type DSSCerts struct {
	Certs []*core.PdfObjectStream
	OCSPs []*core.PdfObjectStream
	CLRs  []*core.PdfObjectStream
}

func (dss *DSSCerts) loadFromDict(dict *core.PdfObjectDictionary, suffix string) error {
	if obj := dict.Get(core.PdfObjectName("Cert" + suffix)); obj != nil {
		if certs, ok := core.GetArray(obj); ok {
			for _, e := range certs.Elements() {
				if stream, ok := core.GetStream(e); ok {
					dss.Certs = append(dss.Certs, stream)
				}
			}
		}
	}

	if obj := dict.Get(core.PdfObjectName("CRL" + suffix)); obj != nil {
		if certs, ok := core.GetArray(obj); ok {
			for _, e := range certs.Elements() {
				if stream, ok := core.GetStream(e); ok {
					dss.CLRs = append(dss.CLRs, stream)
				}
			}
		}
	}

	if obj := dict.Get(core.PdfObjectName("OCSP" + suffix)); obj != nil {
		if certs, ok := core.GetArray(obj); ok {
			for _, e := range certs.Elements() {
				if stream, ok := core.GetStream(e); ok {
					dss.OCSPs = append(dss.OCSPs, stream)
				}
			}
		}
	}
	return nil
}

// DSS describes PDF DSS info.
type DSS struct {
	DSSCerts
	VRI map[string]DSSCerts
}

func (dss *DSS) loadFromDict(dict *core.PdfObjectDictionary) error {
	if err := dss.DSSCerts.loadFromDict(dict, "s"); err != nil {
		return err
	}
	if obj := dict.Get("OCSPs"); obj != nil {
		if certs, ok := core.GetArray(obj); ok {
			for _, e := range certs.Elements() {
				if stream, ok := core.GetStream(e); ok {
					dss.OCSPs = append(dss.OCSPs, stream)
				}
			}
		}
	}

	if obj := dict.Get("VRI"); obj != nil {
		if vriDict, ok := core.GetDict(obj); ok {
			dss.VRI = make(map[string]DSSCerts)
			for _, key := range vriDict.Keys() {
				var certs DSSCerts
				if cDict, ok := core.GetDict(vriDict.Get(key)); ok {
					if err := certs.loadFromDict(cDict, ""); err != nil {
						return err
					}
					dss.VRI[strings.ToUpper(key.String())] = certs
				}
			}
		}
	}
	return nil
}
