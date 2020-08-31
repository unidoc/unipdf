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

package optimize ;import (_a "bytes";_b "crypto/md5";_bd "errors";_f "fmt";_ecd "github.com/unidoc/unipdf/v3/common";_gg "github.com/unidoc/unipdf/v3/contentstream";_ce "github.com/unidoc/unipdf/v3/core";_ec "github.com/unidoc/unipdf/v3/extractor";_gc "github.com/unidoc/unipdf/v3/internal/textencoding";_d "github.com/unidoc/unipdf/v3/model";_ecf "github.com/unidoc/unitype";_e "golang.org/x/image/draw";_g "image";_ee "math";);

// Append appends optimizers to the chain.
func (_gga *Chain )Append (optimizers ..._d .Optimizer ){_gga ._bdc =append (_gga ._bdc ,optimizers ...)};func _dce (_egd []_ce .PdfObject )[]*imageInfo {_dfb :=_ce .PdfObjectName ("\u0053u\u0062\u0074\u0079\u0070\u0065");_edfc :=make (map[*_ce .PdfObjectStream ]struct{});var _cada error ;var _dde []*imageInfo ;for _ ,_gea :=range _egd {_baeg ,_ggab :=_ce .GetStream (_gea );if !_ggab {continue ;};if _ ,_gbc :=_edfc [_baeg ];_gbc {continue ;};_edfc [_baeg ]=struct{}{};_aeg :=_baeg .PdfObjectDictionary .Get (_dfb );_aae ,_ggab :=_ce .GetName (_aeg );if !_ggab ||string (*_aae )!="\u0049\u006d\u0061g\u0065"{continue ;};_eega :=&imageInfo {BitsPerComponent :8,Stream :_baeg };if _eega .ColorSpace ,_cada =_d .DetermineColorspaceNameFromPdfObject (_baeg .PdfObjectDictionary .Get ("\u0043\u006f\u006c\u006f\u0072\u0053\u0070\u0061\u0063\u0065"));_cada !=nil {_ecd .Log .Error ("\u0045\u0072\u0072\u006f\u0072\u0020\u0064\u0065\u0074\u0065r\u006d\u0069\u006e\u0065\u0020\u0063\u006fl\u006f\u0072\u0020\u0073\u0070\u0061\u0063\u0065\u0020\u0025\u0073",_cada );continue ;};if _acba ,_ebg :=_ce .GetIntVal (_baeg .PdfObjectDictionary .Get ("\u0042\u0069t\u0073\u0050\u0065r\u0043\u006f\u006d\u0070\u006f\u006e\u0065\u006e\u0074"));_ebg {_eega .BitsPerComponent =_acba ;};if _cfe ,_gaea :=_ce .GetIntVal (_baeg .PdfObjectDictionary .Get ("\u0057\u0069\u0064t\u0068"));_gaea {_eega .Width =_cfe ;};if _cefe ,_badb :=_ce .GetIntVal (_baeg .PdfObjectDictionary .Get ("\u0048\u0065\u0069\u0067\u0068\u0074"));_badb {_eega .Height =_cefe ;};switch _eega .ColorSpace {case "\u0044e\u0076\u0069\u0063\u0065\u0052\u0047B":_eega .ColorComponents =3;case "\u0044\u0065\u0076\u0069\u0063\u0065\u0047\u0072\u0061\u0079":_eega .ColorComponents =1;default:_ecd .Log .Warning ("\u004f\u0070\u0074\u0069\u006d\u0069\u007a\u0061t\u0069\u006f\u006e i\u0073\u0020\u006e\u006f\u0074\u0020s\u0075\u0070\u0070\u006f\u0072\u0074\u0065\u0064\u0020\u0066\u006f\u0072\u0020\u0063\u006fl\u006f\u0072\u0020\u0073\u0070\u0061\u0063\u0065 \u0025\u0073",_eega .ColorSpace );continue ;};_dde =append (_dde ,_eega );};return _dde ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_caab *CombineIdenticalIndirectObjects )Optimize (objects []_ce .PdfObject )(_fgfc []_ce .PdfObject ,_aaf error ){_gedg (objects );_adf :=make (map[_ce .PdfObject ]_ce .PdfObject );_gfg :=make (map[_ce .PdfObject ]struct{});_dbf :=make (map[string ][]*_ce .PdfIndirectObject );for _ ,_ebb :=range objects {_edg ,_dgge :=_ebb .(*_ce .PdfIndirectObject );if !_dgge {continue ;};if _fef ,_cad :=_edg .PdfObject .(*_ce .PdfObjectDictionary );_cad {if _gcge ,_fbc :=_fef .Get ("\u0054\u0079\u0070\u0065").(*_ce .PdfObjectName );_fbc &&*_gcge =="\u0050\u0061\u0067\u0065"{continue ;};_dagf :=_b .New ();_dagf .Write ([]byte (_fef .WriteString ()));_gde :=string (_dagf .Sum (nil ));_dbf [_gde ]=append (_dbf [_gde ],_edg );};};for _ ,_afcg :=range _dbf {if len (_afcg )< 2{continue ;};_gaa :=_afcg [0];for _fcd :=1;_fcd < len (_afcg );_fcd ++{_gcd :=_afcg [_fcd ];_adf [_gcd ]=_gaa ;_gfg [_gcd ]=struct{}{};};};_fgfc =make ([]_ce .PdfObject ,0,len (objects )-len (_gfg ));for _ ,_bgee :=range objects {if _ ,_aab :=_gfg [_bgee ];_aab {continue ;};_fgfc =append (_fgfc ,_bgee );};_bcbd (_fgfc ,_adf );return _fgfc ,nil ;};func _cabd (_eafc []_ce .PdfObject )objectStructure {_cce :=objectStructure {};_ebf :=false ;for _ ,_ebfa :=range _eafc {switch _acd :=_ebfa .(type ){case *_ce .PdfIndirectObject :_bcef ,_ffac :=_ce .GetDict (_acd );if !_ffac {continue ;};_bdce ,_ffac :=_ce .GetName (_bcef .Get ("\u0054\u0079\u0070\u0065"));if !_ffac {continue ;};switch _bdce .String (){case "\u0043a\u0074\u0061\u006c\u006f\u0067":_cce ._bbga =_bcef ;_ebf =true ;};};if _ebf {break ;};};if !_ebf {return _cce ;};_fed ,_baec :=_ce .GetDict (_cce ._bbga .Get ("\u0050\u0061\u0067e\u0073"));if !_baec {return _cce ;};_cce ._ffgdb =_fed ;_cba ,_baec :=_ce .GetArray (_fed .Get ("\u004b\u0069\u0064\u0073"));if !_baec {return _cce ;};for _ ,_fae :=range _cba .Elements (){_fad ,_dbg :=_ce .GetIndirect (_fae );if !_dbg {break ;};_cce ._cfeb =append (_cce ._cfeb ,_fad );};return _cce ;};func _ca (_ddf *_gg .ContentStreamOperations )*_gg .ContentStreamOperations {if _ddf ==nil {return nil ;};_cc :=_gg .ContentStreamOperations {};for _ ,_ccb :=range *_ddf {switch _ccb .Operand {case "\u0042\u0044\u0043","\u0042\u004d\u0043","\u0045\u004d\u0043":continue ;case "\u0054\u006d":if len (_ccb .Params )==6{if _ae ,_ad :=_ce .GetNumbersAsFloat (_ccb .Params );_ad ==nil {if _ae [0]==1&&_ae [1]==0&&_ae [2]==0&&_ae [3]==1{_ccb =&_gg .ContentStreamOperation {Params :[]_ce .PdfObject {_ccb .Params [4],_ccb .Params [5]},Operand :"\u0054\u0064"};};};};};_cc =append (_cc ,_ccb );};return &_cc ;};

