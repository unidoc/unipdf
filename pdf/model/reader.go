/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

type PdfReader struct {
	parser      *PdfParser
	root        PdfObject
	pages       *PdfObjectDictionary
	pageList    []*PdfIndirectObject
	PageList    []*PdfPage
	pageCount   int
	catalog     *PdfObjectDictionary
	outlineTree *PdfOutlineTreeNode
	AcroForm    *PdfAcroForm

	modelManager *ModelManager

	// For tracking traversal (cache).
	traversed map[PdfObject]bool
}

func NewPdfReader(rs io.ReadSeeker) (*PdfReader, error) {
	pdfReader := &PdfReader{}
	pdfReader.traversed = map[PdfObject]bool{}

	pdfReader.modelManager = NewModelManager()

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

// Returns a string containing some information about the encryption method used.
// Subject to changes.  May be better to return a standardized struct with information.
// But challenging due to the many different types supported.
func (this *PdfReader) GetEncryptionMethod() string {
	crypter := this.parser.GetCrypter()
	str := crypter.Filter + " - "

	if crypter.V == 0 {
		str += "Undocumented algorithm"
	} else if crypter.V == 1 {
		// RC4 or AES (bits: 40)
		str += "RC4: 40 bits"
	} else if crypter.V == 2 {
		str += fmt.Sprintf("RC4: %d bits", crypter.Length)
	} else if crypter.V == 3 {
		str += "Unpublished algorithm"
	} else if crypter.V == 4 {
		// Look at CF, StmF, StrF
		str += fmt.Sprintf("Stream filter: %s - String filter: %s", crypter.StreamFilter, crypter.StringFilter)
		str += "; Crypt filters:"
		for name, cf := range crypter.CryptFilters {
			str += fmt.Sprintf(" - %s: %s (%d)", name, cf.Cfm, cf.Length)
		}
	}
	perms := crypter.GetAccessPermissions()
	str += fmt.Sprintf(" - %#v", perms)

	return str
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
		common.Log.Debug("ERROR: Fail to load structure (%s)", err)
		return false, err
	}

	return true, nil
}

// Check access rights and permissions for a specified password.  If either user/owner password is specified,
// full rights are granted, otherwise the access rights are specified by the Permissions flag.
//
// The bool flag indicates that the user can access and view the file.
// The AccessPermissions shows what access the user has for editing etc.
// An error is returned if there was a problem performing the authentication.
func (this *PdfReader) CheckAccessRights(password []byte) (bool, AccessPermissions, error) {
	return this.parser.CheckAccessRights(password)
}

// Loads the structure of the pdf file: pages, outlines, etc.
func (this *PdfReader) loadStructure() error {
	if this.parser.GetCrypter() != nil && !this.parser.IsAuthenticated() {
		return fmt.Errorf("File need to be decrypted first")
	}

	trailerDict := this.parser.GetTrailer()
	if trailerDict == nil {
		return fmt.Errorf("Missing trailer")
	}

	// Catalog.
	root, ok := trailerDict.Get("Root").(*PdfObjectReference)
	if !ok {
		return fmt.Errorf("Invalid Root (trailer: %s)", *trailerDict)
	}
	oc, err := this.parser.LookupByReference(*root)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read root element catalog: %s", err)
		return err
	}
	pcatalog, ok := oc.(*PdfIndirectObject)
	if !ok {
		common.Log.Debug("ERROR: Missing catalog: (root %q) (trailer %s)", oc, *trailerDict)
		return errors.New("Missing catalog")
	}
	catalog, ok := (*pcatalog).PdfObject.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid catalog (%s)", pcatalog.PdfObject)
		return errors.New("Invalid catalog")
	}
	common.Log.Trace("Catalog: %s", catalog)

	// Pages.
	pagesRef, ok := catalog.Get("Pages").(*PdfObjectReference)
	if !ok {
		return errors.New("Pages in catalog should be a reference")
	}
	op, err := this.parser.LookupByReference(*pagesRef)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read pages")
		return err
	}
	ppages, ok := op.(*PdfIndirectObject)
	if !ok {
		common.Log.Debug("ERROR: Pages object invalid")
		common.Log.Debug("op: %p", ppages)
		return errors.New("Pages object invalid")
	}
	pages, ok := ppages.PdfObject.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Pages object invalid (%s)", ppages)
		return errors.New("Pages object invalid")
	}
	pageCount, ok := pages.Get("Count").(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("ERROR: Pages count object invalid")
		return errors.New("Pages count invalid")
	}

	this.root = root
	this.catalog = catalog
	this.pages = pages
	this.pageCount = int(*pageCount)
	this.pageList = []*PdfIndirectObject{}

	traversedPageNodes := map[PdfObject]bool{}
	err = this.buildPageList(ppages, nil, traversedPageNodes)
	if err != nil {
		return err
	}
	common.Log.Trace("---")
	common.Log.Trace("TOC")
	common.Log.Trace("Pages")
	common.Log.Trace("%d: %s", len(this.pageList), this.pageList)

	// Outlines.
	this.outlineTree, err = this.loadOutlines()
	if err != nil {
		common.Log.Debug("ERROR: Failed to build outline tree (%s)", err)
		return err
	}

	// Load interactive forms and fields.
	this.AcroForm, err = this.loadForms()
	if err != nil {
		return err
	}

	return nil
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

