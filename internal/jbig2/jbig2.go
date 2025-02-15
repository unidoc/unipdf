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

package jbig2 ;import (_e "github.com/unidoc/unipdf/v3/internal/bitwise";_g "github.com/unidoc/unipdf/v3/internal/jbig2/decoder";_df "github.com/unidoc/unipdf/v3/internal/jbig2/document";_d "github.com/unidoc/unipdf/v3/internal/jbig2/document/segments";
_a "github.com/unidoc/unipdf/v3/internal/jbig2/errors";_ef "sort";);type Globals map[int ]*_d .Header ;func DecodeBytes (encoded []byte ,parameters _g .Parameters ,globals ...Globals )([]byte ,error ){var _af Globals ;if len (globals )> 0{_af =globals [0];
};_gg ,_gc :=_g .Decode (encoded ,parameters ,_af .ToDocumentGlobals ());if _gc !=nil {return nil ,_gc ;};return _gg .DecodeNextPage ();};func (_f Globals )ToDocumentGlobals ()*_df .Globals {if _f ==nil {return nil ;};_ge :=[]*_d .Header {};for _ ,_cf :=range _f {_ge =append (_ge ,_cf );
};_ef .Slice (_ge ,func (_gec ,_fe int )bool {return _ge [_gec ].SegmentNumber < _ge [_fe ].SegmentNumber });return &_df .Globals {Segments :_ge };};func DecodeGlobals (encoded []byte )(Globals ,error ){const _ga ="\u0044\u0065\u0063\u006f\u0064\u0065\u0047\u006c\u006f\u0062\u0061\u006c\u0073";
_b :=_e .NewReader (encoded );_gd ,_dc :=_df .DecodeDocument (_b ,nil );if _dc !=nil {return nil ,_a .Wrap (_dc ,_ga ,"");};if _gd .GlobalSegments ==nil ||(_gd .GlobalSegments .Segments ==nil ){return nil ,_a .Error (_ga ,"\u006eo\u0020\u0067\u006c\u006f\u0062\u0061\u006c\u0020\u0073\u0065\u0067m\u0065\u006e\u0074\u0073\u0020\u0066\u006f\u0075\u006e\u0064");
};_dcc :=Globals {};for _ ,_db :=range _gd .GlobalSegments .Segments {_dcc [int (_db .SegmentNumber )]=_db ;};return _dcc ,nil ;};