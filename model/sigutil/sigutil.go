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

package sigutil ;import (_b "bytes";_a "crypto";_c "crypto/x509";_fg "encoding/asn1";_df "encoding/pem";_f "errors";_gg "fmt";_gf "github.com/unidoc/timestamp";_be "github.com/unidoc/unipdf/v3/common";_dd "golang.org/x/crypto/ocsp";_ff "io";_d "net/http";
_e "time";);

// IsCA returns true if the provided certificate appears to be a CA certificate.
func (_cg *CertClient )IsCA (cert *_c .Certificate )bool {return cert .IsCA &&_b .Equal (cert .RawIssuer ,cert .RawSubject );};

// NewCertClient returns a new certificate client.
func NewCertClient ()*CertClient {return &CertClient {HTTPClient :_db ()}};func _db ()*_d .Client {return &_d .Client {Timeout :5*_e .Second }};

// NewCRLClient returns a new CRL client.
func NewCRLClient ()*CRLClient {return &CRLClient {HTTPClient :_db ()}};

// OCSPClient represents a OCSP (Online Certificate Status Protocol) client.
// It is used to request revocation data from OCSP servers.
type OCSPClient struct{

// HTTPClient is the HTTP client used to make OCSP requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_d .Client ;

// Hash is the hash function  used when constructing the OCSP
// requests. If zero, SHA-1 will be used.
Hash _a .Hash ;};

// GetIssuer retrieves the issuer of the provided certificate.
func (_fe *CertClient )GetIssuer (cert *_c .Certificate )(*_c .Certificate ,error ){for _ ,_bee :=range cert .IssuingCertificateURL {_ee ,_fgc :=_fe .Get (_bee );if _fgc !=nil {_be .Log .Debug ("\u0057\u0041\u0052\u004e\u003a\u0020\u0063\u006f\u0075\u006c\u0064\u0020\u006e\u006f\u0074 \u0064\u006f\u0077\u006e\u006c\u006f\u0061\u0064\u0020\u0069\u0073\u0073\u0075e\u0072\u0020\u0066\u006f\u0072\u0020\u0063\u0065\u0072\u0074\u0069\u0066ic\u0061\u0074\u0065\u0020\u0025\u0076\u003a\u0020\u0025\u0076",cert .Subject .CommonName ,_fgc );
continue ;};return _ee ,nil ;};return nil ,_gg .Errorf ("\u0069\u0073\u0073\u0075e\u0072\u0020\u0063\u0065\u0072\u0074\u0069\u0066\u0069\u0063a\u0074e\u0020\u006e\u006f\u0074\u0020\u0066\u006fu\u006e\u0064");};

// NewOCSPClient returns a new OCSP client.
func NewOCSPClient ()*OCSPClient {return &OCSPClient {HTTPClient :_db (),Hash :_a .SHA1 }};

