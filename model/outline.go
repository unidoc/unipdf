/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"github.com/unidoc/unipdf/v3/core"
)

// OutlineDest represents the destination of an outline item.
// It holds the page and the position on the page an outline item points to.
type OutlineDest struct {
	Page int64
	X    float64
	Y    float64
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
	items []*OutlineItem
}

// NewOutline returns a new outline instance.
func NewOutline() *Outline {
	return &Outline{}
}

// Add appends a top level outline item to the outline.
func (o *Outline) Add(item *OutlineItem) {
	o.items = append(o.items, item)
}

// Insert adds a top level outline item in the outline,
// at the specified index.
func (o *Outline) Insert(index uint, item *OutlineItem) {
	l := uint(len(o.items))
	if index > l {
		index = l
	}

	o.items = append(o.items[:index], append([]*OutlineItem{item}, o.items[index:]...)...)
}

// Items returns all children outline items.
func (o *Outline) Items() []*OutlineItem {
	return o.items
}

// ToPdfOutline returns a low level PdfOutline object, based on the current
// instance.
func (o *Outline) ToPdfOutline() *PdfOutline {
	// Create outline.
	outline := NewPdfOutline()

	// Create outline items.
	var outlineItems []*PdfOutlineItem
	var prev *PdfOutlineItem

	for _, item := range o.items {
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

// ToPdfObject returns a PDF object representation of the outline.
func (o *Outline) ToPdfObject() core.PdfObject {
	return o.ToPdfOutline().ToPdfObject()
}

// OutlineItem represents a PDF outline item dictionary (Table 153 - pp. 376 - 377).
type OutlineItem struct {
	Title string
	Dest  OutlineDest

	items []*OutlineItem
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
	oi.items = append(oi.items, item)
}

// Insert adds an outline item as a child of the current outline item,
// at the specified index.
func (oi *OutlineItem) Insert(index uint, item *OutlineItem) {
	l := uint(len(oi.items))
	if index > l {
		index = l
	}

	oi.items = append(oi.items[:index], append([]*OutlineItem{item}, oi.items[index:]...)...)
}

// Items returns all children outline items.
func (oi *OutlineItem) Items() []*OutlineItem {
	return oi.items
}

// ToPdfOutlineItem returns a low level PdfOutlineItem object,
// based on the current instance.
func (oi *OutlineItem) ToPdfOutlineItem() (*PdfOutlineItem, int64) {
	// Create outline item.
	currItem := NewPdfOutlineItem()
	currItem.Title = core.MakeString(oi.Title)
	currItem.Dest = oi.Dest.ToPdfObject()

	// Create outline items.
	var outlineItems []*PdfOutlineItem
	var lenDescendants int64
	var prev *PdfOutlineItem

	for _, item := range oi.items {
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
