/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/unidoc/unidoc/common"
)

// TODO (v3): Create a new type xrefType which can be an integer and can be used for improved type checking.
// TODO (v3): Unexport these constants and rename with camelCase.
const (
	// XREF_TABLE_ENTRY indicates a normal xref table entry.
	XREF_TABLE_ENTRY = iota

	// XREF_OBJECT_STREAM indicates an xref entry in an xref object stream.
	XREF_OBJECT_STREAM = iota
)

// XrefObject defines a cross reference entry which is a map between object number (with generation number) and the
// location of the actual object, either as a file offset (xref table entry), or as a location within an xref
// stream object (xref object stream).
// TODO (v3): Unexport.
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

// XrefTable is a map between object number and corresponding XrefObject.
// TODO (v3): Unexport.
// TODO: Consider changing to a slice, so can maintain the object order without sorting when analyzing.
type XrefTable map[int]XrefObject

// ObjectStream represents an object stream's information which can contain multiple indirect objects.
// The information specifies the number of objects and has information about offset locations for
// each object.
// TODO (v3): Unexport.
type ObjectStream struct {
	N       int // TODO (v3): Unexport.
	ds      []byte
	offsets map[int]int64
}

// ObjectStreams defines a map between object numbers (object streams only) and underlying ObjectStream information.
type ObjectStreams map[int]ObjectStream

// ObjectCache defines a map between object numbers and corresponding PdfObject. Serves as a cache for PdfObjects that
// have already been parsed.
// TODO (v3): Unexport.
type ObjectCache map[int]PdfObject

// Get an object from an object stream.
func (parser *PdfParser) lookupObjectViaOS(sobjNumber int, objNum int) (PdfObject, error) {
	var bufReader *bytes.Reader
	var objstm ObjectStream
	var cached bool

	objstm, cached = parser.objstms[sobjNumber]
	if !cached {
		soi, err := parser.LookupByNumber(sobjNumber)
		if err != nil {
			common.Log.Debug("Missing object stream with number %d", sobjNumber)
			return nil, err
		}

		so, ok := soi.(*PdfObjectStream)
		if !ok {
			return nil, errors.New("Invalid object stream")
		}

		if parser.crypter != nil && !parser.crypter.isDecrypted(so) {
			return nil, errors.New("Need to decrypt the stream")
		}

		sod := so.PdfObjectDictionary
		common.Log.Trace("so d: %s\n", *sod)
		name, ok := sod.Get("Type").(*PdfObjectName)
		if !ok {
			common.Log.Debug("ERROR: Object stream should always have a Type")
			return nil, errors.New("Object stream missing Type")
		}
		if strings.ToLower(string(*name)) != "objstm" {
			common.Log.Debug("ERROR: Object stream type shall always be ObjStm !")
			return nil, errors.New("Object stream type != ObjStm")
		}

		N, ok := sod.Get("N").(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid N in stream dictionary")
		}
		firstOffset, ok := sod.Get("First").(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid First in stream dictionary")
		}

		common.Log.Trace("type: %s number of objects: %d", name, *N)
		ds, err := DecodeStream(so)
		if err != nil {
			return nil, err
		}

		common.Log.Trace("Decoded: %s", ds)

		// Temporarily change the reader object to this decoded buffer.
		// Change back afterwards.
		bakOffset := parser.GetFileOffset()
		defer func() { parser.SetFileOffset(bakOffset) }()

		bufReader = bytes.NewReader(ds)
		parser.reader = bufio.NewReader(bufReader)

		common.Log.Trace("Parsing offset map")
		// Load the offset map (relative to the beginning of the stream...)
		offsets := map[int]int64{}
		// Object list and offsets.
		for i := 0; i < int(*N); i++ {
			parser.skipSpaces()
			// Object number.
			obj, err := parser.parseNumber()
			if err != nil {
				return nil, err
			}
			onum, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, errors.New("Invalid object stream offset table")
			}

			parser.skipSpaces()
			// Offset.
			obj, err = parser.parseNumber()
			if err != nil {
				return nil, err
			}
			offset, ok := obj.(*PdfObjectInteger)
			if !ok {
				return nil, errors.New("Invalid object stream offset table")
			}

			common.Log.Trace("obj %d offset %d", *onum, *offset)
			offsets[int(*onum)] = int64(*firstOffset + *offset)
		}

		objstm = ObjectStream{N: int(*N), ds: ds, offsets: offsets}
		parser.objstms[sobjNumber] = objstm
	} else {
		// Temporarily change the reader object to this decoded buffer.
		// Point back afterwards.
		bakOffset := parser.GetFileOffset()
		defer func() { parser.SetFileOffset(bakOffset) }()

		bufReader = bytes.NewReader(objstm.ds)
		// Temporarily change the reader object to this decoded buffer.
		parser.reader = bufio.NewReader(bufReader)
	}

	offset := objstm.offsets[objNum]
	common.Log.Trace("ACTUAL offset[%d] = %d", objNum, offset)

	bufReader.Seek(offset, os.SEEK_SET)
	parser.reader = bufio.NewReader(bufReader)

	bb, _ := parser.reader.Peek(100)
	common.Log.Trace("OBJ peek \"%s\"", string(bb))

	val, err := parser.parseObject()
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

