/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"fmt"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// PdfOutlineTreeNode contains common fields used by the outline and outline
// item objects.
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

	primitive *core.PdfIndirectObject
}

// PdfOutlineItem represents an outline item dictionary (Table 153 - pp. 376 - 377).
type PdfOutlineItem struct {
	PdfOutlineTreeNode
	Title  *core.PdfObjectString
	Parent *PdfOutlineTreeNode
	Prev   *PdfOutlineTreeNode
	Next   *PdfOutlineTreeNode
	Count  *int64
	Dest   core.PdfObject
	A      core.PdfObject
	SE     core.PdfObject
	C      core.PdfObject
	F      core.PdfObject

	primitive *core.PdfIndirectObject
}

// NewPdfOutline returns an initialized PdfOutline.
func NewPdfOutline() *PdfOutline {
	outline := &PdfOutline{
		primitive: core.MakeIndirectObject(core.MakeDict()),
	}

	outline.context = outline
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
	outlineItem := &PdfOutlineItem{
		primitive: core.MakeIndirectObject(core.MakeDict()),
	}

	outlineItem.context = outlineItem
	return outlineItem
}

// NewOutlineBookmark returns an initialized PdfOutlineItem for a given bookmark title and page.
func NewOutlineBookmark(title string, page *core.PdfIndirectObject) *PdfOutlineItem {
	bookmark := PdfOutlineItem{}
	bookmark.context = &bookmark

	bookmark.Title = core.MakeString(title)

	destArray := core.MakeArray()
	destArray.Append(page)
	destArray.Append(core.MakeName("Fit"))
	bookmark.Dest = destArray

	return &bookmark
}

// Does not traverse the tree.
func newPdfOutlineFromIndirectObject(container *core.PdfIndirectObject) (*PdfOutline, error) {
	dict, isDict := container.PdfObject.(*core.PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("outline object not a dictionary")
	}

	outline := NewPdfOutline()

	if obj := dict.Get("Type"); obj != nil {
		typeVal, ok := obj.(*core.PdfObjectName)
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
		count, err := core.GetNumberAsInt64(obj)
		if err != nil {
			return nil, err
		}
		outline.Count = &count
	}

	return outline, nil
}

// Does not traverse the tree.
func (r *PdfReader) newPdfOutlineItemFromIndirectObject(container *core.PdfIndirectObject) (*PdfOutlineItem, error) {
	dict, isDict := container.PdfObject.(*core.PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("outline object not a dictionary")
	}

	item := NewPdfOutlineItem()

	// Title (required).
	obj := dict.Get("Title")
	if obj == nil {
		return nil, fmt.Errorf("missing Title from Outline Item (required)")
	}
	title, ok := core.GetString(obj)
	if !ok {
		return nil, fmt.Errorf("title not a string (%T)", obj)
	}
	item.Title = title

	// Count (optional).
	if obj := dict.Get("Count"); obj != nil {
		countVal, ok := obj.(*core.PdfObjectInteger)
		if !ok {
			return nil, fmt.Errorf("count not an integer (%T)", obj)
		}
		count := int64(*countVal)
		item.Count = &count
	}

	// Other keys.
	if obj := dict.Get("Dest"); obj != nil {
		item.Dest = core.ResolveReference(obj)
		if !r.isLazy {
			err := r.traverseObjectData(item.Dest)
			if err != nil {
				return nil, err
			}
		}
	}
	if obj := dict.Get("A"); obj != nil {
		item.A = core.ResolveReference(obj)
		if !r.isLazy {
			err := r.traverseObjectData(item.A)
			if err != nil {
				return nil, err
			}
		}
	}
	if obj := dict.Get("SE"); obj != nil {
		// TODO: To add structure element support.
		// Currently not supporting structure elements.
		item.SE = nil
	}
	if obj := dict.Get("C"); obj != nil {
		item.C = core.ResolveReference(obj)
	}
	if obj := dict.Get("F"); obj != nil {
		item.F = core.ResolveReference(obj)
	}

	return item, nil
}

// GetContext returns the context of the outline tree node, which is either a
// *PdfOutline or a *PdfOutlineItem. The method returns nil for uninitialized
// tree nodes.
func (n *PdfOutlineTreeNode) GetContext() PdfModel {
	if outline, isOutline := n.context.(*PdfOutline); isOutline {
		return outline
	}
	if outlineItem, isOutlineItem := n.context.(*PdfOutlineItem); isOutlineItem {
		return outlineItem
	}

	common.Log.Debug("ERROR Invalid outline tree node item") // Should never happen.
	return nil
}

// GetContainingPdfObject returns the container of the outline tree node (indirect object).
func (n *PdfOutlineTreeNode) GetContainingPdfObject() core.PdfObject {
	return n.GetContext().GetContainingPdfObject()
}

// ToPdfObject returns the PDF representation of the outline tree node.
func (n *PdfOutlineTreeNode) ToPdfObject() core.PdfObject {
	return n.GetContext().ToPdfObject()
}

// GetContainingPdfObject returns the container of the outline (indirect object).
func (o *PdfOutline) GetContainingPdfObject() core.PdfObject {
	return o.primitive
}

// ToPdfObject recursively builds the Outline tree PDF object.
func (o *PdfOutline) ToPdfObject() core.PdfObject {
	container := o.primitive
	dict := container.PdfObject.(*core.PdfObjectDictionary)

	dict.Set("Type", core.MakeName("Outlines"))

	if o.First != nil {
		dict.Set("First", o.First.ToPdfObject())
	}

	if o.Last != nil {
		dict.Set("Last", o.Last.GetContext().GetContainingPdfObject())
	}

	if o.Parent != nil {
		dict.Set("Parent", o.Parent.GetContext().GetContainingPdfObject())
	}

	if o.Count != nil {
		dict.Set("Count", core.MakeInteger(*o.Count))
	}

	return container
}

// GetContainingPdfObject returns the container of the outline item (indirect object).
func (oi *PdfOutlineItem) GetContainingPdfObject() core.PdfObject {
	return oi.primitive
}

// ToPdfObject recursively builds the Outline tree PDF object.
func (oi *PdfOutlineItem) ToPdfObject() core.PdfObject {
	container := oi.primitive
	dict := container.PdfObject.(*core.PdfObjectDictionary)

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
		dict.Set("Count", core.MakeInteger(*oi.Count))
	}
	if oi.Next != nil {
		dict.Set("Next", oi.Next.ToPdfObject())
	}
	if oi.First != nil {
		dict.Set("First", oi.First.ToPdfObject())
	}
	if oi.Prev != nil {
		dict.Set("Prev", oi.Prev.GetContext().GetContainingPdfObject())
	}
	if oi.Last != nil {
		dict.Set("Last", oi.Last.GetContext().GetContainingPdfObject())
	}
	if oi.Parent != nil {
		dict.Set("Parent", oi.Parent.GetContext().GetContainingPdfObject())
	}

	return container
}
