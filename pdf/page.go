/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

//
// Allow higher level manipulation of PDF files and pages.
// This can be continously expanded to support more and more features.
// Generic handling can be done by defining elements as PdfObject which
// can later be replaced and fully defined.
//

package pdf

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type PdfRectangle struct {
	Llx float64 // Lower left corner (ll).
	Lly float64
	Urx float64 // Upper right corner (ur).
	Ury float64
}

func getNumberAsFloat(obj PdfObject) (float64, error) {
	if fObj, ok := obj.(*PdfObjectFloat); ok {
		return float64(*fObj), nil
	}

	if iObj, ok := obj.(*PdfObjectInteger); ok {
		return float64(*iObj), nil
	}

	return 0, errors.New("Not a number")
}

// Create a PDF rectangle object based on an input array of 4 integers.
// Defining the lower left (LL) and upper right (UR) corners with
// floating point numbers.
func NewPdfRectangle(arr PdfObjectArray) (*PdfRectangle, error) {
	rect := PdfRectangle{}
	if len(arr) != 4 {
		return nil, errors.New("Invalid rectangle array, len != 4")
	}

	var err error
	rect.Llx, err = getNumberAsFloat(arr[0])
	if err != nil {
		return nil, err
	}

	rect.Lly, err = getNumberAsFloat(arr[1])
	if err != nil {
		return nil, err
	}

	rect.Urx, err = getNumberAsFloat(arr[2])
	if err != nil {
		return nil, err
	}

	rect.Ury, err = getNumberAsFloat(arr[3])
	if err != nil {
		return nil, err
	}

	return &rect, nil
}

// Convert to a PDF object.
func (rect *PdfRectangle) ToPdfObject() PdfObject {
	arr := PdfObjectArray{}
	arr = append(arr, makeFloat(rect.Llx))
	arr = append(arr, makeFloat(rect.Lly))
	arr = append(arr, makeFloat(rect.Urx))
	arr = append(arr, makeFloat(rect.Ury))
	return &arr
}

// A date is a PDF string of the form:
// (D:YYYYMMDDHHmmSSOHH'mm)
type PdfDate struct {
	year          int64 // YYYY
	month         int64 // MM (01-12)
	day           int64 // DD (01-31)
	hour          int64 // HH (00-23)
	minute        int64 // mm (00-59)
	second        int64 // SS (00-59)
	utOffsetSign  byte  // O ('+' / '-' / 'Z')
	utOffsetHours int64 // HH' (00-23 followed by ')
	utOffsetMins  int64 // mm (00-59)
}

var reDate = regexp.MustCompile(`\s*D\s*:\s*(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})([+-Z])(\d{2})'(\d{2})`)

// Make a new PdfDate object from a PDF date string (see 7.9.4 Dates).
// format: "D: YYYYMMDDHHmmSSOHH'mm"
func NewPdfDate(dateStr string) (PdfDate, error) {
	d := PdfDate{}

	matches := reDate.FindAllStringSubmatch(dateStr, 1)
	if len(matches) < 1 {
		return d, errors.New("Invalid date string")
	}
	if len(matches[0]) != 10 {
		return d, errors.New("Invalid regexp group match length != 10")
	}

	// No need to handle err from ParseInt, as pre-validated via regexp.
	d.year, _ = strconv.ParseInt(matches[0][1], 10, 32)
	d.month, _ = strconv.ParseInt(matches[0][2], 10, 32)
	d.day, _ = strconv.ParseInt(matches[0][3], 10, 32)
	d.hour, _ = strconv.ParseInt(matches[0][4], 10, 32)
	d.minute, _ = strconv.ParseInt(matches[0][5], 10, 32)
	d.second, _ = strconv.ParseInt(matches[0][6], 10, 32)
	d.utOffsetSign = matches[0][7][0]
	d.utOffsetHours, _ = strconv.ParseInt(matches[0][8], 10, 32)
	d.utOffsetMins, _ = strconv.ParseInt(matches[0][9], 10, 32)

	return d, nil
}

// Convert to a PDF string object.
func (date *PdfDate) ToPdfObject() PdfObject {
	str := fmt.Sprintf("D:%.4d%.2d%.2d%.2d%.2d%.2d%c%.2d'%.2d'",
		date.year, date.month, date.day, date.hour, date.minute, date.second,
		date.utOffsetSign, date.utOffsetHours, date.utOffsetMins)
	pdfStr := PdfObjectString(str)
	return &pdfStr
}

// PDF page object (7.7.3.3 - Table 30).
type PdfPage struct {
	Parent               PdfObject
	LastModified         *PdfDate
	Resources            PdfObject
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
	Annots               PdfObject
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
}

