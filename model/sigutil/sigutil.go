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

package sigutil ;import (_ff "bytes";_bd "crypto";_a "crypto/x509";_fb "encoding/asn1";_fd "encoding/pem";_d "errors";_c "fmt";_bc "github.com/unidoc/timestamp";_g "github.com/unidoc/unipdf/v4/common";_bg "golang.org/x/crypto/ocsp";_bf "io";_df "net/http";
_b "time";);

// MakeRequest makes a OCSP request to the specified server and returns
// the parsed and raw responses. If a server URL is not provided, it is
// extracted from the certificate.
func (_ge *OCSPClient )MakeRequest (serverURL string ,cert ,issuer *_a .Certificate )(*_bg .Response ,[]byte ,error ){if _ge .HTTPClient ==nil {_ge .HTTPClient =_cb ();};if serverURL ==""{if len (cert .OCSPServer )==0{return nil ,nil ,_d .New ("\u0063e\u0072\u0074i\u0066\u0069\u0063a\u0074\u0065\u0020\u0064\u006f\u0065\u0073 \u006e\u006f\u0074\u0020\u0073\u0070e\u0063\u0069\u0066\u0079\u0020\u0061\u006e\u0079\u0020\u004f\u0043S\u0050\u0020\u0073\u0065\u0072\u0076\u0065\u0072\u0073");
};serverURL =cert .OCSPServer [0];};_dfc ,_eb :=_bg .CreateRequest (cert ,issuer ,&_bg .RequestOptions {Hash :_ge .Hash });if _eb !=nil {return nil ,nil ,_eb ;};_ae ,_eb :=_ge .HTTPClient .Post (serverURL ,"\u0061p\u0070\u006c\u0069\u0063\u0061\u0074\u0069\u006f\u006e\u002f\u006fc\u0073\u0070\u002d\u0072\u0065\u0071\u0075\u0065\u0073\u0074",_ff .NewReader (_dfc ));
if _eb !=nil {return nil ,nil ,_eb ;};defer _ae .Body .Close ();_cd ,_eb :=_bf .ReadAll (_ae .Body );if _eb !=nil {return nil ,nil ,_eb ;};if _ec ,_ :=_fd .Decode (_cd );_ec !=nil {_cd =_ec .Bytes ;};_gca ,_eb :=_bg .ParseResponseForCert (_cd ,cert ,issuer );
if _eb !=nil {return nil ,nil ,_eb ;};return _gca ,_cd ,nil ;};

