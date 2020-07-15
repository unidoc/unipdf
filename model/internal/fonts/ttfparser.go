/*
 * Copyright (c) 2013 Kurt Jung (Gmail: kurt.w.jung)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */
/*
 * Copyright (c) 2018 FoxyUtils ehf. to modifications of the original.
 * Modifications of the original file are subject to the terms and conditions
 * defined in file 'LICENSE.md', which is part of this source code package.
 */

package fonts

// Utility to parse TTF font files
// Version:    1.0
// Date:       2011-06-18
// Author:     Olivier PLATHEY
// Port to Go: Kurt Jung, 2013-07-15

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/cmap"
	"github.com/unidoc/unipdf/v3/internal/textencoding"
)

// MakeEncoder returns an encoder built from the tables in `rec`.
func (ttf *TtfType) MakeEncoder() (textencoding.SimpleEncoder, error) {
	encoding := make(map[textencoding.CharCode]GlyphName)
	// TODO(dennwc): this is a bit strange, since TTF may contain more than 256 characters
	//				 should probably make a different encoder here
	for code := textencoding.CharCode(0); code <= 256; code++ {
		r := rune(code) // TODO(dennwc): make sure this conversion is valid
		gid, ok := ttf.Chars[r]
		if !ok {
			continue
		}
		var glyph GlyphName
		if int(gid) >= 0 && int(gid) < len(ttf.GlyphNames) {
			glyph = ttf.GlyphNames[gid]
		} else {
			r := rune(gid)
			if g, ok := textencoding.RuneToGlyph(r); ok {
				glyph = g
			}
		}
		if glyph != "" {
			encoding[code] = glyph
		}
	}
	if len(encoding) == 0 {
		common.Log.Debug("WARNING: Zero length TrueType encoding. ttf=%s Chars=[% 02x]",
			ttf, ttf.Chars)
	}
	return textencoding.NewCustomSimpleTextEncoder(encoding, nil)
}

// GID is a glyph index.
type GID = textencoding.GID

// GlyphName is a name of a glyph.
type GlyphName = textencoding.GlyphName

// TtfType describes a TrueType font file.
// http://scripts.sil.org/cms/scripts/page.php?site_id=nrsi&id=iws-chapter08
type TtfType struct {
	UnitsPerEm             uint16
	PostScriptName         string
	Bold                   bool
	ItalicAngle            float64
	IsFixedPitch           bool
	TypoAscender           int16
	TypoDescender          int16
	UnderlinePosition      int16
	UnderlineThickness     int16
	Xmin, Ymin, Xmax, Ymax int16
	CapHeight              int16
	// Widths is a list of glyph widths indexed by GID.
	Widths []uint16

	// Chars maps rune values (unicode) to GIDs (the indexes in GlyphNames). i.e. GlyphNames[Chars[r]] is
	// the glyph corresponding to rune r.
	//
	// TODO(dennwc): CharCode is currently defined as uint16, but some tables may store 32 bit charcodes
	//				 not the case right now, but make sure to update it once we support those tables
	// TODO(dennwc,peterwilliams97): it should map char codes to GIDs
	Chars map[rune]GID
	// GlyphNames is a list of glyphs from the "post" section of the TrueType file.
	GlyphNames []GlyphName
}

// MakeToUnicode returns a ToUnicode CMap based on the encoding of `ttf`.
// TODO(peterwilliams97): This currently gives a bad text mapping for creator_test.go but leads to an
// otherwise valid PDF file that Adobe Reader displays without error.
func (ttf *TtfType) MakeToUnicode() *cmap.CMap {
	codeToUnicode := make(map[cmap.CharCode]rune)
	if len(ttf.GlyphNames) == 0 {
		return cmap.NewToUnicodeCMap(codeToUnicode)
	}

	for code, gid := range ttf.Chars {
		// TODO(dennwc): this function is used only in one place and relies on
		//  			 the fact that the code uses identity CID<->GID mapping
		charcode := cmap.CharCode(code)

		glyph := ttf.GlyphNames[gid]

		// TODO(dennwc): 'code' is already a rune; do we need this extra lookup?
		// TODO(dennwc): this cannot be done here; glyphNames might be empty
		//				 the parent font may specify a different encoding
		//				 so we should remap on a higher level
		r, ok := textencoding.GlyphToRune(glyph)
		if !ok {
			common.Log.Debug("No rune. code=0x%04x glyph=%q", code, glyph)
			r = textencoding.MissingCodeRune
		}
		codeToUnicode[charcode] = r
	}
	return cmap.NewToUnicodeCMap(codeToUnicode)
}

