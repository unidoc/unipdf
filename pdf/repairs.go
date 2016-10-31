/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Routines related to repairing malformed pdf files.

package pdf

import (
	"errors"
	"fmt"
	"os"
	"regexp"

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

// Parse the entire file from top down.
// Currently not supporting object streams...
// Also need to detect object streams and load the object numbers.
func (this *PdfParser) repairRebuildXrefsTopDown() (*XrefTable, error) {
	if this.repairsAttempted {
		// Avoid multiple repairs (only try once).
		return nil, fmt.Errorf("Repair failed (multiple trials)")
	}
	this.repairsAttempted = true

	reRepairIndirectObject := regexp.MustCompile(`^(\d+)\s+(\d+)\s+obj`)

	this.SetFileOffset(0)

	xrefTable := XrefTable{}
	for {
		this.skipComments()

		curOffset := this.GetFileOffset()

		peakBuf, err := this.reader.Peek(10)
		if err != nil {
			// EOF
			break
		}

		// Indirect object?
		results := reRepairIndirectObject.FindIndex(peakBuf)
		if len(results) > 0 {
			obj, err := this.parseIndirectObject()
			if err != nil {
				common.Log.Debug("ERROR: Unable to parse indirect object (%s)", err)
				return nil, err
			}

			if indObj, ok := obj.(*PdfIndirectObject); ok {
				// Make the entry for the cross ref table.
				xrefEntry := XrefObject{}
				xrefEntry.xtype = XREF_TABLE_ENTRY
				xrefEntry.objectNumber = int(indObj.ObjectNumber)
				xrefEntry.generation = int(indObj.GenerationNumber)
				xrefEntry.offset = curOffset
				xrefTable[int(indObj.ObjectNumber)] = xrefEntry
			} else if streamObj, ok := obj.(*PdfObjectStream); ok {
				// Make the entry for the cross ref table.
				xrefEntry := XrefObject{}
				xrefEntry.xtype = XREF_TABLE_ENTRY
				xrefEntry.objectNumber = int(streamObj.ObjectNumber)
				xrefEntry.generation = int(streamObj.GenerationNumber)
				xrefEntry.offset = curOffset
				xrefTable[int(streamObj.ObjectNumber)] = xrefEntry
			} else {
				return nil, fmt.Errorf("Not an indirect object or stream (%T)", obj) // Should never happen.
			}
		} else if string(peakBuf[0:6]) == "endobj" {
			this.reader.Discard(6)
		} else {
			// Stop once we reach xrefs/trailer section etc.  Technically this could fail for complex
			// cases, but lets keep it simple for now.  Add more complexity when needed (problematic user committed files).
			// In general more likely that more complex files would have better understanding of the PDF standard.
			common.Log.Debug("Not an object - stop repair rebuilding xref here (%s)", peakBuf)
			break
		}
	}

	return &xrefTable, nil
}
