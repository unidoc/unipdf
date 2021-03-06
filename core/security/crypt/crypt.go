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

package crypt ;import (_ad "crypto/aes";_bc "crypto/cipher";_gc "crypto/md5";_ge "crypto/rand";_gg "crypto/rc4";_a "fmt";_d "github.com/unidoc/unipdf/v3/common";_ab "github.com/unidoc/unipdf/v3/core/security";_g "io";);func init (){_dfd ("\u0041\u0045\u0053V\u0032",_be )};


// MakeKey implements Filter interface.
func (_ec filterV2 )MakeKey (objNum ,genNum uint32 ,ekey []byte )([]byte ,error ){return _ed (objNum ,genNum ,ekey ,false );};

// KeyLength implements Filter interface.
func (filterAESV2 )KeyLength ()int {return 128/8};

// HandlerVersion implements Filter interface.
func (filterAESV2 )HandlerVersion ()(V ,R int ){V ,R =4,4;return ;};

// KeyLength implements Filter interface.
func (_dg filterV2 )KeyLength ()int {return _dg ._eg };

// PDFVersion implements Filter interface.
func (_ca filterV2 )PDFVersion ()[2]int {return [2]int {}};type filterAESV2 struct{filterAES };

// MakeKey implements Filter interface.
func (filterAESV2 )MakeKey (objNum ,genNum uint32 ,ekey []byte )([]byte ,error ){return _ed (objNum ,genNum ,ekey ,true );};

// DecryptBytes implements Filter interface.
func (filterV2 )DecryptBytes (buf []byte ,okey []byte )([]byte ,error ){_eda ,_aefc :=_gg .NewCipher (okey );if _aefc !=nil {return nil ,_aefc ;};_d .Log .Trace ("\u0052\u00434\u0020\u0044\u0065c\u0072\u0079\u0070\u0074\u003a\u0020\u0025\u0020\u0078",buf );
_eda .XORKeyStream (buf ,buf );_d .Log .Trace ("\u0074o\u003a\u0020\u0025\u0020\u0078",buf );return buf ,nil ;};

// NewFilterV2 creates a RC4-based filter with a specified key length (in bytes).
func NewFilterV2 (length int )Filter {_aea ,_ac :=_aef (FilterDict {Length :length });if _ac !=nil {_d .Log .Error ("E\u0052\u0052\u004f\u0052\u003a\u0020\u0063\u006f\u0075l\u0064\u0020\u006e\u006f\u0074\u0020\u0063re\u0061\u0074\u0065\u0020R\u0043\u0034\u0020\u0056\u0032\u0020\u0063\u0072\u0079pt\u0020\u0066i\u006c\u0074\u0065\u0072\u003a\u0020\u0025\u0076",_ac );
return filterV2 {_eg :length };};return _aea ;};func init (){_dfd ("\u0041\u0045\u0053V\u0033",_bf )};

// MakeKey implements Filter interface.
func (filterAESV3 )MakeKey (_ ,_ uint32 ,ekey []byte )([]byte ,error ){return ekey ,nil };var _ Filter =filterAESV2 {};func (filterIdentity )HandlerVersion ()(V ,R int ){return ;};

// EncryptBytes implements Filter interface.
func (filterV2 )EncryptBytes (buf []byte ,okey []byte )([]byte ,error ){_bfa ,_dda :=_gg .NewCipher (okey );if _dda !=nil {return nil ,_dda ;};_d .Log .Trace ("\u0052\u00434\u0020\u0045\u006ec\u0072\u0079\u0070\u0074\u003a\u0020\u0025\u0020\u0078",buf );
_bfa .XORKeyStream (buf ,buf );_d .Log .Trace ("\u0074o\u003a\u0020\u0025\u0020\u0078",buf );return buf ,nil ;};var _ Filter =filterV2 {};type filterFunc func (_bd FilterDict )(Filter ,error );func (filterIdentity )PDFVersion ()[2]int {return [2]int {}};