func (this *PdfReader) loadOutlines() (*PdfOutlineTreeNode, error) {
	if this.parser.GetCrypter() != nil && !this.parser.IsAuthenticated() {
		return nil, fmt.Errorf("File need to be decrypted first")
	}

	// Has outlines? Otherwise return an empty outlines structure.
	catalog := this.catalog
	outlinesObj := catalog.Get("Outlines")
	if outlinesObj == nil {
		return nil, nil
	}

	common.Log.Trace("-Has outlines")
	// Trace references to the object.
	outlineRootObj, err := this.traceToObject(outlinesObj)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read outlines")
		return nil, err
	}
	common.Log.Trace("Outline root: %v", outlineRootObj)

	if _, isNull := outlineRootObj.(*PdfObjectNull); isNull {
		common.Log.Trace("Outline root is null - no outlines")
		return nil, nil
	}

	outlineRoot, ok := outlineRootObj.(*PdfIndirectObject)
	if !ok {
		return nil, errors.New("Outline root should be an indirect object")
	}

	dict, ok := outlineRoot.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return nil, errors.New("Outline indirect object should contain a dictionary")
	}

	common.Log.Trace("Outline root dict: %v", dict)

	outlineTree, _, err := this.buildOutlineTree(outlineRoot, nil, nil)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("Resulting outline tree: %v", outlineTree)

	return outlineTree, nil
}

