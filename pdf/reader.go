/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"errors"
	"fmt"
	"io"

	"github.com/unidoc/unidoc/common"
)

// The root catalog dictionary (table 28).
type PdfCatalog struct {
	Version           *PdfObjectName
	Extensions        PdfObject
	Pages             PdfPages
	PageLabels        PdfObject
	Names             PdfObject
	Dests             PdfObject
	ViewerPreferences PdfObject
	PageLayout        *PdfObjectName
	PageMode          *PdfObjectName
	Outlines          PdfObject //*PdfOutlines
	Threads           PdfObject
	OpenAction        PdfObject
	AA                PdfObject
	URI               PdfObject
	AcroForm          PdfObject
	Metadata          PdfObject
	StructTreeRoot    PdfObject
	MarkInfo          PdfObject
	Lang              *PdfObjectString
	SpiderInfo        PdfObject
	OutputIntents     PdfObject
	PieceInfo         PdfObject
	OCProperties      PdfObject
	Perms             PdfObject
	Legal             PdfObject
	Requirements      PdfObject
	Collection        PdfObject
	NeedsRendering    PdfObject
}

// Build the PDF catalog object based on the Catalog dictionary.
/*
func NewPdfCatalogFromDict(dict *PdfObjectDictionary) (*PdfCatalog, error) {
	if dict == nil {
		return nil, errors.New("Catalog dict is nil")
	}

	catalog := PdfCatalog{}

	dType, ok := dict["Type"].(*PdfObjectName)
	if !ok {
		return nil, errors.New("Missing/Invalid Catalog dictionary type")
	}
	if *dType != "Catalog" {
		return nil, errors.New("Catalog dictionary Type != Catalog")
	}

	if obj, isDefined := dict["Version"]; isDefined {
		version, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, errors.New("Version not a name object")
		}
		catalog.Version = version
	}

	if obj, isDefined := dict["Extensions"]; isDefined {
		catalog.Extensions = obj
	}

	if obj, isDefined := dict["Pages"]; isDefined {
		// Can be indirect.. Make a function that traces indirect objects to its object.
		pagesDict, ok := obj.(*PdfObjectDictionary)
		if !ok {
			return nil, errors.New("Pages value not a dictionary")
		}
		catalog.Pages, err = NewPagesFromDict(pagesDict)
		if err != nil {
			return nil, err
		}
	}

	if obj, isDefined := dict["PageLabels"]; isDefined {
		catalog.PageLabels = obj
	}

	if obj, isDefined := dict["Names"]; isDefined {
		catalog.Names = obj
	}

	if obj, isDefined := dict["Dests"]; isDefined {
		catalog.Dests = obj
	}

	if obj, isDefined := dict["ViewerPreferences"]; isDefined {
		catalog.ViewerPreferences = obj
	}

	if obj, isDefined := dict["PageLayout"]; isDefined {
		layout, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, errors.New("PageLayout not a name object")
		}
		catalog.PageLayout = layout
	}

	if obj, isDefined := dict["PageMode"]; isDefined {
		layout, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, errors.New("PageMode not a name object")
		}
		catalog.PageMode = layout
	}

	if obj, isDefined := dict["Outlines"]; isDefined {
		// Can be indirect.. Make a function that traces indirect objects to its object.
		// obj = TraceToDirect(obj)
		outlinesDict, ok := obj.(*PdfObjectDictionary)
		if !ok {
			return nil, errors.New("Outlines value not a dictionary")
		}
		catalog.Outlines, err = NewOutlinesFromDict(outlinesDict)
		if err != nil {
			return nil, err
		}
	}

	if obj, isDefined := dict["Threads"]; isDefined {
		catalog.Threads = obj
	}
	if obj, isDefined := dict["OpenAction"]; isDefined {
		catalog.OpenAction = obj
	}
	if obj, isDefined := dict["AA"]; isDefined {
		catalog.AA = obj
	}
	if obj, isDefined := dict["URI"]; isDefined {
		catalog.URI = obj
	}
	if obj, isDefined := dict["AcroForm"]; isDefined {
		catalog.URI = obj
	}
	if obj, isDefined := dict["Metadata"]; isDefined {
		catalog.URI = obj
	}
	if obj, isDefined := dict["StructTreeRoot"]; isDefined {
		catalog.URI = obj
	}
	if obj, isDefined := dict["MarkInfo"]; isDefined {
		catalog.URI = obj
	}
	if obj, isDefined := dict["URI"]; isDefined {
		catalog.URI = obj
	}

	if obj, isDefined := dict["Lang"]; isDefined {
		lang, ok := obj.(*PdfObjectString)
		if !ok {
			return nil, errors.New("Lang not a string object")
		}
		catalog.Lang = lang
	}

	if obj, isDefined := dict["SpiderInfo"]; isDefined {
		catalog.SpiderInfo = obj
	}
	if obj, isDefined := dict["OutputIntents"]; isDefined {
		catalog.OutputIntents = obj
	}
	if obj, isDefined := dict["PieceInfo"]; isDefined {
		catalog.PieceInfo = obj
	}
	if obj, isDefined := dict["OCProperties"]; isDefined {
		catalog.OCProperties = obj
	}
	if obj, isDefined := dict["Perms"]; isDefined {
		catalog.Perms = obj
	}
	if obj, isDefined := dict["Legal"]; isDefined {
		catalog.Legal = obj
	}
	if obj, isDefined := dict["Requirements"]; isDefined {
		catalog.Requirements = obj
	}
	if obj, isDefined := dict["Collection"]; isDefined {
		catalog.Collection = obj
	}
	if obj, isDefined := dict["NeedsRendering"]; isDefined {
		catalog.NeedsRendering = obj
	}

	return &catalog
}
*/

