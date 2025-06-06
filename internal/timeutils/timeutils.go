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

package timeutils ;import (_a "errors";_b "fmt";_ecc "regexp";_ec "strconv";_g "time";);func FormatPdfTime (in _g .Time )string {_bd :=in .Format ("\u002d\u0030\u0037\u003a\u0030\u0030");_f ,_ :=_ec .ParseInt (_bd [1:3],10,32);_af ,_ :=_ec .ParseInt (_bd [4:6],10,32);
_fd :=int64 (in .Year ());_gd :=int64 (in .Month ());_aa :=int64 (in .Day ());_fa :=int64 (in .Hour ());_ef :=int64 (in .Minute ());_bf :=int64 (in .Second ());_ca :=_bd [0];return _b .Sprintf ("\u0044\u003a\u0025\u002e\u0034\u0064\u0025\u002e\u0032\u0064\u0025\u002e\u0032\u0064\u0025\u002e\u0032\u0064\u0025\u002e\u0032\u0064\u0025\u002e2\u0064\u0025\u0063\u0025\u002e2\u0064\u0027%\u002e\u0032\u0064\u0027",_fd ,_gd ,_aa ,_fa ,_ef ,_bf ,_ca ,_f ,_af );
};func ParsePdfTime (pdfTime string )(_g .Time ,error ){_d :=_acb .FindAllStringSubmatch (pdfTime ,1);if len (_d )< 1{if len (pdfTime )> 0&&pdfTime [0]!='D'{pdfTime =_b .Sprintf ("\u0044\u003a\u0025\u0073",pdfTime );return ParsePdfTime (pdfTime );};return _g .Time {},_b .Errorf ("\u0069n\u0076\u0061\u006c\u0069\u0064\u0020\u0064\u0061\u0074\u0065\u0020s\u0074\u0072\u0069\u006e\u0067\u0020\u0028\u0025\u0073\u0029",pdfTime );
};if len (_d [0])!=10{return _g .Time {},_a .New ("\u0069\u006e\u0076\u0061\u006c\u0069\u0064\u0020\u0072\u0065\u0067\u0065\u0078p\u0020\u0067\u0072\u006f\u0075\u0070 \u006d\u0061\u0074\u0063\u0068\u0020\u006c\u0065\u006e\u0067\u0074\u0068\u0020!\u003d\u0020\u0031\u0030");
};_bg ,_ :=_ec .ParseInt (_d [0][1],10,32);_agd ,_ :=_ec .ParseInt (_d [0][2],10,32);_ab ,_ :=_ec .ParseInt (_d [0][3],10,32);_gdc ,_ :=_ec .ParseInt (_d [0][4],10,32);_cf ,_ :=_ec .ParseInt (_d [0][5],10,32);_fg ,_ :=_ec .ParseInt (_d [0][6],10,32);var (_cg byte ;
_ce int64 ;_bfc int64 ;);_cg ='+';if len (_d [0][7])> 0{if _d [0][7]=="\u002d"{_cg ='-';}else if _d [0][7]=="\u005a"{_cg ='Z';};};if len (_d [0][8])> 0{_ce ,_ =_ec .ParseInt (_d [0][8],10,32);}else {_ce =0;};if len (_d [0][9])> 0{_bfc ,_ =_ec .ParseInt (_d [0][9],10,32);
}else {_bfc =0;};_ac :=int (_ce *60*60+_bfc *60);switch _cg {case '-':_ac =-_ac ;case 'Z':_ac =0;};_bgd :=_b .Sprintf ("\u0055\u0054\u0043\u0025\u0063\u0025\u002e\u0032\u0064\u0025\u002e\u0032\u0064",_cg ,_ce ,_bfc );_fe :=_g .FixedZone (_bgd ,_ac );return _g .Date (int (_bg ),_g .Month (_agd ),int (_ab ),int (_gdc ),int (_cf ),int (_fg ),0,_fe ),nil ;
};var _acb =_ecc .MustCompile ("\u005cs\u002a\u0044\u005cs\u002a\u003a\u005cs\u002a(\\\u0064\u007b\u0034\u007d\u0029\u0028\u005cd\u007b\u0032\u007d\u0029\u0028\u005c\u0064\u007b\u0032\u007d\u0029\u0028\u005c\u0064\u007b\u0032\u007d\u0029\u0028\u005c\u0064\u007b\u0032\u007d\u0029\u0028\u005c\u0064{2\u007d)\u003f\u0028\u005b\u002b\u002d\u005a]\u0029\u003f\u0028\u005c\u0064{\u0032\u007d\u0029\u003f\u0027\u003f\u0028\u005c\u0064\u007b\u0032}\u0029\u003f");
