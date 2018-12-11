/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Default writing implementation.  Basic output with version 1.3
// for compatibility.

package model

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/common/license"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/core/security"
	"github.com/unidoc/unidoc/pdf/core/security/crypt"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

type crossReference struct {
	Type int
	// Type 1
	Offset     int64
	Generation int64 // and Type 0
	// Type 2
	ObjectNumber int // and Type 0
	Index        int
}

var pdfCreator = ""

func getPdfProducer() string {
	licenseKey := license.GetLicenseKey()
	return fmt.Sprintf("UniDoc v%s (%s) - http://unidoc.io", getUniDocVersion(), licenseKey.TypeToString())
}

func getPdfCreator() string {
	if len(pdfCreator) > 0 {
		return pdfCreator
	}

	// Return default.
	return "UniDoc - http://unidoc.io"
}

// SetPdfCreator sets the Creator attribute of the output PDF.
func SetPdfCreator(creator string) {
	pdfCreator = creator
}

// PdfWriter handles outputing PDF content.
type PdfWriter struct {
	root        *PdfIndirectObject
	pages       *PdfIndirectObject
	objects     []PdfObject
	objectsMap  map[PdfObject]bool // Quick lookup table.
	writer      *bufio.Writer
	writePos    int64 // Represents the current position within output file.
	outlines    []*PdfIndirectObject
	outlineTree *PdfOutlineTreeNode
	catalog     *PdfObjectDictionary
	fields      []PdfObject
	infoObj     *PdfIndirectObject

	// Encryption
	crypter     *PdfCrypt
	encryptDict *PdfObjectDictionary
	encryptObj  *PdfIndirectObject
	ids         *PdfObjectArray

	// PDF version
	majorVersion int
	minorVersion int

	// Objects to be followed up on prior to writing.
	// These are objects that are added and reference objects that are not included
	// for writing.
	// The map stores the object and the dictionary it is contained in.
	// Only way so we can access the dictionary entry later.
	pendingObjects map[PdfObject]*PdfObjectDictionary

	// Forms.
	acroForm *PdfAcroForm

	optimizer         Optimizer
	crossReferenceMap map[int]crossReference
	writeOffset       int64 // used by PdfAppender
	ObjNumOffset      int
	appendMode        bool
	appendToXrefs     XrefTable
}

// NewPdfWriter initializes a new PdfWriter.
func NewPdfWriter() PdfWriter {
	w := PdfWriter{}

	w.objectsMap = map[PdfObject]bool{}
	w.objects = []PdfObject{}
	w.pendingObjects = map[PdfObject]*PdfObjectDictionary{}

	// PDF Version.  Can be changed if using more advanced features in PDF.
	// By default it is set to 1.3.
	w.majorVersion = 1
	w.minorVersion = 3

	// Creation info.
	infoDict := MakeDict()
	infoDict.Set("Producer", MakeString(getPdfProducer()))
	infoDict.Set("Creator", MakeString(getPdfCreator()))
	infoObj := PdfIndirectObject{}
	infoObj.PdfObject = infoDict
	w.infoObj = &infoObj
	w.addObject(&infoObj)

	// Root catalog.
	catalog := PdfIndirectObject{}
	catalogDict := MakeDict()
	catalogDict.Set("Type", MakeName("Catalog"))
	catalog.PdfObject = catalogDict

	w.root = &catalog
	w.addObject(w.root)

	// Pages.
	pages := PdfIndirectObject{}
	pagedict := MakeDict()
	pagedict.Set("Type", MakeName("Pages"))
	kids := PdfObjectArray{}
	pagedict.Set("Kids", &kids)
	pagedict.Set("Count", MakeInteger(0))
	pages.PdfObject = pagedict

	w.pages = &pages
	w.addObject(w.pages)

	catalogDict.Set("Pages", &pages)
	w.catalog = catalogDict

	common.Log.Trace("Catalog %s", catalog)

	return w
}

