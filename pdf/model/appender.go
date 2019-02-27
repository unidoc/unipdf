/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

// PdfAppender appends new PDF content to an existing PDF document via incremental updates.
type PdfAppender struct {
	rs       io.ReadSeeker
	parser   *core.PdfParser
	roReader *PdfReader
	Reader   *PdfReader
	pages    []*PdfPage
	acroForm *PdfAcroForm

	xrefs          core.XrefTable
	greatestObjNum int

	// List of new objects and a map for quick lookups.
	newObjects   []core.PdfObject
	hasNewObject map[core.PdfObject]struct{}

	written bool
}

func getPageResources(p *PdfPage) map[core.PdfObjectName]core.PdfObject {
	resources := make(map[core.PdfObjectName]core.PdfObject)
	if p.Resources == nil {
		return resources
	}
	if p.Resources.Font != nil {
		if dict, found := core.GetDict(p.Resources.Font); found {
			for _, key := range dict.Keys() {
				resources[key] = dict.Get(key)
			}
		}
	}
	if p.Resources.ExtGState != nil {
		if dict, found := core.GetDict(p.Resources.ExtGState); found {
			for _, key := range dict.Keys() {
				resources[key] = dict.Get(key)
			}
		}
	}
	if p.Resources.XObject != nil {
		if dict, found := core.GetDict(p.Resources.XObject); found {
			for _, key := range dict.Keys() {
				resources[key] = dict.Get(key)
			}
		}
	}
	if p.Resources.Pattern != nil {
		if dict, found := core.GetDict(p.Resources.Pattern); found {
			for _, key := range dict.Keys() {
				resources[key] = dict.Get(key)
			}
		}
	}
	if p.Resources.Shading != nil {
		if dict, found := core.GetDict(p.Resources.Shading); found {
			for _, key := range dict.Keys() {
				resources[key] = dict.Get(key)
			}
		}
	}
	if p.Resources.ProcSet != nil {
		if dict, found := core.GetDict(p.Resources.ProcSet); found {
			for _, key := range dict.Keys() {
				resources[key] = dict.Get(key)
			}
		}
	}
	if p.Resources.Properties != nil {
		if dict, found := core.GetDict(p.Resources.Properties); found {
			for _, key := range dict.Keys() {
				resources[key] = dict.Get(key)
			}
		}
	}

	return resources
}

// NewPdfAppender creates a new Pdf appender from a Pdf reader.
func NewPdfAppender(reader *PdfReader) (*PdfAppender, error) {
	a := &PdfAppender{
		rs:     reader.rs,
		Reader: reader,
		parser: reader.parser,
	}
	if _, err := a.rs.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	var err error

	// Create a readonly (immutable) reader. It increases memory use but is necessary to be able
	// to detect changes in the original reader objects.
	//
	// In the case where an existing page is modified, the page contents are replaced upon merging
	// (appending). The new page will refer to objects from the read-only reader and new instances
	// of objects that have been changes. Objects from the original reader are not appended, only
	// new objects that modify the PDF. The change detection check is not resource demanding. It
	// only checks owners (source) of indirect objects.
	a.roReader, err = NewPdfReader(a.rs)
	if err != nil {
		return nil, err
	}
	for _, idx := range a.Reader.GetObjectNums() {
		if a.greatestObjNum < idx {
			a.greatestObjNum = idx
		}
	}
	a.xrefs = a.parser.GetXrefTable()
	a.hasNewObject = make(map[core.PdfObject]struct{})
	for _, p := range a.roReader.PageList {
		a.pages = append(a.pages, p)
	}

	// Load interactive forms and fields.
	a.roReader.AcroForm, err = a.roReader.loadForms()
	if err != nil {
		return nil, err
	}
	a.acroForm = a.roReader.AcroForm

	return a, nil
}