// Build a PdfPage based on the underlying dictionary.
func NewPdfPage(p PdfObjectDictionary) (*PdfPage, error) {
	page := PdfPage{}

	pType, ok := p["Type"].(*PdfObjectName)
	if !ok {
		return nil, errors.New("Missing/Invalid Page dictionary Type")
	}
	if *pType != "Page" {
		return nil, errors.New("Page dictionary Type != Page")
	}

	if obj, isDefined := p["Parent"]; isDefined {
		page.Parent = obj
	}

	if obj, isDefined := p["LastModified"]; isDefined {
		strObj, ok := obj.(*PdfObjectString)
		if !ok {
			return nil, errors.New("Page dictionary LastModified != string")
		}
		lastmod, err := NewPdfDate(string(*strObj))
		if err != nil {
			return nil, err
		}
		page.LastModified = &lastmod
	}

	if obj, isDefined := p["Resources"]; isDefined {
		page.Resources = obj
	}

	if obj, isDefined := p["MediaBox"]; isDefined {
		boxArr, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page MediaBox not an array")
		}
		var err error
		page.MediaBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj, isDefined := p["CropBox"]; isDefined {
		boxArr, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page CropBox not an array")
		}
		var err error
		page.CropBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj, isDefined := p["BleedBox"]; isDefined {
		boxArr, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page BleedBox not an array")
		}
		var err error
		page.BleedBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj, isDefined := p["TrimBox"]; isDefined {
		boxArr, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page TrimBox not an array")
		}
		var err error
		page.TrimBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj, isDefined := p["ArtBox"]; isDefined {
		boxArr, ok := obj.(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Page ArtBox not an array")
		}
		var err error
		page.ArtBox, err = NewPdfRectangle(*boxArr)
		if err != nil {
			return nil, err
		}
	}
	if obj, isDefined := p["BoxColorInfo"]; isDefined {
		page.BoxColorInfo = obj
	}
	if obj, isDefined := p["Contents"]; isDefined {
		page.Contents = obj
	}
	if obj, isDefined := p["Rotate"]; isDefined {
		iObj, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid Page Rotate object")
		}
		iVal := int64(*iObj)
		page.Rotate = &iVal
	}
	if obj, isDefined := p["Group"]; isDefined {
		page.Group = obj
	}
	if obj, isDefined := p["Thumb"]; isDefined {
		page.Thumb = obj
	}
	if obj, isDefined := p["B"]; isDefined {
		page.B = obj
	}
	if obj, isDefined := p["Dur"]; isDefined {
		page.Dur = obj
	}
	if obj, isDefined := p["Trans"]; isDefined {
		page.Trans = obj
	}
	if obj, isDefined := p["Annots"]; isDefined {
		page.Annots = obj
	}
	if obj, isDefined := p["AA"]; isDefined {
		page.AA = obj
	}
	if obj, isDefined := p["Metadata"]; isDefined {
		page.Metadata = obj
	}
	if obj, isDefined := p["PieceInfo"]; isDefined {
		page.PieceInfo = obj
	}
	if obj, isDefined := p["StructParents"]; isDefined {
		page.StructParents = obj
	}
	if obj, isDefined := p["ID"]; isDefined {
		page.ID = obj
	}
	if obj, isDefined := p["PZ"]; isDefined {
		page.PZ = obj
	}
	if obj, isDefined := p["SeparationInfo"]; isDefined {
		page.SeparationInfo = obj
	}
	if obj, isDefined := p["Tabs"]; isDefined {
		page.Tabs = obj
	}
	if obj, isDefined := p["TemplateInstantiated"]; isDefined {
		page.TemplateInstantiated = obj
	}
	if obj, isDefined := p["PresSteps"]; isDefined {
		page.PresSteps = obj
	}
	if obj, isDefined := p["UserUnit"]; isDefined {
		page.UserUnit = obj
	}
	if obj, isDefined := p["VP"]; isDefined {
		page.VP = obj
	}

	return &page, nil
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

// Convert the Page to a PDF object dictionary.
func (this *PdfPage) GetPageDict() *PdfObjectDictionary {
	p := &PdfObjectDictionary{}
	(*p)["Type"] = makeName("Page")
	(*p)["Parent"] = this.Parent

	if this.LastModified != nil {
		p.Set("LastModified", this.LastModified.ToPdfObject())
	}
	p.SetIfNotNil("Resources", this.Resources)
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
		p.Set("Rotate", makeInteger(*this.Rotate))
	}

	p.SetIfNotNil("Group", this.Group)
	p.SetIfNotNil("Thumb", this.Thumb)
	p.SetIfNotNil("B", this.B)
	p.SetIfNotNil("Dur", this.Dur)
	p.SetIfNotNil("Trans", this.Trans)
	p.SetIfNotNil("Annots", this.Annots)
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

	return p
}