// CombineDuplicateDirectObjects combines duplicated direct objects by its data hash.
// It implements interface model.Optimizer.
type CombineDuplicateDirectObjects struct{};

// CleanFonts cleans up embedded fonts, reducing font sizes.
type CleanFonts struct{

// Subset embedded fonts if encountered (if true).
// Otherwise attempts to reduce the font program.
Subset bool ;};func _eg (_ead *_ce .PdfObjectStream ,_gdg []rune ,_dga []_ecf .GlyphIndex )error {_ead ,_egg :=_ce .GetStream (_ead );if !_egg {_ecd .Log .Debug ("\u0045\u006d\u0062\u0065\u0064\u0064\u0065\u0064\u0020\u0066\u006f\u006e\u0074\u0020\u006f\u0062\u006a\u0065c\u0074\u0020\u006e\u006f\u0074\u0020\u0066o\u0075\u006e\u0064\u0020\u002d\u002d\u0020\u0041\u0042\u004f\u0052T\u0020\u0073\u0075\u0062\u0073\u0065\u0074\u0074\u0069\u006e\u0067");return _bd .New ("\u0066\u006f\u006e\u0074fi\u006c\u0065\u0032\u0020\u006e\u006f\u0074\u0020\u0066\u006f\u0075\u006e\u0064");};_fac ,_edd :=_ce .DecodeStream (_ead );if _edd !=nil {_ecd .Log .Debug ("\u0044\u0065c\u006f\u0064\u0065 \u0065\u0072\u0072\u006f\u0072\u003a\u0020\u0025\u0076",_edd );return _edd ;};_deg ,_edd :=_ecf .Parse (_a .NewReader (_fac ));if _edd !=nil {_ecd .Log .Debug ("\u0045\u0072\u0072\u006f\u0072\u0020\u0070\u0061\u0072\u0073\u0069n\u0067\u0020\u0025\u0064\u0020\u0062\u0079\u0074\u0065\u0020f\u006f\u006e\u0074",len (_ead .Stream ));return _edd ;};_cgg :=_dga ;if len (_gdg )> 0{_ffa :=_deg .LookupRunes (_gdg );_cgg =append (_cgg ,_ffa ...);};_deg ,_edd =_deg .SubsetKeepIndices (_cgg );if _edd !=nil {_ecd .Log .Debug ("\u0045R\u0052\u004f\u0052\u0020s\u0075\u0062\u0073\u0065\u0074t\u0069n\u0067 \u0066\u006f\u006e\u0074\u003a\u0020\u0025v",_edd );return _edd ;};var _ddbd _a .Buffer ;_edd =_deg .Write (&_ddbd );if _edd !=nil {_ecd .Log .Debug ("\u0045\u0052\u0052\u004fR \u0057\u0072\u0069\u0074\u0069\u006e\u0067\u0020\u0066\u006f\u006e\u0074\u003a\u0020%\u0076",_edd );return _edd ;};if _ddbd .Len ()> len (_fac ){_ecd .Log .Debug ("\u0052\u0065-\u0077\u0072\u0069\u0074\u0074\u0065\u006e\u0020\u0066\u006f\u006e\u0074\u0020\u0069\u0073\u0020\u006c\u0061\u0072\u0067\u0065\u0072\u0020\u0074\u0068\u0061\u006e\u0020\u006f\u0072\u0069\u0067\u0069\u006e\u0061\u006c\u0020\u002d\u0020\u0073\u006b\u0069\u0070");return nil ;};_bec ,_edd :=_ce .MakeStream (_ddbd .Bytes (),_ce .NewFlateEncoder ());if _edd !=nil {_ecd .Log .Debug ("\u0045\u0052\u0052\u004fR \u0057\u0072\u0069\u0074\u0069\u006e\u0067\u0020\u0066\u006f\u006e\u0074\u003a\u0020%\u0076",_edd );return _edd ;};*_ead =*_bec ;_ead .Set ("\u004ce\u006e\u0067\u0074\u0068\u0031",_ce .MakeInteger (int64 (_ddbd .Len ())));return nil ;};

// CleanContentstream cleans up redundant operands in content streams, including Page and XObject Form
// contents. This process includes:
// 1. Marked content operators are removed.
// 2. Some operands are simplified (shorter form).
// TODO: Add more reduction methods and improving the methods for identifying unnecessary operands.
type CleanContentstream struct{};func _ed (_ge *_ce .PdfObjectStream )error {_gfc ,_fb :=_ce .DecodeStream (_ge );if _fb !=nil {return _fb ;};_adb :=_gg .NewContentStreamParser (string (_gfc ));_cea ,_fb :=_adb .Parse ();if _fb !=nil {return _fb ;};_cea =_ca (_cea );_bc :=_cea .Bytes ();if len (_bc )>=len (_gfc ){return nil ;};_gd ,_fb :=_ce .MakeStream (_cea .Bytes (),_ce .NewFlateEncoder ());if _fb !=nil {return _fb ;};_ge .Stream =_gd .Stream ;_ge .Merge (_gd .PdfObjectDictionary );return nil ;};