func (a *PdfAppender) addNewObjects(obj core.PdfObject) {
	if _, ok := a.hasNewObject[obj]; ok || obj == nil {
		return
	}
	switch v := obj.(type) {
	case *core.PdfIndirectObject:
		// If the current parser is different from the read-only parser, then
		// the object has changed.
		if v.GetParser() != a.roReader.parser {
			a.newObjects = append(a.newObjects, obj)
			a.hasNewObject[obj] = struct{}{}
			a.addNewObjects(v.PdfObject)
		}
	case *core.PdfObjectArray:
		for _, o := range v.Elements() {
			a.addNewObjects(o)
		}
	case *core.PdfObjectDictionary:
		for _, key := range v.Keys() {
			a.addNewObjects(v.Get(key))
		}
	case *core.PdfObjectStreams:
		// If the current parser is different from the read-only parser, then
		// the object has changed.
		if v.GetParser() != a.roReader.parser {
			for _, o := range v.Elements() {
				a.addNewObjects(o)
			}
		}
	case *core.PdfObjectStream:
		// If the current parser is different from the read-only parser, then
		// the object has changed.
		parser := v.GetParser()
		if parser == a.roReader.parser {
			return
		}

		// If the current parser is different from the parser of the reader,
		// then the object may have changed.
		if parser == a.Reader.parser {
			// Check if data has changed.
			if streamObj, err := a.roReader.parser.LookupByReference(v.PdfObjectReference); err == nil {
				var isNotChanged bool
				if stream, ok := core.GetStream(streamObj); ok && bytes.Equal(stream.Stream, v.Stream) {
					isNotChanged = true
				}
				if dict, ok := core.GetDict(streamObj); isNotChanged && ok {
					isNotChanged = dict.WriteString() == v.PdfObjectDictionary.WriteString()
				}
				if isNotChanged {
					return
				}
			}
		}
		a.newObjects = append(a.newObjects, obj)
		a.hasNewObject[obj] = struct{}{}
		a.addNewObjects(v.PdfObjectDictionary)
	}
}

// mergeResources adds new named resources from src to dest. If the resources have the same name its will be renamed.
// The dest and src are resources dictionary. resourcesRenameMap is a rename map for resources.
func (a *PdfAppender) mergeResources(dest, src core.PdfObject, resourcesRenameMap map[core.PdfObjectName]core.PdfObjectName) core.PdfObject {
	if src == nil && dest == nil {
		return nil
	}
	if src == nil {
		return dest
	}

	srcDict, ok := core.GetDict(src)
	if !ok {
		return dest
	}
	if dest == nil {
		dict := core.MakeDict()
		dict.Merge(srcDict)
		return src
	}

	destDict, ok := core.GetDict(dest)
	if !ok {
		common.Log.Error("Error resource is not a dictionary")
		destDict = core.MakeDict()
	}

	for _, key := range srcDict.Keys() {
		if newKey, found := resourcesRenameMap[key]; found {
			destDict.Set(newKey, srcDict.Get(key))
		} else {
			destDict.Set(key, srcDict.Get(key))
		}
	}
	return destDict
}

// MergePageWith appends page content to source Pdf file page content.
func (a *PdfAppender) MergePageWith(pageNum int, page *PdfPage) error {
	pageIndex := pageNum - 1
	var srcPage *PdfPage
	for i, p := range a.pages {
		if i == pageIndex {
			srcPage = p
		}
	}
	if srcPage == nil {
		return fmt.Errorf("ERROR: Page dictionary %d not found in the source document", pageNum)
	}
	if srcPage.primitive != nil && srcPage.primitive.GetParser() == a.roReader.parser {
		srcPage = srcPage.Duplicate()
		a.pages[pageIndex] = srcPage
	}

	page = page.Duplicate()
	procPage(page)

	srcResources := getPageResources(srcPage)
	pageResources := getPageResources(page)
	resourcesRenameMap := make(map[core.PdfObjectName]core.PdfObjectName)

	for key := range pageResources {
		if _, found := srcResources[key]; found {
			for i := 1; true; i++ {
				newKey := core.PdfObjectName(string(key) + strconv.Itoa(i))
				if _, exists := srcResources[newKey]; !exists {
					resourcesRenameMap[key] = newKey
					break
				}
			}
		}
	}

	contentStreams, err := page.GetContentStreams()
	if err != nil {
		return err
	}

	srcContentStreams, err := srcPage.GetContentStreams()
	if err != nil {
		return err
	}

	for i, stream := range contentStreams {
		for oldName, newName := range resourcesRenameMap {
			stream = strings.Replace(stream, "/"+string(oldName), "/"+string(newName), -1)
		}
		contentStreams[i] = stream
	}

	srcContentStreams = append(srcContentStreams, contentStreams...)

	if err := srcPage.SetContentStreams(srcContentStreams, core.NewFlateEncoder()); err != nil {
		return err
	}

	for _, a := range page.Annotations {
		srcPage.Annotations = append(srcPage.Annotations, a)
	}

	if srcPage.Resources == nil {
		srcPage.Resources = NewPdfPageResources()
	}

	if page.Resources != nil {
		srcPage.Resources.Font = a.mergeResources(srcPage.Resources.Font, page.Resources.Font, resourcesRenameMap)
		srcPage.Resources.XObject = a.mergeResources(srcPage.Resources.XObject, page.Resources.XObject, resourcesRenameMap)
		srcPage.Resources.Properties = a.mergeResources(srcPage.Resources.Properties, page.Resources.Properties, resourcesRenameMap)
		if srcPage.Resources.ProcSet == nil {
			srcPage.Resources.ProcSet = page.Resources.ProcSet
		}
		srcPage.Resources.Shading = a.mergeResources(srcPage.Resources.Shading, page.Resources.Shading, resourcesRenameMap)
		srcPage.Resources.ExtGState = a.mergeResources(srcPage.Resources.ExtGState, page.Resources.ExtGState, resourcesRenameMap)
	}

	srcMediaBox, err := srcPage.GetMediaBox()
	if err != nil {
		return err
	}

	pageMediaBox, err := page.GetMediaBox()
	if err != nil {
		return err
	}
	var mediaBoxChanged bool

	if srcMediaBox.Llx > pageMediaBox.Llx {
		srcMediaBox.Llx = pageMediaBox.Llx
		mediaBoxChanged = true
	}
	if srcMediaBox.Lly > pageMediaBox.Lly {
		srcMediaBox.Lly = pageMediaBox.Lly
		mediaBoxChanged = true
	}
	if srcMediaBox.Urx < pageMediaBox.Urx {
		srcMediaBox.Urx = pageMediaBox.Urx
		mediaBoxChanged = true
	}
	if srcMediaBox.Ury < pageMediaBox.Ury {
		srcMediaBox.Ury = pageMediaBox.Ury
		mediaBoxChanged = true
	}

	if mediaBoxChanged {
		srcPage.MediaBox = srcMediaBox
	}

	return nil
}

