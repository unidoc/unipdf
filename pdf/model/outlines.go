/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"fmt"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

type PdfOutlineTreeNode struct {
	context interface{} // Allow accessing outer structure.
	First   *PdfOutlineTreeNode
	Last    *PdfOutlineTreeNode
}

// PDF outline dictionary (Table 152 - p. 376).
type PdfOutline struct {
	PdfOutlineTreeNode
	Parent *PdfOutlineTreeNode
	Count  *int64

	primitive *PdfIndirectObject
}

// Pdf outline item dictionary (Table 153 - pp. 376 - 377).
type PdfOutlineItem struct {
	PdfOutlineTreeNode
	Title  *PdfObjectString
	Parent *PdfOutlineTreeNode
	Prev   *PdfOutlineTreeNode
	Next   *PdfOutlineTreeNode
	Count  *int64
	Dest   PdfObject
	A      PdfObject
	SE     PdfObject
	C      PdfObject
	F      PdfObject

	primitive *PdfIndirectObject
}

func NewPdfOutline() *PdfOutline {
	outline := &PdfOutline{}

	container := &PdfIndirectObject{}
	container.PdfObject = MakeDict()

	outline.primitive = container

	return outline
}

func NewPdfOutlineTree() *PdfOutline {
	outlineTree := NewPdfOutline()
	outlineTree.context = &outlineTree
	return outlineTree
}

func NewPdfOutlineItem() *PdfOutlineItem {
	outlineItem := &PdfOutlineItem{}

	container := &PdfIndirectObject{}
	container.PdfObject = MakeDict()

	outlineItem.primitive = container
	return outlineItem
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
func newPdfOutlineFromIndirectObject(container *PdfIndirectObject) (*PdfOutline, error) {
	dict, isDict := container.PdfObject.(*PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("Outline object not a dictionary")
	}

	outline := PdfOutline{}
	outline.primitive = container
	outline.context = &outline

	if obj := dict.Get("Type"); obj != nil {
		typeVal, ok := obj.(*PdfObjectName)
		if ok {
			if *typeVal != "Outlines" {
				common.Log.Debug("ERROR Type != Outlines (%s)", *typeVal)
				// Should be "Outlines" if there, but some files have other types
				// Log as an error but do not quit.
				// Might be a good idea to log this kind of deviation from the standard separately.
			}
		}
	}

	if obj := dict.Get("Count"); obj != nil {
		// This should always be an integer, but in a few cases has been a float.
		count, err := getNumberAsInt64(obj)
		if err != nil {
			return nil, err
		}
		outline.Count = &count
	}

	return &outline, nil
}

// Does not traverse the tree.
func (this *PdfReader) newPdfOutlineItemFromIndirectObject(container *PdfIndirectObject) (*PdfOutlineItem, error) {
	dict, isDict := container.PdfObject.(*PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("Outline object not a dictionary")
	}

	item := PdfOutlineItem{}
	item.primitive = container
	item.context = &item

	// Title (required).
	obj := dict.Get("Title")
	if obj == nil {
		return nil, fmt.Errorf("Missing Title from Outline Item (required)")
	}
	obj, err := this.traceToObject(obj)
	if err != nil {
		return nil, err
	}
	title, ok := TraceToDirectObject(obj).(*PdfObjectString)
	if !ok {
		return nil, fmt.Errorf("Title not a string (%T)", obj)
	}
	item.Title = title

	// Count (optional).
	if obj := dict.Get("Count"); obj != nil {
		countVal, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, fmt.Errorf("Count not an integer (%T)", obj)
		}
		count := int64(*countVal)
		item.Count = &count
	}

	// Other keys.
	if obj := dict.Get("Dest"); obj != nil {
		item.Dest, err = this.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		err := this.traverseObjectData(item.Dest)
		if err != nil {
			return nil, err
		}
	}
	if obj := dict.Get("A"); obj != nil {
		item.A, err = this.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		err := this.traverseObjectData(item.A)
		if err != nil {
			return nil, err
		}
	}
	if obj := dict.Get("SE"); obj != nil {
		// XXX: To add structure element support.
		// Currently not supporting structure elements.
		item.SE = nil
		/*
			item.SE, err = this.traceToObject(obj)
			if err != nil {
				return nil, err
			}
		*/
	}
	if obj := dict.Get("C"); obj != nil {
		item.C, err = this.traceToObject(obj)
		if err != nil {
			return nil, err
		}
	}
	if obj := dict.Get("F"); obj != nil {
		item.F, err = this.traceToObject(obj)
		if err != nil {
			return nil, err
		}
	}

	return &item, nil
}