// Recursive build outline tree.
// prev PdfObject,
// Input: The indirect object containing an Outlines or Outline item dictionary.
// Parent, Prev are the parent or previous node in the hierarchy.
// The function returns the corresponding tree node and the last node which is used
// for setting the Last pointer of the tree node structures.
func (this *PdfReader) buildOutlineTree(obj PdfObject, parent *PdfOutlineTreeNode, prev *PdfOutlineTreeNode) (*PdfOutlineTreeNode, *PdfOutlineTreeNode, error) {
	container, isInd := obj.(*PdfIndirectObject)
	if !isInd {
		return nil, nil, fmt.Errorf("Outline container not an indirect object %T", obj)
	}
	dict, ok := container.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return nil, nil, errors.New("Not a dictionary object")
	}
	common.Log.Trace("build outline tree: dict: %v (%v) p: %p", dict, container, container)

	if obj := dict.Get("Title"); obj != nil {
		// Outline item has a title. (required)
		outlineItem, err := this.newPdfOutlineItemFromIndirectObject(container)
		if err != nil {
			return nil, nil, err
		}
		outlineItem.Parent = parent
		outlineItem.Prev = prev

		if firstObj := dict.Get("First"); firstObj != nil {
			firstObj, err = this.traceToObject(firstObj)
			if err != nil {
				return nil, nil, err
			}
			if _, isNull := firstObj.(*PdfObjectNull); !isNull {
				first, last, err := this.buildOutlineTree(firstObj, &outlineItem.PdfOutlineTreeNode, nil)
				if err != nil {
					return nil, nil, err
				}
				outlineItem.First = first
				outlineItem.Last = last
			}
		}

		// Resolve the reference to next
		if nextObj := dict.Get("Next"); nextObj != nil {
			nextObj, err = this.traceToObject(nextObj)
			if err != nil {
				return nil, nil, err
			}
			if _, isNull := nextObj.(*PdfObjectNull); !isNull {
				next, last, err := this.buildOutlineTree(nextObj, parent, &outlineItem.PdfOutlineTreeNode)
				if err != nil {
					return nil, nil, err
				}
				outlineItem.Next = next
				return &outlineItem.PdfOutlineTreeNode, last, nil
			}
		}

		return &outlineItem.PdfOutlineTreeNode, &outlineItem.PdfOutlineTreeNode, nil
	} else {
		// Outline dictionary (structure element).

		outline, err := newPdfOutlineFromIndirectObject(container)
		if err != nil {
			return nil, nil, err
		}
		outline.Parent = parent
		//outline.Prev = parent

		if firstObj := dict.Get("First"); firstObj != nil {
			// Has children...
			firstObj, err = this.traceToObject(firstObj)
			if err != nil {
				return nil, nil, err
			}
			if _, isNull := firstObj.(*PdfObjectNull); !isNull {
				first, last, err := this.buildOutlineTree(firstObj, &outline.PdfOutlineTreeNode, nil)
				if err != nil {
					return nil, nil, err
				}
				outline.First = first
				outline.Last = last
			}
		}

		/*
			if nextObj, hasNext := (*dict)["Next"]; hasNext {
				nextObj, err = this.traceToObject(nextObj)
				if err != nil {
					return nil, nil, err
				}
				if _, isNull := nextObj.(*PdfObjectNull); !isNull {
					next, last, err := this.buildOutlineTree(nextObj, parent, &outline.PdfOutlineTreeNode)
					if err != nil {
						return nil, nil, err
					}
					outline.Next = next
					return &outline.PdfOutlineTreeNode, last, nil
				}
			}*/

		return &outline.PdfOutlineTreeNode, &outline.PdfOutlineTreeNode, nil
	}
}

// Get the outline tree.
func (this *PdfReader) GetOutlineTree() *PdfOutlineTreeNode {
	return this.outlineTree
}

// Return a flattened list of tree nodes and titles.
func (this *PdfReader) GetOutlinesFlattened() ([]*PdfOutlineTreeNode, []string, error) {
	outlineNodeList := []*PdfOutlineTreeNode{}
	flattenedTitleList := []string{}

	// Recursive flattening function.
	var flattenFunc func(*PdfOutlineTreeNode, *[]*PdfOutlineTreeNode, *[]string, int)
	flattenFunc = func(node *PdfOutlineTreeNode, outlineList *[]*PdfOutlineTreeNode, titleList *[]string, depth int) {
		if node == nil {
			return
		}
		if node.context == nil {
			common.Log.Debug("ERROR: Missing node.context") // Should not happen ever.
			return
		}

		if item, isItem := node.context.(*PdfOutlineItem); isItem {
			*outlineList = append(*outlineList, &item.PdfOutlineTreeNode)
			title := strings.Repeat(" ", depth*2) + string(*item.Title)
			*titleList = append(*titleList, title)
			if item.Next != nil {
				flattenFunc(item.Next, outlineList, titleList, depth)
			}
		}

		if node.First != nil {
			title := strings.Repeat(" ", depth*2) + "+"
			*titleList = append(*titleList, title)
			flattenFunc(node.First, outlineList, titleList, depth+1)
		}
	}
	flattenFunc(this.outlineTree, &outlineNodeList, &flattenedTitleList, 0)
	return outlineNodeList, flattenedTitleList, nil
}

