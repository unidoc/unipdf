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

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

// PdfPage represents a page in a PDF document. (7.7.3.3 - Table 30).
type PdfPage struct {
	Parent               PdfObject
	LastModified         *PdfDate
	Resources            *PdfPageResources
	CropBox              *PdfRectangle
	MediaBox             *PdfRectangle
	BleedBox             *PdfRectangle
	TrimBox              *PdfRectangle
	ArtBox               *PdfRectangle
	BoxColorInfo         PdfObject
	Contents             PdfObject
	Rotate               *int64
	Group                PdfObject
	Thumb                PdfObject
	B                    PdfObject
	Dur                  PdfObject
	Trans                PdfObject
	AA                   PdfObject
	Metadata             PdfObject
	PieceInfo            PdfObject
	StructParents        PdfObject
	ID                   PdfObject
	PZ                   PdfObject
	SeparationInfo       PdfObject
	Tabs                 PdfObject
	TemplateInstantiated PdfObject
	PresSteps            PdfObject
	UserUnit             PdfObject
	VP                   PdfObject

	Annotations []*PdfAnnotation

	// Primitive container.
	pageDict  *PdfObjectDictionary
	primitive *PdfIndirectObject
}

func NewPdfPage() *PdfPage {
	page := PdfPage{}
	page.pageDict = MakeDict()

	container := PdfIndirectObject{}
	container.PdfObject = page.pageDict
	page.primitive = &container

	return &page
}

func (p *PdfPage) setContainer(container *PdfIndirectObject) {
	container.PdfObject = p.pageDict
	p.primitive = container
}

func (p *PdfPage) Duplicate() *PdfPage {
	var dup PdfPage
	dup = *p
	dup.pageDict = MakeDict()
	dup.primitive = MakeIndirectObject(dup.pageDict)

	return &dup
}

// Build a PdfPage based on the underlying dictionary.
// Used in loading existing PDF files.
// Note that a new container is created (indirect object).
func (r *PdfReader) newPdfPageFromDict(p *PdfObjectDictionary) (*PdfPage, error) {
	page := NewPdfPage()
	page.pageDict = p // TODO

	d := *p

	pType, ok := d.Get("Type").(*PdfObjectName)
	if !ok {
		return nil, errors.New("Missing/Invalid Page dictionary Type")
	}
	if *pType != "Page" {
		return nil, errors.New("Page dictionary Type != Page")
	}

	if obj := d.Get("Parent"); obj != nil {
		page.Parent = obj
	}

	if obj := d.Get("LastModified"); obj != nil {
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		strObj, ok := TraceToDirectObject(obj).(*PdfObjectString)
		if !ok {
			return nil, errors.New("Page dictionary LastModified != string")
		}
		lastmod, err := NewPdfDate(strObj.Str())
		if err != nil {
			return nil, err
		}
		page.LastModified = &lastmod
	}

	if obj := d.Get("Resources"); obj != nil {
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}

		dict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
		if !ok {
			return nil, fmt.Errorf("Invalid resource dictionary (%T)", obj)
		}

		page.Resources, err = NewPdfPageResourcesFromDict(dict)
		if err != nil {
			return nil, err
		}
	} else {
		// If Resources not explicitly defined, look up the tree (Parent objects) using
		// the getResources() function. Resources should always be accessible.
		resources, err := page.getResources()
		if err != nil {
			return nil, err
		}
		if resources == nil {
			resources = NewPdfPageResources()
		}
		page.Resources = resources
	}

	if obj := d.Get("MediaBox"); obj != nil {
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		boxArr, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page MediaBox not an array")
		}
		page.MediaBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("CropBox"); obj != nil {
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		boxArr, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page CropBox not an array")
		}
		page.CropBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("BleedBox"); obj != nil {
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		boxArr, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page BleedBox not an array")
		}
		page.BleedBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("TrimBox"); obj != nil {
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		boxArr, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page TrimBox not an array")
		}
		page.TrimBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj := d.Get("ArtBox"); obj != nil {
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		boxArr, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page ArtBox not an array")
		}
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
		var err error
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		iObj, ok := TraceToDirectObject(obj).(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid Page Rotate object")
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
	//if obj := d.Get("Annots"); obj != nil {
	//	page.Annots = obj
	//}
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

	var err error
	page.Annotations, err = r.LoadAnnotations(&d)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (r *PdfReader) LoadAnnotations(d *PdfObjectDictionary) ([]*PdfAnnotation, error) {
	annotsObj := d.Get("Annots")
	if annotsObj == nil {
		return nil, nil
	}

	var err error
	annotsObj, err = r.traceToObject(annotsObj)
	if err != nil {
		return nil, err
	}
	annotsArr, ok := TraceToDirectObject(annotsObj).(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Annots not an array")
	}

	var annotations []*PdfAnnotation
	for _, obj := range annotsArr.Elements() {
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}

		// Technically all annotation dictionaries should be inside indirect objects.
		// In reality, sometimes the annotation dictionary is inline within the Annots array.
		if _, isNull := obj.(*PdfObjectNull); isNull {
			// Can safely ignore.
			continue
		}

		annotDict, isDict := obj.(*PdfObjectDictionary)
		indirectObj, isIndirect := obj.(*PdfIndirectObject)
		if isDict {
			// Create a container; indirect object; around the dictionary.
			indirectObj = &PdfIndirectObject{}
			indirectObj.PdfObject = annotDict
		} else {
			if !isIndirect {
				return nil, fmt.Errorf("Annotation not in an indirect object")
			}
		}

		annot, err := r.newPdfAnnotationFromIndirectObject(indirectObj)
		if err != nil {
			return nil, err
		}
		annotations = append(annotations, annot)
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
		dictObj, ok := node.(*PdfIndirectObject)
		if !ok {
			return nil, errors.New("Invalid parent object")
		}

		dict, ok := dictObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			return nil, errors.New("Invalid parent objects dictionary")
		}

		if obj := dict.Get("MediaBox"); obj != nil {
			arr, ok := obj.(*PdfObjectArray)
			if !ok {
				return nil, errors.New("Invalid media box")
			}
			rect, err := NewPdfRectangle(*arr)

			if err != nil {
				return nil, err
			}

			return rect, nil
		}

		node = dict.Get("Parent")
	}

	return nil, errors.New("Media box not defined")
}