// CombineIdenticalIndirectObjects combines identical indirect objects.
// It implements interface model.Optimizer.
type CombineIdenticalIndirectObjects struct{};func _eba (_agb _ce .PdfObject )(_acbb string ,_gbd []_ce .PdfObject ){var _cbff _a .Buffer ;switch _aaed :=_agb .(type ){case *_ce .PdfIndirectObject :_gbd =append (_gbd ,_aaed );_agb =_aaed .PdfObject ;};switch _eegg :=_agb .(type ){case *_ce .PdfObjectStream :if _fcf ,_bgef :=_ce .DecodeStream (_eegg );_bgef ==nil {_cbff .Write (_fcf );_gbd =append (_gbd ,_eegg );};case *_ce .PdfObjectArray :for _ ,_ecbg :=range _eegg .Elements (){switch _gbf :=_ecbg .(type ){case *_ce .PdfObjectStream :if _gdf ,_daa :=_ce .DecodeStream (_gbf );_daa ==nil {_cbff .Write (_gdf );_gbd =append (_gbd ,_gbf );};};};};return _cbff .String (),_gbd ;};

// New creates a optimizers chain from options.
func New (options Options )*Chain {_efb :=new (Chain );if options .CleanFonts ||options .SubsetFonts {_efb .Append (&CleanFonts {Subset :options .SubsetFonts });};if options .CleanContentstream {_efb .Append (new (CleanContentstream ));};if options .ImageUpperPPI > 0{_eaa :=new (ImagePPI );_eaa .ImageUpperPPI =options .ImageUpperPPI ;_efb .Append (_eaa );};if options .ImageQuality > 0{_gafe :=new (Image );_gafe .ImageQuality =options .ImageQuality ;_efb .Append (_gafe );};if options .CombineDuplicateDirectObjects {_efb .Append (new (CombineDuplicateDirectObjects ));};if options .CombineDuplicateStreams {_efb .Append (new (CombineDuplicateStreams ));};if options .CombineIdenticalIndirectObjects {_efb .Append (new (CombineIdenticalIndirectObjects ));};if options .UseObjectStreams {_efb .Append (new (ObjectStreams ));};if options .CompressStreams {_efb .Append (new (CompressStreams ));};return _efb ;};

// CombineDuplicateStreams combines duplicated streams by its data hash.
// It implements interface model.Optimizer.
type CombineDuplicateStreams struct{};type imageInfo struct{ColorSpace _ce .PdfObjectName ;BitsPerComponent int ;ColorComponents int ;Width int ;Height int ;Stream *_ce .PdfObjectStream ;PPI float64 ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_abaf *CombineDuplicateStreams )Optimize (objects []_ce .PdfObject )(_eeg []_ce .PdfObject ,_bdg error ){_eggc :=make (map[_ce .PdfObject ]_ce .PdfObject );_gdb :=make (map[_ce .PdfObject ]struct{});_cca :=make (map[string ][]*_ce .PdfObjectStream );for _ ,_fce :=range objects {if _abc ,_cffe :=_fce .(*_ce .PdfObjectStream );_cffe {_cbg :=_b .New ();_cbg .Write ([]byte (_abc .Stream ));_ceb :=string (_cbg .Sum (nil ));_cca [_ceb ]=append (_cca [_ceb ],_abc );};};for _ ,_cgd :=range _cca {if len (_cgd )< 2{continue ;};_egb :=_cgd [0];for _abe :=1;_abe < len (_cgd );_abe ++{_gca :=_cgd [_abe ];_eggc [_gca ]=_egb ;_gdb [_gca ]=struct{}{};};};_eeg =make ([]_ce .PdfObject ,0,len (objects )-len (_gdb ));for _ ,_fca :=range objects {if _ ,_dgcg :=_gdb [_fca ];_dgcg {continue ;};_eeg =append (_eeg ,_fca );};_bcbd (_eeg ,_eggc );return _eeg ,nil ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_dgg *CleanFonts )Optimize (objects []_ce .PdfObject )(_bbb []_ce .PdfObject ,_bggc error ){var _edb map[*_ce .PdfObjectStream ]struct{};if _dgg .Subset {var _bba error ;_edb ,_bba =_cg (objects );if _bba !=nil {return nil ,_bba ;};};for _ ,_fc :=range objects {_dcf ,_fge :=_ce .GetStream (_fc );if !_fge {continue ;};if _ ,_dggc :=_edb [_dcf ];_dggc {continue ;};_gae ,_ada :=_ce .NewEncoderFromStream (_dcf );if _ada !=nil {_ecd .Log .Debug ("\u0045\u0052RO\u0052\u0020\u0067e\u0074\u0074\u0069\u006eg e\u006eco\u0064\u0065\u0072\u003a\u0020\u0025\u0076 -\u0020\u0069\u0067\u006e\u006f\u0072\u0069n\u0067",_ada );continue ;};_agaf ,_ada :=_gae .DecodeStream (_dcf );if _ada !=nil {_ecd .Log .Debug ("\u0044\u0065\u0063\u006f\u0064\u0069\u006e\u0067\u0020\u0065r\u0072\u006f\u0072\u0020\u003a\u0020\u0025v\u0020\u002d\u0020\u0069\u0067\u006e\u006f\u0072\u0069\u006e\u0067",_ada );continue ;};if len (_agaf )< 4{continue ;};_edc :=string (_agaf [:4]);if _edc =="\u004f\u0054\u0054\u004f"{continue ;};if _edc !="\u0000\u0001\u0000\u0000"&&_edc !="\u0074\u0072\u0075\u0065"{continue ;};_gfd ,_ada :=_ecf .Parse (_a .NewReader (_agaf ));if _ada !=nil {_ecd .Log .Debug ("\u0045\u0052\u0052\u004f\u0052\u0020P\u0061\u0072\u0073\u0069\u006e\u0067\u0020\u0066\u006f\u006e\u0074\u003a\u0020%\u0076\u0020\u002d\u0020\u0069\u0067\u006eo\u0072\u0069\u006e\u0067",_ada );continue ;};_ada =_gfd .Optimize ();if _ada !=nil {continue ;};var _cfb _a .Buffer ;_ada =_gfd .Write (&_cfb );if _ada !=nil {_ecd .Log .Debug ("\u0045\u0052\u0052\u004f\u0052\u0020W\u0072\u0069\u0074\u0069\u006e\u0067\u0020\u0066\u006f\u006e\u0074\u003a\u0020%\u0076\u0020\u002d\u0020\u0069\u0067\u006eo\u0072\u0069\u006e\u0067",_ada );continue ;};if _cfb .Len ()> len (_agaf ){_ecd .Log .Debug ("\u0052\u0065-\u0077\u0072\u0069\u0074\u0074\u0065\u006e\u0020\u0066\u006f\u006e\u0074\u0020\u0069\u0073\u0020\u006c\u0061\u0072\u0067\u0065\u0072\u0020\u0074\u0068\u0061\u006e\u0020\u006f\u0072\u0069\u0067\u0069\u006e\u0061\u006c\u0020\u002d\u0020\u0073\u006b\u0069\u0070");continue ;};_aca ,_ada :=_ce .MakeStream (_cfb .Bytes (),_ce .NewFlateEncoder ());if _ada !=nil {continue ;};*_dcf =*_aca ;_dcf .Set ("\u004ce\u006e\u0067\u0074\u0068\u0031",_ce .MakeInteger (int64 (_cfb .Len ())));};return objects ,nil ;};

