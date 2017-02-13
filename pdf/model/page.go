/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

//
// Allow higher level manipulation of PDF files and pages.
// This can be continously expanded to support more and more features.
// Generic handling can be done by defining elements as PdfObject which
// can later be replaced and fully defined.
//

package model

import (
	"errors"
	"fmt"

	. "github.com/unidoc/unidoc/pdf/core"
)

// PDF page object (7.7.3.3 - Table 30).
type PdfPage struct {
	Parent       PdfObject
	LastModified *PdfDate
	Resources    *PdfPageResources
	CropBox      *PdfRectangle
	MediaBox     *PdfRectangle
	BleedBox     *PdfRectangle
	TrimBox      *PdfRectangle
	ArtBox       *PdfRectangle
	BoxColorInfo PdfObject
	Contents     PdfObject
	Rotate       *int64
	Group        PdfObject
	Thumb        PdfObject
	B            PdfObject
	Dur          PdfObject
	Trans        PdfObject
	//Annots               PdfObject
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
	//Annotations
	Annotations []*PdfAnnotation

	// Primitive container.
	pageDict  *PdfObjectDictionary
	primitive *PdfIndirectObject
}

func NewPdfPage() *PdfPage {
	page := PdfPage{}
	page.pageDict = &PdfObjectDictionary{}

	container := PdfIndirectObject{}
	container.PdfObject = page.pageDict
	page.primitive = &container

	return &page
}

func (this *PdfPage) setContainer(container *PdfIndirectObject) {
	container.PdfObject = this.pageDict
	this.primitive = container
}