func (this *PdfReader) loadForms() (*PdfAcroForm, error) {
	if this.parser.GetCrypter() != nil && !this.parser.IsAuthenticated() {
		return nil, fmt.Errorf("File need to be decrypted first")
	}

	// Has forms?
	catalog := this.catalog
	obj := catalog.Get("AcroForm")
	if obj == nil {
		// Nothing to load.
		return nil, nil
	}
	var err error
	obj, err = this.traceToObject(obj)
	if err != nil {
		return nil, err
	}
	obj = TraceToDirectObject(obj)
	if _, isNull := obj.(*PdfObjectNull); isNull {
		common.Log.Trace("Acroform is a null object (empty)\n")
		return nil, nil
	}

	formsDict, ok := obj.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Invalid AcroForm entry %T", obj)
		common.Log.Debug("Does not have forms")
		return nil, fmt.Errorf("Invalid acroform entry %T", obj)
	}
	common.Log.Trace("Has Acro forms")
	// Load it.

	// Ensure we have access to everything.
	common.Log.Trace("Traverse the Acroforms structure")
	err = this.traverseObjectData(formsDict)
	if err != nil {
		common.Log.Debug("ERROR: Unable to traverse AcroForms (%s)", err)
		return nil, err
	}

	// Create the acro forms object.
	acroForm, err := this.newPdfAcroFormFromDict(formsDict)
	if err != nil {
		return nil, err
	}

	return acroForm, nil
}

func (this *PdfReader) lookupPageByObject(obj PdfObject) (*PdfPage, error) {
	// can be indirect, direct, or reference
	// look up the corresponding page
	return nil, errors.New("Page not found")
}