// ObjectStreams groups PDF objects to object streams.
// It implements interface model.Optimizer.
type ObjectStreams struct{};

// CompressStreams compresses uncompressed streams.
// It implements interface model.Optimizer.
type CompressStreams struct{};func _gedg (_ggaf []_ce .PdfObject ){for _gafc ,_afb :=range _ggaf {switch _eed :=_afb .(type ){case *_ce .PdfIndirectObject :_eed .ObjectNumber =int64 (_gafc +1);_eed .GenerationNumber =0;case *_ce .PdfObjectStream :_eed .ObjectNumber =int64 (_gafc +1);_eed .GenerationNumber =0;case *_ce .PdfObjectStreams :_eed .ObjectNumber =int64 (_gafc +1);_eed .GenerationNumber =0;};};};

// Chain allows to use sequence of optimizers.
// It implements interface model.Optimizer.
type Chain struct{_bdc []_d .Optimizer };func _gccd (_cfg *_ce .PdfObjectStream ,_cdd float64 )error {_dab ,_gef :=_d .NewXObjectImageFromStream (_cfg );if _gef !=nil {return _gef ;};_ggg ,_gef :=_dab .ToImage ();if _gef !=nil {return _gef ;};_abd ,_gef :=_ggg .ToGoImage ();if _gef !=nil {return _gef ;};_acga :=int (_ee .RoundToEven (float64 (_ggg .Width )*_cdd ));_dfg :=int (_ee .RoundToEven (float64 (_ggg .Height )*_cdd ));_dcg :=_g .Rect (0,0,_acga ,_dfg );var _bdab _e .Image ;var _cag func (_g .Image )(*_d .Image ,error );switch _dab .ColorSpace .String (){case "\u0044e\u0076\u0069\u0063\u0065\u0052\u0047B":_bdab =_g .NewRGBA (_dcg );_cag =_d .ImageHandling .NewImageFromGoImage ;case "\u0044\u0065\u0076\u0069\u0063\u0065\u0047\u0072\u0061\u0079":_bdab =_g .NewGray (_dcg );_cag =_d .ImageHandling .NewGrayImageFromGoImage ;default:return _f .Errorf ("\u006f\u0070\u0074\u0069\u006d\u0069\u007a\u0061t\u0069\u006f\u006e i\u0073\u0020\u006e\u006f\u0074\u0020s\u0075\u0070\u0070\u006f\u0072\u0074\u0065\u0064\u0020\u0066\u006f\u0072\u0020\u0063\u006fl\u006f\u0072\u0020\u0073\u0070\u0061\u0063\u0065 \u0025\u0073",_dab .ColorSpace .String ());};_e .CatmullRom .Scale (_bdab ,_bdab .Bounds (),_abd ,_abd .Bounds (),_e .Over ,&_e .Options {});if _ggg ,_gef =_cag (_bdab );_gef !=nil {return _gef ;};_ded :=_ce .MakeDict ();_ded .Set ("\u0051u\u0061\u006c\u0069\u0074\u0079",_ce .MakeInteger (100));_ded .Set ("\u0050r\u0065\u0064\u0069\u0063\u0074\u006fr",_ce .MakeInteger (1));_dab .Filter .UpdateParams (_ded );if _gef =_dab .SetImage (_ggg ,nil );_gef !=nil {return _gef ;};_dab .ToPdfObject ();return nil ;};

