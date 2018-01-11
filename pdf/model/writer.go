/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Default writing implementation.  Basic output with version 1.3
// for compatibility.

package model

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/common/license"
	. "github.com/unidoc/unidoc/pdf/core"
)

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

func SetPdfCreator(creator string) {
	pdfCreator = creator
}

type PdfWriter struct {
	root        *PdfIndirectObject
	pages       *PdfIndirectObject
	objects     []PdfObject
	objectsMap  map[PdfObject]bool // Quick lookup table.
	writer      *bufio.Writer
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
}

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
	w.addObject(&catalog)

	// Pages.
	pages := PdfIndirectObject{}
	pagedict := MakeDict()
	pagedict.Set("Type", MakeName("Pages"))
	kids := PdfObjectArray{}
	pagedict.Set("Kids", &kids)
	pagedict.Set("Count", MakeInteger(0))
	pages.PdfObject = pagedict

	w.pages = &pages
	w.addObject(&pages)

	catalogDict.Set("Pages", &pages)
	w.catalog = catalogDict

	common.Log.Trace("Catalog %s", catalog)

	return w
}

// Set the PDF version of the output file.
func (this *PdfWriter) SetVersion(majorVersion, minorVersion int) {
	this.majorVersion = majorVersion
	this.minorVersion = minorVersion
}