// Build the table of contents.
// tree, ex: Pages -> Pages -> Pages -> Page
// Traverse through the whole thing recursively.
func (this *PdfReader) buildPageList(node *PdfIndirectObject, parent *PdfIndirectObject, traversedPageNodes map[PdfObject]bool) error {
	if node == nil {
		return nil
	}

	if _, alreadyTraversed := traversedPageNodes[node]; alreadyTraversed {
		common.Log.Debug("Cyclic recursion, skipping")
		return nil
	}
	traversedPageNodes[node] = true

	nodeDict, ok := node.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return errors.New("Node not a dictionary")
	}

	objType, ok := (*nodeDict).Get("Type").(*PdfObjectName)
	if !ok {
		return errors.New("Node missing Type (Required)")
	}
	common.Log.Trace("buildPageList node type: %s", *objType)
	if *objType == "Page" {
		p, err := this.newPdfPageFromDict(nodeDict)
		if err != nil {
			return err
		}
		p.setContainer(node)

		if parent != nil {
			// Set the parent (in case missing or incorrect).
			nodeDict.Set("Parent", parent)
		}
		this.pageList = append(this.pageList, node)
		this.PageList = append(this.PageList, p)

		return nil
	}
	if *objType != "Pages" {
		common.Log.Debug("ERROR: Table of content containing non Page/Pages object! (%s)", objType)
		return errors.New("Table of content containing non Page/Pages object!")
	}

	// A Pages object.  Update the parent.
	if parent != nil {
		nodeDict.Set("Parent", parent)
	}

	// Resolve the object recursively.
	err := this.traverseObjectData(node)
	if err != nil {
		return err
	}

	kidsObj, err := this.parser.Trace(nodeDict.Get("Kids"))
	if err != nil {
		common.Log.Debug("ERROR: Failed loading Kids object")
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
	common.Log.Trace("Kids: %s", kids)
	for idx, child := range *kids {
		child, ok := child.(*PdfIndirectObject)
		if !ok {
			common.Log.Debug("ERROR: Page not indirect object - (%s)", child)
			return errors.New("Page not indirect object")
		}
		(*kids)[idx] = child
		err = this.buildPageList(child, node, traversedPageNodes)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get the number of pages in the document.
func (this *PdfReader) GetNumPages() (int, error) {
	if this.parser.GetCrypter() != nil && !this.parser.IsAuthenticated() {
		return 0, fmt.Errorf("File need to be decrypted first")
	}
	return len(this.pageList), nil
}

// Resolves a reference, returning the object and indicates whether or not
// it was cached.
func (this *PdfReader) resolveReference(ref *PdfObjectReference) (PdfObject, bool, error) {
	cachedObj, isCached := this.parser.ObjCache[int(ref.ObjectNumber)]
	if !isCached {
		common.Log.Trace("Reader Lookup ref: %s", ref)
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
 *
 * GH: Are we fully protected against circular references? (Add tests).
 */
func (this *PdfReader) traverseObjectData(o PdfObject) error {
	common.Log.Trace("Traverse object data")
	if _, isTraversed := this.traversed[o]; isTraversed {
		common.Log.Trace("-Already traversed...")
		return nil
	}
	this.traversed[o] = true

	if io, isIndirectObj := o.(*PdfIndirectObject); isIndirectObj {
		common.Log.Trace("io: %s", io)
		common.Log.Trace("- %s", io.PdfObject)
		err := this.traverseObjectData(io.PdfObject)
		return err
	}

	if so, isStreamObj := o.(*PdfObjectStream); isStreamObj {
		err := this.traverseObjectData(so.PdfObjectDictionary)
		return err
	}

	if dict, isDict := o.(*PdfObjectDictionary); isDict {
		common.Log.Trace("- dict: %s", dict)
		for _, name := range dict.Keys() {
			v := dict.Get(name)
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj, _, err := this.resolveReference(ref)
				if err != nil {
					return err
				}
				dict.Set(name, resolvedObj)
				err = this.traverseObjectData(resolvedObj)
				if err != nil {
					return err
				}
			} else {
				err := this.traverseObjectData(v)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	if arr, isArray := o.(*PdfObjectArray); isArray {
		common.Log.Trace("- array: %s", arr)
		for idx, v := range *arr {
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj, _, err := this.resolveReference(ref)
				if err != nil {
					return err
				}
				(*arr)[idx] = resolvedObj

				err = this.traverseObjectData(resolvedObj)
				if err != nil {
					return err
				}
			} else {
				err := this.traverseObjectData(v)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	if _, isRef := o.(*PdfObjectReference); isRef {
		common.Log.Debug("ERROR: Reader tracing a reference!")
		return errors.New("Reader tracing a reference!")
	}

	return nil
}

// Get a page by the page number. Indirect object with type /Page.
func (this *PdfReader) GetPageAsIndirectObject(pageNumber int) (PdfObject, error) {
	if this.parser.GetCrypter() != nil && !this.parser.IsAuthenticated() {
		return nil, fmt.Errorf("File needs to be decrypted first")
	}
	if len(this.pageList) < pageNumber {
		return nil, errors.New("Invalid page number (page count too short)")
	}
	page := this.pageList[pageNumber-1]

	// Look up all references related to page and load everything.
	err := this.traverseObjectData(page)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("Page: %T %s", page, page)
	common.Log.Trace("- %T %s", page.PdfObject, page.PdfObject)

	return page, nil
}

// Get a page by the page number.
// Returns the PdfPage entry.
func (this *PdfReader) GetPage(pageNumber int) (*PdfPage, error) {
	if this.parser.GetCrypter() != nil && !this.parser.IsAuthenticated() {
		return nil, fmt.Errorf("File needs to be decrypted first")
	}
	if len(this.pageList) < pageNumber {
		return nil, errors.New("Invalid page number (page count too short)")
	}
	page := this.PageList[pageNumber-1]

	return page, nil
}

// Get optional content properties
func (this *PdfReader) GetOCProperties() (PdfObject, error) {
	dict := this.catalog
	obj := dict.Get("OCProperties")
	var err error
	obj, err = this.traceToObject(obj)
	if err != nil {
		return nil, err
	}

	// Resolve all references...
	// Should be pretty safe. Should not be referencing to pages or
	// any large structures.  Local structures and references
	// to OC Groups.
	err = this.traverseObjectData(obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Inspect the object types, subtypes and content in the PDF file.
func (this *PdfReader) Inspect() (map[string]int, error) {
	return this.parser.Inspect()
}

// GetObjectNums returns the object numbers of the PDF objects in the file
// Numbered objects are either indirect objects or stream objects.
// e.g. objNums := pdfReader.GetObjectNums()
// The underlying objects can then be accessed with
// pdfReader.GetIndirectObjectByNumber(objNums[0]) for the first available object.
func (r *PdfReader) GetObjectNums() []int {
	return r.parser.GetObjectNums()
}

// Get specific object number.
func (this *PdfReader) GetIndirectObjectByNumber(number int) (PdfObject, error) {
	obj, err := this.parser.LookupByNumber(number)
	return obj, err
}

func (this *PdfReader) GetTrailer() (*PdfObjectDictionary, error) {
	trailerDict := this.parser.GetTrailer()
	if trailerDict == nil {
		return nil, errors.New("Trailer missing")
	}

	return trailerDict, nil
}
