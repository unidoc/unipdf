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

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/core/security"
)

// PdfReader represents a PDF file reader. It is a frontend to the lower level parsing mechanism and provides
// a higher level access to work with PDF structure and information, such as the page structure etc.
type PdfReader struct {
	parser         *core.PdfParser
	root           core.PdfObject
	pagesContainer *core.PdfIndirectObject
	pages          *core.PdfObjectDictionary
	pageList       []*core.PdfIndirectObject
	PageList       []*PdfPage
	pageCount      int
	catalog        *core.PdfObjectDictionary
	outlineTree    *PdfOutlineTreeNode
	AcroForm       *PdfAcroForm

	modelManager *modelManager

	// Lazy loading: When enabled reference objects need to be resolved (via lookup, disk access) rather
	// than loading entire document into memory on load.
	isLazy bool

	// For tracking traversal (cache).
	traversed map[core.PdfObject]struct{}
	rs        io.ReadSeeker
}

// NewPdfReader returns a new PdfReader for an input io.ReadSeeker interface. Can be used to read PDF from
// memory or file. Immediately loads and traverses the PDF structure including pages and page contents (if
// not encrypted). Loads entire document structure into memory.
// Alternatively a lazy-loading reader can be created with NewPdfReaderLazy which loads only references,
// and references are loaded from disk into memory on an as-needed basis.
func NewPdfReader(rs io.ReadSeeker) (*PdfReader, error) {
	pdfReader := &PdfReader{
		rs:           rs,
		traversed:    map[core.PdfObject]struct{}{},
		modelManager: newModelManager(),
		isLazy:       false,
	}

	// Create the parser, loads the cross reference table and trailer.
	parser, err := core.NewParser(rs)
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

// NewPdfReaderLazy creates a new PdfReader for `rs` in lazy-loading mode. The difference
// from NewPdfReader is that in lazy-loading mode, objects are only loaded into memory when needed
// rather than entire structure being loaded into memory on reader creation.
// Note that it may make sense to use the lazy-load reader when processing only parts of files,
// rather than loading entire file into memory. Example: splitting a few pages from a large PDF file.
func NewPdfReaderLazy(rs io.ReadSeeker) (*PdfReader, error) {
	pdfReader := &PdfReader{
		rs:           rs,
		traversed:    map[core.PdfObject]struct{}{},
		modelManager: newModelManager(),
		isLazy:       true,
	}

	// Create the parser, loads the cross reference table and trailer.
	parser, err := core.NewParser(rs)
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
func (r *PdfReader) PdfVersion() core.Version {
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
	root, ok := trailerDict.Get("Root").(*core.PdfObjectReference)
	if !ok {
		return fmt.Errorf("invalid Root (trailer: %s)", trailerDict)
	}
	oc, err := r.parser.LookupByReference(*root)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read root element catalog: %s", err)
		return err
	}
	pcatalog, ok := oc.(*core.PdfIndirectObject)
	if !ok {
		common.Log.Debug("ERROR: Missing catalog: (root %q) (trailer %s)", oc, *trailerDict)
		return errors.New("missing catalog")
	}
	catalog, ok := (*pcatalog).PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Invalid catalog (%s)", pcatalog.PdfObject)
		return errors.New("invalid catalog")
	}
	common.Log.Trace("Catalog: %s", catalog)

	// Pages.
	pagesRef, ok := catalog.Get("Pages").(*core.PdfObjectReference)
	if !ok {
		return errors.New("pages in catalog should be a reference")
	}
	op, err := r.parser.LookupByReference(*pagesRef)
	if err != nil {
		common.Log.Debug("ERROR: Failed to read pages")
		return err
	}
	ppages, ok := op.(*core.PdfIndirectObject)
	if !ok {
		common.Log.Debug("ERROR: Pages object invalid")
		common.Log.Debug("op: %p", ppages)
		return errors.New("pages object invalid")
	}
	pages, ok := ppages.PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("ERROR: Pages object invalid (%s)", ppages)
		return errors.New("pages object invalid")
	}
	pageCount, ok := core.GetInt(pages.Get("Count"))
	if !ok {
		common.Log.Debug("ERROR: Pages count object invalid")
		return errors.New("pages count invalid")
	}
	if _, ok = core.GetName(pages.Get("Type")); !ok {
		common.Log.Debug("Pages dict Type field not set. Setting Type to Pages.")
		pages.Set("Type", core.MakeName("Pages"))
	}

	r.root = root
	r.catalog = catalog
	r.pages = pages
	r.pagesContainer = ppages
	r.pageCount = int(*pageCount)
	r.pageList = []*core.PdfIndirectObject{}

	traversedPageNodes := map[core.PdfObject]struct{}{}
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
	outlineRootObj := core.ResolveReference(outlinesObj)
	common.Log.Trace("Outline root: %v", outlineRootObj)

	if isNull := core.IsNullObject(outlineRootObj); isNull {
		common.Log.Trace("Outline root is null - no outlines")
		return nil, nil
	}

	outlineRoot, ok := outlineRootObj.(*core.PdfIndirectObject)
	if !ok {
		if _, ok := core.GetDict(outlineRootObj); !ok {
			common.Log.Debug("Invalid outline root - skipping")
			return nil, nil
		}

		common.Log.Debug("Outline root is a dict. Should be an indirect object")
		outlineRoot = core.MakeIndirectObject(outlineRootObj)
	}

	dict, ok := outlineRoot.PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		return nil, errors.New("outline indirect object should contain a dictionary")
	}

	common.Log.Trace("Outline root dict: %v", dict)

	outlineTree, _, err := r.buildOutlineTree(outlineRoot, nil, nil, nil)
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
func (r *PdfReader) buildOutlineTree(obj core.PdfObject, parent *PdfOutlineTreeNode, prev *PdfOutlineTreeNode, visited map[core.PdfObject]struct{}) (*PdfOutlineTreeNode, *PdfOutlineTreeNode, error) {
	if visited == nil {
		visited = map[core.PdfObject]struct{}{}
	}
	visited[obj] = struct{}{}

	container, isInd := obj.(*core.PdfIndirectObject)
	if !isInd {
		return nil, nil, fmt.Errorf("outline container not an indirect object %T", obj)
	}
	dict, ok := container.PdfObject.(*core.PdfObjectDictionary)
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

		// Build outline tree for node children.
		firstObj := core.ResolveReference(dict.Get("First"))
		if _, processed := visited[firstObj]; firstObj != nil && firstObj != container && !processed {
			if !core.IsNullObject(firstObj) {
				first, last, err := r.buildOutlineTree(firstObj, &outlineItem.PdfOutlineTreeNode, nil, visited)
				if err != nil {
					common.Log.Debug("DEBUG: could not build outline item tree: %v. Skipping node children.", err)
				} else {
					outlineItem.First = first
					outlineItem.Last = last
				}
			}
		}

		// Build outline tree for the next item.
		nextObj := core.ResolveReference(dict.Get("Next"))
		if _, processed := visited[nextObj]; nextObj != nil && nextObj != container && !processed {
			if !core.IsNullObject(nextObj) {
				next, last, err := r.buildOutlineTree(nextObj, parent, &outlineItem.PdfOutlineTreeNode, visited)
				if err != nil {
					common.Log.Debug("DEBUG: could not build outline tree for Next node: %v. Skipping node.", err)
				} else {
					outlineItem.Next = next
					return &outlineItem.PdfOutlineTreeNode, last, nil
				}
			}
		}

		return &outlineItem.PdfOutlineTreeNode, &outlineItem.PdfOutlineTreeNode, nil
	}

	// Outline dictionary (structure element).
	outline, err := newPdfOutlineFromIndirectObject(container)
	if err != nil {
		return nil, nil, err
	}
	outline.Parent = parent

	if firstObj := dict.Get("First"); firstObj != nil {
		// Has children...
		firstObj = core.ResolveReference(firstObj)
		firstObjDirect := core.TraceToDirectObject(firstObj)
		if _, isNull := firstObjDirect.(*core.PdfObjectNull); !isNull && firstObjDirect != nil {
			first, last, err := r.buildOutlineTree(firstObj, &outline.PdfOutlineTreeNode, nil, visited)
			if err != nil {
				common.Log.Debug("DEBUG: could not build outline tree: %v. Skipping node children.", err)
			} else {
				outline.First = first
				outline.Last = last
			}
		}
	}

	return &outline.PdfOutlineTreeNode, &outline.PdfOutlineTreeNode, nil
}