// copyObject creates deep copy of the Pdf object and
// fills objectToObjectCopyMap to replace the old object to the copy of object if needed.
// Parameter objectToObjectCopyMap is needed to replace object references to its copies.
// Because many objects can contain references to another objects like pages to images.
func copyObject(obj PdfObject, objectToObjectCopyMap map[PdfObject]PdfObject) PdfObject {
	if newObj, ok := objectToObjectCopyMap[obj]; ok {
		return newObj
	}

	switch t := obj.(type) {
	case *PdfObjectArray:
		newObj := &PdfObjectArray{}
		objectToObjectCopyMap[obj] = newObj
		for _, val := range t.Elements() {
			newObj.Append(copyObject(val, objectToObjectCopyMap))
		}
		return newObj
	case *PdfObjectStreams:
		newObj := &PdfObjectStreams{PdfObjectReference: t.PdfObjectReference}
		objectToObjectCopyMap[obj] = newObj
		for _, val := range t.Elements() {
			newObj.Append(copyObject(val, objectToObjectCopyMap))
		}
		return newObj
	case *PdfObjectStream:
		newObj := &PdfObjectStream{
			Stream:             t.Stream,
			PdfObjectReference: t.PdfObjectReference,
		}
		objectToObjectCopyMap[obj] = newObj
		newObj.PdfObjectDictionary = copyObject(t.PdfObjectDictionary, objectToObjectCopyMap).(*PdfObjectDictionary)
		return newObj
	case *PdfObjectDictionary:
		newObj := MakeDict()
		objectToObjectCopyMap[obj] = newObj
		for _, key := range t.Keys() {
			val := t.Get(key)
			newObj.Set(key, copyObject(val, objectToObjectCopyMap))
		}
		return newObj
	case *PdfIndirectObject:
		newObj := &PdfIndirectObject{
			PdfObjectReference: t.PdfObjectReference,
		}
		objectToObjectCopyMap[obj] = newObj
		newObj.PdfObject = copyObject(t.PdfObject, objectToObjectCopyMap)
		return newObj
	case *PdfObjectString:
		newObj := &PdfObjectString{}
		*newObj = *t
		objectToObjectCopyMap[obj] = newObj
		return newObj
	case *PdfObjectName:
		newObj := PdfObjectName(*t)
		objectToObjectCopyMap[obj] = &newObj
		return &newObj
	case *PdfObjectNull:
		newObj := PdfObjectNull{}
		objectToObjectCopyMap[obj] = &newObj
		return &newObj
	case *PdfObjectInteger:
		newObj := PdfObjectInteger(*t)
		objectToObjectCopyMap[obj] = &newObj
		return &newObj
	case *PdfObjectReference:
		newObj := PdfObjectReference(*t)
		objectToObjectCopyMap[obj] = &newObj
		return &newObj
	case *PdfObjectFloat:
		newObj := PdfObjectFloat(*t)
		objectToObjectCopyMap[obj] = &newObj
		return &newObj
	case *PdfObjectBool:
		newObj := PdfObjectBool(*t)
		objectToObjectCopyMap[obj] = &newObj
		return &newObj
	default:
		common.Log.Info("TODO(a5i): implement copyObject for %+v", obj)
	}
	// return other objects as is
	return obj
}

// copyObjects makes objects copy and set as working.
func (this *PdfWriter) copyObjects() {
	objectToObjectCopyMap := make(map[PdfObject]PdfObject)
	objects := make([]PdfObject, len(this.objects))
	objectsMap := make(map[PdfObject]bool)
	for i, obj := range this.objects {
		newObject := copyObject(obj, objectToObjectCopyMap)
		objects[i] = newObject
		if this.objectsMap[obj] {
			objectsMap[newObject] = true
		}
	}

	this.objects = objects
	this.objectsMap = objectsMap
	this.infoObj = copyObject(this.infoObj, objectToObjectCopyMap).(*PdfIndirectObject)
	this.root = copyObject(this.root, objectToObjectCopyMap).(*PdfIndirectObject)
	if this.encryptObj != nil {
		this.encryptObj = copyObject(this.encryptObj, objectToObjectCopyMap).(*PdfIndirectObject)
	}
}

