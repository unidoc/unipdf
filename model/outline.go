/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// OutlineDest represents the destination of an outline item.
// It holds the page and the position on the page an outline item points to.
type OutlineDest struct {
	PageObj *core.PdfIndirectObject `json:"-"`
	Page    int64                   `json:"page"`
	Mode    string                  `json:"mode"`
	X       float64                 `json:"x"`
	Y       float64                 `json:"y"`
	Zoom    float64                 `json:"zoom"`
}

// NewOutlineDest returns a new outline destination which can be used
// with outline items.
func NewOutlineDest(page int64, x, y float64) OutlineDest {
	return OutlineDest{
		Page: page,
		Mode: "XYZ",
		X:    x,
		Y:    y,
	}
}

// newOutlineDestFromPdfObject creates a new outline destination from the
// specified PDF object.
func newOutlineDestFromPdfObject(o core.PdfObject, r *PdfReader) (*OutlineDest, error) {
	// Validate input PDF object.
	destArr, ok := core.GetArray(o)
	if !ok {
		return nil, errors.New("outline destination object must be an array")
	}

	destArrLen := destArr.Len()
	if destArrLen < 2 {
		return nil, fmt.Errorf("invalid outline destination array length: %d", destArrLen)
	}

	// Extract page number.
	dest := &OutlineDest{Mode: "Fit"}

	pageObj := destArr.Get(0)
	if pageInd, ok := core.GetIndirect(pageObj); ok {
		// Page object is provided. Identify page number using the reader.
		if _, pageNum, err := r.PageFromIndirectObject(pageInd); err == nil {
			dest.Page = int64(pageNum - 1)
		} else {
			common.Log.Debug("WARN: could not get page index for page %+v", pageInd)
		}
		dest.PageObj = pageInd
	} else if pageIdx, ok := core.GetIntVal(pageObj); ok {
		// Page index is provided. Get indirect object to page.
		if pageIdx >= 0 && pageIdx < len(r.PageList) {
			dest.PageObj = r.PageList[pageIdx].GetPageAsIndirectObject()
		} else {
			common.Log.Debug("WARN: could not get page container for page %d", pageIdx)
		}
		dest.Page = int64(pageIdx)
	} else {
		return nil, fmt.Errorf("invalid outline destination page: %T", pageObj)
	}

	// Extract magnification mode.
	mode, ok := core.GetNameVal(destArr.Get(1))
	if !ok {
		common.Log.Debug("invalid outline destination magnification mode: %v", destArr.Get(1))
		return dest, nil
	}

	// Parse magnification mode parameters.
	// See section 12.3.2.2 "Explicit Destinations" (page 374).
	switch mode {
	// [pageObj|pageNum /Fit]
	// [pageObj|pageNum /FitB]
	case "Fit", "FitB":
	// [pageObj|pageNum /FitH top]
	// [pageObj|pageNum /FitBH top]
	case "FitH", "FitBH":
		if destArrLen > 2 {
			dest.Y, _ = core.GetNumberAsFloat(core.TraceToDirectObject(destArr.Get(2)))
		}
	// [pageObj|pageNum /FitV left]
	// [pageObj|pageNum /FitBV left]
	case "FitV", "FitBV":
		if destArrLen > 2 {
			dest.X, _ = core.GetNumberAsFloat(core.TraceToDirectObject(destArr.Get(2)))
		}
	// [pageObj|pageNum /XYZ x y zoom]
	case "XYZ":
		if destArrLen > 4 {
			dest.X, _ = core.GetNumberAsFloat(core.TraceToDirectObject(destArr.Get(2)))
			dest.Y, _ = core.GetNumberAsFloat(core.TraceToDirectObject(destArr.Get(3)))
			dest.Zoom, _ = core.GetNumberAsFloat(core.TraceToDirectObject(destArr.Get(4)))
		}
	default:
		mode = "Fit"
	}

	dest.Mode = mode
	return dest, nil
}

// ToPdfObject returns a PDF object representation of the outline destination.
func (od OutlineDest) ToPdfObject() core.PdfObject {
	if (od.PageObj == nil && od.Page < 0) || od.Mode == "" {
		return core.MakeNull()
	}

	// Add destination page.
	dest := core.MakeArray()
	if od.PageObj != nil {
		// Internal outline.
		dest.Append(od.PageObj)
	} else {
		// External outline.
		dest.Append(core.MakeInteger(od.Page))
	}

	// Add destination mode.
	dest.Append(core.MakeName(od.Mode))

	// See section 12.3.2.2 "Explicit Destinations" (page 374).
	switch od.Mode {
	// [pageObj|pageNum /Fit]
	// [pageObj|pageNum /FitB]
	case "Fit", "FitB":
	// [pageObj|pageNum /FitH top]
	// [pageObj|pageNum /FitBH top]
	case "FitH", "FitBH":
		dest.Append(core.MakeFloat(od.Y))
	// [pageObj|pageNum /FitV left]
	// [pageObj|pageNum /FitBV left]
	case "FitV", "FitBV":
		dest.Append(core.MakeFloat(od.X))
	// [pageObj|pageNum /XYZ x y zoom]
	case "XYZ":
		dest.Append(core.MakeFloat(od.X))
		dest.Append(core.MakeFloat(od.Y))
		dest.Append(core.MakeFloat(od.Zoom))
	default:
		dest.Set(1, core.MakeName("Fit"))
	}

	return dest
}

// Outline represents a PDF outline dictionary (Table 152 - p. 376).
// Currently, the Outline object can only be used to construct PDF outlines.
type Outline struct {
	Entries []*OutlineItem `json:"entries,omitempty"`
}

// NewOutline returns a new outline instance.
func NewOutline() *Outline {
	return &Outline{}
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
	var lenDescendants int64
	var prev *PdfOutlineItem

	for _, item := range o.Entries {
		outlineItem, lenChildren := item.ToPdfOutlineItem()
		outlineItem.Parent = &outline.PdfOutlineTreeNode

		if prev != nil {
			prev.Next = &outlineItem.PdfOutlineTreeNode
			outlineItem.Prev = &prev.PdfOutlineTreeNode
		}

		outlineItems = append(outlineItems, outlineItem)
		lenDescendants += lenChildren
		prev = outlineItem
	}

	// Add outline linked list properties.
	lenOutlineItems := int64(len(outlineItems))
	lenDescendants += int64(lenOutlineItems)

	if lenOutlineItems > 0 {
		outline.First = &outlineItems[0].PdfOutlineTreeNode
		outline.Last = &outlineItems[lenOutlineItems-1].PdfOutlineTreeNode
		outline.Count = &lenDescendants
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
	Entries []*OutlineItem `json:"entries,omitempty"`
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
