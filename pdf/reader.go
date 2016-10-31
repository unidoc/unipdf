/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/unidoc/unidoc/common"
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
	forms       *PdfObjectDictionary

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
		common.Log.Debug("ERROR: Fail to load structure (%s)", err)
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
		common.Log.Debug("ERROR: Failed to read root element catalog: %s", err)
		return err
	}
	pcatalog, ok := oc.(*PdfIndirectObject)
	if !ok {
		common.Log.Debug("ERROR: Missing catalog: (root %q) (trailer %s)", oc, *(this.parser.trailer))
		return errors.New("Missing catalog")
	}
	catalog, ok := (*pcatalog).PdfObject.(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid catalog (%s)", pcatalog.PdfObject)
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
	pageCount, ok := (*pages)["Count"].(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("ERROR: Pages count object invalid")
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

	// Outlines.
	this.outlineTree, err = this.loadOutlines()
	if err != nil {
		common.Log.Debug("ERROR: Failed to build outline tree (%s)", err)
		return err
	}

	// Get forms.
	this.forms, err = this.GetForms()
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
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
	}

	// Has outlines? Otherwise return an empty outlines structure.
	catalog := this.catalog
	outlinesObj, hasOutlines := (*catalog)["Outlines"]
	if !hasOutlines {
		return nil, nil
	}

	common.Log.Debug("-Has outlines")
	// Trace references to the object.
	outlineRootObj, err := this.traceToObject(outlinesObj)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read outlines")
		return nil, err
	}
	common.Log.Debug("Outline root: %v", outlineRootObj)

	if _, isNull := outlineRootObj.(*PdfObjectNull); isNull {
		common.Log.Debug("Outline root is null - no outlines")
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

	common.Log.Debug("Outline root dict: %v", dict)

	outlineTree, err := this.buildOutlineTree(dict)
	if err != nil {
		return nil, err
	}
	common.Log.Debug("Resulting outline tree: %v", outlineTree)

	return outlineTree, nil
}

// Recursive build outline tree.
func (this *PdfReader) buildOutlineTree(obj PdfObject) (*PdfOutlineTreeNode, error) {
	dict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
	if !ok {
		return nil, errors.New("Not a dictionary object")
	}
	common.Log.Debug("build outline tree: dict: %v", dict)

	if _, hasTitle := (*dict)["Title"]; hasTitle {
		// Outline item has a title.
		outlineItem, err := this.newPdfOutlineItemFromDict(dict)
		if err != nil {
			return nil, err
		}
		// Resolve the reference to next
		if nextObj, hasNext := (*dict)["Next"]; hasNext {
			nextObj, err = this.traceToObject(nextObj)
			if err != nil {
				return nil, err
			}
			nextObj = TraceToDirectObject(nextObj)
			if _, isNull := nextObj.(*PdfObjectNull); !isNull {
				nextDict, ok := nextObj.(*PdfObjectDictionary)
				if !ok {
					return nil, fmt.Errorf("Next not a dictionary object (%T)", nextObj)
				}
				outlineItem.Next, err = this.buildOutlineTree(nextDict)
				if err != nil {
					return nil, err
				}
			}
		}
		if firstObj, hasChildren := (*dict)["First"]; hasChildren {
			firstObj, err = this.traceToObject(firstObj)
			if err != nil {
				return nil, err
			}
			firstObj = TraceToDirectObject(firstObj)
			if _, isNull := firstObj.(*PdfObjectNull); !isNull {
				firstDict, ok := firstObj.(*PdfObjectDictionary)
				if !ok {
					return nil, fmt.Errorf("First not a dictionary object (%T)", firstObj)
				}
				outlineItem.First, err = this.buildOutlineTree(firstDict)
				if err != nil {
					return nil, err
				}
			}
		}
		return &outlineItem.PdfOutlineTreeNode, nil
	} else {
		// Outline dictionary (structure element).
		outline, err := newPdfOutlineFromDict(dict)
		if err != nil {
			return nil, err
		}

		if firstObj, hasChildren := (*dict)["First"]; hasChildren {
			firstObj, err = this.traceToObject(firstObj)
			if err != nil {
				return nil, err
			}
			firstObj = TraceToDirectObject(firstObj)
			if _, isNull := firstObj.(*PdfObjectNull); !isNull {
				firstDict, ok := firstObj.(*PdfObjectDictionary)
				if !ok {
					return nil, fmt.Errorf("First not a dictionary object (%T)", firstObj)
				}
				outline.First, err = this.buildOutlineTree(firstDict)
				if err != nil {
					return nil, err
				}
			}
		}
		return &outline.PdfOutlineTreeNode, nil
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
			common.Log.Debug("ERROR: Failed to read forms")
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
	err := this.traverseObjectData(formsDict)
	if err != nil {
		common.Log.Debug("ERROR: Unable to traverse AcroForms (%s)", err)
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
		p, err := this.newPdfPageFromDict(nodeDict)
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
		common.Log.Debug("ERROR: Table of content containing non Page/Pages object! (%s)", objType)
		return errors.New("Table of content containing non Page/Pages object!")
	}

	// A Pages object.  Update the parent.
	if parent != nil {
		(*nodeDict)["Parent"] = parent
	}

	// Resolve the object recursively.
	err := this.traverseObjectData(node)
	if err != nil {
		return err
	}

	kidsObj, err := this.parser.Trace((*nodeDict)["Kids"])
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
	common.Log.Debug("Kids: %s", kids)
	for idx, child := range *kids {
		child, ok := child.(*PdfIndirectObject)
		if !ok {
			common.Log.Debug("ERROR: Page not indirect object - (%s)", child)
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
 *
 * GH: Are we fully protected against circular references? (Add tests).
 */
func (this *PdfReader) traverseObjectData(o PdfObject) error {
	common.Log.Debug("Traverse object data")
	if _, isTraversed := this.traversed[o]; isTraversed {
		return nil
	}
	this.traversed[o] = true

	if io, isIndirectObj := o.(*PdfIndirectObject); isIndirectObj {
		common.Log.Debug("io: %s", io)
		common.Log.Debug("- %s", io.PdfObject)
		err := this.traverseObjectData(io.PdfObject)
		return err
	}

	if so, isStreamObj := o.(*PdfObjectStream); isStreamObj {
		err := this.traverseObjectData(so.PdfObjectDictionary)
		return err
	}

	if dict, isDict := o.(*PdfObjectDictionary); isDict {
		common.Log.Debug("- dict: %s", dict)
		for name, v := range *dict {
			if ref, isRef := v.(*PdfObjectReference); isRef {
				resolvedObj, _, err := this.resolveReference(ref)
				if err != nil {
					return err
				}
				(*dict)[name] = resolvedObj
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
		common.Log.Debug("- array: %s", arr)
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
// Rename to GetPageAsIndirectObject in the future?
func (this *PdfReader) GetPage(pageNumber int) (PdfObject, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
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
	common.Log.Debug("Page: %T %s", page, page)
	common.Log.Debug("- %T %s", page.PdfObject, page.PdfObject)

	return page, nil
}

// Get a page by the page number.
// Returns the PdfPage entry.
func (this *PdfReader) GetPageAsPdfPage(pageNumber int) (*PdfPage, error) {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return nil, fmt.Errorf("File need to be decrypted first")
	}
	if len(this.pageList) < pageNumber {
		return nil, errors.New("Invalid page number (page count too short)")
	}
	page := this.PageList[pageNumber-1]

	return page, nil
}