// GetOutlineTree returns the outline tree.
func (r *PdfReader) GetOutlineTree() *PdfOutlineTreeNode {
	return r.outlineTree
}

// GetOutlinesFlattened returns a flattened list of tree nodes and titles.
// NOTE: for most use cases, it is recommended to use the high-level GetOutlines
// method instead, which also provides information regarding the destination
// of the outline items.
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

		item, isItem := node.context.(*PdfOutlineItem)
		if isItem {
			*outlineList = append(*outlineList, &item.PdfOutlineTreeNode)
			title := strings.Repeat(" ", depth*2) + item.Title.Decoded()
			*titleList = append(*titleList, title)
		}

		if node.First != nil {
			title := strings.Repeat(" ", depth*2) + "+"
			*titleList = append(*titleList, title)
			flattenFunc(node.First, outlineList, titleList, depth+1)
		}

		if isItem && item.Next != nil {
			flattenFunc(item.Next, outlineList, titleList, depth)
		}
	}
	flattenFunc(r.outlineTree, &outlineNodeList, &flattenedTitleList, 0)
	return outlineNodeList, flattenedTitleList, nil
}

// GetOutlines returns a high-level Outline object, based on the outline tree
// of the reader.
func (r *PdfReader) GetOutlines() (*Outline, error) {
	if r == nil {
		return nil, errors.New("cannot create outline from nil reader")
	}

	outlineTree := r.GetOutlineTree()
	if outlineTree == nil {
		return nil, errors.New("the specified reader does not have an outline tree")
	}

	var traverseFunc func(node *PdfOutlineTreeNode, entries *[]*OutlineItem)
	traverseFunc = func(node *PdfOutlineTreeNode, entries *[]*OutlineItem) {
		if node == nil {
			return
		}
		if node.context == nil {
			common.Log.Debug("ERROR: missing outline entry context")
			return
		}

		// Check if node is an outline item.
		var entry *OutlineItem
		if item, ok := node.context.(*PdfOutlineItem); ok {
			// Search for outline destination object.
			destObj := item.Dest
			if (destObj == nil || core.IsNullObject(destObj)) && item.A != nil {
				if actionDict, ok := core.GetDict(item.A); ok {
					destObj, _ = core.GetArray(actionDict.Get("D"))
				}
			}

			// Parse outline destination object.
			var dest OutlineDest
			if destObj != nil && !core.IsNullObject(destObj) {
				if d, err := newOutlineDestFromPdfObject(destObj, r); err == nil {
					dest = *d
				} else {
					common.Log.Debug("WARN: could not parse outline dest (%v): %v", destObj, err)
				}
			}

			entry = NewOutlineItem(item.Title.Decoded(), dest)
			*entries = append(*entries, entry)

			// Traverse next node.
			if item.Next != nil {
				traverseFunc(item.Next, entries)
			}
		}

		// Check if node has children.
		if node.First != nil {
			if entry != nil {
				entries = &entry.Entries
			}

			// Traverse node children.
			traverseFunc(node.First, entries)
		}
	}

	outline := NewOutline()
	traverseFunc(outlineTree, &outline.Entries)
	return outline, nil
}