// Build a PdfPage based on the underlying dictionary.
// Used in loading existing PDF files.
// Note that a new container is created (indirect object).
func (reader *PdfReader) newPdfPageFromDict(p *PdfObjectDictionary) (*PdfPage, error) {
	page := NewPdfPage()

	d := *p

	pType, ok := d["Type"].(*PdfObjectName)
	if !ok {
		return nil, errors.New("Missing/Invalid Page dictionary Type")
	}
	if *pType != "Page" {
		return nil, errors.New("Page dictionary Type != Page")
	}

	if obj, isDefined := d["Parent"]; isDefined {
		page.Parent = obj
	}

	if obj, isDefined := d["LastModified"]; isDefined {
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

	if obj, isDefined := d["Resources"]; isDefined {
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
	}

	if obj, isDefined := d["MediaBox"]; isDefined {
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
	if obj, isDefined := d["CropBox"]; isDefined {
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
	if obj, isDefined := d["BleedBox"]; isDefined {
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
	if obj, isDefined := d["TrimBox"]; isDefined {
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
	if obj, isDefined := d["ArtBox"]; isDefined {
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
	if obj, isDefined := d["BoxColorInfo"]; isDefined {
		page.BoxColorInfo = obj
	}
	if obj, isDefined := d["Contents"]; isDefined {
		page.Contents = obj
	}
	if obj, isDefined := d["Rotate"]; isDefined {
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
	if obj, isDefined := d["Group"]; isDefined {
		page.Group = obj
	}
	if obj, isDefined := d["Thumb"]; isDefined {
		page.Thumb = obj
	}
	if obj, isDefined := d["B"]; isDefined {
		page.B = obj
	}
	if obj, isDefined := d["Dur"]; isDefined {
		page.Dur = obj
	}
	if obj, isDefined := d["Trans"]; isDefined {
		page.Trans = obj
	}
	//if obj, isDefined := d["Annots"]; isDefined {
	//	page.Annots = obj
	//}
	if obj, isDefined := d["AA"]; isDefined {
		page.AA = obj
	}
	if obj, isDefined := d["Metadata"]; isDefined {
		page.Metadata = obj
	}
	if obj, isDefined := d["PieceInfo"]; isDefined {
		page.PieceInfo = obj
	}
	if obj, isDefined := d["StructParents"]; isDefined {
		page.StructParents = obj
	}
	if obj, isDefined := d["ID"]; isDefined {
		page.ID = obj
	}
	if obj, isDefined := d["PZ"]; isDefined {
		page.PZ = obj
	}
	if obj, isDefined := d["SeparationInfo"]; isDefined {
		page.SeparationInfo = obj
	}
	if obj, isDefined := d["Tabs"]; isDefined {
		page.Tabs = obj
	}
	if obj, isDefined := d["TemplateInstantiated"]; isDefined {
		page.TemplateInstantiated = obj
	}
	if obj, isDefined := d["PresSteps"]; isDefined {
		page.PresSteps = obj
	}
	if obj, isDefined := d["UserUnit"]; isDefined {
		page.UserUnit = obj
	}
	if obj, isDefined := d["VP"]; isDefined {
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
	annotsObj, hasAnnots := (*d)["Annots"]
	if !hasAnnots {
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

		if obj, hasMediaBox := (*dict)["MediaBox"]; hasMediaBox {
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

		node = (*dict)["Parent"]
	}

	return nil, errors.New("Media box not defined")
}

// Get the inheritable resources, either from the page or
// or a higher up page/pages struct.
func (this *PdfPage) GetResources() (*PdfPageResources, error) {
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

		if obj, hasResources := (*dict)["Resources"]; hasResources {
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
		node = (*dict)["Parent"]
	}

	return nil, errors.New("Resources not defined (required)")
}

// Convert the Page to a PDF object dictionary.
func (this *PdfPage) GetPageDict() *PdfObjectDictionary {
	p := this.pageDict
	(*p)["Type"] = MakeName("Page")
	(*p)["Parent"] = this.Parent

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
	//p.SetIfNotNil("AnnotsOLD", this.Annots)
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
			arr = append(arr, annot.GetContext().ToPdfObject())
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
	if this.Resources == nil {
		// Get the resources (is required, should be there..), can be defined by parent object.
		resources, err := this.GetResources()
		if err != nil {
			return err
		}
		if resources != nil {
			this.Resources = resources
		} else {
			this.Resources = NewPdfPageResources()
		}
	}

	var xresDict *PdfObjectDictionary
	if this.Resources.XObject == nil {
		xresDict = &PdfObjectDictionary{}
		this.Resources.XObject = xresDict
	} else {
		var ok bool
		xresDict, ok = (this.Resources.XObject).(*PdfObjectDictionary)
		if !ok {
			return errors.New("Invalid xres dict type")
		}

	}
	// Make a stream object container.
	(*xresDict)[name] = ximg.ToPdfObject()

	return nil
}

// Check if has image resource by name.
func (this *PdfPage) HasImageResource(name PdfObjectName) bool {
	resources, err := this.GetResources()
	if err != nil {
		return false
	}

	xresDict, has := resources.XObject.(*PdfObjectDictionary)
	if !has {
		return false
	}

	if _, has := (*xresDict)[name]; has {
		return true
	} else {
		return false
	}
}

// Add a graphics state to the XObject resources.
func (this *PdfPage) AddExtGState(name PdfObjectName, egs *PdfObjectDictionary) {
	if this.Resources == nil {
		//this.Resources = &PdfPageResources{}
		this.Resources = NewPdfPageResources()
	}

	if this.Resources.ExtGState == nil {
		this.Resources.ExtGState = &PdfObjectDictionary{}
	}

	egsDict := this.Resources.ExtGState.(*PdfObjectDictionary)
	(*egsDict)[name] = egs
}

// Add a font dictionary to the Font resources.
func (this *PdfPage) AddFont(name PdfObjectName, font *PdfObjectDictionary) {
	if this.Resources == nil {
		//this.Resources = &PdfPageResources{}
		this.Resources = NewPdfPageResources()
	}

	if this.Resources.Font == nil {
		this.Resources.Font = &PdfObjectDictionary{}
	}

	fontDict := this.Resources.Font.(*PdfObjectDictionary)
	(*fontDict)[name] = font
}

type WatermarkImageOptions struct {
	Alpha               float64
	FitToWidth          bool
	PreserveAspectRatio bool
}

// Add a watermark to the page.
func (this *PdfPage) AddWatermarkImage(ximg *XObjectImage, opt WatermarkImageOptions) error {
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

	imgName := PdfObjectName("Imw0")
	this.AddImageResource(imgName, ximg)

	gs0 := PdfObjectDictionary{}
	gs0["BM"] = MakeName("Normal")
	gs0["CA"] = MakeFloat(opt.Alpha)
	gs0["ca"] = MakeFloat(opt.Alpha)
	this.AddExtGState("GS0", &gs0)

	contentStr := fmt.Sprintf("q\n"+
		"/GS0 gs\n"+
		"%.0f 0 0 %.0f %.4f %.4f cm\n"+
		"/%s Do\n"+
		"Q", wWidth, wHeight, xOffset, yOffset, imgName)
	this.AddContentStreamByString(contentStr)

	return nil
}

// Add content stream by string.  Puts the content string into a stream
// object and points the content stream towards it.
func (this *PdfPage) AddContentStreamByString(contentStr string) {
	stream := PdfObjectStream{}

	sDict := PdfObjectDictionary{}
	stream.PdfObjectDictionary = &sDict

	sDict["Length"] = MakeInteger(int64(len(contentStr)))
	stream.Stream = []byte(contentStr)

	if this.Contents == nil {
		// If not set, place it directly.
		this.Contents = &stream
	} else if contArray, isArray := this.Contents.(*PdfObjectArray); isArray {
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
		sDict := encoder.MakeEncodingDict()

		encoded, err := encoder.EncodeBytes([]byte(cStream))
		if err != nil {
			return err
		}

		(*sDict)["Length"] = MakeInteger(int64(len(encoded)))

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
	return "", fmt.Errorf("Invalid content stream object holder (%T)", cstreamObj)
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

// Page resources model.
// Implements PdfModel.
type PdfPageResources struct {
	ExtGState  PdfObject
	ColorSpace PdfObject
	Pattern    PdfObject
	Shading    PdfObject
	XObject    PdfObject
	Font       PdfObject
	ProcSet    PdfObject
	// Primitive.
	primitive *PdfObjectDictionary
}

func NewPdfPageResources() *PdfPageResources {
	r := &PdfPageResources{}
	r.primitive = &PdfObjectDictionary{}
	return r
}

func NewPdfPageResourcesFromDict(dict *PdfObjectDictionary) (*PdfPageResources, error) {
	r := NewPdfPageResources()

	if obj, isDefined := (*dict)["ExtGState"]; isDefined {
		r.ExtGState = obj
	}
	if obj, isDefined := (*dict)["ColorSpace"]; isDefined {
		r.ColorSpace = obj
	}
	if obj, isDefined := (*dict)["Pattern"]; isDefined {
		r.Pattern = obj
	}
	if obj, isDefined := (*dict)["Shading"]; isDefined {
		r.Shading = obj
	}
	if obj, isDefined := (*dict)["XObject"]; isDefined {
		r.XObject = obj
	}
	if obj, isDefined := (*dict)["Font"]; isDefined {
		r.Font = obj
	}
	if obj, isDefined := (*dict)["ProcSet"]; isDefined {
		r.ProcSet = obj
	}

	return r, nil
}

func (r *PdfPageResources) GetContainingPdfObject() PdfObject {
	return r.primitive
}

func (r *PdfPageResources) ToPdfObject() PdfObject {
	d := r.primitive
	d.SetIfNotNil("ExtGState", r.ExtGState)
	d.SetIfNotNil("ColorSpace", r.ColorSpace)
	d.SetIfNotNil("Pattern", r.Pattern)
	d.SetIfNotNil("Shading", r.Shading)
	d.SetIfNotNil("XObject", r.XObject)
	d.SetIfNotNil("Font", r.Font)
	d.SetIfNotNil("ProcSet", r.ProcSet)
	return d
}

// XObjectImage (Table 89 in 8.9.5.1).
// Implements PdfModel interface.
type XObjectImage struct {
	Width            *int64
	Height           *int64
	ColorSpace       PdfObject
	BitsPerComponent *int64
	Intent           PdfObject
	ImageMask        PdfObject
	Mask             PdfObject
	Decode           PdfObject
	Interpolate      PdfObject
	Alternatives     PdfObject
	SMask            PdfObject
	SMaskInData      PdfObject
	Name             PdfObject
	StructParent     PdfObject
	ID               PdfObject
	OPI              PdfObject
	Metadata         PdfObject
	OC               PdfObject
	Stream           []byte
	// Primitive
	primitive *PdfObjectStream
}

func NewXObjectImage() *XObjectImage {
	xobj := &XObjectImage{}
	stream := &PdfObjectStream{}
	stream.PdfObjectDictionary = &PdfObjectDictionary{}
	xobj.primitive = stream
	return xobj
}

// Creates a new XObject Image from an image object with default
// options.
func NewXObjectImageFromImage(name PdfObjectName, img *Image) (*XObjectImage, error) {
	xobj := NewXObjectImage()

	xobj.Name = &name
	xobj.Stream = img.Data.Bytes()

	// Width and height.
	imWidth := img.Width
	imHeight := img.Height
	xobj.Width = &imWidth
	xobj.Height = &imHeight

	// Bits.
	bitDepth := int64(8)
	xobj.BitsPerComponent = &bitDepth

	xobj.ColorSpace = MakeName("DeviceRGB")

	return xobj, nil
}

// Build the image xobject from a stream object.
func NewXObjectImageFromStream(stream PdfObjectStream) (*XObjectImage, error) {
	img := NewXObjectImage()

	dict := *(stream.PdfObjectDictionary)

	if obj, isDefined := dict["Width"]; isDefined {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image width object")
		}
		iVal := int64(*iObj)
		img.Width = &iVal
	}

	if obj, isDefined := dict["Height"]; isDefined {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image height object")
		}
		iVal := int64(*iObj)
		img.Height = &iVal
	}

	if obj, isDefined := dict["ColorSpace"]; isDefined {
		img.ColorSpace = obj
	}

	if obj, isDefined := dict["BitsPerComponent"]; isDefined {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid image height object")
		}
		iVal := int64(*iObj)
		img.BitsPerComponent = &iVal
	}

	if obj, isDefined := dict["Intent"]; isDefined {
		img.Intent = obj
	}
	if obj, isDefined := dict["ImageMask"]; isDefined {
		img.ImageMask = obj
	}
	if obj, isDefined := dict["Mask"]; isDefined {
		img.Mask = obj
	}
	if obj, isDefined := dict["Decode"]; isDefined {
		img.Decode = obj
	}
	if obj, isDefined := dict["Interpolate"]; isDefined {
		img.Interpolate = obj
	}
	if obj, isDefined := dict["Alternatives"]; isDefined {
		img.Alternatives = obj
	}
	if obj, isDefined := dict["SMask"]; isDefined {
		img.SMask = obj
	}
	if obj, isDefined := dict["SMaskInData"]; isDefined {
		img.SMaskInData = obj
	}
	if obj, isDefined := dict["Name"]; isDefined {
		img.Name = obj
	}
	if obj, isDefined := dict["StructParent"]; isDefined {
		img.StructParent = obj
	}
	if obj, isDefined := dict["ID"]; isDefined {
		img.ID = obj
	}
	if obj, isDefined := dict["OPI"]; isDefined {
		img.OPI = obj
	}
	if obj, isDefined := dict["Metadata"]; isDefined {
		img.Metadata = obj
	}
	if obj, isDefined := dict["OC"]; isDefined {
		img.OC = obj
	}

	img.Stream = stream.Stream

	return img, nil
}

func (ximg *XObjectImage) GetContainingPdfObject() PdfObject {
	return ximg.primitive
}

// Return a stream object.
func (ximg *XObjectImage) ToPdfObject() PdfObject {
	stream := ximg.primitive
	stream.Stream = ximg.Stream

	dict := stream.PdfObjectDictionary

	dict.Set("Type", MakeName("XObject"))
	dict.Set("Subtype", MakeName("Image"))
	dict.Set("Width", MakeInteger(*(ximg.Width)))
	dict.Set("Height", MakeInteger(*(ximg.Height)))
	dict.Set("Filter", MakeName("DCTDecode"))

	if ximg.BitsPerComponent != nil {
		dict.Set("BitsPerComponent", MakeInteger(*(ximg.BitsPerComponent)))
	}

	dict.SetIfNotNil("ColorSpace", ximg.ColorSpace)
	dict.SetIfNotNil("Intent", ximg.Intent)
	dict.SetIfNotNil("ImageMask", ximg.ImageMask)
	dict.SetIfNotNil("Mask", ximg.Mask)
	dict.SetIfNotNil("Decode", ximg.Decode)
	dict.SetIfNotNil("Interpolate", ximg.Interpolate)
	dict.SetIfNotNil("Alternatives", ximg.Alternatives)
	dict.SetIfNotNil("SMask", ximg.SMask)
	dict.SetIfNotNil("SMaskInData", ximg.SMaskInData)
	dict.SetIfNotNil("Name", ximg.Name)
	dict.SetIfNotNil("StructParent", ximg.StructParent)
	dict.SetIfNotNil("ID", ximg.ID)
	dict.SetIfNotNil("OPI", ximg.OPI)
	dict.SetIfNotNil("Metadata", ximg.Metadata)
	dict.SetIfNotNil("OC", ximg.OC)

	dict.Set("Length", MakeInteger(int64(len(ximg.Stream))))
	stream.Stream = ximg.Stream

	return stream
}
