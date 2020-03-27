Acknowledgements
----------------

The UniDoc library uses resources from the following open source projects:

* [The standard Go library](https://golang.org/pkg/#stdlib), 
and [Go supplementary image libraries](https://godoc.org/golang.org/x/image/tiff/lzw), BSD-3 license:
  - Used for TIFF LZW encoding support.
  - PNG paeth algorithm
  
```
Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```


* [fpdf - Kurt Jung](https://github.com/jung-kurt/gofpdf), MIT license.

  - Used for TrueType (TTF) font file parsing (unidoc/pdf/model/fonts/ttfparser.go).
```
MIT License

Copyright (c) 2017 Kurt Jung and contributors acknowledged in the documentation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

* [Adobe Font Metrics PDF Core 14 fonts](http://www.adobe.com/devnet/font.html), with the following license:
```
This file and the 14 PostScript(R) AFM files it accompanies may be used,
copied, and distributed for any purpose and without charge, with or without
modification, provided that all copyright notices are retained; that the
AFM files are not distributed without this file; that all modifications
to this file or any of the AFM files are prominently noted in the modified
file(s); and that this paragraph is not modified. Adobe Systems has no
responsibility or obligation to support the use of the AFM files.
```

  - Used for support of the 14 core fonts (see unidoc/pdf/model/fonts/afms).

* [Adobe Glyph List](https://github.com/adobe-type-tools/agl-aglfn), BSD-3 license.
  - Used for glyph and textencoding support (see unidoc/pdf/model/textencoding/glyphlist).

```
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright
notice, this list of conditions and the following disclaimer in the
documentation and/or other materials provided with the distribution.

Neither the name of Adobe Systems Incorporated nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

* [Apache Java PDFBox JBIG2 Decoder](https://github.com/apache/pdfbox-jbig2), Apache License 2.0.
    In order to achieve full support for the JBIG2 Decoder, it was necessary to implement all possible decoding
    combinations defined in the JBIG2 standard, aka ITU T.88 and ISO/IEC 14492.
    With a lack of Golang JBIG2 Open Source package, we’ve decided that it would be best to base our own implementation
    on some solid and reliable library.
    The Apache PDFBox JBIG2 library fulfilled all our requirements. It has a really good quality of the code along with
    the detailed comments on each function and class. It also implemented  MMR, Huffman tables and arithmetic
    decompressors along with all JBIG2 segments.

* [AGL JBIG2 Encoder](https://github.com/agl/jbig2enc), Apache License 2.0.
    The complexity and lack of comprehensive documentation for the JBIG2 encoding process, lead us to look at the
    AGL JBIG2 Encoder library. At the moment of implementing our encoder it was the only Open Source JBIG2 encoder.
    It’s a C++ based library that implements both lossless and lossy encoding methods, where most of the image
    operations are done using DanBloomberg Leptonica library.

    The core encoding processes in the UniPDF JBIG2 Encoder were based on that well documented and solid library


* [DanBloomberg Leptonica](https://github.com/DanBloomberg/leptonica), The 2-Clause BSD License,
    DanBloomberg Leptonica is an amazing C/C++ Open Source library. It provides raster operations, binary expansion and
    reduction, JBIG2 component creators, correlation scoring and a lot more perfectly commented image operation functions.
    That library was used as a very solid base for our image operation algorithms used by the JBIG2 Encoder.
