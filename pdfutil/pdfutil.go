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

package pdfutil ;import (_a "github.com/unidoc/unipdf/v3/common";_ag "github.com/unidoc/unipdf/v3/contentstream";_c "github.com/unidoc/unipdf/v3/contentstream/draw";_f "github.com/unidoc/unipdf/v3/core";_e "github.com/unidoc/unipdf/v3/model";);

// NormalizePage performs the following operations on the passed in page:
//   - Normalize the page rotation.
//     Rotates the contents of the page according to the Rotate entry, thus
//     flattening the rotation. The Rotate entry of the page is set to nil.
//   - Normalize the media box.
//     If the media box of the page is offsetted (Llx != 0 or Lly != 0),
//     the contents of the page are translated to (-Llx, -Lly). After
//     normalization, the media box is updated (Llx and Lly are set to 0 and
//     Urx and Ury are updated accordingly).
//   - Normalize the crop box.
//     The crop box of the page is updated based on the previous operations.
//
// After normalization, the page should look the same if openend using a
// PDF viewer.
// NOTE: This function does not normalize annotations, outlines other parts
// that are not part of the basic geometry and page content streams.
func NormalizePage (page *_e .PdfPage )error {_fa ,_fc :=page .GetMediaBox ();if _fc !=nil {return _fc ;};_cf ,_fc :=page .GetRotate ();if _fc !=nil {_a .Log .Debug ("\u0045\u0052R\u004f\u0052\u003a\u0020\u0025\u0073\u0020\u002d\u0020\u0069\u0067\u006e\u006f\u0072\u0069\u006e\u0067\u0020\u0061\u006e\u0064\u0020\u0061\u0073\u0073\u0075\u006d\u0069\u006e\u0067\u0020\u006e\u006f\u0020\u0072\u006f\u0074\u0061\u0074\u0069\u006f\u006e\u000a",_fc .Error ());
};_cb :=_cf %360!=0&&_cf %90==0;_fa .Normalize ();_fac ,_d ,_ea ,_b :=_fa .Llx ,_fa .Lly ,_fa .Width (),_fa .Height ();_ed :=_fac !=0||_d !=0;if !_cb &&!_ed {return nil ;};_agd :=func (_eg ,_ca ,_bd float64 )_c .BoundingBox {return _c .Path {Points :[]_c .Point {_c .NewPoint (0,0).Rotate (_bd ),_c .NewPoint (_eg ,0).Rotate (_bd ),_c .NewPoint (0,_ca ).Rotate (_bd ),_c .NewPoint (_eg ,_ca ).Rotate (_bd )}}.GetBoundingBox ();
};_cc :=_ag .NewContentCreator ();var _faa float64 ;if _cb {_faa =-float64 (_cf );_df :=_agd (_ea ,_b ,_faa );_cc .Translate ((_df .Width -_ea )/2+_ea /2,(_df .Height -_b )/2+_b /2);_cc .RotateDeg (_faa );_cc .Translate (-_ea /2,-_b /2);_ea ,_b =_df .Width ,_df .Height ;
};if _ed {_cc .Translate (-_fac ,-_d );};_ga :=_cc .Operations ();_cd ,_fc :=_f .MakeStream (_ga .Bytes (),_f .NewFlateEncoder ());if _fc !=nil {return _fc ;};_ac :=_f .MakeArray (_cd );_ac .Append (page .GetContentStreamObjs ()...);*_fa =_e .PdfRectangle {Urx :_ea ,Ury :_b };
if _ee :=page .CropBox ;_ee !=nil {_ee .Normalize ();_dc ,_gb ,_db ,_cce :=_ee .Llx -_fac ,_ee .Lly -_d ,_ee .Width (),_ee .Height ();if _cb {_cg :=_agd (_db ,_cce ,_faa );_db ,_cce =_cg .Width ,_cg .Height ;};*_ee =_e .PdfRectangle {Llx :_dc ,Lly :_gb ,Urx :_dc +_db ,Ury :_gb +_cce };
};_a .Log .Debug ("\u0052\u006f\u0074\u0061\u0074\u0065\u003d\u0025\u0066\u00b0\u0020\u004f\u0070\u0073\u003d%\u0071 \u004d\u0065\u0064\u0069\u0061\u0042\u006f\u0078\u003d\u0025\u002e\u0032\u0066",_faa ,_ga ,_fa );page .Contents =_ac ;page .Rotate =nil ;
return nil ;};