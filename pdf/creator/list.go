/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
)

// listItem represents a list item used in the list component.
type listItem struct {
	drawable VectorDrawable
	marker   TextChunk
}

// List represents a list of items.
// The representation of a list item is as follows:
//       [marker] [content]
// e.g.:        • This is the content of the item.
// The supported components to add content to list items are:
// - Paragraph
// - StyledParagraph
// - List
type List struct {
	// The items of the list.
	items []*listItem

	// Margins to be applied around the block when drawing on Page.
	margins margins

	// The marker symbol of the list items.
	marker TextChunk

	// The left offset of the list when nested into another list.
	indent float64

	// Specifies if the user has changed the default indent value of the list.
	defaultIndent bool

	// Positioning: relative/absolute.
	positioning positioning

	// Default style used for internal operations.
	defaultStyle TextStyle
}

// newList returns a new instance of List.
func newList(style TextStyle) *List {
	return &List{
		marker: TextChunk{
			Text:  "\u2022 ",
			Style: style,
		},
		indent:        0,
		defaultIndent: true,
		positioning:   positionRelative,
		defaultStyle:  style,
	}
}

// Add appends a new item to the list.
// The supported components are: *Paragraph, *StyledParagraph and *List.
// Returns the marker used for the newly added item. The returned marker
// object can be used to change the text and style of the marker for the
// current item.
func (l *List) Add(item VectorDrawable) (*TextChunk, error) {
	listItem := &listItem{
		drawable: item,
		marker:   l.marker,
	}

	switch t := item.(type) {
	case *Paragraph:
	case *StyledParagraph:
	case *List:
		if t.defaultIndent {
			t.indent = 15
		}
	default:
		return nil, errors.New("this type of drawable is not supported in list")
	}

	l.items = append(l.items, listItem)
	return &listItem.marker, nil
}

// AddTextItem appends a new item with the specified text to the list.
// The method creates a styled paragraph with the specified text and returns
// it so that the item style can be customized.
// The method also returns the marker used for the newly added item.
// The marker object can be used to change the text and style of the marker
// for the current item.
func (l *List) AddTextItem(text string) (*StyledParagraph, *TextChunk, error) {
	item := newStyledParagraph(l.defaultStyle)
	item.Append(text)

	marker, err := l.Add(item)
	return item, marker, err
}

// Marker returns the marker used for the list items.
// The marker instance can be used the change the text and the style
// of newly added list items.
func (l *List) Marker() *TextChunk {
	return &l.marker
}

// Indent returns the left offset of the list when nested into another list.
func (l *List) Indent() float64 {
	return l.indent
}

// SetIndent sets the left offset of the list when nested into another list.
func (l *List) SetIndent(indent float64) {
	l.indent = indent
	l.defaultIndent = false
}

// Margins returns the margins of the list: left, right, top, bottom.
func (l *List) Margins() (float64, float64, float64, float64) {
	return l.margins.left, l.margins.right, l.margins.top, l.margins.bottom
}

// SetMargins sets the margins of the paragraph.
func (l *List) SetMargins(left, right, top, bottom float64) {
	l.margins.left = left
	l.margins.right = right
	l.margins.top = top
	l.margins.bottom = bottom
}

// Width is not used. The list component is designed to fill into the available
// width depending on the context. Returns 0.
func (l *List) Width() float64 {
	return 0
}

// Height returns the height of the list.
func (l *List) Height() float64 {
	var height float64
	for _, item := range l.items {
		height += item.drawable.Height()
	}

	return height
}

// tableHeight returns the height of the list when used inside a table.
func (l *List) tableHeight(width float64) float64 {
	var height float64
	for _, item := range l.items {
		switch t := item.drawable.(type) {
		case *Paragraph:
			p := t
			if p.enableWrap {
				p.SetWidth(width)
			}

			height += p.Height() + p.margins.bottom + p.margins.bottom
			height += 0.5 * p.fontSize * p.lineHeight
		case *StyledParagraph:
			sp := t
			if sp.enableWrap {
				sp.SetWidth(width)
			}

			height += sp.Height() + sp.margins.top + sp.margins.bottom
			height += 0.5 * sp.getTextHeight()
		default:
			height += item.drawable.Height()
		}
	}

	return height
}

// GeneratePageBlocks generate the Page blocks. Multiple blocks are generated
// if the contents wrap over multiple pages.
func (l *List) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	// Create item markers.
	var markerWidth float64
	var markers []*StyledParagraph

	for _, item := range l.items {
		marker := newStyledParagraph(l.defaultStyle)
		marker.SetEnableWrap(false)
		marker.SetTextAlignment(TextAlignmentRight)
		marker.Append(item.marker.Text).Style = item.marker.Style

		width := marker.getTextWidth() / 1000.0 / ctx.Width
		if markerWidth < width {
			markerWidth = width
		}

		markers = append(markers, marker)
	}

	// Draw items.
	table := newTable(2)
	table.SetColumnWidths(markerWidth, 1-markerWidth)
	table.SetMargins(l.indent, 0, 0, 0)

	for i, item := range l.items {
		cell := table.NewCell()
		cell.SetIndent(0)
		cell.SetContent(markers[i])

		cell = table.NewCell()
		cell.SetIndent(0)
		cell.SetContent(item.drawable)
	}

	return table.GeneratePageBlocks(ctx)
}