// NewEncoder returns a new TrueType font encoder.
func (ttf *TtfType) NewEncoder() textencoding.TextEncoder {
	return textencoding.NewTrueTypeFontEncoder(ttf.Chars)
}

// String returns a human readable representation of `ttf`.
func (ttf *TtfType) String() string {
	return fmt.Sprintf("FONT_FILE2{%#q UnitsPerEm=%d Bold=%t ItalicAngle=%f "+
		"CapHeight=%d Chars=%d GlyphNames=%d}",
		ttf.PostScriptName, ttf.UnitsPerEm, ttf.Bold, ttf.ItalicAngle,
		ttf.CapHeight, len(ttf.Chars), len(ttf.GlyphNames))
}

// ttfParser contains some state variables used to parse a TrueType file.
type ttfParser struct {
	rec              TtfType
	f                io.ReadSeeker
	tables           map[string]uint32
	numberOfHMetrics uint16
	numGlyphs        uint16
}

// NewFontFile2FromPdfObject returns a TtfType describing the TrueType font file in PdfObject `obj`.
func NewFontFile2FromPdfObject(obj core.PdfObject) (TtfType, error) {
	obj = core.TraceToDirectObject(obj)
	streamObj, ok := obj.(*core.PdfObjectStream)
	if !ok {
		common.Log.Debug("ERROR: FontFile2 must be a stream (%T)", obj)
		return TtfType{}, core.ErrTypeError
	}
	data, err := core.DecodeStream(streamObj)
	if err != nil {
		return TtfType{}, err
	}

	// Uncomment these lines to see the contents of the font file. For debugging.
	// fmt.Println("===============&&&&===============")
	// fmt.Printf("%#q", string(data))
	// fmt.Println("===============####===============")

	t := ttfParser{f: bytes.NewReader(data)}
	return t.Parse()
}

// TtfParseFile returns a TtfType describing the TrueType font file in disk file `fileStr`.
func TtfParseFile(fileStr string) (TtfType, error) {
	f, err := os.Open(fileStr)
	if err != nil {
		return TtfType{}, err
	}
	defer f.Close()

	return TtfParse(f)
}

// TtfParse returns a TtfType describing the TrueType font.
func TtfParse(r io.ReadSeeker) (TtfType, error) {
	t := &ttfParser{f: r}
	return t.Parse()
}

// Parse returns a TtfType describing the TrueType font file in io.Reader `t`.f.
func (t *ttfParser) Parse() (TtfType, error) {

	version, err := t.ReadStr(4)
	if err != nil {
		return TtfType{}, err
	}
	if version == "OTTO" {
		// See https://docs.microsoft.com/en-us/typography/opentype/spec/otff
		return TtfType{}, fmt.Errorf("fonts based on PostScript outlines are not supported (%v)",
			core.ErrNotSupported)
	}
	if version != "\x00\x01\x00\x00" && version != "true" {
		// This is not an error. In the font_test.go example axes.txt we see version "true".
		common.Log.Debug("Unrecognized TrueType file format. version=%q", version)
	}
	numTables := int(t.ReadUShort())
	t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
	t.tables = make(map[string]uint32)
	var tag string
	for j := 0; j < numTables; j++ {
		tag, err = t.ReadStr(4)
		if err != nil {
			return TtfType{}, err
		}
		t.Skip(4) // checkSum
		offset := t.ReadULong()
		t.Skip(4) // length
		t.tables[tag] = offset
	}

	common.Log.Trace(describeTables(t.tables))

	if err = t.ParseComponents(); err != nil {
		return TtfType{}, err
	}
	return t.rec, nil
}

// describeTables returns a string describing `tables`, the tables in a TrueType font file.
func describeTables(tables map[string]uint32) string {
	var tags []string
	for tag := range tables {
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool { return tables[tags[i]] < tables[tags[j]] })
	parts := []string{fmt.Sprintf("TrueType tables: %d", len(tables))}
	for _, tag := range tags {
		parts = append(parts, fmt.Sprintf("\t%q %5d", tag, tables[tag]))
	}
	return strings.Join(parts, "\n")
}

