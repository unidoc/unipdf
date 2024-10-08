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

package sanitize ;import (_f "github.com/unidoc/unipdf/v3/common";_c "github.com/unidoc/unipdf/v3/core";);

// Sanitizer represents a sanitizer object.
// It implements the Optimizer interface to access the objects field from the writer.
type Sanitizer struct{_g SanitizationOpts ;_a map[string ]int ;};func (_gf *Sanitizer )analyze (_cae []_c .PdfObject ){_beb :=map[string ]int {};for _ ,_ab :=range _cae {switch _ba :=_ab .(type ){case *_c .PdfIndirectObject :_fab ,_bb :=_c .GetDict (_ba .PdfObject );
if _bb {if _adf ,_ddb :=_c .GetName (_fab .Get ("\u0054\u0079\u0070\u0065"));_ddb &&*_adf =="\u0043a\u0074\u0061\u006c\u006f\u0067"{if _ ,_fae :=_c .GetIndirect (_fab .Get ("\u004f\u0070\u0065\u006e\u0041\u0063\u0074\u0069\u006f\u006e"));_fae {_beb ["\u004f\u0070\u0065\u006e\u0041\u0063\u0074\u0069\u006f\u006e"]++;
};}else if _dcf ,_fc :=_c .GetName (_fab .Get ("\u0053"));_fc {_aecg :=_dcf .String ();if _aecg =="\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074"||_aecg =="\u0055\u0052\u0049"||_aecg =="\u0047\u006f\u0054\u006f"||_aecg =="\u0047\u006f\u0054o\u0052"||_aecg =="\u004c\u0061\u0075\u006e\u0063\u0068"{_beb [_aecg ]++;
}else if _aecg =="\u0052e\u006e\u0064\u0069\u0074\u0069\u006fn"{if _ ,_gg :=_c .GetStream (_fab .Get ("\u004a\u0053"));_gg {_beb [_aecg ]++;};};}else if _eg :=_fab .Get ("\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074");_eg !=nil {_beb ["\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074"]++;
}else if _de ,_gdd :=_c .GetIndirect (_fab .Get ("\u0050\u0061\u0072\u0065\u006e\u0074"));_gdd {if _aeb ,_ffc :=_c .GetDict (_de .PdfObject );_ffc {if _aca ,_bea :=_c .GetDict (_aeb .Get ("\u0041\u0041"));_bea {_dee :=_aca .Get ("\u004b");_df ,_cgc :=_c .GetIndirect (_dee );
if _cgc {if _bd ,_dcfe :=_c .GetDict (_df .PdfObject );_dcfe {if _cafb ,_deec :=_c .GetName (_bd .Get ("\u0053"));_deec &&*_cafb =="\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074"{_beb ["\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074"]++;
}else if _ ,_baaf :=_c .GetString (_bd .Get ("\u004a\u0053"));_baaf {_beb ["\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074"]++;}else {_ffg :=_aca .Get ("\u0046");if _ffg !=nil {_da ,_gdc :=_c .GetIndirect (_ffg );if _gdc {if _bgd ,_ggd :=_c .GetDict (_da .PdfObject );
_ggd {if _aef ,_bgf :=_c .GetName (_bgd .Get ("\u0053"));_bgf {_edf :=_aef .String ();_beb [_edf ]++;};};};};};};};};};};};};};_gf ._a =_beb ;};