// Filter is a common interface for crypt filter methods.
type Filter interface{

// Name returns a name of the filter that should be used in CFM field of Encrypt dictionary.
Name ()string ;

// KeyLength returns a length of the encryption key in bytes.
KeyLength ()int ;

// PDFVersion reports the minimal version of PDF document that introduced this filter.
PDFVersion ()[2]int ;

// HandlerVersion reports V and R parameters that should be used for this filter.
HandlerVersion ()(V ,R int );

// MakeKey generates a object encryption key based on file encryption key and object numbers.
// Used only for legacy filters - AESV3 doesn't change the key for each object.
MakeKey (_egg ,_caba uint32 ,_ce []byte )([]byte ,error );

// EncryptBytes encrypts a buffer using object encryption key, as returned by MakeKey.
// Implementation may reuse a buffer and encrypt data in-place.
EncryptBytes (_eb []byte ,_dga []byte )([]byte ,error );

// DecryptBytes decrypts a buffer using object encryption key, as returned by MakeKey.
// Implementation may reuse a buffer and decrypt data in-place.
DecryptBytes (_age []byte ,_ecc []byte )([]byte ,error );};var (_dfbd =make (map[string ]filterFunc ););func _cb (_dba string )(filterFunc ,error ){_fde :=_dfbd [_dba ];if _fde ==nil {return nil ,_a .Errorf ("\u0075\u006e\u0073\u0075p\u0070\u006f\u0072\u0074\u0065\u0064\u0020\u0063\u0072\u0079p\u0074 \u0066\u0069\u006c\u0074\u0065\u0072\u003a \u0025\u0071",_dba );
};return _fde ,nil ;};type filterV2 struct{_eg int };func (filterIdentity )DecryptBytes (p []byte ,okey []byte )([]byte ,error ){return p ,nil };func init (){_dfd ("\u0056\u0032",_aef )};func _bf (_ada FilterDict )(Filter ,error ){if _ada .Length ==256{_d .Log .Debug ("\u0041\u0045S\u0056\u0033\u0020c\u0072\u0079\u0070\u0074\u0020f\u0069\u006c\u0074\u0065\u0072 l\u0065\u006e\u0067\u0074\u0068\u0020\u0061\u0070\u0070\u0065\u0061\u0072\u0073\u0020\u0074\u006f\u0020\u0062e\u0020i\u006e\u0020\u0062\u0069\u0074\u0073 ra\u0074\u0068\u0065\u0072\u0020\u0074\u0068\u0061\u006e\u0020\u0062\u0079te\u0073 \u002d\u0020\u0061\u0073s\u0075m\u0069n\u0067\u0020b\u0069\u0074s \u0028\u0025\u0064\u0029",_ada .Length );
_ada .Length /=8;};if _ada .Length !=0&&_ada .Length !=32{return nil ,_a .Errorf ("\u0069\u006e\u0076\u0061\u006c\u0069\u0064\u0020\u0041\u0045\u0053\u0056\u0033\u0020\u0063\u0072\u0079\u0070\u0074\u0020\u0066\u0069\u006c\u0074e\u0072\u0020\u006c\u0065\u006eg\u0074\u0068 \u0028\u0025\u0064\u0029",_ada .Length );
};return filterAESV3 {},nil ;};

// Name implements Filter interface.
func (filterAESV3 )Name ()string {return "\u0041\u0045\u0053V\u0033"};func (filterIdentity )EncryptBytes (p []byte ,okey []byte )([]byte ,error ){return p ,nil };func (filterAES )EncryptBytes (buf []byte ,okey []byte )([]byte ,error ){_ae ,_ag :=_ad .NewCipher (okey );
if _ag !=nil {return nil ,_ag ;};_d .Log .Trace ("A\u0045\u0053\u0020\u0045nc\u0072y\u0070\u0074\u0020\u0028\u0025d\u0029\u003a\u0020\u0025\u0020\u0078",len (buf ),buf );const _cfc =_ad .BlockSize ;_agd :=_cfc -len (buf )%_cfc ;for _aa :=0;_aa < _agd ;
_aa ++{buf =append (buf ,byte (_agd ));};_d .Log .Trace ("\u0050a\u0064d\u0065\u0064\u0020\u0074\u006f \u0025\u0064 \u0062\u0079\u0074\u0065\u0073",len (buf ));_bea :=make ([]byte ,_cfc +len (buf ));_dc :=_bea [:_cfc ];if _ ,_fe :=_g .ReadFull (_ge .Reader ,_dc );
_fe !=nil {return nil ,_fe ;};_aab :=_bc .NewCBCEncrypter (_ae ,_dc );_aab .CryptBlocks (_bea [_cfc :],buf );buf =_bea ;_d .Log .Trace ("\u0074\u006f\u0020(\u0025\u0064\u0029\u003a\u0020\u0025\u0020\u0078",len (buf ),buf );return buf ,nil ;};type filterAES struct{};