// ParseComponents parses the tables in a TrueType font file.
// The standard TrueType tables are
// "head"
// "hhea"
// "loca"
// "maxp"
// "cvt "
// "prep"
// "glyf"
// "hmtx"
// "fpgm"
// "gasp"
func (t *ttfParser) ParseComponents() error {

	// Mandatory tables.
	if err := t.ParseHead(); err != nil {
		return err
	}
	if err := t.ParseHhea(); err != nil {
		return err
	}
	if err := t.ParseMaxp(); err != nil {
		return err
	}
	if err := t.ParseHmtx(); err != nil {
		return err
	}

	// Optional tables.
	if _, ok := t.tables["name"]; ok {
		if err := t.ParseName(); err != nil {
			return err
		}
	}
	if _, ok := t.tables["OS/2"]; ok {
		if err := t.ParseOS2(); err != nil {
			return err
		}
	}
	if _, ok := t.tables["post"]; ok {
		if err := t.ParsePost(); err != nil {
			return err
		}
	}
	if _, ok := t.tables["cmap"]; ok {
		if err := t.ParseCmap(); err != nil {
			return err
		}
	}

	return nil
}

func (t *ttfParser) ParseHead() error {
	if err := t.Seek("head"); err != nil {
		return err
	}
	t.Skip(3 * 4) // version, fontRevision, checkSumAdjustment
	magicNumber := t.ReadULong()
	if magicNumber != 0x5F0F3CF5 {
		// outputmanager.pdf displays in Adobe Reader but has a bad magic
		// number so we don't return an error here.
		// TODO(dennwc): check if it's a "head" table different format - this should not blindly accept anything
		common.Log.Debug("ERROR: Incorrect magic number. Font may not display correctly. %s", t)
	}
	t.Skip(2) // flags
	t.rec.UnitsPerEm = t.ReadUShort()
	t.Skip(2 * 8) // created, modified
	t.rec.Xmin = t.ReadShort()
	t.rec.Ymin = t.ReadShort()
	t.rec.Xmax = t.ReadShort()
	t.rec.Ymax = t.ReadShort()
	return nil
}

func (t *ttfParser) ParseHhea() error {
	if err := t.Seek("hhea"); err != nil {
		return err
	}
	t.Skip(4 + 15*2)
	t.numberOfHMetrics = t.ReadUShort()
	return nil
}

func (t *ttfParser) ParseMaxp() error {
	if err := t.Seek("maxp"); err != nil {
		return err
	}
	t.Skip(4)
	t.numGlyphs = t.ReadUShort()
	return nil
}

// ParseHmtx parses the Horizontal Metrics table in a TrueType.
func (t *ttfParser) ParseHmtx() error {
	if err := t.Seek("hmtx"); err != nil {
		return err
	}

	t.rec.Widths = make([]uint16, 0, 8)
	for j := uint16(0); j < t.numberOfHMetrics; j++ {
		t.rec.Widths = append(t.rec.Widths, t.ReadUShort())
		t.Skip(2) // lsb
	}
	if t.numberOfHMetrics < t.numGlyphs && t.numberOfHMetrics > 0 {
		lastWidth := t.rec.Widths[t.numberOfHMetrics-1]
		for j := t.numberOfHMetrics; j < t.numGlyphs; j++ {
			t.rec.Widths = append(t.rec.Widths, lastWidth)
		}
	}

	return nil
}