// Set the optional content properties.
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
		for _, v := range *arr {
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

// Add a page to the PDF file. The new page should be an indirect
// object.
func (this *PdfWriter) AddPage(page *PdfPage) error {
	obj := page.ToPdfObject()
	common.Log.Trace("==========")
	common.Log.Trace("Appending to page list %T", obj)

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
	*kids = append(*kids, pageObj)
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

// Add outlines to a PDF file.
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

// Add Acroforms to a PDF file.  Sets the specified form for writing.
func (this *PdfWriter) SetForms(form *PdfAcroForm) error {
	this.acroForm = form
	return nil
}

// Write out an indirect / stream object.
func (this *PdfWriter) writeObject(num int, obj PdfObject) {
	common.Log.Trace("Write obj #%d\n", num)

	if pobj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		outStr += pobj.PdfObject.DefaultWriteString()
		outStr += "\nendobj\n"
		this.writer.WriteString(outStr)
		return
	}

	// XXX/TODO: Add a default encoder if Filter not specified?
	// Still need to make sure is encrypted.
	if pobj, isStream := obj.(*PdfObjectStream); isStream {
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		outStr += pobj.PdfObjectDictionary.DefaultWriteString()
		outStr += "\nstream\n"
		this.writer.WriteString(outStr)
		this.writer.Write(pobj.Stream)
		this.writer.WriteString("\nendstream\nendobj\n")
		return
	}

	this.writer.WriteString(obj.DefaultWriteString())
}

// Update all the object numbers prior to writing.
func (this *PdfWriter) updateObjectNumbers() {
	// Update numbers
	for idx, obj := range this.objects {
		if io, isIndirect := obj.(*PdfIndirectObject); isIndirect {
			io.ObjectNumber = int64(idx + 1)
			io.GenerationNumber = 0
		}
		if so, isStream := obj.(*PdfObjectStream); isStream {
			so.ObjectNumber = int64(idx + 1)
			so.GenerationNumber = 0
		}
	}
}

type EncryptOptions struct {
	Permissions AccessPermissions
}

// Encrypt the output file with a specified user/owner password.
func (this *PdfWriter) Encrypt(userPass, ownerPass []byte, options *EncryptOptions) error {
	crypter := PdfCrypt{}
	this.crypter = &crypter

	crypter.EncryptedObjects = map[PdfObject]bool{}

	crypter.CryptFilters = CryptFilters{}
	crypter.CryptFilters["Default"] = CryptFilter{Cfm: "V2", Length: 128}

	// Set
	crypter.P = -1
	crypter.V = 2
	crypter.R = 3
	crypter.Length = 128
	crypter.EncryptMetadata = true
	if options != nil {
		crypter.P = int(options.Permissions.GetP())
	}

	// Prepare the ID object for the trailer.
	hashcode := md5.Sum([]byte(time.Now().Format(time.RFC850)))
	id0 := PdfObjectString(hashcode[:])
	b := make([]byte, 100)
	rand.Read(b)
	hashcode = md5.Sum(b)
	id1 := PdfObjectString(hashcode[:])
	common.Log.Trace("Random b: % x", b)

	this.ids = &PdfObjectArray{&id0, &id1}
	common.Log.Trace("Gen Id 0: % x", id0)

	crypter.Id0 = string(id0)

	// Make the O and U objects.
	O, err := crypter.Alg3(userPass, ownerPass)
	if err != nil {
		common.Log.Debug("ERROR: Error generating O for encryption (%s)", err)
		return err
	}
	crypter.O = []byte(O)
	common.Log.Trace("gen O: % x", O)
	U, key, err := crypter.Alg5(userPass)
	if err != nil {
		common.Log.Debug("ERROR: Error generating O for encryption (%s)", err)
		return err
	}
	common.Log.Trace("gen U: % x", U)
	crypter.U = []byte(U)
	crypter.EncryptionKey = key

	// Generate the encryption dictionary.
	encDict := MakeDict()
	encDict.Set("Filter", MakeName("Standard"))
	encDict.Set("P", MakeInteger(int64(crypter.P)))
	encDict.Set("V", MakeInteger(int64(crypter.V)))
	encDict.Set("R", MakeInteger(int64(crypter.R)))
	encDict.Set("Length", MakeInteger(int64(crypter.Length)))
	encDict.Set("O", &O)
	encDict.Set("U", &U)
	this.encryptDict = encDict

	// Make an object to contain it.
	io := MakeIndirectObject(encDict)
	this.encryptObj = io
	this.addObject(io)

	return nil
}

// Write the pdf out.
func (this *PdfWriter) Write(ws io.WriteSeeker) error {
	common.Log.Trace("Write()")
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

	w := bufio.NewWriter(ws)
	this.writer = w

	w.WriteString(fmt.Sprintf("%%PDF-%d.%d\n", this.majorVersion, this.minorVersion))
	w.WriteString("%âãÏÓ\n")
	w.Flush()

	this.updateObjectNumbers()

	offsets := []int64{}

	// Write objects
	common.Log.Trace("Writing %d obj", len(this.objects))
	for idx, obj := range this.objects {
		common.Log.Trace("Writing %d", idx)
		this.writer.Flush()
		offset, _ := ws.Seek(0, os.SEEK_CUR)
		offsets = append(offsets, offset)

		// Encrypt prior to writing.
		// Encrypt dictionary should not be encrypted.
		if this.crypter != nil && obj != this.encryptObj {
			err := this.crypter.Encrypt(obj, int64(idx+1), 0)
			if err != nil {
				common.Log.Debug("ERROR: Failed encrypting (%s)", err)
				return err
			}

		}
		this.writeObject(idx+1, obj)
	}
	w.Flush()

	xrefOffset, _ := ws.Seek(0, os.SEEK_CUR)
	// Write xref table.
	this.writer.WriteString("xref\r\n")
	outStr := fmt.Sprintf("%d %d\r\n", 0, len(this.objects)+1)
	this.writer.WriteString(outStr)
	outStr = fmt.Sprintf("%.10d %.5d f\r\n", 0, 65535)
	this.writer.WriteString(outStr)
	for _, offset := range offsets {
		outStr = fmt.Sprintf("%.10d %.5d n\r\n", offset, 0)
		this.writer.WriteString(outStr)
	}

	// Generate & write trailer
	trailer := MakeDict()
	trailer.Set("Info", this.infoObj)
	trailer.Set("Root", this.root)
	trailer.Set("Size", MakeInteger(int64(len(this.objects)+1)))
	// If encrypted!
	if this.crypter != nil {
		trailer.Set("Encrypt", this.encryptObj)
		trailer.Set("ID", this.ids)
		common.Log.Trace("Ids: %s", this.ids)
	}
	this.writer.WriteString("trailer\n")
	this.writer.WriteString(trailer.DefaultWriteString())
	this.writer.WriteString("\n")

	// Make offset reference.
	outStr = fmt.Sprintf("startxref\n%d\n", xrefOffset)
	this.writer.WriteString(outStr)
	this.writer.WriteString("%%EOF\n")
	w.Flush()

	return nil
}