// NewFilterAESV2 creates an AES-based filter with a 128 bit key (AESV2).
func NewFilterAESV2 ()Filter {_df ,_af :=_be (FilterDict {});if _af !=nil {_d .Log .Error ("E\u0052\u0052\u004f\u0052\u003a\u0020\u0063\u006f\u0075l\u0064\u0020\u006e\u006f\u0074\u0020\u0063re\u0061\u0074\u0065\u0020A\u0045\u0053\u0020\u0056\u0032\u0020\u0063\u0072\u0079pt\u0020\u0066i\u006c\u0074\u0065\u0072\u003a\u0020\u0025\u0076",_af );
return filterAESV2 {};};return _df ;};

// NewFilterAESV3 creates an AES-based filter with a 256 bit key (AESV3).
func NewFilterAESV3 ()Filter {_ged ,_gb :=_bf (FilterDict {});if _gb !=nil {_d .Log .Error ("E\u0052\u0052\u004f\u0052\u003a\u0020\u0063\u006f\u0075l\u0064\u0020\u006e\u006f\u0074\u0020\u0063re\u0061\u0074\u0065\u0020A\u0045\u0053\u0020\u0056\u0033\u0020\u0063\u0072\u0079pt\u0020\u0066i\u006c\u0074\u0065\u0072\u003a\u0020\u0025\u0076",_gb );
return filterAESV3 {};};return _ged ;};func (filterIdentity )Name ()string {return "\u0049\u0064\u0065\u006e\u0074\u0069\u0074\u0079"};

// NewIdentity creates an identity filter that bypasses all data without changes.
func NewIdentity ()Filter {return filterIdentity {}};var _ Filter =filterAESV3 {};

// PDFVersion implements Filter interface.
func (filterAESV3 )PDFVersion ()[2]int {return [2]int {2,0}};

// HandlerVersion implements Filter interface.
func (_cab filterV2 )HandlerVersion ()(V ,R int ){V ,R =2,3;return ;};

// Name implements Filter interface.
func (filterV2 )Name ()string {return "\u0056\u0032"};

// Name implements Filter interface.
func (filterAESV2 )Name ()string {return "\u0041\u0045\u0053V\u0032"};func (filterIdentity )KeyLength ()int {return 0};type filterAESV3 struct{filterAES };func _ed (_dcb ,_fc uint32 ,_dfb []byte ,_fcb bool )([]byte ,error ){_ggb :=make ([]byte ,len (_dfb )+5);
for _abef :=0;_abef < len (_dfb );_abef ++{_ggb [_abef ]=_dfb [_abef ];};for _bfg :=0;_bfg < 3;_bfg ++{_cg :=byte ((_dcb >>uint32 (8*_bfg ))&0xff);_ggb [_bfg +len (_dfb )]=_cg ;};for _fd :=0;_fd < 2;_fd ++{_gd :=byte ((_fc >>uint32 (8*_fd ))&0xff);_ggb [_fd +len (_dfb )+3]=_gd ;
};if _fcb {_ggb =append (_ggb ,0x73);_ggb =append (_ggb ,0x41);_ggb =append (_ggb ,0x6C);_ggb =append (_ggb ,0x54);};_cgd :=_gc .New ();_cgd .Write (_ggb );_bcg :=_cgd .Sum (nil );if len (_dfb )+5< 16{return _bcg [0:len (_dfb )+5],nil ;};return _bcg ,nil ;
};

// HandlerVersion implements Filter interface.
func (filterAESV3 )HandlerVersion ()(V ,R int ){V ,R =5,6;return ;};

