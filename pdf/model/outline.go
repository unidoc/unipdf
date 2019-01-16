/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"github.com/unidoc/unidoc/pdf/core"
)

/*
OutlineDest
*/
type OutlineDest struct {
	Page int64
	X    float64
	Y    float64
}

func NewOutlineDest(page int64, x, y float64) OutlineDest {
	return OutlineDest{
		Page: page,
		X:    x,
		Y:    y,
	}
}

func (od OutlineDest) ToPdfObject() core.PdfObject {
	return core.MakeArray(
		core.MakeInteger(od.Page),
		core.MakeName("XYZ"),
		core.MakeFloat(od.X),
		core.MakeFloat(od.Y),
		core.MakeFloat(0),
	)
}

/*
Outline
*/
type Outline struct {
	items []*OutlineItem
}

func NewOutline() *Outline {
	return &Outline{}
}

func (o *Outline) Add(item *OutlineItem) {
	o.items = append(o.items, item)
}

func (o *Outline) Insert(index uint, item *OutlineItem) {
	l := uint(len(o.items))
	if index > l {
		index = l
	}

	o.items = append(o.items[:index], append([]*OutlineItem{item}, o.items[index:]...)...)
}

func (o *Outline) Items() []*OutlineItem {
	return o.items
}

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

func (o *Outline) ToPdfObject() core.PdfObject {
	return o.ToPdfOutline().ToPdfObject()
}

/*
OutlineItem
*/
type OutlineItem struct {
	Title string
	Dest  OutlineDest

	items []*OutlineItem
}

func NewOutlineItem(title string, dest OutlineDest) *OutlineItem {
	return &OutlineItem{
		Title: title,
		Dest:  dest,
	}
}

func (oi *OutlineItem) Add(item *OutlineItem) {
	oi.items = append(oi.items, item)
}

func (oi *OutlineItem) Insert(index uint, item *OutlineItem) {
	l := uint(len(oi.items))
	if index > l {
		index = l
	}

	oi.items = append(oi.items[:index], append([]*OutlineItem{item}, oi.items[index:]...)...)
}

func (oi *OutlineItem) Items() []*OutlineItem {
	return oi.items
}

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

func (oi *OutlineItem) ToPdfObject() core.PdfObject {
	outlineItem, _ := oi.ToPdfOutlineItem()
	return outlineItem.ToPdfObject()
}
