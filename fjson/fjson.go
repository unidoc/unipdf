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

// Package fjson provides support for loading PDF form field data from JSON data/files.
package fjson ;import (_b "encoding/json";_fg "github.com/unidoc/unipdf/v3/common";_fde "github.com/unidoc/unipdf/v3/core";_fd "github.com/unidoc/unipdf/v3/model";_g "io";_f "os";);

// LoadFromJSON loads JSON form data from `r`.
func LoadFromJSON (r _g .Reader )(*FieldData ,error ){var _ff FieldData ;_d :=_b .NewDecoder (r ).Decode (&_ff ._bd );if _d !=nil {return nil ,_d ;};return &_ff ,nil ;};

// FieldData represents form field data loaded from JSON file.
type FieldData struct{_bd []fieldValue };

// FieldImageValues implements model.FieldImageProvider interface.
func (_ddf *FieldData )FieldImageValues ()(map[string ]*_fd .Image ,error ){_ad :=make (map[string ]*_fd .Image );for _ ,_ed :=range _ddf ._bd {if _ed .ImageValue !=nil {_ad [_ed .Name ]=_ed .ImageValue ;};};return _ad ,nil ;};

// LoadFromPDF loads form field data from a PDF.
func LoadFromPDF (rs _g .ReadSeeker )(*FieldData ,error ){_c ,_ca :=_fd .NewPdfReader (rs );if _ca !=nil {return nil ,_ca ;};if _c .AcroForm ==nil {return nil ,nil ;};var _da []fieldValue ;_df :=_c .AcroForm .AllFields ();for _ ,_fe :=range _df {var _eg []string ;
_ea :=make (map[string ]struct{});_eff ,_eae :=_fe .FullName ();if _eae !=nil {return nil ,_eae ;};if _a ,_af :=_fe .V .(*_fde .PdfObjectString );_af {_da =append (_da ,fieldValue {Name :_eff ,Value :_a .Decoded ()});continue ;};var _dd string ;for _ ,_ee :=range _fe .Annotations {_cae ,_be :=_fde .GetName (_ee .AS );
if _be {_dd =_cae .String ();};_ab ,_bea :=_fde .GetDict (_ee .AP );if !_bea {continue ;};_de ,_ :=_fde .GetDict (_ab .Get ("\u004e"));for _ ,_bc :=range _de .Keys (){_bg :=_bc .String ();if _ ,_ga :=_ea [_bg ];!_ga {_eg =append (_eg ,_bg );_ea [_bg ]=struct{}{};
};};_ce ,_ :=_fde .GetDict (_ab .Get ("\u0044"));for _ ,_bbc :=range _ce .Keys (){_cg :=_bbc .String ();if _ ,_fba :=_ea [_cg ];!_fba {_eg =append (_eg ,_cg );_ea [_cg ]=struct{}{};};};};_afd :=fieldValue {Name :_eff ,Value :_dd ,Options :_eg };_da =append (_da ,_afd );
};_cad :=FieldData {_bd :_da };return &_cad ,nil ;};

// FieldValues implements model.FieldValueProvider interface.
func (_bde *FieldData )FieldValues ()(map[string ]_fde .PdfObject ,error ){_eb :=make (map[string ]_fde .PdfObject );for _ ,_efc :=range _bde ._bd {if len (_efc .Value )> 0{_eb [_efc .Name ]=_fde .MakeString (_efc .Value );};};return _eb ,nil ;};

// SetImage assign model.Image to a specific field identified by fieldName.
func (_acf *FieldData )SetImage (fieldName string ,img *_fd .Image ,opt []string )error {_dcb :=fieldValue {Name :fieldName ,ImageValue :img ,Options :opt };_acf ._bd =append (_acf ._bd ,_dcb );return nil ;};

// LoadFromJSONFile loads form field data from a JSON file.
func LoadFromJSONFile (filePath string )(*FieldData ,error ){_ba ,_ef :=_f .Open (filePath );if _ef !=nil {return nil ,_ef ;};defer _ba .Close ();return LoadFromJSON (_ba );};

// LoadFromPDFFile loads form field data from a PDF file.
func LoadFromPDFFile (filePath string )(*FieldData ,error ){_gc ,_beb :=_f .Open (filePath );if _beb !=nil {return nil ,_beb ;};defer _gc .Close ();return LoadFromPDF (_gc );};type fieldValue struct{Name string `json:"name"`;Value string `json:"value"`;
ImageValue *_fd .Image `json:"-"`;

// Options lists allowed values if present.
Options []string `json:"options,omitempty"`;};

// JSON returns the field data as a string in JSON format.
func (_dec FieldData )JSON ()(string ,error ){_fea ,_dc :=_b .MarshalIndent (_dec ._bd ,"","\u0020\u0020\u0020\u0020");return string (_fea ),_dc ;};

// SetImageFromFile assign image file to a specific field identified by fieldName.
func (_cd *FieldData )SetImageFromFile (fieldName string ,imagePath string ,opt []string )error {_fff ,_fge :=_f .Open (imagePath );if _fge !=nil {return _fge ;};defer _fff .Close ();_db ,_fge :=_fd .ImageHandling .Read (_fff );if _fge !=nil {_fg .Log .Error ("\u0045\u0072\u0072or\u0020\u006c\u006f\u0061\u0064\u0069\u006e\u0067\u0020\u0069\u006d\u0061\u0067\u0065\u003a\u0020\u0025\u0073",_fge );
return _fge ;};return _cd .SetImage (fieldName ,_db ,opt );};