// LookupByNumber looks up a PdfObject by object number.  Returns an error on failure.
// TODO (v3): Unexport.
func (parser *PdfParser) LookupByNumber(objNumber int) (PdfObject, error) {
	// Outside interface for lookupByNumberWrapper.  Default attempts repairs of bad xref tables.
	obj, _, err := parser.lookupByNumberWrapper(objNumber, true)
	return obj, err
}

// Wrapper for lookupByNumber, checks if object encrypted etc.
func (parser *PdfParser) lookupByNumberWrapper(objNumber int, attemptRepairs bool) (PdfObject, bool, error) {
	obj, inObjStream, err := parser.lookupByNumber(objNumber, attemptRepairs)
	if err != nil {
		return nil, inObjStream, err
	}

	// If encrypted, decrypt it prior to returning.
	// Do not attempt to decrypt objects within object streams.
	if !inObjStream && parser.crypter != nil && !parser.crypter.isDecrypted(obj) {
		err := parser.crypter.Decrypt(obj, 0, 0)
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
func (parser *PdfParser) lookupByNumber(objNumber int, attemptRepairs bool) (PdfObject, bool, error) {
	obj, ok := parser.ObjCache[objNumber]
	if ok {
		common.Log.Trace("Returning cached object %d", objNumber)
		return obj, false, nil
	}

	xref, ok := parser.xrefs[objNumber]
	if !ok {
		// An indirect reference to an undefined object shall not be
		// considered an error by a conforming reader; it shall be
		// treated as a reference to the null object.
		common.Log.Trace("Unable to locate object in xrefs! - Returning null object")
		var nullObj PdfObjectNull
		return &nullObj, false, nil
	}

	common.Log.Trace("Lookup obj number %d", objNumber)
	if xref.xtype == XREF_TABLE_ENTRY {
		common.Log.Trace("xrefobj obj num %d", xref.objectNumber)
		common.Log.Trace("xrefobj gen %d", xref.generation)
		common.Log.Trace("xrefobj offset %d", xref.offset)

		parser.rs.Seek(xref.offset, os.SEEK_SET)
		parser.reader = bufio.NewReader(parser.rs)

		obj, err := parser.ParseIndirectObject()
		if err != nil {
			common.Log.Debug("ERROR Failed reading xref (%s)", err)
			// Offset pointing to a non-object.  Try to repair the file.
			if attemptRepairs {
				common.Log.Debug("Attempting to repair xrefs (top down)")
				xrefTable, err := parser.repairRebuildXrefsTopDown()
				if err != nil {
					common.Log.Debug("ERROR Failed repair (%s)", err)
					return nil, false, err
				}
				parser.xrefs = *xrefTable
				return parser.lookupByNumber(objNumber, false)
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
				err := parser.rebuildXrefTable()
				if err != nil {
					return nil, false, err
				}
				// Empty the cache.
				parser.ObjCache = ObjectCache{}
				// Try looking up again and return.
				return parser.lookupByNumberWrapper(objNumber, false)
			}
		}

		common.Log.Trace("Returning obj")
		parser.ObjCache[objNumber] = obj
		return obj, false, nil
	} else if xref.xtype == XREF_OBJECT_STREAM {
		common.Log.Trace("xref from object stream!")
		common.Log.Trace(">Load via OS!")
		common.Log.Trace("Object stream available in object %d/%d", xref.osObjNumber, xref.osObjIndex)

		if xref.osObjNumber == objNumber {
			common.Log.Debug("ERROR Circular reference!?!")
			return nil, true, errors.New("Xref circular reference")
		}
		_, exists := parser.xrefs[xref.osObjNumber]
		if exists {
			optr, err := parser.lookupObjectViaOS(xref.osObjNumber, objNumber) //xref.osObjIndex)
			if err != nil {
				common.Log.Debug("ERROR Returning ERR (%s)", err)
				return nil, true, err
			}
			common.Log.Trace("<Loaded via OS")
			parser.ObjCache[objNumber] = optr
			if parser.crypter != nil {
				// Mark as decrypted (inside object stream) for caching.
				// and avoid decrypting decrypted object.
				parser.crypter.DecryptedObjects[optr] = true
			}
			return optr, true, nil
		} else {
			common.Log.Debug("?? Belongs to a non-cross referenced object ...!")
			return nil, true, errors.New("OS belongs to a non cross referenced object")
		}
	}
	return nil, false, errors.New("Unknown xref type")
}

// LookupByReference looks up a PdfObject by a reference.
func (parser *PdfParser) LookupByReference(ref PdfObjectReference) (PdfObject, error) {
	common.Log.Trace("Looking up reference %s", ref.String())
	return parser.LookupByNumber(int(ref.ObjectNumber))
}

// Trace traces a PdfObject to direct object, looking up and resolving references as needed (unlike TraceToDirect).
// TODO (v3): Unexport.
func (parser *PdfParser) Trace(obj PdfObject) (PdfObject, error) {
	ref, isRef := obj.(*PdfObjectReference)
	if !isRef {
		// Direct object already.
		return obj, nil
	}

	bakOffset := parser.GetFileOffset()
	defer func() { parser.SetFileOffset(bakOffset) }()

	o, err := parser.LookupByReference(*ref)
	if err != nil {
		return nil, err
	}

	io, isInd := o.(*PdfIndirectObject)
	if !isInd {
		// Not indirect (Stream or null object).
		return o, nil
	}
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
