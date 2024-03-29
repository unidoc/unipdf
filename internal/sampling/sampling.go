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

package sampling ;import (_aa "github.com/unidoc/unipdf/v3/internal/bitwise";_b "github.com/unidoc/unipdf/v3/internal/imageutil";_d "io";);type SampleReader interface{ReadSample ()(uint32 ,error );ReadSamples (_db []uint32 )error ;};func (_aee *Writer )WriteSamples (samples []uint32 )error {for _eaa :=0;
_eaa < len (samples );_eaa ++{if _gce :=_aee .WriteSample (samples [_eaa ]);_gce !=nil {return _gce ;};};return nil ;};func ResampleUint32 (data []uint32 ,bitsPerInputSample int ,bitsPerOutputSample int )[]uint32 {var _eeb []uint32 ;_gf :=bitsPerOutputSample ;
var _dad uint32 ;var _gfg uint32 ;_ae :=0;_fb :=0;_bfg :=0;for _bfg < len (data ){if _ae > 0{_cf :=_ae ;if _gf < _cf {_cf =_gf ;};_dad =(_dad <<uint (_cf ))|(_gfg >>uint (bitsPerInputSample -_cf ));_ae -=_cf ;if _ae > 0{_gfg =_gfg <<uint (_cf );}else {_gfg =0;
};_gf -=_cf ;if _gf ==0{_eeb =append (_eeb ,_dad );_gf =bitsPerOutputSample ;_dad =0;_fb ++;};}else {_dc :=data [_bfg ];_bfg ++;_fc :=bitsPerInputSample ;if _gf < _fc {_fc =_gf ;};_ae =bitsPerInputSample -_fc ;_dad =(_dad <<uint (_fc ))|(_dc >>uint (_ae ));
if _fc < bitsPerInputSample {_gfg =_dc <<uint (_fc );};_gf -=_fc ;if _gf ==0{_eeb =append (_eeb ,_dad );_gf =bitsPerOutputSample ;_dad =0;_fb ++;};};};for _ae >=bitsPerOutputSample {_ab :=_ae ;if _gf < _ab {_ab =_gf ;};_dad =(_dad <<uint (_ab ))|(_gfg >>uint (bitsPerInputSample -_ab ));
_ae -=_ab ;if _ae > 0{_gfg =_gfg <<uint (_ab );}else {_gfg =0;};_gf -=_ab ;if _gf ==0{_eeb =append (_eeb ,_dad );_gf =bitsPerOutputSample ;_dad =0;_fb ++;};};if _gf > 0&&_gf < bitsPerOutputSample {_dad <<=uint (_gf );_eeb =append (_eeb ,_dad );};return _eeb ;
};type Writer struct{_fce _b .ImageBase ;_agc *_aa .Writer ;_gca ,_cc int ;_cfd bool ;};func ResampleBytes (data []byte ,bitsPerSample int )[]uint32 {var _bc []uint32 ;_gdd :=bitsPerSample ;var _cg uint32 ;var _ag byte ;_eb :=0;_gb :=0;_bdd :=0;for _bdd < len (data ){if _eb > 0{_gc :=_eb ;
if _gdd < _gc {_gc =_gdd ;};_cg =(_cg <<uint (_gc ))|uint32 (_ag >>uint (8-_gc ));_eb -=_gc ;if _eb > 0{_ag =_ag <<uint (_gc );}else {_ag =0;};_gdd -=_gc ;if _gdd ==0{_bc =append (_bc ,_cg );_gdd =bitsPerSample ;_cg =0;_gb ++;};}else {_cge :=data [_bdd ];
_bdd ++;_da :=8;if _gdd < _da {_da =_gdd ;};_eb =8-_da ;_cg =(_cg <<uint (_da ))|uint32 (_cge >>uint (_eb ));if _da < 8{_ag =_cge <<uint (_da );};_gdd -=_da ;if _gdd ==0{_bc =append (_bc ,_cg );_gdd =bitsPerSample ;_cg =0;_gb ++;};};};for _eb >=bitsPerSample {_ed :=_eb ;
if _gdd < _ed {_ed =_gdd ;};_cg =(_cg <<uint (_ed ))|uint32 (_ag >>uint (8-_ed ));_eb -=_ed ;if _eb > 0{_ag =_ag <<uint (_ed );}else {_ag =0;};_gdd -=_ed ;if _gdd ==0{_bc =append (_bc ,_cg );_gdd =bitsPerSample ;_cg =0;_gb ++;};};return _bc ;};type SampleWriter interface{WriteSample (_ea uint32 )error ;
WriteSamples (_cfg []uint32 )error ;};func NewReader (img _b .ImageBase )*Reader {return &Reader {_fg :_aa .NewReader (img .Data ),_f :img ,_bd :img .ColorComponents ,_bf :img .BytesPerLine *8!=img .ColorComponents *img .BitsPerComponent *img .Width };
};func NewWriter (img _b .ImageBase )*Writer {return &Writer {_agc :_aa .NewWriterMSB (img .Data ),_fce :img ,_cc :img .ColorComponents ,_cfd :img .BytesPerLine *8!=img .ColorComponents *img .BitsPerComponent *img .Width };};type Reader struct{_f _b .ImageBase ;
_fg *_aa .Reader ;_e ,_g ,_bd int ;_bf bool ;};func (_fa *Writer )WriteSample (sample uint32 )error {if _ ,_gdf :=_fa ._agc .WriteBits (uint64 (sample ),_fa ._fce .BitsPerComponent );_gdf !=nil {return _gdf ;};_fa ._cc --;if _fa ._cc ==0{_fa ._cc =_fa ._fce .ColorComponents ;
_fa ._gca ++;};if _fa ._gca ==_fa ._fce .Width {if _fa ._cfd {_fa ._agc .FinishByte ();};_fa ._gca =0;};return nil ;};func (_c *Reader )ReadSample ()(uint32 ,error ){if _c ._g ==_c ._f .Height {return 0,_d .EOF ;};_ee ,_bb :=_c ._fg .ReadBits (byte (_c ._f .BitsPerComponent ));
if _bb !=nil {return 0,_bb ;};_c ._bd --;if _c ._bd ==0{_c ._bd =_c ._f .ColorComponents ;_c ._e ++;};if _c ._e ==_c ._f .Width {if _c ._bf {_c ._fg .ConsumeRemainingBits ();};_c ._e =0;_c ._g ++;};return uint32 (_ee ),nil ;};func (_bdb *Reader )ReadSamples (samples []uint32 )(_fe error ){for _gd :=0;
_gd < len (samples );_gd ++{samples [_gd ],_fe =_bdb .ReadSample ();if _fe !=nil {return _fe ;};};return nil ;};