// Get the inheritable resources, either from the page or or a higher up page/pages struct.
func (p *PdfPage) getResources() (*PdfPageResources, error) {
	if p.Resources != nil {
		return p.Resources, nil
	}

	node := p.Parent
	for node != nil {
		dictObj, ok := node.(*PdfIndirectObject)
		if !ok {
			return nil, errors.New("Invalid parent object")
		}

		dict, ok := dictObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			return nil, errors.New("Invalid parent objects dictionary")
		}

		if obj := dict.Get("Resources"); obj != nil {
			prDict, ok := TraceToDirectObject(obj).(*PdfObjectDictionary)
			if !ok {
				return nil, errors.New("Invalid resource dict!")
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

// GetPageDict convert the Page to a PDF object dictionary.
func (p *PdfPage) GetPageDict() *PdfObjectDictionary {
	d := p.pageDict
	d.Clear()
	d.Set("Type", MakeName("Page"))
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
		d.Set("Rotate", MakeInteger(*p.Rotate))
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

	if p.Annotations != nil {
		arr := MakeArray()
		for _, annot := range p.Annotations {
			if subannot := annot.GetContext(); subannot != nil {
				arr.Append(subannot.ToPdfObject())
			} else {
				// Generic annotation dict (without subtype).
				arr.Append(annot.ToPdfObject())
			}
		}
		d.Set("Annots", arr)
	}

	return d
}

// GetPageAsIndirectObject returns the page as a dictionary within an PdfIndirectObject.
func (p *PdfPage) GetPageAsIndirectObject() *PdfIndirectObject {
	return p.primitive
}

// GetContainingPdfObject returns the page as a dictionary within an PdfIndirectObject.
func (p *PdfPage) GetContainingPdfObject() PdfObject {
	return p.primitive
}

// ToPdfObject converts the PdfPage to a dictionary within an indirect object container.
func (p *PdfPage) ToPdfObject() PdfObject {
	container := p.primitive
	p.GetPageDict() // update.
	return container
}

// AddImageResource adds an image to the XObject resources.
func (p *PdfPage) AddImageResource(name PdfObjectName, ximg *XObjectImage) error {
	var xresDict *PdfObjectDictionary
	if p.Resources.XObject == nil {
		xresDict = MakeDict()
		p.Resources.XObject = xresDict
	} else {
		var ok bool
		xresDict, ok = (p.Resources.XObject).(*PdfObjectDictionary)
		if !ok {
			return errors.New("Invalid xres dict type")
		}

	}
	// Make a stream object container.
	xresDict.Set(name, ximg.ToPdfObject())

	return nil
}

// HasXObjectByName checks if has XObject resource by name.
func (p *PdfPage) HasXObjectByName(name PdfObjectName) bool {
	xresDict, has := p.Resources.XObject.(*PdfObjectDictionary)
	if !has {
		return false
	}

	if obj := xresDict.Get(name); obj != nil {
		return true
	} else {
		return false
	}
}

// GetXObjectByName get XObject by name.
func (p *PdfPage) GetXObjectByName(name PdfObjectName) (PdfObject, bool) {
	xresDict, has := p.Resources.XObject.(*PdfObjectDictionary)
	if !has {
		return nil, false
	}

	if obj := xresDict.Get(name); obj != nil {
		return obj, true
	} else {
		return nil, false
	}
}

// HasFontByName checks if has font resource by name.
func (p *PdfPage) HasFontByName(name PdfObjectName) bool {
	fontDict, has := p.Resources.Font.(*PdfObjectDictionary)
	if !has {
		return false
	}

	if obj := fontDict.Get(name); obj != nil {
		return true
	} else {
		return false
	}
}

// HasExtGState checks if ExtGState name is available.
func (p *PdfPage) HasExtGState(name PdfObjectName) bool {
	if p.Resources == nil {
		return false
	}

	if p.Resources.ExtGState == nil {
		return false
	}

	egsDict, ok := TraceToDirectObject(p.Resources.ExtGState).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected ExtGState dictionary is not a dictionary: %v", TraceToDirectObject(p.Resources.ExtGState))
		return false
	}

	// Update the dictionary.
	obj := egsDict.Get(name)
	has := obj != nil

	return has
}

// AddExtGState adds a graphics state to the XObject resources.
func (p *PdfPage) AddExtGState(name PdfObjectName, egs *PdfObjectDictionary) error {
	if p.Resources == nil {
		//p.Resources = &PdfPageResources{}
		p.Resources = NewPdfPageResources()
	}

	if p.Resources.ExtGState == nil {
		p.Resources.ExtGState = MakeDict()
	}

	egsDict, ok := TraceToDirectObject(p.Resources.ExtGState).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected ExtGState dictionary is not a dictionary: %v", TraceToDirectObject(p.Resources.ExtGState))
		return errors.New("Type check error")
	}

	egsDict.Set(name, egs)
	return nil
}

// AddFont adds a font dictionary to the Font resources.
func (p *PdfPage) AddFont(name PdfObjectName, font PdfObject) error {
	if p.Resources == nil {
		p.Resources = NewPdfPageResources()
	}

	if p.Resources.Font == nil {
		p.Resources.Font = MakeDict()
	}

	fontDict, ok := TraceToDirectObject(p.Resources.Font).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected font dictionary is not a dictionary: %v", TraceToDirectObject(p.Resources.Font))
		return errors.New("Type check error")
	}

	// Update the dictionary.
	fontDict.Set(name, font)

	return nil
}

