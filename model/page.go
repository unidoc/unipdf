/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

//
// Allow higher level manipulation of PDF files and pages.
// This can be continuously expanded to support more and more features.
// Generic handling can be done by defining elements as PdfObject which
// can later be replaced and fully defined.
//

package model

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// PdfPage represents a page in a PDF document. (7.7.3.3 - Table 30).
type PdfPage struct {
	Parent               core.PdfObject
	LastModified         *PdfDate
	Resources            *PdfPageResources
	CropBox              *PdfRectangle
	MediaBox             *PdfRectangle
	BleedBox             *PdfRectangle
	TrimBox              *PdfRectangle
	ArtBox               *PdfRectangle
	BoxColorInfo         core.PdfObject
	Contents             core.PdfObject
	Rotate               *int64
	Group                core.PdfObject
	Thumb                core.PdfObject
	B                    core.PdfObject
	Dur                  core.PdfObject
	Trans                core.PdfObject
	AA                   core.PdfObject
	Metadata             core.PdfObject
	PieceInfo            core.PdfObject
	StructParents        core.PdfObject
	ID                   core.PdfObject
	PZ                   core.PdfObject
	SeparationInfo       core.PdfObject
	Tabs                 core.PdfObject
	TemplateInstantiated core.PdfObject
	PresSteps            core.PdfObject
	UserUnit             core.PdfObject
	VP                   core.PdfObject
	Annots               core.PdfObject

	annotations []*PdfAnnotation

	// Primitive container.
	pageDict  *core.PdfObjectDictionary
	primitive *core.PdfIndirectObject

	reader *PdfReader
}

// NewPdfPage returns a new PDF page.
func NewPdfPage() *PdfPage {
	page := PdfPage{}
	page.pageDict = core.MakeDict()

	page.Resources = NewPdfPageResources()

	container := core.PdfIndirectObject{}
	container.PdfObject = page.pageDict
	page.primitive = &container

	return &page
}

func (p *PdfPage) setContainer(container *core.PdfIndirectObject) {
	container.PdfObject = p.pageDict
	p.primitive = container
}

// Duplicate creates a duplicate page based on the current one and returns it.
func (p *PdfPage) Duplicate() *PdfPage {
	var dup PdfPage
	dup = *p
	dup.pageDict = core.MakeDict()
	dup.primitive = core.MakeIndirectObject(dup.pageDict)

	return &dup
}