// AddPages adds pages to be appended to the end of the source PDF.
func (a *PdfAppender) AddPages(pages ...*PdfPage) {
	for _, page := range pages {
		page = page.Duplicate()
		procPage(page)
		a.pages = append(a.pages, page)
	}
	return
}

// RemovePage removes a page by number.
func (a *PdfAppender) RemovePage(pageNum int) {
	pageIndex := pageNum - 1
	pages := make([]*PdfPage, 0, len(a.pages))
	for i, p := range a.pages {
		if i == pageIndex {
			continue
		}
		if p.primitive != nil && p.primitive.GetParser() == a.roReader.parser {
			p = p.Duplicate()
			procPage(p)
		}
		pages = append(pages, p)
	}
	a.pages = pages
}

// ReplacePage replaces the original page to a new page.
func (a *PdfAppender) ReplacePage(pageNum int, page *PdfPage) {
	pageIndex := pageNum - 1
	for i := range a.pages {
		if i == pageIndex {
			p := page.Duplicate()
			procPage(p)
			a.pages[i] = p
		}
	}
}

// Sign signs a specific page with a digital signature.
// The signature field parameter must have a valid signature dictionary
// specified by its V field.
func (a *PdfAppender) Sign(pageNum int, field *PdfFieldSignature) error {
	if field == nil {
		return errors.New("signature field cannot be nil")
	}

	signature := field.V
	if signature == nil {
		return errors.New("signature dictionary cannot be nil")
	}
	a.addNewObjects(signature.container)

	// Get a copy of the selected page.
	pageIndex := pageNum - 1
	if pageIndex < 0 || pageIndex > len(a.pages)-1 {
		return fmt.Errorf("page %d not found", pageNum)
	}
	page := a.pages[pageIndex].Duplicate()

	// Add signature field annotations to the page annotations.
	field.P = page.ToPdfObject()
	if field.T == nil || field.T.String() == "" {
		field.T = core.MakeString(fmt.Sprintf("Signature %d", pageNum))
	}

	page.Annotations = append(page.Annotations, field.PdfAnnotationWidget.PdfAnnotation)

	// Add signature field to the form.
	acroForm := a.acroForm
	if acroForm == nil {
		acroForm = a.Reader.AcroForm
		if acroForm == nil {
			acroForm = NewPdfAcroForm()
		}
	}
	acroForm.SigFlags = core.MakeInteger(3)

	fields := append(acroForm.AllFields(), field.PdfField)
	acroForm.Fields = &fields
	a.ReplaceAcroForm(acroForm)

	// Replace original page.
	procPage(page)
	a.pages[pageIndex] = page

	return nil
}

// ReplaceAcroForm replaces the acrobat form. It appends a new form to the Pdf which
// replaces the original AcroForm.
func (a *PdfAppender) ReplaceAcroForm(acroForm *PdfAcroForm) {
	a.acroForm = acroForm
}

