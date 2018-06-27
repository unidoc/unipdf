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

package fonts

// Utility to parse TTF font files
// Version:    1.0
// Date:       2011-06-18
// Author:     Olivier PLATHEY
// Port to Go: Kurt Jung, 2013-07-15

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TtfType contains metrics of a TrueType font.
type TtfType struct {
	Embeddable             bool
	UnitsPerEm             uint16
	PostScriptName         string
	Bold                   bool
	ItalicAngle            int16
	IsFixedPitch           bool
	TypoAscender           int16
	TypoDescender          int16
	UnderlinePosition      int16
	UnderlineThickness     int16
	Xmin, Ymin, Xmax, Ymax int16
	CapHeight              int16
	Widths                 []uint16

	// Map of rune values (unicode) to glyph index.
	Chars map[uint16]uint16
}

type ttfParser struct {
	rec              TtfType
	f                *os.File
	tables           map[string]uint32
	numberOfHMetrics uint16
	numGlyphs        uint16
}

// TtfParse extracts various metrics from a TrueType font file.
func TtfParse(fileStr string) (TtfRec TtfType, err error) {
	var t ttfParser
	t.f, err = os.Open(fileStr)
	if err != nil {
		return
	}
	defer t.f.Close()

	version, err := t.ReadStr(4)
	if err != nil {
		return
	}
	if version == "OTTO" {
		err = fmt.Errorf("fonts based on PostScript outlines are not supported")
		return
	}
	if version != "\x00\x01\x00\x00" {
		err = fmt.Errorf("unrecognized file format")
		return
	}
	numTables := int(t.ReadUShort())
	t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
	t.tables = make(map[string]uint32)
	var tag string
	for j := 0; j < numTables; j++ {
		tag, err = t.ReadStr(4)
		if err != nil {
			return
		}
		t.Skip(4) // checkSum
		offset := t.ReadULong()
		t.Skip(4) // length
		t.tables[tag] = offset
	}
	err = t.ParseComponents()
	if err != nil {
		return
	}

	TtfRec = t.rec
	return
}

func (t *ttfParser) ParseComponents() (err error) {
	err = t.ParseHead()
	if err == nil {
		err = t.ParseHhea()
		if err == nil {
			err = t.ParseMaxp()
			if err == nil {
				err = t.ParseHmtx()
				if err == nil {
					err = t.ParseCmap()
					if err == nil {
						err = t.ParseName()
						if err == nil {
							err = t.ParseOS2()
							if err == nil {
								err = t.ParsePost()
							}
						}
					}
				}
			}
		}
	}
	return
}

func (t *ttfParser) ParseHead() (err error) {
	err = t.Seek("head")
	t.Skip(3 * 4) // version, fontRevision, checkSumAdjustment
	magicNumber := t.ReadULong()
	if magicNumber != 0x5F0F3CF5 {
		err = fmt.Errorf("incorrect magic number")
		return
	}
	t.Skip(2) // flags
	t.rec.UnitsPerEm = t.ReadUShort()
	t.Skip(2 * 8) // created, modified
	t.rec.Xmin = t.ReadShort()
	t.rec.Ymin = t.ReadShort()
	t.rec.Xmax = t.ReadShort()
	t.rec.Ymax = t.ReadShort()
	return
}

func (t *ttfParser) ParseHhea() (err error) {
	err = t.Seek("hhea")
	if err == nil {
		t.Skip(4 + 15*2)
		t.numberOfHMetrics = t.ReadUShort()
	}
	return
}

func (t *ttfParser) ParseMaxp() (err error) {
	err = t.Seek("maxp")
	if err == nil {
		t.Skip(4)
		t.numGlyphs = t.ReadUShort()
	}
	return
}

