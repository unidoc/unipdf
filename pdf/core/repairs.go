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

	"github.com/unidoc/unidoc/common"
)

var repairReXrefTable = regexp.MustCompile(`[\r\n]\s*(xref)\s*[\r\n]`)

// Locates a standard Xref table by looking for the "xref" entry.
// Xref object stream not supported.
func (this *PdfParser) repairLocateXref() (int64, error) {
	readBuf := int64(1000)
	this.rs.Seek(-readBuf, os.SEEK_CUR)

	curOffset, err := this.rs.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}
	b2 := make([]byte, readBuf)
	this.rs.Read(b2)

	results := repairReXrefTable.FindAllStringIndex(string(b2), -1)
	if len(results) < 1 {
		common.Log.Debug("ERROR: Repair: xref not found!")
		return 0, errors.New("Repair: xref not found")
	}

	localOffset := int64(results[len(results)-1][0])
	xrefOffset := curOffset + localOffset
	return xrefOffset, nil
}

// Renumbers the xref table.
// Useful when the cross reference is pointing to an object with the wrong number.
// Update the table.
func (this *PdfParser) rebuildXrefTable() error {
	newXrefs := XrefTable{}
	for objNum, xref := range this.xrefs {
		obj, _, err := this.lookupByNumberWrapper(objNum, false)
		if err != nil {
			common.Log.Debug("ERROR: Unable to look up object (%s)", err)
			common.Log.Debug("ERROR: Xref table completely broken - attempting to repair ")
			xrefTable, err := this.repairRebuildXrefsTopDown()
			if err != nil {
				common.Log.Debug("ERROR: Failed xref rebuild repair (%s)", err)
				return err
			}
			this.xrefs = *xrefTable
			common.Log.Debug("Repaired xref table built")
			return nil
		}
		actObjNum, actGenNum, err := getObjectNumber(obj)
		if err != nil {
			return err
		}

		xref.objectNumber = int(actObjNum)
		xref.generation = int(actGenNum)
		newXrefs[int(actObjNum)] = xref
	}

	this.xrefs = newXrefs
	common.Log.Debug("New xref table built")
	printXrefTable(this.xrefs)
	return nil
}

// Parses and returns the object and generation number from a string such as "12 0 obj" -> (12,0,nil).
func parseObjectNumberFromString(str string) (int, int, error) {
	result := reIndirectObject.FindStringSubmatch(str)
	if len(result) < 3 {
		return 0, 0, errors.New("Unable to detect indirect object signature")
	}

	on, _ := strconv.Atoi(result[1])
	gn, _ := strconv.Atoi(result[2])

	return on, gn, nil
}

// Parse the entire file from top down.
// Goes through the file byte-by-byte looking for "<num> <generation> obj" patterns.
// N.B. This collects the XREF_TABLE_ENTRY data only.
func (this *PdfParser) repairRebuildXrefsTopDown() (*XrefTable, error) {
	if this.repairsAttempted {
		// Avoid multiple repairs (only try once).
		return nil, fmt.Errorf("Repair failed")
	}
	this.repairsAttempted = true

	// Go to beginning, reset reader.
	this.rs.Seek(0, os.SEEK_SET)
	this.reader = bufio.NewReader(this.rs)

	// Keep a running buffer of last bytes.
	bufLen := 20
	last := make([]byte, bufLen)

	xrefTable := XrefTable{}
	for {
		b, err := this.reader.ReadByte()
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

			objOffset := this.GetFileOffset() - int64(bufLen-i)

			objstr := append(last[i+1:], b)
			objNum, genNum, err := parseObjectNumberFromString(string(objstr))
			if err != nil {
				common.Log.Debug("Unable to parse object number: %v", err)
				return nil, err
			}

			// Create and insert the XREF entry if not existing, or the generation number is higher.
			if curXref, has := xrefTable[objNum]; !has || curXref.generation < genNum {
				// Make the entry for the cross ref table.
				xrefEntry := XrefObject{}
				xrefEntry.xtype = XREF_TABLE_ENTRY
				xrefEntry.objectNumber = int(objNum)
				xrefEntry.generation = int(genNum)
				xrefEntry.offset = objOffset
				xrefTable[objNum] = xrefEntry
			}
		}

		last = append(last[1:bufLen], b)
	}

	return &xrefTable, nil
}
