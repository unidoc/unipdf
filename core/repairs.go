/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Routines related to repairing malformed pdf files.

package core

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"bufio"
	"io"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
)

var repairReXrefTable = regexp.MustCompile(`[\r\n]\s*(xref)\s*[\r\n]`)

// Locates a standard Xref table by looking for the "xref" entry.
// Xref object stream not supported.
func (parser *PdfParser) repairLocateXref() (int64, error) {
	readBuf := int64(1000)
	parser.rs.Seek(-readBuf, os.SEEK_CUR)

	curOffset, err := parser.rs.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}
	b2 := make([]byte, readBuf)
	parser.rs.Read(b2)

	results := repairReXrefTable.FindAllStringIndex(string(b2), -1)
	if len(results) < 1 {
		common.Log.Debug("ERROR: Repair: xref not found!")
		return 0, errors.New("repair: xref not found")
	}

	localOffset := int64(results[len(results)-1][0])
	xrefOffset := curOffset + localOffset
	return xrefOffset, nil
}

// Renumbers the xref table.
// Useful when the cross reference is pointing to an object with the wrong number.
// Update the table.
func (parser *PdfParser) rebuildXrefTable() error {
	newXrefs := XrefTable{}
	newXrefs.ObjectMap = map[int]XrefObject{}
	for objNum, xref := range parser.xrefs.ObjectMap {
		obj, _, err := parser.lookupByNumberWrapper(objNum, false)
		if err != nil {
			common.Log.Debug("ERROR: Unable to look up object (%s)", err)
			common.Log.Debug("ERROR: Xref table completely broken - attempting to repair ")
			xrefTable, err := parser.repairRebuildXrefsTopDown()
			if err != nil {
				common.Log.Debug("ERROR: Failed xref rebuild repair (%s)", err)
				return err
			}
			parser.xrefs = *xrefTable
			common.Log.Debug("Repaired xref table built")
			return nil
		}
		actObjNum, actGenNum, err := getObjectNumber(obj)
		if err != nil {
			return err
		}

		xref.ObjectNumber = int(actObjNum)
		xref.Generation = int(actGenNum)
		newXrefs.ObjectMap[int(actObjNum)] = xref
	}

	parser.xrefs = newXrefs
	common.Log.Debug("New xref table built")
	printXrefTable(parser.xrefs)
	return nil
}

// Parses and returns the object and generation number from a string such as "12 0 obj" -> (12,0,nil).
func parseObjectNumberFromString(str string) (int, int, error) {
	result := reIndirectObject.FindStringSubmatch(str)
	if len(result) < 3 {
		return 0, 0, errors.New("unable to detect indirect object signature")
	}

	on, _ := strconv.Atoi(result[1])
	gn, _ := strconv.Atoi(result[2])

	return on, gn, nil
}

// Parse the entire file from top down.
// Goes through the file byte-by-byte looking for "<num> <generation> obj" patterns.
// N.B. This collects the XrefTypeTableEntry data only.
func (parser *PdfParser) repairRebuildXrefsTopDown() (*XrefTable, error) {
	if parser.repairsAttempted {
		// Avoid multiple repairs (only try once).
		return nil, fmt.Errorf("repair failed")
	}
	parser.repairsAttempted = true

	// Go to beginning, reset reader.
	parser.rs.Seek(0, os.SEEK_SET)
	parser.reader = bufio.NewReader(parser.rs)

	// Keep a running buffer of last bytes.
	bufLen := 20
	last := make([]byte, bufLen)

	xrefTable := XrefTable{}
	xrefTable.ObjectMap = make(map[int]XrefObject)
	for {
		b, err := parser.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		// Format:
		// object number - whitespace - generation number - obj
		// e.g. "12 0 obj"
		if b == 'j' && last[bufLen-1] == 'b' && last[bufLen-2] == 'o' && IsWhiteSpace(last[bufLen-3]) {
			i := bufLen - 4
			// Go past whitespace
			for IsWhiteSpace(last[i]) && i > 0 {
				i--
			}
			if i == 0 || !IsDecimalDigit(last[i]) {
				continue
			}
			// Go past generation number
			for IsDecimalDigit(last[i]) && i > 0 {
				i--
			}
			if i == 0 || !IsWhiteSpace(last[i]) {
				continue
			}
			// Go past whitespace
			for IsWhiteSpace(last[i]) && i > 0 {
				i--
			}
			if i == 0 || !IsDecimalDigit(last[i]) {
				continue
			}
			// Go past object number.
			for IsDecimalDigit(last[i]) && i > 0 {
				i--
			}
			if i == 0 {
				continue // Probably too long to be a valid object...
			}

			objOffset := parser.GetFileOffset() - int64(bufLen-i)

			objstr := append(last[i+1:], b)
			objNum, genNum, err := parseObjectNumberFromString(string(objstr))
			if err != nil {
				common.Log.Debug("Unable to parse object number: %v", err)
				return nil, err
			}

			// Create and insert the XREF entry if not existing, or the generation number is higher.
			if curXref, has := xrefTable.ObjectMap[objNum]; !has || curXref.Generation < genNum {
				// Make the entry for the cross ref table.
				xrefEntry := XrefObject{}
				xrefEntry.XType = XrefTypeTableEntry
				xrefEntry.ObjectNumber = int(objNum)
				xrefEntry.Generation = int(genNum)
				xrefEntry.Offset = objOffset
				xrefTable.ObjectMap[objNum] = xrefEntry
			}
		}

		last = append(last[1:bufLen], b)
	}

	return &xrefTable, nil
}