// KeyLength implements Filter interface.
func (filterAESV3 )KeyLength ()int {return 256/8};func (filterIdentity )MakeKey (objNum ,genNum uint32 ,fkey []byte )([]byte ,error ){return fkey ,nil };func _be (_abd FilterDict )(Filter ,error ){if _abd .Length ==128{_d .Log .Debug ("\u0041\u0045S\u0056\u0032\u0020c\u0072\u0079\u0070\u0074\u0020f\u0069\u006c\u0074\u0065\u0072 l\u0065\u006e\u0067\u0074\u0068\u0020\u0061\u0070\u0070\u0065\u0061\u0072\u0073\u0020\u0074\u006f\u0020\u0062e\u0020i\u006e\u0020\u0062\u0069\u0074\u0073 ra\u0074\u0068\u0065\u0072\u0020\u0074\u0068\u0061\u006e\u0020\u0062\u0079te\u0073 \u002d\u0020\u0061\u0073s\u0075m\u0069n\u0067\u0020b\u0069\u0074s \u0028\u0025\u0064\u0029",_abd .Length );
_abd .Length /=8;};if _abd .Length !=0&&_abd .Length !=16{return nil ,_a .Errorf ("\u0069\u006e\u0076\u0061\u006c\u0069\u0064\u0020\u0041\u0045\u0053\u0056\u0032\u0020\u0063\u0072\u0079\u0070\u0074\u0020\u0066\u0069\u006c\u0074e\u0072\u0020\u006c\u0065\u006eg\u0074\u0068 \u0028\u0025\u0064\u0029",_abd .Length );
};return filterAESV2 {},nil ;};type filterIdentity struct{};func _dfd (_dfg string ,_bg filterFunc ){if _ ,_ea :=_dfbd [_dfg ];_ea {panic ("\u0061l\u0072e\u0061\u0064\u0079\u0020\u0072e\u0067\u0069s\u0074\u0065\u0072\u0065\u0064");};_dfbd [_dfg ]=_bg ;
};

// NewFilter creates CryptFilter from a corresponding dictionary.
func NewFilter (d FilterDict )(Filter ,error ){_fda ,_cad :=_cb (d .CFM );if _cad !=nil {return nil ,_cad ;};_beac ,_cad :=_fda (d );if _cad !=nil {return nil ,_cad ;};return _beac ,nil ;};

// FilterDict represents information from a CryptFilter dictionary.
type FilterDict struct{CFM string ;AuthEvent _ab .AuthEvent ;Length int ;};func (filterAES )DecryptBytes (buf []byte ,okey []byte )([]byte ,error ){_dd ,_dcd :=_ad .NewCipher (okey );if _dcd !=nil {return nil ,_dcd ;};if len (buf )< 16{_d .Log .Debug ("\u0045R\u0052\u004f\u0052\u0020\u0041\u0045\u0053\u0020\u0069\u006e\u0076a\u006c\u0069\u0064\u0020\u0062\u0075\u0066\u0020\u0025\u0073",buf );
return buf ,_a .Errorf ("\u0041\u0045\u0053\u003a B\u0075\u0066\u0020\u006c\u0065\u006e\u0020\u003c\u0020\u0031\u0036\u0020\u0028\u0025d\u0029",len (buf ));};_abe :=buf [:16];buf =buf [16:];if len (buf )%16!=0{_d .Log .Debug ("\u0020\u0069\u0076\u0020\u0028\u0025\u0064\u0029\u003a\u0020\u0025\u0020\u0078",len (_abe ),_abe );
_d .Log .Debug ("\u0062\u0075\u0066\u0020\u0028\u0025\u0064\u0029\u003a\u0020\u0025\u0020\u0078",len (buf ),buf );return buf ,_a .Errorf ("\u0041\u0045\u0053\u0020\u0062\u0075\u0066\u0020\u006c\u0065\u006e\u0067\u0074\u0068\u0020\u006e\u006f\u0074\u0020\u006d\u0075\u006c\u0074\u0069p\u006c\u0065\u0020\u006f\u0066 \u0031\u0036 \u0028\u0025\u0064\u0029",len (buf ));
};_ee :=_bc .NewCBCDecrypter (_dd ,_abe );_d .Log .Trace ("A\u0045\u0053\u0020\u0044ec\u0072y\u0070\u0074\u0020\u0028\u0025d\u0029\u003a\u0020\u0025\u0020\u0078",len (buf ),buf );_d .Log .Trace ("\u0063\u0068\u006f\u0070\u0020\u0041\u0045\u0053\u0020\u0044\u0065c\u0072\u0079\u0070\u0074\u0020\u0028\u0025\u0064\u0029\u003a \u0025\u0020\u0078",len (buf ),buf );
_ee .CryptBlocks (buf ,buf );_d .Log .Trace ("\u0074\u006f\u0020(\u0025\u0064\u0029\u003a\u0020\u0025\u0020\u0078",len (buf ),buf );if len (buf )==0{_d .Log .Trace ("\u0045\u006d\u0070\u0074\u0079\u0020b\u0075\u0066\u002c\u0020\u0072\u0065\u0074\u0075\u0072\u006e\u0069\u006e\u0067 \u0065\u006d\u0070\u0074\u0079\u0020\u0073t\u0072\u0069\u006e\u0067");
return buf ,nil ;};_cd :=int (buf [len (buf )-1]);if _cd > len (buf ){_d .Log .Debug ("\u0049\u006c\u006c\u0065g\u0061\u006c\u0020\u0070\u0061\u0064\u0020\u006c\u0065\u006eg\u0074h\u0020\u0028\u0025\u0064\u0020\u003e\u0020%\u0064\u0029",_cd ,len (buf ));
return buf ,_a .Errorf ("\u0069n\u0076a\u006c\u0069\u0064\u0020\u0070a\u0064\u0020l\u0065\u006e\u0067\u0074\u0068");};buf =buf [:len (buf )-_cd ];return buf ,nil ;};