// ImagePPI optimizes images by scaling images such that the PPI (pixels per inch) is never higher than ImageUpperPPI.
// TODO(a5i): Add support for inline images.
// It implements interface model.Optimizer.
type ImagePPI struct{ImageUpperPPI float64 ;};func _bdd (_gfca _ce .PdfObject )[]content {if _gfca ==nil {return nil ;};_fbdb ,_geg :=_ce .GetArray (_gfca );if !_geg {_ecd .Log .Debug ("\u0041\u006e\u006e\u006fts\u0020\u006e\u006f\u0074\u0020\u0061\u006e\u0020\u0061\u0072\u0072\u0061\u0079");return nil ;};var _fbg []content ;for _ ,_gge :=range _fbdb .Elements (){_eeca ,_aa :=_ce .GetDict (_gge );if !_aa {_ecd .Log .Debug ("I\u0067\u006e\u006f\u0072\u0069\u006eg\u0020\u006e\u006f\u006e\u002d\u0064i\u0063\u0074\u0020\u0065\u006c\u0065\u006de\u006e\u0074\u0020\u0069\u006e\u0020\u0041\u006e\u006e\u006ft\u0073");continue ;};_ggae ,_aa :=_ce .GetDict (_eeca .Get ("\u0041\u0050"));if !_aa {_ecd .Log .Debug ("\u004e\u006f\u0020\u0041P \u0065\u006e\u0074\u0072\u0079\u0020\u002d\u0020\u0073\u006b\u0069\u0070\u0070\u0069n\u0067");continue ;};_bcc :=_ce .TraceToDirectObject (_ggae .Get ("\u004e"));if _bcc ==nil {_ecd .Log .Debug ("N\u006f\u0020\u004e\u0020en\u0074r\u0079\u0020\u002d\u0020\u0073k\u0069\u0070\u0070\u0069\u006e\u0067");continue ;};var _eb *_ce .PdfObjectStream ;switch _cbe :=_bcc .(type ){case *_ce .PdfObjectDictionary :_gb ,_caad :=_ce .GetName (_eeca .Get ("\u0041\u0053"));if !_caad {_ecd .Log .Debug ("\u004e\u006f\u0020\u0041S \u0065\u006e\u0074\u0072\u0079\u0020\u002d\u0020\u0073\u006b\u0069\u0070\u0070\u0069n\u0067");continue ;};_eb ,_caad =_ce .GetStream (_cbe .Get (*_gb ));if !_caad {_ecd .Log .Debug ("\u0046o\u0072\u006d\u0020\u006eo\u0074\u0020\u0066\u006f\u0075n\u0064 \u002d \u0073\u006b\u0069\u0070\u0070\u0069\u006eg");continue ;};case *_ce .PdfObjectStream :_eb =_cbe ;};if _eb ==nil {_ecd .Log .Debug ("\u0046\u006f\u0072m\u0020\u006e\u006f\u0074 \u0066\u006f\u0075\u006e\u0064\u0020\u0028n\u0069\u006c\u0029\u0020\u002d\u0020\u0073\u006b\u0069\u0070\u0070\u0069\u006e\u0067");continue ;};_bge ,_cfc :=_d .NewXObjectFormFromStream (_eb );if _cfc !=nil {_ecd .Log .Debug ("\u0045\u0072\u0072\u006f\u0072\u0020l\u006f\u0061\u0064\u0069\u006e\u0067\u0020\u0066\u006f\u0072\u006d\u003a\u0020%\u0076\u0020\u002d\u0020\u0069\u0067\u006eo\u0072\u0069\u006e\u0067",_cfc );continue ;};_ecff ,_cfc :=_bge .GetContentStream ();if _cfc !=nil {_ecd .Log .Debug ("E\u0072\u0072\u006f\u0072\u0020\u0064e\u0063\u006f\u0064\u0069\u006e\u0067\u0020\u0063\u006fn\u0074\u0065\u006et\u0073:\u0020\u0025\u0076",_cfc );continue ;};_fbg =append (_fbg ,content {_eee :string (_ecff ),_eef :_bge .Resources });};return _fbg ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_ac *Chain )Optimize (objects []_ce .PdfObject )(_de []_ce .PdfObject ,_af error ){_de =objects ;for _ ,_bg :=range _ac ._bdc {_de ,_af =_bg .Optimize (_de );if _af !=nil {return _de ,_af ;};};return _de ,nil ;};

// Image optimizes images by rewrite images into JPEG format with quality equals to ImageQuality.
// TODO(a5i): Add support for inline images.
// It implements interface model.Optimizer.
type Image struct{ImageQuality int ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_ceae *CompressStreams )Optimize (objects []_ce .PdfObject )(_cbgf []_ce .PdfObject ,_ffgf error ){_cbgf =make ([]_ce .PdfObject ,len (objects ));copy (_cbgf ,objects );for _ ,_egc :=range objects {_beeg ,_gda :=_ce .GetStream (_egc );if !_gda {continue ;};if _ceac :=_beeg .Get ("\u0046\u0069\u006c\u0074\u0065\u0072");_ceac !=nil {if _ ,_eecd :=_ce .GetName (_ceac );_eecd {continue ;};if _ceea ,_bda :=_ce .GetArray (_ceac );_bda &&_ceea .Len ()> 0{continue ;};};_ffb :=_ce .NewFlateEncoder ();var _age []byte ;_age ,_ffgf =_ffb .EncodeBytes (_beeg .Stream );if _ffgf !=nil {return _cbgf ,_ffgf ;};_gffb :=_ffb .MakeStreamDict ();if len (_age )+len (_gffb .WriteString ())< len (_beeg .Stream ){_beeg .Stream =_age ;_beeg .PdfObjectDictionary .Merge (_gffb );_beeg .PdfObjectDictionary .Set ("\u004c\u0065\u006e\u0067\u0074\u0068",_ce .MakeInteger (int64 (len (_beeg .Stream ))));};};return _cbgf ,nil ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_cgbc *ObjectStreams )Optimize (objects []_ce .PdfObject )(_gfbb []_ce .PdfObject ,_ddaa error ){_agd :=&_ce .PdfObjectStreams {};_dgcd :=make ([]_ce .PdfObject ,0,len (objects ));for _ ,_cda :=range objects {if _bcce ,_bbgc :=_cda .(*_ce .PdfIndirectObject );_bbgc &&_bcce .GenerationNumber ==0{_agd .Append (_cda );}else {_dgcd =append (_dgcd ,_cda );};};if _agd .Len ()==0{return _dgcd ,nil ;};_gfbb =make ([]_ce .PdfObject ,0,len (_dgcd )+_agd .Len ()+1);if _agd .Len ()> 1{_gfbb =append (_gfbb ,_agd );};_gfbb =append (_gfbb ,_agd .Elements ()...);_gfbb =append (_gfbb ,_dgcd ...);return _gfbb ,nil ;};func _bcbd (_fec []_ce .PdfObject ,_ace map[_ce .PdfObject ]_ce .PdfObject ){if _ace ==nil ||len (_ace )==0{return ;};for _caaf ,_cdaa :=range _fec {if _fbff ,_dad :=_ace [_cdaa ];_dad {_fec [_caaf ]=_fbff ;continue ;};_ace [_cdaa ]=_cdaa ;switch _dac :=_cdaa .(type ){case *_ce .PdfObjectArray :_ggad :=make ([]_ce .PdfObject ,_dac .Len ());copy (_ggad ,_dac .Elements ());_bcbd (_ggad ,_ace );for _abb ,_deeb :=range _ggad {_dac .Set (_abb ,_deeb );};case *_ce .PdfObjectStreams :_bcbd (_dac .Elements (),_ace );case *_ce .PdfObjectStream :_aeff :=[]_ce .PdfObject {_dac .PdfObjectDictionary };_bcbd (_aeff ,_ace );_dac .PdfObjectDictionary =_aeff [0].(*_ce .PdfObjectDictionary );case *_ce .PdfObjectDictionary :_bag :=_dac .Keys ();_ggbe :=make ([]_ce .PdfObject ,len (_bag ));for _eac ,_bfbc :=range _bag {_ggbe [_eac ]=_dac .Get (_bfbc );};_bcbd (_ggbe ,_ace );for _gebc ,_ceaf :=range _bag {_dac .Set (_ceaf ,_ggbe [_gebc ]);};case *_ce .PdfIndirectObject :_bacc :=[]_ce .PdfObject {_dac .PdfObject };_bcbd (_bacc ,_ace );_dac .PdfObject =_bacc [0];};};};