type WatermarkImageOptions struct {
	Alpha               float64
	FitToWidth          bool
	PreserveAspectRatio bool
}

// AddWatermarkImage add a watermark to the page.
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
	imgName := PdfObjectName(fmt.Sprintf("Imw%d", i))
	for p.Resources.HasXObjectByName(imgName) {
		i++
		imgName = PdfObjectName(fmt.Sprintf("Imw%d", i))
	}

	err = p.AddImageResource(imgName, ximg)
	if err != nil {
		return err
	}

	i = 0
	gsName := PdfObjectName(fmt.Sprintf("GS%d", i))
	for p.HasExtGState(gsName) {
		i++
		gsName = PdfObjectName(fmt.Sprintf("GS%d", i))
	}
	gs0 := MakeDict()
	gs0.Set("BM", MakeName("Normal"))
	gs0.Set("CA", MakeFloat(opt.Alpha))
	gs0.Set("ca", MakeFloat(opt.Alpha))
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

// AddContentStreamByString adds content stream by string.  Puts the content string into a stream
// object and points the content stream towards it.
func (p *PdfPage) AddContentStreamByString(contentStr string) error {
	stream, err := MakeStream([]byte(contentStr), NewFlateEncoder())
	if err != nil {
		return err
	}

	if p.Contents == nil {
		// If not set, place it directly.
		p.Contents = stream
	} else if contArray, isArray := TraceToDirectObject(p.Contents).(*PdfObjectArray); isArray {
		// If an array of content streams, append it.
		contArray.Append(stream)
	} else {
		// Only 1 element in place. Wrap inside a new array and add the new one.
		contArray := MakeArray(p.Contents, stream)
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
		return p.SetContentStreams(cstreams, NewFlateEncoder())
	}

	var buf bytes.Buffer
	buf.WriteString(cstreams[len(cstreams)-1])
	buf.WriteString("\n")
	buf.WriteString(contentStr)
	cstreams[len(cstreams)-1] = buf.String()

	return p.SetContentStreams(cstreams, NewFlateEncoder())
}

