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

package errors ;import (_f "fmt";_e "golang.org/x/xerrors";);var _ _e .Wrapper =(*processError )(nil );func Wrapf (err error ,processName ,message string ,arguments ...interface{})error {if _da ,_eg :=err .(*processError );_eg {_da ._c ="";};_ec :=_ad (_f .Sprintf (message ,arguments ...),processName );
_ec ._b =err ;return _ec ;};func (_ee *processError )Error ()string {var _g string ;if _ee ._c !=""{_g =_ee ._c ;};_g +="\u0050r\u006f\u0063\u0065\u0073\u0073\u003a "+_ee ._ef ;if _ee ._d !=""{_g +="\u0020\u004d\u0065\u0073\u0073\u0061\u0067\u0065\u003a\u0020"+_ee ._d ;
};if _ee ._b !=nil {_g +="\u002e\u0020"+_ee ._b .Error ();};return _g ;};func (_cc *processError )Unwrap ()error {return _cc ._b };type processError struct{_c string ;_ef string ;_d string ;_b error ;};func Wrap (err error ,processName ,message string )error {if _cf ,_deb :=err .(*processError );
_deb {_cf ._c ="";};_ca :=_ad (message ,processName );_ca ._b =err ;return _ca ;};func Errorf (processName ,message string ,arguments ...interface{})error {return _ad (_f .Sprintf (message ,arguments ...),processName );};func Error (processName ,message string )error {return _ad (message ,processName )};
func _ad (_fd ,_de string )*processError {return &processError {_c :"\u005b\u0055\u006e\u0069\u0050\u0044\u0046\u005d",_d :_fd ,_ef :_de };};