// parseCmapSubtable31 parses information from an (3,1) subtable (Windows Unicode).
func (t *ttfParser) parseCmapSubtable31(offset31 int64) error {
	startCount := make([]rune, 0, 8)
	endCount := make([]rune, 0, 8)
	idDelta := make([]int16, 0, 8)
	idRangeOffset := make([]uint16, 0, 8)
	t.rec.Chars = make(map[rune]GID)
	t.f.Seek(int64(t.tables["cmap"])+offset31, os.SEEK_SET)
	format := t.ReadUShort()
	if format != 4 {
		return fmt.Errorf("unexpected subtable format: %d (%v)", format, core.ErrNotSupported)
	}
	t.Skip(2 * 2) // length, language
	segCount := int(t.ReadUShort() / 2)
	t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
	for j := 0; j < segCount; j++ {
		endCount = append(endCount, rune(t.ReadUShort()))
	}
	t.Skip(2) // reservedPad
	for j := 0; j < segCount; j++ {
		startCount = append(startCount, rune(t.ReadUShort()))
	}
	for j := 0; j < segCount; j++ {
		idDelta = append(idDelta, t.ReadShort())
	}
	offset, _ := t.f.Seek(int64(0), os.SEEK_CUR)
	for j := 0; j < segCount; j++ {
		idRangeOffset = append(idRangeOffset, t.ReadUShort())
	}
	for j := 0; j < segCount; j++ {
		c1 := startCount[j]
		c2 := endCount[j]
		d := idDelta[j]
		ro := idRangeOffset[j]
		if ro > 0 {
			t.f.Seek(offset+2*int64(j)+int64(ro), os.SEEK_SET)
		}
		for c := c1; c <= c2; c++ {
			if c == 0xFFFF {
				break
			}
			var gid int32
			if ro > 0 {
				gid = int32(t.ReadUShort())
				if gid > 0 {
					gid += int32(d)
				}
			} else {
				gid = int32(c) + int32(d)
			}
			if gid >= 65536 {
				gid -= 65536
			}
			if gid > 0 {
				t.rec.Chars[c] = GID(gid)
			}
		}
	}
	return nil
}

// parseCmapSubtable10 parses information from an (1,0) subtable (symbol).
func (t *ttfParser) parseCmapSubtable10(offset10 int64) error {

	if t.rec.Chars == nil {
		t.rec.Chars = make(map[rune]GID)
	}

	t.f.Seek(int64(t.tables["cmap"])+offset10, os.SEEK_SET)
	var length, language uint32
	format := t.ReadUShort()
	if format < 8 {
		length = uint32(t.ReadUShort())
		language = uint32(t.ReadUShort())
	} else {
		t.ReadUShort()
		length = t.ReadULong()
		language = t.ReadULong()
	}
	common.Log.Trace("parseCmapSubtable10: format=%d length=%d language=%d",
		format, length, language)

	if format != 0 {
		return errors.New("unsupported cmap subtable format")
	}

	dataStr, err := t.ReadStr(256)
	if err != nil {
		return err
	}
	data := []byte(dataStr)

	for code, gid := range data {
		t.rec.Chars[rune(code)] = GID(gid)
		if gid != 0 {
			fmt.Printf("\t0x%02x ➞ 0x%02x=%c\n", code, gid, rune(gid))
		}
	}
	return nil
}

// ParseCmap parses the cmap table in a TrueType font.
func (t *ttfParser) ParseCmap() error {
	var offset int64
	if err := t.Seek("cmap"); err != nil {
		return err
	}
	common.Log.Trace("ParseCmap")
	t.ReadUShort() // version is ignored.
	numTables := int(t.ReadUShort())
	offset10 := int64(0)
	offset31 := int64(0)
	for j := 0; j < numTables; j++ {
		platformID := t.ReadUShort()
		encodingID := t.ReadUShort()
		offset = int64(t.ReadULong())
		if platformID == 3 && encodingID == 1 {
			// (3,1) subtable. Windows Unicode.
			offset31 = offset
		} else if platformID == 1 && encodingID == 0 {
			offset10 = offset
		}
	}

	// Many non-Latin fonts (including asian fonts) use subtable (1,0).
	if offset10 != 0 {
		if err := t.parseCmapVersion(offset10); err != nil {
			return err
		}
	}

	// Latin font support based on (3,1) table encoding.
	if offset31 != 0 {
		if err := t.parseCmapSubtable31(offset31); err != nil {
			return err
		}
	}

	if offset31 == 0 && offset10 == 0 {
		common.Log.Debug("ttfParser.ParseCmap. No 31 or 10 table.")
	}

	return nil
}

func (t *ttfParser) parseCmapVersion(offset int64) error {
	common.Log.Trace("parseCmapVersion: offset=%d", offset)

	if t.rec.Chars == nil {
		t.rec.Chars = make(map[rune]GID)
	}

	t.f.Seek(int64(t.tables["cmap"])+offset, os.SEEK_SET)
	var length, language uint32
	format := t.ReadUShort()
	if format < 8 {
		length = uint32(t.ReadUShort())
		language = uint32(t.ReadUShort())
	} else {
		t.ReadUShort()
		length = t.ReadULong()
		language = t.ReadULong()
	}
	common.Log.Debug("parseCmapVersion: format=%d length=%d language=%d",
		format, length, language)

	switch format {
	case 0:
		return t.parseCmapFormat0()
	case 6:
		return t.parseCmapFormat6()
	case 12:
		return t.parseCmapFormat12()
	default:
		common.Log.Debug("ERROR: Unsupported cmap format=%d", format)
		return nil // TODO(peterwilliams97): Can't return an error here if creator_test.go is to pass.
	}
}

