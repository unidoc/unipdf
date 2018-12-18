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
	"github.com/unidoc/unidoc/pdf/core/security"
)

// PdfReader represents a PDF file reader. It is a frontend to the lower level parsing mechanism and provides
// a higher level access to work with PDF structure and information, such as the page structure etc.
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

	modelManager *modelManager

	// For tracking traversal (cache).
	traversed map[PdfObject]bool
	rs        io.ReadSeeker
}

// NewPdfReader returns a new PdfReader for an input io.ReadSeeker interface. Can be used to read PDF from
// memory or file. Immediately loads and traverses the PDF structure including pages and page contents (if
// not encrypted).
func NewPdfReader(rs io.ReadSeeker) (*PdfReader, error) {
	pdfReader := &PdfReader{}
	pdfReader.rs = rs
	pdfReader.traversed = map[PdfObject]bool{}

	pdfReader.modelManager = newModelManager()

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

// PdfVersion returns version of the PDF file.
func (r *PdfReader) PdfVersion() Version {
	return r.parser.PdfVersion()
}

// IsEncrypted returns true if the PDF file is encrypted.
func (r *PdfReader) IsEncrypted() (bool, error) {
	return r.parser.IsEncrypted()
}

// GetEncryptionMethod returns a descriptive information string about the encryption method used.
func (r *PdfReader) GetEncryptionMethod() string {
	crypter := r.parser.GetCrypter()
	return crypter.String()
}

// Decrypt decrypts the PDF file with a specified password.  Also tries to
// decrypt with an empty password.  Returns true if successful,
// false otherwise.
func (r *PdfReader) Decrypt(password []byte) (bool, error) {
	success, err := r.parser.Decrypt(password)
	if err != nil {
		return false, err
	}
	if !success {
		return false, nil
	}

	err = r.loadStructure()
	if err != nil {
		common.Log.Debug("ERROR: Fail to load structure (%s)", err)
		return false, err
	}

	return true, nil
}

// CheckAccessRights checks access rights and permissions for a specified password.  If either user/owner
// password is specified,  full rights are granted, otherwise the access rights are specified by the
// Permissions flag.
//
// The bool flag indicates that the user can access and view the file.
// The AccessPermissions shows what access the user has for editing etc.
// An error is returned if there was a problem performing the authentication.
func (r *PdfReader) CheckAccessRights(password []byte) (bool, security.Permissions, error) {
	return r.parser.CheckAccessRights(password)
}

// Loads the structure of the pdf file: pages, outlines, etc.
func (r *PdfReader) loadStructure() error {
	if r.parser.GetCrypter() != nil && !r.parser.IsAuthenticated() {
		return fmt.Errorf("file need to be decrypted first")
	}

	trailerDict := r.parser.GetTrailer()
	if trailerDict == nil {
		return fmt.Errorf("missing trailer")
	}

	// Catalog.
	root, ok := trailerDict.Get("Root").(*PdfObjectReference)
	if !ok {
		return fmt.Errorf("invalid Root (trailer: %s)", trailerDict)
	}
	oc, err := r.parser.LookupByReference(*root)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read root element catalog: %s", err)
		return err
	}
	pcatalog, ok := oc.(*PdfIndirectObject)
	if !ok {
		common.Log.Debug("ERROR: Missing catalog: (root %q) (trailer %s)", oc, *trailerDict)
		return errors.New("missing catalog")
	}
	catalog, ok := (*pcatalog).PdfObject.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid catalog (%s)", pcatalog.PdfObject)
		return errors.New("invalid catalog")
	}
	common.Log.Trace("Catalog: %s", catalog)

	// Pages.
	pagesRef, ok := catalog.Get("Pages").(*PdfObjectReference)
	if !ok {
		return errors.New("pages in catalog should be a reference")
	}
	op, err := r.parser.LookupByReference(*pagesRef)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read pages")
		return err
	}
	ppages, ok := op.(*PdfIndirectObject)
	if !ok {
		common.Log.Debug("ERROR: Pages object invalid")
		common.Log.Debug("op: %p", ppages)
		return errors.New("pages object invalid")
	}
	pages, ok := ppages.PdfObject.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Pages object invalid (%s)", ppages)
		return errors.New("pages object invalid")
	}
	pageCount, ok := pages.Get("Count").(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("ERROR: Pages count object invalid")
		return errors.New("pages count invalid")
	}

	r.root = root
	r.catalog = catalog
	r.pages = pages
	r.pageCount = int(*pageCount)
	r.pageList = []*PdfIndirectObject{}

	traversedPageNodes := map[PdfObject]bool{}
	err = r.buildPageList(ppages, nil, traversedPageNodes)
	if err != nil {
		return err
	}
	common.Log.Trace("---")
	common.Log.Trace("TOC")
	common.Log.Trace("Pages")
	common.Log.Trace("%d: %s", len(r.pageList), r.pageList)

	// Outlines.
	r.outlineTree, err = r.loadOutlines()
	if err != nil {
		common.Log.Debug("ERROR: Failed to build outline tree (%s)", err)
		return err
	}

	// Load interactive forms and fields.
	r.AcroForm, err = r.loadForms()
	if err != nil {
		return err
	}

	return nil
}

