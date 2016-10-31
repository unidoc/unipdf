/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/unidoc/unidoc/common"
)

const (
	XREF_TABLE_ENTRY   = iota
	XREF_OBJECT_STREAM = iota
)

// Either can be in a normal xref table, or in an xref stream.
// Can point either to a file offset, or to an object stream.
// XrefFileOffset or XrefObjectStream...

type XrefObject struct {
	xtype        int
	objectNumber int
	generation   int
	// For normal xrefs (defined by OFFSET)
	offset int64
	// For xrefs to object streams.
	osObjNumber int
	osObjIndex  int
}

type XrefTable map[int]XrefObject

type ObjectStream struct {
	N       int
	ds      []byte
	offsets map[int]int64
}

type ObjectStreams map[int]ObjectStream

type ObjectCache map[int]PdfObject

// Get an object from an object stream.
func (this *PdfParser) lookupObjectViaOS(sobjNumber int, objNum int) (PdfObject, error) {
	var bufReader *bytes.Reader
	var objstm ObjectStream
	var cached bool

	objstm, cached = this.objstms[sobjNumber]
	if !cached {
		soi, err := this.LookupByNumber(sobjNumber)
		if err != nil {
			common.Log.Debug("Missing object stream with number %d", sobjNumber)
			return nil, err
		}

		so, ok := soi.(*PdfObjectStream)
		if !ok {
			return nil, errors.New("Invalid object stream")
		}

		if this.crypter != nil && !this.crypter.isDecrypted(so) {
			return nil, errors.New("Need to decrypt the stream !")
		}

		sod := so.PdfObjectDictionary
		common.Log.Debug("so d: %s\n", *sod)
		name, ok := (*sod)["Type"].(*PdfObjectName)
		if !ok {
			common.Log.Debug("ERROR: Object stream should always have a Type")
			return nil, errors.New("Object stream missing Type")
		}
		if strings.ToLower(string(*name)) != "objstm" {
			common.Log.Debug("ERROR: Object stream type shall always be ObjStm !")
			return nil, errors.New("Object stream type != ObjStm")
		}

		N, ok := (*sod)["N"].(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid N in stream dictionary")
		}
		firstOffset, ok := (*sod)["First"].(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid First in stream dictionary")
		}

		common.Log.Debug("type: %s number of objects: %d", name, *N)
		ds, err := this.decodeStream(so)
		if err != nil {
			return nil, err
		}

		common.Log.Debug("Decoded: %s", ds)

		// Temporarily change the reader object to this decoded buffer.
		// Change back afterwards.
		bakOffset := this.GetFileOffset()
		defer func() { this.SetFileOffset(bakOffset) }()

		bufReader = bytes.NewReader(ds)
		this.reader = bufio.NewReader(bufReader)

		common.Log.Debug("Parsing offset map")
		// Load the offset map (relative to the beginning of the stream...)
		var offsets map[int]int64 = make(map[int]int64)
		// Object list and offsets.
		for i := 0; i < int(*N); i++ {
			this.skipSpaces()
			// Object number.
			obj, err := this.parseNumber()
			if err != nil {
				return nil, err
			}
			onum, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, errors.New("Invalid object stream offset table")
			}

			this.skipSpaces()
			// Offset.
			obj, err = this.parseNumber()
			if err != nil {
				return nil, err
			}
			offset, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, errors.New("Invalid object stream offset table")
			}

			common.Log.Debug("obj %d offset %d", *onum, *offset)
			offsets[int(*onum)] = int64(*firstOffset + *offset)
		}

		objstm = ObjectStream{N: int(*N), ds: ds, offsets: offsets}
		this.objstms[sobjNumber] = objstm
	} else {
		// Temporarily change the reader object to this decoded buffer.
		// Point back afterwards.
		bakOffset := this.GetFileOffset()
		defer func() { this.SetFileOffset(bakOffset) }()

		bufReader = bytes.NewReader(objstm.ds)
		// Temporarily change the reader object to this decoded buffer.
		this.reader = bufio.NewReader(bufReader)
	}

	offset := objstm.offsets[objNum]
	common.Log.Debug("ACTUAL offset[%d] = %d", objNum, offset)

	bufReader.Seek(offset, os.SEEK_SET)
	this.reader = bufio.NewReader(bufReader)

	bb, _ := this.reader.Peek(100)
	common.Log.Debug("OBJ peek \"%s\"", string(bb))

	val, err := this.parseObject()
	if err != nil {
		common.Log.Debug("ERROR Fail to read object (%s)", err)
		return nil, err
	}
	if val == nil {
		return nil, errors.New("Object cannot be null")
	}

	// Make an indirect object around it.
	io := PdfIndirectObject{}
	io.ObjectNumber = int64(objNum)
	io.PdfObject = val

	return &io, nil
}

// Currently a bit messy.. multiple wrappers.  Can we clean up?

// Outside interface for lookupByNumberWrapper.  Default attempts
// repairs of bad xref tables.
func (this *PdfParser) LookupByNumber(objNumber int) (PdfObject, error) {
	obj, _, err := this.lookupByNumberWrapper(objNumber, true)
	return obj, err
}

// Wrapper for lookupByNumber, checks if object encrypted etc.
func (this *PdfParser) lookupByNumberWrapper(objNumber int, attemptRepairs bool) (PdfObject, bool, error) {
	obj, inObjStream, err := this.lookupByNumber(objNumber, attemptRepairs)
	if err != nil {
		return nil, inObjStream, err
	}

	// If encrypted, decrypt it prior to returning.
	// Do not attempt to decrypt objects within object streams.
	if !inObjStream && this.crypter != nil && !this.crypter.isDecrypted(obj) {
		err := this.crypter.Decrypt(obj, 0, 0)
		if err != nil {
			return nil, inObjStream, err
		}
	}

	return obj, inObjStream, nil
}

