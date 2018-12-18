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

// PdfOutline represents a PDF outline dictionary (Table 152 - p. 376).
type PdfOutline struct {
	PdfOutlineTreeNode
	Parent *PdfOutlineTreeNode
	Count  *int64

	primitive *PdfIndirectObject
}

// PdfOutlineItem represents an outline item dictionary (Table 153 - pp. 376 - 377).
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

// NewPdfOutline returns an initialized PdfOutline.
func NewPdfOutline() *PdfOutline {
	outline := &PdfOutline{}

	container := &PdfIndirectObject{}
	container.PdfObject = MakeDict()

	outline.primitive = container

	return outline
}

// NewPdfOutlineTree returns an initialized PdfOutline tree.
func NewPdfOutlineTree() *PdfOutline {
	outlineTree := NewPdfOutline()
	outlineTree.context = &outlineTree
	return outlineTree
}

// NewPdfOutlineItem returns an initialized PdfOutlineItem.
func NewPdfOutlineItem() *PdfOutlineItem {
	outlineItem := &PdfOutlineItem{}

	container := &PdfIndirectObject{}
	container.PdfObject = MakeDict()

	outlineItem.primitive = container
	return outlineItem
}

// NewOutlineBookmark returns an initialized PdfOutlineItem for a given bookmark title and page.
func NewOutlineBookmark(title string, page *PdfIndirectObject) *PdfOutlineItem {
	bookmark := PdfOutlineItem{}
	bookmark.context = &bookmark

	bookmark.Title = MakeString(title)

	destArray := MakeArray()
	destArray.Append(page)
	destArray.Append(MakeName("Fit"))
	bookmark.Dest = destArray

	return &bookmark
}

// Does not traverse the tree.
func newPdfOutlineFromIndirectObject(container *PdfIndirectObject) (*PdfOutline, error) {
	dict, isDict := container.PdfObject.(*PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("outline object not a dictionary")
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
		count, err := GetNumberAsInt64(obj)
		if err != nil {
			return nil, err
		}
		outline.Count = &count
	}

	return &outline, nil
}

// Does not traverse the tree.
func (r *PdfReader) newPdfOutlineItemFromIndirectObject(container *PdfIndirectObject) (*PdfOutlineItem, error) {
	dict, isDict := container.PdfObject.(*PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("outline object not a dictionary")
	}

	item := PdfOutlineItem{}
	item.primitive = container
	item.context = &item

	// Title (required).
	obj := dict.Get("Title")
	if obj == nil {
		return nil, fmt.Errorf("missing Title from Outline Item (required)")
	}
	obj, err := r.traceToObject(obj)
	if err != nil {
		return nil, err
	}
	title, ok := TraceToDirectObject(obj).(*PdfObjectString)
	if !ok {
		return nil, fmt.Errorf("title not a string (%T)", obj)
	}
	item.Title = title

	// Count (optional).
	if obj := dict.Get("Count"); obj != nil {
		countVal, ok := obj.(*PdfObjectInteger)
		if !ok {
			return nil, fmt.Errorf("count not an integer (%T)", obj)
		}
		count := int64(*countVal)
		item.Count = &count
	}

	// Other keys.
	if obj := dict.Get("Dest"); obj != nil {
		item.Dest, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		err := r.traverseObjectData(item.Dest)
		if err != nil {
			return nil, err
		}
	}
	if obj := dict.Get("A"); obj != nil {
		item.A, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		err := r.traverseObjectData(item.A)
		if err != nil {
			return nil, err
		}
	}
	if obj := dict.Get("SE"); obj != nil {
		// TODO: To add structure element support.
		// Currently not supporting structure elements.
		item.SE = nil
		/*
			item.SE, err = r.traceToObject(obj)
			if err != nil {
				return nil, err
			}
		*/
	}
	if obj := dict.Get("C"); obj != nil {
		item.C, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
	}
	if obj := dict.Get("F"); obj != nil {
		item.F, err = r.traceToObject(obj)
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

func (n *PdfOutlineTreeNode) GetContainingPdfObject() PdfObject {
	return n.getOuter().GetContainingPdfObject()
}

func (n *PdfOutlineTreeNode) ToPdfObject() PdfObject {
	return n.getOuter().ToPdfObject()
}

func (o *PdfOutline) GetContainingPdfObject() PdfObject {
	return o.primitive
}

// ToPdfObject recursively builds the Outline tree PDF object.
func (o *PdfOutline) ToPdfObject() PdfObject {
	container := o.primitive
	dict := container.PdfObject.(*PdfObjectDictionary)

	dict.Set("Type", MakeName("Outlines"))

	if o.First != nil {
		dict.Set("First", o.First.ToPdfObject())
	}

	if o.Last != nil {
		dict.Set("Last", o.Last.getOuter().GetContainingPdfObject())
		//PdfObjectConverterCache[o.Last.getOuter()]
	}

	if o.Parent != nil {
		dict.Set("Parent", o.Parent.getOuter().GetContainingPdfObject())
	}

	return container
}

func (oi *PdfOutlineItem) GetContainingPdfObject() PdfObject {
	return oi.primitive
}

// ToPdfObject recursively builds the Outline tree PDF object.
func (oi *PdfOutlineItem) ToPdfObject() PdfObject {
	container := oi.primitive
	dict := container.PdfObject.(*PdfObjectDictionary)

	dict.Set("Title", oi.Title)
	if oi.A != nil {
		dict.Set("A", oi.A)
	}
	if obj := dict.Get("SE"); obj != nil {
		// TODO: Currently not supporting structure element hierarchy.
		// Remove it.
		dict.Remove("SE")
		//	delete(*dict, "SE")
	}
	/*
		if oi.SE != nil {
			(*dict)["SE"] = oi.SE
		}
	*/
	if oi.C != nil {
		dict.Set("C", oi.C)
	}
	if oi.Dest != nil {
		dict.Set("Dest", oi.Dest)
	}
	if oi.F != nil {
		dict.Set("F", oi.F)
	}
	if oi.Count != nil {
		dict.Set("Count", MakeInteger(*oi.Count))
	}
	if oi.Next != nil {
		dict.Set("Next", oi.Next.ToPdfObject())
	}
	if oi.First != nil {
		dict.Set("First", oi.First.ToPdfObject())
	}
	if oi.Prev != nil {
		dict.Set("Prev", oi.Prev.getOuter().GetContainingPdfObject())
		//PdfObjectConverterCache[oi.Prev.getOuter()]
	}
	if oi.Last != nil {
		dict.Set("Last", oi.Last.getOuter().GetContainingPdfObject())
		// PdfObjectConverterCache[oi.Last.getOuter()]
	}
	if oi.Parent != nil {
		dict.Set("Parent", oi.Parent.getOuter().GetContainingPdfObject())
		//PdfObjectConverterCache[oi.Parent.getOuter()]
	}

	return container
}