// PDFVersion implements Filter interface.
func (filterAESV2 )PDFVersion ()[2]int {return [2]int {1,5}};func _aef (_afe FilterDict )(Filter ,error ){if _afe .Length %8!=0{return nil ,_a .Errorf ("\u0063\u0072\u0079p\u0074\u0020\u0066\u0069\u006c\u0074\u0065\u0072\u0020\u006c\u0065\u006e\u0067\u0074\u0068\u0020\u006e\u006f\u0074\u0020\u006d\u0075\u006c\u0074\u0069\u0070\u006c\u0065\u0020o\u0066\u0020\u0038\u0020\u0028\u0025\u0064\u0029",_afe .Length );
};if _afe .Length < 5||_afe .Length > 16{if _afe .Length ==40||_afe .Length ==64||_afe .Length ==128{_d .Log .Debug ("\u0053\u0054\u0041\u004e\u0044AR\u0044\u0020V\u0049\u004f\u004c\u0041\u0054\u0049\u004f\u004e\u003a\u0020\u0043\u0072\u0079\u0070\u0074\u0020\u004c\u0065\u006e\u0067\u0074\u0068\u0020\u0061\u0070\u0070\u0065\u0061\u0072s\u0020\u0074\u006f \u0062\u0065\u0020\u0069\u006e\u0020\u0062\u0069\u0074\u0073\u0020\u0072\u0061t\u0068\u0065\u0072\u0020\u0074h\u0061\u006e\u0020\u0062\u0079\u0074\u0065\u0073\u0020-\u0020\u0061s\u0073u\u006d\u0069\u006e\u0067\u0020\u0062\u0069t\u0073\u0020\u0028\u0025\u0064\u0029",_afe .Length );
_afe .Length /=8;}else {return nil ,_a .Errorf ("\u0063\u0072\u0079\u0070\u0074\u0020\u0066\u0069\u006c\u0074\u0065\u0072\u0020\u006c\u0065\u006e\u0067\u0074h\u0020\u006e\u006f\u0074\u0020\u0069\u006e \u0072\u0061\u006e\u0067\u0065\u0020\u0034\u0030\u0020\u002d\u00201\u0032\u0038\u0020\u0062\u0069\u0074\u0020\u0028\u0025\u0064\u0029",_afe .Length );
};};return filterV2 {_eg :_afe .Length },nil ;};