// CertClient represents a X.509 certificate client. Its primary purpose
// is to download certificates.
type CertClient struct{

// HTTPClient is the HTTP client used to make certificate requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_df .Client ;};

// TimestampClient represents a RFC 3161 timestamp client.
// It is used to obtain signed tokens from timestamp authority servers.
type TimestampClient struct{

// HTTPClient is the HTTP client used to make timestamp requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_df .Client ;

// Callbacks.
BeforeHTTPRequest func (_ebc *_df .Request )error ;};

// NewTimestampClient returns a new timestamp client.
func NewTimestampClient ()*TimestampClient {return &TimestampClient {HTTPClient :_cb ()}};func _cb ()*_df .Client {return &_df .Client {Timeout :5*_b .Second }};

// CRLClient represents a CRL (Certificate revocation list) client.
// It is used to request revocation data from CRL servers.
type CRLClient struct{

// HTTPClient is the HTTP client used to make CRL requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_df .Client ;};

// NewCRLClient returns a new CRL client.
func NewCRLClient ()*CRLClient {return &CRLClient {HTTPClient :_cb ()}};

// MakeRequest makes a CRL request to the specified server and returns the
// response. If a server URL is not provided, it is extracted from the certificate.
func (_ac *CRLClient )MakeRequest (serverURL string ,cert *_a .Certificate )([]byte ,error ){if _ac .HTTPClient ==nil {_ac .HTTPClient =_cb ();};if serverURL ==""{if len (cert .CRLDistributionPoints )==0{return nil ,_d .New ("\u0063e\u0072\u0074i\u0066\u0069\u0063\u0061t\u0065\u0020\u0064o\u0065\u0073\u0020\u006e\u006f\u0074\u0020\u0073\u0070ec\u0069\u0066\u0079 \u0061\u006ey\u0020\u0043\u0052\u004c\u0020\u0073e\u0072\u0076e\u0072\u0073");
};serverURL =cert .CRLDistributionPoints [0];};_daf ,_ed :=_ac .HTTPClient .Get (serverURL );if _ed !=nil {return nil ,_ed ;};defer _daf .Body .Close ();_bgb ,_ed :=_bf .ReadAll (_daf .Body );if _ed !=nil {return nil ,_ed ;};if _ca ,_ :=_fd .Decode (_bgb );
_ca !=nil {_bgb =_ca .Bytes ;};return _bgb ,nil ;};

// NewCertClient returns a new certificate client.
func NewCertClient ()*CertClient {return &CertClient {HTTPClient :_cb ()}};

// GetEncodedToken executes the timestamp request and returns the DER encoded
// timestamp token bytes.
func (_ffa *TimestampClient )GetEncodedToken (serverURL string ,req *_bc .Request )([]byte ,error ){if serverURL ==""{return nil ,_c .Errorf ("\u006d\u0075\u0073\u0074\u0020\u0070r\u006f\u0076\u0069\u0064\u0065\u0020\u0074\u0069\u006d\u0065\u0073\u0074\u0061m\u0070\u0020\u0073\u0065\u0072\u0076\u0065r\u0020\u0055\u0052\u004c");
};if req ==nil {return nil ,_c .Errorf ("\u0074\u0069\u006de\u0073\u0074\u0061\u006dp\u0020\u0072\u0065\u0071\u0075\u0065\u0073t\u0020\u0063\u0061\u006e\u006e\u006f\u0074\u0020\u0062\u0065\u0020\u006e\u0069\u006c");};_gd ,_de :=req .Marshal ();if _de !=nil {return nil ,_de ;
};_acd ,_de :=_df .NewRequest ("\u0050\u004f\u0053\u0054",serverURL ,_ff .NewBuffer (_gd ));if _de !=nil {return nil ,_de ;};_acd .Header .Set ("\u0043\u006f\u006et\u0065\u006e\u0074\u002d\u0054\u0079\u0070\u0065","a\u0070\u0070\u006c\u0069\u0063\u0061t\u0069\u006f\u006e\u002f\u0074\u0069\u006d\u0065\u0073t\u0061\u006d\u0070-\u0071u\u0065\u0072\u0079");
if _ffa .BeforeHTTPRequest !=nil {if _bbe :=_ffa .BeforeHTTPRequest (_acd );_bbe !=nil {return nil ,_bbe ;};};_gcb :=_ffa .HTTPClient ;if _gcb ==nil {_gcb =_cb ();};_gg ,_de :=_gcb .Do (_acd );if _de !=nil {return nil ,_de ;};defer _gg .Body .Close ();
_aee ,_de :=_bf .ReadAll (_gg .Body );if _de !=nil {return nil ,_de ;};if _gg .StatusCode !=_df .StatusOK {return nil ,_c .Errorf ("\u0075\u006e\u0065x\u0070\u0065\u0063\u0074e\u0064\u0020\u0048\u0054\u0054\u0050\u0020s\u0074\u0061\u0074\u0075\u0073\u0020\u0063\u006f\u0064\u0065\u003a\u0020\u0025\u0064",_gg .StatusCode );
};var _bad struct{Version _fb .RawValue ;Content _fb .RawValue ;};if _ ,_de =_fb .Unmarshal (_aee ,&_bad );_de !=nil {return nil ,_de ;};return _bad .Content .FullBytes ,nil ;};

// OCSPClient represents a OCSP (Online Certificate Status Protocol) client.
// It is used to request revocation data from OCSP servers.
type OCSPClient struct{

// HTTPClient is the HTTP client used to make OCSP requests.
// By default, an HTTP client with a 5 second timeout per request is used.
HTTPClient *_df .Client ;

// Hash is the hash function  used when constructing the OCSP
// requests. If zero, SHA-1 will be used.
Hash _bd .Hash ;};

// NewOCSPClient returns a new OCSP client.
func NewOCSPClient ()*OCSPClient {return &OCSPClient {HTTPClient :_cb (),Hash :_bd .SHA1 }};

// GetIssuer retrieves the issuer of the provided certificate.
func (_bb *CertClient )GetIssuer (cert *_a .Certificate )(*_a .Certificate ,error ){for _ ,_ffg :=range cert .IssuingCertificateURL {_fgc ,_dbd :=_bb .Get (_ffg );if _dbd !=nil {_g .Log .Debug ("\u0057\u0041\u0052\u004e\u003a\u0020\u0063\u006f\u0075\u006c\u0064\u0020\u006e\u006f\u0074 \u0064\u006f\u0077\u006e\u006c\u006f\u0061\u0064\u0020\u0069\u0073\u0073\u0075e\u0072\u0020\u0066\u006f\u0072\u0020\u0063\u0065\u0072\u0074\u0069\u0066ic\u0061\u0074\u0065\u0020\u0025\u0076\u003a\u0020\u0025\u0076",cert .Subject .CommonName ,_dbd );
continue ;};return _fgc ,nil ;};return nil ,_c .Errorf ("\u0069\u0073\u0073\u0075e\u0072\u0020\u0063\u0065\u0072\u0074\u0069\u0066\u0069\u0063a\u0074e\u0020\u006e\u006f\u0074\u0020\u0066\u006fu\u006e\u0064");};

// IsCA returns true if the provided certificate appears to be a CA certificate.
func (_gc *CertClient )IsCA (cert *_a .Certificate )bool {return cert .IsCA &&_ff .Equal (cert .RawIssuer ,cert .RawSubject );};

// Get retrieves the certificate at the specified URL.
func (_fg *CertClient )Get (url string )(*_a .Certificate ,error ){if _fg .HTTPClient ==nil {_fg .HTTPClient =_cb ();};_e ,_da :=_fg .HTTPClient .Get (url );if _da !=nil {return nil ,_da ;};defer _e .Body .Close ();_fdf ,_da :=_bf .ReadAll (_e .Body );
if _da !=nil {return nil ,_da ;};if _ffe ,_ :=_fd .Decode (_fdf );_ffe !=nil {_fdf =_ffe .Bytes ;};_db ,_da :=_a .ParseCertificate (_fdf );if _da !=nil {return nil ,_da ;};return _db ,nil ;};

// NewTimestampRequest returns a new timestamp request based
// on the specified options.
func NewTimestampRequest (body _bf .Reader ,opts *_bc .RequestOptions )(*_bc .Request ,error ){if opts ==nil {opts =&_bc .RequestOptions {};};if opts .Hash ==0{opts .Hash =_bd .SHA256 ;};if !opts .Hash .Available (){return nil ,_a .ErrUnsupportedAlgorithm ;
};_dg :=opts .Hash .New ();if _ ,_ffef :=_bf .Copy (_dg ,body );_ffef !=nil {return nil ,_ffef ;};return &_bc .Request {HashAlgorithm :opts .Hash ,HashedMessage :_dg .Sum (nil ),Certificates :opts .Certificates ,TSAPolicyOID :opts .TSAPolicyOID ,Nonce :opts .Nonce },nil ;
};