// CertClient represents a X.509 certificate client. Its primary purpose
// is to download certificates.
type CertClient struct{

// HTTPClient is the HTTP client used to make certificate requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_d .Client ;};

// MakeRequest makes a CRL request to the specified server and returns the
// response. If a server URL is not provided, it is extracted from the certificate.
func (_aa *CRLClient )MakeRequest (serverURL string ,cert *_c .Certificate )([]byte ,error ){if _aa .HTTPClient ==nil {_aa .HTTPClient =_db ();};if serverURL ==""{if len (cert .CRLDistributionPoints )==0{return nil ,_f .New ("\u0063e\u0072\u0074i\u0066\u0069\u0063\u0061t\u0065\u0020\u0064o\u0065\u0073\u0020\u006e\u006f\u0074\u0020\u0073\u0070ec\u0069\u0066\u0079 \u0061\u006ey\u0020\u0043\u0052\u004c\u0020\u0073e\u0072\u0076e\u0072\u0073");
};serverURL =cert .CRLDistributionPoints [0];};_bg ,_ce :=_aa .HTTPClient .Get (serverURL );if _ce !=nil {return nil ,_ce ;};defer _bg .Body .Close ();_gc ,_ce :=_ff .ReadAll (_bg .Body );if _ce !=nil {return nil ,_ce ;};if _cde ,_ :=_df .Decode (_gc );
_cde !=nil {_gc =_cde .Bytes ;};return _gc ,nil ;};

// GetEncodedToken executes the timestamp request and returns the DER encoded
// timestamp token bytes.
func (_cc *TimestampClient )GetEncodedToken (serverURL string ,req *_gf .Request )([]byte ,error ){if serverURL ==""{return nil ,_gg .Errorf ("\u006d\u0075\u0073\u0074\u0020\u0070r\u006f\u0076\u0069\u0064\u0065\u0020\u0074\u0069\u006d\u0065\u0073\u0074\u0061m\u0070\u0020\u0073\u0065\u0072\u0076\u0065r\u0020\u0055\u0052\u004c");
};if req ==nil {return nil ,_gg .Errorf ("\u0074\u0069\u006de\u0073\u0074\u0061\u006dp\u0020\u0072\u0065\u0071\u0075\u0065\u0073t\u0020\u0063\u0061\u006e\u006e\u006f\u0074\u0020\u0062\u0065\u0020\u006e\u0069\u006c");};_ccb ,_bgf :=req .Marshal ();if _bgf !=nil {return nil ,_bgf ;
};_ffc ,_bgf :=_d .NewRequest ("\u0050\u004f\u0053\u0054",serverURL ,_b .NewBuffer (_ccb ));if _bgf !=nil {return nil ,_bgf ;};_ffc .Header .Set ("\u0043\u006f\u006et\u0065\u006e\u0074\u002d\u0054\u0079\u0070\u0065","a\u0070\u0070\u006c\u0069\u0063\u0061t\u0069\u006f\u006e\u002f\u0074\u0069\u006d\u0065\u0073t\u0061\u006d\u0070-\u0071u\u0065\u0072\u0079");
if _cc .BeforeHTTPRequest !=nil {if _gb :=_cc .BeforeHTTPRequest (_ffc );_gb !=nil {return nil ,_gb ;};};_bec :=_cc .HTTPClient ;if _bec ==nil {_bec =_db ();};_cdg ,_bgf :=_bec .Do (_ffc );if _bgf !=nil {return nil ,_bgf ;};defer _cdg .Body .Close ();_fce ,_bgf :=_ff .ReadAll (_cdg .Body );
if _bgf !=nil {return nil ,_bgf ;};if _cdg .StatusCode !=_d .StatusOK {return nil ,_gg .Errorf ("\u0075\u006e\u0065x\u0070\u0065\u0063\u0074e\u0064\u0020\u0048\u0054\u0054\u0050\u0020s\u0074\u0061\u0074\u0075\u0073\u0020\u0063\u006f\u0064\u0065\u003a\u0020\u0025\u0064",_cdg .StatusCode );
};var _bf struct{Version _fg .RawValue ;Content _fg .RawValue ;};if _ ,_bgf =_fg .Unmarshal (_fce ,&_bf );_bgf !=nil {return nil ,_bgf ;};return _bf .Content .FullBytes ,nil ;};

// CRLClient represents a CRL (Certificate revocation list) client.
// It is used to request revocation data from CRL servers.
type CRLClient struct{

// HTTPClient is the HTTP client used to make CRL requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_d .Client ;};

// NewTimestampClient returns a new timestamp client.
func NewTimestampClient ()*TimestampClient {return &TimestampClient {HTTPClient :_db ()}};

// NewTimestampRequest returns a new timestamp request based
// on the specified options.
func NewTimestampRequest (body _ff .Reader ,opts *_gf .RequestOptions )(*_gf .Request ,error ){if opts ==nil {opts =&_gf .RequestOptions {};};if opts .Hash ==0{opts .Hash =_a .SHA256 ;};if !opts .Hash .Available (){return nil ,_c .ErrUnsupportedAlgorithm ;
};_gfe :=opts .Hash .New ();if _ ,_cba :=_ff .Copy (_gfe ,body );_cba !=nil {return nil ,_cba ;};return &_gf .Request {HashAlgorithm :opts .Hash ,HashedMessage :_gfe .Sum (nil ),Certificates :opts .Certificates ,TSAPolicyOID :opts .TSAPolicyOID ,Nonce :opts .Nonce },nil ;
};

// Get retrieves the certificate at the specified URL.
func (_cd *CertClient )Get (url string )(*_c .Certificate ,error ){if _cd .HTTPClient ==nil {_cd .HTTPClient =_db ();};_ba ,_dc :=_cd .HTTPClient .Get (url );if _dc !=nil {return nil ,_dc ;};defer _ba .Body .Close ();_ad ,_dc :=_ff .ReadAll (_ba .Body );
if _dc !=nil {return nil ,_dc ;};if _ec ,_ :=_df .Decode (_ad );_ec !=nil {_ad =_ec .Bytes ;};_gd ,_dc :=_c .ParseCertificate (_ad );if _dc !=nil {return nil ,_dc ;};return _gd ,nil ;};

// MakeRequest makes a OCSP request to the specified server and returns
// the parsed and raw responses. If a server URL is not provided, it is
// extracted from the certificate.
func (_eeg *OCSPClient )MakeRequest (serverURL string ,cert ,issuer *_c .Certificate )(*_dd .Response ,[]byte ,error ){if _eeg .HTTPClient ==nil {_eeg .HTTPClient =_db ();};if serverURL ==""{if len (cert .OCSPServer )==0{return nil ,nil ,_f .New ("\u0063e\u0072\u0074i\u0066\u0069\u0063a\u0074\u0065\u0020\u0064\u006f\u0065\u0073 \u006e\u006f\u0074\u0020\u0073\u0070e\u0063\u0069\u0066\u0079\u0020\u0061\u006e\u0079\u0020\u004f\u0043S\u0050\u0020\u0073\u0065\u0072\u0076\u0065\u0072\u0073");
};serverURL =cert .OCSPServer [0];};_fgd ,_ef :=_dd .CreateRequest (cert ,issuer ,&_dd .RequestOptions {Hash :_eeg .Hash });if _ef !=nil {return nil ,nil ,_ef ;};_fc ,_ef :=_eeg .HTTPClient .Post (serverURL ,"\u0061p\u0070\u006c\u0069\u0063\u0061\u0074\u0069\u006f\u006e\u002f\u006fc\u0073\u0070\u002d\u0072\u0065\u0071\u0075\u0065\u0073\u0074",_b .NewReader (_fgd ));
if _ef !=nil {return nil ,nil ,_ef ;};defer _fc .Body .Close ();_eb ,_ef :=_ff .ReadAll (_fc .Body );if _ef !=nil {return nil ,nil ,_ef ;};if _eef ,_ :=_df .Decode (_eb );_eef !=nil {_eb =_eef .Bytes ;};_ebd ,_ef :=_dd .ParseResponseForCert (_eb ,cert ,issuer );
if _ef !=nil {return nil ,nil ,_ef ;};return _ebd ,_eb ,nil ;};

// TimestampClient represents a RFC 3161 timestamp client.
// It is used to obtain signed tokens from timestamp authority servers.
type TimestampClient struct{

// HTTPClient is the HTTP client used to make timestamp requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_d .Client ;

// Callbacks.
BeforeHTTPRequest func (_ag *_d .Request )error ;};