// Get the outer object of the tree node (Outline or OutlineItem).
func (n *PdfOutlineTreeNode) getOuter() PdfModel {
	if outline, isOutline := n.context.(*PdfOutline); isOutline {
		return outline
	}
	if outlineItem, isOutlineItem := n.context.(*PdfOutlineItem); isOutlineItem {
		return outlineItem
	}

	common.Log.Debug("ERROR Invalid outline tree node item") // Should never happen.
	return nil
}

func (this *PdfOutlineTreeNode) GetContainingPdfObject() PdfObject {
	return this.getOuter().GetContainingPdfObject()
}

func (this *PdfOutlineTreeNode) ToPdfObject() PdfObject {
	return this.getOuter().ToPdfObject()
}

func (this *PdfOutline) GetContainingPdfObject() PdfObject {
	return this.primitive
}

// Recursively build the Outline tree PDF object.
func (this *PdfOutline) ToPdfObject() PdfObject {
	container := this.primitive
	dict := container.PdfObject.(*PdfObjectDictionary)

	dict.Set("Type", MakeName("Outlines"))

	if this.First != nil {
		dict.Set("First", this.First.ToPdfObject())
	}

	if this.Last != nil {
		dict.Set("Last", this.Last.getOuter().GetContainingPdfObject())
		//PdfObjectConverterCache[this.Last.getOuter()]
	}

	if this.Parent != nil {
		dict.Set("Parent", this.Parent.getOuter().GetContainingPdfObject())
	}

	return container
}

func (this *PdfOutlineItem) GetContainingPdfObject() PdfObject {
	return this.primitive
}

// Outline item.
// Recursively build the Outline tree PDF object.
func (this *PdfOutlineItem) ToPdfObject() PdfObject {
	container := this.primitive
	dict := container.PdfObject.(*PdfObjectDictionary)

	dict.Set("Title", this.Title)
	if this.A != nil {
		dict.Set("A", this.A)
	}
	if obj := dict.Get("SE"); obj != nil {
		// XXX: Currently not supporting structure element hierarchy.
		// Remove it.
		dict.Remove("SE")
		//	delete(*dict, "SE")
	}
	/*
		if this.SE != nil {
			(*dict)["SE"] = this.SE
		}
	*/
	if this.C != nil {
		dict.Set("C", this.C)
	}
	if this.Dest != nil {
		dict.Set("Dest", this.Dest)
	}
	if this.F != nil {
		dict.Set("F", this.F)
	}
	if this.Count != nil {
		dict.Set("Count", MakeInteger(*this.Count))
	}
	if this.Next != nil {
		dict.Set("Next", this.Next.ToPdfObject())
	}
	if this.First != nil {
		dict.Set("First", this.First.ToPdfObject())
	}
	if this.Prev != nil {
		dict.Set("Prev", this.Prev.getOuter().GetContainingPdfObject())
		//PdfObjectConverterCache[this.Prev.getOuter()]
	}
	if this.Last != nil {
		dict.Set("Last", this.Last.getOuter().GetContainingPdfObject())
		// PdfObjectConverterCache[this.Last.getOuter()]
	}
	if this.Parent != nil {
		dict.Set("Parent", this.Parent.getOuter().GetContainingPdfObject())
		//PdfObjectConverterCache[this.Parent.getOuter()]
	}

	return container
}
