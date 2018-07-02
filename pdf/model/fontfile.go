package model

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

type fontFile struct {
	name    string
	encoder textencoding.TextEncoder
	// binary  []byte
}

// newFontFileFromPdfObject loads a FontFile from a PdfObject.  Can either be a
// *PdfIndirectObject or a *PdfObjectDictionary.
func newFontFileFromPdfObject(obj PdfObject) (*fontFile, error) {
	common.Log.Debug("newFontFileFromPdfObject: obj=%s", obj)
	fontfile := &fontFile{}

	if o, ok := obj.(*PdfIndirectObject); ok {
		obj = o.PdfObject
	}

	streamObj, ok := obj.(*PdfObjectStream)
	if !ok {
		common.Log.Debug("ERROR: FontFile must be a stream (%T)", obj)
		return nil, ErrTypeError
	}
	d := streamObj.PdfObjectDictionary
	data, err := DecodeStream(streamObj)
	if err != nil {
		common.Log.Error("err=%v", err)
		return nil, err
	}

	fmt.Printf("d=%s\n", d)
	fmt.Printf("data=%d\n", len(data))

	length1 := *(d.Get("Length1").(*PdfObjectInteger))
	length2 := *(d.Get("Length2").(*PdfObjectInteger))

	if len(data) > 0 && data[0] == 0x80 {
		// some bad files embed the entire PFB, see PDFBOX-2607
		// t1 = Type1Font.createWithPFB(bytes);
	} else {
		segment1 := data[:length1]
		segment2 := data[length1:length2]

		// empty streams are simply ignored
		if length1 > 0 && length2 > 0 {
			err := fontfile.loadFromSegments(segment1, segment2)
			if err != nil {
				common.Log.Error("err=%v", err)
				fmt.Println("+++++++++++++++++++++++++++++++++")
				fmt.Printf("%s\n", data)
				fmt.Println("+++++++++++++++++++++++++++++++++")
				panic(err)
				return nil, err
			}
			fmt.Printf("fontfile=%#v\n", fontfile)
		}
	}
	common.Log.Debug("fontfile=%#v", fontfile)
	return fontfile, nil
}

/**
 * Constructs a new Type1Font object from two header-less .pfb segments.
 *
 * @param segment1 The first segment, without header
 * @param segment2 The second segment, without header
 * @return A new Type1Font instance
 * @throws IOException if something went wrong
 */
func (fontfile *fontFile) loadFromSegments(segment1, segment2 []byte) error {
	common.Log.Debug("loadFromSegments: %d %d", len(segment1), len(segment2))
	err := fontfile.parseASCII(segment1)
	if err != nil {
		common.Log.Debug("err=%v", err)
		return err
	}
	common.Log.Debug("fontfile=%#v", fontfile)
	if len(segment2) == 0 {
		return nil
	}
	err = fontfile.parseBinary(segment2)
	if err != nil {
		common.Log.Debug("err=%v", err)
		return err
	}

	common.Log.Debug("fontfile=%#v", fontfile)
	return nil
}

func (fontfile *fontFile) parseASCII(data []byte) error {
	common.Log.Debug("parseASCII: %d ", len(data))
	// %!FontType1-1.0
	// %!PS-AdobeFont-1.0
	if len(data) < 2 || string(data[:2]) != "%!" {
		return errors.New("Invalid start of ASCII segment")
	}

	keySection, encodingSection, err := getSections(data)
	if err != nil {
		common.Log.Debug("err=%v", err)
		return err
	}
	keyValues := getKeyValues(keySection)

	fontfile.name = keyValues["FontName"]
	if fontfile.name == "" {
		panic("no name")
	}

	encodingName, ok := keyValues["Encoding"]
	if ok {
		encoder, err := textencoding.NewSimpleTextEncoder(encodingName, nil)
		if err != nil {
			common.Log.Debug("err=%v", err)
			return err
		}
		fontfile.encoder = encoder
	}
	if encodingSection != "" {
		encodings, err := getEncodings(encodingSection)
		if err != nil {
			common.Log.Debug("err=%v", err)
			// panic(err)
			return err
		}
		encoder, err := textencoding.NewCustomSimpleTextEncoder(encodings, nil)
		if err != nil {
			common.Log.Debug("err=%v", err)
			panic(err)
			return err
		}
		fontfile.encoder = encoder
		common.Log.Debug("encoder=%s", encoder)
		// panic("mmmss")
	}
	return nil
}