// Write writes the Appender output to io.Writer.
// It can only be called once and further invocations will result in an error.
func (a *PdfAppender) Write(w io.Writer) error {
	if a.written {
		return errors.New("appender write can only be invoked once")
	}

	writer := NewPdfWriter()

	pagesDict, ok := core.GetDict(writer.pages)
	if !ok {
		return errors.New("Invalid Pages obj (not a dict)")
	}
	kids, ok := pagesDict.Get("Kids").(*core.PdfObjectArray)
	if !ok {
		return errors.New("Invalid Pages Kids obj (not an array)")
	}
	pageCount, ok := pagesDict.Get("Count").(*core.PdfObjectInteger)
	if !ok {
		return errors.New("Invalid Pages Count object (not an integer)")
	}

	parser := a.roReader.parser
	trailer := parser.GetTrailer()
	if trailer == nil {
		return fmt.Errorf("Missing trailer")
	}
	// Catalog.
	root, ok := trailer.Get("Root").(*core.PdfObjectReference)
	if !ok {
		return fmt.Errorf("Invalid Root (trailer: %s)", trailer)
	}

	oc, err := parser.LookupByReference(*root)
	if err != nil {
		return err
	}
	catalog, ok := core.GetDict(oc)
	if !ok {
		common.Log.Debug("ERROR: Missing catalog: (root %q) (trailer %s)", oc, *trailer)
		return errors.New("Missing catalog")
	}

	for _, key := range catalog.Keys() {
		if writer.catalog.Get(key) == nil {
			obj := catalog.Get(key)
			writer.catalog.Set(key, obj)
		}
	}

	inheritedFields := []core.PdfObjectName{"Resources", "MediaBox", "CropBox", "Rotate"}

	for _, p := range a.pages {
		// Update the count.
		obj := p.ToPdfObject()
		*pageCount = *pageCount + 1
		// Check the object is not changing.
		// If the indirect object has the parser which equals to the readonly then the object is not changed.
		if ind, ok := obj.(*core.PdfIndirectObject); ok && ind.GetParser() == a.roReader.parser {
			kids.Append(&ind.PdfObjectReference)
			continue
		}
		if pDict, ok := core.GetDict(obj); ok {
			parent, hasParent := pDict.Get("Parent").(*core.PdfIndirectObject)
			for hasParent {
				common.Log.Trace("Page Parent: %T", parent)
				parentDict, ok := parent.PdfObject.(*core.PdfObjectDictionary)
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
				parent, hasParent = parentDict.Get("Parent").(*core.PdfIndirectObject)
				common.Log.Trace("Next parent: %T", parentDict.Get("Parent"))
			}
			pDict.Set("Parent", writer.pages)
		}
		a.addNewObjects(obj)
		kids.Append(obj)
	}
	if a.acroForm != nil {
		writer.SetForms(a.acroForm)
	}

	if _, err := a.rs.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Digital signature handling: Check if any of the new objects represent a signature dictionary.
	// The byte range is later updated dynamically based on the position of the actual signature
	// Contents.
	digestWriters := make(map[SignatureHandler]io.Writer)
	byteRange := core.MakeArray()
	for _, obj := range a.newObjects {
		if ind, found := core.GetIndirect(obj); found {
			if sigDict, found := ind.PdfObject.(*pdfSignDictionary); found {
				handler := *sigDict.handler
				var err error
				digestWriters[handler], err = handler.NewDigest(sigDict.signature)
				if err != nil {
					return err
				}
				byteRange.Append(core.MakeInteger(0xfffff), core.MakeInteger(0xfffff))
			}
		}
	}
	if byteRange.Len() > 0 {
		byteRange.Append(core.MakeInteger(0xfffff), core.MakeInteger(0xfffff))
	}
	for _, obj := range a.newObjects {
		if ind, found := core.GetIndirect(obj); found {
			if sigDict, found := ind.PdfObject.(*pdfSignDictionary); found {
				sigDict.Set("ByteRange", byteRange)
			}
		}
	}

	hasSigDict := len(digestWriters) > 0

	var reader io.Reader = a.rs
	if hasSigDict {
		writers := make([]io.Writer, 0, len(digestWriters))
		for _, hash := range digestWriters {
			writers = append(writers, hash)
		}
		reader = io.TeeReader(a.rs, io.MultiWriter(writers...))
	}

	// Write the original PDF.
	offset, err := io.Copy(w, reader)
	if err != nil {
		return err
	}

	if len(a.newObjects) == 0 {
		return nil
	}

	writer.writeOffset = offset
	writer.ObjNumOffset = a.greatestObjNum
	writer.appendMode = true
	writer.appendToXrefs = a.xrefs
	writer.minorVersion = 7

	for _, obj := range a.newObjects {
		writer.addObject(obj)
	}

	writerW := w
	if hasSigDict {
		// For signatures, we need to write twice. First to find the byte offset
		// of the Contents and then dynamically update the file with the
		// signature and ByteRange.
		writerW = bytes.NewBuffer(nil)
	}

	// Perform the write. For signatures will do a mock write to a buffer.
	if err := writer.Write(writerW); err != nil {
		return err
	}

	// TODO(gunnsth): Consider whether the dynamic content can be handled efficiently with generic write hooks?
	// Logic is getting pretty complex here.
	if hasSigDict {
		// Update the byteRanges based on mock write.
		bufferData := writerW.(*bytes.Buffer).Bytes()
		byteRange := core.MakeArray()
		var sigDicts []*pdfSignDictionary
		var lastPosition int64
		for _, obj := range writer.objects {
			if ind, found := core.GetIndirect(obj); found {
				if sigDict, found := ind.PdfObject.(*pdfSignDictionary); found {
					sigDicts = append(sigDicts, sigDict)
					newPosition := sigDict.fileOffset + int64(sigDict.contentsOffsetStart)
					byteRange.Append(
						core.MakeInteger(lastPosition),
						core.MakeInteger(newPosition-lastPosition),
					)
					lastPosition = sigDict.fileOffset + int64(sigDict.contentsOffsetEnd)
				}
			}
		}
		byteRange.Append(
			core.MakeInteger(lastPosition),
			core.MakeInteger(offset+int64(len(bufferData))-lastPosition),
		)
		// set the ByteRange value
		byteRangeData := []byte(byteRange.WriteString())
		for _, sigDict := range sigDicts {
			bufferOffset := int(sigDict.fileOffset - offset)
			for i := sigDict.byteRangeOffsetStart; i < sigDict.byteRangeOffsetEnd; i++ {
				bufferData[bufferOffset+i] = ' '
			}
			dst := bufferData[bufferOffset+sigDict.byteRangeOffsetStart : bufferOffset+sigDict.byteRangeOffsetEnd]
			copy(dst, byteRangeData)
		}
		var prevOffset int
		for _, sigDict := range sigDicts {
			bufferOffset := int(sigDict.fileOffset - offset)
			data := bufferData[prevOffset : bufferOffset+sigDict.contentsOffsetStart]
			handler := *sigDict.handler
			digestWriters[handler].Write(data)
			prevOffset = bufferOffset + sigDict.contentsOffsetEnd
		}
		for _, sigDict := range sigDicts {
			data := bufferData[prevOffset:]
			handler := *sigDict.handler
			digestWriters[handler].Write(data)
		}
		for _, sigDict := range sigDicts {
			bufferOffset := int(sigDict.fileOffset - offset)
			handler := *sigDict.handler
			digest := digestWriters[handler]
			if err := handler.Sign(sigDict.signature, digest); err != nil {
				return err
			}
			sigDict.signature.ByteRange = byteRange
			contents := []byte(sigDict.signature.Contents.WriteString())

			// Empty out the ByteRange and Content data.
			// FIXME(gunnsth): Is this needed?  Seems like the correct data is copied below?  Prefer
			// to keep the rest space?
			for i := sigDict.byteRangeOffsetStart; i < sigDict.byteRangeOffsetEnd; i++ {
				bufferData[bufferOffset+i] = ' '
			}
			for i := sigDict.contentsOffsetStart; i < sigDict.contentsOffsetEnd; i++ {
				bufferData[bufferOffset+i] = ' '
			}

			// Copy the actual ByteRange and Contents data into the buffer prepared by first write.
			dst := bufferData[bufferOffset+sigDict.byteRangeOffsetStart : bufferOffset+sigDict.byteRangeOffsetEnd]
			copy(dst, byteRangeData)
			dst = bufferData[bufferOffset+sigDict.contentsOffsetStart : bufferOffset+sigDict.contentsOffsetEnd]
			copy(dst, contents)
		}

		buffer := bytes.NewBuffer(bufferData)
		_, err = io.Copy(w, buffer)
		if err != nil {
			return err
		}
	}

	a.written = true
	return nil
}

// WriteToFile writes the Appender output to file specified by path.
func (a *PdfAppender) WriteToFile(outputPath string) error {
	fWrite, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer fWrite.Close()
	return a.Write(fWrite)
}