func (t *ttfParser) parseCmapFormat0() error {
	dataStr, err := t.ReadStr(256)
	if err != nil {
		return err
	}
	data := []byte(dataStr)
	common.Log.Trace("parseCmapFormat0: %s\ndataStr=%+q\ndata=[% 02x]", t.rec.String(), dataStr, data)

	for code, glyphID := range data {
		t.rec.Chars[rune(code)] = GID(glyphID)
	}
	return nil
}

func (t *ttfParser) parseCmapFormat6() error {

	firstCode := int(t.ReadUShort())
	entryCount := int(t.ReadUShort())

	common.Log.Trace("parseCmapFormat6: %s firstCode=%d entryCount=%d",
		t.rec.String(), firstCode, entryCount)

	for i := 0; i < entryCount; i++ {
		glyphID := GID(t.ReadUShort())
		t.rec.Chars[rune(i+firstCode)] = glyphID
	}

	return nil
}

func (t *ttfParser) parseCmapFormat12() error {

	numGroups := t.ReadULong()

	common.Log.Trace("parseCmapFormat12: %s numGroups=%d", t.rec.String(), numGroups)

	for i := uint32(0); i < numGroups; i++ {
		firstCode := t.ReadULong()
		endCode := t.ReadULong()
		startGlyph := t.ReadULong()

		if firstCode > 0x0010FFFF || (0xD800 <= firstCode && firstCode <= 0xDFFF) {
			return errors.New("invalid characters codes")
		}

		if endCode < firstCode || endCode > 0x0010FFFF || (0xD800 <= endCode && endCode <= 0xDFFF) {
			return errors.New("invalid characters codes")
		}

		for j := uint32(0); j <= endCode-firstCode; j++ {
			glyphID := startGlyph + j
			if firstCode+j > 0x10FFFF {
				common.Log.Debug("Format 12 cmap contains character beyond UCS-4")
			}

			t.rec.Chars[rune(i+firstCode)] = GID(glyphID)
		}

	}

	return nil
}

// ParseName parses the "name" table.
func (t *ttfParser) ParseName() error {
	if err := t.Seek("name"); err != nil {
		return err
	}
	tableOffset, _ := t.f.Seek(0, os.SEEK_CUR)
	t.rec.PostScriptName = ""
	t.Skip(2) // format
	count := t.ReadUShort()
	stringOffset := t.ReadUShort()
	for j := uint16(0); j < count && t.rec.PostScriptName == ""; j++ {
		t.Skip(3 * 2) // platformID, encodingID, languageID
		nameID := t.ReadUShort()
		length := t.ReadUShort()
		offset := t.ReadUShort()
		if nameID == 6 {
			// PostScript name
			t.f.Seek(int64(tableOffset)+int64(stringOffset)+int64(offset), os.SEEK_SET)
			s, err := t.ReadStr(int(length))
			if err != nil {
				return err
			}
			s = strings.Replace(s, "\x00", "", -1)
			re, err := regexp.Compile("[(){}<> /%[\\]]")
			if err != nil {
				return err
			}
			t.rec.PostScriptName = re.ReplaceAllString(s, "")
		}
	}
	if t.rec.PostScriptName == "" {
		common.Log.Debug("ParseName: The name PostScript was not found.")
	}
	return nil
}

func (t *ttfParser) ParseOS2() error {
	if err := t.Seek("OS/2"); err != nil {
		return err
	}
	version := t.ReadUShort()
	t.Skip(4 * 2) // xAvgCharWidth, usWeightClass, usWidthClass
	t.Skip(11*2 + 10 + 4*4 + 4)
	fsSelection := t.ReadUShort()
	t.rec.Bold = (fsSelection & 32) != 0
	t.Skip(2 * 2) // usFirstCharIndex, usLastCharIndex
	t.rec.TypoAscender = t.ReadShort()
	t.rec.TypoDescender = t.ReadShort()
	if version >= 2 {
		t.Skip(3*2 + 2*4 + 2)
		t.rec.CapHeight = t.ReadShort()
	} else {
		t.rec.CapHeight = 0
	}
	return nil
}

