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

func (dss *DSSCerts) toDict(suffix string) *core.PdfObjectDictionary {
	dict := core.MakeDict()
	var objects []core.PdfObject
	for _, e := range dss.Certs {
		objects = append(objects, e)
	}
	dict.Set(core.PdfObjectName("Cert"+suffix), core.MakeArray(objects...))
	objects = nil
	for _, e := range dss.CLRs {
		objects = append(objects, e)
	}
	dict.Set(core.PdfObjectName("CRL"+suffix), core.MakeArray(objects...))

	objects = nil
	for _, e := range dss.OCSPs {
		objects = append(objects, e)
	}
	dict.Set(core.PdfObjectName("OCSP"+suffix), core.MakeArray(objects...))

	return dict
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

func (dss *DSS) toDict() *core.PdfObjectDictionary {
	dict := dss.DSSCerts.toDict("s")
	vri := core.MakeDict()
	dict.Set("VRI", vri)
	for key, value := range dss.VRI {
		vri.Set(*core.MakeName(key), value.toDict(""))
	}
	return dict
}
