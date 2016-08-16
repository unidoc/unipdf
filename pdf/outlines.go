package pdf

import (
	"fmt"
)

type PdfOutlineTreeNode struct {
	First *PdfOutlineTreeNode
	Last  *PdfOutlineTreeNode
}

// PDF outline dictionary (Table 152 - p. 376)
type PdfOutline struct {
	PdfOutlineTreeNode
	//First *PdfOutlineItem
	//Last  *PdfOutlineItem
	Count *int64
}

// Pdf outline item dictionary (Table 153 - pp. 376 - 377)
type PdfOutlineItem struct {
	PdfOutlineTreeNode
	Title *PdfObjectString
	//First *PdfOutlineItem
	//Last  *PdfOutlineItem
	Prev  *PdfOutlineTreeNode
	Next  *PdfOutlineTreeNode
	Count *int64
	Dest  PdfObject
	A     PdfObject
	SE    PdfObject
	C     PdfObject
	F     PdfObject
}

// Destinations.  Can be explicit (defined by dest arrays - table 151)
// or named destinations through the document catalog Dests definitions.
type PdfDestination interface {
	ToPdfObject() PdfObject
}

type PdfDestinationXYZ struct {
	Page            *PdfPage
	Left, Top, Zoom *float64
}

type PdfDestinationFit struct {
	Page *PdfPage
}

// Horizontal fit (fit width), with a specified top location.
type PdfDestinationFitH struct {
	Page *PdfPage
	Top  *float64
}

// Vertical fit (fit height), with a specified horizontal left location.
type PdfDestinationFitV struct {
	Page *PdfPage
	Left *float64
}

// Rectangle fit.
type PdfDestinationFitR struct {
	Page                     *PdfPage
	Left, Bottom, Right, Top *float64
}

// Bounding box fit.
type PdfDestinationFitB struct {
	Page *PdfPage
}

type PdfDestinationFitBH struct {
	Page *PdfPage
	Top  *float64
}

type PdfDestinationFitBV struct {
	Page *PdfPage
	Left *float64
}

type PdfDestinationNamed struct {
	name *PdfObjectName
}

// Does not traverse the tree.
func newPdfOutlineFromDict(dict *PdfObjectDictionary) (*PdfOutline, error) {
	outline := PdfOutline{}

	if obj, hasType := (*dict)["Type"]; hasType {
		typeVal, ok := obj.(*PdfObjectName)
		if ok {
			if *typeVal != "Outlines" {
				return nil, fmt.Errorf("Type != Outlines (%s)", *typeVal)
			}
		}
	}

	if obj, hasCount := (*dict)["Count"]; hasCount {
		countVal, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, fmt.Errorf("Count not an integer (%T)", obj)
		}
		count := int64(*countVal)
		outline.Count = &count
	}

	return &outline, nil
}

// Does not traverse the tree.
func newPdfOutlineItemFromDict(dict *PdfObjectDictionary) (*PdfOutlineItem, error) {
	item := PdfOutlineItem{}

	// Title (required).
	obj, hasTitle := (*dict)["Title"]
	if !hasTitle {
		return nil, fmt.Errorf("Missing Title from Outline Item (required)")
	}
	title, ok := obj.(*PdfObjectString)
	if !ok {
		return nil, fmt.Errorf("Title not a string (%T)", obj)
	}
	item.Title = title

	// Count (optional).
	if obj, hasCount := (*dict)["Count"]; hasCount {
		countVal, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, fmt.Errorf("Count not an integer (%T)", obj)
		}
		count := int64(*countVal)
		item.Count = &count
	}

	// Dest.
	// A.
	// SE.
	// C.
	// F.

	return &item, nil
}

