/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// OutlineDest represents the destination of an outline item.
// It holds the page and the position on the page an outline item points to.
type OutlineDest struct {
	Page int64   `json:"page"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// NewOutlineDest returns a new outline destination which can be used
// with outline items.
func NewOutlineDest(page int64, x, y float64) OutlineDest {
	return OutlineDest{
		Page: page,
		X:    x,
		Y:    y,
	}
}

// newOutlineDestFromPdfObject creates a new outline destination from the
// specified PDF object.
func newOutlineDestFromPdfObject(o core.PdfObject, r *PdfReader) OutlineDest {
	dest := OutlineDest{}

	destArr, ok := core.GetArray(o)
	if !ok {
		return dest
	}

	// Covered destination formats:
	// [pageObj|pageNum /Fit]
	// [pageObj|pageNum /FitB]
	// [pageObj|pageNum /FitH top]
	// [pageObj|pageNum /FitV left]
	// [pageObj|pageNum /FitBH top]
	// [pageObj|pageNum /XYZ x y zoom]
	// [pageObj|pageNum /FitR left bottom right top]
	// See section 12.3.2.2 "Explicit Destinations" (page 374).
	destArrLen := destArr.Len()
	if destArrLen < 2 {
		return dest
	}

	// Extract page number.
	pageObj := destArr.Get(0)
	if pageInd, ok := core.GetIndirect(pageObj); ok {
		// Page object is provided. Identify page number using the reader.
		if _, pageNum, err := r.PageFromIndirectObject(pageInd); err == nil {
			dest.Page = int64(pageNum - 1)
		}
	} else if pageNum, ok := core.GetIntVal(pageObj); ok {
		// Page number is provided.
		dest.Page = int64(pageNum)
	}

	// Extract destination coordinates.
	if destArrLen == 5 {
		if xyz, ok := core.GetName(destArr.Get(1)); ok && xyz.String() == "XYZ" {
			dest.X, _ = core.GetFloatVal(destArr.Get(2))
			dest.Y, _ = core.GetFloatVal(destArr.Get(3))
		}
	}

	return dest
}

// ToPdfObject returns a PDF object representation of the outline destination.
func (od OutlineDest) ToPdfObject() core.PdfObject {
	return core.MakeArray(
		core.MakeInteger(od.Page),
		core.MakeName("XYZ"),
		core.MakeFloat(od.X),
		core.MakeFloat(od.Y),
		core.MakeFloat(0),
	)
}

// Outline represents a PDF outline dictionary (Table 152 - p. 376).
// Currently, the Outline object can only be used to construct PDF outlines.
type Outline struct {
	Entries []*OutlineItem `json:"entries"`
}

// NewOutline returns a new outline instance.
func NewOutline() *Outline {
	return &Outline{}
}

// NewOutlineFromReaderOutline returns a new outline from the outline tree of
// the passed in reader.
func NewOutlineFromReaderOutline(r *PdfReader) (*Outline, error) {
	if r == nil {
		return nil, errors.New("cannot create outline from nil reader")
	}

	outlineTree := r.GetOutlineTree()
	if outlineTree == nil {
		return nil, errors.New("the specified reader does not have an outline tree")
	}

	var traverseFunc func(node *PdfOutlineTreeNode, entries *[]*OutlineItem)
	traverseFunc = func(node *PdfOutlineTreeNode, entries *[]*OutlineItem) {
		if node == nil {
			return
		}
		if node.context == nil {
			common.Log.Debug("ERROR: missing outline entry context")
			return
		}

		// Check if node is an outline item.
		var entry *OutlineItem
		if item, ok := node.context.(*PdfOutlineItem); ok {
			entry = NewOutlineItem(item.Title.Decoded(), newOutlineDestFromPdfObject(item.Dest, r))
			*entries = append(*entries, entry)

			// Traverse next node.
			if item.Next != nil {
				traverseFunc(item.Next, entries)
			}
		}

		// Check if node has children.
		if node.First != nil {
			if entry != nil {
				entries = &entry.Entries
			}

			// Traverse node children.
			traverseFunc(node.First, entries)
		}
	}

	outline := NewOutline()
	traverseFunc(outlineTree, &outline.Entries)
	return outline, nil
}

// Add appends a top level outline item to the outline.
func (o *Outline) Add(item *OutlineItem) {
	o.Entries = append(o.Entries, item)
}

// Insert adds a top level outline item in the outline,
// at the specified index.
func (o *Outline) Insert(index uint, item *OutlineItem) {
	l := uint(len(o.Entries))
	if index > l {
		index = l
	}

	o.Entries = append(o.Entries[:index], append([]*OutlineItem{item}, o.Entries[index:]...)...)
}

// Items returns all children outline items.
func (o *Outline) Items() []*OutlineItem {
	return o.Entries
}

// ToPdfOutline returns a low level PdfOutline object, based on the current
// instance.
func (o *Outline) ToPdfOutline() *PdfOutline {
	// Create outline.
	outline := NewPdfOutline()

	// Create outline items.
	var outlineItems []*PdfOutlineItem
	var prev *PdfOutlineItem

	for _, item := range o.Entries {
		outlineItem, _ := item.ToPdfOutlineItem()
		outlineItem.Parent = &outline.PdfOutlineTreeNode

		if prev != nil {
			prev.Next = &outlineItem.PdfOutlineTreeNode
			outlineItem.Prev = &prev.PdfOutlineTreeNode
		}

		outlineItems = append(outlineItems, outlineItem)
		prev = outlineItem
	}

	// Add outline linked list properties.
	lenOutlineItems := int64(len(outlineItems))
	if lenOutlineItems > 0 {
		outline.First = &outlineItems[0].PdfOutlineTreeNode
		outline.Last = &outlineItems[lenOutlineItems-1].PdfOutlineTreeNode
		outline.Count = &lenOutlineItems
	}

	return outline
}

// ToOutlineTree returns a low level PdfOutlineTreeNode object, based on
// the current instance.
func (o *Outline) ToOutlineTree() *PdfOutlineTreeNode {
	return &o.ToPdfOutline().PdfOutlineTreeNode
}

// ToPdfObject returns a PDF object representation of the outline.
func (o *Outline) ToPdfObject() core.PdfObject {
	return o.ToPdfOutline().ToPdfObject()
}

// OutlineItem represents a PDF outline item dictionary (Table 153 - pp. 376 - 377).
type OutlineItem struct {
	Title   string         `json:"title"`
	Dest    OutlineDest    `json:"dest"`
	Entries []*OutlineItem `json:"entries"`
}

// NewOutlineItem returns a new outline item instance.
func NewOutlineItem(title string, dest OutlineDest) *OutlineItem {
	return &OutlineItem{
		Title: title,
		Dest:  dest,
	}
}

// Add appends an outline item as a child of the current outline item.
func (oi *OutlineItem) Add(item *OutlineItem) {
	oi.Entries = append(oi.Entries, item)
}

// Insert adds an outline item as a child of the current outline item,
// at the specified index.
func (oi *OutlineItem) Insert(index uint, item *OutlineItem) {
	l := uint(len(oi.Entries))
	if index > l {
		index = l
	}

	oi.Entries = append(oi.Entries[:index], append([]*OutlineItem{item}, oi.Entries[index:]...)...)
}

// Items returns all children outline items.
func (oi *OutlineItem) Items() []*OutlineItem {
	return oi.Entries
}

// ToPdfOutlineItem returns a low level PdfOutlineItem object,
// based on the current instance.
func (oi *OutlineItem) ToPdfOutlineItem() (*PdfOutlineItem, int64) {
	// Create outline item.
	currItem := NewPdfOutlineItem()
	currItem.Title = core.MakeEncodedString(oi.Title, true)
	currItem.Dest = oi.Dest.ToPdfObject()

	// Create outline items.
	var outlineItems []*PdfOutlineItem
	var lenDescendants int64
	var prev *PdfOutlineItem

	for _, item := range oi.Entries {
		outlineItem, lenChildren := item.ToPdfOutlineItem()
		outlineItem.Parent = &currItem.PdfOutlineTreeNode

		if prev != nil {
			prev.Next = &outlineItem.PdfOutlineTreeNode
			outlineItem.Prev = &prev.PdfOutlineTreeNode
		}

		outlineItems = append(outlineItems, outlineItem)
		lenDescendants += lenChildren
		prev = outlineItem
	}

	// Add outline item linked list properties.
	lenOutlineItems := len(outlineItems)
	lenDescendants += int64(lenOutlineItems)

	if lenOutlineItems > 0 {
		currItem.First = &outlineItems[0].PdfOutlineTreeNode
		currItem.Last = &outlineItems[lenOutlineItems-1].PdfOutlineTreeNode
		currItem.Count = &lenDescendants
	}

	return currItem, lenDescendants
}

// ToPdfObject returns a PDF object representation of the outline item.
func (oi *OutlineItem) ToPdfObject() core.PdfObject {
	outlineItem, _ := oi.ToPdfOutlineItem()
	return outlineItem.ToPdfObject()
}