// AcroFormRepairOptions contains options for rebuilding the AcroForm.
type AcroFormRepairOptions struct {
}

// RepairAcroForm attempts to rebuild the AcroForm fields using the widget
// annotations present in the document pages. Pass nil for the opts parameter
// in order to use the default options.
// NOTE: Currently, the opts parameter is declared in order to enable adding
// future options, but passing nil will always result in the default options
// being used.
func (r *PdfReader) RepairAcroForm(opts *AcroFormRepairOptions) error {
	var fields []*PdfField
	fieldCache := map[*core.PdfIndirectObject]struct{}{}
	for _, page := range r.PageList {
		annotations, err := page.GetAnnotations()
		if err != nil {
			return err
		}

		for _, annotation := range annotations {
			var field *PdfField
			switch t := annotation.GetContext().(type) {
			case *PdfAnnotationWidget:
				if t.parent != nil {
					field = t.parent
					break
				}
				if parentObj, ok := core.GetIndirect(t.Parent); ok {
					field, err = r.newPdfFieldFromIndirectObject(parentObj, nil)
					if err == nil {
						break
					}
					common.Log.Debug("WARN: could not parse form field %+v: %v", parentObj, err)
				}
				if t.container != nil {
					field, err = r.newPdfFieldFromIndirectObject(t.container, nil)
					if err == nil {
						break
					}
					common.Log.Debug("WARN: could not parse form field %+v: %v", t.container, err)
				}
			}
			if field == nil {
				continue
			}
			if _, ok := fieldCache[field.container]; ok {
				continue
			}
			fieldCache[field.container] = struct{}{}
			fields = append(fields, field)
		}
	}

	if len(fields) == 0 {
		return nil
	}
	if r.AcroForm == nil {
		r.AcroForm = NewPdfAcroForm()
	}
	r.AcroForm.Fields = &fields
	return nil
}

