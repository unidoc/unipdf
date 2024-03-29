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

package jbig2 ;import (_ff "github.com/unidoc/unipdf/v3/internal/bitwise";_b "github.com/unidoc/unipdf/v3/internal/jbig2/decoder";_e "github.com/unidoc/unipdf/v3/internal/jbig2/document";_fb "github.com/unidoc/unipdf/v3/internal/jbig2/document/segments";
_a "github.com/unidoc/unipdf/v3/internal/jbig2/errors";_d "sort";);func DecodeBytes (encoded []byte ,parameters _b .Parameters ,globals ...Globals )([]byte ,error ){var _g Globals ;if len (globals )> 0{_g =globals [0];};_ge ,_ac :=_b .Decode (encoded ,parameters ,_g .ToDocumentGlobals ());
if _ac !=nil {return nil ,_ac ;};return _ge .DecodeNextPage ();};type Globals map[int ]*_fb .Header ;func (_be Globals )ToDocumentGlobals ()*_e .Globals {if _be ==nil {return nil ;};_ba :=[]*_fb .Header {};for _ ,_ab :=range _be {_ba =append (_ba ,_ab );
};_d .Slice (_ba ,func (_bb ,_ec int )bool {return _ba [_bb ].SegmentNumber < _ba [_ec ].SegmentNumber });return &_e .Globals {Segments :_ba };};func DecodeGlobals (encoded []byte )(Globals ,error ){const _ffa ="\u0044\u0065\u0063\u006f\u0064\u0065\u0047\u006c\u006f\u0062\u0061\u006c\u0073";
_fg :=_ff .NewReader (encoded );_dc ,_ga :=_e .DecodeDocument (_fg ,nil );if _ga !=nil {return nil ,_a .Wrap (_ga ,_ffa ,"");};if _dc .GlobalSegments ==nil ||(_dc .GlobalSegments .Segments ==nil ){return nil ,_a .Error (_ffa ,"\u006eo\u0020\u0067\u006c\u006f\u0062\u0061\u006c\u0020\u0073\u0065\u0067m\u0065\u006e\u0074\u0073\u0020\u0066\u006f\u0075\u006e\u0064");
};_bc :=Globals {};for _ ,_dce :=range _dc .GlobalSegments .Segments {_bc [int (_dce .SegmentNumber )]=_dce ;};return _bc ,nil ;};