func (t *ttfParser) ParseHmtx() (err error) {
	err = t.Seek("hmtx")
	if err == nil {
		t.rec.Widths = make([]uint16, 0, 8)
		for j := uint16(0); j < t.numberOfHMetrics; j++ {
			t.rec.Widths = append(t.rec.Widths, t.ReadUShort())
			t.Skip(2) // lsb
		}
		if t.numberOfHMetrics < t.numGlyphs {
			lastWidth := t.rec.Widths[t.numberOfHMetrics-1]
			for j := t.numberOfHMetrics; j < t.numGlyphs; j++ {
				t.rec.Widths = append(t.rec.Widths, lastWidth)
			}
		}
	}
	return
}

// parseCmapSubtable31 parses information from an (3,1) subtable (Windows Unicode).
func (t *ttfParser) parseCmapSubtable31(offset31 int64) (err error) {
	startCount := make([]uint16, 0, 8)
	endCount := make([]uint16, 0, 8)
	idDelta := make([]int16, 0, 8)
	idRangeOffset := make([]uint16, 0, 8)
	t.rec.Chars = make(map[uint16]uint16)
	t.f.Seek(int64(t.tables["cmap"])+offset31, os.SEEK_SET)
	format := t.ReadUShort()
	if format != 4 {
		err = fmt.Errorf("unexpected subtable format: %d", format)
		return
	}
	t.Skip(2 * 2) // length, language
	segCount := int(t.ReadUShort() / 2)
	t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
	for j := 0; j < segCount; j++ {
		endCount = append(endCount, t.ReadUShort())
	}
	t.Skip(2) // reservedPad
	for j := 0; j < segCount; j++ {
		startCount = append(startCount, t.ReadUShort())
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
				t.rec.Chars[c] = uint16(gid)
			}
		}
	}
	return
}

// parseCmapSubtable10 parses information from an (1,0) subtable (symbol).
func (t *ttfParser) parseCmapSubtable10(offset10 int64) error {

	//startCount := make([]uint16, 0, 8)
	//endCount := make([]uint16, 0, 8)
	//idDelta := make([]int16, 0, 8)
	//idRangeOffset := make([]uint16, 0, 8)

	if t.rec.Chars == nil {
		t.rec.Chars = make(map[uint16]uint16)
	}

	//t.rec.Chars = make(map[uint16]uint16)

	t.f.Seek(int64(t.tables["cmap"])+offset10, os.SEEK_SET)
	format := t.ReadUShort()
	if format != 6 {
		return fmt.Errorf("unexpected subtable format: %d", format)
	}
	length := t.ReadUShort()
	language := t.ReadUShort()
	firstCode := t.ReadUShort()
	entryCount := t.ReadUShort()

	curCode := firstCode
	for i := 0; i < int(entryCount); i++ {
		glyphId := t.ReadUShort()
		t.rec.Chars[curCode] = glyphId
		fmt.Printf("%d -> %d\n", curCode, glyphId)

		curCode++
	}

	fmt.Printf("Length: %d, language: %d\n", length, language)
	fmt.Printf("First code: %d, entry count: %d\n", firstCode, entryCount)
	return nil
}

func (t *ttfParser) ParseCmap() (err error) {
	var offset int64
	if err = t.Seek("cmap"); err != nil {
		return
	}
	t.Skip(2) // version
	numTables := int(t.ReadUShort())
	//offset10 := int64(0)
	offset31 := int64(0)
	for j := 0; j < numTables; j++ {
		platformID := t.ReadUShort()
		encodingID := t.ReadUShort()
		offset = int64(t.ReadULong())
		if platformID == 3 && encodingID == 1 {
			// (3,1) subtable. Windows Unicode.
			offset31 = offset
		} /*else if platformID == 1 && encodingID == 0 {
			// (1,0) subtable.
			offset10 = offset

		}*/
		//fmt.Printf("(%d,%d) subtable @ %d\n", platformID, encodingID, offset)
	}

	// Latin font support based on (3,1) table encoding.
	if offset31 != 0 {
		err = t.parseCmapSubtable31(offset31)
		if err != nil {
			return
		}
	}

	// Many non-Latin fonts (including asian fonts) use subtable (1,0).
	/*
		if offset10 != 0 {
			fmt.Printf("Offset 10: %d\n", offset10)
			err = t.parseCmapSubtable10(offset10)
			if err != nil {
				return
			}
		}
	*/

	return
}