type PdfReader struct {
	parser    *PdfParser
	root      PdfObject
	pages     *PdfObjectDictionary
	pageList  []*PdfIndirectObject
	PageList  []*PdfPage
	pageCount int
	catalog   *PdfObjectDictionary
	outlines  []*PdfIndirectObject
	forms     *PdfObjectDictionary

	// For tracking traversal (cache).
	traversed map[PdfObject]bool
}

func NewPdfReader(rs io.ReadSeeker) (*PdfReader, error) {
	pdfReader := &PdfReader{}
	pdfReader.traversed = map[PdfObject]bool{}

	// Create the parser, loads the cross reference table and trailer.
	parser, err := NewParser(rs)
	if err != nil {
		return nil, err
	}
	pdfReader.parser = parser

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}

	// Load pdf doc structure if not encrypted.
	if !isEncrypted {
		err = pdfReader.loadStructure()
		if err != nil {
			return nil, err
		}
	}

	return pdfReader, nil
}

func (this *PdfReader) IsEncrypted() (bool, error) {
	return this.parser.IsEncrypted()
}

// Decrypt the PDF file with a specified password.  Also tries to
// decrypt with an empty password.  Returns true if successful,
// false otherwise.
func (this *PdfReader) Decrypt(password []byte) (bool, error) {
	success, err := this.parser.Decrypt(password)
	if err != nil {
		return false, err
	}
	if !success {
		return false, nil
	}

	err = this.loadStructure()
	if err != nil {
		common.Log.Error("Fail to load structure (%s)", err)
		return false, err
	}

	return true, nil
}

