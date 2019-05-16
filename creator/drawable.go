/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// Drawable is a widget that can be used to draw with the Creator.
type Drawable interface {
	// GeneratePageBlocks draw onto blocks representing Page contents. As the content can wrap over many pages, multiple
	// templates are returned, one per Page.  The function also takes a draw context containing information
	// where to draw (if relative positioning) and the available height to draw on accounting for Margins etc.
	GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error)
}

// VectorDrawable is a Drawable with a specified width and height.
type VectorDrawable interface {
	Drawable

	// Width returns the width of the Drawable.
	Width() float64

	// Height returns the height of the Drawable.
	Height() float64
}

// DrawContext defines the drawing context. The DrawContext is continuously used and updated when
// drawing the page contents in relative mode.  Keeps track of current X, Y position, available
// height as well as other page parameters such as margins and dimensions.
type DrawContext struct {
	// Current page number.
	Page int

	// Current position.  In a relative positioning mode, a drawable will be placed at these coordinates.
	X, Y float64

	// Context dimensions.  Available width and height (on current page).
	Width, Height float64

	// Page Margins.
	Margins margins

	// Absolute Page size, widths and height.
	PageWidth  float64
	PageHeight float64

	// Controls whether the components are stacked horizontally
	Inline bool
}