// AcroFormNeedsRepair returns true if the document contains widget annotations
// linked to fields which are not referenced in the AcroForm. The AcroForm can
// be repaired using the RepairAcroForm method of the reader.
func (r *PdfReader) AcroFormNeedsRepair() (bool, error) {
	var fields []*PdfField
	if r.AcroForm != nil {
		fields = r.AcroForm.AllFields()
	}

	fieldMap := make(map[*PdfField]struct{}, len(fields))
	for _, field := range fields {
		fieldMap[field] = struct{}{}
	}

	for _, page := range r.PageList {
		annotations, err := page.GetAnnotations()
		if err != nil {
			return false, err
		}

		for _, annotation := range annotations {
			widget, ok := annotation.GetContext().(*PdfAnnotationWidget)
			if !ok {
				continue
			}

			field := widget.Field()
			if field == nil {
				return true, nil
			}
			if _, ok := fieldMap[field]; !ok {
				return true, nil
			}
		}
	}

	return false, nil
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

	formsContainer, _ := core.GetIndirect(obj)

	obj = core.TraceToDirectObject(obj)
	if core.IsNullObject(obj) {
		common.Log.Trace("Acroform is a null object (empty)\n")
		return nil, nil
	}

	formsDict, ok := core.GetDict(obj)
	if !ok {
		common.Log.Debug("Invalid AcroForm entry %T", obj)
		common.Log.Debug("Does not have forms")
		return nil, fmt.Errorf("invalid acroform entry %T", obj)
	}
	common.Log.Trace("Has Acro forms")
	// Load it.

	// Ensure we have access to everything.
	common.Log.Trace("Traverse the Acroforms structure")
	if !r.isLazy {
		err := r.traverseObjectData(formsDict)
		if err != nil {
			common.Log.Debug("ERROR: Unable to traverse AcroForms (%s)", err)
			return nil, err
		}
	}

	// Create the acro forms object.
	acroForm, err := r.newPdfAcroFormFromDict(formsContainer, formsDict)
	if err != nil {
		return nil, err
	}

	return acroForm, nil
}

func (r *PdfReader) lookupPageByObject(obj core.PdfObject) (*PdfPage, error) {
	// can be indirect, direct, or reference
	// look up the corresponding page
	return nil, errors.New("page not found")
}