/**
 * Parses the binary portion of a Type 1 fontfile.
 */
func (fontfile *fontFile) parseBinary(data []byte) error {
	// Sometimes, fonts use  hex format
	if !isBinary(data) {
		decoded, err := hex.DecodeString(string(data))
		if err != nil {
			return err
		}
		data = decoded
	}
	decrypted := decrypt(data)
	fmt.Println("+++++++++++++++++++++^^^^^+++++++++++++++++++++++++++")
	fmt.Printf("decrypted=%s\n", decrypted[:100])
	fmt.Println("+++++++++++++++++++++@@@@@+++++++++++++++++++++++++++")
	// fontfile.binary = decrypted
	return nil
}

/*
   %!PS-AdobeFont-1.0: CMBX12 003.002
   %%Title: CMBX12
   %Version: 003.002
   %%CreationDate: Mon Jul 13 16:17:00 2009
   %%Creator: David M. Jones
   %Copyright: Copyright (c) 1997, 2009 American Mathematical Society
   %Copyright: (<http://www.ams.org>), with Reserved Font Name CMBX12.
   % This Font Software is licensed under the SIL Open Font License, Version 1.1.
   % This license is in the accompanying file OFL.txt, and is also
   % available with a FAQ at: http://scripts.sil.org/OFL.
   %%EndComments
   FontDirectory/CMBX12 known{/CMBX12 findfont dup/UniqueID known{dup
   /UniqueID get 5000769 eq exch/FontType get 1 eq and}{pop false}ifelse
   {save true}{false}ifelse}{false}ifelse
   11 dict begin
   /FontType 1 def
   /FontMatrix [0.001 0 0 0.001 0 0 ]readonly def
   /FontName /YHELPQ+CMBX12 def
   /FontBBox {-53 -251 1139 750 }readonly def
   /UniqueID 5000769 def
   /PaintType 0 def
   /FontInfo 9 dict dup begin
        /version (003.002) readonly def
        /Notice (Copyright \050c\051 1997, 2009 American Mathematical Society \050<http://www.ams.org>\051, with Reserved Font Name CMBX12.) readonly def
        /FullName (CMBX12) readonly def
       /FamilyName (Computer Modern) readonly def
       /Weight (Bold) readonly def
       /ItalicAngle 0 def
       /isFixedPitch false def
       /UnderlinePosition -100 def
       /UnderlineThickness 50 def
   end readonly def
   /Encoding 256 array
   0 1 255 {1 index exch /.notdef put} for
       dup 65 /A put
       dup 67 /C put
       dup 68 /D put
       dup 69 /E put
       dup 73 /I put
       dup 76 /L put
       dup 77 /M put
       dup 78 /N put
       dup 80 /P put
       dup 82 /R put
       dup 83 /S put
       dup 84 /T put
       dup 87 /W put
       dup 97 /a put
       dup 98 /b put
       dup 99 /c put
       dup 58 /colon put
       dup 100 /d put
       dup 101 /e put
       dup 102 /f put
       dup 14 /ffi put
       dup 53 /five put
       dup 52 /four put
       dup 103 /g put
       dup 104 /h put
       dup 45 /hyphen put
       dup 105 /i put
       dup 107 /k put
       dup 108 /l put
       dup 109 /m put
       dup 110 /n put
       dup 111 /o put
       dup 49 /one put
       dup 112 /p put
       dup 46 /period put
       dup 63 /question put
       dup 114 /r put
       dup 115 /s put
       dup 116 /t put
       dup 51 /three put
       dup 50 /two put
       dup 117 /u put
       dup 118 /v put
       dup 119 /w put
       dup 120 /x put
       dup 121 /y put
   readonly def
   currentdict end
   currentfile eexec

*/