// Loads the structure of the pdf file: pages, outlines, etc.
func (this *PdfReader) loadStructure() error {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return fmt.Errorf("File need to be decrypted first")
	}

	// Catalog.
	root, ok := (*(this.parser.trailer))["Root"].(*PdfObjectReference)
	if !ok {
		return fmt.Errorf("Invalid Root (trailer: %s)", *(this.parser.trailer))
	}
	oc, err := this.parser.LookupByReference(*root)
	if err != nil {
		common.Log.Error("Failed to read root element catalog: %s", err)
		return err
	}
	pcatalog, ok := oc.(*PdfIndirectObject)
	if !ok {
		common.Log.Error("Missing catalog: (root %q) (trailer %s)", oc, *(this.parser.trailer))
		return errors.New("Missing catalog")
	}
	catalog, ok := (*pcatalog).PdfObject.(*PdfObjectDictionary)
	if !ok {
		common.Log.Error("Invalid catalog (%s)", pcatalog.PdfObject)
		return errors.New("Invalid catalog")
	}
	common.Log.Debug("Catalog: %s", catalog)

	// Pages.
	pagesRef, ok := (*catalog)["Pages"].(*PdfObjectReference)
	if !ok {
		return errors.New("Pages in catalog should be a reference")
	}
	op, err := this.parser.LookupByReference(*pagesRef)
	if err != nil {
		common.Log.Error("Failed to read pages")
		return err
	}
	ppages, ok := op.(*PdfIndirectObject)
	if !ok {
		common.Log.Error("Pages object invalid")
		common.Log.Error("op: %p", ppages)
		return errors.New("Pages object invalid")
	}
	pages, ok := ppages.PdfObject.(*PdfObjectDictionary)
	if !ok {
		common.Log.Error("Pages object invalid (%s)", ppages)
		return errors.New("Pages object invalid")
	}
	pageCount, ok := (*pages)["Count"].(*PdfObjectInteger)
	if !ok {
		common.Log.Error("Pages count object invalid")
		return errors.New("Pages count invalid")
	}

	this.root = root
	this.catalog = catalog
	this.pages = pages
	this.pageCount = int(*pageCount)
	this.pageList = []*PdfIndirectObject{}

	err = this.buildPageList(ppages, nil)
	if err != nil {
		return err
	}
	common.Log.Debug("---")
	common.Log.Debug("TOC")
	common.Log.Debug("Pages")
	common.Log.Debug("%d: %s", len(this.pageList), this.pageList)

	// Get outlines.
	this.outlines, err = this.GetOutlines()
	if err != nil {
		return err
	}
	// Get forms.
	this.forms, err = this.GetForms()
	if err != nil {
		return err
	}

	return nil
}

// Load the document outlines.
// Returns a list of the outermost layer of the Outlines dictionary,
// which then has connections to the inner layers.
// The inner layers are also fully traversed and references traced
// to their objects which are fully loaded in memory.
func (this *PdfReader) GetOutlines() ([]*PdfIndirectObject, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
	}
	outlinesList := []*PdfIndirectObject{}

	// Has outlines?
	catalog := this.catalog
	outlinesRef, hasOutlines := (*catalog)["Outlines"].(*PdfObjectReference)
	if !hasOutlines {
		return outlinesList, nil
	}

	common.Log.Debug("Has outlines")
	outlinesObj, err := this.parser.LookupByReference(*outlinesRef)
	if err != nil {
		common.Log.Error("Failed to read outlines")
		return outlinesList, err
	}
	common.Log.Debug("Traverse outlines")
	nofollowList := map[PdfObjectName]bool{
		"Parent": true,
	}
	// XXX: Make less generic, only follow certain fields...
	// make into our own static object.. follow whats needed..
	// BuildOutlinesFromDict

	err = this.traverseObjectData(outlinesObj, nofollowList)
	if err != nil {
		return nil, err
	}
	common.Log.Debug("Traverse outlines - done")

	outlines, ok := outlinesObj.(*PdfIndirectObject)
	if !ok {
		return outlinesList, nil
	}

	dict, ok := outlines.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return outlinesList, nil
	}

	traversed := map[*PdfIndirectObject]bool{}

	node, ok := (*dict)["First"].(*PdfIndirectObject)
	for ok {
		if _, alreadyTraversed := traversed[node]; alreadyTraversed {
			common.Log.Error("Circular outline reference")
			return outlinesList, errors.New("Circular outline reference")
		}
		traversed[node] = true
		dict, ok := node.PdfObject.(*PdfObjectDictionary)
		if !ok {
			common.Log.Debug("Invalid outline objects (not dict)")
			break
		}
		outlinesList = append(outlinesList, node)

		node, ok = (*dict)["Next"].(*PdfIndirectObject)
		if !ok {
			break
		}
	}

	return outlinesList, nil
}

//
// Trace to object.  Keeps a list of already visited references to avoid circular references.
//
// Example circular reference.
// 1 0 obj << /Next 2 0 R >>
// 2 0 obj << /Next 1 0 R >>
//
func (this *PdfReader) traceToObjectWrapper(obj PdfObject, refList map[*PdfObjectReference]bool) (PdfObject, error) {
	// Keep a list of references to avoid circular references.

	ref, isRef := obj.(*PdfObjectReference)
	if isRef {
		// Make sure not already visited (circular ref).
		if _, alreadyTraversed := refList[ref]; alreadyTraversed {
			return nil, errors.New("Circular reference")
		}
		refList[ref] = true
		obj, err := this.parser.LookupByReference(*ref)
		if err != nil {
			return nil, err
		}
		return this.traceToObjectWrapper(obj, refList)
	}

	// Not a reference, an object.  Can be indirect or any direct pdf object (other than reference).
	return obj, nil
}

