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

type PdfReader struct {
	parser    *PdfParser
	root      PdfObject
	pages     *PdfObjectDictionary
	pageList  []*PdfIndirectObject
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

func (this *PdfReader) loadStructure() error {
	if this.parser.crypter != nil && !this.parser.crypter.authenticated {
		return fmt.Errorf("File need to be decrypted first")
	}

	root, ok := (*(this.parser.trailer))["Root"].(*PdfObjectReference)
	if !ok {
		return fmt.Errorf("Invalid Root (trailer: %s)", *(this.parser.trailer))
	}

	oc, err := this.parser.LookupByReference(*root)
	if err != nil {
		common.Log.Error("Failed to read root element catalog: %s", err)
		return err
	}
	// Can the root be in an object stream?

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

	err = this.buildToc(ppages, nil)
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
	err = this.traverseObjectData(outlinesObj)
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
	err := this.traverseObjectData(formsDict)
	if err != nil {
		common.Log.Error("Unable to traverse AcroForms (%s)", err)
		return nil, err
	}

	return formsDict, nil
}

// Build the table of contents.
// tree, ex: Pages -> Pages -> Pages -> Page
// Traverse through the whole thing recursively.
func (this *PdfReader) buildToc(node *PdfIndirectObject, parent *PdfIndirectObject) error {
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
	common.Log.Debug("buildToc node type: %s", *objType)
	if *objType == "Page" {
		if parent != nil {
			// Set the parent (in case missing or incorrect).
			(*nodeDict)["Parent"] = parent
		}
		this.pageList = append(this.pageList, node)
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

	// Resolve the object recursively.
	err := this.traverseObjectData(node)
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
		child, ok := child.(*PdfIndirectObject)
		if !ok {
			common.Log.Error("Page not indirect object - (%s)", child)
			return errors.New("Page not indirect object")
		}
		(*kids)[idx] = child
		err = this.buildToc(child, node)
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

	// Look up all references related to page and load everything.
	err := this.traverseObjectData(page)
	if err != nil {
		return nil, err
	}
	common.Log.Debug("Page: %T %s", page, page)
	common.Log.Debug("- %T %s", page.PdfObject, page.PdfObject)

	return page, nil
}
