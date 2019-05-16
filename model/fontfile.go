/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 *

 /*
  * A font file is a stream containing a Type 1 font program. It appears in PDF files as a
  * /FontFile entry in a /FontDescriptor dictionary.
  *
  * 9.9 Embedded Font Programs (page 289)
  *
  * TODO: Add Type1C support
*/

package model

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/textencoding"
)

// fontFile represents a font file.
// Currently this is just the identifying information and the text encoder created from the font
// file's encoding section.
type fontFile struct {
	name    string
	subtype string
	encoder textencoding.SimpleEncoder
}

// String returns a human readable description of `fontfile`.
func (fontfile *fontFile) String() string {
	encoding := "[None]"
	if fontfile.encoder != nil {
		encoding = fontfile.encoder.String()
	}
	return fmt.Sprintf("FONTFILE{%#q encoder=%s}", fontfile.name, encoding)
}

// newFontFileFromPdfObject loads a FontFile from a PdfObject.  Can either be a
// *PdfIndirectObject or a *PdfObjectDictionary.
func newFontFileFromPdfObject(obj core.PdfObject) (*fontFile, error) {
	common.Log.Trace("newFontFileFromPdfObject: obj=%s", obj)
	fontfile := &fontFile{}

	obj = core.TraceToDirectObject(obj)

	streamObj, ok := obj.(*core.PdfObjectStream)
	if !ok {
		common.Log.Debug("ERROR: FontFile must be a stream (%T)", obj)
		return nil, core.ErrTypeError
	}
	d := streamObj.PdfObjectDictionary
	data, err := core.DecodeStream(streamObj)
	if err != nil {
		return nil, err
	}

	subtype, ok := core.GetNameVal(d.Get("Subtype"))
	if !ok {
		fontfile.subtype = subtype
		if subtype == "Type1C" {
			// TODO: Add Type1C support
			common.Log.Debug("Type1C fonts are currently not supported")
			return nil, ErrType1CFontNotSupported
		}
	}

	length1, _ := core.GetIntVal(d.Get("Length1"))
	length2, _ := core.GetIntVal(d.Get("Length2"))

	if length1 > len(data) {
		length1 = len(data)
	}
	if length1+length2 > len(data) {
		length2 = len(data) - length1
	}

	segment1 := data[:length1]
	var segment2 []byte
	if length2 > 0 {
		segment2 = data[length1 : length1+length2]
	}

	// empty streams are ignored
	if length1 > 0 && length2 > 0 {
		err := fontfile.loadFromSegments(segment1, segment2)
		if err != nil {
			return nil, err
		}
	}

	return fontfile, nil
}

// loadFromSegments loads a Type1Font object from two header-less .pfb segments.
// Based on pdfbox
func (fontfile *fontFile) loadFromSegments(segment1, segment2 []byte) error {
	common.Log.Trace("loadFromSegments: %d %d", len(segment1), len(segment2))
	err := fontfile.parseASCIIPart(segment1)
	if err != nil {
		return err
	}
	common.Log.Trace("fontfile=%s", fontfile)
	if len(segment2) == 0 {
		return nil
	}
	common.Log.Trace("fontfile=%s", fontfile)
	return nil
}

// parseASCIIPart parses the ASCII part of the FontFile.
func (fontfile *fontFile) parseASCIIPart(data []byte) error {

	// Uncomment these lines to see the contents of the font file. For debugging.
	// fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~^^^~~~~~~~~~~~~~~~~~~~~~~~")
	// fmt.Printf("data=%s\n", string(data))
	// fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~!!!~~~~~~~~~~~~~~~~~~~~~~~")

	// The start of a FontFile looks like
	//     %!PS-AdobeFont-1.0: MyArial 003.002
	//     %%Title: MyArial
	// or
	//     %!FontType1-1.0
	if len(data) < 2 || string(data[:2]) != "%!" {
		return errors.New("invalid start of ASCII segment")
	}

	keySection, encodingSection, err := getASCIISections(data)
	if err != nil {
		return err
	}
	keyValues := getKeyValues(keySection)

	fontfile.name = keyValues["FontName"]
	if fontfile.name == "" {
		common.Log.Debug(" FontFile has no /FontName")
	}

	if encodingSection != "" {
		encodings, err := getEncodings(encodingSection)
		if err != nil {
			return err
		}
		encoder, err := textencoding.NewCustomSimpleTextEncoder(encodings, nil)
		if err != nil {
			// NOTE(peterwilliams97): Logging an error because we need to fix all these misses.
			common.Log.Debug("ERROR :UNKNOWN GLYPH: err=%v", err)
			return nil
		}
		fontfile.encoder = encoder
	}
	return nil
}