// ParsePost reads the "post" section in a TrueType font table and sets t.rec.GlyphNames.
func (t *ttfParser) ParsePost() error {
	if err := t.Seek("post"); err != nil {
		return err
	}

	formatType := t.Read32Fixed()
	t.rec.ItalicAngle = t.Read32Fixed()
	t.rec.UnderlinePosition = t.ReadShort()
	t.rec.UnderlineThickness = t.ReadShort()
	t.rec.IsFixedPitch = t.ReadULong() != 0
	t.ReadULong() // minMemType42 ignored.
	t.ReadULong() // maxMemType42 ignored.
	t.ReadULong() // mimMemType1 ignored.
	t.ReadULong() // maxMemType1 ignored.

	common.Log.Trace("ParsePost: formatType=%f", formatType)

	switch formatType {
	case 1.0: // This font file contains the standard Macintosh TrueTyp 258 glyphs.
		t.rec.GlyphNames = macGlyphNames
	case 2.0:
		numGlyphs := int(t.ReadUShort())
		glyphNameIndex := make([]int, numGlyphs)
		t.rec.GlyphNames = make([]GlyphName, numGlyphs)
		maxIndex := -1
		for i := 0; i < numGlyphs; i++ {
			index := int(t.ReadUShort())
			glyphNameIndex[i] = index
			// Index numbers between 0x7fff and 0xffff are reserved for future use
			if index <= 0x7fff && index > maxIndex {
				maxIndex = index
			}
		}
		var nameArray []GlyphName
		if maxIndex >= len(macGlyphNames) {
			nameArray = make([]GlyphName, maxIndex-len(macGlyphNames)+1)
			for i := 0; i < maxIndex-len(macGlyphNames)+1; i++ {
				numberOfChars := int(t.readByte())
				names, err := t.ReadStr(numberOfChars)
				if err != nil {
					return err
				}
				nameArray[i] = GlyphName(names)
			}
		}
		for i := 0; i < numGlyphs; i++ {
			index := glyphNameIndex[i]
			if index < len(macGlyphNames) {
				t.rec.GlyphNames[i] = macGlyphNames[index]
			} else if index >= len(macGlyphNames) && index <= 32767 {
				t.rec.GlyphNames[i] = nameArray[index-len(macGlyphNames)]
			} else {
				t.rec.GlyphNames[i] = ".undefined"
			}
		}
	case 2.5:
		glyphNameIndex := make([]int, t.numGlyphs)
		for i := 0; i < len(glyphNameIndex); i++ {
			offset := int(t.ReadSByte())
			glyphNameIndex[i] = i + 1 + offset
		}
		t.rec.GlyphNames = make([]GlyphName, len(glyphNameIndex))
		for i := 0; i < len(t.rec.GlyphNames); i++ {
			name := macGlyphNames[glyphNameIndex[i]]
			t.rec.GlyphNames[i] = name
		}
	case 3.0:
		// No PostScript information is provided.
		common.Log.Debug("No PostScript name information is provided for the font.")
	default:
		common.Log.Debug("ERROR: Unknown formatType=%f", formatType)
	}

	return nil
}