func (t *ttfParser) ParseName() (err error) {
	if err = t.Seek("name"); err != nil {
		return
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
			var s string
			s, err = t.ReadStr(int(length))
			if err != nil {
				return
			}
			s = strings.Replace(s, "\x00", "", -1)
			var re *regexp.Regexp
			if re, err = regexp.Compile("[(){}<> /%[\\]]"); err != nil {
				return
			}
			t.rec.PostScriptName = re.ReplaceAllString(s, "")
		}
	}
	if t.rec.PostScriptName == "" {
		err = fmt.Errorf("the name PostScript was not found")
	}
	return
}

func (t *ttfParser) ParseOS2() (err error) {
	if err = t.Seek("OS/2"); err != nil {
		return
	}
	version := t.ReadUShort()
	t.Skip(3 * 2) // xAvgCharWidth, usWeightClass, usWidthClass
	fsType := t.ReadUShort()
	t.rec.Embeddable = (fsType != 2) && (fsType&0x200) == 0
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
	return
}

func (t *ttfParser) ParsePost() (err error) {
	if err = t.Seek("post"); err != nil {
		return
	}
	//versionUpper := t.ReadShort()
	//versionFraction := t.ReadUShort()

	t.Skip(4) // version
	//fmt.Printf("Post version: %d.%d\n", versionUpper, versionFraction)

	t.rec.ItalicAngle = t.ReadShort()
	t.Skip(2) // Skip decimal part
	t.rec.UnderlinePosition = t.ReadShort()
	t.rec.UnderlineThickness = t.ReadShort()
	t.rec.IsFixedPitch = t.ReadULong() != 0

	/*
		t.Skip(4 * 4) // Skip over memory specs.

		if versionUpper == 2 && versionFraction == 0 {
			numGlyps := t.ReadUShort()

			glyphNameIndexList := []uint16{}
			maxIndex := uint16(0)
			for i := 0; i < int(numGlyps); i++ {
				index := t.ReadUShort()
				glyphNameIndexList = append(glyphNameIndexList, index)
				if index > maxIndex {
					maxIndex = index
				}
			}

			numberNewGlyphs := maxIndex + 1
			if maxIndex > 257 {
				numberNewGlyphs -= 258
			}
			for i := 0; i < int(numberNewGlyphs); i++ {
				len := t.ReadByte()
				glyphName, err := t.ReadStr(int(len))
				if err != nil {
					return err
				}
			}
		}
	*/
	return
}

func (t *ttfParser) Seek(tag string) error {
	ofs, ok := t.tables[tag]
	if !ok {
		return fmt.Errorf("table not found: %s", tag)
	}
	t.f.Seek(int64(ofs), os.SEEK_SET)
	return nil
}

func (t *ttfParser) Skip(n int) {
	t.f.Seek(int64(n), os.SEEK_CUR)
}

func (t *ttfParser) ReadStr(length int) (str string, err error) {
	var n int
	buf := make([]byte, length)
	n, err = t.f.Read(buf)
	if err != nil {
		return
	}
	if n == length {
		str = string(buf)
	} else {
		err = fmt.Errorf("unable to read %d bytes", length)
	}
	return
}

func (t *ttfParser) ReadByte() (val uint8) {
	binary.Read(t.f, binary.BigEndian, &val)
	return
}

func (t *ttfParser) ReadUShort() (val uint16) {
	binary.Read(t.f, binary.BigEndian, &val)
	return
}

func (t *ttfParser) ReadShort() (val int16) {
	binary.Read(t.f, binary.BigEndian, &val)
	return
}

func (t *ttfParser) ReadULong() (val uint32) {
	binary.Read(t.f, binary.BigEndian, &val)
	return
}
