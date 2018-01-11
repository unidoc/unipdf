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
	"errors"
	"fmt"
	"strings"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

// PDF page object (7.7.3.3 - Table 30).
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

func (this *PdfPage) setContainer(container *PdfIndirectObject) {
	container.PdfObject = this.pageDict
	this.primitive = container
}

func (this *PdfPage) Duplicate() *PdfPage {
	var dup PdfPage
	dup = *this
	dup.pageDict = MakeDict()
	dup.primitive = MakeIndirectObject(dup.pageDict)

	return &dup
}

// Build a PdfPage based on the underlying dictionary.
// Used in loading existing PDF files.
// Note that a new container is created (indirect object).
func (reader *PdfReader) newPdfPageFromDict(p *PdfObjectDictionary) (*PdfPage, error) {
	page := NewPdfPage()
	page.pageDict = p //XXX?

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
		obj, err = reader.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		strObj, ok := TraceToDirectObject(obj).(*PdfObjectString)
		if !ok {
			return nil, errors.New("Page dictionary LastModified != string")
		}
		lastmod, err := NewPdfDate(string(*strObj))
		if err != nil {
			return nil, err
		}
		page.LastModified = &lastmod
	}

	if obj := d.Get("Resources"); obj != nil {
		var err error
		obj, err = reader.traceToObject(obj)
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
		obj, err = reader.traceToObject(obj)
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
		obj, err = reader.traceToObject(obj)
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
		obj, err = reader.traceToObject(obj)
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
		obj, err = reader.traceToObject(obj)
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
		obj, err = reader.traceToObject(obj)
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
		obj, err = reader.traceToObject(obj)
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
	page.Annotations, err = reader.LoadAnnotations(&d)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (reader *PdfReader) LoadAnnotations(d *PdfObjectDictionary) ([]*PdfAnnotation, error) {
	annotsObj := d.Get("Annots")
	if annotsObj == nil {
		return nil, nil
	}

	var err error
	annotsObj, err = reader.traceToObject(annotsObj)
	if err != nil {
		return nil, err
	}
	annotsArr, ok := TraceToDirectObject(annotsObj).(*PdfObjectArray)
	if !ok {
		return nil, fmt.Errorf("Annots not an array")
	}

	annotations := []*PdfAnnotation{}
	for _, obj := range *annotsArr {
		obj, err = reader.traceToObject(obj)
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

		annot, err := reader.newPdfAnnotationFromIndirectObject(indirectObj)
		if err != nil {
			return nil, err
		}
		annotations = append(annotations, annot)
	}

	return annotations, nil
}

// Get the inheritable media box value, either from the page
// or a higher up page/pages struct.
func (this *PdfPage) GetMediaBox() (*PdfRectangle, error) {
	if this.MediaBox != nil {
		return this.MediaBox, nil
	}

	node := this.Parent
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
func (this *PdfPage) getResources() (*PdfPageResources, error) {
	if this.Resources != nil {
		return this.Resources, nil
	}

	node := this.Parent
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
			prDict, ok := obj.(*PdfObjectDictionary)
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

// Convert the Page to a PDF object dictionary.
func (this *PdfPage) GetPageDict() *PdfObjectDictionary {
	p := this.pageDict
	p.Set("Type", MakeName("Page"))
	p.Set("Parent", this.Parent)

	if this.LastModified != nil {
		p.Set("LastModified", this.LastModified.ToPdfObject())
	}
	if this.Resources != nil {
		p.Set("Resources", this.Resources.ToPdfObject())
	}
	if this.CropBox != nil {
		p.Set("CropBox", this.CropBox.ToPdfObject())
	}
	if this.MediaBox != nil {
		p.Set("MediaBox", this.MediaBox.ToPdfObject())
	}
	if this.BleedBox != nil {
		p.Set("BleedBox", this.BleedBox.ToPdfObject())
	}
	if this.TrimBox != nil {
		p.Set("TrimBox", this.TrimBox.ToPdfObject())
	}
	if this.ArtBox != nil {
		p.Set("ArtBox", this.ArtBox.ToPdfObject())
	}
	p.SetIfNotNil("BoxColorInfo", this.BoxColorInfo)
	p.SetIfNotNil("Contents", this.Contents)

	if this.Rotate != nil {
		p.Set("Rotate", MakeInteger(*this.Rotate))
	}

	p.SetIfNotNil("Group", this.Group)
	p.SetIfNotNil("Thumb", this.Thumb)
	p.SetIfNotNil("B", this.B)
	p.SetIfNotNil("Dur", this.Dur)
	p.SetIfNotNil("Trans", this.Trans)
	p.SetIfNotNil("AA", this.AA)
	p.SetIfNotNil("Metadata", this.Metadata)
	p.SetIfNotNil("PieceInfo", this.PieceInfo)
	p.SetIfNotNil("StructParents", this.StructParents)
	p.SetIfNotNil("ID", this.ID)
	p.SetIfNotNil("PZ", this.PZ)
	p.SetIfNotNil("SeparationInfo", this.SeparationInfo)
	p.SetIfNotNil("Tabs", this.Tabs)
	p.SetIfNotNil("TemplateInstantiated", this.TemplateInstantiated)
	p.SetIfNotNil("PresSteps", this.PresSteps)
	p.SetIfNotNil("UserUnit", this.UserUnit)
	p.SetIfNotNil("VP", this.VP)

	if this.Annotations != nil {
		arr := PdfObjectArray{}
		for _, annot := range this.Annotations {
			if subannot := annot.GetContext(); subannot != nil {
				arr = append(arr, subannot.ToPdfObject())
			} else {
				// Generic annotation dict (without subtype).
				arr = append(arr, annot.ToPdfObject())
			}
		}
		p.Set("Annots", &arr)
	}

	return p
}

// Get the page object as an indirect objects.  Wraps the Page
// dictionary into an indirect object.
func (this *PdfPage) GetPageAsIndirectObject() *PdfIndirectObject {
	return this.primitive
}

func (this *PdfPage) GetContainingPdfObject() PdfObject {
	return this.primitive
}

func (this *PdfPage) ToPdfObject() PdfObject {
	container := this.primitive
	this.GetPageDict() // update.
	return container
}

// Add an image to the XObject resources.
func (this *PdfPage) AddImageResource(name PdfObjectName, ximg *XObjectImage) error {
	var xresDict *PdfObjectDictionary
	if this.Resources.XObject == nil {
		xresDict = MakeDict()
		this.Resources.XObject = xresDict
	} else {
		var ok bool
		xresDict, ok = (this.Resources.XObject).(*PdfObjectDictionary)
		if !ok {
			return errors.New("Invalid xres dict type")
		}

	}
	// Make a stream object container.
	xresDict.Set(name, ximg.ToPdfObject())

	return nil
}

// Check if has XObject resource by name.
func (this *PdfPage) HasXObjectByName(name PdfObjectName) bool {
	xresDict, has := this.Resources.XObject.(*PdfObjectDictionary)
	if !has {
		return false
	}

	if obj := xresDict.Get(name); obj != nil {
		return true
	} else {
		return false
	}
}

// Get XObject by name.
func (this *PdfPage) GetXObjectByName(name PdfObjectName) (PdfObject, bool) {
	xresDict, has := this.Resources.XObject.(*PdfObjectDictionary)
	if !has {
		return nil, false
	}

	if obj := xresDict.Get(name); obj != nil {
		return obj, true
	} else {
		return nil, false
	}
}

// Check if has font resource by name.
func (this *PdfPage) HasFontByName(name PdfObjectName) bool {
	fontDict, has := this.Resources.Font.(*PdfObjectDictionary)
	if !has {
		return false
	}

	if obj := fontDict.Get(name); obj != nil {
		return true
	} else {
		return false
	}
}

// Check if ExtGState name is available.
func (this *PdfPage) HasExtGState(name PdfObjectName) bool {
	if this.Resources == nil {
		return false
	}

	if this.Resources.ExtGState == nil {
		return false
	}

	egsDict, ok := TraceToDirectObject(this.Resources.ExtGState).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected ExtGState dictionary is not a dictionary: %v", TraceToDirectObject(this.Resources.ExtGState))
		return false
	}

	// Update the dictionary.
	obj := egsDict.Get(name)
	has := obj != nil

	return has
}

// Add a graphics state to the XObject resources.
func (this *PdfPage) AddExtGState(name PdfObjectName, egs *PdfObjectDictionary) error {
	if this.Resources == nil {
		//this.Resources = &PdfPageResources{}
		this.Resources = NewPdfPageResources()
	}

	if this.Resources.ExtGState == nil {
		this.Resources.ExtGState = MakeDict()
	}

	egsDict, ok := TraceToDirectObject(this.Resources.ExtGState).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected ExtGState dictionary is not a dictionary: %v", TraceToDirectObject(this.Resources.ExtGState))
		return errors.New("Type check error")
	}

	egsDict.Set(name, egs)
	return nil
}

// Add a font dictionary to the Font resources.
func (this *PdfPage) AddFont(name PdfObjectName, font PdfObject) error {
	if this.Resources == nil {
		this.Resources = NewPdfPageResources()
	}

	if this.Resources.Font == nil {
		this.Resources.Font = MakeDict()
	}

	fontDict, ok := TraceToDirectObject(this.Resources.Font).(*PdfObjectDictionary)
	if !ok {
		common.Log.Debug("Expected font dictionary is not a dictionary: %v", TraceToDirectObject(this.Resources.Font))
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

// Add a watermark to the page.
func (this *PdfPage) AddWatermarkImage(ximg *XObjectImage, opt WatermarkImageOptions) error {
	// Page dimensions.
	bbox, err := this.GetMediaBox()
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

	if this.Resources == nil {
		this.Resources = NewPdfPageResources()
	}

	// Find available image name for this page.
	i := 0
	imgName := PdfObjectName(fmt.Sprintf("Imw%d", i))
	for this.Resources.HasXObjectByName(imgName) {
		i++
		imgName = PdfObjectName(fmt.Sprintf("Imw%d", i))
	}

	err = this.AddImageResource(imgName, ximg)
	if err != nil {
		return err
	}

	i = 0
	gsName := PdfObjectName(fmt.Sprintf("GS%d", i))
	for this.HasExtGState(gsName) {
		i++
		gsName = PdfObjectName(fmt.Sprintf("GS%d", i))
	}
	gs0 := MakeDict()
	gs0.Set("BM", MakeName("Normal"))
	gs0.Set("CA", MakeFloat(opt.Alpha))
	gs0.Set("ca", MakeFloat(opt.Alpha))
	err = this.AddExtGState(gsName, gs0)
	if err != nil {
		return err
	}

	contentStr := fmt.Sprintf("q\n"+
		"/%s gs\n"+
		"%.0f 0 0 %.0f %.4f %.4f cm\n"+
		"/%s Do\n"+
		"Q", gsName, wWidth, wHeight, xOffset, yOffset, imgName)
	this.AddContentStreamByString(contentStr)

	return nil
}

// Add content stream by string.  Puts the content string into a stream
// object and points the content stream towards it.
func (this *PdfPage) AddContentStreamByString(contentStr string) {
	stream := PdfObjectStream{}

	sDict := MakeDict()
	stream.PdfObjectDictionary = sDict

	sDict.Set("Length", MakeInteger(int64(len(contentStr))))
	stream.Stream = []byte(contentStr)

	if this.Contents == nil {
		// If not set, place it directly.
		this.Contents = &stream
	} else if contArray, isArray := TraceToDirectObject(this.Contents).(*PdfObjectArray); isArray {
		// If an array of content streams, append it.
		*contArray = append(*contArray, &stream)
	} else {
		// Only 1 element in place. Wrap inside a new array and add the new one.
		contArray := PdfObjectArray{}
		contArray = append(contArray, this.Contents)
		contArray = append(contArray, &stream)
		this.Contents = &contArray
	}
}

// Set the content streams based on a string array.  Will make 1 object stream
// for each string and reference from the page Contents.  Each stream will be
// encoded using the encoding specified by the StreamEncoder, if empty, will
// use identity encoding (raw data).
func (this *PdfPage) SetContentStreams(cStreams []string, encoder StreamEncoder) error {
	if len(cStreams) == 0 {
		this.Contents = nil
		return nil
	}

	// If encoding is not set, use default raw encoder.
	if encoder == nil {
		encoder = NewRawEncoder()
	}

	streamObjs := []*PdfObjectStream{}
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
		this.Contents = streamObjs[0]
	} else {
		contArray := PdfObjectArray{}
		for _, streamObj := range streamObjs {
			contArray = append(contArray, streamObj)
		}
		this.Contents = &contArray
	}

	return nil
}

func getContentStreamAsString(cstreamObj PdfObject) (string, error) {
	if cstream, ok := TraceToDirectObject(cstreamObj).(*PdfObjectString); ok {
		return string(*cstream), nil
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

// Get Content Stream as an array of strings.
func (this *PdfPage) GetContentStreams() ([]string, error) {
	if this.Contents == nil {
		return nil, nil
	}

	contents := TraceToDirectObject(this.Contents)
	if contArray, isArray := contents.(*PdfObjectArray); isArray {
		// If an array of content streams, append it.
		cstreams := []string{}
		for _, cstreamObj := range *contArray {
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

// Get all the content streams for a page as one string.
func (this *PdfPage) GetAllContentStreams() (string, error) {
	cstreams, err := this.GetContentStreams()
	if err != nil {
		return "", err
	}
	return strings.Join(cstreams, " "), nil
}

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

// Set the colorspace corresponding to key.  Add to Names if not set.
func (this *PdfPageResourcesColorspaces) Set(key PdfObjectName, val PdfColorspace) {
	if _, has := this.Colorspaces[string(key)]; !has {
		this.Names = append(this.Names, string(key))
	}
	this.Colorspaces[string(key)] = val
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

func (this *PdfPageResourcesColorspaces) ToPdfObject() PdfObject {
	dict := MakeDict()
	for _, csName := range this.Names {
		dict.Set(PdfObjectName(csName), this.Colorspaces[csName].ToPdfObject())
	}

	if this.container != nil {
		this.container.PdfObject = dict
		return this.container
	}

	return dict
}