func (this *PdfReader) traceToObject(obj PdfObject) (PdfObject, error) {
	refList := map[*PdfObjectReference]bool{}
	return this.traceToObjectWrapper(obj, refList)
}

// Testing.
func (this *PdfReader) buildOutlines2() (*PdfOutline, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
	}

	outlines := &PdfOutline{}

	// Has outlines?
	catalog := this.catalog
	outlinesObj, hasOutlines := (*catalog)["Outlines"]
	if !hasOutlines {
		return outlines, nil
	}

	common.Log.Debug("Has outlines")
	outlineRootObj, err := this.traceToObject(outlinesObj)
	if err != nil {
		common.Log.Error("Failed to read outlines")
		return nil, err
	}

	outlineRoot, ok := outlineRootObj.(*PdfIndirectObject)
	if !ok {
		return outlines, nil
	}

	dict, ok := outlineRoot.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return outlines, nil
	}

	outlineDict := TraceToDirectObject(outlinesObj)

	traversed := map[*PdfIndirectObject]bool{}

	node, ok := (*dict)["First"].(*PdfIndirectObject)
	for ok {
		if _, alreadyTraversed := traversed[node]; alreadyTraversed {
			common.Log.Error("Circular outline reference")
			return outlines, errors.New("Circular outline reference")
		}
		traversed[node] = true
		dict, ok := node.PdfObject.(*PdfObjectDictionary)
		if !ok {
			common.Log.Debug("Invalid outline objects (not dict)")
			break
		}
		outlinesList = append(outlinesList, node)

		node, ok = (*dict)["Next"].(*PdfIndirectObject)
		if !ok {
			break
		}
	}

	return outlinesList, nil
}

func (this *pdfReader) buildOutlineTree(obj PdfObject) error {
	dict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
	if !ok {
		return errors.New("Not a dictionary object")
	}

	if _, hasTitle := d["Title"]; hasTitle {
		// OutlineItem
		item := PdfOutlineItem{}
		newOutlineItemFromDict(d)
	}
}

// Get document form data.
func (this *PdfReader) GetForms() (*PdfObjectDictionary, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
	}
	// Has forms?
	catalog := this.catalog

	var formsDict *PdfObjectDictionary

	if dict, hasFormsDict := (*catalog)["AcroForm"].(*PdfObjectDictionary); hasFormsDict {
		common.Log.Debug("Has Acro forms - dictionary under Catalog")
		formsDict = dict
	} else if formsRef, hasFormsRef := (*catalog)["AcroForm"].(*PdfObjectReference); hasFormsRef {
		common.Log.Debug("Has Acro forms - Indirect object")
		formsObj, err := this.parser.LookupByReference(*formsRef)
		if err != nil {
			common.Log.Error("Failed to read forms")
			return nil, err
		}
		if iobj, ok := formsObj.(*PdfIndirectObject); ok {
			if dict, ok := iobj.PdfObject.(*PdfObjectDictionary); ok {
				formsDict = dict
			}
		}
	}
	if formsDict == nil {
		common.Log.Debug("Does not have forms")
		return nil, nil
	}

	common.Log.Debug("Has Acro forms")

	common.Log.Debug("Traverse the Acroforms structure")
	nofollowList := map[PdfObjectName]bool{
		"Parent": true,
	}
	err := this.traverseObjectData(formsDict, nofollowList)
	if err != nil {
		common.Log.Error("Unable to traverse AcroForms (%s)", err)
		return nil, err
	}

	return formsDict, nil
}

func (this *PdfReader) lookupPageByObject(obj PdfObject) (*PdfPage, error) {
	// can be indirect, direct, or reference
	// look up the corresponding page
	return nil, errors.New("Page not found")
}