// Build the table of contents.
// tree, ex: Pages -> Pages -> Pages -> Page
// Traverse through the whole thing recursively.
func (r *PdfReader) buildPageList(node *core.PdfIndirectObject, parent *core.PdfIndirectObject, traversedPageNodes map[core.PdfObject]struct{}) error {
	if node == nil {
		return nil
	}

	if _, alreadyTraversed := traversedPageNodes[node]; alreadyTraversed {
		common.Log.Debug("Cyclic recursion, skipping (%v)", node.ObjectNumber)
		return nil
	}
	traversedPageNodes[node] = struct{}{}

	nodeDict, ok := node.PdfObject.(*core.PdfObjectDictionary)
	if !ok {
		return errors.New("node not a dictionary")
	}

	objType, ok := (*nodeDict).Get("Type").(*core.PdfObjectName)
	if !ok {
		if nodeDict.Get("Kids") == nil {
			return errors.New("node missing Type (Required)")
		}

		common.Log.Debug("ERROR: node missing Type, but has Kids. Assuming Pages node.")
		objType = core.MakeName("Pages")
		nodeDict.Set("Type", objType)
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
		return errors.New("table of content containing non Page/Pages object")
	}

	// A Pages object.  Update the parent.
	if parent != nil {
		nodeDict.Set("Parent", parent)
	}

	// Resolve the object recursively.
	if !r.isLazy {
		err := r.traverseObjectData(node)
		if err != nil {
			return err
		}
	}

	kidsObj, err := r.parser.Resolve(nodeDict.Get("Kids"))
	if err != nil {
		common.Log.Debug("ERROR: Failed loading Kids object")
		return err
	}

	var kids *core.PdfObjectArray
	kids, ok = kidsObj.(*core.PdfObjectArray)
	if !ok {
		kidsIndirect, isIndirect := kidsObj.(*core.PdfIndirectObject)
		if !isIndirect {
			return errors.New("invalid Kids object")
		}
		kids, ok = kidsIndirect.PdfObject.(*core.PdfObjectArray)
		if !ok {
			return errors.New("invalid Kids indirect object")
		}
	}
	common.Log.Trace("Kids: %s", kids)
	for idx, child := range kids.Elements() {
		child, ok := core.GetIndirect(child)
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
func (r *PdfReader) resolveReference(ref *core.PdfObjectReference) (core.PdfObject, bool, error) {
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
func (r *PdfReader) traverseObjectData(o core.PdfObject) error {
	return core.ResolveReferencesDeep(o, r.traversed)
}

// PageFromIndirectObject returns the PdfPage and page number for a given indirect object.
func (r *PdfReader) PageFromIndirectObject(ind *core.PdfIndirectObject) (*PdfPage, int, error) {
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
func (r *PdfReader) GetOCProperties() (core.PdfObject, error) {
	dict := r.catalog
	obj := dict.Get("OCProperties")
	obj = core.ResolveReference(obj)

	// Resolve all references...
	// Should be pretty safe. Should not be referencing to pages or
	// any large structures.  Local structures and references
	// to OC Groups.
	if !r.isLazy {
		err := r.traverseObjectData(obj)
		if err != nil {
			return nil, err
		}
	}

	return obj, nil
}

// GetNamedDestinations returns the Names entry in the PDF catalog.
// See section 12.3.2.3 "Named Destinations" (p. 367 PDF32000_2008).
func (r *PdfReader) GetNamedDestinations() (core.PdfObject, error) {
	obj := core.ResolveReference(r.catalog.Get("Names"))
	if obj == nil {
		return nil, nil
	}

	// Resolve references.
	if !r.isLazy {
		err := r.traverseObjectData(obj)
		if err != nil {
			return nil, err
		}
	}

	return obj, nil
}

// GetPageLabels returns the PageLabels entry in the PDF catalog.
// See section 12.4.2 "Page Labels" (p. 382 PDF32000_2008).
func (r *PdfReader) GetPageLabels() (core.PdfObject, error) {
	obj := core.ResolveReference(r.catalog.Get("PageLabels"))
	if obj == nil {
		return nil, nil
	}

	// Resolve references.
	if !r.isLazy {
		err := r.traverseObjectData(obj)
		if err != nil {
			return nil, err
		}
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
func (r *PdfReader) GetIndirectObjectByNumber(number int) (core.PdfObject, error) {
	obj, err := r.parser.LookupByNumber(number)
	return obj, err
}

// GetTrailer returns the PDF's trailer dictionary.
func (r *PdfReader) GetTrailer() (*core.PdfObjectDictionary, error) {
	trailerDict := r.parser.GetTrailer()
	if trailerDict == nil {
		return nil, errors.New("trailer missing")
	}

	return trailerDict, nil
}
