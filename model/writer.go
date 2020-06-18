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
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/core/security"
	"github.com/unidoc/unipdf/v3/core/security/crypt"
)

var pdfAuthor = ""
var pdfCreationDate time.Time
var pdfCreator = ""
var pdfKeywords = ""
var pdfModifiedDate time.Time
var pdfProducer = ""
var pdfSubject = ""
var pdfTitle = ""

type crossReference struct {
	Type int
	// Type 1
	Offset     int64
	Generation int64 // and Type 0
	// Type 2
	ObjectNumber int // and Type 0
	Index        int
}

func getPdfAuthor() string {
	return pdfAuthor
}

// SetPdfAuthor sets the Author attribute of the output PDF.
func SetPdfAuthor(author string) {
	pdfAuthor = author
}

func getPdfCreationDate() time.Time {
	return pdfCreationDate
}

// SetPdfCreationDate sets the CreationDate attribute of the output PDF.
func SetPdfCreationDate(creationDate time.Time) {
	pdfCreationDate = creationDate
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

func getPdfKeywords() string {
	return pdfKeywords
}

// SetPdfKeywords sets the Keywords attribute of the output PDF.
func SetPdfKeywords(keywords string) {
	pdfKeywords = keywords
}

func getPdfModifiedDate() time.Time {
	return pdfModifiedDate
}

// SetPdfModifiedDate sets the ModDate attribute of the output PDF.
func SetPdfModifiedDate(modifiedDate time.Time) {
	pdfCreationDate = modifiedDate
}

func getPdfProducer() string {
	licenseKey := license.GetLicenseKey()
	if len(pdfProducer) > 0 && (licenseKey.IsLicensed() || flag.Lookup("test.v") != nil) {
		return pdfProducer
	}

	// Return default.
	return fmt.Sprintf("UniDoc v%s (%s) - http://unidoc.io", getUniDocVersion(), licenseKey.TypeToString())
}

// SetPdfProducer sets the Producer attribute of the output PDF.
func SetPdfProducer(producer string) {
	pdfProducer = producer
}

func getPdfSubject() string {
	return pdfSubject
}

// SetPdfSubject sets the Subject attribute of the output PDF.
func SetPdfSubject(subject string) {
	pdfSubject = subject
}

func getPdfTitle() string {
	return pdfTitle
}

// SetPdfTitle sets the Title attribute of the output PDF.
func SetPdfTitle(title string) {
	pdfTitle = title
}

// PdfWriter handles outputing PDF content.
type PdfWriter struct {
	root        *core.PdfIndirectObject
	pages       *core.PdfIndirectObject
	pagesMap    map[core.PdfObject]struct{} // Pages lookup table.
	objects     []core.PdfObject            // Objects to write.
	objectsMap  map[core.PdfObject]struct{} // Quick lookup table.
	outlines    []*core.PdfIndirectObject
	outlineTree *PdfOutlineTreeNode
	catalog     *core.PdfObjectDictionary
	fields      []core.PdfObject
	infoObj     *core.PdfIndirectObject

	// `writer` is the buffered writer for writing, `writePos` tracks the current writing
	// position, needed to generate cross-reference tables, `werr` is the first error
	// encountered during writing. All writes after the first error become no-ops.
	writer   *bufio.Writer
	writePos int64 // Represents the current position within output file.
	werr     error

	// Encryption
	crypter     *core.PdfCrypt
	encryptDict *core.PdfObjectDictionary
	encryptObj  *core.PdfIndirectObject
	ids         *core.PdfObjectArray

	// PDF version
	majorVersion int
	minorVersion int

	// Force whether or not to use cross reference streams.
	// Otherwise is used/not used depending on the PDF version (1.5 and above).
	useCrossReferenceStream *bool

	// Objects to be followed up on prior to writing.
	// These are objects that are added and reference objects that are not included
	// for writing.
	// The map stores the object and the dictionary it is contained in.
	// Only way so we can access the dictionary entry later.
	pendingObjects map[core.PdfObject][]*core.PdfObjectDictionary

	// Forms.
	acroForm *PdfAcroForm

	optimizer              Optimizer
	crossReferenceMap      map[int]crossReference
	writeOffset            int64 // used by PdfAppender
	ObjNumOffset           int
	appendMode             bool
	appendToXrefs          core.XrefTable
	appendXrefPrevOffset   int64
	appendPrevRevisionSize int64
	// Map of object to object number for replacements.
	appendReplaceMap map[core.PdfObject]int64

	// Cache of objects traversed while resolving references.
	traversed map[core.PdfObject]struct{}
}

// NewPdfWriter initializes a new PdfWriter.
func NewPdfWriter() PdfWriter {
	w := PdfWriter{}

	w.objectsMap = map[core.PdfObject]struct{}{}
	w.objects = []core.PdfObject{}
	w.pendingObjects = map[core.PdfObject][]*core.PdfObjectDictionary{}
	w.traversed = map[core.PdfObject]struct{}{}

	// PDF Version. Can be changed if using more advanced features in PDF.
	// By default it is set to 1.3.
	w.majorVersion = 1
	w.minorVersion = 3

	// Creation info.
	infoDict := core.MakeDict()
	metadata := []struct {
		key   core.PdfObjectName
		value string
	}{
		{"Producer", getPdfProducer()},
		{"Creator", getPdfCreator()},
		{"Author", getPdfAuthor()},
		{"Subject", getPdfSubject()},
		{"Title", getPdfTitle()},
		{"Keywords", getPdfKeywords()},
	}
	for _, tuple := range metadata {
		if tuple.value != "" {
			infoDict.Set(tuple.key, core.MakeString(tuple.value))
		}
	}

	// Set creation and modified dates.
	if creationDate := getPdfCreationDate(); !creationDate.IsZero() {
		if cd, err := NewPdfDateFromTime(creationDate); err == nil {
			infoDict.Set("CreationDate", cd.ToPdfObject())
		}
	}
	if modifiedDate := getPdfModifiedDate(); !modifiedDate.IsZero() {
		if md, err := NewPdfDateFromTime(modifiedDate); err == nil {
			infoDict.Set("ModDate", md.ToPdfObject())
		}
	}

	infoObj := core.PdfIndirectObject{}
	infoObj.PdfObject = infoDict
	w.infoObj = &infoObj
	w.addObject(&infoObj)

	// Root catalog.
	catalog := core.PdfIndirectObject{}
	catalogDict := core.MakeDict()
	catalogDict.Set("Type", core.MakeName("Catalog"))
	catalog.PdfObject = catalogDict

	w.root = &catalog
	w.addObject(w.root)

	// Pages.
	pages := core.PdfIndirectObject{}
	pagedict := core.MakeDict()
	pagedict.Set("Type", core.MakeName("Pages"))
	kids := core.PdfObjectArray{}
	pagedict.Set("Kids", &kids)
	pagedict.Set("Count", core.MakeInteger(0))
	pages.PdfObject = pagedict

	w.pages = &pages
	w.pagesMap = map[core.PdfObject]struct{}{}
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
// If a skip map is provided and the writer is not set to append mode, the
// children objects of pages which are not present in the catalog are added to
// the map and the page dictionaries are replaced with null objects.
func (w *PdfWriter) copyObject(obj core.PdfObject,
	objectToObjectCopyMap map[core.PdfObject]core.PdfObject,
	skipMap map[core.PdfObject]struct{}, skip bool) core.PdfObject {
	if newObj, ok := objectToObjectCopyMap[obj]; ok {
		return newObj
	}

	newObj := obj
	skipUnusedPages := !w.appendMode && skipMap != nil
	switch t := obj.(type) {
	case *core.PdfObjectArray:
		arrObj := core.MakeArray()
		newObj = arrObj
		objectToObjectCopyMap[obj] = newObj
		for _, val := range t.Elements() {
			arrObj.Append(w.copyObject(val, objectToObjectCopyMap, skipMap, skip))
		}
	case *core.PdfObjectStreams:
		streamsObj := &core.PdfObjectStreams{PdfObjectReference: t.PdfObjectReference}
		newObj = streamsObj
		objectToObjectCopyMap[obj] = newObj
		for _, val := range t.Elements() {
			streamsObj.Append(w.copyObject(val, objectToObjectCopyMap, skipMap, skip))
		}
	case *core.PdfObjectStream:
		streamObj := &core.PdfObjectStream{
			Stream:             t.Stream,
			PdfObjectReference: t.PdfObjectReference,
		}
		newObj = streamObj
		objectToObjectCopyMap[obj] = newObj
		streamObj.PdfObjectDictionary = w.copyObject(t.PdfObjectDictionary, objectToObjectCopyMap, skipMap, skip).(*core.PdfObjectDictionary)
	case *core.PdfObjectDictionary:
		// Check if the object is a page dictionary and search it in the
		// writer pages. If not found, replace it with a null object and add
		// the chain of children objects to the skip map.
		var unused bool
		if skipUnusedPages && !skip {
			if dictType, _ := core.GetNameVal(t.Get("Type")); dictType == "Page" {
				_, ok := w.pagesMap[t]
				skip = !ok
				unused = skip
			}
		}

		dictObj := core.MakeDict()
		newObj = dictObj
		objectToObjectCopyMap[obj] = newObj
		for _, key := range t.Keys() {
			dictObj.Set(key, w.copyObject(t.Get(key), objectToObjectCopyMap, skipMap, skip))
		}

		// If an unused page dictionary is found, replace it with a null object.
		if unused {
			newObj = core.MakeNull()
			skip = false
		}
	case *core.PdfIndirectObject:
		indObj := &core.PdfIndirectObject{
			PdfObjectReference: t.PdfObjectReference,
		}
		newObj = indObj
		objectToObjectCopyMap[obj] = newObj
		indObj.PdfObject = w.copyObject(t.PdfObject, objectToObjectCopyMap, skipMap, skip)
	case *core.PdfObjectString:
		strObj := *t
		newObj = &strObj
		objectToObjectCopyMap[obj] = newObj
	case *core.PdfObjectName:
		nameObj := core.PdfObjectName(*t)
		newObj = &nameObj
		objectToObjectCopyMap[obj] = newObj
	case *core.PdfObjectNull:
		newObj = core.MakeNull()
		objectToObjectCopyMap[obj] = newObj
	case *core.PdfObjectInteger:
		intObj := core.PdfObjectInteger(*t)
		newObj = &intObj
		objectToObjectCopyMap[obj] = newObj
	case *core.PdfObjectReference:
		refObj := core.PdfObjectReference(*t)
		newObj = &refObj
		objectToObjectCopyMap[obj] = newObj
	case *core.PdfObjectFloat:
		floatObj := core.PdfObjectFloat(*t)
		newObj = &floatObj
		objectToObjectCopyMap[obj] = newObj
	case *core.PdfObjectBool:
		boolObj := core.PdfObjectBool(*t)
		newObj = &boolObj
		objectToObjectCopyMap[obj] = newObj
	case *pdfSignDictionary:
		sigObj := &pdfSignDictionary{
			PdfObjectDictionary: core.MakeDict(),
			handler:             t.handler,
			signature:           t.signature,
		}
		newObj = sigObj
		objectToObjectCopyMap[obj] = newObj
		for _, key := range t.Keys() {
			sigObj.Set(key, w.copyObject(t.Get(key), objectToObjectCopyMap, skipMap, skip))
		}
	default:
		common.Log.Info("TODO(a5i): implement copyObject for %+v", obj)
	}

	if skipUnusedPages && skip {
		skipMap[obj] = struct{}{}
	}

	return newObj
}

// copyObjects makes objects copy and set as working.
func (w *PdfWriter) copyObjects() {
	objectToObjectCopyMap := make(map[core.PdfObject]core.PdfObject)
	objects := make([]core.PdfObject, 0, len(w.objects))
	objectsMap := make(map[core.PdfObject]struct{}, len(w.objects))
	skipMap := make(map[core.PdfObject]struct{})
	for _, obj := range w.objects {
		newObject := w.copyObject(obj, objectToObjectCopyMap, skipMap, false)
		if _, ok := skipMap[obj]; ok {
			continue
		}
		objects = append(objects, newObject)
		objectsMap[newObject] = struct{}{}
	}

	w.objects = objects
	w.objectsMap = objectsMap
	w.infoObj = w.copyObject(w.infoObj, objectToObjectCopyMap, nil, false).(*core.PdfIndirectObject)
	w.root = w.copyObject(w.root, objectToObjectCopyMap, nil, false).(*core.PdfIndirectObject)
	if w.encryptObj != nil {
		w.encryptObj = w.copyObject(w.encryptObj, objectToObjectCopyMap, nil, false).(*core.PdfIndirectObject)
	}

	// Update replace map.
	if w.appendMode {
		appendReplaceMap := make(map[core.PdfObject]int64)
		for obj, replaceNum := range w.appendReplaceMap {
			if objCopy, has := objectToObjectCopyMap[obj]; has {
				appendReplaceMap[objCopy] = replaceNum
			} else {
				common.Log.Debug("ERROR: append mode - object copy not in map")
			}
		}
		w.appendReplaceMap = appendReplaceMap
	}
}

// SetVersion sets the PDF version of the output file.
func (w *PdfWriter) SetVersion(majorVersion, minorVersion int) {
	w.majorVersion = majorVersion
	w.minorVersion = minorVersion
}

// SetOCProperties sets the optional content properties.
func (w *PdfWriter) SetOCProperties(ocProperties core.PdfObject) error {
	dict := w.catalog

	if ocProperties != nil {
		common.Log.Trace("Setting OC Properties...")
		dict.Set("OCProperties", ocProperties)
		// Any risk of infinite loops?
		return w.addObjects(ocProperties)
	}

	return nil
}

// SetNamedDestinations sets the Names entry in the PDF catalog.
// See section 12.3.2.3 "Named Destinations" (p. 367 PDF32000_2008).
func (w *PdfWriter) SetNamedDestinations(names core.PdfObject) error {
	if names == nil {
		return nil
	}

	common.Log.Trace("Setting catalog Names...")
	w.catalog.Set("Names", names)
	return w.addObjects(names)
}

// SetPageLabels sets the PageLabels entry in the PDF catalog.
// See section 12.4.2 "Page Labels" (p. 382 PDF32000_2008).
func (w *PdfWriter) SetPageLabels(pageLabels core.PdfObject) error {
	if pageLabels == nil {
		return nil
	}

	common.Log.Trace("Setting catalog PageLabels...")
	w.catalog.Set("PageLabels", pageLabels)
	return w.addObjects(pageLabels)
}

// SetOptimizer sets the optimizer to optimize PDF before writing.
func (w *PdfWriter) SetOptimizer(optimizer Optimizer) {
	w.optimizer = optimizer
}

// GetOptimizer returns current PDF optimizer.
func (w *PdfWriter) GetOptimizer() Optimizer {
	return w.optimizer
}

func (w *PdfWriter) hasObject(obj core.PdfObject) bool {
	_, found := w.objectsMap[obj]
	return found
}

// Adds the object to list of objects and returns true if the obj was
// not already added. Returns false if the object was previously added.
func (w *PdfWriter) addObject(obj core.PdfObject) bool {
	hasObj := w.hasObject(obj)
	if !hasObj {
		err := core.ResolveReferencesDeep(obj, w.traversed)
		if err != nil {
			common.Log.Debug("ERROR: %v - skipping", err)
		}

		w.objects = append(w.objects, obj)
		w.objectsMap[obj] = struct{}{}
		return true
	}

	return false
}

func (w *PdfWriter) addObjects(obj core.PdfObject) error {
	common.Log.Trace("Adding objects!")

	if io, isIndirectObj := obj.(*core.PdfIndirectObject); isIndirectObj {
		common.Log.Trace("Indirect")
		common.Log.Trace("- %s (%p)", obj, io)
		common.Log.Trace("- %s", io.PdfObject)
		if w.addObject(io) {
			err := w.addObjects(io.PdfObject)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if so, isStreamObj := obj.(*core.PdfObjectStream); isStreamObj {
		common.Log.Trace("Stream")
		common.Log.Trace("- %s %p", obj, obj)
		if w.addObject(so) {
			err := w.addObjects(so.PdfObjectDictionary)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if dict, isDict := obj.(*core.PdfObjectDictionary); isDict {
		common.Log.Trace("Dict")
		common.Log.Trace("- %s", obj)
		for _, k := range dict.Keys() {
			v := core.ResolveReference(dict.Get(k))
			if k != "Parent" {
				err := w.addObjects(v)
				if err != nil {
					return err
				}
			} else {
				if _, parentIsNull := dict.Get("Parent").(*core.PdfObjectNull); parentIsNull {
					// Parent is null.  We can ignore it.
					continue
				}

				if hasObj := w.hasObject(v); !hasObj {
					common.Log.Debug("Parent obj not added yet!! %T %p %v", v, v, v)
					w.pendingObjects[v] = append(w.pendingObjects[v], dict)
					// Although it is missing at this point, it could be added later...
				}
				// How to handle the parent? Make sure it is present?
				if parentObj, parentIsRef := dict.Get("Parent").(*core.PdfObjectReference); parentIsRef {
					// Parent is a reference.  Means we can drop it?
					// Could refer to somewhere outside of the scope of the output doc.
					// Should be done by the reader already.
					// -> ERROR.
					common.Log.Debug("ERROR: Parent is a reference object - Cannot be in writer (needs to be resolved)")
					return fmt.Errorf("parent is a reference object - Cannot be in writer (needs to be resolved) - %s", parentObj)
				}
			}
		}
		return nil
	}

	if arr, isArray := obj.(*core.PdfObjectArray); isArray {
		common.Log.Trace("Array")
		common.Log.Trace("- %s", obj)
		if arr == nil {
			return errors.New("array is nil")
		}
		for _, v := range arr.Elements() {
			err := w.addObjects(core.ResolveReference(v))
			if err != nil {
				return err
			}
		}
		return nil
	}

	if _, isReference := obj.(*core.PdfObjectReference); isReference {
		// Should never be a reference, should already be resolved.
		common.Log.Debug("ERROR: Cannot be a reference - got %#v!", obj)
		return errors.New("reference not allowed")
	}

	return nil
}

// AddPage adds a page to the PDF file. The new page should be an indirect object.
func (w *PdfWriter) AddPage(page *PdfPage) error {
	procPage(page)
	obj := page.ToPdfObject()

	common.Log.Trace("==========")
	common.Log.Trace("Appending to page list %T", obj)

	pageObj, ok := core.GetIndirect(obj)
	if !ok {
		return errors.New("page should be an indirect object")
	}
	common.Log.Trace("%s", pageObj)
	common.Log.Trace("%s", pageObj.PdfObject)

	pDict, ok := core.GetDict(pageObj.PdfObject)
	if !ok {
		return errors.New("page object should be a dictionary")
	}

	otype, ok := core.GetName(pDict.Get("Type"))
	if !ok {
		return fmt.Errorf("page should have a Type key with a value of type name (%T)", pDict.Get("Type"))

	}
	if otype.String() != "Page" {
		return errors.New("field Type != Page (Required)")
	}

	// Copy inherited fields if missing.
	inheritedFields := []core.PdfObjectName{"Resources", "MediaBox", "CropBox", "Rotate"}
	parent, hasParent := core.GetIndirect(pDict.Get("Parent"))
	common.Log.Trace("Page Parent: %T (%v)", pDict.Get("Parent"), hasParent)
	for hasParent {
		common.Log.Trace("Page Parent: %T", parent)
		parentDict, ok := core.GetDict(parent.PdfObject)
		if !ok {
			return errors.New("invalid Parent object")
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
		parent, hasParent = core.GetIndirect(parentDict.Get("Parent"))
		common.Log.Trace("Next parent: %T", parentDict.Get("Parent"))
	}

	common.Log.Trace("Traversal done")

	// Update the dictionary.
	// Reuses the input object, updating the fields.
	pDict.Set("Parent", w.pages)
	pageObj.PdfObject = pDict

	// Add to Pages.
	pagesDict, ok := core.GetDict(w.pages.PdfObject)
	if !ok {
		return errors.New("invalid Pages obj (not a dict)")
	}
	kids, ok := core.GetArray(pagesDict.Get("Kids"))
	if !ok {
		return errors.New("invalid Pages Kids obj (not an array)")
	}
	kids.Append(pageObj)
	w.pagesMap[pDict] = struct{}{}

	pageCount, ok := core.GetInt(pagesDict.Get("Count"))
	if !ok {
		return errors.New("invalid Pages Count object (not an integer)")
	}
	// Update the count.
	*pageCount = *pageCount + 1

	w.addObject(pageObj)

	// Traverse the page and record all object references.
	err := w.addObjects(pDict)
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

	// Add font, if needed.
	fontName := core.PdfObjectName("UF1")
	if !p.Resources.HasFontByName(fontName) {
		p.Resources.SetFontByName(fontName, DefaultFont().ToPdfObject())
	}

	var ops []string
	ops = append(ops, "q")
	ops = append(ops, "BT")
	ops = append(ops, fmt.Sprintf("/%s 14 Tf", fontName.String()))
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
func (w *PdfWriter) AddOutlineTree(outlineTree *PdfOutlineTreeNode) {
	w.outlineTree = outlineTree
}

// Look for a specific key.  Returns a list of entries.
// What if something appears on many pages?
func (w *PdfWriter) seekByName(obj core.PdfObject, followKeys []string, key string) ([]core.PdfObject, error) {
	common.Log.Trace("Seek by name.. %T", obj)
	var list []core.PdfObject

	if io, isIndirectObj := obj.(*core.PdfIndirectObject); isIndirectObj {
		return w.seekByName(io.PdfObject, followKeys, key)
	}

	if so, isStreamObj := obj.(*core.PdfObjectStream); isStreamObj {
		return w.seekByName(so.PdfObjectDictionary, followKeys, key)
	}

	if dict, isDict := obj.(*core.PdfObjectDictionary); isDict {
		common.Log.Trace("Dict")
		for _, k := range dict.Keys() {
			v := dict.Get(k)
			if string(k) == key {
				list = append(list, v)
			}
			for _, followKey := range followKeys {
				if string(k) == followKey {
					common.Log.Trace("Follow key %s", followKey)
					items, err := w.seekByName(v, followKeys, key)
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
func (w *PdfWriter) SetForms(form *PdfAcroForm) error {
	w.acroForm = form
	return nil
}

// writeObject writes out an indirect / stream object.
func (w *PdfWriter) writeObject(num int, obj core.PdfObject) {
	common.Log.Trace("Write obj #%d\n", num)

	if pobj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		w.crossReferenceMap[num] = crossReference{Type: 1, Offset: w.writePos, Generation: pobj.GenerationNumber}
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		if sDict, ok := pobj.PdfObject.(*pdfSignDictionary); ok {
			sDict.fileOffset = w.writePos + int64(len(outStr))
		}
		if pobj.PdfObject == nil {
			common.Log.Debug("Error: indirect object's PdfObject should never be nil - setting to PdfObjectNull")
			pobj.PdfObject = core.MakeNull()
		}
		outStr += pobj.PdfObject.WriteString()
		outStr += "\nendobj\n"
		w.writeString(outStr)
		return
	}

	// TODO: Add a default encoder if Filter not specified?
	// Still need to make sure is encrypted.
	if pobj, isStream := obj.(*core.PdfObjectStream); isStream {
		w.crossReferenceMap[num] = crossReference{Type: 1, Offset: w.writePos, Generation: pobj.GenerationNumber}
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		outStr += pobj.PdfObjectDictionary.WriteString()
		outStr += "\nstream\n"
		w.writeString(outStr)
		w.writeBytes(pobj.Stream)
		w.writeString("\nendstream\nendobj\n")
		return
	}

	if ostreams, isObjStreams := obj.(*core.PdfObjectStreams); isObjStreams {
		w.crossReferenceMap[num] = crossReference{Type: 1, Offset: w.writePos, Generation: ostreams.GenerationNumber}
		outStr := fmt.Sprintf("%d 0 obj\n", num)
		var offsets []string
		var objData string
		var offset int64

		for index, obj := range ostreams.Elements() {
			io, isIndirect := obj.(*core.PdfIndirectObject)
			if !isIndirect {
				common.Log.Debug("ERROR: Object streams N %d contains non indirect pdf object %v", num, obj)
				continue
			}
			data := io.PdfObject.WriteString() + " "
			objData = objData + data
			offsets = append(offsets, fmt.Sprintf("%d %d", io.ObjectNumber, offset))
			w.crossReferenceMap[int(io.ObjectNumber)] = crossReference{Type: 2, ObjectNumber: num, Index: index}
			offset = offset + int64(len([]byte(data)))
		}
		offsetsStr := strings.Join(offsets, " ") + " "
		encoder := core.NewFlateEncoder()
		// For debugging:
		//encoder := core.NewRawEncoder()
		dict := encoder.MakeStreamDict()
		dict.Set(core.PdfObjectName("Type"), core.MakeName("ObjStm"))
		n := int64(ostreams.Len())
		dict.Set(core.PdfObjectName("N"), core.MakeInteger(n))
		first := int64(len(offsetsStr))
		dict.Set(core.PdfObjectName("First"), core.MakeInteger(first))

		data, _ := encoder.EncodeBytes([]byte(offsetsStr + objData))
		length := int64(len(data))

		dict.Set(core.PdfObjectName("Length"), core.MakeInteger(length))
		outStr += dict.WriteString()
		outStr += "\nstream\n"
		w.writeString(outStr)
		w.writeBytes(data)
		w.writeString("\nendstream\nendobj\n")
		return
	}

	w.writeString(obj.WriteString())
}

// Update all the object numbers prior to writing.
func (w *PdfWriter) updateObjectNumbers() {
	offset := w.ObjNumOffset

	// Update numbers
	i := 0
	for _, obj := range w.objects {
		objNum := int64(i + 1 + offset)
		increase := true
		if w.appendMode {
			if replaceNum, has := w.appendReplaceMap[obj]; has {
				objNum = replaceNum
				increase = false
			}
		}

		switch o := obj.(type) {
		case *core.PdfIndirectObject:
			o.ObjectNumber = objNum
			o.GenerationNumber = 0
		case *core.PdfObjectStream:
			o.ObjectNumber = objNum
			o.GenerationNumber = 0
		case *core.PdfObjectStreams:
			o.ObjectNumber = objNum
			o.GenerationNumber = 0
		default:
			common.Log.Debug("ERROR: Unknown type %T - skipping", o)
			continue
		}

		if increase {
			i++
		}
	}

	getObjNum := func(obj core.PdfObject) int64 {
		switch o := obj.(type) {
		case *core.PdfIndirectObject:
			return o.ObjectNumber
		case *core.PdfObjectStream:
			return o.ObjectNumber
		case *core.PdfObjectStreams:
			return o.ObjectNumber
		}
		return 0
	}
	// Sort the output by object numbers so that they appear in descending order.
	sort.SliceStable(w.objects, func(i, j int) bool {
		return getObjNum(w.objects[i]) < getObjNum(w.objects[j])
	})
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
func (w *PdfWriter) Encrypt(userPass, ownerPass []byte, options *EncryptOptions) error {
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
	crypter, info, err := core.PdfCryptNewEncrypt(cf, userPass, ownerPass, perm)
	if err != nil {
		return err
	}
	w.crypter = crypter
	if info.Major != 0 {
		w.SetVersion(info.Major, info.Minor)
	}
	w.encryptDict = info.Encrypt

	w.ids = core.MakeArray(core.MakeHexString(info.ID0), core.MakeHexString(info.ID1))

	// Make an object to contain the encryption dictionary.
	io := core.MakeIndirectObject(info.Encrypt)
	w.encryptObj = io
	w.addObject(io)

	return nil
}

// Wrapper function to handle writing out string.
func (w *PdfWriter) writeString(s string) {
	if w.werr != nil {
		return
	}
	n, err := w.writer.WriteString(s)
	w.writePos += int64(n)
	w.werr = err
}

// Wrapper function to handle writing out bytes.
func (w *PdfWriter) writeBytes(bb []byte) {
	if w.werr != nil {
		return
	}
	n, err := w.writer.Write(bb)
	w.writePos += int64(n)
	w.werr = err
}

// Write writes out the PDF.
func (w *PdfWriter) Write(writer io.Writer) error {
	common.Log.Trace("Write()")

	lk := license.GetLicenseKey()
	if lk == nil || !lk.IsLicensed() {
		fmt.Printf("Unlicensed copy of unidoc\n")
		fmt.Printf("To get rid of the watermark - Please get a license on https://unidoc.io\n")
	}

	// Outlines.
	if w.outlineTree != nil {
		common.Log.Trace("OutlineTree: %+v", w.outlineTree)
		outlines := w.outlineTree.ToPdfObject()
		common.Log.Trace("Outlines: %+v (%T, p:%p)", outlines, outlines, outlines)
		w.catalog.Set("Outlines", outlines)
		err := w.addObjects(outlines)
		if err != nil {
			return err
		}
	}

	// Form fields.
	if w.acroForm != nil {
		common.Log.Trace("Writing acro forms")
		indObj := w.acroForm.ToPdfObject()
		common.Log.Trace("AcroForm: %+v", indObj)
		w.catalog.Set("AcroForm", indObj)
		err := w.addObjects(indObj)
		if err != nil {
			return err
		}
	}

	// Check pending objects prior to write.
	for pendingObj, pendingObjDicts := range w.pendingObjects {
		if !w.hasObject(pendingObj) {
			common.Log.Debug("WARN Pending object %+v %T (%p) never added for writing", pendingObj, pendingObj, pendingObj)
			for _, pendingObjDict := range pendingObjDicts {
				for _, key := range pendingObjDict.Keys() {
					val := pendingObjDict.Get(key)
					if val == pendingObj {
						common.Log.Debug("Pending object found! and replaced with null")
						pendingObjDict.Set(key, core.MakeNull())
						break
					}
				}
			}
		}
	}
	// Set version in the catalog.
	w.catalog.Set("Version", core.MakeName(fmt.Sprintf("%d.%d", w.majorVersion, w.minorVersion)))

	// Make a copy of objects prior to optimizing as this can alter the objects.
	// TODO: Copying wastes memory. Might be worth making user responsible for handling properly.
	//       Is copy needed for optimization?
	w.copyObjects()

	if w.optimizer != nil {
		var err error
		w.objects, err = w.optimizer.Optimize(w.objects)
		if err != nil {
			return err
		}
		objMap := make(map[core.PdfObject]struct{}, len(w.objects))
		for _, obj := range w.objects {
			objMap[obj] = struct{}{}
		}
		w.objectsMap = objMap
	}

	w.writePos = w.writeOffset
	w.writer = bufio.NewWriter(writer)
	useCrossReferenceStream := w.majorVersion > 1 || (w.majorVersion == 1 && w.minorVersion > 4)
	if w.useCrossReferenceStream != nil {
		useCrossReferenceStream = *w.useCrossReferenceStream
	}

	// Make a map of objects within object streams (if used).
	objectsInObjectStreams := make(map[core.PdfObject]bool)
	for _, obj := range w.objects {
		if objStm, isObjectStreams := obj.(*core.PdfObjectStreams); isObjectStreams {
			// Objects in object streams can only be referenced from an xref stream (not table).
			useCrossReferenceStream = true
			for _, obj := range objStm.Elements() {
				objectsInObjectStreams[obj] = true
				if io, isIndirectObj := obj.(*core.PdfIndirectObject); isIndirectObj {
					objectsInObjectStreams[io.PdfObject] = true
				}
			}
		}
	}

	if useCrossReferenceStream && w.majorVersion == 1 && w.minorVersion < 5 {
		w.minorVersion = 5
	}

	if w.appendMode {
		w.writeString("\n")
	} else {
		w.writeString(fmt.Sprintf("%%PDF-%d.%d\n", w.majorVersion, w.minorVersion))
		w.writeString("%âãÏÓ\n")
	}

	w.updateObjectNumbers()

	// Write objects
	common.Log.Trace("Writing %d obj", len(w.objects))
	w.crossReferenceMap = make(map[int]crossReference)
	w.crossReferenceMap[0] = crossReference{Type: 0, ObjectNumber: 0, Generation: 0xFFFF}
	if w.appendToXrefs.ObjectMap != nil {
		for idx, xref := range w.appendToXrefs.ObjectMap {
			if idx == 0 {
				continue
			}
			if xref.XType == core.XrefTypeObjectStream {
				cr := crossReference{Type: 2, ObjectNumber: xref.OsObjNumber, Index: xref.OsObjIndex}
				w.crossReferenceMap[idx] = cr
			}
			if xref.XType == core.XrefTypeTableEntry {
				cr := crossReference{Type: 1, ObjectNumber: xref.ObjectNumber, Offset: xref.Offset}
				w.crossReferenceMap[idx] = cr
			}
		}
	}

	// Write out indirect/stream objects that are not in object streams.
	for _, obj := range w.objects {
		if skip := objectsInObjectStreams[obj]; skip {
			continue
		}

		objectNumber := int64(0)
		switch t := obj.(type) {
		case *core.PdfIndirectObject:
			objectNumber = t.ObjectNumber
		case *core.PdfObjectStream:
			objectNumber = t.ObjectNumber
		case *core.PdfObjectStreams:
			objectNumber = t.ObjectNumber
		default:
			common.Log.Debug("ERROR: Unsupported type in writer objects: %T", obj)
			return ErrTypeCheck
		}

		// Encrypt prior to writing.
		// Encrypt dictionary should not be encrypted.
		if w.crypter != nil && obj != w.encryptObj {
			err := w.crypter.Encrypt(obj, int64(objectNumber), 0)
			if err != nil {
				common.Log.Debug("ERROR: Failed encrypting (%s)", err)
				return err
			}
		}
		w.writeObject(int(objectNumber), obj)
	}

	xrefOffset := w.writePos
	var maxIndex int
	for idx := range w.crossReferenceMap {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	// Write trailer / cross reference stream (depending on which used).
	if useCrossReferenceStream {
		crossObjNumber := maxIndex + 1
		w.crossReferenceMap[crossObjNumber] = crossReference{Type: 1, ObjectNumber: crossObjNumber, Offset: xrefOffset}
		crossReferenceData := bytes.NewBuffer(nil)

		index := core.MakeArray()
		for idx := 0; idx <= maxIndex; {
			// Find next to write.
			for ; idx <= maxIndex; idx++ {
				ref, has := w.crossReferenceMap[idx]
				if has && (!w.appendMode || w.appendMode && (ref.Type == 1 && ref.Offset >= w.appendPrevRevisionSize || ref.Type == 0)) {
					break
				}
			}

			var j int
			for j = idx + 1; j <= maxIndex; j++ {
				ref, has := w.crossReferenceMap[j]
				if has && (!w.appendMode || w.appendMode && (ref.Type == 1 && ref.Offset > w.appendPrevRevisionSize)) {
					continue
				}
				break
			}
			index.Append(core.MakeInteger(int64(idx)), core.MakeInteger(int64(j-idx)))

			for k := idx; k < j; k++ {
				ref := w.crossReferenceMap[k]
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

			idx = j + 1
		}

		crossReferenceStream, err := core.MakeStream(crossReferenceData.Bytes(), core.NewFlateEncoder())
		if err != nil {
			return err
		}
		crossReferenceStream.ObjectNumber = int64(crossObjNumber)
		crossReferenceStream.PdfObjectDictionary.Set("Type", core.MakeName("XRef"))
		crossReferenceStream.PdfObjectDictionary.Set("W", core.MakeArray(core.MakeInteger(1), core.MakeInteger(4), core.MakeInteger(2)))
		crossReferenceStream.PdfObjectDictionary.Set("Index", index)
		crossReferenceStream.PdfObjectDictionary.Set("Size", core.MakeInteger(int64(crossObjNumber+1)))
		crossReferenceStream.PdfObjectDictionary.Set("Info", w.infoObj)
		crossReferenceStream.PdfObjectDictionary.Set("Root", w.root)
		if w.appendMode && w.appendXrefPrevOffset > 0 {
			crossReferenceStream.PdfObjectDictionary.Set("Prev", core.MakeInteger(w.appendXrefPrevOffset))
		}
		// If encrypted!
		if w.crypter != nil {
			crossReferenceStream.Set("Encrypt", w.encryptObj)
			crossReferenceStream.Set("ID", w.ids)
			common.Log.Trace("Ids: %s", w.ids)
		}

		w.writeObject(int(crossReferenceStream.ObjectNumber), crossReferenceStream)
	} else {
		w.writeString("xref\r\n")
		for idx := 0; idx <= maxIndex; {
			// Find next to write.
			for ; idx <= maxIndex; idx++ {
				ref, has := w.crossReferenceMap[idx]
				if has && (!w.appendMode || w.appendMode && (ref.Type == 1 && ref.Offset >= w.appendPrevRevisionSize || ref.Type == 0)) {
					break
				}
			}

			var j int
			for j = idx + 1; j <= maxIndex; j++ {
				ref, has := w.crossReferenceMap[j]
				if has && (!w.appendMode || w.appendMode && (ref.Type == 1 && ref.Offset > w.appendPrevRevisionSize)) {
					continue
				}
				break
			}

			outStr := fmt.Sprintf("%d %d\r\n", idx, j-idx)
			w.writeString(outStr)
			for k := idx; k < j; k++ {
				ref := w.crossReferenceMap[k]
				switch ref.Type {
				case 0:
					outStr = fmt.Sprintf("%.10d %.5d f\r\n", 0, 65535)
					w.writeString(outStr)
				case 1:
					outStr = fmt.Sprintf("%.10d %.5d n\r\n", ref.Offset, 0)
					w.writeString(outStr)
				}
			}

			idx = j + 1
		}

		// Generate & write trailer
		trailer := core.MakeDict()
		trailer.Set("Info", w.infoObj)
		trailer.Set("Root", w.root)
		trailer.Set("Size", core.MakeInteger(int64(maxIndex+1)))
		if w.appendMode && w.appendXrefPrevOffset > 0 {
			trailer.Set("Prev", core.MakeInteger(w.appendXrefPrevOffset))
		}
		// If encrypted!
		if w.crypter != nil {
			trailer.Set("Encrypt", w.encryptObj)
			trailer.Set("ID", w.ids)
			common.Log.Trace("Ids: %s", w.ids)
		}
		w.writeString("trailer\n")
		w.writeString(trailer.WriteString())
		w.writeString("\n")
	}

	// Make offset reference.
	outStr := fmt.Sprintf("startxref\n%d\n", xrefOffset)
	w.writeString(outStr)
	w.writeString("%%EOF\n")

	if w.werr == nil {
		w.werr = w.writer.Flush()
	}

	return w.werr
}
