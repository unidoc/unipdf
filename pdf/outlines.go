package pdf

import (
	"fmt"

	"github.com/unidoc/unidoc/common"
)

type PdfObjectConverter interface {
	ToPdfObject() PdfObject
}

var PdfObjectConverterCache map[PdfObjectConverter]PdfObject = map[PdfObjectConverter]PdfObject{}

type PdfOutlineTreeNode struct {
	context interface{} // Allow accessing outer structure.
	First   *PdfOutlineTreeNode
	Last    *PdfOutlineTreeNode
}

// PDF outline dictionary (Table 152 - p. 376).
type PdfOutline struct {
	PdfOutlineTreeNode
	Count *int64
}

// Pdf outline item dictionary (Table 153 - pp. 376 - 377).
type PdfOutlineItem struct {
	PdfOutlineTreeNode
	Title *PdfObjectString
	Prev  *PdfOutlineTreeNode
	Next  *PdfOutlineTreeNode
	Count *int64
	Dest  PdfObject
	A     PdfObject
	SE    PdfObject
	C     PdfObject
	F     PdfObject
}

func NewPdfOutlineTree() *PdfOutline {
	outlineTree := PdfOutline{}
	outlineTree.context = &outlineTree
	return &outlineTree
}

func NewOutlineBookmark(title string, page *PdfIndirectObject) *PdfOutlineItem {
	bookmark := PdfOutlineItem{}
	bookmark.context = &bookmark

	bookmark.Title = MakeString(title)

	destArray := PdfObjectArray{}
	destArray = append(destArray, page)
	destArray = append(destArray, MakeName("Fit"))
	bookmark.Dest = &destArray

	return &bookmark
}

// Does not traverse the tree.
func newPdfOutlineFromDict(dict *PdfObjectDictionary) (*PdfOutline, error) {
	outline := PdfOutline{}
	outline.context = &outline

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
	item.context = &item

	// Title (required).
	obj, hasTitle := (*dict)["Title"]
	if !hasTitle {
		return nil, fmt.Errorf("Missing Title from Outline Item (required)")
	}
	title, ok := TraceToDirectObject(obj).(*PdfObjectString)
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

	// Other keys.
	if obj, hasKey := (*dict)["Dest"]; hasKey {
		item.Dest = obj
	}
	if obj, hasKey := (*dict)["A"]; hasKey {
		item.A = obj
	}
	if obj, hasKey := (*dict)["SE"]; hasKey {
		item.SE = obj
	}
	if obj, hasKey := (*dict)["C"]; hasKey {
		item.C = obj
	}
	if obj, hasKey := (*dict)["F"]; hasKey {
		item.F = obj
	}

	return &item, nil
}

func (this *PdfOutlineTreeNode) ToPdfObject() PdfObject {
	if outline, isOutline := this.context.(*PdfOutline); isOutline {
		return outline.ToPdfObject()
	}
	if outlineItem, isOutlineItem := this.context.(*PdfOutlineItem); isOutlineItem {
		return outlineItem.ToPdfObject()
	}

	common.Log.Error("Invalid outline tree node item") // Should never happen.
	return nil
}

// Recursively build the Outline tree PDF object.
func (this *PdfOutline) ToPdfObject() PdfObject {
	if cachedObj, isCached := PdfObjectConverterCache[this]; isCached {
		return cachedObj
	}

	outlines := PdfIndirectObject{}

	outlinesDict := PdfObjectDictionary{}
	outlinesDict[PdfObjectName("Type")] = MakeName("Outlines")

	if this.First != nil {
		outlinesDict[PdfObjectName("First")] = this.First.ToPdfObject()
	}

	if this.Last != nil {
		outlinesDict[PdfObjectName("Last")] = this.Last.ToPdfObject()
	}

	outlines.PdfObject = &outlinesDict

	PdfObjectConverterCache[this] = &outlines

	return &outlines
}

// Outline item.
// Recursively build the Outline tree PDF object.
func (this *PdfOutlineItem) ToPdfObject() PdfObject {
	if cachedObj, isCached := PdfObjectConverterCache[this]; isCached {
		return cachedObj
	}

	container := PdfIndirectObject{}

	dict := PdfObjectDictionary{}
	dict["Title"] = this.Title
	if this.A != nil {
		dict["A"] = this.A
	}
	if this.C != nil {
		dict["C"] = this.C
	}
	if this.Dest != nil {
		dict["Dest"] = this.Dest
	}
	if this.F != nil {
		dict["F"] = this.F
	}
	if this.Count != nil {
		dict["Count"] = MakeInteger(*this.Count)
	}

	if this.Next != nil {
		dict["Next"] = this.Next.ToPdfObject()
	}
	if this.First != nil {
		dict["First"] = this.First.ToPdfObject()
	}
	if this.Prev != nil {
		dict["Prev"] = this.Prev.ToPdfObject()
	}
	if this.Last != nil {
		dict["Last"] = this.Last.ToPdfObject()
	}

	container.PdfObject = &dict
	PdfObjectConverterCache[this] = &container

	return &container
}