// Optimize optimizes PDF objects to decrease PDF size.
func (_ba *CleanContentstream )Optimize (objects []_ce .PdfObject )(_ga []_ce .PdfObject ,_fe error ){_gcc :=map[*_ce .PdfObjectStream ]struct{}{};var _edf []*_ce .PdfObjectStream ;_eec :=func (_be *_ce .PdfObjectStream ){if _ ,_cb :=_gcc [_be ];!_cb {_gcc [_be ]=struct{}{};_edf =append (_edf ,_be );};};for _ ,_bde :=range objects {switch _dg :=_bde .(type ){case *_ce .PdfIndirectObject :switch _cae :=_dg .PdfObject .(type ){case *_ce .PdfObjectDictionary :if _bae ,_fg :=_ce .GetName (_cae .Get ("\u0054\u0079\u0070\u0065"));!_fg ||_bae .String ()!="\u0050\u0061\u0067\u0065"{continue ;};if _adg ,_dc :=_ce .GetStream (_cae .Get ("\u0043\u006f\u006e\u0074\u0065\u006e\u0074\u0073"));_dc {_eec (_adg );}else if _ea ,_eca :=_ce .GetArray (_cae .Get ("\u0043\u006f\u006e\u0074\u0065\u006e\u0074\u0073"));_eca {for _ ,_ab :=range _ea .Elements (){if _bga ,_cf :=_ce .GetStream (_ab );_cf {_eec (_bga );};};};};case *_ce .PdfObjectStream :if _bce ,_bee :=_ce .GetName (_dg .Get ("\u0054\u0079\u0070\u0065"));!_bee ||_bce .String ()!="\u0058O\u0062\u006a\u0065\u0063\u0074"{continue ;};if _cee ,_ged :=_ce .GetName (_dg .Get ("\u0053u\u0062\u0074\u0079\u0070\u0065"));!_ged ||_cee .String ()!="\u0046\u006f\u0072\u006d"{continue ;};_eec (_dg );};};for _ ,_geb :=range _edf {_fe =_ed (_geb );if _fe !=nil {return nil ,_fe ;};};return objects ,nil ;};type objectStructure struct{_bbga *_ce .PdfObjectDictionary ;_ffgdb *_ce .PdfObjectDictionary ;_cfeb []*_ce .PdfIndirectObject ;};func _cg (_abg []_ce .PdfObject )(_bf map[*_ce .PdfObjectStream ]struct{},_fbb error ){_bf =map[*_ce .PdfObjectStream ]struct{}{};_ag :=map[*_d .PdfFont ]struct{}{};_fa :=_cabd (_abg );for _ ,_eda :=range _fa ._cfeb {_df ,_abga :=_ce .GetDict (_eda .PdfObject );if !_abga {continue ;};_dec ,_abga :=_ce .GetDict (_df .Get ("\u0052e\u0073\u006f\u0075\u0072\u0063\u0065s"));if !_abga {continue ;};_afe ,_ :=_eba (_df .Get ("\u0043\u006f\u006e\u0074\u0065\u006e\u0074\u0073"));_afc ,_caa :=_d .NewPdfPageResourcesFromDict (_dec );if _caa !=nil {return nil ,_caa ;};_def :=[]content {{_eee :_afe ,_eef :_afc }};_fbe :=_bdd (_df .Get ("\u0041\u006e\u006e\u006f\u0074\u0073"));if _fbe !=nil {_def =append (_def ,_fbe ...);};for _ ,_gfe :=range _def {_bfg ,_cbb :=_ec .NewFromContents (_gfe ._eee ,_gfe ._eef );if _cbb !=nil {return nil ,_cbb ;};_cbbg ,_ ,_ ,_cbb :=_bfg .ExtractPageText ();if _cbb !=nil {return nil ,_cbb ;};for _ ,_dee :=range _cbbg .Marks ().Elements (){if _dee .Font ==nil {continue ;};if _ ,_acg :=_ag [_dee .Font ];!_acg {_ag [_dee .Font ]=struct{}{};};};};};_fd :=map[*_ce .PdfObjectStream ][]*_d .PdfFont {};for _eag :=range _ag {_ff :=_eag .FontDescriptor ();if _ff ==nil ||_ff .FontFile2 ==nil {continue ;};_fab ,_ffc :=_ce .GetStream (_ff .FontFile2 );if !_ffc {continue ;};_fd [_fab ]=append (_fd [_fab ],_eag );};for _bab :=range _fd {var _bea []rune ;var _agc []_ecf .GlyphIndex ;for _ ,_dda :=range _fd [_bab ]{switch _gee :=_dda .Encoder ().(type ){case *_gc .IdentityEncoder :_ddb :=_gee .RegisteredRunes ();_cbbf :=make ([]_ecf .GlyphIndex ,len (_ddb ));for _cd ,_cgf :=range _ddb {_cbbf [_cd ]=_ecf .GlyphIndex (_cgf );};_agc =append (_agc ,_cbbf ...);case *_gc .TrueTypeFontEncoder :_deb :=_gee .RegisteredRunes ();_bea =append (_bea ,_deb ...);case _gc .SimpleEncoder :_bgg :=_gee .Charcodes ();for _ ,_fgc :=range _bgg {_gff ,_dba :=_gee .CharcodeToRune (_fgc );if !_dba {_ecd .Log .Debug ("\u0043\u0068a\u0072\u0063\u006f\u0064\u0065\u003c\u002d\u003e\u0072\u0075\u006e\u0065\u0020\u006e\u006f\u0074\u0020\u0066\u006f\u0075\u006e\u0064: \u0025\u0064",_fgc );continue ;};_bea =append (_bea ,_gff );};};};_fbb =_eg (_bab ,_bea ,_agc );if _fbb !=nil {_ecd .Log .Debug ("\u0045\u0052\u0052\u004f\u0052\u0020\u0073\u0075\u0062\u0073\u0065\u0074\u0074\u0069\u006eg\u0020f\u006f\u006e\u0074\u0020\u0073\u0074\u0072\u0065\u0061\u006d\u003a\u0020\u0025\u0076",_fbb );return nil ,_fbb ;};_bf [_bab ]=struct{}{};};return _bf ,nil ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_gag *CombineDuplicateDirectObjects )Optimize (objects []_ce .PdfObject )(_acb []_ce .PdfObject ,_beb error ){_gedg (objects );_ceab :=make (map[string ][]*_ce .PdfObjectDictionary );var _gcg func (_ecae *_ce .PdfObjectDictionary );_gcg =func (_eadb *_ce .PdfObjectDictionary ){for _ ,_aba :=range _eadb .Keys (){_ggeg :=_eadb .Get (_aba );if _cgfc ,_dgc :=_ggeg .(*_ce .PdfObjectDictionary );_dgc {_dbd :=_b .New ();_dbd .Write ([]byte (_cgfc .WriteString ()));_bac :=string (_dbd .Sum (nil ));_ceab [_bac ]=append (_ceab [_bac ],_cgfc );_gcg (_cgfc );};};};for _ ,_agae :=range objects {_bfd ,_gfb :=_agae .(*_ce .PdfIndirectObject );if !_gfb {continue ;};if _bad ,_cab :=_bfd .PdfObject .(*_ce .PdfObjectDictionary );_cab {_gcg (_bad );};};_baf :=make ([]_ce .PdfObject ,0,len (_ceab ));_gaga :=make (map[_ce .PdfObject ]_ce .PdfObject );for _ ,_dgad :=range _ceab {if len (_dgad )< 2{continue ;};_cef :=_ce .MakeDict ();_cef .Merge (_dgad [0]);_fgf :=_ce .MakeIndirectObject (_cef );_baf =append (_baf ,_fgf );for _ffg :=0;_ffg < len (_dgad );_ffg ++{_aeb :=_dgad [_ffg ];_gaga [_aeb ]=_fgf ;};};_acb =make ([]_ce .PdfObject ,len (objects ));copy (_acb ,objects );_acb =append (_baf ,_acb ...);_bcbd (_acb ,_gaga );return _acb ,nil ;};