var (
	reDictBegin   = regexp.MustCompile(`\d+ dict begin`)
	reKeyVal      = regexp.MustCompile(`^\s*/(\S+?)\s+(.+?)\s+def\s*$`)
	reEncoding    = regexp.MustCompile(`dup\s+(\d+)\s*/(\w+)\s+put`)
	dictEnd       = "end readonly def"
	encodingBegin = "/Encoding 256 array"
	encodingEnd   = "readonly def"
	binaryStart   = "currentfile eexec"
)

func getSections(data []byte) (keySection, encodingSection string, err error) {
	common.Log.Debug("getSections: %d ", len(data))
	loc := reDictBegin.FindIndex(data)
	if loc == nil {
		err = ErrTypeError
		common.Log.Debug("err=%v", err)
		return
	}
	i0 := loc[1]
	i := strings.Index(string(data[i0:]), dictEnd)
	if i < 0 {
		err = ErrTypeError
		common.Log.Debug("err=%v", err)
		return
	}
	i1 := i0 + i
	keySection = string(data[i0:i1])
	fmt.Printf("i0=%d i1=%d len=%d\n", i0, i1, len(data))

	i = strings.Index(string(data[i1:]), encodingBegin)
	if i < 0 {
		err = ErrTypeError
		common.Log.Debug("err=%v", err)
		return
	}
	i2 := i1 + i
	i = strings.Index(string(data[i2:]), encodingEnd)
	if i < 0 {
		err = ErrTypeError
		common.Log.Debug("err=%v", err)
		return
	}
	i3 := i2 + i
	encodingSection = string(data[i2:i3])
	fmt.Printf("keySection=%q\n", truncate(keySection, 200))
	fmt.Printf("encodingSection=%q\n", truncate(encodingSection, 200))
	// panic("pppppp")
	return
}

// truncate returns the first `n` characters in string `s`
func truncate(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
func getKeyValues(data string) map[string]string {
	lines := strings.Split(data, "\n")
	keyValues := map[string]string{}
	for _, line := range lines {
		matches := reKeyVal.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		k, v := matches[1], matches[2]
		keyValues[k] = v
		// common.Log.Debug("%3d: line=%q k=%#q v=%q", i, line, k, v)
	}
	return keyValues
}

func getEncodings(data string) (map[uint16]string, error) {
	common.Log.Debug("getEncodings: data=%q", data)
	lines := strings.Split(data, "\n")
	keyValues := map[uint16]string{}
	for _, line := range lines {
		matches := reEncoding.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		k, glyph := matches[1], matches[2]
		code, err := strconv.Atoi(k)
		if err != nil {
			common.Log.Debug("ERROR: Bad encoding line. %q", line)
			return nil, ErrTypeCheck
		}
		if !textencoding.KnownGlyph(glyph) {
			common.Log.Debug("ERROR: Unknown glyph %q. line=%q", glyph, line)
			return nil, ErrTypeCheck
		}
		keyValues[uint16(code)] = glyph
	}
	common.Log.Debug("getEncodings: keyValues=%#v", keyValues)
	return keyValues, nil
}

/**
 * Type 1 Decryption (eexec, charstring).
 *
 * @param cipherBytes cipher text
 * @param r key
 * @param n number of random bytes (lenIV)
 * @return plain text
 */
func decrypt(cipherBytes []byte) []byte {
	const c1 = 52845
	const c2 = 22719

	r := 55665 // eexec key
	plainBytes := make([]byte, len(cipherBytes)-4)
	for i, b := range cipherBytes {
		cipher := int(b)
		plain := cipher ^ r>>8
		if i >= 4 {
			plainBytes[i-4] = byte(plain)
		}
		r = (cipher+r)*c1 + c2&0xffff
	}
	return plainBytes
}

// Check whether binary or hex encoded. See Adobe Type 1 Font Format specification
// 7.2 eexec encryption
func isBinary(data []byte) bool {
	if len(data) < 4 {
		return true
	}
	// "At least one of the first 4 ciphertext bytes must not be one of
	// the ASCII hexadecimal character codes (a code for 0-9, A-F, or a-f)."
	for b := range data[:4] {
		r := rune(b)
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
