package pdf

import (
	"fmt"
)

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