// SetContentStreams sets the content streams based on a string array.  Will make 1 object stream
// for each string and reference from the page Contents.  Each stream will be
// encoded using the encoding specified by the StreamEncoder, if empty, will
// use identity encoding (raw data).
func (p *PdfPage) SetContentStreams(cStreams []string, encoder StreamEncoder) error {
	if len(cStreams) == 0 {
		p.Contents = nil
		return nil
	}

	// If encoding is not set, use default raw encoder.
	if encoder == nil {
		encoder = NewRawEncoder()
	}

	var streamObjs []*PdfObjectStream
	for _, cStream := range cStreams {
		stream := &PdfObjectStream{}

		// Make a new stream dict based on the encoding parameters.
		sDict := encoder.MakeStreamDict()

		encoded, err := encoder.EncodeBytes([]byte(cStream))
		if err != nil {
			return err
		}

		sDict.Set("Length", MakeInteger(int64(len(encoded))))

		stream.PdfObjectDictionary = sDict
		stream.Stream = []byte(encoded)

		streamObjs = append(streamObjs, stream)
	}

	// Set the page contents.
	// Point directly to the object stream if only one, or embed in an array.
	if len(streamObjs) == 1 {
		p.Contents = streamObjs[0]
	} else {
		contArray := MakeArray()
		for _, streamObj := range streamObjs {
			contArray.Append(streamObj)
		}
		p.Contents = contArray
	}

	return nil
}

func getContentStreamAsString(cstreamObj PdfObject) (string, error) {
	if cstream, ok := TraceToDirectObject(cstreamObj).(*PdfObjectString); ok {
		return cstream.Str(), nil
	}

	if cstream, ok := TraceToDirectObject(cstreamObj).(*PdfObjectStream); ok {
		buf, err := DecodeStream(cstream)
		if err != nil {
			return "", err
		}

		return string(buf), nil
	}
	return "", fmt.Errorf("Invalid content stream object holder (%T)", TraceToDirectObject(cstreamObj))
}

// GetContentStreams returns the content stream as an array of strings.
func (p *PdfPage) GetContentStreams() ([]string, error) {
	if p.Contents == nil {
		return nil, nil
	}

	contents := TraceToDirectObject(p.Contents)
	if contArray, isArray := contents.(*PdfObjectArray); isArray {
		// If an array of content streams, append it.
		var cstreams []string
		for _, cstreamObj := range contArray.Elements() {
			cstreamStr, err := getContentStreamAsString(cstreamObj)
			if err != nil {
				return nil, err
			}
			cstreams = append(cstreams, cstreamStr)
		}
		return cstreams, nil
	} else {
		// Only 1 element in place. Wrap inside a new array and add the new one.
		cstreamStr, err := getContentStreamAsString(contents)
		if err != nil {
			return nil, err
		}
		cstreams := []string{cstreamStr}
		return cstreams, nil
	}
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

	container *PdfIndirectObject
}

func NewPdfPageResourcesColorspaces() *PdfPageResourcesColorspaces {
	colorspaces := &PdfPageResourcesColorspaces{}
	colorspaces.Names = []string{}
	colorspaces.Colorspaces = map[string]PdfColorspace{}
	colorspaces.container = &PdfIndirectObject{}
	return colorspaces
}

// Set sets the colorspace corresponding to key.  Add to Names if not set.
func (r *PdfPageResourcesColorspaces) Set(key PdfObjectName, val PdfColorspace) {
	if _, has := r.Colorspaces[string(key)]; !has {
		r.Names = append(r.Names, string(key))
	}
	r.Colorspaces[string(key)] = val
}

func newPdfPageResourcesColorspacesFromPdfObject(obj PdfObject) (*PdfPageResourcesColorspaces, error) {
	colorspaces := &PdfPageResourcesColorspaces{}

	if indObj, isIndirect := obj.(*PdfIndirectObject); isIndirect {
		colorspaces.container = indObj
		obj = indObj.PdfObject
	}

	dict, ok := obj.(*PdfObjectDictionary)
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

func (r *PdfPageResourcesColorspaces) ToPdfObject() PdfObject {
	dict := MakeDict()
	for _, csName := range r.Names {
		dict.Set(PdfObjectName(csName), r.Colorspaces[csName].ToPdfObject())
	}

	if r.container != nil {
		r.container.PdfObject = dict
		return r.container
	}

	return dict
}