// Build the PDF destination from an explicit destination array.
// Use for parsing existing PDF structures.
/*
func NewPdfDestinationFromDestArray(arr *PdfObjectArray) (PdfDestination, error) {
	if arr == nil {
		return nil, fmt.Errorf("Nil dest array")
	}
	if len(*arr) < 2 {
		return nil, fmt.Errorf("Dest array length < 2 (%d)", len(*arr))
	}

	page, ok := (*arr)[0].(*PdfIndirectObject)
	if !ok {
		return nil, fmt.Errorf("Invalid page type (arr: %v)", arr)
	}

	dtype, ok := (*arr)[1].(*PdfObjectName)
	if !ok {
		return nil, fmt.Errorf("Invalid dest type (%v)", dtype)
	}

	// Explicit XYZ destination.
	if *dtype == "XYZ" {
		if len(*arr) != 5 {
			return nil, fmt.Errorf("XYZ arr len != 5 (%v)", arr)
		}
		dest := NewPdfDestinationXYZ(page)

		// Left.
		num, err := getNumberAsFloatOrNull(*arr[2])
		if err != nil {
			return nil, err
		}
		dest.Left = num

		// Top.
		num, err = getNumberAsFloatOrNull(arr[3])
		if err != nil {
			return nil, err
		}
		dest.Top = num

		// Zoom level.
		num, err = getNumberAsFloatOrNull(arr[4])
		if err != nil {
			return nil, err
		}
		dest.Zoom = num
		return &dest, nil
	}

	// Explicit Fit destination.
	if dtype == "Fit" {
		dest := PdfDestinationFit{}
		dest.Page = page
		if len(arr) != 2 {
			return nil, fmt.Errorf("Fit arr len != 2 (%v)", arr)
		}
		return &dest, nil
	}

	// Explicit FitH destination.
	if dtype == "FitH" {
		if len(arr) != 3 {
			return nil, fmt.Errorf("FitH arr len != 3 (%v)", arr)
		}

		dest := PdfDestinationFitH{}
		dest.Page = page

		// Top.
		num, err = getNumberAsFloatOrNull(arr[2])
		if err != nil {
			return nil, err
		}
		dest.Top = num

		return &dest, nil
	}

	// Explicit FitV destination.
	if dtype == "FitV" {
		if len(arr) != 3 {
			return nil, fmt.Errorf("FitV arr len != 3 (%v)", arr)
		}

		dest := PdfDestinationFitV{}
		dest.Page = page

		// Left.
		num, err = getNumberAsFloatOrNull(arr[2])
		if err != nil {
			return nil, err
		}
		dest.Left = num

		return &dest, nil
	}

	// Explicit FitR destination.
	if dtype == "FitR" {
		if len(arr) != 6 {
			return nil, fmt.Errorf("FitR arr len != 6 (%v)", arr)
		}
		dest := PdfDestinationFitR{}
		dest.Page = page

		// Left.
		num, err := getNumberAsFloatOrNull(arr[2])
		if err != nil {
			return nil, err
		}
		dest.Left = num

		// Bottom.
		num, err = getNumberAsFloatOrNull(arr[3])
		if err != nil {
			return nil, err
		}
		dest.Bottom = num

		// Right.
		num, err = getNumberAsFloatOrNull(arr[4])
		if err != nil {
			return nil, err
		}
		dest.Right = num

		// Top.
		num, err = getNumberAsFloatOrNull(arr[5])
		if err != nil {
			return nil, err
		}
		dest.Top = num

		return &dest, nil
	}

	// Explicit FitB destination.
	if dtype == "FitB" {
		dest := PdfDestinationFitB{}
		dest.Page = page
		if len(arr) != 2 {
			return nil, fmt.Errorf("FitB arr len != 2 (%v)", arr)
		}
		return &dest, nil
	}

	// Explicit FitBH destination.
	if dtype == "FitBH" {
		dest := PdfDestinationFitBH{}
		dest.Page = page
		if len(arr) != 3 {
			return nil, fmt.Errorf("FitB arr len != 3 (%v)", arr)
		}
		dest.Page = page

		// Top.
		num, err = getNumberAsFloatOrNull(arr[2])
		if err != nil {
			return nil, err
		}
		dest.Top = num

		return &dest, nil
	}

	// Explicit FitBV destination.
	if dtype == "FitBV" {
		dest := PdfDestinationFitBV{}
		dest.Page = page
		if len(arr) != 3 {
			return nil, fmt.Errorf("FitB arr len != 3 (%v)", arr)
		}
		dest.Page = page

		// Left.
		num, err = getNumberAsFloatOrNull(arr[2])
		if err != nil {
			return nil, err
		}
		dest.Left = num

		return &dest, nil
	}

	return nil, fmt.Errorf("Invalid destination type (%v)", dtype)

}

// New destination from a PDF object: either explicit destination (array), or name/string (named destination).
func NewPdfDestinationFromPdfObject(obj PdfObject) (error, PdfDestination) {
	if array, isArray := obj.(*PdfObjectArray); isArray {
		return NewPdfDestinationFromDestArray(arr)
	}

	// Named destination can either be a Name
	if name, isName := obj.(*PdfObjectName); isName {
		return NewPdfDestinationFromName(name)
	}

	// or a byte string.  Handle in the same fashion.
	if name, isName := obj.(*PdfObjectString); isName {
		return NewPdfDestinationFromName(PdfObjectName(name))
	}

	return errors.New("Invalid object"), nil
}
*/
func NewPdfDestinationXYZ(page *PdfPage) *PdfDestinationXYZ {
	dest := PdfDestinationXYZ{}
	dest.Page = page
	return &dest
}

// Return a float pdf object if defined, otherwise a pdf null object.
func FloatOrNull(val *float64) PdfObject {
	if val != nil {
		return MakeFloat(*val)
	} else {
		return MakeNull()
	}
}

// Return the XYZ destination in an array form.
func (dest *PdfDestinationXYZ) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("XYZ"))
	array = append(array, FloatOrNull(dest.Left))
	array = append(array, FloatOrNull(dest.Top))
	array = append(array, FloatOrNull(dest.Zoom))
	return &array
}

func (dest *PdfDestinationFit) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("Fit"))
	return &array
}

func (dest *PdfDestinationFitH) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("FitH"))
	array = append(array, FloatOrNull(dest.Top))
	return &array
}

func (dest *PdfDestinationFitV) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("FitV"))
	array = append(array, FloatOrNull(dest.Left))
	return &array
}

func (dest *PdfDestinationFitR) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("FitR"))
	array = append(array, FloatOrNull(dest.Left))
	array = append(array, FloatOrNull(dest.Bottom))
	array = append(array, FloatOrNull(dest.Right))
	array = append(array, FloatOrNull(dest.Top))
	return &array
}

func (dest *PdfDestinationFitB) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("FitB"))
	return &array
}

func (dest *PdfDestinationFitBH) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("FitBH"))
	array = append(array, FloatOrNull(dest.Top))
	return &array
}

func (dest *PdfDestinationFitBV) ToPdfObject() *PdfObjectArray {
	array := PdfObjectArray{}
	array = append(array, dest.Page.GetPageDict())
	array = append(array, MakeName("FitBV"))
	array = append(array, FloatOrNull(dest.Left))
	return &array
}

func (dest *PdfDestinationNamed) ToPdfObject() *PdfObjectName {
	name := PdfObjectName(*dest.name)
	return &name
}