// Set the PDF version of the output file.
func (this *PdfWriter) SetVersion(majorVersion, minorVersion int) {
	this.majorVersion = majorVersion
	this.minorVersion = minorVersion
}

// SetOCProperties sets the optional content properties.
func (this *PdfWriter) SetOCProperties(ocProperties PdfObject) error {
	dict := this.catalog

	if ocProperties != nil {
		common.Log.Trace("Setting OC Properties...")
		dict.Set("OCProperties", ocProperties)
		// Any risk of infinite loops?
		this.addObjects(ocProperties)
	}

	return nil
}

// SetOptimizer sets the optimizer to optimize PDF before writing.
func (this *PdfWriter) SetOptimizer(optimizer Optimizer) {
	this.optimizer = optimizer
}

// GetOptimizer returns current PDF optimizer.
func (this *PdfWriter) GetOptimizer() Optimizer {
	return this.optimizer
}

func (this *PdfWriter) hasObject(obj PdfObject) bool {
	// Check if already added.
	for _, o := range this.objects {
		// GH: May perform better to use a hash map to check if added?
		if o == obj {
			return true
		}
	}
	return false
}

// Adds the object to list of objects and returns true if the obj was
// not already added.
// Returns false if the object was previously added.
func (this *PdfWriter) addObject(obj PdfObject) bool {
	hasObj := this.hasObject(obj)
	if !hasObj {
		this.objects = append(this.objects, obj)
		return true
	}

	return false
}