// Look for first sign of xref table from end of file.
func (parser *PdfParser) repairSeekXrefMarker() error {
	// Get the file size.
	fSize, err := parser.rs.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}

	reXrefTableStart := regexp.MustCompile(`\sxref\s*`)

	// Define the starting point (from the end of the file) to search from.
	var offset int64

	// Define an buffer length in terms of how many bytes to read from the end of the file.
	var buflen int64 = 1000

	for offset < fSize {
		if fSize <= (buflen + offset) {
			buflen = fSize - offset
		}

		// Move back enough (as we need to read forward).
		_, err := parser.rs.Seek(-offset-buflen, os.SEEK_END)
		if err != nil {
			return err
		}

		// Read the data.
		b1 := make([]byte, buflen)
		parser.rs.Read(b1)

		common.Log.Trace("Looking for xref : \"%s\"", string(b1))
		ind := reXrefTableStart.FindAllStringIndex(string(b1), -1)
		if ind != nil {
			// Found it.
			lastInd := ind[len(ind)-1]
			common.Log.Trace("Ind: % d", ind)
			parser.rs.Seek(-offset-buflen+int64(lastInd[0]), os.SEEK_END)
			parser.reader = bufio.NewReader(parser.rs)
			// Go past whitespace, finish at 'x'.
			for {
				bb, err := parser.reader.Peek(1)
				if err != nil {
					return err
				}
				common.Log.Trace("B: %d %c", bb[0], bb[0])
				if !IsWhiteSpace(bb[0]) {
					break
				}
				parser.reader.Discard(1)
			}

			return nil
		}

		common.Log.Debug("Warning: EOF marker not found! - continue seeking")
		offset += buflen
	}

	common.Log.Debug("Error: Xref table marker was not found.")
	return errors.New("xref not found ")
}

// Called when Pdf version not found normally.  Looks for the PDF version by scanning top-down.
// %PDF-1.7
func (parser *PdfParser) seekPdfVersionTopDown() (int, int, error) {
	// Go to beginning, reset reader.
	parser.rs.Seek(0, os.SEEK_SET)
	parser.reader = bufio.NewReader(parser.rs)

	// Keep a running buffer of last bytes.
	bufLen := 20
	last := make([]byte, bufLen)

	for {
		b, err := parser.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return 0, 0, err
			}
		}

		// Format:
		// object number - whitespace - generation number - obj
		// e.g. "12 0 obj"
		if IsDecimalDigit(b) && last[bufLen-1] == '.' && IsDecimalDigit(last[bufLen-2]) && last[bufLen-3] == '-' &&
			last[bufLen-4] == 'F' && last[bufLen-5] == 'D' && last[bufLen-6] == 'P' {
			major := int(last[bufLen-2] - '0')
			minor := int(b - '0')
			return major, minor, nil
		}

		last = append(last[1:bufLen], b)
	}

	return 0, 0, errors.New("version not found")
}
