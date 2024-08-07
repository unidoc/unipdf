//
// Copyright 2020 FoxyUtils ehf. All rights reserved.
//
// This is a commercial product and requires a license to operate.
// A trial license can be obtained at https://unidoc.io
//
// DO NOT EDIT: generated by unitwist Go source code obfuscator.
//
// Use of this source code is governed by the UniDoc End User License Agreement
// terms that can be accessed at https://unidoc.io/eula/

package pdfaid ;import (_cb "fmt";_b "github.com/trimmer-io/go-xmp/xmp";_cf "github.com/unidoc/unipdf/v3/model/xmputil/pdfaextension";);

// CanTag implements xmp.Model interface.
func (_eb *Model )CanTag (tag string )bool {_ ,_be :=_b .GetNativeField (_eb ,tag );return _be ==nil };func init (){_b .Register (Namespace ,_b .XmpMetadata );_cf .RegisterSchema (Namespace ,&Schema )};

// SyncToXMP implements xmp.Model interface.
func (_ce *Model )SyncToXMP (d *_b .Document )error {return nil };var Namespace =_b .NewNamespace ("\u0070\u0064\u0066\u0061\u0069\u0064","\u0068\u0074\u0074p\u003a\u002f\u002f\u0077w\u0077\u002e\u0061\u0069\u0069\u006d\u002eo\u0072\u0067\u002f\u0070\u0064\u0066\u0061\u002f\u006e\u0073\u002f\u0069\u0064\u002f",NewModel );


// Can implements xmp.Model interface.
func (_d *Model )Can (nsName string )bool {return Namespace .GetName ()==nsName };var Schema =_cf .Schema {NamespaceURI :Namespace .URI ,Prefix :Namespace .Name ,Schema :"\u0050D\u0046/\u0041\u0020\u0049\u0044\u0020\u0053\u0063\u0068\u0065\u006d\u0061",Property :[]_cf .Property {{Category :_cf .PropertyCategoryInternal ,Description :"\u0050\u0061\u0072\u0074 o\u0066\u0020\u0050\u0044\u0046\u002f\u0041\u0020\u0073\u0074\u0061\u006e\u0064\u0061r\u0064",Name :"\u0070\u0061\u0072\u0074",ValueType :_cf .ValueTypeNameInteger },{Category :_cf .PropertyCategoryInternal ,Description :"A\u006d\u0065\u006e\u0064\u006d\u0065n\u0074\u0020\u006f\u0066\u0020\u0050\u0044\u0046\u002fA\u0020\u0073\u0074a\u006ed\u0061\u0072\u0064",Name :"\u0061\u006d\u0064",ValueType :_cf .ValueTypeNameText },{Category :_cf .PropertyCategoryInternal ,Description :"C\u006f\u006e\u0066\u006f\u0072\u006da\u006e\u0063\u0065\u0020\u006c\u0065v\u0065\u006c\u0020\u006f\u0066\u0020\u0050D\u0046\u002f\u0041\u0020\u0073\u0074\u0061\u006e\u0064\u0061r\u0064",Name :"c\u006f\u006e\u0066\u006f\u0072\u006d\u0061\u006e\u0063\u0065",ValueType :_cf .ValueTypeNameText }},ValueType :nil };


// Namespaces implements xmp.Model interface.
func (_gg *Model )Namespaces ()_b .NamespaceList {return _b .NamespaceList {Namespace }};

// Model is the XMP model for the PdfA metadata.
type Model struct{Part int `xmp:"pdfaid:part"`;Conformance string `xmp:"pdfaid:conformance"`;};

// SetTag implements xmp.Model interface.
func (_ee *Model )SetTag (tag ,value string )error {if _gdd :=_b .SetNativeField (_ee ,tag ,value );_gdd !=nil {return _cb .Errorf ("\u0025\u0073\u003a\u0020\u0025\u0076",Namespace .GetName (),_gdd );};return nil ;};

// GetTag implements xmp.Model interface.
func (_gc *Model )GetTag (tag string )(string ,error ){_ec ,_cd :=_b .GetNativeField (_gc ,tag );if _cd !=nil {return "",_cb .Errorf ("\u0025\u0073\u003a\u0020\u0025\u0076",Namespace .GetName (),_cd );};return _ec ,nil ;};

// SyncModel implements xmp.Model interface.
func (_cc *Model )SyncModel (d *_b .Document )error {return nil };

// MakeModel gets or create sa new model for PDF/A ID namespace.
func MakeModel (d *_b .Document )(*Model ,error ){_e ,_ga :=d .MakeModel (Namespace );if _ga !=nil {return nil ,_ga ;};return _e .(*Model ),nil ;};

// SyncFromXMP implements xmp.Model interface.
func (_gd *Model )SyncFromXMP (d *_b .Document )error {return nil };

// NewModel creates a new model.
func NewModel (name string )_b .Model {return &Model {}};var _ _b .Model =(*Model )(nil );