func getObjectNumber(obj PdfObject) (int64, int64, error) {
	if io, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		return io.ObjectNumber, io.GenerationNumber, nil
	}
	if so, isStream := obj.(*PdfObjectStream); isStream {
		return so.ObjectNumber, so.GenerationNumber, nil
	}
	return 0, 0, errors.New("Not an indirect/stream object")
}

// LookupByNumber
// Repair signals whether to repair if broken.
func (this *PdfParser) lookupByNumber(objNumber int, attemptRepairs bool) (PdfObject, bool, error) {
	obj, ok := this.ObjCache[objNumber]
	if ok {
		common.Log.Debug("Returning cached object %d", objNumber)
		return obj, false, nil
	}

	xref, ok := this.xrefs[objNumber]
	if !ok {
		// An indirect reference to an undefined object shall not be
		// considered an error by a conforming reader; it shall be
		// treated as a reference to the null object.
		common.Log.Debug("Unable to locate object in xrefs! - Returning null object")
		var nullObj PdfObjectNull
		return &nullObj, false, nil
	}

	common.Log.Debug("Lookup obj number %d", objNumber)
	if xref.xtype == XREF_TABLE_ENTRY {
		common.Log.Debug("xrefobj obj num %d", xref.objectNumber)
		common.Log.Debug("xrefobj gen %d", xref.generation)
		common.Log.Debug("xrefobj offset %d", xref.offset)

		this.rs.Seek(xref.offset, os.SEEK_SET)
		this.reader = bufio.NewReader(this.rs)

		obj, err := this.parseIndirectObject()
		if err != nil {
			common.Log.Debug("ERROR Failed reading xref (%s)", err)
			// Offset pointing to a non-object.  Try to repair the file.
			if attemptRepairs {
				common.Log.Debug("Attempting to repair xrefs (top down)")
				xrefTable, err := this.repairRebuildXrefsTopDown()
				if err != nil {
					common.Log.Debug("ERROR Failed repair (%s)", err)
					return nil, false, err
				}
				this.xrefs = *xrefTable
				return this.lookupByNumber(objNumber, false)
			}
			return nil, false, err
		}

		if attemptRepairs {
			// Check the object number..
			// If it does not match, then try to rebuild, i.e. loop through
			// all the items in the xref and look each one up and correct.
			realObjNum, _, _ := getObjectNumber(obj)
			if int(realObjNum) != objNumber {
				common.Log.Debug("Invalid xrefs: Rebuilding")
				err := this.rebuildXrefTable()
				if err != nil {
					return nil, false, err
				}
				// Empty the cache.
				this.ObjCache = ObjectCache{}
				// Try looking up again and return.
				return this.lookupByNumberWrapper(objNumber, false)
			}
		}

		common.Log.Debug("Returning obj")
		this.ObjCache[objNumber] = obj
		return obj, false, nil
	} else if xref.xtype == XREF_OBJECT_STREAM {
		common.Log.Debug("xref from object stream!")
		common.Log.Debug(">Load via OS!")
		common.Log.Debug("Object stream available in object %d/%d", xref.osObjNumber, xref.osObjIndex)

		if xref.osObjNumber == objNumber {
			common.Log.Debug("ERROR Circular reference!?!")
			return nil, true, errors.New("Xref circular reference")
		}
		_, exists := this.xrefs[xref.osObjNumber]
		if exists {
			optr, err := this.lookupObjectViaOS(xref.osObjNumber, objNumber) //xref.osObjIndex)
			if err != nil {
				common.Log.Debug("ERROR Returning ERR (%s)", err)
				return nil, true, err
			}
			common.Log.Debug("<Loaded via OS")
			this.ObjCache[objNumber] = optr
			if this.crypter != nil {
				// Mark as decrypted (inside object stream) for caching.
				// and avoid decrypting decrypted object.
				this.crypter.decryptedObjects[optr] = true
			}
			return optr, true, nil
		} else {
			common.Log.Debug("?? Belongs to a non-cross referenced object ...!")
			return nil, true, errors.New("OS belongs to a non cross referenced object")
		}
	}
	return nil, false, errors.New("Unknown xref type")
}

// LookupByReference
func (this *PdfParser) LookupByReference(ref PdfObjectReference) (PdfObject, error) {
	common.Log.Debug("Looking up reference %s", ref.String())
	return this.LookupByNumber(int(ref.ObjectNumber))
}

// Trace to direct object.
func (this *PdfParser) Trace(obj PdfObject) (PdfObject, error) {
	ref, isRef := obj.(*PdfObjectReference)
	if !isRef {
		// Direct object already.
		return obj, nil
	}

	bakOffset := this.GetFileOffset()
	defer func() { this.SetFileOffset(bakOffset) }()

	o, err := this.LookupByReference(*ref)
	if err != nil {
		return nil, err
	}

	io, _ := o.(*PdfIndirectObject)
	o = io.PdfObject
	_, isRef = o.(*PdfObjectReference)
	if isRef {
		return io, errors.New("Multi depth trace pointer to pointer")
	}

	return o, nil
}

func printXrefTable(xrefTable XrefTable) {
	common.Log.Debug("=X=X=X=")
	common.Log.Debug("Xref table:")
	i := 0
	for _, xref := range xrefTable {
		common.Log.Debug("i+1: %d (obj num: %d gen: %d) -> %d", i+1, xref.objectNumber, xref.generation, xref.offset)
		i++
	}
}