var (
	reDictBegin   = regexp.MustCompile(`\d+ dict\s+(dup\s+)?begin`)
	reKeyVal      = regexp.MustCompile(`^\s*/(\S+?)\s+(.+?)\s+def\s*$`)
	reEncoding    = regexp.MustCompile(`^\s*dup\s+(\d+)\s*/(\w+?)(?:\.\d+)?\s+put$`)
	encodingBegin = "/Encoding 256 array"
	encodingEnd   = "readonly def"
	binaryStart   = "currentfile eexec"
)

// getASCIISections returns two sections of `data`, the ASCII part of the FontFile
//   - the general key values in `keySection`
//   - the encoding in `encodingSection`
func getASCIISections(data []byte) (keySection, encodingSection string, err error) {
	common.Log.Trace("getASCIISections: %d ", len(data))
	loc := reDictBegin.FindIndex(data)
	if loc == nil {
		common.Log.Debug("ERROR: getASCIISections. No dict.")
		return "", "", core.ErrTypeError
	}
	i0 := loc[1]
	i := strings.Index(string(data[i0:]), encodingBegin)
	if i < 0 {
		keySection = string(data[i0:])
		return keySection, "", nil
	}
	i1 := i0 + i
	keySection = string(data[i0:i1])

	i2 := i1
	i = strings.Index(string(data[i2:]), encodingEnd)
	if i < 0 {
		common.Log.Debug("ERROR: getASCIISections. err=%v", err)
		return "", "", core.ErrTypeError
	}
	i3 := i2 + i
	encodingSection = string(data[i2:i3])
	return keySection, encodingSection, nil
}

// ~/testdata/private/invoice61781040.pdf has \r line endings
var reEndline = regexp.MustCompile(`[\n\r]+`)

// getKeyValues returns the map encoded in `data`.
func getKeyValues(data string) map[string]string {
	lines := reEndline.Split(data, -1)
	keyValues := map[string]string{}
	for _, line := range lines {
		matches := reKeyVal.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		k, v := matches[1], matches[2]
		keyValues[k] = v
	}
	return keyValues
}

// getEncodings returns the encodings encoded in `data`.
func getEncodings(data string) (map[textencoding.CharCode]textencoding.GlyphName, error) {
	lines := strings.Split(data, "\n")
	keyValues := make(map[textencoding.CharCode]textencoding.GlyphName)
	for _, line := range lines {
		matches := reEncoding.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		k, glyph := matches[1], matches[2]
		code, err := strconv.Atoi(k)
		if err != nil {
			common.Log.Debug("ERROR: Bad encoding line. %q", line)
			return nil, core.ErrTypeError
		}
		keyValues[textencoding.CharCode(code)] = textencoding.GlyphName(glyph)
	}
	common.Log.Trace("getEncodings: keyValues=%#v", keyValues)
	return keyValues, nil
}

// decodeEexec returns the decoding of the eexec bytes `data`
func decodeEexec(data []byte) []byte {
	const c1 = 52845
	const c2 = 22719

	seed := 55665 // eexec key
	// Run the seed through the encoder 4 times
	for _, b := range data[:4] {
		seed = (int(b)+seed)*c1 + c2
	}
	decoded := make([]byte, len(data)-4)
	for i, b := range data[4:] {
		decoded[i] = byte(int(b) ^ seed>>8)
		seed = (int(b)+seed)*c1 + c2
	}
	return decoded
}

// isBinary returns true if `data` is binary. See Adobe Type 1 Font Format specification
// 7.2 eexec encryption
func isBinary(data []byte) bool {
	if len(data) < 4 {
		return true
	}
	for b := range data[:4] {
		r := rune(b)
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