// The 258 standard mac glyph names used in 'post' format 1 and 2.
var macGlyphNames = []GlyphName{
	".notdef", ".null", "nonmarkingreturn", "space", "exclam", "quotedbl",
	"numbersign", "dollar", "percent", "ampersand", "quotesingle",
	"parenleft", "parenright", "asterisk", "plus", "comma", "hyphen",
	"period", "slash", "zero", "one", "two", "three", "four", "five",
	"six", "seven", "eight", "nine", "colon", "semicolon", "less",
	"equal", "greater", "question", "at", "A", "B", "C", "D", "E", "F",
	"G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S",
	"T", "U", "V", "W", "X", "Y", "Z", "bracketleft", "backslash",
	"bracketright", "asciicircum", "underscore", "grave", "a", "b",
	"c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o",
	"p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "braceleft",
	"bar", "braceright", "asciitilde", "Adieresis", "Aring",
	"Ccedilla", "Eacute", "Ntilde", "Odieresis", "Udieresis", "aacute",
	"agrave", "acircumflex", "adieresis", "atilde", "aring",
	"ccedilla", "eacute", "egrave", "ecircumflex", "edieresis",
	"iacute", "igrave", "icircumflex", "idieresis", "ntilde", "oacute",
	"ograve", "ocircumflex", "odieresis", "otilde", "uacute", "ugrave",
	"ucircumflex", "udieresis", "dagger", "degree", "cent", "sterling",
	"section", "bullet", "paragraph", "germandbls", "registered",
	"copyright", "trademark", "acute", "dieresis", "notequal", "AE",
	"Oslash", "infinity", "plusminus", "lessequal", "greaterequal",
	"yen", "mu", "partialdiff", "summation", "product", "pi",
	"integral", "ordfeminine", "ordmasculine", "Omega", "ae", "oslash",
	"questiondown", "exclamdown", "logicalnot", "radical", "florin",
	"approxequal", "Delta", "guillemotleft", "guillemotright",
	"ellipsis", "nonbreakingspace", "Agrave", "Atilde", "Otilde", "OE",
	"oe", "endash", "emdash", "quotedblleft", "quotedblright",
	"quoteleft", "quoteright", "divide", "lozenge", "ydieresis",
	"Ydieresis", "fraction", "currency", "guilsinglleft",
	"guilsinglright", "fi", "fl", "daggerdbl", "periodcentered",
	"quotesinglbase", "quotedblbase", "perthousand", "Acircumflex",
	"Ecircumflex", "Aacute", "Edieresis", "Egrave", "Iacute",
	"Icircumflex", "Idieresis", "Igrave", "Oacute", "Ocircumflex",
	"apple", "Ograve", "Uacute", "Ucircumflex", "Ugrave", "dotlessi",
	"circumflex", "tilde", "macron", "breve", "dotaccent", "ring",
	"cedilla", "hungarumlaut", "ogonek", "caron", "Lslash", "lslash",
	"Scaron", "scaron", "Zcaron", "zcaron", "brokenbar", "Eth", "eth",
	"Yacute", "yacute", "Thorn", "thorn", "minus", "multiply",
	"onesuperior", "twosuperior", "threesuperior", "onehalf",
	"onequarter", "threequarters", "franc", "Gbreve", "gbreve",
	"Idotaccent", "Scedilla", "scedilla", "Cacute", "cacute", "Ccaron",
	"ccaron", "dcroat",
}

// Seek moves the file pointer to the table named `tag`.
func (t *ttfParser) Seek(tag string) error {
	ofs, ok := t.tables[tag]
	if !ok {
		return fmt.Errorf("table not found: %s", tag)
	}
	t.f.Seek(int64(ofs), os.SEEK_SET)
	return nil
}

// Skip moves the file point n bytes forward.
func (t *ttfParser) Skip(n int) {
	t.f.Seek(int64(n), os.SEEK_CUR)
}

// ReadStr reads `length` bytes from the file and returns them as a string, or an error if there was
// a problem.
func (t *ttfParser) ReadStr(length int) (string, error) {
	buf := make([]byte, length)
	n, err := t.f.Read(buf)
	if err != nil {
		return "", err
	} else if n != length {
		return "", fmt.Errorf("unable to read %d bytes", length)
	}
	return string(buf), nil
}

// readByte reads a byte and returns it as unsigned.
func (t *ttfParser) readByte() (val uint8) {
	binary.Read(t.f, binary.BigEndian, &val)
	return val
}

// ReadSByte reads a byte and returns it as signed.
func (t *ttfParser) ReadSByte() (val int8) {
	binary.Read(t.f, binary.BigEndian, &val)
	return val
}

// ReadUShort reads 2 bytes and returns them as a big endian unsigned 16 bit integer.
func (t *ttfParser) ReadUShort() (val uint16) {
	binary.Read(t.f, binary.BigEndian, &val)
	return val
}

// ReadShort reads 2 bytes and returns them as a big endian signed 16 bit integer.
func (t *ttfParser) ReadShort() (val int16) {
	binary.Read(t.f, binary.BigEndian, &val)
	return val
}

// ReadULong reads 4 bytes and returns them as a big endian unsigned 32 bit integer.
func (t *ttfParser) ReadULong() (val uint32) {
	binary.Read(t.f, binary.BigEndian, &val)
	return val
}

// Read32Fixed reads 4 bytes and returns them as a float, the first 2 bytes for the whole number and
// the second 2 bytes for the fraction.
func (t *ttfParser) Read32Fixed() float64 {
	whole := float64(t.ReadShort())
	frac := float64(t.ReadUShort()) / 65536.0
	return whole + frac
}