// Options describes PDF optimization parameters.
type Options struct{CombineDuplicateStreams bool ;CombineDuplicateDirectObjects bool ;ImageUpperPPI float64 ;ImageQuality int ;UseObjectStreams bool ;CombineIdenticalIndirectObjects bool ;CompressStreams bool ;CleanFonts bool ;SubsetFonts bool ;CleanContentstream bool ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_ced *ImagePPI )Optimize (objects []_ce .PdfObject )(_aac []_ce .PdfObject ,_ffd error ){if _ced .ImageUpperPPI <=0{return objects ,nil ;};_dcc :=_dce (objects );if len (_dcc )==0{return objects ,nil ;};_bdb :=make (map[_ce .PdfObject ]struct{});for _ ,_deee :=range _dcc {_bgc :=_deee .Stream .PdfObjectDictionary .Get (_ce .PdfObjectName ("\u0053\u004d\u0061s\u006b"));_bdb [_bgc ]=struct{}{};};_dcd :=make (map[*_ce .PdfObjectStream ]*imageInfo );for _ ,_ccd :=range _dcc {_dcd [_ccd .Stream ]=_ccd ;};var _gbe *_ce .PdfObjectDictionary ;for _ ,_acbc :=range objects {if _cead ,_abgb :=_ce .GetDict (_acbc );_gbe ==nil &&_abgb {if _gdab ,_dgb :=_ce .GetName (_cead .Get (_ce .PdfObjectName ("\u0054\u0079\u0070\u0065")));_dgb &&*_gdab =="\u0043a\u0074\u0061\u006c\u006f\u0067"{_gbe =_cead ;};};};if _gbe ==nil {return objects ,nil ;};_ecc ,_daef :=_ce .GetDict (_gbe .Get (_ce .PdfObjectName ("\u0050\u0061\u0067e\u0073")));if !_daef {return objects ,nil ;};_dgbc ,_deef :=_ce .GetArray (_ecc .Get (_ce .PdfObjectName ("\u004b\u0069\u0064\u0073")));if !_deef {return objects ,nil ;};_fgd :=make (map[string ]*imageInfo );for _ ,_ade :=range _dgbc .Elements (){_eea ,_bfgd :=_ce .GetDict (_ade );if !_bfgd {continue ;};_eccc ,_dfe :=_ce .GetArray (_eea .Get ("\u0043\u006f\u006e\u0074\u0065\u006e\u0074\u0073"));if !_dfe {continue ;};_bfe ,_fda :=_ce .GetDict (_eea .Get ("\u0052e\u0073\u006f\u0075\u0072\u0063\u0065s"));if !_fda {continue ;};_geaa ,_ecg :=_ce .GetDict (_bfe .Get ("\u0058O\u0062\u006a\u0065\u0063\u0074"));if !_ecg {continue ;};_gbb :=_geaa .Keys ();for _ ,_bdaa :=range _gbb {if _ceff ,_aef :=_ce .GetStream (_geaa .Get (_bdaa ));_aef {if _aad ,_bed :=_dcd [_ceff ];_bed {_fgd [string (_bdaa )]=_aad ;};};};for _ ,_cgb :=range _eccc .Elements (){if _ddff ,_edea :=_ce .GetStream (_cgb );_edea {_dfcd ,_fag :=_ce .NewEncoderFromStream (_ddff );if _fag !=nil {return nil ,_fag ;};_cec ,_fag :=_dfcd .DecodeStream (_ddff );if _fag !=nil {return nil ,_fag ;};_edac :=_gg .NewContentStreamParser (string (_cec ));_ddfc ,_fag :=_edac .Parse ();if _fag !=nil {return nil ,_fag ;};_ef ,_cbgd :=1.0,1.0;for _ ,_adeg :=range *_ddfc {if _adeg .Operand =="\u0051"{_ef ,_cbgd =1.0,1.0;};if _adeg .Operand =="\u0063\u006d"&&len (_adeg .Params )==6{if _ggb ,_fcg :=_ce .GetFloatVal (_adeg .Params [0]);_fcg {_ef =_ef *_ggb ;};if _becgf ,_bbg :=_ce .GetFloatVal (_adeg .Params [3]);_bbg {_cbgd =_cbgd *_becgf ;};if _cbgg ,_eab :=_ce .GetIntVal (_adeg .Params [0]);_eab {_ef =_ef *float64 (_cbgg );};if _dcda ,_bcca :=_ce .GetIntVal (_adeg .Params [3]);_bcca {_cbgd =_cbgd *float64 (_dcda );};};if _adeg .Operand =="\u0044\u006f"&&len (_adeg .Params )==1{_afea ,_cbf :=_ce .GetName (_adeg .Params [0]);if !_cbf {continue ;};if _abed ,_adc :=_fgd [string (*_afea )];_adc {_bfb ,_ebe :=_ef /72.0,_cbgd /72.0;_ddae ,_dedd :=float64 (_abed .Width )/_bfb ,float64 (_abed .Height )/_ebe ;if _bfb ==0||_ebe ==0{_ddae =72.0;_dedd =72.0;};_abed .PPI =_ee .Max (_abed .PPI ,_ddae );_abed .PPI =_ee .Max (_abed .PPI ,_dedd );};};};};};};for _ ,_faf :=range _dcc {if _ ,_ggdg :=_bdb [_faf .Stream ];_ggdg {continue ;};if _faf .PPI <=_ced .ImageUpperPPI {continue ;};_daefg :=_ced .ImageUpperPPI /_faf .PPI ;if _dcaf :=_gccd (_faf .Stream ,_daefg );_dcaf !=nil {_ecd .Log .Debug ("\u0045\u0072\u0072\u006f\u0072 \u0073\u0063\u0061\u006c\u0065\u0020\u0069\u006d\u0061\u0067\u0065\u0020\u006be\u0065\u0070\u0020\u006f\u0072\u0069\u0067\u0069\u006e\u0061\u006c\u0020\u0069\u006d\u0061\u0067\u0065\u003a\u0020\u0025\u0073",_dcaf );}else {if _deba ,_ffba :=_ce .GetStream (_faf .Stream .PdfObjectDictionary .Get (_ce .PdfObjectName ("\u0053\u004d\u0061s\u006b")));_ffba {if _abca :=_gccd (_deba ,_daefg );_abca !=nil {return nil ,_abca ;};};};};return objects ,nil ;};