// Build a PdfPage based on the underlying dictionary.
// Used in loading existing PDF files.
// Note that a new container is created (indirect object).
func (r *PdfReader) newPdfPageFromDict(p *core.PdfObjectDictionary) (*PdfPage, error) {
	page := NewPdfPage()
	page.pageDict = p

	d := *p

	pType, ok := d.Get("Type").(*core.PdfObjectName)
	if !ok {
		return nil, errors.New("missing/invalid Page dictionary Type")
	}
	if *pType != "Page" {
		return nil, errors.New("page dictionary Type != Page")
	}

	if obj := d.Get("Parent"); obj != nil {
		page.Parent = obj
	}

	if obj := d.Get("LastModified"); obj != nil {
		strObj, ok := core.GetString(obj)
		if !ok {
			return nil, errors.New("page dictionary LastModified != string")
		}
		lastmod, err := NewPdfDate(strObj.Str())
		if err != nil {
			return nil, err
		}
		page.LastModified = &lastmod
	}

	if obj := d.Get("Resources"); obj != nil && !core.IsNullObject(obj) {
		dict, ok := core.GetDict(obj)
		if !ok {
			return nil, fmt.Errorf("invalid resource dictionary (%T)", obj)
		}

		var err error
		page.Resources, err = NewPdfPageResourcesFromDict(dict)
		if err != nil {
			return nil, err
		}
	} else {
		// If Resources not explicitly defined, look up the tree (Parent objects) using
		// the getParentResources() function. Resources should always be accessible.
		resources, err := page.getParentResources()
		if err != nil {
			return nil, err
		}
		if resources == nil {
			resources = NewPdfPageResources()
		}
		page.Resources = resources
	}

	if obj := d.Get("MediaBox"); obj != nil {
		boxArr, ok := core.GetArray(obj)
		if !ok {
			return nil, errors.New("page MediaBox not an array")
		}
		var err error
		page.MediaBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}

	if obj := d.Get("CropBox"); obj != nil {
		boxArr, ok := core.GetArray(obj)
		if !ok {
			return nil, errors.New("page CropBox not an array")
		}
		var err error
		page.CropBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("BleedBox"); obj != nil {
		boxArr, ok := core.GetArray(obj)
		if !ok {
			return nil, errors.New("page BleedBox not an array")
		}
		var err error
		page.BleedBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("TrimBox"); obj != nil {
		boxArr, ok := core.GetArray(obj)
		if !ok {
			return nil, errors.New("page TrimBox not an array")
		}
		var err error
		page.TrimBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("ArtBox"); obj != nil {
		boxArr, ok := core.GetArray(obj)
		if !ok {
			return nil, errors.New("page ArtBox not an array")
		}
		var err error
		page.ArtBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("BoxColorInfo"); obj != nil {
		page.BoxColorInfo = obj
	}
	if obj := d.Get("Contents"); obj != nil {
		page.Contents = obj
	}
	if obj := d.Get("Rotate"); obj != nil {
		iObj, ok := core.GetInt(obj)
		if !ok {
			return nil, errors.New("invalid Page Rotate object")
		}
		iVal := int64(*iObj)
		page.Rotate = &iVal
	}
	if obj := d.Get("Group"); obj != nil {
		page.Group = obj
	}
	if obj := d.Get("Thumb"); obj != nil {
		page.Thumb = obj
	}
	if obj := d.Get("B"); obj != nil {
		page.B = obj
	}
	if obj := d.Get("Dur"); obj != nil {
		page.Dur = obj
	}
	if obj := d.Get("Trans"); obj != nil {
		page.Trans = obj
	}
	if obj := d.Get("AA"); obj != nil {
		page.AA = obj
	}
	if obj := d.Get("Metadata"); obj != nil {
		page.Metadata = obj
	}
	if obj := d.Get("PieceInfo"); obj != nil {
		page.PieceInfo = obj
	}
	if obj := d.Get("StructParents"); obj != nil {
		page.StructParents = obj
	}
	if obj := d.Get("ID"); obj != nil {
		page.ID = obj
	}
	if obj := d.Get("PZ"); obj != nil {
		page.PZ = obj
	}
	if obj := d.Get("SeparationInfo"); obj != nil {
		page.SeparationInfo = obj
	}
	if obj := d.Get("Tabs"); obj != nil {
		page.Tabs = obj
	}
	if obj := d.Get("TemplateInstantiated"); obj != nil {
		page.TemplateInstantiated = obj
	}
	if obj := d.Get("PresSteps"); obj != nil {
		page.PresSteps = obj
	}
	if obj := d.Get("UserUnit"); obj != nil {
		page.UserUnit = obj
	}
	if obj := d.Get("VP"); obj != nil {
		page.VP = obj
	}
	if obj := d.Get("Annots"); obj != nil {
		page.Annots = obj
	}

	page.reader = r

	return page, nil
}

// GetAnnotations returns the list of page annotations for `page`. If not loaded attempts to load the
// annotations, otherwise returns the loaded list.
func (p *PdfPage) GetAnnotations() ([]*PdfAnnotation, error) {
	if p.annotations != nil {
		return p.annotations, nil
	}
	if p.Annots == nil {
		p.annotations = []*PdfAnnotation{}
		return nil, nil
	}
	if p.reader == nil {
		p.annotations = []*PdfAnnotation{}
		return nil, nil
	}

	annots, err := p.reader.loadAnnotations(p.Annots)
	if err != nil {
		return nil, err
	}
	if annots == nil {
		p.annotations = []*PdfAnnotation{}
	}

	p.annotations = annots
	return p.annotations, nil
}

// AddAnnotation appends `annot` to the list of page annotations.
func (p *PdfPage) AddAnnotation(annot *PdfAnnotation) {
	if p.annotations == nil {
		p.GetAnnotations() // Ensure has been loaded.
	}
	p.annotations = append(p.annotations, annot)
}

// SetAnnotations sets the annotations list.
func (p *PdfPage) SetAnnotations(annotations []*PdfAnnotation) {
	p.annotations = annotations
}

// loadAnnotations loads and returns the PDF annotations from the input annotations object (array).
func (r *PdfReader) loadAnnotations(annotsObj core.PdfObject) ([]*PdfAnnotation, error) {
	annotsArr, ok := core.GetArray(annotsObj)
	if !ok {
		return nil, fmt.Errorf("Annots not an array")
	}

	var annotations []*PdfAnnotation
	for _, obj := range annotsArr.Elements() {
		obj = core.ResolveReference(obj)

		// Technically all annotation dictionaries should be inside indirect objects.
		// In reality, sometimes the annotation dictionary is inline within the Annots array.
		if _, isNull := obj.(*core.PdfObjectNull); isNull {
			// Can safely ignore.
			continue
		}

		annotDict, isDict := obj.(*core.PdfObjectDictionary)
		indirectObj, isIndirect := obj.(*core.PdfIndirectObject)
		if isDict {
			// Create a container; indirect object; around the dictionary.
			indirectObj = &core.PdfIndirectObject{}
			indirectObj.PdfObject = annotDict
		} else {
			if !isIndirect {
				return nil, fmt.Errorf("annotation not in an indirect object")
			}
		}

		annot, err := r.newPdfAnnotationFromIndirectObject(indirectObj)
		if err != nil {
			return nil, err
		}
		switch t := annot.GetContext().(type) {
		case *PdfAnnotationWidget:
			// Link widget annotation with form field (parent).
			for _, field := range r.AcroForm.AllFields() {
				if field.container == t.Parent {
					t.parent = field
					break
				}
			}
		}
		if annot != nil {
			annotations = append(annotations, annot)
		}
	}

	return annotations, nil
}

// GetMediaBox gets the inheritable media box value, either from the page
// or a higher up page/pages struct.
func (p *PdfPage) GetMediaBox() (*PdfRectangle, error) {
	if p.MediaBox != nil {
		return p.MediaBox, nil
	}

	node := p.Parent
	for node != nil {
		dict, ok := core.GetDict(node)
		if !ok {
			return nil, errors.New("invalid parent objects dictionary")
		}

		if obj := dict.Get("MediaBox"); obj != nil {
			arr, ok := core.GetArray(obj)
			if !ok {
				return nil, errors.New("invalid media box")
			}
			rect, err := NewPdfRectangle(*arr)

			if err != nil {
				return nil, err
			}

			return rect, nil
		}

		node = dict.Get("Parent")
	}

	return nil, errors.New("media box not defined")
}

// getParentResources searches for page resources in the parent nodes of the page.
func (p *PdfPage) getParentResources() (*PdfPageResources, error) {
	node := p.Parent
	for node != nil {
		dict, ok := core.GetDict(node)
		if !ok {
			common.Log.Debug("ERROR: invalid parent node")
			return nil, errors.New("invalid parent object")
		}

		if obj := dict.Get("Resources"); obj != nil {
			prDict, ok := core.GetDict(obj)
			if !ok {
				return nil, errors.New("invalid resource dict")
			}
			resources, err := NewPdfPageResourcesFromDict(prDict)

			if err != nil {
				return nil, err
			}

			return resources, nil
		}

		// Keep moving up the tree...
		node = dict.Get("Parent")
	}

	// No resources defined...
	return nil, nil
}

// GetPageDict converts the Page to a PDF object dictionary.
func (p *PdfPage) GetPageDict() *core.PdfObjectDictionary {
	d := p.pageDict
	d.Clear()
	d.Set("Type", core.MakeName("Page"))
	d.Set("Parent", p.Parent)

	if p.LastModified != nil {
		d.Set("LastModified", p.LastModified.ToPdfObject())
	}
	if p.Resources != nil {
		d.Set("Resources", p.Resources.ToPdfObject())
	}
	if p.CropBox != nil {
		d.Set("CropBox", p.CropBox.ToPdfObject())
	}
	if p.MediaBox != nil {
		d.Set("MediaBox", p.MediaBox.ToPdfObject())
	}
	if p.BleedBox != nil {
		d.Set("BleedBox", p.BleedBox.ToPdfObject())
	}
	if p.TrimBox != nil {
		d.Set("TrimBox", p.TrimBox.ToPdfObject())
	}
	if p.ArtBox != nil {
		d.Set("ArtBox", p.ArtBox.ToPdfObject())
	}
	d.SetIfNotNil("BoxColorInfo", p.BoxColorInfo)
	d.SetIfNotNil("Contents", p.Contents)

	if p.Rotate != nil {
		d.Set("Rotate", core.MakeInteger(*p.Rotate))
	}

	d.SetIfNotNil("Group", p.Group)
	d.SetIfNotNil("Thumb", p.Thumb)
	d.SetIfNotNil("B", p.B)
	d.SetIfNotNil("Dur", p.Dur)
	d.SetIfNotNil("Trans", p.Trans)
	d.SetIfNotNil("AA", p.AA)
	d.SetIfNotNil("Metadata", p.Metadata)
	d.SetIfNotNil("PieceInfo", p.PieceInfo)
	d.SetIfNotNil("StructParents", p.StructParents)
	d.SetIfNotNil("ID", p.ID)
	d.SetIfNotNil("PZ", p.PZ)
	d.SetIfNotNil("SeparationInfo", p.SeparationInfo)
	d.SetIfNotNil("Tabs", p.Tabs)
	d.SetIfNotNil("TemplateInstantiated", p.TemplateInstantiated)
	d.SetIfNotNil("PresSteps", p.PresSteps)
	d.SetIfNotNil("UserUnit", p.UserUnit)
	d.SetIfNotNil("VP", p.VP)

	if p.annotations != nil {
		arr := core.MakeArray()
		for _, annot := range p.annotations {
			if subannot := annot.GetContext(); subannot != nil {
				arr.Append(subannot.ToPdfObject())
			} else {
				// Generic annotation dict (without subtype).
				arr.Append(annot.ToPdfObject())
			}
		}

		if arr.Len() > 0 {
			d.Set("Annots", arr)
		}
	} else if p.Annots != nil {
		d.SetIfNotNil("Annots", p.Annots)
	}

	return d
}

// GetPageAsIndirectObject returns the page as a dictionary within an PdfIndirectObject.
func (p *PdfPage) GetPageAsIndirectObject() *core.PdfIndirectObject {
	return p.primitive
}

// GetContainingPdfObject returns the page as a dictionary within an PdfIndirectObject.
func (p *PdfPage) GetContainingPdfObject() core.PdfObject {
	return p.primitive
}

// ToPdfObject converts the PdfPage to a dictionary within an indirect object container.
func (p *PdfPage) ToPdfObject() core.PdfObject {
	container := p.primitive
	p.GetPageDict() // update.
	return container
}

// AddImageResource adds an image to the XObject resources.
func (p *PdfPage) AddImageResource(name core.PdfObjectName, ximg *XObjectImage) error {
	var xresDict *core.PdfObjectDictionary
	if p.Resources.XObject == nil {
		xresDict = core.MakeDict()
		p.Resources.XObject = xresDict
	} else {
		var ok bool
		xresDict, ok = (p.Resources.XObject).(*core.PdfObjectDictionary)
		if !ok {
			return errors.New("invalid xres dict type")
		}

	}
	// Make a stream object container.
	xresDict.Set(name, ximg.ToPdfObject())

	return nil
}

// HasXObjectByName checks if has XObject resource by name.
func (p *PdfPage) HasXObjectByName(name core.PdfObjectName) bool {
	xresDict, has := p.Resources.XObject.(*core.PdfObjectDictionary)
	if !has {
		return false
	}
	if obj := xresDict.Get(name); obj != nil {
		return true
	}

	return false
}

// GetXObjectByName gets XObject by name.
func (p *PdfPage) GetXObjectByName(name core.PdfObjectName) (core.PdfObject, bool) {
	xresDict, has := p.Resources.XObject.(*core.PdfObjectDictionary)
	if !has {
		return nil, false
	}
	if obj := xresDict.Get(name); obj != nil {
		return obj, true
	}

	return nil, false
}

// HasFontByName checks if has font resource by name.
func (p *PdfPage) HasFontByName(name core.PdfObjectName) bool {
	fontDict, has := p.Resources.Font.(*core.PdfObjectDictionary)
	if !has {
		return false
	}
	if obj := fontDict.Get(name); obj != nil {
		return true
	}

	return false
}

// HasExtGState checks if ExtGState name is available.
func (p *PdfPage) HasExtGState(name core.PdfObjectName) bool {
	if p.Resources == nil {
		return false
	}

	if p.Resources.ExtGState == nil {
		return false
	}

	egsDict, ok := core.TraceToDirectObject(p.Resources.ExtGState).(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected ExtGState dictionary is not a dictionary: %v", core.TraceToDirectObject(p.Resources.ExtGState))
		return false
	}

	// Update the dictionary.
	obj := egsDict.Get(name)
	has := obj != nil

	return has
}

// AddExtGState adds a graphics state to the XObject resources.
func (p *PdfPage) AddExtGState(name core.PdfObjectName, egs *core.PdfObjectDictionary) error {
	if p.Resources == nil {
		//p.Resources = &PdfPageResources{}
		p.Resources = NewPdfPageResources()
	}

	if p.Resources.ExtGState == nil {
		p.Resources.ExtGState = core.MakeDict()
	}

	egsDict, ok := core.TraceToDirectObject(p.Resources.ExtGState).(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected ExtGState dictionary is not a dictionary: %v", core.TraceToDirectObject(p.Resources.ExtGState))
		return errors.New("type check error")
	}

	egsDict.Set(name, egs)
	return nil
}

// AddFont adds a font dictionary to the Font resources.
func (p *PdfPage) AddFont(name core.PdfObjectName, font core.PdfObject) error {
	if p.Resources == nil {
		p.Resources = NewPdfPageResources()
	}

	if p.Resources.Font == nil {
		p.Resources.Font = core.MakeDict()
	}

	fontDict, ok := core.TraceToDirectObject(p.Resources.Font).(*core.PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected font dictionary is not a dictionary: %v", core.TraceToDirectObject(p.Resources.Font))
		return errors.New("type check error")
	}

	// Update the dictionary.
	fontDict.Set(name, font)

	return nil
}

// WatermarkImageOptions contains options for configuring the watermark process.
type WatermarkImageOptions struct {
	Alpha               float64
	FitToWidth          bool
	PreserveAspectRatio bool
}

// AddWatermarkImage adds a watermark to the page.
func (p *PdfPage) AddWatermarkImage(ximg *XObjectImage, opt WatermarkImageOptions) error {
	// Page dimensions.
	bbox, err := p.GetMediaBox()
	if err != nil {
		return err
	}
	pWidth := bbox.Urx - bbox.Llx
	pHeight := bbox.Ury - bbox.Lly

	wWidth := float64(*ximg.Width)
	xOffset := (float64(pWidth) - float64(wWidth)) / 2
	if opt.FitToWidth {
		wWidth = pWidth
		xOffset = 0
	}
	wHeight := pHeight
	yOffset := float64(0)
	if opt.PreserveAspectRatio {
		wHeight = wWidth * float64(*ximg.Height) / float64(*ximg.Width)
		yOffset = (pHeight - wHeight) / 2
	}

	if p.Resources == nil {
		p.Resources = NewPdfPageResources()
	}

	// Find available image name for this page.
	i := 0
	imgName := core.PdfObjectName(fmt.Sprintf("Imw%d", i))
	for p.Resources.HasXObjectByName(imgName) {
		i++
		imgName = core.PdfObjectName(fmt.Sprintf("Imw%d", i))
	}

	err = p.AddImageResource(imgName, ximg)
	if err != nil {
		return err
	}

	i = 0
	gsName := core.PdfObjectName(fmt.Sprintf("GS%d", i))
	for p.HasExtGState(gsName) {
		i++
		gsName = core.PdfObjectName(fmt.Sprintf("GS%d", i))
	}
	gs0 := core.MakeDict()
	gs0.Set("BM", core.MakeName("Normal"))
	gs0.Set("CA", core.MakeFloat(opt.Alpha))
	gs0.Set("ca", core.MakeFloat(opt.Alpha))
	err = p.AddExtGState(gsName, gs0)
	if err != nil {
		return err
	}

	contentStr := fmt.Sprintf("q\n"+
		"/%s gs\n"+
		"%.0f 0 0 %.0f %.4f %.4f cm\n"+
		"/%s Do\n"+
		"Q", gsName, wWidth, wHeight, xOffset, yOffset, imgName)
	p.AddContentStreamByString(contentStr)

	return nil
}

// AddContentStreamByString adds content stream by string. Puts the content
// string into a stream object and points the content stream towards it.
func (p *PdfPage) AddContentStreamByString(contentStr string) error {
	stream, err := core.MakeStream([]byte(contentStr), core.NewFlateEncoder())
	if err != nil {
		return err
	}

	if p.Contents == nil {
		// If not set, place it directly.
		p.Contents = stream
	} else if contArray, isArray := core.GetArray(p.Contents); isArray {
		// If an array of content streams, append it.
		contArray.Append(stream)
	} else {
		// Only 1 element in place. Wrap inside a new array and add the new one.
		contArray := core.MakeArray(p.Contents, stream)
		p.Contents = contArray
	}

	return nil
}

// AppendContentStream adds content stream by string.  Appends to the last
// contentstream instance if many.
func (p *PdfPage) AppendContentStream(contentStr string) error {
	cstreams, err := p.GetContentStreams()
	if err != nil {
		return err
	}
	if len(cstreams) == 0 {
		cstreams = []string{contentStr}
		return p.SetContentStreams(cstreams, core.NewFlateEncoder())
	}

	var buf bytes.Buffer
	buf.WriteString(cstreams[len(cstreams)-1])
	buf.WriteString("\n")
	buf.WriteString(contentStr)
	cstreams[len(cstreams)-1] = buf.String()

	return p.SetContentStreams(cstreams, core.NewFlateEncoder())
}

// SetContentStreams sets the content streams based on a string array. Will make
// 1 object stream for each string and reference from the page Contents.
// Each stream will be encoded using the encoding specified by the StreamEncoder,
// if empty, will use identity encoding (raw data).
func (p *PdfPage) SetContentStreams(cStreams []string, encoder core.StreamEncoder) error {
	if len(cStreams) == 0 {
		p.Contents = nil
		return nil
	}

	// If encoding is not set, use default raw encoder.
	if encoder == nil {
		encoder = core.NewRawEncoder()
	}

	var streamObjs []*core.PdfObjectStream
	for _, cStream := range cStreams {
		stream := &core.PdfObjectStream{}

		// Make a new stream dict based on the encoding parameters.
		sDict := encoder.MakeStreamDict()

		encoded, err := encoder.EncodeBytes([]byte(cStream))
		if err != nil {
			return err
		}

		sDict.Set("Length", core.MakeInteger(int64(len(encoded))))

		stream.PdfObjectDictionary = sDict
		stream.Stream = []byte(encoded)

		streamObjs = append(streamObjs, stream)
	}

	// Set the page contents.
	// Point directly to the object stream if only one, or embed in an array.
	if len(streamObjs) == 1 {
		p.Contents = streamObjs[0]
	} else {
		contArray := core.MakeArray()
		for _, streamObj := range streamObjs {
			contArray.Append(streamObj)
		}
		p.Contents = contArray
	}

	return nil
}

func getContentStreamAsString(cstreamObj core.PdfObject) (string, error) {
	cstreamObj = core.TraceToDirectObject(cstreamObj)

	switch v := cstreamObj.(type) {
	case *core.PdfObjectString:
		return v.Str(), nil
	case *core.PdfObjectStream:
		buf, err := core.DecodeStream(v)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}

	return "", fmt.Errorf("invalid content stream object holder (%T)", cstreamObj)
}

// GetContentStreams returns the content stream as an array of strings.
func (p *PdfPage) GetContentStreams() ([]string, error) {
	if p.Contents == nil {
		return nil, nil
	}
	contents := core.TraceToDirectObject(p.Contents)

	var cStreamObjs []core.PdfObject
	if contArray, ok := contents.(*core.PdfObjectArray); ok {
		cStreamObjs = contArray.Elements()
	} else {
		cStreamObjs = []core.PdfObject{contents}
	}

	var cStreams []string
	for _, cStreamObj := range cStreamObjs {
		cStreamStr, err := getContentStreamAsString(cStreamObj)
		if err != nil {
			return nil, err
		}
		cStreams = append(cStreams, cStreamStr)
	}

	return cStreams, nil
}

// GetAllContentStreams gets all the content streams for a page as one string.
func (p *PdfPage) GetAllContentStreams() (string, error) {
	cstreams, err := p.GetContentStreams()
	if err != nil {
		return "", err
	}
	return strings.Join(cstreams, " "), nil
}

// PdfPageResourcesColorspaces contains the colorspace in the PdfPageResources.
// Needs to have matching name and colorspace map entry. The Names define the order.
type PdfPageResourcesColorspaces struct {
	Names       []string
	Colorspaces map[string]PdfColorspace

	container *core.PdfIndirectObject
}

// NewPdfPageResourcesColorspaces returns a new PdfPageResourcesColorspaces object.
func NewPdfPageResourcesColorspaces() *PdfPageResourcesColorspaces {
	colorspaces := &PdfPageResourcesColorspaces{}
	colorspaces.Names = []string{}
	colorspaces.Colorspaces = map[string]PdfColorspace{}
	colorspaces.container = &core.PdfIndirectObject{}
	return colorspaces
}

// Set sets the colorspace corresponding to key. Add to Names if not set.
func (rcs *PdfPageResourcesColorspaces) Set(key core.PdfObjectName, val PdfColorspace) {
	if _, has := rcs.Colorspaces[string(key)]; !has {
		rcs.Names = append(rcs.Names, string(key))
	}
	rcs.Colorspaces[string(key)] = val
}

func newPdfPageResourcesColorspacesFromPdfObject(obj core.PdfObject) (*PdfPageResourcesColorspaces, error) {
	colorspaces := &PdfPageResourcesColorspaces{}

	if indObj, isIndirect := obj.(*core.PdfIndirectObject); isIndirect {
		colorspaces.container = indObj
		obj = indObj.PdfObject
	}

	dict, ok := core.GetDict(obj)
	if !ok {
		return nil, errors.New("CS attribute type error")
	}

	colorspaces.Names = []string{}
	colorspaces.Colorspaces = map[string]PdfColorspace{}

	for _, csName := range dict.Keys() {
		csObj := dict.Get(csName)
		colorspaces.Names = append(colorspaces.Names, string(csName))
		cs, err := NewPdfColorspaceFromPdfObject(csObj)
		if err != nil {
			return nil, err
		}
		colorspaces.Colorspaces[string(csName)] = cs
	}

	return colorspaces, nil
}

// ToPdfObject returns the PDF representation of the colorspace.
func (rcs *PdfPageResourcesColorspaces) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()
	for _, csName := range rcs.Names {
		dict.Set(core.PdfObjectName(csName), rcs.Colorspaces[csName].ToPdfObject())
	}

	if rcs.container != nil {
		rcs.container.PdfObject = dict
		return rcs.container
	}

	return dict
}