// SanitizationOpts specifies the objects to be removed during sanitization.
type SanitizationOpts struct{

// JavaScript specifies wether JavaScript action should be removed. JavaScript Actions, section 12.6.4.16 of PDF32000_2008
JavaScript bool ;

// URI specifies if URI actions should be removed. 12.6.4.7 URI Actions, PDF32000_2008.
URI bool ;

// GoToR removes remote GoTo actions. 12.6.4.3 Remote Go-To Actions, PDF32000_2008.
GoToR bool ;

// GoTo specifies wether GoTo actions should be removed. 12.6.4.2 Go-To Actions, PDF32000_2008.
GoTo bool ;

// RenditionJS enables removing of `JS` entry from a Rendition Action.
// The `JS` entry has a value of text string or stream containing a JavaScript script that shall be executed when the action is triggered.
// 12.6.4.13 Rendition Actions Table 214, PDF32000_2008.
RenditionJS bool ;

// OpenAction removes OpenAction entry from the document catalog.
OpenAction bool ;

// Launch specifies wether Launch Action should be removed.
// A launch action launches an application or opens or prints a document.
// 12.6.4.5 Launch Actions, PDF32000_2008.
Launch bool ;};func (_e *Sanitizer )processObjects (_ec []_c .PdfObject )([]_c .PdfObject ,error ){_ca :=[]_c .PdfObject {};_dc :=_e ._g ;for _ ,_ea :=range _ec {switch _dd :=_ea .(type ){case *_c .PdfIndirectObject :_gb ,_fb :=_c .GetDict (_dd );if _fb {if _fbc ,_caf :=_c .GetName (_gb .Get ("\u0054\u0079\u0070\u0065"));
_caf &&*_fbc =="\u0043a\u0074\u0061\u006c\u006f\u0067"{if _ ,_af :=_c .GetIndirect (_gb .Get ("\u004f\u0070\u0065\u006e\u0041\u0063\u0074\u0069\u006f\u006e"));_af &&_dc .OpenAction {_gb .Remove ("\u004f\u0070\u0065\u006e\u0041\u0063\u0074\u0069\u006f\u006e");
};}else if _dg ,_ae :=_c .GetName (_gb .Get ("\u0053"));_ae {switch *_dg {case "\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074":if _dc .JavaScript {if _be ,_eaa :=_c .GetStream (_gb .Get ("\u004a\u0053"));_eaa {_cg :=[]byte {};_bec ,_gc :=_c .MakeStream (_cg ,nil );
if _gc ==nil {*_be =*_bec ;};};_f .Log .Debug ("\u004a\u0061\u0076\u0061\u0073\u0063\u0072\u0069\u0070\u0074\u0020a\u0063\u0074\u0069\u006f\u006e\u0020\u0073\u006b\u0069\u0070p\u0065\u0064\u002e");continue ;};case "\u0055\u0052\u0049":if _dc .URI {_f .Log .Debug ("\u0055\u0052\u0049\u0020ac\u0074\u0069\u006f\u006e\u0020\u0073\u006b\u0069\u0070\u0070\u0065\u0064\u002e");
continue ;};case "\u0047\u006f\u0054\u006f":if _dc .GoTo {_f .Log .Debug ("G\u004fT\u004f\u0020\u0061\u0063\u0074\u0069\u006f\u006e \u0073\u006b\u0069\u0070pe\u0064\u002e");continue ;};case "\u0047\u006f\u0054o\u0052":if _dc .GoToR {_f .Log .Debug ("R\u0065\u006d\u006f\u0074\u0065\u0020G\u006f\u0054\u004f\u0020\u0061\u0063\u0074\u0069\u006fn\u0020\u0073\u006bi\u0070p\u0065\u0064\u002e");
continue ;};case "\u004c\u0061\u0075\u006e\u0063\u0068":if _dc .Launch {_f .Log .Debug ("\u004a\u0061\u0076\u0061\u0073\u0063\u0072\u0069\u0070\u0074\u0020a\u0063\u0074\u0069\u006f\u006e\u0020\u0073\u006b\u0069\u0070p\u0065\u0064\u002e");continue ;};case "\u0052e\u006e\u0064\u0069\u0074\u0069\u006fn":if _cb ,_cbc :=_c .GetStream (_gb .Get ("\u004a\u0053"));
_cbc {_fa :=[]byte {};_fad ,_aec :=_c .MakeStream (_fa ,nil );if _aec ==nil {*_cb =*_fad ;};};};}else if _ce :=_gb .Get ("\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074");_ce !=nil &&_dc .JavaScript {continue ;}else if _dcc ,_gd :=_c .GetName (_gb .Get ("\u0054\u0079\u0070\u0065"));
_gd &&*_dcc =="\u0041\u006e\u006eo\u0074"&&_dc .JavaScript {if _fag ,_eb :=_c .GetIndirect (_gb .Get ("\u0050\u0061\u0072\u0065\u006e\u0074"));_eb {if _ge ,_bf :=_c .GetDict (_fag .PdfObject );_bf {if _cf ,_aea :=_c .GetDict (_ge .Get ("\u0041\u0041"));
_aea {_cfb ,_gbe :=_c .GetIndirect (_cf .Get ("\u004b"));if _gbe {if _cge ,_ebc :=_c .GetDict (_cfb .PdfObject );_ebc {if _ed ,_geg :=_c .GetName (_cge .Get ("\u0053"));_geg &&*_ed =="\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074"{_cge .Clear ();
}else if _cc :=_cf .Get ("\u0046");_cc !=nil {if _dda ,_cd :=_c .GetIndirect (_cc );_cd {if _ff ,_bg :=_c .GetDict (_dda .PdfObject );_bg {if _ecg ,_bfe :=_c .GetName (_ff .Get ("\u0053"));_bfe &&*_ecg =="\u004a\u0061\u0076\u0061\u0053\u0063\u0072\u0069\u0070\u0074"{_ff .Clear ();
};};};};};};};};};};};case *_c .PdfObjectStream :_f .Log .Debug ("\u0070d\u0066\u0020\u006f\u0062j\u0065\u0063\u0074\u0020\u0073t\u0072e\u0061m\u0020\u0074\u0079\u0070\u0065\u0020\u0025T",_dd );case *_c .PdfObjectStreams :_f .Log .Debug ("\u0070\u0064\u0066\u0020\u006f\u0062\u006a\u0065\u0063\u0074\u0020s\u0074\u0072\u0065\u0061\u006d\u0073\u0020\u0074\u0079\u0070e\u0020\u0025\u0054",_dd );
default:_f .Log .Debug ("u\u006e\u006b\u006e\u006fwn\u0020p\u0064\u0066\u0020\u006f\u0062j\u0065\u0063\u0074\u0020\u0025\u0054",_dd );};_ca =append (_ca ,_ea );};_e .analyze (_ca );return _ca ,nil ;};

// Optimize optimizes `objects` and returns updated list of objects.
func (_d *Sanitizer )Optimize (objects []_c .PdfObject )([]_c .PdfObject ,error ){return _d .processObjects (objects );};

// New returns a new sanitizer object.
func New (opts SanitizationOpts )*Sanitizer {return &Sanitizer {_g :opts }};

// GetSuspiciousObjects returns a count of each detected suspicious object.
func (_aad *Sanitizer )GetSuspiciousObjects ()map[string ]int {return _aad ._a };