// Build the table of contents.
// tree, ex: Pages -> Pages -> Pages -> Page
// Traverse through the whole thing recursively.
func (this *PdfReader) buildPageList(node *PdfIndirectObject, parent *PdfIndirectObject) error {
	if node == nil {
		return nil
	}

	nodeDict, ok := node.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return errors.New("Node not a dictionary")
	}

	objType, ok := (*nodeDict)["Type"].(*PdfObjectName)
	if !ok {
		return errors.New("Node missing Type (Required)")
	}
	common.Log.Debug("buildPageList node type: %s", *objType)
	if *objType == "Page" {
		p, err := NewPdfPageFromDict(nodeDict)
		if err != nil {
			return err
		}

		if parent != nil {
			// Set the parent (in case missing or incorrect).
			(*nodeDict)["Parent"] = parent
		}
		this.pageList = append(this.pageList, node)
		this.PageList = append(this.PageList, p)

		return nil
	}
	if *objType != "Pages" {
		common.Log.Error("Table of content containing non Page/Pages object! (%s)", objType)
		return errors.New("Table of content containing non Page/Pages object!")
	}

	// A Pages object.  Update the parent.
	if parent != nil {
		(*nodeDict)["Parent"] = parent
	}

	// Resolve the object recursively, not following Parents or Kids fields.
	// Later can refactor and use only one smart recursive function.
	nofollowList := map[PdfObjectName]bool{
		"Parent": true,
		"Kids":   true,
	}
	err := this.traverseObjectData(node, nofollowList)
	if err != nil {
		return err
	}

	kidsObj, err := this.parser.Trace((*nodeDict)["Kids"])
	if err != nil {
		common.Log.Error("Failed loading Kids object")
		return err
	}

	var kids *PdfObjectArray
	kids, ok = kidsObj.(*PdfObjectArray)
	if !ok {
		kidsIndirect, isIndirect := kidsObj.(*PdfIndirectObject)
		if !isIndirect {
			return errors.New("Invalid Kids object")
		}
		kids, ok = kidsIndirect.PdfObject.(*PdfObjectArray)
		if !ok {
			return errors.New("Invalid Kids indirect object")
		}
	}
	common.Log.Debug("Kids: %s", kids)
	for idx, child := range *kids {
		childRef, ok := child.(*PdfObjectReference)
		if !ok {
			return errors.New("Invalid kid, non-reference")
		}

		common.Log.Debug("look up ref %s", childRef)
		pchild, err := this.parser.LookupByReference(*childRef)
		if err != nil {
			common.Log.Error("Unable to lookup page ref")
			return errors.New("Unable to lookup page ref")
		}
		child, ok := pchild.(*PdfIndirectObject)
		if !ok {
			common.Log.Error("Page not indirect object - %s (%s)", childRef, pchild)
			return errors.New("Page not indirect object")
		}
		(*kids)[idx] = child
		err = this.buildPageList(child, node)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get the number of pages in the document.
func (this *PdfReader) GetNumPages() (int, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return -1, fmt.Errorf("File need to be decrypted first")
	}
	return len(this.pageList), nil
}

// Resolves a reference, returning the object and indicates whether or not
// it was cached.
func (this *PdfReader) resolveReference(ref *PdfObjectReference) (PdfObject, bool, error) {
	cachedObj, isCached := this.parser.ObjCache[int(ref.ObjectNumber)]
	if !isCached {
		common.Log.Debug("Reader Lookup ref: %s", ref)
		obj, err := this.parser.LookupByReference(*ref)
		if err != nil {
			return nil, false, err
		}
		this.parser.ObjCache[int(ref.ObjectNumber)] = obj
		return obj, false, nil
	}
	return cachedObj, true, nil
}

/*
 * Recursively traverse through the page object data and look up
 * references to indirect objects.
 * GH: Consider to define a smarter traversing engine, defining explicitly
 * - how deep we can go in terms of following certain Trees by name etc.
 * GH: Are we fully protected against circular references?
 */