func (this *PdfWriter) addObjects(obj PdfObject) error {
	common.Log.Trace("Adding objects!")

	if io, isIndirectObj := obj.(*PdfIndirectObject); isIndirectObj {
		common.Log.Trace("Indirect")
		common.Log.Trace("- %s (%p)", obj, io)
		common.Log.Trace("- %s", io.PdfObject)
		if this.addObject(io) {
			err := this.addObjects(io.PdfObject)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if so, isStreamObj := obj.(*PdfObjectStream); isStreamObj {
		common.Log.Trace("Stream")
		common.Log.Trace("- %s %p", obj, obj)
		if this.addObject(so) {
			err := this.addObjects(so.PdfObjectDictionary)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if dict, isDict := obj.(*PdfObjectDictionary); isDict {
		common.Log.Trace("Dict")
		common.Log.Trace("- %s", obj)
		for _, k := range dict.Keys() {
			v := dict.Get(k)
			common.Log.Trace("Key %s", k)
			if k != "Parent" {
				err := this.addObjects(v)
				if err != nil {
					return err
				}
			} else {
				if _, parentIsNull := dict.Get("Parent").(*PdfObjectNull); parentIsNull {
					// Parent is null.  We can ignore it.
					continue
				}

				if hasObj := this.hasObject(v); !hasObj {
					common.Log.Debug("Parent obj is missing!! %T %p %v", v, v, v)
					this.pendingObjects[v] = dict
					// Although it is missing at this point, it could be added later...
				}
				// How to handle the parent?  Make sure it is present?
				if parentObj, parentIsRef := dict.Get("Parent").(*PdfObjectReference); parentIsRef {
					// Parent is a reference.  Means we can drop it?
					// Could refer to somewhere outside of the scope of the output doc.
					// Should be done by the reader already.
					// -> ERROR.
					common.Log.Debug("ERROR: Parent is a reference object - Cannot be in writer (needs to be resolved)")
					return fmt.Errorf("Parent is a reference object - Cannot be in writer (needs to be resolved) - %s", parentObj)
				}
			}
		}
		return nil
	}

	if arr, isArray := obj.(*PdfObjectArray); isArray {
		common.Log.Trace("Array")
		common.Log.Trace("- %s", obj)
		if arr == nil {
			return errors.New("Array is nil")
		}
		for _, v := range arr.Elements() {
			err := this.addObjects(v)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if _, isReference := obj.(*PdfObjectReference); isReference {
		// Should never be a reference, should already be resolved.
		common.Log.Debug("ERROR: Cannot be a reference!")
		return errors.New("Reference not allowed")
	}

	return nil
}

// AddPage adds a page to the PDF file. The new page should be an indirect object.
func (this *PdfWriter) AddPage(page *PdfPage) error {
	obj := page.ToPdfObject()
	common.Log.Trace("==========")
	common.Log.Trace("Appending to page list %T", obj)
	procPage(page)

	pageObj, ok := obj.(*PdfIndirectObject)
	if !ok {
		return errors.New("Page should be an indirect object")
	}
	common.Log.Trace("%s", pageObj)
	common.Log.Trace("%s", pageObj.PdfObject)

	pDict, ok := pageObj.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return errors.New("Page object should be a dictionary")
	}

	otype, ok := pDict.Get("Type").(*PdfObjectName)
	if !ok {
		return fmt.Errorf("Page should have a Type key with a value of type name (%T)", pDict.Get("Type"))

	}
	if *otype != "Page" {
		return errors.New("Type != Page (Required).")
	}

	// Copy inherited fields if missing.
	inheritedFields := []PdfObjectName{"Resources", "MediaBox", "CropBox", "Rotate"}
	parent, hasParent := pDict.Get("Parent").(*PdfIndirectObject)
	common.Log.Trace("Page Parent: %T (%v)", pDict.Get("Parent"), hasParent)
	for hasParent {
		common.Log.Trace("Page Parent: %T", parent)
		parentDict, ok := parent.PdfObject.(*PdfObjectDictionary)
		if !ok {
			return errors.New("Invalid Parent object")
		}
		for _, field := range inheritedFields {
			common.Log.Trace("Field %s", field)
			if pDict.Get(field) != nil {
				common.Log.Trace("- page has already")
				continue
			}

			if obj := parentDict.Get(field); obj != nil {
				// Parent has the field.  Inherit, pass to the new page.
				common.Log.Trace("Inheriting field %s", field)
				pDict.Set(field, obj)
			}
		}
		parent, hasParent = parentDict.Get("Parent").(*PdfIndirectObject)
		common.Log.Trace("Next parent: %T", parentDict.Get("Parent"))
	}

	common.Log.Trace("Traversal done")

	// Update the dictionary.
	// Reuses the input object, updating the fields.
	pDict.Set("Parent", this.pages)
	pageObj.PdfObject = pDict

	// Add to Pages.
	pagesDict, ok := this.pages.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return errors.New("Invalid Pages obj (not a dict)")
	}
	kids, ok := pagesDict.Get("Kids").(*PdfObjectArray)
	if !ok {
		return errors.New("Invalid Pages Kids obj (not an array)")
	}
	kids.Append(pageObj)
	pageCount, ok := pagesDict.Get("Count").(*PdfObjectInteger)
	if !ok {
		return errors.New("Invalid Pages Count object (not an integer)")
	}
	// Update the count.
	*pageCount = *pageCount + 1

	this.addObject(pageObj)

	// Traverse the page and record all object references.
	err := this.addObjects(pDict)
	if err != nil {
		return err
	}

	return nil
}

func procPage(p *PdfPage) {
	lk := license.GetLicenseKey()
	if lk != nil && lk.IsLicensed() {
		return
	}

	// Add font as needed.
	f := fonts.NewFontHelvetica()
	p.Resources.SetFontByName("UF1", f.ToPdfObject())

	ops := []string{}
	ops = append(ops, "q")
	ops = append(ops, "BT")
	ops = append(ops, "/UF1 14 Tf")
	ops = append(ops, "1 0 0 rg")
	ops = append(ops, "10 10 Td")
	s := "Unlicensed UniDoc - Get a license on https://unidoc.io"
	ops = append(ops, fmt.Sprintf("(%s) Tj", s))
	ops = append(ops, "ET")
	ops = append(ops, "Q")
	contentstr := strings.Join(ops, "\n")

	p.AddContentStreamByString(contentstr)

	// Update page object.
	p.ToPdfObject()
}

// AddOutlineTree adds outlines to a PDF file.
func (this *PdfWriter) AddOutlineTree(outlineTree *PdfOutlineTreeNode) {
	this.outlineTree = outlineTree
}

// Look for a specific key.  Returns a list of entries.
// What if something appears on many pages?
func (this *PdfWriter) seekByName(obj PdfObject, followKeys []string, key string) ([]PdfObject, error) {
	common.Log.Trace("Seek by name.. %T", obj)
	list := []PdfObject{}
	if io, isIndirectObj := obj.(*PdfIndirectObject); isIndirectObj {
		return this.seekByName(io.PdfObject, followKeys, key)
	}

	if so, isStreamObj := obj.(*PdfObjectStream); isStreamObj {
		return this.seekByName(so.PdfObjectDictionary, followKeys, key)
	}

	if dict, isDict := obj.(*PdfObjectDictionary); isDict {
		common.Log.Trace("Dict")
		for _, k := range dict.Keys() {
			v := dict.Get(k)
			if string(k) == key {
				list = append(list, v)
			}
			for _, followKey := range followKeys {
				if string(k) == followKey {
					common.Log.Trace("Follow key %s", followKey)
					items, err := this.seekByName(v, followKeys, key)
					if err != nil {
						return list, err
					}
					for _, item := range items {
						list = append(list, item)
					}
					break
				}
			}
		}
		return list, nil
	}

	return list, nil
}

// SetForms sets the Acroform for a PDF file.
func (this *PdfWriter) SetForms(form *PdfAcroForm) error {
	this.acroForm = form
	return nil
}

// writeObject writes out an indirect / stream object.
func (this *PdfWriter) writeObject(num int, obj PdfObject) {
	common.Log.Trace("Write obj #%d\n", num)

	if pobj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		this.crossReferenceMap[num] = crossReference{Type: 1, Offset: this.writePos, Generation: pobj.GenerationNumber}
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		outStr += pobj.PdfObject.DefaultWriteString()
		outStr += "\nendobj\n"
		this.writeString(outStr)
		return
	}

	// XXX/TODO: Add a default encoder if Filter not specified?
	// Still need to make sure is encrypted.
	if pobj, isStream := obj.(*PdfObjectStream); isStream {
		this.crossReferenceMap[num] = crossReference{Type: 1, Offset: this.writePos, Generation: pobj.GenerationNumber}
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		outStr += pobj.PdfObjectDictionary.DefaultWriteString()
		outStr += "\nstream\n"
		this.writeString(outStr)
		this.writeBytes(pobj.Stream)
		this.writeString("\nendstream\nendobj\n")
		return
	}

	if ostreams, isObjStreams := obj.(*PdfObjectStreams); isObjStreams {
		this.crossReferenceMap[num] = crossReference{Type: 1, Offset: this.writePos, Generation: ostreams.GenerationNumber}
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		var offsets []string
		var objData string
		var offset int64

		for index, obj := range ostreams.Elements() {
			io, isIndirect := obj.(*PdfIndirectObject)
			if !isIndirect {
				common.Log.Error("Object streams N %d contains non indirect pdf object %v", num, obj)
			}
			data := io.PdfObject.DefaultWriteString() + " "
			objData = objData + data
			offsets = append(offsets, fmt.Sprintf("%d %d", io.ObjectNumber, offset))
			this.crossReferenceMap[int(io.ObjectNumber)] = crossReference{Type: 2, ObjectNumber: num, Index: index}
			offset = offset + int64(len([]byte(data)))
		}
		offsetsStr := strings.Join(offsets, " ") + " "
		encoder := NewFlateEncoder()
		//encoder := NewRawEncoder()
		dict := encoder.MakeStreamDict()
		dict.Set(PdfObjectName("Type"), MakeName("ObjStm"))
		n := int64(ostreams.Len())
		dict.Set(PdfObjectName("N"), MakeInteger(n))
		first := int64(len(offsetsStr))
		dict.Set(PdfObjectName("First"), MakeInteger(first))

		data, _ := encoder.EncodeBytes([]byte(offsetsStr + objData))
		length := int64(len(data))

		dict.Set(PdfObjectName("Length"), MakeInteger(length))
		outStr += dict.DefaultWriteString()
		outStr += "\nstream\n"
		this.writeString(outStr)
		this.writeBytes(data)
		this.writeString("\nendstream\nendobj\n")
		return
	}

	this.writer.WriteString(obj.DefaultWriteString())
}

// Update all the object numbers prior to writing.
func (this *PdfWriter) updateObjectNumbers() {
	offset := this.ObjNumOffset
	// Update numbers
	for idx, obj := range this.objects {
		switch o := obj.(type) {
		case *PdfIndirectObject:
			o.ObjectNumber = int64(idx + 1 + offset)
			o.GenerationNumber = 0
		case *PdfObjectStream:
			o.ObjectNumber = int64(idx + 1 + offset)
			o.GenerationNumber = 0
		case *PdfObjectStreams:
			o.ObjectNumber = int64(idx + 1 + offset)
			o.GenerationNumber = 0
		}
	}
}

// EncryptOptions represents encryption options for an output PDF.
type EncryptOptions struct {
	Permissions security.Permissions
	Algorithm   EncryptionAlgorithm
}

// EncryptionAlgorithm is used in EncryptOptions to change the default algorithm used to encrypt the document.
type EncryptionAlgorithm int

const (
	// RC4_128bit uses RC4 encryption (128 bit)
	RC4_128bit = EncryptionAlgorithm(iota)
	// AES_128bit uses AES encryption (128 bit, PDF 1.6)
	AES_128bit
	// AES_256bit uses AES encryption (256 bit, PDF 2.0)
	AES_256bit
)

// Encrypt encrypts the output file with a specified user/owner password.
func (this *PdfWriter) Encrypt(userPass, ownerPass []byte, options *EncryptOptions) error {
	algo := RC4_128bit
	if options != nil {
		algo = options.Algorithm
	}
	perm := security.PermOwner
	if options != nil {
		perm = options.Permissions
	}

	var cf crypt.Filter
	switch algo {
	case RC4_128bit:
		cf = crypt.NewFilterV2(16)
	case AES_128bit:
		cf = crypt.NewFilterAESV2()
	case AES_256bit:
		cf = crypt.NewFilterAESV3()
	default:
		return fmt.Errorf("unsupported algorithm: %v", options.Algorithm)
	}
	crypter, info, err := PdfCryptNewEncrypt(cf, userPass, ownerPass, perm)
	if err != nil {
		return err
	}
	this.crypter = crypter
	if info.Major != 0 {
		this.SetVersion(info.Major, info.Minor)
	}
	this.encryptDict = info.Encrypt

	this.ids = MakeArray(MakeHexString(info.ID0), MakeHexString(info.ID1))

	// Make an object to contain the encryption dictionary.
	io := MakeIndirectObject(info.Encrypt)
	this.encryptObj = io
	this.addObject(io)

	return nil
}

// Wrapper function to handle writing out string.
func (this *PdfWriter) writeString(s string) error {
	n, err := this.writer.WriteString(s)
	if err != nil {
		return err
	}
	this.writePos += int64(n)
	return nil
}

// Wrapper function to handle writing out bytes.
func (this *PdfWriter) writeBytes(bb []byte) error {
	n, err := this.writer.Write(bb)
	if err != nil {
		return err
	}
	this.writePos += int64(n)
	return nil
}

// Write writes out the PDF.
func (this *PdfWriter) Write(writer io.Writer) error {
	common.Log.Trace("Write()")

	lk := license.GetLicenseKey()
	if lk == nil || !lk.IsLicensed() {
		fmt.Printf("Unlicensed copy of unidoc\n")
		fmt.Printf("To get rid of the watermark - Please get a license on https://unidoc.io\n")
	}

	// Outlines.
	if this.outlineTree != nil {
		common.Log.Trace("OutlineTree: %+v", this.outlineTree)
		outlines := this.outlineTree.ToPdfObject()
		common.Log.Trace("Outlines: %+v (%T, p:%p)", outlines, outlines, outlines)
		this.catalog.Set("Outlines", outlines)
		err := this.addObjects(outlines)
		if err != nil {
			return err
		}
	}

	// Form fields.
	if this.acroForm != nil {
		common.Log.Trace("Writing acro forms")
		indObj := this.acroForm.ToPdfObject()
		common.Log.Trace("AcroForm: %+v", indObj)
		this.catalog.Set("AcroForm", indObj)
		err := this.addObjects(indObj)
		if err != nil {
			return err
		}
	}

	// Check pending objects prior to write.
	for pendingObj, pendingObjDict := range this.pendingObjects {
		if !this.hasObject(pendingObj) {
			common.Log.Debug("ERROR Pending object %+v %T (%p) never added for writing", pendingObj, pendingObj, pendingObj)
			for _, key := range pendingObjDict.Keys() {
				val := pendingObjDict.Get(key)
				if val == pendingObj {
					common.Log.Debug("Pending object found! and replaced with null")
					pendingObjDict.Set(key, MakeNull())
					break
				}
			}
		}
	}
	// Set version in the catalog.
	this.catalog.Set("Version", MakeName(fmt.Sprintf("%d.%d", this.majorVersion, this.minorVersion)))

	// Make a copy of objects prior to optimizing as this can alter the objects.
	this.copyObjects()

	if this.optimizer != nil {
		var err error
		this.objects, err = this.optimizer.Optimize(this.objects)
		if err != nil {
			return err
		}
	}

	this.writePos = this.writeOffset

	w := bufio.NewWriter(writer)
	this.writer = w

	useCrossReferenceStream := this.majorVersion > 1 || (this.majorVersion == 1 && this.minorVersion > 4)
	objectsInObjectStreams := make(map[PdfObject]bool)
	if !useCrossReferenceStream {
		for _, obj := range this.objects {
			if objStm, isObjectStreams := obj.(*PdfObjectStreams); isObjectStreams {
				useCrossReferenceStream = true
				for _, obj := range objStm.Elements() {
					objectsInObjectStreams[obj] = true
					if io, isIndirectObj := obj.(*PdfIndirectObject); isIndirectObj {
						objectsInObjectStreams[io.PdfObject] = true
					}
				}
			}
		}
	}

	if useCrossReferenceStream && this.majorVersion == 1 && this.minorVersion < 5 {
		this.minorVersion = 5
	}

	if this.appendMode {
		this.writeString("\n")
	} else {
		this.writeString(fmt.Sprintf("%%PDF-%d.%d\n", this.majorVersion, this.minorVersion))
		this.writeString("%âãÏÓ\n")
	}

	this.updateObjectNumbers()

	// Write objects
	common.Log.Trace("Writing %d obj", len(this.objects))
	this.crossReferenceMap = make(map[int]crossReference)
	this.crossReferenceMap[0] = crossReference{Type: 0, ObjectNumber: 0, Generation: 0xFFFF}
	if this.appendToXrefs != nil {
		for idx, xref := range this.appendToXrefs {
			if idx == 0 {
				continue
			}
			if xref.XType == XrefTypeObjectStream {
				cr := crossReference{Type: 2, ObjectNumber: xref.OsObjNumber, Index: xref.OsObjIndex}
				this.crossReferenceMap[idx] = cr
			}
			if xref.XType == XrefTypeTableEntry {
				cr := crossReference{Type: 1, ObjectNumber: xref.ObjectNumber, Offset: xref.Offset}
				this.crossReferenceMap[idx] = cr
			}
		}
	}

	offset := this.ObjNumOffset
	for idx, obj := range this.objects {
		if skip := objectsInObjectStreams[obj]; skip {
			continue
		}
		common.Log.Trace("Writing %d", idx)

		objectNumber := int64(idx + 1 + offset)
		// Encrypt prior to writing.
		// Encrypt dictionary should not be encrypted.
		if this.crypter != nil && obj != this.encryptObj {
			err := this.crypter.Encrypt(obj, int64(objectNumber), 0)
			if err != nil {
				common.Log.Debug("ERROR: Failed encrypting (%s)", err)
				return err
			}
		}
		this.writeObject(int(objectNumber), obj)
	}

	xrefOffset := this.writePos
	var maxIndex int
	for idx := range this.crossReferenceMap {
		if idx > maxIndex {
			maxIndex = idx
		}
	}
	if useCrossReferenceStream {

		crossObjNumber := maxIndex + 1
		this.crossReferenceMap[crossObjNumber] = crossReference{Type: 1, ObjectNumber: crossObjNumber, Offset: xrefOffset}
		crossReferenceData := bytes.NewBuffer(nil)

		for idx := 0; idx <= maxIndex+1; idx++ {
			ref := this.crossReferenceMap[idx]
			switch ref.Type {
			case 0:
				binary.Write(crossReferenceData, binary.BigEndian, byte(0))
				binary.Write(crossReferenceData, binary.BigEndian, uint32(0))
				binary.Write(crossReferenceData, binary.BigEndian, uint16(0xFFFF))
			case 1:
				binary.Write(crossReferenceData, binary.BigEndian, byte(1))
				binary.Write(crossReferenceData, binary.BigEndian, uint32(ref.Offset))
				binary.Write(crossReferenceData, binary.BigEndian, uint16(ref.Generation))
			case 2:
				binary.Write(crossReferenceData, binary.BigEndian, byte(2))
				binary.Write(crossReferenceData, binary.BigEndian, uint32(ref.ObjectNumber))
				binary.Write(crossReferenceData, binary.BigEndian, uint16(ref.Index))
			}
		}
		crossReferenceStream, err := MakeStream(crossReferenceData.Bytes(), NewFlateEncoder())
		if err != nil {
			return err
		}
		crossReferenceStream.ObjectNumber = int64(crossObjNumber)
		crossReferenceStream.PdfObjectDictionary.Set("Type", MakeName("XRef"))
		crossReferenceStream.PdfObjectDictionary.Set("W", MakeArray(MakeInteger(1), MakeInteger(4), MakeInteger(2)))
		crossReferenceStream.PdfObjectDictionary.Set("Index", MakeArray(MakeInteger(0), MakeInteger(crossReferenceStream.ObjectNumber+1)))
		crossReferenceStream.PdfObjectDictionary.Set("Size", MakeInteger(crossReferenceStream.ObjectNumber+1))
		crossReferenceStream.PdfObjectDictionary.Set("Info", this.infoObj)
		crossReferenceStream.PdfObjectDictionary.Set("Root", this.root)
		// If encrypted!
		if this.crypter != nil {
			crossReferenceStream.Set("Encrypt", this.encryptObj)
			crossReferenceStream.Set("ID", this.ids)
			common.Log.Trace("Ids: %s", this.ids)
		}

		this.writeObject(int(crossReferenceStream.ObjectNumber), crossReferenceStream)

	} else {
		this.writeString("xref\r\n")
		outStr := fmt.Sprintf("%d %d\r\n", 0, len(this.crossReferenceMap))
		this.writeString(outStr)
		for idx := 0; idx <= maxIndex; idx++ {
			ref := this.crossReferenceMap[idx]
			switch ref.Type {
			case 0:
				outStr = fmt.Sprintf("%.10d %.5d f\r\n", 0, 65535)
				this.writeString(outStr)
			case 1:
				outStr = fmt.Sprintf("%.10d %.5d n\r\n", ref.Offset, 0)
				this.writeString(outStr)
			}
		}

		// Generate & write trailer
		trailer := MakeDict()
		trailer.Set("Info", this.infoObj)
		trailer.Set("Root", this.root)
		trailer.Set("Size", MakeInteger(int64(len(this.crossReferenceMap))))
		// If encrypted!
		if this.crypter != nil {
			trailer.Set("Encrypt", this.encryptObj)
			trailer.Set("ID", this.ids)
			common.Log.Trace("Ids: %s", this.ids)
		}
		this.writeString("trailer\n")
		this.writeString(trailer.DefaultWriteString())
		this.writeString("\n")

	}

	// Make offset reference.
	outStr := fmt.Sprintf("startxref\n%d\n", xrefOffset)
	this.writeString(outStr)
	this.writeString("%%EOF\n")

	this.writer.Flush()

	return nil
}