// Trace to object.  Keeps a list of already visited references to avoid circular references.
//
// Example circular reference.
// 1 0 obj << /Next 2 0 R >>
// 2 0 obj << /Next 1 0 R >>
//
func (r *PdfReader) traceToObjectWrapper(obj PdfObject, refList map[*PdfObjectReference]bool) (PdfObject, error) {
	// Keep a list of references to avoid circular references.

	ref, isRef := obj.(*PdfObjectReference)
	if isRef {
		// Make sure not already visited (circular ref).
		if _, alreadyTraversed := refList[ref]; alreadyTraversed {
			return nil, errors.New("circular reference")
		}
		refList[ref] = true
		obj, err := r.parser.LookupByReference(*ref)
		if err != nil {
			return nil, err
		}
		return r.traceToObjectWrapper(obj, refList)
	}

	// Not a reference, an object.  Can be indirect or any direct pdf object (other than reference).
	return obj, nil
}

func (r *PdfReader) traceToObject(obj PdfObject) (PdfObject, error) {
	refList := map[*PdfObjectReference]bool{}
	return r.traceToObjectWrapper(obj, refList)
}

func (r *PdfReader) loadOutlines() (*PdfOutlineTreeNode, error) {
	if r.parser.GetCrypter() != nil && !r.parser.IsAuthenticated() {
		return nil, fmt.Errorf("file need to be decrypted first")
	}

	// Has outlines? Otherwise return an empty outlines structure.
	catalog := r.catalog
	outlinesObj := catalog.Get("Outlines")
	if outlinesObj == nil {
		return nil, nil
	}

	common.Log.Trace("-Has outlines")
	// Trace references to the object.
	outlineRootObj, err := r.traceToObject(outlinesObj)
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
		return nil, errors.New("outline root should be an indirect object")
	}

	dict, ok := outlineRoot.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return nil, errors.New("outline indirect object should contain a dictionary")
	}

	common.Log.Trace("Outline root dict: %v", dict)

	outlineTree, _, err := r.buildOutlineTree(outlineRoot, nil, nil)
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
func (r *PdfReader) buildOutlineTree(obj PdfObject, parent *PdfOutlineTreeNode, prev *PdfOutlineTreeNode) (*PdfOutlineTreeNode, *PdfOutlineTreeNode, error) {
	container, isInd := obj.(*PdfIndirectObject)
	if !isInd {
		return nil, nil, fmt.Errorf("outline container not an indirect object %T", obj)
	}
	dict, ok := container.PdfObject.(*PdfObjectDictionary)
	if !ok {
		return nil, nil, errors.New("not a dictionary object")
	}
	common.Log.Trace("build outline tree: dict: %v (%v) p: %p", dict, container, container)

	if obj := dict.Get("Title"); obj != nil {
		// Outline item has a title. (required)
		outlineItem, err := r.newPdfOutlineItemFromIndirectObject(container)
		if err != nil {
			return nil, nil, err
		}
		outlineItem.Parent = parent
		outlineItem.Prev = prev

		if firstObj := dict.Get("First"); firstObj != nil {
			firstObj, err = r.traceToObject(firstObj)
			if err != nil {
				return nil, nil, err
			}
			if _, isNull := firstObj.(*PdfObjectNull); !isNull {
				first, last, err := r.buildOutlineTree(firstObj, &outlineItem.PdfOutlineTreeNode, nil)
				if err != nil {
					return nil, nil, err
				}
				outlineItem.First = first
				outlineItem.Last = last
			}
		}

		// Resolve the reference to next
		if nextObj := dict.Get("Next"); nextObj != nil {
			nextObj, err = r.traceToObject(nextObj)
			if err != nil {
				return nil, nil, err
			}
			if _, isNull := nextObj.(*PdfObjectNull); !isNull {
				next, last, err := r.buildOutlineTree(nextObj, parent, &outlineItem.PdfOutlineTreeNode)
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
			firstObj, err = r.traceToObject(firstObj)
			if err != nil {
				return nil, nil, err
			}
			if _, isNull := firstObj.(*PdfObjectNull); !isNull {
				first, last, err := r.buildOutlineTree(firstObj, &outline.PdfOutlineTreeNode, nil)
				if err != nil {
					return nil, nil, err
				}
				outline.First = first
				outline.Last = last
			}
		}

		/*
			if nextObj, hasNext := (*dict)["Next"]; hasNext {
				nextObj, err = r.traceToObject(nextObj)
				if err != nil {
					return nil, nil, err
				}
				if _, isNull := nextObj.(*PdfObjectNull); !isNull {
					next, last, err := r.buildOutlineTree(nextObj, parent, &outline.PdfOutlineTreeNode)
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

// GetOutlineTree returns the outline tree.
func (r *PdfReader) GetOutlineTree() *PdfOutlineTreeNode {
	return r.outlineTree
}

// GetOutlinesFlattened returns a flattened list of tree nodes and titles.
func (r *PdfReader) GetOutlinesFlattened() ([]*PdfOutlineTreeNode, []string, error) {
	var outlineNodeList []*PdfOutlineTreeNode
	var flattenedTitleList []string

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
			title := strings.Repeat(" ", depth*2) + item.Title.Str()
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
	flattenFunc(r.outlineTree, &outlineNodeList, &flattenedTitleList, 0)
	return outlineNodeList, flattenedTitleList, nil
}

// loadForms loads the AcroForm.
func (r *PdfReader) loadForms() (*PdfAcroForm, error) {
	if r.parser.GetCrypter() != nil && !r.parser.IsAuthenticated() {
		return nil, fmt.Errorf("file need to be decrypted first")
	}

	// Has forms?
	catalog := r.catalog
	obj := catalog.Get("AcroForm")
	if obj == nil {
		// Nothing to load.
		return nil, nil
	}
	var err error
	obj, err = r.traceToObject(obj)
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
		return nil, fmt.Errorf("invalid acroform entry %T", obj)
	}
	common.Log.Trace("Has Acro forms")
	// Load it.

	// Ensure we have access to everything.
	common.Log.Trace("Traverse the Acroforms structure")
	err = r.traverseObjectData(formsDict)
	if err != nil {
		common.Log.Debug("ERROR: Unable to traverse AcroForms (%s)", err)
		return nil, err
	}

	// Create the acro forms object.
	acroForm, err := r.newPdfAcroFormFromDict(formsDict)
	if err != nil {
		return nil, err
	}

	return acroForm, nil
}

func (r *PdfReader) lookupPageByObject(obj PdfObject) (*PdfPage, error) {
	// can be indirect, direct, or reference
	// look up the corresponding page
	return nil, errors.New("page not found")
}

// Build the table of contents.
// tree, ex: Pages -> Pages -> Pages -> Page
// Traverse through the whole thing recursively.
func (r *PdfReader) buildPageList(node *PdfIndirectObject, parent *PdfIndirectObject, traversedPageNodes map[PdfObject]bool) error {
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
		return errors.New("node not a dictionary")
	}

	objType, ok := (*nodeDict).Get("Type").(*PdfObjectName)
	if !ok {
		return errors.New("node missing Type (Required)")
	}
	common.Log.Trace("buildPageList node type: %s (%+v)", *objType, node)
	if *objType == "Page" {
		p, err := r.newPdfPageFromDict(nodeDict)
		if err != nil {
			return err
		}
		p.setContainer(node)

		if parent != nil {
			// Set the parent (in case missing or incorrect).
			nodeDict.Set("Parent", parent)
		}
		r.pageList = append(r.pageList, node)
		r.PageList = append(r.PageList, p)

		return nil
	}
	if *objType != "Pages" {
		common.Log.Debug("ERROR: Table of content containing non Page/Pages object! (%s)", objType)
		return errors.New("table of content containing non Page/Pages object!")
	}

	// A Pages object.  Update the parent.
	if parent != nil {
		nodeDict.Set("Parent", parent)
	}

	// Resolve the object recursively.
	err := r.traverseObjectData(node)
	if err != nil {
		return err
	}

	kidsObj, err := r.parser.Resolve(nodeDict.Get("Kids"))
	if err != nil {
		common.Log.Debug("ERROR: Failed loading Kids object")
		return err
	}

	var kids *PdfObjectArray
	kids, ok = kidsObj.(*PdfObjectArray)
	if !ok {
		kidsIndirect, isIndirect := kidsObj.(*PdfIndirectObject)
		if !isIndirect {
			return errors.New("invalid Kids object")
		}
		kids, ok = kidsIndirect.PdfObject.(*PdfObjectArray)
		if !ok {
			return errors.New("invalid Kids indirect object")
		}
	}
	common.Log.Trace("Kids: %s", kids)
	for idx, child := range kids.Elements() {
		child, ok := child.(*PdfIndirectObject)
		if !ok {
			common.Log.Debug("ERROR: Page not indirect object - (%s)", child)
			return errors.New("page not indirect object")
		}
		kids.Set(idx, child)
		err = r.buildPageList(child, node, traversedPageNodes)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetNumPages returns the number of pages in the document.
func (r *PdfReader) GetNumPages() (int, error) {
	if r.parser.GetCrypter() != nil && !r.parser.IsAuthenticated() {
		return 0, fmt.Errorf("file need to be decrypted first")
	}
	return len(r.pageList), nil
}

// Resolves a reference, returning the object and indicates whether or not
// it was cached.
func (r *PdfReader) resolveReference(ref *PdfObjectReference) (PdfObject, bool, error) {
	cachedObj, isCached := r.parser.ObjCache[int(ref.ObjectNumber)]
	if !isCached {
		common.Log.Trace("Reader Lookup ref: %s", ref)
		obj, err := r.parser.LookupByReference(*ref)
		if err != nil {
			return nil, false, err
		}
		r.parser.ObjCache[int(ref.ObjectNumber)] = obj
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
func (r *PdfReader) traverseObjectData(o PdfObject) error {
	common.Log.Trace("Traverse object data")
	if _, isTraversed := r.traversed[o]; isTraversed {
		common.Log.Trace("-Already traversed...")
		return nil
	}
	r.traversed[o] = true

	if io, isIndirectObj := o.(*PdfIndirectObject); isIndirectObj {
		common.Log.Trace("io: %s", io)
		common.Log.Trace("- %s", io.PdfObject)
		err := r.traverseObjectData(io.PdfObject)
		return err
	}

	if so, isStreamObj := o.(*PdfObjectStream); isStreamObj {
		err := r.traverseObjectData(so.PdfObjectDictionary)
		return err
	}

	if dict, isDict := o.(*PdfObjectDictionary); isDict {
		common.Log.Trace("- dict: %s", dict)
		for _, name := range dict.Keys() {
			v := dict.Get(name)
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj, _, err := r.resolveReference(ref)
				if err != nil {
					return err
				}
				dict.Set(name, resolvedObj)
				err = r.traverseObjectData(resolvedObj)
				if err != nil {
					return err
				}
			} else {
				err := r.traverseObjectData(v)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	if arr, isArray := o.(*PdfObjectArray); isArray {
		common.Log.Trace("- array: %s", arr)
		for idx, v := range arr.Elements() {
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj, _, err := r.resolveReference(ref)
				if err != nil {
					return err
				}
				arr.Set(idx, resolvedObj)

				err = r.traverseObjectData(resolvedObj)
				if err != nil {
					return err
				}
			} else {
				err := r.traverseObjectData(v)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	if _, isRef := o.(*PdfObjectReference); isRef {
		common.Log.Debug("ERROR: Reader tracing a reference!")
		return errors.New("reader tracing a reference!")
	}

	return nil
}

// GetPageAsIndirectObject returns the indirect object representing a page fro a given page number.
// Indirect object with type /Page.
func (r *PdfReader) GetPageAsIndirectObject(pageNumber int) (PdfObject, error) {
	if r.parser.GetCrypter() != nil && !r.parser.IsAuthenticated() {
		return nil, fmt.Errorf("file needs to be decrypted first")
	}
	if len(r.pageList) < pageNumber {
		return nil, errors.New("invalid page number (page count too short)")
	}
	page := r.pageList[pageNumber-1]

	// Look up all references related to page and load everything.
	// TODO: Use of traverse object data will be limited when lazy-loading is supported.
	err := r.traverseObjectData(page)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("Page: %T %s", page, page)
	common.Log.Trace("- %T %s", page.PdfObject, page.PdfObject)

	return page, nil
}

// PageFromIndirectObject returns the PdfPage and page number for a given indirect object.
func (r *PdfReader) PageFromIndirectObject(ind *PdfIndirectObject) (*PdfPage, int, error) {
	if len(r.PageList) != len(r.pageList) {
		return nil, 0, errors.New("page list invalid")
	}

	for i, pageind := range r.pageList {
		if pageind == ind {
			return r.PageList[i], i + 1, nil
		}
	}
	return nil, 0, errors.New("page not found")
}

// GetPage returns the PdfPage model for the specified page number.
func (r *PdfReader) GetPage(pageNumber int) (*PdfPage, error) {
	if r.parser.GetCrypter() != nil && !r.parser.IsAuthenticated() {
		return nil, fmt.Errorf("file needs to be decrypted first")
	}
	if len(r.pageList) < pageNumber {
		return nil, errors.New("invalid page number (page count too short)")
	}
	idx := pageNumber - 1
	if idx < 0 {
		return nil, fmt.Errorf("page numbering must start at 1")
	}
	page := r.PageList[idx]

	return page, nil
}

// GetOCProperties returns the optional content properties PdfObject.
func (r *PdfReader) GetOCProperties() (PdfObject, error) {
	dict := r.catalog
	obj := dict.Get("OCProperties")
	var err error
	obj, err = r.traceToObject(obj)
	if err != nil {
		return nil, err
	}

	// Resolve all references...
	// Should be pretty safe. Should not be referencing to pages or
	// any large structures.  Local structures and references
	// to OC Groups.
	err = r.traverseObjectData(obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Inspect inspects the object types, subtypes and content in the PDF file returning a map of
// object type to number of instances of each.
func (r *PdfReader) Inspect() (map[string]int, error) {
	return r.parser.Inspect()
}

// GetObjectNums returns the object numbers of the PDF objects in the file
// Numbered objects are either indirect objects or stream objects.
// e.g. objNums := pdfReader.GetObjectNums()
// The underlying objects can then be accessed with
// pdfReader.GetIndirectObjectByNumber(objNums[0]) for the first available object.
func (r *PdfReader) GetObjectNums() []int {
	return r.parser.GetObjectNums()
}

// GetIndirectObjectByNumber retrieves and returns a specific PdfObject by object number.
func (r *PdfReader) GetIndirectObjectByNumber(number int) (PdfObject, error) {
	obj, err := r.parser.LookupByNumber(number)
	return obj, err
}

// GetTrailer returns the PDF's trailer dictionary.
func (r *PdfReader) GetTrailer() (*PdfObjectDictionary, error) {
	trailerDict := r.parser.GetTrailer()
	if trailerDict == nil {
		return nil, errors.New("trailer missing")
	}

	return trailerDict, nil
}