func (this *PdfReader) traverseObjectData(o PdfObject, nofollowKeys map[PdfObjectName]bool) error {
	common.Log.Debug("Traverse object data")
	if _, isTraversed := this.traversed[o]; isTraversed {
		return nil
	}
	this.traversed[o] = true

	if io, isIndirectObj := o.(*PdfIndirectObject); isIndirectObj {
		common.Log.Debug("io: %s", io)
		common.Log.Debug("- %s", io.PdfObject)
		err := this.traverseObjectData(io.PdfObject, nofollowKeys)
		return err
	}

	if so, isStreamObj := o.(*PdfObjectStream); isStreamObj {
		err := this.traverseObjectData(so.PdfObjectDictionary, nofollowKeys)
		return err
	}

	if dict, isDict := o.(*PdfObjectDictionary); isDict {
		common.Log.Debug("- dict: %s", dict)
		for name, v := range *dict {
			if nofollowKeys != nil {
				if _, nofollow := nofollowKeys[name]; nofollow {
					// Do not retraverse up the tree.
					continue
				}
			}

			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj, _, err := this.resolveReference(ref)
				if err != nil {
					return err
				}
				(*dict)[name] = resolvedObj
				err = this.traverseObjectData(resolvedObj, nofollowKeys)
				if err != nil {
					return err
				}
			} else {
				err := this.traverseObjectData(v, nofollowKeys)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	if arr, isArray := o.(*PdfObjectArray); isArray {
		common.Log.Debug("- array: %s", arr)
		for idx, v := range *arr {
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj, _, err := this.resolveReference(ref)
				if err != nil {
					return err
				}
				(*arr)[idx] = resolvedObj

				err = this.traverseObjectData(resolvedObj, nofollowKeys)
				if err != nil {
					return err
				}
			} else {
				err := this.traverseObjectData(v, nofollowKeys)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	if _, isRef := o.(*PdfObjectReference); isRef {
		common.Log.Error("Reader tracing a reference!")
		return errors.New("Reader tracing a reference!")
	}

	return nil
}

// Get outlines referring to a specific page.  Only checks the outermost
// outlines.
func (this *PdfReader) GetOutlinesForPage(page PdfObject) ([]*PdfIndirectObject, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
	}
	pageOutlines := []*PdfIndirectObject{}

	for _, outlineObj := range this.outlines {
		dict, ok := (*outlineObj).PdfObject.(*PdfObjectDictionary)
		if !ok {
			common.Log.Error("Invalid outlines entry")
			return pageOutlines, fmt.Errorf("Invalid outlines entry")
		}

		if dest, hasDest := (*dict)["Dest"].(*PdfObjectArray); hasDest {
			if len(*dest) > 0 {
				if (*dest)[0] == page {
					pageOutlines = append(pageOutlines, outlineObj)
				}
			}
		}
		// Action: GoTo destination (page) can refer directly to a page.
		// TODO: Support more potential actions.  Make generic.
		// Can we make those sub conditionals cleaner?  Some kind of
		// generic tree traversal / unmarshalling.
		if dict, hasAdict := (*dict)["A"].(*PdfObjectDictionary); hasAdict {
			if s, hasS := (*dict)["S"].(*PdfObjectName); hasS {
				if *s == "GoTo" {
					if d, hasD := (*dict)["D"].(*PdfObjectArray); hasD {
						if len(*d) > 0 {
							if (*d)[0] == page {
								pageOutlines = append(pageOutlines, outlineObj)
							}
						}
					}
				}
			}
		}

		if a, hasA := (*dict)["A"].(*PdfIndirectObject); hasA {
			if dict, ok := a.PdfObject.(*PdfObjectDictionary); ok {
				if s, hasS := (*dict)["S"].(*PdfObjectName); hasS {
					if *s == "GoTo" {
						if d, hasD := (*dict)["D"].(*PdfObjectArray); hasD {
							if len(*d) > 0 {
								if (*d)[0] == page {
									pageOutlines = append(pageOutlines, outlineObj)
								}
							}
						}
					}
				}
			}
		}
	}
	return pageOutlines, nil
}

// Get a page by the page number.
// Indirect object with type /Page.
func (this *PdfReader) GetPage(pageNumber int) (PdfObject, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
	}
	if len(this.pageList) < pageNumber {
		return nil, errors.New("Invalid page number (page count too short)")
	}
	page := this.pageList[pageNumber-1]

	nofollowList := map[PdfObjectName]bool{
		"Parent": true,
	}
	// Look up all references related to page and load everything.
	err := this.traverseObjectData(page, nofollowList)
	if err != nil {
		return nil, err
	}
	common.Log.Debug("Page: %T %s", page, page)
	common.Log.Debug("- %T %s", page.PdfObject, page.PdfObject)

	return page, nil
}
