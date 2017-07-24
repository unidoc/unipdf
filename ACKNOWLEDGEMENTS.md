Acknowledgements
----------------

The UniDoc library uses resources from the following open source projects:

* [The standard Go library](https://golang.org/pkg/#stdlib), MIT license.

* [Go supplementary image libraries](https://godoc.org/golang.org/x/image/tiff/lzw), BSD-3 license.

  - Used for TIFF LZW encoding support.

* [fpdf - Kurt Jung](https://github.com/jung-kurt/gofpdf), MIT license.

  - Used for TrueType (TTF) font file parsing (unidoc/pdf/model/fonts/ttfparser.go).

* [Adobe Font Metrics PDF Core 14 fonts](http://www.adobe.com/devnet/font.html), with the following license:

  This file and the 14 PostScript(R) AFM files it accompanies may be used,
copied, and distributed for any purpose and without charge, with or without
modification, provided that all copyright notices are retained; that the
AFM files are not distributed without this file; that all modifications
to this file or any of the AFM files are prominently noted in the modified
file(s); and that this paragraph is not modified. Adobe Systems has no
responsibility or obligation to support the use of the AFM files.

  - Used for support of the 14 core fonts (see unidoc/pdf/model/fonts/afms).

* [Adobe Glyph List](https://github.com/adobe-type-tools/agl-aglfn), BSD-3 license.

  - Used for glyph and textencoding support (see unidoc/pdf/model/textencoding/glyphlist).