// Optimize optimizes PDF objects to decrease PDF size.
func (_cbeb *Image )Optimize (objects []_ce .PdfObject )(_fbf []_ce .PdfObject ,_ggd error ){if _cbeb .ImageQuality <=0{return objects ,nil ;};_fbbf :=_dce (objects );if len (_fbbf )==0{return objects ,nil ;};_fabb :=make (map[_ce .PdfObject ]_ce .PdfObject );_fdg :=make (map[_ce .PdfObject ]struct{});for _ ,_cfaa :=range _fbbf {_dggb :=_cfaa .Stream .PdfObjectDictionary .Get (_ce .PdfObjectName ("\u0053\u004d\u0061s\u006b"));_fdg [_dggb ]=struct{}{};};for _gfbe ,_adgd :=range _fbbf {_ffad :=_adgd .Stream ;if _ ,_eaf :=_fdg [_ffad ];_eaf {continue ;};_bfa ,_ffgd :=_ce .NewEncoderFromStream (_ffad );if _ffgd !=nil {_ecd .Log .Warning ("\u0045\u0072\u0072\u006f\u0072 \u0067\u0065\u0074\u0020\u0065\u006e\u0063\u006f\u0064\u0065\u0072\u0020\u0066o\u0072\u0020\u0074\u0068\u0065\u0020\u0069\u006d\u0061\u0067\u0065\u0020\u0073\u0074\u0072\u0065\u0061\u006d\u0020\u0025\u0073");continue ;};_ebc ,_ffgd :=_bfa .DecodeStream (_ffad );if _ffgd !=nil {_ecd .Log .Warning ("\u0045\u0072\u0072\u006f\u0072\u0020\u0064\u0065\u0063\u006f\u0064\u0065\u0020\u0074\u0068e\u0020i\u006d\u0061\u0067\u0065\u0020\u0073\u0074\u0072\u0065\u0061\u006d\u0020\u0025\u0073");continue ;};_fdb :=_ce .NewDCTEncoder ();_fdb .ColorComponents =_adgd .ColorComponents ;_fdb .Quality =_cbeb .ImageQuality ;_fdb .BitsPerComponent =_adgd .BitsPerComponent ;_fdb .Width =_adgd .Width ;_fdb .Height =_adgd .Height ;_egdg ,_ffgd :=_fdb .EncodeBytes (_ebc );if _ffgd !=nil {_ecd .Log .Debug ("\u0045R\u0052\u004f\u0052\u003a\u0020\u0025v",_ffgd );return nil ,_ffgd ;};var _feb _ce .StreamEncoder ;_feb =_fdb ;{_dae :=_ce .NewFlateEncoder ();_gaf :=_ce .NewMultiEncoder ();_gaf .AddEncoder (_dae );_gaf .AddEncoder (_fdb );_deeg ,_bfc :=_gaf .EncodeBytes (_ebc );if _bfc !=nil {return nil ,_bfc ;};if len (_deeg )< len (_egdg ){_ecd .Log .Debug ("\u004d\u0075\u006c\u0074\u0069\u0020\u0065\u006e\u0063\u0020\u0069\u006d\u0070\u0072\u006f\u0076\u0065\u0073\u003a\u0020\u0025\u0064\u0020\u0074o\u0020\u0025\u0064\u0020\u0028o\u0072\u0069g\u0020\u0025\u0064\u0029",len (_egdg ),len (_deeg ),len (_ffad .Stream ));_egdg =_deeg ;_feb =_gaf ;};};_cdb :=len (_ffad .Stream );if _cdb < len (_egdg ){continue ;};_dfcc :=&_ce .PdfObjectStream {Stream :_egdg };_dfcc .PdfObjectReference =_ffad .PdfObjectReference ;_dfcc .PdfObjectDictionary =_ce .MakeDict ();_dfcc .Merge (_ffad .PdfObjectDictionary );_dfcc .Merge (_feb .MakeStreamDict ());_dfcc .Set ("\u004c\u0065\u006e\u0067\u0074\u0068",_ce .MakeInteger (int64 (len (_egdg ))));_fabb [_ffad ]=_dfcc ;_fbbf [_gfbe ].Stream =_dfcc ;};_fbf =make ([]_ce .PdfObject ,len (objects ));copy (_fbf ,objects );_bcbd (_fbf ,_fabb );return _fbf ,nil ;};type content struct{_eee string ;_eef *_d .PdfPageResources ;};